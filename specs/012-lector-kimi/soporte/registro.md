# 012 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007–011, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. Techo: **300** líneas
*(tasks §Presupuesto)*.

**Cómo se corre cada suite** *(tasks §Una suite a la vez)*: antes, **en una orden sola y en segundo plano**, `bash ~/bin/libre.sh` *(md5 `ab647c01…`)*,
hasta «inicio:» y `libre=0`; la suite va en la orden siguiente. Se anotan `/proc/loadavg` y la duración. **Ninguna orden que espere a `libre.sh` lleva en su
línea de órdenes el nombre de una suite** *(011 registro §B0, incidencia)*.

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

## B2 · Una línea e identidad

*(T010–T017)*

## B3 · Fichero, cabeza e índice

*(T019–T024)*

## B4 · Integración

*(T026–T032)*

## B5 · `--scan`

*(T034–T036)*

## B6 · README y CHANGELOG

*(T038–T040)*

## Cierre

*(T042–T051)*
