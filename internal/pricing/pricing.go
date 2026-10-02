// Package pricing calcula el coste en LOCAL. El agente nunca depende del backend
// para conocer el coste: la tabla viaja empaquetada en el binario (Principio II).
package pricing

// Rate en USD por millón de tokens.
type Rate struct {
	Input      float64
	Output     float64
	CacheWrite float64
	CacheRead  float64
}

// Table es la tabla de tarifas empaquetada con el binario: ESPEJO EXACTO del catálogo de la
// plataforma en un commit concreto (P-006 FR-014; specs/006-medicion-fiel/contracts/tarifas.md).
// Mismas claves, ni una más ni una menos, y las mismas cuatro cifras por clave, en USD por millón de
// tokens. La vigila `pricing_test.go` contra una copia escrita aparte (P-006 FR-020).
//
// Cabecera (P-006 FR-015, FR-018, FR-019; SC-011):
//
//   - Fuente: https://platform.claude.com/docs/en/about-claude/pricing
//   - Verificación: 2026-08-07 (catálogo); filas claude-opus-5-5 y claude-sonnet-5, 2026-10-02.
//   - Aprobación: Basilio, 2026-08-07 (catálogo); filas claude-opus-5-5 y claude-sonnet-5, 2026-10-02.
//   - Catálogo replicado: permea-dev/permea-platform · backend/config/pricing.php · e50d0a5
//   - Casamiento: exacto con el identificador de modelo tal como llega en el evento; no se normalizan
//     sufijos de fecha, prefijos ni mayúsculas (P-006 FR-018). Un modelo sin fila sale con
//     cost_available=false y cost_usd=0 (P-006 FR-017).
//   - Limitación 1: la escritura de caché va a la tarifa de 5 MINUTOS; una escritura de 1 hora
//     quedaría infravalorada (el evento no la distingue).
//   - Limitación 2: el «modo rápido» no se distingue; un evento en modo rápido quedaría infravalorado.
//
// Un cambio del catálogo se absorbe en un solo commit: la fila aquí, la fila en la tabla esperada del
// test, el commit replicado de esta cabecera, la limitación afectada si la hay, y M4 en la spec.
var Table = map[string]Rate{
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

// Cost devuelve el coste de una llamada y un booleano de disponibilidad (R5): un
// modelo ausente de la tabla devuelve (0, false) —"no disponible", distinto de un
// coste 0 real—; los tokens se contabilizan aparte aunque el coste no esté disponible.
func Cost(model string, in, out, cacheCreate, cacheRead int) (float64, bool) {
	r, ok := Table[model]
	if !ok {
		return 0, false
	}
	perM := func(tokens int, rate float64) float64 { return float64(tokens) / 1_000_000 * rate }
	return perM(in, r.Input) + perM(out, r.Output) + perM(cacheCreate, r.CacheWrite) + perM(cacheRead, r.CacheRead), true
}
