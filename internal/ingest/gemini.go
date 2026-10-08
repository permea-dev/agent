package ingest

import (
	"encoding/json"
	"time"

	"github.com/permea-dev/agent/internal/event"
)

// ═══ P-009 · UNA APARICIÓN DE UNA RESPUESTA DE GEMINI CLI ES, COMO MUCHO, UN EVENTO ═══════════════
//
// Gemini CLI escribe cada respuesta del modelo como un mensaje `type: "gemini"` con un objeto `tokens` de seis
// partidas, copiadas de `usageMetadata` (P-009 descubrimiento Q2). El mismo mensaje APARECE varias veces: en su
// línea, en las líneas que lo reescriben y dentro de `$set.messages`. Esta función lee UNA aparición; quién es
// la primera, y qué es repetida, lo decide quien recorre el fichero (B3, FR-008).
//
// ═══ LO QUE SE LEE, Y LO QUE NO ════════════════════════════════════════════════════════════════
//
// Del mensaje se leen sólo `id`, `timestamp`, `type`, `model` y `tokens`. El contenido, los `thoughts` (el texto
// del razonamiento) y `toolCalls` no se leen. Todo se decodifica primero EN CRUDO: un campo con otra forma
// clasifica la aparición, nunca la convierte en corrupta (P-009 FR-019, FR-020; 008 E-4).
//
// El `id` sólo entra en el hash del `event_id` (`gemini_eventid.go`); el `sessionId` y el texto de
// `.project_root`, sólo con sal (FR-017). El coste no se calcula (D-4): aquí no se consulta `internal/pricing`.

// ContextoGemini son los datos locales que añade el agente a una respuesta de Gemini CLI. Reutiliza el
// `Context` de Claude Code por sus datos locales y su resolutor; su `Pasada` no se usa con Gemini.
type ContextoGemini struct {
	Context
	// SessionID es el `sessionId` de la cabecera del fichero (P-009 FR-015). Vacío si no hay cabecera.
	SessionID string
	// ProjectRoot es el texto de `.project_root` de la carpeta `<slug>` (P-009 FR-014). Vacío si no existe.
	ProjectRoot string
}

// ClaseGemini clasifica una aparición de un mensaje de Gemini CLI (P-009 FR-019).
type ClaseGemini int

const (
	// NoEsRespuestaGemini es cualquier mensaje que no sea `type: "gemini"` con un objeto `tokens`: el de
	// usuario, el de información, y el de Gemini sin `tokens` o con `tokens: null` (P-009 FR-006).
	NoEsRespuestaGemini ClaseGemini = iota
	// EventoGemini es una respuesta que produce un evento.
	EventoGemini
	// SinIdentificadorGemini es una respuesta sin `id` textual no vacío (P-009 FR-009).
	SinIdentificadorGemini
	// IncoherenteGemini es una respuesta cuyos datos no se pueden usar (P-009 FR-019).
	IncoherenteGemini
)

// MarcasGemini son las marcas de un evento (P-009 FR-019): subconjuntos de «eventos», no clases.
type MarcasGemini struct {
	// SinModelo: el mensaje no trae un `model` textual no vacío (P-6).
	SinModelo bool
	// TotalDescuadrado: `total` está y no es `input + output + thoughts + tool` (P-5).
	TotalDescuadrado bool
}

// tipoRespuestaGemini es el `type` del mensaje que lleva el consumo de una respuesta.
const tipoRespuestaGemini = "gemini"

// mensajeGemini es la envoltura de una aparición, toda en crudo.
type mensajeGemini struct {
	ID        json.RawMessage `json:"id"`
	Timestamp json.RawMessage `json:"timestamp"`
	Type      json.RawMessage `json:"type"`
	Model     json.RawMessage `json:"model"`
	Tokens    json.RawMessage `json:"tokens"`
}

// tokensGemini son las partidas de D-2, tal como las escribe la CLI. Una ausente vale 0; `total` es un puntero
// para saber si vino (P-5).
type tokensGemini struct {
	Input    int64  `json:"input"`
	Output   int64  `json:"output"`
	Cached   int64  `json:"cached"`
	Thoughts int64  `json:"thoughts"`
	Tool     int64  `json:"tool"`
	Total    *int64 `json:"total"`
}

// coherente: ninguna partida negativa, y la caché cabe en la entrada, porque va DENTRO de ella (Q3).
func (k tokensGemini) coherente() bool {
	if k.Input < 0 || k.Output < 0 || k.Cached < 0 || k.Thoughts < 0 || k.Tool < 0 {
		return false
	}
	return k.Cached <= k.Input
}

// textoDe devuelve el valor de un campo si es un texto JSON, y "" si falta, es `null` u otra cosa.
func textoDe(crudo json.RawMessage) string {
	var s string
	if len(crudo) == 0 || json.Unmarshal(crudo, &s) != nil {
		return ""
	}
	return s
}

// ausente dice si un campo no vino o vino como `null`.
func ausente(crudo json.RawMessage) bool {
	return len(crudo) == 0 || string(crudo) == "null"
}

// RespuestaGemini lee una aparición de un mensaje y devuelve su clase y, si es un evento, el evento y sus marcas.
//
// Un error es SÓLO una aparición cuya envoltura no es un objeto JSON: quien lee decide qué hacer con ella. Lo
// demás clasifica, en el orden de D-009-P14: no es respuesta → sin identificador → incoherente → evento. Las
// repetidas las cuenta quien recorre el fichero.
func RespuestaGemini(crudo []byte, ctx ContextoGemini) (*event.Event, ClaseGemini, MarcasGemini, error) {
	var m mensajeGemini
	if err := json.Unmarshal(crudo, &m); err != nil {
		return nil, NoEsRespuestaGemini, MarcasGemini{}, err
	}
	if textoDe(m.Type) != tipoRespuestaGemini || ausente(m.Tokens) {
		return nil, NoEsRespuestaGemini, MarcasGemini{}, nil
	}
	id, ok := derivarEventIDGemini(textoDe(m.ID))
	if !ok {
		return nil, SinIdentificadorGemini, MarcasGemini{}, nil
	}
	var k tokensGemini
	errTokens := json.Unmarshal(m.Tokens, &k) // un número, un texto o una partida no entera: no decodifica
	var momento time.Time
	errMomento := json.Unmarshal(m.Timestamp, &momento) // sin `timestamp`, o mal formado
	if errTokens != nil || errMomento != nil || !k.coherente() {
		return nil, IncoherenteGemini, MarcasGemini{}, nil
	}

	modelo := textoDe(m.Model)
	marcas := MarcasGemini{
		SinModelo:        modelo == "",
		TotalDescuadrado: k.Total != nil && *k.Total != k.Input+k.Output+k.Thoughts+k.Tool,
	}
	ev := event.Event{
		SchemaVersion:       event.SchemaVersion,
		AgentVersion:        ctx.AgentVersion,
		EventID:             id,
		OccurredAt:          momento,
		Tool:                herramientaGemini,
		Model:               modelo,
		TokensInput:         int(k.Input - k.Cached + k.Tool),
		TokensOutput:        int(k.Output + k.Thoughts),
		TokensCacheCreation: 0,
		TokensCacheRead:     int(k.Cached),
		CostUSD:             0,
		CostAvailable:       false,
		ProjectRef:          ctx.Resolutor.Derivar(ctx.ProjectRoot, ctx.Salt),
		SessionRef:          event.Ref(ctx.Salt, ctx.SessionID),
		MachineRef:          event.Ref(ctx.Salt, ctx.MachineID),
		DevID:               ctx.DevID,
		OrgID:               ctx.OrgID,
	}
	return &ev, EventoGemini, marcas, nil
}
