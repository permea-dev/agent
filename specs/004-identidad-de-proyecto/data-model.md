# Data Model · P-004 Identidad de proyecto

**Fecha**: 2026-08-09 · **Spec**: [spec.md](./spec.md) · **Research**: [research.md](./research.md)

Esta feature **no añade ninguna entidad persistida ni ningún campo al evento** (FR-020). Lo que
describe este documento son las entidades **de proceso** —las que viven durante una pasada— y las
reglas de derivación que las relacionan.

---

## Lo que NO cambia (y por qué se escribe)

| Elemento | Estado | Cita |
|---|---|---|
| `event.Event` — struct cerrado de 17 campos | **Intacto.** Ni un campo nuevo, ni un tipo cambiado | `internal/event/event.go:18-35` |
| `event.Ref(salt, value)` | **Intacta.** Misma función, mismo algoritmo, mismo salt, mismo ámbito | `internal/event/event.go:40-46` |
| `Event.SessionRef` / `Event.MachineRef` | **Intactos.** Misma fuente (`r.SessionID`, `ctx.MachineID`), mismo valor byte a byte | `internal/ingest/claudecode.go:79-80` |
| El salt y su ciclo de vida | **Intacto.** Ni generación, ni rotación, ni ámbito | `config.LoadOrCreateSalt` |
| `rawRecord` | **Intacto.** No se decodifica ningún campo nuevo del log | `internal/ingest/claudecode.go:24-38` |
| Formato de `queue.jsonl`, `state.json` | **Intactos** | — |

**Lo único que cambia en el evento es el VALOR de `ProjectRef`**, no su presencia, ni su tipo, ni su
forma (FR-018: sigue siendo hex-64 o ausente).

---

## Entidades de proceso

### 1 · Directorio de lanzamiento declarado

**Qué es**: la cadena que el log de la herramienta trae como directorio de trabajo. Es la **entrada**
de toda la derivación.

| Atributo | Valor |
|---|---|
| Origen | `rawRecord.Cwd` (`json:"cwd"`) — `internal/ingest/claudecode.go:28` |
| Tipo | cadena; puede estar **vacía** |
| Confianza | **ninguna**: no se asume que exista, que sea absoluta, ni que esté normalizada |
| Vacía → | identidad **ausente** (FR-008); `event.Ref` ya devuelve `""` para entrada vacía (`event.go:41-43`) |

### 2 · Ubicación real

**Qué es**: el directorio de lanzamiento tras resolver enlaces simbólicos, **cuando el sistema lo
permite**.

| Atributo | Valor |
|---|---|
| Derivación | `filepath.EvalSymlinks(declarado)` — mejor esfuerzo |
| Si falla | **no es error**: se continúa con la forma léxica del declarado (FR-009) |
| Precedencia | Se calcula **antes** del reconocimiento de raíz (FR-006a / D-004-2) |

**Por qué precede**: un enlace que apunta al interior de un proyecto debe recibir la identidad de ese
proyecto. Si el reconocimiento fuera primero, el enlace no llegaría nunca a reconocerse.

### 3 · Raíz de proyecto

**Qué es**: el directorio más cercano, ascendiendo desde la ubicación real, que contiene una entrada
`.git` — **fichero o directorio** — y que no está excluido por FR-004a.

| Atributo | Valor |
|---|---|
| Marcador | entrada `.git`, comprobada con `os.Lstat` |
| **Fichero o directorio** | **Los dos cuentan.** Directorio = repositorio principal; fichero = **árbol de trabajo paralelo** o submódulo. Verificado empíricamente (research §P1) |
| Dirección | Ascendente desde la ubicación real |
| **Techo (exclusivo)** | El **directorio personal** si la ruta cuelga de él; la **raíz del FS** en caso contrario. El ascenso **se detiene ahí y no lo sobrepasa**: el techo mismo **no se examina** (FR-004a) |
| Resultado | El **primer** marcador encontrado bajo el techo gana → proyecto más cercano (FR-004) |
| Si no hay | Se cae a la entidad 4 |
| Origen del home | `os.UserHomeDir()` — **y esa elección es deliberada**: hace FR-004a testeable por entorno (ver abajo) |
| Forma del techo | **Parámetro interno inyectable.** En producción se resuelve solo (el home por `os.UserHomeDir()`, la raíz del FS por la estructura de la ruta); en test se **inyecta**, porque el techo de la **raíz del sistema** no es falseable por entorno sin privilegios y solo así puede ejercitarse el caso 8 (T012) |

**Techo, no salto.** Ningún marcador se descarta «para seguir subiendo»: si existe un `.git` por
debajo del techo, ese gana sin excepción. Un `.git` en `~/dev/proy` cuenta con total normalidad; lo
único que no cuenta es un `.git` **en el home mismo** o **en la raíz misma**, porque el recorrido
termina antes de mirar ahí.

**Por qué el home se lee vía `os.UserHomeDir()`**: porque respeta `HOME` (Linux/macOS) y
`USERPROFILE` (Windows), y eso es lo que permite **probar FR-004a por entorno** —un HOME temporal que
contiene un `.git`— sin tocar el directorio personal real de nadie. Una constante compilada, o una
detección por prefijo de ruta, dejaría el requisito escrito y sin forma razonable de verificarlo. La
testabilidad es aquí una restricción de diseño, no una consecuencia.

**El contenido del fichero `.git` NO se lee.** La identidad se deriva de la **ruta del directorio que
contiene el marcador**, que ya es distinta por worktree. Leer el `gitdir:` daría la ruta del
repositorio común y **fusionaría los árboles paralelos** — lo contrario de lo que la spec promete.

### 4 · Directorio normalizado (fallback)

**Qué es**: el mejor valor disponible, pasado por normalización puramente léxica.

| Atributo | Valor |
|---|---|
| Derivación | `filepath.Clean(mejor_valor)` |
| «Mejor valor» | la ubicación real si se resolvió; si no, el declarado |
| Qué hace `Clean` | colapsa separadores repetidos, elimina `.`, resuelve `..` léxicamente, **quita la barra final** |
| Qué **NO** hace | no expande `~`; no aplica `filepath.Abs`; no aplica case-folding (research §P2) |

**Por qué no `Abs`**: anclaría una ruta relativa al cwd **del proceso agente**, que no tiene relación
con el del log — produciría una identidad inventada.

### 5 · Identidad de proyecto

**Qué es**: el valor que viaja en el evento.

| Atributo | Valor |
|---|---|
| Derivación | `event.Ref(salt, X)` donde `X` es la raíz (3) o el normalizado (4) |
| Forma | hex-64, o **cadena vacía** (ausente) — sin cambio respecto de hoy (FR-018) |
| Opacidad | El receptor **no puede** distinguir por la forma si viene de raíz o de fallback |
| Reversibilidad | Ninguna: el salt no se transmite |

### 6 · Caché de resolución (por pasada)

| Atributo | Valor |
|---|---|
| Ámbito | **Una pasada** (`generate()` / `tick()`), en memoria. **No** por proceso (research §P3) |
| Clave | El **directorio declarado**, tal cual viene del log |
| Valor | La **identidad ya derivada** (evita también el hash en el acierto) |
| Propósito | SC-006 **por diseño**: una resolución por directorio distinto, no por evento |
| Efecto en la identidad | **Ninguno.** La caché optimiza; no decide |

---

## Reglas de derivación — la máquina completa

```
ENTRADA: cwd declarado (cadena), salt

  cwd == ""  ──────────────────────────────────► identidad = ""        [FR-008]

  ¿en caché por cwd?  ─── sí ──────────────────► identidad cacheada    [SC-006]
       │ no
       ▼
  1. real := EvalSymlinks(cwd)                                          [FR-006a]
       └─ error → real := cwd            (best-effort, no aborta)       [FR-009]

  2. techo := home si `real` cuelga del home, si no la raíz del FS      [FR-004a]
     raiz  := ascender desde `real` hasta el techo (EXCLUSIVO),          [FR-001]
              buscando entrada `.git` (fichero o directorio)
       ├─ primer marcador encontrado → identidad = Ref(salt, raiz)      [FR-004]
       │                                └─ el MÁS CERCANO; nada se salta
       └─ techo alcanzado sin marcador → paso 3

  3. identidad = Ref(salt, Clean(real))                                 [FR-005/FR-006]

  guardar en caché bajo la clave `cwd`
SALIDA: identidad
```

**Ninguna rama devuelve error.** El peor caso —todo falla— produce
`Ref(salt, Clean(cwd_declarado))`, que es exactamente el comportamiento actual salvo por el `Clean`.
Es la garantía de FR-010: la emisión nunca se interrumpe.

---

## Configuración: lo que se retira y lo que se construye

### Se retira

| Elemento | Fichero:línea | Acción |
|---|---|---|
| `type ProjectRefMode` | `internal/config/config.go:15` | borrar |
| `ModeHash` / `ModePlain` | `:19`, `:21` | borrar |
| Campo `ProjectRefMode` de `Config` | `:30` | borrar |
| Default en `Default()` | `:41` | borrar |
| Relleno en `applyDefaults()` | `:120-121` | borrar |
| Aserciones de test | `config_test.go:36`, `:52` | **obligatorio**: dejan de compilar |

### Se construye — la detección no sale gratis del borrado

Tras retirar el campo, `encoding/json` **descarta la clave desconocida en silencio**. Para que FR-013
tenga efecto hace falta una lectura deliberada:

| Elemento | Comportamiento |
|---|---|
| Detector de clave obsoleta | Segundo unmarshal a `map[string]json.RawMessage`, mirando **solo** `project_ref_mode` |
| Valor `"plain"` | **Parada** con error visible (FR-013) |
| Valor `"hash"` | **Silencio total**: ni error, ni aviso (FR-013a) |
| Cualquier otro valor | **Silencio total** (FR-013b) |
| Clave ausente | **Silencio total** |
| Escritura (`config.Save`) | La clave **no se reescribe**: al no estar en el struct, desaparece del fichero en el siguiente `Save` |

**Consecuencia útil y deliberada**: como `enroll` hace `Load` + `Save` (`enroll.go:68,78`), un
enrolamiento posterior **limpia la clave obsoleta** del fichero sin que nadie la borre a mano. Es la
razón operativa por la que P4 recomienda que `enroll` **no** se detenga.

---

## Trazabilidad requisito → entidad

| Requisito | Dónde vive |
|---|---|
| FR-001, FR-002, FR-003 | Entidad 3 (raíz) + regla 2 |
| FR-004 | Regla 2: el **primer** marcador ascendiendo |
| FR-004a | Regla 2: exclusión de home y raíz del FS |
| FR-005, FR-006, FR-007 | Entidad 4 (normalizado) + regla 3 |
| FR-006a | Entidad 2 + **orden** de las reglas 1→2 |
| FR-008 | Guarda de entrada vacía |
| FR-009, FR-010 | Todas las ramas de error → continuación, nunca aborto |
| FR-011, FR-012 | Entidades 2 y 3: `EvalSymlinks` + `Lstat`, solo lectura, nunca contenido |
| FR-013, FR-013a, FR-013b | §Configuración → detector |
| FR-014, FR-015 | §Configuración → se retira (**6 sitios documentales en alcance**; el 7.º, `Roadmap.md`, está gitignoreado — research §alcance documental) |
| FR-016, FR-017, FR-018 | Entidad 5: `Ref` intacta, forma intacta |
| FR-019, FR-020 | §Lo que NO cambia |
