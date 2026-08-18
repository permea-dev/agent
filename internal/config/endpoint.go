package config

import "net/url"

// ═══ P-005 T004 · EL JUICIO DE ESQUEMA, EN UN SOLO SITIO ═══════════════════════════════════
//
// Hasta P-005 esta comprobación estaba escrita CUATRO veces —`enrollment.go`, `config.go`, y dos en
// `internal/transport` (`Send` y `Adherir`)—. Cuatro copias de la frontera son cuatro sitios donde
// alguien escribe tres bien y una mal, y el fallo es **emitir por canal en claro**: Principio I, no
// una molestia. Es el defecto de clase que la plataforma ya pagó y dejó escrito —*«una condición
// replicada, corregida en un sitio y olvidada en cuatro»*—.
//
// **Las réplicas las sustituye P-005 T005, no esta tarea.** Aquí solo nace el cuerpo.

// esquemaAdmisible es el único esquema por el que este agente transmite. FR-017 no admite exención,
// modo de desarrollo ni variante: es la misma frontera para la ingesta y para la adhesión.
const esquemaAdmisible = "https"

// JuzgarEndpoint responde las DOS preguntas que las réplicas hacían, **por separado y sin decidir qué
// hacer con ellas**:
//
//	errAnalisis != nil   → el endpoint NO es analizable como URL, y viene la causa
//	admisible == false   → es analizable, pero su esquema no es el admisible
//
// ═══ POR QUÉ DOS HECHOS Y NO UN BOOLEANO ═══════════════════════════════════════════════════
//
// **Las tres réplicas NO son idénticas** (`research.md` §R4): `enrollment.go` **funde** el fallo de
// análisis y el esquema equivocado en **un solo desenlace** —y nunca reproduce el argumento, porque el
// `pmea2` lleva el token dentro—, mientras que `config.go` y `Send` los **separan en dos ramas** con
// mensajes distintos. Un `bool` obligaría a elegir: o `enrollment.go` pasa a distinguir dos casos
// donde hoy hay uno, o los otros pierden la distinción que hoy tienen. **Cualquiera de las dos
// cambiaría comportamiento**, y el síntoma sería uno de los cuatro tests de la red en rojo, con la
// puerta de T005 diciendo «parar» sin que nadie entienda por qué.
//
// ═══ POR QUÉ DEVUELVE EL ERROR Y NO UN SEGUNDO BOOLEANO ════════════════════════════════════
//
// Porque **dos de los llamantes ENVUELVEN la causa** en su mensaje: `config.go:102` hace
// `fmt.Errorf("endpoint inválido %q: %w", …, err)` y `transport.go:143` lo mismo. Con un booleano no
// tendrían qué envolver, y su mensaje **cambiaría** — que es exactamente lo que la condición de parada
// de T005 prohíbe. Devolver la causa **no es formatear**: el mensaje lo sigue construyendo cada
// llamante, que es el punto entero de la función.
//
// ═══ LO QUE ESTA FUNCIÓN NO HACE, Y ES DELIBERADO ══════════════════════════════════════════
//
// **No formatea ningún mensaje y no decide nada.** No sabe si el llamante quiere fundir los dos
// hechos o separarlos, ni si puede reproducir el argumento —`enrollment.go` NO puede—. Devuelve el
// juicio; la presentación es de quien llama. **Se unifica el juicio, no la presentación.**
func JuzgarEndpoint(endpoint string) (errAnalisis error, admisible bool) {
	u, err := url.Parse(endpoint)
	if err != nil {
		// No analizable: el segundo hecho **no se puede juzgar**, y decir `true` sería afirmar algo
		// que no se ha comprobado. Los llamantes que funden los dos hechos dependen de esto.
		return err, false
	}
	return nil, u.Scheme == esquemaAdmisible
}
