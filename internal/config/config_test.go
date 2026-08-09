package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	in := Default()
	in.Endpoint = "https://ingest.example.com/v1/events"
	in.DeviceToken = "tok-123"
	in.OrgID = "org-1"
	in.DevID = "dev-42"
	if err := Save(path, in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	out, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if out.Endpoint != in.Endpoint || out.DeviceToken != in.DeviceToken || out.OrgID != in.OrgID || out.DevID != in.DevID {
		t.Errorf("round-trip no preserva campos: %+v", out)
	}
}

func TestLoad_MissingAppliesDefaults(t *testing.T) {
	// Fichero inexistente -> defaults, sin error.
	out, err := Load(filepath.Join(t.TempDir(), "no-existe.json"))
	if err != nil {
		t.Fatalf("Load de fichero inexistente no debe fallar: %v", err)
	}
	if out.SyncInterval != "60s" || len(out.Tools) != 1 || out.Tools[0] != "claude_code" {
		t.Errorf("defaults no aplicados: %+v", out)
	}
}

func TestLoad_PartialAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	// Config parcial: solo endpoint; el resto debe rellenarse por defecto.
	if err := os.WriteFile(path, []byte(`{"endpoint":"https://x.example/y"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if out.SyncInterval != "60s" || len(out.Tools) != 1 {
		t.Errorf("defaults no aplicados a config parcial: %+v", out)
	}
}

func TestValidate_RejectsNonHTTPS(t *testing.T) {
	c := Default()
	c.Endpoint = "http://inseguro.example/y"
	if err := c.Validate(); err == nil {
		t.Errorf("endpoint http:// debe rechazarse (FR-009)")
	}
	c.Endpoint = "https://seguro.example/y"
	if err := c.Validate(); err != nil {
		t.Errorf("endpoint https:// debe aceptarse, got %v", err)
	}
	// Endpoint vacío: aún no configurado, no es error en carga.
	c.Endpoint = ""
	if err := c.Validate(); err != nil {
		t.Errorf("endpoint vacío no debe fallar la validación: %v", err)
	}
}

func TestDataDir_Resolves(t *testing.T) {
	dir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir: %v", err)
	}
	if dir == "" {
		t.Fatalf("DataDir vacío")
	}
	if filepath.Base(dir) != "permea" {
		t.Errorf("DataDir debe terminar en 'permea': %q", dir)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("DataDir debe existir tras resolver: %v", err)
	}
}

func TestClaudeCodeLogsRoot_Override(t *testing.T) {
	c := Default()
	c.LogsRoot = "/ruta/custom/projects"
	got, err := ClaudeCodeLogsRoot(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/ruta/custom/projects" {
		t.Errorf("override de LogsRoot no respetado: %q", got)
	}

	c.LogsRoot = ""
	got, err = ClaudeCodeLogsRoot(c)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "projects" || filepath.Base(filepath.Dir(got)) != ".claude" {
		t.Errorf("ruta por defecto debe ser ~/.claude/projects: %q", got)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════════════
// P-004 T022b · FR-015 — el ajuste retirado DESAPARECE del fichero
// ═══════════════════════════════════════════════════════════════════════════════════════

// TestConfig_LaClaveRetiradaDesapareceDelFichero comprueba lo que FR-015 promete sobre el
// DISCO, no sobre el struct: tras un ciclo Load→Save, `project_ref_mode` ya no está en el
// `config.json`.
//
// SE RELEE EL JSON CRUDO, y es el punto del test: comprobarlo a través del struct sería
// TAUTOLÓGICO —una vez retirado el campo, jamás aparecería— y no probaría nada sobre lo que
// quedó escrito. Lo que respalda la afirmación de contracts/cli-config.md de que «un enroll
// posterior limpia la clave por sí solo» es exactamente esta relectura.
//
// El valor usado es "hash" a propósito: es el residual que FR-013a manda ignorar en
// silencio, así que el ciclo Load→Save debe completarse sin error y aun así dejar el fichero
// limpio. Con "plain" el arranque para (FR-013) y este ciclo no llegaría a ocurrir.
func TestConfig_LaClaveRetiradaDesapareceDelFichero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	original := `{
  "endpoint": "https://ejemplo.invalid",
  "device_token": "tok",
  "org_id": "org-1",
  "dev_id": "dev-1",
  "project_ref_mode": "hash",
  "tools": ["claude_code"],
  "sync_interval": "60s"
}`
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("sembrar config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	crudo, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("releer: %v", err)
	}

	var mapa map[string]any
	if err := json.Unmarshal(crudo, &mapa); err != nil {
		t.Fatalf("el fichero reescrito debe seguir siendo JSON válido: %v", err)
	}
	if _, presente := mapa["project_ref_mode"]; presente {
		t.Errorf("FR-015: tras Load→Save la clave `project_ref_mode` NO debe seguir en el fichero.\n"+
			"Mientras el campo exista en el struct, Save lo reescribe y la clave nunca se va.\n"+
			"contenido reescrito:\n%s", crudo)
	}

	// Contra-prueba: el ciclo no puede haberse «limpiado» borrando lo demás.
	for _, clave := range []string{"endpoint", "device_token", "org_id", "dev_id", "tools", "sync_interval"} {
		if _, presente := mapa[clave]; !presente {
			t.Errorf("el ciclo Load→Save perdió la clave válida %q: la retirada debe ser quirúrgica", clave)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════════════
// P-004 T022 · Los cuatro casos de la detección
//
// Se escriben JUNTO A T025 y no antes: un test contra una función inexistente NO COMPILA, y
// un rojo de compilación no dice nada sobre el comportamiento — solo que falta un símbolo.
// Nacen VERDES, y su valor lo dan las cuatro mutaciones que los validan (una por caso).
// ═══════════════════════════════════════════════════════════════════════════════════════

// TestCheckRetiredProjectRefMode cubre la tabla de contracts/cli-config.md §«Comportamiento
// ante la clave obsoleta». El criterio que separa los casos es QUÉ PEDÍA EL USUARIO, no qué
// clave escribió (D-004-1).
func TestCheckRetiredProjectRefMode(t *testing.T) {
	casos := []struct {
		nombre    string
		contenido string
		quiereErr bool
		porque    string
	}{
		{
			nombre:    `"plain" → PARA`,
			contenido: `{"endpoint":"https://x.invalid","project_ref_mode":"plain"}`,
			quiereErr: true,
			porque:    "FR-013: pidió algo que el producto ya no promete",
		},
		{
			nombre:    `"hash" → silencio`,
			contenido: `{"endpoint":"https://x.invalid","project_ref_mode":"hash"}`,
			quiereErr: false,
			porque:    "FR-013a: pidió exactamente lo que el agente hace siempre",
		},
		{
			nombre:    "otro valor → silencio",
			contenido: `{"endpoint":"https://x.invalid","project_ref_mode":"valor-inventado"}`,
			quiereErr: false,
			porque:    "FR-013b: no pide el modo retirado",
		},
		{
			nombre:    "clave ausente → silencio",
			contenido: `{"endpoint":"https://x.invalid"}`,
			quiereErr: false,
			porque:    "no hay nada que advertir",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(c.contenido), 0o600); err != nil {
				t.Fatalf("sembrar config: %v", err)
			}

			err := CheckRetiredProjectRefMode(path)
			if c.quiereErr && err == nil {
				t.Errorf("se esperaba señal de parada y no la hubo — %s", c.porque)
			}
			if !c.quiereErr && err != nil {
				t.Errorf("NO debía haber señal de parada y la hubo — %s\nerror: %v", c.porque, err)
			}
		})
	}
}

// TestCheckRetiredProjectRefMode_MensajeUtil comprueba lo que el contrato exige del mensaje:
// que nombre la clave, el valor y la RUTA REAL del fichero. Sin las tres cosas, el usuario
// tiene que adivinar cuál de sus ajustes es y dónde está.
//
// No se compara el texto completo —eso lo volvería frágil ante cualquier reescritura—; se
// comprueba que las tres piezas de información están.
func TestCheckRetiredProjectRefMode_MensajeUtil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"project_ref_mode":"plain"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := CheckRetiredProjectRefMode(path)
	if err == nil {
		t.Fatal("se esperaba señal de parada")
	}
	for _, pieza := range []string{"project_ref_mode", "plain", path} {
		if !strings.Contains(err.Error(), pieza) {
			t.Errorf("el mensaje debe nombrar %q para que el usuario sepa qué corregir y dónde\nmensaje: %v", pieza, err)
		}
	}
}

// TestCheckRetiredProjectRefMode_FallaHaciaNoInterrumpir: un fichero que no se puede leer o
// que no es JSON válido NO pide el modo retirado, así que esta comprobación calla. Su
// diagnóstico es trabajo de Load; adelantarlo aquí convertiría este error en un cajón de
// sastre y su mensaje dejaría de significar lo que dice.
func TestCheckRetiredProjectRefMode_FallaHaciaNoInterrumpir(t *testing.T) {
	dir := t.TempDir()

	t.Run("fichero inexistente", func(t *testing.T) {
		if err := CheckRetiredProjectRefMode(filepath.Join(dir, "no-existe.json")); err != nil {
			t.Errorf("una config inexistente no pide nada: %v", err)
		}
	})

	t.Run("JSON inválido", func(t *testing.T) {
		path := filepath.Join(dir, "roto.json")
		if err := os.WriteFile(path, []byte(`{esto no es json`), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := CheckRetiredProjectRefMode(path); err != nil {
			t.Errorf("una config ilegible no pide el modo retirado; su diagnóstico es de Load: %v", err)
		}
	})
}
