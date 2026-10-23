# 013 · Registro de la implementación

**Por qué existe**: como en 007–012, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. **Techo: 250 líneas**, fijado por el
dueño el 23-10 *(`tasks.md` §Presupuesto)*. ⛔ **Ni rutas reales, ni nombres de carpeta, ni contenido** *(PARA 8)*: sólo recuentos, tiempos, tamaños y md5.

- **§B0** *(línea base 1074 / 0 / 0, tamaños de la 0.9.0, la referencia de SC-012 `e6b011db…` y su incidencia de tamaños)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).

- **§B1** *(leer sin escribir: 1084 / 0 / 0; m1–m3 y M-B1a = declarado, con su PARA 2 por censo incompleto y la lección de los llamantes)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).
- **§B2** *(la derivación con origen: 1094 / 0 / 0; SC-012 repetida, byte a byte; m4–m6 = declarado)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).
- **§B3** *(el recorrido común: 1101 / 0 / 0; SC-012 byte a byte; m7 y m8 = declarado, con su PARA 2 por no bajar a las hojas)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).

## B4 · Agregación y nombres *(`cmd/permea/lista.go`; sobre `8ef0d68`)*

- **T019** · Esqueleto en `cmd/permea/lista.go`: `fila`, `nombresComerciales` vacía y un `agregar` que devuelve `nil`.
- **T020** · **Rojos (7) a (12)**, en `cmd/permea/lista_agregar_test.go`, con eventos hechos a mano y `origen` y `existe` inyectados.
  - Se declaró antes de correr: caen las **13 hojas** porque no hay filas, más los 3 padres con hojas; **ninguna nace verde**.
  - Se esperaba **1101 / 16 / 0** y salió **1101 / 16 / 0**, sin pánicos *(libre 0.20, tras unos 13 min de espera a la plataforma)*.
- **T021** · **Verde**, según plan §4 y §5:
  - `agregar` salta los eventos sin identidad y los `event_id` ya vistos, en cualquier identidad *(el primero gana)*;
  - lleva por identidad el recuento, el conjunto de `Tool` y el mínimo y el máximo de `OccurredAt.UTC()`;
  - pone la marca con `existe` primero y `HuboRaiz` después, y ordena por eventos y luego por identidad;
  - `conNombre` pasa a nombres comerciales en el orden de la tabla, con los desconocidos detrás y por orden alfabético;
  - la tabla tiene los seis de D-1.

  `existeEnDisco`, el `existe` de producción, **se deja para B7**, donde lo usa la tubería y lo prueba (22). Suite *(libre 0.32)*: **1117 / 0 / 0**. md5:
  `lista.go` `3cce9963…` y `lista_agregar_test.go` `163b525b…`.
- **T022** · Censo por llamantes y bajando a las hojas. `agregar` *(y `conNombre`)* no tiene llamantes de producción hasta B7, sólo (7)–(12):
  - **m9** → `TestAgregar_EventIDDistintos`, `/mismo_id_dos_veces`, `/en_otra_identidad` y `/solo_repetidos`;
  - **m10** → `TestAgregar_FechasUTC`;
  - **m11** → `TestAgregar_Marcas` y `/borrada`. El plan decía `/borrada_con_git`, pero la hoja que lo distingue es `/borrada`.
  - **m9 interrumpida antes de correr, por turno con la plataforma; sin resultado.** La orden se paró *(orquestador, 23-10)* con m9 aplicada y `libre.sh` aún
    esperando: la suite no llegó a correr. Revertida por edición inversa: `lista.go` vuelve a `3cce9963…` ✅. m9, m10 y m11 se repiten desde el principio
    cuando la plataforma termine.
- **Repetidas desde el principio**, cuando el dueño dijo «la plataforma ha terminado», con la suite entera tras `libre.sh` en la misma orden:

  | Mutación | Declarado *(y nada más)* | Cayó | Reversión |
  |---|---|---|---|
  | **m9** contar eventos y no `event_id` *(sin `vistos`)* | `TestAgregar_EventIDDistintos`, `/mismo_id_dos_veces`, `/en_otra_identidad`, `/solo_repetidos` | **lo mismo** *(1113 / 4)* | `3cce9963…` ✅ |
  | **m10** fecha sin `.UTC()` | `TestAgregar_FechasUTC` | **lo mismo** *(1116 / 1)* | `3cce9963…` ✅ |
  | **m11** `(sin repositorio)` antes que `(ya no existe)` | `TestAgregar_Marcas`, `/borrada` | **lo mismo** *(1115 / 2)* | `3cce9963…` ✅ |

- **Corte de B4**:
  - suite *(libre 0.49; 15:13:09 → :15)*: **1117 / 0 / 0**;
  - gofmt vacío, vet limpio y lint 0 *(libre 0.27)*;
  - frontera 0 B, ningún `_test.go` existente modificado, `go.mod` sin `require` y sin `go.sum`;
  - `PRIV` 0 en el diff y en los dos ficheros nuevos.

## B5 · La tabla *(`cmd/permea/lista.go`, `internal/project`; sobre `23c0a8f`)* — PARA por `libre.sh`, **cerrada**

- **Antes de B5** *(decisión del 23-10)*: §B1, §B2 y §B3 pasan a `registro-archivo.md`, con una remisión de una línea cada uno.
- **T024** · Esqueletos: `formatear`, `miles` y `aCasa`, en `lista.go`, que devuelven `""`; `HogarReal` → `("", false)` y `CuelgaDe` → `false`, en
  `internal/project/resolve.go`. Compila; vet limpio.
- **T025** · Rojos (13) a (16), escritos en `cmd/permea/lista_formato_test.go`. (13) saca el bloque de spec §Textos por programa. La razón se declaró antes:
  caen 17 nodos, las 14 hojas y los 3 padres con hojas, y ninguna nace verde.
- ⛔ **PARA** *(regla del encargo: «si `libre.sh` espera más de 2 minutos»)*. `timeout 120 bash ~/bin/libre.sh` agotó los 2 minutos **sin ninguna suite en
  marcha**.
  - **Causa, mía.** `libre.sh` busca procesos cuyo comando contenga `go test` y sólo excluye `$$` y `$PPID`. Envuelto en `timeout`, su padre era `timeout`, y
    mi `bash -c`, cuya línea de órdenes contiene «go test», **quedó sin excluir: se esperó a sí mismo**.
  - Es **muy probable que la espera de la primera tanda de m9 de B4** fuera la misma trampa: `libre.sh` corría dentro de `mutacion4.sh`, cuyo abuelo era la
    orden con el heredoc que contiene «go test». **No queda probado**: no se capturó el `pgrep` en ese momento.

- **Cómo se lanza `libre.sh` desde ahora** *(orquestador, 23-10)*: como **hijo directo** de la orden que corre la suite, en segundo plano y sin `timeout` ni guiones
  por medio. Un bucle con `date` vigila desde fuera los 2 minutos: si se pasan, mata `libre.sh` y PARA. **Las ocho esperas de B5 duraron 22–23 s**, sus 3
  muestras, lo que confirma la causa.
- **T025, corridos.** Se esperaba **1117 / 17 / 0** y salió **1117 / 17 / 0**, sin pánicos.
- **T026** · **Verde**, según plan §6:
  - `formatear` escribe las celdas *(`abreviar` 12 + `…`; `aCasa` más la marca; las herramientas con `, `; `miles`; las fechas)*;
  - mide los anchos en runas *(`ancho`)*, separa con dos espacios, pone `EVENTOS` a la derecha, también la cabecera, y deja la última columna sin relleno;
  - escribe una línea en blanco y el pie, con `contar` en singular y en plural; sin filas, sólo el pie;
  - `aCasa` usa `project.CuelgaDe` y `filepath.Rel`;
  - en `internal/project`, `HogarReal` *(`os.UserHomeDir` + `mejorValorDisponible`)* y `CuelgaDe`, un envoltorio de `cuelgaDe`, que **no cambia**:
    `techoDeProduccion` no se toca.

  Suite: **1134 / 0 / 0**.
- **T027** · Censo por llamantes y por hoja:
  - `ancho` sólo lo llama `formatear`;
  - `miles` lo llaman `formatear`, `contar` y (14);
  - `CuelgaDe` sólo lo llama `aCasa`, y a ésta, `formatear` y (15). El `cuelgaDe` del techo no cambia.

  | Mutación | Declarado *(y nada más)* | Cayó | Reversión |
  |---|---|---|---|
  | **m12** ancho en bytes | `TestFormatear_ElEjemploDeLaSpec`, `TestFormatear_Alineacion`, `/tildes` *(el `…` ocupa 3 bytes: cae en (13), como pedía T027)* | **lo mismo** *(1131 / 3)* | `lista.go` `dd5a92e3…` ✅ |
  | **m13** sin punto de miles | `TestFormatear_ElEjemploDeLaSpec`, `TestFormatear_MilesYPie`, `/miles_1000`, `/miles_1234567` | **lo mismo** *(1130 / 4)* | `lista.go` `dd5a92e3…` ✅ |
  | **m14** `CuelgaDe` con `strings.HasPrefix` | `TestACasa`, `/hermano` | **lo mismo** *(1132 / 2)* | `resolve.go` `26d3cf94…` ✅ |

- **El primer corte no salió limpio**:
  - `revive`: `redefines-builtin-id`, la variable `real` de (15/enlace), pasa a `casaReal`;
  - `PRIV` 1: el comentario de `CuelgaDe` llevaba una ruta de ejemplo con forma de directorio personal, y pasa a «`<hogar>2` no cuelga de `<hogar>`».

  Son un test y un comentario, sin conducta: tras ellos, `resolve.go` `acb07a6d…` y `lista_formato_test.go` `04e982e7…`, y se **repitió el corte entero**.
- **Corte de B5**:
  - suite *(16:02:01 → :07)*: **1134 / 0 / 0**;
  - gofmt vacío, vet limpio y lint 0;
  - frontera 0 B, ningún `_test.go` existente modificado, `go.mod` sin `require` y sin `go.sum`;
  - `PRIV` 0 en el diff y en lo nuevo.

- **`~/bin/libre.sh` corregido** *(dueño, 23-10)*, md5 `256ffdf0…` *(antes `ab647c01…`)*:
  - sólo cuenta los procesos que **ejecutan** una suite; nunca `bash`, `sh` o `timeout`, ni los antecesores del guion ni sus herramientas;
  - queda anulada la instrucción provisional del hijo directo, y se puede volver a envolver en `timeout`;
  - la regla sigue: **una suite a la vez**, detrás de `libre.sh`. **Si lista un proceso que no es una suite de verdad, PARA**, y se pega en el informe.

  Desde B6 se guarda su salida entera para mirarlo.

## B6 · El comando *(`cmd/permea/lista.go`, `project.go`, `ayuda.go`; sobre `a6d557c`)*

- **T029** · `case "list"` en `runProject`, hacia `projectList` *(en `lista.go`)*, un esqueleto que sale con 0 sin hacer nada. **Rojos (17) a (21)**, en
  `cmd/permea/lista_test.go`, por el binario de prueba y en `testutil.Sandbox`. Cada texto esperado sale de spec §Textos **por programa** *(`bloqueDeLaSpec`)*.
  - Declarado antes de correr:
    - (17) `sin_sal`: stdout vacío, no el texto; `ilegible`: salida 0, no 1;
    - (18) `x` y `secreto`: salida 0, no el error de uso;
    - (19) `-h` y `--help`: stdout vacío; `en_la_general` y `en_project`: sin la fila;
    - (20) `falta` y `desconocido`: dicen «join» sin `list`;
    - (21) `ilegible`: salida 0. **`plain` nace verde**, porque el esqueleto sale con 0 sin escribir; lo valida m17.
  - Se esperaba **1135 / 16 / 0** y salió **1135 / 16 / 0**. `libre.sh` no listó ningún proceso.
- **T030** · **Verde**:
  - **los textos, generados por un guion que lee spec §Textos**, sin teclear ninguno: la fila de ayuda de `project list` *(índice 3 de `filasDeSubcomandos`)*;
    `ayudasDeSubcomando["project list"]` con `uso: permea project list`, y `"project"` → `[]int{2, 3}`; los dos mensajes de verbos *(`project.go:98`, `:110`)*; y
    `textoSinSal`, `textoSinArgumentos`, `textoConfigIlegible` y `textoSalIlegible`;
  - `projectList`, por orden:
    - `-h` y `--help` dan la ayuda; cualquier otro argumento, el error de uso;
    - después, `RutaDataDir` → `LeerSalt`: si es ilegible, el error con salida 1; sin sal, el texto con salida 0;
    - y `Load`: si es ilegible, el error con salida 1; si no, sale con 0 sin escribir nada.

    **No** pasa por `setup()` ni llama a `CheckRetiredProjectRefMode` *(P-8)*. La tubería la conecta B7.

  Suite: **1151 / 0 / 0**. md5: `lista.go` `890a202e…`, `project.go` `d95bb1f6…`, `ayuda.go` `8ef71c7d…` y `lista_test.go` `bd1ff006…`.
- **T031** · Censo por llamantes y por hoja, leyendo cada test entero:
  - **`projectList`** sólo lo ejercen (17)–(21): ningún test existente ejecuta `project list`;
  - **los mensajes de verbos**, además, `TestProject_ErroresDeUsoDeLaGramatica`, que sólo exige salida 1, stderr no vacío y stdout vacío en sus 3 hojas;
  - **la ayuda de `project`**, `TestAyudaSubcomando_NoHaceNada`, `TestAyuda_AnchoMaximo80` y `TestAyuda_SinJergaInterna`, que no tocan los verbos ni la sal.

  | Mutación | Declarado *(y nada más)* | Cayó | Reversión |
  |---|---|---|---|
  | **m15** `LoadOrCreateSalt` en vez de `LeerSalt` | `TestProjectList_LaSal`, `/sin_sal` *(`/ilegible` sobrevive: `LoadOrCreateSalt` también falla con un directorio)* | **lo mismo** *(1149 / 2)* | `lista.go` `890a202e…` ✅ |
  | **m16** «Verbos disponibles: join» en los dos | `TestProjectList_Verbos`, `/falta`, `/desconocido` | **lo mismo** *(1148 / 3)* | `project.go` `d95bb1f6…` ✅ |
  | **m17** se detiene por `plain` | `TestProjectList_Configuracion`, `/plain` *(con JSON inválido, `CheckRetired…` da `nil`)* | **lo mismo** *(1149 / 2)* | `lista.go` `890a202e…` ✅ |

  **Los `TestAyuda_*` y `TestProject*` existentes siguen verdes sin tocarlos**: en el corte, 50 `TestAyuda_*` y 57 `TestProject*`, sin contar
  `TestProjectList_*`, contando hojas, todos en PASS.
- **Incidencia, arreglada**: `PRIV` daba **1** en `git diff cabf8c5`. Era **mi entrada de §B5 en este registro**, ya en el commit `a6d557c`, que citaba la ruta
  de ejemplo del comentario corregido. El corte de B5 midió `PRIV` antes de escribirla. Se reescribe sin la ruta, y `PRIV` vuelve a 0.
- **Corte de B6**:
  - suite *(17:37:21 → :27)*: **1151 / 0 / 0**;
  - gofmt vacío, vet limpio y lint 0;
  - frontera 0 B, ningún `_test.go` existente modificado, `go.mod` sin `require` y sin `go.sum`;
  - `PRIV` 0 en el diff, en lo no versionado y en el testdata *(`-a`)*.

  `libre.sh` *(con `timeout`)* no listó ningún proceso en ninguna de sus seis esperas.

## B7 · Los seis lectores y la frontera *(`cmd/permea/lista.go`; sobre `c90963c`)*

- **T033** · **Rojos (22) a (26)**, en `cmd/permea/lista_frontera_test.go`, por el binario de prueba.
  - El montaje: `hogarConLista`, con las plantillas y `@CASA@` → `<hogar>`.
    - Dos repositorios: `<hogar>/dev/proyecto-ejemplo`, con `src` y `docs`, y `<hogar>/dev/otro-proyecto`.
    - `<hogar>/pruebas/tmp-1`, sin `.git`, y `<hogar>/dev/viejo`, que no se crea.
    - `sana.db`, en `/sintetico/sana`.

    La tabla esperada se arma a mano: la identidad es `event.Ref(sal, raíz)`, y `formatear`, ya probada en (13), le da forma.
  - Declarado antes de correr, con la tubería vacía de B6:
    - (22) `tabla` y `aviso`: stdout vacío, y en `aviso` sin el aviso de Codex;
    - (23): la lista no enseña ninguna identidad, y `--run` encola 5;
    - (24) `tabla`: stdout vacío. **Nacen verdes** `peticiones`, `conexiones` e `importaciones`, porque no hay red ni importación; las valida m19;
    - (25) `huella_y_tabla/tabla`: stdout vacío. **Nacen verdes** `huella` *(la valida m18)* y `sin_datadir` *(la valida M-B7a, nueva)*;
    - (26): la salida base está vacía.
  - Se esperaba **1156 / 10 / 0** y salió **1156 / 10 / 0**. La precondición de (24) se cumplió: `--run`, en un sandbox gemelo, llega al banco, que
    está en TLS con su certificado por `SSL_CERT_FILE` y cuenta las conexiones con `ConnState`.
- **T034** · **Verde**: la tubería en `projectList`, en el orden del plan:
  1. `resolverRaices`;
  2. `ClaudeCodeLogsRoot` y `FindLogs`;
  3. `recorrer`, con `state.New()`, su `Resolutor` y un sumidero que sólo recoge;
  4. `agregar`, con `res.Origen` y `existeEnDisco` *(nace aquí: `os.Stat` sin `fs.ErrNotExist`)*;
  5. `formatear`, con `project.HogarReal()`.

  El `ingest.Context` lleva la sal leída y la versión; nada de `machine_id`, `dev_id` ni `org_id`. **`internal/ingest/` no se toca.**
  - **Incidencia de montaje, mía**: la primera vuelta dio 1164 / 2. En (24/tabla) faltaba Claude Code, porque el test usaba `escribirConfigCon`, que fija
    `logs_root` en una carpeta vacía, y la lista **lo respeta bien** *(P-8)*. (24) pasa a escribir su `config.json`, con `endpoint` y `device_token` y sin
    `logs_root`. La razón de su rojo no dependía de eso.
  - Suite: **1166 / 0 / 0**. md5: `lista.go` `43ab9a95…` y `lista_frontera_test.go` `46491997…`.
- **T035** · Censo por llamantes y por hoja:
  - `projectList` lo ejercen, por el binario, (17)–(21) y (22)–(26), que he leído enteros;
  - `recorrer`, además, `generate()` y (5) y (6). Ninguna mutación toca `recorrer`.

  | Mutación | Declarado *(y nada más)* | Cayó | Reversión |
  |---|---|---|---|
  | **m18** la lista pasa por `setup()` tras `Load` | `TestProjectList_Configuracion`, `/plain` · `TestListaFrontera_SinEscritura`, `/huella_y_tabla`, `/huella_y_tabla/huella` | **lo mismo** *(1161 / 5)* | `43ab9a95…` ✅ |
  | **m19** drena la cola desde `lista.go` | `TestListaFrontera_SinRed`, `/peticiones`, `/conexiones`, `/importaciones` | **lo mismo** *(1162 / 4)* | `43ab9a95…` ✅ |
  | **m20** `ictx.Resolutor = nil` | `…LosSeisLectores`, `/tabla`, `/aviso` · `…SinRed`, `/tabla` · `…SinEscritura`, `/huella_y_tabla`, `/huella_y_tabla/tabla` | **lo mismo** *(1158 / 8)* | `43ab9a95…` ✅ |
  | **M-B7a** `config.DataDir()` en vez de `RutaDataDir()` | `TestListaFrontera_SinEscritura`, `/sin_datadir` | **lo mismo** *(1164 / 2)* | `43ab9a95…` ✅ |

  Por qué sólo eso:
  - **m18**: `setup()` para por `plain` y crea el `machine_id` que falta; en los demás hay `machine_id`, o la lista sale antes;
  - **m19**: sólo (24) tiene cola y `endpoint`, y `Drain` con la cola vacía no hace nada;
  - **m20**: sin origen, `CARPETA` sale vacía; (23) compara identidades y (26) compara dos salidas igual de rotas.
- **Corte de B7**:
  - suite *(17:52:54 → 17:53:00)*: **1166 / 0 / 0**, con (5) **SC-012 en verde**;
  - gofmt vacío, vet limpio y lint 0;
  - frontera 0 B, ningún `_test.go` existente modificado, `go.mod` sin `require` y sin `go.sum`.

  `libre.sh` no listó ningún proceso en sus ocho esperas.
- **`PRIV`, medido al final**, tras escribir este registro y el informe: **0** en `git diff cabf8c5`, en lo no versionado, en el testdata *(`-a`)*, en el
  registro y en el informe.

## B8 · README y CHANGELOG

## Cierre
