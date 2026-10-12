package ingest

import (
	"encoding/json"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/sqlitero"
)

// ═══ P-010 · UNA LLAMADA DE OPENCODE ES, COMO MUCHO, UN EVENTO SIN COSTE ═════════════════════════════════════════
//
// OpenCode 2.x guarda cada respuesta del modelo como una fila `assistant` de `session_message`, con el consumo en el
// JSON de `data`. Esta función lee UNA fila que la pasada ya tiene por llamada (P-1: completada, o nueva con un `data`
// que no es un objeto); qué filas son llamadas, cuáles siguen en curso y cuáles se abandonan lo decide la pasada (B4).
//
// ═══ LO QUE SE LEE, Y LO QUE NO ════════════════════════════════════════════════════════════════════════════════
//
// De `data` se leen sólo `tokens`, `model.id` y `time.completed`, primero EN CRUDO: un campo con otra forma clasifica la
// llamada, nunca rompe la pasada. `$.cost` no se lee (D-1: sólo lo usa el contraste de `--scan`), y aquí no se consulta
// `internal/pricing`. El `id` sólo entra en el hash del `event_id`; `session_id` y `directory`, sólo con sal (FR-020).

// ContextoOpenCode son los datos locales que añade el agente a una llamada de OpenCode. Reutiliza el `Context` de
// Claude Code por sus datos locales y su resolutor; su `Pasada` no se usa con OpenCode.
type ContextoOpenCode struct {
	Context
	// Ahora es el reloj de la pasada (D-010-P9). Nil → `time.Now`.
	Ahora func() time.Time
	// SessionID, Directory y Path son los de la fila de `session_v2` de la llamada (P-6). Vacíos si esa fila falta.
	SessionID string
	Directory string
	Path      string
}

// FilaOpenCode son las columnas de una fila de `session_message` que lee una llamada.
type FilaOpenCode struct {
	ID   sqlitero.Valor
	Data []byte
}

// ClaseOpenCode clasifica una llamada de OpenCode (P-010 FR-021).
type ClaseOpenCode int

// Las clases de FR-021. El cero no es ninguna: una llamada siempre tiene una. «Sin modelo» no es una clase, sino una
// marca sobre el evento: su `model` vacío (P-5).
const (
	// EventoOpenCode es una llamada que produce un evento.
	EventoOpenCode ClaseOpenCode = iota + 1
	// SinIdentificadorOpenCode es una llamada sin `id` textual no vacío (FR-014).
	SinIdentificadorOpenCode
	// IncoherenteOpenCode es una llamada cuyos datos no se pueden usar (FR-021).
	IncoherenteOpenCode
)

// datosOpenCode es lo que se lee de `data`, en crudo.
type datosOpenCode struct {
	Tokens json.RawMessage `json:"tokens"`
	Model  json.RawMessage `json:"model"`
	Time   struct {
		Completed json.RawMessage `json:"completed"`
	} `json:"time"`
}

// tokensOpenCode son las partidas de P-2, tal como las escribe OpenCode. Una ausente vale 0.
type tokensOpenCode struct {
	Input     int64 `json:"input"`
	Output    int64 `json:"output"`
	Reasoning int64 `json:"reasoning"`
	Cache     struct {
		Read  int64 `json:"read"`
		Write int64 `json:"write"`
	} `json:"cache"`
}

// coherente: ninguna partida negativa.
func (k tokensOpenCode) coherente() bool {
	return k.Input >= 0 && k.Output >= 0 && k.Reasoning >= 0 && k.Cache.Read >= 0 && k.Cache.Write >= 0
}

// esObjeto dice si un campo vino como un objeto JSON.
func esObjeto(crudo json.RawMessage) bool {
	return len(crudo) > 0 && crudo[0] == '{'
}

// LlamadaOpenCode lee una llamada y devuelve su clase y, si es un evento, el evento. Clasifica en el orden de
// D-010-P14: sin identificador → incoherente → evento.
func LlamadaOpenCode(fila FilaOpenCode, ctx ContextoOpenCode) (*event.Event, ClaseOpenCode) {
	var id string
	if fila.ID.Tipo == sqlitero.Texto {
		id = string(fila.ID.B)
	}
	eventID, ok := derivarEventIDOpenCode(id)
	if !ok {
		return nil, SinIdentificadorOpenCode
	}
	var d datosOpenCode
	errData := json.Unmarshal(fila.Data, &d) // no es JSON, no es un objeto, o `time` no es un objeto
	var k tokensOpenCode
	errTokens := json.Unmarshal(d.Tokens, &k) // una partida no numérica o no entera, o `cache` no es un objeto
	var completada int64
	errMomento := json.Unmarshal(d.Time.Completed, &completada) // `time.completed` ausente o no entero
	if errData != nil || !esObjeto(d.Tokens) || errTokens != nil || errMomento != nil || ausente(d.Time.Completed) ||
		!k.coherente() {
		return nil, IncoherenteOpenCode
	}

	var m struct {
		ID json.RawMessage `json:"id"`
	}
	_ = json.Unmarshal(d.Model, &m) // un `model` que no es un objeto deja el modelo vacío: «sin modelo» (P-5)
	modelo := textoDe(m.ID)

	var sesion, proyecto string
	if ctx.SessionID != "" { // sin la fila de `session_v2`, las dos vacías (P-6)
		sesion = event.Ref(ctx.Salt, ctx.SessionID)
		proyecto = ctx.Resolutor.Derivar(ctx.Directory, ctx.Salt)
	}
	ev := event.Event{
		SchemaVersion:       event.SchemaVersion,
		AgentVersion:        ctx.AgentVersion,
		EventID:             eventID,
		OccurredAt:          time.UnixMilli(completada).UTC(),
		Tool:                herramientaOpenCode,
		Model:               modelo,
		TokensInput:         int(k.Input),
		TokensOutput:        int(k.Output + k.Reasoning),
		TokensCacheCreation: int(k.Cache.Write),
		TokensCacheRead:     int(k.Cache.Read),
		CostUSD:             0,
		CostAvailable:       false,
		ProjectRef:          proyecto,
		SessionRef:          sesion,
		MachineRef:          event.Ref(ctx.Salt, ctx.MachineID),
		DevID:               ctx.DevID,
		OrgID:               ctx.OrgID,
	}
	return &ev, EventoOpenCode
}
