package ingest

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/permea-dev/agent/internal/event"
)

// ═══ P-011 · UNA LÍNEA DE COPILOT CLI ES, COMO MUCHO, UN EVENTO ═══════════════════════════════════════════════════════
//
// Copilot CLI escribe un JSONL de sólo añadir por sesión (descubrimiento Q1). Son REGISTRO (D-2) sólo dos clases de línea:
//   - `session.usage_record`: una por llamada al modelo principal, también las de un subagente (muestra 3), con el modelo
//     cobrado, las cuatro partidas y su `usageId` (descubrimiento Q3);
//   - `session.compaction_complete` con `success: true`: el consumo de una compactación, que no está en ningún
//     `usage_record` (muestra 2, M2).
//
// Nada más: ni `shutdown`, ni `usage_checkpoint`, ni `model.*` (las llamadas auxiliares, a 0 créditos, son un límite
// declarado), ni ninguna línea con contenido.
//
// ═══ LO QUE SE LEE, Y LO QUE NO ════════════════════════════════════════════════════════════════════════════════════
//
// PREFILTRO DE BYTES (P-11): una línea sólo se decodifica si contiene el marcador de su tipo. Las líneas con contenido
// (prompts, respuestas, el prompt de sistema, 33–56 KB) no se decodifican nunca, y un marcador dentro de un texto va
// escapado (`\"`) y no casa.
//
// De un registro se leen sólo `type`, `timestamp`, y de `data`:
//   - de `usage_record`: `usage.{model, inputTokens, outputTokens, cacheReadTokens, cacheWriteTokens, accounting.usageId}`;
//   - de `compaction_complete`: `{success, requestId, compactionTokensUsed.{model, las cuatro partidas}}` (FR-012).
//
// Nunca `reasoningTokens` *(el razonamiento ya va dentro de `outputTokens`: descubrimiento Q4)*, ni `toolCounts`,
// `summaryContent` o `checkpointPath`. Los ids sólo entran en el hash del `event_id`; la sesión y el `cwd`, con sal
// (FR-019). El coste no se calcula: lo pone la plataforma (D-4, constitución 1.1.0 §II); no se consulta
// `internal/pricing`.

// ContextoCopilot son los datos locales que añade el agente a un registro de Copilot CLI. Reutiliza el `Context` de Claude
// Code por sus datos locales y su resolutor. `SessionID` y `Cwd` son los del tramo, y los pone quien lee el fichero (P-3, B3).
type ContextoCopilot struct {
	Context
	// SessionID es el `data.sessionId` del `session.start` del fichero. Vacío si no hay `start`.
	SessionID string
	// Cwd es el `data.context.cwd` del último `session.start` o `session.resume` anterior. Vacío si no hay.
	Cwd string
	// Decodificadas, si no es nil, cuenta las líneas que pasan el prefiltro y se decodifican (P-11; test (13)).
	Decodificadas *int
	// ConCreditos lee `copilotUsage.totalNanoAiu` en las marcas. Sólo lo pide `--scan`, para su línea de control (P-9, D-4).
	ConCreditos bool
}

// ClaseCopilot clasifica una línea de una sesión de Copilot CLI (P-011 FR-020).
type ClaseCopilot int

const (
	// NoEsRegistroCopilot es cualquier línea que no es `usage_record` ni `compaction_complete` con `success: true`.
	NoEsRegistroCopilot ClaseCopilot = iota
	// EventoCopilot es un registro que produce un evento.
	EventoCopilot
	// SinIdentificadorCopilot es un registro sin `usageId` (o `requestId`) textual no vacío (FR-014).
	SinIdentificadorCopilot
	// IncoherenteCopilot es un registro cuyas partidas o cuyo momento no se pueden usar (FR-020).
	IncoherenteCopilot
)

// MarcasCopilot son las marcas de un evento (FR-020, D-011-P4): subconjuntos de «eventos», no clases.
type MarcasCopilot struct {
	// Compactacion: el evento es de una `compaction_complete`.
	Compactacion bool
	// SinModelo: el registro no trae un `model` textual no vacío (P-5).
	SinModelo bool
	// NanoAIU son los créditos del registro según Copilot, con `ConCreditos`; 0 si no los trae enteros. Nunca van al evento.
	NanoAIU int64
}

// Los tipos de las dos clases de registro, y sus marcadores para el prefiltro de bytes (P-11).
const (
	tipoUsageRecordCopilot  = "session.usage_record"
	tipoCompactionCopilot   = "session.compaction_complete"
	marcadorUsageRecord     = `"type":"` + tipoUsageRecordCopilot + `"`
	marcadorCompactionLinea = `"type":"` + tipoCompactionCopilot + `"`
)

// lineaCopilot es la envoltura de una línea, toda en crudo: un campo con otra forma clasifica, nunca rompe.
type lineaCopilot struct {
	Type      json.RawMessage `json:"type"`
	Timestamp json.RawMessage `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// RegistroCopilot lee una línea de una sesión de Copilot CLI y devuelve su clase y, si es un evento, el evento y sus marcas.
//
// Un error es SÓLO una línea que pasó el prefiltro y cuya envoltura no es un objeto JSON: quien lee decide qué hacer con
// ella. Lo demás clasifica, en este orden (D-011-P4): no es registro → sin identificador → incoherente → evento. Un
// `usage` que no es un objeto no deja leer el `usageId`, y es «sin identificador».
func RegistroCopilot(linea []byte, ctx ContextoCopilot) (*event.Event, ClaseCopilot, MarcasCopilot, error) {
	if !bytes.Contains(linea, []byte(marcadorUsageRecord)) && !bytes.Contains(linea, []byte(marcadorCompactionLinea)) {
		return nil, NoEsRegistroCopilot, MarcasCopilot{}, nil
	}
	if ctx.Decodificadas != nil {
		*ctx.Decodificadas++
	}
	var l lineaCopilot
	if err := json.Unmarshal(linea, &l); err != nil {
		return nil, NoEsRegistroCopilot, MarcasCopilot{}, err
	}

	var tipo, id string
	var partidas map[string]json.RawMessage
	var marcas MarcasCopilot
	switch textoDe(l.Type) {
	case tipoUsageRecordCopilot:
		var d struct {
			Usage json.RawMessage `json:"usage"`
		}
		_ = json.Unmarshal(l.Data, &d) // `data` que no es un objeto deja `usage` vacío
		var acc struct {
			Accounting struct {
				UsageID json.RawMessage `json:"usageId"`
			} `json:"accounting"`
		}
		if objetoDe(d.Usage, &partidas) {
			_ = json.Unmarshal(d.Usage, &acc)
		}
		tipo, id = tipoRespuesta, textoDe(acc.Accounting.UsageID)
	case tipoCompactionCopilot:
		var d map[string]json.RawMessage
		_ = json.Unmarshal(l.Data, &d)
		var exito bool
		if json.Unmarshal(d["success"], &exito) != nil || !exito {
			return nil, NoEsRegistroCopilot, MarcasCopilot{}, nil
		}
		objetoDe(d["compactionTokensUsed"], &partidas)
		tipo, id, marcas.Compactacion = tipoCompactacionCopilot, textoDe(d["requestId"]), true
	default:
		return nil, NoEsRegistroCopilot, MarcasCopilot{}, nil
	}

	eventID, ok := derivarEventIDCopilot(tipo, id)
	if !ok {
		return nil, SinIdentificadorCopilot, MarcasCopilot{}, nil
	}
	p, ok := partidasCopilotDe(partidas)
	var momento time.Time
	if !ok || json.Unmarshal(l.Timestamp, &momento) != nil {
		return nil, IncoherenteCopilot, MarcasCopilot{}, nil
	}

	modelo := textoDe(partidas["model"])
	marcas.SinModelo = modelo == ""
	if ctx.ConCreditos {
		marcas.NanoAIU = creditosDe(partidas["copilotUsage"])
	}
	proyecto := ""
	if ctx.Cwd != "" {
		proyecto = ctx.Resolutor.Derivar(ctx.Cwd, ctx.Salt)
	}
	ev := event.Event{
		SchemaVersion:       event.SchemaVersion,
		AgentVersion:        ctx.AgentVersion,
		EventID:             eventID,
		OccurredAt:          momento.UTC(),
		Tool:                herramientaCopilot,
		Model:               modelo,
		TokensInput:         int(p.entrada - p.lectura - p.escritura),
		TokensOutput:        int(p.salida),
		TokensCacheCreation: int(p.escritura),
		TokensCacheRead:     int(p.lectura),
		CostUSD:             0,
		CostAvailable:       false,
		ProjectRef:          proyecto,
		SessionRef:          event.Ref(ctx.Salt, ctx.SessionID),
		MachineRef:          event.Ref(ctx.Salt, ctx.MachineID),
		DevID:               ctx.DevID,
		OrgID:               ctx.OrgID,
	}
	return &ev, EventoCopilot, marcas, nil
}

// creditosDe es el `totalNanoAiu` de un `copilotUsage`, o 0 si no es un entero no negativo.
func creditosDe(crudo json.RawMessage) int64 {
	var u struct {
		TotalNanoAiu int64 `json:"totalNanoAiu"`
	}
	if json.Unmarshal(crudo, &u) != nil || u.TotalNanoAiu < 0 {
		return 0
	}
	return u.TotalNanoAiu
}

// objetoDe decodifica `crudo` en `m` si es un objeto JSON, y dice si lo era.
func objetoDe(crudo json.RawMessage, m *map[string]json.RawMessage) bool {
	*m = nil
	return len(crudo) > 0 && crudo[0] == '{' && json.Unmarshal(crudo, m) == nil && *m != nil
}

// partidasCopilot son las cuatro partidas de D-3, tal como las escribe Copilot CLI.
type partidasCopilot struct {
	entrada, salida, lectura, escritura int64
}

// partidasCopilotDe lee las cuatro partidas de `usage` o de `compactionTokensUsed`. `ok == false` si no es un objeto, si
// faltan las cuatro, si alguna no es un entero o es negativa, o si la lectura y la escritura no caben en la entrada, que
// las incluye (descubrimiento Q4; D-3: la resta sería negativa). Una partida ausente o `null` vale 0.
func partidasCopilotDe(m map[string]json.RawMessage) (partidasCopilot, bool) {
	var p partidasCopilot
	if m == nil {
		return p, false
	}
	presentes := 0
	for clave, destino := range map[string]*int64{
		"inputTokens": &p.entrada, "outputTokens": &p.salida, "cacheReadTokens": &p.lectura, "cacheWriteTokens": &p.escritura,
	} {
		crudo, esta := m[clave]
		if !esta || ausente(crudo) {
			continue
		}
		presentes++
		if json.Unmarshal(crudo, destino) != nil || *destino < 0 {
			return p, false
		}
	}
	if presentes == 0 || p.lectura+p.escritura > p.entrada {
		return p, false
	}
	return p, true
}
