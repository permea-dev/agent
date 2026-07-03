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

    curl -fsSL https://raw.githubusercontent.com/bfgnet/agente_permea/main/install.sh | sh
    # opcional: PERMEA_VERSION=v1.4.0 PREFIX="$HOME/.local/bin" sh install.sh

**Windows** — Scoop (bucket propio):

    scoop bucket add permea https://github.com/bfgnet/scoop-permea
    scoop install permea

Verifica la instalación con `permea --version`. Detalle de canales e integridad en
[`specs/002-distribucion/contracts/install-contract.md`](specs/002-distribucion/contracts/install-contract.md).
Compilar desde fuente: ver [Portabilidad](#portabilidad).

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
    cmd/permea        punto de entrada (modos scan/run/daemon)
    internal/event    LA FRONTERA (struct cerrado del evento)
    internal/ingest   lectores por herramienta (claude_code) + tests de frontera
    internal/pricing  cálculo de coste local (tabla empaquetada)
    internal/state    escaneo incremental idempotente
    internal/transport cliente HTTPS + cola offline + entrega exactamente-una-vez
    internal/config   configuración local, rutas por SO, salt e identidades

Renombrar el módulo en `go.mod` (`github.com/bfgnet/agente_permea`) al repo real.
