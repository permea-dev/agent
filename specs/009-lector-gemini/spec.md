# 009 · «Lector de Gemini CLI» — Especificación

**Feature Branch**: `009-lector-gemini` · **Created**: 2026-10-08 · **Status**: **Ratificada** el 2026-10-08, 12:20 (Madrid), salvo lo que diga «pendiente» · **Cabeza de
partida**: `15ce93b` *(sobre `c89c5de`, 0.5.0)*
**Input**: las decisiones del dueño `D-1`…`D-4` del 2026-10-08 · el descubrimiento del 08-10 *(`soporte/descubrimiento.md`, Q1–Q13, sobre la
**copia congelada** de una sesión de Gemini CLI 0.63.0)* · el paquete `@google/gemini-cli` 0.63.0 · el contrato nuevo
`contracts/event-id-gemini.md`.

> ⛔ Toda afirmación sobre nuestro código lleva `fichero:línea`, medida sobre `15ce93b`. Las de Gemini CLI citan la pregunta del
> descubrimiento *(Qn)*, que da la ruta bajo `packages/` y la función. Las propuestas son decisiones **✅** desde la ratificación (P-9, **❌**
> rechazada), y conservan sus alternativas como rastro. Lo que añade la spec sin que nadie lo dijera va marcado **`DECIDÍ YO`**.

## Contexto — el agente lee Claude Code y Codex

1. **Dos lectores hoy.** Claude Code, en `~/.claude/projects` *(`internal/config/config.go:112-121`)*. Codex, en `$CODEX_HOME/sessions` o
   `~/.codex/sessions` *(`internal/config/codex.go:18-27`)*. Los dos se leen en `generate()`, antes del **único** `st.Save`
   *(`cmd/permea/main.go:332-345`)*.
2. **Gemini CLI guarda su consumo en local**, un JSONL por sesión en `~/.gemini/tmp/<slug>/chats/`. Cada respuesta del modelo es un
   mensaje `type: "gemini"` con un objeto `tokens` de seis partidas, copiadas de `usageMetadata` *(Q2)*. Sobre la copia, los tokens del
   papel `main` cuadran al token con `/stats` *(descubrimiento §Contraste)*.
3. **El fichero no es un registro limpio.** La misma respuesta aparece varias veces: 16 apariciones para 10 respuestas en la copia. Si se
   reconstruye el estado final, se pierden 7 de 10, porque el `$set.messages` reescribe el historial *(Q4)*. Se lee como bitácora, y
   **nunca** se reconstruye el estado final.
4. **La plataforma ya acepta el evento tal cual.** `tool` es texto libre, y un modelo sin tarifa llega con el coste «ciego y
   señalizado» *(008 §Contexto 3)*. Un `event_id` repetido se descarta *(`permea-platform`, índice único `(org_id, event_id)`:
   `backend/database/migrations/2026_07_04_000004_add_unique_org_event_to_metric_events.php:16`; `on conflict … do nothing`:
   `backend/app/Ingest/IngestBatchService.php:92`)*.

## Decisiones del dueño *(ratificadas el 2026-10-08; no se discuten aquí)*

| # | Decisión |
|---|---|
| **D-1** · 08-10 | **Fuente**: sólo el fichero de sesión (`chats/*.jsonl` y el nivel de subagentes). **Límite declarado**: las llamadas auxiliares (`utility_*`) y los intentos fallidos no están en él, y el lector queda por debajo de la factura. En la copia faltan 5 396 de 113 284 de entrada y 1 016 de 1 492 de salida, según `/stats`. La telemetría a fichero *(Q13)* queda **fuera** y va a ficha |
| **D-2** · 08-10 | **Partidas**: `tokens_input = input − cached + tool` · `tokens_cache_read = cached` · `tokens_cache_creation = 0` · `tokens_output = output + thoughts`. Un campo que falte vale 0 |
| **D-3** · 08-10 | **Formatos**: se lee el JSONL (≥ 0.39). Los `session-*.json` de sesión (≤ 0.38) se **cuentan** en el resumen y **no** se emiten. `logs.json` no es una sesión |
| **D-4** · 08-10 | **Tarifa**: `cost_usd = 0` y `cost_available = false`, como la D-3 de la 008. No hay tabla de Gemini en el agente |

## Decisiones de método *(heredadas de 008; las precisa la medida)*

| # | Decisión | Lo que precisa la medida |
|---|---|---|
| **M-1** | `tool = "gemini"`. El evento no cambia: mismos campos y `SchemaVersion = 1`. `internal/event/`, `eventid.go`, `eventid_test.go`, `codex_eventid.go`, `codex_eventid_test.go`, `boundary_test.go` y los contratos de 006 y 008, con **0 bytes** de diff | — |
| **M-2** | **Un evento por respuesta**: la primera aparición, en la pasada, de cada `id` con `tokens`. **No hay espera ni cierre** | La CLI escribe los tokens al terminar la respuesta *(Q4: `processStreamResponse`)*. Un intento fallido no deja tokens *(D-1)* |
| **M-3** | `event_id` = `["permea/event_id/v1", "gemini", "respuesta", <id>]`, en un fichero y una función nuevos *(contrato; P-2)* | 10 `id` distintos en 16 apariciones, y 0 con tokens distintos *(Q4)* |
| **M-4** | Las partidas, según D-2 | **Medido**: `total = input + output + thoughts + tool` en las 15 líneas de mensaje con tokens *(Q3)*. La caché va dentro de `input`, y el razonamiento fuera de `output` |
| **M-5** | El modelo, el `model` del propio mensaje | Cambia **dentro de la sesión** *(Q5)* |
| **M-6** | `occurred_at` = el `timestamp` del mensaje. `project_ref`, de `.project_root` con el resolutor existente. `dev_id`, `machine_ref`, `org_id` y `agent_version`, como hoy | En la copia, `timestamp` está en UTC y es igual en todas las repeticiones de un `id` *(Q7)* |
| **M-7** | Dónde lee: `$GEMINI_CLI_HOME/.gemini` si la variable está y no es vacía; si no, `~/.gemini`. Estado en el mismo `state.json`, con sus cuatro campos; **el `mtime` no decide nada** | Es la regla de la CLI, que trata la variable vacía como ausente *(Q1: `homedir`)*. Es la **segunda** lectura de entorno en producción, tras `CODEX_HOME` *(`internal/config/codex.go:19`)* |
| **M-8** | Versión **`0.6.0`**. Claude Code y Codex se leen exactamente como en la 0.5.0: los **574** tests actuales siguen verdes **sin tocarlos** | — |
| **M-9** | `internal/testutil/sandbox.go` fija también `GEMINI_CLI_HOME` *(vacía)*, junto a `CODEX_HOME` *(`sandbox.go:62-63`)* | Sin ella, un test de proceso leería el `.gemini` del desarrollador |

## Propuestas — **ratificadas** por el dueño el 2026-10-08, 12:20 (Madrid)

| # | Propuesta | Alternativas *(rastro)* | Recomendación *(rastro)* |
|---|---|---|---|
| **P-1** ✅ **(a)** | **El lector se activa solo** si `<raíz>/tmp` existe y es un directorio, comprobado **en cada pasada**. No hay interruptor | (b) un interruptor | **(a)**, como 008 P-1 |
| **P-2** ✅ **(a)** | **`event_id` sólo con el `id`** del mensaje *(contrato §Por qué basta el `id`)* | (b) `sessionId` + `id` | **(a)**: el `id` es un UUID v4 de la CLI y las copias conservan el `sessionId`, así que (b) no separa nada; si un día se copiaran respuestas a otra sesión, (b) las contaría dos veces; y (b) obliga a leer la cabecera en cada pasada |
| **P-3** ✅ **(a)** | **Una respuesta que ya apareció antes del offset en el mismo fichero no se reemite** en una pasada posterior: es «repetida» | (b) reemitirla y dejar que la plataforma la descarte | **(a)**: cada compresión o reanudación repite **todas** las respuestas vivas en su `$set.messages`. Con (b), reanudar una sesión larga reencolaría toda su historia, y «eventos» engañaría al usuario. Cuesta lo mismo que el contexto de Codex *(SC-013)* |
| **P-4** ✅ **(a)** | **El contexto entre pasadas, por resultado** *(FR-005)*. El mecanismo lo decide el plan, con un tope de coste | (b) guardar contexto en `state.json` *(cambia sus cuatro campos)* | Releer el prefijo sin emitir, como `leerPrefijoCodex` *(`internal/ingest/codex_contexto.go:212`)*; el plan elige |
| **P-5** ✅ **(a)** | **Un `total` descuadrado** *(≠ `input + output + thoughts + tool`, con `total` presente)* **se emite** y se cuenta como «total descuadrado», un subconjunto de «eventos» | (b) emitir sin contar · (c) «incoherente», sin emitir | **(a)**: D-2 no usa `total`, así que las cuatro partidas siguen siendo ciertas y descartar perdería consumo real. Contarlo hace visible una partida nueva de la API que el lector no conozca. 0 casos en la copia |
| **P-6** ✅ **(a)** | **Una respuesta sin `model`** *(ausente o vacío)* se emite con `model` vacío y se cuenta como «sin modelo» | (b) no emitir | **(a)**, como 008 P-4. `recordSyntheticMessage` registra mensajes `gemini` sin modelo que pueden llevarse los tokens pendientes *(Q4)* |
| **P-7** ✅ **(a)** | **La misma respuesta en dos carpetas** *(migración, Q1)*: el evento lleva el `project_ref` de la que tenga `.project_root` | (b) el de la primera que se lea | **(a)**: las carpetas antiguas por hash no tienen `.project_root`, y (b) mandaría un evento sin proyecto que la plataforma ya no corregiría |
| **P-8** ✅ **(a)** | **`--scan`** reconoce una sesión de Gemini por su primera línea: un objeto con `sessionId` y `projectHash`, y sin `type`. Cada evento sale con la línea `evento:` de Codex, y el resumen con la línea de Gemini | (b) `--scan-gemini` | **(a)**, como 008 P-8 |
| **P-9** ❌ **(b)**, rechazada | **El agente no resuelve enlaces simbólicos**: una raíz de sesiones que sea un enlace no se garantiza ni se prueba *(N-11)*. SC-001 mide sobre la segunda copia congelada | (a) *(propuesta, rastro)*: resolver `<raíz>/tmp` si es un enlace, una vez por pasada | **(a)**, para medir `--run` sobre la copia sin copiarla. El dueño eligió **(b)**, con la segunda copia |
| **P-10** ✅ **(a)** | **Los textos**: §Textos propuestos | corregirlos | — |
| **P-11** ✅ **(a)** *(textos aprobados en E-1)* | **La coherencia del README** *(como la E-5 de 008)*: las frases que hoy dicen «Claude Code y Codex CLI» pasan a nombrar también Gemini CLI, con textos que se aprueban en el encargo de B6 | (b) dejarlas | **(a)**: si no, quedarían incompletas con la 0.6.0 |

## User Scenarios & Testing

### Historia 1 — El consumo de Gemini llega a la plataforma *(P1)*
1. **Dado** una sesión de Gemini CLI ≥ 0.39 con dos respuestas, **cuando** corre `--run`, **entonces** salen **dos** eventos `tool =
   gemini`, con sus tokens según D-2, `cost_available = false` y el modelo de cada mensaje.
2. **Dado** el mismo fichero en una segunda pasada sin cambios, **entonces** no sale **ningún** evento.

### Historia 2 — Las repeticiones no inflan nada *(P1)*
1. **Dado** una respuesta que llama a una herramienta *(dos líneas con el mismo `id`)* y una reanudación que la repite en
   `$set.messages`, **entonces** sale **un** evento, en esa pasada y en las siguientes.

### Historia 3 — Quien no usa Gemini no nota nada *(P1)*
1. **Dado** una instalación sin `~/.gemini/tmp` ni `GEMINI_CLI_HOME`, **entonces** la salida de `--run` es **byte a byte** la de la 0.5.0, y
   no hay eventos `gemini`.

### Historia 4 — Gemini sin Claude Code ni Codex *(P2)*
1. **Dado** una instalación con `~/.gemini/tmp` y sin las otras dos raíces, **entonces** los eventos de Gemini salen igual.

### Edge Cases
- **Llamada a herramienta** *(Q4, 5 casos en la copia)*: `recordToolCalls` vuelve a escribir la respuesta con los mismos `tokens`. Es
  «repetida».
- **Tokens que llegan después que el mensaje** *(Q4, `recordMessageTokens`; 0 casos)*: la primera aparición no tiene `tokens`. Cuenta la
  **primera aparición con `tokens`**.
- **Compresión** *(Q8, 1 caso)*: su `$set.messages` cambia el historial por un resumen **sin** tokens. Los mensajes del resumen no son
  respuestas, y lo emitido antes no se toca.
- **Reanudación**, en la misma pasada o días después *(Q8, 1 caso)*: añade al mismo fichero, y su `$set.messages` repite las respuestas
  vivas **con** tokens. Son «repetidas» *(P-3)*.
- **Rebobinado** *(`$rewindTo`; 0 casos)*: no borra lo emitido. Esos tokens se gastaron.
- **`tokens: null`** o ausente: no es una respuesta medida. No se emite ni se cuenta.
- **Cambio de modelo** *(Q5, 1 caso)*: cada evento lleva el `model` de su mensaje.
- **Subagentes** *(sin muestra; Q1)*: `chats/<sessionId del padre>/<sessionId>.jsonl`. Sus respuestas se emiten, y su `.project_root`
  es el de la carpeta `<slug>` que contiene `chats/`.
- **La misma sesión en dos carpetas** *(migración por hash, Q1; sin muestra)*: dan el mismo `event_id`. Dentro de la pasada, una es
  «repetida» *(P-7)*. Entre pasadas, la descarta la plataforma.
- **Sesión `.json` convertida al reanudar** *(Q10; sin muestra)*: el `.json` se cuenta como «formato anterior», y el `.jsonl` nuevo trae
  todas las respuestas antiguas con sus tokens, que se emiten con su `timestamp` original.
- **Fichero reescrito entero** *(la CLI no pudo releerlo, Q4)*: queda una copia `<fichero>.unreadable-<ms>`, que no se lee *(no acaba en
  `.jsonl`)*. Si el nuevo es más corto que el offset, se relee desde 0 *(`internal/state/state.go:113`)*, con los mismos `event_id`.
- **Temporales de la CLI** (`*.jsonl.tmp-<pid>`): no acaban en `.jsonl` y no se leen.
- **Retención** *(Q9)*: la CLI borra por defecto las sesiones con más de 30 días. Lo que no se leyó antes se pierde *(Límites)*.
- **Respuesta rechazada y reintentada** *(`InvalidStreamError`, Q4)* y **`queuedTokens` pisado**: la CLI no escribe esos tokens, así que no
  hay evento *(D-1)*.
- **Carpeta ausente**, `GEMINI_CLI_HOME` vacía o apuntando a algo que no existe: el agente no dice nada, no falla y no escribe la línea de
  Gemini.
- **Línea parcial** *(sesión en curso)*: no se consume *(`state.go:130-133`)*.
- **Línea larga**: el `$set.messages` lleva el historial entero en **una** línea. En la copia, la mayor mide 4 703 B. `Recorrer` no tiene
  tope *(`state.go:126-129`)*, y `--scan` corta en 1 MiB *(008 D-008-P8)*.
- **Fichero ilegible** o **línea corrupta**: como en Codex *(FR-025, FR-026)*.
- **El proceso cae entre encolar y guardar**: la pasada siguiente reencola los mismos `event_id`, y la plataforma los descarta.

## Requirements

### A · Dónde y cuándo lee *(M-7, P-1)*

- **FR-001** *(raíz)*: la raíz de Gemini **DEBE** ser `$GEMINI_CLI_HOME/.gemini` si `GEMINI_CLI_HOME` está definida y no es vacía, y si
  no, `<directorio personal>/.gemini`, con el mismo `os.UserHomeDir()` que las otras dos raíces. Las sesiones están bajo `<raíz>/tmp`.
- **FR-002** *(activación)*: el lector está activo **en una pasada** si y sólo si, en esa pasada, `<raíz>/tmp` existe y es un directorio,
  comprobado en **cada** `generate()`; la ruta se resuelve una vez por proceso. Si no existe, el lector no escribe nada, no devuelve error y no cambia ningún texto.
- **FR-003** *(ficheros)*: se leen los `.jsonl` de `<raíz>/tmp/*/chats/` y de `<raíz>/tmp/*/chats/*/` *(subagentes)*, y ninguno más. Los
  `.json` de esas mismas carpetas se cuentan *(FR-018)*. Nada fuera de `chats/` se lee: ni `logs.json`, ni `logs/`, ni `.project_root`
  como sesión.
- **FR-004** *(estado)*: el estado de cada fichero vive en el mismo `state.json`, con sus **cuatro** campos y la ruta como clave, como en
  Codex *(`internal/state/state.go:16-21`, `:105-153`)*. **El `mtime` no decide nada.**
- **FR-005** *(contexto, por resultado — P-3, P-4)*: una respuesta leída en una pasada posterior **DEBE** salir con el **mismo** `event_id`,
  modelo, `project_ref` y `session_ref` que si el fichero se hubiera leído entero de una vez. **Y DEBE NO salir** si su `id` ya apareció
  con `tokens` antes del offset del mismo fichero *(P-3 (a))*. `state.json` conserva sus cuatro campos. El coste está acotado por
  SC-013.

### B · El evento *(M-2 a M-6, D-2, D-4)*

- **FR-006** *(unidad — M-2)*: es **respuesta** cada aparición de un mensaje `type: "gemini"` con un objeto `tokens`, ya venga en una
  línea de mensaje o como elemento de `$set.messages`. Cada respuesta se clasifica *(FR-019)*, y produce **como mucho un** evento. Nada
  más produce eventos. **Nunca** se reconstruye el estado de la conversación.
- **FR-007** *(identidad — M-3)*: su `event_id` es el de `contracts/event-id-gemini.md`, y **DEBE** reproducir sus vectores byte a byte.
- **FR-008** *(repetidas — P-3)*: un `event_id` ya emitido en la pasada, o un `id` ya aparecido con `tokens` antes en el mismo fichero, no
  se emite y se cuenta como «repetida».
- **FR-009** *(sin identificador)*: una respuesta sin `id`, con `id` vacío o no textual, no se emite y se cuenta como «sin identificador».
- **FR-010** *(partidas — D-2)*: con las de `tokens`:
  - `tokens_input` = `input − cached + tool`;
  - `tokens_cache_read` = `cached`;
  - `tokens_cache_creation` = `0`;
  - `tokens_output` = `output + thoughts`.

  Una partida ausente vale 0. `total` sólo se usa para el recuento de FR-019 *(P-5)*.
- **FR-011** *(modelo — M-5, P-6)*: `model` es el del mensaje, tal cual. Si falta o está vacío, el evento sale con `model` vacío y se cuenta
  como «sin modelo».
- **FR-012** *(momento — M-6)*: `occurred_at` es el `timestamp` del mensaje *(también si viene en un `$set.messages`)*.
- **FR-013** *(herramienta y coste — M-1, D-4)*: `tool = "gemini"`, `cost_usd = 0` y `cost_available = false`, **siempre**. El lector no
  consulta `internal/pricing`.
- **FR-014** *(proyecto — M-6, P-7)*: `project_ref` = `Resolutor.Derivar(r, sal)` *(`internal/project/resolve.go:318`)*. `r` es el texto
  de `<raíz>/tmp/<slug>/.project_root`, sin espacios ni salto de línea al final: el de la carpeta que contiene `chats/`, también para los
  subagentes. Sin `.project_root`, o si está vacío, `project_ref` sale vacío, como hoy en Claude Code *(008 E-4)*. **Nunca** se usa el
  nombre de la carpeta *(se repite, Q1)* ni `projectHash` *(no se puede invertir, Q7)*.
- **FR-015** *(referencias)*: `session_ref` = `event.Ref(sal, sessionId)` *(`internal/event/event.go:40-46`)*, con el `sessionId` de la
  cabecera del fichero *(su primera línea)*. Sin cabecera legible, `session_ref` sale vacío. `machine_ref`, `dev_id`, `org_id` y
  `agent_version`, como en Claude Code.
- **FR-016** *(un fichero no es varias sesiones)* `DECIDÍ YO`: los `$set {sessionId}` de la compresión y la reanudación traen el mismo
  valor *(Q8)*, y no cambian `session_ref`.
- **FR-017** *(nada del proveedor viaja)*: `id`, `sessionId`, `projectHash` y el texto de `.project_root` **NUNCA** se copian a ningún
  campo del evento, a la cola ni a nada que se transmita. `sessionId` y `.project_root` sólo viajan derivados con sal. **Excepción
  declarada**: el nombre del fichero lleva 8 caracteres del `sessionId`, y la carpeta de un subagente lleva el del padre. La ruta es la
  clave de `state.json`: queda **en local**, como en 008 FR-016.

### C · Formatos *(D-3)*

- **FR-018** *(formato anterior)*: cada `.json` de las carpetas de FR-003 se cuenta como «fichero en formato anterior» **una vez**, en la
  pasada en que aparece, y otra vez sólo si cambia, como los `.zst` de Codex *(`internal/ingest/codex_contexto.go:235-246`)*. No se abre
  y no produce eventos.

### D · Clasificación y lo que ve el usuario *(P-5, P-8, P-10)*

- **FR-019** *(la cuenta del resumen)*. En cada pasada, `respuestas = eventos + repetidas + sin identificador + incoherentes`. «Sin
  modelo» y «total descuadrado» son **subconjuntos** de «eventos».
  - **Incoherente** *(como 008 E-3/E-4)*: se cuenta y no se emite cuando `tokens` no es un objeto; alguna de `input`, `output`, `cached`,
    `thoughts` o `tool` no es numérica, o es negativa; `cached > input`; o el `timestamp` falta o está mal formado.
  - **Total descuadrado** *(P-5)*: `total` está y es distinto de `input + output + thoughts + tool`. Se emite.
  - **Orden**: sin identificador → incoherente → repetida → evento.
- **FR-020** *(corrupta)*: es corrupta **sólo** la línea cuya envoltura no es JSON válido. Si está en la parte nueva, se salta con el aviso
  de hoy, `skip (línea corrupta): …` *(`cmd/permea/main.go:295`)*; si está en el prefijo releído, se ignora en silencio. No cuenta en
  «respuestas» y no corta el fichero.
- **FR-021** *(resumen)*: con el lector activo, `--run` escribe **una** línea más en stderr, el texto aprobado, tras la de Codex *(o, si no
  la hay, tras el resumen de Claude Code y su aviso; `main.go:406-413`)*. El demonio la escribe sólo si en ese ciclo hubo respuestas o
  ficheros en formato anterior, como `HayNovedades` *(`codex_contexto.go:251-253`; `main.go:474-475`)*.
- **FR-022** *(`--scan` — P-8)*: un fichero cuya primera línea es la cabecera de Gemini *(objeto con `sessionId` y `projectHash`, sin
  `type`)* se lee como Gemini, con la línea `evento:` de Codex y el resumen de Gemini.
- **FR-023** *(los textos de antes no cambian)*: sin raíz de Gemini, la salida de `--run`, del demonio y de `--scan` es **byte a byte** la
  de la 0.5.0. Con raíz, sólo se añade la línea de FR-021, y «N eventos encolados» cuenta también los de Gemini.

### E · Frontera, durabilidad y errores

- **FR-024** *(orden de escritura)*: Gemini se lee **después** de Codex y **antes** del único `st.Save` *(`main.go:332-345`)*. Sus eventos
  se encolan *(`transport.Append`)* antes de guardar el estado. Si el proceso cae entre medias, la pasada siguiente los reencola con el
  mismo `event_id`.
- **FR-025** *(Gemini no rompe a los demás)*: un error al **leer** un fichero de Gemini *(stat, prefijo, cabecera, `.project_root` o
  `Recorrer`)* **DEBE**:
  - escribir en stderr el texto aprobado `gemini: fichero omitido: %v`;
  - **no** avanzar el offset de ese fichero;
  - **no** cortar la pasada;
  - **no** impedir guardar el estado de los demás, los de Claude Code y Codex incluidos.

  Un `.project_root` que **no existe** no es un error *(FR-014)*. Un fallo al **encolar** sigue siendo fatal, como hoy *(008 FR-028)*.
- **FR-026** *(sin las otras herramientas)*: una instalación con raíz de Gemini y sin las de Claude Code y Codex **DEBE** enviar los eventos
  de Gemini.
- **FR-027** *(frontera — M-1)*: `git diff 15ce93b` de `internal/event/`, `internal/ingest/eventid.go`, `eventid_test.go`,
  `codex_eventid.go`, `codex_eventid_test.go`, `boundary_test.go` y los contratos de 006 y 008, **vacío**.
- **FR-028** *(versión y documentos — M-8, P-10, P-11)*: `0.6.0`. Claude Code y Codex se leen exactamente como en la 0.5.0. El README gana
  la sección «### Gemini CLI», y el CHANGELOG la entrada `0.6.0`, con los textos aprobados.

### Key Entities
- **Respuesta de Gemini**: una aparición de un mensaje `gemini` con `id`, `timestamp`, `model` y `tokens` *(`input`, `output`, `cached`,
  `thoughts`, `tool`, `total`)*.
- **Contexto de fichero**: el `sessionId` de la cabecera, el texto de `.project_root` de su carpeta y los `id` ya vistos con `tokens` antes
  del offset. Vive **sólo en memoria, durante la pasada** *(FR-005)*.

## Success Criteria

Las referencias son las del **contador independiente** *(Python, en un temporal; 2026-10-08)*, que aplica D-2 y la primera aparición por
`id`, `$set.messages` incluido, sobre la copia congelada **en su sitio**. Sus tres huellas SHA-256 están en los reportes de los encargos de
009 y **no** se publican: la de `.project_root` permitiría confirmar una conjetura de ruta.

- **SC-001** *(contador)*: sobre la copia, el número de eventos `gemini` y las sumas de sus cuatro partidas son **iguales** al contador.
  Se mide con `--scan` sobre el fichero de la **primera** copia, y con `--run` en un sandbox con `GEMINI_CLI_HOME` en la **segunda copia
  congelada** *(la misma sesión, bajo `.gemini/tmp/<slug>/`, con las mismas tres huellas)*. Ninguna se escribe: el estado y la cola del
  agente van al sandbox. Resultado:
  - **10** eventos: **9** `gemini-3.8-flash` y **1** `gemini-3.5-flash-lite`;
  - `tokens_input` **95 747** *(84 627 + 11 120)*, `tokens_cache_read` **12 141**, `tokens_cache_creation` **0** y `tokens_output`
    **2 482** *(2 067 + 415)*;
  - resumen: `respuestas 16 · eventos 10 · repetidas 6 · sin identificador 0 · incoherentes 0 · sin modelo 0 · total descuadrado 0 ·
    ficheros en formato anterior 0`.
- **SC-002** *(contra `/stats`)*: Σ(`tokens_input` + `tokens_cache_read`) = **107 888**, la «entrada» del papel `main` en `/stats`. Y
  Σ`tokens_output` − Σ`thoughts` = **476**, su «salida» *(descubrimiento §Contraste)*.
- **SC-003** *(segunda pasada)*: una segunda `--run` sobre la copia da **0** eventos `gemini`. La copia conserva las tres huellas.
- **SC-004** *(reanudación en otra pasada — P-3)*: un fixture **sintético** con la forma de la copia *(llamadas a herramientas, una
  compresión y una reanudación)*, cortado justo antes de la reanudación y leído en dos pasadas, da lo mismo que la copia en ese corte:
  1.ª pasada `respuestas 12 · eventos 8 · repetidas 4`; 2.ª pasada `respuestas 4 · eventos 2 · repetidas 2`. Con P-3 (b) daría `eventos
  3 · repetidas 1`, y el test cae.
- **SC-005** *(sin coste)*: el 100 % de los eventos `gemini` llevan `cost_available = false` y `cost_usd = 0`, y ninguno lleva `tool ≠
  "gemini"`.
- **SC-006** *(identidad)*: la función nueva reproduce los tres vectores normativos del contrato. Un fixture lleva centinelas en `id`,
  `sessionId`, `projectHash`, `.project_root` **y en el nombre del fichero y de la carpeta del subagente**. Tras una pasada, **ninguno**
  está en la cola ni en lo transmitido, y en `state.json` sólo aparecen los de la ruta, **dentro de su clave**.
- **SC-007** *(partidas, sintético)*: `input` 100, `cached` 40, `tool` 5, `output` 7, `thoughts` 3 → `tokens_input` 65,
  `tokens_cache_read` 40, `tokens_cache_creation` 0, `tokens_output` 10. Con `tool` fuera o `thoughts` fuera, el test cae.
- **SC-008** *(clasificación, sintético)*: un fixture con una respuesta de cada clase da el resumen `respuestas 7 · eventos 3 · repetidas
  1 · sin identificador 1 · incoherentes 2 · sin modelo 1 · total descuadrado 1`. Las clases: evento, evento sin modelo, evento con total
  descuadrado, repetida, sin `id`, `cached > input` y partida negativa. Cumple la identidad de FR-019.
- **SC-009** *(tokens tardíos y estado final)*: en un fixture, un `id` aparece primero sin `tokens` y después con ellos: da **1** evento. En
  otro, un `$set.messages` final sin tokens deja a cero el estado reconstruido: los eventos salen igual.
- **SC-010** *(proyecto)*: con `.project_root`, `project_ref` es `Derivar` de su texto. Sin él, sale vacío. Un subagente lleva el de su
  `<slug>`. Si la misma respuesta está en dos carpetas, una con `.project_root` y otra sin él, el evento lleva el proyecto *(P-7)*.
- **SC-011** *(formatos y raíz)*:
  - un `.json` en `chats/` da 0 eventos y «ficheros en formato anterior 1», y en la pasada siguiente, sin cambios, 0;
  - `logs.json`, `*.unreadable-*` y `*.tmp-*` no se leen;
  - con `GEMINI_CLI_HOME=<dir>`, la raíz es `<dir>/.gemini`;
  - con `GEMINI_CLI_HOME=""`, es `~/.gemini`.
- **SC-012** *(sin Gemini)*: sin raíz de Gemini, y también con `GEMINI_CLI_HOME=""` y sin `~/.gemini`, la salida de `--run` sobre fixtures
  de Claude Code y Codex es **byte a byte** igual a un fichero de referencia generado con el binario de `15ce93b` antes de tocar
  `main.go`.
- **SC-013** *(coste del contexto)*: sobre un fichero sintético de **≥ 100 MB**, generado en un temporal y nunca en el repo, una pasada
  que encuentra **1** respuesta nueva tarda **≤ 3 s** *(el 5 % del ciclo de 60 s, `internal/config/config.go:47`)*. Se mide tres veces.
  **Si no se cumple, cambia el mecanismo; el tope no se mueve.**
- **SC-014** *(puertas)*: frontera con 0 bytes de diff *(FR-027)*; los **574** tests actuales, verdes y **sin tocar** *(`git diff 15ce93b
  --stat -- '*_test.go'` sólo con ficheros nuevos)*; `golangci-lint run` → 0.
- **SC-015** *(textos)*: los textos nuevos, byte a byte iguales a los aprobados, también sobre el binario publicado.
- **SC-016** *(Gemini no rompe a los demás)*: con un fichero de Gemini ilegible, junto a fixtures sanos de las tres herramientas:
  - salen los eventos de los sanos;
  - stderr lleva `gemini: fichero omitido: …`;
  - `state.json` se guarda;
  - una segunda pasada no reencola nada de Claude Code ni de Codex;
  - cuando el fichero vuelve a ser legible, sus respuestas salen.
- **SC-017** *(orden de escritura y sin las otras herramientas)*:
  - con el forzado de 008 SC-015 *(`Load` pasa, `Save` falla)*, la cola contiene los eventos de Gemini y una segunda pasada los reencola
    con los **mismos** `event_id`;
  - con sólo la raíz de Gemini, `--run` encola sus eventos y termina con 0.

## Censo de tests — autorización para la fase de tareas

**Ninguno de los existentes.** Los nuevos van en ficheros nuevos de `internal/ingest/`, `internal/config/` y `cmd/permea/`, con fixtures
**sintéticos** nuevos en `testdata/`. **Censo de código fuera de los `_test.go`** *(M-9)*: `internal/testutil/sandbox.go` gana
`t.Setenv("GEMINI_CLI_HOME", "")`. `internal/testutil/sandbox_test.go:20` comprueba las variables de antes, y la nueva no cambia su
veredicto.

## Fuera de alcance

| # | Qué | Por qué |
|---|---|---|
| **N-1** | Las llamadas auxiliares y los intentos fallidos | D-1: no están en el fichero |
| **N-2** | La telemetría a fichero *(Q13)* | D-1: va a ficha |
| **N-3** | Leer los `session-*.json` | D-3 |
| **N-4** | El coste y las tarifas de Google | D-4 |
| **N-5** | El nivel de servicio, la cuota gratuita, el modo de autenticación y el contexto largo | El registro no los trae *(Q12)* |
| **N-6** | **Antigravity** y cualquier otro cliente de Gemini | No escriben este fichero |
| **N-7** | `~/.cache/.gemini` *(macOS con `SANDBOX=sandbox-exec`, Q1)* | Un modo de *sandbox*, no la instalación normal |
| **N-8** | Leer desde WSL las sesiones de Windows | M-7 lee el directorio personal propio |
| **N-9** | `logs.json` *(lo que escribe el usuario)* y `logs/` | No son consumo |
| **N-10** | Un interruptor para desactivar Gemini | P-1 (a) |
| **N-11** | Una raíz de sesiones que sea un enlace simbólico | P-9 rechazada: no se garantiza ni se prueba |

## Límites declarados

- **D-1**: el lector queda **por debajo de la factura** por las llamadas auxiliares y los intentos fallidos.
- **Retención de 30 días** *(Q9)*: la CLI borra por defecto las sesiones viejas al arrancar. Lo que el agente no haya leído antes, se pierde.
- **Antigravity no se lee** *(N-6)*.
- **`/stats` no es una referencia directa**:
  - tras `--resume latest` cuenta sólo la ejecución *(Q8)*;
  - su «entrada» incluye la caché, y su «salida» excluye el razonamiento, al revés que D-2;
  - cuenta las auxiliares y los errores.

  Sólo cuadra como en SC-002, y con la sesión entera en una ejecución.
- **Una respuesta enviada sin proyecto** desde una carpeta antigua *(sin `.project_root`)* no se corrige si después aparece su copia con
  proyecto: la plataforma descarta el `event_id` repetido.

## Textos aprobados (2026-10-08) *(P-10 ✅; las frases de coherencia de P-11, en E-1)*

**Resumen de Gemini** *(stderr, una línea; en `--run`, tras la de Codex; en el demonio, sólo con novedades)*. Añade «total descuadrado»
*(P-5)* a la propuesta del dueño, y no lleva «comprimidos» *(Gemini no comprime)*:
```
gemini: respuestas %d · eventos %d · repetidas %d · sin identificador %d · incoherentes %d · sin modelo %d · total descuadrado %d · ficheros en formato anterior %d
```
**Fichero de Gemini omitido** *(stderr, una línea por fichero)*:
```
gemini: fichero omitido: %v
```
**Línea de `--scan`** *(stdout)*: la de Codex, sin cambios *(008 §Textos aprobados)*.

**README** *(sección nueva, tras «### Codex CLI»)*:
```
### Gemini CLI

Si existe `~/.gemini/tmp` (o `$GEMINI_CLI_HOME/.gemini/tmp`), el agente lee también el consumo de Gemini CLI.
Cada respuesta del modelo es un evento con `tool = gemini`, sus tokens y su modelo; el razonamiento cuenta
como salida. El coste no lo calcula el agente: los eventos salen con `cost_available = false`. Hace falta
Gemini CLI 0.39.0 o posterior; las sesiones anteriores se cuentan como «formato anterior» y no se envían.
Las llamadas internas de la CLI (compresión, enrutado) y los intentos fallidos no quedan en sus sesiones,
así que el agente no los ve. Gemini CLI borra por defecto las sesiones de más de 30 días. Si la carpeta no
existe, no cambia nada. Antigravity no guarda su consumo en estas sesiones y el agente no lo lee.
```
**README, coherencia** *(P-11; E-1, aprobada por el dueño el 2026-10-08, 18:30)*. Sólo cambian las palabras que nombran a Gemini; en el
fichero se parten al ancho de línea, y se comprueban con `grep -F` sobre el texto con los saltos normalizados:
1. «Agente local que lee los logs de uso de herramientas de IA (Claude Code, Codex CLI y Gemini CLI), calcula **en local** el coste de Claude Code y transmite…»
2. «> ⚠️ **La primera pasada envía todo el historial que conserven Claude Code, Codex CLI y Gemini CLI**, no sólo lo»
3. «permea --scan <fichero.jsonl>   # prueba en seco: un evento por mensaje (por respuesta, en Codex y en Gemini), sin tocar estado ni cola»
4. «…Con Codex activo, una tercera línea, `codex: …`, con sus recuentos, y con Gemini activo, otra, `gemini: …`.»
5. «…En una sesión de Codex o de Gemini la línea no lleva `cw5m=` ni `cw1h=`, y el coste es 0.»
6. «- **`--run`** hace una pasada: descubre los logs de Claude Code y las sesiones de Codex y de Gemini, lee lo nuevo»
7. «- **Codex y Gemini**: el agente no calcula su coste; lo pone la plataforma (ver «Codex CLI» y «Gemini CLI»).»
8. «internal/ingest   lectores por herramienta (claude_code, codex, gemini) + tests de frontera»
9. «…espejo exacto de las 17 filas de Anthropic del catálogo de tarifas de la plataforma (`permea-dev/permea-platform` · `backend/config/pricing.php` · `8f147d1`); las de otros proveedores sólo están en la plataforma.»

**CHANGELOG `0.6.0`** *(encabezado con `PENDIENTE` hasta la etiqueta)*:
```
## 0.6.0 — PENDIENTE

### Nuevo
- Lee también el consumo de Gemini CLI, de `~/.gemini/tmp` (o de `$GEMINI_CLI_HOME/.gemini/tmp`), si esa
  carpeta existe. Cada respuesta del modelo es un evento con `tool = gemini`.
  (`specs/009-lector-gemini/spec.md`, FR-001, FR-002, FR-006)
- Los eventos de Gemini llevan sus tokens y su modelo, sin coste: el coste lo calcula la plataforma.
  (`specs/009-lector-gemini/spec.md`, FR-010, FR-011, FR-013)

### Sin cambios
- Claude Code y Codex se leen exactamente como en la 0.5.0. Sin carpeta de Gemini, la salida no cambia.
  (`specs/009-lector-gemini/spec.md`, FR-023, FR-028)

### Limitaciones conocidas
- Las llamadas internas de Gemini CLI y los intentos fallidos no constan en sus sesiones: el consumo
  enviado queda por debajo de la factura. (`specs/009-lector-gemini/spec.md`, D-1)
- Sólo se leen las sesiones de Gemini CLI 0.39.0 o posterior; las anteriores se cuentan y no se envían.
  (`specs/009-lector-gemini/spec.md`, FR-018; D-3)
```

## Preguntas abiertas — **ninguna** *(ratificadas el 2026-10-08, 12:20, con su recomendación)*

| # | Pregunta | Respuesta |
|---|---|---|
| **Q-1** ✅ | ¿Converge en **Windows** el `project_ref` de `.project_root` *(ruta en minúsculas, Q1)* con el de Claude Code para el mismo directorio? | **Comprobarlo en el ensayo de Windows.** Si `Derivar` reconoce la raíz del repositorio, o si `EvalSymlinks` devuelve la grafía real del disco, coinciden. Si el directorio ya no existe, el *fallback* recibe la forma en minúsculas y puede divergir. No se toca el resolutor |
| **Q-2** ✅ | ¿«respuestas» o «apariciones» en el resumen? *(cuenta apariciones: 16 para 10 respuestas)* | **«respuestas»**, por simetría con Codex, y con la identidad de FR-019 a la vista |
| **Q-3** ✅ | ¿Se emite una respuesta con `timestamp` ausente con la hora de la pasada? | **No**: es «incoherente», como 008 E-4 *(FR-019)* |
| **Q-4** ✅ | ¿Ficha para la telemetría a fichero *(N-2)*? | **Sí, aparte**. Con lo que deja abierto Q13: la forma real, emparejar `api_response` con su `operation.details`, y que hay que encenderla en la CLI del usuario |

## Assumptions
1. **En WSL no hay sesiones de Gemini del dueño**: están en **Windows**, y las leerá el agente de Windows.
2. **El formato JSONL es estable desde la 0.39.0**: el objeto `tokens` tiene los mismos seis campos desde antes *(Q10, `v0.38.0`)*.
3. **Los `id` de las respuestas los genera la CLI con `randomUUID`** *(Q4)*, y ninguna función de la 0.63 los reutiliza en otra sesión.

## Dependencias
- **La plataforma**: tarifas de Gemini, **después** del lector *(D-4)*. Dará fila sólo a los identificadores vistos en eventos reales: hoy
  `gemini-3.8-flash` y `gemini-3.5-flash-lite`.
- **006**: el esquema del `event_id`, sin tocarlo. **007**: `Recorrer`, tal cual. **008**: el patrón del lector de Codex *(prefijo, errores
  por fichero, recuentos)*, sin cambiar su conducta.

## Registro de enmiendas

| # | Fecha | Qué cambia | Por qué |
|---|---|---|---|
| **Borrador** | 2026-10-08 | Primera redacción, con D-1 a D-4 ratificadas el 2026-10-08, y P-1 a P-11 y Q-1 a Q-4 para ratificar | Encargo 009 · spec · 2/n |
| **Ratificación** | 2026-10-08 12:20 | **El dueño ratifica** P-1 a P-8, P-10 y P-11 **(a)**, y Q-1 a Q-4 con su recomendación. **Rechaza P-9 (b)**: no se resuelven enlaces *(N-11)*; FR-002 ya no los nombra, y SC-001 mide `--run` sobre la **segunda copia congelada** con `GEMINI_CLI_HOME`. **Aprueba los textos** *(P-10)*, con una frase más en «### Gemini CLI» sobre Antigravity. Las frases de coherencia *(P-11)* quedan pendientes hasta B6 | Ratificación del dueño |
| **E-1** | 2026-10-08 18:30 | **El dueño aprueba las nueve frases de coherencia del README** *(P-11, T043)*, a §Textos aprobados, «README, coherencia». La 9 se escribe tras comprobar que las 17 filas de Anthropic de `internal/pricing` son iguales a las del catálogo de la plataforma en `6569815` *(`soporte/registro.md` §B6 y B7)* | Que el README no quede incompleto con la 0.6.0 |
