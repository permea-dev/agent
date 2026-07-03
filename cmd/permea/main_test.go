package main

import (
	"testing"

	"github.com/bfgnet/agente_permea/internal/config"
	"github.com/bfgnet/agente_permea/internal/ingest"
)

// TestAgentVersion_ReachesEvent verifica el cableado de T036: la versión REAL del binario
// (variable `version` de este paquete) se propaga por newIngestContext hasta
// Event.AgentVersion, sin depender del sistema de ficheros ni de la config local.
func TestAgentVersion_ReachesEvent(t *testing.T) {
	const want = "9.9.9-test"
	ictx := newIngestContext(want, config.Config{DevID: "dev-1", OrgID: "org-1"}, "salt", "machine")

	if ictx.AgentVersion != want {
		t.Fatalf("Context.AgentVersion = %q, want %q", ictx.AgentVersion, want)
	}

	// Una línea de asistente facturable -> el Event resultante debe llevar la versión.
	line := []byte(`{"type":"assistant","timestamp":"2026-06-20T10:15:30Z","sessionId":"s","cwd":"/x","message":{"model":"claude-opus-4-6","usage":{"input_tokens":10,"output_tokens":5}}}`)
	ev, err := ingest.FromClaudeCodeLine(line, ictx)
	if err != nil {
		t.Fatalf("FromClaudeCodeLine: %v", err)
	}
	if ev == nil {
		t.Fatal("se esperaba un evento facturable")
	}
	if ev.AgentVersion != want {
		t.Fatalf("Event.AgentVersion = %q, want %q", ev.AgentVersion, want)
	}
}

// TestVersion_DefaultNonEmpty: la variable version nunca es vacía por defecto (evita
// emitir agent_version="" si no se inyecta -ldflags).
func TestVersion_DefaultNonEmpty(t *testing.T) {
	if version == "" {
		t.Fatal("version por defecto no debe ser vacía")
	}
}
