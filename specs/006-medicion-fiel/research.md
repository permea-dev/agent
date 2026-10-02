# Research — Medición fiel y publicable (Phase 0)

**Feature**: `006-medicion-fiel` | **Fecha de las mediciones**: 2026-10-02 | **Contra**: `03d3734`
(código idéntico a `0311fa1`; el commit sólo añade la spec)

> **Qué es este documento y qué NO es.** Aquí está **lo medido** y la decisión que cada medida
> sostiene; en [`plan.md`](./plan.md) están las consecuencias, ordenadas en bloques. Cada decisión
> lleva su evidencia: `fichero:línea` contra `03d3734`, o una medida con su fecha. Lo no comprobado se
> dice «sin comprobar».
>
> **Citas por línea**: son fotos con fecha de este documento (disciplina 8 de 005,
> `specs/005-adhesion-a-proyecto/tasks.md`). **Caducan en cuanto 006 edite el fichero citado**, y casi
> todos los citados aquí se editan. Las tareas tendrán que citarlos **por contenido o por nombre**.

---

## R0 · Las cuatro cuestiones que la spec dejó al plan — resumen

| Cuestión (spec §Cuestiones para el plan) | Decisión | Dónde |
|---|---|---|
| Dónde vive la derivación del `event_id` y qué pasa con `event.NewID` | En `internal/ingest`, junto al lector de Claude Code. `event.NewID` **se queda** (D-006-3 prohíbe tocar `internal/event`), sin llamantes de producción, y un testigo de conducta impide que vuelva | R1, R2 |
| Cómo se vigila el espejo de tarifas sin depender del otro repo | Los valores esperados se escriben **aparte**, en el test, clave a clave, con el recuento de claves y el commit replicado. El otro repositorio sólo se consulta en una validación manual | R6 |
| Canal del diagnóstico local | **stderr**, un resumen por pasada, sólo con recuentos por categoría | R4 |
| Quién consume el formato del dry-run | **Ningún test depende de su formato**. Un test lo ejecuta y sólo mira el código de salida, y otro declara expresamente que no lo usa. Puede cambiar | R5 |

---

## R1 · La derivación del `event_id`

### R1.1 · Lo que hay hoy

- `internal/event/event.go:48-55`: `NewID` produce 16 bytes de `crypto/rand` en hex, 32 caracteres.
- Su **único llamante de producción** es `internal/ingest/claudecode.go:67`, por cada línea facturable
  (`grep -rn "NewID" --include=*.go`: `event.go:49`, `claudecode.go:67`, y nada más).
- `internal/event/event_test.go` **no lo prueba** (el mismo `grep` sobre el test no da nada).

### R1.2 · Los identificadores de entrada (medido 2026-10-02, sólo longitudes y prefijos)

Sobre `~/.claude/projects` (WSL), en las líneas facturables no sintéticas (22 972 en esta medición; el
historial crece mientras se mide):

| Campo | Longitud | Forma |
|---|---|---|
| `message.id` | **28** en el 100 % | `msg_` + 24 alfanuméricos, en el 100 % |
| `requestId` | **28** en el 100 % | `req_` + 24 alfanuméricos, en el 100 % |

Cuántos bits aleatorios aportan esos 24 caracteres: **sin comprobar**. Es lo que sostiene FR-003, que
exige que no se puedan obtener a partir del `event_id`. Aun en el peor caso razonable, invertir un
SHA-256 sobre esa entrada no es practicable.

### R1.3 · Decisión

**SHA-256 con separación de dominio y versión, longitud prefijada, truncado a 16 bytes (32 hex).** La
propuesta del orquestador se sostiene, con dos precisiones:
- **la herramienta forma parte del dominio**;
- **cada componente va prefijado con su longitud**.

El contrato completo, con vectores de prueba, está en [`contracts/event-id.md`](./contracts/event-id.md).

| Pieza | Decisión | Evidencia o motivo |
|---|---|---|
| Función | SHA-256, `crypto/sha256` de la librería estándar | Principio III: cero dependencias. `gosec` no la marca (sólo marca md5/sha1, G401) |
| Dominio | `permea/event_id/v1` + herramienta (`claude_code`) + tipo (`par` / `solo_message_id` / `solo_request_id`) | FR-006 exige que «sólo uno» no coincida con «par». La **versión** permite cambiar la derivación sin colisionar con la anterior. La **herramienta** evita que dos lectores futuros (Cursor, Copilot…) puedan producir el mismo `event_id` a partir de identificadores que coinciden por azar |
| Codificación | Cada componente con prefijo de longitud de 4 bytes, big-endian | Sin prefijo, `("a","bc")` y `("ab","c")` concatenan igual. Medido con los vectores: con prefijo dan `6e7e350c…` frente a `918d6e4c…` |
| Truncado | Primeros 16 bytes → 32 hex en minúsculas | D-006-8 / FR-004: es la forma actual (M3). 128 bits: una colisión por azar requiere del orden de 2^64 mensajes |
| Sin sal | Ningún dato local entra en el hash | D-006-2 / FR-002. Comprobable: el mismo par con sales distintas da el mismo `event_id` (R2) |

**Alternativas descartadas:**
- **HMAC con clave fija.** Una clave publicada en código abierto no añade nada a un dominio publicado, y
  confunde a quien audita.
- **UUIDv5.** Su forma (36 caracteres con guiones) no es la actual, así que choca con D-006-8.
- **El hash completo (64 hex).** Cabe en la plataforma, pero D-006-8 fija 32.
- **Truncar el `message.id` en claro.** Viola FR-003.

### R1.4 · Dónde vive — `internal/ingest`, no `internal/event`

| Opción | Veredicto | Motivo |
|---|---|---|
| `internal/event` (junto a `NewID`) | ❌ | **D-006-3 / FR-012 / SC-005**: `git diff internal/event` contra `0311fa1` tiene que salir vacío |
| `internal/ingest`, fichero propio junto a `claudecode.go` | ✅ | Los identificadores de entrada **son de Claude Code**: otra herramienta tendrá otros, y su lector derivará con su propio dominio. Es el mismo paquete que ya decodifica `rawRecord` y construye el evento (`claudecode.go:25-39`, `:73-91`) |
| Paquete nuevo (`internal/eventid`) | ❌ | Pasaría de 9 paquetes a 10, y 005 ya registró que esa cuenta es una puerta: «un paquete nuevo lo convertiría en 10 y rompería la puerta» (`internal/ingest/baseline_regresion_test.go:27-29`) |

### R1.5 · Qué pasa con `event.NewID`

**Se queda donde está, exportado y sin llamantes de producción.** Borrarlo tocaría `internal/event`
(D-006-3). El linter no lo marcará: `unused` no informa de funciones exportadas. **Riesgo**: alguien
vuelve a llamarlo. **Contención**: el testigo de conducta de R2, más una comprobación mecánica en el
quickstart (`grep -rn "event.NewID" --include=*.go` fuera de `internal/event` → vacío).

---

## R2 · Que lo aleatorio no vuelva (FR-008)

Un `grep` caza el `NewID` que hoy existe, pero no un `rand.Read` escrito en otro sitio. **Decisión: un
testigo de conducta**, en tres aserciones independientes (`t.Errorf`, disciplina 3 §inmunidad):

1. **La misma línea, dos veces, en dos `Context` distintos** (sal, máquina y desarrollador distintos):
   el mismo `event_id`. Con la conducta de hoy esto **está en ROJO**, porque `NewID` da dos valores
   distintos. Es el rojo de B1.
2. **El vector de prueba de `contracts/event-id.md`**: el par sintético da exactamente
   `43b8b3b6446704ae3cb8bb74683bf3e2`. Fija la derivación, no sólo su estabilidad.
3. **Las tres formas** (par, sólo `message.id`, sólo `requestId`) dan tres valores distintos entre sí,
   todos de 32 hex.

**Por qué no basta el `grep`**: comprueba un nombre, no una propiedad. La propiedad que FR-008 protege
es «misma entrada → misma salida», y sólo la comprueba ejecutar la entrada.

---

## R3 · Deduplicación dentro de la pasada (FR-033) — dónde vive el conjunto

### R3.1 · El precedente exacto

004 ya resolvió «memoria de una pasada, compartida por todos los caminos»:
- **el `Resolutor` de identidad de proyecto** vive en `ingest.Context` como puntero, **y nil es
  válido** (`internal/ingest/claudecode.go:48-54`);
- **se instancia por pasada** en `generate()` (`cmd/permea/main.go:226-233`), con la razón escrita: en
  `--daemon` el proceso vive días, y una caché de proceso serviría datos obsoletos;
- **`tick()` llama a `generate()`** (`main.go:335-340`), así que el daemon estrena una por ciclo.

### R3.2 · Decisión

**Una «pasada» en `internal/ingest`** con el mismo patrón: puntero en `ingest.Context`, nil válido,
instanciada por pasada. Guarda:
- **el conjunto de `event_id` ya emitidos**, con el `usage` de la primera línea (FR-005 necesita
  compararlo);
- **los contadores del diagnóstico** (R4).

| Camino | Quién instancia la pasada | Ámbito |
|---|---|---|
| `--run` | `generate()` | una pasada = una ejecución |
| `--daemon` | `generate()`, llamado desde `tick()` | **una pasada por ciclo**. Entre ciclos no hay memoria, y eso es lo que pide FR-033 |
| `--scan` | `dryRun()` (`main.go:360-392`, que hoy arma su propio `Context` en `:367`) | una pasada = un fichero |

`FromClaudeCodeLine` sigue siendo el punto único y conserva su firma, así que ningún test existente
cambia de forma. Una repetición dentro de la pasada devuelve `(nil, nil)`, como una línea no
facturable. Con la pasada a nil, la línea se emite: **misma** salida que hoy, salvo que el `event_id`
es determinista.

**Por qué no en `cmd/permea`**: habría que duplicar el conjunto en `generate()` y en `dryRun()`, y el
diagnóstico se separaría de la regla que cuenta.

**Por qué no en disco** (entre pasadas): FR-033 lo excluye. Además sería un estado nuevo que puede
corromperse o perderse, cuando la plataforma ya deduplica (M3).

### R3.3 · La cola tolera repetidos entre pasadas — comprobado en el código

`drain` marca confirmados **por `event_id`** y poda por `event_id` (`internal/transport/queue.go:89-117`).
Con dos copias del mismo `event_id` en la cola, confirmar una poda las dos. No se pierde nada, porque
son el mismo evento. **No hace falta tocar `internal/transport`.**

### R3.4 · Coste en memoria — **estimado, sin medir**

Una entrada por mensaje distinto de la pasada. Para la primera pasada de una instalación nueva, que lee
todo el historial, la referencia local son **10 698 mensajes** (M2). Con clave de 16 bytes y 4 enteros
de `usage`, del orden de 1–2 MB. Si Claude Code conservara meses de historial crecería en proporción.
**Sin medir**: se mide en B2.

---

## R4 · El diagnóstico local (FR-005, FR-006) — canal y forma

| Pregunta | Decisión | Evidencia |
|---|---|---|
| Canal | **stderr** | Es donde el agente ya escribe su relato de la pasada: «%d eventos encolados» (`main.go:286`) y, en el dry-run, «%d eventos generados» (`main.go:391`). En `--scan`, stdout son los eventos (`main.go:385-386`), y el resumen no puede mezclarse con ellos |
| Cuándo | Al **final de cada pasada**. En `--daemon`, sólo si la pasada leyó alguna línea | El daemon ya calla cuando no hay nada (`main.go:338-340`). Un resumen vacío cada 60 s es ruido, y el ruido enseña a ignorar los avisos |
| Qué | **Sólo recuentos**: líneas facturables leídas · eventos emitidos · líneas repetidas de un mismo mensaje · sintéticas · sin identificador (no contables) · con consumo distinto de la primera | FR-005 y FR-006 exigen que esas dos categorías sean visibles. Las otras dan contexto, para que «0 con consumo distinto» se lea sobre un total conocido |
| Qué NUNCA | Ningún identificador (ni `message.id`, ni `requestId`, ni `event_id`), ninguna ruta, ningún modelo | La frontera es hacia fuera, pero stderr acaba en logs de servicio y en capturas. No hay razón para estrenar un camino nuevo por el que salga un identificador |

Una línea por pasada. El formato lo fija la tarea; el contrato sólo fija categorías y canal.

---

## R5 · Quién consume el formato del dry-run

| Consumidor | Qué hace con `--scan` | Evidencia | ¿Le afecta cambiar el formato? |
|---|---|---|---|
| `cmd/permea/main_test.go`, subtest `--scan con "plain" presente → procesa sin parar` | Ejecuta `--scan` sobre una línea escrita en el test y **sólo compara el código de salida** | `main_test.go:294-308` | **No.** Pero su línea **no tiene identificadores** (`:297`): con FR-006 dejaría de producir evento y el test seguiría verde **sin ejercer nada**. Hay que añadírselos (R7) |
| `internal/project/resolve_test.go` | **No lo usa, y lo dice**: «NO se usa `--scan` para obtener las identidades. Ese camino usa un salt literal de dry-run» | `resolve_test.go:720-722` | No |
| Documentación de specs anteriores | 001 V2 lo usa a mano («2 eventos»). 004 y 005 lo describen o advierten de no usarlo para identidades | `specs/001-agente-inicial/quickstart.md:97`; `specs/005-adhesion-a-proyecto/research.md:217-238` | Artefactos fechados; no se tocan |
| `Makefile`, objetivo `run` | `go run ./cmd/permea --scan internal/ingest/testdata/claude_code_sample.jsonl` | `Makefile:16-17` | Sigue funcionando; sólo cambia lo que imprime |

**Decisión**: el formato de la línea `evento: …` **puede cambiar sin romper a nadie**. Se amplía con:
- las **cuatro partidas** (FR-010);
- el **`event_id`**, para que SC-003 y SC-021 se puedan medir sobre la copia congelada sin transmitir
  nada.

El `event_id` no depende de la sal y es un hash: imprimirlo en local no revela nada que no salga ya por
la frontera.

---

## R6 · El espejo de tarifas, y cómo se absorbe Q-006-1

### R6.1 · Lo que hay

- La tabla: `internal/pricing/pricing.go:14-18`, un literal Go de 3 claves, sin fuente ni fecha (`:1-2`).
- Su único test de valor fija la cifra errónea: `pricing_test.go:9-19`, `claude-opus-4-6` a 15/75.
- El catálogo de referencia: `permea-platform/backend/config/pricing.php@e50d0a5`, con 16 claves (M4 de
  la spec). *(Enmendado 2026-10-02, Q-006-1 resuelta: antes `865bba0`, también con 16 claves; entre los dos sólo cambia la
  fila de `claude-sonnet-5`.)*

### R6.2 · Decisión

| Pieza | Decisión |
|---|---|
| Tabla | El mismo literal Go, ahora con 16 claves. **Cabecera** con fuente, verificación, aprobación, la referencia `permea-dev/permea-platform · backend/config/pricing.php · e50d0a5`, el casamiento exacto (FR-018) y las **dos** limitaciones de FR-019. *(Enmendado 2026-10-02, coherencia con FR-018: se añade el casamiento entre los elementos de la cabecera.)* *(Enmendado 2026-10-02, Q-006-1 resuelta: decía `865bba0` y «las tres»; la de `claude-sonnet-5` desaparece.)* |
| Vigilancia | `pricing_test.go` con una **tabla esperada escrita aparte**, literal y clave a clave, con su propio comentario de procedencia. Tres aserciones independientes: (1) **recuento de claves = 16**; (2) cada clave esperada existe con sus **cuatro** cifras exactas; (3) **ninguna clave sobra**. Más el caso de coste a mano para `claude-opus-5-5` |
| Sin depender del otro repo | El test **no lee** `pricing.php`. Comparar contra el otro repositorio es una validación manual del quickstart (`git -C ../permea-platform show <commit>:backend/config/pricing.php`) |

**Alternativa descartada: generar la tabla Go desde el PHP.** Haría depender el build de un segundo
repositorio y de un parser de PHP, para 16 filas que cambian pocas veces. La duplicación es deliberada:
**dos copias que un test obliga a coincidir avisan de un cambio; una sola copia generada lo aplica en
silencio**.

### R6.3 · Q-006-1 (`claude-sonnet-5`) sin rehacer el plan

> **Resuelta el 2026-10-02, antes de B3.** El dueño decidió 2.00 / 10.00 / 2.50 / 0.20 y la plataforma lo
> corrigió en `e50d0a5`. Se aplicó el camino «decide antes de B3»: la spec se enmendó, B3 replica
> `e50d0a5`, y el punto 4 de abajo queda en **retirar** la tercera limitación.

El cambio de una fila toca exactamente **cinco sitios**, todos dentro de B3 o de la spec:
1. la fila en `pricing.go`;
2. la fila en la tabla esperada del test;
3. el commit replicado en la cabecera;
4. la tercera limitación de FR-019, en la cabecera;
5. M4, FR-014 y SC-009 en la spec (enmienda fechada).

**Ningún otro bloque depende de una cifra de tarifa**: B2 cuenta tokens y B4 y B5 son superficie.
- Si el dueño decide **antes** de B3, B3 replica el commit nuevo.
- Si decide **después**, la tarea «Q-006-1» de §Cierre son esos cinco sitios y la reverificación de
  SC-009/SC-010.

La puerta de salida es que, **antes de etiquetar**, la cabecera cite el commit vigente del catálogo.

---

## R7 · Los fixtures no traen identificadores — y con FR-006 dejarían de producir eventos

**Medido** (lectura de los JSONL del repositorio): ninguna línea `assistant` de ningún fixture trae
`message.id` ni `requestId`.

| Fichero | Líneas `assistant` | Consumidor que exige N eventos |
|---|---|---|
| `internal/ingest/testdata/claude_code_sample.jsonl` | 2 | `TestBoundary_NoDenylistLeaks` exige 2 (`boundary_test.go:99-101`). `TestSC009_RegresionCeroDelCaminoDeIngesta` exige `events_total = 2` del baseline (`baseline_regresion_test.go:164-166`; `specs/004-identidad-de-proyecto/baseline-sc004.tsv`, `# meta events_total 2`) |
| `internal/ingest/testdata/boundary_sample.jsonl` | 2 | `TestBoundary_TresCaminosHaciaElExterior` exige 2 (`boundary_test.go:142`) |
| Líneas escritas dentro de los tests | 4 en `boundary_test.go` (`:248`, `:276`, `:293`, `:311`) · 2 en `main_test.go` (`:50`, `:297`) | `TestAgentVersion_ReachesEvent` exige un evento (`main_test.go:55-57`); `TestBoundary_CostAvailable`, `TestBoundary_KeepsMetrics`… |
| `internal/project/testdata/*.jsonl` (20 líneas) | 20 | **Ningún test Go los lee** (`grep` sin resultados). Son entrada de validaciones manuales de 004 (`internal/project/testdata/README.md`, tabla de consumidores) |

**Decisión**: añadir `message.id` y `requestId` a las líneas `assistant` de los dos fixtures de
`internal/ingest/testdata/` y a las 6 líneas escritas en tests, en B1, **antes** de que el código las
exija.
- En `boundary_sample.jsonl` los valores son **centinelas** de la denylist (FR-013).
- `internal/project/testdata/` no se toca: nadie lo ejecuta en la suite.

**⚠️ Tensión declarada.** `claude_code_sample.jsonl` es la entrada de la línea base de 004, y su README
dice que «no se toca en toda la feature» (`internal/project/testdata/README.md`, §Lo que estos fixtures
NO son). Esa regla era de 004 y protegía **las tres columnas y el recuento**. Añadir dos campos que
ninguna de las tres columnas lee **no cambia el baseline**: las identidades se derivan de `cwd`,
`sessionId` y la máquina (`claudecode.go:86-88`), y con identificadores distintos el recuento sigue en
2. **El propio `TestSC009_…` lo demuestra**: si sigue verde tras el cambio, el baseline no se movió.
Se declara en vez de hacerse en silencio. Va a §Dudas del informe.

> **Decidido el 2026-10-02 (E-006-P1, orquestador).** Se edita en B1. Dos condiciones:
> - `TestSC009_RegresionCeroDelCaminoDeIngesta` sigue verde **sin tocar sus aserciones**;
> - `internal/project/testdata/README.md` (§Lo que estos fixtures NO son) anota la edición con fecha y
>   motivo.

---

## R8 · La CLI — por qué hoy miente y dónde se corrige

**Medido el 2026-10-02** (`report-agente.md` §10, sobre el binario de `0311fa1`):
- `-h` y `--help` imprimen la ayuda **por defecto del paquete `flag`** (sólo los 4 flags), por stderr,
  con exit 0;
- `help`, la invocación sin argumentos y un subcomando inexistente imprimen `printUsage` por stderr,
  con exit 0;
- `status -h` ejecuta `status`;
- `project join -h` toma `-h` como código y puede llegar a emitir.

| Hecho | Evidencia |
|---|---|
| La escalera de subcomandos va **antes** de `flag.Parse` | `cmd/permea/main.go:44-63`, y `:65-69` |
| El banner se imprime **después** del parseo y **antes** del `switch` | `main.go:78`. Por eso `help` y la invocación sin argumentos llevan banner y `-h` no |
| La ayuda completa es un **literal** que se escribe a mano | `main.go:104-137` |
| `flag.Usage` no se redefine | No hay asignación a `flag.Usage` en el repositorio (`grep`) |
| `-h` en Go sale con 0 tras `flag.Usage` | Medido: exit 0 (`report-agente.md` §10) |
| `runStatus` crea el directorio de datos en su **primera** línea | `status.go:17` → `config.DataDir()`, que hace `MkdirAll` (`internal/config/config.go:54-64`) |
| `enroll` toma cualquier primer argumento distinto de `-` como secreto | `enroll.go:120-121` |
| `project join` toma cualquier primer argumento distinto de `-` como código | `project.go:328-329` |
| El código de fallo existente es 1 | `project.go:55-56` (`codigoExito = 0`, `codigoFallo = 1`) |

**Decisión** (contrato en [`contracts/cli.md`](./contracts/cli.md)):
1. **Una sola fuente del texto.** Un fichero nuevo en `cmd/permea` con una tabla de subcomandos (nombre,
   sinopsis, líneas de ayuda) y de opciones. La ayuda general se **compone** de la tabla entera, y la de
   cada subcomando de su fila. No hay dos textos que puedan divergir: la ayuda de un subcomando es un
   fragmento de la general.
2. **La ayuda se decide antes que nada.** En la escalera, **antes** de `flag.Parse` y **antes** del
   banner, se atienden `help`, `-h`, `--help`, `-help` y la invocación sin argumentos: ayuda por stdout,
   exit 0.
3. **`flag.Usage` sirve la misma ayuda por stdout.** Cubre combinaciones como `permea --run -h`: Go
   atiende `-h` durante el parseo, antes de ejecutar nada. **El mensaje de error de un flag desconocido
   sigue en stderr**: lo escribe `flag` en su salida, que no se toca.
   > **Enmendado el 2026-10-02 (E-006-P3).** Ante una opción desconocida o sin valor ya **no** se
   > escribe nada por stdout. El error, por stderr, nombra la opción y remite a `permea help`, y la
   > salida es 2. Consecuencia de mecanismo: con el `FlagSet` por defecto (`ExitOnError`), Go escribe su
   > propio mensaje **y** el uso, y sale él mismo. Para controlar canal y texto, el parseo pasa a un
   > `FlagSet` con `ContinueOnError` y salida descartada:
   > - `flag.ErrHelp` → la ayuda por stdout, exit 0;
   > - cualquier otro error → el mensaje propio por stderr, exit 2.
   >
   > Comprobado en la librería estándar de Go 1.22: `ExitOnError` sale con 0 ante `ErrHelp` y con 2 ante
   > los demás (medido para `-h`: exit 0, `report-agente.md` §10). **Sin comprobar** el texto exacto
   > del error que devuelve `Parse` para una opción sin valor: la tarea lo mide.
4. **Las ayudas de subcomando, en la primera línea de cada subcomando**, antes de leer stdin, antes de
   `config.DataDir()` y antes de cualquier rehúse:
   - en `runEnroll`, antes de `os.Stdin.Stat()`;
   - en `runStatus`, que pasa a recibir los argumentos;
   - en `runProject` (`project -h`) y en `projectJoin` (`project join -h`).
5. **Subcomando inexistente.** Si el primer argumento no empieza por `-` y no es un subcomando
   conocido, se escribe un error que lo nombra, por stderr, con exit 1. Si empieza por `pmea2.`,
   `pmeaj1.` o `pmea1.`, **no se reproduce** (FR-022, D-006-9).

**Fuera de la spec, y se conserva tal cual**: los argumentos sobrantes tras los flags
(`permea --run extra`) siguen ignorándose, como hoy (`main.go:65-69`: `flag.Parse` se detiene en el
primer posicional). Va a §Dudas. *(Confirmado el 2026-10-02, E-006-P3: se quedan como están.)*

---

## R9 · Los 7 avisos de `golangci-lint` — uno por uno (D-006-13)

Medido el 2026-10-02 con `golangci-lint` 2.12.2 y la configuración `.golangci.yml` del repositorio.

> **Enmendado el 2026-10-02 (E-006-P7).** La tabla está **truncada**: se midió con el tope por defecto
> `max-same-issues: 3`. Sin el tope (`golangci-lint run --max-same-issues 0 --max-issues-per-linter 0`,
> medido el 2026-10-02 al ejecutar B0) salen **15**:
> - **12** `errcheck` en `cmd/permea/project.go`: las cuatro de la fila 1–4 y **8 ocultas**, todas
>   escrituras `fmt.Fprint*` al `stderr`/`stdout` de `projectJoin`;
> - los 2 `revive` y el `staticcheck` de las filas 5–7.
>
> La de `os.Stderr` en `runProjectOS` no sale porque `errcheck` la excluye por defecto. **Decisión del
> orquestador**: el tope se quita en `.golangci.yml` (`issues.max-same-issues: 0`,
> `issues.max-issues-per-linter: 0`), y la corrección de la fila 1–4 se extiende a las 12 escrituras,
> con el mismo idioma.

| # | Aviso | Dónde | Corrección prevista | Alternativa descartada |
|---|---|---|---|---|
| 1–4 *(12 reales, E-006-P7)* | `errcheck`: valor de error de `fmt.Fprintln`/`Fprintf` sin comprobar | `cmd/permea/project.go:93`, `:103`, `:129`, `:139` | `_, _ = fmt.Fprintln(…)`, el idioma que el repositorio ya usa para escrituras de diagnóstico (`status.go:42`, `:53`, `:55`; `enroll.go:111`) | `//nolint:errcheck`: el repositorio no tiene ninguno (`grep -rn nolint` vacío) y el idioma explícito dice lo mismo sin silenciar el linter |
| 5 | `revive` `error-return`: «error should be the last type» | `internal/config/endpoint.go:83`, `JuzgarEndpoint(endpoint) (errAnalisis error, admisible bool)` | **Reordenar a `(admisible bool, errAnalisis error)`**. Llamantes: `enrollment.go:82`, `config.go:100`, `transport.go:148`, `:256`, y 5 sitios de `endpoint_test.go`. El compilador los encuentra todos | `//nolint:revive`. La razón del orden actual («por qué devuelve el error y no un segundo booleano», `endpoint.go:69-72`) habla de **devolver** el error, no de su **posición**: el orden no tiene justificación escrita |
| 6 | `revive` `unused-parameter`: `r` sin usar | `internal/transport/adhesion_test.go:267` | Renombrar a `_` | — |
| 7 | `staticcheck` SA1007: URL inválida en constante | `internal/transport/adhesion_test.go:194`, `url.Parse(endpointNoAnalizableAdhesion)`; la constante en `:112` | **`//nolint:staticcheck` en esa línea, con el motivo**: la URL inválida es **el sujeto** del test (la causa de `url.Parse` que `Adherir` debe conservar). Será la **única** directiva `nolint` del repositorio, y se declara | Construir la cadena en tiempo de ejecución para que staticcheck no la vea: **engaña al instrumento en vez de declarar la excepción**. Un auditor leería una construcción rara sin saber por qué |

**Por qué el linter va primero** (plan, bloque B0): la constitución pide `golangci-lint` limpio **para
cerrar cualquier tarea** (`.specify/memory/constitution.md:72-75`). Con 7 avisos heredados, ninguna
tarea de B1 a B5 podría cerrarse limpia. Además, los avisos 1–4 están en `project.go`, que B4 vuelve a
editar: corregirlos primero evita mezclar los dos cambios en el mismo diff.

**#7 aceptado el 2026-10-02 (E-006-P4, orquestador).**

**Cuestión abierta**: el resultado depende de la versión del linter (2.12.2 local; en CI no se ejecuta,
D-006-6). El quickstart fija la versión con la que se mide el 0.

---

## R10 · Publicación — lo que el procedimiento vigente ya hace y lo que no

| Hecho | Evidencia |
|---|---|
| La release se dispara con `push` de `v*.*.*` | `.github/workflows/release.yml:6-9` |
| El workflow **no** ejecuta tests ni linter | `release.yml:14-38`; D-006-6 |
| El snapshot local construye los 5 archivos sin publicar | `.goreleaser.yaml:3`; `specs/002-distribucion/quickstart.md` V2 |
| Versión de un snapshot | El de `v0.1.0` se estampó `0.1.0-SNAPSHOT-35f8ba2` (`dist/metadata.json`). Con `v0.2.1` como última etiqueta, el de 006 llevará **otra versión que no es `0.3.0`**: probablemente `0.2.2-SNAPSHOT-<sha>`, por la plantilla por defecto de GoReleaser v2. **Sin comprobar**. SC-022 sólo exige «no `0.0.1-dev`» |
| Ajustar la plantilla del snapshot cambiaría configuración | Lo prohíbe FR-029: el diff de `.goreleaser.yaml` sólo puede tocar comentarios |
| Las etiquetas existentes son **anotadas** | `git ls-remote --tags origin` → `v0.2.0^{}`, `v0.2.1^{}` (2026-10-02) |
| Las dos PR anteriores se fusionaron con **merge commit** | `0311fa1 Merge pull request #2`, `158fb1b Merge pull request #1` |
| El tap y el bucket existen bajo `permea-dev` | `gh api repos/permea-dev/scoop-permea/contents/` → `permea.json`; `…/homebrew-permea/contents/` → `Casks` (2026-10-02) |
| Los repositorios `bfgnet/…` del README | `gh api` → **404** (2026-10-02) |

> **Enmendado el 2026-10-02 (C3, T077).** La versión del snapshot de 006 fue
> **`0.2.1-SNAPSHOT-b504268`** (`dist/metadata.json`: `tag` v0.2.1), no `0.2.2-…`: GoReleaser v2.16.0
> usa la última etiqueta tal cual. La conclusión no cambia: **no es `0.3.0`** ni `0.0.1-dev`, que es
> lo único que exige SC-022.

**Decisión**: el orden del cierre es el de §Cierre en [`plan.md`](./plan.md):
1. puertas locales;
2. snapshot;
3. **ensayo en Windows sobre el snapshot (FR-034)**;
4. PR y fusión;
5. etiqueta anotada sobre el commit de fusión;
6. verificación de canales;
7. ensayo final (FR-032).

---

## R11 · Lo que se descartó, y por qué

| Descartado | Motivo |
|---|---|
| Deduplicar entre pasadas con un fichero de `event_id` vistos | FR-033 lo excluye. Sería un estado nuevo que se corrompe o se pierde, cuando la plataforma ya deduplica (M3) |
| Hacer obligatoria la pasada (sin nil válido) | Obligaría a cambiar las 8 llamadas a `FromClaudeCodeLine` que hay en los tests (R7) **sin aportar garantía**: los dos caminos de producción la instancian, y un testigo de proceso lo comprueba (SC-001 por `--scan`, SC-021 por `generate`). Es el mismo patrón que el `Resolutor` |
| Sumar las líneas de un mismo mensaje | Prohibido por FR-005: con `usage` idéntico (M2) es precisamente el defecto ×2,13 |
| Mostrar el diagnóstico por stdout en `--run` | stdout de `--run` está vacío hoy y nadie lo lee. stderr es el relato de la pasada (R4) |
| Un `nolint` general en `.golangci.yml` para `revive` | Silenciaría la regla en todo el repositorio por un solo caso que tiene corrección limpia (R9 #5) |
