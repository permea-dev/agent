package ingest

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/permea-dev/agent/internal/event"
)

// ═══ P-006 · LA PASADA — UN MENSAJE, UN EVENTO ═══════════════════════════════════════════
//
// Claude Code escribe VARIAS líneas por mensaje. Dentro de una pasada, las líneas de un mensaje producen
// UN evento: las siguientes a la primera se cuentan como repetidas y, si su consumo difiere del de la
// primera, también como discrepancia. Las líneas NUNCA se suman (P-006 FR-005).
//
// ═══ P-007 · CADA PARTIDA VALE SU MÁXIMO, Y EL EVENTO SALE CUANDO EL MENSAJE ESTÁ CERRADO ══════
//
// La salida puede crecer de una línea a la siguiente (7 → 1 303 → 89 817), así que «la primera manda»
// se queda corta (P-007 FR-009). La pasada ACUMULA cada mensaje —el máximo de cada partida, y el desglose
// de la escritura de caché de la línea que da su máximo, la última si empatan— y lo emite cuando está
// CERRADO (P-007 FR-010):
//   - (i) ha empezado otro mensaje posterior en el MISMO fichero;
//   - (ii) han pasado ≥ `EsperaDeCierre` desde el `timestamp` de su última línea Y desde la última
//     modificación de su fichero;
//   - (iii) han pasado ≥ `TopeDeEspera` desde el `timestamp` de su última línea, cambie o no su fichero.
// El evento toma de la primera línea su momento, su modelo y sus referencias (`plan.md` D-007-P5). Sin pasada
// no se acumula: cada línea se emite como hasta ahora.
//
// ═══ LO ABIERTO NO SE GUARDA: SE RELEE ═══════════════════════════════════════════════════
//
// Su ámbito sigue siendo UNA pasada: una ejecución de `--run`, un ciclo de `--daemon` o un `--scan` de un
// fichero. Lo que sigue abierto al acabar un fichero no se guarda en ningún sitio (P-007 FR-011): quien
// lee deja el offset de ese fichero en el comienzo del mensaje abierto (`CerrarConReloj` se lo da), y la
// pasada siguiente lo relee del log. Lo ya emitido no se recuerda entre pasadas (P-006 FR-033): si
// reaparece, sale con el MISMO `event_id` y la plataforma lo descarta por `(org_id, event_id)`.
//
// ═══ NUNCA GUARDA NI ESCRIBE UN IDENTIFICADOR DEL PROVEEDOR ════════════════════════════════
//
// La clave del conjunto es el `event_id` ya derivado, en sus 16 bytes, no en la cadena hex
// (`research.md` R3.4). El resumen sólo lleva recuentos: dice CUÁNTO, nunca QUIÉN.

// EsperaDeCierre es T, la espera de la regla (ii), y TopeDeEspera la de la regla (iii). Son constantes: no
// se configuran (P-007 Q-6, E-3).
const (
	EsperaDeCierre = 10 * time.Minute
	TopeDeEspera   = 24 * time.Hour
)

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
// primera línea; el máximo de cada partida, con el desglose de la línea del máximo de la escritura; dónde
// empieza en su fichero y el `timestamp` de su última línea; y en qué punto está.
type mensaje struct {
	base     event.Event
	primera  consumo
	maximo   consumo
	inicio   int64
	ultimo   time.Time
	cerrado  bool // por la regla (i): ya no acepta líneas, y sale al acabar el fichero
	retenido bool // abierto al acabar su fichero: no sale en esta pasada
	emitido  bool
}

// Pasada es el estado de UNA lectura de los logs. El cero no es utilizable: se crea con NuevaPasada.
// Un *Pasada nil es válido en todos sus métodos, que no hacen nada.
type Pasada struct {
	mensajes  map[[bytesEventID]byte]*mensaje
	abiertos  [][bytesEventID]byte // los del fichero en curso, en el orden en que aparecieron
	inicio    int64                // dónde empieza la línea en curso (`Situar`)
	releida   bool                 // si la línea en curso ya se leyó en otra pasada (`Situar`)
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
	EnEspera         int // mensajes que siguen abiertos al terminar la pasada (P-007 FR-014)
	Releidas         int // líneas facturables de un mensaje en espera, leídas otra vez (P-007 FR-021)
	Tardias          int // líneas de un mensaje ya cerrado, que no se emiten (P-007 FR-013)
}

// Cerrado es un mensaje que la pasada da por terminado: su evento, con el máximo por partida, y el reparto por
// duración de su escritura de caché, que NO viaja en el evento (D-1) y sólo lo usa quien lo imprime (`--scan`).
type Cerrado struct {
	Evento      event.Event
	Escritura5m int
	Escritura1h int
}

// NuevaPasada estrena una pasada vacía.
func NuevaPasada() *Pasada {
	return &Pasada{mensajes: make(map[[bytesEventID]byte]*mensaje)}
}

// Situar dice a la pasada dónde empieza en su fichero la línea que se va a leer, y si es releída
// (P-007 FR-011, FR-021). Quien no la llama lee como si todo empezara en el byte 0 y nada fuera releído.
func (p *Pasada) Situar(inicio int64, releida bool) {
	if p != nil {
		p.inicio, p.releida = inicio, releida
	}
}

// CerrarFichero da por terminados TODOS los mensajes del fichero que se acaba de leer y los devuelve en el
// orden en que aparecieron. Es lo que hace `--scan`: el fichero está completo (P-007 FR-016). En una pasada
// nil, ninguno.
func (p *Pasada) CerrarFichero() []Cerrado {
	if p == nil {
		return nil
	}
	cerrados := make([]Cerrado, 0, len(p.abiertos))
	for _, clave := range p.abiertos {
		cerrados = append(cerrados, p.emitir(p.mensajes[clave]))
	}
	p.abiertos = p.abiertos[:0]
	return cerrados
}

// CerrarConReloj termina el fichero que se acaba de leer, a la hora `ahora` y con el fichero modificado por
// última vez en `modificado` (P-007 FR-010). Devuelve, en el orden en que aparecieron, los mensajes cerrados.
// Los que siguen abiertos se quedan sin emitir; si hay alguno, devuelve el comienzo de la primera línea del
// primero, que es donde debe quedarse el offset del fichero (FR-011).
func (p *Pasada) CerrarConReloj(ahora, modificado time.Time) (cerrados []Cerrado, abierto int64, hayAbierto bool) {
	if p == nil {
		return nil, 0, false
	}
	for _, clave := range p.abiertos {
		m := p.mensajes[clave]
		if !m.cerrado && !cierraPorEspera(m.ultimo, modificado, ahora) {
			m.retenido = true
			p.recuentos.EnEspera++
			if !hayAbierto || m.inicio < abierto {
				abierto, hayAbierto = m.inicio, true
			}
			continue
		}
		cerrados = append(cerrados, p.emitir(m))
	}
	p.abiertos = p.abiertos[:0]
	return cerrados, abierto, hayAbierto
}

// cierraPorEspera dice si un mensaje cuya última línea es de `ultimo`, en un fichero modificado por última vez
// en `modificado`, está cerrado a la hora `ahora` por la regla (iii) o por la (ii) (P-007 FR-010).
func cierraPorEspera(ultimo, modificado, ahora time.Time) bool {
	if ahora.Sub(ultimo) >= TopeDeEspera {
		return true // (iii): el tope, cambie o no el fichero
	}
	return ahora.Sub(ultimo) >= EsperaDeCierre && ahora.Sub(modificado) >= EsperaDeCierre // (ii)
}

// emitir saca un mensaje: su evento con el máximo de cada partida y el coste calculado con él.
func (p *Pasada) emitir(m *mensaje) Cerrado {
	m.emitido = true
	p.recuentos.Emitidos++
	if m.maximo.partidas() != m.primera.partidas() {
		p.recuentos.Crecieron++
	}
	return Cerrado{
		Evento:      conConsumo(m.base, m.maximo),
		Escritura5m: m.maximo.escritura5m,
		Escritura1h: m.maximo.escritura1h,
	}
}

// Recuentos devuelve una copia de los contadores. En una pasada nil, ceros.
func (p *Pasada) Recuentos() Recuentos {
	if p == nil {
		return Recuentos{}
	}
	return p.recuentos
}

// HayNovedades dice si la pasada leyó algo que merezca resumen: líneas facturables que no fueran releídas, o
// eventos emitidos. Una pasada que sólo relee un mensaje en espera no las tiene, y el demonio calla (P-007
// FR-021): un resumen repetido cada ciclo es ruido, y el ruido enseña a ignorar los avisos.
func (p *Pasada) HayNovedades() bool {
	r := p.Recuentos()
	return r.Facturables > r.Releidas || r.Emitidos > 0
}

// Resumen devuelve los recuentos de la pasada en DOS líneas, para stderr (P-007 FR-017): la primera, la de
// 006 sin cambiar un byte; la segunda, la de 007. Sólo recuentos: ningún identificador, ninguna ruta,
// ningún modelo. En una pasada nil, vacío.
func (p *Pasada) Resumen() string {
	if p == nil {
		return ""
	}
	r := p.recuentos
	return fmt.Sprintf("pasada: %d líneas facturables · %d eventos · %d repetidas del mismo mensaje · "+
		"%d sintéticas · %d sin identificador (no contables) · %d con consumo distinto de la primera",
		r.Facturables, r.Emitidos, r.Repetidas, r.Sinteticas, r.SinIdentificador, r.ConsumoDistinto) + "\n" +
		fmt.Sprintf("pasada: %d mensajes que crecieron entre líneas · %d en espera de cerrarse · "+
			"%d líneas releídas de un mensaje en espera · %d líneas tardías · %d líneas sin desglose de caché (a 1 hora)",
			r.Crecieron, r.EnEspera, r.Releidas, r.Tardias, r.SinDesglose)
}

// AvisoDeAbiertos es la línea que escribe `--run` cuando quedan mensajes abiertos (P-007 FR-014, P-5), o
// vacío si no queda ninguno.
func (p *Pasada) AvisoDeAbiertos() string {
	if n := p.Recuentos().EnEspera; n > 0 {
		return fmt.Sprintf("%d mensajes siguen abiertos: se enviarán en la próxima pasada", n)
	}
	return ""
}

// contarFacturable anota una línea `assistant` con modelo y, si es releída, también como releída.
func (p *Pasada) contarFacturable() {
	if p != nil {
		p.recuentos.Facturables++
		if p.releida {
			p.recuentos.Releidas++
		}
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

// acumular suma la línea de `eventID`, con consumo `c` y `timestamp` `momento`, a su mensaje: el máximo de cada
// partida y, si su escritura de caché es la mayor hasta ahora o la iguala, su desglose (P-007 FR-009; la
// última, si empatan).
//
// La primera línea de un mensaje lo abre, con el evento base que da `base`, y CIERRA por la regla (i) los
// que estaban abiertos en el fichero (FR-010). Las demás se cuentan como repetidas y, si su consumo difiere
// del de la primera, como discrepancia. La de un mensaje ya cerrado es TARDÍA: se cuenta y no cambia nada
// (FR-013).
//
// Devuelve false si NO la acumula, y entonces quien llama la emite por línea: en una pasada nil (sin memoria,
// todo se emite) y en el caso inalcanzable de un `event_id` que no sea de 32 hex.
func (p *Pasada) acumular(eventID string, c consumo, momento time.Time, base func() event.Event) bool {
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
		for _, otra := range p.abiertos {
			p.mensajes[otra].cerrado = true // (i): empieza otro mensaje en el mismo fichero
		}
		p.mensajes[clave] = &mensaje{base: base(), primera: c, maximo: c, inicio: p.inicio, ultimo: momento}
		p.abiertos = append(p.abiertos, clave)
		return true
	}
	p.recuentos.Repetidas++
	if c != m.primera {
		p.recuentos.ConsumoDistinto++
	}
	if m.cerrado || m.retenido || m.emitido {
		p.recuentos.Tardias++
		return true // tardía: no se emite dos veces ni cambia lo cerrado (P-007 FR-013)
	}
	if c.escrituraCache >= m.maximo.escrituraCache {
		m.maximo.escritura5m, m.maximo.escritura1h = c.escritura5m, c.escritura1h
	}
	m.maximo.entrada = max(m.maximo.entrada, c.entrada)
	m.maximo.salida = max(m.maximo.salida, c.salida)
	m.maximo.escrituraCache = max(m.maximo.escrituraCache, c.escrituraCache)
	m.maximo.lecturaCache = max(m.maximo.lecturaCache, c.lecturaCache)
	if momento.After(m.ultimo) {
		m.ultimo = momento
	}
	return true
}
