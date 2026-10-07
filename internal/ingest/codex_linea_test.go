package ingest

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
)

// P-008 B2 · Una línea de Codex. Fixtures SINTÉTICOS en testdata/codex/ (disciplina 11): identificadores
// inventados y, en `registro_simple.jsonl`, las partidas ya publicadas en `soporte/descubrimiento.md` (F7).

const (
	salCodex       = "sal-de-prueba"
	modeloDePrueba = "modelo-sintetico"
	cwdSintetico   = "/tmp/proyecto-sintetico"
)

func contextoCodexDePrueba() ContextoCodex {
	return ContextoCodex{
		Context: Context{Salt: salCodex, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba", OrgID: "org-de-prueba",
			AgentVersion: "test"},
		DelTurno: func(string) (string, string) { return modeloDePrueba, cwdSintetico },
	}
}

func lineaDeFixture(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "codex", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return bytes.TrimRight(b, "\n")
}

// leerEventoCodex exige que la línea sea un evento: si no, el test no puede seguir.
func leerEventoCodex(t *testing.T, nombre string) event.Event {
	t.Helper()
	ev, clase, err := LineaCodex(lineaDeFixture(t, nombre), contextoCodexDePrueba())
	if err != nil {
		t.Fatalf("LineaCodex(%s): error %v", nombre, err)
	}
	if clase != EventoCodex || ev == nil {
		t.Fatalf("LineaCodex(%s): clase %d, evento %v; se esperaba un evento", nombre, clase, ev)
	}
	return *ev
}

// (4) · FR-006, FR-010: un registro es un evento, con las partidas de M-4. Y ninguna otra línea lo es.
func TestLineaCodex_UnRegistroEsUnEvento(t *testing.T) {
	t.Run("registro", func(t *testing.T) {
		ev := leerEventoCodex(t, "registro_simple.jsonl")
		got := [4]int{ev.TokensInput, ev.TokensCacheCreation, ev.TokensCacheRead, ev.TokensOutput}
		if want := [4]int{2823, 0, 11008, 5}; got != want {
			t.Errorf("partidas (entrada, escritura, lectura, salida) = %v; se esperaba %v (13 831 − 11 008 − 0)", got, want)
		}
		if id, _ := derivarEventIDCodex("r-000000000000000000000001"); ev.EventID != id {
			t.Errorf("event_id = %q; el del contrato es %q", ev.EventID, id)
		}
		if ev.Model != modeloDePrueba || ev.ProjectRef != project.Derivar(cwdSintetico, salCodex) {
			t.Errorf("modelo %q y project_ref %q; se esperaban los del turno", ev.Model, ev.ProjectRef)
		}
		if ev.SchemaVersion != event.SchemaVersion || ev.AgentVersion != "test" || ev.DevID != "dev-de-prueba" ||
			ev.OrgID != "org-de-prueba" || ev.MachineRef != event.Ref(salCodex, "maquina-de-prueba") {
			t.Errorf("datos locales del evento: %+v", ev)
		}
	})
	t.Run("otra_linea_no_es_registro", func(t *testing.T) {
		ev, clase, err := LineaCodex(lineaDeFixture(t, "no_registro.jsonl"), contextoCodexDePrueba())
		if err != nil || clase != NoEsRegistro || ev != nil {
			t.Errorf("una línea turn_context: (%v, clase %d, %v); se esperaba (nil, NoEsRegistro, nil)", ev, clase, err)
		}
	})
}

// (5) · SC-008: la escritura de caché va DENTRO de la entrada (FASE 0 (c)), y se resta.
func TestLineaCodex_EscrituraDeCacheDentroDeLaEntrada(t *testing.T) {
	ev := leerEventoCodex(t, "registro_escritura.jsonl")
	got := [4]int{ev.TokensInput, ev.TokensCacheCreation, ev.TokensCacheRead, ev.TokensOutput}
	if want := [4]int{0, 60, 40, 10}; got != want {
		t.Errorf("100 / 40 / 60 / 10 → %v; se esperaba %v (entrada sin caché = 100 − 40 − 60)", got, want)
	}
}

// (6) · FR-027 y E-3: un registro incoherente no se emite y se clasifica como tal.
func TestLineaCodex_Incoherente(t *testing.T) {
	for _, c := range []struct{ hoja, fixture string }{
		{"cache_y_escritura_mayores_que_la_entrada", "incoherente_cache_mayor.jsonl"},
		{"sin_usage", "incoherente_sin_usage.jsonl"},
		{"partida_negativa", "incoherente_negativa.jsonl"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			ev, clase, err := LineaCodex(lineaDeFixture(t, c.fixture), contextoCodexDePrueba())
			if err != nil || clase != Incoherente || ev != nil {
				t.Errorf("(%v, clase %d, %v); se esperaba (nil, Incoherente, nil)", ev, clase, err)
			}
		})
	}
}

// (7) · FR-009: sin `response_id`, o con uno vacío, no hay evento.
func TestLineaCodex_SinIdentificador(t *testing.T) {
	for _, c := range []struct{ hoja, fixture string }{
		{"ausente", "sin_identificador_ausente.jsonl"},
		{"vacio", "sin_identificador_vacio.jsonl"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			ev, clase, err := LineaCodex(lineaDeFixture(t, c.fixture), contextoCodexDePrueba())
			if err != nil || clase != SinIdentificador || ev != nil {
				t.Errorf("(%v, clase %d, %v); se esperaba (nil, SinIdentificador, nil)", ev, clase, err)
			}
		})
	}
}

// (8) · FR-013, D-3: `tool = codex` y sin coste, siempre.
func TestLineaCodex_SinCoste(t *testing.T) {
	ev := leerEventoCodex(t, "registro_simple.jsonl")
	if ev.Tool != "codex" || ev.CostUSD != 0 || ev.CostAvailable {
		t.Errorf("tool %q, cost_usd %v, cost_available %t; se esperaba codex, 0 y false", ev.Tool, ev.CostUSD, ev.CostAvailable)
	}
}

// (9) · FR-012: `occurred_at` es la marca de la línea del registro.
func TestLineaCodex_MomentoDelRegistro(t *testing.T) {
	ev := leerEventoCodex(t, "registro_simple.jsonl")
	if want := time.Date(2026, 10, 7, 0, 0, 1, 250_000_000, time.UTC); !ev.OccurredAt.Equal(want) {
		t.Errorf("occurred_at = %v; se esperaba %v", ev.OccurredAt, want)
	}
}

// (10) · FR-015, P-6: `session_ref` sale del `session_id` del registro, con sal. En el fixture, `session_id`,
// `thread_id` y `turn_id` son distintos.
func TestLineaCodex_SesionDelRegistro(t *testing.T) {
	ev := leerEventoCodex(t, "registro_simple.jsonl")
	if want := event.Ref(salCodex, "s-sintetica-raiz"); ev.SessionRef != want {
		t.Errorf("session_ref = %q; se esperaba event.Ref(sal, session_id) = %q", ev.SessionRef, want)
	}
}

// T013 · FR-016: ningún identificador del proveedor viaja en el evento, ni entero ni en fragmento. Nace verde
// (se escribe tras el verde de T012); lo valida M-B2a.
func TestLineaCodex_NingunIdentificadorDelProveedorEnElEvento(t *testing.T) {
	ev := leerEventoCodex(t, "centinelas.jsonl")
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if strings.Contains(string(b), "CENTINELA") {
		t.Errorf("el evento lleva un identificador del proveedor: %s", b)
	}
}
