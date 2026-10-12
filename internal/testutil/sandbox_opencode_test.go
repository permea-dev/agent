package testutil

import (
	"os"
	"testing"
)

// (40) · P-010 M-3: el sandbox deja `XDG_DATA_HOME` y `OPENCODE_DB` vacías aunque el desarrollador las tenga definidas.
// Sin eso, un test de proceso leería su base real de OpenCode.
func TestSandbox_VaciaLasVariablesDeOpenCode(t *testing.T) {
	for _, v := range []string{"XDG_DATA_HOME", "OPENCODE_DB"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv(v, t.TempDir())
			Sandbox(t)
			if valor, definida := os.LookupEnv(v); !definida || valor != "" {
				t.Errorf("tras Sandbox, %s = %q (definida: %t); se esperaba vacía", v, valor, definida)
			}
		})
	}
}
