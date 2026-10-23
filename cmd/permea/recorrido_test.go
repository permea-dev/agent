package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/event"
	"github.com/permea-dev/agent/internal/project"
	"github.com/permea-dev/agent/internal/state"
	"github.com/permea-dev/agent/internal/testutil"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ 013 B3 · EL RECORRIDO COMÚN — (5) y (6) de plan §Bloques ═══════════════════════════════════════════════════════════
//
// `generate()` y `permea project list` recorren los seis lectores con EL MISMO código (013 FR-009, D-013-P2), y la emisión no
// cambia ni un byte (FR-024, PARA 6). La referencia es la de T002: `--run` del binario de `cabf8c5` sobre las plantillas de
// `testdata/lista/`, con `@CASA@` → `/sintetico/casa`.

// casaDeLaReferencia es la sustitución de `@CASA@` de la referencia de SC-012: rutas que no existen, así que sus `project_ref`
// no dependen del temporal (plan §9).
const casaDeLaReferencia = "/sintetico/casa"

// carpetasDeLaPlantilla traduce cada carpeta de `testdata/lista/plantillas/` a la del hogar (generar.sh).
var carpetasDeLaPlantilla = map[string]string{"claude": ".claude", "codex": ".codex", "gemini": ".gemini", "copilot": ".copilot",
	"kimi-code": ".kimi-code"}

// hogarDeLista monta en un `SandboxConSemillas` el hogar de las plantillas, con `@CASA@` → `casa`, y la base de OpenCode
// `sana.db`. Las plantillas se leen y se copian: nunca se escriben en su sitio. Devuelve el `DataDir`.
func hogarDeLista(t *testing.T, casa string) string {
	t.Helper()
	dataDir := testutil.SandboxConSemillas(t)
	hogar := os.Getenv("HOME")
	plantillas := filepath.Join("testdata", "lista", "plantillas")
	err := filepath.WalkDir(plantillas, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(plantillas, p)
		if err != nil {
			return err
		}
		partes := strings.SplitN(rel, string(filepath.Separator), 2)
		carpeta, ok := carpetasDeLaPlantilla[partes[0]]
		if !ok || len(partes) != 2 {
			t.Fatalf("precondición: plantilla fuera del mapa: %s", rel)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		escribir(t, filepath.Join(hogar, carpeta, partes[1]), []byte(strings.ReplaceAll(string(b), "@CASA@", casa)))
		return nil
	})
	if err != nil {
		t.Fatalf("precondición: plantillas: %v", err)
	}
	db, err := os.ReadFile(filepath.Join("testdata", "opencode", "sana.db"))
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	escribir(t, filepath.Join(hogar, ".local", "share", "opencode", "opencode.db"), db)
	return dataDir
}

// referenciaDeLista lee un fichero de la referencia de SC-012 (`testdata/lista/referencia-run.<parte>`).
func referenciaDeLista(t *testing.T, parte string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "lista", "referencia-run."+parte))
	if err != nil {
		t.Fatalf("precondición: la referencia %s: %v", parte, err)
	}
	return string(b)
}

// (5) · SC-012, FR-024: `--run` sobre las plantillas de los seis lectores es BYTE A BYTE la referencia del binario de `cabf8c5`
// (T002), con la ruta de datos por `<DATOS>`: la cola, el stdout y el stderr. NACE VERDE (T014, antes de tocar `main.go`); lo
// valida m8.
func TestRecorrido_RunEsLaReferenciaDeLa090(t *testing.T) {
	dataDir := hogarDeLista(t, casaDeLaReferencia)
	codigo, stdout, stderr, expiro := ejecutar(t, 60*time.Second, "--run")
	if codigo != 0 || expiro {
		t.Fatalf("--run: código %d, expiró %v; stderr:\n%s", codigo, expiro, stderr)
	}
	cola, err := os.ReadFile(transport.QueuePath(dataDir))
	if err != nil {
		t.Fatalf("leer la cola: %v", err)
	}
	for _, c := range []struct{ parte, got string }{{"cola", string(cola)}, {"stdout", stdout}, {"stderr", stderr}} {
		t.Run(c.parte, func(t *testing.T) {
			if got, quiere := strings.ReplaceAll(c.got, dataDir, "<DATOS>"), referenciaDeLista(t, c.parte); got != quiere {
				t.Errorf("%s distinto de la referencia de la 0.9.0:\n--- ahora\n%s\n--- referencia\n%s", c.parte, got, quiere)
			}
		})
	}
}

// (6) · FR-009: `recorrer`, con un sumidero que RECOGE y un estado que empieza vacío y no se guarda, da los mismos eventos y en
// el mismo orden que la cola de (5), y no escribe nada en el sandbox. `no_escribe` NACE VERDE con el esqueleto; lo valida m7.
func TestRecorrido_ElSumideroRecogeLoQueSeEncola(t *testing.T) {
	_ = hogarDeLista(t, casaDeLaReferencia)
	raiz := filepath.Dir(os.Getenv("HOME"))
	a, err := setup() // la sal y el `machine_id` ya están sembrados: no crea nada
	if err != nil {
		t.Fatalf("precondición: setup: %v", err)
	}
	root, err := config.ClaudeCodeLogsRoot(a.cfg)
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	logs, err := state.FindLogs(root)
	if err != nil {
		t.Fatalf("precondición: %v", err)
	}
	antes := arbolCompleto(t, raiz)

	ictx := a.ictx
	ictx.Resolutor = project.NuevoResolutor()
	var recogido strings.Builder
	n, _, err := a.recorrer(logs, state.New(), ictx, func(ev event.Event) error {
		b, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		recogido.Write(append(b, '\n'))
		return nil
	})

	t.Run("mismos_eventos", func(t *testing.T) {
		quiere := referenciaDeLista(t, "cola")
		if err != nil || recogido.String() != quiere || n != strings.Count(quiere, "\n") {
			t.Errorf("recorrer: %d eventos, err %v; lo recogido no es la cola de la referencia:\n--- recogido\n%s\n--- referencia\n%s",
				n, err, recogido.String(), quiere)
		}
	})
	t.Run("no_escribe", func(t *testing.T) {
		if dif := diferencias(antes, arbolCompleto(t, raiz)); len(dif) != 0 {
			t.Errorf("recorrer escribió en el sandbox: %v", dif)
		}
	})
}
