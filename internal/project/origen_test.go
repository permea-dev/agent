package project

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/event"
)

// ═══ 013 B2 · EL ORIGEN — (3) y (4) de plan §Bloques ════════════════════════════════════════
//
// `permea project list` enseña la carpeta de cada identidad, y esa carpeta tiene que ser LA RUTA
// QUE ENTRÓ EN EL HASH (013 FR-015): la raíz, si la hubo; si no, la ubicación limpia. La prueba que
// no depende de la implementación es `event.Ref(sal, origen) == identidad`; las de cada familia
// fijan además QUÉ ruta es.

// familia copia un fichero de `testdata/` a un temporal, lo lee desde allí y devuelve sus `cwd`
// con `{{BASE}}` sustituido por `base`. El fichero de `testdata/` nunca se abre en su sitio.
func familia(t *testing.T, nombre, base string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", nombre))
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	copia := filepath.Join(t.TempDir(), nombre)
	if err := os.WriteFile(copia, b, 0o600); err != nil {
		t.Fatalf("precondición: %v", err)
	}
	f, err := os.Open(copia)
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = f.Close() }()
	var cwds []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var l struct {
			Cwd string `json:"cwd"`
		}
		if err := json.Unmarshal(sc.Bytes(), &l); err != nil {
			t.Fatalf("precondición: %s no decodifica: %v", nombre, err)
		}
		cwds = append(cwds, strings.ReplaceAll(l.Cwd, "{{BASE}}", base))
	}
	return cwds
}

// baseReal es un temporal con los enlaces ya resueltos, para que la ruta esperada sea la real.
func baseReal(t *testing.T) string {
	t.Helper()
	b, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	return b
}

// enlazar crea el enlace simbólico `ruta` → `destino`.
func enlazar(t *testing.T, destino, ruta string) {
	t.Helper()
	if err := os.Symlink(destino, ruta); err != nil {
		t.Fatalf("precondición: enlace %s: %v", ruta, err)
	}
}

// esperado es el origen que debe salir para un `cwd` de la familia.
type esperado struct {
	origen   string
	huboRaiz bool
}

// comprobarFamilia deriva cada `cwd` y exige: la misma identidad y la misma raíz que la derivación
// de siempre, el origen esperado, y que el hash del origen SEA la identidad.
func comprobarFamilia(t *testing.T, cwds []string, quiere map[string]esperado) {
	t.Helper()
	if len(cwds) != len(quiere) {
		t.Fatalf("precondición: %d cwd en la familia y %d esperados", len(cwds), len(quiere))
	}
	for _, cwd := range cwds {
		q, ok := quiere[cwd]
		if !ok {
			t.Fatalf("precondición: sin esperado para %q", cwd)
		}
		identidad, origen, huboRaiz := DerivarConOrigen(cwd, saltDePrueba)
		deSiempre, raizDeSiempre := DerivarConRaiz(cwd, saltDePrueba)
		if identidad != deSiempre || identidad != Derivar(cwd, saltDePrueba) {
			t.Errorf("%q: identidad %q; Derivar y DerivarConRaiz dan %q: deben ser LA MISMA", cwd, identidad, deSiempre)
		}
		if huboRaiz != raizDeSiempre || huboRaiz != q.huboRaiz {
			t.Errorf("%q: huboRaiz %v; DerivarConRaiz da %v y se esperaba %v", cwd, huboRaiz, raizDeSiempre, q.huboRaiz)
		}
		if origen != q.origen {
			t.Errorf("%q: origen %q; se esperaba %q", cwd, origen, q.origen)
		}
		if event.Ref(saltDePrueba, origen) != identidad {
			t.Errorf("%q: el hash del origen %q no es la identidad: no es la ruta que entró en el hash", cwd, origen)
		}
	}
}

// (3) · DerivarConOrigen, familia a familia.
func TestDerivarConOrigen(t *testing.T) {
	t.Run("subdirectorio", func(t *testing.T) {
		b := baseReal(t)
		crearRepo(t, filepath.Join(b, "proy"))
		crearRepo(t, filepath.Join(b, "otro-proy"))
		crearDirs(t, filepath.Join(b, "proy", "frontend", "src"))
		comprobarFamilia(t, familia(t, "raiz-subdirectorio.jsonl", b), map[string]esperado{
			b + "/proy":              {b + "/proy", true},
			b + "/proy/frontend/src": {b + "/proy", true},
			b + "/otro-proy":         {b + "/otro-proy", true},
		})
	})
	// El árbol paralelo lleva `.git` FICHERO, como un `git worktree`: su raíz es la suya.
	t.Run("worktree", func(t *testing.T) {
		b := baseReal(t)
		crearRepo(t, filepath.Join(b, "proy"))
		crearDirs(t, filepath.Join(b, "paralelo", "sub"))
		if err := os.WriteFile(filepath.Join(b, "paralelo", ".git"), []byte("gitdir: "+b+"/proy/.git/worktrees/paralelo\n"), 0o600); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		comprobarFamilia(t, familia(t, "worktree.jsonl", b), map[string]esperado{
			b + "/proy":         {b + "/proy", true},
			b + "/paralelo":     {b + "/paralelo", true},
			b + "/paralelo/sub": {b + "/paralelo", true},
		})
	})
	// Sin `.git`: el origen es la ubicación LIMPIA, la misma para las cuatro grafías de `suelto`.
	t.Run("fallback", func(t *testing.T) {
		b := baseReal(t)
		crearDirs(t, filepath.Join(b, "suelto"), filepath.Join(b, "sueltoo"))
		comprobarFamilia(t, familia(t, "sintacticos.jsonl", b), map[string]esperado{
			b + "/suelto":      {b + "/suelto", false},
			b + "/suelto/":     {b + "/suelto", false},
			b + "/suelto/./":   {b + "/suelto", false},
			b + "/x/../suelto": {b + "/suelto", false},
			b + "/sueltoo":     {b + "/sueltoo", false},
		})
	})
	// Un enlace: el origen es la ruta REAL, no la del enlace.
	t.Run("enlace", func(t *testing.T) {
		b := baseReal(t)
		crearRepo(t, filepath.Join(b, "proy"))
		crearDirs(t, filepath.Join(b, "proy", "frontend", "src"), filepath.Join(b, "suelto"))
		enlazar(t, filepath.Join(b, "proy", "frontend"), filepath.Join(b, "enlace-proy"))
		enlazar(t, filepath.Join(b, "suelto"), filepath.Join(b, "enlace-suelto"))
		comprobarFamilia(t, familia(t, "enlaces.jsonl", b), map[string]esperado{
			b + "/enlace-proy/src":   {b + "/proy", true},
			b + "/proy/frontend/src": {b + "/proy", true},
			b + "/enlace-suelto":     {b + "/suelto", false},
			b + "/suelto":            {b + "/suelto", false},
		})
	})
	// Lo que no se puede resolver cae al literal limpio; el vacío no tiene ni identidad ni origen.
	t.Run("degradados", func(t *testing.T) {
		b := baseReal(t)
		enlazar(t, filepath.Join(b, "no-esta"), filepath.Join(b, "enlace-roto"))
		crearDirs(t, filepath.Join(b, "sin-permiso", "dentro"))
		if err := os.Chmod(filepath.Join(b, "sin-permiso"), 0o000); err != nil {
			t.Fatalf("precondición: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(filepath.Join(b, "sin-permiso"), 0o700) })
		comprobarFamilia(t, familia(t, "degradados.jsonl", b), map[string]esperado{
			b + "/no-existe-jamas":    {b + "/no-existe-jamas", false},
			b + "/enlace-roto":        {b + "/enlace-roto", false},
			b + "/sin-permiso/dentro": {b + "/sin-permiso/dentro", false},
			"relativa/sin-anclar":     {"relativa/sin-anclar", false},
			"":                        {"", false},
		})
	})
}

// (4) · El resolutor apunta el origen de cada identidad que deriva, por identidad.
func TestResolutorOrigen(t *testing.T) {
	montar := func(t *testing.T) (proy, sub string) {
		b := baseReal(t)
		proy, sub = filepath.Join(b, "proy"), filepath.Join(b, "proy", "frontend", "src")
		crearRepo(t, proy)
		crearDirs(t, sub)
		return proy, sub
	}
	t.Run("tras_derivar", func(t *testing.T) {
		proy, sub := montar(t)
		r := NuevoResolutor()
		identidad := r.Derivar(sub, saltDePrueba)
		if o, ok := r.Origen(identidad); !ok || o != (Origen{Ruta: proy, HuboRaiz: true}) {
			t.Errorf("Origen = (%+v, %v); se esperaba ({Ruta:%s HuboRaiz:true}, true)", o, ok, proy)
		}
	})
	t.Run("acierto_de_cache", func(t *testing.T) {
		proy, sub := montar(t)
		r := NuevoResolutor()
		primera := r.Derivar(sub, saltDePrueba)
		segunda := r.Derivar(sub, saltDePrueba)
		o, ok := r.Origen(segunda)
		if primera != segunda || !ok || o.Ruta != proy {
			t.Errorf("tras un acierto de caché: identidades %q / %q, Origen = (%+v, %v); se esperaba %s", primera, segunda, o, ok, proy)
		}
	})
	// Dos `cwd` de la misma raíz son una identidad y un origen, y siguen siendo dos claves de caché.
	t.Run("dos_cwd_una_raiz", func(t *testing.T) {
		proy, sub := montar(t)
		r := NuevoResolutor()
		desdeSub := r.Derivar(sub, saltDePrueba)
		desdeRaiz := r.Derivar(proy, saltDePrueba)
		o, ok := r.Origen(desdeRaiz)
		if desdeSub != desdeRaiz || !ok || o.Ruta != proy || len(r.cache) != 2 {
			t.Errorf("identidades %q / %q, Origen = (%+v, %v), %d claves de caché; se esperaba una identidad, %s y 2 claves",
				desdeSub, desdeRaiz, o, ok, len(r.cache), proy)
		}
	})
}
