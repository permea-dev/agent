# Contrato de transporte — Ingesta de eventos

Contrato entre el agente (`internal/transport`) y el backend de equipo. Cumple FR-009 (todo
envío autenticado y cifrado) y da soporte a FR-007 / SC-004 (exactamente una vez vía dedup).

## Endpoint

```
POST <config.endpoint>            # DEBE ser https:// — el cliente rechaza http://
Content-Type: application/json
Authorization: Bearer <device_token>
```

- **TLS obligatorio**: el cliente valida que el esquema del endpoint sea `https`. Un endpoint
  `http://` es un error de configuración y aborta el envío (nunca en claro).
- **Auth**: token de dispositivo por instalación en la cabecera `Authorization`.

## Cuerpo de la petición

Lote de eventos de frontera (ver [boundary-event.md](./boundary-event.md)):

```json
[
  { "schema_version": 1, "event_id": "…", "...": "…" },
  { "schema_version": 1, "event_id": "…", "...": "…" }
]
```

- El lote contiene 1..N eventos pendientes de `queue.jsonl`.
- Cada elemento cumple **exactamente** el esquema de la allowlist; nada fuera de ella.

## Respuestas y semántica

| Código | Significado | Acción del agente |
|---|---|---|
| `2xx` | Lote aceptado (o ya visto) | Eliminar esos eventos de `queue.jsonl` (reescritura atómica). |
| `401` / `403` | Token inválido | No reintentar en bucle; registrar y detener sync (config errónea). |
| `4xx` (otros) | Petición malformada | No reintentar el mismo lote indefinidamente; registrar. |
| `5xx` / error de red | Backend/enlace no disponible | Mantener en cola; reintentar con backoff exponencial acotado. |

## Deduplicación (exactamente una vez)

- La entrega es **at-least-once** desde el cliente: reintenta hasta recibir `2xx`.
- El backend **DEBE** deduplicar por `event_id`: un `event_id` ya almacenado se ignora y se
  responde `2xx`. Así, un reintento tras un `2xx` cuya respuesta se perdió **no** produce
  duplicado (SC-004: nº de eventos recibidos = nº de llamadas reales).
- ~~El `event_id` lo genera el agente con `crypto/rand` (`event.NewID`), estable por evento a lo
  largo de reintentos.~~ **Redescrito por P-006 FR-011 (2026-10-02):** el `event_id` es
  **determinista**. Se deriva por hash, sin sal ni ningún dato local, de los identificadores del
  mensaje del proveedor (`message.id`, `requestId`), así que el mismo mensaje da el mismo `event_id`
  en cualquier instalación y pasada. Sigue estable a lo largo de reintentos. La derivación, sus
  vectores de prueba y la regla «una pasada, un evento por mensaje» están en
  [`specs/006-medicion-fiel/contracts/event-id.md`](../../006-medicion-fiel/contracts/event-id.md).
  **La forma del campo no cambia**: cadena hexadecimal de 32 caracteres.

## Garantías de no-pérdida (offline)

- Un evento se **persiste en `queue.jsonl` antes** de intentar transmitirse; una caída del
  proceso o de la red no lo pierde (FR-007).
- Sin red, la cola crece y la medición local continúa; al volver la red, se drena en orden.

## Notas de prueba

- Los tests usan `httptest.NewTLSServer` (HTTPS) — nunca `http` en claro.
- Escenario SC-004: enviar un lote, simular pérdida de respuesta, reintentar el **mismo** lote
  y verificar que el backend simulado lo trata como idempotente por `event_id`.
