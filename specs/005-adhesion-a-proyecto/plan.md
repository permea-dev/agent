# Implementation Plan: Adhesión a proyecto

**Branch**: `005-adhesion-a-proyecto` | **Date**: 2026-08-18 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/005-adhesion-a-proyecto/spec.md` — **CONGELADA** el
2026-08-18 (ver el bloque de congelación en `checklists/requirements.md`). **Este plan es la última
ventana para editarla.**

---

## Summary

El agente gana **una operación interactiva nueva**: presentar un código de adhesión y quedar unido a un
Proyecto, con el nombre del Proyecto como desenlace. Nada se persiste en local; el efecto vive en el
servidor y es retroactivo por construcción.

**Técnicamente la feature es pequeña y el riesgo está concentrado en cuatro mecanismos que hoy no
sirven tal cual.** El transporte descarta el cuerpo de la respuesta —justo donde viene el nombre del
Proyecto—; la guarda de HTTPS vive dentro del método de ingesta y no se hereda; lo guardado en config
es la ruta completa de ingesta y no una base; y la derivación de identidad oculta el único dato que el
rehúse local necesita. **El plan es, sobre todo, la resolución de esos cuatro, más el tratamiento
explícito de tres criterios que ninguna historia arrastra.**

**La revisión del plan NO encontró ningún requisito inimplementable.** Los veinticuatro se pueden
satisfacer con las decisiones de abajo. No hay parada.

---

## Technical Context

**Language/Version**: Go 1.22.2 (`go.mod`; toolchain medido en la máquina de desarrollo).

**Primary Dependencies**: **ninguna**. Librería estándar exclusivamente — Principio III de la
constitución exige justificar cualquier dependencia externa en la spec que la introduzca, y esta
feature **no necesita ninguna**.

**Storage**: **ninguno nuevo.** FR-019 prohíbe persistir nada; se **lee** `config.json` del directorio
de datos por SO y no se escribe.

**Testing**: `go test ./...` — Pest no, aquí es la librería estándar con `httptest`.
`internal/testutil.Sandbox` para aislar el estado local.

**Target Platform**: Linux, macOS y Windows, como el resto del binario.

**Project Type**: CLI de un solo binario estático.

**Performance Goals**: **no aplican.** La operación es interactiva y de una sola petición; el único
límite relevante es el timeout de cliente ya existente (`transport.go:83`, 10 s).

**Constraints**: HTTPS obligatorio sin exención (FR-017); un solo intento y sin cola (FR-018); cero
escritura local (FR-019); la garantía de 004 intacta (FR-014/015/016).

**Scale/Scope**: un subcomando nuevo, un método nuevo de transporte, una función nueva de derivación
expuesta, y el despacho de dos niveles. **Línea base a preservar: `go test ./...` → 9 paquetes `ok`,
0 `[no test files]`, 0 `FAIL`.**

---

## Constitution Check

*GATE: pasa antes de Phase 0 y se re-verifica tras Phase 1.*

| Principio | Veredicto | Razón |
|---|:--:|---|
| **I · Frontera de datos inviolable** | ✅ | FR-022: no se transmite ningún dato que no cruzara ya. La identidad de proyecto cruza en **la misma forma irreversible**. El código de adhesión **entra**, no sale. Cero rutas, cero nombres de directorio. |
| **II · Privacidad auditable, local-first** | ✅ | La operación es del usuario y explícita. No añade telemetría ni observación de fondo. **No se guarda nada** (FR-019), así que no hay estado nuevo que auditar. |
| **III · Binario único y auditable** | ✅ | Cero dependencias nuevas. Rutas resueltas por SO, sin hardcodear. La decisión **D-005-P2** (unificar la guarda) va justamente en dirección de legibilidad: una condición en un sitio en vez de cuatro. |
| **IV · Test-first en la frontera** | ⚠️ **con acción** | Esta feature **amplía la frontera con una segunda puerta**. El golden test de frontera existente cubre la emisión de eventos; **la adhesión necesita el suyo** —que el cuerpo de la petición contenga exactamente `{código, identidad}` y nada más— y **va antes que el código**, como disciplina de primer commit. Recogido en Phase 2. |
| **V · Desarrollo dirigido por especificaciones** | ✅ | Spec congelada, checklist cerrado, matriz recorrida. Sin salvedad al Principio V en este repositorio (v1.0.0): todo lleva spec, y la lleva. |

**Sin violaciones que justificar** → §Complexity Tracking queda vacía.

---

## Project Structure

### Documentation (this feature)

```text
specs/005-adhesion-a-proyecto/
├── spec.md                    # CONGELADA
├── checklists/requirements.md # cerrado, con la matriz 5 × 11
├── plan.md                    # este fichero
├── research.md                # Phase 0 — ESCRITO (2026-08-18)
├── data-model.md              # Phase 1 — ESCRITO (2026-08-18)
├── contracts/adhesion.md      # Phase 1 — ESCRITO (2026-08-18), D-005-P11 opción A′
├── contracts/cli.md           # Phase 1 — ESCRITO (2026-08-18), forma de 003/contracts/cli.md
├── quickstart.md              # Phase 1 — ESCRITO (2026-08-18), forma de 004/quickstart.md
└── tasks.md                   # Phase 2 — lo genera /speckit.tasks, NO este comando
```

### Source Code (repository root)

```text
cmd/permea/
├── main.go            # MODIFICADO: despacho de dos niveles (D-005-P6)
├── project.go         # NUEVO: runProject + join, patrón de dos capas de enroll.go
├── project_test.go    # NUEVO: comando — entrada, rehúses, canales, salidas, secretos
├── enroll.go          # sin tocar — es el molde
└── status.go          # sin tocar

internal/transport/
├── transport.go               # MODIFICADO: extraer la guarda de esquema (D-005-P2) + Join (D-005-P1)
├── adhesion_test.go           # NUEVO: los cuatro desenlaces + SC-008
├── boundary_adhesion_test.go  # NUEVO: golden de frontera de la adhesión (D-005-P14)
└── queue.go                   # sin tocar

internal/project/
├── resolve.go         # MODIFICADO: exponer «hubo raíz» por vía adicional (D-005-P5)
└── resolve_test.go    # MODIFICADO: SC-001, origen compartido

internal/ingest/
├── baseline_regresion_test.go # NUEVO: SC-009 contra baseline-sc004.tsv — aquí porque las TRES
│                              #        columnas las produce claudecode.go:86-88
└── boundary_test.go           # MODIFICADO: solo la nota de remisión cruzada (D-005-P14)

internal/config/
├── config.go          # MODIFICADO: usar la guarda extraída; derivar destino (D-005-P3)
├── config_test.go     # MODIFICADO: derivación del destino y validación ruidosa
└── enrollment.go      # MODIFICADO: usar la guarda extraída (solo la condición)

internal/testutil/
└── sandbox.go         # sin tocar — se reutiliza (SandboxConSemillas)
```

**Structure Decision**: se conserva la disposición existente. **No se crea ningún paquete nuevo**: la
adhesión es transporte (`internal/transport`), su comando es CLI (`cmd/permea`) y su identidad ya vive
en `internal/project`. Crear un `internal/adhesion` separaría en tres sitios lo que hoy tiene dueño.

---

## LAS CUATRO TRAMPAS DE MECANISMO

### D-005-P1 · El cuerpo de la respuesta — **método nuevo, `Send` intacto**

**El problema, medido**: `transport.go:131` hace `io.Copy(io.Discard, resp.Body)` y `:133-142`
clasifica **solo por código**, con todos los 4xx que no son 401/403 cayendo en la misma rama `default`.
**El nombre del Proyecto no sobrevive, y los dos rechazos que la plataforma distingue se vuelven
indistinguibles para el agente** — que es lo contrario de lo que la spec necesita: FR-011 y FR-012
exigen mensajes **distintos** para «ya perteneces a otro» y «el código no vale».

**La decisión**: **un método nuevo en `Client`, y `Send` no se toca.**

- `Send` seguirá descartando el cuerpo, y **debe seguir haciéndolo**: su comentario `:130` razona que
  vaciarlo permite reutilizar la conexión en los reintentos, y el camino de ingesta drena lotes en
  serie. Cambiarlo por comodidad de esta feature degradaría el camino caliente.
- El método nuevo **lee y decodifica** la respuesta, y **distingue por código**: el desenlace de éxito
  trae el nombre; los dos rechazos son ramas separadas.
- **Reutiliza todo lo demás de `Client`**: el `http.Client` con su timeout, la cabecera de
  autenticación, el `Content-Type`. Solo cambia qué se envía y qué se hace con lo que vuelve.

**Por qué no ampliar `Send` con un parámetro de salida**: convertiría el método del camino caliente en
uno con dos modos, y el modo equivocado en la ingesta sería un error silencioso.

**Riesgo declarado**: el método nuevo **decodifica cuerpo de respuesta**, algo que hoy no hace ninguna
parte del agente. Debe tratar como no-verificable —FR-013, no como éxito— cualquier respuesta cuyo
cuerpo no se pueda interpretar, incluido un cuerpo vacío en el desenlace de éxito. **Un éxito sin
nombre no es un éxito**: FR-002 exige comunicar la denominación.

### D-005-P2 · La guarda de esquema — **se UNIFICA, y es la decisión más importante del plan**

**El problema, medido**: la comprobación está escrita **tres veces** — pero **las tres NO son
idénticas**, y la diferencia es justo la que puede romper la extracción:

```
internal/config/enrollment.go:78-80
    u, err := url.Parse(p.Endpoint)
    if err != nil || u.Scheme != "https" {        ← ANÁLISIS Y ESQUEMA PLEGADOS EN UNA
        return …, fmt.Errorf("%w: el endpoint debe ser https", ErrEnrollmentString)

internal/config/config.go:100-106
    u, err := url.Parse(c.Endpoint)
    if err != nil { return fmt.Errorf("endpoint inválido %q: %w", c.Endpoint, err) }   ← RAMA 1
    if u.Scheme != "https" { return fmt.Errorf("endpoint debe ser https://, got %q", …) } ← RAMA 2

internal/transport/transport.go:105-111
    u, err := url.Parse(c.Endpoint)
    if err != nil { return fmt.Errorf("transport: endpoint inválido %q: %w", c.Endpoint, err) } ← RAMA 1
    if u.Scheme != "https" { return fmt.Errorf("%w: %q", ErrScheme, c.Endpoint) }               ← RAMA 2
```

**Tres diferencias reales, no de estilo:**

1. **`enrollment.go` FUNDE** el fallo de análisis y el esquema equivocado en **un solo desenlace**;
   `config.go` y `transport.go` los **separan en dos**.
2. **`enrollment.go` NUNCA reproduce el endpoint**; los otros dos **sí** lo interpolan con `%q`. Y es
   deliberado: `enrollment.go:30-33` razona que el `pmea2` **contiene el token**, así que su error no
   puede llevar el argumento.
3. Los tres errores son **distintos**, y `transport.go` además envuelve un centinela, `ErrScheme`
   (`:31`), del que dependen tests por `errors.Is`.

Y **la de transporte vive dentro de `Send`**, no en el tipo: **un método nuevo nace sin ella**.

**La decisión: se unifica. No se replica una cuarta vez.** Con la forma que preserva las tres
diferencias:

- **Una función única en `internal/config`** que **devuelve los dos hechos por separado** —«no se pudo
  analizar» y «el esquema no es admisible»— **y no formatea ningún mensaje**.
- **Cada llamante decide qué hace con lo que recibe**, y así conserva su comportamiento exacto:
  `enrollment.go` **funde los dos en su error genérico sin argumento**; `config.go` y `transport.go`
  **los mantienen en dos ramas** con sus mensajes y su centinela.

> **Por qué la función devuelve dos hechos y no un booleano.** Extraer «la condición» a secas —un
> `esAdmisible(string) bool`— **cambiaría el comportamiento de una de las tres**: o `enrollment.go`
> pasa a distinguir dos casos donde hoy hay uno, o los otros dos pierden la distinción que hoy tienen.
> Y el síntoma sería **uno de los cuatro tests de la red en rojo**, con la puerta diciendo «parar» sin
> que nadie entienda por qué. **La unificación correcta unifica el juicio, no la presentación.**

**Por qué unificar y no replicar, con tres razones y no una:**

1. **FR-017 no admite que se olvide.** Una condición replicada cuatro veces es una que alguien
   escribirá tres veces bien y una mal, y el fallo es **emitir por canal en claro** — Principio I, no
   una molestia.
2. **El repositorio ya pagó esta clase de defecto.** `MetricsQuery` de la plataforma lo documenta
   textualmente como *«el defecto de clase que P-011 acaba de pagar — una condición replicada,
   corregida en un sitio y olvidada en cuatro»*. La lección está escrita; ignorarla aquí sería
   deliberado.
3. **Principio III pide legibilidad**: *«un desarrollador escéptico debe entender la frontera de un
   vistazo»*. Cuatro copias de la frontera no se entienden de un vistazo.

**La ubicación COMPILA, y está medida — no supuesta.** Es la pregunta que este plan tenía abierta sin
saberlo, porque el grafo de este repositorio tiene una arista incómoda documentada
(`internal/testutil/sandbox.go:11-20`) y había que descartar que fuera ésta.

```
$ go list -deps ./internal/transport | grep permea
  internal/config          ← transport YA importa config, hoy, en producción
  internal/event
  internal/transport

$ go list -deps ./internal/config | grep permea
  internal/config          ← config es HOJA: no importa NADA del módulo

$ go list -f '…TestImports…XTestImports…' ./internal/...
  internal/config  TEST-> encoding/base64 encoding/json os path/filepath strings testing   XTEST-> (vacío)
```

Tres hechos, y los tres apuntan igual: **la arista `transport → config` ya existe**, así que llamar a
la función desde `transport` **no añade ninguna**; **`config` no importa nada del módulo**, ni en
producción ni en sus tests, así que **no puede cerrar ciclo con nadie**; y **sus tres ficheros de test
son internos** (`package config`) pero **solo usan la librería estándar**.

> **El ciclo que `testutil` documenta es real y es OTRO.** `internal/testutil` no puede importar
> `internal/config` porque `config_test.go` es test **interno** y usaría el helper — `testutil→config`
> y `config(test)→testutil` cerrarían el círculo. Esa arista es **`testutil ↔ config`**, no
> `transport ↔ config`. La cautela estaba bien puesta; el caso no es éste.

**Y una obligación que la unificación se lleva consigo: la ENCONTRABILIDAD de la frontera.** El
Principio III pide que *«un desarrollador escéptico entienda la frontera de un vistazo»*, y mover el
juicio a `internal/config` tiene un coste real: **quien vaya a buscar por qué el agente rechaza un
endpoint mirará en `transport`**, que es donde se transmite. Así que la extracción **incluye remisión
cruzada en los dos sitios**: en `transport` —donde antes estaba la condición— un comentario que dice
dónde vive ahora el juicio y por qué se movió; y en `config` —donde vive— quiénes son sus cuatro
llamantes. **Sin las dos remisiones, unificar mejora el código y empeora el hallazgo**, que en un
producto cuyo argumento es la frontera de datos no es un intercambio aceptable.

**Coste y riesgo, declarados**: toca **tres ficheros que hoy están verdes**, y los tres tienen test que
ejerce la guarda (`transport_test.go:151`, `config_test.go:61`, `enrollment_test.go:97`, más
`enroll_reject_test.go:129`). **Esos cuatro tests son la red**: la unificación es correcta si siguen
en verde sin tocarlos. Si hubiera que modificarlos, la unificación cambió comportamiento y **hay que
parar**.

**Y una guarda estructural sobre la guarda**: la extracción **no basta** si el método nuevo se olvida
de llamarla. Se cubre con SC-008, cuyas cuatro clases incluyen el caso «destino sin transporte seguro y
con un código utilizable» — **ejercido contra el camino de la adhesión**, no contra el de ingesta.

### D-005-P3 · Del endpoint guardado al destino — **derivación con validación ruidosa**

**El problema, medido**: `config.Endpoint` es **la URL completa de ingesta**, no una base. La compone
el backend (`DeviceTokenController.php:34`, en la plataforma) y el agente la postea **tal cual**
(`transport.go:117`). Los tests lo confirman: `"https://inseguro.example/ingest"`.

**La decisión, ya tomada por Basilio: derivar con validación ruidosa.** El plan escribe el cómo:

1. **Se analiza** el endpoint guardado como URL. Si no lo es, rehúse por FR-009.
2. **Se exige que su ruta tenga la forma esperada** —que termine en el segmento de ingesta conocido—.
   **Es la validación ruidosa**: no se conjetura, se comprueba.
3. **Si la tiene**, se sustituye ese último segmento por el de la adhesión, conservando esquema, host,
   puerto y prefijo. Todo lo demás del endpoint se respeta **tal cual**, incluido un puerto no estándar
   —el banco de pruebas local usa uno—.
4. **Si NO la tiene**, se **rehúsa antes de emitir nada**, con el mensaje de FR-009: nombrando **la
   forma** de lo hallado —qué se esperaba y qué se encontró— y **nunca** el valor si pudiera contener
   material sensible, porque **FR-020 manda sobre FR-009** y así lo dice el propio requisito.

**Por qué la validación es la mitad importante, y no un adorno**: sin ella, una cirugía de cadena sobre
un endpoint con otra forma produce **una URL plausible que apunta a ninguna parte**, y el desenlace es
un error de red —FR-013, «no se pudo»— que manda a depurar la conexión en vez de la configuración.
Rehusar en local con el motivo cuesta una ejecución; conjeturar cuesta una tarde.

**Lo que NO se hace, y por qué**: no se añade un campo nuevo a la configuración ni al enrolamiento. El
`pmea2` es struct **cerrada** de tres campos (`enrollment.go:39-45`, `:71-77`), así que ampliarlo es un
formato nuevo **coordinado entre dos repositorios** — y esta feature no lo necesita.

> ### ⚠️ Esta decisión INCRUSTA DOS HECHOS DEL CONTRATO EN EL CÓDIGO
>
> La derivación no es una manipulación de cadenas neutra: **sabe dos cosas sobre la interfaz de la
> plataforma**, y las dos son contrato, no implementación local.
>
> 1. **Que la ruta guardada termina en el segmento de ingesta** —es lo que la validación ruidosa
>    comprueba, y sin ese hecho no habría nada que validar—.
> 2. **Cuál es el segmento de la adhesión** —es lo que se pone en su lugar—.
>
> **Los dos TIENEN que estar documentados en `contracts/adhesion.md`** (D-005-P11), y **el código que
> los usa tiene que citarlo**. Si vivieran solo como literales en el agente, serían dos suposiciones
> sobre un servidor ajeno sin nada que las respalde: el día que la plataforma mueva una ruta, el
> agente fallaría con «configuración de forma inesperada» **culpando al usuario de un cambio del
> servidor**. Con el contrato citado, el fallo tiene dónde mirarse.

### D-005-P4 · El envío interactivo — **se sigue el patrón de `Verify()`, sin reservas**

**El precedente, medido**: `Verify()` (`transport.go:91-99`) es **exactamente** lo que FR-018 pide, y
su comentario ya razona el criterio: *«Un solo intento (sin backoff): en el enrolamiento, un 5xx/red
significa "no verificable → no persistir", no reintentar»*.

**La decisión: se sigue.** El método de adhesión llama a `Client.Send`-equivalente **directamente**,
sin pasar por `sendWithRetry` (`transport.go:149`) ni por `Append`/`Drain` (`queue.go:23`, `:73`).

**Y hay que verlo bien: la cola no está *dentro* del cliente, está *encima*.** Quien llama
`Append`+`Drain` encola; quien llama al método directo transmite y espera. **`project join` no tiene
que pelearse con la cola: tiene que no usarla**, y eso no requiere mecanismo, requiere no escribirlo.

**La sustitución del razonamiento de `Verify` al caso de la adhesión es literal**: donde el
enrolamiento dice «no verificable → no persistir», la adhesión dice «no verificable → no afirmar
desenlace» (FR-013). Y **la incertidumbre remota es inocua** porque FR-013a lo establece: repetir no
tiene consecuencia.

**Verificación**: SC-010, con su observador —la cola inspeccionada antes y después, y el caso positivo
que demuestra que la inspección funciona—.

---

## EL MAPEO DESENLACE → CÓDIGO DE SALIDA

**Por qué está aquí y no se dio por obvio.** D-005-4 excluyó **la taxonomía** —inventar códigos nuevos
por clase de fallo—, **no la correspondencia**. Y **SC-002 y SC-011 comparan «el mismo resultado del
proceso»**, así que el código de salida **es parte de lo verificado**: no escribirlo lo dejaría a
criterio de quien implemente, y una elección razonable pero distinta rompería un criterio sin que nadie
lo viera venir.

| # | Desenlace | Requisito | Canal | **Salida** |
|---|---|---|:--:|:--:|
| 1 | **Unión nueva** — se une y se nombra el Proyecto | FR-001, FR-002 | stdout | **0** |
| 2 | **Ya unido al MISMO Proyecto** — repetición | FR-010 | stdout | **0** |
| 3 | **Ya unido a OTRO Proyecto** | FR-011 | stderr | **1** |
| 4 | **Código no utilizable** | FR-012 | stderr | **1** |
| 5 | **Rehúse: fuera de árbol de proyecto** | FR-006, FR-007 | stderr | **1** |
| 6 | **Rehúse: sin enrolamiento** | FR-008 | stderr | **1** |
| 7 | **Rehúse: configuración de forma inesperada** | FR-009 | stderr | **1** |
| 8 | **No verificable** — servidor inalcanzable o respuesta que no permite establecer el desenlace | FR-013, FR-013a | stderr | **1** |

> ### ⚠️ LA RESTRICCIÓN DURA: **1 y 2 comparten código de salida, y no es negociable**
>
> **FR-010 exige que los dos sean indistinguibles**, y **el resultado del proceso es observable** — es
> lo primero que mira un script, y lo que `SC-002` y `SC-011 (A)` comparan explícitamente. Darle a
> «ya estabas unido» un código propio —2, por ejemplo— parecería una mejora de ergonomía y **rompería
> FR-010**: quien pega un código dos veces sabría cuál de las dos surtió efecto, que es exactamente lo
> que la plataforma declara indistinguible a propósito.
>
> **Se escribe aquí porque es una decisión que se toma sola si nadie la escribe**, y en la dirección
> equivocada.

**Ocho desenlaces, dos valores.** Es la consecuencia directa de D-005-4: el repositorio tiene hoy
`0` y `1`, y esta feature **no amplía el vocabulario**. Los ocho caben porque la distinción que importa
—qué pasó— viaja en el **mensaje**, no en el número.

### D-005-P13 · El ORDEN de los tres rehúses locales

Los tres ocurren **antes de emitir nada**, y **más de uno puede darse a la vez** — una instalación sin
enrolar, lanzada fuera de un proyecto, con la configuración rota. **Sin fijar el orden lo decide el
primer test que se escriba**, que es la misma clase de decisión que el código de salida.

**El orden es: (1) árbol de proyecto → (2) enrolamiento → (3) configuración.**

| Orden | Rehúse | Requisito | Por qué aquí |
|:--:|---|---|---|
| **1** | **Fuera de árbol de proyecto** | FR-006, FR-007 | **Es el único que no depende de nada instalado.** Se responde con el directorio actual y punto: ni lee ficheros, ni necesita config, ni necesita enrolamiento. Comprobarlo primero es lo más barato y **lo más específico**: nombra el error real de quien se equivocó de sitio |
| **2** | **Sin enrolamiento** | FR-008 | Depende de leer la configuración, pero **solo de su existencia**. Y su mensaje —«enrólate así»— es **accionable y completo**: quien no está enrolado no necesita saber además que su endpoint tiene mala forma, porque **todavía no tiene endpoint** |
| **3** | **Configuración de forma inesperada** | FR-009 | Es el más caro de evaluar y **el menos específico de los tres**. Solo tiene sentido preguntarlo cuando ya sabemos que hay enrolamiento: si no lo hay, «la configuración tiene forma inesperada» sería **técnicamente cierto y completamente inútil** |

**El criterio que ordena los tres, y vale para el que venga**: **de más específico a menos, y de más
barato a más caro** — que aquí coinciden. Cada rehúse debe nombrar **el error que la persona puede
corregir ahora**; presentar el genérico cuando hay uno específico disponible convierte un diagnóstico
en una adivinanza.

> **El caso que lo justifica y no es teórico**: alguien recién instalado, sin enrolar, prueba el
> comando desde su directorio personal. Con este orden oye *«ejecútalo dentro del árbol que quieres
> agrupar»* —cierto y accionable—. Con el orden inverso oiría *«la configuración no permite determinar
> el destino»*, que **le manda a mirar un fichero que aún no existe**.

---

## LAS OTRAS SEIS PIEZAS

### D-005-P5 · Exponer «hubo raíz o no» — **función nueva, `Derivar` intacta**

**El dato ya existe y está tapado**: `ascender` devuelve `(string, bool)` (`resolve.go:128`) y
`derivarConTecho` (`:104-114`) usa el booleano para elegir entre raíz y fallback, **y luego lo tira**.

**La decisión**: **una función exportada nueva** en `internal/project` que devuelve la identidad **y**
si se reconoció raíz. `Derivar` **no cambia de firma, ni de comportamiento, ni de nombre**, y la
función nueva y `Derivar` **comparten el mismo cuerpo interno** — que es lo que FR-005 exige del par
identidad-presentada / identidad-estampada: **no dos derivaciones que hoy den lo mismo**.

**Por qué así y no de otra forma:**

- **FR-015** exige vía **adicional**, y prohíbe obtener el dato «haciendo fallar, interrumpir o alterar
  el desenlace de lo que hoy no falla». Una función nueva no altera nada.
- **FR-016** exige que el valor estampado no cambie. Compartir cuerpo lo garantiza por construcción.
- **La garantía de 004 se preserva literalmente**: `Derivar` sigue sin devolver error
  (`004/contracts/project-identity.md:21`), y la emisión sigue sin interrumpirse
  (`004/spec.md:249`, P-004 FR-010).

**Cómo se demuestra que la ingesta no cambió** (SC-009, y es la parte que no se puede dar por
supuesta): repitiendo la pasada de referencia **con las semillas deterministas del bloque
REPRODUCCIÓN** de `specs/004-identidad-de-proyecto/baseline-sc004.tsv` y comparando **las tres
columnas** de ese fichero. Es el mismo procedimiento que 004 usó en su T007 para demostrar neutralidad,
y **existe precisamente para esto**. Con semillas distintas los refs no comparan y un «fallo» no
significaría nada.

### D-005-P6 · El despacho de dos niveles — **anidado, y los flags intactos**

**Lo medido**: `main()` mira **solo `os.Args[1]`** con una escalera de `if` (`main.go:44`, `:51`), y
el comentario de `:42-43` declara que se despacha **antes** del parseo de flags *«para no interferir
con los flags de P-001/P-002 (`--scan`/`--run`/`--daemon`/`--version`), que se conservan intactos»*.

**La decisión**: **un tercer `if` en la misma escalera**, para `project`, que delega en un despachador
propio que mira `os.Args[2]`. **Nada más cambia.**

- Se **conserva la propiedad que el comentario declara**: sigue ocurriendo **antes** de `flag.Parse()`,
  así que los cuatro flags no se tocan ni se reordenan.
- El segundo nivel vive **en el fichero del comando**, no en `main.go`: `main.go` gana **tres líneas**,
  no un `switch` anidado.
- Un verbo desconocido bajo `project` es **error de uso**, por stderr y salida 1, como el resto.

**Por qué anidado y no plano** (`permea join`): la plataforma llama a esto «Proyecto», y el sustantivo
deja sitio a los verbos que la spec excluye hoy pero nombra —`leave`, `status`— sin volver a cambiar la
gramática. Un verbo plano cierra esa puerta.

**Y se toca `printUsage`** (`main.go:98-113`), que es un literal mantenido a mano: si el comando existe
y no aparece ahí, no existe para quien lo busca.

### D-005-P7 · La entrada del código — **se sigue el patrón de dos capas de `enroll`**

**La decisión: sí, y sin variantes.** `enroll.go` resolvió exactamente este problema: `runEnroll`
(`:18`) es la capa sucia que resuelve stdin, su naturaleza pipe/TTY y stdout; `enroll` (`:38`) es la
capa pura con **lector, escritor y verificador inyectados**.

- **`readEnrollmentInput` (`enroll.go:93-110`) es reutilizable en su lógica**: argumento si lo hay y no
  es `-`; stdin si es `-` o si no hay argumento y stdin es un pipe; **error de uso sin colgarse** si no
  hay argumento y stdin es una TTY. Ese último caso importa: un comando que se cuelga esperando entrada
  que nadie va a teclear es peor que uno que falla.
- **Es lo que hace la feature testeable sin arrancar procesos**, y eso no es comodidad: `main_test.go:321-325`
  documenta que un proceso hijo **no confiaría** en el certificado autofirmado del arnés sin montarle
  un almacén. **La inyección es la única vía de probar el camino completo aquí.**

**Verificación**: SC-011 (A), con sus tres piezas y la comparación **por canal separado**.

### D-005-P8 · El arnés — **dos instrumentos nuevos, ambos pequeños, ninguno desde cero**

Lo que ya hay y se reutiliza **sin tocar**: `httptest.NewTLSServer` + `srv.Client()` inyectado en
`Client.HTTP` (patrón de `transport_test.go:42-71` y `enroll_test.go:35-41`), y
`internal/testutil.Sandbox` para aislar el estado local con su aserción de aislamiento.

Lo que **falta**, y sale de leer los criterios:

| Necesidad | De dónde sale | Estado |
|---|---|---|
| **Destino que CUENTA peticiones**, con su caso positivo | **SC-004** | **Casi hecho**: el `recorder` de `transport_test.go:39-71` **ya cuenta** (`rec.requests++`). Hay que **elevarlo a helper compartido**, porque el comando vive en otro paquete y hoy es local a los tests de transporte |
| **Inspección de la cola antes y después**, con su caso positivo | **SC-010** | **Nuevo, pero de tres líneas**: `transport.QueuePath` y `transport.Load` son **exportadas** (`queue.go:18`, `:42`). El caso positivo —una emisión ordinaria que **sí** encola con el destino igualmente caído— se monta con lo mismo |
| **Captura de los dos canales POR SEPARADO** | **SC-011 (B)** | **Nuevo**: hoy los tests inyectan un `io.Writer` para stdout (`enroll.go:38`) y el error vuelve como valor. Para comprobar «stdout **vacío** en rehúse» hay que capturar los dos, y **no combinados** — el criterio dice que mezclados no cuenta |
| **Snapshot íntegro del estado local** | **SC-007** | **Nuevo, pequeño**: sobre el sandbox, capturar el conjunto enumerado antes y después. El caso positivo lo da cualquier operación que sí escriba |

**Ninguno es infraestructura grande**, y **el que más valor tiene es el primero**: elevar el `recorder`
a helper compartido lo pone a disposición de las features siguientes, que van a necesitar contar
peticiones igual que ésta.

### D-005-P9 · ⚠️ LOS TRES CRITERIOS SIN HISTORIA — tratamiento explícito

**SC-001, SC-008 y SC-009 no los arrastra ninguna historia de usuario.** La matriz del checklist lo
dejó medido: sus tres columnas están enteras en «—». **Si el troceo se hace leyendo historias, se
caen.** Aquí quedan enganchados a algo que sí se recorre:

| Criterio | Verifica | Quién lo verifica, y **cuándo** |
|---|---|---|
| **SC-001** | identidad presentada == estampada, **por origen compartido** | **D-005-P5**, y es su puerta de aceptación. No es un test más: es la condición que hace que la feature no mienta. Sus tres piezas —punto único, comparación sobre cuatro clases de árbol, y **que la comparación sepa fallar**— se ejercen contra la función nueva y `Derivar` **a la vez** |
| **SC-008** | sin transporte seguro no se completa, en cuatro clases | **D-005-P2**, y es la prueba de que la unificación llegó al método nuevo. La clase (c) —destino inseguro **con código utilizable**— es la que impide que el rechazo se atribuya al código |
| **SC-009** | la línea base de identidades de 004 intacta | **D-005-P5**, contra `specs/004-identidad-de-proyecto/baseline-sc004.tsv` **con sus semillas**. Es la puerta que demuestra que exponer el dato **no cambió el camino de ingesta** |

**La regla para `/speckit.tasks`, escrita aquí para que no se pierda**: **estos tres criterios generan
tarea propia y no cuelgan de ninguna historia.** Un troceo por historias los dejaría fuera, y son
—los tres— los que protegen lo que ya funciona.

### D-005-P10 · Lo que no es automatizable aquí — **la ceremonia, con sujeto elegido**

**SC-006** está marcado en la spec como **validación contra plataforma real, no automatizable en este
repositorio**, y su nota prohíbe expresamente el test que lo finja. Además, dos escenarios llevan la
misma marca: **US1 escenario 2** y **US2 escenario 3**.

**Se recogen como FASE FINAL DE VALIDACIÓN, no como tests.** Sujeto ya elegido:
**`~/dev/test/RecetApp`**, sin mapear, **1 895 eventos reales** —221 del 12 de agosto y 1 674 del 13—,
identificado en el descubrimiento del 2026-08-18 por coincidencia exacta de recuentos.

**Por qué es el sujeto correcto y no uno de laboratorio**: tiene volumen real, está **sin agrupar
ahora mismo**, y su consumo es **anterior** a la unión — que es exactamente lo que SC-006 mide, el
efecto retroactivo. Con un proyecto recién creado no habría histórico que apareciera.

**La ceremonia comprueba, en este orden**: el desenlace nombra el Proyecto · el consumo previo aparece
bajo él · **el número de eventos no cambia** · nada se escribió en local. Y **contra el banco TLS local
ya montado**, porque FR-017 no tiene exención ni siquiera para la ceremonia.

---

## Fases

### Phase 0 — Research

**Casi todo hecho en el descubrimiento del 2026-08-18**, que midió las cuatro trampas con
`fichero:línea` y verificó que no existe ninguna vía de exención de TLS (barrido de cuatro patrones
sobre los 32 ficheros `.go`, cero coincidencias). **`research.md` ya está escrito** (2026-08-18): es la
transcripción de ese material a un artefacto versionado, con los patrones del barrido, el grafo de
imports, las tres réplicas literales y lo que se descartó.

> ⚠️ **Corregido el 2026-08-18.** Este apartado decía *«no hay preguntas abiertas de investigación»*, y
> **era falso: había dos, y ninguna estaba medida.** Se resolvieron midiendo, y las dos cambiaban una
> decisión del plan:
>
> 1. **¿La ubicación que propone D-005-P2 compila, o cierra un ciclo de importación?** No estaba
>    medido, y el repositorio **tiene** una arista incómoda documentada
>    (`internal/testutil/sandbox.go:11-20`). **Medido con `go list`: compila.** La arista
>    `transport → config` **ya existe en producción** y `config` es **hoja** del módulo —no importa
>    nada suyo, ni en producción ni en sus tres tests internos—. El ciclo de `testutil` es
>    **`testutil ↔ config`**, otro. Evidencia completa en D-005-P2.
> 2. **¿Son idénticas las tres réplicas de la guarda?** El plan afirmaba que eran *«la misma
>    comparación literal»*. **No lo son**: `enrollment.go` funde el fallo de análisis con el esquema en
>    un solo desenlace y nunca reproduce el argumento; `config.go` y `transport.go` los separan en dos
>    ramas con mensajes distintos. **Eso cambió la forma de la función unificada** —devuelve dos hechos,
>    no un booleano— y sin medirlo la extracción habría puesto en rojo uno de los cuatro tests de la
>    red. Detalle en D-005-P2.
>
> **La lección, que es de método**: «no hay preguntas abiertas» es una afirmación que también hay que
> medir. Escribirla sin comprobarla es la misma clase de defecto que el proyecto persigue en las
> puertas — un verde cuyo sujeto no se ha mirado.

### Phase 1 — Diseño y contratos

**`data-model.md` — ESCRITO.** Corto y con su motivo: esta feature **no persiste nada** (FR-019), así
que no hay modelo. Lo que documenta son los **tres valores en tránsito** —código, identidad,
denominación—, ninguno de los cuales sobrevive al proceso, y **cómo se verifica** que no hay estado
(SC-007, SC-010).

**`contracts/adhesion.md` — ESCRITO**, siguiendo la forma del precedente
(`../../003-enrolamiento/contracts/enrollment-string.md`). Cubre las cuatro cosas de D-005-P11: los
cuatro desenlaces con sus tres propiedades irrompibles, la forma de la petición, **que la ruta de
ingesta termina en su segmento conocido** y **cuál es el de la adhesión** —los dos que D-005-P3
incrusta en el código—. Declara qué es (interfaz pública colocada aquí) y **la ventana conocida**: la
plataforma todavía lo define por su lado, y hasta que la cierre **manda la implementación**.

**`contracts/cli.md` — ESCRITO.** Contrata **la superficie de línea de órdenes**, que 003 y 004 también
contratan y 005 necesita más que ninguna: estrena la gramática de dos niveles. Sigue la forma de
`../../003-enrolamiento/contracts/cli.md` —comando, entrada, comportamiento y salidas, garantías
DEBE/NUNCA— y **no** la de `004/contracts/cli-config.md`, que contrata una configuración y no un
comando: **mismo sujeto, misma forma**. Cubre la gramática nueva, las dos vías de entrada, el reparto
de canales, el mapeo desenlace → código de salida y el verbo desconocido. **Remite a `adhesion.md`
para el protocolo y no lo duplica.**

**`quickstart.md` — ESCRITO.** Forma de `../004-identidad-de-proyecto/quickstart.md`: prerrequisitos,
aislamiento obligatorio, validaciones numeradas y checklist de cierre. **Con la separación que esta
feature exige**: **V1–V9 automatizables** y **C1–C4 ceremonia manual contra plataforma real**, con la
prohibición de SC-006 escrita. Y **el banco TLS local como prerrequisito de primera clase** —proxy en
`:8443`, CA en `~/tls-local/ca.crt`, `SSL_CERT_FILE` aditivo— porque **FR-017 no tiene exención y sin
él la feature no se puede ejercer en local**.

### Inventario — **005 iguala a 004**

Con esto, 005 lleva **los mismos artefactos que 004** salvo los que no le corresponden:

| Artefacto | 004 | 005 | |
|---|:--:|:--:|---|
| `spec.md` · `plan.md` · `research.md` · `data-model.md` · `quickstart.md` · `contracts/` · `checklists/` | ✅ | ✅ | **igualado** |
| `tasks.md` | ✅ | — | Phase 2, la genera `/speckit.tasks` |
| `baseline-sc004.tsv` · `verdad-terreno.md` | ✅ | — | **propios de 004**: son su línea base de identidades y su verdad de terreno. 005 no tiene medición de campo equivalente, y crearlos por simetría sería inventar convención |

### Phase 2 — Tasks

**No la genera este comando.** **Cuatro** cosas que `/speckit.tasks` tiene que respetar, y se dejan escritas
aquí porque un troceo ingenuo las rompe:

1. **El golden test de frontera de la adhesión va PRIMERO**, antes de cualquier código — Principio IV,
   disciplina de primer commit. La segunda puerta de la frontera necesita el suyo.
2. **SC-001, SC-008 y SC-009 generan tarea propia** y no cuelgan de historias (D-005-P9).
3. **La unificación de la guarda (D-005-P2) es tarea temprana y aislada**, con los cuatro tests
   existentes como red, **antes** de escribir el método nuevo que la va a usar.
4. **La documentación es tarea de la feature, no un extra** (D-005-P12): la entrada del comando en el
   README **y la sección de comandos que hoy no existe**.

   > ⚠️ **Ajustado el 2026-08-18.** Esta regla decía también «y **el contrato público** (D-005-P11)».
   > **Ya está escrito** —`contracts/adhesion.md` y `contracts/cli.md`, Phase 1— así que **no genera
   > tarea**: generarla sería trabajo para algo hecho. Lo que sí queda de aquella frase es la
   > dependencia que señalaba, y sigue viva: **D-005-P3 cita los dos hechos de ruta del contrato**, y
   > esa cita hay que ponerla en el código cuando se escriba.

---

## Decisiones de plan

| # | Decisión | Alternativa descartada |
|---|---|---|
| **D-005-P1** | Método nuevo que lee el cuerpo; `Send` intacto | Ampliar `Send` con modo de salida — dos modos en el camino caliente |
| **D-005-P2** | **Unificar** la guarda de esquema en una función, cuatro llamantes | Replicar la condición una cuarta vez — FR-017 no admite que se olvide |
| **D-005-P3** | Derivar el destino **con validación ruidosa** de la forma esperada | Cirugía de cadena silenciosa (URL plausible a ninguna parte) · campo nuevo (exige `pmea3`) |
| **D-005-P4** | Seguir `Verify()`: síncrono, un intento, sin cola | Reutilizar `sendWithRetry` — encolar a quien está esperando |
| **D-005-P5** | Función nueva que expone «hubo raíz»; `Derivar` intacta y **cuerpo compartido** | Que `Derivar` devuelva error — rompería P-004 FR-010 en la ingesta |
| **D-005-P6** | Tercer `if` + despachador de segundo nivel en su fichero | `permea join` plano — cierra la puerta a la familia |
| **D-005-P7** | Dos capas inyectables, como `enroll` | Leer stdin en la capa de comando — intestable sin procesos |
| **D-005-P11** | **La interfaz pública vive en el repositorio público**: `contracts/adhesion.md` se escribe aquí | Ver las tres alternativas descartadas abajo |
| **D-005-P12** | **Documentación DENTRO**: entrada del comando en el README, con la sección de comandos que hoy no existe | Dejarlo fuera — un comando que nadie puede descubrir no existe |
| **D-005-P13** | Orden de los rehúses: **árbol → enrolamiento → configuración** | Cualquier otro — dejaría el genérico tapando al específico |
| **D-005-P14** | **Golden test de la adhesión en fichero propio**, en `internal/transport`, con remisión cruzada desde `boundary_test.go` | Ampliar `boundary_test.go` — vive en el paquete de ingesta, y la adhesión no es ingesta |

---

## D-005-P14 · Dónde vive el golden test de la adhesión

**Medido antes de decidir.** `internal/ingest/boundary_test.go`, **306 líneas**, contiene dos tests:

- **`TestBoundary_NoDenylistLeaks`** (`:61`) — *«el test que define el producto: ninguna información de
  la denylist puede sobrevivir al paso por la frontera»*.
- **`TestBoundary_TresCaminosHaciaElExterior`** (`:117`) — y su cabecera (`:104-116`) dice algo que
  importa aquí: *«FR-017, tras D-004-4, acota la garantía a la salida **hacia el exterior** y **la
  nombra entera**: el **evento serializado**, la **cola de envío** y el **transporte**»*. Los tres se
  comprueban **por separado, cada uno sobre SU artefacto**, y no derivando dos del primero.

**La decisión: fichero propio, en `internal/transport`. Con remisión cruzada obligatoria.**

**Tres razones para no ampliarlo:**

1. **`boundary_test.go` vive en `internal/ingest`, y la adhesión no es ingesta.** Su cuerpo lo compone
   el método nuevo de `internal/transport` y lo dispara un comando de `cmd/permea`. Meter la
   comprobación en el paquete de ingesta la pondría donde no está lo que comprueba.
2. **La propiedad que se verifica es distinta.** El de eventos comprueba que **ningún término de una
   denylist sobrevive** a la proyección de un fixture con contenido inyectado. El de la adhesión
   comprueba que el cuerpo contiene **exactamente dos campos y nada más** — es una allowlist de dos
   elementos, no un barrido de denylist. Compartir fichero mezclaría dos formas de comprobar.
3. **La adhesión tiene UN camino hacia el exterior, no tres.** No hay evento serializado —no emite
   eventos— ni cola —FR-018 lo prohíbe—. La estructura de «tres caminos» no le aplica.

**Y la razón para la remisión cruzada, que es lo que evita el defecto:** la cabecera de
`boundary_test.go:106-107` afirma que FR-017 **«la nombra entera»**. En cuanto exista una segunda
puerta con su propio testigo, **esa frase deja de ser cierta si nadie la actualiza**. Así que:

- **En `boundary_test.go`**: nota de que la frontera tiene ahora **una segunda puerta**, con su testigo
  propio y dónde está.
- **En el fichero nuevo**: nota de que es **el segundo golden test de la frontera**, hermano del de
  eventos, y qué comprueba cada uno.

**Sin las dos, la frontera queda comprobada en dos sitios que no se conocen** — y alguien ampliará uno
y olvidará el otro, que es exactamente la clase de defecto que este plan ya combate en D-005-P2.

> **Nota para `/speckit.tasks`**: la nota en `boundary_test.go` **toca un test existente y verde**. Es
> un comentario, no comportamiento, pero va en la misma tarea que crea el fichero nuevo — no suelta.

---

## D-005-P11 · La autoridad del contrato — **DECIDIDA: opción A′** (Basilio, 2026-08-18)

**La decisión: la interfaz pública vive en el repositorio público.** `contracts/adhesion.md` se escribe
**aquí**, en el repositorio del agente, **como interfaz pública de la plataforma**, y **la plataforma
pasa a citarlo en vez de definirlo**.

**La razón, y es de principio, no de conveniencia**: un cliente de código abierto **solo puede depender
de un contrato abierto**. El Principio II exige que *«un revisor externo DEBE poder confirmar el
cumplimiento de la frontera leyendo únicamente el código»*, y una remisión a una ruta privada lo
impide: el auditor se topa con un enlace que no puede abrir justo donde está lo que vino a verificar.
**La implementación de la plataforma sigue siendo privada; su interfaz no puede serlo.**

Es un paso más allá del precedente del `pmea2`
(`specs/003-enrolamiento/contracts/enrollment-string.md`), que se declara «fuente de verdad compartida
por dos repos»: A′ **no comparte la autoridad, la coloca**. Y coloca en el repositorio abierto
exactamente lo que un tercero necesita para escribir otro cliente.

### Qué tiene que documentar, y no es solo la tabla de desenlaces

| # | Contenido | Por qué |
|---|---|---|
| 1 | **Los cuatro desenlaces** de la adhesión y qué distingue a cada uno | Es el núcleo de la feature (FR-002, FR-010, FR-011, FR-012, FR-013) |
| 2 | **La forma de la petición** | Sin ella no hay cliente posible |
| 3 | **Que la ruta de ingesta termina en su segmento conocido** | **D-005-P3 lo incrusta en la validación ruidosa** |
| 4 | **Cuál es el segmento de la adhesión** | **D-005-P3 lo incrusta en la derivación del destino** |

**3 y 4 son la ampliación de alcance que el plan no había visto**: planteaba la autoridad como problema
documental, y resulta que **D-005-P3 ya mete dos hechos del contrato dentro del código del agente**. El
contrato tiene que documentarlos, y el código que los usa tiene que citarlo — o serían dos suposiciones
sobre un servidor ajeno sin nada que las respalde.

### Las tres alternativas descartadas

| | Alternativa | Por qué se descarta |
|---|---|---|
| **A** | Contrato aquí como «fuente de verdad compartida», al modo del `pmea2` | **Reclamaría autoridad sobre algo ya decidido y demostrado en la plataforma.** Si divergieran, el agente estaría equivocado y su contrato diría que no. A′ evita el choque: la interfaz se **coloca** aquí, y la plataforma la cita |
| **B** | Contrato aquí declarado **réplica verificada**, con fecha y commit | Una réplica **puede quedar obsoleta en silencio**, y **la puerta que la revalidaría no existe** en este repositorio. Sería deuda desde el primer día |
| **C** | **Solo remisión** por nombre lógico, sin transcribir | **Deja al auditor externo sin poder verificar nada** — que es justo lo que el Principio II prohíbe. Y esta feature **es** los desenlaces: remitirlos entera la vacía |

### El lado de la plataforma — **se anota, NO se hace**

A′ tiene una segunda mitad: la plataforma **deja de definir** el contrato y **pasa a citarlo**. Eso
**no se toca en esta feature ni en esta rama** — es trabajo de otro día, en el otro repositorio, y
queda anotado aquí para que exista y no se pierda. **Hasta que ocurra, el contrato estará definido en
dos sitios**, y esa ventana es conocida y aceptada.

---

## D-005-P12 · La documentación — **DENTRO el comando, FUERA la deuda vieja**

El plan no decía nada, y **el silencio se lee como olvido** en un proyecto que enumera lo excluido con
su motivo.

**DENTRO de esta feature:**

- **La entrada de `project join` en el README.** Un comando que nadie puede descubrir no existe.
- **Y con ella, la sección de comandos que hoy NO EXISTE.** Medido: el README documenta `--scan`,
  `--run`, `--daemon` y `--version`, y **no menciona `enroll` ni `status`**. No se puede añadir el
  tercer subcomando a una sección que no está, así que **crearla entra**: es el coste mínimo de
  documentar éste, no una mejora oportunista.

**FUERA, al backlog:**

- **`README.md:74` documenta el campo retirado «modo de ref»** —el `project_ref_mode` que P-004
  eliminó—, y **quien lo siga no arranca**: `CheckRetiredProjectRefMode` (`config.go:197-231`) detiene
  el arranque ante `plain`. Es un defecto **real y anterior**, y **no se cuela en esta feature**: se
  anota como deuda del repositorio.

**El criterio que separa las dos listas**: entra lo que esta feature **hace descubrible**; queda fuera
lo que estaba roto antes de ella. Arreglar de paso lo segundo mezclaría en una rama dos cosas que
alguien tendrá que revisar por separado.


---

## Complexity Tracking

**Vacía**: el Constitution Check no arrojó ninguna violación que justificar.
