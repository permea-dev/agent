# 013 · Registro de la implementación

**Por qué existe**: como en 007–012, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. **Techo: 250 líneas**, fijado por el
dueño el 23-10 *(`tasks.md` §Presupuesto)*. ⛔ **Ni rutas reales, ni nombres de carpeta, ni contenido** *(PARA 8)*: sólo recuentos, tiempos, tamaños y md5.

- **§B0** *(línea base 1074 / 0 / 0, tamaños de la 0.9.0, la referencia de SC-012 `e6b011db…` y su incidencia de tamaños)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).

- **§B1** *(leer sin escribir: 1084 / 0 / 0; m1–m3 y M-B1a = declarado, con su PARA 2 por censo incompleto y la lección de los llamantes)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).
- **§B2** *(la derivación con origen: 1094 / 0 / 0; SC-012 repetida, byte a byte; m4–m6 = declarado)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).
- **§B3** *(el recorrido común: 1101 / 0 / 0; SC-012 byte a byte; m7 y m8 = declarado, con su PARA 2 por no bajar a las hojas)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).

- **§B4** *(agregación y nombres: 1117 / 0 / 0; m9–m11 = declarado, con la primera tanda de m9 interrumpida sin resultado)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).
- **§B5** *(la tabla: 1134 / 0 / 0; m12–m14 = declarado; la PARA por `libre.sh` y el guion corregido, md5 `256ffdf0…`)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).
- **§B6** *(el verbo: 1151 / 0 / 0; m15–m17 = declarado; `PRIV` de §B5 corregido)*: archivado íntegro el 2026-10-23 en [`registro-archivo.md`](./registro-archivo.md).

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

## B8 · README y CHANGELOG *(sobre `fee77c2`)*

- **B8a** *(`fee77c2`)*: §B4–§B6, al archivo; las propuestas de texto, en el informe, sin tocar los ficheros.
- **T037** · **El dueño aprueba los textos de B8a tal cual** *(2026-10-23)*: la entradilla de `## Comandos`, la sección `### permea project list` y `## 0.10.0 —
  PENDIENTE`. Se copia el informe de B8a a `tmp/b8a-propuestas.md` *(md5 `cea5b2af…`)*, y de ahí salen los textos.
  - **Rojo**, declarado en el encargo: `grep -c '^## 0.10.0' CHANGELOG.md` → **0**; `grep -c 'project list' README.md` → **0**.
  - **Escritos por programa**:
    - los tres bloques se extraen con `awk` de `tmp/b8a-propuestas.md`, y antes se comprueba con `cmp` que son las propuestas;
    - la entradilla sustituye a `README.md:85-90`, la lista y la frase de HTTPS; las líneas 83–84 no cambian en la propuesta;
    - la sección va tras la de `project join` y antes de `## Modos de ejecución`;
    - la entrada, antes de `## 0.9.0`.
  - **Verde**:
    - `grep` → **1** y **4**;
    - `cmp` sin diferencias de los tres bloques contra lo escrito;
    - `cmp` sin diferencias del ejemplo del README, sin la sangría, contra el bloque «La salida» de spec §Textos *(8 líneas)*.
  - Puertas y `PRIV`: en el informe de B8.

## Cierre *(en tramos)*

### C1 · Puertas *(T038, 2026-10-23, sobre `752c217`)*

`libre.sh` corrió antes de cada suite y del lint, y no listó nada.

| Puerta | Resultado |
|---|---|
| `gofmt -l .` · `go vet ./...` · `golangci-lint run` *(2.15.0)* | vacío · limpio · **0 issues** |
| `go test -count=1 -v ./...` | **1166 PASS · 0 FAIL · 0 SKIP**, 10 paquetes `ok` |
| **Los 1074 de `cabf8c5`, por nombre** | `git archive cabf8c5` en un temporal del scratchpad, sin worktree: su suite, **1074 / 0 / 0**. Contra la rama: **faltan 0**, y **92 son nuevos**, todos de los tests de la 013 |
| **SC-015** | frontera FR-025, **0 B**; 0 `_test.go` existentes modificados; `go.mod` sin `require`; sin `go.sum` |
| **SC-016** | ≤ 51 200 B en las 5 plataformas *(abajo)*. **Compilan Windows y darwin** |
| **SC-017** | `PRIV` 0 en `git diff cabf8c5` y en `testdata/lista/` *(`-a`)*, medido al final |

**SC-016.** La 0.9.0 *(`git archive cabf8c5`, `-X main.version=0.9.0`)* y la rama *(`git archive HEAD`, `-X main.version=0.10.0`)* se compilaron **en esta misma
sesión**: go1.22.2, `CGO_ENABLED=0`, `-trimpath -ldflags "-s -w …"` y sin líneas `vcs.*`.

| Plataforma | 0.9.0 *(B)* | Rama *(B)* | Δ *(B)* | Δ *(KB)* |
|---|---:|---:|---:|---:|
| linux/amd64 | 5 718 168 | 5 738 648 | +20 480 | 20,0 |
| linux/arm64 *(fichero)* | 5 439 640 | 5 439 640 | 0 | 0,0 |
| linux/arm64 *(Σ `LOAD`, 010 E-3)* | 5 398 044 | 5 419 484 | +21 440 | 20,9 |
| darwin/amd64 | 5 881 792 | 5 907 008 | +25 216 | 24,6 |
| darwin/arm64 | 5 638 738 | 5 655 938 | +17 200 | 16,8 |
| windows/amd64 | 5 914 112 | 5 938 176 | +24 064 | 23,5 |

La mayor es **+25 216 B, el 49 % del tope**. La 0.9.0 da las mismas cifras que en B0, también medidas esta sesión. Los árboles y los binarios temporales se
borraron.

### C2 · SC-018, la medida sobre la instalación de WSL *(T039, 2026-10-23, sobre `eb15b14`)* — **D-013-W, estricto**

- **El binario**: `git archive HEAD` en un temporal del scratchpad, con `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=0.10.0"`; `--version` →
  `0.10.0`. El temporal se borró.
- **Cómo se midió**:
  - `<binario> project list`, con stdout a `wc -l`, stderr a `wc -c` y el código aparte;
  - la huella del `DataDir`: el md5 de la lista ordenada de «md5 + tamaño + `mtime` en ns» de cada fichero, **sólo en variables**.

  No se imprimió ninguna ruta, nombre ni identidad, ni se abrió ningún fichero de las herramientas.

| Medida | Resultado |
|---|---|
| tiempo | **0,713 s** *(SC-018: ≤ 2 s)* |
| líneas de stdout | **6** *(cabecera, 2 filas, la línea en blanco y las 2 del pie)* |
| bytes de stderr | **0** |
| código de salida | **0** |
| huella del `DataDir` | **IGUAL** antes y después *(5 ficheros, y 5 después)* |

**SC-018, cerrado** *(dueño, 23-10)*: el recuento independiente del mismo día es el de desc. Q4 *(23-10)*, **2 identidades** en WSL, y la lista da **2 filas**.

### C3 · snapshot *(SC-013; T039, 2026-10-23)* — PARA por `goreleaser`, **cerrada**

- **1.ª vuelta, PARA**: `goreleaser` no estaba instalado *(ni en el `PATH`, ni en `~/go/bin`, `~/bin` o `/usr/local/bin`, ni en la caché de módulos)*, e
  instalarlo pedía red.
- **El dueño lo instala** *(23-10)*: en `~/.local/bin`, **2.18.3**, bajado de la página de versiones de GitHub y con su sha256 comprobado. `goreleaser --version` →
  `GitVersion: 2.18.3`.
- **`GOPROXY=off GOTOOLCHAIN=local goreleaser release --snapshot --clean`** → **rc 0**, versión **`0.9.0-SNAPSHOT-eb15b14`**.
  - La versión sale de la última etiqueta, `v0.9.0`, como `0.8.0-SNAPSHOT-8809c58` en la 012.
  - El único gancho es `go mod tidy`, y no pidió red.
  - La publicación, el tap y el bucket se saltan en el modo snapshot.
- **`strings`**: 10 agujas sacadas por programa de spec §Textos.
  - Son las 5 líneas de la fila de ayuda, `uso: permea project list`, los 2 mensajes de verbos *(con `<verbo>` → `%q`)* y las 2 frases del pie *(«Fechas en
    UTC, como en el panel.» y la segunda línea)*.
  - Se buscan con `LC_ALL=C grep -aF`.

  | Binario | `strings -e s` | `strings -e S` |
  |---|---:|---:|
  | linux/amd64 del snapshot | 6 de 10 | **10 de 10** |
  | `permea.exe` del zip de Windows | 6 de 10 | **10 de 10** |

  Las 4 que no da `-e s` son **justo las que llevan caracteres no ASCII** *(Enseña, Sólo, instalación, también…)*: `-e s` sólo lee 7 bits y las corta.
- **Incidencia de método, mía**: con un `grep` sin `LC_ALL=C -a`, `-e S` dio 0 de 10, porque la salida con bytes de 8 bits se trataba como binaria o inválida. Con
  el `grep` de bytes, 10 de 10. El binario no cambió.
- **sha256**: el del zip de Windows *(`6f7e7cd3…`)* = el de `checksums.txt`, y `sha256sum -c` da OK en los 5 artefactos.
- **Tamaños**: el `.exe` mide 5 938 176 B y el de linux/amd64, 5 738 648 B, iguales a los de la rama en C1.
- **Al acabar**: `dist/`, que git ignora, se borró, igual que la copia del `.exe`. `GOPROXY=off go mod tidy` → rc 0, `go.mod` con el mismo md5 y sin `go.sum`.


### C4 · W1, el cruce con el panel *(SC-019; T040, hecho por el dueño el 2026-10-23, binario de la rama)*

| Instalación | Tiempo *(Q-2)* | Identidades | En el panel | Eventos y fechas |
|---|---:|---:|---|---|
| **WSL** *(binario de la rama)* | **0,41 s** | **2** | las 2, ya asignadas a su proyecto | **iguales** que el panel, en las 2 |
| **Windows** *(`permea.exe` de la rama, en PowerShell)* | **2,09 s** | **3** | 1 asignada · **2 sin asignar** | **iguales** que el panel, en las 3 |

- **SC-019 cumple: 5 de 5.** Entre las «Sin asignar» del panel, **solo una más** tiene la última fecha ≥ 20-10, y no sale en ninguna de las dos listas: es del equipo
  viejo *(el traslado del 20-10)*, con otra sal. Es la limitación «otras instalaciones» de spec §Limitaciones.
- **Las tildes en Windows**: escritas directamente en la consola, se ven bien *(«también», «aún», «…»)*. **Capturadas por una tubería de PowerShell 5.1**
  *(`| Out-File`)*, salen con caracteres rotos, porque PowerShell decodifica con la página de códigos de la consola. **No es de la 013**: pasa igual con cualquier
  texto con tildes del agente. README y CHANGELOG, aprobados, **no cambian**.
- **Observación del dueño, para después de esta versión**: una ruta muy larga *(p. ej., la de una carpeta temporal)* ensancha la columna CARPETA de todas las
  filas. Candidata: recortar la carpeta por el medio con «…» a partir de un ancho. *(Este repo no tiene backlog propio: va en el informe.)*

### C5 · La fecha y el cuerpo del PR *(T041, 2026-10-23)*
- `CHANGELOG.md`: `## 0.10.0 — PENDIENTE` → `## 0.10.0 — 2026-10-23`, y nada más.
- El cuerpo del PR, **fuera del repo**: `tmp/pr-013.md`.

### C6 a C8 · Fusión, etiqueta y canales *(T042–T044, hechos por el dueño el 2026-10-23)*

- **C6**: el PR #10, fusionado con merge commit, `0e1b342`. Sin checks de CI, que es lo normal en este repo: sólo `release.yml` corre con la etiqueta. La rama,
  borrada en local y en remoto.
- **C7**: la etiqueta **`v0.10.0`**, anotada, sobre `0e1b342`; el workflow `release`, en `completed · success`.
- **C8**, descargado por el dueño en un temporal:
  - 5 archivos y `permea_0.10.0_checksums.txt`, con `sha256sum -c` **OK ×5**;
  - el binario de linux/amd64: `--version` → `0.10.0`, y «uso: permea project list» ×1 y «Fechas en UTC, como en el panel.» ×1;
  - **Scoop** *(`scoop-permea/permea.json`)*: `0.10.0`, con el hash del zip de Windows igual al de las sumas;
  - **el cask** *(`homebrew-permea/Casks/permea.rb`)*: `0.10.0`, con sus 4 `sha256`.
- **Las instalaciones del dueño, en 0.10.0**:
  - **WSL**, por `install.sh`, con el checksum verificado. La primera vez dio «Text file busy», porque `permea.service`, de systemd de usuario, estaba en marcha. Se
    paró, se instaló y se arrancó, y queda `active (running)`;
  - **Windows**, por `scoop update permea`, con el hash OK.
- **`goreleaser` 2.18.3**, instalado en WSL en `~/.local/bin` para C3. En el equipo nuevo sólo estaba en Windows, por scoop.
- **La 013, cerrada**: spec §Cierre, con los 19 SC ✅.

