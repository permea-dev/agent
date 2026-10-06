package pricing

import (
	"math"
	"sort"
	"testing"
)

// TestCost comprueba el coste local frente a un cálculo de referencia (SC-001, ±1%).
//
// P-006 FR-016: `claude-opus-4-6` pasa a 5 / 25 / 6,25 / 0,50 USD por millón, la tarifa del catálogo
// de la plataforma (`contracts/tarifas.md`). Antes este test fijaba la cifra errónea, 15 / 75.
func TestCost(t *testing.T) {
	// claude-opus-4-6: Input 5, Output 25, CacheWrite 6.25, CacheRead 0.5 (USD/millón).
	cost, ok := Cost("claude-opus-4-6", 1_000_000, 1_000_000, 1_000_000, 0, 1_000_000)
	if !ok {
		t.Fatalf("modelo conocido debe devolver ok=true")
	}
	want := 5.0 + 25.0 + 6.25 + 0.5
	if math.Abs(cost-want)/want > 0.01 {
		t.Errorf("coste fuera de ±1%%: got %v want %v", cost, want)
	}
}

// TestCost_UnknownModel: un modelo ausente de la tabla no bloquea; ok=false, coste 0.
func TestCost_UnknownModel(t *testing.T) {
	cost, ok := Cost("modelo-inexistente", 1000, 1000, 0, 0, 0)
	if ok {
		t.Errorf("modelo desconocido debe devolver ok=false")
	}
	if cost != 0 {
		t.Errorf("modelo desconocido debe devolver coste 0, got %v", cost)
	}
}

// ═══ P-006 B3 · EL ESPEJO DEL CATÁLOGO DE LA PLATAFORMA (FR-014, FR-020, SC-009) ═══════════
//
// La tabla empaquetada es una copia exacta del catálogo de la plataforma en un commit concreto
// (`specs/007-coste-fiel/contracts/tarifas.md`, que sustituye al de 006). Esta tabla esperada se escribe APARTE, copiada del
// catálogo y NO de `pricing.go`: dos copias que un test obliga a coincidir avisan de un cambio; una
// sola copia lo aplicaría en silencio. El test no lee el otro repositorio.
//
// PROCEDENCIA: permea-dev/permea-platform · backend/config/pricing.php · 8f147d1
// (17 claves y CINCO cifras: la escritura de caché a 5 minutos y a 1 hora; fila nueva `claude-fable-5-1`). Escrita
// a mano desde la tabla literal de `specs/007-coste-fiel/contracts/tarifas.md` (P-007 FR-008), no desde `pricing.go`.
var esperadaDelCatalogo = map[string]Rate{
	"claude-fable-5":    {Input: 10.00, Output: 50.00, CacheWrite: 12.50, CacheWrite1h: 20.00, CacheRead: 1.00},
	"claude-fable-5-1":  {Input: 10.00, Output: 50.00, CacheWrite: 12.50, CacheWrite1h: 20.00, CacheRead: 0.25},
	"claude-mythos-5":   {Input: 10.00, Output: 50.00, CacheWrite: 12.50, CacheWrite1h: 20.00, CacheRead: 1.00},
	"claude-opus-5-5":   {Input: 4.00, Output: 20.00, CacheWrite: 5.00, CacheWrite1h: 8.00, CacheRead: 0.20},
	"claude-opus-5":     {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheWrite1h: 10.00, CacheRead: 0.50},
	"claude-opus-4-8":   {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheWrite1h: 10.00, CacheRead: 0.50},
	"claude-opus-4-7":   {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheWrite1h: 10.00, CacheRead: 0.50},
	"claude-opus-4-6":   {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheWrite1h: 10.00, CacheRead: 0.50},
	"claude-opus-4-5":   {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheWrite1h: 10.00, CacheRead: 0.50},
	"claude-opus-4-1":   {Input: 15.00, Output: 75.00, CacheWrite: 18.75, CacheWrite1h: 30.00, CacheRead: 1.50},
	"claude-opus-4":     {Input: 15.00, Output: 75.00, CacheWrite: 18.75, CacheWrite1h: 30.00, CacheRead: 1.50},
	"claude-sonnet-5":   {Input: 2.00, Output: 10.00, CacheWrite: 2.50, CacheWrite1h: 4.00, CacheRead: 0.20},
	"claude-sonnet-4-6": {Input: 3.00, Output: 15.00, CacheWrite: 3.75, CacheWrite1h: 6.00, CacheRead: 0.30},
	"claude-sonnet-4-5": {Input: 3.00, Output: 15.00, CacheWrite: 3.75, CacheWrite1h: 6.00, CacheRead: 0.30},
	"claude-sonnet-4":   {Input: 3.00, Output: 15.00, CacheWrite: 3.75, CacheWrite1h: 6.00, CacheRead: 0.30},
	"claude-haiku-4-5":  {Input: 1.00, Output: 5.00, CacheWrite: 1.25, CacheWrite1h: 2.00, CacheRead: 0.10},
	"claude-haiku-3-5":  {Input: 0.80, Output: 4.00, CacheWrite: 1.00, CacheWrite1h: 1.60, CacheRead: 0.08},
}

// (15) · P-006 FR-020 — la tabla tiene exactamente las claves del catálogo replicado.
//
// (1) · P-007 FR-006, FR-008 — 17 desde `8f147d1`.
func TestEspejo_RecuentoDeClaves(t *testing.T) {
	if got, want := len(Table), len(esperadaDelCatalogo); got != want {
		t.Errorf("la tabla tiene %d claves; el catálogo replicado tiene %d", got, want)
	}
}

// (16) · P-006 FR-014, FR-020 — cada clave del catálogo existe en la tabla con sus cifras exactas. Un
// subtest por clave, para que el fallo diga cuál.
//
// (2) · P-007 FR-006, FR-008, SC-010 — las cifras son CINCO: también la escritura de caché a 1 hora.
func TestEspejo_CifrasClaveAClave(t *testing.T) {
	claves := make([]string, 0, len(esperadaDelCatalogo))
	for k := range esperadaDelCatalogo {
		claves = append(claves, k)
	}
	sort.Strings(claves)
	for _, clave := range claves {
		quiere := esperadaDelCatalogo[clave]
		t.Run(clave, func(t *testing.T) {
			got, ok := Table[clave]
			if !ok {
				t.Errorf("falta la clave %q", clave)
				return
			}
			if got.Input != quiere.Input {
				t.Errorf("entrada = %v, want %v", got.Input, quiere.Input)
			}
			if got.Output != quiere.Output {
				t.Errorf("salida = %v, want %v", got.Output, quiere.Output)
			}
			if got.CacheWrite != quiere.CacheWrite {
				t.Errorf("escritura de caché = %v, want %v", got.CacheWrite, quiere.CacheWrite)
			}
			if got.CacheWrite1h != quiere.CacheWrite1h {
				t.Errorf("escritura de caché a 1 hora = %v, want %v", got.CacheWrite1h, quiere.CacheWrite1h)
			}
			if got.CacheRead != quiere.CacheRead {
				t.Errorf("lectura de caché = %v, want %v", got.CacheRead, quiere.CacheRead)
			}
		})
	}
}

// (17) · P-006 FR-020 — ninguna clave de la tabla sobra respecto al catálogo replicado.
func TestEspejo_NingunaClaveSobra(t *testing.T) {
	for clave := range Table {
		if _, ok := esperadaDelCatalogo[clave]; !ok {
			t.Errorf("la clave %q está en la tabla y no en el catálogo replicado", clave)
		}
	}
}

// (19) · P-006 FR-014, SC-010 — un evento de `claude-opus-5-5` con todas sus partidas, contra el cálculo
// hecho a mano (por el orquestador, no con las constantes de la tabla).
//
// Es el ÚNICO test que distingue las partidas: `TestCost` usa los mismos tokens en todas, así que no ve
// dos tarifas cruzadas. Por eso los tokens son irregulares y la comparación es ABSOLUTA y estricta
// (≤ 1e-9), no el ±1 % de `TestCost`, que viene de SC-001 de la feature 001: con ±1 % no se ve una tarifa
// desviada en un céntimo ni un coste redondeado a céntimos.
//
// (3) · P-007 FR-002, SC-004 — la escritura de caché, por duración: la de 5 minutos a `CacheWrite` y la de
// 1 hora a `CacheWrite1h`. Dos vectores: con desglose, y con toda la escritura a 1 hora, que es lo que
// recibe `Cost` cuando la línea no trae desglose (P-1). Si se cruzan las dos tarifas, caen los dos.
func TestCost_Opus55AMano(t *testing.T) {
	casos := []struct {
		nombre string
		cw5m   int
		cw1h   int
		want   float64
		aMano  string
	}{
		// A 4 / 20 / 5 / 8 / 0,20 USD por millón.
		{"con_desglose", 12_345, 33_334, 1.1775756,
			"0,493828 + 0,157820 + 0,061725 + 0,266672 + 0,1975306 = 1,1775756"},
		{"sin_desglose_todo_a_1_hora", 0, 45_679, 1.2146106,
			"0,493828 + 0,157820 + 0,365432 + 0,1975306 = 1,2146106"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cost, ok := Cost("claude-opus-5-5", 123_457, 7_891, c.cw5m, c.cw1h, 987_653)
			if !ok {
				t.Fatalf("claude-opus-5-5 debe tener fila (ok=true)")
			}
			if math.Abs(cost-c.want) > 1e-9 {
				t.Errorf("coste = %.10f, want %.10f (%s; diferencia absoluta %.3g > 1e-9)", cost, c.want, c.aMano, math.Abs(cost-c.want))
			}
		})
	}
}
