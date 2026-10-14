package testutil

import (
	"os"
	"testing"
)

// (2) · P-011 M-3: el sandbox deja `COPILOT_HOME` vacía aunque el desarrollador la tenga definida. Sin eso, un test de
// proceso leería sus sesiones reales de Copilot CLI. La variable se pone ANTES de `Sandbox`: si no, el test pasaría
// también sin la línea del sandbox, en una máquina donde no esté definida.
func TestSandbox_VaciaCopilotHome(t *testing.T) {
	t.Setenv("COPILOT_HOME", t.TempDir())
	Sandbox(t)
	if v, definida := os.LookupEnv("COPILOT_HOME"); !definida || v != "" {
		t.Errorf("tras Sandbox, COPILOT_HOME = %q (definida: %t); se esperaba vacía", v, definida)
	}
}
