# Phase 0 — Research: Agente inicial (Claude Code)

Resuelve las decisiones de "cómo" del plan. Cada apartado sigue el formato
**Decisión / Justificación / Alternativas consideradas**. No hay ningún
`NEEDS CLARIFICATION` pendiente tras esta fase.

---

## R1. Localización de los logs de uso de Claude Code (por SO)

**Decisión**: Descubrir las transcripciones en `~/.claude/projects/**/*.jsonl`, resolviendo
`~` por SO:
- Linux/macOS: `$HOME/.claude/projects`.
- Windows nativo: `%USERPROFILE%\.claude\projects`.
- La ruta base es configurable (override en `config.json`) para instalaciones no estándar.

**Justificación**: Claude Code guarda un JSONL por sesión bajo `~/.claude/projects/<ruta-codificada>/`.
Se confirma con el propio entorno de esta sesión, cuyo directorio de proyecto es
`…\.claude\projects\--wsl-localhost-Ubuntu-home-bfgnet-dev-agente\…`. El fixture
(`internal/ingest/testdata/claude_code_sample.jsonl`) reproduce ese formato (una línea JSON
por registro, con `type`, `timestamp`, `sessionId`, `cwd`, `message.usage`). Resolver la
ruta por SO cumple el Principio III ("resolución de rutas por SO; nunca se hardcodean").

**Alternativas consideradas**:
- Hardcodear una ruta absoluta → rechazado: viola el Principio III y rompe en Windows.
- Leer un directorio de logs global del sistema → rechazado: Claude Code es per-usuario.

---

## R2. Identificación del registro facturable dentro del JSONL

**Decisión**: Un registro genera evento **solo si** `type == "assistant"` **y**
`message.model != ""`. Cualquier otro (`user`, mensajes de sistema, herramientas) se ignora
sin error. La decodificación usa un struct que **no incluye** `message.content` ni texto
alguno.

**Justificación**: Los contadores de tokens viven en el turno del asistente
(`message.usage`). El registro `user` contiene el prompt (denylist) y no es facturable. Ya
implementado en `ingest.FromClaudeCodeLine`, validado por `boundary_test.go` (2 eventos
esperados sobre un fixture de 3 líneas). Esto materializa el deny-by-default: lo que no se
decodifica, no puede filtrarse.

**Alternativas consideradas**:
- Filtrar por presencia de `usage` en cualquier `type` → rechazado: acopla la detección a un
  campo opcional en lugar del rol; menos legible.

---

## R3. Escaneo incremental idempotente (FR-006, SC-003)

**Decisión**: `internal/state` mantiene por fichero `{path, size, mod_time, offset}`
persistido en `state.json`. En cada pasada:
1. Enumerar `*.jsonl` bajo la raíz de Claude Code.
2. Para cada fichero, si `size < offset` (truncado/rotado) → reiniciar `offset = 0`.
3. Abrir, `Seek(offset)`, leer **líneas completas** nuevas, emitir eventos.
4. Avanzar `offset` **solo hasta el final de la última línea completa** consumida; una línea
   parcial (fichero aún escribiéndose) no se cuenta hasta estar completa.
5. Persistir `state.json` de forma atómica (escribir a temporal + `os.Rename`).

**Justificación**: El offset por bytes sobre append-only JSONL da idempotencia natural: una
llamada ya leída queda por debajo del offset y nunca se reprocesa (SC-003 = 0 eventos al
reprocesar). Detectar truncado por `size < offset` evita perder datos ante rotación. La
escritura atómica evita estado corrupto ante caída a mitad. Solo librería estándar.

**Alternativas consideradas**:
- Deduplicar por hash de línea en un set persistido → rechazado: crece sin límite y es más
  costoso que un offset; el offset ya garantiza unicidad sobre append-only.
- Recordar solo `mod_time` → rechazado: no distingue líneas nuevas dentro del mismo fichero.

---

## R4. Cola offline y entrega exactamente-una-vez (FR-007, SC-004)

**Decisión**: Separar **generación** de **transmisión**:
- Al procesar, cada evento se **append** a `queue.jsonl` (una línea por evento) *después* de
  avanzar el estado, de modo que el evento existe de forma durable antes de intentar enviarlo.
- Un paso de sync lee lotes de `queue.jsonl`, los envía por HTTPS y, ante `2xx`, los elimina
  de la cola reescribiéndola atómicamente (temporal + `os.Rename`) dejando solo los no
  confirmados.
- La **idempotencia de entrega** se apoya en `event_id`: el backend deduplica por `event_id`,
  así un reintento tras un envío que sí llegó pero cuya respuesta se perdió no cuenta doble.
- Reintentos con backoff exponencial acotado (**máximo 5 reintentos** con **delay máximo de
  5 minutos** por espera); agotados, el lote permanece en cola para el siguiente ciclo. Sin
  red, la cola simplemente crece y la medición local no se detiene.

**Justificación**: "Exactamente una vez" extremo a extremo se logra con **at-least-once** en
el cliente (reintentar hasta `2xx`) + **dedup por `event_id`** en el backend — patrón estándar
y auditable. La cola en disco cumple "sin pérdidas" ante corte de red o de proceso (FR-007).
`event.NewID()` ya provee la clave de dedup. Todo con stdlib.

**Alternativas consideradas**:
- Mantener los eventos solo en memoria hasta enviarlos → rechazado: una caída pierde datos
  (viola FR-007/SC-004).
- Confiar en exactamente-una-vez a nivel de transporte → rechazado: no existe sin dedup de
  aplicación; el `event_id` es la vía simple y verificable.

---

## R5. Coste de modelo desconocido — "no disponible" ≠ 0 (FR-002, edge case)

**Decisión**: `pricing.Cost` devuelve `(coste float64, known bool)`. Un modelo ausente de la
tabla → `(0, false)`. El evento debe distinguir "coste 0 real" de "coste no disponible": se
añade el booleano `cost_available` al `Event` (metadato derivado, no contenido) y el `Event`
transmite `cost_usd` solo cuando `cost_available == true`. Los tokens **siempre** se
contabilizan aunque el coste no esté disponible.

**Justificación**: El spec dice "un modelo desconocido se contabiliza en tokens, con coste
marcado como no disponible; no bloquea el resto". Devolver `0` a secas mezcla "gratis" con
"desconocido" y falsearía SC-001 (±1%). `cost_available` es un flag derivado permitido por la
frontera (no es contenido). **Cambio de allowlist**: es una **ampliación con metadato**, no
con contenido, por lo que respeta el Principio I (la frontera prohíbe *contenido*, no
metadato de calidad de la medición); se documenta en `contracts/boundary-event.md`.

**Alternativas consideradas**:
- Dejar `cost_usd = 0` para desconocidos → rechazado: indistinguible de coste real 0; rompe
  la agregación.
- Omitir el evento de modelos desconocidos → rechazado: perdería consumo real de tokens.

---

## R6. Generación y persistencia del `salt` (FR-005, Principio I)

**Decisión**: Al primer arranque, generar un `salt` aleatorio (`crypto/rand`, 32 bytes hex) y
persistirlo en el fichero `salt` del directorio de datos con permisos restrictivos
(`0600` en POSIX). El `salt` **nunca** se transmite ni se incluye en el `Event`. `event.Ref`
ya calcula `sha256(salt + ":" + valor)`.

**Justificación**: El Principio I exige que ruta de proyecto, sesión y máquina crucen "solo
como hash salado" y que "el `salt` reside en local y NUNCA se transmite". Un salt por
instalación hace irreversible el hash sin acceso a la máquina. Hoy el `salt` está hardcodeado
en `main.go` ("dry-run-salt"): el plan lo sustituye por el salt persistido.

**Alternativas consideradas**:
- Salt fijo compilado → rechazado: común a todas las instalaciones, permitiría diccionarios
  de rutas conocidas.
- Derivar el salt del `machine_id` → rechazado: el `machine_ref` se transmite; derivar el
  salt de algo transmitido debilita la irreversibilidad.

---

## R7. Identidad de máquina (`machine_ref`)

**Decisión**: Derivar un identificador de máquina estable y **no sensible** por SO
(p. ej. `os.Hostname()` combinado con un id de instalación aleatorio persistido junto al
salt) y transmitir **solo su hash salado** (`event.Ref(salt, machineID)`). Preferir el id de
instalación aleatorio como fuente principal para no depender de identificadores del SO que
puedan considerarse sensibles.

**Justificación**: El `machine_ref` debe ser estable entre sesiones (para agregación) pero
irreversible. Un UUID de instalación aleatorio persistido cumple ambas y evita leer
identificadores de hardware. Solo cruza su hash.

**Alternativas consideradas**:
- Leer machine-id del SO (`/etc/machine-id`, registro de Windows) → rechazado: divergente por
  SO, potencialmente sensible; el UUID de instalación es más simple y suficiente.

---

## R8. Directorio de datos y de configuración por SO (FR-010, Principio III)

**Decisión**: Resolver el directorio de estado del agente por SO con stdlib
(`os.UserConfigDir` / `os.UserHomeDir`):
- Linux: `$XDG_CONFIG_HOME/permea` o `$HOME/.config/permea`.
- macOS: `$HOME/Library/Application Support/permea`.
- Windows: `%AppData%\permea`.
Allí viven `config.json`, `state.json`, `queue.jsonl` y `salt`. El fichero de config es
legible y editable por el usuario (FR-010).

**Justificación**: `os.UserConfigDir` ya da la ruta correcta por SO sin dependencias. Separa
el estado del agente de los logs de Claude Code (que están en `~/.claude`). Cumple "config
local en fichero legible por el usuario".

**Alternativas consideradas**:
- Un único fichero junto al binario → rechazado: el binario puede estar en una ruta de solo
  lectura; rompe la portabilidad.

---

## R9. Transporte HTTPS autenticado (FR-009, Principio de transporte)

**Decisión**: El cliente **exige** esquema `https` en el endpoint (rechaza `http://` en
config), envía `Authorization: Bearer <device_token>` y comprueba el código de estado
(`2xx` = confirmado; resto = reintento). El `device_token` se genera/asigna por instalación y
vive en `config.json`.

**Justificación**: FR-009 y la constitución exigen "HTTPS y autenticado con token de
dispositivo por instalación; NUNCA en claro". El cliente actual pone la cabecera pero no
verifica el esquema ni el status; el plan lo endurece. TLS lo aporta `net/http` (stdlib) con
los CAs del sistema.

**Alternativas consideradas**:
- Permitir `http` para pruebas locales → rechazado en producción; para tests se usa
  `httptest.NewTLSServer`, no `http` en claro.

---

## Resumen de artefactos a producir en Fase 1

- `data-model.md` — `Event` (frontera), `FileState`/`Store` (estado), `QueueItem` (cola),
  `Config`, `Rate`/`Table` (pricing), `Context` (datos locales del agente).
- `contracts/boundary-event.md` — allowlist/denylist y esquema JSON del `Event` (incluye
  `cost_available` de R5).
- `contracts/transport.md` — contrato HTTP de ingesta (método, cabeceras, cuerpo, respuestas,
  semántica de dedup por `event_id`).
- `quickstart.md` — cómo arrancar, escanear en dry-run, ejecutar el golden test y validar los
  escenarios de aceptación.
