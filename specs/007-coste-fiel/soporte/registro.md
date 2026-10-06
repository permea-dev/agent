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
