// Package ingest convierte logs de herramientas en Events de frontera.
package ingest

import (
	"encoding/json"
	"time"

	"github.com/TU-USUARIO/permea-agent/internal/event"
	"github.com/TU-USUARIO/permea-agent/internal/pricing"
)

// rawRecord decodifica SOLO los campos permitidos del JSONL de Claude Code.
// Deliberadamente NO incluye message.content ni ningún campo de texto: lo que no
// se decodifica aquí, no entra en el proceso. Esa es la garantía deny-by-default.
type rawRecord struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"sessionId"`
	Cwd       string    `json:"cwd"`
	Message   struct {
		Model string `json:"model"`
		Usage struct {
			InputTokens         int `json:"input_tokens"`
			OutputTokens        int `json:"output_tokens"`
			CacheCreationTokens int `json:"cache_creation_input_tokens"`
			CacheReadTokens     int `json:"cache_read_input_tokens"`
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
}

// FromClaudeCodeLine convierte una línea JSONL en un Event de frontera.
// Devuelve (nil, nil) si la línea no es una llamada facturable.
func FromClaudeCodeLine(line []byte, ctx Context) (*event.Event, error) {
	var r rawRecord
	if err := json.Unmarshal(line, &r); err != nil {
		return nil, err
	}
	if r.Type != "assistant" || r.Message.Model == "" {
		return nil, nil
	}
	id, err := event.NewID()
	if err != nil {
		return nil, err
	}
	u := r.Message.Usage
	return &event.Event{
		SchemaVersion:       event.SchemaVersion,
		AgentVersion:        ctx.AgentVersion,
		EventID:             id,
		OccurredAt:          r.Timestamp,
		Tool:                "claude_code",
		Model:               r.Message.Model,
		TokensInput:         u.InputTokens,
		TokensOutput:        u.OutputTokens,
		TokensCacheCreation: u.CacheCreationTokens,
		TokensCacheRead:     u.CacheReadTokens,
		CostUSD:             pricing.Cost(r.Message.Model, u.InputTokens, u.OutputTokens, u.CacheCreationTokens, u.CacheReadTokens),
		ProjectRef:          event.Ref(ctx.Salt, r.Cwd),
		SessionRef:          event.Ref(ctx.Salt, r.SessionID),
		MachineRef:          event.Ref(ctx.Salt, ctx.MachineID),
		DevID:               ctx.DevID,
		OrgID:               ctx.OrgID,
	}, nil
}
