package ingest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// Términos sensibles del fixture que JAMÁS deben aparecer en un evento transmitido.
// Cubre contenido conocido (message.content, rutas, sesión) y —lo más exigente del
// SC-005— campos NUEVOS y DESCONOCIDOS con contenido inyectados en el origen: un campo
// futuro arbitrario, argumentos de herramienta y un campo de nivel superior no previsto.
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
	// Campos desconocidos/futuros con contenido (deny-by-default por construcción):
	"FUGA_CAMPO_FUTURO_ZzZ", // campo futuro arbitrario dentro de message
	"FUGA_ARG_SECRETA",      // argumentos de una llamada a herramienta
	"FUGA_TOP_LEVEL_QqQ",    // campo desconocido de nivel superior
	"sk-LEAK-DEADBEEF",      // secreto tipo clave de API
	"/etc/passwd",           // comando de herramienta
	"id_rsa",                // ruta sensible en input de herramienta

	// ═══ P-004 T004 · LOS DOS CENTINELAS DE RUTA, Y POR QUÉ SON DOS ══════════════════════
	//
	// Vienen de `testdata/boundary_sample.jsonl`, cuyo cwd es
	// `/tmp/raiz-canaria-9x7/subdir-canario-3z5`: el primer fragmento identifica la RAÍZ del
	// proyecto y el segundo el SUBDIRECTORIO desde el que se lanzó.
	//
	// Separarlos NO es redundancia. Hoy solo circula por el proceso el cwd completo, así que
	// un único centinela bastaría. Pero cuando T014 traiga la resolución, **la ruta de la raíz
	// pasará a ser un valor nuevo** —derivado, distinto del cwd— y podrá filtrarse SOLA: una
	// fuga de la raíz contendría `raiz-canaria-9x7` y NO `subdir-canario-3z5`. Con un solo
	// centinela sobre la ruta completa, esa fuga viajaría de polizón sin que nadie la viera.
	// Cada valor que cruza el proceso necesita su propio testigo.
	"raiz-canaria-9x7",   // fragmento de la RAÍZ del proyecto (valor que T014 hará circular)
	"subdir-canario-3z5", // fragmento del SUBDIRECTORIO de lanzamiento
	"sess-CANARIA-7k2",   // identificador de sesión del fixture dedicado
	"CANARIO_PROMPT_QqQ", // contenido de prompt del fixture dedicado
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

// ═══ P-004 T005 · LA FRONTERA SE COMPRUEBA EN LOS TRES CAMINOS HACIA EL EXTERIOR ═══════════
//
// FR-017, tras D-004-4, acota la garantía a la salida **hacia el exterior** y la nombra entera:
// el **evento serializado**, la **cola de envío** y el **transporte**. Comprobar solo el primero
// dejaba dos caminos sin testigo, y no son teóricos: por la cola pasa cada evento antes de
// transmitirse, y el cuerpo HTTP es literalmente lo que sale de la máquina.
//
// LOS TRES SE COMPRUEBAN POR SEPARADO, cada uno sobre SU artefacto —el JSON del evento, el
// contenido de `queue.jsonl` leído del disco, y el cuerpo recibido por el servidor—, y no
// derivando dos de ellos del primero. Si el camino (2) se comprobara sobre lo que el test cree
// que escribió en vez de sobre lo que hay en el fichero, una fuga introducida **en el encolado**
// pasaría inadvertida; lo mismo con (3) y el cuerpo real de la petición. La demostración de que
// cada camino tiene testigo propio son las tres mutaciones independientes que lo validan.
func TestBoundary_TresCaminosHaciaElExterior(t *testing.T) {
	dataDir := testutil.Sandbox(t)

	eventos := eventosDelFixture(t, "testdata/boundary_sample.jsonl")
	if len(eventos) != 2 {
		t.Fatalf("el fixture dedicado debe producir 2 eventos facturables, produjo %d", len(eventos))
	}

	// ── Camino 1 · el evento serializado ────────────────────────────────────────────────
	for _, ev := range eventos {
		out, err := json.Marshal(ev)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		exigirSinFugas(t, "camino 1 · evento serializado", string(out))
	}

	// ── Camino 2 · la cola durable, LEÍDA DEL DISCO ─────────────────────────────────────
	for _, ev := range eventos {
		if err := transport.Append(dataDir, ev); err != nil {
			t.Fatalf("encolar: %v", err)
		}
	}
	cola, err := os.ReadFile(transport.QueuePath(dataDir))
	if err != nil {
		t.Fatalf("leer la cola: %v", err)
	}
	exigirSinFugas(t, "camino 2 · queue.jsonl", string(cola))

	// ── Camino 3 · el cuerpo que recibe el backend ──────────────────────────────────────
	// httptest.NewTLSServer, nunca http en claro (contracts/transport.md:58). El cuerpo se
	// captura en el servidor: es lo que de verdad salió de la máquina, no lo que el cliente
	// pensaba enviar.
	var recibido []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		recibido = append(recibido, b...)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cliente := transport.New(srv.URL, "token-de-prueba")
	cliente.HTTP = srv.Client() // confía en el certificado del servidor de test
	if err := cliente.Send(eventos); err != nil {
		t.Fatalf("enviar: %v", err)
	}
	if len(recibido) == 0 {
		t.Fatal("el servidor no recibió cuerpo: el camino 3 no se habría comprobado")
	}
	exigirSinFugas(t, "camino 3 · cuerpo transmitido", string(recibido))
}

// eventosDelFixture parsea un JSONL de Claude Code y devuelve los eventos facturables.
func eventosDelFixture(t *testing.T, ruta string) []event.Event {
	t.Helper()

	f, err := os.Open(ruta)
	if err != nil {
		t.Fatalf("no se pudo abrir el fixture %q: %v", ruta, err)
	}
	defer func() { _ = f.Close() }()

	ctx := Context{Salt: "salt-local-de-prueba", MachineID: "m1", DevID: "dev-42", OrgID: "org-1", AgentVersion: "test"}

	var eventos []event.Event
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
			continue
		}
		eventos = append(eventos, *ev)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return eventos
}

// exigirSinFugas contrasta un artefacto de salida contra la denylist completa. El nombre del
// camino viaja en el mensaje: cuando algo falla, lo primero que hay que saber es POR DÓNDE se
// escapó, no solo qué.
func exigirSinFugas(t *testing.T, camino, contenido string) {
	t.Helper()

	for _, bad := range denylist {
		if strings.Contains(contenido, bad) {
			t.Errorf("FUGA DE FRONTERA [%s]: contiene %q\ncontenido: %s", camino, bad, contenido)
		}
	}
}

// TestBoundary_UnknownFutureFieldDoesNotLeak es el caso más exigente del SC-005: una
// futura versión de Claude Code añade un campo DESCONOCIDO con contenido (aquí un campo
// arbitrario que no existe en rawRecord) directamente en el registro. El evento
// serializado NO debe contener ese contenido: deny-by-default por construcción, porque el
// campo no tiene lugar en el struct cerrado ni se decodifica. Este test FALLA si alguien
// amplía rawRecord/Event para dar paso a contenido.
func TestBoundary_UnknownFutureFieldDoesNotLeak(t *testing.T) {
	// Registro de asistente facturable con un campo futuro inédito que transporta
	// contenido sensible, además de message.content y argumentos de herramienta.
	line := []byte(`{"type":"assistant","timestamp":"2026-06-20T10:15:30Z","sessionId":"sess-PRIVATE-9f3a","cwd":"/home/basilio/x","message":{"model":"claude-opus-4-6","usage":{"input_tokens":10,"output_tokens":5},"content":[{"type":"text","text":"LEAK_CONTENT_AAA"}],"brand_new_2027_field":"LEAK_UNKNOWN_BBB"},"another_unknown":"LEAK_TOPLEVEL_CCC"}`)

	ev, err := FromClaudeCodeLine(line, Context{Salt: "s", MachineID: "m", DevID: "d", OrgID: "o", AgentVersion: "t"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ev == nil {
		t.Fatal("se esperaba un evento facturable")
	}
	out, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, leak := range []string{"LEAK_CONTENT_AAA", "LEAK_UNKNOWN_BBB", "LEAK_TOPLEVEL_CCC", "brand_new_2027_field", "another_unknown"} {
		if strings.Contains(string(out), leak) {
			t.Errorf("FUGA DE FRONTERA: un campo desconocido con contenido cruzó la frontera (%q)\nevento: %s", leak, out)
		}
	}
	// Sanidad: las métricas permitidas sí cruzan (el registro no se descartó por completo).
	if ev.TokensInput != 10 || ev.TokensOutput != 5 {
		t.Errorf("las métricas permitidas deben cruzar: %+v", ev)
	}
}

// TestBoundary_CostAvailable verifica el metadato cost_available (R5): el evento
// serializado incluye el campo y distingue "coste no disponible" de "coste 0".
func TestBoundary_CostAvailable(t *testing.T) {
	// Modelo conocido -> cost_available=true, coste > 0.
	known := []byte(`{"type":"assistant","timestamp":"2026-06-20T10:15:30Z","sessionId":"s","cwd":"/x","message":{"model":"claude-opus-4-6","usage":{"input_tokens":1000,"output_tokens":1000}}}`)
	ev, err := FromClaudeCodeLine(known, Context{Salt: "s"})
	if err != nil || ev == nil {
		t.Fatalf("evento esperado: ev=%v err=%v", ev, err)
	}
	if !ev.CostAvailable {
		t.Errorf("modelo conocido debe tener cost_available=true")
	}
	out, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), "cost_available") {
		t.Errorf("el evento serializado debe incluir cost_available: %s", out)
	}

	// Modelo desconocido -> cost_available=false, cost_usd=0, tokens contabilizados.
	unknown := []byte(`{"type":"assistant","timestamp":"2026-06-20T10:15:30Z","sessionId":"s","cwd":"/x","message":{"model":"modelo-futuro-x","usage":{"input_tokens":500,"output_tokens":200}}}`)
	ev2, err := FromClaudeCodeLine(unknown, Context{Salt: "s"})
	if err != nil || ev2 == nil {
		t.Fatalf("evento esperado: ev=%v err=%v", ev2, err)
	}
	if ev2.CostAvailable {
		t.Errorf("modelo desconocido debe tener cost_available=false")
	}
	if ev2.CostUSD != 0 {
		t.Errorf("modelo desconocido debe tener cost_usd=0, got %v", ev2.CostUSD)
	}
	if ev2.TokensInput != 500 || ev2.TokensOutput != 200 {
		t.Errorf("tokens deben contabilizarse aun sin coste: %+v", ev2)
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
