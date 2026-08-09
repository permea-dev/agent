# Feature Specification: Identidad de proyecto — que identifique proyectos, no puntos de lanzamiento

**Feature Branch**: `004-identidad-de-proyecto`

**Created**: 2026-08-09

**Status**: Draft

**Input**: Descripción de Basilio, 2026-08-09, tras el descubrimiento del mismo día en ambos repos
(agente y plataforma). Los requisitos de esta especificación se citan como **P-004 FR-xxx**; la forma
abreviada «FR-xxx» no identifica nada, porque hay requisitos con el mismo número en otras
especificaciones de `specs/`.

---

## Contexto: una dimensión que cuenta directorios y dice contar proyectos

El agente deriva hoy la identidad de proyecto de un evento a partir del directorio de trabajo que el
log de la herramienta declara, **tal cual llega**, sin interpretarlo. Es la decisión más honesta que
se podía tomar sin observar nada: transcribir lo declarado.

El efecto medido es que la dimensión **no agrupa proyectos, agrupa puntos de lanzamiento**. Lanzar
desde un subdirectorio produce una identidad distinta que lanzar desde la raíz. Una ruta con barra
final, otra. Un enlace simbólico y su destino, dos más. En datos reales: **12 identidades de proyecto
para un solo desarrollador en una sola máquina** —con el mismo secreto local, luego la divergencia es
del directorio, no del secreto—, donde los proyectos genuinos son un puñado. La forma de esa
fragmentación es reveladora: cuatro identidades concentran el 93 % de los eventos y **seis viven un
solo día**, nacidas de trabajar puntualmente desde un directorio distinto.

Esto importa ahora y no dentro de seis meses por dos razones que se cruzan:

1. **La plataforma va a construir dinero sobre esta dimensión** —presupuestos, chargeback por
   proyecto—. Un presupuesto asignado a una identidad que en realidad es «la vez que trabajé desde
   `docs/`» no es un presupuesto: es ruido con nombre de proyecto.
2. **El producto no está desplegado en producción.** Esta es la ventana en que el contrato de
   frontera puede corregirse sin coste de compatibilidad, y **la ventana se cierra al lanzar**.

La corrección exige algo que el agente hoy no hace: **observar el sistema de ficheros**. Hasta ahora
el agente solo transcribe lo que el log declara, y esa contención era parte de su argumento. Ampliar
ese comportamiento es una decisión deliberada de esta feature, y se declara como tal —no se cuela en
un plan— porque en un producto cuyo argumento es la confianza, **lo que el agente mira importa tanto
como lo que envía**.

Lo que **no** cambia, y conviene decirlo primero: la frontera. Ninguna ruta cruza en claro por ningún
camino, nuevo o viejo. La identidad de proyecto sigue siendo un valor irreversible derivado con el
secreto local.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Todo el trabajo de un proyecto cuenta como un proyecto (Priority: P1)

Una desarrolladora trabaja en su proyecto a lo largo de la semana: unas veces lanza la herramienta
desde la raíz del proyecto, otras desde un subdirectorio (`frontend/`, `docs/`, un paquete concreto)
porque es donde tiene el foco ese día. Al final del mes, quien mira el gasto por proyecto ve **una
sola línea** con todo el consumo de ese proyecto, no una línea por cada sitio desde el que ella lanzó
la herramienta.

**Why this priority**: Es el problema medido y la razón de existir de la feature. Sin esto, la
dimensión de proyecto no sirve para presupuestar ni para repartir coste, que es exactamente lo que la
plataforma va a construir encima. Cualquier otra mejora es cosmética si esta no está.

**Independent Test**: Se prueba entera por sí sola procesando eventos originados en distintos puntos
de un mismo proyecto y comprobando que todos reciben la misma identidad, sin necesidad de las
historias 2 y 3.

**Acceptance Scenarios**:

1. **Given** dos eventos de un mismo proyecto, uno originado en la raíz del proyecto y otro en un
   subdirectorio cualquiera de ese proyecto, **When** el agente deriva su identidad de proyecto,
   **Then** ambos eventos reciben **la misma** identidad de proyecto.
2. **Given** un evento originado en un subdirectorio anidado a varios niveles de profundidad dentro
   de un proyecto, **When** el agente deriva su identidad, **Then** recibe la misma identidad que un
   evento originado en la raíz de ese proyecto.
3. **Given** dos eventos originados en **proyectos distintos**, **When** el agente deriva sus
   identidades, **Then** reciben identidades **distintas** (la agrupación no colapsa proyectos
   diferentes en uno).
4. **Given** un evento originado en un directorio que es **él mismo** la raíz de su proyecto,
   **When** el agente deriva su identidad, **Then** la identidad es la de ese proyecto, sin
   tratamiento especial ni diferencia respecto de un evento lanzado desde uno de sus subdirectorios.
5. **Given** un proyecto anidado dentro de otro proyecto, **When** el agente deriva la identidad de un
   evento originado dentro del proyecto interior, **Then** la identidad corresponde al proyecto
   **más cercano** que lo contiene, no al exterior.
6. **Given** un evento cuyo directorio de lanzamiento es un **enlace simbólico que apunta a un
   directorio dentro de un proyecto**, y el sistema permite resolverlo, **When** el agente deriva su
   identidad, **Then** recibe **la misma** identidad que un evento originado en ese directorio por su
   ruta real.

---

### User Story 2 - Trabajar fuera de un proyecto no multiplica identidades (Priority: P2)

La misma desarrolladora lanza la herramienta en un directorio suelto que no pertenece a ningún
proyecto reconocible: una carpeta de pruebas, un scratch, un directorio de descargas. A veces la ruta
que el log declara trae una barra final, a veces no; a veces llega a través de un enlace simbólico y
a veces por su ruta real. Todas esas veces son **el mismo sitio**, y el gasto aparece agrupado como
tal, no repartido entre variantes ortográficas de la misma ruta.

**Why this priority**: Recorta la segunda fuente de fragmentación —la puramente sintáctica— y cubre
el caso en que la primera regla no aplica. Aporta valor por sí misma, pero afecta a menos eventos que
la historia 1, que es donde está el grueso medido.

**Independent Test**: Se prueba por sí sola con eventos originados en directorios que no pertenecen a
ningún proyecto, comprobando que variantes sintácticas de la misma ruta convergen en una identidad y
que rutas genuinamente distintas no colisionan.

**Acceptance Scenarios**:

1. **Given** dos eventos cuyo directorio de lanzamiento es la **misma ruta** expresada de dos formas
   sintácticamente distintas (con y sin barra final, o con segmentos redundantes), y que **no**
   pertenecen a ningún proyecto reconocible, **When** el agente deriva su identidad, **Then** ambos
   reciben la misma identidad de proyecto.
2. **Given** dos eventos cuyo directorio de lanzamiento es un enlace simbólico y su destino real
   respectivamente, y el sistema permite resolver el enlace, **When** el agente deriva su identidad,
   **Then** ambos reciben la misma identidad de proyecto.
3. **Given** un evento cuyo directorio de lanzamiento **no puede observarse** (ya no existe, no hay
   permisos, o el enlace está roto), **When** el agente deriva su identidad, **Then** el agente emite
   el evento igualmente, con la identidad derivada de la forma sintácticamente normalizada de la ruta
   declarada, **y** el procesamiento del lote continúa sin interrupción.
4. **Given** dos eventos originados en **directorios sueltos distintos**, **When** el agente deriva
   sus identidades, **Then** reciben identidades distintas.

---

### User Story 3 - El contrato promete exactamente lo que el agente hace (Priority: P3)

Un desarrollador escéptico lee los documentos de contrato del agente antes de instalarlo, que es
justamente el gesto que el producto invita a hacer. Hoy encuentra que el contrato **admite un modo
opcional de enviar la identidad de proyecto en claro** — un modo que el agente nunca ha implementado
y que ningún camino del código consulta. Tras esta feature no encuentra tal cosa: el contrato dice
que la identidad de proyecto cruza siempre de forma irreversible, y eso es exactamente lo que ocurre.

**Why this priority**: No cambia ninguna cifra, y por eso va la última; pero cierra una brecha entre
lo prometido y lo hecho en el único documento cuya credibilidad sostiene el producto. Un contrato que
ofrece una puerta que no existe es una mentira pequeña en el sitio más caro.

**Independent Test**: Se prueba por sí sola revisando los documentos de contrato y ejercitando la
configuración con el valor retirado, sin depender de las historias 1 y 2.

**Acceptance Scenarios**:

1. **Given** los documentos de contrato de frontera del agente, **When** se buscan promesas de envío
   de la identidad de proyecto en claro, **Then** no queda ninguna: el contrato declara la identidad
   de proyecto como valor irreversible, sin modo alternativo.
2. **Given** una configuración local que **solicita el modo de envío en claro retirado**, **When** el
   agente arranca, **Then** **se detiene con un error visible** que nombra la causa y la corrección,
   y **no** procesa ni emite ningún evento.
3. **Given** una configuración local con un valor residual que solicitaba el comportamiento que ya es
   el único —la derivación irreversible—, **When** el agente arranca, **Then** funciona con
   normalidad, **sin error y sin aviso**: pedir lo que el agente hace siempre no es un problema que
   comunicar.
4. **Given** una configuración local que no menciona el ajuste retirado, **When** el agente la carga,
   **Then** funciona con normalidad y la identidad de proyecto cruza de forma irreversible.
5. **Given** cualquier configuración local, **When** el agente emite un evento, **Then** **NUNCA**
   existe combinación de ajustes que haga cruzar una ruta en claro.

---

### Edge Cases

- **El log no declara directorio de trabajo.** La identidad de proyecto queda **ausente**, como hoy.
  No se inventa un valor, no se sustituye por el directorio del propio fichero de log, no se
  reutiliza el de otro evento.
- **El directorio declarado ya no existe** cuando el agente procesa el log. Los logs se procesan en
  diferido y pueden describir sesiones de hace días: el proyecto pudo borrarse, moverse o renombrarse.
  Sin observación posible → forma sintácticamente normalizada, evento emitido igualmente.
- **El agente no tiene permisos** para observar el directorio o alguno de sus ancestros. Mismo trato:
  mejor esfuerzo, degradación silenciosa hacia lo sintáctico, evento emitido.
- **Enlace simbólico roto** o cadena de enlaces que no resuelve. Igual: se usa la forma normalizada de
  lo declarado.
- **El directorio de lanzamiento es un directorio temporal.** Cae en la regla del directorio
  normalizado (FR-005) **porque ningún proyecto lo contiene**, y produce una identidad estable como
  cualquier otro directorio suelto. Si dentro de un temporal hubiera un proyecto genuino —un clon de
  usar y tirar, por ejemplo—, ese anidamiento ya lo resuelve FR-004 sin trato especial.
- **El directorio de lanzamiento es el directorio personal del usuario.** Este caso **no** se resuelve
  por la razón anterior, y la diferencia importa: el directorio personal **sí puede ser la raíz de un
  árbol de trabajo** —es habitual versionar la configuración personal—, así que la regla general lo
  reconocería como proyecto y colapsaría bajo una sola identidad todo lo que cuelga de él que no
  pertenezca a un proyecto más cercano. Queda excluido **por la regla explícita FR-004a**, no por
  suponer que nadie versiona su directorio personal.
- **El directorio de lanzamiento es la raíz del sistema de ficheros.** Se excluye **por la misma
  regla y la misma razón que el directorio personal** (FR-004a), no porque «ningún proyecto la
  contenga»: la raíz **también puede ser raíz de un árbol de trabajo** —hay sistemas gestionados por
  repositorio—, y entonces la regla general colapsaría bajo **una sola identidad** absolutamente todo
  lo que no perteneciera a un proyecto más cercano. Es el mismo defecto que el del home, a la escala
  máxima.
- **Un directorio se convierte en proyecto después** de haber generado eventos (se inicializa el
  repositorio más tarde). Los eventos anteriores conservan la identidad que se les derivó y los
  posteriores reciben la de la raíz: la identidad de un mismo sitio **puede cambiar en el tiempo** si
  cambia el sistema de ficheros. Es consecuencia inevitable de observar el mundo real y se acepta.
- **Dos desarrolladores en el mismo proyecto** siguen produciendo identidades distintas: el secreto
  local es por instalación y no se comparte. Fuera de alcance (lo resuelve el mapeo de la
  plataforma).
- **El mismo proyecto en dos máquinas del mismo desarrollador** produce identidades distintas, por la
  misma razón. Fuera de alcance.
- **Árboles de trabajo paralelos del mismo repositorio** (cada uno con su propia raíz) producen
  identidades distintas: son árboles de trabajo distintos, y la regla es «la raíz del árbol de trabajo
  al que pertenece el directorio».
- **Diferencias de plataforma**: el agente se ejecuta en macOS, Linux y Windows, con separadores,
  unidades y semánticas de enlace distintas. La identidad de un mismo proyecto debe ser estable
  **dentro de una máquina**; entre máquinas ya se sabe que difiere.
- **Muchos eventos del mismo directorio.** Un fichero de log típico contiene cientos de eventos de un
  mismo directorio de trabajo; la resolución no puede convertir el procesamiento en algo perceptible
  para el usuario (ver SC-006).

---

## Requirements *(mandatory)*

### Functional Requirements

#### Identidad de proyecto: qué identifica

- **P-004 FR-001**: El agente **DEBE** derivar la identidad de proyecto de un evento del **proyecto al
  que pertenece** el directorio de lanzamiento declarado por el log, entendido como la **raíz del
  árbol de trabajo** que lo contiene — no del subdirectorio concreto desde el que se trabajó.
- **P-004 FR-002**: Dos eventos originados en **cualquier** punto de trabajo de un mismo proyecto
  **DEBEN** recibir **la misma** identidad de proyecto.
- **P-004 FR-003**: Dos eventos originados en proyectos distintos **NUNCA** deben recibir la misma
  identidad de proyecto por efecto de esta agrupación.
- **P-004 FR-004**: Cuando un directorio pertenece a más de un proyecto anidado, el agente **DEBE**
  atribuirlo al proyecto **más cercano** que lo contiene.
- **P-004 FR-004a**: Un árbol de trabajo cuya raíz es el **directorio personal del usuario o la raíz
  del sistema de ficheros** **NO** se reconoce como proyecto a efectos de esta agrupación. Un
  directorio bajo cualquiera de ellos que no pertenezca a un proyecto más cercano cae en la regla del
  directorio normalizado (FR-005); esos dos directorios, cuando son ellos mismos el de lanzamiento,
  también.
- **P-004 FR-005**: Cuando el directorio de lanzamiento **no pertenece a ningún proyecto
  reconocible**, el agente **DEBE** derivar la identidad del **propio directorio, normalizado**.
- **P-004 FR-006**: La normalización **DEBE** hacer converger las variaciones **puramente
  sintácticas** de una misma ruta —barras finales, segmentos redundantes— en una sola identidad.
- **P-004 FR-006a**: La identidad **DEBE** corresponder a la **ubicación real** del directorio de
  lanzamiento cuando el sistema permite observarla, con independencia de cómo la exprese la ruta
  declarada: la resolución de enlaces simbólicos (de mejor esfuerzo, FR-009) **precede** al
  reconocimiento de proyecto. Un enlace que conduce al interior de un proyecto recibe la identidad de
  ese proyecto, igual que la ruta real.
- **P-004 FR-007**: La normalización **NUNCA** debe hacer converger rutas que designan directorios
  genuinamente distintos.
- **P-004 FR-008**: Un evento cuyo log **no declara** directorio de trabajo **DEBE** seguir
  produciendo identidad de proyecto **ausente**, exactamente como hoy.

#### Mejor esfuerzo: la emisión no se interrumpe jamás

- **P-004 FR-009**: La resolución de la identidad es de **mejor esfuerzo**: si el sistema no permite
  observar (directorio inexistente, permisos denegados, enlace irresoluble, o cualquier otro fallo del
  sistema de ficheros), el agente **DEBE** usar la forma sintácticamente normalizada de la ruta
  declarada.
- **P-004 FR-010**: La emisión de un evento **NUNCA** se interrumpe, se pospone ni se descarta por no
  poder resolver o normalizar su identidad de proyecto. Un fallo de resolución **NUNCA** detiene el
  procesamiento del resto del lote.

#### Observación del sistema de ficheros: alcance declarado

- **P-004 FR-011**: El agente **PASA A OBSERVAR** el sistema de ficheros local para resolver la
  identidad de proyecto. Esta ampliación de comportamiento respecto del estado anterior —en el que el
  agente solo transcribía lo que el log declaraba— queda **declarada como parte del alcance de esta
  feature**.
- **P-004 FR-012**: La observación **DEBE** limitarse a lo necesario para determinar a qué proyecto
  pertenece un directorio y para normalizar una ruta. El agente **NUNCA** lee el contenido de los
  ficheros de trabajo del usuario (código, documentos, historial, configuración de sus herramientas)
  con este fin, y **NUNCA** modifica, crea ni borra nada en el sistema de ficheros observado: la
  observación es **de solo lectura**.

#### Retirada del modo en claro

- **P-004 FR-013**: Una configuración local que **solicita el modo de envío en claro retirado**
  **DEBE** detener el arranque del agente con un error visible que nombre la causa —el modo fue
  retirado— y la corrección —eliminar esa clave de la configuración—. El agente **NUNCA** procesa ni
  emite eventos con esa configuración presente.
- **P-004 FR-013a**: Un valor residual que solicitaba el comportamiento que **ya es el único e
  incondicional** —la derivación irreversible— **NUNCA** debe detener el arranque ni producir aviso:
  se ignora en silencio. Pedir lo que el agente hace siempre no es un error, y advertirlo sería ruido.
- **P-004 FR-013b**: Cualquier **otro** valor en esa clave obsoleta se ignora igualmente en silencio.
  La parada de FR-013 se dispara **solo** por la solicitud explícita del modo en claro retirado — es
  la única que pide algo que el producto ya no promete.
- **P-004 FR-014**: Los documentos del agente **DEBEN** dejar de prometer el modo de envío en claro de
  la identidad de proyecto, y **DEBEN** declarar que esa identidad cruza **siempre** de forma
  irreversible, sin modo alternativo. El alcance documental **se determina por barrido mecánico de
  toda la documentación del repositorio** —especificaciones, contratos, README y cualquier documento
  orientado a usuario—, **no** por la lista de ficheros que alguien recuerde. Los dos ficheros hoy
  conocidos —`specs/001-agente-inicial/contracts/boundary-event.md` y
  `specs/001-agente-inicial/data-model.md`— son el **mínimo que el barrido debe reencontrar**: si no
  aparecen, el barrido está mal hecho.
- **P-004 FR-015**: El ajuste de configuración asociado al modo retirado **DEBE** desaparecer de la
  superficie de configuración **documentada y con significado** del agente: ningún documento lo ofrece
  como opción y ningún valor suyo altera el comportamiento. El trato de los valores residuales que aún
  lo mencionen es exclusivamente el definido en **FR-013 / FR-013a / FR-013b**.

#### Frontera de datos: lo que no cambia

- **P-004 FR-016**: La identidad de proyecto **DEBE** seguir cruzando la frontera **únicamente** como
  valor irreversible derivado con el secreto local. El secreto **NUNCA** se transmite (Constitución,
  Principio I).
- **P-004 FR-017**: **NINGUNA** ruta cruza la frontera en claro por **ningún** camino: ni la ruta
  original declarada por el log, ni su forma normalizada, ni la raíz del proyecto resuelta, ni
  fragmento alguno de cualquiera de ellas. Esto aplica a todo camino de salida **hacia el exterior**:
  el evento serializado, la cola de envío y el transporte. Las salidas diagnósticas locales del agente
  están fuera del alcance de este requisito (su saneado es trabajo aparte, ya identificado).
- **P-004 FR-018**: La identidad de proyecto resultante **DEBE** conservar la **misma forma** que hoy
  —un valor irreversible de longitud fija, o ausente—, de modo que la plataforma receptora no
  distinga por su forma si procede de una raíz de proyecto o de un directorio normalizado. Lo que
  cambia es **qué** se deriva, no **cómo se ve** lo derivado.
- **P-004 FR-019**: Las demás identidades derivadas —**sesión** y **máquina**— **NUNCA** cambian: ni
  su fuente, ni su ámbito, ni su forma. El secreto local y su ciclo de vida tampoco.
- **P-004 FR-020**: El resto del contrato de frontera —el conjunto cerrado de campos del evento, sus
  nombres y sus tipos— **NUNCA** cambia por efecto de esta feature. No se añade ningún campo.

### Key Entities

- **Identidad de proyecto**: el valor irreversible que viaja en cada evento y que la plataforma usa
  como clave de agrupación de la dimensión «por proyecto». Derivada del proyecto (o del directorio
  normalizado) con el secreto local. Puede estar **ausente**.
- **Directorio de lanzamiento**: el directorio de trabajo que el log de la herramienta declara para
  una sesión. Es la **entrada** de la derivación, y hoy es también su salida directa. Puede no estar
  declarado.
- **Proyecto**: el árbol de trabajo al que pertenece un directorio de lanzamiento, identificado por su
  **raíz**. Un directorio pertenece a lo sumo a un proyecto «más cercano»; puede no pertenecer a
  ninguno.
- **Secreto local**: el valor que hace irreversible la derivación. Reside en la máquina del
  desarrollador, **NUNCA** se transmite, y esta feature **no lo toca** — ni su generación, ni su
  ámbito, ni su ciclo de vida.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Reprocesando el conjunto de logs históricos de referencia —el que produjo 12 identidades
  para un solo desarrollador en una sola máquina—, las identidades resultantes son **exactamente las
  que la regla de esta spec predice** sobre la **enumeración de verdad-terreno** de los directorios de
  origen, aportada por Basilio antes de la validación. La predicción se deriva de la regla —raíz de
  proyecto; clones y árboles paralelos conservan identidad propia; directorios sueltos, normalizados—,
  **no** de una cifra prometida de antemano.
- **SC-002**: Para cualquier conjunto de eventos originados en un mismo proyecto **cuyo directorio es
  observable en el momento del procesamiento**, el número de identidades de proyecto distintas
  producidas es **exactamente 1**, con independencia del punto de lanzamiento. Los eventos degradados
  por **FR-009** pueden recibir una identidad distinta: es la divergencia aceptada por diseño.
- **SC-003**: El **100 %** de los eventos que hoy se emiten se siguen emitiendo tras el cambio: cero
  eventos perdidos, pospuestos o descartados por causa de la resolución de identidad, incluidos los
  casos en que el directorio no puede observarse.
- **SC-004**: Las identidades de **sesión** y de **máquina** producidas para un mismo conjunto de
  logs son **idénticas** antes y después del cambio (regresión cero en las dimensiones que no cambian).
- **SC-005**: En el evento serializado, la cola de envío y lo transmitido, el número de apariciones de
  una ruta en claro —original, normalizada o de raíz— es **cero**, verificado con una entrada que
  inyecta rutas reconocibles.
- **SC-006**: Una pasada completa sobre el conjunto de logs de referencia **no aumenta su duración de
  forma perceptible** para el usuario respecto de la línea base previa al cambio (referencia: no más
  de un **10 %** de incremento).
- **SC-007**: Una configuración local que solicita el modo de envío en claro retirado detiene el
  arranque en el **100 %** de los casos, con **cero** eventos emitidos bajo ella; y un valor residual
  que pedía el comportamiento ya único arranca con normalidad en el **100 %** de los casos, sin error
  ni aviso (cero falsos positivos).
- **SC-008**: Un **barrido mecánico —con su comando— de toda la documentación del repositorio**,
  incluido el **README público**, encuentra **cero** menciones de un modo de envío en claro de la
  identidad de proyecto. Es un barrido, no una lectura: lo que no se busca de forma reproducible no
  se puede declarar ausente.

---

## Decisiones tomadas durante la especificación

- **D-004-1 (Basilio, 2026-08-09) — trato de la configuración retirada: parada, pero solo cuando pide
  el modo en claro.** El encargo fijaba que ese valor «no debe ser aceptado en silencio»; entre
  detener el arranque y avisar-y-continuar, se elige **detener**, y se acota a la solicitud
  explícita del modo retirado.

  El criterio que separa los dos casos es **qué pedía el usuario**, no qué clave escribió: quien pidió
  el modo en claro pidió algo que el producto **ya no promete**, y seguir adelante en silencio —o con
  un aviso que puede perderse— le dejaría creyendo que tiene una puerta que no existe; quien tenía el
  valor que pedía la derivación irreversible pidió **exactamente lo que el agente hace siempre**, y
  pararle o avisarle sería ruido sobre una petición ya satisfecha.

  Recogido en **FR-013** (parada), **FR-013a** (silencio para el valor ya satisfecho), **FR-013b**
  (silencio para cualquier otro valor) y **SC-007**.

- **D-004-2 (2026-08-09) — la resolución de enlaces precede al reconocimiento de proyecto.** La
  primera redacción metía los enlaces simbólicos dentro de la normalización del **fallback**, con lo
  que un enlace que apunta al interior de un proyecto no habría llegado nunca a reconocerse como tal.

  El criterio es que **la identidad debe corresponder a dónde se trabajó de verdad**, no a cómo se
  escribió la ruta: si se resuelve primero la ubicación real y después se busca el proyecto, un enlace
  y su destino convergen en las dos ramas —la de proyecto y la de directorio suelto— en vez de solo en
  la segunda. Recogido en **FR-006** (solo sintaxis), **FR-006a** (ubicación real, precedencia
  explícita) y el escenario 6 de US1.

- **D-004-3 (2026-08-09) — ni el directorio personal del usuario ni la raíz del sistema de ficheros
  son proyectos.** Se añade una exclusión explícita en vez de confiar en que la regla general lo
  resuelva.

  El criterio es que el directorio personal **sí puede ser raíz de un árbol de trabajo** —versionar la
  configuración personal es práctica corriente—, así que la regla general lo reconocería como proyecto
  y colapsaría bajo **una sola identidad** todo lo que cuelga de él sin proyecto propio: exactamente
  la fragmentación que esta feature corrige, pero al revés.

  **El mismo criterio alcanza a la raíz del sistema de ficheros** —hay sistemas gestionados por
  repositorio—, y allí el defecto es el mismo **a la escala máxima**: absolutamente todo lo que no
  perteneciera a un proyecto más cercano caería en un único bucket. Por eso la exclusión se enuncia
  sobre los dos y no sobre uno, y por eso **no** se enuncia sobre «directorios que ningún proyecto
  contiene», que es la razón —distinta— por la que caen los temporales.

  Recogido en **FR-004a** (los dos sujetos) y en los edge cases, que dicen **por qué** estos dos casos
  se resuelven por regla explícita y los temporales no.

- **D-004-4 (2026-08-09) — la frontera de esta feature se acota a la salida hacia el exterior.**
  FR-017 pretendía cubrir «todo camino de salida, incluida cualquier salida diagnóstica», y eso
  extendía el requisito a un territorio que esta feature no aborda ni verifica.

  El criterio es que **un requisito que no se va a comprobar no protege: solo aparenta**. La frontera
  que esta feature garantiza y prueba es la de salida al exterior —evento serializado, cola de envío y
  transporte—. El saneado de las **salidas diagnósticas locales** (stderr y equivalentes) queda
  **identificado como candidata aparte**, no silenciado. Recogido en **FR-017** y **SC-005**.

---

## Assumptions

- **SC-001 depende de un insumo que hoy no existe: la enumeración de verdad-terreno.** Validar SC-001
  exige saber **de qué directorios provienen realmente** los logs históricos de referencia —cuáles
  eran raíces de proyecto, cuáles clones o árboles paralelos, cuáles directorios sueltos—, y eso lo
  aporta Basilio **antes** de la validación. Sin esa enumeración, SC-001 no es verificable: no se
  puede comprobar una predicción contra un terreno que nadie ha declarado.
- **Continuidad de identidades históricas: se rompe, y se acepta.** Las identidades emitidas antes de
  esta feature no corresponden con las posteriores para el mismo proyecto. Es aceptable **porque el
  producto no está en producción**, y es precisamente la razón de hacer este cambio ahora. La
  migración o correspondencia de identidades históricas está **fuera de alcance**.
- **«Proyecto» se decide observando el sistema de ficheros, no preguntando al desarrollador.** No hay
  declaración manual de nombres ni de límites de proyecto (fuera de alcance). Si la observación no
  reconoce un proyecto, se cae en la regla del directorio normalizado — nunca se pide intervención.
- **El mecanismo de reconocimiento pertenece al plan.** Qué marca la raíz de un árbol de trabajo, qué
  normalización exacta se aplica y qué se observa para decidirlo son decisiones de `research.md` y
  `plan.md`, guiadas por la decisión previa de Basilio (D-C1/D-C2/D-C3, 2026-08-09): raíz de
  repositorio con fallback a directorio normalizado, enlaces simbólicos de mejor esfuerzo, retirada
  del modo en claro. Esta especificación fija el **qué** y deja el **cómo** fuera a propósito.
- **La identidad sigue siendo opaca para la plataforma.** El receptor no puede distinguir, por la
  forma del valor, si procede de una raíz de proyecto o de un directorio suelto; tampoco puede
  revertirla. La plataforma no necesita esa distinción para su dimensión «por proyecto».
- **La plataforma no se toca.** La entidad de proyecto y el mapeo identidad→proyecto son una feature
  aparte, ya decidida, en el otro repositorio. Esta feature no depende de que aquella exista, y
  aquella se beneficia de que esta se haga antes.
- **El enrolamiento y su formato no se tocan** (fuera de alcance).
- **Multiplataforma**: la estabilidad de la identidad se exige **dentro de una máquina**. Que dos
  máquinas —o dos sistemas operativos— produzcan identidades distintas para el mismo proyecto es
  esperado y está fuera de alcance por diseño (lo resuelve el mapeo de la plataforma).
- **Una sola herramienta de origen hoy.** El agente solo ingiere Claude Code; la regla se define sobre
  el concepto de «directorio de lanzamiento declarado por el log», de modo que una herramienta futura
  que declare lo mismo herede la regla sin rediseñarla.

---

## Fuera de alcance

- Cualquier cambio en la plataforma (entidad de proyecto y mapeo: feature aparte, otro repositorio).
- Compartir identidad de proyecto entre desarrolladores o entre máquinas.
- Declaración manual de nombres o límites de proyecto por el desarrollador.
- Cualquier cambio en el enrolamiento o su formato.
- Migración o correspondencia de identidades históricas.
- Cualquier cambio en las identidades de sesión y máquina, en el secreto local, o en el resto del
  contrato de frontera.
