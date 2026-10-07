# Tasks: 008 · «Lector de Codex»

**Feature**: `008-lector-codex` · **Fecha**: 2026-10-07 · [spec.md](./spec.md) *(ratificada, E-1, E-2)* · [plan.md](./plan.md) ·
[contracts/event-id-codex.md](./contracts/event-id-codex.md) · [quickstart.md](./quickstart.md)
**Base**: `7b8c77c`. **Línea base a preservar**: `go test -count=1 ./...` → **9 paquetes ok, 494 pass, 0 fail**; `golangci-lint run` → **0**.

## Formato: `[ID] [P?] [✋?] Descripción`

- **[P]**: sin dependencia de decisión entre sí. **No** quiere decir que se editen a la vez: dos `[P]` sobre el mismo fichero se escriben
  una detrás de otra.
- **✋**: la ejecuta **el dueño**: commits, la PR, la fusión, la etiqueta y los ensayos en Windows. Claude no hace git de escritura. El
  mensaje de cada commit va previsto, **sin tildes ni ñ**.
- **(1)…(37)** son los rojos y **(m1)…(m27)** las mutaciones de `plan.md` §Bloques, con la misma numeración. Lo propio de este fichero se
  llama **M-B1a**, **M-B2a**…**M-B2e**, **M-B3a**, **M-B3b**, **M-B4a**, **M-B5a**…**M-B5c**, **M-B6a** y **M-B6b**.
- **No se renumera**: una tarea añadida después recibe el siguiente número libre.
- Las transcripciones van a `soporte/registro.md`, que se crea en B1, y en cada tarea queda una remisión.

## Disciplinas y protocolo

- **Disciplinas 1–9 de 006**, la **10** de 007 *(las copias se leen en su sitio)* y la **11** de este plan *(fixtures sólo sintéticos)*.
- **Protocolo de mutación** *(006, con E-5 de 007)*:
  1. el censo se declara **en la tarea, ANTES de mutar**, con las co-caídas, **por HOJA (subtest), nunca por test padre**;
  2. se muta y se ejecuta `go test -count=1 ./... 2>&1`;
  3. si el conjunto de `FAIL` **coincide** con lo declarado, se revierte **por edición inversa**, se verifica con **md5** que el fichero
     vuelve a ser el de antes y se transcribe el fallo;
  4. si **no** coincide, la mutación **se deja puesta y se para**;
  5. una mutación que no compila o que panica no cuenta como superada.
- **Rojo antes de verde**, transcribiendo la razón real. **Un test que nace verde lleva su mutación declarada.**
- **Ningún `_test.go` existente se toca** *(censo: ninguno)*. Sólo cambian `internal/testutil/sandbox.go` *(M-9)*, `cmd/permea/main.go`,
  `README.md` y `CHANGELOG.md`, y entran ficheros nuevos.

**Puertas de cada bloque**, antes de su ✋ commit:
```sh
gofmt -l .                     # → vacío
go vet ./...                   # → sin hallazgos
golangci-lint run              # → 0 issues (2.12.2, sin tope)
go test -count=1 ./...         # → 9 paquetes ok; 494 + los nuevos
git diff 7b8c77c -- internal/event/ internal/ingest/eventid.go internal/ingest/eventid_test.go \
  internal/ingest/boundary_test.go specs/006-medicion-fiel/contracts/event-id.md      # → vacío (FR-022)
git diff 7b8c77c --name-only --diff-filter=M -- '*_test.go'                            # → vacío (M-8)
```

---

## B0 · Documentos

- [x] **T001** Fase 0. Transcribir la línea base sobre `7b8c77c`: 494 pass en 9 paquetes, lint 0 y la frontera sin diff. Si algo difiere,
  **se para**.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B0 · T001.
- [x] **T002** ✋ *(`5f801c5`)* Commit: `008 B0: spec ratificada (E-1, E-2), contrato, plan, tareas y quickstart`.

## B1 · Identidad *(`contracts/event-id-codex.md`; FR-007; SC-006)*

- [x] **T003** Fase 0: en `internal/ingest/codex_eventid.go` *(nuevo)*, `derivarEventIDCodex(responseID string) (string, bool)`, que
  devuelve `"", false`. La suite sigue verde.
- [x] **T004** **Rojos (1), (2) y (3)** en `internal/ingest/codex_eventid_test.go` *(nuevo)*:
  - **(1)**: los dos vectores del contrato, byte a byte;
  - **(2)**: el vector del espacio de Claude Code, como literal *(`128d67bd…`)*, es distinto del de Codex para el mismo valor;
  - **(3)**: `response_id` vacío → `ok = false`.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T004 *((3) nace verde; la valida m3)*.
- [x] **T005** **Verde**: la derivación del contrato. Reutiliza `hashEventID` sólo si su firma lo permite **sin tocar `eventid.go`**; si
  no, se replica en el fichero nuevo *(M-1)*.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T005 *(replicada: `hashEventID` fija `claude_code`)*.
- [x] **T006** Censo de mutaciones, **declarado antes de mutar**, por hoja: **(m1)** `"codex"` → `"claude_code"` → (1) ×2; **(m2)** sin
  prefijo de longitud → (1) ×2 *(no (2): compara con el vector de 006)*; **(m3)** vacío aceptado → (3). **Remate** *(Encargo 5)*:
  **M-B1a**, derivar en el espacio de Claude Code → (1) ×2 y (2).
- [x] **T007** Mutaciones y transcripción, con md5.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T006 *(m2 sin la co-caída de (2), declarado antes de mutar)*, T007 y
  > §Remate de B1 · M-B1a.
- [x] **T008** ✋ *(`7800d3b`)* Puertas y commit: `008 B1: event_id de Codex con espacio de nombres propio`.

## B2 · Una línea *(FR-006, FR-009, FR-010, FR-012, FR-013, FR-015, FR-016, FR-027; SC-004, SC-008)*

- [x] **T009** Fase 0: en `internal/ingest/codex.go` *(nuevo)*, `ContextoCodex` *(el `Context` de Claude Code y `DelTurno`, que da el
  modelo y el `cwd` del turno)*, `ClaseCodex` y `LineaCodex(line, ctx)`, que no clasifica nada. La suite sigue verde.
- [x] **T010** Fixtures **sintéticos** en `internal/ingest/testdata/codex/`, con identificadores `r-0000…` y partidas inventadas o
  publicadas en `descubrimiento.md` *(13 831 / 11 008 / 0 / 5)*.
- [x] **T011** [P] **Rojos (4) a (10)** en `internal/ingest/codex_linea_test.go` *(nuevo)*: **(4)** 13 831 / 11 008 / 0 / 5 → 2 823 / 0 /
  11 008 / 5, y la hoja `otra_linea_no_es_registro` *(FR-006; nace verde, la valida M-B2b)*; **(5)** SC-008, 100 / 40 / 60 / 10 → 0 / 60 /
  40 / 10; **(6)** incoherente, con tres hojas *(E-3: caché + escritura > entrada · sin `usage` · partida negativa)*; **(7)** sin
  `response_id`, o vacío; **(8)** `tool`, coste 0 y `cost_available = false`; **(9)** `occurred_at` = la marca; **(10)** `session_ref` =
  `event.Ref(sal, session_id)`, con `session_id` ≠ `thread_id` ≠ `turn_id`.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B2 · T011.
- [x] **T012** **Verde**: `LineaCodex` decodifica primero `{type, timestamp, payload}` en crudo, y sólo un registro decodifica su payload.
  Aplica FR-010 y la clasificación de D-008-P9 con E-3, y construye el evento con `event.Ref` y el resolutor. **No** llama a `internal/pricing`.
- [x] **T013** Verde de nacimiento, escrito tras T012: **(FR-016)** el evento no contiene ningún centinela del fixture. Lo valida M-B2a.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B2 · T012 y T013.
- [x] **T014** Censo, **antes de mutar**, por hoja: **(m4)** sin restar la escritura → (5); **(m5)** `cost_available = true` → (8);
  **(m6)** el incoherente se emite *(declarada sin pánico: «sin `usage`» se toma a cero)* → (6) ×3; **(m7)** `session_ref` del `thread_id`
  → (10); **(m8)** `occurred_at` = ahora → (9); **M-B2a** `session_ref` sin sal → (10) y T013; **M-B2b** sin la guarda del tipo → (4/otra_linea).
- [x] **T015** Mutaciones y transcripción, con md5.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B2 · T014 y T015 *(las siete coinciden)*, y §Remate de B2 *(E-4: cuatro
  > hojas en (6), la partida ausente, M-B2c, M-B2d y M-B2e)*.
- [x] **T016** ✋ *(`ed97b2e`)* Puertas y commit: `008 B2: una linea de Codex es un evento sin coste`.

## B3 · Contexto y estado *(FR-004, FR-005, FR-008, FR-011, FR-014, FR-017, FR-018, FR-027, FR-029; SC-005, SC-007, SC-009, SC-010, SC-017)*

- [x] **T017** Fase 0: en `internal/ingest/codex_contexto.go` *(nuevo)*, `PasadaCodex` con los ocho recuentos y `Resumen()` *(texto
  aprobado)*; `LeerFicheroCodex(st, ruta, base, p, avisos io.Writer)`, que usa `Recorrer` con `fijar` = lo leído y aún no emite; y
  `ContarComprimido(st, ruta, p)`. Compila, y la suite sigue verde.
- [x] **T018** Fixtures sintéticos *(21, en `testdata/codex/contexto/`)*: dos turnos; un ajuste de modelo **dentro** de un turno *(para m9)*;
  una compactación; sin modelo; formato anterior, mixto y sin consumo; una reanudación y un corte en dos mitades; una bifurcación; la
  cuenta de SC-017; `cwd` y sin `cwd`; una línea corrupta; un fichero largo y su versión truncada; un `.zst`.
- [x] **T019** [P] **Rojos (11) a (20)** en `internal/ingest/codex_contexto_test.go` *(nuevo)*: **(11)** el modelo del turno, no el vigente;
  **(12)** la compactación, el vigente; **(13)** sin modelo, vacío y contado; **(14)** SC-010, dos pasadas = una; **(15)** SC-005, la
  reanudación en dos pasadas da 1 + 3, iguales a una; **(16)** SC-007, un evento y una repetida; **(17)** FR-017, anterior, mixto y sin
  consumo *(nace verde; la valida M-B3b)*; **(18)** FR-018, el `.zst` cuenta 1, 0 y 1 *(la segunda nace verde; la valida m12)*; **(19)**
  SC-017, el resumen literal y la identidad de FR-027; **(20)** FR-014, el `cwd` del turno, el de `session_meta` y, *(E-4)*, sin `cwd` →
  `project_ref` vacío. **(35)** *(Encargo 6, nuevo)*: un fichero truncado se relee sin prefijo.
- [x] **T055** *(E-2)* **Rojo (34)**, FR-029, en `codex_contexto_test.go`, con el aviso hacia un `io.Writer` inyectado: una línea corrupta en
  la parte nueva da **un** aviso `skip (línea corrupta): …`, no cuenta en «respuestas», y el registro siguiente sale. En la pasada
  siguiente, con la línea ya en el prefijo, **ningún** aviso.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B3 · T019, T055 y (35).
- [x] **T020** **Verde**: el prefijo con el filtro de bytes *(D-008-P1)*, y el contexto también con la parte nueva; el modelo de FR-011 y el
  `cwd` de FR-014; el formato de FR-017, sobre el fichero entero; la clasificación de D-008-P9, con el conjunto de `event_id` de la pasada;
  el `.zst` como entrada de `state.json` *(D-008-P2)*; y la línea corrupta de FR-029, con su aviso sólo en la parte nueva *(D-008-P11)*.
  `state.json` sigue con cuatro campos. Medida informativa de SC-014: ~0,25 s.
- [x] **T021** Censo, **antes de mutar**, por hoja: **(m9)** siempre el vigente → (11/ajuste); **(m10)** sin prefijo → (14) y (15/segunda);
  **(m11)** sin repetidas → (16) y (19); **(m12)** el `.zst` en cada pasada → (18/segunda); **(m13)** formato anterior por «hay
  `token_count`» → (17/mixto); **(m14)** `cwd` siempre de `session_meta` → (20/del_turno); **(m26)** la línea corrupta corta el fichero →
  (34) ×2; **(m27)** el aviso también en el prefijo → (34/segunda); **M-B3a** el truncado lee prefijo → (35); **M-B3b** sin registros =
  formato anterior → (17/sin_consumo).
- [x] **T022** Mutaciones y transcripción, con md5.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B3 · T020, T021, T022 *(las diez coinciden)*, la medida de SC-014 y FR-029.
- [x] **T023** ✋ *(`82da9d6`)* Puertas y commit: `008 B3: contexto entre pasadas, formato y recuentos de Codex`.

## B4 · Raíz y activación *(FR-001, FR-002; M-9)*

- [x] **T024** Fase 0: `config.CodexSessionsRoot() (string, error)` en `internal/config/codex.go` *(nuevo)*, que devuelve `"", nil`.
  *(Antes, la Fase previa del Encargo 7: la forma de los fixtures frente a la copia congelada; cuadra. Registro §B3, «comprobación de forma».)*
- [x] **T025** [P] **Rojos (21) y (22)** en `internal/config/codex_test.go` *(nuevo)*, con `t.Setenv`: **(21)** tres subtests, `CODEX_HOME`
  definida → `<valor>/sessions`, vacía → `<home>/.codex/sessions`, ausente → lo mismo; **(22)** *(E-2)* una raíz inexistente se devuelve
  igual, sin error *(la activación se decide en cada pasada, FR-002, (32))*.
- [x] **T026** **Verde**: el único `os.Getenv` de producción *(D-008-P7)*. En `internal/testutil/sandbox.go`, `t.Setenv("CODEX_HOME", "")`
  *(M-9)*.
- [x] **T027** Comprobar y transcribir, sin tocarlo, que `sandbox_test.go` sigue verde. Y `grep -rn 'os.Getenv' --include=*.go cmd internal
  | grep -v _test` → 1.
- [x] **T028** Censo, **antes de mutar**, y mutaciones: **(m15)** `CODEX_HOME=""` tomada como raíz → (21/vacía); **(m16)** ignorar
  `CODEX_HOME` → (21/definida); **M-B4a** exigir que la raíz exista → (22).
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B4 · T024–T028 *(las tres coinciden)*.
- [x] **T029** ✋ *(`c1068d6`)* Puertas y commit: `008 B4: raiz de Codex con CODEX_HOME`.

## B5 · Integración *(FR-002, FR-003, FR-016, FR-019, FR-021, FR-025, FR-026, FR-028; SC-001–SC-003 por fixture, SC-006, SC-011, SC-015, SC-016, SC-018)*

- [x] **T030** Fase 0, en este orden:
  1. **Antes de tocar `main.go`** *(D-008-P12)*: la referencia de SC-011, con el binario de `HEAD`, en
     `cmd/permea/testdata/codex/referencia-run.stderr` *(md5 `17f189f0…`)*.
  2. **El forzado de SC-015** *(D-008-P10)*: **el A**. `Load` pasa, `Append` pasa y `Save` da `permission denied`; uid 1000, sin `t.Skip`.
  3. En `main.go`, `agent.codexRaiz`, resuelta en `setup()`, y `agent.codex`; `generate()` con su firma de hoy *(D-008-P4/P5)*.
- [x] **T031** [P] **Rojos (23) a (28)** en `cmd/permea/codex_test.go` *(nuevo)*: **(23)** SC-011, con `CODEX_HOME` vacía y con la raíz
  inexistente, byte a byte la referencia *(nace verde; lo valida m18)*; **(24)** SC-016, sin Claude Code; **(25)** SC-015, con el forzado A;
  **(26)** la línea `codex:` literal tras los resúmenes, y «4 eventos encolados»; **(27)** el predicado y `tick`; **(28)** la segunda `--run`
  da 0. **(36)** *(nuevo, Encargo 8)*: `ListarCodex`, en `codex_contexto_test.go`.
- [x] **T032** **Verde**: `generarCodex` tras Claude Code y antes del único `st.Save` *(D-008-P6)*; la existencia de la raíz en cada pasada
  *(FR-002)*; los errores por fichero de FR-028 *(aviso, estado sin tocar y siguiente; `Append` fatal, D-008-P11)*; la línea en `runOnce` y,
  con `HayNovedades`, en `tick`. `ListarCodex` y `HayNovedades`, en `codex_contexto.go`. **568 pass, 0 SKIP.**
- [x] **T033** Comprobar y transcribir, sin tocarlos, que `main_test.go`, `retencion_test.go`, `coste_test.go` y `project_test.go` siguen
  verdes.
- [x] **T034** **Rojo (31)**, SC-006: centinelas en `response_id`, `session_id`, `turn_id` **y el nombre del fichero**. Tras `--run`: ninguno
  en la cola; en `state.json`, sólo el del nombre y **dentro de su ruta** *(FR-016)*. Cae primero en su precondición.
- [x] **T056** *(E-2)* **Rojo (32)**, FR-002: `setup()` en el sandbox con `CODEX_HOME` a una carpeta que **no existe**; la primera pasada no
  emite; se crea la carpeta; la segunda **del mismo `agent`** emite.
- [x] **T057** *(E-2)* **Rojo (33)**, SC-018 y FR-028: un fichero de Codex con permisos 000 junto a uno sano y a Claude Code. Salen los
  sanos, el aviso `codex: fichero omitido: …`, `state.json` se guarda, la segunda pasada no reencola Claude Code *(nace verde; la valida m24)*,
  y `se_relee` saca sus registros.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B5 · T030–T034, T056, T057 y (36).
- [x] **T035** Censo, **antes de mutar**, por hoja: **(m17)** `st.Save` antes de encolar Codex → (25); **(m18)** la línea siempre → (23) ×2;
  **(m19)** `tick` sin predicado → (27/tick); **(m20)** Codex sólo con logs de Claude Code → (24), (25), (27/tick), (28), (31) y (32);
  **(m23)** la existencia sólo en `setup()` → (32); **(m24)** el error aborta la pasada → (33/pasada, /segunda_pasada); **(m25)** el omitido
  guarda su offset → (33/se_relee); **M-B5a** el turno en `state.json` → (31); **M-B5b** sin `.zst` → (36); **M-B5c** `HayNovedades` sólo con
  respuestas → (27/predicado).
- [x] **T036** Mutaciones y transcripción, con md5.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B5 · T035, T036 *(las diez coinciden)* y R-9.
- [x] **T037** ✋ *(`d350165`)* Puertas y commit: `008 B5: Codex en run y daemon, antes de guardar el estado`.

## B6 · `--scan` *(FR-020, FR-021; SC-013)*

- [x] **T038** [P] **Rojos (29) y (30)** en `cmd/permea/codex_test.go`, con los textos leídos de la spec por programa: **(29)** una línea
  `evento:` aprobada por evento, la línea `codex:` y nada en disco *(`nada_en_disco` nace verde; la valida M-B6b)*; **(30)** Claude Code →
  byte a byte la referencia del binario anterior *(nace verde; la valida M-B6a)*. **(37)** *(nuevo)*: formato anterior → 0 eventos y la cuenta.
- [x] **T039** **Verde**: `dryRun` lee la primera línea con el mismo `Scanner` de 1 MiB y bifurca a `dryRunCodex`, con `LeerFicheroCodex`
  sobre un estado en memoria *(D-008-P8)*.
- [x] **T040** Censo, **antes de mutar**, y mutaciones: **(m21)** `cw5m=`/`cw1h=` en la línea de Codex → (29/lineas_evento); **(m22)** sin
  detección → (29/lineas_evento, /resumen) y (37); **M-B6a** todo fichero es Codex → (30) y los cuatro `TestScan_*` de Claude Code;
  **M-B6b** `dryRunCodex` guarda su estado → (29/nada_en_disco).
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B6 y B7 · T038–T040 *(las cuatro coinciden)*.
- [ ] **T041** ✋ Puertas y commit: **un solo commit con B7** *(decisión del orquestador, Encargo 9)*, ver T045.

## B7 · README y CHANGELOG *(FR-023, FR-024; SC-013)*

- [x] **T042** **Rojo**, transcrito: `grep -c '^## 0.5.0' CHANGELOG.md` → 0; `grep -c '^### Codex CLI' README.md` → 0.
- [x] **T043** `README.md`: la sección aprobada, literal y **sacada por programa**, al final de «Modos de ejecución», tras lo de Claude Code.
- [x] **T044** `CHANGELOG.md`: `## 0.5.0 — PENDIENTE` encima de la 0.4.0, con el cuerpo **sacado por programa** de spec §Textos aprobados.
  `cmp` de los dos textos *(README y CHANGELOG)*, sin diferencias. `grep -c PENDIENTE CHANGELOG.md` → 1.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B6 y B7 · T042–T044.
- [x] **T045** ✋ *(`9516a08`)* Puertas y commit de B6 y B7: `008 B6 y B7: scan de Codex, README y CHANGELOG de la 0.5.0`.

## Cierre — en tramos, uno por mensaje *(plan §Cierre; si uno falla, se para y se rehace desde C1)*

- [x] **T046** **C1** · Puertas, transcritas: las del bloque, más `go test` → **494 + nuevos** *(SC-012)*; ningún `_test.go` existente
  modificado; `os.Getenv` de producción = 1; compilan Windows y darwin; `PENDIENTE` → 1.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §Cierre · E-5 y C1 *(574 pass, 0 SKIP; todo en verde)*.
- [x] **T047** **C2** · Medidas *(quickstart §Copia, §Contador, §M, §Coste)*: la huella de la copia antes; el contador; en sandbox, `--scan`
  y `--run` dos veces *(SC-001 a SC-005 y SC-009)*; **SC-014**, tres veces, con su tiempo; la huella después, igual; temporales borrados.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §Cierre · C2 *(SC-001–SC-005 y SC-009 cuadran; SC-014: 0,340 · 0,340 · 0,342 s)*.
- [x] **T048** **C3** · `goreleaser release --snapshot --clean`, el SHA-256 del zip de Windows y `strings` con los textos aprobados *(SC-013)*.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §Cierre · C3 *(`0.4.0-SNAPSHOT-9516a08`; zip de Windows `96dfa0ee…`)*.
- [x] **T049** ✋ **C4 · W1** *(quickstart §W1; sandbox, sin enrolar)*: `--version`; `status` → «no enrolado»; dos `--run` sobre la carpeta
  real de Codex, en sólo lectura; la raíz, la de Codex *(Q-6)*.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §Cierre · C4 *(2026-10-07, sin fallos: 7 eventos codex, 4 en formato anterior; la segunda, 0)*.
- [x] **T050** **C5** · El cuerpo del PR y la fecha del encabezado del CHANGELOG *(`PENDIENTE` → 0; `## 0.5.0 — 2026-10-07`)*. ✋ Commit
  `008 C5: fecha de la 0.5.0 en el CHANGELOG`.
- [x] **T051** ✋ **C6** · Fusión del PR con merge commit *(PR #5, `40bb2c1`)*.
- [x] **T052** ✋ **C7** · `git tag -a v0.5.0` sobre `main` y `git push origin v0.5.0` *(sobre `40bb2c1`; `release` en `success`)*.
- [x] **T053** **C8** · Los tres canales en `0.5.0` y `strings` del binario publicado *(SC-013)*. *(El orquestador; `gh release view`,
  la etiqueta y las sumas, comprobados en el Encargo 12.)*
- [x] **T054** ✋ **C9 · W2** *(quickstart §W2)*: `scoop update`, `--run` y, en la plataforma, eventos `codex` = las respuestas nuevas del
  contador.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §Cierre · C6–C9 *(W2, 2026-10-07: 7 eventos codex = SC-001; sólo lo nuevo, +1 con 2828)*.

## Dependencias

- B0 → B1 → B2 → B3 → B4 → B5 → B6 → B7 → C1…C9, en serie: B2 usa la identidad; B3, la línea; B5, el contexto y la raíz; B6, la línea y
  el contexto.
- Dentro de un bloque, los rojos `[P]` van antes del verde, y el censo antes de mutar.

## Cobertura

| FR | Tareas | | SC | Tareas |
|---|---|---|---|---|
| 001, 002 | T025 (21, 22), T056 (32) | | 001 | T031 *(fixture)*, T047 |
| 003 | T031 (23, 24, 36), T019 (18) | | 002 | T019 (17), T047 |
| 004 | T019 (14, 18, 35), T020 | | 003 | T011 (4), T047 |
| 005 | T019 (14, 15), T047 *(SC-014)* | | 004 | T011 (8), T047 |
| 006 | T011 (4) | | 005 | T019 (15), T031 (28), T047 |
| 007 | T004 (1–3) | | 006 | T004, T034 (31) |
| 008 | T019 (16) | | 007 | T019 (16) |
| 009 | T011 (7) | | 008 | T011 (5) |
| 010 | T011 (4, 5) | | 009 | T019 (11–13), T047 |
| 011 | T019 (11–13) | | 010 | T019 (14) |
| 012 | T011 (9) | | 011 | T030 *(referencia)*, T031 (23) |
| 013 | T011 (8) | | 012 | puertas de cada bloque, T046 |
| 014 | T019 (20) | | 013 | T031 (26), T038 (29), T044, T048, T053 |
| 015 | T011 (10) | | 014 | T047 |
| 016 | T013, T034 (31) | | 015 | T031 (25) |
| 017 | T019 (17), T038 (37) | | 016 | T031 (24) |
| 018 | T019 (18) | | 017 | T019 (19) |
| 019 | T031 (26, 27) | | | |
| 020 | T038 (29, 37) | | | |
| 021 | T031 (23), T038 (30) | | | |
| 022 | puertas de cada bloque, T046 | | | |
| 023 | T044, T046 | | | |
| 024 | T043, T044 | | | |
| 025 | T031 (25) | | | |
| 026 | T031 (24) | | | |
| 027 | T011 (6), T019 (19) | | 018 | T057 (33) |
| 028 | T057 (33) | | | |
| 029 | T055 (34) | | | |

**29 / 29 FR y 18 / 18 SC con tarea.**

## Mutaciones

| # | Bloque | Mutación | Debe caer *(y sólo eso)* |
|---|---|---|---|
| m1 | B1 | `"codex"` → `"claude_code"` | (1) ×2 |
| m2 | B1 | sin prefijo de longitud | (1) ×2 |
| m3 | B1 | `response_id` vacío aceptado | (3) |
| M-B1a | B1 | derivar en el espacio de Claude Code | (1) ×2, (2) |
| m4 | B2 | `tokens_input` sin restar la escritura | (5) |
| m5 | B2 | `cost_available = true` | (8) |
| m6 | B2 | el incoherente se emite *(«sin `usage`» a cero)* | (6) ×3 *(E-3)* |
| m7 | B2 | `session_ref` del `thread_id` | (10) |
| m8 | B2 | `occurred_at` = ahora | (9) |
| M-B2a | B2 | `session_ref` = `session_id` sin sal | (10), T013 |
| M-B2b | B2 | sin la guarda del tipo | (4/otra_linea_no_es_registro) |
| M-B2c | B2 | *(E-4)* sin fecha válida devuelve error | (6/sin_timestamp), (6/timestamp_mal_formado) |
| M-B2d | B2 | *(E-4)* una partida ausente es incoherente | `PartidaAusenteValeCero` |
| M-B2e | B2 | *(E-4)* el payload que no decodifica devuelve error | (6/payload_no_decodifica) |
| m9 | B3 | siempre el modelo vigente | (11/ajuste_dentro_del_turno) |
| m10 | B3 | sin prefijo *(empezar en el offset)* | (14), (15, segunda pasada) |
| m11 | B3 | sin repetidas | (16), (19) |
| m12 | B3 | el `.zst` en cada pasada | (18, segunda) |
| m13 | B3 | formato anterior por «hay `token_count`» | (17, mixto) |
| m14 | B3 | `cwd` siempre de `session_meta` | (20) |
| m26 | B3 | *(E-2)* la línea corrupta corta el fichero | (34) ×2 |
| m27 | B3 | *(E-2)* el aviso también en el prefijo | (34/segunda pasada) |
| M-B3a | B3 | el truncado lee prefijo | (35) |
| M-B3b | B3 | sin registros = formato anterior | (17/sin_consumo) |
| m15 | B4 | `CODEX_HOME=""` tomada como raíz | (21/vacía) |
| m16 | B4 | ignorar `CODEX_HOME` | (21/definida) |
| M-B4a | B4 | exigir que la raíz exista | (22) |
| m17 | B5 | `st.Save` antes del Append de Codex | (25), con el forzado A |
| m18 | B5 | la línea de Codex siempre | (23) ×2 |
| m19 | B5 | `tick` escribe siempre | (27/tick) |
| m20 | B5 | Codex sólo si hay logs de Claude Code | (24), (25), (27/tick), (28), (31), (32) |
| M-B5a | B5 | el contexto del turno en `state.json` *(con el `turn_id`)* | (31) |
| M-B5b | B5 | `ListarCodex` sin `.zst` | (36) |
| M-B5c | B5 | `HayNovedades` sólo con respuestas | (27/predicado) |
| m23 | B5 | *(E-2)* la existencia sólo en `setup()` | (32) |
| m24 | B5 | *(E-2)* el error de Codex aborta la pasada | (33/pasada), (33/segunda_pasada) |
| m25 | B5 | *(E-2)* el fichero omitido guarda su offset al final | (33/se_relee) |
| m21 | B6 | `cw5m=`/`cw1h=` en la línea de Codex | (29/lineas_evento) |
| m22 | B6 | sin detección de Codex en `--scan` | (29/lineas_evento), (29/resumen), (37) |
| M-B6a | B6 | todo fichero es Codex | (30) y los cuatro `TestScan_*` de Claude Code |
| M-B6b | B6 | `dryRunCodex` guarda su estado | (29/nada_en_disco) |

**41 mutaciones previstas.** Las co-caídas que aparezcan al declarar el censo se escriben **antes** de mutar, por hoja.

## Lo que este plan de tareas NO hace

- No toca `internal/event`, la derivación del `event_id` de 006, `boundary_test.go`, `internal/state/` ni ningún `_test.go` existente.
- No cambia la plataforma. Las tarifas de Codex van en un encargo aparte y posterior *(D-3)*.
- No lanza `enroll`, `--run` ni `--daemon` sobre la instalación real. Los `--run` de los tests van en sandbox, y los de W1 y W2 los lanza
  el dueño.
