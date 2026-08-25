package migrate

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// Files returns the names of the migrations in fsys matching pattern, sorted.
// An empty pattern is Pattern, which finds them wherever they sit.
//
// Names are paths relative to the root of fsys, and that is what a run records
// in the filename column. Files shorter than two bytes are skipped: the
// shortest valid statement is "--", a comment.
func Files(fsys fs.FS, pattern string) ([]string, error) {
	if fsys == nil {
		return nil, ErrNoFS
	}
	if pattern == "" {
		pattern = Pattern
	}

	// Check the pattern before the walk, so a malformed one is the same
	// error whether or not the filesystem holds a name to try it against.
	if _, err := path.Match(pattern, ""); err != nil {
		return nil, fmt.Errorf("migrations pattern %q: %w", pattern, err)
	}

	// A pattern naming a directory is matched against the whole path, and
	// so says where the migrations are. One that doesn't is matched against
	// the base name, and so finds them wherever the walk goes.
	rooted := strings.Contains(pattern, "/")

	result := []string{}
	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		target := name
		if !rooted {
			target = path.Base(name)
		}

		if ok, _ := path.Match(pattern, target); !ok {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() < 2 {
			return nil
		}

		result = append(result, name)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// WalkDir walks in lexical order per directory, which is not lexical
	// order of the paths it produces.
	sort.Strings(result)

	return result, nil
}

// Statements returns the statements of one migration, with comments removed and
// uuid() calls replaced by a fresh UUID literal.
func Statements(fsys fs.FS, name string) ([]string, error) {
	if fsys == nil {
		return nil, ErrNoFS
	}

	contents, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", name, err)
	}

	return split(contents), nil
}

// Print writes every migration in fsys to w as SQL, each file headed by a
// comment naming it and each statement by its index, so the output is a script
// applying the project by hand. It touches no database.
//
// An empty pattern is Pattern, as it is for Files.
//
// Because Statements replaces uuid() with a fresh UUID, two calls do not
// produce the same bytes for a migration using it.
func Print(w io.Writer, fsys fs.FS, pattern string) error {
	files, err := Files(fsys, pattern)
	if err != nil {
		return err
	}

	for _, name := range files {
		stmts, err := Statements(fsys, name)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "-- Migrations file: %s\n", name); err != nil {
			return err
		}
		for idx, stmt := range stmts {
			if _, err := fmt.Fprintf(w, "\n-- Statement index: %d\n%s;\n", idx, stmt); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	return nil
}
