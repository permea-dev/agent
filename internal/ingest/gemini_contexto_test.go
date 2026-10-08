package ingest

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/state"
)

// P-009 B3 · El fichero y su contexto. Fixtures SINTÉTICOS en testdata/gemini/contexto/ (disciplina 11). Cada test monta
// en un directorio temporal la forma de `<raíz>/tmp/<slug>/chats/…`, con o sin `.project_root`.

const (
	sesionDeFixture = "s-000000000000000000000001" // la cabecera de los fixtures de contexto
	textoRaizGemini = "/tmp/raiz-sintetica-gemini"
)

func contextoBaseGemini() ContextoGemini {
	return ContextoGemini{Context: Context{Salt: salGemini, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba",
		OrgID: "org-de-prueba", AgentVersion: "test"}}
}

func lineasDeFixture(t *testing.T, nombre string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "gemini", "contexto", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return strings.SplitAfter(strings.TrimSuffix(string(b), "\n"), "\n")
}

// sesionGemini monta `<td>/.gemini/tmp/<slug>/chats/<nombre>` con las líneas dadas y, si conRaiz, el `.project_root`.
func sesionGemini(t *testing.T, td, slug, nombre string, lineas []string, conRaiz bool) FicheroGemini {
	t.Helper()
	dirSlug := filepath.Join(td, ".gemini", "tmp", slug)
	chats := filepath.Join(dirSlug, "chats")
	if err := os.MkdirAll(chats, 0o755); err != nil {
		t.Fatal(err)
	}
	if conRaiz {
		if err := os.WriteFile(filepath.Join(dirSlug, ".project_root"), []byte(textoRaizGemini+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ruta := filepath.Join(chats, nombre)
	escribirLineas(t, ruta, lineas, false)
	return FicheroGemini{Ruta: ruta, Slug: dirSlug}
}

func escribirLineas(t *testing.T, ruta string, lineas []string, anadir bool) {
	t.Helper()
	modo := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if anadir {
		modo = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}
	f, err := os.OpenFile(ruta, modo, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	for _, l := range lineas {
		if !strings.HasSuffix(l, "\n") {
			l += "\n"
		}
		if _, err := f.WriteString(l); err != nil {
			t.Fatal(err)
		}
	}
}

func leerGemini(t *testing.T, st *state.Store, f FicheroGemini, avisos *bytes.Buffer) ([]event.Event, *PasadaGemini) {
	t.Helper()
	p := NuevaPasadaGemini()
	evs, err := LeerFicheroGemini(st, f, contextoBaseGemini(), p, avisos)
	if err != nil {
		t.Fatalf("LeerFicheroGemini(%s): %v", filepath.Base(f.Ruta), err)
	}
	return evs, p
}

func cuenta(p *PasadaGemini) string {
	return fmt.Sprintf("respuestas %d · eventos %d · repetidas %d", p.Respuestas, p.Eventos, p.Repetidas)
}

func idDe(t *testing.T, id string) string {
	t.Helper()
	e, _ := derivarEventIDGemini(id)
	return e
}

// (13) · FR-006: cuentan las apariciones en línea de mensaje y en `$set.messages`; `tokens: null` no cuenta.
func TestFicheroGemini_Apariciones(t *testing.T) {
	f := sesionGemini(t, t.TempDir(), "proyecto-a", "apariciones.jsonl", lineasDeFixture(t, "apariciones.jsonl"), true)
	evs, p := leerGemini(t, state.New(), f, nil)
	if got, want := cuenta(p), "respuestas 3 · eventos 2 · repetidas 1"; got != want {
		t.Errorf("%s; se esperaba %s", got, want)
	}
	if len(evs) != 2 || evs[0].EventID != idDe(t, "m-000000000000000000000021") || evs[1].EventID != idDe(t, "m-000000000000000000000022") {
		t.Errorf("eventos %v; se esperaban m-…21 (línea) y m-…22 (sólo en `$set.messages`)", evs)
	}
}

// (14) · SC-009: la primera aparición sin tokens no cuenta; la primera CON tokens es la que sale.
func TestFicheroGemini_TokensTardios(t *testing.T) {
	f := sesionGemini(t, t.TempDir(), "proyecto-a", "tardios.jsonl", lineasDeFixture(t, "tardios.jsonl"), true)
	evs, p := leerGemini(t, state.New(), f, nil)
	if got, want := cuenta(p), "respuestas 1 · eventos 1 · repetidas 0"; got != want || len(evs) != 1 {
		t.Errorf("%s, %d eventos; se esperaba %s y 1", got, len(evs), want)
	}
}

// (15) · SC-009: un `$set.messages` final sin tokens no borra lo emitido. Nunca se reconstruye el estado final.
func TestFicheroGemini_SetFinalNoBorra(t *testing.T) {
	f := sesionGemini(t, t.TempDir(), "proyecto-a", "set_final.jsonl", lineasDeFixture(t, "set_final_sin_tokens.jsonl"), true)
	if evs, p := leerGemini(t, state.New(), f, nil); len(evs) != 2 || p.Eventos != 2 {
		t.Errorf("%d eventos (%s); se esperaban 2", len(evs), cuenta(p))
	}
}

// (16) · FR-008: dentro de la pasada, la misma respuesta sale una vez, también desde otra carpeta.
func TestFicheroGemini_RepetidasEnLaPasada(t *testing.T) {
	t.Run("herramienta", func(t *testing.T) {
		f := sesionGemini(t, t.TempDir(), "proyecto-a", "herramienta.jsonl", lineasDeFixture(t, "herramienta.jsonl"), true)
		if _, p := leerGemini(t, state.New(), f, nil); cuenta(p) != "respuestas 2 · eventos 1 · repetidas 1" {
			t.Errorf("%s; se esperaba respuestas 2 · eventos 1 · repetidas 1", cuenta(p))
		}
	})
	t.Run("dos_carpetas", func(t *testing.T) {
		td := t.TempDir()
		lineas := lineasDeFixture(t, "herramienta.jsonl")
		f1 := sesionGemini(t, td, "proyecto-a", "s.jsonl", lineas, true)
		f2 := sesionGemini(t, td, "copia-por-hash", "s.jsonl", lineas, false)
		st, p := state.New(), NuevaPasadaGemini()
		for _, f := range []FicheroGemini{f1, f2} {
			if _, err := LeerFicheroGemini(st, f, contextoBaseGemini(), p, nil); err != nil {
				t.Fatal(err)
			}
		}
		if got := cuenta(p); got != "respuestas 4 · eventos 1 · repetidas 3" {
			t.Errorf("%s; se esperaba respuestas 4 · eventos 1 · repetidas 3", got)
		}
	})
}

// (17) · SC-004: la forma de la copia, cortada antes de la reanudación y leída en dos pasadas (P-3 (a)).
func TestFicheroGemini_ReanudacionEnDosPasadas(t *testing.T) {
	lineas := lineasDeFixture(t, "forma_copia.jsonl")
	f := sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineas[:42], true)
	st := state.New()
	_, p1 := leerGemini(t, st, f, nil)
	escribirLineas(t, f.Ruta, lineas[42:], true)
	_, p2 := leerGemini(t, st, f, nil)
	t.Run("primera", func(t *testing.T) {
		if got := cuenta(p1); got != "respuestas 12 · eventos 8 · repetidas 4" {
			t.Errorf("%s; se esperaba respuestas 12 · eventos 8 · repetidas 4", got)
		}
	})
	t.Run("segunda", func(t *testing.T) {
		if got := cuenta(p2); got != "respuestas 4 · eventos 2 · repetidas 2" {
			t.Errorf("%s; se esperaba respuestas 4 · eventos 2 · repetidas 2 (la de `$set.messages`, repetida)", got)
		}
	})
}

type huellaEvento struct {
	model, projectRef, sessionRef string
	in, out, cr                   int
}

func huellas(evs []event.Event) map[string]huellaEvento {
	m := make(map[string]huellaEvento)
	for _, e := range evs {
		m[e.EventID] = huellaEvento{e.Model, e.ProjectRef, e.SessionRef, e.TokensInput, e.TokensOutput, e.TokensCacheRead}
	}
	return m
}

// (18) · FR-005: dos pasadas dan lo mismo que una: `event_id`, modelo, `project_ref`, `session_ref` y partidas.
func TestFicheroGemini_DosPasadasIgualAUna(t *testing.T) {
	lineas := lineasDeFixture(t, "forma_copia.jsonl")
	entera, _ := leerGemini(t, state.New(), sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineas, true), nil)
	f := sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineas[:42], true)
	st := state.New()
	a, _ := leerGemini(t, st, f, nil)
	escribirLineas(t, f.Ruta, lineas[42:], true)
	b, _ := leerGemini(t, st, f, nil)
	una, dos := huellas(entera), huellas(append(a, b...))
	if len(una) != 10 || len(dos) != len(una) {
		t.Fatalf("%d eventos de una vez y %d en dos pasadas; se esperaban 10 y 10", len(una), len(dos))
	}
	for id, h := range una {
		if dos[id] != h {
			t.Errorf("event_id %s: en dos pasadas %+v; de una vez %+v", id, dos[id], h)
		}
	}
}

// (19) · FR-015, D-009-P2: con la cabecera ya en el prefijo, `session_ref` es el de la cabecera.
func TestFicheroGemini_CabeceraEnElPrefijo(t *testing.T) {
	lineas := lineasDeFixture(t, "herramienta.jsonl")
	f := sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineas[:1], true)
	st := state.New()
	leerGemini(t, st, f, nil)
	escribirLineas(t, f.Ruta, lineas[1:], true)
	evs, _ := leerGemini(t, st, f, nil)
	if want := event.Ref(salGemini, sesionDeFixture); len(evs) != 1 || evs[0].SessionRef != want {
		t.Errorf("eventos %v; se esperaba 1 con session_ref = Ref(sal, sessionId de la cabecera) = %q", evs, want)
	}
}

// (20) · FR-018, D-009-P5: un `.json` cuenta una vez, y otra sólo si cambia. No se abre.
func TestContarAnteriorGemini(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "session-anterior.json")
	if err := os.WriteFile(ruta, []byte(`{"sessionId":"s-anterior","messages":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	st := state.New()
	contar := func() int {
		p := NuevaPasadaGemini()
		if err := ContarAnteriorGemini(st, ruta, p); err != nil {
			t.Fatal(err)
		}
		return p.FormatoAnterior
	}
	primera, segunda := contar(), contar()
	escribirLineas(t, ruta, []string{" "}, true)
	cambia := contar()
	for _, c := range []struct {
		hoja      string
		got, want int
	}{{"primera", primera, 1}, {"segunda", segunda, 0}, {"cambia", cambia, 1}} {
		t.Run(c.hoja, func(t *testing.T) {
			if c.got != c.want {
				t.Errorf("formato anterior %d; se esperaba %d", c.got, c.want)
			}
		})
	}
	if e := st.Files[ruta]; e.Offset != e.Size || e.Size == 0 {
		t.Errorf("entrada de state.json %+v; se esperaba Offset = Size", e)
	}
}

// textoAprobado devuelve la línea literal de la spec que empieza por `prefijo` (§Textos aprobados, SC-015).
func textoAprobado(t *testing.T, prefijo string) string {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "specs", "009-lector-gemini", "spec.md"))
	if err != nil {
		t.Fatalf("precondición: spec: %v", err)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), prefijo) {
			return sc.Text()
		}
	}
	t.Fatalf("precondición: la spec no tiene una línea que empiece por %q", prefijo)
	return ""
}

// (21) · SC-008, FR-019: el resumen literal, la identidad de la cuenta y el predicado del demonio.
func TestPasadaGemini_ResumenSC008(t *testing.T) {
	f := sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineasDeFixture(t, "cuenta_sc008.jsonl"), true)
	_, p := leerGemini(t, state.New(), f, nil)
	t.Run("literal", func(t *testing.T) {
		want := fmt.Sprintf(textoAprobado(t, "gemini: respuestas %d"), 7, 3, 1, 1, 2, 1, 1, 0)
		if got := p.Resumen(); got != want {
			t.Errorf("Resumen() = %q\nse esperaba     %q", got, want)
		}
	})
	t.Run("identidad", func(t *testing.T) {
		if p.Respuestas != p.Eventos+p.Repetidas+p.SinIdentificador+p.Incoherentes || p.SinModelo > p.Eventos || p.TotalDescuadrado > p.Eventos {
			t.Errorf("la cuenta no cumple FR-019: %+v", *p)
		}
	})
	t.Run("hay_novedades", func(t *testing.T) {
		anterior := NuevaPasadaGemini()
		anterior.FormatoAnterior = 1
		if NuevaPasadaGemini().HayNovedades() || !p.HayNovedades() || !anterior.HayNovedades() {
			t.Errorf("HayNovedades: vacía %t, con respuestas %t, sólo formato anterior %t; se esperaba false, true, true",
				NuevaPasadaGemini().HayNovedades(), p.HayNovedades(), anterior.HayNovedades())
		}
	})
}

// (22) · FR-020: una línea corrupta en lo nuevo se avisa y se salta; en el prefijo, en silencio.
func TestFicheroGemini_LineaCorrupta(t *testing.T) {
	f := sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineasDeFixture(t, "corrupta.jsonl"), true)
	st := state.New()
	var avisos1, avisos2 bytes.Buffer
	_, p1 := leerGemini(t, st, f, &avisos1)
	escribirLineas(t, f.Ruta, lineasDeFixture(t, "corrupta_mas.jsonl"), true)
	_, p2 := leerGemini(t, st, f, &avisos2)
	t.Run("primera", func(t *testing.T) {
		if n := strings.Count(avisos1.String(), "skip (línea corrupta): "); n != 1 || cuenta(p1) != "respuestas 2 · eventos 2 · repetidas 0" {
			t.Errorf("%d avisos y %s; se esperaba 1 aviso y respuestas 2 · eventos 2 · repetidas 0", n, cuenta(p1))
		}
	})
	t.Run("segunda", func(t *testing.T) {
		if avisos2.Len() != 0 || p2.Eventos != 1 {
			t.Errorf("avisos %q y %d eventos; se esperaba ningún aviso y 1 evento", avisos2.String(), p2.Eventos)
		}
	})
}

func tocar(t *testing.T, ruta, contenido string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sinPermisosPOSIX(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("depende de permisos POSIX sin root (R-7)")
	}
}

// (23) · FR-003, D-009-P4: el patrón, las exclusiones y el orden de `ListarGemini`.
func TestListarGemini(t *testing.T) {
	raiz := filepath.Join(t.TempDir(), ".gemini")
	tmp := filepath.Join(raiz, "tmp")
	tocar(t, filepath.Join(tmp, "b-con", ".project_root"), textoRaizGemini+"\n")
	for _, r := range []string{"b-con/chats/s1.jsonl", "b-con/chats/padre/sub.jsonl", "a-sin/chats/s2.jsonl"} {
		tocar(t, filepath.Join(tmp, r), "{}\n")
	}
	for _, r := range []string{"b-con/chats/s1.json", "b-con/chats/padre/sub.json"} {
		tocar(t, filepath.Join(tmp, r), "{}")
	}
	for _, r := range []string{"b-con/chats/x.jsonl.tmp-123", "b-con/chats/y.jsonl.unreadable-1", "b-con/logs.json",
		"b-con/logs/z.jsonl", "b-con/z.jsonl", "b-con/chats/padre/hijo/profundo.jsonl"} {
		tocar(t, filepath.Join(tmp, r), "{}\n")
	}
	sesiones, anteriores, err := ListarGemini(raiz)
	if err != nil {
		t.Fatalf("ListarGemini: %v", err)
	}
	t.Run("patron", func(t *testing.T) {
		var got []string
		for _, s := range sesiones {
			rel, _ := filepath.Rel(tmp, s.Ruta)
			slug, _ := filepath.Rel(tmp, s.Slug)
			got = append(got, filepath.ToSlash(rel)+" @ "+slug)
		}
		var ant []string
		for _, a := range anteriores {
			rel, _ := filepath.Rel(tmp, a)
			ant = append(ant, filepath.ToSlash(rel))
		}
		if want := "b-con/chats/padre/sub.jsonl @ b-con|b-con/chats/s1.jsonl @ b-con|a-sin/chats/s2.jsonl @ a-sin"; strings.Join(got, "|") != want {
			t.Errorf("sesiones %q\nse esperaba %q", strings.Join(got, "|"), want)
		}
		if want := "b-con/chats/padre/sub.json|b-con/chats/s1.json"; strings.Join(ant, "|") != want {
			t.Errorf("anteriores %q; se esperaba %q", strings.Join(ant, "|"), want)
		}
	})
	t.Run("orden", func(t *testing.T) {
		if len(sesiones) != 3 || filepath.Base(sesiones[2].Slug) != "a-sin" {
			t.Errorf("la carpeta sin `.project_root` debía ir al final: %v", sesiones)
		}
	})
	t.Run("subdirectorio_ilegible", func(t *testing.T) {
		sinPermisosPOSIX(t)
		ilegible := filepath.Join(tmp, "c-ilegible", "chats")
		tocar(t, filepath.Join(ilegible, "s3.jsonl"), "{}\n")
		if err := os.Chmod(ilegible, 0o000); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chmod(ilegible, 0o755) }()
		if s, _, err := ListarGemini(raiz); err != nil || len(s) != 3 {
			t.Errorf("con un `chats/` ilegible: %d sesiones, error %v; se esperaban 3 y ninguno", len(s), err)
		}
	})
	t.Run("raiz_inexistente", func(t *testing.T) {
		if _, _, err := ListarGemini(filepath.Join(t.TempDir(), "no-existe")); err == nil {
			t.Error("sin `<raíz>/tmp`, se esperaba el error de la raíz")
		}
	})
}

// (24) · FR-014, FR-025, D-009-P3: el proyecto sale del `.project_root` de la carpeta `<slug>`.
func TestFicheroGemini_Proyecto(t *testing.T) {
	lineas := lineasDeFixture(t, "herramienta.jsonl")
	t.Run("con", func(t *testing.T) {
		evs, _ := leerGemini(t, state.New(), sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineas, true), nil)
		if want := project.Derivar(textoRaizGemini, salGemini); len(evs) != 1 || evs[0].ProjectRef != want {
			t.Errorf("eventos %v; se esperaba project_ref = Derivar(.project_root sin el salto final) = %q", evs, want)
		}
	})
	t.Run("sin", func(t *testing.T) {
		evs, _ := leerGemini(t, state.New(), sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineas, false), nil)
		if len(evs) != 1 || evs[0].ProjectRef != "" {
			t.Errorf("eventos %v; sin `.project_root` se esperaba project_ref vacío", evs)
		}
	})
	t.Run("subagente", func(t *testing.T) {
		td := t.TempDir()
		sesionGemini(t, td, "proyecto-a", "principal.jsonl", lineas[:1], true)
		sub := filepath.Join(td, ".gemini", "tmp", "proyecto-a", "chats", "padre", "sub.jsonl")
		if err := os.MkdirAll(filepath.Dir(sub), 0o755); err != nil {
			t.Fatal(err)
		}
		escribirLineas(t, sub, lineas, false)
		sesiones, _, err := ListarGemini(filepath.Join(td, ".gemini"))
		if err != nil {
			t.Fatal(err)
		}
		var f FicheroGemini
		for _, s := range sesiones {
			if s.Ruta == sub {
				f = s
			}
		}
		evs, _ := leerGemini(t, state.New(), f, nil)
		if want := project.Derivar(textoRaizGemini, salGemini); len(evs) != 1 || evs[0].ProjectRef != want {
			t.Errorf("el subagente: eventos %v; se esperaba el project_ref de su `<slug>` = %q", evs, want)
		}
	})
	t.Run("ilegible", func(t *testing.T) {
		sinPermisosPOSIX(t)
		f := sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineas, true)
		raiz := filepath.Join(f.Slug, ".project_root")
		if err := os.Chmod(raiz, 0o000); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chmod(raiz, 0o644) }()
		st := state.New()
		if _, err := LeerFicheroGemini(st, f, contextoBaseGemini(), NuevaPasadaGemini(), nil); err == nil {
			t.Error("con `.project_root` ilegible se esperaba un error (FR-025: el fichero se omite)")
		}
		if _, visto := st.Files[f.Ruta]; visto {
			t.Error("el fichero omitido no debe tener estado: su offset no avanza")
		}
	})
}

// (25) · Un fichero truncado o rotado se relee desde 0, y entonces NO hay prefijo.
func TestFicheroGemini_TruncadoSinPrefijo(t *testing.T) {
	f := sesionGemini(t, t.TempDir(), "proyecto-a", "s.jsonl", lineasDeFixture(t, "largo.jsonl"), true)
	st := state.New()
	leerGemini(t, st, f, nil)
	escribirLineas(t, f.Ruta, lineasDeFixture(t, "truncado.jsonl"), false)
	if _, p := leerGemini(t, st, f, nil); cuenta(p) != "respuestas 1 · eventos 1 · repetidas 0" {
		t.Errorf("%s; tras truncar se esperaba respuestas 1 · eventos 1 · repetidas 0", cuenta(p))
	}
}
