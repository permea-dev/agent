# 013 · Registro de la implementación

**Por qué existe**: como en 007–012, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. **Techo: 250 líneas**, fijado por el
dueño el 23-10 *(`tasks.md` §Presupuesto)*. ⛔ **Ni rutas reales, ni nombres de carpeta, ni contenido** *(PARA 8)*: sólo recuentos, tiempos, tamaños y md5.

- **§B0** *(línea base 1074 / 0 / 0, tamaños de la 0.9.0, la referencia de SC-012 `e6b011db…` y su incidencia de tamaños)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).

## B1 · Leer sin escribir *(`internal/config`; sobre `6b318c5`)* — PARA 2 en M-B1a, **cerrada**

- **T004** · Esqueletos: `RutaDataDir()` → `("", nil)`, en `config.go`; `LeerSalt(dir)` → `("", false, nil)`, en `identity.go`. Compila.
- **T005** · **Rojos (1) y (2)**, en `internal/config/lectura_test.go`, con `testutil.Sandbox`. La razón se escribió **antes** de correrlos. Se esperaba
  **1077 / 7 / 0** y salió **1077 / 7 / 0** *(libre 0.29; 13:26:19 → :25)*.
  - **Caen, por lo que prueban**:
    - (1/misma_ruta) y (1/no_crea): `RutaDataDir` devuelve `""`, no la ruta;
    - (2/igual) y (2/igual_con_salto): `LeerSalt` devuelve `""`, no la sal;
    - (2/ilegible): no devuelve error.
  - **Nacen verdes, declaradas**: (1/datadir_crea), porque `DataDir` no cambia *(la valida M-B1a)*; (2/ausente) y (2/ausente_con_datadir), porque el esqueleto
    ya responde lo correcto *(las valida m2)*.
- **T006** · **Verde**, según plan §3:
  - `RutaDataDir` = `os.UserConfigDir()` + `permea`, sin `MkdirAll`; `DataDir` = `RutaDataDir` + `MkdirAll`;
  - `LeerSalt` lee como `loadOrCreateSecret`: tal cual, `existe = false` si no existe, y error en cualquier otro fallo.

  Suite *(libre 0.14; 13:27:04 → :10)*: **1084 / 0 / 0**. md5: `config.go` `19b0d970…`, `identity.go` `5ca4478e…` y `lectura_test.go` `317d7102…`.
- **T007** · Mutaciones, con el censo declarado **antes**; cada una con la suite entera tras `libre.sh` y la reversión por edición inversa con md5:

  | Mutación | Declarado | Cayó | md5 tras revertir |
  |---|---|---|---|
  | **m1** `RutaDataDir` con `MkdirAll` | `TestRutaDataDir`, `/no_crea` | **lo mismo** *(1082 / 2)* | `config.go` `19b0d970…` ✅ |
  | **m2** `LeerSalt` delega en `LoadOrCreateSalt` | `TestLeerSalt`, `/ausente`, `/ausente_con_datadir` | **lo mismo** *(1081 / 3)* | `identity.go` `5ca4478e…` ✅ |
  | **m3** `strings.TrimSpace` sobre la sal | `TestLeerSalt`, `/igual_con_salto` | **lo mismo** *(1082 / 2)* | `identity.go` `5ca4478e…` ✅ |
  | **M-B1a**, 1.ª forma: **PARA 2 por censo incompleto** · `DataDir` sin `MkdirAll` | `TestRutaDataDir`, `/datadir_crea` | eso **y además** `TestEnroll_ReenrollNoResidue`, `TestEnroll_Argv_HappyPath` y `TestEnroll_Stdin_And_SC011` *(1078 / 5)* | puesta hasta la decisión; revertida después: `config.go` `19b0d970…` ✅ |
  | **M-B1a**, 2.ª forma *(censo corregido; orquestador, 23-10, opción 1)* | `TestRutaDataDir`, `/datadir_crea`, **más** los tres `TestEnroll_*` | **lo mismo** *(1078 / 5; `diff` del declarado contra lo caído, vacío)* | `config.go` `19b0d970…` ✅ |

- ⛔ **PARA 2.** M-B1a tumba tres tests de `enroll` que el censo no declaró. **La mutación se queda puesta.**
  - **Por qué caen**: `enroll_test.go:65`, `:115` y `enroll_reject_test.go:260`. Esos tests apuntan `XDG_CONFIG_HOME` a un temporal sin `permea/`, y `enroll`
    escribe allí con `atomicWrite`: «no such file or directory».
  - **Es un error del censo, no del código.** El censo sólo buscó los tests que borran el `DataDir`. Esos tres sí prueban que `DataDir` crea, así que la
    conducta de hoy ya estaba cubierta por tests existentes, como decía el encargo.
- **Cierre** *(decisión del orquestador, 23-10: opción 1)*:
  - se revirtió M-B1a por edición inversa *(→ `19b0d970…`)*;
  - se corrigió su censo: los tres `TestEnroll_*` **escriben en el `DataDir` de un `XDG_CONFIG_HOME` vacío**, y sin el `MkdirAll` no pueden;
  - se repitió, y cayó **exactamente** lo declarado.
- **El censo, por los llamantes**: `DataDir()` tiene cuatro en producción, `setup()` *(`main.go:207`)*, `enroll` *(`enroll.go:69`)*, `status` *(`status.go:24`)* y
  `project join` *(`project.go:163`)*. Sólo los tres `TestEnroll_*` los ejercen con un `XDG_CONFIG_HOME` sin `permea/`. Los demás corren en `testutil.Sandbox`,
  que lo crea.
- 📌 **La lección**: el censo de una mutación sobre una función **compartida** se hace buscando **todos sus llamantes**, y los tests que ejercen cada uno, **no
  sólo los tests que parecen relacionados**.
- **Corte de B1** *(sobre el verde)*:
  - suite *(libre 0.19; 13:34:09 → :15)*: **1084 / 0 / 0**;
  - gofmt vacío, vet limpio y lint 0 *(libre 0.24)*;
  - frontera 0 B, ningún `_test.go` existente modificado, `go.mod` sin `require` y sin `go.sum`;
  - `PRIV` 0 en el diff, en `lectura_test.go` y, con `-a`, en el testdata.

## B2 · La derivación con origen *(`internal/project`; sobre `651cbdd`)*

- **T009** · Esqueletos, al final de `resolve.go`: `Origen{Ruta, HuboRaiz}`; `DerivarConOrigen` → `("", "", false)`; `Resolutor.Origen` → `(Origen{}, false)`.
- **T010** · **Rojos (3) y (4)**, en `internal/project/origen_test.go`.
  - (3) **copia** cada familia de `testdata/` a un `t.TempDir()` *(nunca la abre en su sitio)*, construye su árbol y, por cada `cwd`, exige cuatro cosas:
    - la identidad y `huboRaiz` de `Derivar` y `DerivarConRaiz`;
    - el origen esperado;
    - **`event.Ref(sal, origen)` = la identidad**: la prueba que no depende de la implementación.
  - La razón se escribió **antes** de correrlos: se esperaba **1084 / 10 / 0** y salió **1084 / 10 / 0** *(libre 0.66)*.
  - **Caen, por lo que prueban**:
    - (3/subdirectorio), (3/worktree), (3/fallback), (3/enlace) y (3/degradados): `DerivarConOrigen` = `""`, distinta de la derivación de siempre. En
      `degradados` caen los cuatro `cwd` no vacíos;
    - (4/tras_derivar), (4/acierto_de_cache) y (4/dos_cwd_una_raiz): `Origen` = `false`.

    Ninguna nace verde.
- **T011** · **Verde**, según plan §2. **Un solo cuerpo**:
  - `derivarConTechoYRaiz` devuelve `(identidad, origen, huboRaiz)`, con el origen saliendo **del mismo `return`** que la identidad: `raiz` o `ubicacion`;
  - `DerivarConOrigen` es la entrada, con las dos líneas de siempre y el comentario del ORDEN;
  - `DerivarConRaiz` delega en ella y descarta el origen;
  - `derivarConTecho` descarta los dos.

  El `Resolutor` gana `origenes map[string]Origen`, por identidad, y la apunta en un fallo de caché. `cache` no cambia de forma. Suite *(libre 0.48)*:
  **1094 / 0 / 0**. md5: `resolve.go` `009a64d2…` y `origen_test.go` `3d2cb1b1…`.
- **«El ref no cambia ni un byte»**: además de los tests de siempre, se repitió la **referencia de SC-012** con el código de B2, con el test temporal de B0 en una
  copia del árbol: **la cola, el stdout y el stderr, byte a byte iguales** a los de `cabf8c5`.
- **`HogarReal` y `CuelgaDe`** *(plan §2, último punto)* **se dejan para B5**, donde los usa `aCasa` y los prueba (15). Añadirlos aquí sería producción sin rojo.
- **T012** · Mutaciones, **con el censo por llamantes** *(la lección de B1)*:
  - **m4 y m5** tocan el origen que devuelve `DerivarConOrigen`. La llaman `DerivarConRaiz` *(y con ella `Derivar` y `project join`)*, que **descarta el origen**;
    `Resolutor.Derivar`, que lo apunta; y los tests (3) y (4). El origen apuntado sólo lo lee `Resolutor.Origen`, que no tiene llamantes de producción.
  - **m6** toca la clave de `origenes`. A `Resolutor.Derivar` lo llaman los seis lectores, pero la identidad no cambia; `Origen` sólo lo lee (4).

  | Mutación | Declarado *(y nada más)* | Cayó | Reversión |
  |---|---|---|---|
  | **m4** origen = `filepath.Clean(cwd declarado)` | (3) `/subdirectorio`, `/worktree`, `/enlace`, `/degradados` *(el vacío da `.`)* · (4) `/tras_derivar`, `/acierto_de_cache` · y los 2 padres | **lo mismo** *(1086 / 8)* | `009a64d2…` ✅ |
  | **m5** origen = `cwd` declarado, sin `Clean` | (3) `/subdirectorio`, `/worktree`, `/fallback`, `/enlace` · (4) `/tras_derivar`, `/acierto_de_cache` · y los 2 padres | **lo mismo** *(1086 / 8)* | `009a64d2…` ✅ |
  | **m6** `origenes` por `cwd` y no por identidad | (4) `/tras_derivar`, `/acierto_de_cache`, `/dos_cwd_una_raiz` · y su padre | **lo mismo** *(1090 / 4)* | `009a64d2…` ✅ |

  `(4/dos_cwd_una_raiz)` sobrevive a m4 y m5 porque gana la última escritura, la del `cwd` de la raíz, que ya es la raíz.
- **`resolve_test.go` sigue verde y sin tocar**: 0 B de diff; sus tests, en PASS en el verde y en el corte.
- **Corte de B2**:
  - suite *(libre 0.51; 13:46:52 → :59)*: **1094 / 0 / 0**;
  - gofmt vacío, vet limpio y lint 0 *(libre 0.33)*;
  - frontera 0 B, ningún `_test.go` existente modificado, `go.mod` sin `require` y sin `go.sum`;
  - `PRIV` 0 en el diff y en `origen_test.go`.
- **Incidencias**:
  1. la primera repetición de la referencia de SC-012 salió distinta por **mi copia**, no por el código: `tar --exclude=tmp` dejó fuera también
     `plantillas/gemini/tmp/`. Repetida con `git ls-files`, salió igual;
  2. un `go test ./internal/project/` de 0,1 s corrió en la misma orden, tras la suite del corte, **sin otro `libre.sh`** delante.

## B3 · El recorrido común *(`cmd/permea`; sobre `4db2f8e`)* — PARA 2 en m8, **cerrada**

- **Antes de B3** *(decisiones del 23-10)*: §B0 pasa a `registro-archivo.md`, con una remisión de una línea, y plan §2 anota `HogarReal` y `CuelgaDe` en B5.
- **T014** · **(5)**, en `cmd/permea/recorrido_test.go`, **antes de tocar `main.go`**. `--run` sobre las plantillas, con `@CASA@` → `/sintetico/casa`, en
  `SandboxConSemillas`, comparado con la referencia de T002 en la cola, el stdout y el stderr. **Nace verde**, como se declaró *(lo valida m8)*: **1098 / 0 / 0**
  *(libre 1.32)*.
- **T015** · Esqueleto en `main.go`: `type sumidero func(event.Event) error` y un `recorrer` que devuelve `0, nil, nil`. **Rojo (6)**: en el propio proceso,
  `setup()`, `FindLogs` y `recorrer` sobre `state.New()`, con un sumidero que recoge.
  - Se declaró antes de correr: **(6/mismos_eventos) cae**, porque recoge 0 eventos, y **(6/no_escribe) nace verde**, porque el esqueleto no escribe *(lo valida m7)*.
  - Se esperaba **1099 / 2 / 0** y salió **1099 / 2 / 0** *(libre 0.64)*.
- **T016** · **Verde**, según plan §1:
  - el cuerpo de `generate()`, desde la Pasada hasta Kimi, pasa **tal cual** a `recorrer(logs, st, ictx, sumar)`, con los **siete** `transport.Append` cambiados
    por `sumar` *(dos en Claude Code, uno en cada `generar…`)*;
  - `generate()` conserva su firma y su orden: `ClaudeCodeLogsRoot` → `FindLogs` → `state.Load` → `Resolutor` → `recorrer`, con un sumidero que encola → `st.Save`;
  - las cinco raíces salen de `setup()` a `resolverRaices()`, tras `machine_id`, y la parada por `plain` sigue siendo la primera;
  - `gemini.go`, `opencode.go`, `copilot.go` y `kimi.go` ya no importan `transport`.

  Suite *(libre 0.41)*: **1101 / 0 / 0**, y **(5) en verde con el recorrido nuevo: SC-012, byte a byte**. md5: `main.go` `c573590e…`, `gemini.go` `a81cef22…`,
  `opencode.go` `e6271e37…`, `copilot.go` `5b28b23c…`, `kimi.go` `f623a19b…` y `recorrido_test.go` `3f602bbc…`.
- **T017** · Mutaciones. El censo se hizo por llamantes: los tests que montan Codex *(`raizCodex`, `todasLasHerramientas`, `CODEX_HOME`)* y leen la cola,
  **16 candidatos**, leídos uno a uno.

  | Mutación | Declarado | Cayó | md5 |
  |---|---|---|---|
  | **m7** `generarKimi` encola con `transport.Append` | `TestRecorrido_ElSumideroRecogeLoQueSeEncola`, `/mismos_eventos`, `/no_escribe` | **lo mismo** *(1098 / 3)* | `kimi.go` `f623a19b…` ✅ |
  | **m8**, 1.ª forma: **PARA 2, censo sin bajar a las hojas** · el sumidero de `generate()` salta `codex` | 8 tests de Codex, Gemini y OpenCode, más `TestRecorrido_RunEsLaReferenciaDeLa090` y `/cola` | **los mismos tests**, y **5 hojas suyas sin declarar** *(1086 / 15)* | puesta hasta la decisión; revertida después: `main.go` `c573590e…` ✅ |
  | **m8**, 2.ª forma *(censo por hoja; orquestador, 23-10)* | los 10 de arriba **y** `…Codex…FicheroIlegibleNoRompe/pasada`, `/se_relee`, `…Gemini…FicheroIlegibleNoRompe/pasada`, `/segunda_pasada` y `…OpenCode…BaseCorruptaNoRompe/pasada`: **15** | **lo mismo** *(1086 / 15; `diff` del declarado contra lo caído, vacío)* | `main.go` `c573590e…` ✅ |

- ⛔ **PARA 2.** m8 **no tumbó ningún test de más**. Tumbó hojas de tres tests declarados que el censo no listó:
  - `TestCodexGenerate_FicheroIlegibleNoRompe/pasada` y `/se_relee`;
  - `TestGeminiGenerate_FicheroIlegibleNoRompe/pasada` y `/segunda_pasada`;
  - `TestOpenCodeGenerate_BaseCorruptaNoRompe/pasada`.

  **Causa, mía**: leí sus aserciones filtrando las líneas por palabras clave, y el filtro escondió los `t.Run`, así que concluí que no tenían hojas. El
  protocolo cuenta **hojas**. **La mutación se queda puesta**, y las puertas del corte no se han pasado.
- 📌 **Segunda lección**: el censo por llamantes **baja hasta las hojas**. De cada test candidato se listan sus `t.Run`, leyendo el cuerpo entero, no un filtro.
- **Cierre** *(orquestador, 23-10, como en B1)*:
  1. se revirtió m8 por edición inversa *(→ `c573590e…`)*;
  2. se **releyeron enteros, sin filtro, los 16 candidatos**, con sus hojas. Caen, y sólo por los eventos `codex` de la cola:
     - `TestCodexRun_SinClaudeCode`, `…EncolaAntesDeGuardar`, `…SegundaPasadaCero`, `…NadaDelProveedorViaja` y `…EnCadaPasada`, sin hojas;
     - `TestCodexGenerate_FicheroIlegibleNoRompe/pasada` y `/se_relee`, pero no `/segunda_pasada`, que sólo cuenta Claude Code;
     - `TestGeminiGenerate_FicheroIlegibleNoRompe/pasada` y `/segunda_pasada`, pero no `/se_relee`, que sólo cuenta Gemini;
     - `TestOpenCodeGenerate_BaseCorruptaNoRompe/pasada`, pero no `/se_relee`;
     - `TestRecorrido_RunEsLaReferenciaDeLa090/cola`, pero no `/stdout` ni `/stderr`, porque «N eventos encolados» cuenta igual.

     **Sobreviven**:
     - `TestCodexRun_LineaDeResumen`, `TestGeminiRun_LineaDeResumen` y `TestOpenCodeRun_LineaDeResumen`, que sólo miran el stderr;
     - `TestCopilotRun_LineaDeResumen/linea`, `/orden_cola` *(OpenCode→Copilot)* y `TestKimiRun_LineaDeResumen/linea`, `/total`, `/orden_cola`;
     - `TestKimiRun_FicheroIlegibleNoRompe/pasada`, `/se_relee` *(sólo Kimi)*;
     - (6), que usa su propio sumidero.
  3. se repitió, y cayó **exactamente** lo declarado; revertida otra vez.
- **Corte de B3**:
  - suite *(libre 0.57; 14:14:06 → :12)*: **1101 / 0 / 0**;
  - gofmt vacío, vet limpio y lint 0 *(libre 0.42)*;
  - frontera 0 B, ningún `_test.go` existente modificado, `go.mod` sin `require` y sin `go.sum`;
  - `PRIV` 0 en el diff, en lo nuevo y, con `-a`, en el testdata.

## B4 · Agregación y nombres

## B5 · La tabla

## B6 · El comando

## B7 · Los seis lectores y la frontera

## B8 · README y CHANGELOG

## Cierre
