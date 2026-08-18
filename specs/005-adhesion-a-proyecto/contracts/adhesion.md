# Contrato — Adhesión a proyecto

Operación por la que **una instalación del agente se une a un Proyecto de su organización presentando
un código de adhesión**, y recibe como desenlace la **denominación** del Proyecto al que ha quedado
unida. Cumple FR-001, FR-002, FR-010, FR-011, FR-012 y FR-013.

**Interfaz pública de la plataforma, colocada en el repositorio público** (`plan.md` D-005-P11,
decisión de Basilio del 2026-08-18). La **implementación** de la adhesión es privada y seguirá
siéndolo; **su interfaz no puede serlo**: un cliente de código abierto solo puede depender de un
contrato abierto, y el Principio II exige que *«un revisor externo DEBE poder confirmar el cumplimiento
de la frontera leyendo únicamente el código»*. Una remisión a una ruta privada lo impediría — el
auditor se toparía con un enlace que no puede abrir justo donde está lo que vino a verificar.

Este contrato **NO** redefine el esquema del device token ni el endpoint de ingesta, que siguen siendo
fuente de verdad de P-002 / P-001 (`../../001-agente-inicial/contracts/transport.md`), ni el formato
del enrolamiento (`../../003-enrolamiento/contracts/enrollment-string.md`). La adhesión **reutiliza**
la autenticación de la ingesta sin tocarla.

> ⚠️ **Ventana conocida.** A′ tiene una segunda mitad —que la plataforma **deje de definir** este
> contrato y **pase a citarlo**— que **todavía no ha ocurrido**: es trabajo de otro día, en el otro
> repositorio. **Hasta entonces el contrato está definido en dos sitios.** Verificado contra la
> implementación de la plataforma el **2026-08-18**; si los dos divergieran antes de que se cierre la
> ventana, **manda la implementación**, y este documento es el que hay que corregir.

---

## Superficie

```
POST  <base>/projects/adhesion
```

La adhesión es **la segunda puerta de la frontera de datos**, y la primera ampliación desde P-001.
**Su pila de autenticación es EXACTAMENTE la de la ingesta**, elemento a elemento: canal seguro
obligatorio, autenticación por device token, y organización resuelta **del device**.

De esa pila salen **dos garantías que nadie tiene que recordar**:

- **El código de adhesión NO autentica ni sustituye al enrolamiento.** Sin un device válido no se llega
  a evaluar el código.
- **La organización sale del device, NUNCA del cuerpo.** Un `org_id` en la petición **ni se lee**.

### Cómo se obtiene `<base>` — y por qué está en este contrato

El agente **no guarda una URL base**: guarda **la ruta completa del endpoint de ingesta**, tal como se
la entregó el enrolamiento. Los dos hechos siguientes son **contrato, no detalle local**, y el agente
los usa para derivar el destino de la adhesión (`plan.md` D-005-P3):

| # | Hecho | Qué implica |
|---|---|---|
| **1** | **La ruta de ingesta termina en el segmento `/ingest`** | Es lo que el agente **valida** antes de derivar. Si el endpoint guardado no tiene esa forma, **rehúsa sin emitir petición** en vez de conjeturar (FR-009) |
| **2** | **El segmento de la adhesión es `/projects/adhesion`**, hermano de `/ingest` bajo el mismo prefijo | Es lo que el agente **pone en su lugar**, conservando esquema, host, puerto y prefijo |

**Ejemplo de la derivación:**

```
guardado:   https://api.permea.example/api/v1/ingest
derivado:   https://api.permea.example/api/v1/projects/adhesion
                                       └── prefijo conservado ──┘
```

> **Por qué estos dos hechos están aquí y no solo en el código del agente.** Si vivieran como literales
> en el cliente, serían **dos suposiciones sobre un servidor ajeno sin nada que las respalde**: el día
> que la plataforma moviera una ruta, el agente fallaría con «configuración de forma inesperada»,
> **culpando al usuario de un cambio del servidor**. Documentados aquí, el fallo tiene dónde mirarse —y
> quien escriba otro cliente sabe qué puede dar por bueno.

---

## La petición

```json
{
  "type": "object",
  "required": ["code", "project_ref"],
  "properties": {
    "code":        { "type": "string", "description": "el código de adhesión en claro, COMPLETO, con su prefijo" },
    "project_ref": { "type": "string", "description": "la identidad de proyecto de esta instalación, en su forma irreversible" }
  }
}
```

**Struct de dos campos.** A diferencia del `pmea2`, que es cerrada y rechaza campos extra, **aquí un
campo no contemplado se IGNORA**: es la semántica de descarte del Principio I —lo que no está en la
allowlist no se lee—, no un olvido. **No existe rechazo por campo extra.**

**Sobre `code`:** se envía **el claro completo, prefijo incluido** (ver §El código de adhesión). El
agente **no lo interpreta, no lo trocea y no lo valida más allá de su prefijo**: es opaco por diseño.

**Sobre `project_ref`:** es la identidad que esta instalación estampa en sus eventos para ese mismo
árbol de trabajo, **derivada por la misma vía** (FR-004, FR-005). No es un identificador nuevo ni una
forma distinta del mismo valor. La frontera **no se amplía**: cruza en la misma forma irreversible en
la que ya cruza con cada evento.

**Autenticación:** device token, igual que la ingesta. **No se envía `org_id`.**

---

## El código de adhesión

```
pmeaj1.<base64url-sin-padding( 32 bytes aleatorios )>      →  7 + 43 = 50 caracteres
```

- **Prefijo de versión**: literal `pmeaj1.` (incluye el punto). Es la **única versión válida**. Un
  prefijo distinto o ausente **no se intenta interpretar ni migrar**: es el desenlace 1.
- **256 bits de entropía**, la misma que el device token.
- **Token OPACO: sin payload.** El diseño asume que **puede filtrarse**, y un payload convertiría cada
  filtración en fuga de identificadores internos.
- **No es de un solo uso y no caduca.** Sirve a cuantas instalaciones se presenten mientras siga
  vigente — es lo que lo hace útil para un equipo y no para una sola instalación. Deja de valer solo
  cuando se revoca.
- **Como mucho uno vivo por Proyecto**, garantizado por la plataforma.
- **El claro se revela una sola vez**, al acuñarlo. La plataforma persiste solo su hash y **no lo
  re-sirve**.

> ⚠️ **Se hashea el claro COMPLETO, prefijo incluido.** Quien resuelva la adhesión hashea **el valor
> tal como llega**. Se documenta aquí aunque el agente no hashee nada, porque es el acuerdo que hace
> que el código funcione: si una de las dos partes hasheara solo el cuerpo, **ningún código válido
> resolvería jamás**, y el fallo se vería como «todos los códigos son desconocidos» —**indistinguible
> de funcionar mal a propósito**—.

---

## Los cuatro desenlaces, en ESE orden

**La utilizabilidad del código se evalúa ANTES que el estado de la identidad.** El orden es contrato,
no implementación: invertirlo **filtraría que una identidad está mapeada a quien ni siquiera tiene un
código válido**.

| # | Condición | Estado | Cuerpo |
|---:|---|:---:|---|
| **1** | Código **no utilizable**: inexistente · de otra organización · revocado · prefijo desconocido · `project_ref` no conforme | **`422`** | `{"error":"adhesion_rejected"}` |
| **2** | Utilizable · identidad ya en **OTRO** Proyecto | **`409`** | `{"error":"identity_already_assigned"}` — **sin denominación** |
| **3** | Utilizable · identidad ya en **ESE MISMO** Proyecto | **`200`** | `{"project":{"name":"…"}}` |
| **4** | Utilizable · identidad **sin mapear** | **`200`** | `{"project":{"name":"…"}}` |

**Verificado contra la implementación** el 2026-08-18, no solo contra documentación:
`ProjectAdhesionController.php:159` (`422`, y es **un único `return` físico** al que llegan las cinco
causas desde `:65`, `:73`, `:85` y `:91`) · `:103` (`409`) · `:146` (el cuerpo del éxito, **sin segundo
argumento**, luego `200` por defecto) · y los desenlaces **3 y 4 son literalmente la misma llamada**,
`confirmacion($proyecto)`, desde `:110` y `:133`.

### Qué distingue a qué — lo que un tercero necesita para implementarlo

| Par | ¿Se distinguen? | Por qué medio | Qué debe hacer el cliente |
|---|:--:|---|---|
| **éxito (3, 4) vs. no-éxito (1, 2)** | **Sí** | **El estado**: `200` frente a `4xx` | Ramificar por el estado |
| **1 vs. 2** | **Sí** | **Por los dos**: estado distinto (`422` / `409`) **y** cuerpo distinto (`adhesion_rejected` / `identity_already_assigned`) | **Ramificar por el ESTADO**, que es el discriminante barato y el que no depende de interpretar el cuerpo. El cuerpo sirve de confirmación, no de discriminante |
| **3 vs. 4** | **NO, y no puede** | — | **Nada.** Son idénticos en estado y en cuerpo, y salen del mismo método. **El cliente no debe intentar distinguirlos** |
| **Entre las 5 causas del 1** | **NO, y no puede** | — | **Nada.** *«Ni el cuerpo, ni el código HTTP, ni las cabeceras distinguen las cinco»* (`ProjectAdhesionController.php:158`) |
| **Cualquier OTRO estado** (`5xx`, `403`, `404`, un `2xx` que no sea `200`…) | — | **El estado**: no es ninguno de los tres contemplados | **NO VERIFICABLE** (FR-013). **NUNCA éxito, y NUNCA un rechazo**: afirmar un rechazo también es afirmar un desenlace |

> ⚠️ **La primera fila dice «`200` frente a `4xx`», y eso NO es el reparto completo.** Distingue
> *éxito* de *no-éxito* **entre los cuatro desenlaces contratados**, y nada más. **Un estado fuera de
> los tres no es un rechazo**: es una respuesta que no permite determinar el desenlace, y cae en la
> cláusula de §Reglas para el cliente. La fila de arriba está para que nadie lea la primera como
> exhaustiva y escriba `if 200 { éxito } else { rechazo }` — que convertiría un `500` en un **rechazo
> afirmado**, exactamente lo que FR-013 prohíbe.
>
> *(Añadido tras validar el cliente contra esta tabla: la cláusula general de §Reglas para el cliente
> ya lo cubría, pero **la tabla es lo que se lee primero**.)*

> **Un éxito cuyo cuerpo no se pueda interpretar no es un éxito.** El estado `200` no basta: FR-002
> exige comunicar la denominación, así que un `200` sin `project.name` legible se trata como **no
> verificable** (FR-013), no como unión conseguida.

### Las tres propiedades que el cliente NO puede romper

1. **El desenlace 1 es indistinguible por construcción.** Las cinco causas producen **exactamente la
   misma respuesta** — el ajeno y el revocado ni siquiera se encuentran, porque la consulta va acotada
   a la organización del device y a los códigos vivos. **El cliente NUNCA debe intentar deducir la
   causa**, ni presentarla: convertiría el comando en un oráculo para averiguar qué códigos existen.
2. **3 y 4 son idénticos, y eso es deliberado.** Si se distinguieran, quien pega un código dos veces
   sabría **cuál de las dos surtió efecto**. **El cliente debe presentarlos idénticos**: mismo texto,
   mismo canal y **mismo resultado del proceso** — incluido el código de salida.
3. **El desenlace 2 NUNCA nombra el Proyecto ajeno.** La plataforma no lo revela, y **el cliente no
   puede inventar lo que el servidor calla**.

### Efecto: retroactivo, y sin ejecutar nada

**Solo el desenlace 4 escribe.** La agrupación se resuelve **en lectura**, así que en cuanto la unión
existe, **todo el histórico de esa instalación ya cuenta bajo su Proyecto**. **No se toca ni un evento
de consumo**: no hay migración, ni reproceso, ni reenvío — ni en el servidor ni en el cliente.

---

## Reglas para el cliente

El agente **DEBE**, en este orden, y ante cualquier fallo **abortar sin persistir nada y sin reproducir
el código** en el error:

1. **Comprobar que hay un árbol de trabajo con raíz reconocible.** Si no, **rehusar sin emitir
   petición** (FR-006): la identidad de un directorio suelto no es la del proyecto que se quería
   agrupar, y la plataforma **no puede distinguirlo** —aceptaría la unión con un éxito perfectamente
   formado sobre la identidad equivocada—.
2. **Comprobar que hay enrolamiento.** Si no, rehusar e indicar cómo enrolarse (FR-008).
3. **Derivar el destino** del endpoint guardado, con la validación de §Cómo se obtiene `<base>`. Si la
   forma no es la esperada, **rehusar nombrando la forma de lo hallado** —nunca material sensible—
   (FR-009, y FR-020 manda sobre él).
4. **Exigir canal seguro.** Sin exención, sin modo de desarrollo, sin variante: **es la misma frontera
   que la ingesta** (FR-017).
5. **Transmitir y esperar**: un solo intento, **sin cola ni reintento diferido** (FR-018). Encolar la
   petición de alguien que está mirando la pantalla sería mentirle.
6. **Interpretar el desenlace** según la tabla, presentando 3 y 4 de forma **indistinguible**, 2 **sin
   nombrar el Proyecto ajeno**, y 1 **sin indicar la causa**.
7. **No persistir nada** (FR-019). El efecto vive en el servidor.

**Si el desenlace no puede establecerse** —servidor inalcanzable, o respuesta que no permite
determinarlo—, el cliente **informa de que no se pudo completar** y **NUNCA afirma ningún desenlace**
(FR-013). El estado remoto queda **indeterminado**, y esa incertidumbre es **inocua**: como el código
no se agota y unirse dos veces es indistinguible de unirse una, **repetir la operación no tiene
consecuencia** (FR-013a).

> **Un éxito sin denominación no es un éxito.** FR-002 exige comunicar el nombre del Proyecto; una
> respuesta de éxito cuyo cuerpo no se pueda interpretar se trata como **no verificable**, no como
> unión conseguida.

---

## Seguridad

- **El código de adhesión es un secreto de reparto, no una credencial de autenticación.** No da acceso
  por sí solo y no sustituye al enrolamiento; sin device válido no se evalúa siquiera.
- **NUNCA se hace eco.** Ni completo ni en fragmentos, en ningún mensaje, de éxito o de error (FR-020).
  El umbral es verificable: **ninguna subcadena de ocho o más caracteres** del valor presentado puede
  aparecer en la salida (SC-005).
- **El prefijo `pmeaj1.` es un patrón reconocible a propósito**, para que los escáneres de secretos lo
  detecten si se filtra — **el mismo motivo que `../../003-enrolamiento/contracts/enrollment-string.md:130`
  da para `pmea2.`**. *(Fuente: contrato de la plataforma `P013-proyectos/contracts/join-code.md:68-70`,
  que lo declara con esas palabras y remite al precedente.)* Y la **j** del prefijo es de *join*
  (`join-code.md:22`).
- **El cliente no persiste el código** en ningún momento, ni siquiera transitoriamente en disco.
- **La frontera no se amplía**: la adhesión transmite exactamente dos valores, y los dos ya cruzaban o
  son de la misma clase que los que cruzan. **Cero rutas, cero nombres de directorio, cero contenido.**

---

## Compatibilidad y evolución

- **La autoridad de la interfaz es este documento**; la **implementación** es de la plataforma (P-013).
  Cualquier cambio de los desenlaces, de la forma de la petición o de los segmentos de ruta **debe
  reflejarse aquí**, o los clientes de código abierto dejarán de poder depender de él.
- **`pmeaj1.` es la única versión del código.** Una futura `pmeaj2.` se introduciría con su propia
  política de versión, como hizo `pmea2` con `pmea1`.
- **Un campo nuevo en la petición no rompe a los clientes existentes**: la struct **ignora** lo no
  contemplado. Un campo **requerido** nuevo sí rompería, y exigiría versionar.
- **Mientras la ventana de A′ siga abierta** —la plataforma todavía define su lado—, **cualquier
  divergencia se resuelve a favor de la implementación**, y este documento se corrige.
