# 012 · Descubrimiento del lector de Kimi Code CLI *(sin muestra)*

**Fecha**: 2026-10-15 · **Fuentes**: el código público de Kimi Code *(`MoonshotAI/kimi-code`, MIT)* en la etiqueta **`@moonshot-ai/kimi-code@2.2.0`**
*(commit `472192c`, del 2026-10-14; la etiqueta anotada es de las 15:31 UTC)*: **es la última publicada y la que tiene instalada el dueño**. También la documentación
oficial de `kimi.com/code/docs`, y ccstats 0.4.0 sólo como contraste.

**Cómo se lee esto.** Cada afirmación lleva una etiqueta:
- **[C `fichero:línea`]** = **LEÍDO EN EL CÓDIGO** de la 2.2.0. `ac/` abrevia `packages/agent-core-v2/src/`; `man`, el manifiesto generado del protocolo `packages/agent-core-v2/docs/wire-manifest.d.ts`;
- **[D página]** = **LEÍDO EN LA DOCUMENTACIÓN**, en `kimi.com/code/docs/kimi-code-cli/<página>`, cuya fuente es `docs/en/` de la etiqueta. `configuration/data-locations` y
  `kimi-code/models` se cotejaron en la web el 2026-10-15;
- **[T]** = **DE TERCEROS** *(la página de ccstats 0.4.0 en docs.rs)*; **[M]** = **POR MEDIR**, con la sesión que lo resolvería *(Q11)*.

> **Privacidad.** No se instaló ni se ejecutó `kimi` ni se leyó nada del dueño. El clon, en un `mktemp -d`, se borró al acabar. Nada suyo entra en el repo:
> sólo nombres de tipos y campos.

## Q1 · Dónde y en qué forma

- **Raíz**: `KIMI_CODE_HOME`, o `~/.kimi-code` *(en Windows, `%USERPROFILE%\.kimi-code`)* [C `ac/app/bootstrap/bootstrap.ts:167-174`; D `configuration/data-locations`].
  ⚠️ El núcleo usa `??`, así que un `KIMI_CODE_HOME=""` **no** cae al valor por defecto; la CLI sí trata el vacío como ausente [C `apps/kimi-code/src/utils/paths.ts:38-44`].
- **Árbol**: `<raíz>/sessions/<workDirKey>/<sessionId>/agents/<agentId>/wire.jsonl` [C `ac/app/bootstrap/bootstrapService.ts:45`; `ac/workspace/sessionLifecycle/internal/addressing.ts:7-17`; D; T].
  - `workDirKey` = `wd_<slug>_<12 hex>`: el slug sale del **último componente** del directorio de trabajo y el hash es SHA-256 de la ruta con `/`, **sin pasar a minúsculas**
    [C `ac/_base/utils/workdir-slug.ts:17-23`]. ⛔ **El slug revela el nombre de la carpeta del proyecto**: no viaja.
  - `sessionId` = `session_` + un UUID aleatorio [C `ac/workspace/sessionLifecycle/sessionLifecycleService.ts:903-905`]; `agentId` = `main` o `agent-<n>`
    [C `ac/session/agentLifecycle/agentLifecycleService.ts:180-189`; D].
- **`session_index.jsonl`**, en la raíz: una línea `{sessionId, sessionDir, workDir}` por sesión creada o bifurcada [C `ac/app/workspace/workspaceAlias.ts:9-10,20-24`;
  `sessionLifecycleService.ts:325-333`]. ⛔ Son rutas absolutas. **Se compacta reescribiéndolo** [C `workspaceAlias.ts:26-30`].
- **No se leen**, aunque estén en la carpeta de la sesión [D]: `state.json` *(título, `lastPrompt` ⛔ y `forkedFrom`)*, `tasks/` *(salidas ⛔)*, `cron/`, `logs/`, `plans/` y `upcoming-goals.json`.
- **Con secretos o contenido; el agente no los abre nunca**: `config.toml` *(`api_key` y `[providers.<n>.env]`)* [D `configuration/providers`]; `credentials/` *(OAuth, en ficheros
  0600, y `credentials/mcp/`)*, `mcp.json`, `plugins/`, `user-history/` *(prompts)*, `logs/`, `blobs/`, `store/` y `server/` [D; C `bootstrap.ts:67-74`].
- **`region`** *(el que dejó el instalador)*: un **marcador de una línea, `mainland-cn` o `global`**, que elige la región *(`.com` o `.ai`)* **sólo mientras no haya habido un login**.
  Después manda el `oauthHost` de `config.toml` [C `packages/oauth/src/region.ts:116-123,160-169`; orden de resolución en `:11-21`]. El script que lo escribe no está en el repositorio
  *(se sirve desde la CDN)*. **No tiene uso ni secretos.** `bin\` guarda binarios gestionados [D].
- **Codificación y escritura**: `JSON.stringify(registro) + "\n"` en UTF-8, sin BOM y **con LF también en Windows** [C `ac/persistence/backends/node-fs/appendLogStore.ts:311-315`;
  su propio lector tolera `\r\n`, `:91,104`]. Se escribe por **lotes de líneas completas**, con `open(…, 'a')`, `writeFile` y `fsync` [C `appendLogStore.ts:285-296`;
  `fileStorageService.ts:120-131`]. ⚠️ **No siempre es de sólo añadir** *(Q7)*.

## Q2 · El esquema de `wire.jsonl`

- **Versión del protocolo *wire***: **`1.5`** [C `ac/wire/migration/migration.ts:19`]. La primera línea es `{"type":"metadata","protocol_version","created_at"}` [C `ac/wire/record.ts:38-44`].
  **Cada registro** es `{"type", …la carga extendida…, "time"}`: **`type` primero y `time` al final**, en ms desde la época, como número [C `ac/app/event/event2.ts:32-45`; man `:15-18`].
- **Hay 64 tipos duraderos** [man `:27-92`], más los registros «humanos» *(`agent.switched`, `agent.message.appended`, `agent.turn.*`)*, en los que **`type` no va primero**
  [C `ac/wire/human.ts:5-16`; `ac/agent/loop/machine/storeJournal.ts:46`].
- ⛔ **Llevan contenido**: `context.append_message`, `turn.prompt`, `turn.steer`, `config.update.systemPrompt`, `llm.request.systemPrompt`, `llm.tools_snapshot`,
  `subagent.completed.resultSummary`, `task.terminated.outputTail` y `plan.*`, entre otros [man].
- **El registro de uso es `usage.record`**: `{agentId, model, usage, usageScope?}` [C `ac/agent/usage/usageOps.ts:13-24`; man `:952-958`; T]. `usage` es un `TokenUsage`
  [C `ac/human/llm/usage.ts:1-7`], que cada protocolo rellena en `ac/human/llm/requester/bases/` [C `openai/format.ts:125-138`; `openai-responses/format.ts:361-376`;
  `google-genai/format.ts:265-283`; `anthropic/format.ts:95-113`]:

  | Campo | Tipo | Qué es |
  |---|---|---|
  | `inputOther` | int | entrada **sin** caché: prompt − lectura de caché *(OpenAI, Responses, Gemini)*, o el `input_tokens` de Anthropic, que ya la excluye |
  | `inputCacheRead` | int | lectura de caché: `cached_tokens` *(OpenAI y Kimi)*, `cache_read_input_tokens` *(Anthropic)*, `cachedContentTokenCount` *(Gemini)* |
  | `inputCacheCreation` | int | escritura de caché: **sólo en el protocolo Anthropic** *(`cache_creation_input_tokens`)*; en el resto es **0 fijo** |
  | `output` | int | salida: `completion_tokens`, `output_tokens` o `candidatesTokenCount` |
  | `raw` | objeto, opcional | **el objeto de uso del proveedor, tal cual** |

- **Razonamiento**: **no tiene partida propia**. En OpenAI y Anthropic va dentro de la salida; en Gemini, `candidatesTokenCount` lo **excluye**. En Kimi, **[M]**:
  ¿`completion_tokens` lo incluye y `raw` lo desglosa?
- **`usageScope`**: `turn` para el bucle y `session` para las operaciones, como la compactación [C `ac/session/usage/usageAgentModel.ts:36`].
- **El modelo**: `usage.record.model` es el **alias** de configuración *(`request.modelAlias`)*, no el id del proveedor [C `ac/agent/llmRequester/llmRequesterService.ts:500-505`].
  En la membresía es `kimi-code/<id>` [C `packages/oauth/src/managed-kimi-code.ts:224-226`; T: «`kimi-code/k3`»].
  - El id real va en la **`llm.request`** anterior *(`model` = nombre, `provider` = **el protocolo**, `modelAlias`)*, que **puede llevar `systemPrompt`** ⛔
    [C `llmRequesterService.ts:819-848`; `ac/agent/llmRequester/llmRequestOps.ts:41-62`]. **`usage.record` no lleva el proveedor.**
- **Id único: ninguno.** `usage.record` no tiene id de llamada ni del servidor. El `providerMessageId` va en `context.append_message` ⛔, junto con un `usage`
  [man `:142,147`]; `turn.ended.traceId` es uno por turno [man `:830-883`].
- **Marca de tiempo**: `time`, el `Date.now()` del momento en que se crea el evento, **después** del `finish` [C `event2.ts:35`; `llmRequesterService.ts:466-505`].
- **Coste y cuota: no hay.** La cuota de la membresía sólo está en el servidor *(`<base>/usages`, lo que muestra `/usage`)* [C `packages/oauth/src/managed-usage.ts:21-22`; D `reference/slash-commands`].

## Q3 · Llamada a llamada

- **Un `usage.record` por petición terminada.** Lo escribe **un único sitio** del núcleo, `llmRequesterService.ts:500`, y sólo cuando el *stream* llega a `finish`
  [C `:441-471`]. Del *stream* se queda **el último** evento `usage` [C `:450-452`].
- **Reintentos**: cada intento escribe su `llm.request` *(con `attempt`)* [C `:426-427`], y el bucle reintenta [C `:523-560`; man `:933`, `turn.step.retrying`].
  **Un intento fallido o una respuesta interrumpida con Esc no dejan `usage.record`**, aunque el proveedor facture parte *([M] cuánto)*. Si el proveedor no da uso, se escribe
  `emptyUsage()`, **un registro con todo a 0** [C `:503`].
- **Llamadas auxiliares sin `usage.record`**: los **títulos**, que van a un endpoint de herramientas por OAuth y sólo con la membresía
  [C `ac/session/sessionTitle/sessionTitleService.ts:96-128`], y el **`ping`** de comprobación de modelo [C `ac/llm-adapter/model/catalog-service.ts:163-178`]. [M] si la consola los cuenta.

## Q4 · Subagentes

- **Cada subagente tiene su `agents/agent-<n>/wire.jsonl` con sus `usage.record`**, emitidos en el espacio del agente que llama [C `usageAgentModel.ts:34-44`; D; T].
- ⚠️ **Doble cuenta**: el `wire.jsonl` del **principal** recibe `subagent.completed` con un `usage` que es el **total acumulado del subagente** *(`status().total`)*
  [C `ac/session/subagent/mirrorAgentRun.ts:235-240`; `runAgentTurn.ts:82`; man `:598-601`]. **Se ignora.**
- **El modelo**: el **declarado** es `subagent.spawned.model`, que es opcional [man `:620-630`]; el **facturado**, el `model` de los `usage.record` del subagente.
- **`/btw` y los subagentes bifurcados** copian **sólo los mensajes**, no los `usage.record` [C `ac/session/agentLifecycle/agentLifecycleService.ts:409-454`].

## Q5 · Compactación *(automática y `/compact`)*

- **Consume y lo registra**: usa el mismo `llmRequester`, con `source.type = 'operation'` y `requestKind: 'full_compaction'` [C `ac/agent/fullCompaction/fullCompactionService.ts:665-675`].
  Deja un **`usage.record` con `usageScope: 'session'`** y un `llm.request` con `kind: 'compaction'`, entre `full_compaction.begin` *(`source: manual | auto`)* y `.complete`,
  que **no llevan uso** [man `:254-290`]. **No hace falta una fuente aparte**, al contrario que en Copilot.

## Q6 · Reanudación, `/fork`, `/new` y `kimi export` *(crítica para deduplicar)*

- **Reanudar** *(`--continue`, `--session`)* añade al **mismo** `wire.jsonl` [C `ac/wire/wireService.ts:86-93,574-583`], salvo la reescritura de Q7. **`/new`** *(alias
  `/clear`)* abre una sesión **nueva y vacía** [D `reference/slash-commands`]. **`/undo`** añade un trío `agent.switched…` que abre una rama [C `wireService.ts:181-228`]: **las líneas abandonadas se quedan**, y se facturaron.
- **`/fork` COPIA LOS `usage.record`.**
  - `/fork` y `kimi fork` llaman a `forkSession({id})`, sin corte [C `apps/kimi-code/src/tui/commands/session.ts:51-60`]. Eso **copia byte a byte** la carpeta entera, con
    **todos los `wire.jsonl`**, también los de los subagentes; sólo deja fuera `state.json`, `logs/` y `upcoming-goals.json` [C `sessionLifecycleService.ts:579-583,830-883`].
  - Después **añade** a cada fichero `{"type":"forked","agentId","time"}` [C `:614,738-769,907-909`]. Con un corte por turno *(SDK y web)*, reescribe el prefijo
    re-serializado y añade también el `forked` [C `:569-612,713-736`].
  - El `state.json` nuevo lleva `forkedFrom` [C `:650`] y el índice gana una línea [C `:675`]. **ccstats no deduplica las bifurcaciones** [T].
- **Cómo se distingue una copia**: **en un `wire.jsonl`, todo lo anterior a la última línea `forked` es heredado**, porque bifurcar exige que no haya turno en curso
  [C `:509-546`]. La copia conserva `time`, `model` y `usage` idénticos. Si se bifurca una bifurcación, hay dos `forked`: cuenta la última.
- **`kimi export`** escribe un ZIP en el directorio actual, no bajo la raíz [C `ac/app/sessionExport/sessionExportService.ts:229-232`]. ⚠️ Pero **`kimi vis`** lo importa
  a `<raíz>/imported/<id>/agents/<id>/wire.jsonl`, **con la forma de una sesión** [C `apps/vis/server/src/lib/import-store.ts:3-9`]: **sólo se recorre `sessions/`**.
- **Migrar el Kimi CLI antiguo** *(`~/.kimi`)* no escribe `usage.record` [C `packages/migration-legacy/src/sessions/`].

## Q7 · Escritura concurrente y cierre

- **Añadir es lo normal**, serializado dentro de cada proceso [C `wireService.ts:95-125`; `appendLogStore.ts:193-211`]. **[M]**: dos procesos sobre la misma sesión *(TUI y web)*.
- **Se reescribe entero**, con `atomicWrite` *(un `.tmp.<pid>.<hex>` y luego `rename`)* [C `appendLogStore.ts:114-143`; `ac/_base/utils/fs.ts:74-104`]:
  - **al migrar** un fichero anterior a la 1.5. Si no tenía línea `metadata`, **la añade en cabeza y el fichero crece** [C `wireService.ts:337-355,389-402`];
  - **al reparar** una línea corrupta o truncada: guarda el original como **`wire.jsonl.bak`** *(si no existe ya)* y **deja el prefijo válido, perdiendo lo que hubiera detrás**
    [C `ac/wire/repair.ts:17,28-42`; `appendLogStore.ts:59-95`]; y cuando el fichero estaba vacío.
- ⚠️ **Para el lector**: el tamaño no basta, porque un desplazamiento puede caer a mitad de línea en un fichero **más largo**; hay que mirar la cabeza. `.bak` y `.tmp.*` **no se leen**:
  el `.bak` repite lo ya contado.
- **Sesión matada a mitad de respuesta**: la llamada en vuelo **no deja `usage.record`**, que se escribe tras `finish` y de forma asíncrona *(`void`)* [C `llmRequesterService.ts:500`].
  Si muere a mitad de un lote, queda una **última línea sin `\n`**, que la siguiente reanudación repara. **[M]** cuánto se pierde de verdad.

## Q8 · Proveedores

- **Tipos**: `kimi`, `anthropic`, `openai`, `openai_responses` y `google-genai` [D `configuration/providers`; C `ac/llm-adapter/provider/provider-definition.ts:201-256`].
  `type = "kimi"` habla `openai` *(por defecto)*, `anthropic` u `openai_responses`, con `https://api.moonshot.ai/v1` por defecto [C `ac/human/llm-kimi/trait.ts:17-19`;
  `provider.ts:10-33`], y saca el uso del nivel superior **o de `choices[0].usage`** [C `trait.ts:151-162`].
- **Membresía (OAuth)**: el proveedor `managed:kimi-code`, contra `https://api.kimi.com/coding/v1` *(`.ai` en `global`)*. Sus modelos llegan de `/models` en tiempo de ejecución,
  cada uno con su protocolo (`kimi`, `anthropic` o `response`) [C `packages/oauth/src/managed-kimi-code.ts:11-22`; `region.ts:61-77`]. **Las partidas dependen de ese protocolo**:
  con `anthropic` puede haber escritura de caché.
- **Clave de Moonshot y otros**: el alias es la clave de `[models.<alias>]` en `config.toml`, la que elija el dueño.
- **Ids de modelo de Kimi**: en el código, sólo **`kimi-for-coding`**, por defecto en el SDK [C `packages/node-sdk/src/kimi-code-model-provider.ts:43`]. En la documentación:
  `k3`, `k3-256k`, `kimi-for-coding` *(«K2.8 Preview»)* y `kimi-for-coding-highspeed` [D `kimi-code/models`].

## Q9 · Windows

- **Raíz**: `%USERPROFILE%\.kimi-code` [D], y **`KIMI_CODE_HOME` vale igual**, porque es la misma función en todas las plataformas [C `bootstrap.ts:167-174`].
  Las rutas se forman con `path.join`, y **los saltos son LF** [C `appendLogStore.ts:313`].
- **El hash de `workDirKey` distingue mayúsculas** *(`C:` frente a `c:`)*: un mismo proyecto puede repartirse en **dos** carpetas [C `workdir-slug.ts:18-21`]. **`KIMI_SHELL_PATH`**
  sólo apunta al Git Bash de la herramienta de *shell* y **no afecta a los datos** [C `ac/_base/execEnv/environmentProbe.ts:122-127`; D `configuration/env-vars`].
- ⚠️ **Al reescribir en Windows**, borra y después renombra [C `fs.ts:95-104`], así que hay un instante **sin fichero**. **[M]**: si un lector con el fichero abierto hace fallar ese borrado.
- **Desactivar la actualización**: `KIMI_CODE_NO_AUTO_UPDATE=1` [C `apps/kimi-code/src/cli/update/preflight.ts:494-503`; D `env-vars`], o `[upgrade].auto_install = false` en `tui.toml` [D].

## Q10 · Encaje en el agente *(sin código)*

- **`tool`**: `"kimi_code"`, que lo distingue de un futuro Kimi leído por API. **`event_id`** = `["permea/event_id/v1", "kimi_code", "respuesta", I]`. Como **no hay id del servidor**, `I` tendría que salir de una de estas:
  - **(a) la posición**: `sessionId` y `agentId` *(de las carpetas)* y el ordinal del `usage.record` en su fichero. Resiste la reanudación y la reparación *(que conserva el prefijo)*,
    pero una copia de `/fork` da **otro** id: necesita **la frontera `forked`**;
  - **(b) el contenido**: `agentId`, `time`, `model` y las 4 partidas. Una copia da **el mismo** id y la plataforma la descarta, pero **rompe la regla de 011**
    *(ni horas ni partidas en el hash)* y dos registros a 0 en el mismo milisegundo colisionarían;
  - **(c) el `providerMessageId`** del `context.append_message` del mismo paso: es del servidor, pero vive en una línea ⛔.
- **Frontera**: un desplazamiento por `wire.jsonl`, como en Copilot, **más una huella de la cabeza** *(la línea `metadata`)*. Si el fichero encoge o cambia de cabeza, se relee desde 0
  con los mismos `event_id`.
- **Filtrado sin parsear lo ⛔**: se reconocen `{"type":"usage.record"`, `{"type":"forked"` y `{"type":"metadata"` **por el prefijo de la línea**, porque en ellos `type` va primero.
  El resto **no se decodifica**.
- **Se reutiliza de Copilot**: el listado de `ListarCopilot` *(`internal/ingest/copilot_contexto.go:29`)*, aquí con el patrón fijo `sessions/*/*/agents/*/wire.jsonl`; la lectura
  por desplazamiento con `ReadBytes('\n')` *(las líneas ⛔ pueden ser largas)*; la línea corrupta clasificada; `PasadaCopilot` y su resumen; la codificación de `eventid.go`;
  `cost_available = false`; y una raíz como `config.CopilotSessionsRoot` *(la **sexta** lectura de entorno, `KIMI_CODE_HOME`)*.
- **Diferencias**: las partidas pasan **tal cual, sin restas** *(`tokens_input` = `inputOther`)*; no hay id del servidor y hay frontera `forked`; **los subagentes están en
  otro fichero** y `subagent.completed` se ignora; el modelo es un alias; el proyecto sale del `workDir` de `session_index.jsonl`, la única línea limpia con la ruta.

## Q11 · Lo que sólo la muestra puede responder, y el plan para el dueño

| Hueco [M] | Sesión que lo cierra |
|---|---|
| que `usage.record` existe con esta forma en la 2.2.0, y qué lleva Kimi en `raw` *(caché, razonamiento)* | S1 |
| que la reanudación sólo añade, sin reescribir ni repetir | S2 |
| la copia de `/fork`: Σ heredado = Σ del origen, y la línea `forked` | S3 |
| la compactación: un `usage.record` `session`, y si hay escritura de caché | S4 |
| el subagente: su `wire.jsonl` y la cifra de `subagent.completed` | S5 |
| matar la sesión: qué se pierde y si sale un `.bak` | S6 |
| el razonamiento en `output`, los intentos fallidos y el `ping`, frente a la consola | S1 + consola |

**Plan** *(Windows, PowerShell; sin tocar `%USERPROFILE%\.kimi-code` ni su `region`)*:
1. `$env:KIMI_CODE_HOME='C:\permea-prueba-kimi\home'` y `$env:KIMI_CODE_NO_AUTO_UPDATE='1'`, en **la misma consola** para todas las sesiones, y
   `cd C:\permea-prueba-kimi\proyecto`, una carpeta sin datos. Se entra con `/login` *(en la región de la cuenta)* **o** con una clave de Moonshot en el `config.toml` de prueba.
2. **S1**: normal, 2 prompts y `/exit` · **S2**: `kimi --continue` y 1 prompt · **S3**: `/fork` en S2, luego `kimi --session <fork>` y 1 prompt · **S4**: `/compact` tras 3 prompts
   · **S5**: un prompt que pida un subagente · **S6**: un prompt largo y `taskkill /F /IM kimi.exe` a mitad de la respuesta. Hay que anotar **la hora y el modelo** de cada una.
3. **Qué se copia**: **sólo `home\sessions\` y `home\session_index.jsonl`**, con su sha256. **Nunca** `config.toml`, `credentials\`, `mcp.json`, `logs\`, `user-history\` ni `region`.
4. **Cuadre**: con la clave de Moonshot, el uso por día y modelo de la consola de facturación frente a Σ `usage.record` *(entrada con y sin caché, y salida)*, con la hora
   de cada sesión. Con la membresía no hay tokens en la consola: sólo `/usage` antes y después *(cuota, cualitativo)*.

## Propuestas de decisión *(sin decidir)*

- **P-1 · Fuente**: sólo `usage.record`, de **todos** los `wire.jsonl` de `sessions/`, el del principal y los de los subagentes, con las ramas de `/undo` incluidas.
  Se ignoran `subagent.completed`, `llm.request`, `.bak`, `.tmp.*` e `imported/`.
- **P-2 · Partidas**: `tokens_input` = `inputOther`, `tokens_cache_read` = `inputCacheRead`, `tokens_cache_creation` = `inputCacheCreation`, `tokens_output` = `output`.
  El razonamiento no se suma; `raw` no viaja.
- **P-3 · Copias de `/fork`**: **(a)** la frontera de la última `forked` *(lo anterior no se emite)*, **(b)** un `event_id` por contenido, o las dos.
  Recomendación: **(a) con el `event_id` por posición**, que respeta la regla de 011.
- **P-4 · Registros a 0** *(`emptyUsage`)*: no se emiten; son una clase «sin uso».
- **P-5 · Modelo**: el alias tal cual *(`kimi-code/k3`)*, con filas de alias en la tarifa de Permea, **o** el `model` de la `llm.request` anterior, leído con una lista blanca de campos.
- **P-6 · Coste**: sin coste *(`cost_available = false`)*, como en 008–011. **P-7 · Raíz**: `$KIMI_CODE_HOME/sessions`, con el vacío como ausente *(como la CLI)*,
  o `~/.kimi-code/sessions`.
- **P-8 · Proyecto**: el `workDir` de `session_index.jsonl`, buscado por `sessionId`. Nunca el slug de `workDirKey`.

## Preguntas abiertas

1. ¿Lleva el `raw` de Kimi el razonamiento o la escritura de caché? ¿La salida de Kimi incluye el razonamiento? *(Q2, S1)*
2. ¿Factura Moonshot los intentos fallidos, las respuestas interrumpidas y el `ping`, que no dejan `usage.record`? ¿Consumen cuota los títulos de la membresía? *(Q3, Q7)*
3. ¿`event_id` por posición con la frontera `forked`, o por contenido? *(P-3)*
4. ¿Qué pasa con dos procesos sobre una misma sesión, y con un lector que tenga el fichero abierto durante una reescritura en Windows? *(Q7, Q9)*
