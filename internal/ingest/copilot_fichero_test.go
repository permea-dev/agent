package ingest

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// P-011 B1 · Qué ficheros lee el lector de Copilot CLI (P-7, FR-003, D-6). Los fixtures son SINTÉTICOS, de
// `testdata/copilot/generar.sh`, y NUNCA se abren en su sitio: cada test trabaja sobre una copia en `t.TempDir()`.

// copiarArbolCopilot copia `testdata/copilot/<nombre>` entero a un temporal y devuelve su ruta.
func copiarArbolCopilot(t *testing.T, nombre string) string {
	t.Helper()
	origen := filepath.Join("testdata", "copilot", nombre)
	destino := filepath.Join(t.TempDir(), nombre)
	err := filepath.WalkDir(origen, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(origen, p)
		if err != nil {
			return err
		}
		a := filepath.Join(destino, rel)
		if d.IsDir() {
			return os.MkdirAll(a, 0o700)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(a, b, 0o600)
	})
	if err != nil {
		t.Fatalf("copiar %s: %v", origen, err)
	}
	return destino
}

// sesionesDePrueba es la copia de `sesiones/` con el `session.db` de 0 B que `.gitignore` no deja versionar (generar.sh).
func sesionesDePrueba(t *testing.T) string {
	t.Helper()
	raiz := copiarArbolCopilot(t, "sesiones")
	if err := os.WriteFile(filepath.Join(raiz, "s-000000000000000000000001", "session.db"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return raiz
}

// listarCopilotRelativo llama a `ListarCopilot` y devuelve las rutas relativas a la raíz, con `/`.
func listarCopilotRelativo(t *testing.T, raiz string) []string {
	t.Helper()
	rutas, err := ListarCopilot(raiz)
	if err != nil {
		t.Fatalf("ListarCopilot: %v", err)
	}
	var rel []string
	for _, r := range rutas {
		x, err := filepath.Rel(raiz, r)
		if err != nil {
			t.Fatal(err)
		}
		rel = append(rel, filepath.ToSlash(x))
	}
	if len(rel) == 0 {
		t.Fatal("ListarCopilot no devolvió nada")
	}
	return rel
}

// (3) · P-7: sólo `<raíz>/<carpeta>/events.jsonl`, de un nivel; ni `.session-operation-locks/`, ni `checkpoints/`,
// `session.db` o un `events.jsonl` más hondo. La carpeta sin `events.jsonl` no es un error ni se avisa (D-6). Orden léxico.
func TestListarCopilot(t *testing.T) {
	s1 := "s-000000000000000000000001/events.jsonl"
	s2 := "s-000000000000000000000002/events.jsonl"
	t.Run("solo_events", func(t *testing.T) {
		got := listarCopilotRelativo(t, sesionesDePrueba(t))
		if want := []string{s1, s2}; !slices.Equal(got, want) {
			t.Errorf("ListarCopilot = %q; se esperaba %q", got, want)
		}
	})
	t.Run("sin_locks", func(t *testing.T) {
		for _, r := range listarCopilotRelativo(t, sesionesDePrueba(t)) {
			if strings.HasPrefix(r, ".") {
				t.Errorf("lista %q, de una carpeta oculta", r)
			}
		}
	})
	t.Run("sin_checkpoints_db", func(t *testing.T) {
		for _, r := range listarCopilotRelativo(t, sesionesDePrueba(t)) {
			if path.Base(r) != "events.jsonl" || strings.Contains(r, "checkpoints/") {
				t.Errorf("lista %q, que no es el events.jsonl de una sesión", r)
			}
		}
	})
	t.Run("hondo", func(t *testing.T) {
		for _, r := range listarCopilotRelativo(t, sesionesDePrueba(t)) {
			if strings.Count(r, "/") != 1 {
				t.Errorf("lista %q, a más de un nivel", r)
			}
		}
	})
	t.Run("vacia", func(t *testing.T) {
		for _, r := range listarCopilotRelativo(t, sesionesDePrueba(t)) {
			if strings.HasPrefix(r, "s-000000000000000000000003/") {
				t.Errorf("lista %q, de la carpeta sin events.jsonl", r)
			}
		}
	})
	t.Run("orden", func(t *testing.T) {
		got := listarCopilotRelativo(t, sesionesDePrueba(t))
		if len(got) < 2 || !slices.IsSorted(got) {
			t.Errorf("ListarCopilot = %q; se esperaba en orden léxico", got)
		}
	})
}
