package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/permea-dev/agent/internal/testutil"
)

// ═══ 013 B1 · LEER SIN ESCRIBIR — (1) y (2) de plan §Bloques ═══════════════════════════════════
//
// `permea project list` NUNCA crea nada (013 FR-005, FR-007): ni el `DataDir` ni la sal. Por eso lee
// con `RutaDataDir` y `LeerSalt`, que no crean, y no con `DataDir` y `LoadOrCreateSalt`, que sí.
// Cada test corre en `testutil.Sandbox`: lo que se borra y lo que se mira es del temporal.

// noExiste falla si `ruta` existe: lo que la lectura no debe dejar atrás.
func noExiste(t *testing.T, ruta, que string) {
	t.Helper()
	if _, err := os.Lstat(ruta); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("%s existe tras leer (err = %v): la lectura NUNCA crea", que, err)
	}
}

// (1) · `RutaDataDir` da la ruta de `DataDir` y no la crea; `DataDir` la sigue creando.
func TestRutaDataDir(t *testing.T) {
	t.Run("misma_ruta", func(t *testing.T) {
		esperada := testutil.Sandbox(t)
		got, err := RutaDataDir()
		if err != nil || got != esperada {
			t.Errorf("RutaDataDir() = (%q, %v); se esperaba (%q, nil), la ruta de DataDir", got, err, esperada)
		}
	})
	t.Run("no_crea", func(t *testing.T) {
		esperada := testutil.Sandbox(t)
		if err := os.RemoveAll(esperada); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		got, err := RutaDataDir()
		if err != nil || got != esperada {
			t.Errorf("RutaDataDir() = (%q, %v); se esperaba (%q, nil)", got, err, esperada)
		}
		noExiste(t, esperada, "el DataDir")
	})
	// DataDir conserva su conducta de hoy: crea el directorio. `TestDataDir_Resolves` lo mira sobre el
	// directorio de configuración real, que ya existe; aquí, sobre uno borrado. Lo valida M-B1a.
	t.Run("datadir_crea", func(t *testing.T) {
		esperada := testutil.Sandbox(t)
		if err := os.RemoveAll(esperada); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		got, err := DataDir()
		if err != nil || got != esperada {
			t.Fatalf("DataDir() = (%q, %v); se esperaba (%q, nil)", got, err, esperada)
		}
		if info, err := os.Stat(esperada); err != nil || !info.IsDir() {
			t.Errorf("DataDir() debe crear el directorio: %v", err)
		}
	})
}

// (2) · `LeerSalt` lee la sal tal cual, sin crearla nunca, y distingue la ausente de la ilegible.
func TestLeerSalt(t *testing.T) {
	t.Run("ausente", func(t *testing.T) {
		dataDir := testutil.Sandbox(t)
		if err := os.RemoveAll(dataDir); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		sal, existe, err := LeerSalt(dataDir)
		if sal != "" || existe || err != nil {
			t.Errorf("LeerSalt sin DataDir = (%q, %v, %v); se esperaba (\"\", false, nil)", sal, existe, err)
		}
		noExiste(t, filepath.Join(dataDir, "salt"), "la sal")
		noExiste(t, dataDir, "el DataDir")
	})
	t.Run("ausente_con_datadir", func(t *testing.T) {
		dataDir := testutil.Sandbox(t)
		sal, existe, err := LeerSalt(dataDir)
		if sal != "" || existe || err != nil {
			t.Errorf("LeerSalt sin sal = (%q, %v, %v); se esperaba (\"\", false, nil)", sal, existe, err)
		}
		noExiste(t, filepath.Join(dataDir, "salt"), "la sal")
	})
	// La sal es la semilla de las identidades: leída con un byte de más o de menos, la lista daría
	// identidades que el panel no tiene. Por eso se lee como `LoadOrCreateSalt`, también con un `\n` final.
	for _, c := range []struct{ hoja, contenido string }{
		{"igual", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		{"igual_con_salto", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\n"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			dataDir := testutil.Sandbox(t)
			if err := os.WriteFile(filepath.Join(dataDir, "salt"), []byte(c.contenido), 0o600); err != nil {
				t.Fatalf("precondición: %v", err)
			}
			referencia, err := LoadOrCreateSalt(dataDir)
			if err != nil || referencia != c.contenido {
				t.Fatalf("precondición: LoadOrCreateSalt = (%q, %v)", referencia, err)
			}
			sal, existe, err := LeerSalt(dataDir)
			if sal != referencia || !existe || err != nil {
				t.Errorf("LeerSalt = (%q, %v, %v); se esperaba (%q, true, nil), lo mismo que LoadOrCreateSalt", sal, existe, err, referencia)
			}
		})
	}
	// Ilegible: un directorio en el sitio de la sal. Falla también con root, a diferencia de un 0o000.
	t.Run("ilegible", func(t *testing.T) {
		dataDir := testutil.Sandbox(t)
		if err := os.Mkdir(filepath.Join(dataDir, "salt"), 0o700); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		sal, existe, err := LeerSalt(dataDir)
		if err == nil || sal != "" || existe {
			t.Errorf("LeerSalt con la sal ilegible = (%q, %v, %v); se esperaba un error", sal, existe, err)
		}
	})
}
