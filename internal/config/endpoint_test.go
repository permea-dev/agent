package config

import (
	"strings"
	"testing"
)

// ═══ P-005 T004 · LOS DOS HECHOS, COMPROBADOS POR SEPARADO ═════════════════════════════════
//
// Este test NACE VERDE —la función no tiene llamantes todavía— así que su validación es **por
// mutación** (disciplina 3). Y las dos mitades se comprueban en **tests distintos a propósito**: es
// lo que permite que una mutación tumbe UNO y deje el OTRO en pie. Con un solo test de tabla, una
// mutación cualquiera los tumbaría a los dos y **no demostraría que los hechos están separados**,
// que es el punto entero de T004.

// endpointNoAnalizable es una URL que `url.Parse` rechaza. El carácter de control en el host es la
// forma más estable de conseguirlo: `url.Parse` es deliberadamente permisivo y casi todo lo acepta.
const endpointNoAnalizable = "https://ejemplo\x7f.test/ingest"

// TestJuzgarEndpoint_HechoAnalisis cubre SOLO el primer hecho.
func TestJuzgarEndpoint_HechoAnalisis(t *testing.T) {
	t.Run("no analizable devuelve la causa", func(t *testing.T) {
		errAnalisis, _ := JuzgarEndpoint(endpointNoAnalizable)
		if errAnalisis == nil {
			t.Fatalf("JuzgarEndpoint(%q): errAnalisis = nil, se esperaba la causa de url.Parse", endpointNoAnalizable)
		}
	})

	t.Run("analizable no devuelve causa", func(t *testing.T) {
		errAnalisis, _ := JuzgarEndpoint("https://api.permea.example/api/v1/ingest")
		if errAnalisis != nil {
			t.Fatalf("JuzgarEndpoint(url válida): errAnalisis = %v, se esperaba nil", errAnalisis)
		}
	})

	// La causa se devuelve TAL CUAL, sin envolver: es lo que permite que `config.go` y `Send` la
	// metan en su propio `%w` y conserven sus mensajes (condición de parada de T005).
	t.Run("la causa es la de url.Parse, sin envolver", func(t *testing.T) {
		errAnalisis, _ := JuzgarEndpoint(endpointNoAnalizable)
		if errAnalisis == nil || !strings.Contains(errAnalisis.Error(), "parse") {
			t.Fatalf("errAnalisis = %v; se esperaba la causa cruda de url.Parse", errAnalisis)
		}
	})
}

// TestJuzgarEndpoint_HechoEsquema cubre SOLO el segundo hecho, y siempre sobre URLs ANALIZABLES:
// así una mutación del primer hecho no puede tumbarlo de rebote.
func TestJuzgarEndpoint_HechoEsquema(t *testing.T) {
	casos := []struct {
		nombre   string
		endpoint string
		quiere   bool
	}{
		{"https es admisible", "https://api.permea.example/api/v1/ingest", true},
		{"https con puerto no estándar es admisible", "https://localhost:8443/api/v1/ingest", true},
		{"http NO es admisible", "http://api.permea.example/api/v1/ingest", false},
		{"otro esquema NO es admisible", "ftp://api.permea.example/api/v1/ingest", false},
		{"sin esquema NO es admisible", "api.permea.example/api/v1/ingest", false},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			errAnalisis, admisible := JuzgarEndpoint(c.endpoint)
			if errAnalisis != nil {
				t.Fatalf("%q no era analizable (%v): este test solo juzga el esquema", c.endpoint, errAnalisis)
			}
			if admisible != c.quiere {
				t.Errorf("JuzgarEndpoint(%q): admisible = %v, want %v", c.endpoint, admisible, c.quiere)
			}
		})
	}
}

// TestJuzgarEndpoint_NoAnalizableNoAfirmaEsquema es la costura entre los dos hechos: cuando el
// primero falla, el segundo NO puede decir «admisible». Los llamantes que FUNDEN los dos —
// `enrollment.go`— dependen de esto para que su desenlace único siga siendo correcto.
func TestJuzgarEndpoint_NoAnalizableNoAfirmaEsquema(t *testing.T) {
	errAnalisis, admisible := JuzgarEndpoint(endpointNoAnalizable)
	if errAnalisis == nil {
		t.Fatalf("el caso base falló: %q debería no ser analizable", endpointNoAnalizable)
	}
	if admisible {
		t.Error("admisible = true sobre un endpoint no analizable: se estaría afirmando lo que no se comprobó")
	}
}
