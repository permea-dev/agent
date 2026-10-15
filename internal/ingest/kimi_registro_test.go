package ingest

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/event"
)

// P-012 B2 · Una línea de Kimi Code (plan §1; FR-010 a FR-017, FR-020). Los fixtures son SINTÉTICOS (`testdata/kimi/generar.sh`,
// `lineas/`), y se leen de una copia en `t.TempDir()`, nunca en su sitio. Cada test afirma SÓLO su campo.

const salKimi = "sal-de-prueba-kimi"

// lineasKimi devuelve las líneas de `lineas/<nombre>.jsonl`, leídas de una copia.
func lineasKimi(t *testing.T, nombre string) [][]byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "kimi", "lineas", nombre+".jsonl"))
	if err != nil {
		t.Fatalf("precondición: fixture %s: %v", nombre, err)
	}
	copia := filepath.Join(t.TempDir(), nombre+".jsonl")
	if err := os.WriteFile(copia, b, 0o600); err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(copia)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.Split(bytes.TrimRight(b, "\n"), []byte("\n"))
}

func contextoKimiDePrueba() ContextoKimi {
	return ContextoKimi{
		Context: Context{Salt: salKimi, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba", OrgID: "org-de-prueba",
			AgentVersion: "test"},
		S: sesionKimi1, A: "main", N: 1,
		Cwd: "/sintetico/proyecto-a",
	}
}

type eventoKimiDePrueba struct {
	event.Event
	MarcasKimi
}

// registroKimiLinea clasifica una línea con un contexto.
func registroKimiLinea(t *testing.T, linea []byte, ctx ContextoKimi) (*eventoKimiDePrueba, ClaseKimi) {
	t.Helper()
	ev, clase, marcas, err := RegistroKimi(linea, ctx)
	if err != nil {
		t.Fatalf("RegistroKimi: error %v", err)
	}
	if ev == nil {
		return nil, clase
	}
	return &eventoKimiDePrueba{Event: *ev, MarcasKimi: marcas}, clase
}

// eventoKimi exige que la (primera) línea del fixture sea un evento.
func eventoKimi(t *testing.T, nombre string) *eventoKimiDePrueba {
	t.Helper()
	ev, clase := registroKimiLinea(t, lineasKimi(t, nombre)[0], contextoKimiDePrueba())
	if clase != EventoKimi || ev == nil {
		t.Fatalf("%s: clase %d, evento %v; se esperaba un evento", nombre, clase, ev != nil)
	}
	return ev
}

// (5) · SC-005, FR-013: las partidas pasan TAL CUAL, sin sumas ni restas.
func TestRegistroKimi_SC005Partidas(t *testing.T) {
	ev := eventoKimi(t, "sc005") // inputOther 100 · output 7 · inputCacheRead 40 · inputCacheCreation 2
	if ev.TokensInput != 100 || ev.TokensCacheRead != 40 || ev.TokensCacheCreation != 2 || ev.TokensOutput != 7 {
		t.Errorf("partidas = %d/%d/%d/%d; se esperaba 100/40/2/7 (entrada/lectura/escritura/salida)",
			ev.TokensInput, ev.TokensCacheRead, ev.TokensCacheCreation, ev.TokensOutput)
	}
}

// (6) · SC-008, P-9: el modelo es el texto tras la ÚLTIMA `/`; vacío → «sin modelo».
func TestRegistroKimi_Modelo(t *testing.T) {
	for _, c := range []struct{ hoja, fixture, quiero string }{
		{"prefijo", "modelo_prefijo", "kimi-k2.7-code"},
		{"sin_barra", "modelo_sin_barra", "kimi-k2.7-code"},
		{"varias", "modelo_varias", "c"},
		{"inicial", "modelo_inicial", "x"},
		{"acabada", "modelo_acabada", ""},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			ev := eventoKimi(t, c.fixture)
			if ev.Model != c.quiero || ev.SinModelo != (c.quiero == "") {
				t.Errorf("model = %q, sin modelo %t; se esperaba %q", ev.Model, ev.SinModelo, c.quiero)
			}
		})
	}
	t.Run("vacio_o_no_textual", func(t *testing.T) {
		for _, l := range lineasKimi(t, "modelo_vacio_o_no_textual") {
			ev, clase := registroKimiLinea(t, l, contextoKimiDePrueba())
			if clase != EventoKimi || ev == nil || ev.Model != "" || !ev.SinModelo {
				t.Errorf("clase %d, evento %v; se esperaba un evento con model vacío y «sin modelo»", clase, ev)
			}
		}
	})
}

// (7) · Incoherente, siete hojas (FR-020).
func TestRegistroKimi_Incoherente(t *testing.T) {
	for _, hoja := range []string{"usage_no_objeto", "cuatro_ausentes", "no_numerica", "no_entera", "negativa", "time_textual",
		"time_ausente"} {
		t.Run(hoja, func(t *testing.T) {
			ev, clase := registroKimiLinea(t, lineasKimi(t, "incoherente_"+hoja)[0], contextoKimiDePrueba())
			if clase != IncoherenteKimi || ev != nil {
				t.Errorf("clase %d; se esperaba incoherente, sin evento", clase)
			}
		})
	}
}

// (8) · P-8: las cuatro partidas a 0 no se emiten: «a cero».
func TestRegistroKimi_ACero(t *testing.T) {
	if ev, clase := registroKimiLinea(t, lineasKimi(t, "a_cero")[0], contextoKimiDePrueba()); clase != ACeroKimi || ev != nil {
		t.Errorf("clase %d; se esperaba «a cero», sin evento", clase)
	}
}

// (9) · D-3: un registro heredado no se emite, y NO se decodifica.
func TestRegistroKimi_Heredado(t *testing.T) {
	n := 0
	ctx := contextoKimiDePrueba()
	ctx.Heredado, ctx.Decodificadas = true, &n
	if ev, clase := registroKimiLinea(t, lineasKimi(t, "evento")[0], ctx); clase != HeredadoKimi || ev != nil || n != 0 {
		t.Errorf("clase %d, decodificadas %d; se esperaba «heredado», sin evento y 0 decodificadas", clase, n)
	}
}

// (10) · D-1, D-5: `tool = "kimi_code"` y sin coste, siempre; `usageScope: session` marca la compactación.
func TestRegistroKimi_HerramientaCosteYCompactacion(t *testing.T) {
	t.Run("tool", func(t *testing.T) {
		if ev := eventoKimi(t, "evento"); ev.Tool != "kimi_code" {
			t.Errorf("tool = %q; se esperaba kimi_code", ev.Tool)
		}
	})
	t.Run("coste", func(t *testing.T) {
		if ev := eventoKimi(t, "evento"); ev.CostUSD != 0 || ev.CostAvailable {
			t.Errorf("coste = %v, disponible %t; se esperaba 0 y false", ev.CostUSD, ev.CostAvailable)
		}
	})
	t.Run("compactacion", func(t *testing.T) {
		if ev := eventoKimi(t, "compactacion"); !ev.Compactacion {
			t.Errorf("usageScope: session sin la marca de compactación")
		}
	})
	t.Run("turno", func(t *testing.T) {
		if ev := eventoKimi(t, "evento"); ev.Compactacion {
			t.Errorf("usageScope: turn con la marca de compactación")
		}
	})
}

// (11) · P-7: `occurred_at` = `time` en milisegundos, en UTC.
func TestRegistroKimi_MomentoEnUTC(t *testing.T) {
	ev := eventoKimi(t, "ms") // time 1
	quiero := time.Date(1970, 1, 1, 0, 0, 0, int(time.Millisecond), time.UTC)
	if !ev.OccurredAt.Equal(quiero) || ev.OccurredAt.Location() != time.UTC {
		t.Errorf("occurred_at = %v (%v); se esperaba %v en UTC", ev.OccurredAt, ev.OccurredAt.Location(), quiero)
	}
}

// (12) · P-10: el prefiltro está anclado al principio de la línea. El marcador dentro de un texto, una `forked` y un
// `subagent.completed` no se decodifican. NACE VERDE con el stub (que no decodifica nada): lo valida m22.
func TestRegistroKimi_Prefiltro(t *testing.T) {
	for _, fixture := range []string{"falso_texto", "forked", "subagent_completed"} {
		t.Run(fixture, func(t *testing.T) {
			n := 0
			ctx := contextoKimiDePrueba()
			ctx.Decodificadas = &n
			if ev, clase, _, err := RegistroKimi(lineasKimi(t, fixture)[0], ctx); err != nil || ev != nil ||
				clase != NoEsRegistroKimi || n != 0 {
				t.Errorf("clase %d, evento %v, error %v, decodificadas %d; se esperaba «no es registro» y 0", clase, ev != nil, err, n)
			}
		})
	}
}

// (13) · FR-019: nada del proveedor viaja. Centinelas en `agentId`, `usageScope` y un campo de más, y en S, A y el `cwd`
// del contexto. ROJO ANTES DEL VERDE: con el stub no hay evento.
func TestRegistroKimi_CentinelasFueraDelEvento(t *testing.T) {
	ctx := contextoKimiDePrueba()
	ctx.S, ctx.A, ctx.Cwd = "CENTINELA-sesion", "CENTINELA-agente", "/sintetico/CENTINELA-cwd"
	ev, clase, _, err := RegistroKimi(lineasKimi(t, "centinelas")[0], ctx)
	if err != nil || clase != EventoKimi || ev == nil {
		t.Fatalf("clase %d, error %v; se esperaba un evento", clase, err)
	}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "CENTINELA") {
		t.Errorf("el evento lleva un centinela: %s", b)
	}
}
