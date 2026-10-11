# 010 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007, 008 y 009, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. Techo: **300**
líneas *(tasks §Presupuesto)*.

## B0 · Documentos

### T001 · Línea base *(2026-10-11, sobre `0256098`; sólo documentos desde `82e0b2d`)*

```
$ git log --oneline -2
0256098 010 remate del analyze: enmienda E-1, SC-028 y SC-029, mutaciones re-derivadas
4f7658a 010 plan y tareas: lector propio en internal/sqlitero y frontera por rowid
$ git diff 82e0b2d --name-only                      4 rutas, todas bajo specs/010-lector-opencode/
$ gofmt -l .                                        (vacío)
$ go vet ./...                                      (limpio)
$ golangci-lint run
0 issues.
$ go test -count=1 -v ./...                         671 «--- PASS», 0 «--- FAIL», 0 «--- SKIP» · 9 paquetes ok
$ git diff 4c6cafc -- <frontera de FR-030> | wc -l
0
$ git diff 4c6cafc --name-only --diff-filter=M -- '*_test.go' | wc -l
0
$ grep -c require go.mod ; ls go.sum
0 ; no existe
```
