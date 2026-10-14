# 011 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007–010, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. Techo: **300** líneas
*(tasks §Presupuesto; dueño, 14-10)*.

**Cómo se corre cada suite** *(tasks §Una suite a la vez)*: antes, **en una orden aparte y en segundo plano**, `bash ~/bin/libre.sh` *(md5 `ab647c01…`)*,
hasta «inicio:». Se anotan `/proc/loadavg` y la duración. **Ninguna orden que espere a `libre.sh` puede llevar en su línea de órdenes el nombre de una
suite** *(§B0, incidencia)*.

**§B0, §B1 y §B2** *(FASE 0; raíz y ficheros; una línea e identidad: 848 → 861 → 902 pass, m1–m20 = declarado)*: archivados íntegros el 2026-10-14 en [`registro-archivo.md`](./registro-archivo.md), que no cuenta para el techo.

## B3 · Fichero, frontera y tramos *(`LeerFicheroCopilot`, `PasadaCopilot`, en `internal/ingest/copilot_contexto.go`)*

- **Archivo** *(paso 0 del encargo 11/n)*: §B0–§B2, íntegros, en `registro-archivo.md` *(197 líneas, con 5 de cabecera)*. `cmp` del cuerpo contra las líneas
  10–201 de `registro.md` en `e473269`: **idéntico**.
- **T018** *(sobre `e473269`)*: `PasadaCopilot` *(recuentos, `Aperturas`, `Resumen()` con el texto de spec §Textos, `HayNovedades()`)* y el stub de
  `LeerFicheroCopilot`, que no emite.
- **T019**: `generar.sh` `54c0c270…` gana `ficheros/`, 21 ficheros; `sesiones/` y `lineas/` salen idénticos:
  - `dos_llamadas` `ff5fedf2`, `linea_larga` `2da7d939` *(línea de 66 152 B)*, `tramos` `6be61e61`;
  - `cortado/paso1` `60774e39` y `paso2` `c2604ad0` *(64 KB en el prefijo)*;
  - `sin_lf` `69806dcf` y `sin_lf_completo` `9e8607a9`, `truncado/largo` `a29bdeaf` y `corto` `8423a40e`;
  - `repetida` `d0bf7c65`, `clases` `d45f18e7`, `subagente` `9ffb74c3`, `matada` `a7f52241`, `corrupta` `559fd92c`;
  - `sin_start` `8144b6d6`, `sin_cwd` `8af3fdee`;
  - `solo_lectura/` *(con sus dos `.lock` de 0 B)*;
  - y **`esperado.json`** `2a2666a8`, la línea de resumen de cada fichero leído de una vez.

  `sourceSessionId` = `sessionId` en `tramos` y `cortado`; el subagente lleva `s-…09`, `agentId` y `parentToolCallId`.
- **T020 · Rojos** *(libre 1.39; 15:41:54 → :42:03, 8,3 s)*: **903 / 16 / 0**. Caen por el stub, que no emite:
  - (15) ×3 *(«0 eventos», «aperturas: 1.ª 0»)*;
  - (16), (17), (20), (21) ×2, (23), (24) y (25), por el resumen literal de `esperado.json`;
  - (18), (19), (22);
  - y los padres.

  Las 5 hojas de (16) y (24) no llegan a correr: su padre se para antes. **Nace verde y está declarado así**: (26) `TestLeerFicheroCopilot_SoloLectura`
  *(903 − 902)*.
- **T021 · Verde**: `copilot_contexto.go` `c36df6f2…`, calcado de `LeerFicheroCodex`, con prefijo, tramos, repetidas y aviso de línea corrupta.
  - **Suite** *(libre 1.38; 7,4 s)*: **924 / 0 / 0** *(+16 rojos y +5 hojas que ya corren)*. Vet y lint 0.
  - El tipo del contexto se llama `tramoCopilot`, porque `contextoFicheroCopilot` ya era un ayudante del test.
- **(17) reforzado, antes del censo**: comparar dos pasadas con una sola no puede ver m21, porque las dos lecturas se equivocan igual. Se añade que la lectura
  de una vez da el `project_ref` de X y luego el de Y, como fijó el remate *(M5)*. Test `a61eb1d6…`; suite otra vez **924 / 0 / 0** *(libre 1.26)*.

**T023 · Censo** *(ANTES de mutar)*: (15)–(26) = `TestLeerFicheroCopilot_{Pasadas, Tramos, CortadoEnDosPasadas, SinLF, Truncado, Repetida, SinContexto,
Corrupta, SC010Clases, Subagente, Matada, SoloLectura}`. **Un ajuste a plan §Bloques, declarado aquí**: **M-B3a → (24/sesion)**. La hoja compara con el
`session_ref` del `start`, y M-B3a lo cambia por el del `cwd`. Además, m21 tumba (16/dos_proyectos) y (16/valores), no (16/una_sesion).

| Mut. | Edición | Declarado |
|---|---|---|
| m21 | el `session.resume` no cambia el `cwd` | (16/dos_proyectos), (16/valores), (17) |
| m22 | sin `leerPrefijoCopilot` | (17) |
| m23 | en `copilot.go`, la sesión sale de `usage.accounting.sourceSessionId` | (24/sesion), (21/sin_start) |
| m24 | el modelo del último `subagent.configured` sustituye al del registro | (24/modelo) |
| m25 | sin la deduplicación por `event_id` | (20), (23) |
| m26 | una línea corrupta devuelve el error *(corta el fichero)* | (22) |
| m27 | la repetida suma a `Eventos` y no a `Repetidas` | (20), (23) |
| m28 | sin el atajo `Size == Offset` | (15/aperturas) |
| m29 | `LeerFicheroCopilot` crea `<carpeta>/.lector.lock` | (26) |
| M-B3a | en `copilot.go`, `SessionRef: event.Ref(ctx.Salt, ctx.Cwd)` | (16/valores), (16/una_sesion), (24/sesion) |

**T023 · Observado** *(cada una: md5 del verde, `bash ~/bin/libre.sh` sola en segundo plano, suite, `comprobar.py` contra el censo, reversión por
edición inversa y md5 restaurado; antes, una prueba en seco de que las 10 compilan y revierten)*:

| Mut. | md5 mutado | Suite | = declarado |
|---|---|---|:-:|
| m21 | `copilot_contexto.go` `0474c162` | 920 / 4 | ✅ (16/dos_proyectos), (16/valores), (17) |
| m22 | `copilot_contexto.go` `d003edd4` | 923 / 1 | ✅ (17) |
| m23 | `copilot.go` `a2753137` | 920 / 4 | ✅ (24/sesion), (21/sin_start) |
| m24 | `copilot_contexto.go` `a0fb92bf` | 922 / 2 | ✅ (24/modelo) |
| m25 | `copilot_contexto.go` `79f47764` | 922 / 2 | ✅ (20), (23) |
| m26 | `copilot_contexto.go` `1b75721f` | 923 / 1 | ✅ (22) |
| m27 | `copilot_contexto.go` `2028e4c3` | 922 / 2 | ✅ (20), (23) |
| m28 | `copilot_contexto.go` `fca2ee0e` | 922 / 2 | ✅ (15/aperturas) |
| m29 | `copilot_contexto.go` `fcdc48e4` | 923 / 1 | ✅ (26) |
| M-B3a | `copilot.go` `5bfdeb54` | 919 / 5 | ✅ (16/valores), (16/una_sesion), (24/sesion) |

**10 de 10 = declarado.** Libre «inicio» entre 1.01 y 1.93; cada suite, 7 s. Al final, los md5 son los del verde *(`fbd7d7c1`, `c36df6f2`)*. Esta vez,
ninguna espera llevó el nombre de una suite.

**Puertas del corte** *(libre 1.15; 15:59:55 → 16:00:02)*:
- **924 / 0 / 0**; vet y lint 0 *(libre 1.18)*; gofmt vacío;
- frontera de FR-027 vacía; ningún `*_test.go` existente modificado; `go.mod` sin `require`; sin `go.sum`; `os.Getenv` → 5;
- **`PRIV`**: 0 en `git diff 15c4c47`, 0 en los 22 ficheros nuevos sin seguir, y 0 con `grep -raEc` en los 52 de `testdata/copilot/`.

**T022 · SC-007, informativo**: lo mismo que tras B2 *(Δ fichero 0 en las 5; arm64 `LOAD` +256; `__TEXT` darwin/amd64 846 → 718 B)*. El lector sigue sin
enlazar hasta B4.

## B4 · Integración *(`cmd/permea/copilot.go`, `cmd/permea/main.go`)*

- **T025 · Fase 0** *(sobre `57d6b2f`)*, en este orden:
  1. `git diff 15c4c47 -- cmd/permea/main.go` **vacío**. La referencia de SC-015, con **el binario de `15c4c47`**: `git archive` en un `mktemp -d`, un
     test temporal que sólo vivió allí, y el escenario de (27) *(Claude Code, Codex, Gemini y OpenCode en `testutil.Sandbox`, sin Copilot)*. Resultado:
     rc 0, stdout vacío; con `<DATOS>`, **`cmd/permea/testdata/copilot/referencia-run.stderr` `9c5267df…`**, 8 líneas, sin rutas.
  2. Forzado A, en el mismo temporal: `Load` → `<nil>`; `Append` → `<nil>` *(310 B)*; `Save` → *permission denied*; uid 1000.
  3. Esqueleto: `cmd/permea/copilot.go` *(`copilotActivo`, `generarCopilot` sin emitir)*; en `main.go`, los campos `copilotRaiz` y `copilot`, `setup()` y el
     bloque de `generate()` tras OpenCode.
- **Incidencia, la tercera** *(sin efecto en los resultados)*: la espera previa a la referencia **se bloqueó contra sí misma**. Un `heredoc` que escribía el
  guion dejó «go test» en la línea de órdenes de la espera. **Frenó unos 10 min a la sesión de la plataforma.** Se pararon las dos órdenes. Desde entonces,
  los guiones se escriben antes, en otra orden, y la espera sólo mira su fichero.
- **T026 · Rojos** *(libre 1.41; 16:32:44 → :57, 12,5 s)*: **927 / 7 / 0**.
  - **Caen por el esqueleto** *(«0 eventos copilot_cli», «el primer copilot_cli en -1», «se esperaban 3»)*: (28), (29), (30/linea), (30/orden_cola), (31) y (32),
    con su padre.
  - **(27) ×2 nace verde y está declarado** *(927 − 924)*.
- **T027 · Verde**: `copilot.go` `35626e45…` *(`generarCopilot`; errores por fichero con `sinRuta`, D-011-P5)* y `main.go` `1b0efb8b…` *(la línea en
  `runOnce` tras OpenCode, y en `tick` con `HayNovedades`)*. **934 / 0 / 0** *(libre 1.83; 10,0 s)*.
- **T028**: `git diff 15c4c47 --name-only --diff-filter=M -- '*_test.go'` → vacío. Los `_test.go` existentes de `cmd/permea` siguen verdes y sin tocar.
- **T029 · Nacen verdes** (33)–(36), escritos tras el verde: **940 / 0 / 0** *(libre 1.68; 11,8 s)*; `copilot_test.go` `53cf99c7…`. Vet y lint 0; gofmt vacío.
  Los fixtures son los sintéticos de `internal/ingest/testdata/copilot/`: `cmd` no lleva sesiones propias.

**T030 · Censo** *(ANTES de mutar)*: (27)–(36) = `TestCopilot{Run_SinCopilotEsLaSalidaDeLa070, Run_SoloCopilot, Generate_EncolaAntesDeGuardar,
Run_LineaDeResumen, Demonio_SoloConNovedades, Run_SegundaPasadaCero, Run_NadaDelProveedorViaja, Run_SoloLectura, Activacion_EnCadaPasada,
Generate_FicheroIlegibleNoRompe}`.
- **Los tests existentes que comparan el stderr de `--run` entero o cuentan sus líneas** *(grep de `len(lineas)` y `referencia` en `cmd/permea`; los demás
  buscan una subcadena, y los de `tick` usan agentes sin raíz de Copilot)*. Son, POR NOMBRE, los 8 de **«E»**:
  - `TestCodexRun_SinRaizEsLaSalidaDeLa040/{codex_home_vacia,raiz_inexistente}` y `TestCodexRun_LineaDeResumen`;
  - `TestGeminiRun_SinRaizEsLaSalidaDeLa050/{gemini_cli_home_vacia,sin_tmp}` y `TestGeminiRun_LineaDeResumen`;
  - `TestOpenCodeRun_SinOpenCodeEsLaSalidaDeLa060/{opencode_db_vacia,xdg_vacia}` y `TestOpenCodeRun_LineaDeResumen`.
- **Ajustes a plan §Bloques, declarados aquí**:
  - **m31** → además (31), (32) y (36/se_relee): con el estado guardado ANTES, el offset de Copilot no se guarda, y la pasada siguiente lo relee todo;
  - **m34** → además (27) ×2 y «E»: sin comprobar la carpeta, toda `--run` en sandbox da Copilot por activo *(aviso y línea de más)*;
  - **m35** → además (36/se_relee): la pasada abortada no guarda el estado, y la siguiente relee la sesión sana.

| Mut. | Edición | Declarado |
|---|---|---|
| m30 | `a.copilot = nil` → `= ingest.NuevaPasadaCopilot()` | (27) ×2, (35) y «E» |
| m31 | el bloque de Copilot, después del `st.Save` | (29), (31), (32), (36/se_relee) |
| m32 | `tick`: sin `&& a.copilot.HayNovedades()` | (31) |
| m33 | el bloque de Copilot, antes que el de OpenCode | (30/orden_cola) |
| m34 | `copilotActivo`: sólo `a.copilotRaiz != ""`, sin `os.Stat` | (27) ×2, (35) y «E» |
| m35 | un fichero que no se puede leer → `return total, err` | (36/pasada), (36/se_relee) |
| m36 | un fichero que no se puede leer avanza su offset *(`Offset = Size`)* | (36/se_relee) |
| m37 | `ev.ProjectRef = ruta` antes de encolar | (33) |
| M-B4a | `generarCopilot` crea `<raíz>/.lector` | (34) |

**T030 · Observado, hasta el PARA** *(cada una: md5 del verde, `bash ~/bin/libre.sh` sola, suite y `comprobar.py`; antes, una prueba en seco de las 9)*:
- **En la prueba en seco**, la reversión de m32 falló por un ancla duplicada *(`if a.copilot != nil {` está también en `runOnce`)* y dejó m32 puesta en
  `main.go`. **Ninguna suite había corrido.** Se ancló a dos líneas y `main.go` volvió a `1b0efb8b`; m31 y m33 se volvieron a probar sobre el verde.

| Mut. | md5 mutado | Suite | |
|---|---|---|---|
| m30 | `main.go` `f1dcb6ba` | 924 / 16 | ✅ = declarado: (27) ×2, (35) y los 8 de «E» |
| m31 | `main.go` `1d60e136` | 934 / 6 | 🛑 **≠ declarado**. **De más: (33)** `TestCopilotRun_NadaDelProveedorViaja`. De menos: nada |

**🛑 PARA 2** *(16:59)*. **m31 se queda puesta** *(`main.go` `1d60e136`)*. **Motivo de la caída de (33)**: con el `st.Save` antes de Copilot, `state.json` no
llega a tener la entrada del fichero de Copilot, y (33) tiene como precondición que esa clave esté *(«precondición: state.json no tiene la ruta del fichero
como clave»)*. **Es un fallo del censo, no del código**: m31 queda detectada por una hoja más de las que se declararon. **Sin decisión del dueño no se sigue**
*(m32–m37, M-B4a, T031)*.

**Enmienda del censo de m31** *(dueño, 14-10, opción (a); encargo 13/n)*: «Antes: m31 → (29), (31), (32), (36/se_relee)» → **m31 → (29), (31), (32), (33),
(36/se_relee)**. **Por qué**: la precondición de (33) *(que `state.json` tenga como clave la ruta del fichero de Copilot)* depende de «encolar → guardar».
Con el `st.Save` antes de Copilot, esa entrada no llega al disco. **(33) no se toca.** m31 se revierte, se repite contra el censo enmendado y se sigue.

**T030 · Observado tras la enmienda** *(encargo 13/n; m31 revertida por edición inversa → `main.go` `1b0efb8b`, el del verde)*:

| Mut. | md5 mutado | Suite *(libre «inicio»)* | = declarado |
|---|---|---|:-:|
| m31 *(repetida)* | `main.go` `1d60e136` | 934 / 6 *(1.60)* | ✅ con la enmienda: (29), (31), (32), (33), (36/se_relee) |
| m32 | `main.go` `9b0874e6` | 939 / 1 *(1.26)* | ✅ (31) |
| m33 | `main.go` `9b1146cb` | 938 / 2 *(1.02)* | ✅ (30/orden_cola) |
| m34 | `copilot.go` `f0d033ed` | 924 / 16 *(0.77)* | ✅ (27) ×2, (35) y «E» |
| m35 | `copilot.go` `1d62ee5f` | 937 / 3 *(0.61)* | ✅ (36/pasada), (36/se_relee) |
| m36 | `copilot.go` `bcf336f3` | 938 / 2 *(0.90)* | ✅ (36/se_relee) |
| m37 | `copilot.go` `eff393a1` | 939 / 1 *(0.82)* | ✅ (33) |
| M-B4a | `copilot.go` `9d38dd3b` | 939 / 1 *(0.83)* | ✅ (34) |

**B4: 9 de 9 = declarado** *(m30 antes del PARA; m31, con la enmienda)*. Al final, `main.go` `1b0efb8b` y `copilot.go` `35626e45`, los del verde. Cada suite,
7–12 s.

**T031 · SC-007, con el lector ya enlazado** *(`git archive 15c4c47` con `0.7.0` frente al árbol con `0.8.0`; banderas de goreleaser)*:

| Plataforma | 0.7.0 | Tras B4 | Δ fichero | Nota |
|---|---:|---:|---:|---|
| linux/amd64 | 5 648 536 | 5 673 112 | **+24 576** | |
| linux/arm64 | 5 439 640 | 5 439 640 | 0 | Σ `FileSiz` de los `LOAD`: 5 337 116 → 5 358 604, **+21 488** *(normativo, E-3)* |
| darwin/amd64 | 5 810 800 | 5 835 984 | **+25 184** | `__TEXT` saltó una página: holgura 846 → **3 733** B |
| darwin/arm64 | 5 571 330 | 5 588 450 | **+17 120** | holgura de `__TEXT` 14 144 → 2 157 B |
| windows/amd64 | 5 846 528 | 5 870 592 | **+24 064** | |

**Todas ≤ 102 400 B.** La mayor, darwin/amd64, gasta el **24,6 %** del tope. Quedan B5 *(`--scan`)* y B6 *(textos)*.

**Puertas del corte** *(libre 0.71; 18:49:27 → :35)*:
- **940 / 0 / 0**; vet y lint 0 *(libre 0.79)*; gofmt vacío;
- frontera de FR-027 vacía; ningún `*_test.go` existente modificado *(los 848 de la 0.7.0, verdes)*; `go.mod` sin `require`; sin `go.sum`; `os.Getenv` → 5;
- **`PRIV`**: 0 en `git diff 15c4c47`; 0 en los 3 ficheros nuevos; 0 con `grep -raEc` en `cmd/permea/testdata/copilot/` *(1)* y en
  `internal/ingest/testdata/copilot/` *(52)*.

## B5 · `--scan` *(`esSesionCopilot`, `dryRunCopilot`; créditos sólo en `--scan`)*

- **T033 · Testdata**: `generar.sh` gana `ficheros/control.jsonl` *(llamada 123 456 789 nanoAIU, compactación 1 000, incoherente 999 999 y la llamada
  repetida)*, `control_uno.jsonl` *(1 nanoAIU, no múltiplo de 100)* y `sin_copilotversion.jsonl` *(`session.start` sin `copilotVersion`)*; y sus líneas en
  `esperado.json`. Los demás ficheros salen idénticos *(md5)*.
- **Incidencia del generador** *(antes del verde)*: con `n ≥ 60` el `timestamp` salía con segundos 60–65, que no es RFC 3339, y los registros de control
  eran incoherentes. Se corrige la fórmula *(minutos y segundos de `n`; idéntica para `n < 60`)*: sólo cambian los 3 ficheros nuevos. `generar.sh` `ecc59a45…`.
- **T033 · Rojos** *(libre 0.60; 19:16:42 → :50, 8,3 s; y repetidos con los fixtures corregidos y el código de HEAD, libre 1.10, 7,0 s)*: **947 / 6 / 0**,
  las dos veces:
  - **(37/salida)**: «stdout vacío; stderr: el camino de Claude Code»; **(39/{compactacion,exacto,no_multiplo})**: «control: ""». Con su padre.
  - **Nacen verdes y están declarados**: **(38)** ×4 y **(40)**, y la hoja **(37/disco)**: el camino de Claude Code tampoco escribe *(947 − 940)*.
- **T034 · Verde**: `esSesionCopilot` *(1.ª línea, tope de 1 MiB; `type == "session.start"` y `data.copilotVersion` textual)* tras Gemini, y `dryRunCopilot`
  sobre `state.New()`: `imprimirEventos`, el resumen y la línea de control con `%d.%011d` de `nanoAIU / 10¹¹` y `nanoAIU % 10¹¹`. Los créditos sólo con
  `ContextoCopilot.ConCreditos`, que sólo pone `--scan`; la Σ, en `PasadaCopilot.NanoAIU`, de los eventos emitidos. **953 / 0 / 0** *(libre 1.13; 7,2 s)*.
  md5: `copilot.go` `ecfcca5d`, `main.go` `8712d9ba`, `ingest/copilot.go` `eff50438`, `copilot_contexto.go` `67815cf2`, `copilot_test.go` `a490cbdb`.

**T035 · Censo** *(ANTES de mutar)*: (37)–(40) = `TestCopilotScan_{EventosResumenYControl, LosDemasComoLa070, LineaDeControl, SinCopilotVersionNoEsCopilot}`.
- **Los tests existentes de `--scan`** *(grep de `"--scan"` en los `_test.go`)*, POR NOMBRE Y HOJA, los que caen si todo se lee como Copilot CLI *(«S»)*:
  - `TestScan_{UnEventoPorMensaje, LineaConCuatroPartidasYEventID, LineaConDesgloseDeLaEscritura, LaSalidaQueCreceValeSuMaximo}`;
  - `TestCodexScan_EventosYResumen/{lineas_evento,resumen}`, `TestCodexScan_ClaudeCodeComoLa040`, `TestCodexScan_FormatoAnterior`;
  - `TestGeminiScan_EventosYResumen/{lineas_evento,resumen,proyecto,sin_forma}`, `TestGeminiScan_ClaudeYCodexComoLa050/{claude,codex}`;
  - `TestOpenCodeScan_EventosResumenYContraste/{lineas_evento,resumen,contraste}`, `TestOpenCodeScan_LosDemasComoLa060/{claude,codex,gemini}`,
    `TestOpenCodeScan_ContrasteFueraDeLosMensajes`, `TestOpenCodeScan_FicherosCortosComoLa060/{vacio,diez_bytes}`.
  - **No caen**: las hojas `nada_en_disco` y `disco` *(Copilot tampoco escribe)* y `--scan con "plain"` de `main_test.go` *(sólo mira el código)*.
- **Ajustes a plan §Bloques, declarados aquí**:
  - **m38** → además **(40)**: sin detección, `sin_copilotversion.jsonl` también va a Copilot;
  - **m41** → además **(37/salida)**: compara la línea de control entera, y con `%.9f` el 0 sale con 9 decimales.

| Mut. | Edición | Declarado |
|---|---|---|
| m38 | `dryRun` empieza por `return dryRunCopilot(path)` | (38) ×4, (40) y «S» *(35 FAIL con los padres)* |
| m39 | `dryRunCopilot` guarda su estado en `<fichero>.state.json` | (37/disco) |
| m40 | la Σ de créditos, sin las compactaciones | (39/compactacion), (39/exacto) |
| m41 | el equivalente, `%.9f` sobre `float64` | (37/salida), (39/exacto), (39/no_multiplo) |
| m42 | `esSesionCopilot` sin mirar `copilotVersion` | (40) |

**T035 · Observado** *(cada una: md5 del verde, `bash ~/bin/libre.sh` sola, suite y `comprobar.py`; antes, prueba en seco: las 5 compilan y revierten)*:

| Mut. | md5 mutado | Suite *(libre «inicio»)* | = declarado |
|---|---|---|:-:|
| m38 | `main.go` `26de5012` | 916 / 35 *(1.50)* | ✅ (38) ×4, (40) y «S» *(las 2 hojas de `LineaConCuatroPartidasYEventID` no corren: su padre se para)* |
| m39 | `copilot.go` `5361772c` | 951 / 2 *(0.59)* | ✅ (37/disco) |
| m40 | `copilot_contexto.go` `e10012f5` | 950 / 3 *(1.39)* | ✅ (39/compactacion), (39/exacto) |
| m41 | `copilot.go` `5eb03e52` | 948 / 5 *(1.25)* | ✅ (37/salida), (39/exacto), (39/no_multiplo) |
| m42 | `copilot.go` `7045ce0b` | 952 / 1 *(1.18)* | ✅ (40) |

**B5: 5 de 5 = declarado.** Al final, los md5 del verde *(T034)*. Cada suite, 7–11 s.

**SC-007, informativo en el corte** *(como T031: `git archive 15c4c47` con `0.7.0` frente al árbol con `0.8.0`)*:

| Plataforma | 0.7.0 | Tras B5 | Δ fichero | Nota |
|---|---:|---:|---:|---|
| linux/amd64 | 5 648 536 | 5 681 304 | **+32 768** | |
| linux/arm64 | 5 439 640 | 5 439 640 | 0 | Σ `FileSiz` de los `LOAD`: 5 337 116 → 5 363 132, **+26 016** *(normativo, E-3)* |
| darwin/amd64 | 5 810 800 | 5 840 208 | **+29 408** | holgura de `__TEXT` 846 → 876 B |
| darwin/arm64 | 5 571 330 | 5 605 090 | **+33 760** | holgura de `__TEXT` 14 144 → 15 974 B |
| windows/amd64 | 5 846 528 | 5 875 200 | **+28 672** | |

**Todas ≤ 102 400 B.** La mayor, darwin/arm64, gasta el **33,0 %** del tope.

**Puertas del corte** *(libre 1.39; 19:36:58 → :37:05)*: **953 / 0 / 0**; vet limpio y lint 0; gofmt vacío; frontera de FR-027 vacía; ningún `*_test.go`
existente modificado; `go.mod` sin `require`; sin `go.sum`; `os.Getenv` → 5. **`PRIV`**: 0 en `git diff 15c4c47`; 0 en los 3 ficheros nuevos; 0 con
`grep -raEc` en los dos testdata *(56 ficheros, `generar.sh` incluido)*.

## B6 · README y CHANGELOG *(FR-028; P-8, P-10)*

- **T037 ✋**: el dueño aprueba T1, T2 y R1–R8 de `textos-011/propuesta-v1.md` *(md5 `161d5b36…`)* tal cual; **R9 no entra** *(14-10, encargo 16/n)*. Copia
  sin R9, por programa: **`textos-011/aprobado-v1.md` `2a330ed5…`**, 237 líneas, la fuente desde aquí.
- **T038 · Rojo** *(sobre `README.md` `1dc56f4e`, `CHANGELOG.md` `0e322dec`)*: `grep -c '^## 0.8.0' CHANGELOG.md` → **0**; `grep -c '^### Copilot CLI'
  README.md` → **0**.
- **T039 · Verde**: un programa saca los bloques del aprobado. Sustituye cada «antes» de R1–R8, que era literal y único en HEAD. Inserta T1 antes de «## Coste
  y tarifas» y T2 antes de «## 0.7.0». Otro programa vuelve a sacarlos de los ficheros aplicados, y **`cmp` sin diferencias en los 10**:
  - T1 *(17 líneas)*, T2 *(27)*;
  - R1 README:3-7, R2 :72-73, R3 :143, R4 :158-160, R5 :161-167, R6 :168-172, R7 :251-252, R8 :294.
  - Después: 1 y 1; `PENDIENTE` = 1. `README.md` `edd89c99` *(299 líneas)*, `CHANGELOG.md` `21e22593` *(200)*. **`PRIV`** → 0 en los dos y en el diff.
- **T040 · Puertas** *(libre 0.60; 20:02:27 → :36, 8,6 s)*: **953 / 0 / 0**; vet limpio y lint 0; gofmt vacío; frontera de FR-027 vacía; ningún `*_test.go`
  existente modificado; `go.mod` sin `require`; sin `go.sum`; `os.Getenv` → 5. Sin `strings`: es del cierre *(T043, SC-017)*.
