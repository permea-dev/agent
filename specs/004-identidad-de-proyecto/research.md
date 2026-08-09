# Research · P-004 Identidad de proyecto

**Fecha**: 2026-08-09 · **Contra**: HEAD `2ba8363` · **Spec**: [spec.md](./spec.md) (congelada)

Este documento responde las siete preguntas de mecanismo que el plan necesita resueltas. Donde el
mecanismo tiene alternativas con coste, están las alternativas y una **recomendación explícita** — no
se zanja en silencio.

---

## Verificación previa contra HEAD — qué heredé y qué NO

El encargo pedía reverificar lo que D-1 dejó medido en vez de heredarlo. Resultado:

| Afirmación de D-1 | Verificación contra `2ba8363` | Veredicto |
|---|---|---|
| Choke point en `internal/source/claudecode.go:78` | **`internal/source/` NO EXISTE** (`ls internal/` → `config event ingest pricing state transport`). El choke point real es **`internal/ingest/claudecode.go:78`** | ⚠️ **ruta corregida**; la línea `:78` sí coincide |
| `Ref()` compartida por las tres identidades | `internal/ingest/claudecode.go:78-80` → `ProjectRef`, `SessionRef`, `MachineRef`, las tres por `event.Ref`; definición única en `internal/event/event.go:40` | ✅ confirmado |
| El `cwd` viene del log, no de `Getwd` | `rawRecord.Cwd` con etiqueta `json:"cwd"` (`internal/ingest/claudecode.go:28`); no hay ninguna llamada a `os.Getwd` en el repo | ✅ confirmado |
| `ProjectRefMode` consumido solo por `applyDefaults()` | Barrido completo abajo — **confirmado, con dos consumidores más que D-1 no citó** | ⚠️ **ampliado** |
| Alcance documental del `plain` = 4 sitios conocidos | Barrido completo abajo — **son 7, y tres no contienen la palabra «plain»** | ⚠️ **ampliado** |

### Inventario real de `ProjectRefMode` (barrido, no herencia)

`grep -rn "ProjectRefMode\|ModePlain\|ModeHash\|project_ref_mode" --include=*.go .`

| Fichero:línea | Qué es | Efecto al retirar |
|---|---|---|
| `internal/config/config.go:14-15` | comentario + `type ProjectRefMode string` | se borra |
| `internal/config/config.go:18-21` | `ModeHash` / `ModePlain` | se borran |
| `internal/config/config.go:30` | campo `ProjectRefMode` del struct `Config` | se borra |
| `internal/config/config.go:41` | valor por defecto en `Default()` | se borra |
| `internal/config/config.go:120-121` | relleno en `applyDefaults()` | se borra |
| **`internal/config/config_test.go:36`** | aserción `out.ProjectRefMode != ModeHash` | **el test deja de compilar** |
| **`internal/config/config_test.go:52`** | aserción `out.ProjectRefMode != ModeHash` | **el test deja de compilar** |

`Validate()` (`config.go:91-103`) **no lo mira** — confirmado: solo valida el esquema del endpoint.

Los dos tests son el añadido respecto de D-1, y no son un detalle: al retirar el campo **el paquete
`config` no compila** hasta tocarlos. Es trabajo obligatorio, no opcional, y va en el mismo commit.

### Alcance documental real del `plain` — **7 apariciones; 6 en alcance, 1 fuera**

El barrido por `plain` encuentra **cuatro** (los que D-1 conocía). Pero **FR-014 exige derivar el
alcance por barrido de toda la documentación**, y al barrer también por `opt-in`, `project_ref_mode` y
`en claro` aparecen tres más. Esto es exactamente lo que FR-014 anticipaba: **la lista de memoria se
queda corta**.

| # | Fichero:línea | Contiene «plain» | Qué dice |
|---|---|---|---|
| 1 | `specs/001-agente-inicial/contracts/boundary-event.md:35` | sí | «hash salado; ruta en claro solo con opt-in plain» |
| 2 | `specs/001-agente-inicial/contracts/boundary-event.md:86` | sí | «Solo `*_ref` hasheados (salvo opt-in `plain`)» |
| 3 | `specs/001-agente-inicial/data-model.md:29` | sí | «nunca la ruta en claro (salvo opt-in `plain`)» |
| 4 | `specs/001-agente-inicial/data-model.md:105` | sí | fila `project_ref_mode` en la tabla de config |
| 5 | **`specs/001-agente-inicial/spec.md:82`** | **NO** | «Hash pseudónimo, nunca el nombre en claro **(salvo opt-in explícito)**» |
| 6 | **`specs/003-enrolamiento/data-model.md:42`** | **NO** | lista `project_ref_mode` entre los campos que el enroll preserva |
| 7 | **`Roadmap.md:260-266`** | sí | la deuda fichada: «**Decidir**: borrar la opción, o documentarla explícitamente como no soportada» — **fuera del alcance versionado**, ver nota |

**El sitio 5 es el que justifica la regla de FR-014**: promete el opt-in **sin nombrarlo**, así que
un barrido por `plain` lo habría dejado vivo — una promesa de envío en claro superviviente en la
propia spec fundacional.

**El sitio 7 queda FUERA del alcance de FR-014/SC-008, y por un motivo verificable, no por criterio**:
`Roadmap.md` está **gitignoreado** (`.gitignore:17`) y **no está en el índice de git**
(`git ls-files --error-unmatch Roadmap.md` → *did not match any file(s) known to git*). No es
documentación del repositorio: nadie que clone el repositorio lo lee, porque no viaja con él.

Por tanto el barrido de SC-008 se define sobre **la documentación versionada** —lo que un lector del
repositorio puede leer—, y son **6 sitios**, no 7.

Eso no significa que la entrada de deuda se quede como está: dice literalmente «**Decidir**: borrar
la opción, o documentarla explícitamente como no soportada», y esta feature **es** esa decisión.
**Se cierra fuera de la feature, por Basilio, en el cierre de sesión** — donde vive el resto del
mantenimiento de ese documento.

**El README NO necesita cambio**: `README.md:13-14` dice que los identificadores sensibles «solo
cruzan como **hash salado**» — ya es la promesa correcta, sin opt-in. Verificado leyéndolo, no
suponiéndolo. *(SC-008 exige que el barrido lo incluya; incluirlo y encontrar cero es el resultado
esperado, no una omisión.)*

---

## P1 · El marcador de raíz

### La pregunta con dientes: los árboles paralelos

La spec promete en su edge case que «árboles de trabajo paralelos del mismo repositorio (cada uno con
su propia raíz) producen identidades distintas». Esa promesa **se incumple sola** si la detección
busca «un **directorio** `.git`», porque en un árbol paralelo `.git` **no es un directorio**.

Verificado empíricamente (repo desechable en scratchpad, `git worktree add`):

```
repo principal    → .git : directory
worktree paralelo → .git : regular file
                    contenido: "gitdir: /…/repo/.git/worktrees/paralelo"
```

Un submódulo presenta la misma forma (fichero `.git` con `gitdir:`). Luego la regla correcta es
**«existe una entrada `.git`, sea fichero o directorio»**, y basta con eso: no hace falta leer el
fichero ni seguir el `gitdir:`, porque **la identidad se deriva de la ruta de la raíz del árbol**, que
es el directorio que contiene la entrada — y ese directorio ya es distinto para cada worktree. Leer
el contenido daría la ruta del repositorio común, que es justo lo que **no** queremos: fusionaría los
árboles paralelos.

### Alternativas

| Opción | Cómo | Coste |
|---|---|---|
| **A · Observación directa del marcador** (recomendada) | Ascender desde el directorio de lanzamiento comprobando la existencia de una entrada `.git` (fichero **o** directorio) en cada nivel; el primer nivel que la tenga es la raíz | Ninguno relevante. Autocontenida, stdlib (`os.Lstat`), funciona sin git instalado, sin proceso hijo, sin PATH |
| B · Delegar en `git rev-parse --show-toplevel` | Ejecutar git como proceso hijo | **Dependencia de runtime que el binario único no controla** (Principio III): falla si git no está instalado o no está en PATH; coste de un `exec` por directorio; y en un worktree devuelve la raíz **del worktree**, correcto, pero al precio anterior. Además introduce salida de proceso externo en un agente que presume de autocontenido |
| C · Marcador ampliado (`.git`, `.hg`, `go.mod`, `package.json`…) | Lista de marcadores | Amplía el reconocimiento a proyectos sin VCS, pero **rompe FR-004** de forma sutil: un monorepo con `package.json` en cada paquete convertiría **cada paquete** en un proyecto distinto, refragmentando justo lo que la feature une |

### Decisión

**Opción A.** Ascenso por el árbol comprobando la entrada `.git` con `os.Lstat` (no `os.Stat`: `Lstat`
no sigue enlaces y aquí solo importa la existencia de la entrada). Se acepta **fichero o directorio**
— y esa disyunción es el requisito, no un detalle: sin ella, la promesa de los árboles paralelos
queda escrita y sin cumplir.

**Límite del ascenso: es un TECHO, no un salto.** El ascenso **se detiene** al llegar al techo y
**nunca lo sobrepasa**; ningún marcador se «descarta para seguir subiendo».

- **Techo para rutas bajo el directorio personal**: el propio directorio personal, **exclusivo** — se
  examinan todos sus descendientes, él no.
- **Techo para el resto**: la raíz del sistema de ficheros, **exclusiva** — se examinan todos sus
  descendientes, ella no.

Consecuencia, y es la semántica que importa: **si hay un marcador por debajo del techo, ese gana**,
sin excepción. Un `.git` en `~/dev/proy` cuenta con total normalidad aunque esté bajo el home; lo
único que no cuenta es un `.git` **en el home mismo** o **en la raíz misma**, porque el ascenso ni
siquiera llega a mirar ahí. Agotado el ascenso sin marcador, se cae al fallback (FR-005).

Formularlo como techo —y no como «este marcador no cuenta»— elimina la pregunta absurda de qué haría
el algoritmo *después* de descartar un marcador del home: no hay después, porque ahí se acaba el
recorrido.

**Rationale**: es la única opción que satisface los tres principios a la vez — stdlib (Principio III),
auditable de un vistazo (Principio II), y sin refragmentar (FR-004). Y es la única que funciona en la
máquina de un usuario que no tiene git instalado, caso real en equipos que usan otras herramientas.

---

## P2 · Normalización sintáctica exacta

### Las operaciones y su orden

Sobre la ruta declarada por el log, en este orden:

1. **Rechazo de lo no absoluto**: si la ruta declarada no es absoluta, **no se intenta anclarla**.
   `filepath.Abs` la resolvería contra el cwd **del proceso agente**, que no tiene ninguna relación
   con el cwd del log — produciría una identidad inventada. Una ruta relativa se trata como
   irresoluble y se normaliza solo léxicamente.
2. **`filepath.Clean`**: colapsa separadores repetidos, elimina `.`, resuelve `..` léxicamente y
   **quita la barra final** (salvo en la raíz). Cubre literalmente los dos casos que FR-006 nombra
   —barras finales y segmentos redundantes— y nada más. Es puramente léxico: no toca el disco.
3. *(en la rama observable)* **`filepath.EvalSymlinks`** — pero eso es **P1/FR-006a**, no
   normalización sintáctica; va en el orden global de resolución, abajo.

**Windows**: `filepath.Clean` ya convierte `/` en `\` al compilar para Windows, así que la
convergencia de separadores sale gratis. Las letras de unidad no se tocan (ver más abajo).

### El problema difícil: sensibilidad a mayúsculas

FR-007 prohíbe **las dos** direcciones del error: fusionar lo distinto **y** separar lo igual. Y la
sensibilidad a mayúsculas coloca esas dos prohibiciones en conflicto entre plataformas:

- En Linux (ext4/xfs), `/home/A` y `/home/a` son **directorios distintos**. Aplicar `strings.ToLower`
  los fusionaría → **viola FR-007**.
- En Windows y en macOS (APFS por defecto: insensible, preservador), `C:\Dev\Proyecto` y
  `c:\dev\proyecto` son **el mismo directorio**. No aplicar nada los separaría → **viola FR-007**
  también.

No hay una operación léxica que sea correcta en las dos. Alternativas:

| Opción | Comportamiento | Coste |
|---|---|---|
| **A · Sin case-folding; que lo resuelva la observación** (recomendada) | La rama observable (`EvalSymlinks`) devuelve la ruta tal como el sistema la resuelve, lo que en la práctica canonicaliza el caso en los FS que lo hacen; el fallback léxico **no** aplica folding | Residuo conocido: dos grafías distintas del mismo directorio **irresoluble** en un FS insensible producen dos identidades. Ocurre solo cuando el directorio ya no existe **y** el log trajo dos grafías — intersección estrecha |
| B · `ToLower` condicionado a `runtime.GOOS` | Folding en `windows` y `darwin`, no en `linux` | Falso en los dos sentidos: hay volúmenes **sensibles** en macOS (APFS case-sensitive, común en discos de desarrollo) y montajes insensibles en Linux (CIFS). Decide por SO lo que es propiedad **del volumen** |
| C · Preguntar al FS por su sensibilidad | Crear/probar una entrada para detectarlo | **Viola FR-012**: la observación es de solo lectura. Descartada de plano |

### Decisión

**Opción A**, con el residuo declarado. El criterio: entre un error sistemático (B decide por SO lo
que depende del volumen) y un residuo acotado que solo aparece en la intersección de dos condiciones
poco frecuentes, se elige el residuo — y se escribe, en vez de fingir que no existe.

**Nota sobre `~`**: si el log trajera una ruta con `~` sin expandir, `Clean` no la expande y el
resultado sería una identidad basada en literal `~/...`. No se expande **a propósito**: expandir
exigiría asumir que el `~` del log es el mismo usuario que ejecuta el agente, y esa suposición es
falsa en cuanto alguien procesa logs de otra cuenta. El caso no se ha observado en los datos de
referencia (los 12 valores son rutas absolutas).

---

## P3 · La caché de resolución

**Restricción heredada**: SC-006 se garantiza **por diseño** —una resolución por directorio distinto,
no por evento—, y el benchmark **nunca** es puerta de CI.

### Ámbito: por pasada, no por proceso

`generate()` (`cmd/permea/main.go:167-206`) recorre los logs y llama a `ingest.FromClaudeCodeLine` por
línea. Un fichero de log típico contiene cientos de eventos **del mismo `cwd`** — de ahí la necesidad
de la caché.

| Ámbito | A favor | En contra |
|---|---|---|
| **Por pasada** (recomendado) | Cada `--run` (y cada `tick()` del daemon) reevalúa el mundo. Es el corte natural: el estado del FS se relee con la frecuencia con la que el agente hace algo | Recalcula lo ya resuelto entre pasadas — coste despreciable: decenas de directorios distintos, no cientos de miles |
| Por proceso | Máximo ahorro | En `--daemon` el proceso vive días: un directorio que se convierte en repositorio **no se detectaría hasta reiniciar**. La spec acepta que la identidad cambie **en el tiempo**, pero no que quede congelada arbitrariamente por la vida de un proceso |

### Clave: la ruta **declarada**, no la real

La caché se consulta **antes** de resolver, así que la clave tiene que ser lo único que se tiene sin
trabajo: la ruta tal como viene del log. Usar la ruta real como clave obligaría a resolverla para
poder consultar — que es exactamente el trabajo que la caché existe para evitar.

Valor almacenado: **la identidad de proyecto ya derivada** (el `Ref` final), no la ruta intermedia.
Así el acierto de caché evita también el hash.

**Consecuencia deliberada**: dos grafías distintas del mismo directorio son dos claves y se resuelven
por separado — pero **convergen en el mismo valor**, que es lo que FR-006/FR-006a exigen. La caché
optimiza; no decide identidad.

### El edge case «el directorio cambia entre eventos de la misma pasada»

Dentro de una pasada, la primera resolución gana para todos los eventos de esa pasada con el mismo
`cwd` declarado. Es coherente con la spec —que acepta el cambio **en el tiempo**— y elige la
granularidad más defendible: **una pasada es un instante de observación**. Que dos eventos de la misma
pasada discrepen porque alguien inicializó un repositorio a mitad del escaneo sería un no-determinismo
peor que la instantánea.

---

## P4 · Dónde corta la parada de FR-013

### Quién carga la configuración, verificado

`grep -n "config.Load"` sobre `cmd/`:

| Punto de carga | Fichero:línea | ¿Procesa o emite eventos? |
|---|---|---|
| `setup()` | `cmd/permea/main.go:129` | **Sí** — es el constructor de `agent`, y de él cuelgan `runOnce()` (`:225`), `runDaemon()` (`:251`) y `tick()` (`:283`) |
| `runStatus()` | `cmd/permea/status.go:20` | No — solo informa |
| `enroll()` | `cmd/permea/enroll.go:68` | No — lee, muta y **reescribe** (`config.Save`, `:78`) |
| `dryRun()` | — | **No carga config**: usa un `ingest.Context` fijo (`main.go:316`). Sí procesa líneas, pero con salt de dry-run y sin encolar ni transmitir |

**Matiz sobre D-1**: es cierto que el enroll **no invoca `setup()`**, pero **sí llama a
`config.Load`** (`enroll.go:68`). La distinción importa para decidir dónde cortar.

### Las alternativas

| Opción | Dónde corta | Consecuencia |
|---|---|---|
| A · En `config.Load` | Punto único; paran los cuatro caminos | Garantía trivial y máxima… y **deja al usuario sin salida**: `status` no puede diagnosticar y `enroll` —que reescribiría la config y borraría la clave obsoleta— tampoco puede ejecutarse. El producto se cierra sobre sí mismo |
| **B · En `setup()`, con status informando y enroll operativo** (recomendada) | `setup()` para → `--run`, `--daemon`, `tick` no arrancan. `status` **informa del problema** sin parar. `enroll` sigue funcionando | Cumple la garantía de la spec —«cero eventos procesados o emitidos»— porque los tres únicos caminos que procesan o emiten cuelgan de `setup()`. Y deja **dos vías de diagnóstico y reparación** abiertas |
| C · En cada subcomando por separado | Comprobación replicada | Es el defecto de clase de una condición replicada: alguien añade un subcomando y olvida la comprobación. Descartada |

### Decisión

**Opción B**, con la detección **escrita una sola vez** en el paquete `config` (una función que
inspecciona la clave obsoleta) y **consumida** en `setup()` como error de arranque. `status` llama a
la misma función para **informar**, no para abortar.

**Rationale**: FR-013 exige que el agente «NUNCA procesa ni emite eventos con esa configuración
presente», y los tres caminos que procesan o emiten —`runOnce`, `runDaemon`, `tick`— pasan **todos**
por `setup()`. Cortar ahí satisface la garantía completa. Parar además el diagnóstico y la reparación
no añade ni una pizca de seguridad y empeora el producto: un usuario bloqueado sin poder ver por qué
ni arreglarlo.

*(Esta es la única alternativa de P4 donde la recomendación se aparta de la lectura más literal de
«detener el arranque». Se señala explícitamente para que el orquestador la confirme o la corrija: la
letra de SC-007 dice «detiene el arranque», y aquí se propone que `status` y `enroll` sean excepciones
razonadas.)*

### La trampa de config, y cómo se construye la garantía

Tras retirar el campo del struct, `encoding/json` **descarta las claves desconocidas en silencio** —
la detección no sale gratis del borrado. Opciones:

| Opción | Coste |
|---|---|
| **A · Segundo unmarshal a `map[string]json.RawMessage`** (recomendada) | Una pasada extra sobre un fichero de pocos cientos de bytes. Detección **acotada a la clave obsoleta**: se mira `project_ref_mode` y nada más |
| B · Mantener el campo como `string` marcado obsoleto | Contradice FR-015 (el ajuste debe desaparecer de la superficie con significado) y deja el campo vivo en `Save`, reescribiéndolo |
| C · `Decoder.DisallowUnknownFields()` | Rechaza **toda** clave desconocida: rompería cualquier config con campos futuros o comentarios, y convertiría un cambio acotado en una ruptura general |

**Decisión: A.** La comprobación mira exactamente una clave y compara exactamente un valor
(`"plain"`), que es lo que FR-013b exige: solo esa solicitud para; cualquier otro valor de esa clave
se ignora en silencio.

---

## P5 · El mensaje de parada

Forma final: **causa + corrección + código de salida ≠ 0**, en `stderr`, coherente con el estilo
existente (`fmt.Fprintln(os.Stderr, "error:", err)`, `main.go:46`).

```
$ permea --run
Permea 0.1.0
error: el modo `project_ref_mode: "plain"` fue retirado y ya no existe.
       La identidad de proyecto cruza siempre de forma irreversible.
       Elimina la clave "project_ref_mode" de <ruta>/config.json para continuar.
$ echo $?
1
```

Decisiones de forma:

- **Nombra la clave y el valor exactos** — para que el usuario no tenga que adivinar cuál de sus
  ajustes es.
- **Da la ruta real del fichero**, resuelta por SO (`config.DataDir()`), no «tu config.json».
- **Dice qué hace el producto ahora** («cruza siempre de forma irreversible»), porque el usuario que
  puso `plain` creía tener otra cosa: el mensaje corrige la creencia, no solo el fichero.
- **Exit 1**, el mismo que el resto de errores de arranque — sin código especial.
- **NUNCA** vuelca la configuración completa ni ningún secreto (disciplina de P-003 FR-007).

Con la opción elegida en D-004-1, un `project_ref_mode: "hash"` residual **no** produce nada: ni
error, ni aviso, ni línea en stderr.

---

## P6 · Los tests de la frontera

### SC-005 — extender el golden test

El golden existente (`internal/ingest/boundary_test.go`) ya tiene la forma correcta: una `denylist`
de términos (`:16-33`) y un fixture con contenido inyectado. Extensión necesaria:

1. **Añadir a la denylist los fragmentos de ruta reconocibles** que la nueva resolución maneja: la
   raíz de proyecto resuelta, el subdirectorio de lanzamiento y el destino de un enlace. Hoy la
   denylist ya incluye `/home/basilio`, `acme-banca`, `core-pagos` — la extensión es de la misma
   familia, con rutas que ejerciten **raíz** y **subdirectorio** por separado.
2. **Ampliar el alcance de la comprobación de evento serializado a cola y transporte**, que es lo que
   SC-005 pide tras D-004-4. La cola es `queue.jsonl` (`transport.Append`); el transporte se ejercita
   con `httptest.NewTLSServer` capturando el cuerpo — patrón ya usado en el repo
   (`specs/001-agente-inicial/contracts/transport.md:58`).

**El caso que el test debe cazar y hoy no existiría**: que la ruta de la **raíz resuelta** —un valor
nuevo que antes no circulaba por el proceso— se filtre por algún camino. Es el riesgo específico que
introduce esta feature.

### SC-007 — arranque detenido / arranque limpio

Test de proceso con `os/exec`, comprobando `ExitCode()` — **no** comparando texto de error, por el
puente Windows/WSL ya fichado en el repo. Dos casos:

| Caso | Config | Esperado |
|---|---|---|
| Parada | `{"project_ref_mode":"plain"}` | `ExitCode() != 0`, cero eventos en `queue.jsonl` |
| Limpio | `{"project_ref_mode":"hash"}` | `ExitCode() == 0`, **sin** salida de error ni aviso |
| Limpio | sin la clave | `ExitCode() == 0` |

El tercer caso parece redundante y no lo es: distingue «ignoro el valor residual» de «ignoro la clave
entera», que son los dos comportamientos de FR-013a y FR-013b.

### SC-004 — la línea base, y por qué es la primera tarea

**Restricción de orden heredada**: las identidades de sesión y máquina del conjunto de referencia se
capturan a **artefacto versionado antes de tocar código**. Sin eso, «regresión cero» no tiene contra
qué medirse — se estaría comparando el resultado consigo mismo.

Forma: fichero de referencia con las identidades producidas por el fixture actual, generado con el
binario de `2ba8363` y commiteado **antes** del primer cambio de comportamiento.

---

## P7 · La verdad-terreno de SC-001

**Bloquea la validación, no el plan** (spec, Assumptions).

**Formato propuesto**: fichero versionado en la feature —`specs/004-identidad-de-proyecto/verdad-terreno.md`—
con una fila por cada uno de los 12 valores medidos:

| identidad actual (hex-64, prefijo) | directorio de origen | naturaleza | identidad esperada tras el cambio |
|---|---|---|---|
| `daad022c…` | *(a rellenar por Basilio)* | raíz de proyecto / subdirectorio de X / clon / worktree de X / directorio suelto | *(derivada de la regla)* |

Las dos primeras columnas las aporta Basilio; la tercera es su clasificación; **la cuarta la deriva la
regla**, y es lo que SC-001 compara. Que la predicción se escriba **antes** de ejecutar es lo que hace
del criterio una prueba y no una racionalización a posteriori.

---

## Orden global de resolución (D-004-2, no negociable)

Consolidado de P1+P2+P3, para que el plan lo implemente en este orden y no en otro:

```
cwd declarado por el log
  │
  ├─ ¿vacío? ──────────────────────────► identidad AUSENTE (FR-008)
  │
  ├─ 1. resolución de enlaces (best-effort, FR-006a/FR-009)
  │      EvalSymlinks; si falla → seguir con la forma léxica
  │
  ├─ 2. reconocimiento de raíz (FR-001/FR-004/FR-004a)
  │      ascenso buscando entrada `.git` (fichero o directorio)
  │      con TECHO EXCLUSIVO: el home si la ruta cuelga de él, la raíz del FS si no
  │      ¿encontrada bajo el techo? ────► Ref(salt, raíz)   [la más cercana]
  │
  └─ 3. fallback normalizado (FR-005/FR-006)
         filepath.Clean del mejor valor disponible ──► Ref(salt, normalizado)
```

Los tres pasos se consultan **a través de la caché de P3**, cuya clave es el `cwd` declarado.

---

## Lo que este research NO abre

- Nada de la plataforma (repo aparte).
- Ningún campo nuevo del evento (FR-020); ningún cambio en `Ref`, salt, sesión o máquina (FR-019).
- El saneado de `stderr` (D-004-4, trabajo aparte ya identificado). **Nota**: `dryRun` imprime hoy
  `project_ref` truncado a 8 caracteres (`main.go:329-334`) — es un hash, no una ruta, así que no
  entra en conflicto con FR-017 ni con el trabajo aparte.
