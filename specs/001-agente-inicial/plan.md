# Implementation Plan: Agente inicial — ingesta de Claude Code y frontera de datos

**Branch**: `001-agente-inicial` | **Date**: 2026-07-02 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-agente-inicial/spec.md`

## Summary

Agente local en Go que escanea de forma incremental los logs de uso de Claude Code
(JSONL bajo `~/.claude/projects/**`), deriva por cada llamada al modelo sus contadores
de tokens, calcula el coste **en local** con una tabla empaquetada, y transmite al
backend de equipo **únicamente** un `event.Event` de allowlist cerrada —nunca contenido.
Funciona sin conexión: los eventos pendientes se persisten en una cola local y se
transmiten **exactamente una vez** al recuperarse la red. El esqueleto de la frontera
(`internal/event`, `internal/ingest`) y su golden test ya existen; este plan completa el
escaneo incremental idempotente (`internal/state`), la cola offline con reintentos
(`internal/transport`), la configuración local persistida (`internal/config`), la
resolución de rutas por SO y el bucle del agente en `cmd/permea`.

## Technical Context

**Language/Version**: Go 1.22 (toolchain verificada: go1.22.2). Binario único vía `cmd/permea`.

**Primary Dependencies**: **Solo librería estándar** (`net/http`, `encoding/json`,
`crypto/sha256`, `crypto/rand`, `os`, `bufio`, `path/filepath`, `time`). No se introduce
ninguna dependencia externa (Principio III). `golangci-lint` es herramienta de desarrollo,
no dependencia del binario.

**Storage**: Ficheros locales legibles bajo el directorio de datos por SO:
- `config.json` — configuración del usuario (endpoint, identidad, modo de ref, intervalo).
- `state.json` — offset de escaneo por fichero (escaneo incremental idempotente).
- `queue.jsonl` — cola append-only de eventos pendientes de transmitir (offline).
- `salt` — secreto de hashing generado una sola vez; **nunca** se transmite.

**Testing**: `go test ./...`; **golden test de frontera** (`internal/ingest/boundary_test.go`,
ya en verde) como primera línea de defensa; tests de idempotencia (state) y de
exactamente-una-vez (transport/queue).

**Target Platform**: macOS, Linux y Windows nativo. El agente contempla que el cliente
ejecute en Windows nativo aunque el desarrollo ocurra en WSL/Linux. Resolución de rutas
por SO; nunca se hardcodean rutas de logs.

**Project Type**: Agente/CLI de escritorio de proyecto único (un binario estático).

**Performance Goals**: No es crítico en latencia. Escaneo incremental con lectura por
offset y consumo de memoria acotado (streaming línea a línea con `bufio.Scanner`, buffer
1 MiB ya presente); procesar el uso diario típico (miles de llamadas, decenas de MB) en
segundos.

**Constraints**: offline-capable; frontera deny-by-default inviolable; cero dependencias
externas; coste calculado sin backend; transporte HTTPS autenticado; sin reprocesado ni
duplicados.

**Scale/Scope**: una máquina de desarrollador; una sola herramienta soportada en esta
spec (Claude Code); logs de hasta decenas de MB y miles de llamadas.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Puerta | Estado |
|---|---|---|
| **I. Frontera de datos inviolable** | El evento es un struct cerrado (allowlist); no hay passthrough de campos crudos; deny-by-default en la decodificación; identificadores sensibles solo como hash salado con salt local. | ✅ PASS — `event.Event` cerrado y `ingest.rawRecord` decodifica solo campos permitidos. El plan **no amplía** la allowlist con contenido. El `salt` se persiste solo en local. |
| **II. Privacidad auditable, local-first** | Código abierto; parseo/coste/agregación en local; tabla de precios empaquetada; funciona sin red; frontera verificable leyendo solo la serialización del evento. | ✅ PASS — coste en `internal/pricing` con tabla embebida; cola offline garantiza medición sin red; la revisión externa se apoya en `internal/event` (punto único de construcción del evento). |
| **III. Binario único y auditable** | Binario estático único; se favorece stdlib; toda dependencia externa se justifica; resolución de rutas por SO; legibilidad sobre astucia. | ✅ PASS — se mantiene **cero dependencias externas**; rutas resueltas por SO en `internal/config`; sin CGO. |
| **IV. Test-first en la frontera** | Golden test de frontera antes que cualquier parser; escenarios Given/When/Then; requisitos DEBE/NUNCA; nada se cierra con el test de frontera en rojo. | ✅ PASS — `boundary_test.go` ya existe y pasa. Nuevas capacidades (state, queue) añaden sus tests antes/junto a la implementación. |
| **V. Desarrollo dirigido por especificaciones** | Ciclo specify→plan→tasks→implement; artefactos en `specs/NNN-slug/`; spec sin detalles de cómo; consistencia con la constitución. | ✅ PASS — este plan y sus artefactos derivan del `spec.md` sin redefinir la frontera. |

**Resultado del gate**: sin violaciones. La sección *Complexity Tracking* queda vacía.

## Project Structure

### Documentation (this feature)

```text
specs/001-agente-inicial/
├── plan.md              # Este fichero (/speckit.plan)
├── research.md          # Fase 0 (/speckit.plan)
├── data-model.md        # Fase 1 (/speckit.plan)
├── quickstart.md        # Fase 1 (/speckit.plan)
├── contracts/           # Fase 1 (/speckit.plan)
│   ├── boundary-event.md   # Contrato de frontera: allowlist/denylist + esquema del evento
│   └── transport.md        # Contrato HTTPS de ingesta (petición/respuesta, auth)
└── tasks.md             # Fase 2 (/speckit.tasks — NO lo crea /speckit.plan)
```

### Source Code (repository root)

Estructura de referencia fijada por la constitución (ya existente); este plan la completa,
no la reinventa.

```text
cmd/
└── permea/
    └── main.go          # Bucle del agente: cargar config → escanear → encolar → transmitir
                         #   (hoy: dry-run --scan; se amplía a run incremental + sync)

internal/
├── event/
│   └── event.go         # LA FRONTERA: struct cerrado + Ref() (hash salado) + NewID()  [existe]
├── ingest/
│   ├── claudecode.go    # Lector Claude Code: decodifica solo campos permitidos          [existe]
│   ├── boundary_test.go # Golden test de frontera (deny-by-default)                       [existe]
│   └── testdata/
│       └── claude_code_sample.jsonl                                                        [existe]
├── pricing/
│   └── pricing.go       # Coste local con tabla empaquetada; + señal de "coste no disponible"
├── state/
│   └── state.go         # Escaneo incremental idempotente: Load/Save + avance por offset
├── transport/
│   └── transport.go     # Cliente HTTPS + cola offline + reintentos; dedup por event_id
└── config/
    └── config.go        # Config local + resolución de rutas por SO + salt persistido
```

**Structure Decision**: Proyecto único en Go con el layout `cmd/` + `internal/` que exige
la constitución. El código auditable de la frontera se concentra en `internal/event` (punto
único donde se construye y serializa lo que cruza). Cada capacidad nueva del plan cae en su
paquete ya reservado; no se crean paquetes nuevos ni se mueve la frontera.

## Complexity Tracking

> Sin violaciones de la Constitution Check. No hay complejidad que justificar.
