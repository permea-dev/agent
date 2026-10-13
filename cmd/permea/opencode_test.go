package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/state"
	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// P-010 B6 · OpenCode en `--run` y en el demonio. Bases SINTÉTICAS de testdata/opencode/ (`generar.sh`), siempre copiadas a un
// `t.TempDir()` (PARA 12), y ninguna con filas en curso (PARA 11). Reutiliza, sin tocarlos, los ayudantes de `codex_test.go` y
// `gemini_test.go`. Los tests de proceso corren en `testutil.Sandbox`, que deja `XDG_DATA_HOME` y `OPENCODE_DB` vacías (M-3);
// los que necesitan OpenCode apuntan `XDG_DATA_HOME` a un temporal.

func fixtureOpenCode(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "opencode", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return b
}

// datosOpenCode crea `<datos>/opencode/<destino>` con las bases dadas y devuelve `<datos>`, el valor de XDG_DATA_HOME.
func datosOpenCode(t *testing.T, bases map[string]string) string {
	t.Helper()
	datos := filepath.Join(t.TempDir(), "datos")
	for destino, fixture := range bases {
		escribir(t, filepath.Join(datos, "opencode", destino), fixtureOpenCode(t, fixture))
	}
	return datos
}

// agenteOpenCode es `agenteCodex`, sin Claude Code ni Codex, con la carpeta de OpenCode (el campo que `setup()` resuelve).
func agenteOpenCode(t *testing.T, dataDir, datos string) *agent {
	t.Helper()
	a := agenteCodex(t, dataDir, t.TempDir(), "")
	a.opencodeDatos = filepath.Join(datos, "opencode")
	return a
}

// lineaOpenCode es la línea de spec §Textos, leída por programa, con los recuentos dados.
func lineaOpenCode(t *testing.T, recuentos ...any) string {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "specs", "010-lector-opencode", "spec.md"))
	if err != nil {
		t.Fatalf("precondición: spec: %v", err)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 1<<20)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "opencode: llamadas %d") {
			return fmt.Sprintf(sc.Text(), recuentos...)
		}
	}
	t.Fatal("precondición: la spec no tiene la línea del resumen de OpenCode")
	return ""
}

// (41) · SC-016, FR-026: sin OpenCode, el stderr de `--run` sobre Claude Code, Codex y Gemini es byte a byte la referencia
// generada con el binario de `4c6cafc` (T035), con la ruta de datos sustituida por `<DATOS>`. Nace verde; lo valida m35.
func TestOpenCodeRun_SinOpenCodeEsLaSalidaDeLa060(t *testing.T) {
	referencia := string(fixtureOpenCode(t, "referencia-run.stderr"))
	for _, c := range []struct {
		hoja      string
		datosVivo bool
	}{
		{"opencode_db_vacia", true},
		{"xdg_vacia", false},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			dataDir := testutil.Sandbox(t)
			logsDeClaude(t, dataDir)
			t.Setenv("CODEX_HOME", raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"}))
			t.Setenv("GEMINI_CLI_HOME", raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba))
			t.Setenv("OPENCODE_DB", "")
			if c.datosVivo {
				t.Setenv("XDG_DATA_HOME", t.TempDir()) // existe, pero sin `opencode/`
			} else {
				t.Setenv("XDG_DATA_HOME", "") // `<hogar del sandbox>/.local/share/opencode`, que no existe
			}
			codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
			if got := strings.ReplaceAll(stderr, dataDir, "<DATOS>"); codigo != 0 || got != referencia {
				t.Errorf("código %d; stderr:\n%s\nreferencia:\n%s", codigo, got, referencia)
			}
		})
	}
}

// (42) · SC-021, FR-028: con sólo OpenCode, `--run` encola sus eventos y sale con 0.
func TestOpenCodeRun_SoloOpenCode(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	t.Setenv("XDG_DATA_HOME", datosOpenCode(t, map[string]string{"opencode.db": "sana.db"}))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "opencode")); codigo != 0 || n != 2 {
		t.Errorf("código %d y %d eventos opencode; se esperaba 0 y 2. stderr:\n%s", codigo, n, stderr)
	}
}

// (43) · SC-021, FR-027: con el forzado A (D-010-P15), los eventos de OpenCode están en la cola aunque `st.Save` falle, y la
// pasada siguiente los reencola con los mismos `event_id`.
func TestOpenCodeGenerate_EncolaAntesDeGuardar(t *testing.T) {
	sinPermisosPOSIX(t)
	dataDir := t.TempDir()
	a := agenteOpenCode(t, dataDir, datosOpenCode(t, map[string]string{"opencode.db": "sana.db"}))
	escribir(t, transport.QueuePath(dataDir), nil) // la cola existe: `Append` no necesita escribir en el directorio
	if err := os.Chmod(dataDir, 0o500); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = os.Chmod(dataDir, 0o700) }()
	if _, _, err := a.generate(); err == nil {
		t.Fatalf("con el directorio de datos sin escritura, st.Save debía fallar y la pasada devolver error")
	}
	primera := cola(t, dataDir, "opencode")
	if len(primera) != 2 {
		t.Fatalf("la pasada falló al guardar y la cola tiene %d eventos opencode; se esperaban 2, encolados ANTES de guardar", len(primera))
	}
	if err := os.Chmod(dataDir, 0o700); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	if _, _, err := a.generate(); err != nil {
		t.Fatalf("con el estado sano: %v", err)
	}
	todos := cola(t, dataDir, "opencode")
	if len(todos) != 4 || todos[2].EventID != primera[0].EventID || todos[3].EventID != primera[1].EventID {
		t.Errorf("la segunda pasada debía reencolar los mismos dos event_id; cola opencode: %d eventos", len(todos))
	}
}

// (44) · FR-024, FR-026: la línea `opencode:`, literal, justo tras la de Gemini, y «N eventos encolados» cuenta también OpenCode.
func TestOpenCodeRun_LineaDeResumen(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	logsDeClaude(t, dataDir)
	t.Setenv("CODEX_HOME", raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"}))
	t.Setenv("GEMINI_CLI_HOME", raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba))
	t.Setenv("XDG_DATA_HOME", datosOpenCode(t, map[string]string{"opencode.db": "sana.db"}))
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	lineas := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	want := lineaOpenCode(t, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0)
	if codigo != 0 || len(lineas) != 8 || lineas[1] != "8 eventos encolados en "+transport.QueuePath(dataDir) ||
		!strings.HasPrefix(lineas[5], "gemini: respuestas 3 ") || lineas[6] != want || lineas[7] != "sync omitido: sin endpoint configurado" {
		t.Errorf("código %d; stderr:\n%s\nse esperaba «8 eventos encolados», los resúmenes de Claude Code, Codex y Gemini, y después:\n%s",
			codigo, stderr, want)
	}
}

// (45) · FR-024: en el demonio, la línea `opencode:` sólo con novedades; una segunda pasada sin cambios calla.
func TestOpenCodeDemonio_SoloConNovedades(t *testing.T) {
	a := agenteOpenCode(t, t.TempDir(), datosOpenCode(t, map[string]string{"opencode.db": "sana.db"}))
	primero := capturarStderr(t, func() { _ = a.tick() })
	segundo := capturarStderr(t, func() { _ = a.tick() })
	if !strings.Contains(primero, "opencode: llamadas 2 ") || strings.Contains(segundo, "opencode:") {
		t.Errorf("primer ciclo:\n%s\nsegundo ciclo, sin novedades:\n%s\nse esperaba la línea opencode: sólo en el primero", primero, segundo)
	}
}

// (46) · SC-003: una segunda `--run` sin cambios da 0 eventos de OpenCode.
func TestOpenCodeRun_SegundaPasadaCero(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	t.Setenv("XDG_DATA_HOME", datosOpenCode(t, map[string]string{"opencode.db": "sana.db"}))
	ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "opencode")); n != 2 {
		t.Fatalf("precondición: la primera pasada encoló %d eventos opencode; se esperaban 2", n)
	}
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "opencode")); codigo != 0 || n != 2 || !strings.Contains(stderr, lineaOpenCode(t, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)) {
		t.Errorf("código %d, %d eventos opencode en la cola; se esperaba 0, 2 y la línea vacía. stderr:\n%s", codigo, n, stderr)
	}
}

// ═══ Nacen verdes, tras el verde de T037 (T039) ════════════════════════════════════════════════════════════════════

// (47) · SC-014, FR-020: ningún identificador del proveedor ni texto de la base en la cola; en `state.json`, los centinelas
// de la ruta (el nombre de la base) sólo DENTRO de su clave. Nace verde; lo valida m42.
func TestOpenCodeRun_NadaDelProveedorViaja(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	datos := datosOpenCode(t, map[string]string{"opencode-CENTINELA-canal.db": "centinelas.db"})
	t.Setenv("XDG_DATA_HOME", datos)
	ejecutar(t, 30*time.Second, "--run")
	if n := len(cola(t, dataDir, "opencode")); n != 1 {
		t.Fatalf("precondición: %d eventos opencode en la cola; se esperaba 1", n)
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
	ruta := filepath.Join(datos, "opencode", "opencode-CENTINELA-canal.db")
	if !strings.Contains(string(estado), `"`+ruta+`":`) {
		t.Fatalf("precondición: state.json no tiene la ruta de la base como clave:\n%s", estado)
	}
	if resto := strings.ReplaceAll(string(estado), ruta, ""); strings.Contains(resto, "CENTINELA") {
		t.Errorf("state.json lleva un centinela fuera de la ruta de la base:\n%s", estado)
	}
}

// (48) · SC-006, FR-006, FR-007: `--run` sobre I (la base viva sintética de `internal/sqlitero`, con su `-wal`) deja iguales
// las md5 de la base y del `-wal`, y el listado de la carpeta: ni `-shm` ni ningún fichero nuevo. Nace verde; lo valida M-B6a.
// Las filas sin `completed` de I se crearon en 1970: son «abandonadas» desde la primera pasada, y su resultado no depende
// de cuándo corra el test (PARA 11).
func TestOpenCodeRun_SoloLectura(t *testing.T) {
	_ = testutil.Sandbox(t)
	datos := filepath.Join(t.TempDir(), "datos")
	carpeta := filepath.Join(datos, "opencode")
	for _, n := range []string{"opencode.db", "opencode.db-wal"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "internal", "sqlitero", "testdata", "i", n))
		if err != nil {
			t.Fatalf("precondición: %v", err)
		}
		escribir(t, filepath.Join(carpeta, n), b)
	}
	t.Setenv("XDG_DATA_HOME", datos)
	antes := huellaCarpeta(t, carpeta)
	codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run")
	if despues := huellaCarpeta(t, carpeta); codigo != 0 || despues != antes || !strings.Contains(stderr, "opencode: llamadas ") {
		t.Errorf("código %d; antes:\n%s\ndespués:\n%s\nstderr:\n%s", codigo, antes, despues, stderr)
	}
}

// huellaCarpeta es el listado de una carpeta con el md5 de cada fichero.
func huellaCarpeta(t *testing.T, carpeta string) string {
	t.Helper()
	entradas, err := os.ReadDir(carpeta)
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	var b strings.Builder
	for _, e := range entradas {
		contenido, err := os.ReadFile(filepath.Join(carpeta, e.Name()))
		if err != nil {
			t.Fatalf("leer: %v", err)
		}
		fmt.Fprintf(&b, "%s %x\n", e.Name(), md5.Sum(contenido))
	}
	return b.String()
}

// (49) · FR-002, P-10: la ruta se resuelve en `setup()`, y si existe se mira en CADA pasada: el mismo `agent` ve la carpeta
// de OpenCode creada después. Nace verde; lo valida m39.
func TestOpenCodeActivacion_EnCadaPasada(t *testing.T) {
	_ = testutil.Sandbox(t)
	datos := filepath.Join(t.TempDir(), "datos")
	t.Setenv("XDG_DATA_HOME", datos)
	a, err := setup()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, _, err := a.generate(); err != nil || a.opencode != nil || len(cola(t, a.dir, "opencode")) != 0 {
		t.Fatalf("sin la carpeta: err %v, recuentos %v, cola opencode %d; se esperaba inactivo", err, a.opencode, len(cola(t, a.dir, "opencode")))
	}
	escribir(t, filepath.Join(datos, "opencode", "opencode.db"), fixtureOpenCode(t, "sana.db"))
	if _, _, err := a.generate(); err != nil || a.opencode == nil || len(cola(t, a.dir, "opencode")) != 2 {
		t.Errorf("con la carpeta creada después: err %v, recuentos %v, cola opencode %d; se esperaban 2 eventos",
			err, a.opencode, len(cola(t, a.dir, "opencode")))
	}
}

// (50) · SC-020, FR-029: una base corrupta no tumba la pasada ni a Claude Code, Codex y Gemini; el aviso lleva su nombre y no
// su ruta; su frontera no avanza; y cuando vuelve a estar sana, sus llamadas salen. Nace verde; lo validan m40 y m41.
func TestOpenCodeGenerate_BaseCorruptaNoRompe(t *testing.T) {
	dataDir := t.TempDir()
	logs := filepath.Join(t.TempDir(), "logs")
	escribir(t, filepath.Join(logs, "claude.jsonl"), fixtureB5(t, "claude.jsonl"))
	codex := raizCodex(t, map[string]string{"rollout-a.jsonl": "sesion.jsonl"})
	home := raizGemini(t, map[string]string{"s.jsonl": "sesion.jsonl"}, textoRaizDePrueba)
	datos := datosOpenCode(t, map[string]string{"opencode.db": "sana.db", "opencode-roto.db": "corrupta.db"})
	rota := filepath.Join(datos, "opencode", "opencode-roto.db")
	a := agenteGemini(t, dataDir, logs, filepath.Join(codex, "sessions"), home)
	a.opencodeDatos = filepath.Join(datos, "opencode")

	t.Run("pasada", func(t *testing.T) {
		var err error
		stderr := capturarStderr(t, func() { _, _, err = a.generate() })
		_, sinEstado := os.Stat(filepath.Join(dataDir, "state.json"))
		if err != nil || !strings.Contains(stderr, "opencode: base omitida: opencode-roto.db: ") || strings.Contains(stderr, datos) ||
			sinEstado != nil || len(cola(t, dataDir, "claude_code")) != 2 || len(cola(t, dataDir, "codex")) != 2 ||
			len(cola(t, dataDir, "gemini")) != 2 || len(cola(t, dataDir, "opencode")) != 2 {
			t.Errorf("err %v, state.json %v, claude %d, codex %d, gemini %d, opencode %d; se esperaba sin error, el aviso con el "+
				"nombre y sin la ruta, el estado, y 2, 2, 2 y 2. stderr:\n%s", err, sinEstado, len(cola(t, dataDir, "claude_code")),
				len(cola(t, dataDir, "codex")), len(cola(t, dataDir, "gemini")), len(cola(t, dataDir, "opencode")), stderr)
		}
	})
	t.Run("se_relee", func(t *testing.T) {
		st, err := state.Load(filepath.Join(dataDir, "state.json"))
		if err != nil {
			t.Fatalf("leer state.json: %v", err)
		}
		if e, hay := st.Files[rota]; hay {
			t.Errorf("la base corrupta tiene entrada en state.json, %+v: su frontera avanzó", e)
		}
		escribir(t, rota, fixtureOpenCode(t, "otra.db"))
		_ = capturarStderr(t, func() { _, _, _ = a.generate() })
		if n := len(cola(t, dataDir, "opencode")); n != 3 {
			t.Errorf("con la base ya sana: %d eventos opencode; se esperaban 3 (2 + el suyo)", n)
		}
	})
}

// ═══ P-010 B7 · `--scan` DE UNA BASE DE OPENCODE ═══════════════════════════════════════════════════════════════════
//
// La línea `evento:` es la aprobada en 008 (`textoAprobado`), y el resumen y el contraste, los de spec §Textos de 010, leídos
// por programa.

// formatoOpenCode es la línea de spec §Textos que empieza por `prefijo`, leída por programa y sin rellenar.
func formatoOpenCode(t *testing.T, prefijo string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "specs", "010-lector-opencode", "spec.md"))
	if err != nil {
		t.Fatalf("precondición: spec: %v", err)
	}
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, prefijo) {
			return l
		}
	}
	t.Fatalf("precondición: la spec no tiene la línea %q", prefijo)
	return ""
}

// (51) · FR-025, SC-001 y SC-002 en forma: `--scan` de una base da una línea `evento:` aprobada por llamada, la línea
// `opencode:` en stderr y la de contraste al final de stdout; y nada en disco: el listado del directorio de datos y el de
// la carpeta de la base, iguales antes y después.
func TestOpenCodeScan_EventosResumenYContraste(t *testing.T) {
	dataDir := testutil.Sandbox(t)
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	carpeta := t.TempDir()
	base := filepath.Join(carpeta, "opencode.db")
	escribir(t, base, fixtureOpenCode(t, "contraste.db"))
	antesDatos, antesCarpeta := huellaCarpeta(t, dataDir), huellaCarpeta(t, carpeta)
	codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "--scan", base)
	if codigo != 0 {
		t.Fatalf("código %d; stderr:\n%s", codigo, stderr)
	}
	lineas := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	t.Run("lineas_evento", func(t *testing.T) {
		patron := patronDe(textoAprobado(t, "**Línea de `--scan` para Codex**"))
		want := []string{"opencode modelo-sintetico-opencode 100 27 4 30 0.0000 false",
			"opencode modelo-sintetico-opencode 200 13 0 0 0.0000 false"}
		if len(lineas) != 3 {
			t.Fatalf("%d líneas en stdout; se esperaban 2 `evento:` y el contraste:\n%s", len(lineas), stdout)
		}
		for i, l := range lineas[:2] {
			m := patron.FindStringSubmatch(l)
			if m == nil || strings.Join(m[1:9], " ") != want[i] {
				t.Errorf("línea %d:\n%s\nse esperaba la aprobada con %q", i, l, want[i])
			}
		}
	})
	t.Run("resumen", func(t *testing.T) {
		if want := lineaOpenCode(t, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0); !strings.Contains(stderr, want+"\n") {
			t.Errorf("stderr no lleva la línea de spec §Textos %q:\n%s", want, stderr)
		}
	})
	t.Run("contraste", func(t *testing.T) {
		patron := regexp.MustCompile(strings.ReplaceAll(patronDe(formatoOpenCode(t, "opencode (contraste")).String(),
			`%\.6f`, `([0-9]+\.[0-9]{6})`))
		if !patron.MatchString(lineas[len(lineas)-1]) {
			t.Errorf("la última línea de stdout no es la de contraste de spec §Textos:\n%s\npatrón %s", stdout, patron)
		}
	})
	t.Run("disco", func(t *testing.T) {
		despuesDatos, despuesCarpeta := huellaCarpeta(t, dataDir), huellaCarpeta(t, carpeta)
		if despuesDatos != antesDatos || despuesCarpeta != antesCarpeta {
			t.Errorf("--scan dejó algo en disco; datos antes:\n%s\ndespués:\n%s\ncarpeta antes:\n%s\ndespués:\n%s",
				antesDatos, despuesDatos, antesCarpeta, despuesCarpeta)
		}
		for _, f := range []string{"state.json", "queue.jsonl"} {
			if _, err := os.Stat(filepath.Join(dataDir, f)); !os.IsNotExist(err) {
				t.Errorf("--scan dejó %s en el directorio de datos (err %v)", f, err)
			}
		}
	})
}

// (52) · FR-026: sin una base de OpenCode, `--scan` de Claude Code, de Codex y de Gemini da la salida de la 0.6.0, byte a byte.
// Las referencias son las del binario de `4c6cafc`: las de Claude Code y Codex, idénticas a las de 008 y 009 (`cmp`), se
// reutilizan; la de Gemini es nueva. Nace verde; lo valida m44.
func TestOpenCodeScan_LosDemasComoLa060(t *testing.T) {
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
				t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nno es la salida de la 0.6.0", codigo, stdout, stderr)
			}
		})
	}
}

// (53) · SC-002 en forma, D-1, D-2: el contraste es la Σ de `$.cost` de las `assistant` y, fuera de los mensajes, los totales
// de `session_v2` menos su Σ, con la salida como `output + reasoning`. `contraste.db` lleva Δ`reasoning` ≠ 0 (H13): entrada
// 338 − 300, salida (33 + 14) − (30 + 10) y coste 0.017174 − 0.0168. Lo valida m46.
func TestOpenCodeScan_ContrasteFueraDeLosMensajes(t *testing.T) {
	_ = testutil.Sandbox(t)
	base := filepath.Join(t.TempDir(), "opencode.db")
	escribir(t, base, fixtureOpenCode(t, "contraste.db"))
	codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "--scan", base)
	lineas := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	want := fmt.Sprintf(formatoOpenCode(t, "opencode (contraste"), 0.0168, 38, 7, 0.000374)
	if got := lineas[len(lineas)-1]; codigo != 0 || got != want {
		t.Errorf("código %d; la última línea de stdout:\n%s\nse esperaba:\n%s\nstderr:\n%s", codigo, got, want, stderr)
	}
}

// (54) · FR-026: en `--scan`, un `.jsonl` vacío y otro de 10 bytes, más cortos que la firma de SQLite, dan el código, el stdout
// y el stderr de la 0.6.0, byte a byte. Las referencias son las del binario de `4c6cafc`, que salió con 0. Nace verde; lo valida m43.
func TestOpenCodeScan_FicherosCortosComoLa060(t *testing.T) {
	for _, c := range []struct{ hoja, sufijo string }{{"vacio", "vacio"}, {"diez_bytes", "10"}} {
		t.Run(c.hoja, func(t *testing.T) {
			_ = testutil.Sandbox(t)
			fichero := filepath.Join("testdata", "opencode", "corto-"+c.sufijo+".jsonl")
			wantOut := fixtureOpenCode(t, "referencia-scan-corto-"+c.sufijo+".stdout")
			wantErr := fixtureOpenCode(t, "referencia-scan-corto-"+c.sufijo+".stderr")
			if n := len(fixtureOpenCode(t, "corto-"+c.sufijo+".jsonl")); n >= 16 || len(wantErr) == 0 {
				t.Fatalf("precondición: %d bytes (< 16) y la referencia de stderr no vacía (%d bytes)", n, len(wantErr))
			}
			codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "--scan", fichero)
			if codigo != 0 || stdout != string(wantOut) || stderr != string(wantErr) {
				t.Errorf("código %d; stdout:\n%s\nstderr:\n%s\nno es la salida de la 0.6.0", codigo, stdout, stderr)
			}
		})
	}
}
