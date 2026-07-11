package transport

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestVerifyEmptyBatch asevera el contrato del ping de verificación de enrolamiento:
// el cuerpo transmitido DEBE ser exactamente `[]` (lote vacío, no `null`), con
// `Authorization: Bearer`, y la clasificación de respuestas 2xx/401/5xx del contrato
// de transporte. Reutiliza Send([]event.Event{}); no hay endpoint nuevo.
func TestVerifyEmptyBatch(t *testing.T) {
	t.Run("cuerpo exactamente [] con Bearer y 2xx=nil", func(t *testing.T) {
		var gotBody string
		var gotAuth string
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			gotAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := New(srv.URL, "dev_tok_demo")
		c.HTTP = srv.Client()

		if err := c.Verify(); err != nil {
			t.Fatalf("2xx debe devolver nil, got: %v", err)
		}
		if gotBody != "[]" {
			t.Errorf("el ping DEBE enviar el lote vacío literal `[]` (no `null`); got %q", gotBody)
		}
		if gotAuth != "Bearer dev_tok_demo" {
			t.Errorf("Authorization = %q, quiero %q", gotAuth, "Bearer dev_tok_demo")
		}
	})

	t.Run("401 => IsAuth", func(t *testing.T) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		c := New(srv.URL, "tok")
		c.HTTP = srv.Client()

		err := c.Verify()
		if err == nil {
			t.Fatal("401 debe devolver error")
		}
		if !IsAuth(err) {
			t.Errorf("401 debe clasificarse como auth; IsAuth=false para %v", err)
		}
	})

	t.Run("5xx => error no-auth", func(t *testing.T) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c := New(srv.URL, "tok")
		c.HTTP = srv.Client()

		err := c.Verify()
		if err == nil {
			t.Fatal("5xx debe devolver error")
		}
		if IsAuth(err) {
			t.Errorf("5xx NO debe clasificarse como auth: %v", err)
		}
	})
}
