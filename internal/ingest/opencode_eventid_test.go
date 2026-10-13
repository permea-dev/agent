package ingest

import "testing"

// Vectores NORMATIVOS de `specs/010-lector-opencode/contracts/event-id-opencode.md` §Vectores de prueba, calculados el
// 2026-10-11 con una implementación independiente (Python `hashlib` + `struct`). Identificadores sintéticos, sin la forma
// de ningún identificador real.
const (
	llamadaSintetica1 = "o-000000000000000000000001"
	llamadaSintetica2 = "o-000000000000000000000002"

	// El mismo valor que llamadaSintetica1, derivado en el espacio de nombres de Gemini:
	// ["permea/event_id/v1", "gemini", "respuesta", valor].
	vectorGeminiMismoValor = "a53e1c53e69102bc0411f71d2e711a70"
)

// (11) · Los dos vectores del contrato, byte a byte.
func TestEventIDOpenCode_VectoresDelContrato(t *testing.T) {
	casos := []struct {
		nombre, id, esperado string
	}{
		{"llamada", llamadaSintetica1, "e771a4833f173d08b2d3322627083201"},
		{"otra_llamada", llamadaSintetica2, "f7392add336fb033f74ac241cb4571a3"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got, ok := derivarEventIDOpenCode(c.id)
			if !ok {
				t.Fatalf("derivarEventIDOpenCode(%q): ok = false; se esperaba un event_id", c.id)
			}
			if got != c.esperado {
				t.Errorf("derivarEventIDOpenCode(%q) = %q; el contrato dice %q", c.id, got, c.esperado)
			}
		})
	}
}

// (12) · El espacio de nombres separa las herramientas: el mismo valor da otro `event_id` en Gemini. El literal se
// comprueba antes contra la derivación de 009, sin tocarla.
func TestEventIDOpenCode_EspacioDeNombresPropio(t *testing.T) {
	if g, _ := derivarEventIDGemini(llamadaSintetica1); g != vectorGeminiMismoValor {
		t.Fatalf("precondición: derivarEventIDGemini(%q) = %q; el contrato dice %q", llamadaSintetica1, g, vectorGeminiMismoValor)
	}
	got, ok := derivarEventIDOpenCode(llamadaSintetica1)
	if !ok || got == vectorGeminiMismoValor {
		t.Errorf("derivarEventIDOpenCode(%q) = %q, ok %v; debe existir y ser distinto del de Gemini (%q)",
			llamadaSintetica1, got, ok, vectorGeminiMismoValor)
	}
}

// (13) · Sin `id` no queda nada estable de lo que derivar: no hay event_id y la llamada no se emite.
func TestEventIDOpenCode_SinIDNoHayEventID(t *testing.T) {
	if got, ok := derivarEventIDOpenCode(""); ok || got != "" {
		t.Errorf(`derivarEventIDOpenCode("") = %q, ok %v; se esperaba "", false`, got, ok)
	}
}
