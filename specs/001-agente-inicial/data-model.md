# Phase 1 — Data Model: Agente inicial (Claude Code)

Entidades del agente y sus reglas. La entidad rectora es **Evento de métrica** (`event.Event`):
el único dato que cruza la frontera. Todo lo demás vive y muere en local.

Referencias: [spec.md](./spec.md) · [research.md](./research.md) ·
[contracts/boundary-event.md](./contracts/boundary-event.md)

---

## 1. Event — la frontera (cruza) · `internal/event`

Struct **cerrado** (allowlist). No admite passthrough de campos crudos del log.

| Campo | Tipo | Origen | Reglas |
|---|---|---|---|
| `schema_version` | int | constante (`SchemaVersion`) | Fijo por versión de contrato. |
| `agent_version` | string | local (`Context.AgentVersion`) | Trazabilidad del binario. |
| `event_id` | string | `event.NewID()` (crypto/rand) | Único; clave de deduplicación (FR-006). |
| `occurred_at` | time.Time | log (`timestamp`) | Marca de la llamada. |
| `tool` | string | fijo `"claude_code"` | Origen de la métrica. |
| `model` | string | log (`message.model`) | Modelo empleado. |
| `tokens_input` | int | log (`usage.input_tokens`) | ≥ 0. |
| `tokens_output` | int | log (`usage.output_tokens`) | ≥ 0. |
| `tokens_cache_creation` | int | log (`usage.cache_creation_input_tokens`) | ≥ 0. |
| `tokens_cache_read` | int | log (`usage.cache_read_input_tokens`) | ≥ 0. |
| `cost_usd` | float64 | `pricing.Cost` (local) | Válido solo si `cost_available`. |
| `cost_available` | bool | `pricing.Cost` (local) | **Nuevo (R5)**: `false` si el modelo no está en la tabla. |
| `project_ref` | string | `Ref(salt, cwd)` | Hash salado; nunca la ruta en claro (salvo opt-in `plain`). |
| `session_ref` | string | `Ref(salt, sessionId)` | Hash salado. |
| `machine_ref` | string | `Ref(salt, machineID)` | Hash salado. |
| `dev_id` | string | local (`Context.DevID`) | Identidad, no contenido. |
| `org_id` | string | local (`Context.OrgID`) | Identidad, no contenido. |

**Invariantes de frontera** (verificadas por `boundary_test.go`):
- Ningún campo contiene texto de prompt/respuesta, código, diffs, rutas o nombres en claro,
  ni argumentos/resultados de herramientas, ni secretos (denylist).
- `project_ref`/`session_ref`/`machine_ref` son hashes; el `salt` que los produce **no**
  forma parte del `Event` ni se transmite.
- Un campo nuevo o desconocido del log **no** puede aparecer: no existe en el struct.

**Cambio respecto al código actual**: se añade `cost_available` (bool). Es metadato de
calidad de la medición, no contenido → compatible con el Principio I (ver R5). Requiere
extender `pricing.Cost` para devolver `(float64, bool)` y ajustar `ingest.FromClaudeCodeLine`.

---

## 2. Context — datos locales del agente (NO cruza como tal) · `internal/ingest`

Lo que el agente añade y que **nunca** proviene del log.

| Campo | Tipo | Uso |
|---|---|---|
| `Salt` | string | Semilla de `Ref()`. **Nunca** se transmite (R6). |
| `MachineID` | string | Se transmite solo su hash (`machine_ref`) (R7). |
| `DevID` | string | Va tal cual como `dev_id` (identidad). |
| `OrgID` | string | Va tal cual como `org_id` (identidad). |
| `AgentVersion` | string | Va como `agent_version`. |

---

## 3. FileState / Store — escaneo incremental (local) · `internal/state`

Persistido en `state.json`. Garantiza idempotencia (FR-006, SC-003).

| Entidad | Campo | Tipo | Reglas |
|---|---|---|---|
| `FileState` | `path` | string | Clave del fichero de log. |
| | `size` | int64 | Tamaño observado en la última pasada (detección de truncado). |
| | `mod_time` | int64 | Marca de modificación (epoch). |
| | `offset` | int64 | Byte hasta el que se ha consumido (fin de la última línea completa). |
| `Store` | `files` | map[string]FileState | Estado por ruta. |

**Transiciones**:
- `offset` solo **avanza** salvo truncado (`size < offset` → `offset = 0`).
- Persistencia **atómica** (temporal + `rename`).
- Reprocesar un fichero sin cambios (mismo `size`/`mod_time`) produce **0** eventos nuevos.

---

## 4. QueueItem — cola offline (local) · `internal/transport`

`queue.jsonl`: append-only, una línea por evento pendiente. Garantiza FR-007 / SC-004.

| Elemento | Tipo | Reglas |
|---|---|---|
| Línea de cola | `event.Event` serializado | Se añade **tras** avanzar el estado; se elimina solo tras `2xx`. |

**Ciclo de vida**: `pendiente` → (envío `2xx`) → `confirmado` (eliminado de la cola).
Reintentos con backoff; sin red permanece `pendiente`. La dedup extremo a extremo es por
`event_id` (el backend ignora `event_id` ya visto).

---

## 5. Config — configuración local (local) · `internal/config`

`config.json` legible por el usuario (FR-010).

| Campo | Tipo | Por defecto | Reglas |
|---|---|---|---|
| `endpoint` | string | — | **DEBE** ser `https://…` (R9). |
| `device_token` | string | — | Token por instalación; auth Bearer. |
| `org_id` | string | — | Identidad de organización. |
| `dev_id` | string | — | Identidad de desarrollador. |
| `project_ref_mode` | `ProjectRefMode` | `hash` | `hash` (por defecto) u opt-in `plain`. |
| `tools` | []string | `["claude_code"]` | Herramientas activas (esta spec: solo claude_code). |
| `sync_interval` | string | `"60s"` | Cadencia del ciclo de sync. |

Campos derivados/no-config resueltos por SO (R8): directorio de datos, ruta base de logs de
Claude Code, ubicación de `salt`, `state.json`, `queue.jsonl`.

---

## 6. Rate / Table — pricing local (local) · `internal/pricing`

| Elemento | Tipo | Reglas |
|---|---|---|
| `Rate` | struct{Input, Output, CacheWrite, CacheRead float64} | USD por millón de tokens. |
| `Table` | map[string]Rate | Empaquetada en el binario; nunca depende del backend (Principio II). |
| `Cost(model, in, out, cCreate, cRead)` | → `(float64, bool)` | `bool=false` si el modelo no está en `Table` (R5). |

---

## Diagrama de flujo de datos (frontera destacada)

```text
~/.claude/projects/**/*.jsonl
        │  (bufio, línea a línea; solo campos permitidos)
        ▼
  ingest.rawRecord ──► ingest.FromClaudeCodeLine ──► event.Event
        │  (offset)           │ (pricing.Cost local)   │ (Ref = hash salado, salt local)
        ▼                     ▼                         ▼
   state.json          cost_usd/cost_available    queue.jsonl ──HTTPS+Bearer──► backend
                                                        (dedup por event_id)
```

Solo la caja `event.Event` cruza la frontera. `salt`, rutas, prompts y estado permanecen en
local por construcción.
