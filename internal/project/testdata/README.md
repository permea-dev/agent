# Fixtures de derivación de identidad de proyecto (P-004 · T003)

Registros JSONL con la forma que produce Claude Code, uno por **familia** de casos del contrato
`specs/004-identidad-de-proyecto/contracts/project-identity.md`.

## Rutas parametrizables

Todos los `cwd` llevan el marcador **`{{BASE}}`**, que el consumidor sustituye por el directorio
temporal donde ha construido el árbol (`t.TempDir()` en un test, `$(mktemp -d)` en una validación
manual). Sin sustituir, el fixture no apunta a nada real — **es deliberado**: una ruta fija dentro de
un fixture solo sería válida en la máquina de quien la escribió.

## Los ficheros y qué familia cubre cada uno

| Fichero | Familia | Casos del contrato |
|---|---|---|
| `raiz-subdirectorio.jsonl` | eventos en la raíz y en un subdirectorio del mismo proyecto, más un proyecto distinto como contra-prueba | 2, 3, 6 |
| `worktree.jsonl` | repositorio principal y árbol de trabajo paralelo; dentro del paralelo, raíz y subdirectorio | 4, 5 |
| `sintacticos.jsonl` | la misma ruta con barra final, `./` y `..` redundantes, más una ruta genuinamente distinta con prefijo común (`suelto` vs `sueltoo`) | 9, 10 |
| `enlaces.jsonl` | enlace hacia el interior de un proyecto junto a su ruta real, y enlace hacia un directorio suelto | 11 |
| `degradados.jsonl` | inexistente, enlace roto, sin permisos, ruta relativa y `cwd` vacío | 1, 12, 13, 14, 15 |

**Los casos 7, 7b y 8** (techo del directorio personal y de la raíz del sistema) **no tienen fixture
aquí**: se prueban por entorno (`HOME` falso) y por inyección del techo, no leyendo un log. Un fixture
no puede expresar «este directorio es el home».

## Consumidores — y quién NO los consume

| Consumidor | Uso |
|---|---|
| **T004 / T005** — golden de frontera ampliado | Como entrada adicional del golden. **Su fixture propio es otro**: `internal/ingest/testdata/boundary_sample.jsonl`, con los fragmentos de ruta de la denylist, y **se crea en T004, no aquí** |
| **V1–V6** — validaciones manuales de `quickstart.md` | Sustituyendo `{{BASE}}` por el temporal del escenario |

**Los tests unitarios del resolutor (T008..T020) NO consumen estos ficheros.** Construyen sus árboles
con `t.TempDir()` y llaman al resolutor directamente con el salt explícito: no necesitan pasar por el
parseo del log, y un fixture en disco no podría expresar un `HOME` falso ni un worktree real.

## Lo que estos fixtures NO son

**No sustituyen a `internal/ingest/testdata/claude_code_sample.jsonl`**, que quedó **congelado en
T001** como conjunto de referencia de `baseline-sc004.tsv`. Ese fichero **no se toca en toda la
feature**: ampliarlo cambiaría el conjunto de la línea base y rompería la comparación de neutralidad
de T007 y la regresión-cero de V8 por la razón equivocada — parecería que cambió la derivación cuando
lo que cambió fue la entrada.
