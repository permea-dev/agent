# Contrato — Derivación del `event_id` de Copilot CLI (P-011)

**Feature**: `011-lector-copilot` | **Fecha**: 2026-10-14 | **Estado**: **propuesto**, nacido de P-1 *(ratificada el 2026-10-14)* |
**Requisitos**: P-011 FR-010, FR-011, FR-014, FR-019

Este contrato fija **qué valor** tiene el `event_id` de un evento de Copilot CLI y **cuándo se emite**. Es **hermano** de los de 006, 008, 009 y 010
*(`event-id.md`, `event-id-codex.md`, `event-id-gemini.md`, `event-id-opencode.md`)*: el mismo esquema, con un espacio de nombres propio. **No modifica ninguno.**

## Entrada

Un **registro** es una línea de `events.jsonl` *(spec FR-010)* de una de estas dos clases. De cada una se lee **un identificador técnico del servidor**:

| Clase | Línea | Campo | Nombre aquí | Tipo |
|---|---|---|---|---|
| llamada | `type: "session.usage_record"` | `data.usage.accounting.usageId` | `U` | `respuesta` |
| compactación | `type: "session.compaction_complete"` con `data.success = true` | `data.requestId` | `R` | `compactacion` |

`U` es único por llamada *(9 de 9 en m2 y 4 de 4 en m3, también en los registros de un subagente)* e igual a `usage.apiCallId`. `R` es el de la petición
al servidor y coincide con el de su `model.model_call_success` *(m2 M2)*. Los dos se usan **sólo** para derivar el `event_id`: **NUNCA** se copian a
ningún campo del evento, a la cola ni a `state.json`, ni enteros ni en fragmento.

## Derivación

```
componentes = [ "permea/event_id/v1", "copilot_cli", T, I ]       con (T, I) = ("respuesta", U) o ("compactacion", R)
codificado  = concatenación de  ( longitud_big_endian_uint32(c) ‖ bytes_utf8(c) )  para cada c
event_id    = hex_minúsculas( SHA-256(codificado)[0:16] )        → 32 caracteres [0-9a-f]
```

| Caso | Emite |
|---|---|
| `usage_record` con `U` textual no vacío | sí *(salvo incoherente, FR-020)* |
| `compaction_complete` con `success = true` y `R` textual no vacío | sí *(salvo incoherente)* |
| `U` o `R` ausente, vacío o no textual | **NO**; se cuenta como «sin identificador» |
| `compaction_complete` con `success ≠ true` | **NO**; no es registro |
| el mismo `event_id` otra vez en la misma pasada | **NO**; se cuenta como «repetidas» |

**Ningún otro dato entra en el hash**: ni sal, ni máquina, ni desarrollador, ni organización, ni `sessionId`, `sourceSessionId` o `sequence`, ni la
ruta o el nombre de la carpeta, ni el `id` de la línea, ni el modelo, ni las horas, ni las partidas. `"copilot_cli"` va donde 006 pone `"claude_code"`
*(D-1)*. `"respuesta"` es `tipoRespuesta`, como en 008–010; `"compactacion"` es nuevo y separa un `R` de un `U` iguales por azar.

## Por qué ids del servidor *(P-1 (a))*

- **Únicos y estables**: el fichero es de sólo añadir *(desc. Q1; m2 M2)*, y los ids no cambian al reanudar *(desc. Q5)*.
- **Una copia de las líneas** *(una bifurcación: `.workspace-fork.lock`, `fork_count`)* conservaría `U` y `R`. Daría el mismo `event_id`, y la
  plataforma la descartaría. El `id` de la línea es un UUID local que una copia podría regenerar.
- **`(sourceSessionId, sequence)`** también es único, pero sólo porque el subagente tiene su propio `sourceSessionId` *(m3 S2)*, y no existe en la
  compactación.

## Garantías

- **Determinista**: el mismo registro da el mismo `event_id` en cualquier pasada, fichero e instalación.
- **Fichero truncado** *(spec P-2)*: se relee desde 0, y lo ya emitido sale con el **mismo** `event_id`. La plataforma lo descarta por el índice
  único `(org_id, event_id)` *(009, contrato §Garantías)*.
- **Nada del proveedor viaja** *(spec FR-019)*.

## Vectores de prueba (normativos)

Identificadores **sintéticos**, sin la forma de ningún identificador real:

| Caso | Entrada | `event_id` esperado |
|---|---|---|
| llamada | `("respuesta", "u-000000000000000000000001")` | `3cf3d5fa7697681ce1ad1bd6e5abbd4e` |
| otra llamada | `("respuesta", "u-000000000000000000000002")` | `e63ae9472c63b251ce53d8f0246735e4` |
| compactación | `("compactacion", "r-000000000000000000000001")` | `722a18b8a60a4ebc97f16925cbdbfe61` |
| el valor de la 1.ª como compactación | `("compactacion", "u-000000000000000000000001")` | `a76be27e9ed99dcd6427f38b47785dac` *(≠ la 1.ª)* |
| el valor de la 1.ª en el espacio de Codex | `["permea/event_id/v1", "codex", "respuesta", "u-…0001"]` | `aa69ad2965ddf47bfc427e5e1cc101aa` *(≠ la 1.ª)* |
| rastro de P-1 (b), **no normativo** | `[…, "copilot_cli", "respuesta", "s-000000000000000000000001", "1"]` | `5eec56db2fe3246a41741b9fa11769a8` |

Calculados el 2026-10-14 con una implementación independiente *(Python `hashlib` + `struct`)*, sin código Go del repo. Como control, la misma
implementación reproduce el vector «llamada» de 010 *(`e771a483…`)*. La implementación en Go **DEBE** reproducir los cinco primeros byte a byte.

## Lo que este contrato NO cambia

- Los contratos de 006, 008, 009 y 010, y su implementación *(`eventid.go`, `codex_eventid.go`, `gemini_eventid.go`, `opencode_eventid.go` y sus tests)*:
  0 bytes de diff.
- La forma del campo *(32 hex)*, la allowlist del evento, `SchemaVersion = 1` e `internal/event`.
- El dominio `permea/event_id/v1`.
