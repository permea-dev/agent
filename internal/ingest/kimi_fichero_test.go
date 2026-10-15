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

// P-012 B1 · Qué ficheros lee el lector de Kimi Code (P-3, FR-003). Los fixtures son SINTÉTICOS, de `testdata/kimi/generar.sh`,
// y NUNCA se abren en su sitio: cada test trabaja sobre una copia en `t.TempDir()`.

// copiarArbolKimi copia `testdata/kimi/<nombre>` entero a un temporal y devuelve su ruta.
func copiarArbolKimi(t *testing.T, nombre string) string {
	t.Helper()
	origen := filepath.Join("testdata", "kimi", nombre)
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

// listarKimiRelativo llama a `ListarKimi` sobre una copia de `raiz/` y devuelve las rutas relativas a ella, con `/`.
func listarKimiRelativo(t *testing.T) []string {
	t.Helper()
	raiz := copiarArbolKimi(t, "raiz")
	rutas, err := ListarKimi(raiz)
	if err != nil {
		t.Fatalf("ListarKimi: %v", err)
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
		t.Fatal("ListarKimi no devolvió nada")
	}
	return rel
}

// (3) · P-3: sólo `<raíz>/sessions/<a>/<b>/agents/<c>/wire.jsonl`, con `a`, `b` y `c` que no empiezan por `.`, y `wire.jsonl` un
// fichero regular. Ni `.index-*`, ni carpetas ocultas, ni `.bak` o `.tmp.*`, ni `<raíz>/imported/`, ni un `wire.jsonl` más hondo.
// Una sesión sin `agents/` no es un error. Orden léxico.
func TestListarKimi(t *testing.T) {
	const wd = "sessions/wd_proyecto-a_000000000001/"
	t.Run("forma", func(t *testing.T) {
		want := []string{
			wd + "session_sintetica-0001/agents/agent-0/wire.jsonl",
			wd + "session_sintetica-0001/agents/main/wire.jsonl",
			wd + "session_sintetica-0002/agents/main/wire.jsonl",
		}
		if got := listarKimiRelativo(t); !slices.Equal(got, want) {
			t.Errorf("ListarKimi = %q; se esperaba %q", got, want)
		}
	})
	t.Run("sin_punto", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			for _, c := range strings.Split(r, "/") {
				if strings.HasPrefix(c, ".") {
					t.Errorf("lista %q, de una carpeta oculta", r)
				}
			}
		}
	})
	t.Run("sin_bak_tmp", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			if path.Base(r) != "wire.jsonl" {
				t.Errorf("lista %q, que no es un wire.jsonl", r)
			}
		}
	})
	t.Run("sin_imported", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			if !strings.HasPrefix(r, "sessions/") {
				t.Errorf("lista %q, fuera de sessions/", r)
			}
		}
	})
	t.Run("hondo", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			if strings.Contains(r, "/sub/") {
				t.Errorf("lista %q, más hondo que agents/<agente>/", r)
			}
		}
	})
	t.Run("vacia", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			if strings.Contains(r, "session_sintetica-0003") {
				t.Errorf("lista %q, de la sesión sin agents/", r)
			}
		}
	})
	t.Run("orden", func(t *testing.T) {
		if got := listarKimiRelativo(t); !slices.IsSorted(got) {
			t.Errorf("ListarKimi = %q; se esperaba en orden léxico", got)
		}
	})
}
