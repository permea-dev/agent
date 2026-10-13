# 010 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007, 008 y 009, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. Techo: **300**
líneas *(tasks §Presupuesto)*.

## B0 a B4 remate · Archivados *(13-10, íntegros, en [`registro-mutaciones-archivo.md`](./registro-mutaciones-archivo.md), que no cuenta para el techo)*

- **B0** · T001 sobre `0256098`: **671 / 0 / 0**, lint 0, frontera 0; sin mutaciones.
- **B1** · `lector.go` `6acf4700…`: mB1–mB10, M-B1a y M-B1b, **12 de 12 = declarado**.
- **B2** · `opencode_eventid.go` `0a3c19d4…`: m1–m4, **4 de 4 = declarado**.
- **B3** · `opencode.go` `1b00c194…`, `814d5a7e…` desde el remate: m5–m14, **10 de 10 = declarado**.
- **B4** · `opencode_base.go` `7c3a011b…`, `4e2c4c19…` desde el remate; `state.go` `7622236a…`: m15–m29, **15 de 15 = declarado**.
- **B4 remate** · SC-017 cumple *(5,1–5,2 s; 75 páginas)*; m9–m11 repetidas sobre `814d5a7e…` y `4e2c4c19…`: **3 de 3 = declarado**.

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

## Re-censo de las 55 *(dueño, 12-10; 2026-10-13, sobre `e515258`, **835 / 0 / 0**)*

**Sobre el papel, sin mutar**: para cada mutación, sobre el código de hoy, las hojas de la suite actual que caerían, frente a lo declarado. Bases de
hoy = las finales: `lector.go` `6acf4700…` · `opencode_eventid.go` `0a3c19d4…` · `opencode.go` `814d5a7e…` · `opencode_base.go` `4e2c4c19…` · `state.go`
`7622236a…` · `config/opencode.go` `a8593750…` · `sandbox.go` `98c03742…` · `main.go` `f4c07e53…` · `cmd/permea/opencode.go` `a47bebc2…`. Las «igual»
quedan con su evidencia original; las «cambia» se repiten con el protocolo, con el censo nuevo **declarado aquí ANTES de mutar** *(en negrita, lo nuevo)*.
(17) `TestLlamadaOpenCode_Identificador` · (18) `_SinCoste` · (23)–(38) `TestBaseOpenCode_*` *(B4)* · (41)–(50) `cmd/permea` *(B6)*.
| Mut. | Veredicto · razón | Censo nuevo *(sólo «cambia»)* | Cayó · md5 mutado |
|---|---|---|---|
| mB1 | **cambia**: con un `-wal`, `walHuella` queda nil y `Estable()` nunca es cierto: `frontera/` da «cambió 5 veces». (33/sin_ruta) espera ese error; (48) sólo mira md5 y la línea | TestInstantaneas/I, /R · TestEstable/sin_cambio · **(23) · (24)** *(el padre: sus hojas no corren)* · **(25) · (28) · (33/una), (33/quieta) · (36/bases), (36/estado_060)** | = · `4feffcb3` |
| mB2 | igual: ningún `-wal` posterior tiene frames con sales buenas y suma mala; el byte de `tocarWAL` no completa un frame | — | original |
| mB3 | **cambia**: en `frontera/` *(páginas de 1 KB)* el `CREATE` de `session_v2` *(999 B)* desborda la página 1: `Tabla` da `errCorta`. Las demás bases, de 4 KB, no desbordan | TestInstantaneas/I, /R, /S · TestDesbordamiento/8k, /64k · **(23) · (24)** *(el padre)* · **(25) · (28) · (33/una), (33/quieta) · (36/bases), (36/estado_060)** | = · `b9e69007` |
| mB4 | igual: los `-wal` de `frontera/` acaban en commit; I, en (48), sólo se mira por md5 y la línea | — | original |
| mB5 | **cambia**: sin poda, `RecorrerDesde(25)` lee las 25 hojas de `frontera/`: 30 páginas > esquema + 6 | TestRecorrerDesde/paginas · **(24/paginas)** | = · `4c8e7ff8` |
| mB6 | igual: en `frontera/`, H y la fila en curso son el rowid 25, al que lleva el puntero derecho; las demás bases son de una hoja | — | original |
| mB7 | igual: en B4 y B6, `os.Stat` va antes de `Abrir`: ninguna base que falte llega a `Abrir` | — | original |
| mB8 | igual: ninguna base posterior declara una carga mayor que la base | — | original |
| mB9 | **cambia**: (33) depende de que `Estable()` falle | TestEstable/cambia · **(33/una), (33/quieta), (33/sin_ruta)** | = · `81ddd484` |
| mB10 | **cambia**: `formatos/without_rowid` se acepta y su raíz de índice *(0x0a)* da «tipo de página desconocido»: error, no «esquema desconocido» | TestTablas/without_rowid · **(34/without_rowid)** | = · `89700202` |
| M-B1a | igual: `formatos/utf16` sigue en «esquema desconocido» por otro camino: en UTF-16, `"table"` no casa y la tabla «no existe» | — | original |
| M-B1b | **cambia**: el `-shm` aparece en la carpeta de (48) | TestSoloLectura/listado, /md5 · **(48)** | = · `3c99191f` |
| m1 | **cambia**: `tool` sale `"gemini"`: (18/tool), y `cola(…, "opencode")` vacía en B6. (44), (45) y (48) no cuentan por herramienta | VectoresDelContrato/llamada, /otra_llamada · EspacioDeNombresPropio · **(18/tool) · (42) · (43) · (46) · (47) · (49) · (50/pasada), (50/se_relee)** | = · `4c2df440` |
| m2 · m4 | igual: los tests posteriores calculan el `event_id` con la misma `derivarEventIDOpenCode`; sólo los vectores son literales | — | original |
| m3 | **cambia**: el `id` vacío deriva: (17) ve evento o incoherente, y la fila `""` de `clases.db` sale como evento | SinIDNoHayEventID · **(17/ausente), (17/vacio), (17/no_textual), (17/antes_que_incoherente) · (32/literal)** | = · `c50474ea` |
| m5 · m6 · m7 | igual: ningún test de B4 ni de B6 mira partidas | — | original |
| m8 | igual: ningún test posterior mira `cost_available`, y `Append` no valida | — | original |
| m9 · m10 · m11 | igual: re-derivadas en el remate con B4; B6 no mira `cost_usd` ni `occurred_at`, y sus fixtures no tienen incoherentes | — | original *(remate)* |
| m12 | **cambia**: `o-clases-sinmodelo` pasa a incoherente | SinModelo/ausente, /vacio, /no_textual · **(32/literal)** | = · `7b2fdcb3` |
| m13 | **cambia**: el `session_id` en claro: (37) con sesión, y `s-CENTINELA-sesion` en la cola | Referencias/sesion · CentinelasFueraDelEvento · **(37/o-dos-1), (37/o-dos-2) · (47)** | = · `f58482e6` |
| m14 | **cambia**: `dos_sesiones.db` no tiene `path`: `Derivar("")`. En (47), `Derivar` del `path` es un hash | Referencias/proyecto · **(37/o-dos-1), (37/o-dos-2)** | = · `3d984f1f` |
| m15 · m16 | igual: en B6, la segunda pasada ve la misma fila H | — | original |
| m17 · m18 · m21 | igual: los fixtures de B6 no tienen filas en curso *(PARA 11)*; (48) hace una sola pasada | — | original |
| m19 | igual: sólo I, en (48), tiene filas viejas en curso, y (48) sólo mira md5 y la línea | — | original |
| m20 | **cambia**: la segunda pasada de B6 vuelve a emitir las de `sana.db`. (43) y (49) no tienen estado previo de OpenCode | SegundaPasadaSinCambios/linea, /paginas · EnCursoSeCompleta · OrdenInverso · FilaDeLaListaBorrada · Abandono/paso2, /paso3, /1h · **(45) · (46) · (50/se_relee)** | = · `0f28e577` |
| m22 | igual: en B6, una entrada lleva `frontera` o no existe *(Offset 0: desde 0, todas)* | — | original |
| m23 · m24 · m25 | igual: en B6 no hay reinicios ni bases que cambien al leerlas | — | original |
| m26 · m27 | igual: B6 no tiene bases de formato anterior ni `opencode-next.db` | — | original |
| m28 | **cambia**: la huella en claro de `o-CENTINELA-llamada` queda en `state.json` | HuellaConSal · **(47)** | = · `2fb91dd8` |
| m29 | igual: `"frontera": null` no lleva identificadores, y B6 no mira la forma de las entradas | — | original |
| m30 | igual: con `XDG_DATA_HOME` vacía, la ruta es `opencode`, relativa al directorio del test, que no existe: inactivo, como sin mutar | — | original |
| m31 · m32 · m33 | igual: el sandbox deja `OPENCODE_DB` vacía, y ningún test de B6 la pone | — | original |
| m34 | igual: el binario ya llama a `OpenCodeDatos`, pero con el `HOME` del sandbox, y sin `XDG_DATA_HOME` ni `OPENCODE_DB` en el entorno del runner *(`env`)* | — | original |
| m35–m42 · M-B6a | igual: corridas el 13-10 con esta misma suite | — | original |

Las trece «cambia», **= censo nuevo**, cada una con la suite entera y restaurada a su md5 antes de la siguiente; directa e inversa únicas y cada
mutante compilado, antes de mutar. Ninguna línea cambió: todas se escriben como en su bloque. (24) cae sólo como padre con mB1 y mB3 *(818 + 15 y
816 + 17 de 835)*. M-B1b no deja `-shm` en el árbol. Con las 42 «igual», **55 de 55 = declarado** sobre la suite de hoy.

## B7 · `--scan` *(`cmd/permea/main.go`, `cmd/permea/opencode.go`)*

- **T042**: es `e515258`. **T043** *(2026-10-13, sobre `f9f5cca`)*: `contraste.db` `b5674858…`, de `generar.sh` *(las cuatro de B6 salen idénticas)*.
  Referencias de la 0.6.0, del binario de `4c6cafc` *(`git archive` + `go build`; `env -i` y HOME temporal)*: Claude Code y Codex = las de 008 y 009
  *(`cmp`)*; Gemini, `referencia-scan-gemini.stdout` `9d1223c9…` y `.stderr` `2597191a…`; (54), `corto-vacio.jsonl` *(0 B)* y `corto-10.jsonl` `23f6edc2…`:
  código 0, stdout vacío, `referencia-scan-corto-vacio.stderr` `0f668709…` y `-10.stderr` `8465902f…`. Rojos (51/lineas_evento, /resumen, /contraste)
  y (53): la base va por Claude Code, stdout vacío y `skip: invalid character 'S'…` *((51/disco) pasa)*; (52) ×3 y (54) ×2 nacen verdes.
- **T044**: verde, `main.go` `080fac42…`, `opencode.go` `05910a1a…`; el resumen, a stderr como `main.go:680` y `gemini.go:123`; lint 0; **848 / 0 / 0**.
- **Desviación de tamaño** *(tasks §Presupuesto: se declara en el corte)*: **+121** de producción *(`main.go` 8, `opencode.go` 113)* frente a ~80: más del 50 %. B7 pasa a **M**; SC-007 se juzga en C1.

**T045 · Censo** *(ANTES de mutar; (51) `TestOpenCodeScan_EventosResumenYContraste`, (52) `…LosDemasComoLa060`, (53) `…ContrasteFueraDeLosMensajes`, (54) `…FicherosCortosComoLa060`)*.
m44: el grep de `"--scan"` da los diez de plan §Bloques y **uno más**, `TestRetirada_LasExcepcionesDeD0045/--scan con "plain"`; los de `ayuda_test.go`
no llegan a `dryRun`. m43: de los `--scan` de los `_test.go`, sólo (54) pasa un fichero de menos de 16 B *(el menor de los demás, 624 B)*.
**La m43 de antes** *(el bloque de `esBaseSQLite` tras el de Codex; `d2b27a71`, 845 / 0 / 0)* es **equivalente** *(dueño, 13-10)*: el `Scanner` de Codex corta
en el primer `0x0A`, y toda base de OpenCode lo tiene en sus primeras páginas *(las hojas de índice)*; esa línea no es JSON, y se sigue por SQLite.
| Mut. | Edición | Declarado | Cayó · md5 mutado |
|---|---|---|---|
| m43 | `io.ReadFull` corto: `return false, nil` → `return false, err` | (54/vacio), (54/diez_bytes) y su padre | = · `2a3cb7c6` *(código 1, `error: EOF` y `unexpected EOF`)* |
| m44 | `return string(firma) == "SQLite format 3\x00", nil` → `return true, nil` | (52) ×3 · `TestScan_UnEventoPorMensaje` · `TestScan_LineaConCuatroPartidasYEventID` · `TestScan_LineaConDesgloseDeLaEscritura` · `TestScan_LaSalidaQueCreceValeSuMaximo` · `TestCodexScan_EventosYResumen` · `TestCodexScan_ClaudeCodeComoLa040` · `TestCodexScan_FormatoAnterior` · `TestGeminiScan_EventosYResumen` · `TestGeminiScan_ClaudeYCodexComoLa050/{claude, codex}` · `TestRetirada_LasExcepcionesDeD0045/--scan con "plain"`; de los que tienen hojas tras un `Fatalf`, sólo el padre | = · `ce7f83cd` *(818 + 17 + las 10 hojas que no corren)* |
| m45 | `dryRunOpenCode` guarda su estado en `<datos>/state.json` | (51/disco) | = · `d02f7cf7` *(y su padre)* |
| m46 | `signo * (salida + razonamiento)` → `signo * salida` | (53) | = · `f04293ca` *(salida 3, no 7)* |

Las cuatro, **= declarado**, restauradas a `080fac42…` y `05910a1a…`; directa e inversa únicas y cada mutante compilado, antes de mutar.

## Cierre · C1 *(`cmd/permea/main.go`, `cmd/permea/opencode.go`, `cmd/permea/gemini.go`)*

- **C1 anterior** *(13-10, sobre `69bb5bf`)*: **PARA 7**. SC-007, darwin/amd64 **+104 528 B** *(tope 102 400)*; las otras cuatro, dentro.
- **Diagnóstico** *(sin tocar el repo)*: el Δ de darwin/amd64 son 101 054 de contenido + 3 474 de alineación de página *(Mach-O, 4 096)*. B7 aporta
  ~8,6 KB de contenido y ningún paquete nuevo. El bucle `evento:` estaba tres veces *(Codex, Gemini, OpenCode)*; compartido, `__TEXT` suelta una
  página. Probados R1, R2 *(crece)*, R3 *(rompe m43 y m44)* y R1G.
- **R1G** *(dueño, 13-10)*: `imprimirEventos(evs)` en `main.go`, compartido por `dryRunCodex`, `dryRunGemini` *(sin proyecto, vacía antes el
  `ProjectRef`)* y `dryRunOpenCode`. Ningún `_test.go` cambia; frontera 0. `main.go` `29bccccd…`, `opencode.go` `5a56abbe…`, `gemini.go` `892d26e4…`.

**Censo** *(ANTES de mutar)*. **mR1G**: de los `--scan` de los `_test.go` *(grep de `"--scan"` y de `referencia-scan`)*, miran la línea
«N eventos generados» por el helper sólo los que comparan stderr byte a byte con una entrada de Codex o de Gemini:
`TestGeminiScan_ClaudeYCodexComoLa050/codex` y `TestOpenCodeScan_LosDemasComoLa060/{codex, gemini}`, y sus padres. Las referencias de Claude Code y
(54) van por la línea de `dryRun` *(`main.go:632`)*; (29), (37) de Codex y de Gemini, (51) y (53) buscan su línea con `Contains` o miran stdout.
**m43–m46**, otra vez sobre R1G: las mismas ediciones *(cada una casa una vez)* y el mismo censo de §B7.

| Mut. | Edición | Declarado | Cayó · md5 mutado |
|---|---|---|---|
| mR1G | `imprimirEventos`: el recuento, `len(evs)` → `len(evs)-1` | `TestGeminiScan_ClaudeYCodexComoLa050/codex` · `TestOpenCodeScan_LosDemasComoLa060/{codex, gemini}` y los padres | = · `main.go` `857cb666` *(843 + 5; «1 eventos generados», no 2)* |
| m43 | la de §B7 | (54) ×2 y su padre | = · `opencode.go` `46133aa7` *(`error: EOF` y `unexpected EOF`)* |
| m44 | la de §B7 | los de §B7 | = · `78137cc8` *(las mismas 17 hojas FAIL; 821 + 17 + 10 que no corren)* |
| m45 | la de §B7 | (51/disco) | = · `6facab99` *(y su padre)* |
| m46 | la de §B7 | (53) | = · `dd519711` *(salida 3, no 7)* |

Las cinco, **= declarado**, restauradas a `29bccccd…`, `5a56abbe…` y `892d26e4…`; directa e inversa únicas y cada mutante compilado, antes de mutar.

**C1** *(2026-10-13, el árbol con R1G)*: gofmt, vet y lint 0; **848 / 0 / 0**; frontera 0 B; los 671 de los 43 `_test.go` de `4c6cafc`, verdes y sin
cambios *(+177 nuevos)*; `os.Getenv` 4; SC-008: sin `require`, sin `go.sum`, `go list -m all` 1. **SC-007** *(Δ frente a `4c6cafc`)*:
linux/amd64 94 208 · linux/arm64 contenido 84 624 *(fichero 131 072, sin juzgar)* · darwin/amd64 **100 432** *(holgura de `__TEXT` 846 B)* ·
darwin/arm64 84 672 · windows/amd64 98 816. **SC-022**: privacidad 0 en las 5 874 líneas añadidas a `82e0b2d` y con `grep -a` en los 74 testdata
*(44 binarios)*; los tres `generar.sh`, regenerados en `mktemp -d`: 41 ficheros iguales con `cmp`, y los 5 `-wal` difieren sólo en las sales y
sus sumas *(0 páginas distintas)*; lo que no sale de un `generar.sh` es de procedencia registrada *(`filas/` de T016; referencias de `4c6cafc`;
`corto-*`)*; ninguna huella de las copias. Compilan Windows y darwin, amd64 y arm64. `PENDIENTE` 1.

**C2** *(2026-10-13, sobre `ffcdb30`; binario de `git archive`; copias en su sitio, sólo lectura)*: el sha256 de las dos copias, **igual antes y
después** y al registrado *(sin publicarlo: SC-022)*; el listado de sus carpetas, igual *(ni `-wal` ni `-shm`)*.
**SC-001** = en `--scan` y en `--run`: 8 eventos `gemini-3.8-flash`, 59 321 · 692 · 3 999 · 0, y el resumen de la spec. **SC-003**: la 2.ª `--run`, 0
*(sandbox `mktemp -d`, sin enrolar ni `endpoint`, en `unshare -rn`: sin red; nada fuera del sandbox)*. **SC-002**: $0.047386 y 538 · 5 · $0.000174.
**Anthropic**: 24 eventos *(19 + 5)*. **SC-009**: `FuzzBase` 10 min, 65 932 entradas; `FuzzWAL` 10 min, 1 382 338; 0 fallos; el corpus, en la
caché de Go. **SC-017** × 3 *(1 GB)*: 1.ª pasada 5,62 · 5,49 · 6,47 s; siguiente 3,7 · 3,4 · 35,6 ms; 75 páginas *(0,030 %)*. **Equivalencia** con
la sonda B: los 13 testdata *(10 bases)* y las dos copias, 53 tablas, **0 diferencias** en recuentos, sumas y sha256 de `data` *(en las copias,
también = `sqlite3` con `immutable=1`)*; las otras 29 bases de 010, 59 tablas, 0. Sólo cambia la redacción de 5 errores *(mismo paso y causa)*.
En las copias, la equivalencia sólo vale sobre `session_message` y `session_v2`; las demás tablas no se leen *(13-10)*.

**C4 · W1** *(2026-10-13, Windows, OpenCode 2.0.26; `permea-w1.exe` `17412735…` y `sonda-a.exe` `9AEED898…` como contador independiente; sandbox con
su propio perfil de usuario, sin enrolar)*: **Paso 1**, OpenCode cerrado, sobre una copia de `opencode.db` y su `-wal`/`-shm`: `--run` encola 24
eventos = 24 filas completadas del contador, 24 `event_id` distintos; input 142 950 · output + reasoning 2 645 · cache.read 53 510 · cache.write 11
995, iguales; 0 con coste, 0 de otras herramientas, 0 líneas ilegibles. **SC-006**: md5 y listado de la copia, iguales antes y después. La 2.ª
`--run`, 0. **Paso 2**, OpenCode abierto y `--daemon` sobre la base real: primer ciclo 24 = 24; tras dos prompts, 2 eventos nuevos = 2 filas nuevas;
26 = 26, cada uno una vez; 142 956 · 2 655 · 62 238 · 20 754, iguales; **E-1**: 0 ficheros nuevos en la carpeta real *(11 entradas antes y después)*.
**SC-006** y **SC-023**, cumplen en Windows.
