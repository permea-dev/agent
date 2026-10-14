# 011 · Descubrimiento, muestra 2

**Fecha**: 2026-10-14 · **Fuente**: la **copia congelada** `congelada-2026-10-14b`, de Copilot CLI **1.0.97** en Windows, plan Free y modelo «Auto». Se aplican las mismas
etiquetas y la misma privacidad que en [descubrimiento.md](descubrimiento.md). **sha256** de los 37 ficheros, antes y después: **iguales**.

## La copia *(MEDIDO)*

- **`session-state/`**: A+B, C y la carpeta del 403, **idénticas byte a byte** a la muestra 1 *(con sus 4 `.lock`)*; **D** *(09:54:08–09:55:12, 29 líneas)*;
  **E** *(09:57:13–09:58:41, 110 líneas, 838 084 B)*, con dos ficheros nuevos, `session.db` y `checkpoints/001-<título>.md` *(⛔ el nombre sale del contenido)*.
- **`session-state-home-alt/`**: **F** *(09:59:31–10:00:20, 20 líneas)*.

## M1 · D, la sesión matada *(MEDIDO)*

- **Después del último `usage_record`** *(09:54:27.663)*: el 1.er prompt se cierra entero *(`assistant.message`, `turn_end`, `usage_checkpoint`)*; del 2.º quedan
  `auto_mode_resolved`, `user.message`, `assistant.turn_start` y los 8 `model.*` de una **auxiliar** `gpt-5.4-nano` a 0 AIU *(la última, 09:55:12.575)*. **Nada más**:
  ni `usage_record`, ni `assistant.message`, ni `turn_end`, ni checkpoint, ni `shutdown`. La última línea está completa *(LF; 29/29 se decodifican)*.
- **Bloqueos**: D conserva su `.session-operation-locks/<id>.lock` y su `.workspace-fork.lock` *(0 B)*, **igual que A, C, E y F, que se cerraron bien**: en la muestra no se
  borra ninguno. Hay 6 `.lock` para 5 carpetas, y el que sobra es el de la muestra 1 *(09:38:23.9)*. **D no deja ningún huérfano nuevo.**
- **El `usage_record` de D tiene la misma forma que el de A**: las mismas 65 rutas y los mismos tipos que el 1.º de A, y que los de E y F. `interTokenLatencyMs` es opcional:
  sólo aparece en A-2 y A-3. Su registro = su último `usage_checkpoint` *(M4)*.
- **Q6, confirmada para lo escrito**: la llamada terminada está entera y **de la cortada no queda consumo en disco** *(si el servidor la facturó, NO SE PUDO MEDIR)*.
  ⇒ Una sesión matada pierde como mucho **la llamada en curso**.

## M2 · E, la compactación *(MEDIDO)*

- **`session.compaction_start`**: `systemTokens`, `conversationTokens`, `toolDefinitionsTokens`, `currentTokens`, `tokenLimit` *(int)* y `trigger` *(texto: `manual` en los dos)*.
- **El intento vacío** deja `compaction_start` y, 2 ms después, `compaction_complete` con `success: false`, `error` *(texto ⛔)* y `trigger`.
  **No tiene `compactionTokensUsed`** ni llamada al modelo entre los dos: **0 tokens y 0 créditos**.
- **`session.compaction_complete` con éxito**: `success` *(bool)*; `preCompactionTokens`, `postCompactionTokens`, `preCompactionMessagesLength`, `messagesRemoved`,
  `tokensRemoved`, `checkpointNumber` *(int)*; `summaryContent` ⛔, `checkpointPath` ⛔, `behaviorModelId` *(`gpt-6-luna`)*, `requestId`, `serviceRequestId`, `trigger`;
  `responsesReasoning.*`, `system`/`conversation`/`toolDefinitionsTokens`, `tokenLimit`; y **`compactionTokensUsed`**: `model`, `modelDisplayName`, `provider.*`,
  `requestCount` *(1)*, las cuatro partidas, `copilotUsage.{tokenDetails[], totalNanoAiu}`, `aiCreditsStatus`, `duration`. **Sin** `reasoningTokens`, `usageId` ni `sequence`.
- **La compactación SÍ consume**: **16 515 de entrada** *(= 75 + 15 335 + 1 105)*, **267 de salida**, 15 335 de lectura, 1 105 de escritura y **43 247 500 nanoAIU**
  *(la fórmula de Q8, exacta)*; **ninguna petición premium** *(`totalPremiumRequests` = 3, una por `usage_record`)*. Lo repite un `model.model_call_success` de
  `gpt-6-luna` *(`initiator: agent`, sin `callId`)*, con el mismo `requestId` y `serviceRequestId`, que declara 48 de razonamiento dentro de los 267.
- **Σ `usage_record` ≠ `session.shutdown`.** La diferencia es **exactamente la compactación**:

  | | entrada sin caché | salida | lectura | escritura | nanoAIU |
  |---|---:|---:|---:|---:|---:|
  | Σ de los 3 `usage_record` | 9 | 15 | 34 749 | 11 744 | 182 389 000 |
  | `compactionTokensUsed` | 75 | 267 | 15 335 | 1 105 | 43 247 500 |
  | **`shutdown.tokenDetails` / `totalNanoAiu`** | **84** | **282** | **50 084** | **12 849** | **225 636 500** |

  `shutdown.modelMetrics.gpt-6-luna.usage` *(46 502 · 15 · 34 749 · 11 744)* **no** la incluye; `tokenDetails`, `totalNanoAiu` y el checkpoint posterior **sí**.
  Es el **19,2 %** de los créditos de E.
- **Además, en E**: **`session.db`** *(09:58:21.805, con la compactación)*, SQLite con 12 tablas `workflow_*`/`factory_*` y `workflow_store_schema`, **0 filas** salvo
  la de esquema, entre ellas `*_credit_charges(run_id, charge_id, nano_aiu)`; `checkpoints/index.md` con 1 fila y `001-<título>.md` *(28 líneas: el resumen ⛔)*.
  `events.jsonl` sigue siendo de sólo añadir *(el offset del `shutdown` cuadra)*; la línea más larga, **56 607 B**.
- **⇒ El lector se queda corto si sólo lee `usage_record`.** Necesita también **`session.compaction_complete` con `success: true` y `compactionTokensUsed`**: un
  evento por compactación, con las partidas de Q4 *(salida = `outputTokens`)*, el modelo de `compactionTokensUsed.model`, la marca de tiempo de la línea y el id de
  `requestId` o del `id` de la línea. `shutdown.tokenDetails` no sirve de alternativa: falta cuando la sesión muere *(M1)*.

## M3 · F, con `COPILOT_HOME` *(MEDIDO)*

- **La forma es la misma** que en `~/.copilot`: `.session-operation-locks/<id>.lock` y, en `<id>/`, `events.jsonl`, `workspace.yaml`, `checkpoints/index.md`,
  `rewind-file-snapshots/tracking.json`, `files/`, `research/` y `.workspace-fork.lock`. UTF-8 sin BOM y LF. `workspace.yaml` tiene 9 claves, sin `remote_steerable`
  ni `mc_*`; D tiene 12 y A, 13, así que estas claves varían y no afectan al lector.
- **La ruta exacta**: el `system.message` ⛔ de cada sesión nombra su carpeta. Enmascarada, en A, C, D y E es `<unidad>/<x>/<x>/.copilot/session-state/<id>`, y en F,
  `<unidad>/<x>/<x>/session-state/<id>` ⇒ **`<COPILOT_HOME>/session-state/<id>/`, sin `.copilot` en medio** *(`<id>` = carpeta = `sessionId`)*.
- F tiene un `session.info` con `infoType: authentication`. Que la carpeta nueva no tuviera credenciales es una inferencia, no una medida.
- **Q1, confirmada**: `COPILOT_HOME` sustituye a `~/.copilot`, y `session-state/` cuelga directamente de ella. Sigue sin medir si hay otras variables *(`XDG_*`)*.

## M4 · Las cinco sesiones *(MEDIDO)*

| | registros | Σ entrada | salida | lectura | escritura | Σ nanoAIU | frente al `shutdown` | USD = créditos × 0,01 |
|---|---:|---:|---:|---:|---:|---:|---|---:|
| A+B | 3 | 46 231 | 30 | 30 470 | 15 752 | 228 960 000 | igual | 0,0022896 |
| C | 1 | 15 368 | 6 | 9 692 | 5 673 | 80 934 500 | igual | 0,000809345 |
| D | 1 | 15 366 | 5 | 9 692 | 5 671 | 80 859 500 | sin `shutdown`; igual al último checkpoint | 0,000808595 |
| E | 3 | 46 502 | 15 | 34 749 | 11 744 | 182 389 000 | + compactación 43 247 500 | 0,00182389 + 0,000432475 |
| F | 1 | 15 369 | 5 | 9 692 | 5 674 | 80 897 000 | igual | 0,00080897 |

- **El control de USD**: `totalNanoAiu / 10⁹ × 0,01` = `(entrada − lectura − escritura) × 0,10 + salida × 0,50 + escritura × 0,125 + lectura × 0,01` *(por millón)*.
  Da **igual exacto**, con `decimal`, en **9 de 9** `usage_record` y en la compactación. Es la misma tarifa expresada en centésimas.
- **`accounting.usageId`**: **9 distintos de 9** en las cinco sesiones, cada uno igual a su `apiCallId`. El `id` de cada línea también es único en todos los ficheros.

## Huecos nuevos

- **La compactación automática** *(`trigger` distinto de `manual`)*: no hay muestra. **`session.db`**: ¿cargan los *workflows* o subagentes créditos en
  `*_credit_charges` sin rastro en `events.jsonl`? Hoy tiene 0 filas.
