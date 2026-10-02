# Contrato — Ayuda y errores de uso de la CLI (P-006)

**Feature**: `006-medicion-fiel` | **Fecha**: 2026-10-02 | **Requisitos**: P-006 FR-021 a FR-024

**Relación con contratos existentes.** Cumple `specs/003-enrolamiento/contracts/cli.md` y
`specs/005-adhesion-a-proyecto/contracts/cli.md`, y **no** los redefine:
- los flags `--scan`, `--run`, `--daemon` y `--version` conservan su comportamiento;
- `project` sin verbo y `project <verbo desconocido>` siguen siendo error de uso con exit 1
  (005 `cli.md:31-32`).

Lo nuevo es **cómo se pide ayuda** y **qué pasa con un subcomando inexistente**. Códigos de salida:
siguen siendo **0 y 1** (D-005-4).

---

## La ayuda general

| Invocación | stdout | stderr | Exit |
|---|---|---|---|
| `permea` (sin argumentos) | la ayuda general | **vacío** | **0** |
| `permea help` | la ayuda general | **vacío** | **0** |
| `permea -h` · `permea --help` · `permea -help` | la ayuda general | **vacío** | **0** |
| `permea --run -h` (y cualquier flag de ingesta con `-h`) | la ayuda general, **sin ejecutar** la ingesta | vacío | 0 |

- Las cuatro primeras filas producen salidas **idénticas byte a byte** (SC-012). **Sin banner**: el
  banner `Permea <versión>` es de los modos de ingesta y no forma parte de la ayuda.
- **Contenido mínimo**:
  - los subcomandos `enroll`, `status` y `project join`, y las opciones `--scan`, `--run`, `--daemon`
    y `--version`;
  - la vía **stdin** recomendada para `enroll` y para `project join` (003 `cli.md:21`, 005 `cli.md:60`).
- **Una sola fuente**: la ayuda de cada subcomando es un fragmento de la general, compuesto de la misma
  tabla.
- ~~**Flag desconocido** (`permea --bogus`): fuera de la spec, y se conserva lo que hace Go. El error va
  por **stderr**, con exit **2**, y la ayuda que lo acompaña sale ahora por stdout.~~ **Revocado el
  2026-10-02 por E-006-P3** (decisión del orquestador): ver §Opción desconocida.

## Opción desconocida *(añadido el 2026-10-02, E-006-P3)*

| Invocación | stdout | stderr | Exit |
|---|---|---|---|
| `permea --bogus` (opción que no existe) | **vacío** | un error que **nombra la opción** y remite a `permea help` | **2**, como hoy |
| `permea --scan` sin valor (opción que exige argumento) | **vacío** | ídem, nombrando la opción | **2** |

- **Nada por stdout**: ni la ayuda ni el uso por defecto de Go. Quien quiera la ayuda, la pide.
- Los **argumentos sobrantes tras las opciones** (`permea --run extra`) siguen como hoy: se ignoran.
- El texto exacto lo fija la tarea. El contrato fija el canal, el código, que se nombre la opción y la
  remisión a `permea help`.

## Las ayudas de subcomando

| Invocación | stdout | Exit | Efectos |
|---|---|---|---|
| `permea enroll -h` · `--help` | ayuda de `enroll` | 0 | **ninguno**: no lee stdin |
| `permea status -h` · `--help` | ayuda de `status` | 0 | **ninguno**: no crea el directorio de datos ni lee la configuración |
| `permea project -h` · `--help` | ayuda de `project` (lista `join`) | 0 | ninguno |
| `permea project join -h` · `--help` | ayuda de `project join` | 0 | **ninguno**: no lee stdin, no lee la configuración, no crea el `salt` y **no emite ninguna petición** |

«Ninguno» se comprueba así (SC-013):
- HOME de sandbox **vacío**, con el árbol idéntico antes y después;
- agente de prueba enrolado contra un servidor de prueba que recibe **0** peticiones;
- ejecución desde dentro de un árbol de proyecto.

Sólo se reconoce `-h`/`--help` como **primer** argumento del subcomando. Ningún *enrollment string*
(`pmea2.`) ni código de adhesión (`pmeaj1.`) válidos empiezan por `-h`, así que no se pierde ningún
valor legítimo (spec, caso límite).

## Subcomando inexistente

| Invocación | stdout | stderr | Exit |
|---|---|---|---|
| `permea <x>`, si `<x>` no empieza por `-` y no es `enroll`, `status`, `project` ni `help` | vacío | `error: subcomando desconocido "<x>"…`, con los subcomandos disponibles | **1** |
| ídem, si `<x>` empieza por `pmea2.`, `pmeaj1.` o `pmea1.` | vacío | el mismo error **sin reproducir `<x>`** | **1** |

El texto exacto lo fija la tarea. El contrato fija el canal, el código, que se nombre lo tecleado y la
excepción de los secretos (FR-022, D-006-9).

## Invariante

**El token NUNCA se imprime**, ni el *enrollment string*, ni el código de adhesión, en ninguna de las
invocaciones de este contrato ni en sus errores (FR-024). Se comprueba con un token **centinela** en
un `config.json` de prueba y la búsqueda literal en stdout y en stderr, **capturados por separado**
(disciplina 7 de 005).
