package state

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFindLogs enumera *.jsonl bajo una raíz anidada e ignora otros ficheros.
func TestFindLogs(t *testing.T) {
	root := t.TempDir()
	proj := filepath.Join(root, "proj-a", "nested")
	if err := os.MkdirAll(proj, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(proj, "s1.jsonl"), "{}\n")
	writeFile(t, filepath.Join(proj, "s2.jsonl"), "{}\n")
	writeFile(t, filepath.Join(proj, "notes.txt"), "x") // ignorado
	writeFile(t, filepath.Join(root, "root.jsonl"), "{}\n")

	got, err := FindLogs(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("esperaba 3 .jsonl, got %d: %v", len(got), got)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Fatalf("resultado no ordenado de forma estable: %v", got)
		}
	}
}

// TestFindLogs_MissingRoot: sin datos (Claude Code nunca usado) no falla (edge case).
func TestFindLogs_MissingRoot(t *testing.T) {
	got, err := FindLogs(filepath.Join(t.TempDir(), "no-existe"))
	if err != nil {
		t.Fatalf("raíz ausente no debe fallar: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("esperaba 0 rutas, got %d", len(got))
	}
}
