# Permea — medidor de coste de IA (local-first, multi-herramienta)

Agente local que lee los logs de uso de herramientas de IA (Claude Code, Codex CLI, Gemini
CLI, OpenCode, Copilot CLI y Kimi Code CLI), calcula **en local** el coste de Claude Code y
transmite al backend de equipo **únicamente metadato derivado** — nunca contenido. Funciona sin
conexión: los eventos pendientes se persisten y se transmiten **exactamente una vez** al
recuperarse la red.

## Garantía de frontera
`internal/event/event.go` define un struct CERRADO: el único dato que puede salir de
la máquina (allowlist de 17 campos, equivalente a `additionalProperties: false`).
`internal/ingest` mapea explícitamente solo los campos permitidos; lo que no se mapea,
se descarta (deny-by-default) — incluso campos **nuevos o desconocidos** de futuras
versiones del log. Los identificadores sensibles (ruta de proyecto, sesión, máquina)
solo cruzan como **hash salado**; el `salt` vive en local y nunca se transmite. Los tests
`internal/ingest/boundary_test.go` y `internal/event` (`TestEvent_OnlyAllowlistKeys`) lo
verifican sobre contenido sensible inyectado a propósito.

## Instalación

Binario estático único, sin dependencias previas. Un comando por sistema operativo; la
integridad se verifica por SHA256 en todos los canales. La versión se inyecta desde la
etiqueta de la release.

**macOS** — Homebrew cask (tap propio):

    brew install --cask permea-dev/permea/permea

**macOS y Linux** — script de instalación (canal **principal en Linux**; el cask de Homebrew
es solo macOS):

    curl -fsSL https://raw.githubusercontent.com/permea-dev/agent/main/install.sh | sh
    # opcional: PERMEA_VERSION=v0.4.0 PREFIX="$HOME/.local/bin" sh install.sh

**Windows** — Scoop (bucket propio):

    scoop bucket add permea https://github.com/permea-dev/scoop-permea
    scoop install permea

**Actualizar**, por canal:

    brew upgrade --cask permea          # macOS, Homebrew
    scoop update; scoop update permea   # Windows, Scoop (primero el bucket, luego permea)
    curl -fsSL https://raw.githubusercontent.com/permea-dev/agent/main/install.sh | sh
                                        # script: relanzarlo instala la última release encima
                                        # (con el mismo PREFIX, si indicaste uno)

Verifica la instalación con `permea --version`. Las versiones publicadas están en
<https://github.com/permea-dev/agent/releases>. Detalle de canales e integridad en
[`specs/002-distribucion/contracts/install-contract.md`](specs/002-distribucion/contracts/install-contract.md).
Compilar desde fuente: ver [Desarrollo](#desarrollo).

## Primeros pasos

Los mismos tres pasos que muestra `permea help`:

**1. `permea enroll`** — conecta este ordenador con tu organización. Tres formas:

- Pega el comando que te da la aplicación al añadir un agente.
- Recomendado, por stdin, para que el secreto no quede en el historial del shell (bash/zsh):

      echo "$ENROLL" | permea enroll -

- Recomendado, por stdin (PowerShell): en la aplicación, copia **sólo el código** con «Copiar», y:

      Get-Clipboard | permea enroll -

**2. `permea status`** — comprueba que ha quedado conectado. Nunca muestra el token.

**3. Medir y enviar.**

> ⚠️ **La primera pasada envía todo el historial que conserven Claude Code, Codex CLI, Gemini CLI,
> OpenCode, Copilot CLI y Kimi Code CLI**, no sólo lo que uses a partir de ahora. Las siguientes
> envían sólo lo nuevo.

    permea --run       # mide y envía una vez
    permea --daemon    # o lo deja en marcha: mide y envía cada cierto tiempo

Ayuda en cualquier momento: `permea help`, y la de cada subcomando con `permea <subcomando> -h`.

## Comandos

Tres subcomandos, y el orden en que aparecen es el orden en que se usan:

    permea enroll [<enrollment-string>]   empareja la instalación con su backend
    permea status                         informa si la instalación está enrolada, y contra qué
    permea project join [<código>]        une este árbol de trabajo a un Proyecto

`enroll` y `project join` exigen **HTTPS**, sin exención ni modo de desarrollo: es la misma
frontera que la emisión de eventos. `status` no contacta con nadie.

### `permea enroll` — emparejar la instalación con su backend

    echo "$ENROLL" | permea enroll -      # recomendada
    permea enroll <enrollment-string>     # equivalente, pero ver el aviso

El *enrollment string* lo emite quien administra la organización. Verifica el token contra el
backend y **solo si lo acepta** guarda endpoint y token en `config.json`. Un enrolamiento
rechazado **no escribe nada**: el estado queda idéntico al de no haberlo intentado.

> **La vía de entrada estándar es la recomendada**, y la razón es concreta: pasado **por
> argumento**, el valor queda en el **historial del intérprete de órdenes** y a la vista de quien
> pueda enumerar procesos. El comando no controla eso; lo que sí garantiza es que **existe una vía
> que no obliga a ponerlo en la línea de órdenes**. Por stdin **nunca se hace eco**.

### `permea status` — diagnóstico local

    permea status

Dice si la instalación está enrolada y contra qué backend. Es **local**: no contacta con nadie.
**Nunca imprime el token**, a lo sumo un indicador de presencia.

### `permea project join` — unir este árbol de trabajo a un Proyecto

    cd ~/dev/mi-proyecto                  # DENTRO del árbol que se quiere agrupar
    echo "$CODIGO" | permea project join -   # recomendada
    permea project join <código>             # equivalente, mismo aviso que en `enroll`

Une la instalación a un **Proyecto** del panel para que su consumo cuente bajo él — **incluido el
que ya estaba medido**, sin reenviar ni reprocesar un solo evento: la agrupación ocurre **en el
servidor**, sobre lo que ya llegó.

- **El código lo acuña quien administra la organización**, desde el panel del backend contra el que
  te enrolaste. No se genera en local, y no hay forma de fabricarlo desde aquí.
- **Se ejecuta dentro del árbol de trabajo** que se quiere agrupar: si el directorio actual no
  pertenece a un árbol con raíz reconocible, el comando **rehúsa y no emite ninguna petición**.
- **Al completarse, dice el nombre del Proyecto** al que ha quedado unida la instalación. Es la
  confirmación: sale por la salida estándar, y los errores por la de error.
- **Repetirlo no tiene ninguna consecuencia.** El código no se agota al usarse, y unirse dos veces
  es indistinguible de unirse una — así que si una ejecución no llega a completarse, **volver a
  intentarlo es seguro y no duplica nada**. La segunda vez verás **exactamente la misma salida** que
  la primera: el comando **no revela cuál de las dos surtió efecto**, y es a propósito.
- **Si el servidor no responde, lo dice sin afirmar nada.** No dará por hecho que la unión ocurrió
  ni que no ocurrió: desde aquí las dos cosas se ven igual, y el comando **no conjetura lo que no
  puede establecer**. Es **el único desenlace tras el que alguien querría repetir**, así que es
  justo aquí donde se cobra la promesa de arriba: **vuelve a intentarlo cuando quieras**.
- **No persiste nada en local**: ni el código, ni el Proyecto, ni el hecho de haberse unido. El
  efecto vive en el servidor, y los ficheros del directorio de datos quedan igual que estaban.
- La operación es **de un solo intento**: transmite y espera. Nunca queda en la cola de envío
  diferido, ni siquiera con el servidor inalcanzable.

## Modos de ejecución

    permea --scan <fichero>         # prueba en seco: un evento por mensaje (por respuesta en Codex y Gemini; por llamada en una base de OpenCode, en una sesión de Copilot CLI o en un `wire.jsonl` de Kimi Code CLI), sin tocar estado ni cola
    permea --run                    # una pasada: escanea, encola y drena al backend
    permea --daemon                 # bucle continuo: cada sync_interval genera y transmite

- **Un evento por mensaje, contado entero.** Claude Code escribe varias líneas por mensaje, y a veces
  —en las conversaciones de subagentes— la salida crece de una línea a la siguiente. El agente emite
  **uno** por mensaje, con un `event_id` derivado del propio mensaje (el mismo en cualquier pasada o
  instalación), y **cada partida vale lo más alto que alcanza** entre sus líneas; nunca se suman. La
  plataforma descarta los repetidos. Las líneas `<synthetic>` no se emiten.
- **Un mensaje sale cuando está completo.** El último de cada conversación se envía cuando empieza el
  siguiente o tras **10 minutos sin cambios** en la conversación (y, como tope, a las 24 horas de su
  última línea). Lo que sigue abierto no se guarda aparte: el agente lo relee del log en la pasada
  siguiente.
- **Al final de cada pasada, dos líneas por stderr con sólo recuentos**: lo leído y lo emitido; y los
  mensajes que crecieron entre líneas, los que esperan a cerrarse, las líneas releídas, las tardías y
  las que no traen el desglose de la caché. Con Codex activo, una tercera línea, `codex: …`, con sus
  recuentos, y con Gemini activo, otra, `gemini: …`. Con OpenCode activo, una línea más,
  `opencode: …`; con Copilot CLI activo, otra, `copilot_cli: …`, y con Kimi Code CLI activo, otra,
  `kimi_code: …`.
- **`--scan`** imprime por evento las cuatro partidas de tokens (`in=`, `out=`, `cw=`, `cr=`), el
  desglose de la escritura de caché por duración (`cw5m=`, `cw1h=`), el coste y el `event_id`. Como
  lee un fichero completo, cierra todos sus mensajes al final. En una sesión de Codex, de Gemini, de
  OpenCode, de Copilot CLI o de Kimi Code CLI la línea no lleva `cw5m=` ni `cw1h=`, y el coste es 0.
  `--scan` admite también una base de OpenCode, y añade una línea de contraste con el coste según
  OpenCode, que no se envía. Con una sesión de Copilot CLI añade una línea de control con sus
  créditos y su equivalente en USD, que tampoco se envía. Con un `wire.jsonl` de Kimi Code CLI no
  hay línea de control: sólo sus eventos y su resumen.
- **`--run`** hace una pasada: descubre los logs de Claude Code, las sesiones de Codex, de
  Gemini, de Copilot CLI y de Kimi Code CLI y las bases de OpenCode, lee lo nuevo por offset,
  encola de forma durable en `queue.jsonl` lo que está cerrado y, si hay `endpoint` configurado,
  drena la cola por HTTPS autenticado. Lo que sigue abierto se queda para la pasada siguiente, y lo
  avisa: «N mensajes siguen abiertos: se enviarán en la próxima pasada».
- **`--daemon`** repite lo anterior cada `sync_interval`; un mensaje en espera se cierra a su hora
  aunque su fichero no crezca, y el resumen sólo se escribe en los ciclos con novedades. Errores de
  red/5xx se reintentan con backoff acotado (máx. 5 reintentos, tope 5 min) y el lote permanece en
  cola; un error de autenticación (401/403) detiene el sync por configuración errónea. `Ctrl-C` para
  parar.
- Sin `endpoint` configurado, la medición local funciona igual: los eventos quedan en la
  cola y nada se transmite.
- **Windows, PowerShell 5.1**: al redirigir la salida a un fichero (`2>`, `>`), las tildes pueden
  verse mal. Es la codificación de PowerShell, y no afecta a lo que se mide ni a lo que se envía.

### Codex CLI

Si existe `~/.codex/sessions` (o `$CODEX_HOME/sessions`), el agente lee también el consumo de Codex CLI.
Cada respuesta del modelo es un evento con `tool = codex`, sus tokens y su modelo. El coste no lo calcula
el agente: los eventos de Codex salen con `cost_available = false`, y el coste lo pone la plataforma
cuando tenga las tarifas de esos modelos. Hace falta Codex 0.153.0 o posterior; las sesiones anteriores
se cuentan como «formato anterior» y no se envían. Si la carpeta no existe, no cambia nada.

### Gemini CLI

Si existe `~/.gemini/tmp` (o `$GEMINI_CLI_HOME/.gemini/tmp`), el agente lee también el consumo de Gemini CLI.
Cada respuesta del modelo es un evento con `tool = gemini`, sus tokens y su modelo; el razonamiento cuenta
como salida. El coste no lo calcula el agente: los eventos salen con `cost_available = false`. Hace falta
Gemini CLI 0.39.0 o posterior; las sesiones anteriores se cuentan como «formato anterior» y no se envían.
Las llamadas internas de la CLI (compresión, enrutado) y los intentos fallidos no quedan en sus sesiones,
así que el agente no los ve. Gemini CLI borra por defecto las sesiones de más de 30 días. Si la carpeta no
existe, no cambia nada. Antigravity no guarda su consumo en estas sesiones y el agente no lo lee.

### OpenCode

Si existe `~/.local/share/opencode` (o `$XDG_DATA_HOME/opencode`), el agente lee también el consumo
de OpenCode: su base `opencode.db` y las de sus canales (`opencode-<canal>.db`), con la misma regla
en Linux, macOS y Windows. Con `OPENCODE_DB` apuntando a una base por ruta absoluta, lee sólo esa.
La base se abre sólo para leer. Cada llamada al modelo es un evento con `tool = opencode`, sus
tokens y su modelo; el razonamiento cuenta como salida. El coste no lo calcula el agente: los
eventos salen con `cost_available = false`, aunque OpenCode guarde el suyo. Las partidas se han
comprobado con modelos de Google y de Anthropic; con otros proveedores, un modelo cuyo nombre no
case con el catálogo de la plataforma llega sin coste. Hace falta OpenCode 2.x; las bases de la 1.x
se cuentan como «formato anterior» y no se envían. La llamada que pone título a cada sesión no queda
entre sus mensajes y el agente no la ve: en la medición fue un 0,37 % del coste. Una llamada en
curso se espera hasta 24 horas; pasado ese tiempo se cuenta como «abandonada» y no se envía. Si la
carpeta no existe, no cambia nada.

### Copilot CLI

Si existe `~/.copilot/session-state` (o `$COPILOT_HOME/session-state`), el agente lee también el
consumo de Copilot CLI: el `events.jsonl` de cada sesión. La carpeta sólo se lee: el agente no crea,
cambia ni borra nada en ella, tampoco sus `.lock`. Cada llamada al modelo principal es un evento con
`tool = copilot_cli`, sus tokens y su modelo, y cada compactación con éxito (`/compact`), otro. El
modelo es el de cada llamada, el que se cobra, también en un subagente. El razonamiento ya va dentro
de la salida y no se suma otra vez. El coste no lo calcula el agente: los eventos salen con
`cost_available = false`, y los créditos de Copilot no se envían. Las llamadas auxiliares de la CLI
(modelos pequeños, a 0 créditos en el plan Free) no se cuentan. Si se mata una sesión, lo ya escrito
se envía, pero la llamada en curso no está en disco y no se cuenta. `permea --scan` sobre el
`events.jsonl` de una sesión imprime sus eventos, el resumen y una línea de control con los créditos
según Copilot y su equivalente en USD, a 0,01 por crédito, que no se envía. Medido sólo con Copilot
CLI 1.0.97, en Windows y en el plan Free; los subagentes en segundo plano, los *workflows*, la
compactación automática y los planes de pago, sin muestra. Con un plan de pago, un modelo con punto
en el nombre (como `claude-haiku-4.5`) no casa con el catálogo de la plataforma y llega sin coste.
Copilot dentro de VS Code o JetBrains no se lee. Si la carpeta no existe, no cambia nada.

### Kimi Code CLI

Si existe `~/.kimi-code/sessions` (o `$KIMI_CODE_HOME/sessions`), el agente lee también el consumo
de Kimi Code CLI: el `wire.jsonl` de cada agente de cada sesión, el principal y los de sus
subagentes, y `session_index.jsonl`, de donde saca el proyecto. Nada más: ni `config.toml` ni
`credentials/`. La carpeta sólo se lee: el agente no crea, cambia ni borra nada en ella. Cada
petición terminada es un evento con `tool = kimi_code`, sus tokens y su modelo, y cada compactación
(`/compact`), otro. El modelo es el alias sin su proveedor: `moonshot-ai/kimi-k2.7-code` llega como
`kimi-k2.7-code`. Lo que un `/fork` copia de la sesión de origen no se cuenta otra vez. El coste no
lo calcula el agente: los eventos salen con `cost_available = false`, y lo pone la plataforma con su
tarifa. La tiene para cuatro modelos desde el 2026-10-15 (`permea-platform` · `6e7cedd`):
`kimi-k2.7-code`, `kimi-k2.7-code-highspeed`, `kimi-k2.6` y `kimi-k3`; otro modelo llega sin coste
hasta que tenga su fila. Si `session_index.jsonl` no se puede leer, el agente avisa
(«kimi_code: fichero omitido: …») y deja las sesiones para la pasada siguiente. Si se mata una
sesión, lo ya escrito se envía, pero la petición en curso, que Kimi sí cobra, no está en disco y no
se cuenta; tampoco los intentos fallidos o interrumpidos con Esc. Si una pasada cae justo entre la
copia de un `/fork` y su marca, lo copiado sale como propio: es cuestión de milisegundos. En
Windows, la misma carpeta de trabajo escrita con otras mayúsculas da otro proyecto. `permea --scan`
sobre un `wire.jsonl` imprime sus eventos y el resumen, sin línea de control. Medido sólo con Kimi
Code 2.2.0, en Windows y con clave de API de Kimi Platform; la membresía (OAuth) y los otros
proveedores, sin muestra. Si la carpeta no existe, no cambia nada.

## Coste y tarifas

El coste de Claude Code se calcula **en local**, en **USD**, con una tabla empaquetada en el binario
(`internal/pricing`): **17 modelos con cinco cifras** cada uno —entrada, salida, escritura de caché a
5 minutos, escritura de caché a 1 hora y lectura de caché—, espejo exacto de las 17 filas de
Anthropic del catálogo de tarifas de la plataforma (`permea-dev/permea-platform` ·
`backend/config/pricing.php` · `8f147d1`); las de otros proveedores sólo están en la plataforma.

- **Casamiento exacto**: la tarifa se busca por el identificador de modelo tal como llega en el log.
  No se normalizan sufijos de fecha, prefijos ni mayúsculas. Un modelo sin fila sale con
  `cost_available=false` y coste 0, y sus tokens se cuentan igual.
- **La escritura de caché, por su duración**: la de 5 minutos a su tarifa y la de 1 hora a la suya (el
  doble de la entrada). El log trae el desglose; el evento sigue llevando sólo el total.
- **Hipótesis P-1**: una línea **sin el desglose** de la caché, o con uno que no suma el total, se
  tarifa entera a **1 hora**. Es la misma hipótesis que declara el catálogo de la plataforma.
- **Limitación 2**: el **«modo rápido»** no se distingue; un evento en modo rápido quedaría
  infravalorado.
- **Codex, Gemini, OpenCode, Copilot CLI y Kimi Code CLI**: el agente no calcula su coste; lo pone la plataforma (ver
  «Codex CLI», «Gemini CLI», «OpenCode», «Copilot CLI» y «Kimi Code CLI»).

## Configuración y rutas por SO

Toda la configuración y el estado viven en el **directorio de datos por SO** (creado al
primer arranque), resuelto vía `os.UserConfigDir` — nunca se hardcodean rutas:

| SO | Directorio de datos |
|---|---|
| Linux | `$XDG_CONFIG_HOME/permea` (o `~/.config/permea`) |
| macOS | `~/Library/Application Support/permea` |
| Windows | `%AppData%\permea` |

Ahí se guardan `config.json` (endpoint, token, identidad, `sync_interval`),
`state.json` (offset de escaneo), `queue.jsonl` (cola offline) y `salt` (secreto local,
`0600`, nunca transmitido). Los logs de Claude Code se resuelven en `~/.claude/projects`
por SO, con override opcional `logs_root` en la config. Escrituras siempre atómicas
(temporal + `os.Rename`).

## Desarrollo

Para quien trabaja en el código del agente (no hace falta para usarlo):

    make test    # suite completa; empieza por el test de frontera
    make run     # dry-run: imprime los eventos del fixture, sin transmitir
    make build   # binario en bin/permea
    make lint    # golangci-lint (puerta de calidad: 0 avisos)

### Portabilidad
Binario estático único, **sin CGO ni dependencias externas** (solo stdlib). Compila para
Linux, macOS y Windows:

    CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build ./cmd/permea
    CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build ./cmd/permea
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./cmd/permea

La versión del binario (`agent_version` en el evento) se inyecta con
`-ldflags "-X main.version=<versión>"`.

### Estructura
    cmd/permea        punto de entrada (subcomandos enroll/status/project join + modos scan/run/daemon)
    internal/event    LA FRONTERA (struct cerrado del evento)
    internal/ingest   lectores por herramienta (claude_code, codex, gemini, opencode, copilot, kimi) + tests de frontera
    internal/sqlitero lector propio de SQLite, de sólo lectura (bases de OpenCode)
    internal/pricing  cálculo de coste local (tabla empaquetada)
    internal/state    escaneo incremental idempotente
    internal/transport cliente HTTPS + cola offline + entrega exactamente-una-vez
    internal/config   configuración local, rutas por SO, salt e identidades
