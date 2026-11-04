package migrate

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing/fstest"
)

// Collect returns the migrations of every directory named dir under root, as
// one filesystem with each file at its base name. An empty pattern is Pattern.
// See the package comment for the layout it is for and why it answers flat.
//
// It fails with ErrRootedPattern for a pattern naming a directory, ErrNoDir for
// a name that is empty or that no directory under root is called, and
// ErrDuplicate for two directories carrying one migration.
func Collect(root, dir, pattern string) (fs.FS, error) {
	if pattern == "" {
		pattern = Pattern
	}
	if strings.Contains(pattern, "/") {
		return nil, fmt.Errorf("%w: %q", ErrRootedPattern, pattern)
	}
	if dir = strings.Trim(dir, "/"); dir == "" {
		return nil, fmt.Errorf("%w: %q", ErrNoDir, dir)
	}

	merged := fstest.MapFS{}
	found := 0
	fsys := os.DirFS(root)

	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			// A directory that cannot be read is skipped and the walk
			// goes on; the root failing is the one error that stands.
			if name == "." {
				return err
			}
			return fs.SkipDir
		}
		if !entry.IsDir() {
			return nil
		}
		if name != "." && strings.HasPrefix(path.Base(name), ".") {
			return fs.SkipDir
		}
		if name != dir && !strings.HasSuffix(name, "/"+dir) {
			return nil
		}
		found++
		return collect(merged, fsys, name, pattern)
	})
	if err != nil {
		return nil, err
	}
	if found == 0 {
		return nil, fmt.Errorf("%w: %q under %q", ErrNoDir, dir, root)
	}

	return merged, nil
}

// collect reads the migrations of one directory into merged.
func collect(merged fstest.MapFS, fsys fs.FS, dir, pattern string) error {
	names, err := Files(subFS(fsys, dir), pattern)
	if err != nil {
		return err
	}
	for _, name := range names {
		base := path.Base(name)
		if _, taken := merged[base]; taken {
			return fmt.Errorf("%w: %s", ErrDuplicate, path.Join(dir, name))
		}
		contents, err := fs.ReadFile(fsys, path.Join(dir, name))
		if err != nil {
			return err
		}
		merged[base] = &fstest.MapFile{Data: contents}
	}
	return nil
}

// subFS roots fsys at dir, and answers fsys itself for the root.
func subFS(fsys fs.FS, dir string) fs.FS {
	if dir == "." {
		return fsys
	}
	at, err := fs.Sub(fsys, dir)
	if err != nil {
		return fsys
	}
	return at
}
