package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/permea-dev/agent/internal/project"
)

// ═══ 013 B5 · LA TABLA — (13) a (16) de plan §Bloques ═════════════════════════════════════════════════════════════════
//
// `formatear` es pura *(D-013-P5)*: recibe las filas y el hogar, y no mira el disco. El texto esperado de (13) sale de spec §Textos
// POR PROGRAMA: el bloque que sigue a «**La salida**». Un cambio en la spec cambia la prueba, y no hay una segunda copia.

// hogarDelEjemplo es el hogar de las filas de (13): con él, `/sintetico/casa/dev/…` se escribe `~/dev/…`.
const hogarDelEjemplo = "/sintetico/casa"

// salidaDeLaSpec devuelve el bloque de código que sigue a «**La salida**» en spec §Textos, con un `\n` por línea.
func salidaDeLaSpec(t *testing.T) string {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "specs", "013-lista-de-proyectos", "spec.md"))
	if err != nil {
		t.Fatalf("precondición: spec: %v", err)
	}
	defer func() { _ = f.Close() }()
	var b strings.Builder
	enTitulo, enBloque := false, false
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 1<<20)
	for sc.Scan() {
		l := sc.Text()
		switch {
		case strings.HasPrefix(l, "**La salida**"):
			enTitulo = true
		case enTitulo && !enBloque && l == "```":
			enBloque = true
		case enBloque && l == "```":
			return b.String()
		case enBloque:
			b.WriteString(l + "\n")
		}
	}
	t.Fatal("precondición: spec §Textos no tiene el bloque de «La salida»")
	return ""
}

// filaDelEjemplo es una fila de spec §Textos: la identidad empieza por los 12 caracteres que se ven, y la ruta cuelga del hogar.
func filaDelEjemplo(t *testing.T, doce, resto, marca, desde, hasta string, eventos int, herramientas ...string) fila {
	t.Helper()
	d, err1 := time.Parse("2006-01-02", desde)
	h, err2 := time.Parse("2006-01-02", hasta)
	if err1 != nil || err2 != nil {
		t.Fatalf("precondición: fechas %q, %q", desde, hasta)
	}
	return fila{identidad: doce + strings.Repeat("0", 52), ruta: hogarDelEjemplo + resto, marca: marca, herramientas: herramientas,
		eventos: eventos, desde: d, hasta: h}
}

// filasDelEjemplo son las cuatro filas del ejemplo aprobado, en su orden.
func filasDelEjemplo(t *testing.T) []fila {
	return []fila{
		filaDelEjemplo(t, "08926050c4b8", "/dev/proyecto-ejemplo", "", "2026-10-03", "2026-10-23", 1240, "Claude Code", "Codex"),
		filaDelEjemplo(t, "3f1c9a77d02e", "/dev/otro-proyecto", "", "2026-10-12", "2026-10-22", 412, "Claude Code"),
		filaDelEjemplo(t, "b7e40d915a63", "/pruebas/tmp-1", "(sin repositorio)", "2026-10-20", "2026-10-20", 18, "Gemini CLI"),
		filaDelEjemplo(t, "c05d2e8a1f94", "/dev/viejo", "(ya no existe)", "2026-10-05", "2026-10-06", 7, "Claude Code"),
	}
}

// (13) · SC-001: las cuatro filas del ejemplo producen, BYTE A BYTE, el texto de spec §Textos.
func TestFormatear_ElEjemploDeLaSpec(t *testing.T) {
	quiere := salidaDeLaSpec(t)
	if got := formatear(filasDelEjemplo(t), hogarDelEjemplo); got != quiere {
		t.Errorf("la tabla no es la de spec §Textos:\n--- ahora\n%s--- spec\n%s", got, quiere)
	}
}

// (14) · FR-019, FR-022, P-5: el punto de miles, el pie en singular y, sin filas, sólo el pie.
func TestFormatear_MilesYPie(t *testing.T) {
	for _, c := range []struct {
		hoja   string
		n      int
		quiere string
	}{
		{"miles_0", 0, "0"}, {"miles_999", 999, "999"}, {"miles_1000", 1000, "1.000"}, {"miles_1234567", 1234567, "1.234.567"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if got := miles(c.n); got != c.quiere {
				t.Errorf("miles(%d) = %q; se esperaba %q", c.n, got, c.quiere)
			}
		})
	}
	pie := strings.Join(strings.Split(salidaDeLaSpec(t), "\n")[6:8], "\n") + "\n" // las dos líneas del pie de la spec
	t.Run("singular", func(t *testing.T) {
		f := filasDelEjemplo(t)[3]
		f.eventos = 1
		quiere := "1 identidad, 1 evento. Fechas en UTC, como en el panel.\n" + strings.SplitN(pie, "\n", 2)[1]
		if got := formatear([]fila{f}, hogarDelEjemplo); !strings.HasSuffix(got, "\n\n"+quiere) {
			t.Errorf("con una identidad y un evento, el pie no es:\n%s\nsalida:\n%s", quiere, got)
		}
	})
	t.Run("sin_filas", func(t *testing.T) {
		quiere := "0 identidades, 0 eventos. Fechas en UTC, como en el panel.\n" + strings.SplitN(pie, "\n", 2)[1]
		if got := formatear(nil, hogarDelEjemplo); got != quiere {
			t.Errorf("sin filas, se esperaba sólo el pie:\n%s\nsalida:\n%s", quiere, got)
		}
	})
}

// (15) · P-10, SC-011: la carpeta personal como `~`, comparada por segmentos y con el hogar real *(`HogarReal`)*.
func TestACasa(t *testing.T) {
	sep := string(filepath.Separator)
	for _, c := range []struct{ hoja, ruta, quiere string }{
		{"hogar", hogarDelEjemplo, "~"},
		{"dentro", hogarDelEjemplo + "/dev/x", "~" + sep + filepath.Join("dev", "x")},
		{"hermano", hogarDelEjemplo + "2/x", hogarDelEjemplo + "2/x"},
		{"fuera", "/sintetico/otra/x", "/sintetico/otra/x"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if got := aCasa(c.ruta, hogarDelEjemplo); got != c.quiere {
				t.Errorf("aCasa(%q, %q) = %q; se esperaba %q", c.ruta, hogarDelEjemplo, got, c.quiere)
			}
		})
	}
	// Un hogar al que se llega por un enlace: la ruta que entró en el hash es la real, y `HogarReal` también lo es.
	t.Run("enlace", func(t *testing.T) {
		base, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatalf("precondición: %v", err)
		}
		casaReal, enlace := filepath.Join(base, "casa-real"), filepath.Join(base, "casa-enlace")
		if err := os.MkdirAll(filepath.Join(casaReal, "dev", "x"), 0o700); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		if err := os.Symlink(casaReal, enlace); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		t.Setenv("HOME", enlace)
		t.Setenv("USERPROFILE", enlace)
		hogar, ok := project.HogarReal()
		if got := aCasa(filepath.Join(casaReal, "dev", "x"), hogar); !ok || hogar != casaReal || got != "~"+sep+filepath.Join("dev", "x") {
			t.Errorf("HogarReal = (%q, %v); aCasa = %q; se esperaba (%q, true) y ~/dev/x", hogar, ok, got, casaReal)
		}
	})
}

// (16) · FR-021, P-12: ninguna línea acaba en espacio, y una ruta con tildes no descuadra: el ancho va en runas.
func TestFormatear_Alineacion(t *testing.T) {
	filas := []fila{
		filaDelEjemplo(t, "aaaaaaaaaaaa", "/dev/camión-ñandú", "", "2026-10-01", "2026-10-02", 12, "Claude Code"),
		filaDelEjemplo(t, "bbbbbbbbbbbb", "/dev/sencillo", "", "2026-10-01", "2026-10-02", 3, "Codex"),
	}
	lineas := strings.Split(strings.TrimSuffix(formatear(filas, hogarDelEjemplo), "\n"), "\n")
	t.Run("sin_espacio_final", func(t *testing.T) {
		if len(lineas) != 6 { // cabecera, 2 filas, la línea en blanco y las 2 del pie
			t.Fatalf("%d líneas; se esperaban 6:\n%s", len(lineas), strings.Join(lineas, "\n"))
		}
		for i, l := range lineas {
			if strings.HasSuffix(l, " ") {
				t.Errorf("la línea %d acaba en espacio: %q", i+1, l)
			}
		}
	})
	t.Run("tildes", func(t *testing.T) {
		if len(lineas) < 3 {
			t.Fatalf("%d líneas; se esperaban la cabecera y 2 filas", len(lineas))
		}
		columna := func(l, que string) int { return utf8.RuneCountInString(l[:strings.Index(l, que)]) }
		if a, b, c := columna(lineas[0], "HERRAMIENTAS"), columna(lineas[1], "Claude Code"), columna(lineas[2], "Codex"); a != b || b != c {
			t.Errorf("HERRAMIENTAS empieza en la runa %d, %d y %d: una ruta con tildes descuadra la tabla:\n%s", a, b, c,
				strings.Join(lineas[:3], "\n"))
		}
	})
}
