package config

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

// mkEnroll construye un enrollment string pmea1.<base64url(json)> para los tests,
// con el mismo formato que P-002 emitirá (contracts/enrollment-string.md).
func mkEnroll(t *testing.T, endpoint, token string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"endpoint": endpoint, "token": token})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return "pmea1." + base64.RawURLEncoding.EncodeToString(b)
}

// TestParseEnrollmentString_Valid cubre la variante (a): un enrollment string
// bien formado decodifica al par (endpoint, token).
func TestParseEnrollmentString_Valid(t *testing.T) {
	const endpoint = "https://api.permea.example/ingest"
	const token = "dev_tok_demo"

	gotEndpoint, gotToken, err := ParseEnrollmentString(mkEnroll(t, endpoint, token))
	if err != nil {
		t.Fatalf("esperaba éxito, got error: %v", err)
	}
	if gotEndpoint != endpoint || gotToken != token {
		t.Fatalf("decodificado = (%q, %q), quiero (%q, %q)", gotEndpoint, gotToken, endpoint, token)
	}
}

// TestParseEnrollmentString_Rejects cubre las variantes (b)-(f) y la garantía de
// higiene del secreto (g): ningún error reproduce el argumento ni filtra el token.
func TestParseEnrollmentString_Rejects(t *testing.T) {
	// Marca de secreto para la aserción de higiene (variante g).
	const secret = "dev_tok_SUPERSECRET_zzz"
	validJSON := `{"endpoint":"https://x.example/ingest","token":"` + secret + `"}`
	b64valid := base64.RawURLEncoding.EncodeToString([]byte(validJSON))

	cases := []struct {
		name string
		in   string
	}{
		{"prefijo desconocido", "permea1." + b64valid},
		{"sin prefijo", b64valid},
		{"base64 inválido", "pmea1.###" + secret + "###"},
		{"json malformado", "pmea1." + base64.RawURLEncoding.EncodeToString([]byte(`{"endpoint":"https://x.example/ingest","token":`+secret))},
		{"endpoint http (no https)", mkEnroll(t, "http://inseguro.example/ingest", secret)},
		{"endpoint vacío", mkEnroll(t, "", secret)},
		{"token vacío", mkEnroll(t, "https://api.permea.example/ingest", "")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			endpoint, token, err := ParseEnrollmentString(tc.in)
			if err == nil {
				t.Fatalf("esperaba error, got (%q, %q)", endpoint, token)
			}
			if endpoint != "" || token != "" {
				t.Fatalf("en error, endpoint/token deben ser vacíos; got (%q, %q)", endpoint, token)
			}
			// (g) El error NUNCA reproduce el argumento ni el token (FR-007/FR-013).
			msg := err.Error()
			if strings.Contains(msg, tc.in) {
				t.Errorf("el error reproduce el argumento de entrada: %q", msg)
			}
			if secret != "" && strings.Contains(msg, secret) {
				t.Errorf("el error filtra el token secreto: %q", msg)
			}
		})
	}
}
