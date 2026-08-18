# Contrato — Comando `permea project join`

Contrato observable del comando de usuario de 005. Define **la gramática, las entradas, las salidas,
los canales y los códigos de salida**. No prescribe implementación.

**El protocolo NO está aquí**: los cuatro desenlaces de la adhesión, la forma de la petición y los
estados HTTP son [`adhesion.md`](./adhesion.md). Aquí va **la superficie**; allí, **el protocolo**.

> **Forma.** Sigue la de [`003-enrolamiento/contracts/cli.md`](../../003-enrolamiento/contracts/cli.md)
> —comando, entrada, comportamiento y salidas, garantías DEBE/NUNCA— y **no** la de
> `004/contracts/cli-config.md`, que es de otra clase: 004 contrata **una configuración y su efecto
> sobre los comandos**, mientras que 003 y esto contratan **un comando**. Mismo sujeto, misma forma.

---

## La gramática — **el primer comando de dos niveles del repositorio**

```
permea project join [<código>]
        └──────┘ └──┘
       sustantivo verbo
```

Hasta 005 el binario tiene **dos subcomandos, los dos planos** —`enroll` y `status`— y cuatro flags
—`--scan`, `--run`, `--daemon`, `--version`—. **`project join` estrena el segundo nivel.**

| Regla | Contrato |
|---|---|
| **Despacho** | El primer nivel (`project`) se resuelve **antes del parseo de flags**, igual que `enroll` y `status`. El segundo nivel (`join`) lo resuelve el propio `project` |
| **Flags intactos** | `--scan`, `--run`, `--daemon` y `--version` **conservan su comportamiento sin cambio alguno**. La gramática nueva **NUNCA** los intercepta ni los reordena |
| **`project` sin verbo** | **Error de uso**: mensaje por **stderr** con los verbos disponibles, **exit `1`** (§Los códigos de salida). NUNCA hace nada por defecto |
| **Verbo desconocido** | `permea project <lo-que-sea>` → **error de uso**, por **stderr**, **exit `1`**, nombrando el verbo no reconocido y los disponibles. **NUNCA** se intenta interpretar ni corregir |
| **Ayuda** | La ayuda del binario **DEBE** listar `project join` junto a `enroll` y `status`. Un comando que no aparece en la ayuda no existe para quien lo busca |

**Por qué sustantivo + verbo y no un verbo plano**: deja sitio a la familia (`project leave`,
`project status`) sin volver a cambiar la gramática, y alinea el vocabulario con el de la plataforma,
que llama «Proyecto» a lo que se une. *(Decisión: `plan.md` D-005-P6.)*

---

## Entrada

El código de adhesión (`pmeaj1.…`, ver [`adhesion.md`](./adhesion.md)) se acepta por **dos vías
equivalentes que producen el mismo flujo** (FR-023):

| Vía | Invocación | Notas |
|---|---|---|
| Argumento posicional | `permea project join <código>` | Cómodo, pero en muchos entornos el valor **suele** quedar registrado por el intérprete de órdenes y a la vista de quien pueda enumerar procesos. **El comando no controla eso** |
| stdin (**recomendada**) | `permea project join` (sin argumento) o `permea project join -` | Se lee de stdin (p. ej. `… \| permea project join -`). **NUNCA se hace eco** |

**Lo que el comando SÍ garantiza** es que **existe una vía que no obliga a poner el código en la línea
de órdenes**. Qué haga después el entorno con lo que se teclee es del entorno.

| Situación de entrada | Efecto |
|---|---|
| Con argumento posicional | Usa ese valor |
| Sin argumento, con stdin disponible (pipe) | Lee el código de stdin y sigue el flujo |
| Sin argumento y sin stdin (terminal interactiva sin pipe) | **Error de uso, exit `1`**. NUNCA un prompt que se cuelgue esperando |

**La ayuda del comando DEBE mencionar la vía stdin como la recomendada**, igual que hace `enroll`.

> **Las dos vías son indistinguibles en el resultado**: mismo texto, mismo canal y mismo código de
> salida. **La vía elegida NUNCA es observable en la salida** (FR-023, verificado por SC-011 (A)).

---

## Comportamiento y salidas

**Los tres rehúses locales se evalúan en este orden, y el orden es contrato**: **árbol → enrolamiento
→ configuración**. Más específico primero, y cada uno nombra el error que la persona puede corregir
ahora. *(Motivo: `plan.md` D-005-P13.)*

| # | Situación | Efecto | Canal | Exit |
|---:|---|---|:--:|:--:|
| **R1** | **Fuera de un árbol de trabajo con raíz reconocible** | Rehúsa. **NO emite ninguna petición** | stderr | **1** |
| **R2** | **Sin enrolamiento** | Rehúsa e indica cómo enrolarse. **NO emite petición** | stderr | **1** |
| **R3** | **Configuración de forma inesperada** (no permite determinar el destino) | Rehúsa nombrando **la forma** de lo hallado. **NO emite petición** | stderr | **1** |
| **D4** | **Unión nueva** — código utilizable, identidad sin mapear | Se une; comunica **la denominación del Proyecto** | stdout | **0** |
| **D3** | **Ya unido a ESE MISMO Proyecto** | Comunica la denominación. **Indistinguible de D4** | stdout | **0** |
| **D2** | **Ya unido a OTRO Proyecto** | Lo informa **sin nombrar el Proyecto ajeno** | stderr | **1** |
| **D1** | **Código no utilizable** | Lo informa **sin indicar la causa** | stderr | **1** |
| **NV** | **No verificable** — servidor inalcanzable o respuesta que no permite establecer el desenlace | Informa de que no se pudo completar. **NUNCA afirma ningún desenlace** | stderr | **1** |

### El reparto de canales (FR-021)

**stdout es la respuesta; stderr es todo lo demás.** En concreto, y es verificable:

- **En un desenlace de éxito**: **stdout no vacío** y **stderr sin el desenlace**.
- **En un desenlace de rehúse o error**: **stderr no vacío** y **stdout VACÍO**.

**Los dos canales se comprueban por separado.** Capturados mezclados, la comprobación no distingue
nada: una salida combinada no vacía es compatible con cualquier reparto, incluido el equivocado.
*(Verificado por SC-011 (B).)*

### Los códigos de salida — **dos valores, ocho desenlaces**

El binario tiene hoy **`0` y `1`**, y esta feature **no amplía el vocabulario**. La distinción viaja
en el **mensaje**, no en el número.

> **El ERROR DE USO sale con `1`.** Cubre los tres casos que **no son desenlaces de la adhesión**
> —`project` sin verbo, verbo desconocido, y entrada ausente sin pipe (§La gramática, §Entrada)—, y por
> eso **no figura en la tabla de arriba**: en los tres, **el comando no llegó a intentar la adhesión**,
> así que no hay desenlace que numerar. Pero **el valor sí queda fijado aquí**: `1`, como todo lo que
> no es éxito.
>
> **Y se fija con un número y no con «≠ 0» a propósito.** «Distinto de cero» deja elegir `2`, o `70`,
> y los deja a los dos conformes con el contrato — con lo que un test que compare «≠ 0» **pasa contra
> cualquier andamiaje** que también salga distinto de cero, y deja de probar nada.

> ### ⚠️ **D3 y D4 comparten código de salida, y no es negociable**
>
> El resultado del proceso es **observable** —es lo primero que mira un script—, así que darle a «ya
> estabas unido» un código propio **rompería FR-010**: quien pega un código dos veces sabría cuál de
> las dos surtió efecto. Es exactamente lo que la plataforma declara indistinguible a propósito
> ([`adhesion.md`](./adhesion.md) §Los cuatro desenlaces).

---

## Garantías (DEBE / NUNCA)

**Sobre el secreto:**

- **NINGÚN** mensaje, de éxito o de error, **DEBE** reproducir el código ni ninguna credencial, ni
  completos ni en fragmentos. El umbral es verificable: **ninguna subcadena de ocho o más caracteres**
  del valor presentado puede aparecer en la salida (SC-005).
- Leído por **stdin**, el código **NUNCA** se hace eco a la terminal ni queda en salida alguna, y
  **NUNCA** aparece en la lista de procesos.

**Sobre el estado local:**

- El comando **NUNCA** persiste nada: ni el código, ni el Proyecto, ni su denominación, ni el hecho de
  haberse unido (FR-019). **El efecto vive en el servidor.**
- Tras **cualquier** ejecución —con cualquier desenlace, incluidos los de rehúse y error— los
  artefactos locales quedan **byte a byte iguales** (SC-007).

**Sobre la red:**

- **DEBE** exigir transporte seguro, **con la misma exigencia que la emisión de eventos y sin ninguna
  exención, modo de desarrollo ni variante** (FR-017). Es la misma frontera.
- **DEBE** transmitir y esperar: **un solo intento**, **sin cola ni reintento diferido** (FR-018). Una
  petición de este comando **NUNCA** aparece en la cola de envío, **ni siquiera con el servidor
  inalcanzable** (SC-010).
- Los tres rehúses locales ocurren **antes de emitir nada**: en R1, R2 y R3 el número de peticiones
  emitidas es **exactamente cero** (SC-004).

**Sobre lo que no se revela:**

- **NUNCA DEBE** nombrar el Proyecto ajeno en D2, ni indicar la causa en D1, ni distinguir D3 de D4.
  El comando **NUNCA** infiere ni conjetura lo que el servidor calla.

**Sobre la identidad:**

- La identidad que presenta **DEBE** ser **exactamente la misma** que la instalación estampa en sus
  eventos para ese mismo árbol de trabajo, y **por la misma derivación** — no por dos que hoy
  coincidan (FR-004, FR-005).

---

## Notas

- Este comando se añade al binario existente **sin alterar** los flags de P-001/P-002
  (`--scan`/`--run`/`--daemon`/`--version`) ni los subcomandos `enroll` y `status`.
- **No amplía la allowlist de la frontera** (Principio I): el único tráfico que genera es la petición
  de adhesión, con **dos valores** que ya cruzaban o son de la misma clase que los que cruzan.
- **Un endpoint no-https se rechaza antes de transmitir**, por la misma guarda que la emisión de
  eventos. A diferencia de `enroll` —donde el rechazo ocurre ya en la decodificación del enrollment
  string—, aquí el endpoint **viene de la configuración persistida**, así que la guarda del transporte
  **sí se ejercita** en este flujo.
- **La consulta de a qué Proyecto se pertenece, y el abandono, están fuera de alcance** de 005 y no
  tienen verbo. `project` queda con **un solo verbo** hasta que alguna spec añada otro.
