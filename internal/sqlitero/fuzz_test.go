package sqlitero

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ═══ P-010 B1 · FUZZING (SC-009) ═══════════════════════════════════════════════════════════════════════════════════
//
// En la suite sólo corren las semillas, todas sintéticas. Los 10 minutos por objetivo son de C2 (T052):
//
//	go test ./internal/sqlitero -run '^$' -fuzz '^FuzzBase$' -fuzztime 10m
//	go test ./internal/sqlitero -run '^$' -fuzz '^FuzzWAL$' -fuzztime 10m
//
// Cada entrada se lee en una goroutine CON recover y con un tope de 1 s (H7): un pánico o un exceso son un fallo del
// test, nunca un binario caído.

const topePorEntrada = time.Second

// sinPanicoNiExceso lee la base de ruta con leerTodo (que recupera los pánicos) y falla si hay pánico o si tarda más
// de un segundo. Los errores de lectura son lo esperado con bytes arbitrarios.
func sinPanicoNiExceso(t *testing.T, ruta string) {
	t.Helper()
	hecho := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				hecho <- errPanico{r}
			}
		}()
		hecho <- leerTodo(ruta)
	}()
	select {
	case err := <-hecho:
		if p, ok := err.(errPanico); ok {
			t.Fatalf("pánico: %v", p.valor)
		}
		if err != nil && strings.HasPrefix(err.Error(), "pánico") {
			t.Fatalf("%v", err)
		}
	case <-time.After(topePorEntrada):
		t.Fatalf("la lectura tardó más de %v", topePorEntrada)
	}
}

type errPanico struct{ valor any }

func (e errPanico) Error() string { return "pánico" }

func semilla(f *testing.F, rel string) []byte {
	f.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", rel))
	if err != nil {
		f.Fatal(err)
	}
	return b
}

// FuzzBase · bytes arbitrarios como base, sin -wal.
func FuzzBase(f *testing.F) {
	for _, rel := range []string{"s/opencode.db", "i/opencode.db", "profundo.db", "raros/p512.db", "raros/p65536.db",
		"raros/reserva.db", "raros/utf16.db", "raros/without_rowid.db", "raros/desborda.db"} {
		f.Add(semilla(f, rel))
	}
	enorme := semilla(f, "profundo.db") // la última semilla: la carga de 2⁶² de (9), la que tumba mB8
	tamPag := int(enorme[16])<<8 | int(enorme[17])
	cargaEnorme(enorme, tamPag, raizDeFila(f))
	f.Add(enorme)
	f.Fuzz(func(t *testing.T, base []byte) {
		ruta := filepath.Join(t.TempDir(), "opencode.db")
		if err := os.WriteFile(ruta, base, 0o600); err != nil {
			t.Fatal(err)
		}
		sinPanicoNiExceso(t, ruta)
	})
}

// FuzzWAL · la base de I, sana, con un -wal arbitrario al lado.
func FuzzWAL(f *testing.F) {
	base := semilla(f, "i/opencode.db")
	wal := semilla(f, "i/opencode.db-wal")
	f.Add(wal)
	f.Add(semilla(f, "r/opencode.db-wal"))
	f.Add([]byte{})
	f.Add(wal[:32])
	f.Add(wal[:32+24+4096+10])
	f.Fuzz(func(t *testing.T, w []byte) {
		dir := t.TempDir()
		ruta := filepath.Join(dir, "opencode.db")
		if err := os.WriteFile(ruta, base, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta+"-wal", w, 0o600); err != nil {
			t.Fatal(err)
		}
		sinPanicoNiExceso(t, ruta)
	})
}

// raizDeFila lee la raíz de la tabla `fila` de una copia de profundo.db.
func raizDeFila(f *testing.F) uint32 {
	f.Helper()
	dir := f.TempDir()
	ruta := filepath.Join(dir, "profundo.db")
	if err := os.WriteFile(ruta, semilla(f, "profundo.db"), 0o600); err != nil {
		f.Fatal(err)
	}
	b, err := Abrir(ruta)
	if err != nil {
		f.Fatal(err)
	}
	defer b.Cerrar()
	tb, err := b.Tabla("fila")
	if err != nil {
		f.Fatal(err)
	}
	return tb.raiz
}
