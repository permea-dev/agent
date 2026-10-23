package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
)

// ═══ 013 · `permea project list` ════════════════════════════════════════════════════════════════════════════════════
//
// Enseña, por identidad, la carpeta, las herramientas, los eventos y las fechas, para cruzarlos con el panel. Es la emisión con
// otro destino: el mismo recorrido de los seis lectores *(`recorrer`, FR-009)*, con un sumidero que agrega en memoria y un estado
// que nunca se guarda. NUNCA abre la red ni escribe nada *(D-6)*. Es el primer comando que imprime rutas reales *(D-10)*: salen a
// la terminal de quien lo ejecuta, y a ningún otro sitio.

// Las dos marcas de la carpeta, letra a letra *(013 D-3, spec §Textos)*.
const (
	marcaSinRepositorio = "(sin repositorio)"
	marcaYaNoExiste     = "(ya no existe)"
)

// fila es una identidad de la lista *(013 FR-011 a FR-016)*. `ruta` es la que entró en su hash *(FR-015)*; `herramientas`, ya
// con su nombre comercial y en el orden de P-3; `desde` y `hasta`, en UTC *(FR-013)*.
type fila struct {
	identidad    string
	ruta         string
	marca        string
	herramientas []string
	eventos      int
	desde, hasta time.Time
}

// nombresComerciales son los de las herramientas, en el orden en que `recorrer` las lee *(013 D-1, P-3)*. Es copia del panel
// *(permea-platform `frontend/src/catalogo/es/Operador.ts`)* más `copilot_cli`, que el panel aún no nombra. Vive aquí porque es
// presentación, y sólo la usa este comando: los lectores emiten el id, y la frontera no se toca *(plan §5)*.
var nombresComerciales = []struct{ id, nombre string }{
	{"claude_code", "Claude Code"},
	{"codex", "Codex"},
	{"gemini", "Gemini CLI"},
	{"opencode", "OpenCode"},
	{"copilot_cli", "GitHub Copilot CLI"},
	{"kimi_code", "Kimi Code CLI"},
}

// acumulado es lo que `agregar` lleva de una identidad mientras recorre los eventos.
type acumulado struct {
	eventos      int
	tools        map[string]bool
	desde, hasta time.Time
}

// agregar reúne los eventos por identidad *(013 plan §4)*. Es pura: el disco entra por `origen` *(la ruta que entró en el hash,
// del `Resolutor`)* y por `existe`, para probarla sin él.
//   - Un evento sin identidad no cuenta *(P-1)*.
//   - Se cuentan `event_id` distintos, y el primero que llega gana: un repetido no aporta nada, ni a su identidad ni a otra
//     *(P-2)*. Una identidad hecha sólo de repetidos no sale.
//   - Las fechas, en UTC *(FR-013)*.
//   - Una sola marca *(P-4)*: sin la carpeta, «(ya no existe)», haya habido raíz o no; con ella y sin raíz, «(sin repositorio)».
//   - El orden, por eventos de más a menos y, si empatan, por la identidad *(D-2, FR-020)*.
func agregar(evs []event.Event, origen func(ref string) (project.Origen, bool), existe func(ruta string) bool) []fila {
	vistos := map[string]bool{}
	porRef := map[string]*acumulado{}
	for _, ev := range evs {
		if ev.ProjectRef == "" || vistos[ev.EventID] {
			continue
		}
		vistos[ev.EventID] = true
		ac := porRef[ev.ProjectRef]
		if ac == nil {
			ac = &acumulado{tools: map[string]bool{}}
			porRef[ev.ProjectRef] = ac
		}
		cuando := ev.OccurredAt.UTC()
		if ac.eventos == 0 || cuando.Before(ac.desde) {
			ac.desde = cuando
		}
		if ac.eventos == 0 || cuando.After(ac.hasta) {
			ac.hasta = cuando
		}
		ac.eventos++
		ac.tools[ev.Tool] = true
	}

	filas := make([]fila, 0, len(porRef))
	for ref, ac := range porRef {
		f := fila{identidad: ref, herramientas: conNombre(ac.tools), eventos: ac.eventos, desde: ac.desde, hasta: ac.hasta}
		if o, ok := origen(ref); ok {
			f.ruta = o.Ruta
			switch {
			case !existe(o.Ruta):
				f.marca = marcaYaNoExiste
			case !o.HuboRaiz:
				f.marca = marcaSinRepositorio
			}
		}
		filas = append(filas, f)
	}
	sort.Slice(filas, func(i, j int) bool {
		if filas[i].eventos != filas[j].eventos {
			return filas[i].eventos > filas[j].eventos
		}
		return filas[i].identidad < filas[j].identidad
	})
	return filas
}

// conNombre pasa los ids de herramienta a su nombre comercial, en el orden de `nombresComerciales`; los desconocidos van detrás,
// tal cual y por orden alfabético *(D-1, P-3)*.
func conNombre(tools map[string]bool) []string {
	var nombres []string
	for _, n := range nombresComerciales {
		if tools[n.id] {
			nombres = append(nombres, n.nombre)
		}
	}
	var desconocidos []string
	for id := range tools {
		if !conocido(id) {
			desconocidos = append(desconocidos, id)
		}
	}
	sort.Strings(desconocidos)
	return append(nombres, desconocidos...)
}

func conocido(id string) bool {
	for _, n := range nombresComerciales {
		if n.id == id {
			return true
		}
	}
	return false
}

// ═══ 013 B6 · EL VERBO ══════════════════════════════════════════════════════════════════════════════════════════════

// Los textos de `project list`, letra a letra *(spec §Textos; los errores, aprobados por el dueño el 23-10)*. Generados desde la spec.
const (
	textoSinSal         = "Esta instalación aún no ha medido nada."
	textoSinArgumentos  = "error: project list no admite argumentos. Ayuda: permea project list -h"
	textoConfigIlegible = "error: la configuración local no se puede leer: %v"
	textoSalIlegible    = "error: la sal de esta instalación no se puede leer: %v"
)

// projectList es `permea project list` *(013 FR-001 a FR-008)*: sólo lee, y NUNCA crea nada *(D-6)*. Por eso no pasa por `setup()`,
// que crea la sal y el `machine_id`: el `DataDir` y la sal se leen sin crearlos *(`RutaDataDir`, `LeerSalt`)*, y la configuración,
// sólo por `logs_root`.
//   - `-h` y `--help` dan la ayuda; cualquier otro argumento, el error de uso, sin reproducirlo *(P-7)*.
//   - Sin sal, no hay nada medido: lo dice y sale con 0 *(D-5, P-11)*.
//   - `project_ref_mode: "plain"` ni detiene ni avisa *(P-8)*: la lista no emite nada.
//
// 013 B6: la tubería *(recorrer → agregar → formatear)* la conecta B7. Hasta entonces, con sal, sale con 0 sin escribir nada.
func projectList(args []string, stdout, stderr io.Writer) int {
	if esPeticionDeAyudaDeSubcomando(args) {
		escribirAyudaDe(stdout, "project list")
		return codigoExito
	}
	if len(args) > 0 {
		_, _ = fmt.Fprintln(stderr, textoSinArgumentos)
		return codigoFallo
	}
	dir, err := config.RutaDataDir()
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "error:", err)
		return codigoFallo
	}
	_, existe, err := config.LeerSalt(dir)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, textoSalIlegible+"\n", err)
		return codigoFallo
	}
	if !existe {
		_, _ = fmt.Fprintln(stdout, textoSinSal)
		return codigoExito
	}
	if _, err := config.Load(filepath.Join(dir, "config.json")); err != nil {
		_, _ = fmt.Fprintf(stderr, textoConfigIlegible+"\n", err)
		return codigoFallo
	}
	return codigoExito
}

// ═══ 013 B5 · LA TABLA ══════════════════════════════════════════════════════════════════════════════════════════════

// Las columnas y el pie, letra a letra *(013 D-2, D-4; spec §Textos)*.
var cabeceraDeLaLista = []string{"IDENTIDAD", "CARPETA", "HERRAMIENTAS", "EVENTOS", "DESDE", "HASTA"}

const (
	columnaEventos  = 3 // la única a la derecha *(FR-021)*
	pieFechas       = "%s, %s. Fechas en UTC, como en el panel."
	pieDisco        = "Cuenta lo que las herramientas conservan en disco, también lo aún no enviado: el panel puede tener más o menos."
	formatoDeLaFila = "2006-01-02"
)

// formatear escribe la tabla y su pie *(013 FR-017 a FR-022, spec §Textos)*. Es pura: las filas ya vienen ordenadas
// *(`agregar`)*, y `hogar` es el real *(`project.HogarReal`)*, o vacío si no se pudo resolver.
//   - Cada columna mide lo que su celda más ancha o su cabecera, **en runas** *(P-12)*, con dos espacios entre columnas.
//   - `EVENTOS` va a la derecha, también su cabecera; las demás, a la izquierda, y la última sin relleno: ninguna línea acaba
//     en espacio.
//   - Tras una línea en blanco, el pie. Sin filas, **sólo** el pie *(P-5)*.
func formatear(filas []fila, hogar string) string {
	var b strings.Builder
	if len(filas) > 0 {
		celdas := [][]string{cabeceraDeLaLista}
		for _, f := range filas {
			carpeta := aCasa(f.ruta, hogar)
			if f.marca != "" {
				carpeta += " " + f.marca
			}
			celdas = append(celdas, []string{abreviar(f.identidad), carpeta, strings.Join(f.herramientas, ", "), miles(f.eventos),
				f.desde.Format(formatoDeLaFila), f.hasta.Format(formatoDeLaFila)})
		}
		anchos := make([]int, len(cabeceraDeLaLista))
		for _, fila := range celdas {
			for i, c := range fila {
				anchos[i] = max(anchos[i], ancho(c))
			}
		}
		for _, fila := range celdas {
			for i, c := range fila {
				relleno := strings.Repeat(" ", anchos[i]-ancho(c))
				switch {
				case i == columnaEventos:
					b.WriteString(relleno + c)
				case i == len(fila)-1:
					b.WriteString(c)
				default:
					b.WriteString(c + relleno)
				}
				if i < len(fila)-1 {
					b.WriteString("  ")
				}
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	eventos := 0
	for _, f := range filas {
		eventos += f.eventos
	}
	b.WriteString(fmt.Sprintf(pieFechas, contar(len(filas), "identidad", "identidades"), contar(eventos, "evento", "eventos")) + "\n")
	b.WriteString(pieDisco + "\n")
	return b.String()
}

// ancho es el ancho de una celda en runas *(P-12)*: una tilde o el `…` de la identidad cuentan uno.
func ancho(s string) int { return utf8.RuneCountInString(s) }

// abreviar deja los 12 primeros caracteres de la identidad y `…`, como el panel *(FR-018)*.
func abreviar(identidad string) string {
	if len(identidad) <= 12 {
		return identidad
	}
	return identidad[:12] + "…"
}

// contar escribe `n` con punto de miles y la palabra en singular o en plural *(D-4: «1 identidad», «0 identidades»)*.
func contar(n int, singular, plural string) string {
	if n == 1 {
		return "1 " + singular
	}
	return miles(n) + " " + plural
}

// miles escribe `n` con un punto cada tres cifras desde la derecha *(013 FR-019)*: «999», «1.240», «1.234.567».
func miles(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0 && s[i-1] != '-'; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return s
}

// aCasa escribe la carpeta personal como `~` *(013 P-10)*: `~` si la ruta es el hogar, `~` + separador + el resto si cuelga de él
// *(por segmentos, `project.CuelgaDe`)*, y la ruta tal cual si no, o si no hay hogar.
func aCasa(ruta, hogar string) string {
	if hogar == "" || !project.CuelgaDe(ruta, hogar) {
		return ruta
	}
	rel, err := filepath.Rel(hogar, ruta)
	if err != nil {
		return ruta
	}
	if rel == "." {
		return "~"
	}
	return "~" + string(filepath.Separator) + rel
}
