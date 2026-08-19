package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// mkCuerpo devuelve solo el base64url del payload, sin prefijo: hace falta para construir los
// argumentos de las ramas que fallan ANTES de mirar el prefijo.
func mkCuerpo(t *testing.T, endpoint, token, devID string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"endpoint": endpoint, "token": token, "dev_id": devID})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// ═══ P-005 T005 · PASO 0 · RED DE PRESERVACIÓN, NO PUERTA TDD ═════════════════════════════
//
// **Nace verde a propósito**: hoy `ParseEnrollmentString` **ya** emite su mensaje genérico sin
// reproducir el argumento. Esto no abre comportamiento nuevo — **protege el que hay** mientras T005
// sustituye la réplica de esquema de `ParseEnrollmentString` por `JuzgarEndpoint`.
//
// ═══ POR QUÉ HACE FALTA, Y POR QUÉ EL TEST QUE YA EXISTÍA NO BASTABA ══════════════════════
//
// `TestParseEnrollmentString_Rejects` **ya comprueba higiene**, y es una red
// real — pero mide `strings.Contains(msg, tc.in)`: **el argumento ENTERO**, y el token entero.
//
// **El peligro que T005 introduce no tiene esa forma.** `JuzgarEndpoint` devuelve el error de
// `url.Parse`, que es un `*url.Error` — y **un `*url.Error` lleva dentro LA URL COMPLETA**. Si
// `enrollment.go` lo envolviera con `%w`, el mensaje filtraría **el endpoint**: que **no es el
// argumento entero** (el argumento es `pmea2.<base64>`) **ni es el token**. Las dos aserciones
// existentes seguirían en verde con el endpoint a la vista.
//
// Por eso este test mide **fragmentos**, con el umbral del contrato: **ninguna subcadena de ocho o
// más caracteres** del argumento puede aparecer en el mensaje (`spec.md` SC-005, `contracts/cli.md`
// §Garantías). Es el modo de T024, aplicado aquí a la puerta del enrolamiento.
//
// ═══ Y POR ESO enrollment.go NO ENVUELVE CON %w ═══════════════════════════════════════════
//
// De las cuatro puertas que T005 unifica, **ésta es la única donde DESCARTAR la causa es la conducta
// correcta**: su argumento **lleva el token dentro**, así que el error tiene que ser genérico. Es
// exactamente lo contrario de lo que se exige en `Adherir` (T011 caso 3, que **obliga** a conservar la
// causa). **Cada llamante decide, y eso es el punto de que `JuzgarEndpoint` no formatee ningún
// mensaje.**

// longitudFragmento es el umbral del contrato (SC-005): ninguna subcadena de esta longitud del valor
// sensible puede aparecer en la salida.
const longitudFragmento = 8

// fragmentoFiltrado devuelve la primera subcadena de `longitudFragmento` caracteres de `valor` que
// aparezca en `texto`, o "" si ninguna aparece. Devuelve el fragmento y no un booleano para que el
// mensaje de fallo pueda decir QUÉ se filtró, no sólo que algo se filtró.
func fragmentoFiltrado(texto, valor string) string {
	r := []rune(valor)
	for i := 0; i+longitudFragmento <= len(r); i++ {
		frag := string(r[i : i+longitudFragmento])
		if strings.Contains(texto, frag) {
			return frag
		}
	}
	return ""
}

// ═══ LA TABLA CUBRE LAS NUEVE RAMAS DE ERROR, Y ESO ES PARTE DEL CONTRATO ════════════════
//
// `ParseEnrollmentString` tiene **nueve desenlaces de error**, y la invariante —«ningún mensaje
// reproduce un fragmento de ningún campo»— **es de la función, no de unas cuantas ramas suyas**.
//
// ⛔ **UNA RAMA DE ERROR NUEVA OBLIGA A UNA FILA NUEVA.** Cubrir seis de nueve no se verifica de un
// vistazo; cubrir nueve de nueve sí, porque el número está escrito y se compara con el del código.
//
// **Y las tres primeras se incluyen aunque hoy sean triviales.** Disparan **antes de decodificar**, así
// que la struct está vacía y la aserción se cumple sola. Se quedan por dos razones: la de arriba —que
// «nueve de nueve» es comprobable y «las que importan» no—, y porque **el día que alguien mueva la
// decodificación más arriba, esas tres dejan de ser triviales sin que nadie se entere**. Una fila
// trivial cuesta nada; descubrir que dejó de serlo tres meses después cuesta el incidente.

// TestParseEnrollmentString_ElErrorNoReproduceFragmentosDelArgumento es la red de preservación de
// T005 sobre la puerta del enrolamiento: LAS NUEVE RAMAS × LOS TRES CAMPOS.
func TestParseEnrollmentString_ElErrorNoReproduceFragmentosDelArgumento(t *testing.T) {
	// ⚠️ FIXTURES DE ALTA ENTROPÍA, Y SE APRENDIÓ TROPEZANDO: un test de ausencia por subcadenas es
	// **sensible al valor elegido**. Un `dev_id` de esta tabla fue `"maquina no permitida!01"` y dio un
	// ROJO FALSO — el mensaje genérico dice «dev_id con caracteres **no permitidos**», y las dos cadenas
	// comparten `" no perm"`. **No había fuga: había vocabulario común.** Los valores tienen que
	// parecerse a lo que son —identificadores y secretos— y **no a la prosa de los mensajes**, o el test
	// grita sin motivo y acaba desactivado por ruidoso.
	const (
		secreto       = "dev_tok_9HpQ3mZv7KxR2wLb"
		endpointOK    = "https://r4Vx8Kq2.example/ingest"
		endpointClaro = "http://r4Vx8Kq2.example/ingest"
		endpointRoto  = "https://r4Vx8Kq2\x7f.example/ingest" // carácter de control: url.Parse falla
		devIDOK       = "n7Kq2Vx8"
		devIDCharset  = "n7K!q2Vx8L"
	)
	devIDLargo := strings.Repeat("n7Kq2Vx8", 9) // 72 > 64

	// El cuerpo válido, y el mismo cuerpo con un campo de más: `DisallowUnknownFields` falla, pero
	// **la struct ya quedó COMPLETA** — medido — así que en esa rama el token está vivo.
	cuerpoValido := mkCuerpo(t, endpointOK, secreto, devIDOK)
	cuerpoCampoExtra := base64.RawURLEncoding.EncodeToString([]byte(
		`{"endpoint":"` + endpointOK + `","token":"` + secreto + `","dev_id":"` + devIDOK + `","extra":"x"}`))

	casos := []struct {
		rama     int
		nombre   string
		in       string
		endpoint string
		token    string
		devID    string
	}{
		// ── Ramas 1-3: disparan ANTES de decodificar. Triviales HOY, y se comprueban igual. ──
		{1, "pmea1 obsoleto", mkEnrollV1(t, endpointOK, secreto), endpointOK, secreto, ""},
		{2, "prefijo no reconocido", "permea2." + cuerpoValido, endpointOK, secreto, devIDOK},
		{3, "base64 ilegible", "pmea2.###" + secreto + "###", endpointOK, secreto, devIDOK},

		// ── Rama 4: EL HUECO REAL. `unknown field` deja la struct COMPLETA antes de fallar, con el
		//    token vivo en `p`. Hoy no filtra; lo único que separa «no filtra» de «filtra» es que
		//    alguien añada contexto al mensaje mientras depura. Esta fila es la que lo impide. ──
		{4, "json ilegible (campo extra, struct COMPLETA)", "pmea2." + cuerpoCampoExtra, endpointOK, secreto, devIDOK},

		// ── Ramas 5-6: decodificado completo. La 5 es la que T005 sustituyó. ──
		{5, "endpoint no analizable — la rama que T005 sustituyó", mkEnroll(t, endpointRoto, secreto, devIDOK), endpointRoto, secreto, devIDOK},
		{5, "endpoint en claro — la otra mitad de esa misma rama", mkEnroll(t, endpointClaro, secreto, devIDOK), endpointClaro, secreto, devIDOK},
		{6, "token ausente", mkEnroll(t, endpointOK, "", devIDOK), endpointOK, "", devIDOK},

		// ── Rama 7: SEGUNDO HUECO. `dev_id` vacío, pero endpoint y token POBLADOS y sin testigo. ──
		{7, "dev_id ausente — endpoint y token poblados", mkEnroll(t, endpointOK, secreto, ""), endpointOK, secreto, ""},

		// ── Ramas 8-9 ──
		{8, "dev_id excede 64 caracteres", mkEnroll(t, endpointOK, secreto, devIDLargo), endpointOK, secreto, devIDLargo},
		{9, "dev_id con caracteres no permitidos", mkEnroll(t, endpointOK, secreto, devIDCharset), endpointOK, secreto, devIDCharset},
	}

	for _, c := range casos {
		t.Run(fmt.Sprintf("rama%d_%s", c.rama, c.nombre), func(t *testing.T) {
			endpoint, token, devID, err := ParseEnrollmentString(c.in)
			if err == nil {
				t.Fatalf("premisa rota: esperaba error, got (%q, %q, %q)", endpoint, token, devID)
			}
			msg := err.Error()

			// El bucle es sobre LOS CAMPOS, y ésa es la diferencia con el test que ya existía
			// (`TestParseEnrollmentString_Rejects`): él compara el argumento entero, y **ninguno de los tres
			// es subcadena de `pmea2.<base64>`** — van codificados dentro. Un mensaje que reprodujera
			// cualquiera de los tres pasaría su comprobación sin despeinarse.
			sensibles := []struct {
				campo string
				valor string
			}{
				{"del ARGUMENTO entero", c.in},
				{"del campo endpoint", c.endpoint},
				{"del campo token", c.token},
				{"del campo dev_id", c.devID},
			}

			for _, sn := range sensibles {
				// Un valor más corto que el umbral no tiene subcadenas de esa longitud: comprobarlo
				// daría un verde vacío, y hay que decir que no se comprobó en vez de contarlo como
				// comprobado. (Los campos ausentes de las ramas 1, 6 y 7 caen aquí: valor vacío.)
				if len([]rune(sn.valor)) < longitudFragmento {
					continue
				}
				if frag := fragmentoFiltrado(msg, sn.valor); frag != "" {
					t.Errorf("rama %d: el error reproduce un fragmento %s (≥%d caracteres): %q\n"+
						"  mensaje: %q\n"+
						"  SC-005: el argumento lleva el token dentro, así que el error debe ser genérico",
						c.rama, sn.campo, longitudFragmento, frag, msg)
				}
			}
		})
	}
}
