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
	"net/url"
	"time"

	"github.com/bfgnet/agente_permea/internal/event"
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

// Send transmite un lote de eventos por HTTPS autenticado e interpreta el código de
// estado según el contrato (2xx=aceptado, 401/403=auth, 5xx=reintentar, otros 4xx=error).
// La deduplicación extremo a extremo se apoya en event_id.
func (c *Client) Send(events []event.Event) error {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return fmt.Errorf("transport: endpoint inválido %q: %w", c.Endpoint, err)
	}
	if u.Scheme != "https" {
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
