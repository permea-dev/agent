package config

import (
	"os"
	"path/filepath"
)

// OpenCodeDatos devuelve dónde están las bases de OpenCode (P-010 P-8, FR-001), con la misma regla en los tres SO:
//   - `OPENCODE_DB` absoluta: `unica`, y sólo se lee esa base;
//   - si no, `dir` = `$XDG_DATA_HOME/opencode` si no es vacía, y si no, `<directorio personal>/.local/share/opencode`,
//     con el mismo `os.UserHomeDir()` que las otras raíces. Nunca `AppData` ni `Library`.
//
// `OPENCODE_DB` vacía, `":memory:"` o relativa cae en la regla general: `filepath.IsAbs` es falso para las tres.
//
// Devuelve la ruta AUNQUE NO EXISTA: si existe se comprueba en cada pasada (P-10, FR-002), porque el demonio vive días y
// OpenCode puede instalarse después.
//
// ⚠️ Son dos de las CINCO lecturas de entorno de producción, con `CODEX_HOME`, `GEMINI_CLI_HOME` y `COPILOT_HOME` (M-5;
// P-011 M-5). Los tests las neutralizan en `internal/testutil.Sandbox` (M-3).
func OpenCodeDatos() (dir, unica string, err error) {
	db := os.Getenv("OPENCODE_DB")
	if filepath.IsAbs(db) {
		return "", db, nil
	}
	if datos := os.Getenv("XDG_DATA_HOME"); datos != "" {
		return filepath.Join(datos, "opencode"), "", nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	return filepath.Join(home, ".local", "share", "opencode"), "", nil
}
