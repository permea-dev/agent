// Package event define el ÚNICO dato que cruza la frontera hacia el backend.
package event

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// SchemaVersion es la versión del contrato de frontera (spec-001).
const SchemaVersion = 1

// Event es un struct CERRADO: contiene exclusivamente los campos de la
// allowlist del Contrato de frontera. No existe passthrough de campos crudos
// del log. Lo que no tiene un campo aquí, no puede salir de la máquina.
type Event struct {
	SchemaVersion       int       `json:"schema_version"`
	AgentVersion        string    `json:"agent_version"`
	EventID             string    `json:"event_id"`
	OccurredAt          time.Time `json:"occurred_at"`
	Tool                string    `json:"tool"`
	Model               string    `json:"model"`
	TokensInput         int       `json:"tokens_input"`
	TokensOutput        int       `json:"tokens_output"`
	TokensCacheCreation int       `json:"tokens_cache_creation"`
	TokensCacheRead     int       `json:"tokens_cache_read"`
	CostUSD             float64   `json:"cost_usd"`
	ProjectRef          string    `json:"project_ref"`
	SessionRef          string    `json:"session_ref"`
	MachineRef          string    `json:"machine_ref"`
	DevID               string    `json:"dev_id"`
	OrgID               string    `json:"org_id"`
}

// Ref produce un identificador seguro para cruzar la frontera: hash salado del
// valor sensible (ruta de proyecto, id de sesión, id de máquina). El salt vive
// solo en local y NUNCA se transmite, por lo que el backend no puede revertirlo.
func Ref(salt, value string) string {
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(salt + ":" + value))
	return hex.EncodeToString(sum[:])
}

// NewID genera un event_id aleatorio local (clave de deduplicación).
func NewID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
