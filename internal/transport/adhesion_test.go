package transport

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/testutil"
)

// ═══ P-005 T011 · P-005 FR-017 · LA GUARDA MUERDE EN LA SEGUNDA PUERTA ═══════════════════════════
//
// La adhesión es **la segunda puerta de la frontera de datos**, y P-005 FR-017 exige que su transporte
// seguro se juzgue **con la misma exigencia que la emisión de eventos, sin exención ni variante**.
// «La misma exigencia» incluye **el mismo centinela**: quien compruebe `errors.Is(err, ErrScheme)`
// debe obtener la misma respuesta por las dos puertas.
//
// ═══ ⚠️ NACIÓ ROJO, Y LA RAZÓN FUE EL CENTINELA (registro histórico) ══════════════════════
//
// **No era que `Adherir` no tuviera guarda**: la tenía desde P-005 T002 —una réplica inline— y
// rechazaba el canal en claro. Lo que hacía era devolver un centinela de andamiaje **distinto** de
// `ErrScheme`, precisamente para que este test no pudiera nacer verde acertando contra una condición
// copiada a mano. **Ese centinela ya no existe: lo retiró T005**, y por eso no se nombra aquí — un
// identificador muerto en un comentario se lee como código vivo.
//
// **Lo puso en verde P-005 T005**, al sustituir esa réplica por la función unificada de T004. Y es
// **la mitad de comportamiento** del criterio positivo de T005 —la otra mitad, la estructural, es la
// mutación de la función unificada que exige que los cuatro llamantes se muevan a la vez—.
//
// ═══ LÍMITE, y está escrito en tasks.md ════════════════════════════════════════════════════
//
// Este test acredita **que la puerta nueva devuelve el centinela correcto**. NO acredita que el
// juicio esté unificado: eso es estructura, y ningún test la ve — devolver `ErrScheme` se satisface
// también cambiando una palabra en la réplica inline.

// TestAdherir_RechazaCanalEnClaro es la garantía de P-005 FR-017 sobre la puerta de la adhesión.
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
				t.Fatalf("Adherir(%q) devolvió err = nil: el canal en claro DEBE rechazarse (P-005 FR-017)", c.endpoint)
			}
			if denominacion != "" {
				t.Errorf("Adherir(%q) devolvió denominación %q: no puede haber desenlace de éxito", c.endpoint, denominacion)
			}

			// (2) Y lo hace con EL CENTINELA de la ingesta. Es lo que hace que las dos puertas
			//     respondan igual a `errors.Is`, que es lo que P-005 FR-017 exige de verdad.
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
// ═══ ⚠️ NACIÓ ROJO, Y POR INDISTINGUIBILIDAD (registro histórico) ═════════════════════════
//
// La guarda inline de T002 devolvía **el mismo centinela de andamiaje para los dos hechos**, así que
// las dos respuestas coincidían y **eran indistinguibles**. Lo puso en verde **T005**, que además
// tuvo que demostrar que esta aserción **es falsable** — hasta entonces pasaba por accidente.
//
// ═══ LO QUE ESTE TEST DELIBERADAMENTE **NO** FIJA ═════════════════════════════════════════
//
// **No exige que el error de «no analizable» conserve la causa de `url.Parse`.** Hoy `Adherir` es una
// **cuarta variante** de la guarda: ofrece centinela y **tira la causa**, al revés que `Send`
// (`Client.Send`), que conserva la causa y no ofrece centinela. Esa rama **no la mira nadie**, y
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
// de `url.Parse`**, con **la misma forma que `Client.Send`**, que la envuelve con `%w`.
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
// ═══ ⚠️ NACIÓ ROJO — `Adherir` hacía LO CONTRARIO (registro histórico) ════════════════════
//
// La réplica de andamiaje de T002 era **la cuarta variante** de la guarda: envolvía **el centinela de
// andamiaje y DESCARTABA la causa**, al revés que `Send`. Lo puso en verde **T005**.
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
			"  este test se ancla en la forma de Client.Send: revísese la referencia antes que Adherir",
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

// ═══ P-005 Phase 4 · LOS DESENLACES DE LA ADHESIÓN ════════════════════════════════════════
//
// `contracts/adhesion.md` §Los cuatro desenlaces. Los pone en verde **T012**; aquí nacen **rojos**
// contra el andamiaje de T002, que devolvía un centinela propio para todo — retirado por T012.
//
// ═══ ⛔ LAS ASERCIONES SE DERIVAN DE LO QUE LA RESPUESTA LLEVA ═════════════════════════════
//
// Los cuerpos se escriben **como JSON literal, tal cual sale del servidor**, y **nunca** serializando
// una struct de este repositorio. Serializarla haría que el test viera **lo que ese tipo admite** en
// vez de **lo que el servidor mandó** — es el defecto que ya se midió tres veces en esta feature
// (`tasks.md` disciplina 8 §un grep que da cero certifica). Aquí importa el doble, porque **T010
// existe precisamente para los cuerpos que una struct decodifica sin protestar**.

// backendAdhesion levanta un backend HTTPS de test que responde SIEMPRE con `estado` y con `cuerpo`
// **literal**, y devuelve un Client que confía en su certificado. Nada de red real (disciplina 6).
func backendAdhesion(t *testing.T, estado int, cuerpo string) *Client {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(estado)
		_, _ = w.Write([]byte(cuerpo))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "tok-de-prueba")
	c.HTTP = srv.Client()
	return c
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T008 · EL DESENLACE DE ÉXITO — desenlaces 3 y 4
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestAdherir_ExitoDevuelveLaDenominacion cubre los desenlaces 3 y 4, que **son la misma respuesta**:
// `200` con `{"project":{"name":…}}`. El contrato prohíbe distinguirlos —quien pega un código dos
// veces no debe poder saber cuál de las dos surtió efecto—, así que aquí hay **un solo caso**, y eso
// es fiel al contrato, no un atajo.
func TestAdherir_ExitoDevuelveLaDenominacion(t *testing.T) {
	_ = testutil.Sandbox(t)

	const denominacionEsperada = "Plataforma Permea"
	cliente := backendAdhesion(t, http.StatusOK, `{"project":{"name":"Plataforma Permea"}}`)

	denominacion, err := cliente.Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")

	if err != nil {
		t.Fatalf("Adherir con 200 y denominación legible devolvió err = %v; se esperaba éxito", err)
	}
	if denominacion != denominacionEsperada {
		t.Errorf("denominación = %q, want %q — P-005 FR-002 exige comunicar la del Proyecto",
			denominacion, denominacionEsperada)
	}
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T009 · LOS DOS RECHAZOS, DISTINTOS ENTRE SÍ — desenlaces 1 y 2
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestAdherir_LosDosRechazos cubre los desenlaces 1 (`422`) y 2 (`409`) **y que no son el mismo**.
//
// ⛔ **El discriminante es EL ESTADO, no el cuerpo** (`adhesion.md` §Qué distingue a qué): es el
// barato y el que no depende de interpretar nada. El cuerpo **confirma**, no decide — y este test lo
// exige de la única forma que lo demuestra: **cruzando los cuerpos**. Un `422` que llegue con el
// cuerpo del `409` **sigue siendo el desenlace 1**. Una implementación que ramificara por el cuerpo
// pasaría los dos casos rectos y caería aquí.
func TestAdherir_LosDosRechazos(t *testing.T) {
	_ = testutil.Sandbox(t)

	casos := []struct {
		nombre    string
		estado    int
		cuerpo    string
		centinela error
		otro      error // el centinela del OTRO rechazo: nunca debe casar
	}{
		{"422 · código no utilizable", http.StatusUnprocessableEntity,
			`{"error":"adhesion_rejected"}`, ErrCodigoNoUtilizable, ErrIdentidadYaAsignada},
		{"409 · identidad ya asignada", http.StatusConflict,
			`{"error":"identity_already_assigned"}`, ErrIdentidadYaAsignada, ErrCodigoNoUtilizable},

		// EL DISCRIMINANTE ES EL ESTADO — cuerpos cruzados, y el desenlace no se mueve.
		{"422 con el cuerpo del 409 · manda el estado", http.StatusUnprocessableEntity,
			`{"error":"identity_already_assigned"}`, ErrCodigoNoUtilizable, ErrIdentidadYaAsignada},
		{"409 con el cuerpo del 422 · manda el estado", http.StatusConflict,
			`{"error":"adhesion_rejected"}`, ErrIdentidadYaAsignada, ErrCodigoNoUtilizable},

		// Y sin cuerpo interpretable: el estado basta, porque el cuerpo nunca fue el discriminante.
		{"422 sin cuerpo · el estado basta", http.StatusUnprocessableEntity,
			``, ErrCodigoNoUtilizable, ErrIdentidadYaAsignada},
		{"409 con cuerpo ilegible · el estado basta", http.StatusConflict,
			`no soy json`, ErrIdentidadYaAsignada, ErrCodigoNoUtilizable},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cliente := backendAdhesion(t, c.estado, c.cuerpo)

			denominacion, err := cliente.Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")

			if denominacion != "" {
				t.Errorf("un rechazo devolvió denominación %q: no hay desenlace de éxito", denominacion)
			}
			if !errors.Is(err, c.centinela) {
				t.Errorf("estado %d: errors.Is(err, %v) = false; err = %v", c.estado, c.centinela, err)
			}
			if errors.Is(err, c.otro) {
				t.Errorf("estado %d: casa TAMBIÉN con %v — los dos rechazos se han fundido", c.estado, c.otro)
			}
		})
	}
}

// TestAdherir_LosDosRechazosSonDistinguibles es la aserción de DISTINGUIBILIDAD, y va aparte de las
// individuales a propósito.
//
// **Dos aserciones individuales correctas pasan igual si la implementación funde los dos desenlaces en
// uno solo** cuyo error case con los dos centinelas —envolviendo ambos, por ejemplo—. Y **refundir es
// la forma natural de simplificar**: es exactamente lo que ya pasó en la segunda puerta con «no
// analizable» y «esquema no admisible» (caso 2 de T011). Esto lo impide.
func TestAdherir_LosDosRechazosSonDistinguibles(t *testing.T) {
	_ = testutil.Sandbox(t)

	_, err422 := backendAdhesion(t, http.StatusUnprocessableEntity,
		`{"error":"adhesion_rejected"}`).Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")
	_, err409 := backendAdhesion(t, http.StatusConflict,
		`{"error":"identity_already_assigned"}`).Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")

	if err422 == nil || err409 == nil {
		t.Fatalf("premisa rota: los dos deben rechazar; 422 → %v, 409 → %v", err422, err409)
	}

	// Las respuestas de los DOS centinelas a los DOS errores, comparadas entre sí.
	for _, cent := range []struct {
		nombre string
		err    error
	}{
		{"ErrCodigoNoUtilizable", ErrCodigoNoUtilizable},
		{"ErrIdentidadYaAsignada", ErrIdentidadYaAsignada},
	} {
		a, b := errors.Is(err422, cent.err), errors.Is(err409, cent.err)
		if a == b {
			t.Errorf("los desenlaces 1 y 2 son INDISTINGUIBLES: errors.Is(·, %s) contesta %v a los dos\n"+
				"  422 → %v\n  409 → %v\n"+
				"  el contrato los distingue por estado; fundirlos borra esa distinción",
				cent.nombre, a, err422, err409)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T010 · LA RESPUESTA ININTERPRETABLE — P-005 FR-002 + P-005 FR-013
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestAdherir_DoscientosSinNombreLegibleEsNoVerificable exige que **un `200` sin `project.name`
// legible sea NO VERIFICABLE y NUNCA un éxito**: *«un éxito cuyo cuerpo no se pueda interpretar no es
// un éxito»* (`contracts/adhesion.md`).
//
// ⛔ **Se ENUMERAN las formas de «sin nombre legible», no se elige una con suerte.** Cada fila es **un
// camino distinto por el que un decodificador descuidado devuelve la cadena vacía como si fuera una
// denominación**: con `encoding/json` y una struct de campos, la clave ausente, el `null` y el objeto
// anidado ausente **producen los tres el mismo cero silencioso**, y el tipo equivocado produce un
// error que es fácil tratar como «campo vacío». Probar una sola forma deja las otras abiertas.
func TestAdherir_DoscientosSinNombreLegibleEsNoVerificable(t *testing.T) {
	_ = testutil.Sandbox(t)

	casos := []struct {
		nombre string
		cuerpo string
	}{
		{"clave name ausente", `{"project":{}}`},
		{"objeto project ausente", `{}`},
		{"name null", `{"project":{"name":null}}`},
		{"name cadena vacía", `{"project":{"name":""}}`},
		{"name sólo espacios", `{"project":{"name":"   "}}`},
		{"name numérico", `{"project":{"name":42}}`},
		{"name objeto", `{"project":{"name":{"es":"Plataforma"}}}`},
		{"name lista", `{"project":{"name":["Plataforma"]}}`},
		{"project no es objeto", `{"project":"Plataforma"}`},
		{"project null", `{"project":null}`},
		{"cuerpo vacío", ``},
		{"cuerpo no JSON", `no soy json`},
		{"JSON que no es objeto", `[1,2,3]`},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cliente := backendAdhesion(t, http.StatusOK, c.cuerpo)

			denominacion, err := cliente.Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")

			// (1) NUNCA éxito. Un `200` no basta: P-005 FR-002 exige comunicar la denominación.
			if err == nil {
				t.Fatalf("200 con cuerpo %q devolvió err = nil y denominación %q: "+
					"un éxito sin nombre NO es un éxito (P-005 FR-002)", c.cuerpo, denominacion)
			}
			// (2) Y ninguna denominación inventada: la cadena vacía tampoco vale como nombre.
			if denominacion != "" {
				t.Errorf("200 con cuerpo %q devolvió denominación %q: no había nombre que leer",
					c.cuerpo, denominacion)
			}
			// (3) El desenlace es NO VERIFICABLE, no un rechazo: el estado remoto queda
			//     indeterminado, y afirmar un rechazo sería afirmar un desenlace (P-005 FR-013).
			if !errors.Is(err, ErrNoVerificable) {
				t.Errorf("200 con cuerpo %q: errors.Is(err, ErrNoVerificable) = false; err = %v",
					c.cuerpo, err)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T010-E · EL ESTADO QUE EL CONTRATO NO ENUMERA — Principio I
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestAdherir_EstadoNoContempladoEsNoVerificable exige que un estado fuera de los tres del contrato
// —`200`, `422`, `409`— sea **NO VERIFICABLE**, y **NUNCA** un éxito.
//
// ═══ POR QUÉ NO ES DISCRECIONAL ═══════════════════════════════════════════════════════════
//
// **Principio I: un desenlace que no se reconoce NO PUEDE tratarse como conforme.** La implementación
// natural de «distinguir por estado» es un `switch`, y **un `switch` tiene un `default` que hará algo
// que nadie especificó**. Sin este test, ese `default` es una decisión tomada por descuido en el punto
// exacto donde el cliente decide si afirmar que se unió a un Proyecto.
//
// ═══ QUÉ DICE EL CONTRATO, MEDIDO ═════════════════════════════════════════════════════════
//
// `contracts/adhesion.md` **no enumera ningún estado fuera de esos tres**, pero **tampoco calla**: su
// cláusula general dice que *«si el desenlace no puede establecerse —servidor inalcanzable, o
// **respuesta que no permite determinarlo**—, el cliente informa de que no se pudo completar y NUNCA
// afirma ningún desenlace (FR-013)»*, y `contracts/cli.md` repite la fórmula en su fila **NV**. Un
// `500` es exactamente «una respuesta que no permite determinar el desenlace».
//
// ⚠️ **Lo que sí es un hueco, y va anotado al backlog** (no se arregla aquí: `adhesion.md` es artefacto
// de Phase 1 del plan): la tabla §Qué distingue a qué resume el reparto como **«`200` frente a `4xx`»**,
// **como si fuera exhaustivo**. Quien implemente leyendo esa fila escribe `if 200 {éxito} else
// {rechazo}` — y convierte un `500` en un rechazo afirmado. **La cláusula general lo cubre; la tabla
// invita a lo contrario.**
func TestAdherir_EstadoNoContempladoEsNoVerificable(t *testing.T) {
	_ = testutil.Sandbox(t)

	casos := []struct {
		nombre string
		estado int
		cuerpo string
	}{
		{"500 · error del servidor", http.StatusInternalServerError, `{"error":"boom"}`},
		{"403 · prohibido", http.StatusForbidden, `{"error":"forbidden"}`},
		{"404 · no encontrado", http.StatusNotFound, `{"error":"not_found"}`},
		// El más traicionero: es 2xx, así que un `if estado/100 == 2` lo daría por éxito.
		{"204 · sin contenido (2xx que NO es 200)", http.StatusNoContent, ``},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cliente := backendAdhesion(t, c.estado, c.cuerpo)

			denominacion, err := cliente.Adherir("pmeaj1.codigo-de-prueba", "ref-de-prueba")

			if err == nil {
				t.Fatalf("estado %d devolvió err = nil y denominación %q: un desenlace no reconocido "+
					"NO puede tratarse como conforme (Principio I)", c.estado, denominacion)
			}
			if denominacion != "" {
				t.Errorf("estado %d devolvió denominación %q: no hubo unión que afirmar", c.estado, denominacion)
			}
			// NO VERIFICABLE, y no un rechazo: afirmar un rechazo también es afirmar un desenlace.
			if !errors.Is(err, ErrNoVerificable) {
				t.Errorf("estado %d: errors.Is(err, ErrNoVerificable) = false; err = %v", c.estado, err)
			}
			for _, prohibido := range []struct {
				nombre string
				err    error
			}{{"ErrCodigoNoUtilizable", ErrCodigoNoUtilizable}, {"ErrIdentidadYaAsignada", ErrIdentidadYaAsignada}} {
				if errors.Is(err, prohibido.err) {
					t.Errorf("estado %d se clasificó como %s: el contrato NO lo asigna a ningún rechazo",
						c.estado, prohibido.nombre)
				}
			}
		})
	}
}
