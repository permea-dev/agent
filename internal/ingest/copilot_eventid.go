package ingest

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
)

// ═══ P-011 · LA IDENTIDAD DE UN EVENTO DE COPILOT CLI SE DERIVA DE UN ID DEL SERVIDOR ═════════════════════════════════
//
// Contrato: `specs/011-lector-copilot/contracts/event-id-copilot.md`. Mismo esquema que los de Claude Code (`eventid.go`,
// 006), Codex (008), Gemini (009) y OpenCode (010), con espacio de nombres propio. Los tests de `copilot_eventid_test.go`
// reproducen sus vectores byte a byte: si esta función no los da, la que está mal es la función.
//
// Dos clases de registro, dos TIPOS: `respuesta` *(`tipoRespuesta`, el de 008–010)* con el `usageId` de un
// `session.usage_record`, y `compactacion` con el `requestId` de un `session.compaction_complete`. El tipo separa un
// `requestId` de un `usageId` iguales por azar. Los dos son ids del SERVIDOR: una copia de las líneas (una bifurcación de
// la sesión) da el mismo `event_id`, y la plataforma la descarta. El id entra sólo en el hash: nunca se copia al evento.
//
// Se replica el bucle de codificación, como 009 y 010, porque `eventid.go`, `codex_eventid.go`, `gemini_eventid.go` y
// `opencode_eventid.go` tienen 0 bytes de diff (FR-027). Se reutilizan sus CONSTANTES.

const (
	// herramientaCopilot ocupa la posición en la que 006 pone `claude_code` (P-011 D-1).
	herramientaCopilot = "copilot_cli"
	// tipoCompactacionCopilot es el TIPO de los registros `session.compaction_complete` (P-011 P-1).
	tipoCompactacionCopilot = "compactacion"
)

// derivarEventIDCopilot devuelve el `event_id` del registro de Copilot CLI de tipo `tipo` con este id. `ok == false`
// cuando el id está vacío: no queda nada estable de lo que derivar, y el registro no debe emitirse (FR-014).
func derivarEventIDCopilot(tipo, id string) (eventID string, ok bool) {
	if id == "" {
		return "", false
	}
	componentes := []string{dominioEventID, herramientaCopilot, tipo, id}

	h := sha256.New()
	var longitud [4]byte
	for _, c := range componentes {
		n := len(c)
		if n > math.MaxInt32 {
			// Inalcanzable con ids de decenas de caracteres; el mismo resguardo que `hashEventID`.
			n = math.MaxInt32
		}
		binary.BigEndian.PutUint32(longitud[:], uint32(n))
		_, _ = h.Write(longitud[:]) // hash.Hash.Write nunca devuelve error
		_, _ = h.Write([]byte(c[:n]))
	}
	return hex.EncodeToString(h.Sum(nil)[:bytesEventID]), true
}
