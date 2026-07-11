package config

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

// mkEnroll construye un enrollment string pmea2.<base64url(json)> para los tests, con el
// mismo formato que el backend (P-002b) emite (contracts/enrollment-string.md): struct
// cerrada de tres campos {endpoint, token, dev_id}.
func mkEnroll(t *testing.T, endpoint, token, devID string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"endpoint": endpoint, "token": token, "dev_id": devID})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return "pmea2." + base64.RawURLEncoding.EncodeToString(b)
}

// mkEnrollV1 construye un enrollment string pmea1 (formato viejo {endpoint, token}) para
// comprobar que se RECHAZA con el error de formato obsoleto (política de versión).
func mkEnrollV1(t *testing.T, endpoint, token string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"endpoint": endpoint, "token": token})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return "pmea1." + base64.RawURLEncoding.EncodeToString(b)
}

// TestParseEnrollmentString_Valid: un pmea2 bien formado decodifica a la terna
// (endpoint, token, dev_id) — los tres valores autoritativos.
func TestParseEnrollmentString_Valid(t *testing.T) {
	const endpoint = "https://api.permea.example/ingest"
	const token = "dev_tok_demo"
	const devID = "acme-dev-01"

	gotEndpoint, gotToken, gotDevID, err := ParseEnrollmentString(mkEnroll(t, endpoint, token, devID))
	if err != nil {
		t.Fatalf("esperaba éxito, got error: %v", err)
	}
	if gotEndpoint != endpoint || gotToken != token || gotDevID != devID {
		t.Fatalf("decodificado = (%q, %q, %q), quiero (%q, %q, %q)",
			gotEndpoint, gotToken, gotDevID, endpoint, token, devID)
	}
}

// TestParseEnrollmentString_RejectsPmea1: el formato viejo pmea1 se RECHAZA con el mensaje
// de formato obsoleto del contrato, sin filtrar el argumento ni el token.
func TestParseEnrollmentString_RejectsPmea1(t *testing.T) {
	const secret = "dev_tok_SUPERSECRET_zzz"
	in := mkEnrollV1(t, "https://x.example/ingest", secret)

	endpoint, token, devID, err := ParseEnrollmentString(in)
	if err == nil {
		t.Fatalf("pmea1 debe rechazarse, got (%q, %q, %q)", endpoint, token, devID)
	}
	if endpoint != "" || token != "" || devID != "" {
		t.Fatalf("en error los tres valores deben ser vacíos; got (%q, %q, %q)", endpoint, token, devID)
	}
	msg := err.Error()
	// Mensaje del contrato: "formato de enrollment obsoleto; solicita uno nuevo desde el panel".
	if !strings.Contains(msg, "obsoleto") {
		t.Errorf("el error de pmea1 debe indicar formato obsoleto; got %q", msg)
	}
	// Higiene del secreto: ni el enrollment string ni el token aparecen en el error.
	if strings.Contains(msg, in) {
		t.Errorf("el error de pmea1 reproduce el argumento: %q", msg)
	}
	if strings.Contains(msg, secret) {
		t.Errorf("el error de pmea1 filtra el token: %q", msg)
	}
}

// TestParseEnrollmentString_Rejects cubre prefijo desconocido, cuerpo malformado, struct
// cerrada, validación de endpoint/token y las tres reglas de dev_id (vacío, charset, >64).
// En todos los casos: error, terna vacía y CERO filtración del argumento o del token.
func TestParseEnrollmentString_Rejects(t *testing.T) {
	const secret = "dev_tok_SUPERSECRET_zzz"
	validBody := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"endpoint":"https://x.example/ingest","token":"` + secret + `","dev_id":"acme-dev-01"}`))
	longDevID := strings.Repeat("a", 65) // 65 > 64

	cases := []struct {
		name string
		in   string
	}{
		{"prefijo desconocido", "permea2." + validBody},
		{"sin prefijo", validBody},
		{"base64 inválido", "pmea2.###" + secret + "###"},
		{"json malformado", "pmea2." + base64.RawURLEncoding.EncodeToString(
			[]byte(`{"endpoint":"https://x.example/ingest","token":`+secret))},
		{"campo extra (struct cerrada)", "pmea2." + base64.RawURLEncoding.EncodeToString(
			[]byte(`{"endpoint":"https://x.example/ingest","token":"`+secret+`","dev_id":"acme-dev-01","extra":"x"}`))},
		{"endpoint http (no https)", mkEnroll(t, "http://inseguro.example/ingest", secret, "acme-dev-01")},
		{"endpoint vacío", mkEnroll(t, "", secret, "acme-dev-01")},
		{"token vacío", mkEnroll(t, "https://api.permea.example/ingest", "", "acme-dev-01")},
		{"dev_id vacío", mkEnroll(t, "https://api.permea.example/ingest", secret, "")},
		{"dev_id charset inválido", mkEnroll(t, "https://api.permea.example/ingest", secret, "acme dev!01")},
		{"dev_id demasiado largo (>64)", mkEnroll(t, "https://api.permea.example/ingest", secret, longDevID)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			endpoint, token, devID, err := ParseEnrollmentString(tc.in)
			if err == nil {
				t.Fatalf("esperaba error, got (%q, %q, %q)", endpoint, token, devID)
			}
			if endpoint != "" || token != "" || devID != "" {
				t.Fatalf("en error los tres valores deben ser vacíos; got (%q, %q, %q)", endpoint, token, devID)
			}
			// Higiene: el error NUNCA reproduce el argumento ni filtra el token (FR-007/FR-013).
			msg := err.Error()
			if strings.Contains(msg, tc.in) {
				t.Errorf("el error reproduce el argumento de entrada: %q", msg)
			}
			if strings.Contains(msg, secret) {
				t.Errorf("el error filtra el token secreto: %q", msg)
			}
		})
	}
}
