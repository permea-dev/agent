# Contrato — Derivación del `event_id` de Codex (P-008)

**Feature**: `008-lector-codex` | **Fecha**: 2026-10-07 | **Requisitos**: P-008 M-3, FR-006 a FR-009, FR-016

Este contrato fija **qué valor** tiene el `event_id` de un evento de Codex CLI y **cuándo se emite**. Es **hermano** del de 006
(`specs/006-medicion-fiel/contracts/event-id.md`), con el mismo esquema y un espacio de nombres propio. **No lo modifica**: Claude Code
sigue derivándose exactamente como allí. La forma del campo es la del contrato de frontera.

---

## Entrada

De cada registro `token_usage_record` de una sesión de Codex *(formato actual, Codex ≥ 0.153.0; `soporte/descubrimiento.md` §FASE 0 (b))*
se lee **un identificador técnico del proveedor**:

| Campo del registro | Nombre aquí | Forma observada *(2026-10-07, 7 registros)* |
|---|---|---|
| `payload.response_id` | `P` | identificador de respuesta de la Responses API, 55 caracteres |

Se usa **sólo** para derivar y deduplicar. **NUNCA** se copia a ningún campo del evento, ni entero ni en fragmento.

## Derivación

```
componentes = [ "permea/event_id/v1", "codex", "respuesta", P ]
codificado  = concatenación de  ( longitud_big_endian_uint32(c) ‖ bytes_utf8(c) )  para cada c
event_id    = hex_minúsculas( SHA-256(codificado)[0:16] )        → 32 caracteres [0-9a-f]
```

| Caso | Identificadores | Emite |
|---|---|---|
| `P` no vacío | `P` | sí |
| `P` ausente o vacío | — | **NO**; se cuenta como «sin identificador» |
| Formato anterior *(sin `token_usage_record`)* | — | **NO**; el fichero se cuenta como «formato anterior» *(D-2)* |

**Ningún otro dato entra en el hash**: ni sal, ni máquina, ni desarrollador, ni organización, ni sesión, ni turno, ni fichero, ni
contenido. Tampoco el modelo, el `timestamp` ni las partidas de consumo.

**El espacio de nombres separa las herramientas.** `"codex"` va en la posición en la que 006 pone `"claude_code"`, y `"respuesta"` en la
de `TIPO`. Así, el mismo valor llegado como identificador de Codex y como `message.id` de Claude Code da `event_id` distintos por
construcción.

## Garantías

- **Determinista**: el mismo registro da el mismo `event_id` en cualquier pasada, fichero e instalación.
- **Bifurcación**: Codex copia los registros del padre al fichero del hijo *(`core/src/session/mod.rs:1655-1660`, `rust-v0.160.1`)*.
  Una respuesta copiada da el **mismo** `event_id` que la original, y la plataforma la descarta por `(org_id, event_id)`.
- **Reanudación**: añade al mismo fichero sin repetir registros *(medido en F6)*. No cambia nada aquí.
- **Releer un fichero** *(truncado, rotado o estado perdido)*: vuelve a dar los mismos `event_id`. Lo enviado no se duplica en la
  plataforma.
- **El identificador del proveedor nunca viaja**: ni en el evento, ni en la cola, ni en nada que se transmita. **Excepción declarada**
  *(E-1)*: el nombre del fichero de sesión lleva el identificador del hilo, y su ruta es la clave de `state.json`; queda **en local** y no viaja.

## Vectores de prueba (normativos)

Identificadores **sintéticos**, sin la forma de ningún identificador real:

| Caso | Entrada | `event_id` esperado |
|---|---|---|
| respuesta | `P = r-000000000000000000000001` | `31439e3953a0916dde2f98b748ceb985` |
| otra respuesta | `P = r-000000000000000000000002` | `e2c798be5f23e44680ab2c70690f8cf9` |
| el mismo valor en el espacio de Claude Code | `["permea/event_id/v1", "claude_code", "solo_message_id", "r-…0001"]` | `128d67bd072f4488bc3854e82bbcaeb2` *(≠ el primero)* |

Calculados el 2026-10-07 con una implementación independiente *(Python `hashlib` + `struct`)*. Como control del instrumento, la misma
implementación reproduce el vector «sólo `message.id`» de 006 *(`1692269369e3bb2b418279566f4b093f`)*. La implementación en Go **DEBE**
reproducirlos byte a byte.

## Una pasada, un evento por respuesta

- **Cada `token_usage_record` es una respuesta terminada** *(M-2)*. No hay espera ni cierre, y la pasada de 007 no se aplica.
- **Dentro de una pasada** *(P-7 ✅)*, un `event_id` ya emitido no se vuelve a emitir: se cuenta como «repetida». Es el caso
  de una bifurcación leída junto a su padre.
- **Entre pasadas** no hay memoria. Un registro releído puede volver a encolarse con el **mismo** `event_id`, y la plataforma lo descarta.
- **Qué es una pasada**: lo mismo que en 006 *(una ejecución de `--run`, un ciclo de `--daemon` o una ejecución de `--scan`)*.

## Lo que este contrato NO cambia

- **El contrato de 006 y su implementación**: `eventid.go`, `eventid_test.go` y `event-id.md`, con 0 bytes de diff *(M-1)*. La derivación
  de Codex vive en un fichero y una función nuevos.
- **La forma del campo**: 32 hex.
- **La allowlist del evento**, `SchemaVersion = 1` e `internal/event`.
- **El dominio** `permea/event_id/v1`. Codex no estrena una `v2`: se distingue por el espacio de nombres, no por la versión.
