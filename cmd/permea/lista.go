package main

import (
	"sort"
	"time"

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
