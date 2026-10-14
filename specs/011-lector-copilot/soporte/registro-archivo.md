# 011 · Registro de rojos, verdes y mutaciones — archivo

**§B0, §B1 y §B2, movidos íntegros** desde [`registro.md`](./registro.md) el 2026-10-14 *(dueño, como en 010)*; y **§B3 a §B6**, el mismo día, al
empezar el cierre *(dueño, encargo 17/n)*. **No cuenta para el techo** de 300 líneas.
Texto cerrado: no se edita.

## B0 · FASE 0 y registro

### T001 · Línea base y puertas *(2026-10-14, sobre `e9ad364`; desde `15c4c47`, sólo documentos y `.specify`)*

**La línea base de la 0.7.0** *(encargo 5/n; `git archive 15c4c47` en un temporal, tras `libre.sh`)*: `go test -count=1 -v ./...` → **848 / 0 / 0** en
10 paquetes; `go vet` limpio; lint **0**. **Tamaños** *(`CGO_ENABLED=0 -trimpath -ldflags "-s -w -X main.version=0.7.0"`; referencia de SC-007)*:

| Plataforma | Fichero *(B)* | Nota |
|---|---:|---|
| linux/amd64 | 5 648 536 | |
| linux/arm64 | 5 439 640 | Σ `FileSiz` de los `LOAD`: **5 337 116** |
| darwin/amd64 | 5 810 800 | holgura de `__TEXT`: 846 B |
| darwin/arm64 | 5 571 330 | holgura de `__TEXT`: 14 144 B |
| windows/amd64 | 5 846 528 | |

**La FASE 0 sobre `e9ad364`**, en el árbol de trabajo:

| Puerta | Resultado | Carga *(`/proc/loadavg`)* y duración |
|---|---|---|
| `bash ~/bin/libre.sh` | «inicio» a las 13:45:50, tras esperar desde las 13:39:28 *(otra carga en la máquina)* | 1.42 3.39 2.66 |
| `go test -count=1 -v ./...` | **848 «--- PASS» · 0 FAIL · 0 SKIP**, 10 paquetes `ok`, rc 0 | 13:46:00 → 13:46:09, **8,5 s** · 1.28 → 1.38 |
| `bash ~/bin/libre.sh` | «inicio» a las 13:57:02 *(3 muestras: 1.33 · 1.43 · 1.36)* | 1.36 1.50 1.90 |
| `go vet ./...` · `golangci-lint run` | limpio · **0 issues** | 13:57:10 → 13:57:12, **1,7 s** · 1.45 |
| `gofmt -l .` | vacío | — |
| frontera de FR-027, `git diff 15c4c47` | **vacío** | — |
| `git diff 15c4c47 --name-only --diff-filter=M -- '*_test.go'` | vacío | — |
| `git diff d81c39d --name-only -- '*.go'` | vacío | — |
| `grep -c require go.mod` · `go.sum` | 0 · no existe | — |
| `os.Getenv` de producción | **4** *(`codex.go:19`, `gemini.go:20`, `opencode.go:21`, `:25`)*; pasa a **5** en B1 | — |
| `PRIV` *(tasks §Puertas)* sobre `git diff 15c4c47` | **0** | — |
| `testdata/copilot/` | aún no existen *(nacen en B1)* | — |

**`git diff 15c4c47 --name-only`**: `.specify/feature.json`, `.specify/memory/constitution.md`, y en `specs/011-lector-copilot/`, `contracts/event-id-copilot.md`,
`plan.md`, `spec.md`, `tasks.md` y los tres `soporte/descubrimiento*.md`. **Ningún `.go`.** ⇒ La FASE 0 = la línea base: **848 / 0 / 0**, como en el C1 de 010.

**Incidencia** *(sin efecto en los resultados)*: la primera espera antes de lint se **bloqueó contra sí misma**. La orden que esperaba a `libre.sh`
llevaba `go vet` y `golangci-lint` en su línea de órdenes, y `libre.sh` la veía como una suite en marcha *(la cosecha 279 de la plataforma)*. Se pararon
las dos, se relanzó `libre.sh` solo, y la espera siguiente miró sólo su fichero de salida. **Regla**, al principio de este registro.

**Sin mutaciones en B0.**

## B1 · Raíz y ficheros *(`internal/config/copilot.go`, `ListarCopilot`, `internal/testutil/sandbox.go`)*

- **T003** *(2026-10-14, sobre `7eadff6`)*: stubs `CopilotSessionsRoot` → `("", nil)` e `ingest.ListarCopilot` → `(nil, nil)`; vet limpio.
- **T004**: `internal/ingest/testdata/copilot/generar.sh` `bb9b7c30…` *(`python3 -I`)* → `sesiones/`: `s-…01/events.jsonl` `448bd776…`,
  `s-…01/checkpoints/index.md` `87612f2c…`, `s-…02/events.jsonl` `b989848d…`, `s-…02/sub/events.jsonl` `fcf7d524…` *(a dos niveles)*,
  `s-…03/checkpoints/index.md` `e9733dcb…` *(sin `events.jsonl`)* y `.session-operation-locks/s-…01.lock` *(0 B)*. **El `session.db` de 0 B no se
  versiona** *(`.gitignore`: `*.db`; tocar `.gitignore` sería PARA 1)*: lo crea el test en su copia de `t.TempDir()`. Ningún test abre `testdata/` en su sitio.
- **T005 · Rojos** *(libre 0.58; 14:24:13 → :24, 10,8 s)*: **848 / 13 / 0**. Caen, por el stub:
  - (1) `TestCopilotSessionsRoot_Ubicacion/{definida,vacia,ausente,inexistente}`: `= ("", <nil>); se esperaba (<ruta>, nil)`;
  - (2) `TestSandbox_VaciaCopilotHome`: `COPILOT_HOME = <tmp> (definida: true)`, porque la variable se puso antes de `Sandbox`;
  - (3) `TestListarCopilot/{solo_events,sin_locks,sin_checkpoints_db,hondo,vacia,orden}`: «ListarCopilot no devolvió nada»;
  - y los 2 padres.
- **T006 · Verde**:
  - `copilot.go` `7e1b20a9…`, `copilot_contexto.go` `19d09e34…`;
  - `sandbox.go` `2c6b3301…`: una línea tras `OPENCODE_DB`, con su comentario;
  - los comentarios de `codex.go`, `gemini.go` y `opencode.go`, «una de las CINCO» / «dos de las CINCO».

  Resultados *(libre 0.72; 14:25:28 → :35, 7,3 s)*: **861 / 0 / 0**. Vet y lint 0 *(3,3 s)*; gofmt vacío.
  Tests: `copilot_test.go` `b7479e18…`, `copilot_fichero_test.go` `48190957…`, `sandbox_copilot_test.go` `59cd7ac4…`.
- **T007**: `sandbox_test.go`, verde y sin tocar *(`TestSandbox_AislaLasTresVariables`, `…DistingueSegmentos` ×6, `…ConSemillas…`)*; `os.Getenv` de
  producción → **5** *(`copilot.go:19`, `codex.go:19`, `gemini.go:20`, `opencode.go:21`, `:25`)*.

**T008 · Censo** *(ANTES de mutar)*: (1) = `TestCopilotSessionsRoot_Ubicacion/`, (2) = `TestSandbox_VaciaCopilotHome`, (3) = `TestListarCopilot/`.
- **Quién más las lee**: nadie *(grep)*. Ni el binario ni otros tests llaman aún a `CopilotSessionsRoot` o `ListarCopilot` *(B4)*, ni leen `COPILOT_HOME`.
- **Co-caídas declaradas aquí que plan §Bloques no listaba**: **m2 → (1/vacia)**, porque la rama del hogar pierde también `session-state`. Y **m5 →
  (3/solo_events)**, porque la lista exacta gana `s-…02/sub/events.jsonl`. Con cada caída de hoja cae también su padre.

| Mut. | Edición | Declarado |
|---|---|---|
| m1 | `propia := os.Getenv(…); propia != ""` → `propia, definida := os.LookupEnv(…); definida` *(vacía como definida)* | (1/vacia) |
| m2 | sin `"session-state"` en los dos `Join` | (1/definida), (1/vacia), (1/ausente), (1/inexistente) |
| m3 | `propia != ""` → `false && propia != ""` | (1/definida) |
| m4 | sin la línea `t.Setenv("COPILOT_HOME", "")` del sandbox *(y su comentario)* | (2) |
| m5 | `os.ReadDir` de un nivel → `filepath.WalkDir` recursivo: todo `events.jsonl` a cualquier profundidad, saltando las carpetas que empiezan por `.` | (3/solo_events), (3/hondo) |
| m6 | carpeta sin `events.jsonl` → `return nil, error` | (3) ×6 |

**T008 · Observado** *(cada mutación: md5 de partida comprobado, `bash ~/bin/libre.sh` aparte, suite, reversión por edición inversa y md5 restaurado)*:

| Mut. | Fichero · md5 mutado | Suite *(libre «inicio»)* | Cayó | = declarado |
|---|---|---|---|:-:|
| m1 | `copilot.go` `bb243797` | 859 / 2 *(1.13)* | (1/vacia) + padre | ✅ |
| m2 | `copilot.go` `638f5868` | 856 / 5 *(0.77)* | (1/definida, vacia, ausente, inexistente) + padre | ✅ |
| m3 | `copilot.go` `16bd7b2c` | 859 / 2 *(0.64)* | (1/definida) + padre | ✅ |
| m4 | `sandbox.go` `98c03742` *(= el de 010 B5: sólo falta la línea nueva)* | 860 / 1 *(1.77)* | (2) | ✅ |
| m5 | `copilot_contexto.go` `d6f4b25f` | 858 / 3 *(1.36)* | (3/solo_events, hondo) + padre | ✅ |
| m6 | `copilot_contexto.go` `2a04c9b3` | 854 / 7 *(1.56)* | (3) ×6 + padre | ✅ |

**6 de 6 = declarado.** Tras m6, los tres md5 son los del verde *(`7e1b20a9`, `19d09e34`, `2c6b3301`)*. Cada suite tardó 7–8 s.
- **Incidencia, sin efecto**: la primera reversión de m4 no se aplicó. La edición inversa partía de una cadena vacía y no tenía ancla única. El guion se paró
  en su aserción, **sin escribir nada**. Se ancló a la línea de `OPENCODE_DB`, se revirtió y el md5 dio el del verde, antes de tocar m5.

**Puertas del corte** *(libre 1.69; 14:35:05 → :12)*:
- **861 / 0 / 0**; vet y lint 0 *(libre 1.45)*; gofmt vacío;
- frontera de FR-027 vacía; ningún `*_test.go` modificado; `go.mod` sin `require`; sin `go.sum`; `os.Getenv` → **5**;
- **`PRIV`** *(tasks §Puertas)*: 0 en `git diff 15c4c47`, 0 en cada fichero nuevo sin seguir, y 0 con `grep -raEc` en los 7 ficheros de
  `internal/ingest/testdata/copilot/` *(`generar.sh` incluido)*.

## B2 · Una línea e identidad *(`internal/ingest/copilot.go`, `copilot_eventid.go`)*

- **T010** *(2026-10-14, sobre `89fbe14`)*: stubs de `derivarEventIDCopilot` → `("", false)` y de `RegistroCopilot`, que no clasifica, con
  `ContextoCopilot`, `ClaseCopilot` y `MarcasCopilot`. La firma devuelve también las marcas *(`compactación`, `sin modelo`)*, como `RespuestaGemini`:
  plan §1 la escribía con tres retornos, y D-011-P4 las define como marcas.
- **T011**: `generar.sh` `295bae49…` gana `lineas/`, 24 ficheros *(`sesiones/` sale idéntico a B1)*:
  - eventos: `evento` `22860ea8`, `sc005` `810d78be`, `razonamiento` `8cf82608`, `desfase` `858626d3`, `sin_modelo` `70f22f3e`;
  - sin identificador: `sin_usageid_{ausente,vacio,no_textual}` `28f57fe8` · `f207b1a5` · `a2ec0268`;
  - incoherentes: `incoherente_{cuatro_ausentes,no_numerica,no_entera,negativa,no_cabe,timestamp_malo,ctu_no_objeto}`;
  - compactación: `compactacion` `c086f818`, `compactacion_{fallida,sin_ctu,sin_requestid}`;
  - `centinelas` `12501d6a`, `centinelas_compactacion` `9f236bd4`;
  - no registro: `falso_user` `77c8edbc`, `falso_system` `d66e1030`, `model_aux` `f0e6d4b8` *(tres `model.*`)*.

  **(8/usage_no_objeto)** es una compactación con `compactionTokensUsed: 5` y `requestId`. En un `usage_record`, un `usage` que no es objeto no deja leer
  el `usageId`, y por el orden de D-011-P4 es «sin identificador».
- **T012 · Rojos** *(libre 1.53; 14:48:11 → :20, 8,3 s)*: **867 / 35 / 0**.
  - **Caen por el stub**: (4) ×5 *(«no dio valor»)*; (6); (7/partidas, sin_ctu, sin_requestid); (8) ×7; (9) ×3; (10) ×4; (11) ×2; (12); (14) ×2
    *(«se esperaba un evento»)*; y sus padres.
  - **Nacen verdes y están declarados**: (5) *(`TestEventIDCopilot_VacioNoDerivaNada`)*, (13) ×3 *(`TestRegistroCopilot_Prefiltro/{falso_user,
    falso_system,model_aux}`)* y la hoja (7/fallida). Son las 6 de 867 − 861.
- **T013 · Verde**: `copilot_eventid.go` `ce92c11d…` *(replica el bucle; reutiliza `dominioEventID`, `bytesEventID` y `tipoRespuesta`)* y `copilot.go`
  `fbd7d7c1…`. **No** usa `internal/pricing`.
  - **Suite** *(libre 1.70; 7,5 s)*: **902 / 0 / 0**.
  - **Lint**: 1 issue en el test nuevo *(S1011, un bucle de `append`)*. Se corrigió en `copilot_registro_test.go` `9eea7ffb…`; vet y lint 0; suite otra vez
    **902 / 0 / 0** *(libre 0.92; 7,1 s)*.
- **T014**: (14) ×2, rojo en T012, verde con T013.
- **Incidencia** *(sin efecto en los resultados)*: la espera del segundo lint **se volvió a bloquear contra sí misma**, porque llevaba `golangci-lint` en
  su línea de órdenes. Es la regla del principio de este registro, incumplida. **Además frenó a la sesión de la plataforma**, que esperaba en su propio
  `libre.sh`. Se pararon las dos órdenes; la siguiente espera sólo miró su fichero.

**T015 · Censo** *(ANTES de mutar)*: (4) = `TestEventIDCopilot_Vectores/`, (5) = `…_VacioNoDerivaNada`, (6)–(14) = `TestRegistroCopilot_{SC005Partidas,
Compactacion, Incoherente, SinIdentificador, HerramientaYCoste, MomentoEnUTC, SinModelo, Prefiltro, CentinelasFueraDelEvento}`. Nadie más llama a
`RegistroCopilot` ni a `derivarEventIDCopilot` *(grep)*.

**Dos ajustes a plan §Bloques, declarados aquí**:
- **m7 → (10/tool)**: `herramientaCopilot` es la misma constante para el hash y para `tool` *(D-1)*;
- **m16 → sólo (10/coste)**: (10/nanoaiu) sólo mira `cost_usd`.

| Mut. | Edición | Declarado |
|---|---|---|
| m7 | `herramientaCopilot = "copilot_cli"` → `"codex"` | (4) ×5, (10/tool) |
| m8 | sin el prefijo de longitud *(no se escribe `longitud`)* | (4) ×5 |
| m9 | sin `if id == "" { return "", false }` | (5), (9) ×3, (7/sin_requestid) |
| m10 | `tipoCompactacionCopilot = "compactacion"` → `"respuesta"` | (4/compactacion), (4/mismo_valor) |
| m11 | la salida suma `reasoningTokens` | (6), (10/razonamiento) |
| m12 | `entrada − lectura − escritura` → `entrada − lectura` | (6), (7/partidas) |
| m13 | `TokensCacheRead: int(p.lectura)` → `int(p.escritura)` | (6), (7/partidas) |
| m14 | sin la comprobación de `success` | (7/fallida) |
| m15 | sin `p.lectura+p.escritura > p.entrada` | (8/no_cabe) |
| m16 | `CostAvailable: false` → `true` | (10/coste) |
| m17 | `momento.UTC()` → `momento` | (11/desfase) |
| m18 | sin modelo → `return nil, IncoherenteCopilot, …` | (12) |
| m19 | sin el prefiltro de bytes | (13) ×3 |
| m20 | `SessionRef: event.Ref(ctx.Salt, ctx.SessionID)` → `ctx.SessionID` | (14) ×2 |

**T015 · Observado** *(cada una: md5 del verde comprobado, `bash ~/bin/libre.sh` aparte, suite, comparación por programa con el censo
—`comprobar.py`, hojas y padres—, reversión por edición inversa y md5 restaurado)*:

| Mut. | md5 mutado | Suite | = declarado |
|---|---|---|:-:|
| m7 | `copilot_eventid.go` `7e4a2e37` | 894 / 8 | ✅ (4) ×5 + (10/tool) |
| m8 | `copilot_eventid.go` `2422a98b` | 896 / 6 | ✅ (4) ×5 |
| m9 | `copilot_eventid.go` `117b77d5` | 895 / 7 | ✅ (5), (9) ×3, (7/sin_requestid) |
| m10 | `copilot_eventid.go` `126f20a0` | 899 / 3 | ✅ (4/compactacion), (4/mismo_valor) |
| m11 | `copilot.go` `5494337d` | 899 / 3 | ✅ (6), (10/razonamiento) |
| m12 | `copilot.go` `cb71d241` | 899 / 3 | ✅ (6), (7/partidas) |
| m13 | `copilot.go` `110c7979` | 899 / 3 | ✅ (6), (7/partidas) |
| m14 | `copilot.go` `36a9541f` | 900 / 2 | ✅ (7/fallida) |
| m15 | `copilot.go` `3bee8a69` | 900 / 2 | ✅ (8/no_cabe) |
| m16 | `copilot.go` `9f0bdac7` | 900 / 2 | ✅ (10/coste) |
| m17 | `copilot.go` `82304a0a` | 900 / 2 | ✅ (11/desfase) |
| m18 | `copilot.go` `5f3583ca` | 901 / 1 | ✅ (12) |
| m19 | `copilot.go` `139dbc16` | 898 / 4 | ✅ (13) ×3 |
| m20 | `copilot.go` `8f708e1b` | 899 / 3 | ✅ (14) ×2 |

**14 de 14 = declarado.** Libre «inicio» entre 0.93 y 1.94; cada suite, 7–8 s. Tras m20, los md5 son los del verde *(`fbd7d7c1`, `ce92c11d`)*.

**Puertas del corte** *(libre 1.86; 15:15:52 → :16:00)*:
- **902 / 0 / 0**; vet y lint 0 *(libre 1.95)*; gofmt vacío;
- frontera de FR-027 vacía; ningún `*_test.go` modificado; `go.mod` sin `require`; sin `go.sum`; `os.Getenv` → 5;
- **`PRIV`**: 0 en `git diff 15c4c47`, 0 en los 28 ficheros nuevos sin seguir, y 0 con `grep -raEc` en los 31 de `testdata/copilot/` *(T016)*.

**T050 · SC-007, informativo tras B2** *(`git archive 15c4c47` con `-X main.version=0.7.0`, frente al árbol con `0.8.0`; banderas de goreleaser)*:

| Plataforma | 0.7.0 | Tras B2 | Δ fichero | Nota |
|---|---:|---:|---:|---|
| linux/amd64 | 5 648 536 | 5 648 536 | 0 | |
| linux/arm64 | 5 439 640 | 5 439 640 | 0 | Σ `FileSiz` de los `LOAD`: 5 337 116 → 5 337 372, **+256** |
| darwin/amd64 | 5 810 800 | 5 810 800 | 0 | holgura de `__TEXT`: 846 → **718** B |
| darwin/arm64 | 5 571 330 | 5 571 330 | 0 | holgura de `__TEXT`: 14 144 → 14 016 B |
| windows/amd64 | 5 846 528 | 5 846 528 | 0 | |

**Lectura**: el lector todavía **no está enlazado**. Nadie llama a `RegistroCopilot` ni a `CopilotSessionsRoot` hasta B4, y el enlazador los elimina.
La medida que cuenta es la del corte de B4 *(T031)*, con el lector ya llamado.

---

<!-- §B3–§B6: desde aquí, íntegros, las líneas 12–261 de registro.md en «011 B6: README y CHANGELOG de la 0.8.0» -->

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
