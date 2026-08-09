// Package testutil da a los tests un sandbox donde el estado del agente
// (config.json, state.json, queue.jsonl, salt, machine_id) vive en un directorio temporal
// y NUNCA en la instalación real del desarrollador.
//
// POR QUÉ ES UN HELPER COMPARTIDO Y NO CÓDIGO REPETIDO (plan §I2): tres familias de tests
// necesitan el mismo aislamiento —los de config, los de proceso y las validaciones—, y
// replicarlo a mano en cada una es la vía segura de que a alguna se le olvide y escriba en
// el `config.json` del desarrollador que ejecuta la suite. Un test que ensucia la máquina
// de quien lo corre se descubre tarde y mal.
//
// ═══ POR QUÉ ESTE PAQUETE NO IMPORTA `internal/config` ══════════════════════════════════
//
// Sería lo natural —la aserción de aislamiento querría preguntar por `config.DataDir()`—, y
// es exactamente lo que NO se puede hacer: `internal/config/config_test.go` es un test
// INTERNO (`package config`), y los tests de la clave obsoleta (T022/T022b) van ahí. Si este
// paquete importara `config`, esos tests no podrían usar el helper: ciclo de importación.
//
// Así que la aserción se apoya en `os.UserConfigDir()`, que es la primitiva sobre la que
// `config.DataDir()` está construido (`internal/config/config.go:51`). No es un rodeo: es
// comprobar el mecanismo real de resolución en vez de nuestra envoltura de él.
package testutil

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// dirDatos es el nombre del subdirectorio del agente dentro del directorio de configuración
// del sistema. Coincide con `internal/config.DataDir()`; se replica aquí, y solo aquí, para
// no importar ese paquete (ver la cabecera).
const dirDatos = "permea"

// Sandbox aísla el estado del agente en un directorio temporal del test y devuelve el
// dataDir que el agente resolverá dentro de él.
//
// Establece HOME, USERPROFILE y XDG_CONFIG_HOME —las tres, porque `os.UserConfigDir()` mira
// una u otra según el sistema operativo— con `t.Setenv`, que las restaura al terminar el
// test. Como `t.Setenv` es incompatible con `t.Parallel()`, un test que use este helper NO
// puede marcarse paralelo; es el precio de no tocar el entorno global de forma duradera.
//
// La ASERCIÓN DE AISLAMIENTO es parte del helper, no una recomendación: si tras establecer
// las variables el sistema sigue resolviendo fuera del temporal, el test MUERE aquí mismo
// con t.Fatal. Un aviso serviría de poco — el test seguiría, escribiría donde no debe, y el
// daño estaría hecho antes de que nadie leyera la advertencia.
func Sandbox(t *testing.T) string {
	t.Helper()

	raiz := t.TempDir()
	hogar := filepath.Join(raiz, "home")
	configuracion := filepath.Join(hogar, ".config")
	if err := os.MkdirAll(configuracion, 0o700); err != nil {
		t.Fatalf("sandbox: no se pudo crear el hogar temporal %q: %v", configuracion, err)
	}

	t.Setenv("HOME", hogar)
	t.Setenv("USERPROFILE", hogar)
	t.Setenv("XDG_CONFIG_HOME", configuracion)

	base, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("sandbox: os.UserConfigDir() falló tras aislar el entorno: %v", err)
	}

	// LA ASERCIÓN. `raiz` viene de t.TempDir() y es absoluta y limpia; `base` sale del
	// sistema. Si `base` no cuelga de `raiz`, el aislamiento NO se aplicó y cualquier cosa
	// que el test escriba a continuación iría a la instalación real.
	if !dentroDe(base, raiz) {
		t.Fatalf(
			"sandbox: AISLAMIENTO NO APLICADO — os.UserConfigDir() resolvió %q, fuera del temporal %q.\n"+
				"El test se detiene aquí a propósito: continuar escribiría en la instalación real del desarrollador.",
			base, raiz,
		)
	}

	dataDir := filepath.Join(base, dirDatos)
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		t.Fatalf("sandbox: no se pudo crear el dataDir %q: %v", dataDir, err)
	}
	return dataDir
}

// SandboxConSemillas es Sandbox más las DOS SEMILLAS DETERMINISTAS de la línea base.
//
// PARA QUÉ: `config.LoadOrCreateSalt` genera un salt ALEATORIO cuando el fichero no existe
// (`internal/config/identity.go:13-14`), y un sandbox nace vacío. Como las tres identidades
// del evento son `Ref(salt, valor)`, un salt aleatorio hace que los refs NO se repitan entre
// pasadas — y entonces no hay nada que comparar contra la línea base. `loadOrCreateSecret`
// lee el fichero si existe (`identity.go:24-28`), así que sembrarlo basta.
//
// QUIÉN LA NECESITA: la verificación de neutralidad de T007 y cualquier test de proceso que
// compare contra `baseline-sc004.tsv`. Los tests unitarios del resolutor (T008..T020) NO:
// construyen sus propios árboles y pasan el salt explícitamente, así que no dependen de
// ninguna semilla persistida.
//
// LA FUENTE DE VERDAD ES EL ARTEFACTO, NO ESTE FICHERO: los valores se LEEN de
// `specs/004-identidad-de-proyecto/baseline-sc004.tsv`, del bloque REPRODUCCIÓN de su
// cabecera. No se copian aquí como literales: dos copias de un valor son dos copias que
// pueden divergir, y el día que divergieran, la comparación fallaría sin que nadie
// entendiera por qué.
func SandboxConSemillas(t *testing.T) string {
	t.Helper()

	dataDir := Sandbox(t)
	salt, machineID := SemillasDeLaLineaBase(t)

	escribirSecreto(t, filepath.Join(dataDir, "salt"), salt)
	escribirSecreto(t, filepath.Join(dataDir, "machine_id"), machineID)

	return dataDir
}

// SemillasDeLaLineaBase lee `salt` y `machine_id` del bloque REPRODUCCIÓN de
// `baseline-sc004.tsv`. Falla el test si el artefacto no está o si no las declara: es
// preferible a devolver un valor por defecto que produciría una comparación silenciosamente
// inválida.
func SemillasDeLaLineaBase(t *testing.T) (salt, machineID string) {
	t.Helper()

	ruta := filepath.Join(raizDelRepo(t), "specs", "004-identidad-de-proyecto", "baseline-sc004.tsv")
	f, err := os.Open(ruta)
	if err != nil {
		t.Fatalf("semillas: no se pudo abrir la línea base %q: %v", ruta, err)
	}
	defer func() { _ = f.Close() }()

	// Formato del bloque: líneas de comentario `#   <clave>   <valor>`.
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		campos := strings.Fields(strings.TrimPrefix(strings.TrimSpace(sc.Text()), "#"))
		if len(campos) < 2 {
			continue
		}
		switch campos[0] {
		case "salt":
			salt = campos[1]
		case "machine_id":
			machineID = campos[1]
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("semillas: error leyendo %q: %v", ruta, err)
	}
	if salt == "" || machineID == "" {
		t.Fatalf(
			"semillas: %q no declara las dos semillas en su bloque REPRODUCCIÓN (salt=%q machine_id=%q).\n"+
				"Sin ellas, cualquier comparación contra la línea base es inválida.",
			ruta, salt, machineID,
		)
	}
	return salt, machineID
}

// raizDelRepo deriva la raíz del repositorio de la ubicación de ESTE fichero en tiempo de
// compilación, no del directorio de trabajo del test: `go test` lo fija en el directorio del
// paquete, que es distinto para cada uno.
func raizDelRepo(t *testing.T) string {
	t.Helper()

	_, fichero, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("raizDelRepo: runtime.Caller(0) no devolvió la ruta de sandbox.go")
	}
	// <raíz>/internal/testutil/sandbox.go → subir tres niveles.
	return filepath.Dir(filepath.Dir(filepath.Dir(fichero)))
}

// escribirSecreto persiste una semilla con los mismos permisos restrictivos que usa el
// agente para salt y machine_id (0600).
func escribirSecreto(t *testing.T, ruta, valor string) {
	t.Helper()

	if err := os.WriteFile(ruta, []byte(valor), 0o600); err != nil {
		t.Fatalf("semillas: no se pudo escribir %q: %v", ruta, err)
	}
}

// dentroDe indica si `ruta` es `base` o cuelga de ella. Compara por SEGMENTOS —vía
// filepath.Rel— y no por prefijo de cadena: `/tmp/abc` NO está dentro de `/tmp/ab`, y una
// comparación con strings.HasPrefix diría que sí.
func dentroDe(ruta, base string) bool {
	rel, err := filepath.Rel(base, ruta)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
