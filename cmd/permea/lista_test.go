package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/permea-dev/agent/internal/testutil"
)

// ═══ 013 B6 · EL VERBO — (17) a (21) de plan §Bloques ═════════════════════════════════════════════════════════════════
//
// Se ejercen por el binario de prueba *(`ejecutar`)*, en `testutil.Sandbox`. Cada texto esperado sale de spec §Textos POR
// PROGRAMA *(`bloqueDeLaSpec`)*: no hay una segunda copia que pueda envejecer.

// bloqueDeLaSpec devuelve, línea a línea, el bloque de código que sigue a la línea de spec §Textos que empieza por `titulo`.
func bloqueDeLaSpec(t *testing.T, titulo string) []string {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "specs", "013-lista-de-proyectos", "spec.md"))
	if err != nil {
		t.Fatalf("precondición: spec: %v", err)
	}
	defer func() { _ = f.Close() }()
	var lineas []string
	enTitulo, enBloque := false, false
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 1<<20)
	for sc.Scan() {
		l := sc.Text()
		switch {
		case !enTitulo && strings.HasPrefix(l, titulo):
			enTitulo = true
		case enTitulo && !enBloque && l == "```":
			enBloque = true
		case enBloque && l == "```":
			return lineas
		case enBloque:
			lineas = append(lineas, l)
		}
	}
	t.Fatalf("precondición: spec §Textos no tiene el bloque de %q", titulo)
	return nil
}

// errorDeLaSpec es la línea de «Errores del comando» que empieza por `prefijo`, sin su `<error>` final.
func errorDeLaSpec(t *testing.T, prefijo string) string {
	t.Helper()
	for _, l := range bloqueDeLaSpec(t, "**Errores del comando**") {
		if strings.HasPrefix(l, prefijo) {
			return strings.TrimSuffix(l, "<error>")
		}
	}
	t.Fatalf("precondición: spec §Textos no tiene el error %q", prefijo)
	return ""
}

// (17) · D-5, P-11, FR-007, FR-008: sin sal, el texto por stdout, salida 0, y la sal sigue sin existir; con la sal ilegible, el
// error y salida 1.
func TestProjectList_LaSal(t *testing.T) {
	t.Run("sin_sal", func(t *testing.T) {
		dataDir := testutil.Sandbox(t)
		codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, "project", "list")
		quiere := strings.Join(bloqueDeLaSpec(t, "**Sin sal**"), "\n") + "\n"
		if codigo != 0 || stdout != quiere || stderr != "" {
			t.Errorf("código %d, stdout %q, stderr %q; se esperaba 0, %q y nada", codigo, stdout, stderr, quiere)
		}
		if _, err := os.Lstat(filepath.Join(dataDir, "salt")); !os.IsNotExist(err) {
			t.Errorf("la sal existe tras `project list` (err = %v): NUNCA la crea", err)
		}
	})
	t.Run("ilegible", func(t *testing.T) {
		dataDir := testutil.Sandbox(t)
		if err := os.Mkdir(filepath.Join(dataDir, "salt"), 0o700); err != nil { // un directorio en su sitio: ilegible también con root
			t.Fatalf("precondición: %v", err)
		}
		codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, "project", "list")
		if prefijo := errorDeLaSpec(t, "error: la sal"); codigo != 1 || stdout != "" || !strings.HasPrefix(stderr, prefijo) {
			t.Errorf("código %d, stdout %q, stderr %q; se esperaba 1, nada y %q…", codigo, stdout, stderr, prefijo)
		}
	})
}

// (18) · P-7, FR-002, SC-014: con un argumento, el error de uso, salida 1, y NUNCA lo tecleado.
func TestProjectList_Argumentos(t *testing.T) {
	for _, c := range []struct{ hoja, arg string }{{"x", "x"}, {"secreto", "pmea2.CENTINELA-de-la-lista"}} {
		t.Run(c.hoja, func(t *testing.T) {
			_ = testutil.SandboxConSemillas(t)
			codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, "project", "list", c.arg)
			if quiere := errorDeLaSpec(t, "error: project list no admite") + "\n"; codigo != 1 || stdout != "" || stderr != quiere {
				t.Errorf("código %d, stdout %q, stderr %q; se esperaba 1, nada y %q", codigo, stdout, stderr, quiere)
			}
			if strings.Contains(stdout+stderr, c.arg) {
				t.Errorf("la salida reproduce el argumento %q", c.arg)
			}
		})
	}
}

// (19) · D-11, FR-003, SC-013: la ayuda de spec §Textos, literal; ≤ 80 runas y sin jerga; pedirla no escribe nada; y la fila, en
// la ayuda general y en la de `project`.
func TestProjectList_Ayuda(t *testing.T) {
	fila := strings.Join(bloqueDeLaSpec(t, "**La ayuda**"), "\n") + "\n"
	for _, forma := range []string{"-h", "--help"} {
		t.Run(forma, func(t *testing.T) {
			_ = testutil.Sandbox(t)
			raiz := filepath.Dir(os.Getenv("HOME"))
			antes := arbolCompleto(t, raiz)
			codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, "project", "list", forma)
			if quiere := "uso: permea project list\n\n" + fila; codigo != 0 || stdout != quiere || stderr != "" {
				t.Errorf("código %d, stdout:\n%s\nstderr %q; se esperaba 0 y:\n%s", codigo, stdout, stderr, quiere)
			}
			for i, l := range strings.Split(stdout, "\n") {
				if n := utf8.RuneCountInString(l); n > 80 {
					t.Errorf("línea %d: %d runas (> 80): %q", i+1, n, l)
				}
			}
			for _, jerga := range []string{"P-0", "FR-", "D-0", "sync_interval", "project_ref"} {
				if strings.Contains(stdout, jerga) {
					t.Errorf("la ayuda lleva jerga interna %q", jerga)
				}
			}
			if dif := diferencias(antes, arbolCompleto(t, raiz)); len(dif) != 0 {
				t.Errorf("pedir la ayuda escribió: %v", dif)
			}
		})
	}
	for _, c := range []struct {
		hoja string
		args []string
	}{{"en_la_general", []string{"help"}}, {"en_project", []string{"project", "-h"}}} {
		t.Run(c.hoja, func(t *testing.T) {
			_ = testutil.Sandbox(t)
			if codigo, stdout, _, _ := ejecutar(t, 20*time.Second, c.args...); codigo != 0 || !strings.Contains(stdout, fila) {
				t.Errorf("código %d; la ayuda de %v no trae la fila de `project list`:\n%s", codigo, c.args, stdout)
			}
		})
	}
}

// (20) · FR-001: los dos mensajes del verbo, literales, con `join, list`.
func TestProjectList_Verbos(t *testing.T) {
	verbos := bloqueDeLaSpec(t, "**Los verbos**")
	if len(verbos) != 2 {
		t.Fatalf("precondición: spec §Textos trae %d mensajes de verbos; se esperaban 2", len(verbos))
	}
	for _, c := range []struct {
		hoja   string
		args   []string
		quiere string
	}{
		{"falta", []string{"project"}, verbos[0]},
		{"desconocido", []string{"project", "adherirse"}, strings.Replace(verbos[1], "<verbo>", "adherirse", 1)},
	} {
		t.Run(c.hoja, func(t *testing.T) {
			_ = testutil.Sandbox(t)
			if codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, c.args...); codigo != 1 || stdout != "" || stderr != c.quiere+"\n" {
				t.Errorf("código %d, stdout %q, stderr %q; se esperaba 1, nada y %q", codigo, stdout, stderr, c.quiere)
			}
		})
	}
}

// (21) · P-8, FR-006: una configuración ilegible es un error; `project_ref_mode: "plain"` ni detiene ni avisa. `plain` NACE VERDE
// con el esqueleto *(sale con 0 sin escribir)*; lo valida m17.
func TestProjectList_Configuracion(t *testing.T) {
	t.Run("ilegible", func(t *testing.T) {
		dataDir := testutil.SandboxConSemillas(t)
		escribir(t, filepath.Join(dataDir, "config.json"), []byte("{ no es json"))
		codigo, stdout, stderr, _ := ejecutar(t, 20*time.Second, "project", "list")
		if prefijo := errorDeLaSpec(t, "error: la configuración"); codigo != 1 || stdout != "" || !strings.HasPrefix(stderr, prefijo) {
			t.Errorf("código %d, stdout %q, stderr %q; se esperaba 1, nada y %q…", codigo, stdout, stderr, prefijo)
		}
	})
	t.Run("plain", func(t *testing.T) {
		dataDir := testutil.SandboxConSemillas(t)
		escribir(t, filepath.Join(dataDir, "config.json"), []byte(`{"project_ref_mode": "plain"}`+"\n"))
		codigo, _, stderr, _ := ejecutar(t, 20*time.Second, "project", "list")
		if codigo != 0 || strings.Contains(stderr, "plain") || strings.Contains(stderr, "error") {
			t.Errorf("código %d, stderr %q; se esperaba 0, sin el error ni el aviso del modo retirado", codigo, stderr)
		}
	})
}
