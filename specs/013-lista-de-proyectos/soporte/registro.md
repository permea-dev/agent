# 013 · Registro de la implementación

**Por qué existe**: como en 007–012, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. **Techo: 250 líneas**, fijado por el
dueño el 23-10 *(`tasks.md` §Presupuesto)*. ⛔ **Ni rutas reales, ni nombres de carpeta, ni contenido** *(PARA 8)*: sólo recuentos, tiempos, tamaños y md5.

## B0 · Fase 0, referencia y registro

### T001 · Línea base y puertas *(2026-10-23, sobre `afeee26`; desde `cabf8c5`, sólo documentos y `.specify`)*

**`git diff cabf8c5 --name-only`**: `.specify/feature.json` y, en `specs/013-lista-de-proyectos/`, `spec.md`, `plan.md`, `tasks.md` y
`soporte/descubrimiento.md`. **Ningún `.go`** ⇒ el código de este HEAD es el de `cabf8c5`. La Fase 0 corre sobre `git archive cabf8c5` en un temporal.

| Puerta | Resultado | Carga *(`/proc/loadavg`)* y duración |
|---|---|---|
| `bash ~/bin/libre.sh` *(md5 `ab647c01…`)* | «inicio» a las 13:12:17 *(3 muestras: 0.69 · 0.74 · 0.78)* | 0.78 0.51 0.49 |
| `go test -count=1 -v ./...` | **1074 «--- PASS» · 0 FAIL · 0 SKIP**, 10 paquetes `ok`, rc 0 | 13:12:22 → :29, **6,6 s** · 0.73 |
| `bash ~/bin/libre.sh` | «inicio» a las 13:12:55 *(0.75 · 0.72 · 0.76)* | 0.76 0.53 0.50 |
| `go vet ./...` · `golangci-lint run` | limpio · **0 issues** | 13:13:00 → :01 |
| `gofmt -l .` | vacío | — |
| frontera de FR-025, `git diff cabf8c5` *(tasks §Puertas)* | **vacío** *(0 B)* | — |
| `git diff cabf8c5 --name-only --diff-filter=M -- '*_test.go'` | vacío | — |
| `grep -c require go.mod` · `go.sum` | 0 · no existe | — |
| `PRIV` *(tasks §Puertas)* sobre `git diff cabf8c5` | **0** | — |

**Paquetes**: `cmd/permea` 6,0 s · `config` · `event` · `ingest` 0,6 s · `pricing` · `project` · `sqlitero` · `state` · `testutil` · `transport`.

**`golangci-lint` es la 2.15.0** *(go1.27.0)*, no la 2.12.2 que cita plan §Technical Context. Da 0 igual. Va en §Incidencias.

**Tamaños de la 0.9.0** *(referencia de SC-016; `git archive cabf8c5` en un temporal; `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=0.9.0"`,
las banderas de `.goreleaser.yaml`; go1.22.2; sin líneas `vcs.*`)*:

| Plataforma | Fichero *(B)* | Nota |
|---|---:|---|
| linux/amd64 | 5 718 168 | |
| linux/arm64 | 5 439 640 | Σ `FileSiz` de los `LOAD` *(010 E-3)*: **5 398 044** |
| darwin/amd64 | 5 881 792 | |
| darwin/arm64 | 5 638 738 | |
| windows/amd64 | 5 914 112 | |

**Contraste con la 012, que no cuadra entero.**
- Con el mismo instrumento, en la misma sesión, **la 0.8.0** *(`git archive e9e5449`)* reproduce **exactas** dos cifras del B0 de la 012:
  - linux/amd64, 5 681 304;
  - el fichero de linux/arm64, 5 439 640.
- **Las otras tres de la 0.8.0 salen por encima**: darwin/amd64 +4 272, darwin/arm64 +176, windows/amd64 +6 656 y la Σ `LOAD` de arm64 +5 840.
- **La 0.9.0 de hoy**, frente a «0.8.0 de la 012 + Δ normativo de la 012» *(012 registro, T042)*, sale entre **+5 632 y +16 688**. En linux/amd64, +8 192.
- **Causa no encontrada**:
  - el código es el mismo *(ningún `.go` cambia entre `7127831` y `cabf8c5`)*, y la toolchain también *(go1.22.2, `/usr/lib/go-1.22`)*;
  - `GOFLAGS`, `GOEXPERIMENT` y `GOARM64` están vacíos, y `GOAMD64` vale `v1`.
- ⇒ **SC-016 se mide en C1 recompilando las dos versiones en la misma sesión**, `git archive cabf8c5` contra `git archive HEAD`, **y no contra esta tabla**.
  La tabla sirve de orden de magnitud.

### T002 · Testdata y la referencia de SC-012

- **Nace `cmd/permea/testdata/lista/generar.sh`** `96304965…` *(bash + `python3 -I`)*: genera `plantillas/`, **10 ficheros sintéticos**, un hogar con sus
  carpetas sin el punto inicial y `@CASA@` en cada `cwd`. **Seis lectores**:
  - Claude Code, 4 ficheros y 9 líneas, con 2 repetidas;
  - Codex, 1 sesión;
  - Gemini, una sesión y su `.project_root`;
  - Copilot CLI, 1 sesión;
  - Kimi Code, un `wire.jsonl` y `session_index.jsonl`;
  - OpenCode, no generado: se copia `cmd/permea/testdata/opencode/sana.db`.

  Ids `m-lista-…`, `r-lista-…`, `s-lista-…` y `l-lista-…`. Ninguno es `msg_` ni `ses_`, ni tiene forma UUID. md5, en el orden de `find | sort` sobre `plantillas/`:

  | Fichero | md5 |
  |---|---|
  | `claude/…/proyecto-ejemplo/s-lista-claude-1.jsonl` · `-2.jsonl` | `c1a53e84` · `46177c09` |
  | `claude/…/pruebas-tmp-1/s-lista-claude-3.jsonl` · `claude/…/viejo/s-lista-claude-4.jsonl` | `131be6a6` · `bed96ff9` |
  | `codex/…/rollout-lista-1.jsonl` · `copilot/…/events.jsonl` | `6486d2cc` · `69123941` |
  | `gemini/…/.project_root` · `gemini/…/s-lista-gemini.jsonl` | `ff50ee18` · `44fff7b9` |
  | `kimi-code/session_index.jsonl` · `kimi-code/…/wire.jsonl` | `4e20914c` · `8e80e80b` |

- **La referencia de SC-012** *(D-013-P6)* sale del **binario de `cabf8c5`**, con el patrón de la T026 de la 012: un test temporal que **sólo vivió en la copia
  `git archive`** del scratchpad *(`TestZZReferenciaLista`; libre 1.26)*.
  - En `testutil.SandboxConSemillas`, copia `plantillas/` al hogar con `@CASA@` → `/sintetico/casa` *(plan §9)* y `sana.db` a `~/.local/share/opencode/opencode.db`.
  - Ejecuta `--run`: rc 0.
  - Vuelca la cola, el stdout y el stderr, con la ruta de datos por `<DATOS>`.
- **Repetida** *(libre otra vez)*: los tres ficheros, **byte a byte iguales** ⇒ la referencia es determinista.

  | `cmd/permea/testdata/lista/` | md5 | Contenido |
  |---|---|---|
  | `referencia-run.cola` | `e6b011db77b995046c72356fa2f796ab` | 17 eventos: claude_code 7, codex 2, gemini 2, opencode 2, copilot_cli 2, kimi_code 2 |
  | `referencia-run.stdout` | `d41d8cd98f00b204e9800998ecf8427e` | vacío |
  | `referencia-run.stderr` | `1799737c95a64d0f7ba63d84bd3560df` | 11 líneas: el banner, «17 eventos encolados en `<DATOS>/queue.jsonl`», las dos de la pasada, una por lector y «sync omitido» |

- **La cola, sin rutas.** «sintetico» sólo aparece en `model` *(`modelo-sintetico…`)*. Los `project_ref` son hashes de `/sintetico/casa/…`, con la sal de las
  semillas. Hay **7** identidades distintas: Claude Code da 5, porque en la referencia `src` y `docs` no tienen `.git` y no suben a la raíz; Codex comparte
  la de `src`; Kimi Code, la de la raíz; Gemini y Copilot CLI, una; y OpenCode, la suya.
- **`PRIV`** → **0**:
  - en `git diff cabf8c5`;
  - en los ficheros sin versionar;
  - y con `grep -raEc` sobre `cmd/permea/testdata/lista/`, los 13 ficheros, `generar.sh` incluido.
- El temporal, con el test, los binarios y las copias, se **borró** al acabar.

### Corte de B0 *(sobre el árbol de trabajo, con el testdata nuevo)*
- Suite *(libre 1.14; 13:20:19 → :27)*: **1074 / 0 / 0**, rc 0, y **los mismos 1074 nombres** que la de `cabf8c5`.
- Vet limpio, lint 0 *(libre 0.96)* y gofmt vacío. Frontera, 0 B; ningún `_test.go` modificado; `PRIV` 0.

### Incidencias de B0
1. **`golangci-lint` 2.15.0 en vez de 2.12.2.** Da 0 issues sobre `cabf8c5`. No PARA, porque ninguna regla de `tasks.md` fija la versión. Las puertas siguientes
   usan la instalada.
2. **Los tamaños no cuadran con la 012** *(arriba)*. No PARA: SC-016 compara las dos versiones medidas en la misma sesión.

**Sin mutaciones en B0.**

## B1 · Leer sin escribir

## B2 · La derivación con origen

## B3 · El recorrido común

## B4 · Agregación y nombres

## B5 · La tabla

## B6 · El comando

## B7 · Los seis lectores y la frontera

## B8 · README y CHANGELOG

## Cierre
