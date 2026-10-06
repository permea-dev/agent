package ingest

import (
	"fmt"
	"math"
	"testing"

	"github.com/permea-dev/agent/internal/event"
)

// ═══ P-007 B3 · UN MENSAJE ESCRITO EN VARIAS LÍNEAS SE CUENTA ENTERO (FR-009, SC-002, SC-003) ════════
//
// Claude Code puede escribir un mensaje en varias líneas con la salida creciendo (7 → 1 303 → 89 817 en
// el Hallazgo de W2). Dentro de una pasada, cada partida vale el MÁXIMO entre sus líneas, nunca la suma.
// El desglose de la escritura de caché es el de la línea que da el máximo de la escritura, la última si
// empatan (P-2). Identificadores sintéticos (disciplina 9).

// lineaDeMensaje es una línea `claude-opus-5-5` del mensaje `sufijo`, con salida, escritura de caché y su
// desglose dados. Entrada y lectura a 0, para que el coste dependa sólo de lo que mira cada test.
func lineaDeMensaje(sufijo string, salida, cw5m, cw1h int) []byte {
	return []byte(fmt.Sprintf(`{"type":"assistant","timestamp":"2026-10-06T12:00:00Z","sessionId":"s","cwd":"/tmp/x",`+
		`"requestId":"req_MAXIMO%021s","message":{"id":"msg_MAXIMO%021s","model":"claude-opus-5-5","usage":{`+
		`"input_tokens":0,"output_tokens":%d,"cache_creation_input_tokens":%d,"cache_read_input_tokens":0,`+
		`"cache_creation":{"ephemeral_5m_input_tokens":%d,"ephemeral_1h_input_tokens":%d}}}}`,
		sufijo, sufijo, salida, cw5m+cw1h, cw5m, cw1h))
}

// salidaDe suma la salida de los eventos.
func salidaDe(evs []*event.Event) int {
	n := 0
	for _, ev := range evs {
		n += ev.TokensOutput
	}
	return n
}

// unEvento exige que las líneas de UN mensaje produzcan un solo evento, y lo devuelve.
func unEvento(t *testing.T, evs []*event.Event) *event.Event {
	t.Helper()
	if len(evs) != 1 {
		t.Fatalf("precondición: un mensaje debe producir un evento; produjo %d", len(evs))
	}
	return evs[0]
}

// (9) · P-007 FR-009, SC-002 — la salida crece 7 → 1 303 → 89 817: UN evento con 89 817.
func TestMaximo_LaSalidaQueCreceValeSuMaximo(t *testing.T) {
	ev := unEvento(t, leerEnUnaPasada(t, NuevaPasada(),
		lineaDeMensaje("1", 7, 0, 0), lineaDeMensaje("1", 1303, 0, 0), lineaDeMensaje("1", 89817, 0, 0)))
	if ev.TokensOutput != 89817 {
		t.Errorf("salida = %d; se esperaba 89 817, el máximo (la primera daría 7)", ev.TokensOutput)
	}
}

// (10) · P-007 FR-009, SC-002 — el máximo EN MEDIO: 7 → 89 817 → 1 303. Vale 89 817, no la última.
func TestMaximo_ElMaximoEnMedioNoLaUltima(t *testing.T) {
	ev := unEvento(t, leerEnUnaPasada(t, NuevaPasada(),
		lineaDeMensaje("2", 7, 0, 0), lineaDeMensaje("2", 89817, 0, 0), lineaDeMensaje("2", 1303, 0, 0)))
	if ev.TokensOutput != 89817 {
		t.Errorf("salida = %d; se esperaba 89 817, el máximo (la última daría 1 303)", ev.TokensOutput)
	}
}

// desgloseCerrado lee las líneas de UN mensaje en una pasada y devuelve el evento y el desglose con que
// se cerró. Si la pasada no cerró nada, el desglose no se puede observar: falla.
func desgloseCerrado(t *testing.T, lineas ...[]byte) (coste float64, cw5m, cw1h int, cerrado bool) {
	t.Helper()
	p := NuevaPasada()
	ctx := Context{Salt: "s", Pasada: p}
	var emitidos []*event.Event
	for _, l := range lineas {
		ev, err := FromClaudeCodeLine(l, ctx)
		if err != nil {
			t.Fatalf("precondición: línea corrupta: %v", err)
		}
		if ev != nil {
			emitidos = append(emitidos, ev)
		}
	}
	cerrados := p.CerrarFichero()
	for _, c := range cerrados {
		ev := c.Evento
		emitidos = append(emitidos, &ev)
	}
	ev := unEvento(t, emitidos)
	if len(cerrados) == 1 {
		return ev.CostUSD, cerrados[0].Escritura5m, cerrados[0].Escritura1h, true
	}
	return ev.CostUSD, 0, 0, false
}

// exigirDesglose comprueba el coste (sólo escritura de caché, a mano) y el desglose del cerrado.
func exigirDesglose(t *testing.T, coste float64, cw5m, cw1h int, cerrado bool, wantCoste float64, want5m, want1h int) {
	t.Helper()
	t.Run("coste", func(t *testing.T) {
		if math.Abs(coste-wantCoste) > 1e-12 {
			t.Errorf("coste = %.10f, want %.10f", coste, wantCoste)
		}
	})
	t.Run("desglose", func(t *testing.T) {
		if !cerrado {
			t.Fatalf("la pasada no cerró el mensaje: el desglose no se puede observar")
		}
		if cw5m != want5m || cw1h != want1h {
			t.Errorf("desglose = %d / %d; se esperaba %d / %d", cw5m, cw1h, want5m, want1h)
		}
	})
}

// (11) · P-007 FR-009, P-2 (E-4) — el desglose es el de la línea que da el máximo de la escritura, que aquí
// NO es la última: 100 (todo 5 m) → 300 (todo 1 h) → 200 (todo 5 m). Vale 0 / 300.
// Coste a mano: 300 × 8 / 10⁶ = 0,0024 USD (la de la última, 200 × 5 / 10⁶, daría 0,001).
func TestMaximo_ElDesgloseEsElDeLaLineaDelMaximo(t *testing.T) {
	coste, c5, c1, cerrado := desgloseCerrado(t,
		lineaDeMensaje("3", 0, 100, 0), lineaDeMensaje("3", 0, 0, 300), lineaDeMensaje("3", 0, 200, 0))
	exigirDesglose(t, coste, c5, c1, cerrado, 0.0024, 0, 300)
}

// (25) · P-007 FR-009, P-2 (E-4) — EMPATE: dos líneas con la misma escritura máxima y distinto reparto.
// Vale la ÚLTIMA: 300 (todo 5 m) → 300 (todo 1 h) da 0 / 300, y 0,0024 USD (la primera daría 0,0015).
func TestMaximo_EnEmpateValeLaUltima(t *testing.T) {
	coste, c5, c1, cerrado := desgloseCerrado(t, lineaDeMensaje("4", 0, 300, 0), lineaDeMensaje("4", 0, 0, 300))
	exigirDesglose(t, coste, c5, c1, cerrado, 0.0024, 0, 300)
}

// (12) · P-007 FR-017 — la pasada cuenta los mensajes que CRECIERON entre líneas, y sólo ésos.
func TestMaximo_LaPasadaCuentaLosQueCrecieron(t *testing.T) {
	p := NuevaPasada()
	_ = leerEnUnaPasada(t, p,
		lineaDeMensaje("5", 7, 0, 0), lineaDeMensaje("5", 1303, 0, 0), lineaDeMensaje("5", 89817, 0, 0),
		lineaDeMensaje("6", 40, 0, 0), lineaDeMensaje("6", 40, 0, 0))
	if got := p.Recuentos().Crecieron; got != 1 {
		t.Errorf("Crecieron = %d; se esperaba 1 (uno crece y otro repite)", got)
	}
}

// (14) · P-007 SC-003 — duplicar una línea no cambia ni los eventos ni las sumas: el máximo es idempotente.
// NACE VERDE (con «la primera manda» tampoco cambiaba); la valida la mutación que SUMA (m11).
func TestMaximo_UnaLineaDuplicadaNoCambiaNada(t *testing.T) {
	a, b, c := lineaDeMensaje("7", 7, 0, 0), lineaDeMensaje("7", 1303, 0, 0), lineaDeMensaje("7", 89817, 0, 0)
	sin := leerEnUnaPasada(t, NuevaPasada(), a, b, c)
	con := leerEnUnaPasada(t, NuevaPasada(), a, b, b, c)
	if len(con) != len(sin) || salidaDe(con) != salidaDe(sin) {
		t.Errorf("con una línea duplicada: %d eventos y %d de salida; sin ella: %d y %d",
			len(con), salidaDe(con), len(sin), salidaDe(sin))
	}
}
