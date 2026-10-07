package ingest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/state"
)

// ═══ P-008 B3 · CONTEXTO, ESTADO Y RECUENTOS DE CODEX ═════════════════════════════════════════
//
// Un registro de Codex no trae ni su modelo ni su directorio: están ANTES, en `session_meta`, en el
// `turn_context` de su turno y en los `thread_settings_applied`. Con lectura por offset, la pasada siguiente
// empieza después de ellos. FR-005 exige que un registro leído en una pasada posterior lleve lo mismo que
// leyendo el fichero de una vez, sin cambiar `state.json` (cuatro campos).
//
// ═══ EL MECANISMO (plan D-008-P1) ══════════════════════════════════════════════════════════════
//
// Un fichero se abre sólo si creció. Su PREFIJO, `[0, offset)`, se lee hacia delante con un filtro de bytes:
// sólo se decodifica la línea que contiene uno de los marcadores. El filtro es sólo un prefiltro; el tipo lo
// confirma el JSON. Después, `Recorrer` lee lo nuevo, y el contexto se sigue actualizando con las líneas nuevas
// (un `turn_context` llega a menudo en la misma pasada que su registro). Un fichero truncado o rotado se relee
// desde 0, y entonces NO hay prefijo: el contexto sale sólo de lo que se relee, en orden.
//
// ═══ LOS AVISOS (FR-029) ═══════════════════════════════════════════════════════════════════════
//
// `internal/ingest` no escribe en stderr por su cuenta: el aviso de línea corrupta va al `io.Writer` que da
// quien llama (en `cmd/permea`, stderr), con el texto que ya existe. Sólo se avisa en la parte nueva: en el
// prefijo releído se ignora en silencio, para no repetir el aviso en cada ciclo del demonio.

// PasadaCodex son los recuentos de Codex de UNA pasada (P-008 FR-027) y su línea de resumen (§Textos aprobados).
// Cumplen, por construcción, `Respuestas = Eventos + Repetidas + SinIdentificador + Incoherentes`, y
// `SinModelo` cuenta eventos.
type PasadaCodex struct {
	Respuestas, Eventos, Repetidas, SinIdentificador, Incoherentes, SinModelo int
	FormatoAnterior, Comprimidos                                              int

	// emitidos son los `event_id` que ya salieron en la pasada (FR-008): una respuesta copiada a otro fichero
	// (bifurcación) sale una vez. Entre pasadas no hay memoria: los repetidos los descarta la plataforma.
	emitidos map[string]struct{}
}

// NuevaPasadaCodex crea los recuentos de una pasada.
func NuevaPasadaCodex() *PasadaCodex { return &PasadaCodex{emitidos: make(map[string]struct{})} }

// Resumen es la línea aprobada (P-008 §Textos aprobados, E-2/P-9): sólo recuentos, nunca identificadores.
func (p *PasadaCodex) Resumen() string {
	return fmt.Sprintf("codex: respuestas %d · eventos %d · repetidas %d · sin identificador %d · incoherentes %d · sin modelo %d"+
		" · ficheros en formato anterior %d · ficheros comprimidos %d",
		p.Respuestas, p.Eventos, p.Repetidas, p.SinIdentificador, p.Incoherentes, p.SinModelo,
		p.FormatoAnterior, p.Comprimidos)
}

// marcadoresCodex son el prefiltro de bytes: una línea sin ninguno no se decodifica (D-008-P1).
var marcadoresCodex = [][]byte{
	[]byte(`"session_meta"`), []byte(`"turn_context"`), []byte(`"thread_settings_applied"`),
	[]byte(`"token_usage_record"`), []byte(`"token_count"`),
}

func pasaFiltroCodex(line []byte) bool {
	for _, m := range marcadoresCodex {
		if bytes.Contains(line, m) {
			return true
		}
	}
	return false
}

// turnoCodex es lo que un `turn_context` dice de su turno.
type turnoCodex struct{ modelo, cwd string }

// contextoFichero es lo que se sabe de un fichero hasta la línea que se está leyendo. Vive sólo durante la pasada.
type contextoFichero struct {
	cwdSesion    string                // `session_meta.cwd`
	turnos       map[string]turnoCodex // por `turn_id`
	vigente      string                // el modelo de la última `turn_context` o `thread_settings_applied`
	hayRegistros bool                  // algún `token_usage_record`: formato actual (FR-017)
	hayContador  bool                  // algún `token_count` con `info`: formato anterior si no hay registros
}

// payloadContexto son los campos de contexto de las líneas que no son registros.
type payloadContexto struct {
	Type           string                  `json:"type"`
	Cwd            string                  `json:"cwd"`
	TurnID         string                  `json:"turn_id"`
	Model          string                  `json:"model"`
	ThreadSettings *struct{ Model string } `json:"thread_settings"`
	Info           json.RawMessage         `json:"info"`
}

// aprender actualiza el contexto con una línea que ya pasó el filtro. Devuelve el error de su envoltura si no es
// JSON (una línea corrupta, E-4): quien llama decide si avisa.
func (c *contextoFichero) aprender(line []byte) error {
	var l lineaCodex
	if err := json.Unmarshal(line, &l); err != nil {
		return err
	}
	if l.Type == tipoRegistroCodex {
		c.hayRegistros = true
		return nil
	}
	var p payloadContexto
	if json.Unmarshal(l.Payload, &p) != nil {
		return nil // una línea de contexto ilegible no aporta nada; no es una respuesta
	}
	switch {
	case l.Type == "session_meta":
		c.cwdSesion = p.Cwd
	case l.Type == "turn_context":
		c.turnos[p.TurnID] = turnoCodex{modelo: p.Model, cwd: p.Cwd}
		if p.Model != "" {
			c.vigente = p.Model
		}
	case l.Type == "event_msg" && p.Type == "thread_settings_applied" && p.ThreadSettings != nil:
		if p.ThreadSettings.Model != "" {
			c.vigente = p.ThreadSettings.Model
		}
	case l.Type == "event_msg" && p.Type == "token_count" && len(p.Info) > 0 && string(p.Info) != "null":
		c.hayContador = true
	}
	return nil
}

// delTurno aplica FR-011 y FR-014: el modelo y el `cwd` del `turn_context` del turno; si no lo hay, el modelo
// vigente y el `cwd` de `session_meta`.
func (c *contextoFichero) delTurno(turnID string) (modelo, cwd string) {
	t, ok := c.turnos[turnID]
	modelo, cwd = t.modelo, t.cwd
	if !ok || modelo == "" {
		modelo = c.vigente
	}
	if !ok || cwd == "" {
		cwd = c.cwdSesion
	}
	return modelo, cwd
}

// LeerFicheroCodex lee lo nuevo de una sesión de Codex y devuelve sus eventos, en orden. Actualiza el estado de
// ese fichero SÓLO en memoria (`Recorrer`): quien llama encola los eventos y DESPUÉS guarda el estado (FR-025).
// Un error de lectura deja el estado de ese fichero como estaba (FR-028); quien llama decide si sigue.
func LeerFicheroCodex(st *state.Store, ruta string, base ContextoCodex, p *PasadaCodex, avisos io.Writer) ([]event.Event, error) {
	info, err := os.Stat(ruta)
	if err != nil {
		return nil, err
	}
	offset := st.Files[ruta].Offset
	if info.Size() == offset {
		return nil, nil // sin bytes nuevos: no se abre (D-008-P1)
	}
	ctx := &contextoFichero{turnos: make(map[string]turnoCodex)}
	if offset > 0 && info.Size() > offset { // truncado o rotado (tamaño < offset): `Recorrer` relee desde 0, sin prefijo
		if err := leerPrefijoCodex(ruta, offset, ctx); err != nil {
			return nil, err
		}
	}

	base.DelTurno = ctx.delTurno
	var evs []event.Event
	err = st.Recorrer(ruta, func(line []byte, _ int64, _ bool) error {
		if !pasaFiltroCodex(line) {
			return nil
		}
		if err := ctx.aprender(line); err != nil {
			if avisos != nil {
				_, _ = fmt.Fprintln(avisos, "skip (línea corrupta):", err) // el texto de `cmd/permea/main.go:282`
			}
			return nil // FR-029: se salta, no corta el fichero
		}
		ev, clase, err := LineaCodex(line, base)
		if err != nil || clase == NoEsRegistro {
			return nil
		}
		p.Respuestas++
		switch clase {
		case SinIdentificador:
			p.SinIdentificador++
		case Incoherente:
			p.Incoherentes++
		case EventoCodex:
			if _, visto := p.emitidos[ev.EventID]; visto {
				p.Repetidas++
				return nil
			}
			p.emitidos[ev.EventID] = struct{}{}
			p.Eventos++
			if ev.Model == "" {
				p.SinModelo++
			}
			evs = append(evs, *ev)
		}
		return nil
	}, func(leido int64, _ time.Time) int64 { return leido })
	if err != nil {
		return nil, err
	}
	if !ctx.hayRegistros && ctx.hayContador {
		p.FormatoAnterior++
	}
	return evs, nil
}

// leerPrefijoCodex lee `[0, hasta)` sólo para el contexto: no emite ni avisa (FR-029).
func leerPrefijoCodex(ruta string, hasta int64, ctx *contextoFichero) error {
	f, err := os.Open(ruta)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }() // sólo lectura
	r := bufio.NewReader(io.LimitReader(f, hasta))
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 && pasaFiltroCodex(line) {
			_ = ctx.aprender(line) // en el prefijo, una línea corrupta se ignora en silencio (FR-029)
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// ContarComprimido cuenta un `.zst` de Codex una vez, y otra sólo si cambia (P-008 FR-018, plan D-008-P2). Su
// entrada en `state.json` lleva los cuatro campos de siempre, con `Offset = Size`: no se abre ni se lee.
func ContarComprimido(st *state.Store, ruta string, p *PasadaCodex) error {
	info, err := os.Stat(ruta)
	if err != nil {
		return err
	}
	previo, visto := st.Files[ruta]
	if visto && previo.Size == info.Size() && previo.ModTime == info.ModTime().Unix() {
		return nil
	}
	st.Files[ruta] = state.FileState{Path: ruta, Size: info.Size(), ModTime: info.ModTime().Unix(), Offset: info.Size()}
	p.Comprimidos++
	return nil
}

// HayNovedades dice si el demonio escribe la línea de Codex en este ciclo (P-008 FR-019): si hubo respuestas, ficheros en
// formato anterior o comprimidos. Un resumen vacío cada ciclo es ruido, como en Claude Code (P-006).
func (p *PasadaCodex) HayNovedades() bool {
	return p.Respuestas > 0 || p.FormatoAnterior > 0 || p.Comprimidos > 0
}

// ListarCodex enumera, bajo la raíz de Codex y a cualquier profundidad, las sesiones (`.jsonl`) y los comprimidos
// (`.zst`), cada lista en orden de recorrido (léxico). Un subdirectorio que no se puede leer se salta: su error no impide
// leer lo demás (P-008 FR-028).
func ListarCodex(raiz string) (sesiones, comprimidos []string, err error) {
	err = filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == raiz {
				return err
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		switch filepath.Ext(p) {
		case ".jsonl":
			sesiones = append(sesiones, p)
		case ".zst":
			comprimidos = append(comprimidos, p)
		}
		return nil
	})
	return sesiones, comprimidos, err
}
