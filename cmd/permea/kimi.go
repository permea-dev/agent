package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/state"
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
func (a *agent) generarKimi(st *state.Store, ictx ingest.Context, sumar sumidero) (int, error) {
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
			if err := sumar(ev); err != nil {
				return total, err
			}
			total++
		}
	}
	return total, nil
}

// ═══ P-012 B5 · `--scan` DE UN `wire.jsonl` DE KIMI CODE ═════════════════════════════════════════════════════════════

// esSesionKimi dice si la primera línea del fichero es el `metadata` de un `wire.jsonl` de Kimi Code (P-12): con `protocol_version`
// textual y `created_at` numérico. Se lee con el mismo tope de 1 MiB que `esSesionCopilot`, y se mira TRAS Copilot CLI.
func esSesionKimi(path string) (bool, error) {
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
		Type            string   `json:"type"`
		ProtocolVersion *string  `json:"protocol_version"`
		CreatedAt       *float64 `json:"created_at"`
	}
	if json.Unmarshal(sc.Bytes(), &primera) != nil {
		return false, nil // no es JSON, o no tiene esta forma: lo trata el camino de Claude Code, como hoy
	}
	return primera.Type == "metadata" && primera.ProtocolVersion != nil && primera.CreatedAt != nil, nil
}

// dryRunKimi imprime los eventos de un `wire.jsonl` de Kimi Code sin tocar estado ni cola (P-012 FR-022, P-12). Aplica las MISMAS
// reglas que la emisión (`LeerFicheroKimi`, con la frontera `forked`), sobre un estado sólo en memoria que nunca se guarda: S y A salen de
// la ruta *(`<S>/agents/<A>/wire.jsonl`)*, y sin índice el proyecto va vacío. La línea `evento:` es la compartida, y el resumen, el de
// Kimi Code. Sin línea de control: Kimi Code no da créditos (D-5).
func dryRunKimi(path string) error {
	pk := ingest.NuevaPasadaKimi()
	ctx := ingest.ContextoKimi{Context: ingest.Context{Salt: "dry-run-salt", MachineID: "local", DevID: "dev-local",
		OrgID: "org-local", AgentVersion: version}}
	evs, err := ingest.LeerFicheroKimi(state.New(), path, ctx, pk, os.Stderr)
	if err != nil {
		return err
	}
	imprimirEventos(evs)
	fmt.Fprintln(os.Stderr, pk.Resumen())
	return nil
}
