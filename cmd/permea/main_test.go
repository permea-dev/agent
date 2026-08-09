package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/testutil"
)

// TestVersionFlag verifica el contrato de `--version` (contracts/artifacts.md): imprime
// EXACTAMENTE la versión en stdout (una línea, sin el prefijo "Permea") para que la
// verificación de release sea comprobable (SC-002). La versión es la que inyecta el
// -ldflags de GoReleaser desde la etiqueta.
func TestVersionFlag(t *testing.T) {
	var buf bytes.Buffer
	printVersion(&buf)
	got := buf.String()

	if got != version+"\n" {
		t.Fatalf("printVersion = %q, want %q", got, version+"\n")
	}
	if strings.Contains(got, "Permea") {
		t.Fatalf("--version no debe incluir el prefijo Permea: %q", got)
	}
}

// TestAgentVersion_ReachesEvent verifica el cableado de T036: la versión REAL del binario
// (variable `version` de este paquete) se propaga por newIngestContext hasta
// Event.AgentVersion, sin depender del sistema de ficheros ni de la config local.
func TestAgentVersion_ReachesEvent(t *testing.T) {
	const want = "9.9.9-test"
	ictx := newIngestContext(want, config.Config{DevID: "dev-1", OrgID: "org-1"}, "salt", "machine")

	if ictx.AgentVersion != want {
		t.Fatalf("Context.AgentVersion = %q, want %q", ictx.AgentVersion, want)
	}

	// Una línea de asistente facturable -> el Event resultante debe llevar la versión.
	line := []byte(`{"type":"assistant","timestamp":"2026-06-20T10:15:30Z","sessionId":"s","cwd":"/x","message":{"model":"claude-opus-4-6","usage":{"input_tokens":10,"output_tokens":5}}}`)
	ev, err := ingest.FromClaudeCodeLine(line, ictx)
	if err != nil {
		t.Fatalf("FromClaudeCodeLine: %v", err)
	}
	if ev == nil {
		t.Fatal("se esperaba un evento facturable")
	}
	if ev.AgentVersion != want {
		t.Fatalf("Event.AgentVersion = %q, want %q", ev.AgentVersion, want)
	}
}

// TestVersion_DefaultNonEmpty: la variable version nunca es vacía por defecto (evita
// emitir agent_version="" si no se inyecta -ldflags).
func TestVersion_DefaultNonEmpty(t *testing.T) {
	if version == "" {
		t.Fatal("version por defecto no debe ser vacía")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════════════
// P-004 T023/T024 · La retirada del modo en claro — tests DE PROCESO
//
// PATRÓN NUEVO EN EL REPO, y se declara: hasta ahora los tests de `cmd/permea` llamaban a
// las funciones internas. La parada de FR-013 solo es observable desde fuera —es un CÓDIGO
// DE SALIDA—, así que hace falta ejecutar el binario. `TestMain` lo compila una vez a un
// temporal; no se usa `./bin/permea` ni nada del PATH, que podría ser otra versión.
//
// Se compara SIEMPRE `ExitCode()`, NUNCA texto (tasks.md, disciplina 4): el puente
// Windows/WSL hace que el texto de error no sea estable. Y la AUSENCIA de aviso se comprueba
// por `stderr` VACÍO (disciplina 5), no por matching de un mensaje que no está: comprobar
// que un texto no aparece también pasa cuando aparece otro distinto.
// ═══════════════════════════════════════════════════════════════════════════════════════

// binarioDePrueba es la ruta del binario compilado para los tests de proceso.
var binarioDePrueba string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "permea-test-bin")
	if err != nil {
		fmt.Fprintln(os.Stderr, "TestMain: no se pudo crear el temporal:", err)
		os.Exit(1)
	}
	binarioDePrueba = filepath.Join(dir, "permea-test")
	build := exec.Command("go", "build", "-o", binarioDePrueba, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: no se pudo compilar el binario de prueba: %v\n%s", err, out)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// entornoDePrueba prepara el sandbox de la receta común de tasks.md:150-156 y devuelve el
// dataDir. La `config.json` la escribe el llamante con la clave que esté probando.
//
// Los tres ingredientes existen para que CADA CÓDIGO DE SALIDA CAIGA POR SU PROPIA RAZÓN:
// (a) sandbox aislado; (b) config bien formada salvo la clave bajo prueba —un `endpoint`
// inválido haría fallar `config.Validate()` y daría un !=0 por otro motivo—; y (c)
// `logs_root` a un directorio VACÍO, para que la pasada no encole nada y el !=0 no pueda
// venir de un fallo de transporte.
func entornoDePrueba(t *testing.T, valorDeLaClaveRetirada string) string {
	t.Helper()

	dataDir := testutil.Sandbox(t)
	logsVacios := filepath.Join(dataDir, "logs-vacios")
	if err := os.MkdirAll(logsVacios, 0o755); err != nil {
		t.Fatalf("crear logs_root vacío: %v", err)
	}

	cfg := map[string]any{
		"endpoint":      "https://127.0.0.1:1",
		"device_token":  "0123456789abcdef0123456789abcdef",
		"org_id":        "org-prueba",
		"dev_id":        "dev-prueba",
		"tools":         []string{"claude_code"},
		"sync_interval": "60s",
		"logs_root":     logsVacios,
	}
	if valorDeLaClaveRetirada != "" {
		cfg["project_ref_mode"] = valorDeLaClaveRetirada
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "config.json"), b, 0o600); err != nil {
		t.Fatalf("escribir config: %v", err)
	}
	return dataDir
}

// ejecutar corre el binario de prueba y devuelve su código de salida y sus dos flujos.
// `limite` acota la espera: un subcomando que NO para —como el daemon hoy— se quedaría vivo
// para siempre y colgaría la suite.
func ejecutar(t *testing.T, limite time.Duration, args ...string) (codigo int, stdout, stderr string, expiro bool) {
	t.Helper()

	ctx, cancelar := context.WithTimeout(context.Background(), limite)
	defer cancelar()

	cmd := exec.CommandContext(ctx, binarioDePrueba, args...)
	var salida, errores bytes.Buffer
	cmd.Stdout = &salida
	cmd.Stderr = &errores
	err := cmd.Run()

	expiro = ctx.Err() != nil
	codigo = 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			codigo = ee.ExitCode()
		} else {
			codigo = -1
		}
	}
	return codigo, salida.String(), errores.String(), expiro
}

// lineasEnCola cuenta las líneas de queue.jsonl; una cola inexistente son cero.
func lineasEnCola(t *testing.T, dataDir string) int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dataDir, "queue.jsonl"))
	if err != nil {
		return 0
	}
	return len(bytes.Split(bytes.TrimSpace(b), []byte("\n"))) - btoi(len(bytes.TrimSpace(b)) == 0)
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T023 · SC-007 · la parada — tabla de contracts/cli-config.md, anclada a D-004-5
// ───────────────────────────────────────────────────────────────────────────────────────

func TestRetirada_LaParadaDeLosCaminosQueEmiten(t *testing.T) {
	t.Run(`--run con "plain" → para y no encola`, func(t *testing.T) {
		dataDir := entornoDePrueba(t, "plain")

		codigo, _, _, _ := ejecutar(t, 20*time.Second, "--run")
		if codigo == 0 {
			t.Errorf("FR-013/SC-007: `--run` con `project_ref_mode:\"plain\"` DEBE detenerse con código != 0.\n"+
				"código obtenido: %d", codigo)
		}
		if n := lineasEnCola(t, dataDir); n != 0 {
			t.Errorf("FR-013: la garantía es CERO eventos procesados o emitidos; queue.jsonl tiene %d líneas", n)
		}
	})

	t.Run(`--daemon con "plain" → para`, func(t *testing.T) {
		entornoDePrueba(t, "plain")

		// El daemon es un bucle: si NO para, vive para siempre. El límite corto lo mata y
		// `expiro` distingue «se detuvo por la config» de «siguió vivo hasta que lo matamos».
		codigo, _, _, expiro := ejecutar(t, 3*time.Second, "--daemon")
		if expiro {
			t.Errorf("FR-013/SC-007 (segunda fila de la tabla): el daemon ARRANCÓ Y SIGUIÓ VIVO con la\n" +
				"configuración retirada; hubo que matarlo al vencer el plazo. Debe detenerse igual que\n" +
				"`--run`: si solo se probara `--run`, el daemon podría quedarse sin la parada y nadie lo vería.")
			return
		}
		if codigo == 0 {
			t.Errorf("FR-013/SC-007: `--daemon` con \"plain\" DEBE detenerse con código != 0; obtuvo %d", codigo)
		}
	})

	t.Run(`--run con "hash" → arranca en silencio`, func(t *testing.T) {
		entornoDePrueba(t, "hash")

		codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "--run")
		if codigo != 0 {
			t.Errorf("FR-013a: el valor que pedía el comportamiento YA ÚNICO no puede detener nada; código %d\nstderr: %s", codigo, stderr)
		}
		exigirSinAvisoDeLaClave(t, stderr, `--run con "hash"`)
	})

	t.Run("--run sin la clave → arranca en silencio", func(t *testing.T) {
		entornoDePrueba(t, "")

		codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "--run")
		if codigo != 0 {
			t.Errorf("una config sin la clave obsoleta no puede detener nada; código %d\nstderr: %s", codigo, stderr)
		}
		exigirSinAvisoDeLaClave(t, stderr, "--run sin la clave")
	})
}

// exigirSinAvisoDeLaClave comprueba la AUSENCIA de aviso sobre la clave retirada.
//
// No se puede exigir `stderr` completamente vacío: el arranque normal escribe el banner
// («Permea <versión>») y el resumen de la pasada. Lo que se exige es que NINGUNA línea
// mencione la clave — que es la ausencia que FR-013a/FR-013b prometen, y se comprueba por
// inspección de TODAS las líneas, no buscando un mensaje concreto que podría cambiar.
func exigirSinAvisoDeLaClave(t *testing.T, stderr, contexto string) {
	t.Helper()
	if strings.Contains(stderr, "project_ref_mode") || strings.Contains(stderr, "plain") {
		t.Errorf("FR-013a/FR-013b: %s NO debe producir aviso alguno sobre la clave retirada.\nstderr:\n%s",
			contexto, stderr)
	}
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T024 · D-004-5 · las tres excepciones
// ───────────────────────────────────────────────────────────────────────────────────────

func TestRetirada_LasExcepcionesDeD0045(t *testing.T) {
	t.Run(`status con "plain" → informa SIN parar`, func(t *testing.T) {
		entornoDePrueba(t, "plain")

		codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "status")
		if codigo != 0 {
			t.Errorf("D-004-5: `status` es DIAGNÓSTICO y no se detiene; código %d", codigo)
		}
		// La presencia del aviso se comprueba por NO-VACUIDAD, simétrica a cómo T023
		// comprueba su ausencia. Ninguno de los dos hace matching de un texto concreto.
		if strings.TrimSpace(stderr) == "" {
			t.Errorf("D-004-5: `status` con \"plain\" DEBE informar del problema — su función es explicar\n" +
				"el estado, y esto es parte del estado. stderr llegó vacío.")
		}
	})

	t.Run(`status con "hash" → ni una línea sobre la clave`, func(t *testing.T) {
		// El blanco limpio de la mutación «aviso demasiado ancho»: si el aviso se disparara por
		// la presencia de la clave en vez de por su valor, este sub-caso lo caza y el de
		// `"plain"` seguiría pasando — que es justo lo que lo hace útil.
		entornoDePrueba(t, "hash")

		codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "status")
		if codigo != 0 {
			t.Errorf("`status` con \"hash\" no puede detenerse: código %d", codigo)
		}
		exigirSinAvisoDeLaClave(t, stderr, `status con "hash"`)
	})

	t.Run(`--scan con "plain" presente → procesa sin parar`, func(t *testing.T) {
		dataDir := entornoDePrueba(t, "plain")
		fixture := filepath.Join(dataDir, "muestra.jsonl")
		linea := `{"type":"assistant","timestamp":"2026-08-09T10:00:00Z","sessionId":"s","cwd":"/tmp/x","message":{"model":"claude-opus-4-6","usage":{"input_tokens":10,"output_tokens":5}}}` + "\n"
		if err := os.WriteFile(fixture, []byte(linea), 0o600); err != nil {
			t.Fatalf("escribir fixture: %v", err)
		}

		codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "--scan", fixture)
		if codigo != 0 {
			t.Errorf("D-004-5: `--scan` es PROCESAMIENTO DIAGNÓSTICO —salt de dry-run, sin cola, sin\n"+
				"transporte— y no se detiene aunque la config retirada esté presente; código %d\nstderr: %s",
				codigo, stderr)
		}
	})
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T024 (continuación) · `enroll` como VÍA DE REPARACIÓN
// ───────────────────────────────────────────────────────────────────────────────────────

// TestRetirada_EnrollNoParaYLimpiaLaClave comprueba las DOS mitades de lo que D-004-5
// promete de `enroll`: que no se detiene, Y que repara.
//
// Comprobar solo la primera dejaría en pie un `enroll` que arranca y no limpia nada — que es
// exactamente el caso que haría FALSO el argumento con el que D-004-5 justifica no
// detenerlo.
//
// SE EJECUTA EN PROCESO, no como binario, y es una decisión razonada: `enroll` hace un ping
// real contra el endpoint, y un proceso hijo NO confiaría en el certificado autofirmado del
// `httptest.NewTLSServer` sin montarle un almacén de certificados — complejidad que no
// añade cobertura. El repo ya prueba `enroll` así (`enroll_test.go`), con `verifyVia` y el
// `pmea2` fabricado por `mkEnrollStr`, es decir con el codificador propio y NUNCA un backend
// real. El «código de salida 0» se comprueba aquí como «error nil», que es lo que `main`
// convierte en código.
func TestRetirada_EnrollNoParaYLimpiaLaClave(t *testing.T) {
	srv, _ := okBackend(t)
	dataDir := testutil.Sandbox(t)

	original := `{
  "endpoint": "https://viejo.invalid",
  "device_token": "tok-viejo",
  "org_id": "org-1",
  "dev_id": "dev-1",
  "project_ref_mode": "plain",
  "tools": ["claude_code"],
  "sync_interval": "60s"
}`
	rutaCfg := filepath.Join(dataDir, "config.json")
	if err := os.WriteFile(rutaCfg, []byte(original), 0o600); err != nil {
		t.Fatalf("sembrar config: %v", err)
	}

	var stdout bytes.Buffer
	err := enroll([]string{mkEnrollStr(t, srv.URL, "tok-nuevo")}, strings.NewReader(""), false, &stdout, verifyVia(srv))
	if err != nil {
		t.Errorf("D-004-5: `enroll` es la VÍA DE REPARACIÓN y no puede detenerse ante la clave retirada: %v", err)
	}

	crudo, errLectura := os.ReadFile(rutaCfg)
	if errLectura != nil {
		t.Fatalf("releer config: %v", errLectura)
	}
	var mapa map[string]any
	if err := json.Unmarshal(crudo, &mapa); err != nil {
		t.Fatalf("la config reescrita debe ser JSON válido: %v", err)
	}
	if _, presente := mapa["project_ref_mode"]; presente {
		t.Errorf("D-004-5/FR-015: `enroll` hace Load+Save, así que DEBE dejar el fichero sin la clave\n"+
			"retirada — es lo que sostiene el argumento de que no haga falta detenerlo.\ncontenido:\n%s", crudo)
	}
}

// ───────────────────────────────────────────────────────────────────────────────────────
// T027 · El ORDEN de la comprobación dentro de setup()
// ───────────────────────────────────────────────────────────────────────────────────────

// TestRetirada_ElErrorQueGanaEsElDeLaClave comprueba la parte más sutil de T027: que la
// detección va ANTES de cualquier otro modo de fallo de `setup()`.
//
// POR QUÉ IMPORTA: si un fallo de salt o de directorio de datos pudiera adelantarse, el
// usuario con `"plain"` recibiría **un error distinto del suyo** y la parada quedaría
// indistinguible de cualquier otra avería. La garantía de FR-013 incluye que el usuario sepa
// QUÉ le paró — un `exit != 0` acompañado del mensaje equivocado cumple la letra y falla el
// propósito.
//
// EL MONTAJE: un dataDir de SOLO LECTURA hace fallar a `LoadOrCreateSalt`, que necesita
// escribir el fichero `salt`. Con la comprobación en su sitio gana el error de la clave; si
// alguien la moviera detrás, ganaría el del salt y este test lo diría.
//
// La presencia del aviso se comprueba POR FRAGMENTO —el nombre de la clave—, que es el
// inverso exacto de `exigirSinAvisoDeLaClave`. Nunca se compara el mensaje entero: eso lo
// volvería frágil ante cualquier reescritura del texto.
func TestRetirada_ElErrorQueGanaEsElDeLaClave(t *testing.T) {
	// El t.Skip documentado se permite aquí por la misma asimetría razonada de T018
	// (tasks.md:132): como root, un directorio de solo lectura no deniega nada, así que el
	// montaje es físicamente inejecutable. La diferencia con T010 —donde el skip está
	// prohibido— es que allí mataría la ÚNICA cobertura de una promesa; aquí el resto de la
	// tabla de la parada sigue cubierto por T023.
	if os.Geteuid() == 0 {
		t.Skip("ejecutando como root: un dataDir de solo lectura no impide escribir, el montaje no existe")
	}

	dataDir := entornoDePrueba(t, "plain")

	if err := os.Chmod(dataDir, 0o500); err != nil {
		t.Fatalf("no se pudo dejar el dataDir en solo lectura: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dataDir, 0o700) }) // para que t.TempDir pueda limpiar

	codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "--run")
	if codigo == 0 {
		t.Fatalf("con la config retirada el arranque debe detenerse; código %d", codigo)
	}
	if !strings.Contains(stderr, "project_ref_mode") {
		t.Errorf("T027: con OTRO fallo de setup() disponible —el salt no se puede escribir—, el error\n"+
			"que gana debe seguir siendo el de la CLAVE RETIRADA. El usuario tiene que saber qué le\n"+
			"paró; un error cualquiera cumple el código de salida y falla el propósito.\nstderr:\n%s", stderr)
	}
}
