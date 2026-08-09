# Verdad-terreno · SC-001 — las 12 identidades medidas y su predicción

**Fecha**: 2026-08-09 · **Aportada por**: Basilio (T035) · **Formato**: research.md §P7
**Fuente de las identidades**: base de desarrollo de la plataforma (12 refs, 643 eventos,
2026-07-13 → 2026-08-05). **Fuente de los directorios**: logs de Claude Code supervivientes
(~/.claude/projects, íntegros desde 2026-07-13) + memoria del desarrollador.

## Contexto declarado
- Una sola máquina, un solo usuario del SO, un solo salt en todo el periodo (pregunta 3: sí).
- Nunca se lanzó Claude Code desde clones de prueba, worktrees ni directorios sin repositorio
  (pregunta 2: no). Los 18 cwd de los logs caen todos bajo dos raíces con `.git`.
- El salt original no se conserva (no hay instalación de permea en la máquina), así que el
  casado ref↔directorio es por magnitud + rango de fechas + ranking de logs, no por recálculo.
  La confianza se declara por fila; para la predicción es irrelevante: toda fila cae bajo una
  de las dos raíces, sea cual sea su directorio exacto.

## Las 12 filas

| identidad (prefijo) | eventos | fechas | directorio de origen | naturaleza | confianza | identidad esperada |
|---|---|---|---|---|---|---|
| daad022cc2 | 330 | 07-15→07-25 | /home/bfgnet/dev/permea-platform | raíz de proyecto | alta | **A** |
| 3f3714b98c | 128 | 07-13→08-05 | …/permea-platform/frontend | subdirectorio de A | alta | **A** |
| d33a8d913f | 92 | 07-18→08-05 | …/permea-platform/backend | subdirectorio de A | alta | **A** |
| 167c141674 | 61 | 07-13→08-05 | /home/bfgnet/dev/agente | raíz de proyecto | alta | **B** |
| ede3828798 | 8 | 07-16→17 | …/specs/P003b-auth-onboarding (probable) | subdirectorio de A | media | **A** |
| 6d68c40076 | 7 | 07-26→08-05 | …/design (probable) | subdirectorio de A | media | **A** |
| 5a0c5989ca | 7 | 08-01 | …/specs/P0xx (uno de los del 08-01) | subdirectorio de A | baja | **A** |
| cb37cb7319 | 3 | 08-05 | …/specs/P012 u otro del 08-05 | subdirectorio de A | media | **A** |
| b34058125b | 3 | 08-01 | …/specs/P0xx | subdirectorio de A | baja | **A** |
| e3f12f9910 | 2 | 07-30 | subdirectorio puntual de A (¿frontend/src?) | subdirectorio de A | baja | **A** |
| 7beb65f968 | 1 | 07-24 | subdirectorio puntual de A | subdirectorio de A | baja | **A** |
| 5d157e1084 | 1 | 08-01 | …/specs/P0xx | subdirectorio de A | baja | **A** |

**A** = identidad de la raíz `/home/bfgnet/dev/permea-platform` · **B** = identidad de la raíz
`/home/bfgnet/dev/agente`

## Predicción — derivada de la regla ANTES de ejecutar (lo que T040 compara)

1. **Las 12 identidades colapsan en exactamente 2 clases**: A (11 refs, 582 eventos) y
   B (1 ref, 61 eventos).
2. **Cero filas caen al fallback**: ningún directorio medido está fuera de un repositorio.
3. El reprocesado íntegro de los logs supervivientes (que incluyen trabajo posterior al
   08-05) debe producir **exactamente 2 identidades de proyecto distintas** — todo cwd
   conocido, incluidos los node_modules/ y los specs/, asciende a una de las dos raíces.
4. Los recuentos por clase no son criterio de SC-001 (los logs abarcan más días que la
   base); el criterio es el número y composición de las clases.

La confianza baja de filas individuales no debilita la predicción: cualquier subdirectorio
de la plataforma pertenece a la clase A por la regla, se llame como se llame.