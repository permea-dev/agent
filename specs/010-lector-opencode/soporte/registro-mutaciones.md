# 010 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007, 008 y 009, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. Techo: **300**
líneas *(tasks §Presupuesto)*.

## B0 · Documentos

### T001 · Línea base *(2026-10-11, sobre `0256098`; sólo documentos desde `82e0b2d`)*

```
$ git log --oneline -2
0256098 010 remate del analyze: enmienda E-1, SC-028 y SC-029, mutaciones re-derivadas
4f7658a 010 plan y tareas: lector propio en internal/sqlitero y frontera por rowid
$ git diff 82e0b2d --name-only                      4 rutas, todas bajo specs/010-lector-opencode/
$ gofmt -l .                                        (vacío)
$ go vet ./...                                      (limpio)
$ golangci-lint run
0 issues.
$ go test -count=1 -v ./...                         671 «--- PASS», 0 «--- FAIL», 0 «--- SKIP» · 9 paquetes ok
$ git diff 4c6cafc -- <frontera de FR-030> | wc -l
0
$ git diff 4c6cafc --name-only --diff-filter=M -- '*_test.go' | wc -l
0
$ grep -c require go.mod ; ls go.sum
0 ; no existe
```

## B1 · Lector *(`internal/sqlitero`)*

### T003 · Fase 0 *(2026-10-11, sobre `f486c55`)*
Origen `tmp/sonda-sqlite/b-propio/lector/lector.go` **`f6b47fa8…`** → portado `7e89b932…` *(`package sqlitero`, su comentario, stubs y `ErrCorrupta`/`ErrNoSoportada`)*. Suite **671 / 0 / 0**.

### T004 · Testdata *(`generar.sh`: `sqlite3` 3.45.1 y `python3` 3.12; 13 ficheros, 1 183 216 B; sin `-shm`)*
md5: `esperado.json 2aac2bcd` · `i/opencode.db 68833bf4` · `i/…-wal d2b041e4` · `r/opencode.db 68833bf4` · `r/…-wal df39662e` · `s/opencode.db 1ef488f7` ·
`profundo.db de0010d9` · `raros/`: `desborda 577ef045` · `p512 8dc7c104` · `p65536 8fdced77` · `reserva e7329296` · `utf16 8d057446` · `without_rowid 65490471`.

### T005 · Rojos *(55 tests nuevos = 40 hojas + 15 semillas; los 10 numerados del plan)*
```
FAIL TestCabecera/utf16, TestTablas/without_rowid   errores de texto, sin ErrNoSoportada
FAIL TestBuscar ×4, TestRecorrerDesde ×2            «sin implementar» (stubs)
FAIL TestErroresSinRuta/inexistente                 open /tmp/…/CENTINELA-dir/opencode.db: no such file or directory
FAIL TestCorruptas ×4                               «ciclo en el arbol» · «pagina 2147483647 fuera…» · «base corrupta o truncada» ×2, sin ErrCorrupta
```
Nacen verdes las otras 26 hojas y las 15 semillas.

### T006 · Verde *(md5 `6acf4700…`)*
Poda, `Buscar`, `Paginas`, errores tipados y `sinRuta`; 17 arreglos de lint sin `nolint`; regresión de uno de ellos *(22 FAIL)* corregida con `finCab`;
(10) con `leerTodo`. **726 / 0 / 0** · lint 0.

### T007 y T008 · Censo *(declarado ANTES de mutar)* y mutaciones *(base `6acf4700…`; suite entera; edición inversa; md5)*
| Mut. | Edición | Declarado | Cayó · md5 mutado |
|---|---|---|---|
| mB1 | `if cab[18] == 2 \|\| cab[19] == 2 {` → `if false {` | TestInstantaneas/I, /R · TestEstable/sin_cambio | = · `3ba4bfee` |
| mB2 | `if s0 != be.Uint32(fr[16:]) \|\| …` → `if false {` | TestInstantaneas/R | = · `72cb83c6` |
| mB3 | `if local == total {` → `if true {` | TestInstantaneas/I, /R, /S · TestDesbordamiento/8k, /64k | = · `a65c0040` |
| mB4 | `cadena[:ultimoCommit]` → `cadena[:len(cadena)]` | TestInstantaneas/I, **/R** | = · `efd8ea23` |
| mB5 | `if clave <= desde {` → `… && false {` | TestRecorrerDesde/paginas | = · `ccfdfb13` |
| mB6 | `if clave >= rowid {` → `if clave > rowid {` | TestBuscar/clave | = · `c69e8143` |
| mB7 | en `Abrir`, `return nil, sinRuta(err)` → `return nil, err` | TestErroresSinRuta/inexistente | = · `f3be0653` |
| mB8 | el tope `… \|\| int64(total) > …` fuera | TestCorruptas/carga_enorme · FuzzBase/seed#9 | = · `55be00aa` |
| mB9 | `return err == nil && …` → `return true \|\| (…)` | TestEstable/cambia | = · `95049b2b` |
| mB10 | el `if` del `CREATE` → `if false {` **y** `case 0x02, 0x0a:` de `recorrer` → `case 0xFF:` | TestTablas/without_rowid | = · `be78256c` |
| M-B1a | `enc != 1 && enc != 0 {` → `… && false {` | TestCabecera/utf16 | = · `de485931` |
| M-B1b | tras `b := &Base{…}`, `os.Create(ruta + "-shm")` | TestSoloLectura/listado, /md5 | = · `196bddd1` |

- **mB5**: la primera forma *(`if false {`)* no compilaba y no cuenta; redefinida el 11-10.
- **mB7**: el guion no pudo revertirla, porque su texto mutado no era único; se revirtió anclada en `os.Open` *(md5 = base)* y se repitió limpia.
- **M-B1b**: co-caída (10/md5) no declarada en la primera derivación: la hoja compara todos los ficheros de la carpeta; corregida por decisión del dueño (11-10), sin tocar el test.

### Desviaciones ratificadas por el dueño *(2026-10-11 16:50; equivalencia con la prueba, en C2)*
| Desviación | Dónde |
|---|---|
| R se rompe en la `}` de una celda de `data`, en el frame 34, no en el último byte de la hoja | T004 |
| (1/utf16), (6/without_rowid) y (9) nacen rojos | T005 |
| La co-caída (2/R) de mB4, declarada antes de mutar | T007 |
| 17 arreglos de lint, con su regresión cazada y corregida: el lector ya no es byte a byte el de la prueba *(desviación de D-3)* | T006 |
| Tests al +52 % *(683 líneas frente a ~450)* | T005 |

## B2 · Identidad *(`internal/ingest/opencode_eventid.go`)*

### T010–T012 *(2026-10-11, sobre `f373016`)*
Stub → **726 / 0 / 0**. Rojos de T011, con su razón real: `TestEventIDOpenCode_VectoresDelContrato/{llamada,otra_llamada}`
*(«ok = false; se esperaba un event_id»)* y `TestEventIDOpenCode_EspacioDeNombresPropio` *(«= "", ok false»; su precondición, el literal de Gemini contra
`derivarEventIDGemini`, pasa)*. Nace verde `TestEventIDOpenCode_SinIDNoHayEventID` *(m3)*. Verde: la derivación replicada, con `dominioEventID`,
`bytesEventID` y `tipoRespuesta`; md5 `0a3c19d4…`; lint 0; **731 / 0 / 0**.

### T013 · Censo *(declarado ANTES de mutar)* y mutaciones *(base `0a3c19d4…`; suite entera; edición inversa; md5)*
| Mut. | Edición | Declarado | Cayó · md5 mutado |
|---|---|---|---|
| m1 | `herramientaOpenCode = "opencode"` → `"gemini"` | VectoresDelContrato/llamada, /otra_llamada · EspacioDeNombresPropio | = · `496b16d5` |
| m2 | `_, _ = h.Write(longitud[:])` → `// m2: sin prefijo de longitud` | VectoresDelContrato/llamada, /otra_llamada | = · `e5868ac3` |
| m3 | `if id == "" {` → `if id == "" && false {` | SinIDNoHayEventID | = · `30ad5835` |
| m4 | `…, tipoRespuesta, id}` → `…, tipoRespuesta, "m4", id}` | VectoresDelContrato/llamada, /otra_llamada | = · `bb31b82d` |

Las cuatro, **= declarado**, y restauradas a `0a3c19d4…` antes de la siguiente; directa e inversa comprobadas únicas antes de mutar.

## B3 · Una llamada *(`internal/ingest/opencode.go`)*

### T015–T019 *(2026-10-12, sobre `b65752d`)*
- **T015**: stub `427e86a1…` → **731 / 0 / 0**. Para compilar los rojos, las tres constantes de `ClaseOpenCode` *(el cero no es ninguna)*: `89eaf955…`.
- **T016**: 16 `data` sintéticos en `testdata/opencode/filas/`. El `id` es columna: sus formas *(nulo, `""`, entero)* las pone el test sobre `evento.json`.
- **T017**: 9 rojos, 25 hojas; todos por el stub, «clase 0, evento <nil>». (22): «se esperaba un evento que inspeccionar» *(H19)*. Hojas
  añadidas al plan, nacidas rojas: (17/antes_que_incoherente), (17/event_id), (18/coste_cero) y (20/tal_cual).
- **T018–T019**: verde `1b00c194…`, sin `internal/pricing`; lint 0; **765 / 0 / 0**; (22) en verde.

### T020 · Censo *(declarado ANTES de mutar; prefijo `TestLlamadaOpenCode_`)* y mutaciones *(base `1b00c194…`; suite entera; edición inversa; md5)*
| Mut. | Edición | Declarado | Cayó · md5 mutado |
|---|---|---|---|
| m5 | `int(k.Output + k.Reasoning)` → `int(k.Output)` | PartidasSC012 | = · `ea606ea7` |
| m6 | `TokensCacheCreation: 0` **y** `TokensCacheRead: int(k.Cache.Read + k.Cache.Write)` | PartidasSC012 | = · `1669ecbe` |
| m7 | `int(k.Input)` → `int(k.Input + k.Cache.Read)` | PartidasSC012 | = · `e759c74a` |
| m8 | `CostAvailable: false` → `true` | SinCoste/cost_available | = · `1be824c4` |
| m9 | `Cost float64` en `datosOpenCode` **y** `CostUSD: d.Cost` | SinCoste/cost | = · `3ee005d9` |
| m10 | `Created int64` en `Time` **y** `time.UnixMilli(d.Time.Created)` | OccurredAtEsCompleted | = · `1afab4b1` |
| m11 | `if false && (errData != nil \|\| …)` | Incoherente/ ×7 | = · `55164d41` |
| m12 | tras `modelo := …`, `if modelo == "" { return nil, IncoherenteOpenCode }` | SinModelo/ausente, /vacio, /no_textual | = · `fc955985` |
| m13 | `sesion = ctx.SessionID` *(sin derivar)* | Referencias/sesion · CentinelasFueraDelEvento | = · `1aaa0dc9` |
| m14 | `Derivar(ctx.Directory, …)` → `Derivar(ctx.Path, …)` | Referencias/proyecto | = · `875b52d2` |

**Re-derivación por hoja**: (15) no lleva `reasoning` ni `cache` *(m5–m7 no la tocan)*; (18/coste_cero) usa `evento.json`, sin `$.cost` *(m9 no la
toca)*; (21/sin_sesion) tiene sesión vacía *(m13 da `""` y m14 queda dentro de la guarda)*; (22) compara el JSON entero, pero m14 y m9 sólo meten
hashes y un `0.25` sin centinela; todos los fixtures de evento llevan `model.id` *(m12)*, y `time.created` numérico *(m10)*.

Las diez, **= declarado**, y restauradas a `1b00c194…` antes de la siguiente; directa e inversa comprobadas únicas, y cada mutante compilado, antes de mutar.

## B4 · Base y frontera *(`internal/ingest/opencode_base.go`, `internal/state/state.go`)*

### T022–T023 *(2026-10-12, sobre `c63cf85`)*
- **T022**: `FileState.Frontera` *(`omitempty`)*, `Frontera` y `EnCurso` → `state.go` `7622236a…`; esqueleto → **765 / 0 / 0**; `internal/state`: 10 «--- PASS», sin tocar sus tests.
- **T023**: `generar.sh` `d773524d…` *(Python `sqlite3`, SQLite 3.45.1; conexión retenida para los `-wal` de `frontera/`)*: 28 ficheros, 668 505 B. md5:
  `frontera/paso1–2` `c564552b` + `-wal 3535469c` · `paso3` `c564552b` + `-wal 07e01470` · `inverso/paso1–3` `1ea02836 4d32c776 271e9131` ·
  `lista_borrada/paso1–2` `7d560547 5a13808c` · `vacuum/antes` `35931d5d` · `despues` `7bca728f` · `borrada_h/paso1–2` `abdbe429 78732391` ·
  `abandono/paso1–3` `5c9a9bbc 0106ea3d 2e74ae4d` · `abandono_1h` `f8436203` · `clases` `53af85e7` · `dos_sesiones` `d5324c9b` · `centinelas` `50893cd6` ·
  `formatos/`: `v1 38041872` · `without_rowid b153e8d0` · `utf16 4e70c51b` · `sin_columna 930dda16` · `corrupta` `6daa9b6b` · `esperado.json` `88c8eb27`.
- **Hallazgo (SC-024)**: con `id text PRIMARY KEY`, el `VACUUM` de SQLite 3.45.1 **conserva** los rowid *(la tabla tiene índice)*; sin índices, los
  renumera. `vacuum/` lleva `session_message` sin clave ni índices, y el guion comprueba que el `VACUUM` renumeró.
- **`.gitignore`**: `!internal/ingest/testdata/opencode/**/*.db`; `git check-ignore`: 0 de 24 bases y 0 de 3 `-wal`. Privacidad con `grep -a`: 0.

### T024–T025
- **Rojos**: 16 tests y 43 «--- FAIL», todos por el esqueleto *(resumen `""`, sin eventos, sin entrada, sin error)*. Nace verde (36/otras): la valida m29.
- **Verde** `7c3a011b…`; lint 0; **809 / 0 / 0**. (24/paginas) cayó con 12 > 10: las `assistant` de `frontera/` desbordaban la página de 1 KB y cada
  `Buscar(25)` leía su desbordamiento. Se corrige el **fixture** *(relleno 500, y el guion lo comprueba)*, no la cota.
- **SC-017** *(informativa, una vez; 1 039 265 792 B, 253 727 páginas, 250 000 filas, 2 500 sesiones)*: 1.ª pasada **29,3 s > 20 s: no cumple**; la
  siguiente, con 10 nuevas, **43,6 ms y 75 páginas (0,030 %): cumple**. Se anota y se sigue *(normativa en C2)*.

### T026 · Censo *(declarado ANTES de mutar; prefijo `TestBaseOpenCode_`)* y T027 · mutaciones *(base `7c3a011b…`; suite entera; edición inversa; md5)*
| Mut. | Edición | Declarado | Cayó · md5 mutado |
|---|---|---|---|
| m15 | `case c.huella(fila) != fr.Huella:` → `… && false:` *(`case false:` no compilaba: `fila` sin usar)* | VacuumRenumera/ninguna_falta, /literal | = · `8443cd75` |
| m16 | `return false, nil // la fila H ya no está` → `return true, …` | BorradaH/salen, /reinicio | = · `ad9ec2b9` |
| m17 | en el bucle de la lista, `if true { c.lista = append(c.lista, e); continue }` | EnCursoSeCompleta · OrdenInverso · FilaDeLaListaBorrada · Abandono/paso2, /paso3 · **Inestable/quieta** | = · `49906409` |
| m18 | antes de `c.clasificar(e.Rowid, fila)`, `if c.abandonada(fila) { …Abandonadas++; continue }` | Abandono/paso2 | = · `05c0d7ae` |
| m19 | `return false && c.ahora.Sub(…) > abandonoOpenCode` | Abandono/paso2, /paso3, **/1h** · Clases/literal | = · `80ec170d` |
| m20 | `desde, c.h, c.huellaH = previo.Offset, …` → `desde, … = math.MinInt64, …` | SegundaPasadaSinCambios/linea, /paginas · EnCursoSeCompleta · OrdenInverso · FilaDeLaListaBorrada · Abandono/paso2, /paso3, **/1h** | = · `a2a79d6b` |
| m21 | antes del `continue // la sesión se borró`, `c.cuenta.Abandonadas++` | FilaDeLaListaBorrada | = · `00f323dc` |
| m22 | `fr != nil {` → `fr == nil { desde = previo.Offset } else {` | EntradaSinCampo · **EstadoJSON/estado_060** | = · `2ff3148c` |
| m23 | `c.cuenta.Reiniciadas++` → comentario | VacuumRenumera/literal · BorradaH/reinicio | = · `e5d8c81d` |
| m24 | `intentosOpenCode = 5` → `1` | Inestable/una | = · `6fa948ec` |
| m25 | antes del `return` de `errInestableOpenCode`, `st.Files[ruta] = l.estado` | Inestable/quieta | = · `d20bbe16` |
| m26 | `if visto && previo.Size == …` → `if false && visto && …` | Formatos/anterior_segunda | = · `ed322233` |
| m27 | `n != "opencode-next.db"` → `true` | `TestListarOpenCode` | = · `becdd4ad` |
| m28 | `return event.Ref(c.sal, textoDeValor(…))` → `return textoDeValor(…)` | HuellaConSal | = · `7868fe1a` |
| m29 | `state.go`: `json:"frontera,omitempty"` → `json:"frontera"` | EstadoJSON/otras · `TestRetencion_SC009_NadaDelProveedorEnDisco/state_json_cuatro_campos` | = · `9835d068` |

**Re-derivación** *(en negrita, las co-caídas que plan §Bloques no lista)*. Sólo m22 toca las entradas sin `frontera`: (28) y (36/estado_060) recorren
desde 23. Los formatos se deciden antes de la frontera *(m22 no toca (34))*. Las de la lista y del recorrido sólo tocan pasadas con frontera:
- **m17**: (24) no cae *(la fila sigue en curso, y lee menos páginas)*, ni (31/1h) *(su fila sigue en curso)*. Sí (33/quieta): su tercera pasada no
  consulta la 25.
- **m19 y m20**: caen en (31/1h): la primera pasada no abandona la de 25 h, y la segunda recorre otra vez desde el principio.
- **m20**: (29) y (30) no caen *(reinician)*, ni (33/quieta) *(sólo exige que salga la 25)*. **m18**: (31/paso3) no cae *(la lista ya está vacía)*.
- **m15**: sólo (29) tiene una fila distinta en H. **m28**: H y la lista se comparan con la misma función; sólo (38) mira la huella.
- **m29**: el `"frontera": null` no lleva identificadores *(los tests de 008 y 009 sobre `state.json` no caen)*; (36/bases) busca `"frontera":{`.

Las quince, **= declarado**, y restauradas a su md5 antes de la siguiente; directa e inversa únicas y cada mutante compilado, antes de mutar.

## B4 remate · SC-017 en tope y `VACUUM` puede renumerar

**Decisiones del dueño (2026-10-12)**: `tokens` nulo es incoherente *(B3)*; las desviaciones de B4, aceptadas; y la enmienda E-2 de la spec *(`VACUUM` **puede** renumerar)*.

### SC-017 · Perfil antes de tocar *(base `7c3a011b…` y `1b00c194…`; la base de 1 GB de `TestSC017`, en `t.TempDir()`)*
1.ª pasada **23,6 s** *(en B4, 29,3 s)*; 16,8 s de CPU en la pasada. Las 10 de más tiempo acumulado:
| Función | Acum. | % |
|---|---|---|
| `testing.tRunner` | 16,81 s | 98,0 |
| `ingest.LeerBaseOpenCode` · `TestSC017_CosteDeLaPasada` · `leerBaseOpenCode` | 16,80 s | 97,9 |
| `ingest.(*clasificadorOpenCode).emitir` | 10,24 s | 59,7 |
| `ingest.LlamadaOpenCode` | 8,63 s | 50,3 |
| `encoding/json.Unmarshal` | 8,53 s | 49,7 |
| `sqlitero.(*Base).recorrer` · `recorrer.func1` · `(*Tabla).RecorrerDesde` | 6,56 s | 38,2 |

**Dónde se va**: el JSON, **8,5 s** *(51 %)*, con `checkValid` 5,3 s: **la hipótesis se confirma**, `data` se analiza dos veces por fila
*(`clasificar` 3,4 s y `LlamadaOpenCode` 5,0 s)*. Además, `Resolutor.Derivar` con receptor nil, **2,8 s** *(250 000 `os.Lstat`: el test no
lleva resolutor; `generate()` crea uno por pasada)*, y el `growslice` de `emitir`, 1,25 s. `sqlitero` propio, ≈ 2,6 s *(`carga` 1,9 s, `ReadAt` 1,2 s)*:
**no es el cuello**, y `lector.go` no se toca. Memoria: 3,7 GB asignados, 1,5 GB en `carga`, y cada llamada retenía su fila entera hasta `emitir`.

**Cambios, sin cambiar la conducta**:
- `clasificar` analiza `data` **una vez**, con `datosOpenCode` *(da error en los mismos casos que `tiempoOpenCode`: sus campos de más son
  `json.RawMessage`)*, y pasa lo leído a `LlamadaOpenCode` por `FilaOpenCode.leida` *(no exportado; nil → analiza `Data`, como en B3)*. La llamada
  guarda una copia del `id`, no la fila.
- `emitir`: sin resolutor, uno para la base *(la caché optimiza, no decide: P-004)*; `evs` con su capacidad.

### SC-017 · Después *(`814d5a7e…` y `4e2c4c19…`; **809 / 0 / 0**)*
Perfil: 1.ª pasada **5,4 s**, 5,3 s de CPU *(antes 16,8)*; 2,8 GB asignados *(antes 3,7)*. Las 10 de más tiempo acumulado:
`tRunner` 5,32 s · `LeerBaseOpenCode`, `TestSC017`, `leerBaseOpenCode` 5,31 s · `recorrer`, `recorrer.func1`, `RecorrerDesde` 4,33 s ·
`json.Unmarshal` 3,74 s *(un análisis por fila)* · `RecorrerDesde.func3` 3,46 s · `leerBaseOpenCode.func2` 3,38 s. `Lstat` ya no aparece.

**Tres medidas**, cada una con su ventana y su base nueva de 1 039 265 792 B:
| Medida | 1.ª pasada *(≤ 20 s)* | Siguiente *(≤ 50 ms)* | Páginas *(≤ 1 %)* |
|---|---|---|---|
| 1 | 5,205 s | 3,309 ms | 75 *(0,030 %)* |
| 2 | 5,114 s | 2,715 ms | 75 *(0,030 %)* |
| 3 | 5,149 s | 9,281 ms | 75 *(0,030 %)* |

**Cumple**, con margen de ≈ 3,9× en la primera y ≥ 5× en la siguiente *(W1, en Windows, será más lenta)*.

### Mutaciones tocadas *(declarado ANTES de mutar)*
Ninguna línea de anclaje de m5–m29 cambia ni se mueve de función. Pero `datosOpenCode` y `errData` ahora también vienen de `clasificar`: se repiten
**m9, m10 y m11** con el protocolo completo *(bases `814d5a7e…` y `4e2c4c19…`)*, y se re-derivan con los tests de B4, que en B3 no existían:
- **m9** *(`Cost float64`)* y **m10** *(`Created int64`)*: todas las bases de B4 llevan `$.cost` y `$.time.created` numéricos *(o sin `created`)*, y
  ningún test de B4 mira `cost_usd` ni `occurred_at`: **mismo declarado** que en B3.
- **m11** *(sin incoherentes)*: además de las 7 de B3, **(32) Clases/literal** *(`o-clases-nojson` y `o-clases-negativa` salen como eventos:
  incoherentes 0)*. (32/identidad) no cae: la identidad se cumple igual.

Las demás siguen valiendo, con su md5 de base nuevo: m5–m8 y m12–m14 sobre `opencode.go` `814d5a7e…`; m15–m28 sobre `opencode_base.go` `4e2c4c19…`;
m29, `state.go` sin cambios *(`7622236a…`)*.

| Mut. | Edición *(la de B3)* | Declarado | Cayó · md5 mutado |
|---|---|---|---|
| m9 | `Cost float64` en `datosOpenCode` **y** `CostUSD: d.Cost` | `TestLlamadaOpenCode_SinCoste/cost` | = · `eb641a34` |
| m10 | `Created int64` en `Time` **y** `time.UnixMilli(d.Time.Created)` | `TestLlamadaOpenCode_OccurredAtEsCompleted` | = · `59320c5f` |
| m11 | `if false && (errData != nil \|\| …)` | `TestLlamadaOpenCode_Incoherente/` ×7 · **`TestBaseOpenCode_Clases/literal`** | = · `00ce22d1` |

Las tres, **= declarado**, y restauradas a `814d5a7e…`; directa e inversa únicas y cada mutante compilado, antes de mutar.

## B5 · Ubicación *(`internal/config/opencode.go`, `internal/testutil/sandbox.go`)*

- **T029–T031** *(2026-10-12, sobre `936589e`)*: esqueleto `2f1d1d8c…` → **809 / 0 / 0**. Rojos (39) ×8 y (40) ×2, por el esqueleto *(`("", "", nil)`; las
  dos variables siguen definidas)*. Verde `a8593750…`, `sandbox.go` `98c03742…`; lint 0; **821 / 0 / 0**. `codex.go` y `gemini.go`: sólo el comentario.
- **T032**: `sandbox_test.go`, verde y sin tocar; `os.Getenv` de producción → **4** *(`codex.go:19`, `gemini.go:20`, `opencode.go:21`, `:25`)*.

**T033 · Censo** *(ANTES de mutar; (39) = `TestOpenCodeDatos_Ubicacion/`, (40) = `TestSandbox_VaciaLasVariablesDeOpenCode/`)*. Ningún otro código lee
las dos variables *(grep)*, y `os.UserConfigDir` mira `XDG_CONFIG_HOME`, no `XDG_DATA_HOME`: ninguna co-caída fuera de (39) y (40).
| Mut. | Edición | Declarado | Cayó · md5 mutado |
|---|---|---|---|
| m30 | `Getenv("XDG_DATA_HOME"); datos != ""` → `LookupEnv(…); definida` *(vacía como raíz; ausente sigue igual)* | (39/xdg_vacia) | = · `d4aedafe` |
| m31 | `IsAbs(db)` → `IsAbs(db) \|\| db != "" && db != ":memory:"` | (39/relativa) | = · `eff42a83` |
| m32 | `IsAbs(db)` → `IsAbs(db) \|\| db == ":memory:"` | (39/memory) | = · `8cc512a5` |
| m33 | `IsAbs(db)` → `false && IsAbs(db)` *(ninguna otra hoja es absoluta)* | (39/absoluta) | = · `b43942f6` |
| m34 | sin las dos líneas del sandbox *(y su comentario)* | (40/XDG_DATA_HOME), (40/OPENCODE_DB). Los tests de proceso de B1–B4 **no caen**: el binario aún no llama a `OpenCodeDatos` *(B6)* | = · `f0ebaae7` |

Las cinco, **= declarado**, y restauradas a `a8593750…` y `98c03742…`; directa e inversa únicas y cada mutante compilado, antes de mutar.

## B6 · Integración *(`cmd/permea/main.go`, `cmd/permea/opencode.go`)*

- **T035** *(2026-10-12, sobre `b0d8535`)*. **(1)** `git diff 4c6cafc -- cmd/permea/main.go` vacío. Binario de `4c6cafc` *(`git archive` + `go build`)*,
  en `env -i` con `HOME`, `USERPROFILE` y `XDG_CONFIG_HOME` temporales, Claude Code, Codex y Gemini *(los fixtures de 008 y 009)* y nada de OpenCode:
  `rc=0`, stdout vacío. Con `<DATOS>`, `testdata/opencode/referencia-run.stderr`, **`83ede4146b4a9131089baa11a896b935`**: 8 líneas, sin rutas
  *(privacidad 0, también `/tmp`)*. **(2)** Forzado A, con un test temporal ya borrado: `Load` → `<nil>`; `Append` → `<nil>`, 323 B; `Save` →
  `permission denied`; uid 1000. **(3)** Esqueleto → **821 / 0 / 0**. Después, el campo pasa de `opencodeDir` a `opencodeDatos`, el nombre del plan.
- **Fixtures**: `testdata/opencode/generar.sh` *(comprueba 0 filas en curso: PARA 11)*: `sana.db 7014d754` · `otra.db 210aa12b` · `centinelas.db
  46ee35a1` · `corrupta.db f6f5be9a`. `.gitignore`: `!cmd/permea/testdata/opencode/**/*.db`; `git check-ignore` → 0. (48) usa I de `internal/sqlitero`,
  sin copiarla a `cmd`: sus filas sin `completed` son de 1970, «abandonadas» desde la primera pasada; su resultado no depende del reloj.
- **T036**: rojos (42)–(46), 5 «--- FAIL» por el esqueleto *(0 eventos `opencode`, sin línea)*; (41) ×2 nace verde. **T037**: verde, `main.go`
  `f4c07e53…`, `opencode.go` `a47bebc2…`; lint 0; **829 / 0 / 0**. **T038**: los 237 de `cmd/permea`, verdes y sin tocar *(`git diff` → 0)*.
  **T039**: (47)–(50) nacen verdes → **835 / 0 / 0**, `opencode_test.go` `7efb8c04…`.

**T040 · Censo** *(ANTES de mutar)*. m35: los tests que comparan el stderr de `--run` entero o cuentan sus líneas son, además de (41), sólo los
de la lista *(grep de `"--run"`: los demás buscan una subcadena o el código)*. Las bases de (50) se leen en orden léxico: `opencode-roto.db` antes
que `opencode.db`, así que con m40 nada de OpenCode se encola en la 1.ª pasada, y (50/se_relee) ve 3.
| Mut. | Edición | Declarado | Cayó · md5 mutado |
|---|---|---|---|
| m35 | `a.opencode = nil` → `= ingest.NuevaPasadaOpenCode()` | (41) ×2 · (49) · `TestCodexRun_SinRaizEsLaSalidaDeLa040/{codex_home_vacia, raiz_inexistente}` · `TestCodexRun_LineaDeResumen` · `TestGeminiRun_SinRaizEsLaSalidaDeLa050/{gemini_cli_home_vacia, sin_tmp}` · `TestGeminiRun_LineaDeResumen` | = · `29752fdf` |
| m36 | un `st.Save` antes de OpenCode | (43) | = · `760673dc` |
| m37 | `tick`: `a.opencode != nil && a.opencode.HayNovedades()` → `a.opencode != nil` | (45) | = · `ceaec01a` |
| m38 | `runOnce`: la línea `opencode:` antes que la de Gemini | (44) | = · `8ede8340` |
| m39 | `setup()`: si la carpeta no existe, `opencodeDatos = ""` | (49) | = · `e8c91fd0` |
| m40 | el error de una base: `continue` → `return total, err` | (50/pasada) | = · `c27e90e2` |
| m41 | el error de una base escribe su entrada *(`Frontera` vacía)* | (50/se_relee) | = · `df624c23` |
| m42 | `ev.ProjectRef = ruta` antes de encolar | (47) | = · `af2fe6b0` |
| M-B6a | `generarOpenCode` escribe `m-b6a.tmp` en la carpeta de OpenCode | (48) | = · `8c3c0ac4` |

**T040** *(13-10; base `f4c07e53…` · `a47bebc2…`, **835 / 0 / 0**; md5 del fichero mutado)*. m38 quedó puesta por una sesión cortada el 13-10; revertida
por edición inversa; T040 rehecha entera. Las nueve, **= declarado**, restauradas a su md5; directa e inversa únicas y cada mutante compilado, antes de mutar.

**T041 · SC-007** *(informativo; 2026-10-13; `git archive` de `4c6cafc` y el árbol de B6; `CGO_ENABLED=0 -trimpath -ldflags "-s -w -X main.version=0.7.0"`; go 1.22.2)*:
| Plataforma | `4c6cafc` | B6 | Δ fichero | Δ contenido *(E-3)* |
|---|---:|---:|---:|---:|
| linux/arm64 *(fichero · Σ `FileSiz` de los `LOAD`)* | 5 308 568 · 5 252 492 | 5 439 640 · 5 331 036 | 131 072 *(sin juzgar)* | **78 544** |
| linux/amd64 · darwin/amd64 · darwin/arm64 · windows/amd64 | 5 554 328 · 5 710 368 · 5 486 658 · 5 747 712 | 5 644 440 · 5 802 512 · 5 554 722 · 5 840 384 | 90 112 · 92 144 · 68 064 · 92 672 | — |

Con E-3 *(el dueño, 12-10)*, las cinco ≤ 102 400 B. `readelf -lW` *(y `debug/elf`, igual)*: R E +38 416, R +39 968, RW +160; dos cruzan 64 KB *(2 × 65 536)*.
