package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ═══ P-005 T015 · EL DESTINO DE LA ADHESIÓN SE DERIVA DEL DE LA INGESTA ═══════════════════
//
// **Contrato: `contracts/adhesion.md` §Cómo se obtiene `<base>`.** El agente **no guarda una URL
// base**: guarda **la ruta completa del endpoint de ingesta**, tal como se la entregó el enrolamiento.
// De ahí salen los dos hechos que esta función usa.
//
// ⛔ **LOS DOS HECHOS SON CONTRATO, NO LITERALES LOCALES** (D-005-P3, D-005-P11). Si vivieran aquí
// como constantes sin respaldo, serían **dos suposiciones sobre un servidor ajeno**: el día que la
// plataforma moviera una ruta, el agente fallaría con «forma inesperada» **culpando al usuario de un
// cambio del servidor**. Están escritos en el contrato, que es **interfaz pública** —por eso el fallo
// tiene dónde mirarse, y quien escriba otro cliente sabe qué puede dar por bueno.

// segmentoIngesta es el hecho 1 del contrato: la ruta de ingesta **termina en este segmento**. Es lo
// que se VALIDA antes de derivar; si no está, se rehúsa en vez de conjeturar (P-005 FR-009).
const segmentoIngesta = "ingest"

// segmentosAdhesion es el hecho 2: el segmento de la adhesión, **hermano de `/ingest` bajo el mismo
// prefijo**. Es lo que se pone en su lugar, conservando esquema, host, puerto y prefijo.
var segmentosAdhesion = []string{"projects", "adhesion"}

// ErrFormaDeEndpointInesperada es el rehúse de P-005 FR-009: el endpoint guardado no tiene la forma
// que permite derivar el destino **con confianza**.
var ErrFormaDeEndpointInesperada = errors.New("config: el endpoint configurado no tiene la forma esperada")

// DerivarEndpointDeAdhesion traduce el endpoint de ingesta guardado al de la adhesión, conservando
// **esquema, host, puerto y prefijo** y sustituyendo **solo el último segmento**.
//
//	guardado:   https://api.permea.example/api/v1/ingest
//	derivado:   https://api.permea.example/api/v1/projects/adhesion
//	                                       └── prefijo conservado ──┘
//
// ═══ EL REHÚSE NOMBRA LA FORMA, NUNCA LO HALLADO (P-005 FR-009 + FR-020) ══════════════════
//
// FR-009 pide **nombrar la forma de lo hallado**; FR-020 prohíbe reproducir material sensible **y
// manda sobre él**. Así que la forma se describe **estructuralmente** —ruta vacía, cuántos segmentos,
// qué segmento se esperaba— y **nunca citando lo hallado**: el último segmento de un endpoint mal
// configurado puede ser cualquier cosa, incluido un token pegado por error.
//
// **Por eso la causa de `url.Parse` se DESCARTA aquí**, al revés que en `transport.Adherir`: un
// `*url.Error` **lleva la URL entera dentro**, y envolverlo sería la fuga exacta que FR-020 prohíbe.
// Es la misma decisión que toma `ParseEnrollmentString`, y por el mismo motivo. *(Se unifica el
// juicio, no la presentación: ver `JuzgarEndpoint`.)*
func DerivarEndpointDeAdhesion(endpointIngesta string) (string, error) {
	u, err := url.Parse(endpointIngesta)
	if err != nil {
		return "", fmt.Errorf("%w: no es analizable como URL", ErrFormaDeEndpointInesperada)
	}
	if u.Path == "" || u.Path == "/" {
		return "", fmt.Errorf("%w: la ruta está vacía y se esperaba que terminara en el segmento %q",
			ErrFormaDeEndpointInesperada, segmentoIngesta)
	}

	segmentos := strings.Split(u.Path, "/")
	if segmentos[len(segmentos)-1] != segmentoIngesta {
		// Se nombra CUÁNTOS segmentos hay y CUÁL se esperaba — nunca cuál se encontró.
		return "", fmt.Errorf("%w: la ruta tiene %d segmentos y el último no es %q",
			ErrFormaDeEndpointInesperada, len(segmentos)-1, segmentoIngesta)
	}

	// Se sustituye EL ÚLTIMO segmento, no el primero que coincida: un prefijo puede contener el mismo
	// nombre (`/ingest/v1/ingest`) y cortar por el primero mandaría la petición a otro sitio.
	derivado := *u // copia: conserva esquema, host, puerto, y lo demás que traiga
	derivado.Path = strings.Join(append(segmentos[:len(segmentos)-1:len(segmentos)-1], segmentosAdhesion...), "/")
	return derivado.String(), nil
}
