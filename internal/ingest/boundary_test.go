package ingest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// Términos sensibles del fixture que JAMÁS deben aparecer en un evento transmitido.
var denylist = []string{
	"SECRETO_DEL_PROMPT",
	"refactoriza",
	"refactor",
	"comisiones",
	"bancarias",
	"/home/basilio",
	"acme-banca",
	"core-pagos",
	"sess-PRIVATE-9f3a",
}

// TestBoundary_NoDenylistLeaks es el test que define el producto: ninguna
// información de la denylist puede sobrevivir al paso por la frontera.
func TestBoundary_NoDenylistLeaks(t *testing.T) {
	f, err := os.Open("testdata/claude_code_sample.jsonl")
	if err != nil {
		t.Fatalf("no se pudo abrir fixture: %v", err)
	}
	defer f.Close()

	ctx := Context{Salt: "salt-local-de-prueba", MachineID: "m1", DevID: "dev-42", OrgID: "org-1", AgentVersion: "test"}

	events := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		ev, err := FromClaudeCodeLine(line, ctx)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if ev == nil {
			continue // línea no facturable, correctamente omitida
		}
		events++
		out, err := json.Marshal(ev)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		for _, bad := range denylist {
			if strings.Contains(string(out), bad) {
				t.Errorf("FUGA DE FRONTERA: el evento contiene %q\nevento: %s", bad, out)
			}
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if events != 2 {
		t.Fatalf("se esperaban 2 eventos de asistente, se generaron %d", events)
	}
}

// TestBoundary_KeepsMetrics confirma que lo permitido SÍ cruza correctamente.
func TestBoundary_KeepsMetrics(t *testing.T) {
	line := []byte(`{"type":"assistant","timestamp":"2026-06-20T10:15:30Z","sessionId":"s","cwd":"/x/y","message":{"model":"claude-opus-4-6","usage":{"input_tokens":1200,"output_tokens":800,"cache_creation_input_tokens":300,"cache_read_input_tokens":5000}}}`)
	ev, err := FromClaudeCodeLine(line, Context{Salt: "s"})
	if err != nil || ev == nil {
		t.Fatalf("evento esperado, got ev=%v err=%v", ev, err)
	}
	if ev.TokensInput != 1200 || ev.TokensOutput != 800 {
		t.Errorf("tokens mal mapeados: %+v", ev)
	}
	if ev.Tool != "claude_code" {
		t.Errorf("tool = %q", ev.Tool)
	}
	if ev.ProjectRef == "" || strings.Contains(ev.ProjectRef, "x/y") {
		t.Errorf("project_ref debe ser hash no vacío sin la ruta cruda: %q", ev.ProjectRef)
	}
	if ev.CostUSD <= 0 {
		t.Errorf("el coste debe calcularse en local: %v", ev.CostUSD)
	}
}
