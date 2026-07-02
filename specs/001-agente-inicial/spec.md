# Especificación: Agente inicial — ingesta de Claude Code y frontera de datos

**Feature Branch**: `001-agente-inicial`
**Created**: 2026-07-02
**Status**: Draft
**Input**: Un agente local que lee el uso de Claude Code registrado en la máquina del desarrollador, calcula el coste en local y transmite al backend de equipo únicamente metadato derivado —nunca contenido—, funcionando también sin conexión.

## Clarifications

- El alcance de esta especificación es **una sola herramienta: Claude Code**. Otras herramientas (Cursor, Copilot, aider, Codex) son especificaciones posteriores que se mapearán contra la misma frontera.
- El agente **observa**; no intercepta ni modifica el tráfico de Claude Code.
- Esta especificación describe qué datos cruzan la frontera y qué garantías se cumplen. El diseño técnico (estructura de código, formato de logs, transporte) corresponde al plan.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Sincronización de métricas tras una sesión (Priority: P1)

El desarrollador usa Claude Code con normalidad. El agente, en segundo plano, detecta el uso nuevo, calcula su coste y envía al backend de la organización solo metadato derivado.

**Acceptance Scenarios:**

1. **Given** Claude Code ha registrado N llamadas nuevas al modelo, **When** el agente procesa ese uso, **Then** se encola un evento de métrica por llamada (duradero en `queue.jsonl`), y ninguno contiene texto de prompt, respuesta, código ni rutas o nombres en claro. La transmisión al backend y su validación end-to-end se cubren en la Historia 2 (US2).
2. **Given** una llamada ya procesada anteriormente, **When** el agente vuelve a examinar el mismo origen, **Then** esa llamada NO se reenvía.

### User Story 2 — Funcionamiento sin conexión (Priority: P1)

La máquina está sin red. El agente sigue midiendo en local y sincroniza cuando la red vuelve, sin pérdidas ni duplicados.

**Acceptance Scenarios:**

1. **Given** no hay conexión con el backend, **When** el agente procesa uso nuevo, **Then** los eventos quedan pendientes en local y la medición no se detiene.
2. **Given** la conexión se restablece, **When** el agente reintenta, **Then** los eventos pendientes se transmiten exactamente una vez.

### User Story 3 — Contenido inesperado en el origen (Priority: P1, seguridad)

Una versión futura de Claude Code añade al registro un dato nuevo que contiene contenido (por ejemplo, parte de un prompt).

**Acceptance Scenarios:**

1. **Given** el origen contiene un dato no contemplado en la allowlist, **When** el agente genera el evento, **Then** ese dato NO aparece en lo transmitido.

### Edge Cases

- Un registro que no corresponde a una llamada facturable (p. ej. un mensaje del usuario) se ignora, no genera evento.
- Un modelo desconocido para la tabla de precios se contabiliza en tokens, con coste marcado como no disponible; no bloquea el resto.
- Un registro corrupto o incompleto se omite sin detener el procesamiento del resto.
- Ausencia total de datos (Claude Code nunca usado): el agente no falla, simplemente no transmite nada.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001:** El agente DEBE derivar, por cada llamada al modelo registrada por Claude Code, sus contadores de tokens y el modelo empleado.
- **FR-002:** El coste DEBE calcularse en local. El agente NUNCA depende de un servicio externo para conocer el coste.
- **FR-003:** El agente DEBE transmitir un evento por llamada que contenga únicamente los campos de la allowlist del Contrato de frontera.
- **FR-004:** El agente NUNCA DEBE transmitir ningún elemento de la denylist: prompts, respuestas, código, diffs, rutas de fichero o nombres de proyecto en claro, argumentos o resultados de herramientas, ni secretos.
- **FR-005:** Los identificadores sensibles (proyecto, sesión, máquina) DEBEN cruzar solo como referencia pseudónima. El material que permitiría revertirlos permanece en local y NUNCA se transmite.
- **FR-006:** El procesamiento DEBE ser incremental e idempotente: una llamada ya transmitida NUNCA se reenvía.
- **FR-007:** El agente DEBE funcionar sin conexión, conservar los eventos pendientes y transmitirlos sin pérdida ni duplicado al recuperarse la red.
- **FR-008:** Un dato nuevo o desconocido en el origen NUNCA puede aparecer en el evento transmitido (deny-by-default).
- **FR-009:** Toda transmisión DEBE ir autenticada y cifrada en tránsito.
- **FR-010:** El comportamiento del agente (destino, identidad de organización y desarrollador, modo de las referencias de proyecto) DEBE ser configurable por el usuario en local.

### Key Entities

- **Evento de métrica**: la unidad de información que cruza la frontera. Representa una única llamada al modelo, reducida a sus métricas y a referencias pseudónimas.
- **Contrato de frontera**: el conjunto cerrado de campos permitidos (allowlist) y la enumeración de lo prohibido (denylist). Es la materialización del Principio I de la constitución.

## Contrato de frontera de datos

### Cruza la frontera (allowlist)

| Dato | Descripción |
|---|---|
| Versión de contrato y de agente | Trazabilidad del formato |
| Identificador de evento | Permite deduplicar |
| Marca de tiempo de la llamada | Cuándo ocurrió |
| Herramienta | Origen (aquí: Claude Code) |
| Modelo | Modelo empleado |
| Tokens (entrada, salida, escritura de caché, lectura de caché) | Métricas de consumo |
| Coste | Calculado en local |
| Referencia de proyecto | Hash pseudónimo, nunca el nombre en claro (salvo opt-in explícito) |
| Referencia de sesión | Hash pseudónimo |
| Referencia de máquina | Hash pseudónimo |
| Identidad de desarrollador y de organización | Identidad, no contenido |

### NUNCA cruza la frontera (denylist)

| Categoría | Ejemplos |
|---|---|
| Texto de mensajes | Prompts del usuario y respuestas del asistente |
| Código | Contenido de ficheros, diffs, fragmentos |
| Identificadores en claro | Rutas de fichero, nombres de proyecto, nombres de rama |
| Llamadas a herramientas | Argumentos y resultados |
| Secretos | Variables de entorno, claves de API |
| Cualquier dato del origen no presente en la allowlist | Se descarta por defecto |

## Success Criteria *(mandatory)*

- **SC-001:** El coste agregado por el agente coincide (±1%) con el de una herramienta de referencia sobre los mismos datos.
- **SC-002:** Inspeccionando lo que el agente transmite, no aparece ningún elemento de la denylist.
- **SC-003:** Reprocesar datos ya tratados produce cero eventos nuevos.
- **SC-004:** Tras una interrupción y posterior recuperación de red, el número de eventos recibidos por el backend es exactamente igual al de llamadas reales.
- **SC-005:** Una entrada con contenido sensible inyectado a propósito no produce ningún contenido en lo transmitido.
- **SC-006:** Un revisor externo confirma el cumplimiento de la frontera examinando únicamente el punto donde se construye el evento.
- **SC-007:** `make build` produce el binario estático `./bin/permea` y este arranca en el SO de desarrollo sin dependencias previas. Nota: el instalador multiplataforma de un solo comando (Homebrew / Scoop / paquetes por SO) es objeto de una **especificación de distribución posterior**; esta feature acota SC-007 a lo verificable localmente (build + arranque).
