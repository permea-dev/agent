# Contrato — Derivación del `event_id` (P-006)

**Feature**: `006-medicion-fiel` | **Fecha**: 2026-10-02 | **Requisitos**: P-006 FR-001 a FR-008, FR-033

Este contrato fija **qué valor** tiene el `event_id` de un evento de Claude Code y **cuándo se
emite**. **No** redefine la forma del campo, que sigue siendo la del contrato de frontera
(`specs/001-agente-inicial/contracts/boundary-event.md:25`: cadena hexadecimal, clave de
deduplicación). Lo cumple y precisa su origen.

Sustituye la frase de origen de `specs/001-agente-inicial/contracts/transport.md:47-48` («lo genera el
agente con `crypto/rand`»), que 006 redescribe (P-006 FR-011).

---

## Entrada

De cada línea facturable del log (`type = "assistant"`, `message.model` no vacío y distinto de
`<synthetic>`) se leen **dos identificadores técnicos del proveedor**:

| Campo del log | Nombre aquí | Forma observada (2026-10-02, `research.md` R1.2) |
|---|---|---|
| `message.id` | `M` | `msg_` + 24 alfanuméricos |
| `requestId` (nivel superior) | `R` | `req_` + 24 alfanuméricos |

Se usan **sólo** para derivar y deduplicar. **NUNCA** se copian a ningún campo del evento, ni enteros
ni en fragmento (P-006 FR-003, FR-013).

## Derivación

```
componentes = [ "permea/event_id/v1", "claude_code", TIPO, …identificadores ]
codificado  = concatenación de  ( longitud_big_endian_uint32(c) ‖ bytes_utf8(c) )  para cada c
event_id    = hex_minúsculas( SHA-256(codificado)[0:16] )        → 32 caracteres [0-9a-f]
```

| Caso | `TIPO` | Identificadores | Emite |
|---|---|---|---|
| Están los dos | `par` | `M`, `R` | sí |
| Sólo `message.id` | `solo_message_id` | `M` | sí |
| Sólo `requestId` | `solo_request_id` | `R` | sí |
| Ninguno | — | — | **NO**; se cuenta como «sin identificador» (FR-006) |
| `model = <synthetic>` | — | — | **NO**; se cuenta como «sintética» (FR-007) |

**Ningún otro dato entra en el hash**: ni sal, ni máquina, ni desarrollador, ni organización, ni
contenido (FR-002). Por eso el mismo mensaje da el mismo `event_id` en cualquier instalación.

**Versión del dominio.** `v1` sólo cambia si cambia la derivación. Una `v2` produciría valores
disjuntos de los de `v1` por construcción, y su adopción se especificaría aparte (como hizo `pmea2` con
`pmea1`).

## Vectores de prueba (normativos)

Identificadores **sintéticos**, no observados en ningún log real:

| Caso | Entrada | `event_id` esperado |
|---|---|---|
| par | `M = msg_000000000000000000000001`, `R = req_000000000000000000000001` | `43b8b3b6446704ae3cb8bb74683bf3e2` |
| par, otro `R` | `M` igual, `R = req_000000000000000000000002` | `e68d468337bad98f114449bbe5a37f7b` |
| sólo `message.id` | `M = msg_000000000000000000000001` | `1692269369e3bb2b418279566f4b093f` |
| sólo `requestId` | `R = req_000000000000000000000001` | `5ab5f8575791456e5d951eb35bf36370` |
| ambigüedad sin prefijo de longitud | par `("a","bc")` frente a par `("ab","c")` | `6e7e350c923581ba78d5c81e0897ed53` ≠ `918d6e4cd99d09b918802196669fe50e` |

Calculados el 2026-10-02 con una implementación independiente (Python `hashlib`). La implementación en
Go **DEBE** reproducirlos byte a byte; es el testigo que fija la derivación (`research.md` R2).

## Una pasada, un evento por mensaje (FR-001, FR-005, FR-033)

- **Dentro de una pasada**, la primera línea de un `event_id` produce el evento. Las siguientes con el
  mismo `event_id` **no se emiten**. Si su `usage` (las cuatro partidas) difiere del de la primera, se
  cuentan como «con consumo distinto». **Las líneas NUNCA se suman.**
- **Entre pasadas** no hay memoria. Un mensaje que reaparece puede volver a encolarse con el **mismo**
  `event_id`, y la plataforma lo descarta por `(org_id, event_id)` (`ON CONFLICT DO NOTHING`).
- **Qué es una pasada**: una ejecución de `--run`; un ciclo de `--daemon`; una ejecución de `--scan`
  sobre un fichero.

## Lo que este contrato NO cambia

- La **forma** del campo: 32 hex, como hoy (P-006 FR-004, D-006-8).
- La **allowlist** del evento, `SchemaVersion = 1` e `internal/event` (P-006 FR-012, D-006-3).
- Los eventos **ya encolados** antes de actualizar: salen con su `event_id` aleatorio original
  (P-006 FR-009).
