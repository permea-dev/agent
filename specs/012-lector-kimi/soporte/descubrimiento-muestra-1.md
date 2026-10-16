# 012 · Descubrimiento, muestra 1

**Fecha**: 2026-10-15 · **Fuente**: la **copia congelada** `congelada-2026-10-15`, de Kimi Code **2.2.0** en Windows, con `KIMI_CODE_HOME` de prueba, Kimi Platform
con clave de API, modelo `kimi-k2.7-code` y razonamiento activado. Las etiquetas son **MEDIDO** *(leído en la copia con Python `-I`, desde scripts y temporales
fuera del repo)* y **POR MEDIR**; la privacidad, la de [descubrimiento.md](descubrimiento.md). **sha256** de los 23 ficheros, antes y después: **iguales**.

> **Privacidad.** De la copia salen sólo tipos de línea, claves con su tipo JSON, recuentos, tokens, nombres de modelo, la versión y horas UTC. Los ids se
> comparan sin imprimirse *(`<s1>`…`<s4>`, por `createdAt`)*. De las líneas que no son `usage.record`, `forked`, `metadata` o `subagent.*` sólo se leyeron el tipo y las claves.
> **Corrección del dueño**: S3 *(`/fork`)* **no se hizo**; queda para la muestra 2 *(`congelada-2026-10-15b`)*.

## La copia *(MEDIDO)*

- **Una sola carpeta `wd_*` y 4 sesiones**, con **5 `wire.jsonl`**. Todo es UTF-8 sin BOM y con **LF** *(0 CRLF)*; los 5 `wire.jsonl` y `session_index.jsonl` terminan en `\n`:

  | Carpeta | Agente | Líneas | Horas UTC | Sesión del encargo |
  |---|---|---:|---|---|
  | `<s1>` | `main` | 38 | 09:02:35–09:03:08 | **S1** *(¿y S2? Ver M2)* |
  | `<s2>` | `main` | 60 | 09:04:17–09:06:00 | **S4** *(3 prompts + `/compact`)* |
  | `<s3>` | `main` y `agent-0` | 37 y 45 | 09:06:40–09:07:14 | **S5** *(principal + subagente)* |
  | `<s4>` | `main` | 12 | 09:08:06 | **S6** *(la matada)* |

- **Los otros 18 ficheros**:
  - `session_index.jsonl` *(4 líneas `{sessionId, sessionDir, workDir}`, todas con el mismo `workDir`)*;
  - por sesión, `state.json` ×4 *(13 claves; **sin `forkedFrom`**)*, `logs/kimi-code.log` ×4 *(texto, una línea por arranque y por llamada)* y `notify/state.json` ×4 *(`{enabled}`)*;
    y en S5, `agents/main/file-history/<hash>` *(35 B: el subagente editó un fichero)*.
  - **Nuevos respecto al descubrimiento**: `sessions/.index-cache/scan.json` *(`{version: 1, sessions}`, con 3 sesiones: falta S6)* y `sessions/.index-dirty/<sessionId>.<ms>` ×3
    *(0 B, todos de S6, a las 09:08:06.260, .298 y .333)*. **No son `wd_*`: el lector no entra.**

## M1 · La forma real de `usage.record` *(MEDIDO, 11 registros)*

- **Claves, siempre en este orden**: `type`, `agentId` *(= la carpeta del agente, 11/11)*, `model` *(texto)*, `usage`, `usageScope` *(texto)* y `time` *(int, ms)*.
  Las 11 líneas empiezan por `{"type":"usage.record"`: **el filtro por prefijo funciona**.
- **`usage` = `{inputOther, output, inputCacheRead, inputCacheCreation}`**, los cuatro int. **No hay `raw`**: 0 de 11, también buscado en los bytes.
  ⇒ El razonamiento **no aparece aparte**. Si va dentro de `output` es **POR MEDIR**, contra la consola *(M8)*.
- **Escritura de caché: 0** en los 11 *(protocolo OpenAI de Kimi)*. **La lectura de caché sí es grande**, incluso en la 1.ª llamada de una sesión nueva *(16 896 en S4 y 18 688
  en S5)*: Kimi reutiliza la caché **entre sesiones**. La 1.ª llamada del subagente lee 0.
- **El modelo**: `moonshot-ai/kimi-k2.7-code` en los 11. Es **el alias**: igual al `modelAlias` de la `llm.request` anterior *(11/11)* y **distinto** de su `model` *(11/11)*.
- **El protocolo *wire*** es `1.5` en los 5 ficheros, con `metadata` en la 1.ª línea. Hay **una `llm.request` por cada `usage.record`** *(2/2, 4/4, 2/2 y 3/3)*, **sin `attempt`**:
  no hubo reintentos. S6 tiene 1 `llm.request` y 0 registros *(M6)*.

## M2 · S1+S2, la reanudación: **no consta en la copia** *(MEDIDO; la explicación, POR MEDIR)*

- `<s1>` tiene **2** `turn.prompt` *(09:02:36 y 09:03:03)* y **2** `usage.record`. Con S1 = 2 prompts y S2 = 1, cabrían 3.
- **No hay ninguna marca de un segundo proceso**: `metadata`, `runtime.set_binding`, `profile.bind`, `plugin.session_start` y `llm.tools_snapshot` aparecen **una vez**, a las 09:02:35–37.
  Las otras sesiones dejan ese bloque **al arrancar cada proceso o agente**: `agent-0` tiene el suyo, y el log de S5 repite la pareja de arranque *(184 y 56 B)* a las 09:06:48.381.
  El log de `<s1>` tiene una sola línea de 184 B.
- `<s1>` **no cambia después de las 09:03:08**: ni su `wire.jsonl`, ni su `state.json` *(`updatedAt` 09:03:08.346)*, ni su entrada en `scan.json` *(09:03:08.637)*. Entre S1 y S4 hay 69 s.
- **Lectura**: los dos prompts de `<s1>` son de **un mismo proceso**. **El `kimi --continue` de S2 no escribió en este `KIMI_CODE_HOME`**: o se lanzó desde una consola sin la
  variable *(y fue a la raíz real)*, o no llegó a enviar el prompt. **POR MEDIR**: repetir S2 en la muestra 2, con `echo $env:KIMI_CODE_HOME` antes.

## M3 · S3, `/fork` *(POR MEDIR)*: no se hizo *(corrección del dueño)*. Hay 0 `forked` y 0 `forkedFrom`; todo queda para la muestra 2

## M4 · S4, la compactación *(MEDIDO)*

- `full_compaction.begin` *(09:05:39.861)* → `llm.request` con **`kind: compaction`** → **`usage.record` con `usageScope: session`** *(09:06:00.558)* →
  `context.apply_compaction` → `token_counting.rebased` → `full_compaction.complete` *(09:06:00.581)*.
- **Partidas**: `inputOther` 1 969, `inputCacheRead` 19 200, **`inputCacheCreation` 0** y `output` 661. **Es la única con `session`**; las 10 restantes son `turn`.
  ⇒ **La compactación consume y queda en un `usage.record` normal**: no hace falta una fuente aparte.

## M5 · S5, el subagente *(MEDIDO)*

- **Tiene su propio `agents/agent-0/wire.jsonl`**: 45 líneas y **3 `usage.record`** *(09:06:53.153, 09:07:05.645 y 09:07:08.361)*, intercalados en el tiempo con los 2 del principal.
  `state.json` lo declara `type: sub` y `parentAgentId: main`. El modelo de `subagent.spawned.model` = el de sus `usage.record`.
- **`subagent.completed.usage`, en el `wire.jsonl` del principal** *(09:07:08.370)*: 15 798 · 268 · 30 976 · 0. Es **= Σ de los 3 registros del subagente** en las 4 partidas.
  ⇒ **Leerlo contaría el subagente dos veces: se ignora** *(P-1)*. `subagentId` = la carpeta `agent-0`.

## M6 · S6, la sesión matada *(MEDIDO)*

- 12 líneas: el arranque, `turn.prompt` y **una `llm.request` a las 09:08:06.352**. **Ningún `usage.record`**: la llamada en vuelo **no deja consumo en disco**.
- **La última línea está completa** *(LF; 12/12 se decodifican)* y **no hay `.bak`**. `state.json` no tiene `lastTurnReason` *(los otros tres dicen `completed`)*.
  ⇒ Como en Copilot, **se pierde como mucho la llamada en curso**. **POR MEDIR**: si Kimi Platform la facturó *(M8)*.

## M7 · Registros a 0 *(MEDIDO)*: **ninguno** *(0 de 11)*

## M8 · Por sesión y por modelo *(MEDIDO; todos `moonshot-ai/kimi-k2.7-code`)*

| Sesión | Fichero | Registros | Primera–última *(UTC)* | `inputOther` | `inputCacheRead` | `inputCacheCreation` | `output` |
|---|---|---:|---|---:|---:|---:|---:|
| S1 | `<s1>/main` | 2 | 09:02:44.105–09:03:08.264 | 9 531 | 28 160 | 0 | 286 |
| S4 | `<s2>/main` | 4 *(1 `session`)* | 09:04:27.330–09:06:00.558 | 4 604 | 73 728 | 0 | 2 047 |
| S5 | `<s3>/main` | 2 | 09:06:48.302–09:07:13.986 | 616 | 37 376 | 0 | 462 |
| S5 | `<s3>/agent-0` | 3 | 09:06:53.153–09:07:08.361 | 15 798 | 30 976 | 0 | 268 |
| S6 | `<s4>/main` | 0 | — | 0 | 0 | 0 | 0 |
| **Σ** | | **11** | 09:02:44–09:07:14 | **30 549** | **170 240** | **0** | **3 063** |

- **Entrada total** *(sin caché + caché)*: **200 789**. **Cuadre con la consola de Kimi Platform** *(POR MEDIR)*: entrada facturada ≟ 200 789, con la caché a su tarifa. Salida ≟ 3 063: si la consola da **más**, el
  razonamiento queda fuera de `output`. Y si cuenta una llamada más a las 09:08 *(la de S6)*.

## M9 · El `event_id` con estos datos *(MEDIDO; sin decidir)*

- **(a) Posición** *(`sessionId`, `agentId`, ordinal)*: **11 distintos de 11**. **(b) Contenido** *(`agentId`, `time`, `model` y las 4 partidas)*: **11 distintos de 11**.
  Ninguna colisión, y (a) ∩ (b) = 0. Sin registros a 0, el riesgo de (b) no aparece.
- **Las copias de `/fork` no se pueden probar**: no hay bifurcación *(M3)*. Que (a) con la frontera `forked` y (b) las deduplican es **POR MEDIR** en la muestra 2. **Recomendación** *(sin decidir)*: **(a)**,
  a confirmar con la muestra 2. Aquí funciona igual que (b), y no mete horas ni partidas en el hash *(regla de 011)*.

## Para la muestra 2 *(`congelada-2026-10-15b`)*

1. **S2 de nuevo**: `echo $env:KIMI_CODE_HOME` y luego `kimi --continue` sobre una sesión con un prompt terminado, más 1 prompt.
2. **S3**: `/fork` en esa sesión, `kimi --session <fork>` y 1 prompt.
3. Una **captura de la consola de Kimi Platform** con el uso por modelo de la ventana de 09:02 a 09:09 UTC, para cuadrar M8.
