// Package project deriva la IDENTIDAD DE PROYECTO de un evento a partir del directorio de
// trabajo que el log de la herramienta declara.
//
// ═══ ESTADO ACTUAL: ENLACES + ASCENSO CON TECHO (T014, T015). FALTAN DOS PIEZAS ════════════
//
// Implementado:
//
//	T015  resolución de enlaces PREVIA al ascenso (FR-006a / D-004-2) — `EvalSymlinks`
//	T014  ascenso con techo hasta la raíz del árbol (FR-001/FR-004/FR-004a)
//
// TODAVÍA NO, y por eso el fallback de aquí abajo devuelve el mejor valor disponible SIN
// normalizar:
//
//	T021  normalización del fallback (FR-005/FR-006)  — `filepath.Clean`
//	T032  caché de pasada (SC-006)
//
// Mientras falte T021, dos grafías sintácticas de la misma ruta **que no exista** dan
// identidades distintas. Las que sí existen ya convergen —`EvalSymlinks` las canonicaliza de
// paso—, pero eso es un efecto lateral, no la garantía: G6 se cumple cuando T021 entre.
//
// Lo que SÍ es definitivo es la SUPERFICIE (contrato §Superficie): `Derivar` no devuelve error,
// y no lo devolverá nunca. Esa firma es la garantía estructural de FR-010 —un fallo de
// resolución no puede detener el procesamiento del lote porque no hay por dónde propagarlo—, y
// cambiarla convertiría una garantía que se cumple por construcción en una que habría que
// probar caso por caso.
package project

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/permea-dev/agent/internal/event"
)

// marcador es la entrada que identifica la raíz de un árbol de trabajo.
const marcador = ".git"

// Derivar devuelve la identidad de proyecto de un evento: un valor irreversible derivado de la
// raíz del proyecto que contiene el directorio de lanzamiento —o del propio directorio, si
// ninguno lo contiene—, con el secreto local.
//
// NUNCA devuelve error (contrato §Superficie, G8).
func Derivar(cwdDeclarado, salt string) string {
	// ORDEN, y es el requisito (FR-006a / D-004-2): PRIMERO la ubicación real, DESPUÉS el
	// reconocimiento de raíz. Al revés, un enlace que apunta al interior de un proyecto nunca
	// llegaría a reconocerse: su cadena literal ascendería por otro sitio y caería al fallback.
	//
	// El techo se calcula sobre la MISMA ruta real, no sobre la declarada: si un enlace de fuera
	// del directorio personal apuntara adentro (o al revés), un techo calculado sobre la forma
	// literal acotaría un ascenso que ocurre en otro sitio.
	ubicacion := ubicacionReal(cwdDeclarado)
	return derivarConTecho(ubicacion, salt, techoDeProduccion(ubicacion))
}

// ubicacionReal resuelve los enlaces simbólicos del directorio declarado. Es de MEJOR ESFUERZO
// (FR-009): si el sistema no permite resolver —el directorio ya no existe, faltan permisos, el
// enlace está roto, la ruta es relativa—, devuelve la forma literal y **nunca** un error.
//
// Que la rama de fallo exista desde ya no significa que esté probada: su testigo es el caso 14
// del contrato, que llega con T018. Aquí se declara, no se da por cubierta.
func ubicacionReal(declarado string) string {
	if declarado == "" {
		return ""
	}
	if resuelto, err := filepath.EvalSymlinks(declarado); err == nil {
		return resuelto
	}
	return declarado
}

// derivarConTecho es Derivar con el TECHO DEL ASCENSO EXPLÍCITO, y existe por una razón que no
// es de estilo: el techo de la raíz del sistema de ficheros **no se puede falsear por entorno
// sin privilegios**, así que sin esta costura el caso 8 del contrato no tendría forma de
// probarse (data-model §entidad 3, «Forma del techo»). En producción solo la llama `Derivar`,
// con la ubicación real y el techo que le corresponde.
//
// `ubicacion` es el MEJOR VALOR DISPONIBLE (data-model §entidad 4): la ruta real si se pudo
// resolver, y la declarada si no.
func derivarConTecho(ubicacion, salt, techo string) string {
	// Entrada vacía → identidad ausente (FR-008/G1). No hay guarda propia: `event.Ref` ya
	// devuelve "" para valor vacío (`internal/event/event.go:41-43`), y duplicar aquí la
	// definición de «ausente» crearía dos que podrían divergir.
	if raiz, encontrada := ascender(ubicacion, techo); encontrada {
		return event.Ref(salt, raiz)
	}

	// Fallback: el mejor valor disponible. TODAVÍA SIN `filepath.Clean` — eso es T021.
	return event.Ref(salt, ubicacion)
}

// ascender sube desde `desde` buscando el marcador, y se detiene AL LLEGAR AL TECHO.
//
// ═══ TECHO, NO SALTO ══════════════════════════════════════════════════════════════════════
//
// El techo es EXCLUSIVO: el ascenso examina todos sus descendientes y **a él no lo examina
// nunca**. Ningún marcador se «descarta para seguir subiendo» — si hay un marcador por debajo
// del techo, ese gana, sin excepción. Un `.git` en `~/dev/proy` cuenta con total normalidad
// aunque esté bajo el home; lo único que no cuenta es un `.git` **en el home mismo** o **en la
// raíz misma**, porque el recorrido termina antes de mirar ahí.
//
// Formularlo como techo —y no como «este marcador no cuenta»— elimina la pregunta absurda de
// qué haría el algoritmo DESPUÉS de descartar un marcador del home: no hay después.
func ascender(desde, techo string) (string, bool) {
	if desde == "" {
		return "", false
	}

	actual := desde
	for {
		if actual == techo {
			return "", false // techo alcanzado: no se examina
		}
		if hayMarcador(actual) {
			return actual, true // el PRIMERO encontrado gana → el más cercano (FR-004)
		}
		padre := filepath.Dir(actual)
		if padre == actual {
			return "", false // raíz del sistema de ficheros: no hay más arriba
		}
		actual = padre
	}
}

// hayMarcador indica si `dir` contiene la entrada que marca la raíz de un árbol de trabajo.
//
// ACEPTA FICHERO O DIRECTORIO, y esa disyunción ES el requisito, no un detalle: en el
// repositorio principal `.git` es un directorio, pero en un **árbol de trabajo paralelo** —y en
// un submódulo— es un **fichero**. Comprobar solo «es directorio» haría que los worktrees
// cayeran al fallback y la promesa del contrato (casos 4 y 5) quedara escrita y sin cumplir.
//
// Se usa `os.Lstat` y no `os.Stat` porque aquí solo importa que la entrada EXISTA: seguir un
// enlace para responder eso sería trabajo de más y una decisión que no toca a esta función.
//
// EL CONTENIDO DEL FICHERO `.git` NUNCA SE LEE. Contiene un `gitdir:` que apunta al repositorio
// común, y seguirlo **fusionaría todos los árboles paralelos en una sola identidad** — lo
// contrario exacto de lo que el contrato promete. La identidad sale de la ruta del directorio
// que CONTIENE el marcador, que ya es distinta por árbol (data-model §entidad 3).
func hayMarcador(dir string) bool {
	_, err := os.Lstat(filepath.Join(dir, marcador))
	return err == nil
}

// techoDeProduccion resuelve el techo del ascenso para una ruta: el directorio personal del
// usuario si la ruta cuelga de él, y la raíz del sistema de ficheros en caso contrario
// (FR-004a).
//
// El home se lee por `os.UserHomeDir()` —que respeta HOME en Linux/macOS y USERPROFILE en
// Windows— y esa elección es deliberada: es lo que permite **probar FR-004a por entorno**, con
// un HOME temporal que contiene un `.git`, sin tocar el directorio personal real de nadie.
//
// Si el home no se puede resolver, el techo pasa a ser la raíz del sistema. Falla hacia el lado
// que NO inventa identidades: el peor efecto es reconocer como proyecto un `.git` situado en el
// home, mientras que abortar dejaría al evento sin identidad — y la emisión nunca se interrumpe
// por no poder resolver (FR-009/FR-010).
// El home se compara TAMBIÉN en su forma real: en macOS `/tmp` es un enlace a `/private/tmp`, y
// hay instalaciones donde el directorio personal cuelga de un enlace. Comparar una ruta ya
// resuelta contra un home sin resolver diría «no cuelga del home» para rutas que sí cuelgan, y
// el techo se calcularía mal justo en los sistemas donde más enlaces hay.
func techoDeProduccion(ruta string) string {
	if hogar, err := os.UserHomeDir(); err == nil && hogar != "" {
		hogarReal := ubicacionReal(hogar)
		if cuelgaDe(ruta, hogarReal) {
			return hogarReal
		}
	}
	return raizDelSistema(ruta)
}

// raizDelSistema devuelve el ancestro más alto de `ruta` (en Linux/macOS `/`; en Windows, la
// raíz de la unidad). Se deriva de la propia ruta en vez de fijarse como constante para no
// hardcodear una semántica de sistema operativo (Principio III).
func raizDelSistema(ruta string) string {
	actual := filepath.Clean(ruta)
	for {
		padre := filepath.Dir(actual)
		if padre == actual {
			return actual
		}
		actual = padre
	}
}

// cuelgaDe indica si `ruta` es `base` o desciende de ella. Compara por SEGMENTOS —vía
// `filepath.Rel`— y no por prefijo de cadena: `/home/ana2` NO cuelga de `/home/ana`, y un
// `strings.HasPrefix` diría que sí.
func cuelgaDe(ruta, base string) bool {
	rel, err := filepath.Rel(base, ruta)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
