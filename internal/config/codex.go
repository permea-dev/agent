package config

import (
	"os"
	"path/filepath"
)

// CodexSessionsRoot devuelve la raíz de las sesiones de Codex CLI (P-008 FR-001, M-7): `$CODEX_HOME/sessions` si
// `CODEX_HOME` está definida y no es vacía, y si no, `<directorio personal>/.codex/sessions`, con el mismo
// `os.UserHomeDir()` que la raíz de Claude Code. Es la regla de Codex (`utils/home-dir/src/lib.rs:14-16,52-59`,
// `rust-v0.160.1`): una variable vacía cuenta como no definida.
//
// Devuelve la ruta AUNQUE NO EXISTA (P-008 E-2): si existe se comprueba en cada pasada, porque el demonio vive
// días y Codex puede instalarse después (FR-002).
//
// ⚠️ Es la ÚNICA lectura de una variable de entorno en producción (plan D-008-P7). Los tests la neutralizan en
// `internal/testutil.Sandbox` (M-9).
func CodexSessionsRoot() (string, error) {
	if propia := os.Getenv("CODEX_HOME"); propia != "" {
		return filepath.Join(propia, "sessions"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex", "sessions"), nil
}
