# Tasks: 010 · «Lector de OpenCode»

**Feature**: `010-lector-opencode` · **Fecha**: 2026-10-11 · [spec.md](./spec.md) *(ratificada el 2026-10-11; Q-5 cerrada)* · [plan.md](./plan.md) ·
[contracts/event-id-opencode.md](./contracts/event-id-opencode.md)
**Base**: `82e0b2d`. **Línea base a preservar**: `go test -count=1 -v ./...` → **671** «--- PASS», 0 FAIL, 0 SKIP · `go vet` limpio ·
`golangci-lint run` → **0**.
**59 tareas · 14 ✋ · 53 tests nuevos** *(39 rojos y 14 nacidos verdes)* **· 59 mutaciones** *(plan §Bloques, más M-B1a y M-B1b)*. **La release queda bloqueada por SC-027** *(P-033)*.
*Unidades*: aquí, «tests» son los rojos **numerados** del plan, (1)–(53); el registro cuenta **hojas y semillas**. B1: 10 numerados = 40 hojas + 15 semillas = 55.

## Formato: `[ID] [P?] [✋?] Descripción`

- **[P]**: sin dependencia de decisión entre sí. No quiere decir que se editen a la vez.
- **✋**: la ejecuta **el dueño**: commits, PR, fusión, etiqueta, aprobación de textos y ensayos en Windows. Claude no hace git de escritura. Cada
  commit lleva su mensaje previsto, **sin tildes ni ñ**.
- **(1)…(53)** son los rojos, y **mB1…mB10** y **m1…m46** las mutaciones de plan §Bloques. Lo que se añada aquí se llama **M-B1a**, **M-B4a**…
- **No se renumera**: una tarea añadida después recibe el siguiente número libre.
- **Las transcripciones** van a `soporte/registro-mutaciones.md`, que nace en T001, y cada tarea deja una remisión.

## Presupuesto *(techos fijados por el dueño ANTES de la fase; ninguno se sube: si no cabe, se condensa)*

| Fichero de `specs/010-lector-opencode/` | Techo | Hoy |
|---|---:|---:|
| `spec.md` | 450 | 450 |
| `plan.md` | **250** | 229 |
| `tasks.md` *(este)* | **350** | 309 |
| `contracts/event-id-opencode.md` | **80** | 73 |
| `soporte/registro-mutaciones.md` *(nace en T001)* | **300** | 291 |
| `soporte/prueba-sqlite.md` · `soporte/medicion-anthropic.md` | 200 · 150 *(cerrados)* | 168 · 144 |

| Bloque | Producción | Test | Fixtures | Registro |
|---|---:|---:|---|---:|
| B1 | ~560 | ~450 | `generar.sh`, `esperado.json`, I, R, S, raros, `profundo.db` | ≤ 50 |
| B2 | ~35 | ~60 | — | ≤ 20 |
| B3 | ~150 | ~260 | filas JSON sintéticas | ≤ 35 |
| B4 | ~260 | ~520 | `generar.sh`, `esperado.json`, instantáneas por paso | ≤ 60 |
| B5 | ~30 | ~70 | — | ≤ 20 |
| B6 | ~120 | ~330 | la referencia de SC-016 | ≤ 40 |
| B7 + B8 | ~80 + ~50 de texto | ~120 | — | ≤ 25 |
| Cierre C1–C9 | — | — | — | ≤ 50 |

**~1 235 de producción y ~1 800 de test.** Una desviación de más del 50 % se declara en el corte. **SC-007** *(≤ 100 KB)* manda sobre la estimación.

## 🛑 Lo que PARA

1. **Un `_test.go` existente** necesita cambiar *(censo: ninguno)*, o se toca un fichero fuera de plan §Project Structure.
2. **En una mutación cae algo que no se declaró**, o no cae lo declarado. **La mutación se queda puesta** y se para.
3. **El md5 de la reversión** no es el de antes al empezar la mutación siguiente.
4. **La frontera de FR-030 tiene diff**, o `os.Getenv` de producción ≠ 4 tras B5.
5. **El lector escribe algo**: un md5 o un listado distinto en (10), en (48) o en W1.
6. **Un `require` en `go.mod`**, o un `go.sum` *(FR-011)*.
7. **SC-007 pasa de 102 400 B** en alguna plataforma en C1. En el corte de B6, la medida es informativa, y si pasa se adelgaza antes de B7.
8. **Un dato real** en testdata, en el registro o en un informe: un `id` con la forma real *(4 + 26)*, una ruta personal, un título o un texto. Nada
   que no salga de un `generar.sh`, y nada que case con el grep de privacidad, **tampoco con `grep -a` dentro de las bases** *(centinelas `CENTINELA-…`,
   rutas `/sintetico/…`)*.
9. **Una copia congelada cambia su sha256.**
10. **Crear `v0.7.0` sin SC-027.**
11. **Un fixture de `cmd/permea` con una fila en curso** *(su resultado cambiaría 24 h después de generarlo; D-010-P9)*.
12. **Un test que abre un testdata en su sitio**: siempre una copia en `t.TempDir()`.

## Puertas en cada corte

```sh
gofmt -l . ; go vet ./... ; golangci-lint run          # → vacío, limpio, 0 issues
go test -count=1 ./...                                 # → 671 + los nuevos, 0 FAIL, 0 SKIP; UNA suite a la vez
git diff 4c6cafc -- internal/event/ internal/ingest/eventid.go internal/ingest/eventid_test.go internal/ingest/codex_eventid.go \
  internal/ingest/codex_eventid_test.go internal/ingest/gemini_eventid.go internal/ingest/gemini_eventid_test.go \
  internal/ingest/boundary_test.go specs/006-medicion-fiel/contracts/ specs/008-lector-codex/contracts/ specs/009-lector-gemini/contracts/  # → vacío
git diff 4c6cafc --name-only --diff-filter=M -- '*_test.go'                                # → vacío
grep -c require go.mod ; ls go.sum                                                         # → 0 ; no existe
```

**Protocolo de mutación** *(006, con E-5 de 007)*:
1. El censo se declara **en la tarea, ANTES de mutar**, por **hoja**, con las co-caídas.
2. Se muta y se ejecuta la suite entera.
3. Si los `FAIL` coinciden con lo declarado, se revierte **por edición inversa**, se comprueba el **md5** y se transcribe.
4. Si no coinciden, PARA 2.

Una mutación que no compila o que panica **no cuenta**: (9) recupera el pánico y lo convierte en fallo.

**Rojo antes de verde**, con la razón real transcrita. **Un test que nace verde lleva su mutación.**

---

## B0 · Documentos

- [x] **T001** Fase 0: transcribir la línea base sobre `82e0b2d` *(671, vet, lint 0, frontera sin diff)*. Nace `soporte/registro-mutaciones.md`, con
  su techo. Si algo difiere, **PARA**.
  > Transcripción: [`soporte/registro-mutaciones.md`](./soporte/registro-mutaciones.md) §B0 · T001 *(sobre `0256098`, sólo documentos desde `82e0b2d`; 671 pass en 9 paquetes, 0 SKIP; lint 0; frontera 0)*.
- [ ] **T002** ✋ **Línea base y registro**. La enmienda **E-1** ya está en la spec desde el remate del analyze *(9/n)*: T002 no la aplica. Commit:
  `010 B0: linea base y registro de mutaciones`.

## B1 · Lector *(`internal/sqlitero`; FR-006 a FR-010; SC-004, SC-005, SC-009)* — nacen `lector_test.go` y `fuzz_test.go`

- [x] **T003** Fase 0: `internal/sqlitero/lector.go` **= el de la prueba**:
  - se transcribe el md5 de origen, `f6b47fa8…`;
  - sólo cambia el nombre del paquete;
  - `RecorrerDesde`, `Buscar` y `Paginas` son stubs que devuelven error o 0.

  La suite sigue verde. Se transcribe el md5 del fichero portado.
- [x] **T004** Testdata: `internal/sqlitero/testdata/generar.sh` *(plan §6)* → `i/`, `r/`, `s/`, `raros/` *(utf16, without_rowid, p512, p65536,
  reserva, desborda)*, `profundo.db` *(≥ 3 niveles, sin WAL, filas < 1 KB)* y `esperado.json`.
  - Las cifras de I, R y S deben ser las de la prueba en forma: R sin T3, e I sin la transacción a medias.
  - **R**: el último byte de una hoja de `session_message` de T3, comprobado **dentro de una celda** por el guion *(H9)*.
  - md5 de cada fichero al registro.
- [x] **T005** [P] **Rojos (1) a (10)** *(plan §Bloques)* en `lector_test.go`, y `FuzzBase` y `FuzzWAL` con sus semillas en `fuzz_test.go`. **El fixture de cada
  rojo**, el de plan §Bloques *(H8)*. Las que nacen verdes *(lógica portada: (1)–(3), (6), **(7)**, (9), (10))* llevan su mutación: mB1–mB4, mB8,
  **mB9**, mB10, y dos nuevas:
  - **M-B1a**, aceptar UTF-16 → (1/utf16);
  - **M-B1b**, `Abrir` crea un `-shm` vacío → (10/listado).

  (9/carga_enorme) declara **2⁶²** y llama con `recover`. La goroutine de los fuzz, también con `recover` *(H7)*.
- [x] **T006** **Verde**:
  - `RecorrerDesde` con poda; `Buscar` por descenso; `Paginas()`;
  - errores **sin ruta** *(`PathError` → `Op: Err`)*; `ErrCorrupta` y `ErrNoSoportada`.

  Se transcribe el md5.
- [x] **T007** Censo, **antes de mutar**, por hoja: **mB1–mB10**, **M-B1a** y **M-B1b**, con lo que debe caer en plan §Bloques. Entre las
  co-caídas, (7/sin_cambio) para mB1 y `FuzzBase` *(semilla `carga_enorme`)* para mB8.
  > Transcripción: [`soporte/registro-mutaciones.md`](./soporte/registro-mutaciones.md) §B1 · T003–T008 *(726 pass; mB4 también tumba (2/R); T008: 12 de 12 = declarado, tras corregir la co-caída (10/md5) de M-B1b)*.
- [x] **T008** Mutaciones y transcripción, con md5. **mB1–mB4 son las de la prueba, rehechas** sobre el código definitivo *(D-3)*.
- [ ] **T009** ✋ Puertas y commit: `010 B1: lector SQLite de solo lectura (internal/sqlitero)`.

## B2 · Identidad *(contrato; FR-013; SC-014)* — nace `opencode_eventid_test.go`

- [x] **T010** Fase 0: `derivarEventIDOpenCode(id string) (string, bool)` en `internal/ingest/opencode_eventid.go`, que devuelve `"", false`.
- [x] **T011** [P] **Rojos (11) a (13)**: los dos vectores; el de Gemini *(`a53e1c53…`)*, distinto; vacío → no ok *(nace verde; m3)*.
- [x] **T012** **Verde**: la derivación, **replicada** *(reutiliza `dominioEventID`, `bytesEventID` y `tipoRespuesta`; `eventid.go:31,43`,
  `codex_eventid.go:29`)*.
- [x] **T013** Censo y mutaciones **m1–m4**, con md5.
  > Transcripción: [`soporte/registro-mutaciones.md`](./soporte/registro-mutaciones.md) §B2 *(731 pass; m1–m4 = declarado)*.
- [ ] **T014** ✋ Puertas y commit: `010 B2: event_id de OpenCode con espacio de nombres propio`.

## B3 · Una llamada *(FR-012, FR-014 a FR-021; SC-012, SC-013)* — nace `opencode_llamada_test.go`

- [x] **T015** Fase 0: en `internal/ingest/opencode.go`, `ContextoOpenCode` *(`Context`, `Ahora`, `SessionID`, `Directory` y `Path`; H17)*, `ClaseOpenCode` y `LlamadaOpenCode`, que aún no
  clasifica.
- [x] **T016** Fixtures: filas sintéticas *(`data` JSON en `testdata/opencode/filas/`)*: una por clase, las partidas de SC-012 y una con centinelas.
- [x] **T017** [P] **Rojos (14) a (22)**: cada uno afirma **sólo su campo**. (22) cae porque el stub no emite evento *(H19)*.
- [x] **T018** **Verde**: P-2, la clasificación de D-010-P14, `occurred_at` y las referencias. **No** llama a `internal/pricing`.
- [x] **T019** **(22)**, escrito en T017 como rojo, pasa a verde con T018: el evento no contiene ningún centinela. Lo valida m13.
- [x] **T020** Censo y mutaciones **m5–m14**, con md5.
  > Transcripción: [`soporte/registro-mutaciones.md`](./soporte/registro-mutaciones.md) §B3 *(765 pass; m5–m14 = declarado)*.
- [ ] **T021** ✋ Puertas y commit: `010 B3: una llamada de OpenCode es un evento sin coste`.

## B4 · Base y frontera *(FR-003 a FR-005, FR-008, FR-022, FR-023; SC-010, SC-011, SC-013, SC-015, SC-017, SC-024, SC-025)* — nacen `opencode_base_test.go` y `opencode_coste_test.go`

- [x] **T022** Fase 0:
  - `internal/state/state.go`: `FileState.Frontera` *(D-010-P2)*. **Ningún test de `internal/state` cambia**, y siguen verdes, transcrito;
  - `opencode_base.go`, con `PasadaOpenCode`, `LeerBaseOpenCode`, `ListarOpenCode` y `ContarFormatoOpenCode`, que no emiten.
- [x] **T023** Testdata: `internal/ingest/testdata/opencode/generar.sh` →
  - `frontera/paso1…paso3/` *(4 completadas y 1 en curso)*, `inverso/`, `lista_borrada/` *(con una completada)* y `vacuum/{antes,despues}/` *(2
    borradas + `VACUUM` + 3 nuevas)*;
  - `borrada_h/`: se borran las de rowid máximo y **entran 2 que reutilizan rowids ≤ H** *(H10)*;
  - `abandono/paso1…paso3/` y `abandono_1h/` *(`time.created` fijos)*, `clases.db` *(con una en curso de 25 h)*, `dos_sesiones.db` y `centinelas.db`
    *(H23)*;
  - `formatos/{v1,without_rowid,utf16,sin_columna}.db` y `corrupta.db`;
  - `esperado.json`, con los rowids de H antes y después de `VACUUM`.

  md5 al registro.
- [x] **T024** [P] **Rojos (23) a (38)**, con `Ahora` inyectado en **todos** *(H6)* y la línea de resumen literal en cada pasada. Para (33), un
  gancho de test cambia el `-wal` de la copia entre la lectura y `Estable()`.
- [x] **T025** **Verde**: el algoritmo de plan §2, D-010-P5 a P-9, y `Resumen()` con el texto de spec §Textos.
  - **SC-017 informativa**: `go test -tags sc017 …` una vez, transcrita.
- [x] **T026** Censo, **antes de mutar**, por hoja: **m15–m29**, con las co-caídas de plan §Bloques, incluida la de un test **existente** para m29:
  `TestRetencion_SC009_NadaDelProveedorEnDisco/state_json_cuatro_campos`.
- [x] **T027** Mutaciones y transcripción, con md5.
  > Transcripción: [`soporte/registro-mutaciones.md`](./soporte/registro-mutaciones.md) §B4 *(809 pass; SC-017: 1.ª pasada 29,3 s > 20 s, informativa; m15–m29 = declarado)*.
  > **Remate** *(§B4 remate)*: un análisis de `data` por fila; SC-017 **cumple** × 3 *(1.ª pasada 5,1–5,2 s; siguiente ≤ 9,3 ms, 0,030 %)*; m9–m11 repetidas, = declarado.
- [ ] **T028** ✋ Puertas y commit: `010 B4: frontera por rowid y pasada de una base de OpenCode`.

## B5 · Ubicación *(FR-001, FR-002; M-3, M-5; SC-015)* — nacen `internal/config/opencode_test.go` y `internal/testutil/sandbox_opencode_test.go`

- [x] **T029** Fase 0: `config.OpenCodeDatos()`, que devuelve `"", "", nil`.
- [x] **T030** [P] **Rojos (39) y (40)**, con `t.Setenv`.
- [x] **T031** **Verde**: P-8, en `internal/config/opencode.go`. Tras `sandbox.go:65`, los dos `t.Setenv`. Los comentarios de `codex.go:16` y
  `gemini.go:17` pasan a «una de las CUATRO».
- [x] **T032** Comprobar y transcribir, sin tocarlo, que `sandbox_test.go` sigue verde, y `os.Getenv` de producción → **4**.
- [x] **T033** Censo y mutaciones **m30–m34**, con md5.
  > Transcripción: [`soporte/registro-mutaciones.md`](./soporte/registro-mutaciones.md) §B5 *(821 pass; `os.Getenv` → 4; m30–m34 = declarado)*.
- [ ] **T034** ✋ Puertas y commit: `010 B5: ubicacion de OpenCode con XDG_DATA_HOME y OPENCODE_DB`.

## B6 · Integración *(FR-002, FR-024, FR-026 a FR-029; SC-003, SC-006, SC-014, SC-016, SC-020, SC-021)* — nace `cmd/permea/opencode_test.go`

- [x] **T035** Fase 0, **en este orden**:
  1. **Antes de tocar `main.go`** *(D-010-P10)*: la referencia de SC-016 con el binario de `4c6cafc`, y su md5. **Si `main.go` ya cambió, PARA.**
  2. El forzado A, comprobado *(`Load` pasa, `Append` pasa, `Save` falla)*.
  3. `cmd/permea/opencode.go` *(`generarOpenCode`, sin emitir)*, y en `main.go` los campos, `setup()` y `generate()`, con su firma de hoy.
- [x] **T036** [P] **Rojos (41) a (46)**, con `cola` y `capturarStderr` de `codex_test.go`, sin tocarlo. (41) nace verde; la valida m35.
- [x] **T037** **Verde**: OpenCode tras Gemini y antes del único `st.Save`; la activación en cada pasada; los errores por base *(D-010-P6, P7)*; la
  línea en `runOnce` y, con `HayNovedades`, en `tick`.
- [x] **T038** Comprobar y transcribir, sin tocarlos, que los `_test.go` de `cmd/permea` siguen verdes.
- [x] **T039** **Nacen verdes (47) a (50)**, porque se escriben tras el verde de T037: centinelas en la cola y en `state.json`; md5 y listado de I
  tras `--run`; la carpeta creada después; y la base corrupta. Los validan m42, **M-B6a**, m39 y m40–m41. Ningún fixture de `cmd` con filas en curso
  *(PARA 11)*.
- [x] **T040** Censo y mutaciones **m35–m42** y **M-B6a**, con md5. **m35** declara, por nombre y hoja, los tests existentes de plan §Bloques
  *(`TestCodexRun_*` y `TestGeminiRun_*`)*, y se comprueba antes de mutar que no hay más.
  > Transcripción: [`soporte/registro-mutaciones.md`](./soporte/registro-mutaciones.md) §B6 *(835 pass; m35–m42 y M-B6a = declarado; rehecha el 13-10)*.
- [x] **T041** **SC-007, informativo**: las 5 plataformas, con el método de D-010-P12. Si pasa de 102 400 B, se adelgaza **antes** de B7 *(PARA 7)*.
  > PARA 7 *(12-10)* resuelto con **E-3**: en linux/arm64, contenido +78 544 B *(fichero +131 072, sin juzgar)*; las otras, 68 064–92 672. [`soporte/registro-mutaciones.md`](./soporte/registro-mutaciones.md) §B6.
- [ ] **T042** ✋ Puertas y commit: `010 B6: OpenCode en run y daemon, antes de guardar el estado`.

## B7 · `--scan` *(FR-025, FR-026; SC-001, SC-002 en forma)*

- [ ] **T043** [P] **Rojos (51) y (53)**, y **(52) nacido verde**, en `cmd/permea/opencode_test.go`, con los textos de spec §Textos leídos por
  programa. El fixture de (53) lleva **Δ`reasoning` ≠ 0** *(H13)*.
- [ ] **T044** **Verde**: `esBaseSQLite`, **antes** de `esSesionCodex`, y `dryRunOpenCode` con la línea de contraste *(plan §5)*.
- [ ] **T045** Censo y mutaciones **m43–m46**, con md5. **m44** declara, por nombre y hoja, los diez tests existentes de `--scan` de plan §Bloques
  *(como m36 en 009)*.
- [ ] **T046** ✋ Puertas y commit: `010 B7: scan de OpenCode con la linea de contraste`.

## B8 · README y CHANGELOG *(FR-031; P-12)*

- [ ] **T047** ✋ **El dueño aprueba los textos**: «### OpenCode» en el README y la entrada `0.7.0` del CHANGELOG. Claude los propone en un encargo
  aparte, con estos límites:
  - D-2, el 0,37 %;
  - el abandono a las 24 h;
  - partidas verificadas con Google y Anthropic;
  - la dependencia de P-033.
- [ ] **T048** **Rojo**, transcrito: `grep -c '^## 0.7.0' CHANGELOG.md` → 0; `grep -c '^### OpenCode' README.md` → 0.
- [ ] **T049** README y CHANGELOG *(`## 0.7.0 — PENDIENTE`)*, con los textos **sacados por programa** de lo aprobado. `cmp` sin diferencias.
- [ ] **T050** ✋ Puertas y commit: `010 B8: README y CHANGELOG de la 0.7.0`.

## Cierre — en tramos, uno por mensaje *(plan §Cierre; si uno falla, se para y se rehace desde C1)*

- [ ] **T051** **C1** · Puertas, con 0 SKIP en Linux:
  - **SC-007** en las 5 plataformas;
  - **SC-008**;
  - **SC-018**;
  - **SC-022**: el grep de privacidad, más el prefijo de las claves de Anthropic, sobre `git diff 82e0b2d`, y testdata sólo de `generar.sh`;
  - compilan Windows y darwin; `PENDIENTE` → 1.
- [ ] **T052** **C2** · Medidas, con los sha256 de las dos copias congeladas antes y después *(PARA 9)*:
  - **SC-001** y **SC-003**, con `--scan` sobre la copia del 09-10 y dos `--run` con `OPENCODE_DB` = la copia *(sandbox, sin enrolar)*;
  - **SC-002**, la línea de contraste;
  - la copia de `anthropic`, de control: 24 eventos;
  - **SC-009**: `FuzzBase` y `FuzzWAL`, 10 min cada uno;
  - **SC-017**, tres veces;
  - **equivalencia**: el lector portado = el de la prueba *(`tmp/sonda-sqlite/b-propio/lector/`)* sobre los 13 testdata y las dos copias congeladas. Se
    comparan recuentos, sumas y sha256 de `data` por tabla. Las copias se abren en su sitio, en solo lectura e `immutable=1`, con su sha256 antes y
    después. Una diferencia → **PARA**.
- [ ] **T053** **C3** · `goreleaser release --snapshot --clean`, el SHA-256 del zip de Windows y `strings` con los textos *(SC-019)*.
- [ ] **T054** ✋ **C4 · W1** *(plan D-010-W)*: OpenCode cerrado, y luego abierto; `sonda-a.exe` como contador. **SC-006** y **SC-023**.
- [ ] **T055** **C5** · El cuerpo del PR, que dice **«la etiqueta espera a SC-027 (P-033)»**.
- [ ] **T056** ✋ **C6** · Fusión del PR con merge commit *(pregunta 1 del reporte del plan)*.
- [ ] **T057** ✋ **C7** · 🛑 **BLOQUEADA hasta SC-027**:
  - P-033 en producción;
  - la plataforma repara las 5 filas de la medición en **0,043348**;
  - después, `## 0.7.0 — <fecha>` *(`PENDIENTE` → 0)*, `git tag -a v0.7.0` y push.
- [ ] **T058** **C8** · Los tres canales en `0.7.0`, y `strings` del binario publicado.
- [ ] **T059** ✋ **C9 · W2**: `scoop update`, `--run`, y en la plataforma, eventos `opencode` = las filas nuevas del contador *(SC-023)*.

## Dependencias

- **Bloques**: B0 → B1 → B2 → B3 → B4 → B5 → B6 → B7 → B8 → C1…C6, y **C7 sólo tras P-033**. B3 usa B2; B4 usa B1 y B3; B6 usa B4 y B5; B7 usa B4.
- **Dentro de un bloque**: los rojos `[P]` antes del verde, y el censo antes de mutar.

## Censo de tests *(spec §Censo; ninguno existente cambia)*

| Tarea | Nace |
|---|---|
| T005 | `internal/sqlitero/lector_test.go`, `fuzz_test.go` |
| T011 | `internal/ingest/opencode_eventid_test.go` |
| T017 | `internal/ingest/opencode_llamada_test.go` |
| T024 | `internal/ingest/opencode_base_test.go` · T025: `opencode_coste_test.go` *(`sc017`)* |
| T030 | `internal/config/opencode_test.go`, `internal/testutil/sandbox_opencode_test.go` |
| T036, T039, T043 | `cmd/permea/opencode_test.go` |

**Fuera de `_test.go`**:
- `internal/testutil/sandbox.go`, en T031;
- `internal/state/state.go`, el campo, en T022 *(DISCREPANCIA 2 del reporte)*.

## Cobertura *(SC → tarea que lo acredita)*

| SC | Tareas | | SC | Tareas | | SC | Tareas |
|---|---|---|---|---|---|---|---|
| 001 | T043 *(forma)*, **T052** | | 010 | T024 (25) | | 019 | T036 (44), T043, T049, **T053**, T058 |
| 002 | T043 (53), **T052** | | 011 | T024 (23, 24, 26–28) | | 020 | T039 (50) |
| 003 | T036 (46), **T052** | | 012 | T017 (14) | | 021 | T036 (42, 43) |
| 004 | T005 (2) | | 013 | T017 (16, 17, 20), T024 (32) | | 022 | **T051** |
| 005 | T007–T008 *(mB1–mB4)* | | 014 | T011, T019 (22), T024 (38), T039 (47) | | 023 | **T054** *(W1)*, T059 *(W2)* |
| 006 | T005 (10), T039 (48), **T054** | | 015 | T005 (1, 6), T024 (34, 35), T030 (39) | | 024 | T024 (29) |
| 007 | T041 *(informativo)*, **T051** | | 016 | T035 *(referencia)*, T036 (41) | | 025 | T024 (31) |
| 008 | **T051** | | 017 | T025 *(informativa)*, **T052** *(× 3)* | | 026 | ✅ **cumplido** *(spec, 11-10; `soporte/medicion-anthropic.md`)* |
| 009 | T005 *(semillas)*, **T052** *(10 min × 2)* | | 018 | puertas de cada corte, **T051** | | 027 | **T057** 🛑 |
| 028 | T017 (18, 19, 21) | | 029 | T024 (33) | | | |

**31 / 31 FR** *(plan §Trazabilidad)* **y 29 / 29 SC con tarea.**

## Mutaciones — **59** *(plan §Bloques, más M-B1a y M-B1b; lo que debe caer, allí y en T005)*

| Bloque | Tarea | Mutaciones | Tipo |
|---|---|---|---|
| B1 | T008 | mB1–mB4 *(las de la prueba, rehechas)* · mB5–mB10 · M-B1a · M-B1b | conducta → suite |
| B2 | T013 | m1–m4 | conducta → suite |
| B3 | T020 | m5–m14 | conducta → suite |
| B4 | T027 | m15–m29 *(frontera, renumeración, abandono, formatos, `state.json`)* | conducta → suite |
| B5 | T033 | m30–m34 | conducta → suite |
| B6 | T040 | m35–m42 · M-B6a | conducta → suite |
| B7 | T045 | m43–m46 | conducta → suite |

Las co-caídas que aparezcan al declarar el censo se escriben **antes** de mutar, por hoja.

## Lo que este plan de tareas NO hace

- No toca `internal/event`, las derivaciones de 006, 008 y 009, `boundary_test.go`, los lectores de Claude Code, Codex y Gemini, ni ningún `_test.go`
  existente.
- No cambia la plataforma: P-033 es su encargo *(SC-027)*.
- No ejecuta `opencode`, ni lanza `--run` o `--daemon` sobre la instalación real: los de los tests y C2 van en sandbox, y los de W1 y W2 los lanza
  el dueño.
- No escribe en ninguna copia congelada, ni pone su ruta en el repo. No versiona la base de 1 GB.
