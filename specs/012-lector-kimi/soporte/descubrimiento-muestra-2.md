# 012 · Descubrimiento, muestra 2: el `/fork`

**Fecha**: 2026-10-15 · **Fuente**: la **copia congelada** `congelada-2026-10-15b` *(hacia las 09:17 UTC)*, sobre el mismo `KIMI_CODE_HOME` que la
[muestra 1](descubrimiento-muestra-1.md), comparada con ella. Lleva sus etiquetas y su privacidad *(ids como `<s1>`…`<s5>`, por `createdAt`)*.
**sha256**, antes y después, de los 23 ficheros de la 1.ª copia y los 27 de la 2.ª: **iguales**.
**Lo que hizo el dueño**: `kimi --session` → la sesión de S1 → `/fork` → `/exit`, sin preguntar nada; después, `kimi --resume <fork>`, 1 prompt corto y `/exit`.

## M1 · Qué cambió entre las dos copias *(MEDIDO)*

- **Iguales**: 18 de 23. **La sesión original `<s1>` no cambió nada**: ni su `wire.jsonl` *(38 líneas, 98 171 B)*, ni su `state.json`.
- **Nuevos**: la carpeta del fork `<s5>`, con `state.json`, `agents/main/wire.jsonl` *(56 líneas, 103 475 B)*, `logs/kimi-code.log` y `notify/state.json`, y 3 marcadores
  `.index-dirty/<s5>.<ms>` *(09:16:31.900, .948 y 09:16:37.814)*.
- **Borrados**: los 3 `.index-dirty/<s4>.<ms>` de la sesión matada. El índice se reparó en el siguiente arranque.
- **Cambiados**: `session_index.jsonl` *(900 → 1 127 B; **creció sólo por el final, con el prefijo idéntico**)* y `.index-cache/scan.json` *(1 605 → 2 643 B, **reescrito**)*.

## M2 · La original tras abrirla y bifurcar *(MEDIDO)*

- **No se añadió ninguna línea** y hay **0 `usage.record` nuevos**: abrir una sesión y hacer `/fork` **no escribe nada en ella**. Tampoco el bloque de arranque
  *(`runtime.set_binding`, `profile.bind`, `plugin.session_start`)*.

## M3 · El fork *(MEDIDO)*

- **Las líneas 1–38 son el `wire.jsonl` original entero, byte a byte** *(98 171 B)*, con sus **2 `usage.record`**: idénticos en bytes, en número y en Σ
  *(9 531 · 286 · 28 160 · 0)*.
- **La línea 39 es `{"type":"forked", agentId: texto, time: int}`**, a las 09:14:45.432. `state.json` del fork lleva **`forkedFrom` → `<s1>`** *(la 14.ª clave)*;
  su `createdAt` es 09:14:45.437.
- **Después de `forked`**, la reanudación **añadió al mismo fichero sin reescribir el prefijo**: 17 líneas entre 09:16:31.908 y 09:16:37.802.
  - Hay **1 `llm.request` y 1 `usage.record`** *(09:16:37.790; 3 282 · 107 · 15 872 · 0; `moonshot-ai/kimi-k2.7-code`)*.
  - **No hay bloque de arranque en el `wire.jsonl`**: sólo `turn.prompt`, `context.*`, `agent.*`, `llm.request`, `usage.record`, `token_counting.*`, `turn.ended` y `prompt.completed`.
    El log sí abre con la pareja de **184 y 56 B** *(09:16:31.923)*.
- **Corrige M2 de la muestra 1**:
  - **el argumento del `wire.jsonl` no vale**, porque un proceso que reanuda no repite el bloque de arranque;
  - **el del log, sí**, porque la línea de 184 B sale en la 1.ª petición de cada proceso, y `<s1>` sólo tiene una.
  - La conclusión se mantiene: los 2 prompts de `<s1>` son de un solo proceso, y **S2 no consta**.

## M4 · `session_index.jsonl` *(MEDIDO)*

- **Una línea nueva al final**, con las claves `sessionId` *(= `<s5>`)*, `sessionDir` y `workDir`. El fork está bajo la **misma** carpeta `wd_*`.
- ⚠️ **Su `workDir` es distinto del de las otras 4 sólo en el estilo de las barras** *(misma longitud, y igual al normalizar `\`→`/`)*. Coincide con el `cwd` de su
  `state.json`. ⇒ **El proyecto se normaliza antes del hash** *(P-8)*.

## M5 · El `event_id` con las dos copias *(MEDIDO; sin decidir)*

- **14 `usage.record` en la 2.ª copia: 12 propios y 2 heredados.**
- **(a) Posición** *(`sessionId`, `agentId`, ordinal)*:
  - **sin frontera**, 14 ids distintos de 14, así que **los 2 heredados contarían dos veces**;
  - **con la frontera de la última `forked`**, los 2 heredados se descartan y quedan **12 emitidos, 12 distintos**.
- **(b) Contenido** *(`agentId`, `time`, `model` y las 4 partidas)*: 12 distintos de 14.
  - **Las 2 repeticiones son exactamente los heredados del fork frente a sus originales** *(09:02:44.105 y 09:03:08.264)*, así que la plataforma descartaría la copia.
  - **Ninguna colisión entre los 12 registros legítimos.**
- **Las dos deduplican esta muestra.** (a) necesita la frontera, que es una línea con forma fija y medida; (b) no la necesita, pero mete horas y partidas en el hash.
- **Recomendación para P-3** *(sin decidir)*: **(a) con la frontera `forked`**, porque respeta la regla de 011. **(b)** quedaría como red, si se acepta romperla.

## M6 · El total de la mañana *(MEDIDO; cada petición una vez, sin los heredados)*

| Modelo | Peticiones | Primera–última *(UTC)* | `inputOther` | `inputCacheRead` | `inputCacheCreation` | `output` |
|---|---:|---|---:|---:|---:|---:|
| `moonshot-ai/kimi-k2.7-code` | **12** | 09:02:44.105–09:16:37.790 | **33 831** | **186 112** | **0** | **3 170** |

- **Entrada total: 219 943.** Es la muestra 1 *(11 peticiones)* más la del fork. **Sigue POR MEDIR**: el cuadre con la consola de Kimi Platform, el razonamiento dentro de `output` y si se facturó la petición matada de S6.
