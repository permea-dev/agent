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
// **Las réplicas las sustituyó P-005 T005.** Aquí nació el cuerpo; allí se fueron las copias.
//
// ═══ QUIÉN LLAMA AQUÍ — LOS CUATRO, Y LOS CUATRO EXISTEN ═══════════════════════════════════
//
// Esta lista es **la contrapartida de haber movido el juicio**: quien busque dónde se decide si un
// destino es admisible ya **no** lo encuentra donde estaba, así que el camino de vuelta tiene que
// estar escrito. **Son cuatro, ninguno pendiente:**
//
//	internal/config/enrollment.go   ParseEnrollmentString  — FUNDE los dos hechos y DESCARTA la causa
//	internal/config/config.go       Config.Validate        — DOS ramas, dos mensajes
//	internal/transport/transport.go Client.Send            — DOS ramas · centinela ErrScheme · causa con %w
//	internal/transport/transport.go Client.Adherir         — ídem: la SEGUNDA PUERTA de la frontera
//
// **Los cuatro difieren en el desenlace y coinciden en el juicio, que es exactamente el objetivo.**
// `enrollment.go` descarta la causa porque su argumento **lleva el token dentro**; `Adherir` la
// conserva porque su usuario necesita saber **qué** está mal en su configuración. Unificar el juicio y
// uniformar además el desenlace habría sido pasarse de largo.
//
// ═══ CÓMO SE COMPRUEBA QUE NO QUEDA NINGUNA COPIA — Y NO ES LEYENDO ════════════════════════
//
// Hay **dos pruebas, y hacen falta las dos** porque cubren mitades distintas:
//
//  1. **De comportamiento (P-005 T005 §PASO 2)**: mutar este cuerpo hace caer **los cuatro** tests de
//     los cuatro llamantes, con cuatro mensajes distintos. Demuestra que **todos dependen de aquí**.
//  2. **Estructural, y salió sola**: `enrollment.go` y `config.go` **perdieron el import de
//     `net/url`** —no por limpieza: **porque el compilador dejó de aceptarlo**—. Una réplica
//     disfrazada, una copia local que devolviera lo mismo, **habría conservado ese import**, porque
//     seguiría llamando a `url.Parse`. Demuestra que **ninguno guarda copia**.
//
// **Ninguna de las dos sola basta**: la primera no distingue «usa el original» de «tiene una copia que
// hace lo mismo»; la segunda no dice que el original se use de verdad. **Y la segunda se verifica con
// `go build`, no leyendo el código** — que es lo que la hace útil dentro de seis meses.
//
// La retirada del andamiaje de P-005 T002 se comprobó igual, por barrido: `grep -rn
// errEsquemaAndamiaje internal/ cmd/` → cero.

// esquemaAdmisible es el único esquema por el que este agente transmite. P-005 FR-017 no admite exención,
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
// **Las réplicas NO eran idénticas** (`research.md` §R4 las contó tres; T002 añadió la cuarta): `enrollment.go` **funde** el fallo de
// análisis y el esquema equivocado en **un solo desenlace** —y nunca reproduce el argumento, porque el
// `pmea2` lleva el token dentro—, mientras que `config.go` y `Send` los **separan en dos ramas** con
// mensajes distintos. Un `bool` obligaría a elegir: o `enrollment.go` pasa a distinguir dos casos
// donde hoy hay uno, o los otros pierden la distinción que hoy tienen. **Cualquiera de las dos
// cambiaría comportamiento**, y el síntoma sería uno de los cuatro tests de la red en rojo, con la
// puerta de T005 diciendo «parar» sin que nadie entienda por qué.
//
// ═══ POR QUÉ DEVUELVE EL ERROR Y NO UN SEGUNDO BOOLEANO ════════════════════════════════════
//
// Porque **TRES de los cuatro llamantes ENVUELVEN la causa** en su mensaje —`Config.Validate`,
// `Send` y `Adherir`— con `fmt.Errorf("endpoint inválido %q: %w", …, errAnalisis)`.
// *(Sin número de línea a propósito: las citas de línea envejecen y ésta ya envejeció una vez.)* Con un booleano no
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
