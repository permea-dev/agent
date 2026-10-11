# Implementation Plan: 010 · «Lector de OpenCode»

**Branch**: `010-lector-opencode` | **Date**: 2026-10-11 | **Spec**: [spec.md](./spec.md) *(ratificada el 2026-10-11; Q-5 cerrada)* | **Base**: `82e0b2d`
**Input**: la spec · [`soporte/prueba-sqlite.md`](./soporte/prueba-sqlite.md) · [`soporte/medicion-anthropic.md`](./soporte/medicion-anthropic.md) ·
[`contracts/event-id-opencode.md`](./contracts/event-id-opencode.md) · el plan y las tareas de 009 · `tmp/sonda-sqlite/b-propio/lector/lector.go` *(md5 `f6b47fa8…`; no versionado)*

## Summary

La `0.7.0` lee también OpenCode 2.x. Cada **llamada** *(fila `assistant` completada de `session_message`)* es un evento `tool = opencode`, con las
partidas de P-2 y sin coste *(D-1)*. El evento no cambia *(M-1)*, y Claude Code, Codex y Gemini se leen como en la 0.6.0 *(M-2)*. Riesgos:
1. **Un lector SQLite propio** *(D-3)*: se **porta** el de la prueba *(exactitud, mutaciones y ensayo ya pasados)*, con el descenso por rowid y un contador.
2. **La frontera por rowid** *(§P-3)*: el primer campo nuevo de `state.json`, opcional; y un coste que sigue a lo nuevo *(SC-017)*.
3. **Sólo lectura absoluta** *(FR-006)*: nada en la carpeta de OpenCode, ni un `-shm`.
4. **La tercera y la cuarta variables de entorno** *(M-5)*, aisladas en `internal/config` y en el sandbox *(M-3)*.
5. **≤ 100 KB** en las 5 plataformas *(SC-007)*, con el lector solo en 51–66 KB.

## Technical Context

- **Go 1.22.2** · `golangci-lint` 2.12.2 · `goreleaser` v2. **Ninguna dependencia** *(FR-011)*.
- **Línea base** *(`82e0b2d`; el código, igual que en `4c6cafc`)*: `go test -count=1 -v ./...` → **671** «--- PASS», 0 FAIL, 0 SKIP; lint → 0.
- **Storage**: un campo opcional en la entrada de `state.json` *(D-010-P2)*. La cola no cambia.
- **Sólo para regenerar testdata** *(nunca en el binario ni en la suite)*: `sqlite3` ≥ 3.44 y `python3 -I`.

## Constitution Check *(antes y después del diseño: sin cambios)*

| Principio | | Razón |
|---|:--:|---|
| I · Frontera inviolable | ✅ | El evento no cambia. El `id` sólo entra en el hash y en una huella con sal, en local. `session_id` y `directory`, sólo con sal |
| II · Local-first | ✅ | Sin coste *(D-1)*. Se lee una base local, en sólo lectura |
| III · Binario único | ✅ | Sin dependencias y sin cgo. Un paquete nuevo, de la librería estándar |
| IV · Test-first en la frontera | ✅ | La frontera no se toca. SC-014 siembra centinelas también en `state.json`, y (22) es **rojo antes del verde** de B3 |
| Fixtures | ✅ | Bases SQLite **sintéticas** de `generar.sh`. La constitución dice «JSONL anonimizados»; el principio que se cumple es **nunca datos reales** |
| V · Specs | ✅ | Spec ratificada y contrato antes del código |

## Project Structure

```text
internal/sqlitero/lector.go · lector_test.go · fuzz_test.go          # NUEVOS, B1: el lector portado (D-010-P1)
internal/sqlitero/testdata/{generar.sh, esperado.json, i/, r/, s/, raros/, profundo.db}   # NUEVOS, B1: sintéticos (D-010-P11)
internal/ingest/opencode_eventid.go · opencode_eventid_test.go      # NUEVOS, B2: el contrato
internal/ingest/opencode.go · opencode_llamada_test.go              # NUEVOS, B3: una fila → evento o clase
internal/ingest/opencode_base.go · opencode_base_test.go            # NUEVOS, B4: frontera, pasada, formatos, listado
internal/ingest/opencode_coste_test.go                              # NUEVO, B4: SC-017, sólo con -tags sc017 (D-010-P13)
internal/ingest/testdata/opencode/{generar.sh, esperado.json, …}    # NUEVOS, B3–B4
internal/state/state.go                                             # B4: SÓLO el campo Frontera de FileState (D-010-P2)
internal/config/opencode.go · opencode_test.go                      # NUEVOS, B5 (P-8)
internal/testutil/sandbox.go · sandbox_opencode_test.go             # B5: dos líneas tras :65 · NUEVO
cmd/permea/opencode.go · opencode_test.go · testdata/opencode/**    # NUEVOS, B6–B7
cmd/permea/main.go                                                  # B6: agent, setup, generate, runOnce, tick · B7: dryRun
README.md · CHANGELOG.md                                            # B8
```

**Existentes que se tocan** *(sobre `82e0b2d`)*:
- **`cmd/permea/main.go`**: `agent` gana `opencodeDatos`, `opencodeUnica` y `opencode *ingest.PasadaOpenCode`, tras `gemini` *(:183-185)*; `setup()`, `config.OpenCodeDatos()` con la regla de Gemini *(:230-233, :238-239)*; `generate()`, el bloque **entre** Gemini *(:355-366)* y el único `st.Save` *(:368)*; `runOnce()` y `tick()`, la línea tras la de Gemini *(:439-441, :505-507)*; `dryRun()`, la detección **antes** de Codex *(:536-542)*.
- **`internal/state/state.go:16-21`**: el campo `Frontera`. `Recorrer` *(:105-153)* no cambia. **`internal/testutil/sandbox.go:65`**: dos `t.Setenv`.
- **`internal/config/codex.go:16` y `gemini.go:17`**: sólo el comentario, de «una de las DOS» a «una de las CUATRO». **`README.md`** y **`CHANGELOG.md`**.

**Nada más**: ni `internal/event/`, `project/`, `transport/` o `pricing/`, ni los lectores de Claude Code, Codex y Gemini.

## 1 · El lector, `internal/sqlitero` *(D-3; spec FR-006 a FR-010)*

**Se porta** `lector.go` de la prueba *(525 líneas)* con su lógica intacta *(cabecera, WAL, registros, desbordamiento, columnas, `Estable()`, tope de
carga)*. Sólo cambian la API y el contador; el registro guarda el md5 de partida.

| API | Hace | Spec |
|---|---|---|
| `Abrir(ruta) (*Base, error)` · `Cerrar()` | `O_RDONLY`; el `-wal` con `ReadFile`; **nunca** abre el `-shm` | FR-006, FR-007 |
| `(*Base).Tabla(nombre) (*Tabla, error)` · `(*Tabla).Col(nombre) int` | la raíz y las columnas desde `sqlite_schema`. `WITHOUT ROWID` y UTF-16 → `ErrNoSoportada` | FR-009 |
| `(*Tabla).RecorrerDesde(h, fn)` **nuevo** | sólo rowid > h, en orden. **Poda** las celdas interiores con clave ≤ h: cuesta la profundidad más lo nuevo | §P-3 3, SC-017 |
| `(*Tabla).Buscar(rowid) (fila, ok, err)` **nuevo** | descenso por las interiores: la primera celda con clave ≥ rowid, o el puntero derecho | §P-3 1-2 |
| `(*Base).Estable() bool` | el `-wal` *(tamaño y cabecera)* y el tamaño de la base, como al abrir | FR-008 |
| `(*Base).Paginas() int64` **nuevo** | las páginas leídas, de la base o del WAL. Es para los tests | SC-011, SC-017 |
| errores | `ErrCorrupta` *(estructura imposible)*, `ErrNoSoportada` y los de E/S **sin ruta** *(`PathError` → `Op: Err`)* | FR-010, FR-029 |

## 2 · La frontera *(spec §P-3; FR-004, FR-005)*

- **El campo** *(D-010-P2)*:
  - `FileState` gana `Frontera *Frontera` con la etiqueta `json:"frontera,omitempty"`;
  - ``Frontera{Huella string `json:"huella"`; EnCurso []EnCurso `json:"en_curso,omitempty"`}`` y `EnCurso{Rowid int64; Huella string}`;
  - `Offset` = **H**; `Size` y `ModTime`, los de la base.
  - **Hacia atrás**: los otros lectores no lo llevan *(su JSON no cambia)*; sin `frontera` = primera pasada; un 0.6.0 **ignora** la clave
    *(`encoding/json` descarta lo desconocido)* y la **pierde** al guardar, y al volver a la 0.7.0 hay una pasada completa. **Sin pérdida.**
- **La huella** *(D-010-P3)*: `event.Ref(ictx.Salt, id)`, la misma sal de `Context` *(`internal/ingest/claudecode.go:61`)* que usan Codex y
  Gemini para `session_ref`. Son 64 hex, en local. El `id` no aparece en `state.json` *(SC-014)*.
- **Una pasada por base**, sobre **una** instantánea de `sqlitero`, **sin encolar nada hasta que `Estable()`** *(D-010-P6)*:
  1. **Entrada sin `frontera`** → **pasada completa**: `RecorrerDesde(0)`; se emiten las completadas; la lista se rehace con las en curso, o como
     «abandonadas» si tienen más de 24 h; y no se cuenta «frontera reiniciada».
  2. **Si no**, `Buscar(H)`: si **falta**, o `Ref(sal, id) ≠ Huella` → pasada completa y **«frontera reiniciada» +1**. Si no, cada `EnCurso`, con
     `Buscar(rowid)`: **completada** → se clasifica y sale · **falta** → sale sin contarse · **huella distinta** → reinicio · **sigue en curso** con
     `$.time.created` *(o `time_created`)* de más de 24 h según `ctx.Ahora()` → «abandonada». Después, `RecorrerDesde(H)`.
  3. H = el mayor rowid visto; `Huella` = la de esa fila; la lista **no repite rowids**; se guarda la entrada.
- **`session_v2`** *(D-010-P5)*: se lee **sólo si** la pasada tiene alguna llamada, **una vez por pasada y base**, en un mapa `id → directory`.
  Cuesta el número de sesiones, no de mensajes, y entra en el 1 % de SC-017 *(DISCREPANCIA 1 del reporte)*.

## 3 · Una llamada *(B3; FR-012 a FR-021)*

`LlamadaOpenCode(fila, ctx) (*event.Event, Clase)`, con `ContextoOpenCode{Context, Ahora, SessionID, Directory, Path}` *(B4 los saca de `session_v2`)*. **Orden**: sin identificador → incoherente → evento *(D-010-P14)*. **Partidas** *(P-2)*:
`input` / `output + reasoning` / `cache.read` / `cache.write`, ausente = 0. **Modelo**: `model.id`; si no es texto no vacío, vacío y «sin modelo».
**`occurred_at`** = `time.UnixMilli(time.completed).UTC()`. **Sin coste**: `tool = "opencode"`, 0 y `false`; `$.cost` sólo lo lee `--scan`.
**Referencias**: `event.Ref(sal, session_id)` y `ctx.Resolutor.Derivar(directory, sal)`, como `gemini.go:147-148`; sin sesión, vacías.
**`PasadaOpenCode`** *(nueva)*: `Resumen()` con el texto de spec §Textos; `HayNovedades()` = llamadas, abandonadas, reinicios o bases contadas > 0.

## 4 · Ubicación y activación *(B5; P-8, P-10, M-3, M-5)*

- **`config.OpenCodeDatos() (dir, unica string, err)`**: `OPENCODE_DB` **absoluta** *(`filepath.IsAbs`)* → `unica`; si no, `dir` =
  `$XDG_DATA_HOME/opencode` si no es vacía, o `<UserHomeDir>/.local/share/opencode`. Devuelve la ruta **aunque no exista**.
- **`ingest.ListarOpenCode(dir)`**: `^opencode(-[A-Za-z0-9._-]+)?\.db$`, **menos** `opencode-next.db`, en orden léxico.
- **Activación**: en `generate()`, en **cada** pasada, `os.Stat(unica)` *(un fichero)* o `os.Stat(dir)` *(un directorio)*, como D-009-P8.
- **`sandbox.go`**: tras `:65`, `t.Setenv("XDG_DATA_HOME", "")` y `t.Setenv("OPENCODE_DB", "")`. `XDG_DATA_HOME` no la lee hoy nadie *(grep: 0)*.

## 5 · `--scan` y el contraste *(B7; P-11, FR-025, D-1, D-2)*

- **`esBaseSQLite(path)`** lee 16 bytes y va **antes** que `esSesionCodex` *(`main.go:536`)*: si no, el `Scanner` de Codex leería una base
  binaria como texto.
- **`dryRunOpenCode`**: la pasada de B4 sobre `state.New()` *(completa)*. Escribe las líneas `evento:` *(`main.go:637`)*, el resumen y **el
  contraste** *(Σ `$.cost` de las `assistant`, y `session_v2` − su Σ, con salida = `output + reasoning`)*. Ni estado ni cola.

## 6 · Testdata, fuzzing, tamaño, SC-016 y SC-018

- **Testdata** *(D-010-P11)*: `internal/sqlitero/testdata/generar.sh` e `internal/ingest/testdata/opencode/generar.sh` *(`sqlite3` ≥ 3.44 y
  `python3 -I`; la técnica de la prueba: conexión retenida, transacción a medias, frame roto)*. Relleno determinista *(`hex(zeroblob(n))`)*; como
  las sales del WAL son aleatorias, **`esperado.json`** guarda lo que dio `sqlite3` *(recuentos, sumas, sha256 de `data`, rowids)*. Regenerar =
  correr el guion y versionar, con el md5 en el registro. Las secuencias entre pasadas son **instantáneas versionadas** *(`paso1/`, `paso2/`…)*
  copiadas sobre la misma ruta: **los tests no necesitan `sqlite3`**. **R** rompe el último byte de una hoja de `session_message` de T3, y
  `generar.sh` comprueba que cae **dentro de una celda** *(H9)*. **Ningún test abre un testdata en su sitio**: siempre una copia en `t.TempDir()` *(H14)*.
  Centinelas `CENTINELA-…` y rutas `/sintetico/…`: nada que case con el grep de privacidad, tampoco con `grep -a` en las bases *(H23)*. Raros de SC-015: 1.x, `WITHOUT ROWID`, UTF-16, sin columna, 512 y 65 536 B, reserva.
- **SC-017** *(D-010-P13)*: `//go:build sc017`; el GB, en `t.TempDir()` con `sqlite3`, **nunca en el repo**, y la suite normal no lo compila *(0
  SKIP)*. `go test -tags sc017 -run TestSC017 -count=1 -timeout 90m ./internal/ingest`, tres veces.
- **Fuzzing** *(SC-009)*: `FuzzBase` *(bytes como base)* y `FuzzWAL` *(I sana y un `-wal` arbitrario)*, con semillas de testdata y un tope de **1 s**
  por entrada *(goroutine **con `recover`**: un pánico o un exceso son `t.Fatal`, nunca un binario caído)*. En la suite, sólo las semillas. 10 min: `go test ./internal/sqlitero -run '^$' -fuzz
  '^FuzzBase$' -fuzztime 10m`, e ídem `FuzzWAL`. Un hallazgo queda en `testdata/fuzz/` como regresión.
- **SC-007** *(D-010-P12)*: `git archive` de `4c6cafc` y de `HEAD`, `CGO_ENABLED=0 -trimpath -ldflags "-s -w -X main.version=…"`, 5 plataformas,
  Δ ≤ 102 400 B en cada una. Informativo en el corte de B6 *(lector ya enlazado)*; normativo en C1.
- **SC-016** *(D-010-P10)*: en la Fase 0 de B6, **antes de tocar `main.go`**, el binario de `4c6cafc` escribe el stderr de `--run` *(Claude Code,
  Codex y Gemini, en sandbox, sin OpenCode)* en `cmd/permea/testdata/opencode/referencia-run.stderr`, con su md5.
- **SC-018**: `git diff 4c6cafc` de la frontera de FR-030 y `--diff-filter=M -- '*_test.go'` → vacíos; `os.Getenv` de producción → **4**.

## Decisiones de plan

| ID | Decisión | Por qué |
|---|---|---|
| **D-010-P1** | Paquete **`internal/sqlitero`**: el `lector.go` de la prueba, portado, con la API de §1 | D-3: lo que se probó es lo que entra |
| **D-010-P2** | `FileState.Frontera`, opcional, con `Offset` = H | FR-004. Es el único cambio en `internal/state` |
| **D-010-P3** | La huella es `event.Ref(ictx.Salt, id)` | La sal que ya usan los lectores. Nunca viaja |
| **D-010-P4** | El algoritmo de §2. Las pasadas completas se cuentan aparte de la primera | §P-3 |
| **D-010-P5** | `session_v2`, una vez por pasada y base, sólo si hay llamadas | Sin índices *(N-8)*; acotado por SC-017 |
| **D-010-P6** | **Clasificar en memoria y encolar tras `Estable()`**. Si no se estabiliza, hasta 5 intentos con 50 ms; al quinto, `opencode: base omitida: <base>: <error>`, sin tocar la frontera | FR-008, FR-029. Nada a medias en la cola |
| **D-010-P7** | Errores por base: el `%v` lleva **el nombre** de la base, no su ruta | FR-029. Hay varias bases por canal |
| **D-010-P8** | La 1.x y el esquema desconocido se cuentan con la regla `Size`/`ModTime` de D-009-P5 | FR-022, FR-023 |
| **D-010-P9** | El reloj, `ContextoOpenCode.Ahora` *(nil → `time.Now`)*, **inyectado en todo test de ingest con filas en curso**. En `cmd/permea`, **ningún fixture con filas en curso** *(PARA 11)* | SC-025 sin esperar 24 h, y ningún test que cambie de resultado 24 h después de generar sus fixtures *(H6)* |
| **D-010-P10** | La referencia de SC-016, antes de tocar `main.go` | Como D-009-P11 |
| **D-010-P11** | Los testdata de §6: guiones versionados, `esperado.json` e instantáneas por paso | Los tests no dependen de `sqlite3` |
| **D-010-P12** | SC-007 se mide con `git archive` de las dos versiones, como en la prueba | El mismo instrumento que dio 51–66 KB |
| **D-010-P13** | SC-017 detrás de `-tags sc017` | El GB nunca va al repo, y la suite no tiene SKIP |
| **D-010-P14** | Clasificación: sin identificador → incoherente → evento; «sin modelo», marca sobre el evento | FR-021, por construcción |
| **D-010-P15** | El forzado de SC-021 es **el A de 009** *(D-009-P12)* | Probado |

## Bloques

**Fase 0** de cada bloque: las puertas en verde, el censo y el esqueleto que haga compilar los rojos. **Corte**: ✋ commit del dueño. **S** < 50
líneas de producción · **M** 50–150 · **L** > 150.

| Bloque | FR · SC | Rojos | Mutaciones *(→ lo que debe caer)* | |
|---|---|---|---|---|
| **B0** Documentos | — | — | — | docs |
| **B1** Lector | FR-006 a FR-010 · SC-004, SC-005, SC-009 *(semillas)* | **Fixture por rojo** *(H8; siempre una copia en `t.TempDir()`)*: **(1)** `raros/{p512,p65536,reserva,utf16}.db`: firma, tamaños, reserva; UTF-16 → `ErrNoSoportada` · **(2)** I, R y S = `esperado.json` · **(3)** `raros/desborda.db` *(sin WAL)*: > 8 KB y > 64 KB · **(4)** `Buscar` en `profundo.db` *(sin WAL, filas < 1 KB, ≥ 3 niveles)*: en hoja, en la clave de una interior, ausente, > máximo · **(5)** `RecorrerDesde` en `profundo.db`: sólo > h, en orden, y `Paginas()` ≤ profundidad + hojas nuevas + `sqlite_schema` · **(6)** `raros/without_rowid.db`, tabla ausente, columnas con comillas *(S; sólo `sqlite_schema`)* · **(7)** `Estable()` sobre una copia de I: `sin_cambio` → cierto, `cambia` → falso · **(8)** errores sin la ruta: `inexistente`, `truncada` · **(9)** corruptas hechas por el test sobre `profundo.db`: ciclo, puntero fuera, carga de **2⁶²** *(con `recover`)*, varint corto → error, sin pánico · **(10)** sólo md5 y listado de una copia de I tras leer | **mB1** ignorar el WAL → (2/I), (2/R), (7/sin_cambio) · **mB2** sin suma → (2/R) · **mB3** sin desbordamiento → (2/I), (2/R), (2/S), (3) ×2 · **mB4** frames tras el último commit → (2/I) · **mB5** `RecorrerDesde` sin poda → (5/paginas) · **mB6** `Buscar` con `>` en vez de `≥` → (4/clave) · **mB7** error con ruta → (8/inexistente) · **mB8** sin tope de carga → (9/carga_enorme) y `FuzzBase` *(su semilla `carga_enorme`)* · **mB9** `Estable` siempre cierto → (7/cambia) · **mB10** sin **las dos** detecciones de `WITHOUT ROWID` *(el `CREATE` y el tipo de página)* → (6/without_rowid) *(sale «tipo de página desconocido», no `ErrNoSoportada`)* | **L** |
| **B2** Identidad | FR-013 · SC-014 *(vectores)* | **(11)** los dos vectores · **(12)** el de Gemini, distinto · **(13)** `id` vacío → no ok | **m1** `"opencode"` → `"gemini"` → (11) ×2, (12) · **m2** sin prefijo de longitud → (11) ×2 · **m3** vacío aceptado → (13) · **m4** un componente constante más en el hash → (11) ×2 | **S** |
| **B3** Una llamada | FR-012, FR-014 a FR-021 · SC-012, SC-013 *(clases)*, SC-028 | Cada rojo afirma **sólo su campo**. `ContextoOpenCode` lleva `SessionID`, `Directory` y `Path` *(H17)*. **(14)** SC-012: 100/7/3/40/2 → 100/10/40/2 · **(15)** ausente = 0 · **(16)** incoherente, siete hojas *(`data` no JSON · no objeto · `tokens` no objeto · no numérica · no entera · negativa · `completed` no numérico)* · **(17)** sin identificador: ausente, vacío, no textual · **(18)** `tool`, coste 0, `cost_available = false`, y `$.cost` = 999 no entra · **(19)** `occurred_at` = `completed`, no `created` · **(20)** sin modelo: emitido y marcado · **(21)** refs, `proyecto` de `directory`, y sin sesión → vacías · **(22)** centinelas fuera del evento: **rojo antes del verde** *(el stub no emite; H19)* | **m5** salida sin `reasoning` → (14) · **m6** `cache.write` a lectura → (14) · **m7** `input + cache.read` → (14) · **m8** `cost_available = true` → (18) · **m9** `$.cost` al evento → (18/cost) · **m10** `occurred_at` = `created` → (19) · **m11** el incoherente se emite → (16) ×7 · **m12** sin modelo no se emite → (20) · **m13** `session_ref` sin sal → (21), (22) · **m14** `project_ref` de `path` → (21/proyecto) | **M** |
| **B4** Base y frontera | FR-003 a FR-005, FR-008, FR-022, FR-023 · SC-010, SC-011, SC-013, SC-015 *(base)*, SC-024, SC-025, SC-029, SC-017 *(tag)* | `Ahora` inyectado en **todos** *(H6)*; cada pasada afirma su **línea de resumen literal**. **(23)** `frontera/paso1` *(4 completadas y 1 en curso)* · **(24)** `paso2` sin cambios: 0, y `Paginas()` sólo en los descensos y `sqlite_schema` · **(25)** SC-010 *(`paso3`)* · **(26)** orden inverso · **(27)** fila de la lista borrada *(con una completada en la base)* · **(28)** entrada **presente** sin campo → completa, sin reinicio · **(29)** SC-024, con la aserción de rowids cambiados · **(30)** `borrada_h/`: se borran las de rowid máximo y entran 2 que **reutilizan** rowids ≤ H → `salen`, `reinicio` · **(31)** SC-025, `abandono/paso1–3` *(Ahora +1 h, +25 h, +26 h)* y `1h` · **(32)** SC-013, literal *(con una en curso de 25 h)* e identidad · **(33)** SC-029: inestable 1 vez → reintenta; 5 → omitida y frontera quieta · **(34)** 1.x y esquema desconocido: 1, 0, cambia → 1 · **(35)** `ListarOpenCode` · **(36)** JSON: `frontera` sólo en las bases; un `state.json` de 0.6.0 se lee · **(37)** dos sesiones, dos `directory` · **(38)** huella con sal, sin el `id` | **m15** sin comprobar la huella de H → (29) · **m16** si H falta, se sigue con `RecorrerDesde(H)` sin reinicio → (30/salen), (30/reinicio) · **m17** la lista no se consulta → (25), (26), (27), (31/paso2), (31/paso3) · **m18** abandono antes de consultar → (31/paso2) · **m19** sin abandono → (31/paso2), (31/paso3), (32/literal) · **m20** siempre `RecorrerDesde(0)` → (24), (25), (26), (27), (31/paso2), (31/paso3) · **m21** la borrada cuenta como abandonada → (27) · **m22** entrada presente sin campo = frontera desde `Offset` → (28) · **m23** el reinicio no se cuenta → (29/literal), (30/reinicio) · **m24** sin reintento → (33/una) · **m25** el error avanza la frontera → (33/quieta) · **m26** formato contado siempre → (34/segunda) · **m27** `opencode-next.db` incluida → (35) · **m28** huella sin sal → (38) · **m29** `frontera` sin `omitempty` → (36/otras) y **`TestRetencion_SC009_NadaDelProveedorEnDisco/state_json_cuatro_campos`** *(`retencion_test.go:299`)* | **L** |
| **B5** Ubicación | FR-001, FR-002 · M-3, M-5 · SC-015 *(ubicación)* | **(39)** `OPENCODE_DB` absoluta, relativa, `":memory:"`, vacía; `XDG_DATA_HOME` definida, vacía, ausente; inexistente sin error · **(40)** el sandbox las deja vacías | **m30** XDG vacía como raíz → (39/xdg_vacia) · **m31** relativa aceptada → (39/relativa) · **m32** `":memory:"` aceptada → (39/memory) · **m33** ignorar `OPENCODE_DB` → (39/absoluta) · **m34** sin las líneas del sandbox → (40) | **S** |
| **B6** Integración | FR-002, FR-024, FR-026 a FR-029 · SC-003, SC-006, SC-014, SC-016, SC-020, SC-021 | Fixtures de `cmd` **sin filas en curso** *(H6)*. **(41)** SC-016 contra la referencia, `opencode_db_vacia` y `xdg_vacia` *(nace verde)* · **(42)** sólo OpenCode · **(43)** forzado A · **(44)** la línea, tras Gemini, y el total · **(45)** `tick` · **(46)** SC-003 · **nacen verdes, tras el verde**: **(47)** SC-014: cola y `state.json` · **(48)** SC-006: md5 y listado de I · **(49)** carpeta creada después, mismo `agent` · **(50)** SC-020 | **m35** `a.opencode = NuevaPasadaOpenCode()` siempre *(como m26 de 009)* → (41) ×2, (49), `TestCodexRun_SinRaizEsLaSalidaDeLa040/{codex_home_vacia, raiz_inexistente}`, `TestCodexRun_LineaDeResumen`, `TestGeminiRun_SinRaizEsLaSalidaDeLa050/{gemini_cli_home_vacia, sin_tmp}`, `TestGeminiRun_LineaDeResumen` · **m36** `st.Save` antes de OpenCode → (43) · **m37** `tick` sin predicado → (45) · **m38** OpenCode antes que Gemini → (44) · **m39** existencia sólo en `setup()` → (49) · **m40** el error aborta la pasada → (50/pasada) · **m41** el error avanza la frontera → (50/se_relee) · **m42** la ruta de la base en el evento → (47) · **M-B6a** `generarOpenCode` crea un fichero en la carpeta de datos de OpenCode → (48) | **M** |
| **B7** `--scan` | FR-025, FR-026 *(scan)* · SC-001 y SC-002 *(forma)* | **(51)** base → `evento:`, resumen, contraste y nada en disco · **(52)** `claude`, `codex` y `gemini` como en la 0.6.0 *(nace verde)* · **(53)** contraste: fuera de los mensajes = `session_v2` − Σ, con un fixture de **Δ`reasoning` ≠ 0** *(H13)* | **m43** la detección tras Codex → (51), (53) · **m44** todo es SQLite → (52) ×3, `TestScan_UnEventoPorMensaje`, `TestScan_LineaConCuatroPartidasYEventID`, `TestScan_LineaConDesgloseDeLaEscritura`, `TestScan_LaSalidaQueCreceValeSuMaximo`, `TestCodexScan_EventosYResumen` *(el padre: sus hojas no llegan a correr)*, `TestCodexScan_ClaudeCodeComoLa040`, `TestCodexScan_FormatoAnterior`, `TestGeminiScan_EventosYResumen` *(el padre)*, `TestGeminiScan_ClaudeYCodexComoLa050/{claude, codex}` · **m45** `dryRunOpenCode` guarda estado → (51/disco) · **m46** contraste sin `reasoning` → (53) | **S** |
| **B8** Docs | FR-031 · P-12 | `grep -c '^## 0.7.0'` = 0 · `grep -c '^### OpenCode'` = 0 | — *(textos aprobados ✋; `cmp`)* | **S** |

**59 mutaciones**: mB1–mB10, m1–m46, M-B6a y, en `tasks.md`, M-B1a y M-B1b. Estimación: B1 ~560 + 450 de test · B2 ~35 + 60 · B3 ~150 + 260 · B4 ~260 + 520 · B5 ~30 + 70 ·
B6 ~120 + 330 · B7 ~80 + 120 · B8 ~50 de texto. **~1 235 de producción, ~1 800 de test.** El riesgo de SC-007 es B4 + B6.

## Cierre — en tramos, uno por mensaje *(si uno falla, se para y se rehace desde C1)*

| Tramo | Qué | Quién |
|---|---|---|
| **C1** · Puertas | gofmt, vet y lint 0; 671 + nuevos, 0 SKIP; SC-018; **SC-007** en las 5; compilan Windows y darwin | Claude |
| **C2** · Medidas | **SC-001** y **SC-003**, sobre la copia del 09-10 en su sitio *(`--scan`, y `--run` con `OPENCODE_DB`)* · **SC-002** · la de `anthropic` como segundo control *(24 eventos)* · **SC-009**, 10 min × 2 · **SC-017**, × 3 · las huellas de las copias, antes y después | Claude |
| **C3** · Snapshot | `goreleaser release --snapshot --clean`, `strings` con los textos | Claude |
| **C4** · **W1** | En Windows, en sandbox y sin enrolar *(D-010-W)* | ✋ dueño |
| **C5** · PR | El cuerpo, con «**bloqueada por SC-027**» | Claude redacta |
| **C6** · Fusión | Merge commit *(pregunta 1 del reporte)* | ✋ dueño |
| **C7** · Etiqueta | **Bloqueada hasta SC-027**: P-033 en producción, acreditada con el vector de 0,043348 | ✋ dueño |
| **C8–C9** | Canales y **W2** *(SC-023 en la instalación real)* | tras C7 |

**D-010-W · W1**, con un contador independiente: `sonda-a.exe` *(modernc; no comparte código con `sqlitero`)*, sobre una **copia** de la carpeta.
1. **OpenCode cerrado**: md5 y listado de la carpeta → `--run` → **iguales** *(SC-006)*; eventos = filas completadas del contador, y Σ de partidas =
   Σ de P-2 sobre él *(SC-023)*; un segundo `--run` → 0.
2. **OpenCode abierto**: dos prompts con `--daemon` *(60 s)*; al cerrar, Σ de eventos = las completadas nuevas, una vez cada una; sin ficheros nuevos.

## Trazabilidad *(FR → bloque)*

| FR | Bloque | | FR | Bloque | | FR | Bloque | | FR | Bloque |
|---|---|---|---|---|---|---|---|---|---|---|
| 001 | B5 | | 009 | B1 · B4 | | 017 | B3 | | 025 | B7 |
| 002 | B5 · B6 | | 010 | B1 · C2 | | 018 | B3 | | 026 | B6 · B7 |
| 003 | B4 | | 011 | C1 | | 019 | B3 · B4 | | 027 | B6 |
| 004 | B4 | | 012 | B3 · B4 | | 020 | B3 · B4 · B6 | | 028 | B6 |
| 005 | B4 | | 013 | B2 | | 021 | B3 · B4 | | 029 | B4 · B6 |
| 006 | B1 · B6 | | 014 | B3 | | 022 | B4 | | 030 | todos |
| 007 | B1 | | 015 | B3 | | 023 | B4 | | 031 | B8 · C7 |
| 008 | B1 · B4 | | 016 | B3 | | 024 | B6 | | | |

**31 / 31 FR con bloque.** Los 29 SC, en `tasks.md` §Cobertura.

## Riesgos

| # | Riesgo | Medida o mitigación |
|---|---|---|
| R-1 | SC-007: B4 + B6 pasan de los ~35 KB que quedan | Medida informativa en el corte de B6. Si no cabe, se adelgaza; el tope no sube |
| R-2 | La primera pasada de un historial grande encola de una vez | SC-017 la acota en la función, sin encolar. La cola ya lo admite *(primera pasada de los otros lectores)* |
| R-3 | Sin bloqueos, una lectura concurrente | `Estable()` y D-010-P6. En vivo, 0 reintentos en 176 pares: (33) lo fuerza |
| R-4 | Un test hereda la base real del desarrollador | M-3, y que los `agent` hechos a mano no la lean |
| R-5 | Los forzados de SC-020 y SC-021 dependen de permisos POSIX | `t.Skip` con root o en Windows, como 009 R-7 |
| R-6 | Regenerar los testdata cambia los bytes *(las sales)* | `esperado.json` se regenera con ellos, y el md5 va al registro |
| R-7 | `opencode import` *(Q-4)* | Abierta. Si mueve rowids, la frontera se reinicia y no se pierde nada |

## Complexity Tracking

*Vacío.* El paquete nuevo lo pide D-3. El campo nuevo de `state.json`, la spec *(FR-004)*.
