package testutil

import (
	"os"
	"testing"
)

// (2) · P-012 M-3: el sandbox deja `KIMI_CODE_HOME` vacía aunque el desarrollador la tenga definida. Sin eso, un test de
// proceso leería sus sesiones reales de Kimi Code. La variable se pone ANTES de `Sandbox`: si no, el test pasaría también sin la
// línea del sandbox, en una máquina donde no esté definida.
func TestSandbox_VaciaKimiCodeHome(t *testing.T) {
	t.Setenv("KIMI_CODE_HOME", t.TempDir())
	Sandbox(t)
	if v, definida := os.LookupEnv("KIMI_CODE_HOME"); !definida || v != "" {
		t.Errorf("tras Sandbox, KIMI_CODE_HOME = %q (definida: %t); se esperaba vacía", v, definida)
	}
}
