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
   > ### ⛔ TODA ASERCIÓN DETRÁS DE UN `t.Fatalf` ES INMUNE A LA MUTACIÓN
   > **Y es un verde INALCANZABLE, no un verde vacío**, que es peor: no hay forma de tumbarla **mientras
   > la anterior aguante**. La mutación que debería probarla dispara el `Fatalf`, **el subtest se corta
   > ahí** y la aserción de abajo **nunca llega a evaluarse**. Pasa el test, pasa la revisión, y esa
   > línea **no ha probado nada en su vida**.
   >
   > **Medido en T010** (`tasks.md` T012 §PASO 2): sus dos aserciones «nunca éxito» y «ninguna
   > denominación» necesitaron **dos mutaciones distintas**, porque la primera es un `t.Fatalf`. Una
   > sola no podía tumbar las dos, y con una sola se habría dado por validada la segunda.
   >
   > **Regla:**
   > - **Aserciones independientes → `t.Errorf`**, para que **TODAS se evalúen** en la misma pasada.
   > - **`t.Fatalf` sólo para PRECONDICIONES**: aquello sin lo cual continuar no significa nada —el
   >   servidor no respondió, el fixture no existe, el caso base no se cumple—.
   > - **Si se encadenan de todas formas, hacen falta TANTAS MUTACIONES COMO ASERCIONES**, y cada una
   >   tiene que dejar en pie las anteriores para alcanzar la suya. Se escribe en la tarea, no se
   >   descubre al validar.
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
8. **⚠️ MEDIDA EN 005 · En COMENTARIOS DE CÓDIGO se cita por NOMBRE, nunca por línea.** Fichero,
   función, símbolo: `internal/config.JuzgarEndpoint`, `Client.Adherir`. **En artefactos fechados
   —`spec.md`, `plan.md`, `tasks.md`, `research.md`— sí se puede citar por línea**, porque son **fotos
   con fecha**: dicen lo que era cierto el día que se escribieron y nadie espera otra cosa. **Un
   comentario en el código no tiene fecha**: se lee como si fuera cierto ahora.
   > **Se dedujo tropezando, y en el peor sitio posible.** La cabecera de `internal/config/endpoint.go`
   > citaba `transport.go:143` como el llamante que envuelve la causa. Al ejecutarse T005 esa línea dejó
   > de ser esa — **dentro del texto de la tarea (T006) que existe precisamente para que las remisiones
   > de la frontera no envejezcan**. La cita no falló por descuido: falló **porque las citas de línea en
   > código fallan por construcción**, y el sitio donde más duele es justo donde más se cuidan.
   >
   > **El nombre sobrevive al refactor y la línea no**, y hay una asimetría que decide: **una cita por
   > nombre que se rompe la caza el compilador o un `grep`; una cita por línea que se rompe no la caza
   > nadie** — sigue apuntando a una línea que existe y que ahora dice otra cosa. **Falla en silencio, y
   > apuntando a algo plausible**, que es la peor forma de fallar.
   >
   > #### ✅ LA REGLA ES MECÁNICA — y el patrón SE DERIVA, no se enumera de memoria
   > Una regla que se comprueba leyendo es una regla que nadie comprueba. **Ésta se comprueba así:**
   >
   > ```sh
   > grep -rnE '\.(go|md|jsonl|json|sh|yaml|yml|tsv|mod|gitignore):[0-9]+|[(`]:[0-9]+' \
   >      --include='*.go' --include='*.sh' .
   > ```
   >
   > **Salida esperada: NADA.**
   >
   > ##### ⛔ LA SEGUNDA RAMA — `\(:[0-9]+` — Y POR QUÉ HIZO FALTA AÑADIRLA
   > La primera rama exige **`fichero.ext:N`**, así que **no ve el idioma `(:131)`**: la cita corta que
   > se escribe cuando el fichero «se sobreentiende» porque es el mismo en el que se está. **Medido en
   > T012**: la cabecera de `Adherir` remitía a `Send` con «(`:131`)» y a `Verify` con «`:91-99`», y el
   > grep daba **cero**. Es la forma **más frágil de todas** —no dice ni de qué fichero habla— y era
   > **invisible para el instrumento**.
   >
   > **`[(`]:[0-9]+` la caza con casi cero falsos positivos**: un paréntesis o un acento grave seguidos
   > de dos puntos y un número no aparecen en Go fuera de este idioma. Van **las dos variantes** porque
   > **las dos estaban escritas**: «(`:131`)» y «`:91-99`».
   >
   > **Y aquí el patrón PARA, medido.** La forma general —«dos puntos y un número tras algo que no es
   > identificador, dentro de un comentario»— se probó y **da falso positivo**: caza `"name":42`, que es
   > **JSON dentro de un comentario**, no una cita. Un grep ruidoso acaba desactivado, así que **se
   > prefiere el par de idiomas concretos a la forma general**. *(Límite conocido: una cita escrita de
   > una tercera manera seguiría siendo invisible. Se anota, no se disimula.)*
   >
   > ##### ⛔ LA LISTA DE EXTENSIONES NO SE INVENTA — SE MIDE, Y ASÍ
   > ```sh
   > git ls-files | grep -oE '\.[A-Za-z0-9]+$' | sort -u
   > ```
   > **El patrón es esa salida.** Y **un tipo de fichero nuevo en el repositorio obliga a rehacerlo**:
   > mientras no se rehaga, el grep sigue dando cero **sin haber mirado los ficheros nuevos**.
   >
   > **Por qué esto y no una lista mejor:** la primera versión enumeró `go|md|json|tsv|yml` **de
   > memoria**, y el conjunto real son **diez** —faltaban `jsonl`, `sh`, `yaml`, `mod`, `gitignore`—.
   > **`.jsonl` es el caso didáctico**: `json` **no casa** con `queue.jsonl:12`, porque tras `json` viene
   > una `l` y no los dos puntos. Una extensión que sea **prefijo de otra** crea un agujero silencioso,
   > y sólo se ve derivando la lista. *(Por eso `jsonl` va antes que `json` en la alternancia: el orden
   > no importa en POSIX ERE, pero sí para quien lea el patrón o lo lleve a otra herramienta.)*
   >
   > ##### ⛔ UN GREP QUE DA CERO CON CITAS VIVAS DENTRO ES PEOR QUE NO TENER GREP: **CERTIFICA**
   > No es que falle: es que **firma que está limpio**. Y nadie vuelve a mirar lo que ya está firmado.
   >
   > **Es la TERCERA vez en esta feature, y las tres el instrumento parecía proteger:**
   >
   > | Instrumento | Daba verde… | …y la fuga pasaba por |
   > |---|---|---|
   > | `TestParseEnrollmentString_Rejects` | `ok` con el endpoint sembrado | comparar **el argumento entero**, no sus campos |
   > | Un golden decodificado a `peticionAdhesion` | pasaría con un campo de más | ver **lo que la struct admite**, no lo que el cuerpo lleva |
   > | Este grep, con la lista a ojo | **cero** | una extensión **que no estaba en la lista** |
   >
   > **El patrón común**: los tres miran **el molde que ya tenían** en vez de **lo que hay**. La defensa
   > es la misma en los tres: **derivar el criterio del artefacto real** —los campos decodificados, las
   > claves del cuerpo recibido, las extensiones del repositorio— y **no de lo que se recuerda de él**.
   >
   > **Por qué está ajustado así, lo demás:**
   > - **Cubre `.md` y no sólo `.go`.** El defecto no es citar código por línea, es **citar por línea
   >   desde un comentario**, que no tiene fecha. Y midiendo resultó que **la mayoría de las citas
   >   apuntaban a artefactos**, no a código.
   > - **`--include` acota a CÓDIGO —`.go` y `.sh`—, y el ámbito es el ÁRBOL ENTERO.** Los artefactos
   >   fechados —`spec.md`, `plan.md`, `tasks.md`— **sí pueden citar por línea**: son fotos con fecha, y
   >   **medido, ahí viven las 205 citas del repositorio**, todas legítimas. Lo que se acota es **qué
   >   ficheros se registran**, no **qué directorios**: limitarlo a `internal/` y `cmd/` era otra lista
   >   a ojo.
   > - **No caza URLs con puerto** (`https://ejemplo.test:8443/guia.md` no casa: `.test` no es una
   >   extensión del repositorio), **y sí caza los rangos** (`x.go:41-43`), que es como estaban escritas
   >   tres de las que había.
   > - **Única excepción conocida**: una salida de herramienta **transcrita literalmente** (p. ej.
   >   `endpoint.go:49:2: u declared and not used`) es un registro, no una remisión. Hoy no hay ninguna
   >   en código; si se añade, se anota aquí y no se cambia el grep por un caso.
   >
   > ##### 📏 MEDIDO al rehacer el patrón — el agujero era real, la fuga no
   > Con el patrón derivado y el árbol entero: **cero citas nuevas**. Barrido completo del repositorio:
   > **205 citas por línea, TODAS dentro de `.md`** —artefactos fechados, legítimas—, **cero en `.go`,
   > cero en `.sh`**, y **cero a `.jsonl` o `.tsv` en ningún sitio**: esos fixtures se citan **por
   > nombre** (`testdata/claude_code_sample.jsonl`, `baseline-sc004.tsv`), nunca por línea.
   > **El agujero del patrón era real —`.jsonl:12` no habría casado— pero detrás no había ninguna cita
   > viva.** Se deja anotado tal cual: **la ausencia de hallazgo no valida el instrumento que la
   > encontró**, y por eso lo que se arregla es **la derivación**, no el resultado.
   >
   > #### ⚠️ EL BARRIDO CRECIÓ AL HACERLO, Y ESO ES EL ARGUMENTO
   > Primer barrido, sólo `\.go:` → **seis**. Se arreglaron las dos de 005, incluida **la peor**:
   > `transport.go:143` **dentro de un mensaje de fallo**, o sea, texto que un humano lee justo cuando
   > algo se ha roto, apuntando a una línea equivocada.
   >
   > **Al ampliar el grep a `.md` aparecieron CINCO MÁS que la primera versión no veía**, y **tres ya
   > estaban muertas** —medido, no supuesto—:
   >
   > | Cita | Qué hay hoy ahí |
   > |---|---|
   > | `tasks.md:132` (×2, desde `resolve_test.go` y `main_test.go`) | **línea vacía** en 004; en 005, la nota de ejecución de git. **Y no dice de qué `tasks.md` habla** |
   > | `tasks.md:150-156` | en 004, el *Goal* de US3 — no la receta de sandbox que promete |
   > | `spec.md:249` | **acertada, por suerte**: sigue cayendo en P-004 FR-010 |
   > | `contracts/transport.md:58` | **acertada**, pero el fichero vive en **001**, y la cita no lleva ruta |
   >
   > **Las nueve pasan a nombre** —fichero + símbolo, o fichero + § con título—, y los anclajes se
   > verificaron uno a uno antes de escribirlos: `T018 §«Por qué aquí SÍ se permite el t.Skip que T010
   > prohíbe»`, `§«Receta común de T023/T024»`, `§Notas de prueba`.
   >
   > **Por qué se arreglaron también las cuatro anteriores a 005**, que estaban fuera de alcance: **la
   > disciplina 8 es un artefacto de esta feature, y una regla que se publica con violaciones conocidas
   > no es una regla, es una recomendación.** Son ediciones de comentario —riesgo cero— y `resolve.go`
   > lo vigila T029 igualmente.
   >
   > **Lo que SÍ se queda fuera: `README.md:74`.** No es una cita por línea desde código: es que **el
   > README documenta un campo retirado** (el `project_ref_mode` que P-004 quitó). **Es deuda de
   > contenido anterior, y ninguna regla de 005 la toca** — está al backlog por su propia vía.
9. **⚠️ MEDIDA EN 005 · Los identificadores de requisito se citan CON PREFIJO DE SPEC.** `P-004
   FR-017`, `P-005 FR-017` — **nunca `FR-017` a secas**. Es la convención que ya usa la spec, y aquí se
   volvió obligatoria al medirla: **`P-004 FR-017` es el ALCANCE de la frontera** y **`P-005 FR-017` es
   el TRANSPORTE SEGURO** de la adhesión. **Números iguales, garantías distintas, y los dos se citaban
   en la misma cabecera** (`internal/ingest/boundary_test.go`). Barrido y cerrado en los cuatro ficheros
   que lo mencionaban; verificable con `grep -rn 'FR-017' --include='*.go' . | grep -v 'P-00[0-9] '` →
   **cero**.
   > #### ⚠️ MEDIDO EN PHASE 7 · EL GREP SÓLO MIRA `FR-017`, Y LA REGLA ES GENERAL
   > La regla dice **«los identificadores de requisito»**, en plural y sin excepción; el instrumento
   > que la acompaña **sólo busca uno**. Generalizado a `FR-0[0-9][0-9]` sobre `cmd/` aparecieron
   > **tres citas sin prefijo**, dos escritas en 005. **Las tres corregidas.**
   >
   > **Y una de ellas enseña algo que el instrumento no ve: EL SALTO DE LÍNEA.** `(P-005\n// FR-023,
   > …)` **está bien citada** y el grep la marca igual, porque el prefijo quedó en la línea anterior.
   > Al revés también pasa: una cita mal escrita puede esconderse de un grep de una línea. **La
   > defensa es de estilo, no de patrón**: se reflowea el comentario para que **prefijo y número
   > vivan en la misma línea**, y así lo que el grep ve coincide con lo que hay escrito.
   >
   > **El patrón general —`grep -rn 'FR-0[0-9][0-9]' --include='*.go' . | grep -v 'P-00[0-9] '`— NO
   > se adopta todavía**: sobre el árbol entero da **82 líneas**, casi todas de 001–004, donde la
   > feature se sobreentiende por el paquete. Adoptarlo hoy sería publicar un instrumento que grita
   > siempre, y **un grep ruidoso acaba desactivado** —es la lección que la disciplina 8 ya pagó—.
   > **Queda al backlog como barrido propio**, y mientras tanto la regla se aplica a lo que 005
   > escribe, que es lo que 005 puede sostener.

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

- [x] **T005** Sustituir **las CUATRO réplicas** por llamadas a T004:
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
  > ### ✅ MEDIDO — T005 ejecutada: cuatro réplicas sustituidas, los cuatro llamantes se movieron
  > **PASO 0 — la quinta red.** Barrido de `internal/config` y `cmd/permea`: **la hay**,
  > `internal/config/enrollment_test.go:116` (`TestParseEnrollmentString_Rejects`, casos «endpoint http»
  > y «endpoint vacío»). **Pero al medirla NO cubre el peligro de T005**: compara
  > `strings.Contains(msg, tc.in)` —**el argumento entero**— y el token entero, mientras que un `%w` del
  > `*url.Error` filtra **el endpoint**, que no es subcadena de ninguno de los dos. **Medido con la
  > mutación**: con el endpoint reproducido en el mensaje, ese test **pasa en verde** (`ok`).
  > #### ⛔ POR QUÉ `enrollment_higiene_test.go` NO ES UN DUPLICADO — léase antes de borrarlo
  > **A primera vista solapa con `enrollment_test.go:116`, y no solapa.** Aquel compara
  > `strings.Contains(msg, tc.in)`: **el argumento ENTERO**. Y **ninguno de los tres campos
  > decodificados es subcadena de `pmea2.<base64>`** — endpoint, token y `dev_id` viajan **codificados
  > dentro**, así que un mensaje que reprodujera cualquiera de los tres **pasa su comprobación sin
  > despeinarse**.
  >
  > **No es deducción, está medido**: con el endpoint sembrado en el mensaje de error,
  > `TestParseEnrollmentString_Rejects` **da `ok`**. La fuga a la vista y la red en verde.
  >
  > **Y el agujero es de FAMILIA, no del endpoint**: el endpoint fue sólo el campo que T005 puso en
  > peligro. Si mañana un mensaje reprodujera el `dev_tok_…` o el identificador de máquina, **tampoco
  > lo cazaría nadie**. Por eso el fichero nuevo recorre **los tres campos** con el umbral del contrato
  > (SC-005: **subcadenas de ocho**), en **un bucle sobre una tabla** y no en tres tests copiados.
  >
  > **Los dos ficheros se quedan, y cubren cosas distintas**: uno el argumento entero, el otro sus
  > partes. Borrar el nuevo por «duplicado» devuelve el agujero entero.
  >
  > Va en **fichero aparte** para no tocar el existente (condición de parada). Nació **verde**; mutación
  > válida (`go build` pasa) y la **tumba**:
  > ```
  > enrollment_higiene_test.go:94: el error reproduce un fragmento del ENDPOINT incrustado (≥8 caracteres): "http://i"
  >       mensaje: "enrollment string inválido: el endpoint \"http://inseguro.example/ingest\" debe ser https"
  > ```
  > #### ✅ MEDIDO — TRES CAMPOS, TRES MUTACIONES INDEPENDIENTES
  > Las tres se siembran en **el mismo mensaje** (la rama de charset de `dev_id`) para que sólo cambie
  > **el campo** y las tres sean comparables. Las tres **compilan** (`go build ./...` pasa) y **cada una
  > tumba UNA SOLA aserción, la de su campo**, con las otras dos en pie:
  > ```
  > A endpoint  enrollment_higiene_test.go:114: … fragmento del campo endpoint (≥8 caracteres): "https://"
  > B token     enrollment_higiene_test.go:114: … fragmento del campo token    (≥8 caracteres): "dev_tok_"
  > C dev_id    enrollment_higiene_test.go:114: … fragmento del campo dev_id   (≥8 caracteres): "maq!Z7Q3"
  > ```
  > **Ninguna tumbó las tres**, que era la condición: si una sola las hubiera tumbado, el bucle estaría
  > comprobando el argumento entero otra vez y no los campos.
  >
  > ### ⚠️ TRAMPA MEDIDA — un test de ausencia por subcadenas es sensible AL VALOR ELEGIDO
  > El primer `dev_id` de la tabla fue `"maquina no permitida!01"` y **dio un ROJO FALSO al nacer**: el
  > mensaje genérico dice «dev_id con caracteres **no permitidos**», y las dos cadenas comparten
  > `" no perm"`. **No había fuga: había vocabulario común.** Los valores de prueba tienen que parecerse
  > a lo que son —identificadores y secretos, alta entropía— y **no a la prosa de los mensajes**, o el
  > test grita sin motivo y acaba desactivado por ruidoso. **Aplica a T024**, que usa esta misma técnica
  > sobre los ocho desenlaces del comando.
  >
  > Las tres revertidas por edición inversa; `enrollment.go` verificado idéntico.
  >
  > #### ✅ AMPLIADO DESPUÉS — LAS NUEVE RAMAS, no cuatro
  > La familia se cerró **por campo**; faltaba el otro eje, **por rama**. Medido con `-coverprofile`:
  > `ParseEnrollmentString` tiene **NUEVE desenlaces de error** y la tabla visitaba **cuatro**. Ahora
  > son **diez filas para nueve ramas** (la 5 lleva dos: no analizable y en claro), **nueve de nueve
  > cubiertas**, verificado otra vez por cobertura y no leyendo.
  >
  > **Los dos huecos que había, y no eran iguales:**
  > - **Rama 4 «json ilegible»** — el hueco de verdad. Con `unknown field`, `DisallowUnknownFields`
  >   falla **pero la struct ya quedó COMPLETA**: el token está vivo en `p`. **Hoy no filtra, y lo único
  >   que separa «no filtra» de «filtra» es que alguien añada contexto al mensaje mientras depura.**
  > - **Rama 7 «dev_id ausente»** — endpoint y token poblados, sin testigo.
  > - **Ramas 1, 2 y 3** disparan antes de decodificar: la aserción es **trivialmente cierta** hoy. **Se
  >   incluyen igual**, y está escrito en el fichero: «seis de nueve» no se verifica de un vistazo y
  >   «nueve de nueve» sí, y **si alguien mueve la decodificación más arriba dejan de ser triviales sin
  >   que nadie lo note**.
  >
  > **Mutación** —una sola, en la rama 4, que es la que tiene la struct poblada—: sembrar el token como
  > «contexto para depurar», que es exactamente como pasa en la vida real. `go build` **PASA** y la
  > **tumba**:
  > ```
  > enrollment_higiene_test.go:169: rama 4: el error reproduce un fragmento del campo token (≥8 caracteres): "dev_tok_"
  >       mensaje: "enrollment string inválido: json ilegible (token=dev_tok_9HpQ3mZv7KxR2wLb)"
  > ```
  > **Mató sólo ese hecho**: **9 de las 10 filas quedaron en pie**. Revertida por edición inversa;
  > `enrollment.go` verificado idéntico.
  >
  > **PASO 1 — las cuatro sustituciones.** `enrollment.go` **funde** los dos hechos y **descarta la
  > causa** (su argumento lleva el token); `config.go` y `Send` conservan **sus dos mensajes**; `Send`
  > conserva además **causa (`%w`) y centinela**; `Adherir` pasa a **`ErrScheme` + causa conservada** y
  > **`errEsquemaAndamiaje` desaparece del fichero**. **Efecto colateral medido**: `enrollment.go` y
  > `config.go` dejan de usar `net/url` y pierden el import. `transport.go` gana el import de `config`;
  > **sin ciclo**, y no es nuevo: `transport/queue.go:11` ya lo importaba.
  > #### ✅ LA PRUEBA ESTRUCTURAL QUE SALIÓ SOLA — el grafo de imports
  > El import de `net/url` **no se retiró por limpieza: se retiró porque el compilador ya no lo
  > aceptaba.** `enrollment.go` y `config.go` dejaron de usarlo.
  >
  > **Y eso es exactamente lo que ningún test puede ver.** T011 y los demás observan **comportamiento**;
  > dónde vive el juicio es **estructura**. Una réplica disfrazada —una copia local que devolviera lo
  > mismo— **habría conservado el import**, porque seguiría llamando a `url.Parse`. **El import que
  > desaparece es la firma de que la réplica se fue de verdad.**
  >
  > **Es el COMPLEMENTO de la mutación del PASO 2, no un adorno**: la mutación demuestra que los cuatro
  > llamantes **dependen** del cuerpo único; el grafo de imports demuestra que **ya no conservan el
  > suyo**. Juntas cierran las dos mitades —«todos usan el original» y «nadie guarda una copia»—, y
  > **ninguna de las dos sola lo hace**.
  >
  > ⚠️ **Escrito aquí para T006**, que documenta precisamente la unificación y su encontrabilidad: éste
  > es el argumento verificable de que ocurrió, y **se comprueba con `go build`, no leyendo el código**.
  >
  > **Y la retirada del andamiaje se verificó igual, por barrido**: `grep -rn errEsquemaAndamiaje
  > internal/ cmd/` → **cero**. Las tres apariciones que quedaban eran **comentarios de
  > `adhesion_test.go` en presente** —«hoy la guarda devuelve…»— que tras T005 **decían algo falso**; se
  > pasaron a **registro histórico** y se les quitó el identificador muerto, porque un nombre que ya no
  > existe citado en presente se lee como código vivo.
  >
  > **PASO 2 — la prueba positiva.** `go build ./...` **PASA** con la mutación puesta (a). Los **cuatro
  > llamantes cayeron, con cuatro mensajes DISTINTOS** (b):
  > ```
  > 1 enrollment.go  enrollment_test.go:109: esperaba error, got ("http://inseguro.example/ingest", "dev_tok_…", "acme-dev-01")
  > 2 config.go      config_test.go:63: endpoint http:// debe rechazarse (FR-009)
  > 3 Send           transport_test.go:157: esperaba rechazo de http://, got transport: error de red: Post "http://insecure.example/ingest": dial tcp: lookup …
  > 4 Adherir        adhesion_test.go:67: errors.Is(err, ErrScheme) = false, err = transport: adhesión no implementada
  > ```
  > **Los cuatro hablan con su propia voz** —uno acepta la terna, otro nombra FR-009, otro **llega a
  > intentar la emisión por la red**, el cuarto atraviesa la guarda hasta el `return` de andamiaje—, así
  > que **no es «se rompió algo común»**. El **caso 3 de T011 NO cayó**, y es correcto: la mutación
  > altera **sólo el hecho del esquema**, y ese test mira **la causa del parseo** (disciplina 3, «mata
  > sólo el hecho que altera»). Revertida por edición inversa; `endpoint.go` idéntico.
  >
  > **PASO 3 — falsabilidad del caso 2.** Con la rama «no analizable» de `Adherir` devolviendo
  > `ErrScheme`, la aserción de sentido **CAE**, y en su línea exacta:
  > ```
  > adhesion_test.go:148: errors.Is(err, ErrScheme) = true para "https://ejemplo\x7f.test/…": el esquema NO se pudo leer, así que no se puede juzgar no admisible
  > ```
  > **Ya no pasa por accidente.** Y el **caso 3 también la caza** (`errors.As` → false): la regresión de
  > refundir los dos hechos tiene ahora **dos redes independientes**. Revertida por edición inversa.
- [x] **T006** Remisión cruzada en los dos sitios (D-005-P2 §encontrabilidad): en
  `internal/transport/transport.go` —donde estaba la condición— un comentario que dice **dónde vive
  ahora el juicio y por qué se movió**; en `internal/config/` —donde vive— **quiénes son sus
  llamantes**. Sin las dos, unificar mejora el código y **empeora el hallazgo de la frontera**
  (Principio III).
  ⚠️ **Son CUATRO, y los cuatro existen ya** —`enrollment.go`, `config.go`, `transport.go` (`Send`) y
  `transport.go` (`Adherir`)—, porque **T002 creó el cuarto en Phase 1**. La lista se escribe con los
  cuatro, sin ningún «pendiente».
  *(Esta nota decía «hoy son tres, el cuarto llega en T012». **Dejó de ser cierto al ejecutarse T002**:
  el llamante existe desde Phase 1, y quien lo busque lo encuentra.)*
  > ### ✅ MEDIDO — T006 escrita en los dos sitios, y en un tercero que salió al medir
  > **(a) Donde estaba la condición** — `internal/transport/transport.go`, en **`Send`**: nota larga con
  > **dónde vive ahora** (`internal/config.JuzgarEndpoint`), **por qué se movió** (cuatro copias de la
  > frontera = cuatro sitios donde se arreglan tres y se olvida una, y el fallo es **emitir en claro**),
  > y **qué NO se movió**: el desenlace. Esta puerta conserva causa y centinela; `enrollment.go` los
  > descarta. **Se unificó el juicio, no la presentación.**
  >
  > **(a-bis) Y en `Adherir` también**, remisión breve que apunta a la nota de `Send`. No estaba pedido
  > y hacía falta: **`Adherir` es el llamante que no existía cuando se escribió el plan**, y quien
  > entre por la segunda puerta buscando la frontera entra por ahí, no por `Send`.
  >
  > **(b) Donde vive** — `internal/config/endpoint.go`, junto a `JuzgarEndpoint`: **los CUATRO
  > llamantes, ninguno pendiente**, cada uno con su desenlace anotado (quién funde, quién separa, quién
  > conserva la causa y quién la descarta).
  >
  > **(b-bis) La prueba estructural, escrita ahí**: la mutación de T005 demuestra que **los cuatro
  > dependen** del cuerpo único; **el import de `net/url` que desapareció** de `enrollment.go` y
  > `config.go` demuestra que **ninguno guarda copia** —se fue porque **el compilador dejó de
  > aceptarlo**, no por limpieza; una réplica disfrazada lo habría conservado—. **Ninguna de las dos
  > sola basta**, y la segunda **se verifica con `go build`, no leyendo el código**.
  >
  > ### ⚠️ DOS CITAS ENVEJECIDAS, corregidas al escribir esto
  > El encabezado de `endpoint.go` seguía diciendo **«las réplicas las sustituye T005, no esta tarea»**
  > —falso desde que T005 corrió— y citaba **`transport.go:143`**, línea que ya no es esa. La segunda se
  > sustituyó por el **nombre de los llamantes, sin número**: *las citas de línea envejecen, y ésta ya
  > envejeció una vez*. Es el mismo defecto que T006 existe para prevenir, en el propio texto de T006.

**Checkpoint**: **9 ok** · los cuatro tests de la red **sin una línea tocada** · **y T011 EN VERDE**.

> **Las dos mitades del checkpoint dicen cosas distintas, y hacen falta las dos.** «Los cuatro
> intactos» es el criterio **negativo**: nada se rompió. **«T011 en verde» es el positivo**: la
> unificación **llegó al método nuevo**. Sin él, una T005 que extrajera la función y **se olvidara de
> `Adherir`** pasaría el checkpoint entero — porque «nada se rompió» es compatible con «no pasó nada».

---

## Phase 3: LA FRONTERA — el golden test va PRIMERO (Principio IV)

> **Disciplina de primer commit.** Esta feature **abre la segunda puerta de la frontera de datos**, y
> el golden existente cubre la emisión de eventos, no la adhesión.

- [x] **T007** Crear el **golden test de frontera de la adhesión** en
  `internal/transport/boundary_adhesion_test.go` (nuevo): el cuerpo de la petición contiene
  **exactamente `{code, project_ref}` y nada más** — allowlist de dos elementos—, y **ningún otro dato
  de la instalación** viaja en él. **Nace verde** (T002 ya construye el cuerpo) → **se valida por
  mutación**: añadir un tercer campo al cuerpo debe **tumbarlo**, y el mensaje debe leerse.
  > ### ✅ MEDIDO — el golden nació verde y la mutación lo tumbó
  > `internal/transport/boundary_adhesion_test.go` ·
  > `TestFronteraAdhesion_ElCuerpoLlevaExactamenteDosClaves`. Captura **el cuerpo REAL recibido por el
  > servidor** (`httptest.NewTLSServer` + `srv.Client()`), no lo que el test cree que envió, y lo
  > decodifica a **`map[string]any` y NO a `peticionAdhesion`**: decodificar a la struct del propio
  > código haría que el test viera **lo que la struct admite** en vez de **lo que el cuerpo lleva**, y
  > un campo de más sería invisible.
  >
  > **Mutación** — tercer campo (`hostname`, un dato de la instalación). `go build ./...` **PASA**
  > (mutación válida) y **la tumba**:
  > ```
  > boundary_adhesion_test.go:99: el cuerpo de adhesión NO lleva exactamente la allowlist de dos elementos:
  >       claves emitidas: [code hostname project_ref]
  >       allowlist:       [code project_ref]
  >       Principio I: ningún dato de la instalación viaja por esta puerta salvo estos dos
  > ```
  > **Mató sólo ese hecho**: ningún otro test del paquete se movió. Revertida por edición inversa;
  > `transport.go` verificado idéntico.
  >
  > ### ✅ MEDIDO — LA NOTA RECÍPROCA (D-005-P14), en los dos ficheros
  > La cabecera de `internal/ingest/boundary_test.go` afirmaba que FR-017 **«la nombra entera»**
  > nombrando **tres caminos**. **Desde P-005 es falso**: la adhesión emite por un camino propio que no
  > es ninguno de los tres. Nota puesta **en los dos ficheros, cada uno nombrando al otro**.
  >
  > **Y al escribirla salió una colisión, que después se cerró en todo el código** (disciplina 9): el
  > **`P-004 FR-017`** (alcance de la frontera) **no es** el **`P-005 FR-017`** (transporte seguro de la
  > adhesión) — **mismo número, garantías distintas, y los dos citados en la misma cabecera**. Barrido
  > completo: **nueve apariciones sueltas en tres ficheros**, todas prefijadas, más la convención escrita
  > en **los dos ficheros de frontera**. Verificable: `grep -rn 'FR-017' --include='*.go' . | grep -v
  > 'P-00[0-9] '` → **cero**.
  ⚠️ **En ESTA MISMA TAREA**, la nota en `internal/ingest/boundary_test.go` (D-005-P14): su cabecera
  `:106-107` afirma que FR-017 *«la nombra entera»* nombrando tres caminos, y **esa frase deja de ser
  cierta** en cuanto exista una segunda puerta. Nota en los dos ficheros, **y no suelta**: separarla
  es cómo se queda sin hacer.

**Checkpoint**: la frontera tiene **dos testigos**, y cada uno sabe del otro.

---

## Phase 4: El transporte — los cuatro desenlaces (D-005-P1)

### Tests (rojo antes de verde)

- [x] **T008** [P] Test del **desenlace de éxito** en `internal/transport/adhesion_test.go` —
  **el fichero ya existe: lo creó T011 en Phase 2** al mudarse allí—
  **Garantía**: `adhesion.md` desenlaces 3 y 4. `200` con `{"project":{"name":…}}` → devuelve **la
  denominación**. Nace en **rojo**: T002 devuelve siempre «adhesión no implementada». Transcribir la
  razón.
- [x] **T009** [P] Test de **los dos rechazos, distintos entre sí** en el mismo fichero —
  **Garantía**: desenlaces 1 y 2. `422` → rechazo; `409` → conflicto; **y son ramas distintas**. Nace
  en **rojo**.
  ⚠️ **Ramificar por el ESTADO**, no por el cuerpo (`adhesion.md` §Qué distingue a qué). El cuerpo se
  comprueba como confirmación, no como discriminante.
- [x] **T010** [P] Test de **respuesta ininterpretable** en el mismo fichero — **Garantía**: FR-002 +
  FR-013. Un `200` **sin `project.name` legible** → **no verificable**, **NUNCA éxito**. Nace en
  **rojo** porque **T002 devuelve «adhesión no implementada», que es OTRO centinela**. *Es el riesgo
  declarado de D-005-P1: un éxito sin nombre no es un éxito.*
  ⚠️ **Este es el rojo más frágil del fichero** y por eso T002 lo protege explícitamente: si el
  andamiaje devolviera «no verificable», **este test nacería verde acertando contra el stub**.
  > ### 🔴 MEDIDO — los TRES nacieron rojos, y con el mensaje previsto
  > **T008** `TestAdherir_ExitoDevuelveLaDenominacion`:
  > ```
  > adhesion_test.go:291: Adherir con 200 y denominación legible devolvió err = transport: adhesión no implementada; se esperaba éxito
  > ```
  > **T009** `TestAdherir_LosDosRechazos` (6 subtests, los 6 en rojo) y
  > `TestAdherir_LosDosRechazosSonDistinguibles`:
  > ```
  > adhesion_test.go:348: estado 422: errors.Is(err, transport: el código de adhesión no es utilizable) = false; err = transport: adhesión no implementada
  > adhesion_test.go:348: estado 409: errors.Is(err, transport: esta identidad ya pertenece a otro proyecto) = false; err = transport: adhesión no implementada
  > adhesion_test.go:386: los desenlaces 1 y 2 son INDISTINGUIBLES: errors.Is(·, ErrCodigoNoUtilizable) contesta false a los dos
  >       422 → transport: adhesión no implementada
  >       409 → transport: adhesión no implementada
  > ```
  > **T010** `TestAdherir_DoscientosSinNombreLegibleEsNoVerificable` (**13 formas, 13 en rojo**):
  > ```
  > adhesion_test.go:448: 200 con cuerpo "{\"project\":{}}": errors.Is(err, ErrNoVerificable) = false; err = transport: adhesión no implementada
  > ```
  >
  > ### ⚠️ PARA QUE LOS TRES PUDIERAN NACER ROJOS HUBO QUE DECLARAR TRES CENTINELAS
  > `ErrCodigoNoUtilizable`, `ErrIdentidadYaAsignada` y `ErrNoVerificable`, en `transport.go`,
  > **declarados y SIN CABLEAR**: `Adherir` sigue devolviendo `ErrAdhesionNoImplementada` para todo, y
  > **ninguno de los tres se devuelve todavía**. Los cablea **T012**.
  >
  > **No es adelantar T012, es la condición para que estos tests existan.** Un test que no compila da
  > `[build failed]`, y eso **no es un rojo legible**: falla igual con un test correcto y con uno vacío
  > (disciplina 3). Es **exactamente el criterio que ya aplicó T002** con `ErrAdhesionNoImplementada`:
  > el andamiaje declara lo que los tests necesitan y **devuelve algo que ninguno espera**.
  >
  > ⛔ **Y T010 no existiría sin esto, que es el punto.** Sin `ErrNoVerificable` su única aserción
  > posible sería «hay error y no hay denominación» — **y el andamiaje ya lo cumple**, así que
  > **habría nacido VERDE**. Medido en el rojo real: sus aserciones (1) «nunca éxito» y (2) «ninguna
  > denominación» **pasan hoy**; lo único que falla es **el centinela**. Era el rojo más frágil del
  > fichero y **sólo el centinela distinto lo sostiene**.
  >
  > ### ✅ DOS DECISIONES DE T009 QUE VAN MÁS ALLÁ DEL ENUNCIADO
  > 1. **«Ramificar por el estado» se comprueba CRUZANDO LOS CUERPOS.** Dos casos rectos —`422` con su
  >    cuerpo, `409` con el suyo— **los pasa igual una implementación que ramifique por el cuerpo**. Así
  >    que la tabla lleva `422` con el cuerpo del `409`, `409` con el del `422`, `422` sin cuerpo y
  >    `409` con cuerpo ilegible. **Sólo así el estado queda demostrado como discriminante** y el cuerpo
  >    como confirmación.
  > 2. **La distinguibilidad va en test APARTE**, y compara **las respuestas de los dos centinelas a los
  >    dos errores**. Un error que envolviera ambos pasaría las aserciones individuales; esto no.
  >
  > ### ✅ T010 — se ENUMERAN 13 formas, y la razón es el decodificador
  > clave ausente · objeto `project` ausente · `null` · cadena vacía · sólo espacios · numérico ·
  > objeto · lista · `project` no objeto · `project` null · cuerpo vacío · cuerpo no JSON · JSON que no
  > es objeto. **Con `encoding/json` y una struct de campos, las tres primeras producen el mismo cero
  > silencioso**, y el tipo equivocado da un error fácil de tratar como «campo vacío». **Probar una sola
  > forma deja las otras abiertas.**
  >
  > **Y los cuerpos se escriben como JSON literal, nunca serializando una struct de este repositorio**
  > (disciplina de las tres veces): serializarla haría que el test viera **lo que ese tipo admite** en
  > vez de **lo que el servidor mandó** — y T010 existe precisamente para los cuerpos que una struct
  > decodifica sin protestar.

### Implementación

- [x] **T012** Implementar el método de adhesión en `internal/transport/transport.go`: llama a la
  guarda (T004), compone la petición, **lee y decodifica** la respuesta y **distingue por estado**.
  Reutiliza el `http.Client`, su timeout y la cabecera de autenticación. **Un solo intento, sin cola**
  (D-005-P4, siguiendo `Verify()` en `transport.go:91-99`). Pone en verde **T008, T009 y T010**.
  ⚠️ **Ya NO pone en verde T011**: esa la puso **T005**, en Phase 2, al sustituir la réplica de
  andamiaje de `Adherir` por la guarda unificada. Cuando T012 llega, **T011 lleva dos fases en verde**.
  ⚠️ **T012 ya NO añade ningún llamante a la lista de T006**: el cuarto lo creó **T002** en Phase 1 y
  lo reconvirtió **T005**. Aquí solo se implementan **los desenlaces** —leer el cuerpo, distinguir por
  estado—; **la guarda ya está puesta y ya es la unificada**. *(Esta tarea llevaba el deber de
  «añadirse a la lista»; se retiró al ejecutarse T002, que adelantó el llamante a Phase 1.)*
- [x] **T010-E** [P] Test del **estado que el contrato NO enumera** en `internal/transport/adhesion_test.go`
  — **Garantía**: Principio I. `500`, `403`, `404` y `204` → **no verificable**, y **NUNCA éxito**.
  **Nació rojo**, y se observó **antes de escribir una línea de T012**.
  > ### 📏 MEDIDO — QUÉ DICE EL CONTRATO DE UN ESTADO FUERA DE LOS TRES
  > **No enumera ninguno** —`adhesion.md` sólo nombra `200`, `422` y `409`— **pero NO calla**: su
  > cláusula general dice *«si el desenlace no puede establecerse —servidor inalcanzable, **o respuesta
  > que no permite determinarlo**—, el cliente informa de que no se pudo completar y **NUNCA afirma
  > ningún desenlace** (FR-013)»*, y `cli.md` repite la fórmula en su fila **NV**. **Un `500` es
  > exactamente eso**, así que el comportamiento está cubierto.
  >
  > ### ⚠️ HUECO DEL CONTRATO — AL BACKLOG, no se arregla aquí
  > **`adhesion.md` §Qué distingue a qué resume el reparto como «`200` frente a `4xx`»**, y esa fila
  > **se lee como exhaustiva**. Quien implemente desde ella escribe `if 200 {éxito} else {rechazo}` — y
  > **convierte un `500` en un rechazo afirmado**, que es precisamente lo que FR-013 prohíbe. **La
  > cláusula general lo cubre; la tabla invita a lo contrario**, y la tabla es lo que se lee primero.
  >
  > **No se toca aquí**: `adhesion.md` es artefacto de la Phase 1 del plan, y además está **definido en
  > dos sitios** hasta que se cierre la ventana de A′ — corregirlo es otra conversación. *(Lo que sí
  > queda cerrado es el comportamiento: el test existe y `Adherir` lo cumple.)*
  >
  > ### 🔴 El rojo, transcrito
  > ```
  > adhesion_test.go:512: estado 500: errors.Is(err, ErrNoVerificable) = false; err = transport: adhesión no implementada
  > adhesion_test.go:512: estado 204: errors.Is(err, ErrNoVerificable) = false; err = transport: adhesión no implementada
  > ```
  > **`204` es el caso que justifica el test entero**: es `2xx`, así que un `if estado/100 == 2` lo
  > daría por unión conseguida. Los cuatro casos comprueban además que **no** casan con ninguno de los
  > dos centinelas de rechazo: afirmar un rechazo también es afirmar un desenlace.
  > ### ✅ MEDIDO — T012 ejecutada: los cuatro rojos en verde y el andamiaje retirado
  > **La implementación**: guarda unificada → cuerpo → **un solo intento, sin cola** (`HTTP.Do`, sin
  > `sendWithRetry` ni `Append`/`Drain`) → lectura acotada del cuerpo → **`switch` por ESTADO** con
  > `default` explícito a `ErrNoVerificable`. Reutiliza `http.Client`, su timeout y la cabecera de
  > autenticación. **Ningún llamante nuevo en la lista de T006**, como estaba previsto.
  >
  > **La lectura del nombre lleva DOS comprobaciones, y cubren cosas distintas**: el error de
  > `Unmarshal` caza **los tipos equivocados** (`"name":42`, `"name":{…}`, `"project":"…"`, cuerpo no
  > JSON); la comprobación de vacío caza **lo que decodifica sin protestar** —clave ausente, `null`,
  > `project` ausente, cadena vacía—, que **producen los cuatro el mismo cero silencioso**. Sin la
  > segunda, esos cuatro devolverían `""` **como si fuera una denominación**.
  >
  > ### ✅ PASO 2 — las dos aserciones «verdes dentro del rojo» de T010 SON FALSABLES
  > Hicieron falta **dos mutaciones, y el motivo es del propio test**: la aserción (1) es un
  > `t.Fatalf`, así que **corta el subtest** y la (2) no llega a evaluarse. Una sola mutación no podía
  > tumbar las dos.
  >
  > **Mutación A** —se retira la comprobación de vacío: un `200` sin nombre pasa por éxito—. Compila, y
  > cae **la aserción (1)**:
  > ```
  > adhesion_test.go:437: 200 con cuerpo "{\"project\":{}}" devolvió err = nil y denominación "": un éxito sin nombre NO es un éxito (P-005 FR-002)
  > ```
  > **Mutación B** —error correcto pero denominación inventada, que **aísla la (2)**—. Compila, y cae:
  > ```
  > adhesion_test.go:442: 200 con cuerpo "{\"project\":{}}" devolvió denominación "(sin nombre)": no había nombre que leer
  > ```
  > **Las dos mataron sólo su hecho**: con A cayeron **6 de 13** —las seis que decodifican en silencio;
  > las otras siete siguen fallando en `Unmarshal`, que es otro hecho—, y con B **ningún otro test del
  > paquete se movió**. Las dos revertidas por edición inversa; `transport.go` verificado idéntico.
  >
  > ### ✅ PASO 3 — andamiaje retirado
  > `grep -rn ErrAdhesionNoImplementada internal/ cmd/` → **cero**. Con él se fueron su declaración y
  > su bloque de motivos. **Los comentarios que lo citaban pasaron a pasado y sin el identificador
  > muerto** —en `transport.go` y en `adhesion_test.go`—: *un comentario en presente que nombra un
  > símbolo retirado no es historia, es una afirmación falsa.*
  >
  > **Y de paso cayó una cita por línea que la disciplina 8 NO caza**: la cabecera de `Adherir` decía
  > que `Send` descarta el cuerpo «(`:131`)» y que `Verify` está en «`:91-99`». **Sin nombre de fichero
  > delante, el grep no las ve** — el patrón exige `fichero.ext:N`. Pasadas a nombre (`Client.Verify`,
  > `Send`). **Queda anotado como límite conocido del grep**, no como excepción: una cita `:NNN` a secas
  > es igual de frágil y **es invisible para el instrumento**.

---

## Phase 5: El destino derivado (D-005-P3)

- [x] **T013** [P] Test de **derivación correcta** en `internal/config/config_test.go` — **Garantía**:
  `adhesion.md` §Cómo se obtiene `<base>`. Conserva **esquema, host, puerto y prefijo**, y sustituye
  solo el último segmento. **El puerto no estándar es caso obligatorio**: el banco local usa `:8443`.
  Nace en **rojo**.
- [x] **T014** [P] Test de **forma inesperada** en el mismo fichero — **Garantía**: FR-009 **y
  FR-020**. Un endpoint cuya ruta no termina en el segmento conocido → **rehúsa**, nombrando **la
  forma** de lo hallado, y **el mensaje NO contiene material sensible** aunque estuviera en lo
  hallado. Nace en **rojo**.
- [x] **T015** Implementar la derivación con **validación ruidosa** en `internal/config/`. **Cita el
  contrato** (`contracts/adhesion.md` §Cómo se obtiene `<base>`) en el comentario: los dos hechos de
  ruta son **contrato, no literal local** (D-005-P3, D-005-P11). Pone en verde **T013–T014**.
  > ### 🔴 MEDIDO — los dos rojos, observados ANTES de T015
  > La derivación no existía, así que hubo que declarar **la firma y el centinela** en
  > `internal/config/adhesion_endpoint.go` para que los tests **compilaran** —un `[build failed]` no es
  > un rojo legible—, con un `errDerivacionAndamiaje` **distinto de `ErrFormaDeEndpointInesperada`**
  > para que T014 no naciera verde. **T015 lo retiró**; `grep` → cero.
  > ```
  > T013  config_test.go:345: DerivarEndpointDeAdhesion("https://api.permea.example/api/v1/ingest") devolvió err = config: andamiaje P-005 — derivación no implementada; es una forma válida
  >       config_test.go:348:   obtenido: ""   quiere: "https://api.permea.example/api/v1/projects/adhesion"
  > T014  config_test.go:399: errors.Is(err, ErrFormaDeEndpointInesperada) = false; err = config: andamiaje P-005 — derivación no implementada
  >       config_test.go:434: el rehúse NO nombra la forma: las dos formas dan el MISMO mensaje "config: andamiaje P-005 — derivación no implementada"
  > ```
  > **Las DOS aserciones de T013 dispararon en el mismo subtest**, que es la disciplina 3 §inmunidad
  > funcionando: con `t.Fatalf` la segunda no se habría visto.
  >
  > ### ✅ T013 — siete casos, y tres no son adorno
  > Además del puerto `:8443` **obligatorio** (el banco local; perderlo manda la adhesión al 443, a un
  > servidor que no es el que se estaba probando), se cubren **los tres sitios donde «sustituir el
  > último» se confunde con otra cosa**: prefijo de **varios segmentos** —cortar por el primero da otro
  > destino—, **host que contiene «ingest»** —no se toca—, y **segmento intermedio homónimo**
  > (`/ingest/v1/ingest`), donde sustituir *el primero que coincida* manda la petición a otro sitio.
  >
  > ### ✅ T014 — «nombrar la forma» se comprueba SIN casar texto
  > FR-009 pide nombrar la forma; la disciplina prohíbe aserciones sobre texto de mensajes. Se resuelve
  > por **distinguibilidad**, en test aparte: **dos formas estructuralmente distintas deben dar mensajes
  > distintos**. Un mensaje constante —«forma inesperada»— **no nombra nada**, y eso es lo que impide.
  >
  > **Y FR-020 manda sobre FR-009**: la forma se describe **estructuralmente** —ruta vacía, cuántos
  > segmentos, cuál se esperaba— y **nunca citando lo hallado**, porque el último segmento de un
  > endpoint mal configurado puede ser cualquier cosa. **Por eso T015 DESCARTA la causa de `url.Parse`**
  > —un `*url.Error` lleva la URL entera dentro—, al revés que `transport.Adherir`, que la conserva.
  > **Tercera puerta, tercera decisión, mismo juicio**: es exactamente para esto que `JuzgarEndpoint` no
  > formatea mensajes.
  >
  > ### ✅ CERRADO — lo que no se reconoce, se REHÚSA (userinfo · query · fragmento)
  > La primera versión **conservaba** lo que la URL trajera de más, «por ser lo menos sorprendente».
  > **No es criterio de frontera, y T015 es justamente la tarea de la validación ruidosa**: conservar en
  > silencio algo inesperado es lo contrario de rehusar en vez de conjeturar.
  >
  > **Lo decide USERINFO, y el rojo lo enseñó literalmente**:
  > ```
  > config_test.go:399: DerivarEndpointDeAdhesion("https://usuario:9HpQ3mZv7KxR2wLb@api.permea.example/api/v1/ingest") NO rehusó,
  >   devolvió "https://usuario:9HpQ3mZv7KxR2wLb@api.permea.example/api/v1/projects/adhesion"
  > config_test.go:399: … QUERY      → devolvió ".../projects/adhesion?token=9HpQ3mZv7KxR2wLb"
  > config_test.go:399: … FRAGMENTO  → devolvió ".../projects/adhesion#9HpQ3mZv7KxR2wLb"
  > ```
  > **Las credenciales copiadas a un destino nuevo**, que es lo que ningún criterio justifica. Query y
  > fragmento van con ellas por el mismo motivo: no se reconocen, no se conjetura qué significan.
  > **Cuatro casos nuevos en la tabla de T014** (userinfo con y sin contraseña, query, fragmento),
  > observados en rojo **antes** de tocar T015. El rehúse **nombra qué parte sobra, nunca su
  > contenido** (P-005 FR-020).

---

## Phase 6: El comando (D-005-P6, D-005-P7, D-005-P13)

### Tests (rojo antes de verde)

- [x] **T016** [P] Test de **las dos vías de entrada** en `cmd/permea/project_test.go` (nuevo) —
  **Garantía**: FR-023 + SC-011 (A). Mismo código por argumento y por entrada estándar → desenlaces
  **idénticos**, con las **tres piezas**: no vacío y del tipo que toca · idénticos entre sí · **y la
  comparación sabe fallar**. Nace en **rojo**.
- [x] **T017** [P] Test de **entrada ausente** en el mismo fichero — **Garantía**: `cli.md` §Entrada.
  Sin argumento y sin pipe → **error de uso** con **`ExitCode() == 1`**, y **NUNCA un prompt que se
  cuelgue**. Nace en **rojo**: T003 sale con **70**.
  ⚠️ **Se compara el código EXACTO, no «≠ 0»**, y es lo que hace que el 70 de T003 sirva de algo: con
  «≠ 0» esta tarea **nacería VERDE contra el andamiaje** —70 también es ≠ 0— y su rojo no existiría.
  El valor es **1** porque el binario tiene **dos** códigos, `0` y `1` (`cli.md` §Los códigos de
  salida), y el error de uso no es éxito.
  - [x] **T017b · el ENTRELAZADO entrada ↔ rehúses** — `join` **sin argumento Y fuera de árbol** →
    gana **el error de uso**. Es el único caso que lo observa: T017 monta árbol y config buenos, así
    que allí no hay competencia.
    **Razón, y va escrita**: rehusar por falta de árbol cuando la persona **no ha dado ningún código**
    es contestar a una pregunta que no hizo — se le señala un sitio al que ir con algo que todavía no
    tiene. El error de uso nombra lo que le falta **ahora**, que es el criterio de D-005-P13 aplicado a
    lo que **no** es uno de los tres rehúses.
    **Se observa comparando salidas entre sí** (disciplina 4), no casando texto: el `stderr` de la
    combinación debe ser **el mismo** que el del error de uso puro y **distinto** que el del rehúse del
    árbol puro.
    > ### ⚠️ NACIÓ VERDE — T021 ya lo resolvió así. VALIDADO POR MUTACIÓN
    > **La mutación**: mover la lectura de la entrada **detrás de los tres rehúses**, justo antes de
    > resolver el salt. **Válida por las dos mitades** (disciplina 3): **compila** —`go build ./...`
    > limpio, ni panic ni `build failed`— y **mata sólo ese hecho**: el único `FAIL` de `cmd/permea`
    > fue este test; T016, T017, T018, T019 y T020 siguieron en verde, porque T017 monta el árbol y no
    > hay competencia que invertir.
    > ```
    > --- FAIL: TestProjectJoin_LaEntradaGanaAlRehuseDelArbol
    >   sin argumento y fuera de árbol NO gana el error de uso
    >     obtenido:     "error: este directorio no pertenece a un árbol de trabajo con raíz
    >                    reconocible.\n       Ejecuta el comando dentro del árbol de trabajo…"
    >     error de uso: "error: uso: permea project join <código>  (recomendado: pásalo por stdin…)"
    >   el rehúse obtenido es indistinguible del rehúse del árbol: los dos dan «el del árbol»
    > ```
    > **Revertida por edición inversa**, y la reversión se comprobó contra el árbol de git:
    > `git diff --exit-code cmd/permea/project.go` → **sin diferencias**. No es «parece igual»: es el
    > mismo fichero.
- [x] **T018** [P] Test del **ORDEN de los tres rehúses** en el mismo fichero — **Garantía**:
  D-005-P13, `cli.md` §Comportamiento. Con las tres condiciones a la vez —sin árbol, sin enrolamiento
  y con configuración rota— el mensaje es **el del árbol**. Nace en **rojo**.
  *Sin este test el orden lo fija el primer camino que alguien escriba.*
  - [x] **T018b · las CUATRO combinaciones, con las tres condiciones capaces de coexistir** —
    árbol↔enrolamiento · árbol↔configuración · **enrolamiento↔configuración** · las tres a la vez.
    **Corrección medida antes de T021**: ver §«🔴 MEDIDO — T018 CERTIFICABA DOS PAREJAS DE TRES».
- [x] **T019** [P] Test de **cero peticiones fuera de árbol** en el mismo fichero — **Garantía**:
  SC-004, FR-006. **Con su observador declarado**: un destino instrumentado que **cuenta**. **Y con su
  caso positivo**: el **mismo** destino, con el comando lanzado **dentro** de un árbol, **registra
  exactamente una**. Nace en **rojo**.
  ⚠️ **Sin el caso positivo el test se cumple por no mirar**, y entonces no distingue «no se emitió» de
  «el observador no estaba conectado».
  ⚠️ **Por qué T019 NO lleva la marca «nace verde» que sí llevan T024–T026**, aunque también verifique
  una ausencia: **su caso positivo está DENTRO del mismo test** —el observador debe registrar
  exactamente una petición **dentro** de un árbol— y **eso sí falla** con el andamiaje de T003, que no
  emite nunca. **Nace rojo de verdad, por la mitad positiva.** No es un olvido.
- [x] **T020** [P] Test de **verbo desconocido y `project` sin verbo** en el mismo fichero —
  **Garantía**: `cli.md` §La gramática. Error de uso por stderr, **`ExitCode() == 1`**, nombrando lo no
  reconocido. Nace en **rojo**: T003 sale con **70**.
  ⚠️ **El código EXACTO, por la misma razón que T017**: con «≠ 0» el andamiaje la satisface y el rojo
  desaparece.
  > ### 🔴 MEDIDO — los CINCO nacieron rojos, y los cinco por su propia razón
  > `cmd/permea/project_test.go` (nuevo). Tests **de proceso**: se compara `ExitCode()`, nunca texto
  > (disciplina 4). **8 ok · 1 FAIL**, y el FAIL es `cmd/permea` con estos cinco **y nada más**.
  > ```
  > T016 project_test.go:182: vía argumento: ExitCode() = 70, se esperaba 0 (unión conseguida)
  >      project_test.go:185: vía argumento: stdout VACÍO; el éxito comunica la denominación por stdout
  > T017 project_test.go:229: ExitCode() = 70, se esperaba EXACTAMENTE 1 (error de uso)
  > T018 project_test.go:373: el rehúse obtenido es indistinguible de el del enrolamiento: los dos dan
  >      "error: `permea project join` todavía no está implementado\n"
  > T019 project_test.go:415: ExitCode() = 70 dentro de un árbol con enrolamiento y destino vivo; se esperaba 0
  >      project_test.go:418: el destino recibió 0 peticiones, se esperaba EXACTAMENTE 1
  > T020 project_test.go:448: ExitCode() = 70, se esperaba EXACTAMENTE 1 (error de uso)
  > ```
  > **El `70` aparece en cuatro de los cinco, y es la prueba de que el andamiaje de T003 hace su
  > trabajo**: con `1` —el valor del contrato— **T017 y T020 habrían nacido VERDES** acertando contra un
  > comando sin implementar.
  >
  > ### ✅ T016 — las TRES piezas, y por qué la tercera no es ceremonia
  > (1) **no vacío y del tipo que toca** —código `0` y stdout con contenido—, porque «iguales» no dice
  > nada si los dos son el mismo error; (2) **idénticos entre sí**, canal a canal; (3) **la comparación
  > sabe fallar**: se le pasan tres desenlaces que difieren en **un solo canal cada uno** y debe
  > declararlos distintos. Sin (3), un comparador que devolviera siempre `true` dejaría (2) en verde
  > para siempre.
  >
  > ### ✅ T018 — las TRES PAREJAS, y el orden comprobado SIN CASAR TEXTO
  > «Las tres condiciones a la vez → gana el árbol» **lo pasa igual un orden equivocado entre
  > enrolamiento y configuración**, porque el árbol gana de todas formas. Por eso van **las tres
  > parejas** además del caso triple.
  >
  > **Y el orden se observa comparando SALIDAS ENTRE SÍ**, no contra literales: el `stderr` de una
  > combinación debe ser **el mismo** que el del rehúse que gana **y distinto** del de los que pierden.
  > No depende de ninguna redacción ni del idioma del sistema —que es lo que la disciplina 4 protege—.
  > **El rojo llegó por la desigualdad**: hoy los tres rehúses dan **el mismo** mensaje de andamiaje.
  >
  > ### 🔴 MEDIDO — T018 CERTIFICABA DOS PAREJAS DE TRES, y se arregló ANTES de T021
  > **El bloque de arriba dice «las tres parejas» y el test tenía cuatro casos en su tabla. Los dos
  > eran ciertos y la garantía era falsa igual**, porque el defecto no estaba en la tabla: estaba en
  > **el montaje**.
  >
  > **La causa**: «sin enrolamiento» se representaba como **`config.json` NO EXISTE**, y eso es
  > **mutuamente excluyente** con «configuración de forma inesperada», que exige lo contrario —que el
  > fichero **exista** con un endpoint del que no se puede derivar el destino—. El montaje de la
  > pareja escribía la configuración rota **y acto seguido borraba el fichero**: la segunda condición
  > se destruía al montar la primera. Así, la combinación «enrolamiento + configuración» era en
  > realidad **sólo «sin enrolamiento»**, y «las tres a la vez» eran **dos**.
  >
  > **Y la que se perdía era LA ÚNICA QUE IMPORTA**: el árbol gana en todas las demás combinaciones,
  > así que **enrolamiento↔configuración es la única que vería un orden invertido entre esos dos**.
  >
  > **El remedio**: representar «sin enrolamiento» como **la configuración EXISTE SIN `device_token`**
  > —es «no está enrolada» con la misma literalidad, porque ese campo lo escribe `enroll`— y entonces
  > las tres condiciones **coexisten de verdad**. Las dos representaciones caen en el mismo rehúse
  > (fichero ausente → configuración por defecto, con token vacío), y por eso el cambio **no rompe las
  > combinaciones que ya montaba el fichero ausente**.
  >
  > **Por qué ANTES de T021 y no después**: T018 estaba en rojo y T021 la pone en verde. Arreglarla
  > después deja un tramo en el que **un test VERDE certifica dos parejas de tres con nombre de
  > comprobar el orden completo**. El rojo no engaña a nadie; el verde sí.
  >
  > **El rojo, transcrito** (`go test ./cmd/permea/ -run TestProjectJoin_ElOrdenDeLosTresRehuses`).
  > Con el andamiaje de T003 los tres rehúses dan **el mismo** mensaje, así que el rojo llega **por la
  > desigualdad** — las cuatro combinaciones, cinco aserciones (la triple dispara dos):
  > ```
  > --- FAIL: TestProjectJoin_ElOrdenDeLosTresRehuses/PAREJA_árbol_+_enrolamiento
  >   el rehúse obtenido es indistinguible de el del enrolamiento: los dos dan
  >   "error: `permea project join` todavía no está implementado\n"
  > --- FAIL: .../PAREJA_enrolamiento_+_configuración
  >   el rehúse obtenido es indistinguible de el de la configuración: los dos dan  «ídem»
  > --- FAIL: .../PAREJA_árbol_+_configuración
  >   el rehúse obtenido es indistinguible de el de la configuración: los dos dan  «ídem»
  > --- FAIL: .../LAS_TRES_a_la_vez
  >   el rehúse obtenido es indistinguible de el del enrolamiento: los dos dan     «ídem»
  >   el rehúse obtenido es indistinguible de el de la configuración: los dos dan  «ídem»
  > ```
  > **La mitad de igualdad —«el obtenido es el que gana»— siguió pasando con el montaje nuevo**, que
  > es la comprobación de que no se rompió nada de lo que ya funcionaba. La verificación completa
  > llega con T021: las cuatro en verde exige que los tres rehúses sean **distinguibles entre sí**, y
  > eso sólo lo puede cumplir un montaje en el que las tres condiciones existan de verdad.
  >
  > ### ✅ T019 — observador real, y el caso positivo con EL MISMO destino
  > El banco es un `httptest.NewTLSServer` que **cuenta las peticiones que le llegan de verdad**, y el
  > binario hijo confía en su certificado por **`SSL_CERT_FILE`** —que en Go es **aditivo**, medido en
  > este proyecto—. **El caso positivo es la mitad que importa**: «cero peticiones» lo cumple igual un
  > comando que no emite nunca, y hoy **es exactamente lo que pasa** — por eso el rojo está en el
  > positivo (`0`, se esperaba `1`) y no en el negativo.
  > *(Límite: en Windows Go usa el almacén del sistema y `SSL_CERT_FILE` se ignora.)*
  >
  > ### ✅ Disciplina 3 §inmunidad, aplicada
  > **Todas las aserciones independientes usan `t.Errorf`**, y se nota en el rojo: T016 y T019
  > dispararon **dos aserciones en el mismo subtest**. Con `t.Fatalf` la segunda de cada uno habría
  > quedado inalcanzable para cualquier mutación futura.

### Implementación

- [x] **T021** Implementar el comando en `cmd/permea/project.go` con el **patrón de dos capas de
  `enroll`** (D-005-P7, `enroll.go:18` y `:38`): capa sucia que resuelve stdin/TTY/stdout, capa pura
  con **lector, escritor y ejecutor inyectados**. Los tres rehúses **en el orden de T018**, **antes de
  emitir nada**. Pone en verde **T016–T020**.
  ⚠️ **La inyección no es comodidad**: `main_test.go:321-325` documenta que un proceso hijo **no
  confiaría** en el certificado del arnés, así que es **la única vía** de probar el camino completo.
  > ### ✅ MEDIDO — T016–T020 en verde, y la línea base VUELTA
  > `go test ./...` → **9 `ok` · 0 `[no test files]` · 0 `FAIL`**. Las cuatro combinaciones de T018
  > pasan, que es lo que acredita que **los tres rehúses son distinguibles entre sí** y que el montaje
  > corregido monta las tres condiciones de verdad.
  >
  > **Las tres capas, como manda D-005-P7**: `runProjectOS` (sucia: stdin, su naturaleza pipe/TTY, los
  > dos canales y el ejecutor real) → `runProject` (despacho del verbo) → `projectJoin` (pura: entrada,
  > los tres rehúses y el camino hasta emitir). El ejecutor es un `func(destino, token, código,
  > project_ref) (denominación, error)` y no un `*transport.Client`: el test inyecta una función, no
  > fabrica un cliente.
  >
  > ### ⛔ EL ANDAMIAJE DE T003 SE RETIRA AQUÍ, NO EN T023
  > T017 y T020 exigen `ExitCode() == 1` **exacto**, así que el `70` no podía sobrevivir a esta tarea.
  > **T023 queda de respaldo, no de responsable.** Barrido: el identificador que lo declaraba → **cero
  > en todo el repositorio**; `return 70` / `= 70` → **cero**. Disciplina 8 → **cero**.
  >
  > ### ⚠️ UNA PREMISA QUE CAYÓ AL MEDIRLA — `config.IsEnrolled` NO SIRVE PARA R2
  > Era el candidato obvio: es la función que ya responde «¿está enrolada?», y `status` la usa. **Y
  > usarla habría hecho INALCANZABLE la guarda de esquema del transporte.** `IsEnrolled` funde **tres**
  > hechos —hay endpoint, hay token y el esquema es admisible—, y el tercero **no es este rehúse**: un
  > endpoint en claro lo rechaza la frontera del transporte, «con la misma exigencia y sin exención»
  > (FR-017). `cli.md` §Notas lo dice **de este flujo en concreto**: *«aquí el endpoint viene de la
  > configuración persistida, así que la guarda del transporte SÍ se ejercita»*. Con `IsEnrolled`, un
  > `http://…` habría salido como «no estás enrolado» y esa nota del contrato habría quedado **escrita
  > y falsa**. R2 mira **lo que `enroll` escribe**: endpoint y token, y nada más.
  >
  > ### ⚠️ DOS DECISIONES QUE EL CONTRATO NO ORDENA, Y SE DECIDIERON AQUÍ
  > - **La entrada va ANTES de los tres rehúses.** El error de uso **no es uno de los tres** —`cli.md`
  >   lo deja fuera de la tabla a propósito—, así que D-005-P13 no lo ordena. Va primero porque
  >   responde a **qué se ha pedido**, y los tres responden a **si puede hacerse**. Es también el orden
  >   de `enroll`. **Ningún test observa el entrelazado**: T017 monta árbol y config buenos.
  > - **Una configuración ILEGIBLE cae en R3, y no rompe el orden.** El orden decide qué rehúse se
  >   reporta **cuando varios se cumplen**; aquí R2 ni siquiera es **evaluable** —sin poder leer el
  >   fichero no se sabe si hay enrolamiento, y afirmar que no lo hay sería afirmar lo que no se ha
  >   podido establecer—. Es la misma postura que FR-013 toma con el desenlace remoto.
  >
  > ### ⚠️ EL SALT SE RESUELVE DESPUÉS DE LOS TRES REHÚSES — y `huboRaiz` se pregunta sin él
  > `config.LoadOrCreateSalt` **crea** el secreto si no existe, así que pedirlo antes de tiempo haría
  > que **un camino de rehúse escribiera en local**. R1 no lo necesita: `huboRaiz` sale de `ascender`,
  > que sólo mira el sistema de ficheros, y el salt sólo entra en el hash. Se pregunta con salt vacío
  > y **el camino de rehúse no toca el secreto**. Eso sigue siendo cierto y sigue siendo el diseño.
  >
  > ### 🔴 LO QUE ESTA NOTA DECÍA MAL — «el camino de éxito» eran CUATRO, y hoy son CERO
  > **Primera corrección**: decía que el salt se crea «en el camino de éxito». **Medido: eran los
  > cuatro que EMITEN** —éxito, D1 (`422`), D2 (`409`) y no verificable—, porque el salt hace falta
  > para componer `project_ref` **antes de emitir**, y por tanto antes de saber cómo va a responder la
  > plataforma. Sólo los **tres rehúses locales** quedaban limpios.
  >
  > **Segunda corrección, y deja la primera sin efecto práctico**: **el salt pasa a nacer en el
  > enrolamiento** (`cmd/permea/enroll.go`, al terminar un enrolamiento correcto). Sobre una
  > instalación enrolada con el binario actual, **`project join` no crea nada en ninguno de los ocho
  > desenlaces** — medido, tabla en T025. El aviso «para T025, siembra las semillas» ya no describe un
  > apaño: describe **el estado que el enrolamiento deja**.
  >
  > ### ⛔ LO QUE ESTA TAREA NO HACE, Y ES DE T027
  > **La presentación de los desenlaces.** Todo error del ejecutor sale hoy por **stderr** con código
  > **1**, sin distinguir cuál fue: **canal y código ya son los definitivos**, el mensaje es
  > provisional. El mapeo de `cli.md` §Comportamiento —D3≡D4, D2 sin nombrar el Proyecto ajeno, D1 sin
  > indicar la causa— lo pone T027, y está marcado en el código donde va.
  >
  > ### 📏 ESTADO DEL LINTER, tal cual
  > `golangci-lint run` → **9 issues**, **6 de ellos anteriores a esta tarea** (`endpoint.go`,
  > `adhesion_test.go`, `project_test.go`, y 3 `errcheck` del propio andamiaje de T003 en
  > `project.go`). Los **2 nuevos** son `errcheck` sobre `fmt.Fprintln`/`Fprintf` a stderr, en los dos
  > caminos de rehúse que esta tarea añade — mismo estilo que `main.go` ya usa para sus mensajes de
  > error. **CI no ejecuta el linter** (`release.yml`); se anota, no se disimula.

---

## Phase 7: Canales, salidas y secretos

### Tests (rojo antes de verde)

- [x] **T022** [P] Test del **reparto de canales** en `cmd/permea/project_test.go` — **Garantía**:
  FR-021, SC-011 (B). Éxito → **stdout no vacío y stderr sin el desenlace**; rehúse o error → **stderr
  no vacío y stdout VACÍO**. **Capturados por separado** (disciplina 7). ~~Nace en **rojo**~~ →
  **NACIÓ VERDE**, ver abajo.
  - [x] **T022b · una denominación HOSTIL no inyecta en la terminal** — **Garantía**: P-005 FR-021 +
    Principio I. La denominación es **dato controlado por el SERVIDOR impreso en la consola del
    usuario**: inyección de terminal **a través de la frontera**. Con `\n`, `\r` y un escape ANSI
    dentro del nombre, la salida de éxito **no gana líneas** y **no emite bytes de control**.
    > **Existe porque el `%q` de T027 estaba sujeto por una DECISIÓN, no por una garantía.** Quien
    > mañana lo reponga a `%v` «porque las comillas quedan feas» la deshace **y nada se entera**.
    > **La aserción es ESTRUCTURAL** —número de líneas y presencia de bytes de control—, no contra un
    > texto esperado: así sobrevive a cualquier reescritura de la redacción, que es de T027.
    >
    > ### ⚠️ NACIÓ VERDE — VALIDADO POR MUTACIÓN, y **las dos aserciones dispararon**
    > `%q` → `%v`. Compila, y **mata sólo ese hecho**: el único `FAIL` en todo el repositorio fue este
    > test.
    > ```
    > la salida de éxito tiene 2 saltos de línea y debe tener EXACTAMENTE 1: una denominación con
    > `\n` dentro está FABRICANDO LÍNEAS que la persona leerá como del comando.
    > la salida de éxito emite 4 byte(s) de control (posiciones [26 56 64 74]): un `\r` borra lo que
    > el comando dijo y un escape ANSI repinta la terminal. La denominación viene DEL SERVIDOR.
    > stdout: "unido al Proyecto RecetApp\n  error: no se pudo completar\rborrado\x1b[31m rojo\x1b[0m\n"
    > ```
    > **Las DOS dispararon en la misma pasada** —van con `t.Errorf`—, así que ninguna quedó
    > aparentando validada detrás de la otra. Revertida por edición inversa.
  > ### 🟢 LA PREDICCIÓN DE §Las puertas de rojo CAYÓ AL MEDIRLA — T022 NACE VERDE
  > **La fila decía**: *«T021 implementa entrada y rehúses; la presentación llega en T027, así que no
  > hay reparto de canales que comprobar»*. **La premisa era que T021 no emitiría salida de desenlace**,
  > y **T016 no se lo permitió**: exige `stdout` **no vacío** en el éxito, así que T021 tuvo que
  > imprimir la denominación por stdout y los errores por stderr. **Los ocho desenlaces ya reparten
  > bien los canales antes de T027.**
  >
  > **No se «arregla» retorciendo el test: se declara.** Lo que le queda a T027 no es el reparto —ya
  > está— sino **la redacción** de cada desenlace.
  >
  > ### ✅ VALIDADO POR MUTACIÓN (disciplina 3, porque nació verde)
  > **La mutación**: que el camino de error del ejecutor escriba **también por stdout**. **Válida por
  > las dos mitades**: compila (`go build ./...` limpio) y **mata sólo ese hecho** — el único `FAIL` fue
  > T022, en sus tres desenlaces remotos de error; T016, T017, T018, T019, T020, T023, T024, T025 y
  > T026 siguieron en verde.
  > ```
  > --- FAIL: TestProjectJoin_ElRepartoDeCanales/{D2, D1, NV}
  >   stdout = "MUTACION: un rehúse remoto escribiendo por stdout\n" en un desenlace de REHÚSE o
  >   ERROR; se esperaba VACÍO: no hay respuesta que dar, y quien canalice stdout a un fichero no
  >   debe encontrarse un error dentro
  > ```
  > **Revertida por edición inversa**, comprobado contra git: `git diff --exit-code
  > cmd/permea/project.go` → sin diferencias.
  >
  > ### ⛔ «stderr SIN el desenlace» SE COMPRUEBA POR CANAL VACÍO
  > Buscar que la denominación **no aparezca** en stderr es la forma que la **disciplina 5** prohíbe:
  > pasa igual **cuando aparece otro texto distinto**. El canal vacío satisface «sin el desenlace» y
  > además es observable. Los dos canales viven en campos distintos de `desenlace` y **no se concatenan
  > en ningún punto** (disciplina 7).
- [x] **T023** [P] Test de **los ocho códigos de salida** en el mismo fichero — **Garantía**: `cli.md`
  §Los códigos de salida. Compara **`ExitCode()`, nunca texto** (disciplina 4).
  > ### 🟢 TAMBIÉN NACIÓ VERDE, Y POR LA MISMA CAUSA
  > **La fila decía** *«ídem: los ocho códigos los fija T027»*. **T017 y T020 se adelantaron**: exigen
  > `ExitCode() == 1` **exacto**, así que T021 tuvo que fijar los dos códigos del contrato **y retirar
  > el andamiaje de T003 ahí mismo**. Los ocho valores ya son los del contrato.
  >
  > ⛔ **Consecuencia sobre la nota de esta tarea**: decía que **T023 es quien retira el `70`**. **Ya no
  > lo es** —lo retiró T021—, y pasa a ser **lo que lo vigila**: el test sigue tumbando cualquier
  > superviviente, pero la retirada no dependía de él.
  >
  > ### ✅ VALIDADO POR MUTACIÓN
  > **La mutación**: que los desenlaces remotos de error salgan con `2` en vez de `1`. Compila, y
  > **mata sólo T023** —los tres desenlaces remotos de error—; los otros nueve tests, en verde.
  > ```
  > --- FAIL: TestProjectJoin_LosOchoCodigosDeSalida/{D2, D1, NV}
  >   ExitCode() = 2, se esperaba EXACTAMENTE 1. El binario tiene DOS códigos, `0` y `1`, y esta
  >   feature no amplía el vocabulario
  > ```
  > **Revertida por edición inversa**, `git diff --exit-code` → sin diferencias.
  >
  > ### ⚠️ D3 ≡ D4 NO SE PUEDE MUTAR, Y ESO ES LA GARANTÍA — no un hueco
  > El subtest que los compara **entre sí** está escrito y pasa, pero **ninguna mutación honesta puede
  > separarlos hoy**: el comando **nunca llega a saber cuál de los dos fue** —la plataforma responde lo
  > mismo (`adhesion.md` §Los cuatro desenlaces)—, así que darles códigos distintos exigiría **inventar
  > una distinción que no existe**. La indistinguibilidad es **estructural**, y este subtest es la
  > **guardia contra el futuro**: el día que alguien introduzca esa distinción, cae aquí. Se anota tal
  > cual porque un «validado por mutación» que no se puede ejecutar es peor que decir por qué no.
  > ⚠️ **El caso que no puede faltar**: **D3 y D4 comparten código**. Si alguien le diera a «ya
  > estabas unido» un código propio, **rompería FR-010** sin darse cuenta — el resultado del proceso
  > es observable. Este test es lo único que lo impide.
  >
  > ⚠️ **Y ES QUIEN RETIRA EL 70 DE T003.** Comparar los ocho códigos exactos **tumba cualquier
  > superviviente del andamiaje**, así que la retirada está **garantizada por el test** y no confiada
  > a que alguien se acuerde.
- [x] **T024** [P] Test de **no filtración del código** en el mismo fichero — **Garantía**: FR-020,
  SC-005. Para **cada uno de los ocho desenlaces** (los del comando, `cli.md`): generar las subcadenas
  de **longitud ocho** del valor presentado y buscarlas **en los dos canales**. **Cero apariciones.**
  ⚠️ **NACE VERDE, y hay que decirlo**: un comando sin implementar **no imprime el código** —no imprime
  casi nada—, así que la ausencia se cumple sola. **Se valida por SU CASO POSITIVO**, que ya está
  escrito en la propia tarea: sembrar deliberadamente el código en una salida debe **tumbar el test**,
  y hay que **leer el mensaje**. Sin eso, no distingue «no se filtra» de «no hay salida».
  > ### ✅ ESCRITO Y DECLARADO VERDE — el caso positivo queda SIN EJECUTAR, es deber de T027
  > Los ocho desenlaces × los dos canales, con las **43 subcadenas de longitud ocho** del código. Cero
  > apariciones. El caso positivo del **instrumento** —una salida fabricada que sí lleva el código debe
  > producir apariciones— está **escrito y con `t.Skip`**, y T027 lo retira junto con la siembra en la
  > salida REAL del comando, que es la mitad que demuestra que el detector está **conectado**.
  >
  > ### ⛔ EL CÓDIGO DEL ARNÉS ES DE ALTA ENTROPÍA, Y AHORA SE COMPRUEBA
  > Un fixture legible reinventaría la piedra del `" no perm"` **sobre los ocho desenlaces a la vez**:
  > una subcadena de ocho que choca por azar con la prosa de los mensajes. La regla deja de ser una
  > recomendación: el test **exige la forma del contrato** —`pmeaj1.` + 43 base64url = **50**
  > (`adhesion.md` §El código)— con un `t.Fatalf`, que aquí es lo correcto porque sin código conforme
  > **lo que siga no significa nada**. Nadie puede aflojarlo sin que el test lo diga.
- [x] **T025** [P] Test de **nada se escribe en local** en el mismo fichero — **Garantía**: FR-019,
  SC-007. Captura **íntegra** del conjunto enumerado de artefactos —configuración, estado, cola,
  secretos— **antes**, y comparación byte a byte **contra esa captura** tras cada desenlace.
  ⚠️ **NACE VERDE**: un comando sin implementar **no escribe nada**, así que la ausencia se cumple
  sola. **Se valida por SU CASO POSITIVO**, ya escrito: una operación de la instalación que **sí**
  modifica el estado local **hace fallar la misma comparación**. Si no falla ahí, el test no distingue
  «no cambió» de «no miré».
  > ### ✅ EL CONJUNTO ENUMERADO ESTÁ COMPLETO, Y NO EXCLUYE NADA
  > **Configuración, estado, cola y secretos**, sin cláusula, sin excepción y sin nota al pie. **Ya no
  > hace falta excluir nada, y ésa es la novedad**: la decisión de que **el salt nazca en el
  > enrolamiento** —`cmd/permea/enroll.go`, al terminar un enrolamiento correcto— deja a `project
  > join` **sin nada que crear**.
  >
  > ### 📏 POR QUÉ HIZO FALTA LA DECISIÓN, Y QUÉ SE MIDIÓ
  > **El salt es la semilla de `event.Ref`**, así que hace falta para componer el `project_ref` que la
  > adhesión presenta — o sea **antes de emitir**. Mientras lo creaba `project join`, lo creaba en
  > **los cuatro caminos que emiten**: éxito, D1 (`422`), D2 (`409`) y no verificable. Y P-005 SC-007
  > dice **«sin modificar» tras cualquier ejecución, con cualquier desenlace, incluidos los de rehúse
  > y los de error**, enumerando «secretos» dentro del conjunto. **P-005 FR-019 se dejaba leer acotado
  > a rastro de la adhesión; SC-007 no.** Los dos textos no podían ser ciertos a la vez.
  >
  > **La decisión los hace ciertos a los dos, literalmente y sin excepción escrita** — y lo hace
  > moviendo el hecho, no reinterpretando el texto. La spec está congelada y **no ha hecho falta
  > tocarla**.
  >
  > ### 🔬 MEDIDO CON EL BINARIO NUEVO — enrolamiento real, HOME temporal, banco TLS
  > Se **enrola de verdad**, se captura el directorio de datos, y se ejecutan los **cuatro caminos que
  > emiten** más un rehúse local. Comparación por hash, fichero a fichero:
  >
  > | Camino | Salida | `rc` | Directorio de datos |
  > |---|---|:--:|---|
  > | **Éxito (200)** | `unido al Proyecto "RecetApp"` | 0 | **cero diferencias** |
  > | **D1 (`422`)** | el código no es utilizable | 1 | **cero diferencias** |
  > | **D2 (`409`)** | esta identidad ya pertenece a otro proyecto | 1 | **cero diferencias** |
  > | **NV (servidor caído)** | no se pudo verificar el desenlace | 1 | **cero diferencias** |
  > | **R1 (fuera de árbol)** | ejecuta dentro del árbol de trabajo | 1 | **cero diferencias** |
  >
  > Tras el enrolamiento el directorio contiene exactamente **`config.json` y `salt`**, y **ninguno de
  > los cinco caminos lo mueve**.
  >
  > ### 🛡️ LA RED DEL ENROLAMIENTO SE AMPLIÓ AL DIRECTORIO ENTERO — y era necesario
  > `assertNoPersist` (`cmd/permea/enroll_reject_test.go`) comprobaba **sólo `config.json`**. Bastaba
  > mientras `enroll` sólo escribiera ese fichero. **Ya no**: el `salt` se crea dentro de una rama, y
  > **el día que esas líneas se muevan** —un refactor, un `defer` mal puesto— un enrolamiento rechazado
  > dejaría un secreto en el disco **y la red estaría mirando el fichero equivocado**. Un `config.json`
  > ausente es compatible con un directorio lleno.
  >
  > **No es ajustar la red: es ampliarla para cubrir el cambio que se acaba de introducir.** La
  > aserción anterior **no se toca** y la nueva se le suma, las dos con `t.Errorf` para que ninguna
  > quede inalcanzable. El conjunto se **deriva recorriendo el directorio**, no de una lista de nombres.
  >
  > **Nació verde, y se validó por mutación**: sacar la creación del salt fuera de la rama del
  > enrolamiento correcto —al principio de `enroll`—. **Compila**, y **mata sólo ese hecho**: cayeron
  > los **cinco** tests de rechazo, **todos por la aserción nueva**, y nada más en todo el repositorio.
  > ```
  > --- FAIL: TestEnroll_Reject_{401,5xx,Network,Malformed,Pmea1}
  >   un enrolamiento rechazado dejó 1 fichero(s) en el directorio de datos: [salt]
  >   El estado indistinguible (SC-004) es el del directorio ENTERO, no el de `config.json`:
  >   cualquier otro artefacto —el `salt` entre ellos— delata que hubo un intento
  > ```
  > **La aserción vieja NO disparó** —`config.json` seguía sin escribirse—, que es la prueba de que lo
  > que cazó la fuga fue la ampliación y no lo que ya había. Revertida por edición inversa;
  > `git diff --exit-code cmd/permea/enroll.go` → sin diferencias.
  >
  > ### 📏 PERMISOS DEL `salt`: **600** — medido, no supuesto
  > `stat` sobre cuatro enrolamientos reales: `600 config.json` y **`600 salt`**. No es suerte del
  > umask: `loadOrCreateSecret` pasa `0o600` a `atomicWrite`, que hace `Chmod` **antes** del `rename`.
  > **No hay deuda de permisos que decidir.**
  >
  > ### 🔬 QUÉ MONTA EL TEST, Y POR QUÉ SIGUE SEMBRANDO
  > El montaje escribe `config.json` **a mano** —no pasa por `enroll`—, así que **siembra los
  > secretos** para representar lo que una instalación enrolada tiene. **Ya no es un rodeo: es un
  > espejo.** Antes el sembrado tapaba un hueco; ahora **reproduce lo que el enrolamiento deja**, y
  > eso se puede comprobar comparándolo con la medida de arriba.
  > **Y la captura se DERIVA del directorio**, recorriéndolo entero, no de una lista escrita a mano:
  > una lista a ojo es el defecto que la disciplina 8 ya registró tres veces, y un artefacto nuevo se
  > le escaparía sin que nadie se enterara.

- [x] **T026** [P] Test de **la petición nunca se encola** en el mismo fichero — **Garantía**: FR-018,
  SC-010. Con el servidor **inalcanzable**, la cola inspeccionada **antes y después** no crece.
  ⚠️ **NACE VERDE**: un comando sin implementar **no encola nada**, así que la ausencia se cumple sola.
  **Se valida por SU CASO POSITIVO**, ya escrito: una emisión ordinaria de eventos con el destino
  igualmente caído **sí la hace crecer**. Si la cola no crece en ninguno de los dos casos, **el
  observador no está mirando** y el criterio no cuenta como pasado.
  > ### ✅ ESCRITO Y DECLARADO VERDE — el caso positivo queda SIN EJECUTAR
  > La cola se inspecciona con `transport.Load` —la lectura exportada, no una copia— antes y después,
  > con el banco **cerrado** para que el destino esté escrito y muerto. **Con precondición**: si el
  > comando saliera con `0` contra un destino cerrado, la circunstancia de SC-010 no se habría dado y
  > la cola no probaría nada; ahí `t.Fatalf` es correcto (disciplina 3).
  > El caso positivo —`--run` con el fixture de logs del repositorio y el mismo destino caído— está
  > **escrito y con `t.Skip`**; lo ejecuta T027.

### Implementación

- [x] **T027** Implementar la **presentación de los desenlaces** en `cmd/permea/project.go`: el mapeo
  de `cli.md` §Comportamiento —canal y código de salida por desenlace—, con **D3 y D4 produciendo
  salida idéntica**, **D2 sin nombrar el Proyecto ajeno** y **D1 sin indicar la causa**.
  ~~**Pone en verde T022 y T023.**~~ ⛔ **Los dos nacieron VERDES** (medido, ver T022 y T023): T021 ya
  había fijado canales y códigos. **Lo que esta tarea añade es LA REDACCIÓN**, en dos funciones —
  `mensajeDeUnion` y `mensajeDeRehuseRemoto`—, y **los tres casos positivos**.
  > ### ✅ LAS CUATRO PROMESAS, Y CUÁL SE SOSTIENE SOLA
  > - **D3 ≡ D4, byte a byte** — **estructural, no vigilada**: el comando **nunca llega a saber cuál de
  >   los dos fue** (la plataforma responde `200` igual en los dos), así que sólo existe **un camino**.
  >   Introducir una diferencia exigiría **inventar antes la distinción**. Reforzado igualmente: el
  >   subtest de T023 pasa de comparar sólo el código a comparar **los tres canales** con
  >   `mismoDesenlace` — que es lo que FR-010 pide («mismo texto, mismo canal, mismo resultado»).
  > - **D2 sin nombrar el Proyecto ajeno** — *«este árbol de trabajo ya pertenece a otro Proyecto»*.
  >   Tampoco podría: la plataforma no lo revela.
  > - **D1 sin indicar la causa** — las cinco causas del `422` llegan **como el mismo centinela**, así
  >   que producen el mismo mensaje **por construcción**: no hay dónde escribir la diferencia (FR-012).
  > - **P-005 FR-013a** — el no verificable dice que el desenlace **no se pudo establecer**, **no
  >   afirma ninguno**, y **deja volver a intentarlo**: *«el código no se agota al usarse y unirse dos
  >   veces es indistinguible de unirse una, así que repetir no duplica nada»*. **No es consuelo: es el
  >   requisito** — lo que la spec exige es que la incertidumbre sea **inocua**, y callarlo dejaría a la
  >   persona sin saber si repetir es seguro.
  >
  > **Se discrimina por CENTINELA (`errors.Is`), nunca por el texto del error**: casar cadenas ataría
  > esta redacción a la de `internal/transport`. Y **ningún mensaje recibe el código**, así que no puede
  > filtrarlo ni por descuido.
  >
  > **`%q` en la denominación no es decoración**: el nombre viene **del servidor**, y `%q` escapa los
  > caracteres de control — sin él, un Proyecto con un salto de línea o una secuencia ANSI podría
  > **fabricar líneas de salida** que la persona leería como del comando.
  >
  > **Dónde SÍ viaja la causa, y por qué sólo ahí**: en el no verificable. Red, TLS y DNS son
  > **accionables**; la causa de un rechazo no lo es **y además revela**. *(Residuo: esa causa puede
  > llevar la URL dentro y `url.Error` no redacta el `userinfo` — es la deuda ya anotada en §Lo que
  > este plan de tareas NO hace, con sus otros seis sitios. Ni se estrena aquí ni se amplía aquí.)*
  >
  > ### ✅ LOS TRES CASOS POSITIVOS — EJECUTADOS, y con el rojo LEÍDO
  > Los `t.Skip` retirados. **Dos mitades, y hacían falta las dos**: el subtest demuestra que el
  > instrumento **sabe disparar**; la siembra en el comando demuestra que está **conectado a él**.
  >
  > | | Subtest positivo | Siembra en el comando → rojo leído |
  > |---|---|---|
  > | **T024** | el detector encuentra el código en una salida fabricada | **8/8 desenlaces**: `stderr reproduce el código de adhesión: 43 subcadena(s) de 8 caracteres, la primera "pmeaj1.9"` |
  > | **T025** | `--run` escribe `state.json` y la comparación lo ve | **4/5**: `el comando modificó el estado local: APARECIÓ adhesion.json` |
  > | **T026** | `--run` con el destino caído hace crecer la cola | `la cola pasó de 0 a 1 eventos` |
  >
  > **Ninguno de los tres se quedó callado**: SC-005, SC-007 y SC-010 cuentan como pasados.
  >
  > ### 📐 D3 Y D4 SON COMPLEMENTARIAS — el hueco de cada una lo cierra la otra
  > La siembra de contenido **constante** tumbó **D4, D2, D1 y NV**, y **D3 pasó**. La explicación es
  > el momento de la captura, y **acota el hueco con precisión**:
  >
  > | Fila | Captura | Ve | No ve |
  > |---|---|---|---|
  > | **D4** | **antes de toda unión** | cualquier escritura de la **primera** unión | una que sólo empiece **en la segunda** |
  > | **D3** | tras una unión (eso la hace «segunda presentación») | lo que **cambie** en la segunda | una escritura **idéntica en cada unión** |
  >
  > **El hueco de D3 es ése y no más: una escritura que se repita IDÉNTICA.** Todo lo demás lo caza
  > D4, **incluido el marcador de «ya me uní»** —que es lo que P-005 FR-019 prohíbe **por su nombre**—,
  > porque su captura es anterior a cualquier unión. **Segunda mitad, también medida**: una escritura
  > **acumulativa** —contenido distinto en cada pasada— tumbó **las cinco filas, D3 incluida**.
  >
  > **Conclusión, y es lo contrario de «D3 no acredita nada»**: las dos filas cubren mitades distintas
  > y **T025 no tiene hueco**. Lo tendría si se quitara una de las dos. *(Y el caso que escaparía a
  > las dos —escribir sólo a partir de la segunda unión— exigiría que el comando **distinguiera D3 de
  > D4**, que es justo lo que T023 prohíbe y comprueba.)*
  >
  > ### ⚠️ SOLAPES ENTRE MUTACIONES, y por qué NO invalidan ninguna
  > La siembra de T024 tumbó **también T022** —escribir en stderr en un desenlace de éxito rompe «stderr
  > vacío»—, y la de T026 tumbó **también T025** —encolar **es** escribir en local—. En los dos casos es
  > **UNA alteración observada por dos criterios distintos**, no una mutación que mate de más: no hay
  > forma de filtrar un código sin escribir en un canal, ni de encolar sin tocar el disco. Se anota
  > porque «mata sólo ese hecho» se comprueba sobre **el hecho**, no sobre el recuento de tests.
  >
  > ### 📏 EL TIEMPO DEL CASO POSITIVO DE T026
  > Contra un destino caído, el drenaje entra en el **backoff acotado** del transporte y tardaría más
  > de medio minuto en rendirse. **La cola ya creció antes**: el agente encola de forma durable **antes**
  > de drenar (durabilidad, R4), que es justo lo que SC-010 observa. El límite se bajó a **5 s** y el
  > proceso se corta a propósito — mide lo mismo sin pagar el backoff.

---

## Phase 8: Los TRES criterios que ninguna historia arrastra (D-005-P9)

> ⚠️ **Tarea propia, y no cuelgan de ninguna historia.** La matriz del checklist lo dejó medido: las
> columnas de SC-001, SC-008 y SC-009 están **enteras en «—»**. **Un troceo por historias los habría
> dejado fuera**, y son los tres que protegen lo que ya funciona.

- [x] **T028** **SC-001** — en **`internal/project/resolve_test.go`**: la identidad presentada **es** la
  estampada, demostrado por **origen compartido**: (a) punto único del que salen las dos, y alterarlo **cambia las dos a la vez**; (b)
  comparación sobre **cuatro clases de árbol** —raíz, subdirectorio profundo, árbol paralelo,
  directorio sin raíz—; (c) **una alteración deliberada en el punto único pone (b) en rojo**.
  **Sin (c) el criterio no es falsable.**
  > ### 🔴 UNA PREMISA QUE CAYÓ AL MEDIRLA — **`DerivarConRaiz` NO TENÍA NI UN TEST**
  > Se añadió en Phase 1 (D-005-P5) y **`grep` la encontró sin cobertura ninguna**: los quince tests
  > de `resolve_test.go` son todos de `Derivar` y del `Resolutor`. T028 es lo primero que la mira.
  >
  > ### ⚠️ Y OTRA, QUE CAMBIA LA FORMA DE (a) — `Derivar` DELEGA EN `DerivarConRaiz`
  > El plan habla de «un punto único del que salen las dos» como si fueran hermanas. **No lo son**:
  > `Derivar` **llama a `DerivarConRaiz`**, que llama a `derivarConTechoYRaiz`. La vía del comando está
  > **aguas arriba** de la de la ingesta. La consecuencia es la que decide el diseño de (c):
  > **alterar el cuerpo compartido NO pone en rojo la igualdad** —mueve las dos a la vez y siguen
  > iguales—, que es exactamente lo que (a) promete y lo que haría **infalsable** un test que sólo
  > comparase las dos entre sí.
  >
  > ### ✅ (b) · LAS CUATRO CLASES, Y LA CUARTA NO COMPARA IDENTIDADES
  > Clases 1–3 —raíz, subdirectorio profundo, árbol paralelo— comparan **carácter a carácter** la
  > presentada (`DerivarConRaiz`) contra **las dos formas** de la estampada: `Derivar` y
  > **`Resolutor.Derivar`**, que es la que estampa de verdad en la pasada. Con **la vacuidad
  > comprobada primero**: dos cadenas vacías son iguales carácter a carácter, y sin esa aserción las
  > demás pasarían en verde justo con la derivación rota del todo.
  > **Clase 4 —directorio sin raíz— compara EL JUICIO**: que `huboRaiz` sea `false`. No hay identidad
  > presentada que comparar (el comando rehúsa antes de emitir, FR-006), y el valor que la derivación
  > sí produce —el fallback, P-004 FR-005— **nunca se presenta**. Se comprueba además que **no es
  > vacío**, porque quien tomase «vacío» por «no hubo raíz» dejaría de rehusar en silencio el día que
  > el fallback devolviera valor.
  > *(Los marcadores se fabrican a mano —directorio y FICHERO—: que git produzca esas dos formas ya lo
  > acredita `TestDerivar_ArbolesDeTrabajoParalelos`, y aquí el sujeto es la comparación, no git.)*
  >
  > ### ✅ (c) · **DOS MUTACIONES, PORQUE SON DOS HECHOS** (disciplina 3 §inmunidad)
  > Todas las aserciones de (b) usan `t.Errorf`, así que **todas se evalúan en la misma pasada** — y se
  > vio: cada mutación tumbó su grupo y dejó el otro en pie, que es la prueba de que ninguna quedaba
  > aparentando validada detrás de otra.
  >
  > **C1 · romper el REPARTO** (`Derivar` deja de delegar) → caen **las dos igualdades**, en las tres
  > clases. El anclaje **sobrevive**.
  > ```
  > la identidad PRESENTADA no es la ESTAMPADA (P-005 FR-004, SC-001)
  >   presentada (DerivarConRaiz): 7a3c118aeb93d94d…    estampada (Derivar):            e14653bf7348ff48…
  > la identidad PRESENTADA no es la que estampa la INGESTA por su camino real
  >   presentada (DerivarConRaiz): 7a3c118aeb93d94d…    estampada (Resolutor.Derivar):  e14653bf7348ff48…
  > ```
  > **C2 · alterar el CUERPO COMPARTIDO** → las igualdades **siguen verdes** (las dos se movieron a la
  > vez, que es (a)) y cae **el anclaje**… **y a la vez, en otro paquete, la regresión cero de la
  > ingesta**. **Las dos mitades moviéndose juntas, observadas en una sola pasada**:
  > ```
  > internal/project  la identidad presentada no es `event.Ref(salt, raíz)` sobre la raíz esperada:
  >                     presentada: 7f2e16974189ea59…   anclaje: e2ab77a43046374 0…
  > internal/ingest   TestSC009_RegresionCeroDelCaminoDeIngesta: identidad 0 difiere del baseline:
  >                     got:  3a32e2741c8d2301…        want: ae35623a2d739685…
  > ```
  > **El anclaje es lo que hace observable (a)**: ancla que el cuerpo compartido compone
  > `event.Ref(salt, raízDelAscenso)` —recalculado con las mismas piezas públicas, no un hash escrito a
  > mano—, así que una alteración ahí sí lo tumba. Sin él, C2 no habría puesto nada en rojo del lado
  > presentado y (a) se habría quedado en afirmación.
  > **Las dos revertidas por edición inversa**: `git diff --exit-code internal/project/resolve.go` →
  > sin diferencias.
  >
  > ⛔ **NO se usó `--scan`**: su salt es literal de dry-run (`cmd/permea/main.go`, `dryRun`) y sus refs
  > **no comparan con nada** (`research.md` §R5). Las dos vías se piden **a las funciones**, con el
  > mismo salt.
- [x] **T029-R** **SC-009 · RE-EJECUCIÓN, puerta de SC-009** — volver a pasar
  **`internal/ingest/baseline_regresion_test.go`** (el de T029) con **todo el código de la feature ya
  escrito**. Es lo que **acredita** que la feature entera **no cambió el camino de ingesta**.
  > ### ✅ ACREDITADO — `TestSC009_RegresionCeroDelCaminoDeIngesta` **PASS**
  > Con Phase 2 a Phase 8 en su sitio, incluido el `salt` naciendo en el enrolamiento. **Nada escrito
  > aquí**: se vuelve a pasar, que es lo que la tarea pide.
  >
  > **Y esta vez su falsabilidad no es teórica**: las **dos mutaciones de T028** lo tumbaron —C1 y C2
  > cambian la identidad que la ingesta estampa, y el baseline lo vio con las semillas deterministas—.
  > Un baseline que no cae con nada es un baseline que no compara; éste cae.
- [x] **T030** **SC-008** — en **`internal/transport/adhesion_test.go`**: sin transporte seguro no se
  completa, en **las cuatro clases enumeradas**:
  (a) destino en claro · (b) destino en claro sobre la máquina local · (c) destino inseguro **con un
  código utilizable**, para que el rechazo **no pueda atribuirse al código** · (d) los tres anteriores
  con **cada** ajuste de configuración que la instalación admita.
  > ### 📏 (d) · LOS AJUSTES QUE LA INSTALACIÓN ADMITE — **SIETE**, derivados del struct
  > `endpoint` · `device_token` · `org_id` · `dev_id` · `tools` · `sync_interval` · `logs_root`.
  > **No es una lista escrita a mano**: el test la **deriva por reflexión de las etiquetas `json` de
  > `config.Config`**, así que un ajuste nuevo entra solo en la matriz. Y si es de un tipo que el test
  > no sabe mover, **para con `t.Fatalf`** en vez de dejar su casilla en verde sin variarlo.
  >
  > ### ✅ NINGUNO EXIME LA GUARDA — barrido, no impresión
  > | Vía de exención | Medida |
  > |---|---|
  > | `InsecureSkipVerify` | **0** en todo el repositorio |
  > | `tls.Config` en el cliente | **0** fuera de tests |
  > | `os.Getenv` / `os.LookupEnv` | **0** fuera de tests → **ninguna variable de entorno puede eximir** |
  > | parámetro de modo en `JuzgarEndpoint` | **ninguno**: `JuzgarEndpoint(endpoint string)` y nada más |
  > | literales `http://` admitidos en código | **0** fuera de tests |
  >
  > *(Y `config.Load` **no** usa `DisallowUnknownFields`: una clave inventada como `allow_insecure` se
  > descarta **en silencio**, que es el lado seguro — no hay dónde encender nada.)*
  >
  > ### ✅ LA MATRIZ: **7 ajustes × 4 clases = 28 casillas**, más el control positivo
  > Las cuatro clases: en claro contra un host cualquiera · en claro sobre **`localhost`** · en claro
  > sobre **`127.0.0.1`** · en claro contra un banco **VIVO y dispuesto a aceptar el código**. Las 28
  > devuelven el centinela **`ErrScheme`** y **denominación vacía**.
  >
  > ⛔ **EL CONTROL POSITIVO VA PRIMERO Y NO ES CEREMONIA**: el **mismo** código, contra transporte
  > seguro, **sí completa** y devuelve la denominación. Sin él, «no se completó» lo explicaría igual un
  > código inservible y la clase (c) no significaría nada.
  >
  > ⛔ **Y LA MITAD QUE CIERRA (c): el banco en claro está VIVO y recibió CERO peticiones.** Contra un
  > host inexistente, «no se completó» lo explicaría un fallo de red. Con un destino que **aceptaría**,
  > cero peticiones sólo tiene una lectura: **la guarda rechazó ANTES de transmitir**. Es la diferencia
  > entre «no se completó» y «no se intentó siquiera», y sólo la segunda es lo que FR-017 promete.

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
  > ### ✅ EL `salt` YA NO ES UN RESIDUO DE ESTA CEREMONIA
  > Hubo un tramo en el que **`project join` creaba el `salt`** si no existía —en los cuatro caminos
  > que emiten—, y esta tarea habría salido bien **por la historia del sujeto**: `~/dev/test/RecetApp`
  > lleva miles de eventos y su `salt` existe desde hace tiempo. **Eso ya no es lo que la sostiene.**
  >
  > **El salt nace en el enrolamiento** (`cmd/permea/enroll.go`), así que la comparación byte a byte
  > de C4 pasa **por una propiedad del comando**: `project join` no tiene nada que crear, con ningún
  > desenlace. Medido con binario real y enrolamiento real sobre los cuatro caminos que emiten más un
  > rehúse local — **cero diferencias en los cinco** (la tabla está en T025).
  >
  > **Lo que sigue siendo deber de esta tarea**: que la captura previa la tome **T033**, porque
  > `quickstart.md` §C4 dice «con su estado previo» y **no toma ninguno**. Sin él, esto compara contra
  > un recuerdo.
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
| **T022** | T001–T021 | ~~rojo~~ → ⚠️ **VERDE, medido** | **la predicción cayó**: T016 exige stdout no vacío en el éxito, así que T021 tuvo que repartir los canales. **Validado por mutación** (ver T022) |
| **T023** | T001–T021 | ~~rojo~~ → ⚠️ **VERDE, medido** | **ídem**: T017 y T020 exigen `ExitCode() == 1` exacto, así que T021 fijó los dos códigos **y retiró el andamiaje**. **Validado por mutación** (ver T023) |
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

### ⚠️ Verificación de recuento — **37 tareas, 38 casillas**, y no es un desajuste

**Un barrido mecánico contará 38 `- [ ]` y sólo 36 identificadores `T0nn` distintos** —porque los dos
sufijados colapsan sobre su número— **y dirá que sobran dos.** No sobran, y son dos casos distintos:

**1 · `T029` y `T029-R` son DOS CASILLAS DE UN SOLO MECANISMO** —el mismo fichero de test, escrito una
vez y pasado dos—. Tienen casilla separada porque **se cierran en momentos distintos**: T029 al acabar
Phase 1, cuando su única dependencia (T001) ya está y **un fallo tiene un solo sospechoso**; T029-R al
acabar Phase 8, con todo el código escrito, **acreditando** SC-009.

**Sin las dos casillas, una de las dos ejecuciones no tendría dónde registrarse** — y fue exactamente
el defecto que se corrigió el 2026-08-18: un checkpoint que exige una comprobación sin casilla es una
intención, no una puerta.

**2 · `T010-E` es una TAREA NUEVA, no una segunda casilla de T010.** Nació en Phase 4 (2026-08-18): el
**estado que el contrato no enumera** —`5xx`, `403`, `404`, un `2xx` que no sea `200`— tiene que ser
**no verificable** por Principio I, y **el enunciado de T010 no lo cubría**: T010 mira un `200` cuyo
cuerpo no se puede leer, que es otro hecho. Lleva sufijo en vez de número propio para **no renumerar
las veintiséis tareas siguientes**, que romperían todas las referencias cruzadas del fichero.

**Su casilla es la razón de que exista, igual que la de T029-R**: sin ella, su rojo —observado antes de
T012— **no habría tenido dónde registrarse**, y el hueco del contrato que destapó no tendría ancla.

**Recuento oficial: 37 tareas · 38 casillas · 1 tarea con dos casillas (T029/T029-R) · 1 tarea con
identificador sufijado (T010-E).**

**Y cuatro tests nacen verdes, no tres**: **T024, T025, T026** (ausencias, validadas por su caso
positivo) **y T029** (regresión cero, validada por mutación del cuerpo compartido de T001).

### El barrido en el sentido que faltaba — de la TAREA al requisito

La comprobación de arriba va **de requisito a tarea**: garantiza que ningún FR/SC se queda sin quien lo
sirva. **Pero le falta el recíproco**, y es donde se esconde el desajuste silencioso: **toda tarea que
se declare a sí misma sirviendo un FR o un SC tiene que aparecer en la fila de ese FR o SC.** Si una
tarea dice «Garantía: FR-021» y la fila de FR-021 no la nombra, una de las dos miente — y como la tabla
es lo que se lee al cerrar, gana la tabla y el trabajo de esa tarea se pierde de vista.

**Recorridas las 36 tareas contra sus propias declaraciones de garantía. Tres desajustes, corregidos:**

> *(Barrido del 2026-08-18 sobre las 36 de entonces. **T010-E se añadió después**, en Phase 4, y se
> comprobó al crearla: declara **Principio I** —no un FR ni un SC—, así que **no le corresponde fila**
> en la tabla de requisitos, igual que a T020. Queda dicho aquí para que el próximo barrido no la
> cuente como huérfana.)*

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
- **No migra las instalaciones enroladas con el binario ANTERIOR**: al backlog, **acotado y medido**.
  > El `salt` pasa a nacer en el enrolamiento (`cmd/permea/enroll.go`), pero **quien se enroló antes de
  > ese cambio no lo tiene**. Su **primer** `permea project join` seguirá creándolo —lo crea
  > `config.LoadOrCreateSalt` al componer `project_ref`—, así que en esa instalación, y **sólo la
  > primera vez**, P-005 SC-007 vería aparecer un fichero.
  >
  > **Por qué se anota en vez de arreglarse**: es **residuo conocido, acotado y que se agota solo** —a
  > partir de la segunda ejecución no vuelve a pasar, y cualquier `--run` o `--daemon` lo agota
  > también—. Y el remedio —una migración al arranque, o un `enroll` de reparación— es una decisión de
  > producto, no un parche de esta feature. **Se escribe ahora para que nadie lo encuentre en
  > producción sin explicación.**
- **No pone el linter en CI**: al backlog, **medido**.
  > `golangci-lint run` → **8 issues** (5 `errcheck` en `cmd/permea/project.go`, 2 `revive`, 1
  > `staticcheck`), y **`.github/workflows/release.yml` NO lo ejecuta** — comprobado. `make lint`
  > existe y hay `.golangci.yml`, así que la configuración está y lo que falta es **la puerta**. Un
  > linter que nadie ejecuta es un linter cuyo recuento sólo sube.
- **No arregla la fuga de USERINFO en los mensajes de endpoint**: al backlog, **medido y no impresión**.
  > **`url.Error` NO REDACTA** —medido: `url.Parse("https://usuario:CLAVE@…/ingest\x7f")` devuelve
  > `parse "https://usuario:CLAVE@…"`, con la contraseña dentro—. Pero el problema **es más ancho que
  > la causa envuelta**: **siete sitios imprimen el endpoint verbatim**, y seis de ellos lo hacen con
  > `%q` de `c.Endpoint` **directamente**, así que filtrarían igual aunque nadie envolviera nada:
  > `config.go:102` · `config.go:105` · `transport.go:150` · `transport.go:153` · `transport.go:258` ·
  > `transport.go:261`.
  >
  > **El séptimo es el peor y no es un error**: `cmd/permea/status.go:53` imprime `cfg.Endpoint` **por
  > stdout en la salida normal de `status`** —y a propósito: `status_test.go` comprueba que aparezca—.
  > No es una fuga en un caso raro: es la salida de todos los días.
  >
  > **Cobertura actual: CERO.** Ningún test de `internal/` ni de `cmd/` usa un endpoint con userinfo.
  > **Remedio conocido de la stdlib**: `(*url.URL).Redacted()`, que sustituye la contraseña por
  > `xxxxx` —medido—. **Tensión con el caso 3 de T011**, que **exige** que `Adherir` conserve la causa
  > de `url.Parse`: conservarla y redactarla a la vez requiere decidir qué se conserva exactamente, y
  > eso es una conversación, no un parche. *(Citas por línea: esto es `tasks.md`, artefacto fechado —
  > disciplina 8.)*
- **No toca la plataforma.** La segunda mitad de D-005-P11 —que la plataforma cite el contrato en vez
  de definirlo— es **de otro día y de otra rama**.
- **No automatiza la ceremonia.** Está prohibido por la spec.
