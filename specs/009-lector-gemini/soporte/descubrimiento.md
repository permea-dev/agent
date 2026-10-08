# 009 · Descubrimiento del lector de Gemini CLI

**Fecha**: 2026-10-08 · **Fuentes**: la **copia congelada** de una sesión de Gemini CLI 0.63.0 y el código de esa versión.

**Cómo se lee esto.** Cada afirmación lleva una etiqueta:
- **MEDIDO**: en la copia, con un contador independiente (Python, fuera del repo);
- **LEÍDO EN LA FUENTE**: el paquete npm `@google/gemini-cli` **0.63.0** instalado. Viene empaquetado con esbuild, que conserva
  la ruta de origen de cada módulo: se cita esa ruta bajo `packages/` y la función. Todo el comportamiento de la 0.63 sale de ahí.
  Del repo público (`gh api`, sólo para fechar cambios): `core/src/services/chatRecordingService.ts` en `v0.30.0`–`v0.40.0`
  (las once), `v0.50.0` y `v0.63.0`, para fechar el paso a `.jsonl` (Q10; `v0.38.0` entera, por la forma antigua de `tokens`);
  `core/src/config/projectRegistry.ts` en `v0.25.0`, `v0.28.0`, `v0.29.0` y `v0.30.0`, para fechar los slugs (Q1); y la
  publicación `v0.39.0`, por su fecha;
- **NO SE PUDO MEDIR**: y qué muestra haría falta.

> **Privacidad.** De la copia salen sólo recuentos, sumas, fechas, nombres de modelo y nombres de campo. Las rutas son
> genéricas (`~/.gemini`, `<copia congelada>`, `<slug>`). Los identificadores se compararon entre sí sin imprimirse.

## La copia congelada *(MEDIDO)*

3 ficheros, que se leen **en su sitio**: `tmp/<slug>/.project_root` (32 B), `tmp/<slug>/logs.json` (2 528 B) y
`tmp/<slug>/chats/session-AAAA-MM-DDTHH-MM-<8 hex>.jsonl` (29 394 B, 53 líneas, todas terminadas en `\n`, 0 corruptas). Hay
además una carpeta `tmp/<slug>/logs/`, vacía. La sesión es del 2026-10-08 y dura 21 min: dos ejecuciones (la segunda con
`--resume latest`), una compresión manual y un cambio de modelo.

## Q1 · Ruta y estructura

- **Ruta**: `~/.gemini/tmp/<slug>/chats/session-<AAAA-MM-DDTHH-MM>-<8 primeros del sessionId>.jsonl`, con la hora en **UTC**
  (`toISOString`) y `-1`, `-2`… ante colisión *(LEÍDO: `core/src/services/chatRecordingService.ts`,
  `ChatRecordingService.initialize`)*. **MEDIDO**: la hora coincide con `startTime`, y el id corto con el de `sessionId`.
- **Subagentes**: `chats/<sessionId del padre>/<sessionId>.jsonl`, con `kind: "subagent"` *(`core/src/agents/local-executor.ts`
  inicia su `GeminiChat` con `initialize(undefined, "subagent")`)*. **NO SE PUDO MEDIR**: la copia no tiene subagentes.
- **`<slug>`**: el **nombre base** del directorio del proyecto, en minúsculas y `[a-z0-9-]`. Si ya lo tiene otra ruta, se usa
  `-1`, `-2`… El mapa ruta → slug está en `~/.gemini/projects.json` *(LEÍDO: `core/src/config/projectRegistry.ts`,
  `ProjectRegistry.getShortId`, `claimNewSlug`)*. Dos proyectos homónimos dan `nombre` y `nombre-1`: **el slug no lo identifica**.
- **`.project_root`**: es la **marca de propiedad** del slug. Guarda la ruta absoluta normalizada del proyecto (`path.resolve`, y
  **en minúsculas en Windows**). Con ella se verifica el slug *(`verifySlugOwnership`)* y se recupera si se pierde
  `projects.json` *(`findExistingSlugForPath`)*. **MEDIDO**: ruta absoluta de Windows, en minúsculas, cuyo nombre base es el slug.
- **Reubicación**: `GEMINI_CLI_HOME` sustituye al **home**, no a la carpeta. Las sesiones quedan en
  `$GEMINI_CLI_HOME/.gemini/tmp/…` *(LEÍDO: `core/src/utils/paths.ts`, `homedir`)*. Con `SANDBOX=sandbox-exec` (macOS), la raíz
  es `~/.cache/.gemini` *(`core/src/config/storage.ts`, `Storage.getGlobalRuntimeDir`)*.
- **Carpetas antiguas**: antes de los slugs (entre `v0.28.0` y `v0.29.0`), la carpeta era el SHA-256 de la ruta. Al abrir el
  proyecto, la 0.63 la **copia** a `<slug>` (`fs.cp`) sin borrar el original *(`Storage.performMigration`,
  `core/src/config/storageMigration.ts`, `migrateDirectory`)*. Una sesión puede estar en **dos carpetas**.

## Q2 · Forma de los registros *(MEDIDO; LEÍDO: `loadConversationRecord`)*

Cada línea es un JSON. Hay cuatro formas, y se distinguen por sus claves, no por un `type` común:

| forma | clave que la marca | líneas en la copia | ¿tokens? |
|---|---|---:|---|
| cabecera | `sessionId` + `projectHash` (+ `startTime`, `lastUpdated`, `kind`) | 1 | no |
| mensaje | `id` (+ `timestamp`, `type`, `content`) | 26 | sólo `type: "gemini"` |
| metadatos | `$set` (`lastUpdated`, `sessionId`, `messages`, `summary`, `directories`, `memoryScratchpad`) | 26 | dentro de `$set.messages` |
| rebobinado | `$rewindTo` (id de mensaje) | 0 | no |

- Los mensajes tienen `type` `user` (10), `gemini` (15) e `info` (1, sin tokens). Los `gemini` llevan `id`, `timestamp`, `type`,
  `content`, `model`, `thoughts` (lista de `{subject, description, timestamp}`), `tokens` y, a veces, `toolCalls`.
- **Los tokens** están en el objeto `tokens` de cada mensaje `gemini`: `input`, `output`, `cached`, `thoughts`, `tool` y `total`.
  **LEÍDO** (`ChatRecordingService.recordMessageTokens`): vienen de `usageMetadata`. `input` ← `promptTokenCount`, `output` ←
  `candidatesTokenCount`, `cached` ← `cachedContentTokenCount`, `thoughts` ← `thoughtsTokenCount`, `tool` ←
  `toolUsePromptTokenCount` y `total` ← `totalTokenCount`, todos `?? 0`.
- Las 26 líneas `$set`: 21 sólo cambian `lastUpdated`, 2 cambian `sessionId` y 3 reescriben `messages` entera (al arrancar, en la
  compresión y en la reanudación).
- **Apariciones de tokens**: 15 líneas de mensaje con tokens (10 `id` + 5 repetidas por `recordToolCalls`) **+ 1 aparición
  dentro de `$set.messages`** = **16 apariciones**.

## Q3 · Partidas

- **La caché va DENTRO de la entrada.** **MEDIDO**: `total = input + output + thoughts + tool` en las 15 líneas de mensaje con
  tokens, sin ninguna excepción. La aparición de `$set.messages` es copia de una de ellas. En las dos respuestas con caché,
  `cached < input` (4 051 de 9 204 y 8 090 de 12 420). **LEÍDO**: `promptTokenCount` incluye la caché. `/stats model` muestra
  «Input» = `prompt − cached` *(`core/src/telemetry/uiTelemetry.ts`, `processApiResponse`)*.
- **El razonamiento va APARTE de la salida.** **MEDIDO**: suma en `total`, no en `output` (Σ`output` 476, Σ`thoughts` 2 006). La
  «salida» de `/stats` (409 en la ejecución 1) es Σ`output`, sin los *thoughts*.
- **`tool`** es `toolUsePromptTokenCount`, el *prompt* de las herramientas que ejecuta la propia API: entrada, fuera de `input`,
  sumada en `total`. **MEDIDO**: 0 en todas. Las herramientas locales no cuentan aquí (sus resultados van al siguiente `input`).
- **No hay escritura de caché**: `usageMetadata` no tiene esa partida, y la CLI no crea cachés explícitas (no hay `caches.create`;
  `cachedContent` sólo se reenvía) *(LEÍDO: `core/src/code_assist/converter.ts`)*. La caché es la implícita de la API.

## Q4 · Identidad y repeticiones

- **Identificador**: un `id` UUID v4 por mensaje *(`ChatRecordingService.newMessage`)*, que se conserva cada vez que el mensaje se
  reescribe *(`pushMessage`, `updateMessagesFromHistory`)*. **MEDIDO**: 16 apariciones con tokens (Q2), 10 `id` distintos y **0**
  `id` con tokens distintos.
- **El fichero es de sólo añadir** (`appendFileSync`). La excepción: si al reanudar no se puede releer, se reescribe entero
  (temporal + `rename`), y el viejo queda como `<fichero>.unreadable-<ms>` *(`rewriteConversationFile`)*.
- **Las repeticiones** son 6 en la copia (16 − 10), por tres vías *(LEÍDO)*: (1) `recordToolCalls` vuelve a añadir el mensaje
  entero con `toolCalls`, mismos `tokens` y `timestamp` (5 en la copia); (2) `recordMessageTokens`, si los tokens llegan tras el
  mensaje (0); (3) `$set.messages` (compresión, reanudación, rebobinado) repite los mensajes vivos **con sus tokens** (1: la de
  `gemini-3.5-flash-lite`, tras la reanudación). Sumar todo da 148 024 en `gemini-3.8-flash` (son 96 768) y 22 240 en `-lite` (11 120).
- **Al reconstruir, `$set.messages` también BORRA.** **MEDIDO**: si se reconstruye el estado final como lo hace la CLI
  (`loadConversationRecord`), quedan **3** respuestas con tokens (Σ`input` 33 663) de las **10** que hubo (107 888). La
  compresión cambia el historial por un resumen sin tokens, y `$rewindTo` hace lo mismo con lo rebobinado. **El fichero hay que
  leerlo como bitácora, línea a línea, y quedarse con la primera aparición de cada `id`.**
- **Dos pérdidas posibles** *(LEÍDO; ninguna en la copia)*: `queuedTokens` se pisa si llegan dos `usageMetadata` antes del
  mensaje; y una respuesta rechazada por `InvalidStreamError` (sin texto, sólo *thoughts*, `MAX_TOKENS`…) se reintenta **sin
  registrar sus tokens** *(`core/src/core/geminiChat.ts`, `processStreamResponse`)*.

## Q5 · Modelo

- Viene en `model`, en cada mensaje `gemini`, y es el modelo **pedido**, tras el enrutado y la *fallback* (`lastModelToUse`), no
  el `modelVersion` de la respuesta *(`GeminiChat.makeApiCallAndProcessStream` → `processStreamResponse`)*. `/stats` usa
  `modelVersion` si viene *(`core/src/core/loggingContentGenerator.ts`, `loggingStreamWrapper`)*.
- **MEDIDO**: 9 respuestas de `gemini-3.8-flash` y 1 de `gemini-3.5-flash-lite`, con los mismos nombres que en `/stats`. El modelo
  cambia **dentro de la sesión**: hay que leerlo en cada mensaje.

## Q6 · Llamadas auxiliares ⭐

- **No quedan en el fichero.** **LEÍDO**: sólo escribe en él `GeminiChat` (agente principal y subagentes). El compresor hace
  **dos** llamadas (resumen y `promptId-verify`) por `BaseLlmClient.generateContent` *(`core/src/context/chatCompressionService.ts`)*;
  el enrutador, por `BaseLlmClient.generateJson` *(`core/src/routing/strategies/classifierStrategy.ts`)*; y
  `core/src/core/baseLlmClient.ts` no toca la grabación. Hay más papeles `utility_*` *(`core/src/telemetry/llmRole.ts`)*; no se
  comprobó uno a uno por dónde van.
- **MEDIDO**: no hay ningún registro de `gemini-3-flash-preview`, ni de 3 997/984 ni de 1 399/32 (contraste abajo).

## Q7 · Proyecto y tiempo

- **No hay `cwd` en los registros.** La cabecera trae `projectHash`: el SHA-256 de la ruta **con sus mayúsculas originales**
  *(`core/src/utils/paths.ts`, `getProjectHash`)*. `directories` sólo aparece en subagentes o tras `/dir add`.
- **La ruta legible** está en `.project_root`. **MEDIDO**: el SHA-256 de su texto **no** es `projectHash` (que sale de la misma
  ruta con sus mayúsculas): en Windows no son intercambiables.
- **Tiempo**: `timestamp` en ISO-8601 UTC (`Z`) en cada mensaje, fijado al acabar la respuesta. **MEDIDO**: es monótona, y es
  igual en todas las repeticiones de un mismo `id`.

## Q8 · Reanudación *(MEDIDO; LEÍDO: `ChatRecordingService.initialize` con `resumedSessionData`)*

- **Es el mismo fichero**, con el mismo `sessionId`. **Se duplica** la respuesta de `-lite`, con tokens, en el `$set.messages`
  de la reanudación; deduplicar por `id` basta.
- **Deja una pareja**: `$set {sessionId}` (el **mismo** valor) y `$set {messages}`, igual que la compresión (que reinicia el chat).
  En la copia las distingue que la compresión trae `id` nuevos (4, el resumen) y la reanudación ninguno. No hay registro de «arranque».
- **`/stats` no se hidrata con `--resume latest`**: `uiTelemetryService.hydrate` sólo lo llama el navegador de sesiones
  *(`cli/src/ui/hooks/useSessionBrowser.ts`)*.

## Q9 · Retención *(LEÍDO)*

- **Borra por defecto.** Lo controla `general.sessionRetention` *(`cli/src/config/settingsSchema.ts`)*: `enabled` = **true**,
  `maxAge` = **"30d"**, `maxCount` sin valor y `minRetention` = "1d" *(`cli/src/utils/sessionCleanup.ts`)*. Los valores por
  defecto valen aunque falte el objeto *(`cli/src/config/settings.ts`, `getDefaultsFromSchema`)*.
- Corre en cada arranque *(`cli/src/gemini.tsx`, `main` → `cleanupExpiredSessions`)*, sobre `chats/` del proyecto abierto. Borra
  las sesiones con `lastUpdated` de más de 30 días, nunca la actual. **Al salir**, borra la actual si no tiene nada reanudable
  *(`deleteCurrentSessionIfNotResumableAsync`)*.

## Q10 · Estabilidad del formato *(LEÍDO en el repo público)*

- **Hasta `v0.38.0`**: un único `session-*.json` (`ConversationRecord` con `messages[]`) reescrito entero (`writeFileSync`), con
  los mismos seis campos en `tokens`. **Desde `v0.39.0`** (publicada el 2026-04-23): JSONL de sólo añadir.
- **¿Hacen falta los dos?** La 0.63 aún lee `.json` *(`parseLegacyRecordFallback`)*. Al **reanudar** uno, lo convierte en
  `<nombre>.jsonl` con los mismos `id`, **sin borrar** el `.json` *(`initialize`)*: la sesión puede estar en los dos formatos.
- En un `.json` antiguo, una compresión **reescribía** el historial y borraba los tokens previos. **NO SE PUDO MEDIR** (no hay `.json`).

## Q11 · `logs.json` *(LEÍDO: `core/src/core/logger.ts`, `Logger.logMessage`; MEDIDO)*

Es el historial de lo que **escribe el usuario**: indicaciones y órdenes `/…`, las que recupera la flecha arriba. Es una lista JSON
de **13** entradas, todas `type: "user"` y de la misma sesión, con los campos `sessionId`, `messageId` (entero, de 0 a 12), `type`,
`message` y `timestamp`. No tiene tokens ni modelo. La carpeta `logs/` es de depuración *(`Storage.getProjectTempLogsDir`)*.

## Q12 · Coste

- **Para tarifar** basta con cruzar `model` (`gemini-3.8-flash`, `gemini-3.5-flash-lite`) con la tabla. Las partidas: entrada sin
  caché (`input − cached`), caché leída (`cached`), salida (`output`, más `thoughts` si se cobran como salida) y `tool` (en
  principio, entrada). Lo confirma el dueño. Hoy `internal/pricing` no tiene ningún modelo Gemini.
- **Lo que puede cambiar la tarifa**: un umbral de *contexto largo* se puede aplicar (`input` es **por petición**). El *nivel de
  servicio*, el lote, la cuota gratuita y el modo de autenticación (cuenta de Google, clave o Vertex) **no** constan; no hay
  `trafficType` ni `serviceTier` en el código, y de eso depende que haya coste. Las *llamadas auxiliares* (Q6) cuestan y no están.

## Q13 · ¿Hay una fuente completa en la propia CLI? *(LEÍDO; nada medido, no hay muestra)*

- **Sí, a fichero local, pero opcional y APAGADA por defecto**: la telemetría OpenTelemetry.
  - **Se enciende** con `telemetry.enabled` o `GEMINI_TELEMETRY_ENABLED=true|1`; por defecto, **false**
    *(`core/src/config/config.ts`, `Config.getTelemetryEnabled`)*. **El fichero**, con `telemetry.outfile` o
    `GEMINI_TELEMETRY_OUTFILE`: **sin ruta por defecto** (una relativa cuenta desde donde se arranca); el entorno gana a los
    ajustes. No hay opción de línea de órdenes: `loadCliConfig` *(`cli/src/config/config.ts`)* sólo pasa entorno y ajustes a
    `resolveTelemetrySettings` *(`core/src/telemetry/config.ts`)*.
  - **Precedencia** *(`core/src/telemetry/sdk.ts`, `initializeTelemetry`)*: `target: "gcp"` sin colector va a Google Cloud y **se
    ignora el fichero**; si no, `outfile` gana a OTLP (por defecto `localhost:4317`); sin nada, consola. `target` por defecto: `local`.
- **Eventos por llamada** *(`core/src/telemetry/loggers.ts`, `logApiResponse` / `logApiError`; `types.ts`, `toLogRecord`)*:
  - **`gemini_cli.api_response`**: `model`, `role`, los seis `*_token_count` (`input_`, `output_`, `cached_content_`, `thoughts_`,
    `tool_`, `total_`), `duration_ms`, `status_code`, `prompt_id`, `auth_type`, `finish_reasons` y `event.timestamp`. También los
    comunes: `session.id`, `installation.id`, `interactive`, `user.email` (si hay cuenta) y `experiments.ids`. Con `logPrompts`
    (por defecto **true**) añade `response_text`.
  - **`gemini_cli.api_error`**: `model`, `role`, `error`, `error.type`, `status_code`, `duration_ms`, `prompt_id`, `auth_type`; **sin tokens**.
  - **Identificador**: `api_response` **no** trae uno por petición (`prompt_id` es del turno; el compresor usa `<id>` y `<id>-verify`).
    Sólo lo trae el `gen_ai.client.inference.operation.details` que sale junto a cada respuesta (`toSemanticLogRecord`,
    `gen_ai.response.id`), que con `traces` y `logPrompts` lleva además los mensajes.
- **Cobertura: es lo mismo que cuenta `/stats`.** `logApiResponse` alimenta `uiTelemetryService.addEvent` y emite el registro
  OTel; el evento nace en `LoggingContentGenerator`, y `BaseLlmClient` usa ese mismo generador
  (`new BaseLlmClient(this.contentGenerator, …)`). Entran el compresor, el enrutador y los demás `utility_*`, cada intento
  fallido (`_logApiError`) y las respuestas que luego descarta `InvalidStreamError`.
- **El fichero** *(`core/src/telemetry/file-exporters.ts`, `FileExporter`)* es de **sólo añadir** (`createWriteStream(…, {flags:
  "a"})`); el código no lo rota ni lo borra. Mezcla trazas, registros y métricas (acumuladas, cada 10 s). Cada objeto va con
  `JSON.stringify(…, 2)`: **JSON indentado en varias líneas, no JSONL**. Va por lotes (`BatchLogRecordProcessor`): si el proceso
  muere, puede faltar lo último.
- **Privacidad**: con los valores por defecto lleva el texto de las respuestas y el correo de la cuenta; exige lista blanca de campos.

## Contraste con `/stats` *(MEDIDO)*

**LEÍDO**: `/stats` (= `/stats session`) pinta `ModelUsageTable` *(`cli/src/ui/components/StatsDisplay.tsx`)*. Su «Input
Tokens» es Σ`promptTokenCount`, **con la caché dentro**. `/stats model`, en cambio, da «Input» = `prompt − cached`.

| ejec. | modelo | papel | `/stats` pet. · entr. · caché · sal. | fichero: resp. · Σ`input` · Σ`cached` · Σ`output` | cuadra |
|---|---|---|---|---|---|
| 1 | gemini-3.8-flash | main | 17 · 74 225 · 12 141 · 409 | 7 · 74 225 · 12 141 · 409 | tokens sí; peticiones no (+10) |
| 1 | gemini-3-flash-preview | utility_compressor | 2 · 3 997 · 0 · 984 | — | no está (Q6) |
| 1 | gemini-3.5-flash-lite | main | 1 · 11 120 · 0 · 18 | 1 · 11 120 · 0 · 18 | sí |
| 2 | gemini-3.5-flash-lite | utility_router | 1 · 1 399 · 0 · 32 | — | no está (Q6) |
| 2 | gemini-3.8-flash | main | 3 · 22 543 · 0 · 49 | 2 · 22 543 · 0 · 49 | tokens sí; peticiones no (+1) |
| | **total** | | 24 · 113 284 · 12 141 · 1 492 | 10 · 107 888 · 12 141 · 476 | falta lo auxiliar: 3 · 5 396 · 0 · 1 016 |

- **El papel `main` cuadra al token**: 107 888 + 3 997 + 1 399 = 113 284, y 476 + 984 + 32 = 1 492. Que la «entrada» de `/stats`
  sea Σ`input` sin restar la caché confirma que la incluye. Esa vista no enseña Σ`thoughts` = 2 006 (1 365 + 397 + 244) ni Σ`tool` = 0.
- **Las 11 peticiones de más** (10 + 1) **no aportan ningún token**. `/stats` cuenta cada error de la API como petición
  (`processApiError` suma a `totalRequests`, no a los tokens), y `GeminiChat` reintenta: lo coherente es que fueran intentos
  fallidos. **NO SE PUDO MEDIR** sin la fila «Errors» de `/stats model`.

## Preguntas abiertas

1. ¿Las 11 peticiones sin tokens son errores? Lo diría la fila «Errors» de `/stats model` en una sesión de ensayo.
2. **Subagentes**: ¿van a `chats/<padre>/<id>.jsonl` con `kind: "subagent"`, como dice el código? Falta una sesión que los lance.
3. **`.json` (≤ 0.38)**: ¿quedan en máquinas reales, y conviven con su `.jsonl` convertido? Hace falta una muestra.
4. **Carpetas por hash** (< 0.29) copiadas a `<slug>`: ¿es frecuente ver la misma sesión en dos sitios?
5. `info` (1 registro, contenido vacío): no se identificó quién lo emite. No lleva tokens.
6. Si `modelVersion` no coincidiera con el modelo pedido, fichero y `/stats` darían nombres distintos. No pasa en la copia.
7. **Retención de 30 días**: ¿pasa el lector antes de que se borre?
8. **Telemetría a fichero (Q13)**: falta una muestra para ver la forma exacta de cada objeto (el SDK serializa sus `LogRecord`) y
   cómo emparejar `api_response` con su `operation.details` sin identificador por petición. Encenderla toca el `settings.json` o
   el entorno del usuario, y por defecto el fichero lleva texto y correo.

## Qué cambia respecto al lector de Codex

**Reutilizable**
- La lectura local, línea a línea y con desplazamiento: es JSONL de sólo añadir, como en Codex 0.160.
- El `event_id` determinista con prefijo de longitud (`hashEventID` / `derivarEventIDCodex`), con la herramienta `gemini` y el
  `id` del mensaje como identificador, quizá con el `sessionId`.
- La deduplicación de la pasada (`emitidos`), y el recuento de repetidas e incoherentes.
- `event.Ref` para `session_ref` (`sessionId`) y `project_ref`; `cost_available: false` sin tarifa. El contrato de `Event` no cambia.

**Nuevo**
- **Varias formas de línea, sin `type` común**: el consumo está en los `tokens` de un mensaje `gemini`, no en un registro propio.
- **Las repeticiones son la norma** (Q4): primera aparición por `id`, **también en `$set.messages`**; nunca el estado final.
- **Seis partidas, no cuatro**: la caché va dentro de `input`, como en Codex, pero **no hay escritura de caché**; el razonamiento
  va **fuera** de `output`, al revés que en Codex; `tool` va fuera de `input`. Qué va a `tokens_input`/`tokens_output`, el dueño.
- **Proyecto**: no hay `cwd` por registro. Sale de `.project_root` (minúsculas en Windows) o de `projectHash` (que no se puede
  invertir), nunca del nombre de la carpeta, que se repite.
- **Qué ficheros leer**: `~/.gemini/tmp/*/chats/`, más el nivel de subagentes. Hay que contar con `GEMINI_CLI_HOME`, los `.json`
  antiguos y las carpetas duplicadas por la migración. `logs.json` no es una sesión.
- **Lo que no está en el fichero**: las llamadas auxiliares y los intentos rechazados. El lector se queda por debajo de la
  factura: aquí, 5 396 de 113 284 de entrada (4,8 %) y 1 016 de 1 492 de salida.
- **Segunda fuente posible (Q13)**: la telemetría a fichero sí cubre todo esto, pero está apagada por defecto. Es JSON indentado,
  no JSONL, y no trae identificador por petición en `api_response`. Codex no tiene nada equivalente en lo descubierto en 008.
- **Borrado automático** a los 30 días, por defecto.
