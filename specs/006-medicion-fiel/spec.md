# Feature Specification: Medición fiel y publicable — que cada mensaje cuente una vez, a su precio, y que la versión se pueda instalar

**Feature Branch**: `006-medicion-fiel`

**Created**: 2026-10-02

**Status**: Draft

**Amended**: 2026-10-02 — enmiendas D-006-7 a D-006-13 y pregunta abierta Q-006-1
(§Decisiones › Enmiendas del 2026-10-02). Los números ya asignados no se renumeran;
lo revocado se marca, no se borra.

**Input**: Descripción y decisiones de Basilio, 2026-10-02, tras el descubrimiento del mismo día en el
agente (`~/dev/permea-platform/tmp/report-agente.md`, versión del 2026-10-02). Los requisitos se citan
como **P-006 FR-xxx** y los criterios como **P-006 SC-xxx**: la forma abreviada no identifica nada,
porque hay números repetidos en otras especificaciones de `specs/`.

---

## Contexto: el agente mide de más, a precio equivocado, y la versión publicada no se puede enrolar

El 2026-10-02 alguien instaló el agente por Scoop siguiendo el camino normal. Recibió la `0.2.1`, la
última publicada (2026-07-04), y la aplicación le pidió `permea enroll <secreto>`. **La `0.2.1` no tiene
`enroll`**: el enrolamiento (003) se escribió el 2026-07-11, después de esa etiqueta. Desde entonces
`main` acumula **42 commits sin publicar** —003, 004 y 005 enteras—, y la versión que se instala no
sirve para el flujo que la plataforma ya ofrece.

Publicar sin más no basta, porque el descubrimiento sacó a la luz **dos defectos de medición** que
están en `main` y en la `0.2.1` por igual. Publicar sin corregirlos sería repartir una versión nueva que
mide igual de mal:

1. **Cada mensaje se cuenta varias veces.** Claude Code escribe **varias líneas por mensaje**, todas con
   el mismo consumo, y el agente emite **un evento por línea**, cada uno con un identificador aleatorio.
   La plataforma no puede detectar el duplicado, porque para ella son eventos distintos. Sobre los logs
   locales, los tokens salen **inflados ×2,13** (cifras en §Contexto de partida). Contradice el contrato
   de 001: *«nº de eventos recibidos = nº de llamadas reales»*
   (`specs/001-agente-inicial/contracts/transport.md:46`).
2. **El precio no se conoce, o está mal.** La tabla empaquetada tiene **3 claves**, sin fuente ni fecha
   (`internal/pricing/pricing.go:1-2,14-18`). **Ninguno de los modelos que se usan hoy tiene fila**
   (`claude-opus-5-5`, `claude-opus-5`, `claude-sonnet-5`), así que el 100 % de los eventos recientes
   sale con `cost_available=false`. Y la única fila de Opus, `claude-opus-4-6`, vale 15 / 75, **el
   triple** de la tarifa vigente.

A esto se suman dos defectos de superficie que hacen que la versión no se pueda usar sin ayuda. **La
ayuda miente por omisión**: `permea --help` no lista ningún subcomando, y `permea project join -h`
llega a emitir una petición autenticada con `-h` como código. **El README instala desde repositorios
que dan 404** (`bfgnet/…`).

**Esta feature deja la `0.3.0` publicable**: un mensaje, un evento; tarifas fieles; una ayuda que dice
la verdad; y un README con el que se instala. Cierra **publicando** la versión y con el ensayo del dueño
en Windows, que es donde apareció el problema.

**Y no repara lo que la plataforma ya recibió.** Los eventos duplicados ya ingeridos llevan
identificadores aleatorios distintos entre sí, y nada en ellos permite reconocer desde el agente cuáles
son el mismo mensaje. Lo ya recibido queda **fuera de alcance**, con motivo (§Fuera de alcance). Se
escribe aquí porque, sin decirlo, la ausencia de una corrección retroactiva se leería como un olvido.

---

## Contexto de partida (medido)

Todas las cifras proceden del descubrimiento del 2026-10-02 o se midieron de nuevo ese mismo día al
escribir esta spec. De los logs se leyeron **sólo** `type`, `timestamp`, `message.model`, `message.id`,
`requestId` y `usage`; **ningún contenido**. Fuente: `~/.claude/projects` en WSL, 27 ficheros `.jsonl`.
El `%USERPROFILE%\.claude` de Windows nativo **no se midió**.

**M1 · Líneas facturables sin identificador** (medida al escribir la spec, todo el historial local).
Se toma la definición de facturable que usa el agente: `type=assistant` y `model` no vacío,
`internal/ingest/claudecode.go:64`.

| Concepto | Recuento |
|---|---|
| Líneas facturables | **22 860**, de ellas **3** con `model=<synthetic>` |
| Sin `message.id` | **0** |
| Sin `requestId` | **2**, las dos `<synthetic>` |
| Sin ninguno de los dos | **0** |
| Líneas no sintéticas a las que les falta alguno | **0** |

**M2 · ¿El consumo es idéntico entre las líneas de un mismo mensaje?** **Se confirma.**

| Concepto | Recuento |
|---|---|
| Mensajes distintos (`message.id`, `requestId`), sin `<synthetic>` | **10 698**, que ocupan 22 857 líneas |
| Mensajes de más de una línea | **7 612** |
| … con `usage` idéntico (las 4 partidas) en todas sus líneas | **7 612 de 7 612** (el descubrimiento, con el historial de unas horas antes: 7 553 de 7 553) |
| … con `usage` distinto | **0** |
| Mensajes repartidos en más de un fichero | **0** |
| `message.id` con más de un `requestId`, o al revés | **0 y 0**: correspondencia 1:1 |
| Tokens (4 partidas) sumando línea a línea | **10 292 101 351** |
| Tokens contando una vez por mensaje | **4 835 167 368**, es decir **×2,13** |

Por modelo, como mensajes / líneas: `claude-opus-5` 7 318 / 15 221 · `claude-opus-5-5` 3 377 / 7 632 ·
`claude-sonnet-5` 3 / 4. Ninguno lleva sufijo de fecha.

**M3 · Restricciones actuales de `event_id`.**

| Dónde | Qué dice |
|---|---|
| Agente | 16 bytes de `crypto/rand` en hexadecimal: **32 caracteres `[0-9a-f]`**, acuñados al ingerir (`internal/event/event.go:48-55`, `internal/ingest/claudecode.go:67`). Es estable en los reintentos porque viaja en la cola |
| Contrato de frontera (001) | `"type": "string", "description": "hex; clave de deduplicación"` (`specs/001-agente-inicial/contracts/boundary-event.md:25`), con un ejemplo de 32 hex (`:53`). **No fija longitud** |
| Contrato de transporte (001) | «El `event_id` lo genera el agente con `crypto/rand` (`event.NewID`), estable por evento a lo largo de reintentos» (`specs/001-agente-inicial/contracts/transport.md:47-48`). El backend deduplica por `event_id` (`:44-46`) |
| Modelo de datos (001) | «`event.NewID()` (crypto/rand) · Único; clave de deduplicación» (`specs/001-agente-inicial/data-model.md:19`) |
| Plataforma (referencia, no contrato del agente) | Columna `string('event_id')`, NOT NULL (`permea-platform/backend/database/migrations/2026_07_04_000003_create_metric_events_table.php:22`). Validación: requerido, no vacío, string (`backend/app/Ingest/EventAllowlist.php:78`). Unicidad por `(org_id, event_id)` con `ON CONFLICT DO NOTHING` (`…000004_add_unique_org_event_to_metric_events.php:16`, `IngestBatchService.php:92`). Longitud de la columna: la del `string` de Laravel; **no comprobada en la base de datos** |

**M4 · Catálogo de la plataforma** (`permea-platform/backend/config/pricing.php` en `main` = `865bba0`).
Cabecera: `FUENTE: https://platform.claude.com/docs/en/about-claude/pricing` · `VERIFICADA: 2026-08-07 ·
APROBADA POR: Basilio, 2026-08-07` (`:8-9`). La fila de `claude-opus-5-5` se verificó y aprobó el
2026-10-02 (comentario sobre `:117`). USD por millón de tokens; la escritura de caché va a la tarifa de
5 minutos.

| Clave | input | output | cache_write | cache_read | Agente hoy |
|---|---|---|---|---|---|
| `claude-fable-5` | 10.00 | 50.00 | 12.50 | 1.00 | — |
| `claude-mythos-5` | 10.00 | 50.00 | 12.50 | 1.00 | — |
| `claude-opus-5-5` | 4.00 | 20.00 | 5.00 | 0.20 | — |
| `claude-opus-5` | 5.00 | 25.00 | 6.25 | 0.50 | — |
| `claude-opus-4-8` | 5.00 | 25.00 | 6.25 | 0.50 | — |
| `claude-opus-4-7` | 5.00 | 25.00 | 6.25 | 0.50 | — |
| `claude-opus-4-6` | 5.00 | 25.00 | 6.25 | 0.50 | **15 / 75 / 18,75 / 1,5** ❌ |
| `claude-opus-4-5` | 5.00 | 25.00 | 6.25 | 0.50 | — |
| `claude-opus-4-1` | 15.00 | 75.00 | 18.75 | 1.50 | — |
| `claude-opus-4` | 15.00 | 75.00 | 18.75 | 1.50 | — |
| `claude-sonnet-5` ⚠️ Q-006-1 | 3.00 | 15.00 | 3.75 | 0.30 | — |
| `claude-sonnet-4-6` | 3.00 | 15.00 | 3.75 | 0.30 | 3 / 15 / 3,75 / 0,3 ✅ |
| `claude-sonnet-4-5` | 3.00 | 15.00 | 3.75 | 0.30 | — |
| `claude-sonnet-4` | 3.00 | 15.00 | 3.75 | 0.30 | — |
| `claude-haiku-4-5` | 1.00 | 5.00 | 1.25 | 0.10 | 1 / 5 / 1,25 / 0,1 ✅ |
| `claude-haiku-3-5` | 0.80 | 4.00 | 1.00 | 0.08 | — |

**16 claves.** Las tres del agente están entre ellas. `excluded` está vacío (`:271`). Limitaciones que
el catálogo declara:
- escritura de caché de 1 hora no distinguida (`:16-24`);
- «modo rápido» de Opus 5.5 a 8.00 / 40.00 no distinguido (`:114-116`);
- `claude-sonnet-5` a precio estándar, con el introductorio de 2.00 / 10.00 vigente hasta el 2026-08-31 (`:190-196`).

**Otros puntos de partida** (del descubrimiento):
- la ayuda (`report-agente.md` §10);
- los 7 avisos de `golangci-lint`, todos en código de 005: 4 `errcheck` en `cmd/permea/project.go`, 2 `revive` y 1 `staticcheck` (§3);
- las URLs `bfgnet/…` del README, que dan 404 (§1);
- la ausencia de CHANGELOG (§1).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Un mensaje, un evento (Priority: P1)

Alguien usa Claude Code y el agente mide. Cada respuesta del modelo es **un** consumo, y aparece **una
vez** en la plataforma, con sus tokens contados una sola vez. Da igual que el log la escriba en varias
líneas, que el agente la lea en dos pasadas, que se pierda el estado local y se relea todo, o que la
lean dos instalaciones de la misma organización.

**Why this priority**: es la cifra del producto. Con el defecto actual **todo lo demás miente en
proporción**: presupuestos, chargeback por proyecto, comparativas entre modelos. Tarifas perfectas
multiplicadas por ×2,13 siguen dando un coste falso. Va primero porque es lo único que hace verdaderas
las otras historias.

**Independent Test**: sobre una copia congelada de los logs reales, el dry-run emite **tantos eventos
como mensajes distintos** (sin `<synthetic>`), y la suma de sus tokens coincide con contar cada mensaje
una vez (P-006 SC-001, SC-002). Dos lecturas independientes de la misma copia, con estado, directorio de
datos y secreto local distintos, producen **el mismo conjunto** de `event_id` (P-006 SC-003).

**Acceptance Scenarios**:

1. **Given** un log en el que un mensaje ocupa tres líneas con el mismo par (`message.id`,
   `requestId`), **When** el agente lo procesa, **Then** produce **un** evento, con el consumo de ese
   mensaje contado una vez.
2. **Given** un mensaje del que el agente leyó la primera línea en una pasada y las demás en la
   siguiente, **When** la plataforma recibe lo emitido en las dos pasadas, **Then** el mensaje cuenta
   **una vez**, porque todo lo emitido para él lleva el mismo `event_id`.
3. **Given** una instalación que pierde su estado de lectura y relee todos los logs, **When** emite,
   **Then** cada `event_id` coincide con el que ya emitió para el mismo mensaje, y la plataforma lo
   descarta como repetido.
4. **Given** dos instalaciones distintas, con secretos locales distintos, que leen el mismo log,
   **When** emiten a la misma organización, **Then** el mensaje cuenta una vez.
5. **Given** una línea con `model=<synthetic>`, **When** el agente la procesa, **Then** no produce
   evento.
6. **Given** una instalación de una versión anterior, con estado de lectura y eventos en cola, **When**
   se actualiza a esta versión y hace una pasada, **Then** no relee lo ya leído, y lo que estaba en cola
   se envía tal cual, con su `event_id` original.
7. **Given** cualquier evento emitido, **When** se inspecciona su contenido, **Then** no contiene en
   claro ni el `message.id` ni el `requestId` del proveedor, y su forma es la del contrato de frontera
   vigente.

---

### User Story 2 - Tarifas fieles (Priority: P1)

Quien mira el coste en la plataforma ve, para cada evento de un modelo conocido, **el coste a la tarifa
vigente**, calculado en local, como manda la constitución. Para un modelo desconocido ve «no
disponible», no un cero que parezca gratuito. Y quien mantiene el agente sabe **de dónde sale cada
cifra**, cuándo se verificó y quién la aprobó.

**Why this priority**: es P1 porque hoy el 100 % de los eventos recientes sale sin coste, y la única
fila de Opus triplica su precio. El coste es lo que el agente promete medir; si no lo da, no queda
producto. Va detrás de US1 sólo en el orden de lectura: las dos son condición para publicar.

**Independent Test**: se compara la tabla empaquetada con el catálogo de la plataforma de M4: 16 claves
y 64 cifras, todas iguales (P-006 SC-009). Sobre la copia congelada de los logs, todos los eventos de
modelos con fila salen con coste disponible (P-006 SC-010).

**Acceptance Scenarios**:

1. **Given** un evento de `claude-opus-5-5`, **When** se calcula su coste, **Then** se aplican 4.00 /
   20.00 / 5.00 / 0.20 USD por millón a entrada, salida, escritura de caché y lectura de caché.
2. **Given** un evento de `claude-opus-4-6`, **When** se calcula su coste, **Then** se aplican 5.00 /
   25.00 / 6.25 / 0.50, no 15 / 75.
3. **Given** un evento de un modelo sin fila, **When** se calcula su coste, **Then** sale con
   `cost_available=false` y `cost_usd=0`, y sus tokens se cuentan igual.
4. **Given** un modelo que llega con sufijo de fecha (p. ej. `claude-haiku-4-5-20251001`), **When** se
   busca su tarifa, **Then** **no** casa con `claude-haiku-4-5`. Es un límite declarado, no un fallo.
5. **Given** que alguien cambia una cifra o una clave de la tabla, **When** se ejecuta la suite,
   **Then** al menos un test se pone rojo.

---

### User Story 3 - La ayuda dice la verdad, y pedirla no ejecuta nada (Priority: P2)

Alguien acaba de instalar y no sabe qué hacer. Escribe `permea --help`, `permea -h` o `permea help`, y
recibe la ayuda completa: los subcomandos y las opciones. Escribe `permea project join -h` para saber
cómo se usa, y recibe la ayuda de ese subcomando, **sin que se emita nada ni se toque nada**. Si se
equivoca de subcomando, el error se lo dice.

**Why this priority**: P2 porque las historias P1 entregan su valor sin ésta. Pero es lo primero que ve
un usuario nuevo, y hoy falla en las dos direcciones. **Por omisión**: `--help` no muestra `enroll`,
justo lo que la aplicación le pide. **Por acción**: `project join -h` llega a emitir una petición
autenticada con `-h` como código (`report-agente.md` §10).

**Independent Test**: las cuatro formas de pedir la ayuda general (la invocación sin argumentos
incluida, D-006-7) dan salidas idénticas y completas, por stdout y con salida 0. Las ayudas de los subcomandos, lanzadas contra un directorio de datos vacío y sin
red, no crean nada ni contactan con nadie (P-006 SC-012, SC-013).

**Acceptance Scenarios**:

1. **Given** el binario, **When** se ejecuta `permea -h`, `permea --help`, `permea help` o `permea`
   sin argumentos, **Then** las cuatro imprimen la misma ayuda —`enroll`, `status`, `project join` y
   las cuatro opciones— por stdout, con salida 0. *(Enmendado 2026-10-02, D-006-7: se añade la
   invocación sin argumentos.)*
2. **Given** el binario, **When** se ejecuta `permea enrol`, **Then** falla nombrando `enrol` como
   subcomando desconocido, con el mismo código de salida que `permea project <verbo desconocido>`.
3. **Given** un agente enrolado, dentro de un árbol de proyecto, **When** se ejecuta
   `permea project join -h`, **Then** muestra la ayuda de `project join`, no carga la configuración, no
   crea directorios ni ficheros y no emite ninguna petición.
4. **Given** un agente enrolado, **When** se ejecuta `permea status -h`, **Then** muestra la ayuda de
   `status` y no el estado.
5. **Given** un agente enrolado, **When** se recorre cualquier invocación de ayuda o de error de uso,
   **Then** el token no aparece en ninguna salida.

---

### User Story 4 - Instalar siguiendo el README funciona (Priority: P2)

Alguien llega al repositorio, copia el comando de instalación de su sistema, y el agente se instala.
Después, el mismo README le dice los tres pasos siguientes: enrolar, comprobar el estado, medir. Quien
ya usaba la `0.2.1` encuentra en un CHANGELOG qué ha cambiado y por qué sus cifras bajan.

**Why this priority**: P2 por la misma razón que US3: sin ella, la versión es correcta pero no se puede
adoptar. Hoy dos de los tres comandos de instalación apuntan a repositorios que no responden. Y quien
compare sus cifras de antes y después verá que bajan a menos de la mitad. Sin una explicación escrita,
eso parece una avería.

**Independent Test**: cada comando de instalación del README se resuelve contra el repositorio real, y
la versión publicada aparece en cada canal (P-006 SC-016, SC-019). El CHANGELOG existe y contiene los
tres avisos exigidos (P-006 SC-017).

**Acceptance Scenarios**:

1. **Given** el README, **When** se sigue el comando de Homebrew, el de Scoop o el de `install.sh`,
   **Then** cada uno apunta a un repositorio de `permea-dev` que existe y sirve la versión publicada.
2. **Given** una instalación recién hecha, **When** se sigue la sección de primeros pasos, **Then** el
   orden es `enroll` → `status` → `-run` o `-daemon`, y el README avisa **antes** de la primera pasada
   de que envía todo el historial que conserve Claude Code. *(Enmendado 2026-10-02, D-006-10.)*
3. **Given** quien venía de la `0.2.1`, **When** lee el CHANGELOG de la `0.3.0`, **Then** encuentra lo
   nuevo (003, 004, 005 y 006), la ruptura aceptada de `project_ref` (004) y que **las cifras bajan
   porque antes se contaba de más**.

---

### Cierre de la feature - Publicar `v0.3.0` y ensayarla donde falló

No es una historia de usuario, sino la condición de cierre: la feature **no está terminada hasta que la
versión está publicada** y el dueño la ha ensayado en Windows **dos veces**:

1. **Antes de etiquetar** *(añadido el 2026-10-02, D-006-12)*, con el binario de Windows del
   `--snapshot`: `permea --version` → `enroll` por stdin → `status` → `--scan` sobre un log real de
   Windows. Si algo falla, se corrige **antes** de la etiqueta (P-006 FR-034, SC-022). Windows es donde
   apareció el problema y donde nada de 003–005 se ha ejecutado nunca: descubrirlo después de publicar
   costaría una versión.
2. **Después de publicar**, por el mismo camino que siguió el usuario del 2026-10-02:
   `scoop update` → `permea --version` → `enroll` por stdin → `status` → `-run` → el consumo aparece
   **una sola vez** en la plataforma (P-006 FR-032, SC-020).

---

### Edge Cases

- **¿Qué pasa si las líneas de un mismo mensaje traen consumos distintos?** Hoy no ocurre: 0 de 7 612
  (M2). Pero el formato del log no es nuestro. **Regla: cuenta la primera línea leída del mensaje, y
  NUNCA se suman las líneas** (P-006 FR-005). Se elige «la primera» porque es lo que el agente puede
  sostener leyendo de forma incremental: cuando llega la primera línea no sabe si habrá más, y las que
  lleguen en otra pasada llevarán el mismo `event_id`. La plataforma conserva el primero que recibe
  (`ON CONFLICT DO NOTHING`), así que las dos puntas aplican la misma regla. La discrepancia **se hace
  visible** en local para que no pase en silencio.
- **¿Y una línea facturable sin `message.id` o sin `requestId`?** Hoy no hay ninguna fuera de
  `<synthetic>` (M1). **Regla** (P-006 FR-006):
  - si falta uno, el identificador se deriva del que queda, de forma que no pueda coincidir con uno
    derivado del par;
  - si faltan los dos, la línea **no se emite**, y se cuenta como no contable en el diagnóstico de la
    pasada.

  Por qué así:
  - Con uno solo hay identidad suficiente: M2 mide una correspondencia 1:1 entre los dos.
  - Sin ninguno no queda nada estable de lo que derivar. Inventar un identificador aleatorio
    reintroduciría el defecto que esta feature corrige, y derivarlo del contenido violaría la frontera.
  - Se acepta perder esa línea **a la vista** en lugar de duplicarla en silencio.
- **¿Y un mensaje que cruza la actualización?** Si la versión anterior leyó su primera línea (con
  identificador aleatorio) y esta versión lee las siguientes, ese mensaje cuenta **dos veces, y una sola
  vez en la vida de la instalación**. Es un residuo acotado que se agota solo, como el del `salt` de
  005. Se declara en lugar de construir una migración para él.
- **¿Y si el mismo mensaje llega a dos organizaciones?** Cuenta una vez en cada una, porque la
  plataforma deduplica por (`org_id`, `event_id`). Es lo correcto: son dos tenants.
- **¿Qué revela un `event_id` determinista y sin sal?** Que dos eventos con el mismo `event_id` son el
  mismo mensaje, **desde cualquier instalación**: es justamente lo que se busca. Quien ya tenga el log
  original podría recalcularlo y comprobar que un evento corresponde a un mensaje suyo. Quien no lo
  tenga no puede obtener de él el `message.id` ni el `requestId`. La decisión es del dueño (D-006-2) y
  se analiza en §Verificación de la frontera.
- **¿Qué pasa con el coste de lo ya emitido sin tarifa?** El agente fija el coste **al ingerir**. Lo ya
  encolado o enviado sin coste no se recalcula al actualizar. Que lo repare la plataforma es asunto
  suyo, y queda fuera de alcance aquí.
- **¿Y si una instalación anterior tiene eventos en cola con `dev_id` vacío o con el `project_ref`
  antiguo?** Salen tal cual (P-006 FR-009). Reetiquetarlos sería reescribir eventos, y esta feature no
  reescribe nada.
- **¿Y `permea enroll -h` cuando alguien quiere enrolar con un secreto que empieza por `-h`?** No puede
  ocurrir: un *enrollment string* válido empieza por `pmea2.` y un código de adhesión por `pmeaj1.`.
  Reservar `-h` y `--help` no le quita a nadie un valor legítimo.
- **¿Y si alguien escribe por error el secreto como subcomando (`permea pmea2.…`, o un `pmea1.…`
  antiguo)?** El error de subcomando desconocido **no lo reproduce** (P-006 FR-022; `pmea1.` añadido
  el 2026-10-02 por D-006-9). Nombrar lo tecleado es útil para una
  errata; repetir un secreto en pantalla, nunca.

---

## Requirements *(mandatory)*

### Functional Requirements

#### Un mensaje, un evento

- **P-006 FR-001**: Las líneas facturables del log que comparten el par (`message.id`, `requestId`)
  **DEBEN** producir **exactamente un** evento.
- **P-006 FR-002**: El `event_id` de un mensaje **DEBE** ser **determinista**. Se deriva por hash **sin
  sal ni ningún dato local** del par (`message.id`, `requestId`), de modo que el mismo mensaje dé el
  mismo `event_id` en cualquier instalación, máquina y pasada.
- **P-006 FR-003**: Ni el `message.id` ni el `requestId` **NUNCA** cruzan la frontera en claro, ni en el
  `event_id` ni en ningún otro campo. Tampoco **DEBEN** poder obtenerse a partir del `event_id` sin
  disponer ya de ellos.
- **P-006 FR-004**: El `event_id` **DEBE** conservar **su forma actual: 32 caracteres hexadecimales
  `[0-9a-f]`** (M3), que cumple el contrato de frontera (cadena hexadecimal, `boundary-event.md:25`) y
  cabe en lo que la plataforma acepta hoy. Pares distintos **DEBEN** dar `event_id` distintos, salvo
  colisión criptográficamente despreciable. *(Enmendado 2026-10-02, D-006-8: antes decía «cadena
  hexadecimal no vacía» y dejaba la longitud al plan.)*
- **P-006 FR-005**: Si las líneas de un mismo mensaje traen consumos distintos, el evento **DEBE**
  reflejar el de **la primera línea leída**, y las líneas **NUNCA** se suman. La discrepancia **DEBE**
  quedar visible en el diagnóstico local de la pasada.
- **P-006 FR-006**: Una línea facturable con **sólo uno** de los dos identificadores **DEBE** derivar su
  `event_id` del que tiene, sin posibilidad de coincidir con un `event_id` derivado del par. Una línea
  **sin ninguno** **NUNCA** se emite, y **DEBE** contarse como no contable en el diagnóstico local de la
  pasada.
- **P-006 FR-007**: Las líneas con `model=<synthetic>` **NUNCA** producen evento.
- **P-006 FR-008**: El agente **NUNCA** acuña un `event_id` aleatorio para una línea facturable. La
  única fuente de identidad es la de FR-002 y FR-006.
- **P-006 FR-009**: Actualizar una instalación existente **NUNCA** reenvía historial. Lo ya leído no se
  relee por el mero hecho de actualizar, y lo que estaba en cola se envía **tal cual**, con su `event_id`
  original.
- **P-006 FR-010**: El dry-run (`--scan`) **DEBE** aplicar las mismas reglas (FR-001, FR-005, FR-006,
  FR-007) y **DEBE** mostrar por evento sus **cuatro** partidas de tokens. Es el instrumento con el que
  se miden P-006 SC-001 y SC-002 sin transmitir nada.
- **P-006 FR-011**: Los contratos de 001 que describen el origen del `event_id`
  (`contracts/transport.md:47-48`, `data-model.md:19`) **DEBEN** redescribirse para decir que es
  determinista, como hizo 004 con el contrato de frontera. La forma del campo **no** cambia.
- **P-006 FR-033** *(nuevo, 2026-10-02, D-006-11)*: **Dentro de una misma pasada**, el agente **NUNCA**
  encola dos veces el mismo `event_id`. **Entre pasadas** no guarda memoria de lo ya encolado: un
  mensaje que reaparezca en otra pasada puede volver a encolarse, y **se confía en la plataforma**, que
  descarta repetidos por (`org_id`, `event_id`). Así la cola no crece por las líneas repetidas de un
  mensaje, y no hace falta un estado nuevo que pueda corromperse o perderse.

#### La frontera no se toca

- **P-006 FR-012**: La allowlist del evento **NO** cambia: los mismos 17 campos, con los mismos nombres
  y tipos. `SchemaVersion` sigue en `1`. **Restricción del dueño (D-006-3)**: `internal/event` no se
  modifica.
- **P-006 FR-013**: El golden test de frontera **DEBE** seguir en verde y **DEBE** ampliarse antes de la
  implementación (Principio IV), con `message.id` y `requestId` como centinelas: ninguno de los dos, ni
  entero ni como fragmento reconocible, puede aparecer en la salida. Leer esos dos campos del log es una
  ampliación de lo que el lector decodifica, y se declara aquí para que no entre de pasada
  (`internal/ingest/claudecode.go:17-24`).

#### Tarifas

- **P-006 FR-014**: La tabla empaquetada **DEBE** ser **espejo exacto** del catálogo de la plataforma
  (M4): las mismas 16 claves y las mismas cuatro cifras por clave. Es decisión del dueño (D-006-4): una
  sola verdad de precios, replicada, no dos verdades que se parezcan. *(Nota 2026-10-02, Q-006-1: la
  fila de `claude-sonnet-5` está pendiente de decisión del dueño. Si el catálogo de la plataforma
  cambia, este requisito replica **el commit nuevo**, y M4 se rehace sobre él.)*
- **P-006 FR-015**: La tabla **DEBE** llevar en su cabecera la **fuente** (la misma URL que el
  catálogo), la **fecha** de verificación, la **aprobación** (quién y cuándo) y **qué versión del
  catálogo replica** (repositorio, fichero y commit).
- **P-006 FR-016**: Como consecuencia de FR-014, `claude-opus-4-6` queda en 5.00 / 25.00 / 6.25 / 0.50.
  Se enuncia aparte porque es la única cifra existente que **cambia**, y el test actual fija la errónea.
- **P-006 FR-017**: Un modelo sin fila **DEBE** seguir produciendo `cost_available=false` y
  `cost_usd=0`, con los tokens contados (comportamiento de 001, sin cambios).
- **P-006 FR-018**: La tarifa se busca por **coincidencia exacta** del identificador de modelo. No se
  normalizan sufijos de fecha, prefijos ni mayúsculas. Es un **límite declarado** (en la cabecera y en
  el README), coherente con el catálogo, cuya clave es el identificador «tal como llega en el evento».
- **P-006 FR-019**: La cabecera **DEBE** declarar las limitaciones que hereda del catálogo:
  - la escritura de caché va a la tarifa de **5 minutos**, y una de 1 hora quedaría infravalorada;
  - el **«modo rápido»** no se distingue, y quedaría infravalorado;
  - `claude-sonnet-5` va a **precio estándar**, y los eventos anteriores al 2026-09-01 quedarían
    sobrevalorados. *(Depende de Q-006-1: si el catálogo cambia esa fila, esta limitación se revisa
    con él.)*
- **P-006 FR-020**: La tabla **DEBE** quedar **vigilada clave a clave**: añadir, quitar o cambiar
  cualquier clave o cualquier cifra **DEBE** poner rojo al menos un test.

#### La ayuda y los errores de uso

- **P-006 FR-021**: `permea -h`, `permea --help`, `permea help` y `permea` **sin argumentos** **DEBEN**
  producir **la misma** ayuda completa —los subcomandos `enroll`, `status` y `project join`, y las
  opciones `--scan`, `--run`, `--daemon` y `--version`— por **stdout**, con salida **0**. La ayuda
  **DEBE** tener **una sola fuente**, para que no vuelva a haber dos textos que diverjan. Y **DEBE**
  recomendar la vía stdin para los dos valores sensibles (`003/contracts/cli.md:21`,
  `005/contracts/cli.md:60`). *(Enmendado 2026-10-02, D-006-7: se añade la invocación sin argumentos,
  que hoy imprime la ayuda por stderr. El cambio de canal va al CHANGELOG, FR-028.)*
- **P-006 FR-022**: Un subcomando inexistente **DEBE** fallar con un mensaje que lo **nombre**, con el
  **mismo código de salida** que ya usa `permea project <verbo desconocido>` (1,
  `cmd/permea/project.go:55-56`). Si lo tecleado tiene la forma de un secreto conocido (`pmea2.`,
  `pmeaj1.` o `pmea1.`), el mensaje **NUNCA** lo reproduce. *(Enmendado 2026-10-02, D-006-9: se añade
  `pmea1.`. Que esté rechazado como enrolamiento no lo deja de convertir en un secreto.)*
- **P-006 FR-023**: `enroll -h`, `status -h`, `project -h` y `project join -h`, y sus formas `--help`,
  **DEBEN** mostrar la ayuda del subcomando por stdout, con salida 0. **NUNCA** deben cargar la
  configuración, crear directorios o ficheros, ni emitir peticiones.
- **P-006 FR-024**: **Invariante que se conserva**: el token **NUNCA** se imprime por ningún camino,
  incluidos los mensajes nuevos de esta feature. Lo mismo vale para el *enrollment string* y el código
  de adhesión.

#### README, CHANGELOG y comentarios

- **P-006 FR-025**: Los comandos de instalación del README (Homebrew, Scoop e `install.sh`) **DEBEN**
  apuntar a `permea-dev/…`, y cada uno **DEBE** haberse comprobado contra el repositorio real antes de
  publicar.
- **P-006 FR-026**: El README **DEBE** incluir unos primeros pasos **para quien instala**: `enroll` (por
  stdin) → `status` → `-run` o `-daemon`. Los pasos de desarrollo (`make test`, etc.) quedan
  diferenciados. **DEBE avisar**, antes de la primera pasada, de que **la primera pasada tras instalar
  envía todo el historial que conserve Claude Code**. *(Enmendado 2026-10-02, D-006-10.)*
- **P-006 FR-027**: El README **NUNCA** menciona el «modo de ref» retirado (deuda anotada en
  `specs/005-adhesion-a-proyecto/tasks.md:2251-2257`), y **DEBE** declarar el límite de FR-018.
- **P-006 FR-028**: **`CHANGELOG.md` DEBE nacer con la `0.3.0`** e incluir:
  - lo nuevo de 003, 004, 005 y 006;
  - la **ruptura aceptada** de la continuidad de `project_ref` (`specs/004-identidad-de-proyecto/spec.md:424-427`);
  - que **las cifras bajan respecto a la `0.2.1` porque antes se contaba de más**;
  - que **`permea` sin argumentos imprime ahora la ayuda por stdout** y no por stderr *(añadido el
    2026-10-02, D-006-7)*.
- **P-006 FR-029**: Los comentarios que dicen `bfgnet/…` en `.github/workflows/release.yml` y
  `.goreleaser.yaml` **DEBEN** corregirse **sin tocar ninguna línea de configuración**: el diff de esos
  dos ficheros **sólo** contiene comentarios.

#### Publicación

- **P-006 FR-030**: La feature **DEBE** cerrarse publicando **`v0.3.0`** con el procedimiento vigente
  (`specs/002-distribucion/quickstart.md:98-110`), sobre `main` con el árbol limpio.
- **P-006 FR-031**: Antes de etiquetar **DEBEN** estar en verde, en local:
  - `gofmt`, `go vet` y la suite completa;
  - `golangci-lint` **limpio, con 0 avisos**, como exige la constitución (`constitution.md:74`);
  - `goreleaser check` y un `--snapshot`.

  *(Enmendado 2026-10-02, D-006-13: antes decía «`golangci-lint`, con la regla de la excepción
  D-006-5: el recuento no aumenta, y el código nuevo o modificado por 006 entra limpio». D-006-5 queda
  revocada.)*
- **P-006 FR-032**: La feature **DEBE** cerrarse con el ensayo del dueño en Windows **tras publicar**,
  descrito en §Cierre de la feature (paso 2), con su resultado anotado.
- **P-006 FR-034** *(nuevo, 2026-10-02, D-006-12)*: **Antes de etiquetar**, el dueño **DEBE** ensayar
  en Windows el binario de Windows del `--snapshot`: `--version`, `enroll` por stdin, `status` y
  `--scan` sobre un log real de Windows. El resultado se anota. **Si algún paso falla, se corrige antes
  de la etiqueta**, y el ensayo se repite con el snapshot corregido. No sustituye al ensayo tras
  publicar (FR-032): se suma a él.

### Key Entities

- **Mensaje facturable**: la unidad de consumo. Una respuesta del modelo, identificada en el log por el
  par (`message.id`, `requestId`), que puede ocupar una o varias líneas. Es lo que un evento representa.
- **Evento**: el único dato que cruza la frontera. No cambia de forma; cambia **qué representa**: un
  mensaje, ya no una línea.
- **Identificador de evento (`event_id`)**: clave de deduplicación. Pasa de aleatoria a **derivada del
  mensaje**, sin sal y sin revelar sus identificadores.
- **Tabla de tarifas**: precio por millón de tokens en cuatro partidas, por modelo, empaquetado en el
  binario. Es espejo del catálogo de la plataforma y lleva su procedencia en la cabecera.
- **Ayuda**: el texto que describe la gramática del binario. Una sola fuente, que se sirve de la misma
  forma a todas las maneras de pedirla.

---

## Verificación de la frontera

La constitución exige que toda especificación verifique la frontera de forma explícita
(`.specify/memory/constitution.md:83`). Se verifica aquí.

- **La allowlist no cambia** (FR-012). Ningún campo nuevo, ninguno renombrado, el mismo `SchemaVersion`.
- **Lo que cruza de nuevo es un valor, no un campo.** El `event_id` sigue siendo el mismo campo; lo que
  cambia es cómo se calcula. Pasa de no derivar de nada a derivar de dos identificadores técnicos del
  proveedor, que **no son contenido**: no son prompts, respuestas, código, rutas ni nombres. Y **no
  cruzan en claro** (FR-003).
- **Por qué sin sal, siendo la sal la norma de la frontera.** El Principio I exige hash salado para «los
  identificadores sensibles (ruta de proyecto, sesión, máquina)» (`constitution.md:14`). El `event_id`
  no es uno de ellos, y su función exige lo contrario de la sal: que dos instalaciones produzcan **el
  mismo** valor para el mismo mensaje. Con sal no habría deduplicación entre instalaciones, ni tras
  perder el estado. La decisión es del dueño (D-006-2). Su coste, escrito:
  - quien posea el log original puede recalcular el `event_id` y reconocer el evento como suyo;
  - quien no lo posea no obtiene de él nada del mensaje.
- **El lector decodifica dos campos más** (FR-013). La guarda de `claudecode.go:17-24` admite
  «métricas y metadatos»; estos dos son metadatos técnicos, se usan sólo para derivar y deduplicar, y el
  golden test los incorpora como centinelas antes de que exista el código que los lee.

---

## Success Criteria *(mandatory)*

Los criterios que se miden sobre logs reales se miden sobre **una copia congelada**, fechada y con su
recuento de ficheros anotado. El historial vivo cambia mientras se mide, y un criterio sobre un
blanco móvil no se puede poner rojo. Las cifras de M1 y M2 son la referencia de la fecha de esta spec,
no el valor esperado en otra copia.

### Measurable Outcomes

#### Un mensaje, un evento

- **P-006 SC-001**: Sobre la copia congelada, **el número de eventos que emite el dry-run es igual al
  número de mensajes distintos** (`message.id`, `requestId`) con `model` no vacío y distinto de
  `<synthetic>`. Ese número lo cuenta un recuento independiente que lee sólo identificadores. Referencia
  en la fecha de la spec: **10 698** mensajes frente a 22 857 líneas, cuando hoy se emitirían 22 860
  eventos. **Falsable**: duplicada a propósito una línea de un mensaje en la copia, el recuento de
  eventos **no** cambia; quitado el último par de un mensaje, baja en uno.
- **P-006 SC-002**: Sobre la misma copia, **la suma de las cuatro partidas de tokens de los eventos del
  dry-run es igual a la suma contando cada mensaje una vez**. Referencia: **4 835 167 368**, frente a
  10 292 101 351 sumando por línea. Se mide con el recuento independiente de SC-001.
- **P-006 SC-003**: **Determinismo.** Dos lecturas independientes de la misma copia, con estado de
  lectura vacío, directorios de datos distintos y **secretos locales distintos**, producen
  **exactamente el mismo conjunto** de `event_id`. El número de `event_id` distintos es igual al de
  mensajes de SC-001, sin colisiones. **Todos** tienen exactamente 32 caracteres `[0-9a-f]` (FR-004).
  **Falsable**: alterado un solo carácter del `message.id` de una línea, su `event_id` cambia, y sólo el
  suyo. *(Enmendado 2026-10-02, D-006-8: se añade la comprobación de forma.)*
- **P-006 SC-004**: **Nada del proveedor cruza en claro.** El golden test ampliado (FR-013) está en
  verde con los centinelas. Sobre la copia congelada, **ninguno** de los `message.id` ni `requestId`
  aparece en la salida del dry-run ni en ningún evento serializado (búsqueda literal: 0 apariciones).
- **P-006 SC-005**: **La frontera no cambia.** `git diff` de `internal/event` contra `0311fa1`
  **vacío**. El test de la allowlist (`TestEvent_OnlyAllowlistKeys`) en verde sin modificarlo.
  `SchemaVersion == 1`.
- **P-006 SC-006**: **`<synthetic>` no se emite.** Sobre la copia congelada, 0 eventos con ese modelo.
  Referencia: 3 líneas en la fecha de la spec.
- **P-006 SC-007**: **Actualizar no reenvía.** Con un directorio de datos preparado por la versión
  anterior —estado de lectura a mitad de un log y eventos en cola—, una pasada de esta versión:
  - encola **sólo** lo que hay más allá de los offsets guardados;
  - deja los eventos que ya estaban en cola **byte a byte** iguales, `event_id` incluido.

  Se ejercita en un sandbox, nunca sobre el directorio de datos real.
- **P-006 SC-008**: **Las reglas de los casos límite tienen test.** Hay un caso, en rojo antes de su
  implementación, para cada regla:
  - (a) consumo distinto entre líneas: cuenta la primera, no se suma, se informa;
  - (b) un solo identificador: `event_id` estable y distinto del derivado del par;
  - (c) ningún identificador: no se emite, se cuenta.
- **P-006 SC-021** *(nuevo, 2026-10-02, D-006-11)*: **Deduplicación local dentro de la pasada.** En un
  sandbox, una pasada sobre un log con un mensaje de tres líneas deja en la cola **un** evento para él.
  Sobre la copia congelada, una pasada completa deja en la cola **tantos eventos como mensajes de
  SC-001**, sin dos con el mismo `event_id`. **Entre pasadas**: tras perder el estado y repetir la
  pasada, la cola puede contener el `event_id` repetido, y un servidor de prueba con la regla de la
  plataforma (`ON CONFLICT DO NOTHING` por `(org_id, event_id)`) registra el mensaje **una vez**.
  **Falsable**: desactivada la deduplicación local, el primer caso deja tres eventos y se pone rojo.

#### Tarifas

- **P-006 SC-009**: **Espejo exacto.** La tabla empaquetada tiene **16/16** claves y **64/64** cifras
  iguales a `pricing.php@865bba0`, comprobado fila a fila (M4). **Falsable**: alterada cualquier cifra
  o quitada cualquier clave, al menos un test se pone rojo (FR-020). *(Nota 2026-10-02, Q-006-1: si el
  catálogo cambia, el commit de referencia y los recuentos se toman del nuevo.)*
- **P-006 SC-010**: **Coste disponible para lo que se usa.** Sobre la copia congelada, el **100 %** de
  los eventos de modelos con fila sale con `cost_available=true`. En la fecha de la spec son
  `claude-opus-5-5`, `claude-opus-5` y `claude-sonnet-5`, y hoy salen al **0 %**. El coste de un evento
  de referencia de cada uno coincide con el cálculo a mano a la tarifa de M4.
- **P-006 SC-011**: **La cabecera está completa.** Contiene la fuente, la fecha de verificación, la
  aprobación, la versión replicada del catálogo y las tres limitaciones de FR-019. Lo comprueba la
  revisión, con una casilla por elemento.

#### La ayuda y los errores de uso

- **P-006 SC-012**: `permea -h`, `permea --help`, `permea help` y `permea` sin argumentos producen
  salidas por stdout **no vacías**, **idénticas byte a byte entre sí**, con stderr vacío y salida 0.
  Cada salida contiene `enroll`, `status`, `project join` y las cuatro opciones. **Falsable**: una
  diferencia deliberada en una de las cuatro pone roja la comparación. *(Enmendado 2026-10-02,
  D-006-7: de tres invocaciones a cuatro.)*
- **P-006 SC-013**: Las ayudas de subcomando (FR-023), ejecutadas con un HOME de sandbox **vacío**, con
  un agente de prueba enrolado contra un servidor de prueba y desde dentro de un árbol de proyecto:
  - el árbol del sandbox queda **idéntico** antes y después (0 ficheros y 0 directorios creados);
  - el servidor de prueba recibe **0** peticiones;
  - la salida es 0.
- **P-006 SC-014**: `permea enrol` sale con **1** y nombra `enrol`. `permea pmea2.<centinela>`,
  `permea pmeaj1.<centinela>` y `permea pmea1.<centinela>` salen con 1, y su salida **no** contiene el
  centinela. *(Enmendado 2026-10-02, D-006-9: se añaden `pmea1.`, y `pmeaj1.`, que FR-022 ya cubría
  pero el criterio no comprobaba.)*
- **P-006 SC-015**: Con un `config.json` de prueba enrolado con un token **centinela**, ninguna
  invocación del conjunto —las cuatro ayudas generales (la invocación sin argumentos incluida, D-006-7),
  las ocho de subcomando, `status` y subcomando inexistente— contiene el centinela en stdout ni en
  stderr.

#### README, CHANGELOG y comentarios

- **P-006 SC-016**: Cada comando de instalación del README se resuelve contra un repositorio real de
  `permea-dev`:
  - el tap y el bucket existen;
  - la URL de `install.sh` responde;
  - **0** apariciones de `bfgnet` en `README.md`, `release.yml` y `.goreleaser.yaml`;
  - el diff de configuración de esos dos ficheros (líneas que no son comentario) está **vacío**;
  - los primeros pasos contienen el aviso de que la primera pasada envía todo el historial que conserve
    Claude Code, **antes** del paso `-run`/`-daemon` *(añadido el 2026-10-02, D-006-10)*.
- **P-006 SC-017**: `CHANGELOG.md` existe y su entrada `0.3.0` contiene, cada uno comprobable por
  separado:
  - lo nuevo de 003, 004, 005 y 006;
  - la ruptura de `project_ref`;
  - el aviso de que las cifras bajan, y por qué;
  - el cambio de canal de `permea` sin argumentos, de stderr a stdout *(añadido el 2026-10-02,
    D-006-7)*.

#### Publicación

- **P-006 SC-018**: **Puertas.** `gofmt -l .` vacío, `go vet` sin hallazgos y suite completa en verde.
  `golangci-lint run` con **0 avisos**, los 7 heredados de 005 incluidos. *(Enmendado 2026-10-02,
  D-006-13: antes decía «recuento ≤ 7 y 0 hallazgos en líneas nuevas o modificadas por 006».)*
- **P-006 SC-019**: **Publicada.**
  - `gh release view v0.3.0` muestra 5 archivos y los checksums;
  - el manifiesto del bucket y el cask están en `0.3.0`;
  - el binario descargado responde `0.3.0` a `--version`.
- **P-006 SC-020**: **Ensayo en Windows.** El dueño completa la secuencia de §Cierre de la feature. El
  número de eventos que la plataforma registra para esa instalación en la ventana del ensayo es igual al
  de mensajes distintos de sus logs en esa ventana, contados como en SC-001. Resultado anotado con fecha.
- **P-006 SC-022** *(nuevo, 2026-10-02, D-006-12)*: **Ensayo en Windows antes de la etiqueta**, con el
  binario de Windows del `--snapshot`, anotado con fecha y con el commit del snapshot:
  - `permea --version` responde la versión del snapshot (no `0.0.1-dev`);
  - `enroll` por stdin termina con éxito, y su salida no contiene el token;
  - `status` responde «enrolado contra …» con `token: configurado`;
  - `--scan` sobre un log real de Windows emite tantos eventos como mensajes distintos de ese fichero,
    contados como en SC-001.

  **La etiqueta no se crea** mientras falte alguna de las cuatro anotaciones o alguna diga «falla».

---

## Decisiones tomadas durante la especificación

Registradas aquí porque son de Basilio, están fechadas, y su ausencia haría que la spec pareciera
haber elegido sola.

- **D-006-1 · Qué entra en la `0.3.0`** (2026-10-02): contar una vez por mensaje, tarifas, ayuda y
  errores de la CLI, y README. Todo lo demás espera.
- **D-006-2 · `event_id` determinista por hash SIN sal a partir de (`message.id`, `requestId`)**
  (2026-10-02). El mismo mensaje da el mismo identificador desde cualquier instalación. Ningún
  identificador del proveedor sale en claro. Se descartó la sal porque anularía la deduplicación entre
  instalaciones y tras perder el estado, que es la razón de ser del cambio.
- **D-006-3 · La frontera no se toca** (2026-10-02): `internal/event` y `SchemaVersion` intactos.
- **D-006-4 · La tabla de tarifas es espejo de la de la plataforma** (2026-10-02): mismas claves y
  cifras, incluida `claude-opus-5-5` a 4 / 20 / 5 / 0,20, con fuente, fecha y aprobación en su cabecera.
- ~~**D-006-5 · Excepción decidida por el dueño a la puerta del linter** (2026-10-02).~~
  **REVOCADA el 2026-10-02 por D-006-13**: el linter entra limpio. Se conserva el texto original para
  que quede constancia de qué se decidió y qué se deshizo:

  La constitución
  exige `golangci-lint run` limpio (`constitution.md:74`), y hoy hay **7** avisos heredados de 005. **No
  se corrigen en esta feature.** Se registra como excepción **decidida por el dueño**, con una regla que
  impide que crezca: **el recuento no aumenta, y el código nuevo o modificado entra limpio**
  (FR-031, SC-018).
- **D-006-6 · Los tests siguen fuera de la tubería de publicación** (2026-10-02). La publicación
  continúa sin ejecutar tests ni linter en CI. Las puertas son locales y previas a la etiqueta (FR-031).
  Deuda declarada, no olvido. *(Sin cambios tras las enmiendas del 2026-10-02.)*

### Enmiendas del 2026-10-02

Decisiones del orquestador y del dueño tomadas después de la primera redacción, el mismo día. Cada una
dice qué toca. Ningún número existente se ha reasignado.

- **D-006-7 · `permea` sin argumentos se comporta como `permea help`**: misma ayuda, por stdout, con
  salida 0. Toca FR-021, FR-028, US3 (escenario 1 e Independent Test), SC-012, SC-015 y SC-017. El
  cambio de canal respecto a hoy (stderr → stdout) va al CHANGELOG. Cierra una duda que la primera
  redacción dejaba al plan.
- **D-006-8 · El `event_id` conserva su forma actual: 32 caracteres hexadecimales.** Toca FR-004 y
  SC-003. Cierra una duda que la primera redacción dejaba al plan. La plataforma no ve ningún cambio
  de forma.
- **D-006-9 · El error de subcomando desconocido tampoco reproduce `pmea1.`.** Toca FR-022, SC-014 y
  un caso límite.
- **D-006-10 · El README avisa de que la primera pasada tras instalar envía todo el historial que
  conserve Claude Code.** Toca FR-026, US4 (escenario 2) y SC-016. Quien instala debe saber que su
  primer `-run` no mide «desde ahora».
- **D-006-11 · Deduplicación local sólo dentro de la pasada; entre pasadas se confía en la
  plataforma.** Requisito nuevo **FR-033**, con su criterio **SC-021**. Cierra una duda que la primera
  redacción dejaba al plan.
- **D-006-12 · Ensayo en Windows ANTES de etiquetar**, con el binario de Windows del snapshot:
  `--version`, `enroll` por stdin, `status` y `--scan`. El resultado se anota, y si falla se corrige
  antes de la etiqueta. Requisito nuevo **FR-034**, criterio nuevo **SC-022**; toca también §Cierre de
  la feature y FR-032, que pasa a decir «tras publicar». El ensayo final tras publicar (FR-032, SC-020)
  se mantiene.
- **D-006-13 · El linter entra: `golangci-lint` limpio, 0 avisos** (decisión del dueño, 2026-10-02 por
  la tarde). **Revoca D-006-5.** Toca FR-031 y SC-018, que pasan a exigir 0, y saca «corregir los 7
  avisos» de §Fuera de alcance. Con esto la feature cumple la puerta de la constitución
  (`constitution.md:74`) sin excepción. D-006-6 no cambia.

---

## Preguntas abiertas

- **Q-006-1 · `claude-sonnet-5`: ¿2 / 10 o 3 / 15?** (registrada el 2026-10-02; **pendiente de decisión
  del dueño**).
  - **La página oficial de precios** muestra 2 / 10 / 2,50 / 0,20 (captura del dueño, 2026-10-02).
  - **El catálogo de la plataforma** (`865bba0`) dice 3 / 15 / 3,75 / 0,30, que su cabecera presenta como
    precio estándar vigente desde el 2026-09-01 (`pricing.php:190-196`).

  **Mientras no se decida**, FR-014 replica el catálogo tal cual. **Si el catálogo cambia**, FR-014
  replica el commit nuevo, M4 se rehace sobre él, SC-009 toma de él su referencia y la tercera
  limitación de FR-019 se revisa. Afecta también a SC-010, porque `claude-sonnet-5` es uno de los tres
  modelos en uso (M2: 3 mensajes).

---

## Cuestiones para el plan

No se resuelven en esta spec; quedan listadas para que el plan las tome.

- **Dónde vive la derivación del `event_id`**, dado que `internal/event` no se toca (D-006-3), y **qué
  pasa con `event.NewID`**, que deja de usarse para las líneas facturables (FR-008).
- **Cómo se vigila el espejo de tarifas sin depender del otro repositorio** (FR-020, SC-009).
- **Canal del diagnóstico local** de FR-005 y FR-006: qué se informa, dónde y cuándo.
- **Quién consume el formato del dry-run**, que FR-010 amplía a las cuatro partidas. `--scan` aparece
  en `cmd/permea/main_test.go` y en `internal/project/resolve_test.go`; sin revisar si dependen del
  formato de la línea.

---

## Assumptions

- **El formato del log de Claude Code mantiene `message.id` y `requestId`** en las líneas facturables,
  con la correspondencia 1:1 medida en M2. Si dejara de traerlos, las reglas de FR-006 deciden, y la
  pérdida queda a la vista.
- **La plataforma sigue deduplicando por (`org_id`, `event_id`)** y conservando el primero
  (`ON CONFLICT DO NOTHING`). Toda la deduplicación entre pasadas e instalaciones se apoya en eso.
  Esta feature no la cambia; en local sólo deduplica dentro de una pasada (FR-033, D-006-11).
- **El catálogo de M4 es la verdad de precios** en la fecha de la spec. Si la plataforma cambia una
  tarifa, el espejo queda desfasado hasta la siguiente versión del agente. La cabecera dice qué versión
  replica, para que el desfase sea **visible**.
- **El coste se fija al ingerir.** El agente no recalcula lo ya emitido. Que la plataforma repare costes
  en lectura es asunto suyo.
- **Los mensajes no se reparten entre ficheros** (M2: 0). Si ocurriera, FR-002 seguiría dando el mismo
  `event_id`, y la plataforma deduplicaría.

---

## Fuera de alcance

Enumerado con su motivo, para que la ausencia no se lea como olvido.

- **Los datos que la plataforma ya recibió.** Sus `event_id` son aleatorios y no hay forma de reconocer
  desde el agente qué eventos son el mismo mensaje. Corregirlos es trabajo de la plataforma, si decide
  hacerlo.
- **Instrucciones de instalación dentro de la aplicación.** Son de la plataforma. Aquí sólo se arregla
  el README.
- **Instalar el agente como servicio** (arranque automático en cada SO). Es una capacidad nueva con sus
  propias preguntas de permisos; la `0.3.0` se usa con `-run` o `-daemon`.
- **Tests en Windows dentro de CI.** Requieren tubería de CI con tests, que D-006-6 deja fuera. Los
  dos ensayos manuales del dueño, antes de etiquetar (FR-034) y tras publicar (FR-032), cubren el camino
  crítico de esta versión. *(Ajustado 2026-10-02, D-006-12.)*
- ~~**Corregir los 7 avisos de `golangci-lint`.** Ver D-006-5.~~ **Retirado de §Fuera de alcance el
  2026-10-02 por D-006-13**: los 7 avisos se corrigen en esta feature (FR-031, SC-018).
- **Normalizar sufijos de fecha o prefijos de proveedor en el identificador de modelo.** Se declara
  como límite (FR-018). La plataforma tampoco lo hace, y hacerlo sólo en el agente rompería el espejo.
- **Distinguir la escritura de caché de 1 hora o el «modo rápido».** El evento no trae la información
  necesaria, y ampliar la allowlist está fuera (D-006-3). Se declaran como limitaciones (FR-019).
- **Recalcular el coste de eventos ya encolados.** Reescribir eventos va contra FR-009.
- **Una taxonomía de códigos de salida.** Siguen siendo 0 y 1 (D-005-4). El subcomando inexistente
  reutiliza el 1.
- **Actualizar los artefactos fechados de 002** que dicen `bfgnet/…` o `Formula/permea.rb`. Son
  registro de cuando se escribieron; los que se corrigen aquí son el README y los comentarios de la
  configuración viva.
