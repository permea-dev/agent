# 013 · «Lista de proyectos» — Especificación

**Feature Branch**: `013-lista-de-proyectos` · **Created**: 2026-10-23 · **Status**: **Ratificada el 2026-10-23**
*(D-1 a D-11 y P-1 a P-12, del dueño; la 0.10.0 y SC-016, también)* · **Cabeza de partida**: `6417548` *(«013 B0: descubrimiento»; el código es el de `cabf8c5`, 0.9.0)*
**Input**:
- las decisiones del dueño D-1 a D-11 *(2026-10-23)*;
- el descubrimiento del 23-10 *(`soporte/descubrimiento.md`, Q1–Q7; «desc. Qn»)*.

> ⛔ Toda afirmación sobre nuestro código lleva `fichero:línea`, medida sobre `6417548`; las de la plataforma, `[P fichero:línea]` sobre `permea-platform`
> `f473ae0`. Lo que añadió la spec sin que nadie lo dijera va marcado **`DECIDÍ YO`**; P-1 a P-12 están ratificadas. Ningún dato procede de las carpetas o registros
> del dueño: sólo recuentos y tiempos *(desc. §Privacidad)*.
>
> **Ortografía**: «sólo», **con tilde**. Es la convención de los textos aprobados más recientes *(`CHANGELOG.md`: 6 «sólo», 0 «solo»)* y de las specs
> 008–012. El README mezcla las dos: sus secciones antiguas escriben «solo», y las nuevas, «sólo» *(D-11)*.

## Contexto — por qué hace falta y qué hay ya

1. **El panel enseña las identidades sin asignar como un hash truncado**, `key.slice(0, 12)` + `…`, con sus eventos y su `first_seen → last_seen` *(UTC,
   `AAAA-MM-DD`, de toda la vida del ref)* [P `frontend/src/screens/metrics/PorProyecto.tsx:1137-1157`; `backend/app/Services/Metrics/MetricsQuery.php:478-495`].
   El dueño no sabe de qué carpeta es cada una *(desc. §Por qué)*.
2. **La identidad es local y recalculable**: `event.Ref(sal, raíz o directorio)` *(`internal/event/event.go:40-46`; `internal/project/resolve.go:157-171`)*,
   con la sal de `<DataDir>/salt` *(`internal/config/identity.go:13-14`)*, que nunca ha cambiado por código *(desc. Q2)*.
3. **Los seis lectores pasan por el mismo punto**, `Resolutor.Derivar(cwd, sal)` *(`resolve.go:318-328`)*. Se pueden recorrer enteros con un estado en
   memoria, sin cola: es lo que hace `--scan`, fichero a fichero *(desc. Q3)*.
4. **El estado local no sirve para contar**: la cola sólo guarda lo no enviado *(`internal/transport/queue.go:84-128`)*, y `state.json` no lleva refs
   ni carpetas *(`internal/state/state.go:16-45`)*.
5. **`project join` no guarda nada en local** *(`cmd/permea/project.go:131-240`)*: no se puede marcar lo ya unido *(desc. Q5)*.
6. **Coste medido**: Claude Code, 49 MB en WSL en 0,24–0,27 s; 72 MB por `/mnt/c` en 0,7–1,0 s. Las otras cinco herramientas no tienen muestra en este
   equipo *(desc. Q4)*.

## Decisiones del dueño *(2026-10-23; no se discuten aquí)*

| # | Decisión |
|---|---|
| **D-1** | **Una fila por identidad.** Las herramientas, como dato, separadas por comas y con su **nombre comercial**, como el panel: `claude_code` → Claude Code · `codex` → Codex · `gemini` → Gemini CLI · `opencode` → OpenCode · `kimi_code` → Kimi Code CLI · `copilot_cli` → **GitHub Copilot CLI** *(el panel aún no lo nombra: §Dependencias)* · un id desconocido, **tal cual** |
| **D-2** | **Columnas**, en este orden: `IDENTIDAD` *(12 caracteres y `…`)* · `CARPETA` *(la raíz del repositorio, o el directorio si no hay repositorio; la carpeta personal como `~`)* · `HERRAMIENTAS` · `EVENTOS` *(con punto de miles: «1.240»)* · `DESDE` y `HASTA` *(UTC, `AAAA-MM-DD`)*. **Orden**: por eventos, de más a menos; si empatan, por identidad |
| **D-3** | **Marcas** tras la carpeta: `(sin repositorio)` y `(ya no existe)` |
| **D-4** | **El pie**, letra a letra: `N identidades, N eventos. Fechas en UTC, como en el panel.` *(singular para 1)* y `Cuenta lo que las herramientas conservan en disco, también lo aún no enviado: el panel puede tener más o menos.` *(la segunda, enmendada en la ratificación)* |
| **D-5** | **Sin sal**: no se crea. Dice «esta instalación aún no ha medido nada» y sale con 0 |
| **D-6** | **No exige enrolamiento, no usa la red y no escribe nada**: ni la cola, ni los cursores, ni `state.json`, ni la sal |
| **D-7** | **Sin las identidades de antes de la 0.3.0** en la v1. Se declara en §Limitaciones |
| **D-8** | **Sin `--json`** |
| **D-9** | **`project join` no cambia**: ni con una carpeta ni con un hash |
| **D-10** | Es el **primer comando que imprime rutas reales**. Sólo salen a la terminal y no cruzan la frontera. Lo dicen el plan y §Constitución |
| **D-11** | **La ayuda**: la del descubrimiento *(desc. Q6)*, con «Verbos disponibles: join, list», y con la convención de «sólo» del repositorio |

## Propuestas — **ratificadas por el dueño (23-10)**, tal cual *(todas `DECIDÍ YO`)*

| # | Propuesta | Alternativas | Recomendación |
|---|---|---|---|
| **P-1** | **Eventos sin identidad** *(`project_ref` vacío: sin `cwd`, o una sesión de Kimi sin línea en el índice)*: **ni fila ni recuento** | (b) una fila «(sin carpeta)» | **(a)**: no hay nada que asignar. El panel también los aparta [P `MetricsQuery.php:506-509`] |
| **P-2** | **Se cuentan `event_id` distintos**, como la plataforma *(`ON CONFLICT (org_id, event_id) DO NOTHING`)* [P `backend/app/Ingest/IngestBatchService.php:10`]. Un `event_id` que aparezca con dos identidades cuenta en la **primera leída**; en la segunda **no aporta nada** *(ni recuento, ni herramienta, ni fechas)*, y una identidad que sólo tenga repetidos **no sale** | (b) contar eventos emitidos | **(a)**: Claude Code copia el historial al reanudar, y (b) contaría dos veces lo que el panel cuenta una |
| **P-3** | **Orden de las herramientas en la fila**: el de lectura de `generate()` *(Claude Code, Codex, Gemini CLI, OpenCode, GitHub Copilot CLI, Kimi Code CLI; `cmd/permea/main.go:300-436`)*. Los ids desconocidos van detrás, por orden alfabético | (b) por eventos | **(a)**: es estable y no cambia por un evento más |
| **P-4** | **Una sola marca**: si la carpeta ya no existe, `(ya no existe)`, **aunque** la derivación haya caído al fallback; si existe y no tiene repositorio, `(sin repositorio)` | (b) las dos a la vez | **(a)**: de una carpeta borrada no se sabe si tenía `.git`. «Sin repositorio» diría algo que no consta |
| **P-5** | **Con sal y sin ningún evento**: sin cabecera ni filas, sólo el pie: `0 identidades, 0 eventos. …` | (b) un texto propio | **(a)**: no añade texto |
| **P-6** | **Canales**: la tabla y el pie, por **stdout**, con salida 0. Los avisos que ya escriben los lectores salen por **stderr**, sin cambios: `skip (línea corrupta): …` *(`main.go:342`)*; `codex: fichero omitido: …`, `gemini: …`, `copilot_cli: …` y `kimi_code: …` *(`main.go:450`; `gemini.go`, `copilot.go`, `kimi.go`)*; `opencode: base omitida: …` *(`opencode.go:50`, `:59`)*. **Sin** las líneas de resumen de `--run`. **Un error de lectura de Claude Code** *(su raíz o un fichero)* **detiene la lista**, como detiene `--run` *(`main.go:301-308`, `:364-366`)*: `error: <error>` y salida 1 | (b) callar los avisos · (c) que Claude Code avise y siga | **(a)**: no se esconde un fichero que no se pudo leer, y (c) cambiaría la emisión, que comparte el código *(FR-024)* |
| **P-7** | **Argumentos**: ninguno. `-h` y `--help` dan la ayuda. Cualquier otro da un error de uso, con salida 1, como `project <verbo desconocido>` *(`project.go:108`)*, **sin reproducir lo tecleado** | (b) reproducirlo, con la regla `pareceSecreto` | **(a)**: es más simple, y nunca enseña un secreto |
| **P-8** | **Configuración**: se lee *(`config.Load`, `internal/config/config.go:68-82`)* sólo por `logs_root`. Si es ilegible: error con salida 1. **Con `project_ref_mode: "plain"`** *(el valor retirado; `config.go:163-166`)*, la lista **ni se detiene ni avisa**: no emite nada, y el aviso es de `status` *(`status.go:48`)* y la parada, de `setup()` *(`main.go:224`)* | (b) ignorarla | **(a)**: sin `logs_root`, una instalación no estándar listaría otra carpeta |
| **P-9** | **Qué se cuenta**: lo que la emisión **enviaría hoy**, con las mismas reglas de los seis lectores y el mismo reloj. Un mensaje de Claude Code **todavía abierto** *(menos de 10 min sin cambios; 007 FR-010)* no cuenta aún | (b) cerrarlo todo, como `--scan` | **(a)**: un solo recorrido para los dos usos, así que lo listado y lo enviado no pueden divergir |
| **P-10** | **La carpeta personal**: la real *(con los enlaces resueltos, como el techo; `resolve.go:241-249`)*. Si la carpeta **es** el directorio personal, `~`; si cuelga de él, `~` + separador + resto, con el separador del sistema *(en Windows, `~\dev\x`)*. Se compara por segmentos, no por prefijo *(`resolve.go:265-275`)* | (b) siempre `/` | **(a)**: es la ruta que el usuario reconoce en su sistema |
| **P-11** | **Sin sal**, letra a letra: `Esta instalación aún no ha medido nada.`, por stdout y con salida 0. Igual si no existe el `DataDir` | (b) por stderr | **(a)**: es la respuesta, no un error *(D-5)* |
| **P-12** | **Ancho de las columnas, en runas**: el máximo entre la cabecera y sus valores | (b) en bytes | **(a)**: una ruta con tildes no descuadra. Los caracteres de doble ancho sí: §Limitaciones |

## User Scenarios & Testing

### Historia 1 — El dueño cruza el panel con su carpeta *(P1)*
1. **Dada** una instalación con eventos de Claude Code en `~/dev/proyecto-ejemplo` *(con `.git`)*, **cuando** ejecuta `permea project list`, **entonces**
   ve una fila con los 12 primeros caracteres de su identidad, que son los del panel; `~/dev/proyecto-ejemplo`, `Claude Code`, los eventos y las fechas.
2. **Dados** eventos de Claude Code y de Codex en la misma raíz, **cuando** ejecuta `permea project list`, **entonces** sale **una** fila con
   `Claude Code, Codex`.
3. **Dados** eventos lanzados desde dos subdirectorios del mismo repositorio, **cuando** ejecuta `permea project list`, **entonces** sale **una** fila,
   con la raíz.

### Historia 2 — Las carpetas raras se distinguen *(P1)*
1. **Dada** una carpeta sin `.git`, **cuando** ejecuta `permea project list`, **entonces** su fila lleva `(sin repositorio)`.
2. **Dada** una carpeta que ya no existe, **cuando** ejecuta `permea project list`, **entonces** lleva `(ya no existe)`, y sólo esa marca *(P-4)*.

### Historia 3 — No envía ni toca nada *(P1)*
1. **Dada** una instalación enrolada, con `endpoint` y `device_token`, **cuando** ejecuta `permea project list`, **entonces** no sale ninguna petición
   de red, y ningún fichero del sistema cambia ni aparece.
2. **Dada** una instalación sin sal, **cuando** ejecuta `permea project list`, **entonces** dice `Esta instalación aún no ha medido nada.`, sale con 0, y
   la sal sigue sin existir.

### Historia 4 — Lo de siempre no cambia *(P1)*
1. **Dada** una instalación de la 0.9.0, **cuando** se ejecutan `--run`, `--daemon`, `--scan`, `enroll`, `status` o `project join`, **entonces** su conducta
   y su salida son las de la 0.9.0. Sólo cambian la ayuda y «Verbos disponibles».

### Edge Cases
- **Una herramienta sin carpeta**: su lector no está activo, como en `--run`. No dice nada y no falla.
- **Registros que la emisión no emite** *(`.zst` de Codex, Gemini < 0.39.0, OpenCode 1.x, `<synthetic>` de Claude Code)*: no cuentan, como en `--run`.
- **Un fichero ilegible de Codex, Gemini, OpenCode, Copilot CLI o Kimi Code**: el aviso de su lector por stderr, y la lista sigue. **De Claude Code**: la
  lista se detiene con `error: <error>`, como `--run` *(P-6)*.
- **Un `event_id` repetido** *(una sesión de Claude Code reanudada)*: cuenta una vez *(P-2)*.
- **Un `cwd` con forma de Windows leído en Linux** *(una copia)*: no existe, así que sale `(ya no existe)`, con la ruta tal cual.
- **La carpeta es el propio directorio personal**: `~` *(P-10)*.
- **Una carpeta `<hogar>2/x`, hermana del hogar**: no cuelga del hogar; sale la ruta completa *(P-10)*.
- **`project list` con argumentos**: error de uso *(P-7)*.
- **`project_ref_mode: "plain"` en la configuración**: la lista ni se detiene ni avisa *(P-8)*.

## Requirements

### A · El comando *(D-9, D-11, P-7)*
- **FR-001** *(verbo)*: `permea project` **DEBE** aceptar el verbo `list`. Los dos mensajes de error del verbo **DEBEN** decir `Verbos disponibles: join, list`
  *(`project.go:98`, `:108`)*.
- **FR-002** *(argumentos — P-7)*: el comando **NUNCA** acepta argumentos, salvo `-h` y `--help`. Con otro, **DEBE** dar el error de uso de §Textos, con
  salida 1, y **NUNCA** reproduce lo tecleado.
- **FR-003** *(ayuda — D-11)*: `permea project list -h` y `--help` **DEBEN** dar la ayuda de §Textos y **NUNCA** ejecutan nada. La fila **DEBE** aparecer
  también en la ayuda general y en la de `permea project`. Ninguna línea **DEBE** pasar de 80 runas.

### B · Lo que nunca hace *(D-6)*
- **FR-004** *(red)*: el comando **NUNCA** abre una conexión de red ni consulta la plataforma, esté o no enrolada la instalación.
- **FR-005** *(escritura)*: el comando **NUNCA** crea, escribe, renombra ni borra ningún fichero ni directorio: ni la cola, ni `state.json`, ni la sal, ni
  `machine_id`, ni `config.json`, ni el `DataDir`, ni nada bajo las carpetas de las herramientas.
- **FR-006** *(sin enrolamiento — P-8)*: el comando **DEBE** funcionar igual sin `endpoint` ni `device_token`, y **NUNCA** se detiene ni avisa por
  `project_ref_mode: "plain"`.

### C · La sal *(D-5, P-11)*
- **FR-007** *(sin sal)*: si no existe el fichero de la sal, o el `DataDir`, el comando **DEBE** escribir el texto de P-11 y salir con 0. **NUNCA** crea la sal.
- **FR-008** *(sal ilegible)*: si la sal existe pero no se puede leer, el comando **DEBE** dar el error de §Textos, con salida 1.

### D · Qué se cuenta *(P-1, P-2, P-9)*
- **FR-009** *(recorrido)*: el comando **DEBE** recorrer los seis lectores con el **mismo código** que la emisión, sobre todo lo que conservan en disco,
  con un estado que empieza vacío y **NUNCA** se guarda *(P-9)*.
- **FR-010** *(identidad)*: la identidad de cada evento **DEBE** ser la que la emisión le estamparía en ese momento, por el **mismo cuerpo** de la derivación
  *(P-005 FR-005; `resolve.go:150-156`)*.
- **FR-011** *(sin identidad — P-1)*: un evento sin identidad **NUNCA** da fila ni cuenta.
- **FR-012** *(recuento — P-2)*: `EVENTOS` **DEBE** ser el número de `event_id` distintos de la identidad, según P-2.
- **FR-013** *(fechas)*: `DESDE` y `HASTA` **DEBEN** ser la fecha UTC *(`AAAA-MM-DD`)* del `occurred_at` más antiguo y del más reciente de la identidad.
- **FR-014** *(herramientas — D-1, P-3)*: `HERRAMIENTAS` **DEBE** ser el conjunto de `tool` de sus eventos, con el nombre comercial de D-1, en el orden de
  P-3 y separado por `, `.

### E · La carpeta *(D-2, D-3, P-4, P-10)*
- **FR-015** *(carpeta)*: `CARPETA` **DEBE** ser **la ruta que entró en el hash** de la identidad: la raíz del repositorio, o el directorio si no hubo raíz.
  La carpeta personal, según P-10.
- **FR-016** *(marcas — D-3, P-4)*: tras la carpeta, con un espacio, **DEBE** ir `(ya no existe)` si la ruta no existe hoy; si existe y la derivación no
  encontró raíz, `(sin repositorio)`; si no, nada. **NUNCA** van las dos.

### F · La tabla *(D-2, D-4, P-5, P-12)*
- **FR-017** *(columnas)*: la tabla **DEBE** llevar `IDENTIDAD`, `CARPETA`, `HERRAMIENTAS`, `EVENTOS`, `DESDE` y `HASTA`, con esa cabecera en mayúsculas.
- **FR-018** *(identidad)*: `IDENTIDAD` **DEBE** ser los **12** primeros caracteres de los 64 de la identidad, seguidos de `…`.
- **FR-019** *(miles)*: `EVENTOS` **DEBE** llevar `.` cada tres cifras desde la derecha *(«999», «1.240», «1.234.567»)*. Lo mismo en el pie.
- **FR-020** *(orden — D-2)*: las filas **DEBEN** ir por `EVENTOS`, de más a menos. Si empatan, por la identidad completa, de menor a mayor.
- **FR-021** *(formato — P-12)*: cada columna **DEBE** tener el ancho de P-12, con **dos espacios** entre columnas. `EVENTOS` va a la derecha, **también su
  cabecera**; las demás, a la izquierda, y la última, sin relleno. **Ninguna línea acaba en espacio.**
- **FR-022** *(pie — D-4, P-5)*: tras una línea en blanco, **DEBEN** ir las dos líneas de D-4. Con 1, `1 identidad` y `1 evento`. Sin eventos, según P-5.
- **FR-023** *(canales — P-6)*: la tabla **DEBE** salir por stdout, y los avisos de los lectores, por stderr, según P-6. **NUNCA** escribe las líneas de
  resumen de `--run`. Los errores **DEBEN** salir como `error: …` por stderr, con salida 1 y stdout vacío.

### G · Lo de antes no cambia *(D-9)*
- **FR-024** *(conducta)*: `--run`, `--daemon`, `--scan`, `enroll`, `status` y `project join` **DEBEN** comportarse como en la 0.9.0, **byte a byte**: lo
  encolado, en el mismo orden, y lo escrito por stdout y stderr. Cambian sólo las ayudas *(FR-003)* y los mensajes de FR-001.
- **FR-025** *(frontera)*: `git diff cabf8c5` **DEBE** ser **vacío** en `internal/event/`, `internal/ingest/`, `internal/state/`, `internal/transport/`,
  `internal/pricing/` y los contratos de 005, 006 y 008–012. Las dos frases de 005 y 006 que 013 deja atrás se declaran en §Dependencias, sin tocar los
  contratos.

### Key Entities
- **Identidad**: los 64 hex de `project_ref`, con la ruta que entró en su hash y si hubo raíz.
- **Fila**: una identidad con sus `event_id` distintos, sus herramientas y su `occurred_at` mínimo y máximo.

## Success Criteria

- **SC-001** *(el ejemplo, byte a byte)*: las cuatro filas del ejemplo de §Textos, dadas al formato, producen **exactamente** el texto de §Textos.
- **SC-002** *(los seis lectores)*: en un sandbox con fixtures **sintéticos** de las seis herramientas, la salida es exactamente la esperada:
  - una fila por identidad, con las herramientas de D-1 y los recuentos y fechas del fixture;
  - la identidad esperada se calcula en el test, con la sal del sandbox;
  - un fichero ilegible de Codex da su aviso por stderr, la tabla sale igual y no hay líneas de resumen.
- **SC-003** *(misma identidad que la emisión)*: sobre los mismos fixtures, el conjunto de identidades de la lista = el de `project_ref` no vacíos que `--run`
  encola en un sandbox gemelo.
- **SC-004** *(sin red)*: con `endpoint` en un servidor de prueba **en el que el binario confía**, `device_token` y un evento en la cola: **0** peticiones y
  **0** conexiones TCP contadas en el propio servidor.
- **SC-005** *(sin escritura)*: la huella del sandbox entero *(cada ruta, con tipo, tamaño, `mtime` y md5)* es **igual** antes y después. Con el `DataDir`
  borrado, sigue sin existir.
- **SC-006** *(la sal)*: sin sal, el texto de P-11 por stdout, stderr vacío, salida 0, y el fichero de la sal sigue sin existir. Con la sal ilegible, el
  error de §Textos, salida 1 y stdout vacío.
- **SC-007** *(marcas)*: con `.git`, sin marca; sin `.git`, `(sin repositorio)`; borrada, `(ya no existe)`; borrada y con `.git` antes, sólo `(ya no existe)`.
- **SC-008** *(orden)*: 7 · 412 · 412 · 1 240 eventos salen como 1.240, 412, 412 y 7, y los dos de 412, por identidad. Una identidad con eventos de Kimi Code CLI,
  un id desconocido `zeta` y Claude Code sale `Claude Code, Kimi Code CLI, zeta`.
- **SC-009** *(recuento)*: un fichero de Claude Code y su copia en otra carpeta del proyecto *(reanudar)* cuentan los mensajes **una vez**. Una identidad
  hecha sólo de repetidos no sale.
- **SC-010** *(números y singular)*: 0 → `0`, 999 → `999`, 1000 → `1.000`, 1234567 → `1.234.567`. Con 1 identidad y 1 evento: `1 identidad, 1 evento.`
- **SC-011** *(la carpeta personal)*: el hogar → `~`; `<hogar>/dev/x` → `~/dev/x`; `<hogar>2/x` → la ruta completa; un hogar con un enlace en medio → `~/…`.
- **SC-012** *(sin regresión)*: sobre los fixtures de los seis lectores, la cola, el stdout y el stderr de `--run` son **byte a byte** los del binario de
  `cabf8c5`, con la ruta del sandbox sustituida por un marcador. Los **1074** tests de `cabf8c5`, en PASS **por nombre** y **sin tocarlos**.
- **SC-013** *(ayuda)*: los textos de §Textos, byte a byte; ≤ 80 runas por línea; sin jerga interna; `-h` no escribe nada.
- **SC-014** *(argumentos)*: `permea project list x` → el texto de §Textos por stderr, stdout vacío y salida 1. Con un argumento con forma de secreto, el
  secreto no aparece.
- **SC-015** *(puertas)*: `go vet` limpio; `golangci-lint run` → 0; FR-025 con 0 bytes de diff; `go.mod` sin `require`; sin `go.sum`.
- **SC-016** *(binario)*: crece **≤ 50 KB** frente al de la 0.9.0, en las 5 plataformas, con el instrumento de 010 E-3 *(ratificado el 23-10)*.
- **SC-017** *(privacidad del repo)*: el patrón `PRIV` de 012 da **0** en el diff de 013 y en sus testdata.
- **SC-018** *(en este equipo, en lectura)*: sobre los registros reales de Claude Code de WSL, la lista termina en **≤ 2 s**, y su número de identidades =
  el de un recuento independiente hecho **el mismo día** *(con el método de desc. Q4)*. Sólo se transcriben **recuentos y tiempos**.
- **SC-019** *(el cruce con el panel — W1, del dueño)*: en su instalación real, cada identidad del panel **sin asignar** de esa instalación con `last_seen` ≥
  la primera fecha que conservan las herramientas aparece en la lista, con los mismos 12 caracteres. Lo que no aparezca se explica con §Limitaciones.

## Textos *(letra a letra)*

**La salida** *(stdout; el ejemplo aprobado, con los nombres comerciales)*:
```
IDENTIDAD      CARPETA                            HERRAMIENTAS        EVENTOS  DESDE       HASTA
08926050c4b8…  ~/dev/proyecto-ejemplo             Claude Code, Codex    1.240  2026-10-03  2026-10-23
3f1c9a77d02e…  ~/dev/otro-proyecto                Claude Code             412  2026-10-12  2026-10-22
b7e40d915a63…  ~/pruebas/tmp-1 (sin repositorio)  Gemini CLI               18  2026-10-20  2026-10-20
c05d2e8a1f94…  ~/dev/viejo (ya no existe)         Claude Code               7  2026-10-05  2026-10-06

4 identidades, 1.677 eventos. Fechas en UTC, como en el panel.
Cuenta lo que las herramientas conservan en disco, también lo aún no enviado: el panel puede tener más o menos.
```
**El pie**, en singular: `1 identidad, 1 evento. Fechas en UTC, como en el panel.` *(cada mitad, por su número)*.

**Sin sal** *(stdout, salida 0)*:
```
Esta instalación aún no ha medido nada.
```

**La ayuda** *(fila de la general y de `permea project -h`; sangrías 2 y 6, como `cmd/permea/ayuda.go:33-34`)*:
```
  project list
      Enseña los proyectos que ha medido esta instalación: la identidad
      que ves en el panel, su carpeta, las herramientas, los eventos y las
      fechas. Sirve para saber de qué carpeta es cada identidad sin
      asignar. Sólo lee lo que conservan las herramientas: no envía nada
      ni cambia lo que queda por enviar.
```
**`permea project list -h`**: `uso: permea project list`, una línea en blanco y la fila de arriba.

**Los verbos** *(stderr, salida 1)*:
```
error: falta el verbo. Verbos disponibles: join, list
error: verbo desconocido "<verbo>". Verbos disponibles: join, list
```

**Errores del comando** *(stderr, salida 1, stdout vacío)* `DECIDÍ YO`:
```
error: project list no admite argumentos. Ayuda: permea project list -h
error: la configuración local no se puede leer: <error>
error: la sal de esta instalación no se puede leer: <error>
error: <error>
```
La última es la de un error de lectura de Claude Code *(P-6)*: el mismo `error:` que `--run` escribe *(`main.go:134-137`)*.

## Limitaciones *(se declaran; README y CHANGELOG las recogen)*

- **D-7 · Sin las identidades de antes de la 0.3.0.** Las que envió una 0.2.x, calculadas con el `cwd` crudo *(desc. Q2)*, no se listan. Cuando se lanzaba
  desde la raíz y sin enlaces, coinciden con la de hoy. En este equipo no queda ningún registro anterior al 2026-10-02 *(desc. Q4)*.
- **El sistema de ficheros de hoy** *(desc. Q1)*: la identidad se recalcula con las carpetas **como están ahora**. Si una carpeta se movió, si se hizo
  `git init` después, o si cambió un enlace, la identidad listada **no es** la que viajó. La de una carpeta borrada lleva `(ya no existe)`, y también puede
  diferir. Las otras dos no se detectan.
- **El panel puede tener más o menos** *(lo dice el pie)*: más, por lo que las herramientas ya borraron, lo que envió la 0.2.x por línea y lo de otras
  instalaciones, que tienen otra sal; menos, por lo aún no enviado.
- **Un mensaje de Claude Code todavía abierto** *(menos de 10 min sin cambios)* no cuenta hasta que se cierra *(P-9)*.
- **Windows**: la misma carpeta escrita con otras mayúsculas da dos identidades *(`resolve.go:119-122`; 012 §Límites)*.
- **Los caracteres de doble ancho** *(CJK, emoji)* en una ruta descuadran la tabla *(P-12)*.

## Constitución *(1.1.0)*

| Principio | | Razón |
|---|:--:|---|
| I · Frontera inviolable | ✅ | **No cruza nada**: el comando no abre la red *(FR-004)*, y el evento y la frontera no cambian *(FR-025)*. **D-10**: es el **primer comando que imprime rutas reales**, y sólo a la terminal de quien lo ejecuta, en su máquina. El Principio I prohíbe que salgan de la máquina, y no salen. El plan lo vigila con SC-004 y SC-005 |
| II · Local-first | ✅ | Sólo lectura local, sin red y sin enrolamiento *(FR-004 a FR-006)* |
| III · Binario único | ✅ | Sin dependencias *(SC-015)*; rutas por sistema *(P-10)* |
| IV · Test-first en la frontera | ✅ | `boundary_test.go` no cambia. La prueba de «nada sale y nada se escribe» *(SC-004, SC-005)* es roja antes del verde: afirma también la tabla, que el esqueleto no escribe |
| Fixtures | ✅ | Sintéticos; ninguna ruta real *(SC-017)* |
| V · Specs | ⚠️ | Spec, plan y tareas antes del código. **Pero** la spec lleva citas `fichero:línea` y algún detalle del cómo *(P-8, P-12)*, como 008–012. El Principio V pide «sin detalles de implementación»: es la convención de la casa. **El dueño (23-10)**: se sigue, y la enmienda menor queda como candidata en el backlog *(`permea-platform/docs/backlog.md`, §Deuda del repo del AGENTE)*; la constitución no se toca |

## Censo de tests *(detalle en el plan §7)*

**Ninguno de los 1074 existentes cambia.** Ninguno afirma «Verbos disponibles: join» literal: `TestProject_ErroresDeUsoDeLaGramatica` sólo exige
salida 1, stderr no vacío y stdout vacío *(`cmd/permea/project_test.go:555-583`)*. `lasCincoAyudas` *(`ayuda_test.go:299-307`)* no se amplía: la sexta
ayuda tiene sus pruebas en un fichero nuevo. Si alguno tuviera que cambiar, **PARA**.

## Fuera de alcance

| # | Qué | Por qué |
|---|---|---|
| **N-1** | `--json` | D-8 |
| **N-2** | `project join` con carpeta o con hash | D-9; desc. Q7 |
| **N-3** | Las identidades de la 0.2.x | D-7 |
| **N-4** | Marcar las identidades ya unidas | No hay dato local *(desc. Q5)*, y preguntar a la plataforma es red *(D-6)* |
| **N-5** | Otra instalación u otro usuario | Otra sal *(desc. Q2)* |
| **N-6** | Filtros, por herramienta o por fecha | No pedidos |

## Dependencias
- **La plataforma**: el nombre `copilot_cli` → **GitHub Copilot CLI** falta en el mapa del panel
  [P `frontend/src/catalogo/es/Operador.ts:21-27`; copia en `en/Operador.ts`]. **Anotado, no bloquea**: el agente lo nombra por su cuenta *(D-1)*.
- **005**: el cuerpo compartido de la derivación *(P-005 FR-005)*, que esta feature amplía con una salida más, sin otro cuerpo.
- **Dos frases que 013 deja atrás, sin tocar sus contratos** *(FR-025)*:
  - «`project` queda con **un solo verbo** hasta que alguna spec añada otro» *(`specs/005-adhesion-a-proyecto/contracts/cli.md:170`)*: 013 es esa spec;
  - «ayuda de `project` (lista `join`)» *(`specs/006-medicion-fiel/contracts/cli.md:64`)*: desde 013 lista `join` y `list`.

## Preguntas abiertas

| # | Pregunta | Mientras tanto | Cómo se cierra |
|---|---|---|---|
| **Q-1** ✅ *(cerrada: el dueño, 23-10)* | ¿Qué versión? | — | **0.10.0** |
| **Q-2** | ¿Tarda lo mismo con Codex, Gemini, OpenCode, Copilot y Kimi? | La referencia de OpenCode, ≈ 1 ms/MB *(desc. Q4)* | W1, en Windows, con lo que haya |

## Registro de enmiendas

| # | Fecha | Qué cambia | Por qué |
|---|---|---|---|
| **Borrador** | 2026-10-23 | Primera redacción, con D-1 a D-11 del dueño, y P-1 a P-12 y Q-1 a Q-2 para ratificar | Encargo 013 · B1 |
| **Analyze** | 2026-10-23 | Redacción, sin tocar D-1 a D-11: DEBE/NUNCA en cada FR y «cuando» en cada escenario; P-2 precisa los repetidos; P-6 cita los avisos reales y que un error de lectura de Claude Code detiene la lista, como `--run`; P-8 nombra `project_ref_mode: "plain"`; FR-021, la cabecera de `EVENTOS` a la derecha; FR-025 suma `internal/state/` e `internal/pricing/`; SC-002, SC-004, SC-006, SC-009, SC-012 y SC-018 se vuelven comprobables; §Dependencias, las dos frases de 005 y 006 | Encargo 013 · B1, `/speckit-analyze` |
| **Ratificación** | 2026-10-23 | **El dueño ratifica** P-1 a P-12 tal cual, la **0.10.0** *(Q-1 cerrada)* y SC-016 *(≤ 50 KB)*. **D-4**: «Antes: `Cuenta lo que las herramientas conservan en disco: el panel puede tener más.`» → **Después**: `Cuenta lo que las herramientas conservan en disco, también lo aún no enviado: el panel puede tener más o menos.` §Limitaciones, «el panel puede tener más» y «puede tener menos», en una línea. «1.240» se queda con el punto, como el panel *(«3.908 mensajes»)*. La costumbre de citar `fichero:línea`, al backlog | Encargo 013 · B1, ratificación |
