package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/testutil"
)

// ═══ P-007 B2 · `--scan` IMPRIME EL DESGLOSE DE LA ESCRITURA DE CACHÉ (FR-016) ═══════════════════════
//
// La línea `evento:` añade `cw5m=` y `cw1h=` detrás de `cw=`, que no cambia: `cw=` sigue siendo el total
// que cruza la frontera, y las dos cifras nuevas son el instrumento de SC-005 sobre las copias del dueño.
// Identificadores sintéticos (disciplina 9). Tests de proceso: se compara `ExitCode()` (disciplina 4).

// lineaConDesglose es una línea `claude-opus-5-5` con 12 345 tokens de escritura a 5 minutos y 33 334 a 1 hora.
const lineaConDesglose = `{"type":"assistant","timestamp":"2026-10-06T12:00:00Z","sessionId":"s","cwd":"/tmp/x",` +
	`"requestId":"req_COSTESCAN0000000000000001","message":{"id":"msg_COSTESCAN0000000000000001","model":"claude-opus-5-5",` +
	`"usage":{"input_tokens":123457,"output_tokens":7891,"cache_creation_input_tokens":45679,"cache_read_input_tokens":987653,` +
	`"cache_creation":{"ephemeral_5m_input_tokens":12345,"ephemeral_1h_input_tokens":33334}}}}` + "\n"

// (8) · P-007 FR-016 — `--scan` imprime el desglose, y `cw=` sigue siendo el total.
func TestScan_LineaConDesgloseDeLaEscritura(t *testing.T) {
	_ = testutil.Sandbox(t)
	fichero := filepath.Join(t.TempDir(), "con-desglose.jsonl")
	if err := os.WriteFile(fichero, []byte(lineaConDesglose), 0o600); err != nil {
		t.Fatalf("escribir el fichero de prueba: %v", err)
	}

	codigo, stdout, _, _ := ejecutar(t, 20*time.Second, "--scan", fichero)
	eventos := lineasEvento(stdout)
	if codigo != 0 || len(eventos) != 1 {
		t.Fatalf("precondición: `--scan` debe salir con 0 e imprimir un evento (código %d, eventos %d)", codigo, len(eventos))
	}
	linea := eventos[0] + " "

	for _, campo := range []string{" cw=45679 ", " cw5m=12345 ", " cw1h=33334 "} {
		if !strings.Contains(linea, campo) {
			t.Errorf("P-007 FR-016: la línea `evento:` no lleva %q: %q", strings.TrimSpace(campo), eventos[0])
		}
	}
}
