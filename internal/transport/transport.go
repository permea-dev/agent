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

// ErrAdhesionNoImplementada es el centinela del PUNTO DE EXTENSIÓN de la adhesión (P-005 T002).
//
// ═══ POR QUÉ ES UN CENTINELA PROPIO Y NO «NO VERIFICABLE» ══════════════════════════════════
//
// Es andamiaje, y desaparece cuando P-005 T012 implemente los desenlaces. Existe para que los tests
// de desenlace **nazcan en rojo por comportamiento** y no por «símbolo no existe» —un rojo de
// compilación no dice nada—, y **DEBE ser distinto del desenlace de "no verificable"**: si fueran el
// mismo valor, el test que espera «no verificable» ante una respuesta ininterpretable (T010)
// **nacería VERDE acertando contra el andamiaje**, no contra el comportamiento, y su rojo no
// existiría.
var ErrAdhesionNoImplementada = errors.New("transport: adhesión no implementada")

// ═══ P-005 · LOS TRES DESENLACES DE LA ADHESIÓN — DECLARADOS, TODAVÍA NO CABLEADOS ════════
//
// `contracts/adhesion.md` §Los cuatro desenlaces. **Existen desde Phase 4 y los cablea T012**: hasta
// entonces `Adherir` devuelve `ErrAdhesionNoImplementada` para todo, y **ninguno de estos tres se
// devuelve jamás**.
//
// **Por qué se declaran antes de usarse, que parece al revés.** Los tests de Phase 4 —T008, T009,
// T010— tienen que **compilar para poder nacer rojos**, y un `[build failed]` **no es un rojo
// legible**: falla igual con un test correcto y con uno vacío (disciplina 3). Es el mismo criterio que
// aplicó T002 con `ErrAdhesionNoImplementada`: **el andamiaje declara lo que los tests necesitan y
// devuelve algo que ninguno espera.**
//
// ⚠️ **T010 depende de esto para existir.** Sin `ErrNoVerificable`, su única aserción posible sería
// «hay error y no hay denominación» — **y eso ya lo cumple el andamiaje**, así que **nacería VERDE**
// contra el stub. Es el rojo más frágil del fichero, y esta declaración es lo que lo hace posible.

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
// ═══ ESTADO: PUNTO DE EXTENSIÓN (P-005 T002). NO INTERPRETA DESENLACES TODAVÍA ═════════════
//
// Hoy **compone y emite la petición con su cuerpo definitivo** y devuelve siempre
// `ErrAdhesionNoImplementada`. Componer y emitir de verdad **no es adorno**: sin cuerpo real, el
// golden test de frontera de la adhesión (P-005 T007) **no tendría sujeto que observar** —no habría
// nada sobre lo que comprobar que solo viajan dos campos—.
//
// La interpretación de los cuatro desenlaces —leer el cuerpo, distinguir por estado— llega en
// **P-005 T012**, y con ella desaparece el centinela.
//
// ═══ LO QUE YA ES DEFINITIVO ═══════════════════════════════════════════════════════════════
//
//	· la FIRMA — para que T012 no obligue a tocar a sus llamantes;
//	· el CUERPO de la petición — dos campos, `contracts/adhesion.md` §La petición;
//	· UN SOLO INTENTO, SIN COLA (P-005 FR-018) — se emite aquí y se espera, sin pasar por
//	  `sendWithRetry` ni por `Append`/`Drain`. La cola no está DENTRO del cliente, está ENCIMA:
//	  quien llama a este método directamente transmite y espera, y eso no requiere desactivar nada.
//	  Es el mismo criterio que `Verify` (`:91-99`) ya razonó para el enrolamiento.
//
// ⚠️ **`Send` no se toca.** Su descarte del cuerpo (`:131`) está razonado —reutiliza la conexión en
// los reintentos del camino caliente— y esta operación **no puede** reutilizarlo: `Send` serializa
// `[]event.Event`, y aquí el cuerpo es otro.
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

	resp, err := c.HTTP.Do(req)
	if err != nil {
		// Sin respuesta del enlace. El desenlace real —«no verificable», P-005 FR-013— lo decide
		// T012; aquí el andamiaje responde lo mismo que en cualquier otro caso, a propósito.
		return "", ErrAdhesionNoImplementada
	}
	defer func() { _ = resp.Body.Close() }() // solo lectura de la respuesta
	_, _ = io.Copy(io.Discard, resp.Body)

	return "", ErrAdhesionNoImplementada
}
