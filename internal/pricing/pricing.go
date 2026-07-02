// Package pricing calcula el coste en LOCAL. El agente nunca depende del backend
// para conocer el coste. Tabla ilustrativa: actualizable con precios vigentes.
package pricing

// Rate en USD por millón de tokens.
type Rate struct {
	Input      float64
	Output     float64
	CacheWrite float64
	CacheRead  float64
}

// Table empaquetada con el binario.
var Table = map[string]Rate{
	"claude-opus-4-6":   {Input: 15, Output: 75, CacheWrite: 18.75, CacheRead: 1.5},
	"claude-sonnet-4-6": {Input: 3, Output: 15, CacheWrite: 3.75, CacheRead: 0.3},
	"claude-haiku-4-5":  {Input: 1, Output: 5, CacheWrite: 1.25, CacheRead: 0.1},
}

// Cost devuelve el coste de una llamada. Modelo desconocido -> 0 (se marca n/a aparte).
func Cost(model string, in, out, cacheCreate, cacheRead int) float64 {
	r, ok := Table[model]
	if !ok {
		return 0
	}
	perM := func(tokens int, rate float64) float64 { return float64(tokens) / 1_000_000 * rate }
	return perM(in, r.Input) + perM(out, r.Output) + perM(cacheCreate, r.CacheWrite) + perM(cacheRead, r.CacheRead)
}
