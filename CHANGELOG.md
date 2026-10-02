# Changelog

Cambios visibles del agente `permea`, de la versión más reciente a la más antigua. Cada punto cita la
especificación de la que sale (`specs/NNN-…/spec.md`).

## 0.3.0 — PENDIENTE

Primera versión desde la 0.2.1. Trae el enrolamiento, la identidad y la adhesión a proyecto, y
corrige la medición. **Si venías de la 0.2.1, lee «Cambios que rompen» y «Las cifras bajan».**

### Nuevo

- **`permea enroll`**: conecta la instalación con su organización a partir del *enrollment string*
  que da la aplicación. Verifica el token contra el servidor antes de guardarlo, y sólo si lo acepta
  lo guarda en `config.json` con permisos 0600; un rechazo no escribe nada. Acepta el valor por
  argumento o, recomendado, por stdin. Formato `pmea2.`, con el `dev_id` que asigna el servidor; el
  antiguo `pmea1.` se rechaza.
  (`specs/003-enrolamiento/spec.md`, FR-001 a FR-007, FR-013;
  `specs/003-enrolamiento/contracts/enrollment-string.md`)
- **`permea status`**: dice si la instalación está enrolada y contra qué servidor. Es local y nunca
  muestra el token. (`specs/003-enrolamiento/spec.md`, FR-008)
- **Identidad de proyecto por árbol de trabajo**: los eventos de un mismo proyecto comparten
  identidad, se lance Claude Code desde la raíz o desde cualquier subdirectorio. Antes cada punto de
  lanzamiento era una identidad distinta. (`specs/004-identidad-de-proyecto/spec.md`, P-004 FR-001 a
  FR-005)
- **`permea project join`**: une la instalación a un Proyecto de la organización presentando el código
  que se genera en el panel, y responde con el nombre del Proyecto. Se ejecuta dentro del árbol de
  trabajo, exige HTTPS y no guarda nada en local. (`specs/005-adhesion-a-proyecto/spec.md`, P-005
  FR-001, FR-002, FR-006, FR-017, FR-019)
- **Un evento por mensaje** (006). Claude Code escribe varias líneas por mensaje con el mismo consumo,
  y la 0.2.1 emitía un evento por línea. Ahora sale **uno por mensaje**. (`specs/006-medicion-fiel/spec.md`,
  P-006 FR-001, FR-033)
- **`event_id` determinista** (006): se deriva del propio mensaje, sin sal, así que el mismo mensaje
  da el mismo identificador en cualquier pasada o instalación, y la plataforma descarta los
  repetidos. Las respuestas **`<synthetic>`** de Claude Code ya no se emiten. (P-006 FR-002, FR-007,
  FR-008; `specs/006-medicion-fiel/contracts/event-id.md`)
- **Tarifas de 16 modelos** (006), espejo del catálogo de la plataforma, con su procedencia en la
  cabecera de la tabla. Los modelos en uso (`claude-opus-5-5`, `claude-opus-5`, `claude-sonnet-5`)
  tienen por fin coste. **`claude-opus-4-6` pasa de 15 / 75 a 5 / 25 USD por millón: su coste baja a
  un tercio.** (P-006 FR-014 a FR-020; `specs/006-medicion-fiel/contracts/tarifas.md`)
- **La ayuda** (006): una sola, por stdout, con «Primeros pasos». Cada subcomando tiene la suya
  (`permea <subcomando> -h`), y pedirla no ejecuta nada. (P-006 FR-021, FR-023, D-006-14)
- **`--scan`** imprime por evento las cuatro partidas de tokens y el `event_id`, y cada pasada termina
  con un resumen de recuentos por stderr. (P-006 FR-010, FR-005, FR-006)

### Las cifras bajan

**Las cifras bajan respecto a la 0.2.1 porque antes se contaba de más.** Con un evento por línea,
cada mensaje se contaba tantas veces como líneas ocupaba en el log: en los datos medidos, los tokens
salían **×2,13**. A partir de la 0.3.0 cada mensaje cuenta una vez. Es una corrección, no una pérdida.
(`specs/006-medicion-fiel/spec.md`, §Contexto de partida, M2)

Lo enviado antes de actualizar **no se corrige ni se reenvía**, y lo que la 0.2.1 dejó en cola se
envía tal cual: la corrección aplica a lo que se mida desde la 0.3.0. (P-006 FR-009; §Fuera de alcance,
«Los datos que la plataforma ya recibió»)

### Cambios que rompen

- **`project_ref` cambia para el mismo proyecto.** La identidad de proyecto se deriva ahora de la raíz
  del árbol de trabajo, así que las identidades emitidas por la 0.2.1 no corresponden con las de la
  0.3.0. Es una ruptura aceptada, y la correspondencia de identidades históricas queda fuera de
  alcance. (`specs/004-identidad-de-proyecto/spec.md`,
  §Assumptions, «Continuidad de identidades históricas: se rompe, y se acepta»)
- **El modo de envío en claro de `project_ref` se retiró.** Una configuración que todavía lo pida
  hace que `--run` y `--daemon` se detengan con un error que lo explica; `permea enroll` lo repara.
  (`specs/004-identidad-de-proyecto/spec.md`, P-004 FR-013 a FR-015;
  `specs/004-identidad-de-proyecto/plan.md`, D-004-5)
- **La ayuda cambia de canal y de forma** (006):
  - **`permea` sin argumentos ya no escribe por stderr**: la ayuda sale por stdout, con exit 0.
  - **Un subcomando inexistente sale con exit 1**, nombrándolo (antes, exit 0 y la ayuda).
  - **Una opción desconocida o sin valor sale con exit 2 sin imprimir el uso**, con un mensaje propio
    que remite a `permea help`.

  (P-006 FR-021, FR-022, D-006-7; `specs/006-medicion-fiel/contracts/cli.md`)
