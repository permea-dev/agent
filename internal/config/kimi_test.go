package config

import (
	"path/filepath"
	"testing"
)

// P-012 B1 · Dónde están los datos de Kimi Code (P-1, FR-001; SC-012). `KimiHome` resuelve una RUTA: si `<ruta>/sessions`
// existe se mira en cada pasada (P-2, B4), así que una ruta inexistente se devuelve igual, sin error. Reutiliza `hogarDePrueba`
// de `codex_test.go` y `sinVariableOpenCode` de `opencode_test.go`, sin tocarlos. Las rutas son de `t.TempDir()`.

// comprobarKimiHome llama a `KimiHome` y compara la ruta.
func comprobarKimiHome(t *testing.T, esperada string) {
	t.Helper()
	got, err := KimiHome()
	if err != nil || got != esperada {
		t.Errorf("KimiHome() = (%q, %v); se esperaba (%q, nil)", got, err, esperada)
	}
}

// (1) · P-1: `$KIMI_CODE_HOME` si la variable no es vacía; vacía o ausente, `<hogar>/.kimi-code`. Vacía cuenta como ausente,
// como `COPILOT_HOME` y como la CLI de Kimi (desc. Q1).
func TestKimiHome_Ubicacion(t *testing.T) {
	t.Run("definida", func(t *testing.T) {
		hogarDePrueba(t)
		propia := t.TempDir()
		t.Setenv("KIMI_CODE_HOME", propia)
		comprobarKimiHome(t, propia)
	})
	t.Run("vacia", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		t.Setenv("KIMI_CODE_HOME", "")
		comprobarKimiHome(t, filepath.Join(hogar, ".kimi-code"))
	})
	t.Run("ausente", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		sinVariableOpenCode(t, "KIMI_CODE_HOME")
		comprobarKimiHome(t, filepath.Join(hogar, ".kimi-code"))
	})
	t.Run("inexistente", func(t *testing.T) {
		hogar := filepath.Join(t.TempDir(), "no-existe")
		t.Setenv("HOME", hogar)
		t.Setenv("USERPROFILE", hogar)
		sinVariableOpenCode(t, "KIMI_CODE_HOME")
		comprobarKimiHome(t, filepath.Join(hogar, ".kimi-code"))
	})
}
