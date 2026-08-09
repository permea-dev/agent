# Specification Quality Checklist: Identidad de proyecto

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-08-09
**Re-validado**: 2026-08-09 — **contra la spec editada por los pases E-1 (8 hallazgos) y E-1b**. El
16/16 anterior validaba una versión que ya no existe y **no se ha heredado**: cada ítem se ha vuelto
a comprobar contra el texto actual, con su cita. **+E-1b**: la extensión de FR-004a a la raíz del
sistema de ficheros y la separación del edge case de los temporales quedan cubiertas por esta misma
re-validación — no alteran ningún veredicto, porque amplían el **sujeto** de una regla ya validada sin
cambiar su forma ni su criterio.
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] **No implementation details** — la spec sigue sin nombrar lenguaje, librería ni mecanismo. Las
      dos incorporaciones de E-1 que rozan el límite no lo cruzan: FR-014 nombra **documentos**
      (`spec.md:269-276`), no código, y SC-008 exige un **método de verificación** reproducible
      (`spec.md:347-350`), que es criterio de aceptación, no diseño del producto.
- [x] **Focused on user value and business needs** — el marco sigue siendo el gasto por proyecto y la
      línea de dinero de la plataforma (`spec.md:16-46`).
- [x] **Written for non-technical stakeholders** — FR-004a y sus edge cases se explican por su
      consecuencia («colapsaría bajo una sola identidad todo lo que cuelga de él»,
      `spec.md:175-180`), no por su mecánica.
- [x] **All mandatory sections completed** — User Scenarios, Requirements, Success Criteria y
      Assumptions presentes.

## Requirement Completeness

- [x] **No [NEEDS CLARIFICATION] markers remain** — cero marcadores. Las cuatro decisiones están
      **tomadas y escritas con su criterio**: D-004-1 (`spec.md:356-368`), D-004-2 (`:370-378`),
      D-004-3 (`:380-387`), D-004-4 (`:389-396`).
- [x] **Requirements are testable and unambiguous** — **este ítem cambió de veredicto con E-1** y es
      la razón principal de re-validar: en la versión anterior, FR-015 («no se queda como valor
      aceptado sin efecto») **contradecía** a FR-013a/FR-013b, que mandan justamente ignorar en
      silencio los valores residuales. Un implementador tenía dos órdenes opuestas. Resuelto en
      `spec.md:277-280`: FR-015 habla ahora de la superficie **documentada y con significado**, y
      remite explícitamente el trato de residuales a FR-013/013a/013b.
- [x] **Success criteria are measurable** — con dos correcciones de E-1 que este ítem sí exigía:
      SC-002 ya no es un absoluto que FR-009 desmentía (`spec.md:328-331`, cualificado a «directorio
      observable» y con la divergencia declarada), y SC-001 ya no promete un desenlace que las reglas
      no garantizan (`spec.md:322-327`, ahora predicción-contra-verdad-terreno).
- [x] **Success criteria are technology-agnostic** — ninguno nombra tecnología. SC-006 mide duración
      percibida; SC-008 mide ausencia por barrido.
- [x] **All acceptance scenarios are defined** — US1 tiene 6 (el 6.º, de enlace hacia el interior de
      un proyecto, entró por D-004-2: `spec.md:85-88`), US2 tiene 4, US3 tiene 5.
- [x] **Edge cases are identified** — 13 casos. E-1 partió el que mezclaba raíz del sistema,
      temporales y directorio personal, y **E-1b terminó de partirlo en tres** (`spec.md:172-187`):
      los **temporales** caen en FR-005 porque ningún proyecto los contiene; el **directorio
      personal** y la **raíz del sistema** quedan excluidos por **regla explícita** (FR-004a), no por
      suposición — que era el defecto. La razón es la misma para los dos: ambos pueden ser raíz de un
      árbol de trabajo, y la regla general los colapsaría.
- [x] **Scope is clearly bounded** — y **mejor acotado que antes**: D-004-4 saca las salidas
      diagnósticas locales del alcance de FR-017 (`spec.md:287-291`) y las deja **nombradas como
      candidata aparte** en vez de cubiertas por un requisito que nadie iba a verificar.
- [x] **Dependencies and assumptions identified** — con una dependencia **nueva** que E-1 hizo
      aflorar: SC-001 solo es verificable si Basilio aporta la enumeración de verdad-terreno de los
      directorios de origen. Registrada como primera asunción (`spec.md:402-406`).

## Feature Readiness

- [x] **All functional requirements have clear acceptance criteria** — 23 requisitos (FR-001..020 más
      FR-004a, FR-006a, FR-013a/b), cada uno con escenario o criterio que lo ejercita.
- [x] **User scenarios cover primary flows** — proyecto (US1), directorio suelto (US2), contrato y
      configuración (US3).
- [x] **Feature meets measurable outcomes defined in Success Criteria** — SC-001..008 cubren las tres
      historias y las tres invariantes que no cambian.
- [x] **No implementation details leak into specification** — el mecanismo sigue diferido a
      `research.md` / `plan.md` (`spec.md:414-418`).

## Constitución (Permea v1.0.0) — verificación de sujeto y de fondo

**Verificación de sujeto (F8, hecha ANTES de editar)**: los cinco nombres y su numeración se
contrastaron contra `.specify/memory/constitution.md` y **coinciden exactamente** —
`### I. Frontera de datos inviolable (NO NEGOCIABLE)` (`:7`), `### II. Privacidad auditable, no
prometida (local-first)` (`:18`), `### III. Binario único y auditable` (`:27`), `### IV. Test-first en
la frontera (NO NEGOCIABLE)` (`:36`), `### V. Desarrollo dirigido por especificaciones` (`:45`) —, y
la versión declarada es `**Version**: 1.0.0` (`:87`). Sin desajuste: el checklist verifica contra el
documento real, no contra uno imaginado.

- [x] **Principio I — Frontera de datos**: la feature NO amplía la allowlist ni añade campos
      (FR-020); la identidad sigue cruzando solo como valor irreversible (FR-016). FR-017 **se acotó**
      en E-1 al camino hacia el exterior (`spec.md:287-291`): sigue siendo un estrechamiento de la
      frontera —nunca una relajación—, con el resto identificado como trabajo aparte.
- [x] **Principio II — Privacidad auditable**: la ampliación de comportamiento (observar el sistema
      de ficheros) se declara en el alcance (FR-011) y se acota a solo lectura de lo necesario
      (FR-012).
- [x] **Principio III — Binario único**: la regla se enuncia de forma multiplataforma; la estabilidad
      se exige dentro de una máquina.
- [x] **Principio IV — Test-first en la frontera**: escenarios en Given/When/Then; requisitos con
      DEBE/NUNCA; SC-005 es el golden test de frontera de esta feature, ahora con su alcance dicho con
      precisión (evento serializado, cola y transmitido).
- [x] **Principio V — SDD**: la spec dice QUÉ; el mecanismo queda diferido a research/plan.

## Notes

- **Sin bloqueantes**: los 16 ítems del checklist estándar pasan **contra el texto editado**, más los
  5 de constitución. La spec está lista para `/speckit-plan`.
- **Lo que E-1 cambió de veredicto**, y por qué re-validar no era ceremonia: «Requirements are
  testable and unambiguous» **no lo cumplía** la versión anterior (FR-015 contra FR-013a/b), y
  «Success criteria are measurable» lo cumplía solo en apariencia (SC-002 absoluto, SC-001 sin
  verdad-terreno). Heredar el 16/16 habría dado por buenas tres afirmaciones falsas.
- **Insumo pendiente, no bloqueante para planificar**: la enumeración de verdad-terreno que SC-001
  necesita. Bloquea la **validación**, no el plan.
- **Lo que el plan tiene que resolver y la spec deja fuera a propósito**: qué marca la raíz de un
  árbol de trabajo, qué normalización sintáctica exacta se aplica, cómo se ordena la resolución de
  enlaces respecto del reconocimiento (FR-006a fija el **orden**, no el mecanismo), y cómo se evita
  que la observación se repita por evento (SC-006).
- **Candidata identificada, fuera de esta feature**: saneado de las salidas diagnósticas locales
  (stderr y equivalentes) — D-004-4.
