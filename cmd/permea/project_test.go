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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/testutil"
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

	dataDir = testutil.Sandbox(t)
	banco = &bancoDeAdhesion{}
	banco.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		banco.mu.Lock()
		banco.peticiones++
		banco.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"project":{"name":"`+denominacion+`"}}`)
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
