# Tasks: 009 · «Lector de Gemini CLI»

**Feature**: `009-lector-gemini` · **Fecha**: 2026-10-08 · [spec.md](./spec.md) *(ratificada el 2026-10-08, 12:20)* · [plan.md](./plan.md) ·
[contracts/event-id-gemini.md](./contracts/event-id-gemini.md)
**Base**: `15ce93b`. **Línea base a preservar**: `go test -count=1 ./...` → **9 paquetes ok, 574 pass, 0 fail, 0 SKIP**; `golangci-lint run` → **0**.

## Formato: `[ID] [P?] [✋?] Descripción`

- **[P]**: sin dependencia de decisión entre sí. **No** quiere decir que se editen a la vez: dos `[P]` sobre el mismo fichero se escriben
  una detrás de otra.
- **✋**: la ejecuta **el dueño**: commits, la PR, la fusión, la etiqueta y los ensayos en Windows. Claude no hace git de escritura. El
  mensaje de cada commit va previsto, **sin tildes ni ñ**.
- **(1)…(38)** son los rojos y **(m1)…(m38)** las mutaciones de `plan.md` §Bloques, con la misma numeración. Lo que se añada en este fichero
  se llama **M-B1a**, **M-B2a**…
- **No se renumera**: una tarea añadida después recibe el siguiente número libre.
- Las transcripciones van a `soporte/registro.md`, que se crea en B1, y en cada tarea queda una remisión.

## Presupuesto *(escrito antes de la fase; ningún techo se sube: si no cabe, se condensa)*

| Fichero de `specs/009-lector-gemini/` | Techo | Hoy |
|---|---:|---:|
| `spec.md` | 450 | 392 |
| `contracts/event-id-gemini.md` | 120 | 104 |
| `plan.md` | 220 | 194 |
| `tasks.md` | 320 | 298 |
| `soporte/descubrimiento.md` | 250 *(cerrado)* | 250 |
| `soporte/registro.md` *(nace en B1)* | 450 | — |

| Bloque | Producción | Test | Fixtures | Registro |
|---|---:|---:|---|---:|
| B1 | ~35 | ~60 | — | ≤ 30 |
| B2 | ~110 | ~220 | ~15 `.jsonl` | ≤ 60 |
| B3 | ~210 | ~380 | ~20 `.jsonl` y carpetas `tmp/<slug>/` | ≤ 90 |
| B4 | ~25 | ~50 | — | ≤ 30 |
| B5 | ~110 | ~300 | la referencia y ~6 ficheros | ≤ 80 |
| B6 + B7 | ~55 + ~40 de texto | ~90 | ~2 | ≤ 50 |
| Cierre C1–C9 | — | — | — | ≤ 110 |

**~545 de producción y ~1 100 de test.** Una desviación de más del 50 % en un bloque se declara en su corte.

## Disciplinas y protocolo

- **Disciplinas 1–11 de 008** y la **12** de este plan *(la segunda copia congelada sólo se lee)*.
- **Protocolo de mutación** *(006, con E-5 de 007)*:
  1. el censo se declara **en la tarea, ANTES de mutar**, con las co-caídas, **por HOJA (subtest), nunca por test padre**;
  2. se muta y se ejecuta `go test -count=1 ./... 2>&1`;
  3. si el conjunto de `FAIL` **coincide** con lo declarado, se revierte **por edición inversa**, se comprueba con **md5** que el fichero
     vuelve a ser el de antes y se transcribe el fallo;
  4. si **no** coincide, la mutación **se deja puesta y se para**;
  5. una mutación que no compila o que panica no cuenta como superada.
- **Rojo antes de verde**, transcribiendo la razón real. **Un test que nace verde lleva su mutación declarada.**
- **Ningún `_test.go` existente se toca** *(censo: ninguno)*. Cambian `cmd/permea/main.go`, `internal/testutil/sandbox.go` *(M-9)*, el
  comentario de `internal/config/codex.go:16`, `README.md` y `CHANGELOG.md`, y entran ficheros nuevos.
- **Fixtures sólo sintéticos**: `id` `m-0000…`, `sessionId` `s-0000…`, `projectHash` `h-0000…`, rutas inventadas. Nada con forma de UUID.

**Puertas de cada bloque**, antes de su ✋ commit:
```sh
gofmt -l .                     # → vacío
go vet ./...                   # → sin hallazgos
golangci-lint run              # → 0 issues (2.12.2, sin tope)
go test -count=1 ./...         # → 9 paquetes ok; 574 + los nuevos
git diff 15ce93b -- internal/event/ internal/ingest/eventid.go internal/ingest/eventid_test.go \
  internal/ingest/codex_eventid.go internal/ingest/codex_eventid_test.go internal/ingest/boundary_test.go \
  specs/006-medicion-fiel/contracts/event-id.md specs/008-lector-codex/contracts/event-id-codex.md   # → vacío (FR-027)
git diff 15ce93b --name-only --diff-filter=M -- '*_test.go'                                        # → vacío (M-8)
```

---

## B0 · Documentos

- [x] **T001** Fase 0. Transcribir la línea base sobre `15ce93b`: 574 pass en 9 paquetes, 0 SKIP, lint 0 y la frontera sin diff. Si algo
  difiere, **se para**.
- [x] **T002** ✋ *(`19cda8e`)* Commit: `009 B0: spec ratificada, contrato, plan y tareas`.

## B1 · Identidad *(contrato; FR-007; SC-006)*

- [x] **T003** Fase 0: en `internal/ingest/gemini_eventid.go` *(nuevo)*, `derivarEventIDGemini(id string) (string, bool)`, que devuelve `"",
  false`. Se crea `soporte/registro.md`. La suite sigue verde.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T003.
- [x] **T004** **Rojos (1), (2) y (3)** en `internal/ingest/gemini_eventid_test.go` *(nuevo)*:
  - **(1)**: los dos vectores normativos, byte a byte;
  - **(2)**: el vector del espacio de Codex, como literal *(`07d1f3ea…`)*, es distinto del de Gemini para el mismo valor;
  - **(3)**: `id` vacío → `ok = false` *(nace verde; la valida m3)*.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T004.
- [x] **T005** **Verde**: la derivación del contrato, **replicada** en el fichero nuevo *(plan §Qué se reutiliza)*.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T005 *(md5 `60ce5a10…`)*.
- [x] **T006** Censo, **declarado antes de mutar**, por hoja: **(m1)** `"gemini"` → `"codex"` → (1) ×2 y (2); **(m2)** sin prefijo de
  longitud → (1) ×2; **(m3)** vacío aceptado → (3); **(m4)** el `sessionId` en el hash → (1) ×2.
- [x] **T007** Mutaciones y transcripción, con md5.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T006 y T007 *(las cuatro coinciden)*, y T008 *(puertas: 579 pass)*.
- [ ] **T008** ✋ Puertas y commit: `009 B1: event_id de Gemini con espacio de nombres propio`.

## B2 · Una aparición *(FR-009 a FR-015, FR-017, FR-019; SC-005, SC-007)*

- [ ] **T009** Fase 0: en `internal/ingest/gemini.go` *(nuevo)*:
  - `ContextoGemini`: el `Context` de siempre, el `sessionId` de la cabecera y el texto de `.project_root`;
  - `ClaseGemini`;
  - `RespuestaGemini(crudo, ctx)`: una aparición en crudo → evento o clase. Todavía no clasifica nada.
- [ ] **T010** Fixtures **sintéticos** en `internal/ingest/testdata/gemini/`: una aparición por clase de FR-019, y las partidas de SC-007.
- [ ] **T011** [P] **Rojos (4) a (12)** en `internal/ingest/gemini_respuesta_test.go` *(nuevo)*:
  - **(4)** SC-007: 100/40/5/7/3 → `tokens_input` 65, `tokens_cache_read` 40, `tokens_cache_creation` 0, `tokens_output` 10;
  - **(5)** una partida ausente vale 0, y se emite;
  - **(6)** incoherente, seis hojas: `tokens` no objeto · partida no numérica · negativa · `cached > input` · sin `timestamp` · mal formado;
  - **(7)** sin identificador: ausente, vacío y no textual;
  - **(8)** `tool = "gemini"`, `cost_usd = 0` y `cost_available = false`;
  - **(9)** `occurred_at` = `timestamp`;
  - **(10)** sin `model` → emitido con `model` vacío y marcado;
  - **(11)** `total` descuadrado → emitido y marcado; sin `total`, sin marca;
  - **(12)** `session_ref = Ref(sal, sessionId)`; `project_ref = Derivar(.project_root)`; sin `.project_root`, vacío.
- [ ] **T012** **Verde**: D-2, la clasificación de D-009-P14 y el evento con `event.Ref` y el resolutor. **No** llama a `internal/pricing`.
- [ ] **T013** Verde de nacimiento, escrito tras T012: **(FR-017)** el evento no contiene ningún centinela *(`id`, `sessionId`,
  `projectHash`, texto de `.project_root`)*. Lo valida m12.
- [ ] **T014** Censo, **antes de mutar**, por hoja: m5 a m12, con lo que debe caer en §Mutaciones.
- [ ] **T015** Mutaciones y transcripción, con md5.
- [ ] **T016** ✋ Puertas y commit: `009 B2: una respuesta de Gemini es un evento sin coste`.

## B3 · Fichero y contexto *(FR-004 a FR-006, FR-008, FR-014, FR-016, FR-018 a FR-020; SC-004, SC-008 a SC-011, SC-013)*

- [ ] **T017** Fase 0: en `internal/ingest/gemini_contexto.go` *(nuevo)*:
  - `PasadaGemini`: los ocho recuentos, `Resumen()` *(texto aprobado)*, `HayNovedades()`, `emitidos` y la caché de `.project_root`;
  - `LeerFicheroGemini(st, ruta, base, p, avisos io.Writer)`, sobre `Recorrer`, que aún no emite;
  - `ContarAnteriorGemini(st, ruta, p)` y `ListarGemini(raiz)`.

  Compila, y la suite sigue verde.
- [ ] **T018** Fixtures sintéticos en `testdata/gemini/contexto/`, con forma de `tmp/<slug>/chats/…`: una llamada a herramienta; tokens
  tardíos; un `$set.messages` final sin tokens; una sesión **con la forma de la copia** *(herramientas, compresión, reanudación)* y su corte
  antes de la reanudación; la cuenta de SC-008; una línea corrupta; un `.json`, `logs.json`, `*.unreadable-*` y `*.tmp-*`; un subagente;
  dos carpetas con la misma sesión, una sin `.project_root`; un fichero largo y su versión truncada.
- [ ] **T019** [P] **Rojos (13) a (25)** en `internal/ingest/gemini_contexto_test.go` *(nuevo)*: **(13)** cuentan las apariciones de
  mensaje y de `$set.messages`, y `tokens: null` no; **(14)** SC-009, tokens tardíos → 1; **(15)** SC-009, el `$set` final no borra eventos;
  **(16)** repetidas en la pasada; **(17)** SC-004, `12 · 8 · 4` y luego `4 · 2 · 2`; **(18)** dos pasadas = una *(`event_id`, modelo,
  `project_ref`, `session_ref`)*; **(19)** la cabecera en el prefijo; **(20)** el `.json` cuenta 1, 0, 1 *(la segunda nace verde; la valida
  m18)*; **(21)** SC-008, el resumen literal y la identidad de FR-019; **(22)** la línea corrupta, con aviso en lo nuevo y sin él en el
  prefijo; **(23)** `ListarGemini`, con patrón, exclusiones y orden de D-009-P4; **(24)** el proyecto: con, sin, subagente, e ilegible →
  error; **(25)** el truncado se relee sin prefijo.
- [ ] **T020** **Verde**: el prefijo con el filtro de bytes *(D-009-P1)*; la cabecera *(D-009-P2)*; el proyecto *(D-009-P3)*; el orden
  *(D-009-P4)*; el `.json` *(D-009-P5)*; la clasificación con `emitidos` y `vistos`; y la línea corrupta *(D-009-P10)*. `state.json` sigue
  con cuatro campos. **Medida informativa de SC-013**, transcrita.
- [ ] **T021** Censo, **antes de mutar**, por hoja: m13 a m22 y M-B3a, con lo que debe caer en §Mutaciones.
- [ ] **T022** Mutaciones y transcripción, con md5.
- [ ] **T023** ✋ Puertas y commit: `009 B3: apariciones, contexto entre pasadas y formato de Gemini`.

## B4 · Raíz *(FR-001; M-9)*

- [ ] **T024** Fase 0: `config.GeminiRoot() (string, error)` en `internal/config/gemini.go` *(nuevo)*, que devuelve `"", nil`.
- [ ] **T025** [P] **Rojos (26) y (27)** en `internal/config/gemini_test.go` *(nuevo)*, con `t.Setenv`:
  - **(26)** tres subtests: definida → `<valor>/.gemini`; vacía → `<home>/.gemini`; ausente → lo mismo;
  - **(27)** una raíz inexistente se devuelve sin error.
- [ ] **T026** **Verde**: la segunda lectura de entorno *(D-009-P9)*. El comentario de `codex.go:16` pasa a «una de las dos». En
  `internal/testutil/sandbox.go`, tras `:63`, `t.Setenv("GEMINI_CLI_HOME", "")` *(M-9)*.
- [ ] **T027** Comprobar y transcribir, sin tocarlo, que `sandbox_test.go` sigue verde. Y `grep -rn 'os.Getenv' --include=*.go cmd internal
  | grep -v _test` → 2.
- [ ] **T028** Censo, **antes de mutar**, y mutaciones: **(m23)** la vacía tomada como raíz → (26/vacía); **(m24)** ignorar la variable →
  (26/definida); **(m25)** exigir que exista → (27).
- [ ] **T029** ✋ Puertas y commit: `009 B4: raiz de Gemini con GEMINI_CLI_HOME`.

## B5 · Integración *(FR-002, FR-003, FR-017, FR-021, FR-023 a FR-026; SC-003, SC-006, SC-012, SC-016, SC-017)*

- [ ] **T030** Fase 0, **en este orden**:
  1. **Antes de tocar `main.go`** *(D-009-P11)*: con el binario de `15ce93b`, el stderr de `--run` sobre fixtures de Claude Code y de
     Codex, en `cmd/permea/testdata/gemini/referencia-run.stderr`, con `<DATOS>`. Se transcribe su md5. **Si `main.go` ya tiene un cambio,
     se para.**
  2. **El forzado A** de SC-017 *(D-009-P12)*, comprobado y transcrito: `Load` pasa, `Append` pasa y `Save` falla.
  3. En `cmd/permea/gemini.go` *(nuevo)*, `generarGemini`, que no emite. En `main.go`: `agent.geminiRaiz` y `agent.gemini`; `setup()`; y
     `generate()` con su firma de hoy *(D-009-P7, P8)*.
- [ ] **T031** [P] **Rojos (28) a (33)** en `cmd/permea/gemini_test.go` *(nuevo)*, con `cola` y `capturarStderr` de `codex_test.go`, sin
  tocarlo:
  - **(28)** SC-012, con `GEMINI_CLI_HOME` vacía y sin `~/.gemini`: byte a byte la referencia *(nace verde; la valida m26)*;
  - **(29)** sólo Gemini;
  - **(30)** SC-017, con el forzado A;
  - **(31)** la línea `gemini:` literal, tras la de Codex, y «N eventos encolados» la incluye;
  - **(32)** el predicado y `tick`;
  - **(33)** SC-003: la segunda `--run` da 0.
- [ ] **T032** **Verde**: Gemini tras Codex y antes del único `st.Save`; `<raíz>/tmp` en cada pasada; los errores por fichero
  *(D-009-P10)*; la línea en `runOnce` y, con `HayNovedades`, en `tick`.
- [ ] **T033** Comprobar y transcribir, sin tocarlos, que `main_test.go`, `retencion_test.go`, `coste_test.go`, `project_test.go` y
  `codex_test.go` siguen verdes.
- [ ] **T034** **Rojo (34)**, SC-006: centinelas en `id`, `sessionId`, `projectHash`, `.project_root`, el nombre del fichero y la carpeta del
  subagente. Tras `--run`: ninguno en la cola; en `state.json`, sólo los de las rutas y **dentro de su clave** *(FR-017)*.
- [ ] **T035** **Rojo (35)**, FR-002: `GEMINI_CLI_HOME` en una carpeta sin `.gemini/tmp`. La primera pasada no emite. Se crea, y la segunda
  pasada **del mismo `agent`** emite.
- [ ] **T036** **Rojo (36)**, SC-016: un fichero de Gemini con permisos 000 junto a fixtures sanos de las tres herramientas. Salen los sanos y
  el aviso `gemini: fichero omitido: …`; `state.json` se guarda; la segunda pasada no reencola Claude Code ni Codex; con `se_relee`, sale.
- [ ] **T037** Censo, **antes de mutar**, por hoja: m26 a m34, con lo que debe caer en §Mutaciones.
- [ ] **T038** Mutaciones y transcripción, con md5.
- [ ] **T039** ✋ Puertas y commit: `009 B5: Gemini en run y daemon, antes de guardar el estado`.

## B6 · `--scan` *(FR-022, FR-023; SC-015)*

- [ ] **T040** [P] **Rojos (37) y (38)** en `cmd/permea/gemini_test.go`, con los textos leídos de las specs por programa:
  - **(37)** un fichero de Gemini → la línea `evento:` de 008, una por evento; la línea `gemini:`; y nada en disco *(`nada_en_disco` nace
    verde; la valida m37)*;
  - **(38)** Claude Code y Codex → byte a byte su salida de la 0.5.0 *(nace verde; la valida m36)*.
- [ ] **T041** **Verde**: en `dryRun`, tras Codex, `esSesionGemini` y `dryRunGemini` *(D-009-P13)*.
- [ ] **T042** Censo, **antes de mutar**, y mutaciones: **(m35)** sin detección → (37); **(m36)** todo es Gemini → (38); **(m37)**
  `dryRunGemini` guarda estado → (37/nada_en_disco); **(m38)** detección sólo por `sessionId` → (38/claude).
- [ ] **T043** ✋ **El dueño aprueba las frases de coherencia del README** *(P-11)*, que se proponen en este bloque, literales, en
  `soporte/registro.md`.

## B7 · README y CHANGELOG *(FR-028; SC-015)*

- [ ] **T044** **Rojo**, transcrito: `grep -c '^## 0.6.0' CHANGELOG.md` → 0; `grep -c '^### Gemini CLI' README.md` → 0.
- [ ] **T045** `README.md`: «### Gemini CLI», literal y **sacada por programa** de spec §Textos aprobados, tras «### Codex CLI». Y las frases
  de coherencia aprobadas en T043.
- [ ] **T046** `CHANGELOG.md`: `## 0.6.0 — PENDIENTE` encima de la 0.5.0, con el cuerpo **sacado por programa**. `cmp` de los dos textos, sin
  diferencias. `grep -c PENDIENTE CHANGELOG.md` → 1.
- [ ] **T047** ✋ Puertas y commit de B6 y B7: `009 B6 y B7: scan de Gemini, README y CHANGELOG de la 0.6.0`.

## Cierre — en tramos, uno por mensaje *(plan §Cierre; si uno falla, se para y se rehace desde C1)*

- [ ] **T048** **C1** · Puertas, transcritas: las del bloque; `go test` → **574 + nuevos**, 0 SKIP en Linux *(SC-014)*; ningún `_test.go`
  existente modificado; `os.Getenv` de producción = 2; compilan Windows y darwin; `PENDIENTE` → 1.
- [ ] **T049** **C2** · Medidas *(plan §Contador)*: las huellas de las dos copias, antes; el contador y `--scan` sobre la primera; `--run` dos
  veces con `GEMINI_CLI_HOME` en la **segunda copia congelada** *(SC-001 a SC-004)*; **SC-013**, tres veces; las huellas después, iguales;
  los temporales, borrados.
- [ ] **T050** **C3** · `goreleaser release --snapshot --clean`, el SHA-256 del zip de Windows y `strings` con los textos aprobados
  *(SC-015)*.
- [ ] **T051** ✋ **C4 · W1** *(sandbox, sin enrolar)*: `--version`; `status`; dos `--run` sobre el `.gemini` real en **sólo lectura**,
  contrastados con el contador; **Q-1**: el `project_ref` de un directorio con sesiones de Claude Code y de Gemini, igual en los dos.
- [ ] **T052** **C5** · El cuerpo del PR y la fecha del encabezado del CHANGELOG *(`PENDIENTE` → 0)*. ✋ Commit `009 C5: fecha de la 0.6.0 en el CHANGELOG`.
- [ ] **T053** ✋ **C6** · Fusión del PR con merge commit.
- [ ] **T054** ✋ **C7** · `git tag -a v0.6.0` sobre `main` y `git push origin v0.6.0`.
- [ ] **T055** **C8** · Los tres canales en `0.6.0` y `strings` del binario publicado *(SC-015)*.
- [ ] **T056** ✋ **C9 · W2**: `scoop update`, `--run` y, en la plataforma, eventos `gemini` = las respuestas nuevas del contador.

## Dependencias

- B0 → B1 → B2 → B3 → B4 → B5 → B6 → B7 → C1…C9, en serie. B2 usa la identidad; B3, la aparición; B5, el fichero y la raíz; B6, el
  fichero.
- Dentro de un bloque, los rojos `[P]` van antes del verde, y el censo antes de mutar.

## Cobertura *(SC → bloque que lo acredita)*

| SC | Bloque | Tareas | | SC | Bloque | Tareas |
|---|---|---|---|---|---|---|
| 001 | C2 | T049 | | 010 | B3 | T019 (24), (23/orden) |
| 002 | C2 | T049 | | 011 | B3 · B4 | T019 (20), (23) · T025 (26) |
| 003 | B5 · C2 | T031 (33), T049 | | 012 | B5 | T030 *(referencia)*, T031 (28) |
| 004 | B3 · C2 | T019 (17), T049 | | 013 | B3 · C2 | T020 *(informativa)*, T049 |
| 005 | B2 · C2 | T011 (8), T049 | | 014 | todos · C1 | puertas, T048 |
| 006 | B1 · B5 | T004 (1–3), T034 (34) | | 015 | B5 · B6 · B7 · C3 · C8 | T031 (31), T040 (37), T046, T050, T055 |
| 007 | B2 | T011 (4) | | 016 | B5 | T036 (36) |
| 008 | B2 · B3 | T011 (6, 7, 10, 11), T019 (21) | | 017 | B5 | T031 (29, 30) |
| 009 | B3 | T019 (14), (15) | | | | |

**28 / 28 FR** *(plan §Trazabilidad)* **y 17 / 17 SC con bloque.**

## Mutaciones

| # | Bloque | Mutación | Debe caer *(y sólo eso)* |
|---|---|---|---|
| m1 | B1 | `"gemini"` → `"codex"` | (1) ×2, (2) |
| m2 | B1 | sin prefijo de longitud | (1) ×2 |
| m3 | B1 | `id` vacío aceptado | (3) |
| m4 | B1 | el `sessionId` en el hash | (1) ×2 |
| m5 | B2 | `tokens_input` sin `tool` | (4) |
| m6 | B2 | `tokens_output` sin `thoughts` | (4) |
| m7 | B2 | `tokens_input` sin restar `cached` | (4) |
| m8 | B2 | `cost_available = true` | (8) |
| m9 | B2 | el incoherente se emite | (6) ×6 |
| m10 | B2 | `occurred_at` = ahora | (9) |
| m11 | B2 | total descuadrado = incoherente | (11/descuadrado) |
| m12 | B2 | `session_ref` sin sal | (12/session_ref), T013 |
| m13 | B3 | sin prefijo | (17/segunda), (18), (19) |
| m14 | B3 | prefijo sin `vistos` *(P-3 (b))* | (17/segunda) |
| m15 | B3 | `$set.messages` reemplaza el estado | (15) |
| m16 | B3 | la aparición sin tokens bloquea el `id` | (14) |
| m17 | B3 | sin repetidas | (16), (21) |
| m18 | B3 | el `.json` en cada pasada | (20/segunda) |
| m19 | B3 | el aviso también en el prefijo | (22/segunda) |
| m20 | B3 | orden sólo léxico | (23/orden) |
| m21 | B3 | `.project_root` del subagente en `chats/<padre>/` | (24/subagente) |
| m22 | B3 | sin `.project_root` es error | (24/sin) |
| M-B3a | B3 | el truncado lee prefijo | (25) |
| m23 | B4 | `GEMINI_CLI_HOME=""` tomada como raíz | (26/vacía) |
| m24 | B4 | ignorar `GEMINI_CLI_HOME` | (26/definida) |
| m25 | B4 | exigir que la raíz exista | (27) |
| m26 | B5 | la línea de Gemini siempre | (28) |
| m27 | B5 | `st.Save` antes de encolar Gemini | (30), con el forzado A |
| m28 | B5 | `tick` escribe siempre | (32/tick) |
| m29 | B5 | Gemini sólo con raíz de Claude Code | (29), (33), (35) |
| m30 | B5 | la existencia sólo en `setup()` | (35) |
| m31 | B5 | el error de Gemini aborta la pasada | (36/pasada) |
| m32 | B5 | el omitido guarda su offset | (36/se_relee) |
| m33 | B5 | la ruta de `.project_root` en el evento | (34) |
| m34 | B5 | Gemini antes que Codex en stderr | (31) |
| m35 | B6 | sin detección de Gemini en `--scan` | (37) |
| m36 | B6 | todo fichero es Gemini | (38) |
| m37 | B6 | `dryRunGemini` guarda su estado | (37/nada_en_disco) |
| m38 | B6 | detección sólo por `sessionId` | (38/claude) |

**39 mutaciones previstas.** Las co-caídas que aparezcan al declarar el censo se escriben **antes** de mutar, por hoja.

## Lo que este plan de tareas NO hace

- No toca `internal/event`, las derivaciones de 006 y 008, `boundary_test.go`, `internal/state/`, los ficheros de Codex de `internal/ingest/`
  ni ningún `_test.go` existente.
- No cambia la plataforma. Las tarifas de Gemini van en un encargo aparte y posterior *(D-4)*.
- No lanza `gemini`, `enroll`, `--run` ni `--daemon` sobre la instalación real. Los `--run` de los tests y de C2 van en sandbox; los de W1 y
  W2 los lanza el dueño.
- No escribe en ninguna copia congelada, ni pone su ruta en el repo.
