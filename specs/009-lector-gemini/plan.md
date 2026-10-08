# Implementation Plan: 009 · «Lector de Gemini CLI»

**Branch**: `009-lector-gemini` | **Date**: 2026-10-08 | **Spec**: [spec.md](./spec.md) *(ratificada el 2026-10-08, 12:20)* | **Base**: `15ce93b`
**Input**: la spec · [`soporte/descubrimiento.md`](./soporte/descubrimiento.md) · [`contracts/event-id-gemini.md`](./contracts/event-id-gemini.md) · el plan y las tareas de 008

## Summary

La `0.6.0` lee también Gemini CLI. Cada **respuesta** *(primera aparición de un `id` con `tokens`, también dentro de `$set.messages`)* es
un evento `tool = gemini` con las partidas de D-2 y sin coste *(D-4)*. **El evento no cambia** *(M-1)*, y Claude Code y Codex se leen
como en la 0.5.0 *(M-8)*. El riesgo está en cuatro sitios:
1. **Las repeticiones** *(FR-006, FR-008, P-3)*. La misma respuesta sale hasta tres veces en el fichero, y vuelve en el `$set.messages`
   de una reanudación días después. Hay que saber, sin cambiar `state.json`, qué `id` ya aparecieron antes del offset *(D-009-P1)*.
2. **No reconstruir nunca el estado final** *(Q4)*: el lector recorre apariciones, no mensajes vivos.
3. **No tocar ningún test existente** *(M-8)*. La raíz de Gemini se resuelve en `setup()`, como la de Codex *(D-008-P5)*.
4. **La segunda variable de entorno de producción**, `GEMINI_CLI_HOME` *(M-7)*, aislada en `internal/config` y en el sandbox *(M-9)*.

## Technical Context

- **Go** 1.22.2 · `golangci-lint` 2.12.2 · `goreleaser` v2.16.0. **Ninguna dependencia ni paquete nuevo.**
- **Línea base**, sobre `15ce93b`: `go test -count=1 ./...` → **9 paquetes ok, 574 pass, 0 fail, 0 SKIP**; `golangci-lint run` → **0**.
- **Storage**: **ninguno nuevo**. `state.json` conserva sus cuatro campos *(`internal/state/state.go:16-21`)*, y la cola no cambia.
- **Restricciones**: la frontera de FR-027 sin diff; **ningún `_test.go` existente** se toca; lint a 0; fixtures **sólo sintéticos**,
  sin forma de UUID.

## Constitution Check *(antes del Phase 0 del plan y tras su Phase 1: sin cambios)*

| Principio | Veredicto | Razón |
|---|:--:|---|
| I · Frontera inviolable | ✅ | El evento no cambia. `id` sólo entra en el hash *(contrato)*; `sessionId` y `.project_root`, sólo con sal. Las rutas quedan en el `state.json` local *(FR-017)* |
| II · Local-first | ✅ | Sin coste de Gemini *(D-4)*. Sólo ficheros locales |
| III · Binario único | ✅ | Ni dependencias ni paquetes nuevos |
| IV · Test-first en la frontera | ✅ | La frontera **no se toca**. SC-006 siembra centinelas también en las rutas |
| V · Specs | ✅ | Spec ratificada y contrato antes del código |
| Puertas | ✅ | vet, lint a 0 y suite verde al cerrar cada bloque |

## Project Structure

```text
internal/ingest/gemini_eventid.go · gemini_eventid_test.go    # NUEVOS, B1: derivación del contrato y sus vectores
internal/ingest/gemini.go · gemini_respuesta_test.go          # NUEVOS, B2: una aparición → evento o clase (D-2, refs, coste)
internal/ingest/gemini_contexto.go · gemini_contexto_test.go  # NUEVOS, B3: apariciones, prefijo, repetidas, formato, listado, proyecto
internal/ingest/testdata/gemini/**                            # NUEVOS, B2–B6: fixtures sintéticos (ids «m-0000…», «s-0000…»)
internal/config/gemini.go · gemini_test.go                    # NUEVOS, B4: raíz de Gemini (GEMINI_CLI_HOME)
internal/config/codex.go                                      # B4: SÓLO el comentario de :16 («la ÚNICA lectura» deja de ser cierto)
internal/testutil/sandbox.go                                  # B4 (M-9): una línea tras :63, t.Setenv("GEMINI_CLI_HOME", "")
cmd/permea/gemini.go                                          # NUEVO, B5–B6: generarGemini, esSesionGemini, dryRunGemini
cmd/permea/main.go                                            # B5: agent, setup, generate, runOnce, tick · B6: dryRun
cmd/permea/gemini_test.go · testdata/gemini/**                # NUEVOS, B5–B6: en proceso, en sandbox; la referencia de SC-012
README.md · CHANGELOG.md                                      # B7
```

**Existentes que se tocan, uno a uno**, sobre `15ce93b`:
- `cmd/permea/main.go`:
  - `agent` *(:167-180)*: los campos `geminiRaiz string` y `gemini *ingest.PasadaGemini`, junto a `codexRaiz` y `codex`;
  - `setup()` *(:218-228)*: `config.GeminiRoot()`, con la misma regla de error que Codex *(:219-222)*;
  - `generate()`: el bloque de Gemini **entre** el de Codex *(:330-341)* y el único `st.Save` *(:343)*;
  - `runOnce()`: la línea `gemini:` tras la de Codex *(:411-413)*;
  - `tick()`: tras la de Codex *(:474-476)*, con `HayNovedades`;
  - `dryRun()`: tras la detección de Codex *(:506-512)*, la de Gemini.
- `internal/testutil/sandbox.go:63`: no es un test, es un ayudante *(M-9)*.
- `internal/config/codex.go:16`: sólo el comentario. **Ningún** cambio de conducta, y FR-027 no lo incluye.
- `README.md` y `CHANGELOG.md`: los textos aprobados.

**Ninguno más.** `internal/state/`, `internal/event/`, `internal/project/`, `internal/transport/` y los ficheros de Codex en
`internal/ingest/` no cambian.

## Qué se reutiliza y qué se duplica

| Pieza | Decisión | Por qué |
|---|---|---|
| `state.Store.Recorrer`, `state.New`, `Files` *(`state.go:105-153`)* | **Se reutiliza**, tal cual | Offset, línea parcial *(:130-133)* y truncado *(:113)*, idénticos a Codex |
| `event.Ref`, `project.Resolutor.Derivar` *(`resolve.go:318`)*, `transport.Append`, `ingest.Context` | **Se reutilizan** | La frontera y la identidad de proyecto son las de siempre |
| `cola` y `capturarStderr` *(`cmd/permea/codex_test.go:70,103`)* | **Se reutilizan** desde `gemini_test.go` | Son del paquete de test, y usarlas no toca `codex_test.go` |
| El hash del `event_id` *(`eventid.go:66`, `codex_eventid.go:34`)* | **Se replica** en `gemini_eventid.go` | `hashEventID` fija `claude_code` *(`eventid.go:34`)* y `derivarEventIDCodex` fija `codex`. Parametrizarlos cambiaría la frontera *(FR-027)* |
| `PasadaCodex`, `ContarComprimido`, `ListarCodex`, `leerPrefijoCodex` *(`codex_contexto.go:43-278`)* | **Se duplica la forma**, no el código | Otros recuentos y otro texto, otro patrón de rutas, otro filtro de prefijo. Generalizarlos cambiaría la conducta de Codex |
| La línea `evento:` de `dryRunCodex` *(`main.go:598-600`)* | **Se duplica** el `Printf` en `dryRunGemini` | Un solo literal aprobado, comprobado por test contra 008 §Textos aprobados, sin tocar el camino de Codex |

## Decisiones de plan

| ID | Decisión | Por qué |
|---|---|---|
| **D-009-P1** · *mecanismo de P-4* | Un fichero se abre sólo si su tamaño supera el offset. Si el offset es > 0, se lee el **prefijo** `[0, offset)` hacia delante, con un **filtro de bytes**: la línea 1 siempre *(cabecera: `sessionId`)*, y cualquier otra que **no** empiece por `{"$set"` y contenga `"tokens":{` *(mensaje con tokens)*. De éstas sólo se guarda su `id` en el conjunto `vistos`. Los `$set` del prefijo **no** se decodifican. Después, `Recorrer` lee lo nuevo y clasifica cada aparición | Cada vez que la CLI pone tokens a un mensaje, vuelve a escribir el mensaje entero como línea propia *(`recordMessage`, `recordMessageTokens` y `recordToolCalls` → `pushMessage`; también la conversión `.json` y la reescritura; descubrimiento Q4, Q10)*. Así, un `id` con tokens de un `$set` ya está antes en una línea de mensaje. **Medido**: 1 de 1 en la copia. **Plan B**, si SC-013 no se cumple: sacar el `id` del principio de la línea (`{"id":"…"`, el orden de claves de `newMessage`) sin decodificar |
| **D-009-P2** · *la cabecera* | `session_ref` sale del `sessionId` de la línea 1, se lea en el prefijo o en lo nuevo. Si no es una cabecera *(objeto con `sessionId` y `projectHash`)*, `session_ref` va vacío | FR-015, FR-016 |
| **D-009-P3** · *el proyecto* | `.project_root` se lee **una vez por pasada y carpeta `<slug>`**, con una caché en la `PasadaGemini`. Su texto, sin espacios finales, entra en `ictx.Resolutor.Derivar`. Si no existe, `project_ref` vacío; cualquier otro error omite el fichero *(FR-025)* | FR-014. El subagente sube desde `chats/<padre>/` hasta `<slug>` |
| **D-009-P4** · *orden (P-7)* | `ListarGemini` devuelve primero los `.jsonl` de las carpetas `<slug>` **con** `.project_root` y después los demás, cada grupo en orden léxico. Y aparte, los `.json` | Dentro de la pasada, la primera aparición gana *(FR-008)*, y así lleva proyecto si alguna copia lo tiene |
| **D-009-P5** · *formato anterior* | Un `.json` de `chats/` tiene su entrada en `state.json` con `Offset = Size`, y se cuenta cuando no existe o cambian `Size` o `ModTime`, como D-008-P2 | FR-018, sin abrirlo |
| **D-009-P6** | Los recuentos viven en `PasadaGemini` *(nuevo, `internal/ingest`)*, con `Resumen()` *(texto aprobado)* y `HayNovedades()` = respuestas > 0 o formato anterior > 0. **`Pasada` y `PasadaCodex` no se tocan** | Sin tocar textos ni tests de 007 y 008 |
| **D-009-P7** | `generate()` conserva su firma. El total suma Gemini, y la `PasadaGemini` queda en `agent.gemini` | Como D-008-P4 |
| **D-009-P8** | La **ruta** de la raíz se resuelve en `setup()` y se guarda en `agent.geminiRaiz`. En **cada** `generate()` se mira `os.Stat(<raíz>/tmp)` y si es un directorio *(sin resolver enlaces; P-9 rechazada)*. Ruta vacía = no se lee | FR-002, como D-008-P5. Los `agent{…}` hechos a mano no leen Gemini |
| **D-009-P9** | `config.GeminiRoot()` es la **segunda** lectura de entorno de producción. El comentario de `codex.go:16` pasa a decir «una de las dos» | M-7; `grep os.Getenv` debe dar 2 |
| **D-009-P10** | **Errores por fichero**, como D-008-P11: `gemini: fichero omitido: %v`, el estado de ese fichero como estaba, y sigue. `transport.Append` sigue siendo fatal. Línea corrupta: en lo nuevo, `skip (línea corrupta): …` *(`main.go:295`)*; en el prefijo, nada | FR-020, FR-025 |
| **D-009-P11** · *la referencia de SC-012* | En la Fase 0 de B5, **antes de tocar `main.go`**, el binario de `15ce93b` genera el stderr de `--run` sobre fixtures de Claude Code y de Codex. Se guarda en `cmd/permea/testdata/gemini/referencia-run.stderr`, con la ruta de datos como `<DATOS>`, y se transcribe su md5 | Como D-008-P12 |
| **D-009-P12** · *el forzado de SC-017* | El **candidato A** de D-008-P10: cola vacía creada antes, y el directorio de datos sin escritura. `Load` pasa *(`state.go:32-48`)*, `Append` abre con `O_APPEND` *(`internal/transport/queue.go:28`)* y `Save` falla en `CreateTemp` *(`state.go:58`)*. `t.Skip` con root o en Windows | Probado en 008 B5 |
| **D-009-P13** · *`--scan`* | `dryRun` mira Codex primero y después Gemini. `esSesionGemini` lee la primera línea con el mismo `Scanner` de 1 MiB: es Gemini si es un objeto con `sessionId` y `projectHash` textuales **y sin** `type`. Entonces `dryRunGemini` aplica `LeerFicheroGemini` sobre `state.New()` | P-8. Las líneas de Claude Code también llevan `sessionId`, pero siempre con `type` |
| **D-009-P14** | Clasificación de cada aparición con `tokens`: sin identificador → incoherente → repetida → evento. «Sin modelo» y «total descuadrado» se marcan sobre el evento | FR-019 cumple la identidad por construcción |

## Disciplinas

**Las once de 008**, sin cambios: las nueve de 006, la décima de 007 *(las copias se leen en su sitio)* y la undécima *(fixtures sólo
sintéticos)*. **Una nueva:**

12. **La segunda copia congelada sólo se lee.** El `--run` de C2 la usa como `GEMINI_CLI_HOME`, con el directorio de datos y la cola del
    agente en el sandbox. Su huella se toma antes y después, y la ruta no entra en el repo.

## Bloques

**Fase 0** de cada bloque: las puertas en verde sobre el commit anterior, el censo del bloque en `tasks.md` y el esqueleto que haga
compilar los rojos. **Corte**: ✋ commit del dueño, con las transcripciones. Tamaños: **S** < 50 líneas de producción · **M** 50–150 · **L** > 150.

| Bloque | FR · SC | Producto | Rojos | Mutaciones *(lo que debe caer)* | Tamaño |
|---|---|---|---|---|---|
| **B0 · Documentos** | — | sólo `specs/` | — | — | docs |
| **B1 · Identidad** | FR-007 · SC-006 *(vectores)* | `gemini_eventid.go` | **(1)** los dos vectores · **(2)** el del espacio de Codex, distinto · **(3)** `id` vacío → no ok | **(m1)** `"gemini"` → `"codex"` → (1), (2) · **(m2)** sin prefijo de longitud → (1) · **(m3)** vacío aceptado → (3) · **(m4)** el `sessionId` en el hash → (1) | **S** |
| **B2 · Una aparición** | FR-009 a FR-015, FR-017, FR-019 · SC-005, SC-007 | `gemini.go` | **(4)** SC-007: 100/40/5/7/3 → 65/40/0/10 · **(5)** partida ausente = 0 · **(6)** incoherente, seis hojas *(`tokens` no objeto · no numérica · negativa · `cached > input` · sin `timestamp` · mal formado)* · **(7)** sin identificador: ausente, vacío, no textual · **(8)** `tool`, coste 0, `cost_available = false` · **(9)** `occurred_at` = `timestamp` · **(10)** sin modelo, emitido y marcado · **(11)** total descuadrado, emitido y marcado; sin `total`, no · **(12)** `session_ref` y `project_ref` | **(m5)** sin `tool` → (4) · **(m6)** sin `thoughts` → (4) · **(m7)** sin restar `cached` → (4) · **(m8)** `cost_available = true` → (8) · **(m9)** el incoherente se emite → (6) ×6 · **(m10)** `occurred_at` = ahora → (9) · **(m11)** descuadrado = incoherente → (11) · **(m12)** `session_ref` sin sal → (12) y la hoja de centinelas | **M** |
| **B3 · Fichero y contexto** | FR-004 a FR-006, FR-008, FR-014, FR-016, FR-018 a FR-020 · SC-004, SC-008 a SC-011, SC-013 *(informativa)* | `gemini_contexto.go` | **(13)** apariciones de mensaje y de `$set.messages`; `tokens: null` no cuenta · **(14)** SC-009: tokens tardíos → 1 · **(15)** SC-009: un `$set` final sin tokens no borra eventos · **(16)** repetidas en la pasada · **(17)** SC-004: 8 + 2, repetidas 4 + 2 · **(18)** dos pasadas = una *(`event_id`, modelo, `project_ref`, `session_ref`)* · **(19)** la cabecera en el prefijo · **(20)** `.json` cuenta 1, 0, 1 · **(21)** SC-008: el resumen literal y la identidad · **(22)** línea corrupta: aviso en lo nuevo, silencio en el prefijo · **(23)** `ListarGemini`: patrón, exclusiones *(`logs.json`, `*.unreadable-*`, `*.tmp-*`)* y orden de P-7 · **(24)** proyecto: con, sin, subagente, y `.project_root` ilegible → error · **(25)** truncado → se relee sin prefijo | **(m13)** sin prefijo → (17/segunda), (18), (19) · **(m14)** prefijo sin `vistos` *(P-3 (b))* → (17/segunda) · **(m15)** `$set.messages` reemplaza el estado → (15) · **(m16)** la aparición sin tokens bloquea el `id` → (14) · **(m17)** sin repetidas → (16), (21) · **(m18)** el `.json` en cada pasada → (20/segunda) · **(m19)** el aviso también en el prefijo → (22/segunda) · **(m20)** orden sólo léxico → (23/orden) · **(m21)** el `.project_root` del subagente en `chats/<padre>/` → (24/subagente) · **(m22)** sin `.project_root` es error → (24/sin) | **L** |
| **B4 · Raíz** | FR-001 · M-9 | `internal/config/gemini.go`, `sandbox.go`, comentario de `codex.go` | **(26)** `GEMINI_CLI_HOME` definida, vacía y ausente · **(27)** una raíz inexistente se devuelve sin error | **(m23)** la vacía tomada como raíz → (26/vacía) · **(m24)** ignorar la variable → (26/definida) · **(m25)** exigir que exista → (27) | **S** |
| **B5 · Integración** | FR-002, FR-003, FR-017, FR-021, FR-023 a FR-026 · SC-003, SC-006, SC-012, SC-016, SC-017 | `cmd/permea/gemini.go`, `main.go` | **(28)** SC-012, contra la referencia de D-009-P11 · **(29)** sólo Gemini · **(30)** SC-017, forzado A · **(31)** la línea `gemini:` literal, tras la de Codex, y «N eventos encolados» la incluye · **(32)** el demonio calla sin novedades · **(33)** SC-003: la segunda pasada da 0 · **(34)** SC-006: centinelas en `id`, `sessionId`, `projectHash`, `.project_root`, nombre del fichero y carpeta del subagente · **(35)** la raíz creada después: la pasada siguiente del mismo `agent` emite · **(36)** SC-016: fichero ilegible | **(m26)** la línea siempre → (28) · **(m27)** `st.Save` antes de Gemini → (30) · **(m28)** `tick` sin predicado → (32/tick) · **(m29)** Gemini sólo con raíz de Claude → (29), (33), (35) · **(m30)** la existencia sólo en `setup()` → (35) · **(m31)** el error aborta la pasada → (36/pasada) · **(m32)** el omitido guarda su offset → (36/se_relee) · **(m33)** la ruta de `.project_root` en el evento → (34) · **(m34)** Gemini antes que Codex en stderr → (31) | **M** |
| **B6 · `--scan`** | FR-022, FR-023 · SC-015 *(línea)* | `gemini.go`, `main.go` *(dryRun)* | **(37)** un fichero de Gemini → líneas `evento:` literales, resumen y nada en disco · **(38)** Claude Code y Codex → la salida de la 0.5.0 | **(m35)** sin detección → (37) · **(m36)** todo es Gemini → (38) · **(m37)** `dryRunGemini` guarda estado → (37/nada_en_disco) · **(m38)** detección sólo por `sessionId` → (38/claude) | **S** |
| **B7 · README y CHANGELOG** | FR-028 · SC-015 | `README.md`, `CHANGELOG.md` | `grep -c '^## 0.6.0'` = 0 · `grep -c '^### Gemini CLI'` = 0 | — *(`cmp` con la spec)* | **S** |

Estimación: B1 ~35 + 60 de test · B2 ~110 + 220 · B3 ~210 + 380 · B4 ~25 + 50 · B5 ~110 + 300 · B6 ~55 + 90 · B7 ~40 de texto.
**~545 de producción y ~1 100 de test.**

## El contador independiente *(SC-001 a SC-004)*

- **Dónde vive**: en un `mktemp -d /tmp/permea-009-…`, nunca en el repo. Es Python con `-I` y sin código del agente. Aplica D-2 y la primera
  aparición con `tokens` por `id`, `$set.messages` incluido, y nunca reconstruye el estado. Sólo imprime recuentos, sumas y nombres de modelo.
- **Las medidas de C2**:
  - el contador, sobre el `.jsonl` de la **primera** copia congelada, en su sitio;
  - `--scan` sobre ese mismo fichero, sumando cada partida con `awk`;
  - `--run` dos veces en `env -i`, con `HOME` y `XDG_CONFIG_HOME` en el sandbox, `CODEX_HOME` vacía y `GEMINI_CLI_HOME` en la **segunda copia
    congelada**, sin enrolar y sin endpoint. La segunda debe dar 0 eventos `gemini`;
  - las huellas de las dos copias, antes y después: iguales.
- **SC-013**: un `.jsonl` sintético de ≥ 100 MB *(muchas apariciones y `$set.messages` largos)*, generado en el temporal; una respuesta
  nueva al final; tres medidas de la pasada.

## Cierre — en tramos, uno por mensaje *(si un tramo falla, se para y se rehace desde C1)*

| Tramo | Qué | Quién |
|---|---|---|
| **C1** · Puertas | `gofmt -l`, `go vet`, `golangci-lint run` → 0, `go test -count=1 ./...` → 574 + nuevos, 0 SKIP en Linux. La frontera de FR-027, sin diff. Ningún `_test.go` existente cambiado. `os.Getenv` de producción → 2. Compilan Windows y darwin. `PENDIENTE` = 1 | Claude |
| **C2** · Medidas | SC-001 a SC-004 contra el contador *(§Contador)*; SC-013, tres veces; las huellas de las dos copias, antes y después | Claude |
| **C3** · Snapshot | `goreleaser release --snapshot --clean`, la huella del zip de Windows y `strings` con los textos aprobados | Claude |
| **C4** · **W1** | En Windows, en sandbox y sin enrolar: `--version`; `status`; `--run` con el `.gemini` real en **sólo lectura**, contrastado con el contador; un segundo `--run` → 0 `gemini`. **Q-1**: el `project_ref` de Gemini frente al de Claude Code para el mismo directorio | ✋ dueño |
| **C5** · PR | El cuerpo del PR y `## 0.6.0 — <fecha>` *(`PENDIENTE` → 0)* | Claude redacta |
| **C6** · Fusión | Con merge commit | ✋ dueño |
| **C7** · Etiqueta | `git tag -a v0.6.0` sobre `main` | ✋ dueño |
| **C8** · Canales | `gh release view v0.6.0`; Scoop y el cask en `0.6.0`; `strings` del binario publicado | Claude |
| **C9** · **W2** | En la instalación real: `scoop update`, `--version` `0.6.0`, `status`, `--run`. En la plataforma, eventos `gemini` = respuestas nuevas del contador | ✋ dueño |

## Trazabilidad *(FR → bloque → rojo)*

| FR | Bloque | Rojo | | FR | Bloque | Rojo |
|---|---|---|---|---|---|---|
| 001 | B4 | (26), (27) | | 015 | B2 · B3 | (12), (19) |
| 002 | B5 | (35), (28) | | 016 | B3 | (19) |
| 003 | B3 · B5 | (23), (29) | | 017 | B2 · B5 | (12) · (34) |
| 004 | B3 | (18), (20), (25) | | 018 | B3 | (20) |
| 005 | B3 | (17), (18) · SC-013 en C2 | | 019 | B2 · B3 | (6), (10), (11), (21) |
| 006 | B3 | (13), (15) | | 020 | B3 | (22) |
| 007 | B1 | (1)–(3) | | 021 | B5 | (31), (32) |
| 008 | B3 | (16), (17) | | 022 | B6 | (37) |
| 009 | B2 | (7) | | 023 | B5 · B6 | (28), (38) |
| 010 | B2 | (4), (5) | | 024 | B5 | (30) |
| 011 | B2 | (10) | | 025 | B3 · B5 | (24) · (36) |
| 012 | B2 | (9) | | 026 | B5 | (29) |
| 013 | B2 | (8) | | 027 | todos *(puerta)* | — |
| 014 | B2 · B3 | (12), (24) | | 028 | B7 | `cmp` |

**28 / 28 requisitos con bloque.** Los 17 SC, en `tasks.md` §Cobertura.

## Riesgos medidos

| # | Riesgo | Medida o mitigación |
|---|---|---|
| R-1 | Un `id` con tokens que sólo esté en un `$set` del prefijo se reemitiría | D-009-P1 lo descarta por el código *(Q4)*, y en la copia es 1 de 1. Si ocurre, la plataforma lo descarta por `(org_id, event_id)` |
| R-2 | Las líneas `$set.messages` llevan el historial entero | En la copia, la mayor mide 4 703 B. `Recorrer` no tiene tope; en `--scan`, el tope de 1 MiB sólo afecta a la primera línea *(D-009-P13)* |
| R-3 | Un test de proceso hereda el `GEMINI_CLI_HOME` o el `~/.gemini` del desarrollador | M-9, y D-009-P8 para los `agent` hechos a mano |
| R-4 | El prefiltro da falsos positivos *(una herramienta devuelve `"tokens":{`)* | Sólo cuestan un decodificado. Lo confirma el JSON |
| R-5 | `project_ref` en Windows, con `.project_root` en minúsculas | Q-1: se comprueba en W1 |
| R-6 | La retención de 30 días borra sesiones no leídas | Límite declarado. El demonio lee cada 60 s |
| R-7 | Los forzados de SC-016 y SC-017 dependen de permisos POSIX | `t.Skip` con root o en Windows, declarado, como 008 R-8 |
| R-8 | La CLI reescribe un fichero **más largo** que el offset | Sólo ocurre si no pudo releerlo *(Q4)*. Se sigue en mitad de una línea, que se salta como corrupta. Sin caso medible |
| R-9 | Una lectura que falla **a mitad** ya sumó a la `PasadaGemini` | Como 008 R-9: el estado no avanza, y la línea de esa pasada cuenta de más |
| R-10 | La raíz es un enlace simbólico | N-11: ni se garantiza ni se prueba |

## Complexity Tracking

*Vacío: sin violaciones.*

## Enmiendas

- **Ratificación** *(2026-10-08, 12:20; registro en `spec.md`)*: P-9 rechazada. D-009-P8 no resuelve enlaces, y C2 mide `--run` sobre la
  segunda copia congelada *(disciplina 12)*.
- **Ratificación de las preguntas del plan** *(2026-10-08, 12:35, Madrid; el dueño, como se recomendaron)*: **(1)** D-009-P1 no decodifica
  los `$set` del prefijo, con su plan B escrito; **(2)** el comentario de `internal/config/codex.go:16` se corrige en el bloque que toca ese
  fichero *(B4)*; **(3)** B6 y B7 en un solo commit *(T047)*; **(4)** SC-013 sobre un sintético de ≥ 100 MB con muchos `$set.messages` largos.
