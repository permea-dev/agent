# 008 · Descubrimiento del lector de Codex

**Fecha**: 2026-10-07 · **Fuentes**: el 🔍 del 07-10, sobre las sesiones vivas, y la FASE 0 del Encargo 2, sobre la **copia congelada**.

**Cómo se lee esto.** Cada afirmación lleva una etiqueta:
- **MEDIDO**: en la copia, con el guión citado;
- **LEÍDO EN LA FUENTE**: Codex, etiqueta `rust-v0.160.1` *(commit `d27764b`)*, rutas bajo `codex-rs/`; o nuestro código en `7b8c77c`;
- **NO SE PUDO MEDIR**: y qué muestra haría falta.

> **Privacidad.** De las sesiones salen sólo nombres de campo, tipos, recuentos, sumas, horas, nombres de modelo y versiones. Los
> ficheros se nombran **F1…F8**. Las rutas son genéricas: `~/.codex/sessions` y `<copia congelada>`.

## La copia congelada

- **Qué es**: 8 `.jsonl`, de los mismos tamaños que las sesiones del 🔍. Se lee **en su sitio**.
- **Huella SHA-256 del conjunto**: `984d483c12a15601`, con 8 ficheros. Igual antes y después del Encargo 2.
  ```sh
  (cd "<copia congelada>" && find . -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum | cut -c1-16)
  ```

## Los ficheros *(MEDIDO)*

Patrón: `sessions/AAAA/MM/DD/rollout-AAAA-MM-DDTHH-MM-SS-<uuid>.jsonl`. La carpeta y la hora del nombre son **locales**: F5 se llama
`…T11-16-03`, y su `session_meta` dice 09:16:03Z.

| F | Fecha *(UTC)* | Bytes | Líneas | Línea máx. | Codex | `source` | Modelos | `token_count` con `info` | `token_usage_record` |
|---|---|---:|---:|---:|---|---|---|---:|---:|
| F1 | 2026-02-15 10:33–10:59 | 949 055 | 621 | 20 199 | 0.101.0 | cli | gpt-5.3-codex | 107 | 0 |
| F2 | 2026-02-15 16:35–16:38 | 286 866 | 189 | 19 501 | 0.101.0 | cli | gpt-5.3-codex | 27 | 0 |
| F3 | 2026-02-15 17:23–17:30 | 611 600 | 414 | 44 745 | 0.101.0 | cli | gpt-5.3-codex | 93 | 0 |
| F4 | 2026-02-15 17:33–17:36 | 287 098 | 191 | 41 681 | 0.101.0 | cli | gpt-5.3-codex | 39 | 0 |
| F5 | 2026-10-07 09:16 | 31 642 | 8 | 22 496 | 0.160.1 | exec | gpt-5.3-codex | 0 | 0 |
| F6 | 2026-10-07 09:17–09:25 | 154 662 | 61 | 22 497 | 0.160.1 | vscode | gpt-5.3-codex → gpt-6-luna | 4 | 4 |
| F7 | 2026-10-07 09:20 | 39 564 | 15 | 18 983 | 0.160.1 | exec | gpt-6-luna | 1 | 1 |
| F8 | 2026-10-07 09:20–09:21 | 47 870 | 22 | 18 983 | 0.160.1 | exec | gpt-6-luna | 2 | 2 |

0 líneas corruptas, y todas acaban en `\n`. Un único `session_meta` por fichero, en la línea 1.

## Dos formatos

- **0.101.0 *(F1–F4)***: líneas `{timestamp, type, payload}`. El consumo sólo está en `event_msg/token_count.info`, con
  `total_token_usage` *(acumulado)* y `last_token_usage` *(parcial)*, **sin ningún identificador**.
  - **MEDIDO**: el primero de cada fichero trae `info` nulo, y cada uno sale **dos veces** con `info` idéntico *(53/107, 13/27, 46/93,
    19/39)*.
  - Σ `last` de todos da el doble. Sumando sólo los de acumulado nuevo sale el total final exacto.
  - Un `turn_context` **por petición** *(F1: 54 en 2 turnos)*, sin `effort`.
- **0.160.1 *(F5–F8)***: las líneas ganan `ordinal`, contiguo desde 0, y aparecen `token_usage_record`, `world_state`, `compacted`,
  `event_msg/thread_settings_applied` y `event_msg/item_completed`.
  - **`token_usage_record`**: `usage` *(la respuesta)*, `turn_token_usage`, `thread_token_usage`, `response_id`, `turn_id`,
    `root_turn_id`, `session_id` y `thread_id`. Las partidas son `input_tokens`, `cached_input_tokens`, `cache_write_input_tokens`,
    `output_tokens`, `reasoning_output_tokens` y `total_tokens`.
  - Un `turn_context` **por turno**, con `model` y `effort`.
- **LEÍDO EN LA FUENTE**:
  - un `token_usage_record` por respuesta completada con uso *(`core/src/session/mod.rs:4768-4800`)*, siempre persistido
    *(`rollout/src/policy.rs:20`)*;
  - «Persist one `TurnContextItem` per real user turn» *(`core/src/session/mod.rs:4745-4747`)*.

## El consumo

**Σ `usage` = `thread_token_usage` final** en F6, F7 y F8. Sin repetidos ni retrocesos. **MEDIDO**.

**El acumulado de `token_count` no es fiable en la 0.160.1.** **MEDIDO**: en F6 acaba en 48 629 de entrada y 33 de salida. Le faltan los
17 647 / 17 152 / 55 de la respuesta de compactación, que `thread_token_usage` sí cuenta. Tras la compactación, un `token_count` trae un
parcial que es una estimación: sólo `total_tokens` = 5394.

**LEÍDO EN LA FUENTE**:
- la compactación remota apunta su uso en el registro, no en `total_token_usage` *(`core/src/compact_remote_v2.rs:304-305`)*;
- el parcial se recalcula *(`:373`)*.

**Caché y razonamiento**:
- **LEÍDO EN LA FUENTE**: la copia de la API es literal *(`codex-api/src/sse/responses.rs:128-151`)*. La caché es un subconjunto de la
  entrada: `non_cached_input = input − cached` *(`protocol/src/protocol.rs:2430-2432`)*. El razonamiento va dentro de la salida *(test
  `responses.rs:802-829`)*.
- **MEDIDO**: caché ≤ entrada y razonamiento ≤ salida en todos los registros. `total` = entrada + salida en todos menos la estimación.

**Peticiones fallidas.**
- **MEDIDO**: en F6, los dos primeros turnos acaban con `task_complete.error`, con un mensaje que contiene «400». No dejan ni
  `token_count` ni `token_usage_record`.
- **MEDIDO**: F5 acaba tras el mensaje del usuario, sin `task_complete` y sin consumo.
- **LEÍDO EN LA FUENTE**: sin `response.completed` no hay uso *(`responses.rs:413-450`)*.

**Coste.** **MEDIDO**: no hay ningún campo de coste. Hay límites de uso *(`rate_limits`: `primary`/`secondary`, ventanas de 300 y
10 080 minutos)*, y `service_tier` = `default` sólo en `thread_settings_applied`.

## El contraste 2828 / 4117 *(MEDIDO)*

El dueño vio «tokens used» **2828** y **4117** en las dos sesiones `exec` con consumo, F7 y F8 por orden de creación.

| F | `thread_token_usage` final *(entrada / caché / salida)* | entrada − caché + salida |
|---|---|---:|
| F7 | 13 831 / 11 008 / 5 | **2828** ✅ |
| F8 | 27 944 / 24 064 / 237 | **4117** ✅ |

**LEÍDO EN LA FUENTE**: `blended_total` = `(input − cached).max(0) + output` *(`exec/src/event_processor_with_human_output.rs:392-397,510-514`)*.

## El modelo

**MEDIDO**:
- en F6: T1 y T2 con `gpt-5.3-codex`, fallidos; tres `thread_settings_applied` a `gpt-6-luna`; T3, T4 y T6 con `turn_context`
  `gpt-6-luna`/`low`;
- 3 de los 4 registros de F6 tienen un `turn_context` con su `turn_id`. El de la compactación, T5, no lo tiene: el modelo vigente,
  según el `turn_context` y el `thread_settings_applied` anteriores, es `gpt-6-luna`;
- en F7 y F8, todos tienen el suyo.

**LEÍDO EN LA FUENTE**: la compactación usa el modelo del hilo, salvo en un cambio de modelo a la baja, en el que usa el anterior
*(`core/src/compact_remote_v2.rs:253-290,402-404`)*.

## Identidad *(MEDIDO)*

- `response_id`: sólo en `token_usage_record`, 0 repetidos entre los 7.
- `session_id` = `thread_id` = `session_meta.id` en todos los registros.
- `token_count` no lleva ningún identificador.

**LEÍDO EN LA FUENTE**:
- el `id` del hilo es el mismo al reanudar y nuevo al bifurcar *(`core/src/session/session.rs:858-861`)*;
- `session_id` es el del hilo raíz *(`protocol.rs:3130`)*.

## Proyecto *(MEDIDO)*

`cwd` presente en `session_meta` y en cada `turn_context`. `git`: con 3 claves en F1–F4 y vacío en F5–F8. **LEÍDO EN LA FUENTE**: `git`
se omite fuera de un repositorio *(`git-utils/src/info.rs:77-86`)*.

## Dónde, y cómo cambian los ficheros *(LEÍDO EN LA FUENTE)*

- **Dónde**: `CODEX_HOME` si está y no es vacía; si no, el directorio personal + `.codex` *(`utils/home-dir/src/lib.rs:14-16,52-59`)*.
  Dentro, `sessions/AAAA/MM/DD/`, en hora local *(`rollout/src/recorder.rs:1722-1744`)*. En Windows, el directorio personal lo resuelve
  el crate externo `dirs`: **NO SE PUDO LEER**.
- **En vivo sólo se añade al final** *(`recorder.rs:1756-1760`)*.
- **Dos tareas reescriben ficheros renombrando**, desactivadas por defecto *(`features/src/lib.rs:1171-1186`)*: la compresión a
  `.zst` de lo que tiene más de 7 días y la migración de formato.
- **Bifurcación**: otro fichero, con **copia** de los registros del padre *(`core/src/session/mod.rs:1655-1660`)*.
- **Retención**: no hay borrado automático. Sólo borra un borrado explícito del hilo *(`thread-store/src/local/delete_thread.rs:1-7`)*.
- **El `mtime`**: Codex lo fija a mano al abrir un fichero nuevo *(`recorder.rs:1747,1761-1762`)*. **MEDIDO** en el 🔍, sobre las vivas:
  en F5, F7 y F8 el `mtime` era anterior al último registro *(1,2 s, 4,0 s y 18,4 s)*. Por qué: **NO SE PUDO MEDIR**.

## FASE 0 *(Encargo 2)*

### (a) La reanudación de F6 — ✅ hay una, en el mismo fichero

**MEDIDO**: horas de inicio de los turnos de F6, numerados dentro de F6.

| Turno | Inicio *(UTC)* | Hueco desde el turno anterior | Qué es |
|---|---|---|---|
| T1 | 09:17:59.115 | — | error |
| T2 | 09:18:33.344 | 34 s | error |
| T3 | 09:19:30.851 | 57 s | pregunta, `gpt-6-luna` |
| — | **09:23:48.450** | **4 min 14,5 s** tras el fin de T3 *(09:19:33.972)* | **`thread_settings_applied` suelto, sin turno** |
| T4 | 09:24:45.719 | 57 s tras el anterior | pregunta |
| T5 | 09:24:54.550 | 9 s | compactación *(`compacted`, sin `turn_context`)* |
| T6 | 09:25:12.061 | 18 s | pregunta |

**Qué se ve**:
- El acumulado sigue a través del hueco: `thread_token_usage` pasa de 17 600 a **35 216** = 17 600 + 17 616 de entrada.
- El `ordinal` sigue contiguo *(29 → 30)*.
- No hay otro `session_meta`, ni se repite ningún registro *(4 `response_id` distintos, y su suma = el acumulado final)*.

**LEÍDO EN LA FUENTE**:
- la reanudación abre el mismo fichero para añadir, sin `session_meta` nuevo *(`rollout/src/recorder.rs:981-992`)*;
- el `ordinal` continúa *(`rollout/src/ordinal.rs:95-99`)*;
- escribe un `thread_settings_applied` «even when no turn follows the resume» *(`core/src/session/mod.rs:1607-1611`)*.

**Veredicto**: F6 contiene una reanudación a las 09:23:48. Demuestra tres cosas:
1. la reanudación añade al **mismo** fichero;
2. **no repite** consumo ya contado;
3. el acumulado del hilo continúa.

El criterio del 🔍, «un `session_meta` por fichero», no podía verla. **No hay ningún registro propio de «reanudación»**: el
`thread_settings_applied` también se escribe en otros momentos.

### (b) Versión mínima — ✅ Codex `0.153.0`

**LEÍDO EN LA FUENTE**, con un clon sin árboles de la historia y búsqueda por etiquetas estables:
- `TokenUsageRecord` no está en `0.152.1` *(2026-09-01)*;
- está en **`0.153.0`** *(2026-09-02)*, como variante de `RolloutItem` *(`history/src/lib.rs:111`)*, persistida *(`rollout/src/policy.rs:20`)*
  y escrita por `record_observed_response_completed` *(`core/src/session/mod.rs:4342`)*;
- los campos de `TokenUsageRecord` y de `TokenUsage` son idénticos en `0.153.0` y en `0.160.1`.

No se miraron las versiones preliminares *(`-alpha`)*.

### (c) `cache_write_input_tokens` va DENTRO de `input_tokens` — ✅

**LEÍDO EN LA FUENTE** *(documentación oficial, <https://developers.openai.com/api/docs/guides/prompt-caching>, consultada el
2026-10-07)*:
- «Cache-write pricing is not an additive fee: input tokens use the uncached-input, cached-input, or cache-write rate».
- Su ejemplo calcula `ordinaryInputTokens = inputTokens - cachedTokens - cacheWriteTokens`.

**LEÍDO EN LA FUENTE** *(Codex)*:
- el test `responses.rs:802-829` usa entrada 100 = 40 de caché + 60 de escritura, con total 110 = 100 + 10;
- Codex **no** resta la escritura en su `non_cached_input` *(`protocol.rs:2430-2432`)*.

**MEDIDO**: `cache_write_input_tokens` = 0 en los 7 registros. La copia **no puede distinguir** las dos lecturas: hace falta un fixture
sintético.

**Consecuencia para M-4**: `tokens_input` = entrada − caché − escritura.

### (d) Líneas largas — ✅ ninguna llega al tope

- **MEDIDO**: la línea máxima por fichero está en la tabla de arriba; la mayor, 44 745 B, en F3. Los `token_usage_record` ocupan
  861–867 B. Ninguna línea pasa de 1 MiB.
- **LEÍDO** *(agente)*:
  - `state.Recorrer` lee con `bufio.Reader.ReadBytes('\n')`, **sin tope** *(`internal/state/state.go:126,129`)*;
  - una última línea sin `\n` no se consume *(`:130-133`)*;
  - un fichero más corto que el offset se relee desde 0 *(`:113-115`)*;
  - `--scan` usa un `bufio.Scanner` con tope de **1 MiB** *(`cmd/permea/main.go:447-448`)*, y una línea mayor acaba en `sc.Err()`
    *(`:474`)*.
- **NO SE PUDO MEDIR**: el tamaño de un `compacted` de una sesión larga. Lleva `replacement_history`, y en F6 ocupa menos que su
  `session_meta`.

### (e) `session_ref` y `machine_ref`

**LEÍDO** *(agente)*:
- `SessionRef: event.Ref(ctx.Salt, r.SessionID)` y `MachineRef: event.Ref(ctx.Salt, ctx.MachineID)` *(`internal/ingest/claudecode.go:126-127`)*;
- `event.Ref` = SHA-256 de `sal:valor`, vacío si no hay valor *(`internal/event/event.go:40-46`)*;
- `MachineID` es un id aleatorio de la instalación, persistido *(`internal/config/identity.go:17-20`)*. No sale del log.

**En Codex**:
- `machine_ref`: igual que hoy;
- `session_ref`: candidato, el `session_id` del propio registro.
  - **MEDIDO**: igual a `session_meta.id` en los 7.
  - **LEÍDO**: es el del hilo raíz, y en una bifurcación los registros copiados conservan el del padre.

### (f) El contador independiente

No comparte código con el agente. Lee la copia **en su sitio** e imprime sólo recuentos, sumas y nombres de modelo. Las partidas son las
de M-4.

```sh
T="$(mktemp -d /tmp/permea-008-XXXXXX)"
cat > "$T/contador.py" <<'PY'
import json, os, sys
raiz = sys.argv[1]
fs = sorted(os.path.join(dp, n) for dp, _, ns in os.walk(raiz) for n in ns if n.endswith('.jsonl'))
for i, f in enumerate(fs, 1):
    regs = sin_id = tc = 0; s = [0, 0, 0, 0]; hilo = None; modelo_turno = {}; vigente = None; modelos = {}
    for raw in open(f, 'rb'):
        r = json.loads(raw); t = r.get('type'); p = r.get('payload') or {}
        st = p.get('type') if isinstance(p, dict) else None
        if t == 'turn_context':
            modelo_turno[p.get('turn_id')] = p.get('model'); vigente = p.get('model') or vigente
        elif st == 'thread_settings_applied':
            vigente = (p.get('thread_settings') or {}).get('model') or vigente
        elif st == 'token_count' and p.get('info') is not None:
            tc += 1
        elif t == 'token_usage_record':
            if not p.get('response_id'):
                sin_id += 1; continue
            regs += 1; u = p['usage']
            inp, cr, cw, out = (u.get(k) or 0 for k in ('input_tokens', 'cached_input_tokens', 'cache_write_input_tokens', 'output_tokens'))
            s = [a + b for a, b in zip(s, (inp - cr - cw, cw, cr, out))]
            m = modelo_turno.get(p.get('turn_id')) or vigente
            modelos[m] = modelos.get(m, 0) + 1; hilo = p['thread_token_usage']
    formato = 'actual' if regs or sin_id else ('anterior' if tc else 'sin consumo')
    codex = hilo['input_tokens'] - hilo['cached_input_tokens'] + hilo['output_tokens'] if hilo else '-'
    print(f"F{i} formato={formato} registros={regs} sin_response_id={sin_id} input={s[0]} cache_creation={s[1]}"
          f" cache_read={s[2]} output={s[3]} modelos={modelos} tokens_used_codex={codex}")
PY
(cd "$T" && python3 -I contador.py "<copia congelada>/sessions")
```

**MEDIDO** el 2026-10-07. Es la referencia de los SC:

| F | Formato | Registros | `tokens_input` | `tokens_cache_creation` | `tokens_cache_read` | `tokens_output` | Modelo | «tokens used» |
|---|---|---:|---:|---:|---:|---:|---|---:|
| F1–F4 | anterior | 0 | 0 | 0 | 0 | 0 | — | — |
| F5 | sin consumo | 0 | 0 | 0 | 0 | 0 | — | — |
| F6 | actual | 4 | 15 076 | 0 | 51 200 | 88 | gpt-6-luna ×4 *(3 por turno, 1 vigente)* | 15 164 |
| F7 | actual | 1 | 2 823 | 0 | 11 008 | 5 | gpt-6-luna | **2828** |
| F8 | actual | 2 | 3 880 | 0 | 24 064 | 237 | gpt-6-luna ×2 | **4117** |
| **Total** | | **7** | **21 779** | **0** | **86 272** | **330** | | |

0 registros sin `response_id`. En F7 y F8: `tokens_input` + `tokens_cache_creation` + `tokens_output` = «tokens used».

## Premisas que cayeron

- **«`token_count` acumulado» como fuente**: existe, pero no lleva identificador, pierde la compactación remota y en la 0.101.0 sale
  duplicado. La fuente fiel es `token_usage_record`.
- **«Dos sesiones `exec`»**: hay tres. F5 no tiene consumo.
- **El cambio de modelo y los errores 400** están en la misma sesión, F6, que es de `source` = `vscode`.
- **«0 reanudaciones»** *(el 🔍)*: F6 tiene una *(FASE 0 (a))*.

## Muestras que faltan

- **Una bifurcación**, para medir los registros copiados y sus `response_id`.
- **Un registro con `cache_write_input_tokens` > 0**, para fijar M-4 en una medida y no sólo en la documentación.
- **Una sesión larga con compactación**, para medir el tamaño de `compacted`.
- **Un cambio de modelo a la baja con compactación**, el caso en que el modelo vigente no sería el de la compactación.
- **Una sesión antigua reanudada con Codex ≥ 0.153.0**: un fichero con los dos formatos.
- **Un fichero `.zst`**.
- **Un `mtime` releído más tarde en NTFS**.
