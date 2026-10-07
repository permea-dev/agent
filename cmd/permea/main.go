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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/project"
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
	// ═══ P-006 · LA AYUDA SE DECIDE ANTES QUE NADA ════════════════════════════════════════════
	//
	// Sin argumentos, `help`, `-h`, `--help` y `-help` dan la ayuda general por STDOUT, con salida 0,
	// antes del parseo de opciones y antes del banner: pedir ayuda no ejecuta nada ni anuncia nada
	// (`specs/006-medicion-fiel/contracts/cli.md` §La ayuda general). La fuente única del texto está
	// en `ayuda.go`.
	if len(os.Args) < 2 || esPeticionDeAyudaGeneral(os.Args[1]) {
		escribirAyudaGeneral(os.Stdout)
		return
	}

	// Subcomandos (P-003): se despachan ANTES del parseo de flags para no interferir con
	// los flags de P-001/P-002 (--scan/--run/--daemon/--version), que se conservan intactos.
	if os.Args[1] == "enroll" {
		if err := runEnroll(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}
	if os.Args[1] == "status" {
		if err := runStatus(os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}
	// P-005 T003 — `project` es el PRIMER subcomando con verbo. Va en esta misma escalera y **antes
	// de `flag.Parse()`**, por la razón que el comentario de arriba ya declara: los flags de
	// P-001/P-002 se conservan intactos. El segundo nivel lo resuelve `runProject`, no `main`.
	if os.Args[1] == "project" {
		os.Exit(runProjectOS(os.Args[2:]))
	}

	// P-006 FR-022 · SUBCOMANDO INEXISTENTE. Un primer argumento que no es opción ni subcomando es un
	// error de uso, con el mismo código que `project <verbo desconocido>`. Se nombra lo tecleado,
	// salvo que tenga forma de secreto: entonces NUNCA se reproduce (FR-024).
	if !strings.HasPrefix(os.Args[1], "-") {
		if pareceSecreto(os.Args[1]) {
			fmt.Fprintln(os.Stderr, "error: subcomando desconocido (no se reproduce: tiene forma de secreto). "+
				"Subcomandos: enroll, status, project. Ayuda: permea help")
		} else {
			fmt.Fprintf(os.Stderr, "error: subcomando desconocido %q. Subcomandos: enroll, status, project. "+
				"Ayuda: permea help\n", os.Args[1])
		}
		os.Exit(codigoFallo)
	}

	// P-006 · LAS OPCIONES SE PARSEAN CON `ContinueOnError` Y SALIDA DESCARTADA. Con el `FlagSet`
	// por defecto (`ExitOnError`), Go escribe su propio mensaje y su propio uso y sale él mismo. Así
	// se controlan canal y texto (contrato §Opción desconocida): `-h` combinado con otras opciones da
	// la ayuda por stdout con salida 0, y una opción desconocida o sin valor, un mensaje propio por
	// stderr con salida 2 y NADA por stdout.
	opciones := flag.NewFlagSet("permea", flag.ContinueOnError)
	opciones.SetOutput(io.Discard)
	scan := opciones.String("scan", "", "ruta a un JSONL de Claude Code para dry-run (imprime eventos, no envía)")
	run := opciones.Bool("run", false, "una pasada: escanea, encola en queue.jsonl y drena al backend (US1 + US2)")
	daemon := opciones.Bool("daemon", false, "bucle continuo: cada sync_interval genera y transmite (US2)")
	showVersion := opciones.Bool("version", false, "imprime la versión en stdout y termina")
	if err := opciones.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			escribirAyudaGeneral(os.Stdout)
			return
		}
		fmt.Fprintln(os.Stderr, mensajeDeOpcionInvalida(err))
		os.Exit(2)
	}

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
		// Opciones válidas que no piden ningún modo (p. ej. `--run=false`): se conserva lo de siempre,
		// la ayuda por stderr con salida 0, ahora desde la fuente única.
		escribirAyudaGeneral(os.Stderr)
	}
}

// mensajeDeOpcionInvalida traduce el error de `FlagSet.Parse` a un mensaje propio que NOMBRA la
// opción y remite a `permea help` (contrato §Opción desconocida). Si lo tecleado tiene forma de
// secreto, no se reproduce (P-006 FR-024).
func mensajeDeOpcionInvalida(err error) string {
	texto := err.Error()
	if pareceSecreto(texto) {
		return "error: opción no válida (no se reproduce: tiene forma de secreto). Ayuda: permea help"
	}
	const desconocida, sinValor = "flag provided but not defined: ", "flag needs an argument: "
	switch {
	case strings.HasPrefix(texto, desconocida):
		return "error: opción desconocida: " + strings.TrimPrefix(texto, desconocida) + ". Ayuda: permea help"
	case strings.HasPrefix(texto, sinValor):
		return "error: la opción " + strings.TrimPrefix(texto, sinValor) + " necesita un valor. Ayuda: permea help"
	default:
		return "error: opción no válida: " + texto + ". Ayuda: permea help"
	}
}

// agent agrupa el contexto resuelto una sola vez (directorio de datos, config, salt e
// identidades) para las pasadas de generación y sync.
type agent struct {
	dir  string
	cfg  config.Config
	ictx ingest.Context
	// reloj da la hora contra la que se cierran los mensajes por espera (P-007 FR-010). Nil es time.Now; los
	// tests lo fijan por código, nunca por entorno ni por bandera.
	reloj func() time.Time
	// codexRaiz es la RUTA de las sesiones de Codex (P-008 FR-001), resuelta una vez en `setup()`. Si existe se
	// mira en cada pasada (FR-002, E-2). Vacía —un `agent` construido a mano en un test— significa que no se lee
	// Codex (plan D-008-P5).
	codexRaiz string
	// codex son los recuentos de Codex de la última pasada; nil si el lector no estuvo activo (plan D-008-P4).
	codex *ingest.PasadaCodex
}

// setup resuelve el directorio de datos por SO, carga la config y las identidades locales
// (salt/machineID persistidos). El salt nunca cruza la frontera (R6).
func setup() (*agent, error) {
	dir, err := config.DataDir()
	if err != nil {
		return nil, err
	}
	rutaCfg := filepath.Join(dir, "config.json")

	// ═══ P-004 T027 · LA PARADA POR MODO RETIRADO VA ANTES QUE CUALQUIER OTRO FALLO ═══════
	//
	// Va aquí, y no más abajo, por una razón que no es de estilo: si un fallo de salt o de
	// directorio de datos pudiera adelantarse, el usuario con `"plain"` recibiría **un error
	// distinto del suyo** y la parada quedaría indistinguible de cualquier otra avería. La
	// garantía de FR-013 incluye que el usuario sepa **QUÉ** le paró.
	//
	// Solo detiene los caminos que procesan hacia una emisión —`--run`, `--daemon` y el
	// `tick()` del daemon—, porque los tres pasan por `setup()`. `status` y `enroll` NO
	// cuelgan de aquí y siguen operativos a propósito (D-004-5): el primero diagnostica, el
	// segundo repara.
	if err := config.CheckRetiredProjectRefMode(rutaCfg); err != nil {
		return nil, err
	}

	cfg, err := config.Load(rutaCfg)
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
	// P-008 FR-001, FR-002: sin directorio personal no hay raíz de Codex, y el lector queda inactivo sin error.
	codexRaiz, err := config.CodexSessionsRoot()
	if err != nil {
		codexRaiz = ""
	}
	return &agent{
		dir:       dir,
		cfg:       cfg,
		ictx:      newIngestContext(version, cfg, salt, machineID),
		codexRaiz: codexRaiz,
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
//
// Devuelve también la PASADA (P-006), para que quien llama escriba su resumen: cuántas líneas
// facturables se leyeron, cuántos eventos salieron y cuántas se descartaron y por qué.
func (a *agent) generate() (int, *ingest.Pasada, error) {
	root, err := config.ClaudeCodeLogsRoot(a.cfg)
	if err != nil {
		return 0, nil, err
	}
	logs, err := state.FindLogs(root)
	if err != nil {
		return 0, nil, err
	}

	statePath := filepath.Join(a.dir, "state.json")
	st, err := state.Load(statePath)
	if err != nil {
		return 0, nil, err
	}

	// P-004 T032 · LA CACHÉ VIVE AQUÍ, Y POR ESO SU ÁMBITO ES LA PASADA. Se instancia dentro
	// de `generate()` —no en `setup()`— porque `a.ictx` se construye una vez por PROCESO y en
	// `--daemon` ese proceso vive días: una caché de proceso serviría identidad obsoleta a un
	// directorio que entretanto se hubiera convertido en repositorio. Cada pasada estrena la
	// suya y la abandona al terminar; `tick()` llama a `generate()`, así que el daemon estrena
	// una por ciclo.
	ictx := a.ictx
	ictx.Resolutor = project.NuevoResolutor()
	// P-006 FR-001/FR-033 · LA PASADA, POR EL MISMO MOTIVO Y EN EL MISMO SITIO. Un mensaje cuyas
	// líneas se lean en esta llamada produce UN evento. Una por llamada: en `--daemon`, una por ciclo,
	// y entre ciclos no hay memoria (los repetidos los descarta la plataforma por `event_id`).
	pasada := ingest.NuevaPasada()
	ictx.Pasada = pasada
	// P-007 FR-010: la hora contra la que se cierran los mensajes por espera, una para toda la pasada.
	ahora := time.Now()
	if a.reloj != nil {
		ahora = a.reloj()
	}

	total := 0
	for _, logPath := range logs {
		var cerrados []ingest.Cerrado
		err := st.Recorrer(logPath, func(line []byte, inicio int64, releida bool) error {
			pasada.Situar(inicio, releida)
			ev, err := ingest.FromClaudeCodeLine(line, ictx)
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
		}, func(leido int64, modificado time.Time) int64 {
			// P-007 FR-010, FR-011: salen los mensajes cerrados; lo que sigue abierto no se consume, y el offset
			// se queda en su comienzo para que la pasada siguiente lo relea.
			var abierto int64
			var hayAbierto bool
			cerrados, abierto, hayAbierto = pasada.CerrarConReloj(ahora, modificado)
			if hayAbierto {
				return abierto
			}
			return leido
		})
		if err != nil {
			return total, pasada, err
		}
		// P-007 FR-012: lo cerrado se encola ANTES de guardar el estado. Si algo falla antes de este punto, el
		// estado no avanza y la pasada siguiente relee esas líneas.
		for _, c := range cerrados {
			if err := transport.Append(a.dir, c.Evento); err != nil {
				return total, pasada, err
			}
			total++
		}
	}

	// P-008: Codex, tras Claude Code y ANTES del único `st.Save` (FR-025, plan D-008-P6). La ruta se resolvió en
	// `setup()`; si existe se mira en CADA pasada, porque el demonio vive días y Codex puede instalarse después (FR-002, E-2).
	a.codex = nil
	if a.codexRaiz != "" {
		if info, err := os.Stat(a.codexRaiz); err == nil && info.IsDir() {
			n, err := a.generarCodex(st, ictx)
			total += n
			if err != nil {
				return total, pasada, err
			}
		}
	}

	if err := st.Save(statePath); err != nil {
		return total, pasada, err
	}
	return total, pasada, nil
}

// generarCodex lee la raíz de Codex y ENCOLA sus eventos (P-008 B5). Los errores son POR FICHERO (FR-028, plan
// D-008-P11): un fichero que no se puede leer se omite con un aviso, su estado queda como estaba —la pasada siguiente lo
// relee— y se sigue con el siguiente, sin impedir guardar el estado de los demás. Encolar sigue siendo fatal, como hoy.
func (a *agent) generarCodex(st *state.Store, ictx ingest.Context) (int, error) {
	pc := ingest.NuevaPasadaCodex()
	a.codex = pc
	sesiones, comprimidos, err := ingest.ListarCodex(a.codexRaiz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codex: fichero omitido: %v\n", err)
		return 0, nil
	}
	base := ingest.ContextoCodex{Context: ictx}
	total := 0
	for _, ruta := range sesiones {
		evs, err := ingest.LeerFicheroCodex(st, ruta, base, pc, os.Stderr) // FR-029: el aviso de línea corrupta, a stderr
		if err != nil {
			fmt.Fprintf(os.Stderr, "codex: fichero omitido: %v\n", err)
			continue
		}
		for _, ev := range evs {
			if err := transport.Append(a.dir, ev); err != nil {
				return total, err
			}
			total++
		}
	}
	for _, ruta := range comprimidos {
		if err := ingest.ContarComprimido(st, ruta, pc); err != nil {
			fmt.Fprintf(os.Stderr, "codex: fichero omitido: %v\n", err)
		}
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
	n, pasada, err := a.generate()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d eventos encolados en %s\n", n, transport.QueuePath(a.dir))
	fmt.Fprintln(os.Stderr, pasada.Resumen()) // P-006: sólo recuentos, nunca identificadores
	if aviso := pasada.AvisoDeAbiertos(); aviso != "" {
		fmt.Fprintln(os.Stderr, aviso) // P-007 FR-014: lo abierto sale en la próxima pasada
	}
	if a.codex != nil {
		fmt.Fprintln(os.Stderr, a.codex.Resumen()) // P-008 FR-019: con el lector activo en esta pasada
	}

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
	if n, pasada, err := a.generate(); err != nil {
		fmt.Fprintln(os.Stderr, "generación:", err)
	} else {
		if n > 0 {
			fmt.Fprintf(os.Stderr, "%d eventos encolados\n", n)
		}
		// P-006: el resumen, sólo si la pasada leyó algo. Un resumen vacío cada ciclo es ruido, y el
		// ruido enseña a ignorar los avisos.
		if pasada.HayNovedades() {
			fmt.Fprintln(os.Stderr, pasada.Resumen())
		}
		if a.codex != nil && a.codex.HayNovedades() {
			fmt.Fprintln(os.Stderr, a.codex.Resumen()) // P-008 FR-019: sólo con novedades
		}
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
//
// P-006 FR-010: aplica las MISMAS reglas que la emisión —una pasada por fichero, así que un mensaje
// de varias líneas es UN evento; `<synthetic>` y las líneas sin identificadores no salen—, e imprime
// por evento las cuatro partidas de tokens y el `event_id`. Es el instrumento con el que se mide sobre
// una copia de los logs sin transmitir nada. El `event_id` no depende de la sal y es un hash:
// imprimirlo en local no revela nada que no salga ya por la frontera. Los identificadores del
// proveedor NUNCA se imprimen.
func dryRun(path string) error {
	// P-008 FR-020, P-8: una sesión de Codex se reconoce por su primera línea y se lee como Codex.
	codex, err := esSesionCodex(path)
	if err != nil {
		return err
	}
	if codex {
		return dryRunCodex(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }() // solo lectura: el error de Close no afecta a datos

	pasada := ingest.NuevaPasada()
	ctx := ingest.Context{Salt: "dry-run-salt", MachineID: "local", DevID: "dev-local", OrgID: "org-local", AgentVersion: version, Pasada: pasada}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	n := 0
	imprimir := func(ev event.Event, cw5m, cw1h int) {
		n++
		ref := ev.ProjectRef
		if len(ref) > 8 {
			ref = ref[:8] + "…"
		}
		// P-007 FR-016: el desglose de la escritura de caché, detrás de `cw=`, que sigue siendo el total. Viaja
		// con el mensaje cerrado: el evento no lo lleva (D-1).
		fmt.Printf("evento: tool=%s model=%s in=%d out=%d cw=%d cw5m=%d cw1h=%d cr=%d cost=$%.4f cost_avail=%t project_ref=%s event_id=%s\n",
			ev.Tool, ev.Model, ev.TokensInput, ev.TokensOutput, ev.TokensCacheCreation, cw5m, cw1h, ev.TokensCacheRead,
			ev.CostUSD, ev.CostAvailable, ref, ev.EventID)
	}
	for sc.Scan() {
		ev, err := ingest.FromClaudeCodeLine(sc.Bytes(), ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "skip:", err)
			continue
		}
		if ev != nil {
			// Sólo si la pasada no la acumula, que con un `event_id` de 32 hex es inalcanzable. Su desglose no
			// se conoce aquí, y se imprime a 0.
			imprimir(*ev, 0, 0)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	// P-007 FR-016: el fichero está completo, así que al final se cierra todo lo que tenga.
	for _, c := range pasada.CerrarFichero() {
		imprimir(c.Evento, c.Escritura5m, c.Escritura1h)
	}
	fmt.Fprintf(os.Stderr, "%d eventos generados (dry-run, nada transmitido)\n", n)
	fmt.Fprintln(os.Stderr, pasada.Resumen()) // P-006: sólo recuentos, nunca identificadores
	return nil
}

// esSesionCodex dice si la primera línea del fichero es un `session_meta` de Codex (P-008 P-8): Claude Code no escribe
// ninguno. Se lee con el mismo tope de 1 MiB que el `--scan` de Claude Code, y con el mismo error si se pasa (R-2).
func esSesionCodex(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }() // solo lectura
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	if !sc.Scan() {
		return false, sc.Err()
	}
	var primera struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(sc.Bytes(), &primera) != nil {
		return false, nil // no es JSON: lo trata el camino de Claude Code, como hoy
	}
	return primera.Type == "session_meta", nil
}

// dryRunCodex imprime los eventos de una sesión de Codex sin tocar estado ni cola (P-008 FR-020). Aplica las MISMAS
// reglas que la emisión (`LeerFicheroCodex`), sobre un estado sólo en memoria que nunca se guarda. La línea `evento:` es
// la aprobada (§Textos aprobados): la de Claude Code sin `cw5m=` ni `cw1h=`, que en Codex no existen.
func dryRunCodex(path string) error {
	pc := ingest.NuevaPasadaCodex()
	ctx := ingest.ContextoCodex{Context: ingest.Context{Salt: "dry-run-salt", MachineID: "local", DevID: "dev-local",
		OrgID: "org-local", AgentVersion: version}}
	evs, err := ingest.LeerFicheroCodex(state.New(), path, ctx, pc, os.Stderr)
	if err != nil {
		return err
	}
	for _, ev := range evs {
		ref := ev.ProjectRef
		if len(ref) > 8 {
			ref = ref[:8] + "…"
		}
		fmt.Printf("evento: tool=%s model=%s in=%d out=%d cw=%d cr=%d cost=$%.4f cost_avail=%t project_ref=%s event_id=%s\n",
			ev.Tool, ev.Model, ev.TokensInput, ev.TokensOutput, ev.TokensCacheCreation, ev.TokensCacheRead,
			ev.CostUSD, ev.CostAvailable, ref, ev.EventID)
	}
	fmt.Fprintf(os.Stderr, "%d eventos generados (dry-run, nada transmitido)\n", len(evs))
	fmt.Fprintln(os.Stderr, pc.Resumen()) // P-008 FR-020: el resumen de Codex
	return nil
}
