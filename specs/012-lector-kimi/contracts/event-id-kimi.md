# Contrato — Derivación del `event_id` de Kimi Code CLI, y la frontera con huella (P-012)

**Feature**: `012-lector-kimi` | **Fecha**: 2026-10-15 | **Estado**: **propuesto**, nacido de D-3, P-4 *(con E-1)* y P-5 *(ratificadas el 2026-10-15)* |
**Requisitos**: P-012 FR-004, FR-005, FR-008, FR-011, FR-014, FR-019

Este contrato fija **qué valor** tiene el `event_id` de un evento de Kimi Code y **cuándo se emite**, y la **huella de la cabeza** que guarda `state.json`. Es
**hermano** de los de 006 y 008–011: el mismo esquema, con un espacio de nombres propio. **No modifica ninguno.**

## Entrada

Un **registro** es una línea de `wire.jsonl` que **empieza** por `{"type":"usage.record"` *(spec P-10)*. Kimi no le da id *(desc. Q2; m1 1)*, así que se
identifica por su **posición**:

| Nombre | Qué es | De dónde |
|---|---|---|
| `S` | la sesión | el nombre de la carpeta de la sesión: `sessions/<wd>/<S>/agents/<A>/wire.jsonl` |
| `A` | el agente | el nombre de la carpeta del agente *(`main`, `agent-0`…)* |
| `N` | el ordinal | el número, **desde 1**, de las líneas completas del fichero que empiezan por ese prefijo, contadas **desde el principio** e **incluidos** los heredados y las corruptas, en decimal sin ceros a la izquierda |

`S` y `A` se usan **sólo** para derivar el `event_id`, y `S` además para `session_ref` *(con sal)*. **NUNCA** se copian a ningún campo del evento, a la cola ni a
`state.json`, ni enteros ni en fragmento *(FR-019)*.

## Derivación

```
componentes = [ "permea/event_id/v1", "kimi_code", "respuesta", S, A, N ]
codificado  = concatenación de  ( longitud_big_endian_uint32(c) ‖ bytes_utf8(c) )  para cada c
event_id    = hex_minúsculas( SHA-256(codificado)[0:16] )        → 32 caracteres [0-9a-f]
```

| Caso | Emite |
|---|---|
| registro **posterior** a la última línea `{"type":"forked"` del fichero *(o fichero sin `forked`)*, coherente y con alguna partida ≠ 0 | sí |
| registro **anterior** a la última `forked` | **NO**; «heredados» *(D-3, FR-014)* |
| las cuatro partidas a 0 | **NO**; «a cero» *(P-8)* |
| el mismo `event_id` otra vez en la misma pasada | **NO**; «repetidas» |

**Ningún otro dato entra en el hash**: ni sal, ni máquina, ni desarrollador, ni organización, ni `wd`, ni la ruta, ni el `agentId` de la línea, ni el modelo, ni
`time`, ni las partidas *(D-3)*. `"respuesta"` es `tipoRespuesta` *(`codex_eventid.go:29`)*.

## La frontera con huella *(P-4, E-1)*

- Por cada `wire.jsonl`, `state.json` guarda los cuatro campos de siempre y **`cabeza`** = `event.Ref(sal, L)` *(`internal/event/event.go:40-46`: SHA-256 de
  `sal + ":" + L`, 64 hex)*, donde **L** es la primera línea completa del fichero **sin su LF** *(ni un CR final)*. Sin línea completa, `cabeza` vacía.
- **No se abre** si el tamaño y el `mtime` *(en segundos)* son los guardados.
- **Se relee desde 0** si el fichero **encoge**, si la huella de su primera línea no es `cabeza`, o si `cabeza` falta con `Offset > 0` *(un estado anterior al campo)*.
  Lo releído da **los mismos** `(S, A, N)` y, por tanto, **los mismos `event_id`**: la plataforma los descarta por el índice único `(org_id, event_id)`.

## Por qué la posición *(D-3, P-5)*

- **Estable**: `N` es un **recuento**, no un número de línea. La reparación de Kimi **quita la cola** y la migración **añade una línea `metadata` en cabeza**
  *(desc. Q7)*; ninguna cambia el `N` de un registro ya escrito.
- **El fork**: la copia lleva otra `S` y daría **otro** `event_id`. Por eso lo anterior a la última `forked` **no se emite** *(m2 3, m2 5: 2 heredados, 0 dobles)*.
  **Sin horas ni partidas en el hash**: la regla de 011.

## Vectores de prueba (normativos)

| Caso | `(S, A, N)` | `event_id` esperado |
|---|---|---|
| 1.er registro | `("session_sintetica-0001", "main", "1")` | `e1ad45be19ab01a906be10cf22f4e75c` |
| 2.º registro | `("session_sintetica-0001", "main", "2")` | `c7f5518b2da765df74f784f4a5737d40` |
| subagente | `("session_sintetica-0001", "agent-0", "1")` | `70ee1df9071f3a49800d62483b50dcb3` *(≠ el 1.º)* |
| fork: su 1.er registro propio, tras 2 heredados | `("session_sintetica-0002", "main", "3")` | `17456f5b3d83b75590d3f65f21a9a267` |
| la frontera de los componentes | `("a", "bc", "1")` | `eccac1f146057198b16b839f837fc42b` |
| la frontera de los componentes | `("ab", "c", "1")` | `4adb17be1c8f18ab3a2b7b322915fffd` *(≠ el anterior)* |
| el 1.º en el espacio de Codex | `[…, "codex", "respuesta", "session_sintetica-0001", "main", "1"]` | `a37c3049e29731d108347b52ed3e0b5e` *(≠ el 1.º)* |
| un heredado del fork, **no se emite** *(no normativo)* | `("session_sintetica-0002", "main", "1")` | `0d96b672691678fe4b194d8de782383b` |

**`cabeza`** *(normativo)*: con `sal = "sal-sintetica"` y `L = {"type":"metadata","protocol_version":"1.5","created_at":1}` →
`f0ad969649b2af3e649c8af85d5fc2ef6757978024f08ca8926b49f21186c7f9`. Con `"created_at":2` → `572bb1d23c7da57096f084cc31a0b7d4c760530398ca41d5273695e6102ebabe`.

Identificadores **sintéticos**, sin la forma de ningún identificador real. Calculados el 2026-10-15 con una implementación independiente *(Python `hashlib` + `struct`)*, sin código Go del repo. Como control, la misma implementación
reproduce el vector «llamada» de 011 *(`3cf3d5fa…`)*. La implementación en Go **DEBE** reproducir los siete primeros y los dos de `cabeza`, byte a byte.

## Lo que este contrato NO cambia

- Los contratos de 006 y 008–011, y su implementación *(`eventid.go`, los `*_eventid.go` y sus tests)*: 0 bytes de diff.
- La forma del campo *(32 hex)*, la allowlist del evento, `SchemaVersion = 1`, `internal/event` y el dominio `permea/event_id/v1`. Las entradas de `state.json`
  de los otros lectores no llevan `cabeza`, y su JSON no cambia *(`omitempty`)*.
