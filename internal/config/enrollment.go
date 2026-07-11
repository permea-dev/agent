package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// enrollmentPrefix versiona el formato del enrollment string (contracts/enrollment-string.md).
// Un prefijo distinto o ausente es un enrollment string no reconocido.
const enrollmentPrefix = "pmea1."

// ErrEnrollmentString señala un enrollment string no reconocido o inválido. Su mensaje
// NUNCA incluye el argumento ni el token: el enrollment string es un secreto del mismo
// calibre que el salt (FR-007), es encoding y contiene el token en claro.
var ErrEnrollmentString = errors.New("enrollment string inválido")

// enrollmentPayload es la estructura cerrada que viaja dentro del enrollment string.
type enrollmentPayload struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

// ParseEnrollmentString decodifica un enrollment string `pmea1.<base64url(json)>` en su
// par (endpoint, token). Es encoding, no cifrado. La validez del token la decide el
// backend en el ping de verificación, no este formato; aquí solo se comprueba que el
// envoltorio está bien formado y que el endpoint es https. Ante cualquier fallo devuelve
// un error que NO reproduce el argumento ni el token (FR-007/FR-013).
func ParseEnrollmentString(s string) (endpoint, token string, err error) {
	body, ok := strings.CutPrefix(s, enrollmentPrefix)
	if !ok {
		return "", "", fmt.Errorf("%w: prefijo no reconocido", ErrEnrollmentString)
	}
	raw, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return "", "", fmt.Errorf("%w: base64 ilegible", ErrEnrollmentString)
	}
	var p enrollmentPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return "", "", fmt.Errorf("%w: json ilegible", ErrEnrollmentString)
	}
	u, err := url.Parse(p.Endpoint)
	if err != nil || u.Scheme != "https" {
		return "", "", fmt.Errorf("%w: el endpoint debe ser https", ErrEnrollmentString)
	}
	if p.Token == "" {
		return "", "", fmt.Errorf("%w: token ausente", ErrEnrollmentString)
	}
	return p.Endpoint, p.Token, nil
}

// IsEnrolled indica si la config representa un enrolamiento válido: endpoint y token
// presentes y endpoint https (reutiliza Validate). Es el estado que expone `permea status`.
func IsEnrolled(c Config) bool {
	return c.Endpoint != "" && c.DeviceToken != "" && c.Validate() == nil
}
