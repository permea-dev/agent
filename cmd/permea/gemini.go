package main

import (
	"fmt"
	"os"

	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/state"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ P-009 B5 · GEMINI CLI EN `--run` Y EN EL DEMONIO ════════════════════════════════════════════
//
// `generate()` lee Gemini DESPUÉS de Codex y ANTES del único `st.Save` (FR-024, plan D-009-P7). La raíz se resolvió en
// `setup()`; si `<raíz>/tmp` existe se mira en CADA pasada, sin resolver enlaces (FR-002, N-11), porque el demonio vive
// días y Gemini CLI puede instalarse después.

// generarGemini lee la raíz de Gemini CLI y ENCOLA sus eventos (P-009 B5). Los errores son POR FICHERO (FR-025, plan
// D-009-P10): un fichero que no se puede leer se omite con un aviso, su estado queda como estaba —la pasada siguiente lo
// relee— y se sigue con el siguiente, sin impedir guardar el estado de los demás. Encolar sigue siendo fatal, como hoy.
func (a *agent) generarGemini(st *state.Store, ictx ingest.Context) (int, error) {
	pg := ingest.NuevaPasadaGemini()
	a.gemini = pg
	sesiones, anteriores, err := ingest.ListarGemini(a.geminiRaiz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gemini: fichero omitido: %v\n", err)
		return 0, nil
	}
	base := ingest.ContextoGemini{Context: ictx}
	total := 0
	for _, f := range sesiones {
		evs, err := ingest.LeerFicheroGemini(st, f, base, pg, os.Stderr) // FR-020: el aviso de línea corrupta, a stderr
		if err != nil {
			fmt.Fprintf(os.Stderr, "gemini: fichero omitido: %v\n", err)
			continue
		}
		for _, ev := range evs {
			if err := transport.Append(a.dir, ev); err != nil {
				return total, err
			}
			total++
		}
	}
	for _, ruta := range anteriores {
		if err := ingest.ContarAnteriorGemini(st, ruta, pg); err != nil {
			fmt.Fprintf(os.Stderr, "gemini: fichero omitido: %v\n", err)
		}
	}
	return total, nil
}
