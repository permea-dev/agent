package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

// esSesionGemini dice si la primera línea del fichero es la cabecera de una sesión de Gemini CLI (P-009 FR-022, P-8,
// D-009-P13): un objeto con `sessionId` y `projectHash` textuales y SIN `type`. Las líneas de Claude Code también llevan
// `sessionId`, pero siempre con `type`. Se lee con el mismo tope de 1 MiB que `esSesionCodex`.
func esSesionGemini(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }() // sólo lectura
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	if !sc.Scan() {
		return false, sc.Err()
	}
	var primera struct {
		SessionID   *string         `json:"sessionId"`
		ProjectHash *string         `json:"projectHash"`
		Type        json.RawMessage `json:"type"`
	}
	if json.Unmarshal(sc.Bytes(), &primera) != nil {
		return false, nil // no es JSON: lo trata el camino de Claude Code, como hoy
	}
	return primera.SessionID != nil && primera.ProjectHash != nil && primera.Type == nil, nil
}

// slugDeScan deduce la carpeta `<slug>` de un fichero suelto de `--scan` por su POSICIÓN (DECIDÍ YO, B6): si el padre se
// llama `chats`, es el abuelo; si el abuelo se llama `chats` (subagente), el bisabuelo. Es la única deducción por nombre del
// lector, y sólo vale aquí: en `--run`, el `<slug>` sale de `ListarGemini`. Si no encaja ninguna, "".
func slugDeScan(path string) string {
	padre := filepath.Dir(path)
	switch {
	case filepath.Base(padre) == "chats":
		return filepath.Dir(padre)
	case filepath.Base(filepath.Dir(padre)) == "chats":
		return filepath.Dir(filepath.Dir(padre))
	}
	return ""
}

// dryRunGemini imprime los eventos de una sesión de Gemini CLI sin tocar estado ni cola (P-009 FR-022). Aplica las MISMAS
// reglas que la emisión (`LeerFicheroGemini`), sobre un estado sólo en memoria que nunca se guarda. La línea `evento:` es la
// de Codex (008 §Textos aprobados), y el resumen, el de Gemini.
func dryRunGemini(path string) error {
	pg := ingest.NuevaPasadaGemini()
	ctx := ingest.ContextoGemini{Context: ingest.Context{Salt: "dry-run-salt", MachineID: "local", DevID: "dev-local",
		OrgID: "org-local", AgentVersion: version}}
	slug := slugDeScan(path)
	f := ingest.FicheroGemini{Ruta: path, Slug: slug}
	if slug == "" {
		f.Slug = filepath.Dir(path) // sólo para que la lectura no falle: el proyecto se descarta abajo
	}
	evs, err := ingest.LeerFicheroGemini(state.New(), f, ctx, pg, os.Stderr)
	if err != nil {
		return err
	}
	for _, ev := range evs {
		ref := ev.ProjectRef
		if slug == "" {
			ref = "" // sin la forma `<slug>/chats/…`, no hay proyecto
		}
		if len(ref) > 8 {
			ref = ref[:8] + "…"
		}
		fmt.Printf("evento: tool=%s model=%s in=%d out=%d cw=%d cr=%d cost=$%.4f cost_avail=%t project_ref=%s event_id=%s\n",
			ev.Tool, ev.Model, ev.TokensInput, ev.TokensOutput, ev.TokensCacheCreation, ev.TokensCacheRead,
			ev.CostUSD, ev.CostAvailable, ref, ev.EventID)
	}
	fmt.Fprintf(os.Stderr, "%d eventos generados (dry-run, nada transmitido)\n", len(evs))
	fmt.Fprintln(os.Stderr, pg.Resumen()) // P-009 FR-022: el resumen de Gemini
	return nil
}
