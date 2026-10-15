package ingest

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"strconv"
)

// ═══ P-012 · LA IDENTIDAD DE UN EVENTO DE KIMI CODE ES SU POSICIÓN ══════════════════════════════════════════════════
//
// Contrato: `specs/012-lector-kimi/contracts/event-id-kimi.md`. Mismo esquema que los de 006 y 008–011, con espacio de nombres
// propio. Los tests de `kimi_eventid_test.go` reproducen sus vectores byte a byte: si esta función no los da, la que está mal es
// la función.
//
// Kimi Code no da ningún id por llamada (descubrimiento Q2; muestra 1), así que un registro se identifica por su POSICIÓN (D-3):
// S, el nombre de la carpeta de la sesión; A, el de la carpeta del agente; y N, el ordinal del registro en su `wire.jsonl`, desde
// 1 y contando los heredados. Ni horas ni partidas entran en el hash. Una copia de `/fork` lleva otra S y daría otro valor: por eso
// lo anterior a la última `forked` no se emite (FR-014, B3).
//
// Se replica el bucle de codificación, como 009–011, porque los `*_eventid.go` anteriores tienen 0 bytes de diff (FR-027). Se
// reutilizan sus CONSTANTES.

// herramientaKimi ocupa la posición en la que 006 pone `claude_code` (P-012 D-1).
const herramientaKimi = "kimi_code"

// derivarEventIDKimi devuelve el `event_id` del registro de Kimi Code en la posición (S, A, N). N va en decimal, sin ceros a la
// izquierda.
func derivarEventIDKimi(s, a string, n int) string {
	componentes := []string{dominioEventID, herramientaKimi, tipoRespuesta, s, a, strconv.Itoa(n)}

	h := sha256.New()
	var longitud [4]byte
	for _, c := range componentes {
		l := len(c)
		if l > math.MaxInt32 {
			// Inalcanzable con nombres de carpeta; el mismo resguardo que `hashEventID`.
			l = math.MaxInt32
		}
		binary.BigEndian.PutUint32(longitud[:], uint32(l))
		_, _ = h.Write(longitud[:]) // hash.Hash.Write nunca devuelve error
		_, _ = h.Write([]byte(c[:l]))
	}
	return hex.EncodeToString(h.Sum(nil)[:bytesEventID])
}
