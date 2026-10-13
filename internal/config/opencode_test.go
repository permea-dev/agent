package config

import (
	"os"
	"path/filepath"
	"testing"
)

// P-010 B5 · Dónde están las bases de OpenCode (P-8, FR-001; SC-015). `OpenCodeDatos` resuelve RUTAS: si existen se mira
// en cada pasada (P-10, B6), así que una ruta inexistente se devuelve igual, sin error. Reutiliza `hogarDePrueba` de
// `codex_test.go`, sin tocarlo. Las rutas son de `t.TempDir()` o sintéticas.

// sinVariableOpenCode deja una variable sin definir durante el test; t.Setenv la restaura al terminar.
func sinVariableOpenCode(t *testing.T, nombre string) {
	t.Helper()
	t.Setenv(nombre, "restaurada-al-terminar")
	if err := os.Unsetenv(nombre); err != nil {
		t.Fatalf("precondición: %v", err)
	}
}

// comprobarOpenCodeDatos llama a `OpenCodeDatos` y compara las dos rutas.
func comprobarOpenCodeDatos(t *testing.T, dir, unica string) {
	t.Helper()
	gotDir, gotUnica, err := OpenCodeDatos()
	if err != nil || gotDir != dir || gotUnica != unica {
		t.Errorf("OpenCodeDatos() = (%q, %q, %v); se esperaba (%q, %q, nil)", gotDir, gotUnica, err, dir, unica)
	}
}

// (39) · P-8: `OPENCODE_DB` absoluta → sólo esa base; vacía, `":memory:"` o relativa → la regla general. Y la regla
// general: `$XDG_DATA_HOME/opencode` si no es vacía; vacía o ausente, `<hogar>/.local/share/opencode`. Nunca `AppData`
// ni `Library`: la misma regla en los tres SO.
func TestOpenCodeDatos_Ubicacion(t *testing.T) {
	t.Run("absoluta", func(t *testing.T) {
		hogarDePrueba(t)
		t.Setenv("XDG_DATA_HOME", t.TempDir())
		unica := filepath.Join(t.TempDir(), "opencode-sintetico.db")
		t.Setenv("OPENCODE_DB", unica)
		comprobarOpenCodeDatos(t, "", unica)
	})
	for _, c := range []struct{ hoja, valor string }{
		{"relativa", filepath.Join("sintetico", "opencode.db")},
		{"memory", ":memory:"},
		{"db_vacia", ""},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			hogarDePrueba(t)
			datos := t.TempDir()
			t.Setenv("XDG_DATA_HOME", datos)
			t.Setenv("OPENCODE_DB", c.valor)
			comprobarOpenCodeDatos(t, filepath.Join(datos, "opencode"), "")
		})
	}
	t.Run("xdg_definida", func(t *testing.T) {
		hogarDePrueba(t)
		datos := t.TempDir()
		t.Setenv("XDG_DATA_HOME", datos)
		sinVariableOpenCode(t, "OPENCODE_DB")
		comprobarOpenCodeDatos(t, filepath.Join(datos, "opencode"), "")
	})
	t.Run("xdg_vacia", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		t.Setenv("XDG_DATA_HOME", "")
		sinVariableOpenCode(t, "OPENCODE_DB")
		comprobarOpenCodeDatos(t, filepath.Join(hogar, ".local", "share", "opencode"), "")
	})
	t.Run("xdg_ausente", func(t *testing.T) {
		hogar := hogarDePrueba(t)
		sinVariableOpenCode(t, "XDG_DATA_HOME")
		sinVariableOpenCode(t, "OPENCODE_DB")
		comprobarOpenCodeDatos(t, filepath.Join(hogar, ".local", "share", "opencode"), "")
	})
	t.Run("inexistente", func(t *testing.T) {
		hogar := filepath.Join(t.TempDir(), "no-existe")
		t.Setenv("HOME", hogar)
		t.Setenv("USERPROFILE", hogar)
		sinVariableOpenCode(t, "XDG_DATA_HOME")
		sinVariableOpenCode(t, "OPENCODE_DB")
		comprobarOpenCodeDatos(t, filepath.Join(hogar, ".local", "share", "opencode"), "")
	})
}
