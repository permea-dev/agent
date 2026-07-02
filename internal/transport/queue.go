package transport

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/bfgnet/agente_permea/internal/event"
)

const queueFile = "queue.jsonl"

// QueuePath devuelve la ruta del fichero de cola dentro del directorio de datos.
func QueuePath(dir string) string { return filepath.Join(dir, queueFile) }

// Append añade un evento a la cola append-only (una línea JSON por evento). El
// evento queda DURABLE en disco antes de intentar transmitirse (FR-007): una caída
// de red o de proceso no lo pierde. El drenaje/transmisión se implementa en US2.
func Append(dir string, ev event.Event) error {
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(QueuePath(dir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		return errors.Join(err, f.Close())
	}
	// Close DEBE comprobarse: confirma el volcado del evento a la cola durable antes de
	// darlo por encolado; un fallo aquí se propaga sin perder el error (FR-007).
	return f.Close()
}

// Load lee todos los eventos pendientes de la cola en orden. Una cola inexistente
// devuelve una lista vacía sin error.
func Load(dir string) ([]event.Event, error) {
	f, err := os.Open(QueuePath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }() // solo lectura: el error de Close no afecta a datos

	var out []event.Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev event.Event
		if err := json.Unmarshal(line, &ev); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, sc.Err()
}
