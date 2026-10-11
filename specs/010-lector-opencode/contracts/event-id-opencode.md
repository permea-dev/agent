# Contrato — Derivación del `event_id` de OpenCode (P-010)

**Feature**: `010-lector-opencode` | **Fecha**: 2026-10-11 | **Estado**: **propuesto**, nacido de P-4 *(ratificada el 2026-10-11)* |
**Requisitos**: P-010 FR-012 a FR-014, FR-020

Este contrato fija **qué valor** tiene el `event_id` de un evento de OpenCode y **cuándo se emite**. Es **hermano** de los de 006, 008 y 009
*(`event-id.md`, `event-id-codex.md`, `event-id-gemini.md`)*: el mismo esquema, con un espacio de nombres propio. **No modifica ninguno.**

## Entrada

Una **llamada** es una fila `assistant` de `session_message`, completada *(`$.time.completed` presente y no nulo; spec P-1)*, en una base de
OpenCode 2.x. De ella se lee **un identificador técnico del proveedor**:

| Columna | Nombre aquí | Forma observada *(2026-10-09; 14 filas, 14 valores)* |
|---|---|---|
| `session_message.id` | `O` | un prefijo de 4 caracteres y 26 alfanuméricos *(desc. D-5)* |

`O` se usa **sólo** para derivar el `event_id` y la huella de la frontera *(spec §P-3)*. **NUNCA** se copia a ningún campo del evento, ni
entero ni en fragmento.

## Derivación

```
componentes = [ "permea/event_id/v1", "opencode", "respuesta", O ]
codificado  = concatenación de  ( longitud_big_endian_uint32(c) ‖ bytes_utf8(c) )  para cada c
event_id    = hex_minúsculas( SHA-256(codificado)[0:16] )        → 32 caracteres [0-9a-f]
```

| Caso | Emite |
|---|---|
| `O` no vacío y textual, fila completada que la pasada clasifica *(spec §P-3)* | sí *(salvo incoherente, FR-021)* |
| `O` ausente, vacío o no textual | **NO**; se cuenta como «sin identificador» |
| fila en curso *(sin `time.completed`)* | **NO**; entra en la lista en curso |
| fila abandonada *(más de 24 h en curso)* | **NO**, nunca |

**Ningún otro dato entra en el hash**: ni sal, ni máquina, ni desarrollador, ni organización, ni `session_id`, ni la ruta de la base, ni el
rowid, ni el modelo, ni las horas, ni las partidas. `"opencode"` va donde 006 pone `"claude_code"`; el tipo es `"respuesta"`, como en 008 y 009.

## Por qué basta el `id` *(P-4 (a))*

- **Es único por llamada**: 14 filas y 14 `id` en la copia del 09-10, y `(session_id, seq)` también único *(desc. D-5)*.
- **Es estable**: la fila se reescribe **en su sitio** mientras se genera, con el mismo `id` *(desc. D-5)*.
- **El rowid no sirve de identidad**: `VACUUM` lo renumera *(spec §P-3)*. El `id`, no.
- **La misma llamada en dos bases** *(canales, P-8)* da el mismo `event_id`, y la plataforma la descarta.

## Garantías

- **Determinista**: la misma llamada da el mismo `event_id` en cualquier pasada, base e instalación.
- **Frontera reiniciada** *(spec §P-3)*: las ya emitidas se reencolan con el **mismo** `event_id`, y la plataforma las descarta por el índice
  único `(org_id, event_id)` *(009, contrato §Garantías)*.
- **Nada del proveedor viaja**: ni `id` ni `session_id` *(éste, sólo con sal en `session_ref`)*. En `state.json` sólo queda una **huella con sal**
  del `id` *(spec FR-020)*, en local.

## Vectores de prueba (normativos)

Identificadores **sintéticos**, sin la forma de ningún identificador real:

| Caso | Entrada | `event_id` esperado |
|---|---|---|
| llamada | `O = o-000000000000000000000001` | `e771a4833f173d08b2d3322627083201` |
| otra llamada | `O = o-000000000000000000000002` | `f7392add336fb033f74ac241cb4571a3` |
| el mismo valor en el espacio de Gemini | `["permea/event_id/v1", "gemini", "respuesta", "o-…0001"]` | `a53e1c53e69102bc0411f71d2e711a70` *(≠ el primero)* |
| rastro de P-4 (b), **no normativo** | `[…, "opencode", "respuesta", "s-000000000000000000000001", "o-…0001"]` | `dcfb5de539ceec04942d7465fbaf0431` |

Calculados el 2026-10-11 con una implementación independiente *(Python `hashlib` + `struct`)*, sin código Go del repo.
- Como control, la misma implementación reproduce el vector «respuesta» de 009 *(`027c36e1…`)* y el de Codex de 009 *(`07d1f3ea…`)*.
- La implementación en Go **DEBE** reproducir los tres primeros byte a byte.

## Lo que este contrato NO cambia

- Los contratos de 006, 008 y 009 y su implementación *(`eventid.go`, `codex_eventid.go`, `gemini_eventid.go` y sus tests)*: 0 bytes de diff.
- La forma del campo *(32 hex)*, la allowlist del evento, `SchemaVersion = 1` e `internal/event`.
- El dominio `permea/event_id/v1`.
