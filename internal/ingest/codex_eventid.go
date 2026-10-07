package ingest

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
)

// ═══ P-008 · LA IDENTIDAD DE UN EVENTO DE CODEX SE DERIVA DE SU RESPUESTA ══════════════════════
//
// Contrato: `specs/008-lector-codex/contracts/event-id-codex.md`. Mismo esquema que el de Claude Code
// (`eventid.go`, 006) con espacio de nombres propio. Los tests de `codex_eventid_test.go` reproducen sus
// vectores byte a byte: si esta función no los da, la que está mal es la función.
//
// ═══ POR QUÉ SE REPLICA LA CODIFICACIÓN Y NO SE REUTILIZA `hashEventID` ════════════════════════
//
// `hashEventID` fija la herramienta `claude_code` dentro de la función (`eventid.go:67`). Darle otra
// exigiría tocar `eventid.go`, que P-008 M-1 deja con 0 bytes de diff. Se reutilizan sus CONSTANTES
// (`dominioEventID`, `bytesEventID`) y se replica el bucle de codificación.
//
// El identificador de la respuesta entra sólo en el hash: nunca se copia a ningún campo del evento.

const (
	// herramientaCodex ocupa la posición en la que 006 pone `claude_code`: el mismo valor llegado de las
	// dos herramientas da `event_id` distintos por construcción.
	herramientaCodex = "codex"
	// tipoRespuesta ocupa la posición de TIPO: Codex sólo tiene una forma, la respuesta.
	tipoRespuesta = "respuesta"
)

// derivarEventIDCodex devuelve el `event_id` de la respuesta de Codex identificada por `responseID`.
// `ok == false` cuando no viene: no queda nada estable de lo que derivar, y el registro no debe emitirse.
func derivarEventIDCodex(responseID string) (eventID string, ok bool) {
	if responseID == "" {
		return "", false
	}
	componentes := []string{dominioEventID, herramientaCodex, tipoRespuesta, responseID}

	h := sha256.New()
	var longitud [4]byte
	for _, c := range componentes {
		n := len(c)
		if n > math.MaxInt32 {
			// Inalcanzable con identificadores de 55 caracteres; el mismo resguardo que `hashEventID`.
			n = math.MaxInt32
		}
		binary.BigEndian.PutUint32(longitud[:], uint32(n))
		_, _ = h.Write(longitud[:]) // hash.Hash.Write nunca devuelve error
		_, _ = h.Write([]byte(c[:n]))
	}
	return hex.EncodeToString(h.Sum(nil)[:bytesEventID]), true
}
