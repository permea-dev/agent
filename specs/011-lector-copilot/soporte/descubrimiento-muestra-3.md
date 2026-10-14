# 011 · Descubrimiento, muestra 3: un subagente

**Fecha**: 2026-10-14 · **Fuente**: la **copia congelada** `congelada-2026-10-14c/session-state/` *(1.0.97, Windows, Free, «Auto»)*: las sesiones de la muestra 2,
**idénticas byte a byte**, y **G**: 41 líneas *(10:51:57–10:53:08)*, 1 prompt, un subagente `task` y `/exit`. Etiquetas y privacidad de [descubrimiento.md](descubrimiento.md); **sha256** de los 37 ficheros, antes y después: **iguales**.

## S1 · Esquemas *(MEDIDO)*

La envoltura es la de siempre; las líneas **del subagente** *(`subagent.*`, sus `tool.execution_*` y **sus dos `usage_record`**)* añaden **`agentId`** *(texto)*.
| Tipo | Claves de `data` *(tipo)* |
|---|---|
| `subagent.started` | `toolCallId`, `agentName` *(`task`)*, `agentDisplayName`, `agentDescription` ⛔, `model` *(`gpt-5.6-luna`)*, `taskModelSource`, `modelSelectionSource`, `agentType`, `executionMode` *(`sync`)* *(texto)*; `resumable` *(bool)* |
| `subagent.configured` | `model` *(`gpt-5.6-luna`)*; `multiTurn` *(bool)* |
| `subagent.completed` | `toolCallId`, `agentName`, `agentDisplayName`, `model` *(`gpt-5.6-luna`)*, `firstDispatchedModel` *(**`gpt-6-luna`**)*, `modelSelectionSource` *(texto)*; `totalToolCalls`, **`totalTokens`** *(22 284)*, `durationMs` *(int)* |
| `tool.execution_start` / `_complete` | `toolCallId`, `toolName`, `turnId`, `model`, `arguments.*` ⛔ / `toolCallId`, `model`, `interactionId`, `turnId`, `success`, `rte`, `result.content` ⛔, `toolTelemetry.*`; en el comando del subagente, además, `toolTitle` ⛔, `shellToolInfo.*`, `shellExecution.exitCode` y `parentToolCallId` |

- **Créditos: ninguno. Tokens: dos cifras, y ninguna es una llamada nueva**: `subagent.completed.totalTokens` = 22 284 = entrada + salida de los 2 registros
  del subagente *(22 234 + 50)*, y `toolTelemetry.metrics.result_token_count` mide el resultado de la herramienta. Al configurar el subagente se escribe un
  `session.model_change` *(`source: agent`, `newModel: gpt-5.6-luna`)*.

## S2 · Qué registros son del subagente *(MEDIDO)*

| # | `initiator` | `interactionType` | `sourceSessionId` | `sequence` | `agentId` en la línea | `parentToolCallId` | `model` |
|---|---|---|---|---:|---|---|---|
| 1 | `user` | `conversation-agent` | = carpeta | 1 | no | no | `gpt-6-luna` |
| 2 | `sub-agent` | `conversation-subagent` | **otro** *(no es carpeta)* | **1** | sí | = `toolCallId` de la tarea | `gpt-6-luna` |
| 3 | `sub-agent` | `conversation-subagent` | el mismo otro | **2** | sí | ídem | `gpt-6-luna` |
| 4 | `agent` | `conversation-agent` | = carpeta | 2 | no | no | `gpt-6-luna` |

**Del subagente son el 2 y el 3**: su `sequence` vuelve a empezar, con un `sourceSessionId` propio sin carpeta. `agentId` = la clave de `agentMetrics` del
  `shutdown` = `restrictedProperties.agent_id`, y **no** = `sourceSessionId`. Los 4 `usageId` son distintos *(cada uno = su `apiCallId`)*. Los registros del
  subagente cuestan **0** peticiones premium *(`cost` 0,0)*; el `shutdown` cuenta 1. **El subagente declara `gpt-5.6-luna`** *(`started`, `configured`, `completed.model` y la pantalla)*, pero **sus registros dicen `gpt-6-luna`**, y
  `completed.firstDispatchedModel` también. El marcador del prefiltro *(`"type":"session.usage_record"`)* está en las 4 líneas.

## S3 · ¿Qué modelo cobró Copilot? *(MEDIDO; tarifas de `config/pricing.php`, sólo esas dos filas)*

| # | créditos × 0,01 | tokens × `gpt-6-luna` *(0,10 / 0,50 / 0,125 / 0,01)* | tokens × `gpt-5.6-luna` *(0,20 / 1,20 / 0,25 / 0,02)* |
|---|---:|---:|---:|
| 1 | 0,002066675 | **0,002066675** ✓ | 0,00419095 |
| 2 *(sub)* | 0,00140755 | **0,00140755** ✓ | 0,0028241 |
| 3 *(sub)* | 0,00012258 | **0,00012258** ✓ | 0,00024616 |
| 4 | 0,000193965 | **0,000193965** ✓ | 0,00038893 |
| **Σ** | **0,00379077** | **0,00379077** | 0,00765014 |

**Cuadra `gpt-6-luna`, exacto, en 4 de 4**; `gpt-5.6-luna`, en ninguno. `copilotUsage.tokenDetails[].model` y sus precios por lote dicen también `gpt-6-luna`.
  ⇒ **El modelo del registro es el que Copilot cobró**; el que declara el subagente no se usó. Σ = 0,379 créditos: los «0,38» de la pantalla.

## S4 · Σ `usage_record` = `shutdown` *(MEDIDO)*

- **Tokens**: entrada **53 297** *(31 063 del orquestador + 22 234 del subagente)*, salida **343**, lectura **26 457**, escritura **26 828** = `shutdown.modelMetrics`
  y `tokenDetails` *(entrada sin caché: 12)*. **Créditos**: **379 077 000 nanoAIU** = `shutdown.totalNanoAiu` = `agentMetrics.main` *(226 064 000)* +
  `agentMetrics.<agentId>` *(153 013 000)*. **No hay consumo del subagente fuera de los `usage_record`**: van en el mismo `events.jsonl`, sin carpeta nueva ni `session.db`. La salida incluye también aquí
  el razonamiento *(registro 1: 288, 219 de ellos de razonamiento, y se facturan 288)*.

## S5 · Qué cambia para la spec

- **Fuente** *(D-2)*: nada; son `session.usage_record` del mismo fichero. **Q-1, cerrada para un subagente síncrono.** **Modelo** *(P-5)*: `usage.model` es el que
  se cobra; **nunca** el `model` de `subagent.*`.
- **Sesión** *(P-3 (a))*: los registros del subagente llevan el `session_ref` de la sesión que los contiene; P-3 (c) los habría separado en una sesión sin
  carpeta. **`event_id`** *(P-1 (a))*: `usageId` sigue siendo único; `(sourceSessionId, sequence)` también, pero sólo porque `sourceSessionId` cambia.
- **Límites** *(P-10)*: los subagentes **síncronos** salen; siguen sin muestra los de segundo plano y los *workflows* *(`session.db`)*.
