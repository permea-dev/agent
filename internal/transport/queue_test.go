package transport

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/event"
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

// TestQueue_AtomicRewrite_KeepsUnconfirmed (US2, T028): tras confirmar (2xx) solo un
// subconjunto, queue.jsonl conserva EXACTAMENTE los no confirmados y se reescribe vía
// temporal + os.Rename (sin borrado in-place ni temporales residuales).
func TestQueue_AtomicRewrite_KeepsUnconfirmed(t *testing.T) {
	dir := t.TempDir()
	evs := seed(t, dir, 4)

	// El backend acepta (2xx) el primer lote y falla (5xx) el segundo: solo los dos
	// primeros eventos se confirman.
	client, _ := newBackend(t, func(reqNum int) int {
		if reqNum == 1 {
			return http.StatusOK
		}
		return http.StatusServiceUnavailable
	})
	client.BatchSize = 2
	client.MaxRetries = 1

	confirmed, err := drain(dir, client, evs)
	if err == nil {
		t.Fatal("esperaba error del segundo lote (5xx)")
	}
	if confirmed != 2 {
		t.Fatalf("esperaba 2 confirmados, got %d", confirmed)
	}

	// La cola conserva SOLO los no confirmados, en orden.
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 2 || got[0].EventID != "id-2" || got[1].EventID != "id-3" {
		t.Fatalf("la cola debe conservar los no confirmados en orden, got %+v", got)
	}

	// La reescritura atómica no deja temporales residuales en el directorio.
	tmps, _ := filepath.Glob(filepath.Join(dir, ".tmp-*"))
	if len(tmps) != 0 {
		t.Fatalf("la reescritura atómica no debe dejar temporales: %v", tmps)
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
