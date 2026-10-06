# 007 · Registro de rojos, verdes y mutaciones

**Por qué existe**: el techo de `tasks.md` es de 320 líneas, y no se sube *(decisión del orquestador, 2026-10-06)*. En `tasks.md` quedan
las casillas y una remisión. Aquí van las transcripciones, desde B1, **movidas sin cambiar un carácter**. Desde B2 se escriben aquí directamente.

## B1 · Tarifas

### T005

  > **Rojo transcrito**: `la tabla tiene 16 claves; el catálogo replicado tiene 17`. En `CifrasClaveAClave` fallan **17** subtests:
  > 16 con `escritura de caché a 1 hora = 0, want …` *(20, 10, 30, 8, 6, 4, 2, 1.6 según la fila)* y `falta la clave "claude-fable-5-1"`.
  > `NingunaClaveSobra` sigue verde.

### T006

  > **Verde**: 9/9 paquetes ok, **433 pass** *(432 + el subtest `claude-fable-5-1`)*. La «Limitación 1» añade que `CacheWrite1h` se
  > replica pero `Cost` todavía no la usa.

### T007

  > **Hecho el 2026-10-06.** Lo caído coincide con lo declarado en las tres, y el md5 de `pricing.go` es el mismo antes y después
  > *(`e3a744c9…`)*.
  > - **m1**: `CifrasClaveAClave/claude-fable-5-1`, «escritura de caché a 1 hora = 0, want 20».
  > - **m2**: `RecuentoDeClaves` *(«18 claves; … 17»)* y `NingunaClaveSobra` *(«"claude-sobrante-de-mutacion" está en la tabla…»)*.
  > - **m3**: `CifrasClaveAClave/claude-opus-5-5` *(«= 8, want 5» y «= 5, want 8»)* y `TestCost_Opus55AMano` *(«coste = 1.2146106000, want
  >   1.0775736000»)*.

## B2 · Desglose y coste por duración

### T009 · Fase 0

- `Cost(model, in, out, cacheWrite5m, cacheWrite1h, cacheRead)` con la conducta de hoy: las dos a `CacheWrite`. `FromClaudeCodeLine`
  pasa el total y 0; `TestCost`, `TestCost_UnknownModel` y `TestCost_Opus55AMano` cambian sólo la llamada. **433 pass, 9/9**: el mismo
  número que en B1.
- Esqueleto para que el rojo (7) falle como test y no como compilación: `Recuentos.SinDesglose`, sin usar. La suite, verde.

### T010 · Rojo (3)

`TestCost_Opus55AMano` pasa a dos subtests, los vectores de SC-004. Las cifras coinciden con las de la spec y con las del orquestador.
```
--- FAIL: TestCost_Opus55AMano/con_desglose
    coste = 1.0775736000, want 1.1775756000 (0,493828 + 0,157820 + 0,061725 + 0,266672 + 0,1975306 = 1,1775756; diferencia absoluta 0.1 > 1e-9)
--- FAIL: TestCost_Opus55AMano/sin_desglose_todo_a_1_hora
    coste = 1.0775736000, want 1.2146106000 (0,493828 + 0,157820 + 0,365432 + 0,1975306 = 1,2146106; diferencia absoluta 0.137 > 1e-9)
```
**Razón**: `Cost` aún tarifa las dos duraciones a `CacheWrite` (5,00).

### T011 · Rojos (4) a (7), en `internal/ingest/desglose_test.go` *(nuevo)*

```
--- FAIL: TestDesglose_LineaConDesgloseSeTarifaPorDuracion/coste     coste = 1.0775736000, want 1.1775756000
--- FAIL: TestDesglose_SinDesgloseTodoA1Hora/coste                   coste = 1.0775736000, want 1.2146106000
--- FAIL: TestDesglose_DesgloseQueNoSumaEsSinDesglose/coste          coste = 1.0775736000, want 1.2146106000
--- FAIL: TestDesglose_LaPasadaCuentaLasLineasSinDesglose            SinDesglose = 0; se esperaba 2
```
**Razón**: `rawRecord` no decodifica `cache_creation`, y nadie cuenta las líneas sin desglose. Los tres subtests
`tokens_cache_creation_*` **nacen verdes**, porque el evento ya lleva el total. Los validan m7, M-B2a y M-B2b.

### T012 · Rojo (8), en `cmd/permea/coste_test.go` *(nuevo, de proceso)*

```
--- FAIL: TestScan_LineaConDesgloseDeLaEscritura
    la línea `evento:` no lleva "cw5m=12345": "evento: tool=claude_code model=claude-opus-5-5 in=123457 out=7891 cw=45679 cr=987653 cost=$1.0776 …"
    la línea `evento:` no lleva "cw1h=33334": …
```
**Razón**: `dryRun` imprime cuatro partidas. `cw=45679` ya estaba.

### T013 · Verde

- `internal/pricing/pricing.go`: `Cost` tarifa `cacheWrite5m` a `CacheWrite` y `cacheWrite1h` a `CacheWrite1h`. En la cabecera, la
  **hipótesis P-1** sustituye a la «Limitación 1».
- `internal/ingest/claudecode.go`:
  - `rawRecord` decodifica `usage.cache_creation.ephemeral_5m_input_tokens` y `…_1h_…`, como punteros;
  - la guarda de frontera se amplía **sólo** a esos dos números;
  - `desglosarEscritura` aplica P-1 y Q-4;
  - la línea sin desglose, o con uno que no suma, se cuenta;
  - el evento sigue llevando el total.
- `internal/ingest/pasada.go`: `consumo` gana el desglose; `contarSinDesglose`; `Desglose(eventID)` para quien imprime.
- `cmd/permea/main.go` (`dryRun`): `cw5m=` y `cw1h=` detrás de `cw=`, tomados de la pasada.
- **9/9 ok, 446 pass**: 433 + 13 *(2 subtests de SC-004; 4 tests y 6 subtests en `desglose_test.go`; 1 en `coste_test.go`)*.

### T014 · Lo que no se toca

- `git diff` de `boundary_test.go`, `eventid_test.go`, `internal/event/`, `main_test.go` y `pasada_test.go`: **0 bytes**.
- `TestScan_LineaConCuatroPartidasYEventID` *(y sus dos subtests)* **PASS**, porque `" cw=7 "` sigue en la línea.
- `TestBoundary_NoDenylistLeaks`, `_TresCaminosHaciaElExterior`, `_UnknownFutureFieldDoesNotLeak`, `_CostAvailable` y `_KeepsMetrics`:
  **PASS**.

### T015 · Censo declarado ANTES de mutar *(2026-10-06)*

Cada mutación lista lo que **debe** caer, padres incluidos implícitamente. Si cae otra cosa o falta algo, se para. Respecto a
`plan.md` y `tasks.md`, **m4** y **m6** declaran más de lo que preveían, y **M-B2a** y **M-B2b** son nuevas: los subtests
`tokens_cache_creation_*` de (4), (5) y (6) nacen verdes.

| # | Mutación | Debe caer |
|---|---|---|
| m4 | `Cost`: `cacheWrite5m` a `CacheWrite1h` y `cacheWrite1h` a `CacheWrite` | `TestCost` *(36,75 → 40,5, fuera de ±1 %)*; `TestCost_Opus55AMano/con_desglose` y `/sin_desglose_todo_a_1_hora`; `TestDesglose_LineaConDesgloseSeTarifaPorDuracion/coste`; `TestDesglose_SinDesgloseTodoA1Hora/coste`; `TestDesglose_DesgloseQueNoSumaEsSinDesglose/coste` |
| m5 | `desglosarEscritura`: sin desglose → `(total, 0)` | `TestDesglose_SinDesgloseTodoA1Hora/coste`; `TestDesglose_DesgloseQueNoSumaEsSinDesglose/coste` |
| m6 | `desglosarEscritura`: sin la comprobación de la suma | `TestDesglose_DesgloseQueNoSumaEsSinDesglose/coste`; `TestDesglose_LaPasadaCuentaLasLineasSinDesglose` *(SinDesglose = 1)* |
| m7 | el evento lleva 5 m + 1 h **del log** cuando trae las dos claves | `TestDesglose_DesgloseQueNoSumaEsSinDesglose/tokens_cache_creation_es_el_total_del_log` |
| m8 | `dryRun` imprime `cw5m=` con el total | `TestScan_LineaConDesgloseDeLaEscritura` |
| M-B2a | el evento lleva `cw1h` como `tokens_cache_creation` | `TestDesglose_LineaConDesgloseSeTarifaPorDuracion/tokens_cache_creation_es_el_total`; `TestScan_LineaConDesgloseDeLaEscritura` *(`cw=33334`)* |
| M-B2b | sin desglose, el evento lleva `tokens_cache_creation` = 0 | `TestDesglose_SinDesgloseTodoA1Hora/tokens_cache_creation_es_el_total`; `TestDesglose_DesgloseQueNoSumaEsSinDesglose/tokens_cache_creation_es_el_total_del_log`; `TestScan_LineaConCuatroPartidasYEventID/cuatro_partidas` *(`cw=0`)* |

### T015 · Resultado *(2026-10-06)*

**Las siete coinciden con lo declarado.** Reversión por edición inversa, con el md5 igual antes y después:
- `pricing.go` `868828c2…`;
- `claudecode.go` `a8ca004e…`;
- `main.go` `6f8147f4…`.

| # | Caído *(hojas)* | Mensaje |
|---|---|---|
| m4 | los 6 declarados | `got 40.5 want 36.75`; `coste = 1.1146086000, want 1.1775756000` *(×2)*; `1.0775736000, want 1.2146106000` *(×3)* |
| m5 | los 2 declarados | `coste = 1.0775736000, want 1.2146106000` *(×2)* |
| m6 | los 2 declarados | `coste = 0.9791786000, want 1.2146106000`; `SinDesglose = 1; se esperaba 2` |
| m7 | el declarado | `tokens_cache_creation = 20000, want 45679` |
| m8 | el declarado | `no lleva "cw5m=12345": "… cw=45679 cw5m=45679 cw1h=33334 …"` |
| M-B2a | los 2 declarados | `tokens_cache_creation = 33334, want 45679`; `no lleva "cw=45679": "… cw=33334 …"` |
| M-B2b | los 3 declarados | `tokens_cache_creation = 0, want 45679` *(×2)*; `la línea no lleva "cw=7": "… cw=0 cw5m=0 cw1h=7 …"` |

**m8, primera forma, inválida**: imprimir `ev.TokensCacheCreation` en el sitio de `cw5m` dejaba la variable sin usar, y `cmd/permea`
no compilaba *(`[build failed]`)*. Se revirtió por edición inversa, con el md5 de `main.go` otra vez en `6f8147f4…`, y **no cuenta**.
La forma válida descarta `cw5m` de la pasada y lo toma del total.

## B3 · El máximo dentro de la pasada

### T017 · Fase 0

En `pasada.go`: el tipo `Cerrado` *(evento y desglose)*, `CerrarFichero()` devolviendo `nil` y seguro con pasada nula, y el recuento
`Crecieron`, sin usar. **Suite verde, 9/9.**

### T018 · Rojos

```
(9)  --- FAIL: TestMaximo_LaSalidaQueCreceValeSuMaximo        salida = 7; se esperaba 89 817, el máximo (la primera daría 7)
(10) --- FAIL: TestMaximo_ElMaximoEnMedioNoLaUltima           salida = 7; se esperaba 89 817, el máximo (la última daría 1 303)
(11) --- FAIL: TestMaximo_ElDesgloseEsElDeLaLineaDelMaximo/coste     coste = 0.0005000000, want 0.0024000000
     --- FAIL: …/desglose                                     la pasada no cerró el mensaje: el desglose no se puede observar
(25) --- FAIL: TestMaximo_EnEmpateValeLaUltima/coste          coste = 0.0015000000, want 0.0024000000
     --- FAIL: …/desglose                                     la pasada no cerró el mensaje: el desglose no se puede observar
(12) --- FAIL: TestMaximo_LaPasadaCuentaLosQueCrecieron       Crecieron = 0; se esperaba 1 (uno crece y otro repite)
(13) --- FAIL: TestScan_LaSalidaQueCreceValeSuMaximo          … "evento: … in=3 out=7 cw=0 cw5m=0 cw1h=0 cr=0 cost=$0.0002 …"
(24) --- FAIL: TestDesglose_SinEscrituraNoCuentaComoSinDesglose  SinDesglose = 1; se esperaba 0
```
**Razones**:
- «la primera manda»: el evento sale con la primera línea, que en (11) da 100 a 5 min y en (25) 300 a 5 min;
- la pasada no cierra nada;
- `SinDesglose` cuenta aunque la escritura sea 0.

**(14)**, `TestMaximo_UnaLineaDuplicadaNoCambiaNada`, **nace verde**: compara la lectura con y sin la línea duplicada, y con «la primera»
también coincidían. La valida m11.

### T019 · Censo de `internal/ingest/pasada_test.go`

**Antes**, literal:
```go
func leerEnUnaPasada(t *testing.T, p *Pasada, lineas ...[]byte) []*event.Event {
	t.Helper()
	ctx := Context{Salt: "s", Pasada: p}
	var emitidos []*event.Event
	for _, l := range lineas {
		ev, err := FromClaudeCodeLine(l, ctx)
		if err != nil {
			t.Fatalf("precondición: línea corrupta: %v", err)
		}
		if ev != nil {
			emitidos = append(emitidos, ev)
		}
	}
	return emitidos
}
```
y en `TestCasoLimite_ConsumoDistinto` el subtest `conserva_el_consumo_de_la_primera`, que esperaba **140** *(«los 140 de la PRIMERA línea»)*.

**Después**:
- `leerEnUnaPasada` añade al final `for _, c := range p.CerrarFichero() { … emitidos = append(emitidos, &ev) }`;
- el subtest pasa a `cada_partida_vale_su_maximo` y espera **1039** *(999 + 40)*;
- el comentario cita P-007 FR-009, y `cuenta_la_discrepancia` sigue igual.

Con la Fase 0, el subtest nuevo cayó por su razón: `los eventos emitidos suman 140 tokens; se esperaban 999 + 40 = 1039`.

### T020 · Verde

- `internal/ingest/pasada.go`:
  - `mensaje` *(base, primera, máximo y emitido)*, y la `Pasada` con `mensajes` y `abiertos`;
  - `acumular` aplica el máximo por partida y el desglose de la línea del máximo de la escritura *(`>=`, la última si empatan)*. Un
    mensaje ya emitido en la pasada no se reemite;
  - `CerrarFichero` emite en orden de aparición, cuenta `Emitidos` y `Crecieron`, y devuelve el desglose;
  - fuera `registrar`, `vistos` y `Desglose`.
- `internal/ingest/claudecode.go`:
  - `SinDesglose` sólo con escritura > 0 *(E-4)*;
  - el evento base sale de la primera línea;
  - **sin pasada, por línea como hoy**;
  - `conConsumo` pone tokens y coste en un solo sitio.
- `cmd/permea/main.go`:
  - `generate()` encola los cerrados **al final de cada fichero, antes de `st.Save`**;
  - `dryRun()` los imprime al final, con su desglose;
  - se importa `internal/event`.
- **9/9 ok, 458 pass** *(446 + 12)*. `golangci-lint run` → 0.

### T021 · (14)

Nace verde *(ver T018)*. La validó m11: `con una línea duplicada: 1 eventos y 92430 de salida; sin ella: 1 y 91127`.

### T022 · Censo declarado ANTES de mutar *(2026-10-06)*

m9, m10 y m11 tocan **sólo las cuatro partidas** en `acumular`, y dejan la regla del desglose como está. Respecto a `plan.md` y
`tasks.md`, **m9, m10 y m11 declaran más** de lo previsto. Las razones:
- `TestCasoLimite_ConsumoDistinto`, ya con «máximo», cae con cualquier otra regla;
- (11) depende del máximo de la escritura;
- (12) cuenta los que crecen;
- `TestPasada_UnMensajeDeTresLineasEsUnEvento` suma tokens.

| # | Mutación *(en `internal/ingest/pasada.go`, `acumular`, salvo m23)* | Debe caer *(hojas)* |
|---|---|---|
| m9 | máximo → **primera**: no se actualiza ninguna de las cuatro partidas | `TestMaximo_LaSalidaQueCreceValeSuMaximo`, `TestMaximo_ElMaximoEnMedioNoLaUltima`, `TestScan_LaSalidaQueCreceValeSuMaximo`, `TestCasoLimite_ConsumoDistinto/cada_partida_vale_su_maximo`, `TestMaximo_ElDesgloseEsElDeLaLineaDelMaximo/coste` y `/desglose` *(la escritura se queda en 100, y la de 200 la supera)*, `TestMaximo_LaPasadaCuentaLosQueCrecieron` |
| m10 | máximo → **última**: las cuatro partidas toman la de la línea | `TestMaximo_ElMaximoEnMedioNoLaUltima`, `TestCasoLimite_ConsumoDistinto/cada_partida_vale_su_maximo` *(999 + 1)* |
| m11 | **sumar** las cuatro partidas | `TestMaximo_LaSalidaQueCreceValeSuMaximo`, `TestMaximo_ElMaximoEnMedioNoLaUltima`, `TestScan_LaSalidaQueCreceValeSuMaximo`, `TestCasoLimite_ConsumoDistinto/cada_partida_vale_su_maximo`, `TestMaximo_UnaLineaDuplicadaNoCambiaNada` *(14)*, `TestPasada_UnMensajeDeTresLineasEsUnEvento/tokens_de_una_linea`, `TestMaximo_LaPasadaCuentaLosQueCrecieron` *(el que repite también «crece»)* |
| m12 | desglose de la **última** línea *(`if true`)* | `TestMaximo_ElDesgloseEsElDeLaLineaDelMaximo/coste` y `/desglose` |
| m23 | *(`claudecode.go`)* contar sin desglose aunque la escritura sea 0 | `TestDesglose_SinEscrituraNoCuentaComoSinDesglose` |
| m24 | la **primera** entre las empatadas *(`>=` → `>`)* | `TestMaximo_EnEmpateValeLaUltima/coste` y `/desglose` |

### T022 · Resultado *(2026-10-06)*

**Las seis coinciden con lo declarado.** Reversión por edición inversa, con el md5 igual antes y después: `pasada.go` `90a531b9…` y
`claudecode.go` `f29a6519…`.

| # | Mensajes |
|---|---|
| m9 | `salida = 7` ×2; `out=7`; `coste = 0.0010000000, want 0.0024000000`; `desglose = 200 / 0`; `Crecieron = 0`; `suman 140 tokens` |
| m10 | `salida = 1303`; `suman 1000 tokens` |
| m11 | `salida = 91127` ×2; `out=91127`; `Crecieron = 2`; `92430 de salida; sin ella: 1 y 91127`; `suman 420 tokens`; `suman 1140 tokens` |
| m12 | `coste = 0.0010000000, want 0.0024000000`; `desglose = 200 / 0; se esperaba 0 / 300` |
| m23 | `SinDesglose = 1; se esperaba 0` |
| m24 | `coste = 0.0015000000, want 0.0024000000`; `desglose = 300 / 0; se esperaba 0 / 300` |

## B4 · Retención: offset, cierre, `--run`, `--daemon` y resumen

### T024 · Fase 0

- `internal/state`: `Recorrer(path, fn(line, inicio, releida), fijar(leido, modificado))`, todavía con el offset de hoy y sin
  releídas. `ScanFile` es su envoltorio.
- `agent.reloj`.
- El esqueleto de la `Pasada`: `Situar`, `CerrarConReloj` *(cerraba todo)*, `HayNovedades` *(facturables > 0, el predicado de hoy, ya
  usado por `tick`)*, `AvisoDeAbiertos` *(vacío)* y los recuentos `EnEspera`, `Releidas` y `Tardias`.
- **Suite verde, 9/9**, y `state_test.go` sin tocar.

### T025–T028 · Rojos

```
R-B4a TestRecorrer_GuardaElOffsetPedidoYMarcaLasReleidas/offset_pedido   offset guardado = 24; se pidió 8 (…), no el final
      …/releidas                                                         releídas = [false]; se esperaba [true true false]
(i)   TestCierre_ReglaI/mismo_fichero    cerrados = 2 · abierto = 0, hay = false · EnEspera = 0; se esperaba 1
      TestCierre_ReglaI/otro_fichero     cerrados 1 + 1, abiertos false y false
(ii)  TestCierre_ReglaII/b_…, /c_…       con el mtime / el timestamp a T − 1 s, el mensaje debe seguir abierto
(iii) TestCierre_ReglaIII/a_24h_menos_1s_retenido   … debe seguir abierto
(21)  TestCierre_LineaTardia             cerrados = 2 · Tardias = 0; se esperaba 1
      TestCierre_Releidas                Releidas = 0; se esperaba 2 · una pasada que sólo relee y no emite no tiene novedades
(15)  TestRetencion_SC006_DosProcesos/…  [5 7] (sólo P) · offset = 646; se esperaba 323 · [5 7 89817 3]
(16)  TestRetencion_SC007_CierrePorT/b_…, /c_…      cola [42]
(23)  TestRetencion_TopeDe24Horas/a_24h_menos_1s_retenido   cola [42]
(17)  TestRetencion_SC008_CierrePorMensajePosterior/mismo_fichero, /otro_fichero   cola [11 22]
(19)  TestRetencion_AvisoDeRun           stderr no lleva el aviso aprobado
(20)  TestRetencion_TickCallaSiSoloRelee Releidas = 0
(22)  TestRetencion_SC009_NadaDelProveedorEnDisco   precondición: la pasada debe dejar el mensaje abierto
(18)  TestPasada_ElResumenNoLlevaIdentificadores/dos_lineas_aprobadas   el resumen era una línea
```

**Razones**:
- el offset avanzaba al final;
- se cerraba todo al acabar el fichero;
- no había tardías, ni releídas, ni aviso;
- el resumen tenía una línea.

**Nacen verdes**:
- `ReglaII/a` y `SC007/a`: los valida m21;
- `ReglaIII/a_24h` y `TopeDe24Horas/a_24h`: los valida m22;
- `R-B4a/size_y_modtime_del_stat`: lo valida M-B4b.

**(22) no nace verde**: cae en su **precondición**, porque en la Fase 0 no queda ningún mensaje abierto. Sus dos aserciones se cumplían, y
las validan m19 y M-B4d.

### T029 · Censo de `cmd/permea/main_test.go`

**No dieron rojo sin el reloj.** Sus líneas son de 2026-10-02, más de 24 h antes del reloj real, y la regla (iii) *(E-3)* las cierra.
- **Diagnóstico** *(no es una mutación del protocolo)*: sin la (iii), los dos dieron el rojo previsto. El md5 de `pasada.go` fue el mismo
  antes y después *(`4eb7a909…`)*:
  - `TestPasada_GenerateEncolaUnoPorMensaje`: «un mensaje de tres líneas dejó 0 eventos en la cola; se esperaba 1»;
  - `TestActualizar_NoReenviaNiReescribeLaCola`: «la pasada encoló 0 eventos nuevos ([]); se esperaba sólo el de la línea posterior al
    offset».
- **Cambio**: los dos fijan el `mtime` en el `timestamp` de sus líneas *(2026-10-02 12:00Z)* y el reloj a T + 1 s. Desde ahí se cierran
  por la regla (ii) y no dependen de la fecha real.

### T030 · Verde

- **`internal/state/state.go`**: `Recorrer` marca como releída la línea que empieza por debajo del `Size` anterior *(ninguna, si hubo
  truncado)*. Guarda el offset que devuelve `fijar`, acotado a [previo, leído], y `Size` y `ModTime` del stat.
- **`internal/ingest/pasada.go`**:
  - las constantes `EsperaDeCierre` *(10 min)* y `TopeDeEspera` *(24 h)*;
  - la regla (i) en `acumular`, y las tardías;
  - `CerrarConReloj`, con (i), (ii) y (iii), y el comienzo del primer abierto;
  - el resumen en dos líneas;
  - `HayNovedades` *(facturables no releídas, o emitidos)*;
  - `AvisoDeAbiertos`;
  - la cabecera: lo abierto se relee.
- **`internal/ingest/claudecode.go`**: `acumular` recibe el `timestamp`.
- **`cmd/permea/main.go`**:
  - `generate()` usa `Recorrer` y `Situar`, una hora por pasada, y deja el offset en el abierto;
  - `runOnce` escribe el aviso;
  - `tick` usa `HayNovedades`.
- **9/9 ok y `golangci-lint run` → 0**, antes de mutar.

### T031 · Verdes de nacimiento

`ReglaII/a`, `SC007/a`, `ReglaIII/a_24h`, `TopeDe24Horas/a_24h` y `R-B4a/size_y_modtime_del_stat` nacieron verdes, y las dos aserciones de
(22) se cumplían. Todos los validan sus mutaciones *(abajo)*.

### FR-012 y `state.json` *(comprobado el 2026-10-06)*

- **Encolar antes de guardar, como hoy.** En `generate()`, `Recorrer` sólo actualiza el estado **en memoria** *(`s.Files[path]`)*. Los
  cerrados se encolan con `transport.Append` después de `Recorrer`, y **después** va `st.Save(statePath)`, que es lo único que escribe
  `state.json`. Una caída entre medias re-encola, y nunca pierde.
- **Cuatro campos.** `type FileState struct` es **idéntico** al de `24cf6b3` *(comparado con `cmp`)*: `path`, `size`, `mod_time` y
  `offset`. `TestRetencion_SC009_NadaDelProveedorEnDisco/state_json_cuatro_campos` lo comprueba sobre el `state.json` real, en PASS, y
  M-B4d lo valida.

### T032 · Censo declarado ANTES de mutar *(2026-10-06)*

**Dos cambios respecto a `tasks.md`, declarados aquí antes de mutar:**
- **m20 NO tumba `TestProjectJoin_LaPeticionNuncaSeEncola/CASO_POSITIVO`.** Su fixture lleva `timestamp` de 2026-06-20, más de 24 h
  antes del reloj real, así que la regla (iii) de E-3 cierra sus dos mensajes aunque falte la (i). La co-caída que preveía el plan
  *(R-3)* la escribimos antes de E-3.
- **m21 también tumba los dos tests del censo de `main_test.go`**: desde T029 se cierran por la regla (ii).

Otras mutaciones también hacen caer más de lo que preveían las tareas, porque los tests de `ingest` y los de `cmd` cubren las mismas
reglas a dos niveles. **M-B4b** y **M-B4d** son nuevas: validan los subtests que nacen verdes, `size_y_modtime_del_stat` de R-B4a y
`state_json_cuatro_campos` de (22).

| # | Mutación | Debe caer *(hojas)* |
|---|---|---|
| m13 | `generate`: el offset avanza al final *(sin `return abierto`)* | `TestRetencion_SC006_DosProcesos/offset_en_el_comienzo_de_M`, `TestRetencion_TickCallaSiSoloRelee` |
| m14 | (ii) sin la condición del `mtime` | `TestCierre_ReglaII/b_mtime_a_T_menos_1s_retenido`, `TestCierre_ReglaIII/a_24h_menos_1s_retenido`, `TestRetencion_SC007_CierrePorT/b_mtime_a_T_menos_1s_retenido`, `TestRetencion_TopeDe24Horas/a_24h_menos_1s_retenido` |
| m15 | (ii) sin la condición del `timestamp` | `TestCierre_ReglaII/c_timestamp_a_T_menos_1s_retenido`, `TestRetencion_SC007_CierrePorT/c_timestamp_a_T_menos_1s_retenido` |
| m16 | la regla (i) alcanza a los retenidos de otros ficheros | `TestCierre_ReglaI/otro_fichero`, `TestRetencion_SC008_CierrePorMensajePosterior/otro_fichero` |
| m17 | `generate` emite todo al acabar cada fichero, como B3 | `TestRetencion_SC006_DosProcesos/` *(los tres)*, `TestRetencion_SC007_CierrePorT/b_…` y `/c_…`, `TestRetencion_TopeDe24Horas/a_24h_menos_1s_retenido`, `TestRetencion_SC008_…/mismo_fichero` y `/otro_fichero`, `TestRetencion_AvisoDeRun`, `TestRetencion_TickCallaSiSoloRelee`, `TestRetencion_SC009_NadaDelProveedorEnDisco` *(precondición)* |
| m18 | `Recorrer`: ninguna línea es releída | `TestRecorrer_GuardaElOffsetPedidoYMarcaLasReleidas/releidas`, `TestRetencion_TickCallaSiSoloRelee` |
| m19 | `generate` escribe cada línea leída en `pendientes.json` | `TestRetencion_SC009_NadaDelProveedorEnDisco/sin_centinelas` |
| m20 | sin la regla (i) | `TestCierre_ReglaI/mismo_fichero`, `TestCierre_LineaTardia`, `TestRetencion_SC008_…/mismo_fichero`, `TestRetencion_SC006_DosProcesos/` *(los tres)* |
| m21 | (ii) nunca cierra | `TestCierre_ReglaII/a_las_dos_a_T_emitido`, `TestRetencion_SC007_CierrePorT/a_las_dos_a_T_emitido`, `TestPasada_GenerateEncolaUnoPorMensaje`, `TestActualizar_NoReenviaNiReescribeLaCola` |
| m22 | sin la regla (iii) | `TestCierre_ReglaIII/a_24h_emitido`, `TestRetencion_TopeDe24Horas/a_24h_emitido` |
| M-B4a | `Recorrer` guarda el final y no el offset pedido | `TestRecorrer_…/offset_pedido` y `/releidas`, `TestRetencion_SC006_DosProcesos/offset_en_el_comienzo_de_M`, `TestRetencion_TickCallaSiSoloRelee` |
| M-B4b | `Recorrer` guarda como `Size` el offset pedido | `TestRecorrer_…/size_y_modtime_del_stat` y `/releidas`, `TestRetencion_TickCallaSiSoloRelee` |
| M-B4d | `FileState` gana un quinto campo | `TestRetencion_SC009_NadaDelProveedorEnDisco/state_json_cuatro_campos` |

#### m21 · Parada y corrección *(2026-10-06)*

**Parada.** m21 **no coincidió** con lo declarado, y se dejó puesta *(md5 de `pasada.go` `a4d92dd4…`)*.
- **Declarado**: `TestActualizar_NoReenviaNiReescribeLaCola`, es decir, el test entero.
- **Caído**: sólo su hoja `TestActualizar_NoReenviaNiReescribeLaCola/solo_lo_posterior_al_offset`.
- **Verde**: la otra hoja, `…/la_cola_previa_byte_a_byte`, como debe. m21 no toca la cola previa.
- **Lo demás coincidió**: cayeron los otros tres declarados *(`TestCierre_ReglaII/a_las_dos_a_T_emitido`,
  `TestRetencion_SC007_CierrePorT/a_las_dos_a_T_emitido` y `TestPasada_GenerateEncolaUnoPorMensaje`)*.

**Causa**: se nombró el test y no su hoja.

**Corrección autorizada por el orquestador el 2026-10-06** *(y regla E-5: el censo se declara por HOJA, nunca por test padre)*. La
declaración corregida de m21:

| # | Mutación | Debe caer *(hojas)* |
|---|---|---|
| m21 | (ii) nunca cierra | `TestCierre_ReglaII/a_las_dos_a_T_emitido`, `TestRetencion_SC007_CierrePorT/a_las_dos_a_T_emitido`, `TestPasada_GenerateEncolaUnoPorMensaje`, `TestActualizar_NoReenviaNiReescribeLaCola/solo_lo_posterior_al_offset` |

**Las pendientes, declaradas otra vez por hoja con su nombre completo** *(E-5; mismo contenido que la tabla de arriba, ahora sin abreviar)*:

| # | Debe caer *(hojas)* |
|---|---|
| m22 | `TestCierre_ReglaIII/a_24h_emitido`, `TestRetencion_TopeDe24Horas/a_24h_emitido` |
| M-B4a | `TestRecorrer_GuardaElOffsetPedidoYMarcaLasReleidas/offset_pedido`, `TestRecorrer_GuardaElOffsetPedidoYMarcaLasReleidas/releidas`, `TestRetencion_SC006_DosProcesos/offset_en_el_comienzo_de_M`, `TestRetencion_TickCallaSiSoloRelee` *(sin subtests)* |
| M-B4b | `TestRecorrer_GuardaElOffsetPedidoYMarcaLasReleidas/size_y_modtime_del_stat`, `TestRecorrer_GuardaElOffsetPedidoYMarcaLasReleidas/releidas`, `TestRetencion_TickCallaSiSoloRelee` |
| M-B4d | `TestRetencion_SC009_NadaDelProveedorEnDisco/state_json_cuatro_campos` |

### T032 · Resultado *(2026-10-06)*

**Las trece coinciden con lo declarado** *(m21, con el censo corregido)*. Reversión por edición inversa, con el md5 igual antes y después:
- `pasada.go` `4eb7a909…`;
- `state.go` `918a6d46…`;
- `main.go` `589a3533…`.

| # | Mensajes |
|---|---|
| m13 | `offset = 646; se esperaba 323`; `Releidas = 0` |
| m14 | `cola [42]` ×2; «debe seguir abierto» ×2 |
| m15 | `cola [42]`; «debe seguir abierto» |
| m16 | `B en otro fichero no cierra A; cola [11]`; `cerrados 0 + 1` |
| m17 | `[5 7]`; `offset = 644; se esperaba 322`; `[5 7 89817 3]`; `cola [42]` ×3; `cola [11 22]` ×2; sin aviso; `Releidas = 0`; precondición de (22) |
| m18 | `releídas = [false false false]`; `Releidas = 0`; «no debe escribir su resumen» |
| m19 | `pendientes.json contiene un identificador del proveedor (…)` ×3 |
| m20 | `[]` ×2; `offset = 0`; `cola []`; `cerrados = 0` ×2; `EnEspera = 2`; `Tardias = 0`. **`project_test` no cae** |
| m21 | `dejó 0 eventos en la cola`; `encoló 0 eventos nuevos ([])`; `cola []`; «debe cerrarse» |
| m22 | `cola []`; «debe cerrarse aunque el fichero cambie» |
| M-B4a | `offset guardado = 24; se pidió 8`; `releídas = [false]`; `offset = 646`; `Releidas = 0` |
| M-B4b | `Size/ModTime = 8/…; los del stat son 24/…`; `releídas = [false false false]`; `Releidas = 0` |
| M-B4d | `una entrada de state.json tiene 5 campos; se esperaban 4` |

**Inválidas, no cuentan:**
- **m17, primera forma** *(`CerrarFichero()` dejaba `ahora` sin usar; no compilaba)*: revertida con el md5 en `589a3533…`. Se rehízo
  cerrando con el reloj adelantado mil horas.
- **m20**: el script no la pudo revertir porque borraba un bloque. Se revirtió a mano por edición inversa, con el md5 en `4eb7a909…`.
  m22 usa un marcador.

(m17 imprime 644/322 donde la ejecución sin mutar imprimía 646/323: los `timestamp` de ahora cambian de longitud entre ejecuciones.)
