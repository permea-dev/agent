# Implementation Plan: Medición fiel y publicable

**Branch**: `006-medicion-fiel` | **Date**: 2026-10-02 | **Spec**: [spec.md](./spec.md)

**Input**: `specs/006-medicion-fiel/spec.md`, commit `03d3734` (Draft, enmendada el 2026-10-02:
D-006-7 a D-006-13 y Q-006-1). Mediciones en [`research.md`](./research.md), contra `03d3734`.

---

## Summary

La `0.3.0` tiene que contar **una vez por mensaje**, a la **tarifa vigente**, con una **ayuda que dice
la verdad** y un **README con el que se instala**, y publicarse tras un ensayo en Windows.

**Técnicamente es pequeña, y el riesgo está en tres sitios**:
1. **La identidad del evento** pasa de aleatoria a derivada. Un error ahí no se ve en ningún test
   existente, y en producción duplica o funde mensajes en silencio. Por eso la derivación tiene
   contrato con **vectores de prueba calculados aparte** (`contracts/event-id.md`).
2. **Los fixtures del repositorio no traen identificadores** (`research.md` R7). Con la regla de FR-006,
   la suite entera se quedaría sin eventos. Se resuelve en B1, **antes** de que el código lo exija.
3. **La CLI decide hoy la ayuda después de cargar cosas.** Hay que moverla al principio de cada camino
   sin tocar los flags de ingesta (`research.md` R8).

**La revisión del plan NO encontró ningún requisito inimplementable ni choque con la constitución.**
Hay dos tensiones que se declaran y no se resuelven en silencio:
- la edición de un fixture que 004 declaró congelado (R7);
- la mitad de SC-003 que `--scan` no puede medir (quickstart V4).

Ninguna obliga a parar. Las dos van a §Dudas. *(Resueltas el 2026-10-02 por E-006-P1 y E-006-P2:
§Enmiendas del 2026-10-02.)*

---

## Technical Context

**Language/Version**: Go 1.22.2 (`go.mod`: `go 1.22`; CI: `1.22`).

**Primary Dependencies**: **ninguna nueva**. `crypto/sha256` y `encoding/hex` son de la librería
estándar (Principio III).

**Storage**: **ninguno nuevo** (`data-model.md`). La deduplicación vive en memoria, por pasada.

**Testing**: `go test ./...`, con `httptest` e `internal/testutil.Sandbox`.
- Línea base a preservar: **9 paquetes ok, 297 tests pass, 0 FAIL** (medido 2026-10-02 sobre
  `0311fa1`).
- **Ningún paquete nuevo** (`research.md` R1.4).

**Target Platform**: Linux, macOS y Windows (amd64; arm64 salvo Windows), como hasta ahora.

**Project Type**: CLI de un solo binario estático.

**Performance Goals**: no aplican como objetivo. Memoria de la pasada: del orden de 1–2 MB para la
referencia local de 10 698 mensajes, **estimado, se mide en B2b** (`research.md` R3.4).

**Constraints**:
- `internal/event` intacto (D-006-3);
- `event_id` de 32 hex (D-006-8);
- linter a 0 (D-006-13);
- el diff de `.goreleaser.yaml` y `release.yml` sólo en comentarios (FR-029);
- el espejo de tarifas no depende del otro repositorio en tiempo de test.

**Scale/Scope**: unos 10 ficheros de producción tocados y 3 nuevos, todos en paquetes existentes.
Detalle en §Project Structure.

---

## Constitution Check

*GATE: pasa antes de Phase 0 y se re-verifica tras Phase 1 del plan (contratos escritos).*

| Principio | Veredicto | Razón |
|---|:--:|---|
| **I · Frontera de datos inviolable** | ✅ | La allowlist no cambia: 17 campos, `SchemaVersion=1`, `internal/event` sin diff (FR-012, SC-005). Cambia el **valor** de `event_id`: un hash de dos identificadores técnicos del proveedor que **no son contenido** y **no cruzan en claro** (FR-003). El lector decodifica dos campos más, admitidos por la guarda de `rawRecord` como «metadatos derivados de la allowlist». Su comentario se actualiza por nombre. Análisis completo en spec §Verificación de la frontera. **No hay sal, a propósito** (D-006-2), y el Principio I sólo la exige para «ruta de proyecto, sesión, máquina» (`constitution.md:14`) |
| **II · Privacidad auditable, local-first** | ✅ | El coste sigue calculándose en local con la tabla empaquetada (`constitution.md:24`). El espejo es una **copia en el binario**, no una consulta al backend. El diagnóstico local no emite nada: sólo recuentos por stderr |
| **III · Binario único y auditable** | ✅ | Cero dependencias y cero paquetes nuevos. Una sola directiva `nolint`, **declarada y razonada** en el código (R9 #7). La alternativa engañaría al instrumento |
| **IV · Test-first en la frontera** | ✅ **con acción** | El golden se amplía con centinelas de identificador **en B1, antes** de que B2 escriba el código que los lee. **Nace verde**, porque hoy no se decodifican, y se valida por mutación (disciplina 3). El rojo de B1 es otro: el testigo de determinismo (`research.md` R2) |
| **V · Desarrollo dirigido por especificaciones** | ✅ | Spec commiteada, enmiendas fechadas, contratos escritos antes del código |
| **Puertas de calidad** (`constitution.md:72-75`) | ✅ **con acción** | `golangci-lint` limpio **para cerrar cualquier tarea**. Hoy hay 7 avisos, así que **B0 los corrige antes que nada** (D-006-13, R9). Desde B0, cada bloque cierra con vet, lint a 0 y suite verde |

**Re-verificación tras Phase 1 del plan**: los tres contratos (`event-id`, `cli`, `tarifas`) **no
introducen ningún campo ni camino de salida nuevo**. Sin cambios en el veredicto.

**Sin violaciones que justificar** → §Complexity Tracking queda vacía.

---

## Project Structure

### Documentation (this feature)

```text
specs/006-medicion-fiel/
├── spec.md                 # Draft, commiteada en 03d3734
├── plan.md                 # este fichero
├── research.md             # Phase 0 — R0 a R11
├── data-model.md           # Phase 1 — sin entidad persistente nueva; la «pasada» en memoria
├── contracts/event-id.md   # Phase 1 — derivación, vectores de prueba, una pasada = un evento por mensaje
├── contracts/cli.md        # Phase 1 — ayuda, ayudas de subcomando, subcomando inexistente
├── contracts/tarifas.md    # Phase 1 — espejo, cabecera, vigilancia, cómo se absorbe un cambio
├── quickstart.md           # Phase 1 — V1–V18, W1, P1, W2
└── tasks.md                # Phase 2 — lo genera /speckit.tasks, NO este comando
```

### Source Code (repository root)

```text
cmd/permea/
├── main.go           # MODIFICADO: escalera de ayuda y subcomando inexistente (B4); pasada y
│                     #   resumen en generate()/dryRun(), formato del dry-run (B2b, B2c)
├── ayuda.go          # NUEVO (B4): la fuente ÚNICA del texto de ayuda (tabla de subcomandos y opciones)
├── ayuda_test.go     # NUEVO (B4): SC-012 a SC-015
├── enroll.go         # MODIFICADO (B4): `-h`/`--help` antes de tocar stdin
├── status.go         # MODIFICADO (B4): recibe argumentos; `-h` antes de DataDir()
├── status_test.go    # MODIFICADO (B4): nueva firma de runStatus
├── project.go        # MODIFICADO: `_, _ =` en 4 escrituras (B0); `project -h` y `project join -h` (B4)
└── main_test.go      # MODIFICADO: identificadores en 2 líneas literales (B1); testigos de proceso
                      #   --scan/generate (B2b, B2c); actualizar sin reenviar (B2c)

internal/ingest/
├── claudecode.go     # MODIFICADO (B2a/B2b): rawRecord + message.id/requestId; <synthetic>;
│                     #   derivación; pasada
├── eventid.go        # NUEVO (B2a): derivación del event_id (contracts/event-id.md)
├── eventid_test.go   # NUEVO (B1 en rojo → B2a en verde): determinismo, vectores, tres formas
├── pasada.go         # NUEVO (B2b): conjunto de la pasada + contadores + resumen sin identificadores
├── pasada_test.go    # NUEVO (B2b): FR-033, SC-008, resumen sin identificadores
├── boundary_test.go  # MODIFICADO (B1): centinelas de identificador en la denylist; ids en 4 literales
└── testdata/
    ├── claude_code_sample.jsonl  # MODIFICADO (B1): message.id + requestId en las 2 líneas assistant
    └── boundary_sample.jsonl     # MODIFICADO (B1): ids CENTINELA en las 2 líneas assistant

internal/pricing/
├── pricing.go        # MODIFICADO (B3): 16 claves + cabecera
└── pricing_test.go   # MODIFICADO (B3): tabla esperada aparte; TestCost a 5/25

internal/config/
├── endpoint.go       # MODIFICADO (B0): JuzgarEndpoint → (admisible bool, errAnalisis error)
├── endpoint_test.go  # MODIFICADO (B0): orden de retorno
├── enrollment.go     # MODIFICADO (B0): orden de retorno
└── config.go         # MODIFICADO (B0): orden de retorno

internal/transport/
├── transport.go      # MODIFICADO (B0): orden de retorno en Send y Adherir (sin otro cambio)
└── adhesion_test.go  # MODIFICADO (B0): `r` → `_`; `//nolint:staticcheck` razonado

internal/event/       # SIN TOCAR (D-006-3)

specs/001-agente-inicial/contracts/transport.md   # MODIFICADO (B1): origen del event_id (FR-011)
specs/001-agente-inicial/data-model.md            # MODIFICADO (B1): ídem
README.md · CHANGELOG.md (NUEVO)                  # B5
.github/workflows/release.yml · .goreleaser.yaml  # B5: SÓLO comentarios
```

**Structure Decision**: se conserva la disposición. Lo nuevo vive en el paquete que ya tiene el dueño
del concepto:
- la identidad del evento de Claude Code, junto a su lector;
- la ayuda, junto al despacho.

`internal/event` no se toca, y la cuenta de 9 paquetes se conserva.

---

## Decisiones de plan

Todas razonadas, con su evidencia, en `research.md`. Aquí, la consecuencia.

| ID | Decisión | Research |
|---|---|---|
| **D-006-P1** | `event_id` = SHA-256 de un dominio versionado (`permea/event_id/v1` + herramienta + tipo) con longitud prefijada, truncado a 32 hex. Vive en `internal/ingest` | R1 |
| **D-006-P2** | `event.NewID` se queda sin llamantes de producción; un testigo de conducta (no un `grep`) impide que lo aleatorio vuelva | R1.5, R2 |
| **D-006-P3** | La pasada vive en `ingest.Context`, nil válido, y se instancia en `generate()` (un ciclo de daemon = una pasada) y en `dryRun()`. Entre pasadas no hay memoria | R3 |
| **D-006-P4** | Diagnóstico: un resumen por pasada, por stderr, sólo recuentos. En `--daemon`, sólo si la pasada leyó algo | R4 |
| **D-006-P5** | El formato del dry-run se amplía con las 4 partidas y el `event_id`; no lo consume ningún test | R5 |
| **D-006-P6** | Espejo de tarifas: literal Go con cabecera, y test con tabla esperada aparte, recuento y «ninguna sobra». Sin leer el otro repositorio | R6 |
| **D-006-P7** | Los fixtures de `internal/ingest/testdata` y las 6 líneas literales reciben identificadores en B1; `internal/project/testdata` no se toca | R7 |
| **D-006-P8** | CLI: una fuente de ayuda; ayuda antes de `flag.Parse` y del banner; `flag.Usage` por stdout; ayudas de subcomando en la primera línea de cada uno; subcomando inexistente con exit 1 y sin reproducir secretos | R8 |
| **D-006-P9** | Linter: `_, _ =` (×4), reordenar `JuzgarEndpoint`, `r`→`_`, un `nolint:staticcheck` razonado | R9 |
| **D-006-P10** | B0 es el linter, no los contratos: la puerta de la constitución aplica a toda tarea | R9 |

---

## Disciplinas

**Las de 005, sin cambios** (`specs/005-adhesion-a-proyecto/tasks.md`, §Disciplinas transversales):
1. una garantía por tarea, anclada al contrato;
2. rojo antes de verde, con la razón del fallo transcrita;
3. todo verde de nacimiento se valida por mutación, que compila y mata sólo su hecho, y se revierte
   por edición inversa;
4. los tests de proceso comparan `ExitCode()`;
5. la ausencia se comprueba por canal vacío;
6. aislamiento obligatorio;
7. los dos canales se capturan por separado;
8. en el código se cita por nombre, nunca por línea.

**Una nueva en 006:**

9. **Sobre la copia congelada de logs reales, sólo recuentos.** Ningún test, validación ni mensaje de
   diagnóstico imprime un `message.id`, un `requestId` o un `event_id` de un log real. Los **vectores
   de prueba** usan identificadores **sintéticos**. La copia vive fuera del repositorio y se borra al
   cerrar (quickstart §La copia congelada).

---

## Bloques de implementación

Orden propuesto y por qué difiere del de partida:
- **B0 pasa a ser el linter**, por D-006-P10;
- **«un mensaje, un evento» se parte en tres** (B2a, B2b y B2c), para que cada bloque tenga un rojo
  pequeño.

Tamaños: **S** < 50 líneas de producción · **M** 50–150 · **L** > 150 (los tests aparte).

| Bloque | Qué | Ficheros | Rojo (antes) | Verde (después) | Mutaciones que validan | Acredita | Tamaño |
|---|---|---|---|---|---|---|---|
| **B0 · Linter a 0** | Los 7 avisos (R9) | `project.go`; `endpoint.go`, `endpoint_test.go`, `enrollment.go`, `config.go`, `transport.go`; `adhesion_test.go` | `golangci-lint run` → 7 issues | 0 issues; los 297 tests verdes **sin tocar ninguna aserción** | No hay test nuevo: es un refactor que conserva la conducta. La red es la suite existente, incluidos los 5 sitios de `endpoint_test.go` que miran los dos valores de `JuzgarEndpoint` | FR-031 (lint), SC-018 (parcial; se re-mide en cada bloque) | **S** (~15 prod, ~8 test) |
| **B1 · Contratos y testigos** | Redescribir el origen del `event_id` en 001. Identificadores en los fixtures y en las 6 líneas literales. Centinelas en la denylist. Testigos de la derivación **en rojo** | `specs/001…/contracts/transport.md`, `specs/001…/data-model.md`; `testdata/claude_code_sample.jsonl`, `testdata/boundary_sample.jsonl`; `boundary_test.go`, `main_test.go`; `eventid_test.go` (nuevo) | Por `FromClaudeCodeLine`, la API existente, para que el rojo sea de **test** y no de compilación: **(1)** misma línea en dos `Context` → mismo `event_id`: ROJO (hoy aleatorio); **(2)** vector del par → `43b8…f3e2`: ROJO; **(3)** `<synthetic>` → `nil`: ROJO (hoy emite); **(4)** línea sin ningún identificador → `nil`: ROJO | Golden ampliado **verde de nacimiento**. `TestSC009_…` **sigue verde** con los fixtures editados: prueba que el baseline no se movió (R7) | Golden: copiar `message.id` a un campo del evento (decodificándolo sólo para la mutación) → ROJO; revertir por edición inversa | FR-011, FR-013; SC-004 (golden); deja en rojo SC-003a, SC-006 y SC-008c | **S** prod (sólo docs) · ~120 test |
| **B2a · La derivación** | `rawRecord` + ids; `eventid.go`; descartar `<synthetic>`; formas de un solo identificador; no emitir sin ninguno | `claudecode.go`, `eventid.go` (nuevo) | Los 4 rojos de B1 | Los 4 en verde, más: (5) las 3 formas distintas entre sí, todas de 32 hex; (6) vectores de una sola forma y de la ambigüedad | (m1) quitar el tipo del dominio → (5) ROJO · (m2) quitar el prefijo de longitud → vector de ambigüedad ROJO · (m3) volver a emitir `<synthetic>` → (3) ROJO · (m4) meter la sal en el hash → (1) ROJO | FR-002, FR-003, FR-004, FR-006 (derivar), FR-007, FR-008; SC-003a, SC-006 | **M** (~70 prod, ~100 test) |
| **B2b · La pasada** | El conjunto de la pasada, los contadores y el resumen. Instanciada en `generate()` | `pasada.go` (nuevo), `claudecode.go`, `main.go` (`generate`, `runOnce`, `tick`) | Con un tipo vacío nil-seguro creado **primero**, para que compile: **(7)** un mensaje de 3 líneas en una pasada → 1 evento: ROJO; **(8)** consumo distinto → 1 evento con el de la primera + contador = 1: ROJO; **(9)** sin identificador → contador = 1: ROJO; **(10)** el resumen no contiene ningún centinela de identificador; **(11)** `generate()` en sandbox → la cola tiene 1 evento por mensaje: ROJO | Los 5 en verde. Con la pasada a nil, una línea repetida **se emite** (patrón `Resolutor`) | (m5) desactivar el conjunto → (7) y (11) ROJO · (m6) sumar en vez de quedarse con la primera → (8) ROJO · (m7) escribir el `event_id` en el resumen → (10) ROJO · (m8) `generate()` sin instanciar la pasada → (11) ROJO **y (7) verde**: prueba que (11) mira el camino real | FR-001, FR-005, FR-006 (contar), FR-033; SC-008, SC-021 (sandbox) | **M** (~90 prod, ~180 test) |
| **B2c · Dry-run y actualización** | Pasada en `dryRun()`; línea `evento:` con las 4 partidas y el `event_id`; resumen. Actualizar sin reenviar | `main.go` (`dryRun`), `main_test.go` | **(12)** `--scan` sobre un fichero con un mensaje de 3 líneas → 1 `evento:` (proceso, `ExitCode` + recuento): ROJO; **(13)** la línea lleva las 4 partidas y `event_id=` de 32 hex: ROJO; **(14)** sandbox con `state.json` a mitad del log y un evento antiguo en cola → sólo se encola lo posterior, y el antiguo queda byte a byte | (12) y (13) en verde. (14) **nace verde** (no cambia nada del estado), y por eso necesita mutación | (m9) reiniciar el offset al actualizar → (14) ROJO · (m10) reescribir la cola al derivar → (14) ROJO | FR-009, FR-010; SC-001, SC-002 y SC-003b (V2–V4 sobre la copia), SC-007, SC-021 (V10) | **S** (~30 prod, ~120 test) |
| **B3 · Tarifas** | Las 16 claves con cabecera; test espejo | `pricing.go`, `pricing_test.go` | **(15)** recuento = 16: ROJO (3); **(16)** las 64 cifras: ROJO; **(17)** ninguna sobra; **(18)** `TestCost` a 5/25/6,25/0,50: ROJO; **(19)** coste a mano de `claude-opus-5-5` | Los 5 en verde | (m11) alterar una cifra → (16) ROJO · (m12) añadir una clave sobrante → (17) ROJO y (15) ROJO · (m13) quitar una clave → (15) y (16) ROJO | FR-014 a FR-020; SC-009, SC-010 (con V12), SC-011 (revisión) | **S–M** (~90 prod con la cabecera, ~120 test) |
| **B4 · CLI** | Fuente única de ayuda; escalera; `flag.Usage`; ayudas de subcomando; subcomando inexistente | `ayuda.go` (nuevo), `main.go`, `enroll.go`, `status.go`, `status_test.go`, `project.go`, `ayuda_test.go` (nuevo) | Por proceso, con los canales por separado: **(20)** las 4 ayudas generales idénticas, por stdout, con stderr vacío: ROJO; **(21)** contenido mínimo: ROJO (`-h` hoy no lista subcomandos); **(22)** las 8 ayudas de subcomando: 0 peticiones a `httptest` y el árbol del sandbox idéntico: ROJO (`status -h` crea el directorio; `project join -h` emite); **(23)** `enrol` → exit 1, nombrado: ROJO (hoy 0); **(24)** los 3 prefijos de secreto no se reproducen; **(25)** token centinela ausente en todo; **(26)** *(añadido el 2026-10-02, E-006-P3)* opción desconocida y opción sin valor: stdout vacío, stderr nombra la opción y remite a `permea help`, exit 2: ROJO (hoy Go no remite a `permea help` y añade su uso por defecto) | Los 7 en verde | (m14) una de las 4 ayudas por stderr → (20) ROJO · (m15) mover la comprobación de `status -h` detrás de `DataDir()` → (22) ROJO · (m16) quitar el `-h` de `projectJoin` → (22) ROJO por peticiones · (m17) reproducir lo tecleado siempre → (24) ROJO · (m18) imprimir la configuración en la ayuda de `status` → (25) ROJO · (m19) *(E-006-P3)* escribir la ayuda por stdout ante una opción desconocida → (26) ROJO | FR-021 a FR-024; SC-012 a SC-015; contrato de CLI §Opción desconocida | **M** (~140 prod, ~350 test) |
| **B5 · README, CHANGELOG y comentarios** | Instalación con `permea-dev`; primeros pasos para quien instala, con el aviso del historial; fuera el «modo de ref»; límite de casamiento; CHANGELOG `0.3.0`; comentarios `bfgnet` | `README.md`, `CHANGELOG.md` (nuevo), `release.yml`, `.goreleaser.yaml` | V18 en rojo: `grep -c bfgnet` > 0, CHANGELOG ausente | V18 en verde: los repos responden, 0 `bfgnet`, diff de configuración vacío | No aplica (documentación); V18 es la comprobación mecánica | FR-025 a FR-029; SC-016, SC-017 | **M** en texto (~150 líneas), 0 de código |
| **Cierre** | Puertas; Q-006-1; snapshot; **W1** (FR-034); PR y fusión; etiqueta anotada; **P1**; **W2** | — | — | quickstart, checklist de cierre | — | FR-030, FR-031, FR-032, FR-034; SC-018, SC-019, SC-020, SC-022 | operativo |

**Por qué este orden:**
- **B0** primero: la puerta de la constitución (D-006-P10), y así los avisos 1–4 de `project.go` no se
  mezclan con B4.
- **B1** antes que **B2**: Principio IV, el golden primero, y los fixtures arreglados antes de que el
  código los exija.
- **B2a → B2b → B2c**: derivar, deduplicar, exponer.
- **B3** es independiente de B2 y podría ir en paralelo tras B0. Se pone detrás para que SC-010 se mida
  ya con un evento por mensaje.
- **B4** después de B2c: los dos tocan `main.go`, y B2c antes evita conflictos.
- **B5** el último: documenta la CLI y el CHANGELOG ya definitivos.

**Puerta de cada bloque** (además de su verde):
- `gofmt -l .` vacío;
- `go vet ./...` limpio;
- `golangci-lint run` → 0;
- `go test ./...` → 9 paquetes ok;
- `git diff 0311fa1 -- internal/event` vacío (SC-005);
- `grep -rnE` de la disciplina 8 → nada.

---

## Trazabilidad

| Requisito | Bloque | | Criterio | Bloque |
|---|---|---|---|---|
| FR-001 | B2b | | SC-001 | B2c (V2) |
| FR-002 | B2a | | SC-002 | B2c (V3) |
| FR-003 | B1 + B2a | | SC-003 | B2a (a) · B2c (b, V4) |
| FR-004 | B2a | | SC-004 | B1 (golden) · B2c (V5) |
| FR-005 | B2b | | SC-005 | todos (puerta) |
| FR-006 | B2a (derivar) · B2b (contar) | | SC-006 | B2a · V7 |
| FR-007 | B2a | | SC-007 | B2c |
| FR-008 | B2a | | SC-008 | B2b |
| FR-009 | B2c | | SC-009 | B3 |
| FR-010 | B2c | | SC-010 | B3 + V12 |
| FR-011 | B1 | | SC-011 | B3 (revisión) |
| FR-012 | todos (puerta) | | SC-012 | B4 |
| FR-013 | B1 | | SC-013 | B4 |
| FR-014 – FR-020 | B3 | | SC-014 | B4 |
| FR-021 – FR-024 | B4 | | SC-015 | B4 |
| FR-025 – FR-029 | B5 | | SC-016, SC-017 | B5 |
| FR-030 | Cierre | | SC-018 | B0 + cada bloque + Cierre |
| FR-031 | B0 + Cierre | | SC-019 | Cierre (P1) |
| FR-032 | Cierre (W2) | | SC-020 | Cierre (W2) |
| FR-033 | B2b | | SC-021 | B2b + B2c (V10) |
| FR-034 | Cierre (W1) | | SC-022 | Cierre (W1) |

**34 / 34 requisitos y 22 / 22 criterios con bloque.**

---

## Q-006-1 — cómo la absorbe el plan sin rehacerse

`claude-sonnet-5` sólo afecta a **B3** y a la spec (`research.md` R6.3; `contracts/tarifas.md` §Cómo
se absorbe):
- **Si el dueño decide antes de B3**: B3 replica directamente el commit nuevo del catálogo.
- **Si decide después**: una tarea de Cierre toca los cinco sitios en un solo commit (fila, fila
  esperada, commit replicado en la cabecera, limitación 3, y M4/FR-014/SC-009 por enmienda fechada) y
  re-verifica SC-009 y SC-010.
- **Puerta**: no se etiqueta mientras la cabecera no cite el commit vigente del catálogo (checklist de
  cierre del quickstart).

Ningún otro bloque lee una cifra de tarifa.

---

## Riesgos

| Riesgo | Mitigación |
|---|---|
| Una derivación sutilmente mal (orden de componentes, codificación, mayúsculas) duplica o funde mensajes **en producción, en silencio** | Vectores normativos calculados con una implementación **independiente** (`contracts/event-id.md`). Mutaciones m1, m2 y m4 |
| Editar `claude_code_sample.jsonl` (entrada del baseline de 004) parece romper una regla de 004 | La regla protegía las 3 columnas y el recuento. `TestSC009_…` en verde tras la edición lo demuestra (R7). Declarado en §Dudas. *(Decidido el 2026-10-02, E-006-P1: se edita, con `TestSC009_…` verde sin tocar sus aserciones y la anotación fechada en el README del fixture.)* |
| Un mensaje repartido en dos ficheros: `--scan` (una pasada por fichero) lo contaría dos veces y SC-001 fallaría | M2: 0 casos medidos. `generate()` usa una pasada para todos los ficheros, así que la conducta real es correcta. Si V2 discrepa, primero se mira esto |
| Un mensaje que cruza la actualización cuenta dos veces, una sola vez | Residuo declarado en la spec (caso límite). No se mitiga |
| La memoria de la pasada en la primera ejecución sobre un historial grande | Estimada en 1–2 MB para 10 698 mensajes. Se mide en B2b. Clave de 16 bytes, no la cadena hex |
| Cambiar la ayuda de stderr a stdout rompe a quien la canalizaba | Lo declara el CHANGELOG (FR-028, D-006-7) |
| Windows: nada de 003–005 se ha ejecutado nunca allí (permisos y rename de `config.json`, rutas, TLS) | **W1 antes de la etiqueta** (FR-034). Si falla, no hay etiqueta |
| El «0 avisos» depende de la versión del linter | El quickstart fija 2.12.2. Si se mide con otra, se anota |
| El snapshot no lleva la versión `0.3.0` | Esperado (R10). SC-022 sólo exige «no `0.0.1-dev`»; W1 anota el commit |
| `TAP_GITHUB_TOKEN` caducado o sin permisos | **Sin comprobar.** Si falla, la release se publica y el tap y el bucket no (contrato de 002). P1 verifica los tres canales. *(Enmendado 2026-10-02, E-006-P5: el dueño lo renovó el 2026-10-01. Procedencia: el orquestador; Claude no lo ha comprobado, y P1 sigue siendo la verificación.)* |
| La plataforma verá caer el volumen por instalación (×2,13 en los datos medidos) | Lo anuncia el CHANGELOG. Un aviso del lado de la plataforma queda fuera de la spec |

---

## Fases

- **Phase 0 del plan — Research**: hecho, `research.md` R0–R11.
- **Phase 1 del plan — Diseño y contratos**: hecho: `data-model.md`, `contracts/event-id.md`,
  `contracts/cli.md`, `contracts/tarifas.md` y `quickstart.md`.
- **Phase 2 del plan — Tasks**: `/speckit.tasks` sobre este plan. Una tarea por garantía (disciplina
  1), agrupadas por los bloques B0–B5 y el Cierre, con los rojos numerados (1)–(25) y las mutaciones
  (m1)–(m18) de la tabla. *(Enmendado 2026-10-02, E-006-P3: más el rojo (26) y la mutación (m19).)*

---

## Enmiendas del 2026-10-02

Decisiones del orquestador sobre las dudas del plan, tomadas el mismo día, después de commitear el plan
(`cbee0cd`). Ningún número existente se reasigna; lo revocado se marca.

- **E-006-P1 · El fixture se edita.** `internal/ingest/testdata/claude_code_sample.jsonl` recibe
  `message.id` y `requestId` en B1. Condiciones:
  - `TestSC009_RegresionCeroDelCaminoDeIngesta` sigue verde **sin tocar sus aserciones**;
  - el README del fixture (`internal/project/testdata/README.md`, §Lo que estos fixtures NO son)
    lo anota con fecha.

  Toca: R7, riesgo 2, B1 (`tasks.md`).
- **E-006-P2 · SC-003 se mide entero.** Se autoriza `--run` **sólo** en un sandbox aislado:
  - `env -i`, con `HOME` y `XDG_CONFIG_HOME` temporales;
  - **sin enrolar**: antes, `permea status` en ese mismo entorno debe decir «no enrolado»;
  - sobre la copia congelada (`logs_root`).

  Fuera de ese caso, `--run`, `--daemon` y `enroll` siguen prohibidos a Claude. Toca: quickstart V4,
  Cierre (tramo de medidas).
- **E-006-P3 · Opción desconocida.** Error por stderr que **nombra la opción** y remite a
  `permea help`, exit 2 como hoy, **nada por stdout**. Los argumentos sobrantes tras las opciones
  siguen como están. Toca: `contracts/cli.md` (§Opción desconocida, que revoca la línea anterior), R8,
  B4 (rojo (26) y mutación m19), quickstart V14.
- **E-006-P4 · El `nolint` razonado de SA1007 queda aceptado** (R9 #7). Sin cambio en el plan.
- **E-006-P5 · `TAP_GITHUB_TOKEN` renovado por el dueño el 2026-10-01.** Procedencia: el orquestador.
  Toca: §Riesgos.
- **E-006-P6 · El secreto de enrolamiento de W1 lo prepara el dueño**: tarea ✋ suya en el Cierre.
  Toca: quickstart W1, `tasks.md`.

---

## Complexity Tracking

*Vacío: el Constitution Check no registra violaciones.*
