---
description: "Task list for feature implementation"
---

# Tasks: Agente inicial — ingesta de Claude Code y frontera de datos

**Input**: Design documents from `/specs/001-agente-inicial/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅ (boundary-event.md, transport.md), quickstart.md ✅

**Tests**: Test tasks ARE included. Este proyecto es **test-first en la frontera** por mandato de la constitución (Principio IV) y del input del usuario: el struct de frontera (`internal/event`) y su golden test van **antes que cualquier parser**. Los tests de cada capacidad nueva (state, transport, pricing) se escriben y **deben fallar** antes de implementar.

**Organization**: Tareas agrupadas por historia de usuario. Las tres historias son P1; US1 es el MVP.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Puede ejecutarse en paralelo (fichero distinto, sin dependencias sobre tareas incompletas)
- **[Story]**: US1, US2, US3 (mapea a las historias de spec.md)
- Toda descripción incluye ruta de fichero exacta

## Path Conventions

Proyecto único en Go (layout `cmd/` + `internal/` fijado por la constitución). Rutas relativas a la raíz del repo: `cmd/permea/`, `internal/event/`, `internal/ingest/`, `internal/pricing/`, `internal/state/`, `internal/transport/`, `internal/config/`.

## Estado de partida (no reinventar)

- ✅ `internal/event/event.go` — struct cerrado `Event`, `Ref()`, `NewID()` (en verde).
- ✅ `internal/ingest/claudecode.go` + `boundary_test.go` — golden test de frontera en verde (2 eventos sobre fixture de 3 líneas).
- ⚠️ Pendiente de ampliación con metadato (R5): `cost_available` en `Event` y `pricing.Cost → (float64, bool)`. Es cambio de frontera ⇒ **test-first** (Fase 2).
- ⛔ Stub: `internal/state`, `internal/transport`, `internal/config` (solo structs). `cmd/permea` solo dry-run `--scan`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirmar línea base verde y utillaje antes de tocar la frontera.

- [X] T001 Verificar toolchain y baseline verde desde la raíz del repo: `go version` (go1.22.x), `go build ./...`, `go test ./...`, `go vet ./...` y `golangci-lint run` — dejar constancia de que `internal/ingest` (golden test) pasa antes de empezar
- [X] T002 [P] Añadir targets de test por paquete en `Makefile` (`test-state`, `test-transport`, `test-pricing`, `test-config`) reutilizando `go test ./internal/<pkg>`
- [X] T003 [P] Confirmar que `.golangci.yml` cubre los paquetes nuevos (state, transport, config) sin excepciones que oculten fugas de frontera

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Infraestructura compartida por TODAS las historias: ampliación de frontera con `cost_available` (test-first) y resolución local por SO (directorio de datos, `salt`, identidades, ruta de logs).

**⚠️ CRITICAL**: Ninguna historia puede completarse hasta terminar esta fase. La ampliación de la frontera se hace **test-first**: el golden test se actualiza y falla antes de tocar `event.go`.

### Frontera: ampliación con metadato `cost_available` (test-first, R5)

- [X] T004 [P] Ampliar el golden test de frontera en `internal/ingest/boundary_test.go`: aseverar que el evento serializado incluye `cost_available` y que un modelo desconocido produce `cost_available=false` con `cost_usd=0` pero tokens contabilizados — **debe FALLAR** antes de T006–T008
- [X] T005 [P] Añadir tests de pricing en `internal/pricing/pricing_test.go`: `TestCost` (coste ±1% frente a cálculo de referencia, SC-001) y `TestCost_UnknownModel` (`(0,false)`) — **deben FALLAR** antes de T007
- [X] T006 Añadir el campo cerrado `CostAvailable bool` con tag `json:"cost_available"` a `Event` en `internal/event/event.go`, manteniendo el struct como allowlist (sin passthrough)
- [X] T007 Cambiar la firma de `Cost` a `(float64, bool)` en `internal/pricing/pricing.go`: modelo ausente de `Table` ⇒ `(0, false)`; presente ⇒ `(coste, true)`
- [X] T008 Actualizar `FromClaudeCodeLine` en `internal/ingest/claudecode.go` para poblar `CostUSD` y `CostAvailable` desde `pricing.Cost` (depende de T006, T007) — dejar T004 y T005 en verde

### Configuración local, rutas por SO, salt e identidades (R6, R7, R8)

- [X] T009 [P] Tests de configuración en `internal/config/config_test.go`: resolución del directorio de datos por SO, `Load`/`Save` round-trip, aplicación de `Default()`, y rechazo de endpoint no-`https` — **deben FALLAR** antes de T010–T014
- [X] T010 Implementar `DataDir()` (resolución por SO vía `os.UserConfigDir`→`permea`, creación del directorio con permisos) en `internal/config/config.go`
- [X] T011 Implementar `Load(path)`/`Save(path)` de `config.json` con escritura atómica (temporal + `os.Rename`), aplicando `Default()` a campos vacíos, en `internal/config/config.go`
- [X] T012 Implementar `LoadOrCreateSalt(dir)` en `internal/config/identity.go`: generar `salt` (`crypto/rand`, 32 bytes hex) al primer arranque, persistir en fichero `salt` con permisos `0600`; **nunca** se transmite (R6)
- [X] T013 Implementar `LoadOrCreateMachineID(dir)` en `internal/config/identity.go`: UUID de instalación aleatorio persistido (fuente estable y no sensible); solo su hash cruza como `machine_ref` (R7)
- [X] T014 Implementar `ClaudeCodeLogsRoot(cfg)` en `internal/config/config.go`: resolver `~/.claude/projects` por SO (Linux/macOS `$HOME`, Windows `%USERPROFILE%`) con override configurable (R1)

**Checkpoint**: Frontera ampliada y en verde; configuración, salt, identidades y rutas resueltas por SO. Las historias pueden empezar.

---

## Phase 3: User Story 1 — Sincronización de métricas tras una sesión (Priority: P1) 🎯 MVP

**Goal**: Detectar uso nuevo de Claude Code de forma **incremental e idempotente**, generar un evento de frontera por llamada facturable (métricas + referencias pseudónimas) y persistirlo de forma durable en la cola, sin reprocesar llamadas ya emitidas.

**Independent Test**: Ejecutar el escaneo dos veces sobre el mismo origen sin cambios ⇒ la segunda pasada produce **0** eventos nuevos (offset persistido en `state.json`); cada evento generado lleva `project_ref` hasheado y ninguna ruta/prompt en claro (V2, V3, SC-003).

### Tests for User Story 1 (test-first) ⚠️

- [X] T015 [P] [US1] Test de idempotencia `TestIncremental_NoReprocess` en `internal/state/state_test.go`: dos pasadas sobre el mismo fichero sin cambios ⇒ 0 eventos nuevos; truncado (`size < offset`) ⇒ reinicio a offset 0 — **debe FALLAR**
- [X] T016 [P] [US1] Test de recorrido de directorio en `internal/state/scan_test.go`: enumera `*.jsonl` bajo una raíz de proyectos temporal (subdirectorios anidados) e ignora no-`.jsonl` — **debe FALLAR**
- [X] T017 [P] [US1] Test de cola de generación `TestQueue_Append` en `internal/transport/queue_test.go`: `Append` añade una línea JSON por evento y `Load` las recupera en orden — **debe FALLAR**

### Implementation for User Story 1

- [X] T018 [US1] Implementar `Store.Load(path)`/`Store.Save(path)` con escritura atómica (temporal + `os.Rename` en el mismo sistema de ficheros) en `internal/state/state.go`
- [X] T019 [US1] Implementar el escaneo incremental por offset en `internal/state/state.go`: `Seek(offset)`, leer solo **líneas completas** nuevas, avanzar `offset` hasta el fin de la última línea completa, y detectar truncado (`size < offset` ⇒ `offset = 0`) — sin `internal/state/scan.go` aún (separado en T020)
- [X] T020 [P] [US1] Implementar el recorrido del directorio de proyectos en `internal/state/scan.go`: enumerar `~/.claude/projects/**/*.jsonl` (usando `ClaudeCodeLogsRoot`) con `filepath.WalkDir`, devolviendo rutas ordenadas (independiente del offset de T019)
- [X] T021 [P] [US1] Implementar `Queue.Append(dir, ev)` y `Queue.Load(dir)` sobre `queue.jsonl` append-only (una línea `event.Event` por evento) en `internal/transport/queue.go`
- [X] T022 [US1] Cablear el pipeline de generación en `cmd/permea/main.go`: cargar config+salt+machineID → recorrer directorio (T020) → por fichero, leer incrementalmente (T019) → `ingest.FromClaudeCodeLine` → `Queue.Append` **y solo entonces** avanzar y persistir `Store.Save` (orden de durabilidad de R4) — sustituye el salt hardcodeado `"dry-run-salt"`
- [X] T023 [US1] Test de integración `TestScan_TwoPasses_NoDuplicates` en `internal/state/state_test.go` (o `cmd/permea`): primera pasada genera N eventos en cola, segunda pasada 0 nuevos (SC-003)

**Checkpoint**: US1 funcional y testeable de forma independiente — escaneo incremental idempotente que deja eventos de frontera durables en `queue.jsonl`, sin contenido en claro.

---

## Phase 4: User Story 2 — Funcionamiento sin conexión (Priority: P1)

**Goal**: Medir sin red conservando los eventos pendientes en `queue.jsonl` y, al recuperarse la red, transmitirlos por HTTPS autenticado **exactamente una vez** (at-least-once en cliente + dedup por `event_id`), eliminando de la cola solo los confirmados mediante **reescritura atómica**.

**Independent Test**: Sin backend, procesar uso ⇒ los eventos quedan en `queue.jsonl` y la medición no se detiene; al volver la red (backend simulado con `httptest.NewTLSServer`), el nº de eventos recibidos == nº de llamadas reales y reenviar un lote ya aceptado no duplica (V4, SC-004).

### Tests for User Story 2 (test-first) ⚠️

- [ ] T024 [P] [US2] Tests de cola offline/drenaje `TestQueue_OfflineThenDrain` y `TestQueue_ExactlyOnce` en `internal/transport/transport_test.go` con `httptest.NewTLSServer`: sin red la cola crece; al drenar, el backend recibe exactamente N; reenvío del mismo lote es idempotente por `event_id` — **deben FALLAR**
- [ ] T025 [P] [US2] Test de reescritura atómica `TestQueue_AtomicRewrite_KeepsUnconfirmed` en `internal/transport/queue_test.go`: tras `2xx` de un subconjunto, `queue.jsonl` conserva solo los no confirmados y se reescribe vía temporal + `os.Rename` (mismo sistema de ficheros) — **debe FALLAR**
- [ ] T026 [P] [US2] Test de contrato de transporte `TestSend_RejectsHTTP`/`TestSend_StatusSemantics` en `internal/transport/transport_test.go`: endpoint `http://` se rechaza; `2xx`=confirmar, `401/403`=detener sync, `5xx`/error de red=reintentar — **debe FALLAR**

### Implementation for User Story 2

- [ ] T027 [US2] Endurecer `Client.Send` en `internal/transport/transport.go`: validar esquema `https` (rechazar `http://`), fijar `Content-Type` y `Authorization: Bearer`, e interpretar el código de estado según `contracts/transport.md` (`2xx`/`401`/`403`/`4xx`/`5xx`)
- [ ] T028 [US2] Implementar la reescritura atómica de `queue.jsonl` tras `2xx` en `internal/transport/queue.go`: escribir los no confirmados a temporal + `os.Rename` en el mismo directorio (mismo sistema de ficheros); nunca borrado in-place
- [ ] T029 [US2] Implementar reintentos con backoff exponencial acotado en `internal/transport/transport.go`: `5xx`/error de red reintentan hasta un **máximo de 5 reintentos** con **delay máximo de 5 minutos** por espera (tope del backoff); agotados los reintentos, el lote permanece en cola para el siguiente ciclo de sync; `401/403` detienen el sync (config errónea); sin red la cola permanece
- [ ] T030 [US2] Implementar el paso de sync `Drain(cfg)` en `internal/transport/queue.go`: `Queue.Load` en lotes → `Client.Send` → en `2xx` reescritura atómica (T028); la dedup extremo a extremo se apoya en `event_id`
- [ ] T031 [US2] Cablear el sync en el bucle del agente en `cmd/permea/main.go`: ticker con `sync_interval` de la config, ejecutando generación (US1) y drenaje (T030); modo `run` frente al `--scan` dry-run existente

**Checkpoint**: US1 + US2 funcionan de forma independiente — medición offline sin pérdida y entrega exactamente-una-vez al recuperar la red.

---

## Phase 5: User Story 3 — Contenido inesperado en el origen (Priority: P1, seguridad)

**Goal**: Garantizar que un dato nuevo o desconocido en el log (p. ej. una futura versión de Claude Code que añada contenido de prompt) **nunca** aparece en el evento transmitido (deny-by-default). El evento e ingest ya están en verde; esta historia **blinda** la garantía con tests de regresión y una aserción de "solo allowlist".

**Independent Test**: Inyectar en el fixture un campo no contemplado con contenido sensible (p. ej. `message.content`, un campo extra arbitrario) y confirmar que no aparece en el JSON del evento; el conjunto de claves del evento serializado == allowlist exacta (V1, SC-002, SC-005).

### Tests for User Story 3 (test-first) ⚠️

- [ ] T032 [P] [US3] Ampliar `internal/ingest/testdata/claude_code_sample.jsonl` y `internal/ingest/boundary_test.go`: inyectar un campo futuro con contenido (`message.content`, ruta en claro, argumentos de herramienta, un campo desconocido arbitrario) y aseverar que **ningún** término de la denylist sobrevive al evento serializado — **debe FALLAR** si algo se filtra
- [ ] T033 [P] [US3] Añadir `TestEvent_OnlyAllowlistKeys` en `internal/event/event_test.go`: serializar un `Event` y comprobar que el conjunto de claves JSON es **exactamente** la allowlist de `contracts/boundary-event.md` (equivalente a `additionalProperties:false`) — **debe FALLAR** si aparece una clave nueva

### Implementation for User Story 3

- [ ] T034 [US3] Confirmar y documentar en `internal/ingest/claudecode.go` que `rawRecord` decodifica **solo** campos de la allowlist (sin `message.content` ni texto); añadir comentario-guardia que prohíba ampliar `rawRecord` con contenido — dejar T032/T033 en verde
- [ ] T035 [US3] Revisión de frontera (SC-006): verificar leyendo **solo** `internal/event/event.go` e `internal/ingest/claudecode.go` que ningún contenido cruza; anotar el resultado en `quickstart.md` (sección "Revisión de frontera")

**Checkpoint**: Las tres historias funcionan de forma independiente; la frontera está blindada frente a datos futuros.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Puertas de calidad, portabilidad y validación de extremo a extremo.

- [ ] T036 [P] Cablear `agent_version` real: propagar la variable `version` de `cmd/permea/main.go` al `Context.AgentVersion` y verificar que llega a `Event.AgentVersion`
- [ ] T037 [P] Verificación multiplataforma de build: `GOOS=linux/darwin/windows go build ./cmd/permea` sin CGO ni dependencias externas (SC-007)
- [ ] T038 Ejecutar la validación de `quickstart.md` (V1–V7) y dejar constancia de resultados
- [ ] T039 [P] Puertas de calidad de la constitución desde la raíz: `go vet ./...` sin hallazgos, `golangci-lint run` limpio, `go test ./...` en verde
- [ ] T040 [P] Actualizar `README.md`: modo `run` (config, sync), resolución de rutas/datos por SO, y garantía de frontera

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias — empieza de inmediato.
- **Foundational (Phase 2)**: depende de Setup — **BLOQUEA** todas las historias (amplía la frontera y resuelve rutas/salt/identidades).
- **User Stories (Phase 3–5)**: todas dependen de Foundational.
  - US1 (P1, MVP): escaneo incremental idempotente + generación durable en cola.
  - US2 (P1): reutiliza `queue.jsonl` de US1 para drenaje/transporte; conceptualmente independiente y testeable por sí sola con backend simulado.
  - US3 (P1, seguridad): independiente — refuerza la frontera ya verde; no depende de US1/US2.
- **Polish (Phase 6)**: depende de las historias deseadas completas.

### User Story Dependencies

- **US1 (P1)**: arranca tras Foundational. Sin dependencias sobre otras historias.
- **US2 (P1)**: arranca tras Foundational. Consume la cola producida por US1 en el flujo completo, pero sus tests (backend simulado + cola sembrada) la hacen testeable de forma independiente.
- **US3 (P1)**: arranca tras Foundational. Totalmente independiente (tests sobre `event`/`ingest`).

### Within Each User Story

- Los tests se escriben y **deben FALLAR** antes de la implementación (test-first).
- Frontera (`internal/event`) y su golden test **antes** que cualquier parser (mandato del usuario / Principio IV).
- state (offset) y recorrido de directorio son tareas **separadas** (T019 vs T020).
- La cola se **append** (US1) antes de drenarse/reescribirse atómicamente (US2), y el evento se persiste **antes** de intentar transmitirse (durabilidad, R4).

### Parallel Opportunities

- Setup: T002, T003 en paralelo.
- Foundational: T004 y T005 en paralelo (ficheros de test distintos); T006 (event) y T007 (pricing) en paralelo; T009 en paralelo con los tests de frontera. T010–T014 tocan `config.go`/`identity.go` — secuenciar los que comparten fichero.
- US1: T015, T016, T017 en paralelo (tests en ficheros distintos); T020 (scan.go) y T021 (queue.go) en paralelo entre sí y con T018/T019 salvo por el orden de fichero de `state.go`.
- US2: T024, T025, T026 en paralelo (tests).
- US3: T032 y T033 en paralelo (tests en paquetes distintos).
- Polish: T036, T037, T039, T040 en paralelo.

---

## Parallel Example: User Story 1

```bash
# Lanzar los tests de US1 juntos (deben fallar primero):
Task: "TestIncremental_NoReprocess en internal/state/state_test.go"
Task: "Test de recorrido de directorio en internal/state/scan_test.go"
Task: "TestQueue_Append en internal/transport/queue_test.go"

# Implementar en paralelo las piezas de ficheros distintos:
Task: "Recorrido del directorio de proyectos en internal/state/scan.go"
Task: "Queue.Append/Load en internal/transport/queue.go"
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Fase 1: Setup (baseline verde).
2. Fase 2: Foundational (ampliar frontera con `cost_available` test-first + config/salt/rutas). **CRÍTICO**: bloquea todo.
3. Fase 3: US1 — escaneo incremental idempotente que deja eventos durables en `queue.jsonl`.
4. **PARAR y VALIDAR**: dos pasadas ⇒ 0 duplicados (SC-003); sin contenido en claro (SC-002).
5. Entregar/demostrar el MVP (medición local correcta y auditable).

### Incremental Delivery

1. Setup + Foundational → base lista.
2. + US1 → medición incremental idempotente (MVP).
3. + US2 → offline + entrega exactamente-una-vez.
4. + US3 → frontera blindada frente a datos futuros.
5. Cada historia añade valor sin romper las anteriores.

### Parallel Team Strategy

Tras Foundational, con varios desarrolladores:

- Dev A: US1 (state + traversal + pipeline).
- Dev B: US2 (transport + cola offline + backoff).
- Dev C: US3 (blindaje de frontera).

---

## Notes

- [P] = ficheros distintos, sin dependencias.
- La etiqueta [Story] mapea la tarea a su historia para trazabilidad.
- Verificar que los tests fallan antes de implementar (Principio IV).
- Escritura atómica **siempre** vía temporal + `os.Rename` en el mismo sistema de ficheros (`state.json`, `config.json`, `queue.jsonl`).
- Cero dependencias externas: solo stdlib (Principio III).
- Commit tras cada tarea o grupo lógico; parar en cualquier checkpoint para validar la historia de forma independiente.
