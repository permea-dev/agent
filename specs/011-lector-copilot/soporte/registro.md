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
