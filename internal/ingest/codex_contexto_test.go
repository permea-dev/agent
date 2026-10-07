package ingest

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/state"
)

// P-008 B3 · Contexto, estado y recuentos. Fixtures SINTÉTICOS en testdata/codex/contexto/ (disciplina 11). Los tests
// los copian a un temporal, porque leer actualiza el estado y algunos casos añaden líneas o truncan el fichero.

func contextoBaseCodex() ContextoCodex {
	return ContextoCodex{Context: Context{Salt: salCodex, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba",
		OrgID: "org-de-prueba", AgentVersion: "test"}}
}

func fixtureContexto(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "codex", "contexto", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return b
}

// sesionDePrueba escribe en un temporal la concatenación de los fixtures y devuelve su ruta.
func sesionDePrueba(t *testing.T, dir, nombre string, fixtures ...string) string {
	t.Helper()
	var b bytes.Buffer
	for _, f := range fixtures {
		b.Write(fixtureContexto(t, f))
	}
	ruta := filepath.Join(dir, nombre)
	if err := os.WriteFile(ruta, b.Bytes(), 0o600); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	return ruta
}

func anexar(t *testing.T, ruta, fixture string) {
	t.Helper()
	f, err := os.OpenFile(ruta, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(fixtureContexto(t, fixture)); err != nil {
		t.Fatalf("precondición: %v", err)
	}
}

func leerCodex(t *testing.T, st *state.Store, ruta string, p *PasadaCodex, avisos *bytes.Buffer) []event.Event {
	t.Helper()
	evs, err := LeerFicheroCodex(st, ruta, contextoBaseCodex(), p, avisos)
	if err != nil {
		t.Fatalf("LeerFicheroCodex(%s): %v", filepath.Base(ruta), err)
	}
	return evs
}

// deUnaVez lee el fichero entero, con un estado nuevo: la referencia de FR-005. Exige los eventos que el fixture trae,
// para que una referencia vacía no haga pasar la comparación.
func deUnaVez(t *testing.T, dir string, eventos int, fixtures ...string) []event.Event {
	t.Helper()
	evs := leerCodex(t, state.New(), sesionDePrueba(t, dir, "de-una-vez.jsonl", fixtures...), NuevaPasadaCodex(), nil)
	if len(evs) != eventos {
		t.Fatalf("precondición: leyendo de una vez salen %d eventos; el fixture trae %d", len(evs), eventos)
	}
	return evs
}

func modelos(evs []event.Event) []string {
	var m []string
	for _, e := range evs {
		m = append(m, e.Model)
	}
	return m
}

// mismoContexto compara lo que FR-005 exige igual entre leer en dos pasadas y leer de una vez.
func mismoContexto(t *testing.T, got, want []event.Event) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%d eventos; leyendo de una vez salen %d", len(got), len(want))
	}
	for i := range got {
		g, w := got[i], want[i]
		if g.EventID != w.EventID || g.Model != w.Model || g.ProjectRef != w.ProjectRef || g.SessionRef != w.SessionRef {
			t.Errorf("evento %d: (id %s, modelo %q, proyecto %q, sesión %q); de una vez (id %s, modelo %q, proyecto %q, sesión %q)",
				i, g.EventID, g.Model, g.ProjectRef, g.SessionRef, w.EventID, w.Model, w.ProjectRef, w.SessionRef)
		}
	}
}

// (11) · FR-011: el modelo es el del `turn_context` del turno, no el vigente.
func TestContextoCodex_ModeloDelTurno(t *testing.T) {
	t.Run("dos_turnos", func(t *testing.T) {
		got := modelos(leerCodex(t, state.New(), sesionDePrueba(t, t.TempDir(), "s.jsonl", "dos_turnos.jsonl"), NuevaPasadaCodex(), nil))
		if strings.Join(got, ",") != "modelo-a,modelo-b" {
			t.Errorf("modelos %v; se esperaba [modelo-a modelo-b]", got)
		}
	})
	t.Run("ajuste_dentro_del_turno", func(t *testing.T) {
		got := modelos(leerCodex(t, state.New(), sesionDePrueba(t, t.TempDir(), "s.jsonl", "ajuste_dentro_del_turno.jsonl"), NuevaPasadaCodex(), nil))
		if strings.Join(got, ",") != "modelo-a" {
			t.Errorf("modelos %v; se esperaba el del turno, [modelo-a], no el vigente (modelo-b)", got)
		}
	})
}

// (12) · FR-011, M-5: la compactación no tiene `turn_context` y lleva el modelo vigente.
func TestContextoCodex_CompactacionLlevaElVigente(t *testing.T) {
	got := modelos(leerCodex(t, state.New(), sesionDePrueba(t, t.TempDir(), "s.jsonl", "compactacion.jsonl"), NuevaPasadaCodex(), nil))
	if strings.Join(got, ",") != "modelo-a,modelo-b" {
		t.Errorf("modelos %v; se esperaba [modelo-a modelo-b]: la compactación, el vigente", got)
	}
}

// (13) · FR-011, P-4 (b): sin ningún modelo, el evento sale con `model` vacío y se cuenta.
func TestContextoCodex_SinModelo(t *testing.T) {
	p := NuevaPasadaCodex()
	evs := leerCodex(t, state.New(), sesionDePrueba(t, t.TempDir(), "s.jsonl", "sin_modelo.jsonl"), p, nil)
	if len(evs) != 1 || evs[0].Model != "" || p.SinModelo != 1 {
		t.Errorf("%d eventos, modelos %v, sin modelo %d; se esperaba 1 evento con modelo vacío y sin modelo = 1", len(evs), modelos(evs), p.SinModelo)
	}
}

// (14) · SC-010, FR-005: cortado tras su `turn_context`, el registro de la segunda pasada lleva lo mismo que leyendo de una vez.
func TestContextoCodex_ContextoEntrePasadas(t *testing.T) {
	dir := t.TempDir()
	st := state.New()
	ruta := sesionDePrueba(t, dir, "s.jsonl", "corte_1.jsonl")
	if evs := leerCodex(t, st, ruta, NuevaPasadaCodex(), nil); len(evs) != 0 {
		t.Fatalf("precondición: la primera pasada no tiene registros y dio %d eventos", len(evs))
	}
	anexar(t, ruta, "corte_2.jsonl")
	mismoContexto(t, leerCodex(t, st, ruta, NuevaPasadaCodex(), nil), deUnaVez(t, dir, 1, "corte_1.jsonl", "corte_2.jsonl"))
}

// (15) · SC-005: una reanudación leída en dos pasadas da 1 + 3 eventos, sin repetir, iguales a leer de una vez.
func TestContextoCodex_ReanudacionEnDosPasadas(t *testing.T) {
	dir := t.TempDir()
	st := state.New()
	ruta := sesionDePrueba(t, dir, "s.jsonl", "reanudacion_1.jsonl")
	entera := deUnaVez(t, dir, 4, "reanudacion_1.jsonl", "reanudacion_2.jsonl")
	t.Run("primera", func(t *testing.T) {
		mismoContexto(t, leerCodex(t, st, ruta, NuevaPasadaCodex(), nil), entera[:1])
	})
	t.Run("segunda", func(t *testing.T) {
		anexar(t, ruta, "reanudacion_2.jsonl")
		mismoContexto(t, leerCodex(t, st, ruta, NuevaPasadaCodex(), nil), entera[1:])
	})
}

// (16) · SC-007, FR-008: una respuesta copiada a otro fichero (bifurcación) sale una vez y se cuenta como repetida.
func TestContextoCodex_BifurcacionUnEvento(t *testing.T) {
	dir := t.TempDir()
	st, p := state.New(), NuevaPasadaCodex()
	n := len(leerCodex(t, st, sesionDePrueba(t, dir, "padre.jsonl", "bifurcacion_padre.jsonl"), p, nil))
	n += len(leerCodex(t, st, sesionDePrueba(t, dir, "hijo.jsonl", "bifurcacion_hijo.jsonl"), p, nil))
	if n != 1 || p.Repetidas != 1 {
		t.Errorf("%d eventos y %d repetidas; se esperaba 1 y 1", n, p.Repetidas)
	}
}

// (17) · FR-017, Q-2 (a): el formato anterior no emite y se cuenta; el mixto es actual; sin consumo, ni lo uno ni lo otro.
func TestContextoCodex_Formatos(t *testing.T) {
	for _, c := range []struct {
		hoja, fixture       string
		eventos, anteriores int
	}{
		{"anterior", "formato_anterior.jsonl", 0, 1},
		{"mixto", "mixto.jsonl", 1, 0},
		{"sin_consumo", "sin_consumo.jsonl", 0, 0},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			p := NuevaPasadaCodex()
			evs := leerCodex(t, state.New(), sesionDePrueba(t, t.TempDir(), "s.jsonl", c.fixture), p, nil)
			if len(evs) != c.eventos || p.FormatoAnterior != c.anteriores {
				t.Errorf("%d eventos y %d en formato anterior; se esperaba %d y %d", len(evs), p.FormatoAnterior, c.eventos, c.anteriores)
			}
		})
	}
}

// (18) · FR-018, D-008-P2: un `.zst` cuenta una vez, y otra sólo si cambia. Su entrada en `state.json` tiene los cuatro campos.
func TestContextoCodex_ComprimidoUnaVez(t *testing.T) {
	dir := t.TempDir()
	ruta := filepath.Join(dir, "rollout-sintetico.jsonl.zst")
	if err := os.WriteFile(ruta, fixtureContexto(t, "comprimido.jsonl.zst"), 0o600); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	st := state.New()
	contar := func(t *testing.T, want int) {
		t.Helper()
		p := NuevaPasadaCodex()
		if err := ContarComprimido(st, ruta, p); err != nil {
			t.Fatalf("ContarComprimido: %v", err)
		}
		if p.Comprimidos != want {
			t.Errorf("comprimidos = %d; se esperaba %d", p.Comprimidos, want)
		}
	}
	t.Run("primera", func(t *testing.T) { contar(t, 1) })
	t.Run("segunda", func(t *testing.T) { contar(t, 0) })
	t.Run("tras_cambiar", func(t *testing.T) {
		f, err := os.OpenFile(ruta, os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			t.Fatalf("precondición: %v", err)
		}
		_, _ = f.Write([]byte("más bytes"))
		_ = f.Close()
		contar(t, 1)
	})
}

// (19) · SC-017, FR-027: una respuesta de cada clase; la línea de resumen y su identidad.
func TestContextoCodex_CuentaDelResumen(t *testing.T) {
	p := NuevaPasadaCodex()
	leerCodex(t, state.New(), sesionDePrueba(t, t.TempDir(), "s.jsonl", "cuenta.jsonl"), p, nil)
	want := "codex: respuestas 5 · eventos 2 · repetidas 1 · sin identificador 1 · incoherentes 1 · sin modelo 1" +
		" · ficheros en formato anterior 0 · ficheros comprimidos 0"
	if got := p.Resumen(); got != want {
		t.Errorf("resumen:\n got %q\nwant %q", got, want)
	}
	if p.Respuestas != p.Eventos+p.Repetidas+p.SinIdentificador+p.Incoherentes || p.SinModelo > p.Eventos {
		t.Errorf("la identidad de FR-027 no se cumple: %+v", *p)
	}
}

// (20) · FR-014, P-5, E-4: el `cwd` del turno; si falta, el de `session_meta`; si no hay ninguno, `project_ref` vacío.
func TestContextoCodex_Cwd(t *testing.T) {
	evs := leerCodex(t, state.New(), sesionDePrueba(t, t.TempDir(), "s.jsonl", "cwd.jsonl"), NuevaPasadaCodex(), nil)
	if len(evs) != 2 {
		t.Fatalf("precondición: %d eventos; se esperaban 2", len(evs))
	}
	t.Run("del_turno", func(t *testing.T) {
		if want := project.Derivar("/tmp/turno-cwd", salCodex); evs[0].ProjectRef != want {
			t.Errorf("project_ref = %q; se esperaba el del cwd del turno, %q", evs[0].ProjectRef, want)
		}
	})
	t.Run("de_session_meta", func(t *testing.T) {
		if want := project.Derivar("/tmp/sesion-cwd", salCodex); evs[1].ProjectRef != want {
			t.Errorf("project_ref = %q; se esperaba el del cwd de session_meta, %q", evs[1].ProjectRef, want)
		}
	})
	t.Run("sin_cwd", func(t *testing.T) {
		got := leerCodex(t, state.New(), sesionDePrueba(t, t.TempDir(), "s.jsonl", "sin_cwd.jsonl"), NuevaPasadaCodex(), nil)
		if len(got) != 1 || got[0].ProjectRef != "" {
			t.Errorf("%d eventos, project_ref %v; se esperaba 1 evento con project_ref vacío", len(got), got)
		}
	})
}

// (34) · T055, FR-029, E-4: una línea corrupta en la parte nueva se avisa una vez y se salta; en el prefijo, en silencio.
func TestContextoCodex_LineaCorrupta(t *testing.T) {
	st, ruta := state.New(), sesionDePrueba(t, t.TempDir(), "s.jsonl", "corrupta_1.jsonl")
	t.Run("primera_pasada", func(t *testing.T) {
		var avisos bytes.Buffer
		p := NuevaPasadaCodex()
		evs := leerCodex(t, st, ruta, p, &avisos)
		if n := strings.Count(avisos.String(), "skip (línea corrupta): "); n != 1 || len(evs) != 1 || p.Respuestas != 1 {
			t.Errorf("%d avisos, %d eventos, %d respuestas; se esperaba 1, 1 y 1. Avisos: %q", n, len(evs), p.Respuestas, avisos.String())
		}
	})
	t.Run("segunda_pasada", func(t *testing.T) {
		anexar(t, ruta, "corrupta_2.jsonl")
		var avisos bytes.Buffer
		evs := leerCodex(t, st, ruta, NuevaPasadaCodex(), &avisos)
		if avisos.Len() != 0 || len(evs) != 1 {
			t.Errorf("avisos %q y %d eventos; la línea corrupta ya está en el prefijo: se esperaba ningún aviso y 1 evento", avisos.String(), len(evs))
		}
	})
}

// (35) · Un fichero truncado o rotado se relee desde 0 (`Recorrer`), y entonces NO se lee prefijo: el contexto sale sólo
// de lo que se relee, en orden. En el fichero nuevo el `turn_context` va DESPUÉS de su registro: leído como prefijo, daría
// al registro un modelo que todavía no tenía.
func TestContextoCodex_TruncadoNoLeePrefijo(t *testing.T) {
	dir := t.TempDir()
	st := state.New()
	ruta := sesionDePrueba(t, dir, "s.jsonl", "truncado_largo.jsonl")
	if evs := leerCodex(t, st, ruta, NuevaPasadaCodex(), nil); len(evs) != 1 {
		t.Fatalf("precondición: la primera pasada dio %d eventos; se esperaba 1", len(evs))
	}
	sesionDePrueba(t, dir, "s.jsonl", "truncado_corto.jsonl") // reescribe el fichero, más corto que su offset
	p := NuevaPasadaCodex()
	evs := leerCodex(t, st, ruta, p, nil)
	if len(evs) != 1 || evs[0].Model != "" || p.SinModelo != 1 {
		t.Errorf("%d eventos, modelos %v, sin modelo %d; se esperaba 1 evento sin modelo (su turn_context va después)", len(evs), modelos(evs), p.SinModelo)
	}
}

// (36) · FR-003, FR-018 (Encargo 8): bajo la raíz, las sesiones `.jsonl` a cualquier profundidad y los `.zst`, cada lista
// en orden; nada más.
func TestContextoCodex_ListarRaiz(t *testing.T) {
	raiz := t.TempDir()
	for _, f := range []string{"2026/10/07/a.jsonl", "2026/10/08/b.jsonl", "2026/10/07/c.jsonl.zst", "notas.txt"} {
		ruta := filepath.Join(raiz, f)
		if err := os.MkdirAll(filepath.Dir(ruta), 0o700); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		if err := os.WriteFile(ruta, []byte("x"), 0o600); err != nil {
			t.Fatalf("precondición: %v", err)
		}
	}
	sesiones, comprimidos, err := ListarCodex(raiz)
	rel := func(rutas []string) string {
		var r []string
		for _, x := range rutas {
			s, _ := filepath.Rel(raiz, x)
			r = append(r, filepath.ToSlash(s))
		}
		return strings.Join(r, ",")
	}
	if err != nil || rel(sesiones) != "2026/10/07/a.jsonl,2026/10/08/b.jsonl" || rel(comprimidos) != "2026/10/07/c.jsonl.zst" {
		t.Errorf("ListarCodex = (%q, %q, %v); se esperaba ([2026/10/07/a.jsonl 2026/10/08/b.jsonl], [2026/10/07/c.jsonl.zst], nil)",
			rel(sesiones), rel(comprimidos), err)
	}
}
