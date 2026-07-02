// Package config define la configuración local del agente (fichero legible por el usuario).
package config

// ProjectRefMode controla si project_ref se envía como hash (por defecto) o en claro (opt-in).
type ProjectRefMode string

const (
	ModeHash  ProjectRefMode = "hash"
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
}

// Default devuelve la configuración por defecto: máxima privacidad, Claude Code.
func Default() Config {
	return Config{
		ProjectRefMode: ModeHash,
		Tools:          []string{"claude_code"},
		SyncInterval:   "60s",
	}
}
