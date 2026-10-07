package config

import (
	"os"
	"path/filepath"
	"testing"
)

// P-008 B4 · La raíz de Codex (FR-001, FR-002, E-2). `CodexSessionsRoot` resuelve la RUTA; si existe se mira en
// cada pasada (B5), así que aquí una raíz inexistente se devuelve igual, sin error.

// hogarDePrueba fija el directorio personal en un temporal (HOME y USERPROFILE, por sistema) y lo devuelve.
func hogarDePrueba(t *testing.T) string {
	t.Helper()
	hogar := t.TempDir()
	t.Setenv("HOME", hogar)
	t.Setenv("USERPROFILE", hogar)
	return hogar
}

// sinCodexHome deja CODEX_HOME sin definir durante el test; t.Setenv la restaura al terminar.
func sinCodexHome(t *testing.T) {
	t.Helper()
	t.Setenv("CODEX_HOME", "restaurada-al-terminar")
	if err := os.Unsetenv("CODEX_HOME"); err != nil {
		t.Fatalf("precondición: %v", err)
	}
}

func mkdir(t *testing.T, ruta string) {
	t.Helper()
	if err := os.MkdirAll(ruta, 0o700); err != nil {
		t.Fatalf("precondición: %v", err)
	}
}

// (21) · FR-001: `CODEX_HOME` definida y no vacía manda; vacía o ausente, `<home>/.codex/sessions`.
func TestCodexSessionsRoot_Raiz(t *testing.T) {
	t.Run("definida", func(t *testing.T) {
		hogarDePrueba(t)
		propia := t.TempDir()
		mkdir(t, filepath.Join(propia, "sessions"))
		t.Setenv("CODEX_HOME", propia)
		got, err := CodexSessionsRoot()
		if want := filepath.Join(propia, "sessions"); err != nil || got != want {
			t.Errorf("CodexSessionsRoot() = (%q, %v); se esperaba (%q, nil)", got, err, want)
		}
	})
	t.Run("vacia", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		mkdir(t, filepath.Join(hogar, ".codex", "sessions"))
		t.Setenv("CODEX_HOME", "")
		got, err := CodexSessionsRoot()
		if want := filepath.Join(hogar, ".codex", "sessions"); err != nil || got != want {
			t.Errorf(`con CODEX_HOME="": (%q, %v); se esperaba (%q, nil)`, got, err, want)
		}
	})
	t.Run("ausente", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		mkdir(t, filepath.Join(hogar, ".codex", "sessions"))
		sinCodexHome(t)
		got, err := CodexSessionsRoot()
		if want := filepath.Join(hogar, ".codex", "sessions"); err != nil || got != want {
			t.Errorf("sin CODEX_HOME: (%q, %v); se esperaba (%q, nil)", got, err, want)
		}
	})
}

// (22) · E-2: una raíz que no existe se devuelve igual, sin error. La activación no se decide aquí, sino en cada pasada
// (FR-002, rojo (32) de B5).
func TestCodexSessionsRoot_InexistenteSinError(t *testing.T) {
	hogar := filepath.Join(t.TempDir(), "no-existe")
	t.Setenv("HOME", hogar)
	t.Setenv("USERPROFILE", hogar)
	sinCodexHome(t)
	got, err := CodexSessionsRoot()
	if want := filepath.Join(hogar, ".codex", "sessions"); err != nil || got != want {
		t.Errorf("con la raíz inexistente: (%q, %v); se esperaba (%q, nil)", got, err, want)
	}
}
