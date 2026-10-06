package ingest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/event"
)

// ═══ P-006 B2b · LA PASADA — UN MENSAJE, UN EVENTO ═══════════════════════════════════════
//
// Claude Code escribe VARIAS líneas por mensaje, todas con el mismo consumo. Dentro de una pasada,
// la primera línea de un mensaje produce el evento y las demás no: se cuentan, y NUNCA se suman
// (`specs/006-medicion-fiel/contracts/event-id.md`, §Una pasada, un evento por mensaje).
//
// Todos los identificadores son SINTÉTICOS (disciplina 9 de P-006).

// lineaConConsumo arma una línea facturable con identificadores y tokens de entrada/salida dados.
func lineaConConsumo(messageID, requestID string, entrada, salida int) []byte {
	return []byte(fmt.Sprintf(`{"type":"assistant","timestamp":"2026-10-02T12:00:00Z","sessionId":"s","cwd":"/x","requestId":%q,"message":{"id":%q,"model":"claude-opus-4-6","usage":{"input_tokens":%d,"output_tokens":%d}}}`,
		requestID, messageID, entrada, salida))
}

// leerEnUnaPasada pasa las líneas por FromClaudeCodeLine con UNA pasada compartida, como un fichero, y
// devuelve los eventos emitidos: los que salgan por línea y los que la pasada cierre al acabar el fichero
// (P-007 FR-009). Una línea corrupta es precondición rota (t.Fatalf).
func leerEnUnaPasada(t *testing.T, p *Pasada, lineas ...[]byte) []*event.Event {
	t.Helper()
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
	for _, c := range p.CerrarFichero() {
		ev := c.Evento
		emitidos = append(emitidos, &ev)
	}
	return emitidos
}

// tokensDe suma entrada y salida de todos los eventos emitidos.
func tokensDe(evs []*event.Event) int {
	n := 0
	for _, ev := range evs {
		n += ev.TokensInput + ev.TokensOutput
	}
	return n
}

// (7) · P-006 FR-001, FR-033 — tres líneas del mismo mensaje, en una pasada, son UN evento con los
// tokens de UNA línea, y la pasada lo cuenta.
func TestPasada_UnMensajeDeTresLineasEsUnEvento(t *testing.T) {
	const m, r = "msg_PASADA000000000000000001", "req_PASADA000000000000000001"
	linea := lineaConConsumo(m, r, 100, 40)
	p := NuevaPasada()
	evs := leerEnUnaPasada(t, p, linea, linea, linea)

	t.Run("un_evento", func(t *testing.T) {
		if len(evs) != 1 {
			t.Errorf("P-006 FR-001: tres líneas del mismo mensaje produjeron %d eventos; se esperaba 1", len(evs))
		}
	})
	t.Run("tokens_de_una_linea", func(t *testing.T) {
		if got := tokensDe(evs); got != 140 {
			t.Errorf("P-006 FR-001: los eventos emitidos suman %d tokens; un mensaje de 140 se contó de más", got)
		}
	})
	t.Run("recuentos", func(t *testing.T) {
		rc := p.Recuentos()
		if rc.Facturables != 3 || rc.Emitidos != 1 || rc.Repetidas != 2 {
			t.Errorf("recuentos = %+v; se esperaba Facturables=3, Emitidos=1, Repetidas=2", rc)
		}
	})
}

// (8) · P-006 FR-005, SC-008 (a) — la segunda línea del mismo mensaje trae OTRO consumo: un solo evento,
// nunca se suman, y la discrepancia se cuenta.
//
// P-007 FR-009 sustituye «la primera manda» por EL MÁXIMO POR PARTIDA: 100/40 y 999/1 dan 999/40.
func TestCasoLimite_ConsumoDistinto(t *testing.T) {
	const m, r = "msg_DISCREPA00000000000000001", "req_DISCREPA00000000000000001"
	p := NuevaPasada()
	evs := leerEnUnaPasada(t, p, lineaConConsumo(m, r, 100, 40), lineaConConsumo(m, r, 999, 1))

	t.Run("un_solo_evento", func(t *testing.T) {
		if len(evs) != 1 {
			t.Errorf("P-006 FR-005: dos líneas del mismo mensaje produjeron %d eventos; se esperaba 1", len(evs))
		}
	})
	t.Run("cada_partida_vale_su_maximo", func(t *testing.T) {
		if got := tokensDe(evs); got != 1039 {
			t.Errorf("P-007 FR-009: los eventos emitidos suman %d tokens; se esperaban 999 + 40 = 1039, el máximo de cada partida, nunca una suma", got)
		}
	})
	t.Run("cuenta_la_discrepancia", func(t *testing.T) {
		if got := p.Recuentos().ConsumoDistinto; got != 1 {
			t.Errorf("P-006 FR-005: ConsumoDistinto = %d; se esperaba 1 (la discrepancia DEBE quedar visible)", got)
		}
	})
}

// (9) · P-006 FR-006, FR-007, SC-008 (c) — una línea sin identificadores no se emite y SE CUENTA; una
// `<synthetic>` tampoco se emite y se cuenta APARTE.
func TestCasoLimite_SinIdentificador(t *testing.T) {
	sinIDs := []byte(`{"type":"assistant","timestamp":"2026-10-02T12:00:00Z","sessionId":"s","cwd":"/x","message":{"model":"claude-opus-4-6","usage":{"input_tokens":10,"output_tokens":5}}}`)
	sintetica := []byte(`{"type":"assistant","timestamp":"2026-10-02T12:00:00Z","sessionId":"s","cwd":"/x","requestId":"req_SINTETIC000000000000000001","message":{"id":"msg_SINTETIC000000000000000001","model":"<synthetic>","usage":{"input_tokens":0,"output_tokens":0}}}`)
	p := NuevaPasada()
	_ = leerEnUnaPasada(t, p, sinIDs, sintetica)

	t.Run("cuenta_sin_identificador", func(t *testing.T) {
		if got := p.Recuentos().SinIdentificador; got != 1 {
			t.Errorf("P-006 FR-006: SinIdentificador = %d; se esperaba 1 (la pérdida DEBE quedar a la vista)", got)
		}
	})
	t.Run("sinteticas_aparte", func(t *testing.T) {
		if got := p.Recuentos().Sinteticas; got != 1 {
			t.Errorf("P-006 FR-007: Sinteticas = %d; se esperaba 1", got)
		}
	})
}

// (10) · P-006 R4 — el resumen de la pasada dice CUÁNTO, nunca QUIÉN: no está vacío, trae el número de
// facturables, y no contiene ni los identificadores de entrada ni ningún `event_id` emitido.
func TestPasada_ElResumenNoLlevaIdentificadores(t *testing.T) {
	ids := [][2]string{
		{"msg_RESUMEN0000000000000000001", "req_RESUMEN0000000000000000001"},
		{"msg_RESUMEN0000000000000000002", "req_RESUMEN0000000000000000002"},
	}
	// 7 facturables: el primer mensaje en 4 líneas, el segundo en 3.
	var lineas [][]byte
	for i := 0; i < 4; i++ {
		lineas = append(lineas, lineaConConsumo(ids[0][0], ids[0][1], 10, 5))
	}
	for i := 0; i < 3; i++ {
		lineas = append(lineas, lineaConConsumo(ids[1][0], ids[1][1], 20, 5))
	}
	p := NuevaPasada()
	evs := leerEnUnaPasada(t, p, lineas...)
	resumen := p.Resumen()

	t.Run("no_vacio_con_facturables", func(t *testing.T) {
		if strings.TrimSpace(resumen) == "" || !strings.Contains(resumen, "7") {
			t.Errorf("el resumen debe decir cuántas líneas facturables se leyeron (7); resumen = %q", resumen)
		}
	})
	t.Run("sin_identificadores_de_entrada", func(t *testing.T) {
		for _, par := range ids {
			for _, id := range par {
				if strings.Contains(resumen, id) || strings.Contains(resumen, "RESUMEN") {
					t.Errorf("el resumen contiene un identificador de entrada (%s): %q", id, resumen)
				}
			}
		}
	})
	// (18) · P-007 FR-017, P-5 — DOS líneas: la primera, la de 006 sin cambiar un byte; la segunda, la aprobada.
	t.Run("dos_lineas_aprobadas", func(t *testing.T) {
		want := "pasada: 7 líneas facturables · 2 eventos · 5 repetidas del mismo mensaje · 0 sintéticas · " +
			"0 sin identificador (no contables) · 0 con consumo distinto de la primera\n" +
			"pasada: 0 mensajes que crecieron entre líneas · 0 en espera de cerrarse · " +
			"0 líneas releídas de un mensaje en espera · 0 líneas tardías · 0 líneas sin desglose de caché (a 1 hora)"
		if resumen != want {
			t.Errorf("resumen =\n%s\nse esperaba, byte a byte,\n%s", resumen, want)
		}
	})
	t.Run("sin_event_id", func(t *testing.T) {
		for _, ev := range evs {
			if strings.Contains(resumen, ev.EventID) {
				t.Errorf("el resumen contiene un event_id emitido (%s): %q", ev.EventID, resumen)
			}
		}
	})
}
