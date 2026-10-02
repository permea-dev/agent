# Data Model — Medición fiel y publicable

**Feature**: `006-medicion-fiel` | **Fecha**: 2026-10-02

## No hay entidad persistente nueva, y es deliberado

006 **no añade ningún fichero** al directorio de datos (`config.json`, `state.json`, `queue.jsonl`,
`salt` y `machine_id` siguen siendo los mismos), **ni ningún campo** al evento. Lo que cambia es **qué
representa** un evento y **cómo se calcula** un valor que ya existía.

| Elemento | Antes | Después | Persistido |
|---|---|---|---|
| **Evento** | una línea facturable del log | **un mensaje** facturable | en la cola, como hoy |
| **`event_id`** | 16 bytes aleatorios en hex (`internal/event/event.go:48-55`) | `SHA-256` truncado del dominio y los identificadores del proveedor (`contracts/event-id.md`) | en la cola, como hoy |
| **Tabla de tarifas** | 3 claves, sin procedencia | 16 claves, espejo del catálogo, con cabecera (`contracts/tarifas.md`) | en el binario |

## Lo nuevo vive en memoria: la pasada

**Pasada** (en `internal/ingest`): el estado de **una** lectura de los logs. Nace y muere con ella
(`research.md` R3).

| Atributo | Qué es | Para qué |
|---|---|---|
| vistos | `event_id` → `usage` de la primera línea (4 enteros) | FR-033: no emitir dos veces en la pasada. FR-005: detectar consumo distinto |
| facturables | líneas `assistant` con modelo leídas | contexto del resumen |
| emitidos | eventos producidos | contexto del resumen |
| repetidas | líneas de un mensaje ya emitido en la pasada | contexto del resumen; con M2 son la mayoría |
| sintéticas | líneas `<synthetic>` descartadas | FR-007 |
| sin identificador | líneas sin `message.id` ni `requestId`, no emitidas | FR-006: visible |
| consumo distinto | repetidas cuyo `usage` difiere del de la primera | FR-005: visible |

**Ciclo de vida**:
- `--run`: una por ejecución;
- `--daemon`: una por ciclo;
- `--scan`: una por fichero.

**Nil es válido**: sin pasada se deriva igual y no se deduplica (patrón del `Resolutor`,
`internal/ingest/claudecode.go:48-54`).

**Invariante**: la pasada **nunca** guarda ni escribe identificadores del proveedor. Su clave es el
`event_id` ya derivado. Al resumen sólo salen recuentos.

## Lo que se lee del log, y no se guarda

`message.id` y `requestId` se añaden a lo que el lector decodifica (`rawRecord`,
`internal/ingest/claudecode.go:25-39`). Son metadatos técnicos, no contenido. Se usan para derivar el
`event_id` y se descartan. La guarda de frontera de `rawRecord` los admite como «metadatos derivados
de la allowlist»: el `event_id` está en ella. Su comentario se actualiza **por nombre** (disciplina 8).

## Verificación

- **Ningún fichero nuevo** en el directorio de datos tras una pasada en sandbox (lista de ficheros
  antes y después).
- **`git diff 0311fa1 -- internal/event` vacío** (SC-005).
- **La cola conserva su forma**: los mismos 17 campos por línea, `event_id` de 32 hex (SC-003).
