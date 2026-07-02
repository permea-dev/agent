package state

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestIncremental_NoReprocess: una llamada ya leída queda por debajo del offset y
// nunca se reprocesa; una línea nueva sí (FR-006, SC-003).
func TestIncremental_NoReprocess(t *testing.T) {
	dir := t.TempDir()
	logp := filepath.Join(dir, "a.jsonl")
	writeFile(t, logp, "line1\nline2\n")

	st := New()
	var got []string
	collect := func(line []byte) error {
		got = append(got, string(line))
		return nil
	}
	if err := st.ScanFile(logp, collect); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("primera pasada: esperaba 2 líneas, got %d", len(got))
	}

	// Segunda pasada sin cambios -> 0 líneas nuevas.
	got = nil
	if err := st.ScanFile(logp, collect); err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("segunda pasada: esperaba 0 líneas nuevas, got %d", len(got))
	}

	// Añadir una línea -> solo esa se procesa.
	f, err := os.OpenFile(logp, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("line3\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	got = nil
	if err := st.ScanFile(logp, collect); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "line3\n" {
		t.Fatalf("esperaba solo line3, got %v", got)
	}
}

// TestIncremental_PartialLine: una línea sin '\n' final (fichero aún escribiéndose)
// no se cuenta hasta completarse.
func TestIncremental_PartialLine(t *testing.T) {
	dir := t.TempDir()
	logp := filepath.Join(dir, "a.jsonl")
	writeFile(t, logp, "complete\npartial-sin-salto")

	st := New()
	var got []string
	if err := st.ScanFile(logp, func(l []byte) error { got = append(got, string(l)); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "complete\n" {
		t.Fatalf("solo la línea completa debe procesarse, got %v", got)
	}

	// Completar la línea parcial -> ahora se procesa.
	f, _ := os.OpenFile(logp, os.O_APPEND|os.O_WRONLY, 0o600)
	f.WriteString("-ahora-completa\n")
	f.Close()
	got = nil
	if err := st.ScanFile(logp, func(l []byte) error { got = append(got, string(l)); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "partial-sin-salto-ahora-completa\n" {
		t.Fatalf("la línea completada debe procesarse entera, got %v", got)
	}
}

// TestIncremental_Truncation: si el fichero encoge por debajo del offset (rotado/
// truncado), se relee desde 0 para no perder datos.
func TestIncremental_Truncation(t *testing.T) {
	dir := t.TempDir()
	logp := filepath.Join(dir, "a.jsonl")
	writeFile(t, logp, "aaaa\nbbbb\n")
	st := New()
	n := 0
	if err := st.ScanFile(logp, func(_ []byte) error { n++; return nil }); err != nil {
		t.Fatal(err)
	}
	writeFile(t, logp, "c\n") // más pequeño que el offset previo
	n = 0
	if err := st.ScanFile(logp, func(_ []byte) error { n++; return nil }); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("tras truncado esperaba releer 1 línea, got %d", n)
	}
}

// TestScan_TwoPasses_NoDuplicates: con el estado persistido en state.json entre
// pasadas, la segunda produce 0 eventos nuevos (SC-003, T023).
func TestScan_TwoPasses_NoDuplicates(t *testing.T) {
	dir := t.TempDir()
	logp := filepath.Join(dir, "s.jsonl")
	writeFile(t, logp, "x\ny\nz\n")
	statePath := filepath.Join(dir, "state.json")

	pass := func() int {
		st, err := Load(statePath)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		if err := st.ScanFile(logp, func(_ []byte) error { n++; return nil }); err != nil {
			t.Fatal(err)
		}
		if err := st.Save(statePath); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if got := pass(); got != 3 {
		t.Fatalf("primera pasada esperaba 3, got %d", got)
	}
	if got := pass(); got != 0 {
		t.Fatalf("segunda pasada esperaba 0 (idempotente), got %d", got)
	}
}
