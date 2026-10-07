# 008 · «Lector de Codex» — Especificación

**Feature Branch**: `008-lector-codex` · **Created**: 2026-10-07 · **Status**: **Ratificada** el 2026-10-07, 12:16 (Madrid) · **Cabeza de partida**: `7b8c77c` *(0.4.0)* · **Enmiendas**: E-1, E-2, E-3
**Input**: las decisiones del dueño `D-1`…`D-3` · las decisiones de método del orquestador `M-1`…`M-9` · el descubrimiento del 07-10 y la
FASE 0 *(`soporte/descubrimiento.md`, sobre la **copia congelada** de 8 sesiones, F1…F8)* · la fuente pública de Codex, etiqueta
`rust-v0.160.1` · el contrato nuevo `contracts/event-id-codex.md`.

> ⛔ Toda afirmación sobre nuestro código lleva `fichero:línea`, medida sobre `7b8c77c`. Las de Codex, ruta bajo `codex-rs/` en
> `rust-v0.160.1`. Las propuestas son decisiones **✅** desde la ratificación, y conservan sus alternativas como rastro. Lo que añade la
> spec sin que nadie lo dijera va marcado **`DECIDÍ YO`**.

## Contexto — el agente sólo ve Claude Code

1. **Lee una sola herramienta.** La raíz es `~/.claude/projects` *(`internal/config/config.go:112-121`)*, y `tool` es siempre
   `"claude_code"` *(`internal/ingest/claudecode.go:123`)*. `Config.Tools` vale `["claude_code"]` *(`config.go:46`)*, y en producción
   sólo se lee para rellenarse *(`:125-126`)*.
2. **Codex CLI guarda su consumo en local**, un JSONL por sesión en `~/.codex/sessions/AAAA/MM/DD/`. Desde la **0.153.0** escribe un
   `token_usage_record` por respuesta terminada, con su `response_id` *(descubrimiento §FASE 0 (b))*. Sobre la copia congelada, la suma
   de esos registros reproduce el acumulado del hilo y las cifras que Codex imprime, **2828** y **4117** en F7 y F8. El `token_count`
   acumulado **no** sirve: pierde la compactación y no lleva identificador *(descubrimiento §El consumo)*.
3. **La plataforma ya acepta el evento tal cual.** `tool` es texto libre, y un modelo sin tarifa que llega sin coste se guarda: sus
   tokens cuentan y su coste queda «ciego y señalizado» hasta que tenga fila *(plataforma: `backend/app/Ingest/EventAllowlist.php:23,56,140-147`;
   `backend/config/pricing.php:85-92`)*.

## Decisiones del dueño *(no se discuten aquí)*

| # | Decisión |
|---|---|
| **D-1** · 06-10 | Tras la 0.4.0, el agente lee también el consumo de **Codex CLI** |
| **D-2** · 07-10 | **Sólo el formato actual**, el que trae un registro por respuesta (`token_usage_record`). Las sesiones en formato anterior se cuentan como «formato no soportado» y **no generan eventos** |
| **D-3** · 07-10 | **Las tarifas viven sólo en la plataforma y entran después del lector.** El agente envía los eventos de Codex con sus tokens y **sin coste** (`cost_usd = 0`, `cost_available = false`). No hay tabla de OpenAI en el agente |

## Decisiones de método *(orquestador; con lo que precisa la FASE 0)*

| # | Decisión | Lo que precisa la medida |
|---|---|---|
| **M-1** | `tool = "codex"`. El evento no cambia: los mismos campos y `SchemaVersion = 1`. `internal/event/`, `eventid.go`, `eventid_test.go`, `boundary_test.go` y `specs/006-medicion-fiel/contracts/event-id.md`, con **0 bytes** de diff | — |
| **M-2** | **Un evento por `token_usage_record`**. Una petición fallida no deja registro ni evento. **No hay espera ni cierre**: la pasada de 007 no se aplica a Codex | **Confirmado**: F6 tiene dos turnos fallidos sin registro, y cada registro es una respuesta completada *(`core/src/session/mod.rs:4768-4800`)* |
| **M-3** | `event_id` = `["permea/event_id/v1", "codex", "respuesta", <response_id>]`, en fichero y función nuevos *(`contracts/event-id-codex.md`)*. El `response_id` no viaja, y una respuesta copiada por una bifurcación da el mismo `event_id` | **Confirmado**: 0 `response_id` repetidos en 7 registros |
| **M-4** | Partidas **sin solape**: caché leída → `tokens_cache_read`; salida, con el razonamiento dentro → `tokens_output`; escritura de caché → `tokens_cache_creation`; entrada sin caché → `tokens_input` | **FASE 0 (c)**: la escritura va **dentro** de la entrada *(documentación oficial)*, así que `tokens_input = input_tokens − cached_input_tokens − cache_write_input_tokens`. En la copia la escritura es 0; lo fija un fixture sintético *(SC-008)* |
| **M-5** | El modelo es el del `turn_context` con el mismo `turn_id`. Si el turno no lo tiene *(la compactación)*, el vigente en el hilo | **Precisado** en FR-011: en F6, 3 de 4 por turno y 1 vigente. Una compactación tras un cambio de modelo **a la baja** usa el modelo anterior *(`core/src/compact_remote_v2.rs:253-290`)*; se aplica el vigente *(Q-4 (a))* |
| **M-6** | `occurred_at` = la marca del registro. `project_ref`, del `cwd` con el resolutor existente. `dev_id`, `machine_ref` y `session_ref`, como hoy | **FASE 0 (e)**: `machine_ref` y `dev_id`, iguales; `session_ref`, del `session_id` del registro *(P-6)*. 7 de 7 registros traen marca, y el `cwd` de cada `turn_context` es igual al de `session_meta` en los 8 ficheros |
| **M-7** | Dónde lee: `CODEX_HOME` si está y no es vacía; si no, `~/.codex`; dentro, `sessions/`. Si no existe, el agente no dice nada y no falla. Estado en el mismo `state.json`, por tamaño y offset; **el `mtime` no decide nada** | Coincide con Codex *(`utils/home-dir/src/lib.rs:14-16,52-59`)*. ⚠️ Es la **primera** lectura de una variable de entorno en producción: hoy no hay ningún `Getenv` *(007 C1)* |
| **M-8** | Versión **`0.5.0`**. La conducta con Claude Code no cambia: los **494** tests actuales siguen verdes **sin tocarlos** | — |
| **M-9** *(Q-3 (a), E-1)* | `internal/testutil/sandbox.go` fija también `CODEX_HOME` *(vacía)*. Entra en el censo **de código** como ayudante. **Ningún `_test.go` existente se toca** | Sin ella, un test de proceso heredaría el `CODEX_HOME` del desarrollador *(`sandbox.go:59-61` fija sólo `HOME`, `USERPROFILE` y `XDG_CONFIG_HOME`)* |

## Propuestas — **ratificadas** por el dueño el 2026-10-07, 12:16 (Madrid)

| # | Propuesta → **decisión** | Alternativas *(rastro)* | Recomendación *(rastro)* |
|---|---|---|---|
| **P-1** ✅ **(a)** | **El lector se activa solo, si la carpeta existe.** Sin interruptor en la 0.5.0 *(Q-5 (a))* | (b) carpeta y `"codex"` en `Config.Tools`: `enroll` guarda la configuración entera con `"tools": ["claude_code"]` *(`cmd/permea/enroll.go:84`, `config.go:85-90`)*, y dejaría Codex apagado en lo ya enrolado · (c) un interruptor de exclusión | **(a)**: cumple D-1 sin migrar configuraciones |
| **P-2** ✅ *(por resultado, E-1)* | **El contexto entre pasadas**: un registro emitido en una pasada posterior lleva el **mismo** modelo, `project_ref` y `session_ref` que si el fichero se hubiera leído de una vez, y `state.json` conserva sus cuatro campos. **El mecanismo lo decide el plan**, con un tope de coste *(SC-014)* | (b) guardar el contexto en `state.json` · (c) emitir lo que no tenga contexto como «sin modelo» | Releer el prefijo sin emitir *(rastro; el plan elige)* |
| **P-3** ✅ **(a), (a)** | **Formato anterior por contenido** *(`token_count` con `info` y ningún `token_usage_record`)*. **Los `.zst` se cuentan** como «comprimido», sin leerlos. Un fichero sin consumo no se cuenta | (b) por versión, `cli_version` < 0.153.0 · `.zst`: (b) ignorarlo · (c) descomprimirlo *(dependencia nueva, Principio III)* | **(a), (a)** |
| **P-4** ✅ **(b)** | **Una respuesta sin modelo establecible** se emite con `model` vacío y se cuenta *(la plataforma admite el nulo, `EventAllowlist.php:56`)* | (a) no emitir · (c) un modelo inventado | **(b)**: la identidad y los tokens son ciertos |
| **P-5** ✅ **(a)** | **El `cwd`** es el del `turn_context` del mismo turno, y si falta, el de `session_meta` | (b) siempre el de `session_meta` | **(a)** |
| **P-6** ✅ **(a)** | **`session_ref`** sale del `session_id` del propio registro: el del hilo raíz *(`protocol/src/protocol.rs:3130`)*, que las copias de una bifurcación conservan | (b) `thread_id` · (c) `session_meta.id` del fichero | **(a)** |
| **P-7** ✅ **(a)** | **Repetidos en la pasada**: se emiten una vez y se cuentan | (b) emitir todos | **(a)** |
| **P-8** ✅ **(a)** | **`--scan`** reconoce Codex por la primera línea, `type = "session_meta"` | (b) `--scan-codex` | **(a)** |
| **P-9** ✅ | **Los textos**, en §Textos aprobados. El resumen de Codex, **en forma de etiqueta y número** *(sustituye al propuesto)*; el resto, tal como se propuso | corregirlos | — |

## User Scenarios & Testing

### Historia 1 — El consumo de Codex llega a la plataforma *(P1)*
1. **Dado** una sesión de Codex ≥ 0.153.0 con dos respuestas, **cuando** corre `--run`, **entonces** salen **dos** eventos `tool = codex`,
   con sus tokens, `cost_available = false` y el modelo de su turno.
2. **Dado** el mismo fichero en una segunda pasada sin cambios, **entonces** no sale **ningún** evento.

### Historia 2 — Las cifras son las de Codex *(P1)*
1. **Dado** F7, **entonces** `tokens_input + tokens_cache_creation + tokens_output` = **2828**, lo que Codex imprimió como «tokens used».

### Historia 3 — Quien no usa Codex no nota nada *(P1)*
1. **Dado** una instalación sin `~/.codex` ni `CODEX_HOME`, **entonces** la salida de `--run` es **byte a byte** la de la 0.4.0, y no hay
   eventos `codex`.

### Historia 4 — Codex sin Claude Code *(P2, E-1)*
1. **Dado** una instalación con carpeta de Codex y sin `~/.claude/projects`, **entonces** los eventos de Codex salen igual.

### Edge Cases
- **Reanudación** *(F6, descubrimiento §FASE 0 (a))*: Codex añade al mismo fichero. Los registros nuevos salen en la pasada siguiente,
  sin repetir los anteriores.
- **Compactación** *(F6)*: su respuesta tiene registro y **no** tiene `turn_context`. Sale con el modelo vigente *(FR-011)*.
- **Bifurcación** *(sin muestra)*: Codex copia los registros del padre a otro fichero. Dan el mismo `event_id`: dentro de la pasada son
  «repetidas» *(P-7)*, y entre pasadas los descarta la plataforma.
- **Subagentes** *(sin muestra; E-1)*: un subagente escribe su propio fichero y **no hereda** el consumo del padre
  *(`core/src/agent/control/spawn.rs:124-125`)*. Sus registros son respuestas propias: se emiten, y su `session_id` es el del hilo raíz
  *(P-6)*.
- **Petición fallida** *(F6, dos turnos)*: sin registro, sin evento y sin recuento.
- **Cambio de modelo** *(F6)*: cada registro lleva el modelo del `turn_context` de su turno.
- **Carpeta ausente**, `CODEX_HOME` vacía o apuntando a algo que no existe: el agente no dice nada, no falla y no escribe la línea de Codex.
- **Fichero truncado o rotado**: se relee desde 0, como hoy *(`internal/state/state.go:113-115`)*. Los registros vuelven a dar los mismos
  `event_id`, y la plataforma descarta los ya recibidos.
- **Línea más larga que el tope**: `Recorrer` no tiene tope *(`state.go:126,129`)*. `--scan` corta en 1 MiB *(`cmd/permea/main.go:447-448`)*
  y termina con error *(`:474`)*. En la copia, la mayor mide 44 745 B. Un `compacted` de una sesión larga no se ha medido.
- **Sesión en curso leída a medias**: una última línea sin `\n` no se consume *(`state.go:130-133`)*. Los registros completos salen ya,
  sin espera *(M-2)*, con su contexto *(FR-005)*.
- **Un fichero con los dos formatos** *(sesión antigua reanudada con ≥ 0.153.0; sin muestra)*: tiene registros, así que es formato actual.
  Se emiten sus registros, y sus `token_count` antiguos se ignoran *(Q-2 (a))*.
- **Partidas incoherentes** *(caché + escritura > entrada; 0 casos)*: no se emite y se cuenta como «incoherente» *(Q-1 (a), FR-027)*.
- **Un fichero de Codex ilegible** *(permisos, enlace roto, error de E/S; E-2)*: se omite con un aviso, sin avanzar su offset, y la pasada
  sigue y guarda el estado de los demás, también los de Claude Code *(FR-028)*.
- **Una línea corrupta en un fichero de Codex** *(E-2)*: si está en la parte nueva, se salta con el aviso de hoy. Si está en el prefijo
  releído, se ignora en silencio. No corta el fichero ni cuenta como respuesta *(FR-029)*.
- **El proceso cae entre encolar y guardar el estado**: la pasada siguiente reencola los mismos `event_id`, y la plataforma los descarta
  *(FR-025)*.

## Requirements

### A · Dónde y cuándo lee *(M-7, P-1)*

- **FR-001** *(raíz)*: la raíz de Codex **DEBE** ser `$CODEX_HOME/sessions` si `CODEX_HOME` está definida y no es vacía, y si no,
  `<directorio personal>/.codex/sessions`, con el mismo `os.UserHomeDir()` que la raíz de Claude Code *(`config.go:116`)*.
- **FR-002** *(activación — P-1, E-2)*: el lector está activo **en una pasada** si y sólo si, **en esa pasada**, la raíz existe y es un
  directorio. La ruta se resuelve una vez por proceso, pero su existencia se comprueba en **cada** `generate()`: el demonio vive días, y
  Codex puede instalarse después. Si no existe, no escribe nada, no devuelve error y no cambia ningún texto.
- **FR-003** *(ficheros)*: se leen todos los `.jsonl` bajo la raíz, a cualquier profundidad. Los `.zst` no se leen y se cuentan *(P-3)*.
- **FR-004** *(estado)*: el estado de cada fichero vive en el mismo `state.json`, con sus **cuatro** campos y la ruta como clave. El offset
  queda en el fin de la última línea completa. Una línea sin `\n` no se consume, y un fichero más corto que su offset se relee desde 0,
  como en `Recorrer` *(`state.go:105-153`)*. **El `mtime` no decide nada.**
- **FR-005** *(contexto, por resultado — P-2, E-1)*: un registro emitido en una pasada posterior **DEBE** llevar el **mismo** `model`,
  `project_ref` y `session_ref` que si el fichero se hubiera leído entero en una sola pasada. `state.json` conserva sus cuatro campos. El
  coste está acotado por SC-014.

### B · El evento *(M-2 a M-6, D-3)*

- **FR-006** *(unidad — M-2)*: cada línea `type = "token_usage_record"` con `payload.response_id` no vacío y partidas coherentes **DEBE**
  producir **un** evento. Ningún otro tipo de línea produce eventos.
- **FR-007** *(identidad — M-3)*: su `event_id` es el de `contracts/event-id-codex.md`, y **DEBE** reproducir sus vectores byte a byte.
- **FR-008** *(repetidas — P-7)*: dentro de una pasada, un `event_id` ya emitido no se vuelve a emitir y se cuenta como «repetida».
- **FR-009** *(sin identificador)*: un registro sin `response_id`, o con uno vacío, no se emite y se cuenta como «sin identificador».
- **FR-010** *(partidas — M-4)*: con las de `payload.usage`:
  - `tokens_input` = `input_tokens − cached_input_tokens − cache_write_input_tokens`;
  - `tokens_cache_creation` = `cache_write_input_tokens`;
  - `tokens_cache_read` = `cached_input_tokens`;
  - `tokens_output` = `output_tokens`, con el razonamiento dentro.

  `reasoning_output_tokens`, `total_tokens` y los acumulados *(`turn_token_usage`, `thread_token_usage`)* **no** se usan.
- **FR-011** *(modelo — M-5)*: el `model` del evento es el del `turn_context` con el mismo `turn_id`. Si no lo hay, es el **vigente**: el
  `model` de la última línea `turn_context` o `thread_settings_applied` anterior al registro en el mismo fichero. Si tampoco lo hay, el
  evento sale con `model` vacío y se cuenta como «sin modelo» *(P-4 (b))*.
- **FR-012** *(momento — M-6)*: `occurred_at` es el `timestamp` de la línea del registro.
- **FR-013** *(herramienta y coste — M-1, D-3)*: `tool = "codex"`, `cost_usd = 0` y `cost_available = false`, **siempre**. El lector no
  consulta `internal/pricing`.
- **FR-014** *(proyecto — M-6, P-5)*: `project_ref` = `Resolutor.Derivar(cwd, sal)` *(`internal/project/resolve.go:318`)*, con el `cwd` del
  `turn_context` del mismo turno, o el de `session_meta` si el turno no lo tiene.
- **FR-015** *(referencias — M-6, P-6)*: `session_ref` = `event.Ref(sal, payload.session_id)` *(`internal/event/event.go:40-46`)*.
  `machine_ref`, `dev_id`, `org_id` y `agent_version`, como en Claude Code *(`claudecode.go:119-129`)*.
- **FR-016** *(nada del proveedor viaja — E-1)*: `response_id`, `session_id`, `thread_id` y `turn_id` **NUNCA** se copian a ningún campo del
  evento, a la cola ni a nada que se transmita. `session_id` sólo viaja como `event.Ref` con sal. **Excepción declarada**: el nombre del
  fichero de sesión de Codex contiene el identificador del hilo, y la ruta es la clave de `state.json`. Esa ruta queda en el `state.json`
  **local**, como hoy las de Claude Code, y **no viaja**.

### C · Formatos *(D-2, P-3)*

- **FR-017** *(formato anterior)*: un fichero con `event_msg/token_count` con `info` no nulo y ningún `token_usage_record` produce **0**
  eventos y se cuenta como «formato anterior». Un fichero con registros es formato actual aunque tenga `token_count` antiguos *(Q-2 (a))*.
  Un fichero sin consumo no se cuenta.
- **FR-018** *(comprimidos)*: cada `.zst` bajo la raíz se cuenta como «comprimido» **una vez**, en la pasada en que aparece, y otra vez
  sólo si cambia.

### D · Lo que ve el usuario *(P-8, P-9)*

- **FR-019** *(resumen)*: con el lector activo, `--run` escribe **una** línea más en stderr, el texto aprobado, tras el resumen de Claude
  Code y su aviso *(`cmd/permea/main.go:346-350`)*. El demonio la escribe sólo si en ese ciclo hubo respuestas, ficheros en formato anterior
  o comprimidos, como `HayNovedades` *(`main.go:408`)*.
- **FR-020** *(`--scan` — P-8)*: un fichero cuya primera línea es `session_meta` se lee como Codex. Cada evento sale con la línea `evento:`
  de Codex, y el resumen con la línea de Codex.
- **FR-021** *(los textos de Claude Code no cambian)*: sin carpeta de Codex, la salida de `--run`, del demonio y de `--scan` es **byte a
  byte** la de la 0.4.0. Con carpeta, sólo se añade la línea de FR-019, y «N eventos encolados» *(`main.go:346`)* cuenta también los de
  Codex, con el mismo texto.

### E · Frontera, versión y documentos

- **FR-022** *(frontera — M-1)*: `git diff 7b8c77c` de `internal/event/`, `internal/ingest/eventid.go`, `eventid_test.go`,
  `boundary_test.go` y `specs/006-medicion-fiel/contracts/event-id.md`, **vacío**.
- **FR-023** *(versión — M-8)*: `0.5.0`. Claude Code se lee exactamente como en la 0.4.0.
- **FR-024** *(documentos)*: el README gana la sección de Codex, y el CHANGELOG la entrada `0.5.0`, con los textos aprobados.

### F · Durabilidad, errores y cuentas *(E-1, E-2)*

- **FR-025** *(orden de escritura)*: los eventos de Codex de un fichero se encolan **antes** de guardar el estado, como hoy hace
  `generate()` con Claude Code: `transport.Append` *(`cmd/permea/main.go:288,310`)* antes de `st.Save` *(`:317`)*. Si el proceso cae entre
  medias, la pasada siguiente los reencola con el mismo `event_id`, y la plataforma los descarta.
- **FR-026** *(Codex sin Claude Code)*: una instalación con carpeta de Codex y sin `~/.claude/projects` **DEBE** enviar los eventos de
  Codex. `state.FindLogs` ya tolera la raíz ausente *(`internal/state/scan.go:14-17`)*.
- **FR-027** *(la cuenta del resumen — Q-1 (a))*: en cada pasada, `respuestas = eventos + repetidas + sin identificador + incoherentes`.
  «Sin modelo» es un **subconjunto** de «eventos». Un registro con `cached_input_tokens + cache_write_input_tokens > input_tokens` no se
  emite y se cuenta como «incoherente». *(E-3)* También es «incoherente» un registro **sin `usage`** y uno con **alguna partida negativa**
  *(de las cuatro de FR-010)*. El orden de clasificación es sin identificador → incoherente → repetida → evento *(`DECIDÍ YO`)*.
- **FR-028** *(Codex no rompe a Claude Code — E-2)*: hoy cualquier error en `generate()` devuelve antes de `st.Save`
  *(`cmd/permea/main.go:304-306,317`)*. Un error al **leer** un fichero de Codex *(el stat, el prefijo o `Recorrer`)* **DEBE**:
  - escribir en stderr el texto aprobado `codex: fichero omitido: %v`;
  - **no** avanzar el offset de ese fichero;
  - **no** cortar la pasada;
  - **no** impedir guardar el estado de los demás ficheros, los de Claude Code incluidos.

  Un fallo al **encolar** *(`transport.Append`)* sigue siendo fatal, como hoy. Los errores de Claude Code no cambian *(FR-021)*.
- **FR-029** *(línea corrupta en Codex — E-2)*: una línea que pasa el filtro y no decodifica:
  - en la parte nueva *(del offset en adelante)*, se salta con el mensaje que ya existe, `skip (línea corrupta): …`
    *(`cmd/permea/main.go:282`)*;
  - en el prefijo releído, se ignora en silencio, para no repetir el aviso en cada ciclo.

  No cuenta en «respuestas» y no corta el fichero.

### Key Entities
- **Registro de Codex**: una línea `token_usage_record` con `response_id`, `turn_id`, `session_id`, la marca y las partidas de `usage`.
- **Contexto de fichero**: el `cwd` de `session_meta`, el modelo y el `cwd` del último `turn_context`, y el modelo vigente. Vive **sólo en
  memoria, durante la pasada** *(FR-005)*.

## Success Criteria

Las referencias son las del **contador independiente** *(descubrimiento §FASE 0 (f))*, sobre la copia congelada de huella
`984d483c12a15601`. Las medidas se toman en un sandbox con `CODEX_HOME=<copia congelada>`, sin enrolar y sin endpoint.

- **SC-001** *(contador, fichero a fichero)*: por fichero, el número de eventos `codex` y la suma de sus cuatro partidas son **iguales** al
  contador. Total: **7** eventos · `tokens_input` **21 779** · `tokens_cache_creation` **0** · `tokens_cache_read` **86 272** ·
  `tokens_output` **330**. F6: 4 · 15 076 · 0 · 51 200 · 88.
- **SC-002** *(formatos)*: F1–F4 dan **0** eventos y **4** ficheros «formato anterior». F5 da **0** eventos y no se cuenta.
- **SC-003** *(las cifras de Codex)*: `tokens_input + tokens_cache_creation + tokens_output` suma **2828** en F7 y **4117** en F8.
- **SC-004** *(sin coste)*: el 100 % de los eventos `codex` llevan `cost_available = false` y `cost_usd = 0`, y **ninguno** lleva
  `tool ≠ "codex"`.
- **SC-005** *(segunda pasada)*: una segunda `--run` sobre la copia da **0** eventos `codex`. Un fichero sintético cortado en la reanudación,
  leído en dos pasadas, da 1 + 3 eventos sin repetir ninguno.
- **SC-006** *(identidad — E-1)*: la función nueva reproduce los tres vectores del contrato. Un fixture lleva centinelas en `response_id`,
  `session_id` y `turn_id`, **y también en el nombre del fichero**. Tras una pasada, **ninguno** está en la cola ni en lo transmitido, y en
  `state.json` sólo aparece el del nombre, **dentro de su clave de ruta** *(la excepción de FR-016, a la vista)*.
- **SC-007** *(bifurcación, sintético)*: el mismo registro en dos ficheros da **un** evento y **1** «repetida» en la pasada.
- **SC-008** *(escritura de caché, sintético)*: entrada 100, caché 40 y escritura 60 → `tokens_input` 0, `tokens_cache_creation` 60 y
  `tokens_cache_read` 40. Con `tokens_input` = entrada − caché, el test cae.
- **SC-009** *(modelo)*: en F6, los 4 eventos llevan `gpt-6-luna`, el de la compactación por la regla del vigente. Un fixture sin ningún
  modelo da 1 evento con `model` vacío y 1 «sin modelo».
- **SC-010** *(contexto entre pasadas)*: un fichero leído en dos pasadas, cortado después de su `turn_context`, da en la segunda el mismo
  modelo, `project_ref` y `session_ref` que una lectura de una vez.
- **SC-011** *(sin Codex — E-2)*: sin raíz de Codex, y también con `CODEX_HOME=""`, la salida de `--run` sobre un fixture de Claude Code es
  **byte a byte** igual a un **fichero de referencia**. Se genera en la Fase 0 de B5 con el código del commit anterior, **antes** de tocar
  `main.go`, y se guarda en `testdata/` con su md5 transcrito.
- **SC-012** *(puertas)*: frontera con 0 bytes de diff; los **494** tests actuales, verdes y **sin tocar** *(`git diff 7b8c77c --stat --
  '*_test.go'` sólo con ficheros nuevos)*; `golangci-lint run` → 0.
- **SC-013** *(textos)*: los textos nuevos, byte a byte iguales a los aprobados, también sobre el binario publicado.
- **SC-014** *(coste del contexto — E-1)*: sobre un fichero sintético de **≥ 100 MB**, generado en un temporal y nunca en el repo, una
  pasada que encuentra **1** registro nuevo tarda **≤ 3 s**: el 5 % del ciclo por defecto del demonio, 60 s *(`internal/config/config.go:47`)*.
  Se mide tres veces *(quickstart §Coste)*. **Si no se cumple, cambia el mecanismo; el tope no se mueve.**
- **SC-015** *(orden de escritura — E-1, E-2)*: con un forzado en el que `state.Load` pasa y `Save` falla *(`internal/state/state.go:32-48,52-75`;
  se elige y se declara en la Fase 0 de B5, plan D-008-P10)*, la cola contiene los eventos y la pasada devuelve error. Una segunda pasada
  con el estado sano los reencola con los **mismos** `event_id`.
- **SC-016** *(Codex sin Claude Code — E-1)*: con un fixture de Codex y sin raíz de Claude Code, `--run` encola los eventos de Codex y
  termina con código 0.
- **SC-017** *(la cuenta del resumen — E-1)*: sobre un fixture con una respuesta de cada clase *(evento, repetida, sin identificador,
  incoherente y evento sin modelo)*, el resumen cumple la identidad de FR-027 y da `respuestas 5 · eventos 2 · repetidas 1 · sin
  identificador 1 · incoherentes 1 · sin modelo 1`.
- **SC-018** *(Codex no rompe a Claude Code — E-2)*: con un fichero de Codex ilegible, junto a un fixture de Claude Code y otro de Codex sano:
  - salen los eventos de los dos sanos;
  - stderr lleva `codex: fichero omitido: …`;
  - `state.json` se guarda;
  - una segunda pasada **no** reencola nada de Claude Code;
  - cuando el fichero vuelve a ser legible, sus registros salen, porque su offset no avanzó.

## Censo de tests — **autorización para la fase de tareas**

**Ninguno de los existentes.** Los tests nuevos van en ficheros nuevos de `internal/ingest/`, `internal/config/` y `cmd/permea/`, con
fixtures **sintéticos** nuevos en `testdata/`.

**Censo de código fuera de los `_test.go`** *(M-9)*: `internal/testutil/sandbox.go` gana `t.Setenv("CODEX_HOME", "")`.

**Medido y FUERA del censo**:
- `internal/ingest/boundary_test.go` y `eventid_test.go` *(M-1)*;
- `internal/state/state_test.go` y `retener_test.go`: `Recorrer` se usa tal cual;
- `cmd/permea/main_test.go`, `retencion_test.go` y `coste_test.go`: construyen `agent{…}` a mano *(`main_test.go:441`,
  `retencion_test.go:57`)*. Con la raíz de Codex resuelta en `setup()`, no la tienen y no leen Codex *(plan D-008-P5)*;
- `internal/testutil/sandbox_test.go:20`: comprueba las tres variables de antes, y la cuarta no cambia su veredicto.

## Fuera de alcance

| # | Qué | Por qué |
|---|---|---|
| **N-1** | El formato anterior *(0.101.0)* | D-2 |
| **N-2** | El coste y las tarifas de OpenAI | D-3: van en la plataforma, después |
| **N-3** | Los niveles de servicio *(Flex, Fast, Batch)* y el contexto largo *(> 272K)* | Su tarifa depende de datos que el registro no trae |
| **N-4** | Los límites de uso *(`rate_limits`)* | No son consumo |
| **N-5** | Las bases SQLite de Codex | Su total se extrae del JSONL |
| **N-6** | Gemini CLI y OpenCode | Lectores aparte |
| **N-7** | Leer desde WSL las sesiones de Codex de Windows | M-7 lee el directorio personal propio |
| **N-8** | Un interruptor para desactivar Codex | Q-5 (a) |

## Textos aprobados *(P-9 ✅ — literales)*

**Resumen de Codex** *(stderr, una línea; en `--run`, tras el resumen de Claude Code; en el demonio, sólo con novedades)*:
```
codex: respuestas %d · eventos %d · repetidas %d · sin identificador %d · incoherentes %d · sin modelo %d · ficheros en formato anterior %d · ficheros comprimidos %d
```
**Fichero de Codex omitido** *(stderr, una línea por fichero; E-2, aprobado por el dueño el 2026-10-07 salvo aviso en contra)*:
```
codex: fichero omitido: %v
```
**Línea de `--scan` para Codex** *(stdout)*:
```
evento: tool=%s model=%s in=%d out=%d cw=%d cr=%d cost=$%.4f cost_avail=%t project_ref=%s event_id=%s
```
**README** *(sección nueva, tras la de Claude Code)*:
```
### Codex CLI

Si existe `~/.codex/sessions` (o `$CODEX_HOME/sessions`), el agente lee también el consumo de Codex CLI.
Cada respuesta del modelo es un evento con `tool = codex`, sus tokens y su modelo. El coste no lo calcula
el agente: los eventos de Codex salen con `cost_available = false`, y el coste lo pone la plataforma
cuando tenga las tarifas de esos modelos. Hace falta Codex 0.153.0 o posterior; las sesiones anteriores
se cuentan como «formato anterior» y no se envían. Si la carpeta no existe, no cambia nada.
```
**CHANGELOG `0.5.0`** *(encabezado con `PENDIENTE` hasta la etiqueta, como en 007)*:
```
## 0.5.0 — PENDIENTE

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
```

## Preguntas abiertas — **ninguna** *(ratificación del 2026-10-07)*

| # | Pregunta | Respuesta |
|---|---|---|
| **Q-1** ✅ **(a)** | ¿Un registro con caché + escritura > entrada? *(0 casos)* | No se emite y se cuenta como «incoherente» *(FR-027)* |
| **Q-2** ✅ **(a)** | ¿Un fichero con los dos formatos? | Tiene registros: se emiten, y sus `token_count` antiguos se ignoran *(FR-017)* |
| **Q-3** ✅ **(a)** | ¿Fija el sandbox de test `CODEX_HOME`? | Sí *(M-9)* |
| **Q-4** ✅ **(a)** | ¿Qué modelo lleva la compactación tras un cambio a la baja? | El vigente, hasta tener muestra *(FR-011)* |
| **Q-5** ✅ **(a)** | ¿Un interruptor para desactivar Codex? | No, en la 0.5.0 *(N-8)* |
| **Q-6** ✅ **(a)** | ¿Coincide `os.UserHomeDir()` con el directorio de Codex en Windows? | Se asume que sí, y se comprueba en W1 |

## Assumptions
1. **En WSL no hay Codex**: `~/.codex` no existe allí. Las sesiones del dueño están en **Windows**, y las leerá el agente de Windows.
2. **El formato de `token_usage_record` es estable** desde la 0.153.0: los mismos campos en 0.153.0 y en 0.160.1.
3. **`cache_write_input_tokens` va dentro de `input_tokens`**, según la documentación oficial *(FASE 0 (c))*. En la copia vale 0 siempre.
4. **Codex no borra sesiones**: no hay borrado automático, y la compresión a `.zst` está desactivada por defecto.

## Dependencias
- **El parche de tarifas de la plataforma**, **posterior** a este lector *(D-3)*. Da fila sólo a los identificadores **vistos en eventos
  reales**: hoy, `gpt-6-luna`. `gpt-5.3-codex` no ha producido ningún evento *(sus turnos de F5 y F6 fallaron)*. Hasta entonces, los eventos
  de Codex llegan con el coste ciego y señalizado *(E-1)*.
- **006**: el esquema del `event_id`, sin tocarlo. **007**: `Recorrer`, que se usa tal cual.

## Registro de enmiendas

| # | Fecha | Qué cambia | Por qué |
|---|---|---|---|
| **Ratificación** | 2026-10-07 12:16 | **El dueño ratifica**: P-1 (a) y Q-5 (a), sin interruptor; P-2 a P-8 según la recomendación, con lo que precisa E-1; Q-1 a Q-6, (a). P-9: el resumen de Codex pasa a **etiqueta y número**, y la línea de `--scan`, el README y el CHANGELOG se aprueban tal cual. ✋ → ✅, §Preguntas abiertas a cero, y §Textos propuestos pasa a «Textos aprobados» | Ratificación del dueño |
| **E-1** | 2026-10-07 | **Enmienda del orquestador.** **(1)** P-2 y FR-005 **por resultado**, con el mecanismo en el plan y **SC-014**, de coste: ≤ 3 s sobre ≥ 100 MB. **(2)** FR-016, SC-006 y el contrato: **la excepción de la ruta** en `state.json`. **(3)** Nuevos **FR-025** y **SC-015**, del orden de escritura; **FR-026** y **SC-016**, de Codex sin Claude Code. **(4)** **FR-027** y **SC-017**: la cuenta del resumen, con «incoherentes». **(5)** Dependencias: sólo identificadores vistos. **(6)** Edge case de los subagentes. **(7)** **M-9**: Q-3 (a) como decisión de método, `sandbox.go` en el censo de código | Que P-2 se juzgue por su efecto y su coste; que la privacidad declare lo que queda en local; y que no falten la durabilidad, la instalación sin Claude Code ni la cuenta del resumen |
| **E-2** | 2026-10-07 | **Enmienda del orquestador.** **(1)** FR-002: la activación se evalúa **en cada pasada**. **(2)** Nuevos **FR-028** y **SC-018**: un error de lectura de un fichero de Codex se omite con `codex: fichero omitido: %v` *(a §Textos aprobados)*, sin avanzar su offset ni cortar la pasada; encolar sigue siendo fatal. **(3)** Nuevo **FR-029**: la línea corrupta, con aviso sólo en la parte nueva. **(4)** SC-015: un forzado en el que `Load` pasa y `Save` falla, porque «`state.json` como directorio» hace fallar `Load` antes de encolar. **(5)** SC-011: un fichero de referencia generado con el commit anterior | El demonio vive días; un fichero de Codex no debe tumbar a Claude Code; y dos SC necesitaban un instrumento que sí mida |
| **E-3** | 2026-10-07 | **Enmienda del orquestador.** «Incoherente» *(FR-027, Q-1)* cubre también un `token_usage_record` **sin `usage`** y uno con **alguna partida negativa**: no se emite y se cuenta. El rojo (6) gana una hoja por caso *(caché + escritura > entrada · sin `usage` · partida negativa)*, y m6 debe tumbar las tres | Que un registro roto no viaje como evento con ceros, ni con negativos que resten en la plataforma |
