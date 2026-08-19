// Package transport envía los eventos de frontera al backend de equipo por HTTPS
// autenticado (FR-008/FR-009) y sostiene la entrega exactamente-una-vez: at-least-once
// desde el cliente (reintentos con backoff acotado) + deduplicación por event_id en el
// backend. La cola offline y su reescritura atómica viven en queue.go.
package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/event"
)

// Valores por defecto del backoff acotado (T029): reintentar 5xx/red hasta 5 veces,
// con una espera que se dobla hasta un tope de 5 minutos.
const (
	defaultMaxRetries = 5
	defaultMaxBackoff = 5 * time.Minute
	defaultBatchSize  = 256
	baseBackoff       = 1 * time.Second
)

// ErrScheme señala un endpoint no-https: error de configuración que aborta el envío
// antes de transmitir nada en claro (FR-009).
var ErrScheme = errors.New("transport: el endpoint debe usar https://")

// ═══ P-005 · LOS DESENLACES DE LA ADHESIÓN ════════════════════════════════════════════════
//
// `contracts/adhesion.md` §Los cuatro desenlaces. Los devuelve `Adherir`, **distinguiendo por el
// ESTADO** de la respuesta; el cuerpo confirma, nunca decide.
//
// *(Se declararon una fase antes de cablearse, para que los tests de desenlace pudieran **compilar y
// nacer rojos**: un `[build failed]` no es un rojo legible. El centinela de andamiaje que `Adherir`
// devolvía entretanto lo retiró T012 junto con esta nota.)*

// ErrCodigoNoUtilizable es el desenlace 1 (`422`): inexistente · de otra organización · revocado ·
// prefijo desconocido · `project_ref` no conforme. **Las cinco causas son indistinguibles por
// construcción**, y el cliente NUNCA debe intentar deducir cuál fue.
var ErrCodigoNoUtilizable = errors.New("transport: el código de adhesión no es utilizable")

// ErrIdentidadYaAsignada es el desenlace 2 (`409`): la identidad ya está en OTRO Proyecto. **NUNCA
// nombra el Proyecto ajeno**: la plataforma no lo revela y el cliente no puede inventarlo.
var ErrIdentidadYaAsignada = errors.New("transport: esta identidad ya pertenece a otro proyecto")

// ErrNoVerificable es P-005 FR-013: el desenlace **no se pudo establecer**. Cubre el servidor
// inalcanzable y —lo que es menos evidente— **un `200` cuyo cuerpo no permite leer la denominación**:
// *«un éxito cuyo cuerpo no se pueda interpretar no es un éxito»* (`contracts/adhesion.md`). El estado
// remoto queda indeterminado, y el cliente **NUNCA afirma ningún desenlace**.
var ErrNoVerificable = errors.New("transport: no se pudo verificar el desenlace de la adhesión")

// sendError clasifica el fallo de un envío para decidir la acción del agente según
// contracts/transport.md: reintentar (5xx/red), detener el sync (401/403 auth), o
// registrar sin reintentar en bucle (otros 4xx).
type sendError struct {
	status    int   // código HTTP; 0 cuando el fallo es de red (sin respuesta)
	retryable bool  // 5xx o error de red -> reintentar con backoff
	auth      bool  // 401/403 -> token inválido, detener sync
	cause     error // causa subyacente (error de red), si la hay
}

func (e *sendError) Error() string {
	if e.status == 0 {
		return fmt.Sprintf("transport: error de red: %v", e.cause)
	}
	return fmt.Sprintf("transport: respuesta HTTP %d", e.status)
}

func (e *sendError) Unwrap() error { return e.cause }

// Retryable indica si el error de Send justifica reintento (5xx / error de red). Un
// lote con error reintentable permanece en la cola para el siguiente ciclo de sync.
func Retryable(err error) bool {
	var se *sendError
	return errors.As(err, &se) && se.retryable
}

// IsAuth indica si el error es de autenticación (401/403): el sync debe detenerse por
// configuración errónea, sin reintentar en bucle.
func IsAuth(err error) bool {
	var se *sendError
	return errors.As(err, &se) && se.auth
}

// Client habla con el backend de equipo por HTTPS.
type Client struct {
	Endpoint    string
	DeviceToken string
	HTTP        *http.Client
	MaxRetries  int
	MaxBackoff  time.Duration
	BatchSize   int
	// sleep espera entre reintentos; inyectable en tests para no dormir de verdad.
	sleep func(time.Duration)
}

// New crea un cliente con timeout razonable y el backoff/batching por defecto.
func New(endpoint, token string) *Client {
	return &Client{
		Endpoint:    endpoint,
		DeviceToken: token,
		HTTP:        &http.Client{Timeout: 10 * time.Second},
		MaxRetries:  defaultMaxRetries,
		MaxBackoff:  defaultMaxBackoff,
		BatchSize:   defaultBatchSize,
		sleep:       time.Sleep,
	}
}

// Verify comprueba el device token con un ping de ingesta de **lote vacío** contra el
// mismo `/ingest`: reutiliza el contrato de transporte (Send), sin inventar ningún
// endpoint nuevo. El slice es **no-nil** (`[]event.Event{}`) para que el cuerpo sea `[]`
// y no `null` —un `null` no es un lote válido—. Un solo intento (sin backoff): en el
// enrolamiento, un 5xx/red significa "no verificable → no persistir", no reintentar.
// No cruza la frontera: cero eventos, cero metadato (Principio I).
func (c *Client) Verify() error {
	return c.Send([]event.Event{})
}

// Send transmite un lote de eventos por HTTPS autenticado e interpreta el código de
// estado según el contrato (2xx=aceptado, 401/403=auth, 5xx=reintentar, otros 4xx=error).
// La deduplicación extremo a extremo se apoya en event_id.
func (c *Client) Send(events []event.Event) error {
	// ═══ AQUÍ ESTABA LA GUARDA DE ESQUEMA, Y YA NO ESTÁ (P-005 T006 §encontrabilidad) ═══════
	//
	// **El juicio vive ahora en `internal/config.JuzgarEndpoint`** (`internal/config/endpoint.go`), y
	// allí está escrita la lista de sus CUATRO llamantes: `ParseEnrollmentString`, `Config.Validate`,
	// este `Send` y `Adherir`.
	//
	// **Por qué se movió**: la condición estaba escrita **cuatro veces**. Cuatro copias de la frontera
	// son cuatro sitios donde alguien arregla tres y olvida una, y el fallo no es cosmético — es
	// **emitir por canal en claro**, Principio I. Es el defecto de clase que la plataforma ya pagó.
	//
	// **Y esta nota existe porque unificar tiene un coste**: mejora el código y **empeora el hallazgo**
	// —quien buscara la frontera la encontraba justo aquí—. Sin este puntero, D-005-P2 habría cambiado
	// una deuda por otra.
	//
	// **Lo que NO se movió es el desenlace, y es deliberado**: esta puerta CONSERVA la causa con `%w` y
	// CONSERVA el centinela `ErrScheme` —de los dos dependen tests por `errors.Is`—, mientras que
	// `ParseEnrollmentString` los descarta a propósito porque su argumento lleva el token dentro.
	// **Se unificó el juicio, no la presentación.**
	errAnalisis, admisible := config.JuzgarEndpoint(c.Endpoint)
	if errAnalisis != nil {
		return fmt.Errorf("transport: endpoint inválido %q: %w", c.Endpoint, errAnalisis)
	}
	if !admisible {
		return fmt.Errorf("%w: %q", ErrScheme, c.Endpoint)
	}

	body, err := json.Marshal(events)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.DeviceToken)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		// Sin respuesta del backend/enlace: reintentable con backoff.
		return &sendError{status: 0, retryable: true, cause: err}
	}
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
}

// sendWithRetry envía con backoff exponencial acotado (T029): reintenta solo los
// errores reintentables (5xx / red) hasta MaxRetries, doblando la espera hasta MaxBackoff.
// Los errores de auth y de petición malformada no se reintentan. Agotados los reintentos,
// devuelve el último error reintentable y el lote queda en cola para el próximo ciclo.
func (c *Client) sendWithRetry(events []event.Event) error {
	delay := baseBackoff
	var err error
	for attempt := 0; ; attempt++ {
		err = c.Send(events)
		if err == nil || !Retryable(err) {
			return err
		}
		if attempt >= c.MaxRetries {
			return err // reintentos agotados: el lote permanece en cola
		}
		c.sleep(delay)
		delay *= 2
		if delay > c.MaxBackoff {
			delay = c.MaxBackoff
		}
	}
}

// ═══ P-005 T002 · PUNTO DE EXTENSIÓN DE LA ADHESIÓN ════════════════════════════════════════

// peticionAdhesion es el cuerpo de la adhesión: **exactamente dos campos y nada más**.
//
// Es la allowlist de la SEGUNDA PUERTA de la frontera de datos (Principio I), y su forma la fija
// `specs/005-adhesion-a-proyecto/contracts/adhesion.md` §La petición. Struct **cerrada por
// construcción**: lo que no está aquí no puede viajar, y añadir un campo es ampliar la frontera.
type peticionAdhesion struct {
	Code       string `json:"code"`
	ProjectRef string `json:"project_ref"`
}

// Adherir presenta un código de adhesión y devuelve la denominación del Proyecto al que la
// instalación queda unida.
//
// ═══ HISTORIA — nació como PUNTO DE EXTENSIÓN en P-005 T002 ════════════════════════════════
//
// Durante Phase 1–3 **componía y emitía la petición con su cuerpo definitivo** pero devolvía siempre
// un centinela de andamiaje, ya retirado. Componer y emitir de verdad **no fue adorno**: sin cuerpo
// real, el golden test de frontera de la adhesión (P-005 T007) **no habría tenido sujeto que
// observar** —nada sobre lo que comprobar que solo viajan dos campos—.
//
// **P-005 T012 cerró la interpretación de los desenlaces** —leer el cuerpo, distinguir por estado— y
// retiró el centinela con ella.
//
// ═══ LO QUE NO CAMBIÓ AL IMPLEMENTARSE, Y ERA EL PUNTO ═════════════════════════════════════
//
//	· la FIRMA — T012 no obligó a tocar a ningún llamante;
//	· el CUERPO de la petición — dos campos, `contracts/adhesion.md` §La petición;
//	· UN SOLO INTENTO, SIN COLA (P-005 FR-018) — se emite aquí y se espera, sin pasar por
//	  `sendWithRetry` ni por `Append`/`Drain`. La cola no está DENTRO del cliente, está ENCIMA:
//	  quien llama a este método directamente transmite y espera, y eso no requiere desactivar nada.
//	  Es el mismo criterio que `Client.Verify` ya razonó para el enrolamiento.
//
// ⚠️ **`Send` no se toca.** Su descarte del cuerpo está razonado —reutiliza la conexión en los
// reintentos del camino caliente— y esta operación **no puede** reutilizarlo: `Send` serializa
// `[]event.Event`, y aquí el cuerpo es otro. Por eso `Adherir` lee el suyo y `Send` lo descarta.
func (c *Client) Adherir(codigo, projectRef string) (denominacion string, err error) {
	// El juicio de esquema vive en `internal/config.JuzgarEndpoint` — ver la nota larga en `Send`, y la
	// lista de llamantes en `internal/config/endpoint.go`. Aquí estuvo la réplica inline de andamiaje
	// de P-005 T002, que T005 retiró.
	//
	// La segunda puerta de la frontera usa EL MISMO juicio que la ingesta y da EL MISMO desenlace:
	// centinela `ErrScheme` para el esquema, causa conservada con %w para el parseo (D-005-P2). Unificar la condición y divergir en el desenlace sería unificar el
	// código y mantener la diferencia justo en lo que la persona lee cuando su config está rota.
	errAnalisis, admisible := config.JuzgarEndpoint(c.Endpoint)
	if errAnalisis != nil {
		return "", fmt.Errorf("transport: endpoint inválido %q: %w", c.Endpoint, errAnalisis)
	}
	if !admisible {
		return "", fmt.Errorf("%w: %q", ErrScheme, c.Endpoint)
	}

	body, err := json.Marshal(peticionAdhesion{Code: codigo, ProjectRef: projectRef})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.DeviceToken)

	// UN SOLO INTENTO, SIN COLA (D-005-P4), igual que `Verify()`: encolar la petición de alguien que
	// está mirando la pantalla sería mentirle, y reintentar no aporta nada —el código no se agota y
	// unirse dos veces es indistinguible de unirse una (FR-013a)—.
	resp, err := c.HTTP.Do(req)
	if err != nil {
		// Servidor inalcanzable: el desenlace remoto queda INDETERMINADO. No se afirma ninguno.
		return "", fmt.Errorf("%w: %v", ErrNoVerificable, err)
	}
	defer func() { _ = resp.Body.Close() }() // solo lectura de la respuesta

	// Se lee siempre —también en los rechazos— para drenar la conexión y poder reutilizarla. El tope
	// es defensivo: un cuerpo truncado no decodifica, y eso cae en «no verificable», que es el lado
	// seguro. Un nombre de Proyecto no se acerca de lejos a este tamaño.
	cuerpo, err := io.ReadAll(io.LimitReader(resp.Body, maxCuerpoAdhesion))
	if err != nil {
		return "", fmt.Errorf("%w: no se pudo leer la respuesta: %v", ErrNoVerificable, err)
	}

	// ═══ EL DISCRIMINANTE ES EL ESTADO. EL CUERPO CONFIRMA, NUNCA DECIDE ═══════════════════
	//
	// `contracts/adhesion.md` §Qué distingue a qué: el estado es «el discriminante barato y el que no
	// depende de interpretar el cuerpo». Un `422` que llegue con el cuerpo del `409` sigue siendo el
	// desenlace 1.
	switch resp.StatusCode {
	case http.StatusOK:
		return denominacionDe(cuerpo)

	case http.StatusUnprocessableEntity: // desenlace 1
		// Las CINCO causas son indistinguibles por construcción, y el cliente NUNCA debe intentar
		// deducir cuál fue: lo convertiría en un oráculo para averiguar qué códigos existen.
		return "", ErrCodigoNoUtilizable

	case http.StatusConflict: // desenlace 2
		// NUNCA nombra el Proyecto ajeno: la plataforma no lo revela y el cliente no lo inventa.
		return "", ErrIdentidadYaAsignada

	default:
		// ⛔ PRINCIPIO I — un desenlace que NO SE RECONOCE no puede tratarse como conforme.
		//
		// El contrato enumera tres estados y su cláusula general manda sobre el resto: «una respuesta
		// que no permite determinar el desenlace» es **no verificable** (FR-013). Clasificarlo como
		// rechazo también sería afirmar un desenlace, así que tampoco.
		//
		// ⚠️ Esto NO es el `default` que sobra de un `switch`: es el punto exacto donde un cliente
		// descuidado decide que un `500` fue un rechazo, o que un `204` fue una unión. Lo cubre
		// `TestAdherir_EstadoNoContempladoEsNoVerificable`.
		return "", fmt.Errorf("%w: estado %d no contemplado por el contrato", ErrNoVerificable, resp.StatusCode)
	}
}

// maxCuerpoAdhesion acota la lectura de la respuesta. Ver el motivo en `Adherir`.
const maxCuerpoAdhesion = 1 << 20 // 1 MiB

// denominacionDe extrae la denominación del Proyecto de un cuerpo de éxito, o devuelve
// `ErrNoVerificable`.
//
// ═══ UN ÉXITO CUYO CUERPO NO SE PUEDE INTERPRETAR NO ES UN ÉXITO ══════════════════════════
//
// El estado `200` **no basta**: P-005 FR-002 exige comunicar la denominación, así que un `200` sin
// `project.name` legible se trata como **no verificable** (P-005 FR-013), no como unión conseguida.
//
// **Las dos comprobaciones hacen falta, y cubren cosas distintas:**
//   - el error de `Unmarshal` caza los **tipos equivocados** —`"name":42`, `"name":{…}`,
//     `"project":"…"`— y el cuerpo que ni siquiera es JSON de objeto;
//   - la comprobación de vacío caza lo que **decodifica sin protestar**: clave ausente, `null`,
//     objeto `project` ausente y cadena vacía **producen los cuatro el mismo cero silencioso**.
//
// Sin la segunda, esos cuatro devolverían `""` **como si fuera una denominación**. Es el desenlace que
// `TestAdherir_DoscientosSinNombreLegibleEsNoVerificable` enumera en trece formas.
func denominacionDe(cuerpo []byte) (string, error) {
	var respuesta struct {
		Project struct {
			Name string `json:"name"`
		} `json:"project"`
	}
	if err := json.Unmarshal(cuerpo, &respuesta); err != nil {
		return "", fmt.Errorf("%w: la respuesta de éxito no se pudo interpretar: %v", ErrNoVerificable, err)
	}
	nombre := strings.TrimSpace(respuesta.Project.Name)
	if nombre == "" {
		return "", fmt.Errorf("%w: la respuesta de éxito no trae denominación legible", ErrNoVerificable)
	}
	return nombre, nil
}
