# Implementation Plan: 008 · «Lector de Codex»

**Branch**: `008-lector-codex` | **Date**: 2026-10-07 | **Spec**: [spec.md](./spec.md) *(ratificada, E-1, E-2)* | **Base**: `7b8c77c`
**Input**: la spec · [`soporte/descubrimiento.md`](./soporte/descubrimiento.md) · [`contracts/event-id-codex.md`](./contracts/event-id-codex.md) · [`quickstart.md`](./quickstart.md)

## Summary

La `0.5.0` lee también Codex CLI. Cada `token_usage_record` es un evento `tool = codex`, con sus tokens y sin coste *(D-3)*. **El
evento no cambia** *(M-1)*, y Claude Code se lee como en la 0.4.0 *(M-8)*. El riesgo está en tres sitios:
1. **El contexto entre pasadas** *(FR-005, SC-014)*. El modelo, el `cwd` y la sesión están **antes** del offset guardado. Hay que
   recuperarlos sin cambiar `state.json` y en ≤ 3 s sobre 100 MB *(D-008-P1)*.
2. **No tocar ningún test existente** *(M-8)*. `generate()` conserva su firma, y la raíz de Codex se resuelve en `setup()`, así que los
   tests que construyen `agent{…}` a mano no leen Codex *(D-008-P5)*.
3. **La primera variable de entorno de producción**, `CODEX_HOME` *(M-7)*. Se aísla en `internal/config` y en el sandbox de test *(M-9)*.

## Technical Context

- **Go** 1.22.2 · `golangci-lint` 2.12.2 · `goreleaser` v2.16.0. **Ninguna dependencia ni paquete nuevo**: los `.zst` se cuentan, no se
  descomprimen *(P-3)*.
- **Línea base**, sobre `7b8c77c`: `go test -count=1 ./...` → **9 paquetes ok, 494 pass, 0 fail**; `golangci-lint run` → **0 issues**.
- **Storage**: **ninguno nuevo**. `state.json` conserva sus cuatro campos *(FR-004)*, y la cola no cambia.
- **Restricciones**:
  - la frontera de M-1, sin diff;
  - **ningún `_test.go` existente** se toca;
  - lint a 0;
  - fixtures **sólo sintéticos**.

## Constitution Check *(antes del Phase 0 del plan y tras su Phase 1: sin cambios)*

| Principio | Veredicto | Razón |
|---|:--:|---|
| I · Frontera inviolable | ✅ | El evento no cambia. Los identificadores del proveedor sólo entran en un hash *(contrato)* o en `event.Ref` con sal. La ruta del fichero queda en `state.json` local y no viaja *(FR-016, E-1)* |
| II · Local-first | ✅ | El agente no calcula coste de Codex *(D-3)*. Lee ficheros locales y no llama a nada |
| III · Binario único | ✅ | Ni dependencias ni paquetes nuevos. Por eso los `.zst` no se leen |
| IV · Test-first en la frontera | ✅ | La frontera **no se toca**. SC-006 siembra centinelas también en el nombre del fichero |
| V · Specs | ✅ | Spec ratificada y contrato antes del código |
| Puertas | ✅ | vet, lint a 0 y suite verde al cerrar cada bloque |

## Project Structure

```text
internal/ingest/codex_eventid.go · codex_eventid_test.go    # NUEVOS, B1: derivación del contrato y sus vectores
internal/ingest/codex.go · codex_linea_test.go              # NUEVOS, B2: una línea → evento o clase (partidas, refs, coste)
internal/ingest/codex_contexto.go · codex_contexto_test.go  # NUEVOS, B3: contexto, modelo, formato, recuentos, repetidas
internal/ingest/testdata/codex/*.jsonl                      # NUEVOS, B2–B6: fixtures sintéticos
internal/config/codex.go · codex_test.go                    # NUEVOS, B4: raíz de Codex (CODEX_HOME)
internal/testutil/sandbox.go                                # B4 (M-9): una línea, t.Setenv("CODEX_HOME", "")
cmd/permea/main.go                                          # B5: setup (raíz), generate (Codex tras Claude), runOnce y tick (línea); B6: dryRun
cmd/permea/codex_test.go                                    # NUEVO, B5–B6: de proceso, en sandbox
README.md · CHANGELOG.md                                    # B7
```

**Existentes que se tocan, uno a uno**:
- `internal/testutil/sandbox.go`: no es un test, es un ayudante *(M-9)*;
- `cmd/permea/main.go`: la integración;
- `README.md` y `CHANGELOG.md`: los textos aprobados.

**Ninguno más.** `internal/state/` no cambia, porque `Recorrer` se usa tal cual.

## Decisiones de plan

| ID | Decisión | Por qué |
|---|---|---|
| **D-008-P1** · *mecanismo de E-1.1* | Un fichero se abre sólo si su tamaño supera el offset guardado. Primero se lee el **prefijo** `[0, offset)` hacia delante, con un **filtro de bytes**: sólo se decodifica una línea que contenga `"session_meta"`, `"turn_context"`, `"thread_settings_applied"`, `"token_usage_record"` o `"token_count"`. El filtro es sólo un prefiltro, y el decodificado confirma el tipo. Después, `Recorrer` lee lo nuevo y emite | Un solo sentido de lectura y un solo analizador. El coste lo dominan la E/S y `bytes.Contains`, no el JSON. **Plan B**, si SC-014 no se cumple: buscar hacia atrás desde el offset, por bloques, el último `turn_context` y el último `thread_settings_applied`, y leer sólo la línea 1. Bastan para FR-011 y FR-014, porque los turnos de un fichero son secuenciales |
| **D-008-P2** · *un `.zst`, una vez* | Cada `.zst` tiene su entrada en `state.json`, con su ruta como clave, `Size` y `ModTime` del stat y `Offset = Size`. Se cuenta cuando la entrada no existe o cambian `Size` o `ModTime` | La forma de `state.json` no cambia, porque sólo se añade una clave. No se abre el fichero |
| **D-008-P3** | Los recuentos de Codex viven en un tipo nuevo, `PasadaCodex`, en `internal/ingest`, con su línea de resumen. **La `Pasada` de Claude Code no se toca** | Sin contaminar los textos ni los tests de 007 |
| **D-008-P4** | `generate()` conserva su firma `(int, *ingest.Pasada, error)`. El total incluye los eventos de Codex, y la `PasadaCodex` queda en un campo del `agent` | `main_test.go:443` y `retencion_test.go:79,239,262` la llaman con esa firma |
| **D-008-P5** *(E-2)* | La **ruta** de la raíz se resuelve en **`setup()`** y se guarda en `agent.codexRaiz`. **Si existe** se comprueba en **cada** `generate()` *(FR-002)*. Una ruta vacía *(sólo los `agent{…}` construidos a mano)* significa que no se lee Codex | Los tests que construyen `agent{…}` a mano *(`main_test.go:441`, `retencion_test.go:57`)* no la tienen y no leen Codex. Y un demonio arrancado antes de instalar Codex lo ve en el ciclo siguiente |
| **D-008-P6** | En `generate()`, primero Claude Code y después Codex, con **un solo** `st.Save` al final. Todo `transport.Append` va antes | FR-025: el orden de hoy *(`main.go:288,310` antes de `:317`)* |
| **D-008-P7** | `config.CodexSessionsRoot()` es el **único** `os.Getenv` de producción, comentado como tal | M-7, y poder buscarlo con un solo `grep` |
| **D-008-P8** | `--scan`: `dryRun` lee la primera línea; si es `session_meta`, sigue por Codex con el mismo `Scanner` de 1 MiB | P-8. Una línea mayor termina en error, como hoy *(R-2)* |
| **D-008-P10** *(E-2)* · *el forzado de SC-015* | **Se elige y se declara en la Fase 0 de B5** *(T030)*. **Candidato A**: crear antes `queue.jsonl` vacía y quitar al directorio de datos el permiso de escritura. `Load` lee, y si `state.json` no existe devuelve un `Store` vacío *(`state.go:32-48`)*. `Append` abre la cola existente con `O_APPEND` *(`internal/transport/queue.go:28`)*, que no necesita escribir en el directorio. `Save` falla en `os.CreateTemp` *(`state.go:58`)*. El test lleva `t.Skip` con root o en Windows. **Candidato B**, si A no se sostiene: un campo `guardar` en el `agent`, `nil` = `st.Save`, como el `reloj` de 007 *(D-007-P6)* | «`state.json` como directorio» hace fallar `Load` antes de encolar. 007 acreditó su FR-012 **leyendo el código**, sin forzado *(`specs/007-coste-fiel/soporte/registro.md` §FR-012)*. A no añade nada a producción |
| **D-008-P11** *(E-2)* | **Los errores de Codex son por fichero.** Un error de lectura escribe `codex: fichero omitido: %v`, deja el estado de ese fichero como estaba y sigue con el siguiente. Un error de `transport.Append` devuelve, como hoy. Una línea corrupta en la parte nueva escribe `skip (línea corrupta): …` *(`main.go:282`)*, y en el prefijo releído no escribe nada | FR-028 y FR-029. Hoy cualquier error devuelve antes de `st.Save` *(`main.go:304-306,317`)* |
| **D-008-P12** *(E-2)* · *la referencia de SC-011* | En la Fase 0 de B5, **antes** de tocar `main.go`, el binario del commit anterior genera el stderr de `--run` sobre el fixture de Claude Code. Se guarda en `cmd/permea/testdata/codex/referencia-run.stderr`, con la ruta del directorio de datos sustituida por `<DATOS>`, y su md5 se transcribe. El test (23) aplica la misma sustitución y compara | «N eventos encolados en <ruta>» *(`main.go:346`)* lleva la ruta del sandbox, que cambia en cada ejecución |
| **D-008-P9** | Clasificación de cada registro: sin identificador → incoherente → repetida → evento, y «sin modelo» se marca sobre el evento | FR-027 hace cumplir la identidad por construcción |

## Disciplinas

**Las nueve de 006 y la décima de 007** *(las copias del dueño se leen en su sitio)*, sin cambios. **Una nueva:**

11. **Fixtures sólo sintéticos.** Ningún fichero del repo lleva datos de las sesiones del dueño. Los identificadores de prueba tienen
    forma sintética *(`r-0000…`)*. El fichero de 100 MB de SC-014 se genera en un `mktemp -d` y se borra.

## Bloques

**Fase 0** de cada bloque: las puertas en verde sobre el commit anterior, el censo del bloque en `tasks.md` y el esqueleto que haga
compilar los rojos. **Corte**: ✋ commit del dueño, con las transcripciones. Tamaños: **S** < 50 líneas de producción · **M** 50–150 · **L** > 150.

| Bloque | FR · SC | Producto | Rojos | Mutaciones *(lo que debe caer)* | Tamaño |
|---|---|---|---|---|---|
| **B0 · Documentos** | — | sólo `specs/` | — | — | docs |
| **B1 · Identidad** | FR-007 · SC-006 *(vectores)* | `codex_eventid.go` | **(1)** los dos vectores de Codex · **(2)** el valor en el espacio de Claude Code es distinto · **(3)** `response_id` vacío → no ok | **(m1)** `"codex"` → `"claude_code"` → (1) · **(m2)** sin prefijo de longitud → (1) · **(m3)** vacío aceptado → (3) | **S** |
| **B2 · Una línea** | FR-006, FR-009, FR-010, FR-012, FR-013, FR-015, FR-016, FR-027 *(incoherente)* · SC-004, SC-008 | `codex.go` | **(4)** partidas de F7 sintético · **(5)** SC-008: 100/40/60 → 0/60/40 · **(6)** incoherente · **(7)** sin `response_id` · **(8)** `tool`, coste a 0 y `cost_available = false` · **(9)** `occurred_at` = marca · **(10)** `session_ref` = `Ref(sal, session_id)`, con `session_id` ≠ `thread_id` | **(m4)** sin restar la escritura → (5) · **(m5)** `cost_available = true` → (8) · **(m6)** incoherente emitido → (6) · **(m7)** `session_ref` del `thread_id` → (10) · **(m8)** `occurred_at` = ahora → (9) · **M-B2a** `session_ref` sin sal → (10) y T013 | **M** |
| **B3 · Contexto y estado** | FR-004, FR-005, FR-008, FR-011, FR-014, FR-017, FR-018, FR-027, FR-029 · SC-005, SC-007, SC-009, SC-010, SC-017 | `codex_contexto.go` | **(11)** modelo del turno frente al vigente · **(12)** compactación → vigente · **(13)** sin modelo → vacío y contado · **(14)** SC-010 en dos pasadas · **(15)** SC-005: reanudación en dos pasadas, 1 + 3 · **(16)** SC-007 · **(17)** formato anterior y fichero mixto *(Q-2)* · **(18)** `.zst` contado una vez *(D-008-P2)* · **(19)** SC-017 · **(20)** `cwd` del turno, y si no, el de `session_meta` · **(34)** *(E-2)* una línea corrupta en la parte nueva se avisa y se salta; en el prefijo, en silencio | **(m9)** siempre el vigente → (11) · **(m10)** sin prefijo *(empezar en el offset)* → (14), (15, la 2.ª) · **(m11)** sin repetidas → (16), (19) · **(m12)** el `.zst` cada pasada → (18) · **(m13)** formato por «hay `token_count`» → (17, mixto) · **(m14)** `cwd` siempre de `session_meta` → (20) · **(m26)** la línea corrupta corta el fichero → (34) · **(m27)** el aviso también en el prefijo → (34/segunda pasada) | **L** |
| **B4 · Raíz y activación** | FR-001, FR-002 · M-9 | `internal/config/codex.go`, `sandbox.go` | **(21)** `CODEX_HOME` definida, vacía y ausente · **(22)** *(E-2)* una raíz inexistente se devuelve igual, sin error: la ruta no decide la activación | **(m15)** `CODEX_HOME=""` tomada como raíz → (21, vacía) · **(m16)** ignorar `CODEX_HOME` → (21, definida) | **S** |
| **B5 · Integración** | FR-002, FR-003, FR-016, FR-019, FR-021, FR-025, FR-026, FR-028 · SC-006, SC-011, SC-015, SC-016, SC-018, SC-001–SC-003 *(fixture)* | `main.go` *(setup, generate, runOnce, tick)* | **(23)** SC-011, contra la referencia de D-008-P12 · **(24)** SC-016 · **(25)** SC-015, con el forzado de D-008-P10 · **(26)** la línea de resumen, literal · **(27)** el demonio calla sin novedades · **(28)** la segunda pasada da 0 · **(31)** SC-006: centinelas, también en el nombre del fichero · **(32)** *(E-2)* sin carpeta, se crea, y la pasada siguiente del mismo `agent` emite · **(33)** *(E-2)* SC-018 | **(m17)** `st.Save` antes del Append de Codex → (25) *(con el forzado, la cola queda vacía)* · **(m18)** la línea siempre → (23) · **(m19)** `tick` escribe siempre → (27) · **(m20)** Codex sólo si hay raíz de Claude → (24) · **M-B5a** el `response_id` en `state.json` → (31) · **(m23)** la existencia sólo en `setup()` → (32) · **(m24)** el error de Codex aborta la pasada → (33) · **(m25)** el fichero omitido guarda su offset al final → (33/se_relee) | **M** |
| **B6 · `--scan`** | FR-020 · SC-013 *(línea)* | `main.go` *(dryRun)* | **(29)** un fichero de Codex → líneas `evento:` literales y resumen · **(30)** un fichero de Claude Code → salida de la 0.4.0 | **(m21)** imprimir `cw5m=`/`cw1h=` → (29) · **(m22)** sin detección → (29) · **M-B6a** todo es Codex → (30) | **S** |
| **B7 · README y CHANGELOG** | FR-023, FR-024 · SC-013 | `README.md`, `CHANGELOG.md` | `grep -c '^## 0.5.0'` = 0 · `grep -c '### Codex CLI'` = 0 | — *(`cmp` con la spec)* | **S** |

Estimación: B1 ~30 + 60 de test · B2 ~90 + 180 · B3 ~170 + 320 · B4 ~25 + 50 · B5 ~80 + 260 · B6 ~40 + 80 · B7 ~30 líneas de texto.
**~435 de producción y ~950 de test.**

## El contador independiente *(SC-001 a SC-003)*

- **Dónde vive**: en `quickstart.md` §Contador, que es el de `descubrimiento.md` §FASE 0 (f). Python con `-I`, sin código del agente.
- **Cómo se lanza**: `python3 -I contador.py <copia congelada>/sessions`, desde un `mktemp -d`. Sólo imprime recuentos, sumas y nombres de
  modelo.
- **Las medidas** *(quickstart §M)*:
  - el binario de la rama, en `env -i` con `HOME`, `XDG_CONFIG_HOME` y `CODEX_HOME=<copia congelada>`, sin enrolar ni endpoint;
  - `--scan` fichero a fichero, y la suma de cada partida con `awk`;
  - `--run` dos veces: la segunda debe dar 0 eventos `codex`.

## Cierre — en tramos, uno por mensaje *(si un tramo falla, se para y se rehace desde C1)*

| Tramo | Qué | Quién |
|---|---|---|
| **C1** · Puertas | `gofmt -l`, `go vet`, `golangci-lint run` → 0, `go test -count=1 ./...` → 494 + nuevos. La frontera, sin diff. Ningún `_test.go` existente cambiado. `grep -rn 'os.Getenv' --include=*.go cmd internal \| grep -v _test` → 1. Compilan Windows y darwin. `PENDIENTE` = 1 | Claude |
| **C2** · Medidas | Sobre la copia congelada en sandbox: SC-001 a SC-005 contra el contador, con la huella antes = después. **SC-014**: tres medidas de coste | Claude |
| **C3** · Snapshot | `goreleaser release --snapshot --clean`, la huella del zip de Windows y `strings` con los textos aprobados | Claude |
| **C4** · **W1** | En Windows, **en sandbox y sin enrolar** *(007 E-7)*. `--version`; `status` → «no enrolado»; `--run` con la carpeta real de Codex en **sólo lectura**, contrastado con el contador; un segundo `--run` → 0 eventos `codex`. Q-6: la raíz es la de Codex | ✋ dueño |
| **C5** · PR | El cuerpo del PR y `## 0.5.0 — <fecha>` *(`PENDIENTE` → 0)* | Claude redacta |
| **C6** · Fusión | Con merge commit | ✋ dueño |
| **C7** · Etiqueta | `git tag -a v0.5.0` sobre `main` | ✋ dueño |
| **C8** · Canales | `gh release view v0.5.0`; Scoop y el cask en `0.5.0`; `strings` del binario publicado | Claude |
| **C9** · **W2** | En la instalación real: `scoop update` → `--version` `0.5.0` → `status` → `--run`. En la plataforma, eventos `codex` = respuestas nuevas del contador en la ventana | ✋ dueño |

## Trazabilidad *(FR → bloque → rojo)*

| FR | Bloque | Rojo | | FR | Bloque | Rojo |
|---|---|---|---|---|---|---|
| 001, 002 | B4 · B5 | (21), (22), (32) | | 015 | B2 | (10) |
| 003 | B5 | (23), (24) | | 016 | B2 · B5 | T013 · (31) |
| 004 | B3 | (14), (18) | | 017 | B3 | (17) |
| 005 | B3 | (14), (15) · SC-014 en C2 | | 018 | B3 | (18) |
| 006 | B2 | (4) | | 019 | B5 | (26), (27) |
| 007 | B1 | (1)–(3) | | 020 | B6 | (29) |
| 008 | B3 | (16) | | 021 | B5 · B6 | (23), (30) |
| 009 | B2 | (7) | | 022 | todos *(puerta)* | — |
| 010 | B2 | (4), (5) | | 023 | B7 · C1 | — |
| 011 | B3 | (11)–(13) | | 024 | B7 | `cmp` |
| 012 | B2 | (9) | | 025 | B5 | (25) |
| 013 | B2 | (8) | | 026 | B5 | (24) |
| 014 | B3 | (20) | | 027 | B2 · B3 | (6), (19) |
| | | | | 028 | B5 | (33) |
| | | | | 029 | B3 | (34) |

**29 / 29 requisitos y 18 / 18 criterios con bloque** *(los SC, en `tasks.md` §Cobertura)*.

## Riesgos medidos

| # | Riesgo | Medida o mitigación |
|---|---|---|
| R-1 | La primera pasada de quien tenga muchas sesiones en formato anterior las lee **enteras una vez** | F1–F4 suman 2 134 619 B *(descubrimiento)*. Después, sin bytes nuevos, no se abren *(D-008-P1)*. Se mide en W1 |
| R-2 | Una línea de más de 1 MiB en `--scan` | 0 en la copia *(máx. 44 745 B)*. `compacted` crece con la historia y no se ha medido. `--scan` termina con error; `--run` no tiene tope |
| R-3 | Un test de proceso hereda el `CODEX_HOME` del desarrollador | M-9 en el sandbox, y D-008-P5 para los que construyen el `agent` a mano |
| R-4 | El prefiltro da falsos positivos *(el texto de una herramienta contiene `"turn_context"`)* | Sólo cuestan un decodificado. El tipo lo confirma el JSON |
| R-5 | En Windows, `os.UserHomeDir()` no es el directorio de Codex | Q-6 (a): se comprueba en W1 |
| R-6 | `gpt-6-luna` aún no tiene tarifa en la plataforma | D-3: llega ciego y señalizado hasta el parche posterior |
| R-7 | SC-014 no se cumple en Windows *(E/S más lenta)* | Se mide en C2 en Linux. Si W1 muestra una pasada lenta, se para y se pasa al plan B de D-008-P1 |
| R-8 | El forzado A de SC-015 y el fichero ilegible de SC-018 dependen de permisos POSIX | `t.Skip` con root o en Windows, declarado. Si la integración continua corre como root, se pasa al candidato B de D-008-P10 |

## Complexity Tracking

*Vacío: sin violaciones.*

## Enmiendas

- **E-1** *(2026-10-07, orquestador; registro en `spec.md`)*: P-2 por resultado, con el mecanismo D-008-P1 y SC-014; la excepción de la
  ruta *(FR-016, SC-006)*; FR-025/SC-015, FR-026/SC-016 y FR-027/SC-017; M-9.
- **E-2** *(2026-10-07, orquestador)*: la activación en cada pasada *(D-008-P5)*; FR-028/SC-018 y FR-029 *(D-008-P11)*; el forzado de SC-015
  *(D-008-P10)*; la referencia de SC-011 *(D-008-P12)*. Rojos (32)–(34) y mutaciones m23–m27.
