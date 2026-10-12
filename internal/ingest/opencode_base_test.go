package ingest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/sqlitero"
	"github.com/permea-dev/agent/internal/state"
)

// P-010 B4 · La pasada de una base de OpenCode y su frontera por rowid. Bases SINTÉTICAS de testdata/opencode/generar.sh:
// ningún test abre una en su sitio, todos trabajan sobre una copia en t.TempDir() (PARA 12). Las secuencias entre pasadas
// son instantáneas (`paso1/`, `paso2/`…) copiadas sobre la misma ruta. `Ahora` va inyectado en todas las pasadas (H6), y
// cada pasada afirma su línea de resumen literal.

const salBaseOpenCode = "sal-de-prueba-opencode-base"

// formatoResumenOpenCode es la línea de spec §Textos, literal.
const formatoResumenOpenCode = "opencode: llamadas %d · eventos %d · en curso %d · abandonadas %d · sin identificador %d · " +
	"incoherentes %d · sin modelo %d · fronteras reiniciadas %d · bases en formato anterior %d · bases con esquema desconocido %d"

// resumenOpenCode da la línea esperada, con los diez recuentos en el orden del texto.
func resumenOpenCode(n ...any) string { return fmt.Sprintf(formatoResumenOpenCode, n...) }

// esperadoOpenCode es lo que `generar.sh` sacó con sqlite3 al generar.
type esperadoOpenCode struct {
	T0       int64
	Frontera struct {
		Completadas []string
		EnCurso     string `json:"en_curso"`
		Niveles     int
	}
	Inverso struct{ Orden []string }
	Vacuum  struct {
		HAntes             int64    `json:"h_antes"`
		IDHAntes           string   `json:"id_h_antes"`
		RowidDeIDHDespues  int64    `json:"rowid_de_id_h_despues"`
		IDEnHDespues       string   `json:"id_en_h_despues"`
		CompletadasDespues []string `json:"completadas_despues"`
	}
	BorradaH    struct{ Nuevas []string } `json:"borrada_h"`
	Abandono    struct{ Tarde string }
	DosSesiones map[string][2]string `json:"dos_sesiones"`
	Centinelas  struct {
		H       string
		EnCurso []string `json:"en_curso"`
	}
}

func esperadoOpenCodeDe(t *testing.T) esperadoOpenCode {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "opencode", "esperado.json"))
	if err != nil {
		t.Fatalf("esperado.json: %v", err)
	}
	var e esperadoOpenCode
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatalf("esperado.json: %v", err)
	}
	return e
}

// ahoraOpenCode es `t0` más d: el reloj de una pasada.
func (e esperadoOpenCode) ahora(d time.Duration) time.Time { return time.UnixMilli(e.T0).Add(d) }

// ponerOpenCode deja en ruta la instantánea rel de testdata/opencode (un directorio con `opencode.db` y quizá su -wal, o
// una base suelta), en lugar de lo que hubiera.
func ponerOpenCode(t *testing.T, ruta, rel string) {
	t.Helper()
	for _, suf := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(ruta + suf); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	origen := filepath.Join("testdata", "opencode", rel)
	info, err := os.Stat(origen)
	if err != nil {
		t.Fatalf("testdata %s: %v", rel, err)
	}
	if !info.IsDir() {
		copiarOpenCode(t, origen, ruta)
		return
	}
	copiarOpenCode(t, filepath.Join(origen, "opencode.db"), ruta)
	if _, err := os.Stat(filepath.Join(origen, "opencode.db-wal")); err == nil {
		copiarOpenCode(t, filepath.Join(origen, "opencode.db-wal"), ruta+"-wal")
	}
}

func copiarOpenCode(t *testing.T, de, a string) {
	t.Helper()
	b, err := os.ReadFile(de)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(a, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func contextoBaseOpenCode(ahora time.Time) ContextoOpenCode {
	return ContextoOpenCode{
		Context: Context{Salt: salBaseOpenCode, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba", OrgID: "org-de-prueba",
			AgentVersion: "test"},
		Ahora: func() time.Time { return ahora },
	}
}

// pasarOpenCode hace una pasada que debe acabar bien.
func pasarOpenCode(t *testing.T, st *state.Store, ruta string, ahora time.Time) ([]event.Event, *PasadaOpenCode) {
	t.Helper()
	p := NuevaPasadaOpenCode()
	evs, err := LeerBaseOpenCode(st, ruta, contextoBaseOpenCode(ahora), p)
	if err != nil {
		t.Fatalf("LeerBaseOpenCode: %v", err)
	}
	return evs, p
}

// idsOpenCode son los `event_id` de unos eventos, en su orden.
func idsOpenCode(evs []event.Event) []string {
	ids := make([]string, 0, len(evs))
	for _, ev := range evs {
		ids = append(ids, ev.EventID)
	}
	return ids
}

// eventIDsOpenCode son los `event_id` del contrato para unos `id` de fila.
func eventIDsOpenCode(ids ...string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		e, _ := derivarEventIDOpenCode(id)
		out = append(out, e)
	}
	return out
}

func comprobarResumen(t *testing.T, p *PasadaOpenCode, want string) {
	t.Helper()
	if got := p.Resumen(); got != want {
		t.Errorf("resumen:\n  %q\nse esperaba\n  %q", got, want)
	}
}

// (23) · SC-011: la primera pasada de 4 completadas y 1 en curso.
func TestBaseOpenCode_PrimeraPasada(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	ponerOpenCode(t, ruta, "frontera/paso1")
	evs, p := pasarOpenCode(t, state.New(), ruta, esp.ahora(time.Hour))
	comprobarResumen(t, p, resumenOpenCode(4, 4, 1, 0, 0, 0, 0, 0, 0, 0))
	if got, want := idsOpenCode(evs), eventIDsOpenCode(esp.Frontera.Completadas...); !slices.Equal(got, want) {
		t.Errorf("event_id %v; se esperaban los de las 4 completadas, %v", got, want)
	}
}

// paginasDelEsquema son las páginas que cuesta buscar las dos tablas en `sqlite_schema`, como hace la pasada.
func paginasDelEsquema(t *testing.T, ruta string) int64 {
	t.Helper()
	b, err := sqlitero.Abrir(ruta)
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	defer func() { _ = b.Cerrar() }()
	for _, n := range []string{"session_message", "session_v2"} {
		if _, err := b.Tabla(n); err != nil {
			t.Fatalf("precondición: %v", err)
		}
	}
	return b.Paginas()
}

// (24) · SC-011: la segunda pasada, sin cambios, no emite, y sólo lee `sqlite_schema` y tres descensos: a H, a la fila en
// curso y el de `RecorrerDesde(H)`.
func TestBaseOpenCode_SegundaPasadaSinCambios(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	st := state.New()
	ponerOpenCode(t, ruta, "frontera/paso1")
	pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
	ponerOpenCode(t, ruta, "frontera/paso2")
	evs, p := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour+time.Minute))
	t.Run("linea", func(t *testing.T) {
		comprobarResumen(t, p, resumenOpenCode(0, 0, 1, 0, 0, 0, 0, 0, 0, 0))
		if len(evs) != 0 {
			t.Errorf("%d eventos; se esperaba 0", len(evs))
		}
	})
	t.Run("paginas", func(t *testing.T) {
		cota := paginasDelEsquema(t, ruta) + 3*int64(esp.Frontera.Niveles)
		if p.paginas == 0 || p.paginas > cota {
			t.Errorf("%d páginas leídas; se esperaban entre 1 y %d (el esquema y 3 descensos de %d niveles)", p.paginas, cota,
				esp.Frontera.Niveles)
		}
	})
}

// (25) · SC-010: la fila en curso, completada, sale en la pasada siguiente.
func TestBaseOpenCode_EnCursoSeCompleta(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	st := state.New()
	ponerOpenCode(t, ruta, "frontera/paso1")
	pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
	ponerOpenCode(t, ruta, "frontera/paso3")
	evs, p := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour+time.Minute))
	comprobarResumen(t, p, resumenOpenCode(1, 1, 0, 0, 0, 0, 0, 0, 0, 0))
	if got, want := idsOpenCode(evs), eventIDsOpenCode(esp.Frontera.EnCurso); !slices.Equal(got, want) {
		t.Errorf("event_id %v; se esperaba el de la fila en curso, %v", got, want)
	}
}

// (26) · SC-011: dos filas que se completan al revés de su rowid salen una vez cada una.
func TestBaseOpenCode_OrdenInverso(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	st := state.New()
	for i, c := range []struct {
		resumen string
		ids     []string
	}{
		{resumenOpenCode(0, 0, 2, 0, 0, 0, 0, 0, 0, 0), nil},
		{resumenOpenCode(1, 1, 1, 0, 0, 0, 0, 0, 0, 0), eventIDsOpenCode(esp.Inverso.Orden[0])},
		{resumenOpenCode(1, 1, 0, 0, 0, 0, 0, 0, 0, 0), eventIDsOpenCode(esp.Inverso.Orden[1])},
	} {
		ponerOpenCode(t, ruta, fmt.Sprintf("inverso/paso%d", i+1))
		evs, p := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour+time.Duration(i)*time.Minute))
		comprobarResumen(t, p, c.resumen)
		if got := idsOpenCode(evs); !slices.Equal(got, c.ids) {
			t.Errorf("paso %d: event_id %v; se esperaba %v", i+1, got, c.ids)
		}
	}
}

// (27) · SC-011: una fila de la lista que desaparece sale de ella sin contarse. La fila H, completada, sigue.
func TestBaseOpenCode_FilaDeLaListaBorrada(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	st := state.New()
	ponerOpenCode(t, ruta, "lista_borrada/paso1")
	_, p := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
	comprobarResumen(t, p, resumenOpenCode(1, 1, 1, 0, 0, 0, 0, 0, 0, 0))
	ponerOpenCode(t, ruta, "lista_borrada/paso2")
	_, p = pasarOpenCode(t, st, ruta, esp.ahora(time.Hour+time.Minute))
	comprobarResumen(t, p, resumenOpenCode(0, 0, 0, 0, 0, 0, 0, 0, 0, 0))
}

// (28) · SC-011: una entrada presente, sin el campo nuevo, da una pasada completa, sin «frontera reiniciada».
func TestBaseOpenCode_EntradaSinCampo(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	ponerOpenCode(t, ruta, "frontera/paso1")
	st := state.New()
	st.Files[ruta] = state.FileState{Path: ruta, Size: 1, ModTime: 1, Offset: 23}
	evs, p := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
	comprobarResumen(t, p, resumenOpenCode(4, 4, 1, 0, 0, 0, 0, 0, 0, 0))
	if got, want := idsOpenCode(evs), eventIDsOpenCode(esp.Frontera.Completadas...); !slices.Equal(got, want) {
		t.Errorf("event_id %v; se esperaban los de las 4 completadas, %v", got, want)
	}
}

// (29) · SC-024: tras borrar 2 filas anteriores a H, `VACUUM` y 3 nuevas, la pasada reinicia la frontera y no falta
// ninguna completada. Antes, el test comprueba que los rowid cambiaron.
func TestBaseOpenCode_VacuumRenumera(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	v := esp.Vacuum
	if v.RowidDeIDHDespues == v.HAntes || v.IDEnHDespues == "" || v.IDEnHDespues == v.IDHAntes {
		t.Fatalf("precondición: el VACUUM no renumeró (%+v)", v)
	}
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	st := state.New()
	ponerOpenCode(t, ruta, "vacuum/antes")
	pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
	ponerOpenCode(t, ruta, "vacuum/despues")
	evs, p := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour+time.Minute))
	t.Run("ninguna_falta", func(t *testing.T) {
		got := idsOpenCode(evs)
		for i, id := range eventIDsOpenCode(v.CompletadasDespues...) {
			if !slices.Contains(got, id) {
				t.Errorf("falta la completada %s", v.CompletadasDespues[i])
			}
		}
	})
	t.Run("literal", func(t *testing.T) {
		comprobarResumen(t, p, resumenOpenCode(9, 9, 0, 0, 0, 0, 0, 1, 0, 0))
	})
}

// (30) · H10: se borran las filas de rowid máximo y entran 2 que reutilizan rowids ≤ H: H falta, y se reinicia.
func TestBaseOpenCode_BorradaH(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	st := state.New()
	ponerOpenCode(t, ruta, "borrada_h/paso1")
	pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
	ponerOpenCode(t, ruta, "borrada_h/paso2")
	evs, p := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour+time.Minute))
	t.Run("salen", func(t *testing.T) {
		got := idsOpenCode(evs)
		for i, id := range eventIDsOpenCode(esp.BorradaH.Nuevas...) {
			if !slices.Contains(got, id) {
				t.Errorf("no salió la nueva %s", esp.BorradaH.Nuevas[i])
			}
		}
	})
	t.Run("reinicio", func(t *testing.T) {
		comprobarResumen(t, p, resumenOpenCode(4, 4, 0, 0, 0, 0, 0, 1, 0, 0))
	})
}

// (31) · SC-025: con `Ahora` +1 h, +25 h y +26 h, una en curso de más de 24 h se abandona y no vuelve a contarse; la que se
// completa con más de 24 h sale como evento. Y, en una primera pasada, la de hace 25 h se abandona y la de hace 1 h sigue.
func TestBaseOpenCode_Abandono(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	st := state.New()
	t.Run("paso1", func(t *testing.T) {
		ponerOpenCode(t, ruta, "abandono/paso1")
		_, p := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
		comprobarResumen(t, p, resumenOpenCode(1, 1, 2, 0, 0, 0, 0, 0, 0, 0))
	})
	t.Run("paso2", func(t *testing.T) {
		ponerOpenCode(t, ruta, "abandono/paso2")
		evs, p := pasarOpenCode(t, st, ruta, esp.ahora(25*time.Hour))
		comprobarResumen(t, p, resumenOpenCode(1, 1, 0, 1, 0, 0, 0, 0, 0, 0))
		if got, want := idsOpenCode(evs), eventIDsOpenCode(esp.Abandono.Tarde); !slices.Equal(got, want) {
			t.Errorf("event_id %v; se esperaba el de la que se completó tarde, %v", got, want)
		}
	})
	t.Run("paso3", func(t *testing.T) {
		ponerOpenCode(t, ruta, "abandono/paso3")
		_, p := pasarOpenCode(t, st, ruta, esp.ahora(26*time.Hour))
		comprobarResumen(t, p, resumenOpenCode(0, 0, 0, 0, 0, 0, 0, 0, 0, 0))
	})
	t.Run("1h", func(t *testing.T) {
		ruta := filepath.Join(t.TempDir(), "opencode.db")
		st := state.New()
		ponerOpenCode(t, ruta, "abandono_1h.db")
		_, p := pasarOpenCode(t, st, ruta, esp.ahora(25*time.Hour))
		comprobarResumen(t, p, resumenOpenCode(1, 1, 1, 1, 0, 0, 0, 0, 0, 0))
		_, p = pasarOpenCode(t, st, ruta, esp.ahora(25*time.Hour+time.Minute))
		comprobarResumen(t, p, resumenOpenCode(0, 0, 1, 0, 0, 0, 0, 0, 0, 0))
	})
}

// (32) · SC-013: una fila de cada clase, con `Ahora` = t0 + 25 h; y la identidad de FR-021.
func TestBaseOpenCode_Clases(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	ponerOpenCode(t, ruta, "clases.db")
	evs, p := pasarOpenCode(t, state.New(), ruta, esp.ahora(25*time.Hour))
	t.Run("literal", func(t *testing.T) {
		comprobarResumen(t, p, resumenOpenCode(5, 2, 1, 1, 1, 2, 1, 0, 0, 0))
	})
	t.Run("identidad", func(t *testing.T) {
		if p.Llamadas == 0 || p.Llamadas != p.Eventos+p.SinIdentificador+p.Incoherentes || len(evs) != p.Eventos {
			t.Errorf("llamadas %d, eventos %d (%d emitidos), sin identificador %d, incoherentes %d; no cumple FR-021",
				p.Llamadas, p.Eventos, len(evs), p.SinIdentificador, p.Incoherentes)
		}
	})
}

// ganchoOpenCode pone el gancho de entre la lectura y `Estable()` durante un test.
func ganchoOpenCode(t *testing.T, f func()) {
	t.Helper()
	antesDeEstableOpenCode = f
	t.Cleanup(func() { antesDeEstableOpenCode = nil })
}

// tocarWAL añade un byte al -wal: `Estable()` ve otro tamaño, y la lectura siguiente lo ignora (no es un frame entero).
func tocarWAL(t *testing.T, ruta string) {
	t.Helper()
	f, err := os.OpenFile(ruta+"-wal", os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Errorf("tocar el -wal: %v", err)
		return
	}
	_, _ = f.Write([]byte{0})
	_ = f.Close()
}

// (33) · SC-029: una base que cambia entre la lectura y `Estable()` se relee y sus llamadas salen una vez; si cambia en los
// 5 intentos, la pasada da error, la frontera no se mueve y la pasada siguiente la lee. El error lleva el nombre de la base,
// no su ruta (D-010-P7).
func TestBaseOpenCode_Inestable(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	t.Run("una", func(t *testing.T) {
		ruta := filepath.Join(t.TempDir(), "opencode.db")
		ponerOpenCode(t, ruta, "frontera/paso1")
		veces := 0
		ganchoOpenCode(t, func() {
			if veces++; veces == 1 {
				tocarWAL(t, ruta)
			}
		})
		evs, p := pasarOpenCode(t, state.New(), ruta, esp.ahora(time.Hour))
		comprobarResumen(t, p, resumenOpenCode(4, 4, 1, 0, 0, 0, 0, 0, 0, 0))
		if len(evs) != 4 || veces != 2 {
			t.Errorf("%d eventos en %d lecturas; se esperaban 4 en 2", len(evs), veces)
		}
	})
	t.Run("quieta", func(t *testing.T) {
		ruta := filepath.Join(t.TempDir(), "opencode.db")
		st := state.New()
		ponerOpenCode(t, ruta, "frontera/paso1")
		pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
		antes, _ := json.Marshal(st.Files[ruta])
		ponerOpenCode(t, ruta, "frontera/paso3")
		ganchoOpenCode(t, func() { tocarWAL(t, ruta) })
		if _, err := LeerBaseOpenCode(st, ruta, contextoBaseOpenCode(esp.ahora(time.Hour)), NuevaPasadaOpenCode()); err == nil {
			t.Error("cinco lecturas inestables, sin error")
		}
		if despues, _ := json.Marshal(st.Files[ruta]); string(despues) != string(antes) {
			t.Errorf("la frontera se movió con el error:\n  %s\nantes\n  %s", despues, antes)
		}
		antesDeEstableOpenCode = nil
		evs, _ := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour+time.Minute))
		if id := eventIDsOpenCode(esp.Frontera.EnCurso)[0]; !slices.Contains(idsOpenCode(evs), id) {
			t.Errorf("la pasada siguiente no sacó la fila completada (%s)", esp.Frontera.EnCurso)
		}
	})
	t.Run("sin_ruta", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "CENTINELA-carpeta")
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		ruta := filepath.Join(dir, "opencode.db")
		ponerOpenCode(t, ruta, "frontera/paso1")
		ganchoOpenCode(t, func() { tocarWAL(t, ruta) })
		_, err := LeerBaseOpenCode(state.New(), ruta, contextoBaseOpenCode(esp.ahora(time.Hour)), NuevaPasadaOpenCode())
		if err == nil || strings.Contains(err.Error(), "CENTINELA") || !strings.Contains(err.Error(), "opencode.db") {
			t.Errorf("error %v; se esperaba uno con el nombre de la base y sin su carpeta", err)
		}
	})
}

// (34) · SC-015, FR-022, FR-023: la 1.x cuenta 1, 0 en la pasada siguiente, y 1 si cambia; `WITHOUT ROWID`, UTF-16 y una
// columna que falta cuentan «esquema desconocido», sin eventos. Una base corrupta es un error, y no se cuenta.
func TestBaseOpenCode_Formatos(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ahora := esp.ahora(time.Hour)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	ponerOpenCode(t, ruta, "formatos/v1.db")
	st := state.New()
	t.Run("anterior_primera", func(t *testing.T) {
		_, p := pasarOpenCode(t, st, ruta, ahora)
		comprobarResumen(t, p, resumenOpenCode(0, 0, 0, 0, 0, 0, 0, 0, 1, 0))
	})
	t.Run("anterior_segunda", func(t *testing.T) {
		_, p := pasarOpenCode(t, st, ruta, ahora)
		comprobarResumen(t, p, resumenOpenCode(0, 0, 0, 0, 0, 0, 0, 0, 0, 0))
	})
	t.Run("anterior_cambia", func(t *testing.T) {
		if err := os.Chtimes(ruta, ahora, ahora.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		_, p := pasarOpenCode(t, st, ruta, ahora)
		comprobarResumen(t, p, resumenOpenCode(0, 0, 0, 0, 0, 0, 0, 0, 1, 0))
	})
	for _, nombre := range []string{"without_rowid", "utf16", "sin_columna"} {
		t.Run(nombre, func(t *testing.T) {
			ruta := filepath.Join(t.TempDir(), "opencode.db")
			ponerOpenCode(t, ruta, "formatos/"+nombre+".db")
			evs, p := pasarOpenCode(t, state.New(), ruta, ahora)
			comprobarResumen(t, p, resumenOpenCode(0, 0, 0, 0, 0, 0, 0, 0, 0, 1))
			if len(evs) != 0 {
				t.Errorf("%d eventos; se esperaba 0", len(evs))
			}
		})
	}
	t.Run("corrupta", func(t *testing.T) {
		ruta := filepath.Join(t.TempDir(), "opencode.db")
		ponerOpenCode(t, ruta, "corrupta.db")
		st := state.New()
		p := NuevaPasadaOpenCode()
		_, err := LeerBaseOpenCode(st, ruta, contextoBaseOpenCode(ahora), p)
		if _, hay := st.Files[ruta]; err == nil || hay || p.HayNovedades() {
			t.Errorf("error %v, entrada %v, novedades %v; se esperaba un error, sin entrada y sin recuentos", err, hay, p.HayNovedades())
		}
	})
}

// (35) · P-8, FR-003: `opencode.db` y `opencode-<canal>.db`, menos `opencode-next.db`, en orden léxico; nada más.
func TestListarOpenCode(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"opencode.db", "opencode-beta.db", "opencode-dev_2.db", "opencode-next.db", "opencode.db-wal",
		"opencode.db-shm", "opencode-x y.db", "opencode-.db", "otro.db", "opencode.sqlite"} {
		if err := os.WriteFile(filepath.Join(dir, n), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{"storage", "log", "opencode-carpeta.db"} {
		if err := os.Mkdir(filepath.Join(dir, d), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	got, err := ListarOpenCode(dir)
	want := []string{filepath.Join(dir, "opencode-beta.db"), filepath.Join(dir, "opencode-dev_2.db"), filepath.Join(dir, "opencode.db")}
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("ListarOpenCode = %v, %v; se esperaba %v", got, err, want)
	}
}

// (36) · FR-004, D-010-P2: `frontera` sólo en las entradas de las bases; un `state.json` de la 0.6.0 se lee, y su entrada sin
// el campo da una pasada completa.
func TestBaseOpenCode_EstadoJSON(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	t.Run("bases", func(t *testing.T) {
		ruta := filepath.Join(t.TempDir(), "opencode.db")
		st := state.New()
		ponerOpenCode(t, ruta, "frontera/paso1")
		pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
		if b, _ := json.Marshal(st.Files[ruta]); !strings.Contains(string(b), `"frontera":{`) {
			t.Errorf("la entrada de la base no lleva `frontera`: %s", b)
		}
	})
	t.Run("otras", func(t *testing.T) {
		b, _ := json.Marshal(state.FileState{Path: "/sintetico/sesion.jsonl", Size: 10, ModTime: 20, Offset: 10})
		if strings.Contains(string(b), "frontera") {
			t.Errorf("una entrada de otro lector lleva `frontera`: %s", b)
		}
	})
	t.Run("estado_060", func(t *testing.T) {
		dir := t.TempDir()
		ruta := filepath.Join(dir, "opencode.db")
		ponerOpenCode(t, ruta, "frontera/paso1")
		viejo := fmt.Sprintf(`{"files":{%q:{"path":%q,"size":1,"mod_time":1,"offset":23}}}`, ruta, ruta)
		if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(viejo), 0o600); err != nil {
			t.Fatal(err)
		}
		st, err := state.Load(filepath.Join(dir, "state.json"))
		if err != nil {
			t.Fatalf("state.Load de un state.json de la 0.6.0: %v", err)
		}
		_, p := pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
		comprobarResumen(t, p, resumenOpenCode(4, 4, 1, 0, 0, 0, 0, 0, 0, 0))
	})
}

// (37) · P-6: dos sesiones, dos `directory`; y una fila cuya sesión no está en `session_v2`, con las referencias vacías.
func TestBaseOpenCode_DosSesiones(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	ponerOpenCode(t, ruta, "dos_sesiones.db")
	evs, _ := pasarOpenCode(t, state.New(), ruta, esp.ahora(time.Hour))
	porID := map[string]event.Event{}
	for _, ev := range evs {
		porID[ev.EventID] = ev
	}
	for _, id := range []string{"o-dos-1", "o-dos-2", "o-dos-3"} {
		t.Run(id, func(t *testing.T) {
			ev, ok := porID[eventIDsOpenCode(id)[0]]
			if !ok {
				t.Fatalf("no salió %s", id)
			}
			s := esp.DosSesiones[id]
			want := [2]string{event.Ref(salBaseOpenCode, s[0]), ""}
			if s[0] != "" {
				want[1] = project.Derivar(s[1], salBaseOpenCode)
			}
			if got := [2]string{ev.SessionRef, ev.ProjectRef}; got != want {
				t.Errorf("(session_ref, project_ref) = %q; se esperaba %q", got, want)
			}
		})
	}
}

// (38) · D-010-P3, SC-014: las huellas de `state.json` llevan sal, y ningún identificador del proveedor queda fuera de la
// ruta de la base (que también lleva un centinela).
func TestBaseOpenCode_HuellaConSal(t *testing.T) {
	esp := esperadoOpenCodeDe(t)
	ruta := filepath.Join(t.TempDir(), "opencode-CENTINELA-base.db")
	ponerOpenCode(t, ruta, "centinelas.db")
	st := state.New()
	pasarOpenCode(t, st, ruta, esp.ahora(time.Hour))
	fr := st.Files[ruta].Frontera
	if fr == nil || fr.Huella != event.Ref(salBaseOpenCode, esp.Centinelas.H) || len(fr.EnCurso) != 1 ||
		fr.EnCurso[0].Huella != event.Ref(salBaseOpenCode, esp.Centinelas.EnCurso[0]) {
		t.Errorf("frontera %+v; se esperaban las huellas con sal de la fila H y de la que sigue en curso", fr)
	}
	b, _ := json.Marshal(st)
	if fuera := strings.ReplaceAll(string(b), ruta, ""); strings.Contains(fuera, "CENTINELA") {
		t.Errorf("state.json lleva un centinela fuera de la ruta de la base:\n%s", b)
	}
}
