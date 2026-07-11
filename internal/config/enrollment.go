package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// enrollmentPrefix versiona el formato del enrollment string (contracts/enrollment-string.md).
// Solo pmea2. es válido; un prefijo distinto o ausente es un enrollment string no reconocido.
const enrollmentPrefix = "pmea2."

// enrollmentPrefixV1 es el formato viejo {endpoint, token} (contracts/enrollment-string.md):
// queda RETIRADO. Se detecta explícitamente para dar un error de "formato obsoleto" accionable
// en vez de un genérico "no reconocido": el pmea1 no trae dev_id autoritativo y es insuficiente.
const enrollmentPrefixV1 = "pmea1."

// devIDPattern es el charset autoritativo del dev_id (contracts/enrollment-string.md): 1–64
// caracteres de `[A-Za-z0-9._-]`. El límite y el charset lo hacen seguro de usar en rutas,
// cabeceras y nombres de fichero sin escaping.
var devIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

const devIDMaxLen = 64

// ErrEnrollmentString señala un enrollment string no reconocido o inválido. Su mensaje
// NUNCA incluye el argumento ni el token: el enrollment string es un secreto del mismo
// calibre que el salt (FR-007), es encoding y contiene el token en claro.
var ErrEnrollmentString = errors.New("enrollment string inválido")

// ErrEnrollmentObsolete señala el formato viejo pmea1, ya retirado. Su mensaje es accionable
// y NO reproduce el argumento ni el token (se detecta solo por el prefijo).
var ErrEnrollmentObsolete = errors.New("formato de enrollment obsoleto; solicita uno nuevo desde el panel")

// enrollmentPayload es la estructura CERRADA (additionalProperties:false) que viaja dentro
// del enrollment string pmea2: exactamente tres campos, ni uno más ni uno menos.
type enrollmentPayload struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
	DevID    string `json:"dev_id"`
}

// ParseEnrollmentString decodifica un enrollment string `pmea2.<base64url(json)>` en su
// terna (endpoint, token, dev_id). Es encoding, no cifrado. El par (endpoint, token) alimenta
// el ping de verificación; el dev_id es la identidad de desarrollador ASIGNADA por el backend
// (Owner/Admin), autoritativa: el agente la adopta, ya no la genera en local. La validez del
// token la decide el backend en el ping, no este formato; aquí solo se comprueba que el
// envoltorio está bien formado, que el endpoint es https y que el dev_id cumple su charset.
// Política de versión: pmea1 se RECHAZA (formato obsoleto); cualquier otro prefijo es "no
// reconocido". Ante cualquier fallo devuelve un error que NO reproduce el argumento ni el
// token (FR-007/FR-013).
func ParseEnrollmentString(s string) (endpoint, token, devID string, err error) {
	// Política de versión: pmea1 (formato viejo) se rechaza explícitamente ANTES que el
	// genérico "no reconocido", con un mensaje accionable. No se intenta migrar: el pmea1
	// no trae dev_id autoritativo, así que es intrínsecamente insuficiente.
	if strings.HasPrefix(s, enrollmentPrefixV1) {
		return "", "", "", fmt.Errorf("%w", ErrEnrollmentObsolete)
	}
	body, ok := strings.CutPrefix(s, enrollmentPrefix)
	if !ok {
		return "", "", "", fmt.Errorf("%w: prefijo no reconocido", ErrEnrollmentString)
	}
	raw, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return "", "", "", fmt.Errorf("%w: base64 ilegible", ErrEnrollmentString)
	}
	// Struct cerrada: DisallowUnknownFields rechaza campos extra (additionalProperties:false).
	var p enrollmentPayload
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return "", "", "", fmt.Errorf("%w: json ilegible", ErrEnrollmentString)
	}
	u, err := url.Parse(p.Endpoint)
	if err != nil || u.Scheme != "https" {
		return "", "", "", fmt.Errorf("%w: el endpoint debe ser https", ErrEnrollmentString)
	}
	if p.Token == "" {
		return "", "", "", fmt.Errorf("%w: token ausente", ErrEnrollmentString)
	}
	// dev_id autoritativo: no vacío, longitud 1–64, charset [A-Za-z0-9._-]. El valor inválido
	// NO se reproduce en el error (no es secreto, pero mantenemos el error libre de entrada).
	if p.DevID == "" {
		return "", "", "", fmt.Errorf("%w: dev_id ausente", ErrEnrollmentString)
	}
	if len(p.DevID) > devIDMaxLen {
		return "", "", "", fmt.Errorf("%w: dev_id excede %d caracteres", ErrEnrollmentString, devIDMaxLen)
	}
	if !devIDPattern.MatchString(p.DevID) {
		return "", "", "", fmt.Errorf("%w: dev_id con caracteres no permitidos", ErrEnrollmentString)
	}
	return p.Endpoint, p.Token, p.DevID, nil
}

// IsEnrolled indica si la config representa un enrolamiento válido: endpoint y token
// presentes y endpoint https (reutiliza Validate). Es el estado que expone `permea status`.
func IsEnrolled(c Config) bool {
	return c.Endpoint != "" && c.DeviceToken != "" && c.Validate() == nil
}
