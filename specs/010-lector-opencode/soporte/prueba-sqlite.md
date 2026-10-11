# 010 · Prueba de SQLite: A *(modernc)* frente a B *(lector propio)*

**Fecha**: 2026-10-11 *(prueba en WSL, sobre `4c6cafc`, y ensayo en Windows)* · **Decisión del dueño**: **B** *(11-10; D-3 de la spec)*.

**Cómo se lee esto.**
- **MEDIDO**: en WSL, con Go 1.22.2 y `sqlite3` 3.45.1. El código vivió en un directorio de trabajo no versionado, como **módulos Go propios**, sin tocar el
  `go.mod`, `internal/` ni `cmd/` del agente.
- **ENSAYO**: en Windows, con OpenCode 2.0.26 escribiendo. Lo ejecutó el dueño, y el orquestador leyó los logs.

> **Privacidad.** Sólo recuentos, sumas, tamaños y md5 del código de la prueba. Las sondas nunca imprimen ids, títulos, textos ni rutas, y sus errores se
> limpian de rutas. La copia congelada se abrió sólo en lectura y su sha256 se comprobó al empezar y al terminar. Las huellas del contenido *(sha256 de `data`)*
> se compararon y **no se publican**, como en 009.

## Las dos sondas

Las dos tienen la misma interfaz: `sonda <opencode.db>` imprime una línea JSON con lo siguiente:
- filas `assistant` y filas completadas *(`$.time.completed`)*;
- sobre las completadas, Σ de `tokens.input`, `.output`, `.reasoning`, `.cache.read` y `.cache.write`, y de `cost`;
- los totales de `session_v2`;
- filas de `session_message` y el sha256 de su `data`, en orden de `session_id, seq`;
- `ms` y `error`.

- **A** · `modernc.org/sqlite` **v1.36.1** *(SQLite 3.49.1)*, con `database/sql`:
  - abre `file:…?mode=ro`, y con `--immutable` añade `immutable=1`;
  - las sumas las hace en SQL, con `json_extract` y `total()`.
- **B** · **lector propio**, sólo con la librería estándar:
  - `lector.go`: **525 líneas, 472 de código**;
  - `main.go` de la sonda: 250;
  - las sumas las hace en Go *(`encoding/json`; el coste, con suma compensada, como `total()`)*.

## Tamaños *(MEDIDO; bytes; `CGO_ENABLED=0 -trimpath -ldflags "-s -w"`, las banderas de `.goreleaser.yaml`)*

El aumento se midió **enlazando cada lector en el propio agente** *(una copia de `HEAD` hecha con `git archive`, fuera del repo)*. El lector va detrás de un
`init` que no se ejecuta, para que el enlazador no lo elimine.

| Plataforma | Agente | **+A** | **+B** |
|---|---:|---:|---:|
| darwin/amd64 | 5 714 464 | +3 822 960 *(+66,9 %)* | **+62 928** *(+1,1 %)* |
| darwin/arm64 | 5 486 658 | +3 691 792 *(+67,3 %)* | **+51 008** *(+0,9 %)* |
| linux/amd64 | 5 558 424 | +3 764 224 *(+67,7 %)* | **+61 440** *(+1,1 %)* |
| linux/arm64 | 5 308 568 | +3 670 016 *(+69,1 %)* | **+65 536** *(+1,2 %)* |
| windows/amd64 | 5 748 736 | +3 867 648 *(+67,3 %)* | **+62 464** *(+1,1 %)* |

Frente a un «hola mundo» con las mismas banderas, las sondas sueltas añaden de 4,5 a 4,9 MB *(A)* y de 0,50 a 0,64 MB *(B)*. Esta medida exagera: la
sonda arrastra `encoding/json` y `crypto/sha256`, que el agente ya tiene.

## A frente a B *(MEDIDO)*

| | **A** · modernc | **B** · propio |
|---|---|---|
| Aumento del binario | +3,7–3,9 MB *(≈ +67 %)* | **+51–66 KB** *(≈ +1 %)* |
| `go.mod` | 1 `require` directo; `go list -m all` = 28 módulos, **9 enlazados** | **sin `require`** |
| Licencias | BSD-3, MIT *(musl incluido)*, dominio público *(SQLite)*; Apache-2.0 sólo en el grafo, no enlazado | — |
| Go | v1.36.1 es la última que acepta Go 1.22; la última publicada, v1.60.1, pide **Go ≥ 1.26** | Go 1.22 |
| cgo | compila en las 5 con `CGO_ENABLED=0` | ídem |
| ¿Escribe? | **sí**: reescribe el `-shm`; sin `-wal`, **crea** `-wal` *(0 B)* y `-shm` y los deja | **nunca** *(md5 de los tres ficheros iguales)* |
| Coherencia frente a un escritor | el protocolo de SQLite *(bloqueos en el `-shm`)* | sin bloqueos: **comprobación de estabilidad** y reintento *(§Lo que cubre B)* |
| Exactitud *(C, I, R, S)* | igual a `sqlite3` en los 4 | igual a `sqlite3` en los 4 |

**`immutable` no ve el WAL** *(MEDIDO)*: A con `immutable=1` sobre la instantánea **I** da 1 sesión y 5 filas, en vez de 3 sesiones y 55 filas, aunque el
`-wal` esté al lado. ⇒ «Copiar la base y abrirla `immutable`» *(descubrimiento, P-10)* **no sirve si hay WAL**.

## Exactitud *(MEDIDO)*

**Las cuatro bases**:
- **C**: la copia congelada *(sin `-wal`; el `data` mayor mide 1 688 B, así que no hay desbordamiento)*.
- **I**: la instantánea sintética *(base, `-wal` y `-shm`)*.
- **R**: I con **un bit cambiado** en una hoja de tabla de la última transacción confirmada, sin `-shm`.
- **S**: la base sintética **sola**, tras cerrarla.

**Cómo se hizo la sintética** *(con `sqlite3`, en un `mktemp -d` que se borró al acabar)*:
1. **Estado 0**: modo WAL, `session_v2` y `session_message` con **las columnas de la 2.0.26**, y 1 sesión con 5 mensajes, ya volcada a la base.
2. **Una conexión se queda abierta**, para que nadie vuelque el WAL.
3. **Tres transacciones confirmadas**, con `wal_autocheckpoint=0`:
   - **T1**: una sesión con 20 mensajes, 4 de ≈ 10 KB, que **desbordan**;
   - **T2**: **reescribe en su sitio** filas de T1, como hace OpenCode, y agranda una a ≈ 12 KB;
   - **T3**: una sesión con 30 mensajes, 6 de ≈ 20 KB.
4. **Una transacción a medias**: `cache_size=10`, 2 000 filas de ≈ 2 KB y **sin commit**, así que vuelca frames al WAL sin confirmarlos.
5. **La instantánea**: copia de `db`, `-wal` y `-shm`.
   - 519 frames en el fichero, **93 en cadena válida** y 3 commits: **73 confirmados** y 20 tras el último commit.
   - SQLite **reescribe en su sitio** los frames aún no confirmados y no rehace las sumas siguientes hasta el commit. Por eso la cadena se corta en el 93.

**Cada lector leyó su propia copia**, con md5 de cada fichero antes y después.

| Base | `sqlite3` *(referencia)* | A | B |
|---|---|:--:|:--:|
| **C** | 8 `assistant`, **8 completadas** · **59 321 / 230 / 462 / 3 999 / 0 / 0,047385675** · `session_v2` *(1)*: **59 859 / 235 / 462 / 3 999 / 0 / 0,047559575** · 14 filas | = | = |
| **I** | 44 · 39 · 46 873 / 4 699 / 40 / 5 600 / 88 / 0,052926000 · **3 sesiones** *(la «a medias» no aparece)* · 55 filas | = *(¹)* | = |
| **R** | 20 · 18 · 15 126 / 1 518 / 18 / 2 200 / 36 / 0,017104500 · **2 sesiones** *(T3 se descarta por la suma rota)* · 25 filas | = *(¹)* | = |
| **S** | las cifras de I · 3 sesiones · 55 filas | = *(¹)* | = |

- *(¹)* Las mismas cifras, pero **escribe**: en I reescribe el `-shm`; en R lo crea; en S crea `-shm` y un `-wal` de 0 B.
- **El sha256 de la concatenación de `data` es igual** en `sqlite3`, A y B, en las cuatro bases. En C, también con los `.exe` de Windows. En la sintética
  cambia de una ejecución a otra *(relleno aleatorio)*, y en cada una coincidió.
- **ms**: en C, A 1 y B 0; en I, B 5–10, con 73 frames y 0 reintentos.

## Mutaciones de B *(declaradas antes de ejecutarlas)*

- **Cada una es un cambio de una línea** con `perl`; el patrón tiene que casar exactamente una vez.
- Tras compilar, se restaura el fichero y se comprueba la md5.
- md5 del `lector.go` probado: **`f6b47fa8…`**; igual tras cada restauración.

| | Cambio | Declarado: cae en | Observado | md5 mutado |
|---|---|---|---|---|
| **mB1** ignorar el WAL | `if cab[18] == 2 \|\| cab[19] == 2 {` → `if false {` | I, R | **I y R**: sólo el estado 0, y además `error` *(la estabilidad ve un `-wal` que no leyó)*; C y S, iguales ✓ | `0e906084…` |
| **mB2** no comprobar la suma | `if s0 != …fr[16:]… {` → `if false {` | R | **R, en silencio**: acepta T3 con el bit roto *(`cache_read` 64 813 en vez de 64 812)*; C, I y S, iguales ✓ | `ffc60fc7…` |
| **mB3** no seguir el desbordamiento | `if local == total {` → `if true {` | I, R, S | **I, R y S**: `base corrupta o truncada`; **C no cambia** ✓ | `1923e05f…` |
| **mB4** frames tras el último commit | `cadena[:ultimoCommit]` → `cadena[:len(cadena)]` | I | **I**: aparece la transacción a medias *(4 sesiones, 61 filas)*; C, R y S, iguales ✓ | `f069afc8…` |

**Las 4 caen donde se declaró y sólo ahí.**
- **mB2 sólo la detecta R**: con un WAL sano, quitar la suma no cambia nada.
- **mB3 no la detecta la copia real**, porque no tiene ninguna fila que desborde.

⇒ Los tests de la implementación necesitan **una base con un frame roto** y **filas que desborden**.

## Lo que cubre B

- **Cabecera**:
  - firma; tamaño de página *(también 65 536)*; bytes reservados;
  - codificación *(sólo UTF-8)*;
  - páginas de la base: las de la cabecera si son válidas, y si no, por el tamaño del fichero.
- **`sqlite_schema`**: la raíz y el `CREATE TABLE` de cada tabla. Las columnas salen del SQL, con comillas y restricciones de tabla. También el alias de
  rowid `INTEGER PRIMARY KEY`.
- **Árbol de tabla**:
  - interiores *(0x05)* y hojas *(0x0D)*, en orden de rowid;
  - con defensa ante ciclos, profundidad > 64, punteros fuera de página y páginas fuera de la base.
- **Registros**: varints de 1 a 9 bytes y todos los tipos de serie. **Desbordamiento**: la parte local *(U−35, M, K)* y la cadena, con tope de saltos y de
  tamaño.
- **WAL**, sólo si la cabecera dice modo WAL:
  - cabecera: magia LE/BE, versión, tamaño de página y suma;
  - cada frame: **sales** y **suma acumulada** en cadena, que se corta en el primer fallo;
  - se aceptan **sólo los frames hasta el último commit válido**, y **el más reciente por página**;
  - el tamaño de la base pasa a ser el del commit;
  - un WAL inválido se ignora entero, como en SQLite;
  - el `-shm` no se lee.
- **Sin `-wal`**: la base sola.
- **Nunca escribe**: base con `O_RDONLY` y `-wal` con `ReadFile`.
- **Estabilidad**:
  - al acabar, comprueba que el `-wal` *(tamaño y cabecera, con las sales)* y el tamaño de la base no han cambiado;
  - si cambiaron, repite, hasta 5 veces *(`reintentos`)*.
  - Un checkpoint sólo vuelca a la base páginas que B ya toma del WAL.

## Lo que NO cubre B

- **Índices**: no los usa, y recorre la tabla entera *(en el ensayo, 6–22 ms)*.
- **Tablas `WITHOUT ROWID`** y **texto UTF-16**: se detectan y dan error.
- **`DEFAULT` de columnas añadidas con `ALTER TABLE ADD COLUMN`**: en filas antiguas salen NULL.
- **Los bloqueos de SQLite**: los sustituye la estabilidad.
- **Un diario de rollback «caliente»** *(`-journal`)*.
- **Autovacuum**: no probado.
- **Formatos de esquema 1–3**: no probados.
- **El WAL entero va a memoria.**
- **Un esquema distinto al de la 2.0.26**: no se ha probado.

## Ensayo en Windows *(ENSAYO; 2026-10-11, OpenCode 2.0.26 escribiendo)*

Un bucle de PowerShell lanzó `sonda-a.exe` y `sonda-b.exe` cada 2 s contra la base viva *(sin `--immutable`)*.

| | Ventana | Pares | Resultado |
|---|---|---:|---|
| **1** | 10:03:52–10:06:50 | 87 | **A = B en 87/87**, sin errores ni reintentos. **El WAL se reinició** durante el ensayo *(1 490 → 33 frames, checkpoint de OpenCode)* y B lo leyó bien. Sólo entraron una sesión nueva y la llamada de su título *(+545 / +8 en los totales)*; las respuestas llegaron después de la ventana |
| **2** | 10:24:49–10:25:28 | 13 | **A = B en 13/13**. Dos lecturas lentas al arrancar OpenCode *(A, 5 378 ms; B, 3 615 ms)*, que no se repitieron |
| **3** | 10:26:22–10:28:58 | 76 | **A = B en 76/76**, sin errores ni reintentos. ms: A, mediana 1 y máximo 4; B, mediana 6 y máximo 22. **A las 10:26:56 las dos vieron 18 filas `assistant` con 17 completadas** *(una respuesta a medio escribir, excluida)*, y a las 10:26:58, 18 de 18. El WAL creció de 36 a 264 frames mientras OpenCode escribía |

⇒ **OpenCode 2.0.26 usa WAL** y hace checkpoints mientras está abierto.
- B lee bien **antes y después de un reinicio del WAL**.
- La regla «sólo filas completadas» *(P-1)* **funciona en vivo**.
- La comprobación de estabilidad **no tuvo que repetir ninguna lectura** en 176 pares: el reintento **no llegó a ejercitarse en vivo**.
