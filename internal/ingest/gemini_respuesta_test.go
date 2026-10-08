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

// P-009 B2 · Una aparición de un mensaje de Gemini CLI. Fixtures SINTÉTICOS en testdata/gemini/ (disciplina 11):
// identificadores `m-0000…` y `s-0000…`, sin forma de UUID, y partidas inventadas.

const (
	salGemini          = "sal-de-prueba-gemini"
	modeloGeminiPrueba = "modelo-sintetico-gemini"
	sesionGeminiPrueba = "s-000000000000000000000001"
	raizGeminiPrueba   = "/tmp/proyecto-sintetico-gemini"
)

func contextoGeminiDePrueba() ContextoGemini {
	return ContextoGemini{
		Context: Context{Salt: salGemini, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba", OrgID: "org-de-prueba",
			AgentVersion: "test"},
		SessionID:   sesionGeminiPrueba,
		ProjectRoot: raizGeminiPrueba,
	}
}

func aparicionDeFixture(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "gemini", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return bytes.TrimRight(b, "\n")
}

// leerEventoGemini exige que la aparición sea un evento: si no, el test no puede seguir.
func leerEventoGemini(t *testing.T, nombre string, ctx ContextoGemini) (event.Event, MarcasGemini) {
	t.Helper()
	ev, clase, marcas, err := RespuestaGemini(aparicionDeFixture(t, nombre), ctx)
	if err != nil {
		t.Fatalf("RespuestaGemini(%s): error %v", nombre, err)
	}
	if clase != EventoGemini || ev == nil {
		t.Fatalf("RespuestaGemini(%s): clase %d, evento %v; se esperaba un evento", nombre, clase, ev)
	}
	return *ev, marcas
}

// claseDe devuelve la clase de una aparición, que no debe ser un error ni producir evento.
func claseDe(t *testing.T, nombre string) ClaseGemini {
	t.Helper()
	ev, clase, _, err := RespuestaGemini(aparicionDeFixture(t, nombre), contextoGeminiDePrueba())
	if err != nil {
		t.Fatalf("RespuestaGemini(%s): error %v; una aparición que es JSON no es corrupta", nombre, err)
	}
	if ev != nil && clase != EventoGemini {
		t.Fatalf("RespuestaGemini(%s): clase %d con evento", nombre, clase)
	}
	return clase
}

// (4) · SC-007, FR-010 (D-2): 100/40/5/7/3 → 65/40/0/10. Y lo que no es respuesta no se confunde con incoherente.
func TestRespuestaGemini_PartidasD2(t *testing.T) {
	t.Run("partidas", func(t *testing.T) {
		ev, marcas := leerEventoGemini(t, "respuesta_partidas.jsonl", contextoGeminiDePrueba())
		got := [4]int{ev.TokensInput, ev.TokensCacheRead, ev.TokensCacheCreation, ev.TokensOutput}
		if want := [4]int{65, 40, 0, 10}; got != want {
			t.Errorf("partidas (entrada, lectura, escritura, salida) = %v; se esperaba %v (100 − 40 + 5; 7 + 3)", got, want)
		}
		if id, _ := derivarEventIDGemini("m-000000000000000000000001"); ev.EventID != id {
			t.Errorf("event_id = %q; el del contrato es %q", ev.EventID, id)
		}
		if ev.Model != modeloGeminiPrueba || marcas != (MarcasGemini{}) {
			t.Errorf("modelo %q y marcas %+v; se esperaban el del mensaje y ninguna", ev.Model, marcas)
		}
		if ev.SchemaVersion != event.SchemaVersion || ev.AgentVersion != "test" || ev.DevID != "dev-de-prueba" ||
			ev.OrgID != "org-de-prueba" || ev.MachineRef != event.Ref(salGemini, "maquina-de-prueba") {
			t.Errorf("datos locales del evento: %+v", ev)
		}
	})
	for _, c := range []struct{ hoja, fixture string }{
		{"usuario_no_es_respuesta", "no_respuesta_usuario.jsonl"},
		{"tokens_null_no_es_respuesta", "no_respuesta_tokens_null.jsonl"},
		{"sin_tokens_no_es_respuesta", "no_respuesta_sin_tokens.jsonl"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if clase := claseDe(t, c.fixture); clase != NoEsRespuestaGemini {
				t.Errorf("%s: clase %d; se esperaba NoEsRespuestaGemini (%d)", c.fixture, clase, NoEsRespuestaGemini)
			}
		})
	}
}

// (5) · D-2: una partida ausente vale 0, y se emite.
func TestRespuestaGemini_PartidaAusenteValeCero(t *testing.T) {
	ev, _ := leerEventoGemini(t, "respuesta_partida_ausente.jsonl", contextoGeminiDePrueba())
	got := [4]int{ev.TokensInput, ev.TokensCacheRead, ev.TokensCacheCreation, ev.TokensOutput}
	if want := [4]int{10, 0, 0, 2}; got != want {
		t.Errorf("partidas = %v; se esperaba %v", got, want)
	}
}

// (6) · FR-019: incoherente, se cuenta y no se emite. Seis hojas.
func TestRespuestaGemini_Incoherente(t *testing.T) {
	for _, c := range []struct{ hoja, fixture string }{
		{"tokens_no_objeto", "incoherente_tokens_no_objeto.jsonl"},
		{"partida_no_numerica", "incoherente_partida_no_numerica.jsonl"},
		{"partida_negativa", "incoherente_negativa.jsonl"},
		{"cache_mayor_que_entrada", "incoherente_cache_mayor.jsonl"},
		{"sin_timestamp", "incoherente_sin_timestamp.jsonl"},
		{"timestamp_mal_formado", "incoherente_timestamp_mal_formado.jsonl"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if clase := claseDe(t, c.fixture); clase != IncoherenteGemini {
				t.Errorf("%s: clase %d; se esperaba IncoherenteGemini (%d)", c.fixture, clase, IncoherenteGemini)
			}
		})
	}
}

// (7) · FR-009: sin `id` textual no vacío, «sin identificador» (y no «incoherente» ni error).
func TestRespuestaGemini_SinIdentificador(t *testing.T) {
	for _, c := range []struct{ hoja, fixture string }{
		{"ausente", "sin_identificador_ausente.jsonl"},
		{"vacio", "sin_identificador_vacio.jsonl"},
		{"no_textual", "sin_identificador_no_textual.jsonl"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if clase := claseDe(t, c.fixture); clase != SinIdentificadorGemini {
				t.Errorf("%s: clase %d; se esperaba SinIdentificadorGemini (%d)", c.fixture, clase, SinIdentificadorGemini)
			}
		})
	}
}

// (8) · FR-013, D-4: `tool = "gemini"`, sin coste y sin coste disponible.
func TestRespuestaGemini_SinCoste(t *testing.T) {
	ev, _ := leerEventoGemini(t, "respuesta_partidas.jsonl", contextoGeminiDePrueba())
	if ev.Tool != "gemini" || ev.CostUSD != 0 || ev.CostAvailable {
		t.Errorf("tool %q, cost_usd %v, cost_available %t; se esperaba gemini, 0 y false", ev.Tool, ev.CostUSD, ev.CostAvailable)
	}
}

// (9) · FR-012: `occurred_at` es el `timestamp` del mensaje.
func TestRespuestaGemini_MomentoDelMensaje(t *testing.T) {
	ev, _ := leerEventoGemini(t, "respuesta_partidas.jsonl", contextoGeminiDePrueba())
	if want := time.Date(2026, 10, 8, 0, 0, 1, 250_000_000, time.UTC); !ev.OccurredAt.Equal(want) {
		t.Errorf("occurred_at = %v; se esperaba %v", ev.OccurredAt, want)
	}
}

// (10) · FR-011, P-6: sin `model`, el evento sale con modelo vacío y la marca «sin modelo».
func TestRespuestaGemini_SinModelo(t *testing.T) {
	ev, marcas := leerEventoGemini(t, "sin_modelo.jsonl", contextoGeminiDePrueba())
	if ev.Model != "" || !marcas.SinModelo || marcas.TotalDescuadrado {
		t.Errorf("modelo %q, marcas %+v; se esperaba modelo vacío y sólo «sin modelo»", ev.Model, marcas)
	}
}

// (11) · FR-019, P-5: un `total` descuadrado se emite y se marca. Sin `total`, no hay marca.
func TestRespuestaGemini_TotalDescuadrado(t *testing.T) {
	t.Run("descuadrado", func(t *testing.T) {
		_, marcas := leerEventoGemini(t, "total_descuadrado.jsonl", contextoGeminiDePrueba())
		if !marcas.TotalDescuadrado || marcas.SinModelo {
			t.Errorf("marcas %+v; se esperaba sólo «total descuadrado»", marcas)
		}
	})
	t.Run("sin_total", func(t *testing.T) {
		if _, marcas := leerEventoGemini(t, "respuesta_partida_ausente.jsonl", contextoGeminiDePrueba()); marcas.TotalDescuadrado {
			t.Errorf("sin `total`, marcas %+v; no debía marcarse", marcas)
		}
	})
}

// (12) · FR-014, FR-015: `session_ref` con sal, `project_ref` del texto de `.project_root`, y vacío sin él.
func TestRespuestaGemini_Referencias(t *testing.T) {
	t.Run("session_ref", func(t *testing.T) {
		ev, _ := leerEventoGemini(t, "respuesta_partidas.jsonl", contextoGeminiDePrueba())
		if want := event.Ref(salGemini, sesionGeminiPrueba); ev.SessionRef != want || want == "" {
			t.Errorf("session_ref = %q; se esperaba Ref(sal, sessionId) = %q", ev.SessionRef, want)
		}
	})
	t.Run("project_ref", func(t *testing.T) {
		ev, _ := leerEventoGemini(t, "respuesta_partidas.jsonl", contextoGeminiDePrueba())
		if want := project.Derivar(raizGeminiPrueba, salGemini); ev.ProjectRef != want || want == "" {
			t.Errorf("project_ref = %q; se esperaba Derivar(.project_root, sal) = %q", ev.ProjectRef, want)
		}
	})
	t.Run("sin_project_root", func(t *testing.T) {
		ctx := contextoGeminiDePrueba()
		ctx.ProjectRoot = ""
		if ev, _ := leerEventoGemini(t, "respuesta_partidas.jsonl", ctx); ev.ProjectRef != "" {
			t.Errorf("sin .project_root, project_ref = %q; se esperaba vacío", ev.ProjectRef)
		}
	})
}

// T013 · FR-017: ningún identificador del proveedor ni el texto de `.project_root` llegan al evento.
func TestRespuestaGemini_NadaDelProveedorEnElEvento(t *testing.T) {
	ctx := contextoGeminiDePrueba()
	ctx.SessionID = "s-CENTINELA-SESION-000001"
	ctx.ProjectRoot = "/tmp/CENTINELA-RAIZ-000001"
	ev, _ := leerEventoGemini(t, "centinelas.jsonl", ctx)
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	for _, centinela := range []string{"CENTINELA", "m-CENTINELA-ID", "s-CENTINELA-SESION", "h-CENTINELA-HASH", "CENTINELA-RAIZ"} {
		if strings.Contains(string(b), centinela) {
			t.Errorf("el evento contiene %q: %s", centinela, b)
		}
	}
}
