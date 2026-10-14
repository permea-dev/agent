# 011 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007–010, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. Techo: **300** líneas
*(tasks §Presupuesto; dueño, 14-10)*.

**Cómo se corre cada suite** *(tasks §Una suite a la vez)*: antes, **en una orden aparte y en segundo plano**, `bash ~/bin/libre.sh` *(md5 `ab647c01…`)*,
hasta «inicio:». Se anotan `/proc/loadavg` y la duración. **Ninguna orden que espere a `libre.sh` puede llevar en su línea de órdenes el nombre de una
suite** *(§B0, incidencia)*.

## B0 · FASE 0 y registro

### T001 · Línea base y puertas *(2026-10-14, sobre `e9ad364`; desde `15c4c47`, sólo documentos y `.specify`)*

**La línea base de la 0.7.0** *(encargo 5/n; `git archive 15c4c47` en un temporal, tras `libre.sh`)*: `go test -count=1 -v ./...` → **848 / 0 / 0** en
10 paquetes; `go vet` limpio; lint **0**. **Tamaños** *(`CGO_ENABLED=0 -trimpath -ldflags "-s -w -X main.version=0.7.0"`; referencia de SC-007)*:

| Plataforma | Fichero *(B)* | Nota |
|---|---:|---|
| linux/amd64 | 5 648 536 | |
| linux/arm64 | 5 439 640 | Σ `FileSiz` de los `LOAD`: **5 337 116** |
| darwin/amd64 | 5 810 800 | holgura de `__TEXT`: 846 B |
| darwin/arm64 | 5 571 330 | holgura de `__TEXT`: 14 144 B |
| windows/amd64 | 5 846 528 | |

**La FASE 0 sobre `e9ad364`**, en el árbol de trabajo:

| Puerta | Resultado | Carga *(`/proc/loadavg`)* y duración |
|---|---|---|
| `bash ~/bin/libre.sh` | «inicio» a las 13:45:50, tras esperar desde las 13:39:28 *(otra carga en la máquina)* | 1.42 3.39 2.66 |
| `go test -count=1 -v ./...` | **848 «--- PASS» · 0 FAIL · 0 SKIP**, 10 paquetes `ok`, rc 0 | 13:46:00 → 13:46:09, **8,5 s** · 1.28 → 1.38 |
| `bash ~/bin/libre.sh` | «inicio» a las 13:57:02 *(3 muestras: 1.33 · 1.43 · 1.36)* | 1.36 1.50 1.90 |
| `go vet ./...` · `golangci-lint run` | limpio · **0 issues** | 13:57:10 → 13:57:12, **1,7 s** · 1.45 |
| `gofmt -l .` | vacío | — |
| frontera de FR-027, `git diff 15c4c47` | **vacío** | — |
| `git diff 15c4c47 --name-only --diff-filter=M -- '*_test.go'` | vacío | — |
| `git diff d81c39d --name-only -- '*.go'` | vacío | — |
| `grep -c require go.mod` · `go.sum` | 0 · no existe | — |
| `os.Getenv` de producción | **4** *(`codex.go:19`, `gemini.go:20`, `opencode.go:21`, `:25`)*; pasa a **5** en B1 | — |
| `PRIV` *(tasks §Puertas)* sobre `git diff 15c4c47` | **0** | — |
| `testdata/copilot/` | aún no existen *(nacen en B1)* | — |

**`git diff 15c4c47 --name-only`**: `.specify/feature.json`, `.specify/memory/constitution.md`, y en `specs/011-lector-copilot/`, `contracts/event-id-copilot.md`,
`plan.md`, `spec.md`, `tasks.md` y los tres `soporte/descubrimiento*.md`. **Ningún `.go`.** ⇒ La FASE 0 = la línea base: **848 / 0 / 0**, como en el C1 de 010.

**Incidencia** *(sin efecto en los resultados)*: la primera espera antes de lint se **bloqueó contra sí misma**. La orden que esperaba a `libre.sh`
llevaba `go vet` y `golangci-lint` en su línea de órdenes, y `libre.sh` la veía como una suite en marcha *(la cosecha 279 de la plataforma)*. Se pararon
las dos, se relanzó `libre.sh` solo, y la espera siguiente miró sólo su fichero de salida. **Regla**, al principio de este registro.

**Sin mutaciones en B0.**
