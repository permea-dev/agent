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

## B5 · README y CHANGELOG

### De paso · la marca «releída» es una aproximación *(2026-10-06; sólo se declara, sin cambiar conducta)*

`Recorrer` marca como releída la línea que empieza por debajo del `Size` guardado. Ese `Size` es el del stat de la pasada anterior, no lo
que esa pasada leyó. Tiene dos imprecisiones:
1. **Una línea a medio escribir** en la pasada anterior, sin `\n`, no se consumió, pero quedaba por debajo del `Size`. Al completarse se
   cuenta como **releída** sin haberse leído.
2. **Si el fichero creció entre el stat y la lectura**, la pasada leyó más allá del `Size` que guardó. Si esas líneas se releen después,
   se cuentan como **nuevas**.

**Efecto, comprobado en el código.** La marca va de `Recorrer` a `Pasada.Situar`, de ahí a `contarFacturable` y de ahí sólo al recuento
`Releidas`. Ese recuento lo usan **únicamente**:
- la segunda línea del resumen;
- `HayNovedades`, que decide si el demonio escribe el resumen de ese ciclo.

**No** entra en `acumular` ni en `CerrarConReloj`, así que **nunca** cambia lo que se emite. Queda declarado en el comentario de
`Recorrer`.

### T034 · Rojos *(por `grep`)*

```
$ grep -n "Limitación conocida" README.md
150:  **Limitación conocida**: si las líneas de un mismo mensaje traen tokens de salida crecientes, se
$ grep -n "Limitación 1" README.md
175:- **Limitación 1**: la escritura de caché va a la tarifa de **5 minutos**; una escritura de caché de
$ grep -c "^## 0.4.0" CHANGELOG.md
0
```

### T035 · `README.md`

- **§Modos de ejecución**:
  - fuera la «Limitación conocida»;
  - **un evento por mensaje, contado entero** *(cada partida vale lo más alto que alcanza)*;
  - **un mensaje sale cuando está completo** *(cuando empieza el siguiente o tras 10 minutos sin cambios; tope de 24 h; lo abierto se
    relee)*;
  - el resumen, en **dos líneas**;
  - `--scan`, con `cw5m=` y `cw1h=`, cierra todo al final del fichero;
  - `--run` deja lo abierto y avisa, con el texto aprobado;
  - `--daemon` cierra a su hora aunque el fichero no crezca, y sólo escribe el resumen si hay novedades.
- **§Coste y tarifas**:
  - **17 modelos con cinco cifras** y `8f147d1`;
  - la escritura de caché, por duración;
  - **fuera la «Limitación 1»**, y entra la **hipótesis P-1**;
  - la «Limitación 2» sigue.
- **§Instalación**: el ejemplo `PERMEA_VERSION=v0.3.0` pasa a `v0.4.0`.

### T036 · `CHANGELOG.md`

`## 0.4.0 — PENDIENTE` va encima de la 0.3.0, con el cuerpo **sacado por programa** del bloque de spec §Textos aprobados, sin copiarlo
a mano.

```
$ cmp <cuerpo 0.4.0 de CHANGELOG.md> <spec §Textos aprobados>
cmp: sin diferencias (rc=0)
$ grep -c PENDIENTE CHANGELOG.md
1
```

## Cierre

### C1 · T038 · Puertas *(2026-10-06, sobre `07904af`, árbol limpio; Go 1.22.2, golangci-lint 2.12.2)*

Las arquitecturas del release salen de `.goreleaser.yaml`: darwin/amd64, darwin/arm64 y windows/amd64. windows/arm64 está fuera.
`Getenv` no aparece en producción ni en `222c824` ni en `HEAD`.

```
$ gofmt -l .
(rc=0)
$ go vet ./...
(rc=0)
$ golangci-lint run
0 issues.
(rc=0)
$ go test -count=1 ./...
ok  	github.com/permea-dev/agent/cmd/permea	6.027s
ok  	github.com/permea-dev/agent/internal/config	0.015s
ok  	github.com/permea-dev/agent/internal/event	0.002s
ok  	github.com/permea-dev/agent/internal/ingest	0.019s
ok  	github.com/permea-dev/agent/internal/pricing	0.003s
ok  	github.com/permea-dev/agent/internal/project	0.121s
ok  	github.com/permea-dev/agent/internal/state	0.011s
ok  	github.com/permea-dev/agent/internal/testutil	0.008s
ok  	github.com/permea-dev/agent/internal/transport	0.126s
(rc=0)
recuento: tests {'pass': 494} · paquetes {'pass': 9}
$ git diff 222c824 -- internal/event/ internal/ingest/eventid.go internal/ingest/eventid_test.go internal/ingest/boundary_test.go specs/006-medicion-fiel/contracts/event-id.md | wc -c
0
$ git diff 222c824 -- internal/state/state_test.go cmd/permea/project_test.go | wc -c
0
$ grep -n "8f147d1" internal/pricing/pricing.go
25://   - Catálogo replicado: permea-dev/permea-platform · backend/config/pricing.php · 8f147d1
$ grep -rn nolint --include=*.go . | wc -l
1
internal/transport/adhesion_test.go:194:	_, errDirecto := url.Parse(endpointNoAnalizableAdhesion) //nolint:staticcheck // SA1007: la URL inválida es el SUJETO del test — la causa de url.Parse que Adherir debe conservar (P-006 E-006-P4)
$ grep -c PENDIENTE CHANGELOG.md
1
$ CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o /dev/null ./cmd/permea
(rc=0)
$ CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /dev/null ./cmd/permea
(rc=0)
$ CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /dev/null ./cmd/permea
(rc=0)
$ git diff 222c824 -- "*.go" | grep "^+" | grep -v "^+++" | grep -c Getenv
0
$ grep -rn Getenv --include=*.go cmd internal | grep -v _test.go   (HEAD)
$ git grep -n Getenv 222c824 -- "cmd/*.go" "internal/*.go" | grep -v _test.go   (222c824)
$ git log --oneline 222c824..HEAD
07904af 007 B5: README y CHANGELOG de la 0.4.0
c4596b4 007 B4: un mensaje se envia cuando esta cerrado
24cf6b3 007 B3: cada partida vale el maximo de sus lineas
5e3e5a0 007 B2: la escritura de cache se tarifa por su duracion
3dc064c 007 B1: tarifas de 17 modelos con cinco cifras (8f147d1)
0d90b5c 007 B0: enmienda E-3 y contrato de tarifas de 006 sustituido
49cf73c 007: spec ratificada (E-1, E-2), plan, tareas, contrato de tarifas y quickstart
$ git diff --stat 222c824
 .specify/feature.json                          |   2 +-
 CHANGELOG.md                                   |  25 ++
 README.md                                      |  50 ++-
 cmd/permea/coste_test.go                       |  70 ++++
 cmd/permea/main.go                             |  66 +++-
 cmd/permea/main_test.go                        |  18 +-
 cmd/permea/retencion_test.go                   | 309 ++++++++++++++++
 internal/ingest/cierre_test.go                 | 149 ++++++++
 internal/ingest/claudecode.go                  |  82 +++--
 internal/ingest/desglose_test.go               | 122 +++++++
 internal/ingest/maximo_test.go                 | 147 ++++++++
 internal/ingest/pasada.go                      | 248 +++++++++++--
 internal/ingest/pasada_test.go                 |  31 +-
 internal/pricing/pricing.go                    |  78 ++--
 internal/pricing/pricing_test.go               | 101 ++++--
 internal/state/retener_test.go                 |  78 ++++
 internal/state/state.go                        |  39 +-
 specs/006-medicion-fiel/contracts/tarifas.md   |   2 +
 specs/007-coste-fiel/contracts/tarifas.md      | 111 ++++++
 specs/007-coste-fiel/plan.md                   | 195 ++++++++++
 specs/007-coste-fiel/quickstart.md             | 229 ++++++++++++
 specs/007-coste-fiel/soporte/descubrimiento.md | 350 ++++++++++++++++++
 specs/007-coste-fiel/soporte/registro.md       | 472 +++++++++++++++++++++++++
 specs/007-coste-fiel/spec.md                   | 306 ++++++++++++++++
 specs/007-coste-fiel/tasks.md                  | 308 ++++++++++++++++
 25 files changed, 3417 insertions(+), 171 deletions(-)
```

**Resultado: todo cuadra.**
- **SC-012**: 494 pass en 9/9 paquetes, y lint a 0.
- **SC-009 y FR-018**: la frontera sin diff; `state_test.go` y `project_test.go`, tampoco.
- **La cabecera** cita `8f147d1`.
- **Un `nolint`**, el de SA1007 *(E-006-P4)*.
- **`PENDIENTE` = 1**: se resuelve en C5.
- **Compilan** las tres arquitecturas del release.
- **Ningún `Getenv` nuevo**.
- **Siete commits** desde `222c824`.

### C2 · T039 · Medidas sobre las copias del dueño *(2026-10-06; binario de la rama sobre `07904af`)*

**Cómo**:
- **Binario y sandbox**: el binario de la rama, en un `mktemp -d /tmp/permea-007-XXXXXX`, dentro de `env -i` con `HOME` y
  `XDG_CONFIG_HOME` temporales. Antes de medir, `permea status` → «no enrolado».
- **`--scan`**: fichero a fichero, con las copias **leídas en su sitio**. 0 ficheros con rc ≠ 0, 0 líneas `skip` y 0 ficheros escritos en
  el sandbox.
- **Scripts**: el contador y la comprobación del coste, extraídos de `quickstart.md`.
- **Lo que queda**: las salidas con `event_id`, sólo en el temporal, que se borró tras comprobar el prefijo. Aquí, sólo recuentos y sumas.

**SC-001 *(copia del dueño 2026-10-06-wsl)* — ✅**

| Partida | Referencia | Agente *(`--scan`)* | Contador *(máximo)* | Contador *(primera)* |
|---|---:|---:|---:|---:|
| mensajes / eventos | 10 121 | 10 121 | 10 121 | — |
| entrada | 26 392 | 26 392 | 26 392 | — |
| salida | 11 713 955 | 11 713 955 | 11 713 955 | 11 713 955 |
| escritura de caché | 32 765 802 | 32 765 802 | 32 765 802 | — |
| lectura de caché | 4 499 158 833 | 4 499 158 833 | 4 499 158 833 | — |

Del contador: facturables 22 181 · sintéticas 1 · sin identificador 0 · corruptas 0 · `crecen=0` · líneas sin desglose 0 · desglose que no
suma 0.

**SC-003 *(-wsl)* — ✅**: 10 121 eventos = 10 121 mensajes, y **0** `event_id` repetidos.

**SC-002 *(copia del dueño 2026-10-06-windows)* — ✅**

| Partida | Referencia | Agente | Contador *(máximo)* | Contador *(primera)* |
|---|---:|---:|---:|---:|
| mensajes / eventos | 6 074 | 6 074 | 6 074 | — |
| entrada | 12 206 | 12 206 | 12 206 | — |
| **salida** | **6 723 801** | **6 723 801** | **6 723 801** | **6 635 290** |
| escritura de caché *(5 min / 1 h)* | 22 307 249 *(971 559 / 21 335 690)* | ídem | ídem | — |
| lectura de caché | 2 423 738 410 | 2 423 738 410 | 2 423 738 410 | — |

Los que crecen *(auxiliar independiente, sólo sumas)*: **143**. Su salida es **89 817** con «máximo» y 1 306 con «primera», como dice la
spec. En -wsl, 0. `event_id` repetidos: 0. Facturables 13 869 · sintéticas 2.

**SC-005 *(-wsl)***:
- **El reparto — ✅**: a 5 min, **247 506**; a 1 h, **32 518 296**. Coinciden la referencia, el agente y el contador.
- **El coste por evento — ✅ tras E-6**:

| | Instrumento | -wsl *(10 121)* | -windows *(6 074)* |
|---|---|---|---|
| Antes *(Encargo 10)* | `float`, distinto si `abs(…) > 5e-5` | `coste_distinto=3` | `coste_distinto=4` |
| Después *(Encargo 10-bis, E-6)* | `Decimal` exacto | `coste_distinto=0 empates=31 sin_tarifa=0` | `coste_distinto=0 empates=18 sin_tarifa=0` |

**Las dos paradas y su resolución**:
1. **Encargo 10, PARA en SC-005**: el script daba `coste_distinto=3` *(-wsl)*. Medidas en `Decimal`, las 3 *(y las 4 de -windows)* estaban
   a **exactamente** 0,00005 del exacto. Eran empates de redondeo: el coste exacto acaba en 5 en la quinta cifra y `%.4f` lo redondea
   desde el binario. La comparación en coma flotante los empujaba por encima del borde de su tolerancia.
   **Resolución, E-6**: el script pasa a aritmética decimal exacta, y el empate vale.
2. **Encargo 10-bis, PARA en el paso 2**: el esperado era «3 y 4 empates» y salieron **31 y 18**. Un diagnóstico de sólo lectura
   clasificó cada empate decimal según lo que daba la comparación en `float`.
   **Resolución**: el orquestador corrigió su esperado, que contaba sólo los empates visibles en coma flotante. Las cifras correctas son
   las medidas.

| Copia | Empates en decimal | …en `float` daban «distinto» | …en `float` pasaban sin verse | No empates que en `float` daban «distinto» |
|---|---:|---:|---:|---:|
| -wsl | 31 | 3 | 28 | 0 |
| -windows | 18 | 4 | 14 | 0 |

**Medida informativa** *(no es un SC)*: por modelo, el coste de la 0.4.0 y la parte debida a tarifar la escritura de 1 hora a su cifra,
Σ `cw1h` × (`cache_write_1h` − `cache_write`) / 10⁶. En USD, recalculado con la tabla del contrato.

| Copia | Modelo | Eventos | Coste 0.4.0 | Por la escritura de 1 h | % |
|---|---|---:|---:|---:|---:|
| -wsl | `claude-opus-5` | 5 727 | 1 689,604355 | 66,074771 | 3,91 % |
| -wsl | `claude-opus-5-5` | 4 364 | 575,607939 | 43,330350 | 7,53 % |
| -wsl | `claude-sonnet-5` | 30 | 2,853042 | 0,682361 | 23,92 % |
| -wsl | **total** | **10 121** | **2 268,065336** | **110,087482** | **4,85 %** |
| -windows | `claude-opus-4-8` | 7 | 0,627783 | 0,165862 | 26,42 % |
| -windows | `claude-opus-5` | 1 453 | 362,655623 | 17,215882 | 4,75 % |
| -windows | `claude-opus-5-5` | 4 614 | 617,702031 | 50,101674 | 8,11 % |
| -windows | **total** | **6 074** | **980,985437** | **67,483419** | **6,88 %** |

`sin_tarifa=0` en las dos copias. En -windows, la subida frente a la 0.3.0 incluye además los 88 511 tokens de salida que la 0.3.0 no
contaba.

**Huellas de las copias**, antes = después en las **cuatro** ejecuciones *(Encargo 10; Encargo 10-bis, SC-005 y diagnóstico)* e iguales a
las de E-1:

| Copia | Huella SHA-256 del conjunto | `.jsonl` | Ficheros |
|---|---|---:|---:|
| 2026-10-06-wsl | `82d79809411f61a3` | 31 | 50 |
| 2026-10-06-windows | `d65f53cf7c56325c` | 29 | 78 |

**Temporales**: `/tmp/permea-007-eKE9Jx`, `/tmp/permea-007-98z1vn` y `/tmp/permea-007-eSxU4g`, borrados tras comprobar el prefijo. No
queda ningún `/tmp/permea-007-*`.

### C3 · T040 · Snapshot *(2026-10-06, sobre `07904af`; goreleaser v2.16.0)*

**Lo que se construyó y desde dónde**:
- `goreleaser release --snapshot --clean` → rc 0, «skipping announce, publish, and validate». **No se publicó nada**: no había ninguna
  credencial en el entorno, y el cask y el manifiesto de Scoop sólo se escribieron en `dist/`.
- El hook `go mod tidy` dejó `go.mod` igual *(md5)*.
- El árbol tenía modificados sólo cuatro documentos de `specs/007-coste-fiel/`, así que el código es el de `07904af`.

**Versión inyectada: `0.3.0-SNAPSHOT-07904af`.** goreleaser parte de la última etiqueta, `v0.3.0`. No es ni `0.0.1-dev` ni `0.3.0`.

| Artefacto | Bytes | SHA-256 |
|---|---:|---|
| `dist/permea_0.3.0-SNAPSHOT-07904af_windows_amd64.zip` | 2 484 829 | `897ffdd170e4f52e9d73912da8daec4052d8e9f75f670bf5b4f6c7f91cd8de1f` |
| `dist/permea_0.3.0-SNAPSHOT-07904af_linux_amd64.tar.gz` | 2 416 049 | `b5f15d51aa9403eb136357297fbd929bbd0849051321a7baafcad15011ce5280` |
| `dist/permea_0.3.0-SNAPSHOT-07904af_linux_arm64.tar.gz` | 2 229 960 | `b7c431c487a18d96e9c8b1f2781fbad2c54120d1e6ca1c763aa9eed9bef89495` |
| `dist/permea_0.3.0-SNAPSHOT-07904af_darwin_amd64.tar.gz` | 2 460 846 | `26e95bf03af2ba4e0a9e7a58cb0ad3b432ae3f32081e18cf24902406b7234e57` |
| `dist/permea_0.3.0-SNAPSHOT-07904af_darwin_arm64.tar.gz` | 2 318 501 | `dbb0c189123ce4d371eee5616e1b53e879c54d7470ea4e573e084a8063b72944` |

`sha256sum -c` del fichero `permea_0.3.0-SNAPSHOT-07904af_checksums.txt`: OK en los 5.

**Las comprobaciones**:
- **Sandbox**: el binario de Linux, sacado del `.tar.gz`, en `env -i` con `HOME` temporal.
  - `--version` → **`0.3.0-SNAPSHOT-07904af`** *(rc 0)*;
  - `help` → rc 0, 41 líneas por stdout;
  - 0 ficheros creados en el `HOME`.
- **SC-011**: los textos se extrajeron de spec §Textos aprobados por programa. En `strings -e S` del ejecutable de Linux y en el de
  Windows *(`permea.exe`)* aparece **1** vez cada uno:
  - la segunda línea del resumen;
  - el aviso de `--run`;
  - la línea de `--scan`, con `cw5m=` y `cw1h=`;
  - la primera línea de 006.

  La comprobación cruzada sobre los bytes de cada ejecutable da lo mismo.
- **README y LICENSE** dentro del `.zip` de Windows y del `.tar.gz` de Linux: `cmp` con los del repo → **idénticos** los cuatro. El README
  empaquetado es el de la 0.4.0: cita `8f147d1` y los 17 modelos, y no lleva «Limitación conocida» ni «Limitación 1».
- **`git status`**: no muestra `dist/`, que ignora `.gitignore:20 /dist/`.

**Una trampa del instrumento**: en el shell de la sesión, `grep` es una función del perfil que llama a ugrep con `-I`, y salta los
ficheros que parecen binarios. Sobre la salida de `strings` daba 0 coincidencias sin avisar. Las comprobaciones de SC-011 se hicieron con
`command grep -a` *(GNU grep)*. Las medidas anteriores con `grep` fueron sobre texto, y sus cuentas coincidieron con las de `awk` o
Python. Las de C2, además, fueron sobre las salidas de `--scan`: `grep -c '^evento:'` dio lo mismo que `awk`.

### C4 · T041 · W1 *(hecho por el dueño el 2026-10-06, de 21:47 a 22:04, Madrid, en Windows; en sandbox, E-7)*

**Lo que se probó y cómo**:
- **Zip**: `permea_0.3.0-SNAPSHOT-07904af_windows_amd64.zip`. SHA-256 comprobado:
  `897FFDD170E4F52E9D73912DA8DAEC4052D8E9F75F670BF5B4F6C7F91CD8DE1F`. Es el de C3, en mayúsculas.
- **Sandbox** *(E-7)*: un directorio de datos aparte *(la variable `APPDATA` de esa consola)*, sin enrolar.
- `--version` → `0.3.0-SNAPSHOT-07904af` · `status` → `no enrolado`.
- Antes de la primera pasada, un mensaje nuevo con Claude Code en Windows.

**Primera pasada** *(21:50)*, salida literal:
```
Permea 0.3.0-SNAPSHOT-07904af
6154 eventos encolados en <directorio de datos del ensayo>\permea\queue.jsonl
pasada: 14041 líneas facturables · 6154 eventos · 7883 repetidas del mismo mensaje · 2 sintéticas · 0 sin identificador (no contables) · 158 con consumo distinto de la primera
pasada: 143 mensajes que crecieron entre líneas · 2 en espera de cerrarse · 0 líneas releídas de un mensaje en espera · 0 líneas tardías · 0 líneas sin desglose de caché (a 1 hora)
2 mensajes siguen abiertos: se enviarán en la próxima pasada
sync omitido: sin endpoint configurado
```

**Segunda pasada** *(22:04, 14 minutos después, sin usar Claude Code en Windows entre medias)*, salida literal:
```
Permea 0.3.0-SNAPSHOT-07904af
2 eventos encolados en <directorio de datos del ensayo>\permea\queue.jsonl
pasada: 3 líneas facturables · 2 eventos · 1 repetidas del mismo mensaje · 0 sintéticas · 0 sin identificador (no contables) · 0 con consumo distinto de la primera
pasada: 0 mensajes que crecieron entre líneas · 0 en espera de cerrarse · 3 líneas releídas de un mensaje en espera · 0 líneas tardías · 0 líneas sin desglose de caché (a 1 hora)
sync omitido: sin endpoint configurado
```

**Lectura del dueño**:
- Las cuentas cierran: 6154 + 7883 + 2 + 2 = 14041.
- Los 143 que crecen son los de la copia.
- Los 2 abiertos salieron en la pasada siguiente, releídos del log.
- La regla (ii) funciona en NTFS.
- Nada se transmitió.

**Contraste con C2** *(E-7: sustituye al `--scan` contra el contador)*. La pasada lee el historial completo de esa instalación, no la copia
-windows, así que las líneas y los mensajes no tienen por qué coincidir:

| | Copia -windows *(C2)* | Primera pasada de W1 |
|---|---:|---:|
| mensajes que crecen | 143 | 143 |
| sintéticas | 2 | 2 |
| líneas facturables | 13 869 | 14041 |
| mensajes / eventos | 6 074 | 6154 eventos + 2 en espera |

**Resultado: W1 sin fallos.** Los cuatro pasos de quickstart §W1 *(E-7)* dieron lo esperado.

### C6 · T043 · Fusión *(el dueño, 2026-10-07)*

- El PR #4, fusionado con merge commit **`03ae23c`**. La rama remota se borró tras `MERGED`.
- Sobre `main`: 9 paquetes ok, lint 0 y `PENDIENTE` 0.

### C7 · T044 · Etiqueta *(el dueño, 2026-10-07)*

- La etiqueta anotada **`v0.4.0`**, sobre `03ae23c`.
- El flujo `release` terminó en verde *(GoReleaser, 56 s)*.

### C8 · T045 · Canales *(el orquestador, 2026-10-07, con una descarga anónima de lo publicado)*

- **Los cinco archivos de la release**: `sha256sum -c` del fichero de checksums, OK en todos.

| Archivo | SHA-256 |
|---|---|
| `permea_0.4.0_darwin_amd64.tar.gz` | `f9342465600fc70612fadcc954a0d3e497e90e2e43a739e792a7e6d5f90e99a8` |
| `permea_0.4.0_darwin_arm64.tar.gz` | `c2bb79812f45b5d44623ad1b661c3ddaa0524a42f6374f2c68dae8d35a162602` |
| `permea_0.4.0_linux_amd64.tar.gz` | `e31041de11d0ad9b71bda82bd18552a6a42d58ae614ad258d0f72327e22fc883` |
| `permea_0.4.0_linux_arm64.tar.gz` | `6e5a7cd0fee4a33c6b575dce894d88558eb6045a9caac900986778f678f174a1` |
| `permea_0.4.0_windows_amd64.zip` | `0f360e0397c13ebb080e5d84a604ea06f330d3452ba16d24cde5fc3acf8c27e0` |

- **Scoop**: el manifiesto, en `0.4.0`, con el hash del zip de Windows. **Homebrew**: el cask, en `0.4.0`, con los cuatro hashes de macOS
  y Linux.
- **El binario de Linux publicado, en sandbox**: `--version` → `0.4.0`; `status` → `no enrolado`; `help` → 41 líneas.
- **SC-011**: la segunda línea del resumen, el aviso de `--run` y `cw=%d cw5m=%d cw1h=%d cr=%d` aparecen **una** vez en el ejecutable de
  Linux y **una** en el de Windows.
- **El README empaquetado** es el de la 0.4.0: 17 modelos y `8f147d1`, sin «Limitación conocida» ni «Limitación 1».

**Comprobado aquí** *(Encargo 13, sólo lectura)*:
- `gh release view v0.4.0`:
  - título `v0.4.0`, etiqueta `v0.4.0`, `draft: false`, `prerelease: false`;
  - autor `github-actions[bot]`, creada `2026-10-07T08:02:42Z` y publicada `2026-10-07T08:03:47Z`;
  - seis assets: el fichero de checksums y los cinco archivos de arriba;
  - el changelog de la release lista los commits de `222c824` a `03ae23c`.
- El fichero `permea_0.4.0_checksums.txt` publicado, descargado a un temporal y borrado después, trae **los mismos cinco hashes**
  *(`diff` vacío)*.
- En el repo, `git cat-file -t v0.4.0` → `tag` *(anotada)*, y apunta a `03ae23c`.

**Coincide con lo del orquestador.**

### C9 · T046 · W2 *(el dueño, 2026-10-07, en la instalación real de Windows, enrolada)*

- `scoop update permea`: de 0.3.0 a 0.4.0, con el hash comprobado por Scoop.
- `permea --version` → `0.4.0`. `permea status` → enrolado, token configurado.

**Primera pasada** *(10:10, Madrid)*, salida literal:
```
Permea 0.4.0
1005 eventos encolados en <directorio de datos>\queue.jsonl
pasada: 2290 líneas facturables · 1005 eventos · 1283 repetidas del mismo mensaje · 1 sintéticas · 0 sin identificador (no contables) · 0 con consumo distinto de la primera
pasada: 0 mensajes que crecieron entre líneas · 1 en espera de cerrarse · 0 líneas releídas de un mensaje en espera · 0 líneas tardías · 0 líneas sin desglose de caché (a 1 hora)
1 mensajes siguen abiertos: se enviarán en la próxima pasada
1005 eventos transmitidos y confirmados
```
En la plataforma, eventos de la versión `0.4.0` a las 08:11:27 UTC: 1005 eventos · entrada 2038 · salida 1169812 · escritura 4210955 ·
lectura 412521097 · coste 139.619963 · con coste 1005.

**Segunda pasada** *(10:26, Madrid; Claude Code seguía en uso en otra carpeta)*, salida literal:
```
Permea 0.4.0
45 eventos encolados en <directorio de datos>\queue.jsonl
pasada: 108 líneas facturables · 45 eventos · 62 repetidas del mismo mensaje · 0 sintéticas · 0 sin identificador (no contables) · 0 con consumo distinto de la primera
pasada: 0 mensajes que crecieron entre líneas · 1 en espera de cerrarse · 2 líneas releídas de un mensaje en espera · 0 líneas tardías · 0 líneas sin desglose de caché (a 1 hora)
1 mensajes siguen abiertos: se enviarán en la próxima pasada
45 eventos transmitidos y confirmados
```
En la plataforma, a las 08:26:29 UTC: 1050 eventos · entrada 2128 · salida 1234124 · escritura 4294409 · lectura 428932769 · coste
144.856529 · con coste 1050.

**Lectura del orquestador**:
- La primera pasada envió sólo lo posterior al offset de la 0.3.0 *(FR-019)*.
- 1005 + 1283 + 1 + 1 = 2290, y 45 + 62 + 1 = 108.
- La plataforma = lo transmitido en las dos pasadas *(1005, y 1005 + 45)*.
- El mensaje en espera de la primera salió en la segunda, releído del log.
- El incremento de coste, 5,236566, es exactamente el de `claude-opus-5-5` con la escritura a 1 hora sobre los incrementos de tokens
  *(90 de entrada, 64 312 de salida, 83 454 de escritura y 16 411 672 de lectura)*.

> ⚠️ **Declarado**: el quickstart pedía contrastar W2 con el contador independiente y con «en espera» a 0 tras 10 minutos. No se hizo
> así: la instalación estaba en uso *(siempre quedó 1 mensaje abierto)*, y el contraste fue transmitidos = plataforma, más el coste
> recalculado. El contador independiente quedó acreditado en C2 sobre las copias, con el mismo código, y «en espera → 0», en W1.

**WSL** *(10:28, Madrid)*:
- `install.sh` con el checksum verificado, `0.4.0`, enrolado.
- Una pasada de 1841 líneas facturables · 816 eventos · 1025 repetidas · 0 en espera.
- «816 eventos transmitidos y confirmados».

**Comprobado aquí** *(aritmética sobre las cifras de arriba, en `Decimal`)*:
- Los incrementos de la plataforma son 90 · 64 312 · 83 454 · 16 411 672, y 1050 − 1005 = 45 eventos.
- (90 × 4 + 64 312 × 20 + 83 454 × 8 + 16 411 672 × 0,2) / 10⁶ = **5,2365664** USD, con la tabla del contrato para `claude-opus-5-5`
  *(entrada, salida, escritura a 1 h y lectura)*. Redondeado a 6 decimales es **5,236566** = 144,856529 − 139,619963.
- 1005 + 1283 + 1 + 1 = 2290 · 45 + 62 + 1 = 108 · y, en WSL, 816 + 1025 = 1841.

**Resultado: W2 sin fallos**, con el contraste declarado arriba.
