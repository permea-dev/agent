package config

import (
	"os"
	"path/filepath"
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
	if out.ProjectRefMode != ModeHash || out.SyncInterval != "60s" || len(out.Tools) != 1 || out.Tools[0] != "claude_code" {
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
	if out.ProjectRefMode != ModeHash || out.SyncInterval != "60s" || len(out.Tools) != 1 {
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
