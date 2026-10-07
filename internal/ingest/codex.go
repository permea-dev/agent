package ingest

import (
	"encoding/json"
	"time"

	"github.com/permea-dev/agent/internal/event"
)

// ═══ P-008 · UNA LÍNEA DE CODEX ES UN EVENTO ═══════════════════════════════════════════════════
//
// Desde Codex 0.153.0, cada respuesta terminada deja una línea `token_usage_record` con su consumo y su
// `response_id` (P-008 descubrimiento §FASE 0 (b)). Cada una es UN evento (M-2): no hay espera ni cierre,
// y la pasada de Claude Code (P-007) no se aplica.
//
// ═══ LO QUE SE LEE, Y LO QUE NO ════════════════════════════════════════════════════════════════
//
// Del registro se leen sólo el `type`, el `timestamp` y, del `payload`, `response_id`, `session_id`,
// `turn_id` y las cuatro partidas de `usage` que usa FR-010. Los acumulados (`turn_token_usage`,
// `thread_token_usage`), `reasoning_output_tokens` y `total_tokens` no se leen.
//
// Los identificadores del proveedor NUNCA se copian al evento (FR-016): `response_id` sólo entra en el
// hash del `event_id` (`codex_eventid.go`) y `session_id`, con sal, en `event.Ref`. El coste no se
// calcula: lo pone la plataforma (D-3), así que aquí no se consulta `internal/pricing` (FR-013).

// ContextoCodex son los datos locales que añade el agente a un registro de Codex. Reutiliza el `Context`
// de Claude Code por sus datos locales y su resolutor; su `Pasada` no se usa con Codex (P-008 M-2).
type ContextoCodex struct {
	Context
	// DelTurno da el modelo y el directorio de trabajo del turno `turnID`. Lo pone quien lee el fichero
	// (P-008 FR-011, FR-014). Nil es válido: modelo y directorio vacíos.
	DelTurno func(turnID string) (modelo, cwd string)
}

// ClaseCodex clasifica una línea de una sesión de Codex (P-008 FR-027).
type ClaseCodex int

const (
	// NoEsRegistro es cualquier línea que no sea `token_usage_record`. No es una respuesta.
	NoEsRegistro ClaseCodex = iota
	// EventoCodex es una respuesta que produce un evento.
	EventoCodex
	// SinIdentificador es una respuesta sin `response_id` (P-008 FR-009).
	SinIdentificador
	// Incoherente es una respuesta cuyas partidas no se pueden usar (P-008 FR-027, E-3).
	Incoherente
)

// tipoRegistroCodex es el `type` de la línea que lleva el consumo de una respuesta.
const tipoRegistroCodex = "token_usage_record"

// lineaCodex es la envoltura de cualquier línea: el payload sólo se decodifica si es un registro, para que
// una línea válida de otro tipo, con otra forma, no cuente como corrupta.
type lineaCodex struct {
	Type      string          `json:"type"`
	Timestamp json.RawMessage `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// registroCodex es lo que se lee del payload de un `token_usage_record`. `usage` se guarda en crudo y se
// decodifica aparte, para que un defecto suyo no impida saber si hay `response_id` (E-4: sin identificador va antes).
type registroCodex struct {
	ResponseID string          `json:"response_id"`
	SessionID  string          `json:"session_id"`
	TurnID     string          `json:"turn_id"`
	Usage      json.RawMessage `json:"usage"`
}

// usoCodex son las cuatro partidas de FR-010, tal como las escribe Codex (`TokenUsage`, i64). Una partida
// ausente vale 0, como en Codex (`#[serde(default)]`; E-4).
type usoCodex struct {
	Entrada   int64 `json:"input_tokens"`
	Cache     int64 `json:"cached_input_tokens"`
	Escritura int64 `json:"cache_write_input_tokens"`
	Salida    int64 `json:"output_tokens"`
}

// usoDe decodifica `usage`. `ok == false` si no viene, es `null` o no decodifica (una partida no numérica).
func usoDe(crudo json.RawMessage) (*usoCodex, bool) {
	if len(crudo) == 0 || string(crudo) == "null" {
		return nil, false
	}
	var u usoCodex
	if err := json.Unmarshal(crudo, &u); err != nil {
		return nil, false
	}
	return &u, true
}

// coherente dice si las partidas se pueden usar (FR-027, E-3): ninguna negativa, y la caché y la escritura
// caben en la entrada, porque las dos van DENTRO de ella (FASE 0 (c)).
func (u *usoCodex) coherente() bool {
	if u == nil {
		return false
	}
	if u.Entrada < 0 || u.Cache < 0 || u.Escritura < 0 || u.Salida < 0 {
		return false
	}
	return u.Cache+u.Escritura <= u.Entrada
}

// LineaCodex lee una línea de una sesión de Codex. Devuelve su clase y, si es `EventoCodex`, su evento.
//
// Un error es SÓLO una línea «corrupta»: su envoltura no es JSON válido y no se puede leer su `type` (P-008
// FR-029, E-4). Quien lee decide qué hacer con ella. Establecido que es un registro, cualquier defecto de sus
// datos lo clasifica, no lo rompe (FR-027, E-3, E-4), en este orden: sin identificador → incoherente →
// evento. Las repetidas las cuenta quien lee el fichero, porque sólo él conoce la pasada.
func LineaCodex(line []byte, ctx ContextoCodex) (*event.Event, ClaseCodex, error) {
	var l lineaCodex
	if err := json.Unmarshal(line, &l); err != nil {
		return nil, NoEsRegistro, err
	}
	if l.Type != tipoRegistroCodex {
		return nil, NoEsRegistro, nil
	}
	var r registroCodex
	if err := json.Unmarshal(l.Payload, &r); err != nil {
		return nil, Incoherente, nil // E-4: el payload no decodifica
	}
	id, ok := derivarEventIDCodex(r.ResponseID)
	if !ok {
		return nil, SinIdentificador, nil
	}
	var momento time.Time
	if err := json.Unmarshal(l.Timestamp, &momento); err != nil {
		return nil, Incoherente, nil // E-4: sin `timestamp`, o mal formado
	}
	u, ok := usoDe(r.Usage)
	if !ok || !u.coherente() {
		return nil, Incoherente, nil // E-3: sin `usage`, negativa o que no cabe; E-4: no numérica
	}

	var modelo, cwd string
	if ctx.DelTurno != nil {
		modelo, cwd = ctx.DelTurno(r.TurnID)
	}
	ev := event.Event{
		SchemaVersion:       event.SchemaVersion,
		AgentVersion:        ctx.AgentVersion,
		EventID:             id,
		OccurredAt:          momento,
		Tool:                "codex",
		Model:               modelo,
		TokensInput:         int(u.Entrada - u.Cache - u.Escritura),
		TokensOutput:        int(u.Salida),
		TokensCacheCreation: int(u.Escritura),
		TokensCacheRead:     int(u.Cache),
		CostUSD:             0,
		CostAvailable:       false,
		ProjectRef:          ctx.Resolutor.Derivar(cwd, ctx.Salt),
		SessionRef:          event.Ref(ctx.Salt, r.SessionID),
		MachineRef:          event.Ref(ctx.Salt, ctx.MachineID),
		DevID:               ctx.DevID,
		OrgID:               ctx.OrgID,
	}
	return &ev, EventoCodex, nil
}
