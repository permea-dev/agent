package ingest

import (
	"encoding/hex"
	"fmt"

	"github.com/permea-dev/agent/internal/event"
)

// ═══ P-006 · LA PASADA — UN MENSAJE, UN EVENTO ═══════════════════════════════════════════
//
// Claude Code escribe VARIAS líneas por mensaje. Dentro de una pasada, las líneas de un mensaje producen
// UN evento: las siguientes a la primera se cuentan como repetidas y, si su consumo difiere del de la
// primera, también como discrepancia. Las líneas NUNCA se suman (P-006 FR-005).
//
// ═══ P-007 · CADA PARTIDA VALE SU MÁXIMO, Y EL EVENTO SALE AL CERRAR EL FICHERO ════════════════
//
// La salida puede crecer de una línea a la siguiente (7 → 1 303 → 89 817), así que «la primera manda»
// se queda corta (P-007 FR-009). La pasada ACUMULA cada mensaje —el máximo de cada partida, y el desglose
// de la escritura de caché de la línea que da su máximo, la última si empatan— y lo emite al cerrar el
// fichero (`CerrarFichero`). El evento toma de la primera línea su momento, su modelo y sus referencias
// (`plan.md` D-007-P5). Sin pasada no se acumula: cada línea se emite como hasta ahora.
//
// ═══ SU ÁMBITO ES UNA PASADA, Y POR ESO VIVE EN MEMORIA ════════════════════════════════════
//
// Una ejecución de `--run`, un ciclo de `--daemon` o un `--scan` de un fichero. Entre pasadas no hay
// memoria (P-006 FR-033): un mensaje que reaparece se vuelve a encolar con el MISMO `event_id`, y la
// plataforma lo descarta por `(org_id, event_id)`. Guardarlo en disco sería un estado nuevo que puede
// corromperse o perderse, para resolver algo que la plataforma ya resuelve.
//
// ═══ NUNCA GUARDA NI ESCRIBE UN IDENTIFICADOR DEL PROVEEDOR ════════════════════════════════
//
// La clave del conjunto es el `event_id` ya derivado, en sus 16 bytes, no en la cadena hex
// (`research.md` R3.4). El resumen sólo lleva recuentos: dice CUÁNTO, nunca QUIÉN.

// consumo son las cuatro partidas de tokens de una línea y el reparto de su escritura de caché por duración
// (P-007 FR-002), ya resuelto por `desglosarEscritura`.
type consumo struct {
	entrada, salida, escrituraCache, lecturaCache int
	escritura5m, escritura1h                      int
}

// partidas son las cuatro partidas de tokens de un consumo, sin el desglose.
func (c consumo) partidas() [4]int {
	return [4]int{c.entrada, c.salida, c.escrituraCache, c.lecturaCache}
}

// mensaje es lo que la pasada sabe de un mensaje: su evento base, de la primera línea; el consumo de esa
// primera línea; el máximo de cada partida, con el desglose de la línea del máximo de la escritura; y si ya
// salió.
type mensaje struct {
	base    event.Event
	primera consumo
	maximo  consumo
	emitido bool
}

// Pasada es el estado de UNA lectura de los logs. El cero no es utilizable: se crea con NuevaPasada.
// Un *Pasada nil es válido en todos sus métodos, que no hacen nada.
type Pasada struct {
	mensajes  map[[bytesEventID]byte]*mensaje
	abiertos  [][bytesEventID]byte // los del fichero en curso, en el orden en que aparecieron
	recuentos Recuentos
}

// Recuentos son los contadores de una pasada (`data-model.md`, §La pasada).
type Recuentos struct {
	Facturables      int // líneas `assistant` con modelo leídas
	Emitidos         int // eventos producidos
	Repetidas        int // líneas de un mensaje ya emitido en la pasada
	Sinteticas       int // líneas `<synthetic>` descartadas (P-006 FR-007)
	SinIdentificador int // líneas sin `message.id` ni `requestId`, no emitidas (P-006 FR-006)
	ConsumoDistinto  int // repetidas cuyo consumo difiere del de la primera (P-006 FR-005)
	SinDesglose      int // líneas sin desglose de la escritura de caché, o con uno que no suma: todo a 1 hora (P-007 FR-003, FR-004)
	Crecieron        int // mensajes cuyo máximo supera a su primera línea en alguna partida (P-007 FR-009)
}

// Cerrado es un mensaje que la pasada da por terminado: su evento, con el máximo por partida, y el reparto por
// duración de su escritura de caché, que NO viaja en el evento (D-1) y sólo lo usa quien lo imprime (`--scan`).
type Cerrado struct {
	Evento      event.Event
	Escritura5m int
	Escritura1h int
}

// CerrarFichero da por terminados los mensajes del fichero que se acaba de leer y los devuelve en el orden en
// que aparecieron (P-007 FR-009, FR-016). En una pasada nil, ninguno.
func (p *Pasada) CerrarFichero() []Cerrado {
	if p == nil {
		return nil
	}
	cerrados := make([]Cerrado, 0, len(p.abiertos))
	for _, clave := range p.abiertos {
		m := p.mensajes[clave]
		m.emitido = true
		p.recuentos.Emitidos++
		if m.maximo.partidas() != m.primera.partidas() {
			p.recuentos.Crecieron++
		}
		cerrados = append(cerrados, Cerrado{
			Evento:      conConsumo(m.base, m.maximo),
			Escritura5m: m.maximo.escritura5m,
			Escritura1h: m.maximo.escritura1h,
		})
	}
	p.abiertos = p.abiertos[:0]
	return cerrados
}

// NuevaPasada estrena una pasada vacía.
func NuevaPasada() *Pasada {
	return &Pasada{mensajes: make(map[[bytesEventID]byte]*mensaje)}
}

// Recuentos devuelve una copia de los contadores. En una pasada nil, ceros.
func (p *Pasada) Recuentos() Recuentos {
	if p == nil {
		return Recuentos{}
	}
	return p.recuentos
}

// Resumen devuelve una línea con los recuentos de la pasada, para stderr. Sólo recuentos: ningún
// identificador, ninguna ruta, ningún modelo. En una pasada nil, vacío.
func (p *Pasada) Resumen() string {
	if p == nil {
		return ""
	}
	r := p.recuentos
	return fmt.Sprintf("pasada: %d líneas facturables · %d eventos · %d repetidas del mismo mensaje · "+
		"%d sintéticas · %d sin identificador (no contables) · %d con consumo distinto de la primera",
		r.Facturables, r.Emitidos, r.Repetidas, r.Sinteticas, r.SinIdentificador, r.ConsumoDistinto)
}

// contarFacturable anota una línea `assistant` con modelo.
func (p *Pasada) contarFacturable() {
	if p != nil {
		p.recuentos.Facturables++
	}
}

// contarSintetica anota una línea `<synthetic>` descartada.
func (p *Pasada) contarSintetica() {
	if p != nil {
		p.recuentos.Sinteticas++
	}
}

// contarSinDesglose anota una línea sin desglose de la escritura de caché, o con uno que no suma (P-007 FR-003).
func (p *Pasada) contarSinDesglose() {
	if p != nil {
		p.recuentos.SinDesglose++
	}
}

// contarSinIdentificador anota una línea que no se puede emitir por no traer identificadores.
func (p *Pasada) contarSinIdentificador() {
	if p != nil {
		p.recuentos.SinIdentificador++
	}
}

// acumular suma la línea de `eventID`, con consumo `c`, a su mensaje: el máximo de cada partida y, si su
// escritura de caché es la mayor hasta ahora o la iguala, su desglose (P-007 FR-009; la última, si empatan).
// La primera línea de un mensaje lo abre, con el evento base que da `base`. Las demás se cuentan como
// repetidas y, si su consumo difiere del de la primera, como discrepancia.
//
// Devuelve false si NO la acumula, y entonces quien llama la emite por línea: en una pasada nil (sin memoria,
// todo se emite) y en el caso inalcanzable de un `event_id` que no sea de 32 hex.
func (p *Pasada) acumular(eventID string, c consumo, base func() event.Event) bool {
	if p == nil {
		return false
	}
	var clave [bytesEventID]byte
	if n, err := hex.Decode(clave[:], []byte(eventID)); err != nil || n != bytesEventID {
		// Inalcanzable: `derivarEventID` produce siempre 32 hex. Si no lo fuera, emitir es el lado
		// seguro: la plataforma descarta el repetido, y perder un mensaje no tiene arreglo.
		p.recuentos.Emitidos++
		return false
	}
	m, visto := p.mensajes[clave]
	if !visto {
		p.mensajes[clave] = &mensaje{base: base(), primera: c, maximo: c}
		p.abiertos = append(p.abiertos, clave)
		return true
	}
	p.recuentos.Repetidas++
	if c != m.primera {
		p.recuentos.ConsumoDistinto++
	}
	if m.emitido {
		return true // ya salió en esta pasada, desde otro fichero: no se emite dos veces (P-007 FR-013)
	}
	if c.escrituraCache >= m.maximo.escrituraCache {
		m.maximo.escritura5m, m.maximo.escritura1h = c.escritura5m, c.escritura1h
	}
	m.maximo.entrada = max(m.maximo.entrada, c.entrada)
	m.maximo.salida = max(m.maximo.salida, c.salida)
	m.maximo.escrituraCache = max(m.maximo.escrituraCache, c.escrituraCache)
	m.maximo.lecturaCache = max(m.maximo.lecturaCache, c.lecturaCache)
	return true
}
