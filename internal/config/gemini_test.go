package config

import (
	"os"
	"path/filepath"
	"testing"
)

// P-009 B4 · La raíz de Gemini CLI (FR-001). `GeminiRoot` resuelve la RUTA de `.gemini`; si `<raíz>/tmp` existe se mira
// en cada pasada (B5), así que aquí una raíz inexistente se devuelve igual, sin error. Reutiliza `hogarDePrueba` y `mkdir`
// de `codex_test.go`, sin tocarlo.

// sinGeminiHome deja GEMINI_CLI_HOME sin definir durante el test; t.Setenv la restaura al terminar.
func sinGeminiHome(t *testing.T) {
	t.Helper()
	t.Setenv("GEMINI_CLI_HOME", "restaurada-al-terminar")
	if err := os.Unsetenv("GEMINI_CLI_HOME"); err != nil {
		t.Fatalf("precondición: %v", err)
	}
}

// (26) · FR-001: `GEMINI_CLI_HOME` definida y no vacía manda; vacía o ausente, `<home>/.gemini`.
func TestGeminiRoot_Raiz(t *testing.T) {
	t.Run("definida", func(t *testing.T) {
		hogarDePrueba(t)
		propia := t.TempDir()
		mkdir(t, filepath.Join(propia, ".gemini", "tmp"))
		t.Setenv("GEMINI_CLI_HOME", propia)
		got, err := GeminiRoot()
		if want := filepath.Join(propia, ".gemini"); err != nil || got != want {
			t.Errorf("GeminiRoot() = (%q, %v); se esperaba (%q, nil)", got, err, want)
		}
	})
	t.Run("vacia", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		mkdir(t, filepath.Join(hogar, ".gemini", "tmp"))
		t.Setenv("GEMINI_CLI_HOME", "")
		got, err := GeminiRoot()
		if want := filepath.Join(hogar, ".gemini"); err != nil || got != want {
			t.Errorf(`con GEMINI_CLI_HOME="": (%q, %v); se esperaba (%q, nil)`, got, err, want)
		}
	})
	t.Run("ausente", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		mkdir(t, filepath.Join(hogar, ".gemini", "tmp"))
		sinGeminiHome(t)
		got, err := GeminiRoot()
		if want := filepath.Join(hogar, ".gemini"); err != nil || got != want {
			t.Errorf("sin GEMINI_CLI_HOME: (%q, %v); se esperaba (%q, nil)", got, err, want)
		}
	})
}

// (27) · Una raíz que no existe se devuelve igual, sin error. La activación se decide en cada pasada (FR-002, (35) de B5).
func TestGeminiRoot_InexistenteSinError(t *testing.T) {
	hogar := filepath.Join(t.TempDir(), "no-existe")
	t.Setenv("HOME", hogar)
	t.Setenv("USERPROFILE", hogar)
	sinGeminiHome(t)
	got, err := GeminiRoot()
	if want := filepath.Join(hogar, ".gemini"); err != nil || got != want {
		t.Errorf("con la raíz inexistente: (%q, %v); se esperaba (%q, nil)", got, err, want)
	}
}
