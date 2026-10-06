package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ═══ P-007 B4 · EL OFFSET LO DECIDE QUIEN LLAMA (FR-011, FR-021) ════════════════════════════════════
//
// Un mensaje abierto no se consume: `Recorrer` deja que quien llama fije el offset en el comienzo de ese
// mensaje, y la pasada siguiente lo relee del log. `state.json` no gana ningún campo: `Size` y `ModTime`
// siguen siendo los del stat, y con el `Size` anterior se sabe qué líneas son releídas.

// R-B4a · P-007 FR-011, FR-021 — el offset guardado es el pedido; `Size` y `ModTime` son los del stat; y en
// la pasada siguiente, lo que empieza por debajo del `Size` anterior llega como releído.
func TestRecorrer_GuardaElOffsetPedidoYMarcaLasReleidas(t *testing.T) {
	dir := t.TempDir()
	logp := filepath.Join(dir, "a.jsonl")
	l1, l2, l3 := "{\"n\":1}\n", "{\"n\":2}\n", "{\"n\":3}\n"
	if err := os.WriteFile(logp, []byte(l1+l2+l3), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(logp)
	if err != nil {
		t.Fatal(err)
	}
	inicioDeLaSegunda := int64(len(l1))

	st := New()
	var inicios []int64
	err = st.Recorrer(logp,
		func(_ []byte, inicio int64, _ bool) error { inicios = append(inicios, inicio); return nil },
		func(_ int64, _ time.Time) int64 { return inicioDeLaSegunda })
	if err != nil {
		t.Fatalf("precondición: Recorrer falló: %v", err)
	}
	if len(inicios) != 3 || inicios[1] != inicioDeLaSegunda {
		t.Fatalf("precondición: se esperaban 3 líneas con la segunda en %d; inicios = %v", inicioDeLaSegunda, inicios)
	}

	t.Run("offset_pedido", func(t *testing.T) {
		if got := st.Files[logp].Offset; got != inicioDeLaSegunda {
			t.Errorf("offset guardado = %d; se pidió %d (el comienzo de la segunda línea), no el final", got, inicioDeLaSegunda)
		}
	})
	t.Run("size_y_modtime_del_stat", func(t *testing.T) {
		fs := st.Files[logp]
		if fs.Size != info.Size() || fs.ModTime != info.ModTime().Unix() {
			t.Errorf("Size/ModTime = %d/%d; los del stat son %d/%d", fs.Size, fs.ModTime, info.Size(), info.ModTime().Unix())
		}
	})
	t.Run("releidas", func(t *testing.T) {
		l4 := "{\"n\":4}\n"
		f, err := os.OpenFile(logp, os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.WriteString(l4); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		var releidas []bool
		err = st.Recorrer(logp,
			func(_ []byte, _ int64, releida bool) error { releidas = append(releidas, releida); return nil },
			func(leido int64, _ time.Time) int64 { return leido })
		if err != nil {
			t.Fatalf("precondición: la segunda pasada falló: %v", err)
		}
		// Desde el offset retenido: la 2 y la 3 ya se leyeron; la 4 es nueva.
		if want := []bool{true, true, false}; len(releidas) != 3 || releidas[0] != want[0] || releidas[1] != want[1] || releidas[2] != want[2] {
			t.Errorf("releídas = %v; se esperaba %v", releidas, want)
		}
	})
}
