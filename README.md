# Permea — medidor de coste de IA (local-first, multi-herramienta)

Agente local que lee los logs de uso de herramientas de IA (Claude Code y Codex CLI), calcula
**en local** el coste de Claude Code y transmite al backend de equipo **únicamente metadato
derivado** — nunca contenido. Funciona sin conexión: los eventos pendientes se persisten y se
transmiten **exactamente una vez** al recuperarse la red.

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

> ⚠️ **La primera pasada envía todo el historial que conserven Claude Code y Codex CLI**, no sólo lo
> que uses a partir de ahora. Las siguientes envían sólo lo nuevo.

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

    permea --scan <fichero.jsonl>   # prueba en seco: un evento por mensaje (por respuesta, en Codex), sin tocar estado ni cola
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
  recuentos.
- **`--scan`** imprime por evento las cuatro partidas de tokens (`in=`, `out=`, `cw=`, `cr=`), el
  desglose de la escritura de caché por duración (`cw5m=`, `cw1h=`), el coste y el `event_id`. Como
  lee un fichero completo, cierra todos sus mensajes al final. En una sesión de Codex la línea no lleva
  `cw5m=` ni `cw1h=`, y el coste es 0.
- **`--run`** hace una pasada: descubre los logs de Claude Code y las sesiones de Codex, lee lo nuevo
  por offset, encola de forma durable en `queue.jsonl` lo que está cerrado y, si hay `endpoint`
  configurado, drena la cola por HTTPS autenticado. Lo que sigue abierto se queda para la pasada
  siguiente, y lo avisa: «N mensajes siguen abiertos: se enviarán en la próxima pasada».
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

## Coste y tarifas

El coste de Claude Code se calcula **en local**, en **USD**, con una tabla empaquetada en el binario
(`internal/pricing`): **17 modelos con cinco cifras** cada uno —entrada, salida, escritura de caché a
5 minutos, escritura de caché a 1 hora y lectura de caché—, espejo exacto del catálogo de tarifas de
la plataforma (`permea-dev/permea-platform` · `backend/config/pricing.php` · `8f147d1`).

- **Casamiento exacto**: la tarifa se busca por el identificador de modelo tal como llega en el log.
  No se normalizan sufijos de fecha, prefijos ni mayúsculas. Un modelo sin fila sale con
  `cost_available=false` y coste 0, y sus tokens se cuentan igual.
- **La escritura de caché, por su duración**: la de 5 minutos a su tarifa y la de 1 hora a la suya (el
  doble de la entrada). El log trae el desglose; el evento sigue llevando sólo el total.
- **Hipótesis P-1**: una línea **sin el desglose** de la caché, o con uno que no suma el total, se
  tarifa entera a **1 hora**. Es la misma hipótesis que declara el catálogo de la plataforma.
- **Limitación 2**: el **«modo rápido»** no se distingue; un evento en modo rápido quedaría
  infravalorado.
- **Codex**: el agente no calcula su coste; lo pone la plataforma (ver «Codex CLI»).

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
    internal/ingest   lectores por herramienta (claude_code, codex) + tests de frontera
    internal/pricing  cálculo de coste local (tabla empaquetada)
    internal/state    escaneo incremental idempotente
    internal/transport cliente HTTPS + cola offline + entrega exactamente-una-vez
    internal/config   configuración local, rutas por SO, salt e identidades
