# Contrato — Derivación del `event_id` de Gemini CLI (P-009)

**Feature**: `009-lector-gemini` | **Fecha**: 2026-10-08 | **Estado**: **ratificado** el 2026-10-08, 12:20 *(P-2 (a), P-3 (a))* |
**Requisitos**: P-009 FR-006 a FR-010, FR-017

Este contrato fija **qué valor** tiene el `event_id` de un evento de Gemini CLI y **cuándo se emite**. Es **hermano** de los de 006
(`specs/006-medicion-fiel/contracts/event-id.md`) y 008 (`specs/008-lector-codex/contracts/event-id-codex.md`): el mismo esquema, con un
espacio de nombres propio. **No modifica ninguno de los dos.** La forma del campo es la del contrato de frontera.

---

## Entrada

Una **respuesta** es cada aparición de un mensaje `type: "gemini"` con un objeto `tokens`, en un `.jsonl` de sesión de Gemini CLI
*(≥ 0.39.0)*. Aparece en una línea de mensaje o como elemento de un `$set.messages` *(`soporte/descubrimiento.md` Q2, Q4)*. De ella se
lee **un identificador técnico del proveedor**:

| Campo del mensaje | Nombre aquí | Forma observada *(2026-10-08; 16 apariciones, 10 valores)* |
|---|---|---|
| `id` | `M` | UUID v4 que genera la CLI (`randomUUID`) *(Q4)* |

Se usa **sólo** para derivar y deduplicar. **NUNCA** se copia a ningún campo del evento, ni entero ni en fragmento.

## Derivación

```
componentes = [ "permea/event_id/v1", "gemini", "respuesta", M ]
codificado  = concatenación de  ( longitud_big_endian_uint32(c) ‖ bytes_utf8(c) )  para cada c
event_id    = hex_minúsculas( SHA-256(codificado)[0:16] )        → 32 caracteres [0-9a-f]
```

| Caso | Identificadores | Emite |
|---|---|---|
| `M` no vacío, primera aparición con `tokens` en la pasada | `M` | sí |
| `M` ya emitido en la pasada, o ya visto antes en el mismo fichero *(P-3)* | `M` | **NO**; se cuenta como «repetida» |
| `M` ausente, vacío o no textual | — | **NO**; se cuenta como «sin identificador» |
| `session-*.json` *(≤ 0.38)* | — | **NO**; el fichero se cuenta como «formato anterior» *(D-3)* |

**Ningún otro dato entra en el hash**: ni sal, ni máquina, ni desarrollador, ni organización, ni `sessionId`, ni `projectHash`, ni
fichero, ni carpeta, ni contenido. Tampoco el modelo, el `timestamp` ni las partidas.

**El espacio de nombres separa las herramientas.** `"gemini"` va donde 006 pone `"claude_code"` y 008 pone `"codex"`, y `"respuesta"` en
la posición de `TIPO`, como en 008. El mismo valor llegado de dos herramientas da `event_id` distintos por construcción.

## Por qué basta el `id` *(P-2)*

- **Es único**: la CLI lo genera con `randomUUID` en `ChatRecordingService.newMessage`. Las respuestas con tokens nacen en `recordMessage`,
  `recordSyntheticMessage` o `recordToolCalls`, y en la 0.63.0 ningún llamador les pasa un `id` *(`core/src/services/chatRecordingService.ts`
  y `core/src/core/geminiChat.ts`; descubrimiento Q4)*.
- **Es estable**: `pushMessage` y `updateMessagesFromHistory` lo conservan cada vez que reescriben el mensaje *(Q4)*. **Medido**: 16
  apariciones, 10 `id`, y 0 `id` con dos juegos de tokens distintos.
- **Las copias tienen el mismo `id`, y también el mismo `sessionId`**. Pasa con la copia de carpeta por la migración *(Q1)*, la conversión
  `.json` → `.jsonl` al reanudar *(Q10)* y el `$set.messages` de una reanudación *(Q8)*. Añadir el `sessionId` no separaría nada que
  deba separarse, y obligaría a leer la cabecera del fichero en cada pasada.
- **Si una función futura copiara respuestas a otra sesión** *(como la bifurcación de Codex)*, con sólo el `id` darían **un** evento. Con
  el `sessionId`, darían dos.

## Garantías

- **Determinista**: la misma respuesta da el mismo `event_id` en cualquier pasada, fichero, carpeta e instalación.
- **Repeticiones en el fichero** *(`recordToolCalls`, `recordMessageTokens`, `$set.messages`)*: el mismo `event_id`. Se emite **la
  primera aparición con `tokens`**, y las demás son «repetidas».
- **Reanudación días después**: el `$set.messages` vuelve a traer respuestas ya emitidas. Con P-3 (a), no se reemiten. Si aun así se
  reencolan *(estado perdido, fichero reescrito)*, la plataforma las descarta.
- **Idempotencia en la plataforma**: índice único `(org_id, event_id)` *(`permea-platform`
  `backend/database/migrations/2026_07_04_000004_add_unique_org_event_to_metric_events.php:16`)* y
  `insert … on conflict ("org_id", "event_id") do nothing` *(`backend/app/Ingest/IngestBatchService.php:92`)*. Es la misma base que
  sostiene la bifurcación de Codex *(008, contrato §Garantías)*.
- **El identificador del proveedor nunca viaja**: ni `id`, ni `sessionId`, ni `projectHash`, ni en el evento, ni en la cola, ni en nada
  que se transmita. **Excepción declarada**: los nombres de fichero llevan los 8 primeros caracteres del `sessionId`, y la carpeta de un
  subagente, el `sessionId` del padre. La ruta es la clave de `state.json`: queda **en local** y no viaja *(FR-017)*.

## Vectores de prueba (normativos)

Identificadores **sintéticos**, sin la forma de UUID ni de ningún identificador real:

| Caso | Entrada | `event_id` esperado |
|---|---|---|
| respuesta | `M = m-000000000000000000000001` | `027c36e1f42d164e670770cee1a06f83` |
| otra respuesta | `M = m-000000000000000000000002` | `53120e5f5cfd342c9c48f37cd857c697` |
| el mismo valor en el espacio de Codex | `["permea/event_id/v1", "codex", "respuesta", "m-…0001"]` | `07d1f3ea70c0b94474fe27e18fdb89b7` *(≠ el primero)* |
| rastro de P-2 (b), **no normativo** | `["permea/event_id/v1", "gemini", "respuesta", "s-000000000000000000000001", "m-…0001"]` | `4b416a0d7c2b1ead78845a9bcdf62ca3` |

Calculados el 2026-10-08 con una implementación independiente *(Python `hashlib` + `struct`)*, sin código Go del repo. Como control del
instrumento, la misma implementación reproduce el vector «sólo `message.id`» de 006 *(`1692269369e3bb2b418279566f4b093f`)* y el
«respuesta» de 008 *(`31439e3953a0916dde2f98b748ceb985`)*. La implementación en Go **DEBE** reproducir los tres primeros byte a byte.

## Una pasada, un evento por respuesta

- **No hay espera ni cierre**: cada aparición con `tokens` es una respuesta terminada (la CLI los registra al acabar el *stream*), y la
  pasada de 007 no se aplica.
- **Dentro de una pasada**, un `event_id` ya emitido no se vuelve a emitir: es una «repetida». Es el caso de la misma sesión en dos
  carpetas.
- **Entre pasadas, dentro del mismo fichero** *(P-3 (a))*: tampoco se reemite la respuesta que ya apareció antes del offset.
- **Entre ficheros y entre pasadas** no hay memoria: un `event_id` repetido puede volver a encolarse, y la plataforma lo descarta.
- **Qué es una pasada**: lo mismo que en 006 *(una ejecución de `--run`, un ciclo de `--daemon` o una ejecución de `--scan`)*.

## Lo que este contrato NO cambia

- **Los contratos de 006 y 008, y su implementación**: `eventid.go`, `eventid_test.go`, `codex_eventid.go`, `codex_eventid_test.go`,
  `event-id.md` y `event-id-codex.md`, con 0 bytes de diff. La derivación de Gemini vive en un fichero y una función nuevos.
- **La forma del campo**: 32 hex.
- **La allowlist del evento**, `SchemaVersion = 1` e `internal/event`.
- **El dominio** `permea/event_id/v1`: Gemini se distingue por el espacio de nombres, no por la versión.
