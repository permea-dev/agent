package main

import (
	"bufio"
	"encoding/json"
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

// ═══ P-011 B5 · `--scan` DE UNA SESIÓN DE COPILOT CLI ═══════════════════════════════════════════════════════════════════

// nanoAIUPorDolar: un crédito son 10⁹ nanoAIU y vale 0,01 USD (D-4), así que 1 USD son 10¹¹ nanoAIU.
const nanoAIUPorDolar = 100_000_000_000

// esSesionCopilot dice si la primera línea del fichero es el `session.start` de una sesión de Copilot CLI (P-9): con
// `data.copilotVersion` textual. Se lee con el mismo tope de 1 MiB que `esSesionCodex`, y se mira TRAS Gemini.
func esSesionCopilot(path string) (bool, error) {
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
		Type string `json:"type"`
		Data struct {
			CopilotVersion *string `json:"copilotVersion"`
		} `json:"data"`
	}
	if json.Unmarshal(sc.Bytes(), &primera) != nil {
		return false, nil // no es JSON, o no tiene esta forma: lo trata el camino de Claude Code, como hoy
	}
	return primera.Type == "session.start" && primera.Data.CopilotVersion != nil, nil
}

// dryRunCopilot imprime los eventos de una sesión de Copilot CLI sin tocar estado ni cola (P-011 FR-022). Aplica las MISMAS
// reglas que la emisión (`LeerFicheroCopilot`), sobre un estado sólo en memoria que nunca se guarda. La línea `evento:` es la
// compartida, y el resumen, el de Copilot CLI. Tras él, la línea de control (D-4, D-011-P6): la Σ de los créditos de los
// registros que dieron evento y su equivalente en USD, EXACTO con 11 decimales, en enteros. Nunca se envía.
func dryRunCopilot(path string) error {
	pc := ingest.NuevaPasadaCopilot()
	ctx := ingest.ContextoCopilot{Context: ingest.Context{Salt: "dry-run-salt", MachineID: "local", DevID: "dev-local",
		OrgID: "org-local", AgentVersion: version}, ConCreditos: true}
	evs, err := ingest.LeerFicheroCopilot(state.New(), path, ctx, pc, os.Stderr)
	if err != nil {
		return err
	}
	imprimirEventos(evs)
	fmt.Fprintln(os.Stderr, pc.Resumen())
	fmt.Printf("copilot_cli (control, no se envía): créditos según Copilot %d nanoAIU · equivalente $%d.%011d a 0,01 por crédito\n",
		pc.NanoAIU, pc.NanoAIU/nanoAIUPorDolar, pc.NanoAIU%nanoAIUPorDolar)
	return nil
}
