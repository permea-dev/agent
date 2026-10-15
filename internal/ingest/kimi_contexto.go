package ingest

import (
	"os"
	"path/filepath"
	"strings"
)

// ═══ P-012 · LOS FICHEROS DE KIMI CODE ══════════════════════════════════════════════════════════════════════════════
//
// Kimi Code guarda cada agente de cada sesión en `<raíz>/sessions/<wd>/<sesión>/agents/<agente>/wire.jsonl` (descubrimiento Q1;
// muestra 1). Junto a ellos hay ficheros que NO son consumo y que el lector no abre nunca (P-3): `state.json` (título y último
// prompt), `logs/`, `notify/`, `tasks/`, `cron/`, `file-history/`, los `.bak` y `.tmp.*` de la reparación y la escritura atómica
// (desc. Q7), y `sessions/.index-cache/` y `.index-dirty/` (m1). `<raíz>/imported/` son exportaciones con forma de sesión
// (desc. Q6): tampoco se leen. Una sesión sin `agents/` no es un error ni se avisa.

// ListarKimi devuelve los `wire.jsonl` de las sesiones de Kimi Code bajo `raiz`, en orden léxico (P-3, FR-003): exactamente
// `<raiz>/sessions/<a>/<b>/agents/<c>/wire.jsonl`, sin recursión, sin las carpetas que empiezan por `.` en ninguno de los tres
// niveles, y sólo si `wire.jsonl` es un fichero regular. Sólo lee directorios: no abre ningún fichero.
func ListarKimi(raiz string) ([]string, error) {
	sesiones := filepath.Join(raiz, "sessions")
	entradas, err := os.ReadDir(sesiones) // ordenadas por nombre
	if err != nil {
		return nil, err
	}
	var rutas []string
	for _, wd := range visiblesKimi(sesiones, entradas) {
		bs, err := subcarpetasKimi(wd)
		if err != nil {
			return nil, err
		}
		for _, sesion := range bs {
			agentes, err := subcarpetasKimi(filepath.Join(sesion, "agents"))
			if err != nil {
				return nil, err
			}
			for _, agente := range agentes {
				fs, err := os.ReadDir(agente)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return nil, err
				}
				for _, f := range fs {
					if f.Name() == "wire.jsonl" && f.Type().IsRegular() {
						rutas = append(rutas, filepath.Join(agente, f.Name()))
					}
				}
			}
		}
	}
	return rutas, nil
}

// subcarpetasKimi devuelve las subcarpetas visibles de `dir`, en orden. Una carpeta que no existe *(una sesión sin `agents/`, o
// borrada entre dos lecturas)* no tiene ninguna.
func subcarpetasKimi(dir string) ([]string, error) {
	entradas, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return visiblesKimi(dir, entradas), nil
}

// visiblesKimi se queda con los directorios de `entradas` que no empiezan por `.`, como rutas bajo `dir`.
func visiblesKimi(dir string, entradas []os.DirEntry) []string {
	var dirs []string
	for _, e := range entradas {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, filepath.Join(dir, e.Name()))
		}
	}
	return dirs
}
