package ingest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/sqlitero"
)

// P-010 B3 · Una llamada de OpenCode. Fixtures SINTÉTICOS en testdata/opencode/filas/: el `data` de cada fila, con
// identificadores `o-…` y `s-…` sin la forma real, centinelas `CENTINELA-…` y rutas `/sintetico/…`. El `id` es una
// columna de la fila, no va en `data`: sus formas las pone cada test. Cada test afirma SÓLO su campo (plan §Bloques).

const (
	salOpenCode        = "sal-de-prueba-opencode"
	idOpenCodePrueba   = "o-000000000000000000000001"
	modeloOpenCode     = "modelo-sintetico-opencode"
	sesionOpenCode     = "s-000000000000000000000001"
	directorioOpenCode = "/sintetico/proyecto-opencode"
	rutaOpenCode       = "/sintetico/ruta-opencode"
	creadaOpenCode     = 1760000000000 // `time.created` de los fixtures, en ms
	completadaOpenCode = 1760000004321 // `time.completed`
)

func contextoOpenCodeDePrueba() ContextoOpenCode {
	return ContextoOpenCode{
		Context: Context{Salt: salOpenCode, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba", OrgID: "org-de-prueba",
			AgentVersion: "test"},
		SessionID: sesionOpenCode,
		Directory: directorioOpenCode,
		Path:      rutaOpenCode,
	}
}

func idTexto(id string) sqlitero.Valor { return sqlitero.Valor{Tipo: sqlitero.Texto, B: []byte(id)} }

// filaDeFixture arma una fila con el `id` dado y el `data` del fixture.
func filaDeFixture(t *testing.T, nombre string, id sqlitero.Valor) FilaOpenCode {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "opencode", "filas", nombre))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	return FilaOpenCode{ID: id, Data: b}
}

// eventoOpenCode exige que la fila sea un evento: si no, el test no puede seguir.
func eventoOpenCode(t *testing.T, nombre string, ctx ContextoOpenCode) event.Event {
	t.Helper()
	ev, clase := LlamadaOpenCode(filaDeFixture(t, nombre, idTexto(idOpenCodePrueba)), ctx)
	if clase != EventoOpenCode || ev == nil {
		t.Fatalf("LlamadaOpenCode(%s): clase %d, evento %v; se esperaba un evento", nombre, clase, ev)
	}
	return *ev
}

// partidas es (entrada, salida, lectura, escritura).
func partidas(ev event.Event) [4]int {
	return [4]int{ev.TokensInput, ev.TokensOutput, ev.TokensCacheRead, ev.TokensCacheCreation}
}

// (14) · SC-012, P-2: 100/7/3/40/2 → 100/10/40/2.
func TestLlamadaOpenCode_PartidasSC012(t *testing.T) {
	ev := eventoOpenCode(t, "partidas_sc012.json", contextoOpenCodeDePrueba())
	if got, want := partidas(ev), [4]int{100, 10, 40, 2}; got != want {
		t.Errorf("partidas (entrada, salida, lectura, escritura) = %v; se esperaba %v (salida = 7 + 3)", got, want)
	}
}

// (15) · P-2: una partida ausente vale 0.
func TestLlamadaOpenCode_PartidaAusenteVale0(t *testing.T) {
	for _, c := range []struct {
		hoja, fixture string
		want          [4]int
	}{
		{"sin_cache_ni_razonamiento", "partidas_ausentes.json", [4]int{5, 2, 0, 0}},
		{"tokens_vacio", "tokens_vacio.json", [4]int{0, 0, 0, 0}},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if got := partidas(eventoOpenCode(t, c.fixture, contextoOpenCodeDePrueba())); got != c.want {
				t.Errorf("partidas (entrada, salida, lectura, escritura) = %v; se esperaba %v", got, c.want)
			}
		})
	}
}

// (16) · FR-021: las siete formas de «incoherente». No producen evento.
func TestLlamadaOpenCode_Incoherente(t *testing.T) {
	for _, c := range []struct{ hoja, fixture string }{
		{"data_no_json", "incoherente_no_json.json"},
		{"data_no_objeto", "incoherente_no_objeto.json"},
		{"tokens_no_objeto", "incoherente_tokens_no_objeto.json"},
		{"partida_no_numerica", "incoherente_no_numerica.json"},
		{"partida_no_entera", "incoherente_no_entera.json"},
		{"partida_negativa", "incoherente_negativa.json"},
		{"completed_no_numerico", "incoherente_completed_no_numerico.json"},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			ev, clase := LlamadaOpenCode(filaDeFixture(t, c.fixture, idTexto(idOpenCodePrueba)), contextoOpenCodeDePrueba())
			if clase != IncoherenteOpenCode || ev != nil {
				t.Errorf("%s: clase %d, evento %v; se esperaba IncoherenteOpenCode (%d), sin evento", c.fixture, clase, ev,
					IncoherenteOpenCode)
			}
		})
	}
}

// (17) · FR-014, D-010-P14: sin un `id` textual no vacío no hay evento, y eso se mira antes que el `data`. Con él, el
// `event_id` es el del contrato.
func TestLlamadaOpenCode_Identificador(t *testing.T) {
	for _, c := range []struct {
		hoja, fixture string
		id            sqlitero.Valor
	}{
		{"ausente", "evento.json", sqlitero.Valor{Tipo: sqlitero.Nulo}},
		{"vacio", "evento.json", idTexto("")},
		{"no_textual", "evento.json", sqlitero.Valor{Tipo: sqlitero.Entero, I: 7}},
		{"antes_que_incoherente", "incoherente_no_json.json", sqlitero.Valor{Tipo: sqlitero.Nulo}},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			ev, clase := LlamadaOpenCode(filaDeFixture(t, c.fixture, c.id), contextoOpenCodeDePrueba())
			if clase != SinIdentificadorOpenCode || ev != nil {
				t.Errorf("%s: clase %d, evento %v; se esperaba SinIdentificadorOpenCode (%d), sin evento", c.fixture, clase, ev,
					SinIdentificadorOpenCode)
			}
		})
	}
	t.Run("event_id", func(t *testing.T) {
		ev := eventoOpenCode(t, "evento.json", contextoOpenCodeDePrueba())
		if want, _ := derivarEventIDOpenCode(idOpenCodePrueba); ev.EventID != want {
			t.Errorf("event_id = %q; el del contrato es %q", ev.EventID, want)
		}
	})
}

// (18) · FR-018, SC-028, D-1: `tool = "opencode"`, sin coste, y el `$.cost` de OpenCode no entra.
func TestLlamadaOpenCode_SinCoste(t *testing.T) {
	t.Run("tool", func(t *testing.T) {
		if ev := eventoOpenCode(t, "evento.json", contextoOpenCodeDePrueba()); ev.Tool != "opencode" {
			t.Errorf("tool = %q; se esperaba \"opencode\"", ev.Tool)
		}
	})
	t.Run("coste_cero", func(t *testing.T) {
		if ev := eventoOpenCode(t, "evento.json", contextoOpenCodeDePrueba()); ev.CostUSD != 0 {
			t.Errorf("cost_usd = %v; se esperaba 0", ev.CostUSD)
		}
	})
	t.Run("cost_available", func(t *testing.T) {
		if ev := eventoOpenCode(t, "evento.json", contextoOpenCodeDePrueba()); ev.CostAvailable {
			t.Error("cost_available = true; se esperaba false")
		}
	})
	t.Run("cost", func(t *testing.T) {
		if ev := eventoOpenCode(t, "evento_con_coste.json", contextoOpenCodeDePrueba()); ev.CostUSD != 0 {
			t.Errorf("cost_usd = %v con `$.cost` = 999; se esperaba 0", ev.CostUSD)
		}
	})
}

// (19) · FR-017, SC-028, P-7: `occurred_at` = `time.completed`, en UTC, y no `time.created`.
func TestLlamadaOpenCode_OccurredAtEsCompleted(t *testing.T) {
	ev := eventoOpenCode(t, "evento.json", contextoOpenCodeDePrueba())
	if want := time.UnixMilli(completadaOpenCode).UTC(); !ev.OccurredAt.Equal(want) || ev.OccurredAt.Location() != time.UTC {
		t.Errorf("occurred_at = %v; se esperaba %v (`time.completed`, UTC; `time.created` es %v)", ev.OccurredAt, want,
			time.UnixMilli(creadaOpenCode).UTC())
	}
}

// (20) · FR-016, P-5: sin `model.id` textual no vacío, el evento sale con el modelo vacío, que es su marca «sin modelo».
// Con él, tal cual.
func TestLlamadaOpenCode_SinModelo(t *testing.T) {
	for _, c := range []struct{ hoja, fixture, want string }{
		{"ausente", "sin_modelo_ausente.json", ""},
		{"vacio", "sin_modelo_vacio.json", ""},
		{"no_textual", "sin_modelo_no_textual.json", ""},
		{"tal_cual", "evento.json", modeloOpenCode},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			if ev := eventoOpenCode(t, c.fixture, contextoOpenCodeDePrueba()); ev.Model != c.want {
				t.Errorf("model = %q; se esperaba %q", ev.Model, c.want)
			}
		})
	}
}

// (21) · FR-019, SC-028, P-6: `session_ref` de `session_id` y `project_ref` de `directory`, con sal; sin la fila de
// `session_v2`, las dos vacías.
func TestLlamadaOpenCode_Referencias(t *testing.T) {
	t.Run("sesion", func(t *testing.T) {
		ev := eventoOpenCode(t, "evento.json", contextoOpenCodeDePrueba())
		if want := event.Ref(salOpenCode, sesionOpenCode); ev.SessionRef != want {
			t.Errorf("session_ref = %q; se esperaba event.Ref(sal, session_id) = %q", ev.SessionRef, want)
		}
	})
	t.Run("proyecto", func(t *testing.T) {
		ev := eventoOpenCode(t, "evento.json", contextoOpenCodeDePrueba())
		if want := project.Derivar(directorioOpenCode, salOpenCode); ev.ProjectRef != want {
			t.Errorf("project_ref = %q; se esperaba Derivar(directory, sal) = %q", ev.ProjectRef, want)
		}
	})
	t.Run("sin_sesion", func(t *testing.T) {
		ctx := contextoOpenCodeDePrueba()
		ctx.SessionID, ctx.Directory, ctx.Path = "", "", ""
		if ev := eventoOpenCode(t, "evento.json", ctx); ev.SessionRef != "" || ev.ProjectRef != "" {
			t.Errorf("session_ref = %q, project_ref = %q; sin sesión, se esperaban vacías", ev.SessionRef, ev.ProjectRef)
		}
	})
}

// (22) · FR-020, SC-014: ningún centinela del `data`, de la sesión ni del `id` llega al evento. Rojo antes del verde
// (H19): con el stub no hay evento que inspeccionar.
func TestLlamadaOpenCode_CentinelasFueraDelEvento(t *testing.T) {
	ctx := contextoOpenCodeDePrueba()
	ctx.SessionID = "s-CENTINELA-sesion"
	ctx.Directory = "/sintetico/CENTINELA-directorio"
	ctx.Path = "/sintetico/CENTINELA-ruta"
	ev, clase := LlamadaOpenCode(filaDeFixture(t, "centinelas.json", idTexto("o-CENTINELA-llamada")), ctx)
	if clase != EventoOpenCode || ev == nil {
		t.Fatalf("clase %d, evento %v; se esperaba un evento que inspeccionar", clase, ev)
	}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if strings.Contains(string(b), "CENTINELA") {
		t.Errorf("el evento lleva un centinela: %s", b)
	}
}
