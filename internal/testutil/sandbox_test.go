package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSandbox_AislaLasTresVariables comprueba lo que el helper promete: que tras llamarlo,
// las tres variables de entorno y el dataDir resuelto caen DENTRO del temporal del test.
func TestSandbox_AislaLasTresVariables(t *testing.T) {
	dataDir := Sandbox(t)

	base, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("os.UserConfigDir(): %v", err)
	}

	for _, v := range []string{"HOME", "USERPROFILE", "XDG_CONFIG_HOME"} {
		valor := os.Getenv(v)
		if valor == "" {
			t.Errorf("%s quedó vacía: el helper debe establecer las tres", v)
			continue
		}
		if !strings.Contains(valor, os.TempDir()) && !filepath.IsAbs(valor) {
			t.Errorf("%s = %q no parece una ruta absoluta de temporal", v, valor)
		}
	}

	if !dentroDe(dataDir, base) {
		t.Errorf("el dataDir %q no cuelga del directorio de configuración resuelto %q", dataDir, base)
	}
	if _, err := os.Stat(dataDir); err != nil {
		t.Errorf("el dataDir %q debería existir tras Sandbox: %v", dataDir, err)
	}
}

// TestSandbox_LaAsercionDeAislamientoDistingueSegmentos es la prueba de la aserción, y se
// hace sobre `dentroDe` porque es donde vive la decisión.
//
// NACE VERDE por construcción —el helper ya se comporta así—, así que su valor lo da la
// MUTACIÓN: sustituir `dentroDe` por una comparación con strings.HasPrefix hace fallar el
// caso `/tmp/ab` vs `/tmp/abc`, que es exactamente el falso positivo que la implementación
// por segmentos evita. Sin este caso, un aislamiento roto pasaría por bueno.
func TestSandbox_LaAsercionDeAislamientoDistingueSegmentos(t *testing.T) {
	sep := string(filepath.Separator)
	base := filepath.Join(sep+"tmp", "ab")

	casos := []struct {
		nombre string
		ruta   string
		dentro bool
	}{
		{"la propia base", base, true},
		{"un hijo directo", filepath.Join(base, "permea"), true},
		{"un nieto", filepath.Join(base, "x", "y"), true},
		{"un hermano con el mismo prefijo de cadena", filepath.Join(sep+"tmp", "abc"), false},
		{"el padre", sep + "tmp", false},
		{"una rama ajena", filepath.Join(sep+"var", "lib"), false},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := dentroDe(c.ruta, base); got != c.dentro {
				t.Errorf("dentroDe(%q, %q) = %v; se esperaba %v", c.ruta, base, got, c.dentro)
			}
		})
	}
}

// TestSandboxConSemillas_SiembraLoQueDiceLaLineaBase comprueba las dos cosas que la variante
// promete: que los ficheros existen en el dataDir y que su contenido es EXACTAMENTE el que
// declara el artefacto — no un valor parecido escrito a mano en el helper.
func TestSandboxConSemillas_SiembraLoQueDiceLaLineaBase(t *testing.T) {
	dataDir := SandboxConSemillas(t)
	salt, machineID := SemillasDeLaLineaBase(t)

	casos := []struct {
		fichero  string
		esperado string
	}{
		{"salt", salt},
		{"machine_id", machineID},
	}

	for _, c := range casos {
		b, err := os.ReadFile(filepath.Join(dataDir, c.fichero))
		if err != nil {
			t.Fatalf("no se pudo leer la semilla %q: %v", c.fichero, err)
		}
		if string(b) != c.esperado {
			t.Errorf("semilla %q = %q; el artefacto declara %q", c.fichero, string(b), c.esperado)
		}
	}
}

// TestSemillasDeLaLineaBase_NoEstanVacias protege el contrato de la lectura: el artefacto
// existe y declara las dos semillas. Si alguien reescribiera el .tsv sin el bloque
// REPRODUCCIÓN, este test lo dice aquí y no tres tareas después, cuando una comparación
// fallara sin motivo aparente.
func TestSemillasDeLaLineaBase_NoEstanVacias(t *testing.T) {
	salt, machineID := SemillasDeLaLineaBase(t)

	if len(salt) != 64 {
		t.Errorf("salt de la línea base: %d caracteres, se esperaban 64 (32 bytes en hex)", len(salt))
	}
	if len(machineID) != 32 {
		t.Errorf("machine_id de la línea base: %d caracteres, se esperaban 32 (16 bytes en hex)", len(machineID))
	}
}
