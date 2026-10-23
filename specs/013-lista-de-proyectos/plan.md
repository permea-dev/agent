# Implementation Plan: 013 · «Lista de proyectos»

**Branch**: `013-lista-de-proyectos` | **Date**: 2026-10-23 | **Spec**: [spec.md](./spec.md) *(ratificada el 2026-10-23)* | **Base**: el commit de la spec
ratificada *(el código, el de `cabf8c5`, 0.9.0)*
**Input**: la spec · [`soporte/descubrimiento.md`](./soporte/descubrimiento.md) · el plan y las tareas de 012 *(la forma)*

## Summary

`permea project list` enseña, por identidad, la carpeta, las herramientas, los eventos y las fechas, para cruzarlos con el panel. **Es la emisión con otro
destino**:
- **un recorrido común** de los seis lectores, extraído de `generate()` y con un **sumidero**: `generate` encola y guarda el estado, y `project list` agrega
  en memoria sobre un estado vacío que nunca se guarda *(§1)*;
- **una salida más del cuerpo de la derivación**: la ruta que entró en el hash *(§2)*;
- **lectura sin escritura**: el `DataDir` y la sal, sin crearlos *(§3)*;
- la agregación *(§4)*, los nombres comerciales *(§5)* y la tabla *(§6)*.

Riesgos:
1. **Que la emisión cambie al extraer el recorrido.** Lo cubre SC-012, con la referencia del binario de `cabf8c5` tomada **antes** de tocar `main.go`.
2. **Que la lista escriba o abra la red** *(D-6)*. Lo cubren (24) y (25), que son rojos antes del verde, y las mutaciones m18 y m19.
3. **Que lo listado y lo enviado diverjan.** Hay un solo recorrido y un solo cuerpo de derivación, y SC-003 los compara.

## Technical Context

- **Go 1.22.2** · `golangci-lint` 2.12.2 · `goreleaser` v2. **Ninguna dependencia** *(SC-015)*. Sólo la biblioteca estándar ya enlazada: `os`, `io/fs`,
  `path/filepath`, `sort`, `strings`, `time` y `unicode/utf8`.
- **Línea base** *(`cabf8c5`; se transcribe en T001)*: **1074** «--- PASS», 0 FAIL y 0 SKIP *(012 registro §Cierre)*; vet limpio; lint → 0. Los tamaños
  de la 0.9.0 en las 5 plataformas, para SC-016, con el instrumento de 010 E-3.
- **Storage**: ninguno. El comando no escribe nada *(FR-005)*.

## Constitution Check *(1.1.0; antes y después del diseño: sin cambios)*

| Principio | | Razón |
|---|:--:|---|
| I · Frontera inviolable | ✅ | **No cruza nada.** La lista no construye cliente ni importa `internal/transport` *(§1; sonda (24))*. **D-10**: es **el primer comando que imprime rutas reales**: `CARPETA` sale a la terminal de quien lo ejecuta, **en su máquina**, y nunca a la cola, a la red ni a un fichero. El Principio I no se toca: el evento, `boundary_test.go` y los lectores tienen 0 bytes de diff *(FR-025)* |
| II · Local-first | ✅ | Lee los registros locales, en sólo lectura, sin red y sin enrolamiento |
| III · Binario único | ✅ | Sin dependencias, sin cgo y sin paquete nuevo. Rutas por sistema *(P-10, con `filepath`)* |
| IV · Test-first en la frontera | ✅ | La prueba de «ni red ni escritura» *((24), (25))* es **roja antes del verde** de B7: afirma también la tabla, que el esqueleto de B6 no escribe |
| Fixtures | ✅ | Sintéticos. Las rutas son de `t.TempDir()` o de `/sintetico/…` *(§9)* |
| V · Specs | ⚠️ | Spec, plan y tareas antes del código. La spec lleva citas y algún «cómo», como 008–012: el dueño la mantiene, con la enmienda menor en el backlog *(spec §Constitución)* |

## Project Structure

```text
internal/config/config.go · identity.go · lectura_test.go          # B1: RutaDataDir y LeerSalt, sin crear · NUEVO el test
internal/project/resolve.go · origen_test.go                        # B2: DerivarConOrigen y Resolutor.Origen · NUEVO el test
cmd/permea/main.go · codex.go · gemini.go · opencode.go · copilot.go · kimi.go   # B3: el recorrido común y el sumidero
cmd/permea/recorrido_test.go · testdata/lista/**                    # NUEVOS, B0 (referencia) y B3
cmd/permea/lista.go · lista_agregar_test.go · lista_formato_test.go # NUEVOS, B4 y B5: agregación, nombres y tabla
cmd/permea/project.go · ayuda.go · lista_test.go                    # B6: verbo, argumentos, sal, ayuda · NUEVO el test
cmd/permea/lista_frontera_test.go                                   # NUEVO, B7: los seis lectores, sin red y sin escritura
README.md · CHANGELOG.md                                            # B8
specs/013-lista-de-proyectos/soporte/registro.md                    # nace en T001 (techo 250)
```

**Nada más**: ni `internal/event/`, `internal/ingest/`, `internal/state/`, `internal/transport/` o `internal/pricing/`, ni ningún `_test.go` existente.

## 1 · El recorrido común *(B3; FR-009, FR-024; P-9)*

- **`type sumidero func(event.Event) error`**, en `cmd/permea/main.go`.
- **`(a *agent) recorrer(logs []string, st *state.Store, ictx ingest.Context, sumar sumidero) (int, *ingest.Pasada, error)`**. Es **el cuerpo de hoy**
  de `generate()` *(`main.go:300-436`)* desde la Pasada *(`:327`)* hasta Kimi, **sin** cargar ni guardar el estado. Los ficheros de Claude Code llegan
  hechos: `ClaudeCodeLogsRoot` y `FindLogs` *(`main.go:301-308`)* se quedan **antes** de `state.Load` en `generate()`, en su orden de hoy, así que los
  errores salen igual *(FR-024)*. La lista los llama en el mismo orden. Cada `transport.Append(a.dir, ev)` pasa a `sumar(ev)`:
  - en Claude Code, `main.go:348` y `:370`;
  - en `generarCodex`, `main.go:462`;
  - y en `generarGemini`, `generarOpenCode`, `generarCopilot` y `generarKimi`, en `gemini.go:41`, `opencode.go:63`, `copilot.go:50` y `kimi.go:56`.

  Los cinco `generar…` ganan el parámetro `sumar`. Ningún test los llama directamente: los tests llaman a `generate()` *(censo §7)*.
- **`generate()`** conserva su firma *(la usan los tests)*. Resuelve los ficheros de Claude Code, carga el estado, pone `ictx.Resolutor`, llama a `recorrer` con
  `func(ev) error { return transport.Append(a.dir, ev) }` y guarda el estado como hoy *(`main.go:436`)*. Así el orden de encolado no cambia.
- **`project list`** llama a `recorrer` con `state.New()`, su propio `Resolutor` y un sumidero que agrega *(§4)*. **Nunca** llama a `st.Save` ni a
  `transport`.
- **Raíces sin `setup()`.** `setup()` crea la sal y el `machine_id` *(`main.go:232-239`)*, y `DataDir()` crea el directorio *(`config.go:60`)*. La
  resolución de las otras **cinco** raíces *(`main.go:240-264`)* se extrae a `resolverRaices()`, que comparten `setup()` y la lista. `setup()` conserva su
  firma y su orden, con `resolverRaices()` **tras** `machine_id` y la parada por `plain` **la primera**, como exige `main_test.go:371-413`. La llaman cinco
  tests: `codex_test.go:283`, `gemini_test.go:225`, `opencode_test.go:260`, `copilot_test.go:315` y `kimi_test.go:294`.
- **Dónde vive la orquestación**: `projectList`, en **`cmd/permea/lista.go`**. `runProject` *(`project.go`)* sólo le despacha el verbo. `recorrer` vive en
  `main.go`, que importa `transport` porque `generate` encola.
- **El `ingest.Context` de la lista**: la sal leída; `MachineID`, `DevID` y `OrgID` vacíos, porque el evento no sale; `Pasada` y `Resolutor`, nuevos.

## 2 · La derivación con origen *(B2; FR-010, FR-015, FR-016)*

- **`derivarConTechoYRaiz`** *(`resolve.go:157`)* devuelve además **`origen`**: `raíz` o `ubicación`, que es **la cadena que entra en `event.Ref`**.
  Sigue siendo **el único cuerpo**. `derivarConTecho` y `DerivarConRaiz` descartan el valor nuevo y **no cambian de firma**: los tests usan
  `derivarConTecho` *(`resolve_test.go:245-260`)*.
- **`DerivarConOrigen(cwd, sal) (identidad, origen string, huboRaiz bool)`**, exportada, con `mejorValorDisponible` y `techoDeProduccion` como
  `DerivarConRaiz`.
- **`Resolutor`** gana `origenes map[string]Origen` *(clave: la identidad)* y `Origen(identidad) (Origen, bool)`, con `Origen{Ruta string; HuboRaiz bool}`.
  En un fallo de caché, `Resolutor.Derivar` deriva con `DerivarConOrigen` y apunta el origen. `cache` **sigue siendo** `map[string]string`, porque
  `resolve_test.go:625` y `:689` miden su `len`. Con receptor nil, deriva sin apuntar *(como hoy, `resolve.go:318-321`)*.
- **Dos ayudantes exportados**, para que la tabla no copie la lógica del techo: `HogarReal() (string, bool)`, el hogar con `mejorValorDisponible`
  *(como `techoDeProduccion`, `resolve.go:241-249`)*, y `CuelgaDe(ruta, base string) bool`, un envoltorio de `cuelgaDe` *(`resolve.go:268`)*.
  `cuelgaDe` y `mejorValorDisponible` **no cambian de nombre ni de firma**: `resolve_test.go:790` usa la segunda.

## 3 · Leer sin escribir *(B1; FR-005, FR-007, FR-008)*

- **`config.RutaDataDir() (string, error)`**: la ruta de `DataDir()` **sin** `MkdirAll`. `DataDir()` pasa a usarla y conserva su conducta.
- **`config.LeerSalt(dir) (sal string, existe bool, err error)`**: lee `<dir>/salt` **tal cual**, como `loadOrCreateSecret` *(`identity.go:24-29`)*. Si
  no existe, devuelve `("", false, nil)` y **no la crea**.
- **`config.Load`** ya es de sólo lectura *(`config.go:68-82`)*. **No** se llama a `CheckRetiredProjectRefMode` *(P-8)*.

## 4 · La agregación *(B4; FR-011 a FR-014, FR-016; P-1 a P-4)*

`agregar(evs []event.Event, origen func(ref string) (project.Origen, bool), existe func(ruta string) bool) []fila`, con funciones inyectadas para probarla
sin disco:
- descarta `ProjectRef == ""` *(P-1)*;
- por identidad, un conjunto de `EventID`: el primero que llega gana *(P-2)*. Un `EventID` ya visto en otra identidad no cuenta otra vez;
- el conjunto de `Tool`, y el mínimo y el máximo de `OccurredAt.UTC()` *(FR-013)*;
- la marca: `existe(ruta)` falso → «ya no existe»; si no, `!HuboRaiz` → «sin repositorio» *(P-4)*. En producción, `existe` es `os.Stat` sin
  `fs.ErrNotExist`. Otro error **no** marca;
- el orden de FR-020.

## 5 · Los nombres comerciales *(B4; D-1, P-3)*

**Dónde viven**: en **`cmd/permea/lista.go`**, como una tabla ordenada `[]struct{ id, nombre string }`. Su orden es el de P-3. Un id que no está en la
tabla sale tal cual, detrás, por orden alfabético. **Por qué ahí**:
- es **presentación**, y sólo la usa este comando;
- `internal/event` es la frontera y no se toca *(FR-025)*;
- los lectores emiten el id, no el nombre, y `internal/ingest` tampoco se toca.

Es copia de la del panel [P `frontend/src/catalogo/es/Operador.ts:21-27`] **más `copilot_cli` → GitHub Copilot CLI**, que el panel aún no tiene
*(spec §Dependencias)*.

## 6 · La tabla *(B5; FR-017 a FR-023; P-5, P-10, P-12)*

- **Anchos**: por columna, el máximo en **runas** *(`utf8.RuneCountInString`)* entre la cabecera y sus celdas. `IDENTIDAD` siempre mide 13 *(12 + `…`)*;
  `DESDE` y `HASTA`, 10.
- **Alineación**: `EVENTOS`, a la derecha, **también su cabecera**; las demás, a la izquierda, rellenas con espacios. `HASTA`, la última, **sin
  relleno**. Separación de **2 espacios**. En el ejemplo aprobado, los anchos son **13 · 33 · 18 · 7 · 10 · 10**: 101 runas por fila y 96 la cabecera.
- **`CARPETA`**: `aCasa(ruta, hogar)`, con P-10. El hogar sale de `project.HogarReal()` y la comparación, de `project.CuelgaDe` *(§2)*: no se copia la
  lógica del techo. Después, `" (ya no existe)"` o `" (sin repositorio)"`.
- **Miles**: `miles(n int) string`, con `.` cada tres cifras desde la derecha.
- **Pie**: una línea en blanco y las dos líneas de D-4. `identidad`/`identidades` y `evento`/`eventos`, cada una por su número. Sin filas, **sólo** el
  pie *(P-5)*.
- **Funciones puras**: `formatear(filas []fila, hogar string) string`. El disco no entra en el formato: (13) compara bytes sin sandbox.

## 7 · Censo de tests existentes *(ninguno cambia; si uno tuviera que cambiar, PARA 1)*

| Test o grupo | Por qué podría tocarse | Por qué sigue verde sin tocarlo |
|---|---|---|
| `TestProject_ErroresDeUsoDeLaGramatica` *(`project_test.go:555`)* | el texto de «Verbos disponibles» | Sólo exige salida 1, stderr no vacío y stdout vacío *(`:570-580`)* |
| `TestAyuda_ContenidoMinimo` *(`ayuda_test.go:115`)* | la ayuda general gana una fila | Busca subcadenas, que siguen ahí |
| `TestAyuda_AnchoMaximo80`, `TestAyuda_SinJergaInterna` *(`:311`, `:363`)* | la fila nueva sale en `help` y en `project -h` | Las líneas nuevas miden ≤ 74 runas, sin jerga. `lasCincoAyudas` *(`:299-307`)* **no se amplía**: `project list -h` se prueba en `lista_test.go` (19) |
| `TestAyudaSubcomando_NoHaceNada` *(`:164`)* | — | Su lista *(`:144-145`)* no cambia. La de `list` va en (19) |
| `generate()` en los tests de los seis lectores *(`codex_test.go`, `gemini_test.go`, `opencode_test.go`, `copilot_test.go`, `kimi_test.go`, `main_test.go`, `retencion_test.go`)* | §1 cambia su cuerpo | La firma `(int, *ingest.Pasada, error)` y la conducta no cambian *(SC-012)*. Construyen `&agent{…}` con campos que no cambian |
| `setup()` en `codex_test.go:283`, `gemini_test.go:225`, `opencode_test.go:260`, `copilot_test.go:315` y `kimi_test.go:294`; el orden, en `main_test.go:371-413` | §1 extrae las raíces | Misma firma y mismo resultado; la parada por `plain`, la primera, y las raíces tras `machine_id` |
| `resolve_test.go` | §2 toca el cuerpo | `derivarConTecho`, `Derivar`, `DerivarConRaiz` y `mejorValorDisponible` *(`:790`)* conservan firma y valores; `len(r.cache)` *(`:625`, `:689`)* no cambia |
| `TestProjectJoin_*` | `project.go` gana un `case` | `join` no cambia *(D-9)* |

## 8 · La red de pruebas *(numeradas; las de los bloques, en §Bloques)*

- **Salida exacta**: (13) es literal, con el ejemplo de spec §Textos y sin disco. (22) usa los seis lectores en sandbox, y su esperado se arma en el test:
  la identidad con `event.Ref(sal, raízReal)` y el resto, literal.
- **Frontera**:
  - (24) **sin red**: un `httptest.NewTLSServer` como `endpoint`, con `device_token` y un evento en la cola. El binario hijo **confía** en su certificado
    por `SSL_CERT_FILE`, como `entornoDeAdhesion` *(`project_test.go:57-92`)*: sin eso, un drenaje fallaría en el apretón TLS y daría 0 sin probar nada.
    Las conexiones se cuentan **en el propio servidor** *(`Config.ConnState`, `StateNew`)*. **Sin proxy**: `ProxyFromEnvironment` nunca envía por proxy a
    127.0.0.1. Tras `permea project list`, **0 peticiones y 0 conexiones**, y la tabla esperada por stdout, para que el esqueleto no lo pase. A eso se suma
    una sonda estructural: `lista.go` no importa `internal/transport` ni `net/http` *(`go/parser`)*;
  - (25) **sin escritura**: la huella del sandbox entero, antes y después, con cada ruta *(también los directorios)*, su tipo, tamaño, `mtime` en ns y el md5
    de cada fichero, y la tabla esperada por stdout. Además, el caso `sin_datadir`: con el `DataDir` borrado, sigue sin existir.
- **La sal**: (17), sin sal y con la sal ilegible.
- **Las dos marcas**: (11), en la agregación, y (22), en disco, con una carpeta con `.git`, una sin él, una borrada y una borrada que tenía `.git`.
- **El orden**: (12) y (10), con empates y con un id desconocido.

## 9 · Testdata: SINTÉTICOS

`cmd/permea/testdata/lista/generar.sh` *(bash, como 012 §6)* deja **plantillas** con un marcador `@CASA@`. Cada test las copia a `t.TempDir()` y sustituye
`@CASA@` por el hogar del sandbox:
- un JSONL de Claude Code con mensajes desde dos subdirectorios de `@CASA@/dev/proyecto-ejemplo`, y una copia que repite dos mensajes *(SC-009)*;
- una sesión de Codex en la misma raíz;
- una de Gemini con su `.project_root`;
- una de Copilot CLI;
- una de Kimi Code con su `session_index.jsonl`;
- un registro de una carpeta sin `.git` y otro de una carpeta que el test borra.

OpenCode usa una **copia** de `cmd/permea/testdata/opencode/sana.db` *(`directory` = `/sintetico/sana`, que no existe → `(ya no existe)`, ruta
completa)*. La sal es la de `SandboxConSemillas` *(`internal/testutil/sandbox.go`)*. **Ninguna ruta personal**: `PRIV` → 0 *(SC-017)*.

**La referencia de SC-012 no usa `@CASA@`**: sus fixtures llevan `cwd` fijos bajo `/sintetico/…`, así que sus `project_ref` no dependen del temporal. La
línea `N eventos encolados en <ruta>` *(`main.go:499`)* lleva la ruta del sandbox: antes de comparar, se sustituye por `@DATOS@` en las dos salidas.

## Decisiones de plan

| ID | Decisión | Por qué |
|---|---|---|
| **D-013-P1** `DECIDÍ YO` | Sin paquete nuevo: `cmd/permea/lista.go`, más una salida en `internal/project` y dos funciones en `internal/config` | El patrón de 008–012 |
| **D-013-P2** `DECIDÍ YO` | **Un solo recorrido con sumidero** *(§1)*, y no una copia de `generate()` | FR-009, FR-010 y SC-003: dos recorridos podrían divergir |
| **D-013-P3** `DECIDÍ YO` | **El origen sale del cuerpo compartido** *(§2)*, y no de una función aparte | P-005 FR-005 |
| **D-013-P4** `DECIDÍ YO` | **El mapa de nombres, en `cmd/permea/lista.go`** *(§5)* | Es presentación; la frontera y los lectores no se tocan |
| **D-013-P5** `DECIDÍ YO` | `agregar` y `formatear` son **puras**, con el disco inyectado | (13) sin sandbox; mutaciones con sonda limpia |
| **D-013-P6** | La referencia de SC-012, **antes** de tocar `main.go`, con el binario de `cabf8c5` por `git archive` | Como D-012-P9 |
| **D-013-P7** | SC-016 con `git archive` de las dos versiones; en linux/arm64, los `LOAD` | El instrumento de 010 y 012 |
| **D-013-W** `DECIDÍ YO` | **SC-018** lo mide Claude en este equipo con **stdout a `wc -l`**: nunca una ruta en el registro. Antes y después, la huella del `DataDir` real, sólo con md5. **SC-019** lo hace el dueño, con el panel delante | D-10: las rutas salen sólo a una terminal |

## Bloques

**Fase 0** de cada bloque: las puertas en verde, el censo y el esqueleto que haga compilar los rojos. **Corte**: ✋ commit del dueño. **S** < 50 líneas de
producción · **M** 50–150. Cada mutación, con su **sonda**: el test que debe caer.

| Bloque | FR · SC | Rojos | Mutaciones *(→ la sonda)* | |
|---|---|---|---|---|
| **B0** Fase 0, referencia y registro | SC-012 *(referencia)* | — | — | docs |
| **B1** Leer sin escribir | FR-005, FR-007, FR-008 | **(1)** `RutaDataDir` = la ruta de `DataDir` y **no crea** · **(2)** `LeerSalt`: `ausente` *(sin crear)*, `igual` *(byte a byte con `LoadOrCreateSalt`, también con un `\n` final)* e `ilegible` | **m1** `RutaDataDir` con `MkdirAll` → (1/no_crea) · **m2** `LeerSalt` delega en `LoadOrCreateSalt` → (2/ausente) · **m3** `TrimSpace` sobre la sal → (2/igual) | **S** |
| **B2** La derivación con origen | FR-010, FR-015 | **(3)** `DerivarConOrigen` sobre las familias de `internal/project/testdata`: identidad = `Derivar`; origen = raíz *(desde un subdirectorio)*, ubicación limpia *(fallback, con `/./`)* o ruta real *(enlace)*; `huboRaiz` = `DerivarConRaiz` · **(4)** `Resolutor.Origen` tras `Derivar`, con y sin acierto de caché; nil no apunta; `len(cache)` igual | **m4** origen = `cwd` declarado → (3/subdirectorio), (3/enlace) · **m5** origen sin `Clean` → (3/fallback) · **m6** `origenes` por `cwd` y no por identidad → (4/dos_cwd_una_raiz) | **S** |
| **B3** El recorrido común | FR-009, FR-024 · SC-012 | **(5)** SC-012: `--run` sobre los fixtures de los seis = la referencia de B0, byte a byte *(**nace verde**; lo valida m8)* · **(6)** `recorrer` con un sumidero que recoge: los mismos eventos y en el mismo orden que la cola de (5), y la huella del sandbox igual | **m7** `generarKimi` sigue encolando con `transport.Append` → (6/no_escribe) · **m8** el sumidero de `generate` salta Codex → (5) y los `TestCodex_*Run*` *(censo por nombre)* | **M** |
| **B4** Agregación y nombres | FR-011 a FR-014, FR-016, FR-020 · SC-007 a SC-009 | **(7)** P-2: el mismo `event_id` dos veces → 1; en otra identidad, cuenta en la primera y **no aporta** herramienta ni fechas a la segunda; `solo_repetidos` → sin fila · **(8)** P-1: sin `ProjectRef`, ni fila ni recuento · **(9)** fechas: un `occurred_at` en `+02:00` cerca de la medianoche da su día UTC · **(10)** P-3: `kimi_code`, `zeta` y `claude_code` → `Claude Code, Kimi Code CLI, zeta`; los seis nombres de D-1 · **(11)** marcas: `con_git`, `sin_git`, `borrada` y `borrada_con_git` · **(12)** orden: 7 · 412 · 412 · 1 240, con el empate por identidad | **m9** contar eventos y no `event_id` → (7) · **m10** fecha sin `.UTC()` → (9) · **m11** `(sin repositorio)` antes que `(ya no existe)` → (11/borrada_con_git) | **M** |
| **B5** La tabla | FR-017 a FR-022 · SC-001, SC-010, SC-011 | **(13)** SC-001: el ejemplo de spec §Textos, **literal** · **(14)** miles y singular: 0, 999, 1000, 1234567 y `1 identidad, 1 evento.`; sin filas, sólo el pie · **(15)** `aCasa`: `hogar`, `dentro`, `hermano` *(`<hogar>2`)*, `enlace` y `fuera` · **(16)** ninguna línea acaba en espacio; una ruta con tildes no descuadra | **m12** ancho en bytes → (13), (16/tildes) · **m13** sin punto de miles → (13), (14) · **m14** `strings.HasPrefix` en vez de segmentos → (15/hermano) | **M** |
| **B6** El comando *(con la tubería aún vacía: con sal, sale con 0 sin escribir nada)* | FR-001 a FR-003, FR-006 a FR-008, FR-023 · SC-006, SC-013, SC-014 | **(17)** la sal: `sin_sal` *(el texto, salida 0, stderr vacío y **sin crearla**)* e `ilegible` *(el texto, salida 1, stdout vacío)* · **(18)** argumentos: `x`, y uno con forma de secreto → el texto, salida 1 y nada del argumento · **(19)** ayuda: `list -h` y `--help`, literal, ≤ 80 runas, sin jerga, sin escribir nada; la fila en `help` y en `project -h` · **(20)** verbos: los dos mensajes, literales · **(21)** configuración: ilegible → el texto y salida 1; con `plain` → salida 0, sin el error ni el aviso | **m15** `LoadOrCreateSalt` en vez de `LeerSalt` → (17/sin_crear) · **m16** «Verbos disponibles: join» → (20) ×2 · **m17** se detiene por `plain` → (21/plain) | **S** |
| **B7** Los seis lectores y la frontera *(el verde conecta la tubería)* | FR-004, FR-005, FR-009, FR-010, FR-015, FR-016, FR-023 · SC-002 a SC-005, SC-007, SC-009 | **(22)** SC-002: la salida exacta con los seis, con las cuatro marcas y OpenCode como `/sintetico/sana (ya no existe)`; `aviso`: un fichero de Codex ilegible da su aviso por stderr, la tabla igual y ninguna línea de resumen · **(23)** SC-003: las identidades = las `project_ref` de `--run` en un sandbox gemelo · **(24)** sin red: 0 peticiones, 0 conexiones, la tabla, y la sonda de importaciones · **(25)** sin escritura: la huella igual y la tabla; `sin_machine_id` *(con sal y sin `machine_id`)*; `sin_datadir` · **(26)** SC-009: la copia de Claude Code cuenta una vez | **m18** la lista pasa por `setup()` tras leer la sal → (25/sin_machine_id), (21/plain) · **m19** la lista drena la cola desde `lista.go` *(`transport.Drain`)* → (24/peticiones), (24/conexiones), (24/importaciones) · **m20** la lista con `ictx.Resolutor` nil → (22), (24/tabla), (25/tabla) | **S** |
| **B8** README y CHANGELOG | spec §Limitaciones, Q-1 | `grep -c '^## 0.10.0'` = 0 · `grep -c 'project list'` en el README = 0 | — *(textos aprobados ✋; `cmp`)* | **S** |

**20 mutaciones**, m1–m20, entre 2 y 3 por bloque. **26 tests numerados: 25 rojos y 1 nacido verde**, el (5), que valida m8. (22)–(26) son rojos porque
B6 deja la tubería vacía y todos afirman la tabla.

**Estimación** *(horas de Claude, sin las del dueño)*:

| | B0 | B1 | B2 | B3 | B4 | B5 | B6 | B7 | B8 | C1–C3 |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Producción *(líneas)* | — | ~25 | ~50 | ~60 cambiadas | ~90 | ~70 | ~35 | ~25 | ~40 de texto | — |
| Test *(líneas)* | — | ~80 | ~120 | ~120 | ~160 | ~150 | ~150 | ~260 | — | — |
| Horas | 0,75 | 0,75 | 0,75 | 1 | 1 | 1 | 0,75 | 1 | 0,5 | 1,5 |

**~385 de producción y ~1 040 de test; ~9 h.** Para SC-016: en 011, 440 líneas costaron +33 760 B como mucho.

## Cierre — en tramos, uno por mensaje *(si uno falla, se para y se rehace desde C1)*

| Tramo | Qué | Quién |
|---|---|---|
| **C1** · Puertas | gofmt, vet y lint 0; 1074 + nuevos, 0 SKIP; los 1074 **por nombre**; SC-015; **SC-016** en las 5; SC-017; compilan Windows y darwin | Claude |
| **C2** · Medida | **SC-018**, por D-013-W: `permea project list` del binario de la rama sobre la instalación de WSL, con stdout a `wc -l`. Se transcriben el tiempo, las líneas y la huella del `DataDir` antes y después | Claude |
| **C3** · Snapshot | `goreleaser release --snapshot --clean`, `strings` con los textos *(SC-013)* | Claude |
| **C4** · **W1** | En Windows, con el panel delante: **SC-019**. Q-2: el tiempo con las herramientas que haya | ✋ dueño |
| **C5** · PR | El cuerpo, fuera del repo; la fecha de la versión | Claude redacta |
| **C6** · Fusión | Merge commit | ✋ dueño |
| **C7** · Etiqueta | `v0.10.0` | ✋ dueño |
| **C8** | Canales y `strings` del binario publicado | tras C7 |

## Trazabilidad *(FR → bloque)*

| FR | Bloque | | FR | Bloque | | FR | Bloque | | FR | Bloque |
|---|---|---|---|---|---|---|---|---|---|---|
| 001 | B6 | | 008 | B1 · B6 | | 015 | B2 · B5 · B7 | | 022 | B5 |
| 002 | B6 | | 009 | B3 · B7 | | 016 | B4 · B7 | | 023 | B6 |
| 003 | B6 | | 010 | B2 · B7 | | 017 | B5 | | 024 | B3 · C1 |
| 004 | B7 | | 011 | B4 | | 018 | B5 | | 025 | todos |
| 005 | B1 · B7 | | 012 | B4 · B7 | | 019 | B5 | | | |
| 006 | B6 | | 013 | B4 | | 020 | B4 | | | |
| 007 | B1 · B6 | | 014 | B4 | | 021 | B5 | | | |

**25 / 25 FR con bloque.** Los 19 SC, en `tasks.md` §Cobertura.

## Riesgos

| # | Riesgo | Medida o mitigación |
|---|---|---|
| R-1 | **Extraer el recorrido cambia la emisión** | La referencia de SC-012, antes de tocar `main.go` *(D-013-P6)*; (5) y m8 |
| R-2 | **Una ruta real llega al registro o a un informe** *(D-10)* | D-013-W: stdout a `wc -l`; `PRIV` en cada corte; los testdata, con `@CASA@` |
| R-3 | **Un lector escribe en modo lista** | Todos son de sólo lectura por su spec *(FR-006 de 008–012)*; (25) mira el sandbox entero |
| R-4 | **La huella de (25) falla por el `mtime` de un directorio** que el sistema toca al leer | Sólo `mtime`, no `atime`; si ext4 lo cambiara sin escritura, PARA y se enmienda la sonda |
| R-5 | **Un lector nuevo, más adelante, no pasa por `recorrer`** | Hay un solo recorrido; (23) lo compara con `--run` |

## Complexity Tracking

| Qué | Por qué | Alternativa descartada |
|---|---|---|
| El sumidero en los cinco `generar…` | FR-009: un solo recorrido | Copiar `generate()`: lo listado y lo enviado podrían divergir |
| Un valor más en el cuerpo de la derivación | FR-015: la carpeta es la que entró en el hash | El `cwd` más frecuente: no es la raíz |
