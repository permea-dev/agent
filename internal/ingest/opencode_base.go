package ingest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/sqlitero"
	"github.com/permea-dev/agent/internal/state"
)

// ═══ P-010 B4 · UNA PASADA POR UNA BASE DE OPENCODE, CON SU FRONTERA POR ROWID (plan §2) ═══════════════════════════
//
// SQLite serializa a los escritores: una fila confirmada después tiene un rowid mayor. Por eso basta guardar H, el mayor
// rowid visto, con la huella de su `id`, y la lista de las filas `assistant` vistas en curso. Una pasada comprueba H,
// consulta la lista por rowid y recorre desde H; si H falta, o es otra fila (un rowid reutilizado, un `VACUUM`), recorre
// la tabla entera y cuenta «frontera reiniciada». Las ya emitidas vuelven a salir con su `event_id`, y la plataforma
// las descarta: nunca se pierde una completada.
//
// Todo se clasifica EN MEMORIA, sobre una instantánea, y sólo cuenta si `Estable()` (D-010-P6): si no, se relee, hasta
// 5 veces. Un error deja la entrada de la base y los recuentos como estaban (FR-029).

const (
	// abandonoOpenCode es lo que se espera a una fila en curso: con más, se abandona y nunca se emite (P-3).
	abandonoOpenCode = 24 * time.Hour
	// intentosOpenCode y esperaOpenCode son las lecturas de una base que cambia mientras se lee (FR-008).
	intentosOpenCode = 5
	esperaOpenCode   = 50 * time.Millisecond
)

// errInestableOpenCode es una base que cambió en todas las lecturas.
var errInestableOpenCode = fmt.Errorf("la base cambió durante la lectura %d veces", intentosOpenCode)

// antesDeEstableOpenCode es un gancho para los tests de SC-029: se llama entre la lectura y `Estable()`. Nil fuera de ellos.
var antesDeEstableOpenCode func()

// PasadaOpenCode son los recuentos de OpenCode de UNA pasada (P-010 FR-021) y su línea de resumen (§Textos). Cumplen,
// por construcción, `Llamadas = Eventos + SinIdentificador + Incoherentes`; `SinModelo` cuenta eventos.
type PasadaOpenCode struct {
	Llamadas, Eventos, EnCurso, Abandonadas, SinIdentificador, Incoherentes, SinModelo, Reiniciadas int
	FormatoAnterior, EsquemaDesconocido                                                             int

	// paginas son las páginas leídas en las pasadas que acabaron bien: para los tests (SC-011, SC-017).
	paginas int64
}

// NuevaPasadaOpenCode crea los recuentos de una pasada.
func NuevaPasadaOpenCode() *PasadaOpenCode { return &PasadaOpenCode{} }

// Resumen es la línea de spec §Textos: sólo recuentos, nunca identificadores.
func (p *PasadaOpenCode) Resumen() string {
	return fmt.Sprintf("opencode: llamadas %d · eventos %d · en curso %d · abandonadas %d · sin identificador %d · "+
		"incoherentes %d · sin modelo %d · fronteras reiniciadas %d · bases en formato anterior %d · "+
		"bases con esquema desconocido %d", p.Llamadas, p.Eventos, p.EnCurso, p.Abandonadas, p.SinIdentificador,
		p.Incoherentes, p.SinModelo, p.Reiniciadas, p.FormatoAnterior, p.EsquemaDesconocido)
}

// HayNovedades dice si el demonio escribe la línea de OpenCode en este ciclo (FR-024): si hubo llamadas, abandonadas,
// fronteras reiniciadas o bases contadas. Una fila que se queda en curso no la repite en cada ciclo.
func (p *PasadaOpenCode) HayNovedades() bool {
	return p.Llamadas > 0 || p.Abandonadas > 0 || p.Reiniciadas > 0 || p.FormatoAnterior > 0 || p.EsquemaDesconocido > 0
}

// sumar añade los recuentos de una base.
func (p *PasadaOpenCode) sumar(d PasadaOpenCode) {
	p.Llamadas += d.Llamadas
	p.Eventos += d.Eventos
	p.EnCurso += d.EnCurso
	p.Abandonadas += d.Abandonadas
	p.SinIdentificador += d.SinIdentificador
	p.Incoherentes += d.Incoherentes
	p.SinModelo += d.SinModelo
	p.Reiniciadas += d.Reiniciadas
	p.paginas += d.paginas
}

// FormatoOpenCode es una base que no se lee, sólo se cuenta (P-9, FR-022, FR-023).
type FormatoOpenCode int

// Los formatos que se cuentan.
const (
	// FormatoAnteriorOpenCode es una base 1.x: sin `session_message`, con `message` y `part` (P-9).
	FormatoAnteriorOpenCode FormatoOpenCode = iota + 1
	// EsquemaDesconocidoOpenCode es una base 2.x que el lector no reconoce: `WITHOUT ROWID`, UTF-16 o una columna que
	// falta (FR-023, Q-3).
	EsquemaDesconocidoOpenCode
)

// ListarOpenCode enumera las bases de una carpeta de datos de OpenCode (P-8, FR-003): `opencode.db` y cada
// `opencode-<canal>.db`, con el canal en `[A-Za-z0-9._-]`, menos `opencode-next.db`; en orden léxico. Ni `-wal`, ni `-shm`,
// ni `log/` o `storage/`.
func ListarOpenCode(dir string) ([]string, error) {
	entradas, err := os.ReadDir(dir) // ya en orden léxico
	if err != nil {
		return nil, err
	}
	var bases []string
	for _, e := range entradas {
		if n := e.Name(); !e.IsDir() && esBaseOpenCode(n) && n != "opencode-next.db" {
			bases = append(bases, filepath.Join(dir, n))
		}
	}
	return bases, nil
}

// caracteresCanalOpenCode son los del canal de `opencode-<canal>.db` (P-8).
const caracteresCanalOpenCode = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-"

// esBaseOpenCode dice si un nombre es `opencode(-[A-Za-z0-9._-]+)?\.db`.
func esBaseOpenCode(n string) bool {
	resto, empieza := strings.CutPrefix(n, "opencode")
	canal, acaba := strings.CutSuffix(resto, ".db")
	if !empieza || !acaba {
		return false
	}
	return canal == "" || len(canal) > 1 && canal[0] == '-' && strings.Trim(canal[1:], caracteresCanalOpenCode) == ""
}

// ContarFormatoOpenCode cuenta una base que no se lee una vez, y otra sólo si cambia (FR-022, FR-023, D-010-P8): con la
// regla `Size`/`ModTime` de D-009-P5. Su entrada en `state.json` lleva los cuatro campos, con `Offset = Size`.
func ContarFormatoOpenCode(st *state.Store, ruta string, f FormatoOpenCode, p *PasadaOpenCode) error {
	info, err := os.Stat(ruta)
	if err != nil {
		return sinRutaOpenCode(err)
	}
	previo, visto := st.Files[ruta]
	if visto && previo.Size == info.Size() && previo.ModTime == info.ModTime().Unix() {
		return nil
	}
	st.Files[ruta] = state.FileState{Path: ruta, Size: info.Size(), ModTime: info.ModTime().Unix(), Offset: info.Size()}
	if f == FormatoAnteriorOpenCode {
		p.FormatoAnterior++
	} else {
		p.EsquemaDesconocido++
	}
	return nil
}

// sinRutaOpenCode quita la ruta de un error de E/S: en Windows lleva el nombre del usuario (FR-029).
func sinRutaOpenCode(err error) error {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return fmt.Errorf("%s: %w", pe.Op, pe.Err)
	}
	return err
}

// LeerBaseOpenCode hace la pasada de una base y devuelve sus eventos, en orden. Actualiza su entrada de `state.json`
// SÓLO en memoria: quien llama encola los eventos y DESPUÉS guarda el estado (FR-027). Un error lleva el NOMBRE de la
// base, no su ruta (D-010-P7), y deja su entrada y los recuentos como estaban (FR-029); quien llama decide si sigue.
func LeerBaseOpenCode(st *state.Store, ruta string, base ContextoOpenCode, p *PasadaOpenCode) ([]event.Event, error) {
	var l *lecturaOpenCode
	for intento := 1; intento <= intentosOpenCode; intento++ {
		if intento > 1 {
			time.Sleep(esperaOpenCode)
		}
		var err error
		if l, err = leerBaseOpenCode(ruta, st.Files[ruta], base); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(ruta), err)
		}
		if l.formato != 0 {
			if err := ContarFormatoOpenCode(st, ruta, l.formato, p); err != nil {
				return nil, fmt.Errorf("%s: %w", filepath.Base(ruta), err)
			}
			return nil, nil
		}
		if l.estable {
			st.Files[ruta] = l.estado
			p.sumar(l.cuenta)
			return l.eventos, nil
		}
	}
	return nil, fmt.Errorf("%s: %w", filepath.Base(ruta), errInestableOpenCode)
}

// lecturaOpenCode es lo que dio una lectura de una base: o un formato que sólo se cuenta, o sus eventos, sus recuentos y
// su entrada nueva, que sólo valen si la base no cambió mientras se leía.
type lecturaOpenCode struct {
	formato FormatoOpenCode
	eventos []event.Event
	cuenta  PasadaOpenCode
	estado  state.FileState
	estable bool
}

// leerBaseOpenCode lee una instantánea de la base y la clasifica en memoria.
func leerBaseOpenCode(ruta string, previo state.FileState, base ContextoOpenCode) (*lecturaOpenCode, error) {
	info, err := os.Stat(ruta)
	if err != nil {
		return nil, sinRutaOpenCode(err)
	}
	b, err := sqlitero.Abrir(ruta)
	if errors.Is(err, sqlitero.ErrNoSoportada) { // UTF-16
		return &lecturaOpenCode{formato: EsquemaDesconocidoOpenCode}, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = b.Cerrar() }() // sólo lectura: el error de Close no afecta a nada
	msgs, ses, f, err := tablasOpenCode(b)
	if err != nil || f != 0 {
		return &lecturaOpenCode{formato: f}, err
	}

	c := &clasificadorOpenCode{sal: base.Salt, ahora: time.Now(), colID: msgs.Col("id"), colSesion: msgs.Col("session_id"),
		colTipo: msgs.Col("type"), colCreada: msgs.Col("time_created"), colData: msgs.Col("data")}
	if base.Ahora != nil {
		c.ahora = base.Ahora()
	}
	desde := int64(math.MinInt64) // sin `frontera`: pasada completa (la primera, o una entrada de la 0.6.0)
	if fr := previo.Frontera; fr != nil {
		valida, err := c.consultar(msgs, fr, previo.Offset)
		if err != nil {
			return nil, err
		}
		if valida {
			desde, c.h, c.huellaH = previo.Offset, previo.Offset, fr.Huella
		} else {
			c.cuenta.Reiniciadas++
			c.pendientes, c.lista, c.cuenta.Abandonadas = nil, nil, 0
		}
	}
	err = msgs.RecorrerDesde(desde, func(rowid int64, fila []sqlitero.Valor) error {
		c.h, c.huellaH = rowid, c.huella(fila)
		c.clasificar(rowid, fila)
		return nil
	})
	if err != nil {
		return nil, err
	}
	l := &lecturaOpenCode{}
	if len(c.pendientes) > 0 { // `session_v2`, una vez por pasada y base, sólo si hay llamadas (D-010-P5)
		sesiones, err := sesionesOpenCode(ses)
		if err != nil {
			return nil, err
		}
		l.eventos = c.emitir(sesiones, base)
	}
	c.cuenta.EnCurso = len(c.lista)
	c.cuenta.paginas = b.Paginas()
	l.cuenta = c.cuenta
	l.estado = state.FileState{Path: ruta, Size: info.Size(), ModTime: info.ModTime().Unix(), Offset: c.h,
		Frontera: &state.Frontera{Huella: c.huellaH, EnCurso: c.lista}}
	if antesDeEstableOpenCode != nil {
		antesDeEstableOpenCode()
	}
	l.estable = b.Estable()
	return l, nil
}

// tablasOpenCode busca las dos tablas que se leen y sus columnas (FR-003, FR-009, FR-023). Un error de `Tabla` que no
// envuelve a otro —la tabla no existe, o su `CREATE` no se reconoce— o `ErrNoSoportada` (`WITHOUT ROWID`) son de esquema;
// los demás (corrupta, E/S), de la base.
func tablasOpenCode(b *sqlitero.Base) (msgs, ses *sqlitero.Tabla, f FormatoOpenCode, err error) {
	msgs, err = b.Tabla("session_message")
	if err == nil {
		ses, err = b.Tabla("session_v2")
	}
	switch {
	case err == nil:
	case !errors.Is(err, sqlitero.ErrNoSoportada) && errors.Unwrap(err) != nil:
		return nil, nil, 0, err
	case msgs == nil && existeTablaOpenCode(b, "message") && existeTablaOpenCode(b, "part"):
		return nil, nil, FormatoAnteriorOpenCode, nil
	default:
		return nil, nil, EsquemaDesconocidoOpenCode, nil
	}
	for _, c := range []struct {
		t   *sqlitero.Tabla
		col string
	}{{msgs, "id"}, {msgs, "session_id"}, {msgs, "type"}, {msgs, "time_updated"}, {msgs, "data"}, {ses, "id"}, {ses, "directory"}} {
		if c.t.Col(c.col) < 0 {
			return nil, nil, EsquemaDesconocidoOpenCode, nil
		}
	}
	return msgs, ses, 0, nil
}

func existeTablaOpenCode(b *sqlitero.Base, nombre string) bool {
	_, err := b.Tabla(nombre)
	return err == nil
}

// filaSesionOpenCode es lo que se lee de una fila de `session_v2` (P-6).
type filaSesionOpenCode struct{ directory, path string }

// sesionesOpenCode lee `session_v2` entera: cuesta el número de sesiones, no el de mensajes (D-010-P5).
func sesionesOpenCode(ses *sqlitero.Tabla) (map[string]filaSesionOpenCode, error) {
	colID, colDir, colPath := ses.Col("id"), ses.Col("directory"), ses.Col("path")
	m := make(map[string]filaSesionOpenCode)
	err := ses.Recorrer(func(fila []sqlitero.Valor) error {
		s := filaSesionOpenCode{directory: textoDeValor(fila[colDir])}
		if colPath >= 0 {
			s.path = textoDeValor(fila[colPath])
		}
		m[textoDeValor(fila[colID])] = s
		return nil
	})
	return m, err
}

// textoDeValor es el texto de una columna, y "" si es de otro tipo.
func textoDeValor(v sqlitero.Valor) string {
	if v.Tipo != sqlitero.Texto {
		return ""
	}
	return string(v.B)
}

// clasificadorOpenCode reúne lo que una lectura va viendo: H y su huella, la lista en curso, las llamadas a la espera de
// `session_v2` y los recuentos.
type clasificadorOpenCode struct {
	sal                                           string
	ahora                                         time.Time
	colID, colSesion, colTipo, colCreada, colData int

	h          int64
	huellaH    string
	lista      []state.EnCurso
	pendientes []pendienteOpenCode
	cuenta     PasadaOpenCode
}

// pendienteOpenCode es una llamada clasificada, a la espera de su sesión.
type pendienteOpenCode struct {
	fila   FilaOpenCode
	sesion string
}

// huella es la del `id` de una fila, con sal (D-010-P3): `state.json` no guarda el `id`.
func (c *clasificadorOpenCode) huella(fila []sqlitero.Valor) string {
	return event.Ref(c.sal, textoDeValor(fila[c.colID]))
}

// consultar comprueba la frontera (§P-3, paso 1) y consulta la lista por rowid (paso 2), sin barrido. Devuelve falso si
// hay que reiniciarla.
func (c *clasificadorOpenCode) consultar(msgs *sqlitero.Tabla, fr *state.Frontera, h int64) (bool, error) {
	fila, ok, err := msgs.Buscar(h)
	if err != nil {
		return false, err
	}
	switch {
	case !ok:
		return false, nil // la fila H ya no está: se borró, o la renumeró un `VACUUM`
	case c.huella(fila) != fr.Huella:
		return false, nil // otra fila en el rowid H
	}
	for _, e := range fr.EnCurso {
		fila, ok, err := msgs.Buscar(e.Rowid)
		if err != nil {
			return false, err
		}
		if !ok {
			continue // la sesión se borró: sale de la lista sin contarse
		}
		if c.huella(fila) != e.Huella {
			return false, nil
		}
		c.clasificar(e.Rowid, fila)
	}
	return true, nil
}

// tiempoOpenCode son los momentos de una fila, en crudo.
type tiempoOpenCode struct {
	Time struct {
		Created   json.RawMessage `json:"created"`
		Completed json.RawMessage `json:"completed"`
	} `json:"time"`
}

// clasificar mira una fila nueva, o de la lista: una `assistant` completada, o cuyo `data` no es un objeto, es una llamada
// (P-1, FR-012); si no, sigue en curso, o se abandona. Otro `type`, nada. Lo completado se mira ANTES que el abandono.
func (c *clasificadorOpenCode) clasificar(rowid int64, fila []sqlitero.Valor) {
	if textoDeValor(fila[c.colTipo]) != "assistant" {
		return
	}
	// `data` se analiza UNA vez por fila: lo leído pasa a `LlamadaOpenCode`, y la llamada no retiene la fila (SC-017). Da
	// error en los mismos casos que `tiempoOpenCode`: sus campos de más son `json.RawMessage`.
	var d datosOpenCode
	data := fila[c.colData].B
	err := json.Unmarshal(data, &d)
	switch {
	case err != nil || !esObjeto(bytes.TrimSpace(data)) || !ausente(d.Time.Completed):
		id := fila[c.colID]
		id.B = bytes.Clone(id.B)
		c.pendientes = append(c.pendientes, pendienteOpenCode{FilaOpenCode{ID: id, leida: &leidaOpenCode{d, err}},
			textoDeValor(fila[c.colSesion])})
	case c.abandonada(fila):
		c.cuenta.Abandonadas++
	default:
		c.lista = append(c.lista, state.EnCurso{Rowid: rowid, Huella: c.huella(fila)})
	}
}

// abandonada dice si una fila en curso lleva más de 24 h: desde `$.time.created` o, si falta, la columna `time_created`.
func (c *clasificadorOpenCode) abandonada(fila []sqlitero.Valor) bool {
	var d tiempoOpenCode
	var creada int64
	if json.Unmarshal(fila[c.colData].B, &d) != nil || ausente(d.Time.Created) || json.Unmarshal(d.Time.Created, &creada) != nil {
		creada = 0
		if c.colCreada >= 0 && fila[c.colCreada].Tipo == sqlitero.Entero {
			creada = fila[c.colCreada].I
		}
	}
	return c.ahora.Sub(time.UnixMilli(creada)) > abandonoOpenCode
}

// emitir convierte las llamadas en eventos, con su sesión (P-6), y las cuenta (FR-021).
func (c *clasificadorOpenCode) emitir(sesiones map[string]filaSesionOpenCode, base ContextoOpenCode) []event.Event {
	if base.Resolutor == nil { // sin resolutor, uno para esta base: la caché optimiza, no decide (P-004; SC-017)
		base.Resolutor = project.NuevoResolutor()
	}
	evs := make([]event.Event, 0, len(c.pendientes))
	for _, pd := range c.pendientes {
		ctx := base
		ctx.SessionID, ctx.Directory, ctx.Path = "", "", ""
		if s, ok := sesiones[pd.sesion]; ok {
			ctx.SessionID, ctx.Directory, ctx.Path = pd.sesion, s.directory, s.path
		}
		ev, clase := LlamadaOpenCode(pd.fila, ctx)
		c.cuenta.Llamadas++
		switch clase {
		case SinIdentificadorOpenCode:
			c.cuenta.SinIdentificador++
		case IncoherenteOpenCode:
			c.cuenta.Incoherentes++
		case EventoOpenCode:
			c.cuenta.Eventos++
			if ev.Model == "" {
				c.cuenta.SinModelo++
			}
			evs = append(evs, *ev)
		}
	}
	return evs
}
