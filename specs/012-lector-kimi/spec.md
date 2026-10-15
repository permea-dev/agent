# 012 · «Lector de Kimi Code CLI» — Especificación

**Feature Branch**: `012-lector-kimi` · **Created**: 2026-10-15 · **Status**: **Ratificada el 2026-10-15** *(D-1 a D-6, P-1 a P-13 y SC-001 a SC-025, del dueño; **E-1** en P-4, con SC-026)* ·
**Cabeza de partida**: `a7faf5b` *(sobre `e9e5449`, cierre de la 0.8.0)*
**Input**:
- las decisiones del dueño D-1 a D-6 *(2026-10-15)*;
- el descubrimiento del 15-10 *(`soporte/descubrimiento.md`, Q1–Q11, leído en el código de Kimi Code **2.2.0**)*;
- la muestra 1 *(`soporte/descubrimiento-muestra-1.md`, M1–M9; «m1 n»)* y la muestra 2 *(`soporte/descubrimiento-muestra-2.md`, M1–M6; «m2 n»)*;
- el cuadre del dueño con la consola de Kimi Platform *(§Contexto 5)*.

> ⛔ Toda afirmación sobre nuestro código lleva `fichero:línea`, medida sobre `a7faf5b` *(el código es el de `e9e5449`)*. Las de Kimi citan el descubrimiento
> *(«desc. Qn»)* o las muestras *(«m1 n», «m2 n»)*. Lo que añade la spec sin que nadie lo dijera va marcado **`DECIDÍ YO`**. Ningún dato procede del contenido
> de las sesiones: sólo recuentos, tokens, nombres de modelo, claves, tipos y horas.

## Contexto — el agente lee Claude Code, Codex, Gemini CLI, OpenCode y Copilot CLI

1. **Cinco lectores hoy**, en `generate()`, antes del **único** `st.Save` *(`cmd/permea/main.go:414`)*. El último es Copilot CLI *(`main.go:403-412`)*.
2. **Kimi Code escribe un JSONL por agente**: `<KIMI_CODE_HOME | ~/.kimi-code>/sessions/<wd_*>/<sesión>/agents/<agente>/wire.jsonl` *(desc. Q1; m1 §La copia)*.
   - **Cada petición terminada deja una línea `usage.record`**: `{type, agentId, model, usage{inputOther, output, inputCacheRead, inputCacheCreation}, usageScope, time}`,
     sin `raw` y **sin id** *(m1 1; desc. Q2)*. La compactación deja uno más, con `usageScope: session` *(m1 4)*.
   - **Un subagente tiene su propio `wire.jsonl`**, y el `subagent.completed.usage` del principal repite su suma *(m1 5)*.
   - **`/fork` copia byte a byte el `wire.jsonl` del origen** y añade una línea `forked`; lo nuevo va detrás *(m2 3)*. Una petición matada no deja registro *(m1 6)*.
3. **Es casi el lector de Copilot CLI**: una línea = un evento; lectura por offset con `state.Recorrer` *(`internal/state/state.go:121-169`)*; el prefijo
   `[0, offset)` se relee con un prefiltro de bytes, como `leerPrefijoCopilot` *(`internal/ingest/copilot_contexto.go:195`)*. **Cambian**:
   - el `event_id`, por posición, porque Kimi no da id, y la **huella de la cabeza** en `state.json` *(E-1)*;
   - la frontera `forked`;
   - el proyecto, que sale de `session_index.jsonl` y no del fichero;
   - y las partidas, que pasan tal cual.
4. **La plataforma ya acepta el evento tal cual**: `tool` es texto libre, un modelo sin tarifa llega «ciego y señalizado», y un `event_id` repetido se descarta *(009 §Contexto 4)*.
5. **Cuadre con la consola** *(dato del dueño, 15-10)*: «Billing Details → Request Details» lista **13 peticiones** entre las 11:02:42 y las 11:16:38 *(hora de Madrid)*.
   - **12 casan token a token** con los 12 `usage.record` de las muestras, con `entrada de la consola = inputOther + inputCacheRead` y `caché = inputCacheRead`:
     Σ **33 831 · 186 112 · 0 · 3 170** *(entrada sin caché · lectura · escritura · salida)*. **La salida del registro es la que se cobra.**
   - **La 13.ª es S6, matada** *(11:08:12; 18 786 de entrada, 18 176 de caché y 157 de salida)*: **sin registro**.
   - **Coste previsto** a la tarifa oficial: **0,0802 $** para lo que verá Permea, y **≈ 0,0848 $** para lo que cobra Kimi *(la factura diaria se actualiza a las 7:00 del día siguiente)*.

## Decisiones del dueño *(2026-10-15; no se discuten aquí)*

| # | Decisión |
|---|---|
| **D-1** | `tool = "kimi_code"` |
| **D-2** | **Fuente**: los `usage.record` de **todos** los `wire.jsonl` de `sessions/`, el del principal y los de los subagentes, **con** el de la compactación *(`usageScope: session`)*. `subagent.completed.usage` se **ignora**, porque repite la suma del subagente *(m1 5)* |
| **D-3** | **`event_id` por posición**, con la forma de `permea/event_id/v1` de 011: sesión, agente y ordinal del registro en su fichero; **ni horas ni partidas en el hash**. **Con la frontera `forked`**: en un fichero con línea `forked`, lo anterior a la **última** es heredado y **no se emite** *(m2 3, m2 5)* |
| **D-4** | **El modelo es el alias sin su prefijo de proveedor**: `moonshot-ai/kimi-k2.7-code` → `kimi-k2.7-code`, el id que factura Kimi *(consola, 13/13)*. Un alias sin barra va tal cual. Los casos raros, en P-9 |
| **D-5** | **Sin coste**: `cost_usd = 0` y `cost_available = false`, como 008–011; la plataforma lo calcula con su tarifa *(constitución 1.1.0, §II)*. **Dependencia**: la fila de `kimi-k2.7-code` en `pricing.php` *(0,95 · 0,19 · 4,00 USD por millón)*, **desplegada antes de la etiqueta v0.9.0** |
| **D-6** | **Límite declarado**: una sesión matada pierde la petición en vuelo. Medido: S6 se cobró y no dejó `usage.record` *(m1 6; §Contexto 5)* |

## Decisiones de método *(heredadas de 008–011)*

| # | Decisión |
|---|---|
| **M-1** | El evento no cambia: mismos campos y `SchemaVersion = 1` *(`internal/event/event.go:12-35`)*. La frontera de FR-027, con **0 bytes** de diff |
| **M-2** | Versión **`0.9.0`** *(encargo del dueño)*. Los otros cinco lectores, exactamente como en la 0.8.0: los tests de `e9e5449` siguen verdes **sin tocarlos** |
| **M-3** | `internal/testutil/sandbox.go` fija también `KIMI_CODE_HOME` *(vacía)*, junto a las otras cinco *(`sandbox.go:63-70`)* |
| **M-4** | Los errores son **por fichero**, sin la ruta *(`sinRuta`, `cmd/permea/copilot.go:60`)*, como 011. Encolar sigue siendo fatal |
| **M-5** | Es la **6.ª** lectura de entorno en producción, tras `CODEX_HOME`, `GEMINI_CLI_HOME`, `XDG_DATA_HOME`, `OPENCODE_DB` y `COPILOT_HOME` *(`internal/config/codex.go:19`, `gemini.go:20`, `opencode.go:21`, `:25`, `copilot.go:19`)* |

## Propuestas — **ratificadas** por el dueño el 2026-10-15 *(tal cual, con sus `DECIDÍ YO`; P-4, con la enmienda E-1)*

| # | Propuesta | Alternativas | Recomendación |
|---|---|---|---|
| **P-1** ✅ **(a)** `DECIDÍ YO` | **Raíz**: `$KIMI_CODE_HOME` si no es vacía; si no, `<hogar>/.kimi-code`, con el `os.UserHomeDir()` de las otras raíces. **Vacía cuenta como ausente**, como `COPILOT_HOME` *(`config/copilot.go:19`)* y como la CLI de Kimi *(desc. Q1)*. Las sesiones van en `<raíz>/sessions`, y el índice en `<raíz>/session_index.jsonl` | (b) vacía = definida, como el núcleo de Kimi *(`??`; desc. Q1)* | **(a)**: con (b), una variable vacía daría una ruta relativa al proceso del agente, que nada tiene que ver con la de Kimi |
| **P-2** ✅ **(a)** `DECIDÍ YO` | **Activación**: el lector está activo **en una pasada** si existe, en esa pasada, `<raíz>/sessions`. Se comprueba en cada `generate()`, sin interruptor | (b) un interruptor | **(a)**, como 008–011 |
| **P-3** ✅ **(a)** `DECIDÍ YO` | **Ficheros**: sólo `<raíz>/sessions/<a>/<b>/agents/<c>/wire.jsonl`, con `a`, `b` y `c` directorios **que no empiezan por `.`**, y `wire.jsonl` un fichero regular. Nada más: ni `.index-cache/`, `.index-dirty/` *(m1 §La copia)*, `state.json`, `logs/`, `notify/`, `file-history/`, `tasks/`, `cron/`, `.bak`, `.tmp.*` *(desc. Q7)*, ni `<raíz>/imported/` *(desc. Q6)* | (b) exigir el prefijo `wd_` a `a` | **(a)**: el `.bak` repetiría lo contado *(desc. Q7)*, e `imported/` son exportaciones con forma de sesión. (b) ataría el lector a un nombre que Kimi puede cambiar |
| **P-4** ✅ **(a)** · **E-1** | **Frontera: offset y huella de la cabeza, por fichero** *(dueño, 15-10)*. `state.json` guarda, por `wire.jsonl`, los cuatro campos y **`cabeza`** `DECIDÍ YO`: `event.Ref(sal, primera línea completa sin su LF)` *(64 hex, con sal, como las huellas de OpenCode, `state.go:26-29`)*, con `omitempty`. **No se abre** si tamaño = offset **y** `mtime` = el guardado `DECIDÍ YO`. Abierto, **se relee desde 0** si **encoge**, si la huella de su primera línea completa ≠ `cabeza`, o si `cabeza` falta con offset > 0 *(un estado sin el campo)* `DECIDÍ YO`; con D-3 por ordinal, lo releído da los **mismos** `event_id` y la plataforma descarta los repetidos. Línea final sin LF, no se consume *(`state.go:146-148`)*. **El ordinal y la última `forked`** salen de un prepaso del fichero entero con el prefiltro de P-10 | (b) sólo el offset *(el borrador)* · (c) la huella sin sal · (d) mirar la cabeza en **todas** las pasadas | **(a)**: detecta las dos reescrituras de Kimi, la reparación, que **encoge**, y la migración desde una versión anterior a 1.5, que **cambia la cabeza** *(desc. Q7)*. El ordinal es un **recuento**, no un número de línea, y sobrevive a las dos. (c) dejaría en `state.json` algo derivable sin sal. (d) abriría en cada pasada ficheros que no cambiaron *(FR-005)* |
| **P-5** ✅ **(a)** `DECIDÍ YO` | **`event_id`** *(D-3)*: `["permea/event_id/v1", "kimi_code", "respuesta", S, A, N]`, con la codificación de 006 *(`internal/ingest/eventid.go:31`, `:43`)* y `tipoRespuesta` *(`codex_eventid.go:29`)*, en un contrato nuevo, `contracts/event-id-kimi.md`. **S** = nombre de la carpeta de la sesión *(su `sessionId`; desc. Q1)*; **A** = nombre de la carpeta del agente *(= `agentId` del registro, 11/11; m1 1)*; **N** = el ordinal del `usage.record` en el fichero, en decimal, desde 1, **contando los heredados** | (b) **A** = el `agentId` de la línea · (c) **N** sin los heredados | **(a)**: las carpetas existen siempre, y no hay «sin identificador». Contar desde el principio no depende de dónde caiga la frontera *(un fork de un fork tiene dos `forked`)* |
| **P-6** ✅ **(a)** `DECIDÍ YO` | **Proyecto y sesión**: `project_ref` = `Resolutor.Derivar(W, sal)` *(`internal/project/resolve.go:318`)*, donde **W** es el `workDir` de la **última** línea de `session_index.jsonl` con `sessionId` = **S**. Antes de derivar, una **W con forma de Windows** *(`X:\`, `X:/` o `\\`)* pasa sus `/` a `\` *(m2 4: el fork escribe la misma ruta con otras barras)*. `session_ref` = `event.Ref(sal, S)` *(`internal/event/event.go:40-46`)*. **S sin línea en el índice**: `project_ref` vacío. El índice se lee **entero**, una vez por pasada | (b) el `cwd` de `state.json`, que lleva `title` y `lastPrompt` ⛔ · (c) el de `profile.bind`, con `systemPrompt` ⛔ · (d) normalizar sólo en Windows, con el `filepath.Clean` de `mejorValorDisponible` *(`resolve.go:123-131`)* | **(a)**: el índice sólo lleva rutas e ids *(m1 §La copia)*. (d) ya ocurre en Windows, pero no se puede probar en Linux, y una copia de Windows leída en Linux daría dos proyectos. Sin minúsculas *(`C:` frente a `c:`)*, como `resolve.go:119-122`: queda en §Límites |
| **P-7** ✅ **(a)** `DECIDÍ YO` | **Momento**: `occurred_at` = `time` *(milisegundos desde la época; m1 1)*, en UTC | (b) el `time` de la `llm.request` anterior *(el inicio)* | **(a)**: es cuando Kimi registra el consumo |
| **P-8** ✅ **(a)** `DECIDÍ YO` | **Registros a 0** *(las cuatro partidas a 0: `emptyUsage`; desc. Q3)*: **no se emiten** y se cuentan como «a cero» | (b) emitirlos | **(a)**: no hay nada que tarificar. En las muestras hay 0 *(m1 7)*, pero su recuento lo deja visible |
| **P-9** ✅ **(a)** `DECIDÍ YO` | **Modelo** *(D-4)*: el texto **tras la última `/`** del `model` del registro. Sin `/`, tal cual. Si queda vacío *(alias vacío o acabado en `/`)*, o `model` no es texto, el evento sale con `model` **vacío** y cuenta «sin modelo». Así, `a/b/c` → `c`, `/x` → `x` y `x/` → vacío | (b) tras la **primera** `/` *(`a/b/c` → `b/c`)* · (c) quitar el prefijo sólo con **una** `/` | **(a)**: el catálogo de la plataforma usa el último segmento *(`kimi-k2.7-code`)*, y un enrutador escribe `proveedor/fabricante/modelo`. (b) dejaría `b/c`, que no casaría con ninguna tarifa |
| **P-10** ✅ **(a)** `DECIDÍ YO` | **Prefiltro**: una línea se decodifica sólo si **empieza** por `{"type":"usage.record"` *(11/11 en m1 1)*; en el prepaso de P-4, también por `{"type":"forked"`. Lo confirma el JSON. De un `usage.record` sólo se leen `model`, `usage.{inputOther, output, inputCacheRead, inputCacheCreation}`, `usageScope` y `time`; de `forked`, sólo `type` | (b) `bytes.Contains`, como 011 P-11 | **(a)**: `type` va primero en los registros *(desc. Q2)*, y anclar al principio de la línea descarta el mismo texto dentro de un contenido. Las líneas ⛔ *(`context.append_message`, `llm.request.systemPrompt`, `subagent.completed.resultSummary`…)* no se decodifican |
| **P-11** ✅ `DECIDÍ YO` | **Resumen y textos**: los de §Textos. La línea `kimi_code: …` va tras la de Copilot CLI *(`main.go:492`)*; en el demonio, sólo con novedades *(`main.go:563-565`)* | corregirlos | — |
| **P-12** ✅ **(a)** `DECIDÍ YO` | **`--scan <fichero>`**: un fichero cuya **primera línea** es `{"type":"metadata"}` con `protocol_version` textual y `created_at` numérico se lee como Kimi Code, **tras** Copilot CLI *(`main.go:618-624`)*. S y A salen de la ruta *(`<S>/agents/<A>/wire.jsonl`)*; sin índice, `project_ref` vacío. Escribe las líneas `evento:` *(`imprimirEventos`, `main.go:697`)* y el resumen. **Sin línea de control**: no hay créditos *(D-5)*. Nunca escribe estado ni cola | (b) leer el índice de la raíz que contenga el fichero · (c) sin `--scan` | **(a)**: es el contraste de W1 sobre un fichero suelto, con la frontera `forked` aplicada |
| **P-13** ✅ **(a)** `DECIDÍ YO` | **Índice ilegible** *(existe y no se puede leer)*: `kimi_code: fichero omitido: %v`, y en esa pasada **no se lee ningún `wire.jsonl`** *(sus offsets no avanzan)*. **Índice ausente**: se lee todo, con `project_ref` vacío | (b) leer los `wire.jsonl` con `project_ref` vacío | **(a)**: un evento emitido sin proyecto ya no se corrige; mejor esperar a la pasada siguiente |

## User Scenarios & Testing

### Historia 1 — El consumo de Kimi Code llega a la plataforma *(P1)*
1. **Dada** una sesión con dos peticiones terminadas, **cuando** corre `--run`, **entonces** salen **dos** eventos `tool = kimi_code` con las partidas de FR-013,
   `model = kimi-k2.7-code` y `cost_available = false`.
2. **Dada** la misma sesión en una segunda pasada sin cambios, **entonces** no sale nada y el fichero **no se abre**.

### Historia 2 — Un `/fork` no cuenta dos veces *(P1)*
1. **Dada** una sesión con 2 registros y su fork con 1 petición nueva, **entonces** salen **3** eventos: los 2 heredados del fork se cuentan «heredados» y no se emiten.

### Historia 3 — El subagente y la compactación se cuentan una vez *(P1)*
1. **Dado** un subagente con 3 peticiones, **entonces** salen 3 eventos de su `wire.jsonl` y **ninguno** de `subagent.completed`.
2. **Dado** un `/compact`, **entonces** sale **un** evento más, contado en «compactaciones».

### Historia 4 — Una sesión matada no pierde lo ya escrito *(P1)*
1. **Dada** una sesión matada a mitad de su 2.ª petición, **entonces** sale la 1.ª, y nada de la cortada *(D-6)*.

### Historia 5 — El agente nunca toca la carpeta de Kimi *(P1)*
1. **Cuando** corre una pasada, **entonces** todos los ficheros bajo la raíz tienen la **misma md5** antes y después, y no aparece ninguno nuevo.

### Historia 6 — Quien no usa Kimi Code no nota nada *(P1)*
1. **Dada** una instalación sin `~/.kimi-code/sessions` ni `KIMI_CODE_HOME`, **entonces** la salida de `--run` es **byte a byte** la de la 0.8.0.

### Historia 7 — Kimi Code sin las otras herramientas *(P2)*
1. **Dada** sólo la carpeta de Kimi Code, **entonces** sus eventos salen igual, y `--run` termina con 0.

### Edge Cases
- **Reanudar** *(`--continue`, `--session`, `--resume`)*: añade al mismo fichero sin reescribir, y sin bloque de arranque *(m2 3)*. Los ordinales siguen.
- **Abrir una sesión y bifurcar sin preguntar**: el origen no cambia *(m2 2)*.
- **Fork de un fork**: dos `forked`; manda la **última** *(D-3)*. **Ventana de la copia**: entre copiar el fichero y añadir `forked` *(desc. Q6)*, una pasada vería
  los heredados como propios. Son milisegundos: §Límites.
- **`/undo`**: las ramas abandonadas se quedan en el fichero, y sus registros se emiten, porque se facturaron *(desc. Q6)*.
- **Subagente**: su fichero lleva su A y su ordinal, y comparte S, `session_ref` y proyecto con el principal *(m1 5)*.
- **`subagent.completed`, `llm.request`, `context.*`**: no pasan el prefiltro *(P-10)*; no se decodifican ni se cuentan.
- **Registro a 0**: «a cero», no se emite *(P-8)*. **Partida no entera, negativa o ausente del todo**, o `time` no numérico: «incoherente».
- **`model` con varias `/`, vacío o no textual**: P-9.
- **Sesión sin línea en el índice**: `project_ref` vacío *(P-6)*. **Índice ilegible**: P-13.
- **`workDir` con otras barras** *(el del fork; m2 4)*: el mismo `project_ref` *(P-6)*.
- **Fichero reparado por Kimi** *(más corto; deja un `.bak`)*: se relee desde 0 con los mismos `event_id`; el `.bak` no se lee *(P-3, P-4)*.
- **Fichero migrado desde una versión anterior a 1.5** *(crece y cambia la cabeza)*: se relee desde 0, con los mismos `event_id` *(P-4, E-1)*.
- **Una línea que no es JSON y pasa el prefiltro**: `skip (línea corrupta): …`, como Copilot *(`copilot_contexto.go:149`)*; no corta el fichero.
- **El mismo `event_id` dos veces en una pasada** *(una carpeta de sesión copiada en otra `wd_*`)*: sale una vez y se cuenta «repetidas».
- **`KIMI_CODE_HOME` apunta a algo sin `sessions/`**: el lector no está activo; no dice nada y no falla.
- **El proceso cae entre encolar y guardar**: la pasada siguiente reencola los mismos `event_id`, y la plataforma los descarta.

## Requirements

### A · Dónde y cuándo lee *(P-1 a P-4)*

- **FR-001** *(raíz — P-1)*: según P-1. Se resuelve una vez por proceso, en `setup()`, como las demás *(`main.go:235-255`)*.
- **FR-002** *(activación — P-2)*: según P-2. Si no está activo, el lector no escribe nada, no devuelve error y no cambia ningún texto.
- **FR-003** *(ficheros — P-3)*: según P-3.
- **FR-004** *(estado — E-1)*: cada `wire.jsonl` tiene una entrada en `state.json`, con la ruta como clave, los **cuatro** campos de siempre *(`state.go:16-20`)*
  y **`cabeza`** *(P-4)*, **sin** `Frontera`. Las entradas de los otros lectores no la llevan, y su JSON no cambia. El índice no tiene entrada: se lee entero.
- **FR-005** *(entre pasadas — P-4)*: un registro **DEBE** salir **una vez**, salvo tras una relectura desde 0 *(encogido o cabeza cambiada)*, que lo reemite
  con el mismo `event_id`. Un fichero con el mismo tamaño y el mismo `mtime` **no se abre**.

### B · Cómo lee *(P-3, P-6, P-10)*

- **FR-006** *(sólo lectura)*: el lector **NUNCA** escribe, crea, bloquea, renombra ni borra nada bajo la raíz. Abre `wire.jsonl` y el índice en sólo lectura.
- **FR-007** *(prefiltro — P-10)*: según P-10.
- **FR-008** *(posición — P-4, P-5)*: un registro leído en una pasada posterior **DEBE** tener el mismo ordinal y la misma condición de heredado que leyendo
  el fichero de una vez.
- **FR-009** *(líneas largas)*: sin tope de longitud *(`ReadBytes`, `state.go:145`)*: las líneas ⛔, como `llm.tools_snapshot`, pueden ser largas *(desc. Q2)*.

### C · El evento *(D-1 a D-5, P-5 a P-9)*

- **FR-010** *(unidad — D-2)*: es **registro** cada línea `usage.record` de un fichero de P-3. Cada registro se clasifica *(FR-020)* y produce **como mucho un** evento.
- **FR-011** *(identidad — D-3, P-5)*: el `event_id` es el de `contracts/event-id-kimi.md`, y **DEBE** reproducir sus vectores byte a byte.
- **FR-012** *(lo que se lee)*: los campos de P-10; del índice, `sessionId` y `workDir`. **Nunca** `agentId` ni `sessionDir`, ni ninguna otra línea.
- **FR-013** *(partidas)*: **tal cual, sin sumas ni restas**: `tokens_input` = `inputOther`, `tokens_cache_read` = `inputCacheRead`,
  `tokens_cache_creation` = `inputCacheCreation` y `tokens_output` = `output` *(12/12 con la consola; §Contexto 5)*. Una partida ausente vale 0; si faltan las cuatro, «incoherente».
- **FR-014** *(heredados — D-3)*: un registro anterior a la última `forked` de su fichero no se emite y se cuenta «heredados».
- **FR-015** *(modelo — D-4, P-9)*: según P-9.
- **FR-016** *(momento — P-7)*: según P-7.
- **FR-017** *(herramienta y coste — D-1, D-5)*: `tool = "kimi_code"`, `cost_usd = 0` y `cost_available = false`, **siempre**. No consulta `internal/pricing`.
- **FR-018** *(proyecto y sesión — P-6)*: según P-6. `machine_ref`, `dev_id`, `org_id` y `agent_version`, como en Claude Code.
- **FR-019** *(nada del proveedor viaja)*: **NUNCA** van a ningún campo, a la cola ni a nada que se transmita: S, A, `workDir`, `sessionDir`, el nombre de la
  carpeta `wd_*` *(lleva el nombre del proyecto; desc. Q1)*, ni ningún texto. S y `workDir` sólo viajan derivados con sal. La ruta es la clave de `state.json`, **en local**.

### D · Clasificación y textos *(P-8, P-11, P-12)*

- **FR-020** *(la cuenta)*: en cada pasada, `registros = eventos + heredados + repetidas + a cero + incoherentes`. «Compactaciones» *(`usageScope` = `session`)* y
  «sin modelo» son **subconjuntos** de «eventos». **Orden**: heredado → incoherente → a cero → repetida → evento.
- **FR-021** *(resumen)*: con el lector activo, `--run` escribe **una** línea más en stderr *(§Textos)*, tras la de Copilot CLI o tras la anterior que haya. El
  demonio, sólo si en ese ciclo hubo registros *(`HayNovedades`)*.
- **FR-022** *(`--scan` — P-12)*: según P-12. Nunca escribe estado ni cola.
- **FR-023** *(lo de antes no cambia)*: sin Kimi Code activo, la salida de `--run`, del demonio y de `--scan` es **byte a byte** la de la 0.8.0. Con él, sólo se
  añade la línea de FR-021, y «N eventos encolados» cuenta también los suyos.

### E · Frontera, durabilidad y errores

- **FR-024** *(orden)*: Kimi Code se lee **después** de Copilot CLI *(`main.go:403-412`)* y **antes** del único `st.Save` *(`main.go:414`)*.
- **FR-025** *(sin las otras herramientas)*: una instalación sólo con Kimi Code **DEBE** enviar sus eventos.
- **FR-026** *(no rompe a los demás — M-4)*: un error al leer un fichero escribe `kimi_code: fichero omitido: %v` *(sin ruta)*, **no** avanza su offset, **no**
  corta la pasada y **no** impide guardar el estado de los demás. El índice ilegible, según P-13. Encolar sigue siendo fatal.
- **FR-027** *(frontera — M-1)*: `git diff e9e5449`, **vacío**, de:
  - `internal/event/`, `internal/ingest/eventid.go`, `eventid_test.go` y `boundary_test.go`;
  - los `*_eventid.go` y `*_eventid_test.go` de Codex, Gemini, OpenCode y Copilot;
  - los contratos de 006 y 008–011.
- **FR-028** *(versión — M-2)*: `0.9.0`. Lo demás, exactamente como en la 0.8.0.

### Key Entities
- **Registro de Kimi Code**: una línea `usage.record` de un `wire.jsonl`. Su posición *(S, A, N)* le da identidad, y lleva `time`, el alias y las cuatro partidas.
- **Frontera `forked`**: la última línea `forked` de un fichero. Lo anterior es heredado.
- **Índice**: `session_index.jsonl`, de líneas `{sessionId, sessionDir, workDir}` *(m2 4)*. Da W por S.
- **Estado de un fichero**: los cuatro campos de `state.json`, con `Offset` al final de la última línea completa, y `cabeza` *(E-1)*.

## Success Criteria

Las referencias son las copias `congelada-2026-10-15` *(«A»)* y `congelada-2026-10-15b` *(«B»)*, leídas en sólo lectura. Sus huellas **no** se publican.

- **SC-001** *(contra B)*: `--run` en un sandbox con `KIMI_CODE_HOME` = la raíz de B, sin escribir en ella:
  - **12** eventos `kimi_code`, los 12 con `kimi-k2.7-code`; **1** de compactación y 3 del subagente;
  - `tokens_input` **33 831**, `tokens_cache_read` **186 112**, `tokens_cache_creation` **0** y `tokens_output` **3 170**;
  - resumen: `kimi_code: registros 14 · eventos 12 · heredados 2 · repetidas 0 · a cero 0 · incoherentes 0 · compactaciones 1 · sin modelo 0`;
  - **4** `session_ref` distintos y **1** `project_ref` *(el fork comparte proyecto con su origen; m2 4)*; 0 eventos de S6 y de `subagent.completed`.
  - Las copias conservan sus sha256.
- **SC-002** *(incremental, A → B)*: sobre una copia temporal de A, `--run` da **11** eventos; sustituida por B, la pasada siguiente da **1** *(el del fork)* y
  `heredados 2`. **No se abre** ningún fichero que no cambió.
- **SC-003** *(coste previsto — D-5)*: Σ eventos de SC-001 × la tarifa de `kimi-k2.7-code` = **0,08018073 $** *(0,0802)*. La consola cobra ≈ 0,0848 $: la
  diferencia es S6 *(D-6)*.
- **SC-004** *(sesión matada — m1 6)*: el fichero de S6 da 0 eventos y `registros 0`. Un fixture con la última línea sin LF da 0 eventos de esa línea; con el LF, la
  pasada siguiente la emite.
- **SC-005** *(partidas, sintético — FR-013)*: `inputOther` 100, `inputCacheRead` 40, `inputCacheCreation` 2 y `output` 7 dan **100 · 40 · 2 · 7**. Restando la caché
  de la entrada *(la regla de Copilot)*, el test cae.
- **SC-006** *(frontera `forked`, sintético — D-3)*: 2 registros + `forked` + 1 registro → 1 evento y «heredados 2»; con dos `forked`, cuenta la última; sin `forked`,
  todo. Leído en dos pasadas, cortado tras el `forked`, da lo mismo que de una vez.
- **SC-007** *(identidad)*: la función nueva reproduce los vectores del contrato. Tras truncar un fichero por debajo de su offset, sus registros salen con los
  **mismos** `event_id`.
- **SC-008** *(modelo — P-9)*: `moonshot-ai/kimi-k2.7-code` → `kimi-k2.7-code`; `kimi-k2.7-code` → igual; `a/b/c` → `c`; `/x` → `x`; `x/`, vacío o no textual →
  `model` vacío y «sin modelo 1».
- **SC-009** *(subagente — D-2)*: un principal con `subagent.completed.usage` y un `agent-0` con 3 registros dan 3 eventos, con el mismo `session_ref` que el principal.
- **SC-010** *(clasificación, sintético)*: una línea de cada clase *(evento, compactación, sin modelo, heredado, a cero, partida negativa, `time` textual, una
  copia del fichero en otra `wd_*`)* da la identidad de FR-020, con los recuentos exactos fijados en el plan.
- **SC-011** *(privacidad)*: un fixture lleva **centinelas** en S, A, el nombre de `wd_*`, `workDir`, `sessionDir`, `context.append_message`, `profile.bind.systemPrompt`,
  `llm.request.systemPrompt` y `subagent.completed.resultSummary`. Tras una pasada, **ninguno** está en la cola ni en lo transmitido; en `state.json`, sólo la
  ruta, **dentro de su clave**.
- **SC-012** *(ubicación — P-1)*: `KIMI_CODE_HOME=<dir>` → `<dir>/sessions`; `KIMI_CODE_HOME=""` → `<hogar>/.kimi-code/sessions`. `os.Getenv` en producción → **6**.
- **SC-013** *(sólo lectura)*: sobre una raíz sintética con `.index-*`, `.bak`, `.tmp.*`, `imported/` y un fichero creciendo, una `--run` deja **iguales** las md5 y el
  listado. Ninguno de esos ficheros se abre *(contador en el test)*. En W1, lo mismo contra la raíz real con Kimi cerrado; con Kimi abierto, ningún fichero nuevo.
- **SC-014** *(proyecto — P-6)*: dos líneas del índice con la misma ruta de Windows, una con `\` y otra con `/`, dan el **mismo** `project_ref` en Linux y en Windows.
  Sin línea para S, `project_ref` vacío. Con el índice ilegible, 0 eventos y el aviso de P-13.
- **SC-015** *(sin Kimi Code)*: sin carpeta ni variable, la salida de `--run` sobre los fixtures de los otros cinco lectores es **byte a byte** la del binario de `e9e5449`.
- **SC-016** *(puertas)*: frontera con 0 bytes de diff *(FR-027)*; los tests de `e9e5449`, verdes y **sin tocar** *(`git diff e9e5449 --stat -- '*_test.go'` sólo
  con ficheros nuevos)*; `golangci-lint run` → 0.
- **SC-017** *(textos)*: los textos nuevos, byte a byte iguales a los aprobados, también sobre el binario publicado.
- **SC-018** *(no rompe a los demás)*: un `wire.jsonl` ilegible junto a fixtures sanos de los cinco lectores: salen los sanos, stderr lleva `kimi_code: fichero omitido: …`,
  y `state.json` se guarda sin avanzar ese offset.
- **SC-019** *(orden y sin otras herramientas)*: con el forzado de 008 SC-015 *(`Save` falla)*, la cola tiene los eventos de Kimi Code, y la pasada siguiente
  los reencola con los **mismos** `event_id`. Sólo con Kimi Code, `--run` encola y termina con 0.
- **SC-020** *(privacidad del repo)*: el patrón `PRIV` de 011 *(`tasks.md` §Puertas)* da **0** en el diff de 012 y, con `grep -raE`, en `testdata/kimi/` y su
  `generar.sh`. `testdata` sólo lleva ficheros **sintéticos**; nada de las copias se publica.
- **SC-021** *(el evento)*: el 100 % de los eventos `kimi_code` llevan `cost_usd = 0` y `cost_available = false`; `occurred_at` = `time` en UTC; `session_ref` =
  `event.Ref(sal, S)`, y `project_ref` según P-6.
- **SC-022** *(binario)*: frente al de la **0.8.0** *(`e9e5449`)*, con las banderas de `.goreleaser.yaml`, crece **≤ 100 KB** en las 5 plataformas, con el instrumento
  de 010 E-3. Si no cabe, se adelgaza y el tope no sube.
- **SC-023** *(dependencias)*: `go.mod` sin `require`, no hay `go.sum`, y `go list -m all` da una sola línea.
- **SC-024** *(`--scan`)*: sobre el `wire.jsonl` del fork de B da **1** evento y `heredados 2`; sobre el de `agent-0`, **3** eventos *(15 798 · 30 976 · 0 · 268)*.
- **SC-025** *(cuadre con la consola, en Windows)*: con Kimi Code escribiendo, el agente emite cada registro **una vez**, cumple SC-013, y **Σ de eventos = Σ de
  «Request Details», petición a petición**, con `entrada = tokens_input + tokens_cache_read`, `caché = tokens_cache_read` y `salida = tokens_output`. **Cada
  petición de la consola sin evento se declara como matada**, con su hora. Debe incluir un `/fork` y un subagente.
  - **W1**, antes de publicar *(sandbox, `KIMI_CODE_HOME` propio)*;
  - **W2**, en la instalación real, después de publicar y con la fila de D-5 desplegada: además, **el coste que calcula la plataforma** por evento = la tarifa
    oficial × sus partidas.
- **SC-026** *(cabeza — E-1, sintético)*: un fichero leído en una pasada y, antes de la siguiente, **(i)** con otra primera línea y más largo, **(ii)** más corto
  con el prefijo intacto, o **(iii)** con su entrada de `state.json` sin `cabeza`, da en la pasada siguiente **0 eventos nuevos** y **ningún `event_id` distinto** de
  los ya emitidos; y lo nuevo tras el cambio sale una vez. Un fichero sin cambios no se abre *(contador en el test)*.

## Textos propuestos *(P-11, P-12, P-13)*

**Resumen** *(stderr, una línea; en `--run`, tras la de Copilot CLI; en el demonio, sólo con novedades)*:
```
kimi_code: registros %d · eventos %d · heredados %d · repetidas %d · a cero %d · incoherentes %d · compactaciones %d · sin modelo %d
```
**Fichero omitido** *(stderr, uno por fichero, y uno por el índice ilegible)*:
```
kimi_code: fichero omitido: %v
```
`--scan` escribe las líneas `evento:` de siempre y el resumen; **no** hay línea de control *(D-5)*.

## Preguntas abiertas *(con cómo se cerrarían)*

| # | Pregunta | Mientras tanto | Cómo se cierra |
|---|---|---|---|
| **Q-1** ✅ *(cerrada: m2 3)* | ¿Reanudar reescribe o repite? | — | **No**: el fork reanudado añadió sin reescribir el prefijo |
| **Q-2** ✅ *(cerrada: §Contexto 5)* | ¿La salida incluye el razonamiento? | — | La salida del registro **es la que se cobra**, 12/12 |
| **Q-3** | En Windows, ¿un `wire.jsonl` abierto por el agente hace fallar la reescritura atómica de Kimi *(borrar y renombrar; desc. Q9)*? | Lecturas cortas; sin `FILE_SHARE_DELETE` | W1: una sesión reparada *(línea cortada)* mientras el demonio lee, y ver si Kimi avisa |
| **Q-4** | ¿Dos procesos sobre la misma sesión *(TUI y web)* intercalan líneas? | Mismo lector | Una sesión abierta en las dos a la vez, y SC-025 |
| **Q-5** | **Membresía OAuth** *(alias `kimi-code/k3` → `k3`)* y otros protocolos *(`anthropic`, con escritura de caché; desc. Q8)* | Mismo lector; P-9 | Una sesión con membresía y otra con `type = "anthropic"`; las tarifas, en §Dependencias |
| **Q-6** | ¿Cambia algo en Linux o macOS? | Mismo lector | Una sesión corta en cada uno, con SC-001 rehecho |

## Censo de tests — autorización para la fase de tareas

**Ninguno de los existentes cambia.** Los nuevos van en ficheros nuevos, con fixtures **sintéticos** nuevos en `testdata/`.

| Dónde | Nace |
|---|---|
| `internal/config/` | `kimi.go` y `kimi_test.go` *(P-1, SC-012)* |
| `internal/ingest/` | `kimi.go` *(una línea)*, `kimi_contexto.go` *(listado, índice, fichero, prefijo y recuentos)*, `kimi_eventid.go`; `kimi_eventid_test.go`, `kimi_registro_test.go`, `kimi_fichero_test.go`; y `testdata/kimi/` con `generar.sh` |
| `cmd/permea/` | `kimi.go` y `kimi_test.go` *(`--run`, demonio, `--scan`, SC-015, SC-018, SC-019)* |
| `internal/testutil/` | `sandbox_kimi_test.go`, como `sandbox_copilot_test.go` |
| `specs/012-lector-kimi/contracts/` | `event-id-kimi.md` |

**Fuera de los `_test.go`** *(M-3, M-5)*:
- `internal/testutil/sandbox.go` gana `t.Setenv("KIMI_CODE_HOME", "")`;
- en `cmd/permea/main.go`, los campos, `setup()`, `generate()`, `runOnce`, `tick` y `dryRun`;
- el comentario de `internal/config/codex.go:16`, `gemini.go:17`, `opencode.go:18` y `copilot.go:16` pasa a «una de las SEIS».

## Fuera de alcance

| # | Qué | Por qué |
|---|---|---|
| **N-1** | El coste y las tarifas en el agente | D-5 |
| **N-2** | `subagent.completed`, `llm.request`, `token_counting.*` y cualquier línea con contenido | D-2, P-10 |
| **N-3** | `state.json`, `logs/`, `tasks/`, `cron/`, `.index-*`, `.bak`, `imported/`, `config.toml` y `credentials/` | P-3; desc. Q1 |
| **N-4** | La cuota de la membresía *(`/usage`)* | No está en disco *(desc. Q2)* |
| **N-5** | Las llamadas auxiliares *(títulos y `ping`)* | No dejan `usage.record` *(desc. Q3)* |
| **N-6** | Leer desde WSL las sesiones de Windows | P-1 lee el hogar propio |
| **N-7** | Un interruptor | P-2 |
| **N-8** | El alias en la plataforma | P-9 quita el prefijo; el resto, otra feature *(como 011 N-8)* |

## Límites declarados

- **D-6**: una sesión matada pierde la petición en vuelo, **que sí se cobra** *(S6: 18 786 de entrada · 157 de salida; §Contexto 5)*.
- **Intentos fallidos o interrumpidos con Esc**: no dejan registro *(desc. Q3)*; si el proveedor los factura, el lector queda por debajo.
- **La ventana del fork**: si una pasada cae entre la copia y la línea `forked` *(desc. Q6)*, los heredados salen como propios.
- **P-4** *(E-1)*: una reescritura que **conserve la primera línea y no encoja**, o que deje el mismo tamaño y el mismo `mtime` *(en segundos)*, no se detecta. Kimi no la hace *(desc. Q7)*.
- **P-6**: dos grafías de la misma ruta en Windows que sólo difieren en mayúsculas dan dos proyectos *(desc. Q9; `resolve.go:119-122`)*.
- **Q-3 a Q-6**: sin muestra. **Medido sólo con Kimi Code 2.2.0, en Windows, con clave de API de Kimi Platform y un único modelo** *(`kimi-k2.7-code`)*.

## Assumptions
1. **El fichero es de sólo añadir**, salvo reparación o migración *(desc. Q7; m2 1, 3)*.
2. **Una línea, una vez escrita, no cambia**, y los `usage.record` heredados son idénticos byte a byte a los del origen *(m2 3)*.
3. **La carpeta de la sesión es su id, y la del agente su `agentId`** *(desc. Q1; m1 1)*.
4. **`type` va primero en un `usage.record`** *(desc. Q2; m1 1)*.

## Dependencias
- **La plataforma**:
  - `tool = "kimi_code"` como texto libre;
  - la fila **`kimi-k2.7-code`** en `backend/config/pricing.php` *(0,95 · 0,19 · 4,00 USD por millón; oficial en `platform.kimi.ai/docs/pricing/chat-k27-code`)*,
    **desplegada antes de la etiqueta v0.9.0**, en un parche aparte *(D-5)*;
  - las de la membresía *(`k3`, `kimi-for-coding`…)*, **sin decidir**: sin ellas, esos eventos llegan «ciegos y señalizados» *(Q-5)*.
- **006**: el esquema del `event_id`, sin tocarlo. **011**: el patrón de lector por offset y prefijo, sin cambiar su conducta.

## Registro de enmiendas

| # | Fecha | Qué cambia | Por qué |
|---|---|---|---|
| **Borrador** | 2026-10-15 | Primera redacción, con D-1 a D-6 del dueño, y P-1 a P-13 y Q-3 a Q-6 para ratificar | Encargo 012 · la spec |
| **Ratificación** | 2026-10-15 | **El dueño ratifica** D-1 a D-6, P-1 a P-13 **(a)** con sus `DECIDÍ YO`, y SC-001 a SC-025, con la enmienda E-1 | Encargo 012 · ratificación |
| **E-1** | 2026-10-15 | **P-4, la frontera con huella** *(decisión del dueño)*. «Antes: offset por fichero; `state.json` no gana campos; (b) una huella de la cabecera, como alternativa» → **Después**: offset **y** `cabeza` *(`event.Ref(sal, 1.ª línea)`)*; se relee desde 0 si encoge, si la cabeza cambia o si falta. **FR-004**: «Antes: los cuatro campos, sin `Frontera`» → cuatro y `cabeza`. **FR-005**: «Antes: un fichero que no creció no se abre» → mismo tamaño y mismo `mtime`. **§Límites**: «Antes: un fichero reescrito que crece no se detecta» → desaparece; queda la reescritura que conserva la 1.ª línea y no encoge. **Nace SC-026** | Encargo 012 · ratificación |
