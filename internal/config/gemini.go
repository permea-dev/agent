package config

import (
	"os"
	"path/filepath"
)

// GeminiRoot devuelve la raíz `.gemini` de Gemini CLI (P-009 FR-001, M-7): `$GEMINI_CLI_HOME/.gemini` si
// `GEMINI_CLI_HOME` está definida y no es vacía, y si no, `<directorio personal>/.gemini`, con el mismo
// `os.UserHomeDir()` que las raíces de Claude Code y Codex. Es la regla de la CLI (descubrimiento Q1, `homedir` en
// `core/src/utils/paths.ts`, 0.63.0): la variable sustituye al directorio PERSONAL, no a la carpeta, y vacía cuenta
// como no definida. Las sesiones están bajo `<raíz>/tmp`.
//
// Devuelve la ruta AUNQUE NO EXISTA: si `<raíz>/tmp` existe se comprueba en cada pasada (FR-002), porque el demonio
// vive días y Gemini CLI puede instalarse después.
//
// ⚠️ Es una de las DOS lecturas de entorno de producción, con `CODEX_HOME` (plan D-008-P7, D-009-P9). Los tests la
// neutralizan en `internal/testutil.Sandbox` (M-9).
func GeminiRoot() (string, error) {
	if propia := os.Getenv("GEMINI_CLI_HOME"); propia != "" {
		return filepath.Join(propia, ".gemini"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gemini"), nil
}
