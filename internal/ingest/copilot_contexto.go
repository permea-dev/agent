package ingest

import (
	"os"
	"path/filepath"
	"strings"
)

// ═══ P-011 · LOS FICHEROS DE COPILOT CLI ════════════════════════════════════════════════════════════════════════════
//
// Copilot CLI guarda cada sesión en `<raíz>/<carpeta>/events.jsonl` (descubrimiento Q1). Junto a él hay ficheros que NO
// son consumo y que el lector no abre nunca (P-7, D-6): `workspace.yaml`, `checkpoints/` (el nombre de sus ficheros sale
// del contenido de la conversación), `rewind-file-snapshots/`, `files/`, `research/`, `session.db` y los `.lock`, tanto
// `.session-operation-locks/` como `.workspace-fork.lock`. Una carpeta sin `events.jsonl` (un intento fallido) no es un
// error ni se avisa.

// ListarCopilot devuelve los `events.jsonl` de las sesiones bajo `raiz`, en orden léxico (P-7, FR-003): un nivel, sin
// recursión, y sin las carpetas que empiezan por `.`. Sólo lee el directorio: no abre ningún fichero.
func ListarCopilot(raiz string) ([]string, error) {
	entradas, err := os.ReadDir(raiz) // ordenadas por nombre
	if err != nil {
		return nil, err
	}
	var rutas []string
	for _, e := range entradas {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		ruta := filepath.Join(raiz, e.Name(), "events.jsonl")
		if info, err := os.Stat(ruta); err == nil && info.Mode().IsRegular() {
			rutas = append(rutas, ruta)
		}
	}
	return rutas, nil
}
