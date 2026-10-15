package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// P-012 B4 · Kimi Code en `--run` y en el demonio. Las sesiones son los ficheros SINTÉTICOS de
// `internal/ingest/testdata/kimi/ficheros/` (generar.sh), siempre copiados a un `t.TempDir()`: este paquete sólo tiene la referencia
// de SC-015 en `testdata/kimi/`. Reutiliza, sin tocarlos, los ayudantes de `codex_test.go`, `copilot_test.go` y los demás. Los tests
// de proceso corren en `testutil.Sandbox`, que deja `KIMI_CODE_HOME` vacía (M-3); los que necesitan Kimi Code la apuntan a un temporal.

// fixtureKimi es un fichero de `internal/ingest/testdata/kimi/ficheros/`.
func fixtureKimi(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "internal", "ingest", "testdata", "kimi", "ficheros", filepath.FromSlash(nombre)))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return b
}

// wireKimi es la ruta del `wire.jsonl` del agente A de la sesión S bajo la carpeta de Kimi Code `home`.
func wireKimi(home, s, a string) string {
	return filepath.Join(home, "sessions", "wd_sintetico_000000000001", s, "agents", a, "wire.jsonl")
}

// homeKimi crea una carpeta de Kimi Code con el índice sintético y, por cada clave «S/A», su `wire.jsonl` con el fixture dado.
// Devuelve la carpeta, que es el KIMI_CODE_HOME.
func homeKimi(t *testing.T, sesiones map[string]string) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "kimi-home")
	escribir(t, filepath.Join(home, "session_index.jsonl"), fixtureKimi(t, "indice/session_index.jsonl"))
	for sa, fixture := range sesiones {
		s, a, _ := strings.Cut(sa, "/")
		escribir(t, wireKimi(home, s, a), fixtureKimi(t, fixture))
	}
	return home
}

// agenteKimi es `agenteCodex`, sin Claude Code ni Codex, con la carpeta de Kimi Code (el campo que `setup()` resuelve).
func agenteKimi(t *testing.T, dataDir, home string) *agent {
	t.Helper()
	a := agenteCodex(t, dataDir, t.TempDir(), "")
	a.kimiRaiz = home
	return a
}

// lineaKimi es la línea de spec §Textos de 012, leída por programa, con los recuentos dados.
func lineaKimi(t *testing.T, recuentos ...any) string {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "specs", "012-lector-kimi", "spec.md"))
	if err != nil {
		t.Fatalf("precondición: spec: %v", err)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 1<<20)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "kimi_code: registros %d") {
			return fmt.Sprintf(sc.Text(), recuentos...)
		}
	}
	t.Fatal("precondición: la spec no tiene la línea del resumen de Kimi Code")
	return ""
}

// sesionesKimi son dos sesiones de Kimi Code: `dos_registros` (2 eventos) y el fork `un_forked` (1 evento y 2 heredados). Su resumen
// es «registros 5 · eventos 3 · heredados 2».
var sesionesKimi = map[string]string{
	"session_sintetica-0001/main": "dos_registros.jsonl",
	"session_sintetica-0002/main": "fork/un_forked.jsonl",
}

// (30) · SC-015, FR-023: sin Kimi Code, el stderr de `--run` sobre las otras cinco herramientas es byte a byte la referencia
// generada con el binario de `e9e5449` (T026), con la ruta de datos sustituida por `<DATOS>`. NACE VERDE; lo valida m41.
func TestKimiRun_SinKimiEsLaSalidaDeLa080(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "kimi", "referencia-run.stderr"))
	if err != nil || len(b) == 0 {
		t.Fatalf("precondición: la referencia: %v", err)
	}
	for _, c := range []struct {
		hoja    string
		conHome bool
	}{
		{"kimi_home_vacia", false},
		{"sin_raiz", true},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			dataDir := testutil.Sandbox(t)
			todasLasHerramientas(t, dataDir)
			t.Setenv("COPILOT_HOME", homeCopilot(t, dosSesiones))
			if c.conHome {
				t.Setenv("KIMI_CODE_HOME", t.TempDir()) // existe, pero sin `sessions/`
			} else {
				t.Setenv("KIMI_CODE_HOME", "") // `<hogar del sandbox>/.kimi-code`, que no existe
			}
			codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
			if got := strings.ReplaceAll(stderr, dataDir, "<DATOS>"); codigo != 0 || got != string(b) {
				t.Errorf("código %d; stderr:\n%s\nreferencia:\n%s", codigo, got, b)
			}
		})
	}
}

// (31) · SC-019, FR-025: con sólo Kimi Code, `--run` encola sus eventos y sale con 0.
func TestKimiRun_SoloKimi(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	t.Setenv("KIMI_CODE_HOME", homeKimi(t, sesionesKimi))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "kimi_code")); codigo != 0 || n != 3 {
		t.Errorf("código %d y %d eventos kimi_code; se esperaba 0 y 3. stderr:\n%s", codigo, n, stderr)
	}
}

// (32) · SC-019, FR-024: con el forzado A (D-012-P10), los eventos de Kimi Code están en la cola aunque `st.Save` falle, y la
// pasada siguiente los reencola con los mismos `event_id`.
func TestKimiGenerate_EncolaAntesDeGuardar(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := t.TempDir()
	a := agenteKimi(t, dataDir, homeKimi(t, sesionesKimi))
	escribir(t, transport.QueuePath(dataDir), nil) // la cola existe: `Append` no necesita escribir en el directorio
	if err := os.Chmod(dataDir, 0o500); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(dataDir, 0o700) }()
	if _, _, err := a.generate(); err == nil {
		t.Fatalf("con el directorio de datos sin escritura, st.Save debía fallar y la pasada devolver error")
	}
	primera := cola(t, dataDir, "kimi_code")
	if len(primera) != 3 {
		t.Fatalf("la pasada falló al guardar y la cola tiene %d eventos kimi_code; se esperaban 3, encolados ANTES de guardar", len(primera))
	}
	if err := os.Chmod(dataDir, 0o700); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	if _, _, err := a.generate(); err != nil {
		t.Fatalf("con el estado sano: %v", err)
	}
	todos := cola(t, dataDir, "kimi_code")
	if len(todos) != 6 {
		t.Fatalf("la segunda pasada debía reencolar los tres; cola kimi_code: %d eventos", len(todos))
	}
	for i := range primera {
		if todos[3+i].EventID != primera[i].EventID {
			t.Errorf("evento %d reencolado con otro event_id", i)
		}
	}
}

// (33) · FR-021, FR-023, FR-024: la línea `kimi_code:`, literal, justo tras la de Copilot CLI; «N eventos encolados» cuenta también
// Kimi Code; y en la cola, los eventos de Copilot CLI van ANTES que los de Kimi Code.
func TestKimiRun_LineaDeResumen(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	todasLasHerramientas(t, dataDir)
	t.Setenv("COPILOT_HOME", homeCopilot(t, dosSesiones))
	t.Setenv("KIMI_CODE_HOME", homeKimi(t, sesionesKimi))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	lineas := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	t.Run("linea", func(t *testing.T) {
		want := lineaKimi(t, 5, 3, 2, 0, 0, 0, 0, 0)
		if codigo != 0 || len(lineas) != 10 || !strings.HasPrefix(lineas[7], "copilot_cli: registros 3 ") || lineas[8] != want ||
			lineas[9] != "sync omitido: sin endpoint configurado" {
			t.Errorf("código %d; stderr:\n%s\nse esperaban los resúmenes de las otras cinco y después:\n%s", codigo, stderr, want)
		}
	})
	t.Run("total", func(t *testing.T) {
		// 11 de las otras cinco (la referencia de SC-015) + 3 de Kimi Code.
		if len(lineas) < 2 || lineas[1] != "14 eventos encolados en "+transport.QueuePath(dataDir) {
			t.Errorf("stderr:\n%s\nse esperaba «14 eventos encolados»", stderr)
		}
	})
	t.Run("orden_cola", func(t *testing.T) {
		b, err := os.ReadFile(transport.QueuePath(dataDir))
		if err != nil {
			t.Fatalf("leer la cola: %v", err)
		}
		texto := string(b)
		ultimoCopilot, primerKimi := strings.LastIndex(texto, `"tool":"copilot_cli"`), strings.Index(texto, `"tool":"kimi_code"`)
		if ultimoCopilot < 0 || primerKimi < 0 || ultimoCopilot > primerKimi {
			t.Errorf("en la cola, el último copilot_cli está en %d y el primer kimi_code en %d; se esperaba Copilot CLI antes", ultimoCopilot, primerKimi)
		}
	})
}

// (34) · FR-021: en el demonio, la línea `kimi_code:` sólo con novedades; una segunda pasada sin cambios calla.
func TestKimiDemonio_SoloConNovedades(t *testing.T) {
	a := agenteKimi(t, t.TempDir(), homeKimi(t, sesionesKimi))
	primero := capturarStderr(t, func() { _ = a.tick() })
	segundo := capturarStderr(t, func() { _ = a.tick() })
	if !strings.Contains(primero, lineaKimi(t, 5, 3, 2, 0, 0, 0, 0, 0)) || strings.Contains(segundo, "kimi_code:") {
		t.Errorf("primer ciclo:\n%s\nsegundo ciclo, sin novedades:\n%s\nse esperaba la línea kimi_code: sólo en el primero", primero, segundo)
	}
}

// ═══ Nacen verdes, tras el verde de T028 (T030) ════════════════════════════════════════════════════════════════════

// (35) · SC-011, FR-019: nada del proveedor ni del contenido en la cola; en `state.json`, los centinelas de la ruta sólo DENTRO de su
// clave. Centinelas en S, A, el nombre de la `wd_*`, `workDir`, `sessionDir`, `context.append_message`, `profile.bind.systemPrompt`,
// `llm.request.systemPrompt`, `subagent.completed.resultSummary` y los campos de texto del registro. NACE VERDE; lo valida m49.
func TestKimiRun_NadaDelProveedorViaja(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	home := filepath.Join(t.TempDir(), "kimi-home")
	registro, err := os.ReadFile(filepath.Join("..", "..", "internal", "ingest", "testdata", "kimi", "lineas", "centinelas.jsonl"))
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	lineas := `{"type":"metadata","protocol_version":"1.5","created_at":1}` + "\n" +
		`{"type":"profile.bind","agentId":"main","systemPrompt":"CENTINELA-profile","time":1}` + "\n" +
		`{"type":"context.append_message","agentId":"main","message":{"role":"user","content":[{"type":"text","text":"CENTINELA-mensaje"}]},"time":2}` + "\n" +
		`{"type":"llm.request","agentId":"main","kind":"loop","systemPrompt":"CENTINELA-llm","messageCount":1,"time":3}` + "\n" +
		string(registro) +
		`{"type":"subagent.completed","subagentId":"agent-0","resultSummary":"CENTINELA-resumen","usage":{"inputOther":1,"output":1,"inputCacheRead":1,"inputCacheCreation":0},"time":4}` + "\n"
	ruta := filepath.Join(home, "sessions", "wd_CENTINELA-wd", "CENTINELA-S", "agents", "CENTINELA-A", "wire.jsonl")
	escribir(t, ruta, []byte(lineas))
	escribir(t, filepath.Join(home, "session_index.jsonl"),
		[]byte(`{"sessionId":"CENTINELA-S","sessionDir":"/sintetico/CENTINELA-sessionDir","workDir":"/sintetico/CENTINELA-workDir"}`+"\n"))
	t.Setenv("KIMI_CODE_HOME", home)
	ejecutar(t, 30*time.Second, "--run")
	evs := cola(t, dataDir, "kimi_code")
	if len(evs) != 1 || evs[0].ProjectRef == "" {
		t.Fatalf("precondición: %d eventos kimi_code en la cola, con proyecto; se esperaba 1", len(evs))
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

// (36) · SC-013, FR-006: `--run` deja iguales las md5 de todos los ficheros de la carpeta de Kimi Code y su listado: ni un byte, ni
// un fichero nuevo. La carpeta es `raiz/` *(B1: `.index-*`, `.bak`, `.tmp.*`, `imported/`, carpetas con punto y un `wire.jsonl` demasiado
// hondo)*, con registros en TODOS sus `wire.jsonl*`: si se leyera uno que no toca, el resumen pasaría de 7. NACE VERDE; lo valida M-B4a.
func TestKimiRun_SoloLectura(t *testing.T) {
	_ = testutil.Sandbox(t)
	home := filepath.Join(t.TempDir(), "kimi-home")
	origen := filepath.Join("..", "..", "internal", "ingest", "testdata", "kimi", "raiz")
	err := filepath.WalkDir(origen, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(origen, p)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		switch {
		case filepath.Base(filepath.Dir(p)) == "agent-0":
			b = fixtureKimi(t, "subagente/agent-0.jsonl")
		case strings.HasPrefix(filepath.Base(p), "wire.jsonl"):
			b = fixtureKimi(t, "dos_registros.jsonl")
		}
		escribir(t, filepath.Join(home, rel), b)
		return nil
	})
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	escribir(t, filepath.Join(home, "session_index.jsonl"), fixtureKimi(t, "indice/session_index.jsonl"))
	t.Setenv("KIMI_CODE_HOME", home)
	antes := huellaArbol(t, home)
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if despues := huellaArbol(t, home); codigo != 0 || despues != antes || !strings.Contains(stderr, lineaKimi(t, 7, 7, 0, 0, 0, 0, 0, 0)) {
		t.Errorf("código %d; antes:\n%s\ndespués:\n%s\nstderr:\n%s", codigo, antes, despues, stderr)
	}
}

// (37) · FR-002, P-2: la carpeta se resuelve en `setup()`, y si existe `sessions/` se mira en CADA pasada: el mismo `agent` ve la
// carpeta de sesiones creada después. NACE VERDE; lo valida m45 (y m41, que lo daría por activo sin carpeta).
func TestKimiActivacion_EnCadaPasada(t *testing.T) {
	_ = testutil.Sandbox(t)
	home := filepath.Join(t.TempDir(), "kimi-home")
	t.Setenv("KIMI_CODE_HOME", home)
	a, err := setup()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, _, err := a.generate(); err != nil || a.kimi != nil || len(cola(t, a.dir, "kimi_code")) != 0 {
		t.Fatalf("sin la carpeta: err %v, recuentos %v, cola kimi_code %d; se esperaba inactivo", err, a.kimi, len(cola(t, a.dir, "kimi_code")))
	}
	escribir(t, wireKimi(home, "session_sintetica-0001", "main"), fixtureKimi(t, "dos_registros.jsonl"))
	if _, _, err := a.generate(); err != nil || a.kimi == nil || len(cola(t, a.dir, "kimi_code")) != 2 {
		t.Errorf("con la carpeta creada después: err %v, recuentos %v, cola kimi_code %d; se esperaban 2 eventos",
			err, a.kimi, len(cola(t, a.dir, "kimi_code")))
	}
}

// (38) · SC-018, FR-026: un `wire.jsonl` ilegible junto a fixtures sanos de los cinco lectores: salen los sanos, el aviso no lleva su
// ruta, `state.json` se guarda sin avanzar ese offset, y cuando vuelve a ser legible, su registro sale. NACE VERDE; lo validan m46 y m47.
func TestKimiRun_FicheroIlegibleNoRompe(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := testutil.Sandbox(t)
	todasLasHerramientas(t, dataDir)
	t.Setenv("COPILOT_HOME", homeCopilot(t, dosSesiones))
	home := homeKimi(t, sesionesKimi)
	t.Setenv("KIMI_CODE_HOME", home)
	ilegible := wireKimi(home, "session_sintetica-0002", "main")
	if err := os.Chmod(ilegible, 0o000); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(ilegible, 0o600) }()

	t.Run("pasada", func(t *testing.T) {
		codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
		_, sinEstado := os.Stat(filepath.Join(dataDir, "state.json"))
		if codigo != 0 || !strings.Contains(stderr, "\n13 eventos encolados en ") || !strings.Contains(stderr, "kimi_code: fichero omitido: ") ||
			strings.Contains(stderr, home) || sinEstado != nil || len(cola(t, dataDir, "kimi_code")) != 2 {
			t.Errorf("código %d, state.json %v, kimi_code %d; se esperaba 0, «13 eventos encolados» (11 + 2), el aviso sin la ruta, el estado y 2. stderr:\n%s",
				codigo, sinEstado, len(cola(t, dataDir, "kimi_code")), stderr)
		}
	})
	t.Run("se_relee", func(t *testing.T) {
		if err := os.Chmod(ilegible, 0o600); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		_, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
		if n := len(cola(t, dataDir, "kimi_code")); n != 3 {
			t.Errorf("con el fichero ya legible: %d eventos kimi_code; se esperaban 3 (2 + el suyo, porque su offset no avanzó). stderr:\n%s", n, stderr)
		}
	})
}

// (39) · P-13, SC-014: con el índice ilegible, se avisa sin la ruta y NO se lee ningún `wire.jsonl` en esa pasada (0 eventos y ningún
// offset en `state.json`); Claude Code sigue. Cuando vuelve a ser legible, salen. NACE VERDE; lo valida m48.
func TestKimiGenerate_IndiceIlegible(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := t.TempDir()
	logs := filepath.Join(t.TempDir(), "logs")
	escribir(t, filepath.Join(logs, "claude.jsonl"), fixtureB5(t, "claude.jsonl"))
	home := homeKimi(t, sesionesKimi)
	indice := filepath.Join(home, "session_index.jsonl")
	if err := os.Chmod(indice, 0o000); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(indice, 0o600) }()
	a := agenteCodex(t, dataDir, logs, "")
	a.kimiRaiz = home

	t.Run("pasada", func(t *testing.T) {
		var err error
		stderr := capturarStderr(t, func() { _, _, err = a.generate() })
		estado, errEstado := os.ReadFile(filepath.Join(dataDir, "state.json"))
		if err != nil || !strings.Contains(stderr, "kimi_code: fichero omitido: ") || strings.Contains(stderr, home) || errEstado != nil ||
			strings.Contains(string(estado), home) || len(cola(t, dataDir, "claude_code")) != 2 || len(cola(t, dataDir, "kimi_code")) != 0 {
			t.Errorf("err %v, state.json %v, claude %d, kimi_code %d; se esperaba sin error, el aviso sin la ruta, el estado sin Kimi, 2 y 0. stderr:\n%s\nstate.json:\n%s",
				err, errEstado, len(cola(t, dataDir, "claude_code")), len(cola(t, dataDir, "kimi_code")), stderr, estado)
		}
	})
	t.Run("legible", func(t *testing.T) {
		if err := os.Chmod(indice, 0o600); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		_ = capturarStderr(t, func() { _, _, _ = a.generate() })
		if n := len(cola(t, dataDir, "kimi_code")); n != 3 {
			t.Errorf("con el índice ya legible: %d eventos kimi_code; se esperaban 3", n)
		}
	})
}

// ═══ P-012 B5 · `--scan` DE UN `wire.jsonl` DE KIMI CODE ═══════════════════════════════════════════════════════════════
//
// La línea `evento:` es la compartida *(`imprimirEventos`)*, y el resumen, el de spec §Textos de 012, leído por programa. Sin línea de
// control: Kimi Code no da créditos (D-5).

// scanKimi copia un fichero de `ficheros/` a `<tmp>/<s>/agents/<a>/wire.jsonl` y le pasa `--scan`.
func scanKimi(t *testing.T, fixture, s, a string) (codigo int, stdout, stderr, fichero string) {
	t.Helper()
	fichero = filepath.Join(t.TempDir(), s, "agents", a, "wire.jsonl")
	escribir(t, fichero, fixtureKimi(t, fixture))
	codigo, stdout, stderr, _ = ejecutar(t, 30*time.Second, "--scan", fichero)
	return codigo, stdout, stderr, fichero
}

// (40) · FR-022, P-12, SC-024 *(forma)*: `--scan` de un `wire.jsonl` escribe las líneas `evento:` y el resumen, sin línea de control, y
// no escribe nada en disco; el de un fork aplica la frontera `forked`: 1 evento y «heredados 2».
func TestKimiScan_EventosYResumen(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	codigo, stdout, stderr, fichero := scanKimi(t, "dos_registros.jsonl", "session_sintetica-0001", "main")
	t.Run("salida", func(t *testing.T) {
		lineas := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
		if codigo != 0 || len(lineas) != 2 || !strings.HasPrefix(lineas[0], "evento: tool=kimi_code model=kimi-k2.7-code ") ||
			!strings.HasPrefix(lineas[1], "evento: tool=kimi_code model=kimi-k2.7-code ") ||
			!strings.Contains(stderr, "2 eventos generados (dry-run, nada transmitido)\n"+lineaKimi(t, 2, 2, 0, 0, 0, 0, 0, 0)+"\n") {
			t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nse esperaban 2 `evento:`, sin línea de control, y el resumen", codigo, stdout, stderr)
		}
	})
	t.Run("disco", func(t *testing.T) {
		junto, err := os.ReadDir(filepath.Dir(fichero))
		_, sinEstado := os.Stat(filepath.Join(dataDir, "state.json"))
		_, sinCola := os.Stat(transport.QueuePath(dataDir))
		if err != nil || len(junto) != 1 || sinEstado == nil || sinCola == nil {
			t.Errorf("junto al fichero: %d entradas (%v); state.json: %v; cola: %v; se esperaba nada en disco", len(junto), err, sinEstado, sinCola)
		}
	})
	t.Run("fork", func(t *testing.T) {
		codigo, stdout, stderr, _ := scanKimi(t, "fork/un_forked.jsonl", "session_sintetica-0002", "main")
		if codigo != 0 || strings.Count(stdout, "evento: tool=kimi_code ") != 1 ||
			!strings.Contains(stderr, "1 eventos generados (dry-run, nada transmitido)\n"+lineaKimi(t, 3, 1, 2, 0, 0, 0, 0, 0)+"\n") {
			t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nse esperaba 1 evento y «heredados 2»", codigo, stdout, stderr)
		}
	})
	t.Run("subagente", func(t *testing.T) {
		// SC-024: el `wire.jsonl` de `agent-0`, 3 eventos; el 1.º, con el vector «subagente» del contrato, tal cual. NACE VERDE (tras el
		// verde de T035); lo valida m53.
		const subagente = "70ee1df9071f3a49800d62483b50dcb3"
		codigo, stdout, stderr, _ := scanKimi(t, "subagente/agent-0.jsonl", "session_sintetica-0001", "agent-0")
		lineas := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
		if codigo != 0 || len(lineas) != 3 || !strings.HasSuffix(lineas[0], " event_id="+subagente) ||
			!strings.Contains(stderr, lineaKimi(t, 3, 3, 0, 0, 0, 0, 0, 0)+"\n") {
			t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nse esperaban 3 eventos, el 1.º con event_id=%s", codigo, stdout, stderr, subagente)
		}
	})
}

// (41) · FR-023: `--scan` de un fichero de Claude Code, Codex, Gemini, OpenCode o Copilot CLI da lo de la 0.8.0: las referencias de 008,
// 009 y 010 para los tres primeros, el contraste de OpenCode y la línea de control de Copilot CLI, y nada de Kimi Code. NACE VERDE; lo
// valida m50 *(la hoja `claude`: los otros cuatro formatos se reconocen ANTES que Kimi, P-12)*.
func TestKimiScan_LosDemasComoLa080(t *testing.T) {
	for _, c := range []struct{ hoja, fichero, referencia string }{
		{"claude", filepath.Join("testdata", "codex", "claude.jsonl"), filepath.Join("codex", "referencia-scan-claude")},
		{"codex", filepath.Join("testdata", "codex", "sesion.jsonl"), filepath.Join("gemini", "referencia-scan-codex")},
		{"gemini", filepath.Join("testdata", "gemini", "sesion.jsonl"), filepath.Join("opencode", "referencia-scan-gemini")},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			_ = testutil.Sandbox(t)
			wantOut, errOut := os.ReadFile(filepath.Join("testdata", c.referencia+".stdout"))
			wantErr, errErr := os.ReadFile(filepath.Join("testdata", c.referencia+".stderr"))
			if errOut != nil || errErr != nil || len(wantOut) == 0 {
				t.Fatalf("precondición: la referencia %s: %v, %v", c.referencia, errOut, errErr)
			}
			codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "--scan", c.fichero)
			if codigo != 0 || stdout != string(wantOut) || stderr != string(wantErr) {
				t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nno es la salida de la 0.8.0", codigo, stdout, stderr)
			}
		})
	}
	t.Run("opencode", func(t *testing.T) {
		_ = testutil.Sandbox(t)
		base := filepath.Join(t.TempDir(), "opencode.db")
		escribir(t, base, fixtureOpenCode(t, "sana.db"))
		codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "--scan", base)
		if codigo != 0 || !strings.Contains(stdout, "opencode (contraste, no se envía)") || strings.Contains(stdout+stderr, "kimi_code") {
			t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nse esperaba la salida de OpenCode y nada de Kimi Code", codigo, stdout, stderr)
		}
	})
	t.Run("copilot", func(t *testing.T) {
		_ = testutil.Sandbox(t)
		codigo, stdout, stderr, _ := scanCopilot(t, "dos_llamadas.jsonl")
		if codigo != 0 || !strings.Contains(stdout, "copilot_cli (control, no se envía)") || strings.Contains(stdout+stderr, "kimi_code") {
			t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nse esperaba la salida de Copilot CLI y nada de Kimi Code", codigo, stdout, stderr)
		}
	})
}

// (42) · P-12: un fichero cuya 1.ª línea es `metadata` SIN `created_at` no es de Kimi Code: sigue el camino de Claude Code, como en la
// 0.8.0. NACE VERDE; lo valida m52.
func TestKimiScan_SinCreatedAtNoEsKimi(t *testing.T) {
	_ = testutil.Sandbox(t)
	fichero := filepath.Join(t.TempDir(), "session_sintetica-0001", "agents", "main", "wire.jsonl")
	registro, err := os.ReadFile(filepath.Join("..", "..", "internal", "ingest", "testdata", "kimi", "lineas", "evento.jsonl"))
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	escribir(t, fichero, append([]byte(`{"type":"metadata","protocol_version":"1.5"}`+"\n"), registro...))
	codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "--scan", fichero)
	if codigo != 0 || strings.Contains(stdout+stderr, "kimi_code") {
		t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nse esperaba el camino de Claude Code, sin nada de Kimi Code", codigo, stdout, stderr)
	}
}

// (43) · P-12, D-3: S y A salen de la ruta *(`<S>/agents/<A>/wire.jsonl`)*: el 1.er registro de `session_sintetica-0001`/`main` da el
// `event_id` del vector 1 del contrato, tal cual.
func TestKimiScan_SesionYAgenteDeLaRuta(t *testing.T) {
	_ = testutil.Sandbox(t)
	const vector1 = "e1ad45be19ab01a906be10cf22f4e75c" // contracts/event-id-kimi.md, 1.er registro
	codigo, stdout, stderr, _ := scanKimi(t, "dos_registros.jsonl", "session_sintetica-0001", "main")
	lineas := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if codigo != 0 || !strings.HasSuffix(lineas[0], " event_id="+vector1) {
		t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nse esperaba el 1.er evento con event_id=%s", codigo, stdout, stderr, vector1)
	}
}
