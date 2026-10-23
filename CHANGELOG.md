# Changelog

Cambios visibles del agente `permea`, de la versión más reciente a la más antigua. Cada punto cita la
especificación de la que sale (`specs/NNN-…/spec.md`).

## 0.10.0 — PENDIENTE

### Nuevo
- `permea project list`: enseña, en esta instalación, qué carpeta hay detrás de cada identidad del panel, para poder asignar las
  que están sin asignar. Una fila por identidad, con sus 12 primeros caracteres, la carpeta, las herramientas por su nombre, los
  eventos y las fechas en UTC, como el panel. (`specs/013-lista-de-proyectos/spec.md`, D-1, D-2, D-4, FR-001, FR-009 a FR-022)
- Marca `(sin repositorio)` la carpeta sin `.git`, y `(ya no existe)` la que ya no está. (`specs/013-lista-de-proyectos/spec.md`,
  D-3, P-4, FR-016)
- Es local y de solo lectura: no envía nada, no abre la red, no escribe en disco ni crea la sal, y no necesita enrolamiento. Es el
  primer comando que imprime rutas reales del disco, sólo en tu terminal. (`specs/013-lista-de-proyectos/spec.md`, D-5, D-6, D-10,
  FR-004 a FR-008)

### Sin cambios
- `--run`, `--daemon`, `--scan`, `enroll`, `status` y `project join` funcionan como en la 0.9.0, y lo enviado es el mismo, byte a
  byte. Cambian sólo la ayuda y el mensaje de «Verbos disponibles», que ahora dice `join, list`. (`specs/013-lista-de-proyectos/spec.md`,
  D-9, FR-001, FR-003, FR-024, SC-012)

### Limitaciones conocidas
- La lista cuenta lo que las herramientas conservan en disco, también lo aún no enviado: el panel puede tener más o menos.
  (`specs/013-lista-de-proyectos/spec.md`, D-4, §Limitaciones)
- No lista las identidades de antes de la 0.3.0. (`specs/013-lista-de-proyectos/spec.md`, D-7)
- La identidad se recalcula con las carpetas como están hoy: si una se movió, si se hizo `git init` después o si cambió un enlace,
  no es la que viajó. (`specs/013-lista-de-proyectos/spec.md`, §Limitaciones)
- Un mensaje de Claude Code todavía abierto, con menos de 10 minutos sin cambios, no cuenta aún. (`specs/013-lista-de-proyectos/spec.md`,
  P-9)
- En Windows, la misma carpeta escrita con otras mayúsculas da dos identidades, y una ruta con caracteres de doble ancho descuadra la
  tabla. (`specs/013-lista-de-proyectos/spec.md`, P-12, §Limitaciones)

## 0.9.0 — 2026-10-16

### Nuevo
- Lee también el consumo de Kimi Code CLI, de `~/.kimi-code` (o de `$KIMI_CODE_HOME`), si tiene `sessions/`: el `wire.jsonl` de
  cada agente, también de los subagentes, y `session_index.jsonl`; nunca `config.toml` ni `credentials/`. Cada petición terminada
  y cada compactación es un evento con `tool = kimi_code`. La carpeta sólo se lee. (`specs/012-lector-kimi/spec.md`, D-1, D-2,
  P-1, P-2, P-3, FR-006, FR-010)
- Los eventos de Kimi Code CLI llevan sus tokens y su modelo, el alias sin su proveedor, sin coste: el coste lo calcula la
  plataforma con su tarifa. Lo que un `/fork` copia de la sesión de origen no se cuenta otra vez. (`specs/012-lector-kimi/spec.md`,
  D-3, D-4, D-5, P-9, FR-014, FR-017)
- `--scan` acepta también un `wire.jsonl` de Kimi Code CLI, sin línea de control. (`specs/012-lector-kimi/spec.md`, P-12, FR-022)

### Sin cambios
- Claude Code, Codex, Gemini, OpenCode y Copilot CLI se leen exactamente como en la 0.8.0. Sin carpeta de Kimi Code CLI, la salida
  no cambia. (`specs/012-lector-kimi/spec.md`, FR-023, FR-028, SC-015)

### Requiere
- La tarifa de la plataforma para `kimi-k2.7-code`, `kimi-k2.7-code-highspeed`, `kimi-k2.6` y `kimi-k3`, desplegada el 2026-10-15
  (`permea-platform` · `6e7cedd`). Otro modelo llega sin coste hasta que tenga su fila. (`specs/012-lector-kimi/spec.md`, D-5,
  §Dependencias)

### Limitaciones conocidas
- Si se mata una sesión, la petición que estaba en curso no está en disco y no se cuenta, aunque Kimi la cobra. Tampoco se cuentan
  los intentos fallidos o interrumpidos con Esc. (`specs/012-lector-kimi/spec.md`, D-6, §Límites declarados)
- Si una pasada cae justo entre la copia de un `/fork` y su marca, lo copiado sale como propio: es cuestión de milisegundos.
  (`specs/012-lector-kimi/spec.md`, §Límites declarados)
- En Windows, la misma carpeta de trabajo escrita con otras mayúsculas da otro proyecto. (`specs/012-lector-kimi/spec.md`, P-6,
  §Límites declarados)
- Medido sólo con Kimi Code 2.2.0, en Windows y con clave de API de Kimi Platform. La membresía (OAuth) y los otros proveedores no
  tienen muestra. (`specs/012-lector-kimi/spec.md`, §Límites declarados, Q-5)

## 0.8.0 — 2026-10-14

### Nuevo
- Lee también el consumo de Copilot CLI, de `~/.copilot/session-state` (o de `$COPILOT_HOME/session-state`), si existe. Cada llamada
  al modelo principal y cada compactación con éxito es un evento con `tool = copilot_cli`. La carpeta sólo se lee.
  (`specs/011-lector-copilot/spec.md`, D-5, FR-001, FR-002, FR-006, FR-010)
- Los eventos de Copilot CLI llevan sus tokens y el modelo de cada llamada, sin coste: el coste lo calcula la plataforma. El
  razonamiento ya va dentro de la salida. Los créditos de Copilot no se envían. (`specs/011-lector-copilot/spec.md`, D-3, D-4,
  FR-015, FR-017)
- `--scan` acepta también una sesión de Copilot CLI y añade una línea de control con sus créditos y su equivalente en USD, que no
  se envía. (`specs/011-lector-copilot/spec.md`, P-9, FR-022)

### Sin cambios
- Claude Code, Codex, Gemini y OpenCode se leen exactamente como en la 0.7.0. Sin carpeta de Copilot CLI, la salida no cambia.
  (`specs/011-lector-copilot/spec.md`, FR-023, FR-028, SC-015)

### Limitaciones conocidas
- Las llamadas auxiliares de Copilot CLI (a 0 créditos en el plan Free) no se cuentan: en tokens, el consumo enviado queda por
  debajo del real. (`specs/011-lector-copilot/spec.md`, D-2)
- Si se mata una sesión, la llamada que estaba en curso no está en disco y no se cuenta. (`specs/011-lector-copilot/spec.md`,
  §Límites declarados, SC-004)
- Los subagentes en segundo plano, los *workflows* y la compactación automática no tienen muestra, y `session.db` no se lee. Un
  subagente síncrono sí se cuenta. (`specs/011-lector-copilot/spec.md`, P-10, SC-023)
- Medido sólo con Copilot CLI 1.0.97, en Windows y en el plan Free. Con un plan de pago, un modelo con punto en el nombre (como
  `claude-haiku-4.5`) no casa con el catálogo de la plataforma, y el evento llega sin coste. (`specs/011-lector-copilot/spec.md`,
  §Límites declarados, §Dependencias, P-5)
- Copilot dentro de VS Code y JetBrains no se lee. (`specs/011-lector-copilot/spec.md`, D-1, N-5)

## 0.7.0 — 2026-10-13

### Nuevo
- Lee también el consumo de OpenCode 2.x, de `~/.local/share/opencode` (o de `$XDG_DATA_HOME/opencode`, o de la base que indique
  `OPENCODE_DB`), si existe. Cada llamada al modelo es un evento con `tool = opencode`. La base sólo se lee.
  (`specs/010-lector-opencode/spec.md`, FR-001, FR-002, FR-003, FR-012)
- Los eventos de OpenCode llevan sus tokens y su modelo, sin coste: el coste lo calcula la plataforma. El razonamiento cuenta como salida.
  (`specs/010-lector-opencode/spec.md`, D-1, P-2)
- `--scan` acepta también una base de OpenCode y añade una línea de contraste con el coste según OpenCode, que no se envía.
  (`specs/010-lector-opencode/spec.md`, FR-025)

### Sin cambios
- Claude Code, Codex y Gemini se leen exactamente como en la 0.6.0. Sin carpeta de OpenCode, la salida no cambia.
  (`specs/010-lector-opencode/spec.md`, FR-026)

### Requiere
- Una plataforma que cobre a 5 minutos la escritura de caché de OpenCode con modelos de Anthropic. Por eso esta versión se publica
  cuando la plataforma ya lo hace. (`specs/010-lector-opencode/spec.md`, SC-027)

### Limitaciones conocidas
- La llamada que pone título a cada sesión no queda entre los mensajes: el consumo enviado queda por debajo de la factura, un 0,37 %
  del coste en la medición. (`specs/010-lector-opencode/spec.md`, D-2)
- Una llamada en curso más de 24 horas se cuenta como «abandonada» y no se envía. (`specs/010-lector-opencode/spec.md`, P-3)
- Las bases de OpenCode 1.x se cuentan como «formato anterior» y no se envían. (`specs/010-lector-opencode/spec.md`, P-9, FR-022)
- Con proveedores distintos de Google y Anthropic, el modelo puede no casar con el catálogo de la plataforma, y el evento llega sin
  coste. (`specs/010-lector-opencode/spec.md`, P-5)

## 0.6.0 — 2026-10-08

### Nuevo
- Lee también el consumo de Gemini CLI, de `~/.gemini/tmp` (o de `$GEMINI_CLI_HOME/.gemini/tmp`), si esa
  carpeta existe. Cada respuesta del modelo es un evento con `tool = gemini`.
  (`specs/009-lector-gemini/spec.md`, FR-001, FR-002, FR-006)
- Los eventos de Gemini llevan sus tokens y su modelo, sin coste: el coste lo calcula la plataforma.
  (`specs/009-lector-gemini/spec.md`, FR-010, FR-011, FR-013)

### Sin cambios
- Claude Code y Codex se leen exactamente como en la 0.5.0. Sin carpeta de Gemini, la salida no cambia.
  (`specs/009-lector-gemini/spec.md`, FR-023, FR-028)

### Limitaciones conocidas
- Las llamadas internas de Gemini CLI y los intentos fallidos no constan en sus sesiones: el consumo
  enviado queda por debajo de la factura. (`specs/009-lector-gemini/spec.md`, D-1)
- Sólo se leen las sesiones de Gemini CLI 0.39.0 o posterior; las anteriores se cuentan y no se envían.
  (`specs/009-lector-gemini/spec.md`, FR-018; D-3)

## 0.5.0 — 2026-10-07

### Nuevo
- Lee también el consumo de Codex CLI, de `~/.codex/sessions` (o de `$CODEX_HOME/sessions`), si esa carpeta
  existe. Cada respuesta del modelo es un evento con `tool = codex`.
  (`specs/008-lector-codex/spec.md`, FR-001, FR-002, FR-006)
- Los eventos de Codex llevan sus tokens y su modelo, sin coste: el coste lo calcula la plataforma.
  (`specs/008-lector-codex/spec.md`, FR-010, FR-011, FR-013)

### Sin cambios
- Claude Code se lee exactamente como en la 0.4.0. Sin carpeta de Codex, la salida no cambia.
  (`specs/008-lector-codex/spec.md`, FR-021, FR-023)

### Limitaciones conocidas
- Sólo se leen las sesiones de Codex 0.153.0 o posterior. Las anteriores se cuentan como «formato anterior»
  y no se envían.
  (`specs/008-lector-codex/spec.md`, FR-017; D-2)
- Los ficheros comprimidos de Codex (`.zst`) no se leen; se cuentan.
  (`specs/008-lector-codex/spec.md`, FR-018)

## 0.4.0 — 2026-10-07

### Cambia
- El coste de la escritura de caché usa su duración: la de 5 minutos a su tarifa y la de 1 hora a la suya (el doble de la entrada).
  Antes toda iba a la de 5 minutos, y el coste salía por debajo.
  (`specs/007-coste-fiel/spec.md`, FR-001, FR-002, FR-007)
- Un mensaje que Claude Code escribe en varias líneas se cuenta entero: cada partida vale lo más alto que alcanza. Antes contaba
  la primera línea, y en las conversaciones de subagentes podía faltar parte de la salida.
  (`specs/007-coste-fiel/spec.md`, FR-009)
- El último mensaje de cada conversación se envía cuando empieza el siguiente o tras 10 minutos sin cambios. Si la pasada
  termina antes, sale en la siguiente.
  (`specs/007-coste-fiel/spec.md`, FR-010, FR-011, FR-014)
- Tarifas de 17 modelos, con la escritura de caché a 1 hora; nueva: claude-fable-5-1.
  (`specs/007-coste-fiel/spec.md`, FR-006, FR-008; `specs/007-coste-fiel/contracts/tarifas.md`)

### Lo ya enviado
- No se corrige ni se reenvía. Un mensaje que la 0.3.0 envió incompleto se queda como llegó.
  (`specs/007-coste-fiel/spec.md`, FR-019)

### Limitaciones conocidas
- El «modo rápido» no se distingue: un mensaje en modo rápido queda por debajo de su coste.
  (`specs/007-coste-fiel/spec.md`, FR-007, N-3)
- Una línea sin el desglose de la caché se tarifa entera a 1 hora.
  (`specs/007-coste-fiel/spec.md`, FR-003, FR-004)

## 0.3.0 — 2026-10-02

Primera versión desde la 0.2.1. Trae el enrolamiento, la identidad y la adhesión a proyecto, y
corrige la medición. **Si venías de la 0.2.1, lee «Cambios que rompen» y «Las cifras bajan».**

### Nuevo

- **`permea enroll`**: conecta la instalación con su organización a partir del *enrollment string*
  que da la aplicación. Verifica el token contra el servidor antes de guardarlo, y sólo si lo acepta
  lo guarda en `config.json` con permisos 0600; un rechazo no escribe nada. Acepta el valor por
  argumento o, recomendado, por stdin. Formato `pmea2.`, con el `dev_id` que asigna el servidor; el
  antiguo `pmea1.` se rechaza.
  (`specs/003-enrolamiento/spec.md`, FR-001 a FR-007, FR-013;
  `specs/003-enrolamiento/contracts/enrollment-string.md`)
- **`permea status`**: dice si la instalación está enrolada y contra qué servidor. Es local y nunca
  muestra el token. (`specs/003-enrolamiento/spec.md`, FR-008)
- **Identidad de proyecto por árbol de trabajo**: los eventos de un mismo proyecto comparten
  identidad, se lance Claude Code desde la raíz o desde cualquier subdirectorio. Antes cada punto de
  lanzamiento era una identidad distinta. (`specs/004-identidad-de-proyecto/spec.md`, P-004 FR-001 a
  FR-005)
- **`permea project join`**: une la instalación a un Proyecto de la organización presentando el código
  que se genera en el panel, y responde con el nombre del Proyecto. Se ejecuta dentro del árbol de
  trabajo, exige HTTPS y no guarda nada en local. (`specs/005-adhesion-a-proyecto/spec.md`, P-005
  FR-001, FR-002, FR-006, FR-017, FR-019)
- **Un evento por mensaje** (006). Claude Code escribe varias líneas por mensaje con el mismo consumo,
  y la 0.2.1 emitía un evento por línea. Ahora sale **uno por mensaje**. (`specs/006-medicion-fiel/spec.md`,
  P-006 FR-001, FR-033)
- **`event_id` determinista** (006): se deriva del propio mensaje, sin sal, así que el mismo mensaje
  da el mismo identificador en cualquier pasada o instalación, y la plataforma descarta los
  repetidos. Las respuestas **`<synthetic>`** de Claude Code ya no se emiten. (P-006 FR-002, FR-007,
  FR-008; `specs/006-medicion-fiel/contracts/event-id.md`)
- **Tarifas de 16 modelos** (006), espejo del catálogo de la plataforma, con su procedencia en la
  cabecera de la tabla. Los modelos en uso (`claude-opus-5-5`, `claude-opus-5`, `claude-sonnet-5`)
  tienen por fin coste. **`claude-opus-4-6` pasa de 15 / 75 a 5 / 25 USD por millón: su coste baja a
  un tercio.** (P-006 FR-014 a FR-020; `specs/006-medicion-fiel/contracts/tarifas.md`)
- **La ayuda** (006): una sola, por stdout, con «Primeros pasos». Cada subcomando tiene la suya
  (`permea <subcomando> -h`), y pedirla no ejecuta nada. (P-006 FR-021, FR-023, D-006-14)
- **`--scan`** imprime por evento las cuatro partidas de tokens y el `event_id`, y cada pasada termina
  con un resumen de recuentos por stderr. (P-006 FR-010, FR-005, FR-006)

### Las cifras bajan

**Las cifras bajan respecto a la 0.2.1 porque antes se contaba de más.** Con un evento por línea,
cada mensaje se contaba tantas veces como líneas ocupaba en el log: en los datos medidos, los tokens
salían **×2,13**. A partir de la 0.3.0 cada mensaje cuenta una vez. Es una corrección, no una pérdida.
(`specs/006-medicion-fiel/spec.md`, §Contexto de partida, M2)

Lo enviado antes de actualizar **no se corrige ni se reenvía**, y lo que la 0.2.1 dejó en cola se
envía tal cual: la corrección aplica a lo que se mida desde la 0.3.0. (P-006 FR-009; §Fuera de alcance,
«Los datos que la plataforma ya recibió»)

### Limitación conocida

Cuando Claude Code escribe varias líneas de un mismo mensaje con los **tokens de salida creciendo**, se
cuenta el valor de **la primera** línea. En los datos medidos en Windows faltan así **un 1,6 % de los
tokens de salida**; la entrada y las cachés no se ven afectadas. Se corregirá en la versión siguiente.
(`specs/006-medicion-fiel/spec.md`, P-006 FR-005, «Hallazgo de W2, 2026-10-02»)

### Cambios que rompen

- **`project_ref` cambia para el mismo proyecto.** La identidad de proyecto se deriva ahora de la raíz
  del árbol de trabajo, así que las identidades emitidas por la 0.2.1 no corresponden con las de la
  0.3.0. Es una ruptura aceptada, y la correspondencia de identidades históricas queda fuera de
  alcance. (`specs/004-identidad-de-proyecto/spec.md`,
  §Assumptions, «Continuidad de identidades históricas: se rompe, y se acepta»)
- **El modo de envío en claro de `project_ref` se retiró.** Una configuración que todavía lo pida
  hace que `--run` y `--daemon` se detengan con un error que lo explica; `permea enroll` lo repara.
  (`specs/004-identidad-de-proyecto/spec.md`, P-004 FR-013 a FR-015;
  `specs/004-identidad-de-proyecto/plan.md`, D-004-5)
- **La ayuda cambia de canal y de forma** (006):
  - **`permea` sin argumentos ya no escribe por stderr**: la ayuda sale por stdout, con exit 0.
  - **Un subcomando inexistente sale con exit 1**, nombrándolo (antes, exit 0 y la ayuda).
  - **Una opción desconocida o sin valor sale con exit 2 sin imprimir el uso**, con un mensaje propio
    que remite a `permea help`.

  (P-006 FR-021, FR-022, D-006-7; `specs/006-medicion-fiel/contracts/cli.md`)
