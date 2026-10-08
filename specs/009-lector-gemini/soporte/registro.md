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
