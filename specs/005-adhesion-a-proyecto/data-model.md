# Data Model — Adhesión a proyecto

**Feature**: `005-adhesion-a-proyecto` | **Fecha**: 2026-08-18

---

## No hay modelo de datos, y este es el motivo

**Esta feature no persiste nada.** No es que el modelo sea pequeño: **es que no existe**, y por
requisito explícito:

> **P-005 FR-019**: El comando **NUNCA** persiste nada en local: ni el código, ni el Proyecto, ni su
> denominación, ni el hecho de haberse unido. **El efecto vive en el servidor.**

No hay tabla, ni fichero nuevo, ni campo añadido a `config.json`, ni entrada en `state.json`, ni línea
en `queue.jsonl`. **El directorio de datos del agente queda byte a byte igual** después de cualquier
ejecución del comando —con cualquier desenlace, incluidos los de rehúse y los de error—, y eso **está
verificado**, no supuesto: es **SC-007**, con su observador y su caso positivo.

**Un `data-model.md` vacío no valdría.** El silencio se leería como olvido en un flujo que genera este
artefacto para cada feature. Lo que sigue es lo que **sí** hay: valores que atraviesan la operación y
mueren con ella.

---

## Lo que sí hay: tres valores en tránsito

Ninguno sobrevive al proceso. Se enumeran porque **son el dato de la feature**, aunque no sean estado.

| Valor | De dónde viene | A dónde va | Vida |
|---|---|---|---|
| **El código de adhesión** | de fuera: argumento del comando **o** entrada estándar (FR-023) | al servidor, en la petición | **Muere con el proceso.** Nunca se escribe, y **nunca aparece en ningún mensaje** (FR-020, verificado por SC-005 con su umbral de ocho caracteres) |
| **La identidad de proyecto** | **derivada** del árbol de trabajo, con el secreto local | al servidor, en la petición | **Muere con el proceso.** No es nueva: es **exactamente la misma** que la instalación estampa en sus eventos (FR-004), por **la misma derivación** y no por dos que hoy coincidan (FR-005) |
| **La denominación del Proyecto** | **del servidor**, en el desenlace de éxito | a la pantalla de quien ejecutó (stdout, FR-021) | **Muere con el proceso.** No se cachea, no se guarda, no se reutiliza (FR-019) |

### Y dos que se LEEN y no se tocan

| Valor | Dónde vive | Qué se hace con él |
|---|---|---|
| **`endpoint`** | `config.json`, escrito por el enrolamiento | Se **lee** y de él se **deriva** el destino de la adhesión, con validación ruidosa (`plan.md` D-005-P3). **No se reescribe** |
| **`device_token`** | `config.json`, escrito por el enrolamiento | Se **lee** para autenticar. **No se reescribe, y nunca aparece en un mensaje** (FR-020) |

---

## Por qué la ausencia de estado es una propiedad, no una carencia

**El efecto de la unión es retroactivo sin que el agente haga nada**, y esa es la razón de fondo por la
que no hay nada que guardar:

> **P-005 FR-003**: La feature **NUNCA** emite, reenvía, reescribe ni reprocesa eventos de consumo. El
> alcance retroactivo de la unión lo aporta la plataforma al resolver la agrupación **en lectura**, y
> **el comando no tiene que hacer nada para conseguirlo**.

La agrupación vive **enteramente del lado del servidor**. En cuanto la unión existe, el histórico
entero de esa instalación ya cuenta bajo su Proyecto. Guardar una copia local de «a qué Proyecto
pertenezco» crearía **un segundo estado que puede quedar obsoleto y contradecir al primero** — y por
eso está **explícitamente fuera de alcance** en la spec, con ese motivo escrito.

**Consecuencia para quien implemente**: si en algún momento parece que hace falta guardar algo, **eso
no es un detalle de implementación: es una violación de FR-019**, y la spec está congelada.

---

## Entidades del dominio — dónde viven de verdad

La spec describe cuatro en §Key Entities. **Ninguna es del agente**, y conviene tenerlo delante:

| Entidad | Dónde vive | Qué sabe el agente de ella |
|---|---|---|
| **Código de adhesión** | acuñado y persistido **en la plataforma** (solo su hash) | Lo recibe de fuera y lo presenta. **No lo interpreta**: es opaco por diseño |
| **Identidad de proyecto** | **derivada en local**, no almacenada | La deriva cada vez. No hay registro de identidades en el agente |
| **Proyecto** | **enteramente en la plataforma** | Solo la **denominación** que recibe como desenlace, y solo durante el proceso |
| **Unión** | **solo en la plataforma** | **Nada.** El agente no puede consultarla — y consultarla está fuera de alcance |

---

## Verificación

Que no hay modelo **no se declara: se comprueba**.

- **SC-007** — tras cualquier ejecución, el conjunto enumerado de artefactos locales
  —configuración, estado de lectura, cola de envío, secretos— queda **sin modificar**, comparado byte a
  byte contra una captura previa, **y con el caso positivo que demuestra que el observador funciona**.
- **SC-010** — la petición **nunca** aparece en la cola de envío diferido, ni siquiera con el servidor
  inalcanzable, **con la cola inspeccionada antes y después** y su caso positivo.

Las dos mitades importan: sin el caso positivo, «no cambió» no se distingue de «no miré».
