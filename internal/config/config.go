// Package config define la configuración local del agente (fichero legible por el usuario)
// y la resolución de rutas por sistema operativo (Principio III: nunca se hardcodean rutas).
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

// ProjectRefMode controla si project_ref se envía como hash (por defecto) o en claro (opt-in).
type ProjectRefMode string

const (
	// ModeHash envía project_ref como hash salado (por defecto: máxima privacidad).
	ModeHash ProjectRefMode = "hash"
	// ModePlain envía project_ref en claro (opt-in explícito del usuario).
	ModePlain ProjectRefMode = "plain"
)

// Config es el contenido del fichero de configuración local.
type Config struct {
	Endpoint       string         `json:"endpoint"`
	DeviceToken    string         `json:"device_token"`
	OrgID          string         `json:"org_id"`
	DevID          string         `json:"dev_id"`
	ProjectRefMode ProjectRefMode `json:"project_ref_mode"`
	Tools          []string       `json:"tools"`
	SyncInterval   string         `json:"sync_interval"`
	// LogsRoot es un override opcional de la raíz de logs de Claude Code para
	// instalaciones no estándar (R1). Vacío -> se resuelve ~/.claude/projects por SO.
	LogsRoot string `json:"logs_root,omitempty"`
}

// Default devuelve la configuración por defecto: máxima privacidad, Claude Code.
func Default() Config {
	return Config{
		ProjectRefMode: ModeHash,
		Tools:          []string{"claude_code"},
		SyncInterval:   "60s",
	}
}

// DataDir resuelve el directorio de datos del agente por SO (R8) y lo crea si no
// existe: Linux `$XDG_CONFIG_HOME/permea`, macOS `~/Library/Application Support/permea`,
// Windows `%AppData%\permea`. Allí viven config.json, state.json, queue.jsonl y salt.
func DataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "permea")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// Load lee config.json aplicando Default() a los campos vacíos. Un fichero
// inexistente devuelve la configuración por defecto sin error (arranque limpio).
func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Default(), err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Default(), err
	}
	cfg.applyDefaults()
	return cfg, nil
}

// Save escribe config.json de forma atómica (temporal + rename en el mismo dir).
func Save(path string, cfg Config) error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, append(b, '\n'), 0o600)
}

// Validate comprueba invariantes de transporte: si hay endpoint configurado, DEBE
// ser https:// (FR-009, R9). Un endpoint vacío aún no configurado no es error.
func (c Config) Validate() error {
	if c.Endpoint == "" {
		return nil
	}
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return fmt.Errorf("endpoint inválido %q: %w", c.Endpoint, err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("endpoint debe ser https://, got %q", c.Endpoint)
	}
	return nil
}

// ClaudeCodeLogsRoot resuelve la raíz de logs de Claude Code por SO (R1): usa el
// override LogsRoot si está presente, o `~/.claude/projects` resuelto por SO.
func ClaudeCodeLogsRoot(c Config) (string, error) {
	if c.LogsRoot != "" {
		return c.LogsRoot, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

func (c *Config) applyDefaults() {
	d := Default()
	if c.ProjectRefMode == "" {
		c.ProjectRefMode = d.ProjectRefMode
	}
	if len(c.Tools) == 0 {
		c.Tools = d.Tools
	}
	if c.SyncInterval == "" {
		c.SyncInterval = d.SyncInterval
	}
}

// atomicWrite escribe data en path de forma atómica: fichero temporal en el mismo
// directorio + os.Rename (mismo sistema de ficheros). Evita estado corrupto ante caída.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Limpieza best-effort: tras un rename correcto el temporal ya no existe.
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Chmod(perm); err != nil {
		return errors.Join(err, tmp.Close())
	}
	// Close DEBE comprobarse: un fallo aquí puede significar datos sin volcar; en ese
	// caso NO se hace rename y se devuelve el error (durabilidad, FR-007).
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
