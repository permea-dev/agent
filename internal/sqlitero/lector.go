// Package sqlitero lee tablas de una base SQLite 3 sin escribir nada, solo con la
// libreria estandar: cabecera, WAL (frames hasta el ultimo commit valido),
// arboles de tabla, registros y paginas de desbordamiento.
//
// No cubre: indices, tablas WITHOUT ROWID, texto que no sea UTF-8, valores
// DEFAULT de columnas anadidas con ALTER TABLE (salen NULL), ni bloqueos: la
// coherencia frente a un escritor se comprueba con Estable() y se reintenta.
package sqlitero

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

var be = binary.BigEndian

// Base es una instantanea de solo lectura: fichero principal + WAL ya validado.
type Base struct {
	f         *os.File
	ruta      string
	tamPag    int            // tamano de pagina
	util      int            // bytes utiles por pagina (sin la reserva)
	nPag      uint32         // paginas de la base en esta instantanea
	wal       []byte         // contenido del -wal, si lo hay
	walPag    map[uint32]int // pagina -> desplazamiento de su ultimo frame confirmado
	walHuella []byte         // cabecera + tamano del -wal al abrir, para Estable()
	tamDB     int64
	leidas    int64 // páginas leídas, de la base o del WAL (Paginas, para los tests)
	FramesWAL int   // frames aceptados (hasta el ultimo commit valido)
}

// Abrir lee la cabecera y, si la base esta en modo WAL, valida el -wal.
func Abrir(ruta string) (*Base, error) {
	f, err := os.Open(ruta) // O_RDONLY
	if err != nil {
		return nil, sinRuta(err)
	}
	b := &Base{f: f, ruta: ruta}
	if err := b.cargar(); err != nil {
		_ = f.Close() // sólo lectura: el error de Close no afecta a nada
		return nil, err
	}
	return b, nil
}

// Cerrar cierra la base. Nunca escribe: el fichero se abrió en sólo lectura.
func (b *Base) Cerrar() error { return b.f.Close() }

func (b *Base) cargar() error {
	st, err := b.f.Stat()
	if err != nil {
		return sinRuta(err)
	}
	b.tamDB = st.Size()
	cab := make([]byte, 100)
	if _, err := io.ReadFull(io.NewSectionReader(b.f, 0, 100), cab); err != nil {
		return errors.New("cabecera corta")
	}
	if string(cab[:16]) != "SQLite format 3\x00" {
		return errors.New("no es una base SQLite 3")
	}
	b.tamPag = int(be.Uint16(cab[16:]))
	if b.tamPag == 1 {
		b.tamPag = 65536
	}
	if b.tamPag < 512 || b.tamPag&(b.tamPag-1) != 0 {
		return fmt.Errorf("tamano de pagina invalido %d", b.tamPag)
	}
	b.util = b.tamPag - int(cab[20])
	if b.util < 480 {
		return errors.New("reserva por pagina demasiado grande")
	}
	if enc := be.Uint32(cab[56:]); enc != 1 && enc != 0 {
		return fmt.Errorf("%w: codificación de texto %d (sólo UTF-8)", ErrNoSoportada, enc)
	}
	// Tamano en paginas: el de la cabecera si es valido; si no, el del fichero.
	if n := be.Uint32(cab[28:]); n != 0 && be.Uint32(cab[24:]) == be.Uint32(cab[92:]) {
		b.nPag = n
	} else {
		n := b.tamDB / int64(b.tamPag)
		if n < 0 || n > math.MaxUint32 {
			return errCorta
		}
		b.nPag = uint32(n)
	}
	if cab[18] == 2 || cab[19] == 2 { // modo WAL
		if err := b.leerWAL(); err != nil {
			return err
		}
	}
	return nil
}

// leerWAL acepta los frames con sales y sumas correctas, en cadena, y se queda
// solo con los que llegan hasta el ultimo frame de commit; el mas reciente por pagina.
func (b *Base) leerWAL() error {
	w, err := os.ReadFile(b.ruta + "-wal")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return sinRuta(err)
	}
	b.walHuella = huellaWAL(w)
	if len(w) < 32 {
		return nil
	}
	magia := be.Uint32(w)
	if magia&^1 != 0x377f0682 || be.Uint32(w[4:]) != 3007000 {
		return nil // WAL no valido: se ignora entero, como hace SQLite
	}
	grande := magia&1 == 1 // sumas en big-endian
	if int(be.Uint32(w[8:])) != b.tamPag {
		return nil
	}
	s0, s1 := suma(w[:24], 0, 0, grande)
	if s0 != be.Uint32(w[24:]) || s1 != be.Uint32(w[28:]) {
		return nil
	}
	sal1, sal2 := be.Uint32(w[16:]), be.Uint32(w[20:])
	type frame struct {
		pag, commit uint32
		off         int
	}
	var cadena []frame // frames con sales y sumas en cadena
	ultimoCommit := 0  // frames de cadena[] hasta el ultimo commit, incluido
	tamFrame := 24 + b.tamPag
	for off := 32; off+tamFrame <= len(w); off += tamFrame {
		fr := w[off : off+tamFrame]
		pag, commit := be.Uint32(fr), be.Uint32(fr[4:])
		if pag == 0 || be.Uint32(fr[8:]) != sal1 || be.Uint32(fr[12:]) != sal2 {
			break
		}
		s0, s1 = suma(fr[:8], s0, s1, grande)
		s0, s1 = suma(fr[24:], s0, s1, grande)
		if s0 != be.Uint32(fr[16:]) || s1 != be.Uint32(fr[20:]) {
			break
		}
		cadena = append(cadena, frame{pag, commit, off + 24})
		if commit != 0 {
			ultimoCommit = len(cadena)
		}
	}
	b.wal = w
	b.walPag = map[uint32]int{}
	for _, f := range cadena[:ultimoCommit] { // el mas reciente por pagina gana
		b.walPag[f.pag] = f.off
	}
	if ultimoCommit > 0 {
		b.nPag = cadena[ultimoCommit-1].commit
	}
	b.FramesWAL = ultimoCommit
	return nil
}

func huellaWAL(w []byte) []byte {
	h := make([]byte, 0, 40)
	h = binary.BigEndian.AppendUint64(h, uint64(len(w)))
	return append(h, w[:min(32, len(w))]...)
}

// suma es la suma de comprobacion del WAL sobre palabras de 32 bits.
func suma(d []byte, s0, s1 uint32, grande bool) (uint32, uint32) {
	var bo binary.ByteOrder = binary.LittleEndian
	if grande {
		bo = binary.BigEndian
	}
	for i := 0; i+8 <= len(d); i += 8 {
		s0 += bo.Uint32(d[i:]) + s1
		s1 += bo.Uint32(d[i+4:]) + s0
	}
	return s0, s1
}

// Estable dice si el -wal y la base siguen como al abrir: si no, la lectura
// pudo mezclar estados y hay que repetirla.
func (b *Base) Estable() bool {
	w, err := os.ReadFile(b.ruta + "-wal")
	var h []byte
	if err == nil {
		h = huellaWAL(w)
	}
	st, err := os.Stat(b.ruta)
	return err == nil && st.Size() == b.tamDB && bytes.Equal(h, b.walHuella)
}

func (b *Base) pagina(n uint32) ([]byte, error) {
	if n < 1 || n > b.nPag {
		return nil, fmt.Errorf("%w: página %d fuera de la base", ErrCorrupta, n)
	}
	b.leidas++
	if off, ok := b.walPag[n]; ok {
		return b.wal[off : off+b.tamPag], nil
	}
	p := make([]byte, b.tamPag)
	if _, err := b.f.ReadAt(p, int64(n-1)*int64(b.tamPag)); err != nil {
		return nil, fmt.Errorf("página %d: %w", n, sinRuta(err))
	}
	return p, nil
}

// varint lee un entero de longitud variable de SQLite (1..9 bytes).
// Devuelve los 64 bits como int64, como SQLite: el rowid es un entero con signo.
func varint(d []byte) (int64, int) {
	var v int64
	for i := 0; i < 8; i++ {
		if i >= len(d) {
			return 0, 0
		}
		v = v<<7 | int64(d[i]&0x7f)
		if d[i]&0x80 == 0 {
			return v, i + 1
		}
	}
	if len(d) < 9 {
		return 0, 0
	}
	return v<<8 | int64(d[8]), 9
}

var errCorta = fmt.Errorf("%w o truncada", ErrCorrupta)

// todas es el «desde» de un recorrido completo: ningún rowid es menor.
const todas = math.MinInt64

// recorrer visita en orden de rowid las filas del arbol de tabla con raiz en raiz que tienen rowid > desde. Poda las
// celdas interiores con clave ≤ desde: su hijo sólo tiene rowids ≤ clave (P-010 §P-3, SC-017).
func (b *Base) recorrer(raiz uint32, desde int64, fn func(rowid int64, carga []byte) error) error {
	vistas := map[uint32]bool{}
	var rec func(n uint32, prof int) error
	rec = func(n uint32, prof int) error {
		if vistas[n] || prof > 64 {
			return fmt.Errorf("%w: ciclo en el árbol", ErrCorrupta)
		}
		vistas[n] = true
		p, err := b.pagina(n)
		if err != nil {
			return err
		}
		h := 0
		if n == 1 {
			h = 100
		}
		tipo, nCeldas := p[h], int(be.Uint16(p[h+3:]))
		switch tipo {
		case 0x05: // interior de tabla
			ptrs := h + 12
			if ptrs+2*nCeldas > b.util {
				return errCorta
			}
			for i := 0; i < nCeldas; i++ {
				c := int(be.Uint16(p[ptrs+2*i:]))
				if c+4 > b.util {
					return errCorta
				}
				clave, k := varint(p[c+4 : b.util])
				if k == 0 {
					return errCorta
				}
				if clave <= desde {
					continue // poda: el hijo izquierdo sólo tiene rowids <= clave
				}
				if err := rec(be.Uint32(p[c:]), prof+1); err != nil {
					return err
				}
			}
			return rec(be.Uint32(p[h+8:]), prof+1)
		case 0x0d: // hoja de tabla
			ptrs := h + 8
			if ptrs+2*nCeldas > b.util {
				return errCorta
			}
			for i := 0; i < nCeldas; i++ {
				c := int(be.Uint16(p[ptrs+2*i:]))
				if c >= b.util {
					return errCorta
				}
				tam, n1 := varint(p[c:b.util])
				rowid, n2 := varint(p[c+n1 : b.util])
				if n1 == 0 || n2 == 0 {
					return errCorta
				}
				if rowid <= desde {
					continue
				}
				carga, err := b.carga(p, c+n1+n2, int(tam))
				if err != nil {
					return err
				}
				if err := fn(rowid, carga); err != nil {
					return err
				}
			}
			return nil
		case 0x02, 0x0a:
			return fmt.Errorf("%w: árbol de índice (tabla WITHOUT ROWID)", ErrNoSoportada)
		default:
			return fmt.Errorf("%w: tipo de página %#x desconocido", ErrCorrupta, tipo)
		}
	}
	return rec(raiz, 0)
}

// carga junta la parte local de la celda y sus paginas de desbordamiento.
func (b *Base) carga(p []byte, ini, total int) ([]byte, error) {
	u := b.util
	if total < 0 || int64(total) > int64(b.nPag)*int64(b.tamPag) {
		return nil, errCorta
	}
	local := total
	if x := u - 35; total > x {
		m := (u-12)*32/255 - 23
		k := m + (total-m)%(u-4)
		local = m
		if k <= x {
			local = k
		}
	}
	if ini+local > u || (local < total && ini+local+4 > u) {
		return nil, errCorta
	}
	out := make([]byte, 0, total)
	out = append(out, p[ini:ini+local]...)
	if local == total {
		return out, nil
	}
	sig := be.Uint32(p[ini+local:])
	for resto, saltos := total-local, uint32(0); resto > 0; saltos++ {
		if sig == 0 || saltos > b.nPag {
			return nil, errCorta
		}
		op, err := b.pagina(sig)
		if err != nil {
			return nil, err
		}
		n := min(resto, u-4)
		out = append(out, op[4:4+n]...)
		resto -= n
		sig = be.Uint32(op)
	}
	return out, nil
}

// Tipo de un valor almacenado.
type Tipo int

// Los tipos de un Valor: los de SQLite.
const (
	Nulo Tipo = iota
	Entero
	Real
	Texto
	Blob
)

// Valor es una columna de un registro.
type Valor struct {
	Tipo Tipo
	I    int64
	F    float64
	B    []byte // texto (UTF-8) o blob
}

// Num da el valor numerico de un entero o real; 0 en otro caso.
func (v Valor) Num() float64 {
	if v.Tipo == Entero {
		return float64(v.I)
	}
	return v.F
}

// registro descodifica el formato de registro: cabecera de tipos de serie y cuerpo.
func registro(d []byte) ([]Valor, error) {
	tamCab, n := varint(d)
	if n == 0 || tamCab < 0 || tamCab > int64(len(d)) {
		return nil, errCorta
	}
	var vals []Valor
	finCab := len(d[:tamCab]) // == tamCab, sin conversión
	cuerpo := finCab
	for i := n; i < finCab; {
		st, k := varint(d[i:finCab])
		if k == 0 {
			return nil, errCorta
		}
		i += k
		var v Valor
		lon := 0
		switch {
		case st == 0:
		case st >= 1 && st <= 6:
			switch st {
			case 5:
				lon = 6
			case 6:
				lon = 8
			default:
				lon = int(st)
			}
			v.Tipo = Entero
		case st == 7:
			lon, v.Tipo = 8, Real
		case st == 8 || st == 9:
			v.Tipo, v.I = Entero, int64(st-8)
		case st >= 12:
			lon = int((st - 12) / 2)
			v.Tipo = Blob
			if st%2 == 1 {
				v.Tipo = Texto
			}
		default:
			return nil, errCorta
		}
		if cuerpo+lon > len(d) {
			return nil, errCorta
		}
		c := d[cuerpo : cuerpo+lon]
		switch {
		case v.Tipo == Entero && lon > 0:
			x := int64(c[0])
			if x >= 0x80 {
				x -= 0x100 // con signo
			}
			for _, by := range c[1:] {
				x = x<<8 | int64(by)
			}
			v.I = x
		case v.Tipo == Real:
			v.F = math.Float64frombits(be.Uint64(c))
		case v.Tipo == Texto || v.Tipo == Blob:
			v.B = c
		}
		cuerpo += lon
		vals = append(vals, v)
	}
	return vals, nil
}

// Tabla es una tabla con rowid: raiz y nombres de columna sacados del CREATE TABLE.
type Tabla struct {
	b        *Base
	raiz     uint32
	columnas []string
	alias    int // columna INTEGER PRIMARY KEY (alias del rowid), o -1
}

// Tabla busca nombre en sqlite_schema.
func (b *Base) Tabla(nombre string) (*Tabla, error) {
	var t *Tabla
	err := b.recorrer(1, todas, func(_ int64, carga []byte) error {
		v, err := registro(carga)
		if err != nil {
			return err
		}
		if len(v) < 5 || string(v[0].B) != "table" || !strings.EqualFold(string(v[1].B), nombre) {
			return nil
		}
		sql := string(v[4].B)
		cols, alias, err := columnas(sql)
		if err != nil {
			return err
		}
		raiz := v[3].I
		if raiz < 1 || raiz > math.MaxUint32 {
			return errCorta
		}
		t = &Tabla{b: b, raiz: uint32(raiz), columnas: cols, alias: alias}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("no existe la tabla %s", nombre)
	}
	return t, nil
}

// Col da la posicion de una columna, o -1.
func (t *Tabla) Col(nombre string) int {
	for i, c := range t.columnas {
		if strings.EqualFold(c, nombre) {
			return i
		}
	}
	return -1
}

// Recorrer entrega cada fila con tantas columnas como el esquema; las que el
// registro no trae (ALTER TABLE ADD COLUMN) salen NULL.
func (t *Tabla) Recorrer(fn func(fila []Valor) error) error {
	return t.RecorrerDesde(todas, func(_ int64, fila []Valor) error { return fn(fila) })
}

// fila descodifica una carga con tantas columnas como el esquema.
func (t *Tabla) fila(rowid int64, carga []byte) ([]Valor, error) {
	v, err := registro(carga)
	if err != nil {
		return nil, err
	}
	fila := make([]Valor, len(t.columnas))
	copy(fila, v)
	if t.alias >= 0 {
		fila[t.alias] = Valor{Tipo: Entero, I: rowid}
	}
	return fila, nil
}

// columnas saca los nombres de columna de un CREATE TABLE.
func columnas(sql string) ([]string, int, error) {
	if strings.Contains(strings.ToUpper(sql), "WITHOUT ROWID") {
		return nil, -1, fmt.Errorf("%w: tabla WITHOUT ROWID", ErrNoSoportada)
	}
	a, z := strings.IndexByte(sql, '('), strings.LastIndexByte(sql, ')')
	if a < 0 || z < a {
		return nil, -1, errors.New("CREATE TABLE no reconocido")
	}
	var defs []string
	prof, ini := 0, a+1
	var comilla byte
	for i := a + 1; i < z; i++ {
		c := sql[i]
		switch {
		case comilla != 0:
			if c == comilla {
				comilla = 0
			}
		case c == '\'' || c == '"' || c == '`':
			comilla = c
		case c == '[':
			comilla = ']'
		case c == '(':
			prof++
		case c == ')':
			prof--
		case c == ',' && prof == 0:
			defs = append(defs, sql[ini:i])
			ini = i + 1
		}
	}
	defs = append(defs, sql[ini:z])
	var cols []string
	alias := -1
	for _, d := range defs {
		f := strings.Fields(d)
		if len(f) == 0 {
			continue
		}
		switch strings.ToUpper(f[0]) {
		case "CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN":
			continue
		}
		up := strings.ToUpper(d)
		if len(f) > 1 && strings.ToUpper(f[1]) == "INTEGER" && strings.Contains(up, "PRIMARY KEY") {
			alias = len(cols)
		}
		cols = append(cols, strings.Trim(f[0], "`\"[]'"))
	}
	return cols, alias, nil
}

// ═══ P-010 B1 · LA API NUEVA (plan §1) ══════════════════════════════════════════════════════════════════════════════
//
// El lector de arriba es el de la prueba (`soporte/prueba-sqlite.md`), portado con su lógica intacta: cabecera, WAL
// hasta el último commit válido, registros, desbordamiento, columnas y Estable(). B1 añade el recorrido desde un rowid
// con poda, la búsqueda por descenso, el contador de páginas y los errores tipados y sin ruta.

// ErrCorrupta marca una estructura imposible: ciclos, punteros fuera de la página o de la base, registros cortos.
var ErrCorrupta = errors.New("base corrupta")

// ErrNoSoportada marca una base válida que el lector no lee: tablas WITHOUT ROWID o texto UTF-16.
var ErrNoSoportada = errors.New("formato no soportado")

// sinRuta quita la ruta de un error de E/S: en Windows lleva el nombre del usuario (spec FR-029).
func sinRuta(err error) error {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return fmt.Errorf("%s: %w", pe.Op, pe.Err)
	}
	return err
}

// RecorrerDesde visita, en orden, sólo las filas con rowid > h. Cuesta la profundidad del árbol más lo nuevo.
func (t *Tabla) RecorrerDesde(h int64, fn func(rowid int64, fila []Valor) error) error {
	return t.b.recorrer(t.raiz, h, func(rowid int64, carga []byte) error {
		fila, err := t.fila(rowid, carga)
		if err != nil {
			return err
		}
		return fn(rowid, fila)
	})
}

// Buscar devuelve la fila con ese rowid bajando por las páginas interiores: en cada una, la primera celda con clave
// ≥ rowid, o el puntero derecho. ok es falso, sin error, si no está.
func (t *Tabla) Buscar(rowid int64) ([]Valor, bool, error) {
	b, n := t.b, t.raiz
	for prof := 0; prof <= 64; prof++ {
		p, err := b.pagina(n)
		if err != nil {
			return nil, false, err
		}
		h := 0
		if n == 1 {
			h = 100
		}
		tipo, nCeldas := p[h], int(be.Uint16(p[h+3:]))
		switch tipo {
		case 0x05:
			ptrs := h + 12
			if ptrs+2*nCeldas > b.util {
				return nil, false, errCorta
			}
			sig := be.Uint32(p[h+8:])
			for i := 0; i < nCeldas; i++ {
				c := int(be.Uint16(p[ptrs+2*i:]))
				if c+4 > b.util {
					return nil, false, errCorta
				}
				clave, k := varint(p[c+4 : b.util])
				if k == 0 {
					return nil, false, errCorta
				}
				if clave >= rowid {
					sig = be.Uint32(p[c:])
					break
				}
			}
			n = sig
		case 0x0d:
			ptrs := h + 8
			if ptrs+2*nCeldas > b.util {
				return nil, false, errCorta
			}
			for i := 0; i < nCeldas; i++ {
				c := int(be.Uint16(p[ptrs+2*i:]))
				if c >= b.util {
					return nil, false, errCorta
				}
				tam, n1 := varint(p[c:b.util])
				rid, n2 := varint(p[c+n1 : b.util])
				if n1 == 0 || n2 == 0 {
					return nil, false, errCorta
				}
				if rid != rowid {
					continue
				}
				carga, err := b.carga(p, c+n1+n2, int(tam))
				if err != nil {
					return nil, false, err
				}
				fila, err := t.fila(rowid, carga)
				return fila, err == nil, err
			}
			return nil, false, nil
		case 0x02, 0x0a:
			return nil, false, fmt.Errorf("%w: árbol de índice (tabla WITHOUT ROWID)", ErrNoSoportada)
		default:
			return nil, false, fmt.Errorf("%w: tipo de página %#x desconocido", ErrCorrupta, tipo)
		}
	}
	return nil, false, fmt.Errorf("%w: árbol demasiado profundo", ErrCorrupta)
}

// Paginas dice cuántas páginas se han leído, de la base o del WAL, desde que se abrió la base. Es para los tests.
func (b *Base) Paginas() int64 { return b.leidas }
