package ingest

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
)

// ═══ P-009 · LA IDENTIDAD DE UN EVENTO DE GEMINI CLI SE DERIVA DEL `id` DE SU MENSAJE ═══════════════
//
// Contrato: `specs/009-lector-gemini/contracts/event-id-gemini.md`. Mismo esquema que los de Claude Code
// (`eventid.go`, 006) y Codex (`codex_eventid.go`, 008), con espacio de nombres propio. Los tests de
// `gemini_eventid_test.go` reproducen sus vectores byte a byte: si esta función no los da, la que está mal es
// la función.
//
// ═══ POR QUÉ SE REPLICA LA CODIFICACIÓN UNA VEZ MÁS ═════════════════════════════════════════════════
//
// `hashEventID` fija `claude_code` y `derivarEventIDCodex` fija `codex` dentro de la función. Parametrizarlas
// exigiría tocar `eventid.go` o `codex_eventid.go`, que P-009 FR-027 deja con 0 bytes de diff. Se reutilizan
// sus CONSTANTES (`dominioEventID`, `bytesEventID`, `tipoRespuesta`) y se replica el bucle de codificación.
//
// ═══ POR QUÉ BASTA EL `id` (P-2 (a)) ═════════════════════════════════════════════════════════════════
//
// La CLI lo genera con `randomUUID` y lo conserva cada vez que reescribe el mensaje. Las copias de una sesión
// (migración de carpeta, conversión `.json` → `.jsonl`, `$set.messages`) conservan también el `sessionId`, así que
// añadirlo no separaría nada; y si una copia llegara a otra sesión, con sólo el `id` sigue siendo UN evento.
//
// El `id` entra sólo en el hash: nunca se copia a ningún campo del evento.

const (
	// herramientaGemini ocupa la posición en la que 006 pone `claude_code` y 008 pone `codex`: el mismo valor
	// llegado de dos herramientas da `event_id` distintos por construcción. El TIPO es `tipoRespuesta`, como en 008.
	herramientaGemini = "gemini"
)

// derivarEventIDGemini devuelve el `event_id` de la respuesta de Gemini CLI cuyo mensaje tiene este `id`.
// `ok == false` cuando no viene: no queda nada estable de lo que derivar, y la respuesta no debe emitirse.
func derivarEventIDGemini(id string) (eventID string, ok bool) {
	if id == "" {
		return "", false
	}
	componentes := []string{dominioEventID, herramientaGemini, tipoRespuesta, id}

	h := sha256.New()
	var longitud [4]byte
	for _, c := range componentes {
		n := len(c)
		if n > math.MaxInt32 {
			// Inalcanzable con identificadores de 36 caracteres; el mismo resguardo que `hashEventID`.
			n = math.MaxInt32
		}
		binary.BigEndian.PutUint32(longitud[:], uint32(n))
		_, _ = h.Write(longitud[:]) // hash.Hash.Write nunca devuelve error
		_, _ = h.Write([]byte(c[:n]))
	}
	return hex.EncodeToString(h.Sum(nil)[:bytesEventID]), true
}
