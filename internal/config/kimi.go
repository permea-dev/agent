package config

import (
	"os"
	"path/filepath"
)

// KimiHome devuelve la carpeta de datos de Kimi Code (P-012 P-1, FR-001): `$KIMI_CODE_HOME` si está definida y no es vacía, y
// si no, `<directorio personal>/.kimi-code`, con el mismo `os.UserHomeDir()` que las otras raíces. Vacía cuenta como no
// definida, como `COPILOT_HOME` y como la CLI de Kimi (descubrimiento Q1): el núcleo de Kimi la tomaría por una ruta relativa.
// Las sesiones cuelgan de `<carpeta>/sessions`, y el índice es `<carpeta>/session_index.jsonl`.
//
// Devuelve la ruta AUNQUE NO EXISTA: si existe `<carpeta>/sessions` se comprueba en cada pasada (P-2, FR-002), porque el demonio
// vive días y Kimi Code puede instalarse después.
//
// ⚠️ Es una de las SEIS lecturas de entorno de producción, con `CODEX_HOME`, `GEMINI_CLI_HOME`, `XDG_DATA_HOME`, `OPENCODE_DB` y
// `COPILOT_HOME` (P-012 M-5). Los tests la neutralizan en `internal/testutil.Sandbox` (M-3).
func KimiHome() (string, error) {
	if propia := os.Getenv("KIMI_CODE_HOME"); propia != "" {
		return propia, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kimi-code"), nil
}
