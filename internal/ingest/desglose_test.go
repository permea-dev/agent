package ingest

import (
	"fmt"
	"math"
	"testing"
)

// ═══ P-007 B2 · LA ESCRITURA DE CACHÉ, POR DURACIÓN (FR-001 a FR-005, SC-004) ═══════════════════════
//
// El log trae la escritura de caché partida en `usage.cache_creation.ephemeral_5m_input_tokens` y
// `…ephemeral_1h_input_tokens`. El coste la tarifa por duración; el evento sigue llevando el TOTAL
// (`cache_creation_input_tokens`), y el desglose no cruza la frontera (D-1).
//
// Vector de SC-004, calculado a mano por el orquestador: `claude-opus-5-5` a 4 / 20 / 5 / 8 / 0,20 USD
// por millón, con 123 457 de entrada, 7 891 de salida, 45 679 de escritura y 987 653 de lectura.
// Identificadores sintéticos (disciplina 9).

const (
	costeConDesglose = 1.1775756 // 0,493828 + 0,157820 + 0,061725 + 0,266672 + 0,1975306
	costeTodoA1Hora  = 1.2146106 // 0,493828 + 0,157820 + 0,365432 + 0,1975306
)

// lineaConEscritura es una línea `claude-opus-5-5` del vector de SC-004. `cacheCreation` es el JSON de
// `usage.cache_creation`, o "" para una línea sin desglose.
func lineaConEscritura(sufijo, cacheCreation string) []byte {
	desglose := ""
	if cacheCreation != "" {
		desglose = `,"cache_creation":` + cacheCreation
	}
	return []byte(fmt.Sprintf(`{"type":"assistant","timestamp":"2026-10-06T12:00:00Z","sessionId":"s","cwd":"/tmp/x",`+
		`"requestId":"req_DESGLOSE%019s","message":{"id":"msg_DESGLOSE%019s","model":"claude-opus-5-5","usage":{`+
		`"input_tokens":123457,"output_tokens":7891,"cache_creation_input_tokens":45679,"cache_read_input_tokens":987653%s}}}`,
		sufijo, sufijo, desglose))
}

// eventoDe convierte UNA línea, sin pasada, y exige que salga evento.
func eventoDe(t *testing.T, linea []byte, ctx Context) (coste float64, escritura int) {
	t.Helper()
	ev, err := FromClaudeCodeLine(linea, ctx)
	if err != nil || ev == nil {
		t.Fatalf("precondición: la línea debe producir un evento (ev=%v, err=%v)", ev, err)
	}
	if !ev.CostAvailable {
		t.Fatalf("precondición: claude-opus-5-5 debe tener tarifa")
	}
	return ev.CostUSD, ev.TokensCacheCreation
}

// exigirCoste compara en ABSOLUTO, ≤ 1e-9, como `TestCost_Opus55AMano`.
func exigirCoste(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("coste = %.10f, want %.10f (diferencia absoluta %.3g > 1e-9)", got, want, math.Abs(got-want))
	}
}

// (4) · P-007 FR-001, FR-002, FR-005 — una línea con desglose se tarifa por duración, y el evento lleva
// el TOTAL de la escritura.
func TestDesglose_LineaConDesgloseSeTarifaPorDuracion(t *testing.T) {
	coste, escritura := eventoDe(t, lineaConEscritura("1",
		`{"ephemeral_5m_input_tokens":12345,"ephemeral_1h_input_tokens":33334}`), Context{Salt: "s"})

	t.Run("coste", func(t *testing.T) { exigirCoste(t, coste, costeConDesglose) })
	t.Run("tokens_cache_creation_es_el_total", func(t *testing.T) {
		if escritura != 45679 {
			t.Errorf("tokens_cache_creation = %d, want 45679 (el total del log, D-1)", escritura)
		}
	})
}

// (5) · P-007 FR-003, P-1 — una línea SIN desglose tarifa toda su escritura a 1 hora.
func TestDesglose_SinDesgloseTodoA1Hora(t *testing.T) {
	coste, escritura := eventoDe(t, lineaConEscritura("2", ""), Context{Salt: "s"})

	t.Run("coste", func(t *testing.T) { exigirCoste(t, coste, costeTodoA1Hora) })
	t.Run("tokens_cache_creation_es_el_total", func(t *testing.T) {
		if escritura != 45679 {
			t.Errorf("tokens_cache_creation = %d, want 45679", escritura)
		}
	})
}

// (6) · P-007 FR-004, Q-4 — un desglose que NO suma el total se trata como si no lo hubiera: todo a 1 hora,
// y el evento lleva el total del log, no la suma del desglose.
func TestDesglose_DesgloseQueNoSumaEsSinDesglose(t *testing.T) {
	coste, escritura := eventoDe(t, lineaConEscritura("3",
		`{"ephemeral_5m_input_tokens":10000,"ephemeral_1h_input_tokens":10000}`), Context{Salt: "s"})

	t.Run("coste", func(t *testing.T) { exigirCoste(t, coste, costeTodoA1Hora) })
	t.Run("tokens_cache_creation_es_el_total_del_log", func(t *testing.T) {
		if escritura != 45679 {
			t.Errorf("tokens_cache_creation = %d, want 45679 (el total del log, no 5 m + 1 h)", escritura)
		}
	})
}

// (7) · P-007 FR-003 — la pasada cuenta las líneas sin desglose y las que traen uno que no suma; la que
// lo trae bien no cuenta.
func TestDesglose_LaPasadaCuentaLasLineasSinDesglose(t *testing.T) {
	p := NuevaPasada()
	_ = leerEnUnaPasada(t, p,
		lineaConEscritura("4", `{"ephemeral_5m_input_tokens":12345,"ephemeral_1h_input_tokens":33334}`),
		lineaConEscritura("5", ""),
		lineaConEscritura("6", `{"ephemeral_5m_input_tokens":10000,"ephemeral_1h_input_tokens":10000}`),
	)
	if got := p.Recuentos().SinDesglose; got != 2 {
		t.Errorf("SinDesglose = %d; se esperaba 2 (una sin desglose y una con un desglose que no suma)", got)
	}
}
