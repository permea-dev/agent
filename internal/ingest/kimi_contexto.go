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

// ═══ P-012 · LOS FICHEROS DE KIMI CODE ══════════════════════════════════════════════════════════════════════════════
//
// Kimi Code guarda cada agente de cada sesión en `<raíz>/sessions/<wd>/<sesión>/agents/<agente>/wire.jsonl` (descubrimiento Q1;
// muestra 1). Junto a ellos hay ficheros que NO son consumo y que el lector no abre nunca (P-3): `state.json` (título y último
// prompt), `logs/`, `notify/`, `tasks/`, `cron/`, `file-history/`, los `.bak` y `.tmp.*` de la reparación y la escritura atómica
// (desc. Q7), y `sessions/.index-cache/` y `.index-dirty/` (m1). `<raíz>/imported/` son exportaciones con forma de sesión
// (desc. Q6): tampoco se leen. Una sesión sin `agents/` no es un error ni se avisa.

// ListarKimi devuelve los `wire.jsonl` de las sesiones de Kimi Code bajo `raiz`, en orden léxico (P-3, FR-003): exactamente
// `<raiz>/sessions/<a>/<b>/agents/<c>/wire.jsonl`, sin recursión, sin las carpetas que empiezan por `.` en ninguno de los tres
// niveles, y sólo si `wire.jsonl` es un fichero regular. Sólo lee directorios: no abre ningún fichero.
func ListarKimi(raiz string) ([]string, error) {
	sesiones := filepath.Join(raiz, "sessions")
	entradas, err := os.ReadDir(sesiones) // ordenadas por nombre
	if err != nil {
		return nil, err
	}
	var rutas []string
	for _, wd := range visiblesKimi(sesiones, entradas) {
		bs, err := subcarpetasKimi(wd)
		if err != nil {
			return nil, err
		}
		for _, sesion := range bs {
			agentes, err := subcarpetasKimi(filepath.Join(sesion, "agents"))
			if err != nil {
				return nil, err
			}
			for _, agente := range agentes {
				fs, err := os.ReadDir(agente)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return nil, err
				}
				for _, f := range fs {
					if f.Name() == "wire.jsonl" && f.Type().IsRegular() {
						rutas = append(rutas, filepath.Join(agente, f.Name()))
					}
				}
			}
		}
	}
	return rutas, nil
}

// subcarpetasKimi devuelve las subcarpetas visibles de `dir`, en orden. Una carpeta que no existe *(una sesión sin `agents/`, o
// borrada entre dos lecturas)* no tiene ninguna.
func subcarpetasKimi(dir string) ([]string, error) {
	entradas, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return visiblesKimi(dir, entradas), nil
}

// visiblesKimi se queda con los directorios de `entradas` que no empiezan por `.`, como rutas bajo `dir`.
func visiblesKimi(dir string, entradas []os.DirEntry) []string {
	var dirs []string
	for _, e := range entradas {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, filepath.Join(dir, e.Name()))
		}
	}
	return dirs
}

// ═══ P-012 B3 · UN FICHERO: ATAJO, CABEZA, PREPASO Y FRONTERA `forked` ═══════════════════════════════════════════
//
// Calcado de `LeerFicheroCopilot` (P-4, E-1): se lee por offset con `state.Recorrer`, y además:
//   - ATAJO: un fichero con el mismo tamaño y el mismo `mtime` que la pasada anterior no se abre;
//   - CABEZA: `state.json` guarda la huella con sal de la primera línea completa. Si el fichero ENCOGE (la reparación de Kimi),
//     si la cabeza CAMBIA (la migración desde una versión anterior a 1.5) o si el estado NO la tiene (la 0.8.0), se relee desde 0
//     (descubrimiento Q7). Como la identidad es la posición (D-3), lo releído da los MISMOS `event_id`;
//   - PREPASO: el fichero entero, sin decodificar nada salvo las `forked`, da el comienzo de cada registro (para N, que cuenta los
//     heredados) y el de la última `forked` (la frontera: lo anterior es heredado y no se emite; muestra 2, M3).
//
// S y A salen de la ruta, `<S>/agents/<A>/wire.jsonl`; el `cwd`, del índice (P-6). NUNCA escribe, crea ni borra nada bajo la raíz.

// prefijoForkedKimi es el comienzo de la línea `forked` (muestra 2, M3). El tipo lo confirma el JSON.
var prefijoForkedKimi = []byte(`{"type":"forked"`)

// PasadaKimi son los recuentos de Kimi Code de UNA pasada (P-012 FR-020): `registros = eventos + heredados + repetidas + a cero +
// incoherentes`; «compactaciones» y «sin modelo» son subconjuntos de «eventos». La clave de los emitidos es el `event_id` ya derivado.
type PasadaKimi struct {
	Registros, Eventos, Heredados, Repetidas, ACero, Incoherentes, Compactaciones, SinModelo int
	// Aperturas cuenta los ficheros abiertos en la pasada (plan §2; tests (14) y (19)).
	Aperturas int
	emitidos  map[string]struct{}
}

// NuevaPasadaKimi crea los recuentos de una pasada.
func NuevaPasadaKimi() *PasadaKimi { return &PasadaKimi{emitidos: make(map[string]struct{})} }

// Resumen es la línea de spec §Textos (P-11): sólo recuentos, nunca identificadores.
func (p *PasadaKimi) Resumen() string {
	return fmt.Sprintf("kimi_code: registros %d · eventos %d · heredados %d · repetidas %d · a cero %d · incoherentes %d · "+
		"compactaciones %d · sin modelo %d", p.Registros, p.Eventos, p.Heredados, p.Repetidas, p.ACero, p.Incoherentes,
		p.Compactaciones, p.SinModelo)
}

// HayNovedades dice si la pasada vio algún registro (FR-021): el demonio sólo escribe la línea entonces.
func (p *PasadaKimi) HayNovedades() bool { return p.Registros > 0 }

// prepasoKimi es lo que da una lectura del fichero entero, sólo de sus líneas completas.
type prepasoKimi struct {
	// cabeza es la huella de la primera línea completa; vacía si no hay ninguna.
	cabeza string
	// inicios es el comienzo de cada línea con el prefijo de registro, también las corruptas (contrato: N las cuenta).
	inicios []int64
	// ultimaForked es el comienzo de la última `forked`, o -1.
	ultimaForked int64
}

// leerPrepasoKimi lee el fichero entero sin tope de línea (FR-009) y sin decodificar nada salvo las líneas `forked`.
func leerPrepasoKimi(ruta, sal string) (prepasoKimi, error) {
	pp := prepasoKimi{ultimaForked: -1}
	f, err := os.Open(ruta)
	if err != nil {
		return pp, err
	}
	defer func() { _ = f.Close() }() // sólo lectura
	r := bufio.NewReader(f)
	var pos int64
	for primera := true; ; primera = false {
		linea, err := r.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			return pp, nil // una línea final sin LF no cuenta
		}
		if err != nil {
			return pp, err
		}
		if primera {
			pp.cabeza = huellaCabezaKimi(sal, linea)
		}
		if bytes.HasPrefix(linea, []byte(prefijoRegistroKimi)) {
			pp.inicios = append(pp.inicios, pos)
		}
		if bytes.HasPrefix(linea, prefijoForkedKimi) && esForkedKimi(linea) {
			pp.ultimaForked = pos
		}
		pos += int64(len(linea))
	}
}

// esForkedKimi confirma por el JSON que una línea es una `forked`.
func esForkedKimi(linea []byte) bool {
	var l struct {
		Type string `json:"type"`
	}
	return json.Unmarshal(linea, &l) == nil && l.Type == "forked"
}

// contarAntesKimi es el número de registros que empiezan antes de `offset`.
func contarAntesKimi(inicios []int64, offset int64) int {
	n := 0
	for _, i := range inicios {
		if i < offset {
			n++
		}
	}
	return n
}

// huellaCabezaKimi es la huella con sal de la primera línea, sin su LF ni un CR final (contrato, E-1): `event.Ref`, 64 hex.
func huellaCabezaKimi(sal string, linea []byte) string {
	return event.Ref(sal, string(bytes.TrimSuffix(bytes.TrimSuffix(linea, []byte("\n")), []byte("\r"))))
}

// sesionYAgenteKimi saca S y A de la ruta de un `wire.jsonl`: `<S>/agents/<A>/wire.jsonl` (P-5).
func sesionYAgenteKimi(ruta string) (string, string) {
	agente := filepath.Dir(ruta)
	return filepath.Base(filepath.Dir(filepath.Dir(agente))), filepath.Base(agente)
}

// LeerFicheroKimi lee lo nuevo de un `wire.jsonl` de Kimi Code y devuelve sus eventos, en orden. Actualiza el estado de ese fichero
// SÓLO en memoria: quien llama encola los eventos y DESPUÉS guarda el estado (FR-024). Un error de lectura deja el estado de ese
// fichero como estaba (FR-026). NUNCA escribe, crea ni borra nada bajo la raíz (FR-006).
func LeerFicheroKimi(st *state.Store, ruta string, base ContextoKimi, p *PasadaKimi, avisos io.Writer) ([]event.Event, error) {
	info, err := os.Stat(ruta)
	if err != nil {
		return nil, err
	}
	previo, existia := st.Files[ruta]
	if info.Size() == previo.Offset && info.ModTime().Unix() == previo.ModTime {
		return nil, nil // sin cambios: no se abre (E-1)
	}
	p.Aperturas++
	pp, err := leerPrepasoKimi(ruta, base.Salt)
	if err != nil {
		return nil, err
	}
	relee := info.Size() < previo.Offset // encogido: la reparación de Kimi
	if previo.Offset > 0 {
		if previo.Cabeza == "" {
			relee = true // un estado sin `cabeza` (la 0.8.0)
		} else if previo.Cabeza != pp.cabeza {
			relee = true // la cabeza cambió: el fichero se reescribió
		}
	}
	if relee {
		st.Files[ruta] = state.FileState{Path: ruta}
	}
	desde := st.Files[ruta].Offset
	previos := contarAntesKimi(pp.inicios, desde)
	base.S, base.A = sesionYAgenteKimi(ruta)
	base.Cwd = normalizarWorkDirKimi(base.Indice[base.S])

	vistos := 0
	var evs []event.Event
	err = st.Recorrer(ruta, func(linea []byte, inicio int64, _ bool) error {
		if !bytes.HasPrefix(linea, []byte(prefijoRegistroKimi)) {
			return nil
		}
		vistos++
		base.N, base.Heredado = previos+vistos, inicio < pp.ultimaForked
		ev, clase, marcas, err := RegistroKimi(linea, base)
		if err != nil {
			if avisos != nil {
				_, _ = fmt.Fprintln(avisos, "skip (línea corrupta):", err) // el texto de `cmd/permea/main.go:320`
			}
			return nil // una línea corrupta se salta: no corta el fichero
		}
		p.Registros++
		switch clase {
		case HeredadoKimi:
			p.Heredados++
		case ACeroKimi:
			p.ACero++
		case IncoherenteKimi:
			p.Incoherentes++
		case EventoKimi:
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
		if existia {
			st.Files[ruta] = previo
		} else {
			delete(st.Files, ruta)
		}
		return nil, err
	}
	e := st.Files[ruta]
	e.Cabeza = pp.cabeza // `Recorrer` rehace la entrada sin ella
	st.Files[ruta] = e
	return evs, nil
}

// ═══ P-012 · EL ÍNDICE: `session_index.jsonl` (P-6, P-13) ═══════════════════════════════════════════════════════════

// LeerIndiceKimi lee `<raiz>/session_index.jsonl` entero y devuelve el `workDir` de cada sesión: gana la ÚLTIMA línea de cada una
// (P-6). De cada línea sólo se leen `sessionId` y `workDir` (FR-012). Un índice que no existe es un mapa vacío; uno que existe y no se
// puede leer, un error (P-13). Una línea que no es JSON se ignora.
func LeerIndiceKimi(raiz string) (map[string]string, error) {
	idx := map[string]string{}
	b, err := os.ReadFile(filepath.Join(raiz, "session_index.jsonl"))
	if err != nil {
		if os.IsNotExist(err) {
			return idx, nil
		}
		return nil, err
	}
	for _, linea := range bytes.Split(b, []byte("\n")) {
		var l struct {
			SessionID string `json:"sessionId"`
			WorkDir   string `json:"workDir"`
		}
		if json.Unmarshal(linea, &l) != nil || l.SessionID == "" {
			continue
		}
		idx[l.SessionID] = l.WorkDir
	}
	return idx, nil
}

// normalizarWorkDirKimi pasa a `\` las `/` de una ruta con forma de Windows (`X:\`, `X:/` o `\\`), para que un fork, que escribe la
// misma ruta con otras barras (muestra 2, M4), dé el mismo proyecto también fuera de Windows (P-6). Las demás rutas, tal cual.
func normalizarWorkDirKimi(w string) string {
	unidad := len(w) >= 3 && (w[0] >= 'A' && w[0] <= 'Z' || w[0] >= 'a' && w[0] <= 'z') && w[1] == ':' && (w[2] == '\\' || w[2] == '/')
	if unidad || strings.HasPrefix(w, `\\`) {
		return strings.ReplaceAll(w, "/", `\`)
	}
	return w
}
