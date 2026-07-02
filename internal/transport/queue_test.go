package transport

import (
	"testing"
	"time"

	"github.com/bfgnet/agente_permea/internal/event"
)

// TestQueue_Append: la cola es append-only, una línea JSON por evento, y Load las
// recupera en orden (base de la durabilidad offline, US1 -> US2).
func TestQueue_Append(t *testing.T) {
	dir := t.TempDir()
	evs := []event.Event{
		{SchemaVersion: 1, EventID: "id-1", Tool: "claude_code", Model: "m", OccurredAt: time.Unix(0, 0).UTC()},
		{SchemaVersion: 1, EventID: "id-2", Tool: "claude_code", Model: "m", OccurredAt: time.Unix(0, 0).UTC()},
	}
	for _, e := range evs {
		if err := Append(dir, e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("esperaba 2 eventos en cola, got %d", len(got))
	}
	if got[0].EventID != "id-1" || got[1].EventID != "id-2" {
		t.Fatalf("orden de cola incorrecto: %+v", got)
	}
}

// TestQueue_LoadEmpty: una cola inexistente no es error (arranque limpio / sin uso).
func TestQueue_LoadEmpty(t *testing.T) {
	got, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("cola inexistente no debe fallar: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("esperaba 0, got %d", len(got))
	}
}
