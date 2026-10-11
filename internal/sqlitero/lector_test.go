package sqlitero

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// ═══ P-010 B1 · ROJOS (1)–(10) DEL LECTOR (plan §Bloques, fila B1) ════════════════════════════════════════════
//
// Los fixtures son SINTÉTICOS (testdata/generar.sh) y cada rojo usa el suyo (H8). Ningún test abre un testdata en su
// sitio: todos trabajan sobre una copia en t.TempDir() (PARA 12). Lo esperado lo calculó el módulo sqlite3 de Python al
// generar, sin código Go (testdata/esperado.json).

type esperadoBase struct {
	FramesConfirmados int `json:"frames_confirmados"`
	SessionMessage    struct {
		Filas      int    `json:"filas"`
		Sha256Data string `json:"sha256_data"`
	} `json:"session_message"`
	SessionV2 struct {
		Filas        int    `json:"filas"`
		TokensInput  int64  `json:"tokens_input"`
		TokensOutput int64  `json:"tokens_output"`
		Sha256Ids    string `json:"sha256_ids"`
	} `json:"session_v2"`
}

type esperadoTabla struct {
	Filas      int    `json:"filas"`
	Sha256Data string `json:"sha256_data"`
}

type esperado struct {
	I, R, S  esperadoBase
	Raros    map[string]esperadoTabla `json:"raros"`
	Desborda map[string]struct {
		Bytes  int    `json:"bytes"`
		Sha256 string `json:"sha256"`
	} `json:"desborda"`
	Profundo struct {
		Niveles       int    `json:"niveles"`
		ClaveInterior int64  `json:"clave_interior"`
		EnHoja        int64  `json:"en_hoja"`
		Ausente       int64  `json:"ausente"`
		MaxRowid      int64  `json:"max_rowid"`
		Filas         int    `json:"filas"`
		DatosEnHoja   [3]any `json:"datos_en_hoja"`
	} `json:"profundo"`
}

func leerEsperado(t *testing.T) esperado {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "esperado.json"))
	if err != nil {
		t.Fatalf("esperado.json: %v", err)
	}
	var e esperado
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatalf("esperado.json: %v", err)
	}
	return e
}

// copia lleva a t.TempDir() los ficheros de testdata indicados (un directorio entero o ficheros sueltos) y devuelve la
// ruta del primero copiado con su nombre base: la base que se abrirá.
func copia(t *testing.T, rel string) string {
	t.Helper()
	destino := t.TempDir()
	origen := filepath.Join("testdata", rel)
	info, err := os.Stat(origen)
	if err != nil {
		t.Fatalf("testdata %s: %v", rel, err)
	}
	if !info.IsDir() {
		copiarFichero(t, origen, filepath.Join(destino, filepath.Base(origen)))
		return filepath.Join(destino, filepath.Base(origen))
	}
	entradas, err := os.ReadDir(origen)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entradas {
		copiarFichero(t, filepath.Join(origen, e.Name()), filepath.Join(destino, e.Name()))
	}
	return filepath.Join(destino, "opencode.db")
}

func copiarFichero(t *testing.T, de, a string) {
	t.Helper()
	b, err := os.ReadFile(de)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(a, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// resumen lee una base de OpenCode con el lector y da lo mismo que esperado.json: recuentos, sumas y sha256.
func resumen(t *testing.T, ruta string) esperadoBase {
	t.Helper()
	b, err := Abrir(ruta)
	if err != nil {
		t.Fatalf("Abrir: %v", err)
	}
	defer b.Cerrar()
	var r esperadoBase
	r.FramesConfirmados = b.FramesWAL
	sm, err := b.Tabla("session_message")
	if err != nil {
		t.Fatalf("Tabla(session_message): %v", err)
	}
	type trozo struct {
		ses  string
		seq  int64
		data []byte
	}
	var trozos []trozo
	cs, cq, cd := sm.Col("session_id"), sm.Col("seq"), sm.Col("data")
	if err := sm.Recorrer(func(f []Valor) error {
		trozos = append(trozos, trozo{string(f[cs].B), f[cq].I, f[cd].B})
		return nil
	}); err != nil {
		t.Fatalf("Recorrer(session_message): %v", err)
	}
	sort.Slice(trozos, func(i, j int) bool {
		if trozos[i].ses != trozos[j].ses {
			return trozos[i].ses < trozos[j].ses
		}
		return trozos[i].seq < trozos[j].seq
	})
	h := sha256.New()
	for _, x := range trozos {
		h.Write(x.data)
	}
	r.SessionMessage.Filas, r.SessionMessage.Sha256Data = len(trozos), hex.EncodeToString(h.Sum(nil))

	sv, err := b.Tabla("session_v2")
	if err != nil {
		t.Fatalf("Tabla(session_v2): %v", err)
	}
	var ids []string
	ci, cin, cout := sv.Col("id"), sv.Col("tokens_input"), sv.Col("tokens_output")
	if err := sv.Recorrer(func(f []Valor) error {
		ids = append(ids, string(f[ci].B))
		r.SessionV2.TokensInput += f[cin].I
		r.SessionV2.TokensOutput += f[cout].I
		return nil
	}); err != nil {
		t.Fatalf("Recorrer(session_v2): %v", err)
	}
	sort.Strings(ids)
	sum := sha256.Sum256([]byte(strings.Join(ids, "")))
	r.SessionV2.Filas, r.SessionV2.Sha256Ids = len(ids), hex.EncodeToString(sum[:])
	return r
}

// tablaT resume la tabla `t` de un raro: filas y sha256 de `data`, en orden de rowid.
func tablaT(t *testing.T, ruta string) (esperadoTabla, error) {
	t.Helper()
	b, err := Abrir(ruta)
	if err != nil {
		return esperadoTabla{}, err
	}
	defer b.Cerrar()
	tb, err := b.Tabla("t")
	if err != nil {
		return esperadoTabla{}, err
	}
	h := sha256.New()
	n := 0
	cd := tb.Col("data")
	err = tb.Recorrer(func(f []Valor) error {
		h.Write(f[cd].B)
		n++
		return nil
	})
	return esperadoTabla{Filas: n, Sha256Data: hex.EncodeToString(h.Sum(nil))}, err
}

// (1) · FR-009: la cabecera. Páginas de 512 y de 65 536 B y bytes reservados se leen igual que sqlite3; UTF-16 se rechaza
// como no soportada.
func TestCabecera(t *testing.T) {
	e := leerEsperado(t)
	for _, n := range []string{"p512", "p65536", "reserva"} {
		t.Run(n, func(t *testing.T) {
			got, err := tablaT(t, copia(t, filepath.Join("raros", n+".db")))
			if err != nil || got != e.Raros[n] {
				t.Errorf("%s: %+v, err %v; sqlite3 dio %+v", n, got, err, e.Raros[n])
			}
		})
	}
	t.Run("utf16", func(t *testing.T) {
		_, err := Abrir(copia(t, filepath.Join("raros", "utf16.db")))
		if !errors.Is(err, ErrNoSoportada) {
			t.Errorf("UTF-16: err %v; se esperaba ErrNoSoportada", err)
		}
	})
}

// (2) · FR-007, SC-004: I, R y S dan exactamente lo que dio sqlite3. En I no aparece la transacción a medias; en R, la
// transacción con el frame roto no existe; S es la base sola.
func TestInstantaneas(t *testing.T) {
	e := leerEsperado(t)
	for _, c := range []struct {
		hoja string
		want esperadoBase
	}{{"I", e.I}, {"R", e.R}, {"S", e.S}} {
		t.Run(c.hoja, func(t *testing.T) {
			if got := resumen(t, copia(t, strings.ToLower(c.hoja))); got != c.want {
				t.Errorf("%s:\n lector %+v\nsqlite3 %+v", c.hoja, got, c.want)
			}
		})
	}
}

// (3) · FR-009, SC-004: filas que desbordan más de una página (> 8 KB) y muchas (> 64 KB), en una base sin WAL.
func TestDesbordamiento(t *testing.T) {
	e := leerEsperado(t)
	for _, c := range []struct{ hoja, id string }{{"8k", "o-desborda-8k"}, {"64k", "o-desborda-64k"}} {
		t.Run(c.hoja, func(t *testing.T) {
			b, err := Abrir(copia(t, filepath.Join("raros", "desborda.db")))
			if err != nil {
				t.Fatal(err)
			}
			defer b.Cerrar()
			tb, err := b.Tabla("t")
			if err != nil {
				t.Fatal(err)
			}
			var data []byte
			err = tb.Recorrer(func(f []Valor) error {
				if string(f[tb.Col("id")].B) == c.id {
					data = append([]byte(nil), f[tb.Col("data")].B...)
				}
				return nil
			})
			sum := sha256.Sum256(data)
			want := e.Desborda[c.id]
			if err != nil || len(data) != want.Bytes || hex.EncodeToString(sum[:]) != want.Sha256 {
				t.Errorf("%s: %d B, err %v; sqlite3 dio %d B", c.id, len(data), err, want.Bytes)
			}
		})
	}
}

func abrirProfundo(t *testing.T) (*Base, *Tabla) {
	t.Helper()
	b, err := Abrir(copia(t, "profundo.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Cerrar() })
	tb, err := b.Tabla("fila")
	if err != nil {
		t.Fatal(err)
	}
	return b, tb
}

// (4) · §P-3 1-2: Buscar baja por las páginas interiores hasta la fila de ese rowid.
func TestBuscar(t *testing.T) {
	e := leerEsperado(t).Profundo
	t.Run("en_hoja", func(t *testing.T) {
		_, tb := abrirProfundo(t)
		f, ok, err := tb.Buscar(e.EnHoja)
		want := fmt.Sprint(e.DatosEnHoja[0], int64(e.DatosEnHoja[1].(float64)), e.DatosEnHoja[2])
		if err != nil || !ok || fmt.Sprint(string(f[0].B), f[1].I, string(f[2].B)) != want {
			t.Errorf("Buscar(%d): ok %v, err %v; se esperaba %s", e.EnHoja, ok, err, want)
		}
	})
	t.Run("clave", func(t *testing.T) {
		_, tb := abrirProfundo(t)
		f, ok, err := tb.Buscar(e.ClaveInterior)
		if err != nil || !ok || f[1].I != e.ClaveInterior || string(f[0].B) != fmt.Sprintf("o-p-%05d", e.ClaveInterior) {
			t.Errorf("Buscar(%d), la clave de una página interior: ok %v, err %v", e.ClaveInterior, ok, err)
		}
	})
	for _, c := range []struct {
		hoja  string
		rowid int64
	}{{"ausente", e.Ausente}, {"mayor", e.MaxRowid + 1}} {
		t.Run(c.hoja, func(t *testing.T) {
			_, tb := abrirProfundo(t)
			if _, ok, err := tb.Buscar(c.rowid); err != nil || ok {
				t.Errorf("Buscar(%d): ok %v, err %v; se esperaba «no está», sin error", c.rowid, ok, err)
			}
		})
	}
}

// (5) · §P-3 3, SC-017: RecorrerDesde da sólo los rowid > h, en orden, y poda el árbol: lee la profundidad más las
// hojas nuevas, no la tabla.
func TestRecorrerDesde(t *testing.T) {
	e := leerEsperado(t).Profundo
	t.Run("solo_mayores", func(t *testing.T) {
		_, tb := abrirProfundo(t)
		var vistos []int64
		err := tb.RecorrerDesde(1500, func(rowid int64, _ []Valor) error {
			vistos = append(vistos, rowid)
			return nil
		})
		ordenados := sort.SliceIsSorted(vistos, func(i, j int) bool { return vistos[i] < vistos[j] })
		if err != nil || len(vistos) != int(e.MaxRowid-1500) || !ordenados || vistos[0] != 1501 {
			t.Errorf("RecorrerDesde(1500): %d filas, err %v, ordenadas %v; se esperaban %d, de 1501 a %d",
				len(vistos), err, ordenados, e.MaxRowid-1500, e.MaxRowid)
		}
	})
	t.Run("paginas", func(t *testing.T) {
		b, tb := abrirProfundo(t)
		antes := b.Paginas()
		n := 0
		err := tb.RecorrerDesde(e.MaxRowid-3, func(int64, []Valor) error { n++; return nil })
		leidas := b.Paginas() - antes
		if err != nil || n != 3 || leidas < 1 || leidas > int64(e.Niveles+1) {
			t.Errorf("RecorrerDesde(max-3): %d filas, %d páginas, err %v; se esperaban 3 filas y ≤ %d páginas",
				n, leidas, err, e.Niveles+1)
		}
	})
}

// (6) · FR-009: las tablas WITHOUT ROWID no se leen; una tabla que no está es un error; las columnas salen del CREATE
// TABLE con sus comillas.
func TestTablas(t *testing.T) {
	t.Run("without_rowid", func(t *testing.T) {
		b, err := Abrir(copia(t, filepath.Join("raros", "without_rowid.db")))
		if err != nil {
			t.Fatal(err)
		}
		defer b.Cerrar()
		tb, err := b.Tabla("t")
		if err == nil {
			err = tb.Recorrer(func([]Valor) error { return nil })
		}
		if !errors.Is(err, ErrNoSoportada) {
			t.Errorf("WITHOUT ROWID: err %v; se esperaba ErrNoSoportada", err)
		}
	})
	t.Run("tabla_ausente", func(t *testing.T) {
		b, err := Abrir(copia(t, "s"))
		if err != nil {
			t.Fatal(err)
		}
		defer b.Cerrar()
		if _, err := b.Tabla("no_existe"); err == nil {
			t.Error("una tabla que no está debe dar error")
		}
	})
	t.Run("comillas", func(t *testing.T) {
		b, err := Abrir(copia(t, "s"))
		if err != nil {
			t.Fatal(err)
		}
		defer b.Cerrar()
		sm, err1 := b.Tabla("session_message")
		sv, err2 := b.Tabla("session_v2")
		if err1 != nil || err2 != nil || sm.Col("id") != 0 || sm.Col("seq") != 3 || sm.Col("data") != 6 ||
			sv.Col("directory") != 7 || sv.Col("tokens_input") != 18 || sv.Col("resume_attempts") != 35 {
			t.Errorf("columnas: %v %v %v / %v", err1, err2, sm.columnas, sv.columnas)
		}
	})
}

// (7) · FR-008: Estable dice si el -wal y la base siguen como al abrir.
func TestEstable(t *testing.T) {
	for _, c := range []struct {
		hoja   string
		cambia bool
	}{{"sin_cambio", false}, {"cambia", true}} {
		t.Run(c.hoja, func(t *testing.T) {
			ruta := copia(t, "i")
			b, err := Abrir(ruta)
			if err != nil {
				t.Fatal(err)
			}
			defer b.Cerrar()
			if c.cambia {
				f, err := os.OpenFile(ruta+"-wal", os.O_APPEND|os.O_WRONLY, 0o600)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = f.Write(make([]byte, 24))
				_ = f.Close()
			}
			if got := b.Estable(); got == c.cambia {
				t.Errorf("Estable() = %v con el -wal %s", got, map[bool]string{true: "cambiado", false: "intacto"}[c.cambia])
			}
		})
	}
}

// (8) · FR-029: los errores no llevan la ruta de la base.
func TestErroresSinRuta(t *testing.T) {
	dir := t.TempDir()
	t.Run("inexistente", func(t *testing.T) {
		ruta := filepath.Join(dir, "CENTINELA-dir", "opencode.db")
		_, err := Abrir(ruta)
		if err == nil || strings.Contains(err.Error(), "CENTINELA-dir") {
			t.Errorf("err %v: no debe ser nil ni llevar la ruta", err)
		}
	})
	t.Run("truncada", func(t *testing.T) {
		ruta := filepath.Join(dir, "CENTINELA-truncada.db")
		if err := os.WriteFile(ruta, []byte("SQLite format 3\x00corta"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := Abrir(ruta)
		if err == nil || strings.Contains(err.Error(), "CENTINELA-truncada") {
			t.Errorf("err %v: no debe ser nil ni llevar la ruta", err)
		}
	})
}

// corrupta escribe en un temporal profundo.db con un cambio y devuelve su ruta.
func corrupta(t *testing.T, cambio func(b []byte, tamPag int, raiz uint32)) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "profundo.db"))
	if err != nil {
		t.Fatal(err)
	}
	_, tb := abrirProfundo(t)
	tamPag := int(b[16])<<8 | int(b[17])
	cambio(b, tamPag, tb.raiz)
	ruta := filepath.Join(t.TempDir(), "corrupta.db")
	if err := os.WriteFile(ruta, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return ruta
}

// pagina devuelve el trozo de b de la página n (desde 1).
func paginaDe(b []byte, tamPag int, n uint32) []byte { return b[int(n-1)*tamPag : int(n)*tamPag] }

// primeraHoja baja por el primer hijo hasta una hoja de tabla.
func primeraHoja(b []byte, tamPag int, n uint32) uint32 {
	for p := paginaDe(b, tamPag, n); p[0] == 0x05; p = paginaDe(b, tamPag, n) {
		c := int(p[12])<<8 | int(p[13])
		n = uint32(p[c])<<24 | uint32(p[c+1])<<16 | uint32(p[c+2])<<8 | uint32(p[c+3])
	}
	return n
}

// cargaEnorme declara, en la celda de menor desplazamiento de la primera hoja, una carga de 2⁶² bytes: un varint de 9
// bytes escrito en su sitio. Con el tope de carga es un error; sin él, `make` revienta (mB8).
func cargaEnorme(b []byte, tamPag int, raiz uint32) {
	p := paginaDe(b, tamPag, primeraHoja(b, tamPag, raiz))
	n := int(p[3])<<8 | int(p[4])
	menor := tamPag
	for i := 0; i < n; i++ {
		if c := int(p[8+2*i])<<8 | int(p[9+2*i]); c < menor {
			menor = c
		}
	}
	copy(p[menor:], []byte{0xA0, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00})
}

// leerTodo abre la base y recorre las tablas que conoce, convirtiendo un pánico en error.
func leerTodo(ruta string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pánico: %v", r)
		}
	}()
	b, err := Abrir(ruta)
	if err != nil {
		return err
	}
	defer b.Cerrar()
	for _, n := range []string{"session_message", "session_v2", "t", "fila"} {
		tb, e := b.Tabla(n)
		if e != nil {
			continue
		}
		if e := tb.Recorrer(func([]Valor) error { return nil }); e != nil {
			return e
		}
	}
	return nil
}

// (9) · FR-010: una estructura imposible es ErrCorrupta, sin pánico ni bucle.
func TestCorruptas(t *testing.T) {
	for _, c := range []struct {
		hoja   string
		cambio func(b []byte, tamPag int, raiz uint32)
	}{
		{"ciclo", func(b []byte, tamPag int, raiz uint32) { // el primer hijo de la raíz apunta a la raíz
			p := paginaDe(b, tamPag, raiz)
			c := int(p[12])<<8 | int(p[13])
			p[c], p[c+1], p[c+2], p[c+3] = byte(raiz>>24), byte(raiz>>16), byte(raiz>>8), byte(raiz)
		}},
		{"puntero_fuera", func(b []byte, tamPag int, raiz uint32) { // el puntero derecho de la raíz, fuera de la base
			p := paginaDe(b, tamPag, raiz)
			p[8], p[9], p[10], p[11] = 0x7F, 0xFF, 0xFF, 0xFF
		}},
		{"carga_enorme", cargaEnorme},
		{"varint_corto", func(b []byte, tamPag int, raiz uint32) { // una celda que empieza en el último byte, con varint abierto
			p := paginaDe(b, tamPag, primeraHoja(b, tamPag, raiz))
			p[8], p[9] = byte((tamPag-1)>>8), byte(tamPag-1)
			p[tamPag-1] = 0xFF
		}},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if err := leerTodo(corrupta(t, c.cambio)); !errors.Is(err, ErrCorrupta) {
				t.Errorf("%s: err %v; se esperaba ErrCorrupta, sin pánico", c.hoja, err)
			}
		})
	}
}

func huella(t *testing.T, dir string) (md5s string, listado string) {
	t.Helper()
	entradas, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var nombres []string
	var sumas bytes.Buffer
	for _, e := range entradas {
		nombres = append(nombres, e.Name())
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&sumas, "%s %x\n", e.Name(), md5.Sum(b))
	}
	return sumas.String(), strings.Join(nombres, ",")
}

// (10) · FR-006: el lector no escribe nada. Tras abrir, leer las dos tablas y comprobar la estabilidad, la carpeta de
// una copia de I tiene los mismos ficheros con el mismo md5.
func TestSoloLectura(t *testing.T) {
	ruta := copia(t, "i")
	dir := filepath.Dir(ruta)
	md5Antes, listaAntes := huella(t, dir)
	_ = leerTodo(ruta) // sólo md5 y listado (H8): un error de lectura no es de este test
	if b, err := Abrir(ruta); err == nil {
		_ = b.Estable()
		_ = b.Cerrar()
	}
	md5Despues, listaDespues := huella(t, dir)
	t.Run("md5", func(t *testing.T) {
		if md5Antes != md5Despues {
			t.Errorf("md5 antes:\n%sdespués:\n%s", md5Antes, md5Despues)
		}
	})
	t.Run("listado", func(t *testing.T) {
		if listaAntes != listaDespues {
			t.Errorf("listado antes %q, después %q", listaAntes, listaDespues)
		}
	})
}
