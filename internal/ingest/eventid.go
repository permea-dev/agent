package ingest

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
)

// ═══ P-006 · LA IDENTIDAD DEL EVENTO SE DERIVA DEL MENSAJE ═════════════════════════════════
//
// El `event_id` deja de ser aleatorio. Claude Code escribe VARIAS líneas por mensaje, todas con el
// mismo consumo, y un identificador aleatorio por línea contaba cada mensaje varias veces sin que la
// plataforma pudiera notarlo. Derivado del mensaje, el mismo mensaje da el mismo `event_id` en
// cualquier línea, pasada o instalación, y la plataforma descarta los repetidos por
// `(org_id, event_id)`.
//
// El contrato —dominio, codificación, truncado y VECTORES NORMATIVOS— está en
// `specs/006-medicion-fiel/contracts/event-id.md`. Los tests de `eventid_test.go` reproducen esos
// vectores byte a byte: si esta función no los da, la que está mal es la función.
//
// ═══ LO QUE NO ENTRA EN EL HASH, Y POR QUÉ ═════════════════════════════════════════════════
//
// Ni sal, ni máquina, ni desarrollador, ni organización, ni contenido. La sal particulariza los
// identificadores SENSIBLES (ruta, sesión, máquina); aquí haría lo contrario de lo que se busca, que
// es que dos instalaciones produzcan el MISMO valor para el mismo mensaje. Los identificadores del
// proveedor entran sólo en el hash y nunca se copian a ningún campo del evento.

const (
	// dominioEventID versiona la derivación: una v2 daría valores disjuntos de los de v1.
	dominioEventID = "permea/event_id/v1"
	// herramientaClaudeCode separa este lector de los de otras herramientas, cuyos identificadores
	// podrían coincidir por azar con los de Claude Code.
	herramientaClaudeCode = "claude_code"

	// Los tres tipos separan las formas: sin ellos, el mismo valor llegado como `message.id` solo y
	// como `requestId` solo daría el mismo `event_id`.
	tipoPar           = "par"
	tipoSoloMessageID = "solo_message_id"
	tipoSoloRequestID = "solo_request_id"

	// bytesEventID es el truncado del SHA-256: 16 bytes, 32 hex, la forma que el campo ya tenía.
	bytesEventID = 16
)

// derivarEventID devuelve el `event_id` del mensaje identificado por `messageID` y `requestID`.
// `ok == false` cuando no viene ninguno de los dos: no queda nada estable de lo que derivar, y la
// línea no debe emitirse (P-006 FR-006).
func derivarEventID(messageID, requestID string) (eventID string, ok bool) {
	switch {
	case messageID != "" && requestID != "":
		return hashEventID(tipoPar, messageID, requestID), true
	case messageID != "":
		return hashEventID(tipoSoloMessageID, messageID), true
	case requestID != "":
		return hashEventID(tipoSoloRequestID, requestID), true
	default:
		return "", false
	}
}

// hashEventID aplica la derivación del contrato: cada componente —dominio, herramienta, tipo y los
// identificadores— va precedido de su longitud en 4 bytes big-endian, y del SHA-256 del resultado se
// toman los primeros 16 bytes en hexadecimal. El prefijo de longitud impide que `("a","bc")` y
// `("ab","c")` se codifiquen igual.
func hashEventID(tipo string, identificadores ...string) string {
	componentes := append([]string{dominioEventID, herramientaClaudeCode, tipo}, identificadores...)

	h := sha256.New()
	var longitud [4]byte
	for _, c := range componentes {
		n := len(c)
		if n > math.MaxInt32 {
			// Inalcanzable con identificadores de 28 caracteres. Existe porque la longitud se codifica
			// en 4 bytes: un componente que no cupiera se codifica por sus primeros MaxInt32 bytes con
			// esa longitud, bien formado en vez de desbordado. MaxInt32 y no MaxUint32 para que la
			// comparación compile también donde `int` es de 32 bits.
			n = math.MaxInt32
		}
		binary.BigEndian.PutUint32(longitud[:], uint32(n))
		_, _ = h.Write(longitud[:]) // hash.Hash.Write nunca devuelve error
		_, _ = h.Write([]byte(c[:n]))
	}
	return hex.EncodeToString(h.Sum(nil)[:bytesEventID])
}
