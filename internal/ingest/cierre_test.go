package ingest

import (
	"fmt"
	"testing"
	"time"
)

// ═══ P-007 B4 · CUÁNDO ESTÁ CERRADO UN MENSAJE (FR-010, FR-013, FR-021) ═══════════════════════════════
//
// Un mensaje se emite cuando está cerrado:
//   - (i) ha empezado otro mensaje posterior en el MISMO fichero;
//   - (ii) han pasado ≥ T (10 min) desde el `timestamp` de su última línea Y desde la última modificación de
//     su fichero;
//   - (iii) han pasado ≥ 24 h desde el `timestamp` de su última línea, cambie o no su fichero (E-3).
// Lo que no está cerrado se queda abierto, y la pasada devuelve el comienzo de su primera línea. Una línea de
// un mensaje ya cerrado es TARDÍA: se cuenta y no se emite. Identificadores sintéticos (disciplina 9).

// t0 es el `timestamp` de las líneas de estos tests.
var t0 = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

// lineaEn es una línea del mensaje `sufijo` con salida `salida` y `timestamp` `momento`.
func lineaEn(sufijo string, salida int, momento time.Time) []byte {
	return []byte(fmt.Sprintf(`{"type":"assistant","timestamp":%q,"sessionId":"s","cwd":"/tmp/x",`+
		`"requestId":"req_CIERRE%021s","message":{"id":"msg_CIERRE%021s","model":"claude-opus-5-5","usage":{`+
		`"input_tokens":1,"output_tokens":%d,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`,
		momento.Format(time.RFC3339Nano), sufijo, sufijo, salida))
}

// leerFichero pasa las líneas por la pasada como un fichero que empieza en el byte 0, con su comienzo y sin
// releer, y lo cierra con el reloj dado.
func leerFichero(t *testing.T, p *Pasada, ahora, modificado time.Time, lineas ...[]byte) (cerrados []Cerrado, abierto int64, hay bool) {
	t.Helper()
	ctx := Context{Salt: "s", Pasada: p}
	var inicio int64
	for _, l := range lineas {
		p.Situar(inicio, false)
		ev, err := FromClaudeCodeLine(l, ctx)
		if err != nil || ev != nil {
			t.Fatalf("precondición: con pasada, la línea se acumula (ev=%v, err=%v)", ev, err)
		}
		inicio += int64(len(l)) + 1
	}
	return p.CerrarConReloj(ahora, modificado)
}

// (i) · P-007 FR-010 — empezar otro mensaje en el MISMO fichero cierra el anterior; en OTRO fichero, no.
func TestCierre_ReglaI(t *testing.T) {
	t.Run("mismo_fichero", func(t *testing.T) {
		p := NuevaPasada()
		a, b := lineaEn("1", 10, t0), lineaEn("2", 20, t0)
		cerrados, abierto, hay := leerFichero(t, p, t0, t0, a, b)
		if len(cerrados) != 1 || cerrados[0].Evento.TokensOutput != 10 {
			t.Errorf("se esperaba cerrado sólo A (salida 10); cerrados = %d", len(cerrados))
		}
		if !hay || abierto != int64(len(a))+1 {
			t.Errorf("B debe quedar abierto, desde el byte %d; abierto = %d, hay = %v", len(a)+1, abierto, hay)
		}
		if got := p.Recuentos().EnEspera; got != 1 {
			t.Errorf("EnEspera = %d; se esperaba 1", got)
		}
	})
	t.Run("otro_fichero", func(t *testing.T) {
		p := NuevaPasada()
		c1, _, hay1 := leerFichero(t, p, t0, t0, lineaEn("3", 10, t0))
		c2, _, hay2 := leerFichero(t, p, t0, t0, lineaEn("4", 20, t0))
		if len(c1)+len(c2) != 0 || !hay1 || !hay2 {
			t.Errorf("un mensaje de otro fichero no cierra el anterior: cerrados %d + %d, abiertos %v y %v", len(c1), len(c2), hay1, hay2)
		}
	})
}

// cierraAl lee un mensaje de una línea con `timestamp` `momento`, y dice si se cierra a la hora `ahora` con el
// fichero modificado en `modificado`.
func cierraAl(t *testing.T, momento, ahora, modificado time.Time) bool {
	t.Helper()
	cerrados, _, hay := leerFichero(t, NuevaPasada(), ahora, modificado, lineaEn("5", 10, momento))
	if len(cerrados) == 1 == hay {
		t.Fatalf("precondición: el mensaje o se cierra o queda abierto (cerrados %d, abierto %v)", len(cerrados), hay)
	}
	return len(cerrados) == 1
}

// (ii) · P-007 FR-010, SC-007 — T desde la última línea Y desde la última modificación del fichero.
func TestCierre_ReglaII(t *testing.T) {
	const T = 10 * time.Minute
	t.Run("a_las_dos_a_T_emitido", func(t *testing.T) {
		if !cierraAl(t, t0, t0.Add(T), t0) {
			t.Errorf("con el timestamp y el mtime a T, el mensaje debe cerrarse")
		}
	})
	t.Run("b_mtime_a_T_menos_1s_retenido", func(t *testing.T) {
		if cierraAl(t, t0, t0.Add(T), t0.Add(time.Second)) {
			t.Errorf("con el mtime a T − 1 s, el mensaje debe seguir abierto")
		}
	})
	t.Run("c_timestamp_a_T_menos_1s_retenido", func(t *testing.T) {
		if cierraAl(t, t0.Add(time.Second), t0.Add(T), t0) {
			t.Errorf("con el timestamp a T − 1 s, el mensaje debe seguir abierto")
		}
	})
}

// (iii) · P-007 FR-010, SC-007 (E-3) — 24 h desde la última línea cierran, aunque el fichero acabe de cambiar.
func TestCierre_ReglaIII(t *testing.T) {
	ahora := t0.Add(24 * time.Hour)
	t.Run("a_24h_emitido", func(t *testing.T) {
		if !cierraAl(t, t0, ahora, ahora) {
			t.Errorf("a las 24 h de su última línea, el mensaje debe cerrarse aunque el fichero cambie")
		}
	})
	t.Run("a_24h_menos_1s_retenido", func(t *testing.T) {
		if cierraAl(t, t0.Add(time.Second), ahora, ahora) {
			t.Errorf("a las 24 h − 1 s, con el fichero recién cambiado, el mensaje debe seguir abierto")
		}
	})
}

// (21) · P-007 FR-013 — una línea de A después de que empiece B es TARDÍA: se cuenta, no se emite y no mueve
// el máximo de A.
func TestCierre_LineaTardia(t *testing.T) {
	p := NuevaPasada()
	cerrados, _, _ := leerFichero(t, p, t0, t0, lineaEn("6", 10, t0), lineaEn("7", 20, t0), lineaEn("6", 999, t0))
	if len(cerrados) != 1 || cerrados[0].Evento.TokensOutput != 10 {
		t.Errorf("A debe salir con la salida que tenía al empezar B (10); cerrados = %d", len(cerrados))
	}
	if got := p.Recuentos().Tardias; got != 1 {
		t.Errorf("Tardias = %d; se esperaba 1", got)
	}
}

// P-007 FR-021 — las líneas releídas se cuentan aparte, y una pasada que sólo relee no tiene novedades.
func TestCierre_Releidas(t *testing.T) {
	p := NuevaPasada()
	ctx := Context{Salt: "s", Pasada: p}
	for i, l := range [][]byte{lineaEn("8", 10, t0), lineaEn("8", 20, t0)} {
		p.Situar(int64(i*1000), true)
		if _, err := FromClaudeCodeLine(l, ctx); err != nil {
			t.Fatalf("precondición: %v", err)
		}
	}
	_, _, _ = p.CerrarConReloj(t0, t0)
	if got := p.Recuentos().Releidas; got != 2 {
		t.Errorf("Releidas = %d; se esperaba 2", got)
	}
	if p.HayNovedades() {
		t.Errorf("una pasada que sólo relee y no emite no tiene novedades (FR-021)")
	}
}
