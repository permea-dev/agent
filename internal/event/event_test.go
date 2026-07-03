package event

import (
	"encoding/json"
	"sort"
	"testing"
	"time"
)

// allowlist es el conjunto EXACTO de claves que pueden cruzar la frontera, según
// contracts/boundary-event.md (`additionalProperties: false`). Cualquier clave nueva en
// el Event serializado que no esté aquí es una fuga; cualquier clave que falte es una
// ruptura del contrato. Este test es el equivalente en Go a `additionalProperties: false`.
var allowlist = []string{
	"schema_version",
	"agent_version",
	"event_id",
	"occurred_at",
	"tool",
	"model",
	"tokens_input",
	"tokens_output",
	"tokens_cache_creation",
	"tokens_cache_read",
	"cost_usd",
	"cost_available",
	"project_ref",
	"session_ref",
	"machine_ref",
	"dev_id",
	"org_id",
}

// TestEvent_OnlyAllowlistKeys serializa un Event completamente poblado y comprueba que el
// conjunto de claves JSON es EXACTAMENTE la allowlist del contrato de frontera. Falla si
// aparece una clave nueva (posible passthrough de contenido) o si desaparece una esperada.
func TestEvent_OnlyAllowlistKeys(t *testing.T) {
	// Todos los campos con valor no-cero para que ningún `omitempty` (si existiera) los
	// oculte: queremos ver todas las claves que el struct puede emitir.
	ev := Event{
		SchemaVersion:       SchemaVersion,
		AgentVersion:        "0.1.0",
		EventID:             "abc123",
		OccurredAt:          time.Unix(1, 0).UTC(),
		Tool:                "claude_code",
		Model:               "claude-opus-4-6",
		TokensInput:         1,
		TokensOutput:        2,
		TokensCacheCreation: 3,
		TokensCacheRead:     4,
		CostUSD:             0.5,
		CostAvailable:       true,
		ProjectRef:          "pr",
		SessionRef:          "sr",
		MachineRef:          "mr",
		DevID:               "dev-42",
		OrgID:               "org-1",
	}

	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := map[string]bool{}
	for _, k := range allowlist {
		want[k] = true
	}

	var extra, missing []string
	for k := range got {
		if !want[k] {
			extra = append(extra, k)
		}
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			missing = append(missing, k)
		}
	}
	sort.Strings(extra)
	sort.Strings(missing)

	if len(extra) > 0 {
		t.Errorf("FUGA DE FRONTERA: claves fuera de la allowlist en el evento: %v", extra)
	}
	if len(missing) > 0 {
		t.Errorf("contrato roto: faltan claves de la allowlist en el evento: %v", missing)
	}
}
