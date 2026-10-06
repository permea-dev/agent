// Package ingest convierte logs de herramientas en Events de frontera.
package ingest

import (
	"encoding/json"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/pricing"
	"github.com/permea-dev/agent/internal/project"
)

// rawRecord decodifica SOLO los campos permitidos del JSONL de Claude Code.
// Deliberadamente NO incluye message.content ni ningún campo de texto: lo que no
// se decodifica aquí, no entra en el proceso. Esa es la garantía deny-by-default.
//
// GUARDIA DE FRONTERA (Principio I, no negociable — NO AMPLIAR con contenido):
// Está PROHIBIDO añadir a rawRecord cualquier campo que transporte contenido del
// usuario o del modelo: message.content, texto de respuestas, código, diffs,
// argumentos o resultados de herramientas, rutas en claro, secretos, o CUALQUIER
// campo nuevo/desconocido del log con contenido. Los campos desconocidos del origen
// se ignoran por construcción (encoding/json descarta lo no declarado). Solo se
// admiten métricas y metadatos derivados de la allowlist de contracts/boundary-event.md.
// El golden test (boundary_test.go) y TestEvent_OnlyAllowlistKeys fallan si esto se viola.
//
// P-006 · LOS DOS IDENTIFICADORES DEL PROVEEDOR (`message.id` y `requestId`): son metadatos
// técnicos, no contenido, y se admiten SÓLO para derivar el `event_id` de la allowlist
// (`derivarEventID`, specs/006-medicion-fiel/contracts/event-id.md). NUNCA se copian a ningún
// campo del evento, ni enteros ni en fragmento (P-006 FR-003): la denylist del golden lleva sus
// centinelas y sus núcleos.
//
// P-007 · EL DESGLOSE DE LA ESCRITURA DE CACHÉ (FR-001): de `usage.cache_creation` se admiten SÓLO
// `ephemeral_5m_input_tokens` y `ephemeral_1h_input_tokens`, dos NÚMEROS de consumo. Nada más de ese
// objeto, y la ampliación no cubre ningún otro campo. El desglose sirve para tarifar y NUNCA cruza la
// frontera: el evento lleva el total, `cache_creation_input_tokens` (FR-005).
type rawRecord struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"sessionId"`
	Cwd       string    `json:"cwd"`
	RequestID string    `json:"requestId"`
	Message   struct {
		ID    string `json:"id"`
		Model string `json:"model"`
		Usage struct {
			InputTokens         int `json:"input_tokens"`
			OutputTokens        int `json:"output_tokens"`
			CacheCreationTokens int `json:"cache_creation_input_tokens"`
			CacheReadTokens     int `json:"cache_read_input_tokens"`
			// Punteros para distinguir «no viene» de «viene a 0» (P-007 FR-003).
			CacheCreation struct {
				Ephemeral5m *int `json:"ephemeral_5m_input_tokens"`
				Ephemeral1h *int `json:"ephemeral_1h_input_tokens"`
			} `json:"cache_creation"`
		} `json:"usage"`
	} `json:"message"`
}

// Context son los datos locales que añade el agente (nunca provienen del log).
type Context struct {
	Salt         string
	MachineID    string
	DevID        string
	OrgID        string
	AgentVersion string
	// Resolutor da a la derivación de identidad de proyecto memoria de UNA PASADA (P-004
	// T032). Su ámbito es el de este Context, así que quien quiera la caché lo instancia por
	// pasada — ver `generate()` en cmd/permea.
	//
	// NIL ES VÁLIDO: sin resolutor se deriva igual, solo que sin el ahorro. Ningún punto de
	// construcción existente tiene que cambiar para seguir funcionando.
	Resolutor *project.Resolutor
	// Pasada da a la emisión memoria de UNA PASADA (P-006 FR-001, FR-033): un mensaje cuyas líneas
	// se lean en la misma pasada produce UN evento, y la pasada cuenta lo que lee. Mismo patrón que
	// el Resolutor: quien la quiera la instancia por pasada —`generate()` y `dryRun()` en cmd/permea—.
	//
	// NIL ES VÁLIDO: sin pasada se deriva y se emite igual, sólo que sin deduplicar ni contar. Entre
	// pasadas no hay memoria: los repetidos los descarta la plataforma por `(org_id, event_id)`.
	Pasada *Pasada
}

// modeloSintetico es el modelo con el que Claude Code marca los mensajes que genera él mismo, sin
// llamada al modelo. No son consumo y NUNCA producen evento (P-006 FR-007).
const modeloSintetico = "<synthetic>"

// FromClaudeCodeLine convierte una línea JSONL en un Event de frontera.
// Devuelve (nil, nil) si la línea no es una llamada facturable, si es `<synthetic>`, o si no trae
// ninguno de los dos identificadores del mensaje (P-006 FR-006): sin ellos no hay `event_id` estable,
// y uno aleatorio volvería a contar el mismo mensaje varias veces.
func FromClaudeCodeLine(line []byte, ctx Context) (*event.Event, error) {
	var r rawRecord
	if err := json.Unmarshal(line, &r); err != nil {
		return nil, err
	}
	if r.Type != "assistant" || r.Message.Model == "" {
		return nil, nil
	}
	ctx.Pasada.contarFacturable()
	if r.Message.Model == modeloSintetico {
		ctx.Pasada.contarSintetica()
		return nil, nil
	}
	// P-006 FR-002/FR-008: el `event_id` se DERIVA del mensaje; ya no se acuña uno aleatorio.
	id, ok := derivarEventID(r.Message.ID, r.RequestID)
	if !ok {
		ctx.Pasada.contarSinIdentificador()
		return nil, nil
	}
	u := r.Message.Usage
	cw5m, cw1h, conDesglose := desglosarEscritura(u.CacheCreationTokens, u.CacheCreation.Ephemeral5m, u.CacheCreation.Ephemeral1h)
	if !conDesglose && u.CacheCreationTokens > 0 {
		ctx.Pasada.contarSinDesglose() // P-007 FR-003 (E-4): sin escritura no hubo hipótesis que contar
	}
	c := consumo{u.InputTokens, u.OutputTokens, u.CacheCreationTokens, u.CacheReadTokens, cw5m, cw1h}
	// El evento base sale de la primera línea del mensaje: identidad, momento, modelo y referencias. Los
	// tokens y el coste los pone `conConsumo`.
	base := func() event.Event {
		return event.Event{
			SchemaVersion: event.SchemaVersion,
			AgentVersion:  ctx.AgentVersion,
			EventID:       id,
			OccurredAt:    r.Timestamp,
			Tool:          "claude_code",
			Model:         r.Message.Model,
			ProjectRef:    ctx.Resolutor.Derivar(r.Cwd, ctx.Salt),
			SessionRef:    event.Ref(ctx.Salt, r.SessionID),
			MachineRef:    event.Ref(ctx.Salt, ctx.MachineID),
			DevID:         ctx.DevID,
			OrgID:         ctx.OrgID,
		}
	}
	// P-007 FR-009: con pasada, la línea se ACUMULA en su mensaje, que sale al cerrar el fichero con el
	// máximo de cada partida. Sin pasada, cada línea se emite como hasta ahora.
	if ctx.Pasada.acumular(id, c, r.Timestamp, base) {
		return nil, nil
	}
	ev := conConsumo(base(), c)
	return &ev, nil
}

// conConsumo completa el evento base con los tokens de `c` y su coste, tarifando la escritura de caché por
// duración (P-007 FR-002). El evento lleva el TOTAL de la escritura; el desglose no cruza la frontera (FR-005).
func conConsumo(base event.Event, c consumo) event.Event {
	base.TokensInput = c.entrada
	base.TokensOutput = c.salida
	base.TokensCacheCreation = c.escrituraCache
	base.TokensCacheRead = c.lecturaCache
	base.CostUSD, base.CostAvailable = pricing.Cost(base.Model, c.entrada, c.salida, c.escritura5m, c.escritura1h, c.lecturaCache)
	return base
}

// desglosarEscritura reparte la escritura de caché de una línea por duración (P-007 FR-003, FR-004). Si la
// línea trae las DOS cifras del desglose y suman el total, son las del log. Si no las trae, o no suman, toda
// la escritura va a 1 hora (hipótesis P-1; Q-4) y `conDesglose` es false, para que la pasada lo cuente. El
// total no se toca nunca: es lo que cruza la frontera (FR-005).
func desglosarEscritura(total int, a5m, a1h *int) (cw5m, cw1h int, conDesglose bool) {
	if a5m != nil && a1h != nil && *a5m+*a1h == total {
		return *a5m, *a1h, true
	}
	return 0, total, false
}
