# Tasks: Adhesión a proyecto

**Feature**: `005-adhesion-a-proyecto` · **Fecha**: 2026-08-18 · **Spec**: [spec.md](./spec.md)
**Plan**: [plan.md](./plan.md) · **Contratos**: [adhesion.md](./contracts/adhesion.md) ·
[cli.md](./contracts/cli.md) · **Validación**: [quickstart.md](./quickstart.md)

**Base**: `6c32ded`. **Línea base a preservar**: `go test ./...` → **9 paquetes `ok`, 0
`[no test files]`, 0 `FAIL`**.

---

## Format: `[ID] [P?] Description`

- **[P]**: **sin dependencia entre sí — escribibles en cualquier orden**, y cada una nace en su rojo
  con independencia de las otras del grupo.
  > ⚠️ **NO significa «ficheros distintos», y aquí no lo son.** Cuatro grupos marcados `[P]`
  > —T008–T010, T013/T014, T016–T020, T022–T026— **comparten fichero**: es lo normal cuando lo que se
  > paraleliza son casos de un mismo sujeto. Lo que `[P]` promete es que **el orden entre ellas no
  > cambia el resultado**, no que puedan escribirse a la vez.
  >
  > **Un mismo fichero NO admite dos escritores simultáneos**: si dos tareas `[P]` tocan el mismo
  > `_test.go`, se hacen **una detrás de otra**. El paralelismo de este fichero es **de decisión**
  > —qué escribir sin esperar a nada—, no de edición concurrente.
- Ruta de fichero exacta en cada tarea
- **No hay `[Story]`**: ver §Por qué el troceo NO va por historias

### Vocabulario — dos numeraciones distintas, y no son la misma

Este fichero usa las dos y **conviene no confundirlas**:

| Notación | Qué es | Dónde vive | Cuántos |
|---|---|---|:--:|
| **D1 … D4** | **los desenlaces REMOTOS de la adhesión** — lo que responde la plataforma | `contracts/adhesion.md` §Los cuatro desenlaces | **4** |
| **«los ocho desenlaces»** | **los del COMANDO** — incluye los tres rehúses locales y el no verificable, que nunca llegan a la plataforma | `contracts/cli.md` §Comportamiento y salidas | **8** |

**Los cuatro remotos son un subconjunto de los ocho del comando.** Cuando una tarea dice «los ocho»
se refiere **siempre** a los del comando; cuando dice `D3`/`D4`, a los remotos.

### Convenio de «Phase» — dos numeraciones que colisionan

**Dentro de este fichero, «Phase N» significa SIEMPRE la fase de ejecución de este fichero** (Phase 1 …
Phase 10). El plan usa **otra** numeración —Phase 0 research · Phase 1 diseño y contratos · Phase 2
tasks—, así que **toda cita al plan se escribe «Phase N del plan»**, cualificada. Sin el convenio,
«Phase 1» son dos cosas distintas y ninguna de las dos avisa.

## Path Conventions

Proyecto único Go. Rutas desde la raíz del repositorio: `cmd/permea/`, `internal/`,
`specs/005-adhesion-a-proyecto/`.

---

## Disciplinas transversales — aplican a TODA tarea de test de este fichero

No se repiten en cada tarea; se dan por incluidas en todas. **Son las de 004 (`004/tasks.md:33-56`),
sin cambios**, más una que esta feature añade.

1. **Una garantía por tarea**, anclada al **contrato** (`contracts/adhesion.md`,
   `contracts/cli.md`), **nunca** a la redacción de la spec. Si contrato y spec discrepan, se para y
   se reporta — no se elige.
2. **Rojo antes de verde.** El encargo que observa el fallo **transcribe la razón del fallo** (el
   mensaje real, no «falla como se esperaba»).
3. **Todo test que nazca verde se valida por mutación**, y la validación exige **leer el mensaje de
   fallo** que produce la mutación. La mutación se revierte **por edición inversa**, nunca por
   `git checkout` — para que el diff final demuestre que se revirtió lo mismo que se introdujo.
   > **Y una mutación VÁLIDA deja el paquete COMPILANDO y mata SÓLO el hecho que altera.** Esta mitad
   > faltaba, y **se aprendió midiendo**: de las cuatro alteraciones que se probaron en T004, **dos no
   > acreditaban nada** (registro en T004 §HALLAZGO DE MÉTODO). Hay **dos formas de mutación inválida**,
   > y las dos se leen como éxito si no se mira:
   >
   > - **La que PANICA.** Un panic **aborta el binario de test entero**, y con él **la observación del
   >   otro hecho** — que era el punto de la mutación. Se ve un `FAIL` rojo y abundante que **no dice
   >   qué murió**.
   > - **La que NO COMPILA.** Un `[build failed]` **falla igual aunque el test no mire nada**, así que
   >   **no dice nada del test**: es el mismo desenlace con un test correcto y con uno vacío.
   >
   > **Si una alteración produce cualquiera de las dos, se DESCARTA y se busca otra forma de alterar el
   > mismo hecho** — no se apunta como mutación superada. Y si no existe ninguna forma que compile y
   > mate sólo ese hecho, **se PARA y se reporta**: eso ya no es un problema de la mutación, es que los
   > hechos no están separados.
   >
   > **Rige las CUATRO mutaciones que quedan**: **T007**, la **mutación de unificación de T005**, y las
   > cláusulas **(a)** y **(c)** de **T028**.
   >
   > ⚠️ **La de T005 es la más expuesta de las cuatro, y por su propia naturaleza**: muta la función que
   > usan **cuatro llamantes en tres paquetes**, así que una mutación mal escrita **rompe la compilación
   > de los cuatro a la vez** — y **un `build failed` en los cuatro se lee exactamente igual que “los
   > cuatro se movieron”**, que es lo que esa mutación existe para demostrar. **La forma inválida
   > FALSIFICA el resultado que se busca, no sólo lo oculta.** La mutación de T005 tiene que **compilar**
   > y hacer fallar **tests**, no el `go build`.
4. **Tests de proceso**: comparar `ExitCode()`, **nunca** texto (puente Windows/WSL).
5. **La ausencia de aviso se comprueba por canal VACÍO**, no por *matching* de mensaje: comprobar que
   un texto no aparece pasa también cuando aparece otro texto distinto.
6. **Aislamiento obligatorio**: todo test que toque config o cola usa `HOME`/`USERPROFILE`/
   `XDG_CONFIG_HOME` apuntando a un temporal (`quickstart.md` §Aislamiento). Ningún test escribe jamás
   en la instalación real.
7. **⚠️ NUEVA EN 005 · Los dos canales se capturan POR SEPARADO.** Nunca combinados: una salida
   conjunta no vacía es compatible con cualquier reparto, **incluido el equivocado**, así que un
   montaje que los una **no cuenta como pasado** (SC-011 B). Es el caso particular de la disciplina 5
   para esta feature, y se escribe aparte porque aquí **el canal es requisito**, no detalle.

**Nota de ejecución**: Claude Code **no ejecuta git de escritura**. El marcado de casillas de este
fichero viaja en **el mismo commit** que el código que documenta, y ese commit lo hace Basilio.

> **Dónde se registran los rojos y las mutaciones.** **Aquí, en este fichero** — es la convención
> medida de este repositorio: 001–004 **no tienen** `registro-rojos.md` ni `registro-mutaciones.md`
> (esos son de la plataforma). El rojo se declara en la tarea que lo espera («Nace en **rojo**: …») y
> su razón se transcribe al observarlo; la mutación se declara en la tarea que la exige. **No se crea
> ningún artefacto nuevo de registro.**

---

## Por qué el troceo NO va por historias

004 troceó por historias (US1/US2/US3) y fue correcto: allí cada historia tocaba **una regla distinta**
de la derivación. **Aquí no.** Las cinco historias de 005 se realizan sobre **un solo comando** y **un
solo método de transporte**: US1 es el éxito, US2 la repetición del mismo éxito, US3 un rehúse previo,
US4 dos desenlaces de error y US5 tres estados. Trocear por historia produciría **cinco fases que
reabren los mismos dos ficheros**, y —peor— violaría la regla de que *un bloque no implementa un
desenlace cuyo disparador llega en otro bloque*: US2 no puede probarse sin el éxito de US1, y US4/US5
no pueden probarse sin el comando de US3.

**Así que las fases van por MECANISMO**, siguiendo las decisiones del plan (D-005-P1 … D-005-P14). La
cobertura de historias no se pierde: está en §Tabla de cobertura, fila a fila.

---

## Phase 1: Puntos de extensión — para que los rojos sean LEGIBLES

**El criterio es de 004, y está sentado**: un test contra un símbolo que no existe **no compila**, y un
rojo de compilación **no es un rojo legible** — no dice nada del comportamiento, solo que falta un
nombre. 004 lo resolvió con su T006; aquí hacen falta tres puntos de extensión.

- [x] **T001** [P] En `internal/project/resolve.go`, añadir una función exportada que devuelva
  **(identidad, huboRaíz)**, **compartiendo cuerpo interno** con `Derivar`. `Derivar` **NO cambia de
  firma, ni de comportamiento, ni de nombre** (D-005-P5, FR-015, FR-016). Es el punto de extensión que
  permite a los tests de rehúse local nacer en rojo **por la razón correcta**.
  ⚠️ **Cuerpo compartido, no duplicado**: es lo que FR-005 exige —*no dos derivaciones que hoy den lo
  mismo*—, y lo que hace que SC-001 sea demostrable en **T028**.
- [x] **T002** [P] En `internal/transport/transport.go`, añadir el método de adhesión con **su firma
  final**, que **COMPONE Y EMITE la petición con el cuerpo definitivo `{code, project_ref}`** y, sea
  cual sea la respuesta, devuelve **un centinela propio: «adhesión no implementada»**.
  **`Send` no se toca** (D-005-P1).
  ⚠️ **El centinela DEBE ser distinto de «no verificable»**, y es el punto de la tarea. Dos razones,
  y las dos son de que los rojos digan la verdad:
  - **T007 necesita un cuerpo que observar.** Un punto de extensión que no compone la petición deja al
    golden de frontera sin sujeto: no habría nada sobre lo que comprobar la allowlist de dos campos.
  - **T010 espera «no verificable».** Si el andamiaje devolviera ese mismo valor, **T010 nacería verde
    por accidente** —acertaría contra el stub, no contra el comportamiento— y su rojo no existiría.
- [x] **T003** En `cmd/permea/main.go` + `cmd/permea/project.go` (nuevo), el **despacho de dos
  niveles**: un tercer `if` para `project` en la escalera existente, **antes de `flag.Parse()`**, que
  delega en un despachador propio. De momento `join` **rehúsa siempre** con «no implementado».
  **Los flags de P-001/P-002 no se tocan** (D-005-P6, `main.go:42-43`).
  ⚠️ **El rehúse de andamiaje sale con un código que NO es ninguno de los ocho del contrato** —los
  del comando son **0** y **1**, así que el andamiaje usa **70**—. Es lo que hace que **T017 y T020
  nazcan rojas por COMPORTAMIENTO** y no por casualidad: con `1` acertarían contra el andamiaje
  —que también rehúsa y también sale ≠ 0— sin que nadie hubiera implementado nada.
  ⚠️ **Quién RETIRA el 70, y es mecánico: T023.** Ese test compara **los ocho códigos exactos** del
  contrato, así que **cualquier superviviente del andamiaje lo tumba** — no hace falta acordarse de
  quitarlo, ni revisarlo a ojo. La retirada deja de ser **intención** y pasa a ser **garantía**.
  *(T017 y T020 cazan además los dos casos de error de uso, por la misma vía.)*

- [x] **T029** **SC-009 · escritura y PRIMERA ejecución** — en
  **`internal/ingest/baseline_regresion_test.go`** (nuevo, **paquete EXISTENTE**): comparar las **tres
  columnas** de `specs/004-identidad-de-proyecto/baseline-sc004.tsv` contra las identidades que produce
  una pasada, **con las semillas del bloque REPRODUCCIÓN** vía `testutil.SandboxConSemillas`.
  **Su única dependencia es T001**, así que se escribe y se ejecuta **aquí**, no en Phase 8.
  ⚠️ **Por qué en `internal/ingest` y no en `internal/project`** —medido—: las tres columnas las
  produce `internal/ingest/claudecode.go:86-88` (`ProjectRef`, `SessionRef`, `MachineRef`) en
  `FromClaudeCodeLine`. Desde `internal/project` **solo se podría comparar UNA de las tres**, y el
  baseline existe precisamente para comparar las tres. `internal/ingest` **ya importa `testutil` en sus
  tests**, así que no añade ninguna arista.
  ⚠️ **Y NO en un paquete nuevo, que es lo crítico**: el checkpoint de abajo compara contra **9 ok**.
  Un paquete nuevo lo convertiría en **10** y **rompería la puerta que dice «nada cambió» justo al
  escribir la prueba de que nada cambió**.
  > ### ⚠️ NACE VERDE — y aquí la mutación no es ceremonia, es lo único que lo sostiene
  > **Es un test de regresión cero**: si T001 está bien hecha, **no ha cambiado nada** y el test da
  > verde **desde el primer instante**. La disciplina 3 exige validarlo por mutación, y en éste **es
  > crítico**: T029 es **la ÚNICA puerta que dice que el camino de ingesta no se movió**, y
  > **un T029 que lea mal el `.tsv`, que reciba cero semillas o que compare dos conjuntos vacíos DA
  > EXACTAMENTE EL MISMO VERDE que uno correcto.**
  >
  > **Validación**: alterar deliberadamente **el cuerpo compartido de T001**, comprobar que T029
  > **CAE**, **transcribir el mensaje de fallo real** y revertir **por edición inversa** —nunca por
  > `git checkout`—. **Si la mutación no lo tumba, el test no está mirando: se PARA y se reporta.**
  > ### ⚠️ LÍMITE CONOCIDO — qué acredita T029 y qué NO
  > Compara **el CONJUNTO deduplicado de identidades** y **el recuento de eventos**. Eso demuestra que
  > **aparecen las mismas identidades y no se ha perdido ningún evento** — pero **NO que una entrada
  > dada siga produciendo la misma identidad**: el dedup colapsa la correspondencia entrada→salida, y
  > el conjunto podría coincidir con las filas permutadas entre entradas distintas.
  >
  > **Y la base es fina**: `# meta events_total 2`. **Dos eventos y una sola fila de identidades.** Un
  > cambio que afectara solo a entradas que el fixture no ejerce **pasaría inadvertido**.
  >
  > **Es limitación del artefacto de 004, no de esta tarea, y no se arregla aquí**: ampliar el fixture
  > cambiaría el conjunto y **rompería la comparación por la razón equivocada** —lo dice la cabecera del
  > propio `.tsv`, que ya lo sufrió con `boundary_sample.jsonl`—. Se escribe como límite para que nadie
  > lea el verde de T029 como más de lo que dice. **Quien necesite la correspondencia entrada→salida
  > tiene SC-001 (T028)**, que sí compara por entrada sobre cuatro clases de árbol.

**Checkpoint**: la suite sigue en **9 ok**. **Ningún comportamiento observable ha cambiado** — y eso
**se comprueba contra artefacto, no de memoria**: comparar las identidades derivadas contra
`specs/004-identidad-de-proyecto/baseline-sc004.tsv`, **reutilizando las semillas del bloque
REPRODUCCIÓN**. T001 toca `internal/project/resolve.go`, y **éste es el momento en que la regresión
es barata de encontrar**: antes de que nada más cambie.

> **Medido antes de escribirlo, y la mitad ya existe.** `internal/testutil` **ya lee el artefacto y
> extrae las dos semillas**: `SemillasDeLaLineaBase` (`sandbox.go:116-120`) y `SandboxConSemillas`
> (`:104`), con sus propios tests (`sandbox_test.go:75` y `:102`). **Así que el checkpoint NO
> construye la lectura: reutiliza `testutil.SandboxConSemillas`.**
>
> **Lo que NO existe** en ninguno de los 9 paquetes es **la comparación de las tres columnas de
> identidades** contra el `.tsv`. En 004 esa comparación era **manual** —V8, ejecutada en
> `004/tasks.md` **T037**, que es de OTRA feature y no de este fichero—, no de la suite.
>
> **Así que hay que escribirla, y tiene tarea con casilla propia AQUÍ: es T029**, arriba en esta misma
> fase. En Phase 8 vive **T029-R**, que la **re-ejecuta** como puerta formal de SC-009.
>
> **Se hace dos veces y a propósito**: **aquí para DETECTAR** —un fallo de T001 se ve en el acto, antes
> de que nada más cambie y con un solo sospechoso— y **en Phase 8 para ACREDITAR**, con todo el código
> ya escrito. No es duplicación: **es la diferencia entre encontrar el fallo barato y demostrar la
> propiedad**. Un checkpoint que exige una comprobación **sin tarea ni casilla** es una intención, no
> una puerta.

---

## Phase 2: La guarda de esquema — unificación AISLADA (D-005-P2)

> ⚠️ **Va antes que el método que la usará, y sola.** Es la tarea de mayor alcance del plan y **la
> única que puede romper algo que hoy funciona**: toca tres ficheros verdes que esta feature no
> necesitaba tocar.

- [x] **T004** En `internal/config/`, extraer el juicio de esquema a una función que devuelva **los dos
  hechos por separado** —«no analizable» y «esquema no admisible»— y **que NO formatee ningún
  mensaje**. **Nace verde** (no hay llamantes aún) → **se valida por mutación**: alterar cada uno de
  los dos hechos debe producir un fallo legible, y el otro quedar en pie.
  ⚠️ **Devuelve dos hechos y no un booleano**, y es el punto entero de la tarea: las tres réplicas
  **no son idénticas** (`research.md` §R4) — `enrollment.go` **funde** análisis y esquema en un
  desenlace, las otras dos los **separan** —, así que un booleano cambiaría el comportamiento de una
  de las tres.

  > ### ✅ MEDIDO — T004 nació verde y las dos mutaciones son independientes
  > `internal/config/endpoint.go` · `JuzgarEndpoint(endpoint string) (errAnalisis error, admisible bool)`.
  > **Devuelve el error, no un segundo booleano**, y no fue una preferencia: `config.go:102` y
  > `transport.go:143` **envuelven la causa con `%w`**, así que dos booleanos habrían roto sus mensajes.
  >
  > El test se partió en **tres funciones** —`_HechoAnalisis`, `_HechoEsquema`,
  > `_NoAnalizableNoAfirmaEsquema`— precisamente para que **cada mutación pueda tumbar una y dejar la
  > otra en pie**; con una sola función de tabla, las dos mutaciones habrían sido indistinguibles.
  >
  > **Mutación (a)** — se altera «no analizable» (`return err, false` → `return nil, false`):
  > ```
  > --- FAIL: TestJuzgarEndpoint_HechoAnalisis/no_analizable_devuelve_la_causa
  >     endpoint_test.go:25: JuzgarEndpoint("https://ejemplo\x7f.test/ingest"): errAnalisis = nil, se esperaba la causa de url.Parse
  > --- FAIL: TestJuzgarEndpoint_NoAnalizableNoAfirmaEsquema
  >     endpoint_test.go:80: el caso base falló: "https://ejemplo\x7f.test/ingest" debería no ser analizable
  > ```
  > `TestJuzgarEndpoint_HechoEsquema` → **PASS** (el otro hecho, EN PIE).
  >
  > **Mutación (b)** — se altera «esquema no admisible» (`u.Scheme == esquemaAdmisible` → `!=`):
  > ```
  > --- FAIL: TestJuzgarEndpoint_HechoEsquema/https_es_admisible
  >     endpoint_test.go:68: JuzgarEndpoint("https://api.permea.example/api/v1/ingest"): admisible = false, want true
  > --- FAIL: TestJuzgarEndpoint_HechoEsquema/http_NO_es_admisible
  >     endpoint_test.go:68: JuzgarEndpoint("http://api.permea.example/api/v1/ingest"): admisible = true, want false
  > ```
  > `TestJuzgarEndpoint_HechoAnalisis` → **PASS** (el otro hecho, EN PIE).
  >
  > Las dos revertidas **por edición inversa**, fichero verificado **idéntico byte a byte** al original.
  >
  > ### ⚠️ HALLAZGO DE MÉTODO — dos alteraciones DESCARTADAS por no ser fallos legibles
  > La disciplina 3 pide que la mutación **falle legiblemente**, y **no toda alteración lo hace**:
  > - `if err != nil` → `if false` dejó `u` nil y produjo **`panic: nil pointer dereference`**. Un panic
  >   **aborta el binario de test entero**, así que el otro hecho **no se pudo observar en pie** — que es
  >   justo lo que la mutación tenía que demostrar.
  > - `return nil, u.Scheme == …` → `return nil, true` **no compiló** (`u declared and not used`).
  >   Un fallo de compilación **no dice nada del test**: falla igual aunque el test no mire nada.
  >
  > **Una mutación que revienta o que no compila no acredita separación.** Descartadas y sustituidas
  > por alteraciones que **compilan y fallan por comportamiento**.
- [x] **T011** [P] Test de que **la guarda de esquema muerde en el método nuevo** en
  `internal/transport/adhesion_test.go` — **Garantía**: FR-017, y la separación que T004 existe para
  sostener. **DOS CASOS**, y hacen falta los dos:
  1. **El centinela.** Destino en claro (analiza bien, esquema `http`) → no se completa, **con el
     centinela `ErrScheme`** — el mismo que la ingesta, para que `errors.Is` conteste igual por las dos
     puertas. Nace **rojo** porque **T002 SÍ tiene guarda —una réplica inline— pero devuelve
     `errEsquemaAndamiaje`**: rojo **por el centinela, no por ausencia de guarda**.
  2. **La distinguibilidad.** Los **dos hechos** producen desenlaces **distinguibles**: en claro →
     `errors.Is(err, ErrScheme)`; **no analizable** → no se completa **y NO es `ErrScheme`**. Nace
     **rojo** porque hoy **los dos devuelven el mismo `errEsquemaAndamiaje`**, así que **son
     indistinguibles**.
  3. **La causa conservada** — **Garantía: D-005-P2**. Con un endpoint **no analizable**, el error
     **conserva la causa de `url.Parse`**, **con la misma forma que `Send`** (`transport.go:143`, que
     la envuelve con `%w`). Nace **rojo** porque `Adherir` hoy hace **lo contrario**: envuelve el
     centinela y **descarta la causa**.
  > ### ⚠️ El caso 2 es lo que impide REFUNDIR lo que T004 separó
  > **Con sólo el caso 1, T005 puede ponerlo verde devolviendo `ErrScheme` TAMBIÉN cuando la URL no se
  > puede analizar** — verde, y habiendo fundido los dos hechos en uno. El caso 1 no lo ve: sólo mira
  > el canal en claro, y ahí la respuesta sería la correcta por accidente.
  >
  > **La aserción del caso 2 es la DISTINGUIBILIDAD, no el texto de ningún mensaje**: compara **las dos
  > respuestas entre sí** —`errors.Is(·, ErrScheme)` no puede contestar lo mismo a los dos hechos— más
  > el sentido (`ErrScheme` es un juicio **sobre** el esquema: sólo puede afirmarlo quien pudo leerlo).
  > Comparar texto lo ataría a una redacción, y **T004 no formatea ninguna**: la redacción es de cada
  > llamante.
  >
  > **Lo que el caso 2 deliberadamente NO fija**: que el error de «no analizable» conserve la causa de
  > `url.Parse`. Hoy `Adherir` es una **CUARTA VARIANTE** de la guarda —**ofrece centinela y tira la
  > causa**, al revés que `Send` (`transport.go:143`), que **conserva la causa y no ofrece centinela**.
  > El caso 2 sólo le prohíbe a T005 **la opción que borra la diferencia entre los dos hechos**; **la
  > causa la fija el CASO 3**, y ya no queda como decisión libre.
  > ### ⚠️ El caso 3 es la PREMISA de D-005-P2, no una preferencia de estilo
  > **Unificar el juicio y dejar que una puerta conserve la causa y la otra la tire es unificar el
  > código y mantener la divergencia justo donde se nota: en lo que la persona lee cuando su
  > configuración está rota.** El juicio compartido no le sirve de nada a quien tiene un endpoint mal
  > escrito si una puerta le dice **qué** está mal y la otra sólo que algo lo está. **D-005-P2 unifica
  > EL DESENLACE, no sólo la condición** — si no, el trabajo se queda a medias exactamente en la mitad
  > que el usuario ve.
  >
  > **La aserción es `errors.As` sobre `*url.Error`, nunca texto** — y **no puede ser `errors.Is`**:
  > **medido**, `url.Parse` devuelve un `*url.Error` **nuevo en cada llamada** y `url.Error` **no
  > implementa `Is`**, así que `errors.Is(err, <causa de una segunda llamada>)` da **false aunque la
  > causa esté perfectamente conservada** — compara punteros. Se comprueba con `errors.As` (tipo que
  > **sólo** produce `url.Parse`) más sus **campos estructurados** `Op` y `URL`.
  >
  > **Y `Send` es EL ORÁCULO del caso 3**: se comprueba **la misma propiedad en las dos puertas**, así
  > que «misma forma que `Send`» **es la aserción y no una afirmación del comentario**. Si la rama de
  > `Send` fallara, lo que cambió es **la referencia**, y entonces se revisa el test antes que
  > `Adherir` — el test lo dice en su propio mensaje.
  > ### ⚠️ LÍMITE CONOCIDO — qué acredita T011 y qué NO
  > Este enunciado decía **«es la prueba de que la unificación LLEGÓ aquí»**, y **no puede serlo.**
  >
  > **T011 observa comportamiento: que `Adherir` devuelva `ErrScheme`.** Y eso **se satisface cambiando
  > una palabra en la réplica inline** —`errEsquemaAndamiaje` → `ErrScheme`— **sin unificar nada**: las
  > cuatro condiciones seguirían copiadas a mano y el test estaría verde.
  >
  > **Dónde vive el juicio es ESTRUCTURA, y ningún test la ve.** Un test no puede distinguir «llama a
  > la función unificada» de «tiene una copia que devuelve lo mismo».
  >
  > **Lo que T011 SÍ acredita, y no es poco**: que la segunda puerta de la frontera **tiene guarda de
  > esquema y devuelve el centinela correcto** —el mismo que la ingesta—, así que quien compruebe
  > `errors.Is(err, ErrScheme)` obtiene la misma respuesta por las dos puertas. **Es la mitad de
  > comportamiento.**
  >
  > **La garantía de unificación es la OTRA mitad, y está en T005**: la mutación de la función
  > unificada que exige que **los cuatro llamantes cambien a la vez**. Esa sí ve la estructura, porque
  > **una copia no se mueve cuando muta el original**.
  > ### ⚠️ VIVE EN PHASE 2, ENTRE T004 Y T005 — y el sitio es el mecanismo
  > **En Phase 4 este test NACERÍA VERDE.** T005 —Phase 2— sustituye la réplica de andamiaje de
  > `Adherir` por la función unificada, así que **para cuando llegara el turno de Phase 4, `Adherir`
  > llevaría dos fases devolviendo `ErrScheme`** y el test pasaría a la primera. **El rojo dependía de
  > un estado que T005 destruye antes de que a T011 le llegue el turno.**
  >
  > Aquí, en cambio, **nace ROJO contra `errEsquemaAndamiaje`** —el centinela de andamiaje que T002
  > dejó puesto— y **T005 lo pone VERDE**.
  >
  > **Y con eso T005 gana un criterio POSITIVO que no tenía.** Su único gate era «los cuatro tests
  > existentes siguen verdes», y **«nada se rompió» es compatible con «no pasó nada»**: la tarea de
  > mayor alcance del plan no tenía ninguna prueba de que **hizo su trabajo**. T011 aporta **la mitad
  > de comportamiento**; la de estructura la aporta **la mutación de T005**.
  >
  > ### 🔴 MEDIDO — T011 nació ROJO, y por EL CENTINELA
  > `internal/transport/adhesion_test.go` · `TestAdherir_RechazaCanalEnClaro`, dos casos (`http` remoto
  > y `http` a la máquina local). Mensaje de fallo **real**:
  > ```
  > --- FAIL: TestAdherir_RechazaCanalEnClaro/http_en_claro
  >     adhesion_test.go:65: Adherir("http://api.permea.example/api/v1/projects/adhesion"): errors.Is(err, ErrScheme) = false, err = transport: andamiaje P-005 — el endpoint debe usar https:// (guarda inline, se retira en T005): "http://api.permea.example/api/v1/projects/adhesion"
  >           la guarda de la segunda puerta no devuelve el centinela de la ingesta
  > ```
  > **El rojo es exactamente el previsto**: las dos primeras aserciones —no se completa, no hay
  > denominación— **ya pasan**, porque la guarda inline de T002 sí muerde. Lo único que falla es
  > `errors.Is(err, ErrScheme)`. **Queda en rojo a propósito; la pone en verde T005.**
  >
  > ### 🔴 MEDIDO — el CASO 2 nació ROJO, y por INDISTINGUIBILIDAD
  > `TestAdherir_LosDosHechosSonDistinguibles`. Mensaje de fallo **real**:
  > ```
  > adhesion_test.go:133: los dos hechos son INDISTINGUIBLES: errors.Is(·, ErrScheme) contesta false a los dos
  >       en claro       ("http://api.permea.example/api/v1/projects/adhesion") → false
  >       no analizable  ("https://ejemplo\x7f.test/api/v1/projects/adhesion") → false
  >       la segunda puerta funde «no analizable» y «esquema no admisible» en un solo desenlace
  > adhesion_test.go:143: errors.Is(err, ErrScheme) = false para "http://…": el esquema se leyó y no es admisible
  > ```
  > **La premisa del test SÍ se cumple hoy**: los dos endpoints se rechazan y ninguno devuelve
  > denominación — la guarda inline muerde en ambos. Lo que falla es que **muerde igual**. Y la aserción
  > de sentido en la rama «no analizable» (**no debe ser `ErrScheme`**) **pasa hoy** por accidente: hoy
  > no es `ErrScheme` porque **nada** lo es. **Es exactamente la que T005 puede romper.**
  >
  > ### 🔴 MEDIDO — el CASO 3 nació ROJO, y con las dos puertas lado a lado
  > `TestAdherir_ConservaLaCausaDelParseo`. Mensaje de fallo **real**:
  > ```
  > adhesion_test.go:220: Adherir("https://ejemplo\x7f.test/api/v1/projects/adhesion"): errors.As(err, **url.Error) = false — LA CAUSA SE PERDIÓ.
  >       err = transport: andamiaje P-005 — el endpoint debe usar https:// (guarda inline, se retira en T005): endpoint inválido "https://ejemplo\x7f.test/…"
  >       Send, con el mismo endpoint, SÍ la conserva: transport: endpoint inválido "https://ejemplo\x7f.test/…": parse "https://ejemplo\x7f.test/…": net/url: invalid control character in URL
  >       D-005-P2: las dos puertas deben dar el mismo desenlace, no sólo compartir la condición
  > ```
  > **La rama del oráculo PASÓ**: `Send` conserva la causa hoy, así que la referencia se sostiene y el
  > rojo es **inequívocamente de `Adherir`**. El mensaje **enseña las dos puertas juntas**, que es lo
  > que hace legible la divergencia sin comparar texto en ninguna aserción.

- [ ] **T005** Sustituir **las CUATRO réplicas** por llamadas a T004:
  1. `internal/config/enrollment.go:78-80` — **funde los dos hechos** en su error genérico, **sin
     reproducir el argumento** (lleva el token dentro).
  2. `internal/config/config.go:100-106` — **dos ramas**, con sus mensajes.
  3. `internal/transport/transport.go`, en **`Send`** — **dos ramas**, y **conserva el centinela
     `ErrScheme`**, del que dependen tests por `errors.Is`.
  4. `internal/transport/transport.go`, en **`Adherir`** — **la réplica de andamiaje que introdujo
     T002**, que hoy devuelve `errEsquemaAndamiaje`. **Al sustituirla pasa a devolver `ErrScheme`**, y
     eso es **lo que pone T011 en verde**. Con ella desaparece también la declaración de
     `errEsquemaAndamiaje`.
  > ⚠️ **Son CUATRO, no tres.** La cuarta nació en T002 —Phase 1— y **un ejecutor que siga la lista
  > vieja al pie de la letra la deja viva**: la segunda puerta de la frontera seguiría con una
  > condición copiada a mano, que es exactamente lo que D-005-P2 existe para eliminar.
  > ### ⛔ CONDICIÓN DE PARADA
  > **Los cuatro tests existentes deben seguir en verde SIN TOCARLOS**:
  > `internal/transport/transport_test.go:151` · `internal/config/config_test.go:61` ·
  > `internal/config/enrollment_test.go:97` · `cmd/permea/enroll_reject_test.go:129`.
  > **Si hay que modificar alguno, la unificación cambió comportamiento: SE PARA Y SE REPORTA.**
  > No se «ajusta el test»: el test es la red, y una red que se ajusta no sujeta nada.
  > ### ✅ PRUEBA POSITIVA DE UNIFICACIÓN — mutar el original y ver moverse a los cuatro
  > La condición de parada de arriba es **negativa**: dice que nada se rompió. Y **«nada se rompió» es
  > compatible con «no pasó nada»**: una T005 que **dejara las cuatro réplicas en pie** y se limitara a
  > cambiar el centinela de `Adherir` **pasaría el checkpoint entero** —los cuatro tests verdes, T011
  > verde— sin haber unificado nada.
  >
  > **T011 no lo caza**, y no puede: observa comportamiento, y dónde vive el juicio es estructura (ver
  > su §LÍMITE CONOCIDO). **Esto sí lo caza:**
  >
  > **Al terminar la sustitución, MUTAR LA FUNCIÓN UNIFICADA de T004** —alterar el juicio de esquema
  > en su único cuerpo— **y comprobar que LOS CUATRO LLAMANTES CAMBIAN A LA VEZ**:
  >
  > 1. `internal/config/enrollment.go` — su test de tabla debe caer;
  > 2. `internal/config/config.go` — su test de `Validate()` debe caer;
  > 3. `internal/transport/transport.go` (`Send`) — `TestSend_RejectsHTTP` debe caer;
  > 4. `internal/transport/transport.go` (`Adherir`) — **T011 debe caer**.
  >
  > **⛔ Si UNO SOLO no se mueve, ese llamante NO está unificado: conserva su copia. SE PARA Y SE
  > REPORTA.** Es la misma forma que este fichero ya usa en **T001** —cuerpo compartido: alterarlo
  > cambia las dos identidades a la vez— y en la **cláusula (a) de T028** —punto único del que salen
  > las dos—. **Una copia no se mueve cuando muta el original**, y eso sí es observable.
  >
  > **TRANSCRIBIR los cuatro mensajes de fallo** (disciplina 2 y 3) y **revertir POR EDICIÓN INVERSA**,
  > nunca por `git checkout` — para que el diff final demuestre que se revirtió lo mismo que se
  > introdujo.
  > #### ⛔ CÓMO SABER QUE LA MUTACIÓN VALE — dos comprobaciones, y son obligatorias
  > El párrafo de arriba dice **qué mutar**. Esto dice **cómo saber que la mutación acredita algo**, que
  > no es lo mismo y **se aprendió midiendo** en T004 (disciplina 3 §mutación válida).
  >
  > **(a) ANTES de ejecutar ningún test: `go build ./...` DEBE PASAR con la mutación puesta.** Si no
  > compila, **la mutación se descarta y se busca otra forma de alterar el mismo juicio** — no se apunta
  > como mutación superada.
  > > **Y aquí no es higiene, es la diferencia entre medir y engañarse.** Ésta es la **única mutación de
  > > la feature donde el fallo de método produce EL ASPECTO DEL ÉXITO**: muta la función que usan
  > > **cuatro llamantes en tres paquetes**, así que una mutación mal escrita **rompe la compilación de
  > > los cuatro a la vez** — y **un `build failed` en los cuatro paquetes SE LEE EXACTAMENTE IGUAL que
  > > «los cuatro llamantes se movieron»**, que es justo lo que esta mutación existe para demostrar.
  > > **La mutación tiene que hacer fallar TESTS, no el `go build`.**
  >
  > **(b) DESPUÉS: los cuatro tests deben fallar con CUATRO MENSAJES DISTINTOS**, transcritos **uno a
  > uno y atribuidos a su llamante** (enrollment · config · `Send` · `Adherir`). **Cuatro fallos
  > idénticos no distinguen «se movieron los cuatro» de «se rompió algo común»** — y con un cuerpo
  > compartido recién introducido, «algo común» es la hipótesis más probable, no la menos. Cuatro
  > mensajes distintos, cada uno con la voz de su llamante, **son la prueba de que el juicio llegó a
  > los cuatro sitios**; cuatro copias del mismo texto no prueban nada.
  > ### ⛔ DEBER DE CIERRE — hacer FALSABLE la aserción de sentido del CASO 2 de T011
  > **Al terminar la unificación**, y además de la mutación de arriba: comprobar que la aserción de
  > sentido del **caso 2 de T011** —«**no analizable NO es `ErrScheme`**»— **es falsable**. Hacer que la
  > rama no analizable devuelva `ErrScheme` **debe TUMBARLA**.
  >
  > **Por qué hace falta, y por qué SÓLO PUEDE HACERSE AQUÍ:** hoy esa aserción **pasa por accidente**
  > —no es `ErrScheme` porque **nada** lo es, ni siquiera el canal en claro—, así que **verde no
  > significa nada** mientras dure el andamiaje. **El comportamiento que podría violarla no existe hasta
  > que T005 unifica**: sólo después hay un `ErrScheme` de verdad en esta puerta, y sólo entonces
  > «devolverlo también para lo no analizable» es un error posible en vez de una imposibilidad.
  >
  > **Es exactamente el fallo que T005 puede introducir**: refundir los dos hechos en un solo desenlace
  > es la forma natural de simplificar al unificar, y **el caso 2 es lo único que lo impide**. Una
  > guardia que no se puede disparar no guarda nada — y ésta se queda así de por vida si nadie la
  > dispara **el día en que empieza a poder dispararse**.
  >
  > **Revertir POR EDICIÓN INVERSA**, y transcribir el mensaje. **Si NO la tumba, la aserción no está
  > mirando: SE PARA Y SE REPORTA** — la unificación habrá dejado el caso 2 verde y vacío, que es peor
  > que no tenerlo, porque parece una red.
- [ ] **T006** Remisión cruzada en los dos sitios (D-005-P2 §encontrabilidad): en
  `internal/transport/transport.go` —donde estaba la condición— un comentario que dice **dónde vive
  ahora el juicio y por qué se movió**; en `internal/config/` —donde vive— **quiénes son sus
  llamantes**. Sin las dos, unificar mejora el código y **empeora el hallazgo de la frontera**
  (Principio III).
  ⚠️ **Son CUATRO, y los cuatro existen ya** —`enrollment.go`, `config.go`, `transport.go` (`Send`) y
  `transport.go` (`Adherir`)—, porque **T002 creó el cuarto en Phase 1**. La lista se escribe con los
  cuatro, sin ningún «pendiente».
  *(Esta nota decía «hoy son tres, el cuarto llega en T012». **Dejó de ser cierto al ejecutarse T002**:
  el llamante existe desde Phase 1, y quien lo busque lo encuentra.)*

**Checkpoint**: **9 ok** · los cuatro tests de la red **sin una línea tocada** · **y T011 EN VERDE**.

> **Las dos mitades del checkpoint dicen cosas distintas, y hacen falta las dos.** «Los cuatro
> intactos» es el criterio **negativo**: nada se rompió. **«T011 en verde» es el positivo**: la
> unificación **llegó al método nuevo**. Sin él, una T005 que extrajera la función y **se olvidara de
> `Adherir`** pasaría el checkpoint entero — porque «nada se rompió» es compatible con «no pasó nada».

---

## Phase 3: LA FRONTERA — el golden test va PRIMERO (Principio IV)

> **Disciplina de primer commit.** Esta feature **abre la segunda puerta de la frontera de datos**, y
> el golden existente cubre la emisión de eventos, no la adhesión.

- [ ] **T007** Crear el **golden test de frontera de la adhesión** en
  `internal/transport/boundary_adhesion_test.go` (nuevo): el cuerpo de la petición contiene
  **exactamente `{code, project_ref}` y nada más** — allowlist de dos elementos—, y **ningún otro dato
  de la instalación** viaja en él. **Nace verde** (T002 ya construye el cuerpo) → **se valida por
  mutación**: añadir un tercer campo al cuerpo debe **tumbarlo**, y el mensaje debe leerse.
  ⚠️ **En ESTA MISMA TAREA**, la nota en `internal/ingest/boundary_test.go` (D-005-P14): su cabecera
  `:106-107` afirma que FR-017 *«la nombra entera»* nombrando tres caminos, y **esa frase deja de ser
  cierta** en cuanto exista una segunda puerta. Nota en los dos ficheros, **y no suelta**: separarla
  es cómo se queda sin hacer.

**Checkpoint**: la frontera tiene **dos testigos**, y cada uno sabe del otro.

---

## Phase 4: El transporte — los cuatro desenlaces (D-005-P1)

### Tests (rojo antes de verde)

- [ ] **T008** [P] Test del **desenlace de éxito** en `internal/transport/adhesion_test.go` —
  **el fichero ya existe: lo creó T011 en Phase 2** al mudarse allí—
  **Garantía**: `adhesion.md` desenlaces 3 y 4. `200` con `{"project":{"name":…}}` → devuelve **la
  denominación**. Nace en **rojo**: T002 devuelve siempre «adhesión no implementada». Transcribir la
  razón.
- [ ] **T009** [P] Test de **los dos rechazos, distintos entre sí** en el mismo fichero —
  **Garantía**: desenlaces 1 y 2. `422` → rechazo; `409` → conflicto; **y son ramas distintas**. Nace
  en **rojo**.
  ⚠️ **Ramificar por el ESTADO**, no por el cuerpo (`adhesion.md` §Qué distingue a qué). El cuerpo se
  comprueba como confirmación, no como discriminante.
- [ ] **T010** [P] Test de **respuesta ininterpretable** en el mismo fichero — **Garantía**: FR-002 +
  FR-013. Un `200` **sin `project.name` legible** → **no verificable**, **NUNCA éxito**. Nace en
  **rojo** porque **T002 devuelve «adhesión no implementada», que es OTRO centinela**. *Es el riesgo
  declarado de D-005-P1: un éxito sin nombre no es un éxito.*
  ⚠️ **Este es el rojo más frágil del fichero** y por eso T002 lo protege explícitamente: si el
  andamiaje devolviera «no verificable», **este test nacería verde acertando contra el stub**.

### Implementación

- [ ] **T012** Implementar el método de adhesión en `internal/transport/transport.go`: llama a la
  guarda (T004), compone la petición, **lee y decodifica** la respuesta y **distingue por estado**.
  Reutiliza el `http.Client`, su timeout y la cabecera de autenticación. **Un solo intento, sin cola**
  (D-005-P4, siguiendo `Verify()` en `transport.go:91-99`). Pone en verde **T008, T009 y T010**.
  ⚠️ **Ya NO pone en verde T011**: esa la puso **T005**, en Phase 2, al sustituir la réplica de
  andamiaje de `Adherir` por la guarda unificada. Cuando T012 llega, **T011 lleva dos fases en verde**.
  ⚠️ **T012 ya NO añade ningún llamante a la lista de T006**: el cuarto lo creó **T002** en Phase 1 y
  lo reconvirtió **T005**. Aquí solo se implementan **los desenlaces** —leer el cuerpo, distinguir por
  estado—; **la guarda ya está puesta y ya es la unificada**. *(Esta tarea llevaba el deber de
  «añadirse a la lista»; se retiró al ejecutarse T002, que adelantó el llamante a Phase 1.)*

---

## Phase 5: El destino derivado (D-005-P3)

- [ ] **T013** [P] Test de **derivación correcta** en `internal/config/config_test.go` — **Garantía**:
  `adhesion.md` §Cómo se obtiene `<base>`. Conserva **esquema, host, puerto y prefijo**, y sustituye
  solo el último segmento. **El puerto no estándar es caso obligatorio**: el banco local usa `:8443`.
  Nace en **rojo**.
- [ ] **T014** [P] Test de **forma inesperada** en el mismo fichero — **Garantía**: FR-009 **y
  FR-020**. Un endpoint cuya ruta no termina en el segmento conocido → **rehúsa**, nombrando **la
  forma** de lo hallado, y **el mensaje NO contiene material sensible** aunque estuviera en lo
  hallado. Nace en **rojo**.
- [ ] **T015** Implementar la derivación con **validación ruidosa** en `internal/config/`. **Cita el
  contrato** (`contracts/adhesion.md` §Cómo se obtiene `<base>`) en el comentario: los dos hechos de
  ruta son **contrato, no literal local** (D-005-P3, D-005-P11). Pone en verde **T013–T014**.

---

## Phase 6: El comando (D-005-P6, D-005-P7, D-005-P13)

### Tests (rojo antes de verde)

- [ ] **T016** [P] Test de **las dos vías de entrada** en `cmd/permea/project_test.go` (nuevo) —
  **Garantía**: FR-023 + SC-011 (A). Mismo código por argumento y por entrada estándar → desenlaces
  **idénticos**, con las **tres piezas**: no vacío y del tipo que toca · idénticos entre sí · **y la
  comparación sabe fallar**. Nace en **rojo**.
- [ ] **T017** [P] Test de **entrada ausente** en el mismo fichero — **Garantía**: `cli.md` §Entrada.
  Sin argumento y sin pipe → **error de uso** con **`ExitCode() == 1`**, y **NUNCA un prompt que se
  cuelgue**. Nace en **rojo**: T003 sale con **70**.
  ⚠️ **Se compara el código EXACTO, no «≠ 0»**, y es lo que hace que el 70 de T003 sirva de algo: con
  «≠ 0» esta tarea **nacería VERDE contra el andamiaje** —70 también es ≠ 0— y su rojo no existiría.
  El valor es **1** porque el binario tiene **dos** códigos, `0` y `1` (`cli.md` §Los códigos de
  salida), y el error de uso no es éxito.
- [ ] **T018** [P] Test del **ORDEN de los tres rehúses** en el mismo fichero — **Garantía**:
  D-005-P13, `cli.md` §Comportamiento. Con las tres condiciones a la vez —sin árbol, sin enrolamiento
  y con configuración rota— el mensaje es **el del árbol**. Nace en **rojo**.
  *Sin este test el orden lo fija el primer camino que alguien escriba.*
- [ ] **T019** [P] Test de **cero peticiones fuera de árbol** en el mismo fichero — **Garantía**:
  SC-004, FR-006. **Con su observador declarado**: un destino instrumentado que **cuenta**. **Y con su
  caso positivo**: el **mismo** destino, con el comando lanzado **dentro** de un árbol, **registra
  exactamente una**. Nace en **rojo**.
  ⚠️ **Sin el caso positivo el test se cumple por no mirar**, y entonces no distingue «no se emitió» de
  «el observador no estaba conectado».
  ⚠️ **Por qué T019 NO lleva la marca «nace verde» que sí llevan T024–T026**, aunque también verifique
  una ausencia: **su caso positivo está DENTRO del mismo test** —el observador debe registrar
  exactamente una petición **dentro** de un árbol— y **eso sí falla** con el andamiaje de T003, que no
  emite nunca. **Nace rojo de verdad, por la mitad positiva.** No es un olvido.
- [ ] **T020** [P] Test de **verbo desconocido y `project` sin verbo** en el mismo fichero —
  **Garantía**: `cli.md` §La gramática. Error de uso por stderr, **`ExitCode() == 1`**, nombrando lo no
  reconocido. Nace en **rojo**: T003 sale con **70**.
  ⚠️ **El código EXACTO, por la misma razón que T017**: con «≠ 0» el andamiaje la satisface y el rojo
  desaparece.

### Implementación

- [ ] **T021** Implementar el comando en `cmd/permea/project.go` con el **patrón de dos capas de
  `enroll`** (D-005-P7, `enroll.go:18` y `:38`): capa sucia que resuelve stdin/TTY/stdout, capa pura
  con **lector, escritor y ejecutor inyectados**. Los tres rehúses **en el orden de T018**, **antes de
  emitir nada**. Pone en verde **T016–T020**.
  ⚠️ **La inyección no es comodidad**: `main_test.go:321-325` documenta que un proceso hijo **no
  confiaría** en el certificado del arnés, así que es **la única vía** de probar el camino completo.

---

## Phase 7: Canales, salidas y secretos

### Tests (rojo antes de verde)

- [ ] **T022** [P] Test del **reparto de canales** en `cmd/permea/project_test.go` — **Garantía**:
  FR-021, SC-011 (B). Éxito → **stdout no vacío y stderr sin el desenlace**; rehúse o error → **stderr
  no vacío y stdout VACÍO**. **Capturados por separado** (disciplina 7). Nace en **rojo**.
- [ ] **T023** [P] Test de **los ocho códigos de salida** en el mismo fichero — **Garantía**: `cli.md`
  §Los códigos de salida. Compara **`ExitCode()`, nunca texto** (disciplina 4).
  > ⚠️ **El caso que no puede faltar**: **D3 y D4 comparten código**. Si alguien le diera a «ya
  > estabas unido» un código propio, **rompería FR-010** sin darse cuenta — el resultado del proceso
  > es observable. Este test es lo único que lo impide.
  >
  > ⚠️ **Y ES QUIEN RETIRA EL 70 DE T003.** Comparar los ocho códigos exactos **tumba cualquier
  > superviviente del andamiaje**, así que la retirada está **garantizada por el test** y no confiada
  > a que alguien se acuerde.
- [ ] **T024** [P] Test de **no filtración del código** en el mismo fichero — **Garantía**: FR-020,
  SC-005. Para **cada uno de los ocho desenlaces** (los del comando, `cli.md`): generar las subcadenas
  de **longitud ocho** del valor presentado y buscarlas **en los dos canales**. **Cero apariciones.**
  ⚠️ **NACE VERDE, y hay que decirlo**: un comando sin implementar **no imprime el código** —no imprime
  casi nada—, así que la ausencia se cumple sola. **Se valida por SU CASO POSITIVO**, que ya está
  escrito en la propia tarea: sembrar deliberadamente el código en una salida debe **tumbar el test**,
  y hay que **leer el mensaje**. Sin eso, no distingue «no se filtra» de «no hay salida».
- [ ] **T025** [P] Test de **nada se escribe en local** en el mismo fichero — **Garantía**: FR-019,
  SC-007. Captura **íntegra** del conjunto enumerado de artefactos —configuración, estado, cola,
  secretos— **antes**, y comparación byte a byte **contra esa captura** tras cada desenlace.
  ⚠️ **NACE VERDE**: un comando sin implementar **no escribe nada**, así que la ausencia se cumple
  sola. **Se valida por SU CASO POSITIVO**, ya escrito: una operación de la instalación que **sí**
  modifica el estado local **hace fallar la misma comparación**. Si no falla ahí, el test no distingue
  «no cambió» de «no miré».
- [ ] **T026** [P] Test de **la petición nunca se encola** en el mismo fichero — **Garantía**: FR-018,
  SC-010. Con el servidor **inalcanzable**, la cola inspeccionada **antes y después** no crece.
  ⚠️ **NACE VERDE**: un comando sin implementar **no encola nada**, así que la ausencia se cumple sola.
  **Se valida por SU CASO POSITIVO**, ya escrito: una emisión ordinaria de eventos con el destino
  igualmente caído **sí la hace crecer**. Si la cola no crece en ninguno de los dos casos, **el
  observador no está mirando** y el criterio no cuenta como pasado.

### Implementación

- [ ] **T027** Implementar la **presentación de los desenlaces** en `cmd/permea/project.go`: el mapeo
  de `cli.md` §Comportamiento —canal y código de salida por desenlace—, con **D3 y D4 produciendo
  salida idéntica**, **D2 sin nombrar el Proyecto ajeno** y **D1 sin indicar la causa**.
  **Pone en verde T022 y T023.**
  > ### ⚠️ Y AQUÍ SE EJECUTAN LOS TRES CASOS POSITIVOS — es deber de esta tarea
  > **T024, T025 y T026 nacieron VERDES** (una ausencia la satisface un comando sin implementar), así
  > que **su validación no es «ponerlos en verde»: es ejecutar su caso positivo ahora que hay
  > comportamiento que pueda violarlos.** Los tres, con el **mensaje de fallo leído** (disciplina 3):
  >
  > - **T024** — sembrar deliberadamente el código en una salida **debe tumbarlo**;
  > - **T025** — una operación que **sí** escribe en local **debe hacer fallar la comparación**;
  > - **T026** — una emisión ordinaria con el destino caído **debe hacer crecer la cola**.
  >
  > **Si alguno no falla, no está mirando**, y el criterio que dice sostener —SC-005, SC-007, SC-010—
  > **no cuenta como pasado**. Decir «T027 pone en verde T022–T026» habría sido cómodo y falso: tres de
  > los cinco ya estaban verdes, y lo que les faltaba era exactamente esto.

---

## Phase 8: Los TRES criterios que ninguna historia arrastra (D-005-P9)

> ⚠️ **Tarea propia, y no cuelgan de ninguna historia.** La matriz del checklist lo dejó medido: las
> columnas de SC-001, SC-008 y SC-009 están **enteras en «—»**. **Un troceo por historias los habría
> dejado fuera**, y son los tres que protegen lo que ya funciona.

- [ ] **T028** **SC-001** — en **`internal/project/resolve_test.go`**: la identidad presentada **es** la
  estampada, demostrado por **origen compartido**: (a) punto único del que salen las dos, y alterarlo **cambia las dos a la vez**; (b)
  comparación sobre **cuatro clases de árbol** —raíz, subdirectorio profundo, árbol paralelo,
  directorio sin raíz—; (c) **una alteración deliberada en el punto único pone (b) en rojo**.
  **Sin (c) el criterio no es falsable.**
  ⚠️ **En las TRES primeras clases se compara carácter a carácter que las dos identidades son la
  misma. En la cuarta —directorio sin raíz— NO hay identidades que comparar**, y hay que escribirlo
  porque invita a un test que compara dos valores vacíos y da verde por nada:
  - **el comando NO presenta identidad ninguna**: rehúsa antes de emitir (FR-006), así que no hay
    valor presentado;
  - **lo que se compara ahí es el JUICIO, no el valor**: que la vía nueva reporte «no hubo raíz» y que
    el comando **rehúse por esa razón** — las dos coinciden en el mismo veredicto;
  - *(la derivación **sí produce un valor** en esa clase —cae al fallback del directorio, P-004
    FR-005— pero ese valor **nunca se presenta**, y compararlo contra nada sería inventar un sujeto.)*
  ⚠️ **NO usar `--scan` para obtener las identidades**: usa un salt literal (`cmd/permea/main.go:342`)
  y sus refs **no comparan con nada** (`research.md` §R5).
- [ ] **T029-R** **SC-009 · RE-EJECUCIÓN, puerta de SC-009** — volver a pasar
  **`internal/ingest/baseline_regresion_test.go`** (el de T029) con **todo el código de la feature ya
  escrito**. Es lo que **acredita** que la feature entera **no cambió el camino de ingesta**.
  ⚠️ **Aquí NO se escribe nada: se vuelve a pasar.** El test se escribió y se ejecutó en **T029, Phase
  1**, donde su única dependencia —T001— ya estaba satisfecha. **Detectar en Phase 1, acreditar aquí.**
  ⚠️ **Con otras semillas los refs no comparan y un «fallo» no significaría nada.** Mismo procedimiento
  que 004 usó en su T007.
- [ ] **T030** **SC-008** — en **`internal/transport/adhesion_test.go`**: sin transporte seguro no se
  completa, en **las cuatro clases enumeradas**:
  (a) destino en claro · (b) destino en claro sobre la máquina local · (c) destino inseguro **con un
  código utilizable**, para que el rechazo **no pueda atribuirse al código** · (d) los tres anteriores
  con **cada** ajuste de configuración que la instalación admita.

---

## Phase 9: Documentación (D-005-P12)

> **Es tarea de la feature, no un extra.** Un comando que nadie puede descubrir no existe.

- [ ] **T031** En `README.md`, **crear la sección de comandos que hoy NO existe** y documentar en ella
  `project join` **junto a `enroll` y `status`** — medido: el README documenta los cuatro flags y **no
  menciona los dos subcomandos**. Con la vía de entrada estándar señalada como **recomendada**, igual
  que hace `cli.md` de 003.
  ⚠️ **No se arregla `README.md:74`** —documenta el campo retirado «modo de ref», y quien lo siga **no
  arranca**—: es **deuda anterior**, va al backlog y **no se cuela en esta feature** (D-005-P12).
- [ ] **T032** Actualizar `printUsage` en `cmd/permea/main.go:98-113` —literal mantenido a mano— para
  que liste `project join`. **Un comando que no aparece en la ayuda no existe para quien lo busca**
  (`cli.md` §La gramática).

---

## Phase 10: Ceremonia — validación manual contra plataforma real (ejecuta Basilio)

> ## ⚠️ ESTO NO SON TAREAS DE TEST, Y AUTOMATIZARLAS ESTÁ PROHIBIDO
>
> **SC-006** lo declara en la spec: lo que observan es comportamiento **de la plataforma**, y la suite
> **tendría que fabricar la agrupación para probarlo**, con lo que estaría comprobando su propio
> simulacro. **Un test automático que lo finja es peor que no tenerlo.**

Estas tareas **no las ejecuta Claude Code**. Referencian los **C-números** de `quickstart.md` §Parte B
y **no reescriben sus pasos**.

- [ ] **T033** ⛔ **PRERREQUISITO** — banco TLS local levantado (`quickstart.md` §El banco TLS), agente
  enrolado contra `https://localhost:8443/…`, y **un código de adhesión acuñado desde el panel**.
  **FR-017 no tiene exención ni para la ceremonia.**
  ⚠️ **Y ANOTAR, ANTES DE EMPEZAR, el recuento de eventos de `~/dev/test/RecetApp` DEL DÍA.** Es el
  número contra el que compara T034, y **hay que capturarlo antes de tocar nada**: si alguien ha
  trabajado ahí desde que se identificó el sujeto, la cifra es otra — y **lo que la ceremonia
  comprueba no es una cifra absoluta, sino que NO CAMBIE durante la unión**. Sin anotarlo antes, T034
  compararía contra un recuerdo.
  ⚠️ **Y TOMAR TAMBIÉN LA CAPTURA PREVIA DEL DIRECTORIO DE DATOS DEL AGENTE**, por la misma razón:
  **medido — `quickstart.md` §C4 dice «comparar el directorio de datos con su estado previo» y NINGÚN
  paso de la Parte B toma ese estado previo.** «Byte a byte igual» necesita un «igual a qué» **tomado
  antes**, o T036 compara contra un recuerdo igual que lo haría T034 sin el recuento. Las dos capturas
  —recuento y directorio— se toman **aquí, antes de tocar nada**.
- [ ] **T034** Ejecutar **C1** y **C2** de `quickstart.md` sobre **`~/dev/test/RecetApp`** —sin
  mapear—: la unión nombra el Proyecto, **el consumo previo aparece bajo él**, y **el número de eventos
  NO cambia respecto al recuento ANOTADO EN T033**. *(SC-006, US1 escenario 2.)*
  ⚠️ **El recuento es la comprobación que distingue «se agrupó en lectura» de «se reprocesó algo»**: si
  se moviera, algo escribió, y FR-003 lo prohíbe.
- [ ] **T035** Ejecutar **C3** — repetir es indistinguible, **en salida y en código de salida**, y en el
  panel la instalación sigue unida **una sola vez**. *(US2 escenario 3, FR-010.)*
- [ ] **T036** Ejecutar **C4** — el directorio de datos del agente, **byte a byte igual a la captura
  tomada en T033**. *(FR-019.)*
  ⚠️ **La captura previa la toma T033, no C4**: `quickstart.md` §C4 dice «con su estado previo» pero
  **no toma ninguno** —comprobado—, así que el «estado previo» es el de T033. Sin él, esta tarea no
  tiene contra qué comparar.

---

## Dependencies & Execution Order

```
Phase 1 (extensión)     T001, T002, T003  ─►  T029  (se ESCRIBE y ejecuta aquí)
        │
Phase 2 (guarda)        T004 ─► T011 (rojo) ─► T005 ─► T006   ⛔ parada si hay que tocar los 4 tests
        │
Phase 3 (frontera)      T007                         Principio IV — antes de la lógica
        │
Phase 4 (transporte)    T008, T009, T010 (rojo) ─► T012
        │
Phase 5 (destino)       T013, T014 (rojo) ─► T015
        │
Phase 6 (comando)       T016..T020 (rojo) ─► T021
        │
Phase 7 (canales)       T022..T026 (rojo) ─► T027
        │
Phase 8 (sin historia)  T028, T030  ·  T029-R (RE-ejecución de T029, puerta de SC-009)
        │
Phase 9 (docs)          T031, T032
        │
Phase 10 (ceremonia)    T033 ─► T034 ─► T035 ─► T036          manual, Basilio · SECUENCIAL
```

### Dependencias que no son de fase

- **T005 depende de T011, y no al revés — la relación se invirtió.** Antes se leía «T011 depende de
  T005»: había que esperar a la unificación para que el test pasara. **Ahora es T005 quien necesita que
  T011 exista y esté ROJA antes de empezar**, porque **T011 es su criterio positivo**: es lo único que
  demuestra que la unificación **llegó al método nuevo** en vez de limitarse a no romper nada.
  **Orden obligado en Phase 2: T004 → T011 (rojo) → T005 (la pone verde) → T006.**
- **T029 depende SOLO de T001**, y por eso **tiene dos casillas**: **T029** la escribe y ejecuta en
  Phase 1 —en cuanto T001 existe, sin esperar a nada más— y **T029-R** la re-ejecuta en Phase 8 como
  puerta de SC-009. Dejarla solo al final habría hecho que una regresión de T001 se descubriera **con
  veintiocho tareas de por medio** y veintiocho sospechosos.
- **T022–T026 dependen de T012 y T021**: no puede comprobarse el canal de un desenlace que aún no se
  produce (*un bloque no implementa un desenlace cuyo disparador llega en otro bloque*).
- **T007 depende de T002**, no de T012: el golden necesita **el cuerpo**, no la lógica de desenlaces.

### Las puertas de rojo, validadas CONTRA EL ORDEN

Para cada tarea que declara un rojo, se comprobó **qué la precede** y si alguna ya satisface lo que esa
puerta espera ver fallar:

**Y la respuesta se da POR TEST, no por bloque**: agrupar «T022–T026» y responder «no» de una vez es
justo cómo se cuela un verde vacío en un grupo de cinco.

| Test | Lo precede | ¿Nace rojo? | Por qué |
|---|---|:--:|---|
| **T008** | T001–T007 | **rojo** | T002 devuelve «adhesión no implementada», no una denominación |
| **T009** | T001–T007 | **rojo** | T002 no distingue `422` de `409`: devuelve el mismo centinela para los dos |
| **T010** | T001–T007 | **rojo** | **solo porque el centinela de T002 es OTRO.** Si T002 devolviera «no verificable», este test **nacería verde acertando contra el stub** — es la razón de que el cambio 1 exista |
| **T011** (1) *el centinela* | **T001–T004** | **rojo** | **T002 SÍ tiene guarda —una réplica inline— pero devuelve `errEsquemaAndamiaje`, no `ErrScheme`**, que es el centinela que este test exige. Nace rojo **por el centinela, no por ausencia de guarda**. **Vive en Phase 2, entre T004 y T005**: en Phase 4 habría nacido VERDE, porque T005 ya habría puesto `ErrScheme` dos fases antes |
| **T011** (2) *la distinguibilidad* | **T001–T004** | **rojo** | **la réplica de T002 devuelve el MISMO `errEsquemaAndamiaje` para los dos hechos**, así que «no analizable» y «esquema no admisible» son **indistinguibles**. Rojo **por indistinguibilidad**, que es un rojo distinto del de (1): (1) mira **qué** centinela; (2) mira que **no sea el mismo para los dos**. **Sin este caso, T005 puede ponerse verde REFUNDIENDO lo que T004 separó** |
| **T011** (3) *la causa conservada* | **T001–T004** | **rojo** | **la réplica de T002 envuelve el CENTINELA y DESCARTA la causa de `url.Parse`** — la **cuarta variante**, al revés que `Send` (`transport.go:143`), que **conserva la causa**. Tercer rojo **distinto** de los otros dos: (1) y (2) miran **el centinela**; (3) mira **lo que hay debajo de él**. Su rama de oráculo —la misma propiedad sobre `Send`— **pasa hoy**, así que el rojo es **atribuible sólo a `Adherir`** |
| **T013** | T001–T012 | **rojo** | la derivación del destino no existe hasta T015 |
| **T014** | T001–T012 | **rojo** | ídem: no hay validación ruidosa que rehusar |
| **T016** | T001–T015 | **rojo** | T003 rehúsa siempre; no hay dos vías que comparar |
| **T017** | T001–T015 | **rojo** | T003 sale con **70**, no con el código del contrato |
| **T018** | T001–T015 | **rojo** | T003 no evalúa los tres rehúses, así que no hay orden que observar |
| **T019** | T001–T015 | **rojo** | T003 no emite peticiones **y tampoco las emite dentro de un árbol**: el **caso positivo** —que el observador registre exactamente una— es el que falla |
| **T020** | T001–T015 | **rojo** | T003 sale con **70** ante un verbo desconocido, no con el del contrato |
| **T022** | T001–T021 | **rojo** | T021 implementa entrada y rehúses; **la presentación llega en T027**, así que no hay reparto de canales que comprobar |
| **T023** | T001–T021 | **rojo** | ídem: los ocho códigos los fija T027 |
| **T024** | T001–T021 | ⚠️ **VERDE** | **una ausencia la satisface un comando que no imprime.** Se valida por **su caso positivo** |
| **T025** | T001–T021 | ⚠️ **VERDE** | **una ausencia la satisface un comando que no escribe.** Se valida por **su caso positivo** |
| **T026** | T001–T021 | ⚠️ **VERDE** | **una ausencia la satisface un comando que no encola.** Se valida por **su caso positivo** |
| **T029** | T001 | ⚠️ **VERDE** | **es regresión cero: si T001 está bien, nada cambió y da verde de entrada.** Se valida **por mutación del cuerpo compartido de T001** — y es la única puerta del camino de ingesta, así que un test que no mire da el mismo verde que uno correcto |

> **Por qué esta comprobación tiene sección propia**: *«ya mordió dos veces en P-010»*. Un rojo que
> nace verde porque una tarea anterior ya lo satisfizo **no es un rojo**, y el test que lo sigue no
> demuestra nada.
>
> **Y por qué tres de ellos nacen verdes y se declara**: **T024, T025 y T026 verifican AUSENCIAS**
> —que el código no aparezca, que no se escriba, que no se encole— y **una ausencia la satisface
> trivialmente un comando sin implementar**. Declararlos «rojos» habría sido cómodo y falso. Los tres
> ya traían su caso positivo escrito, y **ese caso positivo ES su validación** (disciplina 3): sin él
> no distinguen «no ocurrió» de «no hay nada que pudiera ocurrir».

### Oportunidades de paralelismo

`T001`/`T002` · **`T008`–`T010`** · `T013`/`T014` · `T016`–`T020` · `T022`–`T026` · `T028`/`T030`.

> **T011 salió del grupo de Phase 4** al mudarse a Phase 2, donde **no es paralelizable con nada**: su
> sitio en la secuencia `T004 → T011 → T005` **es el mecanismo**, no una preferencia de orden.

> ⚠️ **La ceremonia NO es paralelizable, y T034–T036 quedan FUERA de esta lista.** Es **estrictamente
> secuencial**: **C3 presupone que C1/C2 ya unieron** —no se puede comprobar que «repetir es
> indistinguible» sin una primera vez— y **C4 comprueba el estado tras todas las ejecuciones
> anteriores**. Ejecutarlas en otro orden, o a la vez, **no comprueba lo que dicen comprobar**.

---

## Tabla de cobertura — **con columna de SUPERFICIE**

> ⚠️ **Por qué la columna existe.** P-013 cerró con **54/54 FR en verde y la pantalla no permitía
> ejecutar la ceremonia**: cinco verbos con prueba de backend, de contrato y de API, **y ninguna que
> comprobara que hubiera desde dónde invocarlos**. Costó una jornada. Un requisito que cruza capas
> necesita **una fila por superficie**, no una fila con tres citas.

**Superficies**: **TR** transporte · **CMD** comando · **CFG** configuración · **PRJ** derivación ·
**DOC** documentación · **CER** ceremonia manual.

### Los 24 requisitos funcionales

| FR | Superficie | Tarea |
|---|:--:|---|
| FR-001 | TR · CMD | T012 · T021 |
| FR-002 | TR · CMD | T008, T010 · T027 |
| FR-003 | CER | T034 *(el recuento que no cambia)* |
| FR-004 | PRJ | **T001** *(impl: cuerpo compartido)* · T028 |
| FR-005 | PRJ | T001 *(cuerpo compartido)* · T028 |
| FR-006 | PRJ · CMD | T001 · T018, T019, T021 |
| FR-007 | CMD | T018, T027 |
| FR-008 | CMD | T018, T021 |
| FR-009 | CFG · CMD | T014, T015 · T018 |
| FR-010 | TR · CMD · **CER** | T008 · **T023** *(mismo código de salida)*, T027 · **T035** |
| FR-011 | TR · CMD | T009 · T027 |
| FR-012 | TR · CMD | T009 · T027 |
| FR-013 | TR · CMD | T010 · T027 |
| FR-013a | CMD | T027 *(el mensaje deja volver a intentarlo)* |
| FR-014 | PRJ | T001 · **T029**, T029-R |
| FR-015 | PRJ | T001 · T029, T029-R |
| FR-016 | PRJ | T001 · T029, T029-R |
| FR-017 | TR · CFG · **CER** | **T004, T005** *(impl, Phase 2)* · **T011** *(su prueba, Phase 2)* · T012, **T030** · T033 *(ni la ceremonia exime)* |
| FR-018 | TR · CMD | T012 · **T026** |
| FR-019 | CMD · **CER** | **T025** · **T036** |
| FR-020 | CFG · CMD | T014, **T015** · **T024**, **T027** |
| FR-021 | CMD | **T022** · **T027** *(impl: el reparto)* |
| FR-022 | TR | **T007** *(golden de frontera)* |
| FR-023 | CMD · DOC | T016, T017, T021 · T031, **T032** |

### Los 11 criterios de éxito

| SC | Superficie | Tarea | |
|---|:--:|---|---|
| SC-001 | PRJ | **T028** | sin historia |
| SC-002 | TR · CMD | T008 · T027 | |
| SC-003 | TR · CMD | T009 · T027 | |
| SC-004 | CMD | **T019** | con observador y caso positivo |
| SC-005 | CMD | **T024** | umbral de ocho |
| SC-006 | **CER** | **T034** | **no automatizable** |
| SC-007 | CMD | **T025** | con caso positivo |
| SC-008 | TR · CFG | **T030** | sin historia |
| SC-009 | PRJ | **T029-R** *(acredita; T029 la escribe en Phase 1)* | sin historia |
| SC-010 | CMD | **T026** | con caso positivo |
| SC-011 | CMD | T016 *(A)* · **T022** *(B)* | canales por separado |

**Los 24 FR y los 11 SC tienen tarea. Ningún hueco.**

### ⚠️ Verificación de recuento — **36 tareas, 37 casillas**, y no es un desajuste

**Un barrido mecánico contará 37 `- [ ]` y 36 identificadores `T0xx`, y dirá que sobra una.** No sobra:

**`T029` y `T029-R` son DOS CASILLAS DE UN SOLO MECANISMO** —el mismo fichero de test, escrito una vez
y pasado dos—. Tienen casilla separada porque **se cierran en momentos distintos**: T029 al acabar
Phase 1, cuando su única dependencia (T001) ya está y **un fallo tiene un solo sospechoso**; T029-R al
acabar Phase 8, con todo el código escrito, **acreditando** SC-009.

**Sin las dos casillas, una de las dos ejecuciones no tendría dónde registrarse** — y fue exactamente
el defecto que se corrigió el 2026-08-18: un checkpoint que exige una comprobación sin casilla es una
intención, no una puerta.

**Recuento oficial: 36 tareas · 37 casillas · 1 tarea con dos casillas (T029/T029-R).**

**Y cuatro tests nacen verdes, no tres**: **T024, T025, T026** (ausencias, validadas por su caso
positivo) **y T029** (regresión cero, validada por mutación del cuerpo compartido de T001).

### El barrido en el sentido que faltaba — de la TAREA al requisito

La comprobación de arriba va **de requisito a tarea**: garantiza que ningún FR/SC se queda sin quien lo
sirva. **Pero le falta el recíproco**, y es donde se esconde el desajuste silencioso: **toda tarea que
se declare a sí misma sirviendo un FR o un SC tiene que aparecer en la fila de ese FR o SC.** Si una
tarea dice «Garantía: FR-021» y la fila de FR-021 no la nombra, una de las dos miente — y como la tabla
es lo que se lee al cerrar, gana la tabla y el trabajo de esa tarea se pierde de vista.

**Recorridas las 36 tareas contra sus propias declaraciones de garantía. Tres desajustes, corregidos:**

| Hallazgo | Qué pasaba | Corregido |
|---|---|---|
| **T032** | se declara sirviendo la ayuda del binario —vía de entrada de FR-023, `cli.md` §La gramática— y **no aparecía en ninguna fila** | añadido a **FR-023 · DOC** |
| **T033** | se declara con *«FR-017 no tiene exención ni para la ceremonia»* y **no aparecía en FR-017** | añadido a **FR-017 · CER** |
| **T001** | citaba *«SC-001 demostrable en T027»*, y **T027 es la presentación de desenlaces**: SC-001 es **T028** | referencia corregida a **T028** |

Las otras 33 coinciden con sus filas. **Las tareas que NO declaran FR/SC no incumplen nada**: T020
sirve a `cli.md` §La gramática, que la spec no cubre con requisito propio, y así queda dicho.

#### ⚠️ Y un hueco que el barrido destapó en la SPEC, no en las tareas

**T013 —la derivación CORRECTA del destino, el camino feliz— no tiene FR que la cubra, y no se le ha
forzado una fila.** Medido sobre los 24 requisitos buscando `destino|dirigir|endpoint|ruta|alcanzar`:

- **FR-009 cubre solo el camino INFELIZ**: *«si la configuración **no permite determinar con
  confianza** a dónde dirigirse → rehúsa»*. Es el rehúse, no el acierto.
- **FR-001** dice *«DEBE poder unirla presentando un código»* — **el resultado, no el mecanismo**.
- **Ningún otro FR menciona el destino.**

**Y es coherente, no un descuido de la spec**: derivar el destino es **D-005-P3**, una decisión de
plan, y la spec **no puede tocar el CÓMO**. La spec no sabe —ni debe— que lo guardado sea la ruta
completa de ingesta. **T013 prueba una decisión de plan, no un requisito**, y por eso queda **sin fila
y declarada aquí** en vez de colgada de FR-001 para que la tabla parezca completa.

---

## Implementation Strategy

### El MVP no es una historia: es la Phase 4

Trocear por mecanismo tiene una consecuencia que conviene decir: **no hay un «solo US1» entregable**.
El primer punto con valor demostrable es **el final de Phase 6** —comando que se une y lo dice—, y el
primer punto **seguro** es el final de Phase 2, porque hasta ahí **nada ha cambiado de comportamiento**.

### Lo que este plan de tareas NO hace

- **No escribe el contrato**: `contracts/adhesion.md` y `contracts/cli.md` **ya existen**
  (**Phase 1 del plan**). La regla 4 de **`plan.md` §Phase 2 del plan** los mencionaba y **quedó
  ajustada el 2026-08-18**.
- **No arregla la deuda vieja del README** (`README.md:74`): al backlog.
- **No toca la plataforma.** La segunda mitad de D-005-P11 —que la plataforma cite el contrato en vez
  de definirlo— es **de otro día y de otra rama**.
- **No automatiza la ceremonia.** Está prohibido por la spec.
