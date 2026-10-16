package ingest

import "testing"

// P-012 B2 · El `event_id` de Kimi Code (contracts/event-id-kimi.md, FR-011). Los vectores son las CADENAS DEL CONTRATO,
// calculadas por una implementación independiente (Python `hashlib` + `struct`): si la función no las da, la que está mal es la
// función. Cada hoja afirma su valor EXACTO.

const (
	sesionKimi1 = "session_sintetica-0001"
	sesionKimi2 = "session_sintetica-0002"

	vectorKimiUno          = "e1ad45be19ab01a906be10cf22f4e75c"
	vectorKimiDos          = "c7f5518b2da765df74f784f4a5737d40"
	vectorKimiSubagente    = "70ee1df9071f3a49800d62483b50dcb3"
	vectorKimiForkPropio   = "17456f5b3d83b75590d3f65f21a9a267"
	vectorKimiFronteraABC  = "eccac1f146057198b16b839f837fc42b"
	vectorKimiFronteraABxC = "4adb17be1c8f18ab3a2b7b322915fffd"
	vectorKimiUnoEnCodex   = "a37c3049e29731d108347b52ed3e0b5e"
)

// (4) · Los siete vectores normativos del contrato.
func TestEventIDKimi_Vectores(t *testing.T) {
	for _, c := range []struct {
		hoja, s, a string
		n          int
		quiero     string
	}{
		{"uno", sesionKimi1, "main", 1, vectorKimiUno},
		{"dos", sesionKimi1, "main", 2, vectorKimiDos},
		{"subagente", sesionKimi1, "agent-0", 1, vectorKimiSubagente},
		{"fork_propio", sesionKimi2, "main", 3, vectorKimiForkPropio},
		{"frontera_a_bc", "a", "bc", 1, vectorKimiFronteraABC},
		{"frontera_ab_c", "ab", "c", 1, vectorKimiFronteraABxC},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if got := derivarEventIDKimi(c.s, c.a, c.n); got != c.quiero {
				t.Errorf("derivarEventIDKimi(%q, %q, %d) = %q; se esperaba %s", c.s, c.a, c.n, got, c.quiero)
			}
		})
	}
	t.Run("codex", func(t *testing.T) {
		// El 1.º en el espacio de Codex da otro valor: el espacio de nombres separa las herramientas.
		if got := derivarEventIDKimi(sesionKimi1, "main", 1); got != vectorKimiUno || got == vectorKimiUnoEnCodex {
			t.Errorf("= %q; se esperaba %s, distinto del de Codex %s", got, vectorKimiUno, vectorKimiUnoEnCodex)
		}
	})
}
