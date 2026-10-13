# 010 · «Lector de OpenCode» — Especificación

**Feature Branch**: `010-lector-opencode` · **Created**: 2026-10-11 · **Status**: **Cerrada (2026-10-13)** *(ratificada el 2026-10-11: P-1 a P-12; P-3, reescrita
por el dueño como frontera por rowid; Q-1 a Q-4 siguen abiertas; Q-5, cerrada; E-1 el 11-10)* · **Cabeza de partida**: `4c6cafc` *(sobre `6a952bf`, cierre de la 0.6.0)*
**Input**:
- las decisiones del dueño D-1 y D-2 *(2026-10-09)* y D-3 *(2026-10-11)*;
- el descubrimiento del 09-10 *(`soporte/descubrimiento.md`, D-1…D-10, sobre la **copia congelada** de una sesión de OpenCode 2.0.26)*;
- la prueba de SQLite y el ensayo en Windows del 11-10 *(`soporte/prueba-sqlite.md`)*.

> ⛔ Toda afirmación sobre nuestro código lleva `fichero:línea`, medida sobre `4c6cafc`. Las de OpenCode citan el descubrimiento *(«desc. D-n»)* o la
> prueba. Lo que añade la spec sin que nadie lo dijera va marcado **`DECIDÍ YO`**. Para no confundirse con las D-n del descubrimiento, las
> decisiones del dueño se citan como **D-1…D-3** y las del descubrimiento como **desc. D-n**.

## Contexto — el agente lee Claude Code, Codex y Gemini CLI

1. **Tres lectores hoy**, en `generate()`, antes del **único** `st.Save` *(`cmd/permea/main.go:368`)*: Claude Code, Codex *(`main.go:342-353`)* y Gemini,
   tras Codex *(`main.go:355-366`)*.
2. **OpenCode 2.x guarda su consumo en una base SQLite**: `<XDG_DATA_HOME | ~/.local/share>/opencode/opencode.db`, igual en los tres SO *(desc. D-9)*.
   - Cada llamada al modelo es **una fila `assistant`** de `session_message`, con tokens, coste y modelo en `data` *(JSON; desc. D-2)*.
   - La fila **se reescribe en su sitio** hasta completarse *(desc. D-5)*.
   - El coste de OpenCode coincide con la tarifa de Permea, salvo el redondeo, si la salida es `output + reasoning` *(desc. D-4)*: con `google` y con
     `anthropic` *(`soporte/medicion-anthropic.md`)*.
3. **Leer SQLite sin cgo y sin dependencias es posible.** El lector propio de la prueba *(B)* añade **≈ 60 KB** y **nunca escribe**; es exacto con WAL vivo,
   WAL roto y desbordamiento; y en Windows dio lo mismo que `modernc.org/sqlite` en **176 de 176** lecturas *(prueba §Ensayo)*.
4. **La plataforma ya acepta el evento tal cual.** `tool` es texto libre; un modelo sin tarifa llega «ciego y señalizado»; y un `event_id` repetido se
   descarta *(009 §Contexto 4)*.

## Decisiones del dueño *(ratificadas; no se discuten aquí)*

| # | Decisión |
|---|---|
| **D-1** · 09-10 | **Sin coste**: `cost_usd = 0` y `cost_available = false`, y la plataforma lo repara, como en Codex y Gemini. El `$.cost` de OpenCode **sólo** como contraste en el dry-run *(`--scan`)*; nunca viaja |
| **D-2** · 09-10 | **Límite declarado**: la llamada del título, y cualquier otra que no esté en `session_message`, no se emite. En la medición, **0,37 %** del coste *(538 de entrada y 5 de salida; desc. D-3, D-8)*. **No hay evento residuo** |
| **D-3** · 11-10 | **SQLite con el lector propio** de la prueba *(B)*:<br>• sólo librería estándar, `go.mod` **sin `require`**, y **nunca** escribe en la carpeta de OpenCode;<br>• WAL hasta el último commit válido, con la comprobación de estabilidad.<br>Antes de entrar en `internal/` lleva:<br>• tests con bases **sintéticas** generadas por `sqlite3`, en `testdata`;<br>• **fuzzing**: una entrada corrupta no cuelga ni revienta;<br>• las mutaciones **mB1–mB4**, rehechas sobre el código definitivo |

## Decisiones de método *(heredadas de 008 y 009)*

| # | Decisión |
|---|---|
| **M-1** | `tool = "opencode"`. El evento no cambia: mismos campos y `SchemaVersion = 1` *(`internal/event/event.go:12-35`)*. La frontera de FR-030, con **0 bytes** de diff |
| **M-2** | Versión **`0.7.0`**. Claude Code, Codex y Gemini se leen exactamente como en la 0.6.0: los **671** tests actuales *(`go test -count=1 -v ./...` en `4c6cafc`: 671 «--- PASS», 0 FAIL, 0 SKIP)* siguen verdes **sin tocarlos** |
| **M-3** | `internal/testutil/sandbox.go` fija también `XDG_DATA_HOME` y `OPENCODE_DB` *(vacías)*, junto a `CODEX_HOME` y `GEMINI_CLI_HOME` *(`sandbox.go:63`, `:65`)*. Sin ellas, un test de proceso leería la base real del desarrollador |
| **M-4** | Los errores son **por base**, como por fichero en 008 y 009 *(`cmd/permea/gemini.go:21-23`)*. Encolar sigue siendo fatal |
| **M-5** | Son la **tercera y la cuarta** lecturas de entorno en producción, tras `CODEX_HOME` y `GEMINI_CLI_HOME` *(`internal/config/codex.go:19`, `gemini.go:20`)* |

## Propuestas — **ratificadas** por el dueño el 2026-10-11 *(P-3, reescrita: §P-3)*

| # | Propuesta | Alternativas | Recomendación |
|---|---|---|---|
| **P-1** ✅ **(a)** | **Fuente**: las filas de `session_message` con `type = 'assistant'` y `$.time.completed` presente y no nulo. Una fila = una llamada = **como mucho un** evento | (b) también las no completadas con tokens | **(a)**: la fila se reescribe hasta completarse *(desc. D-5)*. En vivo, la regla excluyó una respuesta a medio escribir y la tomó 2 s después *(prueba §Ensayo 3)* |
| **P-2** ✅ **(a)** | **Partidas**:<br>• `tokens_input` = `tokens.input`;<br>• `tokens_output` = `tokens.output + tokens.reasoning`;<br>• `tokens_cache_read` = `tokens.cache.read`;<br>• `tokens_cache_creation` = `tokens.cache.write`.<br>Una partida ausente vale 0 | (b) la salida sin `reasoning` | **(a)**: `input` ya viene sin la caché, y el coste sólo cuadra así; sin el razonamiento, Permea daría un 3,7 % menos *(desc. D-3, D-4)*. Es la regla de Gemini *(009 D-2)*. **Verificada con `google` y `anthropic`** *(Q-5 cerrada, SC-026)* |
| **P-3** ✅ *(del dueño)* | **Estado entre pasadas: frontera por rowid** *(§P-3)* | *(rastro)* (b) la marca por `time_updated` con corte de 5 s, del borrador · (c) guardar los `id` emitidos · (d) reemitirlo todo | **Decisión del dueño, 11-10**: el coste por pasada sigue a lo nuevo, no a la tabla, y una renumeración nunca pierde filas |
| **P-4** ✅ **(a)** | **`event_id` sólo con el `id` de la fila** *(`msg_` + 26; desc. D-5)*: `["permea/event_id/v1", "opencode", "respuesta", <id>]`, en un fichero y una función nuevos, con un contrato nuevo, `contracts/event-id-opencode.md` | (b) `session_id` + `id` | **(a)**, como 009 P-2: el `id` es único por llamada *(14 de 14)* y estable mientras la fila se reescribe. El tipo reutiliza `tipoRespuesta`, como en 008 y 009 |
| **P-5** | **Modelo**: `$.model.id` tal cual. `providerID` y `variant` no viajan. Sin `model.id` *(ausente, vacío o no textual)*: el evento sale con `model` vacío y se cuenta como «sin modelo» | (b) `providerID/model.id` · (c) normalizarlo en el agente | **(a)**. Con `google`, `model.id` es ya la clave de Permea *(desc. D-6)*. Con otros proveedores *(p. ej. pasarelas o nubes que añaden prefijos o sufijos)*, el `id` puede no casar con el catálogo: el evento llega «ciego y señalizado», y en la plataforma **queda sin reparar y parcial**, como hoy. Un alias en la plataforma sería otra feature *(N-10)*. El agente nunca traduce modelos *(como 008 y 009)* |
| **P-6** ✅ **(a)** | **Proyecto y sesión**:<br>• `project_ref` = `Resolutor.Derivar` *(`internal/project/resolve.go:318`)* de `session_v2.directory`;<br>• `session_ref` = `event.Ref(sal, session_id)` *(`internal/event/event.go:40-46`)*.<br>Un subagente *(`parent_id`)* lleva su propia sesión y su propio `directory`. Si la fila de `session_v2` falta, las dos referencias salen vacías | (b) el `session_ref` del padre para los subagentes | **(a)**: es lo que la base dice, sin inferir. No hay muestra de subagentes *(desc. D-7)* |
| **P-7** ✅ **(a)** `DECIDÍ YO` | **Momento**: `occurred_at` = `$.time.completed` *(ms epoch, en UTC)* | (b) `$.time.created` | **(a)**: P-1 garantiza que está, y es cuando la llamada termina y se factura. Una llamada que cruza la medianoche cae en el día en que acaba |
| **P-8** ✅ **(a)** | **Ubicación**, resuelta una vez por proceso y **con la misma regla en los tres SO** *(desc. D-9)*:<br>• `OPENCODE_DB` **absoluta**: sólo esa base;<br>• si no: `<datos>/opencode/`, con `<datos>` = `XDG_DATA_HOME` si no es vacía y, si no, `<hogar>/.local/share`;<br>• bases: `opencode.db` y cada `opencode-<canal>.db` *(canal en `[A-Za-z0-9._-]`)*, **salvo `opencode-next.db`** *(Q-1)* | (b) sólo `opencode.db` | **(a)**: las versiones de canal escriben en `opencode-<canal>.db`. `OPENCODE_DB` vacía, `":memory:"` o relativa cae en la regla general `DECIDÍ YO`: vacía es «no definida» en la CLI *(`\|\|` de JS)*, y no se midió cómo resuelve una relativa |
| **P-9** ✅ **(a)** | **1.x = «formato anterior»**: una base sin `session_message` y con `message` y `part` se **cuenta**, no se lee ni produce eventos. `storage/` *(JSON de 1.x)* no se lee ni se cuenta | (b) leer 1.x | **(a)**, como el `.json` de Gemini *(009 D-3)*. No hay muestra 1.x *(desc. D-9)*, y las migraciones de 2.x pasan `storage/` a la base |
| **P-10** ✅ **(a)** `DECIDÍ YO` | **Activación**: el lector está activo **en una pasada** si existe, en esa pasada, `<datos>/opencode/` *(o el fichero de `OPENCODE_DB`)*. Se comprueba en cada `generate()`. No hay interruptor | (b) un interruptor | **(a)**, como 008 P-1 y 009 P-1 |
| **P-11** ✅ **(a)** `DECIDÍ YO` | **`--scan <base>`**: una base se reconoce por sus 16 primeros bytes *(`SQLite format 3\0`)*, **antes** que Codex y Gemini. La comprobación de hoy lee la primera línea como texto *(`main.go:537-551`)* | (b) `--scan-opencode` | **(a)**, como 008 P-8 y 009 P-8. Va antes porque una base binaria no tiene «primera línea» |
| **P-12** ✅ **(a)** | **Textos**: los de §Textos propuestos. El README *(«### OpenCode»)* y el CHANGELOG `0.7.0`, en el encargo de documentos, como 009 P-11 | corregirlos | — |

### P-3 ✅ · Frontera por rowid *(decisión del dueño, 2026-10-11; el mecanismo exacto, al plan)*

**Por qué rowid.** `session_message` tiene rowid implícito: su `id` es texto *(desc. D-1)*. **SQLite serializa a los escritores**, también entre procesos:
una fila confirmada después tiene un rowid mayor. Dos excepciones, que la frontera detecta: un rowid **reutilizado** tras borrar la fila de rowid máximo,
y una **renumeración**, que `VACUUM` **puede** hacer *(E-2)*. Medido el 2026-10-12 con SQLite 3.45.1, conserva los rowid con `id text PRIMARY KEY` y los
renumera sin índices; la defensa *(huella de H y pasada completa)* sigue para otras versiones y esquemas, y `vacuum/`, sin índices, prueba el caso.

**Estado por base** *(FR-004)*: **H**, el mayor rowid visto de `session_message`; la **huella** del `id` de la fila H *(con sal, no el `id`, para que
`state.json` no guarde identificadores del proveedor; `DECIDÍ YO`)*; y la **lista en curso**, con el rowid y la huella de cada fila `assistant` vista sin
completar *(pequeña)*.

**Una pasada**:
1. **Comprueba la frontera** bajando por el árbol hasta H. Si la fila no existe, o su huella no es la guardada, hace una **pasada completa** y cuenta
   «frontera reiniciada».
2. **Consulta por rowid** cada fila de la lista, bajando por el árbol, sin barrido. Según lo que encuentre:
   - **completada**: se clasifica *(FR-021)* y sale de la lista;
   - **desaparecida** *(sesión borrada)*: sale de la lista, sin contarse;
   - **huella distinta**: pasada completa;
   - **sigue en curso** y su `$.time.created` *(o, si falta, la columna `time_created`)* tiene más de **24 h**: «abandonada». Sale de la lista y **nunca**
     se emite.

   La consulta va **antes** que el abandono: una fila que se completó mientras el agente estaba parado sale como llamada.
3. **Recorre desde el primer rowid > H** hasta el final: una completada se clasifica; una en curso entra en la lista *(o es «abandonada», si ya tiene
   más de 24 h)*; otro `type`, nada. H pasa a ser el mayor rowid visto.

**Pasada completa**: la primera pasada *(entrada sin el campo nuevo)* o una frontera reiniciada. Recorre la tabla entera, emite todas las completadas y
rehace la lista; las ya emitidas vuelven a salir con el mismo `event_id` y la plataforma las descarta: **nunca se pierde una fila completada**.

**Coste**: proporcional a lo nuevo, al tamaño de la lista y al número de sesiones *(E-1)*, no al de la tabla *(SC-017)*. **Nunca hay un barrido completo** fuera de la pasada completa.

## User Scenarios & Testing

### Historia 1 — El consumo de OpenCode llega a la plataforma *(P1)*
1. **Dado** una base de OpenCode 2.x con dos llamadas completadas, **cuando** corre `--run`, **entonces** salen **dos** eventos `tool = opencode`, con sus
   partidas según P-2, `cost_available = false` y el `model.id` de cada fila.
2. **Dada** la misma base en una segunda pasada sin cambios, **entonces** no sale **ningún** evento.

### Historia 2 — Una respuesta a medio escribir no se cuenta dos veces ni a medias *(P1)*
1. **Dada** una fila `assistant` sin `time.completed` *(OpenCode escribiendo)*, **entonces** sale 0 eventos y la línea dice «en curso 1».
2. **Cuando** se completa, **entonces** la pasada siguiente emite **un** evento, con las partidas finales.
3. **Si** no se completa en 24 h, **entonces** se cuenta como «abandonada» y deja de consultarse.

### Historia 3 — El agente nunca toca la carpeta de OpenCode *(P1)*
1. **Dada** una base viva en modo WAL, **cuando** corre una pasada, **entonces** `opencode.db`, `-wal` y `-shm` tienen la **misma md5** antes y después, y no
   aparece ningún fichero nuevo.

### Historia 4 — Quien no usa OpenCode no nota nada *(P1)*
1. **Dada** una instalación sin `<datos>/opencode/` ni `OPENCODE_DB`, **entonces** la salida de `--run` es **byte a byte** la de la 0.6.0.

### Historia 5 — OpenCode sin las otras herramientas *(P2)*
1. **Dada** sólo la carpeta de OpenCode, **entonces** sus eventos salen igual, y `--run` termina con 0.

### Edge Cases
- **WAL reiniciado por un checkpoint** *(ensayo 1: de 1 490 a 33 frames)*: los frames viejos tienen otras sales y la cadena se corta. Se lee bien.
- **La base cambia durante la lectura** *(otra transacción, un checkpoint)*: se repite hasta 5 veces. Si sigue cambiando, la base se omite en esa
  pasada *(FR-008)* y su frontera no avanza.
- **Transacción a medias** *(frames sin commit, o frames reescritos en su sitio con sumas viejas)*: no se ve nunca *(prueba, mB4)*.
- **WAL roto** *(un frame con la suma mal)*: la cadena se corta allí, y lo de después no existe, como en SQLite *(prueba, R)*.
- **Arranque de OpenCode**: lecturas de hasta 3,6 s en el ensayo 2. Entran en el ciclo de 60 s *(`internal/config/config.go:47`)*.
- **Filas `user` e `idle`**, y cualquier otro `type`: no son llamadas. No se emiten ni se cuentan.
- **La llamada del título** y las que no están en `session_message`: D-2. Sólo las muestra el contraste de `--scan`.
- **Sesión borrada en OpenCode** *(`ON DELETE CASCADE`)*: sus filas desaparecen. Lo emitido no se toca, y lo no leído se pierde *(§Límites)*. Si
  se lleva la fila H, la frontera se reinicia *(P-3)*.
- **`VACUUM`** *(**puede** renumerar los rowid implícitos)*: si la fila H cambia de `id`, la frontera se reinicia sin perder nada *(P-3)*.
- **Varias bases** *(canales)*: cada una tiene su entrada en `state.json` y su frontera. Una misma llamada en dos bases da el mismo `event_id`, y la
  plataforma la descarta.
- **La tabla `event`** *(vacía en 2.0.26; desc. D-5)*: no se lee. Si una versión la llenara, no se cuenta dos veces.
- **`OPENCODE_DB` apunta a algo que no existe**, o `<datos>/opencode/` no existe: el agente no dice nada, no falla y no escribe la línea de OpenCode.
- **Columnas de más** en las tablas: se ignoran. **Columnas que faltan** *(`id`, `session_id`, `type`, `time_updated` o `data`; `id` o `directory` en
  `session_v2`)*: «esquema desconocido» *(Q-3)*.
- **`data` que no es JSON, o tokens no numéricos o negativos**: la fila es «incoherente» *(FR-021)*.
- **El proceso cae entre encolar y guardar**: la pasada siguiente reencola los mismos `event_id`, y la plataforma los descarta.

## Requirements

### A · Dónde y cuándo lee *(P-8, P-10, P-3)*

- **FR-001** *(raíz)*: según P-8. `<hogar>` es el `os.UserHomeDir()` de las otras raíces. **No** se usa `AppData` ni `Library`.
- **FR-002** *(activación)*: según P-10. Si no está activo, el lector no escribe nada, no devuelve error y no cambia ningún texto.
- **FR-003** *(bases)*: se leen las bases de P-8, y **ningún otro fichero**: ni `-wal` ni `-shm` como bases, ni `opencode-next.db`, `log/` o `storage/`.
  De cada base sólo se leen `sqlite_schema`, `session_message` y `session_v2`.
- **FR-004** *(estado)*: cada base tiene una entrada en el mismo `state.json`, con la ruta como clave y sus **cuatro** campos *(`internal/state/state.go:16-21`)*.
  - `Offset` guarda H, y `Size` y `ModTime` son los de la base; **el `mtime` no decide nada**.
  - La huella de H y la lista en curso van en **un campo nuevo, opcional y compatible hacia atrás** *(`omitempty`)*:
    - una entrada sin él es una primera pasada;
    - las entradas de las otras herramientas no lo llevan;
    - un 0.6.0 lo ignora al leer y lo pierde al guardar, y al volver a 0.7.0 hay una pasada completa, sin pérdida.
- **FR-005** *(entre pasadas — P-3)*:
  - una fila completada **DEBE** salir **una vez**, salvo tras una frontera reiniciada, que la reemite con el mismo `event_id`;
  - **NUNCA** se pierde, salvo si se abandona *(24 h en curso)*;
  - el coste de una pasada **DEBE** ser proporcional a lo nuevo, a la lista en curso **y al número de sesiones** *(la lectura de `session_v2`)*, no al
    tamaño de `session_message` *(SC-017; E-1)*.

### B · Cómo lee la base *(D-3)*

- **FR-006** *(sólo lectura)*: el lector **NUNCA** escribe, crea, bloquea ni borra nada en la carpeta de OpenCode.
  - Abre la base en sólo lectura y el `-wal` con una lectura entera; **no abre el `-shm`**.
  - Ni SQLite, ni cgo, ni un `sqlite3` externo.
- **FR-007** *(WAL)*: si la cabecera dice modo WAL, valida en cadena la cabecera, las sales y las sumas.
  - Toma **sólo los frames hasta el último commit válido**, y el más reciente por página.
  - Un `-wal` inválido o ausente se ignora, y queda la base sola.
- **FR-008** *(estabilidad)*: tras leer, comprueba que el `-wal` *(tamaño y cabecera)* y el tamaño de la base no han cambiado; si cambiaron, repite, hasta
  **5** veces. Si no se estabiliza, es un error de la base *(FR-029)*.
- **FR-009** *(formato)*: lee tablas con rowid, texto UTF-8 y páginas de 512 a 65 536 B, con bytes reservados y desbordamiento.
  - Una base `WITHOUT ROWID` en las tablas que lee, o en UTF-16, es «esquema desconocido» *(Q-3)*.
  - Una estructura imposible *(ciclos, punteros fuera, registros cortos)* es un error de la base *(FR-029)*.
- **FR-010** *(robustez)*: ninguna entrada, por corrupta que esté, puede provocar un pánico, un bucle sin fin o una reserva de memoria mayor que el fichero
  *(SC-009)*.
- **FR-011** *(dependencias — D-3)*: `go.mod` **sin `require`** y sin `go.sum`.

### C · El evento *(P-1, P-2, P-4 a P-7, D-1)*

- **FR-012** *(unidad — P-1)*: es **llamada** cada fila `assistant` con `$.time.completed` presente y no nulo que la pasada clasifica *(P-3: nueva, de
  la lista o de una pasada completa)*. Una fila `assistant` nueva cuyo `data` no es un objeto JSON también es llamada, «incoherente». Cada llamada se clasifica *(FR-021)* y produce
  **como mucho un** evento. Una fila `assistant` sin `time.completed` está «en curso».
- **FR-013** *(identidad — P-4)*: el `event_id` es el de `contracts/event-id-opencode.md`, y **DEBE** reproducir sus vectores byte a byte.
- **FR-014** *(sin identificador)*: una llamada cuyo `id` está vacío o no es texto no se emite, y se cuenta como «sin identificador».
- **FR-015** *(partidas — P-2)*: según P-2.
- **FR-016** *(modelo — P-5)*: según P-5.
- **FR-017** *(momento — P-7)*: `occurred_at` = `$.time.completed`, convertido de ms a UTC.
- **FR-018** *(herramienta y coste — D-1)*: `tool = "opencode"`, `cost_usd = 0` y `cost_available = false`, **siempre**. El lector no consulta
  `internal/pricing`, y `$.cost` **nunca** entra en el evento.
- **FR-019** *(proyecto y sesión — P-6)*: según P-6. `machine_ref`, `dev_id`, `org_id` y `agent_version`, como en Claude Code.
- **FR-020** *(nada del proveedor viaja)*: **NUNCA** se copian a ningún campo del evento, a la cola ni a nada que se transmita:
  - el `id` de la fila, `session_id`, `directory`, `path`, `title`, `slug` o `project_id`;
  - ningún texto de `data`.

  `session_id` y `directory` sólo viajan derivados con sal. La ruta de la base es la clave de `state.json`, y queda **en local**. Del `id`, `state.json` sólo
  guarda huellas con sal *(P-3)*.

### D · Clasificación, formatos y lo que ve el usuario *(P-9, P-11, P-12)*

- **FR-021** *(la cuenta del resumen)*: en cada pasada, `llamadas = eventos + sin identificador + incoherentes`. «Sin modelo» es un **subconjunto** de
  «eventos». Se cuentan aparte «en curso» *(la lista al acabar)*, «abandonadas» y «fronteras reiniciadas» *(P-3)*.
  - **Incoherente**: `data` no es un objeto JSON; `tokens` no es un objeto; alguna partida no es numérica, no es entera o es negativa; o
    `time.completed` no es numérico.
  - **Orden**: sin identificador → incoherente → evento.
- **FR-022** *(formato anterior — P-9)*: una base 1.x se cuenta como «base en formato anterior» **una vez**, en la pasada en que aparece, y otra vez sólo si
  cambia *(como los `.json` de 009 FR-018)*.
- **FR-023** *(esquema desconocido — Q-3)*: una base de 2.x que el lector no reconoce *(FR-009, o columnas que faltan)* se cuenta como «base con esquema
  desconocido», con la regla de FR-022. **No** se lee, **no** es un error y **no** corta nada.
- **FR-024** *(resumen)*: con el lector activo, `--run` escribe **una** línea más en stderr *(§Textos)*, tras la de Gemini *(`main.go:440`)*, o tras la
  anterior que haya. El demonio la escribe sólo si en ese ciclo hubo llamadas, abandonadas, fronteras reiniciadas o bases contadas *(como `HayNovedades`, `main.go:505-507`)*: una fila que
  se quede «en curso» no la repite en cada ciclo.
- **FR-025** *(`--scan` — P-11, D-1, D-2)*:
  - una base se lee entera, con su `-wal` si está al lado;
  - escribe una línea `evento:` por llamada *(la de Codex y Gemini, `main.go:637`)*, la línea de resumen y **la línea de contraste**;
  - nunca escribe estado ni cola.
- **FR-026** *(los textos de antes no cambian)*: sin OpenCode activo, la salida de `--run`, del demonio y de `--scan` es **byte a byte** la de la 0.6.0. Con
  OpenCode, sólo se añade la línea de FR-024, y «N eventos encolados» cuenta también los de OpenCode.

### E · Frontera, durabilidad y errores

- **FR-027** *(orden de escritura)*: OpenCode se lee **después** de Gemini y **antes** del único `st.Save` *(`main.go:368`)*. Sus eventos se encolan antes
  de guardar el estado.
- **FR-028** *(sin las otras herramientas)*: una instalación sólo con OpenCode **DEBE** enviar sus eventos.
- **FR-029** *(OpenCode no rompe a los demás)*: un error al leer una base *(E/S, estructura imposible o inestable tras 5 intentos)* **DEBE**:
  - escribir `opencode: base omitida: %v`;
  - **no** avanzar su frontera ni tocar su lista;
  - **no** cortar la pasada;
  - **no** impedir guardar el estado de los demás.

  Un fallo al **encolar** sigue siendo fatal.
- **FR-030** *(frontera — M-1)*: `git diff 4c6cafc` de lo siguiente, **vacío**:
  - `internal/event/` e `internal/ingest/eventid.go`;
  - `eventid_test.go`, `codex_eventid.go`, `codex_eventid_test.go`, `gemini_eventid.go`, `gemini_eventid_test.go` y `boundary_test.go`;
  - los contratos de 006, 008 y 009.
- **FR-031** *(versión — M-2)*: `0.7.0`. Lo demás, exactamente como en la 0.6.0. **No se publica** hasta que **P-033** de la plataforma esté en producción
  *(SC-027; dueño, 11-10)*.

### Key Entities
- **Llamada de OpenCode**: una fila `assistant` completada. Lleva `id`, `session_id` y `time_updated`, y en `data`, `tokens.{input, output, reasoning,
  cache.read, cache.write}`, `model.id`, `time.completed` y `cost` *(sólo contraste)*.
- **Instantánea de una base**: la base más los frames confirmados del `-wal`, estable durante la lectura. Vive **sólo en memoria, durante la pasada**.
- **Frontera de una base**: H, la huella del `id` de la fila H y la lista en curso *(P-3)*.

## Success Criteria

Las referencias son las de la prueba *(`sqlite3` sobre la copia congelada en su sitio, `file:…?mode=ro&immutable=1`)*. Las huellas del contenido **no** se
publican.

- **SC-001** *(contra la copia congelada)*: con `--scan` sobre la copia y con `--run` en un sandbox con `OPENCODE_DB` = la copia *(sin escribir en ella)*:
  - **8** eventos `opencode`, los 8 con `gemini-3.8-flash`;
  - `tokens_input` **59 321**, `tokens_output` **692** *(230 + 462)*, `tokens_cache_read` **3 999** y `tokens_cache_creation` **0**;
  - resumen: `opencode: llamadas 8 · eventos 8 · en curso 0 · abandonadas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0 · fronteras reiniciadas 0 · bases en formato anterior 0 · bases con esquema desconocido 0`.
  - La copia conserva su sha256.
- **SC-002** *(contraste — D-1, D-2)*:
  - en `--scan` sobre la copia, la línea de contraste da **$0.047386** según OpenCode;
  - fuera de los mensajes, **entrada 538 · salida 5 · $0.000174**, que es el **0,37 %** de D-2.
- **SC-003** *(segunda pasada)*: una segunda `--run` sobre la copia da **0** eventos `opencode`.
- **SC-004** *(WAL sintético)*: bases de `testdata`, generadas por `sqlite3` con la forma de la prueba *(I, R, S)*, dan exactamente las cifras que dio
  `sqlite3` al generarlas:
  - la transacción a medias **nunca** aparece;
  - en R, la transacción con el frame roto no existe;
  - con filas que desbordan *(> 8 KB)*.
- **SC-005** *(mutaciones — D-3)*: mB1–mB4, rehechas sobre el código definitivo, caen **exactamente** donde caían en la prueba: mB1 en
  I y R; mB2 sólo en R; mB3 en I, R y S; mB4 sólo en I. Cada una se declara antes de mutar, con md5 antes, mutado y restaurado.
- **SC-006** *(sólo lectura)*: sobre una base sintética viva *(una conexión de `sqlite3` abierta y transacciones en el WAL)*, una `--run` deja **iguales**
  las md5 de `opencode.db`, `-wal` y `-shm`, y el listado de la carpeta. En el ensayo de Windows *(W1)*, lo mismo contra la base real **con OpenCode
  cerrado**; con OpenCode abierto, ningún fichero nuevo en la carpeta *(E-1)*.
- **SC-007** *(binario)*: frente al binario de `4c6cafc`, con las banderas de `.goreleaser.yaml`, crece **≤ 100 KB** en **las 5 plataformas**: el fichero; en
  linux/arm64, su contenido *(Σ `FileSiz` de los `LOAD`, sin el relleno a 64 KB; E-3)*. **Se mantiene** *(dueño, 11-10)*: si no cabe, se adelgaza y el tope no sube.
- **SC-008** *(dependencias)*: `go.mod` sin `require`, no hay `go.sum`, y `go list -m all` da una sola línea.
- **SC-009** *(fuzzing)*:
  - `go test -fuzz`, durante **10 minutos por objetivo**, sobre dos objetivos: la base con bytes arbitrarios, y un `-wal` arbitrario junto a una base sana;
  - **0** pánicos y **0** entradas que tarden más de 1 s;
  - el corpus semilla, sintético.
- **SC-010** *(en curso)*: una fila sin `time.completed` da `eventos 0 · en curso 1`. Al completarla, la pasada siguiente da `eventos 1 · en curso 0`. Con P-1
  (b), el test cae.
- **SC-011** *(frontera — P-3)*:
  - **la pasada que sigue**: con 4 filas completadas y 1 en curso, la 1.ª pasada da `eventos 4 · en curso 1`. La 2.ª, sin cambios, da 0 eventos, y el
    contador de páginas sólo ve los descensos a H y a la fila en curso;
  - **orden inverso**: dos filas que se completan al revés de su rowid salen una vez cada una;
  - **fila borrada**: una fila de la lista que desaparece sale de ella sin contarse;
  - **entrada sin el campo nuevo**: hay una pasada completa, sin «frontera reiniciada».
- **SC-012** *(partidas, sintético)*: `input` 100, `output` 7, `reasoning` 3, `cache.read` 40 y `cache.write` 2 dan:
  - `tokens_input` 100, `tokens_output` **10**, `tokens_cache_read` 40 y `tokens_cache_creation` 2;
  - sin `reasoning` en la salida, el test cae.
- **SC-013** *(clasificación, sintético)*: una fila de cada clase da `llamadas 5 · eventos 2 · en curso 1 · abandonadas 1 · sin identificador 1 ·
  incoherentes 2 · sin modelo 1 · fronteras reiniciadas 0`. Las clases:
  - evento;
  - evento sin modelo;
  - sin `id`;
  - `data` que no es JSON;
  - partida negativa;
  - una en curso;
  - y una en curso de hace 25 h.

  Cumple la identidad de FR-021.
- **SC-014** *(identidad y privacidad)*:
  - la función nueva reproduce los vectores del contrato;
  - un fixture lleva **centinelas** en `id`, `session_id`, `directory`, `path`, `title`, el texto de una fila `user` y el nombre de la base;
  - tras una pasada, **ninguno** está en la cola ni en lo transmitido;
  - en `state.json` sólo está la ruta de la base, **dentro de su clave**.
- **SC-015** *(formatos y ubicación)*:
  - una base 1.x da «formato anterior 1», y 0 en la pasada siguiente;
  - `WITHOUT ROWID`, UTF-16 o una columna que falta dan «esquema desconocido 1», con salida 0 y los demás eventos fuera;
  - `opencode-next.db`, `-wal`, `-shm` y `storage/` no se leen;
  - `XDG_DATA_HOME=<dir>` → `<dir>/opencode`, y `XDG_DATA_HOME=""` → `<hogar>/.local/share/opencode`;
  - `OPENCODE_DB` absoluta → sólo esa base; `""`, `":memory:"` o relativa → la regla general.
- **SC-016** *(sin OpenCode)*: sin carpeta ni variables, la salida de `--run` sobre los fixtures de las otras tres herramientas es **byte a byte** la del
  binario de `4c6cafc`.
- **SC-017** *(coste — P-3)*: sobre una base **sintética de 1 GB**, generada en un temporal y nunca en el repo. Lleva ≈ 250 000 filas `assistant` con el
  `data` de la 2.0.26, de 1 a 12 KB, en ≈ 2 500 sesiones. Se mide la función del lector *(leer y clasificar)*:
  - **la primera pasada**, en **≤ 20 s**. Es una propuesta: en el ensayo, B leyó 6,3 MB en 6 ms de mediana *(≈ 1 ms/MB, ≈ 1–2 s por GB)*, y el margen
    cubre la E/S en frío de WSL y el JSON de 250 000 filas;
  - **cada pasada siguiente**, con **10 filas nuevas**, en **≤ 50 ms**, leyendo **≤ 1 %** de las páginas de la base. Lo mide un contador de páginas
    instrumentado en el test, que incluye `session_v2` y la lista.

  **No entran**: encolar *(≈ 250 000 `transport.Append`, una apertura por evento: `internal/transport/queue.go:23-38`)*, como en la primera pasada de los
  otros lectores; y el `-wal`, que se lee entero *(FR-007)* y lo acota el checkpoint de OpenCode *(≤ 1 490 frames en el ensayo)*.

  Cada medida, tres veces. Si no se cumple, cambia el mecanismo; los topes no se mueven.
- **SC-018** *(puertas)*: frontera con 0 bytes de diff *(FR-030)*; los **671** tests actuales, verdes y **sin tocar** *(`git diff 4c6cafc --stat --
  '*_test.go'` sólo con ficheros nuevos)*; `golangci-lint run` → 0; `os.Getenv` → 4.
- **SC-019** *(textos)*: los textos nuevos, byte a byte iguales a los aprobados *(los de README y CHANGELOG, en T047; E-1)*, también sobre el binario
  publicado.
- **SC-020** *(OpenCode no rompe a los demás)*: una base corrupta, junto a fixtures sanos de las tres herramientas:
  - salen los eventos sanos, stderr lleva `opencode: base omitida: …`, y `state.json` se guarda sin avanzar la frontera de esa base;
  - cuando la base vuelve a estar sana, sus llamadas salen.
- **SC-021** *(orden de escritura y sin otras herramientas)*:
  - con el forzado de 008 SC-015 *(`Save` falla)*, la cola tiene los eventos de OpenCode, y la pasada siguiente los reencola con los **mismos** `event_id`;
  - sólo con OpenCode, `--run` encola y termina con 0.
- **SC-022** *(privacidad del repo público)*:
  - en el diff de 010, el `grep` de privacidad de los encargos da **0**. Busca correos, tokens `dev_…`, ids `ses_…` y `msg_…`, el usuario de la
    máquina, rutas personales de Linux y de Windows y claves de Google;
  - `testdata` sólo lleva bases **sintéticas**;
  - ninguna huella del contenido de la copia se publica.
- **SC-023** *(ensayo en Windows)*: con OpenCode escribiendo, el agente emite cada llamada completada **una vez**, cumple SC-006, y sus eventos cuadran
  **al token** con un contador independiente: en **W1**, antes de publicar *(sandbox; `sonda-a.exe` como contador)*, y en **W2**, en la instalación real,
  **después de P-033** *(como 009 W2; E-1)*.
- **SC-024** *(renumeración — P-3)*: sobre una base sintética:
  1. una pasada emite;
  2. se borran 2 filas anteriores a H y se hace `VACUUM` con `sqlite3`. El test **comprueba que los rowid cambiaron**: la fila H tiene otro rowid u otro `id`;
  3. se añaden 3 filas completadas.

  La pasada siguiente emite las 3, **ninguna** completada de la base falta entre sus `event_id` *(frente a `sqlite3`)*, y cuenta `fronteras reiniciadas 1`.
- **SC-025** *(abandono — P-3)*: dos filas en curso:
  - la de `$.time.created` de **hace 25 h** cuenta `abandonadas 1` y sale de la lista; en la pasada siguiente, `abandonadas 0`;
  - la de **hace 1 h** sigue, `en curso 1`;
  - una fila de la lista que se completa con más de 24 h sale como **evento**, no como abandonada.
- **SC-026** ✅ *(segundo proveedor — Q-5; **cumplido** el 2026-10-11)*: en la copia congelada de `anthropic` *(`3a8e563f…`)*, las 5 filas
  `claude-sonnet-5` cumplen la identidad de desc. D-4 con las partidas de P-2, con **Δ = 0**. La **consola de Anthropic** coincide al token: entrada 10 ·
  lectura 46 851 · escritura 11 995 «(5m)» · salida 397 = 369 + 28 *(`soporte/medicion-anthropic.md`)*. P-2 **no** se enmienda por proveedor.
- **SC-027** *(publicación — dependencia de P-033)*: la etiqueta `v0.7.0` no se crea hasta que **P-033** esté en producción. Se acredita reparando en la
  plataforma las 5 filas de la medición *(escritura a 5 min, `ROUND(·, 6)` por evento)*: **0,043348**, y no 0,061341 `DECIDÍ YO` *(el vector)*.
- **SC-028** *(el evento — FR-017 a FR-019, D-1; E-1)*: el 100 % de los eventos `opencode` llevan `cost_usd = 0` y `cost_available = false`, y ninguno
  `tool ≠ "opencode"`; `occurred_at` = `$.time.completed` en UTC *(y no `time.created`)*; `session_ref` = `event.Ref(sal, session_id)` y `project_ref` =
  `Derivar(directory, sal)`; sin fila de sesión, las dos vacías.
- **SC-029** *(estabilidad — FR-008; E-1)*: una base que cambia entre la lectura y la comprobación se relee, y sus llamadas salen **una vez**; si cambia
  en los **5** intentos, se omite con `opencode: base omitida: …`, su frontera no se mueve y la pasada siguiente la lee.

## Textos propuestos *(P-12)*

**Resumen de OpenCode** *(stderr, una línea; en `--run`, tras la de Gemini; en el demonio, sólo con novedades)*:
```
opencode: llamadas %d · eventos %d · en curso %d · abandonadas %d · sin identificador %d · incoherentes %d · sin modelo %d · fronteras reiniciadas %d · bases en formato anterior %d · bases con esquema desconocido %d
```
**Base omitida** *(stderr, una por base)*:
```
opencode: base omitida: %v
```
**Contraste** *(sólo en `--scan`, stdout, tras el resumen; D-1, D-2)*:
```
opencode (contraste, no se envía): coste según OpenCode $%.6f · fuera de los mensajes: entrada %d · salida %d · coste $%.6f
```
«Fuera de los mensajes» = los totales de `session_v2` menos la Σ de las filas `assistant` de la base, con la salida como `output + reasoning`.

## Preguntas abiertas *(con cómo se cerrarían)*

| # | Pregunta | Propuesta mientras tanto | Cómo se cierra |
|---|---|---|---|
| **Q-1** | ¿Qué es `opencode-next.db` *(«Incompatible opencode-next.db…», «Reassigned previous V2 session»)*, y cuándo existe? | **No se lee**: podría duplicar sesiones, aunque la plataforma descartaría los `event_id` repetidos | Leer en el binario 2.0.26 quién llama a `nextDatabasePath`, sin ejecutarlo. Y mirar si existe en la instalación del dueño **por su nombre**, sin abrirla |
| **Q-2** | ¿Una **compactación** o un resumen dejan fila `assistant` en `session_message`, o sólo suben los totales de `session_v2`? ¿Y una respuesta **interrumpida** por el usuario llega a tener `time.completed`? | Si no dejan fila completada, son otro caso del límite de D-2 | Una sesión de medición con compactación *(`/compact` o automática)* y una interrupción, en copia congelada: recuento por `type` y con `time.completed`, y `session_v2` − Σ de los mensajes, como desc. D-3 |
| **Q-3** | ¿Qué hacer ante `WITHOUT ROWID`, UTF-16 o un esquema distinto al de la 2.0.26? | **No leer, contarlo en el resumen y nunca fallar** *(FR-023)*. Un error de E/S o una estructura imposible sí avisan *(FR-029)* | Ratificación del dueño y SC-015. Ante cada versión nueva de OpenCode, rehacer SC-001 sobre una copia nueva |
| **Q-4** | ¿Qué hace **`opencode import`** en `session_message`: inserta filas nuevas, o mueve rowids? Con varios procesos no hay pregunta: SQLite serializa a los escritores, y el orden de los rowid confirmados está garantizado *(§P-3)* | P-3 tal cual. Una fila insertada tiene rowid > H y se emite; si trae el `id` de otra máquina, la plataforma descarta el `event_id` repetido. Un rowid movido reinicia la frontera | Una importación en copia congelada: recuento de filas nuevas y de rowids cambiados frente a la copia de antes |
| **Q-5** ✅ *(del dueño; **cerrada** el 2026-10-11)* | ¿Normaliza OpenCode las partidas igual con otros proveedores? | — | **Sí**: P-2 queda verificada con `google` y `anthropic`, con Δ = 0 y sin doble conteo de caché ni de razonamiento. Anthropic informa el razonamiento dentro de su salida, y OpenCode lo reparte. La escritura es de 5 min *(copia `3a8e563f…` y consola; `soporte/medicion-anthropic.md`; SC-026)* |

## Censo de tests — autorización para la fase de tareas

**Ninguno de los existentes cambia.** Los nuevos van en ficheros nuevos, con fixtures **sintéticos** nuevos en `testdata/`.

| Dónde | Nace |
|---|---|
| `internal/sqlitero/` *(nuevo)* | `lector.go` *(el de la prueba)*; `lector_test.go` *(cabecera, árbol, registros, desbordamiento y WAL)* y `fuzz_test.go`; las bases I, R y S *(`db`, `-wal` y `-shm`)*, los raros, y el guion `sqlite3` que las genera, versionado y fuera del binario |
| `internal/config/` | `opencode.go` y `opencode_test.go` *(P-8)* |
| `internal/ingest/` | `opencode.go`, `opencode_base.go`, `opencode_eventid.go`; `opencode_eventid_test.go`, `opencode_llamada_test.go`, `opencode_base_test.go`, `opencode_coste_test.go` *(sólo con `-tags sc017`)*; y `testdata/opencode/` |
| `cmd/permea/` | `opencode.go` y `opencode_test.go` *(`--run`, demonio, `--scan`, SC-016, SC-020, SC-021)* |
| `internal/testutil/` | `sandbox_opencode_test.go`, como `sandbox_gemini_test.go` |
| `specs/010-lector-opencode/contracts/` | `event-id-opencode.md` |

**Fuera de los `_test.go`** *(M-3; E-1)*: `internal/testutil/sandbox.go` gana `t.Setenv("XDG_DATA_HOME", "")` y `t.Setenv("OPENCODE_DB", "")`;
`internal/state/state.go`, el campo opcional de FR-004; e `internal/config/codex.go:16` y `gemini.go:17`, sólo el comentario *(«una de las CUATRO»)*.
- `sandbox_test.go:20` comprueba `HOME`, `USERPROFILE` y `XDG_CONFIG_HOME`: las nuevas no cambian su veredicto.
- `XDG_DATA_HOME` no la lee hoy ningún código del agente *(`grep`: 0)*, así que vaciarla no cambia nada más.

## Fuera de alcance

| # | Qué | Por qué |
|---|---|---|
| **N-1** | El coste y las tarifas | D-1 |
| **N-2** | Las llamadas fuera de `session_message` y un evento residuo | D-2 |
| **N-3** | Leer bases 1.x y `storage/` | P-9 |
| **N-4** | `opencode-next.db` | Q-1 |
| **N-5** | Las tablas `event`, `session_pending` y `session_inbox`, y las de cuentas y credenciales | No son consumo; de las de cuentas no se lee ni una fila |
| **N-6** | `log/opencode.log` | No lleva consumo *(desc. D-8)* |
| **N-7** | Leer desde WSL las bases de Windows | P-8 lee el hogar propio |
| **N-8** | Usar los índices de SQLite | El lector baja por el árbol de la tabla por rowid *(P-3)*; lo acota SC-017 |
| **N-9** | Un interruptor | P-10 |
| **N-10** | Un alias de modelos en la plataforma | Otra feature *(P-5, 11-10)* |

## Límites declarados

- **D-2**: la llamada del título y las que no están en `session_message` quedan **por debajo de la factura**. En la medición, el **0,37 %**.
- **P-3**:
  - una fila en curso más de 24 h no se emite nunca, aunque después se complete con el agente parado más tiempo *(«abandonada»)*;
  - una frontera reiniciada reencola **todas** las completadas de esa base. La plataforma descarta las repetidas, pero la cola crece una vez.
- **Dependencia de la plataforma**: con proveedores que escriben caché a 5 min *(Anthropic vía OpenCode)*, la reparación de la plataforma anterior a
  **P-033** cobra la escritura a 1 h *(+41,5 % en la medición)*. La 0.7.0 **no se publica** hasta que P-033 esté en producción *(decisión del dueño, 11-10)*.
- **Sesiones borradas antes de leerlas**: se pierden *(`ON DELETE CASCADE`)*.
- **Sin bloqueos**:
  - la coherencia descansa en la estabilidad *(FR-008)*, que en vivo **no llegó a reintentar** *(176 pares, 0 reintentos)*;
  - una lectura inestable cinco veces seguidas se omite y se reintenta en la pasada siguiente.
- **Lo que el lector no cubre** *(prueba §Lo que NO cubre B)*: se cuenta como esquema desconocido o como base omitida, y nunca se lee a medias.

## Assumptions
1. **La regla de ubicación es la misma en los tres SO** *(desc. D-9, leído en el binario)*. En Windows: `%USERPROFILE%\.local\share\opencode`. El ensayo la
   confirma.
2. **Una fila `assistant` no cambia sus tokens una vez completada.** Si lo hiciera, cuenta la primera versión completada que se vea *(P-3)*.
3. **El `id` de la fila es único y estable por llamada** *(desc. D-5)*.

## Dependencias
- **La plataforma**: `tool = "opencode"` como texto libre, las tarifas de los modelos que lleguen *(hoy, `gemini-3.8-flash` y `claude-sonnet-5`)*, y
  **P-033** en producción antes de publicar *(SC-027)*.
- **006**: el esquema del `event_id`, sin tocarlo. **008** y **009**: el patrón de lector, con errores por fichero, recuentos y la línea de resumen, sin
  cambiar su conducta.

## Registro de enmiendas

| # | Fecha | Qué cambia | Por qué |
|---|---|---|---|
| **Borrador** | 2026-10-11 | Primera redacción, con D-1 a D-3 ratificadas, y P-1 a P-12 y Q-1 a Q-4 para ratificar | Encargo 010 · constancia y spec · 3/n |
| **Ratificación** | 2026-10-11 | **El dueño ratifica** P-1, P-2 y P-4 a P-12 **(a)**:<br>• **P-5**, con la redacción corregida: un modelo que no casa queda sin reparar y parcial, y el alias es otra feature *(N-10)*;<br>• **SC-007** *(≤ 100 KB)* se mantiene.<br>**P-3 se sustituye** por la **frontera por rowid** *(§P-3)*, con FR-004, FR-005, FR-012, FR-020, FR-021, FR-024 y FR-029 ajustados. SC-011 y SC-017 se reescriben; nacen SC-024 *(renumeración)* y SC-025 *(abandono)*. La línea de resumen gana `abandonadas` y `fronteras reiniciadas` *(SC-001, SC-013)*. Q-4 queda sólo para `opencode import` | Encargo 010 · remate de la spec · 4/n |
| **Segundo proveedor** | 2026-10-11 11:00 | **El dueño** abre **Q-5**: P-2 está medida sólo con `google`, y con otros proveedores la caché podría contarse mal. Mientras tanto, P-2 vale para todos, con el límite en §Límites. Nace **SC-026**: medición con `anthropic` antes de la 0.7.0; si falla, P-2 por proveedor con su mutación; sin medición, límite en README y CHANGELOG | Encargo 010 · segundo proveedor · 4b/n |
| **Cierre de Q-5** | 2026-10-11 12:05 | **La consola de Anthropic** *(el dueño, 12:02)* coincide al token con la copia: salida 397 = 369 + 28 y escritura «(5m)». **Q-5, cerrada**; **SC-026, cumplido**. En §Límites, «partidas verificadas sólo con Google» deja paso a la **dependencia de P-033**. Nace **SC-027**, que FR-031 cita: la 0.7.0 no se publica sin P-033 en producción. Se condensan los puntos 1 y 3 del Contexto y la nota de SC-017, sin tocar nada normativo | Encargo 010 · cierre de Q-5 · 6/n |
| **E-1** | 2026-10-11 12:35 | **El dueño** resuelve las discrepancias del plan *(7/n)* y los hallazgos del analyze *(8/n)*. **FR-005**: «Antes: el coste de una pasada **DEBE** ser proporcional a lo nuevo y a la lista en curso, no al tamaño de la tabla *(SC-017)*.» → «… a lo nuevo, a la lista en curso **y al número de sesiones** …» *(y §P-3 «Coste»)*. **SC-006**: «Antes: En el ensayo de Windows, lo mismo, contra la base real.» → W1 con OpenCode **cerrado**; abierto, sin ficheros nuevos. **SC-019**: «Antes: los textos nuevos, byte a byte iguales a los aprobados, también sobre el binario publicado.» → los de README y CHANGELOG, en T047. **SC-023**: «Antes: … con un contador independiente *(como 009 W2)*.» → W1 antes de publicar, W2 después de P-033. **Censo**: «Antes: `opencode.go`, `opencode_eventid.go`, `opencode_eventid_test.go`, `opencode_test.go` y `testdata/opencode/`» → los nombres del plan; y fuera de los `_test.go`, `internal/state/state.go` y los comentarios de `internal/config/{codex,gemini}.go`. **Nacen SC-028** *(el evento: FR-017 a FR-019)* **y SC-029** *(estabilidad: FR-008)* | Plan 7/n, DISCREPANCIAS 1–5; analyze 8/n, H1, H12, H18 |
| **E-2** | 2026-10-12 | **El dueño** enmienda §P-3 y §Edge Cases. «Antes: una **renumeración** por `VACUUM`» y «*(renumera los rowid implícitos)*» → `VACUUM` **puede** renumerar: con SQLite 3.45.1 y `id text PRIMARY KEY` conserva los rowid; sin índices, los renumera *(medido en B4, T023)*. La defensa se mantiene y `vacuum/` prueba el caso. Para caber, §P-3 se condensa sin tocar nada normativo | Encargo 010 · B4 remate |
| **E-3** | 2026-10-12 | **El dueño** cambia el instrumento de SC-007 en linux/arm64; el tope no cambia. «Antes: el aumento frente al binario de `4c6cafc` […] es **≤ 100 KB** en **las 5 plataformas**», medido en el fichero → en linux/arm64 se mide **el contenido de los segmentos cargables** *(Σ `FileSiz` de los `LOAD`)*, y el Δ del fichero se informa al lado, sin juzgar; en las otras cuatro, el fichero. Para caber, SC-007 y §Contexto 2 se condensan sin tocar nada normativo *(sale «El lector solo midió de 51 a 66 KB», que sigue en plan §Summary)* | Encargo 010 · B6 · PARA 7: la alineación a 64 KB *(fichero +131 072 B, contenido +78 544 B)* |
