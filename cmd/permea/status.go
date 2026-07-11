package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/permea-dev/agent/internal/config"
)

// runStatus informa el estado de enrolamiento. Es una operación **local**: deriva el
// estado de la config persistida (endpoint + token + https) y NUNCA contacta al backend
// (a diferencia de enroll). NUNCA imprime el device_token (SC-005): a lo sumo un indicador
// de presencia.
func runStatus(stdout io.Writer) error {
	dir, err := config.DataDir()
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		return err
	}

	if config.IsEnrolled(cfg) {
		_, _ = fmt.Fprintf(stdout, "enrolado contra %s (token: configurado)\n", cfg.Endpoint)
	} else {
		_, _ = fmt.Fprintln(stdout, "no enrolado")
	}
	return nil
}
