# 010 · Descubrimiento del lector de OpenCode

**Fecha**: 2026-10-09 · **Fuentes**: la **copia congelada** de una sesión de OpenCode **2.0.26** *(Windows, proveedor `google`, Gemini 3.8 Flash con
clave de pago)* y el binario del paquete instalado.

**Cómo se lee esto.** Cada afirmación lleva una etiqueta:
- **MEDIDO**: en la copia, con `sqlite3` abriéndola como `file:<copia>?mode=ro&immutable=1`, y las cuentas con `decimal` de Python, fuera del repo;
- **LEÍDO EN EL BINARIO**: cadenas del paquete npm `@opencode/cli` **2.0.26** *(`bin/opencode.exe`, un ejecutable de Bun que lleva el JS dentro)*,
  leídas con `grep` sin ejecutarlo;
- **NO SE PUDO MEDIR**: y qué muestra haría falta.

> **Privacidad.** De la copia salen sólo recuentos, sumas, fechas, nombres de modelo y nombres de campo. Las rutas son genéricas
> (`~/.local/share/opencode`, `<copia congelada>`). Los identificadores *(`ses_…`, `msg_…`)* se compararon entre sí sin imprimirse. De las tablas de
> cuentas y credenciales *(`account`, `account_state`, `control_account`, `credential`, `kv`, `permission`)* sólo se leyó el esquema.

## La copia congelada *(MEDIDO)*

- **Una base consolidada**, de 6 348 800 B, sin `-wal` pendiente; sha256 comprobado al empezar y al terminar.
- **El log**, `log/opencode.log`, de 65 líneas *(63 `INFO`, 1 `WARN`, 1 `ERROR`)*.
- **La sesión**: una sola, del 2026-10-09, de **16:03:44 a 16:09:20 UTC**; 3 mensajes del usuario y herramientas de lectura, de shell y de edición. El
  título lo puso OpenCode.

## D-1 · Esquema *(MEDIDO)*

- **20 tablas**:
  - de sesión: `session_v2`, `session_message`, `session_pending`, `session_inbox`;
  - de eventos: `event`, `event_sequence`;
  - de proyecto: `project`, `project_directory`, `worktree`, `workspace`;
  - de instrucciones: `instruction_blob`, `instruction_entry`, `instruction_state`;
  - de cuentas: `account`, `account_state`, `control_account`, `credential`;
  - y `kv`, `permission`, `migration`.
- **`migration`**: **49** versiones aplicadas, de `20260127222353_familiar_lady_ursula` a `20261007190000_azure_cli_external_credential`. Entre ellas,
  `20260510033149_session_usage`, `20260323234822_events`, `20260604172448_event_sourced_session_input` y `20260910120000_clear_v1_session_permission`.
- **`session_message`**: `id`, `session_id`, `type`, `seq`, `time_created`, `time_updated`, `data` *(JSON)*. Índice único `(session_id, seq)`.
- **`session_v2`**: `id`, `project_id`, `parent_id`, `fork_session_id`, `directory`, `path`, `title`, `version`, `agent`, `model` *(JSON)* y los
  **totales de la sesión**: `cost`, `tokens_input`, `tokens_output`, `tokens_reasoning`, `tokens_cache_read`, `tokens_cache_write`. Más tiempos
  *(`time_created`, `time_updated`, `time_idle`, `time_compacting`, `time_archived`…)*.
- **`event`**: `id`, `aggregate_id`, `seq`, `created`, `type`, `data`. **`event_sequence`**: `aggregate_id`, `seq`, `owner_id`.

## D-2 · Dónde están los tokens y el coste *(MEDIDO)*

Filas de `session_message` por `type`: **`assistant` 8 · `user` 3 · `idle` 3**.

**Sólo las `assistant` traen consumo**, todas con las mismas rutas en `data`:

| Partida | Ruta JSON | Tipo |
|---|---|---|
| entrada **sin caché** | `$.tokens.input` | entero |
| salida | `$.tokens.output` | entero |
| razonamiento | `$.tokens.reasoning` | entero |
| lectura de caché | `$.tokens.cache.read` | entero |
| escritura de caché | `$.tokens.cache.write` | entero |
| coste *(USD)* | `$.cost` | real |
| modelo | `$.model.providerID`, `$.model.id`, `$.model.variant` | texto |
| fin | `$.finish` *(`tool-calls`, `stop`)*, `$.rawFinish` *(`STOP`)* | texto |
| tiempos *(ms epoch)* | `$.time.created`, `$.time.streamed`, `$.time.completed` | entero |

`user` lleva `$.text` y `$.time.created`; `idle` lleva `$.outcome` *(`succeeded` las 3)* y `$.time.created`. Las herramientas van dentro de la
`assistant` que las pide *(`$.content[].type`, `.name`, `.state.status`…)*: **no son filas aparte**.

## D-3 · Cuadre interno *(MEDIDO)*

| | entrada | salida | razonamiento | lectura | escritura | coste |
|---|---:|---:|---:|---:|---:|---:|
| Σ de las 8 `assistant` | 59 321 | 230 | 462 | 3 999 | 0 | 0,047385675 |
| `session_v2` | **59 859** | **235** | 462 | 3 999 | 0 | **0,047559575** |
| diferencia | **+538** | **+5** | 0 | 0 | 0 | **+0,000173900** |

- **Llamadas: 8**, las 8 `assistant`; `opencode stats` dice «8 steps». Los «3 prompts» son las 3 filas `user`.
- **Los «64k» de `stats`**:
  - las cinco partidas de `session_v2` suman **64 555**;
  - las de los mensajes, **64 012**;
  - las dos son «64k» si `stats` trunca.
  - **NO SE PUDO MEDIR** la definición exacta sin ejecutar `stats`. Las dos sumas incluyen la lectura de caché.
- **La diferencia** es una llamada que **no está en `session_message`**: D-8.
- **La entrada de OpenCode ya viene SIN la caché**: en la única llamada con lectura de caché *(la 5.ª: 3 963 de entrada y 3 999 de lectura)*, el coste
  sólo cuadra si `input` es lo no cacheado *(D-4)*.

## D-4 · El coste de OpenCode frente a la tarifa de Permea *(MEDIDO)*

**La fórmula de OpenCode, en las 8 llamadas, sin una excepción**: `input × 0,75 + (output + reasoning) × 3,75 + cache.read × 0,075` *(por millón)*. **Es la de
la fila `gemini-3.8-flash` de Permea**, con el razonamiento a precio de salida.

| # | entrada | salida + razon. | lectura | escritura | coste de OpenCode | Permea, `ROUND(·, 6)` | Δ |
|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 7 495 | 66 | 0 | 0 | 0,00586875 | 0,005869 | +0,00000025 |
| 2 | 7 599 | 73 | 0 | 0 | 0,005973 | 0,005973 | 0 |
| 3 | 7 686 | 149 | 0 | 0 | 0,00632325 | 0,006323 | −0,00000025 |
| 4 | 7 863 | 61 | 0 | 0 | 0,006126 | 0,006126 | 0 |
| 5 | 3 963 | 181 | 3 999 | 0 | 0,003950925 | 0,003951 | +0,000000075 |
| 6 | 8 168 | 24 | 0 | 0 | 0,006216 | 0,006216 | 0 |
| 7 | 8 244 | 49 | 0 | 0 | 0,00636675 | 0,006367 | +0,00000025 |
| 8 | 8 303 | 89 | 0 | 0 | 0,006561 | 0,006561 | 0 |
| **Σ** | | | | | **0,047385675** | **0,047386** | **+0,000000325** |

- **La única diferencia es el redondeo por evento** de la plataforma.
- **Sin el razonamiento en la salida**, Permea daría **0,045652**: un 3,7 % por debajo. ⇒ **salida = `output` + `reasoning`**, como `output + thoughts` en
  el lector de Gemini CLI.
- **De dónde saca OpenCode el precio** *(LEÍDO EN EL BINARIO)*:
  - de **`models.dev`** *(6 apariciones)*;
  - las claves de la tarifa son `cost.input`, `cost.output`, `cost.cache_read` y `cost.cache_write`.
  - El log **no** registra ni el catálogo ni los precios.

## D-5 · Identidad y repetición *(MEDIDO)*

- **Un id estable por llamada**: el `id` de la fila *(`msg_` + 26 caracteres, 30 en total; 14 filas y 14 ids distintos)*. `(session_id, seq)` también es único.
- **Las filas se REESCRIBEN en su sitio**: las 8 `assistant` tienen `time_updated > time_created`, hasta **55,4 s** después. La misma llamada es **una sola
  fila** que se actualiza mientras se genera, no varias. Las `user` *(2 de 3, 1 ms)* también; las `idle`, no.
- **`event` está VACÍA** *(0 filas)*; `event_sequence` tiene 1. **Hoy no hay riesgo de contar doble** desde `event`. ⚠️ Las migraciones `events` y
  `event_sourced_session_input` dicen que el diseño es por eventos: **otra versión podría llenarla**.

## D-6 · Modelo *(MEDIDO)*

`$.model.providerID` = **`google`** · `$.model.id` = **`gemini-3.8-flash`** · `$.model.variant` = **`default`**, en las 8. En `session_v2.model`, el JSON
`{"id": …, "providerID": …}`. **`model.id` es ya la clave de la fila de Permea**, sin transformar.

## D-7 · Proyecto y sesión *(MEDIDO; sólo nombres de campo)*

- **Directorio del proyecto**: `session_v2.directory` *(presente; **igual** a `project.worktree` en la copia)*. `session_v2.path`, también presente.
  `project_directory`: 0 filas.
- **Sesión**: `session_v2.id` *(`ses_` + 26)*. `session_message.session_id` la enlaza. `parent_id` y `fork_session_id`, nulos *(no hay subagentes ni bifurcaciones)*.
- `project.id` *(40 caracteres)*, `project.vcs` vacío y `project.name` nulo.

## D-8 · Llamadas que no quedan en `session_message`

- **El título** *(`session_v2.title`, presente)*. Su llamada no está en los mensajes, pero **sí en los totales de la sesión**: es la diferencia de D-3.
- **Su modelo, por la cuenta** *(MEDIDO, no registrado)*:

  | La diferencia, 538 de entrada y 5 de salida, a… | Coste |
  |---|---:|
  | `gemini-3.8-flash` *(0,75 / 3,75)* | 0,00042225 |
  | **`gemini-3.5-flash-lite`** *(0,30 / 2,50)* | **0,00017390** |
  | `gemini-3.1-flash-lite` *(0,25 / 1,50)* | 0,000142 |

  Sólo la de `gemini-3.5-flash-lite` da **exactamente** la diferencia de coste, **0,000173900**. ⇒ El título lo hizo un modelo **más barato** que el de la
  sesión. Coincide con el **`small_model`** que el binario declara *(5 apariciones, LEÍDO)*. ⚠️ **El modelo no queda escrito en ningún sitio de la copia**:
  es una inferencia por aritmética.
- **El log**:
  - `role=cli` 31, `role=server` 25, `component=plugin` 10;
  - `event.type`: `model.updated` 3, `provider.updated` 3, `agent.updated` 1, `command.updated` 1;
  - **ningún nombre de modelo, ni título, ni resumen o compactación**.
- **Compactación y resúmenes**: no hubo *(`time_compacting` nulo, `session_pending` vacía)*. **NO SE PUDO MEDIR** si dejan fila propia: haría falta una
  sesión con compactación.

## D-9 · Ubicación y variantes *(LEÍDO EN EL BINARIO)*

- **Carpeta de datos**: `process.env.XDG_DATA_HOME || join(homedir(), ".local", "share")`. Es el **mismo código en Windows, Linux y macOS**, sin
  `AppData` ni `Library` para esto. ⇒ **`<XDG_DATA_HOME | ~/.local/share>/opencode/`**. El propio binario cita `~/.local/share/opencode/log/opencode.log`.
- **Nombre de la base**:
  - `OPENCODE_DB` manda si existe *(`":memory:"` = sin fichero)*;
  - si el canal está en `["latest","dev","beta","next","prod"]`, o `OPENCODE_DISABLE_CHANNEL_DB` es `"1"` o `"true"` → **`opencode.db`**;
  - si no → **`opencode-<canal>.db`**, con el canal saneado a `[a-zA-Z0-9._-]`.
- Hay **además** una `opencode-next.db` *(`nextDatabasePath`; mensajes «Incompatible opencode-next.db… missing required columns» y «Reassigned previous V2
  session»)*: **otra base, de propósito sin aclarar** *(pregunta abierta)*.
- Otras variables: `OPENCODE_TUI_CHANNEL`, `OPENCODE_CONFIG_DIR`, `XDG_CONFIG_HOME`, `XDG_CACHE_HOME`, `XDG_STATE_HOME`.
- **Formato 1.x**: el binario conserva `CREATE TABLE message` y `CREATE TABLE part` *(2 cada una)* y la cadena `"storage"`. Las migraciones de la
  copia incluyen `normalize_storage_paths`, `data_migration_state` y `clear_v1_session_permission`. **NO SE PUDO MEDIR** con una base 1.x: no hay muestra.
- **El paquete se llama `@opencode/cli`**, no `opencode-ai`. Tiene **212 567 080 B**.

## D-10 · Leer SQLite desde Go *(hechos; ninguna dependencia añadida)*

**Hoy**: `go.mod` es `module github.com/permea-dev/agent`, `go 1.22`, **sin un solo `require`**, y no hay `go.sum`. La publicación compila con
**`CGO_ENABLED=0`** *(«binario estático, sin dependencias de runtime», FR-003)*. La matriz es **5 plataformas**: `darwin`, `linux` × `amd64`, `arm64`, y
`windows/amd64` *(`windows/arm64`, excluida)*.

| Opción | ¿En `go.mod`? | ¿Con `CGO_ENABLED=0` y las 5? | ¿Lectura de una base con WAL mientras OpenCode escribe? |
|---|---|---|---|
| **Go puro** *(`modernc.org/sqlite`, C traducido a Go)* | no | sí, sin cgo, en las cinco *(documentación del módulo; no compilado aquí)* | es el motor de SQLite: con `mode=ro` necesita el `-shm`, y con `immutable=1` **no ve el WAL** *(documentación de SQLite)* |
| **WASM** *(`github.com/ncruces/go-sqlite3`, sobre `wazero`)* | no | sí, sin cgo *(documentación)*; ⚠️ el soporte de WAL depende del SO y la versión | ídem, con las limitaciones de memoria compartida del port |
| **cgo** *(`github.com/mattn/go-sqlite3`)* | no | **no**: exige cgo, y choca con FR-003 | — |
| **El `sqlite3` del sistema** | — | sin dependencia de Go, pero **sí de runtime**: Windows no lo trae, y en Linux depende de la distribución; choca con FR-003 | ídem |

**NO SE PUDO MEDIR** ninguna opción contra una base viva con WAL: no se añadió dependencia ni se ejecutó OpenCode. La copia está consolidada, sin `-wal`.

## Contraste con la interfaz *(MEDIDO)*

Por `seq`, el mensaje 1 lleva las llamadas 1–2; el 2, las 3–7; el 3, la 8.
- **Tras el 2.º mensaje**, el coste acumulado es **0,040825** *(0,040999 con el título)* → los «$0.04» de la barra.
- **Los «8.3K»** son la entrada más la salida de la última llamada de ese mensaje *(8 244 + 43)*.
- **Los «1 %»**, la ventana del modelo; no medido.

## Propuestas de decisión *(sin decidir)*

- **P-1 · Fuente**: `session_message` con `type = 'assistant'`, **una fila = un evento**. **Sólo filas terminadas** *(`$.time.completed` presente)*,
  porque se reescriben mientras se generan *(D-5)*.
- **P-2 · Partidas**:
  - entrada = `tokens.input` *(ya sin caché)*;
  - salida = `tokens.output + tokens.reasoning` *(D-4)*;
  - lectura = `tokens.cache.read`;
  - escritura = `tokens.cache.write`.
- **P-3 · Coste**: ¿**sin coste** *(`cost_available = false`; repara la plataforma, como Codex y Gemini)*, o el `$.cost` de OpenCode como observado?
  Hoy coincide con Permea al céntimo, pero sale de `models.dev`, que el agente no controla.
- **P-4 · `event_id`**: del `id` de la fila, con espacio de nombres propio *(`["permea/event_id/v1","opencode", …, id]`)*. El id no viaja.
- **P-5 · Modelo**: `model.id` tal cual; `providerID` y `variant` no viajan.
- **P-6 · Proyecto y sesión**: `project_ref` desde `session_v2.directory`; `session_ref` desde `session_v2.id`, con el mismo pseudónimo que Codex y Gemini.
- **P-7 · La llamada del título** *(y cualquier otra fuera de los mensajes)*:
  - **(a)** un evento «residuo» por sesión, `session_v2` menos la Σ de los mensajes, con el modelo **desconocido**;
  - **(b)** límite declarado: queda por debajo de la factura, aquí un **0,37 %** del coste.
- **P-8 · Ubicación**: `OPENCODE_DB`, o `<XDG_DATA_HOME | ~/.local/share>/opencode/opencode.db` y `opencode-<canal>.db`. **¿Y `opencode-next.db`?**
- **P-9 · 1.x**: tablas `message`/`part` o JSON en `storage/` = «formato anterior», no soportado *(como el `.json` de Gemini CLI)*.
- **P-10 · SQLite**: Go puro frente a WASM *(D-10)*, o, sin añadir dependencia, **copiar** la base *(y su `-wal`)* a un temporal y abrirla `immutable`.

## Preguntas abiertas

1. ¿Qué es `opencode-next.db`, y cuándo existe? *(P-8)*
2. ¿Una compactación o un resumen dejan fila en `session_message`, o sólo suben los totales? *(D-8; hace falta una muestra)*
3. ¿Qué cuenta exactamente «64k» en `stats`? *(D-3)*
4. ¿Existen todavía instalaciones 1.x? *(P-9)*
