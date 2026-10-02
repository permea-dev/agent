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

// lineaCon arma una línea facturable con los identificadores dados; uno vacío se OMITE de la línea,
// no se escribe vacío, que es como llega un campo ausente del log.
func lineaCon(messageID, requestID string) []byte {
	req := ""
	if requestID != "" {
		req = `"requestId":"` + requestID + `",`
	}
	id := ""
	if messageID != "" {
		id = `"id":"` + messageID + `",`
	}
	return []byte(`{"type":"assistant","timestamp":"2026-10-02T12:00:00Z","sessionId":"s","cwd":"/x",` + req +
		`"message":{` + id + `"model":"claude-opus-4-6","usage":{"input_tokens":10,"output_tokens":5}}}`)
}

// eventIDDe devuelve el event_id que produce la línea. Que haya evento es PRECONDICIÓN (t.Fatalf).
func eventIDDe(t *testing.T, linea []byte) string {
	t.Helper()
	ev, err := FromClaudeCodeLine(linea, Context{Salt: "s"})
	if err != nil || ev == nil {
		t.Fatalf("precondición: la línea debe producir un evento (ev=%v, err=%v)", ev, err)
	}
	return ev.EventID
}

// esHex32 dice si s es exactamente la forma del contrato: 32 caracteres [0-9a-f].
func esHex32(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// (5) · P-006 FR-004, FR-006, SC-008 (b) — las tres formas (par, sólo `message.id`, sólo `requestId`)
// dan event_id estables, de 32 hex y DISTINTOS entre sí, incluido el caso en que el MISMO valor llega
// como `message.id` solo y como `requestId` solo: sólo el tipo del dominio los separa
// (`contracts/event-id.md`, §Derivación).
func TestCasoLimite_UnSoloIdentificador(t *testing.T) {
	const m, r = "msg_000000000000000000000001", "req_000000000000000000000001"
	formas := map[string][]byte{
		"par":             lineaCon(m, r),
		"solo_message_id": lineaCon(m, ""),
		"solo_request_id": lineaCon("", r),
	}

	t.Run("estable", func(t *testing.T) {
		for nombre, linea := range formas {
			if a, b := eventIDDe(t, linea), eventIDDe(t, linea); a != b {
				t.Errorf("forma %s: la misma línea dio dos event_id distintos: %s y %s", nombre, a, b)
			}
		}
	})
	t.Run("tres_formas_distintas", func(t *testing.T) {
		vistos := map[string]string{}
		for nombre, linea := range formas {
			id := eventIDDe(t, linea)
			if otra, ya := vistos[id]; ya {
				t.Errorf("las formas %s y %s dieron el mismo event_id %s", otra, nombre, id)
			}
			vistos[id] = nombre
		}
	})
	t.Run("mismo_valor_en_las_dos_formas_solas", func(t *testing.T) {
		const v = "msg_000000000000000000000001"
		if a, b := eventIDDe(t, lineaCon(v, "")), eventIDDe(t, lineaCon("", v)); a == b {
			t.Errorf("el mismo valor como message.id solo y como requestId solo dio el mismo event_id %s", a)
		}
	})
	t.Run("forma_32_hex", func(t *testing.T) {
		for nombre, linea := range formas {
			if id := eventIDDe(t, linea); !esHex32(id) {
				t.Errorf("forma %s: event_id %q no es 32 hex en minúsculas", nombre, id)
			}
		}
	})
}

// (6) · P-006 FR-002, FR-006 — los vectores normativos de una sola forma y los de la ambigüedad sin
// prefijo de longitud (`contracts/event-id.md`, §Vectores de prueba). Los de la ambigüedad se alcanzan
// por la API pública usando como `message.id`/`requestId` los dos componentes del contrato.
func TestEventID_VectoresDeUnaSolaFormaYAmbiguedad(t *testing.T) {
	casos := []struct {
		nombre, messageID, requestID, quiere string
	}{
		{"solo_message_id", "msg_000000000000000000000001", "", "1692269369e3bb2b418279566f4b093f"},
		{"solo_request_id", "", "req_000000000000000000000001", "5ab5f8575791456e5d951eb35bf36370"},
		{"ambiguedad_a_bc", "a", "bc", "6e7e350c923581ba78d5c81e0897ed53"},
		{"ambiguedad_ab_c", "ab", "c", "918d6e4cd99d09b918802196669fe50e"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := eventIDDe(t, lineaCon(c.messageID, c.requestID)); got != c.quiere {
				t.Errorf("event_id = %q, want %q", got, c.quiere)
			}
		})
	}
	t.Run("ambiguedad_distintos", func(t *testing.T) {
		if a, b := eventIDDe(t, lineaCon("a", "bc")), eventIDDe(t, lineaCon("ab", "c")); a == b {
			t.Errorf(`("a","bc") y ("ab","c") dieron el mismo event_id %s: falta el prefijo de longitud`, a)
		}
	})
}
