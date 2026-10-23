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
  - `PRIV` 1: el comentario de `CuelgaDe` decía «/home/ana2», y pasa a «`<hogar>2` no cuelga de `<hogar>`».

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

## B6 · El comando

## B7 · Los seis lectores y la frontera

## B8 · README y CHANGELOG

## Cierre
