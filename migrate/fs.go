package migrate

import (
	"os"
	"sort"

	"path/filepath"
)

// FS represents a mapping between filename => contents
type FS map[string][]byte

// NewFS returns a new FS instance
func NewFS() FS {
	return make(FS)
}

// Migrations returns list of SQL files to execute
func (fs FS) Migrations() []string {
	result := []string{}
	for filename, contents := range fs {
		// skip empty files (minimum valid statement is `--`, a comment)
		if len(contents) < 2 {
			continue
		}
		if matched, _ := filepath.Match("*.up.sql", filename); matched {
			result = append(result, filename)
		}
	}
	sort.Strings(result)
	return result
}

// ReadFile returns decoded file contents from FS
func (fs FS) ReadFile(filename string) ([]byte, error) {
	if val, ok := fs[filename]; ok {
		return val, nil
	}
	return nil, os.ErrNotExist
}
