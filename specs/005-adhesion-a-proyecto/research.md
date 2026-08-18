# Research — Adhesión a proyecto (Phase 0)

**Feature**: `005-adhesion-a-proyecto` | **Fecha de las mediciones**: 2026-08-18 | **Contra**: `158fb1b`

> **Qué es este documento y qué NO es.** Aquí está **lo medido**; en [`plan.md`](./plan.md) está **lo
> decidido**. Cuando algo aparece en los dos, aquí va la medición y allí la consecuencia, con remisión.
>
> **Por qué existe.** Todas estas mediciones se hicieron en el descubrimiento del 2026-08-18 y vivían
> **solo en reportes de `tmp/`, que se sobrescriben en cada encargo**. La regla de la casa es que *lo
> que deba sobrevivir se transcribe en el encargo que lo observa*: esto es esa transcripción.

---

## R1 · Las cuatro trampas de mecanismo

Cuatro sitios donde el código existente **no sirve tal cual** para lo que la feature necesita. Las
cuatro medidas contra `158fb1b`, no recordadas.

### R1.1 · El transporte descarta el cuerpo de la respuesta

```go
// internal/transport/transport.go:129-142
defer func() { _ = resp.Body.Close() }() // solo lectura de la respuesta
// Vaciar el cuerpo permite reutilizar la conexión en los reintentos.
_, _ = io.Copy(io.Discard, resp.Body)

switch {
case resp.StatusCode >= 200 && resp.StatusCode < 300:
    return nil // aceptado (o ya visto por dedup): confirmar
case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
    return &sendError{status: resp.StatusCode, auth: true} // detener sync
case resp.StatusCode >= 500:
    return &sendError{status: resp.StatusCode, retryable: true} // reintentar
default:
    return &sendError{status: resp.StatusCode} // otros 4xx: registrar, no reintentar en bucle
}
```

**Dos hechos medidos, no uno:**

1. **El cuerpo se descarta** (`:131`). El nombre del Proyecto —el desenlace entero de la feature— no
   sobrevive a `Send`.
2. **La clasificación es solo por código, y los dos rechazos de la adhesión colapsan**: `422` y `409`
   caen ambos en el `default` de `:141`, indistinguibles para el llamante.

**El descarte es deliberado y está razonado en su comentario** (`:130`): vaciar el cuerpo permite
reutilizar la conexión en los reintentos, y el camino de ingesta drena lotes en serie. **No es un
descuido que corregir.** → Consecuencia: [`plan.md` D-005-P1](./plan.md).

### R1.2 · La guarda de esquema viaja con el método, no con el tipo

`transport.go:105-111`, **dentro de `Send`**. Un método nuevo del mismo tipo **nace sin ella**. Y hay
tres réplicas de la comprobación: ver **R4**, que las mide una a una.

### R1.3 · Lo guardado es la ruta completa de ingesta, no una base

| Lado | Cita | Qué hace |
|---|---|---|
| Plataforma | `backend/app/Http/Controllers/Api/V1/DeviceTokenController.php:34` | `$endpoint = rtrim(Config::string('app.url'), '/').'/api/v1/ingest';` — **compone la ruta completa** |
| Agente | `internal/transport/transport.go:117` | `http.NewRequest(http.MethodPost, c.Endpoint, …)` — **postea contra ella tal cual**, sin componer nada |

Corroborado por los propios tests del agente, que usan la forma completa:
`"https://inseguro.example/ingest"` (`internal/config/enrollment_test.go:97`),
`"http://insecure.example/ingest"` (`internal/transport/transport_test.go:154`).

→ Consecuencia: [`plan.md` D-005-P3](./plan.md).

### R1.4 · `Verify()` es el precedente exacto de la operación interactiva

```go
// internal/transport/transport.go:91-99
// Verify comprueba el device token con un ping de ingesta de **lote vacío** contra el
// mismo `/ingest`: reutiliza el contrato de transporte (Send), sin inventar ningún
// endpoint nuevo. […] Un solo intento (sin backoff): en el enrolamiento, un 5xx/red
// significa "no verificable → no persistir", no reintentar.
func (c *Client) Verify() error {
    return c.Send([]event.Event{})
}
```

**Y el hecho estructural que lo acompaña**: la cola **no está dentro del cliente, está encima**.
`Append` (`queue.go:23`) y `Drain` (`queue.go:73`) son funciones de paquete que **usan** el `Client`;
`Send` no las conoce. Quien llama a `Send` directamente **transmite y espera**, sin cola, sin que haya
que desactivar nada. → Consecuencia: [`plan.md` D-005-P4](./plan.md).

---

## R2 · El barrido de exenciones de TLS — **cero coincidencias**

**Qué acredita**: que FR-017 —transporte seguro sin exención, sin modo de desarrollo, sin variante—
**no tiene hoy ninguna puerta trasera** en el agente. **Sin los patrones escritos, el barrido no está
acreditado**: un «no encontré nada» cuyo patrón nadie puede reproducir no es una medición.

**Universo**: **32 ficheros `.go`** (15 de producción + 17 de test), excluyendo `.git/` y `dist/`.

| # | Patrón | Ámbito | Resultado |
|---|---|---|---|
| **A** | `grep -rniE 'https?\|scheme\|tls\|insecure\|skipverify\|allow\.?http\|no\.?verify' --include='*.go' .` *(excluyendo `_test.go` y `dist/`)* | producción | **28 coincidencias, TODAS son las tres guardas, sus mensajes, sus comentarios o usos de `net/http`.** Ni un `InsecureSkipVerify`, ni un `allow_http`, ni un `no_verify` |
| **B** | `grep -rnE 'os\.(Getenv\|LookupEnv)' --include='*.go' .` | todo el repo | **1 coincidencia, y es de test**: `internal/testutil/sandbox_test.go:21`. **Cero en producción**: el agente **no lee ninguna variable de entorno** para su comportamiento |
| **C** | `grep -rn '//go:build' --include='*.go' .` | todo el repo | **cero.** No hay build tags de ningún tipo |
| **D** | lectura del struct `Config` (`internal/config/config.go:32-42`) y de los flags de `main` (`cmd/permea/main.go:59-63`) | — | Campos: `endpoint`, `device_token`, `org_id`, `dev_id`, `tools`, `sync_interval`, `logs_root`. Flags: `--scan`, `--run`, `--daemon`, `--version`. **Ninguno es una exención** |

**Veredicto: barrido A+B+C+D sobre los 32 ficheros, cero vías de exención.** No existe forma —de
configuración, de entorno, de compilación ni de bandera— de hacer que el agente hable por `http://`.

> **Y hay una decisión de diseño detrás, no un olvido.** `internal/config/config.go:14-29` documenta
> que P-004 **retiró** un modo de configuración (`project_ref_mode: plain`) precisamente por ser *«una
> promesa inerte en la configuración de un binario de código abierto, a la vista de cualquier
> auditor»*. Añadir una exención de TLS a la config sería **exactamente la clase de cosa que este
> repositorio acaba de quitarse de encima** — y en Principio I, no en Principio III.

---

## R3 · El grafo de imports — **la ubicación de D-005-P2 compila**

**Por qué se midió**: la decisión de unificar la guarda coloca la función en `internal/config` y hace
que `internal/transport` la llame, y **eso nadie lo había comprobado**. El repositorio **tiene** una
arista prohibida documentada (`internal/testutil/sandbox.go:11-20`), así que había que descartar que
fuera ésta.

**Las tres consultas, literales:**

```
$ go list -deps ./internal/transport | grep permea
github.com/permea-dev/agent/internal/config
github.com/permea-dev/agent/internal/event
github.com/permea-dev/agent/internal/transport

$ go list -deps ./internal/config | grep permea
github.com/permea-dev/agent/internal/config

$ go list -f '{{.ImportPath}} TEST-> {{join .TestImports " "}} XTEST-> {{join .XTestImports " "}}' ./internal/...
…/internal/config    TEST-> encoding/base64 encoding/json os path/filepath strings testing   XTEST->
…/internal/transport TEST-> encoding/json errors fmt …/internal/event io net/http net/http/httptest path/filepath sync testing time   XTEST->
…/internal/ingest    TEST-> bufio bytes encoding/json …/internal/event …/internal/testutil …/internal/transport io net/http net/http/httptest os strings testing   XTEST->
…/internal/project   TEST-> …/internal/testutil os os/exec path/filepath regexp testing   XTEST->
```

**Tres hechos:**

1. **`transport → config` ya existe, hoy, en producción.** Llamar a la función desde `transport` **no
   añade ninguna arista**.
2. **`config` es HOJA del módulo**: `go list -deps` devuelve solo a sí mismo, y sus imports son
   **exclusivamente librería estándar**. **No puede cerrar ciclo con nadie.**
3. **Sus tres ficheros de test son internos** (`package config`, verificado en `config_test.go`,
   `enrollment_test.go`, `identity_test.go`) **pero solo usan la librería estándar**, y `XTestImports`
   está vacío.

### El ciclo documentado es OTRO: `testutil ↔ config`

```
// internal/testutil/sandbox.go:11-20
// ═══ POR QUÉ ESTE PAQUETE NO IMPORTA `internal/config` ═══
// Sería lo natural […] y es exactamente lo que NO se puede hacer:
// `internal/config/config_test.go` es un test INTERNO (`package config`) […]
// Si este paquete importara `config`, esos tests no podrían usar el helper: ciclo de importación.
```

La arista prohibida es **`testutil ↔ config`**, y sigue prohibida. **No es `transport ↔ config`**, que
ya está tendida y en la dirección buena. → Consecuencia: [`plan.md` D-005-P2](./plan.md).

---

## R4 · Las tres réplicas de la guarda — **NO son idénticas**

Medidas literalmente. La creencia de que eran «la misma comparación» era falsa, y la diferencia es
justo la que puede romper una extracción ingenua.

```go
// internal/config/enrollment.go:78-80
u, err := url.Parse(p.Endpoint)
if err != nil || u.Scheme != "https" {                                   // ← PLEGADAS EN UNA
    return "", "", "", fmt.Errorf("%w: el endpoint debe ser https", ErrEnrollmentString)
}

// internal/config/config.go:100-106
u, err := url.Parse(c.Endpoint)
if err != nil {
    return fmt.Errorf("endpoint inválido %q: %w", c.Endpoint, err)       // ← RAMA 1
}
if u.Scheme != "https" {
    return fmt.Errorf("endpoint debe ser https://, got %q", c.Endpoint)  // ← RAMA 2
}

// internal/transport/transport.go:105-111
u, err := url.Parse(c.Endpoint)
if err != nil {
    return fmt.Errorf("transport: endpoint inválido %q: %w", c.Endpoint, err)  // ← RAMA 1
}
if u.Scheme != "https" {
    return fmt.Errorf("%w: %q", ErrScheme, c.Endpoint)                          // ← RAMA 2
}
```

### Las tres diferencias

| # | Diferencia | Dónde | Por qué existe |
|---|---|---|---|
| **1** | `enrollment.go` **funde** el fallo de análisis y el esquema equivocado en **un solo desenlace**; los otros dos los **separan en dos ramas** | `:78-80` vs `:100-106` y `:105-111` | — |
| **2** | `enrollment.go` **NUNCA reproduce el endpoint**; los otros dos lo interpolan con `%q` | ídem | **Deliberado**: `enrollment.go:30-33` razona que el `pmea2` **contiene el token en claro**, así que su error no puede llevar el argumento |
| **3** | `transport.go` envuelve el centinela **`ErrScheme`** (`:31`); los otros dos no tienen centinela | `:110` | Tests dependen de él por `errors.Is` (`transport_test.go:156`) |

### Los cuatro tests que las vigilan — **la red de la unificación**

| Guarda | Caso | Cita |
|---|---|---|
| `transport.go` | `TestSend_RejectsHTTP`, exige `errors.Is(err, ErrScheme)` | `internal/transport/transport_test.go:151-158` |
| `config.go` | endpoint `http://` debe fallar `Validate()` | `internal/config/config_test.go:61-64` |
| `enrollment.go` | caso de tabla «endpoint http (no https)» | `internal/config/enrollment_test.go:97` |
| `enrollment.go` | `T011(c)` — aborta **antes del ping** y **sin filtrar el token** | `cmd/permea/enroll_reject_test.go:129-160` |

→ Consecuencia: [`plan.md` D-005-P2](./plan.md), donde la unificación pasa a devolver **dos hechos** en
vez de un booleano, precisamente para preservar estas tres diferencias.

---

## R5 · `--scan` usa un salt literal — **sus refs no son comparables con producción**

```go
// cmd/permea/main.go:342
ctx := ingest.Context{Salt: "dry-run-salt", MachineID: "local", DevID: "dev-local", OrgID: "org-local", AgentVersion: version}
```

**`--scan` es local puro** —comprobado: `dryRun` (`main.go:335-368`) no llama a `transport.*`, ni a
`state.*`, ni a `setup()`, así que no toca cola, ni estado, ni el directorio de datos—. **Pero usa un
salt literal**, no el real de la instalación.

**Consecuencia práctica, y por eso se anota**: los `project_ref` que `--scan` imprime **están en otro
espacio de valores** que los de producción. **No se pueden comparar con los de la base de datos ni con
los de la línea base.** Además se truncan a 8 caracteres al imprimirlos (`:357-359`).

**Lo que sí sirve de `--scan` son los RECUENTOS.** Se usó así el 2026-08-18 para identificar el origen
de las identidades sin asignar, y funcionó: los recuentos por directorio cuadraron al evento con los de
la base.

> **Por qué está aquí y no en el plan**: no cambia ninguna decisión de esta feature, pero **es una
> trampa que ya costó una confusión** y no estaba escrita en ningún artefacto versionado. Quien vaya a
> verificar SC-001 comparando identidades **no debe usar `--scan` para obtenerlas**.

---

## R6 · Lo que se descartó, y por qué

Las alternativas evaluadas y rechazadas. Se conservan porque saber que algo se consideró y se descartó
—y por qué— es parte del expediente: sin esto, la primera persona que lea el plan volverá a proponerlas.

| # | Alternativa descartada | Por qué |
|---|---|---|
| **1** | **Ampliar `Send` para que devuelva el cuerpo** (parámetro de salida o segundo valor de retorno) | Convertiría el método del **camino caliente** en uno con dos modos, y el modo equivocado en la ingesta sería un **error silencioso**. Además el descarte del cuerpo está razonado en `transport.go:130`: reutilizar la conexión en los reintentos |
| **2** | **Añadir un campo nuevo al `pmea2`** (una URL base, en vez de derivar el destino) | El payload es **struct cerrada** con `additionalProperties: false` (`internal/config/enrollment.go:39-45`, `:71-77`; contrato en `specs/003-enrolamiento/contracts/enrollment-string.md`). Un campo más **no es una ampliación: es un formato nuevo**, `pmea3`, coordinado entre dos repositorios — y esta feature no lo necesita |
| **3** | **`permea join` plano**, sin sustantivo | Cerraría la puerta a la familia de verbos que la spec excluye hoy pero nombra (`leave`, `status` de proyecto). Y la plataforma llama a esto «Proyecto»: el sustantivo alinea las dos gramáticas |
| **4** | **Que `Derivar` devuelva error cuando no encuentre raíz** | Rompería **P-004 FR-010 en el camino de la ingesta** (`specs/004-identidad-de-proyecto/spec.md:249`), que es donde la garantía existe: un lote entero se detendría porque un directorio dejó de existir. El dato se expone por **vía adicional**, que es lo que FR-015 exige |
| **5** | **Cirugía de cadena silenciosa** sobre el endpoint guardado | Produciría **una URL plausible que apunta a ninguna parte**, y el desenlace sería un error de red que manda a depurar la conexión en vez de la configuración. De ahí la **validación ruidosa** de D-005-P3 |
| **6** | **Replicar la guarda de esquema una cuarta vez** | Ver R4 y `plan.md` D-005-P2: FR-017 no admite que se olvide, y el repositorio ya pagó esta clase de defecto —una condición replicada, corregida en un sitio y olvidada en cuatro— |

---

## Preguntas abiertas

**Ninguna de investigación.** Las dos que quedaban —si la ubicación de D-005-P2 compila (**R3**) y si
las tres réplicas eran idénticas (**R4**)— se resolvieron midiendo, y las dos **cambiaron una decisión
del plan**.

La única decisión pendiente no es de investigación sino de producto, y está **tomada**: la autoridad
del contrato de adhesión, resuelta como **opción A′** en [`plan.md` D-005-P11](./plan.md).
