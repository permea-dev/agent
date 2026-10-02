package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/permea-dev/agent/internal/testutil"
)

// ═══ P-006 B4 · LA AYUDA DICE LA VERDAD, Y PEDIRLA NO EJECUTA NADA ═════════════════════════
//
// Contrato: `specs/006-medicion-fiel/contracts/cli.md`. Todos son tests de PROCESO contra el binario
// que compila `TestMain`: se compara `ExitCode()`, los dos canales se capturan POR SEPARADO
// (disciplinas 4 y 7) y todo corre en sandbox (disciplina 6). Los secretos y el token son CENTINELAS
// sintéticos.

// tokenCentinelaB4 es el device_token de prueba de (25): no puede aparecer en ninguna salida.
const tokenCentinelaB4 = "dev_tok_CENTINELA_B4_7c1e9a"

// arbolCompleto es la huella de un directorio CON sus subdirectorios: ruta → contenido (o "dir").
// `huellaDelDirectorio` (project_test.go) sólo mira ficheros, y lo que `status -h` crea hoy es un
// directorio VACÍO: con aquella huella el cambio sería invisible.
func arbolCompleto(t *testing.T, raiz string) map[string]string {
	t.Helper()
	huella := map[string]string{}
	err := filepath.WalkDir(raiz, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(raiz, ruta)
		if d.IsDir() {
			huella[rel] = "dir"
			return nil
		}
		b, err := os.ReadFile(ruta)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		huella[rel] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatalf("recorrer %q: %v", raiz, err)
	}
	return huella
}

// arbolDeTrabajo crea un árbol con raíz reconocible (un `.git`) fuera del sandbox.
func arbolDeTrabajo(t *testing.T) string {
	t.Helper()
	arbol := filepath.Join(t.TempDir(), "proyecto")
	if err := os.MkdirAll(filepath.Join(arbol, ".git"), 0o755); err != nil {
		t.Fatalf("crear el árbol de trabajo: %v", err)
	}
	return arbol
}

// ── (20) ──────────────────────────────────────────────────────────────────────────────────

// (20) · P-006 FR-021, SC-012 — sin argumentos, `help`, `-h`, `--help` y `-help` dan la MISMA ayuda,
// por stdout, con stderr vacío y salida 0. (`-help` lo añade el contrato a las cuatro de la spec.)
func TestAyuda_LasCuatroFormasSonIdenticas(t *testing.T) {
	_ = testutil.Sandbox(t)
	formas := []struct {
		nombre string
		args   []string
	}{
		{"sin_argumentos", nil},
		{"help", []string{"help"}},
		{"-h", []string{"-h"}},
		{"--help", []string{"--help"}},
		{"-help", []string{"-help"}},
	}
	salidas := map[string]string{}
	for _, f := range formas {
		codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, f.args...)
		salidas[f.nombre] = stdout
		t.Run(f.nombre, func(t *testing.T) {
			if codigo != 0 {
				t.Errorf("ExitCode() = %d, se esperaba 0", codigo)
			}
			if stderr != "" {
				t.Errorf("stderr debe estar VACÍO; trae %q", stderr)
			}
			if strings.TrimSpace(stdout) == "" {
				t.Errorf("stdout vacío: la ayuda debe salir por stdout")
			}
		})
	}
	t.Run("identicas", func(t *testing.T) {
		ref := salidas["help"]
		if strings.TrimSpace(ref) == "" {
			t.Errorf("la ayuda de `help` está vacía: no hay nada que comparar")
		}
		for _, f := range formas {
			if salidas[f.nombre] != ref {
				t.Errorf("la salida de %s difiere de la de `help`:\n--- %s\n%s\n--- help\n%s", f.nombre, f.nombre, salidas[f.nombre], ref)
			}
		}
	})
}

// ── (21) ──────────────────────────────────────────────────────────────────────────────────

// (21) · P-006 FR-021, contrato §La ayuda general — la ayuda (forma `help`) contiene los tres
// subcomandos, las cuatro opciones y la vía stdin recomendada para los DOS valores sensibles.
func TestAyuda_ContenidoMinimo(t *testing.T) {
	_ = testutil.Sandbox(t)
	_, stdout, _, _ := ejecutar(t, 20*time.Second, "help")
	elementos := []struct{ nombre, texto string }{
		{"enroll", "enroll"},
		{"status", "status"},
		{"project_join", "project join"},
		{"--scan", "--scan"},
		{"--run", "--run"},
		{"--daemon", "--daemon"},
		{"--version", "--version"},
		{"stdin_enroll", `| permea enroll -`},
		{"stdin_project_join", `| permea project join -`},
	}
	for _, e := range elementos {
		t.Run(e.nombre, func(t *testing.T) {
			if !strings.Contains(stdout, e.texto) {
				t.Errorf("la ayuda (stdout de `permea help`) no contiene %q", e.texto)
			}
		})
	}
}

// ── (22) ──────────────────────────────────────────────────────────────────────────────────

// lasOchoAyudasDeSubcomando son las de `contracts/cli.md` §Las ayudas de subcomando.
var lasOchoAyudasDeSubcomando = [][]string{
	{"enroll", "-h"}, {"enroll", "--help"},
	{"status", "-h"}, {"status", "--help"},
	{"project", "-h"}, {"project", "--help"},
	{"project", "join", "-h"}, {"project", "join", "--help"},
}

// (22) · P-006 FR-023, SC-013 — las ocho ayudas de subcomando salen por stdout con exit 0 y NO HACEN
// NADA, desde dentro de un árbol de proyecto y en dos montajes:
//
//	(a) `vacio`    — sandbox SIN nada: el árbol del HOME queda idéntico, directorios incluidos. Se
//	    borra el directorio de datos VACÍO que `testutil.Sandbox` crea por adelantado; con él puesto,
//	    que `status -h` lo cree sería invisible.
//	(b) `enrolado` — agente enrolado contra el banco del propio test (`entornoDeAdhesion`): el banco
//	    recibe 0 peticiones y el HOME no cambia.
//
// El 0 de (b) no es vacuo porque el hijo CONFÍA en el banco por `SSL_CERT_FILE` (montaje de
// `entornoDeAdhesion`): sin esa confianza, una petición moriría en el apretón TLS antes de llegar al
// manejador que cuenta, y el 0 saldría igual con el defecto puesto. Lo demuestra m16, que tumba (b)
// por peticiones.
//
// Además, stdout tiene que ser LA AYUDA y no otra salida: empieza por `uso: permea <subcomando>`.
// `status` sin ayuda imprime su estado por stdout con exit 0, y sin esta comprobación pasaría.
func TestAyudaSubcomando_NoHaceNada(t *testing.T) {
	for _, montaje := range []string{"vacio", "enrolado"} {
		for _, args := range lasOchoAyudasDeSubcomando {
			nombre := montaje + "/" + strings.Join(args, "_")
			t.Run(nombre, func(t *testing.T) {
				var banco *bancoDeAdhesion
				var arbol string
				if montaje == "vacio" {
					dataDir := testutil.Sandbox(t)
					if err := os.Remove(dataDir); err != nil {
						t.Fatalf("vaciar el sandbox: %v", err)
					}
					arbol = arbolDeTrabajo(t)
				} else {
					_, arbol, banco = entornoDeAdhesion(t, "RecetApp")
				}
				hogar := os.Getenv("HOME")
				antes := arbolCompleto(t, hogar)

				d := ejecutarEn(t, arbol, nil, 20*time.Second, args...)

				if d.codigo != 0 {
					t.Errorf("ExitCode() = %d, se esperaba 0", d.codigo)
				}
				if prefijo := "uso: permea " + strings.Join(args[:len(args)-1], " "); !strings.HasPrefix(d.stdout, prefijo) {
					t.Errorf("stdout debe ser la ayuda del subcomando (empezar por %q); trae %q", prefijo, d.stdout)
				}
				if d.stderr != "" {
					t.Errorf("stderr debe estar VACÍO; trae %q", d.stderr)
				}
				if dif := diferencias(antes, arbolCompleto(t, hogar)); len(dif) != 0 {
					t.Errorf("pedir ayuda cambió el HOME del sandbox: %v", dif)
				}
				if banco != nil && banco.recibidas() != 0 {
					t.Errorf("pedir ayuda emitió %d peticiones al banco; se esperaban 0", banco.recibidas())
				}
			})
		}
	}
}

// ── (23) y (24) ───────────────────────────────────────────────────────────────────────────

// (23) · P-006 FR-022, SC-014 — un subcomando inexistente falla con exit 1, NOMBRÁNDOLO por stderr, y
// stdout vacío. Se busca con comillas (`"enrol"`) porque `enroll` sin comillas aparece en la ayuda.
func TestDesconocido_SubcomandoNombradoYSalida1(t *testing.T) {
	_ = testutil.Sandbox(t)
	codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, "enrol")
	if codigo != 1 {
		t.Errorf("ExitCode() = %d, se esperaba 1", codigo)
	}
	if !strings.Contains(stderr, `"enrol"`) {
		t.Errorf("stderr debe nombrar el subcomando desconocido (\"enrol\"); trae %q", stderr)
	}
	if stdout != "" {
		t.Errorf("stdout debe estar vacío; trae %q", stdout)
	}
}

// (24) · P-006 FR-022, FR-024, SC-014 — si lo tecleado tiene forma de secreto conocido, el error NO lo
// reproduce: el centinela no aparece en ninguno de los dos canales.
func TestDesconocido_NoReproduceSecretos(t *testing.T) {
	_ = testutil.Sandbox(t)
	const centinela = "CENTINELASECRETOB4x9"
	for _, prefijo := range []string{"pmea2.", "pmeaj1.", "pmea1."} {
		t.Run(prefijo, func(t *testing.T) {
			codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, prefijo+centinela)
			if codigo != 1 {
				t.Errorf("ExitCode() = %d, se esperaba 1", codigo)
			}
			if strings.Contains(stdout, centinela) || strings.Contains(stderr, centinela) {
				t.Errorf("el secreto tecleado se reprodujo:\nstdout: %q\nstderr: %q", stdout, stderr)
			}
		})
	}
}

// ── (25) ──────────────────────────────────────────────────────────────────────────────────

// (25) · P-006 FR-024, SC-015 — con un agente enrolado con token CENTINELA, ninguna invocación del
// contrato de la CLI lo imprime, por ningún canal.
func TestTokenCentinela_NuncaSale(t *testing.T) {
	invocaciones := [][]string{
		nil, {"help"}, {"-h"}, {"--help"}, {"-help"},
		{"status"}, {"enrol"}, {"--bogus"}, {"--scan"},
	}
	invocaciones = append(invocaciones, lasOchoAyudasDeSubcomando...)
	for _, args := range invocaciones {
		nombre := strings.Join(args, "_")
		if nombre == "" {
			nombre = "sin_argumentos"
		}
		t.Run(nombre, func(t *testing.T) {
			dataDir, arbol, _ := entornoDeAdhesion(t, "RecetApp")
			escribirConfigCon(t, dataDir, "https://127.0.0.1:1"+rutaDeIngesta, tokenCentinelaB4)
			d := ejecutarEn(t, arbol, nil, 20*time.Second, args...)
			if strings.Contains(d.stdout, tokenCentinelaB4) || strings.Contains(d.stderr, tokenCentinelaB4) {
				t.Errorf("el token centinela apareció en la salida:\nstdout: %q\nstderr: %q", d.stdout, d.stderr)
			}
		})
	}
}

// ── (26) ──────────────────────────────────────────────────────────────────────────────────

// (26) · contrato §Opción desconocida (E-006-P3) — una opción desconocida o sin valor: exit 2, stdout
// VACÍO, y stderr nombra la opción y remite a `permea help`.
func TestOpcionDesconocida_StderrYSalida2(t *testing.T) {
	_ = testutil.Sandbox(t)
	casos := []struct{ nombre, arg, nombra string }{
		{"desconocida", "--bogus", "bogus"},
		{"sin_valor", "--scan", "scan"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, c.arg)
			if codigo != 2 {
				t.Errorf("ExitCode() = %d, se esperaba 2", codigo)
			}
			if stdout != "" {
				t.Errorf("stdout debe estar VACÍO; trae %q", stdout)
			}
			if !strings.Contains(stderr, c.nombra) {
				t.Errorf("stderr debe nombrar la opción (%q); trae %q", c.nombra, stderr)
			}
			if !strings.Contains(stderr, "permea help") {
				t.Errorf("stderr debe remitir a `permea help`; trae %q", stderr)
			}
		})
	}
}

// ── (27) (28) (29) · D-006-14: el texto aprobado por el dueño ─────────────────────────────

// lasCincoAyudas son la general y las cuatro de subcomando, por su invocación.
var lasCincoAyudas = []struct {
	nombre string
	args   []string
}{
	{"general", []string{"help"}},
	{"enroll", []string{"enroll", "-h"}},
	{"status", []string{"status", "-h"}},
	{"project", []string{"project", "-h"}},
	{"project_join", []string{"project", "join", "-h"}},
}

// (27) · D-006-14 (1) — ninguna línea de ninguna ayuda pasa de 80 caracteres (runas, no bytes).
func TestAyuda_AnchoMaximo80(t *testing.T) {
	_ = testutil.Sandbox(t)
	for _, a := range lasCincoAyudas {
		t.Run(a.nombre, func(t *testing.T) {
			codigo, stdout, _, _ := ejecutar(t, 20*time.Second, a.args...)
			if codigo != 0 || strings.TrimSpace(stdout) == "" {
				t.Fatalf("precondición: la ayuda debe salir por stdout con exit 0 (código %d)", codigo)
			}
			for i, linea := range strings.Split(stdout, "\n") {
				if n := utf8.RuneCountInString(linea); n > 80 {
					t.Errorf("línea %d: %d caracteres (> 80): %q", i+1, n, linea)
				}
			}
		})
	}
}

// (28) · D-006-14 (2) — la ayuda general trae «Primeros pasos» en el orden enroll → status → --run, y
// el aviso de que la primera pasada envía todo el historial que conserve Claude Code.
func TestAyuda_PrimerosPasosYAviso(t *testing.T) {
	_ = testutil.Sandbox(t)
	_, stdout, _, _ := ejecutar(t, 20*time.Second, "help")

	// El bloque: desde «Primeros pasos:» hasta la primera línea en blanco.
	bloque := ""
	if i := strings.Index(stdout, "Primeros pasos:"); i >= 0 {
		bloque = stdout[i:]
		if j := strings.Index(bloque, "\n\n"); j >= 0 {
			bloque = bloque[:j]
		}
	}
	t.Run("bloque", func(t *testing.T) {
		if bloque == "" {
			t.Errorf("la ayuda general no contiene «Primeros pasos:»")
		}
	})
	t.Run("orden", func(t *testing.T) {
		e, s, r := strings.Index(bloque, "enroll"), strings.Index(bloque, "status"), strings.Index(bloque, "--run")
		if e < 0 || s < 0 || r < 0 || e >= s || s >= r {
			t.Errorf("en «Primeros pasos» deben ir enroll → status → --run; posiciones %d, %d, %d en %q", e, s, r, bloque)
		}
	})
	t.Run("aviso", func(t *testing.T) {
		// Con los espacios normalizados: en el texto aprobado el aviso cruza un salto de línea
		// («…envía todo» / «el historial que conserve…»), y la búsqueda literal no casaría nunca.
		if !strings.Contains(strings.Join(strings.Fields(stdout), " "), "todo el historial") {
			t.Errorf("la ayuda general no avisa de que la primera pasada envía «todo el historial»")
		}
	})
}

// (29) · D-006-14 (3) — ninguna ayuda contiene jerga interna.
func TestAyuda_SinJergaInterna(t *testing.T) {
	_ = testutil.Sandbox(t)
	for _, a := range lasCincoAyudas {
		t.Run(a.nombre, func(t *testing.T) {
			_, stdout, _, _ := ejecutar(t, 20*time.Second, a.args...)
			for _, jerga := range []string{"P-001", "P-002", "sync_interval"} {
				if strings.Contains(stdout, jerga) {
					t.Errorf("la ayuda contiene jerga interna %q", jerga)
				}
			}
		})
	}
}
