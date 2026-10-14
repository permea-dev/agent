package ingest

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/state"
)

// P-011 B1 · Qué ficheros lee el lector de Copilot CLI (P-7, FR-003, D-6). Los fixtures son SINTÉTICOS, de
// `testdata/copilot/generar.sh`, y NUNCA se abren en su sitio: cada test trabaja sobre una copia en `t.TempDir()`.

// copiarArbolCopilot copia `testdata/copilot/<nombre>` entero a un temporal y devuelve su ruta.
func copiarArbolCopilot(t *testing.T, nombre string) string {
	t.Helper()
	origen := filepath.Join("testdata", "copilot", nombre)
	destino := filepath.Join(t.TempDir(), nombre)
	err := filepath.WalkDir(origen, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(origen, p)
		if err != nil {
			return err
		}
		a := filepath.Join(destino, rel)
		if d.IsDir() {
			return os.MkdirAll(a, 0o700)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(a, b, 0o600)
	})
	if err != nil {
		t.Fatalf("copiar %s: %v", origen, err)
	}
	return destino
}

// sesionesDePrueba es la copia de `sesiones/` con el `session.db` de 0 B que `.gitignore` no deja versionar (generar.sh).
func sesionesDePrueba(t *testing.T) string {
	t.Helper()
	raiz := copiarArbolCopilot(t, "sesiones")
	if err := os.WriteFile(filepath.Join(raiz, "s-000000000000000000000001", "session.db"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return raiz
}

// listarCopilotRelativo llama a `ListarCopilot` y devuelve las rutas relativas a la raíz, con `/`.
func listarCopilotRelativo(t *testing.T, raiz string) []string {
	t.Helper()
	rutas, err := ListarCopilot(raiz)
	if err != nil {
		t.Fatalf("ListarCopilot: %v", err)
	}
	var rel []string
	for _, r := range rutas {
		x, err := filepath.Rel(raiz, r)
		if err != nil {
			t.Fatal(err)
		}
		rel = append(rel, filepath.ToSlash(x))
	}
	if len(rel) == 0 {
		t.Fatal("ListarCopilot no devolvió nada")
	}
	return rel
}

// (3) · P-7: sólo `<raíz>/<carpeta>/events.jsonl`, de un nivel; ni `.session-operation-locks/`, ni `checkpoints/`,
// `session.db` o un `events.jsonl` más hondo. La carpeta sin `events.jsonl` no es un error ni se avisa (D-6). Orden léxico.
func TestListarCopilot(t *testing.T) {
	s1 := "s-000000000000000000000001/events.jsonl"
	s2 := "s-000000000000000000000002/events.jsonl"
	t.Run("solo_events", func(t *testing.T) {
		got := listarCopilotRelativo(t, sesionesDePrueba(t))
		if want := []string{s1, s2}; !slices.Equal(got, want) {
			t.Errorf("ListarCopilot = %q; se esperaba %q", got, want)
		}
	})
	t.Run("sin_locks", func(t *testing.T) {
		for _, r := range listarCopilotRelativo(t, sesionesDePrueba(t)) {
			if strings.HasPrefix(r, ".") {
				t.Errorf("lista %q, de una carpeta oculta", r)
			}
		}
	})
	t.Run("sin_checkpoints_db", func(t *testing.T) {
		for _, r := range listarCopilotRelativo(t, sesionesDePrueba(t)) {
			if path.Base(r) != "events.jsonl" || strings.Contains(r, "checkpoints/") {
				t.Errorf("lista %q, que no es el events.jsonl de una sesión", r)
			}
		}
	})
	t.Run("hondo", func(t *testing.T) {
		for _, r := range listarCopilotRelativo(t, sesionesDePrueba(t)) {
			if strings.Count(r, "/") != 1 {
				t.Errorf("lista %q, a más de un nivel", r)
			}
		}
	})
	t.Run("vacia", func(t *testing.T) {
		for _, r := range listarCopilotRelativo(t, sesionesDePrueba(t)) {
			if strings.HasPrefix(r, "s-000000000000000000000003/") {
				t.Errorf("lista %q, de la carpeta sin events.jsonl", r)
			}
		}
	})
	t.Run("orden", func(t *testing.T) {
		got := listarCopilotRelativo(t, sesionesDePrueba(t))
		if len(got) < 2 || !slices.IsSorted(got) {
			t.Errorf("ListarCopilot = %q; se esperaba en orden léxico", got)
		}
	})
}

// ═══ P-011 B3 · Un fichero de Copilot CLI: offset, contexto por tramo, repetidas, truncado, sin LF, subagente, sólo lectura ═══
//
// Los fixtures son los de `testdata/copilot/ficheros/` (generar.sh), copiados a `t.TempDir()`. Cada pasada afirma su línea
// de resumen LITERAL, la de `esperado.json` cuando el fichero se lee de una vez.

// ficheroCopilot copia `ficheros/<nombre>` a un temporal, con el nombre `events.jsonl`, y devuelve su ruta.
func ficheroCopilot(t *testing.T, nombre string) string {
	t.Helper()
	ruta := filepath.Join(t.TempDir(), "events.jsonl")
	ponerFicheroCopilot(t, nombre, ruta)
	return ruta
}

// ponerFicheroCopilot escribe `ficheros/<nombre>` sobre `ruta` (para las secuencias entre pasadas).
func ponerFicheroCopilot(t *testing.T, nombre, ruta string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "copilot", "ficheros", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	if err := os.WriteFile(ruta, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// resumenEsperadoCopilot es la línea de `esperado.json` del fichero.
func resumenEsperadoCopilot(t *testing.T, nombre string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "copilot", "ficheros", "esperado.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m[nombre] == "" {
		t.Fatalf("precondición: esperado.json no trae %s", nombre)
	}
	return m[nombre]
}

func contextoFicheroCopilot() ContextoCopilot {
	return ContextoCopilot{Context: Context{Salt: salCopilot, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba",
		OrgID: "org-de-prueba", AgentVersion: "test"}}
}

// pasarCopilot hace una pasada que debe acabar bien y devuelve sus eventos, sus recuentos y sus avisos.
func pasarCopilot(t *testing.T, st *state.Store, ruta string) ([]event.Event, *PasadaCopilot, string) {
	t.Helper()
	p := NuevaPasadaCopilot()
	var avisos bytes.Buffer
	evs, err := LeerFicheroCopilot(st, ruta, contextoFicheroCopilot(), p, &avisos)
	if err != nil {
		t.Fatalf("LeerFicheroCopilot: %v", err)
	}
	return evs, p, avisos.String()
}

// deUnaVezCopilot lee un fixture entero con un estado nuevo, y exige su línea de `esperado.json`.
func deUnaVezCopilot(t *testing.T, nombre string) []event.Event {
	t.Helper()
	evs, p, _ := pasarCopilot(t, state.New(), ficheroCopilot(t, nombre))
	if got, want := p.Resumen(), resumenEsperadoCopilot(t, nombre); got != want {
		t.Fatalf("%s: resumen\n  %s\nse esperaba\n  %s", nombre, got, want)
	}
	return evs
}

// (15) · FR-005: un registro sale una vez; un fichero que no creció no se abre. FR-009: sin tope de línea.
func TestLeerFicheroCopilot_Pasadas(t *testing.T) {
	t.Run("dos_llamadas", func(t *testing.T) {
		st, ruta := state.New(), ficheroCopilot(t, "dos_llamadas.jsonl")
		if evs, _, _ := pasarCopilot(t, st, ruta); len(evs) != 2 {
			t.Fatalf("1.ª pasada: %d eventos; se esperaban 2", len(evs))
		}
		if evs, p, _ := pasarCopilot(t, st, ruta); len(evs) != 0 || p.Registros != 0 {
			t.Errorf("2.ª pasada: %d eventos, %d registros; se esperaba 0", len(evs), p.Registros)
		}
	})
	t.Run("aperturas", func(t *testing.T) {
		st, ruta := state.New(), ficheroCopilot(t, "dos_llamadas.jsonl")
		_, p1, _ := pasarCopilot(t, st, ruta)
		_, p2, _ := pasarCopilot(t, st, ruta)
		if p1.Aperturas == 0 || p2.Aperturas != 0 {
			t.Errorf("aperturas: 1.ª %d, 2.ª %d; se esperaba ≥ 1 y 0", p1.Aperturas, p2.Aperturas)
		}
	})
	t.Run("linea_larga", func(t *testing.T) {
		if evs := deUnaVezCopilot(t, "linea_larga.jsonl"); len(evs) != 2 {
			t.Errorf("%d eventos; se esperaban 2 tras una línea de más de 64 KB", len(evs))
		}
	})
}

// (16) · P-3: el `cwd` es el del último `start` o `resume`; la sesión, la del `start`. SC-021: los valores.
func TestLeerFicheroCopilot_Tramos(t *testing.T) {
	evs := deUnaVezCopilot(t, "tramos.jsonl")
	if len(evs) != 2 {
		t.Fatalf("%d eventos; se esperaban 2", len(evs))
	}
	t.Run("dos_proyectos", func(t *testing.T) {
		if evs[0].ProjectRef == "" || evs[0].ProjectRef == evs[1].ProjectRef {
			t.Errorf("project_ref = %q y %q; se esperaban dos distintos", evs[0].ProjectRef, evs[1].ProjectRef)
		}
	})
	t.Run("una_sesion", func(t *testing.T) {
		if evs[0].SessionRef == "" || evs[0].SessionRef != evs[1].SessionRef {
			t.Errorf("session_ref = %q y %q; se esperaba uno solo", evs[0].SessionRef, evs[1].SessionRef)
		}
	})
	t.Run("valores", func(t *testing.T) {
		sesion := event.Ref(salCopilot, "s-000000000000000000000001")
		var r *project.Resolutor
		x, y := r.Derivar("/sintetico/x", salCopilot), r.Derivar("/sintetico/y", salCopilot)
		if evs[0].SessionRef != sesion || evs[0].ProjectRef != x || evs[1].ProjectRef != y {
			t.Errorf("refs = (%q, %q, %q); se esperaba (%q, %q, %q)", evs[0].SessionRef, evs[0].ProjectRef, evs[1].ProjectRef,
				sesion, x, y)
		}
	})
}

// (17) · FR-008: un registro leído en una pasada posterior lleva las mismas referencias que leyendo el fichero de una vez.
// El prefijo lleva una línea de 64 KB.
func TestLeerFicheroCopilot_CortadoEnDosPasadas(t *testing.T) {
	st, ruta := state.New(), ficheroCopilot(t, "cortado/paso1.jsonl")
	evs, _, _ := pasarCopilot(t, st, ruta)
	ponerFicheroCopilot(t, "cortado/paso2.jsonl", ruta)
	evs2, _, _ := pasarCopilot(t, st, ruta)
	evs = append(evs, evs2...)
	ref := deUnaVezCopilot(t, "cortado/paso2.jsonl")
	if len(evs) != len(ref) {
		t.Fatalf("en dos pasadas, %d eventos; de una vez, %d", len(evs), len(ref))
	}
	// La referencia, además, es la que el fichero dice: X antes del `resume`, Y después (remate del analyze, M5).
	var r *project.Resolutor
	if x, y := r.Derivar("/sintetico/x", salCopilot), r.Derivar("/sintetico/y", salCopilot); ref[0].ProjectRef != x || ref[1].ProjectRef != y {
		t.Errorf("de una vez, project_ref = (%q, %q); se esperaba (%q, %q)", ref[0].ProjectRef, ref[1].ProjectRef, x, y)
	}
	for i := range ref {
		if evs[i].EventID != ref[i].EventID || evs[i].SessionRef != ref[i].SessionRef || evs[i].ProjectRef != ref[i].ProjectRef {
			t.Errorf("evento %d: en dos pasadas %+v; de una vez %+v", i, evs[i], ref[i])
		}
	}
}

// (18) · P-2: la línea final sin LF no se consume; cuando se completa, sale en la pasada siguiente. Una sola hoja.
func TestLeerFicheroCopilot_SinLF(t *testing.T) {
	st, ruta := state.New(), ficheroCopilot(t, "sin_lf.jsonl")
	evs1, _, _ := pasarCopilot(t, st, ruta)
	ponerFicheroCopilot(t, "sin_lf_completo.jsonl", ruta)
	evs2, _, _ := pasarCopilot(t, st, ruta)
	if len(evs1) != 0 || len(evs2) != 1 {
		t.Errorf("sin LF: %d eventos; completada: %d; se esperaba 0 y 1", len(evs1), len(evs2))
	}
}

// (19) · P-2, SC-014: un fichero truncado por debajo de su offset se relee, con los MISMOS `event_id`.
func TestLeerFicheroCopilot_Truncado(t *testing.T) {
	st, ruta := state.New(), ficheroCopilot(t, "truncado/largo.jsonl")
	evs1, _, _ := pasarCopilot(t, st, ruta)
	ponerFicheroCopilot(t, "truncado/corto.jsonl", ruta)
	evs2, _, _ := pasarCopilot(t, st, ruta)
	if len(evs1) != 2 || len(evs2) != 1 || evs2[0].EventID != evs1[0].EventID {
		t.Errorf("largo: %d eventos; corto: %d; se esperaba 2, 1 y el mismo event_id", len(evs1), len(evs2))
	}
}

// (20) · El mismo `usageId` dos veces en una pasada: un evento y «repetidas 1».
func TestLeerFicheroCopilot_Repetida(t *testing.T) {
	if evs := deUnaVezCopilot(t, "repetida.jsonl"); len(evs) != 1 {
		t.Errorf("%d eventos; se esperaba 1", len(evs))
	}
}

// (21) · P-3: sin `start`, `session_ref` vacío; sin `cwd`, `project_ref` vacío.
func TestLeerFicheroCopilot_SinContexto(t *testing.T) {
	t.Run("sin_start", func(t *testing.T) {
		for _, ev := range deUnaVezCopilot(t, "sin_start.jsonl") {
			if ev.SessionRef != "" || ev.ProjectRef != "" {
				t.Errorf("session_ref %q, project_ref %q; se esperaban vacíos", ev.SessionRef, ev.ProjectRef)
			}
		}
	})
	t.Run("sin_cwd", func(t *testing.T) {
		for _, ev := range deUnaVezCopilot(t, "sin_cwd.jsonl") {
			if ev.ProjectRef != "" {
				t.Errorf("project_ref %q; se esperaba vacío", ev.ProjectRef)
			}
		}
	})
}

// (22) · Una línea corrupta se avisa y no corta el fichero.
func TestLeerFicheroCopilot_Corrupta(t *testing.T) {
	evs, p, avisos := pasarCopilot(t, state.New(), ficheroCopilot(t, "corrupta.jsonl"))
	if len(evs) != 2 || strings.Count(avisos, "skip (línea corrupta):") != 1 || p.Resumen() != resumenEsperadoCopilot(t, "corrupta.jsonl") {
		t.Errorf("%d eventos, avisos %q, resumen %q; se esperaban 2 y un aviso", len(evs), avisos, p.Resumen())
	}
}

// (23) · SC-010: una línea de cada clase, la línea de resumen literal y la identidad de FR-020.
func TestLeerFicheroCopilot_SC010Clases(t *testing.T) {
	_, p, _ := pasarCopilot(t, state.New(), ficheroCopilot(t, "clases.jsonl"))
	want := "copilot_cli: registros 6 · eventos 2 · compactaciones 1 · repetidas 1 · sin identificador 1 · incoherentes 2 · sin modelo 1"
	if got := p.Resumen(); got != want || p.Registros != p.Eventos+p.Repetidas+p.SinIdentificador+p.Incoherentes {
		t.Errorf("resumen\n  %s\nse esperaba\n  %s", got, want)
	}
}

// (24) · Un subagente síncrono (muestra 3): sus registros salen con la sesión del `start` y con su `usage.model`, el cobrado;
// nunca con el modelo que declara `subagent.*`.
func TestLeerFicheroCopilot_Subagente(t *testing.T) {
	evs := deUnaVezCopilot(t, "subagente.jsonl")
	t.Run("sesion", func(t *testing.T) {
		sesion := event.Ref(salCopilot, "s-000000000000000000000001")
		for i, ev := range evs {
			if ev.SessionRef != sesion {
				t.Errorf("evento %d: session_ref %q; se esperaba el del start, %q", i, ev.SessionRef, sesion)
			}
		}
	})
	t.Run("modelo", func(t *testing.T) {
		for i, ev := range evs {
			if ev.Model != "modelo-sintetico-1" {
				t.Errorf("evento %d: modelo %q; se esperaba el del registro, modelo-sintetico-1", i, ev.Model)
			}
		}
	})
}

// (25) · Una sesión matada (muestra 2, M1): sale el registro del 1.er prompt, y nada de la llamada cortada.
func TestLeerFicheroCopilot_Matada(t *testing.T) {
	if evs := deUnaVezCopilot(t, "matada.jsonl"); len(evs) != 1 {
		t.Errorf("%d eventos; se esperaba 1", len(evs))
	}
}

// (26) · D-6, FR-006: leer no cambia nada en la carpeta: ni un byte, ni un fichero, ni un `.lock`. NACE VERDE con el stub
// (que no escribe): lo valida m29.
func TestLeerFicheroCopilot_SoloLectura(t *testing.T) {
	raiz := copiarArbolCopilot(t, filepath.Join("ficheros", "solo_lectura"))
	foto := func() map[string]string {
		m := map[string]string{}
		_ = filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				b, _ := os.ReadFile(p)
				m[p] = fmt.Sprintf("%x", md5.Sum(b))
			}
			return nil
		})
		return m
	}
	antes := foto()
	pasarCopilot(t, state.New(), filepath.Join(raiz, "s-000000000000000000000001", "events.jsonl"))
	if despues := foto(); !maps.Equal(antes, despues) {
		t.Errorf("la carpeta cambió al leer:\n  antes   %v\n  después %v", antes, despues)
	}
}
