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

## B1 · Raíz y ficheros *(`internal/config/copilot.go`, `ListarCopilot`, `internal/testutil/sandbox.go`)*

- **T003** *(2026-10-14, sobre `7eadff6`)*: stubs `CopilotSessionsRoot` → `("", nil)` e `ingest.ListarCopilot` → `(nil, nil)`; vet limpio.
- **T004**: `internal/ingest/testdata/copilot/generar.sh` `bb9b7c30…` *(`python3 -I`)* → `sesiones/`: `s-…01/events.jsonl` `448bd776…`,
  `s-…01/checkpoints/index.md` `87612f2c…`, `s-…02/events.jsonl` `b989848d…`, `s-…02/sub/events.jsonl` `fcf7d524…` *(a dos niveles)*,
  `s-…03/checkpoints/index.md` `e9733dcb…` *(sin `events.jsonl`)* y `.session-operation-locks/s-…01.lock` *(0 B)*. **El `session.db` de 0 B no se
  versiona** *(`.gitignore`: `*.db`; tocar `.gitignore` sería PARA 1)*: lo crea el test en su copia de `t.TempDir()`. Ningún test abre `testdata/` en su sitio.
- **T005 · Rojos** *(libre 0.58; 14:24:13 → :24, 10,8 s)*: **848 / 13 / 0**. Caen, por el stub:
  - (1) `TestCopilotSessionsRoot_Ubicacion/{definida,vacia,ausente,inexistente}`: `= ("", <nil>); se esperaba (<ruta>, nil)`;
  - (2) `TestSandbox_VaciaCopilotHome`: `COPILOT_HOME = <tmp> (definida: true)`, porque la variable se puso antes de `Sandbox`;
  - (3) `TestListarCopilot/{solo_events,sin_locks,sin_checkpoints_db,hondo,vacia,orden}`: «ListarCopilot no devolvió nada»;
  - y los 2 padres.
- **T006 · Verde**:
  - `copilot.go` `7e1b20a9…`, `copilot_contexto.go` `19d09e34…`;
  - `sandbox.go` `2c6b3301…`: una línea tras `OPENCODE_DB`, con su comentario;
  - los comentarios de `codex.go`, `gemini.go` y `opencode.go`, «una de las CINCO» / «dos de las CINCO».

  Resultados *(libre 0.72; 14:25:28 → :35, 7,3 s)*: **861 / 0 / 0**. Vet y lint 0 *(3,3 s)*; gofmt vacío.
  Tests: `copilot_test.go` `b7479e18…`, `copilot_fichero_test.go` `48190957…`, `sandbox_copilot_test.go` `59cd7ac4…`.
- **T007**: `sandbox_test.go`, verde y sin tocar *(`TestSandbox_AislaLasTresVariables`, `…DistingueSegmentos` ×6, `…ConSemillas…`)*; `os.Getenv` de
  producción → **5** *(`copilot.go:19`, `codex.go:19`, `gemini.go:20`, `opencode.go:21`, `:25`)*.

**T008 · Censo** *(ANTES de mutar)*: (1) = `TestCopilotSessionsRoot_Ubicacion/`, (2) = `TestSandbox_VaciaCopilotHome`, (3) = `TestListarCopilot/`.
- **Quién más las lee**: nadie *(grep)*. Ni el binario ni otros tests llaman aún a `CopilotSessionsRoot` o `ListarCopilot` *(B4)*, ni leen `COPILOT_HOME`.
- **Co-caídas declaradas aquí que plan §Bloques no listaba**: **m2 → (1/vacia)**, porque la rama del hogar pierde también `session-state`. Y **m5 →
  (3/solo_events)**, porque la lista exacta gana `s-…02/sub/events.jsonl`. Con cada caída de hoja cae también su padre.

| Mut. | Edición | Declarado |
|---|---|---|
| m1 | `propia := os.Getenv(…); propia != ""` → `propia, definida := os.LookupEnv(…); definida` *(vacía como definida)* | (1/vacia) |
| m2 | sin `"session-state"` en los dos `Join` | (1/definida), (1/vacia), (1/ausente), (1/inexistente) |
| m3 | `propia != ""` → `false && propia != ""` | (1/definida) |
| m4 | sin la línea `t.Setenv("COPILOT_HOME", "")` del sandbox *(y su comentario)* | (2) |
| m5 | `os.ReadDir` de un nivel → `filepath.WalkDir` recursivo: todo `events.jsonl` a cualquier profundidad, saltando las carpetas que empiezan por `.` | (3/solo_events), (3/hondo) |
| m6 | carpeta sin `events.jsonl` → `return nil, error` | (3) ×6 |

**T008 · Observado** *(cada mutación: md5 de partida comprobado, `bash ~/bin/libre.sh` aparte, suite, reversión por edición inversa y md5 restaurado)*:

| Mut. | Fichero · md5 mutado | Suite *(libre «inicio»)* | Cayó | = declarado |
|---|---|---|---|:-:|
| m1 | `copilot.go` `bb243797` | 859 / 2 *(1.13)* | (1/vacia) + padre | ✅ |
| m2 | `copilot.go` `638f5868` | 856 / 5 *(0.77)* | (1/definida, vacia, ausente, inexistente) + padre | ✅ |
| m3 | `copilot.go` `16bd7b2c` | 859 / 2 *(0.64)* | (1/definida) + padre | ✅ |
| m4 | `sandbox.go` `98c03742` *(= el de 010 B5: sólo falta la línea nueva)* | 860 / 1 *(1.77)* | (2) | ✅ |
| m5 | `copilot_contexto.go` `d6f4b25f` | 858 / 3 *(1.36)* | (3/solo_events, hondo) + padre | ✅ |
| m6 | `copilot_contexto.go` `2a04c9b3` | 854 / 7 *(1.56)* | (3) ×6 + padre | ✅ |

**6 de 6 = declarado.** Tras m6, los tres md5 son los del verde *(`7e1b20a9`, `19d09e34`, `2c6b3301`)*. Cada suite tardó 7–8 s.
- **Incidencia, sin efecto**: la primera reversión de m4 no se aplicó. La edición inversa partía de una cadena vacía y no tenía ancla única. El guion se paró
  en su aserción, **sin escribir nada**. Se ancló a la línea de `OPENCODE_DB`, se revirtió y el md5 dio el del verde, antes de tocar m5.

**Puertas del corte** *(libre 1.69; 14:35:05 → :12)*:
- **861 / 0 / 0**; vet y lint 0 *(libre 1.45)*; gofmt vacío;
- frontera de FR-027 vacía; ningún `*_test.go` modificado; `go.mod` sin `require`; sin `go.sum`; `os.Getenv` → **5**;
- **`PRIV`** *(tasks §Puertas)*: 0 en `git diff 15c4c47`, 0 en cada fichero nuevo sin seguir, y 0 con `grep -raEc` en los 7 ficheros de
  `internal/ingest/testdata/copilot/` *(`generar.sh` incluido)*.
