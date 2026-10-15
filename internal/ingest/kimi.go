package ingest

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/permea-dev/agent/internal/event"
)

// ═══ P-012 · UNA LÍNEA DE KIMI CODE ES, COMO MUCHO, UN EVENTO ═════════════════════════════════════════════════════════
//
// Kimi Code escribe un `wire.jsonl` por agente (descubrimiento Q1). Es REGISTRO (D-2) sólo la línea `usage.record`: una por
// petición terminada al modelo, también las de la compactación (`usageScope: session`) y las de un subagente, en su propio fichero
// (muestra 1). Nada más: ni `subagent.completed` (repite la suma del subagente), ni `llm.request`, ni ninguna línea con contenido.
//
// ═══ LO QUE SE LEE, Y LO QUE NO ════════════════════════════════════════════════════════════════════════════════════
//
// PREFILTRO ANCLADO (P-10): una línea sólo se decodifica si EMPIEZA por `{"type":"usage.record"` (Kimi escribe `type` primero:
// descubrimiento Q2; muestra 1, 11/11). El mismo texto dentro de un contenido no está al principio, y no casa. Un heredado ni
// siquiera se decodifica.
//
// De un registro se leen sólo `type`, `model`, `usage.{inputOther, output, inputCacheRead, inputCacheCreation}`, `usageScope` y
// `time` (FR-012). Nunca `agentId`: el agente es el de la carpeta (P-5). La sesión y el `cwd`, sólo con sal (FR-019). El coste no se
// calcula: lo pone la plataforma (D-5, constitución 1.1.0 §II); no se consulta `internal/pricing`.

// ContextoKimi son los datos locales que añade el agente a un registro de Kimi Code (P-012 P-5, P-6): la posición del
// registro (S, A, N), si es heredado de un `/fork`, y el `workDir` de su sesión. Los pone quien lee el fichero (B3).
type ContextoKimi struct {
	Context
	// S es el nombre de la carpeta de la sesión, y A el de la carpeta del agente.
	S, A string
	// N es el ordinal del registro en su fichero, desde 1, contando los heredados.
	N int
	// Heredado: el registro va antes de la última línea `forked` de su fichero (D-3).
	Heredado bool
	// Cwd es el `workDir` de la sesión en `session_index.jsonl`, ya normalizado. Vacío si no hay.
	Cwd string
	// Decodificadas, si no es nil, cuenta las líneas que se decodifican (P-10; test (12)).
	Decodificadas *int
}

// ClaseKimi clasifica una línea de un `wire.jsonl` de Kimi Code (P-012 FR-020).
type ClaseKimi int

const (
	// NoEsRegistroKimi es cualquier línea que no empieza por `{"type":"usage.record"`.
	NoEsRegistroKimi ClaseKimi = iota
	// EventoKimi es un registro que produce un evento.
	EventoKimi
	// HeredadoKimi es un registro anterior a la última `forked` de su fichero (D-3, FR-014).
	HeredadoKimi
	// ACeroKimi es un registro con las cuatro partidas a 0 (P-8).
	ACeroKimi
	// IncoherenteKimi es un registro cuyas partidas o cuyo momento no se pueden usar (FR-020).
	IncoherenteKimi
)

// MarcasKimi son las marcas de un evento (FR-020): subconjuntos de «eventos», no clases.
type MarcasKimi struct {
	// Compactacion: el registro es de una compactación (`usageScope: session`).
	Compactacion bool
	// SinModelo: el modelo, tras P-9, queda vacío.
	SinModelo bool
}

// Los datos del registro, y su prefijo anclado (P-10).
const (
	tipoRegistroKimi     = "usage.record"
	prefijoRegistroKimi  = `{"type":"` + tipoRegistroKimi + `"`
	alcanceSesionKimi    = "session"
	partidaEntradaKimi   = "inputOther"
	partidaSalidaKimi    = "output"
	partidaLecturaKimi   = "inputCacheRead"
	partidaEscrituraKimi = "inputCacheCreation"
)

// lineaKimi es un `usage.record`, todo en crudo: un campo con otra forma clasifica, nunca rompe.
type lineaKimi struct {
	Type       json.RawMessage `json:"type"`
	Model      json.RawMessage `json:"model"`
	Usage      json.RawMessage `json:"usage"`
	UsageScope json.RawMessage `json:"usageScope"`
	Time       json.RawMessage `json:"time"`
}

// RegistroKimi lee una línea de un `wire.jsonl` de Kimi Code y devuelve su clase y, si es un evento, el evento y sus marcas.
//
// Un error es SÓLO una línea que pasó el prefiltro y no es un objeto JSON: quien lee decide qué hacer con ella. Lo demás
// clasifica, en este orden (D-012-P4): no es registro → heredado → incoherente → a cero → evento. Las repetidas las cuenta quien
// recorre el fichero.
func RegistroKimi(linea []byte, ctx ContextoKimi) (*event.Event, ClaseKimi, MarcasKimi, error) {
	if !bytes.HasPrefix(linea, []byte(prefijoRegistroKimi)) {
		return nil, NoEsRegistroKimi, MarcasKimi{}, nil
	}
	if ctx.Heredado {
		return nil, HeredadoKimi, MarcasKimi{}, nil
	}
	if ctx.Decodificadas != nil {
		*ctx.Decodificadas++
	}
	var l lineaKimi
	if err := json.Unmarshal(linea, &l); err != nil {
		return nil, NoEsRegistroKimi, MarcasKimi{}, err
	}
	if textoDe(l.Type) != tipoRegistroKimi {
		return nil, NoEsRegistroKimi, MarcasKimi{}, nil
	}

	var partidas map[string]json.RawMessage
	p, ok := partidasKimiDe(l.Usage, &partidas)
	var ms int64
	if !ok || json.Unmarshal(l.Time, &ms) != nil || ms <= 0 {
		return nil, IncoherenteKimi, MarcasKimi{}, nil
	}
	if p == (partidasKimi{}) {
		return nil, ACeroKimi, MarcasKimi{}, nil
	}

	modelo := modeloKimi(textoDe(l.Model))
	marcas := MarcasKimi{Compactacion: textoDe(l.UsageScope) == alcanceSesionKimi, SinModelo: modelo == ""}
	proyecto := ""
	if ctx.Cwd != "" {
		proyecto = ctx.Resolutor.Derivar(ctx.Cwd, ctx.Salt)
	}
	ev := event.Event{
		SchemaVersion:       event.SchemaVersion,
		AgentVersion:        ctx.AgentVersion,
		EventID:             derivarEventIDKimi(ctx.S, ctx.A, ctx.N),
		OccurredAt:          time.UnixMilli(ms).UTC(),
		Tool:                herramientaKimi,
		Model:               modelo,
		TokensInput:         int(p.otra),
		TokensOutput:        int(p.salida),
		TokensCacheCreation: int(p.escritura),
		TokensCacheRead:     int(p.lectura),
		CostUSD:             0,
		CostAvailable:       false,
		ProjectRef:          proyecto,
		SessionRef:          event.Ref(ctx.Salt, ctx.S),
		MachineRef:          event.Ref(ctx.Salt, ctx.MachineID),
		DevID:               ctx.DevID,
		OrgID:               ctx.OrgID,
	}
	return &ev, EventoKimi, marcas, nil
}

// modeloKimi es el texto tras la ÚLTIMA `/` del alias, o el alias entero si no tiene (P-9, D-4): `moonshot-ai/kimi-k2.7-code` →
// `kimi-k2.7-code`; `a/b/c` → `c`; `/x` → `x`; `x/` → vacío.
func modeloKimi(alias string) string {
	return alias[strings.LastIndexByte(alias, '/')+1:]
}

// partidasKimi son las cuatro partidas de `usage`, tal como las escribe Kimi Code: pasan al evento sin sumas ni restas (FR-013).
type partidasKimi struct {
	otra, salida, lectura, escritura int64
}

// partidasKimiDe lee las cuatro partidas. `ok == false` si `usage` no es un objeto, si faltan las cuatro, o si alguna no es un
// entero o es negativa. Una partida ausente o `null` vale 0.
func partidasKimiDe(crudo json.RawMessage, m *map[string]json.RawMessage) (partidasKimi, bool) {
	var p partidasKimi
	if !objetoDe(crudo, m) {
		return p, false
	}
	presentes := 0
	for clave, destino := range map[string]*int64{
		partidaEntradaKimi: &p.otra, partidaSalidaKimi: &p.salida, partidaLecturaKimi: &p.lectura, partidaEscrituraKimi: &p.escritura,
	} {
		v, esta := (*m)[clave]
		if !esta || ausente(v) {
			continue
		}
		presentes++
		if json.Unmarshal(v, destino) != nil || *destino < 0 {
			return p, false
		}
	}
	return p, presentes > 0
}
