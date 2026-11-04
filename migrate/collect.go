package migrate

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing/fstest"
)

// Collector gathers the migrations of a project that keeps each one with the
// code owning its table. See the package comment for the layout and for why
// the answer is flat.
type Collector struct {
	// Root is the directory the walk starts at. Empty is ".".
	Root string

	// Dir is the name of the directory holding migrations, looked for at
	// any depth under Root. A name with a "/" in it is matched as a path
	// suffix, so "db/schema" finds one/db/schema and not one/schema.
	Dir string

	// Pattern selects the migrations inside each one. Empty is Pattern.
	Pattern string

	// Skip is the directories the walk does not enter, named the same way
	// as Dir. A tree carrying a build directory, a vendored copy or a data
	// directory the program itself writes has migrations in it that are not
	// the project's, and those are what Skip is for. Dot directories are
	// skipped without being named.
	Skip []string
}

// Collect returns the migrations as one filesystem, each file at its base name.
//
// It fails with ErrRootedPattern for a pattern naming a directory, ErrNoDir for
// a Dir that is empty or that no directory under Root is called, and
// ErrDuplicate for two directories carrying one migration.
func (c Collector) Collect() (fs.FS, error) {
	pattern := c.Pattern
	if pattern == "" {
		pattern = Pattern
	}
	if strings.Contains(pattern, "/") {
		return nil, fmt.Errorf("%w: %q", ErrRootedPattern, pattern)
	}

	dir := strings.Trim(c.Dir, "/")
	if dir == "" {
		return nil, fmt.Errorf("%w: %q", ErrNoDir, c.Dir)
	}

	root := c.Root
	if root == "" {
		root = "."
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
		if !entry.IsDir() || name == "." {
			return nil
		}
		if strings.HasPrefix(path.Base(name), ".") || c.skipped(name) {
			return fs.SkipDir
		}
		if !named(name, dir) {
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

// skipped reports whether name is one of the directories Skip names.
func (c Collector) skipped(name string) bool {
	for _, skip := range c.Skip {
		if skip = strings.Trim(skip, "/"); skip != "" && named(name, skip) {
			return true
		}
	}
	return false
}

// named reports whether the path name is called dir: equal to it, or ending in
// it on a path boundary.
func named(name, dir string) bool {
	return name == dir || strings.HasSuffix(name, "/"+dir)
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
