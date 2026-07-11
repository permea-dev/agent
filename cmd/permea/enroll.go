package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/transport"
)

// runEnroll es el punto de entrada del subcomando `permea enroll`. Resuelve las
// dependencias reales (stdin, su naturaleza pipe/TTY y stdout) y delega en enroll. El
// enrollment string —y el token que contiene— es un secreto del mismo calibre que el
// salt: NUNCA se hace eco ni aparece en errores.
func runEnroll(args []string) error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("no se pudo inspeccionar stdin: %w", err)
	}
	// Un pipe/fichero NO tiene el bit ModeCharDevice; una TTY interactiva sí.
	stdinIsPipe := fi.Mode()&os.ModeCharDevice == 0
	return enroll(args, os.Stdin, stdinIsPipe, os.Stdout, defaultVerify)
}

// defaultVerify hace el ping real de lote vacío contra el `/ingest` del endpoint,
// reutilizando el contrato de transporte (sin endpoint nuevo).
func defaultVerify(endpoint, token string) error {
	return transport.New(endpoint, token).Verify()
}

// enroll implementa el flujo de enrolamiento con dependencias inyectables (para test):
// resuelve la entrada (argv o stdin), decodifica el enrollment string, verifica el token
// con un ping de lote vacío y —solo si el backend acepta (2xx)— persiste endpoint+token
// en config.json a 0600. En cualquier error NO persiste y el error NUNCA incluye el token.
func enroll(args []string, stdin io.Reader, stdinIsPipe bool, stdout io.Writer, verify func(endpoint, token string) error) error {
	raw, err := readEnrollmentInput(args, stdin, stdinIsPipe)
	if err != nil {
		return err
	}

	// La decodificación valida el envoltorio y que el endpoint sea https; su error no
	// reproduce el argumento ni el token (FR-007/FR-013).
	endpoint, token, err := config.ParseEnrollmentString(raw)
	if err != nil {
		return err
	}

	// Verificar ANTES de persistir. El error del transporte no contiene el token.
	if err := verify(endpoint, token); err != nil {
		return fmt.Errorf("no se pudo verificar el enrolamiento: %w", err)
	}

	dir, err := config.DataDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "config.json")
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}
	cfg.Endpoint = endpoint
	cfg.DeviceToken = token
	if err := config.Save(path, cfg); err != nil {
		return err
	}

	// Confirmación: la URL sí; el token NUNCA. Un fallo al escribir la confirmación no
	// revierte el enrolamiento ya persistido, así que se ignora explícitamente (como printVersion).
	_, _ = fmt.Fprintf(stdout, "enrolado contra %s\n", endpoint)
	return nil
}

// readEnrollmentInput resuelve el enrollment string por dos vías: argv si hay un
// argumento distinto de "-"; stdin si el argumento es "-" (fuerza stdin) o si no hay
// argumento y stdin es un pipe. Sin argumento y con stdin no-pipe (TTY interactiva) →
// error de uso, sin leer (no se cuelga). El valor se recorta; NUNCA se hace eco.
func readEnrollmentInput(args []string, stdin io.Reader, stdinIsPipe bool) (string, error) {
	if len(args) >= 1 && args[0] != "-" {
		return strings.TrimSpace(args[0]), nil
	}
	forceStdin := len(args) >= 1 && args[0] == "-"
	if !forceStdin && !stdinIsPipe {
		return "", fmt.Errorf("uso: permea enroll <enrollment-string>  (recomendado: pásalo por stdin, p. ej. `… | permea enroll -`)")
	}
	b, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el enrollment string de stdin: %w", err)
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return "", fmt.Errorf("no se recibió ningún enrollment string por stdin")
	}
	return s, nil
}
