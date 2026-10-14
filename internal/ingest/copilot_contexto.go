package ingest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/state"
)

// ═══ P-011 · LOS FICHEROS DE COPILOT CLI ════════════════════════════════════════════════════════════════════════════
//
// Copilot CLI guarda cada sesión en `<raíz>/<carpeta>/events.jsonl` (descubrimiento Q1). Junto a él hay ficheros que NO
// son consumo y que el lector no abre nunca (P-7, D-6): `workspace.yaml`, `checkpoints/` (el nombre de sus ficheros sale
// del contenido de la conversación), `rewind-file-snapshots/`, `files/`, `research/`, `session.db` y los `.lock`, tanto
// `.session-operation-locks/` como `.workspace-fork.lock`. Una carpeta sin `events.jsonl` (un intento fallido) no es un
// error ni se avisa.

// ListarCopilot devuelve los `events.jsonl` de las sesiones bajo `raiz`, en orden léxico (P-7, FR-003): un nivel, sin
// recursión, y sin las carpetas que empiezan por `.`. Sólo lee el directorio: no abre ningún fichero.
func ListarCopilot(raiz string) ([]string, error) {
	entradas, err := os.ReadDir(raiz) // ordenadas por nombre
	if err != nil {
		return nil, err
	}
	var rutas []string
	for _, e := range entradas {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		ruta := filepath.Join(raiz, e.Name(), "events.jsonl")
		if info, err := os.Stat(ruta); err == nil && info.Mode().IsRegular() {
			rutas = append(rutas, ruta)
		}
	}
	return rutas, nil
}

// PasadaCopilot son los recuentos de Copilot CLI de UNA pasada (P-011 FR-020): `registros = eventos + repetidas + sin
// identificador + incoherentes`; «compactaciones» y «sin modelo» son subconjuntos de «eventos». La clave de los emitidos es el
// `event_id` ya derivado: nunca un identificador del proveedor.
type PasadaCopilot struct {
	Registros, Eventos, Compactaciones, Repetidas, SinIdentificador, Incoherentes, SinModelo int
	// Aperturas cuenta los ficheros abiertos en la pasada (plan §2; test (15)).
	Aperturas int
	emitidos  map[string]struct{}
}

// NuevaPasadaCopilot crea los recuentos de una pasada.
func NuevaPasadaCopilot() *PasadaCopilot { return &PasadaCopilot{emitidos: make(map[string]struct{})} }

// Resumen es la línea de spec §Textos (P-8): sólo recuentos, nunca identificadores.
func (p *PasadaCopilot) Resumen() string {
	return fmt.Sprintf("copilot_cli: registros %d · eventos %d · compactaciones %d · repetidas %d · sin identificador %d · incoherentes %d · sin modelo %d",
		p.Registros, p.Eventos, p.Compactaciones, p.Repetidas, p.SinIdentificador, p.Incoherentes, p.SinModelo)
}

// HayNovedades dice si la pasada vio algún registro (FR-021): el demonio sólo escribe la línea entonces.
func (p *PasadaCopilot) HayNovedades() bool { return p.Registros > 0 }

// ═══ P-011 B3 · UN FICHERO: OFFSET, PREFIJO Y CONTEXTO POR TRAMO ════════════════════════════════════════════════════════
//
// Calcado de `LeerFicheroCodex` (P-2): el fichero es de SÓLO AÑADIR (descubrimiento Q1), así que se lee por offset con
// `state.Recorrer`. Un fichero que no creció no se abre; uno truncado por debajo del offset se relee desde 0, y lo ya emitido
// sale con el mismo `event_id`; una línea final sin LF no se consume. `state.json` guarda los cuatro campos de siempre (FR-004).
//
// EL CONTEXTO (P-3, D-011-P2): el `cwd` de un registro es el `data.context.cwd` del ÚLTIMO `session.start` o
// `session.resume` anterior; la sesión, el `data.sessionId` del `session.start` (el `resume` no lo trae). Con lectura por
// offset, esas líneas quedan antes: el prefijo `[0, offset)` se relee con un prefiltro de bytes, como `leerPrefijoCodex`,
// para que un registro leído en una pasada posterior lleve lo mismo que leyendo el fichero de una vez (FR-008).
//
// Un subagente (muestra 3) escribe sus registros en el mismo fichero, con otro `sourceSessionId`: salen con la sesión y el
// `cwd` del tramo, y con su `usage.model`. Nada de `subagent.*` se lee.

// Los marcadores de las dos líneas de contexto.
var (
	marcadorInicioCopilot   = []byte(`"type":"session.start"`)
	marcadorReanudarCopilot = []byte(`"type":"session.resume"`)
)

// tramoCopilot es lo que el fichero dice hasta el punto que se está leyendo.
type tramoCopilot struct {
	sessionID, cwd string
}

// esContextoCopilot dice si una línea pasa el prefiltro de las líneas de contexto. El tipo lo confirma el JSON.
func esContextoCopilot(linea []byte) bool {
	return bytes.Contains(linea, marcadorInicioCopilot) || bytes.Contains(linea, marcadorReanudarCopilot)
}

// aprender actualiza el contexto con una línea que ya pasó el prefiltro. Devuelve el error de su envoltura si no es JSON.
func (c *tramoCopilot) aprender(linea []byte) error {
	var l struct {
		Type string `json:"type"`
		Data struct {
			SessionID string `json:"sessionId"`
			Context   struct {
				Cwd string `json:"cwd"`
			} `json:"context"`
		} `json:"data"`
	}
	if err := json.Unmarshal(linea, &l); err != nil {
		return err
	}
	switch l.Type {
	case "session.start":
		c.sessionID, c.cwd = l.Data.SessionID, l.Data.Context.Cwd
	case "session.resume":
		c.cwd = l.Data.Context.Cwd
	}
	return nil
}

// LeerFicheroCopilot lee lo nuevo de una sesión de Copilot CLI y devuelve sus eventos, en orden. Actualiza el estado de ese
// fichero SÓLO en memoria (`Recorrer`): quien llama encola los eventos y DESPUÉS guarda el estado (FR-024). Un error de
// lectura deja el estado de ese fichero como estaba (FR-026). NUNCA escribe, crea ni borra nada bajo la raíz (FR-006).
func LeerFicheroCopilot(st *state.Store, ruta string, base ContextoCopilot, p *PasadaCopilot, avisos io.Writer) ([]event.Event, error) {
	info, err := os.Stat(ruta)
	if err != nil {
		return nil, err
	}
	offset := st.Files[ruta].Offset
	if info.Size() == offset {
		return nil, nil // sin bytes nuevos: no se abre (P-2)
	}
	ctx := &tramoCopilot{}
	if offset > 0 && info.Size() > offset { // truncado (tamaño < offset): `Recorrer` relee desde 0, sin prefijo
		p.Aperturas++
		if err := leerPrefijoCopilot(ruta, offset, ctx); err != nil {
			return nil, err
		}
	}

	p.Aperturas++
	var evs []event.Event
	err = st.Recorrer(ruta, func(linea []byte, _ int64, _ bool) error {
		if esContextoCopilot(linea) {
			if err := ctx.aprender(linea); err != nil && avisos != nil {
				_, _ = fmt.Fprintln(avisos, "skip (línea corrupta):", err) // el texto de `cmd/permea/main.go:320`
			}
			return nil
		}
		base.SessionID, base.Cwd = ctx.sessionID, ctx.cwd
		ev, clase, marcas, err := RegistroCopilot(linea, base)
		if err != nil {
			if avisos != nil {
				_, _ = fmt.Fprintln(avisos, "skip (línea corrupta):", err)
			}
			return nil // una línea corrupta se salta: no corta el fichero
		}
		if clase == NoEsRegistroCopilot {
			return nil
		}
		p.Registros++
		switch clase {
		case SinIdentificadorCopilot:
			p.SinIdentificador++
		case IncoherenteCopilot:
			p.Incoherentes++
		case EventoCopilot:
			if _, visto := p.emitidos[ev.EventID]; visto {
				p.Repetidas++
				return nil
			}
			p.emitidos[ev.EventID] = struct{}{}
			p.Eventos++
			if marcas.Compactacion {
				p.Compactaciones++
			}
			if marcas.SinModelo {
				p.SinModelo++
			}
			evs = append(evs, *ev)
		}
		return nil
	}, func(leido int64, _ time.Time) int64 { return leido })
	if err != nil {
		return nil, err
	}
	return evs, nil
}

// leerPrefijoCopilot lee `[0, hasta)` sólo para el contexto: no emite ni avisa. Sin tope de línea (FR-009).
func leerPrefijoCopilot(ruta string, hasta int64, ctx *tramoCopilot) error {
	f, err := os.Open(ruta)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }() // sólo lectura
	r := bufio.NewReader(io.LimitReader(f, hasta))
	for {
		linea, err := r.ReadBytes('\n')
		if len(linea) > 0 && esContextoCopilot(linea) {
			_ = ctx.aprender(linea) // en el prefijo, una línea corrupta se ignora en silencio
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
