# Implementation Plan: 012 · «Lector de Kimi Code CLI»

**Branch**: `012-lector-kimi` | **Date**: 2026-10-15 | **Spec**: [spec.md](./spec.md) *(ratificada el 2026-10-15; E-1)* | **Base**: el commit de la spec ratificada
*(el código, el de `e9e5449`, 0.8.0)*
**Input**: la spec · [`contracts/event-id-kimi.md`](./contracts/event-id-kimi.md) · [`soporte/descubrimiento.md`](./soporte/descubrimiento.md) y las muestras 1 y 2 ·
el plan y las tareas de 011 *(el lector de Copilot CLI, que es el modelo)*

## Summary

La `0.9.0` lee también Kimi Code. Cada **`usage.record`** de `<raíz>/sessions/<wd>/<S>/agents/<A>/wire.jsonl` es un evento `tool = kimi_code`, con las
partidas **tal cual** *(FR-013)* y sin coste *(D-5)*. Lo heredado de un `/fork` no se emite *(D-3)*. El evento no cambia *(M-1)*, y los otros cinco lectores se
leen como en la 0.8.0 *(M-2)*. Es **el lector de Copilot CLI con otra identidad**:
- offset por fichero con `state.Recorrer`, **más una huella de la cabeza** en `state.json` *(E-1; el único campo nuevo)*;
- un **prepaso** del fichero entero con prefiltro, para el ordinal `N` y la última `forked` *(D-3, P-5)*;
- el proyecto, de `session_index.jsonl` *(P-6)*; recuentos y línea de resumen.

Riesgos:
1. **La posición como identidad** *(D-3)*: N debe ser el mismo en una pasada o en varias, tras una relectura y en `--scan`. Lo cubren (15)–(19) y el contrato.
2. **Privacidad**: las líneas llevan prompts, respuestas y rutas. El prefiltro está **anclado al principio de la línea** *(P-10)*, y los testdata son **sintéticos** *(§6)*.
3. **≤ 100 KB** frente a la 0.8.0 *(SC-022)*: ~600 líneas de producción *(R-1)*.
4. **`state.json` gana un campo** *(E-1)*: compatible en los dos sentidos con la 0.8.0 *(§2, D-012-P3)*.

## Technical Context

- **Go 1.22.2** · `golangci-lint` 2.12.2 · `goreleaser` v2. **Ninguna dependencia** *(SC-023)*. Sólo `encoding/json`, `crypto/sha256` y `bytes`, ya enlazados.
- **Línea base** *(`e9e5449`; se transcribe en T001)*: la del cierre de 011, **953** «--- PASS», 0 FAIL y 0 SKIP *(011 registro §Cierre, T041)*; vet limpio; lint → 0.
  Los **tamaños de la 0.8.0** en las 5 plataformas, para SC-022, se miden en T001 con el instrumento de 011 *(D-011-P7)*.
- **Storage**: una entrada de `state.json` por `wire.jsonl`, con los cuatro campos *(`internal/state/state.go:16-20`)* y **`cabeza`** *(nuevo, `omitempty`)*.
- **Para regenerar testdata** *(nunca en el binario ni en la suite)*: `bash` y `python3 -I`.

## Constitution Check *(1.1.0; antes y después del diseño: sin cambios)*

| Principio | | Razón |
|---|:--:|---|
| I · Frontera inviolable | ✅ | El evento no cambia. S y A sólo entran en el hash; S y `workDir`, sólo con sal *(FR-019)* |
| II · Local-first | ✅ | **§II, 1.1.0**: sin coste fiable de la herramienta, el evento va sin coste y lo calcula la plataforma *(D-5)*. Se lee un JSONL local, en sólo lectura. `cabeza` lleva sal |
| III · Binario único | ✅ | Sin dependencias y sin cgo. Ningún paquete nuevo |
| IV · Test-first en la frontera | ✅ | La frontera no se toca. SC-011 siembra centinelas, y (13) es **rojo antes del verde** de B2 |
| Fixtures | ✅ | JSONL **sintéticos** de `generar.sh`. Nunca una línea de una copia congelada *(§6)* |
| V · Specs | ✅ | Spec ratificada y contrato antes del código |

## Project Structure

```text
internal/config/kimi.go · kimi_test.go                              # NUEVOS, B1 (P-1)
internal/ingest/kimi_eventid.go · kimi_eventid_test.go              # NUEVOS, B2: el contrato
internal/ingest/kimi.go · kimi_registro_test.go                     # NUEVOS, B2: una línea → evento o clase
internal/ingest/kimi_contexto.go · kimi_fichero_test.go             # NUEVOS, B1 (listado) y B3: índice, prepaso, cabeza, fichero, recuentos
internal/ingest/testdata/kimi/{generar.sh, esperado.json, lineas/**, ficheros/**, raiz/**}   # NUEVOS, B1–B3: sintéticos (D-012-P8)
internal/state/state.go                                             # B3: el campo `Cabeza`, tras `Frontera` (:21-23)
internal/testutil/sandbox.go · sandbox_kimi_test.go                 # B1: una línea tras :70 · NUEVO
cmd/permea/kimi.go · kimi_test.go · testdata/kimi/**                # NUEVOS, B4–B5
cmd/permea/main.go                                                  # B4: agent, setup, generate, runOnce, tick · B5: dryRun
README.md · CHANGELOG.md                                            # B6
specs/012-lector-kimi/soporte/registro.md · registro-archivo.md     # el registro (techo 300) y su archivo: §B0–§B2, íntegros, sin techo
```

**Existentes que se tocan** *(sobre el commit de la spec ratificada)*:
- **`cmd/permea/main.go`**:
  - `agent` gana `kimiRaiz string` y `kimi *ingest.PasadaKimi`, tras `copilot` *(:192-196)*;
  - `setup()`: `config.KimiHome()` con la regla de Copilot *(:250-254, :263)*;
  - `generate()`: el bloque **entre** Copilot *(:403-412)* y el único `st.Save` *(:414)*;
  - `runOnce()` y `tick()`: la línea, tras la de Copilot *(:492, :563-565)*;
  - `dryRun()`: la detección, **tras** Copilot *(:618-624)*.
- **`internal/state/state.go`**: `Cabeza string \`json:"cabeza,omitempty"\`` en `FileState`. **Nada más**: `Recorrer`, `Load` y `Save` no cambian *(D-012-P3)*.
- **`internal/testutil/sandbox.go:70`**: una línea más, `t.Setenv("KIMI_CODE_HOME", "")`.
- **`internal/config/codex.go:16`, `gemini.go:17`, `opencode.go:18` y `copilot.go:16`**: sólo el comentario, «una de las SEIS». Y **`README.md`** y **`CHANGELOG.md`**.

**Nada más**: ni `internal/event/`, `project/`, `transport/` o `pricing/`, ni los otros cinco lectores, ni ningún `_test.go` existente.

## 1 · Una línea *(B2; FR-010 a FR-017, FR-020; D-3, D-4, P-5, P-7 a P-10)*

`RegistroKimi(linea []byte, ctx ContextoKimi) (*event.Event, ClaseKimi, error)`, con `ContextoKimi{Context, S, A, N, Heredado, Cwd}`, que pone B3.
- **Orden** *(FR-020)*: `Heredado` → `Heredado`, **sin decodificar**. Si no, prefiltro y decodificación.
- **Prefiltro** *(P-10)*: `bytes.HasPrefix(linea, {"type":"usage.record")`. Sin él → `NoEsRegistro`, **sin decodificar** *(un contador, para (12))*. Error sólo si
  la línea no es un objeto JSON *(corrupta)*; se comprueba que `type` sea `usage.record`.
- **Campos** *(FR-012)*: `model`, `usage.{inputOther, output, inputCacheRead, inputCacheCreation}`, `usageScope` y `time`, en un struct con
  `json.RawMessage` donde hace falta. **Nunca** `agentId`.
- **Incoherente**: `usage` no objeto; las cuatro ausentes; una partida no numérica, no entera o negativa; `time` ausente, no entero o ≤ 0. → **a cero** *(las
  cuatro = 0)* → evento. La deduplicación *(«repetidas»)* es de B3.
- **Identidad** *(P-5)*: `derivarEventIDKimi(S, A, N)`, con `dominioEventID` y `bytesEventID` *(`eventid.go:31`, `:43`)*, `tipoRespuesta` *(`codex_eventid.go:29`)* y la
  codificación **replicada** de `derivarEventIDCopilot` *(`copilot_eventid.go:33-52`)*, como 009–011. N, con `strconv.Itoa`.
- **Partidas**: tal cual *(FR-013)*. **Modelo** *(P-9)*: `strings.LastIndexByte(m, '/')`; lo que sigue, o `m` sin barra; vacío → «sin modelo».
- **`occurred_at`** = `time.UnixMilli(t).UTC()` *(P-7)*. **Compactación**: `usageScope == "session"`, una marca sobre el evento.
- **Sin coste**: `tool = "kimi_code"`, 0 y `false`. **Referencias**: `event.Ref(sal, S)` y `ctx.Resolutor.Derivar(ctx.Cwd, sal)`, como `copilot.go` de 011; vacías si faltan.

## 2 · Un fichero *(B3; FR-004, FR-005, FR-008, FR-014, FR-020; D-3, P-4 con E-1, P-6)*

`LeerFicheroKimi(st, ruta, base ContextoKimi, p *PasadaKimi, avisos io.Writer) ([]event.Event, error)`, **calcado de `LeerFicheroCopilot`**
*(`internal/ingest/copilot_contexto.go:127-192`)*:
1. **Atajo**: `Size == Offset` **y** `ModTime == mtime` → no se abre *(un contador de aperturas, para (14) y (19))*.
2. **Cabeza** *(E-1)*: se lee la 1.ª línea completa y se calcula `huellaCabezaKimi(sal, L)` = `event.Ref(sal, L sin LF ni CR)`. Si el fichero **encoge**, si la
   huella ≠ `Cabeza` o si `Cabeza == ""` con `Offset > 0` → **relectura**: la entrada se pone a cero **en memoria**, y se restaura si hay error *(FR-026)*.
3. **Prepaso** *(D-3, P-5)*: el fichero entero, con `ReadBytes('\n')`, **sin decodificar** salvo lo que pase el prefiltro de `forked` *(confirmado por JSON)*.
   Da `inicio[k]`, el comienzo del registro k, y `ultimaForked`, el comienzo de la última `forked`. Sólo líneas completas.
4. **`st.Recorrer`** desde el offset *(`state.go:121-169`)*: para cada línea con el prefijo de registro, `N` = 1 + los registros que empiezan antes de ella *(una
   búsqueda en `inicio`)*, y `Heredado` = `comienzo < ultimaForked`. Se deduplica por `event_id` *(«repetidas»)*. Una corrupta → `skip (línea corrupta): …`
   *(`copilot_contexto.go:149`)* y sigue: **cuenta para N**, pero no como registro.
5. **Tras `Recorrer`**: `Cabeza` se escribe en la entrada *(`Recorrer` la rehace sin ella, `state.go:162-167`)*. Si el fichero aún no tiene línea completa, vacía.

**Índice** *(P-6, P-13)*: `LeerIndiceKimi(raiz) (map[string]string, error)` lee `session_index.jsonl` entero, **una vez por pasada**. Decodifica `sessionId` y `workDir`
de cada línea; gana la **última**. `os.IsNotExist` → mapa vacío y `nil`. Otro error → se devuelve, y B4 no lee ningún `wire.jsonl` *(P-13)*. Antes de usar W, una
ruta con forma de Windows *(`^[A-Za-z]:[\\/]` o `^\\\\`)* pasa sus `/` a `\` *(`normalizarWorkDirKimi`)*.

`PasadaKimi` *(nueva)*: `Registros`, `Eventos`, `Heredados`, `Repetidas`, `ACero`, `Incoherentes`, `Compactaciones` y `SinModelo`; `Resumen()` con el texto de
spec §Textos; `HayNovedades()` = `Registros > 0`.

`ListarKimi(raiz) ([]string, error)` *(B1)*: `<raiz>/sessions/<a>/<b>/agents/<c>/wire.jsonl`, con `os.ReadDir` en cada nivel, **sin** entradas que empiecen
por `.`, y `wire.jsonl` regular. Orden léxico. Como `ListarCopilot` *(`copilot_contexto.go:29-45`)*, con dos niveles más.

## 3 · Ubicación y activación *(B1; P-1, P-2, M-3, M-5)*

- **`config.KimiHome() (string, error)`**: `$KIMI_CODE_HOME` si no es vacía; si no, `<UserHomeDir>/.kimi-code`. Devuelve la ruta **aunque no exista**, como
  `CopilotSessionsRoot` *(`internal/config/copilot.go:18-27`)*.
- **Activación**: en `generate()`, en **cada** pasada, `os.Stat(<raíz>/sessions)` es un directorio *(como `copilotActivo`, `cmd/permea/copilot.go:21-27`)*.
- **`sandbox.go`**: tras `:70`, `t.Setenv("KIMI_CODE_HOME", "")`. Hoy no la lee nadie *(`grep`: 0)*.

## 4 · Integración *(B4; FR-021, FR-023 a FR-026; P-13)*

`generarKimi(st, ictx)`, en `cmd/permea/kimi.go`, como `generarCopilot` *(`cmd/permea/copilot.go:33-57`)*:
- primero el índice: si es ilegible, `kimi_code: fichero omitido: %v` **sin ruta** *(`sinRuta`, `copilot.go:60-66`)*, y no se lee nada más en esa pasada;
- errores **por fichero**, con el mismo texto; encolar es fatal;
- la línea en `runOnce` tras la de Copilot, y en `tick` con `HayNovedades()`.

## 5 · `--scan` *(B5; P-12, FR-022)*

- **`esSesionKimi(path)`**: la 1.ª línea, con el tope de 1 MiB de `esSesionCopilot` *(`cmd/permea/copilot.go:75-96`)*, decodificada: `type == "metadata"`,
  `protocol_version` textual y `created_at` numérico. Va **tras** Copilot.
- **`dryRunKimi`**: S = el nombre de la carpeta tres niveles arriba y A = el de la carpeta que contiene el fichero; `LeerFicheroKimi` sobre `state.New()`, que nunca se
  guarda, sin índice. Escribe `imprimirEventos` *(`main.go:697`)* y el resumen. Sin línea de control *(D-5)*.

## 6 · Testdata: SINTÉTICOS *(D-012-P8)*

- **Nunca se copia una línea de una copia congelada**, ni entera ni en parte. Los ficheros se **construyen** desde el esquema *(desc. Q2; m1 1)*, con valores
  inventados, por `internal/ingest/testdata/kimi/generar.sh` *(`bash` + `python3 -I`)*, que se versiona con su md5.
- **Textos**: cadenas neutras *(`texto-1`…)*. **Ids**: `session_sintetica-0001`… *(nunca la forma UUID)*. **Rutas**: `/sintetico/proyecto-a` y `C:\sintetico\proyecto-a`
  con las dos barras. **Centinelas**: `CENTINELA-…`. **Modelos**: `proveedor-sintetico/modelo-sintetico-1` y, donde importa, `moonshot-ai/kimi-k2.7-code`.
- **Incluye**: `raiz/` *(sesiones, subagente, `.index-*`, `.bak`, `.tmp.*`, `imported/`, un `wire.jsonl` demasiado hondo)*; `lineas/` *(una por clase)*;
  `ficheros/` *(fork, dos `forked`, cortado, sin LF, encogido, cabeza cambiada, corrupta, matada, subagente con `subagent.completed`)*; un `session_index.jsonl` con las dos barras.
- **`esperado.json`**: los recuentos y las sumas que calcula el guion. **Ningún test abre un testdata en su sitio**: siempre una copia en `t.TempDir()`.
- **Vigilancia** *(PARA 8)*: el patrón `PRIV` de `tasks.md` §Puertas sobre `git diff e9e5449` y, con `grep -raE`, sobre los dos `testdata/kimi/`.

## 7 · Tamaño, SC-015 y puertas

- **SC-022**: `git archive` de `e9e5449` y de `HEAD`, con `CGO_ENABLED=0 -trimpath -ldflags "-s -w -X main.version=0.9.0"` *(0.8.0 en la base)*; Δ ≤ 102 400 B
  en las 5 plataformas; en linux/arm64, Σ `FileSiz` de los `LOAD`. Informativo tras B2, B3 y B4; normativo en C1.
- **SC-015** *(D-012-P9)*: en la Fase 0 de B4, **antes de tocar `main.go`**, el binario de `e9e5449` escribe el stderr de `--run` *(los otros cinco, en sandbox, sin
  Kimi)* en `cmd/permea/testdata/kimi/referencia-run.stderr`, con su md5.
- **SC-016**: `git diff e9e5449` de FR-027 y `--diff-filter=M -- '*_test.go'` → vacíos; `os.Getenv` de producción → **6**.

## Decisiones de plan

| ID | Decisión | Por qué |
|---|---|---|
| **D-012-P1** `DECIDÍ YO` | Sin paquete nuevo: `internal/ingest/kimi*.go` y `internal/config/kimi.go` | El patrón de 008–011 |
| **D-012-P2** `DECIDÍ YO` | El prepaso del fichero entero da N y la última `forked`; `Recorrer` lee lo nuevo | FR-008: N y «heredado» no dependen de dónde corte la pasada |
| **D-012-P3** `DECIDÍ YO` | **`Cabeza`**, en `FileState` tras `Frontera`, `json:"cabeza,omitempty"`: 64 hex de `event.Ref(sal, L)`. **De la 0.8.0 a la 0.9.0**: ninguna entrada de Kimi existe, y las demás no cambian. **De vuelta a la 0.8.0**: el campo se pierde al guardar, y la 0.9.0 relee desde 0 *(«`Cabeza` vacía con `Offset > 0`»)*, con los mismos `event_id` | E-1. Sal, como `Frontera.Huella` *(`state.go:26-29`)*. `Recorrer` no cambia: la conducta de los otros cinco lectores es la de la 0.8.0 |
| **D-012-P4** `DECIDÍ YO` | Clasificación: heredado → incoherente → a cero → repetida → evento; «compactación» y «sin modelo», marcas sobre el evento | FR-020, por construcción |
| **D-012-P5** | Errores por fichero, con el `%v` **sin ruta** | M-4 |
| **D-012-P6** `DECIDÍ YO` | El índice, una vez por pasada y entero; la última línea gana; la normalización de barras, **antes** de `Derivar` | P-6, P-13 |
| **D-012-P7** | SC-022 con `git archive` de las dos versiones; en linux/arm64, los `LOAD` | El instrumento de 010 y 011 |
| **D-012-P8** | Los testdata de §6, de `generar.sh`, y nunca de una copia | SC-020; PARA 8 |
| **D-012-P9** | La referencia de SC-015, antes de tocar `main.go` | Como D-011-P9 |
| **D-012-P10** | El forzado de SC-019 es **el A de 009** | Probado en 009–011 |
| **D-012-W** `DECIDÍ YO` | El cuadre de SC-025 lo hace un guion `python3 -I` **fuera del repo**, que lee la exportación de «Request Details» y los eventos de la cola del sandbox *(W1)* o de la plataforma *(W2)* | §Cierre |

## Bloques

**Fase 0** de cada bloque: las puertas en verde, el censo y el esqueleto que haga compilar los rojos. **Corte**: ✋ commit del dueño. **S** < 50 líneas de
producción · **M** 50–150 · **L** > 150.

| Bloque | FR · SC | Rojos | Mutaciones *(→ lo que debe caer)* | |
|---|---|---|---|---|
| **B0** FASE 0 y registro | — | — | — | docs |
| **B1** Raíz y ficheros | FR-001 a FR-003 · SC-012 · M-3, M-5 | **(1)** `KimiHome`: `definida`, `vacia`, `ausente`, `inexistente` *(devuelve la ruta, sin error)* · **(2)** con `KIMI_CODE_HOME=/x` puesta **antes** de `Sandbox(t)`, después `LookupEnv` = `("", true)` · **(3)** `ListarKimi` sobre `raiz/`, siete hojas: `forma`, `sin_punto` *(`.index-cache`, `.index-dirty`)*, `sin_bak_tmp`, `sin_imported`, `hondo`, `vacia` *(sesión sin `agents/`: sin error)*, `orden` | **m1** vacía como definida → (1/vacia) · **m2** sin `.kimi-code` → (1/ausente), (1/inexistente) · **m3** ignorar la variable → (1/definida) · **m4** sin la línea del sandbox → (2) · **m5** sin el filtro de `.` → (3/sin_punto) · **m6** `HasPrefix("wire.jsonl")` → (3/sin_bak_tmp) · **m7** listar desde `<raíz>` y no desde `sessions` → (3/sin_imported) · **m8** recursivo → (3/hondo) | **S** |
| **B2** Una línea e identidad | FR-010 a FR-017, FR-020 · SC-005, SC-007 *(vectores)*, SC-008, SC-011 *(evento)*, SC-021 | **(4)** los 7 vectores del contrato, **cada hoja con su valor** · **(5)** SC-005: 100/40/2/7 → 100/40/2/7 · **(6)** SC-008, seis hojas · **(7)** incoherente, siete hojas *(`usage` no objeto · cuatro ausentes · no numérica · no entera · negativa · `time` textual · `time` ausente)* · **(8)** a cero · **(9)** heredado, sin decodificar · **(10)** `tool`, coste 0, `false`; `compactacion` *(`usageScope: session`)* · **(11)** `occurred_at`: `time` 1 → `…00.001Z`, en UTC · **(12)** prefiltro: `usage.record` dentro de un texto, `forked` y `context.append_message` → 0 decodificaciones *(**nace verde**: el stub no decodifica; lo valida m22)* · **(13)** centinelas fuera del evento: **rojo antes del verde** | **m9** `"kimi_code"` → `"codex"` → (4) ×7 · **m10** sin prefijo de longitud → (4/frontera) ×2 · **m11** S y A cambiados de orden → (4) ×6 · **m12** N con un cero delante → (4) ×7 · **m13** sumar la caché a la entrada → (5) · **m14** lectura ↔ escritura → (5) · **m15** tras la **primera** `/` → (6/varias) · **m16** vacío → alias entero → (6/acabada) · **m17** negativa aceptada → (7/negativa) · **m18** a cero emitido → (8) · **m19** heredado emitido → (9) · **m20** `cost_available = true` → (10/coste) · **m21** sin `.UTC()` → (11) · **m22** sin prefiltro → (12) · **m23** `session_ref` sin sal → (13) · **m24** compactación sin marca → (10/compactacion) | **M** |
| **B3** Fichero, cabeza, índice | FR-004, FR-005, FR-007 a FR-009, FR-014, FR-018, FR-020 · SC-004, SC-006, SC-007 *(relectura)*, SC-009, SC-010, SC-014, SC-026 | Cada pasada afirma su **línea de resumen literal**. **(14)** dos registros → 2; la 2.ª pasada, 0 y **0 aperturas** · **(15)** fork: `un_forked` *(2 + `forked` + 1 → 1 y heredados 2)*, `dos_forked`, `sin_forked` · **(16)** `fork_en_dos_pasadas`, cortado tras el `forked` = de una vez · **(17)** N: `corrupta_cuenta` *(una corrupta con prefijo entre dos registros: el 2.º es N = 3)* y `dos_pasadas` · **(18)** SC-026: `cabeza_cambiada`, `encogido`, `sin_cabeza` *(0 nuevos, mismos `event_id`)* y `nuevo_tras_cambio` *(1)* · **(19)** `mtime`: mismo tamaño y otro `mtime` → 1 apertura y 0 eventos; mismo tamaño y `mtime` → 0 aperturas · **(20)** sin LF → 0; con LF → 1 · **(21)** índice: `barras`, `ultima`, `sin_linea`, `ausente` · **(22)** índice ilegible → error · **(23)** la misma sesión en otra `wd` → «repetidas 1» · **(24)** SC-010 literal · **(25)** subagente: 3 eventos, mismo `session_ref`, 0 de `subagent.completed` · **(26)** corrupta → aviso y sigue · **(27)** md5 y listado de la raíz iguales tras leer *(**nace verde**; lo valida m39)* · **(28)** `cabeza`: los 2 vectores del contrato · **(29)** `state_compatible`: una entrada sin `Cabeza` se serializa como en la 0.8.0 | **m25** sin frontera → (15) ×3, (24), (25) · **m26** la **primera** `forked` → (15/dos_forked) · **m27** N sin contar los heredados → (15/un_forked), (16) · **m28** N por número de línea → (17), (18/cabeza_cambiada) · **m29** sin prepaso *(N desde el offset)* → (16), (17/dos_pasadas) · **m30** sin comparar la cabeza → (18/cabeza_cambiada) · **m31** cabeza sin sal → (28) · **m32** `Cabeza` vacía tolerada → (18/sin_cabeza) · **m33** sin el atajo → (14/aperturas), (19/igual) · **m34** el atajo sin `mtime` → (19/mtime) · **m35** sin normalizar las barras → (21/barras) · **m36** la **primera** línea del índice → (21/ultima) · **m37** sin dedupe → (23), (24) · **m38** la corrupta corta el fichero → (26) · **m39** el lector crea un fichero → (27) · **m40** `omitempty` quitado → (29) | **L** |
| **B4** Integración | FR-002, FR-021, FR-023 a FR-026 · SC-011, SC-013, SC-015, SC-018, SC-019 | **(30)** SC-015 contra la referencia, `kimi_home_vacia` y `sin_raiz` *(nace verde)* · **(31)** sólo Kimi Code · **(32)** forzado A · **(33)** la línea tras Copilot, el total y `orden_cola` · **(34)** `tick` · **nacen verdes, tras el verde**: **(35)** SC-011, cola y `state.json` · **(36)** SC-013, md5 y listado · **(37)** `sessions/` creada después, mismo `agent` · **(38)** SC-018 · **(39)** índice ilegible: 0 eventos, el aviso y los offsets sin avanzar | **m41** `a.kimi = NuevaPasadaKimi()` siempre → (30) ×2, (37) y los `*_SinRaiz*`/`*_LineaDeResumen` de los otros cinco *(censo por nombre)* · **m42** `st.Save` antes de Kimi → (32) · **m43** `tick` sin predicado → (34) · **m44** Kimi antes que Copilot → (33/orden_cola) · **m45** la existencia, sólo en `setup()` → (37) · **m46** el error aborta la pasada → (38/pasada) · **m47** el error avanza el offset → (38/se_relee) · **m48** índice ilegible ignorado → (39) · **m49** la ruta en el evento → (35) · **M-B4a** `generarKimi` crea un fichero bajo la raíz → (36) | **M** |
| **B5** `--scan` | FR-022, FR-023 *(scan)* · SC-024 *(forma)* | **(40)** fichero → `evento:`, resumen y nada en disco; `fork` → «heredados 2» · **(41)** los otros cinco formatos como en la 0.8.0 *(nace verde)* · **(42)** `metadata` sin `created_at` → no es Kimi *(**nace verde**; lo valida m52)* · **(43)** S y A de la ruta: el `event_id` del vector 1 | **m50** todo es Kimi → (41) ×5 y los `TestScan_*`/`*Scan_*` existentes *(censo por nombre)* · **m51** `dryRunKimi` guarda estado → (40/disco) · **m52** detección sin `created_at` → (42) · **m53** S y A intercambiados → (43) · **m54** `--scan` sin frontera → (40/fork) | **M** |
| **B6** README y CHANGELOG | FR-028 · P-11 | `grep -c '^## 0.9.0'` = 0 · `grep -c '^### Kimi Code'` = 0 | — *(textos aprobados ✋; `cmp`)* | **S** |

**55 mutaciones**: m1–m54 y M-B4a. **43 tests numerados: 33 rojos y 10 nacidos verdes** *((12), (27), (30), (35)–(39), (41), (42))*; cada nacido verde
lleva la mutación que lo valida.

**Estimación** *(horas de Claude, sin las del dueño)*:

| | B0 | B1 | B2 | B3 | B4 | B5 | B6 | C1–C3 |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Producción *(líneas)* | — | ~60 | ~150 | ~230 | ~90 | ~60 | ~50 de texto | — |
| Test *(líneas)* | — | ~110 | ~300 | ~420 | ~300 | ~120 | — | — |
| Horas | 0,5 | 1 | 2 | 3,5 | 2,5 | 1 | 0,5 | 2 |

**~590 de producción y ~1 250 de test; ~13 h.** El lector de Copilot tiene 440 *(`copilot.go` 226 + `copilot_contexto.go` 214)*. El riesgo de SC-022 es B2 + B3.

## Cierre — en tramos, uno por mensaje *(si uno falla, se para y se rehace desde C1)*

| Tramo | Qué | Quién |
|---|---|---|
| **C1** · Puertas | gofmt, vet y lint 0; 953 + nuevos, 0 SKIP; SC-016; **SC-022** en las 5; SC-023; SC-020; compilan Windows y darwin | Claude |
| **C2** · Medidas | Cada suite, tras `~/bin/libre.sh`. Sobre **las copias A y B**, con sus sha256 antes y después *(PARA 9)*: **SC-001** y **SC-024** sobre B *(`--run` con `KIMI_CODE_HOME` = una copia en un temporal; `--scan`)*; **SC-002** *(A, y luego B)*; **SC-003**; **SC-004** sobre S6. Contraste: un guion `python3 -I` fuera del repo cuenta registros, heredados y partidas | Claude |
| **C3** · Snapshot | `goreleaser release --snapshot --clean`, `strings` con los textos *(SC-017)* | Claude |
| **C4** · **W1** | En Windows, en sandbox *(`KIMI_CODE_HOME` propio)*, sin enrolar: **D-012-W** | ✋ dueño |
| **C5** · PR | El cuerpo, fuera del repo | Claude redacta |
| **C6** · Fusión | Merge commit | ✋ dueño |
| **C6b** · Plataforma | **La fila `kimi-k2.7-code` desplegada en producción**, comprobada con `CatalogoTarifas::cargar()->esUtilizable("kimi-k2.7-code")` → `true` *(D-5)*. **Sin ella, no hay C7** | ✋ dueño |
| **C7** · Etiqueta | `v0.9.0` | ✋ dueño |
| **C8–C9** | Canales y **W2** *(SC-025, en la instalación real)* | tras C7 |

**D-012-W · W1** *(SC-025)*: el contador independiente es un guion `python3 -I` que lee sólo líneas `usage.record` y `forked` *(no comparte código con el lector)*.
1. **Kimi cerrado**: md5 y listado → `--run` → **iguales** *(SC-013)*. Eventos = registros del contador, menos los heredados; un segundo `--run` → 0.
2. **Kimi abierto**, con `--daemon` *(60 s)*: dos prompts, un subagente, un `/compact`, un `/fork` con un prompt en el fork, y una petición matada con `taskkill /F`.
3. **El cuadre, petición a petición**: el dueño exporta «Request Details» de la ventana. Cada fila de la consola casa con **un** evento por `entrada = tokens_input +
   tokens_cache_read`, `caché = tokens_cache_read` y `salida = tokens_output`, en orden de hora. **Cada fila sin evento se declara matada, con su hora**; un evento
   sin fila, o una fila con dos eventos, es **PARA**.

**W2**: tras C8, `scoop update` y `--run` en la instalación real. El mismo cuadre y, **evento a evento**, el coste de la plataforma = la tarifa oficial × sus partidas.

## Trazabilidad *(FR → bloque)*

| FR | Bloque | | FR | Bloque | | FR | Bloque | | FR | Bloque |
|---|---|---|---|---|---|---|---|---|---|---|
| 001 | B1 | | 008 | B3 | | 015 | B2 | | 022 | B5 |
| 002 | B1 · B4 | | 009 | B3 | | 016 | B2 | | 023 | B4 · B5 |
| 003 | B1 | | 010 | B2 | | 017 | B2 | | 024 | B4 |
| 004 | B3 | | 011 | B2 | | 018 | B2 · B3 | | 025 | B4 |
| 005 | B3 | | 012 | B2 · B3 | | 019 | B2 · B3 · B4 | | 026 | B4 |
| 006 | B3 · B4 | | 013 | B2 | | 020 | B2 · B3 | | 027 | todos |
| 007 | B2 · B3 | | 014 | B3 | | 021 | B4 | | 028 | B6 · C7 |

**28 / 28 FR con bloque.** Los 26 SC, en `tasks.md` §Cobertura.

## Riesgos

| # | Riesgo | Medida o mitigación |
|---|---|---|
| R-1 | **SC-022**: ~590 líneas. En 011, 440 líneas costaron +33 760 B como mucho *(011 registro §Cierre)* | Sin paquetes nuevos. Medidas informativas tras B2, B3 y B4. Si no cabe, se adelgaza; el tope no sube |
| R-2 | **El prepaso** lee el fichero entero cada vez que crece | Sólo se decodifican las `forked`. Lo mismo que el prefijo de Copilot, con la cola dos veces |
| R-3 | **Un dato real en testdata** | §6 y PARA 8, con su vigilancia en cada corte |
| R-4 | Un test hereda las sesiones reales del desarrollador | M-3, y que los `agent` hechos a mano no las lean |
| R-5 | **Windows**: el agente con el fichero abierto durante la reescritura atómica de Kimi *(spec Q-3)* | Lecturas cortas; W1 lo observa. Si Kimi falla, PARA y enmienda |
| R-6 | Los forzados de SC-018 y SC-019 dependen de permisos POSIX | `t.Skip` con root o en Windows, como 011 R-5 |
| R-7 | Kimi cambia de forma *(una versión nueva)* | Se cuenta como incoherente o como no-registro, nunca se rompe. C2 rehecho sobre una copia nueva |
| R-8 | **La fila de tarifa no llega a producción** | C6b bloquea C7 |

## Complexity Tracking

| Qué | Por qué | Alternativa descartada |
|---|---|---|
| Un campo nuevo en `FileState` *(`Cabeza`)* | E-1, decisión del dueño | Sólo el offset: no detecta la migración que cambia la cabeza |
