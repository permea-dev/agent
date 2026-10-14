package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// P-011 B4 · Copilot CLI en `--run` y en el demonio. Las sesiones son los ficheros SINTÉTICOS de
// `internal/ingest/testdata/copilot/ficheros/` (generar.sh), siempre copiados a un `t.TempDir()`: este paquete no tiene
// fixtures propios de sesiones. Reutiliza, sin tocarlos, los ayudantes de `codex_test.go`, `gemini_test.go` y
// `opencode_test.go`. Los tests de proceso corren en `testutil.Sandbox`, que deja `COPILOT_HOME` vacía (M-3); los que necesitan
// Copilot CLI la apuntan a un temporal.

// fixtureCopilot es un fichero de `internal/ingest/testdata/copilot/ficheros/`.
func fixtureCopilot(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "internal", "ingest", "testdata", "copilot", "ficheros", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return b
}

// homeCopilot crea `<COPILOT_HOME>/session-state/<carpeta>/events.jsonl` con los fixtures dados y devuelve el COPILOT_HOME.
func homeCopilot(t *testing.T, sesiones map[string]string) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "copilot-home")
	for carpeta, fixture := range sesiones {
		escribir(t, filepath.Join(home, "session-state", carpeta, "events.jsonl"), fixtureCopilot(t, fixture))
	}
	return home
}

// agenteCopilot es `agenteCodex`, sin Claude Code ni Codex, con la raíz de Copilot CLI (el campo que `setup()` resuelve).
func agenteCopilot(t *testing.T, dataDir, home string) *agent {
	t.Helper()
	a := agenteCodex(t, dataDir, t.TempDir(), "")
	a.copilotRaiz = filepath.Join(home, "session-state")
	return a
}

// lineaCopilot es la línea de spec §Textos, leída por programa, con los recuentos dados.
func lineaCopilot(t *testing.T, recuentos ...any) string {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "specs", "011-lector-copilot", "spec.md"))
	if err != nil {
		t.Fatalf("precondición: spec: %v", err)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 1<<20)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "copilot_cli: registros %d") {
			return fmt.Sprintf(sc.Text(), recuentos...)
		}
	}
	t.Fatal("precondición: la spec no tiene la línea del resumen de Copilot CLI")
	return ""
}

// dosSesiones son dos sesiones de Copilot CLI con 2 + 1 registros: 3 eventos.
var dosSesiones = map[string]string{"s-000000000000000000000001": "dos_llamadas.jsonl", "s-000000000000000000000002": "matada.jsonl"}

// todasLasHerramientas apunta Claude Code, Codex, Gemini y OpenCode a sus fixtures, como (41) de 010.
func todasLasHerramientas(t *testing.T, dataDir string) {
	t.Helper()
	logsDeClaude(t, dataDir)
	t.Setenv("CODEX_HOME", raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"}))
	t.Setenv("GEMINI_CLI_HOME", raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba))
	t.Setenv("XDG_DATA_HOME", datosOpenCode(t, map[string]string{"opencode.db": "sana.db"}))
}

// (27) · SC-015, FR-023: sin Copilot CLI, el stderr de `--run` sobre las otras cuatro herramientas es byte a byte la referencia
// generada con el binario de `15c4c47` (T025), con la ruta de datos sustituida por `<DATOS>`. NACE VERDE; lo valida m30.
func TestCopilotRun_SinCopilotEsLaSalidaDeLa070(t *testing.T) {
	referencia := string(fixtureCopilot2(t, "referencia-run.stderr"))
	for _, c := range []struct {
		hoja    string
		conHome bool
	}{
		{"copilot_home_vacia", false},
		{"sin_raiz", true},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			dataDir := testutil.Sandbox(t)
			todasLasHerramientas(t, dataDir)
			if c.conHome {
				t.Setenv("COPILOT_HOME", t.TempDir()) // existe, pero sin `session-state/`
			} else {
				t.Setenv("COPILOT_HOME", "") // `<hogar del sandbox>/.copilot/session-state`, que no existe
			}
			codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
			if got := strings.ReplaceAll(stderr, dataDir, "<DATOS>"); codigo != 0 || got != referencia {
				t.Errorf("código %d; stderr:\n%s\nreferencia:\n%s", codigo, got, referencia)
			}
		})
	}
}

// fixtureCopilot2 es un fichero de `cmd/permea/testdata/copilot/` (sólo la referencia de SC-015).
func fixtureCopilot2(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "copilot", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return b
}

// (28) · SC-019, FR-025: con sólo Copilot CLI, `--run` encola sus eventos y sale con 0.
func TestCopilotRun_SoloCopilot(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	t.Setenv("COPILOT_HOME", homeCopilot(t, dosSesiones))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "copilot_cli")); codigo != 0 || n != 3 {
		t.Errorf("código %d y %d eventos copilot_cli; se esperaba 0 y 3. stderr:\n%s", codigo, n, stderr)
	}
}

// (29) · SC-019, FR-024: con el forzado A, los eventos de Copilot CLI están en la cola aunque `st.Save` falle, y la pasada
// siguiente los reencola con los mismos `event_id`.
func TestCopilotGenerate_EncolaAntesDeGuardar(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := t.TempDir()
	a := agenteCopilot(t, dataDir, homeCopilot(t, dosSesiones))
	escribir(t, transport.QueuePath(dataDir), nil) // la cola existe: `Append` no necesita escribir en el directorio
	if err := os.Chmod(dataDir, 0o500); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(dataDir, 0o700) }()
	if _, _, err := a.generate(); err == nil {
		t.Fatalf("con el directorio de datos sin escritura, st.Save debía fallar y la pasada devolver error")
	}
	primera := cola(t, dataDir, "copilot_cli")
	if len(primera) != 3 {
		t.Fatalf("la pasada falló al guardar y la cola tiene %d eventos copilot_cli; se esperaban 3, encolados ANTES de guardar", len(primera))
	}
	if err := os.Chmod(dataDir, 0o700); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	if _, _, err := a.generate(); err != nil {
		t.Fatalf("con el estado sano: %v", err)
	}
	todos := cola(t, dataDir, "copilot_cli")
	if len(todos) != 6 {
		t.Fatalf("la segunda pasada debía reencolar los tres; cola copilot_cli: %d eventos", len(todos))
	}
	for i := range primera {
		if todos[3+i].EventID != primera[i].EventID {
			t.Errorf("evento %d reencolado con otro event_id", i)
		}
	}
}

// (30) · FR-021, FR-023: la línea `copilot_cli:`, literal, justo tras la de OpenCode; «N eventos encolados» cuenta también
// Copilot CLI; y en la cola, los eventos de OpenCode van ANTES que los de Copilot CLI (se lee después: FR-024; H5 del analyze).
func TestCopilotRun_LineaDeResumen(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	todasLasHerramientas(t, dataDir)
	t.Setenv("COPILOT_HOME", homeCopilot(t, dosSesiones))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	t.Run("linea", func(t *testing.T) {
		lineas := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
		want := lineaCopilot(t, 3, 3, 0, 0, 0, 0, 0)
		if codigo != 0 || len(lineas) != 9 || lineas[1] != "11 eventos encolados en "+transport.QueuePath(dataDir) ||
			!strings.HasPrefix(lineas[6], "opencode: llamadas 2 ") || lineas[7] != want || lineas[8] != "sync omitido: sin endpoint configurado" {
			t.Errorf("código %d; stderr:\n%s\nse esperaba «11 eventos encolados», los resúmenes de las otras cuatro y después:\n%s",
				codigo, stderr, want)
		}
	})
	t.Run("orden_cola", func(t *testing.T) {
		b, err := os.ReadFile(transport.QueuePath(dataDir))
		if err != nil {
			t.Fatalf("leer la cola: %v", err)
		}
		texto := string(b)
		ultimoOpenCode, primerCopilot := strings.LastIndex(texto, `"tool":"opencode"`), strings.Index(texto, `"tool":"copilot_cli"`)
		if ultimoOpenCode < 0 || primerCopilot < 0 || ultimoOpenCode > primerCopilot {
			t.Errorf("en la cola, el último opencode está en %d y el primer copilot_cli en %d; se esperaba OpenCode antes", ultimoOpenCode, primerCopilot)
		}
	})
}

// (31) · FR-021: en el demonio, la línea `copilot_cli:` sólo con novedades; una segunda pasada sin cambios calla.
func TestCopilotDemonio_SoloConNovedades(t *testing.T) {
	a := agenteCopilot(t, t.TempDir(), homeCopilot(t, dosSesiones))
	primero := capturarStderr(t, func() { _ = a.tick() })
	segundo := capturarStderr(t, func() { _ = a.tick() })
	if !strings.Contains(primero, lineaCopilot(t, 3, 3, 0, 0, 0, 0, 0)) || strings.Contains(segundo, "copilot_cli:") {
		t.Errorf("primer ciclo:\n%s\nsegundo ciclo, sin novedades:\n%s\nse esperaba la línea copilot_cli: sólo en el primero", primero, segundo)
	}
}

// (32) · SC-003: una segunda `--run` sin cambios da 0 eventos de Copilot CLI.
func TestCopilotRun_SegundaPasadaCero(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	t.Setenv("COPILOT_HOME", homeCopilot(t, dosSesiones))
	ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "copilot_cli")); n != 3 {
		t.Fatalf("precondición: la primera pasada encoló %d eventos copilot_cli; se esperaban 3", n)
	}
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "copilot_cli")); codigo != 0 || n != 3 || !strings.Contains(stderr, lineaCopilot(t, 0, 0, 0, 0, 0, 0, 0)) {
		t.Errorf("código %d, %d eventos copilot_cli en la cola; se esperaba 0, 3 y la línea vacía. stderr:\n%s", codigo, n, stderr)
	}
}

// ═══ Nacen verdes, tras el verde de T027 (T029) ════════════════════════════════════════════════════════════════════

// huellaArbol es el listado RECURSIVO de una carpeta con el md5 de cada fichero, `.lock` incluidos.
func huellaArbol(t *testing.T, raiz string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.WalkDir(raiz, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		contenido, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(raiz, p)
		fmt.Fprintf(&b, "%s %x\n", filepath.ToSlash(rel), md5.Sum(contenido))
		return nil
	})
	if err != nil {
		t.Fatalf("recorrer: %v", err)
	}
	return b.String()
}

// (33) · SC-011, FR-019: nada del proveedor ni del contenido en la cola; en `state.json`, los centinelas de la ruta (el nombre de
// la carpeta) sólo DENTRO de su clave. Centinelas en la sesión, el `cwd`, el nombre de la carpeta y todos los campos de texto del
// registro. NACE VERDE; lo valida m37.
func TestCopilotRun_NadaDelProveedorViaja(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	home := filepath.Join(t.TempDir(), "copilot-home")
	inicio := `{"type":"session.start","data":{"sessionId":"CENTINELA-sessionId","version":1,"copilotVersion":"0.0.0-sintetico",` +
		`"context":{"cwd":"/sintetico/CENTINELA-cwd"}},"id":"l-1","timestamp":"2026-01-01T00:00:00.000Z","parentId":null}` + "\n"
	registro, err := os.ReadFile(filepath.Join("..", "..", "internal", "ingest", "testdata", "copilot", "lineas", "centinelas.jsonl"))
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	ruta := filepath.Join(home, "session-state", "s-CENTINELA-carpeta", "events.jsonl")
	escribir(t, ruta, append([]byte(inicio), registro...))
	t.Setenv("COPILOT_HOME", home)
	ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "copilot_cli")); n != 1 {
		t.Fatalf("precondición: %d eventos copilot_cli en la cola; se esperaba 1", n)
	}
	enCola, err := os.ReadFile(transport.QueuePath(dataDir))
	if err != nil {
		t.Fatalf("leer la cola: %v", err)
	}
	if strings.Contains(string(enCola), "CENTINELA") {
		t.Errorf("la cola lleva un identificador del proveedor, un texto o una ruta:\n%s", enCola)
	}
	estado, err := os.ReadFile(filepath.Join(dataDir, "state.json"))
	if err != nil {
		t.Fatalf("leer state.json: %v", err)
	}
	if !strings.Contains(string(estado), `"`+ruta+`":`) {
		t.Fatalf("precondición: state.json no tiene la ruta del fichero como clave:\n%s", estado)
	}
	if resto := strings.ReplaceAll(string(estado), ruta, ""); strings.Contains(resto, "CENTINELA") {
		t.Errorf("state.json lleva un centinela fuera de la ruta del fichero:\n%s", estado)
	}
}

// (34) · SC-006, FR-006, D-6: `--run` deja iguales las md5 de todos los ficheros de la raíz y su listado, `.lock` incluidos: ni
// un byte, ni un fichero nuevo. NACE VERDE; lo valida M-B4a.
func TestCopilotRun_SoloLectura(t *testing.T) {
	_ = testutil.Sandbox(t)
	home := filepath.Join(t.TempDir(), "copilot-home")
	raiz := filepath.Join(home, "session-state")
	origen := filepath.Join("..", "..", "internal", "ingest", "testdata", "copilot", "ficheros", "solo_lectura")
	err := filepath.WalkDir(origen, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(origen, p)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		escribir(t, filepath.Join(raiz, rel), b)
		return nil
	})
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	t.Setenv("COPILOT_HOME", home)
	antes := huellaArbol(t, raiz)
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if despues := huellaArbol(t, raiz); codigo != 0 || despues != antes || !strings.Contains(stderr, lineaCopilot(t, 2, 2, 0, 0, 0, 0, 0)) {
		t.Errorf("código %d; antes:\n%s\ndespués:\n%s\nstderr:\n%s", codigo, antes, despues, stderr)
	}
}

// (35) · FR-002, P-6: la raíz se resuelve en `setup()`, y si existe se mira en CADA pasada: el mismo `agent` ve la carpeta de
// sesiones creada después. NACE VERDE; lo valida m34 (y m30, que la daría por activa sin carpeta).
func TestCopilotActivacion_EnCadaPasada(t *testing.T) {
	_ = testutil.Sandbox(t)
	home := filepath.Join(t.TempDir(), "copilot-home")
	t.Setenv("COPILOT_HOME", home)
	a, err := setup()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, _, err := a.generate(); err != nil || a.copilot != nil || len(cola(t, a.dir, "copilot_cli")) != 0 {
		t.Fatalf("sin la carpeta: err %v, recuentos %v, cola copilot_cli %d; se esperaba inactivo", err, a.copilot, len(cola(t, a.dir, "copilot_cli")))
	}
	escribir(t, filepath.Join(home, "session-state", "s-000000000000000000000001", "events.jsonl"), fixtureCopilot(t, "dos_llamadas.jsonl"))
	if _, _, err := a.generate(); err != nil || a.copilot == nil || len(cola(t, a.dir, "copilot_cli")) != 2 {
		t.Errorf("con la carpeta creada después: err %v, recuentos %v, cola copilot_cli %d; se esperaban 2 eventos",
			err, a.copilot, len(cola(t, a.dir, "copilot_cli")))
	}
}

// (36) · SC-018, FR-026: un fichero de Copilot CLI ilegible no tumba la pasada ni a Claude Code; el aviso no lleva su ruta; su
// offset no avanza; y cuando vuelve a ser legible, sus registros salen. NACE VERDE; lo validan m35 y m36.
func TestCopilotGenerate_FicheroIlegibleNoRompe(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := t.TempDir()
	logs := filepath.Join(t.TempDir(), "logs")
	escribir(t, filepath.Join(logs, "claude.jsonl"), fixtureB5(t, "claude.jsonl"))
	home := homeCopilot(t, dosSesiones)
	ilegible := filepath.Join(home, "session-state", "s-000000000000000000000002", "events.jsonl")
	if err := os.Chmod(ilegible, 0o000); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(ilegible, 0o600) }()
	a := agenteCodex(t, dataDir, logs, "")
	a.copilotRaiz = filepath.Join(home, "session-state")

	t.Run("pasada", func(t *testing.T) {
		var err error
		stderr := capturarStderr(t, func() { _, _, err = a.generate() })
		_, sinEstado := os.Stat(filepath.Join(dataDir, "state.json"))
		if err != nil || !strings.Contains(stderr, "copilot_cli: fichero omitido: ") || strings.Contains(stderr, home) || sinEstado != nil ||
			len(cola(t, dataDir, "claude_code")) != 2 || len(cola(t, dataDir, "copilot_cli")) != 2 {
			t.Errorf("err %v, state.json %v, claude %d, copilot_cli %d; se esperaba sin error, el aviso sin la ruta, el estado, 2 y 2. stderr:\n%s",
				err, sinEstado, len(cola(t, dataDir, "claude_code")), len(cola(t, dataDir, "copilot_cli")), stderr)
		}
	})
	t.Run("se_relee", func(t *testing.T) {
		if err := os.Chmod(ilegible, 0o600); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		_ = capturarStderr(t, func() { _, _, _ = a.generate() })
		if n := len(cola(t, dataDir, "copilot_cli")); n != 3 {
			t.Errorf("con el fichero ya legible: %d eventos copilot_cli; se esperaban 3 (2 + el suyo, porque su offset no avanzó)", n)
		}
	})
}
