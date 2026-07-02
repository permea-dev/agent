package state

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// FindLogs enumera recursivamente los ficheros *.jsonl bajo root (los logs de
// Claude Code en ~/.claude/projects/**), devolviéndolos en orden estable. Ignora
// cualquier otro fichero. Una raíz inexistente (Claude Code nunca usado) devuelve
// una lista vacía sin error (edge case del spec), no un fallo.
func FindLogs(root string) ([]string, error) {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}

	var paths []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(p) == ".jsonl" {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}
