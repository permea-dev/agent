package transport

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/permea-dev/agent/internal/config"
	"github.com/permea-dev/agent/internal/event"
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

// Drain drena la cola pendiente hacia el backend (T030): carga los eventos, los envía
// en lotes por HTTPS y, tras cada 2xx, los confirma; al terminar reescribe queue.jsonl
// conservando solo los no confirmados (T028). Devuelve el nº de eventos confirmados.
// La deduplicación extremo a extremo se apoya en event_id. Una cola vacía no es error.
func Drain(dir string, cfg config.Config) (int, error) {
	events, err := Load(dir)
	if err != nil {
		return 0, err
	}
	if len(events) == 0 {
		return 0, nil
	}
	return drain(dir, New(cfg.Endpoint, cfg.DeviceToken), events)
}

// drain es el núcleo testeable de Drain: envía events en lotes con client, confirma los
// aceptados y reescribe la cola atómicamente conservando los no confirmados. Un error de
// envío (auth / 4xx / red tras agotar reintentos) detiene el drenaje, pero los lotes ya
// confirmados se eliminan igualmente de la cola antes de devolver el error.
func drain(dir string, client *Client, events []event.Event) (int, error) {
	confirmed := make(map[string]bool, len(events))
	var syncErr error

	size := client.BatchSize
	if size <= 0 {
		size = defaultBatchSize
	}
	for start := 0; start < len(events); start += size {
		end := start + size
		if end > len(events) {
			end = len(events)
		}
		batch := events[start:end]
		if err := client.sendWithRetry(batch); err != nil {
			syncErr = err
			break // auth/4xx/red-agotada: detener; lo confirmado se poda igualmente
		}
		for _, ev := range batch {
			confirmed[ev.EventID] = true
		}
	}

	// Reescritura atómica (T028): conservar SOLO los no confirmados, en orden.
	keep := make([]event.Event, 0, len(events)-len(confirmed))
	for _, ev := range events {
		if !confirmed[ev.EventID] {
			keep = append(keep, ev)
		}
	}
	if err := rewriteQueue(dir, keep); err != nil {
		return len(confirmed), errors.Join(syncErr, err)
	}
	return len(confirmed), syncErr
}

// rewriteQueue reescribe queue.jsonl con exactamente los eventos keep, de forma atómica:
// se vuelca a un temporal en el MISMO directorio y se hace os.Rename (mismo sistema de
// ficheros). Nunca hay borrado in-place; una caída deja la cola vieja o la nueva, jamás
// un estado a medias (durabilidad, FR-007). Con keep vacío queda un fichero de 0 bytes.
func rewriteQueue(dir string, keep []event.Event) error {
	var buf bytes.Buffer
	for _, ev := range keep {
		b, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	return atomicWrite(QueuePath(dir), buf.Bytes(), 0o600)
}

// atomicWrite escribe data en path vía temporal en el mismo directorio + os.Rename.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Limpieza best-effort: tras un rename correcto el temporal ya no existe.
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Chmod(perm); err != nil {
		return errors.Join(err, tmp.Close())
	}
	// Close DEBE comprobarse: un fallo aquí puede dejar datos sin volcar; en ese caso NO
	// se hace rename y se propaga el error (durabilidad, FR-007).
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
