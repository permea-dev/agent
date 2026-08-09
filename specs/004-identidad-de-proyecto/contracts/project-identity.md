# Contrato · Derivación de la identidad de proyecto

**Feature**: P-004 · **Fecha**: 2026-08-09 · **Estado**: propuesto (pendiente de `/speckit-tasks`)

Este contrato fija **el comportamiento observable** de la derivación: qué entra, qué sale, y qué se
garantiza en cada caso. No fija estructura de código.

**Relación con contratos existentes**: este documento **cumple** `specs/001-agente-inicial/contracts/boundary-event.md`
y **nunca lo redefine**. El campo `project_ref` conserva nombre, tipo y forma; lo único que cambia es
cómo se calcula su valor. La descripción de ese campo en el contrato de frontera **debe actualizarse**
para retirar la promesa de envío en claro (FR-014).

---

## Superficie

```
Derivar(cwdDeclarado string, salt string) → identidad string
```

- **Nunca devuelve error.** Es la forma que exige FR-010: un fallo de resolución no puede propagarse
  hasta detener la emisión de un evento.
- **Determinista dentro de una pasada**: la misma entrada produce la misma salida (por la caché de
  pasada, y porque las tres reglas son deterministas dadas las mismas condiciones del FS).
- **No determinista entre pasadas si el FS cambia** — y eso es correcto por spec (edge case «un
  directorio se convierte en proyecto después»).

---

## Garantías

| # | Garantía | Requisito |
|---|---|---|
| G1 | Entrada vacía → salida vacía | FR-008 |
| G2 | Salida no vacía → **hex-64 minúscula** (forma de `event.Ref`) | FR-018 |
| G3 | Dos entradas de cualquier punto del mismo proyecto observable → **misma** salida | FR-002 |
| G4 | Dos entradas de proyectos distintos → salidas **distintas** | FR-003 |
| G5 | Un enlace y su destino, ambos resolubles → **misma** salida | FR-006a |
| G6 | Variaciones sintácticas de la misma ruta → **misma** salida | FR-006 |
| G7 | Directorios genuinamente distintos → salidas **distintas** | FR-007 |
| G8 | Ningún fallo del sistema de ficheros produce error ni salida vacía | FR-009, FR-010 |
| G9 | La salida **nunca** contiene la ruta, ni un fragmento de ella | FR-016, FR-017 |
| G10 | La derivación **no escribe** en el sistema de ficheros | FR-012 |

---

## Tabla de casos

| # | `cwdDeclarado` | Condición del sistema | Salida | Regla |
|---|---|---|---|---|
| 1 | `""` | cualquiera | `""` | G1 / FR-008 |
| 2 | `/dev/proy/sub` | `/dev/proy/.git` existe (dir) | `Ref(salt, "/dev/proy")` | raíz |
| 3 | `/dev/proy` | `/dev/proy/.git` existe (dir) | `Ref(salt, "/dev/proy")` — **igual que 2** | G3 |
| 4 | `/dev/wt-x/sub` | `/dev/wt-x/.git` existe (**fichero**) | `Ref(salt, "/dev/wt-x")` | árbol paralelo |
| 5 | `/dev/proy` y `/dev/wt-x` | ambos con marcador | salidas **distintas** | edge case «árboles paralelos» |
| 6 | `/dev/a/b/c` | `.git` en `/dev/a` y en `/dev/a/b` | `Ref(salt, "/dev/a/b")` — el **más cercano** | FR-004 |
| 7 | `/home/u/sub` | `.git` en `/home/u` (el home) | `Ref(salt, "/home/u/sub")` — el ascenso **se detiene en el techo** (el home, exclusivo) y nunca mira ahí | FR-004a |
| 7b | `/home/u/dev/proy/sub` | `.git` en `/home/u/dev/proy` | `Ref(salt, "/home/u/dev/proy")` — **un marcador bajo el home cuenta con normalidad** | FR-004a (techo, no salto) |
| 8 | `/sub` | `.git` en `/` (raíz del FS) | `Ref(salt, "/sub")` — techo alcanzado sin marcador | FR-004a |
| 9 | `/tmp/x/` y `/tmp/x` | sin marcador | **misma** salida `Ref(salt, "/tmp/x")` | G6 |
| 10 | `/tmp/a/./b/../b` | sin marcador | `Ref(salt, "/tmp/a/b")` | G6 |
| 11 | `/link` → `/real/proy/sub`, con `.git` en `/real/proy` | resoluble | `Ref(salt, "/real/proy")` — **igual que el caso 2 sobre la ruta real** | G5 / FR-006a |
| 12 | `/borrado/x` | no existe | `Ref(salt, "/borrado/x")` (léxico) — **no falla** | G8 |
| 13 | `/sinpermiso/x` | `Lstat` denegado | `Ref(salt, "/sinpermiso/x")` — **no falla** | G8 |
| 14 | `/roto` (enlace roto) | irresoluble | `Ref(salt, "/roto")` — **no falla** | G8 |
| 15 | `relativa/x` | no absoluta | `Ref(salt, "relativa/x")` limpiado — **sin anclar** | research §P2 |

**El caso 5 es el que la implementación tiene que demostrar**, y es la razón de que el marcador
acepte fichero: comprobar solo «directorio `.git`» haría que el caso 4 cayera al fallback y el 5
siguiera pasando por accidente — la promesa quedaría cumplida por el motivo equivocado y se rompería
en cuanto el worktree estuviera dentro de otro repositorio.

---

## Lo que este contrato NO promete

- **No promete estabilidad entre máquinas ni entre usuarios**: el salt es por instalación. Dos
  personas en el mismo proyecto obtienen identidades distintas (fuera de alcance, lo resuelve el
  mapeo de la plataforma).
- **No promete estabilidad en el tiempo si el sistema de ficheros cambia**: inicializar un
  repositorio en un directorio ya usado cambia su identidad a partir de ese momento.
- **No promete distinguir origen**: el consumidor no puede saber por la forma si una identidad viene
  de una raíz o de un directorio suelto (FR-018, deliberado).
- **No promete case-folding**: dos grafías de un directorio **irresoluble** en un FS insensible
  pueden dar dos identidades (residuo declarado, research §P2).
