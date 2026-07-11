package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/config"
)

// writeConfig materializa un config.json enrolado bajo el dir de config por-usuario.
func writeConfig(t *testing.T, cfgDir, endpoint, token string) {
	t.Helper()
	dir := filepath.Join(cfgDir, "permea")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfg := config.Default()
	cfg.Endpoint = endpoint
	cfg.DeviceToken = token
	if err := config.Save(filepath.Join(dir, "config.json"), cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
}

// T015(a) — enrolado: la salida muestra la URL pero NUNCA el token (SC-005).
func TestStatus_Enrolled_ShowsURLNotToken(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	const endpoint = "https://api.permea.example/ingest"
	const token = "dev_tok_status_SECRET"
	writeConfig(t, cfgDir, endpoint, token)

	var out strings.Builder
	if err := runStatus(&out); err != nil {
		t.Fatalf("runStatus: %v", err)
	}
	s := out.String()

	if !strings.Contains(s, "enrolado") {
		t.Errorf("esperaba informar 'enrolado'; got %q", s)
	}
	if !strings.Contains(s, endpoint) {
		t.Errorf("esperaba mostrar la URL del backend; got %q", s)
	}
	if strings.Contains(s, token) {
		t.Errorf("SC-005: status filtró el device_token; got %q", s)
	}
}

// T015(b) — no enrolado: sin config, informa 'no enrolado' sin error (exit 0) ni secreto.
func TestStatus_NotEnrolled(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir) // sin config.json → Load devuelve Default()

	var out strings.Builder
	if err := runStatus(&out); err != nil {
		t.Fatalf("runStatus no debe fallar sin enrolar (exit 0): %v", err)
	}
	if !strings.Contains(out.String(), "no enrolado") {
		t.Errorf("esperaba 'no enrolado'; got %q", out.String())
	}
}
