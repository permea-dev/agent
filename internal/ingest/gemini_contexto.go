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
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/state"
)

// ═══ P-009 B3 · EL FICHERO DE GEMINI CLI, SUS APARICIONES Y SU CONTEXTO ═══════════════════════════
//
// Una sesión de Gemini CLI es una BITÁCORA: la misma respuesta aparece en su línea, en las líneas que la
// reescriben (`recordToolCalls`, `recordMessageTokens`) y dentro de cada `$set.messages` (compresión, reanudación,
// rebobinado). Aquí se recorren las APARICIONES, en orden, y se emite la primera de cada `id` con `tokens`
// (FR-006, FR-008). NUNCA se reconstruye el estado final: un `$set.messages` reemplaza el historial y, si se
// aplicara, borraría respuestas ya gastadas (descubrimiento Q4).
//
// ═══ EL CONTEXTO ENTRE PASADAS (plan D-009-P1, P-3 (a)) ═══════════════════════════════════════════
//
// Un fichero se abre sólo si creció. Su PREFIJO, `[0, offset)`, se lee hacia delante con un filtro de bytes: la
// línea 1 siempre (la cabecera, con el `sessionId`), ninguna `$set`, y de las demás sólo las que contienen
// `"tokens":{`. De ésas se guarda el `event_id` en `vistos`. Basta, porque la CLI reescribe el mensaje entero, como
// línea propia, cada vez que le pone tokens: un `id` con tokens de un `$set` ya apareció antes en una línea de
// mensaje. Así, una reanudación días después no reemite lo que su `$set.messages` repite.
//
// Plan B, si SC-013 no se cumpliera: sacar el `id` del principio de la línea (`{"id":"…"`, el orden de claves de
// `newMessage` en la CLI) sin decodificarla. Escrito y sin usar.
//
// Un fichero truncado o rotado se relee desde 0 (`state.go:113`), y entonces NO hay prefijo.
//
// ═══ LOS AVISOS (FR-020) ═════════════════════════════════════════════════════════════════════════
//
// El aviso de línea corrupta va al `io.Writer` que da quien llama, con el texto que ya existe, y sólo en la parte
// nueva: en el prefijo, en silencio, para no repetirlo en cada ciclo del demonio.

// PasadaGemini son los recuentos de Gemini CLI de UNA pasada (P-009 FR-019) y su línea de resumen (§Textos aprobados).
// Cumplen, por construcción, `Respuestas = Eventos + Repetidas + SinIdentificador + Incoherentes`; `SinModelo` y
// `TotalDescuadrado` cuentan eventos.
type PasadaGemini struct {
	Respuestas, Eventos, Repetidas, SinIdentificador, Incoherentes, SinModelo, TotalDescuadrado int
	FormatoAnterior                                                                             int

	// emitidos son los `event_id` que ya salieron en la pasada (FR-008): la misma sesión en dos carpetas sale una vez.
	emitidos map[string]struct{}
	// raices es la caché de `.project_root` por carpeta `<slug>`, para la pasada (D-009-P3).
	raices map[string]raizGemini
}

// raizGemini es lo leído de un `.project_root`: su texto, o el error que omite los ficheros de esa carpeta.
type raizGemini struct {
	texto string
	err   error
}

// NuevaPasadaGemini crea los recuentos de una pasada.
func NuevaPasadaGemini() *PasadaGemini {
	return &PasadaGemini{emitidos: make(map[string]struct{}), raices: make(map[string]raizGemini)}
}

// Resumen es la línea aprobada (P-009 §Textos aprobados): sólo recuentos, nunca identificadores.
func (p *PasadaGemini) Resumen() string {
	return fmt.Sprintf("gemini: respuestas %d · eventos %d · repetidas %d · sin identificador %d · incoherentes %d · sin modelo %d"+
		" · total descuadrado %d · ficheros en formato anterior %d",
		p.Respuestas, p.Eventos, p.Repetidas, p.SinIdentificador, p.Incoherentes, p.SinModelo,
		p.TotalDescuadrado, p.FormatoAnterior)
}

// HayNovedades dice si el demonio escribe la línea de Gemini en este ciclo (P-009 FR-021): si hubo respuestas o
// ficheros en formato anterior. Un resumen vacío cada ciclo es ruido.
func (p *PasadaGemini) HayNovedades() bool {
	return p.Respuestas > 0 || p.FormatoAnterior > 0
}

// FicheroGemini es una sesión y la carpeta `<slug>` que la contiene, la de su `.project_root` (D-009-P3). El `<slug>`
// sale de la POSICIÓN bajo la raíz, no del nombre de las carpetas: la de un subagente cuelga de `chats/<padre>/`.
type FicheroGemini struct {
	Ruta string
	Slug string
}

// ListarGemini enumera, bajo `<raiz>/tmp`, las sesiones (`.jsonl`) y los ficheros en formato anterior (`.json`) de
// `<slug>/chats/` y de `<slug>/chats/<padre>/` (subagentes), y nada más: ni `logs.json`, ni `logs/`, ni otra
// profundidad. Las sesiones van primero las de las carpetas con `.project_root` y después las demás, cada grupo en
// orden léxico (D-009-P4). Un subdirectorio que no se puede leer se salta; sólo el error de `<raiz>/tmp` se devuelve.
func ListarGemini(raiz string) (sesiones []FicheroGemini, anteriores []string, err error) {
	tmp := filepath.Join(raiz, "tmp")
	slugs, err := os.ReadDir(tmp)
	if err != nil {
		return nil, nil, err
	}
	var conRaiz, sinRaiz []FicheroGemini
	for _, s := range slugs {
		if !s.IsDir() {
			continue
		}
		dirSlug := filepath.Join(tmp, s.Name())
		var propias []FicheroGemini
		recoger := func(dir string) []fs.DirEntry {
			entradas, err := os.ReadDir(dir)
			if err != nil {
				return nil // ilegible o ausente: se salta
			}
			for _, e := range entradas {
				if e.IsDir() {
					continue
				}
				switch filepath.Ext(e.Name()) {
				case ".jsonl":
					propias = append(propias, FicheroGemini{Ruta: filepath.Join(dir, e.Name()), Slug: dirSlug})
				case ".json":
					anteriores = append(anteriores, filepath.Join(dir, e.Name()))
				}
			}
			return entradas
		}
		chats := filepath.Join(dirSlug, "chats")
		for _, e := range recoger(chats) {
			if e.IsDir() {
				recoger(filepath.Join(chats, e.Name()))
			}
		}
		if _, err := os.Stat(filepath.Join(dirSlug, ".project_root")); err == nil {
			conRaiz = append(conRaiz, propias...)
		} else {
			sinRaiz = append(sinRaiz, propias...)
		}
	}
	for _, g := range [][]FicheroGemini{conRaiz, sinRaiz} {
		sort.Slice(g, func(i, j int) bool { return g[i].Ruta < g[j].Ruta })
	}
	sort.Strings(anteriores)
	return append(conRaiz, sinRaiz...), anteriores, nil
}

// raizDe lee el `.project_root` de una carpeta `<slug>`, una vez por pasada. Su texto va sin espacios ni salto de línea
// al final. Si no existe, el texto es vacío y no hay error (FR-014); cualquier otro error omite el fichero (FR-025).
func (p *PasadaGemini) raizDe(dirSlug string) (string, error) {
	if r, visto := p.raices[dirSlug]; visto {
		return r.texto, r.err
	}
	var r raizGemini
	b, err := os.ReadFile(filepath.Join(dirSlug, ".project_root"))
	switch {
	case err == nil:
		r.texto = strings.TrimRightFunc(string(b), unicode.IsSpace)
	case !errors.Is(err, fs.ErrNotExist):
		r.err = err
	}
	p.raices[dirSlug] = r
	return r.texto, r.err
}

// cabeceraGemini es la primera línea de una sesión.
type cabeceraGemini struct {
	SessionID   *string `json:"sessionId"`
	ProjectHash *string `json:"projectHash"`
}

// sesionDeCabecera devuelve el `sessionId` si la línea es una cabecera: un objeto con `sessionId` y `projectHash` (D-009-P2).
func sesionDeCabecera(line []byte) string {
	var c cabeceraGemini
	if json.Unmarshal(line, &c) != nil || c.SessionID == nil || c.ProjectHash == nil {
		return ""
	}
	return *c.SessionID
}

// envolturaGemini distingue una línea `$set` del resto.
type envolturaGemini struct {
	Set *struct {
		Messages []json.RawMessage `json:"messages"`
	} `json:"$set"`
}

// aparicionesDe devuelve las apariciones de mensaje de una línea: la propia línea, o los elementos de su
// `$set.messages`. Una `$set` sin `messages` no tiene ninguna. El error es sólo el de una envoltura que no es JSON.
func aparicionesDe(line []byte) ([]json.RawMessage, error) {
	var e envolturaGemini
	if err := json.Unmarshal(line, &e); err != nil {
		return nil, err
	}
	if e.Set != nil {
		return e.Set.Messages, nil
	}
	return []json.RawMessage{line}, nil
}

// marcaPrefijoSet y marcaPrefijoTokens son el filtro de bytes del prefijo (D-009-P1).
var (
	marcaPrefijoSet    = []byte(`{"$set"`)
	marcaPrefijoTokens = []byte(`"tokens":{`)
)

// leerPrefijoGemini lee `[0, hasta)` sólo para el contexto: el `sessionId` de la línea 1 y los `event_id` de las
// respuestas con tokens. No emite ni avisa: una línea que no decodifica se ignora en silencio (FR-020).
func leerPrefijoGemini(ruta string, hasta int64, sesion *string, vistos map[string]struct{}) error {
	f, err := os.Open(ruta)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }() // sólo lectura
	r := bufio.NewReader(io.LimitReader(f, hasta))
	for primera := true; ; primera = false {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			switch {
			case primera:
				*sesion = sesionDeCabecera(line)
			case bytes.HasPrefix(line, marcaPrefijoSet) || !bytes.Contains(line, marcaPrefijoTokens):
			default:
				var m mensajeGemini
				if json.Unmarshal(line, &m) == nil && textoDe(m.Type) == tipoRespuestaGemini && !ausente(m.Tokens) {
					if id, ok := derivarEventIDGemini(textoDe(m.ID)); ok {
						vistos[id] = struct{}{}
					}
				}
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// LeerFicheroGemini lee lo nuevo de una sesión de Gemini CLI y devuelve sus eventos, en orden. Actualiza el estado de
// ese fichero SÓLO en memoria (`Recorrer`): quien llama encola los eventos y DESPUÉS guarda el estado (FR-024). Un
// error de lectura deja el estado de ese fichero como estaba (FR-025); quien llama decide si sigue.
func LeerFicheroGemini(st *state.Store, f FicheroGemini, base ContextoGemini, p *PasadaGemini, avisos io.Writer) ([]event.Event, error) {
	info, err := os.Stat(f.Ruta)
	if err != nil {
		return nil, err
	}
	offset := st.Files[f.Ruta].Offset
	if info.Size() == offset {
		return nil, nil // sin bytes nuevos: no se abre
	}
	ctx := base
	if ctx.ProjectRoot, err = p.raizDe(f.Slug); err != nil {
		return nil, err
	}
	vistos := make(map[string]struct{})
	if offset > 0 && info.Size() > offset { // truncado o rotado (tamaño < offset): `Recorrer` relee desde 0, sin prefijo
		if err := leerPrefijoGemini(f.Ruta, offset, &ctx.SessionID, vistos); err != nil {
			return nil, err
		}
	}

	var evs []event.Event
	err = st.Recorrer(f.Ruta, func(line []byte, inicio int64, _ bool) error {
		if inicio == 0 {
			ctx.SessionID = sesionDeCabecera(line)
		}
		apariciones, err := aparicionesDe(line)
		if err != nil {
			if avisos != nil {
				_, _ = fmt.Fprintln(avisos, "skip (línea corrupta):", err) // el texto de `cmd/permea/main.go:295`
			}
			return nil // FR-020: se salta, no corta el fichero
		}
		for _, a := range apariciones {
			ev, clase, marcas, err := RespuestaGemini(a, ctx)
			if err != nil || clase == NoEsRespuestaGemini {
				continue // un elemento de `$set.messages` que no es un objeto no es una respuesta
			}
			p.Respuestas++
			switch clase {
			case SinIdentificadorGemini:
				p.SinIdentificador++
			case IncoherenteGemini:
				p.Incoherentes++
			case EventoGemini:
				_, emitido := p.emitidos[ev.EventID]
				_, visto := vistos[ev.EventID]
				if emitido || visto {
					p.Repetidas++
					continue
				}
				p.emitidos[ev.EventID] = struct{}{}
				p.Eventos++
				if marcas.SinModelo {
					p.SinModelo++
				}
				if marcas.TotalDescuadrado {
					p.TotalDescuadrado++
				}
				evs = append(evs, *ev)
			}
		}
		return nil
	}, func(leido int64, _ time.Time) int64 { return leido })
	if err != nil {
		return nil, err
	}
	return evs, nil
}

// ContarAnteriorGemini cuenta un `.json` de sesión (Gemini CLI ≤ 0.38) una vez, y otra sólo si cambia (P-009 FR-018,
// D-009-P5). Su entrada en `state.json` lleva los cuatro campos de siempre, con `Offset = Size`: no se abre ni se lee.
func ContarAnteriorGemini(st *state.Store, ruta string, p *PasadaGemini) error {
	info, err := os.Stat(ruta)
	if err != nil {
		return err
	}
	previo, visto := st.Files[ruta]
	if visto && previo.Size == info.Size() && previo.ModTime == info.ModTime().Unix() {
		return nil
	}
	st.Files[ruta] = state.FileState{Path: ruta, Size: info.Size(), ModTime: info.ModTime().Unix(), Offset: info.Size()}
	p.FormatoAnterior++
	return nil
}
