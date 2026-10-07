# 008 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones.

## B0 · Documentos

### T001 · Línea base *(2026-10-07, sobre `5f801c5`; sólo documentos desde `7b8c77c`)*

```
$ git log --oneline 7b8c77c..HEAD
5f801c5 008 B0: spec ratificada (E-1, E-2), contrato, plan, tareas y quickstart
$ git diff 7b8c77c --stat -- '*.go'                (vacío)
$ gofmt -l .                                        (vacío)
$ go vet ./...                                      (rc=0)
$ golangci-lint run
0 issues.
$ go test -count=1 -json ./...                      tests {'pass': 494} · paquetes {'pass': 9}
$ git diff 7b8c77c -- <frontera de FR-022> | wc -c
0
```

## B1 · Identidad

### T003 · Fase 0

`internal/ingest/codex_eventid.go` *(nuevo)*: `derivarEventIDCodex(responseID string) (string, bool)`, que devuelve `"", false`. Suite verde
en 9/9. `golangci-lint` marca la función como `unused` hasta que T004 la usa; en la Fase 0 sólo se exige la suite.

### T004 · Rojos, en `internal/ingest/codex_eventid_test.go` *(nuevo)*

```
--- FAIL: TestEventIDCodex_VectoresDelContrato/respuesta        derivarEventIDCodex("r-000000000000000000000001"): ok = false; se esperaba un event_id
--- FAIL: TestEventIDCodex_VectoresDelContrato/otra_respuesta   derivarEventIDCodex("r-000000000000000000000002"): ok = false; se esperaba un event_id
--- FAIL: TestEventIDCodex_EspacioDeNombresPropio               derivarEventIDCodex("r-000000000000000000000001"): ok = false; se esperaba un event_id
--- PASS: TestEventIDCodex_SinRespuestaNoHayEventID
```

**Razón**: la Fase 0 no deriva nada. La precondición de (2), que el literal de Claude Code sea el que da la derivación de 006
*(`hashEventID`)*, **pasa**.

**(3) nace verde**: la Fase 0 ya devuelve `("", false)`. La valida m3.

### T005 · Verde

`hashEventID` fija `claude_code` dentro de la función *(`eventid.go:67`)*, así que **no se puede reutilizar sin tocar `eventid.go`**. Se
replica el bucle de codificación en `codex_eventid.go`, reutilizando sólo sus constantes `dominioEventID` y `bytesEventID`. El vacío
devuelve `("", false)`.

- Los tres tests de T004, en **PASS**.
- **9/9 ok** y `golangci-lint run` → **0 issues**.

### T006 · Censo declarado ANTES de mutar *(2026-10-07)*

**Un cambio respecto a `tasks.md`**, declarado aquí antes de mutar. `tasks.md` preveía que **m2** tumbara también (2). **No lo hará**: (2)
compara el valor de Codex con el vector de 006, que sale de `hashEventID`, y m2 sólo altera la función nueva. El valor de Codex cambia, pero
sigue siendo distinto del de Claude Code. La co-caída prevista se retira.

| # | Mutación *(en `internal/ingest/codex_eventid.go`)* | Debe caer *(hojas)* |
|---|---|---|
| m1 | `herramientaCodex = "claude_code"` | `TestEventIDCodex_VectoresDelContrato/respuesta`, `TestEventIDCodex_VectoresDelContrato/otra_respuesta` |
| m2 | sin el prefijo de longitud *(se quita `h.Write(longitud[:])`)* | `TestEventIDCodex_VectoresDelContrato/respuesta`, `TestEventIDCodex_VectoresDelContrato/otra_respuesta` |
| m3 | el vacío se acepta *(se quita la guarda `responseID == ""`)* | `TestEventIDCodex_SinRespuestaNoHayEventID` |

`TestEventIDCodex_EspacioDeNombresPropio` **no debe caer con ninguna de las tres**.

### T007 · Resultado *(2026-10-07)*

**Las tres coinciden con lo declarado.** `TestEventIDCodex_EspacioDeNombresPropio` siguió verde en las tres. Cada una se revirtió por
edición inversa, y el md5 de `codex_eventid.go` volvió a ser el de antes: **`257c349d3e68e2395d88c6785539ad8a`**, antes y después de cada una.

| # | Cayó | Mensaje |
|---|---|---|
| m1 | `…VectoresDelContrato/respuesta`, `…/otra_respuesta` | `= "59999dcf4bf2fcacd45ec556d081cc88"; el contrato dice "31439e3953a0916dde2f98b748ceb985"` · `= "d92bde300578b96ca32672bbf8595443"; el contrato dice "e2c798be5f23e44680ab2c70690f8cf9"` |
| m2 | `…VectoresDelContrato/respuesta`, `…/otra_respuesta` | `= "49e9fee5dbaca403523f9c94d3184677"; el contrato dice "31439e3953a0916dde2f98b748ceb985"` · `= "6a4a17a8d8ce4e2e9e8bec4aaf6d4bbf"; el contrato dice "e2c798be5f23e44680ab2c70690f8cf9"` |
| m3 | `TestEventIDCodex_SinRespuestaNoHayEventID` | `derivarEventIDCodex("") = ("95a68fb7e1c843767c12c65687f28d85", true); se esperaba ("", false)` |
