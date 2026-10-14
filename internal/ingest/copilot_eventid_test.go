package ingest

import "testing"

// P-011 B2 · El `event_id` de Copilot CLI (contracts/event-id-copilot.md, FR-011). Los vectores son los del contrato,
// calculados por una implementación independiente (Python `hashlib` + `struct`): si la función no los da, la que está mal
// es la función. Cada hoja afirma su valor EXACTO (L3 del analyze), también las que comparan con otro espacio.

const (
	idLlamada1Copilot      = "u-000000000000000000000001"
	idLlamada2Copilot      = "u-000000000000000000000002"
	idCompactacion1Copilot = "r-000000000000000000000001"

	vectorLlamada1Copilot          = "3cf3d5fa7697681ce1ad1bd6e5abbd4e"
	vectorLlamada2Copilot          = "e63ae9472c63b251ce53d8f0246735e4"
	vectorCompactacion1Copilot     = "722a18b8a60a4ebc97f16925cbdbfe61"
	vectorLlamada1ComoCompactacion = "a76be27e9ed99dcd6427f38b47785dac"
	vectorLlamada1EnEspacioDeCodex = "aa69ad2965ddf47bfc427e5e1cc101aa"
)

// derivarCopilotOk exige que la derivación dé un valor.
func derivarCopilotOk(t *testing.T, tipo, id string) string {
	t.Helper()
	got, ok := derivarEventIDCopilot(tipo, id)
	if !ok {
		t.Fatalf("derivarEventIDCopilot(%q, %q) no dio valor", tipo, id)
	}
	return got
}

// (4) · Los cinco vectores normativos del contrato.
func TestEventIDCopilot_Vectores(t *testing.T) {
	t.Run("llamada", func(t *testing.T) {
		if got := derivarCopilotOk(t, tipoRespuesta, idLlamada1Copilot); got != vectorLlamada1Copilot {
			t.Errorf("= %s; se esperaba %s", got, vectorLlamada1Copilot)
		}
	})
	t.Run("otra_llamada", func(t *testing.T) {
		if got := derivarCopilotOk(t, tipoRespuesta, idLlamada2Copilot); got != vectorLlamada2Copilot {
			t.Errorf("= %s; se esperaba %s", got, vectorLlamada2Copilot)
		}
	})
	t.Run("compactacion", func(t *testing.T) {
		if got := derivarCopilotOk(t, tipoCompactacionCopilot, idCompactacion1Copilot); got != vectorCompactacion1Copilot {
			t.Errorf("= %s; se esperaba %s", got, vectorCompactacion1Copilot)
		}
	})
	t.Run("mismo_valor", func(t *testing.T) {
		// El valor de la 1.ª como compactación: el tipo separa un requestId de un usageId iguales por azar.
		got := derivarCopilotOk(t, tipoCompactacionCopilot, idLlamada1Copilot)
		if got != vectorLlamada1ComoCompactacion || got == vectorLlamada1Copilot {
			t.Errorf("= %s; se esperaba %s, distinto de %s", got, vectorLlamada1ComoCompactacion, vectorLlamada1Copilot)
		}
	})
	t.Run("codex", func(t *testing.T) {
		// El valor de la 1.ª en el espacio de Codex: el espacio de nombres separa las herramientas.
		codex, okCodex := derivarEventIDCodex(idLlamada1Copilot)
		got := derivarCopilotOk(t, tipoRespuesta, idLlamada1Copilot)
		if !okCodex || codex != vectorLlamada1EnEspacioDeCodex || got != vectorLlamada1Copilot || got == codex {
			t.Errorf("Codex = %s (%t), Copilot = %s; se esperaba Codex %s y Copilot %s", codex, okCodex, got,
				vectorLlamada1EnEspacioDeCodex, vectorLlamada1Copilot)
		}
	})
}

// (5) · Sin id no hay `event_id`. NACE VERDE con el stub (que ya da `"", false`): lo valida m9.
func TestEventIDCopilot_VacioNoDerivaNada(t *testing.T) {
	for _, tipo := range []string{tipoRespuesta, tipoCompactacionCopilot} {
		if got, ok := derivarEventIDCopilot(tipo, ""); ok || got != "" {
			t.Errorf("derivarEventIDCopilot(%q, \"\") = (%q, %t); se esperaba (\"\", false)", tipo, got, ok)
		}
	}
}
