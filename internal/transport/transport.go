// Package transport envía eventos al backend por HTTPS (FR-008).
// Stub inicial: el batching, la cola offline y el backoff se detallan en su spec.
package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bfgnet/agente_permea/internal/event"
)

// Client habla con el backend de equipo.
type Client struct {
	Endpoint    string
	DeviceToken string
	HTTP        *http.Client
}

// New crea un cliente con timeout razonable.
func New(endpoint, token string) *Client {
	return &Client{Endpoint: endpoint, DeviceToken: token, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

// Send transmite un lote de eventos. La deduplicación en backend usa event_id.
func (c *Client) Send(events []event.Event) error {
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
		return err
	}
	defer func() { _ = resp.Body.Close() }() // solo lectura de la respuesta
	return nil
}
