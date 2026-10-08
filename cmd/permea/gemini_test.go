package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// P-009 B5 · Gemini CLI en `--run` y en el demonio. Fixtures SINTÉTICOS en testdata/gemini/ (disciplina 11). Reutiliza, sin
// tocarlos, los ayudantes de `codex_test.go`. Los tests de proceso corren en `testutil.Sandbox`, que deja
// `GEMINI_CLI_HOME` vacía (M-9); los que necesitan Gemini la apuntan a un temporal.

// slugDePrueba es la carpeta `<slug>` de los fixtures; textoRaizDePrueba, el texto de su `.project_root`.
const (
	slugDePrueba      = "proyecto-sintetico"
	textoRaizDePrueba = "/tmp/proyecto-gemini-sintetico"
)

func fixtureGemini(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "gemini", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return b
}

// raizGemini crea `<H>/.gemini/tmp/<slug>/chats/<destino>` con los fixtures dados, y el `.project_root` con `raiz` si no es
// vacío. Devuelve H, el valor de GEMINI_CLI_HOME (la raíz es `<H>/.gemini`).
func raizGemini(t *testing.T, ficheros map[string]string, raiz string) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "gemini-home")
	slug := filepath.Join(home, ".gemini", "tmp", slugDePrueba)
	for destino, fixture := range ficheros {
		escribir(t, filepath.Join(slug, "chats", destino), fixtureGemini(t, fixture))
	}
	if raiz != "" {
		escribir(t, filepath.Join(slug, ".project_root"), []byte(raiz+"\n"))
	}
	return home
}

// agenteGemini es `agenteCodex` con la raíz de Gemini (el campo que `setup()` resuelve).
func agenteGemini(t *testing.T, dataDir, logs, codexRaiz, geminiHome string) *agent {
	t.Helper()
	a := agenteCodex(t, dataDir, logs, codexRaiz)
	a.geminiRaiz = filepath.Join(geminiHome, ".gemini")
	return a
}

// lineaGemini es la línea aprobada de spec §Textos aprobados, leída por programa, con los recuentos dados.
func lineaGemini(t *testing.T, recuentos ...any) string {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "specs", "009-lector-gemini", "spec.md"))
	if err != nil {
		t.Fatalf("precondición: spec: %v", err)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "gemini: respuestas %d") {
			return fmt.Sprintf(sc.Text(), recuentos...)
		}
	}
	t.Fatal("precondición: la spec no tiene la línea del resumen de Gemini")
	return ""
}

// (28) · SC-012, FR-023: sin raíz de Gemini, el stderr de `--run` sobre Claude Code y Codex es byte a byte la referencia
// generada con el binario de `15ce93b` (T030), con la ruta de datos sustituida por `<DATOS>`. Nace verde; lo valida m26.
func TestGeminiRun_SinRaizEsLaSalidaDeLa050(t *testing.T) {
	referencia := string(fixtureGemini(t, "referencia-run.stderr"))
	for _, c := range []struct{ hoja, geminiHome string }{
		{"gemini_cli_home_vacia", ""},
		{"sin_tmp", filepath.Join(os.TempDir(), "permea-009-no-existe", "gemini-home")},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			dataDir := testutil.Sandbox(t)
			logsDeClaude(t, dataDir)
			t.Setenv("CODEX_HOME", raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"}))
			t.Setenv("GEMINI_CLI_HOME", c.geminiHome)
			codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
			if got := strings.ReplaceAll(stderr, dataDir, "<DATOS>"); codigo != 0 || got != referencia {
				t.Errorf("código %d; stderr:\n%s\nreferencia:\n%s", codigo, got, referencia)
			}
		})
	}
}

// (29) · SC-017, FR-026: con sólo la raíz de Gemini, `--run` encola sus eventos y sale con 0.
func TestGeminiRun_SoloGemini(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	t.Setenv("GEMINI_CLI_HOME", raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "gemini")); codigo != 0 || n != 2 {
		t.Errorf("código %d y %d eventos gemini; se esperaba 0 y 2. stderr:\n%s", codigo, n, stderr)
	}
}

// (30) · SC-017, FR-024: con el forzado A (D-009-P12), los eventos de Gemini están en la cola aunque `st.Save` falle, y la
// pasada siguiente los reencola con los mismos `event_id`.
func TestGeminiGenerate_EncolaAntesDeGuardar(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := t.TempDir()
	home := raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba)
	a := agenteGemini(t, dataDir, t.TempDir(), "", home)
	escribir(t, transport.QueuePath(dataDir), nil) // la cola existe: `Append` no necesita escribir en el directorio
	if err := os.Chmod(dataDir, 0o500); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(dataDir, 0o700) }()
	if _, _, err := a.generate(); err == nil {
		t.Fatalf("con el directorio de datos sin escritura, st.Save debía fallar y la pasada devolver error")
	}
	primera := cola(t, dataDir, "gemini")
	if len(primera) != 2 {
		t.Fatalf("la pasada falló al guardar y la cola tiene %d eventos gemini; se esperaban 2, encolados ANTES de guardar", len(primera))
	}
	if err := os.Chmod(dataDir, 0o700); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	if _, _, err := a.generate(); err != nil {
		t.Fatalf("con el estado sano: %v", err)
	}
	todos := cola(t, dataDir, "gemini")
	if len(todos) != 4 || todos[2].EventID != primera[0].EventID || todos[3].EventID != primera[1].EventID {
		t.Errorf("la segunda pasada debía reencolar los mismos dos event_id; cola gemini: %d eventos", len(todos))
	}
}

// (31) · FR-021, FR-023: la línea `gemini:`, literal, justo tras la de Codex, y «N eventos encolados» cuenta también Gemini.
func TestGeminiRun_LineaDeResumen(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	logsDeClaude(t, dataDir)
	t.Setenv("CODEX_HOME", raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"}))
	t.Setenv("GEMINI_CLI_HOME", raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	lineas := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	want := lineaGemini(t, 3, 2, 1, 0, 0, 0, 0, 0)
	if codigo != 0 || len(lineas) != 7 || lineas[1] != "6 eventos encolados en "+transport.QueuePath(dataDir) ||
		!strings.HasPrefix(lineas[4], "codex: respuestas 2 ") || lineas[5] != want || lineas[6] != "sync omitido: sin endpoint configurado" {
		t.Errorf("código %d; stderr:\n%s\nse esperaba «6 eventos encolados», los resúmenes de Claude Code y de Codex, y después:\n%s",
			codigo, stderr, want)
	}
}

// (32) · FR-021: en el demonio, la línea `gemini:` sólo con novedades (respuestas o formato anterior).
func TestGeminiDemonio_SoloConNovedades(t *testing.T) {
	t.Run("predicado", func(t *testing.T) {
		if (&ingest.PasadaGemini{}).HayNovedades() || !(&ingest.PasadaGemini{Respuestas: 1}).HayNovedades() ||
			!(&ingest.PasadaGemini{FormatoAnterior: 1}).HayNovedades() {
			t.Error("HayNovedades: se esperaba false sin nada, y true con respuestas o con formato anterior")
		}
	})
	t.Run("tick", func(t *testing.T) {
		home := raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba)
		a := agenteGemini(t, t.TempDir(), t.TempDir(), "", home)
		primero := capturarStderr(t, func() { _ = a.tick() })
		segundo := capturarStderr(t, func() { _ = a.tick() })
		if !strings.Contains(primero, "gemini: respuestas 3 ") || strings.Contains(segundo, "gemini:") {
			t.Errorf("primer ciclo:\n%s\nsegundo ciclo, sin novedades:\n%s\nse esperaba la línea gemini: sólo en el primero", primero, segundo)
		}
	})
}

// (33) · SC-003: una segunda `--run` sin cambios da 0 eventos de Gemini.
func TestGeminiRun_SegundaPasadaCero(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	t.Setenv("GEMINI_CLI_HOME", raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba))
	ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "gemini")); n != 2 {
		t.Fatalf("precondición: la primera pasada encoló %d eventos gemini; se esperaban 2", n)
	}
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "gemini")); codigo != 0 || n != 2 || !strings.Contains(stderr, lineaGemini(t, 0, 0, 0, 0, 0, 0, 0, 0)) {
		t.Errorf("código %d, %d eventos gemini en la cola; se esperaba 0, 2 y la línea vacía. stderr:\n%s", codigo, n, stderr)
	}
}

// (34) · SC-006, FR-017: ningún identificador del proveedor ni el texto de `.project_root` en la cola; en `state.json`, los
// centinelas de ruta (nombre del fichero y carpeta del subagente) sólo DENTRO de la clave (la excepción declarada).
func TestGeminiRun_NadaDelProveedorViaja(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	destino := filepath.Join("CENTINELA-PADRE", "session-CENTINELA-FICHERO.jsonl")
	home := raizGemini(t, map[string]string{destino: "centinelas.jsonl"}, "/tmp/CENTINELA-RAIZ")
	t.Setenv("GEMINI_CLI_HOME", home)
	ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "gemini")); n != 1 {
		t.Fatalf("precondición: %d eventos gemini en la cola; se esperaba 1", n)
	}
	enCola, err := os.ReadFile(transport.QueuePath(dataDir))
	if err != nil {
		t.Fatalf("leer la cola: %v", err)
	}
	if strings.Contains(string(enCola), "CENTINELA") {
		t.Errorf("la cola lleva un identificador del proveedor o una ruta:\n%s", enCola)
	}
	estado, err := os.ReadFile(filepath.Join(dataDir, "state.json"))
	if err != nil {
		t.Fatalf("leer state.json: %v", err)
	}
	ruta := filepath.Join(home, ".gemini", "tmp", slugDePrueba, "chats", destino)
	if !strings.Contains(string(estado), `"`+ruta+`":`) {
		t.Fatalf("precondición: state.json no tiene la ruta del fichero como clave:\n%s", estado)
	}
	if resto := strings.ReplaceAll(string(estado), ruta, ""); strings.Contains(resto, "CENTINELA") {
		t.Errorf("state.json lleva un centinela fuera de la ruta del fichero:\n%s", estado)
	}
}

// (35) · FR-002: la ruta se resuelve en `setup()`, y si `<raíz>/tmp` existe se mira en CADA pasada.
func TestGeminiActivacion_EnCadaPasada(t *testing.T) {
	_ = testutil.Sandbox(t)
	home := filepath.Join(t.TempDir(), "gemini-home")
	t.Setenv("GEMINI_CLI_HOME", home)
	a, err := setup()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, _, err := a.generate(); err != nil || a.gemini != nil || len(cola(t, a.dir, "gemini")) != 0 {
		t.Fatalf("sin `.gemini/tmp`: err %v, recuentos %v, cola gemini %d; se esperaba inactivo", err, a.gemini, len(cola(t, a.dir, "gemini")))
	}
	escribir(t, filepath.Join(home, ".gemini", "tmp", slugDePrueba, "chats", "s.jsonl"), fixtureGemini(t, "sesion.jsonl"))
	if _, _, err := a.generate(); err != nil || a.gemini == nil || len(cola(t, a.dir, "gemini")) != 2 {
		t.Errorf("con `.gemini/tmp` creada después: err %v, recuentos %v, cola gemini %d; se esperaban 2 eventos",
			err, a.gemini, len(cola(t, a.dir, "gemini")))
	}
}

// (36) · SC-016, FR-025: un fichero de Gemini ilegible no tumba la pasada ni a Claude Code ni a Codex, no avanza su offset y
// se relee cuando vuelve a ser legible.
func TestGeminiGenerate_FicheroIlegibleNoRompe(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := t.TempDir()
	logs := filepath.Join(t.TempDir(), "logs")
	escribir(t, filepath.Join(logs, "claude.jsonl"), fixtureB5(t, "claude.jsonl"))
	codex := raizCodex(t, map[string]string{"rollout-a.jsonl": "otra.jsonl"})
	home := raizGemini(t, map[string]string{"s-sano.jsonl": "sesion.jsonl", "t-ilegible.jsonl": "otra.jsonl"}, textoRaizDePrueba)
	ilegible := filepath.Join(home, ".gemini", "tmp", slugDePrueba, "chats", "t-ilegible.jsonl")
	if err := os.Chmod(ilegible, 0o000); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(ilegible, 0o600) }()
	a := agenteGemini(t, dataDir, logs, filepath.Join(codex, "sessions"), home)

	t.Run("pasada", func(t *testing.T) {
		var err error
		stderr := capturarStderr(t, func() { _, _, err = a.generate() })
		_, sinEstado := os.Stat(filepath.Join(dataDir, "state.json"))
		if err != nil || !strings.Contains(stderr, "gemini: fichero omitido: ") || sinEstado != nil ||
			len(cola(t, dataDir, "claude_code")) != 2 || len(cola(t, dataDir, "codex")) != 1 || len(cola(t, dataDir, "gemini")) != 2 {
			t.Errorf("err %v, state.json %v, claude %d, codex %d, gemini %d; se esperaba sin error, el aviso, el estado, 2, 1 y 2. stderr:\n%s",
				err, sinEstado, len(cola(t, dataDir, "claude_code")), len(cola(t, dataDir, "codex")), len(cola(t, dataDir, "gemini")), stderr)
		}
	})
	t.Run("segunda_pasada", func(t *testing.T) {
		_ = capturarStderr(t, func() { _, _, _ = a.generate() })
		if c, x := len(cola(t, dataDir, "claude_code")), len(cola(t, dataDir, "codex")); c != 2 || x != 1 {
			t.Errorf("la segunda pasada reencoló: claude_code %d, codex %d; se esperaban 2 y 1", c, x)
		}
	})
	t.Run("se_relee", func(t *testing.T) {
		if err := os.Chmod(ilegible, 0o600); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		_ = capturarStderr(t, func() { _, _, _ = a.generate() })
		if n := len(cola(t, dataDir, "gemini")); n != 3 {
			t.Errorf("con el fichero ya legible: %d eventos gemini; se esperaban 3 (2 + el suyo, porque su offset no avanzó)", n)
		}
	})
}

// ═══ P-009 B6 · `--scan` DE UNA SESIÓN DE GEMINI CLI ═════════════════════════════════════════════
//
// La línea `evento:` es la aprobada en 008 (`textoAprobado` de `codex_test.go` la lee por programa), y el resumen, la de
// 009 (`lineaGemini`).

// (37) · FR-022, SC-015: una línea `evento:` aprobada por evento, la línea `gemini:`, el proyecto de la posición del
// fichero, y nada en disco.
func TestGeminiScan_EventosYResumen(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	home := raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba)
	fichero := filepath.Join(home, ".gemini", "tmp", slugDePrueba, "chats", "s.jsonl")
	codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "--scan", fichero)
	if codigo != 0 {
		t.Fatalf("código %d; stderr:\n%s", codigo, stderr)
	}
	lineas := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	t.Run("lineas_evento", func(t *testing.T) {
		patron := patronDe(textoAprobado(t, "**Línea de `--scan` para Codex**"))
		want := []string{"gemini modelo-sintetico-gemini 100 10 0 0 0.0000 false", "gemini modelo-sintetico-gemini 200 12 0 0 0.0000 false"}
		if len(lineas) != 2 {
			t.Fatalf("%d líneas en stdout; se esperaban 2:\n%s", len(lineas), stdout)
		}
		for i, l := range lineas {
			m := patron.FindStringSubmatch(l)
			if m == nil || strings.Join(m[1:9], " ") != want[i] {
				t.Errorf("línea %d:\n%s\nse esperaba la aprobada con %q", i, l, want[i])
			}
		}
	})
	t.Run("resumen", func(t *testing.T) {
		if want := lineaGemini(t, 3, 2, 1, 0, 0, 0, 0, 0); !strings.Contains(stderr, want+"\n") {
			t.Errorf("stderr no lleva la línea aprobada %q:\n%s", want, stderr)
		}
	})
	t.Run("proyecto", func(t *testing.T) {
		if want := "project_ref=" + project.Derivar(textoRaizDePrueba, "dry-run-salt")[:8] + "… "; !strings.Contains(lineas[0], want) {
			t.Errorf("la línea no lleva el proyecto de su `<slug>` (%q):\n%s", want, lineas[0])
		}
	})
	t.Run("sin_forma", func(t *testing.T) {
		suelto := filepath.Join(t.TempDir(), "s.jsonl")
		escribir(t, suelto, fixtureGemini(t, "sesion.jsonl"))
		codigo, stdout, _, _ := ejecutar(t, 30*time.Second, "--scan", suelto)
		if codigo != 0 || strings.Count(stdout, "project_ref= event_id=") != 2 {
			t.Errorf("fuera de `<slug>/chats/`, código %d; se esperaban 2 eventos con project_ref vacío:\n%s", codigo, stdout)
		}
	})
	t.Run("nada_en_disco", func(t *testing.T) {
		for _, f := range []string{"state.json", "queue.jsonl"} {
			if _, err := os.Stat(filepath.Join(dataDir, f)); !os.IsNotExist(err) {
				t.Errorf("--scan dejó %s en el directorio de datos (err %v)", f, err)
			}
		}
	})
}

// (38) · FR-023: `--scan` de Claude Code y de Codex da la salida de la 0.5.0, byte a byte. Las referencias son las del binario
// de `15ce93b` (Claude Code, la misma de la 0.4.0). Nace verde; lo valida m36.
func TestGeminiScan_ClaudeYCodexComoLa050(t *testing.T) {
	for _, c := range []struct{ hoja, fichero, stdout, stderr string }{
		{"claude", filepath.Join("testdata", "codex", "claude.jsonl"),
			filepath.Join("codex", "referencia-scan-claude.stdout"), filepath.Join("codex", "referencia-scan-claude.stderr")},
		{"codex", filepath.Join("testdata", "codex", "sesion.jsonl"),
			filepath.Join("gemini", "referencia-scan-codex.stdout"), filepath.Join("gemini", "referencia-scan-codex.stderr")},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			_ = testutil.Sandbox(t)
			codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "--scan", c.fichero)
			wantOut, _ := os.ReadFile(filepath.Join("testdata", c.stdout))
			wantErr, _ := os.ReadFile(filepath.Join("testdata", c.stderr))
			if codigo != 0 || stdout != string(wantOut) || stderr != string(wantErr) || len(wantOut) == 0 {
				t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nno es la salida de la 0.5.0", codigo, stdout, stderr)
			}
		})
	}
}
