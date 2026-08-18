package config

import (
	"strings"
	"testing"
)

// ═══ P-005 T005 · PASO 0 · RED DE PRESERVACIÓN, NO PUERTA TDD ═════════════════════════════
//
// **Nace verde a propósito**: hoy `ParseEnrollmentString` **ya** emite su mensaje genérico sin
// reproducir el argumento. Esto no abre comportamiento nuevo — **protege el que hay** mientras T005
// sustituye la réplica de esquema de `enrollment.go:78-80` por `JuzgarEndpoint`.
//
// ═══ POR QUÉ HACE FALTA, Y POR QUÉ EL TEST QUE YA EXISTÍA NO BASTABA ══════════════════════
//
// `TestParseEnrollmentString_Rejects` (`enrollment_test.go:116`) **ya comprueba higiene**, y es una red
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

// TestParseEnrollmentString_ElErrorNoReproduceFragmentosDelArgumento es la red de preservación de
// T005 sobre la puerta del enrolamiento, y recorre LOS TRES CAMPOS decodificados.
func TestParseEnrollmentString_ElErrorNoReproduceFragmentosDelArgumento(t *testing.T) {
	const secret = "dev_tok_SUPERSECRET_zzz"

	// Los dos primeros casos son LOS QUE T005 TOCA: ambos atraviesan la rama de esquema de
	// `enrollment.go:78-80`, la que pasa a llamar a `JuzgarEndpoint`. Los otros dos cierran la familia:
	// el agujero no era del endpoint, era de CUALQUIER campo decodificado.
	// ⚠️ LOS VALORES SON DE ALTA ENTROPÍA A PROPÓSITO, y se aprendió tropezando: un test de ausencia
	// por subcadenas es **sensible al valor elegido**. El primer `dev_id` de esta tabla fue
	// `"maquina no permitida!01"`, y dio un ROJO FALSO — el mensaje genérico dice «dev_id con caracteres
	// no permitidos», y las dos cadenas comparten `" no perm"`. **No había fuga: había vocabulario
	// común.** Los valores tienen que parecerse a lo que son —identificadores y secretos— y no a la
	// prosa de los mensajes, o el test grita sin motivo y acaba desactivado.
	casos := []struct {
		nombre   string
		endpoint string
		token    string
		devID    string
	}{
		{"endpoint no analizable — la rama que T005 sustituye", "https://ejemplo\x7f.test/ingest", secret, "acme-dev-01"},
		{"endpoint en claro — la otra mitad de esa misma rama", "http://inseguro.example/ingest", secret, "acme-dev-01"},
		{"token ausente", "https://api.permea.example/ingest", "", "acme-dev-01"},
		{"dev_id con caracteres no permitidos", "https://api.permea.example/ingest", secret, "maq!Z7Q3kX9vLp"},
		{"dev_id excede 64 caracteres", "https://api.permea.example/ingest", secret, strings.Repeat("Z7Q3kX9v", 9)},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			in := mkEnroll(t, c.endpoint, c.token, c.devID)

			endpoint, token, devID, err := ParseEnrollmentString(in)
			if err == nil {
				t.Fatalf("premisa rota: esperaba error, got (%q, %q, %q)", endpoint, token, devID)
			}
			msg := err.Error()

			// El bucle es sobre LOS CAMPOS, y ésa es la diferencia con el test que ya existía: él
			// compara el argumento entero, y **ninguno de los tres campos es subcadena de
			// `pmea2.<base64>`** — van codificados dentro. Un mensaje que reprodujera cualquiera de los
			// tres pasaría su comprobación sin despeinarse.
			sensibles := []struct {
				campo string
				valor string
			}{
				{"del ARGUMENTO entero", in},
				{"del campo endpoint", c.endpoint},
				{"del campo token", c.token},
				{"del campo dev_id", c.devID},
			}

			for _, sn := range sensibles {
				// Un valor más corto que el umbral no tiene subcadenas de esa longitud: comprobarlo
				// daría un verde vacío, y hay que decir que no se comprobó en vez de contarlo como
				// comprobado. (El caso «token ausente» cae aquí: valor vacío.)
				if len([]rune(sn.valor)) < longitudFragmento {
					continue
				}
				if frag := fragmentoFiltrado(msg, sn.valor); frag != "" {
					t.Errorf("el error reproduce un fragmento %s (≥%d caracteres): %q\n"+
						"  mensaje: %q\n"+
						"  SC-005: el argumento lleva el token dentro, así que el error debe ser genérico",
						sn.campo, longitudFragmento, frag, msg)
				}
			}
		})
	}
}
