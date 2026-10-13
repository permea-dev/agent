package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/permea-dev/agent/internal/ingest"
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
