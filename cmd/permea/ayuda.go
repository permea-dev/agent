package main

import (
	"fmt"
	"io"
	"strings"
)

// ═══ P-006 · LA FUENTE ÚNICA DEL TEXTO DE AYUDA ════════════════════════════════════════════
//
// Hasta P-006 la ayuda eran dos textos que nadie obligaba a coincidir: el literal de `printUsage` y
// el uso por defecto del paquete `flag`, que no listaba ningún subcomando. Ahora hay UNA tabla, y de
// ella se componen la ayuda general y la de cada subcomando. La de un subcomando es un FRAGMENTO
// LITERAL de la general: las mismas filas, con la misma sangría (`specs/006-medicion-fiel/contracts/cli.md`).
//
// ⛔ Un subcomando nuevo **no aparece solo**: hay que añadir su fila aquí. Lo exige
// `005/contracts/cli.md` §La gramática —«la ayuda del binario DEBE listar `project join` junto a
// `enroll` y `status`»— porque un comando que no aparece en la ayuda no existe para quien lo busca.
//
// La vía stdin se documenta como RECOMENDADA en los dos que llevan un secreto en el argumento
// —`enroll` y `project join`—: por argumento, el valor queda en el historial del intérprete de
// órdenes y a la vista de quien pueda enumerar procesos (`003/contracts/cli.md`,
// `005/contracts/cli.md` §Entrada).

// ═══ D-006-14 · EL TEXTO LO APROBÓ EL DUEÑO, LITERALMENTE (2026-10-02) ═════════════════════
//
// Tres garantías, vigiladas por `ayuda_test.go`: ninguna línea de ninguna ayuda pasa de 80
// caracteres (runas); la general trae «Primeros pasos» y el aviso de que la primera pasada envía todo
// el historial que conserve Claude Code; y ninguna ayuda contiene jerga interna (identificadores de
// especificación, nombres de ajustes de la configuración). Cambiar una línea de este fichero es
// cambiar un texto aprobado.
//
// Forma: la sinopsis va SOLA en su línea (sangría 2), la descripción debajo (sangría 6) y los
// ejemplos debajo de la descripción (sangría 10).

// filaDeAyuda es una entrada de la tabla: la sinopsis, su descripción y sus ejemplos.
type filaDeAyuda struct {
	sinopsis    string
	descripcion []string
	ejemplos    []string
}

// introduccion es lo primero que se lee: qué es Permea, en una frase.
var introduccion = []string{
	"Permea mide el uso de Claude Code en este ordenador y lo envía a tu",
	"organización.",
}

// primerosPasos es la secuencia para quien acaba de instalar (D-006-14).
var primerosPasos = []string{
	"  1. permea enroll ...   pega el comando que te da la aplicación al añadir",
	"                         un agente",
	"  2. permea status       comprueba que ha quedado conectado",
	"  3. permea --run        mide y envía una vez; --daemon lo deja en marcha",
}

// filasDeSubcomandos son los subcomandos, en el orden en que se usan.
var filasDeSubcomandos = []filaDeAyuda{
	{
		sinopsis: "enroll [<enrollment-string>]",
		descripcion: []string{
			"Conecta este ordenador con tu organización: verifica el código y",
			"guarda el token. Recomendado: pásalo por stdin, para que el secreto",
			"no quede en el historial del shell:",
		},
		ejemplos: []string{`echo "$ENROLL" | permea enroll -`},
	},
	{
		sinopsis: "status",
		descripcion: []string{
			"Dice si este ordenador está conectado y con qué servidor. Nunca",
			"muestra el token.",
		},
	},
	{
		sinopsis: "project join [<código>]",
		descripcion: []string{
			"Une esta instalación a un Proyecto, para que su consumo (también el",
			"ya medido) cuente bajo él. Se ejecuta dentro de la carpeta de trabajo",
			"que quieres agrupar. El código lo genera quien administra la",
			"organización, desde el panel. Repetirlo no cambia nada.",
			"Recomendado: pásalo por stdin, para que no quede en el historial:",
		},
		ejemplos: []string{`echo "$CODIGO" | permea project join -`},
	},
}

// filasDeOpciones son las opciones y la de ayuda.
var filasDeOpciones = []filaDeAyuda{
	{sinopsis: "--scan <fichero>", descripcion: []string{
		"Prueba en seco: lee un registro (.jsonl) y muestra un evento por",
		"mensaje. No guarda ni envía nada.",
	}},
	{sinopsis: "--run", descripcion: []string{
		"Una pasada: mide el uso nuevo y lo envía. La primera vez envía todo",
		"el historial que conserve Claude Code.",
	}},
	{sinopsis: "--daemon", descripcion: []string{
		"Se queda en marcha: mide y envía cada cierto tiempo.",
	}},
	{sinopsis: "--version", descripcion: []string{
		"Muestra la versión.",
	}},
	{sinopsis: "-h, --help, help", descripcion: []string{
		"Muestra esta ayuda. La de un subcomando: permea <subcomando> -h",
	}},
}

// ayudasDeSubcomando son las cabeceras `uso:` y las filas de cada ayuda de subcomando. Cada una sale
// de la MISMA tabla que la general, con la misma forma, así que es un fragmento literal suyo.
var ayudasDeSubcomando = map[string]struct {
	uso     string
	titulo  string
	indices []int
}{
	"enroll":       {"uso: permea enroll [<enrollment-string>]", "", []int{0}},
	"status":       {"uso: permea status", "", []int{1}},
	"project":      {"uso: permea project <verbo>", "Verbos:", []int{2}},
	"project join": {"uso: permea project join [<código>]", "", []int{2}},
}

// escribirLineas escribe cada línea con la sangría dada.
func escribirLineas(w io.Writer, sangria int, lineas []string) {
	prefijo := strings.Repeat(" ", sangria)
	for _, l := range lineas {
		_, _ = fmt.Fprintf(w, "%s%s\n", prefijo, l)
	}
}

// escribirFila escribe una fila: sinopsis (sangría 2), descripción (6) y ejemplos (10). La misma
// función sirve a la ayuda general y a la de cada subcomando: por eso una es fragmento literal de la
// otra.
func escribirFila(w io.Writer, f filaDeAyuda) {
	escribirLineas(w, 2, []string{f.sinopsis})
	escribirLineas(w, 6, f.descripcion)
	escribirLineas(w, 10, f.ejemplos)
}

// escribirAyudaGeneral escribe la ayuda completa aprobada (D-006-14). Un fallo al escribir la ayuda
// no es recuperable ni afecta a ningún dato, así que se ignora explícitamente.
func escribirAyudaGeneral(w io.Writer) {
	escribirLineas(w, 0, introduccion)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "uso: permea <subcomando | opción>")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Primeros pasos:")
	escribirLineas(w, 0, primerosPasos)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Subcomandos:")
	for _, f := range filasDeSubcomandos {
		escribirFila(w, f)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Opciones:")
	for _, f := range filasDeOpciones {
		escribirFila(w, f)
	}
}

// escribirAyudaDe escribe la ayuda de un subcomando («enroll», «status», «project», «project join»).
// Devuelve false si el nombre no tiene ayuda, lo que sería un error de programación.
func escribirAyudaDe(w io.Writer, nombre string) bool {
	a, ok := ayudasDeSubcomando[nombre]
	if !ok {
		return false
	}
	_, _ = fmt.Fprintln(w, a.uso)
	_, _ = fmt.Fprintln(w)
	if a.titulo != "" {
		_, _ = fmt.Fprintln(w, a.titulo)
	}
	for _, i := range a.indices {
		escribirFila(w, filasDeSubcomandos[i])
	}
	return true
}

// esPeticionDeAyudaGeneral reconoce las formas de pedir la ayuda general en el PRIMER argumento.
func esPeticionDeAyudaGeneral(arg string) bool {
	switch arg {
	case "help", "-h", "--help", "-help":
		return true
	}
	return false
}

// esPeticionDeAyudaDeSubcomando reconoce `-h` y `--help` como PRIMER argumento de un subcomando.
// Ningún enrollment string (`pmea2.`) ni código de adhesión (`pmeaj1.`) válido empieza por `-h`.
func esPeticionDeAyudaDeSubcomando(args []string) bool {
	return len(args) > 0 && (args[0] == "-h" || args[0] == "--help")
}

// prefijosDeSecreto son las formas conocidas de un secreto tecleado: enrollment strings (`pmea2.`,
// y el `pmea1.` ya rechazado, que sigue siendo un secreto) y códigos de adhesión (`pmeaj1.`).
var prefijosDeSecreto = []string{"pmea2.", "pmeaj1.", "pmea1."}

// pareceSecreto dice si un texto tecleado contiene la forma de un secreto conocido. Si la tiene, los
// mensajes de error NUNCA lo reproducen (P-006 FR-022, FR-024).
func pareceSecreto(s string) bool {
	for _, p := range prefijosDeSecreto {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
