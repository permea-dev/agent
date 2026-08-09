package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// ═══════════════════════════════════════════════════════════════════════════════════════
// US2 · El fallback normalizado — T016..T020
//
// AVISO DE CONTEXTO, y determina sobre qué rutas trabaja cada test:
//
// `EvalSymlinks` (T015) canonicaliza DE PASO las rutas que EXISTEN — es un efecto lateral
// declarado en la cabecera del paquete, no la garantía. Así que un test de G6 sobre rutas
// existentes mediría ese efecto lateral y pasaría en verde SIN que la normalización
// sintáctica exista. Los casos 9-10 se ejercitan por eso sobre rutas INEXISTENTES, que es
// además donde la garantía tiene contenido propio: es justo cuando el sistema no puede
// resolver nada cuando `filepath.Clean` es lo único que hace converger las grafías.
// ═══════════════════════════════════════════════════════════════════════════════════════

// ───────────────────────────────────────────────────────────────────────────────────────
// T016 · Garantía G6 · casos 9 y 10 del contrato
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_VariantesSintacticasConvergen trabaja sobre una ruta que NO EXISTE: es la
// clase de ruta donde G6 tiene contenido propio (ver el aviso de arriba).
//
// Las variantes se construyen por CONCATENACIÓN DE CADENAS y no con filepath.Join, porque
// Join normaliza al construir — usarlo aquí destruiría el caso antes de probarlo.
func TestDerivar_VariantesSintacticasConvergen(t *testing.T) {
	base := t.TempDir()
	canonica := base + "/no-existe-jamas"

	variantes := []struct {
		nombre string
		ruta   string
	}{
		{"canónica", canonica},
		{"con barra final (caso 9)", canonica + "/"},
		{"con ./ intercalado (caso 10)", base + "/./no-existe-jamas"},
		{"con .. redundante (caso 10)", base + "/otro/../no-existe-jamas"},
		{"con separadores repetidos", base + "//no-existe-jamas"},
	}

	esperada := Derivar(canonica, saltDePrueba)
	for _, v := range variantes {
		if got := Derivar(v.ruta, saltDePrueba); got != esperada {
			t.Errorf("G6/casos 9-10: %s debe converger con la forma canónica\n"+
				"  %s → %s\n  canónica %s → %s",
				v.nombre, v.ruta, got, canonica, esperada)
		}
	}
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T017 · Garantía G7
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_RutasDistintasNoColisionan es la contra-prueba de T016: sin ella, sustituir la
// normalización por una constante dejaría T016 en verde.
//
// NACE VERDE, y por construcción: dos cadenas distintas dan dos hashes distintos aunque no
// haya normalización ninguna. Su testigo real es la mutación de I-9 («Clean por constante»),
// que debe tumbarlo.
//
// El par `suelto` / `sueltoo` es el que caza la fusión por PREFIJO DE CADENA: una
// normalización que comparase con strings.HasPrefix los uniría.
func TestDerivar_RutasDistintasNoColisionan(t *testing.T) {
	base := t.TempDir()

	pares := []struct {
		nombre string
		a, b   string
	}{
		{"hermanos con prefijo común", base + "/suelto", base + "/sueltoo"},
		{"hermanos sin relación", base + "/uno", base + "/dos"},
		{"padre e hijo", base + "/a", base + "/a/b"},
		{"mismo nombre, ramas distintas", base + "/x/comun", base + "/y/comun"},
	}

	for _, p := range pares {
		if Derivar(p.a, saltDePrueba) == Derivar(p.b, saltDePrueba) {
			t.Errorf("G7: %s — rutas genuinamente distintas NO pueden converger\n  %s\n  %s",
				p.nombre, p.a, p.b)
		}
	}
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T018 · Garantía G8 · casos 12, 13 y 14 del contrato
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_MejorEsfuerzoNuncaFalla comprueba que ningún fallo del sistema de ficheros
// deja al evento sin identidad. La firma ya impide devolver error (contrato §Superficie);
// lo que se mide aquí es que la salida sea utilizable en las tres formas de fallo.
//
// NACE VERDE: el best-effort existe desde T015 (`ubicacionReal` devuelve la forma literal
// cuando `EvalSymlinks` falla). El **caso 14** es además el testigo pendiente que I-7 dejó
// declarado para esa rama. Su demostración real son las mutaciones de I-9.
func TestDerivar_MejorEsfuerzoNuncaFalla(t *testing.T) {
	base := t.TempDir()

	t.Run("caso 12: el directorio no existe", func(t *testing.T) {
		exigirIdentidadUtilizable(t, base+"/borrado/hace/tiempo")
	})

	t.Run("caso 14: enlace roto", func(t *testing.T) {
		destino := filepath.Join(base, "destino-efimero")
		crearDirs(t, destino)
		enlace := filepath.Join(base, "enlace-roto")
		if err := os.Symlink(destino, enlace); err != nil {
			t.Fatalf("no se pudo crear el enlace: %v", err)
		}
		if err := os.RemoveAll(destino); err != nil {
			t.Fatalf("no se pudo romper el enlace: %v", err)
		}
		exigirIdentidadUtilizable(t, enlace)
	})

	t.Run("caso 13: permisos denegados", func(t *testing.T) {
		// El t.Skip documentado se permite AQUÍ y se prohíbe en T010, y la asimetría es
		// deliberada (tasks.md:132): aquí el caso de permisos es 1 de los 3 que cubren G8
		// —los otros dos siguen ejerciéndola— y es físicamente inejecutable como root, que
		// puede leerlo todo. En T010 el skip mataría la ÚNICA cobertura de una promesa.
		if os.Geteuid() == 0 {
			t.Skip("ejecutando como root: los permisos no deniegan nada, el caso es inejecutable")
		}
		cerrado := filepath.Join(base, "cerrado")
		dentro := filepath.Join(cerrado, "dentro")
		crearDirs(t, dentro)
		if err := os.Chmod(cerrado, 0o000); err != nil {
			t.Fatalf("no se pudo cerrar el directorio: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(cerrado, 0o755) }) // para que t.TempDir pueda limpiar

		exigirIdentidadUtilizable(t, dentro)
	})
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T019 · caso 15 del contrato
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_RutaRelativaNoSeAncla comprueba que una ruta relativa NO se resuelve contra el
// directorio de trabajo DEL PROCESO AGENTE. Anclarla con filepath.Abs produciría una
// identidad inventada: el cwd del agente no tiene ninguna relación con el del log, que pudo
// escribirse en otra máquina y hace días.
//
// La prueba es por CONSTRUCCIÓN, no por inspección: se deriva la MISMA entrada relativa desde
// dos directorios de trabajo distintos y se exige el mismo resultado. Si alguien añadiera
// Abs, los dos resultados divergirían.
//
// NACE VERDE (no hay Abs en ninguna parte); su testigo real es la mutación de I-9.
//
// Este test cambia el cwd del proceso, así que NO puede marcarse paralelo.
func TestDerivar_RutaRelativaNoSeAncla(t *testing.T) {
	relativa := "proyecto/relativo/src"

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("no se pudo leer el directorio de trabajo: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	derivarDesde := func(dir string) string {
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("no se pudo cambiar a %q: %v", dir, err)
		}
		return Derivar(relativa, saltDePrueba)
	}

	primero := t.TempDir()
	segundo := t.TempDir()
	// En el segundo, la ruta relativa SÍ existe: si algo anclara, la diferencia sería máxima.
	crearDirs(t, filepath.Join(segundo, relativa))

	desdePrimero := derivarDesde(primero)
	desdeSegundo := derivarDesde(segundo)

	if desdePrimero != desdeSegundo {
		t.Errorf("caso 15: una ruta relativa NO puede anclarse al cwd del proceso agente\n"+
			"  desde %s → %s\n  desde %s → %s\n"+
			"Anclar con filepath.Abs produciría una identidad inventada: el cwd del agente no\n"+
			"tiene relación con el del log.", primero, desdePrimero, segundo, desdeSegundo)
	}
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T020 · Garantías G1 y G2
// ───────────────────────────────────────────────────────────────────────────────────────

// TestDerivar_EntradaVaciaYFormaDeSalida cubre las dos garantías de FORMA.
//
// G2 protege FR-018: el receptor no puede distinguir, mirando el valor, si la identidad viene
// de una raíz de proyecto o de un directorio suelto. Por eso se comprueba en TODAS las ramas
// —raíz reconocida, fallback sobre ruta existente y fallback sobre ruta inexistente— y no
// solo en una: una rama que devolviera otra forma delataría su origen.
//
// NACE VERDE: G1 y G2 se cumplen por delegación en `event.Ref`. Su testigo es la mutación de
// I-9.
func TestDerivar_EntradaVaciaYFormaDeSalida(t *testing.T) {
	base := t.TempDir()
	proyecto := filepath.Join(base, "proy")
	crearRepo(t, proyecto)
	existente := filepath.Join(base, "suelto")
	crearDirs(t, existente)

	t.Run("G1: entrada vacía → identidad ausente", func(t *testing.T) {
		if got := Derivar("", saltDePrueba); got != "" {
			t.Errorf("G1/caso 1: sin directorio declarado la identidad debe estar AUSENTE, no inventada: %q", got)
		}
	})

	hexadecimal := regexp.MustCompile(`^[0-9a-f]{64}$`)
	ramas := []struct {
		nombre string
		ruta   string
	}{
		{"rama de raíz reconocida", filepath.Join(proyecto, "sub")},
		{"rama de fallback, ruta existente", existente},
		{"rama de fallback, ruta inexistente", base + "/no-existe"},
	}

	for _, r := range ramas {
		t.Run("G2: "+r.nombre+" → hex-64 minúscula", func(t *testing.T) {
			got := Derivar(r.ruta, saltDePrueba)
			if !hexadecimal.MatchString(got) {
				t.Errorf("G2/FR-018: toda identidad no vacía debe tener la MISMA forma en todas las ramas,\n"+
					"para que el receptor no pueda deducir su origen\n  %s → %q", r.ruta, got)
			}
		})
	}
}

// exigirIdentidadUtilizable comprueba lo que G8 promete de una derivación degradada: que hay
// identidad, que tiene la forma del contrato y que no vino acompañada de ningún error (la
// firma ya lo impide, y por eso esto último se comprueba por AUSENCIA de forma vacía).
func exigirIdentidadUtilizable(t *testing.T, ruta string) {
	t.Helper()

	got := Derivar(ruta, saltDePrueba)
	if got == "" {
		t.Errorf("G8: un fallo del sistema de ficheros NUNCA puede dejar al evento sin identidad\n  %s → vacía", ruta)
		return
	}
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(got) {
		t.Errorf("G8+G2: la identidad degradada debe conservar la forma del contrato\n  %s → %q", ruta, got)
	}
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T032 · La caché de pasada
// ───────────────────────────────────────────────────────────────────────────────────────

// TestResolutor_UnaResolucionPorDirectorioDistinto es la comprobación de SC-006, y se hace
// CONTANDO RESOLUCIONES, no midiendo tiempo: la garantía es estructural —una resolución por
// directorio distinto, no por evento— y un cronómetro no la demostraría, solo la sugeriría
// con un número que oscila.
//
// La unicidad se observa por el TAMAÑO DE LA CACHÉ, leyendo el campo no exportado desde el
// propio paquete. No hace falta ninguna costura de producción para contar: si hubiera que
// añadir un contador solo para el test, se estaría deformando el sujeto para que el test
// pudiera morderlo.
func TestResolutor_UnaResolucionPorDirectorioDistinto(t *testing.T) {
	base := t.TempDir()
	proyecto := filepath.Join(base, "proy")
	crearRepo(t, proyecto)
	sub := filepath.Join(proyecto, "a", "b")
	crearDirs(t, sub)
	otro := filepath.Join(base, "suelto")
	crearDirs(t, otro)

	r := NuevoResolutor()

	// El mismo cwd muchas veces —el caso real: cientos de eventos de un fichero de log— más
	// dos directorios distintos.
	for i := 0; i < 50; i++ {
		r.Derivar(sub, saltDePrueba)
	}
	r.Derivar(proyecto, saltDePrueba)
	r.Derivar(otro, saltDePrueba)

	if len(r.cache) != 3 {
		t.Errorf("SC-006: la caché debe tener UNA entrada por directorio DECLARADO distinto (3), no una\n"+
			"por evento (52); tiene %d", len(r.cache))
	}
}

// TestResolutor_LaCacheOptimizaNoDecide es la mitad que impide que la caché se convierta en
// parte de la derivación: con acierto y sin él, la identidad es la MISMA.
//
// Si faltara, una caché que devolviera cualquier cosa recordada —o que se saltara la
// resolución en el primer uso— pasaría el test de unicidad con nota.
func TestResolutor_LaCacheOptimizaNoDecide(t *testing.T) {
	base := t.TempDir()
	proyecto := filepath.Join(base, "proy")
	crearRepo(t, proyecto)
	sub := filepath.Join(proyecto, "sub")
	crearDirs(t, sub)

	casos := []string{sub, proyecto, filepath.Join(base, "no-existe"), ""}
	r := NuevoResolutor()

	for _, cwd := range casos {
		sinCache := Derivar(cwd, saltDePrueba)
		primera := r.Derivar(cwd, saltDePrueba) // fallo de caché
		segunda := r.Derivar(cwd, saltDePrueba) // acierto de caché

		if primera != sinCache || segunda != sinCache {
			t.Errorf("la caché OPTIMIZA, no DECIDE: para %q debe dar lo mismo que Derivar\n"+
				"  sin caché → %s\n  1.ª (fallo) → %s\n  2.ª (acierto) → %s",
				cwd, sinCache, primera, segunda)
		}
	}
}

// TestResolutor_NilDerivaSinCache protege el contrato del receptor nil: quien construya un
// contexto de ingesta sin resolutor —el dry-run de `--scan`, o un test que arme su Context a
// mano— sigue derivando igual. Sin esta garantía, añadir la caché habría roto a todo el que
// no la pidió.
func TestResolutor_NilDerivaSinCache(t *testing.T) {
	base := t.TempDir()
	proyecto := filepath.Join(base, "proy")
	crearRepo(t, proyecto)
	sub := filepath.Join(proyecto, "sub")
	crearDirs(t, sub)

	var nulo *Resolutor
	if got, want := nulo.Derivar(sub, saltDePrueba), Derivar(sub, saltDePrueba); got != want {
		t.Errorf("un *Resolutor nil debe derivar sin caché y dar el mismo resultado\n  nil → %s\n  Derivar → %s", got, want)
	}
}

// TestResolutor_CadaUnoTieneSuPropiaCache: dos resolutores no comparten memoria. Es lo que
// hace que el ámbito sea la PASADA y no el proceso — si compartieran, `tick()` heredaría lo
// resuelto en el ciclo anterior y un directorio que cambiara de naturaleza no se enteraría
// hasta reiniciar el daemon.
func TestResolutor_CadaUnoTieneSuPropiaCache(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "x")
	crearDirs(t, dir)

	primero := NuevoResolutor()
	primero.Derivar(dir, saltDePrueba)

	segundo := NuevoResolutor()
	if len(segundo.cache) != 0 {
		t.Errorf("un resolutor nuevo debe nacer sin memoria del anterior; tiene %d entradas", len(segundo.cache))
	}
}
