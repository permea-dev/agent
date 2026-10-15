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

## B1 · Raíz y ficheros

*(T003–T008)*

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
