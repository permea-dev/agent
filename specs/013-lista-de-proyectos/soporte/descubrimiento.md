# 013 · Descubrimiento de `permea project list` *(B0, sin código)*

**Fecha**: 2026-10-23 · **HEAD**: `cabf8c5` en `013-lista-de-proyectos` *(igual que `main`)* · **Plataforma leída**: `permea-platform` en `f473ae0`,
sólo lectura.

**Cómo se lee esto.** Cada afirmación lleva su fuente:
- **[A `fichero:línea`]** = leído en el código del agente, en HEAD; **[A@vX `fichero:línea`]** = en la etiqueta `vX`;
- **[P `fichero:línea`]** = leído en el código de la plataforma *(rutas relativas a `permea-platform/`)*;
- **[M]** = medido en este equipo el 2026-10-23 *(§Q4)*.

> **Privacidad.** Se midieron sólo tiempos y recuentos. Ninguna ruta, nombre de carpeta ni contenido de un registro entra en este
> documento. No se calculó ningún hash con la sal real: la comparación de identidades antiguas y nuevas se hizo comparando las
> cadenas que entran en el hash, que con la misma sal es equivalente. El binario de medida se compiló desde HEAD a un directorio
> temporal fuera del repositorio.

---

## Q1 · La derivación de hoy

**Qué entra.** Tres cosas: el **directorio de trabajo declarado** por el registro de la herramienta, la **sal** de la instalación y el
**sistema de ficheros tal como está en ese momento**.

1. `mejorValorDisponible`: la ruta real si `EvalSymlinks` la resuelve; si no, la declarada. En los dos casos, `filepath.Clean`
   [A `internal/project/resolve.go:123-131`]. No expande `~`, no aplica `Abs`, no pasa a minúsculas [A `resolve.go:112-118`].
2. `ascender`: sube buscando un `.git` *(fichero o directorio)* y para **antes** del techo: el directorio personal si la ruta cuelga
   de él, la raíz del sistema si no [A `resolve.go:185-203`, `:220-223`, `:241-249`].
3. Con raíz: `event.Ref(sal, raíz)`. Sin raíz *(fallback)*: `event.Ref(sal, ubicación)` [A `resolve.go:157-171`].
4. `event.Ref` = `hex(SHA-256(sal + ":" + valor))`, 64 caracteres en minúscula; vacío si el valor es vacío
   [A `internal/event/event.go:40-46`].

**Las seis herramientas pasan por el mismo punto**, `ctx.Resolutor.Derivar(cwd, sal)` [A `internal/project/resolve.go:318-328`]:

| Herramienta | De dónde sale el `cwd` | Cita |
|---|---|---|
| Claude Code | `cwd` de cada línea | [A `internal/ingest/claudecode.go:125`] |
| Codex | `cwd` del `turn_context` del turno; si no, el de `session_meta` | [A `internal/ingest/codex_contexto.go:135-144`; `codex.go:150`] |
| Gemini | `.project_root` de la carpeta de la sesión | [A `internal/ingest/gemini_contexto.go:149-161`; `gemini.go:147`] |
| OpenCode | `directory` de la fila de `session_v2` | [A `internal/ingest/opencode_base.go:297-300`, `:425-427`; `opencode.go:134`] |
| Copilot CLI | `data.context.cwd` del último `session.start`/`resume` | [A `internal/ingest/copilot_contexto.go:108-119`; `copilot.go:155`] |
| Kimi Code | `workDir` de `session_index.jsonl`, con las barras normalizadas | [A `internal/ingest/kimi_contexto.go:229`, `:314-320`; `kimi.go:127`] |

**Dónde vive la sal.** En `<DataDir>/salt`: 32 bytes aleatorios en hexadecimal, permisos 0600, creada la primera vez que se pide
[A `internal/config/identity.go:13-14`, `:24-41`]. `DataDir` es `$XDG_CONFIG_HOME/permea` *(Linux)*, `~/Library/Application
Support/permea` *(macOS)* o `%AppData%\permea` *(Windows)* [A `internal/config/config.go:51-60`]. La leen `setup()`
[A `cmd/permea/main.go:232`], `enroll` [A `cmd/permea/enroll.go:110`] y `project join` [A `cmd/permea/project.go:221`].
⚠️ Las tres usan `LoadOrCreateSalt`, que **la crea si no existe**: un comando de sólo lectura no debería usarla tal cual (§Q6).

**¿Se puede recalcular?** **Sí.** Con la sal y el `cwd` que declara cada registro, el comando obtiene exactamente el ref que hoy
estamparía la ingesta, porque llamaría al **mismo cuerpo** (`derivarConTechoYRaiz`, el único de todo el paquete, P-005 FR-005)
[A `resolve.go:150-156`]. **Con una condición que hay que decir**: el ref depende del sistema de ficheros **en el momento de
derivarlo**. Se recalcula igual que se envió mientras la carpeta siga donde estaba y con su `.git`. Si después se borró o se movió,
se hizo `git init` encima o cambió un enlace, el ref de hoy **ya no es** el que viajó [A `resolve.go:123-131`, `:157-171`]. En este
equipo, de 18 carpetas declaradas en WSL, **ninguna ha desaparecido** [M].

**Lo que falta para enseñar la carpeta.** Las funciones públicas devuelven el ref y, como mucho, si hubo raíz
(`DerivarConRaiz`) [A `resolve.go:80-90`], pero no **qué ruta entró en el hash**. El `Resolutor` guarda `cwd declarado → ref`
[A `resolve.go:300-328`], y con eso basta para agrupar, pero no para enseñar la raíz. Hará falta una salida más del mismo cuerpo
*(la ruta que entró en el hash)*, no un cuerpo nuevo.

**Cuántos caracteres.** **12**, en minúscula, seguidos de `…`: es lo que pinta el panel, `key.slice(0, 12)` + `…`
[P `frontend/src/screens/metrics/PorProyecto.tsx:1153-1157`; también `piezas/IdentidadesDeProyecto.tsx:108`, `:200`]. La
plataforma guarda el ref tal cual llega, sólo si casa con `^[0-9a-f]{64}$` *(si no, lo deja en NULL)*
[P `backend/app/Support/Identity/ProjectRefFormat.php:50`; `backend/app/Ingest/EventAllowlist.php:180-182`]. Con 48 bits, una
colisión entre las identidades de una organización es despreciable.

**El panel enseña, además, eventos y fechas.** Debajo del hash pinta `event_count` y `first_seen → last_seen`
[P `PorProyecto.tsx:1137-1141`]. Salen de la vida completa del ref: `COUNT(*)`, y `MIN`/`MAX(occurred_at)` como **fecha UTC
`YYYY-MM-DD`** [P `backend/app/Services/Metrics/MetricsQuery.php:478-495`]. Para cruzar con comodidad, el comando debería dar
**las mismas tres cosas, en la misma forma**.

---

## Q2 · Las identidades de antes de la 0.3.0

**Cómo se calculaban.** En la 0.2.x: `ProjectRef = event.Ref(sal, r.Cwd)` con el `cwd` **crudo** de cada línea, sin `Clean`, sin
resolver enlaces y sin subir a la raíz [A@v0.2.1 `internal/ingest/claudecode.go:78`]. Sólo había lector de Claude Code
[A@v0.2.1, `git ls-tree`]. `event.Ref` es idéntica byte a byte en v0.2.0, v0.2.1 y v0.3.0 [A@v0.2.0/v0.2.1/v0.3.0
`internal/event/event.go:40-46`].

> **Nota sobre `CheckRetiredProjectRefMode`.** No tiene que ver con la derivación: detecta el modo `plain` *(envío en claro)*, que
> **nunca se implementó** [A `internal/config/config.go:13-28`, `:167-197`]. En la 0.2.x se enviaba siempre el hash: el `plain` no
> lo consultaba ningún camino [A@v0.2.1 `claudecode.go:78`].

**¿Se pueden recalcular con la misma sal?** **Sí, y es trivial**: `event.Ref(sal, cwdDeclarado)` sobre la cadena tal cual viene
en la línea. Pero valen poco, por tres razones:

1. **Muchas coinciden con la de hoy.** Si se lanzaba desde la raíz del repositorio y sin enlaces, el `cwd` crudo **es** la raíz y
   los dos refs son el mismo. Sólo difieren las de subdirectorios, enlaces o rutas sin limpiar. En WSL: de 18 `cwd` distintos, 2
   dan el mismo ref y 16 no [M].
2. **Los registros de entonces ya no están.** En los dos perfiles de este equipo, la primera línea facturable que queda es del
   **2026-10-20**. Ninguna es anterior al 2026-10-02, día de la 0.3.0 [M]. Para recalcular una identidad antigua hace falta el
   registro que la produjo.
3. **No se sabe qué eventos salieron con la 0.2.x.** El ref lo decide la versión que **envió**, no la fecha del evento, y el
   agente no apunta cuándo se actualizó (§Q5).

**¿Ha cambiado alguna vez la sal entre versiones?** **No por código.** El fichero (`<DataDir>/salt`) y la función son los mismos
desde el MVP (`67e6963`) [A `git log --follow internal/config/identity.go`]. `a5d7204` cambió **cuándo nace** *(en el
enrolamiento)*, no su valor. Ningún camino la reescribe: `loadOrCreateSecret` sólo escribe si el fichero no existe
[A `identity.go:24-41`]. **Cambia de hecho** si se borra el fichero o cambia el `DataDir`: otro usuario del sistema, otro
`XDG_CONFIG_HOME`, un perfil de Windows rehecho o **otro equipo**. Así, una misma carpeta en dos equipos son dos identidades.

**¿Qué ref tiene hoy la plataforma para los eventos viejos?** **El de la versión que los envió.** Razones:
- el agente no reenvía lo enviado: los confirmados salen de la cola [A `internal/transport/queue.go:84-128`] y los cursores impiden
  releer [A `internal/state/state.go:124-…`]. Así lo promete el CHANGELOG de la 0.3.0: «no se corrige ni se reenvía»
  [A `CHANGELOG.md`, §0.3.0];
- si se reenviara un evento ya recibido *(por ejemplo, tras borrar `state.json`)*, la plataforma **se queda con el primero**:
  `INSERT … ON CONFLICT (org_id, event_id) DO NOTHING` [P `backend/app/Ingest/IngestBatchService.php:10`;
  `backend/database/migrations/2026_07_04_000004_add_unique_org_event_to_metric_events.php:16`];
- ⚠️ **salvo con lo de la 0.2.x**: allí el `event_id` era aleatorio [A@v0.2.1 `internal/event/event.go:49-55`] y desde la 0.3.0 es
  determinista [A `CHANGELOG.md`, §0.3.0]. Una resubida de esos mensajes **no deduplica**: entra otra vez, con el ref nuevo. En ese
  caso, la plataforma tendría los dos.

---

## Q3 · Carpetas, eventos y fechas sin enviar nada y sin tocar el estado

**Todos los lectores se pueden recorrer en modo lectura hoy**, con las piezas que ya existen. Los seis siguen el mismo patrón:
`Listar…(raíz)` y `Leer…(st, ruta, contexto, pasada) → []event.Event`. El `*state.Store` se inyecta: con `state.New()`, un estado sólo
en memoria que **nunca se guarda**, leen desde el principio y no mueven ningún cursor [A `internal/state/state.go:48`]. **La cola
sólo la escribe quien llama**, con `transport.Append` [A `cmd/permea/main.go:348`, `:370`]. Así lo hace ya `--scan` en cada lector:

| Herramienta | Listar | Leer | `--scan` lo hace así | Lo que no da eventos |
|---|---|---|---|---|
| Claude Code | `state.FindLogs` [A `internal/state/scan.go`] | `FromClaudeCodeLine` + `Pasada.CerrarFichero` | [A `cmd/permea/main.go:668-705`] | líneas no facturables, `<synthetic>` |
| Codex | `ListarCodex` [A `internal/ingest/codex_contexto.go:258`] | `LeerFicheroCodex` [A `:150`] | [A `cmd/permea/main.go:749-760`] | `.zst` *(se cuentan, no se leen)*; < 0.153.0 |
| Gemini | `ListarGemini` [A `internal/ingest/gemini_contexto.go:97`] | `LeerFicheroGemini` [A `:243`] | [A `cmd/permea/gemini.go:97-106`] | < 0.39.0 *(se cuentan)* |
| OpenCode | `ListarOpenCode` [A `internal/ingest/opencode_base.go:100`] | `LeerBaseOpenCode` [A `:159`], base en sólo lectura | [A `cmd/permea/opencode.go:94-98`] | bases 1.x; llamadas abandonadas |
| Copilot CLI | `ListarCopilot` [A `internal/ingest/copilot_contexto.go:29`] | `LeerFicheroCopilot` [A `:127`] | [A `cmd/permea/copilot.go:102-106`] | llamadas auxiliares |
| Kimi Code | `LeerIndiceKimi` + `ListarKimi` [A `internal/ingest/kimi_contexto.go:290`, `:30`] | `LeerFicheroKimi` [A `:201`] | [A `cmd/permea/kimi.go:95-99`] | lo copiado por `/fork` |

Cada evento trae lo que hace falta para la fila: `ProjectRef`, `Tool` y `OccurredAt` [A `internal/event/event.go:17-35`]. El `cwd` no
viaja en el evento, pero pasa por el `Resolutor` (§Q1).

**¿Lo hace ya `--scan`?** **En parte.** Lee en modo lectura y aplica las mismas reglas que la emisión, pero:
- va **fichero a fichero** y no enumera la raíz de cada herramienta [A `cmd/permea/main.go:621-660`];
- usa una sal falsa, `"dry-run-salt"` [A `main.go:669`, `:751`; `gemini.go:99`; `kimi.go:97`; `opencode.go:96`; `copilot.go:104`],
  así que sus refs **no casan con el panel**;
- trunca el ref a 8 caracteres, no 12 [A `main.go:676-679`, `:736-739`], y no enseña el `cwd` ni agrega.

**Dos caminos, y cuánto cuesta cada uno:**

| | A · desde los registros de cada herramienta | B · desde el estado local del agente |
|---|---|---|
| Lo que hay hoy | Todo: los seis lectores, el `Resolutor` y el patrón de `--scan` | **Nada que sirva**: la cola sólo guarda **lo no enviado**, porque los confirmados se borran [A `internal/transport/queue.go:84-128`]. `state.json` guarda cursores y huellas por fichero, sin ref, `cwd` ni recuentos [A `internal/state/state.go:16-45`] |
| Lo que falta | Sacar de `generate()` un recorrido común con un **sumidero** *(encolar, o agregar en memoria)* y sin `st.Save` [A `cmd/permea/main.go:300-436`]; una salida más del cuerpo de la derivación *(§Q1)*; leer la sal sin crearla | Un registro **nuevo** y persistente de lo enviado, con rutas en claro en disco. Sólo valdría **desde que se despliegue**: el histórico, que es lo que se pide, no lo tendría |
| Cubre | Lo que las herramientas **siguen guardando** en disco | Lo enviado desde la versión que lo traiga |
| Riesgo | Recalcula con el sistema de ficheros de hoy (§Q1) | Más estado que mantener; rutas en reposo; cambia lo que el agente guarda |

**Recomendación: A.** Es la única que responde al histórico, reutiliza la ingesta sin tocar su semántica y no cambia lo que el
agente guarda.

---

## Q4 · El coste *(medido el 2026-10-23)*

**Método.** Binario compilado desde HEAD a un directorio temporal. `--scan` sobre cada fichero `.jsonl` que encontraría `FindLogs`,
dos veces seguidas. Cada fichero arranca un proceso, así que un recorrido dentro de un solo proceso sería igual o más rápido. Los
recuentos de `cwd` y fechas salieron de `jq`, y se imprimieron **sólo números**.

| Herramienta | Dónde | Ficheros | Bytes | Eventos | 1.ª pasada | 2.ª pasada |
|---|---|---:|---:|---:|---:|---:|
| Claude Code | WSL *(ext4)* | 13 | 49,3 MB | 1.767 | 0,27 s | 0,24 s |
| Claude Code | perfil de Windows, leído desde WSL por `/mnt/c` | 12 | 72,3 MB | 2.468 | 1,01 s | 0,69 s |
| Copilot CLI | perfil de Windows | 1 *(0 bytes)* | 0 | 0 | — | — |
| Codex, Gemini, OpenCode, Kimi Code | — | **no hay carpeta** ni en WSL ni en el perfil de Windows | | | | |

**Otros recuentos [M]:**
- **Líneas facturables**: WSL 4.049 y Windows 6.174, para 1.767 y 2.468 eventos. La razón ≈ 2,3 casa con el «×2,13» de la 0.3.0.
- **`cwd` distintos**: WSL 18, que dan **2 identidades** de hoy. Windows 30. Desde WSL no se puede derivar con rutas `C:\`, así que
  sus identidades no se midieron.
- **Ficheros con más de un `cwd`**: 5 de 13 en WSL y 6 de 12 en Windows. Claude Code cambia el `cwd` dentro de una sesión. Por eso
  la fila tiene que ser **por identidad**, no por `cwd` (§Q6).
- **Fechas en disco**: del 2026-10-20 al 2026-10-23 en los dos perfiles. Ninguna línea anterior al 2026-10-02.
- **Derivación**: una resolución por `cwd` distinto y pasada, con el `Resolutor` [A `resolve.go:278-299`]. 18 son despreciables
  frente a la lectura.

**Para las herramientas sin muestra**, la referencia publicada es OpenCode: «6,3 MB en 6 ms de mediana *(≈ 1 ms/MB)*»
[A `specs/010-lector-opencode/spec.md:307`]. Para los JSONL de Codex, Gemini, Copilot y Kimi no hay cifras de pasada completa.
Se espera el mismo orden que Claude Code *(≈ 5 ms/MB en WSL)*, pero **no está medido**.

---

## Q5 · ¿Sabe el agente qué carpetas se unieron?

**No.** `project join` deriva el ref de `os.Getwd()`, lo presenta y escribe la respuesta. No guarda nada en local
[A `cmd/permea/project.go:131-261`: sólo `config.Load`, `LoadOrCreateSalt` y la petición]. Así lo dice el CHANGELOG de la 0.3.0: «no
guarda nada en local». La asignación sólo existe en la plataforma (`ProjectIdentityMapping`) [P
`backend/app/Models/ProjectIdentityMapping.php`]. Saberlo exigiría preguntar a la red, y el comando es local. **El comando no podrá
marcar las unidas.**

---

## Q6 · La salida propuesta *(borrador para aprobar)*

**Forma.**
- **Una fila por identidad.** Dentro de una instalación, identidad y carpeta son 1:1. Lo que no es 1:1 es `cwd`↔identidad (§Q4).
  La herramienta va como dato, y si hay varias, separadas por comas.
- **Columnas**: `IDENTIDAD` *(12 + `…`, como el panel)* · `CARPETA` *(la raíz, o el directorio si no hay repositorio; el home como
  `~`)* · `HERRAMIENTAS` *(el `tool` del evento: `claude_code`, `codex`, `gemini`, `opencode`, `copilot_cli`, `kimi_code`)* ·
  `EVENTOS` · `DESDE` · `HASTA` *(fecha UTC `AAAA-MM-DD`, como el panel)*.
- **Orden**: por eventos, de más a menos. Si empatan, por identidad.
- **Marcas** tras la carpeta:
  - `(sin repositorio)`: es fallback, y `project join` la rehusaría [A `cmd/permea/project.go:157`];
  - `(ya no existe)`: el ref enseñado puede no ser el que viajó (§Q1).
- **Pie**: el total y una línea que avisa de que el panel puede contar más.
- **Sin sal**: no se crea. Se dice «esta instalación aún no ha medido nada» y se sale con 0.
- **No exige enrolamiento ni red**, no pasa por `setup()` y no escribe nada.

**Ejemplo inventado:**

```
$ permea project list
IDENTIDAD      CARPETA                            HERRAMIENTAS        EVENTOS  DESDE       HASTA
08926050c4b8…  ~/dev/proyecto-ejemplo             claude_code, codex    1.240  2026-10-03  2026-10-23
3f1c9a77d02e…  ~/dev/otro-proyecto                claude_code             412  2026-10-12  2026-10-22
b7e40d915a63…  ~/pruebas/tmp-1 (sin repositorio)  gemini                   18  2026-10-20  2026-10-20
c05d2e8a1f94…  ~/dev/viejo (ya no existe)         claude_code               7  2026-10-05  2026-10-06

4 identidades, 1.677 eventos. Fechas en UTC, como en el panel.
Cuenta lo que las herramientas conservan en disco: el panel puede tener más.
```

**Identidades de antes de la 0.3.0.** Propuesta: **no enseñarlas** en la primera versión, y decirlo en las limitaciones. Coste:
calcularlas es una línea, pero no hay registros de entonces (§Q2) y, donde los haya, la mayoría coincide con la identidad de hoy. Si
el dueño las quiere, irían tras una opción, `--antiguas`: sólo Claude Code, una fila por `cwd` crudo y sólo las que difieren de la de
hoy, con una nota de que pueden no estar en el panel.

**La ayuda** *(líneas ≤ 80 caracteres, sin jerga, con la forma de `ayuda.go`)* [A `cmd/permea/ayuda.go:25-34`, `:76-85`]:

```
  project list
      Enseña los proyectos que ha medido esta instalación: la identidad
      que ves en el panel, su carpeta, las herramientas, los eventos y las
      fechas. Sirve para saber de qué carpeta es cada identidad sin
      asignar. Sólo lee lo que conservan las herramientas: no envía nada
      ni cambia lo que queda por enviar.
```

Y la de subcomando: `uso: permea project list`. El mensaje «Verbos disponibles: join» pasa a «join, list»
[A `cmd/permea/project.go:98`, `:108`].

**`--json`.** **No hace falta.** Quien lo usa es una persona que cruza la tabla con el panel. Una salida para máquinas invita a
llevar rutas a otro sitio, y se puede añadir después sin romper nada.

**Un aviso de privacidad, para el plan.** Es el **primer comando que imprime rutas reales**: `--scan` nunca enseña el `cwd`. Sale
sólo a la terminal del usuario y no cruza la frontera, pero conviene decirlo en el plan.

---

## Q7 · Fuera de alcance: `project join` con una carpeta o con una identidad antigua

- **`project join` con una carpeta como argumento.** Técnicamente es cambiar `os.Getwd()` por la ruta dada: el rehúse sin raíz y el
  cuerpo compartido siguen valiendo [A `cmd/permea/project.go:148-157`, `:228`]. Tiene tres pegas:
  - choca con la gramática actual, `project join [<código>]` *(posicional)*, y haría falta una opción;
  - una carpeta que **ya no existe** cae al fallback y se rehúsa, que es justo el caso histórico que motivaría la opción;
  - una ruta existente pero equivocada une **en silencio la identidad equivocada**, el fallo que P-005 FR-006 quiso evitar.

  Aporta poco sobre `cd <carpeta> && permea project join`.
- **Con una identidad antigua** *(un hash)*. El agente enviaría un ref que **no ha derivado**, y se rompería la promesa de P-005
  FR-005: lo que se une es lo que la ingesta estampa. Además sobra: **el panel ya permite asignar cualquier identidad** a un
  Proyecto [P `backend/routes/api.php:237`; `backend/app/Http/Controllers/Api/V1/ProjectIdentityController.php:76-100`]. Con
  `project list` el dueño reconoce el hash y lo asigna desde allí. **Recomendación: ninguna de las dos.**
