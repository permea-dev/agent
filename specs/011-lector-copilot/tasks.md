# Tasks: 011 · «Lector de Copilot CLI»

**Feature**: `011-lector-copilot` · **Fecha**: 2026-10-14 · [spec.md](./spec.md) *(ratificada el 2026-10-14; E-1)* · [plan.md](./plan.md) ·
[contracts/event-id-copilot.md](./contracts/event-id-copilot.md)
**Base**: `d81c39d` *(el código, el de `15c4c47`)*. **Línea base a preservar** *(sobre `15c4c47`, encargo 5/n)*: `go test -count=1 -v ./...` → **848**
«--- PASS», 0 FAIL, 0 SKIP · `go vet` limpio · `golangci-lint run` → **0**. Los tamaños de la 0.7.0, en plan §Technical Context.
**50 tareas · 12 ✋ · 40 tests nuevos** *(30 rojos y 10 nacidos verdes)* **· 44 mutaciones** *(plan §Bloques)*.
*Unidades*: aquí, «tests» son los rojos **numerados** del plan, (1)–(40). El registro cuenta **hojas**.

## Formato: `[ID] [P?] [✋?] Descripción`

- **[P]**: sin dependencia de decisión entre sí. No quiere decir que se editen a la vez.
- **✋**: la ejecuta **el dueño**: commits, PR, fusión, etiqueta, aprobación de textos y ensayos en Windows. Claude no hace git de escritura. Cada
  commit lleva su mensaje previsto, **sin tildes ni ñ**.
- **(1)…(40)** son los rojos, y **m1…m42**, **M-B3a** y **M-B4a** las mutaciones de plan §Bloques. Lo que se añada aquí se llama **M-B2a**, **M-B3b**…
- **No se renumera**: una tarea añadida después recibe el siguiente número libre.
- **Las transcripciones** van a `soporte/registro.md`, que nace en T001, y cada tarea deja una remisión.

## Presupuesto *(techos fijados por el dueño ANTES de la fase; ninguno se sube: si no cabe, se condensa)*

| Fichero de `specs/011-lector-copilot/` | Techo | Hoy |
|---|---:|---:|
| `spec.md` | 400 | 338 |
| `plan.md` | **250** | 236 |
| `tasks.md` *(este)* | **350** | 285 |
| `contracts/event-id-copilot.md` | **80** | 78 |
| `soporte/registro.md` *(nace en T001)* | **300** | 62 |
| `soporte/registro-archivo.md` *(§B0–§B6, íntegros; dueño, 14-10)* | **no cuenta** | 453 |
| `soporte/descubrimiento.md` · `-muestra-2.md` · `-muestra-3.md` | 200 · 80 · 60 *(cerrados)* | 200 · 80 · 60 |

| Bloque | Producción | Test | Fixtures | Registro |
|---|---:|---:|---|---:|
| B1 | ~45 | ~90 | `generar.sh` *(nace)*, `sesiones/` | ≤ 30 |
| B2 | ~140 | ~280 | `lineas/*.jsonl` | ≤ 45 |
| B3 | ~170 | ~330 | `ficheros/**`, `esperado.json` | ≤ 50 |
| B4 | ~90 | ~300 | la referencia de SC-015, `cmd/permea/testdata/copilot/` | ≤ 45 |
| B5 + B6 | ~70 + ~50 de texto | ~130 | — | ≤ 30 |
| Cierre C1–C9 | — | — | — | ≤ 60 |

**~515 de producción y ~1 130 de test.** Una desviación de más del 50 % se declara en el corte. **SC-007** *(≤ 100 KB)* manda sobre la estimación.

## Una suite a la vez, con la carga vigilada

Antes de **cada** `go test` o `golangci-lint` *(puertas, rojos, verdes y mutaciones)*, **en una orden aparte**, se lanza **`~/bin/libre.sh`**:
- 0 procesos de suite, **del agente o de la plataforma**;
- carga de 1 min **< U = 2,0** en 3 muestras seguidas, cada 10 s;
- tope de 45 min.

Mientras dice «no libre», se espera; al tope, **PARA 13**. El guion es el de la plataforma *(14-10)*, en **`~/bin/libre.sh`** *(md5 `ab647c01…`)*: lo usan
los dos repos y **no se versiona** *(dueño, 14-10)*. Excluye su PID y el de su padre.

## 🛑 Lo que PARA

1. **Un `_test.go` existente** necesita cambiar *(censo: ninguno)*, o se toca un fichero fuera de plan §Project Structure.
2. **En una mutación cae algo que no se declaró**, o no cae lo declarado. **La mutación se queda puesta** y se para.
3. **El md5 de la reversión** no es el de antes al empezar la mutación siguiente.
4. **La frontera de FR-027 tiene diff**, o `os.Getenv` de producción ≠ 5 tras B1.
5. **El lector escribe algo**: un md5 o un listado distinto en (26), en (34) o en W1, o un `.lock` nuevo o cambiado.
6. **Un `require` en `go.mod`**, o un `go.sum` *(SC-008)*.
7. **SC-007 pasa de 102 400 B** en alguna plataforma en C1. En los cortes de B3 y B4, la medida es informativa, y si pasa se adelgaza antes del bloque siguiente.
8. **Un dato real** en testdata, en el registro o en un informe: una línea, o un trozo, de una copia congelada; un id con forma UUID; una ruta personal; un
   prompt, una respuesta, un nombre de agente o de fichero de `checkpoints/`. Nada que no salga de `generar.sh`, y nada que case con el grep de privacidad,
   **tampoco con `grep -a`** sobre `testdata/`.
9. **Una copia congelada cambia su sha256.**
10. **Se ejecuta `copilot`**, o un `--run` o `--daemon` sobre la instalación real fuera de W1/W2.
11. **El lector lee una línea de un tipo que no es de D-2** más allá de la envoltura, o un campo fuera de FR-012.
12. **Un test que abre un testdata en su sitio**: siempre una copia en `t.TempDir()`.
13. **`~/bin/libre.sh` llega a su tope**, o su md5 no es `ab647c01…`.

## Puertas en cada corte

```sh
gofmt -l . ; go vet ./... ; golangci-lint run          # → vacío, limpio, 0 issues
go test -count=1 ./...                                 # → 848 + los nuevos, 0 FAIL, 0 SKIP; UNA suite a la vez, tras ~/bin/libre.sh
git diff 15c4c47 -- internal/event/ internal/ingest/eventid.go internal/ingest/eventid_test.go internal/ingest/codex_eventid.go \
  internal/ingest/codex_eventid_test.go internal/ingest/gemini_eventid.go internal/ingest/gemini_eventid_test.go \
  internal/ingest/opencode_eventid.go internal/ingest/opencode_eventid_test.go internal/ingest/boundary_test.go \
  specs/006-medicion-fiel/contracts/ specs/008-lector-codex/contracts/ specs/009-lector-gemini/contracts/ specs/010-lector-opencode/contracts/  # → vacío
git diff 15c4c47 --name-only --diff-filter=M -- '*_test.go'                                # → vacío
grep -c require go.mod ; ls go.sum                                                         # → 0 ; no existe
```

**Vigilancia de privacidad** *(cada corte con testdata, PARA 8)*. El patrón se escribe con clases *(`AI[z]a`…)* para no detectarse a sí mismo, y **el usuario de la
máquina y el nombre del dueño salen del entorno**: nunca se escriben en el repo.
```sh
PRIV='@[a-z0-9.-]+\.[a-z]{2,}|dev_[A-Za-z0-9]{8,}|ses_[A-Za-z0-9]{6,}|msg_[A-Za-z0-9]{6,}|/hom[e]/|C:\\User[s]|AI[z]a|sk-[a]nt|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}'
PRIV="$PRIV|${USER%?}|$(git config user.name)"
git diff 15c4c47 | grep -cE "$PRIV"                                                        # → 0
grep -raEc "$PRIV" internal/ingest/testdata/copilot/ cmd/permea/testdata/copilot/          # → 0 en cada fichero, generar.sh incluido
```

**Protocolo de mutación** *(006, con E-5 de 007)*:
1. El censo se declara **en la tarea, ANTES de mutar**, por **hoja**, con las co-caídas.
2. Se muta y se ejecuta la suite entera, tras `~/bin/libre.sh`.
3. Si los `FAIL` coinciden con lo declarado, se revierte **por edición inversa**, se comprueba el **md5** y se transcribe.
4. Si no coinciden, PARA 2.

Una mutación que no compila o que panica **no cuenta**. **Rojo antes de verde**, con la razón real transcrita. **Un test que nace verde lleva su mutación.**

---

## B0 · FASE 0 y registro

- [x] **T001** Fase 0: transcribir la línea base sobre `15c4c47` *(la del encargo 5/n: 848, vet, lint 0, y los tamaños de la 0.7.0 en las 5 plataformas)* y
  la frontera sin diff sobre `d81c39d`. Nace `soporte/registro.md`, con su techo. Si algo difiere, **PARA**.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B0 · T001 *(sobre `e9ad364`, sólo documentos desde `15c4c47`; 848 / 0 / 0 en 10 paquetes; vet y lint 0; frontera 0; `PRIV` 0)*.
- [ ] **T002** ✋ **Línea base y registro**. Commit: `011 B0: linea base y registro`.

## B1 · Raíz y ficheros *(FR-001 a FR-003; SC-012; M-3, M-5)* — nacen `internal/config/copilot_test.go`, `sandbox_copilot_test.go` y el listado en `copilot_fichero_test.go`

- [x] **T003** Fase 0: `config.CopilotSessionsRoot()`, que devuelve `"", nil`, e `ingest.ListarCopilot`, que devuelve `nil, nil`.
- [x] **T004** Testdata: **nace** `internal/ingest/testdata/copilot/generar.sh` *(plan §6)* con `sesiones/`: dos carpetas con `events.jsonl` mínimo,
  una sin él, `.session-operation-locks/` con un `.lock` de 0 B, `checkpoints/index.md`, `session.db` de 0 B y un `events.jsonl` a dos niveles. md5 al registro.
- [x] **T005** [P] **Rojos (1) a (3)**, con `t.Setenv`.
- [x] **T006** **Verde**: D-5 en `internal/config/copilot.go`; `ListarCopilot` de plan §2; tras `sandbox.go:68`, `t.Setenv("COPILOT_HOME", "")`; los
  comentarios de `codex.go:16`, `gemini.go:17` y `opencode.go:18`, «una de las CINCO».
- [x] **T007** Comprobar y transcribir, sin tocarlo, que `sandbox_test.go` sigue verde, y que `os.Getenv` de producción → **5**.
- [x] **T008** Censo y mutaciones **m1–m6**, con md5.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B1 *(861 pass; `os.Getenv` → 5; m1–m6 = declarado, con m2 → (1/vacia) y m5 → (3/solo_events) declaradas en el censo)*.
- [ ] **T009** ✋ Puertas y commit: `011 B1: raiz de Copilot CLI con COPILOT_HOME`.

## B2 · Una línea e identidad *(FR-010 a FR-017, FR-020; SC-005, SC-009, SC-011, SC-021)* — nacen `copilot_eventid_test.go` y `copilot_registro_test.go`

- [x] **T010** Fase 0:
  - `derivarEventIDCopilot(tipo, id string) (string, bool)`, que devuelve `"", false`;
  - `ContextoCopilot`, `ClaseCopilot` y `RegistroCopilot`, que aún no clasifica, en `internal/ingest/copilot.go`.
- [x] **T011** Testdata: `generar.sh` gana `lineas/`:
  - una línea por clase *(evento, compactación, compactación fallida, sin `compactionTokensUsed`, sin `requestId`, sin `usageId` ×3, las siete incoherentes,
    sin modelo)*;
  - las partidas de SC-005;
  - una con centinelas en todos los campos de texto *(SC-011)*;
  - `user.message` y `system.message` con el **falso marcador escapado**, y tres `model.*`;
  - una compactación con escritura ≠ 0 *(7/partidas)*, una con `reasoningTokens` 999 *(10/razonamiento)* y una `timestamp` con desfase `+02:00` *(11/desfase)*.

  md5 al registro.
- [x] **T012** [P] **Rojos (4) a (14)**: cada uno afirma **sólo su campo**; las 5 hojas de (4), el valor exacto. (14) cae porque el stub no emite *(rojo
  antes del verde)*. **Nacen verdes** con el stub, y se declaran así: **(5)** *(lo valida m9)*, **(13)** *(m19)* y la hoja **(7/fallida)** *(m14)*.
- [x] **T013** **Verde**: plan §1. Reutiliza `dominioEventID`, `bytesEventID` y `tipoRespuesta` *(`eventid.go:31,43`, `codex_eventid.go:29`)*; la
  codificación, **replicada**, como 009 y 010. **No** llama a `internal/pricing`.
- [x] **T014** **(14)**, rojo en T012, pasa a verde con T013: el evento no contiene ningún centinela. Lo valida m20.
- [x] **T015** Censo y mutaciones **m7–m20**, con md5, con las co-caídas de plan §Bloques: m9 → (5), (9) ×3, (7/sin_requestid); m11 → (6), (10/razonamiento);
  m12 y m13 → (6), (7/partidas); m17 → (11/desfase).
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B2 *(902 pass; m7–m20 = declarado, con m7 → (10/tool) y m16 → sólo (10/coste) declaradas en el censo)*.
- [x] **T050** **SC-007, informativo tras B2**: las 5 plataformas, D-011-P7 *(R-1)*. Si pasa de 102 400 B, se adelgaza antes de B3 *(PARA 7)*.
  > Transcripción: §B2 *(Δ fichero 0 en las 5; arm64 `LOAD` +256; `__TEXT` darwin/amd64 846 → 718 B: el lector aún no está enlazado)*.
- [x] **T016** **Vigilancia de privacidad** sobre `lineas/` *(§Puertas)* → 0, transcrito.
- [ ] **T017** ✋ Puertas y commit: `011 B2: un registro de Copilot CLI es un evento sin coste`.

## B3 · Fichero, frontera y tramos *(FR-004, FR-005, FR-007 a FR-009, FR-020; SC-003, SC-004, SC-006, SC-010, SC-013, SC-014, SC-023)* — `copilot_fichero_test.go`

- [x] **T018** Fase 0: `PasadaCopilot` *(`Resumen`, `HayNovedades`)* y `LeerFicheroCopilot`, que no emite, en `internal/ingest/copilot_contexto.go`.
- [x] **T019** Testdata: `generar.sh` gana `ficheros/`:
  - `dos_llamadas.jsonl`;
  - `tramos.jsonl` *(`start` con `/sintetico/x`, registro, `resume` con `/sintetico/y`, registro; `sourceSessionId` = `sessionId`)*;
  - `cortado/{paso1,paso2}.jsonl` *(paso 1: `start` X, registro y un `system.message` de 64 KB; paso 2: además, `resume` Y y registro)*;
  - `linea_larga.jsonl` *(un `system.message` de ≥ 64 KB antes de los registros; FR-009)*;
  - `sin_lf.jsonl` y `sin_lf_completo.jsonl`;
  - `truncado/{largo,corto}.jsonl` *(el corto, más pequeño que el offset del largo)*;
  - `repetida.jsonl`, `clases.jsonl` *(SC-010)*, `subagente.jsonl` *(m3)*, `matada.jsonl` *(m2 M1)*, `corrupta.jsonl` y `sin_start.jsonl`;
  - `esperado.json`.

  md5 al registro.
- [x] **T020** [P] **Rojos (15) a (26)**, cada pasada con su **línea de resumen literal**; (16/valores) afirma `event.Ref` y `Derivar` *(SC-021)*; (18), una
  sola hoja con las dos pasadas. **Nace verde**, y se declara así, **(26)** *(el stub no escribe; lo valida m29)*.
- [x] **T021** **Verde**: plan §2, D-011-P2 y P-4, y `Resumen()` con el texto de spec §Textos.
- [x] **T022** **SC-007, informativo**: las 5 plataformas, D-011-P7. Si pasa de 102 400 B, se adelgaza antes de B4 *(PARA 7)*.
- [x] **T023** Censo y mutaciones **m21–m29** y **M-B3a**, con md5, con las co-caídas de plan §Bloques: m21 → (16), (17); m23 → (24/sesion), (21/sin_start);
  m27 → (20), (23); M-B3a → (16/valores), (16/una_sesion). Vigilancia de privacidad sobre `ficheros/` → 0.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B3 *(924 pass; m21–m29 y M-B3a = declarado, con M-B3a → (24/sesion) declarada en el censo; (17) reforzado antes del censo)*.
- [ ] **T024** ✋ Puertas y commit: `011 B3: fichero de Copilot CLI por offset, con contexto por tramo`.

## B4 · Integración *(FR-002, FR-021, FR-023 a FR-026; SC-003, SC-006, SC-011, SC-015, SC-018, SC-019)* — nace `cmd/permea/copilot_test.go`

- [x] **T025** Fase 0, **en este orden**:
  1. **Antes de tocar `main.go`** *(D-011-P9)*: la referencia de SC-015 con el binario de `15c4c47`, y su md5. **Si `main.go` ya cambió, PARA.**
  2. El forzado A, comprobado *(`Load` pasa, `Append` pasa, `Save` falla)*.
  3. `cmd/permea/copilot.go` *(`generarCopilot`, sin emitir)*, y en `main.go` los campos, `setup()` y `generate()`.
- [x] **T026** [P] **Rojos (28) a (32)**, y **(27)**, nacido verde *(lo valida m30)*, con `cola` y `capturarStderr` de `codex_test.go`, sin tocarlo.
- [x] **T027** **Verde**: Copilot CLI tras OpenCode y antes del único `st.Save`; la activación en cada pasada; los errores por fichero *(D-011-P5)*; la
  línea en `runOnce` y, con `HayNovedades`, en `tick`.
- [x] **T028** Comprobar y transcribir, sin tocarlos, que los `_test.go` de `cmd/permea` siguen verdes.
- [x] **T029** **Nacen verdes (33) a (36)**, escritos tras el verde de T027: centinelas en la cola y en `state.json`; md5 y listado de la raíz y sus `.lock`
  tras `--run`; la carpeta creada después; y el fichero ilegible. Los validan m37, **M-B4a**, m34 y m35–m36.
- [x] **T030** Censo y mutaciones **m30–m37** y **M-B4a**, con md5. **m30** declara, **por nombre y hoja**, los tests existentes que caen *(los
  `*_SinRaiz*` y `*_LineaDeResumen` de Codex, Gemini y OpenCode)*, y se comprueba antes de mutar que no hay más.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B4 *(940 pass; m30–m37 y M-B4a = declarado; PARA 2 en m31 resuelto por la enmienda del censo del dueño, m31 → también (33))*.
- [x] **T031** **SC-007, informativo**, como T022.
  > Transcripción: §B4 *(lector enlazado: Δ máx. +25 184 B en darwin/amd64; arm64 `LOAD` +21 488; todas ≤ 102 400)*.
- [ ] **T032** ✋ Puertas y commit: `011 B4: Copilot CLI en run y daemon, antes de guardar el estado`.

## B5 · `--scan` *(FR-022, FR-023; SC-001, SC-002 y SC-023 en forma)*

- [x] **T033** [P] **Rojos (37) y (39)**, y **(38)** y **(40)**, nacidos verdes *(los validan m38 y m42)*, en `cmd/permea/copilot_test.go`, con los textos de
  spec §Textos **leídos por programa**. (39/no_multiplo) lleva un `nanoAIU` que no es múltiplo de 100, y (39/exacto) los 11 decimales.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B5 *(947 / 6; nacen verdes (38) ×4, (40) y la hoja (37/disco))*.
- [x] **T034** **Verde**: `esSesionCopilot`, **tras** Gemini, y `dryRunCopilot`, con la línea de control *(plan §5, D-011-P6)*.
  > Transcripción: §B5 *(953 / 0)*.
- [x] **T035** Censo y mutaciones **m38–m42**, con md5. **m38** declara, por nombre y hoja, los tests existentes de `--scan` que caen.
  > Transcripción: §B5 *(m38–m42 = declarado; ajustes: m38 → también (40), m41 → también (37/salida))*.
- [ ] **T036** ✋ Puertas y commit: `011 B5: scan de Copilot CLI con la linea de control de creditos`.

## B6 · README y CHANGELOG *(FR-028; P-8, P-10)*

- [x] **T037** ✋ **El dueño aprueba los textos**: «### Copilot CLI» en el README y la entrada `0.8.0` del CHANGELOG. Claude los propone en un encargo
  aparte, con estos límites:
  - las llamadas auxiliares;
  - la llamada en curso de una sesión matada;
  - los subagentes en segundo plano y los *workflows*;
  - la compactación automática;
  - los planes de pago y los nombres de modelo con punto;
  - medido en Windows con 1.0.97.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §B6 *(T1, T2 y R1–R8 aprobados; R9 no entra; `aprobado-v1.md` `2a330ed5`)*.
- [x] **T038** **Rojo**, transcrito: `grep -c '^## 0.8.0' CHANGELOG.md` → 0; `grep -c '^### Copilot CLI' README.md` → 0.
  > Transcripción: §B6 *(0 y 0)*.
- [x] **T039** README y CHANGELOG *(`## 0.8.0 — PENDIENTE`)*, con los textos **sacados por programa** de lo aprobado. `cmp` sin diferencias.
  > Transcripción: §B6 *(10 `cmp` sin diferencias; `PRIV` 0)*.
- [ ] **T040** ✋ Puertas y commit: `011 B6: README y CHANGELOG de la 0.8.0`.

## Cierre — en tramos, uno por mensaje *(plan §Cierre; si uno falla, se para y se rehace desde C1)*

- [x] **T041** **C1** · Puertas, con 0 SKIP en Linux:
  - **SC-007** en las 5 plataformas;
  - **SC-008** y **SC-016**;
  - **SC-020**: el patrón `PRIV` de §Puertas sobre `git diff 15c4c47` y, con `grep -raE`, sobre los testdata;
  - compilan Windows y darwin; `PENDIENTE` → 1.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §Cierre *(953 / 0; los 848 intactos; SC-007 máx. +33 760 B; PRIV 0 en 011)*.
- [x] **T042** **C2** · Medidas sobre **las 3 copias congeladas** *(`-14`, `-14b` y `-14c`)*, en su sitio y en sólo lectura, con sus sha256 antes y
  después *(PARA 9)*:
  - **SC-001**, **SC-002** y **SC-003** sobre `-14b` *(`--run` con `COPILOT_HOME` = la copia, sandbox sin enrolar; `--scan` de F)*;
  - **SC-004** sobre D, **SC-009** sobre E, **SC-023** sobre G;
  - y `-14`, igual que A+B y C en `-14b`.

  Contraste independiente: un guion `python3 -I` fuera del repo cuenta registros, partidas y créditos. Una diferencia → **PARA**.
  > Transcripción: §Cierre *(TODO IGUAL; sha256 de las 3 copias iguales)*.
- [x] **T043** **C3** · `goreleaser release --snapshot --clean`, el SHA-256 del zip de Windows, y `strings` con los textos *(SC-017)*. `dist/` borrado.
  > Transcripción: §Cierre *(zip `8be731d5…`; los 3 textos en `permea.exe`)*.
- [x] **T044** ✋ **C4 · W1** *(plan D-011-W)*: Copilot cerrado, y luego abierto con un `/compact`; el contador `python3 -I`. **SC-006** y **SC-022** *(W1, con el control de créditos)*.
  > Transcripción: [`soporte/registro.md`](./soporte/registro.md) §Cierre *(W1: TODO IGUAL; 16 = 16 y 19 = 19, créditos iguales; **SC-022 cumple en Windows**)*.
- [x] **T045** **C5** · El cuerpo del PR, fuera del repo; la fecha de la 0.8.0 entra en la rama *(`PENDIENTE` → 0)*.
  > Transcripción: §Cierre *(`## 0.8.0 — 2026-10-14`; el cuerpo, en `tmp/pr-body.md`, ignorado por git)*.
- [ ] **T046** ✋ **C6** · Fusión del PR con merge commit.
- [ ] **T047** ✋ **C7** · `git tag -a v0.8.0` y push; el workflow `release` en verde.
- [ ] **T048** **C8** · Los tres canales en `0.8.0`, `sha256sum -c`, y `strings` del binario publicado.
- [ ] **T049** ✋ **C9 · W2**: `scoop update`, `--run`. En la plataforma, eventos `copilot_cli` = los registros nuevos, y **evento a evento**
  `ROUND(créditos × 0,01, 6)` = su coste *(la plataforma redondea por evento; SC-022)*.

## Dependencias

- **Bloques**: B0 → B1 → B2 → B3 → B4 → B5 → B6 → C1…C9. B3 usa B2; B4 usa B1 y B3; B5 usa B3.
- **Dentro de un bloque**: los rojos `[P]` antes del verde, y el censo antes de mutar.

## Censo de tests *(spec §Censo; ninguno existente cambia)*

| Tarea | Nace |
|---|---|
| T005 | `internal/config/copilot_test.go`, `internal/testutil/sandbox_copilot_test.go`, `internal/ingest/copilot_fichero_test.go` *(el listado)* |
| T012 | `internal/ingest/copilot_eventid_test.go`, `copilot_registro_test.go` |
| T020 | `internal/ingest/copilot_fichero_test.go` *(el fichero)* |
| T026, T029, T033 | `cmd/permea/copilot_test.go` |

**Fuera de `_test.go`**: `internal/testutil/sandbox.go`, una línea, en T006.

## Cobertura *(SC → tarea que lo acredita)*

| SC | Tareas | | SC | Tareas | | SC | Tareas |
|---|---|---|---|---|---|---|---|
| 001 | T033 *(forma)*, **T042** | | 009 | T012 (7), **T042** | | 017 | T026 (30), T033, T039, **T043**, T048 |
| 002 | T033 (39), **T042** | | 010 | T020 (23) | | 018 | T029 (36) |
| 003 | T020 (15), T026 (32), **T042** | | 011 | T012 (4, 14), T029 (33) | | 019 | T026 (28, 29) |
| 004 | T020 (18, 25), **T042** | | 012 | T005 (1) | | 020 | T016, T023, **T041** |
| 005 | T012 (6) | | 013 | T020 (16, 17) | | 021 | T012 (10, 11), T020 (16, 21) |
| 006 | T020 (26), T029 (34), **T044** | | 014 | T020 (19) | | 022 | **T044** *(W1)*, **T049** *(W2)* |
| 007 | T050, T022, T031 *(informativos)*, **T041** | | 015 | T025 *(referencia)*, T026 (27) | | 023 | T020 (24), T033 *(forma)*, **T042** |
| 008 | **T041** | | 016 | puertas de cada corte, **T041** | | | |

**28 / 28 FR** *(plan §Trazabilidad)* **y 23 / 23 SC con tarea.**

## Mutaciones — **44** *(plan §Bloques; lo que debe caer, allí)*

| Bloque | Tarea | Mutaciones | Tipo |
|---|---|---|---|
| B1 | T008 | m1–m6 *(raíz, sandbox, listado)* | conducta → suite |
| B2 | T015 | m7–m20 *(identidad, partidas, clases, prefiltro, refs)* | conducta → suite |
| B3 | T023 | m21–m29 · M-B3a *(tramos, prefijo, subagente, repetidas, aperturas, sólo lectura, refs)* | conducta → suite |
| B4 | T030 | m30–m37 · M-B4a | conducta → suite |
| B5 | T035 | m38–m42 | conducta → suite |

Las co-caídas que aparezcan al declarar el censo se escriben **antes** de mutar, por hoja.

## Lo que este plan de tareas NO hace

- No toca `internal/event`, `internal/state`, las derivaciones de 006, 008, 009 y 010, `boundary_test.go`, los otros cuatro lectores, ni ningún `_test.go`
  existente.
- No cambia la plataforma: las tarifas de `gpt-6-luna` y `gpt-5.6-luna` ya están *(spec §Dependencias)*. Los nombres con punto quedan allí, sin decidir.
- No ejecuta `copilot`, ni lanza `--run` o `--daemon` sobre la instalación real: los de los tests y C2 van en sandbox, y los de W1 y W2 los lanza el dueño.
- No escribe en ninguna copia congelada, ni pone su ruta en el repo, ni copia de ellas una sola línea a `testdata/`.
