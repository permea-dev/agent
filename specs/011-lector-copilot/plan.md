# Implementation Plan: 011 · «Lector de Copilot CLI»

**Branch**: `011-lector-copilot` | **Date**: 2026-10-14 | **Spec**: [spec.md](./spec.md) *(ratificada el 2026-10-14; E-1)* | **Base**: `d81c39d` *(el código, el de `15c4c47`, 0.7.0)*
**Input**: la spec · [`contracts/event-id-copilot.md`](./contracts/event-id-copilot.md) · [`soporte/descubrimiento.md`](./soporte/descubrimiento.md) y las
muestras 2 y 3 · el plan y las tareas de 008 *(el lector de Codex, que es el modelo)* y de 010 *(el formato y el cierre)*

## Summary

La `0.8.0` lee también Copilot CLI. Cada **registro** es un evento `tool = copilot_cli`, con las partidas de D-3 y sin coste *(D-4)*. Un registro es una
línea `session.usage_record` o una `session.compaction_complete` con éxito, de `<raíz>/<carpeta>/events.jsonl`. El evento no cambia *(M-1)*, y los
otros cuatro lectores se leen como en la 0.7.0 *(M-2)*. Es **el lector de Codex con otra forma de línea**:
- offset por fichero con `state.Recorrer`, sin campos nuevos en `state.json` *(FR-004)*;
- prefijo releído para el contexto;
- recuentos y línea de resumen.

Riesgos:
1. **El contexto por tramo** *(P-3)*: el `cwd` del último `start`/`resume`, leído en el prefijo, como `turn_context` en Codex.
2. **Privacidad**: los ficheros llevan prompts, respuestas y rutas. Hay un **prefiltro de bytes** *(P-11)*, sólo se leen los campos de FR-012, y los testdata son **sintéticos** *(§6)*.
3. **≤ 100 KB** frente a la 0.7.0 en las 5 plataformas *(SC-007)*: el riesgo es el código *(estimado, sin medir; §Riesgos R-1)*.
4. **La 5.ª variable de entorno**, `COPILOT_HOME` *(M-5)*, aislada en `internal/config` y en el sandbox *(M-3)*.

## Technical Context

- **Go 1.22.2** · `golangci-lint` 2.12.2 · `goreleaser` v2. **Ninguna dependencia** *(SC-008)*. Sólo `encoding/json`, que ya está enlazado.
- **Línea base** *(`15c4c47`, por `git archive` en un temporal, tras `~/bin/libre.sh`; 14-10)*: `go test -count=1 -v ./...` → **848** «--- PASS», 0 FAIL y
  0 SKIP en 10 paquetes; `go vet` limpio; lint → 0. **Tamaños de la 0.7.0** *(referencia de SC-007)*: linux/amd64 5 648 536 · linux/arm64 5 439 640 *(Σ
  `FileSiz` de los `LOAD`: **5 337 116**)* · darwin/amd64 5 810 800 *(holgura de `__TEXT`: **846 B**)* · darwin/arm64 5 571 330 · windows/amd64 5 846 528.
- **Storage**: ninguno nuevo. Una entrada de `state.json` por `events.jsonl`, con los cuatro campos de siempre *(`internal/state/state.go:16-21`)*, sin `Frontera`.
- **Para regenerar testdata** *(nunca en el binario ni en la suite)*: `bash` y `python3 -I`.

## Constitution Check *(antes y después del diseño: sin cambios)*

| Principio | | Razón |
|---|:--:|---|
| I · Frontera inviolable | ✅ | El evento no cambia. `usageId` y `requestId` sólo entran en el hash; `sessionId` y `cwd`, sólo con sal |
| II · Local-first | ✅ | **Constitución 1.1.0** *(14-10)*: sin coste fiable de la herramienta, el evento va sin coste y lo calcula la plataforma *(D-4)*. Se lee un JSONL local, en sólo lectura *(D-6)* |
| III · Binario único | ✅ | Sin dependencias y sin cgo. Ningún paquete nuevo |
| IV · Test-first en la frontera | ✅ | La frontera no se toca. SC-011 siembra centinelas en todos los campos de texto, y (14) es **rojo antes del verde** de B2 |
| Fixtures | ✅ | JSONL **sintéticos** de `generar.sh`. Nunca una línea de una copia congelada *(§6)* |
| V · Specs | ✅ | Spec ratificada y contrato antes del código |

## Project Structure

```text
internal/config/copilot.go · copilot_test.go                        # NUEVOS, B1 (D-5)
internal/ingest/copilot_eventid.go · copilot_eventid_test.go        # NUEVOS, B2: el contrato
internal/ingest/copilot.go · copilot_registro_test.go               # NUEVOS, B2: una línea → evento o clase
internal/ingest/copilot_contexto.go · copilot_fichero_test.go       # NUEVOS, B1 (listado) y B3: fichero, prefijo, tramos y recuentos
internal/ingest/testdata/copilot/{generar.sh, esperado.json, *.jsonl, sesiones/**}   # NUEVOS, B2–B3: sintéticos (D-011-P8)
internal/testutil/sandbox.go · sandbox_copilot_test.go              # B1: una línea tras :68 · NUEVO
cmd/permea/copilot.go · copilot_test.go · testdata/copilot/**       # NUEVOS, B4–B5
cmd/permea/main.go                                                  # B4: agent, setup, generate, runOnce, tick · B5: dryRun
README.md · CHANGELOG.md                                            # B6
```

**Existentes que se tocan** *(sobre `d81c39d`)*:
- **`cmd/permea/main.go`**:
  - `agent` gana `copilotRaiz string` y `copilot *ingest.PasadaCopilot`, tras `opencode` *(:186-191)*;
  - `setup()`: `config.CopilotSessionsRoot()` con la regla de Codex *(:230-243)*;
  - `generate()`: el bloque **entre** OpenCode *(:381-390)* y el único `st.Save` *(:392)*;
  - `runOnce()` y `tick()`: la línea, tras la de OpenCode *(:466-467, :535-536)*;
  - `dryRun()`: la detección, **tras** Gemini *(:582-588)*.
- **`internal/testutil/sandbox.go:68`**: una línea, `t.Setenv("COPILOT_HOME", "")`.
- **`internal/config/codex.go:16`, `gemini.go:17` y `opencode.go:18`**: sólo el comentario, «una de las CINCO». Y **`README.md`** y **`CHANGELOG.md`**.

**Nada más**: ni `internal/event/`, `state/`, `project/`, `transport/`, `pricing/` o `sqlitero/`, ni los otros cuatro lectores, ni ningún `_test.go` existente.

## 1 · Una línea *(B2; FR-010 a FR-017, FR-020; D-3, P-1, P-4, P-5, P-11)*

`RegistroCopilot(linea []byte, ctx ContextoCopilot) (*event.Event, ClaseCopilot, error)`, con `ContextoCopilot{Context, SessionID, Cwd}`, que pone B3.
- **Prefiltro** *(P-11)*: `bytes.Contains` de `"type":"session.usage_record"` o `"type":"session.compaction_complete"`. Sin marcador → `NoEsRegistro`,
  **sin decodificar** *(un contador de decodificaciones, para (13))*.
- **Envoltura en crudo**: `type`, `timestamp` y `data` como `json.RawMessage`. Un error sólo si la envoltura no es un objeto JSON *(línea corrupta)*.
  Se comprueba que el `type` decodificado sea el del marcador.
- **`compaction_complete`**: sólo con `data.success == true` *(bool)*; si no, `NoEsRegistro`.
- **Identidad** *(P-1)*: `derivarEventIDCopilot(tipo, id)`; tipo `respuesta` *(`tipoRespuesta`, `codex_eventid.go:29`)* o `compactacion`.
- **Partidas** *(D-3)*: `inputTokens`, `outputTokens`, `cacheReadTokens` y `cacheWriteTokens`, de `usage` o de `compactionTokensUsed`. Una ausente vale 0;
  las cuatro ausentes, «incoherente». `tokens_input = input − lectura − escritura`, y salida = `outputTokens`.
- **Orden** *(D-011-P4)*: sin identificador → incoherente → evento.
- **`occurred_at`** = `timestamp` *(RFC 3339)*, en UTC. **Modelo**: `usage.model` / `compactionTokensUsed.model`; si no es texto no vacío, vacío y «sin modelo».
- **Sin coste**: `tool = "copilot_cli"`, 0 y `false`. `copilotUsage.totalNanoAiu` sólo lo lee `--scan` *(B5)*.
- **Referencias**: `event.Ref(sal, ctx.SessionID)` y `ctx.Resolutor.Derivar(ctx.Cwd, sal)`, como `codex.go:150-151`; vacías si faltan.

## 2 · Un fichero *(B3; FR-004, FR-005, FR-008, FR-020; P-2, P-3)*

`LeerFicheroCopilot(st, ruta, base ContextoCopilot, p *PasadaCopilot, avisos io.Writer) ([]event.Event, error)`, **calcado de `LeerFicheroCodex`**
*(`codex_contexto.go:150-210`)*:
1. **`Size == Offset`** → no se abre. Un contador de aperturas, para (15).
2. **`0 < Offset < Size`** → **prefijo** `[0, Offset)` con `ReadBytes('\n')`. Sólo se decodifican las líneas con `"type":"session.start"` o
   `"type":"session.resume"`. De `start` salen `data.sessionId` y `data.context.cwd`; de `resume`, sólo `data.context.cwd` *(D-011-P2)*.
3. **`st.Recorrer`** desde el offset: un `start`/`resume` actualiza el contexto en orden. Cada registro se clasifica con el contexto **de ese momento**,
   y se deduplica por `event_id` en la pasada *(«repetidas»)*. Una envoltura corrupta → `skip (línea corrupta): …` *(el texto de `main.go:320`)* y sigue.
4. **Truncado** *(`Size < Offset`)*: `Recorrer` relee desde 0 *(`state.go:129-131`)*, sin prefijo. Una **línea final sin LF** no se consume *(`:146-148`)*.

`PasadaCopilot` *(nueva)*: `Registros`, `Eventos`, `Compactaciones`, `Repetidas`, `SinIdentificador`, `Incoherentes` y `SinModelo`; `Resumen()` con el texto
de spec §Textos; `HayNovedades()` = `Registros > 0`; y, sólo para `--scan`, `NanoAIU`.

`ListarCopilot(raiz) ([]string, error)` *(B1)*: `os.ReadDir(raiz)`. Para cada **directorio** que no empiece por `.`, `<dir>/events.jsonl` si es un
fichero regular. En orden léxico. **Ni recursión ni aviso** *(P-7, D-6)*.

## 3 · Ubicación y activación *(B1; D-5, P-6, M-3, M-5)*

- **`config.CopilotSessionsRoot() (string, error)`**: `$COPILOT_HOME/session-state` si no es vacía; si no, `<UserHomeDir>/.copilot/session-state`.
  Devuelve la ruta **aunque no exista**.
- **Activación**: en `generate()`, en **cada** pasada, `os.Stat(raiz)` es un directorio *(como Codex, `main.go:358-359`)*.
- **`sandbox.go`**: tras `:68`, `t.Setenv("COPILOT_HOME", "")`. Hoy no la lee nadie *(`grep`: 0)*.

## 4 · Integración *(B4; FR-021, FR-023 a FR-026)*

`generarCopilot(st, ictx)`, en `cmd/permea/copilot.go`, como `generarCodex` *(`main.go:401-430`)*:
- errores **por fichero**, `copilot_cli: fichero omitido: %v`, con el `%v` **sin la ruta** *(D-011-P5)*;
- encolar es fatal;
- la línea en `runOnce` tras la de OpenCode, y en `tick` con `HayNovedades()`.

## 5 · `--scan` y la línea de control *(B5; P-9, FR-022)*

- **`esSesionCopilot(path)`**: la 1.ª línea, con el tope de 1 MiB de `esSesionCodex` *(`main.go:639`)*, decodificada: `type == "session.start"` y
  `data.copilotVersion` textual. Va **tras** Gemini.
- **`dryRunCopilot`**: `LeerFicheroCopilot` sobre `state.New()`, que nunca se guarda. Escribe `imprimirEventos` *(`main.go:659-671`)*, el resumen y la
  **línea de control**: Σ `totalNanoAiu` de los registros que dieron evento, y su equivalente en USD, **exacto con 11 decimales**, en aritmética entera *(D-011-P6)*.

## 6 · Testdata: SINTÉTICOS *(D-011-P8)*

- **Nunca se copia una línea de una copia congelada**, ni entera ni en parte. Los ficheros se **construyen** desde el esquema del descubrimiento *(desc. Q2,
  m2 M2, m3 S1)*, con valores inventados, por `internal/ingest/testdata/copilot/generar.sh` *(`bash` + `python3 -I`)*, que se versiona con su md5.
- **Textos**: prompts, respuestas, `summaryContent`, `requestMessages` y `system.message` son cadenas neutras *(`texto-1`, `texto-2`…)*.
  - **Ids**: `u-000…`, `r-000…`, `s-000…` *(nunca la forma UUID)*.
  - **Rutas**: `/sintetico/proyecto-a`.
  - **Centinelas**: `CENTINELA-…`.
  - **Modelos**: `modelo-sintetico-1` y, donde importa la tarifa, `gpt-6-luna`.
- **Incluye**: `start`/`resume` con dos `cwd`; dos llamadas; un subagente *(con `agentId`, `sourceSessionId` propio y `parentToolCallId`)*; una compactación
  con éxito y otra fallida; líneas `model.*`, `system.message` y `user.message` con un falso marcador escapado; una línea corrupta; un fichero sin LF final; y
  carpetas sin `events.jsonl` y con `.lock`, `checkpoints/` y `session.db` de 0 B.
- **`esperado.json`**: los recuentos y las sumas que calcula el guion. **Ningún test abre un testdata en su sitio**: siempre una copia en `t.TempDir()`.
- **Vigilancia** *(T-privacidad en cada corte, PARA 8)*: el patrón `PRIV` de `tasks.md` §Puertas, sobre `git diff 15c4c47` y con `grep -raE` sobre los dos
  `testdata/copilot/` *(con `generar.sh`)*. El usuario y el nombre del dueño salen del entorno: nunca se escriben en el repo.

## 7 · Tamaño, SC-015 y puertas

- **SC-007** *(D-011-P7)*:
  - `git archive` de `15c4c47` y de `HEAD`; `CGO_ENABLED=0 -trimpath -ldflags "-s -w -X main.version=0.8.0"` *(0.7.0 en la base: misma longitud)*;
  - las 5 plataformas, Δ ≤ 102 400 B en cada una; en linux/arm64, el Δ de Σ `FileSiz` de los `LOAD` *(`readelf -lW`; 010 E-3)*;
  - informativo en el corte de B4, normativo en C1.
- **SC-015** *(D-011-P9)*: en la Fase 0 de B4, **antes de tocar `main.go`**, el binario de `15c4c47` escribe el stderr de `--run` *(las otras cuatro
  herramientas, en sandbox, sin Copilot)* en `cmd/permea/testdata/copilot/referencia-run.stderr`, con su md5.
- **SC-016**: `git diff 15c4c47` de la frontera de FR-027 y `--diff-filter=M -- '*_test.go'` → vacíos; `os.Getenv` de producción → **5**.

## Decisiones de plan

| ID | Decisión | Por qué |
|---|---|---|
| **D-011-P1** | Sin paquete nuevo: `internal/ingest/copilot*.go` y `internal/config/copilot.go` | El patrón de 008 |
| **D-011-P2** | El contexto de §2: `sessionId` sólo del `start`; `cwd` del último `start`/`resume` | P-3. `session.resume` no trae `sessionId` *(desc. Q2)* |
| **D-011-P3** | Ningún campo nuevo en `state.json` | FR-004. El prefijo se relee *(como Codex)* |
| **D-011-P4** | Clasificación: sin identificador → incoherente → evento; «sin modelo» y «compactación», marcas sobre el evento | FR-020, por construcción |
| **D-011-P5** | Errores por fichero, con el `%v` **sin ruta** *(`PathError` → `Op: Err`)* | FR-026. La ruta lleva el id de la sesión |
| **D-011-P6** | La línea de control, **exacta**: `nanoAIU / 10¹¹` USD con **11 decimales**, en enteros *(`%d.%011d` de `nanoAIU / 10¹¹` y `nanoAIU % 10¹¹`)*, sin truncar ni redondear | Sin flotantes y sin pérdida *(dueño, 14-10)*. W2 compara evento a evento con `ROUND(·, 6)` |
| **D-011-P7** | SC-007 con `git archive` de las dos versiones; linux/arm64, el contenido de los `LOAD` | El instrumento de 010 |
| **D-011-P8** | Los testdata de §6, de `generar.sh`, y nunca de una copia | Spec SC-020; PARA 8 |
| **D-011-P9** | La referencia de SC-015, antes de tocar `main.go` | Como D-010-P10 |
| **D-011-P10** | El forzado de SC-019 es **el A de 009** | Probado en 009 y 010 |

## Bloques

**Fase 0** de cada bloque: las puertas en verde, el censo y el esqueleto que haga compilar los rojos. **Corte**: ✋ commit del dueño. **S** < 50
líneas de producción · **M** 50–150 · **L** > 150.

| Bloque | FR · SC | Rojos | Mutaciones *(→ lo que debe caer)* | |
|---|---|---|---|---|
| **B0** FASE 0 y registro | — | — | — | docs |
| **B1** Raíz y ficheros | FR-001 a FR-003 · SC-012 · M-3, M-5 | **(1)** `COPILOT_HOME` definida, vacía, ausente; inexistente → **devuelve la ruta**, sin error · **(2)** con `COPILOT_HOME=/x` puesta **antes** de `Sandbox(t)`, después `LookupEnv` = `("", true)` · **(3)** `ListarCopilot` sobre `sesiones/`, seis hojas: `solo_events` *(sólo `<dir>/events.jsonl` de un nivel)*, `sin_locks`, `sin_checkpoints_db`, `hondo`, `vacia` *(la carpeta sin `events.jsonl`: sin error y **fuera de la lista**)*, `orden` *(léxico)* | **m1** vacía como definida → (1/vacia) · **m2** sin `session-state` → (1/definida), (1/ausente), (1/inexistente) · **m3** ignorar `COPILOT_HOME` → (1/definida) · **m4** sin la línea del sandbox → (2) · **m5** recursivo → (3/hondo) · **m6** carpeta sin `events.jsonl` = error → **(3) ×6** *(el fixture tiene siempre la carpeta vacía)* | **S** |
| **B2** Una línea e identidad | FR-010 a FR-017, FR-020 · SC-005, SC-009 *(línea)*, SC-011 *(vectores, evento)*, SC-021 | **(4)** los 5 vectores del contrato, **cada hoja con su valor exacto** · **(5)** id vacío → no ok *(**nace verde**: el stub ya da `"", false`; lo valida m9)* · **(6)** SC-005: 100/40/2/7/3 → 58/7/40/2 · **(7)** compactación: `partidas` *(escritura ≠ 0)*; `fallida` *(`success: false` → no registro; **hoja nacida verde**, la valida m14)*; sin `compactionTokensUsed` → incoherente; `sin_requestid` → sin identificador · **(8)** incoherente, siete hojas *(`usage` no objeto · las cuatro ausentes · no numérica · no entera · negativa · `no_cabe` · `timestamp` malo)* · **(9)** sin identificador: ausente, vacío, no textual · **(10)** `tool`, coste 0, `false`; `totalNanoAiu` no entra; `razonamiento` *(`reasoningTokens` 999 no cambia la salida)* · **(11)** `occurred_at` en UTC, con una hoja `desfase` *(`…+02:00`)* · **(12)** sin modelo: emitido y marcado · **(13)** prefiltro: el falso marcador escapado y las `model.*` → 0 decodificaciones *(**nace verde**: el stub no decodifica; lo valida m19)* · **(14)** centinelas fuera del evento: **rojo antes del verde** | **m7** `"copilot_cli"` → `"codex"` → (4) ×5 · **m8** sin prefijo de longitud → (4) ×5 · **m9** vacío aceptado → (5), **(9) ×3, (7/sin_requestid)** · **m10** `compactacion` → `respuesta` → (4/compactacion), (4/mismo_valor) · **m11** salida + razonamiento → (6), **(10/razonamiento)** · **m12** sin restar la escritura → (6), **(7/partidas)** · **m13** escritura a lectura → (6), **(7/partidas)** · **m14** `success: false` emitida → (7/fallida) · **m15** `cr + cw > input` aceptado → (8/no_cabe) · **m16** `cost_available = true` → (10) · **m17** `occurred_at` sin `.UTC()` → **(11/desfase)** · **m18** sin modelo no se emite → (12) · **m19** sin prefiltro → (13) · **m20** `session_ref` sin sal → (14) | **M** |
| **B3** Fichero, frontera y tramos | FR-004, FR-005, FR-007 a FR-009, FR-020 · SC-003 *(función)*, SC-004, SC-006 *(función)*, SC-010, SC-013, SC-014, SC-021 | Cada pasada afirma su **línea de resumen literal**. **(15)** dos llamadas → 2; la 2.ª pasada, 0 y **0 aperturas**; `linea_larga` *(un `system.message` de ≥ 64 KB antes de los registros; FR-009)* · **(16)** tramos: `start` X, registro, `resume` Y, registro → dos `project_ref`, `una_sesion`; y `valores` = `event.Ref(sal, "s-…")` y `Resolutor.Derivar("/sintetico/x", sal)` *(SC-021)* · **(17)** `cortado/` *(`start` X + registro + `system.message` de 64 KB | `resume` Y + registro)*, en dos pasadas = de una vez · **(18)** una hoja: sin LF → 0; con LF → 1 · **(19)** truncado → mismos `event_id` · **(20)** el mismo `usageId` dos veces → `repetidas 1` · **(21)** `sin_start` → `session_ref` vacío; sin `cwd` → `project_ref` vacío · **(22)** línea corrupta → aviso y sigue · **(23)** SC-010, literal · **(24)** subagente: `sesion` *(el del `start`)* y `modelo` *(`usage.model`, nunca el de `subagent.*`)* · **(25)** sesión matada: 1 evento · **(26)** md5 y listado de la carpeta, y los `.lock`, iguales tras leer *(**nace verde**: el stub no escribe; lo valida m29)*. En `generar.sh`, `sourceSessionId` = `sessionId` en `tramos.jsonl` y `cortado/` | **m21** `cwd` sólo del `start` → (16), **(17)** · **m22** sin prefijo → (17) · **m23** `session_ref` de `sourceSessionId` → (24/sesion), **(21/sin_start)** · **m24** modelo de `subagent.configured` → (24/modelo) · **m25** sin dedupe → (20), (23) · **m26** la corrupta corta el fichero → (22) · **m27** repetida contada como evento → **(20)**, (23) · **m28** sin el atajo `Size == Offset` → (15/aperturas) · **m29** el lector crea un `.lock` → (26) · **M-B3a** `session_ref` = `event.Ref(sal, cwd)` → (16/valores), (16/una_sesion) | **M** |
| **B4** Integración | FR-002, FR-021, FR-023 a FR-026 · SC-003, SC-006, SC-011, SC-015, SC-018, SC-019 | **(27)** SC-015 contra la referencia, `copilot_home_vacia` y `sin_raiz` *(nace verde)* · **(28)** sólo Copilot CLI · **(29)** forzado A · **(30)** la línea, tras OpenCode, el total y **`orden_cola`** *(los eventos de OpenCode antes que los de Copilot)* · **(31)** `tick` · **(32)** SC-003 · **nacen verdes, tras el verde**: **(33)** SC-011, cola y `state.json` · **(34)** SC-006, md5 y listado · **(35)** carpeta creada después, mismo `agent` · **(36)** SC-018 | **m30** `a.copilot = NuevaPasadaCopilot()` siempre → (27) ×2, (35) y los `*_SinRaiz*`/`*_LineaDeResumen` de Codex, Gemini y OpenCode *(censo por nombre en la tarea)* · **m31** `st.Save` antes de Copilot → (29) · **m32** `tick` sin predicado → (31) · **m33** Copilot antes que OpenCode → (30/orden_cola) · **m34** existencia sólo en `setup()` → (35) · **m35** el error aborta la pasada → (36/pasada) · **m36** el error avanza el offset → (36/se_relee) · **m37** la ruta del fichero en el evento → (33) · **M-B4a** `generarCopilot` crea un fichero bajo la raíz → (34) | **M** |
| **B5** `--scan` | FR-022, FR-023 *(scan)* · SC-001, SC-002 *(forma)*, SC-023 *(forma)* | **(37)** fichero → `evento:`, resumen, control y nada en disco · **(38)** `claude`, `codex`, `gemini` y `opencode` como en la 0.7.0 *(nace verde)* · **(39)** control: Σ con la compactación, sin los incoherentes; `exacto` *(11 decimales)*; `no_multiplo` *(un `nanoAIU` que no es múltiplo de 100)* · **(40)** 1.ª línea `session.start` sin `copilotVersion` → no es Copilot *(**nace verde**: sin `esSesionCopilot` ya no lo es; lo valida m42)* | **m38** todo es Copilot → (38) ×4 y los `TestScan_*`/`*Scan_*` existentes *(censo por nombre)* · **m39** `dryRunCopilot` guarda estado → (37/disco) · **m40** control sin la compactación → (39) · **m41** el equivalente con `%.9f` sobre `float64` → (39/exacto), (39/no_multiplo) · **m42** detección sin `copilotVersion` → (40) | **M** |
| **B6** README y CHANGELOG | FR-028 · P-8, P-10 | `grep -c '^## 0.8.0'` = 0 · `grep -c '^### Copilot CLI'` = 0 | — *(textos aprobados ✋; `cmp`)* | **S** |

**44 mutaciones**: m1–m42, M-B3a y M-B4a. **40 tests numerados: 30 rojos y 10 nacidos verdes** *((5), (13), (26), (27), (33)–(36), (38), (40))*.

**Estimación**:

| | B1 | B2 | B3 | B4 | B5 | B6 |
|---|---:|---:|---:|---:|---:|---:|
| Producción | ~45 | ~140 | ~170 | ~90 | ~70 | ~50 de texto |
| Test | ~90 | ~280 | ~330 | ~300 | ~130 | — |

**~515 de producción y ~1 130 de test.** El lector de Codex tiene 435 *(`codex.go` 157 + `codex_contexto.go` 278)*. El riesgo de SC-007 es B2 + B3.

## Cierre — en tramos, uno por mensaje *(si uno falla, se para y se rehace desde C1)*

| Tramo | Qué | Quién |
|---|---|---|
| **C1** · Puertas | gofmt, vet y lint 0; 848 + nuevos, 0 SKIP; SC-016; **SC-007** en las 5; SC-008; SC-020 *(privacidad y testdata, §6)*; compilan Windows y darwin | Claude |
| **C2** · Medidas | Cada suite, tras `~/bin/libre.sh`. Sobre **las 3 copias congeladas**, en su sitio, con sus sha256 antes y después *(PARA 9)*:<br>• **SC-001** y **SC-002**, en `-14b` *(`--run` con `COPILOT_HOME` = la copia, y `--scan` de F)*;<br>• **SC-003**;<br>• **SC-009** sobre E;<br>• **SC-004** sobre D;<br>• **SC-023** sobre G *(`-14c`)*;<br>• en `-14`, A+B y C *(= `-14b`)*.<br>Contraste independiente: un guion `python3 -I` fuera del repo cuenta registros, partidas y créditos, y debe dar lo mismo que el agente | Claude |
| **C3** · Snapshot | `goreleaser release --snapshot --clean`, `strings` con los textos *(SC-017)* | Claude |
| **C4** · **W1** | En Windows, en sandbox *(`COPILOT_HOME` propio)* y sin enrolar *(D-011-W)* | ✋ dueño |
| **C5** · PR | El cuerpo, fuera del repo | Claude redacta |
| **C6** · Fusión | Merge commit | ✋ dueño |
| **C7** · Etiqueta | `v0.8.0` | ✋ dueño |
| **C8–C9** | Canales y **W2** *(SC-022, en la instalación real, con el control de créditos)* | tras C7 |

**D-011-W · W1**: el contador independiente es un guion `python3 -I` que lee sólo `usage_record` y `compaction_complete` *(no comparte código con el lector)*.
1. **Copilot cerrado**: md5 y listado → `--run` → **iguales** *(SC-006)*. Eventos = registros del contador, y Σ partidas iguales. Σ créditos × 0,01 de `--scan` =
   Σ eventos × la tarifa *(SC-022)*. Un segundo `--run` → 0.
2. **Copilot abierto**: dos prompts y un `/compact`, con `--daemon` *(60 s)*. Al cerrar, Σ eventos = registros nuevos, una vez cada uno; sin ficheros nuevos.

**W2**: tras C8, `scoop update` y `--run`. En la plataforma, **evento a evento**: `ROUND(créditos × 0,01, 6)` = el coste del evento *(la plataforma redondea por evento)*.

## Trazabilidad *(FR → bloque)*

| FR | Bloque | | FR | Bloque | | FR | Bloque | | FR | Bloque |
|---|---|---|---|---|---|---|---|---|---|---|
| 001 | B1 | | 008 | B3 | | 015 | B2 | | 022 | B5 |
| 002 | B1 · B4 | | 009 | B3 | | 016 | B2 | | 023 | B4 · B5 |
| 003 | B1 | | 010 | B2 | | 017 | B2 | | 024 | B4 |
| 004 | B3 | | 011 | B2 | | 018 | B2 · B3 | | 025 | B4 |
| 005 | B3 | | 012 | B2 | | 019 | B2 · B3 · B4 | | 026 | B4 |
| 006 | B3 · B4 | | 013 | B2 | | 020 | B2 · B3 | | 027 | todos |
| 007 | B2 · B3 | | 014 | B2 | | 021 | B4 | | 028 | B6 · C7 |

**28 / 28 FR con bloque.** Los 23 SC, en `tasks.md` §Cobertura.

## Riesgos

| # | Riesgo | Medida o mitigación |
|---|---|---|
| R-1 | **SC-007**: el código del lector *(~515 líneas, estimado en 30–50 KB sin medir)*. Un salto de página Mach-O en darwin/amd64 *(`__TEXT`, 846 B de holgura hoy)* cuesta como mucho 4 096 B del tope de 102 400 | Sin paquetes nuevos. Medidas informativas **tras B2**, en el corte de B3 y en el de B4 *(5 `go build`)*. Si no cabe, se adelgaza; el tope no sube |
| R-2 | Un test hereda las sesiones reales del desarrollador | M-3, y que los `agent` hechos a mano no las lean |
| R-3 | **Un dato real en testdata** | §6 y PARA 8, con su tarea de vigilancia en cada corte |
| R-4 | Releer el prefijo de un fichero grande *(838 KB en E)* en cada pasada en que crece | Prefiltro de bytes: sólo se decodifican `start` y `resume`. Lo mismo que Codex |
| R-5 | Los forzados de SC-018 y SC-019 dependen de permisos POSIX | `t.Skip` con root o en Windows, como 009 R-7 |
| R-6 | Copilot cambia de forma *(una versión nueva)* | Se cuenta como incoherente o como no-registro, nunca se rompe. C2 rehecho sobre una copia nueva |

## Complexity Tracking

*Vacío.* Ni paquetes ni campos nuevos.
