package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/state"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ P-012 B4 · KIMI CODE EN `--run` Y EN EL DEMONIO ══════════════════════════════════════════════════════════════════
//
// `generate()` lee Kimi Code DESPUÉS de Copilot CLI y ANTES del único `st.Save` (FR-024). La carpeta se resolvió en `setup()`; si
// existe `<carpeta>/sessions` se mira en CADA pasada (P-2, FR-002), porque el demonio vive días y Kimi Code puede instalarse después.

// kimiActivo dice si Kimi Code se lee EN ESTA PASADA (P-2): si `<carpeta>/sessions` es un directorio.
func (a *agent) kimiActivo() bool {
	if a.kimiRaiz == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(a.kimiRaiz, "sessions"))
	return err == nil && info.IsDir()
}

// generarKimi lee las sesiones de Kimi Code y ENCOLA sus eventos (P-012 B4). Primero el índice, una vez por pasada (D-012-P6): si
// existe y no se puede leer, se avisa y NO se lee ningún `wire.jsonl` en esta pasada (P-13), porque un evento emitido sin proyecto
// ya no se corrige; sus offsets no avanzan y la pasada siguiente los lee. Los errores son POR FICHERO (FR-026, D-012-P5): un
// `wire.jsonl` que no se puede leer se omite con un aviso SIN SU RUTA (la ruta lleva el id de la sesión y, en Windows, el nombre del
// usuario), su entrada de `state.json` queda como estaba y se sigue con el siguiente. Encolar sigue siendo fatal, como hoy.
func (a *agent) generarKimi(st *state.Store, ictx ingest.Context) (int, error) {
	pk := ingest.NuevaPasadaKimi()
	a.kimi = pk
	indice, err := ingest.LeerIndiceKimi(a.kimiRaiz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kimi_code: fichero omitido: %v\n", sinRuta(err))
		return 0, nil
	}
	ficheros, err := ingest.ListarKimi(a.kimiRaiz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kimi_code: fichero omitido: %v\n", sinRuta(err))
		return 0, nil
	}
	base := ingest.ContextoKimi{Context: ictx, Indice: indice}
	total := 0
	for _, ruta := range ficheros {
		evs, err := ingest.LeerFicheroKimi(st, ruta, base, pk, os.Stderr) // el aviso de línea corrupta, a stderr
		if err != nil {
			fmt.Fprintf(os.Stderr, "kimi_code: fichero omitido: %v\n", sinRuta(err))
			continue
		}
		for _, ev := range evs {
			if err := transport.Append(a.dir, ev); err != nil {
				return total, err
			}
			total++
		}
	}
	return total, nil
}
