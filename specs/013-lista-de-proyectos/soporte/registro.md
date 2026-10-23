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

## B1 · Leer sin escribir *(`internal/config`; sobre `6b318c5`)* — PARA 2 en M-B1a, **cerrada**

- **T004** · Esqueletos: `RutaDataDir()` → `("", nil)`, en `config.go`; `LeerSalt(dir)` → `("", false, nil)`, en `identity.go`. Compila.
- **T005** · **Rojos (1) y (2)**, en `internal/config/lectura_test.go`, con `testutil.Sandbox`. La razón se escribió **antes** de correrlos. Se esperaba
  **1077 / 7 / 0** y salió **1077 / 7 / 0** *(libre 0.29; 13:26:19 → :25)*.
  - **Caen, por lo que prueban**:
    - (1/misma_ruta) y (1/no_crea): `RutaDataDir` devuelve `""`, no la ruta;
    - (2/igual) y (2/igual_con_salto): `LeerSalt` devuelve `""`, no la sal;
    - (2/ilegible): no devuelve error.
  - **Nacen verdes, declaradas**: (1/datadir_crea), porque `DataDir` no cambia *(la valida M-B1a)*; (2/ausente) y (2/ausente_con_datadir), porque el esqueleto
    ya responde lo correcto *(las valida m2)*.
- **T006** · **Verde**, según plan §3:
  - `RutaDataDir` = `os.UserConfigDir()` + `permea`, sin `MkdirAll`; `DataDir` = `RutaDataDir` + `MkdirAll`;
  - `LeerSalt` lee como `loadOrCreateSecret`: tal cual, `existe = false` si no existe, y error en cualquier otro fallo.

  Suite *(libre 0.14; 13:27:04 → :10)*: **1084 / 0 / 0**. md5: `config.go` `19b0d970…`, `identity.go` `5ca4478e…` y `lectura_test.go` `317d7102…`.
- **T007** · Mutaciones, con el censo declarado **antes**; cada una con la suite entera tras `libre.sh` y la reversión por edición inversa con md5:

  | Mutación | Declarado | Cayó | md5 tras revertir |
  |---|---|---|---|
  | **m1** `RutaDataDir` con `MkdirAll` | `TestRutaDataDir`, `/no_crea` | **lo mismo** *(1082 / 2)* | `config.go` `19b0d970…` ✅ |
  | **m2** `LeerSalt` delega en `LoadOrCreateSalt` | `TestLeerSalt`, `/ausente`, `/ausente_con_datadir` | **lo mismo** *(1081 / 3)* | `identity.go` `5ca4478e…` ✅ |
  | **m3** `strings.TrimSpace` sobre la sal | `TestLeerSalt`, `/igual_con_salto` | **lo mismo** *(1082 / 2)* | `identity.go` `5ca4478e…` ✅ |
  | **M-B1a**, 1.ª forma: **PARA 2 por censo incompleto** · `DataDir` sin `MkdirAll` | `TestRutaDataDir`, `/datadir_crea` | eso **y además** `TestEnroll_ReenrollNoResidue`, `TestEnroll_Argv_HappyPath` y `TestEnroll_Stdin_And_SC011` *(1078 / 5)* | puesta hasta la decisión; revertida después: `config.go` `19b0d970…` ✅ |
  | **M-B1a**, 2.ª forma *(censo corregido; orquestador, 23-10, opción 1)* | `TestRutaDataDir`, `/datadir_crea`, **más** los tres `TestEnroll_*` | **lo mismo** *(1078 / 5; `diff` del declarado contra lo caído, vacío)* | `config.go` `19b0d970…` ✅ |

- ⛔ **PARA 2.** M-B1a tumba tres tests de `enroll` que el censo no declaró. **La mutación se queda puesta.**
  - **Por qué caen**: `enroll_test.go:65`, `:115` y `enroll_reject_test.go:260`. Esos tests apuntan `XDG_CONFIG_HOME` a un temporal sin `permea/`, y `enroll`
    escribe allí con `atomicWrite`: «no such file or directory».
  - **Es un error del censo, no del código.** El censo sólo buscó los tests que borran el `DataDir`. Esos tres sí prueban que `DataDir` crea, así que la
    conducta de hoy ya estaba cubierta por tests existentes, como decía el encargo.
- **Cierre** *(decisión del orquestador, 23-10: opción 1)*:
  - se revirtió M-B1a por edición inversa *(→ `19b0d970…`)*;
  - se corrigió su censo: los tres `TestEnroll_*` **escriben en el `DataDir` de un `XDG_CONFIG_HOME` vacío**, y sin el `MkdirAll` no pueden;
  - se repitió, y cayó **exactamente** lo declarado.
- **El censo, por los llamantes**: `DataDir()` tiene cuatro en producción, `setup()` *(`main.go:207`)*, `enroll` *(`enroll.go:69`)*, `status` *(`status.go:24`)* y
  `project join` *(`project.go:163`)*. Sólo los tres `TestEnroll_*` los ejercen con un `XDG_CONFIG_HOME` sin `permea/`. Los demás corren en `testutil.Sandbox`,
  que lo crea.
- 📌 **La lección**: el censo de una mutación sobre una función **compartida** se hace buscando **todos sus llamantes**, y los tests que ejercen cada uno, **no
  sólo los tests que parecen relacionados**.
- **Corte de B1** *(sobre el verde)*:
  - suite *(libre 0.19; 13:34:09 → :15)*: **1084 / 0 / 0**;
  - gofmt vacío, vet limpio y lint 0 *(libre 0.24)*;
  - frontera 0 B, ningún `_test.go` existente modificado, `go.mod` sin `require` y sin `go.sum`;
  - `PRIV` 0 en el diff, en `lectura_test.go` y, con `-a`, en el testdata.

## B2 · La derivación con origen

## B3 · El recorrido común

## B4 · Agregación y nombres

## B5 · La tabla

## B6 · El comando

## B7 · Los seis lectores y la frontera

## B8 · README y CHANGELOG

## Cierre
