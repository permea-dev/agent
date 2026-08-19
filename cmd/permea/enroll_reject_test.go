package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/transport"
)

// assertNoPersist verifica el estado indistinguible (SC-004): tras un rechazo/aborto,
// config.json NO existe (idéntico a no haber enrolado).
//
// ═══ P-005 · SE AMPLIÓ AL DIRECTORIO ENTERO, Y NO ES CELO ══════════════════════════════════
//
// Hasta P-005 esta red miraba **un solo fichero**. Bastaba mientras `enroll` **sólo** escribiera
// `config.json`. **Ya no es el caso**: P-005 le añadió la creación del `salt` al terminar un
// enrolamiento CORRECTO (`cmd/permea/enroll.go`), y esa escritura vive **dentro** de una rama.
//
// **El día que esas líneas se muevan** —un refactor, un `defer` mal puesto, alguien subiéndolas para
// "resolver las dependencias antes"— el enrolamiento rechazado empezaría a dejar un secreto en el
// disco, y **la red que existe para cazarlo estaría mirando el fichero equivocado**. Un `config.json`
// ausente es compatible con un directorio lleno.
//
// Por eso se comprueba **el directorio completo, derivado recorriéndolo**, no una lista de nombres:
// una lista escrita a mano se queda corta con el siguiente artefacto, que es exactamente el defecto
// que este repositorio ya registró tres veces.
//
// **Las dos aserciones son independientes y las dos usan `t.Errorf`**: con un `Fatalf` en la primera,
// la segunda quedaría inalcanzable para cualquier mutación y pasaría por validada sin haberse
// evaluado jamás.
func assertNoPersist(t *testing.T, cfgDir string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(cfgDir, "permea", "config.json")); !os.IsNotExist(err) {
		t.Errorf("no debía persistir config.json (estado indistinguible, SC-004); stat err=%v", err)
	}
	if residuos := ficherosDeDatos(t, filepath.Join(cfgDir, "permea")); len(residuos) > 0 {
		t.Errorf("un enrolamiento rechazado dejó %d fichero(s) en el directorio de datos: %v\n"+
			"El estado indistinguible (SC-004) es el del directorio ENTERO, no el de `config.json`: "+
			"cualquier otro artefacto —el `salt` entre ellos— delata que hubo un intento",
			len(residuos), residuos)
	}
}

// ficherosDeDatos enumera los ficheros que hay bajo `dir`, recursivamente. Un directorio inexistente
// son cero ficheros, sin error: es el estado de quien nunca enroló, y es el que se espera aquí.
func ficherosDeDatos(t *testing.T, dir string) []string {
	t.Helper()
	var encontrados []string
	err := filepath.WalkDir(dir, func(ruta string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return filepath.SkipAll
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, ruta)
		if err != nil {
			return err
		}
		encontrados = append(encontrados, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("enumerar el directorio de datos %q: %v", dir, err)
	}
	sort.Strings(encontrados)
	return encontrados
}

// assertNoSecret verifica la higiene: la cadena s no contiene el token ni el enrollment
// string (ni el argumento), en ninguna rama de error (FR-007).
func assertNoSecret(t *testing.T, where, s, token, arg string) {
	t.Helper()
	if token != "" && strings.Contains(s, token) {
		t.Errorf("%s filtra el token: %q", where, s)
	}
	if arg != "" && strings.Contains(s, arg) {
		t.Errorf("%s reproduce el argumento/enrollment string: %q", where, s)
	}
}

// T011(a) — 401/403: token rechazado. No persiste; mensaje de rechazo; sin secreto.
func TestEnroll_Reject_401_NoPersist(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	const token = "dev_tok_revoked"
	es := mkEnrollStr(t, srv.URL, token)

	var out strings.Builder
	err := enroll([]string{es}, strings.NewReader(""), false, &out, verifyVia(srv))
	if err == nil {
		t.Fatal("401 debe devolver error (enrolamiento rechazado)")
	}
	assertNoPersist(t, cfgDir)
	assertNoSecret(t, "el error de 401", err.Error(), token, es)
	assertNoSecret(t, "la salida de 401", out.String(), token, es)

	// Debe distinguirse de "no verificable": mensaje de RECHAZO.
	if !strings.Contains(err.Error(), "rechazado") {
		t.Errorf("el mensaje de 401 debe indicar rechazo del token; got %q", err.Error())
	}
}

// T011(b) — 5xx: no verificable (transitorio). No persiste; mensaje distinto; sin secreto.
func TestEnroll_Reject_5xx_NoPersist(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	const token = "dev_tok_5xx"
	es := mkEnrollStr(t, srv.URL, token)

	var out strings.Builder
	err := enroll([]string{es}, strings.NewReader(""), false, &out, verifyVia(srv))
	if err == nil {
		t.Fatal("5xx debe devolver error (no verificable)")
	}
	assertNoPersist(t, cfgDir)
	assertNoSecret(t, "el error de 5xx", err.Error(), token, es)

	if !strings.Contains(err.Error(), "no se pudo verificar") {
		t.Errorf("el mensaje de 5xx debe indicar 'no se pudo verificar'; got %q", err.Error())
	}
	if strings.Contains(err.Error(), "rechazado") {
		t.Errorf("5xx NO debe clasificarse como rechazo (auth); got %q", err.Error())
	}
}

// T011(b') — error de red: backend inalcanzable (servidor cerrado). No verificable.
func TestEnroll_Reject_Network_NoPersist(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	client := srv.Client()
	url := srv.URL
	srv.Close() // ahora el endpoint es inalcanzable → error de red (no-auth)

	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	const token = "dev_tok_net"
	es := mkEnrollStr(t, url, token)

	verify := func(endpoint, tok string) error {
		c := transport.New(endpoint, tok)
		c.HTTP = client
		return c.Verify()
	}

	var out strings.Builder
	err := enroll([]string{es}, strings.NewReader(""), false, &out, verify)
	if err == nil {
		t.Fatal("error de red debe devolver error (no verificable)")
	}
	assertNoPersist(t, cfgDir)
	assertNoSecret(t, "el error de red", err.Error(), token, es)
	if !strings.Contains(err.Error(), "no se pudo verificar") {
		t.Errorf("el mensaje de red debe indicar 'no se pudo verificar'; got %q", err.Error())
	}
}

// T011(c) — enrollment string malformado y http:// : aborta ANTES del ping (verify no
// se llama), no persiste y no reproduce el argumento.
func TestEnroll_Reject_Malformed_AbortsBeforePing(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	verifyMustNotRun := func(_, _ string) error {
		t.Fatal("verify NO debe ejecutarse: un enrollment string inválido aborta antes del ping")
		return nil
	}

	t.Run("base64 malformado", func(t *testing.T) {
		const arg = "pmea2.###malformado###"
		var out strings.Builder
		err := enroll([]string{arg}, strings.NewReader(""), false, &out, verifyMustNotRun)
		if err == nil {
			t.Fatal("un enrollment string malformado debe devolver error")
		}
		assertNoPersist(t, cfgDir)
		assertNoSecret(t, "el error de malformado", err.Error(), "", arg)
	})

	t.Run("endpoint http (no https)", func(t *testing.T) {
		const token = "dev_tok_http"
		es := mkEnrollStr(t, "http://inseguro.example/ingest", token)
		var out strings.Builder
		err := enroll([]string{es}, strings.NewReader(""), false, &out, verifyMustNotRun)
		if err == nil {
			t.Fatal("un endpoint http:// debe rechazarse en la decodificación")
		}
		assertNoPersist(t, cfgDir)
		assertNoSecret(t, "el error de http://", err.Error(), token, es)
	})
}

// Política de versión — un enrollment string pmea1 (formato viejo) se RECHAZA antes del
// ping (verify no se ejecuta), no persiste y el error indica formato obsoleto sin filtrar
// el argumento ni el token.
func TestEnroll_Reject_Pmea1_Obsolete(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	verifyMustNotRun := func(_, _ string) error {
		t.Fatal("verify NO debe ejecutarse: un pmea1 se rechaza en la decodificación")
		return nil
	}

	const token = "dev_tok_viejo_pmea1"
	b, err := json.Marshal(map[string]string{"endpoint": "https://x.example/ingest", "token": token})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	es := "pmea1." + base64.RawURLEncoding.EncodeToString(b)

	var out strings.Builder
	err = enroll([]string{es}, strings.NewReader(""), false, &out, verifyMustNotRun)
	if err == nil {
		t.Fatal("un pmea1 debe rechazarse (formato obsoleto)")
	}
	assertNoPersist(t, cfgDir)
	assertNoSecret(t, "el error de pmea1", err.Error(), token, es)
	if !strings.Contains(err.Error(), "obsoleto") {
		t.Errorf("el error de pmea1 debe indicar formato obsoleto; got %q", err.Error())
	}
}

// T012 (FR-014/SC-010) — re-enrolar no deja residuo del token viejo.
func TestEnroll_ReenrollNoResidue(t *testing.T) {
	srv, _ := okBackend(t)
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	const tokenA = "dev_tok_AAA_viejo"
	const tokenB = "dev_tok_BBB_nuevo"

	if err := enroll([]string{mkEnrollStr(t, srv.URL, tokenA)}, strings.NewReader(""), false, io.Discard, verifyVia(srv)); err != nil {
		t.Fatalf("primer enroll (A) falló: %v", err)
	}
	if err := enroll([]string{mkEnrollStr(t, srv.URL, tokenB)}, strings.NewReader(""), false, io.Discard, verifyVia(srv)); err != nil {
		t.Fatalf("segundo enroll (B) falló: %v", err)
	}

	dir := filepath.Join(cfgDir, "permea")
	cfg, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.DeviceToken != tokenB {
		t.Errorf("config.json debe contener SOLO el token B; got %q", cfg.DeviceToken)
	}

	// Ningún fichero del directorio de config contiene el token viejo A (ni .bak ni temporal).
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if strings.Contains(string(b), tokenA) {
			t.Errorf("residuo del token viejo en %q (FR-014/SC-010)", e.Name())
		}
	}
}
