// Package project deriva la IDENTIDAD DE PROYECTO de un evento a partir del directorio de
// trabajo que el log de la herramienta declara.
//
// ═══ ESTADO ACTUAL: PUNTO DE EXTENSIÓN NEUTRO — NO ES LA IMPLEMENTACIÓN ═════════════════════
//
// Lo que hay hoy aquí **reproduce exactamente el comportamiento anterior a P-004**: hashear el
// directorio declarado tal cual llega. No observa el sistema de ficheros, no reconoce raíces de
// proyecto, no normaliza y no resuelve enlaces. **Ninguna de las garantías G1..G10 del contrato
// está implementada todavía**, salvo las dos que salen gratis de `event.Ref` (G1 y G2).
//
// Eso es DELIBERADO y es el propósito de este fichero: existir **antes** que la resolución para
// que los tests de US1/US2 puedan nacer en ROJO por la razón correcta —«la identidad no agrupa»—
// en lugar de por «la función no existe». Un test que falla porque no compila no demuestra nada
// sobre el comportamiento.
//
// La resolución real llega después, y en este orden (contrato §Orden de resolución, D-004-2):
//
//	T014  ascenso con techo hasta la raíz del árbol de trabajo   (FR-001/FR-004/FR-004a)
//	T015  resolución de enlaces PREVIA al ascenso                 (FR-006a)
//	T021  fallback normalizado                                    (FR-005/FR-006)
//	T032  caché de pasada                                         (SC-006)
//
// **Que nadie lea esta neutralidad como una implementación terminada.** Mientras este comentario
// siga aquí, el paquete no cumple el contrato: solo ocupa su sitio.
//
// Lo que sí es definitivo desde ya es la SUPERFICIE (contrato §Superficie): `Derivar` no devuelve
// error, y no lo devolverá nunca. Esa firma es la garantía estructural de FR-010 —un fallo de
// resolución no puede detener el procesamiento del lote porque no hay por dónde propagarlo—, y
// cambiarla convertiría una garantía que se cumple por construcción en una que habría que probar
// caso por caso.
package project

import "github.com/permea-dev/agent/internal/event"

// Derivar devuelve la identidad de proyecto de un evento: un valor irreversible derivado del
// directorio de lanzamiento con el secreto local, o la cadena vacía si no hay directorio.
//
// NUNCA devuelve error (contrato §Superficie, G8).
//
// IMPLEMENTACIÓN NEUTRA (ver la cabecera del paquete): hoy delega en `event.Ref` sobre el
// directorio declarado, sin tocar el sistema de ficheros. Es literalmente lo que hacía
// `internal/ingest/claudecode.go` antes de P-004, movido de sitio y nada más.
//
// **No lleva guarda propia para la entrada vacía, y es a propósito**: `event.Ref` ya devuelve ""
// para valor vacío (`internal/event/event.go:41-43`), así que G1 se cumple por delegación. Añadir
// aquí un `if cwdDeclarado == "" { return "" }` sería una segunda definición de «ausente» que
// podría divergir de la primera el día que una de las dos cambiara — y el sitio donde vive esa
// decisión es `event.Ref`, que la comparte con las identidades de sesión y de máquina.
func Derivar(cwdDeclarado, salt string) string {
	return event.Ref(salt, cwdDeclarado)
}
