package testutil

import (
	"os"
	"testing"
)

// (27-bis) · P-009 M-9: el sandbox deja `GEMINI_CLI_HOME` vacía aunque el desarrollador la tenga definida. Sin eso, un
// test de proceso leería sus sesiones reales de Gemini CLI.
func TestSandbox_VaciaGeminiCliHome(t *testing.T) {
	t.Setenv("GEMINI_CLI_HOME", t.TempDir())
	Sandbox(t)
	if v, definida := os.LookupEnv("GEMINI_CLI_HOME"); !definida || v != "" {
		t.Errorf("tras Sandbox, GEMINI_CLI_HOME = %q (definida: %t); se esperaba vacía", v, definida)
	}
}
