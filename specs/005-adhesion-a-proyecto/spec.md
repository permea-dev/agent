# Feature Specification: Adhesión a proyecto — que quien instala se agrupe solo, presentando un código

**Feature Branch**: `005-adhesion-a-proyecto`

**Created**: 2026-08-18

**Status**: Draft

**Input**: Descripción de Basilio, 2026-08-18, tras el descubrimiento de la plataforma: la entidad
Proyecto y su código de adhesión existen y están demostrados sobre base real; falta el lado del agente.

---

## Contexto: agrupar exige hoy que alguien lea cadenas hexadecimales

La plataforma ya sabe agrupar. Tiene la entidad **Proyecto**, sabe reunir bajo ella varias identidades
de instalación, y resuelve esa agrupación **en lectura**: en cuanto la unión existe, todo el consumo
histórico de esa instalación cuenta bajo su Proyecto, sin reprocesar ni migrar nada.

Lo que no tiene es **cómo se declara esa unión desde el lado de quien instala**. Hoy la única vía es
que alguien con privilegios entre en el panel, mire la lista de identidades sin agrupar —cadenas
hexadecimales que no significan absolutamente nada para un humano— y decida a mano cuál pertenece a
qué. Eso tiene tres consecuencias, y las tres son de producto:

1. **No escala con el equipo.** Cada persona que instala el agente genera una identidad nueva que
   alguien tiene que ir a reclamar. El coste de dar de alta a alguien crece con el tamaño del equipo,
   y recae en quien menos contexto tiene sobre el trabajo de los demás.
2. **Se adivina.** Quien mapea a mano no ve rutas —la frontera de datos impide que crucen—, así que
   empareja identidad con proyecto por el volumen, por la fecha o preguntando. Es una decisión sin
   dato, y una decisión sin dato se equivoca.
3. **El alta no es onboarding.** Instalar el agente deja el trabajo a medias: mide, pero no cuenta
   donde debe hasta que un tercero interviene.

Esta feature cierra esa brecha por el único lado que puede cerrarla sin pedirle nada a nadie: **quien
instala presenta un código de adhesión que le han dado, y queda agrupado**. El comando le responde a
qué Proyecto se ha unido, que es lo que convierte la operación en un desenlace y no en un silencio.

**Y no hay migración que hacer, ni la habrá.** Como la agrupación se resuelve del lado de la
plataforma en lectura, el efecto es **retroactivo por construcción**: en el instante en que la unión
existe, el histórico entero de esa instalación ya cuenta bajo su Proyecto. Esta feature **no mueve, no
reescribe y no reenvía ni un solo evento**, y esa ausencia de trabajo no es una simplificación: es la
consecuencia directa de dónde vive la agrupación. Se escribe aquí porque, sin decirlo, la ausencia de
cualquier operación de migración se lee como un olvido.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien instala se agrupa solo y sabe dónde ha quedado (Priority: P1)

Alguien acaba de enrolar su agente. Le han pasado un código de adhesión —por el canal que sea: un
mensaje, la documentación interna del equipo, el propio panel—. Se sitúa dentro del proyecto que
quiere agrupar y presenta el código. El comando le responde **con la denominación del Proyecto al que
ha quedado unido**.

**Why this priority**: es la feature. Sin esto no hay nada: es lo que sustituye el mapeo manual por
autoservicio, y lo que convierte el alta del agente en onboarding completo. Las demás historias
protegen a ésta de sus modos de fallo.

**Independent Test**: se ejercita entero presentando un código utilizable desde dentro de un proyecto
con una instalación aún sin agrupar, y comprobando dos cosas **que NO se verifican de la misma
manera**: **(1)** que el desenlace nombra el Proyecto — automatizable en este repositorio; y **(2)** que
a partir de ese instante el consumo previo de esa instalación aparece contabilizado bajo él **sin que
se haya emitido ningún evento nuevo** — que es **SC-006**, y por tanto **validación manual contra una
plataforma real, NO automatizable aquí**. Automatizar (2) exigiría fabricar la agrupación, y entonces
estaría comprobando su propio simulacro: **es justo lo que SC-006 prohíbe**.

**Acceptance Scenarios**:

1. **Given** una instalación enrolada, situada dentro de un árbol de trabajo con raíz de proyecto
   reconocible, cuya identidad todavía no pertenece a ningún Proyecto, **When** se presenta un código
   de adhesión utilizable, **Then** el desenlace es de éxito e **incluye la denominación del Proyecto**
   al que se ha unido.
2. **Given** la unión recién hecha, **When** se consulta el consumo de esa organización, **Then** el
   consumo **anterior** a la unión de esa misma instalación **ya cuenta bajo ese Proyecto**, sin que se
   haya emitido, reenviado ni reprocesado ningún evento. *(Este escenario **es SC-006**: validación
   contra plataforma real, **no automatizable en este repositorio**.)*
3. **Given** la unión recién hecha, **When** se inspeccionan **los artefactos locales de la
   instalación** —el conjunto que enumera SC-007, no solo la configuración—, **Then** **no han
   cambiado**: la unión no dejó rastro local.
4. **Given** las mismas condiciones de partida, **When** se presenta el código **sin pasarlo como
   argumento**, por la vía de la entrada estándar, **Then** la operación se completa igualmente —
   **existe una vía que no exige que el código aparezca en la línea de órdenes**— y el desenlace es
   **indistinguible** del obtenido por la otra vía, **en el sentido que define SC-011 (A)**: qué se
   compara, con qué piezas, y cómo se demuestra que la comparación sabe fallar. El éxito sale por el
   canal que exige **SC-011 (B)**.

---

### User Story 2 - Repetir la operación no revela nada, y no rompe nada (Priority: P1)

La misma persona, la misma instalación, el mismo código, otra vez —porque no vio la salida, porque
dudó, porque lo pegó dos veces—. El desenlace que ve es **el mismo que la primera vez**.

**Why this priority**: es P1 y no P2 porque **es una propiedad de seguridad, no una comodidad**. La
plataforma declara estos dos desenlaces indistinguibles a propósito: si se distinguieran, quien pega
un código dos veces averiguaría cuál de las dos veces surtió efecto, y con ello si esa identidad
estaba ya unida. El comando **no puede** introducir esa distinción por su cuenta, ni siquiera con
buena intención tipográfica.

**Independent Test**: presentar dos veces seguidas el mismo código desde la misma instalación y
comparar los dos desenlaces completos —texto, canal y resultado del proceso—. Deben ser iguales, **con
las tres exigencias de SC-002**: que cada salida sea del tipo que le toca, que sean idénticas entre sí,
y que la comparación **sepa fallar**. Comprobar solo la igualdad lo satisfarían dos salidas vacías.

**Acceptance Scenarios**:

1. **Given** una instalación **ya unida** a un Proyecto mediante un código, **When** se presenta **el
   mismo código otra vez**, **Then** el desenlace es de éxito y **nombra el mismo Proyecto**.
2. **Given** los desenlaces de la primera y la segunda presentación, **When** se comparan
   íntegramente, **Then** son **indistinguibles**: mismo texto, mismo canal, mismo resultado del
   proceso. Nada permite deducir cuál de las dos produjo el cambio.
3. **Given** la segunda presentación, **When** se examina el estado del lado de la plataforma,
   **Then** la instalación sigue unida **una sola vez** al mismo Proyecto. *(Observa el estado remoto,
   así que es **validación contra plataforma real** y **no automatizable aquí**, por el mismo motivo que
   SC-006.)*

---

### User Story 3 - El comando rehúsa antes de hablar cuando no puede acertar (Priority: P1)

Alguien lanza el comando **fuera de un árbol de proyecto** —desde su directorio personal, desde una
carpeta suelta, desde cualquier sitio que no tenga raíz reconocible—. El comando **rehúsa sin emitir
ninguna petición**, y le dice que tiene que ejecutarlo dentro del árbol que quiere agrupar.

**Why this priority**: es P1 porque **sin esto la feature es peor que no tenerla**. Fuera de un árbol
de proyecto la identidad que el comando presentaría es la del directorio suelto desde el que se lanzó.
Esa identidad **sí puede emitir** —y emite, si alguien trabajó ahí—, pero **con casi total seguridad no
es la del proyecto que se quería agrupar**. La plataforma la aceptaría sin objeción y respondería con
un éxito perfectamente formado. La persona vería «te has unido a X», y el consumo que esperaba ver
agrupado **seguiría apareciendo suelto**. Ese fallo es **indistinguible de una avería del servidor** y
consume el tiempo de dos personas antes de que alguien sospeche del sitio desde el que se lanzó el
comando. Rehusar en local, antes de la petición, es lo único que lo convierte en un error legible.

**Independent Test**: lanzar el comando desde un directorio sin raíz de proyecto por encima y
comprobar **que no se emite ninguna petición** —no que la petición falle: que no exista—, y que el
mensaje dice qué hacer. **La ausencia se verifica con el observador que exige SC-004**, que incluye
demostrar que el mismo instrumento **registra el caso positivo**: sin esa segunda mitad, «no se emitió»
no se distingue de «no se miró». Y el rehúse sale por el canal que exige **SC-011 (B)**, con los dos
capturados por separado.

**Acceptance Scenarios**:

1. **Given** un directorio que no tiene raíz de proyecto reconocible ni por encima de él, **When** se
   lanza el comando con un código perfectamente válido, **Then** el comando **rehúsa**, **no emite
   ninguna petición**, y el mensaje indica que debe ejecutarse dentro del árbol de trabajo que se
   quiere agrupar.
2. **Given** ese mismo rehúse, **When** se examinan **los artefactos locales que enumera SC-007** —con
   su observador— y el estado de la plataforma, **Then** **nada ha cambiado en ninguno de los dos**.
   Aquí **sí** puede afirmarse de los dos: al no emitirse ninguna petición, el estado remoto no pudo
   cambiar.

---

### User Story 4 - Los desenlaces que no son éxito se entienden sin filtrar información (Priority: P2)

El código no vale, o esta instalación ya pertenece a otro Proyecto. En los dos casos la persona
recibe un mensaje que le permite actuar —pedir otro código, hablar con quien administra— **sin que el
comando revele nada que la plataforma calle**.

**Why this priority**: P2 porque la feature entrega valor sin ella —los caminos de éxito son P1—, pero
sin ella la primera incidencia real se convierte en una fuga: un comando que explique *por qué* un
código no vale es un oráculo para averiguar qué códigos existen, y uno que nombre el Proyecto ajeno
revela la estructura interna de la organización a quien solo tenía un código.

**Independent Test**: presentar (a) un código bien formado pero inexistente y (b) un código utilizable
desde una instalación ya unida a **otro** Proyecto; comprobar que ninguno de los dos mensajes contiene
ni la causa concreta ni la denominación ajena. La comparación de los dos rechazos va **con las tres
exigencias de SC-003**, y el contenido de los mensajes **con SC-005**: tampoco puede aparecer en ellos
el código presentado. Los dos son desenlaces de error, así que salen por el canal que exige
**SC-011 (B)**.

**Acceptance Scenarios**:

1. **Given** una instalación cuya identidad ya pertenece a **otro** Proyecto, **When** se presenta un
   código utilizable, **Then** se le informa de que esta instalación ya pertenece a otro Proyecto,
   **sin nombrarlo**.
2. **Given** un código **no utilizable**, **When** se presenta, **Then** se le informa de que el
   código no vale, **sin decir por qué**.
3. **Given** dos códigos no utilizables **por causas distintas**, **When** se presentan uno tras otro,
   **Then** los dos mensajes son **indistinguibles entre sí**.

---

### User Story 5 - Los estados en los que no se puede operar se declaran, no se intentan (Priority: P2)

La instalación no está enrolada; o el servidor no está alcanzable; o la configuración local tiene una
forma que no permite determinar con confianza a dónde dirigirse. En los tres casos el comando **lo
dice y no deja nada a medias en local**.

**Why this priority**: P2 porque son estados excepcionales, pero cada uno tiene una forma de fallar mal
que cuesta cara: intentar sin enrolamiento produce un error de credenciales que sugiere que el código
es el problema; afirmar un desenlace que no se pudo confirmar hace que alguien dé por hecha una unión
que no existe; y adivinar el destino ante una configuración rara produce un fallo que **parece culpa
de quien lanzó el comando**.

**Independent Test**: provocar los tres estados por separado y comprobar en cada uno que el mensaje
nombra la situación real, que no se afirma ningún desenlace, y que **el estado local no queda a
medias**. **Sobre el estado de la plataforma no se comprueba nada**: en el caso del servidor
inalcanzable queda **indeterminado por definición** (FR-013a), y exigir aquí que «no quede a medias»
sería pedir la verificación de algo que el requisito declara inverificable. Los tres son rehúses o
errores, así que salen por el canal que exige **SC-011 (B)**.

**Acceptance Scenarios**:

1. **Given** una instalación **sin enrolamiento**, **When** se lanza el comando, **Then** rehúsa e
   indica **cómo enrolarse**.
2. **Given** un servidor **inalcanzable, o que responde de forma que no permite establecer el
   desenlace**, **When** se presenta un código, **Then** se informa de que **no se pudo completar la
   operación**, **sin afirmar ningún desenlace** —ni que se unió, ni que el código no valía— y **sin
   dejar nada a medias en local**. El estado del lado de la plataforma queda **indeterminado**, y lo
   que hace tolerable esa incertidumbre —que **repetir la operación no tiene consecuencia**— lo
   establece **FR-013a**. Y la petición fallida **no queda encolada**, con el observador que exige
   **SC-010**: es justo la circunstancia que ese criterio nombra.
3. **Given** una configuración local con **forma inesperada**, de la que no puede determinarse con
   confianza a dónde dirigirse, **When** se lanza el comando, **Then** rehúsa **nombrando lo que
   encontró**, en lugar de intentarlo y fallar de una forma que atribuya el problema a quien lanzó el
   comando.

---

### Edge Cases

- **¿Qué pasa si el árbol de trabajo se convierte en proyecto entre dos ejecuciones?** La identidad que
  el comando presenta cambia, exactamente igual que cambia la que se estampa en los eventos. Es
  correcto: las dos derivan de lo mismo, y esa coherencia es justamente FR-002.
- **¿Y si el comando se lanza desde un subdirectorio profundo del proyecto?** La identidad presentada
  es la misma que desde la raíz: la agrupación es por árbol de trabajo, no por punto de lanzamiento.
- **¿Y si la instalación pierde el secreto local que particulariza sus identidades?** Sus identidades
  cambian por completo, y la unión anterior queda referida a una identidad que ya no emite. Esta
  feature **no lo detecta ni lo repara**: es una condición preexistente del agente, ajena a este
  alcance, y se anota como supuesto.
- **¿Y si dos personas del mismo equipo presentan el mismo código?** Cada una une **su** instalación.
  El código no es de un solo uso y no se agota: es la propiedad que lo hace útil para un equipo.
- **¿Y si el Proyecto se renombra después de la unión?** La unión no se ve afectada: lo que se une es
  la identidad al Proyecto, no a su denominación. Una ejecución posterior nombrará la denominación
  vigente en ese momento.
- **¿Y si no se presenta ningún código?** Es un error de uso: el comando lo indica sin emitir petición
  alguna, igual que el rehúse de FR-006.
- **¿Y si el árbol de trabajo existe pero el sistema no permite observarlo?** La derivación de
  identidad **sigue sin fallar** —esa garantía no se toca (FR-014)—, pero el comando **no puede
  afirmar** que hay raíz de proyecto, así que se comporta como en el caso «fuera de un árbol»: rehúsa.
  Rehusar de más cuesta una ejecución; unir de más cuesta un diagnóstico falso.
- **¿Y un proyecto que sencillamente NO tiene raíz reconocible? No podrá unirse por código. Nunca.**
  **El hecho, sin adornos**: un directorio de trabajo sin raíz de proyecto **emite eventos con toda
  normalidad y con identidad propia** —004 lo exige: *«Cuando el directorio de lanzamiento **no
  pertenece a ningún proyecto reconocible**, el agente **DEBE** derivar la identidad del **propio
  directorio, normalizado**»*, `specs/004-identidad-de-proyecto/spec.md:229-230` (P-004 FR-005)—. Pero
  FR-006 rehúsa exactamente en ese caso. **Así que esa identidad consume, aparece en la vista sin
  agrupar, y no hay forma de adherirla con un código.**

  **Por qué se acepta**: el comando no puede distinguir «trabajo real en un directorio suelto» de
  «alguien lanzó esto desde donde no debía» — **las dos situaciones son idénticas desde aquí**. Unir a
  ciegas cuesta un **diagnóstico falso que nadie sabe diagnosticar**: un éxito perfectamente formado
  sobre una identidad que quizá no es la que emite, cuyo síntoma es indistinguible de una avería de
  servidor. Rehusar de más cuesta **una ejecución** y un mensaje que dice qué hacer. La asimetría no
  está reñida.

  **Cuál es la salida para quien está en ese caso**, y son dos: **convertir el directorio en un árbol
  de trabajo reconocible** —tras lo cual la adhesión por código funciona con normalidad, y de forma
  retroactiva sobre lo ya emitido bajo la identidad nueva—, o **pedir el mapeo manual desde el panel**,
  que sigue existiendo y no lo sustituye esta feature. Lo que **no** hay es una tercera vía por la que
  el comando lo resuelva solo, y eso es deliberado.

---

## Requirements *(mandatory)*

### Functional Requirements

#### La capacidad, y su desenlace

- **P-005 FR-001**: Quien haya enrolado una instalación **DEBE** poder unirla a un Proyecto de su
  organización **presentando un código de adhesión**, sin necesitar privilegios de administración y sin
  intervención de terceros.
- **P-005 FR-002**: Cuando la unión queda establecida, el desenlace **DEBE** comunicar **la
  denominación vigente del Proyecto** al que la instalación ha quedado unida. Un éxito mudo no es
  desenlace: la persona necesita saber a qué se ha unido para poder detectar que se ha unido a lo que
  no era.
- **P-005 FR-003**: La feature **NUNCA** emite, reenvía, reescribe ni reprocesa eventos de consumo.
  El alcance retroactivo de la unión lo aporta la plataforma al resolver la agrupación en lectura, y
  **el comando no tiene que hacer nada para conseguirlo**.

#### La identidad presentada: la restricción que sostiene la feature

- **P-005 FR-004**: La identidad de proyecto que el comando presenta **DEBE** ser **exactamente la
  misma** que esta instalación estampa en los eventos que emite **para ese mismo árbol de trabajo**,
  bajo las mismas condiciones observables.
- **P-005 FR-005**: La coincidencia de FR-004 **DEBE** sostenerse **por construcción y no por
  coincidencia**: las dos identidades **DEBEN** proceder de la misma derivación, no de dos
  derivaciones que hoy den lo mismo.

  > **Por qué es la restricción que sostiene la feature.** Si difirieran, se uniría una identidad que
  > **no es la que emite**. La plataforma respondería con un éxito perfectamente formado, la persona
  > lo daría por bueno, y el consumo seguiría apareciendo sin agrupar. El síntoma sería «me he unido y
  > no aparece nada» — **indistinguible de una avería del servidor**, y por tanto diagnosticable solo
  > por quien ya sospeche de esto. Es el modo de fallo más caro de esta feature y el único que no
  > avisa.

#### Rehúse local: lo que se decide antes de hablar con nadie

- **P-005 FR-006**: Si no puede establecerse que el directorio desde el que se lanza el comando
  pertenece a un árbol de trabajo con raíz de proyecto reconocible, el comando **DEBE** rehusar y
  **NUNCA DEBE** emitir petición alguna. **Este rehúse tiene una consecuencia permanente para los
  directorios que nunca tendrán raíz reconocible —emiten con identidad propia y no podrán adherirse por
  código—: está documentada, con sus dos salidas, en §Edge Cases.**
- **P-005 FR-007**: El mensaje de ese rehúse **DEBE** indicar que el comando ha de ejecutarse **dentro
  del árbol de trabajo que se quiere agrupar**.
- **P-005 FR-008**: Si la instalación **no está enrolada**, el comando **DEBE** rehusar antes de emitir
  petición alguna, e **indicar cómo enrolarse**.
- **P-005 FR-009**: Si la configuración local tiene una forma que **no permite determinar con
  confianza** a dónde dirigirse, el comando **DEBE** rehusar **nombrando lo que encontró**, y **NUNCA
  DEBE** intentar la operación con un destino conjeturado. **FR-020 manda sobre este requisito**: lo
  que se nombra es **la FORMA de lo hallado** —qué falta, qué sobra, qué no se entiende— y **NUNCA**
  material sensible, aunque estuviera justo en lo hallado. Ante la duda entre ser informativo y no
  reproducir un secreto, **gana no reproducirlo**.

#### Los desenlaces, y lo que no revelan

- **P-005 FR-010**: Los desenlaces que la plataforma declara **indistinguibles** —haberse unido ahora y
  estar ya unido a ese mismo Proyecto— **DEBEN** presentarse **indistinguibles** **bajo las mismas
  condiciones observables**: mismo texto, mismo canal y mismo resultado del proceso. El comando
  **NUNCA** introduce una diferencia observable entre ellos. **La exigencia es sobre los dos desenlaces,
  no a través del tiempo**: si entre dos presentaciones cambia algo que el desenlace refleja —la
  denominación del Proyecto, por ejemplo—, las dos salidas diferirán, y eso **no** incumple este
  requisito.
- **P-005 FR-011**: Cuando la identidad pertenece **a otro** Proyecto, el desenlace **DEBE** decirlo y
  **NUNCA DEBE** nombrar ese Proyecto. El comando **NUNCA** infiere, deduce ni conjetura lo que la
  plataforma calla.
- **P-005 FR-012**: Cuando el código **no es utilizable**, el desenlace **DEBE** decirlo y **NUNCA
  DEBE** indicar la causa. Distintas causas **DEBEN** producir mensajes **indistinguibles entre sí**.
- **P-005 FR-013**: Cuando el desenlace **no puede establecerse** —el servidor no responde, o responde
  de forma que no permite determinarlo—, el comando **DEBE** informar de que la operación no pudo
  completarse, **NUNCA DEBE** afirmar ningún desenlace concreto, y **NUNCA DEBE** dejar nada a medias
  **en local**.
- **P-005 FR-013a**: En ese mismo caso, **el estado del lado de la plataforma queda INDETERMINADO para
  el comando, y la especificación NO afirma lo contrario**: si la petición llegó y se perdió la
  respuesta, la unión ocurrió. Lo que **DEBE** garantizarse es que esa incertidumbre sea **inocua**:
  **repetir la operación no tiene ninguna consecuencia adversa**, porque el código no se agota al
  usarse y unirse dos veces es indistinguible de unirse una (FR-010). El mensaje **DEBE** dejar a la
  persona en condiciones de volver a intentarlo sin miedo a duplicar nada.

  > **Por qué se escribe así, y por qué la redacción anterior era falsa.** Decir «ninguna operación
  > queda a medias en ninguno de los dos lados» es una afirmación sobre el estado remoto que **el
  > comando no puede sostener**: desde este lado, «la petición no llegó» y «llegó y se perdió la
  > respuesta» son **indistinguibles**, y en el segundo caso el estado remoto **sí** quedó determinado
  > sin que nadie lo sepa. Prometer atomicidad de extremo a extremo sería escribir como garantía algo
  > que no se puede comprobar.
  >
  > **Lo que sí se puede sostener son las tres cosas de arriba**, y la tercera es la que convierte la
  > incertidumbre en tolerable: como la operación es **repetible sin consecuencia**, no hace falta
  > saber si ocurrió. Basta con volver a intentarlo. Esa propiedad no es un accidente afortunado —es
  > consecuencia directa de FR-010 y de que el código no sea de un solo uso— pero **no estaba escrita
  > en ninguna parte**, y sin ella la incertidumbre de FR-013 no tiene salida.

#### Preservación: la garantía de 004 no se toca

- **P-005 FR-014**: **Para toda entrada para la que hoy se deriva una identidad de proyecto, la
  derivación DEBE seguir produciendo el mismo valor y DEBE seguir sin producir error.** La garantía
  está establecida en `specs/004-identidad-de-proyecto/spec.md:249` (P-004 FR-010: *«La emisión de un
  evento **NUNCA** se interrumpe, se pospone ni se descarta por no poder resolver o normalizar su
  identidad de proyecto»*) y en `specs/004-identidad-de-proyecto/contracts/project-identity.md:21`
  (*«**Nunca devuelve error.** Es la forma que exige FR-010: un fallo de resolución no puede propagarse
  hasta detener la emisión de un evento»*). **Esta feature NO la modifica, NO la relaja y NO la
  condiciona.**
- **P-005 FR-015**: El conocimiento que FR-006 necesita —**si se reconoció una raíz de proyecto o
  no**— **DEBE** obtenerse **exponiendo esa información por una vía adicional**, disponible para quien
  la pida. **NUNCA DEBE** obtenerse haciendo fallar, interrumpir o alterar el desenlace de lo que hoy
  no falla.

  > **Por qué se formula así y no enumerando casos.** La tentación es «que la derivación devuelva error
  > cuando no encuentre raíz». Eso rompería P-004 FR-010 **en el camino de la ingesta**, que es donde
  > la garantía existe: un lote entero se detendría porque un directorio dejó de existir. La condición
  > de arriba admite cualquier forma de exponer el dato y prohíbe exactamente una: **que el camino
  > existente cambie de comportamiento**. Enumerar casos habría dejado fuera el que a nadie se le
  > ocurrió.
- **P-005 FR-016**: Añadir esa vía **NUNCA DEBE** alterar el valor de identidad que se estampa en los
  eventos, ni las condiciones bajo las que se estampa.

#### Frontera, canal y secretos

- **P-005 FR-017**: La operación **DEBE** exigir **transporte seguro**, con la **misma exigencia y sin
  ninguna diferencia** respecto a la emisión de eventos. **NUNCA** existe exención, modo de desarrollo,
  variante ni ajuste que la relaje. **Es la misma frontera**, y una frontera con dos puertas de
  distinta altura no es una frontera.
- **P-005 FR-018**: La operación **DEBE** ser **interactiva**: transmite y espera el desenlace, **un
  solo intento**, y **NUNCA** entra en ninguna cola ni mecanismo de reintento diferido.

  > Encolar la petición de alguien que está delante de la pantalla esperando una respuesta es
  > **mentirle**: recibiría un desenlace que no es el de su operación, y la operación real ocurriría
  > más tarde, sin nadie mirando.
- **P-005 FR-019**: El comando **NUNCA** persiste nada en local: ni el código, ni el Proyecto, ni su
  denominación, ni el hecho de haberse unido. **El efecto vive en el servidor.**
- **P-005 FR-020**: **Ningún** mensaje, de éxito o de error, **DEBE** reproducir el código de adhesión
  ni ninguna credencial, ni completos ni en fragmentos.
- **P-005 FR-021**: El desenlace de **éxito** se comunica por **stdout**; **todo** mensaje de error o
  de rehúse, por **stderr**. Es la regla ya establecida en este repositorio: **stdout es la respuesta,
  stderr es todo lo demás**, para que quien canalice la respuesta a un fichero o a otro programa no se
  encuentre avisos mezclados con el dato.
- **P-005 FR-022**: La frontera de datos **no se amplía**: esta feature **NUNCA** transmite rutas,
  nombres de directorio, contenido ni ningún dato que no cruzara ya. La identidad de proyecto cruza en
  la **misma forma irreversible** en la que ya cruza con cada evento.
- **P-005 FR-023**: El código se presenta **como argumento del comando**, y **DEBE existir también la
  vía de leerlo de la entrada estándar**. Las dos vías **DEBEN** producir desenlaces idénticos: la vía
  elegida **NUNCA** es observable en el resultado.

  > **Por qué la segunda vía es obligatoria y no una comodidad.** Un valor pasado como argumento
  > **suele** quedar registrado por el intérprete de órdenes de quien lo teclea, y **en muchos sistemas**
  > queda además a la vista de cualquiera que pueda enumerar los procesos. **Nada de eso lo controla el
  > comando**, y por eso no se promete aquí: lo que el comando **sí** puede ofrecer, y este requisito
  > exige, es que **exista una vía que no obligue a ponerlo en la línea de órdenes**. Qué haga después
  > el entorno con lo que se teclee es del entorno.
  >
  > Este repositorio **ya resolvió exactamente este problema** para otro valor sensible presentado por
  > línea de órdenes: aquí se sigue ese precedente en vez de inventar uno nuevo, para que las dos
  > operaciones sensibles del agente se manejen igual y nadie tenga que recordar cuál se comporta de
  > qué manera.

### Key Entities

- **Código de adhesión**: el valor que alguien con privilegios ha acuñado para un Proyecto y ha
  repartido. Es **opaco**: no significa nada por sí mismo y no revela a qué Proyecto pertenece hasta
  que se presenta. **No es de un solo uso** —sirve a cuantas instalaciones se presenten mientras siga
  vigente—, que es lo que lo hace útil para un equipo. Es un **secreto de reparto**, no una credencial
  de autenticación: no sustituye al enrolamiento y no da acceso por sí solo.
- **Identidad de proyecto de esta instalación**: el valor irreversible que esta instalación deriva del
  árbol de trabajo y estampa en cada evento. Es **lo que se une**, y por eso FR-004 exige que sea el
  mismo que el comando presenta.
- **Proyecto**: la agrupación, propiedad de la organización, bajo la que se reúnen varias identidades
  de instalación. Vive **enteramente en la plataforma**; el agente no lo crea, no lo lista y no lo
  conoce más allá de la denominación que recibe como desenlace.
- **Unión**: el hecho de que una identidad pertenece a un Proyecto. Vive **solo en la plataforma**, es
  **única por identidad** —una identidad no está en dos Proyectos a la vez— y su efecto sobre el
  consumo es **retroactivo sin reproceso**.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: **Se pone ROJO si las dos identidades divergen.** Se demuestra que **comparten origen**
  —lo que FR-005 exige— y no que coincidan caso por caso: **(a)** existe un punto único del que salen
  las dos, y alterarlo cambia **las dos a la vez**; **(b)** sobre un conjunto de árboles de trabajo que
  incluye al menos la raíz de un proyecto, un subdirectorio profundo suyo, un árbol paralelo y un
  directorio sin raíz, la identidad presentada y la estampada se comparan carácter a carácter y **son
  iguales en todos**; y **(c)** una alteración deliberada en ese punto único hace **fallar** la
  comparación de (b). Sin (c) el criterio no es falsable.

  > **Por qué no «para todo árbol de trabajo».** Nadie recorre «todo árbol», así que ese enunciado no
  > se puede poner rojo — y su salida natural, «se cumple por construcción», es exactamente la que este
  > proyecto rechaza: un verde cuyo sujeto no se ha comprobado dice «no se ha mirado». Lo que sí se
  > puede demostrar es **que comparten derivación**, y eso es lo que se pide arriba.
- **SC-002**: Los dos desenlaces que la plataforma declara indistinguibles producen salidas **byte a
  byte idénticas** y el **mismo** resultado del proceso. Se exigen **tres** cosas, y las tres a la vez:
  **(a)** cada salida es **no vacía** y **del tipo que le corresponde** —la de éxito **nombra un
  Proyecto**, comprobado porque **contiene la denominación que existe en ese momento**, no porque
  coincida con un texto fijado—; **(b)** las dos son idénticas **comparadas entre sí**; y **(c)** la
  comparación **sabe fallar**: introducida una diferencia deliberada en una de las dos, (b) se pone
  **rojo**.
- **SC-003**: Dos códigos no utilizables **por causas distintas** producen mensajes **byte a byte
  idénticos**, con las mismas tres exigencias: **(a)** cada mensaje es **no vacío** y **es un rechazo
  de código** —comprobado porque **no contiene ninguna denominación de Proyecto ni ningún fragmento del
  código**, que es lo que lo distingue de los otros desenlaces—; **(b)** los dos son idénticos entre
  sí; **(c)** una diferencia deliberada en uno de ellos pone **rojo** la comparación.

  > **Por qué (a) y por qué así.** Sin (a), «byte a byte idénticas» lo satisfacen **dos salidas
  > vacías**, o dos procesos que revientan igual antes de escribir nada: el criterio pasaría en verde
  > justo cuando el comando está roto del todo. Y (a) se formula **por propiedad y no por literal**
  > —«nombra un Proyecto», «no contiene denominación ni código»— porque comparar contra un texto
  > esperado es lo que D-005-5 prohíbe: pasaría en verde aunque las dos salidas hubieran divergido a la
  > vez del texto de referencia.
- **SC-004**: Lanzado fuera de un árbol de proyecto, el comando emite **exactamente cero** peticiones.
  **Quién observa la ausencia**: un destino instrumentado que **cuenta** las peticiones que recibe y
  está declarado como destino del comando durante la prueba. **Cómo se demuestra que el observador
  funciona**: el **mismo** destino, en el caso positivo —el comando lanzado **dentro** de un árbol de
  proyecto—, **registra exactamente una**. Cero y uno se miden con el mismo instrumento, en la misma
  prueba.
- **SC-005**: La salida completa —ambos canales— de **cada uno** de los desenlaces enumerados **no
  contiene ninguna subcadena de OCHO O MÁS caracteres consecutivos del valor presentado**. Se
  verifica generando las subcadenas de longitud ocho del valor y buscándolas todas en la salida:
  **cero apariciones**.

  > **Por qué ocho, y por qué sobre el valor completo.** Ocho es lo bastante largo para que no choque
  > por azar con prosa en español, y lo bastante corto para cazar una filtración **en cuanto empieza**,
  > en vez de esperar a que se escape el valor entero. Y se aplica **al valor completo, sin exceptuar
  > su parte pública**, porque se comprobó desenlace a desenlace que **ninguno de los siete mensajes
  > que la especificación exige necesita nombrar parte alguna del código**: el de éxito nombra el
  > Proyecto (FR-002), y los seis restantes nombran dónde ejecutar (FR-007), cómo enrolarse (FR-008),
  > lo hallado en la configuración (FR-009), la pertenencia ajena (FR-011), la no utilizabilidad
  > (FR-012) y la imposibilidad de establecer el desenlace (FR-013). **No hay ninguna excepción que
  > tallar**, así que la regla no la tiene.
  >
  > *(La formulación anterior decía «ni fragmento suyo de longitud significativa». Eso no se evalúa:
  > da verde siempre, y un criterio que no puede ponerse rojo no es un criterio.)*
- **SC-006** *(**VALIDACIÓN CONTRA PLATAFORMA REAL — NO AUTOMATIZABLE EN ESTE REPOSITORIO**)*: Tras
  una unión con éxito, el consumo **anterior** a la unión de esa instalación aparece contabilizado bajo
  su Proyecto, **sin que se haya emitido ni reenviado ningún evento**. Se verifica comparando el número
  de eventos antes y después: **debe ser el mismo**.

  > **Por qué se marca, y qué está prohibido hacer con él.** Lo que este criterio observa es
  > comportamiento **de la plataforma** —la agrupación resuelta en lectura—, y la suite de este
  > repositorio **no puede montarlo**: tendría que fabricar la agrupación, y entonces estaría
  > comprobando su propio simulacro. Es un criterio de **validación manual contra una instalación real**
  > y así debe ejecutarse. **Un test automático que lo finja es peor que no tenerlo**: daría verde
  > permanente sobre la propiedad que justifica la feature entera.
- **SC-007**: Tras cualquier ejecución del comando —con cualquier desenlace, incluidos los de rehúse y
  los de error—, los artefactos locales de la instalación quedan **sin modificar**. **Qué se observa**:
  **el conjunto completo** de artefactos locales que la instalación mantiene —configuración, estado de
  lectura, cola de envío, secretos— **enumerado, no «la configuración» a secas**, capturado **íntegro
  antes** de la ejecución y comparado byte a byte **contra esa captura previa**, no contra sí mismo.
  **Cómo se demuestra que el observador funciona**: en el caso positivo —una operación de la
  instalación que **sí** modifica el estado local— **la misma comparación falla**. Si no falla ahí,
  no distingue «no cambió» de «no miré», y el criterio **no cuenta como pasado**.
- **SC-008**: El comando **no completa la operación** contra un destino sin transporte seguro, en
  **todos y cada uno** de los casos de un conjunto **enumerado y reproducible**: **(a)** destino en
  claro; **(b)** destino en claro sobre la máquina local; **(c)** destino sin transporte seguro y con
  un código **utilizable**, para que el rechazo no pueda atribuirse al código; y **(d)** los tres
  anteriores repetidos con **cada** ajuste de configuración que la instalación admita, para sostener
  que **ninguno** altera el resultado. **Cuatro clases enumeradas, no «el 100 % de los intentos»**: un
  porcentaje sobre un conjunto que nadie define se cumple ejecutando un solo caso.
- **SC-009**: Las identidades que la derivación produce **antes y después** de esta feature son
  **idénticas** para el mismo conjunto de entradas, y **ninguna** entrada que hoy no produce error pasa
  a producirlo. Se verifica contra la línea base de identidades ya establecida y versionada por 004:
  `specs/004-identidad-de-proyecto/baseline-sc004.tsv`.
- **SC-010**: Una petición del comando **nunca** aparece en la cola de envío diferido, en **ninguna**
  circunstancia, incluida la de un servidor inalcanzable. **Quién observa la ausencia**: el contenido de
  la propia cola, inspeccionado **antes y después** de la ejecución. **Cómo se demuestra que el
  observador funciona**: en el caso positivo —una emisión ordinaria de eventos con el destino
  igualmente inalcanzable— **la misma inspección SÍ registra el aumento**. Si la cola no crece en
  ninguno de los dos casos, el observador no está mirando y el criterio **no cuenta como pasado**.

  > **Por qué la segunda mitad no es opcional en SC-004 ni en SC-010.** Los dos verifican una
  > **ausencia**, y una ausencia se satisface trivialmente **por no mirar**: un observador mal
  > conectado, un destino que no es el que el comando usa, una cola que se inspecciona en otro sitio.
  > Exigir que el mismo instrumento registre el caso positivo es lo único que distingue «no ocurrió» de
  > «no me enteré».
- **SC-011**: cubre las dos cosas que hasta ahora **no medía ningún criterio** — las dos vías de
  presentar el código (FR-023) y el reparto de canales (FR-021).

  **(A) Las dos vías de entrada.** Presentado el **mismo** código por argumento y por entrada estándar,
  **bajo las mismas condiciones**, los dos desenlaces son **idénticos**: **cada canal por separado,
  byte a byte**, y el **mismo resultado del proceso**. Con **las mismas tres piezas de SC-002**:
  **(a)** cada salida es **no vacía** y **del tipo que le toca**; **(b)** son **idénticas comparadas
  entre sí**; **(c)** la comparación **sabe fallar**, demostrado introduciendo una diferencia
  deliberada en una de las dos.

  **(B) El reparto de canales.** En un desenlace de **éxito**: **stdout no vacío** y **stderr sin el
  desenlace**. En un desenlace de **rehúse o error**: **stderr no vacío** y **stdout VACÍO**. Los dos
  canales se capturan **por separado**, y eso es parte del criterio: **capturados mezclados, la
  comprobación no distingue nada** —una salida combinada no vacía es compatible con cualquier reparto,
  incluido el equivocado— así que un montaje que los una **no cuenta como pasado**.

  > **Por qué hacía falta.** FR-023 y FR-021 eran los dos únicos requisitos que **ningún criterio
  > medía**. Y (A) sin su pieza (c) sería otra vez el «indistinguible» hueco que este documento ya ha
  > corregido **dos veces** —en SC-002 y en SC-003—: dos salidas vacías son indistinguibles entre sí, y
  > el criterio pasaría en verde justo cuando el comando está roto del todo. **SC-002 no cubre (A)**:
  > su sujeto son los dos desenlaces **de la plataforma** —haberse unido ahora y estar ya unido—, no las
  > dos **vías de entrada**, que es una pareja distinta.

---

## Decisiones tomadas durante la especificación

Registradas aquí porque son de Basilio, están fechadas, y su ausencia haría que la spec pareciera
haber elegido sola.

- **D-005-1 · La identidad se deriva por árbol de trabajo, no por punto de lanzamiento** (2026-08-18).
  Consecuencia directa de 004. Es lo que permite lanzar el comando desde cualquier subdirectorio.
- **D-005-2 · El rehúse fuera de un árbol de proyecto es local y previo a cualquier petición**
  (2026-08-18). Se descartó la alternativa —emitir y dejar que la plataforma decida— porque la
  plataforma **no puede** distinguir una identidad legítima de una espuria: las dos tienen la misma
  forma. El único sitio donde ese dato existe es la máquina de quien lanza el comando.
- **D-005-3 · La garantía de 004 se preserva exponiendo información, no propagando fallos**
  (2026-08-18). Formulada como FR-015, por condición y no por enumeración.
- **D-005-4 · Sin taxonomía de códigos de salida** (2026-08-18). El repositorio tiene hoy dos
  resultados de proceso, éxito y fallo. Ampliarlos es una decisión de interfaz de línea de comandos
  que afecta a **todos** los comandos, no solo a éste, y esta feature no la necesita. Queda fuera de
  alcance de forma explícita para que nadie la tome de pasada.
- **D-005-5 · La indistinguibilidad de desenlaces es requisito, no cortesía** (2026-08-18). Se escribe
  como FR-010 y se verifica comparando las dos salidas **entre sí**, nunca contra un literal esperado:
  una comparación contra literal pasaría en verde aunque las dos salidas hubieran divergido a la vez.
- **D-005-6 · El código se presenta como argumento, y la entrada estándar es obligatoria, no opcional**
  (2026-08-18). Escrito como FR-023. Se descartó ofrecer **solo** el argumento: en muchos entornos el
  valor **suele** quedar registrado por el intérprete de órdenes —algo que el comando **no controla** y
  que por eso no se promete— y quien no lo quiera se quedaría sin más salida que renunciar a la
  feature. Lo que sí depende del comando, y es lo que se decide aquí, es que **exista una vía que no
  obligue a poner el código en la línea de órdenes**. Se sigue el precedente que este repositorio ya
  estableció para otro valor sensible presentado por línea de órdenes, en vez de inventar una
  convención nueva para la segunda operación sensible del agente.

---

## Assumptions

- **El código de adhesión llega por un canal ajeno a esta feature.** Cómo se acuña y cómo se reparte es
  asunto de la plataforma y de cada equipo; aquí solo se presenta.
- **La organización se determina a partir del enrolamiento existente**, no se declara ni se elige al
  unirse. Es coherente con cómo se determina hoy para la emisión de eventos.
- **La plataforma ya ofrece la capacidad de resolver una adhesión** y sus desenlaces están establecidos
  y demostrados; esta feature es el lado del agente y **no negocia** esos desenlaces.
- **La identidad de esta instalación depende de dos cosas del entorno que esta feature no controla, y
  cualquiera de las dos puede hacerla divergir en silencio:**
  - **El secreto local que particulariza las identidades.** Si se pierde, **todas** las identidades de
    la instalación cambian, y las uniones previas quedan referidas a identidades que ya no emiten.
  - **El directorio personal, del que depende dónde se detiene la búsqueda de la raíz de proyecto.**
    Un entorno alterado —otro usuario efectivo, otro directorio personal declarado— puede **detener la
    búsqueda en otro punto** y producir, para el mismo árbol de trabajo, **una identidad distinta**.

  **Las dos comparten lo que importa aquí, y por eso van juntas**: afectan **exactamente igual** a la
  identidad que se estampa en los eventos, de modo que FR-004 se sigue cumpliendo —las dos siguen
  coincidiendo— y **el desenlace sigue siendo un éxito perfectamente formado**. Lo que cambia es a qué
  identidad se unió uno. Son condiciones **preexistentes y conocidas** del agente, **ajenas a esta
  feature**, y **no se abordan aquí**: se dejan escritas porque quien diagnostique un «me uní y no
  aparece nada» necesita saber que existen antes de sospechar de la unión.
- **La denominación del Proyecto es texto destinado a una persona**, y se presenta tal como llega, sin
  interpretarla. **Y se supone NO VACÍA, porque la plataforma no admite lo contrario.** Esta
  especificación **no define comportamiento** para una denominación vacía: si algún día la plataforma
  llegara a admitirla, **habría que revisar la pieza (a) de SC-002**, que usa «nombra un Proyecto» como
  marca de tipo y se quedaría sin sujeto.
- **No se asume conectividad estable**: el desenlace de servidor inalcanzable es un camino de primera
  clase (FR-013), no una excepción.

---

## Fuera de alcance

Enumerado con su motivo, para que la ausencia no se lea como olvido.

- **Abandonar un Proyecto, o consultar a cuál se pertenece.** Son operaciones distintas, con sus
  propias preguntas de privacidad —quién puede desagrupar, y qué revela una consulta a quien solo tiene
  el agente—. Unirse no las necesita.
- **Listar Proyectos desde el agente.** Convertiría el agente en una ventana al catálogo de la
  organización a partir de un device enrolado, que es exactamente el oráculo que FR-011 y FR-012 evitan.
- **Que alguien sin privilegios acuñe su propio código.** Acuñar es el acto que crea la agrupación y es
  privilegiado por diseño; permitirlo desde el agente anularía el control de quien administra.
- **Guardar cualquier cosa del Proyecto en la configuración local.** El efecto vive en el servidor
  (FR-019). Una copia local sería un segundo estado que puede quedar obsoleto y contradecirlo.
- **Cambiar el formato del enrolamiento.** Es un contrato compartido entre dos repositorios; tocarlo
  exige coordinarlos y esta feature no lo necesita.
- **Una taxonomía de códigos de salida.** Ver D-005-4: hoy hay dos resultados, y ampliarlos es una
  decisión de interfaz que afecta a todos los comandos.
- **Reprocesar o migrar el consumo histórico.** No hace falta: el efecto es retroactivo por
  construcción (FR-003). Se enumera precisamente porque su ausencia podría leerse como una tarea
  pendiente.
