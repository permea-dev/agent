package ingest

import "testing"

// ═══ P-006 B1 · LOS TESTIGOS DE LA IDENTIDAD DEL EVENTO ════════════════════════════════════
//
// Fijan la conducta que `specs/006-medicion-fiel/contracts/event-id.md` exige y que el lector todavía
// NO tiene: el `event_id` deja de ser aleatorio y pasa a derivarse, sin sal, de `message.id` y
// `requestId`; las líneas `<synthetic>` y las que no traen ninguno de los dos no se emiten.
//
// ═══ POR QUÉ SE ESCRIBEN CONTRA `FromClaudeCodeLine` Y NO CONTRA LA DERIVACIÓN ══════════════
//
// La función de derivación no existe todavía. Un test contra un símbolo inexistente no compila, y un
// `[build failed]` no es un rojo legible: falla igual con un test correcto que con uno vacío
// (disciplina 3 de P-005). Por la API que ya existe, los cuatro caen POR CONDUCTA, y cada uno por su
// razón.
//
// ═══ LOS IDENTIFICADORES SON SINTÉTICOS ════════════════════════════════════════════════════
//
// Ninguno procede de un log real (disciplina 9 de P-006). El par de `lineaDelVectorPar` es el del
// vector normativo del contrato, calculado aparte con una implementación independiente.

// lineaDelVectorPar es una línea facturable con el par sintético del primer vector del contrato.
const lineaDelVectorPar = `{"type":"assistant","timestamp":"2026-10-02T12:00:00Z","sessionId":"s-vector","cwd":"/x","requestId":"req_000000000000000000000001","message":{"id":"msg_000000000000000000000001","model":"claude-opus-4-6","usage":{"input_tokens":10,"output_tokens":5}}}`

// vectorDelPar es el `event_id` normativo de ese par (`contracts/event-id.md`, §Vectores de prueba).
const vectorDelPar = "43b8b3b6446704ae3cb8bb74683bf3e2"

// (1) · P-006 FR-002, FR-008 — la misma línea, leída por dos instalaciones distintas, da el MISMO
// `event_id`. Las dos difieren en todo lo local: sal, máquina, desarrollador y organización.
func TestEventID_LaMismaLineaDaElMismoIDEnDosInstalaciones(t *testing.T) {
	ctxA := Context{Salt: "sal-de-la-instalacion-A", MachineID: "maquina-A", DevID: "dev-A", OrgID: "org-A", AgentVersion: "t"}
	ctxB := Context{Salt: "sal-de-la-instalacion-B", MachineID: "maquina-B", DevID: "dev-B", OrgID: "org-B", AgentVersion: "t"}

	evA, err := FromClaudeCodeLine([]byte(lineaDelVectorPar), ctxA)
	if err != nil || evA == nil {
		t.Fatalf("precondición: la línea del vector debe producir un evento (ev=%v, err=%v)", evA, err)
	}
	evB, err := FromClaudeCodeLine([]byte(lineaDelVectorPar), ctxB)
	if err != nil || evB == nil {
		t.Fatalf("precondición: la línea del vector debe producir un evento (ev=%v, err=%v)", evB, err)
	}

	if evA.EventID != evB.EventID {
		t.Errorf("P-006 FR-002: la misma línea dio dos event_id distintos en dos instalaciones:\n  A: %s\n  B: %s",
			evA.EventID, evB.EventID)
	}
}

// (2) · P-006 FR-002, FR-004 — el par sintético da EXACTAMENTE el vector del contrato. No basta con que
// sea estable: tiene que ser ESTA derivación.
func TestEventID_VectorDelPar(t *testing.T) {
	ev, err := FromClaudeCodeLine([]byte(lineaDelVectorPar), Context{Salt: "cualquier-sal"})
	if err != nil || ev == nil {
		t.Fatalf("precondición: la línea del vector debe producir un evento (ev=%v, err=%v)", ev, err)
	}
	if ev.EventID != vectorDelPar {
		t.Errorf("P-006 FR-002: event_id del par = %q, want %q (contracts/event-id.md, §Vectores de prueba)",
			ev.EventID, vectorDelPar)
	}
}

// (3) · P-006 FR-007 — una línea `<synthetic>` NO produce evento, aunque traiga identificadores.
func TestSintetica_NoSeEmite(t *testing.T) {
	linea := `{"type":"assistant","timestamp":"2026-10-02T12:00:00Z","sessionId":"s","cwd":"/x","requestId":"req_SINTETICA0000000000000001","message":{"id":"msg_SINTETICA0000000000000001","model":"<synthetic>","usage":{"input_tokens":0,"output_tokens":0}}}`
	ev, err := FromClaudeCodeLine([]byte(linea), Context{Salt: "s"})
	if err != nil {
		t.Fatalf("precondición: la línea debe decodificarse: %v", err)
	}
	if ev != nil {
		t.Errorf("P-006 FR-007: una línea <synthetic> NO debe producir evento; se produjo uno de modelo %q", ev.Model)
	}
}

// (4) · P-006 FR-006 — una línea facturable SIN `message.id` NI `requestId` no produce evento: no queda
// nada estable de lo que derivar, y un identificador aleatorio reintroduciría el defecto que P-006
// corrige. Su recuento como «no contable» lo prueba la pasada (P-006 B2b).
func TestCasoLimite_SinIdentificadorNoSeEmite(t *testing.T) {
	linea := `{"type":"assistant","timestamp":"2026-10-02T12:00:00Z","sessionId":"s","cwd":"/x","message":{"model":"claude-opus-4-6","usage":{"input_tokens":10,"output_tokens":5}}}`
	ev, err := FromClaudeCodeLine([]byte(linea), Context{Salt: "s"})
	if err != nil {
		t.Fatalf("precondición: la línea debe decodificarse: %v", err)
	}
	if ev != nil {
		t.Errorf("P-006 FR-006: una línea sin message.id ni requestId NO debe producir evento; se produjo event_id=%q", ev.EventID)
	}
}
