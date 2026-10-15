# 012 · Registro de rojos, verdes y mutaciones — archivo

**§B0, §B1 y §B2, movidos íntegros** desde [`registro.md`](./registro.md) el 2026-10-15, al empezar B5 *(dueño, como en 011)*. **No cuenta para el techo** de
300 líneas. Texto cerrado: no se edita.

## B0 · FASE 0 y registro

### T001 · Línea base y puertas *(2026-10-15, sobre `3779671`; desde `e9e5449`, sólo documentos y `.specify`)*

**`git diff e9e5449 --name-only`**: `.specify/feature.json` y, en `specs/012-lector-kimi/`, `spec.md`, `plan.md`, `tasks.md`, `contracts/event-id-kimi.md` y los tres
`soporte/descubrimiento*.md`. **Ningún `.go`** *(`git diff e9e5449 --name-only -- '*.go'` → vacío)* ⇒ el código de este HEAD es el de `e9e5449`, y la FASE 0 es la línea base.

| Puerta | Resultado | Carga *(`/proc/loadavg`)* y duración |
|---|---|---|
| `bash ~/bin/libre.sh` | **`libre=0`**: «inicio» a las 11:57:13, tras esperar desde las 11:53:23 a una suite `playwright` de la plataforma *(3 muestras: 1.90 · 1.76 · 1.79)* | 1.79 3.08 3.54 |
| `go test -count=1 -v ./...` | **953 «--- PASS» · 0 FAIL · 0 SKIP**, 10 paquetes `ok`, rc 0 | 11:57:23 → 11:57:35, **12,1 s** · 1.75 → 3.72 |
| `bash ~/bin/libre.sh` | **`libre=0`**: «inicio» a las 12:12:17, tras esperar desde las 11:57:41 a suites de la plataforma *(3 muestras: 1.80 · 1.59 · 1.43)* | 1.43 3.43 3.99 |
| `go vet ./...` · `golangci-lint run` *(2.12.2)* | limpio · **0 issues** | 12:12:28 → 12:12:31, **2,5 s** · 1.21 |
| `gofmt -l .` | vacío | — |
| frontera de FR-027, `git diff e9e5449` *(tasks §Puertas)* | **vacío** *(0 B)* | — |
| `git diff e9e5449 --name-only --diff-filter=M -- '*_test.go'` | vacío | — |
| `grep -c require go.mod` · `go.sum` | 0 · no existe | — |
| `os.Getenv` de producción | **5** *(`codex.go:19`, `gemini.go:20`, `opencode.go:21`, `:25`, `copilot.go:19`)*; pasa a **6** en B1 | — |
| `PRIV` *(tasks §Puertas)* sobre `git diff e9e5449` | **0** | — |
| `testdata/kimi/` | aún no existen *(nacen en B1)* | — |

**Paquetes**: `cmd/permea` 9,8 s · `config` · `event` · `ingest` 1,1 s · `pricing` · `project` · `sqlitero` · `state` · `testutil` · `transport`.

**Tamaños de la 0.8.0** *(referencia de SC-022; `git archive e9e5449` en un temporal; `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=0.8.0"`,
las banderas de `.goreleaser.yaml`; go1.22.2)*:

| Plataforma | Fichero *(B)* | Nota |
|---|---:|---|
| linux/amd64 | 5 681 304 | |
| linux/arm64 | 5 439 640 | Σ `FileSiz` de los `LOAD`: **5 362 876** |
| darwin/amd64 | 5 840 208 | holgura de `__TEXT`: **1 004 B** |
| darwin/arm64 | 5 605 090 | holgura de `__TEXT`: 16 102 B |
| windows/amd64 | 5 874 688 | |

**Contraste con 011**:
- linux/amd64, darwin/amd64 y darwin/arm64 son exactamente los de la 0.7.0 más los Δ del C1 de 011 *(011 registro §Cierre)*.
- **windows/amd64 da 512 B menos, y la Σ de `LOAD` de arm64, 256 B menos.**
- **Causa, medida**: compilada en el **árbol de trabajo** *(con `.git`)*, la misma 0.8.0 da **5 875 200** en Windows y **5 363 132** en arm64, las cifras de 011. Go
  incrusta entonces 3 líneas `vcs.*` *(`go version -m`)*, y un `git archive` no tiene ninguna.
- ⇒ **SC-022 se mide `git archive` contra `git archive`** *(plan §7)*, con las cifras de esta tabla.

**Incidencias**: ninguna. Las dos esperas fueron a suites de la plataforma *(P-034 B3)*, como se preveía.

**Sin mutaciones en B0.**

**Nota del dueño** *(15-10)*: el asunto de `8604637` salió sin «012 spec ratificada:» por un fallo del bloque de commit del orquestador *(la barra de continuación antes del heredoc)*; el contenido es el correcto, y no se reescribe la historia.

## B1 · Raíz y ficheros *(`internal/config/kimi.go`, `ListarKimi`, `internal/testutil/sandbox.go`)*

- **T003** *(2026-10-15, sobre `6eece4c`)*: stubs `config.KimiHome` → `("", nil)` e `ingest.ListarKimi` → `(nil, nil)`, en `kimi.go` y `kimi_contexto.go`; compila, vet limpio.
- **T004**: `internal/ingest/testdata/kimi/generar.sh` `24c285f4…` *(`python3 -I`)* → `raiz/`, 13 ficheros: `sessions/wd_proyecto-a_000000000001/`
  con `session_sintetica-0001/agents/{main,agent-0}/wire.jsonl`, `main/wire.jsonl.bak`, `main/wire.jsonl.tmp.1.ab`, `main/sub/wire.jsonl` *(hondo)* y
  `agents/.oculto/wire.jsonl`; `session_sintetica-0002/agents/main/wire.jsonl`; `session_sintetica-0003/logs/texto-1.log` *(sin `agents/`)*; `.oculta/agents/main/wire.jsonl`;
  `sessions/.index-cache/scan.json`, `sessions/.index-dirty/session_sintetica-0002.1` *(0 B)*, `sessions/.oculta/session_sintetica-0009/agents/main/wire.jsonl`; y
  `imported/imp-0001/agents/main/wire.jsonl`. md5 *(en ese orden de `find | sort`)*: `46db5cf4 21063e74 d41d8cd9 3a71dfff 3d0851ec 07e716f4 94e26bd4 0c9130d7 ×4 23d4a375
  e2ec519e`. **Ningún `state.json`** *(`.gitignore`)*. Ningún test abre `testdata/` en su sitio.
- **T005 · Rojos, la razón escrita ANTES de correrlos**: con los stubs, se espera **953 / 14 / 0**:
  - (1) `TestKimiHome_Ubicacion/{definida,vacia,ausente,inexistente}`: `KimiHome() = ("", <nil>); se esperaba (<ruta>, nil)`, porque el stub no resuelve;
  - (2) `TestSandbox_VaciaKimiCodeHome`: `KIMI_CODE_HOME = <tmp> (definida: true)`, porque `Sandbox` aún no la vacía;
  - (3) `TestListarKimi/{forma,sin_punto,sin_bak_tmp,sin_imported,hondo,vacia,orden}`: «ListarKimi no devolvió nada», porque el stub no lista;
  - y los 2 padres.
- **T005 · Rojos, corridos** *(libre 1.31, «inicio» 12:33:45 tras esperar desde 12:28:26; 12:33:55 → :34:04, 9,0 s; 1.27 → 2.04)*: **953 / 14 / 0**, rc 1.
  **Caen exactamente los 14 declarados, con su razón**: (1) ×4 `KimiHome() = ("", <nil>)`; (2) `KIMI_CODE_HOME = "<tmp>" (definida: true)`; (3) ×7
  «ListarKimi no devolvió nada»; y los 2 padres.
- **T006 · Verde**: `internal/config/kimi.go` `1b57111b…` *(P-1)*; `internal/ingest/kimi_contexto.go` `5803a7cc…` *(`ListarKimi`, `subcarpetasKimi`,
  `visiblesKimi`)*; `sandbox.go` `4a10ca7c…` *(una línea tras `COPILOT_HOME`, con su comentario)*; y los comentarios de `codex.go:16`, `gemini.go:17`, `opencode.go:18`
  y `copilot.go:16`, «una de las SEIS» / «dos de las SEIS». gofmt vacío, vet limpio. **Suite** *(libre 1.36, «inicio» 12:42:05; 12:42:15 → :28, 13,3 s; 1.41 → 6.30)*:
  **967 / 0 / 0**, rc 0, 10 paquetes `ok` *(953 + 14)*.
- **T007**: `sandbox_test.go` sin tocar y en verde *(`TestSandbox_AislaLasTresVariables`, `…DistingueSegmentos`, `TestSandboxConSemillas_…`, en PASS en esa suite)*.
  **`os.Getenv` de producción → 6** *(`codex.go:19`, `gemini.go:20`, `opencode.go:21`, `:25`, `copilot.go:19`, `kimi.go:19`)*.
- **T008 · Censo de m1–m8, declarado ANTES de mutar** *(md5 de partida: `kimi.go` `1b57111b…`, `kimi_contexto.go` `5803a7cc…`, `sandbox.go` `4a10ca7c…`)*.
  Ajustes al plan, razonados sobre `raiz/`: **m2** cae también en (1/vacia) *(espera `.kimi-code`)*; **m5**, **m6**, **m7** y **m8** tiran también (3/forma), porque
  cambian la lista; (3/orden) **no** cae en ninguna *(los añadidos quedan en orden léxico; con m7, la lista tiene un solo elemento)*.

  | Mut. | Fichero | Cambio | Debe caer *(hojas; + el padre)* | FAIL |
  |---|---|---|---|---:|
  | m1 | `kimi.go` | `os.Getenv` → `os.LookupEnv`, vacía como definida | (1/vacia) | 2 |
  | m2 | `kimi.go` | sin `.kimi-code`: devuelve el hogar | (1/vacia), (1/ausente), (1/inexistente) | 4 |
  | m3 | `kimi.go` | la variable se ignora *(`propia := ""`)* | (1/definida) | 2 |
  | m4 | `sandbox.go` | sin `t.Setenv("KIMI_CODE_HOME", "")` | (2) | 1 |
  | m5 | `kimi_contexto.go` | `visiblesKimi` sin el filtro de `.` | (3/forma), (3/sin_punto) | 3 |
  | m6 | `kimi_contexto.go` | `f.Name() == "wire.jsonl"` → `strings.HasPrefix(f.Name(), "wire.jsonl")` | (3/forma), (3/sin_bak_tmp) | 3 |
  | m7 | `kimi_contexto.go` | `sesiones := raiz` | (3/forma), (3/sin_imported) | 3 |
  | m8 | `kimi_contexto.go` | recursivo: `filepath.WalkDir` en cada agente | (3/forma), (3/hondo) | 3 |
- **T008 · Mutaciones, corridas** *(cada suite, tras `libre.sh` solo con `libre=0`; edición inversa y md5 comprobados tras cada una)*: **8 de 8 = declarado.**

  | Mut. | md5 puesta | Suite *(libre · hora · duración)* | Cayó | → restaurado |
  |---|---|---|---|---|
  | m1 | `kimi.go` `c013f8fb` | 1.62 · 12:47:28 · 7,6 s | 965 / **2**: (1/vacia) y el padre | `1b57111b` ✓ |
  | m2 | `kimi.go` `146aa961` | 1.60 · 12:52:47 · 9,4 s | 963 / **4**: (1/vacia), (1/ausente), (1/inexistente) y el padre | `1b57111b` ✓ |
  | m3 | `kimi.go` `76256114` | 1.29 · 12:57:01 · 7,5 s | 965 / **2**: (1/definida) y el padre | `1b57111b` ✓ |
  | m4 | `sandbox.go` `e21d96a5` | 0.94 · 12:59:41 · 7,5 s | 966 / **1**: (2) | `4a10ca7c` ✓ *(a mano; incidencia 1)* |
  | m5 | `kimi_contexto.go` `4fbf782a` | 0.79 · 13:00:37 · 7,1 s | 964 / **3**: (3/forma), (3/sin_punto) y el padre | `5803a7cc` ✓ |
  | m6 | `kimi_contexto.go` `3456739c` | 1.50 · 13:01:45 · 7,1 s | 964 / **3**: (3/forma), (3/sin_bak_tmp) y el padre | `5803a7cc` ✓ |
  | m7 | `kimi_contexto.go` `940149cf` | 1.67 · 13:05:34 · 7,7 s | 964 / **3**: (3/forma), (3/sin_imported) y el padre | `5803a7cc` ✓ |
  | m8 | `kimi_contexto.go` `32d82ae3` | 0.57 · 13:08:14 · 9,1 s | 964 / **3**: (3/forma), (3/hondo) y el padre | `5803a7cc` ✓ |

  **Incidencias** *(ninguna cambia un resultado)*:
  1. **m4**: el guion de mutaciones no sabía invertir un **borrado** *(buscaba la cadena vacía)* y abortó tras una suite ya correcta. La línea se repuso a mano, tras
     su comentario, y el md5 volvió a `4a10ca7c`; el guion se corrigió antes de m5.
  2. **m5**: quitar la llamada a `strings.HasPrefix` dejaba `strings` sin usar y **no compilaba** *(md5 `61eb9c78`, sin suite: no cuenta)*. Se revirtió y se reformuló
     con el mismo efecto, `!strings.HasPrefix(e.Name(), "\x00")` *(ningún nombre lo cumple)*: es la que se corrió.
- **T009 · Corte** *(antes del commit del dueño)*: **suite** *(libre 1.53, «inicio» tras 13:11; 13:11:35 → :43, 8,4 s; 1.49 → 1.33)* **967 / 0 / 0**, rc 0; **vet** limpio y
  **lint 0 issues** *(libre 1.01; 13:12:13 → :18, 5,2 s)*; gofmt vacío; frontera de FR-027 0 B; ningún `_test.go` de `e9e5449` modificado; `go.mod` sin `require`, sin `go.sum`;
  `os.Getenv` → **6**; `PRIV` → 0 en `git diff e9e5449`, 0 en los 14 ficheros de `testdata/kimi/` *(`generar.sh` incluido)* y 0 en los nuevos.

## B2 · Una línea e identidad *(`kimi_eventid.go`, `kimi.go`)*

- **T010** *(2026-10-15, sobre `fb69293`)*: stubs `derivarEventIDKimi(S, A, N)` → `""` y `RegistroKimi` → `(nil, NoEsRegistroKimi, {}, nil)`; los tipos `ContextoKimi`
  *(S, A, N, `Heredado`, `Cwd`, `Decodificadas`)*, `ClaseKimi` *(no es registro · evento · heredado · a cero · incoherente)* y `MarcasKimi` *(compactación, sin modelo)*. Compila.
- **T011**: `generar.sh` `812f7762…` gana `lineas/`, 22 ficheros: `evento`, `compactacion`, `sc005`, los seis de SC-008 *(`modelo_prefijo`, `_sin_barra`, `_varias`, `_inicial`,
  `_acabada`, `_vacio_o_no_textual` con dos líneas)*, las siete `incoherente_*`, `a_cero`, `ms` *(time 1)*, `centinelas` y tres que no son registro *(`falso_texto`, `forked`,
  `subagent_completed`)*. `raiz/` no cambia al regenerar.
- **T012 · Rojos, la razón escrita ANTES de correrlos**: con los stubs, se espera **971 / 33 / 0**:
  - (4) ×7 hojas y el padre: `derivarEventIDKimi(…) = ""; se esperaba <vector del contrato>`;
  - (5), (8), (11), (13), y (10) ×4 hojas con su padre: el stub no da evento *(«se esperaba un evento» / «se esperaba “a cero”»)*;
  - (6) ×6 hojas y el padre: el stub no da evento;
  - (7) ×7 hojas y el padre: clase 0 *(no es registro)*; se esperaba incoherente;
  - (9): clase 0; se esperaba «heredado»;
  - **nace verde (12)** ×3 hojas y su padre *(el stub no decodifica nada; lo valida m22)*: 967 + 4 = 971 PASS.
- **T012 · Rojos, corridos** *(libre 1.37, «inicio» tras esperar desde 13:19:25; 13:20:01 → :09, 7,8 s; 1.34 → 2.09)*: **971 / 33 / 0**, rc 1. **Caen exactamente los 33
  declarados, con su razón** *(los 7 vectores con `= ""`; los demás con «clase 0» o «se esperaba un evento»)*, y (12) nace verde con sus 4 PASS.
- **T013 · Verde**: `kimi_eventid.go` `4f48b961…` *(la codificación de `copilot_eventid.go:33-52`, replicada, con `dominioEventID`, `bytesEventID` y `tipoRespuesta`)*
  y `kimi.go` `7d4a5e65…` *(prefiltro anclado → heredado → incoherente → a cero → evento; partidas tal cual; P-9 con `LastIndexByte`; `time.UnixMilli(·).UTC()`)*. No
  llama a `internal/pricing`. **Suite** *(libre 1.75; 13:32:51 → :33:04, 12,9 s; 1.69 → 4.99)*: **1004 / 0 / 0**, rc 0 *(971 + 33)*.
- **T014**: (13), rojo en T012, pasa a verde con T013.
- **T016 · SC-022, informativo tras B2** *(los ficheros del árbol, versionados y nuevos, copiados sin `.git` a un temporal: el equivalente a un `git archive`, sin sellado VCS;
  0.9.0; las banderas de `.goreleaser.yaml`)*: **Δ = 0 en las 5 plataformas** *(5 681 304 · 5 439 640 · 5 840 208 · 5 605 090 · 5 874 688)*; arm64, Σ `LOAD` 5 362 876 *(Δ 0)*;
  **holgura de `__TEXT` en darwin/amd64: 1 004 B, sin cambio** *(arm64: 16 102)*. El lector aún no está enlazado: `main` no lo llama hasta B4.
- **T017 · Vigilancia de privacidad** sobre `lineas/` *(22 ficheros)* y `generar.sh`: **`PRIV` → 0**.
- **T015 · Censo de m9–m24, declarado ANTES de mutar** *(md5 de partida: `kimi_eventid.go` `4f48b961…`, `kimi.go` `7d4a5e65…`; **las 16 compilan**, comprobado aplicando,
  `go vet` y revirtiendo, con el md5 restaurado)*. Ajustes al plan: **m9** cambia sólo el componente del hash *(no la constante: `tool` no cae)*; **m10, m11 y m12** tiran
  las **7** hojas de (4) *(cualquiera cambia los 7 valores)*; m22 tira las 3 hojas de (12).

  | Mut. | Fichero | Cambio | Debe caer *(hojas; + el padre)* | FAIL |
  |---|---|---|---|---:|
  | m9 | `kimi_eventid.go` | `herramientaKimi` → `"codex"` en los componentes | (4) ×7 | 8 |
  | m10 | `kimi_eventid.go` | sin el prefijo de longitud *(borrada la línea `h.Write(longitud[:])`)* | (4) ×7 | 8 |
  | m11 | `kimi_eventid.go` | S y A cambiados de orden | (4) ×7 | 8 |
  | m12 | `kimi_eventid.go` | N con un `0` delante | (4) ×7 | 8 |
  | m13 | `kimi.go` | `TokensInput` = `otra + lectura` | (5) | 1 |
  | m14 | `kimi.go` | lectura ↔ escritura | (5) | 1 |
  | m15 | `kimi.go` | `IndexByte` *(la primera `/`)* | (6/varias) | 2 |
  | m16 | `kimi.go` | vacío tras la barra → el alias entero | (6/acabada) | 2 |
  | m17 | `kimi.go` | `< -1`: la negativa −1 aceptada | (7/negativa) | 2 |
  | m18 | `kimi.go` | `false &&` en «a cero» | (8) | 1 |
  | m19 | `kimi.go` | `false &&` en «heredado» | (9) | 1 |
  | m20 | `kimi.go` | `CostAvailable: true` | (10/coste) | 2 |
  | m21 | `kimi.go` | sin `.UTC()` | (11) | 1 |
  | m22 | `kimi.go` | `false &&` en el prefiltro | (12) ×3 | 4 |
  | m23 | `kimi.go` | `SessionRef: ctx.S` | (13) | 1 |
  | m24 | `kimi.go` | `false &&` en la marca de compactación | (10/compactacion) | 2 |
- **T015 · Mutaciones, corridas** *(cada suite tras `libre.sh` solo con `libre=0`, carga de inicio 1.08–1.99; 6,9–8,6 s cada una; edición inversa y md5 comprobados)*:
  **16 de 16 = declarado**, sin incidencias.

  | Mut. | md5 puesta | Hora | Cayó *(PASS / FAIL)* | → restaurado |
  |---|---|---|---|---|
  | m9 | `865a000f` | 13:38:20 | 996 / 8: (4) ×7 y el padre | `4f48b961` ✓ |
  | m10 | `1c899ca7` | 13:42:11 | 996 / 8: (4) ×7 y el padre | `4f48b961` ✓ |
  | m11 | `1d67d3e2` | 13:42:50 | 996 / 8: (4) ×7 y el padre | `4f48b961` ✓ |
  | m12 | `2cc24313` | 13:43:31 | 996 / 8: (4) ×7 y el padre | `4f48b961` ✓ |
  | m13 | `780b5d7f` | 13:44:11 | 1003 / 1: (5) | `7d4a5e65` ✓ |
  | m14 | `9f4799f8` | 13:44:51 | 1003 / 1: (5) | `7d4a5e65` ✓ |
  | m15 | `d8f93a20` | 13:45:31 | 1002 / 2: (6/varias) y el padre | `7d4a5e65` ✓ |
  | m16 | `b6dcc5d9` | 13:46:11 | 1002 / 2: (6/acabada) y el padre | `7d4a5e65` ✓ |
  | m17 | `426417f7` | 13:46:51 | 1002 / 2: (7/negativa) y el padre | `7d4a5e65` ✓ |
  | m18 | `165a4993` | 13:47:30 | 1003 / 1: (8) | `7d4a5e65` ✓ |
  | m19 | `9e1c0537` | 13:48:10 | 1003 / 1: (9) | `7d4a5e65` ✓ |
  | m20 | `eee0e582` | 13:49:00 | 1002 / 2: (10/coste) y el padre | `7d4a5e65` ✓ |
  | m21 | `d93d79b4` | 13:49:40 | 1003 / 1: (11) | `7d4a5e65` ✓ |
  | m22 | `12646cdb` | 13:50:20 | 1000 / 4: (12) ×3 y el padre | `7d4a5e65` ✓ |
  | m23 | `ab77d034` | 13:51:00 | 1003 / 1: (13) | `7d4a5e65` ✓ |
  | m24 | `dd5e404c` | 13:51:42 | 1002 / 2: (10/compactacion) y el padre | `7d4a5e65` ✓ |
- **T018 · Corte** *(antes del commit del dueño)*: **suite** *(libre 1.52; 13:52:35 → :41, 6,8 s)* **1004 / 0 / 0**, rc 0; **vet** limpio y **lint 0 issues** *(libre 1.29;
  13:53:11 → :16, 4,5 s)*; gofmt vacío; frontera de FR-027 0 B; ningún `_test.go` de `e9e5449` modificado; sin `require` ni `go.sum`; `os.Getenv` → 6; **ningún fichero de
  `testdata/kimi/` ignorado por git** *(`git ls-files --others --ignored --exclude-standard` → vacío)*; los `kimi*.go` no importan `internal/pricing`; `PRIV` → 0 en el diff,
  en los 36 ficheros de `testdata/kimi/` y en los nuevos.
