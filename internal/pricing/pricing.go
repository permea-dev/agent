// Package pricing calcula el coste en LOCAL. El agente nunca depende del backend
// para conocer el coste: la tabla viaja empaquetada en el binario (Principio II).
package pricing

// Rate en USD por millón de tokens. La escritura de caché tiene DOS tarifas, una por duración: a 5 minutos
// (`CacheWrite`) y a 1 hora (`CacheWrite1h`), como el catálogo de la plataforma (P-007 FR-002, FR-006).
type Rate struct {
	Input        float64
	Output       float64
	CacheWrite   float64
	CacheWrite1h float64
	CacheRead    float64
}

// Table es la tabla de tarifas empaquetada con el binario: ESPEJO EXACTO del catálogo de la
// plataforma en un commit concreto (P-007 FR-006; specs/007-coste-fiel/contracts/tarifas.md, que sustituye
// al de 006). Mismas claves, ni una más ni una menos, y las mismas CINCO cifras por clave, en USD por millón
// de tokens. La vigila `pricing_test.go` contra una copia escrita aparte (P-007 FR-008).
//
// Cabecera (P-006 FR-015, FR-018, FR-019; P-007 FR-007):
//
//   - Fuente: https://platform.claude.com/docs/en/about-claude/pricing
//   - Verificación: 2026-10-04 (las 17 filas, con sus cinco cifras, contra la fuente).
//   - Aprobación: Basilio, 2026-08-07 (catálogo); la quinta cifra y la fila claude-fable-5-1, 2026-10-04 (P-031).
//   - Catálogo replicado: permea-dev/permea-platform · backend/config/pricing.php · 8f147d1
//   - Casamiento: exacto con el identificador de modelo tal como llega en el evento; no se normalizan
//     sufijos de fecha, prefijos ni mayúsculas (P-006 FR-018). Un modelo sin fila sale con
//     cost_available=false y cost_usd=0 (P-006 FR-017).
//   - Hipótesis P-1 (sustituye a la «Limitación 1»): la escritura de caché se tarifa por su duración, la
//     de 5 minutos a `CacheWrite` y la de 1 hora a `CacheWrite1h`. Una línea SIN desglose, o con uno que no
//     suma el total, tarifa TODA su escritura a 1 hora (P-007 FR-003, FR-004). Es la misma hipótesis que
//     declara la cabecera del catálogo.
//   - Limitación 2: el «modo rápido» no se distingue; un evento en modo rápido quedaría infravalorado.
//
// Un cambio del catálogo se absorbe en un solo commit, en cinco sitios: la fila aquí, la fila en la tabla
// esperada del test, el commit replicado de esta cabecera, la hipótesis o limitación afectada si la hay, y
// el contrato de tarifas con D-3 de la spec.
var Table = map[string]Rate{
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

// Cost devuelve el coste de una llamada y un booleano de disponibilidad (R5): un
// modelo ausente de la tabla devuelve (0, false) —"no disponible", distinto de un
// coste 0 real—; los tokens se contabilizan aparte aunque el coste no esté disponible.
//
// La escritura de caché llega partida por duración y se tarifa por duración: `cacheWrite5m` a `CacheWrite` y
// `cacheWrite1h` a `CacheWrite1h` (P-007 FR-002). Cuál es cuál lo decide quien lee el log, no esta función.
func Cost(model string, in, out, cacheWrite5m, cacheWrite1h, cacheRead int) (float64, bool) {
	r, ok := Table[model]
	if !ok {
		return 0, false
	}
	perM := func(tokens int, rate float64) float64 { return float64(tokens) / 1_000_000 * rate }
	return perM(in, r.Input) + perM(out, r.Output) + perM(cacheWrite5m, r.CacheWrite) + perM(cacheWrite1h, r.CacheWrite1h) +
		perM(cacheRead, r.CacheRead), true
}
