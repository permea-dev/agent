# Tasks: 008 · «Lector de Codex»

**Feature**: `008-lector-codex` · **Fecha**: 2026-10-07 · [spec.md](./spec.md) *(ratificada, E-1, E-2)* · [plan.md](./plan.md) ·
[contracts/event-id-codex.md](./contracts/event-id-codex.md) · [quickstart.md](./quickstart.md)
**Base**: `7b8c77c`. **Línea base a preservar**: `go test -count=1 ./...` → **9 paquetes ok, 494 pass, 0 fail**; `golangci-lint run` → **0**.

## Formato: `[ID] [P?] [✋?] Descripción`

- **[P]**: sin dependencia de decisión entre sí. **No** quiere decir que se editen a la vez: dos `[P]` sobre el mismo fichero se escriben
  una detrás de otra.
- **✋**: la ejecuta **el dueño**: commits, la PR, la fusión, la etiqueta y los ensayos en Windows. Claude no hace git de escritura. El
  mensaje de cada commit va previsto, **sin tildes ni ñ**.
- **(1)…(34)** son los rojos y **(m1)…(m27)** las mutaciones de `plan.md` §Bloques, con la misma numeración. Lo propio de este fichero se
  llama **M-B2a**, **M-B5a** y **M-B6a**.
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
- [ ] **T002** ✋ Commit: `008 B0: spec ratificada (E-1, E-2), contrato, plan, tareas y quickstart`.

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
  prefijo de longitud → (1) ×2 y (2); **(m3)** vacío aceptado → (3).
- [x] **T007** Mutaciones y transcripción, con md5.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T006 *(m2 sin la co-caída de (2), declarado antes de mutar)* y T007.
- [ ] **T008** ✋ Puertas y commit: `008 B1: event_id de Codex con espacio de nombres propio`.

## B2 · Una línea *(FR-006, FR-009, FR-010, FR-012, FR-013, FR-015, FR-016, FR-027; SC-004, SC-008)*

- [ ] **T009** Fase 0: en `internal/ingest/codex.go` *(nuevo)*, `ContextoCodex` *(sal, máquina, dev, org, versión, resolutor, `cwd` y
  modelo del turno)*, `ClaseCodex` *(evento, sin identificador, incoherente)* y `LineaCodex(line, ctx)`, que devuelve `nil`. La suite sigue
  verde.
- [ ] **T010** Fixtures **sintéticos** en `internal/ingest/testdata/codex/`, con identificadores `r-0000…` y partidas inventadas o
  publicadas en `descubrimiento.md` *(13 831 / 11 008 / 0 / 5)*.
- [ ] **T011** [P] **Rojos (4) a (10)** en `internal/ingest/codex_linea_test.go` *(nuevo)*:
  - **(4)**: 13 831 / 11 008 / 0 / 5 → `tokens_input` 2 823, `tokens_cache_read` 11 008, `tokens_output` 5;
  - **(5)**, SC-008: 100 / 40 / 60 / 10 → 0 / 60 / 40 / 10;
  - **(6)**: caché + escritura > entrada → clase «incoherente» y sin evento;
  - **(7)**: sin `response_id`, o vacío → «sin identificador» y sin evento;
  - **(8)**: `tool = "codex"`, `cost_usd = 0` y `cost_available = false`;
  - **(9)**: `occurred_at` = la marca de la línea;
  - **(10)**: `session_ref` = `event.Ref(sal, session_id)`, con `session_id` ≠ `thread_id` en el fixture.
- [ ] **T012** **Verde**: `LineaCodex` decodifica `type`, `timestamp` y `payload.{response_id, session_id, turn_id, usage}`. Aplica
  FR-010 y la clasificación de D-008-P9, y construye el evento con `event.Ref` y el resolutor. **No** llama a `internal/pricing`.
- [ ] **T013** Verde de nacimiento, con su mutación: **(FR-016)** campo a campo, el evento no contiene ni `response_id` ni `session_id`
  ni `turn_id` *(centinelas en el fixture)*.
- [ ] **T014** Censo, **antes de mutar**, por hoja:
  - **(m4)** sin restar la escritura → (5);
  - **(m5)** `cost_available = true` → (8);
  - **(m6)** el incoherente se emite → (6);
  - **(m7)** `session_ref` del `thread_id` → (10);
  - **(m8)** `occurred_at` = ahora → (9);
  - **M-B2a** `session_ref` = `session_id` sin sal → (10) y T013.
- [ ] **T015** Mutaciones y transcripción, con md5.
- [ ] **T016** ✋ Puertas y commit: `008 B2: una linea de Codex es un evento sin coste`.

## B3 · Contexto y estado *(FR-004, FR-005, FR-008, FR-011, FR-014, FR-017, FR-018, FR-027, FR-029; SC-005, SC-007, SC-009, SC-010, SC-017)*

- [ ] **T017** Fase 0: en `internal/ingest/codex_contexto.go` *(nuevo)*, `PasadaCodex` con los ocho recuentos y `Resumen()` *(texto
  aprobado)*; `LeerFicheroCodex(st *state.Store, ruta string, …)`, que usa `Recorrer` con `fijar` = lo leído y aún no emite; y
  `ContarComprimido(st, ruta)`. Compila, y la suite sigue verde.
- [ ] **T018** Fixtures sintéticos: dos turnos con modelos distintos; `thread_settings_applied` cambiando de modelo **dentro** de un turno
  *(para m9)*; una compactación sin `turn_context`; un fichero sin ningún modelo; un formato anterior *(sólo `token_count`)* y uno mixto;
  una reanudación cortada en dos mitades; dos ficheros con el mismo registro *(bifurcación)*; una línea corrupta; un `.zst` cualquiera.
- [ ] **T019** [P] **Rojos (11) a (20)** en `internal/ingest/codex_contexto_test.go` *(nuevo)*:
  - **(11)**: modelo del `turn_context` del turno, no el vigente;
  - **(12)**: la compactación lleva el vigente;
  - **(13)**: sin modelo → `model` vacío y «sin modelo» = 1;
  - **(14)**, SC-010: dos pasadas cortadas tras el `turn_context` dan el mismo modelo, `project_ref` y `session_ref` que una;
  - **(15)**, SC-005: la reanudación en dos pasadas da 1 + 3 eventos, sin repetir;
  - **(16)**, SC-007: un evento y «repetidas» = 1;
  - **(17)**, FR-017: el formato anterior da 0 eventos y 1 fichero; el mixto da sus eventos y 0 ficheros; sin consumo, 0 y 0;
  - **(18)**, FR-018: el `.zst` cuenta 1 en la primera pasada, 0 en la segunda y 1 si cambia su tamaño;
  - **(19)**, SC-017: `respuestas 5 · eventos 2 · repetidas 1 · sin identificador 1 · incoherentes 1 · sin modelo 1`, y la identidad de
    FR-027;
  - **(20)**, FR-014: el `cwd` del turno, y si falta, el de `session_meta`.
- [ ] **T055** *(E-2)* **Rojo (34)**, FR-029, en `codex_contexto_test.go`, con el aviso hacia un `io.Writer` inyectado: una línea que pasa
  el filtro y no decodifica, en la parte nueva, da **un** aviso `skip (línea corrupta): …`, no cuenta en «respuestas», y el registro
  siguiente sale. En la pasada siguiente, con la línea ya en el prefijo, **ningún** aviso.
- [ ] **T020** **Verde**: el prefijo con el filtro de bytes *(D-008-P1)*; el modelo de FR-011 y el `cwd` de FR-014; el formato de FR-017,
  sobre el fichero entero; la clasificación de D-008-P9, con el conjunto de `event_id` de la pasada; el `.zst` como entrada de `state.json`
  *(D-008-P2)*; y la línea corrupta de FR-029, con su aviso sólo en la parte nueva *(D-008-P11)*. `state.json` sigue con cuatro campos.
- [ ] **T021** Censo, **antes de mutar**, por hoja:
  - **(m9)** siempre el vigente → (11);
  - **(m10)** sin prefijo → (14) y (15, segunda pasada);
  - **(m11)** sin repetidas → (16) y (19);
  - **(m12)** el `.zst` en cada pasada → (18, segunda);
  - **(m13)** formato anterior por «hay `token_count`» → (17, mixto);
  - **(m14)** `cwd` siempre de `session_meta` → (20);
  - **(m26)** *(E-2)* la línea corrupta corta el fichero → (34);
  - **(m27)** *(E-2)* el aviso también en el prefijo → (34/segunda pasada).
- [ ] **T022** Mutaciones y transcripción, con md5.
- [ ] **T023** ✋ Puertas y commit: `008 B3: contexto entre pasadas, formato y recuentos de Codex`.

## B4 · Raíz y activación *(FR-001, FR-002; M-9)*

- [ ] **T024** Fase 0: `config.CodexSessionsRoot() (string, error)` en `internal/config/codex.go` *(nuevo)*, que devuelve `"", nil`.
- [ ] **T025** [P] **Rojos (21) y (22)** en `internal/config/codex_test.go` *(nuevo)*, con `t.Setenv`:
  - **(21)**, tres subtests: `CODEX_HOME` definida → `<valor>/sessions`; vacía → `<home>/.codex/sessions`; ausente → lo mismo;
  - **(22)** *(E-2)*: con una raíz inexistente se devuelve la ruta igual, sin error. La activación no se decide aquí, sino en cada pasada
    *(FR-002, (32))*.
- [ ] **T026** **Verde**: el único `os.Getenv` de producción *(D-008-P7)*. En `internal/testutil/sandbox.go`, `t.Setenv("CODEX_HOME", "")`
  *(M-9)*.
- [ ] **T027** Comprobar y transcribir, sin tocarlo, que `sandbox_test.go` sigue verde. Y `grep -rn 'os.Getenv' --include=*.go cmd internal
  | grep -v _test` → 1.
- [ ] **T028** Censo, **antes de mutar**, y mutaciones: **(m15)** `CODEX_HOME=""` tomada como raíz → (21/vacía); **(m16)** ignorar
  `CODEX_HOME` → (21/definida).
- [ ] **T029** ✋ Puertas y commit: `008 B4: raiz de Codex con CODEX_HOME`.

## B5 · Integración *(FR-002, FR-003, FR-016, FR-019, FR-021, FR-025, FR-026, FR-028; SC-001–SC-003 por fixture, SC-006, SC-011, SC-015, SC-016, SC-018)*

- [ ] **T030** Fase 0, en este orden:
  1. **Antes de tocar `main.go`** *(E-2, D-008-P12)*: con el binario del commit anterior, el stderr de `--run` sobre el fixture de Claude
     Code se guarda en `cmd/permea/testdata/codex/referencia-run.stderr`, con la ruta de datos sustituida por `<DATOS>`. Su md5 se
     transcribe.
  2. **El forzado de SC-015** *(E-2, D-008-P10)*: se prueba el candidato A *(cola creada antes y directorio de datos sin permiso de
     escritura: `Load` pasa, `Append` pasa y `Save` falla)* y se **declara** el elegido, A o B, con la razón.
  3. En `main.go`, `agent.codexRaiz` y `agent.codex *ingest.PasadaCodex`. La ruta se resuelve en `setup()`, y su existencia se comprueba
     en cada pasada *(D-008-P5)*. `generate()` conserva su firma de hoy *(D-008-P4)*.
- [ ] **T031** [P] **Rojos (23) a (28)** en `cmd/permea/codex_test.go` *(nuevo, sandbox, `CODEX_HOME` a un temporal con fixtures)*:
  - **(23)**, SC-011: sin raíz, y con `CODEX_HOME=""`, el stderr de `--run`, con la ruta de datos sustituida por `<DATOS>`, es byte a
    byte el fichero de referencia de T030. **Nace verde**; lo valida m18;
  - **(24)**, SC-016: sin `~/.claude/projects`, `--run` encola los eventos de Codex y sale con 0;
  - **(25)**, SC-015: con el forzado declarado en T030, la cola tiene los eventos y la pasada devuelve error. Con el directorio
    restaurado, la segunda pasada los reencola con los mismos `event_id`;
  - **(26)**: la línea `codex: respuestas …`, literal, tras el resumen de Claude Code;
  - **(27)**: el predicado del demonio es falso sin respuestas, formato anterior ni comprimidos, y `tick()` no escribe la línea;
  - **(28)**: la segunda `--run` da 0 eventos `codex` *(precondición: la primera da > 0)*.
- [ ] **T032** **Verde**: Codex tras Claude Code en `generate()`, con un solo `st.Save` al final *(D-008-P6)*; la existencia de la raíz en
  cada pasada *(FR-002)*; los errores por fichero de FR-028 *(aviso, estado sin tocar y siguiente fichero; `Append` sigue siendo fatal,
  D-008-P11)*; y la línea en `runOnce`, y en `tick` con su predicado.
- [ ] **T033** Comprobar y transcribir, sin tocarlos, que `main_test.go`, `retencion_test.go`, `coste_test.go` y `project_test.go` siguen
  verdes.
- [ ] **T034** **Rojo (31)**, SC-006: un fixture con centinelas en `response_id`, `session_id` y `turn_id`, **y en el nombre del fichero**.
  Tras `--run`: ninguno en `queue.jsonl`; en `state.json`, sólo el del nombre y **dentro de su clave** *(FR-016)*. Cae primero en su
  precondición *(la cola vacía)*.
- [ ] **T056** *(E-2)* **Rojo (32)**, FR-002, en `cmd/permea/codex_test.go`: un mismo `agent` con `codexRaiz` apuntando a una carpeta que
  **no existe**. La primera pasada no emite nada de Codex ni escribe la línea. Se crea la carpeta con un fixture, y la segunda pasada
  **del mismo `agent`** emite.
- [ ] **T057** *(E-2)* **Rojo (33)**, SC-018 y FR-028, en `cmd/permea/codex_test.go`: un fichero de Codex ilegible *(enlace roto o
  permisos 000; se elige en la Fase 0; `t.Skip` con root o en Windows)*, junto a un fixture de Claude Code y otro de Codex sano. Salen
  los eventos de los dos sanos, stderr lleva `codex: fichero omitido: …`, `state.json` se guarda, y una segunda pasada no reencola nada de
  Claude Code. Subtest `se_relee`: con el fichero ya legible, sus registros salen.
- [ ] **T035** Censo, **antes de mutar**, por hoja:
  - **(m17)** `st.Save` antes del Append de Codex → (25): con el forzado, la cola queda vacía;
  - **(m18)** la línea siempre → (23) ×2;
  - **(m19)** `tick` escribe siempre → (27);
  - **(m20)** Codex sólo si hay raíz de Claude Code → (24);
  - **M-B5a** una clave extra en `state.json` con el `response_id` → (31);
  - **(m23)** *(E-2)* la existencia de la raíz sólo en `setup()` → (32);
  - **(m24)** *(E-2)* el error de un fichero de Codex aborta la pasada → (33) en sus hojas de eventos, `state.json` y segunda pasada;
  - **(m25)** *(E-2)* el fichero omitido guarda su offset al final → (33/se_relee).
- [ ] **T036** Mutaciones y transcripción, con md5.
- [ ] **T037** ✋ Puertas y commit: `008 B5: Codex en run y daemon, antes de guardar el estado`.

## B6 · `--scan` *(FR-020, FR-021; SC-013)*

- [ ] **T038** [P] **Rojos (29) y (30)** en `cmd/permea/codex_test.go`:
  - **(29)**: `--scan` de un fichero de Codex → una línea `evento:` por evento, literal *(spec §Textos aprobados)*, y la línea de resumen;
  - **(30)**: `--scan` de un fixture de Claude Code → la salida de la 0.4.0. **Nace verde**; lo valida M-B6a.
- [ ] **T039** **Verde**: `dryRun` lee la primera línea y bifurca *(D-008-P8)*.
- [ ] **T040** Censo, **antes de mutar**, y mutaciones: **(m21)** `cw5m=`/`cw1h=` en la línea de Codex → (29); **(m22)** sin detección →
  (29); **M-B6a** todo fichero es Codex → (30).
- [ ] **T041** ✋ Puertas y commit: `008 B6: scan de una sesion de Codex`.

## B7 · README y CHANGELOG *(FR-023, FR-024; SC-013)*

- [ ] **T042** **Rojo**, transcrito: `grep -c '^## 0.5.0' CHANGELOG.md` → 0; `grep -c '^### Codex CLI' README.md` → 0.
- [ ] **T043** `README.md`: la sección aprobada, literal, tras la de Claude Code.
- [ ] **T044** `CHANGELOG.md`: `## 0.5.0 — PENDIENTE` encima de la 0.4.0, con el cuerpo **sacado por programa** de spec §Textos aprobados.
  Comprobación: `cmp` de los dos cuerpos, sin diferencias. `grep -c PENDIENTE CHANGELOG.md` → 1.
- [ ] **T045** ✋ Puertas y commit: `008 B7: README y CHANGELOG de la 0.5.0`.

## Cierre — en tramos, uno por mensaje *(plan §Cierre; si uno falla, se para y se rehace desde C1)*

- [ ] **T046** **C1** · Puertas, transcritas: las del bloque, más `go test` → **494 + nuevos** *(SC-012)*; ningún `_test.go` existente
  modificado; `os.Getenv` de producción = 1; compilan Windows y darwin; `PENDIENTE` → 1.
- [ ] **T047** **C2** · Medidas *(quickstart §Copia, §Contador, §M, §Coste)*: la huella de la copia antes; el contador; en sandbox, `--scan`
  y `--run` dos veces *(SC-001 a SC-005 y SC-009)*; **SC-014**, tres veces, con su tiempo; la huella después, igual; temporales borrados.
- [ ] **T048** **C3** · `goreleaser release --snapshot --clean`, el SHA-256 del zip de Windows y `strings` con los textos aprobados *(SC-013)*.
- [ ] **T049** ✋ **C4 · W1** *(quickstart §W1; sandbox, sin enrolar)*: `--version`; `status` → «no enrolado»; dos `--run` sobre la carpeta
  real de Codex, en sólo lectura; la raíz, la de Codex *(Q-6)*.
- [ ] **T050** **C5** · El cuerpo del PR y la fecha del encabezado del CHANGELOG *(`PENDIENTE` → 0)*. ✋ Commit
  `008 C5: fecha de la 0.5.0 en el CHANGELOG`.
- [ ] **T051** ✋ **C6** · Fusión del PR con merge commit.
- [ ] **T052** ✋ **C7** · `git tag -a v0.5.0` sobre `main` y `git push origin v0.5.0`.
- [ ] **T053** **C8** · Los tres canales en `0.5.0` y `strings` del binario publicado *(SC-013)*.
- [ ] **T054** ✋ **C9 · W2** *(quickstart §W2)*: `scoop update`, `--run` y, en la plataforma, eventos `codex` = las respuestas nuevas del
  contador.

## Dependencias

- B0 → B1 → B2 → B3 → B4 → B5 → B6 → B7 → C1…C9, en serie: B2 usa la identidad; B3, la línea; B5, el contexto y la raíz; B6, la línea y
  el contexto.
- Dentro de un bloque, los rojos `[P]` van antes del verde, y el censo antes de mutar.

## Cobertura

| FR | Tareas | | SC | Tareas |
|---|---|---|---|---|
| 001, 002 | T025 (21, 22), T056 (32) | | 001 | T031 *(fixture)*, T047 |
| 003 | T031 (23, 24), T019 (18) | | 002 | T019 (17), T047 |
| 004 | T019 (14, 18), T020 | | 003 | T011 (4), T047 |
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
| 017 | T019 (17) | | 016 | T031 (24) |
| 018 | T019 (18) | | 017 | T019 (19) |
| 019 | T031 (26, 27) | | | |
| 020 | T038 (29) | | | |
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
| m2 | B1 | sin prefijo de longitud | (1) ×2, (2) |
| m3 | B1 | `response_id` vacío aceptado | (3) |
| m4 | B2 | `tokens_input` sin restar la escritura | (5) |
| m5 | B2 | `cost_available = true` | (8) |
| m6 | B2 | el incoherente se emite | (6) |
| m7 | B2 | `session_ref` del `thread_id` | (10) |
| m8 | B2 | `occurred_at` = ahora | (9) |
| M-B2a | B2 | `session_ref` = `session_id` sin sal | (10), T013 |
| m9 | B3 | siempre el modelo vigente | (11) |
| m10 | B3 | sin prefijo *(empezar en el offset)* | (14), (15, segunda pasada) |
| m11 | B3 | sin repetidas | (16), (19) |
| m12 | B3 | el `.zst` en cada pasada | (18, segunda) |
| m13 | B3 | formato anterior por «hay `token_count`» | (17, mixto) |
| m14 | B3 | `cwd` siempre de `session_meta` | (20) |
| m26 | B3 | *(E-2)* la línea corrupta corta el fichero | (34) |
| m27 | B3 | *(E-2)* el aviso también en el prefijo | (34/segunda pasada) |
| m15 | B4 | `CODEX_HOME=""` tomada como raíz | (21/vacía) |
| m16 | B4 | ignorar `CODEX_HOME` | (21/definida) |
| m17 | B5 | `st.Save` antes del Append de Codex | (25), con el forzado de T030 |
| m18 | B5 | la línea de Codex siempre | (23) ×2 |
| m19 | B5 | `tick` escribe siempre | (27) |
| m20 | B5 | Codex sólo si hay raíz de Claude Code | (24) |
| M-B5a | B5 | una clave extra en `state.json` con el `response_id` | (31) |
| m23 | B5 | *(E-2)* la existencia sólo en `setup()` | (32) |
| m24 | B5 | *(E-2)* el error de Codex aborta la pasada | (33) |
| m25 | B5 | *(E-2)* el fichero omitido guarda su offset al final | (33/se_relee) |
| m21 | B6 | `cw5m=`/`cw1h=` en la línea de Codex | (29) |
| m22 | B6 | sin detección de Codex en `--scan` | (29) |
| M-B6a | B6 | todo fichero es Codex | (30) |

**30 mutaciones previstas.** Las co-caídas que aparezcan al declarar el censo se escriben **antes** de mutar, por hoja.

## Lo que este plan de tareas NO hace

- No toca `internal/event`, la derivación del `event_id` de 006, `boundary_test.go`, `internal/state/` ni ningún `_test.go` existente.
- No cambia la plataforma. Las tarifas de Codex van en un encargo aparte y posterior *(D-3)*.
- No lanza `enroll`, `--run` ni `--daemon` sobre la instalación real. Los `--run` de los tests van en sandbox, y los de W1 y W2 los lanza
  el dueño.
