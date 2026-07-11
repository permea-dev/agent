package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/transport"
)

// testDevID es el dev_id autoritativo (asignado por el backend) que viaja en los enrollment
// strings pmea2 de los tests; el agente lo adopta como su identidad de desarrollador.
const testDevID = "acme-dev-01"

// mkEnrollStr construye un enrollment string pmea2.<base64url(json{endpoint,token,dev_id})>
// para los tests, con el mismo formato que el backend (P-002b) emite.
func mkEnrollStr(t *testing.T, endpoint, token string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"endpoint": endpoint, "token": token, "dev_id": testDevID})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return "pmea2." + base64.RawURLEncoding.EncodeToString(b)
}

// verifyVia construye un verify que confía en el servidor TLS de test (self-signed) y
// hace el ping real de lote vacío contra él.
func verifyVia(srv *httptest.Server) func(endpoint, token string) error {
	return func(endpoint, token string) error {
		c := transport.New(endpoint, token)
		c.HTTP = srv.Client()
		return c.Verify()
	}
}

func okBackend(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, &gotAuth
}

// T008 — camino feliz por argv: verifica (2xx) y persiste a 0600; salida con URL, sin token.
func TestEnroll_Argv_HappyPath(t *testing.T) {
	srv, gotAuth := okBackend(t)
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir) // controla config.DataDir en Linux

	const token = "dev_tok_happy"
	es := mkEnrollStr(t, srv.URL, token)

	var out bytes.Buffer
	if err := enroll([]string{es}, strings.NewReader(""), false, &out, verifyVia(srv)); err != nil {
		t.Fatalf("enroll por argv falló: %v", err)
	}

	path := filepath.Join(cfgDir, "permea", "config.json")
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config.json no persistido: %v", err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("permisos = %o, quiero 600 (SC-002)", perm)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Endpoint != srv.URL || cfg.DeviceToken != token {
		t.Errorf("config = (%q, %q), quiero (%q, %q)", cfg.Endpoint, cfg.DeviceToken, srv.URL, token)
	}
	// Fuente autoritativa: el dev_id persistido DEBE ser el del enrollment string (pmea2),
	// no uno autodeclarado. Es lo que luego estampa Event.dev_id vía newIngestContext.
	if cfg.DevID != testDevID {
		t.Errorf("dev_id persistido = %q, quiero el del enrollment string %q (autoritativo)", cfg.DevID, testDevID)
	}
	if *gotAuth != "Bearer "+token {
		t.Errorf("el backend recibió Authorization %q, quiero Bearer %q", *gotAuth, token)
	}

	// Higiene: la salida incluye la URL pero NUNCA el token ni el enrollment string (SC-003).
	if !strings.Contains(out.String(), srv.URL) {
		t.Errorf("la salida de éxito no muestra la URL: %q", out.String())
	}
	if strings.Contains(out.String(), token) || strings.Contains(out.String(), es) {
		t.Errorf("la salida de éxito filtra el secreto: %q", out.String())
	}
}

// T009 — vía stdin (`-` y sin argumento) + SC-011 (token no en argv) + sin eco.
func TestEnroll_Stdin_And_SC011(t *testing.T) {
	srv, _ := okBackend(t)
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	const token = "dev_tok_stdin"
	es := mkEnrollStr(t, srv.URL, token)

	// (a) stdin explícito con "-" (echo añade \n, que se recorta).
	args := []string{"-"}
	var out bytes.Buffer
	if err := enroll(args, strings.NewReader(es+"\n"), true, &out, verifyVia(srv)); err != nil {
		t.Fatalf("enroll por stdin (-) falló: %v", err)
	}
	cfg, err := config.Load(filepath.Join(cfgDir, "permea", "config.json"))
	if err != nil || cfg.DeviceToken != token {
		t.Fatalf("stdin no persistió el token: cfg=%q err=%v", cfg.DeviceToken, err)
	}

	// SC-011: el token NUNCA aparece en los argumentos del comando.
	if strings.Contains(strings.Join(args, " "), token) {
		t.Errorf("SC-011: el token aparece en los args del subcomando: %v", args)
	}
	for _, a := range os.Args {
		if strings.Contains(a, token) {
			t.Errorf("SC-011: el token aparece en os.Args: %q", a)
		}
	}
	// Sin eco del secreto en la salida.
	if strings.Contains(out.String(), token) || strings.Contains(out.String(), es) {
		t.Errorf("la vía stdin filtra el secreto en la salida: %q", out.String())
	}

	// (b) sin argumento, stdin es un pipe → mismo flujo.
	var out2 bytes.Buffer
	if err := enroll(nil, strings.NewReader(es), true, &out2, verifyVia(srv)); err != nil {
		t.Fatalf("enroll por stdin (sin argumento, pipe) falló: %v", err)
	}
}

// T009 (guard TTY) — sin argumento y stdin NO-pipe (TTY interactiva) → error de uso,
// exit ≠ 0, no se cuelga (no lee stdin) y no persiste nada.
func TestEnroll_NoArg_NoPipe_UsageError(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	verifyMustNotRun := func(_, _ string) error {
		t.Fatal("verify NO debe ejecutarse en un error de uso")
		return nil
	}

	var out bytes.Buffer
	// stdinIsPipe=false simula una TTY interactiva. El reader tiene contenido a propósito:
	// el guard NO debe leerlo (si lo leyera, en una TTY real se colgaría).
	err := enroll(nil, strings.NewReader("no-debe-leerse"), false, &out, verifyMustNotRun)
	if err == nil {
		t.Fatal("esperaba error de uso cuando no hay argumento y stdin no es un pipe")
	}
	if _, statErr := os.Stat(filepath.Join(cfgDir, "permea", "config.json")); !os.IsNotExist(statErr) {
		t.Errorf("no debía persistir config.json en un error de uso")
	}
}
