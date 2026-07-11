// Comando agente Permea.
//
//	--scan <fichero>  dry-run: imprime eventos de frontera desde un JSONL, sin tocar estado ni cola.
//	--run             una pasada: escanea ~/.claude/projects, ENCOLA de forma durable en
//	                  queue.jsonl y drena la cola al backend por HTTPS (US1 + US2).
//	--daemon          bucle continuo: cada sync_interval genera y transmite (US2).
//	--version         imprime la versión (inyectada desde la etiqueta) en stdout y termina.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/state"
	"github.com/permea-dev/agent/internal/transport"
)

// version es la versión del binario. GoReleaser la sobreescribe desde la etiqueta con
// -ldflags "-X main.version={{.Version}}"; por defecto, un valor de desarrollo.
var version = "0.0.1-dev"

// printVersion escribe SOLO la versión (una línea, sin adornos) en w. Es el contrato de
// `--version`: salida limpia y estable, apta para scripts y para verificar que el binario
// publicado coincide con su etiqueta (SC-002).
func printVersion(w io.Writer) {
	// Escritura de una sola línea; un fallo al escribir la versión no es recuperable ni
	// afecta a la durabilidad (a diferencia de la cola), así que se ignora explícitamente.
	_, _ = fmt.Fprintln(w, version)
}

func main() {
	// Subcomandos (P-003): se despachan ANTES del parseo de flags para no interferir con
	// los flags de P-001/P-002 (--scan/--run/--daemon/--version), que se conservan intactos.
	if len(os.Args) >= 2 && os.Args[1] == "enroll" {
		if err := runEnroll(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	scan := flag.String("scan", "", "ruta a un JSONL de Claude Code para dry-run (imprime eventos, no envía)")
	run := flag.Bool("run", false, "una pasada: escanea, encola en queue.jsonl y drena al backend (US1 + US2)")
	daemon := flag.Bool("daemon", false, "bucle continuo: cada sync_interval genera y transmite (US2)")
	showVersion := flag.Bool("version", false, "imprime la versión en stdout y termina")
	flag.Parse()

	// --version se atiende ANTES de cualquier otra salida: stdout queda con exactamente la
	// versión y nada más (sin el banner de stderr), para verificación e integración.
	if *showVersion {
		printVersion(os.Stdout)
		return
	}

	fmt.Fprintf(os.Stderr, "Permea %s\n", version)

	switch {
	case *scan != "":
		if err := dryRun(*scan); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case *daemon:
		if err := runDaemon(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case *run:
		if err := runOnce(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "sin --scan/--run/--daemon: nada que hacer. config por defecto: %+v\n", config.Default())
	}
}

// agent agrupa el contexto resuelto una sola vez (directorio de datos, config, salt e
// identidades) para las pasadas de generación y sync.
type agent struct {
	dir  string
	cfg  config.Config
	ictx ingest.Context
}

// setup resuelve el directorio de datos por SO, carga la config y las identidades locales
// (salt/machineID persistidos). El salt nunca cruza la frontera (R6).
func setup() (*agent, error) {
	dir, err := config.DataDir()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		return nil, err
	}
	salt, err := config.LoadOrCreateSalt(dir)
	if err != nil {
		return nil, err
	}
	machineID, err := config.LoadOrCreateMachineID(dir)
	if err != nil {
		return nil, err
	}
	return &agent{
		dir:  dir,
		cfg:  cfg,
		ictx: newIngestContext(version, cfg, salt, machineID),
	}, nil
}

// newIngestContext construye el contexto de ingesta a partir de la versión REAL del
// binario (variable `version`, sobreescribible con -ldflags "-X main.version=...") y de
// la identidad local. Aislado como función pura para poder verificar por test que la
// versión del agente llega hasta Event.AgentVersion (T036), sin tocar el sistema de
// ficheros.
func newIngestContext(agentVersion string, cfg config.Config, salt, machineID string) ingest.Context {
	return ingest.Context{
		Salt:         salt,
		MachineID:    machineID,
		DevID:        cfg.DevID,
		OrgID:        cfg.OrgID,
		AgentVersion: agentVersion,
	}
}

// generate ejecuta una pasada de generación incremental (US1): descubre los logs de
// Claude Code, lee solo las líneas nuevas por offset, construye el Event de frontera y lo
// ENCOLA de forma durable. El estado se persiste DESPUÉS de encolar (durabilidad, R4): si
// hay caída entre medias, a lo sumo se re-encola (at-least-once), nunca se pierde.
func (a *agent) generate() (int, error) {
	root, err := config.ClaudeCodeLogsRoot(a.cfg)
	if err != nil {
		return 0, err
	}
	logs, err := state.FindLogs(root)
	if err != nil {
		return 0, err
	}

	statePath := filepath.Join(a.dir, "state.json")
	st, err := state.Load(statePath)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, logPath := range logs {
		err := st.ScanFile(logPath, func(line []byte) error {
			ev, err := ingest.FromClaudeCodeLine(line, a.ictx)
			if err != nil {
				fmt.Fprintln(os.Stderr, "skip (línea corrupta):", err)
				return nil // un registro corrupto se omite sin detener el resto
			}
			if ev == nil {
				return nil // línea no facturable (p. ej. mensaje de usuario)
			}
			if err := transport.Append(a.dir, *ev); err != nil {
				return err
			}
			total++
			return nil
		})
		if err != nil {
			return total, err
		}
	}

	if err := st.Save(statePath); err != nil {
		return total, err
	}
	return total, nil
}

// sync drena la cola pendiente hacia el backend por HTTPS (US2, T030). Sin endpoint
// configurado es un no-op silencioso (medición local sin backend). Valida que el endpoint
// sea https antes de intentar transmitir (FR-009).
func (a *agent) sync() (int, error) {
	if a.cfg.Endpoint == "" {
		return 0, nil
	}
	if err := a.cfg.Validate(); err != nil {
		return 0, err
	}
	return transport.Drain(a.dir, a.cfg)
}

// runOnce ejecuta una única pasada: genera (US1) y drena (US2).
func runOnce() error {
	a, err := setup()
	if err != nil {
		return err
	}
	n, err := a.generate()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d eventos encolados en %s\n", n, transport.QueuePath(a.dir))

	if a.cfg.Endpoint == "" {
		fmt.Fprintln(os.Stderr, "sync omitido: sin endpoint configurado")
		return nil
	}
	m, err := a.sync()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d eventos transmitidos y confirmados\n", m)
	return nil
}

// runDaemon corre el bucle del agente: cada sync_interval de la config genera y drena
// (T031). Un error de auth (401/403) detiene el bucle por configuración errónea; los
// errores de red/5xx se registran y se reintentan en el siguiente ciclo (la cola persiste).
func runDaemon() error {
	a, err := setup()
	if err != nil {
		return err
	}
	interval, err := time.ParseDuration(a.cfg.SyncInterval)
	if err != nil {
		return fmt.Errorf("sync_interval inválido %q: %w", a.cfg.SyncInterval, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	fmt.Fprintf(os.Stderr, "daemon: ciclo cada %s (Ctrl-C para parar)\n", interval)

	for {
		if err := a.tick(); err != nil {
			return err // solo errores terminales (p. ej. auth) llegan aquí
		}
		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "daemon detenido")
			return nil
		case <-ticker.C:
		}
	}
}

// tick es una iteración del daemon: generar + drenar. Devuelve error solo cuando el sync
// debe detenerse (auth); los fallos transitorios se registran y no abortan el bucle.
func (a *agent) tick() error {
	if n, err := a.generate(); err != nil {
		fmt.Fprintln(os.Stderr, "generación:", err)
	} else if n > 0 {
		fmt.Fprintf(os.Stderr, "%d eventos encolados\n", n)
	}

	if a.cfg.Endpoint == "" {
		return nil
	}
	m, err := a.sync()
	if err != nil {
		if transport.IsAuth(err) {
			return fmt.Errorf("sync detenido por autenticación (revisa device_token): %w", err)
		}
		fmt.Fprintln(os.Stderr, "sync (reintento en el próximo ciclo):", err)
		return nil
	}
	if m > 0 {
		fmt.Fprintf(os.Stderr, "%d eventos transmitidos y confirmados\n", m)
	}
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
