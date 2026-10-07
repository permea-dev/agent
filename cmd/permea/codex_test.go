package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ P-008 B5 · CODEX EN `--run` Y EN EL DEMONIO ════════════════════════════════════════════════════
//
// Fixtures SINTÉTICOS en testdata/codex/ (disciplina 11). Los tests de proceso corren en `testutil.Sandbox`, que deja
// `CODEX_HOME` vacía (M-9); los que necesitan Codex la apuntan a un temporal. Se compara `ExitCode()` (disciplina 4).

const lineaCodexVacia = "codex: respuestas 0 · eventos 0 · repetidas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0" +
	" · ficheros en formato anterior 0 · ficheros comprimidos 0"

func fixtureB5(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "codex", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return b
}

func escribir(t *testing.T, ruta string, contenido []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(ruta), 0o700); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	if err := os.WriteFile(ruta, contenido, 0o600); err != nil {
		t.Fatalf("precondición: %v", err)
	}
}

// logsDeClaude crea un `logs_root` con el fixture de Claude Code y escribe la `config.json` del sandbox que lo usa.
func logsDeClaude(t *testing.T, dataDir string) string {
	t.Helper()
	logs := filepath.Join(t.TempDir(), "logs")
	escribir(t, filepath.Join(logs, "claude.jsonl"), fixtureB5(t, "claude.jsonl"))
	escribir(t, filepath.Join(dataDir, "config.json"), []byte(`{"logs_root": "`+logs+`"}`+"\n"))
	return logs
}

// raizCodex crea una raíz de Codex con los fixtures dados y devuelve el CODEX_HOME (la raíz es `<CODEX_HOME>/sessions`).
func raizCodex(t *testing.T, fixtures map[string]string) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "codex")
	for destino, fixture := range fixtures {
		escribir(t, filepath.Join(home, "sessions", "2026", "10", "07", destino), fixtureB5(t, fixture))
	}
	return home
}

// cola lee los eventos encolados, por herramienta.
func cola(t *testing.T, dataDir, tool string) []event.Event {
	t.Helper()
	f, err := os.Open(transport.QueuePath(dataDir))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("leer la cola: %v", err)
	}
	defer func() { _ = f.Close() }()
	var evs []event.Event
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e event.Event
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			t.Fatalf("una línea de la cola no decodifica: %v", err)
		}
		if e.Tool == tool {
			evs = append(evs, e)
		}
	}
	return evs
}

// agenteCodex construye un `agent` en el propio proceso, con su raíz de Codex (el campo que `setup()` resuelve).
func agenteCodex(t *testing.T, dataDir, logs, codexRaiz string) *agent {
	t.Helper()
	cfg := config.Config{LogsRoot: logs}
	return &agent{dir: dataDir, cfg: cfg, ictx: newIngestContext("test", cfg, "sal-de-prueba", "maquina-de-prueba"),
		codexRaiz: codexRaiz}
}

// capturarStderr ejecuta f con os.Stderr redirigido y devuelve lo escrito.
func capturarStderr(t *testing.T, f func()) string {
	t.Helper()
	original := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	hecho := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		hecho <- string(b)
	}()
	f()
	os.Stderr = original
	_ = w.Close()
	return <-hecho
}

func sinPermisosPOSIX(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("depende de permisos POSIX de un usuario normal (plan R-8)")
	}
}

// (23) · SC-011, FR-021: sin raíz de Codex, el stderr de `--run` es byte a byte la referencia generada con el commit
// anterior (T030), con la ruta de datos sustituida por `<DATOS>`. Nace verde; lo valida m18.
func TestCodexRun_SinRaizEsLaSalidaDeLa040(t *testing.T) {
	referencia := string(fixtureB5(t, "referencia-run.stderr"))
	for _, c := range []struct{ hoja, codexHome string }{
		{"codex_home_vacia", ""},
		{"raiz_inexistente", filepath.Join(os.TempDir(), "permea-008-no-existe", "codex")},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			dataDir := testutil.Sandbox(t)
			logsDeClaude(t, dataDir)
			t.Setenv("CODEX_HOME", c.codexHome)
			codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
			if got := strings.ReplaceAll(stderr, dataDir, "<DATOS>"); codigo != 0 || got != referencia {
				t.Errorf("código %d; stderr:\n%s\nreferencia:\n%s", codigo, got, referencia)
			}
		})
	}
}

// (24) · SC-016, FR-026: con Codex y sin `~/.claude/projects`, `--run` encola los eventos de Codex y sale con 0.
func TestCodexRun_SinClaudeCode(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	t.Setenv("CODEX_HOME", raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"}))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "codex")); codigo != 0 || n != 2 {
		t.Errorf("código %d y %d eventos codex; se esperaba 0 y 2. stderr:\n%s", codigo, n, stderr)
	}
}

// (25) · SC-015, FR-025: con el forzado de D-008-P10 (A), los eventos de Codex están en la cola aunque `st.Save` falle,
// y la pasada siguiente los reencola con los mismos `event_id`.
func TestCodexGenerate_EncolaAntesDeGuardar(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := t.TempDir()
	codex := raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"})
	a := agenteCodex(t, dataDir, t.TempDir(), filepath.Join(codex, "sessions"))
	escribir(t, transport.QueuePath(dataDir), nil) // la cola existe: `Append` no necesita escribir en el directorio
	if err := os.Chmod(dataDir, 0o500); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(dataDir, 0o700) }()
	if _, _, err := a.generate(); err == nil {
		t.Fatalf("con el directorio de datos sin escritura, st.Save debía fallar y la pasada devolver error")
	}
	primera := cola(t, dataDir, "codex")
	if len(primera) != 2 {
		t.Fatalf("la pasada falló al guardar y la cola tiene %d eventos codex; se esperaban 2, encolados ANTES de guardar", len(primera))
	}
	if err := os.Chmod(dataDir, 0o700); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	if _, _, err := a.generate(); err != nil {
		t.Fatalf("con el estado sano: %v", err)
	}
	todos := cola(t, dataDir, "codex")
	if len(todos) != 4 || todos[2].EventID != primera[0].EventID || todos[3].EventID != primera[1].EventID {
		t.Errorf("la segunda pasada debía reencolar los mismos dos event_id; cola codex: %d eventos", len(todos))
	}
}

// (26) · FR-019, FR-021: la línea `codex:`, literal, justo tras el resumen de Claude Code, y «N eventos encolados» cuenta
// también los de Codex.
func TestCodexRun_LineaDeResumen(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	logsDeClaude(t, dataDir)
	t.Setenv("CODEX_HOME", raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"}))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	lineas := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	want := "codex: respuestas 2 · eventos 2 · repetidas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0" +
		" · ficheros en formato anterior 0 · ficheros comprimidos 0"
	if codigo != 0 || len(lineas) != 6 || lineas[1] != "4 eventos encolados en "+transport.QueuePath(dataDir) ||
		!strings.HasPrefix(lineas[3], "pasada: 0 mensajes que crecieron") || lineas[4] != want ||
		lineas[5] != "sync omitido: sin endpoint configurado" {
		t.Errorf("código %d; stderr:\n%s\nse esperaba «4 eventos encolados», los dos resúmenes de Claude Code y después:\n%s", codigo, stderr, want)
	}
}

// (27) · FR-019: en el demonio, la línea `codex:` sólo con novedades (respuestas, formato anterior o comprimidos).
func TestCodexDemonio_SoloConNovedades(t *testing.T) {
	t.Run("predicado", func(t *testing.T) {
		for _, c := range []struct {
			nombre string
			p      ingest.PasadaCodex
			want   bool
		}{
			{"nada", ingest.PasadaCodex{}, false},
			{"respuestas", ingest.PasadaCodex{Respuestas: 1}, true},
			{"formato_anterior", ingest.PasadaCodex{FormatoAnterior: 1}, true},
			{"comprimidos", ingest.PasadaCodex{Comprimidos: 1}, true},
		} {
			if got := c.p.HayNovedades(); got != c.want {
				t.Errorf("%s: HayNovedades() = %t; se esperaba %t", c.nombre, got, c.want)
			}
		}
	})
	t.Run("tick", func(t *testing.T) {
		codex := raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"})
		a := agenteCodex(t, t.TempDir(), t.TempDir(), filepath.Join(codex, "sessions"))
		primero := capturarStderr(t, func() { _ = a.tick() })
		segundo := capturarStderr(t, func() { _ = a.tick() })
		if !strings.Contains(primero, "codex: respuestas 2 ") || strings.Contains(segundo, "codex:") {
			t.Errorf("primer ciclo:\n%s\nsegundo ciclo, sin novedades:\n%s\nse esperaba la línea codex: sólo en el primero", primero, segundo)
		}
	})
}

// (28) · SC-005: una segunda `--run` sin cambios da 0 eventos de Codex.
func TestCodexRun_SegundaPasadaCero(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	t.Setenv("CODEX_HOME", raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"}))
	ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "codex")); n != 2 {
		t.Fatalf("precondición: la primera pasada encoló %d eventos codex; se esperaban 2", n)
	}
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "codex")); codigo != 0 || n != 2 || !strings.Contains(stderr, lineaCodexVacia) {
		t.Errorf("código %d, %d eventos codex en la cola; se esperaba 0, 2 y la línea vacía. stderr:\n%s", codigo, n, stderr)
	}
}

// (31) · SC-006, FR-016, E-1: ningún identificador del proveedor en la cola; en `state.json`, sólo el del nombre del
// fichero, dentro de su ruta (la excepción declarada).
func TestCodexRun_NadaDelProveedorViaja(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	home := raizCodex(t, map[string]string{"rollout-2026-10-07T00-00-00-CENTINELA-NOMBRE.jsonl": "centinelas.jsonl"})
	t.Setenv("CODEX_HOME", home)
	ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "codex")); n != 1 {
		t.Fatalf("precondición: %d eventos codex en la cola; se esperaba 1", n)
	}
	cola, err := os.ReadFile(transport.QueuePath(dataDir))
	if err != nil {
		t.Fatalf("leer la cola: %v", err)
	}
	if strings.Contains(string(cola), "CENTINELA") {
		t.Errorf("la cola lleva un identificador del proveedor:\n%s", cola)
	}
	estado, err := os.ReadFile(filepath.Join(dataDir, "state.json"))
	if err != nil {
		t.Fatalf("leer state.json: %v", err)
	}
	ruta := filepath.Join(home, "sessions", "2026", "10", "07", "rollout-2026-10-07T00-00-00-CENTINELA-NOMBRE.jsonl")
	if resto := strings.ReplaceAll(string(estado), ruta, ""); strings.Contains(resto, "CENTINELA") {
		t.Errorf("state.json lleva un identificador del proveedor fuera de la ruta del fichero:\n%s", estado)
	}
}

// (32) · FR-002, E-2: la ruta se resuelve en `setup()`, y si existe se mira en CADA pasada: el mismo `agent` ve una
// carpeta de Codex creada después.
func TestCodexActivacion_EnCadaPasada(t *testing.T) {
	_ = testutil.Sandbox(t)
	home := filepath.Join(t.TempDir(), "codex")
	t.Setenv("CODEX_HOME", home)
	a, err := setup()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, _, err := a.generate(); err != nil || a.codex != nil || len(cola(t, a.dir, "codex")) != 0 {
		t.Fatalf("sin carpeta: err %v, recuentos %v, cola codex %d; se esperaba inactivo", err, a.codex, len(cola(t, a.dir, "codex")))
	}
	escribir(t, filepath.Join(home, "sessions", "2026", "10", "07", "rollout-a.jsonl"), fixtureB5(t, "sesion.jsonl"))
	if _, _, err := a.generate(); err != nil || a.codex == nil || len(cola(t, a.dir, "codex")) != 2 {
		t.Errorf("con la carpeta creada después: err %v, recuentos %v, cola codex %d; se esperaban 2 eventos", err, a.codex, len(cola(t, a.dir, "codex")))
	}
}

// (33) · SC-018, FR-028: un fichero de Codex ilegible no tumba la pasada ni a Claude Code, no avanza su offset y se relee
// cuando vuelve a ser legible.
func TestCodexGenerate_FicheroIlegibleNoRompe(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := t.TempDir()
	logs := filepath.Join(t.TempDir(), "logs")
	escribir(t, filepath.Join(logs, "claude.jsonl"), fixtureB5(t, "claude.jsonl"))
	codex := raizCodex(t, map[string]string{"rollout-sano.jsonl": "otra.jsonl", "rollout-ilegible.jsonl": "sesion.jsonl"})
	ilegible := filepath.Join(codex, "sessions", "2026", "10", "07", "rollout-ilegible.jsonl")
	if err := os.Chmod(ilegible, 0o000); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(ilegible, 0o600) }()
	a := agenteCodex(t, dataDir, logs, filepath.Join(codex, "sessions"))

	t.Run("pasada", func(t *testing.T) {
		var err error
		stderr := capturarStderr(t, func() { _, _, err = a.generate() })
		_, sinEstado := os.Stat(filepath.Join(dataDir, "state.json"))
		if err != nil || !strings.Contains(stderr, "codex: fichero omitido: ") || sinEstado != nil ||
			len(cola(t, dataDir, "claude_code")) != 2 || len(cola(t, dataDir, "codex")) != 1 {
			t.Errorf("err %v, state.json %v, claude %d, codex %d; se esperaba sin error, el aviso, el estado guardado, 2 y 1. stderr:\n%s",
				err, sinEstado, len(cola(t, dataDir, "claude_code")), len(cola(t, dataDir, "codex")), stderr)
		}
	})
	t.Run("segunda_pasada", func(t *testing.T) {
		_ = capturarStderr(t, func() { _, _, _ = a.generate() })
		if n := len(cola(t, dataDir, "claude_code")); n != 2 {
			t.Errorf("la segunda pasada reencoló Claude Code: %d eventos claude_code; se esperaban 2", n)
		}
	})
	t.Run("se_relee", func(t *testing.T) {
		if err := os.Chmod(ilegible, 0o600); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		_ = capturarStderr(t, func() { _, _, _ = a.generate() })
		if n := len(cola(t, dataDir, "codex")); n != 3 {
			t.Errorf("con el fichero ya legible: %d eventos codex; se esperaban 3 (1 + sus 2, porque su offset no avanzó)", n)
		}
	})
}
