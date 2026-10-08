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
