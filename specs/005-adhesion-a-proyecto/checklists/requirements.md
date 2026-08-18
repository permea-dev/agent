# Specification Quality Checklist: Adhesión a proyecto

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-08-18
**Revalidated**: 2026-08-18 — **sexto pase: SC-011 y su columna de la matriz**
**Feature**: [spec.md](../spec.md)

---

# 🔒 CONGELADA

> **Congelada el 2026-08-18**, tras cinco pases adversariales, la matriz de historias × criterios
> recorrida entera y SC-011. **La última ventana para editarla es la fase de plan**: si el plan revela
> que la spec estaba mal, ese es el momento. A partir de tasks, tocarla ya no es corregir — es
> **derogar un requisito**, y eso exige decisión de Basilio y quedar escrito.

*(El campo `Status` de la cabecera **no se toca**: este repositorio usa un único valor, `Draft`, en las
cinco specs y en la plantilla, y las cuatro features ya mergeadas siguen en él. No hay convención de
promoción, y crearla es decisión de Basilio. La congelación se registra **aquí**, que es donde hay
sitio para decir por qué.)*

---

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] **Feature meets measurable outcomes defined in Success Criteria** — verificado por la matriz,
      cruce a cruce, **incluida la columna de SC-011**
- [x] No implementation details leak into specification

---

# LA MATRIZ — 5 historias × 11 criterios

> **—** el criterio no impone nada a esa historia · **✅** impone y ya se reflejaba ·
> **⚠️→✅** impone, no se reflejaba, **alineado**

| | SC-001 | SC-002 | SC-003 | SC-004 | SC-005 | SC-006 | SC-007 | SC-008 | SC-009 | SC-010 | **SC-011** |
|---|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|
| **US1** unión nueva | — | — | — | — | — | ⚠️→✅ | ✅ | — | — | — | **⚠️→✅** |
| **US2** repetición | — | ⚠️→✅ | — | — | — | ⚠️→✅ | — | — | — | — | **✅** |
| **US3** rehúse local | — | — | — | ✅ | — | — | ⚠️→✅ | — | — | — | **⚠️→✅** |
| **US4** desenlaces sin fuga | — | — | ⚠️→✅ | — | ⚠️→✅ | — | — | — | — | — | **⚠️→✅** |
| **US5** estados no operables | — | — | — | — | — | — | ✅ | — | — | ⚠️→✅ | **⚠️→✅** |

**55 cruces · 4 ✅ · 11 ⚠️→✅ · 40 —**

## La columna SC-011, recorrida en el acto

| Historia | Marca | Por qué |
|---|:--:|---|
| **US1** | **⚠️→✅** | Es la única con **(A)**: el escenario 4 decía «indistinguible» **a secas**, que es el hueco corregido ya dos veces. `spec.md:84-86` remite a **SC-011 (A)** —qué se compara, con qué piezas, cómo se demuestra que sabe fallar— y a **(B)** para el canal del éxito |
| **US2** | **✅** | **Ya lo reflejaba**, y por eso no se tocó: su Independent Test compara «texto, **canal** y resultado del proceso» (`:99-100`) y su escenario 2 repite «mismo texto, **mismo canal**» (`:63`). Es la granularidad exacta que pide **(B)**. No le aplica **(A)**: US2 no cambia de vía de entrada |
| **US3** | **⚠️→✅** | **(B)** para rehúse: `spec.md:139-140` — *«el rehúse sale por el canal que exige SC-011 (B), con los dos capturados por separado»* |
| **US4** | **⚠️→✅** | **(B)** para error: `spec.md:170-171` — *«Los dos son desenlaces de error, así que salen por el canal que exige SC-011 (B)»* |
| **US5** | **⚠️→✅** | **(B)** para rehúse y error: `spec.md:201-202` — *«Los tres son rehúses o errores, así que salen por el canal que exige SC-011 (B)»* |

**Cuatro de las cinco historias la necesitaban**, no una. **(A)** solo toca a US1 —es la única que
presenta el código por dos vías— pero **(B) discrimina por clase de desenlace**, y cada historia cae en
una: US1 y US2 son de éxito, US3, US4 y US5 de rehúse o error. Por eso la columna no es «US1 y cuatro
guiones».

## Los tres criterios con la columna vacía — sigue siendo el hallazgo abierto

**SC-001**, **SC-008** y **SC-009** no imponen nada a ninguna historia: verifican requisitos
—FR-004/005, FR-017, FR-014/015/016— que **ninguna historia enuncia**. No se han creado historias para
cubrirlos. Están verificados por criterio; lo que no tienen es escenario.

**Con SC-011, ya no queda ningún requisito sin criterio que lo mida**: FR-021 y FR-023 eran los dos
últimos, y son justamente lo que SC-011 cubre.

---

## Recuentos — **solo SC se movió**

| | Pase 5 | **Pase 6** | |
|---|---:|---:|:--:|
| Requisitos funcionales | 24 | **24** | = |
| **Criterios de éxito** | 10 | **11** | **+1 · SC-011** |
| Historias de usuario | 5 | **5** | = |
| Escenarios de aceptación | 15 | **15** | = |
| Casos límite | 8 | **8** | = |
| Decisiones fechadas | 6 | **6** | = |
| Supuestos | 6 | **6** | = |
| Exclusiones con motivo | 7 | **7** | = |
| `[NEEDS CLARIFICATION]` | 0 | **0** | = |

**Los otros ocho, clavados.** Las cuatro alineaciones de la columna se hicieron **remitiendo**, sin
crear ni endurecer nada.

## Barrido de filtraciones — mismo patrón de los seis pases

**Una coincidencia**, `:164` → «e**struct**ura interna de la organización». El falso positivo de
siempre. **SC-011 lo pasa**: `stdout` y `stderr` no están en el patrón y son vocabulario ya establecido
por FR-021, no implementación.

## Relectura de lo nuevo contra el resto

SC-011 y sus cuatro remisiones, contrastados con lo ya escrito. **No contradicen nada**:

- **con FR-021** — lo hace medible sin cambiarlo: donde FR-021 dice «stdout es la respuesta, stderr es
  todo lo demás», SC-011 (B) añade **cómo se comprueba** (stdout vacío en rehúse, captura por separado);
- **con FR-023** — le da por fin criterio, sin ampliar lo que exige;
- **con SC-002** — sujetos distintos, y SC-011 lo dice en su propia nota: SC-002 son los dos desenlaces
  **de la plataforma**, SC-011 (A) las dos **vías de entrada**;
- **con FR-010 y US2** — «bajo las mismas condiciones» es la misma cláusula que FR-010 ya usa, y US2 ya
  comparaba canal.

---

## Coherencia interna — el expediente

> **Las entradas no se borran: se marcan.** Misma regla que `contratos-verificados.md`.

### ~~C-1 · «no emite eventos» contra «emite con toda normalidad»~~ — ✅ **RESUELTA el 2026-08-18**

US3 afirmaba que la identidad de un directorio suelto *«no emite eventos»*, contra el caso límite 8,
que cita `004/spec.md:229-230`. Reformulada **sin debilitar** el argumento: esa identidad **sí emite**,
pero **no es la del proyecto que se quería agrupar**.

### ~~C-2 · «sin dejar nada a medias» sin acotar~~ — ✅ **RESUELTA el 2026-08-18, en TRES sitios**

Cabecera de US5, escenario 2 y —el no enumerado— Independent Test de US5. Los tres acotados «en local»,
con remisión a **FR-013a**.

### Retoques y matrices — ✅ **2026-08-18**

**M-1** US1 esc 3 → SC-007 · **M-2** US3 IT → SC-004 · **M-3** `Status` → **no se cambió**, medido ·
**Matriz 5×10** recorrida entera · **SC-011 y su columna** recorrida en el mismo pase que lo creó, para
no repetir el defecto de dejar las historias atrás.

## Notes

- **`.specify/feature.json` no se tocó** en ninguno de los seis pases.
- **No se escribieron** `plan.md`, `research.md`, `tasks.md` ni contratos.
- **Spec congelada**: ver el bloque del principio.
