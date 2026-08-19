package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ P-005 · EL COMANDO `permea project join` ══════════════════════════════════════════════
//
// `permea project join` es el PRIMER comando de dos niveles del binario: hasta 005 los dos
// subcomandos —`enroll` y `status`— son planos. La gramática la fija
// `specs/005-adhesion-a-proyecto/contracts/cli.md` §La gramática.
//
// ═══ DOS CAPAS, Y LA INYECCIÓN NO ES COMODIDAD (D-005-P7) ══════════════════════════════════
//
// Es el patrón que `enroll.go` ya resolvió —`runEnroll` sucia, `enroll` pura con dependencias
// inyectadas—, y aquí se sigue **sin variantes**:
//
//	runProjectOS   capa SUCIA: resuelve stdin, su naturaleza pipe/TTY, los dos canales y el ejecutor real
//	runProject     despacho del segundo nivel, ya con todo inyectado
//	projectJoin    capa PURA: entrada, los tres rehúses y el camino hasta emitir
//
// **Por qué inyectado y no resuelto dentro**: el camino completo termina en una petición HTTPS contra
// un destino de prueba con certificado autofirmado, y `cmd/permea/main_test.go`
// —`TestRetirada_EnrollNoParaYLimpiaLaClave`— ya dejó escrito que **un proceso hijo no confiaría en
// ese certificado** sin montarle un almacén. La inyección del ejecutor es **la única vía** de probar
// el camino completo sin arrancar procesos.
//
// ═══ LA PRESENTACIÓN — P-005 T027, y lo que ya no le tocaba hacer ══════════════════════════
//
// El **reparto de canales** y los **ocho códigos** los fijó T021, porque T016 exige stdout no vacío
// en el éxito y T017/T020 exigen `ExitCode() == 1` exacto. Lo que T027 añade es **la redacción**:
// `mensajeDeUnion` y `mensajeDeRehuseRemoto`, con las cuatro promesas de `contracts/cli.md`
// §Comportamiento —D3 ≡ D4 byte a byte, D2 sin nombrar el Proyecto ajeno, D1 sin indicar la causa, y
// el no verificable dejando volver a intentarlo (P-005 FR-013a)—.

// Los DOS códigos de salida del binario (`contracts/cli.md` §Los códigos de salida). **No hay un
// tercero**: los ocho desenlaces del comando caben en dos valores porque la distinción que importa
// —qué pasó— viaja en el mensaje, no en el número.
//
// Aquí vivió el código de andamiaje de P-005 T003 —el `70` que el contrato no usa, elegido para que
// los tests no pudieran acertar contra un verbo sin implementar—, y se retira con esta tarea: P-005
// T017 y T020 exigen `ExitCode() == 1` EXACTO, y cualquier superviviente los tumba. *(P-005 T023, que
// compara los ocho códigos, queda de respaldo — no de responsable.)* La retirada se comprueba por
// barrido, no leyendo: el identificador que lo declaraba no aparece en ningún fichero del repositorio.
const (
	codigoExito = 0
	codigoFallo = 1
)

// ejecutorDeAdhesion emite la petición de adhesión y devuelve la denominación del Proyecto.
//
// Es la costura de la capa pura, y su firma es deliberadamente la del contrato y no la del cliente:
// **destino ya derivado** y **token**, no un `*transport.Client`. Así el test inyecta una función y
// no tiene que fabricar un cliente para observar el camino.
type ejecutorDeAdhesion func(destino, token, codigo, projectRef string) (denominacion string, err error)

// adherirPorRed es el ejecutor real: un solo intento contra el destino derivado, sin cola ni
// reintento (P-005 FR-018). Es el gemelo de `defaultVerify` de `enroll.go`, y por el mismo motivo:
// la capa sucia es el único sitio que conoce el transporte de verdad.
func adherirPorRed(destino, token, codigo, projectRef string) (string, error) {
	return transport.New(destino, token).Adherir(codigo, projectRef)
}

// runProjectOS es la CAPA SUCIA que usa `main`: resuelve stdin y su naturaleza, los dos canales
// reales y el ejecutor de red. No decide nada.
func runProjectOS(args []string) int {
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: no se pudo inspeccionar la entrada estándar:", err)
		return codigoFallo
	}
	// Un pipe/fichero NO tiene el bit ModeCharDevice; una TTY interactiva sí. Es el mismo juicio que
	// hace `runEnroll`, y sostiene la misma garantía: NUNCA un prompt que se cuelgue.
	stdinEsPipe := fi.Mode()&os.ModeCharDevice == 0
	return runProject(args, os.Stdin, stdinEsPipe, os.Stdout, os.Stderr, adherirPorRed)
}

// runProject despacha el SEGUNDO nivel de `permea project <verbo>`.
//
// Devuelve el código de salida en vez de llamar a `os.Exit`: es lo que permite probarlo en proceso,
// sin arrancar un binario hijo. `main` es quien sale.
func runProject(args []string, stdin io.Reader, stdinEsPipe bool, stdout, stderr io.Writer, adherir ejecutorDeAdhesion) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "error: falta el verbo. Verbos disponibles: join")
		return codigoFallo
	}

	switch args[0] {
	case "join":
		return projectJoin(args[1:], stdin, stdinEsPipe, stdout, stderr, adherir)
	default:
		// NUNCA se intenta interpretar ni corregir el verbo (`contracts/cli.md` §La gramática): un
		// comando que adivina lo que quisiste decir ejecuta lo que no pediste.
		fmt.Fprintf(stderr, "error: verbo desconocido %q. Verbos disponibles: join\n", args[0])
		return codigoFallo
	}
}

// projectJoin es la CAPA PURA del verbo: entrada, los tres rehúses locales y el camino hasta emitir.
//
// ═══ EL ORDEN, Y ES CONTRATO (D-005-P13, `contracts/cli.md` §Comportamiento) ═══════════════
//
//	R1  árbol de trabajo con raíz reconocible
//	R2  enrolamiento
//	R3  forma de la configuración
//
// De más específico a menos, y de más barato a más caro —que aquí coinciden—. **Los tres, antes de
// emitir nada** (SC-004): fuera de un árbol el número de peticiones es exactamente cero.
//
// ═══ LA ENTRADA VA ANTES QUE LOS TRES, Y SE DECIDE AQUÍ ════════════════════════════════════
//
// El error de uso **no es uno de los tres rehúses** —`contracts/cli.md` lo deja fuera de la tabla a
// propósito: «en los tres, el comando no llegó a intentar la adhesión»—, así que D-005-P13 no lo
// ordena. Se resuelve primero porque responde a **qué se ha pedido**, y los tres rehúses responden a
// **si puede hacerse**: preguntar lo segundo sin saber lo primero es responder a una pregunta que
// nadie ha terminado de formular. Es también el orden que sigue `enroll`.
func projectJoin(args []string, stdin io.Reader, stdinEsPipe bool, stdout, stderr io.Writer, adherir ejecutorDeAdhesion) int {
	codigo, err := leerCodigoDeAdhesion(args, stdin, stdinEsPipe)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return codigoFallo
	}

	// ── R1 · ÁRBOL DE TRABAJO (P-005 FR-006, FR-007) ──────────────────────────────────────
	//
	// Es el único de los tres que **no depende de nada instalado**: se responde con el directorio
	// actual y punto. Ni lee configuración, ni necesita enrolamiento.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, "error: no se pudo determinar el directorio actual:", err)
		return codigoFallo
	}
	// ⛔ EL SALT NO INFLUYE EN `huboRaiz`, y por eso aquí va vacío. `DerivarConRaiz` obtiene ese
	// booleano de `ascender`, que sólo mira el sistema de ficheros; el salt sólo entra en el hash de
	// la identidad. Preguntarlo sin salt tiene una consecuencia que sí importa: **el camino de rehúse
	// no llega a tocar el secreto local**, y `LoadOrCreateSalt` lo CREARÍA si no existiera.
	if _, hayRaiz := project.DerivarConRaiz(cwd, ""); !hayRaiz {
		fmt.Fprintln(stderr, "error: este directorio no pertenece a un árbol de trabajo con raíz reconocible.\n"+
			"       Ejecuta el comando dentro del árbol de trabajo que quieres agrupar")
		return codigoFallo
	}

	dir, err := config.DataDir()
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return codigoFallo
	}
	cfg, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		// ⛔ UNA CONFIGURACIÓN QUE NO SE PUEDE LEER CAE EN R3, Y NO ROMPE EL ORDEN.
		//
		// El orden decide **qué rehúse se reporta cuando varios se cumplen**, y aquí R2 ni siquiera es
		// evaluable: sin poder leer el fichero **no se sabe** si hay enrolamiento, y afirmar que no lo
		// hay sería afirmar lo que no se ha podido establecer. Lo que sí se sabe es lo que R3 dice —la
		// configuración no permite determinar el destino—, así que es el que se reporta.
		//
		// La causa se conserva: es de `encoding/json` sobre el fichero del usuario, no sobre nada que
		// lleve el código dentro (P-005 FR-020).
		fmt.Fprintln(stderr, "error: la configuración local no permite determinar el destino:", err)
		return codigoFallo
	}

	// ── R2 · ENROLAMIENTO (P-005 FR-008) ──────────────────────────────────────────────────
	//
	// ⛔ **NO se usa `config.IsEnrolled`, y no es un descuido.** Esa función funde tres hechos —hay
	// endpoint, hay token y el esquema es admisible—, y el tercero **no es este rehúse**: un endpoint
	// en claro lo rechaza la guarda del transporte, que es «la misma frontera y sin exención»
	// (P-005 FR-017). `contracts/cli.md` §Notas lo dice de este flujo en concreto —«aquí el endpoint
	// viene de la configuración persistida, así que la guarda del transporte SÍ se ejercita»—, y
	// tratarlo como «no enrolado» la dejaría **inalcanzable**, convirtiendo esa nota en falsa.
	//
	// Lo que este rehúse mira es **lo que `enroll` escribe**: endpoint y token. Sin uno de los dos no
	// hay enrolamiento, y el mensaje dice cómo conseguirlo.
	if cfg.Endpoint == "" || cfg.DeviceToken == "" {
		fmt.Fprintln(stderr, "error: esta instalación no está enrolada.\n"+
			"       Enrólala primero:  permea enroll <enrollment-string>\n"+
			"       (recomendado por stdin:  … | permea enroll -)")
		return codigoFallo
	}

	// ── R3 · FORMA DE LA CONFIGURACIÓN (P-005 FR-009) ─────────────────────────────────────
	//
	// El destino de la adhesión **se deriva** del de la ingesta (D-005-P3): si el endpoint guardado no
	// tiene la forma que permite derivarlo, se rehúsa en vez de conjeturar. El error de
	// `config.DerivarEndpointDeAdhesion` ya nombra **la forma** de lo hallado y NUNCA lo hallado
	// (P-005 FR-020 manda sobre FR-009), así que se propaga tal cual.
	destino, err := config.DerivarEndpointDeAdhesion(cfg.Endpoint)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return codigoFallo
	}

	// ── LA IDENTIDAD, POR LA MISMA DERIVACIÓN QUE LA INGESTA (P-005 FR-004, FR-005) ────────
	//
	// El salt se resuelve **aquí y no antes**: es el mismo que estampa la ingesta, y ninguno de los
	// tres rehúses lo necesita. `LoadOrCreateSalt` lo crea si no existe, y eso es lo correcto y no un
	// efecto colateral: una instalación sin salt **todavía no tiene identidad**, y la que se cree
	// ahora es exactamente la que estampará la primera emisión. Derivar con un salt distinto —o
	// vacío— presentaría una identidad que la ingesta nunca va a usar: el fallo que P-005 FR-005
	// existe para impedir.
	salt, err := config.LoadOrCreateSalt(dir)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return codigoFallo
	}
	// MISMA función que el camino de la ingesta, no una equivalente: `project.Derivar` y
	// `project.DerivarConRaiz` comparten cuerpo por construcción.
	identidad, _ := project.DerivarConRaiz(cwd, salt)

	denominacion, err := adherir(destino, cfg.DeviceToken, codigo, identidad)
	if err != nil {
		fmt.Fprintln(stderr, mensajeDeRehuseRemoto(err))
		return codigoFallo
	}

	// El éxito comunica **la denominación del Proyecto** (P-005 FR-002) por **stdout**, que es la
	// respuesta (P-005 FR-021).
	fmt.Fprintln(stdout, mensajeDeUnion(denominacion))
	return codigoExito
}

// mensajeDeUnion es la salida de **D4 y de D3 a la vez**, y ésa es la garantía.
//
// ═══ D3 ≡ D4 ES ESTRUCTURAL, NO UNA COINCIDENCIA QUE HAYA QUE VIGILAR ══════════════════════
//
// P-005 FR-010 exige que unirse ahora y estar ya unido a ese mismo Proyecto sean **indistinguibles en
// texto, canal y resultado del proceso**. Aquí no hay dos ramas que mantener sincronizadas: **el
// comando nunca llega a saber cuál de los dos fue** —la plataforma responde `200` con la misma forma
// en los dos casos (`contracts/adhesion.md` §Los cuatro desenlaces)—, así que sólo existe **un
// camino**. Introducir una diferencia observable exigiría **inventar antes una distinción que no
// existe**, y eso es más difícil que acordarse de no hacerlo.
//
// ⛔ **`%q` NO ES DECORACIÓN.** La denominación viene **del servidor**, y `%q` escapa los caracteres
// de control: sin él, un nombre de Proyecto con un salto de línea o una secuencia de escape ANSI
// podría **fabricar líneas de salida** que la persona leería como del comando. Es entrada ajena
// impresa en una terminal.
func mensajeDeUnion(denominacion string) string {
	return fmt.Sprintf("unido al Proyecto %q", denominacion)
}

// mensajeDeRehuseRemoto redacta los desenlaces que decide la plataforma, según `contracts/cli.md`
// §Comportamiento y salidas.
//
// ⛔ **Discrimina por CENTINELA, nunca por el texto del error**: los centinelas son el contrato de
// `internal/transport` (`ErrCodigoNoUtilizable`, `ErrIdentidadYaAsignada`, `ErrNoVerificable`), y
// casar cadenas ataría esta redacción a la de otro paquete.
//
// ⛔ **Ningún mensaje reproduce el código de adhesión** (P-005 FR-020, SC-005): no lo recibe, y por
// eso no puede filtrarlo ni por descuido.
func mensajeDeRehuseRemoto(err error) string {
	switch {
	case errors.Is(err, transport.ErrCodigoNoUtilizable):
		// P-005 FR-012 — dice QUE no es utilizable y **NUNCA la causa**. Las cinco causas del `422`
		// —inexistente, de otra organización, revocado, prefijo desconocido, `project_ref` no
		// conforme— llegan aquí **como el mismo centinela**, así que producen el mismo mensaje **por
		// construcción**: no hay dónde escribir la diferencia. Deducirla convertiría al comando en un
		// oráculo para averiguar qué códigos existen.
		return "error: el código de adhesión no es utilizable.\n" +
			"       Pide uno nuevo a quien administre la organización"

	case errors.Is(err, transport.ErrIdentidadYaAsignada):
		// P-005 FR-011 — lo dice y **NUNCA nombra el Proyecto ajeno**. Tampoco puede: la plataforma no
		// lo revela y el comando no infiere lo que ella calla.
		return "error: este árbol de trabajo ya pertenece a otro Proyecto.\n" +
			"       Habla con quien administre la organización si debería cambiar"

	case errors.Is(err, transport.ErrNoVerificable):
		// P-005 FR-013 + FR-013a — informa de que **no se pudo establecer** el desenlace, **no afirma
		// ninguno**, y deja volver a intentarlo.
		//
		// **La segunda frase no es consuelo: es el requisito.** El estado remoto queda INDETERMINADO
		// —si la petición llegó y se perdió la respuesta, la unión ocurrió—, y lo que la spec exige es
		// que esa incertidumbre sea **inocua**. Callarlo dejaría a la persona sin saber si repetir es
		// seguro, y la respuesta es que sí: el código no se agota al usarse y unirse dos veces es
		// indistinguible de unirse una (P-005 FR-010).
		//
		// La causa sí viaja, y es el único desenlace donde procede: red, TLS y DNS son **accionables**,
		// mientras que la causa de un rechazo no lo es y además revela. *(Residuo conocido: la causa
		// puede llevar la URL dentro, y `url.Error` no redacta el `userinfo`. Es la deuda que
		// `tasks.md` §Lo que este plan de tareas NO hace ya tiene anotada, con sus otros seis sitios;
		// no se amplía aquí ni se estrena aquí.)*
		return "error: no se pudo establecer el desenlace de la operación.\n" +
			"       Puede que la unión se haya producido y puede que no: desde aquí no se distingue.\n" +
			"       Vuelve a intentarlo cuando quieras — el código no se agota al usarse y unirse dos\n" +
			"       veces es indistinguible de unirse una, así que repetir no duplica nada.\n" +
			"       causa: " + err.Error()

	default:
		// ⛔ NO ES EL `default` QUE SOBRA DE UN `switch`. Aquí caen los fallos de **configuración** que
		// el transporte detecta al ir a emitir —el endpoint en claro (`transport.ErrScheme`), el no
		// analizable—, y **no son desenlaces de la adhesión**: no los decide la plataforma. Se
		// presentan con su causa, que es lo que la persona tiene que corregir.
		return "error: " + err.Error()
	}
}

// leerCodigoDeAdhesion resuelve el código por las DOS vías equivalentes de `contracts/cli.md`
// §Entrada: argumento posicional si lo hay y no es `-`; entrada estándar si es `-` o si no hay
// argumento y stdin es un pipe. **Sin argumento y con stdin no-pipe → error de uso, sin leer**: un
// comando que se cuelga esperando entrada que nadie va a teclear es peor que uno que falla.
//
// El valor se recorta y **NUNCA se hace eco** (P-005 FR-020): ni aquí ni en ningún mensaje de error,
// que por eso no reproducen el argumento.
//
// ═══ POR QUÉ NO REUTILIZA `readEnrollmentInput`, QUE HACE ESTO MISMO ═══════════════════════
//
// D-005-P7 dice que esa función es reutilizable **en su lógica**, y es exactamente lo que se ha
// hecho: mismo reparto de vías, mismo trato de la TTY, mismo recorte. Lo que **no** se comparte es el
// desenlace —el mensaje de uso nombra otro comando y otra vía recomendada—, y ese mensaje es contrato
// de 003 para `enroll` y de 005 para esto. Fundirlas obligaría a parametrizar el texto de un contrato
// ajeno para ahorrar seis líneas.
func leerCodigoDeAdhesion(args []string, stdin io.Reader, stdinEsPipe bool) (string, error) {
	if len(args) >= 1 && args[0] != "-" {
		return strings.TrimSpace(args[0]), nil
	}
	forzarStdin := len(args) >= 1 && args[0] == "-"
	if !forzarStdin && !stdinEsPipe {
		return "", fmt.Errorf("uso: permea project join <código>  (recomendado: pásalo por stdin, p. ej. `… | permea project join -`)")
	}
	b, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el código de adhesión de la entrada estándar: %w", err)
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return "", fmt.Errorf("no se recibió ningún código de adhesión por la entrada estándar")
	}
	return s, nil
}
