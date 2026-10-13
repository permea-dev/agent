package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/permea-dev/agent/internal/ingest"
	"github.com/permea-dev/agent/internal/sqlitero"
	"github.com/permea-dev/agent/internal/state"
	"github.com/permea-dev/agent/internal/transport"
)

// ═══ P-010 B6 · OPENCODE EN `--run` Y EN EL DEMONIO ═════════════════════════════════════════════════════════════
//
// `generate()` lee OpenCode DESPUÉS de Gemini y ANTES del único `st.Save` (FR-027). Las rutas se resolvieron en `setup()`;
// si existen se miran en CADA pasada (P-10, FR-002), porque el demonio vive días y OpenCode puede instalarse después.

// opencodeActivo dice si OpenCode se lee EN ESTA PASADA (P-10): si existe la base de `OPENCODE_DB` o la carpeta de datos.
func (a *agent) opencodeActivo() bool {
	if a.opencodeUnica != "" {
		info, err := os.Stat(a.opencodeUnica)
		return err == nil && info.Mode().IsRegular()
	}
	if a.opencodeDatos == "" {
		return false
	}
	info, err := os.Stat(a.opencodeDatos)
	return err == nil && info.IsDir()
}

// generarOpenCode lee las bases de OpenCode y ENCOLA sus eventos (P-010 B6). Los errores son POR BASE (FR-029, D-010-P6,
// P7): una base que no se puede leer se omite con un aviso que lleva su NOMBRE, su entrada de `state.json` queda como
// estaba —la pasada siguiente la relee— y se sigue con la siguiente, sin impedir guardar el estado de los demás. Encolar
// sigue siendo fatal, como hoy.
func (a *agent) generarOpenCode(st *state.Store, ictx ingest.Context) (int, error) {
	po := ingest.NuevaPasadaOpenCode()
	a.opencode = po
	bases := []string{a.opencodeUnica}
	if a.opencodeUnica == "" {
		var err error
		if bases, err = ingest.ListarOpenCode(a.opencodeDatos); err != nil {
			var pe *os.PathError
			if errors.As(err, &pe) {
				err = pe.Err // sin la ruta: en Windows lleva el nombre del usuario
			}
			fmt.Fprintf(os.Stderr, "opencode: base omitida: %s: %v\n", filepath.Base(a.opencodeDatos), err)
			return 0, nil
		}
	}
	base := ingest.ContextoOpenCode{Context: ictx, Ahora: a.reloj}
	total := 0
	for _, ruta := range bases {
		evs, err := ingest.LeerBaseOpenCode(st, ruta, base, po)
		if err != nil {
			fmt.Fprintf(os.Stderr, "opencode: base omitida: %v\n", err) // con el nombre de la base, no su ruta (D-010-P7)
			continue
		}
		for _, ev := range evs {
			if err := transport.Append(a.dir, ev); err != nil {
				return total, err
			}
			total++
		}
	}
	return total, nil
}

// ═══ P-010 B7 · `--scan` DE UNA BASE DE OPENCODE ═════════════════════════════════════════════════════════════════
//
// Es lo único que lee `$.cost` (D-1): para la línea de contraste, que no viaja.

// esBaseSQLite dice si el fichero empieza por la firma de SQLite 3 (P-010 FR-025). Uno más corto que la firma no lo es, y
// sigue por el camino de hoy.
func esBaseSQLite(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }() // sólo lectura
	firma := make([]byte, 16)
	if _, err := io.ReadFull(f, firma); err != nil {
		return false, nil
	}
	return string(firma) == "SQLite format 3\x00", nil
}

// dryRunOpenCode imprime los eventos de una base de OpenCode sin tocar estado ni cola (P-010 FR-025, plan §5): la pasada de
// B4 sobre un estado sólo en memoria que nunca se guarda, la línea `evento:` de Codex y Gemini, el resumen de OpenCode y la
// línea de contraste (spec §Textos), si la base se leyó.
func dryRunOpenCode(path string) error {
	po := ingest.NuevaPasadaOpenCode()
	ctx := ingest.ContextoOpenCode{Context: ingest.Context{Salt: "dry-run-salt", MachineID: "local", DevID: "dev-local",
		OrgID: "org-local", AgentVersion: version}}
	evs, err := ingest.LeerBaseOpenCode(state.New(), path, ctx, po)
	if err != nil {
		return err
	}
	for _, ev := range evs {
		ref := ev.ProjectRef
		if len(ref) > 8 {
			ref = ref[:8] + "…"
		}
		fmt.Printf("evento: tool=%s model=%s in=%d out=%d cw=%d cr=%d cost=$%.4f cost_avail=%t project_ref=%s event_id=%s\n",
			ev.Tool, ev.Model, ev.TokensInput, ev.TokensOutput, ev.TokensCacheCreation, ev.TokensCacheRead,
			ev.CostUSD, ev.CostAvailable, ref, ev.EventID)
	}
	fmt.Fprintf(os.Stderr, "%d eventos generados (dry-run, nada transmitido)\n", len(evs))
	fmt.Fprintln(os.Stderr, po.Resumen()) // P-010 FR-025: el resumen de OpenCode, a stderr como el de Codex y Gemini
	if po.FormatoAnterior+po.EsquemaDesconocido > 0 {
		return nil // contada en el resumen: no hay mensajes que contrastar
	}
	var c contrasteOpenCode
	if err := c.leer(path); err != nil {
		return err
	}
	fmt.Printf("opencode (contraste, no se envía): coste según OpenCode $%.6f · fuera de los mensajes: entrada %d · salida %d · coste $%.6f\n",
		c.coste, c.fueraEntrada, c.fueraSalida, c.fueraCoste)
	return nil
}

// contrasteOpenCode es la línea de contraste (D-1, D-2): la Σ de `$.cost` de las filas `assistant`, y lo que queda «fuera de
// los mensajes», los totales de `session_v2` menos la Σ de esas filas. Una columna o una partida que falta vale 0.
type contrasteOpenCode struct {
	coste, fueraCoste         float64
	fueraEntrada, fueraSalida int64
}

// sumar añade una fila con su signo: + las sesiones, − los mensajes. La salida es `output + reasoning`, la de P-2.
func (c *contrasteOpenCode) sumar(signo int64, coste float64, entrada, salida, razonamiento int64) {
	c.fueraCoste += float64(signo) * coste
	c.fueraEntrada += signo * entrada
	c.fueraSalida += signo * (salida + razonamiento)
}

// leer recorre `session_message` y `session_v2` de una base que la pasada ya leyó.
func (c *contrasteOpenCode) leer(path string) error {
	b, err := sqlitero.Abrir(path)
	if err != nil {
		return err
	}
	defer func() { _ = b.Cerrar() }() // sólo lectura
	msgs, err := b.Tabla("session_message")
	if err != nil {
		return err
	}
	ses, err := b.Tabla("session_v2")
	if err != nil {
		return err
	}
	colTipo, colData := msgs.Col("type"), msgs.Col("data")
	err = msgs.Recorrer(func(fila []sqlitero.Valor) error {
		var d struct {
			Cost   float64
			Tokens struct{ Input, Output, Reasoning int64 }
		}
		if string(fila[colTipo].B) == "assistant" && json.Unmarshal(fila[colData].B, &d) == nil {
			c.coste += d.Cost
			c.sumar(-1, d.Cost, d.Tokens.Input, d.Tokens.Output, d.Tokens.Reasoning)
		}
		return nil
	})
	if err != nil {
		return err
	}
	cols := []int{ses.Col("cost"), ses.Col("tokens_input"), ses.Col("tokens_output"), ses.Col("tokens_reasoning")}
	return ses.Recorrer(func(fila []sqlitero.Valor) error {
		var v [4]float64
		for i, col := range cols {
			if col >= 0 {
				v[i] = fila[col].Num()
			}
		}
		c.sumar(1, v[0], int64(v[1]), int64(v[2]), int64(v[3]))
		return nil
	})
}
