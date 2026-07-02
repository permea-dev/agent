package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

// LoadOrCreateSalt devuelve el salt persistido en `dir/salt`, generándolo la
// primera vez (32 bytes aleatorios en hex, permisos 0600). El salt es la semilla
// de event.Ref: NUNCA se transmite ni forma parte del Event (Principio I, R6).
func LoadOrCreateSalt(dir string) (string, error) {
	return loadOrCreateSecret(filepath.Join(dir, "salt"), 32)
}

// LoadOrCreateMachineID devuelve un id de instalación aleatorio persistido en
// `dir/machine_id` (R7): estable entre sesiones para agregación, pero no sensible.
// Solo su hash salado (machine_ref) cruza la frontera; el id crudo se queda en local.
func LoadOrCreateMachineID(dir string) (string, error) {
	return loadOrCreateSecret(filepath.Join(dir, "machine_id"), 16)
}

func loadOrCreateSecret(path string, n int) (string, error) {
	b, err := os.ReadFile(path)
	if err == nil {
		return string(b), nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	val := hex.EncodeToString(raw)
	if err := atomicWrite(path, []byte(val), 0o600); err != nil {
		return "", err
	}
	return val, nil
}
