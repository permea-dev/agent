// Comando agente: por ahora, dry-run que imprime eventos de frontera desde un JSONL.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/tu-org/agente/internal/config"
	"github.com/tu-org/agente/internal/ingest"
)

var version = "0.0.1-dev"

func main() {
	scan := flag.String("scan", "", "ruta a un JSONL de Claude Code para dry-run (imprime eventos, no envía)")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "[PRODUCTO] agente %s\n", version)

	if *scan == "" {
		fmt.Fprintf(os.Stderr, "sin --scan: nada que hacer. config por defecto: %+v\n", config.Default())
		return
	}

	f, err := os.Open(*scan)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer f.Close()

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
		fmt.Printf("evento: tool=%s model=%s in=%d out=%d cost=$%.4f project_ref=%s\n",
			ev.Tool, ev.Model, ev.TokensInput, ev.TokensOutput, ev.CostUSD, ref)
	}
	fmt.Fprintf(os.Stderr, "%d eventos generados (dry-run, nada transmitido)\n", n)
}
