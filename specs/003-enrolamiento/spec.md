# Especificación: Enrolamiento de dispositivo — emparejar el agente `permea` con su backend

**Feature Branch**: `003-enrolamiento`
**Created**: 2026-07-11
**Status**: Draft
**Input**: Emparejamiento agente↔backend por el lado del agente: recibir el **enrollment string** que el backend (P-002) revela una sola vez —un envoltorio que empaqueta la **URL del backend** y el **device token**—, decodificarlo, verificar el token contra `/ingest` con un ping de lote vacío, guardarlo de forma segura en local y exponer el estado de enrolamiento — sin debilitar la frontera de datos (Principio I) y consumiendo el contrato de transporte existente sin redefinirlo.

## Clarifications

- El alcance es **el lado del agente** del emparejamiento: recibir, verificar, persistir de forma segura y exponer el estado del device token. La **emisión, revocación y rotación** del token son del backend (P-002) y quedan **fuera** de esta spec.
- Esta feature es **contract-driven**: el esquema del token y el endpoint de ingesta son la fuente de verdad (P-001 / P-002). 003 **cumple** el `transport.md` de la spec 001 y **no lo redefine**; no inventa ningún endpoint de verificación nuevo. El **formato del enrollment string** sí lo **define 003** (es el lado que lo decodifica) en `contracts/enrollment-string.md`; P-002 lo consumirá para emitirlo.
- El argumento de `enroll` es un **enrollment string**: un envoltorio que empaqueta la **URL del backend** y el **device token**. Es **encoding, no cifrado**: contiene el token en claro. El agente lo decodifica en URL + token y usa ambos para el ping de verificación; el token que P-001 verifica sigue siendo el mismo.
- La verificación reutiliza la ingesta existente: un **ping de lote vacío** (cero eventos, cero metadato) contra `/ingest`. Una respuesta de aceptación confirma el enrolamiento; un rechazo de autenticación lo invalida.
- El device token —y, por contenerlo, el enrollment string— es un secreto **del mismo calibre que el `salt`**: se protege con la misma disciplina (Principio I y restricciones técnicas de la constitución). El enrolamiento **no** toca la frontera de datos: no amplía la allowlist ni añade contenido a ningún payload.
- El detalle técnico (formato del fichero de config, biblioteca de cliente, mecanismo exacto de permisos por SO) corresponde al plan; esta especificación describe **qué** ocurre y **qué garantías** cumple.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Enrolar el agente con un enrollment string válido (Priority: P1)

Un desarrollador acaba de obtener del backend, una sola vez, el **enrollment string** de su instalación (un envoltorio que empaqueta la URL del backend y el device token). Ejecuta `permea enroll <enrollment-string>` pegando ese valor. El agente lo decodifica en URL + token, verifica el token contra esa URL y, si es válido, lo guarda de forma segura en la máquina para poder transmitir metadato a partir de ese momento.

**Why this priority**: Es el propósito central de la feature y el MVP. Sin un enrolamiento válido, el agente no puede autenticarse ante el backend y todo el resto del producto (medición, transmisión) queda inutilizable. Entrega valor por sí solo.

**Independent Test**: Se prueba de extremo a extremo ejecutando `permea enroll <enrollment-string>` con un enrollment string que el backend acepta y comprobando que (a) el agente confirma el enrolamiento y (b) el token queda persistido de forma segura, sin haber tocado la frontera de datos.

**Acceptance Scenarios**:

1. **Given** un enrollment string válido (URL del backend + device token) y un backend alcanzable por HTTPS, **When** el usuario ejecuta `permea enroll <enrollment-string>`, **Then** el agente lo decodifica en URL + token, envía un ping de ingesta de **lote vacío** a esa URL, recibe una respuesta de aceptación (2xx) y **confirma** el enrolamiento.
2. **Given** un enrolamiento recién confirmado, **When** se inspecciona el fichero de configuración local, **Then** contiene el endpoint (la URL decodificada) y el token, y tiene permisos **0600** (solo el dueño lee/escribe), en la ruta de configuración resuelta por el sistema operativo.
3. **Given** el comando de enrolamiento en ejecución, **When** se captura toda su salida (stdout, stderr y logs), **Then** ni el enrollment string ni el token que contiene aparecen en ninguna parte salvo el propio argumento que el usuario pegó.
4. **Given** un enrollment string cuya URL usa el esquema `http://` en claro, **When** el usuario intenta enrolar, **Then** el transporte lo rechaza conforme al contrato (TLS obligatorio) y el enrolamiento aborta sin persistir nada.

---

### User Story 2 — Rechazo seguro de un token inválido o revocado (Priority: P2)

Un usuario introduce un token equivocado, caducado o que el backend ya ha revocado. El agente lo detecta durante la verificación y rechaza el enrolamiento, sin dejar en la máquina un token inútil y sin filtrar el secreto en ningún mensaje.

**Why this priority**: Protege dos propiedades clave: que no quede un estado corrupto («enrolado» con un token que no funciona) y que el secreto no se filtre en la ruta de error. Depende de que exista el flujo de enrolamiento (US1), por eso es P2, pero es imprescindible para la robustez y la higiene del secreto.

**Independent Test**: Se prueba ejecutando `permea enroll <enrollment-string>` cuyo token el backend rechaza (401/403) y verificando que no queda ningún token persistido y que ni el enrollment string ni el token aparecen en la salida de error.

**Acceptance Scenarios**:

1. **Given** un enrollment string cuyo token es inválido o está revocado, **When** el usuario ejecuta `permea enroll <enrollment-string>` y el ping de lote vacío recibe **401/403**, **Then** el agente **rechaza** el enrolamiento y **no** persiste el token.
2. **Given** un enrolamiento rechazado, **When** se inspecciona el estado del sistema, **Then** es **indistinguible** de no haber intentado enrolar: no hay token inútil en el fichero de configuración.
3. **Given** un enrolamiento rechazado, **When** se lee el mensaje de error mostrado al usuario, **Then** explica el fallo **sin** incluir el token en claro.
4. **Given** un backend inalcanzable o que responde con error transitorio (5xx / error de red), **When** el usuario intenta enrolar, **Then** el agente **no** confirma ni persiste el enrolamiento e informa de que no pudo verificarse, sin filtrar el token.

---

### User Story 3 — Consultar el estado de enrolamiento (Priority: P3)

Un usuario quiere saber si su agente está enrolado y contra qué backend está apuntando, sin tener que ver ni manipular el secreto. Ejecuta `permea status`.

**Why this priority**: Aporta visibilidad y confianza (¿estoy conectado? ¿a dónde envío?), pero es una operación de solo lectura que no bloquea la capacidad de enrolar ni de transmitir; por eso es P3.

**Independent Test**: Se prueba ejecutando `permea status` en dos estados —tras un enrolamiento válido y sin enrolar— y verificando que informa el estado correcto y la URL del backend cuando aplica, sin mostrar nunca el token.

**Acceptance Scenarios**:

1. **Given** un agente enrolado, **When** el usuario ejecuta `permea status`, **Then** informa que está **enrolado** y muestra la **URL** del backend contra el que lo está.
2. **Given** un agente enrolado, **When** se lee la salida de `permea status`, **Then** el token **no** se muestra en claro (ni completo ni de forma que permita reconstruirlo).
3. **Given** un agente sin enrolar, **When** el usuario ejecuta `permea status`, **Then** informa que **no está enrolado**, sin error y sin exponer secreto alguno.

---

### Edge Cases

- **Falta el argumento**: `permea enroll` sin enrollment string termina con un error de uso claro y no persiste nada.
- **Enrollment string malformado**: si el argumento no decodifica a un par URL + token válido, el agente aborta con un error que **no** reproduce el argumento y no persiste nada.
- **Fichero de configuración preexistente con permisos laxos**: al persistir un token verificado, el agente **fija** los permisos a 0600 y no confía en el umask heredado.
- **Re-enrolamiento seguro**: ejecutar `permea enroll` con un nuevo enrollment string válido reemplaza el token almacenado (última escritura gana); la sobrescritura **no** deja el token viejo en disco (ni ficheros `.bak` ni temporales con el secreto). La rotación/revocación como ciclo de vida es del backend y queda fuera de alcance.
- **Respuesta ambigua del backend** (2xx sin cuerpo esperable, o corte de red tras enviar): el agente solo confirma con una aceptación explícita; ante la duda, no persiste un enrolamiento «a medias».

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El comando `permea enroll <enrollment-string>` DEBE aceptar como argumento un **enrollment string** que empaqueta la **URL del backend** y el **device token en claro** (el valor que el backend genera y revela una sola vez). El agente DEBE **decodificarlo** en URL + token y usar **ambos** para el ping de verificación. El token que P-001 verifica sigue siendo el mismo; el enrollment string es solo el envoltorio de entrega.
- **FR-002**: Antes de persistir el token, el agente DEBE **verificarlo** contra el backend reutilizando el **contrato de transporte existente** (un ping de ingesta de **lote vacío** contra `/ingest`); NUNCA DEBE inventar un endpoint de verificación nuevo.
- **FR-003**: Una **respuesta de aceptación** del backend (2xx) DEBE confirmar el enrolamiento; una respuesta **401/403** (token inválido o revocado) DEBE rechazarlo.
- **FR-004**: Ante un token rechazado o no verificable, el agente NUNCA DEBE persistir el token; el estado del sistema tras un rechazo DEBE ser indistinguible de no haber intentado enrolar (no queda token inútil).
- **FR-005**: Un token verificado DEBE persistirse en el **fichero de configuración local** con permisos **0600** (solo el dueño lee/escribe).
- **FR-006**: La ruta del fichero de configuración DEBE **resolverse por sistema operativo** (macOS, Linux y Windows nativo) y NUNCA DEBE hardcodearse.
- **FR-007**: El device token es un secreto del **mismo calibre que el `salt`**: NUNCA DEBE escribirse en logs, en stdout (salvo el acto por el que el usuario lo pega como argumento) ni en mensajes de error. El **enrollment string** hereda esta disciplina: es encoding, no cifrado, y contiene el token en claro, por lo que NUNCA DEBE loguearse, mostrarse ni filtrarse en errores.
- **FR-008**: El comando `permea status` DEBE informar si el agente está **enrolado** y contra **qué backend (URL)**; NUNCA DEBE mostrar el token en claro ni de forma que permita reconstruirlo.
- **FR-009**: El enrolamiento NUNCA DEBE ampliar la **allowlist** de la frontera ni añadir contenido a ningún payload (Principio I, no negociable).
- **FR-010**: El ping de verificación DEBE ser un **lote vacío**: cero eventos y cero metadato; no transporta ningún dato de la frontera.
- **FR-011**: El transporte del ping DEBE cumplir el contrato existente: **HTTPS obligatorio** (un endpoint `http://` en claro se rechaza) y autenticación con el device token en la cabecera `Authorization`. 003 **cumple** `transport.md`; NUNCA lo redefine.
- **FR-012**: El **esquema del token** y el **endpoint de ingesta** son la fuente de verdad (P-001 / P-002); 003 los **consume** y NUNCA los define.
- **FR-013**: El **formato del enrollment string** DEBE especificarse en `specs/003-enrolamiento/contracts/enrollment-string.md` (lo generará el plan) como fuente de verdad de ese envoltorio. 003 **define** este formato porque es el lado que lo decodifica; P-002 (backend) lo **consumirá** para emitirlo. Este contrato NUNCA DEBE redefinir el esquema del token ni el endpoint de ingesta (que siguen siendo de P-001 / P-002).
- **FR-014**: El re-enrolamiento DEBE **sobrescribir** el token almacenado sin dejar el token viejo en disco: NUNCA DEBE quedar una copia del secreto en ficheros `.bak`, temporales u otros residuos tras reemplazarlo.

### Key Entities

- **Enrollment string**: envoltorio de entrega que empaqueta la **URL del backend** y el **device token en claro**; el backend lo genera y lo revela una sola vez. Es encoding, no cifrado; el agente lo decodifica en URL + token. Su formato es la fuente de verdad definida en `contracts/enrollment-string.md`. Secreto del mismo calibre que el `salt` (contiene el token).
- **Device token**: credencial de instalación que autentica al dispositivo ante el backend; el backend guarda solo su hash; el agente lo extrae del enrollment string y conserva la forma usable únicamente en el fichero local protegido. Secreto del mismo calibre que el `salt`.
- **Configuración local**: fichero legible por el usuario que asocia el device token con el endpoint del backend (la URL decodificada del enrollment string); su ruta se resuelve por SO y sus permisos son 0600.
- **Estado de enrolamiento**: condición derivada de la presencia de un token verificado y persistido junto a un endpoint conocido; es lo que `permea status` expone (sin revelar el secreto).
- **Ping de verificación (lote vacío)**: petición de ingesta con cero eventos y cero metadato, usada exclusivamente para validar el token frente al backend sin cruzar la frontera de datos.

## Success Criteria *(mandatory)*

- **SC-001**: Un usuario enrola el agente con **un único comando** (`permea enroll <enrollment-string>`) y, con un enrollment string válido, obtiene una **confirmación** explícita de enrolamiento.
- **SC-002**: Tras un enrolamiento exitoso, el fichero de configuración que contiene el token tiene permisos **0600** (solo el dueño), verificable en macOS, Linux y Windows nativo.
- **SC-003**: Ni el enrollment string ni el device token que contiene aparecen **NUNCA** en la salida estándar, los logs ni los mensajes de error del agente, verificable inspeccionando toda la salida de `enroll` y `status` (salvo el argumento que el usuario pega).
- **SC-004**: Un enrolamiento con token inválido o revocado (401/403) se **rechaza** y no deja **ningún** token persistido: el estado posterior es idéntico al de no haber enrolado.
- **SC-005**: `permea status` informa correctamente el estado (**enrolado / no enrolado**) y, si enrolado, la **URL** del backend, sin revelar el token.
- **SC-006**: El ping de verificación transmite un **lote vacío** —cero eventos y cero campos de metadato, verificable en el cuerpo de la petición— y la allowlist de la frontera **no** se amplía.
- **SC-007**: El agente resuelve la ruta de configuración **por SO** sin rutas hardcodeadas: `enroll` y `status` operan sobre la ubicación estándar de cada plataforma.
- **SC-008**: El enrolamiento **reutiliza** el contrato de transporte existente (HTTPS + Bearer) y **no introduce ningún endpoint nuevo**; una URL `http://` en claro se rechaza y el enrolamiento aborta.
- **SC-009**: El ping de verificación (lote vacío) **no persiste ni altera estado** en el backend más allá de confirmar la autenticación: no crea eventos ni deja rastro de metadato (verificable en el estado del backend antes/después del ping).
- **SC-010**: Un re-enrolamiento con un enrollment string válido reemplaza el token y **no deja el token viejo en disco** (sin ficheros `.bak` ni temporales con el secreto), verificable inspeccionando la ruta de configuración tras la operación.

## Assumptions

- El **enrollment string** (que empaqueta URL + device token) lo **emite P-002** (backend, ya implementado); 003 solo lo consume y decodifica. Emisión, revocación y rotación del token quedan **fuera de alcance**.
- El endpoint `/ingest` de **P-001** (ya implementado) acepta un **lote vacío** y responde 2xx para un token válido; esa semántica es la base del ping de verificación.
- La **URL del backend** contra el que enrolar sale del **enrollment string** (viene empaquetada junto al token); no depende de configuración previa ni de un valor por defecto. El agente la decodifica y la persiste junto al token verificado.
- El **override del token por variable de entorno** (p. ej. para CI) queda **fuera de alcance**; posible spec futura.
- El **desenrolar/rotar** el token queda fuera de alcance: hoy se resuelve revocando y reemitiendo desde el backend (P-002).
