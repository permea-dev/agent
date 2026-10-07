package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/state"
	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ P-007 B4 · UN MENSAJE SE ENVÍA CUANDO ESTÁ CERRADO (FR-010 a FR-015, FR-021; SC-006 a SC-009) ════
//
// Lo abierto no se consume: el offset de su fichero se queda en el comienzo del mensaje y la pasada
// siguiente lo relee. En proceso, el reloj se fija por CÓDIGO (`agent.reloj`) y el `mtime` con
// `os.Chtimes`. Los tests de proceso no pueden fijar el reloj —no hay puerta de test en producción—, así
// que sus líneas llevan el `timestamp` de AHORA. Sandbox, sin enrolar y con raíz de logs temporal
// (disciplina 6). Identificadores sintéticos (disciplina 9).

const esperaT = 10 * time.Minute

// lineaRetencion es una línea del mensaje `sufijo` con salida `salida` y `timestamp` `momento`, con
// identificadores sintéticos. `lineaConIDs` permite sembrar centinelas.
func lineaRetencion(sufijo string, salida int, momento time.Time) string {
	return lineaConIDs(fmt.Sprintf("msg_RETEN%021s", sufijo), fmt.Sprintf("req_RETEN%021s", sufijo), salida, momento)
}

func lineaConIDs(mensajeID, peticionID string, salida int, momento time.Time) string {
	return fmt.Sprintf(`{"type":"assistant","timestamp":%q,"sessionId":"s","cwd":"/tmp/x","requestId":%q,`+
		`"message":{"id":%q,"model":"claude-opus-5-5","usage":{"input_tokens":1,"output_tokens":%d,`+
		`"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`+"\n",
		momento.Format(time.RFC3339Nano), peticionID, mensajeID, salida)
}

// escribirLog escribe `contenido` en `ruta` y le pone el `mtime` dado.
func escribirLog(t *testing.T, ruta, contenido string, mtime time.Time) {
	t.Helper()
	if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
		t.Fatalf("escribir el log de prueba: %v", err)
	}
	if err := os.Chtimes(ruta, mtime, mtime); err != nil {
		t.Fatalf("fijar el mtime: %v", err)
	}
}

// agenteConReloj es un agente en sandbox sobre `logs`, con el reloj fijado en `ahora`.
func agenteConReloj(t *testing.T, dataDir, logs string, ahora time.Time) *agent {
	t.Helper()
	cfg := config.Config{LogsRoot: logs}
	return &agent{dir: dataDir, cfg: cfg, ictx: newIngestContext("test", cfg, "sal-de-prueba", "maquina-de-prueba"),
		reloj: func() time.Time { return ahora }}
}

// salidasEnCola devuelve la salida de cada evento de la cola, en orden.
func salidasEnCola(t *testing.T, dataDir string) []int {
	t.Helper()
	cola, err := transport.Load(dataDir)
	if err != nil {
		t.Fatalf("leer la cola: %v", err)
	}
	var out []int
	for _, ev := range cola {
		out = append(out, ev.TokensOutput)
	}
	return out
}

// generarCon ejecuta una pasada de `generate()` con el reloj en `ahora` y devuelve la cola.
func generarCon(t *testing.T, ahora time.Time, logs string) []int {
	t.Helper()
	dataDir := testutil.Sandbox(t)
	if _, _, err := agenteConReloj(t, dataDir, logs, ahora).generate(); err != nil {
		t.Fatalf("precondición: generate() falló: %v", err)
	}
	return salidasEnCola(t, dataDir)
}

// entornoDeRun prepara un sandbox para `--run` en proceso: sin enrolar, sin endpoint (el sync se omite) y
// con la raíz de logs en un temporal. Devuelve el dataDir y la raíz de logs.
func entornoDeRun(t *testing.T) (string, string) {
	t.Helper()
	dataDir := testutil.Sandbox(t)
	logs := t.TempDir()
	b, err := json.Marshal(map[string]any{"tools": []string{"claude_code"}, "sync_interval": "60s", "logs_root": logs})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "config.json"), b, 0o600); err != nil {
		t.Fatalf("escribir config: %v", err)
	}
	return dataDir, logs
}

// (15) · P-007 FR-011, FR-014, SC-006 — DOS PROCESOS `--run`. El primero lee P entero y la primera línea de M:
// cierra P por la regla (i), y M queda abierto, sin encolar y con el offset en su comienzo. Se añaden la línea
// final de M y la primera de N. El segundo relee M desde su comienzo y lo encola UNA vez, con la final.
func TestRetencion_SC006_DosProcesos(t *testing.T) {
	dataDir, logs := entornoDeRun(t)
	ahora := time.Now().UTC()
	log := filepath.Join(logs, "sesion.jsonl")
	p, m1 := lineaRetencion("P", 5, ahora), lineaRetencion("M", 7, ahora)
	escribirLog(t, log, p+m1, ahora)

	if codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "--run"); codigo != 0 {
		t.Fatalf("precondición: el primer `--run` salió con %d: %s", codigo, stderr)
	}
	t.Run("primera_pasada_no_encola_M", func(t *testing.T) {
		if got := salidasEnCola(t, dataDir); len(got) != 1 || got[0] != 5 {
			t.Errorf("tras la primera pasada la cola debe tener sólo P (salida 5); tiene salidas %v", got)
		}
	})
	t.Run("offset_en_el_comienzo_de_M", func(t *testing.T) {
		st, err := state.Load(filepath.Join(dataDir, "state.json"))
		if err != nil {
			t.Fatalf("leer state.json: %v", err)
		}
		if got := st.Files[log].Offset; got != int64(len(p)) {
			t.Errorf("offset = %d; se esperaba %d, el comienzo de la primera línea de M", got, len(p))
		}
	})

	f, err := os.OpenFile(log, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(lineaRetencion("M", 89817, ahora) + lineaRetencion("N", 3, ahora)); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "--run"); codigo != 0 {
		t.Fatalf("precondición: el segundo `--run` salió con %d: %s", codigo, stderr)
	}
	t.Run("segunda_pasada_encola_M_una_vez_con_la_final", func(t *testing.T) {
		if got := salidasEnCola(t, dataDir); len(got) != 2 || got[1] != 89817 {
			t.Errorf("tras la segunda pasada la cola debe ser P y M con la final (5, 89817); es %v", got)
		}
	})
}

// (16) · P-007 FR-010, SC-007 — la regla (ii), con el reloj y el `mtime` fijados.
func TestRetencion_SC007_CierrePorT(t *testing.T) {
	t0 := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	caso := func(t *testing.T, momento, ahora, mtime time.Time) []int {
		logs := t.TempDir()
		escribirLog(t, filepath.Join(logs, "s.jsonl"), lineaRetencion("T", 42, momento), mtime)
		return generarCon(t, ahora, logs)
	}
	t.Run("a_las_dos_a_T_emitido", func(t *testing.T) {
		if got := caso(t, t0, t0.Add(esperaT), t0); len(got) != 1 {
			t.Errorf("con el timestamp y el mtime a T, el mensaje debe encolarse; cola %v", got)
		}
	})
	t.Run("b_mtime_a_T_menos_1s_retenido", func(t *testing.T) {
		if got := caso(t, t0, t0.Add(esperaT), t0.Add(time.Second)); len(got) != 0 {
			t.Errorf("con el mtime a T − 1 s, el mensaje debe seguir retenido; cola %v", got)
		}
	})
	t.Run("c_timestamp_a_T_menos_1s_retenido", func(t *testing.T) {
		if got := caso(t, t0.Add(time.Second), t0.Add(esperaT), t0); len(got) != 0 {
			t.Errorf("con el timestamp a T − 1 s, el mensaje debe seguir retenido; cola %v", got)
		}
	})
}

// (23) · P-007 FR-010, SC-007 (E-3) — el tope de 24 h, con el `mtime` de ahora.
func TestRetencion_TopeDe24Horas(t *testing.T) {
	ahora := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	caso := func(t *testing.T, momento time.Time) []int {
		logs := t.TempDir()
		escribirLog(t, filepath.Join(logs, "s.jsonl"), lineaRetencion("D", 42, momento), ahora)
		return generarCon(t, ahora, logs)
	}
	t.Run("a_24h_emitido", func(t *testing.T) {
		if got := caso(t, ahora.Add(-24*time.Hour)); len(got) != 1 {
			t.Errorf("a las 24 h de su última línea el mensaje debe encolarse aunque el fichero cambie; cola %v", got)
		}
	})
	t.Run("a_24h_menos_1s_retenido", func(t *testing.T) {
		if got := caso(t, ahora.Add(-24*time.Hour+time.Second)); len(got) != 0 {
			t.Errorf("a las 24 h − 1 s, con el fichero recién cambiado, el mensaje debe seguir retenido; cola %v", got)
		}
	})
}

// (17) · P-007 FR-010, SC-008 — A seguido de B en el mismo fichero encola A; con B en otro fichero, A sigue
// retenido.
func TestRetencion_SC008_CierrePorMensajePosterior(t *testing.T) {
	ahora := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	t.Run("mismo_fichero", func(t *testing.T) {
		logs := t.TempDir()
		escribirLog(t, filepath.Join(logs, "s.jsonl"), lineaRetencion("A", 11, ahora)+lineaRetencion("B", 22, ahora), ahora)
		if got := generarCon(t, ahora, logs); len(got) != 1 || got[0] != 11 {
			t.Errorf("se esperaba encolado sólo A (11); cola %v", got)
		}
	})
	t.Run("otro_fichero", func(t *testing.T) {
		logs := t.TempDir()
		escribirLog(t, filepath.Join(logs, "1.jsonl"), lineaRetencion("A", 11, ahora), ahora)
		escribirLog(t, filepath.Join(logs, "2.jsonl"), lineaRetencion("B", 22, ahora), ahora)
		if got := generarCon(t, ahora, logs); len(got) != 0 {
			t.Errorf("B en otro fichero no cierra A; cola %v", got)
		}
	})
}

// (19) · P-007 FR-014, P-5 — `--run` avisa de lo que queda abierto, con el texto aprobado, literal.
func TestRetencion_AvisoDeRun(t *testing.T) {
	_, logs := entornoDeRun(t)
	ahora := time.Now().UTC()
	escribirLog(t, filepath.Join(logs, "s.jsonl"), lineaRetencion("V", 9, ahora), ahora)

	codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "--run")
	if codigo != 0 {
		t.Fatalf("precondición: `--run` salió con %d: %s", codigo, stderr)
	}
	const aviso = "1 mensajes siguen abiertos: se enviarán en la próxima pasada\n"
	if !strings.Contains(stderr, aviso) {
		t.Errorf("stderr no lleva el aviso aprobado %q:\n%s", aviso, stderr)
	}
}

// (20) · P-007 FR-021 — una pasada que sólo relee el mensaje en espera no tiene novedades: el demonio no
// escribe su resumen.
func TestRetencion_TickCallaSiSoloRelee(t *testing.T) {
	ahora := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	dataDir := testutil.Sandbox(t)
	logs := t.TempDir()
	escribirLog(t, filepath.Join(logs, "s.jsonl"), lineaRetencion("R", 9, ahora), ahora)
	a := agenteConReloj(t, dataDir, logs, ahora)
	if _, _, err := a.generate(); err != nil {
		t.Fatalf("precondición: la primera pasada falló: %v", err)
	}
	_, segunda, err := a.generate()
	if err != nil {
		t.Fatalf("precondición: la segunda pasada falló: %v", err)
	}
	if got := segunda.Recuentos().Releidas; got != 1 {
		t.Errorf("la segunda pasada debe releer la línea del mensaje en espera; Releidas = %d", got)
	}
	if segunda.HayNovedades() {
		t.Errorf("una pasada que sólo relee no tiene novedades: el demonio no debe escribir su resumen")
	}
}

// (22) · P-007 FR-011, SC-009 — un mensaje abierto con CENTINELAS de identificador: tras la pasada, ningún
// fichero del directorio de datos los contiene, y `state.json` sigue con sus cuatro campos.
func TestRetencion_SC009_NadaDelProveedorEnDisco(t *testing.T) {
	const centinelaMensaje, centinelaPeticion = "msg_CENTINELA7RETENIDO0000001", "req_CENTINELA7RETENIDO0000001"
	ahora := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	dataDir := testutil.Sandbox(t)
	logs := t.TempDir()
	escribirLog(t, filepath.Join(logs, "s.jsonl"), lineaConIDs(centinelaMensaje, centinelaPeticion, 9, ahora), ahora)
	_, pasada, err := agenteConReloj(t, dataDir, logs, ahora).generate()
	if err != nil || pasada.Recuentos().EnEspera != 1 {
		t.Fatalf("precondición: la pasada debe dejar el mensaje abierto (err %v)", err)
	}

	t.Run("sin_centinelas", func(t *testing.T) {
		err := filepath.WalkDir(dataDir, func(ruta string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			b, err := os.ReadFile(ruta)
			if err != nil {
				return err
			}
			for _, c := range []string{centinelaMensaje, centinelaPeticion, "CENTINELA7RETENIDO"} {
				if strings.Contains(string(b), c) {
					t.Errorf("%s contiene un identificador del proveedor (%s)", filepath.Base(ruta), c)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("recorrer el directorio de datos: %v", err)
		}
	})
	t.Run("state_json_cuatro_campos", func(t *testing.T) {
		b, err := os.ReadFile(filepath.Join(dataDir, "state.json"))
		if err != nil {
			t.Fatalf("leer state.json: %v", err)
		}
		var crudo struct {
			Files map[string]map[string]any `json:"files"`
		}
		if err := json.Unmarshal(b, &crudo); err != nil {
			t.Fatalf("state.json no es JSON: %v", err)
		}
		for _, campos := range crudo.Files {
			if len(campos) != 4 {
				t.Errorf("una entrada de state.json tiene %d campos; se esperaban 4 (path, size, mod_time, offset)", len(campos))
			}
			for _, k := range []string{"path", "size", "mod_time", "offset"} {
				if _, ok := campos[k]; !ok {
					t.Errorf("a una entrada de state.json le falta %q", k)
				}
			}
		}
	})
}
