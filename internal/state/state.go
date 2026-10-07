// Package state persiste el progreso del escaneo incremental (FR-006): una llamada
// ya emitida queda por debajo del offset y NUNCA se reprocesa.
package state

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileState guarda el progreso por fichero para no reprocesar llamadas ya emitidas.
type FileState struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
	Offset  int64  `json:"offset"`
}

// Store es el conjunto persistido de estados por fichero.
type Store struct {
	Files map[string]FileState `json:"files"`
}

// New crea un Store vacío.
func New() *Store { return &Store{Files: make(map[string]FileState)} }

// Load lee state.json; un fichero inexistente devuelve un Store vacío sin error.
func Load(path string) (*Store, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, err
	}
	s := New()
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}
	if s.Files == nil {
		s.Files = make(map[string]FileState)
	}
	return s, nil
}

// Save persiste state.json de forma atómica: temporal + os.Rename en el mismo
// directorio (mismo sistema de ficheros), evitando estado corrupto ante caída.
func (s *Store) Save(path string) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".state-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Limpieza best-effort: tras un rename correcto el temporal ya no existe.
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(b); err != nil {
		return errors.Join(err, tmp.Close())
	}
	// Close DEBE comprobarse: un fallo puede dejar datos sin volcar; en ese caso NO se
	// hace rename y se devuelve el error (durabilidad del estado incremental, FR-007).
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// ScanFile lee las líneas COMPLETAS nuevas de path desde el offset guardado e invoca
// fn por cada una, avanzando el offset solo hasta el fin de la última línea completa.
// Detecta truncado/rotación (size < offset -> relee desde 0). Una línea parcial (el
// fichero se está escribiendo) no se cuenta hasta terminar en '\n'. Solo stdlib.
//
// P-007: es un envoltorio de `Recorrer` que guarda el offset de siempre, el final de lo leído.
func (s *Store) ScanFile(path string, fn func(line []byte) error) error {
	return s.Recorrer(path,
		func(line []byte, _ int64, _ bool) error { return fn(line) },
		func(leido int64, _ time.Time) int64 { return leido })
}

// Recorrer lee las líneas COMPLETAS nuevas de path desde el offset guardado, como ScanFile, con dos
// diferencias (P-007 FR-011, FR-021):
//
//   - fn recibe además el COMIENZO de cada línea en el fichero, y si es RELEÍDA: si empieza por debajo del
//     tamaño que tenía el fichero en la pasada anterior, ya se leyó entonces;
//   - el offset que se guarda lo decide quien llama: `fijar` recibe el fin de la última línea completa
//     leída y la última modificación del fichero, del stat de ESTA pasada, y devuelve el offset que se
//     guarda. Así un mensaje abierto no se consume: el offset se queda en su comienzo y la pasada
//     siguiente lo relee del log.
//
// «Releída» es una APROXIMACIÓN, porque el `Size` anterior es el del stat de la pasada anterior, no lo que
// esa pasada leyó. Tiene dos imprecisiones:
//   - una línea que estaba a medio escribir en la pasada anterior se cuenta como releída sin haberse leído;
//   - si el fichero creció entre el stat y la lectura, una línea ya leída se cuenta como nueva.
//
// Sólo afectan al recuento de releídas y a si el demonio escribe el resumen de ese ciclo. NUNCA a lo que se
// emite: la marca no entra en la acumulación ni en el cierre de los mensajes.
func (s *Store) Recorrer(path string, fn func(line []byte, inicio int64, releida bool) error,
	fijar func(leido int64, modificado time.Time) int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	anterior := s.Files[path]
	offset, leidoAntes := anterior.Offset, anterior.Size
	if info.Size() < offset {
		offset, leidoAntes = 0, 0 // truncado o rotado: releer desde el principio; nada de esto se leyó antes
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }() // solo lectura: el error de Close no afecta a datos
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	consumed := offset
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			// Sin '\n' final: línea parcial. No se consume; se reintenta en la próxima pasada.
			break
		}
		if err != nil {
			return err
		}
		inicio := consumed
		consumed += int64(len(line))
		if cbErr := fn(line, inicio, inicio < leidoAntes); cbErr != nil {
			return cbErr
		}
	}
	// El offset pedido nunca sale de lo que ya estaba guardado y lo leído ahora.
	guardado := min(max(fijar(consumed, info.ModTime()), offset), consumed)

	s.Files[path] = FileState{
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime().Unix(),
		Offset:  guardado,
	}
	return nil
}
