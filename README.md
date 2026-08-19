# Permea — medidor de coste de IA (local-first, multi-herramienta)

Agente local que lee los logs de uso de herramientas de IA, calcula coste **en local**
y transmite al backend de equipo **únicamente metadato derivado** — nunca contenido.
Funciona sin conexión: los eventos pendientes se persisten y se transmiten **exactamente
una vez** al recuperarse la red.

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

    brew install --cask bfgnet/permea/permea

**macOS y Linux** — script de instalación (canal **principal en Linux**; el cask de Homebrew
es solo macOS):

    curl -fsSL https://raw.githubusercontent.com/permea-dev/agent/main/install.sh | sh
    # opcional: PERMEA_VERSION=v1.4.0 PREFIX="$HOME/.local/bin" sh install.sh

**Windows** — Scoop (bucket propio):

    scoop bucket add permea https://github.com/bfgnet/scoop-permea
    scoop install permea

Verifica la instalación con `permea --version`. Detalle de canales e integridad en
[`specs/002-distribucion/contracts/install-contract.md`](specs/002-distribucion/contracts/install-contract.md).
Compilar desde fuente: ver [Portabilidad](#portabilidad).

## Comandos

Tres subcomandos, y el orden en que aparecen es el orden en que se usan:

    permea enroll [<enrollment-string>]   empareja la instalación con su backend
    permea project join [<código>]        une este árbol de trabajo a un Proyecto
    permea status                         informa si la instalación está enrolada, y contra qué

Los tres exigen **HTTPS**, sin exención ni modo de desarrollo: es la misma frontera que la
emisión de eventos.

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

### `permea status` — diagnóstico local

    permea status

Dice si la instalación está enrolada y contra qué backend. Es **local**: no contacta con nadie.
**Nunca imprime el token**, a lo sumo un indicador de presencia.

## Primeros pasos
    make test    # test de frontera en verde (empezar por aquí)
    make run     # dry-run: imprime eventos desde el fixture, sin transmitir
    make build   # binario en bin/permea

## Modos de ejecución

    permea --scan <fichero.jsonl>   # dry-run: imprime eventos de un JSONL, sin tocar estado ni cola
    permea --run                    # una pasada: escanea, encola y drena al backend (US1 + US2)
    permea --daemon                 # bucle continuo: cada sync_interval genera y transmite

- **`--run`** hace una pasada: descubre los logs de Claude Code, lee solo lo nuevo por
  offset (idempotente), encola de forma durable en `queue.jsonl` y, si hay `endpoint`
  configurado, drena la cola por HTTPS autenticado.
- **`--daemon`** repite lo anterior cada `sync_interval`. Errores de red/5xx se reintentan
  con backoff acotado (máx. 5 reintentos, tope 5 min) y el lote permanece en cola; un error
  de autenticación (401/403) detiene el sync por configuración errónea. `Ctrl-C` para parar.
- Sin `endpoint` configurado, la medición local funciona igual: los eventos quedan en la
  cola y nada se transmite.

## Configuración y rutas por SO

Toda la configuración y el estado viven en el **directorio de datos por SO** (creado al
primer arranque), resuelto vía `os.UserConfigDir` — nunca se hardcodean rutas:

| SO | Directorio de datos |
|---|---|
| Linux | `$XDG_CONFIG_HOME/permea` (o `~/.config/permea`) |
| macOS | `~/Library/Application Support/permea` |
| Windows | `%AppData%\permea` |

Ahí se guardan `config.json` (endpoint, token, identidad, `sync_interval`, modo de ref),
`state.json` (offset de escaneo), `queue.jsonl` (cola offline) y `salt` (secreto local,
`0600`, nunca transmitido). Los logs de Claude Code se resuelven en `~/.claude/projects`
por SO, con override opcional `logs_root` en la config. Escrituras siempre atómicas
(temporal + `os.Rename`).

## Portabilidad
Binario estático único, **sin CGO ni dependencias externas** (solo stdlib). Compila para
Linux, macOS y Windows:

    CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build ./cmd/permea
    CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build ./cmd/permea
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./cmd/permea

La versión del binario (`agent_version` en el evento) se inyecta con
`-ldflags "-X main.version=<versión>"`.

## Estructura
    cmd/permea        punto de entrada (subcomandos enroll/status/project join + modos scan/run/daemon)
    internal/event    LA FRONTERA (struct cerrado del evento)
    internal/ingest   lectores por herramienta (claude_code) + tests de frontera
    internal/pricing  cálculo de coste local (tabla empaquetada)
    internal/state    escaneo incremental idempotente
    internal/transport cliente HTTPS + cola offline + entrega exactamente-una-vez
    internal/config   configuración local, rutas por SO, salt e identidades

Renombrar el módulo en `go.mod` (`github.com/permea-dev/agent`) al repo real.
