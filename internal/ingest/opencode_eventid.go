package ingest

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
)

// ═══ P-010 · LA IDENTIDAD DE UN EVENTO DE OPENCODE SE DERIVA DEL `id` DE SU FILA ═════════════════════════════════
//
// Contrato: `specs/010-lector-opencode/contracts/event-id-opencode.md`. Mismo esquema que los de Claude Code (006),
// Codex (008) y Gemini (009), con espacio de nombres propio. Los tests de `opencode_eventid_test.go` reproducen sus
// vectores byte a byte: si esta función no los da, la que está mal es la función.
//
// ═══ POR QUÉ SE REPLICA LA CODIFICACIÓN UNA VEZ MÁS ═════════════════════════════════════════════════════════════════
//
// Las derivaciones de 006, 008 y 009 fijan su herramienta dentro de la función, y parametrizarlas cambiaría la frontera
// (P-010 FR-030). Se reutilizan sus CONSTANTES (`dominioEventID`, `bytesEventID`, `tipoRespuesta`) y se replica el bucle.
//
// ═══ POR QUÉ BASTA EL `id` (P-4) ═══════════════════════════════════════════════════════════════════════════════════
//
// Es único por llamada y estable mientras la fila se reescribe. El rowid no sirve de identidad: `VACUUM` lo renumera.
// La misma llamada en dos bases (canales) da el mismo `event_id`, y la plataforma la descarta. El `id` entra sólo en el
// hash: nunca se copia a ningún campo del evento.

const (
	// herramientaOpenCode ocupa la posición en la que 006 pone `claude_code`, 008 `codex` y 009 `gemini`: el mismo valor
	// llegado de dos herramientas da `event_id` distintos por construcción. El TIPO es `tipoRespuesta`, como en 008 y 009.
	herramientaOpenCode = "opencode"
)

// derivarEventIDOpenCode devuelve el `event_id` de la llamada de OpenCode cuya fila de `session_message` tiene este `id`.
// `ok == false` cuando no viene: no queda nada estable de lo que derivar, y la llamada no debe emitirse.
func derivarEventIDOpenCode(id string) (eventID string, ok bool) {
	if id == "" {
		return "", false
	}
	componentes := []string{dominioEventID, herramientaOpenCode, tipoRespuesta, id}

	h := sha256.New()
	var longitud [4]byte
	for _, c := range componentes {
		n := len(c)
		if n > math.MaxInt32 {
			// Inalcanzable con identificadores de 30 caracteres; el mismo resguardo que `hashEventID`.
			n = math.MaxInt32
		}
		binary.BigEndian.PutUint32(longitud[:], uint32(n))
		_, _ = h.Write(longitud[:]) // hash.Hash.Write nunca devuelve error
		_, _ = h.Write([]byte(c[:n]))
	}
	return hex.EncodeToString(h.Sum(nil)[:bytesEventID]), true
}
