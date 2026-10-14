# 011 · «Lector de Copilot CLI» — Especificación

**Feature Branch**: `011-lector-copilot` · **Created**: 2026-10-14 · **Status**: **Borrador** *(D-1 a D-6, del dueño el 14-10; P-1 a P-11, para ratificar)* ·
**Cabeza de partida**: `872b4f3` *(sobre `15c4c47`, cierre de la 0.7.0)*
**Input**:
- las decisiones del dueño D-1 a D-6 *(2026-10-14)*;
- el descubrimiento del 14-10 *(`soporte/descubrimiento.md`, Q1–Q11, sobre la **copia congelada** de 3 sesiones de Copilot CLI 1.0.97)*;
- la muestra 2 *(`soporte/descubrimiento-muestra-2.md`, M1–M4: sesión matada, compactación y `COPILOT_HOME`)*.

> ⛔ Toda afirmación sobre nuestro código lleva `fichero:línea`, medida sobre `872b4f3` *(el código es el de `15c4c47`)*. Las de Copilot citan el
> descubrimiento *(«desc. Qn»)* o la muestra 2 *(«m2 Mn»)*. Lo que añade la spec sin que nadie lo dijera va marcado **`DECIDÍ YO`**. Ningún dato de esta
> spec procede del contenido de las sesiones: sólo recuentos, tokens, créditos, nombres de modelo, claves y tipos.

## Contexto — el agente lee Claude Code, Codex, Gemini CLI y OpenCode

1. **Cuatro lectores hoy**, en `generate()`, antes del **único** `st.Save` *(`cmd/permea/main.go:392`)*: Claude Code, Codex *(`main.go:355-366`)*, Gemini
   *(`main.go:368-379`)* y OpenCode *(`main.go:381-390`)*.
2. **Copilot CLI escribe un JSONL de sólo añadir por sesión**: `<COPILOT_HOME | ~/.copilot>/session-state/<id>/events.jsonl` *(desc. Q1; m2 M3)*.
   - **Cada llamada al modelo principal deja una línea `session.usage_record`**, con tokens, modelo, un id único por llamada *(`accounting.usageId`)* y
     su marca de tiempo *(desc. Q3)*.
   - **Cada compactación con éxito deja una línea `session.compaction_complete`** cuyo consumo **no está en ningún `usage_record`**: el 19,2 % de los
     créditos de la sesión medida *(m2 M2)*.
   - Los `shutdown` son acumulados de la sesión y faltan si se mata el proceso; los `usage_record` ya escritos no se pierden *(desc. Q5; m2 M1)*.
3. **Es casi el lector de Codex**: una línea = un evento, sin espera de cierre; lectura por offset con `state.Recorrer` *(`internal/state/state.go:121-169`)*;
   y el contexto *(el `cwd` de cada tramo)* en líneas anteriores, que se releen con un prefiltro de bytes, como `leerPrefijoCodex`
   *(`internal/ingest/codex_contexto.go:212`)*.
4. **La plataforma ya acepta el evento tal cual.** `tool` es texto libre; un modelo sin tarifa llega «ciego y señalizado»; y un `event_id` repetido se
   descarta *(009 §Contexto 4)*.

## Decisiones del dueño *(2026-10-14; no se discuten aquí)*

| # | Decisión |
|---|---|
| **D-1** | `tool = "copilot_cli"`, distinto de un futuro Copilot de IDE leído por API desde la plataforma |
| **D-2** | **Fuente**: `session.usage_record` + `session.compaction_complete` con `success: true`. **Nada más**: ni `shutdown`, ni `usage_checkpoint`, ni `checkpoints/`, ni `session.db`, ni `model.*`, ni líneas con contenido. Las **llamadas auxiliares** *(`gpt-4o-mini`, `gpt-5.4-nano`, a 0 créditos; desc. Q3)* **no se cuentan**: límite declarado |
| **D-3** | **Partidas**: `tokens_input` = input − lectura − escritura; `tokens_output` = output, **sin sumarle el razonamiento** *(desc. Q4)*; `tokens_cache_creation` = cacheWrite; `tokens_cache_read` = cacheRead. Si la resta da negativo, el evento es **incoherente** *(como 010)* |
| **D-4** | **Sin coste**: `cost_usd = 0` y `cost_available = false`, como 010 SC-028. Los créditos *(AIU)* **no** van en el evento: sirven sólo de **control** en W1/W2, **Σ créditos × 0,01 = coste de la plataforma** |
| **D-5** | **Raíz**: `$COPILOT_HOME/session-state` si la variable está puesta; si no, `~/.copilot/session-state`. Es la **5.ª** variable de entorno del agente |
| **D-6** | **Sólo lectura**: el agente no crea, toca ni borra `.lock`. Las carpetas sin `events.jsonl` son 0 eventos y **no generan aviso** |

## Decisiones de método *(heredadas de 008–010)*

| # | Decisión |
|---|---|
| **M-1** | El evento no cambia: mismos campos y `SchemaVersion = 1` *(`internal/event/event.go:12-35`)*. La frontera de FR-027, con **0 bytes** de diff |
| **M-2** `DECIDÍ YO` | Versión **`0.8.0`**. Claude Code, Codex, Gemini y OpenCode se leen exactamente como en la 0.7.0: los tests de `15c4c47` *(su recuento, `go test -count=1 -v ./...`, se fija en el plan: en este encargo no se ejecutan suites)* siguen verdes **sin tocarlos** |
| **M-3** | `internal/testutil/sandbox.go` fija también `COPILOT_HOME` *(vacía)*, junto a las otras cuatro *(`sandbox.go:63-68`)*. Sin ella, un test de proceso leería las sesiones reales del desarrollador |
| **M-4** | Los errores son **por fichero**, como en 008 y 009 *(`cmd/permea/main.go:404-426`)*. Encolar sigue siendo fatal |
| **M-5** | Es la **5.ª** lectura de entorno en producción, tras `CODEX_HOME`, `GEMINI_CLI_HOME`, `XDG_DATA_HOME` y `OPENCODE_DB` *(`internal/config/codex.go:19`, `gemini.go:20`, `opencode.go:21`, `:25`)* |

## Propuestas — **para ratificar** por el dueño

| # | Propuesta | Alternativas | Recomendación |
|---|---|---|---|
| **P-1** | **`event_id`**: `["permea/event_id/v1", "copilot_cli", <tipo>, <id>]`, con la codificación de 006 *(`internal/ingest/eventid.go:31`, `:43`)*, en un fichero y una función nuevos y un contrato nuevo, `contracts/event-id-copilot.md`:<br>• `usage_record`: tipo `respuesta` *(`tipoRespuesta`, `codex_eventid.go:29`)* e id = `usage.accounting.usageId`;<br>• `compaction_complete`: tipo **`compactacion`** `DECIDÍ YO` e id = `requestId`.<br>Sin id textual no vacío: «sin identificador», no se emite | (b) para `usage_record`, `(sourceSessionId, sequence)` · (c) para la compactación, el `id` de la línea · (d) un único tipo para los dos | **(a)**: `usageId` es único *(9 de 9; m2 M4)*, igual a `apiCallId`, y **del servidor**: si Copilot copiara las líneas *(una bifurcación: `.workspace-fork.lock`, `fork_count`; desc. Q1)*, daría el mismo `event_id` y la plataforma descartaría la copia. El `id` de la línea es un UUID local que una copia podría regenerar. `requestId` es el del servidor y coincide con el de su `model.model_call_success` *(m2 M2)*. El tipo distinto separa un `requestId` de un `usageId` iguales por azar |
| **P-2** | **Frontera: offset por fichero**, con `state.Recorrer` y las reglas de Codex:<br>• tamaño = offset: no se abre;<br>• **tamaño < offset** *(truncado o reescrito más corto)*: se relee desde 0, y lo ya emitido sale con el **mismo** `event_id` *(`state.go:129-131`)*;<br>• línea final sin LF: no se consume; se reintenta *(`state.go:146-148`)*;<br>• reescrito **del mismo tamaño o mayor**: no se detecta; lo nuevo tras el offset se lee, y una línea cortada se salta como corrupta | (b) además, una huella de la 1.ª línea para detectar reescrituras · (c) releer el fichero entero en cada pasada | **(a)**: el fichero es de **sólo añadir** *(el offset de cada `shutdown` y `resume` cuadra en 5 de 5; desc. Q1, m2 M2)*, como el de Codex. (b) protege de algo que no se ha visto; (c) cuesta lo que el fichero entero en cada pasada *(838 KB en E)*. El hueco de (a) se declara *(§Límites)* |
| **P-3** | **Contexto por tramo**: el `cwd` de un registro es el `data.context.cwd` del **último** `session.start` o `session.resume` anterior en el mismo fichero; la sesión, el `data.sessionId` del `session.start`. Con lectura por offset, el prefijo `[0, offset)` se relee filtrando por bytes *(`"type":"session.start"`, `"type":"session.resume"`)*, como `leerPrefijoCodex` *(`codex_contexto.go:212`)*.<br>• `project_ref` = `Resolutor.Derivar(cwd, sal)` *(`internal/project/resolve.go:318`)*;<br>• `session_ref` = `event.Ref(sal, sessionId)` *(`internal/event/event.go:40-46`)*: la pseudonimización de Claude Code *(`internal/ingest/claudecode.go:125-126`)*.<br>Sin `session.start`: `session_ref` vacío; sin `cwd`: `project_ref` vacío | (b) `session_ref` del nombre de la carpeta · (c) `session_ref` de `usage.accounting.sourceSessionId` · (d) un único `cwd` por fichero, el del `session.start` | **(a)**: es lo que el fichero dice, sin inferir. El nombre de la carpeta = `sessionId` en 5 de 5 *(m2 M3)*, pero es un hecho medido, no un contrato. (c) no existe en `compaction_complete`. (d) se equivoca si `--continue` se lanza desde otro directorio *(`session.resume` trae su propio `cwd`; desc. Q2)* |
| **P-4** | **Momento**: `occurred_at` = la `timestamp` de la línea *(ISO 8601 en UTC, con ms; desc. Q2)* | (b) para `usage_record`, la marca menos `usage.duration` *(el inicio de la llamada)* | **(a)**: es cuando Copilot registra el consumo, y está en las dos clases de línea. Una llamada que cruza la medianoche cae en el día en que acaba, como 010 P-7 |
| **P-5** | **Modelo**: `usage.model` y `compactionTokensUsed.model`, tal cual *(`gpt-6-luna`; nunca `auto`, desc. Q7)*. `modelDisplayName` y `provider.*` no viajan. Sin modelo textual no vacío: el evento sale con `model` vacío y se cuenta «sin modelo» | (b) `modelDisplayName` · (c) normalizarlo en el agente | **(a)**, como 008–010: el agente nunca traduce modelos. Los nombres de otros planes *(§Dependencias)* son cosa de la plataforma |
| **P-6** | **Activación**: el lector está activo **en una pasada** si existe, en esa pasada, la carpeta de D-5. Se comprueba en cada `generate()`. No hay interruptor | (b) un interruptor | **(a)**, como 008 P-1, 009 P-1 y 010 P-10 |
| **P-7** | **Ficheros**: sólo `<raíz>/<carpeta>/events.jsonl`, **un nivel**. Ni `.session-operation-locks/`, ni `workspace.yaml`, `checkpoints/`, `rewind-file-snapshots/`, `files/`, `research/` o `session.db`, ni nada más hondo | (b) recorrer la raíz entera, como `ListarCodex` *(`codex_contexto.go:258-278`)* | **(a)**: D-2 y D-6. Recorrerla entera abriría directorios que no son consumo, y el nombre de los ficheros de `checkpoints/` sale del contenido *(m2 §La copia)* |
| **P-8** | **Resumen y textos**: los de §Textos. Línea `copilot_cli: …` tras la de OpenCode *(`main.go:466-467`)*; en el demonio, sólo con novedades *(`main.go:535-536`)* | corregirlos | — |
| **P-9** | **`--scan <fichero>`**: un fichero cuya **primera línea** es `type: "session.start"` con `data.copilotVersion` textual se lee como Copilot CLI, tras la comprobación de Gemini *(`main.go:582-588`)*. Escribe las líneas `evento:` *(`imprimirEventos`, `main.go:659-671`)*, el resumen y **la línea de control** de §Textos: Σ créditos del fichero y su equivalente × 0,01. Nunca escribe estado ni cola | (b) `--scan` sin línea de control · (c) sin `--scan` | **(a)**: es el contraste de 010 D-1 adaptado a D-4. Da a W1 y W2 el control de créditos sin abrir el fichero a mano. Los créditos salen de los mismos `usage_record` y `compaction_complete` de D-2, y nunca viajan. La primera línea es `session.start` en 5 de 5 *(m2)* |
| **P-10** | **Límites declarados**, en README, CHANGELOG y §Límites:<br>• llamadas auxiliares *(D-2)*;<br>• la llamada en curso de una sesión matada *(m2 M1)*;<br>• **subagentes y *workflows***: sin muestra; `agentMetrics` sólo trae `main`, y `session.db` *(con `*_credit_charges`)* no se lee;<br>• **compactación automática**: sin muestra; se lee si deja el mismo `compaction_complete`;<br>• **Copilot de IDE** *(VS Code, JetBrains)*: no escribe aquí, o no se sabe; D-1 lo deja a la plataforma;<br>• **planes de pago**: sin muestra; los nombres de modelo, en §Dependencias | (b) medirlos antes de la 0.8.0 | **(a)**, salvo **subagentes**: recomiendo **una muestra antes de W1**, porque si un subagente factura fuera de `usage_record` el lector se queda corto sin decirlo, como pasaba con la compactación |
| **P-11** | **Prefiltro de bytes**: una línea sólo se decodifica si contiene `"type":"session.usage_record"`, `"type":"session.compaction_complete"` *(o, en el prefijo, las de P-3)*; el tipo lo confirma el JSON. De una línea decodificada sólo se leen los campos de FR-012 | (b) decodificar todas las líneas | **(a)**: las líneas con contenido *(33 KB de `system.message`, 56 KB de `messages_snapshot`; m2 M2)* no se decodifican. El marcador aparece en 9 de 9 registros y en 2 de 2 compactaciones de la copia; dentro de un texto JSON iría escapado *(`\"`)* y no casa |

## User Scenarios & Testing

### Historia 1 — El consumo de Copilot CLI llega a la plataforma *(P1)*
1. **Dada** una sesión con dos llamadas terminadas, **cuando** corre `--run`, **entonces** salen **dos** eventos `tool = copilot_cli`, con sus partidas según
   D-3, `cost_available = false` y el `usage.model` de cada línea.
2. **Dada** la misma sesión en una segunda pasada sin cambios, **entonces** no sale **ningún** evento y el fichero **no se abre**.

### Historia 2 — La compactación no se pierde *(P1)*
1. **Dada** una sesión con un `/compact` con éxito, **entonces** sale **un** evento más, con las partidas de `compactionTokensUsed`.
2. **Dado** un `/compact` sin nada que compactar *(`success: false`)*, **entonces** no sale nada.

### Historia 3 — Una sesión matada no pierde lo ya escrito *(P1)*
1. **Dada** una sesión cortada a mitad de su 2.º prompt, **entonces** sale el evento del 1.º, y nada de la llamada cortada.
2. **Dada** una última línea a medio escribir *(sin LF)*, **entonces** no se lee; cuando se completa, la pasada siguiente la emite.

### Historia 4 — El agente nunca toca la carpeta de Copilot *(P1)*
1. **Cuando** corre una pasada, **entonces** todos los ficheros de `session-state/` tienen la **misma md5** antes y después, no aparece ningún fichero nuevo y
   ningún `.lock` cambia.

### Historia 5 — Quien no usa Copilot CLI no nota nada *(P1)*
1. **Dada** una instalación sin `~/.copilot/session-state` ni `COPILOT_HOME`, **entonces** la salida de `--run` es **byte a byte** la de la 0.7.0.

### Historia 6 — Copilot CLI sin las otras herramientas *(P2)*
1. **Dada** sólo la carpeta de Copilot CLI, **entonces** sus eventos salen igual, y `--run` termina con 0.

### Edge Cases
- **`--continue`**: la sesión sigue en el mismo fichero, con un `session.resume` y sin nuevo `start`; los `usage_record` no se repiten *(desc. Q5)*. El
  `cwd` del tramo es el del `resume` *(P-3)*.
- **Carpeta sin `events.jsonl`** *(un intento fallido; desc. Q9)*: 0 eventos, sin aviso y sin recuento *(D-6)*.
- **`.lock` huérfanos**: Copilot no los borra *(m2 M1)*. No se leen ni se cuentan.
- **Llamadas auxiliares** *(`model.*`)*: no son registros. No se decodifican ni se cuentan *(D-2, P-11)*.
- **El `shutdown` y los checkpoints**: no se leen. Sus totales no se comparan con nada en `--run` *(D-2)*.
- **Compactación fallida, o con éxito sin `compactionTokensUsed`**: la primera no es registro; la segunda es «incoherente» *(FR-020)*.
- **Sesión matada**: no hay `shutdown`, y no hace falta *(D-2)*. La llamada en curso no está en disco *(§Límites)*.
- **Fichero truncado** *(tamaño < offset)*: se relee desde 0 y la plataforma descarta los `event_id` repetidos *(P-2)*.
- **Fichero reescrito del mismo tamaño o mayor**: no se detecta *(P-2, §Límites)*.
- **Un mismo `usageId` dos veces en una pasada** *(una copia del fichero en otra carpeta)*: sale una vez y se cuenta «repetidas» *(como Codex)*.
- **Un registro con `cacheRead + cacheWrite > inputTokens`**: «incoherente» *(D-3)*.
- **Campos de más en `usage`**: se ignoran *(`interTokenLatencyMs` es opcional; m2 M1)*. **`usage` sin las partidas**: «incoherente».
- **Una línea que no es JSON**: `skip (línea corrupta): …`, como Codex *(`codex_contexto.go:174`)*; no corta el fichero.
- **`COPILOT_HOME` apunta a algo que no existe**: el lector no está activo en esa pasada; no dice nada, no falla y no escribe su línea.
- **El proceso cae entre encolar y guardar**: la pasada siguiente reencola los mismos `event_id`, y la plataforma los descarta.

## Requirements

### A · Dónde y cuándo lee *(D-5, D-6, P-6, P-7, P-2)*

- **FR-001** *(raíz — D-5)*: `$COPILOT_HOME/session-state` si `COPILOT_HOME` no es vacía; si no, `<hogar>/.copilot/session-state`, con el
  `os.UserHomeDir()` de las otras raíces. Se resuelve una vez por proceso, en `setup()`, como las demás *(`main.go:230-243`)*.
- **FR-002** *(activación — P-6)*: según P-6. Si no está activo, el lector no escribe nada, no devuelve error y no cambia ningún texto.
- **FR-003** *(ficheros — P-7)*: según P-7. Una carpeta sin `events.jsonl` no produce nada *(D-6)*.
- **FR-004** *(estado)*: cada `events.jsonl` tiene una entrada en el mismo `state.json`, con la ruta como clave y los **cuatro** campos de siempre
  *(`internal/state/state.go:16-21`)*, **sin** `Frontera` *(la de OpenCode)*. No hay campo nuevo.
- **FR-005** *(entre pasadas — P-2)*: un registro **DEBE** salir **una vez**, salvo tras un fichero truncado, que lo reemite con el mismo `event_id`. Un
  fichero que no creció **no se abre**. El coste de una pasada es el de lo nuevo más el prefijo de los ficheros que crecieron *(P-3)*.

### B · Cómo lee *(D-2, D-6, P-3, P-11)*

- **FR-006** *(sólo lectura — D-6)*: el lector **NUNCA** escribe, crea, bloquea, renombra ni borra nada bajo la raíz. Abre `events.jsonl` en sólo lectura.
- **FR-007** *(prefiltro — P-11)*: según P-11. El tipo lo confirma el JSON; una línea que pasa el filtro y no es de los tipos de D-2 no es registro.
- **FR-008** *(contexto — P-3)*: según P-3. Un registro leído en una pasada posterior **DEBE** llevar las mismas referencias que leyendo el fichero de una vez.
- **FR-009** *(líneas largas)*: no hay tope de longitud de línea *(`ReadBytes`, `state.go:145`)*; las de la copia llegan a 56 607 B *(m2 M2)*.

### C · El evento *(D-1 a D-4, P-1, P-3 a P-5)*

- **FR-010** *(unidad — D-2)*: es **registro** cada línea `session.usage_record`, y cada `session.compaction_complete` con `data.success` = `true`. Cada
  registro se clasifica *(FR-020)* y produce **como mucho un** evento.
- **FR-011** *(identidad — P-1)*: el `event_id` es el de `contracts/event-id-copilot.md`, y **DEBE** reproducir sus vectores byte a byte.
- **FR-012** *(lo que se lee)*: de un registro, **sólo**: `type`, `timestamp`; de `usage_record`, `data.usage.{model, inputTokens, outputTokens,
  cacheReadTokens, cacheWriteTokens, accounting.usageId}` *(y, en `--scan`, `copilotUsage.totalNanoAiu`)*; de `compaction_complete`, `data.{success,
  requestId}` y `data.compactionTokensUsed.{model, inputTokens, outputTokens, cacheReadTokens, cacheWriteTokens}` *(y, en `--scan`,
  `copilotUsage.totalNanoAiu`)*; de `start`/`resume`, `data.sessionId` y `data.context.cwd`. **Nunca** `reasoningTokens`, `toolCounts`, `summaryContent`,
  `checkpointPath` ni ningún tipo de la tabla ⛔ de desc. Q2.
- **FR-013** *(partidas — D-3)*: según D-3, sobre `inputTokens`, `outputTokens`, `cacheReadTokens` y `cacheWriteTokens` de `usage` o de
  `compactionTokensUsed`. Una partida ausente vale 0, como en Codex *(E-4 de 008)*; si las cuatro faltan, es «incoherente».
- **FR-014** *(sin identificador)*: un registro sin el id de P-1 *(ausente, vacío o no textual)* no se emite y se cuenta como «sin identificador».
- **FR-015** *(modelo — P-5)*: según P-5.
- **FR-016** *(momento — P-4)*: según P-4.
- **FR-017** *(herramienta y coste — D-1, D-4)*: `tool = "copilot_cli"`, `cost_usd = 0` y `cost_available = false`, **siempre**. El lector no consulta
  `internal/pricing`, y los créditos **nunca** entran en el evento.
- **FR-018** *(proyecto y sesión — P-3)*: según P-3. `machine_ref`, `dev_id`, `org_id` y `agent_version`, como en Claude Code.
- **FR-019** *(nada del proveedor viaja)*: **NUNCA** se copian a ningún campo del evento, a la cola ni a nada que se transmita: `usageId`, `requestId`,
  `sessionId`, el nombre de la carpeta, `cwd`, ni ningún texto de ninguna línea. `sessionId` y `cwd` sólo viajan derivados con sal. La ruta del fichero es
  la clave de `state.json`, y queda **en local**.

### D · Clasificación y lo que ve el usuario *(P-8, P-9)*

- **FR-020** *(la cuenta del resumen)*: en cada pasada, `registros = eventos + repetidas + sin identificador + incoherentes`. «Compactaciones» y «sin
  modelo» son **subconjuntos** de «eventos».
  - **Incoherente**: `usage` *(o `compactionTokensUsed`)* no es un objeto, o le faltan las cuatro partidas; alguna partida no es numérica, no es entera o
    es negativa; `cacheRead + cacheWrite > inputTokens` *(D-3)*; o la `timestamp` no es una fecha RFC 3339.
  - **Orden**: sin identificador → incoherente → evento, como 008 E-4.
- **FR-021** *(resumen)*: con el lector activo, `--run` escribe **una** línea más en stderr *(§Textos)*, tras la de OpenCode, o tras la anterior que haya. El
  demonio la escribe sólo si en ese ciclo hubo registros *(como `HayNovedades`, `main.go:526-536`)*.
- **FR-022** *(`--scan` — P-9)*: según P-9. Nunca escribe estado ni cola.
- **FR-023** *(los textos de antes no cambian)*: sin Copilot CLI activo, la salida de `--run`, del demonio y de `--scan` es **byte a byte** la de la 0.7.0. Con
  Copilot CLI, sólo se añade la línea de FR-021, y «N eventos encolados» cuenta también los suyos.

### E · Frontera, durabilidad y errores

- **FR-024** *(orden de escritura)*: Copilot CLI se lee **después** de OpenCode y **antes** del único `st.Save` *(`main.go:392`)*. Sus eventos se encolan antes
  de guardar el estado.
- **FR-025** *(sin las otras herramientas)*: una instalación sólo con Copilot CLI **DEBE** enviar sus eventos.
- **FR-026** *(Copilot CLI no rompe a los demás)*: un error al leer un fichero **DEBE** escribir `copilot_cli: fichero omitido: %v`, **no** avanzar su
  offset, **no** cortar la pasada y **no** impedir guardar el estado de los demás. Un fallo al **encolar** sigue siendo fatal.
- **FR-027** *(frontera — M-1)*: `git diff 15c4c47` de lo siguiente, **vacío**:
  - `internal/event/` e `internal/ingest/eventid.go`;
  - `eventid_test.go`, `codex_eventid.go`, `codex_eventid_test.go`, `gemini_eventid.go`, `gemini_eventid_test.go`, `opencode_eventid.go`,
    `opencode_eventid_test.go` y `boundary_test.go`;
  - los contratos de 006, 008, 009 y 010.
- **FR-028** *(versión — M-2)*: `0.8.0`. Lo demás, exactamente como en la 0.7.0.

### Key Entities
- **Registro de Copilot CLI**: una línea `session.usage_record`, o `session.compaction_complete` con éxito. Lleva `timestamp`, un id del servidor *(`usageId` o
  `requestId`)*, el modelo y las cuatro partidas.
- **Tramo**: lo que va de un `session.start` o `session.resume` al siguiente, en un mismo fichero. Da `cwd`; la sesión la da el `start`.
- **Estado de un fichero**: los cuatro campos de `state.json`, con `Offset` al final de la última línea completa.

## Success Criteria

Las referencias son las de `descubrimiento-muestra-2.md` *(copia `congelada-2026-10-14b`, leída en sólo lectura)*. Las huellas del contenido **no** se publican.

- **SC-001** *(contra la copia congelada)*: con `--run` en un sandbox con `COPILOT_HOME` = la raíz de la copia *(lee su `session-state/`: A+B, C, D y E)*, sin
  escribir en ella:
  - **9** eventos `copilot_cli`, los 9 con `gpt-6-luna`: 8 de `usage_record` y **1** de compactación;
  - `tokens_input` **99**, `tokens_output` **323**, `tokens_cache_read` **99 938** y `tokens_cache_creation` **39 945**;
  - resumen: `copilot_cli: registros 9 · eventos 9 · compactaciones 1 · repetidas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0`;
  - 0 eventos de las llamadas auxiliares, de los `shutdown` y de la carpeta sin `events.jsonl`; ningún aviso por ella;
  - y `--scan` sobre el `events.jsonl` de F da 1 evento *(15 369 de entrada con caché → `tokens_input` 3, salida 5, lectura 9 692, escritura 5 674)*.
  - La copia conserva sus sha256.
- **SC-002** *(control de créditos — D-4)*: sobre la copia, **Σ eventos × la tarifa de `gpt-6-luna`** *(USD por millón: entrada 0,10 · salida 0,50 ·
  escritura 0,125 · lectura 0,01)* = **Σ créditos × 0,01** = **$0.006163905** *(616 390 500 nanoAIU)*. La línea de control de `--scan`, sumada sobre los 4
  ficheros, da esa cifra.
- **SC-003** *(segunda pasada)*: una segunda `--run` sobre la copia da **0** eventos `copilot_cli`, y ningún fichero se abre *(contador en el test)*.
- **SC-004** *(sesión matada — m2 M1)*: el fichero de D da **1** evento. Un fixture con la última línea sin LF da 0 eventos de esa línea; al añadirle el LF,
  la pasada siguiente da 1.
- **SC-005** *(partidas, sintético — D-3)*: `inputTokens` 100, `cacheReadTokens` 40, `cacheWriteTokens` 2, `outputTokens` 7 y `reasoningTokens` 3 dan
  `tokens_input` **58**, `tokens_output` **7**, `tokens_cache_read` 40 y `tokens_cache_creation` 2. Con el razonamiento en la salida, o sin restar la
  escritura, el test cae.
- **SC-006** *(sólo lectura — D-6)*: sobre un `session-state/` sintético con `.lock`, una carpeta sin `events.jsonl` y un fichero creciendo, una `--run` deja
  **iguales** las md5 de todos los ficheros y el listado de la raíz. Ningún aviso por la carpeta vacía. En W1, lo mismo contra la carpeta real **con Copilot
  cerrado**; con Copilot abierto, ningún fichero nuevo.
- **SC-007** *(binario)*: frente al binario de la **0.7.0** *(`15c4c47`)*, con las banderas de `.goreleaser.yaml`, crece **≤ 100 KB** en **las 5
  plataformas**, con el instrumento de 010 E-3 *(en linux/arm64, Σ `FileSiz` de los `LOAD`)*. Si no cabe, se adelgaza y el tope no sube.
- **SC-008** *(dependencias)*: `go.mod` sin `require`, no hay `go.sum`, y `go list -m all` da una sola línea.
- **SC-009** *(compactación — D-2, m2 M2)*: el fichero de E da **4** eventos, uno de compactación con `tokens_input` 75, `tokens_output` 267, lectura 15 335 y
  escritura 1 105. Sintético: un `compaction_complete` con `success: false` da 0 registros; uno con `success: true` sin `compactionTokensUsed`, «incoherentes 1»;
  uno sin `requestId`, «sin identificador 1».
- **SC-010** *(clasificación, sintético)*: una línea de cada clase da `registros 6 · eventos 2 · compactaciones 1 · repetidas 1 · sin identificador 1 ·
  incoherentes 2 · sin modelo 1`. Las clases: evento sin modelo; compactación; el mismo `usageId` repetido; sin `usageId`; `cacheRead + cacheWrite > input`;
  partida no entera. Cumple la identidad de FR-020.
- **SC-011** *(identidad y privacidad)*:
  - la función nueva reproduce los vectores del contrato;
  - un fixture lleva **centinelas** en `usageId`, `requestId`, `sessionId`, el nombre de la carpeta, `cwd`, el texto de `user.message`, `system.message`,
    `summaryContent`, `checkpointPath` y `requestMessages`;
  - tras una pasada, **ninguno** está en la cola ni en lo transmitido;
  - en `state.json` sólo está la ruta del fichero, **dentro de su clave**.
- **SC-012** *(ubicación — D-5)*: `COPILOT_HOME=<dir>` → `<dir>/session-state`; `COPILOT_HOME=""` → `<hogar>/.copilot/session-state`. `os.Getenv` en
  producción → **5**.
- **SC-013** *(tramos — P-3)*: un fixture con `start` *(cwd X)*, registro, `resume` *(cwd Y)*, registro da dos `project_ref` distintos y un solo `session_ref`.
  Leído en dos pasadas, cortado entre los dos registros, da **las mismas** referencias que de una vez.
- **SC-014** *(frontera — P-2)*: un fichero truncado por debajo de su offset se relee, y sus registros salen con los **mismos** `event_id`.
- **SC-015** *(sin Copilot CLI)*: sin carpeta ni variable, la salida de `--run` sobre los fixtures de las otras cuatro herramientas es **byte a byte** la del
  binario de `15c4c47`.
- **SC-016** *(puertas)*: frontera con 0 bytes de diff *(FR-027)*; los tests de `15c4c47`, verdes y **sin tocar** *(`git diff 15c4c47 --stat -- '*_test.go'`
  sólo con ficheros nuevos)*; `golangci-lint run` → 0.
- **SC-017** *(textos)*: los textos nuevos, byte a byte iguales a los aprobados, también sobre el binario publicado.
- **SC-018** *(no rompe a los demás)*: un fichero ilegible junto a fixtures sanos de las cuatro herramientas: salen los eventos sanos, stderr lleva
  `copilot_cli: fichero omitido: …`, y `state.json` se guarda sin avanzar ese offset.
- **SC-019** *(orden de escritura y sin otras herramientas)*: con el forzado de 008 SC-015 *(`Save` falla)*, la cola tiene los eventos de Copilot CLI, y la
  pasada siguiente los reencola con los **mismos** `event_id`; sólo con Copilot CLI, `--run` encola y termina con 0.
- **SC-020** *(privacidad del repo público)*: en el diff de 011, el `grep` de privacidad de los encargos da **0**; `testdata` sólo lleva ficheros
  **sintéticos**; nada del contenido de las copias se publica.
- **SC-021** *(el evento — FR-016 a FR-018)*: el 100 % de los eventos `copilot_cli` llevan `cost_usd = 0` y `cost_available = false`; `occurred_at` = la
  `timestamp` de la línea en UTC; `session_ref` = `event.Ref(sal, sessionId)` y `project_ref` = `Derivar(cwd, sal)`; sin `start`, `session_ref` vacío.
- **SC-022** *(ensayo en Windows)*: con Copilot CLI escribiendo, el agente emite cada registro **una vez**, cumple SC-006, y sus eventos cuadran **al token**
  con un contador independiente:
  - **W1**, antes de publicar *(sandbox, `COPILOT_HOME` propio)*: además, **Σ créditos × 0,01** de la línea de control = Σ eventos × la tarifa de SC-002;
  - **W2**, en la instalación real, después de publicar: el **coste que calcula la plataforma** para esos eventos = **Σ créditos × 0,01**, al redondeo por
    evento de la plataforma.

## Textos propuestos *(P-8, P-9)*

**Resumen de Copilot CLI** *(stderr, una línea; en `--run`, tras la de OpenCode; en el demonio, sólo con novedades)*:
```
copilot_cli: registros %d · eventos %d · compactaciones %d · repetidas %d · sin identificador %d · incoherentes %d · sin modelo %d
```
**Fichero omitido** *(stderr, uno por fichero)*:
```
copilot_cli: fichero omitido: %v
```
**Control** *(sólo en `--scan`, stdout, tras el resumen; D-4)*:
```
copilot_cli (control, no se envía): créditos según Copilot %d nanoAIU · equivalente $%.9f a 0,01 por crédito
```
«Créditos» = Σ `copilotUsage.totalNanoAiu` de los registros que dieron evento; el equivalente, `nanoAIU / 10⁹ × 0,01`.

## Preguntas abiertas *(con cómo se cerrarían)*

| # | Pregunta | Mientras tanto | Cómo se cierra |
|---|---|---|---|
| **Q-1** | ¿Un **subagente** o un *workflow* factura fuera de `usage_record` *(`agentMetrics` ≠ `main`, `session.db` con `*_credit_charges`)*? | Límite declarado *(P-10)* | Una sesión que lance un subagente, en copia congelada: Σ registros frente a `shutdown.totalNanoAiu`, como m2 M2 |
| **Q-2** | ¿La **compactación automática** deja el mismo `compaction_complete`? | Se lee si lo deja | Una sesión larga que llegue al límite de contexto |
| **Q-3** | ¿Cambian los registros en **Linux y macOS**, o con un plan de pago *(`isByok`, otros modelos)*? | Mismo lector | Una sesión corta en cada uno, con SC-001 rehecho sobre una copia nueva |

## Censo de tests — autorización para la fase de tareas

**Ninguno de los existentes cambia.** Los nuevos van en ficheros nuevos, con fixtures **sintéticos** nuevos en `testdata/`.

| Dónde | Nace |
|---|---|
| `internal/config/` | `copilot.go` y `copilot_test.go` *(D-5, SC-012)* |
| `internal/ingest/` | `copilot.go` *(una línea)*, `copilot_contexto.go` *(fichero, prefijo y recuentos)*, `copilot_eventid.go`; `copilot_eventid_test.go`, `copilot_registro_test.go`, `copilot_fichero_test.go`; y `testdata/copilot/` |
| `cmd/permea/` | `copilot.go` y `copilot_test.go` *(`--run`, demonio, `--scan`, SC-015, SC-018, SC-019)* |
| `internal/testutil/` | `sandbox_copilot_test.go`, como `sandbox_gemini_test.go` |
| `specs/011-lector-copilot/contracts/` | `event-id-copilot.md` |

**Fuera de los `_test.go`** *(M-3)*: `internal/testutil/sandbox.go` gana `t.Setenv("COPILOT_HOME", "")`; e `internal/config/codex.go:16`, `gemini.go:17` y
`opencode.go:18`, sólo el comentario *(«una de las CINCO»)*. `sandbox_test.go` comprueba `HOME`, `USERPROFILE` y `XDG_CONFIG_HOME`: la nueva no cambia su
veredicto.

## Fuera de alcance

| # | Qué | Por qué |
|---|---|---|
| **N-1** | El coste y las tarifas; los créditos en el evento | D-4 |
| **N-2** | Las llamadas auxiliares | D-2 |
| **N-3** | `shutdown`, `usage_checkpoint`, `model.*` y cualquier línea con contenido | D-2 |
| **N-4** | `workspace.yaml`, `checkpoints/`, `rewind-file-snapshots/`, `session.db` y los `.lock` | D-2, D-6, P-7 |
| **N-5** | Copilot dentro de VS Code y JetBrains | D-1: lo leerá la plataforma por API |
| **N-6** | Leer desde WSL las sesiones de Windows | D-5 lee el hogar propio |
| **N-7** | Un interruptor | P-6 |
| **N-8** | Un alias de modelos en la plataforma | Otra feature, como 010 N-10 |

## Límites declarados

- **D-2**: las llamadas auxiliares *(0 créditos hoy, en el plan Free)* no se cuentan: en tokens, el lector queda por debajo de lo consumido.
- **Sesión matada**: la llamada que estaba en curso no está en disco y no se cuenta *(m2 M1)*.
- **P-2**: un fichero reescrito con el mismo tamaño o mayor no se detecta. No se ha visto: el fichero es de sólo añadir *(desc. Q1; m2 M2)*.
- **P-10**: subagentes y *workflows*, compactación automática, planes de pago y otros SO, **sin muestra** *(Q-1 a Q-3)*.
- **Medido sólo con Copilot CLI 1.0.97, en Windows y en el plan Free**, con un único modelo *(`gpt-6-luna`)*.

## Assumptions
1. **El fichero es de sólo añadir** *(desc. Q1; m2 M2)*. Si no lo fuera, cuenta lo que diga P-2.
2. **`usageId` y `requestId` son únicos y estables por llamada** *(9 de 9 y 1 de 1; m2 M4)*.
3. **Una línea, una vez escrita, no cambia** *(desc. Q1)*.
4. **La regla de ubicación es la misma en los tres SO** *(documentación de GitHub; medida sólo en Windows, m2 M3)*.

## Dependencias
- **La plataforma**:
  - `tool = "copilot_cli"` como texto libre;
  - la tarifa de **`gpt-6-luna`** *(la de SC-002)* antes de W2;
  - **los nombres de modelo de los planes de pago** *(sin decidir)*: Copilot ofrece modelos con nombres como **`claude-haiku-4.5`** *(con punto)*, que **no
    casan** con el catálogo *(`claude-haiku-4-5`)*. El agente no traduce *(P-5)*: sin un alias o una normalización en la plataforma, esos eventos llegan
    «ciegos y señalizados» y quedan sin reparar.
- **006**: el esquema del `event_id`, sin tocarlo. **008**: el patrón de lector por offset, prefijo y recuentos, sin cambiar su conducta.

## Registro de enmiendas

| # | Fecha | Qué cambia | Por qué |
|---|---|---|---|
| **Borrador** | 2026-10-14 | Primera redacción, con D-1 a D-6 del dueño, y P-1 a P-11 y Q-1 a Q-3 para ratificar | Encargo 011 · la spec · 3/n |
