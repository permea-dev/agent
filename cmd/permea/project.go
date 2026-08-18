package main

import (
	"fmt"
	"io"
	"os"
)

// ═══ P-005 T003 · PUNTO DE EXTENSIÓN DEL COMANDO ═══════════════════════════════════════════
//
// `permea project join` es el PRIMER comando de dos niveles del binario: hasta 005 los dos
// subcomandos —`enroll` y `status`— son planos. La gramática la fija
// `specs/005-adhesion-a-proyecto/contracts/cli.md` §La gramática.
//
// Hoy el verbo `join` **rehúsa siempre**. La implementación llega en P-005 T021 (entrada y rehúses)
// y T027 (presentación de los desenlaces).

// codigoAndamiaje es el código de salida del rehúse provisional de `join`.
//
// ═══ POR QUÉ 70 Y NO 1, QUE ES LO QUE FIJA EL CONTRATO ═════════════════════════════════════
//
// El binario tiene **dos** códigos, `0` y `1` (`contracts/cli.md` §Los códigos de salida), y el
// andamiaje usa **uno que no es ninguno de los dos** a propósito: con `1` —que es lo que el contrato
// fija para el error de uso— los tests de P-005 T017 y T020 **acertarían contra el andamiaje**, que
// también rehúsa y también sale distinto de cero, **sin que nadie hubiera implementado nada**.
//
// **Quién lo retira: P-005 T023**, que compara los ocho códigos exactos del contrato — cualquier
// superviviente del andamiaje lo tumba. La retirada no depende de que nadie se acuerde.
//
// 70 es `EX_SOFTWARE` de la convención `sysexits`: «error interno del programa», que es exactamente
// lo que un verbo sin implementar es.
const codigoAndamiaje = 70

// runProject despacha el SEGUNDO nivel de `permea project <verbo>`.
//
// Devuelve el código de salida en vez de llamar a `os.Exit`: es lo que permite probarlo en proceso,
// sin arrancar un binario hijo. `main` es quien sale.
func runProject(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "error: falta el verbo. Verbos disponibles: join")
		return codigoAndamiaje
	}

	switch args[0] {
	case "join":
		// P-005 T021/T027 implementan esto. Hasta entonces rehúsa, y lo hace por **stderr**:
		// stdout es la respuesta y aquí no hay respuesta que dar (P-005 FR-021).
		fmt.Fprintln(stderr, "error: `permea project join` todavía no está implementado")
		return codigoAndamiaje
	default:
		// NUNCA se intenta interpretar ni corregir el verbo (`contracts/cli.md` §La gramática).
		fmt.Fprintf(stderr, "error: verbo desconocido %q. Verbos disponibles: join\n", args[0])
		return codigoAndamiaje
	}
}

// runProjectOS es el envoltorio que usa `main`: resuelve los canales reales.
func runProjectOS(args []string) int { return runProject(args, os.Stdout, os.Stderr) }
