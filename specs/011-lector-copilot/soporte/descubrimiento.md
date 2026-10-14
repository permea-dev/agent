# 011 · Descubrimiento del lector de Copilot CLI

**Fecha**: 2026-10-14 · **Fuentes**: la **copia congelada** de `session-state/` de Copilot CLI **1.0.97** *(Windows, plan Free, modelo «Auto»)*,
tomada el 2026-10-14, y la documentación pública de GitHub.

**Cómo se lee esto.** Cada afirmación lleva una etiqueta:
- **MEDIDO**: en la copia, leída con Python `-I` en modo lectura, desde scripts y temporales fuera del repo. Las cuentas son enteras y exactas;
- **LEÍDO EN LA DOCUMENTACIÓN**: en la página «Configure Copilot CLI» de docs.github.com y en la referencia del directorio de configuración. No se leyó
  el paquete instalado, porque es una carpeta del dueño;
- **NO SE PUDO MEDIR**: y qué muestra haría falta.

> **Privacidad.** De la copia salen sólo tipos de evento, rutas de claves con su tipo JSON, recuentos, tokens, AIU, nombres de modelo, la versión y
> marcas de tiempo *(UTC)*. Los identificadores se compararon entre sí sin imprimirse *(`<id>`)*. De `workspace.yaml`, `checkpoints/index.md` y
> `rewind-file-snapshots/tracking.json` sólo se leyeron las claves. No se ejecutó `copilot`. sha256 de los 17 ficheros antes y después: **iguales**.

## La copia congelada *(MEDIDO)*

- **Tres carpetas de sesión** y la carpeta `.session-operation-locks/`, con **17 ficheros**.
- **Sesión A+B**, una carpeta y un solo `events.jsonl` de **46 líneas** *(165 358 B)*: **A** empieza a las 09:32:17, 2 prompts, `shutdown` a las 09:37:56;
  **B**, `session.resume` a las 09:38:28, 1 prompt, `shutdown` a las 09:39:39.
- **Sesión C**: de 09:39:47 a 09:40:22, 1 prompt, **20 líneas** y 63 973 B. Se cerró con la X y aun así tiene `shutdown`.
- **La tercera carpeta**, de las 09:26:06, **no tiene `events.jsonl`** *(Q9)*.

## Q1 · Dónde y en qué forma

- **Ruta** *(LEÍDO EN LA DOCUMENTACIÓN)*: `~/.copilot` por defecto, y **`COPILOT_HOME` sustituye la ruta entera**. La copia es
  `<raíz>/session-state/`. La documentación no dice explícitamente que `session-state/` se mueva con la variable *(pregunta abierta 1)*.
- **Por sesión** *(MEDIDO)*, en `session-state/<id>/`, donde **el nombre de la carpeta es `data.sessionId` de `session.start`**:

  | Fichero | A+B | C | Forma |
  |---|---|---|---|
  | `events.jsonl` | sí | sí | JSONL, un objeto por línea |
  | `workspace.yaml` | sí *(13 claves)* | sí *(13)* | `id`, `cwd`, `client_name`, `name`, `user_named`, `summary_count`, `fork_count`, `created_at`, `updated_at`, `remote_steerable`, `mc_task_id`, `mc_session_id`, `mc_last_event_id` |
  | `checkpoints/index.md` | sí | sí | plantilla `# Checkpoint History` con la tabla `\| # \| Title \| File \|` y **0 filas**; idéntica en las tres carpetas |
  | `rewind-file-snapshots/tracking.json` | sí | sí | `{schema: int, tracking: bool}` |
  | `files/`, `research/` | vacías | vacías | — |
  | `.workspace-fork.lock` | 0 B | 0 B | fichero vacío |

- **Codificación**: **UTF-8 sin BOM** y **LF** *(0 CRLF)*, también en Windows. Cada fichero termina en `\n`.
- **Bloqueos**: `.session-operation-locks/<id>.lock` y `<id>/.workspace-fork.lock`, **vacíos** *(0 B)*. Hay **4** de operación para 3 carpetas: el cuarto **no tiene
  carpeta** y es de las 09:38:23.9, **4,2 s antes de `session.resume`** *(pregunta 2)*. **El agente no los crea ni los toca**: sólo abre `events.jsonl` en lectura.
- **Sólo se añade al final** *(compatible con lo medido)*: `eventsFileSizeBytes` de cada `shutdown` *(3/3)* y del `resume` es **exactamente el offset donde empieza
  su propia línea**, y B **sigue en el mismo fichero** que A.

## Q2 · Esquema por tipo *(MEDIDO)*

**La envoltura es la misma en todas las líneas**: `type` *(texto)*, `data` *(objeto)*, `id` *(texto, UUID, **46/46 y 20/20 distintos**)*, `timestamp`
*(texto ISO 8601 en UTC, con ms)* y `parentId` *(texto o `null`)*. **21 tipos.** ⛔ = lleva contenido del usuario o del entorno: **el lector nunca lo lee**.

| Tipo | A+B | C | Claves de `data` *(resumen)* | |
|---|---:|---:|---|---|
| `session.start` | 1 | 1 | `sessionId`, `version` *(int, = 1)*, `producer` *(`copilot-agent`)*, `copilotVersion` *(`1.0.97`)*, `startTime`, `context.cwd`, `alreadyInUse`, `remoteSteerable`, `contextTier` | ⛔ `cwd` |
| `session.resume` | 1 | — | `resumeTime`, `eventCount` *(int)*, `eventsFileSizeBytes` *(int)*, `selectedModel`, `reasoningEffort`, `context.cwd`, `sessionLimits`… | ⛔ `cwd` |
| `session.model_change` | 2 | 1 | `cause`, `source`, `newModel`, `previousModel`, `reasoningEffort`, `contextTier` | |
| `session.auto_mode_resolved` | 3 | 1 | `chosenModel`, `candidateModels[]`, `availableModels[]`, `categoryScores.*` *(float)*, `routingMethod`, `fallback`, `endToEndLatencyMs`, `hasImage` | |
| `session.info` | 2 | 1 | `infoType` *(`mcp`, `authentication`, `folder_trust`)*, `message` | ⛔ |
| `user.message` | 3 | 1 | `content`, `transformedContent`, `messageId`, `interactionId`, `turnId`, `delivery`, `responsesReasoning.{model,effort,initialEffort}`… | ⛔ |
| `system.message` | 1 | 1 | `role`, `content`, `metadata.variables.gitCommitTrailer.*`, `interactionId` | ⛔ |
| `assistant.turn_start` / `assistant.turn_end` | 3 / 3 | 1 / 1 | `turnId`, `interactionId` | |
| `assistant.message` | 3 | 1 | `content`, `reasoningText`, `reasoningOpaque`, `encryptedContent`, `reasoningBlocks`, `toolRequests[]`, `model`, `apiCallId`, `phase`… | ⛔ |
| **`session.usage_record`** | **3** | **1** | **`usage.*`**: Q3 | |
| `session.usage_checkpoint` | 3 | 1 | `totalNanoAiu`, `totalPremiumRequests`, `accountingSnapshot.*`, `providerModelMetrics[]`, `modelCacheState[]`, `usageAccountingWatermarks.<id>`, `promptCacheBreakState[]` | ⛔ `promptCacheBreakState` *(nombres de herramienta y hashes del prompt)* |
| `session.shutdown` | 2 | 1 | lo mismo que el checkpoint, y además `shutdownType`, `tokenDetails.*`, `modelMetrics.<modelo>.*`, `currentTokens`, `systemTokens`… | ⛔ `codeChanges.filesModified[]` |
| `model.turn_started` / `model.turn_ended` / `model.model_call_started` | 2 / 2 / 2 | 1 / 1 / 1 | `model`, `modelInfo.*` *(con `billing.token_prices.*`)*, `turn`, `timestampMs` | |
| `model.captured_assignment_context` | 2 | 1 | `assignmentContext` *(texto)* | ⛔ por precaución |
| `model.model_call_success` | 2 | 1 | `callId`, `modelCall.*`, `responseUsage.*`, `copilotUsage.*`, `quotaSnapshots.*`, **`requestMessages`**, `responseChunk.choices[].delta.content`… | ⛔ |
| `model.message` / `model.response` | 2 / 2 | 1 / 1 | `callId`, `modelCall.model`, `message.*` / `response.*` *(`content`, razonamiento)* | ⛔ |
| `model.messages_snapshot` | 2 | 1 | `messages[]` *(conversación entera)* | ⛔ |

- **C** tiene las mismas claves que A+B, salvo las de razonamiento *(`reasoning*`, `encrypted_content`, `phase`)*, que en C no aparecen.
- **La línea más larga** es la de `system.message`, de **33 408 B**, en las dos sesiones. La de `usage_record` llega a 2 857 B y la de `usage_checkpoint`, a 8 518 B.
- **Los `model.*` forman un árbol aparte**: `model.turn_started` tiene `parentId: null`. **No cuentan en `eventCount`**: el `resume` dice 21, y antes de él hay 37
  líneas, de las cuales 16 son `model.*`.

## Q3 · Llamada a llamada *(MEDIDO)*

- **`session.usage_record`**: **una línea por llamada al modelo principal**.
  - Lleva `usage.model`, `inputTokens`, `outputTokens`, `cacheReadTokens`, `cacheWriteTokens`, `reasoningTokens`, `cost` *(float)*, `isAuto` y `copilotUsage.{tokenDetails[], totalNanoAiu}`;
  - y además `accounting.{sourceSessionId, sequence, usageId}`, `apiCallId`, `providerCallId`, `serviceRequestId`, `duration`, `finishReason` y `toolCounts` *(objeto;
    vacío aquí, y llevaría nombres de herramienta)*.
  - **Tokens: sí. Modelo: sí, en la misma línea. Marca de tiempo: sí**, la `timestamp` de la envoltura. ⚠️ **`providerCallId` se repite** *(llamadas 1 y 2): no sirve*.
  - **Id estable**: **`accounting.usageId`**. Es **igual a `apiCallId`** en 4/4 y distinto del `id` de la línea. **`(sourceSessionId, sequence)`** también es único
    *(sequence 1, 2, 3 en A+B; 1 en C)*, y `sourceSessionId` = nombre de la carpeta en 4/4.
- **`model.model_call_success`, `model.message` y `model.response` NO son las llamadas del modelo principal.** Son **llamadas auxiliares**:
  - **3 en total**, `initiator: agent`: `gpt-4o-mini` *(200 de entrada y 11 de salida en A; 200 y 10 en C)* y `gpt-5.4-nano` *(537 y 46, 13 de ellas de razonamiento)*;
  - con **`cost_per_batch` = 0 y `total_nano_aiu` = 0**, y `modelInfo.billing.token_prices.*` = 0;
  - **sin `usage_record`** y **fuera del `shutdown`**. `model.response` sólo trae `callId`, `modelCall.model` y el contenido ⛔. Ningún `model.*` es de `gpt-6-luna`.
- **Σ de los `usage_record` = `session.shutdown`**, en las cuatro partidas, en `totalNanoAiu` y en `requests.cost`:

  | | input | output | cacheRead | cacheWrite | reasoning | nanoAIU | cost |
  |---|---:|---:|---:|---:|---:|---:|---:|
  | 1 · A, 09:36:21.261 | 15 367 | 5 | 0 | 15 364 | 0 | 192 330 000 | 1,0 |
  | 2 · A, 09:36:35.224 | 15 410 | 6 | 15 364 | 43 | 0 | 16 231 500 | 1,0 |
  | 3 · B, 09:39:26.507 | 15 454 | 19 | 15 106 | 345 | 11 | 20 398 500 | 1,0 |
  | **Σ 1–2 = 1.er `shutdown`** | **30 777** | **11** | **15 364** | **15 407** | **0** | **208 561 500** | **2** |
  | **Σ 1–3 = 2.º `shutdown`** | **46 231** | **30** | **30 470** | **15 752** | **11** | **228 960 000** | **3** |
  | **C, 09:40:09.036 = su `shutdown`** | **15 368** | **6** | **9 692** | **5 673** | **0** | **80 934 500** | **1** |

## Q4 · Las partidas *(MEDIDO)*

- **`inputTokens` INCLUYE la lectura y la escritura de caché**: `inputTokens = tokenDetails[input] + cacheRead + cacheWrite` en **4 de 4** llamadas *(15 367 = 3 + 0 + 15 364;
  15 410 = 3 + 15 364 + 43; 15 454 = 3 + 15 106 + 345; 15 368 = 3 + 9 692 + 5 673)*. En el total: 30 777 = **6** + 15 364 + 15 407. Lo no cacheado es 3 tokens por llamada.
- **`outputTokens` INCLUYE el razonamiento**: en la llamada 3 *(19 de salida, 11 de razonamiento)* Copilot factura **19** de `output`, y `totalNanoAiu` sólo cuadra
  con 19 *(con 30 daría 20 948 500, no 20 398 500)*. La auxiliar `gpt-5.4-nano` igual: 46 de `completion_tokens`, 13 de razonamiento, 46 facturados.
  ⚠️ Es **una sola** llamada del modelo principal con razonamiento > 0.
- **El reparto en las cuatro partidas del evento** *(como en Codex)*: `tokens_input` = `inputTokens − cacheReadTokens − cacheWriteTokens` *(= `tokenDetails[input]`)*;
  `tokens_output` = **`outputTokens`, sin sumar `reasoningTokens`** *(al contrario que Gemini CLI y OpenCode)*; `tokens_cache_creation` = `cacheWriteTokens`;
  `tokens_cache_read` = `cacheReadTokens`.

  En A+B, el reparto da **9 · 30 · 15 752 · 30 470**, igual que `shutdown.tokenDetails`. En C da **3 · 6 · 5 673 · 9 692**. La coherencia es la de Codex:
  `cacheRead + cacheWrite ≤ inputTokens`.

## Q5 · La reanudación *(MEDIDO)*

- **`--continue` añade al MISMO `events.jsonl`**: un `session.resume` *(`eventCount` 21, `eventsFileSizeBytes` 135 930, `selectedModel` `auto`)* sin un nuevo `session.start`.
- **Los `usage_record` NO se repiten**: 3 líneas, `sequence` 1, 2 y 3, 3 `usageId` distintos, nada de A reescrito *(Q1)*. `usageAccountingWatermarks.<id de sesión>`
  vale la última `sequence`: 1, 2, 2, 3, 3.
- **Los `shutdown` son ACUMULADOS de la sesión, no del tramo**: el 2.º = el 1.º + la llamada 3 *(30 777 + 15 454 = 46 231)*. Si se leyeran los `shutdown`, la regla sería
  `tramo_n = shutdown_n − shutdown_{n−1}`. **Leyendo los `usage_record` no hace falta ninguna regla.**
- La caché sobrevive a la reanudación *(15 106 de lectura en la llamada 3)*, y **también entre sesiones**: C lee 9 692 de caché en su primera llamada.

## Q6 · El cierre

- **C, cerrada con la X, tiene `shutdown`**, con `shutdownType: routine` como A y B, 13 s después de su `usage_checkpoint` *(MEDIDO)*.
- **Cuándo se escribe cada `usage_record`** *(MEDIDO)*: inmediatamente tras su llamada y **antes** de `assistant.message`, `turn_end` y el checkpoint. En las tres
  sesiones, **Σ `usage_record` = `shutdown`** exactamente.
- **Sin `shutdown`**, con los `usage_record` como fuente, **no se perdería nada de lo ya escrito**; como mucho, la llamada en vuelo al morir.
  **NO SE PUDO MEDIR**: no hay una sesión matada *(`taskkill /F`, `kill -9`, corte de luz)*. Hace falta esa muestra para confirmarlo.

## Q7 · El modelo y Auto *(MEDIDO)*

- **`session.model_change`**: al empezar, `cause: initial_resolution`, `source: automatic`, `newModel: auto`; en A, otro antes del `/exit` *(`previousModel: auto`,
  `reasoningEffort: medium`)*.
- **`session.auto_mode_resolved`**: uno por prompt *(4 en total)*, con `routingMethod: auto_v2`, `fallback: false` y `chosenModel: gpt-6-luna`. En el plan Free,
  `candidateModels` y `availableModels` = `[gpt-6-luna]`.
- **Por llamada llega `usage.model` = `gpt-6-luna`**, nunca `auto`, con `isAuto: true` y `modelDisplayName` `GPT-6 Luna`. Es la clave para la tarifa.
- Las auxiliares traen `gpt-4o-mini` y `gpt-5.4-nano`; su `responseChunk.model` lleva la fecha *(`gpt-4o-mini-2024-07-18`, `gpt-5.4-nano-2026-03-17`)*.

## Q8 · El coste *(MEDIDO)*

- **`usage.cost` y `requests.cost` son PETICIONES PREMIUM, no dinero**: 1,0 por llamada; `totalPremiumRequests` 1 → 2 → 3, igual que `totalPremiumRequestCost`.
- **`totalNanoAiu` son créditos de IA de Copilot** *(«AIU»; `aiCreditsStatus: complete`)*. **`Σ tokenCount × costPerBatch / batchSize` lo reproduce exacto en 4/4.**
  Tarifa de `gpt-6-luna` en AIU por millón: entrada **10**, lectura **1**, escritura **12,5**, salida **50** *(1 : 0,1 : 1,25 : 5)*. **NO SE PUDO MEDIR** la conversión
  de AIU a dólares: no está en la copia.
- **Recomendación** *(sin decidir; como en 008–010)*: **sin coste** *(`cost_available = false`)*; el coste lo calcula la plataforma con su tarifa, y ni las AIU ni las
  peticiones premium viajan. Hace falta una fila `gpt-6-luna` en la tarifa de Permea.

## Q9 · La sesión sin `events.jsonl` *(MEDIDO)*

- **Tiene** `workspace.yaml` *(**8 claves**: `id`, `cwd`, `client_name`, `user_named`, `summary_count`, `fork_count`, `created_at`, `updated_at`)*, la plantilla de
  `checkpoints/index.md`, `files/` y `research/` vacías, `.workspace-fork.lock` y su `.lock` de operación, **todo escrito en 4 ms** *(09:26:06.112–.116)* y nada después.
  **No tiene** `events.jsonl`, `rewind-file-snapshots/` ni las claves `name`, `remote_steerable` y `mc_*`.
- **Lectura**: la carpeta se crea antes de autenticar y el fallo llegó **antes del primer evento**. **El 403 no consta en la copia** *(no incluye `logs/`)*; A tiene
  un `session.info` con `infoType: authentication` a las 09:33:19. **Para el lector**: una carpeta sin `events.jsonl` = 0 eventos, sin aviso.

## Q10 · Encaje en el agente *(sin código)*

- **Se reutiliza del lector de Codex**: **«una línea = un evento»** *(`token_usage_record` ↔ `session.usage_record`)*, sin pasada ni espera de cierre; la lectura por
  offset con `ReadBytes('\n')` *(admite las líneas de 33 KB)*; la línea corrupta clasificada, no fatal; la coherencia `caché + escritura ≤ entrada` y la resta de
  `tokens_input`; `cost_available = false`; el **prefijo releído con marcadores** para el contexto *(aquí `session.start` y `session.resume`, por `context.cwd`)*; el
  esquema de `event_id` con espacio de nombres propio *(`copilot`)*; `event.Ref` con sal; y una raíz como `GeminiRoot` *(`COPILOT_HOME` o `~/.copilot`: la **quinta**
  lectura de entorno)*.
- **Diferencias**: **el modelo viene en la misma línea** *(más simple que `DelTurno`)*; **`cwd` por tramo** *(el del último `start`/`resume` anterior)*; **la salida
  no suma el razonamiento** *(Gemini y OpenCode sí)*; **hay escritura de caché** *(Gemini, 0)*; **sin apariciones repetidas** *(al revés que Gemini)*; carpetas sin
  `events.jsonl` ignoradas; `.lock` intocables; de cada línea sólo la envoltura y los `usage.*` necesarios, **nunca** `toolCounts` ni los tipos ⛔.

## Q11 · Lo que esta muestra no puede responder

| Hueco | Cómo se cerraría |
|---|---|
| **Linux y macOS** *(ruta, LF, `COPILOT_HOME`)* | una sesión corta en cada uno, con y sin `COPILOT_HOME`, con sha256 y `find` de la raíz |
| **Sesiones largas o con compactación** *(¿`usage_record` de la compactación? ¿se reescribe `events.jsonl`?)* | una sesión con `/compact` o que llegue al límite de contexto; contar las líneas y comprobar que los offsets sólo crecen |
| **Varios modelos en una sesión** *(`/model` a mitad)* | sesión con dos modelos de pago; ver `usage.model` por línea y `modelMetrics` con dos claves |
| **Llamadas con herramientas y subagentes** *(`toolCounts`, `agentMetrics` distinto de `main`)* | una sesión que edite ficheros y lance un subagente |
| **Una sesión matada** *(Q6)* | `taskkill /F` a mitad de respuesta; comparar Σ `usage_record` con lo que diga la facturación |
| **Copilot dentro de VS Code y JetBrains** *(¿escriben en `session-state/`?)* | una conversación en cada IDE y buscar ficheros nuevos bajo `~/.copilot` y en la carpeta de datos del IDE |
| **Business y Enterprise** *(¿cambian `cost`, `aiCreditsStatus` o `isByok`?, BYOK)* | una sesión con una licencia de organización, y otra con un proveedor propio |
| **Las llamadas auxiliares en otros planes** *(¿siguen a 0 AIU?)* | la misma muestra en un plan de pago |

## Propuestas de decisión *(sin decidir)*

- **P-1 · Fuente**: `session.usage_record`, **una línea = un evento**. Se ignoran los `model.*`, los `shutdown` y los checkpoints.
- **P-2 · Partidas**: las de Q4. La salida = `outputTokens` *(el razonamiento va dentro)*.
- **P-3 · `event_id`**: de `accounting.usageId`, o de `(sourceSessionId, sequence)` si falta, con espacio de nombres `copilot`. **Nunca** de `providerCallId`.
- **P-4 · Modelo**: `usage.model` tal cual *(`gpt-6-luna`)*. `modelDisplayName` no viaja.
- **P-5 · Sesión y proyecto**: `session_ref` desde `accounting.sourceSessionId` *(= carpeta = `session.start.sessionId`)*; `project_ref` desde el `context.cwd` del
  último `start` o `resume` anterior. No se lee `workspace.yaml`.
- **P-6 · Coste**: sin coste *(Q8)*.
- **P-7 · Llamadas auxiliares** *(0 AIU)*: **(a)** no se cuentan, como la factura de Copilot; **(b)** se cuentan desde `model.model_call_success.responseUsage`, en
  una línea ⛔ con `requestMessages`.
- **P-8 · Raíz**: `$COPILOT_HOME/session-state`, o `~/.copilot/session-state`.

## Preguntas abiertas

1. ¿`session-state/` sigue a `COPILOT_HOME`, y existe alguna otra variable *(`XDG_*`)*? *(Q1, Q11)*
2. ¿Qué es el cuarto `.lock` de operación, sin carpeta, creado 4,2 s antes del `resume`? ¿Un id provisional de `--continue`? *(Q1)*
3. ¿Una sesión matada conserva todos sus `usage_record`? *(Q6)*
4. ¿La compactación deja `usage_record` propio, y `events.jsonl` sigue siendo sólo-añadir? *(Q11)*
5. ¿A cuántos dólares equivale una AIU, y lo necesita Permea? *(Q8)*
6. ¿Se cuentan las llamadas auxiliares a 0 AIU? *(P-7)*
