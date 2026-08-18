// Package project deriva la IDENTIDAD DE PROYECTO de un evento a partir del directorio de
// trabajo que el log de la herramienta declara.
//
// ═══ ESTADO ACTUAL: LA DERIVACIÓN COMPLETA (T014, T015, T021). FALTA UNA PIEZA ═════════════
//
// Implementado, en el orden en que se aplica (contrato §Orden de resolución, D-004-2):
//
//	T015  resolución de enlaces PREVIA al ascenso (FR-006a)     — `EvalSymlinks`
//	T021  normalización sintáctica (FR-005/FR-006)              — `filepath.Clean`
//	T014  ascenso con techo hasta la raíz (FR-001/FR-004/FR-004a)
//
// TODAVÍA NO:
//
//	T032  caché de pasada (SC-006)
//
// Las garantías G1..G8 del contrato están cubiertas. **G6 ya es garantía y no efecto lateral**:
// las grafías de una misma ruta convergen tanto si el directorio existe —donde `EvalSymlinks`
// también canonicalizaría— como si no, que es donde antes divergían.
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
	identidad, _ := DerivarConRaiz(cwdDeclarado, salt)
	return identidad
}

// DerivarConRaiz es `Derivar` **más el único dato que `Derivar` tapa**: si el ascenso llegó a
// reconocer una raíz de proyecto, o si la identidad salió del fallback.
//
// ═══ POR QUÉ EXISTE, Y POR QUÉ NO SE RESOLVIÓ HACIENDO FALLAR A `Derivar` ══════════════════
//
// P-005 FR-006 necesita rehusar la adhesión cuando no hay raíz reconocible: unir una identidad
// derivada de un directorio suelto produciría un **éxito perfectamente formado sobre la identidad
// equivocada**, cuyo síntoma es indistinguible de una avería de servidor.
//
// El dato ya existía —`ascender` devuelve `(string, bool)`— y `derivarConTecho` lo usaba y lo
// tiraba. La tentación era que `Derivar` devolviera error al no encontrar raíz, y **eso rompería
// P-004 FR-010 en el camino de la ingesta** (`specs/004-identidad-de-proyecto/spec.md:249`): un lote
// entero se detendría porque un directorio dejó de existir. Por eso P-005 FR-015 exige exponerlo por
// **vía adicional**, y esto es esa vía.
//
// ═══ CUERPO COMPARTIDO, NO DUPLICADO (P-005 FR-005) ════════════════════════════════════════
//
// `Derivar` y `DerivarConRaiz` **no son dos derivaciones que hoy coincidan**: son **la misma**.
// Las dos pasan por `derivarConTechoYRaiz`, y `Derivar` se limita a descartar el segundo valor.
// Esa es la razón de que P-005 SC-001 se pueda demostrar por **origen compartido** en vez de por
// comparación caso a caso: alterar el cuerpo cambia **las dos a la vez**, por construcción.
//
// **NUNCA devuelve error**, igual que `Derivar`: exponer el dato no puede introducir una vía de
// fallo donde no la había (P-005 FR-016).
func DerivarConRaiz(cwdDeclarado, salt string) (identidad string, huboRaiz bool) {
	// ORDEN, y es el requisito (FR-006a / D-004-2): PRIMERO la ubicación real, DESPUÉS el
	// reconocimiento de raíz. Al revés, un enlace que apunta al interior de un proyecto nunca
	// llegaría a reconocerse: su cadena literal ascendería por otro sitio y caería al fallback.
	//
	// El techo se calcula sobre la MISMA ruta real, no sobre la declarada: si un enlace de fuera
	// del directorio personal apuntara adentro (o al revés), un techo calculado sobre la forma
	// literal acotaría un ascenso que ocurre en otro sitio.
	ubicacion := mejorValorDisponible(cwdDeclarado)
	return derivarConTechoYRaiz(ubicacion, salt, techoDeProduccion(ubicacion))
}

// mejorValorDisponible obtiene la mejor forma del directorio declarado (data-model §entidad 4):
// la UBICACIÓN REAL si el sistema permite resolverla, y la forma literal si no — y en los dos
// casos NORMALIZADA.
//
// Es de MEJOR ESFUERZO (FR-009): si `EvalSymlinks` falla —el directorio ya no existe, faltan
// permisos, el enlace está roto, la ruta es relativa— se sigue con lo que hay, y **nunca** se
// devuelve error.
//
// ═══ POR QUÉ EL `Clean` VA AQUÍ Y NO SOLO EN EL FALLBACK ═══════════════════════════════════
//
// La convergencia sintáctica (G6) solo se exige del fallback, así que limpiar únicamente allí
// bastaría para el contrato. **Pero dejaría un agujero en el techo, y es de corrección:** la
// parada del ascenso compara `actual == techo` como cadenas, y `"/home/u/."` NO es igual a
// `"/home/u"` aunque designen el mismo directorio. Con una grafía sucia del propio directorio
// personal, el ascenso no pararía donde debe y **examinaría el home** — violando FR-004a por
// una barra y un punto.
//
// Normalizar una vez, al principio, cierra ese agujero y deja al ascenso y al fallback
// operando sobre la misma forma. Es idempotente y no cambia ningún resultado de las rutas que
// existen, porque `EvalSymlinks` ya devuelve forma limpia.
//
// Lo que NO hace, y son las tres decisiones confirmadas (plan §Puntos, punto 2):
//
//	· NO aplica `filepath.Abs` — anclaría una ruta relativa al cwd del proceso AGENTE, que no
//	  tiene ninguna relación con el del log: sería una identidad inventada (caso 15).
//	· NO expande `~` — exigiría asumir que el `~` del log es el mismo usuario que ejecuta el
//	  agente, falso en cuanto alguien procesa logs de otra cuenta.
//	· NO aplica case-folding — en Linux fusionaría directorios distintos; no aplicarlo separa
//	  dos grafías de un directorio IRRESOLUBLE en un sistema insensible. Ese residuo está
//	  declarado en el contrato §«Lo que este contrato NO promete», y es preferible a decidir
//	  por sistema operativo lo que es propiedad del volumen.
func mejorValorDisponible(declarado string) string {
	if declarado == "" {
		return ""
	}
	if resuelto, err := filepath.EvalSymlinks(declarado); err == nil {
		return filepath.Clean(resuelto)
	}
	return filepath.Clean(declarado)
}

// derivarConTecho es Derivar con el TECHO DEL ASCENSO EXPLÍCITO, y existe por una razón que no
// es de estilo: el techo de la raíz del sistema de ficheros **no se puede falsear por entorno
// sin privilegios**, así que sin esta costura el caso 8 del contrato no tendría forma de
// probarse (data-model §entidad 3, «Forma del techo»). En producción solo la llama `Derivar`,
// con la ubicación real y el techo que le corresponde.
//
// `ubicacion` es el MEJOR VALOR DISPONIBLE (data-model §entidad 4): la ruta real si se pudo
// resolver, y la declarada si no.
//
// Conserva su firma de una sola salida **a propósito**: es la costura que usan los tests del caso 8
// del contrato, y ampliarla los obligaría a cambiar sin que el cambio les aporte nada. El dato de la
// raíz se obtiene de `derivarConTechoYRaiz`, que es donde vive el cuerpo.
func derivarConTecho(ubicacion, salt, techo string) string {
	identidad, _ := derivarConTechoYRaiz(ubicacion, salt, techo)
	return identidad
}

// derivarConTechoYRaiz es **EL CUERPO COMPARTIDO**: la única implementación de la derivación en todo
// el paquete. `Derivar`, `DerivarConRaiz` y `derivarConTecho` son envoltorios suyos.
//
// Que sea uno solo **no es orden, es el mecanismo de P-005 FR-005**: mientras la identidad que el
// comando presenta y la que la ingesta estampa salgan de aquí, **no pueden divergir**. Dos cuerpos que
// hoy devolvieran lo mismo serían dos cuerpos que mañana pueden no hacerlo, y el síntoma —«me he unido
// y no aparece nada»— es indistinguible de una avería de servidor.
func derivarConTechoYRaiz(ubicacion, salt, techo string) (identidad string, huboRaiz bool) {
	// Entrada vacía → identidad ausente (FR-008/G1). No hay guarda propia: `event.Ref` ya
	// devuelve "" para valor vacío (`internal/event/event.go:41-43`), y duplicar aquí la
	// definición de «ausente» crearía dos que podrían divergir.
	if raiz, encontrada := ascender(ubicacion, techo); encontrada {
		return event.Ref(salt, raiz), true
	}

	// Fallback: el mejor valor disponible, ya normalizado por `mejorValorDisponible` (G6).
	//
	// **`huboRaiz` es false, y el valor NO es vacío**: la derivación sigue produciendo identidad para
	// un directorio suelto (P-004 FR-005). Quien decida rehusar por «no hay raíz» debe mirar el
	// booleano, **nunca** si la identidad está vacía — vacía solo lo está con entrada vacía.
	return event.Ref(salt, ubicacion), false
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
		hogarReal := mejorValorDisponible(hogar)
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

// ═══ P-004 T032 · LA CACHÉ DE PASADA ══════════════════════════════════════════════════════

// Resolutor deriva identidades recordando las que ya resolvió DENTRO DE UNA PASADA.
//
// PARA QUÉ (SC-006): un fichero de log típico trae **cientos de eventos del mismo directorio
// de trabajo**, y cada uno dispararía el mismo `EvalSymlinks` y el mismo ascenso. La garantía
// de SC-006 es **por diseño** —una resolución por directorio distinto, no por evento—, y esto
// es ese diseño. No hay benchmark que la sostenga: los tests de tiempo son la familia de
// flakes ya fichada, y un número que oscila no demuestra una propiedad estructural.
//
// LAS TRES DECISIONES, y sus porqués (data-model §entidad 6, research §P3):
//
//	· CLAVE = el directorio DECLARADO, tal cual viene del log. No la ruta real: la caché se
//	  consulta ANTES de resolver, así que usar la ruta real obligaría a resolverla para poder
//	  preguntar — justo el trabajo que la caché existe para evitar.
//	· VALOR = la identidad YA DERIVADA, no la ruta intermedia. Así el acierto se ahorra
//	  también el hash.
//	· ÁMBITO = una PASADA, no el proceso. En `--daemon` el proceso vive días: una caché de
//	  proceso serviría identidad obsoleta a un directorio que entretanto se convirtió en
//	  repositorio. Una pasada es un instante de observación; el proceso no lo es.
//
// LA CACHÉ OPTIMIZA, NO DECIDE: `Derivar` y `Resolutor.Derivar` devuelven exactamente lo
// mismo para la misma entrada. Dos grafías distintas del mismo directorio son dos claves y se
// resuelven por separado —**y convergen igual**, porque quien decide es la derivación.
type Resolutor struct {
	cache map[string]string
}

// NuevoResolutor crea un resolutor con su caché vacía. Se llama UNA VEZ POR PASADA: ahí vive
// el ámbito, y por eso no hay ninguna instancia global ni perezosa que pudiera sobrevivirle.
func NuevoResolutor() *Resolutor {
	return &Resolutor{cache: make(map[string]string)}
}

// Derivar es `Derivar` con memoria de la pasada.
//
// EL RECEPTOR NIL ES VÁLIDO Y DELIBERADO: un `*Resolutor` nil deriva sin caché. Así, quien
// construya un contexto de ingesta sin resolutor —el dry-run de `--scan`, o cualquier test
// que arme su `ingest.Context` a mano— sigue funcionando exactamente igual, solo que sin el
// ahorro. La alternativa —exigir un resolutor siempre— obligaría a tocar cada punto de
// construcción para ganar nada: la caché es una optimización, y una optimización que rompe a
// quien no la pide está mal puesta.
func (r *Resolutor) Derivar(cwdDeclarado, salt string) string {
	if r == nil {
		return Derivar(cwdDeclarado, salt)
	}
	if identidad, visto := r.cache[cwdDeclarado]; visto {
		return identidad
	}
	identidad := Derivar(cwdDeclarado, salt)
	r.cache[cwdDeclarado] = identidad
	return identidad
}
