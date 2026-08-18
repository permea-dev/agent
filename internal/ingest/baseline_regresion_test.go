package ingest

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/permea-dev/agent/internal/testutil"
)

// ═══ P-005 T029 · SC-009 · REGRESIÓN CERO DEL CAMINO DE INGESTA ════════════════════════════
//
// P-005 expone «hubo raíz» por una vía adicional (T001) y **no puede cambiar lo que la ingesta
// estampa** (FR-016). Esta es la puerta que lo demuestra, y lo hace **contra un artefacto**, no
// contra un recuerdo: `specs/004-identidad-de-proyecto/baseline-sc004.tsv`, capturado por 004 antes
// de tocar nada.
//
// ═══ POR QUÉ VIVE EN `internal/ingest` Y NO EN `internal/project` ══════════════════════════
//
// Las TRES columnas del baseline las produce `claudecode.go:86-88` —`ProjectRef` por
// `project.Derivar`, `SessionRef` y `MachineRef` por `event.Ref`—. Desde `internal/project` solo
// podría compararse UNA de las tres, y el baseline existe precisamente para comparar las tres:
// `project_ref` detecta que la derivación se movió, y las otras dos que P-005 no tocó lo que 004
// declaró intocable (FR-019 de 004).
//
// **Y no en un paquete nuevo**: el checkpoint de Phase 1 compara contra «9 paquetes ok». Un paquete
// nuevo lo convertiría en 10 y rompería la puerta que dice «nada cambió» justo al escribir la prueba
// de que nada cambió.
//
// ═══ ⚠️ ESTE TEST NACE VERDE, Y POR ESO LA MUTACIÓN NO ES CEREMONIA ════════════════════════
//
// Es regresión cero: si T001 está bien hecha, **nada cambió** y da verde desde el primer instante.
// La disciplina 3 exige validarlo por mutación, y aquí es **crítico**: éste es el ÚNICO testigo del
// camino de ingesta, y **un test que lea mal el `.tsv`, que reciba cero semillas o que compare dos
// conjuntos vacíos DA EXACTAMENTE EL MISMO VERDE que uno correcto**.
//
// De ahí las dos guardas explícitas de abajo —conjunto vacío y baseline vacío—: sin ellas, el modo
// de fallo del propio test es indistinguible de su éxito.

// rutaBaseline localiza el artefacto de 004 desde este paquete.
func rutaBaseline(t *testing.T) string {
	t.Helper()
	// internal/ingest → raíz del repositorio.
	return filepath.Join("..", "..", "specs", "004-identidad-de-proyecto", "baseline-sc004.tsv")
}

// baselineEsperada lee del artefacto el CONJUNTO de identidades y el RECUENTO de eventos.
//
// Son dos referencias distintas y **ninguna sustituye a la otra**, como declara la cabecera del
// propio fichero: las filas van deduplicadas —el conjunto—, y el `events_total` es el recuento, que
// el dedup destruiría. Un test que solo comparase el conjunto **no vería eventos perdidos**.
func baselineEsperada(t *testing.T) (identidades []string, eventosTotal int) {
	t.Helper()

	ruta := rutaBaseline(t)
	f, err := os.Open(ruta)
	if err != nil {
		t.Fatalf("baseline: no se pudo abrir %q: %v", ruta, err)
	}
	defer func() { _ = f.Close() }()

	eventosTotal = -1
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		linea := strings.TrimSpace(sc.Text())
		if linea == "" {
			continue
		}
		if strings.HasPrefix(linea, "#") {
			campos := strings.Fields(strings.TrimPrefix(linea, "#"))
			if len(campos) == 3 && campos[0] == "meta" && campos[1] == "events_total" {
				n, err := strconv.Atoi(campos[2])
				if err != nil {
					t.Fatalf("baseline: events_total no es un número: %q", campos[2])
				}
				eventosTotal = n
			}
			continue
		}
		campos := strings.Split(linea, "\t")
		if len(campos) != 3 {
			t.Fatalf("baseline: fila con %d columnas, se esperaban 3: %q", len(campos), linea)
		}
		if campos[0] == "project_ref" { // cabecera de columnas
			continue
		}
		identidades = append(identidades, strings.Join(campos, "\t"))
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("baseline: error leyendo %q: %v", ruta, err)
	}

	// GUARDA 1 — sin esto, un artefacto ilegible o vacío produciría una comparación de
	// «nada contra nada» que pasa en verde y no demuestra absolutamente nada.
	if len(identidades) == 0 {
		t.Fatalf("baseline: %q no declara ninguna fila de identidades — la comparación sería vacua", ruta)
	}
	if eventosTotal < 0 {
		t.Fatalf("baseline: %q no declara `# meta events_total` — no hay recuento contra el que comparar", ruta)
	}
	return identidades, eventosTotal
}

// TestSC009_RegresionCeroDelCaminoDeIngesta reproduce la pasada de referencia de 004 y exige que las
// tres columnas y el recuento de eventos coincidan con el baseline.
func TestSC009_RegresionCeroDelCaminoDeIngesta(t *testing.T) {
	// El sandbox aísla el estado local Y siembra las dos semillas deterministas del bloque
	// REPRODUCCIÓN. Con otras semillas los refs NO comparan y un «fallo» no significaría nada.
	_ = testutil.SandboxConSemillas(t)
	salt, machineID := testutil.SemillasDeLaLineaBase(t)

	// GUARDA 2 — semillas vacías producirían refs consistentes entre sí pero ajenos al baseline;
	// el test fallaría por la razón equivocada, o —peor— pasaría si el baseline también lo fuera.
	if salt == "" || machineID == "" {
		t.Fatalf("semillas vacías (salt=%q machine_id=%q): la comparación no significaría nada", salt, machineID)
	}

	ctx := Context{
		Salt:      salt,
		MachineID: machineID,
		DevID:     "dev-baseline",
		OrgID:     "org-baseline",
	}

	f, err := os.Open(filepath.Join("testdata", "claude_code_sample.jsonl"))
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	vistas := map[string]bool{}
	eventos := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	for sc.Scan() {
		ev, err := FromClaudeCodeLine(sc.Bytes(), ctx)
		if err != nil {
			t.Fatalf("fixture: línea corrupta: %v", err)
		}
		if ev == nil {
			continue // no facturable
		}
		eventos++
		vistas[strings.Join([]string{ev.ProjectRef, ev.SessionRef, ev.MachineRef}, "\t")] = true
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("fixture: %v", err)
	}

	obtenidas := make([]string, 0, len(vistas))
	for k := range vistas {
		obtenidas = append(obtenidas, k)
	}
	sort.Strings(obtenidas)

	esperadas, eventosTotal := baselineEsperada(t)
	sort.Strings(esperadas)

	// SC-003 de 004 — el recuento, que el dedup destruye. Si bajó, se han perdido eventos.
	if eventos != eventosTotal {
		t.Errorf("recuento de eventos: got %d, want %d (baseline `# meta events_total`)", eventos, eventosTotal)
	}

	// SC-004 de 004 + FR-016 de 005 — el CONJUNTO de identidades, las tres columnas.
	if len(obtenidas) != len(esperadas) {
		t.Fatalf("conjunto de identidades: got %d filas, want %d\n got: %v\nwant: %v",
			len(obtenidas), len(esperadas), obtenidas, esperadas)
	}
	for i := range esperadas {
		if obtenidas[i] != esperadas[i] {
			t.Errorf("identidad %d difiere del baseline:\n got: %s\nwant: %s", i, obtenidas[i], esperadas[i])
		}
	}
}
