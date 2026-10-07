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

### Remate de B1 · M-B1a *(Encargo 5, 2026-10-07)*

**Por qué**: con m1–m3, `TestEventIDCodex_EspacioDeNombresPropio` no cae nunca. Una mutación que deriva en el espacio de Claude Code
comprueba que muerde.

**Censo declarado ANTES de mutar**:

| # | Mutación *(en `internal/ingest/codex_eventid.go`)* | Debe caer *(hojas)* |
|---|---|---|
| M-B1a | tras la guarda del vacío, `return hashEventID(tipoSoloMessageID, responseID), true` | `TestEventIDCodex_VectoresDelContrato/respuesta`, `TestEventIDCodex_VectoresDelContrato/otra_respuesta`, `TestEventIDCodex_EspacioDeNombresPropio` |

`TestEventIDCodex_SinRespuestaNoHayEventID` no debe caer, porque la guarda se queda. **Si (2) no cae, se para.**

**Resultado: coincide.** Cayeron las tres hojas declaradas:
- `…/respuesta`: `= "128d67bd072f4488bc3854e82bbcaeb2"; el contrato dice "31439e3953a0916dde2f98b748ceb985"`;
- `…/otra_respuesta`: `= "143e55305c8a840b37397a9c7fd619cd"; el contrato dice "e2c798be5f23e44680ab2c70690f8cf9"`;
- `EspacioDeNombresPropio`: `el event_id de Codex coincide con el de Claude Code para el mismo valor ("128d67bd072f4488bc3854e82bbcaeb2")`.

`SinRespuestaNoHayEventID` siguió verde. `go vet` avisó `unreachable code` *(el código tras el `return` mutado)*; compiló y los tests
corrieron, así que la mutación cuenta. Se revirtió por edición inversa, y el md5 es **`257c349d3e68e2395d88c6785539ad8a`**, antes y después.

## B2 · Una línea

### Enmienda E-3 *(orquestador, 2026-10-07; registrada en `spec.md`)*

«Incoherente» cubre también un registro **sin `usage`** y uno con **alguna partida negativa**. Se aplica a las cuatro partidas de FR-010
*(`DECIDÍ YO`)*. El rojo (6) tiene una hoja por caso.

### T009 · Fase 0

`internal/ingest/codex.go` *(nuevo)* con `ContextoCodex`, `ClaseCodex` y `LineaCodex`, que devuelve `(nil, NoEsRegistro, nil)`.
- `ContextoCodex` embebe el `Context` de Claude Code *(sal, máquina, dev, org, versión y resolutor)*.
- Añade `DelTurno(turnID) (modelo, cwd)`, que pondrá B3.

Suite verde en 9/9.

### T010 · Fixtures sintéticos, en `internal/ingest/testdata/codex/`

Nueve ficheros de una línea, generados por programa:
- `registro_simple` *(13 831 / 11 008 / 0 / 5, las partidas publicadas de F7)*;
- `registro_escritura` *(100 / 40 / 60 / 10)*;
- tres `incoherente_*` *(caché + escritura > entrada; sin `usage`; salida −1)*;
- dos `sin_identificador_*` *(ausente y vacío)*;
- `centinelas`;
- `no_registro` *(un `turn_context`)*.

Identificadores inventados, con `session_id` ≠ `thread_id` ≠ `turn_id`.

### T011 · Rojos, en `internal/ingest/codex_linea_test.go` *(nuevo)*

```
--- FAIL: TestLineaCodex_UnRegistroEsUnEvento/registro                         LineaCodex(registro_simple.jsonl): clase 0, evento <nil>; se esperaba un evento
--- PASS: TestLineaCodex_UnRegistroEsUnEvento/otra_linea_no_es_registro        (nace verde; la valida M-B2b)
--- FAIL: TestLineaCodex_EscrituraDeCacheDentroDeLaEntrada                     LineaCodex(registro_escritura.jsonl): clase 0, evento <nil>; se esperaba un evento
--- FAIL: TestLineaCodex_Incoherente/{cache_y_escritura_mayores_que_la_entrada, sin_usage, partida_negativa}
                                                                              (<nil>, clase 0, <nil>); se esperaba (nil, Incoherente, nil)
--- FAIL: TestLineaCodex_SinIdentificador/{ausente, vacio}                     (<nil>, clase 0, <nil>); se esperaba (nil, SinIdentificador, nil)
--- FAIL: TestLineaCodex_SinCoste · _MomentoDelRegistro · _SesionDelRegistro   LineaCodex(registro_simple.jsonl): clase 0, evento <nil>; se esperaba un evento
```

**Razón**: la Fase 0 no clasifica nada.

**Lo que añade este fichero sobre el plan**: la hoja `otra_linea_no_es_registro` de (4), que cubre «ningún otro tipo produce eventos»
*(FR-006)*, y su mutación **M-B2b**.

### T012 · Verde

- **Decodificación en dos pasos**: primero `{type, timestamp, payload}` en crudo, y sólo un `token_usage_record` decodifica su `payload`.
  Así, una línea válida de otro tipo y otra forma no cuenta como corrupta.
- **Partidas de FR-010**.
- **Clasificación**: sin identificador → incoherente *(FR-027, E-3)* → evento.
- **El evento**: `tool = codex`, coste 0 y `cost_available = false`; `event.Ref` para la sesión y la máquina; el resolutor para el
  proyecto. **No** se consulta `internal/pricing`.
- **Resultado**: los 11 tests y hojas de T011, en **PASS**. Lint: 4 avisos de `revive` en los comentarios de las constantes, corregidos;
  queda en `0 issues`. Suite 9/9.

### T013 · Verde de nacimiento

`TestLineaCodex_NingunIdentificadorDelProveedorEnElEvento`, escrito **tras** T012, en **PASS**. Lo valida M-B2a.

### T014 · Censo declarado ANTES de mutar *(2026-10-07)*

**Dos ajustes respecto a `tasks.md`, declarados aquí antes de mutar**:
- **m6**, tal como la describe `tasks.md` *(«el incoherente se emite»)*, panicaría con el registro sin `usage`, porque el puntero es nil.
  Se declara así: «un registro sin `usage` se toma con partidas a cero, y ninguno se clasifica como incoherente». Emite los tres casos sin
  panicar.
- **M-B2b** es nueva: valida la hoja que nació verde.

| # | Mutación *(en `internal/ingest/codex.go`)* | Debe caer *(hojas)* |
|---|---|---|
| m4 | `TokensInput: int(u.Entrada - u.Cache)` *(sin restar la escritura)* | `TestLineaCodex_EscrituraDeCacheDentroDeLaEntrada` |
| m5 | `CostAvailable: true` | `TestLineaCodex_SinCoste` |
| m6 | la guarda de coherencia → `if r.Usage == nil { r.Usage = &usoCodex{} }` | `TestLineaCodex_Incoherente/cache_y_escritura_mayores_que_la_entrada`, `/sin_usage`, `/partida_negativa` |
| m7 | `session_ref` del `thread_id` *(campo `ThreadID` decodificado)* | `TestLineaCodex_SesionDelRegistro` |
| m8 | `OccurredAt: time.Now()` | `TestLineaCodex_MomentoDelRegistro` |
| M-B2a | `SessionRef: r.SessionID` *(sin sal)* | `TestLineaCodex_SesionDelRegistro`, `TestLineaCodex_NingunIdentificadorDelProveedorEnElEvento` |
| M-B2b | sin la guarda del tipo *(toda línea es registro)* | `TestLineaCodex_UnRegistroEsUnEvento/otra_linea_no_es_registro` |

### T015 · Resultado *(2026-10-07)*

**Las siete coinciden con lo declarado.** Ninguna panicó. Cada una se aplicó con una sustitución exacta *(una sola aparición)*. Desde m5,
se comprobó el resultado **antes** de revertir; m4 se revirtió en la misma orden, tras coincidir. La reversión fue por edición inversa, y el
md5 de `codex.go` volvió a **`1aa5b6e4480ac4db3bf9faefda7bcc5d`** tras cada una.

| # | Cayó | Mensaje |
|---|---|---|
| m4 | `…EscrituraDeCacheDentroDeLaEntrada` | `100 / 40 / 60 / 10 → [60 60 40 10]; se esperaba [0 60 40 10]` |
| m5 | `TestLineaCodex_SinCoste` | `tool "codex", cost_usd 0, cost_available true; se esperaba codex, 0 y false` |
| m6 | `TestLineaCodex_Incoherente/` ×3 | los tres salen como evento *(clase 1)*: con entrada −10, con todo a 0 y con salida −1 |
| m7 | `TestLineaCodex_SesionDelRegistro` | `session_ref = "26750c29…"; se esperaba event.Ref(sal, session_id) = "56ee076e…"` |
| m8 | `TestLineaCodex_MomentoDelRegistro` | `occurred_at = <ahora>; se esperaba 2026-10-07 00:00:01.25 +0000 UTC` |
| M-B2a | `…SesionDelRegistro`, `…NingunIdentificadorDelProveedorEnElEvento` | `session_ref = "s-sintetica-raiz"` · `el evento lleva un identificador del proveedor: {… "session_ref":"s-CENTINELA-SESION" …}` |
| M-B2b | `…UnRegistroEsUnEvento/otra_linea_no_es_registro` | `una línea turn_context: (<nil>, clase 2, <nil>); se esperaba (nil, NoEsRegistro, nil)` |
