<!--
Sync Impact Report
- Version change: 1.0.0 → 1.1.0 (MINOR: se amplía la regla de coste del Principio II; nada se retira)
- Modified principles: II. Privacidad auditable, no prometida (local-first): título sin cambios. La viñeta
  «El cálculo de coste NUNCA depende del backend: la tabla de precios viaja empaquetada en el binario»
  pasa a la regla de dos casos (coste fiable de la herramienta → tabla empaquetada; si no → evento sin coste,
  cost_usd = 0 y cost_available = false, y lo calcula la plataforma con su catálogo). La viñeta del parseo
  precisa «cuando lo hace el agente». Motivo: es el patrón de 008, 009 y 010 (lectores de Codex, Gemini CLI y
  OpenCode) y de 011 (Copilot CLI, D-4); decisión del dueño, 2026-10-14 (encargo 011 · 7/n, hallazgo K1 del analyze).
- Added sections: ninguna. Removed sections: ninguna.
- Templates: ✅ .specify/templates/plan-template.md, spec-template.md, tasks-template.md (no citan la regla de
  coste; sin cambios) · ✅ README.md (describe el coste de Claude Code con la tabla empaquetada: sigue siendo cierto).
- Follow-up TODOs: ninguno.
-->
# Constitución de Permea

Documento rector del repositorio del **agente de Permea**: un binario local, multiplataforma y de código abierto que mide el coste de uso de herramientas de IA y transmite al backend de equipo **únicamente metadato derivado, nunca contenido**. Esta constitución es la fuente de verdad transversal del proyecto: gobierna toda especificación, plan y tarea. Ninguna decisión de implementación puede contradecirla.

## Core Principles

### I. Frontera de datos inviolable (NO NEGOCIABLE)

El agente transmite exclusivamente metadato derivado (tokens, coste, modelo, marcas de tiempo e identificadores pseudónimos). El contenido —prompts, respuestas, código, diffs, rutas de fichero y nombres de proyecto en claro— NUNCA sale de la máquina del desarrollador.

- La frontera se implementa como **allowlist cerrada**: un struct de evento que contiene solo los campos permitidos. No existe passthrough de campos crudos del log.
- El diseño es **deny-by-default**: lo que no se mapea explícitamente a un campo permitido, se descarta. Un campo nuevo o desconocido en el origen no puede filtrarse por construcción.
- Ninguna especificación puede ampliar la allowlist con contenido. Toda herramienta nueva (Cursor, Copilot, aider, Codex…) DEBE mapear su origen contra la frontera existente; si su ingesta exige algo que no encaja, se rediseña la ingesta, no la frontera.
- Los identificadores sensibles (ruta de proyecto, sesión, máquina) cruzan solo como hash salado. El \`salt\` reside en local y NUNCA se transmite.

Este principio es la fuente de verdad del contrato de frontera. Toda especificación lo referencia; ninguna lo redefine.

### II. Privacidad auditable, no prometida (local-first)

La confianza del producto no se sostiene en una política, sino en código que cualquiera puede leer.

- El agente es de **código abierto**. La promesa de privacidad se verifica leyendo el código de la frontera, no confiando en el proveedor.
- El parseo, la agregación y, cuando lo hace el agente, el cálculo de coste ocurren **en local**. El agente funciona sin conexión; la ausencia de red no detiene el procesamiento local.
- **Coste**: si la herramienta da un coste fiable, el agente lo calcula con la **tabla de precios empaquetada en el binario**, sin depender del backend. Si no lo da, el evento va **sin coste** (`cost_usd = 0`, `cost_available = false`) y lo calcula la plataforma con su catálogo. El agente NUNCA consulta el backend para calcular un coste.
- Un revisor externo DEBE poder confirmar el cumplimiento de la frontera leyendo únicamente el código de serialización del evento.

### III. Binario único y auditable

La distribución y la legibilidad son parte del producto, no un añadido.

- El agente se distribuye como **binario estático único**, sin runtime externo, instalable con un solo comando en macOS, Linux y Windows.
- Se favorece la **librería estándar**. Cualquier dependencia externa DEBE justificarse explícitamente en la especificación que la introduce; en caso de duda, no se añade.
- La resolución de rutas es por sistema operativo (nunca se hardcodean rutas de logs); el agente contempla que un cliente ejecute la herramienta en Windows nativo aunque el desarrollo ocurra en WSL/Linux.
- El código prioriza la legibilidad sobre la astucia: un desarrollador escéptico debe entender la frontera de un vistazo.

### IV. Test-first en la frontera (NO NEGOCIABLE)

La propiedad que define el producto se prueba antes de implementarse.

- El **golden test de frontera** —una entrada con contenido sensible inyectado que demuestra que ningún dato de la denylist sobrevive al evento— es disciplina de primer commit y DEBE existir antes que cualquier parser de una herramienta.
- Los escenarios de aceptación se expresan en formato **Given/When/Then**.
- Los requisitos se expresan con **DEBE** (obligatorio) y **NUNCA** (prohibido); no hay requisitos ambiguos.
- Ningún cambio se da por bueno si el test de frontera no está en verde.

### V. Desarrollo dirigido por especificaciones

Todo desarrollo se gobierna por una especificación formal antes de escribir código.

- El ciclo es **\`/speckit.specify\` → \`/speckit.plan\` → \`/speckit.tasks\` → \`/speckit.implement\`**; cada fase lee la anterior.
- Cada funcionalidad vive en \`specs/NNN-slug/\` con su \`spec.md\`, y los artefactos que genere el plan (\`plan.md\`, \`tasks.md\`, y cuando apliquen \`research.md\`, \`data-model.md\`, \`contracts/\`).
- La especificación describe **qué** y **por qué**, no **cómo**: sin detalles de implementación en el \`spec.md\`.
- Ninguna especificación se implementa sin haber superado su checklist previo ni sin ser consistente con esta constitución.

## Restricciones técnicas

- **Lenguaje**: Go. Binario único vía \`cmd/permea\`.
- **Estructura de referencia del repositorio**:
  - \`internal/event/\` — la frontera: struct cerrado del evento + helpers de hash e id. Es el código auditable por excelencia; se mantiene pequeño y obvio.
  - \`internal/ingest/\` — lectores por herramienta (uno por herramienta soportada) + tests de frontera. El lector decodifica solo campos permitidos.
  - \`internal/pricing/\` — cálculo de coste local con tabla empaquetada y actualizable.
  - \`internal/state/\` — escaneo incremental: no se reprocesan llamadas ya emitidas.
  - \`internal/transport/\` — cliente HTTPS con cola offline y reintentos; deduplicación por \`event_id\`.
  - \`internal/config/\` — configuración local en fichero legible por el usuario.
- **Fixtures**: el desarrollo usa datos JSONL anonimizados en \`testdata/\`; NUNCA se commitean logs reales de clientes.
- **Transporte**: todo envío es HTTPS y autenticado con token de dispositivo por instalación. NUNCA se transmite en claro.
- **Cálculo de coste**: en local, con tabla de precios empaquetada; nunca depende del backend.

## Flujo de desarrollo

- **Comandos**: \`/speckit.specify\`, \`/speckit.plan\`, \`/speckit.tasks\`, \`/speckit.implement\`, apoyados por \`analyze\`, \`clarify\`, \`checklist\` cuando aporten.
- **Convención de requisitos**: DEBE / NUNCA; escenarios en Given/When/Then; criterios de éxito medibles (\`SC-XXX\`).
- **Puertas de calidad** antes de dar por cerrada una tarea:
  - \`go vet ./...\` sin hallazgos.
  - \`golangci-lint run\` limpio (el linter es parte del producto, no cosmético).
  - Test de frontera y suite completa (\`go test ./...\`) en verde.
- **Disciplina de commits**: se commitea el esqueleto y el test de frontera antes de construir sobre él, de modo que cualquier regresión de la frontera sea visible en el diff.

## Governance

Esta constitución prevalece sobre cualquier otra práctica del repositorio. En caso de conflicto entre una especificación y esta constitución, gana la constitución.

- El **Principio I (Frontera de datos)** no puede ser anulado, relajado ni excepcionado por ninguna especificación, plan o tarea. Cualquier propuesta que lo requiera se rechaza de plano.
- Toda especificación y toda revisión de código DEBEN verificar el cumplimiento de la frontera de forma explícita.
- Las enmiendas a esta constitución requieren: justificación escrita, incremento de versión y, si alteran la frontera de forma no ampliable, revisión especial documentada.
- La complejidad añadida DEBE justificarse; ante la duda, se elige la opción más simple y auditable.

**Version**: 1.1.0 | **Ratified**: 2026-07-02 | **Last Amended**: 2026-10-14
