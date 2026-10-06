# Tasks: 007 · «Coste fiel»

**Feature**: `007-coste-fiel` · **Fecha**: 2026-10-06 · [spec.md](./spec.md) *(E-1 a E-5)* · [plan.md](./plan.md) ·
[contracts/tarifas.md](./contracts/tarifas.md) · [quickstart.md](./quickstart.md)
**Base**: `222c824`. **Línea base a preservar**: `go test -count=1 ./...` → **9 paquetes ok, 432 pass, 0 fail**; `golangci-lint run` → **0**.
Ambas medidas el 2026-10-06.

## Formato: `[ID] [P?] [✋?] Descripción`

- **[P]**: sin dependencia de decisión entre sí. **No** quiere decir que se editen a la vez: dos `[P]` sobre el mismo fichero se escriben
  una detrás de otra.
- **✋**: la ejecuta **el dueño**: commits, la PR, la fusión, la etiqueta y los ensayos en Windows. Claude no hace git de escritura.
  El mensaje de cada commit va previsto, **sin tildes ni ñ**.
- **(1)…(25)** son los rojos y **(m1)…(m24)** las mutaciones de `plan.md` §Bloques, con la misma numeración. Lo propio de este fichero
  se llama **R-B4a** y **M-B4a**.
- **No se renumera**: una tarea añadida después recibe el siguiente número libre.

## Disciplinas y protocolo

- **Disciplinas 1–9 de 006** *(`specs/006-medicion-fiel/tasks.md` §Disciplinas transversales)* y la **10** de `plan.md` *(las copias del
  dueño se leen en su sitio)*.
- **Protocolo de mutación de 006, sin cambios**:
  1. el censo se declara **en la tarea, antes** de mutar, con las co-caídas, **por HOJA (subtest), nunca por test padre** *(E-5)*;
  2. se muta y se ejecuta `go test -count=1 ./... 2>&1`;
  3. si el conjunto de `FAIL` **coincide** con lo declarado, se revierte por edición inversa y se transcribe el fallo;
  4. si **no** coincide, la mutación **se deja puesta y se para**;
  5. una mutación que no compila o que panica no cuenta como superada.
- **Rojo antes de verde**, transcribiendo la razón real. Lo que nazca verde lleva su mutación.

**Puertas de cada bloque**, antes de su ✋ commit:
```sh
gofmt -l .                     # → vacío
go vet ./...                   # → sin hallazgos
golangci-lint run              # → 0 issues (2.12.2, sin tope)
go test -count=1 ./...         # → 9 paquetes ok
git diff 222c824 -- internal/event/ internal/ingest/eventid.go internal/ingest/eventid_test.go \
  internal/ingest/boundary_test.go specs/006-medicion-fiel/contracts/event-id.md      # → vacío (SC-009, FR-018)
grep -rnE '\.(go|md|jsonl|json|sh|yaml|yml):[0-9]+|[(`]:[0-9]+' --include='*.go' .      # → nada (disciplina 8)
```

---

## B0 · Documentos

- [x] **T001** Fase 0. Transcribir la línea base sobre `222c824`:
  - suite: 432 pass y 9 paquetes;
  - lint: 0;
  - el `git diff` de la frontera, vacío.
  Si algo difiere, **se para**.
  > **Hecho el 2026-10-06**, sobre `49cf73c` *(sólo documentos desde `222c824`)*: 432 pass y 9/9 paquetes ok; `0 issues.`; diff de la
  > frontera de 0 bytes.
- [x] **T002** En `specs/006-medicion-fiel/contracts/tarifas.md`, una línea bajo el título: «*Sustituido el 2026-10-06 por
  `specs/007-coste-fiel/contracts/tarifas.md` (cinco cifras, `8f147d1`).*». Ningún otro cambio en 006.
- [x] **T003** ✋ *(`0d90b5c`)* Commit: `007 B0: enmienda E-3 y contrato de tarifas de 006 sustituido` *(E-3: los documentos ya entraron en el commit
  «007: spec ratificada (E-1, E-2), plan, tareas, contrato de tarifas y quickstart»)*.

## B1 · Tarifas *(`contracts/tarifas.md`; FR-006, FR-007, FR-008; SC-010)*

- [x] **T004** Fase 0: `CacheWrite1h float64` en `Rate`, `internal/pricing/pricing.go`, sin rellenar. La suite sigue verde.
- [x] **T005** **Rojos (1) y (2)** en `internal/pricing/pricing_test.go`:
  - `esperadaDelCatalogo` pasa a las **17 × 5** del contrato, escrita a mano desde `contracts/tarifas.md` *(no copiada de `pricing.go`)*,
    con la procedencia `8f147d1`;
  - `TestEspejo_CifrasClaveAClave` compara también `CacheWrite1h`.
  - **Esperado**: `TestEspejo_RecuentoDeClaves` falla con 16 ≠ 17; `CifrasClaveAClave` falla en 16 subtests por `cache_write_1h` y en
    `claude-fable-5-1` por clave ausente.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T005.
- [x] **T006** **Verde**: `Table` a 17 × 5. La cabecera cita `8f147d1`, la verificación del 2026-10-04 y la aprobación *(contrato
  §La cabecera)*. **La «Limitación 1» se queda hasta B2.**
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T006.
- [x] **T007** Mutaciones, cada una con su censo declarado antes de mutar:
  - **(m1)** `CacheWrite1h` de `claude-fable-5-1` a 0 → cae sólo `CifrasClaveAClave/claude-fable-5-1` *(SC-010)*;
  - **(m2)** una clave sobrante → caen `NingunaClaveSobra` y `RecuentoDeClaves`;
  - **(m3)** cruzar 5 m y 1 h en `claude-opus-5-5` → cae `CifrasClaveAClave/claude-opus-5-5` y, **co-caída declarada antes de
    mutar** *(2026-10-06; el plan no la recogía)*, `TestCost_Opus55AMano`: `Cost` tarifa la escritura con `CacheWrite`, que pasa a 8,00.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 · T007.
- [x] **T008** ✋ *(`3dc064c`)* Puertas y commit: `007 B1: tarifas de 17 modelos con cinco cifras (8f147d1)`.

## B2 · Desglose y coste por duración *(FR-001 a FR-005, FR-007, FR-016; SC-004)*

- [x] **T009** Fase 0 *(paso mecánico)*:
  - `Cost(model, in, out, cw5m, cw1h, cr)` con **la conducta de hoy**: las dos a `CacheWrite`;
  - `FromClaudeCodeLine` pasa el total como `cw5m` y 0 como `cw1h`;
  - `TestCost`, `TestCost_UnknownModel` y `TestCost_Opus55AMano` cambian **sólo la llamada**.
  - La suite sigue verde y con el mismo número de tests.
- [x] **T010** **Rojo (3)** en `pricing_test.go`: `TestCost_Opus55AMano` pasa a los dos vectores de SC-004, comparados en absoluto con
  tolerancia ≤ 1e-9.
  - Con desglose *(12 345 a 5 min / 33 334 a 1 h)*: **1,1775756**. Hoy da 1,0775736.
  - Sin desglose *(0 / 45 679)*: **1,2146106**.
- [x] **T011** [P] **Rojos (4) a (7)** en `internal/ingest/desglose_test.go` *(nuevo)*, con líneas sintéticas y sin pasada:
  - **(4)**: coste 1,1775756 y `tokens_cache_creation` = 45 679;
  - **(5)**: la línea sin `cache_creation` da 1,2146106;
  - **(6)**: con 5 m + 1 h ≠ total *(p. ej. 10 000 + 10 000 frente a 45 679)*, da 1,2146106 y el total del log;
  - **(7)**: con pasada, `Recuentos().SinDesglose` = 2 tras (5) y (6).
- [x] **T012** [P] **Rojo (8)** en `cmd/permea/coste_test.go` *(nuevo, de proceso)*: `--scan` de una línea con desglose imprime
  `cw5m=12345 cw1h=33334` y conserva `cw=45679`. Se compara `ExitCode()`.
- [x] **T013** **Verde**:
  - `rawRecord` decodifica `usage.cache_creation.ephemeral_5m_input_tokens` y `…_1h_…` *(comentario de la guarda: números de consumo)*;
  - se aplican las reglas P-1 y Q-4;
  - `Cost` tarifa por duración;
  - `consumo` gana el desglose y `Recuentos` gana `SinDesglose`;
  - `dryRun` imprime `cw5m=` y `cw1h=` detrás de `cw=`. El desglose viaja por la pasada, no por el evento *(D-007-P2)*;
  - la cabecera cambia la «Limitación 1» por la **hipótesis P-1**.
- [x] **T014** Comprobar y transcribir, sin tocarlos:
  - `TestScan_LineaConCuatroPartidasYEventID` sigue verde, porque `" cw=7 "` sigue en la línea;
  - los cinco tests de `boundary_test.go` siguen verdes.
- [x] **T015** Mutaciones:
  - **(m4)** cruzar las tarifas en `Cost` → caen (3) y (4). Co-caída a declarar: `TestCost`, si su vector usa 1 h;
  - **(m5)** sin desglose a 5 min → caen (5) y (6);
  - **(m6)** aceptar el desglose incoherente → cae (6);
  - **(m7)** `tokens_cache_creation` = 5 m + 1 h → cae (6);
  - **(m8)** `cw5m` con el total → cae (8).
  > Transcripciones de T009–T015, el censo declarado y las mutaciones nuevas M-B2a y M-B2b: [`soporte/registro.md`](./soporte/registro.md) §B2.
- [x] **T016** ✋ *(`5e3e5a0`)* Puertas y commit: `007 B2: la escritura de cache se tarifa por su duracion`.

## B3 · Máximo dentro de la pasada *(FR-009, FR-013, FR-016; SC-002, SC-003)*

- [x] **T017** Fase 0:
  - en `internal/ingest/pasada.go`, el acumulador por mensaje *(máximo por partida, desglose de la línea del máximo de la escritura,
    primera línea para el evento, D-007-P5)*, vacío y nil-seguro;
  - el método que cierra **al final de un fichero** y devuelve los cerrados con su desglose.
  - Compila; la suite sigue verde.
- [x] **T018** **Rojos (9) a (12)** en `internal/ingest/maximo_test.go` *(nuevo)*:
  - **(9)**: 7 → 1 303 → 89 817 → **un** evento con **89 817**;
  - **(10)**: 7 → 89 817 → 1 303 → 89 817;
  - **(11)** *(E-4: el máximo NO en la última)*: escritura 100 *(todo 5 m)* → 300 *(todo 1 h)* → 200 *(todo 5 m)* → desglose 0 / 300;
  - **(12)**: `Crecieron` = 1;
  - **(25)** *(E-4, empate)*: 300 *(todo 5 m)* → 300 *(todo 1 h)* → vale la última, 0 / 300.
  **Rojo (24)** *(E-4)* en `desglose_test.go`: una línea sin desglose y con escritura 0 **no** cuenta en `SinDesglose`; (7) sigue dando 2.
  **Rojo (13)** en `cmd/permea/coste_test.go`: `--scan` del mensaje de (9) → `out=89817`.
- [x] **T019** Censo de `internal/ingest/pasada_test.go`, transcribiendo antes y después:
  - `leerEnUnaPasada` recoge también los cerrados al final;
  - `TestCasoLimite_ConsumoDistinto` pasa de «la primera» a «el máximo» *(100/40 y 999/1 → 999/40)*, y su discrepancia se sigue contando.
- [x] **T020** **Verde**:
  - con pasada, `FromClaudeCodeLine` acumula y devuelve `nil`; **sin pasada, emite por línea como hoy** *(FR-018)*;
  - el coste se calcula al cerrar, con el máximo;
  - `generate()` encola los cerrados **al final de cada fichero, antes** de `st.Save` *(FR-012)*, y `dryRun()` los imprime.
- [x] **T021** **(14)**, que nace verde: duplicar una línea no cambia ni eventos ni sumas *(SC-003)*. La valida (m11).
- [x] **T022** Mutaciones:
  - **(m9)** máximo → primera: caen (9), (10), (11), (12), (13) y `ConsumoDistinto` *(censo en el registro)*;
  - **(m10)** máximo → última: caen (10) y `ConsumoDistinto`;
  - **(m11)** sumar: caen (9), (10), (12), (13), (14), `ConsumoDistinto` y `UnMensajeDeTresLineas…/tokens_de_una_linea`;
  - **(m12)** desglose de la última línea: cae sólo (11);
  - **(m23)** *(E-4)* contar sin desglose también con escritura 0: cae sólo (24);
  - **(m24)** *(E-4)* la primera entre las empatadas: cae sólo (25).
  > Censo, transcripciones y mutaciones: [`soporte/registro.md`](./soporte/registro.md) §B3.
- [ ] **T023** ✋ Puertas y commit: `007 B3: cada partida vale el maximo de sus lineas`.

## B4 · Retención: offset, cierre, `--run`, `--daemon` y resumen *(FR-010 a FR-015, FR-017, FR-019, FR-021; SC-006 a SC-009)*

- [x] **T024** Fase 0:
  - en `internal/state/state.go`, el método nuevo que da al callback el **comienzo** de cada línea y fija el offset que decide quien llama
    *(D-007-P3)*, con `ScanFile` como envoltorio;
  - el campo `reloj func() time.Time` en el `agent`, a `time.Now` por defecto *(D-007-P6)*.
  - Suite verde y `state_test.go` **sin tocar**.
- [x] **T025** [P] **R-B4a** en `internal/state/retener_test.go` *(nuevo)*: el offset guardado es el pedido, no el final, y `Size`
  y `ModTime` siguen siendo los del stat. ROJO mientras el método avance como hoy.
- [x] **T026** [P] Rojos de reglas en `internal/ingest/cierre_test.go` *(nuevo)*:
  - (i): en el mismo fichero cierra, y en otro no;
  - (ii): con `ahora` y `mtime` dados, los tres casos de SC-007;
  - (iii) *(E-3)*: con el `mtime` de ahora, el tope a 24 h y a 24 h − 1 s;
  - **(21)**: línea tardía contada y no emitida;
  - releídas, contadas aparte.
- [x] **T027** [P] **Rojos (15), (16), (17), (19), (20) y (23)** en `cmd/permea/retencion_test.go` *(nuevo, sandbox)*:
  - **(15)**, SC-006 con **dos procesos** `--run` *(sin enrolar, `LogsRoot` temporal)*:
    - el primero, con la línea parcial, no encola nada del mensaje, y el offset queda en su comienzo;
    - se añaden la final y la primera línea de otro mensaje;
    - el segundo encola **un** evento con la final;
  - **(16)**, SC-007 a, b y c: `generate()` con el reloj y `os.Chtimes`;
  - **(23)** *(E-3)*, el tope de SC-007, con el `mtime` de ahora: `timestamp` a 24 h − 1 s → retenido *(rojo)*; a 24 h → emitido
    *(nace verde, T031)*;
  - **(17)**, SC-008;
  - **(19)**: el aviso de `--run`, literal *(spec §Textos)*;
  - **(20)**: el predicado de `tick()` es falso si todo lo leído es releído y no se emitió nada.
- [x] **T028** **Rojo (18)** en `internal/ingest/pasada_test.go`, en el censo: `TestPasada_ElResumenNoLlevaIdentificadores` mira las
  **dos** líneas. La primera, **byte a byte** la de hoy; la segunda, la aprobada.
- [x] **T029** Censo de `cmd/permea/main_test.go`: `TestPasada_GenerateEncolaUnoPorMensaje` y `TestActualizar_NoReenviaNiReescribeLaCola`
  fijan el reloj a T + 1 s sobre el `timestamp` y el `mtime`. *(Ejecución: sin el reloj **no** daban rojo, porque la regla (iii) cierra sus
  líneas de 2026-10-02; dan el rojo previsto sólo sin la (iii). Ver el registro.)*
- [x] **T030** **Verde**:
  - reglas (i) y (ii), esta última con `max(timestamp, mtime)`;
  - el offset en el comienzo del abierto;
  - `--run` deja lo abierto y avisa;
  - `tick` con el predicado de FR-021;
  - el resumen en dos líneas;
  - el comentario de `pasada.go` sobre «entre pasadas no hay memoria» se reescribe: ahora lo abierto se relee.
- [x] **T031** Verdes de nacimiento, con su mutación:
  - **(22)** SC-009: un mensaje abierto con centinelas de `message.id` y `requestId`; tras la pasada, **ningún** fichero del directorio de
    datos los contiene, y `state.json` tiene sus cuatro campos. *(Ejecución: en la Fase 0 cae en su precondición, porque no hay
    abiertos; sus aserciones se cumplían. Las validan m19 y M-B4d.)*
  - **(16 a)** y **(23, 24 h)** *(E-3)*.
- [x] **T032** Mutaciones, con el censo declarado:
  - **(m13)** el offset avanza al final → caen (15, offset) y (20);
  - **(m14)** sin la condición del `mtime` → caen (16 b) y (23, 24 h − 1 s), y sus gemelos de `cierre_test.go`;
  - **(m15)** sin la del `timestamp` → cae (16 c);
  - **(m16)** la regla (i) mira todos los ficheros → cae (17), caso de otro fichero;
  - **(m17)** emitir lo abierto al acabar `--run` → caen (15) ×3, (16 b, c), (17) ×2, (19), (20), (22, precondición) y (23, 24 h − 1 s);
  - **(m18)** las releídas cuentan como nuevas → cae (20);
  - **(m19)** guardar el abierto en `pendientes.json` con sus identificadores → cae (22);
  - **(m20)** sin la regla (i) → caen (17, mismo fichero), (15) ×3, (21) y `ReglaI/mismo_fichero`. *(Ejecución: `project_test.go` **no**
    cae, porque la (iii) cierra su fixture de 2026-06-20.)*;
  - **(m21)** no emitir nunca por T → caen (16 a), `ReglaII/a` y los dos de T029 *(por hoja: `…Actualizar…/solo_lo_posterior_al_offset`)*;
  - **(m22)** *(E-3)* quitar el tope de 24 h → cae (23, 24 h), y sólo ése;
  - **M-B4a**: el método ignora el offset pedido → cae R-B4a.
  > Censo, transcripciones y resultado: [`soporte/registro.md`](./soporte/registro.md) §B4.
- [ ] **T033** ✋ Puertas y commit: `007 B4: un mensaje se envia cuando esta cerrado`.

## B5 · README y CHANGELOG *(FR-020; SC-011)*

- [x] **T034** **Rojo**, transcrito:
  - `grep -n "Limitación conocida" README.md` y `grep -n "Limitación 1" README.md` → > 0;
  - `grep -c '^## 0.4.0' CHANGELOG.md` → 0.
- [x] **T035** En `README.md`:
  - se retiran la limitación de los tokens de salida crecientes y la «Limitación 1»;
  - se declaran la **hipótesis P-1** y la espera del último mensaje *(10 minutos sin cambios, o la pasada siguiente)*;
  - la «Limitación 2» sigue.
- [x] **T036** En `CHANGELOG.md`, `## 0.4.0 — PENDIENTE` encima de la 0.3.0, con el **cuerpo aprobado literal y sus citas** *(spec §Textos,
  E-3)*. Comprobación: el cuerpo extraído de los dos ficheros, **citas incluidas**, comparado con `cmp`, sin diferencias. `grep -c PENDIENTE CHANGELOG.md` → 1.
  > Rojos, cambios, `cmp` y la aproximación de «releída»: [`soporte/registro.md`](./soporte/registro.md) §B5.
- [ ] **T037** ✋ Puertas y commit: `007 B5: README y CHANGELOG de la 0.4.0`.

## Cierre — en tramos, uno por mensaje *(plan §Cierre; si uno falla, se para y se rehace desde C1)*

- [x] **T038** **C1** · Puertas sobre la rama, transcritas: las del bloque más `go test` → **432 + nuevos** *(SC-012)*; `grep -rn nolint` → 1;
  la cabecera cita `8f147d1`; compilan Windows y darwin; `PENDIENTE` → 1.
- [x] **T039** **C2** · Medidas *(quickstart §Copias, §Contador, §M1–M4)*: huella de las dos copias antes; el contador; `--scan` en
  `env -i`; SC-001, SC-002, SC-003 y SC-005 contra las referencias de la spec; huella después, igual; temporales borrados.
- [x] **T040** **C3** · `goreleaser release --snapshot --clean`; SHA-256 del zip de Windows; `strings` del binario contiene los textos
  aprobados *(SC-011)*.
- [ ] **T041** ✋ **C4 · W1** *(quickstart §W1)*: `--version` primero; sólo después `status`, `--scan` y los dos `--run` separados por
  más de T. Se anota cada paso.
- [ ] **T042** **C5** · Cuerpo del PR y fecha del encabezado del CHANGELOG *(`PENDIENTE` → 0)*. ✋ commit
  `007 C5: fecha de la 0.4.0 en el CHANGELOG`.
- [ ] **T043** ✋ **C6** · Fusión del PR con merge commit.
- [ ] **T044** ✋ **C7** · `git tag -a v0.4.0` sobre `main` y `git push origin v0.4.0`.
- [ ] **T045** **C8** · Los tres canales en `0.4.0` *(release, Scoop y cask)*, y `strings` del binario publicado *(SC-011)*.
- [ ] **T046** ✋ **C9 · W2** *(quickstart §W2)*: Scoop, `--run`, recuento en la plataforma con «en espera» descontado; tras T, otro `--run`
  que cuadre con todos.

## Dependencias

- B0 → B1 → B2 → B3 → B4 → B5 → C1…C9, en serie: B2 necesita la quinta cifra; B3, el desglose; B4, el acumulador.
- Dentro de un bloque, los rojos `[P]` van antes del verde.

## Cobertura

| FR | Tareas | | SC | Tareas |
|---|---|---|---|---|
| 001 | T011, T013 | | 001 | T039 |
| 002 | T009, T010, T013 | | 002 | T018 (sintético), T039 (copia -windows) |
| 003, 004 | T011, T013, T015 | | 003 | T021, T039 |
| 005 | T011, T015 (m7) | | 004 | T010, T011 |
| 006, 008 | T005–T007 | | 005 | T012, T039 |
| 007 | T006, T013 | | 006 | T025, T027 (15), T032 |
| 009 | T018–T020, T022 | | 007 | T026, T027 (16, 23), T031, T032 |
| 010 | T026, T027, T030, T032 | | 008 | T026, T027 (17), T032 |
| 011, 012 | T024, T025, T030 | | 009 | T031, T038 + puertas |
| 013 | T020, T026 (21) | | 010 | T007 (m1) |
| 014 | T027 (19), T030 | | 011 | T028, T036, T040, T045 |
| 015 | T027 (16), T030 | | 012 | puertas de cada bloque, T038 |
| 016 | T012, T013, T018, T020 | | | |
| 017 | T028, T030 | | | |
| 018 | puertas de cada bloque, T014 | | | |
| 019 | T029 | | | |
| 020 | T035, T036, T042–T045 | | | |
| 021 | T027 (20), T030, T032 (m18) | | | |

**21 / 21 FR y 12 / 12 SC con tarea.**

## Mutaciones

| # | Bloque | Mutación | Debe caer *(y sólo eso)* |
|---|---|---|---|
| m1 | B1 | `CacheWrite1h` de `claude-fable-5-1` a 0 | `CifrasClaveAClave/claude-fable-5-1` |
| m2 | B1 | clave sobrante | `NingunaClaveSobra`, `RecuentoDeClaves` |
| m3 | B1 | cruzar 5 m/1 h en `claude-opus-5-5` | `CifrasClaveAClave/claude-opus-5-5`, `TestCost_Opus55AMano` *(declarada)* |
| m4 | B2 | cruzar las tarifas en `Cost` | (3) ×2, (4), (5), (6) *(su coste)* y `TestCost` *(declarado en T015)* |
| m5 | B2 | sin desglose a 5 min | (5), (6) |
| m6 | B2 | aceptar el desglose incoherente | (6) *(coste)*, (7) |
| m7 | B2 | `tokens_cache_creation` = 5 m + 1 h | (6) |
| m8 | B2 | `cw5m` con el total | (8) |
| M-B2a | B2 | el evento lleva `cw1h` como total | (4) *(tokens)*, (8) |
| M-B2b | B2 | sin desglose, total = 0 | (5) y (6) *(tokens)*, `TestScan_LineaConCuatroPartidasYEventID` |
| m9 | B3 | máximo → primera | (9), (10), (11), (12), (13), `ConsumoDistinto` |
| m10 | B3 | máximo → última | (10), `ConsumoDistinto` |
| m11 | B3 | sumar | (9), (10), (12), (13), (14), `ConsumoDistinto`, `UnMensajeDeTresLineas…` |
| m12 | B3 | desglose de la última | (11) |
| m23 | B3 | *(E-4)* sin desglose cuenta aunque la escritura sea 0 | (24) |
| m24 | B3 | *(E-4)* la primera entre las empatadas | (25) |
| m13 | B4 | offset al final | (15, offset), (20) |
| m14 | B4 | sin la condición del `mtime` | (16 b), (23, 24 h − 1 s) y sus gemelos de `cierre_test.go` |
| m15 | B4 | sin la del `timestamp` | (16 c) |
| m16 | B4 | regla (i) entre ficheros | (17), caso de otro fichero |
| m17 | B4 | emitir lo abierto al acabar `--run` | (15) ×3, (16 b, c), (17) ×2, (19), (20), (22), (23, 24 h − 1 s) |
| m18 | B4 | releídas como nuevas | (20) |
| m19 | B4 | `pendientes.json` con identificadores | (22) |
| m20 | B4 | sin la regla (i) | (17, mismo fichero), (15) ×3, (21), `ReglaI/mismo_fichero`; **no** `project_test.go` |
| m21 | B4 | no emitir nunca por T | (16 a), `ReglaII/a`, los dos de T029 |
| m22 | B4 | *(E-3)* sin el tope de 24 h | (23, 24 h) |
| M-B4a | B4 | el método ignora el offset pedido | R-B4a ×2, (15, offset), (20) |
| M-B4b | B4 | `Size` = offset pedido | R-B4a ×2, (20) |
| M-B4d | B4 | `FileState` con un quinto campo | (22, cuatro campos) |

## Lo que este plan de tareas NO hace

- No toca `internal/event`, la derivación del `event_id`, `boundary_test.go` ni `state_test.go`.
- No cambia la plataforma. La cabecera de `pricing.php` se enmienda en un encargo aparte *(N-6)*.
- No lanza `enroll`, `--run` ni `--daemon` sobre la instalación real. Los `--run` de los tests van en sandbox, y los de W1 y W2 los
  lanza el dueño.
