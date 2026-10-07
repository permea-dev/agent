package ingest

import "testing"

// Vectores NORMATIVOS de `specs/008-lector-codex/contracts/event-id-codex.md` §Vectores de prueba,
// calculados el 2026-10-07 con una implementación independiente (Python `hashlib` + `struct`).
// Identificadores sintéticos, sin la forma de ninguno real.
const (
	respuestaSintetica1 = "r-000000000000000000000001"
	respuestaSintetica2 = "r-000000000000000000000002"

	// El mismo valor que respuestaSintetica1, derivado en el espacio de nombres de Claude Code como
	// `message.id` solo: ["permea/event_id/v1", "claude_code", "solo_message_id", valor].
	vectorClaudeCodeMismoValor = "128d67bd072f4488bc3854e82bbcaeb2"
)

// (1) · Los dos vectores del contrato, byte a byte.
func TestEventIDCodex_VectoresDelContrato(t *testing.T) {
	casos := []struct {
		nombre, respuesta, esperado string
	}{
		{"respuesta", respuestaSintetica1, "31439e3953a0916dde2f98b748ceb985"},
		{"otra_respuesta", respuestaSintetica2, "e2c798be5f23e44680ab2c70690f8cf9"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got, ok := derivarEventIDCodex(c.respuesta)
			if !ok {
				t.Fatalf("derivarEventIDCodex(%q): ok = false; se esperaba un event_id", c.respuesta)
			}
			if got != c.esperado {
				t.Errorf("derivarEventIDCodex(%q) = %q; el contrato dice %q", c.respuesta, got, c.esperado)
			}
		})
	}
}

// (2) · El espacio de nombres separa las herramientas: el mismo valor da otro `event_id` en Claude Code.
// El literal se comprueba antes contra la derivación de 006, sin tocarla.
func TestEventIDCodex_EspacioDeNombresPropio(t *testing.T) {
	if claude := hashEventID(tipoSoloMessageID, respuestaSintetica1); claude != vectorClaudeCodeMismoValor {
		t.Fatalf("precondición: la derivación de 006 da %q para el mismo valor; el contrato dice %q", claude, vectorClaudeCodeMismoValor)
	}
	got, ok := derivarEventIDCodex(respuestaSintetica1)
	if !ok {
		t.Fatalf("derivarEventIDCodex(%q): ok = false; se esperaba un event_id", respuestaSintetica1)
	}
	if got == vectorClaudeCodeMismoValor {
		t.Errorf("el event_id de Codex coincide con el de Claude Code para el mismo valor (%q)", got)
	}
}

// (3) · Sin `response_id` no hay nada estable de lo que derivar: no se emite (contrato §Derivación).
func TestEventIDCodex_SinRespuestaNoHayEventID(t *testing.T) {
	got, ok := derivarEventIDCodex("")
	if ok || got != "" {
		t.Errorf(`derivarEventIDCodex("") = (%q, %t); se esperaba ("", false)`, got, ok)
	}
}
