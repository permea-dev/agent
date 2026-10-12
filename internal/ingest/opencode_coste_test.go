//go:build sc017

package ingest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/permea-dev/agent/internal/state"
)

// P-010 SC-017 · El coste de la pasada (D-010-P13). Sólo con `-tags sc017`: la suite normal no lo compila, y así no tiene
// SKIP. Genera con sqlite3, en t.TempDir() y nunca en el repo, una base sintética de ≈ 1 GB: ≈ 250 000 filas `assistant`
// completadas, con un `data` de 1 a 12 KB, en 2 500 sesiones. Mide la función del lector (leer y clasificar), sin encolar:
//
//	go test -tags sc017 -run TestSC017 -count=1 -timeout 90m ./internal/ingest
//
// Topes: la primera pasada, ≤ 20 s; la siguiente, con 10 filas nuevas, ≤ 50 ms y ≤ 1 % de las páginas de la base.

const (
	filasSC017    = 250000
	sesionesSC017 = 2500
)

// sqlFilasSC017 inserta las filas [desde, hasta]: un tercio de 1 a 12 KB de relleno, el resto de 1 a 4 KB (≈ 1 GB en total).
func sqlFilasSC017(desde, hasta int) string {
	return fmt.Sprintf(`WITH RECURSIVE n(i) AS (SELECT %d UNION ALL SELECT i+1 FROM n WHERE i<%d)
INSERT INTO session_message(id, session_id, type, seq, time_created, time_updated, data)
SELECT printf('o-sc017-%%07d', i), printf('s-sc017-%%04d', i %% %d), 'assistant', i, 1760000000000+i, 1760000000000+i,
  json_object('role', 'assistant', 'time', json_object('created', 1760000000000+i, 'completed', 1760000001000+i),
    'model', json_object('providerID', 'proveedor-sintetico', 'id', 'modelo-sintetico-opencode', 'variant', 'default'),
    'tokens', json_object('input', 100+i%%977, 'output', 10+i%%97, 'reasoning', i%%13,
       'cache', json_object('read', i%%4000, 'write', i%%7)), 'cost', 0.0001,
    'pad', hex(zeroblob(512*(CASE WHEN i%%3=0 THEN 1+i%%12 ELSE 1+i%%4 END))))
FROM n;`, desde, hasta, sesionesSC017)
}

func sqlite3SC017(t *testing.T, ruta, sql string) {
	t.Helper()
	cmd := exec.Command("sqlite3", ruta)
	cmd.Stdin = strings.NewReader(sql)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("sqlite3: %v\n%s", err, out)
	}
}

func TestSC017_CosteDeLaPasada(t *testing.T) {
	if _, err := exec.LookPath("sqlite3"); err != nil {
		t.Fatal("SC-017 necesita sqlite3 para generar la base")
	}
	ruta := filepath.Join(t.TempDir(), "opencode.db")
	inicio := time.Now()
	sqlite3SC017(t, ruta, "PRAGMA journal_mode=OFF; PRAGMA synchronous=OFF;\n"+
		strings.Replace(`CREATE TABLE session_v2 (id text PRIMARY KEY, project_id text NOT NULL, slug text NOT NULL,
  directory text NOT NULL, path text, title text, version text NOT NULL, time_created integer NOT NULL, time_updated integer NOT NULL);
CREATE TABLE session_message (id text PRIMARY KEY, session_id text NOT NULL, type text NOT NULL, seq integer NOT NULL,
  time_created integer NOT NULL, time_updated integer NOT NULL, data text NOT NULL);
CREATE UNIQUE INDEX session_message_session_seq_idx ON session_message (session_id, seq);
WITH RECURSIVE n(i) AS (SELECT 0 UNION ALL SELECT i+1 FROM n WHERE i<NSES-1)
INSERT INTO session_v2 SELECT printf('s-sc017-%04d', i), 'prj-sintetico', 'slug', printf('/sintetico/sc017/%04d', i), NULL,
  'titulo', '2.0.26', 1760000000000, 1760000000000 FROM n;`, "NSES", fmt.Sprint(sesionesSC017), 1)+"\n"+sqlFilasSC017(1, filasSC017))
	info, err := os.Stat(ruta)
	if err != nil {
		t.Fatal(err)
	}
	paginasBase := info.Size() / 4096
	t.Logf("base: %d B (%d páginas de 4 KB), %d filas en %d sesiones, generada en %v", info.Size(), paginasBase, filasSC017,
		sesionesSC017, time.Since(inicio).Round(time.Second))

	st := state.New()
	ctx := ContextoOpenCode{Context: Context{Salt: "sal-sc017"}, Ahora: func() time.Time { return time.UnixMilli(1760000000000) }}
	p := NuevaPasadaOpenCode()
	inicio = time.Now()
	evs, err := LeerBaseOpenCode(st, ruta, ctx, p)
	primera := time.Since(inicio)
	if err != nil {
		t.Fatalf("primera pasada: %v", err)
	}
	t.Logf("primera pasada: %v · %d eventos · %d páginas · %s", primera.Round(time.Millisecond), len(evs), p.paginas, p.Resumen())
	if primera > 20*time.Second || len(evs) != filasSC017 {
		t.Errorf("primera pasada en %v con %d eventos; tope 20 s y %d eventos", primera, len(evs), filasSC017)
	}

	sqlite3SC017(t, ruta, sqlFilasSC017(filasSC017+1, filasSC017+10))
	p = NuevaPasadaOpenCode()
	inicio = time.Now()
	evs, err = LeerBaseOpenCode(st, ruta, ctx, p)
	siguiente := time.Since(inicio)
	if err != nil {
		t.Fatalf("pasada siguiente: %v", err)
	}
	t.Logf("pasada siguiente: %v · %d eventos · %d páginas (%.3f %% de la base) · %s", siguiente.Round(time.Microsecond),
		len(evs), p.paginas, 100*float64(p.paginas)/float64(paginasBase), p.Resumen())
	if siguiente > 50*time.Millisecond || p.paginas*100 > paginasBase || len(evs) != 10 {
		t.Errorf("pasada siguiente en %v, %d páginas y %d eventos; topes 50 ms, %d páginas (1 %%) y 10 eventos", siguiente,
			p.paginas, len(evs), paginasBase/100)
	}
}
