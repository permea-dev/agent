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
