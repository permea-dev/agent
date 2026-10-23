package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
)

// ═══ 013 B4 · LA AGREGACIÓN — (7) a (12) de plan §Bloques ═════════════════════════════════════════════════════════════
//
// `agregar` es pura *(D-013-P5)*: los eventos, hechos a mano, y el disco, inyectado como `origen` y `existe`. Las identidades
// son 64 hex inventados; las rutas, `/sintetico/…`.

// refDe es una identidad de 64 hex hecha de una letra repetida: legible y ordenable.
func refDe(letra string) string { return strings.Repeat(letra, 64) }

// evLista es un evento con sólo lo que lee la agregación.
func evLista(ref, id, tool string, cuando time.Time) event.Event {
	return event.Event{ProjectRef: ref, EventID: id, Tool: tool, OccurredAt: cuando}
}

// dia es un instante en UTC del día `d` de octubre de 2026, a la hora `h`.
func dia(d, h int) time.Time { return time.Date(2026, 10, d, h, 0, 0, 0, time.UTC) }

// discoDeLista inyecta el disco: el origen de cada identidad y qué rutas existen.
type discoDeLista struct {
	origenes map[string]project.Origen
	existen  map[string]bool
}

func (d discoDeLista) origen(ref string) (project.Origen, bool) {
	o, ok := d.origenes[ref]
	return o, ok
}

func (d discoDeLista) existe(ruta string) bool { return d.existen[ruta] }

// discoConRepos da a cada identidad una raíz `/sintetico/<letra>` que existe y con repositorio.
func discoConRepos(letras ...string) discoDeLista {
	d := discoDeLista{origenes: map[string]project.Origen{}, existen: map[string]bool{}}
	for _, l := range letras {
		d.origenes[refDe(l)] = project.Origen{Ruta: "/sintetico/" + l, HuboRaiz: true}
		d.existen["/sintetico/"+l] = true
	}
	return d
}

// filaDe es la fila esperada de una identidad de `discoConRepos`.
func filaDe(letra string, eventos int, desde, hasta time.Time, herramientas ...string) fila {
	return fila{identidad: refDe(letra), ruta: "/sintetico/" + letra, herramientas: herramientas, eventos: eventos, desde: desde, hasta: hasta}
}

func exigirFilas(t *testing.T, got, quiere []fila) {
	t.Helper()
	if !reflect.DeepEqual(got, quiere) {
		t.Errorf("filas:\n  %+v\nse esperaban:\n  %+v", got, quiere)
	}
}

// (7) · P-2, FR-012: se cuentan `event_id` distintos; el primero que llega gana, y un repetido no aporta nada a otra identidad.
func TestAgregar_EventIDDistintos(t *testing.T) {
	d := discoConRepos("a", "b")
	t.Run("mismo_id_dos_veces", func(t *testing.T) {
		got := agregar([]event.Event{
			evLista(refDe("a"), "x", "claude_code", dia(3, 8)),
			evLista(refDe("a"), "x", "claude_code", dia(3, 8)),
			evLista(refDe("a"), "y", "claude_code", dia(4, 9)),
		}, d.origen, d.existe)
		exigirFilas(t, got, []fila{filaDe("a", 2, dia(3, 8), dia(4, 9), "Claude Code")})
	})
	t.Run("en_otra_identidad", func(t *testing.T) {
		got := agregar([]event.Event{
			evLista(refDe("a"), "x", "claude_code", dia(3, 8)),
			evLista(refDe("b"), "x", "codex", dia(1, 1)), // repetido: ni cuenta, ni Codex, ni su fecha
			evLista(refDe("b"), "y", "gemini", dia(5, 5)),
		}, d.origen, d.existe)
		exigirFilas(t, got, []fila{
			filaDe("a", 1, dia(3, 8), dia(3, 8), "Claude Code"),
			filaDe("b", 1, dia(5, 5), dia(5, 5), "Gemini CLI"),
		})
	})
	t.Run("solo_repetidos", func(t *testing.T) {
		got := agregar([]event.Event{
			evLista(refDe("a"), "x", "claude_code", dia(3, 8)),
			evLista(refDe("b"), "x", "codex", dia(3, 8)),
		}, d.origen, d.existe)
		exigirFilas(t, got, []fila{filaDe("a", 1, dia(3, 8), dia(3, 8), "Claude Code")})
	})
}

// (8) · P-1, FR-011: un evento sin identidad no da fila ni cuenta.
func TestAgregar_SinIdentidad(t *testing.T) {
	d := discoConRepos("a")
	got := agregar([]event.Event{
		evLista("", "p", "claude_code", dia(2, 2)),
		evLista(refDe("a"), "q", "claude_code", dia(3, 3)),
		evLista("", "r", "kimi_code", dia(4, 4)),
	}, d.origen, d.existe)
	exigirFilas(t, got, []fila{filaDe("a", 1, dia(3, 3), dia(3, 3), "Claude Code")})
}

// (9) · FR-013: las fechas son las del instante en UTC. A las 01:30 en `+02:00` del 6 de octubre son las 23:30 UTC del 5.
func TestAgregar_FechasUTC(t *testing.T) {
	d := discoConRepos("a")
	madrid := time.FixedZone("+02:00", 2*3600)
	tardio := time.Date(2026, 10, 7, 0, 10, 0, 0, madrid)   // 2026-10-06 22:10 UTC
	temprano := time.Date(2026, 10, 6, 1, 30, 0, 0, madrid) // 2026-10-05 23:30 UTC
	got := agregar([]event.Event{
		evLista(refDe("a"), "x", "claude_code", tardio),
		evLista(refDe("a"), "y", "claude_code", temprano),
	}, d.origen, d.existe)
	if len(got) != 1 {
		t.Fatalf("filas: %+v; se esperaba una", got)
	}
	f := got[0]
	if f.desde.Location() != time.UTC || f.hasta.Location() != time.UTC ||
		f.desde.Format("2006-01-02") != "2026-10-05" || f.hasta.Format("2006-01-02") != "2026-10-06" ||
		!f.desde.Equal(temprano) || !f.hasta.Equal(tardio) {
		t.Errorf("desde %v, hasta %v; se esperaban 2026-10-05 23:30 y 2026-10-06 22:10, en UTC", f.desde, f.hasta)
	}
}

// (10) · D-1, P-3, FR-014: los nombres comerciales, en el orden de lectura; un id desconocido, tal cual y detrás, por orden alfabético.
func TestAgregar_Herramientas(t *testing.T) {
	d := discoConRepos("a")
	conTools := func(t *testing.T, tools ...string) []string {
		t.Helper()
		var evs []event.Event
		for i, tool := range tools {
			evs = append(evs, evLista(refDe("a"), fmt.Sprintf("e%d", i), tool, dia(3, 3)))
		}
		got := agregar(evs, d.origen, d.existe)
		if len(got) != 1 {
			t.Fatalf("filas: %+v; se esperaba una", got)
		}
		return got[0].herramientas
	}
	for _, c := range []struct {
		hoja   string
		tools  []string
		quiere []string
	}{
		{"orden_y_desconocido", []string{"kimi_code", "zeta", "claude_code"}, []string{"Claude Code", "Kimi Code CLI", "zeta"}},
		{"los_seis", []string{"kimi_code", "copilot_cli", "opencode", "gemini", "codex", "claude_code"},
			[]string{"Claude Code", "Codex", "Gemini CLI", "OpenCode", "GitHub Copilot CLI", "Kimi Code CLI"}},
		{"desconocidos_alfabeticos", []string{"zeta", "alfa", "codex", "zeta"}, []string{"Codex", "alfa", "zeta"}},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if got := conTools(t, c.tools...); !reflect.DeepEqual(got, c.quiere) {
				t.Errorf("herramientas %q; se esperaban %q", got, c.quiere)
			}
		})
	}
}

// (11) · D-3, P-4, FR-016: una sola marca. Sin la carpeta, «(ya no existe)», haya habido raíz o no; con ella y sin raíz,
// «(sin repositorio)»; con las dos, nada.
func TestAgregar_Marcas(t *testing.T) {
	for _, c := range []struct {
		hoja     string
		huboRaiz bool
		existe   bool
		quiere   string
	}{
		{"con_git", true, true, ""},
		{"sin_git", false, true, "(sin repositorio)"},
		{"borrada", false, false, "(ya no existe)"},
		{"borrada_con_git", true, false, "(ya no existe)"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			d := discoDeLista{
				origenes: map[string]project.Origen{refDe("a"): {Ruta: "/sintetico/a", HuboRaiz: c.huboRaiz}},
				existen:  map[string]bool{"/sintetico/a": c.existe},
			}
			got := agregar([]event.Event{evLista(refDe("a"), "x", "claude_code", dia(3, 3))}, d.origen, d.existe)
			if len(got) != 1 || got[0].marca != c.quiere || got[0].ruta != "/sintetico/a" {
				t.Errorf("filas %+v; se esperaba una, en /sintetico/a, con la marca %q", got, c.quiere)
			}
		})
	}
}

// (12) · D-2, FR-020: por eventos, de más a menos; si empatan, por la identidad completa, de menor a mayor.
func TestAgregar_Orden(t *testing.T) {
	d := discoConRepos("a", "b", "c", "d")
	cuantos := map[string]int{"a": 7, "c": 412, "d": 1240, "b": 412}
	var evs []event.Event
	for _, l := range []string{"a", "c", "d", "b"} {
		for i := 0; i < cuantos[l]; i++ {
			evs = append(evs, evLista(refDe(l), fmt.Sprintf("%s-%04d", l, i), "claude_code", dia(3, 3)))
		}
	}
	got := agregar(evs, d.origen, d.existe)
	var orden []string
	for _, f := range got {
		orden = append(orden, f.identidad[:1]+":"+strconv.Itoa(f.eventos))
	}
	if quiere := []string{"d:1240", "b:412", "c:412", "a:7"}; !reflect.DeepEqual(orden, quiere) {
		t.Errorf("orden %q; se esperaba %q", orden, quiere)
	}
}
