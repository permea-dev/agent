package transport

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/testutil"
)

// ═══ P-005 T007 · GOLDEN DE FRONTERA · LA SEGUNDA PUERTA ══════════════════════════════════
//
// **Principio IV — test-first en la frontera.** P-005 **abre la segunda puerta de la frontera de
// datos**: hasta ahora lo único que salía de esta máquina hacia el exterior eran **eventos**, y el
// golden que los vigila está en `internal/ingest/boundary_test.go`. **Ese golden no cubre esto**: mira
// el evento serializado, la cola y el cuerpo de la ingesta — tres caminos que la adhesión no recorre.
//
// La adhesión emite **un cuerpo propio, con una forma propia, por un camino propio**. Sin este
// testigo, la puerta nueva sale al exterior **sin nadie mirando**, que es exactamente lo que el
// Principio I no admite.
//
// ═══ ⛔ ALLOWLIST, NO DENYLIST — Y ES LA DIFERENCIA ENTRE PROBAR ALGO Y NO PROBAR NADA ═════
//
// La aserción es **«el conjunto de claves del cuerpo es EXACTAMENTE {code, project_ref}»**, no «no
// aparece tal campo». Una denylist enumera lo que ya se temía, y **pasa con cualquier campo que nadie
// previó** — que es precisamente el caso que importa: la fuga que llega es la que no estaba en la
// lista. Con allowlist, **un campo nuevo rompe el test el día que se añade**, sin que nadie haya
// tenido que anticiparlo. Es la misma forma que la allowlist cerrada de la ingesta.
//
// ═══ ⚠️ NACE VERDE — y por eso se valida por MUTACIÓN ═════════════════════════════════════
//
// `Adherir` compone el cuerpo desde P-005 T002, así que este test pasa desde el primer instante. **Un
// golden que nunca ha fallado no es un golden, es una decoración.** Se valida añadiendo un TERCER
// campo al cuerpo: debe tumbarlo (registro en `tasks.md` T007).
//
// ═══ 🔗 EL OTRO TESTIGO DE LA FRONTERA ════════════════════════════════════════════════════
//
// **`internal/ingest/boundary_test.go`** vigila la PRIMERA puerta —la emisión de eventos, por sus tres
// caminos—. **Los dos ficheros se leen juntos**: desde P-005 la frontera tiene **dos puertas**, y la
// cabecera de aquél lleva la nota recíproca. Quien toque una y no mire la otra deja media frontera sin
// testigo.
//
// ═══ ⚠️ COLISIÓN DE NUMERACIÓN — SE CITA SIEMPRE CON PREFIJO DE SPEC ══════════════════════
//
// **`P-004 FR-017`** es **el ALCANCE de la frontera** —qué caminos hacia el exterior hay que vigilar—,
// y es el que gobierna el fichero de al lado. **`P-005 FR-017`** es **el TRANSPORTE SEGURO** de la
// adhesión, y es el que vigila `adhesion_test.go`. **Mismo número, specs distintas**, y los dos se
// citan en la cabecera del otro testigo.
//
// **En este repositorio el número NUNCA va suelto: `P-004 FR-017`, `P-005 FR-017`** — la convención que
// ya usa la spec. Este golden no depende de ninguno de los dos: lo suyo es **Principio I**, la
// allowlist. Pero la nota va aquí porque **es el fichero que empareja con el que sí los mezcla**.

// clavesPermitidasAdhesion es LA ALLOWLIST del cuerpo de adhesión: dos elementos, y ni uno más.
// `contracts/adhesion.md` §La petición.
var clavesPermitidasAdhesion = []string{"code", "project_ref"}

// TestFronteraAdhesion_ElCuerpoLlevaExactamenteDosClaves es el golden de la segunda puerta.
func TestFronteraAdhesion_ElCuerpoLlevaExactamenteDosClaves(t *testing.T) {
	_ = testutil.Sandbox(t)

	const (
		codigo     = "pmeaj1.codigo-de-adhesion-de-prueba"
		projectRef = "ref-de-proyecto-de-prueba"
	)

	// Se captura EL CUERPO REAL RECIBIDO POR EL SERVIDOR, no lo que el test cree que se envió.
	// Derivarlo de la struct dejaría fuera precisamente lo que se quiere vigilar: lo que sale de la
	// máquina. Es el mismo criterio que el golden de la ingesta aplica a sus tres caminos.
	var recibido []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recibido, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cliente := New(srv.URL, "tok-de-prueba")
	cliente.HTTP = srv.Client() // confía en el certificado del servidor de test

	// El desenlace es de T012 y aquí da igual: lo que se vigila es EL CUERPO, que ya se emite.
	_, _ = cliente.Adherir(codigo, projectRef)

	if len(recibido) == 0 {
		t.Fatalf("premisa rota: el servidor no recibió cuerpo alguno; sin cuerpo no hay frontera que vigilar")
	}

	// Se decodifica a un mapa genérico, NO a `peticionAdhesion`: decodificar a la struct del propio
	// código haría que el test viera lo que la struct admite en vez de lo que el cuerpo lleva, y un
	// campo de más sería invisible.
	var cuerpo map[string]any
	if err := json.Unmarshal(recibido, &cuerpo); err != nil {
		t.Fatalf("el cuerpo emitido no es JSON de objeto: %v\n  crudo: %s", err, recibido)
	}

	// ═══ LA ASERCIÓN: el conjunto de claves, EXACTO ═══
	var claves []string
	for k := range cuerpo {
		claves = append(claves, k)
	}
	sort.Strings(claves)

	permitidas := append([]string(nil), clavesPermitidasAdhesion...)
	sort.Strings(permitidas)

	if strings.Join(claves, ",") != strings.Join(permitidas, ",") {
		t.Errorf("el cuerpo de adhesión NO lleva exactamente la allowlist de dos elementos:\n"+
			"  claves emitidas: %v\n"+
			"  allowlist:       %v\n"+
			"  Principio I: ningún dato de la instalación viaja por esta puerta salvo estos dos",
			claves, permitidas)
	}

	// Y que los dos valores sean los que se le dieron: una allowlist correcta con los valores
	// cruzados seguiría siendo una fuga, sólo que ordenada.
	if cuerpo["code"] != codigo {
		t.Errorf("code = %v, want %q", cuerpo["code"], codigo)
	}
	if cuerpo["project_ref"] != projectRef {
		t.Errorf("project_ref = %v, want %q", cuerpo["project_ref"], projectRef)
	}
}
