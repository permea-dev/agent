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

### Remate de B2 · Enmienda E-4 *(Encargo 6, 2026-10-07; registrada en `spec.md`)*

**Rojos**: cuatro hojas nuevas en (6), y un test nuevo. Fixtures sintéticos nuevos en `testdata/codex/`.

```
--- FAIL: TestLineaCodex_Incoherente/sin_timestamp           (<nil>, clase 0, unexpected end of JSON input); se esperaba (nil, Incoherente, nil)
--- FAIL: TestLineaCodex_Incoherente/timestamp_mal_formado   (<nil>, clase 0, parsing time "ayer por la tarde" as "2006-01-02T15:04:05Z07:00": …); se esperaba (nil, Incoherente, nil)
--- FAIL: TestLineaCodex_Incoherente/partida_no_numerica     (<nil>, clase 0, json: cannot unmarshal string into Go struct field usoCodex.usage.input_tokens of type int64); …
--- FAIL: TestLineaCodex_Incoherente/payload_no_decodifica   (<nil>, clase 0, json: cannot unmarshal array into Go value of type ingest.registroCodex); …
--- PASS: TestLineaCodex_PartidaAusenteValeCero              (nace verde: el decodificado de Go ya da 0; lo valida M-B2d)
```

**Verde**:
- `registroCodex.Usage` pasa a `json.RawMessage`, y se decodifica aparte con `usoDe`.
- Orden: envoltura *(si falla, error = corrupta)* → tipo → payload *(si no decodifica, incoherente)* → `response_id` *(sin identificador)* →
  `timestamp` *(si falla, incoherente)* → `usage` *(ausente, `null` o que no decodifica, incoherente)* → coherencia.
- Todos los `TestLineaCodex_*` en PASS *(20 líneas PASS entre tests y hojas)*. `0 issues` y 9/9.

**Censo declarado ANTES de mutar**:

| # | Mutación *(en `internal/ingest/codex.go`)* | Debe caer *(hojas)* |
|---|---|---|
| M-B2c | sin fecha válida devuelve error *(`return nil, NoEsRegistro, err`)* | `TestLineaCodex_Incoherente/sin_timestamp`, `/timestamp_mal_formado` |
| M-B2d | una partida ausente es incoherente *(`usoDe` exige `cache_write_input_tokens` en el crudo)* | `TestLineaCodex_PartidaAusenteValeCero` |
| M-B2e | el payload que no decodifica devuelve error | `TestLineaCodex_Incoherente/payload_no_decodifica` |

**Resultado: las tres coinciden.** Reversión por edición inversa, y el md5 de `codex.go` es **`2e674096168e779136ae77647d10c4ca`** antes y
después de cada una.
- **M-B2c**: `…/sin_timestamp` *(`unexpected end of JSON input`)* y `…/timestamp_mal_formado` *(`parsing time "ayer por la tarde" …`)*.
- **M-B2d** *(dos sustituciones, revertidas en orden inverso)*: `…PartidaAusenteValeCero`, `clase 3, evento <nil>; se esperaba un evento`.
- **M-B2e**: `…/payload_no_decodifica`, `json: cannot unmarshal array into Go value of type ingest.registroCodex`.

## B3 · Contexto y estado

### T017 · Fase 0

`internal/ingest/codex_contexto.go` *(nuevo)*:
- `PasadaCodex`, con los ocho recuentos y `Resumen()` *(el texto aprobado)*;
- `LeerFicheroCodex(st, ruta, base, p, avisos io.Writer)`, que recorre con `Recorrer` y `fijar` = lo leído, sin emitir;
- `ContarComprimido(st, ruta, p)`, que no cuenta.

Suite verde. **De paso**: en la primera escritura de `Resumen()` el hueco de «repetidas» llevaba otra expresión; se corrigió antes de
seguir y antes de cualquier test.

### T018 · Fixtures sintéticos, en `internal/ingest/testdata/codex/contexto/`

21 ficheros, generados por programa, con turnos, modelos *(`modelo-a`, `-b`, `-z`)* y directorios *(`/tmp/…-sintetico`)* inventados:
- dos turnos; un ajuste de modelo dentro del turno; una compactación; sin modelo;
- formato anterior, mixto y sin consumo;
- una reanudación en dos mitades; un corte tras el `turn_context`;
- una bifurcación en dos ficheros; la cuenta de SC-017;
- el `cwd` y «sin `cwd`»;
- una línea corrupta en dos mitades;
- un fichero largo y su versión truncada;
- un `.zst` de bytes inventados.

### T019, T055 y (35) · Rojos, en `internal/ingest/codex_contexto_test.go` *(nuevo)*

**Una corrección del propio test antes de dar los rojos por buenos**: en la primera ejecución, `ReanudacionEnDosPasadas` panicó
*(`entera[:1]` sobre una referencia vacía)*, y `ContextoEntrePasadas` pasó comparando 0 con 0. `deUnaVez` exige ahora, como
precondición, los eventos que trae el fixture.

```
--- FAIL: TestContextoCodex_ModeloDelTurno/dos_turnos               modelos []; se esperaba [modelo-a modelo-b]
--- FAIL: TestContextoCodex_ModeloDelTurno/ajuste_dentro_del_turno  modelos []; se esperaba el del turno, [modelo-a], no el vigente (modelo-b)
--- FAIL: TestContextoCodex_CompactacionLlevaElVigente              modelos []; se esperaba [modelo-a modelo-b]: la compactación, el vigente
--- FAIL: TestContextoCodex_SinModelo                               0 eventos, modelos [], sin modelo 0; se esperaba 1 evento con modelo vacío y sin modelo = 1
--- FAIL: TestContextoCodex_ContextoEntrePasadas                    precondición: leyendo de una vez salen 0 eventos; el fixture trae 1
--- FAIL: TestContextoCodex_ReanudacionEnDosPasadas                 precondición: leyendo de una vez salen 0 eventos; el fixture trae 4
--- FAIL: TestContextoCodex_BifurcacionUnEvento                     0 eventos y 0 repetidas; se esperaba 1 y 1
--- FAIL: TestContextoCodex_Formatos/anterior · /mixto              0 eventos y 0 en formato anterior; se esperaba 0 y 1 · … 1 y 0
--- PASS: TestContextoCodex_Formatos/sin_consumo                    (nace verde; la valida M-B3b)
--- FAIL: TestContextoCodex_ComprimidoUnaVez/primera · /tras_cambiar   comprimidos = 0; se esperaba 1
--- PASS: TestContextoCodex_ComprimidoUnaVez/segunda                (nace verde; la valida m12)
--- FAIL: TestContextoCodex_CuentaDelResumen                        resumen: got "codex: respuestas 0 · eventos 0 · …"
--- FAIL: TestContextoCodex_Cwd                                     precondición: 0 eventos; se esperaban 2
--- FAIL: TestContextoCodex_LineaCorrupta/primera_pasada            0 avisos, 0 eventos, 0 respuestas; se esperaba 1, 1 y 1
--- FAIL: TestContextoCodex_LineaCorrupta/segunda_pasada            avisos "" y 0 eventos; … se esperaba ningún aviso y 1 evento
--- FAIL: TestContextoCodex_TruncadoNoLeePrefijo                    precondición: la primera pasada dio 0 eventos; se esperaba 1
```

**Razón**: la Fase 0 no emite, no cuenta y no reconstruye contexto.

**(35) es nuevo** *(Encargo 6)*: «un fichero truncado o rotado no lee prefijo». En el fichero nuevo, el `turn_context` va **después** de
su registro: leído como prefijo, le daría un modelo que todavía no tenía.

### T020 · Verde

- **D-008-P1**: un fichero sin bytes nuevos no se abre. El prefijo `[0, offset)` se lee con el filtro de cinco marcadores, sólo si
  `0 < offset < tamaño`; truncado o rotado, sin prefijo. El contexto se actualiza también con la parte nueva.
- **FR-011 y FR-014**: el turno, y si falta, el vigente o el `cwd` de `session_meta`. Sin ningún `cwd`, `project_ref` vacío *(E-4)*.
- **FR-017**: formato anterior = `token_count` con `info` y ningún registro, sobre el fichero entero *(prefijo y parte nueva)*.
- **D-008-P9**: la clasificación, con los `event_id` emitidos en la pasada *(repetidas)*.
- **D-008-P2**: el `.zst`, como entrada de `state.json` con cuatro campos y `Offset = Size`.
- **FR-029**:
  - el aviso va a un `io.Writer` que da quien llama, con el texto de `cmd/permea/main.go:282`;
  - sale sólo por la parte nueva;
  - en el prefijo, silencio.
- **Resultado**: los 27 tests y hojas de B3, en **PASS**; `gofmt` aplicado; `0 issues`; 9/9.

### T021 · Censo declarado ANTES de mutar *(2026-10-07)*

**Ajustes respecto a `tasks.md`, declarados aquí antes de mutar**:
- **m9** sólo tumba `ajuste_dentro_del_turno`: en `dos_turnos` el vigente coincide con el del turno.
- **m26** tumba **las dos** hojas de (34): el error no deja avanzar el estado, y la segunda pasada vuelve a encontrar la línea en la parte
  nueva.
- **Nuevas**:
  - **M-B3a**, el truncado lee prefijo, valida (35);
  - **M-B3b**, «sin registros es formato anterior», valida `Formatos/sin_consumo`, que nació verde.

| # | Mutación *(en `internal/ingest/codex_contexto.go`)* | Debe caer *(hojas)* |
|---|---|---|
| m9 | `delTurno` da siempre el vigente | `TestContextoCodex_ModeloDelTurno/ajuste_dentro_del_turno` |
| m10 | sin prefijo *(no se llama a `leerPrefijoCodex`)* | `TestContextoCodex_ContextoEntrePasadas`, `TestContextoCodex_ReanudacionEnDosPasadas/segunda` |
| m11 | sin repetidas *(siempre se emite)* | `TestContextoCodex_BifurcacionUnEvento`, `TestContextoCodex_CuentaDelResumen` |
| m12 | el `.zst` cuenta en cada pasada | `TestContextoCodex_ComprimidoUnaVez/segunda` |
| m13 | formato anterior = «hay `token_count`» | `TestContextoCodex_Formatos/mixto` |
| m14 | `cwd` siempre de `session_meta` | `TestContextoCodex_Cwd/del_turno` |
| m26 | la línea corrupta corta el fichero *(el callback devuelve el error)* | `TestContextoCodex_LineaCorrupta/primera_pasada`, `/segunda_pasada` |
| m27 | el aviso también en el prefijo | `TestContextoCodex_LineaCorrupta/segunda_pasada` |
| M-B3a | el truncado lee prefijo *(condición `offset > 0`)* | `TestContextoCodex_TruncadoNoLeePrefijo` |
| M-B3b | sin registros = formato anterior | `TestContextoCodex_Formatos/sin_consumo` |

### T022 · Resultado *(2026-10-07)*

**Las diez coinciden con lo declarado.** Ninguna panicó. Se aplicaron como sustituciones exactas, se miró el resultado antes de revertir, y
las de varias sustituciones *(m27, tres)* se revirtieron en orden inverso. El md5 de `codex_contexto.go` volvió a
**`ad738316cf40cbb6266b4eae85d25ecb`** tras cada una.

| # | Cayó | Mensaje |
|---|---|---|
| m9 | `…ModeloDelTurno/ajuste_dentro_del_turno` | `modelos [modelo-b]; se esperaba el del turno, [modelo-a]` |
| m10 | `…ContextoEntrePasadas`, `…ReanudacionEnDosPasadas/segunda` | modelo `""` y proyecto `""` frente a `modelo-a` y el proyecto de una vez · la compactación, con proyecto `""` *(perdió el `cwd` de `session_meta`)* |
| m11 | `…BifurcacionUnEvento`, `…CuentaDelResumen` | `2 eventos y 0 repetidas; se esperaba 1 y 1` · el resumen cambia |
| m12 | `…ComprimidoUnaVez/segunda` | `comprimidos = 1; se esperaba 0` |
| m13 | `…Formatos/mixto` | `1 eventos y 1 en formato anterior; se esperaba 1 y 0` |
| m14 | `…Cwd/del_turno` | `project_ref = "34a3acaf…"; se esperaba el del cwd del turno, "05725fcd…"` |
| m26 | `…LineaCorrupta/primera_pasada`, `/segunda_pasada` | `LeerFicheroCodex(s.jsonl): unexpected end of JSON input` ×2 |
| m27 | `…LineaCorrupta/segunda_pasada` | `avisos "skip (línea corrupta): unexpected end of JSON input\n" …; se esperaba ningún aviso` |
| M-B3a | `…TruncadoNoLeePrefijo` | `1 eventos, modelos [modelo-z], sin modelo 0; se esperaba 1 evento sin modelo` |
| M-B3b | `…Formatos/sin_consumo` | `0 eventos y 1 en formato anterior; se esperaba 0 y 0` |

### Medida informativa de SC-014 *(no es puerta de B3; 2026-10-07)*

**Cómo**:
- Un fichero sintético de **105 013 608 B**, con 20 745 líneas y 1 497 registros, generado en `mktemp -d /tmp/permea-008-XXXXXX`.
- Contenido: un `session_meta` con 20 KB de instrucciones inventadas, y por turno un `turn_context`, de 10 a 40 líneas de 400 B a 18 KB y
  de 1 a 3 registros.
- La medida, con un test **temporal** en `codex_contexto_test.go`: una lectura entera con `LeerFicheroCodex` y, tres veces, «añadir 1
  registro y leer».
- Al terminar, el test se quitó, y el md5 del fichero volvió a ser el de antes *(`a2f5abb860c0e29182796f8e1f3caf1e`, tras quitar una línea en
  blanco sobrante)*. El temporal se borró tras comprobar el prefijo.

| | Tiempo | Resultado |
|---|---:|---|
| Lectura entera | 306 ms | 1 497 eventos |
| Medida 1 | **250 ms** | 1 evento, `modelo-sintetico` *(el del turno, sacado del prefijo)* |
| Medida 2 | **239 ms** | 1 evento |
| Medida 3 | **260 ms** | 1 evento |

**Máquina**: 16 hilos, Intel Core i5-13400, ext4 en WSL, con la caché de páginas caliente.

**Es la función de biblioteca, no el `--run` entero**: falta el arranque del proceso y la lectura de Claude Code. La medida de SC-014 que
cuenta es la del quickstart, en C2. Con ~0,25 s queda margen frente a los 3 s.

### FR-029 · cómo sale el aviso

- **Por un `io.Writer`** que recibe `LeerFicheroCodex`, con el texto de `cmd/permea/main.go:282` *(`skip (línea corrupta): <error>`)*.
- **Por qué un `io.Writer` y no un retorno**: `internal/ingest` no escribe en stderr por su cuenta, el aviso sale en el orden en que se
  lee, y B5 sólo tiene que pasar `os.Stderr`. Un retorno obligaría a cada llamante a reconstruir el orden y el texto.
- **Un `nil`** calla el aviso.
- **En el prefijo** no se avisa nunca *(T055, m27)*.

### Comprobación de forma contra la copia congelada *(Encargo 7, Fase previa, 2026-10-07)*

**Por qué**: los fixtures de B3 son sintéticos, y el contador comparte con el agente la ruta del modelo de `thread_settings_applied`
*(`payload.thread_settings.model`)*. Se mide **con el código del agente**: `contextoFichero.aprender`, línea a línea con un contexto nuevo, y
`LeerFicheroCodex` en un estado vacío y sin encolar.

**Instrumento**: un test temporal, `internal/ingest/zz_forma_temporal_test.go`, que sólo imprime recuentos y modelos. Se **borró** al
terminar, y el árbol quedó limpio. La copia congelada se leyó en su sitio, con huella **`984d483c12a15601`** *(8 ficheros)* antes y después.

| F | `session_meta` *(con `cwd`)* | `turn_context` *(`turn_id` / modelo / `cwd`)* | `thread_settings_applied` *(con modelo)* | Orden registro / `token_count` |
|---|---|---|---|---|
| F1–F4 | 1 *(1)* cada uno | 54, 14, 47, 20 *(todos con los tres)* | 0 | sólo `token_count` |
| F5 | 1 *(1)* | 1 *(1 / 1 / 1)* | 0 | — |
| F6 | 1 *(1)* | 5 *(5 / 5 / 5)* | **8 *(8)***: `gpt-5.3-codex` ×1, `gpt-6-luna` ×7 | R C R C R C R C |
| F7 | 1 *(1)* | 1 *(1 / 1 / 1)* | 0 | R C |
| F8 | 1 *(1)* | 1 *(1 / 1 / 1)* | 0 | R C R C |

**El agente extrae el dato del 100 % de las líneas de contexto**. Ningún `thread_settings_applied` quedó sin modelo, así que no hubo que
listar campos.

**`LeerFicheroCodex`, frente al contador** *(descubrimiento §FASE 0 (f))*:

| F | Agente: eventos · entrada / escritura / lectura / salida · modelos | Contador | |
|---|---|---|---|
| F1–F4 | 0 · formato anterior = 1 cada uno | 0 · anterior | ✅ |
| F5 | 0 · ni anterior ni comprimido | 0 · sin consumo | ✅ |
| F6 | 4 · 15 076 / 0 / 51 200 / 88 · `gpt-6-luna` ×4 | 4 · 15 076 / 0 / 51 200 / 88 · `gpt-6-luna` ×4 | ✅ |
| F7 | 1 · 2 823 / 0 / 11 008 / 5 · `gpt-6-luna` | igual | ✅ |
| F8 | 2 · 3 880 / 0 / 24 064 / 237 · `gpt-6-luna` ×2 | igual | ✅ |

**El orden**: en F6, F7 y F8 el `token_usage_record` de cada respuesta va **antes** que su `token_count`. Una pasada que cayera entre los dos
vería ya el registro, así que no hay riesgo de contar el fichero como «formato anterior» por ese orden.

**Resultado: la forma de los fixtures es la de los ficheros reales. Se sigue con B4.**

## B4 · Raíz y activación

### T024 · Fase 0

`internal/config/codex.go` *(nuevo)*: `CodexSessionsRoot() (string, error)`, que devuelve `"", nil`. Suite verde.

### T025 · Rojos, en `internal/config/codex_test.go` *(nuevo)*

```
--- FAIL: TestCodexSessionsRoot_Raiz/definida              CodexSessionsRoot() = ("", <nil>); se esperaba ("<temporal>/sessions", nil)
--- FAIL: TestCodexSessionsRoot_Raiz/vacia                 con CODEX_HOME="": ("", <nil>); se esperaba ("<hogar>/.codex/sessions", nil)
--- FAIL: TestCodexSessionsRoot_Raiz/ausente               sin CODEX_HOME: ("", <nil>); se esperaba ("<hogar>/.codex/sessions", nil)
--- FAIL: TestCodexSessionsRoot_InexistenteSinError        con la raíz inexistente: ("", <nil>); se esperaba ("<hogar inexistente>/.codex/sessions", nil)
```

**Razón**: la Fase 0 no resuelve nada. Las rutas son temporales del test.

**Cómo se hace cada caso**:
- **«ausente»**: `t.Setenv` y después `os.Unsetenv`, de modo que `t.Setenv` la restaura al terminar.
- **En (21)**: se crea la carpeta `sessions`, para que sólo **(22)** pruebe la raíz inexistente *(E-2)*.

### T026 · Verde

- **`CodexSessionsRoot()`**: `CODEX_HOME` no vacía → `<valor>/sessions`; si no, `os.UserHomeDir()` + `.codex/sessions`. **Sin mirar si
  existe** *(E-2)*.
- **`internal/testutil/sandbox.go`** *(M-9)*: una línea, `t.Setenv("CODEX_HOME", "")`, con su comentario.
- **Resultado**: los 4 tests y hojas, en PASS; `0 issues`; 9/9.

### T027 · Comprobado, sin tocarlo

- **`sandbox_test.go`** sigue verde *(sus 4 tests)*, y `git diff 7b8c77c -- internal/testutil/sandbox_test.go` está vacío.
- **`os.Getenv` de producción**:
  ```
  $ grep -rn 'os.Getenv' --include=*.go cmd internal | grep -v _test
  internal/config/codex.go:19:	if propia := os.Getenv("CODEX_HOME"); propia != "" {
  ```
  **Una línea.** La primera vez salieron **dos**, porque el comentario nombraba `os.Getenv` literalmente; se reformuló *(«la ÚNICA lectura de
  una variable de entorno en producción»)*.

### T028 · Censo declarado ANTES de mutar *(2026-10-07)*

**Nueva**: **M-B4a**, «exigir que la raíz exista», valida E-2 en (22).

| # | Mutación *(en `internal/config/codex.go`)* | Debe caer *(hojas)* |
|---|---|---|
| m15 | `CODEX_HOME=""` tomada como raíz *(`os.LookupEnv` en vez de «no vacía»)* | `TestCodexSessionsRoot_Raiz/vacia` |
| m16 | ignorar `CODEX_HOME` *(`os.Getenv("CODEX_HOME")` → `""`)* | `TestCodexSessionsRoot_Raiz/definida` |
| M-B4a | exigir que la raíz exista *(un `os.Stat` en la rama del hogar)* | `TestCodexSessionsRoot_InexistenteSinError` |

**Resultado: las tres coinciden**, y ninguna tumbó nada fuera de lo declarado en la suite entera. Reversión por edición inversa, y el md5 de
`internal/config/codex.go` es **`fdc32b1c024af448d021800a85cd5752`** antes y después de cada una.
- **m15**: `…Raiz/vacia`: `con CODEX_HOME="": ("sessions", <nil>); se esperaba ("<hogar>/.codex/sessions", nil)`.
- **m16**: `…Raiz/definida`: devuelve `<hogar>/.codex/sessions` en vez de `<CODEX_HOME>/sessions`.
- **M-B4a**: `…InexistenteSinError`: `("", stat <hogar inexistente>/.codex/sessions: no such file or directory)`.

## B5 · Integración

### T030 · Fase 0 *(Encargo 8, 2026-10-07)*

**1 · La referencia de SC-011** *(D-008-P12)*, generada **antes** de tocar `main.go`, con el binario de `HEAD` *(`c1068d6`)*:
- **El entorno**: `env -i` con `HOME`, `USERPROFILE` y `XDG_CONFIG_HOME` temporales, `CODEX_HOME` vacía, y un `config.json` con sólo `logs_root`.
- **El fixture**: `cmd/permea/testdata/codex/claude.jsonl`, sintético, con dos mensajes de 2026-10-01. Tienen más de 24 h, así que la regla
  (iii) los cierra y la salida no depende del día.
- **La normalización**: la ruta de datos se sustituye por `<DATOS>`.
- **Resultado**: `cmd/permea/testdata/codex/referencia-run.stderr`, md5 **`17f189f03dde79100eaac3d5b32cc32a`**, `rc=0`:
```
Permea 0.0.1-dev
2 eventos encolados en <DATOS>/queue.jsonl
pasada: 2 líneas facturables · 2 eventos · 0 repetidas del mismo mensaje · 0 sintéticas · 0 sin identificador (no contables) · 0 con consumo distinto de la primera
pasada: 0 mensajes que crecieron entre líneas · 0 en espera de cerrarse · 0 líneas releídas de un mensaje en espera · 0 líneas tardías · 0 líneas sin desglose de caché (a 1 hora)
sync omitido: sin endpoint configurado
```

**2 · El forzado de SC-015** *(D-008-P10)*: **se elige el candidato A**. Se comprobó con un test temporal en `cmd/permea`, ya borrado: la
cola existente y vacía, y el directorio de datos en `0500`.

| Paso | Resultado |
|---|---|
| `state.Load` *(sin `state.json`)* | `err <nil>`: un `Store` vacío |
| `transport.Append` a la cola existente | `err <nil>`: 307 bytes escritos |
| `st.Save` | `open <temporal>/.state-…: permission denied` |

La máquina no es root *(uid 1000)*, así que el test **no** se salta aquí. Lleva `t.Skip` sólo con root o en Windows, donde los permisos POSIX
no aplican *(R-8)*. **Candidato B**, el campo inyectable, no hace falta.

**3 · El esqueleto**:
- en `main.go`, los campos `agent.codexRaiz` y `agent.codex`, y `setup()` resuelve la ruta con `config.CodexSessionsRoot()`. Sin directorio
  personal la deja vacía, sin error.
- en `codex_contexto.go`, `HayNovedades()` y `ListarCodex()`, todavía vacíos.

Suite verde.

### T031, T034, T056, T057 y (36) · Rojos

En `cmd/permea/codex_test.go` *(nuevo)* y en `internal/ingest/codex_contexto_test.go`. Los fixtures son sintéticos, en
`cmd/permea/testdata/codex/`: `sesion` con 2 registros, `otra` con 1, y `centinelas`.

```
--- PASS: TestCodexRun_SinRaizEsLaSalidaDeLa040/codex_home_vacia · /raiz_inexistente   (23, nacen verdes; los valida m18)
--- FAIL: TestCodexRun_SinClaudeCode                         (24) código 0 y 0 eventos codex; se esperaba 0 y 2
--- FAIL: TestCodexGenerate_EncolaAntesDeGuardar             (25) la pasada falló al guardar y la cola tiene 0 eventos codex; se esperaban 2, encolados ANTES de guardar
--- FAIL: TestCodexRun_LineaDeResumen                        (26) código 0; stderr: … (sin la línea codex:)
--- FAIL: TestCodexDemonio_SoloConNovedades/predicado        (27) respuestas: HayNovedades() = false; se esperaba true (y formato_anterior, comprimidos)
--- FAIL: TestCodexDemonio_SoloConNovedades/tick             (27) primer ciclo: … se esperaba la línea codex: sólo en el primero
--- FAIL: TestCodexRun_SegundaPasadaCero                     (28) precondición: la primera pasada encoló 0 eventos codex; se esperaban 2
--- FAIL: TestCodexRun_NadaDelProveedorViaja                 (31) precondición: 0 eventos codex en la cola; se esperaba 1
--- FAIL: TestCodexActivacion_EnCadaPasada                   (32) con la carpeta creada después: err <nil>, recuentos <nil>, cola codex 0; se esperaban 2 eventos
--- FAIL: TestCodexGenerate_FicheroIlegibleNoRompe/pasada    (33) err <nil>, state.json <nil>, claude 2, codex 0; se esperaba … el aviso … 2 y 1
--- PASS: TestCodexGenerate_FicheroIlegibleNoRompe/segunda_pasada   (33, nace verde; lo valida m24)
--- FAIL: TestCodexGenerate_FicheroIlegibleNoRompe/se_relee  (33) con el fichero ya legible: 0 eventos codex; se esperaban 3
--- FAIL: TestContextoCodex_ListarRaiz                       (36, nuevo) ListarCodex = ("", "", <nil>); se esperaba ([…a.jsonl …b.jsonl], […c.jsonl.zst], nil)
```

**Razón**: el esqueleto no lee Codex.
- **(32)** pasa por `setup()` de verdad, en el sandbox, para que m23 tenga dónde morder.
- **(25) y (33)** llevan `t.Skip` sólo con root o en Windows *(R-8)*; aquí no se saltan.
- **(36)** es nuevo: `ListarCodex`, el ayudante de enumeración que pide la integración, con su mutación **M-B5b**.

### T032 · Verde

- **`ListarCodex`**: `WalkDir` de la raíz; `.jsonl` y `.zst`, en orden. Un subdirectorio ilegible se salta.
- **`HayNovedades`**: respuestas, formato anterior o comprimidos.
- **`generate()`**: tras el bucle de Claude Code y antes del único `st.Save`, si `codexRaiz` existe **en esta pasada** llama a `generarCodex`
  *(nuevo)*.
- **`generarCodex`**:
  - `ListarCodex`;
  - por fichero, `LeerFicheroCodex` con el aviso a stderr, y un error de lectura → `codex: fichero omitido: %v` y siguiente fichero;
  - `transport.Append` de cada evento, que si falla es fatal;
  - `ContarComprimido` de cada `.zst`.
- **`runOnce`**: la línea `codex:` tras el resumen y el aviso de Claude Code, si `a.codex != nil`.
- **`tick`**: la misma línea, sólo si además `HayNovedades()`.
- **Resultado**: **568 pass, 0 SKIP** *(antes, 551 y 0)*, `0 issues`, 9/9.

### T033 · Comprobado, sin tocarlos

`main_test.go`, `retencion_test.go`, `coste_test.go` y `project_test.go`, en verde y con `git diff 7b8c77c` vacío. Los que construyen
`agent{…}` a mano no traen `codexRaiz`, y no leen Codex.

### T035 · Censo declarado ANTES de mutar *(2026-10-07)*

**Ajustes respecto a `tasks.md`, declarados aquí antes de mutar**:
- **m18** se declara sin pánico: «con el lector inactivo, se escribe la línea con una pasada vacía».
- **m20** tumba mucho más que (24): todos los tests de Codex **sin raíz de Claude Code** *(25, 27/tick, 28, 31 y 32)*.
- **m24** no tumba `se_relee`: el ilegible va antes que el sano en el orden léxico, y con permisos ya se leen los dos.
- **Nuevas**:
  - **M-B5a**: guardar el contexto del turno en `state.json`, la opción (b) de P-2 que se rechazó;
  - **M-B5b**: `ListarCodex` sin `.zst`;
  - **M-B5c**: `HayNovedades` sólo con respuestas.

| # | Mutación | Debe caer *(hojas)* |
|---|---|---|
| m17 | `st.Save` al principio de `generarCodex`, antes de encolar | `TestCodexGenerate_EncolaAntesDeGuardar` |
| m18 | con el lector inactivo, se escribe la línea con una pasada vacía | `TestCodexRun_SinRaizEsLaSalidaDeLa040/codex_home_vacia`, `/raiz_inexistente` |
| m19 | `tick` escribe la línea sin mirar `HayNovedades` | `TestCodexDemonio_SoloConNovedades/tick` |
| m20 | Codex sólo si hay logs de Claude Code | `TestCodexRun_SinClaudeCode`, `TestCodexGenerate_EncolaAntesDeGuardar`, `TestCodexDemonio_SoloConNovedades/tick`, `TestCodexRun_SegundaPasadaCero`, `TestCodexRun_NadaDelProveedorViaja`, `TestCodexActivacion_EnCadaPasada` |
| m23 | la existencia sólo en `setup()` *(`generate` no la mira)* | `TestCodexActivacion_EnCadaPasada` |
| m24 | el error de un fichero de Codex aborta la pasada | `TestCodexGenerate_FicheroIlegibleNoRompe/pasada`, `/segunda_pasada` |
| m25 | el fichero omitido guarda su offset al final | `TestCodexGenerate_FicheroIlegibleNoRompe/se_relee` |
| M-B5a | *(`codex_contexto.go`)* cada turno se guarda en `state.json` con su `turn_id` | `TestCodexRun_NadaDelProveedorViaja` |
| M-B5b | *(`codex_contexto.go`)* `ListarCodex` no enumera los `.zst` | `TestContextoCodex_ListarRaiz` |
| M-B5c | *(`codex_contexto.go`)* `HayNovedades` sólo con respuestas | `TestCodexDemonio_SoloConNovedades/predicado` |

### T036 · Resultado *(2026-10-07)*

**Las diez coinciden con lo declarado**, y ninguna panicó. Se aplicaron como sustituciones exactas, se miró el resultado antes de revertir,
y m23 *(dos sustituciones)* se revirtió en orden inverso. Los md5 volvieron tras cada una:
- `cmd/permea/main.go` → **`070d6938b52dbd691c20aa39d125a1b1`** *(m17–m25)*;
- `internal/ingest/codex_contexto.go` → **`3aeb390b2a7ff7dfd16f8b819aef0cb5`** *(M-B5a, M-B5b y M-B5c)*.

| # | Cayó |
|---|---|
| m17 | `TestCodexGenerate_EncolaAntesDeGuardar` |
| m18 | `TestCodexRun_SinRaizEsLaSalidaDeLa040/codex_home_vacia`, `/raiz_inexistente` |
| m19 | `TestCodexDemonio_SoloConNovedades/tick` |
| m20 | `…SinClaudeCode`, `…EncolaAntesDeGuardar`, `…SoloConNovedades/tick`, `…SegundaPasadaCero`, `…NadaDelProveedorViaja`, `…ActivacionEnCadaPasada` *(las seis declaradas)* |
| m23 | `TestCodexActivacion_EnCadaPasada` |
| m24 | `TestCodexGenerate_FicheroIlegibleNoRompe/pasada`, `/segunda_pasada` *(no `se_relee`, como se declaró)* |
| m25 | `TestCodexGenerate_FicheroIlegibleNoRompe/se_relee` |
| M-B5a | `TestCodexRun_NadaDelProveedorViaja` |
| M-B5b | `TestContextoCodex_ListarRaiz` |
| M-B5c | `TestCodexDemonio_SoloConNovedades/predicado` |

### Un límite conocido · R-9 *(se anota; no se arregla aquí)*

Si la lectura de un fichero de Codex falla **a mitad**, y no al abrirlo:
- `LeerFicheroCodex` ya ha sumado a la `PasadaCodex` respuestas cuyos eventos no se encolan;
- y ha marcado sus `event_id` como emitidos en la pasada.

El estado de ese fichero no avanza, así que se releen en la pasada siguiente. La línea `codex:` de esa pasada cuenta de más. Va a
`plan.md` §Riesgos como **R-9**.

## B6 y B7 · `--scan`, README y CHANGELOG *(un solo commit: decisión del orquestador, Encargo 9; B7 son sólo textos aprobados)*

### T038 · Antes de tocar `dryRun`: la referencia de (30)

El binario de `HEAD` *(`d350165`)* corrió `--scan` sobre `cmd/permea/testdata/codex/claude.jsonl`, en `env -i` con hogar temporal y sin
escribir nada en él. Salida: `referencia-scan-claude.stdout` *(md5 **`d7c4ec34cfdda7840f1c04ac85574376`**)* y `referencia-scan-claude.stderr`
*(**`a801c6a071fa180029e7fcc886f75355`**)*.

### T038 · Rojos, en `cmd/permea/codex_test.go`

Los formatos aprobados se leen de `spec.md` §Textos aprobados **por programa** *(`textoAprobado`, `patronDe`)*: el test no los teclea.

```
--- FAIL: TestCodexScan_EventosYResumen/lineas_evento    (29) 1 líneas en stdout; se esperaban 2 (stdout vacío: hoy la sesión pasa por el camino de Claude Code)
--- FAIL: TestCodexScan_EventosYResumen/resumen          (29) stderr no lleva la línea aprobada "codex: respuestas 2 · eventos 2 · …"
--- PASS: TestCodexScan_EventosYResumen/nada_en_disco    (29) nace verde; la valida M-B6b
--- PASS: TestCodexScan_ClaudeCodeComoLa040              (30) nace verde; la valida M-B6a
--- FAIL: TestCodexScan_FormatoAnterior                  (37, nuevo) código 0; stdout ""; stderr sin «ficheros en formato anterior 1»
```

**(37)** es nuevo *(Encargo 9)*: `--scan` de una sesión en formato anterior, con el fixture `anterior.jsonl`.

### T039 · Verde

- **`dryRun`**: llama antes a `esSesionCodex`. Esta lee la primera línea **con el mismo `Scanner` de 1 MiB** y el mismo error si se pasa; si es
  `session_meta`, sigue por `dryRunCodex` *(nuevas las dos)*.
- **`dryRunCodex`**: `LeerFicheroCodex` sobre `state.New()`, un estado sólo en memoria que **nunca se guarda**. Imprime la línea `evento:`
  aprobada, «N eventos generados (dry-run, nada transmitido)» y la línea `codex:`.
- **`main.go`** importa `encoding/json`.
- **Resultado**: los cinco tests y hojas, en PASS; `0 issues`; suite verde.

**La salida literal sobre `cmd/permea/testdata/codex/sesion.jsonl`** *(binario de la rama, hogar temporal)*:
```
Permea 0.0.1-dev
evento: tool=codex model=modelo-sintetico in=60 out=10 cw=0 cr=40 cost=$0.0000 cost_avail=false project_ref=b6b7efa8… event_id=cb0457c5f1897e1408679f6a3194375e
evento: tool=codex model=modelo-sintetico in=60 out=10 cw=0 cr=40 cost=$0.0000 cost_avail=false project_ref=b6b7efa8… event_id=7d0a8c7b06895d3ffbe79b5061be5021
2 eventos generados (dry-run, nada transmitido)
codex: respuestas 2 · eventos 2 · repetidas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0 · ficheros en formato anterior 0 · ficheros comprimidos 0
```

**Líneas de más de 1 MiB** *(medido, sin cambiar nada)*:
- **Claude Code**: `error: bufio.Scanner: token too long`, como hoy.
- **Codex**:
  - si la de más de 1 MiB es la **primera** línea, el mismo error, porque `esSesionCodex` usa el mismo `Scanner`;
  - si está **a mitad**, se lee sin error, porque `LeerFicheroCodex` no tiene tope, como en `--run`.

### T040 · Censo declarado ANTES de mutar *(2026-10-07)*

**Ajustes respecto a `tasks.md`, declarados aquí antes de mutar**:
- **m22** tumba también `resumen` y (37).
- **M-B6a** tumba, además de (30), **los cuatro tests de `--scan` de Claude Code** de 006 y 007. `TestScan_LineaConCuatroPartidasYEventID` cae
  en su precondición, antes de sus dos hojas.
- **Nueva**: **M-B6b**, que valida `nada_en_disco`, nacida verde.

| # | Mutación *(en `cmd/permea/main.go`)* | Debe caer *(hojas)* |
|---|---|---|
| m21 | la línea de Codex lleva `cw5m=0 cw1h=0` | `TestCodexScan_EventosYResumen/lineas_evento` |
| m22 | sin detección *(`esSesionCodex` devuelve siempre `false`)* | `TestCodexScan_EventosYResumen/lineas_evento`, `/resumen`, `TestCodexScan_FormatoAnterior` |
| M-B6a | todo fichero es Codex *(devuelve siempre `true`)* | `TestCodexScan_ClaudeCodeComoLa040`, `TestScan_UnEventoPorMensaje`, `TestScan_LineaConCuatroPartidasYEventID` *(precondición)*, `TestScan_LineaConDesgloseDeLaEscritura`, `TestScan_LaSalidaQueCreceValeSuMaximo` |
| M-B6b | `dryRunCodex` guarda su estado en el directorio de datos | `TestCodexScan_EventosYResumen/nada_en_disco` |

`TestRetirada_LasExcepcionesDeD0045` también lanza `--scan`, pero sólo mira el código de salida. **No debe caer.**

**Resultado: las cuatro coinciden**, y ninguna panicó. Se miró el resultado antes de revertir. El md5 de `cmd/permea/main.go` volvió a
**`265363f236397fa45d31d4c727d1748e`** tras cada una. `TestRetirada_LasExcepcionesDeD0045` siguió verde con M-B6a.
- **m21**: `…/lineas_evento`.
- **m22**: `…/lineas_evento`, `…/resumen`, `TestCodexScan_FormatoAnterior`.
- **M-B6a**: `TestCodexScan_ClaudeCodeComoLa040`, `TestScan_LineaConDesgloseDeLaEscritura`, `TestScan_LaSalidaQueCreceValeSuMaximo`,
  `TestScan_UnEventoPorMensaje`, `TestScan_LineaConCuatroPartidasYEventID` *(en su precondición, sin hojas)*.
- **M-B6b**: `…/nada_en_disco`.

### T042–T044 · README y CHANGELOG

**Rojo**, por `grep`: `^## 0.5.0` en `CHANGELOG.md` → 0; `^### Codex CLI` en `README.md` → 0.

**Cómo se insertaron**: los dos bloques **se sacaron por programa** de `spec.md` §Textos aprobados *(«**README**» y «**CHANGELOG `0.5.0`**»)*:
- `### Codex CLI` va al final de «Modos de ejecución», antes de «Coste y tarifas». El README no tiene una sección propia de Claude Code, y
  ésa es la que lo describe;
- `## 0.5.0 — PENDIENTE` va encima de la 0.4.0.

```
$ cmp <«### Codex CLI» de README.md> <bloque README de la spec>          → sin diferencias (7 líneas)
$ cmp <«## 0.5.0 — PENDIENTE» de CHANGELOG.md> <bloque de la spec>       → sin diferencias (19 líneas)
$ grep -c PENDIENTE CHANGELOG.md                                          → 1
```

## Cierre

### Enmienda E-5 · coherencia del README *(aprobada por el dueño el 2026-10-07; Encargo 10, fase previa)*

**Qué se hizo**:
- Ocho frases del README que hablaban sólo de Claude Code, cambiadas **literalmente**, al ancho de línea del fichero. Algunos párrafos se
  recolocaron, y la frase 1 queda partida en dos líneas.
- Registrada en `spec.md`, con la fila E-5 y las frases finales en §Textos aprobados.
- Corregidos en `plan.md` D-008-P8 y R-2: en `--scan` de Codex sólo la primera línea pasa por el tope de 1 MiB *(B6)*.

**Comprobación**: `grep -F` de cada frase final sobre el README con los saltos de línea *(y el `> ` de la cita)* normalizados a un espacio.
**9 de 9 presentes**: las ocho frases, más la viñeta nueva de la 7.
- Ninguna línea nueva pasa de 104 caracteres fuera de los bloques de código. Las dos que pasan, las líneas 49 y 180, ya estaban en
  `HEAD`.
- `cmp` de «### Codex CLI» y del CHANGELOG con la spec: **sin diferencias**.

### C1 · T046 · Puertas *(2026-10-07, sobre `9516a08` + los documentos de E-5; Go 1.22.2, golangci-lint 2.12.2)*

```
$ gofmt -l .                                   (vacío, rc=0)
$ go vet ./...                                 (rc=0)
$ golangci-lint run                            0 issues. (rc=0)
$ go test -count=1 ./...                       9/9 ok (rc=0) · recuento -json: tests {'pass': 574} · paquetes {'pass': 9} · SKIP 0
$ git diff 7b8c77c -- <frontera de FR-022> | wc -c                          0
$ git diff 7b8c77c --name-only --diff-filter=M -- '*_test.go'               (vacío)
$ git diff 7b8c77c --stat -- internal/state/ internal/event/ internal/pricing/   (vacío)
$ grep -rn 'os.Getenv' --include=*.go cmd internal | grep -v _test          internal/config/codex.go:19 (1)
$ CGO_ENABLED=0 GOOS=windows GOARCH=amd64 / GOOS=darwin GOARCH=amd64, arm64 go build   rc=0 ×3
$ grep -c PENDIENTE CHANGELOG.md               1
$ grep -rn nolint --include=*.go . | wc -l     1   (la de SA1007, de 006)
$ git log --oneline 7b8c77c..HEAD              7 commits: B0, B1, B2, B3, B4, B5, B6 y B7
```

**Resultado: todo en verde.**

### C2 · T047 · Medidas sobre la copia congelada *(2026-10-07; binario de la rama, `9516a08`)*

**Cómo**:
- **El sandbox**: `mktemp -d /tmp/permea-008-XXXXXX` y `env -i` con `HOME`, `XDG_CONFIG_HOME`, `PATH=/usr/bin:/bin` y
  `CODEX_HOME=<copia congelada>`. Sin enrolar *(`status` → «no enrolado» antes de cualquier `--run`)* y sin endpoint.
- **El binario**: compilado con `-X main.version=0.5.0-c2-9516a08`, para seguir `agent_version` hasta la cola.
- **El contador**: extraído por programa de `descubrimiento.md` §FASE 0 (f).
- **Huella de la copia**: **`984d483c12a15601`** *(8 ficheros)* antes y después.
- **El temporal**: borrado tras comprobar el prefijo.

**`--scan` fichero a fichero, frente al contador** *(SC-001, SC-002, SC-003, SC-009)*:

| F | Agente: eventos · entrada / escritura / lectura / salida · modelo | Línea `codex:` | Contador | |
|---|---|---|---|---|
| F1–F4 | 0 | formato anterior 1, cada uno | anterior | ✅ |
| F5 | 0 | todo a 0 | sin consumo | ✅ |
| F6 | 4 · 15 076 / 0 / 51 200 / 88 · `gpt-6-luna` ×4 | respuestas 4 · eventos 4 | igual | ✅ |
| F7 | 1 · 2 823 / 0 / 11 008 / 5 → **2828** | respuestas 1 · eventos 1 | igual | ✅ |
| F8 | 2 · 3 880 / 0 / 24 064 / 237 → **4117** | respuestas 2 · eventos 2 | igual | ✅ |

Todos con `rc=0` y `cost_avail=false`. `--scan` no dejó ningún fichero en el hogar del sandbox.

**`--run` dos veces** *(SC-002, SC-004, SC-005)*:
- **Primera**: «7 eventos encolados» · los dos resúmenes de Claude Code a 0 *(el sandbox no tiene sus logs)* ·
  `codex: respuestas 7 · eventos 7 · repetidas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0 · ficheros en formato anterior 4 · ficheros comprimidos 0`
  · «sync omitido».
- **Segunda**: «0 eventos encolados» y `codex:` todo a 0, con «ficheros en formato anterior 0».

**La cola**:
- 7 eventos, **todos `tool = codex`** *(0 de Claude Code)*, con 21 779 / 0 / 86 272 / 330;
- `cost_available = false` y `cost_usd = 0` en los 7;
- `agent_version = 0.5.0-c2-9516a08` en los 7, con modelo `gpt-6-luna` ×7;
- 7 `event_id` distintos, `schema_version = 1` y **17 campos** por evento;
- `state.json`, con 8 entradas.

**SC-014, con el binario entero**:
- **El fichero**: una sesión sintética de **115 170 686 B** *(110 003 líneas)*, generada con el guion de `quickstart.md` §Coste, extraído por
  programa, en otro `CODEX_HOME` temporal.
- **Los tiempos**, de pared *(`time`)*:

| | Tiempo | `codex:` |
|---|---:|---|
| Pasada inicial *(lee todo)* | 0,331 s | respuestas 1 · eventos 1 |
| Medida 1 | **0,340 s** | respuestas 1 · eventos 1 |
| Medida 2 | **0,340 s** | respuestas 1 · eventos 1 |
| Medida 3 | **0,342 s** | respuestas 1 · eventos 1 |

- **Lo que salió**: los tres eventos nuevos llevan `modelo-sintetico`, sacado del prefijo, y `project_ref` no vacío.
- **La máquina**: Intel Core i5-13400, 16 hilos, ext4 en `/tmp`, Linux 6.18 WSL2.
- **Tope: 3 s. Se cumple.**

**Resultado: SC-001 a SC-005, SC-009 y SC-014, ✅.**

### C3 · T048 · Snapshot *(2026-10-07; goreleaser v2.16.0)*

**Lo que se construyó**:
- `goreleaser release --snapshot --clean` → `rc=0`, «skipping announce, publish, and validate». **No se publicó nada**: no había ninguna
  credencial en el entorno.
- **El código es el de `9516a08`**. El árbol sólo tenía modificados el README *(E-5)* y tres documentos de `specs/`.
- **Versión inyectada: `0.4.0-SNAPSHOT-9516a08`.** goreleaser parte de la última etiqueta, `v0.4.0`; no es ni `0.0.1-dev` ni `0.4.0`.
- `sha256sum -c` del fichero de checksums: OK en los 5 artefactos.
- **El zip de Windows**: `dist/permea_0.4.0-SNAPSHOT-9516a08_windows_amd64.zip`, 2 495 694 B, SHA-256
  **`96dfa0ee5f49312f908827ac73e0efcf3f4a4606cc51b7d5c09653af8fbe5566`**.

**Las comprobaciones**:
- **El binario de Linux, en sandbox**: `--version` → `0.4.0-SNAPSHOT-9516a08`, sin crear ningún fichero en el hogar.
- **Los textos aprobados**, extraídos de la spec por programa y contados sobre los bytes del ejecutable de Linux y del de Windows: la línea
  `codex:`, `codex: fichero omitido: %v` y la línea `evento:` de Codex aparecen **una** vez cada una en los dos.
- **El README empaquetado**, en el zip de Windows y en el tar de Linux: lleva «### Codex CLI», las 9 frases de E-5, y es **idéntico** al del
  repo.
- **`dist/`** no aparece en `git status`: lo ignora `.gitignore`.

### C4 · T049 · W1 *(hecho por el dueño el 2026-10-07, de 18:37 a 18:42, Madrid, en Windows; en sandbox y sin enrolar)*

**Hechos del dueño**:
- **El zip**: `permea_0.4.0-SNAPSHOT-9516a08_windows_amd64.zip`, en una carpeta aparte, con `APPDATA` propio. SHA-256
  `96dfa0ee5f49312f908827ac73e0efcf3f4a4606cc51b7d5c09653af8fbe5566`, igual al de C3.
- **Antes de pasar**: `permea.exe --version` → `0.4.0-SNAPSHOT-9516a08` · `permea.exe status` → `no enrolado` · `CODEX_HOME` vacía · la
  carpeta de sesiones por defecto de Codex existe *(Q-6)*.
- **El contador independiente**, lanzado desde WSL sobre la carpeta viva de Codex justo antes: 8 ficheros · 7 registros *(F6 4, F7 1, F8 2,
  todos `gpt-6-luna`)* · F1–F4 formato anterior · F5 sin consumo.

**Primera `--run`** *(2,2 s)*, stderr literal:
```
Permea 0.4.0-SNAPSHOT-9516a08
6622 eventos encolados en <DATOS>\permea\queue.jsonl
pasada: 15144 líneas facturables · 6615 eventos · 8527 repetidas del mismo mensaje · 1 sintéticas · 0 sin identificador (no contables) · 158 con consumo distinto de la primera
pasada: 143 mensajes que crecieron entre líneas · 1 en espera de cerrarse · 0 líneas releídas de un mensaje en espera · 0 líneas tardías · 0 líneas sin desglose de caché (a 1 hora)
1 mensajes siguen abiertos: se enviarán en la próxima pasada
codex: respuestas 7 · eventos 7 · repetidas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0 · ficheros en formato anterior 4 · ficheros comprimidos 0
sync omitido: sin endpoint configurado
```

**Segunda `--run`** *(0,1 s)*, stderr literal:
```
Permea 0.4.0-SNAPSHOT-9516a08
0 eventos encolados en <DATOS>\permea\queue.jsonl
pasada: 3 líneas facturables · 0 eventos · 2 repetidas del mismo mensaje · 0 sintéticas · 0 sin identificador (no contables) · 0 con consumo distinto de la primera
pasada: 0 mensajes que crecieron entre líneas · 1 en espera de cerrarse · 3 líneas releídas de un mensaje en espera · 0 líneas tardías · 0 líneas sin desglose de caché (a 1 hora)
1 mensajes siguen abiertos: se enviarán en la próxima pasada
codex: respuestas 0 · eventos 0 · repetidas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0 · ficheros en formato anterior 0 · ficheros comprimidos 0
sync omitido: sin endpoint configurado
```

**Al terminar**: la cola del ensayo tenía 7 eventos con `"tool":"codex"`, y ninguno con `"cost_available":true`. La carpeta del ensayo se
borró, y no se transmitió nada.

**Comprobado aquí** *(Encargo 11, sin abrir nada del dueño)*:
- **Aritmética**:
  - 6622 = 6615 de Claude Code + 7 de Codex;
  - en Claude Code, eventos + repetidas + sintéticas + en espera = facturables, tanto en la primera pasada *(6615 + 8527 + 1 + 1 = 15144)*
    como en la segunda *(0 + 2 + 0 + 1 = 3)*;
  - en Codex, `respuestas = eventos + repetidas + sin identificador + incoherentes` *(FR-027)* en las dos: 7 = 7 y 0 = 0.
- **Frente al contador**: 7 respuestas = 7 registros, y 4 ficheros en formato anterior = F1–F4.
- **Frente a C2**: las mismas dos líneas `codex:` que el `--run` doble sobre la copia congelada *(7 · 4 formato anterior, y luego todo a 0)*.
- **Los textos**: las ocho líneas de recuentos son las aprobadas. Se reconocieron con los formatos sacados por programa de §Textos
  aprobados de 008 *(`codex:`)* y de 007 *(las dos `pasada:` y el aviso)*.
- **El orden** *(FR-019)*: la línea `codex:` va tras el resumen y el aviso de Claude Code, y antes de «sync omitido».
- **Q-6**: el agente encontró las sesiones con `CODEX_HOME` vacía, así que su raíz por defecto es la de Codex.
- **Lo de Claude Code** *(15144 · 6615 · 143 que crecen)* es el historial real de esa instalación, leído en sólo lectura, como en el W1 de 007.

**Resultado: W1 sin fallos. Nada contradice C2.**
