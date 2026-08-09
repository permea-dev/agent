package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/permea-dev/agent/internal/testutil"
)

// saltDePrueba es el secreto local de estos tests. Es explícito y fijo: la derivación se
// comprueba por IGUALDAD y DESIGUALDAD entre identidades del mismo test, nunca contra
// valores literales esperados — un hash escrito a mano en un test solo prueba que alguien
// lo copió bien.
const saltDePrueba = "salt-de-prueba-de-resolucion"

// ───────────────────────────────────────────────────────────────────────────────────────
// T008 · Garantía G3 · casos 2 y 3 del contrato
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_RaizYSubdirectorioCompartenIdentidad es el núcleo de US1: dos eventos
// originados en cualquier punto del mismo proyecto reciben la MISMA identidad.
func TestDerivar_RaizYSubdirectorioCompartenIdentidad(t *testing.T) {
	base := t.TempDir()
	proyecto := filepath.Join(base, "proy")
	sub := filepath.Join(proyecto, "frontend", "src")
	crearRepo(t, proyecto)
	crearDirs(t, sub)

	desdeRaiz := Derivar(proyecto, saltDePrueba)
	desdeSub := Derivar(sub, saltDePrueba)

	if desdeRaiz != desdeSub {
		t.Errorf("G3/casos 2-3: la raíz y un subdirectorio del MISMO proyecto deben compartir identidad\n"+
			"  raíz         %s → %s\n  subdirectorio %s → %s",
			proyecto, desdeRaiz, sub, desdeSub)
	}

	// Profundidad: un subdirectorio anidado a varios niveles tampoco puede divergir.
	hondo := filepath.Join(proyecto, "a", "b", "c", "d")
	crearDirs(t, hondo)
	if got := Derivar(hondo, saltDePrueba); got != desdeRaiz {
		t.Errorf("G3: un subdirectorio profundo debe compartir la identidad de su raíz\n"+
			"  %s → %s\n  raíz → %s", hondo, got, desdeRaiz)
	}
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T009 · Garantía G4 · caso 6 del contrato
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_ProyectosDistintosNoColisionan es la contra-prueba de T008: sin ella, una
// implementación que devolviera una constante pasaría T008 con nota.
func TestDerivar_ProyectosDistintosNoColisionan(t *testing.T) {
	base := t.TempDir()
	uno := filepath.Join(base, "proy-uno")
	otro := filepath.Join(base, "proy-otro")
	crearRepo(t, uno)
	crearRepo(t, otro)

	t.Run("proyectos hermanos dan identidades distintas", func(t *testing.T) {
		if Derivar(uno, saltDePrueba) == Derivar(otro, saltDePrueba) {
			t.Errorf("G4: dos proyectos distintos NO pueden compartir identidad\n  %s\n  %s", uno, otro)
		}
	})

	t.Run("anidado: gana el proyecto MÁS CERCANO", func(t *testing.T) {
		// Un proyecto dentro de otro. Un evento del interior pertenece al interior.
		interior := filepath.Join(uno, "vendor", "libreria")
		crearRepo(t, interior)
		dentroDelInterior := filepath.Join(interior, "src")
		crearDirs(t, dentroDelInterior)

		identidadInterior := Derivar(interior, saltDePrueba)
		if got := Derivar(dentroDelInterior, saltDePrueba); got != identidadInterior {
			t.Errorf("caso 6: un evento dentro del proyecto interior debe recibir la identidad del interior\n"+
				"  %s → %s\n  interior → %s", dentroDelInterior, got, identidadInterior)
		}
		if identidadInterior == Derivar(uno, saltDePrueba) {
			t.Errorf("caso 6: el proyecto interior NO puede confundirse con el exterior que lo contiene")
		}
	})
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T010 · casos 4 y 5 del contrato
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_ArbolesDeTrabajoParalelos cubre la promesa que se incumple sola si el
// reconocimiento busca «un DIRECTORIO .git»: en un árbol de trabajo paralelo el marcador es
// un FICHERO. La aserción (a) existe para que el test documente por qué existe.
//
// Si `git` no está disponible, este test FALLA con mensaje claro — NUNCA t.Skip: es la única
// cobertura de esta promesa, y saltárselo en silencio la dejaría sin testigo mientras la
// suite sigue diciendo «verde». (La implementación NO necesita git: usa os.Lstat. Solo este
// test lo necesita, para fabricar el árbol.)
func TestDerivar_ArbolesDeTrabajoParalelos(t *testing.T) {
	exigirGit(t)

	base := t.TempDir()
	principal := filepath.Join(base, "proy")
	paralelo := filepath.Join(base, "paralelo")

	crearRepo(t, principal)
	git(t, principal, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "--allow-empty", "-m", "base")
	git(t, principal, "worktree", "add", "-q", paralelo)

	t.Run("(a) el marcador del worktree es un FICHERO, no un directorio", func(t *testing.T) {
		infoPrincipal, err := os.Lstat(filepath.Join(principal, ".git"))
		if err != nil {
			t.Fatalf("no se pudo inspeccionar el marcador del repositorio principal: %v", err)
		}
		if !infoPrincipal.IsDir() {
			t.Errorf("el marcador del repositorio principal debería ser un directorio, es %v", infoPrincipal.Mode())
		}

		infoParalelo, err := os.Lstat(filepath.Join(paralelo, ".git"))
		if err != nil {
			t.Fatalf("no se pudo inspeccionar el marcador del árbol paralelo: %v", err)
		}
		if infoParalelo.IsDir() {
			t.Error("el marcador del árbol paralelo es un DIRECTORIO; se esperaba un FICHERO.\n" +
				"Esta aserción es la razón de ser del test: si el reconocimiento buscara solo directorios,\n" +
				"los árboles paralelos caerían al fallback y la promesa del contrato quedaría incumplida.")
		}
	})

	t.Run("(b) caso 5: árboles paralelos dan identidades distintas", func(t *testing.T) {
		if Derivar(principal, saltDePrueba) == Derivar(paralelo, saltDePrueba) {
			t.Errorf("caso 5: dos árboles de trabajo del mismo repositorio deben tener identidades distintas")
		}
	})

	t.Run("(c) caso 4: dentro del worktree, raíz y subdirectorio comparten identidad", func(t *testing.T) {
		sub := filepath.Join(paralelo, "modulo", "src")
		crearDirs(t, sub)

		raiz := Derivar(paralelo, saltDePrueba)
		if got := Derivar(sub, saltDePrueba); got != raiz {
			t.Errorf("caso 4: el marcador-FICHERO del worktree debe reconocerse como raíz\n"+
				"  %s → %s\n  raíz del worktree → %s\n"+
				"Sin esta aserción, (b) pasaría por el motivo equivocado: dos rutas distintas dan\n"+
				"hashes distintos aunque no se reconozca ningún proyecto.", sub, got, raiz)
		}
	})
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T011 · casos 7 y 7b del contrato · FR-004a, POR ENTORNO
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_TechoDelDirectorioPersonal prueba la exclusión del directorio personal con un
// HOME falso QUE CONTIENE un `.git` — el caso realista que el requisito existe para cubrir
// (quien versiona su configuración personal).
//
// Las cuatro filas son las de quickstart.md §V6. Las dos últimas son las que impiden aprobar
// por el motivo equivocado: una implementación que simplemente ignorase todo lo que cuelga
// del home pasaría las dos primeras y fallaría la tercera.
func TestDerivar_TechoDelDirectorioPersonal(t *testing.T) {
	testutil.Sandbox(t) // HOME/USERPROFILE/XDG_CONFIG_HOME → temporal del test

	hogar, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("no se pudo resolver el HOME del sandbox: %v", err)
	}
	crearRepo(t, hogar) // el HOME es ÉL MISMO un repositorio

	proyecto := filepath.Join(hogar, "dev", "proy")
	crearRepo(t, proyecto)
	sueltoA := filepath.Join(hogar, "suelto-a")
	sueltoB := filepath.Join(hogar, "suelto-b")
	crearDirs(t, sueltoA, sueltoB, filepath.Join(proyecto, "sub"))

	t.Run("fila 1-2: dos sueltos bajo el home tienen identidades PROPIAS y distintas", func(t *testing.T) {
		a, b := Derivar(sueltoA, saltDePrueba), Derivar(sueltoB, saltDePrueba)
		if a == b {
			t.Errorf("caso 7/FR-004a: si coincidieran, el home se estaría tratando como proyecto y\n"+
				"todo lo que cuelga de él caería en un único bucket\n  %s → %s\n  %s → %s",
				sueltoA, a, sueltoB, b)
		}
	})

	t.Run("fila 3 (caso 7b): un marcador BAJO el home SÍ cuenta — techo, no zona prohibida", func(t *testing.T) {
		sub := filepath.Join(proyecto, "sub")
		raiz := Derivar(proyecto, saltDePrueba)
		if got := Derivar(sub, saltDePrueba); got != raiz {
			t.Errorf("caso 7b: un proyecto genuino bajo el home debe reconocerse con total normalidad.\n"+
				"El techo DETIENE el ascenso en el home; no prohíbe los marcadores que hay debajo.\n"+
				"  %s → %s\n  raíz del proyecto → %s", sub, got, raiz)
		}
	})

	t.Run("fila 4: el propio home, siendo el cwd, tampoco es proyecto", func(t *testing.T) {
		// Su identidad debe ser la del directorio normalizado, no una compartida con sus
		// descendientes sin proyecto propio.
		if Derivar(hogar, saltDePrueba) == Derivar(sueltoA, saltDePrueba) {
			t.Errorf("caso 7: el home no se examina ni siendo él mismo el directorio de lanzamiento")
		}
	})
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T012 · caso 8 del contrato · techo de la raíz del sistema de ficheros
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_TechoDeLaRaizDelSistema cubre lo que del caso 8 se puede comprobar sin
// privilegios: que el ascenso TERMINA y cae al fallback cuando no hay ningún marcador en
// toda la cadena hasta la raíz.
//
// La mitad de «un `.git` EN la raíz del sistema no cuenta» no se puede probar creando `/.git`
// —privilegios, y contaminaría la máquina de quien corra la suite—, así que se prueba por la
// COSTURA INYECTABLE que T014 expone: un directorio temporal hace de techo, con un marcador
// dentro, y se comprueba que no se reconoce. No es un stub que se pruebe a sí mismo: es la
// misma función que usa producción, con el techo que en producción vale `/`.
func TestDerivar_TechoDeLaRaizDelSistema(t *testing.T) {
	base := t.TempDir()
	sinMarcador := filepath.Join(base, "a", "b", "c")
	crearDirs(t, sinMarcador)

	// No hay ningún `.git` en toda la cadena: el ascenso debe terminar en el techo y caer al
	// fallback, con una identidad propia y no vacía.
	got := Derivar(sinMarcador, saltDePrueba)
	if got == "" {
		t.Errorf("caso 8: agotado el ascenso sin marcador, la identidad debe venir del directorio\n"+
			"normalizado, nunca quedar vacía\n  %s → %q", sinMarcador, got)
	}

	// Y no puede colapsar con un hermano: terminar el ascenso no es devolver una constante.
	hermano := filepath.Join(base, "a", "b", "otro")
	crearDirs(t, hermano)
	if got == Derivar(hermano, saltDePrueba) {
		t.Errorf("caso 8: dos directorios sin proyecto NO pueden compartir identidad")
	}

	t.Run("un marcador situado EN el techo no se reconoce", func(t *testing.T) {
		// `techo` hace aquí el papel que en producción hace la raíz del sistema de ficheros.
		techo := t.TempDir()
		crearRepo(t, techo) // el marcador está EN el techo, que es lo que no debe contar
		dentro := filepath.Join(techo, "x", "y")
		crearDirs(t, dentro)

		identidadDeDentro := derivarConTecho(dentro, saltDePrueba, techo)
		identidadDelTecho := derivarConTecho(techo, saltDePrueba, techo)

		if identidadDeDentro == identidadDelTecho {
			t.Errorf("caso 8: el marcador situado EN el techo NO se reconoce, así que un descendiente\n"+
				"debe caer al fallback con su propia identidad — no compartirla con el techo.\n"+
				"  %s → %s\n  techo %s → %s", dentro, identidadDeDentro, techo, identidadDelTecho)
		}

		// Contra-prueba: bajo ese mismo techo, un marcador POR DEBAJO sí cuenta. Sin ella, la
		// aserción anterior pasaría también con un ascenso completamente roto.
		proyecto := filepath.Join(techo, "x", "proy")
		crearRepo(t, proyecto)
		sub := filepath.Join(proyecto, "sub")
		crearDirs(t, sub)
		if derivarConTecho(sub, saltDePrueba, techo) != derivarConTecho(proyecto, saltDePrueba, techo) {
			t.Errorf("techo, no zona prohibida: un marcador por debajo del techo debe reconocerse")
		}
	})
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T013 · Garantía G5 · caso 11 del contrato · FR-006a (D-004-2)
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_EnlaceHaciaElInteriorDeUnProyecto comprueba la PRECEDENCIA de la resolución de
// enlaces sobre el reconocimiento de raíz, no solo que enlace y destino converjan.
//
// La aserción que lo demuestra es la de igualdad con la identidad del PROYECTO: si el
// reconocimiento fuera primero, el enlace no llegaría nunca a reconocerse y caería al
// fallback — convergería con su destino solo si ambos cayeran igual, que no es lo mismo.
func TestDerivar_EnlaceHaciaElInteriorDeUnProyecto(t *testing.T) {
	base := t.TempDir()
	proyecto := filepath.Join(base, "proy")
	rutaReal := filepath.Join(proyecto, "frontend")
	crearRepo(t, proyecto)
	crearDirs(t, filepath.Join(rutaReal, "src"))

	enlace := filepath.Join(base, "atajo")
	if err := os.Symlink(rutaReal, enlace); err != nil {
		t.Fatalf("no se pudo crear el enlace simbólico (¿sistema sin symlinks?): %v", err)
	}

	porElEnlace := Derivar(filepath.Join(enlace, "src"), saltDePrueba)
	porLaRutaReal := Derivar(filepath.Join(rutaReal, "src"), saltDePrueba)
	delProyecto := Derivar(proyecto, saltDePrueba)

	if porElEnlace != porLaRutaReal {
		t.Errorf("G5/caso 11: un enlace y su destino deben dar la misma identidad\n"+
			"  por el enlace   → %s\n  por la ruta real → %s", porElEnlace, porLaRutaReal)
	}
	if porElEnlace != delProyecto {
		t.Errorf("FR-006a: la resolución de enlaces PRECEDE al reconocimiento de raíz.\n"+
			"Un enlace que conduce al interior de un proyecto debe recibir la identidad de ESE proyecto.\n"+
			"  por el enlace → %s\n  del proyecto  → %s", porElEnlace, delProyecto)
	}
}

// ───────────────────────────────────────────────────────────────────────────────────────
// Helpers
// ───────────────────────────────────────────────────────────────────────────────────────

// crearDirs crea directorios con t.Fatal si falla: un árbol mal construido invalidaría el
// test en silencio.
func crearDirs(t *testing.T, rutas ...string) {
	t.Helper()
	for _, r := range rutas {
		if err := os.MkdirAll(r, 0o755); err != nil {
			t.Fatalf("no se pudo crear %q: %v", r, err)
		}
	}
}

// crearRepo inicializa un repositorio de git real en `ruta`.
func crearRepo(t *testing.T, ruta string) {
	t.Helper()
	exigirGit(t)
	crearDirs(t, ruta)
	git(t, ruta, "init", "-q")
}

// git ejecuta un comando de git en `dir` y falla el test con su salida si no va bien.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v en %q falló: %v\n%s", args, dir, err, out)
	}
}

// exigirGit falla —NUNCA salta— si git no está disponible. Ver la cabecera de
// TestDerivar_ArbolesDeTrabajoParalelos para el motivo.
func exigirGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Fatalf("git no está disponible en PATH y estos tests lo necesitan para fabricar los árboles.\n"+
			"NO se salta el test: es la única cobertura de la promesa de los árboles paralelos, y\n"+
			"saltarlo dejaría la suite en verde sin haber comprobado nada. Instala git y repite.\n"+
			"(La implementación NO necesita git: reconoce el marcador con os.Lstat.)\n"+
			"error: %v", err)
	}
}
