# Quickstart — Validación del agente inicial (Claude Code)

Guía ejecutable para validar que la funcionalidad cumple el [spec.md](./spec.md). No duplica
código; enlaza al [data-model](./data-model.md) y a los [contratos](./contracts/).

## Prerrequisitos

- Go 1.22+ (`go version`).
- Repositorio clonado; sin dependencias externas (solo stdlib).

## Puesta en marcha

```bash
make test    # test de frontera en verde — EMPIEZA POR AQUÍ (Principio IV)
make run     # dry-run: imprime eventos desde el fixture, sin transmitir nada
make build   # binario estático en bin/permea
make lint    # golangci-lint limpio (parte del producto)
```

## Escenarios de validación (Given/When/Then → comando)

### V1 — La frontera no filtra contenido (US1, US3, SC-002, SC-005)
El golden test inyecta contenido sensible (`SECRETO_DEL_PROMPT`, rutas, sesión privada) y
falla si algo de la denylist sobrevive al evento.

```bash
go test ./internal/ingest -run TestBoundary -v
```
**Esperado**: `TestBoundary_NoDenylistLeaks` y `TestBoundary_KeepsMetrics` en verde; 2
eventos generados sobre las 3 líneas del fixture (la línea `user` se ignora). Ver
[contracts/boundary-event.md](./contracts/boundary-event.md).

### V2 — Solo cruzan métricas y referencias pseudónimas (US1, SC-002)
```bash
make run
```
**Esperado**: por cada llamada de asistente, una línea con `tool`, `model`, tokens, coste y un
`project_ref` **hasheado** (nunca la ruta `/home/basilio/...`). Ninguna línea contiene prompt,
respuesta ni ruta en claro.

### V3 — Idempotencia / escaneo incremental (US1, FR-006, SC-003)
*(tras implementar `internal/state`)*
```bash
# Ejecutar el escaneo dos veces sobre el mismo origen sin cambios
go test ./internal/state -run TestIncremental_NoReprocess -v
```
**Esperado**: la segunda pasada produce **0** eventos nuevos (offset persistido en
`state.json`).

### V4 — Sin conexión y recuperación exactamente-una-vez (US2, FR-007, SC-004)
*(tras implementar la cola en `internal/transport`)*
```bash
go test ./internal/transport -run 'TestQueue_OfflineThenDrain|TestQueue_ExactlyOnce' -v
```
**Esperado**:
- Sin backend, los eventos quedan en `queue.jsonl` y la medición no se detiene.
- Al volver la red, el nº de eventos recibidos por el backend simulado (`httptest.NewTLSServer`)
  es **exactamente** el nº de llamadas reales; reenviar un lote ya aceptado no duplica (dedup
  por `event_id`). Ver [contracts/transport.md](./contracts/transport.md).

### V5 — Modelo desconocido: tokens sí, coste no disponible (edge case, FR-002)
*(tras extender `pricing.Cost` a `(float64, bool)`)*
```bash
go test ./internal/pricing -run TestCost_UnknownModel -v
```
**Esperado**: modelo ausente de la tabla → `cost_available=false`, tokens contabilizados, el
resto del procesamiento no se bloquea.

### V6 — Coste correcto frente a referencia (SC-001)
```bash
go test ./internal/pricing -run TestCost -v
```
**Esperado**: el coste agregado coincide (±1%) con el cálculo de referencia sobre los mismos
tokens y tabla.

### V7 — Instalación con un comando, multiplataforma (SC-007)
```bash
make build            # o: go build -o bin/permea ./cmd/permea
./bin/permea          # arranca sin dependencias previas
```
**Esperado**: binario estático único; arranca en macOS, Linux y Windows sin runtime externo.
Las rutas de logs y de datos se resuelven por SO (ver [research.md](./research.md#r1)).

## Puertas de calidad antes de cerrar (constitución)

```bash
go vet ./...          # sin hallazgos
golangci-lint run     # limpio
go test ./...         # frontera + suite completa en verde
```

## Revisión de frontera (SC-006)

Un revisor externo confirma el cumplimiento leyendo **solo** `internal/event/event.go` (el
struct cerrado y `Ref`) y el mapeo explícito en `internal/ingest/claudecode.go`. No hace falta
leer el resto del agente para verificar que ningún contenido cruza.
