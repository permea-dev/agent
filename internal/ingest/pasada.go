package ingest

import (
	"encoding/hex"
	"fmt"
)

// ═══ P-006 · LA PASADA — UN MENSAJE, UN EVENTO ═══════════════════════════════════════════
//
// Claude Code escribe VARIAS líneas por mensaje, todas con el mismo consumo. Dentro de una pasada, la
// primera línea de un mensaje produce el evento y las siguientes con el mismo `event_id` no se
// emiten: se cuentan como repetidas. Si su consumo difiere del de la primera, también se cuenta como
// discrepancia, y las líneas NUNCA se suman (P-006 FR-005). La primera manda porque es lo único que
// se puede sostener leyendo de forma incremental, y la plataforma aplica la misma regla
// (`ON CONFLICT DO NOTHING`).
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

// consumo son las cuatro partidas de tokens de una línea.
type consumo struct {
	entrada, salida, escrituraCache, lecturaCache int
}

// Pasada es el estado de UNA lectura de los logs. El cero no es utilizable: se crea con NuevaPasada.
// Un *Pasada nil es válido en todos sus métodos, que no hacen nada.
type Pasada struct {
	vistos    map[[bytesEventID]byte]consumo
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
}

// NuevaPasada estrena una pasada vacía.
func NuevaPasada() *Pasada {
	return &Pasada{vistos: make(map[[bytesEventID]byte]consumo)}
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

// contarSinIdentificador anota una línea que no se puede emitir por no traer identificadores.
func (p *Pasada) contarSinIdentificador() {
	if p != nil {
		p.recuentos.SinIdentificador++
	}
}

// registrar decide si la línea de `eventID` con consumo `c` produce evento. Devuelve true si es la
// PRIMERA del mensaje en esta pasada, y entonces la anota como emitida. Si no, la cuenta como repetida
// y, si su consumo difiere del de la primera, como discrepancia. En una pasada nil devuelve siempre
// true: sin memoria, todo se emite.
func (p *Pasada) registrar(eventID string, c consumo) bool {
	if p == nil {
		return true
	}
	var clave [bytesEventID]byte
	if n, err := hex.Decode(clave[:], []byte(eventID)); err != nil || n != bytesEventID {
		// Inalcanzable: `derivarEventID` produce siempre 32 hex. Si no lo fuera, emitir es el lado
		// seguro: la plataforma descarta el repetido, y perder un mensaje no tiene arreglo.
		p.recuentos.Emitidos++
		return true
	}
	if primera, visto := p.vistos[clave]; visto {
		p.recuentos.Repetidas++
		if c != primera {
			p.recuentos.ConsumoDistinto++
		}
		return false
	}
	p.vistos[clave] = c
	p.recuentos.Emitidos++
	return true
}
