package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ 013 B7 · LOS SEIS LECTORES Y LA FRONTERA — (22) a (26) de plan §Bloques ═════════════════════════════════════════
//
// Por el binario de prueba, sobre las plantillas de `testdata/lista/` con `@CASA@` → el hogar del sandbox: así cada carpeta
// existe o no, y tiene o no `.git`, de verdad. La tabla esperada se arma en el test: la identidad es `event.Ref(sal, raíz real)`, y
// el resto, literal; `formatear`, ya probada contra spec §Textos en (13), sólo le da forma.

// hogarConLista monta las plantillas en un `SandboxConSemillas`, con `@CASA@` → su hogar, y crea las carpetas: dos repositorios
// *(`<hogar>/dev/proyecto-ejemplo`, con `src` y `docs`, y `<hogar>/dev/otro-proyecto`)*, una carpeta sin `.git`
// *(`<hogar>/pruebas/tmp-1`)*, y ninguna `<hogar>/dev/viejo`, que «ya no existe». OpenCode es la copia de `sana.db`
// *(`/sintetico/sana`, que tampoco existe)*. Devuelve el `DataDir` y el hogar.
func hogarConLista(t *testing.T) (dataDir, hogar string) {
	t.Helper()
	dataDir = testutil.SandboxConSemillas(t)
	hogar = os.Getenv("HOME")
	plantillas := filepath.Join("testdata", "lista", "plantillas")
	err := filepath.WalkDir(plantillas, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(plantillas, p)
		if err != nil {
			return err
		}
		partes := strings.SplitN(rel, string(filepath.Separator), 2)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		escribir(t, filepath.Join(hogar, carpetasDeLaPlantilla[partes[0]], partes[1]), []byte(strings.ReplaceAll(string(b), "@CASA@", hogar)))
		return nil
	})
	if err != nil {
		t.Fatalf("precondición: plantillas: %v", err)
	}
	db, err := os.ReadFile(filepath.Join("testdata", "opencode", "sana.db"))
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	escribir(t, filepath.Join(hogar, ".local", "share", "opencode", "opencode.db"), db)
	for _, d := range []string{"dev/proyecto-ejemplo/.git", "dev/proyecto-ejemplo/src", "dev/proyecto-ejemplo/docs", "dev/otro-proyecto/.git",
		"pruebas/tmp-1"} {
		if err := os.MkdirAll(filepath.Join(hogar, filepath.FromSlash(d)), 0o700); err != nil {
			t.Fatalf("precondición: %v", err)
		}
	}
	return dataDir, hogar
}

// tablaEsperada es la salida de `permea project list` sobre `hogarConLista`, armada a mano: una fila por identidad, con los eventos,
// las herramientas y las fechas de las plantillas *(generar.sh)* y de `sana.db`.
func tablaEsperada(t *testing.T, hogar string) string {
	t.Helper()
	sal, _ := testutil.SemillasDeLaLineaBase(t)
	d := func(s string) time.Time {
		v, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatalf("precondición: %v", err)
		}
		return v
	}
	nueva := func(ruta, marca string, eventos int, desde, hasta string, herramientas ...string) fila {
		return fila{identidad: event.Ref(sal, ruta), ruta: ruta, marca: marca, herramientas: herramientas, eventos: eventos,
			desde: d(desde), hasta: d(hasta)}
	}
	filas := []fila{
		nueva(filepath.Join(hogar, "dev", "proyecto-ejemplo"), "", 8, "2025-10-09", "2026-10-07", "Claude Code", "Codex", "Kimi Code CLI"),
		nueva(filepath.Join(hogar, "dev", "otro-proyecto"), "", 4, "2026-10-08", "2026-10-09", "Gemini CLI", "GitHub Copilot CLI"),
		nueva(filepath.Join(hogar, "dev", "viejo"), marcaYaNoExiste, 2, "2026-10-05", "2026-10-06", "Claude Code"),
		nueva("/sintetico/sana", marcaYaNoExiste, 2, "2025-10-09", "2025-10-09", "OpenCode"),
		nueva(filepath.Join(hogar, "pruebas", "tmp-1"), marcaSinRepositorio, 1, "2026-10-20", "2026-10-20", "Claude Code"),
	}
	sort.Slice(filas, func(i, j int) bool {
		if filas[i].eventos != filas[j].eventos {
			return filas[i].eventos > filas[j].eventos
		}
		return filas[i].identidad < filas[j].identidad
	})
	return formatear(filas, hogar)
}

// (22) · SC-002, SC-007, FR-009, FR-015, FR-016, FR-023: con los seis lectores, la salida exacta, con sus marcas; y un fichero
// ilegible de Codex da su aviso por stderr, la tabla igual y ninguna línea de resumen.
func TestListaFrontera_LosSeisLectores(t *testing.T) {
	t.Run("tabla", func(t *testing.T) {
		_, hogar := hogarConLista(t)
		codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "project", "list")
		if quiere := tablaEsperada(t, hogar); codigo != 0 || stdout != quiere || stderr != "" {
			t.Errorf("código %d, stderr %q; stdout:\n%s\nse esperaba 0, nada por stderr y:\n%s", codigo, stderr, stdout, quiere)
		}
	})
	t.Run("aviso", func(t *testing.T) {
		sinPermisosPOSIX(t)
		_, hogar := hogarConLista(t)
		ilegible := filepath.Join(hogar, ".codex", "sessions", "2026", "10", "07", "rollout-ilegible.jsonl")
		escribir(t, ilegible, []byte("{}\n"))
		if err := os.Chmod(ilegible, 0o000); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		defer func() { _ = os.Chmod(ilegible, 0o600) }()
		codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "project", "list")
		if quiere := tablaEsperada(t, hogar); codigo != 0 || stdout != quiere || !strings.Contains(stderr, "codex: fichero omitido: ") {
			t.Errorf("código %d, stderr %q; stdout:\n%s\nse esperaba 0, el aviso de Codex y:\n%s", codigo, stderr, stdout, quiere)
		}
		for _, resumen := range []string{"codex: respuestas", "pasada:", "eventos encolados"} {
			if strings.Contains(stderr, resumen) {
				t.Errorf("stderr lleva una línea de resumen de `--run` (%q): %q", resumen, stderr)
			}
		}
	})
}

// identidadesDeLaTabla son los 12 caracteres de cada identidad que enseña la tabla.
func identidadesDeLaTabla(stdout string) map[string]bool {
	ids := map[string]bool{}
	for i, l := range strings.Split(stdout, "\n") {
		if l == "" {
			break
		}
		if i > 0 && strings.Contains(l, "…") {
			ids[l[:strings.Index(l, "…")]] = true
		}
	}
	return ids
}

// (23) · SC-003, FR-010: en el MISMO sandbox *(la lista no escribe nada)*, las identidades de la lista son las `project_ref` no
// vacías que `--run` encola después.
func TestListaFrontera_LasIdentidadesDeRun(t *testing.T) {
	dataDir, _ := hogarConLista(t)
	codigo, stdout, _, _ := ejecutar(t, 30*time.Second, "project", "list")
	if codigo != 0 {
		t.Fatalf("project list: código %d", codigo)
	}
	if codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run"); codigo != 0 {
		t.Fatalf("precondición: --run: código %d; stderr:\n%s", codigo, stderr)
	}
	b, err := os.ReadFile(transport.QueuePath(dataDir))
	if err != nil {
		t.Fatalf("precondición: la cola: %v", err)
	}
	deRun := map[string]bool{}
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		var ev event.Event
		if err := json.Unmarshal([]byte(l), &ev); err != nil {
			t.Fatalf("precondición: la cola: %v", err)
		}
		if ev.ProjectRef != "" {
			deRun[ev.ProjectRef[:12]] = true
		}
	}
	if got := identidadesDeLaTabla(stdout); len(deRun) == 0 || fmt.Sprint(sortedKeys(got)) != fmt.Sprint(sortedKeys(deRun)) {
		t.Errorf("identidades de la lista %v; las de `--run`, %v", sortedKeys(got), sortedKeys(deRun))
	}
}

func sortedKeys(m map[string]bool) []string {
	var k []string
	for x := range m {
		k = append(k, x)
	}
	sort.Strings(k)
	return k
}

// bancoDeLaLista es un servidor TLS que cuenta las peticiones y las conexiones TCP que recibe.
type bancoDeLaLista struct {
	mu                     sync.Mutex
	peticiones, conexiones int
	srv                    *httptest.Server
}

func (b *bancoDeLaLista) cuentas() (int, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.peticiones, b.conexiones
}

// (24) · SC-004, FR-004: con `endpoint` en un servidor DE CONFIANZA *(su certificado, por `SSL_CERT_FILE`)*, `device_token` y un
// evento en la cola, `project list` hace 0 peticiones y 0 conexiones. El montaje se valida antes: `--run`, en un sandbox gemelo,
// SÍ llega. Sin proxy: `ProxyFromEnvironment` no envía por proxy a 127.0.0.1. Las tres primeras hojas NACEN VERDES con la tubería
// vacía de B6; las valida m19.
func TestListaFrontera_SinRed(t *testing.T) {
	banco := &bancoDeLaLista{}
	banco.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		banco.mu.Lock()
		banco.peticiones++
		banco.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	banco.srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			banco.mu.Lock()
			banco.conexiones++
			banco.mu.Unlock()
		}
	}
	banco.srv.StartTLS()
	t.Cleanup(banco.srv.Close)
	rutaCert := filepath.Join(t.TempDir(), "banco.pem")
	if err := os.WriteFile(rutaCert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: banco.srv.Certificate().Raw}), 0o600); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	t.Setenv("SSL_CERT_FILE", rutaCert)
	enrolar := func(t *testing.T) (dataDir, hogar string) {
		dataDir, hogar = hogarConLista(t)
		// Una instalación enrolada, SIN `logs_root`: `escribirConfigCon` lo apunta a una carpeta vacía, y la lista lo respeta (P-8).
		cfg, err := json.Marshal(map[string]string{"endpoint": banco.srv.URL + rutaDeIngesta, "device_token": tokenDePrueba})
		if err != nil {
			t.Fatalf("precondición: %v", err)
		}
		escribir(t, filepath.Join(dataDir, "config.json"), cfg)
		linea := strings.SplitN(referenciaDeLista(t, "cola"), "\n", 2)[0] + "\n"
		escribir(t, transport.QueuePath(dataDir), []byte(linea))
		return dataDir, hogar
	}

	// Control del montaje, en un sandbox gemelo: con lo mismo, `--run` drena la cola contra el banco. Sin esto, un 0 no probaría nada.
	_, _ = enrolar(t)
	if codigo, _, stderr, _ := ejecutar(t, 30*time.Second, "--run"); codigo != 0 {
		t.Fatalf("precondición: --run: código %d; stderr:\n%s", codigo, stderr)
	}
	if p, _ := banco.cuentas(); p == 0 {
		t.Fatalf("precondición: `--run` no llegó al banco: el montaje no prueba nada")
	}

	_, hogar := enrolar(t)
	p0, c0 := banco.cuentas()
	codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "project", "list")
	p1, c1 := banco.cuentas()
	t.Run("peticiones", func(t *testing.T) {
		if p1 != p0 {
			t.Errorf("`project list` hizo %d peticiones; se esperaban 0", p1-p0)
		}
	})
	t.Run("conexiones", func(t *testing.T) {
		if c1 != c0 {
			t.Errorf("`project list` abrió %d conexiones; se esperaban 0", c1-c0)
		}
	})
	t.Run("importaciones", func(t *testing.T) {
		f, err := parser.ParseFile(token.NewFileSet(), "lista.go", nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("precondición: %v", err)
		}
		for _, imp := range f.Imports {
			if ruta, _ := strconv.Unquote(imp.Path.Value); ruta == "net/http" || strings.HasSuffix(ruta, "/internal/transport") {
				t.Errorf("lista.go importa %q: la lista no abre la red", ruta)
			}
		}
	})
	t.Run("tabla", func(t *testing.T) {
		if quiere := tablaEsperada(t, hogar); codigo != 0 || stdout != quiere {
			t.Errorf("código %d, stderr %q; stdout:\n%s\nse esperaba:\n%s", codigo, stderr, stdout, quiere)
		}
	})
}

// huellaCompleta es cada ruta bajo `raiz`, con su tipo, tamaño, `mtime` en ns y, si es un fichero, su md5.
func huellaCompleta(t *testing.T, raiz string) map[string]string {
	t.Helper()
	h := map[string]string{}
	err := filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		valor := fmt.Sprintf("%v %d %d", info.Mode(), info.Size(), info.ModTime().UnixNano())
		if info.Mode().IsRegular() && info.Mode().Perm()&0o400 != 0 {
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			suma := md5.Sum(b)
			valor += " " + hex.EncodeToString(suma[:])
		}
		h[p] = valor
		return nil
	})
	if err != nil {
		t.Fatalf("huella: %v", err)
	}
	return h
}

// (25) · SC-005, FR-005: `project list` no escribe nada. `huella` y `tabla`, con la sal y SIN `machine_id`, como deja el
// enrolamiento: la huella del sandbox entero *(cada ruta: tipo, tamaño, `mtime` en ns y md5)*, igual antes y después. Y sin
// `DataDir`, sigue sin existir. `huella` NACE VERDE con la tubería vacía; la valida m18. `sin_datadir`, también; la valida M-B7a.
func TestListaFrontera_SinEscritura(t *testing.T) {
	t.Run("huella_y_tabla", func(t *testing.T) {
		dataDir, hogar := hogarConLista(t)
		if err := os.Remove(filepath.Join(dataDir, "machine_id")); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		raiz := filepath.Dir(hogar)
		antes := huellaCompleta(t, raiz)
		codigo, stdout, stderr, _ := ejecutar(t, 30*time.Second, "project", "list")
		despues := huellaCompleta(t, raiz)
		t.Run("huella", func(t *testing.T) {
			if dif := diferencias(antes, despues); len(dif) != 0 {
				t.Errorf("`project list` escribió: %v", dif)
			}
		})
		t.Run("tabla", func(t *testing.T) {
			if quiere := tablaEsperada(t, hogar); codigo != 0 || stdout != quiere {
				t.Errorf("código %d, stderr %q; stdout:\n%s\nse esperaba:\n%s", codigo, stderr, stdout, quiere)
			}
		})
	})
	t.Run("sin_datadir", func(t *testing.T) {
		dataDir := testutil.Sandbox(t)
		if err := os.Remove(dataDir); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		if codigo, _, _, _ := ejecutar(t, 30*time.Second, "project", "list"); codigo != 0 {
			t.Errorf("código %d; se esperaba 0", codigo)
		}
		if _, err := os.Lstat(dataDir); !os.IsNotExist(err) {
			t.Errorf("el DataDir existe tras `project list` (err = %v): NUNCA lo crea", err)
		}
	})
}

// (26) · SC-009, P-2: una copia de una sesión de Claude Code en otra carpeta del proyecto *(reanudar)* no cambia nada: sus mensajes
// cuentan una vez.
func TestListaFrontera_CopiaUnaVez(t *testing.T) {
	_, hogar := hogarConLista(t)
	_, base, _, _ := ejecutar(t, 30*time.Second, "project", "list")
	original := filepath.Join(hogar, ".claude", "projects", "proyecto-ejemplo", "s-lista-claude-1.jsonl")
	b, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	escribir(t, filepath.Join(hogar, ".claude", "projects", "otra-carpeta", "copia.jsonl"), b)
	_, conCopia, _, _ := ejecutar(t, 30*time.Second, "project", "list")
	sal, _ := testutil.SemillasDeLaLineaBase(t)
	raiz := event.Ref(sal, filepath.Join(hogar, "dev", "proyecto-ejemplo"))[:12]
	if base == "" || conCopia != base || !strings.Contains(base, raiz+"…") {
		t.Errorf("con la copia, la tabla cambia o no trae la identidad %s:\n--- sin la copia\n%s--- con la copia\n%s", raiz, base, conCopia)
	}
}
