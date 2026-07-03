package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/bfgnet/agente_permea/internal/event"
)

// recorder es un backend de ingesta simulado: deduplica por event_id como exige el
// contrato de transporte (exactamente una vez). Cuenta también las peticiones para
// verificar que un reenvío realmente ocurre.
type recorder struct {
	mu       sync.Mutex
	seen     map[string]bool
	requests int
}

func (r *recorder) unique() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.seen)
}

func (r *recorder) reqCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.requests
}

// newBackend levanta un backend HTTPS de test (httptest.NewTLSServer — nunca http en
// claro) que responde con el código devuelto por code(reqNum) y, en 2xx, memoriza los
// event_id recibidos deduplicando. Devuelve un Client ya confiado en el certificado del
// servidor y con backoff sin esperas (sleep no-op) para que los tests sean rápidos.
func newBackend(t *testing.T, code func(reqNum int) int) (*Client, *recorder) {
	t.Helper()
	rec := &recorder{seen: map[string]bool{}}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.mu.Lock()
		rec.requests++
		reqNum := rec.requests
		rec.mu.Unlock()

		status := http.StatusOK
		if code != nil {
			status = code(reqNum)
		}
		if status >= 200 && status < 300 {
			body, _ := io.ReadAll(r.Body)
			var evs []event.Event
			_ = json.Unmarshal(body, &evs)
			rec.mu.Lock()
			for _, e := range evs {
				rec.seen[e.EventID] = true
			}
			rec.mu.Unlock()
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)

	client := New(srv.URL, "test-token")
	client.HTTP = srv.Client()            // confía en el certificado autofirmado del test
	client.sleep = func(time.Duration) {} // no dormir en los reintentos
	return client, rec
}

// seed encola n eventos en dir y devuelve el orden esperado (base de los tests de drenaje).
func seed(t *testing.T, dir string, n int) []event.Event {
	t.Helper()
	var evs []event.Event
	for i := 0; i < n; i++ {
		e := event.Event{
			SchemaVersion: 1,
			EventID:       fmt.Sprintf("id-%d", i),
			Tool:          "claude_code",
			Model:         "m",
			OccurredAt:    time.Unix(0, 0).UTC(),
		}
		if err := Append(dir, e); err != nil {
			t.Fatalf("Append: %v", err)
		}
		evs = append(evs, e)
	}
	return evs
}

// TestQueue_OfflineThenDrain (US2): sin red la cola crece y no se pierde nada; al volver
// la red el backend recibe exactamente N y la cola queda vacía (V4).
func TestQueue_OfflineThenDrain(t *testing.T) {
	dir := t.TempDir()
	evs := seed(t, dir, 3)

	// Sin red: endpoint https muerto -> el drenaje agota reintentos y falla; la cola
	// permanece intacta (durabilidad offline, FR-007).
	dead := New("https://127.0.0.1:1", "tok")
	dead.sleep = func(time.Duration) {}
	dead.MaxRetries = 2
	if _, err := drain(dir, dead, evs); err == nil {
		t.Fatal("esperaba error de red sin backend disponible")
	}
	if got, _ := Load(dir); len(got) != 3 {
		t.Fatalf("sin red la cola debe permanecer intacta: got %d", len(got))
	}

	// Vuelve la red: el drenaje entrega exactamente N y vacía la cola.
	client, rec := newBackend(t, nil)
	pending, _ := Load(dir)
	n, err := drain(dir, client, pending)
	if err != nil {
		t.Fatalf("drain: %v", err)
	}
	if n != 3 || rec.unique() != 3 {
		t.Fatalf("esperaba 3 confirmados/recibidos, got n=%d unique=%d", n, rec.unique())
	}
	if got, _ := Load(dir); len(got) != 0 {
		t.Fatalf("tras drenar, la cola debe quedar vacía: got %d", len(got))
	}
}

// TestQueue_ExactlyOnce (US2, SC-004): reenviar un lote ya aceptado no duplica gracias a
// la deduplicación por event_id en el backend (at-least-once cliente + dedup servidor).
func TestQueue_ExactlyOnce(t *testing.T) {
	dir := t.TempDir()
	evs := seed(t, dir, 4)
	client, rec := newBackend(t, nil)

	// Primer envío: el backend registra 4 event_id únicos.
	if _, err := drain(dir, client, evs); err != nil {
		t.Fatalf("drain 1: %v", err)
	}
	// Reenvío del MISMO lote (simula una respuesta 2xx perdida): dedup por event_id.
	if _, err := drain(dir, client, evs); err != nil {
		t.Fatalf("drain 2: %v", err)
	}
	if rec.unique() != 4 {
		t.Fatalf("dedup por event_id: esperaba 4 únicos, got %d", rec.unique())
	}
	if rec.reqCount() < 2 {
		t.Fatalf("esperaba al menos 2 peticiones (reenvío efectivo), got %d", rec.reqCount())
	}
}

// TestSend_RejectsHTTP (US2): un endpoint http:// es un error de configuración y el
// cliente lo rechaza antes de transmitir nada en claro (FR-009).
func TestSend_RejectsHTTP(t *testing.T) {
	c := New("http://insecure.example/ingest", "tok")
	err := c.Send([]event.Event{{EventID: "x"}})
	if !errors.Is(err, ErrScheme) {
		t.Fatalf("esperaba rechazo de http://, got %v", err)
	}
}

// TestSend_StatusSemantics (US2): 2xx confirma; 401/403 detiene el sync (no reintenta);
// 5xx y error de red reintentan; otros 4xx ni reintentan ni son auth (contracts/transport.md).
func TestSend_StatusSemantics(t *testing.T) {
	ev := []event.Event{{EventID: "x"}}

	ok, _ := newBackend(t, func(int) int { return http.StatusOK })
	if err := ok.Send(ev); err != nil {
		t.Fatalf("2xx no debe ser error: %v", err)
	}

	unauth, _ := newBackend(t, func(int) int { return http.StatusUnauthorized })
	err := unauth.Send(ev)
	if !IsAuth(err) {
		t.Fatalf("401 debe marcar auth, got %v", err)
	}
	if Retryable(err) {
		t.Fatal("401 no debe ser reintentable")
	}

	forbid, _ := newBackend(t, func(int) int { return http.StatusForbidden })
	if !IsAuth(forbid.Send(ev)) {
		t.Fatal("403 debe marcar auth")
	}

	fail, _ := newBackend(t, func(int) int { return http.StatusServiceUnavailable })
	if !Retryable(fail.Send(ev)) {
		t.Fatal("5xx debe ser reintentable")
	}

	dead := New("https://127.0.0.1:1", "tok")
	if !Retryable(dead.Send(ev)) {
		t.Fatal("error de red debe ser reintentable")
	}

	bad, _ := newBackend(t, func(int) int { return http.StatusBadRequest })
	err = bad.Send(ev)
	if err == nil {
		t.Fatal("un 4xx malformado debe ser error")
	}
	if Retryable(err) || IsAuth(err) {
		t.Fatalf("4xx malformado no reintenta ni es auth, got %v", err)
	}
}
