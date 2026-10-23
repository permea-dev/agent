# Tasks: 013 · «Lista de proyectos»

**Feature**: `013-lista-de-proyectos` · **Fecha**: 2026-10-23 · [spec.md](./spec.md) *(ratificada el 2026-10-23)* · [plan.md](./plan.md)
**Base**: el commit de la spec ratificada *(el código, el de `cabf8c5`)*. **Línea base a preservar** *(sobre `cabf8c5`, se transcribe en T001)*:
`go test -count=1 -v ./...` → **1074** «--- PASS», 0 FAIL, 0 SKIP · `go vet` limpio · `golangci-lint run` → **0**. Los tamaños de la 0.9.0, en T001.
**44 tareas · 12 ✋ · 26 tests nuevos** *(25 rojos y 1 nacido verde)* **· 20 mutaciones** *(plan §Bloques)*.
*Unidades*: aquí, «tests» son los rojos **numerados** del plan, (1)–(26). El registro cuenta **hojas**.

## Formato: `[ID] [P?] [✋?] Descripción`

- **[P]**: sin dependencia de decisión entre sí. No quiere decir que se editen a la vez.
- **✋**: la ejecuta **el dueño**: commits, fusión, etiqueta, aprobación de textos y W1. Claude no hace git de escritura. Cada commit lleva su
  mensaje previsto, **sin tildes ni ñ**.
- **(1)…(26)** son los rojos, y **m1…m20** las mutaciones de plan §Bloques, cada una con su sonda. Lo que se añada aquí se llama **M-B3a**, **M-B7a**…
- **`[MUTACIÓN]`** marca las tareas que mutan. **No se renumera**.
- **Las transcripciones** van a `soporte/registro.md`, que **nace en T001**, y cada tarea deja una remisión.
- **Bloques de 45–60 min** de Claude *(plan §Estimación)*; en cada uno, **el rojo antes del verde**, con la razón real transcrita.

## Presupuesto *(techos **fijados por el dueño** el 23-10, como se propusieron; ninguno se sube: si no cabe, se condensa)*

| Fichero de `specs/013-lista-de-proyectos/` | Techo | Hoy |
|---|---:|---:|
| `spec.md` | 320 | 304 |
| `plan.md` | 280 | 269 |
| `tasks.md` *(este)* | 220 | 217 |
| `soporte/registro.md` *(nace en T001)* | 250 | 86 |
| `soporte/descubrimiento.md` | cerrado | 262 |
| `soporte/registro-archivo.md` *(§B0–§B6, desde el 23-10)* | sin techo | 394 |

## Una suite a la vez, con la carga vigilada

Antes de **cada** `go test` o `golangci-lint`, se lanza **`~/bin/libre.sh` en una orden sola**, como en 012 *(012 `tasks.md` §Una suite a la vez)*: 0
procesos de suite, carga de 1 min **< 2,0** en 3 muestras y tope de 45 min. Al tope, **PARA 9**.

## 🛑 Lo que PARA

1. **Un `_test.go` existente** necesita cambiar *(censo de plan §7: ninguno)*, o se toca un fichero fuera de plan §Project Structure.
2. **En una mutación cae algo que no se declaró**, o no cae lo declarado. **La mutación se queda puesta** y se para.
3. **El md5 de la reversión** no es el de antes.
4. **La frontera de FR-025 tiene diff.**
5. **La lista escribe algo o abre la red**: (24) o (25) en rojo tras el verde, o una huella distinta en C2.
6. **La emisión cambia**: (5) en rojo, o un test de los 1074 en FAIL.
7. **Un `require` en `go.mod`**, o un `go.sum`.
8. **Una ruta real, un nombre de carpeta o un contenido** en testdata, en el registro o en un informe. En C2, sólo recuentos, tiempos y md5. `PRIV` → 0,
   también con `grep -a`.
9. **`~/bin/libre.sh` llega a su tope.**
10. **SC-016 pasa de 51 200 B** en alguna plataforma en C1.

## Puertas en cada corte

```sh
gofmt -l . ; go vet ./... ; golangci-lint run                                  # → vacío, limpio, 0 issues
go test -count=1 ./...                                                         # → 1074 + los nuevos, 0 FAIL, 0 SKIP; tras ~/bin/libre.sh
git diff cabf8c5 -- internal/event/ internal/ingest/ internal/transport/ internal/state/ internal/pricing/ \
  specs/005-adhesion-a-proyecto/contracts/ specs/006-medicion-fiel/contracts/ specs/008-lector-codex/contracts/ \
  specs/009-lector-gemini/contracts/ specs/010-lector-opencode/contracts/ specs/011-lector-copilot/contracts/ \
  specs/012-lector-kimi/contracts/                                             # → vacío
git diff cabf8c5 --name-only --diff-filter=M -- '*_test.go'                    # → vacío
grep -c require go.mod ; ls go.sum                                             # → 0 ; no existe
```

**Vigilancia de privacidad**: el patrón `PRIV` de 012 *(`tasks.md` §Puertas)*, con el usuario y el nombre sacados del entorno:
`git diff cabf8c5 | grep -cE "$PRIV"` → 0 y `grep -raEc "$PRIV" cmd/permea/testdata/lista/` → 0.

**Protocolo de mutación** *(006, con E-5 de 007)*: el censo, **por hoja y antes de mutar**; la suite entera; si los `FAIL` = lo declarado, reversión por
edición inversa y md5; si no, PARA 2. Una mutación que no compila no cuenta.

---

## B0 · Fase 0, referencia y registro

- [x] **T001** Fase 0 sobre `cabf8c5` por `git archive` en un temporal: la línea base *(1074, vet, lint 0)*, los **tamaños de la 0.9.0** en las 5
  plataformas *(D-013-P7)* y la frontera sin diff. **Nace `soporte/registro.md`**, con su techo y las secciones §B0…§B8 y §Cierre.
  > Transcripción: [`soporte/registro-archivo.md`](./soporte/registro-archivo.md) §B0 · T001 *(1074 / 0 / 0; vet y lint 0, con lint 2.15.0; frontera 0; `PRIV` 0; los cinco tamaños, que no cuadran con la 012: SC-016 recompila las dos en C1)*.
- [x] **T002** **Nacen `cmd/permea/testdata/lista/generar.sh` y sus plantillas** *(plan §9)*, y **la referencia de SC-012**: `--run` del binario de
  `cabf8c5` sobre los seis fixtures, en sandbox con `SandboxConSemillas`. La cola, el stdout y el stderr van a `testdata/lista/referencia-run.*`.
  md5 al registro *(D-013-P6)*. `PRIV` → 0.
  > Transcripción: `registro-archivo.md` §B0 · T002 *(`generar.sh` `96304965…`, 10 plantillas; la referencia, determinista: cola `e6b011db…` con 17 eventos de los seis lectores, stdout vacío y stderr `1799737c…`; `PRIV` 0, también con `-a`)*.
- [ ] **T003** ✋ Commit: `013 B0: linea base, referencia y registro`.

## B1 · Leer sin escribir *(FR-005, FR-007, FR-008)* — nace `internal/config/lectura_test.go`

- [x] **T004** Fase 0: `RutaDataDir()` y `LeerSalt()`, que devuelven `"", nil` y `"", false, nil`.
- [x] **T005** [P] **Rojos (1) y (2)**, con `testutil.Sandbox`.
- [x] **T006** **Verde**: plan §3. `DataDir()` llama a `RutaDataDir()` y conserva su conducta.
- [x] **T007** `[MUTACIÓN]` Censo y **m1–m3**, con md5. Transcribir.
  > Transcripción: [`soporte/registro-archivo.md`](./soporte/registro-archivo.md) §B1 *(m1–m3 = declarado, revertidas con md5; **M-B1a**: 1.ª forma PARA 2 por censo incompleto; la 2.ª, con los tres `TestEnroll_*`, = declarado; corte 1084 / 0 / 0, vet y lint 0)*.
- [ ] **T008** ✋ Puertas y commit: `013 B1: leer el DataDir y la sal sin crearlos`.

## B2 · La derivación con origen *(FR-010, FR-015)* — nace `internal/project/origen_test.go`

- [x] **T009** Fase 0: `Origen`, `DerivarConOrigen`, que devuelve `"", "", false`, y `Resolutor.Origen`, que devuelve `Origen{}, false`.
- [x] **T010** [P] **Rojos (3) y (4)**, sobre copias en `t.TempDir()` de las familias de `internal/project/testdata`.
- [x] **T011** **Verde**: plan §2. **Un solo cuerpo**: `derivarConTechoYRaiz` devuelve el origen, y no nace una segunda derivación.
- [x] **T012** `[MUTACIÓN]` Censo y **m4–m6**, con md5. Transcribir que `resolve_test.go` sigue verde **sin tocarlo**.
  > Transcripción: [`soporte/registro-archivo.md`](./soporte/registro-archivo.md) §B2 *(rojos 1084 / 10 / 0 = declarado; verde 1094 / 0 / 0; SC-012 repetida, byte a byte; m4–m6 = declarado, con el censo por llamantes; `resolve_test.go` 0 B y verde)*.
- [ ] **T013** ✋ Puertas y commit: `013 B2: la derivacion dice que ruta entro en el hash`.

## B3 · El recorrido común *(FR-009, FR-024; SC-012)* — nace `cmd/permea/recorrido_test.go`

- [x] **T014** **(5)**, **antes de tocar `main.go`** *(con B1 y B2 ya hechos, que no tocan la emisión)*: `--run` = la referencia de T002, **en verde**
  *(nace verde; lo valida m8)*.
- [x] **T015** Fase 0 y **rojo (6)**: `sumidero` y un `recorrer` que aún no recorre.
- [x] **T016** **Verde**: plan §1. Primero, `recorrer` con el cuerpo de `generate()` y `sumar` en los siete puntos de encolado. Después, `generate()`
  sobre `recorrer` y, por último, `resolverRaices` fuera de `setup()`. **(5) y (6) en verde**, y los 1074.
- [x] **T017** `[MUTACIÓN]` Censo *(m8, por nombre: los `TestCodex_*` que pasan por `--run`)* y **m7–m8**, con md5.
  > Transcripción: [`soporte/registro-archivo.md`](./soporte/registro-archivo.md) §B3 *(m7 = declarado; m8: 1.ª forma PARA 2, censo sin bajar a las hojas; la 2.ª, con 15 hojas, = declarado; corte 1101 / 0 / 0, vet y lint 0)*.
- [ ] **T018** ✋ Puertas y commit: `013 B3: un solo recorrido de los seis lectores con sumidero`.

## B4 · Agregación y nombres *(FR-011 a FR-014, FR-016, FR-020)* — nace `cmd/permea/lista_agregar_test.go`

- [x] **T019** Fase 0: `fila`, `agregar`, que devuelve `nil`, y la tabla de nombres vacía.
- [x] **T020** [P] **Rojos (7) a (12)**, con eventos hechos a mano y `origen` y `existe` inyectados.
- [x] **T021** **Verde**: plan §4 y §5. La tabla de nombres, con los seis de D-1 en el orden de P-3.
- [x] **T022** `[MUTACIÓN]` Censo y **m9–m11**, con md5.
  > Transcripción: [`soporte/registro-archivo.md`](./soporte/registro-archivo.md) §B4 *(rojos 1101 / 16 / 0 = declarado; verde 1117 / 0 / 0; m9 interrumpida sin resultado y repetida; m9–m11 = declarado, por hoja; corte 1117 / 0 / 0, vet y lint 0)*.
- [ ] **T023** ✋ Puertas y commit: `013 B4: agregar por identidad con los nombres comerciales`.

## B5 · La tabla *(FR-017 a FR-022; SC-001, SC-010, SC-011)* — nace `cmd/permea/lista_formato_test.go`

- [x] **T024** Fase 0: `formatear`, `miles` y `aCasa`, que devuelven `""`.
- [x] **T025** [P] **Rojos (13) a (16)**. (13), con el bloque de spec §Textos **copiado por programa** *(`awk` sobre `spec.md`)*, no a mano.
- [x] **T026** **Verde**: plan §6.
- [x] **T027** `[MUTACIÓN]` Censo y **m12–m14**, con md5. m12 cae en (13) porque `…` ocupa 3 bytes: se declara así.
  > Transcripción: [`soporte/registro-archivo.md`](./soporte/registro-archivo.md) §B5 *(rojos 1117 / 17 / 0 = declarado; verde 1134 / 0 / 0; m12–m14 = declarado, por hoja; `libre.sh` como hijo directo; corte repetido tras `revive` y `PRIV`: 1134 / 0 / 0, lint 0)*.
- [ ] **T028** ✋ Puertas y commit: `013 B5: la tabla de project list`.

## B6 · El comando *(FR-001 a FR-003, FR-006 a FR-008, FR-023; SC-006, SC-013, SC-014)* — nace `cmd/permea/lista_test.go`

- [x] **T029** Fase 0 y **rojos (17) a (21)**: `case "list"` en `runProject` *(`project.go:103`)*, con un `projectList` *(en `lista.go`)* que sale con 0
  sin hacer nada.
- [x] **T030** **Verde**: `projectList` hasta la tubería: RutaDataDir → LeerSalt → Load. Con sal, **sale con 0 sin escribir nada**: la tubería la conecta
  B7. Más los dos mensajes
  de verbos *(`project.go:98`, `:108`)*, la fila de ayuda y `ayudasDeSubcomando["project list"]`, con `"project"` → `[]int{2, 3}` *(`ayuda.go:115-120`)*.
  Los textos, **sacados por programa** de spec §Textos.
- [x] **T031** `[MUTACIÓN]` Censo y **m15–m17**, con md5. Transcribir que los `TestAyuda_*` y `TestProject*` existentes siguen verdes **sin tocarlos**.
  > Transcripción: [`soporte/registro-archivo.md`](./soporte/registro-archivo.md) §B6 *(rojos 1135 / 16 / 0 = declarado, con (21/plain) nacido verde; verde 1151 / 0 / 0; m15–m17 = declarado, por hoja; los existentes, verdes sin tocar; corte 1151 / 0 / 0, lint 0, `PRIV` 0 tras corregir §B5)*.
- [ ] **T032** ✋ Puertas y commit: `013 B6: el verbo project list`.

## B7 · Los seis lectores y la frontera *(FR-004, FR-005, FR-009, FR-010, FR-015, FR-016; SC-002 a SC-005, SC-007, SC-009)* — nace `cmd/permea/lista_frontera_test.go`

- [x] **T033** **Rojos (22) a (26)** sobre el binario de prueba. Los cinco afirman la tabla, que la tubería vacía de B6 no escribe: son **rojos**, con
  la razón transcrita. (24) monta el servidor de confianza por `SSL_CERT_FILE` *(plan §8)*.
- [x] **T034** **Verde**: la tubería en `projectList`: resolverRaices → `ClaudeCodeLogsRoot` y `FindLogs` → `recorrer` con `state.New()` y su
  `Resolutor` → agregar → formatear *(plan §1, §4 y §6)*. Si algo exige tocar `internal/ingest/`, **PARA 4**.
- [x] **T035** `[MUTACIÓN]` Censo y **m18–m20**, con md5. Vigilancia de privacidad → 0.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B7 *(rojos 1156 / 10 / 0 = declarado, 5 hojas nacidas verdes; verde 1166 / 0 / 0; m18–m20 y M-B7a = declarado, por hoja; corte 1166 / 0 / 0, lint 0, `PRIV` 0 al final)*.
- [ ] **T036** ✋ Puertas y commit: `013 B7: project list no abre la red ni escribe nada`.

## B8 · README y CHANGELOG *(spec §Limitaciones, Q-1)*

- [x] **T037** ✋ **El dueño aprueba los textos**: «`permea project list`» en el README y la entrada `0.10.0` en el CHANGELOG, con las limitaciones de
  spec §Limitaciones. Claude los propone en un encargo aparte. **Rojo**: `grep -c '^## 0.10.0'` y `grep -c 'project list' README.md` → 0. **Verde**: los
  textos **sacados por programa**, y `cmp` sin diferencias. Commit ✋: `013 B8: README y CHANGELOG`.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B8 *(textos de B8a **aprobados tal cual por el dueño el 2026-10-23**; rojo 0 y 0; verde 1 y 4, con los tres `cmp` y el de la tabla contra spec §Textos sin diferencias; 1166 / 0 / 0)*.

## Cierre — en tramos, uno por mensaje *(plan §Cierre)*

- [ ] **T038** **C1** · Puertas, con 0 SKIP: los 1074 **por nombre**; SC-015; **SC-016** en las 5; SC-017; compilan Windows y darwin.
- [ ] **T039** **C2** · **SC-018** *(D-013-W)*: `time permea project list | wc -l` sobre la instalación de WSL, con la huella md5 del `DataDir` antes y
  después. Al registro van sólo el tiempo, las líneas y si la huella es igual. **C3** · `goreleaser release --snapshot --clean` y `strings` con los textos.
- [ ] **T040** ✋ **C4 · W1** *(SC-019)*: en Windows, con el panel delante, cada identidad sin asignar de la instalación, contra la lista. Q-2: el tiempo.
- [ ] **T041** **C5** · El cuerpo del PR, fuera del repo, y la fecha de la versión en la rama.
- [ ] **T042** ✋ **C6** · Fusión del PR con merge commit.
- [ ] **T043** ✋ **C7** · La etiqueta `v0.10.0`; el workflow `release` en verde.
- [ ] **T044** **C8** · Los canales en la versión nueva, `sha256sum -c`, y `strings` del binario publicado.

## Dependencias

- **Bloques**: B0 → B1 → B2 → B3 → B4 → B5 → B6 → B7 → B8 → C1…C8. B4 y B5 no dependen de B3 y pueden adelantarse. B6 usa B1, B3, B4 y B5; B7, todo.
- **Dentro de un bloque**: los rojos `[P]` antes del verde, y el censo antes de mutar.
- **Fuera del repo**: el nombre `copilot_cli` en el panel *(spec §Dependencias)* **no bloquea**.

## Censo de tests *(plan §7; ninguno existente cambia)*

| Tarea | Nace |
|---|---|
| T005 | `internal/config/lectura_test.go` |
| T010 | `internal/project/origen_test.go` |
| T014, T015 | `cmd/permea/recorrido_test.go` |
| T020 | `cmd/permea/lista_agregar_test.go` |
| T025 | `cmd/permea/lista_formato_test.go` |
| T029 | `cmd/permea/lista_test.go` |
| T033 | `cmd/permea/lista_frontera_test.go` |

## Cobertura *(SC → tarea que lo acredita)*

| SC | Tareas | | SC | Tareas | | SC | Tareas |
|---|---|---|---|---|---|---|---|
| 001 | T025 (13) | | 008 | T020 (10, 12) | | 015 | puertas de cada corte, **T038** |
| 002 | T033 (22) | | 009 | T020 (7), T033 (26) | | 016 | **T038** |
| 003 | T033 (23) | | 010 | T025 (14) | | 017 | T002, T035, **T038** |
| 004 | T033 (24) | | 011 | T025 (15) | | 018 | **T039** |
| 005 | T005 (1, 2), T033 (25) | | 012 | T002, T014 (5), T016, **T038** | | 019 | **T040** *(W1)* |
| 006 | T029 (17) | | 013 | T029 (19), **T039** *(C3)* | | | |
| 007 | T020 (11), T033 (22) | | 014 | T029 (18) | | | |

**25 / 25 FR** *(plan §Trazabilidad)* **y 19 / 19 SC con tarea.**

## Mutaciones — **20** *(plan §Bloques; la sonda, allí)*

| Bloque | Tarea | Mutaciones |
|---|---|---|
| B1 | T007 | m1–m3 *(DataDir sin crear, sal sin crear, sal tal cual)* |
| B2 | T012 | m4–m6 *(origen = raíz, `Clean`, clave por identidad)* |
| B3 | T017 | m7–m8 *(el sumidero en todos los lectores; la emisión igual)* |
| B4 | T022 | m9–m11 *(`event_id` distintos, UTC, la marca que gana)* |
| B5 | T027 | m12–m14 *(runas, miles, segmentos)* |
| B6 | T031 | m15–m17 *(sal, verbos, `plain`)* |
| B7 | T035 | m18–m20 *(sin `setup()`, sin red, con resolutor)* |

## Lo que este plan de tareas NO hace

- No toca `internal/event`, `internal/ingest`, `internal/state`, `internal/transport`, `internal/pricing`, `boundary_test.go` ni ningún `_test.go` existente.
- No cambia `project join` *(D-9)* ni la plataforma.
- No lanza `--run` ni `--daemon` sobre la instalación real. `project list` sí se ejecuta sobre ella, en C2, que es de sólo lectura, con stdout a `wc -l`.
- No pone en el repo, ni en el registro, una ruta, una carpeta o un contenido reales.
