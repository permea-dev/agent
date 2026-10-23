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

