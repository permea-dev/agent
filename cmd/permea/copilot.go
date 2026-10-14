package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/state"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ P-011 B4 · COPILOT CLI EN `--run` Y EN EL DEMONIO ════════════════════════════════════════════════════════════
//
// `generate()` lee Copilot CLI DESPUÉS de OpenCode y ANTES del único `st.Save` (FR-024). La raíz se resolvió en `setup()`; si
// existe se mira en CADA pasada (P-6, FR-002), porque el demonio vive días y Copilot CLI puede instalarse después.

// copilotActivo dice si Copilot CLI se lee EN ESTA PASADA (P-6): si existe la carpeta de sus sesiones.
func (a *agent) copilotActivo() bool {
	if a.copilotRaiz == "" {
		return false
	}
	info, err := os.Stat(a.copilotRaiz)
	return err == nil && info.IsDir()
}

// generarCopilot lee las sesiones de Copilot CLI y ENCOLA sus eventos (P-011 B4). Los errores son POR FICHERO (FR-026): un
// `events.jsonl` que no se puede leer se omite con un aviso SIN SU RUTA (D-011-P5: la ruta lleva el id de la sesión y, en
// Windows, el nombre del usuario), su entrada de `state.json` queda como estaba —la pasada siguiente lo relee— y se sigue con
// el siguiente, sin impedir guardar el estado de los demás. Encolar sigue siendo fatal, como hoy.
func (a *agent) generarCopilot(st *state.Store, ictx ingest.Context) (int, error) {
	pc := ingest.NuevaPasadaCopilot()
	a.copilot = pc
	sesiones, err := ingest.ListarCopilot(a.copilotRaiz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "copilot_cli: fichero omitido: %v\n", sinRuta(err))
		return 0, nil
	}
	base := ingest.ContextoCopilot{Context: ictx}
	total := 0
	for _, ruta := range sesiones {
		evs, err := ingest.LeerFicheroCopilot(st, ruta, base, pc, os.Stderr) // el aviso de línea corrupta, a stderr
		if err != nil {
			fmt.Fprintf(os.Stderr, "copilot_cli: fichero omitido: %v\n", sinRuta(err))
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

// sinRuta quita la ruta de un error de E/S (D-011-P5): de `open <ruta>: permission denied` queda `open: permission denied`.
func sinRuta(err error) error {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return fmt.Errorf("%s: %w", pe.Op, pe.Err)
	}
	return err
}
