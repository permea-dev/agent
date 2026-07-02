// Package state persiste el progreso del escaneo incremental (FR-004).
package state

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
