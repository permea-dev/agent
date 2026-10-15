# 012 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007–011, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. Techo: **300** líneas
*(tasks §Presupuesto)*.

**Cómo se corre cada suite** *(tasks §Una suite a la vez)*: antes, **en una orden sola y en segundo plano**, `bash ~/bin/libre.sh` *(md5 `ab647c01…`)*,
hasta «inicio:» y `libre=0`; la suite va en la orden siguiente. Se anotan `/proc/loadavg` y la duración. **Ninguna orden que espere a `libre.sh` lleva en su
línea de órdenes el nombre de una suite** *(011 registro §B0, incidencia)*.

- **§B0** *(FASE 0 y registro: línea base 953 / 0 / 0, tamaños de la 0.8.0 por `git archive`)*: archivado íntegro el 2026-10-15 en [`registro-archivo.md`](./registro-archivo.md).
- **§B1** *(raíz y ficheros: 967 / 0 / 0, m1–m8 = declarado, `os.Getenv` → 6)*: archivado íntegro el 2026-10-15 en [`registro-archivo.md`](./registro-archivo.md).
- **§B2** *(una línea e identidad: 1004 / 0 / 0, m9–m24 = declarado, Δ de tamaño 0)*: archivado íntegro el 2026-10-15 en [`registro-archivo.md`](./registro-archivo.md).

## B3 · Fichero, cabeza e índice *(`kimi_contexto.go`, `state.go`)*

- **T019** *(2026-10-15, sobre `b3a2989`)*: `state.FileState` gana `Cabeza string \`json:"cabeza,omitempty"\`` tras `Frontera`, **sin tocar** `Recorrer`, `Load` ni `Save`;
  `ContextoKimi` gana `Indice`; stubs de `PasadaKimi` *(`Resumen` → `""`)*, `LeerFicheroKimi` *(no emite)*, `LeerIndiceKimi` *(mapa vacío)* y `huellaCabezaKimi` *(`""`)*.
- **T020**: `generar.sh` gana `ficheros/`: 21 ficheros, con `indice/session_index.jsonl` *(las dos barras de Windows; una sesión con dos líneas)* y `esperado.json`, que
  calcula **un clasificador propio del guion** *(sin código del lector)*. `raiz/` y `lineas/` no cambian al regenerar.
- **T021 · Rojos, la razón escrita ANTES de correrlos**: con los stubs, se espera **1010 / 32 / 0**. Caen por el stub *(resumen `""`, sin eventos, huella `""`, sin error)*:
  (14); (15) ×3 y su padre; (16); (17) ×2 y su padre; (18) ×4 y su padre; (19/mtime) y su padre; (20) ×2 y su padre; (21) ×4 y su padre *(«sin eventos»)*; (22); (23);
  (24); (25); (26); (28) ×2 y su padre. **Nacen verdes**, y se declaran así: **(27)** *(el stub no escribe; m39)*, **(29)** ×3 *(el campo ya está desde T019; m40)* y la hoja
  **(19/igual)** *(el stub no abre nada; m33)*: 1004 + 6 = 1010 PASS.
- **T021 · Rojos, corridos** *(libre 0.27; 14:07:47 → :55, 7,5 s)*: **1010 / 32 / 0**, rc 1: **los 32 declarados, con la razón del stub**, y los 6 nacidos verdes en PASS.
- **T022 · Verde**: `kimi_contexto.go` `4645bc4d…` *(atajo tamaño + `mtime`; cabeza; prepaso; N = registros antes del offset + vistos; heredado = antes de la última `forked`;
  `Recorrer`; la cabeza tras `Recorrer`; el índice, gana la última línea; barras de Windows)*; `state.go` `698968ab…`. **Suite** *(libre 0.57; 14:09:09 → :17, 7,3 s)*:
  **1042 / 0 / 0** *(1010 + 32)*.
- **T023 · SC-022, informativo tras B3** *(como en B2)*: linux/amd64, linux/arm64 *(fichero)* y darwin/amd64 Δ **0**; darwin/arm64 **+16**; windows/amd64 **+512**; arm64 Σ `LOAD`
  **+416**; **holgura de `__TEXT` en darwin/amd64: 1 004 → 683 B** *(arm64: 16 102 → 15 781)*. **De dónde sale, medido**: el campo `Cabeza` de `state.FileState`, que sí está
  enlazado *(metadato de tipo y etiqueta JSON)*. Compilado sin él *(y sin sus dos usos en el lector, aún muerto)*, las cifras vuelven **exactamente** a las de la 0.8.0.
- **R-2 · El prepaso, informativo** *(copia del módulo fuera del repo, con un test desechable; tras `libre.sh`)*: un `wire.jsonl` sintético de **10 487 160 B** *(5 010 líneas de
  relleno de ~2 KB y 250 registros)*: **1.ª pasada 126 ms**; **2.ª, tras añadir un registro *(prepaso entero + la cola)*, 7,1 ms**.
- **T024 · Censo de m25–m40, declarado ANTES de mutar** *(md5 de partida: `kimi_contexto.go` `4645bc4d…`, `state.go` `698968ab…`; **las 16 compilan**, comprobado aplicando,
  `go vet` y revirtiendo)*. Ajustes al plan, razonados sobre la testdata: **m25** no tira (15/sin_forked) ni (25), que no tienen `forked`; **m28 y m29** sólo se notan con
  offset > 0 *(N = registros antes del offset + vistos)*: (16) y (17/dos_pasadas); **m38** tira también (17) ×2; **m40** tira también un test existente,
  `TestRetencion_SC009_NadaDelProveedorEnDisco/state_json_cuatro_campos` *(exige 4 campos)*.

  | Mut. | Cambio | Debe caer *(hojas; + padres)* | FAIL |
  |---|---|---|---:|
  | m25 | heredado nunca *(`false &&`)* | (15/un_forked), (15/dos_forked), (16), (24) | 5 |
  | m26 | la **primera** `forked` | (15/dos_forked) | 2 |
  | m27 | N sin contar los heredados | (15/un_forked), (16) | 3 |
  | m28 | el prepaso cuenta **todas** las líneas | (16), (17/dos_pasadas) | 3 |
  | m29 | `previos := 0 * …` | (16), (17/dos_pasadas) | 3 |
  | m30 | sin comparar la cabeza | (18/cabeza_cambiada), (18/nuevo_tras_cambio) | 3 |
  | m31 | cabeza sin sal *(`event.Ref("", …)`)* | (28/v1), (28/v2) | 3 |
  | m32 | `cabeza` vacía tolerada | (18/sin_cabeza) | 2 |
  | m33 | sin el atajo | (14), (19/igual) | 3 |
  | m34 | el atajo sin `mtime` | (19/mtime) | 2 |
  | m35 | sin normalizar las barras | (21/barras) | 2 |
  | m36 | la **primera** línea del índice | (21/ultima) | 2 |
  | m37 | sin dedupe | (23), (24) | 2 |
  | m38 | la corrupta corta *(`return err`)* | (26), (17/corrupta_cuenta), (17/dos_pasadas) | 4 |
  | m39 | el lector crea `<ruta>.leido` | (27) | 1 |
  | m40 | `json:"cabeza"` sin `omitempty` | (29/entrada_0_8_0), (29/ida_y_vuelta), `TestRetencion_SC009_…/state_json_cuatro_campos` | 5 |
- **T024 · Mutaciones, corridas** *(cada suite tras `libre.sh` solo con `libre=0`, carga de inicio 0.22–1.23; 6,8–7,1 s; edición inversa y md5 comprobados)*: **16 de 16 =
  declarado**, sin incidencias. *(md5 puesta → PASS/FAIL)*: m25 `46f9ce33` 1037/5 · m26 `b7b4a0ed` 1040/2 · m27 `a1f01de3` 1039/3 · m28 `763686b4` 1039/3 · m29 `e28ede0a` 1039/3 ·
  m30 `e21bef2b` 1039/3 · m31 `b27e7ed8` 1039/3 · m32 `adb517d2` 1040/2 · m33 `85fae583` 1039/3 · m34 `20db7659` 1040/2 · m35 `17ecf37e` 1040/2 · m36 `82e81649` 1040/2 ·
  m37 `9a6c25e9` 1040/2 · m38 `ef15d54a` 1038/4 · m39 `aea497b8` 1041/1 · m40 `bd1712e7` 1037/5; cada una cayó exactamente en las hojas de la tabla, y tras cada una
  `kimi_contexto.go` volvió a `4645bc4d` y `state.go` a `698968ab`.
- **T025 · Corte** *(antes del commit del dueño)*: **suite** *(libre 0.52; 14:23:14 → :21, 6,6 s)* **1042 / 0 / 0**, rc 0; **vet** limpio y **lint 0 issues** *(14:23:49 → :52, 2,9 s)*;
  gofmt vacío; frontera de FR-027 0 B; ningún `_test.go` de `e9e5449` modificado *(`TestRetencion_…` sólo cayó con m40, sin tocarlo)*; `state.go`, +3 líneas; sin `require` ni
  `go.sum`; `os.Getenv` → 6; **ningún fichero de `testdata/kimi/` ignorado por git**; `PRIV` → 0 en el diff, en los 57 ficheros de `testdata/kimi/` y en los nuevos.

## B4 · Integración

- **T026 · Fase 0** *(sobre `ffc18c1`)*, en este orden:
  1. `git diff e9e5449 -- cmd/permea/` **vacío**. La referencia de SC-015, con **el binario de `e9e5449`**: `git archive` en un `mktemp -d`, un test temporal que sólo
     vivió allí *(libre 1.63)*, y el escenario de (30) *(Claude Code, Codex, Gemini, OpenCode y Copilot CLI en `testutil.Sandbox`, sin Kimi)*: rc 0, stdout vacío; con
     `<DATOS>`, **`cmd/permea/testdata/kimi/referencia-run.stderr` `44a96b93…`**, 9 líneas, sin rutas.
  2. Forzado A, en el mismo temporal *(libre 1.13)*: `Load` → `<nil>`; `Append` → `<nil>` *(316 B)*; `Save` → *permission denied*; uid 1000. **Incidencia menor**: el `sed`
     que tapaba la ruta falló, y la prueba *(un test de 0,4 s, no una suite)* se repitió sin otro `libre.sh`. Temporal borrado.
  3. Esqueleto: `cmd/permea/kimi.go` *(`kimiActivo`, `generarKimi` sin emitir)*; en `main.go`, los campos `kimiRaiz` y `kimi`, `setup()` y el bloque de `generate()` tras Copilot.
- **T027 · Rojos, la razón escrita ANTES de correrlos**: se espera **1045 / 7 / 0**. Caen por el esqueleto *(«0 eventos kimi_code», sin línea `kimi_code:`, «11 eventos
  encolados», «el primer kimi_code en -1»)*: (31), (32), (33/linea), (33/total), (33/orden_cola) y su padre, y (34). **(30) ×2 nace verde** *(m41)*: 1042 + 3 = 1045.
- **T027 · Rojos, corridos** *(libre 0.47; 18:11:36 → :44)*: **1045 / 7 / 0**, rc 1: los 7 declarados, con la razón del esqueleto.
- **T028 · Verde**: `kimi.go` `5509f984…` *(el índice primero, P-13; errores por fichero con `sinRuta`)* y `main.go` `43e0db25…` *(la línea en `runOnce` tras Copilot, y en
  `tick` con `HayNovedades`)*. **1052 / 0 / 0** *(libre 0.46; 18:12:33 → :41)*. **T029**: `git diff e9e5449 --name-only --diff-filter=M -- '*_test.go'` → vacío.
- **T030 · Nacen verdes** (35)–(39), escritos tras el verde: **1061 / 0 / 0** *(libre 0.34; 18:14:47 → :54)*; `kimi_test.go` `a682d7e9…`. (36) pone registros en
  TODOS los `wire.jsonl*` de `raiz/`, señuelos incluidos: si se leyera uno que no toca, el resumen pasaría de 7.
- **T031 · Censo, ANTES de mutar** *(md5 de partida `main.go` `43e0db25`, `kimi.go` `5509f984`; **las 10 compilan**, comprobado aplicando, `vet` y revirtiendo)*.
  **«E»** *(grep de `len(lineas)` y `got != referencia` en `cmd/permea`; los demás `--run` buscan una subcadena, y `tick` sólo escribe con novedades)*:
  `TestCodexRun_SinRaizEsLaSalidaDeLa040/{codex_home_vacia,raiz_inexistente}`, `TestCodexRun_LineaDeResumen`, `TestGeminiRun_SinRaizEsLaSalidaDeLa050/{gemini_cli_home_vacia,
  sin_tmp}`, `TestGeminiRun_LineaDeResumen`, `TestOpenCodeRun_SinOpenCodeEsLaSalidaDeLa060/{opencode_db_vacia,xdg_vacia}`, `TestOpenCodeRun_LineaDeResumen`,
  `TestCopilotRun_SinCopilotEsLaSalidaDeLa070/{copilot_home_vacia,sin_raiz}` y `TestCopilotRun_LineaDeResumen/linea` *(no `orden_cola`)*: 17 líneas FAIL con sus padres.
  **Ajustes al plan**: **m42** → además (34), (35) y (38/se_relee) *(con el `Save` antes, el offset de Kimi no se guarda: la lección de 011 m31)*; **m46** → además
  (38/se_relee) *(la pasada abortada no guarda nada, y la siguiente relee el sano)*; **m45**, formulada para compilar: `setup()` deja la carpeta vacía si no hay `sessions/`, y
  `kimiActivo` conserva el `os.Stat` sin usarlo *(`_, _ = info, err; return true`)*: sin él, `filepath` quedaría sin uso en `kimi.go`.

  | Mut. | Edición | Debe caer *(hojas; + padres)* | FAIL |
  |---|---|---|---:|
  | m41 | `a.kimi = nil` → `= ingest.NuevaPasadaKimi()` | (30) ×2, (37) y «E» | 21 |
  | m42 | el bloque de Kimi, después del `st.Save` | (32), (34), (35), (38/se_relee) | 5 |
  | m43 | `tick`: sin `&& a.kimi.HayNovedades()` | (34) | 1 |
  | m44 | el bloque de Kimi, antes que el de Copilot | (33/orden_cola) | 2 |
  | m45 | la existencia, sólo en `setup()` | (37) | 1 |
  | m46 | un fichero que no se puede leer → `return total, err` | (38/pasada), (38/se_relee) | 3 |
  | m47 | un fichero que no se puede leer avanza su offset *(`Offset = Size`, con su `mtime`)* | (38/se_relee) | 2 |
  | m48 | índice ilegible → mapa vacío, y se sigue | (39/pasada) | 2 |
  | m49 | `ev.ProjectRef = ruta` antes de encolar | (35) | 1 |
  | M-B4a | `generarKimi` crea `<carpeta>/.lector` | (36) | 1 |
- **T031 · Mutaciones, corridas** *(cada suite tras `libre.sh` solo, carga de inicio 0.16–0.85; 7–9 s; edición inversa y md5 comprobados)*: **10 de 10 = declarado**, sin
  incidencias *(md5 puesta → PASS/FAIL)*: m41 `57419b15` 1040/21 · m42 `19150158` 1056/5 · m43 `1a9e9180` 1060/1 · m44 `9afdf65a` 1059/2 · m45 `e358d952` + `66de8b04` 1060/1 ·
  m46 `572c9de6` 1058/3 · m47 `6b49daee` 1059/2 · m48 `1b5782cf` 1059/2 · m49 `4814ee0d` 1060/1 · M-B4a `d36ffc6a` 1060/1; tras cada una, `main.go` `43e0db25` y `kimi.go` `5509f984`.
- **T032 · SC-022, con el lector ya enlazado** *(`git archive e9e5449` con 0.8.0 frente a la copia del árbol sin `.git` con 0.9.0; el instrumento reproduce las cifras de B0)*:
  linux/amd64 **+28 672**; linux/arm64, fichero 0 y Σ `LOAD` **+26 048**; darwin/amd64 **+29 040** *(`__TEXT` 3 268 608 → 3 289 088, cinco páginas: el salto esperado;
  holgura 1 004 → 2 138 B)*; darwin/arm64 **+16 896** *(una página de 16 KB; holgura 16 102 → 15 934)*; windows/amd64 **+30 208**. **Todas ≤ 102 400**; la mayor, Windows,
  gasta el **29,5 %** del tope. Quedan B5 *(`--scan`)* y B6 *(textos)*.
- **T033 · Corte** *(antes del commit del dueño)*: **suite** *(libre 0.60; 18:27:41 → :48)* **1061 / 0 / 0**, rc 0; vet limpio; **lint 0** *(libre 1.43; 5 s)*; gofmt vacío; frontera
  de FR-027 0 B; ningún `_test.go` de `e9e5449` modificado; sin `require` ni `go.sum`; `os.Getenv` → 6; ninguna testdata ignorada; `PRIV` → 0 en el diff, en los 3 ficheros
  nuevos y en los 58 de los dos `testdata/kimi/`.

## B5 · `--scan`

- **Archivo** *(dueño, encargo de B5; sobre `efa3b1d`)*: §B0, §B1 y §B2 *(177 líneas)* bajan íntegros a [`registro-archivo.md`](./registro-archivo.md), con su rótulo;
  `cmp` del bloque extraído de `HEAD` frente al archivado → **idénticos**. Aquí queda una remisión por bloque *(297 → 123 líneas)*. Plan §Project Structure lo nombra *(249)*.
- **T034 · Rojos, la razón escrita ANTES de correrlos**: se espera **1069 / 4 / 0**. Sin `esSesionKimi`, un `wire.jsonl` sigue el camino de Claude Code: 0 `evento:` y su
  resumen. Caen (40/salida), (40/fork) y su padre, y (43) *(«se esperaba el 1.er evento con event_id=…»)*. **Nacen verdes**, y se declaran así: (40/disco) *(nada en
  disco)*, **(41)** ×5 y su padre *(m50)* y **(42)** *(m52)*: 1061 + 1 + 6 + 1 = 1069.
- **T034 · Rojos, corridos** *(libre 1.45, tras esperar a las suites de la plataforma; 18:55:01 → :10)*: **1069 / 4 / 0**, rc 1: los 4 declarados.
- **T035 · Verde**: `kimi.go` `5502aba4…` *(`esSesionKimi`: 1.ª línea con el tope de 1 MiB, `metadata` con `protocol_version` textual y `created_at` numérico; `dryRunKimi`:
  `LeerFicheroKimi` sobre `state.New()`, que nunca se guarda, sin índice, `imprimirEventos` y el resumen, **sin línea de control**)* y `main.go` `ae048ee1…` *(tras Copilot)*.
  **1073 / 0 / 0** *(libre 0.90; 18:55:53 → :01)*.
- **Añadido tras el verde, nacido verde**: la hoja **(40/subagente)**, porque SC-024 pide también el `agent-0` sintético. 3 eventos, y el 1.º con el vector «subagente» del
  contrato *(`70ee1df9…`)*; lo valida m53. **1074 / 0 / 0** *(libre 1.78; 18:59:28 → :35)*; `kimi_test.go` `dae74079…`.
- **T036 · Censo, ANTES de mutar** *(md5 de partida `main.go` `ae048ee1`, `kimi.go` `5502aba4`, `kimi_contexto.go` `4645bc4d`; **las 5 compilan**, comprobado)*.
  **Ajustes al plan, razonados sobre los tests**:
  - **m50** se formula en `dryRun`: `if kimi || true`. Sólo llega ahí lo que no es OpenCode, Codex, Gemini ni Copilot CLI *(P-12: Kimi va tras ellos)*, así que de (41)
    cae la hoja `claude`, no las 5. Caen, por nombre, los `--scan` existentes que siguen el camino de Claude Code: `TestScan_{UnEventoPorMensaje,
    LineaConCuatroPartidasYEventID, LineaConDesgloseDeLaEscritura, LaSalidaQueCreceValeSuMaximo}`, `TestCodexScan_ClaudeCodeComoLa040`,
    `TestGeminiScan_ClaudeYCodexComoLa050/claude`, `TestOpenCodeScan_LosDemasComoLa060/claude`, `TestOpenCodeScan_FicherosCortosComoLa060/{vacio,diez_bytes}` y
    `TestCopilotScan_LosDemasComoLa070/claude`, con sus padres. No caen ni `TestRetirada_…/--scan con "plain"…` ni `TestCopilotScan_SinCopilotVersionNoEsCopilot`:
    sólo miran el código y la ausencia de `copilot_cli`.
  - **m53 y m54 sólo pueden ir en el código compartido**, porque `dryRunKimi` delega en `LeerFicheroKimi` *(plan §5)*: S y A los saca `sesionYAgenteKimi`, y la frontera
    la aplica el lector. **m54 es la edición de m25** *(mismo md5, `46f9ce33`)*. Así que tiran también (15)–(17), (21) y (25) de B3, y (31)–(35) y (38)–(39) de B4.

  | Mut. | Edición | Debe caer *(hojas; + padres)* | FAIL |
  |---|---|---|---:|
  | m50 | `dryRun`: `if kimi \|\| true` | (41/claude), (42) y los 10 `--scan` de arriba | 17 |
  | m51 | `dryRunKimi` guarda el estado junto al fichero | (40/disco) | 2 |
  | m52 | `esSesionKimi` sin `created_at` | (42) | 1 |
  | m53 | `sesionYAgenteKimi`: S y A cambiados | (15/un_forked), (16), (17) ×2, (21/barras), (21/ultima), (25), (35), (40/subagente), (43) | 14 |
  | m54 | heredado nunca *(`false &&`)* | (15/un_forked), (15/dos_forked), (16), (24); (31), (32), (33/linea), (33/total), (34), (38/se_relee), (39/legible), (40/fork) | 17 |
- **T036 · Mutaciones, corridas** *(cada suite tras `libre.sh` solo, que esperó a las de la plataforma; carga de inicio 0.89–1.51; 8 s; edición inversa y md5
  comprobados)*: **5 de 5 = declarado**, sin incidencias *(md5 puesta → PASS/FAIL)*: m50 `main.go` `261edc70` 1055/17 *(las 2 hojas de `CuatroPartidas` no llegan a
  correr: su padre para en la precondición)* · m51 `kimi.go` `ead40557` 1072/2 · m52 `kimi.go` `0a6db245` 1073/1 · m53 `kimi_contexto.go` `ce773be1` 1060/14 · m54
  `kimi_contexto.go` `46f9ce33` 1057/17. Tras cada una, `main.go` `ae048ee1`, `kimi.go` `5502aba4` y `kimi_contexto.go` `4645bc4d`.
- **SC-022 tras B5, informativo** *(como en B4)*: linux/amd64 **+28 672**; linux/arm64, fichero 0 y Σ `LOAD` **+29 328**; darwin/amd64 **+33 200** *(`__TEXT` otra página:
  holgura 4 047 B)*; darwin/arm64 **+16 960** *(holgura 14 040)*; windows/amd64 **+33 792**. **Todas ≤ 102 400**; la mayor, Windows, el **33,0 %**.
- **T037 · Corte** *(antes del commit del dueño)*: **suite** *(libre 1.30; 19:12:20 → :28)* **1074 / 0 / 0**, rc 0; vet limpio; **lint 0** *(libre 1.13; 6 s)*; gofmt vacío;
  frontera 0 B; ningún `_test.go` de `e9e5449` modificado; sin `require` ni `go.sum`; `os.Getenv` → 6; ninguna testdata ignorada; `PRIV` → 0 en el diff, en el archivo y en
  los 58 ficheros de los dos `testdata/kimi/`.

## B6 · README y CHANGELOG

*(T038–T040)*

## Cierre

*(T042–T051)*
