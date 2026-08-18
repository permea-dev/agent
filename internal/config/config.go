// Package config define la configuración local del agente (fichero legible por el usuario)
// y la resolución de rutas por sistema operativo (Principio III: nunca se hardcodean rutas).
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ═══ P-004 T026 · EL MODO EN CLARO SE RETIRÓ; AQUÍ NO QUEDA NADA DE ÉL ════════════════════
//
// Hasta P-004 este paquete declaraba un `ProjectRefMode` con dos valores, `hash` y `plain`.
// El segundo prometía enviar la identidad de proyecto EN CLARO, **nunca se implementó**, y
// ningún camino del código lo consultaba: era una promesa inerte en la configuración de un
// binario de código abierto, a la vista de cualquier auditor (FR-014/FR-015).
//
// La identidad de proyecto cruza SIEMPRE de forma irreversible, sin modo alternativo. Por eso
// el campo no está en `Config`: si estuviera, `Save` lo reescribiría y la clave no se iría
// nunca del fichero del usuario.
//
// Lo que SÍ queda es `CheckRetiredProjectRefMode`, y no es una contradicción: precisamente
// **porque** el campo ya no existe, `encoding/json` descarta la clave EN SILENCIO al cargar
// (deny-by-default del decodificador). Sin una lectura deliberada, quien tuviera `plain` en su
// configuración seguiría creyendo que hace algo. La garantía no sale gratis del borrado: hay
// que construirla.

// Config es el contenido del fichero de configuración local.
type Config struct {
	Endpoint     string   `json:"endpoint"`
	DeviceToken  string   `json:"device_token"`
	OrgID        string   `json:"org_id"`
	DevID        string   `json:"dev_id"`
	Tools        []string `json:"tools"`
	SyncInterval string   `json:"sync_interval"`
	// LogsRoot es un override opcional de la raíz de logs de Claude Code para
	// instalaciones no estándar (R1). Vacío -> se resuelve ~/.claude/projects por SO.
	LogsRoot string `json:"logs_root,omitempty"`
}

// Default devuelve la configuración por defecto: máxima privacidad, Claude Code.
func Default() Config {
	return Config{
		Tools:        []string{"claude_code"},
		SyncInterval: "60s",
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
	// P-005 T005: juicio unificado en `JuzgarEndpoint`; los DOS mensajes son de esta puerta.
	errAnalisis, admisible := JuzgarEndpoint(c.Endpoint)
	if errAnalisis != nil {
		return fmt.Errorf("endpoint inválido %q: %w", c.Endpoint, errAnalisis)
	}
	if !admisible {
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

// ═══ P-004 T025 · LA DETECCIÓN DE LA CLAVE OBSOLETA ════════════════════════════════════════

const (
	// claveRetirada es el ajuste que P-004 retiró de la superficie de configuración.
	claveRetirada = "project_ref_mode"
	// valorRetirado es el ÚNICO valor que detiene el arranque: el que pedía el modo en claro.
	valorRetirado = "plain"
)

// CheckRetiredProjectRefMode devuelve un error si la configuración de `path` SOLICITA el modo
// de envío en claro retirado, y nil en cualquier otro caso.
//
// ═══ POR QUÉ HACE FALTA UNA LECTURA DELIBERADA ════════════════════════════════════════════
//
// Tras retirar el campo del struct, `encoding/json` descarta la clave desconocida EN SILENCIO.
// Un usuario que tuviera `"plain"` cargaría su configuración sin un solo aviso y seguiría
// creyendo que envía rutas en claro. El borrado quita el efecto; **no avisa de la creencia**.
// Por eso esto no reutiliza el `Unmarshal` de `Load`: hace un SEGUNDO decodificado a un mapa
// crudo y mira exactamente UNA clave.
//
// ═══ SOLO `plain` DETIENE, Y ESE ES EL CRITERIO (D-004-1) ═════════════════════════════════
//
// La parada se dispara por **lo que el usuario pedía**, no por la clave que escribió:
//
//	· "plain"          → ERROR. Pidió algo que el producto YA NO PROMETE. Seguir en silencio
//	                     —o con un aviso que puede perderse— le dejaría creyendo que tiene una
//	                     puerta que no existe (FR-013).
//	· "hash"           → nil. Pidió EXACTAMENTE lo que el agente hace siempre. Pararle o
//	                     avisarle sería ruido sobre una petición ya satisfecha (FR-013a).
//	· cualquier otro   → nil. No pide el modo retirado (FR-013b).
//	· clave ausente    → nil.
//
// Los cuatro casos son de `contracts/cli-config.md` §«Comportamiento ante la clave obsoleta».
//
// FALLA HACIA EL LADO QUE NO INTERRUMPE: un fichero inexistente, ilegible o con JSON inválido
// devuelve nil. Esta función responde a UNA pregunta —«¿pide el modo retirado?»— y un fichero
// que no se puede leer no la pide. Diagnosticar una configuración corrupta es trabajo de
// `Load`, y adelantarse aquí convertiría este error en un cajón de sastre donde el mensaje
// dejaría de significar lo que dice.
func CheckRetiredProjectRefMode(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var crudo map[string]json.RawMessage
	if err := json.Unmarshal(b, &crudo); err != nil {
		return nil
	}

	bruto, presente := crudo[claveRetirada]
	if !presente {
		return nil
	}

	var valor string
	if err := json.Unmarshal(bruto, &valor); err != nil {
		return nil // un valor que ni siquiera es una cadena no pide el modo retirado
	}
	if valor != valorRetirado {
		return nil
	}

	// El mensaje es el de contracts/cli-config.md §«Forma del error»: nombra la clave y el
	// valor exactos, da la RUTA REAL del fichero, y dice qué hace el producto ahora — porque
	// quien puso `plain` creía tener otra cosa, y corregir el fichero sin corregir la creencia
	// deja al usuario igual de desinformado. NUNCA vuelca la configuración ni ningún secreto.
	return fmt.Errorf(
		"el modo `%s: %q` fue retirado y ya no existe.\n"+
			"       La identidad de proyecto cruza siempre de forma irreversible.\n"+
			"       Elimina la clave %q de %s para continuar",
		claveRetirada, valorRetirado, claveRetirada, path,
	)
}
