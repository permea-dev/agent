// Comando agente Permea.
//
//	--scan <fichero>  dry-run: imprime eventos de frontera desde un JSONL, sin tocar estado ni cola.
//	--run             una pasada incremental: escanea ~/.claude/projects, ingiere y ENCOLA
//	                  los eventos de forma durable en queue.jsonl (la transmisión es US2).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bfgnet/agente_permea/internal/config"
	"github.com/bfgnet/agente_permea/internal/ingest"
	"github.com/bfgnet/agente_permea/internal/state"
	"github.com/bfgnet/agente_permea/internal/transport"
)

var version = "0.0.1-dev"

func main() {
	scan := flag.String("scan", "", "ruta a un JSONL de Claude Code para dry-run (imprime eventos, no envía)")
	run := flag.Bool("run", false, "ejecuta una pasada incremental: escanea, ingiere y encola en queue.jsonl (no transmite en este MVP)")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "Permea %s\n", version)

	switch {
	case *scan != "":
		if err := dryRun(*scan); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case *run:
		if err := runOnce(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "sin --scan ni --run: nada que hacer. config por defecto: %+v\n", config.Default())
	}
}

// runOnce ejecuta una pasada de generación incremental (US1): descubre los logs de
// Claude Code, lee solo las líneas nuevas por offset, construye el Event de frontera
// y lo ENCOLA de forma durable. El estado se persiste tras encolar (orden de
// durabilidad de R4). No transmite: el drenaje de la cola corresponde a US2.
func runOnce() error {
	dir, err := config.DataDir()
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		return err
	}
	salt, err := config.LoadOrCreateSalt(dir)
	if err != nil {
		return err
	}
	machineID, err := config.LoadOrCreateMachineID(dir)
	if err != nil {
		return err
	}
	root, err := config.ClaudeCodeLogsRoot(cfg)
	if err != nil {
		return err
	}
	logs, err := state.FindLogs(root)
	if err != nil {
		return err
	}

	statePath := filepath.Join(dir, "state.json")
	st, err := state.Load(statePath)
	if err != nil {
		return err
	}

	ctx := ingest.Context{
		Salt:         salt,
		MachineID:    machineID,
		DevID:        cfg.DevID,
		OrgID:        cfg.OrgID,
		AgentVersion: version,
	}

	total := 0
	for _, logPath := range logs {
		err := st.ScanFile(logPath, func(line []byte) error {
			ev, err := ingest.FromClaudeCodeLine(line, ctx)
			if err != nil {
				fmt.Fprintln(os.Stderr, "skip (línea corrupta):", err)
				return nil // un registro corrupto se omite sin detener el resto
			}
			if ev == nil {
				return nil // línea no facturable (p. ej. mensaje de usuario)
			}
			if err := transport.Append(dir, *ev); err != nil {
				return err
			}
			total++
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Persistir el estado DESPUÉS de encolar: si hay caída entre medias, a lo sumo se
	// re-encola (at-least-once), nunca se pierde un evento (FR-007).
	if err := st.Save(statePath); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%d eventos encolados en %s (no transmitido: sync es US2)\n", total, transport.QueuePath(dir))
	return nil
}

// dryRun imprime los eventos de frontera de un JSONL sin tocar estado ni cola.
func dryRun(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }() // solo lectura: el error de Close no afecta a datos

	ctx := ingest.Context{Salt: "dry-run-salt", MachineID: "local", DevID: "dev-local", OrgID: "org-local", AgentVersion: version}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	n := 0
	for sc.Scan() {
		ev, err := ingest.FromClaudeCodeLine(sc.Bytes(), ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "skip:", err)
			continue
		}
		if ev == nil {
			continue
		}
		n++
		ref := ev.ProjectRef
		if len(ref) > 8 {
			ref = ref[:8] + "…"
		}
		fmt.Printf("evento: tool=%s model=%s in=%d out=%d cost=$%.4f cost_avail=%t project_ref=%s\n",
			ev.Tool, ev.Model, ev.TokensInput, ev.TokensOutput, ev.CostUSD, ev.CostAvailable, ref)
	}
	if err := sc.Err(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d eventos generados (dry-run, nada transmitido)\n", n)
	return nil
}
