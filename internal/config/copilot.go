package config

import (
	"os"
	"path/filepath"
)

// CopilotSessionsRoot devuelve la carpeta de las sesiones de Copilot CLI (P-011 D-5, FR-001): `$COPILOT_HOME/session-state`
// si `COPILOT_HOME` está definida y no es vacía, y si no, `<directorio personal>/.copilot/session-state`, con el mismo
// `os.UserHomeDir()` que las otras raíces. La variable sustituye a `~/.copilot` ENTERO: `session-state` cuelga directamente
// de ella, sin `.copilot` en medio (descubrimiento, muestra 2, M3). Vacía cuenta como no definida.
//
// Devuelve la ruta AUNQUE NO EXISTA: si existe se comprueba en cada pasada (P-6, FR-002), porque el demonio vive días y
// Copilot CLI puede instalarse después.
//
// ⚠️ Es una de las SEIS lecturas de entorno de producción, con `CODEX_HOME`, `GEMINI_CLI_HOME`, `XDG_DATA_HOME`, `OPENCODE_DB`
// y `KIMI_CODE_HOME` (P-011 M-5; P-012 M-5). Los tests la neutralizan en `internal/testutil.Sandbox` (M-3).
func CopilotSessionsRoot() (string, error) {
	if propia := os.Getenv("COPILOT_HOME"); propia != "" {
		return filepath.Join(propia, "session-state"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".copilot", "session-state"), nil
}
