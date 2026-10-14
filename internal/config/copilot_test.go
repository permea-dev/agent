package config

import (
	"path/filepath"
	"testing"
)

// P-011 B1 · Dónde están las sesiones de Copilot CLI (D-5, FR-001; SC-012). `CopilotSessionsRoot` resuelve una RUTA: si
// existe se mira en cada pasada (P-6, B4), así que una ruta inexistente se devuelve igual, sin error. Reutiliza
// `hogarDePrueba` de `codex_test.go` y `sinVariableOpenCode` de `opencode_test.go`, sin tocarlos. Las rutas son de
// `t.TempDir()`.

// comprobarCopilotSessionsRoot llama a `CopilotSessionsRoot` y compara la ruta.
func comprobarCopilotSessionsRoot(t *testing.T, esperada string) {
	t.Helper()
	got, err := CopilotSessionsRoot()
	if err != nil || got != esperada {
		t.Errorf("CopilotSessionsRoot() = (%q, %v); se esperaba (%q, nil)", got, err, esperada)
	}
}

// (1) · D-5: `$COPILOT_HOME/session-state` si la variable no es vacía; vacía o ausente, `<hogar>/.copilot/session-state`.
// La variable sustituye a `~/.copilot` entero (m2 M3): `session-state` cuelga directamente de ella.
func TestCopilotSessionsRoot_Ubicacion(t *testing.T) {
	t.Run("definida", func(t *testing.T) {
		hogarDePrueba(t)
		propia := t.TempDir()
		t.Setenv("COPILOT_HOME", propia)
		comprobarCopilotSessionsRoot(t, filepath.Join(propia, "session-state"))
	})
	t.Run("vacia", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		t.Setenv("COPILOT_HOME", "")
		comprobarCopilotSessionsRoot(t, filepath.Join(hogar, ".copilot", "session-state"))
	})
	t.Run("ausente", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		sinVariableOpenCode(t, "COPILOT_HOME")
		comprobarCopilotSessionsRoot(t, filepath.Join(hogar, ".copilot", "session-state"))
	})
	t.Run("inexistente", func(t *testing.T) {
		hogar := filepath.Join(t.TempDir(), "no-existe")
		t.Setenv("HOME", hogar)
		t.Setenv("USERPROFILE", hogar)
		sinVariableOpenCode(t, "COPILOT_HOME")
		comprobarCopilotSessionsRoot(t, filepath.Join(hogar, ".copilot", "session-state"))
	})
}
