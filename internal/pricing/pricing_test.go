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
	cost, ok := Cost("claude-opus-4-6", 1_000_000, 1_000_000, 1_000_000, 1_000_000)
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
	cost, ok := Cost("modelo-inexistente", 1000, 1000, 0, 0)
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
// (`specs/006-medicion-fiel/contracts/tarifas.md`). Esta tabla esperada se escribe APARTE, copiada del
// catálogo y NO de `pricing.go`: dos copias que un test obliga a coincidir avisan de un cambio; una
// sola copia lo aplicaría en silencio. El test no lee el otro repositorio.
//
// PROCEDENCIA: permea-dev/permea-platform · backend/config/pricing.php · e50d0a5
// (16 claves; `claude-sonnet-5` corregida a 2.00 / 10.00 / 2.50 / 0.20 el 2026-10-02, Q-006-1).
var esperadaDelCatalogo = map[string]Rate{
	"claude-fable-5":    {Input: 10.00, Output: 50.00, CacheWrite: 12.50, CacheRead: 1.00},
	"claude-mythos-5":   {Input: 10.00, Output: 50.00, CacheWrite: 12.50, CacheRead: 1.00},
	"claude-opus-5-5":   {Input: 4.00, Output: 20.00, CacheWrite: 5.00, CacheRead: 0.20},
	"claude-opus-5":     {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheRead: 0.50},
	"claude-opus-4-8":   {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheRead: 0.50},
	"claude-opus-4-7":   {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheRead: 0.50},
	"claude-opus-4-6":   {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheRead: 0.50},
	"claude-opus-4-5":   {Input: 5.00, Output: 25.00, CacheWrite: 6.25, CacheRead: 0.50},
	"claude-opus-4-1":   {Input: 15.00, Output: 75.00, CacheWrite: 18.75, CacheRead: 1.50},
	"claude-opus-4":     {Input: 15.00, Output: 75.00, CacheWrite: 18.75, CacheRead: 1.50},
	"claude-sonnet-5":   {Input: 2.00, Output: 10.00, CacheWrite: 2.50, CacheRead: 0.20},
	"claude-sonnet-4-6": {Input: 3.00, Output: 15.00, CacheWrite: 3.75, CacheRead: 0.30},
	"claude-sonnet-4-5": {Input: 3.00, Output: 15.00, CacheWrite: 3.75, CacheRead: 0.30},
	"claude-sonnet-4":   {Input: 3.00, Output: 15.00, CacheWrite: 3.75, CacheRead: 0.30},
	"claude-haiku-4-5":  {Input: 1.00, Output: 5.00, CacheWrite: 1.25, CacheRead: 0.10},
	"claude-haiku-3-5":  {Input: 0.80, Output: 4.00, CacheWrite: 1.00, CacheRead: 0.08},
}

// (15) · P-006 FR-020 — la tabla tiene exactamente las claves del catálogo replicado.
func TestEspejo_RecuentoDeClaves(t *testing.T) {
	if got, want := len(Table), len(esperadaDelCatalogo); got != want {
		t.Errorf("la tabla tiene %d claves; el catálogo replicado tiene %d", got, want)
	}
}

// (16) · P-006 FR-014, FR-020 — cada clave del catálogo existe en la tabla con sus CUATRO cifras
// exactas. Un subtest por clave, para que el fallo diga cuál.
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

// (19) · P-006 FR-014, SC-010 — un evento de `claude-opus-5-5` con las CUATRO partidas, contra el
// cálculo hecho a mano (por el orquestador, no con las constantes de la tabla).
//
// Es el ÚNICO test que distingue las cuatro partidas: `TestCost` usa los mismos tokens en las cuatro,
// así que no ve dos tarifas cruzadas. Por eso los tokens son irregulares y la comparación es ABSOLUTA
// y estricta (≤ 1e-9), no el ±1 % de `TestCost`, que viene de SC-001 de la feature 001: con ±1 % no se
// ve una tarifa desviada en un céntimo ni un coste redondeado a céntimos.
func TestCost_Opus55AMano(t *testing.T) {
	cost, ok := Cost("claude-opus-5-5", 123_457, 7_891, 45_679, 987_653)
	if !ok {
		t.Fatalf("claude-opus-5-5 debe tener fila (ok=true)")
	}
	// A 4 / 20 / 5 / 0,20 USD por millón:
	// 0,493828 + 0,157820 + 0,228395 + 0,1975306 = 1,0775736 USD.
	want := 1.0775736
	if math.Abs(cost-want) > 1e-9 {
		t.Errorf("coste = %.10f, want %.10f (diferencia absoluta %.3g > 1e-9)", cost, want, math.Abs(cost-want))
	}
}
