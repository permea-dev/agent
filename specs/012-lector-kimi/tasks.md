# Tasks: 012 · «Lector de Kimi Code CLI»

**Feature**: `012-lector-kimi` · **Fecha**: 2026-10-15 · [spec.md](./spec.md) *(ratificada el 2026-10-15; E-1)* · [plan.md](./plan.md) ·
[contracts/event-id-kimi.md](./contracts/event-id-kimi.md)
**Base**: el commit de la spec ratificada *(el código, el de `e9e5449`)*. **Línea base a preservar** *(sobre `e9e5449`, se transcribe en T001)*:
`go test -count=1 -v ./...` → **953** «--- PASS», 0 FAIL, 0 SKIP · `go vet` limpio · `golangci-lint run` → **0**. Los tamaños de la 0.8.0, en T001.
**51 tareas · 13 ✋ · 43 tests nuevos** *(33 rojos y 10 nacidos verdes)* **· 55 mutaciones** *(plan §Bloques)*.
*Unidades*: aquí, «tests» son los rojos **numerados** del plan, (1)–(43). El registro cuenta **hojas**.

## Formato: `[ID] [P?] [✋?] Descripción`

- **[P]**: sin dependencia de decisión entre sí. No quiere decir que se editen a la vez.
- **✋**: la ejecuta **el dueño**: commits, PR, fusión, etiqueta, la plataforma, aprobación de textos y ensayos en Windows. Claude no hace git de escritura.
  Cada commit lleva su mensaje previsto, **sin tildes ni ñ**.
- **(1)…(43)** son los rojos, y **m1…m54** y **M-B4a** las mutaciones de plan §Bloques. Lo que se añada aquí se llama **M-B2a**, **M-B3a**…
- **`[MUTACIÓN]`** marca las tareas que mutan. **No se renumera**: una tarea añadida después recibe el siguiente número libre.
- **Las transcripciones** van a `soporte/registro.md`, que **nace en T001**, y cada tarea deja una remisión.

## Presupuesto *(techos fijados por el dueño ANTES de la fase; ninguno se sube: si no cabe, se condensa)*

| Fichero de `specs/012-lector-kimi/` | Techo | Hoy |
|---|---:|---:|
| `spec.md` | 340 | 326 |
| `plan.md` | **250** | 249 |
| `tasks.md` *(este)* | **350** | 287 |
| `contracts/event-id-kimi.md` | **80** | 79 |
| `soporte/registro.md` *(nace en T001; §B0–§B2 en `registro-archivo.md`, sin techo)* | **300** | 169 |
| `soporte/descubrimiento.md` · `-muestra-1.md` · `-muestra-2.md` | 200 · 100 · 60 *(cerrados)* | 200 · 100 · 60 |

| Bloque | Producción | Test | Fixtures | Registro |
|---|---:|---:|---|---:|
| B1 | ~60 | ~110 | `generar.sh` *(nace)*, `raiz/` | ≤ 30 |
| B2 | ~150 | ~300 | `lineas/*.jsonl` | ≤ 45 |
| B3 | ~230 | ~420 | `ficheros/**`, índice, `esperado.json` | ≤ 60 |
| B4 | ~90 | ~300 | la referencia de SC-015, `cmd/permea/testdata/kimi/` | ≤ 45 |
| B5 + B6 | ~60 + ~50 de texto | ~120 | — | ≤ 30 |
| Cierre C1–C9 | — | — | — | ≤ 70 |

**~590 de producción y ~1 250 de test.** Una desviación de más del 50 % se declara en el corte. **SC-022** *(≤ 100 KB)* manda sobre la estimación.

## Una suite a la vez, con la carga vigilada

Antes de **cada** `go test` o `golangci-lint` *(puertas, rojos, verdes y mutaciones)*, se lanza **`~/bin/libre.sh` en una orden sola**: sin nada más en esa línea
de órdenes, y **sin el nombre de ninguna suite en ella** *(011 registro §B0, incidencia)*. Espera:
- 0 procesos de suite, **del agente o de la plataforma**;
- carga de 1 min **< U = 2,0** en 3 muestras seguidas, cada 10 s;
- tope de 45 min.

Mientras dice «no libre», se espera; al tope, **PARA 13**. El guion es el de la plataforma, en **`~/bin/libre.sh`** *(md5 `ab647c01…`)*, y **no se versiona**.

## 🛑 Lo que PARA

1. **Un `_test.go` existente** necesita cambiar *(censo: ninguno)*, o se toca un fichero fuera de plan §Project Structure.
2. **En una mutación cae algo que no se declaró**, o no cae lo declarado. **La mutación se queda puesta** y se para.
3. **El md5 de la reversión** no es el de antes al empezar la mutación siguiente.
4. **La frontera de FR-027 tiene diff**, o `os.Getenv` de producción ≠ 6 tras B1.
5. **El lector escribe algo**: un md5 o un listado distinto en (27), en (36) o en W1, o un fichero nuevo bajo la raíz.
6. **Un `require` en `go.mod`**, o un `go.sum` *(SC-023)*.
7. **SC-022 pasa de 102 400 B** en alguna plataforma en C1. En los cortes de B2, B3 y B4 es informativo, y si pasa se adelgaza antes del bloque siguiente.
8. **Un dato real** en testdata, en el registro o en un informe: una línea, o un trozo, de una copia congelada; un id con forma UUID; una ruta personal; un
   prompt, una respuesta o un nombre de proyecto. Nada que no salga de `generar.sh`, y nada que case con `PRIV`, **tampoco con `grep -a`** sobre `testdata/`.
9. **Una copia congelada cambia su sha256.**
10. **Se ejecuta `kimi`**, o un `--run` o `--daemon` sobre la instalación real fuera de W1/W2. **Se lee `C:\permea-prueba-kimi\home\`** fuera de W1.
11. **El lector decodifica una línea** que no empieza por `{"type":"usage.record"` o `{"type":"forked"`, o lee un campo fuera de FR-012.
12. **Un test que abre un testdata en su sitio**: siempre una copia en `t.TempDir()`.
13. **`~/bin/libre.sh` llega a su tope**, o su md5 no es `ab647c01…`.
14. **El cuadre de W1 o W2** *(SC-025)* tiene un evento sin fila en la consola, o una fila con dos eventos.
15. **C7 sin C6b**: la etiqueta antes de que `esUtilizable("kimi-k2.7-code")` dé `true` en producción.

## Puertas en cada corte

```sh
gofmt -l . ; go vet ./... ; golangci-lint run          # → vacío, limpio, 0 issues
go test -count=1 ./...                                 # → 953 + los nuevos, 0 FAIL, 0 SKIP; UNA suite a la vez, tras ~/bin/libre.sh en una orden sola
git diff e9e5449 -- internal/event/ internal/ingest/eventid.go internal/ingest/eventid_test.go internal/ingest/boundary_test.go \
  internal/ingest/codex_eventid.go internal/ingest/codex_eventid_test.go internal/ingest/gemini_eventid.go internal/ingest/gemini_eventid_test.go \
  internal/ingest/opencode_eventid.go internal/ingest/opencode_eventid_test.go internal/ingest/copilot_eventid.go internal/ingest/copilot_eventid_test.go \
  specs/006-medicion-fiel/contracts/ specs/008-lector-codex/contracts/ specs/009-lector-gemini/contracts/ specs/010-lector-opencode/contracts/ \
  specs/011-lector-copilot/contracts/                                                      # → vacío
git diff e9e5449 --name-only --diff-filter=M -- '*_test.go'                                # → vacío
grep -c require go.mod ; ls go.sum                                                         # → 0 ; no existe
```

**Vigilancia de privacidad** *(cada corte con testdata, PARA 8)*. El patrón se escribe con clases *(`AI[z]a`…)* para no detectarse a sí mismo, y **el usuario de la
máquina y el nombre del dueño salen del entorno**: nunca se escriben en el repo.
```sh
PRIV='@[a-z0-9.-]+\.[a-z]{2,}|dev_[A-Za-z0-9]{8,}|ses_[A-Za-z0-9]{6,}|msg_[A-Za-z0-9]{6,}|/hom[e]/|C:\\User[s]|AI[z]a|sk-[a]nt|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}'
PRIV="$PRIV|${USER%?}|$(git config user.name)"
git diff e9e5449 | grep -cE "$PRIV"                                                        # → 0
grep -raEc "$PRIV" internal/ingest/testdata/kimi/ cmd/permea/testdata/kimi/                # → 0 en cada fichero, generar.sh incluido
```

**Protocolo de mutación** *(006, con E-5 de 007)*:
1. El censo se declara **en la tarea, ANTES de mutar**, por **hoja**, con las co-caídas.
2. Se muta y se ejecuta la suite entera, tras `~/bin/libre.sh`.
3. Si los `FAIL` coinciden con lo declarado, se revierte **por edición inversa**, se comprueba el **md5** y se transcribe.
4. Si no coinciden, PARA 2.

Una mutación que no compila o que panica **no cuenta**. **Rojo antes de verde**, con la razón real transcrita. **Un test que nace verde lleva su mutación.**

---

## B0 · FASE 0 y registro

- [x] **T001** Fase 0, sobre `e9e5449` por `git archive` en un temporal: la línea base *(953, vet, lint 0)*, los **tamaños de la 0.8.0** en las 5 plataformas
  *(D-012-P7)* y la frontera de FR-027 sin diff. **Nace `soporte/registro.md`**, con su techo *(300)* y las secciones §B0…§B6 y §Cierre. Si algo difiere, **PARA**.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B0 · T001 *(sobre `3779671`, sólo documentos desde `e9e5449`; 953 / 0 / 0 en 10 paquetes; vet y lint 0; frontera 0; `PRIV` 0; tamaños de la 0.8.0 por `git archive`, sin sellado VCS)*.
- [ ] **T002** ✋ **Línea base y registro**. Commit: `012 B0: linea base y registro`.

## B1 · Raíz y ficheros *(FR-001 a FR-003; SC-012; M-3, M-5)* — nacen `internal/config/kimi_test.go`, `sandbox_kimi_test.go` y el listado en `kimi_fichero_test.go`

- [x] **T003** Fase 0: `config.KimiHome()`, que devuelve `"", nil`, e `ingest.ListarKimi`, que devuelve `nil, nil`.
- [x] **T004** Testdata: **nace** `internal/ingest/testdata/kimi/generar.sh` *(plan §6; `bash` + `python3 -I`)* con `raiz/`: dos sesiones en una `wd_*`, un
  `agents/agent-0/`, `sessions/.index-cache/` y `.index-dirty/`, un `wire.jsonl.bak` y un `wire.jsonl.tmp.1.ab`, `raiz/imported/<id>/agents/main/wire.jsonl`, un
  `wire.jsonl` un nivel más hondo y una sesión sin `agents/`. md5 al registro.
- [x] **T005** [P] **Rojos (1) a (3)**, con `t.Setenv`.
- [x] **T006** **Verde**: P-1 en `internal/config/kimi.go`; `ListarKimi` de plan §2; tras `sandbox.go:70`, `t.Setenv("KIMI_CODE_HOME", "")`; los comentarios de
  `codex.go:16`, `gemini.go:17`, `opencode.go:18` y `copilot.go:16`, «una de las SEIS».
- [x] **T007** Comprobar y transcribir, sin tocarlo, que `sandbox_test.go` sigue verde, y que `os.Getenv` de producción → **6**.
- [x] **T008** `[MUTACIÓN]` Censo y mutaciones **m1–m8**, con md5.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 *(rojos 953 / 14 / 0 con su razón escrita antes; verde 967 / 0 / 0; `os.Getenv` → 6; m1–m8 = declarado, con m2 → también (1/vacia) y (3/forma) en m5–m8 declaradas en el censo; m4 restaurada a mano y m5 reformulada para compilar)*.
- [ ] **T009** ✋ Puertas y commit: `012 B1: raiz de Kimi Code con KIMI_CODE_HOME`.

## B2 · Una línea e identidad *(FR-010 a FR-017, FR-020; SC-005, SC-007, SC-008, SC-011, SC-021)* — nacen `kimi_eventid_test.go` y `kimi_registro_test.go`

- [x] **T010** Fase 0: `derivarEventIDKimi(S, A string, N int) string`, que devuelve `""`; `ContextoKimi`, `ClaseKimi` y `RegistroKimi`, que aún no
  clasifica, en `internal/ingest/kimi.go`.
- [x] **T011** Testdata: `generar.sh` gana `lineas/`:
  - una línea por clase *(evento, compactación, sin modelo, a cero, las siete incoherentes)*;
  - las partidas de SC-005; los seis modelos de SC-008;
  - una con centinelas en todos los campos de texto *(SC-011)*;
  - un `context.append_message` con `"type":"usage.record"` **dentro** de un texto, una `forked` y un `subagent.completed` *(12)*.

  md5 al registro.
- [x] **T012** [P] **Rojos (4) a (13)**: cada uno afirma **sólo su campo**; las 7 hojas de (4), el valor exacto del contrato. (13) cae porque el stub no emite
  *(rojo antes del verde)*. **Nace verde** con el stub, y se declara así, **(12)** *(lo valida m22)*.
- [x] **T013** **Verde**: plan §1. Reutiliza `dominioEventID`, `bytesEventID` y `tipoRespuesta` *(`eventid.go:31`, `:43`; `codex_eventid.go:29`)*; la
  codificación, **replicada** de `copilot_eventid.go:33-52`. **No** llama a `internal/pricing`.
- [x] **T014** **(13)**, rojo en T012, pasa a verde con T013: el evento no contiene ningún centinela. Lo valida m23.
- [x] **T015** `[MUTACIÓN]` Censo y mutaciones **m9–m24**, con md5, con las co-caídas de plan §Bloques: m9 → (4) ×7; m11 → (4) ×6 *(no la hoja
  `ab|c`, que también cambia: se declara al censar)*; m13 y m14 → (5); m24 → (10/compactacion).
- [x] **T016** **SC-022, informativo tras B2**: las 5 plataformas, D-012-P7 *(R-1)*. Si pasa de 102 400 B, se adelgaza antes de B3 *(PARA 7)*.
- [x] **T017** **Vigilancia de privacidad** sobre `lineas/` *(§Puertas)* → 0, transcrito.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B2 *(rojos 971 / 33 / 0 con su razón escrita antes, (12) nacido verde; verde 1004 / 0 / 0; SC-022 tras B2: Δ 0 en las 5, `__TEXT` darwin/amd64 1 004 B; m9–m24 = declarado, con m10–m12 → (4) ×7 y m9 sólo en el hash, declaradas en el censo)*.
- [ ] **T018** ✋ Puertas y commit: `012 B2: un registro de Kimi Code es un evento sin coste`.

## B3 · Fichero, cabeza e índice *(FR-004, FR-005, FR-007 a FR-009, FR-014, FR-018, FR-020; SC-004, SC-006, SC-007, SC-009, SC-010, SC-014, SC-026)* — `kimi_fichero_test.go`

- [x] **T019** Fase 0:
  - en `internal/state/state.go`, `Cabeza string \`json:"cabeza,omitempty"\`` tras `Frontera` *(D-012-P3)*, **sin tocar** `Recorrer`, `Load` ni `Save`;
  - `PasadaKimi` *(`Resumen`, `HayNovedades`)*, `LeerFicheroKimi`, que no emite, `LeerIndiceKimi`, que devuelve un mapa vacío, y `huellaCabezaKimi`,
    que devuelve `""`, en `internal/ingest/kimi_contexto.go`.
- [x] **T020** Testdata: `generar.sh` gana `ficheros/` y un índice:
  - `dos_registros.jsonl`; `fork/{un_forked,dos_forked,sin_forked}.jsonl`; `fork_cortado/{paso1,paso2}.jsonl` *(corte tras la `forked`)*;
  - `corrupta_cuenta.jsonl` *(una corrupta con prefijo entre dos registros)*;
  - `cabeza/{antes,cambiada,encogido}.jsonl` *(SC-026: `cambiada`, más larga y con otra 1.ª línea; `encogido`, el prefijo de `antes`)*;
  - `sin_lf.jsonl` y `sin_lf_completo.jsonl`; `matada.jsonl` *(arranque, `turn.prompt` y una `llm.request`, sin registro)*;
  - `subagente/{main,agent-0}.jsonl` *(con `subagent.completed.usage` = Σ del subagente)*; `clases.jsonl` *(SC-010)*;
  - `session_index.jsonl` *(una ruta `C:\sintetico\proyecto-a` y la misma con `/`; dos líneas para una S)*;
  - `esperado.json`.

  md5 al registro.
- [x] **T021** [P] **Rojos (14) a (29)**, cada pasada con su **línea de resumen literal**; (18) con un `state.json` sintético sin `cabeza` para `sin_cabeza`;
  (19) con `os.Chtimes`. **Nace verde**, y se declara así, **(27)** *(el stub no escribe; lo valida m39)*.
- [x] **T022** **Verde**: plan §2 *(atajo, cabeza, prepaso, `Recorrer`, la cabeza tras `Recorrer`)*, el índice con la normalización de barras *(D-012-P6)*, y
  `Resumen()` con el texto de spec §Textos.
- [x] **T023** **SC-022, informativo**: las 5 plataformas. Si pasa de 102 400 B, se adelgaza antes de B4 *(PARA 7)*.
- [x] **T024** `[MUTACIÓN]` Censo y mutaciones **m25–m40**, con md5, con las co-caídas de plan §Bloques: m25 → (15) ×3, (24), (25); m27 → (15/un_forked),
  (16); m28 → (17), (18/cabeza_cambiada); m29 → (16), (17/dos_pasadas); m37 → (23), (24). Vigilancia de privacidad sobre `ficheros/` → 0.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B3 *(rojos 1010 / 32 / 0 con su razón escrita antes, nacidos verdes (27), (29) ×3 y (19/igual); verde 1042 / 0 / 0; SC-022 tras B3: el campo `Cabeza` cuesta +512 B en Windows y 321 B de holgura de `__TEXT` en darwin/amd64, medido; R-2: 10 MB en 126 ms; m25–m40 = declarado, con los ajustes del censo: m28 y m29 → (16), (17/dos_pasadas); m38 → (17) ×2; m40 → también `TestRetencion_…/state_json_cuatro_campos`)*.
- [ ] **T025** ✋ Puertas y commit: `012 B3: fichero de Kimi Code por offset y cabeza, con frontera forked e indice`.

## B4 · Integración *(FR-002, FR-021, FR-023 a FR-026; SC-011, SC-013, SC-015, SC-018, SC-019)* — nace `cmd/permea/kimi_test.go`

- [x] **T026** Fase 0, **en este orden**:
  1. **Antes de tocar `main.go`** *(D-012-P9)*: la referencia de SC-015 con el binario de `e9e5449`, y su md5. **Si `main.go` ya cambió, PARA.**
  2. El forzado A, comprobado *(`Load` pasa, `Append` pasa, `Save` falla)*.
  3. `cmd/permea/kimi.go` *(`generarKimi`, sin emitir)*, y en `main.go` los campos, `setup()` y `generate()`.
- [x] **T027** [P] **Rojos (31) a (34)**, y **(30)**, nacido verde *(lo valida m41)*, con `cola` y `capturarStderr` de los tests de `cmd/permea`, sin tocarlos.
- [x] **T028** **Verde**: Kimi Code tras Copilot CLI y antes del único `st.Save`; la activación en cada pasada; el índice primero *(P-13)*; los errores por fichero
  *(D-012-P5)*; la línea en `runOnce` y, con `HayNovedades`, en `tick`.
- [x] **T029** Comprobar y transcribir, sin tocarlos, que los `_test.go` de `cmd/permea` siguen verdes.
- [x] **T030** **Nacen verdes (35) a (39)**, escritos tras el verde de T028: centinelas en la cola y en `state.json`; md5 y listado de la raíz; `sessions/` creada
  después; el `wire.jsonl` ilegible; y el índice ilegible. Los validan m49, **M-B4a**, m45, m46–m47 y m48.
- [x] **T031** `[MUTACIÓN]` Censo y mutaciones **m41–m49** y **M-B4a**, con md5. **m41** declara, **por nombre y hoja**, los tests existentes que caen *(los
  `*_SinRaiz*` y `*_LineaDeResumen` de Codex, Gemini, OpenCode y Copilot)*, y se comprueba antes de mutar que no hay más.
- [x] **T032** **SC-022, informativo**, como T023.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B4 *(referencia de SC-015 con el binario de `e9e5449`, `44a96b93…`, 9 líneas; forzado A comprobado; rojos 1045 / 7 / 0 con su razón escrita antes, (30) ×2 nacido verde; verde 1052 / 0 / 0; nacidos verdes (35)–(39), 1061 / 0 / 0; m41–m49 y M-B4a = declarado, con «E» por nombre y los ajustes del censo: m42 → también (34), (35), (38/se_relee); m46 → también (38/se_relee); m45 formulada para compilar; SC-022 tras B4: +30 208 B como mucho, Windows; `__TEXT` de darwin/amd64 saltó, holgura 2 138 B)*.
- [ ] **T033** ✋ Puertas y commit: `012 B4: Kimi Code en run y daemon, antes de guardar el estado`.

## B5 · `--scan` *(FR-022, FR-023; SC-024 en forma)*

- [x] **T034** [P] **Rojos (40) y (43)**, y **(41)** y **(42)**, nacidos verdes *(los validan m50 y m52)*, en `cmd/permea/kimi_test.go`, con los textos de spec
  §Textos **leídos por programa**. (43) afirma el vector 1 del contrato con un fichero en `<tmp>/session_sintetica-0001/agents/main/wire.jsonl`.
- [x] **T035** **Verde**: `esSesionKimi`, **tras** Copilot, y `dryRunKimi` *(plan §5)*.
- [x] **T036** `[MUTACIÓN]` Censo y mutaciones **m50–m54**, con md5. **m50** declara, por nombre y hoja, los tests existentes de `--scan` que caen.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B5 *(§B0–§B2 archivados íntegros en `registro-archivo.md`, `cmp` idénticos; rojos 1069 / 4 / 0 con su razón escrita antes, (40/disco), (41) ×5 y (42) nacidos verdes; verde 1073 / 0 / 0; (40/subagente) añadida tras el verde por SC-024, 1074 / 0 / 0; m50–m54 = declarado, con los ajustes del censo: m50 → sólo (41/claude) y los `--scan` de Claude Code por nombre; m53 y m54 en el código compartido, m54 = la edición de m25; SC-022 tras B5: +33 792 B como mucho, Windows)*.
- [ ] **T037** ✋ Puertas y commit: `012 B5: scan de Kimi Code con la frontera forked`.

## B6 · README y CHANGELOG *(FR-028; P-11)*

- [x] **T038** ✋ **El dueño aprueba los textos**: «### Kimi Code» en el README y la entrada `0.9.0` del CHANGELOG. Claude los propone en un encargo aparte, con
  los límites de spec §Límites: la petición matada *(que sí se cobra)*, los intentos fallidos, la ventana del fork, las mayúsculas en Windows, y «medido en Windows con
  la 2.2.0 y clave de API»; y la dependencia de la tarifa de `kimi-k2.7-code`.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B6 *(la propuesta v1, tal cual: «### Kimi Code CLI», «### Requiere», R1–R8; `aprobado-v1.md` `5a435d3f`)*.
- [x] **T039** **Rojo**, transcrito: `grep -c '^## 0.9.0' CHANGELOG.md` → 0; `grep -c '^### Kimi Code' README.md` → 0.
  > Transcripción: §B6 *(0 y 0)*.
- [x] **T040** README y CHANGELOG *(`## 0.9.0 — PENDIENTE`)*, con los textos **sacados por programa** de lo aprobado. `cmp` sin diferencias.
  > Transcripción: §B6 *(10 `cmp` sin diferencias; fuera de los bloques, nada cambió; `PRIV` 0)*.
- [ ] **T041** ✋ Puertas y commit: `012 B6: README y CHANGELOG de la 0.9.0`.

## Cierre — en tramos, uno por mensaje *(plan §Cierre; si uno falla, se para y se rehace desde C1)*

- [ ] **T042** **C1** · Puertas, con 0 SKIP en Linux: **SC-022** en las 5 plataformas; **SC-023** y **SC-016**; **SC-020** *(`PRIV` sobre `git diff e9e5449` y,
  con `grep -raE`, sobre los testdata)*; compilan Windows y darwin; `PENDIENTE` → 1. **Los 953 de la 0.8.0, por nombre**, en PASS.
- [ ] **T043** **C2** · Medidas sobre **las copias A y B**, en sólo lectura y con sus sha256 antes y después *(PARA 9)*. `--run` con `KIMI_CODE_HOME` = **una copia
  de la copia** en un `mktemp -d`, en sandbox, sin enrolar ni `endpoint`:
  - **SC-001** sobre B: 12 eventos, 33 831 · 186 112 · 0 · 3 170, el resumen literal, 4 `session_ref` y 1 `project_ref`;
  - **SC-002**: A y luego B sobre el mismo estado: 11 y 1, y 0 aperturas de lo que no cambió;
  - **SC-003**: Σ × la tarifa = 0,08018073 $; **SC-004**: el fichero de S6, 0 eventos; **SC-024**: `--scan` del fork y de `agent-0`.

  Contraste independiente: un guion `python3 -I` fuera del repo cuenta registros, heredados y partidas. Una diferencia → **PARA**.
- [ ] **T044** **C3** · `goreleaser release --snapshot --clean`, el SHA-256 del zip de Windows, y `strings` con los textos *(SC-017)*. `dist/` borrado.
- [ ] **T045** ✋ **C4 · W1** *(plan D-012-W)*: en Windows, en sandbox y sin enrolar. Kimi cerrado *(SC-013)*, y luego abierto con `--daemon`: dos prompts, un
  subagente, un `/compact`, un `/fork` con un prompt en el fork y una petición matada. **SC-025**: el cuadre petición a petición contra «Request Details»; cada
  fila sin evento, declarada como matada con su hora. Observar Q-3 *(¿avisa Kimi al reescribir con el agente leyendo?)*.
- [ ] **T046** **C5** · El cuerpo del PR, fuera del repo; la fecha de la 0.9.0 entra en la rama *(`PENDIENTE` → 0)*.
- [ ] **T047** ✋ **C6** · Fusión del PR con merge commit.
- [ ] **T048** ✋ **C6b** · **La plataforma**: la fila `kimi-k2.7-code` *(0,95 · 0,19 · 4,00 USD por millón)* desplegada **en producción**, en su parche aparte, y
  comprobada allí con `CatalogoTarifas::cargar()->esUtilizable("kimi-k2.7-code")` → **`true`**. Transcrita con la fecha. **Sin ella, PARA 15.**
- [ ] **T049** ✋ **C7** · `git tag -a v0.9.0` y push; el workflow `release` en verde. **Sólo tras T048.**
- [ ] **T050** **C8** · Los tres canales en `0.9.0`, `sha256sum -c`, y `strings` del binario publicado.
- [ ] **T051** ✋ **C9 · W2**: `scoop update`, `--run` en la instalación real. **SC-025** otra vez contra «Request Details», y **evento a evento** el coste que
  calcula la plataforma = la tarifa oficial × sus partidas.

## Dependencias

- **Bloques**: B0 → B1 → B2 → B3 → B4 → B5 → B6 → C1…C9. B3 usa B2; B4 usa B1 y B3; B5 usa B3.
- **Dentro de un bloque**: los rojos `[P]` antes del verde, y el censo antes de mutar.
- **Fuera del repo**: **T048** *(la fila de tarifa en producción)* antes de **T049** *(la etiqueta)*. El parche de la plataforma puede ir en paralelo a B1–C6.

## Censo de tests *(spec §Censo; ninguno existente cambia)*

| Tarea | Nace |
|---|---|
| T005 | `internal/config/kimi_test.go`, `internal/testutil/sandbox_kimi_test.go`, `internal/ingest/kimi_fichero_test.go` *(el listado)* |
| T012 | `internal/ingest/kimi_eventid_test.go`, `kimi_registro_test.go` |
| T021 | `internal/ingest/kimi_fichero_test.go` *(fichero, cabeza e índice)* |
| T027, T030, T034 | `cmd/permea/kimi_test.go` |

**Fuera de `_test.go`**: `internal/testutil/sandbox.go`, una línea, en T006; `internal/state/state.go`, un campo, en T019.

## Cobertura *(SC → tarea que lo acredita)*

| SC | Tareas | | SC | Tareas | | SC | Tareas |
|---|---|---|---|---|---|---|---|
| 001 | **T043** | | 010 | T021 (24) | | 019 | T027 (32) |
| 002 | T021 (14), **T043** | | 011 | T012 (13), T030 (35) | | 020 | T017, T024, **T042** |
| 003 | **T043** | | 012 | T005 (1) | | 021 | T012 (10, 11), T021 (21) |
| 004 | T021 (20), **T043** | | 013 | T021 (27), T030 (36), **T045** | | 022 | T016, T023, T032 *(informativos)*, **T042** |
| 005 | T012 (5) | | 014 | T021 (21, 22), T030 (39) | | 023 | **T042** |
| 006 | T021 (15, 16) | | 015 | T026 *(referencia)*, T027 (30) | | 024 | T034 (40), **T043** |
| 007 | T012 (4), T021 (18) | | 016 | puertas de cada corte, **T042** | | 025 | **T045** *(W1)*, **T051** *(W2)* |
| 008 | T012 (6) | | 017 | T027 (33), T034, T040, **T044**, T050 | | 026 | T021 (18, 19, 29) |
| 009 | T021 (25) | | 018 | T030 (38) | | | |

**28 / 28 FR** *(plan §Trazabilidad)* **y 26 / 26 SC con tarea.**

## Mutaciones — **55** *(plan §Bloques; lo que debe caer, allí)*

| Bloque | Tarea | Mutaciones | Tipo |
|---|---|---|---|
| B1 | T008 | m1–m8 *(raíz, sandbox, listado)* | conducta → suite |
| B2 | T015 | m9–m24 *(identidad, partidas, modelo, clases, prefiltro, refs)* | conducta → suite |
| B3 | T024 | m25–m40 *(frontera `forked`, N, cabeza, atajo, índice, repetidas, sólo lectura, `omitempty`)* | conducta → suite |
| B4 | T031 | m41–m49 · M-B4a | conducta → suite |
| B5 | T036 | m50–m54 | conducta → suite |

Las co-caídas que aparezcan al declarar el censo se escriben **antes** de mutar, por hoja.

## Lo que este plan de tareas NO hace

- No toca `internal/event`, las derivaciones de 006 y 008–011, `boundary_test.go`, los otros cinco lectores ni ningún `_test.go` existente. De `internal/state`,
  sólo el campo de T019.
- No cambia la plataforma: la fila de tarifa va en un parche aparte *(T048)*. Las tarifas de la membresía quedan sin decidir *(spec Q-5)*.
- No ejecuta `kimi`, ni lanza `--run` o `--daemon` sobre la instalación real: los de los tests y C2 van en sandbox, y los de W1 y W2 los lanza el dueño.
- No escribe en ninguna copia congelada, ni pone su ruta en el repo, ni copia de ellas una sola línea a `testdata/`.
