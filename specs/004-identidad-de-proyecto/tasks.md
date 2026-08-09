---

description: "Task list · P-004 Identidad de proyecto"
---

# Tasks: Identidad de proyecto

**Input**: Design documents from `/specs/004-identidad-de-proyecto/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md),
[data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: **REQUERIDOS, no opcionales.** La constitución (Principio IV, NO NEGOCIABLE) exige
test-first en la frontera: *«el golden test […] es disciplina de primer commit y DEBE existir antes
que cualquier parser»*. Cada tarea de test lleva escrita **su garantía del contrato**.

**Organization**: por historia de usuario (US1 P1, US2 P2, US3 P3), cada una entregable e
independientemente testeable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: paralelizable (ficheros distintos, sin dependencias pendientes)
- **[Story]**: US1 / US2 / US3 — solo en fases de historia
- Ruta de fichero exacta en cada tarea

## Path Conventions

Proyecto único Go. Rutas desde la raíz del repositorio: `cmd/permea/`, `internal/`,
`specs/004-identidad-de-proyecto/`.

---

## Disciplinas transversales — aplican a TODA tarea de test de este fichero

No se repiten en cada tarea; se dan por incluidas en todas.

1. **Una garantía por tarea**, anclada al **contrato** (`contracts/project-identity.md` G1..G10 y
   casos 1..15/7b; `contracts/cli-config.md` tabla de invocaciones), **nunca** a la redacción de la
   spec. Si el contrato y la spec discrepan, se para y se reporta — no se elige.
2. **Rojo antes de verde.** El encargo que observa el fallo **transcribe la razón del fallo** (el
   mensaje real, no «falla como se esperaba»).
3. **Todo test que nazca verde se valida por mutación**, y la validación exige **leer el mensaje de
   fallo** que produce la mutación. La mutación se revierte **por edición inversa**, nunca por
   `git checkout` — para que el diff final demuestre que se revirtió lo mismo que se introdujo.
4. **Tests de proceso**: comparar `ExitCode()`, **nunca** texto (puente Windows/WSL).
5. **La ausencia de aviso se comprueba por `stderr` VACÍO**, no por *matching* de mensaje: comprobar
   que un texto no aparece pasa también cuando aparece otro texto distinto.
6. **Aislamiento obligatorio**: todo test que toque config o cola usa `HOME`/`USERPROFILE`/
   `XDG_CONFIG_HOME` apuntando a un temporal (quickstart §Aislamiento). Ningún test escribe jamás en
   la instalación real.

**Nota de ejecución**: Claude Code **no ejecuta git de escritura**. El marcado de casillas de este
fichero viaja en **el mismo commit** que el código que documenta, y ese commit lo hace Basilio.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: capturar la referencia de regresión y preparar los fixtures. **Sin ningún cambio de
comportamiento.**

- [x] T001 Capturar la línea base de SC-004 por la **VÍA B** en `specs/004-identidad-de-proyecto/baseline-sc004.tsv`, con los **TRES** refs por evento (`project_ref`, `session_ref`, `machine_ref`) — el `project_ref` no es para SC-004, es la columna contra la que T007 demostrará su neutralidad. **PRIMERA TAREA; SE COMMITEA ANTES DE CUALQUIER CAMBIO DE COMPORTAMIENTO**: sin ella, SC-004 compararía el resultado consigo mismo. La puerta de decisión está **resuelta en negativo** (`cmd/permea/main.go:333-334` imprime solo `project_ref`, y truncado a 8 caracteres), así que la vía A queda descartada y **no se reevalúa**. Receta:
  1. **Sandbox** de `quickstart.md` §Aislamiento, incluida su comprobación de que `status` dice «no enrolado».
  2. **Siembra del sandbox — `config.json` Y LAS DOS SEMILLAS.**
     a. **`logs_root` a un directorio EFÍMERO del sandbox**, al que se copia —como paso previo— **una copia fresca de `internal/ingest/testdata/claude_code_sample.jsonl` y NADA MÁS**. **Nunca se apunta `logs_root` a `internal/ingest/testdata/`**: `state.FindLogs` recorre **todos** los `.jsonl` bajo la raíz que se le dé (`internal/state/scan.go:20-31`), así que apuntar a una carpeta compartida del repositorio hace que **cualquier fixture futuro contamine la pasada**. Ya ocurrió: el `boundary_sample.jsonl` de T004 cayó en esa carpeta y una verificación de neutralidad pasó de 2 eventos a 4. **El fichero estaba congelado; el directorio no.** Esta receta congela el **directorio** por construcción: es efímero, lo crea la propia pasada, y solo contiene lo que ella pone.
     b. `config.json` bien formada, porque una config incompleta no ejercita el camino real: `endpoint` **https sintácticamente válido** —`config.Validate()` (`internal/config/config.go:99-101`) rechaza cualquier otro esquema— y `device_token` sintáctico; `logs_root` apuntando al directorio efímero del punto anterior, **jamás a los logs reales del desarrollador**.
     c. **`salt` y `machine_id` DETERMINISTAS, escritos ANTES de la pasada.** `LoadOrCreateSalt` genera un valor **aleatorio** si el fichero no existe (`internal/config/identity.go:13-14`, `:32-38`) y el sandbox nace vacío; como los tres refs son `Ref(salt, valor)`, **con un salt aleatorio la línea base no vuelve a salir jamás** y la comparación de T007/V8 sería imposible. `loadOrCreateSecret` **lee el fichero si existe** (`identity.go:24-28`), así que sembrarlos basta. Los valores son los documentados en el bloque **REPRODUCCIÓN** de `specs/004-identidad-de-proyecto/baseline-sc004.tsv` — **ese fichero es la única fuente de verdad**; no se copian aquí para que no puedan divergir.
  3. **Endpoint irrecuperable rápido**: `https://127.0.0.1:1`. El puerto 1 rechaza la conexión de inmediato, así que la pasada no queda colgando en reintentos ni depende de la red.
  4. **Pasada real**: `./bin/permea --run`. **Se espera `ExitCode() != 0`** —`sync()` devuelve el error de transporte y `runOnce()` lo propaga (`cmd/permea/main.go:214-245`)—; **eso no es un fallo de la captura**, es el envío fallido que la receta busca. Confundirlo con un error de la tarea es el malentendido probable.
  5. **VERIFICACIÓN OBLIGATORIA — la premisa de la vía B**: comprobar que `queue.jsonl` **retiene** los eventos tras el envío fallido. La premisa está respaldada por `internal/transport/queue_test.go:40` (`TestQueue_AtomicRewrite_KeepsUnconfirmed`) y por `rewriteQueue` (`internal/transport/queue.go:128`), pero **se comprueba aquí de todos modos**. **Si la cola quedara vacía: PARAR y reportar** — significaría que la vía B no funciona, y ese es exactamente el sitio donde hay que saberlo, no tres tareas después.
  6. Extraer los tres refs de `queue.jsonl` al `.tsv`, ordenado y sin duplicados — **el `.tsv` deduplicado es la referencia de CONJUNTOS** (SC-004: qué identidades existen).
  7. **Capturar TAMBIÉN el recuento de eventos** (`wc -l` de `queue.jsonl`) **como metadato del mismo fichero** —cabecera comentada o fila de metadatos—, porque **es la referencia de SC-003** («el 100 % de los eventos que hoy se emiten se siguen emitiendo»). El `sort -u` del paso 6 **colapsa** los eventos que comparten `session_ref` y `machine_ref`, que es lo normal dentro de una sesión: sin el recuento aparte, la línea base no puede detectar eventos perdidos. Las dos cifras miden cosas distintas y ninguna sustituye a la otra.
- [x] T002 [P] Crear el helper de sandbox de tests en `internal/testutil/sandbox.go` — establece `HOME`/`USERPROFILE`/`XDG_CONFIG_HOME` a un `t.TempDir()` y devuelve la ruta del dataDir resultante; incluye la aserción de aislamiento (un `config.DataDir()` que caiga fuera del temporal es fallo inmediato del test, no aviso).
- [x] T003 [P] Crear los fixtures JSONL de los casos de derivación en `internal/project/testdata/` — un fichero por familia (raíz/subdirectorio, worktree, sintácticos, enlaces, degradados), con rutas parametrizables por el test para poder apuntarlas al `t.TempDir()`. **Consumidores**: el golden ampliado (T004/T005) y las validaciones manuales de `quickstart.md` V1–V6. **NO los consumen los tests unitarios** de T008..T020, que construyen sus árboles con `t.TempDir()` y no necesitan JSONL: un fixture en disco con rutas fijas no podría ejercitar el techo del `HOME` falso ni los worktrees reales.

**Checkpoint**: `baseline-sc004.tsv` commiteado. El comportamiento del agente **no ha cambiado**.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: la frontera se refuerza **antes** de tocar la derivación (Principio IV), y se crea el
punto de extensión que permitirá que los tests de US1 nazcan en rojo.

**⚠️ BLOQUEA TODAS LAS HISTORIAS**: ninguna tarea de US1/US2/US3 empieza hasta cerrar esta fase.

- [x] T004 Ampliar la `denylist` del golden en `internal/ingest/boundary_test.go:16-33` con **fragmentos de raíz de proyecto y de subdirectorio de lanzamiento POR SEPARADO** (dos entradas distintas, no una ruta completa). El `cwd` que los contiene va a un **fixture DEDICADO** —`internal/ingest/testdata/boundary_sample.jsonl`— o a un registro construido dentro del propio test. **Garantía: G9** (la salida nunca contiene la ruta ni un fragmento de ella). Nace **verde** —hoy nada filtra—, así que **se valida por mutación**: hacer que la derivación emita el `cwd` en claro, **leer y transcribir el mensaje de fallo**, y revertir por edición inversa.
  ⚠️ **`internal/ingest/testdata/claude_code_sample.jsonl` NO SE TOCA EN TODA LA FEATURE.** Quedó congelado en T001 como **conjunto de referencia** de la línea base: sus 2 eventos facturables son los que producen `baseline-sc004.tsv`. Ampliarlo cambiaría el conjunto y **rompería la comparación de neutralidad de T007 y la regresión-cero de V8 por la razón equivocada** — parecería que la derivación cambió cuando lo que cambió fue la entrada. Un fixture nuevo para la frontera cuesta un fichero; recuperar una línea base contaminada cuesta rehacer T001 y descubrir por qué tarde.
- [x] T005 Ampliar el alcance del golden en `internal/ingest/boundary_test.go` de «evento serializado» a **evento serializado + `queue.jsonl` + cuerpo transmitido** (`httptest.NewTLSServer`, patrón de `specs/001-agente-inicial/contracts/transport.md:58`). **Garantía: G9 en los tres caminos hacia el exterior** (FR-017 tras D-004-4). Misma disciplina de mutación que T004, ejercitada **en cada uno de los tres caminos** — una mutación que solo falle en el primero deja los otros dos sin cobertura demostrada.
- [x] T006 Crear el paquete `internal/project/` con `resolve.go` exponiendo `Derivar(cwdDeclarado, salt string) string`, **implementado de forma que reproduzca EXACTAMENTE el comportamiento actual** (`event.Ref(salt, cwd)` sin observación). Es el punto de extensión: con él, los tests de US1/US2 **nacen en rojo por la razón correcta** —la identidad no agrupa— en vez de por «función no existe».
- [x] T007 Sustituir en `internal/ingest/claudecode.go:78` la expresión `event.Ref(ctx.Salt, r.Cwd)` por la llamada a `project.Derivar(r.Cwd, ctx.Salt)`, **sin cambiar el resultado**. Las líneas `:79-80` (`SessionRef`, `MachineRef`) **NO se tocan** (FR-019). **Verificación de neutralidad, contra artefacto y no contra un recuerdo**: repetir la pasada de T001 **reutilizando EXACTAMENTE las semillas del bloque REPRODUCCIÓN de `specs/004-identidad-de-proyecto/baseline-sc004.tsv`** (`salt` y `machine_id`; con otras, los refs no comparan y el «fallo» no significaría nada) y comparar las **tres** columnas de ese mismo fichero — incluida la de `project_ref`, que T001 capturó precisamente para esto. Las tres deben coincidir. Decir «idéntico al previo» sin un fichero con el que comparar no es una verificación.

**Checkpoint**: frontera reforzada y punto de extensión colocado. **El comportamiento sigue sin
cambiar** — y eso es comprobable, que es justo lo que hace útil a T007.

---

## Phase 3: User Story 1 — Todo el trabajo de un proyecto cuenta como un proyecto (P1) 🎯 MVP

**Goal**: dos eventos originados en cualquier punto de un mismo proyecto reciben la misma identidad.

**Independent Test**: procesar eventos originados en distintos puntos de un mismo proyecto y
comprobar que comparten identidad, sin necesidad de US2 ni US3.

### Tests de US1 (rojo antes de verde)

- [x] T008 [P] [US1] Test de raíz y subdirectorio en `internal/project/resolve_test.go` — **Garantía G3**, casos **2 y 3** de `contracts/project-identity.md`. Nace en **rojo**: con T006 la identidad es la del `cwd`, así que raíz y subdirectorio difieren. Transcribir la razón del fallo.
- [x] T009 [P] [US1] Test de contra-prueba de proyectos distintos en `internal/project/resolve_test.go` — **Garantía G4**, caso **6** (anidamiento: gana el más cercano). Sin este test, T008 se podría satisfacer devolviendo una constante.
- [x] T010 [P] [US1] Test de **árbol de trabajo paralelo** en `internal/project/resolve_test.go` — **Garantía: casos 4 y 5**. Crea un repositorio y un `git worktree` reales en `t.TempDir()` y asevera: (a) el marcador del worktree es un **fichero**, no un directorio (aserción explícita, para que el test documente por qué existe); (b) dos árboles paralelos dan identidades **distintas**; (c) raíz y subdirectorio **dentro** del worktree comparten identidad — que es lo que impide aprobar el caso 5 por el motivo equivocado. **Si `git` no está disponible, el test FALLA con un mensaje claro que lo diga — NUNCA `t.Skip`**: este es el único test que cubre la promesa de los árboles paralelos, y saltárselo en silencio dejaría la promesa sin cobertura mientras la suite sigue diciendo «verde». *(La implementación no necesita git — usa `os.Lstat`, decisión de research §P1; solo lo necesita este test, para fabricar el árbol.)*
- [x] T011 [P] [US1] Test de **techo del directorio personal** en `internal/project/resolve_test.go` — **Garantía: casos 7 y 7b**, FR-004a. **Por entorno**: `HOME` falso que **contiene** un `.git` (técnica de `quickstart.md` V6), con las cuatro filas de esa tabla. El caso **7b** —marcador bajo el home que **sí** cuenta— es el que distingue «techo» de «zona prohibida»; sin él, una implementación que ignore todo lo que cuelga del home pasaría igual.
- [x] T012 [P] [US1] Test de **techo de la raíz del sistema** en `internal/project/resolve_test.go` — **Garantía: caso 8**. **Test unitario con el techo INYECTADO** como parámetro, no por entorno: no hay variable que falsee la raíz del FS sin privilegios.
- [x] T013 [P] [US1] Test de **enlace hacia el interior de un proyecto** en `internal/project/resolve_test.go` — **Garantía G5**, caso **11** + escenario 6 de US1. Asevera además que el resultado **coincide con el del caso 2** sobre la ruta real: es lo que demuestra que la resolución de enlaces **precede** al reconocimiento (FR-006a / D-004-2), y no solo que converge.

### Implementación de US1

- [x] T014 [US1] Implementar el ascenso con techo en `internal/project/resolve.go` — marcador `.git` por `os.Lstat` aceptando **fichero o directorio**; techo **exclusivo** (directorio personal vía `os.UserHomeDir()` si la ruta cuelga de él, raíz del FS si no), **inyectable** para T012; primer marcador bajo el techo gana (FR-004). Pone en verde T008..T013.
- [x] T015 [US1] Implementar la resolución de enlaces **previa** al ascenso en `internal/project/resolve.go` — `filepath.EvalSymlinks` de mejor esfuerzo; si falla, se continúa con la forma léxica (FR-009). Pone en verde T013.

**Checkpoint**: US1 completa y entregable. El gasto de un proyecto aparece como **una** línea.

---

## Phase 4: User Story 2 — Trabajar fuera de un proyecto no multiplica identidades (P2)

**Goal**: los directorios sin proyecto reconocible producen una identidad estable frente a variaciones
sintácticas.

**Independent Test**: eventos en directorios que no pertenecen a ningún proyecto; variantes
sintácticas convergen y rutas distintas no colisionan.

### Tests de US2

- [x] T016 [P] [US2] Test de variaciones sintácticas en `internal/project/resolve_test.go` — **Garantía G6**, casos **9 y 10** (barra final; `.` y `..` redundantes).
- [x] T017 [P] [US2] Test de no-colisión en `internal/project/resolve_test.go` — **Garantía G7** (no fusionar lo genuinamente distinto). Es la contra-prueba de T016: sin ella, `Clean` podría sustituirse por una constante y T016 seguiría verde.
- [x] T018 [P] [US2] Test de **mejor esfuerzo** en `internal/project/resolve_test.go` — **Garantía G8**, casos **12, 13 y 14** (inexistente, permisos denegados, enlace roto). Asevera que la salida es **no vacía** y que **no hay error**. El caso de permisos se omite con `t.Skip` documentado si la suite corre como root.
  **Por qué aquí SÍ se permite el `t.Skip` que T010 prohíbe** (decidido por el orquestador, 2026-08-09): en T018 el caso de permisos es **1 de los 3** que cubren G8 —los otros dos siguen ejerciéndola— y es **físicamente inejecutable como root**, que puede leerlo todo. En T010, en cambio, el skip mataría **la única** cobertura de una promesa de la spec. La asimetría es deliberada; sin escribirla, el siguiente lector la leería como descuido y «armonizaría» una de las dos.
  **FR-010 es ESTRUCTURAL, no se prueba aquí**: «un fallo de resolución nunca detiene el procesamiento del resto del lote» se garantiza por la **firma** — `Derivar` **no devuelve error**, así que no existe rama que pueda propagar un fallo hacia arriba (contrato G8: *«Nunca devuelve error»*). Su verificación end-to-end es **V5, manual y declarada** (T036). **Si algún día `Derivar` ganara un error de retorno, esta nota es el sitio que lo prohíbe**: cambiar la firma convertiría una garantía estructural en una que habría que probar caso por caso.
- [x] T019 [P] [US2] Test de **ruta no absoluta** en `internal/project/resolve_test.go` — **Garantía: caso 15**. Asevera que **no** se ancla con `filepath.Abs`: la identidad de una ruta relativa no puede depender del directorio de trabajo del proceso agente.
- [x] T020 [P] [US2] Test de **entrada vacía y forma de salida** en `internal/project/resolve_test.go` — **Garantías G1 y G2** (vacío → vacío; no vacío → hex-64 minúscula). G2 protege FR-018: la forma no debe delatar si la identidad viene de raíz o de fallback.

### Implementación de US2

- [x] T021 [US2] Implementar el fallback normalizado en `internal/project/resolve.go` — `filepath.Clean` sobre el mejor valor disponible; **sin** `filepath.Abs`, **sin** expansión de `~` y **sin case-folding** (decisión confirmada, plan §Puntos, punto 2: lo canonicaliza la observación real y el residuo queda declarado). Pone en verde T016..T020.

**Checkpoint**: US1 + US2 completas. La derivación cumple el contrato entero.

---

## Phase 5: User Story 3 — El contrato promete exactamente lo que el agente hace (P3)

**Goal**: retirar la promesa de envío en claro del producto y de su documentación.

**Independent Test**: revisar los documentos y ejercitar la configuración con el valor retirado, sin
depender de US1 ni US2.

### Tests de US3

- [x] T022 [P] [US3] Test de **detección de la clave obsoleta** en `internal/config/config_test.go` — **Garantía: tabla de comportamiento de `contracts/cli-config.md`**. Cuatro casos: `"plain"` → error; `"hash"` → sin error; otro valor → sin error; clave ausente → sin error. **Se escribe JUNTO A T025, no antes**: un test contra una función que aún no existe **no compila**, y un rojo de compilación no es un rojo legible — no dice nada sobre el comportamiento, solo que falta un símbolo (criterio sentado en I-5, y la razón de que T006 existiera como punto de extensión). Por tanto **nace verde** y se valida **por mutación de sus cuatro casos**: cada uno debe caer al alterar la condición que lo distingue, y los otros tres quedar en pie.
- [x] T022b [P] [US3] Test unitario de **FR-015 — la clave desaparece del fichero** en `internal/config/config_test.go`: escribir un `config.json` que **contiene** la clave obsoleta (junto a campos válidos), hacer `config.Load` → `config.Save`, **releer el JSON crudo** y aseverar que la clave **ya no está**. Se relee el fichero como texto/mapa, no a través del struct: comprobarlo con el struct sería tautológico —el campo no existe, así que jamás aparecería— y no probaría nada sobre lo que quedó **en disco**. Es lo que respalda la afirmación de `contracts/cli-config.md` de que un `enroll` posterior limpia la clave por sí solo.
> **Receta común de T023/T024 — para que cada código de salida signifique lo que se cree.** Los dos
> tests usan: (a) el sandbox de T002; (b) una `config.json` **bien formada** salvo por la clave que se
> está probando —`endpoint` https válido y `device_token` sintáctico—, porque un `endpoint` inválido
> haría fallar `config.Validate()` y produciría un `ExitCode() != 0` **por otra razón**; y (c)
> `logs_root` apuntando a un **directorio VACÍO**, para que la pasada no encole nada y el `!= 0` no
> pueda venir de un fallo de transporte. Sin (b) y (c), un test verde no demostraría la parada:
> demostraría que algo falló.

- [x] T023 [P] [US3] Test de proceso de SC-007 en `cmd/permea/main_test.go` — **anclado a D-004-5**, tabla «Alcance de la parada» de `contracts/cli-config.md`. Con la receta común, comprueba por `ExitCode()`: `--run` con `"plain"` → **≠ 0** y `queue.jsonl` sin líneas nuevas; **`--daemon` con `"plain"` → ≠ 0** (segunda entrada de la tabla: si solo se probara `--run`, el daemon podría quedarse sin la parada y el test no se enteraría); `--run` con `"hash"` → **0** y **`stderr` VACÍO**; `--run` sin la clave → **0** y `stderr` vacío.
- [x] T024 [P] [US3] Test de proceso de las **excepciones de D-004-5** en `cmd/permea/main_test.go` — con la receta común y `"plain"` presente: `status` → `ExitCode()==0` **y `stderr` NO vacío** (la presencia del aviso se comprueba por no-vacuidad, simétrica a cómo T023 comprueba su ausencia por vacuidad — ninguno de los dos hace *matching* de texto); `--scan` → `ExitCode()==0` **sin** parada; **`enroll <string-válido>` → `ExitCode()==0` Y la clave `project_ref_mode` AUSENTE del `config.json` releído en crudo**. **Esta tarea es la que ancla D-004-5**: `--scan` procesa líneas con la configuración retirada presente y aun así no se detiene, porque su procesamiento no alcanza la frontera (salt de dry-run, sin cola, sin transporte). Si algún día falla, la pregunta es si `--scan` dejó de ser diagnóstico — no cómo arreglar el test.
  **Las dos aserciones de `enroll` son obligatorias, no una u otra**: «vía de reparación» promete **dos** cosas —que no se detiene **y** que repara—, y comprobar solo la primera dejaría en pie un `enroll` que arranca y no limpia nada, que es exactamente el caso que haría falso el argumento de D-004-5.
  *Nota de mecanismo para el encargo de implementación (la tarea fija la garantía; el mecanismo se resuelve al implementar)*: si `enroll` exigiera red para su ping de verificación, se usa el `httptest.NewTLSServer` del patrón ya existente en el repo; si decodifica *offline*, basta un `pmea2` fabricado con el codificador propio. Lo que **no** puede hacer el test es depender de un backend real.

### Implementación de US3

- [x] T025 [US3] Implementar la detección de la clave obsoleta en `internal/config/config.go` — segundo *unmarshal* a `map[string]json.RawMessage` mirando **solo** `project_ref_mode`; devuelve señal de parada **únicamente** para el valor `"plain"` (FR-013/013a/013b). La detección se escribe **una sola vez** aquí y la consumen `setup()` y `status`. Pone en verde T022.
- [x] T026 [US3] Retirar `ProjectRefMode` de `internal/config/config.go` (tipo `:15`, constantes `:19`/`:21`, campo `:30`, default `:41`, relleno `:120-121`) **y** actualizar `internal/config/config_test.go:36,52` **EN EL MISMO COMMIT** — el paquete **no compila** entre medias, así que separarlos deja el repositorio roto en un commit intermedio.
- [x] T027 [US3] Consumir la detección en `cmd/permea/main.go` (`setup()`, en torno a `:129`) como **error de arranque** con el mensaje de `contracts/cli-config.md` §Forma del error: nombra clave y valor, da la **ruta real** del fichero, dice qué hace el producto ahora, exit 1, y **nunca** vuelca la configuración ni secretos. **La comprobación va ANTES de cualquier otro modo de fallo de `setup()`** —antes de `LoadOrCreateSalt` (`:133`) y de `LoadOrCreateMachineID` (`:137`)—, y el porqué no es estilo: si un fallo de salt o de directorio de datos pudiera adelantarse, el usuario con `"plain"` recibiría **un error distinto del suyo** y la parada quedaría indistinguible de cualquier otra avería. La garantía de FR-013 incluye que el usuario sepa **qué** le paró. Pone en verde T023.
- [x] T028 [US3] Consumir la detección en `cmd/permea/status.go` como **aviso que no aborta** (D-004-5), emitido **solo** ante `"plain"`. Pone en verde T024.
- [x] T029 [US3] Retirar la promesa de envío en claro de la documentación **versionada**, derivando el alcance de `git ls-files` con los **cuatro** términos (`plain`, `opt-in`, `project_ref_mode`, `en claro`) — nunca de una lista escrita a mano (FR-014). Sitios en alcance conocidos, que el barrido **debe reencontrar** como control de que funciona: `specs/001-agente-inicial/contracts/boundary-event.md:35,86`, `specs/001-agente-inicial/data-model.md:29,105`, `specs/001-agente-inicial/spec.md:82` (**no contiene la palabra «plain»**) y `specs/003-enrolamiento/data-model.md:42`. `Roadmap.md` queda **fuera**: está gitignoreado (`.gitignore:17`) y no es documentación del repositorio.
- [x] T030 [US3] Verificar SC-008 ejecutando el barrido de `quickstart.md` V10 sobre `git ls-files` y dejando su salida (cero menciones) en el encargo. **Un «cero» sin haber comprobado antes que el barrido encontraba los seis sitios no significa nada** — el control previo es parte de la tarea.

**Checkpoint**: las tres historias completas.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T031 [P] Verificar **SC-006 POR DISEÑO** leyendo `internal/project/resolve.go` — comprobar que la caché existe, que su **clave es el `cwd` declarado** (no la ruta real: consultarla exigiría resolver primero, que es el trabajo que la caché evita) y que su **ámbito es la pasada**, no el proceso (un daemon de días serviría identidad obsoleta). **El benchmark es orientativo y NUNCA entra en CI**: los tests de tiempo son la familia de flakes ya fichada.
- [x] T031b [P] Verificar **G10 / FR-012 — la observación es de solo lectura** sobre `internal/project/`, con **lectura de diseño + barrido mecánico**. El barrido busca APIs de escritura del sistema de ficheros: `grep -nE "os\.(Create|Mkdir|MkdirAll|Remove|RemoveAll|WriteFile|Rename|Chtimes|Truncate)" internal/project/*.go` excluyendo `_test.go` (los tests **sí** crean árboles temporales, y eso es legítimo). **Resultado exigido: cero apariciones**, con **veredicto por aparición** si hubiera alguna — no un «parece que no escribe». Es la única forma de comprobar FR-012 que no depende de leer con buenos ojos: una escritura colada en un paquete que promete no escribir no se ve revisando, se ve buscando.
- [x] T032 [P] Implementar la caché de pasada en `internal/project/resolve.go` con la forma que T031 verifica — mapa en memoria por pasada, clave `cwd` declarado, valor la identidad ya derivada.
- [x] T033 **DEPENDE DE T029 — no paralelizable con ella** Actualizar la descripción de `project_ref` en `specs/001-agente-inicial/contracts/boundary-event.md` para que refleje la derivación nueva (raíz de proyecto con fallback normalizado) **sin cambiar** nombre, tipo ni forma del campo — el contrato de frontera se **cumple**, nunca se redefine. **Las dos tareas escriben la MISMA línea** (`boundary-event.md:35`): primero el barrido de T029 la limpia de la promesa del `plain`, después T033 la redescribe. Invertir el orden —o correrlas en paralelo— haría que T029 reencontrase texto que T033 acaba de escribir, o que T033 redescribiera una línea que T029 va a reescribir encima.
- [x] T034 Ejecutar las puertas de calidad de la constitución —las definidas en `Makefile` (`test`, `lint`)— y dejar su salida en el encargo: `go vet ./...` sin hallazgos, `golangci-lint run` limpio, `go test ./...` en verde.

---

## Phase 7: Validación (ejecuta Basilio en su máquina)

Estas tareas **no las ejecuta Claude Code**: dependen de la máquina de Basilio y de un insumo suyo.
Referencian los V-números de `quickstart.md` y **no reescriben sus pasos**.

- [ ] T035 ⛔ **BLOQUEANTE, DEPENDE DE BASILIO** — recibir la enumeración de verdad-terreno de los 12 directorios de origen y commitearla en `specs/004-identidad-de-proyecto/verdad-terreno.md`, con el formato de `research.md` §P7. **Sin este artefacto, SC-001 NO SE VALIDA NI SE MARCA**: no se puede comprobar una predicción contra un terreno que nadie ha declarado.
- [ ] T036 [P] Ejecutar las validaciones de derivación de `quickstart.md` **V1–V6** y registrar el resultado. Prerrequisito: sandbox de §Aislamiento activo (con su comprobación de que `status` dice «no enrolado»).
- [ ] T037 [P] Ejecutar las validaciones de frontera y regresión de `quickstart.md` **V7 y V8** — V8 compara contra `baseline-sc004.tsv` de T001 en sus **dos** dimensiones: el **conjunto** de identidades de sesión y máquina (**SC-004** / FR-019 — si cambiaron, se ha tocado algo que FR-019 prohíbe) y el **recuento de eventos** capturado en el paso 7 de T001 (**SC-003** — si bajó, se han perdido eventos). **La pasada de V7 sobre `TestEvent_OnlyAllowlistKeys` ES la verificación de FR-020** («el conjunto cerrado de campos no cambia; no se añade ningún campo»): se ejecuta **con esa intención declarada**, no como cobertura accidental que aprovecha un test preexistente.
- [ ] T038 [P] Ejecutar la validación de configuración de `quickstart.md` **V9**, incluida la fila de `--scan` que ancla D-004-5.
- [ ] T039 Ejecutar `quickstart.md` **V11** (coste, orientativo) y registrar el número **sin convertirlo en puerta**.
- [ ] T040 Validar **SC-001** comparando las identidades reprocesadas contra la predicción derivada de la regla sobre `verdad-terreno.md`. **Depende de T035.** La predicción se escribe **antes** de ejecutar: si se deriva después, deja de ser una prueba y pasa a ser una racionalización.

---

## Dependencies & Execution Order

### Orden de fases

```
Phase 1 (Setup)          T001 ─► T002, T003
   │  T001 COMMITEADA antes de cualquier cambio de comportamiento
   ▼
Phase 2 (Foundational)   T004, T005 ─► T006 ─► T007        ⚠️ BLOQUEA TODO
   │  golden reforzado ANTES de tocar la derivación
   ▼
Phase 3 (US1 · P1)       T008..T013 (rojo) ─► T014 ─► T015   🎯 MVP
   ▼
Phase 4 (US2 · P2)       T016..T020 (rojo) ─► T021
   ▼
Phase 5 (US3 · P3)       T022, T022b, T023, T024 (rojo) ─► T025 ─► T026 ─► T027, T028
                                                            ─► T029 ─► T030
   ▼
Phase 6 (Polish)         T032 ─► T031 ; T031b, T034 ; T029 ─► T033
   ▼
Phase 7 (Validación)     T035 ⛔ ─► T040 ; T036..T039
```

### Dependencias que no son de fase

- **T026 y el arreglo de `config_test.go` son UNA tarea, no dos**: el paquete no compila entre medias.
- **T031 verifica lo que T032 implementa**, así que T032 va primero aunque T031 aparezca antes en el
  texto (la verificación es de lectura de diseño, no de ejecución).
- **T033 depende de T029**, y no es una dependencia de fase sino **de fichero y línea**: las dos
  escriben `specs/001-agente-inicial/contracts/boundary-event.md:35`. T029 (Phase 5) limpia la
  promesa del `plain`; T033 (Phase 6) redescribe el campo. Correrlas en paralelo o al revés hace que
  una pise el trabajo de la otra — por eso T033 **no lleva `[P]`**.
- **T040 depende de T035**, que depende de Basilio. Es la única dependencia externa del plan.
- **US3 es independiente de US1 y US2**: podría entregarse antes. Va la última por prioridad de
  valor, no por dependencia técnica.

### Oportunidades de paralelismo

| Fase | Paralelizable | Motivo |
|---|---|---|
| Phase 1 | T002, T003 | ficheros distintos, tras T001 |
| Phase 3 | T008..T013 | todos en `resolve_test.go` pero **independientes entre sí**: se escriben como funciones de test separadas |
| Phase 4 | T016..T020 | ídem |
| Phase 5 | T022, T022b, T023, T024 | dos ficheros (`config_test.go` para T022/T022b, `main_test.go` para T023/T024) |
| Phase 6 | T031, T031b, T034 | verificaciones y puertas no se estorban. **T033 queda fuera**: comparte fichero y línea con T029 |
| Phase 7 | T036..T039 | validaciones manuales independientes |

> **Aviso sobre el `[P]` dentro de un mismo fichero**: T008..T013 y T016..T020 comparten
> `resolve_test.go`. Son paralelizables **como diseño** (ninguna depende del resultado de otra), pero
> si se editan concurrentemente hay conflicto de fichero. Se marcan `[P]` porque el orden entre ellas
> es libre, no porque puedan escribirse a la vez sin coordinación.

---

## Implementation Strategy

### MVP: solo US1 (Phases 1 → 3)

US1 sola ya entrega el valor medido: **los 12 buckets fragmentados colapsan a los proyectos
genuinos**. US2 refina el resto y US3 es higiene de contrato.

### Entrega incremental

1. **Phases 1–2** → nada cambia de comportamiento, y eso es verificable (T007). La frontera queda
   reforzada antes de tocar nada.
2. **Phase 3** → MVP entregable: el gasto por proyecto agrupa.
3. **Phase 4** → los directorios sueltos dejan de multiplicarse.
4. **Phase 5** → el contrato deja de prometer lo que el agente no hace.
5. **Phases 6–7** → pulido y validación en la máquina de Basilio.

### Lo que este plan de tareas NO hace

- No toca la plataforma (repo aparte).
- No añade campos al evento (FR-020) ni toca `Ref`, salt, sesión o máquina (FR-019).
- No sanea `stderr` (D-004-4: trabajo aparte, ya identificado).
- No cierra la entrada de deuda de `Roadmap.md` (gitignoreado; lo cierra Basilio en el cierre de
  sesión).
