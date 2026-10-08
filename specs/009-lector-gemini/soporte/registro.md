# 009 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007 y 008, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. Techo: **450**
líneas *(tasks §Presupuesto)*.

## B0 · Documentos

### T001 · Línea base *(2026-10-08, sobre `19cda8e`; sólo documentos desde `15ce93b`)*

```
$ git log --oneline -2
19cda8e 009 B0: spec ratificada, contrato, plan y tareas
15ce93b 009 descubrimiento: lector de Gemini CLI
$ go test -count=1 -v ./...                         574 «--- PASS», 0 «--- FAIL», 0 «--- SKIP» · 9 paquetes ok
$ golangci-lint run
0 issues.
$ git diff 15ce93b --stat -- <frontera de FR-027> | wc -l
0
```

## B1 · Identidad

### T003 · Fase 0

`internal/ingest/gemini_eventid.go` *(nuevo)*: `derivarEventIDGemini(id string) (string, bool)`, que devuelve `"", false`. `gofmt` y
`go vet` limpios, y la suite verde en 9/9. Como en 008 T003, sólo se exige la suite: el lint marcaría la función sin uso hasta T004.

### T004 · Rojos, en `internal/ingest/gemini_eventid_test.go` *(nuevo)*

```
gemini_eventid_test.go:29: derivarEventIDGemini("m-000000000000000000000001"): ok = false; se esperaba un event_id
gemini_eventid_test.go:29: derivarEventIDGemini("m-000000000000000000000002"): ok = false; se esperaba un event_id
--- FAIL: TestEventIDGemini_VectoresDelContrato
--- FAIL: TestEventIDGemini_VectoresDelContrato/respuesta
--- FAIL: TestEventIDGemini_VectoresDelContrato/otra_respuesta
gemini_eventid_test.go:46: derivarEventIDGemini("m-000000000000000000000001"): ok = false; se esperaba un event_id
--- FAIL: TestEventIDGemini_EspacioDeNombresPropio
--- PASS: TestEventIDGemini_SinIDNoHayEventID
```

**Razón**: la Fase 0 no deriva nada. La **precondición** de (2), que el literal de Codex `07d1f3ea…` sea el que da la derivación de 008
para `m-…0001`, **pasa**: (2) cae en `gemini_eventid_test.go:46`, la llamada a Gemini, no en `:42`. **(3) nace verde**: la acredita m3.

### T005 · Verde

La derivación del contrato, **replicada** en `gemini_eventid.go`. Reutiliza las constantes `dominioEventID`, `bytesEventID` y
`tipoRespuesta`. `eventid.go` y `codex_eventid.go`, sin diff.

```
--- PASS: TestEventIDGemini_VectoresDelContrato
    --- PASS: TestEventIDGemini_VectoresDelContrato/respuesta
    --- PASS: TestEventIDGemini_VectoresDelContrato/otra_respuesta
--- PASS: TestEventIDGemini_EspacioDeNombresPropio
--- PASS: TestEventIDGemini_SinIDNoHayEventID
$ golangci-lint run
0 issues.
$ md5sum internal/ingest/gemini_eventid.go
60ce5a10de029d5b02d1d8b10fa8880f  internal/ingest/gemini_eventid.go
```

### T006 · Censo, declarado ANTES de mutar *(por hoja; las co-caídas, en la misma fila)*

| # | Mutación, en `gemini_eventid.go` | Debe caer *(y sólo eso)* | Co-caídas |
|---|---|---|---|
| **m1** | `herramientaGemini = "codex"` | `TestEventIDGemini_VectoresDelContrato/respuesta`, `…/otra_respuesta`, `TestEventIDGemini_EspacioDeNombresPropio` | ninguna: ningún otro test usa `herramientaGemini` |
| **m2** | sin prefijo de longitud *(sólo `h.Write([]byte(c))`)* | `…/respuesta`, `…/otra_respuesta` | ninguna: (2) sigue verde, porque el valor sin prefijo tampoco es el de Codex |
| **m3** | sin la guarda del `id` vacío | `TestEventIDGemini_SinIDNoHayEventID` | ninguna |
| **m4** | el `sessionId` en el hash: `"s-000000000000000000000001"` antes del `id` | `…/respuesta`, `…/otra_respuesta` | ninguna. Informativo: `…/respuesta` debe mostrar `4b416a0d7c2b1ead78845a9bcdf62ca3`, el vector de rastro de P-2 (b) |

Cada mutación: aplicar → `go test -count=1 ./... 2>&1` → comparar el conjunto de `FAIL` con esta tabla → revertir por edición inversa →
md5 igual a `60ce5a10…`.

### T007 · Mutaciones *(2026-10-08; `go test -count=1 ./...` entero en cada una; sólo cae `internal/ingest`)*

```
## m1 aplicada: md5 ea20a3d6
--- FAIL: TestEventIDGemini_VectoresDelContrato/respuesta
--- FAIL: TestEventIDGemini_VectoresDelContrato/otra_respuesta
--- FAIL: TestEventIDGemini_EspacioDeNombresPropio
gemini_eventid_test.go:32: derivarEventIDGemini("m-000000000000000000000001") = "07d1f3ea70c0b94474fe27e18fdb89b7"; el contrato dice "027c36e1f42d164e670770cee1a06f83"
gemini_eventid_test.go:32: derivarEventIDGemini("m-000000000000000000000002") = "ddb495bbafa09172fcc3ec6358f6c1b2"; el contrato dice "53120e5f5cfd342c9c48f37cd857c697"
gemini_eventid_test.go:49: el event_id de Gemini coincide con el de Codex para el mismo valor ("07d1f3ea70c0b94474fe27e18fdb89b7")
revertida: md5 60ce5a10de029d5b02d1d8b10fa8880f

## m2 aplicada: md5 33606265
--- FAIL: TestEventIDGemini_VectoresDelContrato/respuesta
--- FAIL: TestEventIDGemini_VectoresDelContrato/otra_respuesta
gemini_eventid_test.go:32: derivarEventIDGemini("m-000000000000000000000001") = "c8b956947b5577989bef4f02ce7f3f2f"; el contrato dice "027c36e1f42d164e670770cee1a06f83"
gemini_eventid_test.go:32: derivarEventIDGemini("m-000000000000000000000002") = "c34df1390ff84b359cc0bb24ccfbc7e1"; el contrato dice "53120e5f5cfd342c9c48f37cd857c697"
revertida: md5 60ce5a10de029d5b02d1d8b10fa8880f

## m3 aplicada: md5 e3bbd4f0
--- FAIL: TestEventIDGemini_SinIDNoHayEventID
gemini_eventid_test.go:57: derivarEventIDGemini("") = ("2ffd3aa08a242f1266cac85cad8198bc", true); se esperaba ("", false)
revertida: md5 60ce5a10de029d5b02d1d8b10fa8880f

## m4 aplicada: md5 490c3b93
--- FAIL: TestEventIDGemini_VectoresDelContrato/respuesta
--- FAIL: TestEventIDGemini_VectoresDelContrato/otra_respuesta
gemini_eventid_test.go:32: derivarEventIDGemini("m-000000000000000000000001") = "4b416a0d7c2b1ead78845a9bcdf62ca3"; el contrato dice "027c36e1f42d164e670770cee1a06f83"
gemini_eventid_test.go:32: derivarEventIDGemini("m-000000000000000000000002") = "17506c70a3abdd4e5a86eda05f0b2666"; el contrato dice "53120e5f5cfd342c9c48f37cd857c697"
revertida: md5 60ce5a10de029d5b02d1d8b10fa8880f
```

| # | Declarado *(T006)* | Observado | |
|---|---|---|:--:|
| **m1** | `…/respuesta`, `…/otra_respuesta`, `EspacioDeNombresPropio` | las tres; `…/respuesta` da justo el vector de Codex | ✅ |
| **m2** | `…/respuesta`, `…/otra_respuesta` | las dos; (2) sigue verde | ✅ |
| **m3** | `SinIDNoHayEventID` | esa | ✅ |
| **m4** | `…/respuesta`, `…/otra_respuesta` | las dos; `…/respuesta` da `4b416a0d…`, el vector de rastro de P-2 (b), ahora también en Go | ✅ |

**Las cuatro coinciden.** Cada reversión deja `gemini_eventid.go` en `60ce5a10de029d5b02d1d8b10fa8880f`, y `cmp` con la copia previa a mutar no
da diferencias. El padre `TestEventIDGemini_VectoresDelContrato` cae con sus hojas; el censo va por hoja.

### T008 · Puertas del bloque *(antes del ✋ commit)*

```
$ gofmt -l .                                        (vacío)
$ go vet ./...                                      (rc=0)
$ golangci-lint run
0 issues.
$ go test -count=1 -v ./...                         579 «--- PASS» (574 + 5), 0 «--- FAIL», 0 «--- SKIP» · 9 paquetes ok
$ git diff 15ce93b --stat -- <frontera de FR-027> | wc -l
0
$ git diff 15ce93b --name-only --diff-filter=M -- '*_test.go' | wc -l
0
```

Commit previsto *(✋ dueño)*: `009 B1: event_id de Gemini con espacio de nombres propio`.

## B2 · Una aparición

### T009 y T010 · Fase 0 y fixtures

`internal/ingest/gemini.go` *(nuevo)*: `ContextoGemini` *(`Context`, `SessionID`, `ProjectRoot`)*, `ClaseGemini`, `MarcasGemini` *(`SinModelo`,
`TotalDescuadrado`)* y `RespuestaGemini(crudo, ctx)`, que aún no clasifica. Suite verde. **17 fixtures** en `testdata/gemini/`, una aparición
por fichero, con `id` `m-0000…`. **Corrección antes del verde** *(DECIDÍ YO)*: el mensaje de usuario lleva `tokens`; sin ellos, la hoja
«usuario» no acreditaba la guarda del `type`.

### T011 · Rojos, en `internal/ingest/gemini_respuesta_test.go` *(nuevo)*

T013 se sacó del fichero y se escribió **después** del verde, como pide la tarea.

```
--- FAIL (hojas): PartidasD2/partidas · PartidaAusenteValeCero · Incoherente/{tokens_no_objeto, partida_no_numerica,
    partida_negativa, cache_mayor_que_entrada, sin_timestamp, timestamp_mal_formado} · SinIdentificador/{ausente, vacio, no_textual} ·
    SinCoste · MomentoDelMensaje · SinModelo · TotalDescuadrado/{descuadrado, sin_total} · Referencias/{session_ref, project_ref,
    sin_project_root}                                                       (21 hojas; con los padres, 24 «--- FAIL»)
--- PASS: PartidasD2/{usuario_no_es_respuesta, tokens_null_no_es_respuesta, sin_tokens_no_es_respuesta}
gemini_respuesta_test.go:73: RespuestaGemini(respuesta_partidas.jsonl): clase 0, evento <nil>; se esperaba un evento      (y las 11 de evento)
gemini_respuesta_test.go:123: incoherente_negativa.jsonl: clase 0; se esperaba IncoherenteGemini (3)                       (y las 6 de (6))
gemini_respuesta_test.go:138: sin_identificador_no_textual.jsonl: clase 0; se esperaba SinIdentificadorGemini (2)         (y las 3 de (7))
```

**Razón**: la Fase 0 devuelve siempre `NoEsRespuestaGemini`, sin evento. **Nacen verdes** las tres hojas de «no es respuesta» de (4),
añadidas para que `RespuestaGemini` distinga «no es respuesta» de «incoherente» *(DE CAMINO de B1)*. Las acreditan M-B2a, M-B2c y M-B2d.

### T012 y T013 · Verde, y el verde de nacimiento

`RespuestaGemini` decodifica la envoltura **en crudo** (`json.RawMessage`): no es respuesta *(`type ≠ "gemini"`, o `tokens` ausente o `null`)*
→ sin identificador *(`id` no textual o vacío)* → incoherente *(`tokens` no decodifica a enteros, falla el `timestamp`, o negativas, o
`cached > input`)* → evento, con D-2 y las marcas. `total` es un puntero. Sin `internal/pricing`. **T013**, escrito después:
`NadaDelProveedorEnElEvento` *(centinelas en `id`, `content`, un `projectHash` metido en el mensaje, `SessionID` y `ProjectRoot`)* nace verde;
lo acredita m12.

```
28 «--- PASS» en TestRespuestaGemini_* · go test ./...: 607 «--- PASS», 0 FAIL, 0 SKIP · golangci-lint: 0 issues
$ md5sum internal/ingest/gemini.go
3c491b0dd8a69b6b592b3a7db59b6406  internal/ingest/gemini.go
```

### T014 · Censo, declarado ANTES de mutar *(por hoja; sólo puede caer `internal/ingest`)*

| # | Mutación, en `gemini.go` | Debe caer *(y sólo eso)* | Co-caídas y por qué no más |
|---|---|---|---|
| **m5** | `TokensInput = Input − Cached` *(sin `tool`)* | `PartidasD2/partidas` | ninguna: sólo ese fixture tiene `tool` ≠ 0 |
| **m6** | `TokensOutput = Output` *(sin `thoughts`)* | `PartidasD2/partidas` | ninguna: sólo ese tiene `thoughts` ≠ 0 |
| **m7** | `TokensInput = Input + Tool` *(sin restar `cached`)* | `PartidasD2/partidas` | ninguna: en (5), `cached` = 0 |
| **m8** | `CostAvailable = true` | `SinCoste` | ninguna |
| **m9** | la condición de incoherente, con `&& false` | `Incoherente/` × 6 | ninguna: el resto de fixtures son coherentes |
| **m10** | `OccurredAt = time.Now()` | `MomentoDelMensaje` | ninguna |
| **m11** | descuadrado → `IncoherenteGemini` | `TotalDescuadrado/descuadrado` | ninguna: los demás `total` cuadran *(115, 11, 11)* |
| **m12** | `SessionRef = ctx.SessionID` *(sin sal)* | `Referencias/session_ref`, `NadaDelProveedorEnElEvento` | — |
| **M-B2a** | `tokens: null` tomado como respuesta *(`ausente` → `len == 0`)* | `PartidasD2/tokens_null_no_es_respuesta` | ninguna: `null` decodifica a ceros y sale un evento |
| **M-B2b** | `id` decodificado como `string` | `SinIdentificador/no_textual` | ninguna: la envoltura entera deja de decodificar sólo con `"id": 7` |
| **M-B2c** | sin la guarda del `type` | `PartidasD2/usuario_no_es_respuesta` | ninguna |
| **M-B2d** | sin la guarda de `tokens` ausente | `PartidasD2/sin_tokens_no_es_respuesta`, `PartidasD2/tokens_null_no_es_respuesta` | ausente → incoherente; `null` → evento |

Cada mutación: aplicar → `go test -count=1 ./... 2>&1` → comparar el conjunto de `FAIL` por hoja → revertir por edición inversa → md5
igual a `3c491b0d…`.

### T015 · Mutaciones *(2026-10-08; `go test -count=1 ./...` entero en cada una; sólo cae `internal/ingest`; las hojas, sin el prefijo `TestRespuestaGemini_`)*

| # | md5 mutado | Observado *(hojas que caen)* | Declarado | md5 tras revertir |
|---|---|---|:--:|---|
| **m5** | `20f8cc48` | `PartidasD2/partidas` | ✅ | `3c491b0d…` |
| **m6** | `e1f702e2` | `PartidasD2/partidas` | ✅ | `3c491b0d…` |
| **m7** | `bd19cf5f` | `PartidasD2/partidas` | ✅ | `3c491b0d…` |
| **m8** | `e0f72675` | `SinCoste` | ✅ | `3c491b0d…` |
| **m9** | `b410f0ea` | `Incoherente/` × 6 | ✅ | `3c491b0d…` |
| **m10** | `67507957` | `MomentoDelMensaje` | ✅ | `3c491b0d…` |
| **m11** | `87c34c77` | `TotalDescuadrado/descuadrado` | ✅ | `3c491b0d…` |
| **m12** | `f6f4f067` | `Referencias/session_ref`, `NadaDelProveedorEnElEvento` | ✅ | `3c491b0d…` |
| **M-B2a** | `6df3fbf7` | `PartidasD2/tokens_null_no_es_respuesta` | ✅ | `3c491b0d…` |
| **M-B2b** | `865c6f56` | `SinIdentificador/no_textual` *(«json: cannot unmarshal number into Go struct field mensajeGemini.id of type string»: lo que habría sido un falso «corrupto»)* | ✅ | `3c491b0d…` |
| **M-B2c** | `52cf890e` | `PartidasD2/usuario_no_es_respuesta` | ✅ | `3c491b0d…` |
| **M-B2d** | `0a122bfb` | `PartidasD2/tokens_null_no_es_respuesta`, `PartidasD2/sin_tokens_no_es_respuesta` | ✅ | `3c491b0d…` |

**Las doce coinciden.** Ningún paquete más que `internal/ingest` cae, y ninguna mutación deja de compilar. M-B2b toca dos líneas *(el tipo
del campo y la llamada)*, y se revierte con las dos ediciones inversas. Tras la última, `cmp` con la copia previa a mutar no da diferencias.

### T016 · Puertas del bloque *(antes del ✋ commit)*

```
gofmt -l . (vacío) · go vet ./... (rc=0) · golangci-lint run (0 issues) · go test -count=1 -v ./... 607 «--- PASS» (579 + 28), 0 FAIL, 0 SKIP
git diff 15ce93b --stat -- <frontera de FR-027> | wc -l → 0 · git diff 15ce93b --name-only --diff-filter=M -- '*_test.go' | wc -l → 0
grep -n pricing internal/ingest/gemini.go → sólo :24, un comentario
```

**Presupuesto** *(tasks §Presupuesto)*: producción 154 líneas *(99 sin comentarios ni blancos)*, frente a ~110; test 221, frente a ~220; 17
fixtures. Esta sección del registro mide **94** líneas frente a ≤ 60: **+57 %, se declara**. Lo pagan los bloques siguientes, y el techo de
450 no se mueve.

Commit previsto *(✋ dueño)*: `009 B2: una respuesta de Gemini es un evento sin coste`.

## B3 · Fichero y contexto

### T017–T019 · Fase 0, fixtures y rojos

`internal/ingest/gemini_contexto.go` *(nuevo)*: `PasadaGemini`, `FicheroGemini{Ruta, Slug}`, `ListarGemini(raiz)`,
`LeerFicheroGemini(st, f FicheroGemini, …)` y `ContarAnteriorGemini`, que no hacen nada. Suite verde. **10 fixtures** en
`testdata/gemini/contexto/`, generados por un guion fuera del repo. `forma_copia.jsonl` *(53 líneas: herramientas, compresión, reanudación)*
da, con un contador independiente, 16 · 10 · 6 de una vez, y 12 · 8 · 4 + 4 · 2 · 2 cortado en la línea 42. El árbol de `ListarGemini` y los
`.project_root` se montan en `t.TempDir()`. Rojos en `gemini_contexto_test.go` *(nuevo)*: **31 «--- FAIL»**, todos porque el esqueleto no
devuelve nada (`respuestas 0 · eventos 0 · repetidas 0`, `Resumen() = ""`, `sesiones ""`…). **Nacen verdes** `ContarAnteriorGemini/segunda`
*(la acredita m18)* y `PasadaGemini_ResumenSC008/identidad` *(con ceros se cumple; la acredita M-B3b, nueva)*.

### T020 · Verde

El prefijo de D-009-P1, la cabecera de D-009-P2, la caché de `.project_root` de D-009-P3, el orden de D-009-P4 y el `.json` de D-009-P5.
`Recorrer` va tal cual. Repetida = `event_id` en `emitidos` o en `vistos`. **Dos correcciones antes del verde completo** *(DECIDÍ YO)*:
- `ListarGemini` ordena cada grupo **por ruta completa**, como dice D-009-P4; mi primer verde seguía el orden de recorrido;
- `truncado.jsonl` tenía un `timestamp` con el segundo 61, que es inválido. El guion usa `%60`, y sólo cambió ese fichero.

```
33 «--- PASS» en los tests de B3 · go test ./...: 640 «--- PASS», 0 FAIL, 0 SKIP · golangci-lint: 0 issues
$ md5sum internal/ingest/gemini_contexto.go
945acb6e4c34a13183529b6106624adf  internal/ingest/gemini_contexto.go
```

**SC-013, medida informativa**. Es un sintético de 105 906 784 B en el temporal, nunca en el repo: 140 529 líneas, 70 089 respuestas y 350
`$set.messages` de hasta 125 165 B con el historial vivo y sus tokens. Una primera pasada entera, y después tres pasadas con **1** respuesta
nueva cada una, con un test temporal que se borró. Máquina: i5-13400, Linux.

```
primera pasada entera: 70089 eventos en 2.046418034s
medida 1: 1 evento nuevo, respuestas 1 · eventos 1 · repetidas 0, en 0.434 s
medida 2: 1 evento nuevo, respuestas 1 · eventos 1 · repetidas 0, en 0.445 s
medida 3: 1 evento nuevo, respuestas 1 · eventos 1 · repetidas 0, en 0.436 s
```

**≤ 3 s en las tres** *(0,434 · 0,445 · 0,436 s)*. El plan B no hace falta.

### T021 · Censo, declarado ANTES de mutar *(por hoja; sin el prefijo `TestFicheroGemini_`, salvo los de otro nombre)*

| # | Mutación, en `gemini_contexto.go` | Debe caer *(y sólo eso)* |
|---|---|---|
| **m13** | sin prefijo *(`false &&` en su condición)* | `ReanudacionEnDosPasadas/segunda`, `DosPasadasIgualAUna`, `CabeceraEnElPrefijo` |
| **m14** | el prefijo no guarda `vistos` | `ReanudacionEnDosPasadas/segunda` |
| **m15** | un `$set.messages` vacía los eventos ya acumulados del fichero | `Apariciones`, `SetFinalNoBorra`, `DosPasadasIgualAUna` |
| **m16** | una aparición sin tokens mete su `event_id` en `emitidos` | `TokensTardios` |
| **m17** | sin repetidas *(`false &&`)* | `Apariciones`, `RepetidasEnLaPasada/herramienta`, `…/dos_carpetas`, `ReanudacionEnDosPasadas/primera`, `…/segunda`, `CabeceraEnElPrefijo`, `PasadaGemini_ResumenSC008/literal`, `Proyecto/con`, `Proyecto/sin`, `Proyecto/subagente` |
| **m18** | el `.json` en cada pasada | `ContarAnteriorGemini/segunda` |
| **m19** | el prefijo avisa de la línea corrupta *(con `avisos` en su firma)* | `LineaCorrupta/segunda` |
| **m20** | orden sólo léxico, sin grupos | `ListarGemini/patron`, `ListarGemini/orden` |
| **m21** | `Slug = filepath.Dir(dir)` *(el del subagente queda en `chats/`)* | `ListarGemini/patron`, `Proyecto/subagente` |
| **m22** | sin `.project_root` es error | `RepetidasEnLaPasada/dos_carpetas`, `Proyecto/sin` |
| **M-B3a** | el truncado lee prefijo *(`offset > 0` a secas)* | `TruncadoSinPrefijo` |
| **M-B3b** | una repetida no suma a `Respuestas` | `Apariciones`, `RepetidasEnLaPasada/herramienta`, `…/dos_carpetas`, `ReanudacionEnDosPasadas/primera`, `…/segunda`, `PasadaGemini_ResumenSC008/literal`, `…/identidad` |

**Co-caídas más allá de `tasks.md`, declaradas aquí**:
- **m17** hace caer todo test que cuenta repetidas o exige un solo evento de una respuesta que se repite;
- **m15** hace caer los que miran los eventos de un fichero con `$set.messages`;
- **m20** y **m21** hacen caer `patron`, que compara el listado completo;
- **m22** hace caer `dos_carpetas`, que tiene una carpeta sin `.project_root`.

md5 de referencia: `945acb6e…`.

### T022 · Mutaciones *(`go test -count=1 ./...` entero en cada una; sólo cae `internal/ingest`; ninguna deja de compilar)*

| # | md5 mutado | Observado | Declarado | md5 tras revertir |
|---|---|---|:--:|---|
| **m13** | `99929e13` | `CabeceraEnElPrefijo`, `DosPasadasIgualAUna`, `ReanudacionEnDosPasadas/segunda` | ✅ | `945acb6e…` |
| **m14** | `1f2c25b6` | `ReanudacionEnDosPasadas/segunda` | ✅ | `945acb6e…` |
| **m15** | `e8a155cc` | `Apariciones`, `DosPasadasIgualAUna`, `SetFinalNoBorra` | ✅ | `945acb6e…` |
| **m16** | `37678b74` | `TokensTardios` | ✅ | `945acb6e…` |
| **m17** | `64815156` | `Apariciones`, `CabeceraEnElPrefijo`, `Proyecto/con`, `Proyecto/sin`, `Proyecto/subagente`, `ReanudacionEnDosPasadas/primera`, `ReanudacionEnDosPasadas/segunda`, `RepetidasEnLaPasada/dos_carpetas`, `RepetidasEnLaPasada/herramienta`, `ResumenSC008/literal` | ✅ | `945acb6e…` |
| **m18** | `62167e2b` | `ContarAnteriorGemini/segunda` | ✅ | `945acb6e…` |
| **m19** | `3441ac96` | `LineaCorrupta/segunda` | ✅ | `945acb6e…` |
| **m20** | `8215b998` | `ListarGemini/orden`, `ListarGemini/patron` | ✅ | `945acb6e…` |
| **m21** | `6ee65e7a` | `Proyecto/subagente`, `ListarGemini/patron` | ✅ | `945acb6e…` |
| **m22** | `cae39177` | `Proyecto/sin`, `RepetidasEnLaPasada/dos_carpetas` | ✅ | `945acb6e…` |
| **M-B3a** | `42810f7d` | `TruncadoSinPrefijo` | ✅ | `945acb6e…` |
| **M-B3b** | `6119ff17` | `Apariciones`, `ReanudacionEnDosPasadas/primera`, `ReanudacionEnDosPasadas/segunda`, `RepetidasEnLaPasada/dos_carpetas`, `RepetidasEnLaPasada/herramienta`, `ResumenSC008/identidad`, `ResumenSC008/literal` | ✅ | `945acb6e…` |

**Las doce coinciden**, hoja a hoja, con T021. Tras la última, `cmp` con la copia previa a mutar no da diferencias.

### T023 · Puertas del bloque *(antes del ✋ commit)*

```
gofmt -l . (vacío) · go vet ./... (rc=0) · golangci-lint run (0 issues) · go test -count=1 -v ./... 640 «--- PASS» (607 + 33), 0 FAIL, 0 SKIP
frontera de FR-027 → 0 · '*_test.go' existentes modificados → 0 · codex.go, codex_contexto.go, gemini.go e internal/state/ → sin cambios
```

**Presupuesto**: producción 326 líneas, frente a ~210 *(+55 %: se declara)*; test 458, frente a ~380; 10 fixtures; esta sección,
**91** líneas frente a ≤ 90. Commit previsto *(✋ dueño)*: `009 B3: apariciones, contexto entre pasadas y formato de Gemini`.

## B4 · Raíz

**T024–T027** · `config.GeminiRoot()` *(nuevo, `internal/config/gemini.go`)*: el esqueleto devolvía `"", nil`. **Rojos**, en ficheros nuevos:
(26) ×3 y (27) en `internal/config/gemini_test.go`, que reutiliza `hogarDePrueba` y `mkdir` de `codex_test.go`, y añade `sinGeminiHome`; y
**(27-bis)** *(T026.1, del orquestador)* en `internal/testutil/sandbox_gemini_test.go`. Los cinco cayeron con `("", <nil>)` frente a la ruta
esperada, y (27-bis) con la variable aún definida. **Verde**: la regla de D-009-P9. El comentario de `codex.go:16` pasa a «una de las DOS lecturas
de entorno de producción, con `GEMINI_CLI_HOME`», y `git diff -U0` sólo toca `:16-17`, que son comentario. `sandbox.go` gana
`t.Setenv("GEMINI_CLI_HOME", "")` tras `:63`. `sandbox_test.go` sigue en verde y sin tocar. Suite: **646** «--- PASS» *(640 + 6)*, 0 FAIL,
0 SKIP; lint 0. `os.Getenv` de producción → **2**: `internal/config/codex.go:19` *(`CODEX_HOME`)* y `internal/config/gemini.go:20`
*(`GEMINI_CLI_HOME`)*.

**T028 · Censo, ANTES de mutar.** md5 `gemini.go` = `2825c479…` y `sandbox.go` = `82639521…`.

| # | Mutación | Debe caer *(y sólo eso)* | Observado | md5 mutado | md5 tras revertir |
|---|---|---|---|---|---|
| **m23** | `GEMINI_CLI_HOME=""` tomada como raíz *(`LookupEnv`)* | `TestGeminiRoot_Raiz/vacia` | `TestGeminiRoot_Raiz/vacia` ✅ | `003ffa86` | `2825c479…` |
| **m24** | ignorar `GEMINI_CLI_HOME` | `TestGeminiRoot_Raiz/definida` | `TestGeminiRoot_Raiz/definida` ✅ | `677f99ac` | `2825c479…` |
| **m25** | exigir que la raíz exista | `TestGeminiRoot_InexistenteSinError` | `TestGeminiRoot_InexistenteSinError` ✅ | `463efd6b` | `2825c479…` |
| **M-B4a** | quitar la línea nueva de `sandbox.go` | `TestSandbox_VaciaGeminiCliHome` | `TestSandbox_VaciaGeminiCliHome` ✅ | `31530de9` | `82639521…` |

**Las cuatro coinciden**; sólo cae el paquete de la hoja *(`config` o `testutil`)*. `cmp` con las copias previas: idénticos. **T029 · Puertas**:
`gofmt`, `vet` y lint a 0; **646** pass, 0 FAIL, 0 SKIP; frontera → 0; `*_test.go` existentes modificados → 0. Commit previsto *(✋ dueño)*:
`009 B4: raiz de Gemini con GEMINI_CLI_HOME`.

## B5 · Integración

**T030 · Fase 0**, en este orden. **(1)** La referencia de SC-012, antes de tocar `main.go` *(`git diff --quiet` → limpio)*. El binario de
`15ce93b` sale de `git archive 15ce93b | tar -x` + `go build` en un temporal. Se ejecuta en `env -i`, con `HOME`, `USERPROFILE` y
`XDG_CONFIG_HOME` temporales, `logs_root` en `claude.jsonl`, `CODEX_HOME` en `sessions/2026/10/07/rollout-a.jsonl` *(`sesion.jsonl`)*, y sin
`GEMINI_CLI_HOME` ni `.gemini`. Sale `rc=0` y nada en stdout. Con `<DATOS>`, queda en `cmd/permea/testdata/gemini/referencia-run.stderr`,
md5 **`aa3314e9838f192829743cc3501f4102`**: siete líneas, la de Codex con `respuestas 2 · eventos 2`. **(2)** El forzado A, con un test temporal
ya borrado: `Load` → `<nil>`; `Append` → `<nil>`, 323 bytes; `Save` → `permission denied`; uid 1000. **(3)** El esqueleto: `agent.geminiRaiz`,
`agent.gemini`, `setup()` con `config.GeminiRoot()` *(`gofmt` realinea el literal)* y `generarGemini` vacío en `cmd/permea/gemini.go` *(nuevo)*.

**T031–T036 · Rojos**, en `cmd/permea/gemini_test.go` *(nuevo)*, con los ayudantes de `codex_test.go` sin tocarlo. Caen 11 «--- FAIL»: 0 eventos
`gemini`. **Nacen verdes** (28) ×2 *(la acredita m26)*, `SoloConNovedades/predicado` *(M-B5a, nueva)* y `FicheroIlegibleNoRompe/segunda_pasada`
*(m31)*. **Verde**: Gemini entre Codex y el único `st.Save`, con `os.Stat(<raíz>/tmp)` en cada pasada, y las líneas de `runOnce` y `tick`.
**662** pass *(646 + 16)*, 0 FAIL, 0 SKIP; lint 0. Los tests existentes, verdes y sin tocar. md5 `main.go` = `c04ec343…`, `gemini.go` = `f2ccec01…`.

**T037 · Censo, ANTES de mutar** *(hojas sin `TestGemini`; las de Codex, con su nombre)*:

| # | Mutación | Debe caer *(y sólo eso)* | Observado | md5 mutado |
|---|---|---|---|---|
| **m26** | `a.gemini = NuevaPasadaGemini()` siempre | `Run_SinRaiz…/gemini_cli_home_vacia`, `…/sin_tmp`, `Activacion_EnCadaPasada`, `TestCodexRun_SinRaiz…/codex_home_vacia`, `…/raiz_inexistente`, `TestCodexRun_LineaDeResumen` | = declarado ✅ | `38fb4c5d` |
| **m27** | un `st.Save` antes de Gemini | `Generate_EncolaAntesDeGuardar` | = declarado ✅ | `95286d8b` |
| **m28** | `tick` sin `HayNovedades()` | `Demonio_SoloConNovedades/tick` | = declarado ✅ | `5d4222f6` |
| **m29** | Gemini sólo con logs de Claude | `Run_SoloGemini`, `Generate_EncolaAntesDeGuardar`, `Demonio…/tick`, `Run_SegundaPasadaCero`, `Run_NadaDelProveedorViaja`, `Activacion_EnCadaPasada` | = declarado ✅ | `0b620a9b` |
| **m30** | la existencia sólo en `setup()` | `Activacion_EnCadaPasada` | = declarado ✅ | `42b9fd2e` |
| **m31** | el error de Gemini aborta la pasada | `FicheroIlegibleNoRompe/pasada`, `…/segunda_pasada`, `…/se_relee` | = declarado ✅ | `70e969fe` |
| **m32** | el omitido guarda su offset al final | `FicheroIlegibleNoRompe/se_relee` | = declarado ✅ | `5ab77917` |
| **m33** | la ruta del fichero en `SessionRef` | `Run_NadaDelProveedorViaja` | = declarado ✅ | `d3b030c5` |
| **m34** | la línea `gemini:` antes que la de Codex | `Run_LineaDeResumen` | = declarado ✅ | `ab90a1e8` |
| **M-B5a** | `HayNovedades` sin formato anterior *(en `gemini_contexto.go`)* | `Demonio…/predicado`, `TestPasadaGemini_ResumenSC008/hay_novedades` | = declarado ✅ | `e69f6380` |

**T038 · Las diez coinciden**, hoja a hoja. Cada reversión devuelve `main.go` a `c04ec343…`, `gemini.go` a `f2ccec01…` y
`gemini_contexto.go` a `945acb6e…`, y `cmp` con las copias previas no da diferencias. **Incidente** *(de mi guion, no del código)*: tras
mutar m28, `if a.gemini != nil {` quedaba dos veces, y la reversión automática no pudo hacerse. m29 y m30 corrieron **encima** de m28, y m30
mostró de más justo `…/tick`, la caída de m28. Se deshizo m28 con una edición inversa de patrón único *(md5 de vuelta a `c04ec343…`)*, se
descartaron esas dos medidas y m28–m30 se repitieron limpias: son las de la tabla. **T039 · Puertas**: `gofmt`, `vet` y lint a 0; **662**
pass, 0 FAIL, 0 SKIP; frontera → 0; `*_test.go` existentes → 0. Commit *(✋ dueño)*: `009 B5: Gemini en run y daemon, antes de guardar el estado`.

## B6 y B7 · `--scan`, README y CHANGELOG

**T040–T041** · Antes de nada, las referencias de la 0.5.0 con el binario de `15ce93b` *(`git archive` + `go build`, `env -i`)*: `--scan` de
`testdata/codex/sesion.jsonl` → `testdata/gemini/referencia-scan-codex.stdout` *(`100aa389…`)* y `.stderr` *(`b530a31f…`)*. La de Claude es
byte a byte la de la 0.4.0 *(`cmp`)*, y se reutiliza. **Rojos** en `gemini_test.go`: (37) cae en 4 hojas, porque el fichero de Gemini aún se lee
como Claude Code *(0 eventos)*. **Nacen verdes** `nada_en_disco` *(m37)* y (38) ×2 *(m36 y m38)*. **Verde**: `esSesionGemini` *(cabecera:
`sessionId` y `projectHash`, sin `type`)* y `dryRunGemini`, en `gemini.go`; en `dryRun`, tras Codex. **DECIDÍ YO** *(`slugDeScan`)*: con
`--scan` el fichero llega suelto, así que el `<slug>` sale de la posición. Padre `chats` → abuelo; abuelo `chats` *(subagente)* → bisabuelo; si no,
`project_ref` vacío. Es la única deducción por nombre del lector, y sólo en `--scan`. **671** pass, 0 FAIL, 0
SKIP; lint 0. md5 `main.go` = `89e1b4ec…`, `gemini.go` = `4c3e95d7…`.

**T042 · Censo, ANTES de mutar.** `Scan_*` de Claude son los cuatro `TestScan_*` *(`main_test.go:478,497`, `coste_test.go:26,49`)*.

| # | Mutación | Debe caer *(y sólo eso)* | Observado | md5 mutado |
|---|---|---|---|---|
| **m35** | `if false && gemini` en `dryRun` | `GeminiScan_EventosYResumen/{lineas_evento, resumen, proyecto, sin_forma}` | = declarado ✅ | `1b89bd13` |
| **m36** | `esSesionGemini` → `true` | `GeminiScan_ClaudeYCodexComoLa050/claude`, `TestCodexScan_ClaudeCodeComoLa040`, los cuatro `TestScan_*` | = declarado ✅ | `91bbc7d4` |
| **m37** | `dryRunGemini` guarda `state.json` | `GeminiScan_EventosYResumen/nada_en_disco` | = declarado ✅ | `fffb321c` |
| **m38** | detección sólo por `sessionId` | lo mismo que m36: las primeras líneas de Claude llevan `sessionId` | = declarado ✅ | `cd4b0918` |

Las cuatro coinciden, y el md5 se comprobó tras cada reversión, **antes** de la siguiente mutación *(corrección de método del dueño)*. El guion,
además, rechaza un reemplazo que ya existiera en el fichero, que es la causa del incidente de m28 en B5.

**T043 ✋** *(el dueño, 2026-10-08, 18:30)*: las nueve frases de coherencia, a spec §Textos aprobados *(E-1)*. **Tarifas, antes de la frase 9**:
con `git show 6569815:backend/config/pricing.php` *(sólo lectura)*, las 17 filas `claude-*` de la plataforma y las de `internal/pricing`, con las cinco
cifras de cada una, son **17/17 iguales**. La plataforma tiene además 11 filas `gpt-*`. **T044–T046**: los rojos dieron 0 y 0. «### Gemini CLI» y
`## 0.6.0 — PENDIENTE` se sacaron por programa de la spec, y `cmp` no da diferencias. Las nueve frases dan 9/9 con `grep -F` sobre el texto normalizado
*(la 2, sobre la línea tal cual)*. Las líneas 187, 201 y 243 de B5 están ahora en 197, 211 y 253, porque la sección nueva las movió. `grep -c` →
`^## 0.6.0` 1, `^### Gemini CLI` 1 y `PENDIENTE` 1; `Claude Code y Codex CLI` → 0. **Puertas**: **671** pass, 0 FAIL, 0 SKIP; lint 0; frontera → 0.
Commit *(✋ dueño, T047)*: `009 B6 y B7: scan de Gemini, README y CHANGELOG de la 0.6.0`.

## Cierre *(2026-10-08; binario de la rama, `4b7f88e`; Go 1.22.2, golangci-lint 2.12.2, goreleaser v2.16.0)*

| **C1** · Puertas | `gofmt`, `vet` y lint 0 · **671** pass, 0 FAIL, 0 SKIP · frontera de FR-027 → 0 B · `*_test.go` existentes modificados → 0 · `internal/state/`, `event/` y `pricing/` → sin diff · `os.Getenv` → 2 *(`codex.go:19`, `gemini.go:20`)* · `windows`/`darwin` × `amd64`/`arm64` → rc 0 ×4 · `PENDIENTE` → 1 · `15ce93b..HEAD` → B0…B7 *(7 commits)* ✅ |
|---|---|
| **C2** · Medidas | Huellas de las dos copias, antes = después *(3 + 3)*, y 0 ficheros modificados en ellas. **Contador** sobre la primera: 16 · 10 · 6, y 95 747 / 12 141 / 0 / 2 482 *(9 + 1)*. **`--scan`** en su sitio: 10 `evento:`; con `awk`, las mismas sumas; `gemini:` literal; 0 `project_ref` vacíos; nada en el hogar. **`--run` ×2** con `GEMINI_CLI_HOME` en la segunda copia y `CODEX_HOME=""`: la 1.ª, 10 `gemini` *(9 y 1)*, mismas sumas, entrada + caché **107 888**, 10/10 sin coste; la 2.ª, 0. `state.json` y la cola, sólo en el sandbox *(SC-001–SC-005)* ✅ |
| **SC-013** | Binario entero, 105 906 784 B *(70 089 respuestas, 350 `$set.messages`)*. Las tres medidas: **2,30 · 2,23 · 2,09 s** ≤ 3 s ✅. Diagnóstico: justo después de escribir ~140 MB, con 140 MB libres en WSL, hubo 1,5–3,25 s, y una de 5,98 s. En reposo, **8 medidas de 0,44 a 0,59 s**. La varianza es de E/S de WSL, no del lector *(sin bytes nuevos: 0,03–0,10 s)* |
| **C3** · Snapshot | `rc=0`, «skipping announce, publish, and validate»; la versión inyectada, `0.5.0-SNAPSHOT-4b7f88e` *(desde la etiqueta `v0.5.0`)*; `sha256sum -c` OK ×5. En `permea.exe`, el resumen `gemini:`, `gemini: fichero omitido: %v` y la línea `evento:`, byte a byte *(1 vez cada uno; `strings` no ve el resumen porque el `·` no es ASCII)*. Zip de Windows `8a478701…e400`, copiado para W1 ✅ |
