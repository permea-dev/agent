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

// P-011 B2 · Una línea de Copilot CLI (plan §1; FR-010 a FR-017, FR-020; D-3). Los fixtures son SINTÉTICOS
// (`testdata/copilot/generar.sh`, `lineas/`), y se leen de una copia en `t.TempDir()`, nunca en su sitio. Cada test
// afirma SÓLO su campo.

const salCopilot = "sal-de-prueba-copilot"

// lineasCopilot devuelve las líneas de `lineas/<nombre>.jsonl`, leídas de una copia.
func lineasCopilot(t *testing.T, nombre string) [][]byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "copilot", "lineas", nombre+".jsonl"))
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

func contextoCopilotDePrueba() ContextoCopilot {
	return ContextoCopilot{
		Context: Context{Salt: salCopilot, MachineID: "maquina-de-prueba", DevID: "dev-de-prueba", OrgID: "org-de-prueba",
			AgentVersion: "test"},
		SessionID: "s-000000000000000000000001",
		Cwd:       "/sintetico/proyecto-a",
	}
}

// registroCopilot lee la (única) línea de un fixture.
func registroCopilot(t *testing.T, nombre string) (*eventoCopilotDePrueba, ClaseCopilot) {
	t.Helper()
	ev, clase, marcas, err := RegistroCopilot(lineasCopilot(t, nombre)[0], contextoCopilotDePrueba())
	if err != nil {
		t.Fatalf("RegistroCopilot(%s): error %v", nombre, err)
	}
	if ev == nil {
		return nil, clase
	}
	return &eventoCopilotDePrueba{Event: *ev, MarcasCopilot: marcas}, clase
}

type eventoCopilotDePrueba struct {
	event.Event
	MarcasCopilot
}

// eventoCopilot exige que el fixture sea un evento.
func eventoCopilot(t *testing.T, nombre string) *eventoCopilotDePrueba {
	t.Helper()
	ev, clase := registroCopilot(t, nombre)
	if clase != EventoCopilot || ev == nil {
		t.Fatalf("%s: clase %d, evento %v; se esperaba un evento", nombre, clase, ev != nil)
	}
	return ev
}

// (6) · SC-005, D-3: entrada = input − lectura − escritura; salida = output, SIN sumarle el razonamiento.
func TestRegistroCopilot_SC005Partidas(t *testing.T) {
	ev := eventoCopilot(t, "sc005") // input 100 · lectura 40 · escritura 2 · salida 7 · razonamiento 3
	if ev.TokensInput != 58 || ev.TokensOutput != 7 || ev.TokensCacheRead != 40 || ev.TokensCacheCreation != 2 {
		t.Errorf("partidas = %d/%d/%d/%d; se esperaba 58/7/40/2 (entrada/salida/lectura/escritura)",
			ev.TokensInput, ev.TokensOutput, ev.TokensCacheRead, ev.TokensCacheCreation)
	}
}

// (7) · La compactación con éxito es un registro, con las partidas de `compactionTokensUsed`.
func TestRegistroCopilot_Compactacion(t *testing.T) {
	t.Run("partidas", func(t *testing.T) {
		ev := eventoCopilot(t, "compactacion") // input 1000 · lectura 600 · escritura 100 · salida 50
		if ev.TokensInput != 300 || ev.TokensOutput != 50 || ev.TokensCacheRead != 600 || ev.TokensCacheCreation != 100 ||
			!ev.Compactacion {
			t.Errorf("partidas = %d/%d/%d/%d, compactación %t; se esperaba 300/50/600/100 y compactación", ev.TokensInput,
				ev.TokensOutput, ev.TokensCacheRead, ev.TokensCacheCreation, ev.Compactacion)
		}
	})
	t.Run("fallida", func(t *testing.T) {
		// HOJA NACIDA VERDE con el stub (que no clasifica nada): la valida m14.
		if ev, clase := registroCopilot(t, "compactacion_fallida"); clase != NoEsRegistroCopilot || ev != nil {
			t.Errorf("success: false → clase %d; se esperaba «no es registro»", clase)
		}
	})
	t.Run("sin_ctu", func(t *testing.T) {
		if _, clase := registroCopilot(t, "compactacion_sin_ctu"); clase != IncoherenteCopilot {
			t.Errorf("sin compactionTokensUsed → clase %d; se esperaba incoherente", clase)
		}
	})
	t.Run("sin_requestid", func(t *testing.T) {
		if _, clase := registroCopilot(t, "compactacion_sin_requestid"); clase != SinIdentificadorCopilot {
			t.Errorf("sin requestId → clase %d; se esperaba sin identificador", clase)
		}
	})
}

// (8) · Incoherente, siete hojas (FR-020).
func TestRegistroCopilot_Incoherente(t *testing.T) {
	for hoja, fixture := range map[string]string{
		"usage_no_objeto": "incoherente_ctu_no_objeto",
		"cuatro_ausentes": "incoherente_cuatro_ausentes",
		"no_numerica":     "incoherente_no_numerica",
		"no_entera":       "incoherente_no_entera",
		"negativa":        "incoherente_negativa",
		"no_cabe":         "incoherente_no_cabe",
		"timestamp_malo":  "incoherente_timestamp_malo",
	} {
		t.Run(hoja, func(t *testing.T) {
			if ev, clase := registroCopilot(t, fixture); clase != IncoherenteCopilot || ev != nil {
				t.Errorf("%s → clase %d; se esperaba incoherente, sin evento", fixture, clase)
			}
		})
	}
}

// (9) · Sin `usageId` textual no vacío: «sin identificador», sin evento (FR-014).
func TestRegistroCopilot_SinIdentificador(t *testing.T) {
	for hoja, fixture := range map[string]string{
		"ausente":    "sin_usageid_ausente",
		"vacio":      "sin_usageid_vacio",
		"no_textual": "sin_usageid_no_textual",
	} {
		t.Run(hoja, func(t *testing.T) {
			if ev, clase := registroCopilot(t, fixture); clase != SinIdentificadorCopilot || ev != nil {
				t.Errorf("%s → clase %d; se esperaba sin identificador, sin evento", fixture, clase)
			}
		})
	}
}

// (10) · D-1, D-4: `tool = "copilot_cli"` y sin coste, siempre. Ni los créditos ni el razonamiento entran.
func TestRegistroCopilot_HerramientaYCoste(t *testing.T) {
	t.Run("tool", func(t *testing.T) {
		if ev := eventoCopilot(t, "evento"); ev.Tool != "copilot_cli" {
			t.Errorf("tool = %q; se esperaba copilot_cli", ev.Tool)
		}
	})
	t.Run("coste", func(t *testing.T) {
		if ev := eventoCopilot(t, "evento"); ev.CostUSD != 0 || ev.CostAvailable {
			t.Errorf("coste = %v, disponible %t; se esperaba 0 y false", ev.CostUSD, ev.CostAvailable)
		}
	})
	t.Run("nanoaiu", func(t *testing.T) {
		if ev := eventoCopilot(t, "compactacion"); ev.CostUSD != 0 {
			t.Errorf("coste = %v; los créditos no entran en el evento", ev.CostUSD)
		}
	})
	t.Run("razonamiento", func(t *testing.T) {
		if ev := eventoCopilot(t, "razonamiento"); ev.TokensOutput != 7 { // outputTokens 7, reasoningTokens 999
			t.Errorf("salida = %d; se esperaba 7 (el razonamiento va dentro de outputTokens)", ev.TokensOutput)
		}
	})
}

// (11) · P-4: `occurred_at` = la `timestamp` de la línea, en UTC.
func TestRegistroCopilot_MomentoEnUTC(t *testing.T) {
	quiero := time.Date(2026, 1, 1, 0, 0, 10, 0, time.UTC)
	for hoja, fixture := range map[string]string{"utc": "evento", "desfase": "desfase"} {
		t.Run(hoja, func(t *testing.T) {
			ev := eventoCopilot(t, fixture)
			if !ev.OccurredAt.Equal(quiero) || ev.OccurredAt.Location() != time.UTC {
				t.Errorf("occurred_at = %v (%v); se esperaba %v en UTC", ev.OccurredAt, ev.OccurredAt.Location(), quiero)
			}
		})
	}
}

// (12) · P-5: sin modelo, el evento sale igual, con `model` vacío y la marca «sin modelo».
func TestRegistroCopilot_SinModelo(t *testing.T) {
	ev := eventoCopilot(t, "sin_modelo")
	if ev.Model != "" || !ev.SinModelo {
		t.Errorf("model = %q, sin modelo %t; se esperaba vacío y marcado", ev.Model, ev.SinModelo)
	}
}

// (13) · P-11: el prefiltro. Un marcador dentro de un texto va escapado y no casa; las `model.*` no se decodifican.
// NACE VERDE con el stub (que no decodifica nada): lo valida m19.
func TestRegistroCopilot_Prefiltro(t *testing.T) {
	for _, fixture := range []string{"falso_user", "falso_system", "model_aux"} {
		t.Run(fixture, func(t *testing.T) {
			n := 0
			ctx := contextoCopilotDePrueba()
			ctx.Decodificadas = &n
			for _, l := range lineasCopilot(t, fixture) {
				if ev, clase, _, err := RegistroCopilot(l, ctx); err != nil || ev != nil || clase != NoEsRegistroCopilot {
					t.Errorf("clase %d, evento %v, error %v; se esperaba «no es registro»", clase, ev != nil, err)
				}
			}
			if n != 0 {
				t.Errorf("decodificadas = %d; se esperaba 0", n)
			}
		})
	}
}

// (14) · FR-019: nada del proveedor viaja. Centinelas en todos los campos de texto de las dos clases de registro, y en la
// sesión y el `cwd` del contexto. ROJO ANTES DEL VERDE: con el stub no hay evento.
func TestRegistroCopilot_CentinelasFueraDelEvento(t *testing.T) {
	for _, fixture := range []string{"centinelas", "centinelas_compactacion"} {
		t.Run(fixture, func(t *testing.T) {
			ctx := contextoCopilotDePrueba()
			ctx.SessionID = "CENTINELA-sessionId"
			ctx.Cwd = "/sintetico/CENTINELA-cwd"
			ev, clase, _, err := RegistroCopilot(lineasCopilot(t, fixture)[0], ctx)
			if err != nil || clase != EventoCopilot || ev == nil {
				t.Fatalf("clase %d, error %v; se esperaba un evento", clase, err)
			}
			b, err := json.Marshal(ev)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(b), "CENTINELA") {
				t.Errorf("el evento lleva un centinela: %s", b)
			}
		})
	}
}
