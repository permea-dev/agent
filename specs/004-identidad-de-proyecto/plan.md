# Implementation Plan: Identidad de proyecto

**Branch**: `004-identidad-de-proyecto` | **Date**: 2026-08-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-identidad-de-proyecto/spec.md` (congelada en
`2ba8363`, salvo la ampliación aprobada de D-004-3)

## Summary

La identidad de proyecto deja de derivarse del **directorio de lanzamiento tal cual lo declara el
log** y pasa a derivarse de **la raíz del árbol de trabajo que lo contiene**, con fallback al propio
directorio normalizado cuando ningún proyecto lo contiene. Para lograrlo, el agente **observa el
sistema de ficheros por primera vez** —hoy solo transcribe— en modo estrictamente de solo lectura y
de mejor esfuerzo: ningún fallo de observación interrumpe jamás la emisión de un evento.

El cambio se concentra en **un único punto de composición** (`internal/ingest/claudecode.go:78`), al
que se le cambia **qué valor entra**, nunca la función que lo hace irreversible: `event.Ref` y el
salt no se tocan, y las identidades de sesión y máquina quedan byte a byte iguales (FR-019).

En paralelo se retira del producto la promesa de envío en claro: el ajuste `project_ref_mode`
desaparece de la superficie con significado, y una configuración que **solicita** el modo retirado
detiene el arranque de los caminos que procesan o emiten eventos — mientras que un valor residual que
pedía el comportamiento que ya es el único se ignora en silencio (D-004-1).

## Technical Context

**Language/Version**: Go 1.x (módulo `github.com/permea-dev/agent`; binario único vía `cmd/permea`)

**Primary Dependencies**: **ninguna nueva — stdlib only** (restricción heredada). La feature usa
`path/filepath` (`Clean`, `EvalSymlinks`, `Dir`, `IsAbs`) y `os` (`Lstat`), ya en uso en el repo

**Storage**: ficheros locales bajo el directorio de datos por SO (`config.json`, `state.json`,
`queue.jsonl`, `salt`). **Esta feature no añade ninguno** y no cambia el formato de los existentes,
salvo la retirada de una clave de `config.json`

**Testing**: `go test ./...` (stdlib `testing`); golden test de frontera en
`internal/ingest/boundary_test.go`; tests de proceso con `os/exec` + `ExitCode()`

**Target Platform**: macOS, Linux y Windows (binario estático único). La resolución de rutas es por
SO y **nunca** se hardcodea

**Project Type**: CLI / agente local (proyecto único Go, sin frontend ni servicio)

**Performance Goals**: **una resolución por directorio distinto, no por evento** (SC-006). Un log
típico trae cientos de eventos del mismo `cwd`; la caché de pasada es la garantía **por diseño**

**Constraints**: la emisión **nunca** se interrumpe por un fallo de resolución (FR-009/FR-010);
observación **de solo lectura** y acotada (FR-012); cero campos nuevos en el evento (FR-020); cero
cambios en `Ref`/salt/sesión/máquina (FR-019)

**Scale/Scope**: decenas de directorios distintos por pasada; miles de eventos por fichero de log.
La escala real medida: 655 eventos, 12 identidades, 1 desarrollador, 1 máquina

## Constitution Check

*GATE: evaluado contra `.specify/memory/constitution.md` v1.0.0 antes de Phase 0 y re-evaluado tras
Phase 1.*

| Principio | Exigencia | Cumplimiento de este plan | Puerta |
|---|---|---|---|
| **I · Frontera de datos inviolable** (NO NEGOCIABLE) | Allowlist cerrada; los identificadores sensibles cruzan solo como hash salado; el salt nunca se transmite | **No se añade ningún campo** al struct cerrado (FR-020). La identidad sigue siendo `Ref(salt, …)`; lo único que cambia es el **valor de entrada**. La retirada del `plain` **estrecha** la frontera: elimina la única promesa documentada de envío en claro | ✅ PASA |
| **II · Privacidad auditable** (local-first) | Verificable leyendo el código; procesamiento local; sin dependencia del backend | La resolución es **enteramente local** y no consulta nada externo. La ampliación de comportamiento —observar el FS— se declara en la spec (FR-011) y se acota por escrito (FR-012: solo lectura, nunca contenido de ficheros de trabajo) en vez de colarse en el plan | ✅ PASA |
| **III · Binario único y auditable** | Stdlib favorecida; toda dependencia justificada; rutas por SO; legibilidad sobre astucia | **Cero dependencias nuevas.** P1 elige `os.Lstat` sobre delegar en `git` precisamente para no introducir una dependencia de runtime que el binario no controla. La resolución de rutas usa `path/filepath`, que es por SO por construcción | ✅ PASA |
| **IV · Test-first en la frontera** (NO NEGOCIABLE) | El golden test es disciplina de primer commit; Given/When/Then; DEBE/NUNCA | El golden existente se **extiende antes** de cambiar la derivación: rutas de raíz y de subdirectorio en la denylist, y alcance ampliado a cola y transporte (SC-005). La línea base de SC-004 se captura **antes de tocar código** | ✅ PASA |
| **V · Desarrollo dirigido por especificaciones** | spec → plan → tasks → implement; la spec no lleva implementación | La spec quedó congelada y este plan solo resuelve mecanismo. Las siete preguntas están respondidas en `research.md` **con alternativas y coste** donde las había | ✅ PASA |

**Resultado de la puerta**: pasa sin violaciones. La sección *Complexity Tracking* queda vacía a
propósito.

**Una desviación señalada, no oculta**: la recomendación de P4 propone que `status` y `enroll` **no**
se detengan ante la configuración retirada, aunque la letra de SC-007 dice «detiene el arranque». No
es una violación de la constitución —la garantía de FR-013 («cero eventos procesados o emitidos») se
cumple entera—, pero es una lectura que el orquestador debe confirmar. Está razonada en `research.md`
§P4 y repetida abajo en *Puntos que requieren confirmación*.

## Project Structure

### Documentation (this feature)

```text
specs/004-identidad-de-proyecto/
├── plan.md              # Este fichero
├── research.md          # Phase 0 — P1..P7 resueltas con alternativas y coste
├── data-model.md        # Phase 1 — entidades y reglas de derivación
├── quickstart.md        # Phase 1 — guía de validación ejecutable
├── contracts/
│   ├── project-identity.md   # Contrato de derivación (entrada → identidad)
│   └── cli-config.md         # Contrato de configuración y parada
├── checklists/
│   └── requirements.md  # Calidad de la spec (re-validado E-1/E-1b)
└── tasks.md             # Phase 2 — NO lo crea /speckit-plan
```

### Source Code (repository root)

```text
cmd/permea/
├── main.go              # setup() ← punto de corte de la parada (P4); generate() ← pasada
├── status.go            # informa del problema sin abortar (P4)
└── enroll.go            # no se detiene: es la vía de reparación (P4)

internal/
├── config/
│   ├── config.go        # RETIRAR ProjectRefMode; AÑADIR detección de la clave obsoleta
│   └── config_test.go   # DEJA DE COMPILAR al retirar el campo (:36, :52) — trabajo obligado
├── event/
│   └── event.go         # Ref() — NO SE TOCA (FR-019)
├── ingest/
│   ├── claudecode.go    # :78 — único punto de composición; cambia el VALOR de entrada
│   ├── boundary_test.go # golden extendido (SC-005)
│   └── testdata/        # fixture ampliado con rutas de raíz y subdirectorio
├── project/             # NUEVO — resolución de identidad de proyecto (P1+P2+P3)
│   ├── resolve.go       # orden: enlaces → raíz → fallback normalizado
│   └── resolve_test.go
└── testutil/            # NUEVO — helper de aislamiento de tests
    └── sandbox.go       # HOME/USERPROFILE/XDG_CONFIG_HOME → t.TempDir()
```

**Sobre `internal/testutil/`**: existe porque **tres familias de tests** —los de config, los de
proceso y las validaciones— necesitan el mismo aislamiento, y replicarlo en cada una es la vía
segura de que a alguna se le olvide y escriba en la instalación real del desarrollador. El helper
incluye su propia aserción: un `config.DataDir()` que caiga fuera del temporal es **fallo inmediato
del test**, no aviso.

**Structure Decision**: proyecto único Go, con la estructura de referencia que fija la constitución
(§Restricciones técnicas). **La resolución vive en un paquete nuevo `internal/project/`** y no dentro
de `internal/ingest/` por dos razones: `ingest` es *«el lector por herramienta»* y esta lógica no es
de ninguna herramienta —una segunda herramienta que declare su directorio de trabajo la heredará sin
tocarla—; y porque `internal/event/` es *«la frontera, pequeña y obvia»* y meterle observación de
sistema de ficheros la haría menos auditable, que es justo lo contrario de su función.

`internal/ingest/claudecode.go:78` sigue siendo el único punto de composición: pasa de recibir
`r.Cwd` a recibir el valor que devuelve el resolutor. La frontera no se mueve.

## Fases

### Phase 0 — Research ✅

`research.md` responde P1..P7 y reverifica contra HEAD lo que D-1 dejó medido. **Dos correcciones
materiales** salieron de esa reverificación (ver *Hallazgos*).

### Phase 1 — Diseño y contratos ✅

`data-model.md`, `contracts/project-identity.md`, `contracts/cli-config.md`, `quickstart.md`.

**Nota sobre el paso «update agent context»**: el repo **no tiene** `update-agent-context.sh` en
`.specify/scripts/bash/` (solo `check-prerequisites`, `common`, `create-new-feature`, `setup-plan`,
`setup-tasks`). El paso se omite por inexistente, no por descuido.

### Phase 2 — Tasks

La genera `/speckit-tasks`. **Restricciones de orden que tasks.md debe respetar**:

1. **PRIMERA tarea, antes de tocar código**: capturar la línea base de SC-004 (identidades de sesión
   y máquina del conjunto de referencia) a artefacto versionado. Sin ella, «regresión cero» no tiene
   contra qué medirse.
2. **El golden test de frontera se extiende antes** de cambiar la derivación (Principio IV).
3. La retirada del `plain` toca `config_test.go:36,52`, que **dejan de compilar** — va en el mismo
   commit que la retirada, no después.
4. El barrido documental de FR-014 usa **cuatro términos** (`plain`, `opt-in`, `project_ref_mode`,
   `en claro`), no uno: con solo `plain` se escapan 3 de los 6 sitios en alcance. Y su alcance sale
   de `git ls-files`, no de una lista a mano.

## Hallazgos de la reverificación contra HEAD

Dos cosas que D-1 dejó medidas **no coinciden** con `2ba8363`, y se corrigen aquí:

1. **El choke point no está donde decía el contexto.** `internal/source/` **no existe**; el fichero es
   `internal/ingest/claudecode.go` (la línea `:78` sí es correcta). El encargo pedía reverificar en
   vez de heredar: esta es la razón.
2. **El `plain` aparece en 7 sitios, no 4** — y **tres no contienen la palabra**. El más importante
   es `specs/001-agente-inicial/spec.md:82`, que promete «salvo opt-in explícito» sin nombrarlo: un
   barrido por `plain` lo habría dejado vivo. De los 7, **6 están en alcance**: el séptimo
   (`Roadmap.md`) está **gitignoreado** y fuera de la documentación del repositorio. Detalle en
   `research.md`.

Y un añadido al inventario: `internal/config/config_test.go:36,52` consumen `ProjectRefMode`, así que
**el paquete `config` no compila** tras la retirada hasta tocarlos. D-1 no los citó.

## Decisiones de plan

Decisiones de **mecanismo** tomadas en esta fase. Las de producto (D-004-1..D-004-4) viven en la
spec y no se repiten aquí.

### D-004-5 (2026-08-09) — «arranque», en FR-013/SC-007, son los caminos que procesan hacia emisión

**Interpretación confirmada.** La garantía de FR-013 —«cero eventos procesados o emitidos»— protege
**la frontera**, no el binario. Por tanto «detener el arranque» se aplica a los caminos que llevan
hacia una emisión, y esos pasan **todos** por `setup()` (`cmd/permea/main.go:129`): `--run`,
`--daemon` y el `tick()` del daemon.

**El criterio, que es lo que hay que conservar**: parar de más no añade ni una pizca de protección a
la frontera y sí retira al usuario las dos herramientas con las que saldría del problema. Un producto
que detecta una configuración obsoleta y a la vez impide diagnosticarla y repararla convierte un aviso
en un encierro.

**Tres excepciones razonadas**, cada una por un motivo distinto y verificable:

| Camino | Por qué no se detiene | Verificación |
|---|---|---|
| `permea status` | Es **diagnóstico**: su función es explicar el estado, y este problema es parte del estado. **Informa solo ante `"plain"`** —nunca ante `"hash"` ni ante otros valores—, así que no contradice FR-013a/FR-013b | `cmd/permea/status.go:20` carga config y no emite nada |
| `permea enroll` | Es la **vía de reparación**: hace `Load` + `Save`, y como la clave deja de estar en el struct, **el propio enroll la limpia del fichero** sin que nadie la borre a mano | `cmd/permea/enroll.go:68` (Load) y `:78` (Save) |
| `permea --scan` | Es **procesamiento diagnóstico**, no emisión: usa un `ingest.Context` con **salt fijo** (`"dry-run-salt"`), **no encola** y **no transporta** | `cmd/permea/main.go:316` |

`--scan` merece una línea aparte porque sí **procesa** líneas, y la garantía de FR-013 nombra
«procesados **o** emitidos». No se detiene porque su procesamiento no puede alcanzar la frontera por
tres barreras simultáneas: el salt es de dry-run (las identidades que produce no son las del usuario),
no escribe en `queue.jsonl`, y no abre ninguna conexión. Además **no lee `config.json` en absoluto**,
así que no hay ningún `"plain"` que pudiera detectar.

**Anclaje del test**: el test de SC-007 se escribe contra esta decisión —qué invocaciones paran y
cuáles no— y no contra una lectura literal de «arranque». La tabla completa está en
[contracts/cli-config.md](./contracts/cli-config.md) §«Alcance de la parada».

## Puntos que requieren confirmación del orquestador

Ninguno bloquea el diseño; los tres están resueltos con recomendación y se señalan por transparencia.

| # | Punto | Estado | Nota |
|---|---|---|---|
| 1 | **Alcance de la parada** (P4) | ✅ **RESUELTO** — confirmado como **D-004-5** | Ya no requiere confirmación: la interpretación y su criterio están registrados arriba |
| 2 | **Residuo de mayúsculas** (P2) | ✅ **RESUELTO** — confirmado por el orquestador el 2026-08-09 | **Opción A de research §P2**: sin case-folding; lo canonicaliza la observación real; **residuo declarado** — dos grafías de un directorio **irresoluble** en un FS insensible dan dos identidades. La alternativa (folding por `runtime.GOOS`) se descarta porque decide por sistema operativo lo que es propiedad **del volumen** |
| 3 | **`Roadmap.md:260-266`** | ✅ **RESUELTO** — fuera de alcance | `Roadmap.md` está **gitignoreado** (`.gitignore:17`) y no está en el índice: no es documentación del repositorio, así que queda fuera de FR-014/SC-008. La entrada de deuda la cierra Basilio en el cierre de sesión, fuera de la feature |

## Complexity Tracking

> Sin violaciones de la Constitution Check. Sección vacía a propósito.
