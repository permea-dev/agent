package ingest

import "testing"

// Vectores NORMATIVOS de `specs/009-lector-gemini/contracts/event-id-gemini.md` §Vectores de prueba,
// calculados el 2026-10-08 con una implementación independiente (Python `hashlib` + `struct`).
// Identificadores sintéticos, sin la forma de UUID ni de ninguno real.
const (
	mensajeSintetico1 = "m-000000000000000000000001"
	mensajeSintetico2 = "m-000000000000000000000002"

	// El mismo valor que mensajeSintetico1, derivado en el espacio de nombres de Codex:
	// ["permea/event_id/v1", "codex", "respuesta", valor].
	vectorCodexMismoValor = "07d1f3ea70c0b94474fe27e18fdb89b7"
)

// (1) · Los dos vectores del contrato, byte a byte.
func TestEventIDGemini_VectoresDelContrato(t *testing.T) {
	casos := []struct {
		nombre, id, esperado string
	}{
		{"respuesta", mensajeSintetico1, "027c36e1f42d164e670770cee1a06f83"},
		{"otra_respuesta", mensajeSintetico2, "53120e5f5cfd342c9c48f37cd857c697"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got, ok := derivarEventIDGemini(c.id)
			if !ok {
				t.Fatalf("derivarEventIDGemini(%q): ok = false; se esperaba un event_id", c.id)
			}
			if got != c.esperado {
				t.Errorf("derivarEventIDGemini(%q) = %q; el contrato dice %q", c.id, got, c.esperado)
			}
		})
	}
}

// (2) · El espacio de nombres separa las herramientas: el mismo valor da otro `event_id` en Codex.
// El literal se comprueba antes contra la derivación de 008, sin tocarla.
func TestEventIDGemini_EspacioDeNombresPropio(t *testing.T) {
	if codex, _ := derivarEventIDCodex(mensajeSintetico1); codex != vectorCodexMismoValor {
		t.Fatalf("precondición: la derivación de 008 da %q para el mismo valor; el contrato dice %q", codex, vectorCodexMismoValor)
	}
	got, ok := derivarEventIDGemini(mensajeSintetico1)
	if !ok {
		t.Fatalf("derivarEventIDGemini(%q): ok = false; se esperaba un event_id", mensajeSintetico1)
	}
	if got == vectorCodexMismoValor {
		t.Errorf("el event_id de Gemini coincide con el de Codex para el mismo valor (%q)", got)
	}
}

// (3) · Sin `id` no hay nada estable de lo que derivar: no se emite (contrato §Derivación).
func TestEventIDGemini_SinIDNoHayEventID(t *testing.T) {
	got, ok := derivarEventIDGemini("")
	if ok || got != "" {
		t.Errorf(`derivarEventIDGemini("") = (%q, %t); se esperaba ("", false)`, got, ok)
	}
}
