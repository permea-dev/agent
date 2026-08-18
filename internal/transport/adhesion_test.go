package transport

import (
	"errors"
	"net/url"
	"testing"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/testutil"
)

// ═══ P-005 T011 · FR-017 · LA GUARDA MUERDE EN LA SEGUNDA PUERTA ═══════════════════════════
//
// La adhesión es **la segunda puerta de la frontera de datos**, y FR-017 exige que su transporte
// seguro se juzgue **con la misma exigencia que la emisión de eventos, sin exención ni variante**.
// «La misma exigencia» incluye **el mismo centinela**: quien compruebe `errors.Is(err, ErrScheme)`
// debe obtener la misma respuesta por las dos puertas.
//
// ═══ ⚠️ ESTE TEST NACE ROJO, Y LA RAZÓN ES EL CENTINELA ════════════════════════════════════
//
// **No es que `Adherir` no tenga guarda**: la tiene desde P-005 T002 —una réplica inline— y rechaza
// el canal en claro. Lo que hace es devolver **`errEsquemaAndamiaje`**, un centinela de andamiaje
// **distinto** de `ErrScheme`, precisamente para que este test no pueda nacer verde acertando contra
// una condición copiada a mano.
//
// **Lo pone en verde P-005 T005**, al sustituir esa réplica por la función unificada de T004. Y es
// **la mitad de comportamiento** del criterio positivo de T005 —la otra mitad, la estructural, es la
// mutación de la función unificada que exige que los cuatro llamantes se muevan a la vez—.
//
// ═══ LÍMITE, y está escrito en tasks.md ════════════════════════════════════════════════════
//
// Este test acredita **que la puerta nueva devuelve el centinela correcto**. NO acredita que el
// juicio esté unificado: eso es estructura, y ningún test la ve — devolver `ErrScheme` se satisface
// también cambiando una palabra en la réplica inline.

// TestAdherir_RechazaCanalEnClaro es la garantía de FR-017 sobre la puerta de la adhesión.
func TestAdherir_RechazaCanalEnClaro(t *testing.T) {
	// Disciplina 6 — aislamiento. Este test no toca config ni cola, pero el sandbox cuesta cero y
	// elimina la pregunta: ninguna prueba de esta suite escribe en la instalación real.
	_ = testutil.Sandbox(t)

	casos := []struct {
		nombre   string
		endpoint string
	}{
		{"http en claro", "http://api.permea.example/api/v1/projects/adhesion"},
		{"http sobre la máquina local", "http://127.0.0.1:8000/api/v1/projects/adhesion"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cliente := New(c.endpoint, "tok-de-prueba")

			denominacion, err := cliente.Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")

			// (1) NO se completa: ningún desenlace de éxito, ninguna denominación.
			if err == nil {
				t.Fatalf("Adherir(%q) devolvió err = nil: el canal en claro DEBE rechazarse (FR-017)", c.endpoint)
			}
			if denominacion != "" {
				t.Errorf("Adherir(%q) devolvió denominación %q: no puede haber desenlace de éxito", c.endpoint, denominacion)
			}

			// (2) Y lo hace con EL CENTINELA de la ingesta. Es lo que hace que las dos puertas
			//     respondan igual a `errors.Is`, que es lo que FR-017 exige de verdad.
			if !errors.Is(err, ErrScheme) {
				t.Errorf("Adherir(%q): errors.Is(err, ErrScheme) = false, err = %v\n"+
					"  la guarda de la segunda puerta no devuelve el centinela de la ingesta", c.endpoint, err)
			}
		})
	}
}

// ═══ P-005 T011 · SEGUNDO CASO · LOS DOS HECHOS, DISTINGUIBLES EN LA SEGUNDA PUERTA ════════
//
// El primer caso comprueba que el canal en claro devuelve `ErrScheme`. **No basta**, y la razón es
// lo que P-005 T004 existe para impedir: un endpoint puede rechazarse por **dos hechos distintos**
// —«no analizable» y «esquema no admisible»— y **fundirlos en un solo desenlace es una regresión
// que ningún test de un solo caso ve**. Con sólo el primer caso, T005 podría satisfacerlo devolviendo
// `ErrScheme` **también** cuando la URL no se puede analizar, y estaría verde habiendo **refundido
// justo lo que T004 separó**.
//
// ═══ LA ASERCIÓN ES LA DISTINGUIBILIDAD, NO EL TEXTO ══════════════════════════════════════
//
// Este test **no compara ningún mensaje** — comparar texto lo ataría a una redacción, y la redacción
// es de cada llamante (T004 no formatea ninguna). Compara **las dos respuestas entre sí**: lo que
// `errors.Is(·, ErrScheme)` contesta a un hecho **no puede ser lo mismo** que contesta al otro.
//
// ═══ ⚠️ NACE ROJO, Y POR INDISTINGUIBILIDAD ═══════════════════════════════════════════════
//
// Hoy la guarda inline de T002 devuelve **el mismo `errEsquemaAndamiaje` para los dos hechos**, así
// que las dos respuestas coinciden y **son indistinguibles**. Lo pone en verde **T005**.
//
// ═══ LO QUE ESTE TEST DELIBERADAMENTE **NO** FIJA ═════════════════════════════════════════
//
// **No exige que el error de «no analizable» conserve la causa de `url.Parse`.** Hoy `Adherir` es una
// **cuarta variante** de la guarda: ofrece centinela y **tira la causa**, al revés que `Send`
// (`transport.go:143`), que conserva la causa y no ofrece centinela. Esa rama **no la mira nadie**, y
// unificarla es decisión de T005 — este test le deja las dos opciones abiertas y sólo le prohíbe
// **la que borra la diferencia**.

// endpointNoAnalizableAdhesion lleva un carácter de control, que hace fallar a url.Parse.
const endpointNoAnalizableAdhesion = "https://ejemplo\x7f.test/api/v1/projects/adhesion"

// endpointEnClaroAdhesion analiza SIN problema; lo que falla es su esquema. La pareja es el punto:
// los dos se rechazan, pero por hechos distintos.
const endpointEnClaroAdhesion = "http://api.permea.example/api/v1/projects/adhesion"

// TestAdherir_LosDosHechosSonDistinguibles es la garantía de que la segunda puerta no funde los dos
// hechos que P-005 T004 separó.
func TestAdherir_LosDosHechosSonDistinguibles(t *testing.T) {
	_ = testutil.Sandbox(t)

	// Los dos rechazos, recogidos por separado (disciplina 7).
	denomClaro, errClaro := New(endpointEnClaroAdhesion, "tok-de-prueba").Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")
	denomNoAnaliz, errNoAnaliz := New(endpointNoAnalizableAdhesion, "tok-de-prueba").Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")

	// (1) Los DOS rechazan, y ninguno se completa. Es la premisa del test: si uno de los dos pasara,
	//     lo que sigue compararía un rechazo con un éxito y no diría nada de la distinguibilidad.
	if errClaro == nil {
		t.Fatalf("Adherir(%q) devolvió err = nil: el canal en claro DEBE rechazarse", endpointEnClaroAdhesion)
	}
	if errNoAnaliz == nil {
		t.Fatalf("Adherir(%q) devolvió err = nil: un endpoint no analizable DEBE rechazarse", endpointNoAnalizableAdhesion)
	}
	if denomClaro != "" || denomNoAnaliz != "" {
		t.Errorf("hubo denominación en un rechazo: en claro = %q, no analizable = %q", denomClaro, denomNoAnaliz)
	}

	// (2) LA ASERCIÓN. Las dos respuestas a la MISMA pregunta, comparadas entre sí.
	claroEsScheme := errors.Is(errClaro, ErrScheme)
	noAnalizEsScheme := errors.Is(errNoAnaliz, ErrScheme)

	if claroEsScheme == noAnalizEsScheme {
		t.Errorf("los dos hechos son INDISTINGUIBLES: errors.Is(·, ErrScheme) contesta %v a los dos\n"+
			"  en claro       (%q) → %v\n"+
			"  no analizable  (%q) → %v\n"+
			"  la segunda puerta funde «no analizable» y «esquema no admisible» en un solo desenlace",
			claroEsScheme, endpointEnClaroAdhesion, claroEsScheme, endpointNoAnalizableAdhesion, noAnalizEsScheme)
	}

	// (3) Y en el sentido correcto: `ErrScheme` es un juicio SOBRE el esquema, así que sólo puede
	//     afirmarlo quien pudo leerlo. Sin esto, dos respuestas distintas pero invertidas pasarían.
	if !claroEsScheme {
		t.Errorf("errors.Is(err, ErrScheme) = false para %q: el esquema se leyó y no es admisible", endpointEnClaroAdhesion)
	}
	if noAnalizEsScheme {
		t.Errorf("errors.Is(err, ErrScheme) = true para %q: el esquema NO se pudo leer, "+
			"así que no se puede juzgar no admisible", endpointNoAnalizableAdhesion)
	}
}

// ═══ P-005 T011 · TERCER CASO · LA CAUSA SE CONSERVA, IGUAL QUE EN `Send` ══════════════════
//
// **Garantía: D-005-P2.** Con un endpoint no analizable, el error de `Adherir` **conserva la causa
// de `url.Parse`**, con **la misma forma que `Send`** (`transport.go:143`), que la envuelve con `%w`.
//
// ═══ POR QUÉ ES LA PREMISA DE D-005-P2, Y NO UNA PREFERENCIA ══════════════════════════════
//
// **Unificar el juicio y dejar que una puerta conserve la causa y la otra la tire es unificar el
// código y mantener la divergencia justo donde se nota: en lo que la persona lee cuando su
// configuración está rota.** El juicio compartido no le sirve de nada a quien tiene un endpoint mal
// escrito si una puerta le dice *qué* está mal y la otra sólo que algo lo está. D-005-P2 unifica **el
// desenlace**, no sólo la condición — si no, el trabajo se queda a medias exactamente en la mitad que
// el usuario ve.
//
// ═══ ⚠️ NACE ROJO — `Adherir` hace HOY LO CONTRARIO ═══════════════════════════════════════
//
// La réplica de andamiaje de T002 es **la cuarta variante** de la guarda: `fmt.Errorf("%w: endpoint
// inválido %q", errEsquemaAndamiaje, c.Endpoint)` — **envuelve el CENTINELA y DESCARTA la causa**, al
// revés que `Send`. Lo pone en verde **T005**.
//
// ═══ POR QUÉ `errors.As` Y NO `errors.Is` ═════════════════════════════════════════════════
//
// **Medido**: `url.Parse` devuelve un `*url.Error` **nuevo en cada llamada**, y `url.Error` no
// implementa `Is`. Así que `errors.Is(err, <causa de una segunda llamada>)` da **false aunque la causa
// esté perfectamente conservada** — compara punteros. El vehículo correcto es **`errors.As`** sobre
// `*url.Error` —tipo que **sólo** produce `url.Parse`— más sus **campos estructurados** (`Op`, `URL`).
// **En ningún punto se compara el texto de un mensaje.**

// TestAdherir_ConservaLaCausaDelParseo es la garantía D-005-P2 sobre la puerta de la adhesión.
func TestAdherir_ConservaLaCausaDelParseo(t *testing.T) {
	_ = testutil.Sandbox(t)

	// La causa de referencia, obtenida de la misma fuente que debería conservarse.
	_, errDirecto := url.Parse(endpointNoAnalizableAdhesion)
	var causaDirecta *url.Error
	if !errors.As(errDirecto, &causaDirecta) {
		t.Fatalf("premisa rota: url.Parse(%q) no devolvió un *url.Error, sino %T",
			endpointNoAnalizableAdhesion, errDirecto)
	}

	cliente := New(endpointNoAnalizableAdhesion, "tok-de-prueba")

	// EL ORÁCULO: `Send` es la forma de referencia. Se comprueba **la misma propiedad en las dos
	// puertas**, así que «misma forma que Send» no es una afirmación del comentario — es la aserción.
	// Si esta rama fallara, lo que cambió es la referencia, y entonces el que hay que revisar es este
	// test y no `Adherir`.
	errSend := cliente.Send([]event.Event{})
	var causaDeSend *url.Error
	if !errors.As(errSend, &causaDeSend) {
		t.Fatalf("LA REFERENCIA CAMBIÓ: Send(%q) ya no conserva la causa de url.Parse; err = %v (%T)\n"+
			"  este test se ancla en la forma de Send (transport.go:143): revísese la referencia antes que Adherir",
			endpointNoAnalizableAdhesion, errSend, errSend)
	}

	denominacion, errAdherir := cliente.Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")

	if errAdherir == nil {
		t.Fatalf("Adherir(%q) devolvió err = nil: un endpoint no analizable DEBE rechazarse",
			endpointNoAnalizableAdhesion)
	}
	if denominacion != "" {
		t.Errorf("Adherir devolvió denominación %q en un rechazo", denominacion)
	}

	// LA ASERCIÓN. La causa tiene que estar EN LA CADENA, no reproducida en el mensaje.
	var causaDeAdherir *url.Error
	if !errors.As(errAdherir, &causaDeAdherir) {
		t.Fatalf("Adherir(%q): errors.As(err, **url.Error) = false — LA CAUSA SE PERDIÓ.\n"+
			"  err = %v\n"+
			"  Send, con el mismo endpoint, SÍ la conserva: %v\n"+
			"  D-005-P2: las dos puertas deben dar el mismo desenlace, no sólo compartir la condición",
			endpointNoAnalizableAdhesion, errAdherir, errSend)
	}

	// Y que sea LA MISMA causa, por campos estructurados — nunca por texto.
	if causaDeAdherir.Op != causaDirecta.Op || causaDeAdherir.URL != causaDirecta.URL {
		t.Errorf("Adherir conserva UNA causa, pero no la de este endpoint:\n"+
			"  obtenida:  Op=%q URL=%q\n"+
			"  esperada:  Op=%q URL=%q",
			causaDeAdherir.Op, causaDeAdherir.URL, causaDirecta.Op, causaDirecta.URL)
	}
	if causaDeAdherir.Op != causaDeSend.Op || causaDeAdherir.URL != causaDeSend.URL {
		t.Errorf("las dos puertas conservan causas DISTINTAS para el mismo endpoint:\n"+
			"  Adherir:  Op=%q URL=%q\n"+
			"  Send:     Op=%q URL=%q",
			causaDeAdherir.Op, causaDeAdherir.URL, causaDeSend.Op, causaDeSend.URL)
	}
}
