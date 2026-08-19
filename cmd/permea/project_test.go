package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ P-005 Phase 6 · LOS TESTS DEL COMANDO `permea project join` ══════════════════════════
//
// Contrato: `specs/005-adhesion-a-proyecto/contracts/cli.md`. **Los pone en verde P-005 T021**
// (entrada y rehúses) y **T027** (presentación); aquí nacen **rojos** contra el andamiaje de T003,
// que rehúsa siempre y **sale con 70** —un código que el contrato no usa, precisamente para que estos
// tests no puedan acertar contra él—.
//
// ⛔ **Tests de PROCESO**: se compara `ExitCode()`, nunca texto de mensajes (disciplina 4). Donde hace
// falta distinguir dos rehúses —T018— **se comparan las salidas ENTRE SÍ**, que no depende de ninguna
// redacción ni del idioma del sistema.

// ─────────────────────────────────────────────────────────────────────────────────────────
// Andamiaje común
// ─────────────────────────────────────────────────────────────────────────────────────────

// bancoDeAdhesion es un destino instrumentado: cuenta las peticiones que REALMENTE le llegan.
type bancoDeAdhesion struct {
	mu         sync.Mutex
	peticiones int
	srv        *httptest.Server
}

func (b *bancoDeAdhesion) recibidas() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.peticiones
}

// entornoDeAdhesion monta el aislamiento (disciplina 6), un árbol de trabajo con raíz reconocible,
// un destino HTTPS instrumentado y una `config.json` que apunta a él.
//
// El binario hijo confía en el certificado del banco por `SSL_CERT_FILE`, que en Go es **aditivo**
// —no sustituye el almacén del sistema—. *(Linux/macOS: en Windows Go usa el almacén del sistema y
// esta variable se ignora, así que los casos que necesitan alcanzar el banco no aplican allí.)*
func entornoDeAdhesion(t *testing.T, denominacion string) (dataDir, arbol string, banco *bancoDeAdhesion) {
	t.Helper()
	return entornoConDesenlace(t, http.StatusOK, `{"project":{"name":"`+denominacion+`"}}`)
}

// entornoConDesenlace es `entornoDeAdhesion` con **el desenlace remoto elegido**: el banco responde
// `estado` y `cuerpo`, que es lo que distingue D1, D2 y D3/D4 (`contracts/adhesion.md` §Los cuatro
// desenlaces — el estado es el discriminante, el cuerpo confirma).
//
// El montaje es UNO SOLO a propósito: si cada familia de tests armara el suyo, «las mismas
// condiciones» dejaría de significar lo mismo en cada sitio sin que nadie lo notara.
func entornoConDesenlace(t *testing.T, estado int, cuerpo string) (dataDir, arbol string, banco *bancoDeAdhesion) {
	t.Helper()

	dataDir = testutil.Sandbox(t)
	banco = &bancoDeAdhesion{}
	banco.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		banco.mu.Lock()
		banco.peticiones++
		banco.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(estado)
		_, _ = io.WriteString(w, cuerpo)
	}))
	t.Cleanup(banco.srv.Close)

	// El certificado del banco, en un fichero que el hijo pueda leer.
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: banco.srv.Certificate().Raw})
	rutaCert := filepath.Join(t.TempDir(), "banco.pem")
	if err := os.WriteFile(rutaCert, pemCert, 0o600); err != nil {
		t.Fatalf("escribir el certificado del banco: %v", err)
	}
	t.Setenv("SSL_CERT_FILE", rutaCert)

	escribirConfig(t, dataDir, banco.srv.URL+rutaDeIngesta)

	// Árbol de trabajo con raíz reconocible: un `.git`, que es el marcador que usa el resolutor.
	arbol = filepath.Join(t.TempDir(), "proyecto")
	if err := os.MkdirAll(filepath.Join(arbol, ".git"), 0o755); err != nil {
		t.Fatalf("crear el árbol de trabajo: %v", err)
	}
	return dataDir, arbol, banco
}

// rutaDeIngesta es la ruta del endpoint que el enrolamiento persiste. La forma la fija
// `contracts/adhesion.md` §Cómo se obtiene `<base>`: el destino de la adhesión se DERIVA de ella
// sustituyendo el último segmento, así que **esta ruta es la forma BUENA** y su contraria
// —`endpointDeFormaInesperada`— es la del tercer rehúse.
const rutaDeIngesta = "/api/v1/ingest"

// endpointDeFormaInesperada es un endpoint **bien formado como URL y https**, cuya ruta **no termina
// en el segmento de ingesta**: no permite derivar el destino de la adhesión, que es la condición del
// tercer rehúse (R3, P-005 FR-009).
//
// ⛔ **No es un endpoint inválido ni en claro**, y la distinción es el test: un endpoint no-https lo
// rehúsa la guarda de esquema del transporte —otro camino, otro mensaje—, así que montarlo así
// mediría un orden que no es el de este test.
const endpointDeFormaInesperada = "https://api.permea.example/api/v1/eventos"

// escribirConfig deja una `config.json` bien formada apuntando a `endpoint`, **con device_token**:
// una instalación enrolada.
func escribirConfig(t *testing.T, dataDir, endpoint string) {
	t.Helper()
	escribirConfigCon(t, dataDir, endpoint, tokenDePrueba)
}

// escribirConfigSinToken deja una `config.json` que **EXISTE y no tiene `device_token`**.
//
// Es la representación de «sin enrolamiento» que **coexiste** con las otras dos condiciones del
// orden — el motivo largo está en `TestProjectJoin_ElOrdenDeLosTresRehuses`, y no es un detalle de
// montaje: es lo que hace que la pareja enrolamiento↔configuración pueda existir.
func escribirConfigSinToken(t *testing.T, dataDir, endpoint string) {
	t.Helper()
	escribirConfigCon(t, dataDir, endpoint, "")
}

// tokenDePrueba es el device_token de una instalación enrolada. Valor inventado, nunca un secreto
// real: lo único que se le pide es no estar vacío.
const tokenDePrueba = "0123456789abcdef0123456789abcdef"

// escribirConfigCon es el cuerpo compartido de las dos anteriores: **una sola forma de config**, y
// el token como único eje que las separa. Dos escrituras independientes serían dos formas que
// pueden divergir, y entonces «lo único que cambia es el enrolamiento» dejaría de ser cierto sin
// que nadie lo notara.
func escribirConfigCon(t *testing.T, dataDir, endpoint, token string) {
	t.Helper()
	logsVacios := filepath.Join(dataDir, "logs-vacios")
	if err := os.MkdirAll(logsVacios, 0o755); err != nil {
		t.Fatalf("crear logs_root vacío: %v", err)
	}
	cfg := map[string]any{
		"endpoint":      endpoint,
		"device_token":  token,
		"org_id":        "org-prueba",
		"dev_id":        "dev-prueba",
		"tools":         []string{"claude_code"},
		"sync_interval": "60s",
		"logs_root":     logsVacios,
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "config.json"), b, 0o600); err != nil {
		t.Fatalf("escribir config: %v", err)
	}
}

// directorioSuelto crea un directorio **sin raíz reconocible por encima**: la condición del primer
// rehúse (R1, P-005 FR-006). Cuelga de un temporal del test, así que el ascenso llega a la raíz del
// sistema sin encontrar marcador.
func directorioSuelto(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "suelto")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("crear directorio suelto: %v", err)
	}
	return dir
}

// desenlace es lo observable de una ejecución: los tres canales por separado (disciplina 7).
type desenlace struct {
	codigo int
	stdout string
	stderr string
	expiro bool
}

// ejecutarEn corre el binario en `dir`, con `entrada` por la entrada estándar. `entrada == nil`
// significa **sin pipe**: el hijo recibe una entrada que da EOF de inmediato, que es lo más cerca de
// «terminal sin pipe» que se puede montar de forma portable — y basta, porque lo que se comprueba es
// que **no se cuelga esperando** y que sale con el código de uso.
func ejecutarEn(t *testing.T, dir string, entrada *string, limite time.Duration, args ...string) desenlace {
	t.Helper()

	ctx, cancelar := context.WithTimeout(context.Background(), limite)
	defer cancelar()

	cmd := exec.CommandContext(ctx, binarioDePrueba, args...)
	cmd.Dir = dir
	if entrada != nil {
		cmd.Stdin = strings.NewReader(*entrada)
	}
	var salida, errores bytes.Buffer
	cmd.Stdout = &salida
	cmd.Stderr = &errores
	err := cmd.Run()

	d := desenlace{stdout: salida.String(), stderr: errores.String(), expiro: ctx.Err() != nil}
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			d.codigo = ee.ExitCode()
		} else {
			d.codigo = -1
		}
	}
	return d
}

func conEntrada(s string) *string { return &s }

const codigoDeAdhesion = "pmeaj1.9HpQ3mZv7KxR2wLbN4tYsE6uJ1cA8dF0gK5rP2xW3zQ"

// ─────────────────────────────────────────────────────────────────────────────────────────
// T016 · LAS DOS VÍAS DE ENTRADA — P-005 FR-023 + SC-011 (A)
// ─────────────────────────────────────────────────────────────────────────────────────────

// mismoDesenlace compara dos ejecuciones por sus TRES canales, por separado.
func mismoDesenlace(a, b desenlace) bool {
	return a.codigo == b.codigo && a.stdout == b.stdout && a.stderr == b.stderr
}

// TestProjectJoin_LasDosViasSonIndistinguibles cubre P-005 FR-023: el mismo código por argumento y
// por entrada estándar produce **el mismo desenlace**, y **la vía elegida NUNCA es observable**.
func TestProjectJoin_LasDosViasSonIndistinguibles(t *testing.T) {
	_, arbol, _ := entornoDeAdhesion(t, "Plataforma Permea")

	porArgumento := ejecutarEn(t, arbol, nil, 20*time.Second, "project", "join", codigoDeAdhesion)
	porEntrada := ejecutarEn(t, arbol, conEntrada(codigoDeAdhesion+"\n"), 20*time.Second, "project", "join")

	// PIEZA 1 — no vacío y DEL TIPO QUE TOCA. Sin esto, dos fracasos idénticos pasarían por éxito
	// idéntico: «iguales» no dice nada si los dos son el mismo error.
	for _, c := range []struct {
		via string
		d   desenlace
	}{{"argumento", porArgumento}, {"entrada estándar", porEntrada}} {
		if c.d.codigo != 0 {
			t.Errorf("vía %s: ExitCode() = %d, se esperaba 0 (unión conseguida)", c.via, c.d.codigo)
		}
		if strings.TrimSpace(c.d.stdout) == "" {
			t.Errorf("vía %s: stdout VACÍO; el éxito comunica la denominación por stdout", c.via)
		}
	}

	// PIEZA 2 — IDÉNTICOS ENTRE SÍ, canal a canal.
	if !mismoDesenlace(porArgumento, porEntrada) {
		t.Errorf("la vía elegida ES observable en la salida (P-005 FR-023):\n"+
			"  argumento: código=%d stdout=%q stderr=%q\n"+
			"  entrada:   código=%d stdout=%q stderr=%q",
			porArgumento.codigo, porArgumento.stdout, porArgumento.stderr,
			porEntrada.codigo, porEntrada.stdout, porEntrada.stderr)
	}

	// PIEZA 3 — LA COMPARACIÓN SABE FALLAR. Sin esto, un `mismoDesenlace` que devolviera siempre
	// `true` dejaría la pieza 2 en verde para siempre, y nadie lo notaría.
	for _, distinto := range []desenlace{
		{codigo: 1, stdout: porArgumento.stdout, stderr: porArgumento.stderr},
		{codigo: porArgumento.codigo, stdout: porArgumento.stdout + "x", stderr: porArgumento.stderr},
		{codigo: porArgumento.codigo, stdout: porArgumento.stdout, stderr: porArgumento.stderr + "x"},
	} {
		if mismoDesenlace(porArgumento, distinto) {
			t.Errorf("la comparación NO sabe fallar: dio iguales %+v y %+v", porArgumento, distinto)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T017 · ENTRADA AUSENTE — error de uso, y NUNCA un prompt que se cuelgue
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestProjectJoin_SinArgumentoYSinEntrada cubre la tercera fila de `cli.md` §Entrada.
//
// ⛔ **El código es EXACTO, no «≠ 0»**: con «distinto de cero» el `70` del andamiaje de T003
// satisfaría el test **sin que nadie hubiera implementado nada** —y `cli.md` §Los códigos de salida
// fija el valor con un número precisamente por esto—.
func TestProjectJoin_SinArgumentoYSinEntrada(t *testing.T) {
	_, arbol, _ := entornoDeAdhesion(t, "Plataforma Permea")

	d := ejecutarEn(t, arbol, nil, 10*time.Second, "project", "join")

	if d.expiro {
		t.Errorf("el comando NO terminó: se quedó esperando entrada. NUNCA un prompt que se cuelgue")
	}
	if d.codigo != 1 {
		t.Errorf("ExitCode() = %d, se esperaba EXACTAMENTE 1 (error de uso)", d.codigo)
	}
	if strings.TrimSpace(d.stderr) == "" {
		t.Errorf("stderr VACÍO: el error de uso se comunica por stderr")
	}
	if d.stdout != "" {
		t.Errorf("stdout = %q, se esperaba VACÍO: en un error no hay respuesta que dar (P-005 FR-021)", d.stdout)
	}
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T017b · EL ENTRELAZADO ENTRADA ↔ REHÚSES — sin argumento Y fuera de árbol
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestProjectJoin_LaEntradaGanaAlRehuseDelArbol fija **el único caso que observa el entrelazado**
// entre la resolución de la entrada y los tres rehúses locales: **sin argumento y fuera de árbol**.
//
// ⛔ **Gana el error de uso.** Rehusar por falta de árbol cuando la persona **no ha dado ningún
// código** es contestar a una pregunta que no hizo: se le señala un sitio al que ir con algo que
// todavía no tiene. El error de uso le dice lo que le falta **ahora**, que es el criterio con el que
// D-005-P13 ordena los tres —el error que se puede corregir ahora— aplicado a lo que **no** es uno de
// los tres (`contracts/cli.md` deja el error de uso fuera de la tabla a propósito).
//
// ⛔ **No se casa texto** (disciplina 4): se comparan **las salidas entre sí**, igual que T018. El
// `stderr` de «sin argumento y fuera de árbol» debe ser **el mismo** que el de «sin argumento dentro
// de un árbol» —el error de uso puro— y **distinto** del de «con código y fuera de árbol» —el rehúse
// del árbol puro—.
//
// ⚠️ **NACE VERDE**: P-005 T021 ya resolvió la entrada antes de los tres rehúses. Se valida **por
// mutación** —mover la lectura de la entrada detrás de los rehúses—, y la transcripción del fallo
// está en `tasks.md` T017b.
func TestProjectJoin_LaEntradaGanaAlRehuseDelArbol(t *testing.T) {
	// El error de uso PURO: dentro de un árbol, con enrolamiento bueno, sin argumento y sin pipe.
	usoPuro := func(t *testing.T) desenlace {
		_, arbol, _ := entornoDeAdhesion(t, "X")
		return ejecutarEn(t, arbol, nil, 10*time.Second, "project", "join")
	}
	// El rehúse del árbol PURO: fuera de árbol, pero con el código presentado.
	arbolPuro := func(t *testing.T) desenlace {
		entornoDeAdhesion(t, "X")
		return ejecutarEn(t, directorioSuelto(t), nil, 10*time.Second, "project", "join", codigoDeAdhesion)
	}
	// LA COMBINACIÓN: sin argumento, sin pipe, Y fuera de árbol.
	combinacion := func(t *testing.T) desenlace {
		entornoDeAdhesion(t, "X")
		return ejecutarEn(t, directorioSuelto(t), nil, 10*time.Second, "project", "join")
	}

	obtenido := combinacion(t)
	uso := usoPuro(t)
	arbol := arbolPuro(t)

	if obtenido.expiro {
		t.Errorf("el comando NO terminó: se quedó esperando entrada. NUNCA un prompt que se cuelgue")
	}
	if obtenido.stderr != uso.stderr {
		t.Errorf("sin argumento y fuera de árbol NO gana el error de uso\n"+
			"  obtenido:   %q\n"+
			"  error de uso: %q", obtenido.stderr, uso.stderr)
	}
	if obtenido.stderr == arbol.stderr {
		t.Errorf("el rehúse obtenido es indistinguible del rehúse del árbol: los dos dan %q\n"+
			"  si no se distinguen, el entrelazado no es observable y este test no prueba nada",
			obtenido.stderr)
	}
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T018 · EL ORDEN DE LOS TRES REHÚSES — D-005-P13
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestProjectJoin_ElOrdenDeLosTresRehuses fija el orden **árbol → enrolamiento → configuración**.
//
// ⛔ **Sin este test el orden lo fija el primer camino que alguien escriba**, y el orden es contrato:
// más específico primero, y cada rehúse nombra el error que la persona puede corregir ahora.
//
// ⛔ **Se comprueban LAS CUATRO COMBINACIONES —las tres parejas Y el caso triple—, y no bastan las
// tres condiciones juntas**: «las tres a la vez → el del árbol» **lo pasa igual un orden equivocado
// entre enrolamiento y configuración**, porque el árbol gana de todas formas. La pareja
// **enrolamiento↔configuración** es **la única** que ve ese cambio, y por eso su montaje es el que
// más cuidado necesita (ver la nota de dentro de la función).
//
// ⛔ **No se casa texto** (disciplina 4): se comparan **las salidas entre sí**. El mensaje de una
// combinación tiene que ser **el mismo** que el del rehúse que debe ganar, y **distinto** del de los
// otros. Eso no depende de ninguna redacción.
func TestProjectJoin_ElOrdenDeLosTresRehuses(t *testing.T) {
	// ═══ CÓMO SE REPRESENTA «SIN ENROLAMIENTO» — Y POR QUÉ NO ES «NO HAY config.json» ══════
	//
	// **Es el defecto que este test tuvo al nacer, y se midió**: representar «sin enrolamiento» como
	// **el fichero de configuración NO EXISTE** lo vuelve **mutuamente excluyente** con «configuración
	// de forma inesperada», que exige justo lo contrario —que el fichero **exista**, con un endpoint
	// que no permite derivar el destino—. El montaje de la pareja escribía la configuración rota y
	// **acto seguido borraba el fichero**, así que **la segunda condición se destruía al montar la
	// primera**: `enrolamientoYConfig` era en realidad **sólo «sin enrolamiento»**, y `lasTres` eran
	// **dos**.
	//
	// **Y la que se perdía era precisamente la que importa.** El árbol gana en todas las demás
	// combinaciones, así que **enrolamiento↔configuración es LA ÚNICA que vería un orden invertido
	// entre esos dos**. Sin ella, un test titulado «el ORDEN de los tres rehúses» certificaba **dos
	// parejas de tres** — y en verde no se distingue de uno completo.
	//
	// **La representación que sí coexiste**: la configuración **EXISTE y no tiene `device_token`**.
	// Es «no está enrolada» con la misma literalidad —`enroll` es quien escribe ese campo, y sin él
	// no hay con qué autenticarse—, y **deja sitio** a que el endpoint tenga a la vez la forma
	// inesperada. Con eso las tres condiciones pueden darse **de verdad** a la vez, que es el
	// supuesto del que parte D-005-P13.
	//
	// ⚠️ **Las dos representaciones caen en el mismo rehúse, y por eso el cambio es seguro**: un
	// fichero ausente carga la configuración por defecto —endpoint y token vacíos— y una config sin
	// `device_token` deja el token vacío. Las dos son «no enrolada» para la misma comprobación, así
	// que las combinaciones que ya montaba el fichero ausente siguen valiendo.

	soloArbol := func(t *testing.T) string {
		entornoDeAdhesion(t, "X")
		return directorioSuelto(t)
	}
	soloEnrolamiento := func(t *testing.T) string {
		dataDir, arbol, banco := entornoDeAdhesion(t, "X")
		// Se conserva el endpoint del banco —forma BUENA— y se retira sólo el token: aquí no hay
		// nada roto salvo el enrolamiento.
		escribirConfigSinToken(t, dataDir, banco.srv.URL+rutaDeIngesta)
		return arbol
	}
	soloConfig := func(t *testing.T) string {
		dataDir, arbol, _ := entornoDeAdhesion(t, "X")
		escribirConfig(t, dataDir, endpointDeFormaInesperada)
		return arbol
	}
	arbolYEnrolamiento := func(t *testing.T) string {
		dataDir, _, banco := entornoDeAdhesion(t, "X")
		escribirConfigSinToken(t, dataDir, banco.srv.URL+rutaDeIngesta)
		return directorioSuelto(t)
	}
	enrolamientoYConfig := func(t *testing.T) string {
		dataDir, arbol, _ := entornoDeAdhesion(t, "X")
		// LAS DOS A LA VEZ, y ahora de verdad: el fichero existe con el endpoint de forma inesperada
		// **y** sin device_token. Si el orden se invirtiera, ganaría la configuración.
		escribirConfigSinToken(t, dataDir, endpointDeFormaInesperada)
		return arbol
	}
	arbolYConfig := func(t *testing.T) string {
		dataDir, _, _ := entornoDeAdhesion(t, "X")
		escribirConfig(t, dataDir, endpointDeFormaInesperada)
		return directorioSuelto(t)
	}
	lasTres := func(t *testing.T) string {
		dataDir, _, _ := entornoDeAdhesion(t, "X")
		escribirConfigSinToken(t, dataDir, endpointDeFormaInesperada)
		return directorioSuelto(t)
	}

	corre := func(t *testing.T, montar func(*testing.T) string) desenlace {
		t.Helper()
		dir := montar(t)
		return ejecutarEn(t, dir, conEntrada(codigoDeAdhesion+"\n"), 20*time.Second, "project", "join")
	}

	casos := []struct {
		nombre      string
		combinacion func(*testing.T) string
		debeGanar   func(*testing.T) string
		nombreGana  string
		noDebeSer   []struct {
			nombre string
			montar func(*testing.T) string
		}
	}{
		{"PAREJA árbol + enrolamiento", arbolYEnrolamiento, soloArbol, "el del árbol",
			[]struct {
				nombre string
				montar func(*testing.T) string
			}{{"el del enrolamiento", soloEnrolamiento}}},
		{"PAREJA enrolamiento + configuración", enrolamientoYConfig, soloEnrolamiento, "el del enrolamiento",
			[]struct {
				nombre string
				montar func(*testing.T) string
			}{{"el de la configuración", soloConfig}}},
		{"PAREJA árbol + configuración", arbolYConfig, soloArbol, "el del árbol",
			[]struct {
				nombre string
				montar func(*testing.T) string
			}{{"el de la configuración", soloConfig}}},
		{"LAS TRES a la vez", lasTres, soloArbol, "el del árbol",
			[]struct {
				nombre string
				montar func(*testing.T) string
			}{{"el del enrolamiento", soloEnrolamiento}, {"el de la configuración", soloConfig}}},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			obtenido := corre(t, c.combinacion)
			gana := corre(t, c.debeGanar)

			if obtenido.stderr != gana.stderr {
				t.Errorf("el rehúse que gana NO es %s (D-005-P13, orden árbol → enrolamiento → configuración)\n"+
					"  obtenido: %q\n"+
					"  %s:  %q", c.nombreGana, obtenido.stderr, c.nombreGana, gana.stderr)
			}
			for _, otro := range c.noDebeSer {
				perdedor := corre(t, otro.montar)
				if obtenido.stderr == perdedor.stderr {
					t.Errorf("el rehúse obtenido es indistinguible de %s: los dos dan %q\n"+
						"  si los mensajes no se distinguen, el orden no es observable y este test no prueba nada",
						otro.nombre, obtenido.stderr)
				}
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T019 · CERO PETICIONES FUERA DE ÁRBOL — SC-004 + P-005 FR-006
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestProjectJoin_FueraDeArbolNoEmiteNiUnaPeticion usa **un destino instrumentado que CUENTA**.
//
// ⛔ **Con su CASO POSITIVO, y es la mitad que importa**: «cero peticiones» lo cumple igual un comando
// que **no emite nunca**, así que sin el positivo el test no distingue «rehusó antes de emitir» de «no
// sabe emitir». El positivo usa **EL MISMO destino** y exige **exactamente una**.
func TestProjectJoin_FueraDeArbolNoEmiteNiUnaPeticion(t *testing.T) {
	t.Run("fuera de árbol → CERO peticiones", func(t *testing.T) {
		_, _, banco := entornoDeAdhesion(t, "Plataforma Permea")
		sinArbol := filepath.Join(t.TempDir(), "suelto")
		if err := os.MkdirAll(sinArbol, 0o755); err != nil {
			t.Fatalf("crear directorio suelto: %v", err)
		}

		d := ejecutarEn(t, sinArbol, conEntrada(codigoDeAdhesion+"\n"), 20*time.Second, "project", "join")

		if d.codigo == 0 {
			t.Errorf("ExitCode() = 0 fuera de un árbol de trabajo: debe rehusar (P-005 FR-006)")
		}
		if n := banco.recibidas(); n != 0 {
			t.Errorf("el destino recibió %d peticiones; el rehúse ocurre ANTES de emitir (SC-004)", n)
		}
	})

	t.Run("CASO POSITIVO · dentro de árbol → EXACTAMENTE UNA", func(t *testing.T) {
		_, arbol, banco := entornoDeAdhesion(t, "Plataforma Permea")

		d := ejecutarEn(t, arbol, conEntrada(codigoDeAdhesion+"\n"), 20*time.Second, "project", "join")

		if d.codigo != 0 {
			t.Errorf("ExitCode() = %d dentro de un árbol con enrolamiento y destino vivo; se esperaba 0", d.codigo)
		}
		if n := banco.recibidas(); n != 1 {
			t.Errorf("el destino recibió %d peticiones, se esperaba EXACTAMENTE 1 "+
				"(un solo intento, sin cola ni reintento — P-005 FR-018)", n)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T020 · VERBO DESCONOCIDO Y `project` SIN VERBO
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestProject_ErroresDeUsoDeLaGramatica cubre `cli.md` §La gramática.
//
// ⛔ **El código es EXACTO**: con «≠ 0» el `70` del andamiaje lo satisface.
func TestProject_ErroresDeUsoDeLaGramatica(t *testing.T) {
	casos := []struct {
		nombre string
		args   []string
	}{
		{"`project` sin verbo", []string{"project"}},
		{"verbo desconocido", []string{"project", "adherirse"}},
		{"verbo que se parece al bueno — NUNCA se corrige", []string{"project", "joln"}},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, arbol, _ := entornoDeAdhesion(t, "X")

			d := ejecutarEn(t, arbol, nil, 10*time.Second, c.args...)

			if d.codigo != 1 {
				t.Errorf("ExitCode() = %d, se esperaba EXACTAMENTE 1 (error de uso)", d.codigo)
			}
			if strings.TrimSpace(d.stderr) == "" {
				t.Errorf("stderr VACÍO: el error de uso se comunica por stderr, nombrando lo no reconocido")
			}
			if d.stdout != "" {
				t.Errorf("stdout = %q, se esperaba VACÍO (P-005 FR-021)", d.stdout)
			}
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════════════════
// P-005 Phase 7 · CANALES, SALIDAS Y SECRETOS — T022, T023, T024, T025, T026
// ═════════════════════════════════════════════════════════════════════════════════════════
//
// **Los OCHO desenlaces del COMANDO**, que no son los cuatro de la plataforma: los cuatro remotos
// —D1..D4— más los **tres rehúses locales** y el **no verificable**, que nunca llegan a la
// plataforma (`contracts/cli.md` §Comportamiento y salidas).
//
// La tabla es **una sola** y la comparten los cuatro tests de esta fase. Si cada uno armara la suya,
// «los ocho» significaría ocho cosas distintas en cuatro sitios, y la primera que se quedara en siete
// no la vería nadie.

// ejecucion es todo lo que hace falta para provocar un desenlace, más el directorio de datos que ese
// montaje ha resuelto —que es lo que inspeccionan T025 y T026—.
type ejecucion struct {
	dataDir string
	dir     string
	entrada *string
	args    []string
}

// desenlaceDelComando describe uno de los ocho: cómo se provoca y qué promete el contrato de él.
type desenlaceDelComando struct {
	nombre  string
	esExito bool // stdout es la respuesta (`contracts/cli.md` §El reparto de canales)
	codigo  int  // `contracts/cli.md` §Los códigos de salida
	montar  func(t *testing.T) ejecucion
}

// entornoDeDesenlace es `entornoConDesenlace` **con los secretos ya sembrados**.
//
// ⛔ **El sembrado NO es un apaño del arnés: es un ESPEJO de lo que deja el enrolamiento.** El
// montaje escribe `config.json` a mano —no pasa por `enroll`—, así que tiene que reponer lo demás.
// Y lo demás es el `salt`, que **`enroll` persiste al terminar un enrolamiento correcto**
// (`cmd/permea/enroll.go`) precisamente para que `project join` no tenga nada que crear: el salt es
// la semilla de `event.Ref` y hace falta para componer `project_ref` **antes de emitir**.
//
// **Medido con binario real** (`tasks.md` T025): sobre una instalación enrolada de verdad, los
// cuatro caminos que emiten —éxito, `422`, `409` y servidor caído— y el rehúse local dejan el
// directorio de datos **sin una sola diferencia**. Sembrar aquí reproduce ese estado, no lo disimula.
func entornoDeDesenlace(t *testing.T, estado int, cuerpo string) (dataDir, arbol string, banco *bancoDeAdhesion) {
	t.Helper()
	dataDir, arbol, banco = entornoConDesenlace(t, estado, cuerpo)
	sembrarSecretos(t, dataDir)
	return dataDir, arbol, banco
}

// sembrarSecretos deja `salt` y `machine_id` con los permisos que usa el agente. Valores fijos e
// inventados: lo único que se les pide es **existir antes** de la ejecución observada.
func sembrarSecretos(t *testing.T, dataDir string) {
	t.Helper()
	for nombre, valor := range map[string]string{
		"salt":       "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90",
		"machine_id": "0f1e2d3c4b5a69788796a5b4c3d2e1f0",
	} {
		if err := os.WriteFile(filepath.Join(dataDir, nombre), []byte(valor), 0o600); err != nil {
			t.Fatalf("sembrar %s: %v", nombre, err)
		}
	}
}

// losOchoDesenlaces enumera los ocho, en el orden de la tabla del contrato.
//
// **Todos presentan el código por ARGUMENTO**: la vía no es observable en el resultado
// (P-005 FR-023, verificado por T016), así que fijar una deja los ocho montajes comparables.
func losOchoDesenlaces() []desenlaceDelComando {
	conCodigo := []string{"project", "join", codigoDeAdhesion}
	exito := `{"project":{"name":"RecetApp"}}`

	return []desenlaceDelComando{
		{"R1 · fuera de un árbol de trabajo", false, 1, func(t *testing.T) ejecucion {
			dataDir, _, _ := entornoDeDesenlace(t, http.StatusOK, exito)
			return ejecucion{dataDir, directorioSuelto(t), nil, conCodigo}
		}},
		{"R2 · sin enrolamiento", false, 1, func(t *testing.T) ejecucion {
			dataDir, arbol, banco := entornoDeDesenlace(t, http.StatusOK, exito)
			escribirConfigSinToken(t, dataDir, banco.srv.URL+rutaDeIngesta)
			return ejecucion{dataDir, arbol, nil, conCodigo}
		}},
		{"R3 · configuración de forma inesperada", false, 1, func(t *testing.T) ejecucion {
			dataDir, arbol, _ := entornoDeDesenlace(t, http.StatusOK, exito)
			escribirConfig(t, dataDir, endpointDeFormaInesperada)
			return ejecucion{dataDir, arbol, nil, conCodigo}
		}},
		{"D4 · unión nueva", true, 0, func(t *testing.T) ejecucion {
			dataDir, arbol, _ := entornoDeDesenlace(t, http.StatusOK, exito)
			return ejecucion{dataDir, arbol, nil, conCodigo}
		}},
		{"D3 · ya unido a ESE MISMO Proyecto", true, 0, func(t *testing.T) ejecucion {
			dataDir, arbol, _ := entornoDeDesenlace(t, http.StatusOK, exito)
			// LA SEGUNDA PRESENTACIÓN del mismo código. La plataforma declara D3 y D4
			// indistinguibles (`contracts/adhesion.md` §Los cuatro desenlaces), así que el banco
			// responde lo mismo — y **eso es el contrato, no una simplificación del arnés**.
			ejecutarEn(t, arbol, nil, 20*time.Second, conCodigo...)
			return ejecucion{dataDir, arbol, nil, conCodigo}
		}},
		{"D2 · ya unido a OTRO Proyecto", false, 1, func(t *testing.T) ejecucion {
			dataDir, arbol, _ := entornoDeDesenlace(t, http.StatusConflict, `{}`)
			return ejecucion{dataDir, arbol, nil, conCodigo}
		}},
		{"D1 · código no utilizable", false, 1, func(t *testing.T) ejecucion {
			dataDir, arbol, _ := entornoDeDesenlace(t, http.StatusUnprocessableEntity, `{}`)
			return ejecucion{dataDir, arbol, nil, conCodigo}
		}},
		{"NV · no verificable (servidor inalcanzable)", false, 1, func(t *testing.T) ejecucion {
			dataDir, arbol, banco := entornoDeDesenlace(t, http.StatusOK, exito)
			// Se cierra el banco DESPUÉS de que la config apunte a él: el destino queda escrito y
			// muerto, que es exactamente «servidor inalcanzable». `Close` es idempotente, así que el
			// `t.Cleanup` del montaje sigue siendo correcto.
			banco.srv.Close()
			return ejecucion{dataDir, arbol, nil, conCodigo}
		}},
	}
}

// correrDesenlace monta y ejecuta uno de los ocho.
func correrDesenlace(t *testing.T, d desenlaceDelComando) (ejecucion, desenlace) {
	t.Helper()
	e := d.montar(t)
	return e, ejecutarEn(t, e.dir, e.entrada, 20*time.Second, e.args...)
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T022 · EL REPARTO DE CANALES — P-005 FR-021 + SC-011 (B)
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestProjectJoin_ElRepartoDeCanales comprueba, **para los ocho desenlaces**, que stdout es la
// respuesta y stderr es todo lo demás.
//
// ⛔ **Los dos canales se capturan POR SEPARADO** (disciplina 7, SC-011 B): `desenlace` los guarda en
// campos distintos y nunca se concatenan. Una salida combinada no vacía es compatible con **cualquier**
// reparto, incluido el equivocado, así que un montaje que los una **no cuenta como pasado**.
//
// ⛔ **«stderr sin el desenlace» se comprueba por CANAL VACÍO, no buscando el texto** (disciplina 5).
// Buscar que la denominación no aparezca pasa igual **cuando aparece otro texto distinto**, y entonces
// el criterio no distingue «no se coló el desenlace» de «se coló otra cosa». El canal vacío satisface
// «sin el desenlace» y además es observable.
func TestProjectJoin_ElRepartoDeCanales(t *testing.T) {
	for _, d := range losOchoDesenlaces() {
		t.Run(d.nombre, func(t *testing.T) {
			_, r := correrDesenlace(t, d)

			if d.esExito {
				if strings.TrimSpace(r.stdout) == "" {
					t.Errorf("stdout VACÍO en un desenlace de ÉXITO: stdout es la respuesta (P-005 FR-021)")
				}
				if r.stderr != "" {
					t.Errorf("stderr = %q en un desenlace de ÉXITO; se esperaba VACÍO: el desenlace no viaja "+
						"por el canal de los avisos (SC-011 B)", r.stderr)
				}
				return
			}
			if strings.TrimSpace(r.stderr) == "" {
				t.Errorf("stderr VACÍO en un desenlace de REHÚSE o ERROR: por ahí se comunica (P-005 FR-021)")
			}
			if r.stdout != "" {
				t.Errorf("stdout = %q en un desenlace de REHÚSE o ERROR; se esperaba VACÍO: no hay respuesta "+
					"que dar, y quien canalice stdout a un fichero no debe encontrarse un error dentro", r.stdout)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T023 · LOS OCHO CÓDIGOS DE SALIDA — `contracts/cli.md` §Los códigos de salida
// ─────────────────────────────────────────────────────────────────────────────────────────

// TestProjectJoin_LosOchoCodigosDeSalida compara **`ExitCode()`, nunca texto** (disciplina 4).
//
// ⛔ **El caso que no puede faltar es D3 ≡ D4.** El resultado del proceso es **observable** —es lo
// primero que mira un script—, así que darle a «ya estabas unido» un código propio **rompería
// P-005 FR-010** sin que nadie lo notara: quien pega un código dos veces sabría cuál de las dos surtió
// efecto, que es justo lo que la plataforma declara indistinguible a propósito. Este test es lo único
// que lo impide, y por eso los compara **entre sí** además de contra su valor.
//
// ⛔ **Y es el RESPALDO que retira cualquier código de andamiaje superviviente**: comparar los ocho
// valores exactos tumba cualquier `70` que quedara vivo. La retirada la hizo T021; esto la vigila.
func TestProjectJoin_LosOchoCodigosDeSalida(t *testing.T) {
	for _, d := range losOchoDesenlaces() {
		t.Run(d.nombre, func(t *testing.T) {
			_, r := correrDesenlace(t, d)
			if r.codigo != d.codigo {
				t.Errorf("ExitCode() = %d, se esperaba EXACTAMENTE %d. El binario tiene DOS códigos, "+
					"`0` y `1`, y esta feature no amplía el vocabulario", r.codigo, d.codigo)
			}
		})
	}

	t.Run("D3 y D4 COMPARTEN CÓDIGO — P-005 FR-010", func(t *testing.T) {
		var d3, d4 desenlaceDelComando
		for _, d := range losOchoDesenlaces() {
			switch {
			case strings.HasPrefix(d.nombre, "D3"):
				d3 = d
			case strings.HasPrefix(d.nombre, "D4"):
				d4 = d
			}
		}
		if d3.montar == nil || d4.montar == nil {
			t.Fatalf("la tabla de los ocho desenlaces ya no trae D3 y D4: no hay nada que comparar")
		}

		_, rD4 := correrDesenlace(t, d4)
		_, rD3 := correrDesenlace(t, d3)

		if rD3.codigo != rD4.codigo {
			t.Errorf("D3 y D4 salen con códigos DISTINTOS (%d y %d): el resultado del proceso revela "+
				"cuál de las dos presentaciones surtió efecto, y P-005 FR-010 lo prohíbe", rD3.codigo, rD4.codigo)
		}

		// P-005 T027 · **LA SALIDA ENTERA, BYTE A BYTE.** P-005 FR-010 exige «mismo texto, mismo canal
		// y mismo resultado del proceso», y el código es sólo el tercero. Se compara con el mismo
		// `mismoDesenlace` que T016 —los tres canales por separado—, cuya capacidad de fallar ya
		// demuestra allí la pieza 3. Aserción aparte y con `t.Errorf`: encadenada tras la del código,
		// una mutación que tumbara aquélla la dejaría sin evaluar (disciplina 3 §inmunidad).
		if !mismoDesenlace(rD3, rD4) {
			t.Errorf("D3 y D4 producen SALIDAS DISTINTAS, y P-005 FR-010 las exige indistinguibles:\n"+
				"  D4 (unión nueva):  código=%d stdout=%q stderr=%q\n"+
				"  D3 (ya unido):     código=%d stdout=%q stderr=%q",
				rD4.codigo, rD4.stdout, rD4.stderr, rD3.codigo, rD3.stdout, rD3.stderr)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T024 · NO FILTRACIÓN DEL CÓDIGO — P-005 FR-020 + SC-005
// ─────────────────────────────────────────────────────────────────────────────────────────

// aparicionesDelCodigo devuelve las subcadenas de longitud `n` de `valor` que aparecen en `texto`.
//
// Es la formulación literal de SC-005: *«se verifica generando las subcadenas de longitud ocho del
// valor y buscándolas todas en la salida: cero apariciones»*. Devuelve **cuáles**, no cuántas: un
// fallo que no dice qué se coló obliga a buscarlo a mano.
func aparicionesDelCodigo(valor string, n int, texto string) []string {
	var vistas []string
	for i := 0; i+n <= len(valor); i++ {
		if trozo := valor[i : i+n]; strings.Contains(texto, trozo) {
			vistas = append(vistas, trozo)
		}
	}
	return vistas
}

// TestProjectJoin_NingunDesenlaceFiltraElCodigo cubre P-005 FR-020 y SC-005 sobre **los ocho
// desenlaces** y **los dos canales**.
//
// ⚠️ **NACE VERDE, y hay que decirlo**: una ausencia la satisface cualquier salida que no nombre el
// código. **Se valida por su CASO POSITIVO, y ése lo ejecuta P-005 T027** — sembrar deliberadamente el
// código en una salida del comando debe tumbar este test, con el mensaje leído. Sin eso, no distingue
// «no se filtra» de «no hay salida que mirar».
func TestProjectJoin_NingunDesenlaceFiltraElCodigo(t *testing.T) {
	const umbral = 8 // SC-005: ocho o más caracteres consecutivos

	// ⛔ PRECONDICIÓN — EL CÓDIGO DEL ARNÉS TIENE QUE SER DE ALTA ENTROPÍA, Y SE COMPRUEBA.
	//
	// Un fixture legible —`codigo-de-prueba`, `AAAAAAAA`— convierte este test en la piedra del
	// «" no perm"`»: una subcadena de ocho que choca por azar con la prosa de los mensajes, y encima
	// sobre los ocho desenlaces a la vez. La forma del contrato (`contracts/adhesion.md` §El código)
	// es `pmeaj1.` + 43 caracteres base64url = **50**, y es lo que se exige aquí para que nadie pueda
	// aflojarlo sin que el test lo diga.
	//
	// `t.Fatalf` es correcto: sin un código conforme, lo que siga **no significa nada** (disciplina 3).
	if !strings.HasPrefix(codigoDeAdhesion, "pmeaj1.") || len(codigoDeAdhesion) != 50 {
		t.Fatalf("el código del arnés no tiene la forma del contrato (`pmeaj1.` + 43 = 50 caracteres): "+
			"len=%d. Un fixture de baja entropía haría este test inútil o ruidoso sobre los ocho desenlaces",
			len(codigoDeAdhesion))
	}

	for _, d := range losOchoDesenlaces() {
		t.Run(d.nombre, func(t *testing.T) {
			_, r := correrDesenlace(t, d)

			// Los dos canales, POR SEPARADO y los dos con `t.Errorf`: si el primero fuera `Fatalf`, una
			// fuga por stdout dejaría stderr sin mirar en esa pasada (disciplina 3 §inmunidad).
			for _, canal := range []struct{ nombre, texto string }{
				{"stdout", r.stdout}, {"stderr", r.stderr},
			} {
				if vistas := aparicionesDelCodigo(codigoDeAdhesion, umbral, canal.texto); len(vistas) > 0 {
					t.Errorf("%s reproduce el código de adhesión: %d subcadena(s) de %d caracteres, "+
						"la primera %q (P-005 FR-020, SC-005)", canal.nombre, len(vistas), umbral, vistas[0])
				}
			}
		})
	}

	t.Run("CASO POSITIVO · el detector sabe encontrar", func(t *testing.T) {
		// P-005 T027 lo puso en marcha. **Es la mitad del instrumento**: una salida fabricada que SÍ
		// lleva el código debe producir apariciones. **La otra mitad —que el detector esté conectado a
		// la salida REAL— no la demuestra ningún subtest**: la demostró la siembra del código en el
		// comando, ejecutada en T027 y transcrita en `tasks.md`. Las dos hacen falta: sin ésta, el
		// detector podría no encontrar nada nunca; sin aquélla, podría estar mirando otra cosa.
		sembrada := "unido al Proyecto, código " + codigoDeAdhesion + "\n"
		if vistas := aparicionesDelCodigo(codigoDeAdhesion, umbral, sembrada); len(vistas) == 0 {
			t.Errorf("el detector NO encuentra el código en una salida que lo lleva entero: "+
				"no distingue «no se filtra» de «no miro». Texto: %q", sembrada)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T025 · NADA SE ESCRIBE EN LOCAL — P-005 FR-019 + SC-007
// ─────────────────────────────────────────────────────────────────────────────────────────

// huellaDelDirectorio captura **íntegro** el contenido de `dir`: ruta relativa → bytes, para TODOS
// los ficheros que haya, recursivamente.
//
// ⛔ **Se DERIVA del directorio, no de una lista escrita a mano.** SC-007 exige «el conjunto completo
// de artefactos locales que la instalación mantiene, enumerado». Una lista a ojo —`config.json`,
// `state.json`, `queue.jsonl`, `salt`— es exactamente el defecto que la disciplina 8 registró tres
// veces: mira **el molde que ya se tenía** en vez de **lo que hay**, y un artefacto nuevo queda fuera
// sin que nadie se entere. Recorrer el directorio caza también **lo que aparece**.
func huellaDelDirectorio(t *testing.T, dir string) map[string]string {
	t.Helper()
	huella := make(map[string]string)
	err := filepath.WalkDir(dir, func(ruta string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(ruta)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, ruta)
		if err != nil {
			return err
		}
		huella[rel] = string(b)
		return nil
	})
	if err != nil {
		t.Fatalf("capturar el directorio de datos %q: %v", dir, err)
	}
	return huella
}

// diferencias describe qué cambió entre dos huellas, nombrando **qué** y de qué clase.
func diferencias(antes, despues map[string]string) []string {
	var d []string
	for ruta, contenido := range despues {
		previo, estaba := antes[ruta]
		switch {
		case !estaba:
			d = append(d, "APARECIÓ "+ruta)
		case previo != contenido:
			d = append(d, "CAMBIÓ "+ruta)
		}
	}
	for ruta := range antes {
		if _, sigue := despues[ruta]; !sigue {
			d = append(d, "DESAPARECIÓ "+ruta)
		}
	}
	sort.Strings(d)
	return d
}

// TestProjectJoin_NoEscribeNadaEnLocal cubre P-005 FR-019 y SC-007 sobre **los ocho desenlaces**.
//
// **Qué se observa**: el directorio de datos **entero**, capturado **antes** y comparado byte a byte
// **contra esa captura previa** —no contra sí mismo—, que es la mitad que SC-007 subraya.
//
// **El conjunto enumerado está COMPLETO y no excluye nada**: configuración, estado, cola y secretos,
// sin cláusula. **Y no hace falta excluir nada porque el `salt` nace en el enrolamiento**
// (`cmd/permea/enroll.go`), no aquí — ver `entornoDeDesenlace` y `tasks.md` T025. Mientras lo creaba
// este comando, lo creaba en **los cuatro caminos que emiten**, incluido un rechazo `422`, y SC-007
// dice «sin modificar» **con cualquier desenlace, incluidos los de rehúse y los de error**: los dos
// textos no podían ser ciertos a la vez. Moviendo el hecho lo son los dos, **sin excepción escrita**.
//
// ⚠️ **NACE VERDE**: una ausencia la satisface un comando que no escribe. **Su CASO POSITIVO se
// ejecuta abajo**, y P-005 T027 añadió la otra mitad: sembrar una escritura en el propio comando y
// leer el rojo (transcrito en `tasks.md` T027).
//
// ⛔ **PUNTO CIEGO CONOCIDO, MEDIDO Y DECLARADO: la fila de D3.** Su montaje **ejecuta el comando una
// vez antes de capturar** —eso es lo que la convierte en «segunda presentación»—, así que un fichero
// que el comando escriba **con contenido constante** ya está dentro de la captura previa y la
// comparación no ve nada. **Medido**: la siembra de T027 tumbó D4, D2, D1 y NV, y **D3 pasó**.
// Se declara en vez de disimularse: la cobertura la dan las **otras cuatro filas**, que sí lo cazan,
// y la fila de D3 sirve a T022/T023/T024 —donde el montaje sí es el correcto— pero **no acredita
// nada aquí**. Cambiar el montaje para taparlo dejaría de medir D3 en las otras tres.
func TestProjectJoin_NoEscribeNadaEnLocal(t *testing.T) {
	for _, d := range losOchoDesenlaces() {
		t.Run(d.nombre, func(t *testing.T) {
			e := d.montar(t)
			antes := huellaDelDirectorio(t, e.dataDir)

			ejecutarEn(t, e.dir, e.entrada, 20*time.Second, e.args...)

			if d := diferencias(antes, huellaDelDirectorio(t, e.dataDir)); len(d) > 0 {
				t.Errorf("el comando modificó el estado local (P-005 FR-019, SC-007):\n  %s",
					strings.Join(d, "\n  "))
			}
		})
	}

	t.Run("CASO POSITIVO · el observador sabe ver un cambio", func(t *testing.T) {
		// P-005 T027 lo puso en marcha. Una operación de la instalación que SÍ escribe en local:
		// `--run` genera y encola, así que deja `state.json`. **La MISMA comparación** —`diferencias`
		// sobre `huellaDelDirectorio`, la de arriba— tiene que verlo. Si no lo viera, el verde de
		// arriba no distinguiría «no cambió» de «no miré».
		dataDir, arbol, _ := entornoDeDesenlace(t, http.StatusOK, `{"project":{"name":"RecetApp"}}`)
		antes := huellaDelDirectorio(t, dataDir)

		ejecutarEn(t, arbol, nil, 20*time.Second, "--run")

		if d := diferencias(antes, huellaDelDirectorio(t, dataDir)); len(d) == 0 {
			t.Errorf("una operación que SÍ escribe en local no movió la comparación: el observador no " +
				"está mirando, y entonces el verde de arriba no distingue «no cambió» de «no miré»")
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────────────────
// T026 · LA PETICIÓN NUNCA SE ENCOLA — P-005 FR-018 + SC-010
// ─────────────────────────────────────────────────────────────────────────────────────────

// colaDe devuelve los eventos pendientes en la cola de `dataDir`. Una cola inexistente son cero
// eventos, sin error — que es el estado de partida de una instalación limpia.
func colaDe(t *testing.T, dataDir string) []event.Event {
	t.Helper()
	pendientes, err := transport.Load(dataDir)
	if err != nil {
		t.Fatalf("inspeccionar la cola de %q: %v", transport.QueuePath(dataDir), err)
	}
	return pendientes
}

// TestProjectJoin_LaPeticionNuncaSeEncola cubre P-005 FR-018 y SC-010 **con el servidor
// inalcanzable**, que es la circunstancia en la que un cliente descuidado encolaría «para luego».
//
// ⛔ **Encolar la petición de alguien que está mirando la pantalla sería mentirle**: recibiría un
// desenlace que no es el de su operación, y la operación real ocurriría más tarde, sin nadie mirando.
//
// ⚠️ **NACE VERDE**: una ausencia la satisface un comando que no encola. **Su CASO POSITIVO lo ejecuta
// P-005 T027** — una emisión ordinaria con el destino igualmente caído **sí** hace crecer la cola. Si
// no creciera en ninguno de los dos casos, el observador no está mirando.
func TestProjectJoin_LaPeticionNuncaSeEncola(t *testing.T) {
	dataDir, arbol, banco := entornoDeDesenlace(t, http.StatusOK, `{"project":{"name":"RecetApp"}}`)
	banco.srv.Close() // destino escrito y muerto

	antes := colaDe(t, dataDir)

	r := ejecutarEn(t, arbol, nil, 20*time.Second, "project", "join", codigoDeAdhesion)

	// Precondición: si el comando hubiera tenido éxito, el destino no estaba caído y este test no
	// habría medido su circunstancia (disciplina 3: `Fatalf` sólo donde continuar no significa nada).
	if r.codigo == 0 {
		t.Fatalf("el comando salió con 0 contra un destino cerrado: la circunstancia de SC-010 —servidor "+
			"inalcanzable— no llegó a darse, así que la cola no prueba nada. stderr: %q", r.stderr)
	}

	if despues := colaDe(t, dataDir); len(despues) != len(antes) {
		t.Errorf("la cola pasó de %d a %d eventos: una petición de este comando NUNCA aparece en la cola "+
			"de envío diferido, ni con el servidor inalcanzable (P-005 FR-018, SC-010)", len(antes), len(despues))
	}

	t.Run("CASO POSITIVO · la cola sabe crecer", func(t *testing.T) {
		// P-005 T027 lo puso en marcha. Una emisión ORDINARIA con el destino **igualmente caído**:
		// `--run` escanea, encola y falla al drenar, así que **la MISMA inspección** tiene que
		// registrar el aumento. Que la cola no crezca en ninguno de los dos casos significaría que el
		// observador no está mirando, y entonces SC-010 no cuenta como pasado.
		dataDir, arbol, banco := entornoDeDesenlace(t, http.StatusOK, `{"project":{"name":"RecetApp"}}`)
		banco.srv.Close()

		// Logs que sí producen eventos facturables: el fixture del repositorio.
		logs := filepath.Join(dataDir, "logs-vacios")
		muestra, err := os.ReadFile(filepath.Join("..", "..", "internal", "ingest", "testdata", "claude_code_sample.jsonl"))
		if err != nil {
			t.Fatalf("leer el fixture de logs: %v", err)
		}
		if err := os.WriteFile(filepath.Join(logs, "sesion.jsonl"), muestra, 0o600); err != nil {
			t.Fatalf("sembrar el log: %v", err)
		}

		// ⚠️ EL PROCESO SE CORTA A PROPÓSITO, Y NO INVALIDA LA MEDIDA. Contra un destino caído, el
		// drenaje entra en el backoff acotado del transporte —hasta cinco reintentos doblando la
		// espera— y tardaría más de medio minuto en rendirse. **La cola ya creció antes de eso**: el
		// agente encola de forma durable ANTES de drenar (durabilidad, R4), que es precisamente lo que
		// SC-010 observa. Un límite corto mide lo mismo sin pagar el backoff.
		antes := colaDe(t, dataDir)
		ejecutarEn(t, arbol, nil, 5*time.Second, "--run")

		if despues := colaDe(t, dataDir); len(despues) <= len(antes) {
			t.Errorf("una emisión ordinaria con el destino caído no hizo crecer la cola (%d → %d): el "+
				"observador no está mirando, y entonces el verde de arriba no cuenta como pasado",
				len(antes), len(despues))
		}
	})
}
