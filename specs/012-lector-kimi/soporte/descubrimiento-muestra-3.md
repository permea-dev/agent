# 012 · Descubrimiento, muestra 3: los otros tres modelos

**Fecha**: 2026-10-15 · **Fuente**: la **copia congelada** `congelada-2026-10-15c` *(hacia las 16:21 UTC)*, sobre el mismo `KIMI_CODE_HOME` que las muestras
[1](descubrimiento-muestra-1.md) y [2](descubrimiento-muestra-2.md), comparada con la 2.ª. Lleva sus etiquetas y su privacidad *(ids como `<s1>`…`<s6>`, por `createdAt`)*.
**sha256**, antes y después, de los 23, 27 y 31 ficheros de las tres copias: **iguales**.
**Lo que hizo el dueño**: una sesión nueva, con el modelo cambiado a mitad desde el selector *(`kimi-k2.7-code-highspeed` → `kimi-k2.6` → `kimi-k3`)*, y salir.

## M1 · Qué cambió entre las copias b y c *(MEDIDO)*

- **Iguales**: 22 de 27, **todos los `wire.jsonl` y `state.json` de `<s1>`…`<s5>` incluidos**.
- **Nuevos**: la carpeta de `<s6>`, bajo la **misma** `wd_*`, con `state.json`, `agents/main/wire.jsonl` *(121 líneas, 121 345 B, la última con LF)*, `logs/kimi-code.log` y
  `notify/state.json`; y 3 marcadores `.index-dirty/<s6>.<ms>` *(16:19:12.037, .054 y 16:19:21.025)*.
- **Borrados**: los 3 `.index-dirty/<s5>.<ms>` de la muestra 2.
- **Cambiados**: `session_index.jsonl` *(1 127 → 1 352 B; **creció sólo por el final, con el prefijo idéntico**)* y `.index-cache/scan.json` *(2 643 → 2 655 B, reescrito)*.
- **El índice**: una línea nueva, `{sessionId, sessionDir, workDir}`, para `<s6>`. Su `workDir` es igual tal cual al de 4 de las 5 anteriores, y normalizado, al de las 5.

## M2 · `<s6>`, registro a registro *(MEDIDO; en orden)*

| # | Hora UTC | `model` *(el alias)* | P-9 / D-4 | `inputOther` | `output` | `inputCacheRead` | `inputCacheCreation` | `usageScope` |
|---:|---|---|---|---:|---:|---:|---:|---|
| 1 | 16:15:25.397 | `moonshot-ai/kimi-k2.7-code-highspeed` | `kimi-k2.7-code-highspeed` | 9 304 | 116 | 9 472 | 0 | `turn` |
| 2 | 16:15:44.880 | `moonshot-ai/kimi-k2.7-code-highspeed` | `kimi-k2.7-code-highspeed` | 220 | 51 | 18 688 | 0 | `turn` |
| 3 | 16:16:41.845 | `moonshot-ai/kimi-k2.6` | `kimi-k2.6` | 9 504 | 88 | 9 472 | 0 | `turn` |
| 4 | 16:17:24.117 | `moonshot-ai/kimi-k2.6` | `kimi-k2.6` | 648 | 165 | 18 432 | 0 | `turn` |
| 5 | 16:17:37.726 | `moonshot-ai/kimi-k2.6` | `kimi-k2.6` | 829 | 70 | 18 432 | 0 | `turn` |
| 6 | 16:18:23.645 | `moonshot-ai/kimi-k3` | `kimi-k3` | 20 479 | 155 | **0** | 0 | `turn` |
| 7 | 16:19:21.005 | `moonshot-ai/kimi-k3` | `kimi-k3` | 181 | 102 | 20 480 | 0 | `turn` |

- **7 `usage.record`, 7 `llm.request` y 7 `turn.prompt`**: uno por pregunta. Ni compactación, ni subagente, ni `forked`. `kimi-k2.6` tuvo 3 preguntas.
- **La forma no cambia**: las mismas claves y los mismos tipos que en las muestras 1 y 2, y `usage` con exactamente las 4 partidas.
- **P-10, en las tres copias**: los **21** `usage.record` empiezan por `{"type":"usage.record"`, y ninguna otra línea. ⚠️ No todas llevan `type` primero:
  `agent.message.appended` *(14)*, `agent.turn.started` y `agent.turn.ended` *(7 y 7)* empiezan por otra clave. El prefiltro sólo depende del registro: 21 de 21.

## M3 · El cambio de modelo a mitad de sesión *(MEDIDO)*

- **Lo marca una línea `config.update`**, con `{type, agentId, modelAlias, thinkingEffort, time}` *(`modelAlias`, el alias con prefijo)*:
  - L39 *(16:16:21.286)*: `moonshot-ai/kimi-k2.6`, `thinkingEffort` `on`;
  - L88 *(16:17:48.942)*: `moonshot-ai/kimi-k3`, `max`; y en L89, 1 ms después, otra `config.update` **sin `modelAlias`** *(sólo `thinkingEffort` `max`)*.
- El modelo de arranque va en `profile.bind.modelAlias` *(L3)*. `llm.request` lleva `model` *(sin prefijo)*, `modelAlias` *(con prefijo)*, `provider` *(`openai`, el protocolo)* y `thinkingEffort`.
- **El `model` del registro cambia con él** *(= `llm.request.modelAlias`, 7 de 7)*. ⇒ **El lector no necesita `config.update`**: cada registro trae su modelo.

## M4 · La escritura de caché *(MEDIDO)*

- **`inputCacheCreation`: 0 en 7 registros de `<s6>`** *(2 de `highspeed`, 3 de `k2.6` y 2 de `k3`)*, **y 0 en los 21 de la copia c**.
- **Nada dice si sería a 5 minutos o a 1 hora**: las únicas claves de caché del fichero son `inputCacheRead` e `inputCacheCreation`, enteras, y ninguna con `ttl`,
  `ephemeral`, `5m`, `1h` ni `expire`. Están en `usage.record` y en dos copias del uso *(`agent.message.appended.message.meta.usage` y
  `context.append_loop_event.event.usage`)*, **iguales al registro en 7 de 7**.
- **El cambio de modelo no hereda la caché**: el 1.er registro de `kimi-k3` lee **0** de caché, y el 2.º, 20 480. Los 1.os de `highspeed` y `k2.6` leen 9 472 cada uno.

## M5 · Σ por modelo *(MEDIDO; para cuadrar con «Request Details»: entrada de la consola = `inputOther` + `inputCacheRead`)*

| Modelo *(P-9)* | Peticiones | Primera–última *(UTC)* | `inputOther` | `inputCacheRead` | `inputCacheCreation` | `output` | **Entrada** |
|---|---:|---|---:|---:|---:|---:|---:|
| `kimi-k2.7-code-highspeed` | 2 | 16:15:25.397–16:15:44.880 | 9 524 | 28 160 | 0 | 167 | **37 684** |
| `kimi-k2.6` | 3 | 16:16:41.845–16:17:37.726 | 10 981 | 46 336 | 0 | 323 | **57 317** |
| `kimi-k3` | 2 | 16:18:23.645–16:19:21.005 | 20 660 | 20 480 | 0 | 257 | **41 140** |
| **`<s6>`** | **7** | 16:15:25.397–16:19:21.005 | **41 165** | **94 976** | **0** | **747** | **136 141** |

**POR DECIDIR** *(el dueño; no cambia el lector)*: D-5 sólo prevé la tarifa de `kimi-k2.7-code` *(C6b)*. Los eventos de estos tres modelos llegarían con su nombre
y **sin fila de tarifa**, salvo que la plataforma la tenga.
