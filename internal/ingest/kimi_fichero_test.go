package ingest

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/state"
)

// P-012 B1 · Qué ficheros lee el lector de Kimi Code (P-3, FR-003). Los fixtures son SINTÉTICOS, de `testdata/kimi/generar.sh`,
// y NUNCA se abren en su sitio: cada test trabaja sobre una copia en `t.TempDir()`.

// copiarArbolKimi copia `testdata/kimi/<nombre>` entero a un temporal y devuelve su ruta.
func copiarArbolKimi(t *testing.T, nombre string) string {
	t.Helper()
	origen := filepath.Join("testdata", "kimi", nombre)
	destino := filepath.Join(t.TempDir(), nombre)
	err := filepath.WalkDir(origen, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(origen, p)
		if err != nil {
			return err
		}
		a := filepath.Join(destino, rel)
		if d.IsDir() {
			return os.MkdirAll(a, 0o700)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(a, b, 0o600)
	})
	if err != nil {
		t.Fatalf("copiar %s: %v", origen, err)
	}
	return destino
}

// listarKimiRelativo llama a `ListarKimi` sobre una copia de `raiz/` y devuelve las rutas relativas a ella, con `/`.
func listarKimiRelativo(t *testing.T) []string {
	t.Helper()
	raiz := copiarArbolKimi(t, "raiz")
	rutas, err := ListarKimi(raiz)
	if err != nil {
		t.Fatalf("ListarKimi: %v", err)
	}
	var rel []string
	for _, r := range rutas {
		x, err := filepath.Rel(raiz, r)
		if err != nil {
			t.Fatal(err)
		}
		rel = append(rel, filepath.ToSlash(x))
	}
	if len(rel) == 0 {
		t.Fatal("ListarKimi no devolvió nada")
	}
	return rel
}

// (3) · P-3: sólo `<raíz>/sessions/<a>/<b>/agents/<c>/wire.jsonl`, con `a`, `b` y `c` que no empiezan por `.`, y `wire.jsonl` un
// fichero regular. Ni `.index-*`, ni carpetas ocultas, ni `.bak` o `.tmp.*`, ni `<raíz>/imported/`, ni un `wire.jsonl` más hondo.
// Una sesión sin `agents/` no es un error. Orden léxico.
func TestListarKimi(t *testing.T) {
	const wd = "sessions/wd_proyecto-a_000000000001/"
	t.Run("forma", func(t *testing.T) {
		want := []string{
			wd + "session_sintetica-0001/agents/agent-0/wire.jsonl",
			wd + "session_sintetica-0001/agents/main/wire.jsonl",
			wd + "session_sintetica-0002/agents/main/wire.jsonl",
		}
		if got := listarKimiRelativo(t); !slices.Equal(got, want) {
			t.Errorf("ListarKimi = %q; se esperaba %q", got, want)
		}
	})
	t.Run("sin_punto", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			for _, c := range strings.Split(r, "/") {
				if strings.HasPrefix(c, ".") {
					t.Errorf("lista %q, de una carpeta oculta", r)
				}
			}
		}
	})
	t.Run("sin_bak_tmp", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			if path.Base(r) != "wire.jsonl" {
				t.Errorf("lista %q, que no es un wire.jsonl", r)
			}
		}
	})
	t.Run("sin_imported", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			if !strings.HasPrefix(r, "sessions/") {
				t.Errorf("lista %q, fuera de sessions/", r)
			}
		}
	})
	t.Run("hondo", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			if strings.Contains(r, "/sub/") {
				t.Errorf("lista %q, más hondo que agents/<agente>/", r)
			}
		}
	})
	t.Run("vacia", func(t *testing.T) {
		for _, r := range listarKimiRelativo(t) {
			if strings.Contains(r, "session_sintetica-0003") {
				t.Errorf("lista %q, de la sesión sin agents/", r)
			}
		}
	})
	t.Run("orden", func(t *testing.T) {
		if got := listarKimiRelativo(t); !slices.IsSorted(got) {
			t.Errorf("ListarKimi = %q; se esperaba en orden léxico", got)
		}
	})
}

// ═══ P-012 B3 · Un fichero de Kimi Code: atajo, cabeza, prepaso, frontera `forked`, índice, recuentos y sólo lectura ═══
//
// Los fixtures son los de `testdata/kimi/ficheros/` (generar.sh), copiados a un árbol `sessions/<wd>/<S>/agents/<A>/wire.jsonl` en
// `t.TempDir()`. Cada pasada afirma su línea de resumen LITERAL: la de `esperado.json` cuando el fichero se lee de una vez.

const wdKimi = "wd_proyecto-a_000000000001"

// rutaKimi es la ruta del `wire.jsonl` del agente `a` de la sesión `s` bajo `raiz`, en la `wd` dada.
func rutaKimi(raiz, wd, s, a string) string {
	return filepath.Join(raiz, "sessions", wd, s, "agents", a, "wire.jsonl")
}

// ponerFicheroKimi escribe `ficheros/<nombre>` en `ruta`, creando sus carpetas.
func ponerFicheroKimi(t *testing.T, nombre, ruta string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "kimi", "ficheros", filepath.FromSlash(nombre)))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	if err := os.MkdirAll(filepath.Dir(ruta), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// ficheroKimi pone `ficheros/<nombre>` en un árbol nuevo, en la sesión 1 y el agente `main`, y devuelve la raíz y la ruta.
func ficheroKimi(t *testing.T, nombre string) (string, string) {
	t.Helper()
	raiz := t.TempDir()
	ruta := rutaKimi(raiz, wdKimi, sesionKimi1, "main")
	ponerFicheroKimi(t, nombre, ruta)
	return raiz, ruta
}

func baseKimiDePrueba() ContextoKimi {
	return ContextoKimi{Context: Context{Salt: salKimi, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba",
		OrgID: "org-de-prueba", AgentVersion: "test"}}
}

// pasarKimi hace una pasada sobre `ruta` con unos recuentos nuevos y devuelve los eventos, los recuentos y los avisos.
func pasarKimi(t *testing.T, st *state.Store, ruta string, base ContextoKimi) ([]event.Event, *PasadaKimi, string) {
	t.Helper()
	p := NuevaPasadaKimi()
	var avisos bytes.Buffer
	evs, err := LeerFicheroKimi(st, ruta, base, p, &avisos)
	if err != nil {
		t.Fatalf("LeerFicheroKimi: %v", err)
	}
	return evs, p, avisos.String()
}

// resumenEsperadoKimi es la línea de `esperado.json` para un fichero leído de una vez.
func resumenEsperadoKimi(t *testing.T, nombre string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "kimi", "ficheros", "esperado.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	r, ok := m[nombre]
	if !ok {
		t.Fatalf("esperado.json no tiene %s", nombre)
	}
	return r
}

// resumenKimi es la línea de resumen con estos recuentos: el texto de spec §Textos, escrito aquí por separado.
func resumenKimi(reg, ev, her, rep, cero, inc, comp, sinMod int) string {
	return fmt.Sprintf("kimi_code: registros %d · eventos %d · heredados %d · repetidas %d · a cero %d · incoherentes %d · "+
		"compactaciones %d · sin modelo %d", reg, ev, her, rep, cero, inc, comp, sinMod)
}

func comprobarResumenKimi(t *testing.T, p *PasadaKimi, quiero string) {
	t.Helper()
	if got := p.Resumen(); got != quiero {
		t.Errorf("resumen = %q; se esperaba %q", got, quiero)
	}
}

func idsKimi(evs []event.Event) []string {
	var ids []string
	for _, ev := range evs {
		ids = append(ids, ev.EventID)
	}
	return ids
}

// reemplazarKimi cambia el contenido de `ruta` por `ficheros/<nombre>` con un `mtime` posterior, como una reescritura de Kimi.
func reemplazarKimi(t *testing.T, nombre, ruta string) {
	t.Helper()
	ponerFicheroKimi(t, nombre, ruta)
	despues := time.Now().Add(time.Hour)
	if err := os.Chtimes(ruta, despues, despues); err != nil {
		t.Fatal(err)
	}
}

// (14) · FR-005: dos registros → 2 eventos; la 2.ª pasada, sin cambios, 0 eventos y el fichero NO se abre.
func TestLeerFicheroKimi_Pasadas(t *testing.T) {
	_, ruta := ficheroKimi(t, "dos_registros.jsonl")
	st := state.New()
	evs, p, _ := pasarKimi(t, st, ruta, baseKimiDePrueba())
	comprobarResumenKimi(t, p, resumenEsperadoKimi(t, "dos_registros.jsonl"))
	if len(evs) != 2 {
		t.Errorf("1.ª pasada: %d eventos; se esperaban 2", len(evs))
	}
	evs, p, _ = pasarKimi(t, st, ruta, baseKimiDePrueba())
	comprobarResumenKimi(t, p, resumenKimi(0, 0, 0, 0, 0, 0, 0, 0))
	if len(evs) != 0 || p.Aperturas != 0 {
		t.Errorf("2.ª pasada: %d eventos, %d aperturas; se esperaba 0 y 0", len(evs), p.Aperturas)
	}
}

// (15) · D-3: la frontera es la ÚLTIMA `forked`; lo anterior es heredado y no se emite. N cuenta los heredados.
func TestLeerFicheroKimi_Fork(t *testing.T) {
	for _, hoja := range []string{"un_forked", "dos_forked", "sin_forked"} {
		t.Run(hoja, func(t *testing.T) {
			nombre := "fork/" + hoja + ".jsonl"
			_, ruta := ficheroKimi(t, nombre)
			evs, p, _ := pasarKimi(t, state.New(), ruta, baseKimiDePrueba())
			comprobarResumenKimi(t, p, resumenEsperadoKimi(t, nombre))
			if hoja == "un_forked" && (len(evs) != 1 || evs[0].EventID != derivarEventIDKimi(sesionKimi1, "main", 3)) {
				t.Errorf("eventos %q; se esperaba uno, el de N = 3", idsKimi(evs))
			}
		})
	}
}

// (16) · FR-008: un fork leído en dos pasadas, cortado tras la `forked`, da lo mismo que de una vez: el registro nuevo, con N = 3.
func TestLeerFicheroKimi_ForkEnDosPasadas(t *testing.T) {
	_, ruta := ficheroKimi(t, "fork_cortado/paso1.jsonl")
	st := state.New()
	evs1, p, _ := pasarKimi(t, st, ruta, baseKimiDePrueba())
	comprobarResumenKimi(t, p, resumenEsperadoKimi(t, "fork_cortado/paso1.jsonl"))
	reemplazarKimi(t, "fork_cortado/paso2.jsonl", ruta)
	evs2, p, _ := pasarKimi(t, st, ruta, baseKimiDePrueba())
	comprobarResumenKimi(t, p, resumenKimi(1, 1, 0, 0, 0, 0, 0, 0))
	_, rutaUna := ficheroKimi(t, "fork_cortado/paso2.jsonl")
	deUnaVez, _, _ := pasarKimi(t, state.New(), rutaUna, baseKimiDePrueba())
	if got, quiero := idsKimi(append(evs1, evs2...)), idsKimi(deUnaVez); !slices.Equal(got, quiero) ||
		len(got) != 1 || got[0] != derivarEventIDKimi(sesionKimi1, "main", 3) {
		t.Errorf("en dos pasadas %q, de una vez %q; se esperaba sólo el de N = 3 en las dos", got, quiero)
	}
}

// (17) · contrato: N cuenta las líneas completas con el prefijo de registro, también una corrupta; igual en una pasada o en dos.
func TestLeerFicheroKimi_Ordinal(t *testing.T) {
	uno, tres := derivarEventIDKimi(sesionKimi1, "main", 1), derivarEventIDKimi(sesionKimi1, "main", 3)
	t.Run("corrupta_cuenta", func(t *testing.T) {
		_, ruta := ficheroKimi(t, "corrupta_cuenta.jsonl")
		evs, _, _ := pasarKimi(t, state.New(), ruta, baseKimiDePrueba())
		if got := idsKimi(evs); !slices.Equal(got, []string{uno, tres}) {
			t.Errorf("eventos %q; se esperaban los de N = 1 y N = 3", got)
		}
	})
	t.Run("dos_pasadas", func(t *testing.T) {
		_, ruta := ficheroKimi(t, "corrupta_cuenta_paso1.jsonl")
		st := state.New()
		pasarKimi(t, st, ruta, baseKimiDePrueba())
		reemplazarKimi(t, "corrupta_cuenta.jsonl", ruta)
		evs, _, _ := pasarKimi(t, st, ruta, baseKimiDePrueba())
		if got := idsKimi(evs); !slices.Equal(got, []string{tres}) {
			t.Errorf("2.ª pasada: eventos %q; se esperaba el de N = 3", got)
		}
	})
}

// (18) · SC-026, E-1: un fichero cuya cabeza cambia, que encoge, o cuyo estado no tiene `cabeza`, se relee desde 0 y no emite NINGÚN
// `event_id` distinto de los ya emitidos; lo nuevo tras el cambio sale una vez.
func TestLeerFicheroKimi_Cabeza(t *testing.T) {
	for _, c := range []struct {
		hoja, antes, despues string
		sinCabeza            bool
		nuevos               int
	}{
		{"cabeza_cambiada", "cabeza/antes.jsonl", "cabeza/cambiada.jsonl", false, 0},
		{"encogido", "cabeza/largo.jsonl", "cabeza/encogido.jsonl", false, 0},
		{"sin_cabeza", "cabeza/cambiada.jsonl", "cabeza/cambiada_mas.jsonl", true, 1},
		{"nuevo_tras_cambio", "cabeza/antes.jsonl", "cabeza/cambiada_mas.jsonl", false, 1},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			_, ruta := ficheroKimi(t, c.antes)
			st := state.New()
			evs1, _, _ := pasarKimi(t, st, ruta, baseKimiDePrueba())
			if c.sinCabeza {
				e := st.Files[ruta]
				e.Cabeza = "" // un estado de la 0.8.0, sin el campo
				st.Files[ruta] = e
			}
			reemplazarKimi(t, c.despues, ruta)
			evs2, p, _ := pasarKimi(t, st, ruta, baseKimiDePrueba())
			comprobarResumenKimi(t, p, resumenEsperadoKimi(t, c.despues)) // se relee desde 0
			antes := idsKimi(evs1)
			nuevos := 0
			for _, id := range idsKimi(evs2) {
				if !slices.Contains(antes, id) {
					nuevos++
				}
			}
			if nuevos != c.nuevos {
				t.Errorf("event_id nuevos en la 2.ª pasada: %d (antes %q, después %q); se esperaban %d", nuevos, antes, idsKimi(evs2), c.nuevos)
			}
		})
	}
}

// (19) · E-1: el atajo. Mismo tamaño y mismo `mtime`: no se abre. Mismo tamaño y otro `mtime`: se abre, la cabeza es la misma y no
// emite nada.
func TestLeerFicheroKimi_Atajo(t *testing.T) {
	t.Run("igual", func(t *testing.T) {
		_, ruta := ficheroKimi(t, "dos_registros.jsonl")
		st := state.New()
		pasarKimi(t, st, ruta, baseKimiDePrueba())
		if _, p, _ := pasarKimi(t, st, ruta, baseKimiDePrueba()); p.Aperturas != 0 {
			t.Errorf("aperturas = %d; se esperaba 0", p.Aperturas)
		}
	})
	t.Run("mtime", func(t *testing.T) {
		_, ruta := ficheroKimi(t, "dos_registros.jsonl")
		st := state.New()
		pasarKimi(t, st, ruta, baseKimiDePrueba())
		despues := time.Now().Add(time.Hour)
		if err := os.Chtimes(ruta, despues, despues); err != nil {
			t.Fatal(err)
		}
		if evs, p, _ := pasarKimi(t, st, ruta, baseKimiDePrueba()); p.Aperturas != 1 || len(evs) != 0 {
			t.Errorf("aperturas = %d, eventos = %d; se esperaba 1 y 0", p.Aperturas, len(evs))
		}
	})
}

// (20) · SC-004: una línea final sin LF no se consume; con el LF, la pasada siguiente la emite. Una sesión matada no deja registro.
func TestLeerFicheroKimi_SinLFYMatada(t *testing.T) {
	t.Run("sin_lf", func(t *testing.T) {
		_, ruta := ficheroKimi(t, "sin_lf.jsonl")
		st := state.New()
		_, p, _ := pasarKimi(t, st, ruta, baseKimiDePrueba())
		comprobarResumenKimi(t, p, resumenEsperadoKimi(t, "sin_lf.jsonl"))
		reemplazarKimi(t, "sin_lf_completo.jsonl", ruta)
		_, p, _ = pasarKimi(t, st, ruta, baseKimiDePrueba())
		comprobarResumenKimi(t, p, resumenKimi(1, 1, 0, 0, 0, 0, 0, 0))
	})
	t.Run("matada", func(t *testing.T) {
		_, ruta := ficheroKimi(t, "matada.jsonl")
		_, p, _ := pasarKimi(t, state.New(), ruta, baseKimiDePrueba())
		comprobarResumenKimi(t, p, resumenEsperadoKimi(t, "matada.jsonl"))
	})
}

// (21) · P-6: el proyecto sale del `workDir` de la ÚLTIMA línea del índice de su sesión, con las barras de Windows normalizadas.
func TestLeerFicheroKimi_Indice(t *testing.T) {
	proyectoDe := func(t *testing.T, conIndice bool, s string) string {
		t.Helper()
		raiz := t.TempDir()
		if conIndice {
			ponerFicheroKimi(t, "indice/session_index.jsonl", filepath.Join(raiz, "session_index.jsonl"))
		}
		ruta := rutaKimi(raiz, wdKimi, s, "main")
		ponerFicheroKimi(t, "dos_registros.jsonl", ruta)
		idx, err := LeerIndiceKimi(raiz)
		if err != nil {
			t.Fatalf("LeerIndiceKimi: %v", err)
		}
		base := baseKimiDePrueba()
		base.Indice = idx
		evs, _, _ := pasarKimi(t, state.New(), ruta, base)
		if len(evs) == 0 {
			t.Fatal("sin eventos")
		}
		return evs[0].ProjectRef
	}
	t.Run("barras", func(t *testing.T) {
		a, b := proyectoDe(t, true, sesionKimi1), proyectoDe(t, true, sesionKimi2)
		if a == "" || a != b {
			t.Errorf("project_ref con `\\` %q y con `/` %q; se esperaba el mismo, no vacío", a, b)
		}
	})
	t.Run("ultima", func(t *testing.T) {
		if got, quiero := proyectoDe(t, true, "session_sintetica-0003"), project.Derivar("/sintetico/proyecto-b", salKimi); got != quiero {
			t.Errorf("project_ref = %q; se esperaba el de la última línea, %q", got, quiero)
		}
	})
	t.Run("sin_linea", func(t *testing.T) {
		if got := proyectoDe(t, true, "session_sintetica-0004"); got != "" {
			t.Errorf("project_ref = %q; se esperaba vacío", got)
		}
	})
	t.Run("ausente", func(t *testing.T) {
		if got := proyectoDe(t, false, sesionKimi1); got != "" {
			t.Errorf("project_ref = %q; se esperaba vacío", got)
		}
	})
}

// (22) · P-13: un índice que existe y no se puede leer es un error *(aquí, una carpeta con su nombre)*.
func TestLeerIndiceKimi_Ilegible(t *testing.T) {
	raiz := t.TempDir()
	if err := os.Mkdir(filepath.Join(raiz, "session_index.jsonl"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := LeerIndiceKimi(raiz); err == nil {
		t.Error("LeerIndiceKimi no dio error con un índice ilegible")
	}
}

// (23) · La misma sesión copiada en otra `wd`: sus registros salen una vez, y la copia cuenta «repetidas».
func TestLeerFicheroKimi_Repetidas(t *testing.T) {
	raiz := t.TempDir()
	p := NuevaPasadaKimi()
	for _, wd := range []string{wdKimi, "wd_copia_000000000002"} {
		ruta := rutaKimi(raiz, wd, sesionKimi1, "main")
		ponerFicheroKimi(t, "dos_registros.jsonl", ruta)
		if _, err := LeerFicheroKimi(state.New(), ruta, baseKimiDePrueba(), p, nil); err != nil {
			t.Fatal(err)
		}
	}
	comprobarResumenKimi(t, p, resumenKimi(4, 2, 0, 2, 0, 0, 0, 0))
}

// (24) · SC-010: una línea de cada clase, y el mismo fichero en otra `wd`; la identidad de FR-020.
func TestLeerFicheroKimi_SC010Clases(t *testing.T) {
	raiz := t.TempDir()
	p := NuevaPasadaKimi()
	for _, wd := range []string{wdKimi, "wd_copia_000000000002"} {
		ruta := rutaKimi(raiz, wd, sesionKimi1, "main")
		ponerFicheroKimi(t, "clases.jsonl", ruta)
		if _, err := LeerFicheroKimi(state.New(), ruta, baseKimiDePrueba(), p, nil); err != nil {
			t.Fatal(err)
		}
	}
	comprobarResumenKimi(t, p, resumenKimi(14, 3, 2, 3, 2, 4, 1, 1))
	if p.Registros != p.Eventos+p.Heredados+p.Repetidas+p.ACero+p.Incoherentes {
		t.Errorf("no se cumple registros = eventos + heredados + repetidas + a cero + incoherentes: %s", p.Resumen())
	}
}

// (25) · D-2: el subagente sale de SU fichero, con la sesión del principal; `subagent.completed` no se cuenta.
func TestLeerFicheroKimi_Subagente(t *testing.T) {
	raiz := t.TempDir()
	principal, sub := rutaKimi(raiz, wdKimi, sesionKimi1, "main"), rutaKimi(raiz, wdKimi, sesionKimi1, "agent-0")
	ponerFicheroKimi(t, "subagente/main.jsonl", principal)
	ponerFicheroKimi(t, "subagente/agent-0.jsonl", sub)
	evsM, pM, _ := pasarKimi(t, state.New(), principal, baseKimiDePrueba())
	evsS, pS, _ := pasarKimi(t, state.New(), sub, baseKimiDePrueba())
	comprobarResumenKimi(t, pM, resumenEsperadoKimi(t, "subagente/main.jsonl"))
	comprobarResumenKimi(t, pS, resumenEsperadoKimi(t, "subagente/agent-0.jsonl"))
	sesion := event.Ref(salKimi, sesionKimi1)
	for _, ev := range append(evsM, evsS...) {
		if ev.SessionRef != sesion {
			t.Errorf("session_ref = %q; se esperaba el de la sesión, %q", ev.SessionRef, sesion)
		}
	}
	if len(evsS) != 3 {
		t.Errorf("subagente: %d eventos; se esperaban 3", len(evsS))
	}
}

// (26) · Una línea corrupta con el prefijo de registro: aviso, y el fichero sigue.
func TestLeerFicheroKimi_Corrupta(t *testing.T) {
	_, ruta := ficheroKimi(t, "corrupta_cuenta.jsonl")
	evs, p, avisos := pasarKimi(t, state.New(), ruta, baseKimiDePrueba())
	comprobarResumenKimi(t, p, resumenEsperadoKimi(t, "corrupta_cuenta.jsonl"))
	if len(evs) != 2 || strings.Count(avisos, "skip (línea corrupta):") != 1 {
		t.Errorf("%d eventos, avisos %q; se esperaban 2 y un aviso", len(evs), avisos)
	}
}

// huellasArbolKimi es el md5 de cada fichero bajo `raiz`, y el listado.
func huellasArbolKimi(t *testing.T, raiz string) map[string]string {
	t.Helper()
	m := map[string]string{}
	err := filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			m[p] = "dir"
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		m[p] = fmt.Sprintf("%x", md5.Sum(b))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// (27) · FR-006: el lector no escribe, crea ni borra nada bajo la raíz. NACE VERDE con el stub (que no escribe): lo valida m39.
func TestLeerFicheroKimi_SoloLectura(t *testing.T) {
	raiz := copiarArbolKimi(t, "raiz")
	ponerFicheroKimi(t, "dos_registros.jsonl", rutaKimi(raiz, wdKimi, "session_sintetica-0005", "main"))
	antes := huellasArbolKimi(t, raiz)
	rutas, err := ListarKimi(raiz)
	if err != nil {
		t.Fatal(err)
	}
	st := state.New()
	for _, ruta := range rutas {
		if _, err := LeerFicheroKimi(st, ruta, baseKimiDePrueba(), NuevaPasadaKimi(), nil); err != nil {
			t.Fatal(err)
		}
	}
	if despues := huellasArbolKimi(t, raiz); !maps.Equal(antes, despues) {
		t.Errorf("el árbol cambió al leer: %d entradas antes, %d después", len(antes), len(despues))
	}
}

// (28) · contrato, `cabeza`: los dos vectores normativos.
func TestHuellaCabezaKimi(t *testing.T) {
	for _, c := range []struct{ hoja, linea, quiero string }{
		{"v1", `{"type":"metadata","protocol_version":"1.5","created_at":1}`, "f0ad969649b2af3e649c8af85d5fc2ef6757978024f08ca8926b49f21186c7f9"},
		{"v2", `{"type":"metadata","protocol_version":"1.5","created_at":2}`, "572bb1d23c7da57096f084cc31a0b7d4c760530398ca41d5273695e6102ebabe"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if got := huellaCabezaKimi("sal-sintetica", []byte(c.linea)); got != c.quiero {
				t.Errorf("huellaCabezaKimi = %q; se esperaba %s", got, c.quiero)
			}
		})
	}
}

// (29) · D-012-P3: `cabeza` lleva `omitempty`. Una entrada sin ella se serializa como en la 0.8.0, y un `state.json` de la 0.8.0 se
// carga y se guarda sin perder ni ganar nada.
func TestEstadoKimi_Compatible(t *testing.T) {
	t.Run("entrada_0_8_0", func(t *testing.T) {
		b, err := json.Marshal(state.FileState{Path: "/sintetico/a.jsonl", Size: 1, ModTime: 2, Offset: 1})
		if err != nil || string(b) != `{"path":"/sintetico/a.jsonl","size":1,"mod_time":2,"offset":1}` {
			t.Errorf("entrada = %s (%v); se esperaba la de la 0.8.0, sin `cabeza`", b, err)
		}
	})
	t.Run("ida_y_vuelta", func(t *testing.T) {
		dir := t.TempDir()
		viejo := `{"files":{"/sintetico/a.jsonl":{"path":"/sintetico/a.jsonl","size":1,"mod_time":2,"offset":1},` +
			`"/sintetico/o.db":{"path":"/sintetico/o.db","size":3,"mod_time":4,"offset":5,"frontera":{"huella":"h","en_curso":[{"rowid":6,"huella":"i"}]}}}}`
		if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(viejo), 0o600); err != nil {
			t.Fatal(err)
		}
		st, err := state.Load(filepath.Join(dir, "state.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := st.Save(filepath.Join(dir, "nuevo.json")); err != nil {
			t.Fatal(err)
		}
		nuevo, err := os.ReadFile(filepath.Join(dir, "nuevo.json"))
		if err != nil {
			t.Fatal(err)
		}
		var a, b any
		if json.Unmarshal([]byte(viejo), &a) != nil || json.Unmarshal(nuevo, &b) != nil || fmt.Sprint(a) != fmt.Sprint(b) {
			t.Errorf("tras cargar y guardar:\n%s\nse esperaba lo mismo que\n%s", nuevo, viejo)
		}
	})
	t.Run("con_cabeza", func(t *testing.T) {
		if b, _ := json.Marshal(state.FileState{Path: "/sintetico/w.jsonl", Cabeza: "c"}); !strings.Contains(string(b), `"cabeza":"c"`) {
			t.Errorf("entrada = %s; se esperaba `cabeza`", b)
		}
	})
}
