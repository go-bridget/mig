package migrate_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-bridget/mig/migrate"
)

// tree writes the files of a map under a fresh directory and returns it.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, contents := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
	}
	return root
}

// names returns the migrations of fsys, in the order a run applies them.
func names(t *testing.T, fsys fs.FS) []string {
	t.Helper()
	got, err := migrate.Files(fsys, "")
	require.NoError(t, err)
	return got
}

// TestCollect covers the layout it is for: a migration living with the code
// that owns its table, found at whatever depth, answered flat and in the order
// the base names give rather than the order the paths do.
func TestCollect(t *testing.T) {
	root := tree(t, map[string]string{
		"access/schema/user.up.sql":            "CREATE TABLE user",
		"keystore/schema/key_info.up.sql":      "CREATE TABLE key_info",
		"app/modules/notes/schema/z.up.sql":    "CREATE TABLE z",
		"app/modules/runs/schema/a.up.sql":     "CREATE TABLE a",
		"access/schema/user.down.sql":          "DROP TABLE user",
		"access/schema/schema.puml":            "not a migration",
		"access/README.md":                     "not a migration",
		"app/modules/notes/templates/i.tpl":    "not a migration",
		".git/objects/schema/sneaky.up.sql":    "CREATE TABLE sneaky",
		"vendor/dep/other/elsewhere.up.sql":    "CREATE TABLE elsewhere",
		"app/modules/runs/schema/empty.up.sql": "-",
	})

	fsys, err := migrate.Collector{Root: root, Dir: "schema"}.Collect()
	require.NoError(t, err)

	require.Equal(t, []string{"a.up.sql", "key_info.up.sql", "user.up.sql", "z.up.sql"}, names(t, fsys),
		"flat, base-name order, and nothing a dot directory or another folder holds")

	contents, err := fs.ReadFile(fsys, "key_info.up.sql")
	require.NoError(t, err)
	require.Equal(t, "CREATE TABLE key_info", string(contents))
}

// A nested directory name is matched as a path suffix, so a project keeping
// its migrations under db/schema says so.
func TestCollectNestedName(t *testing.T) {
	root := tree(t, map[string]string{
		"one/db/schema/a.up.sql": "CREATE TABLE a",
		"two/db/schema/b.up.sql": "CREATE TABLE b",
		"two/schema/c.up.sql":    "CREATE TABLE c",
	})

	fsys, err := migrate.Collector{Root: root, Dir: "db/schema"}.Collect()
	require.NoError(t, err)
	require.Equal(t, []string{"a.up.sql", "b.up.sql"}, names(t, fsys))
}

// TestCollectSkip covers what a tree holds that is not the project's: the copy
// of a module tree the program itself extracted into its data directory is the
// same migration under another name, and without Skip the walk reads it and
// calls the project a duplicate.
func TestCollectSkip(t *testing.T) {
	root := tree(t, map[string]string{
		"modules/notes/schema/note.up.sql":      "CREATE TABLE note",
		"data/modules/notes/schema/note.up.sql": "CREATE TABLE note",
		"ext/dep/schema/dep.up.sql":             "CREATE TABLE dep",
	})

	_, err := migrate.Collector{Root: root, Dir: "schema"}.Collect()
	require.ErrorIs(t, err, migrate.ErrDuplicate)

	fsys, err := migrate.Collector{Root: root, Dir: "schema", Skip: []string{"data", "/ext/"}}.Collect()
	require.NoError(t, err)
	require.Equal(t, []string{"note.up.sql"}, names(t, fsys))

	// A skip naming nothing changes nothing, and an empty one is not a
	// name that skips the root.
	fsys, err = migrate.Collector{Root: root, Dir: "schema", Skip: []string{"", "nope", "data", "ext"}}.Collect()
	require.NoError(t, err)
	require.Equal(t, []string{"note.up.sql"}, names(t, fsys))
}

// The root is not named by its own path: a project pointed straight at its
// migrations wants the plain --path, not this.
func TestCollectRootDirectory(t *testing.T) {
	root := tree(t, map[string]string{"schema/a.up.sql": "CREATE TABLE a"})

	_, err := migrate.Collector{Root: filepath.Join(root, "schema"), Dir: "schema"}.Collect()
	require.ErrorIs(t, err, migrate.ErrNoDir)

	fsys, err := migrate.Collector{Root: root, Dir: "schema"}.Collect()
	require.NoError(t, err)
	require.Equal(t, []string{"a.up.sql"}, names(t, fsys))
}

// An empty Root is the working directory, which is what the CLI hands it.
func TestCollectDefaultRoot(t *testing.T) {
	t.Chdir(tree(t, map[string]string{"a/schema/a.up.sql": "CREATE TABLE a"}))

	fsys, err := migrate.Collector{Dir: "schema"}.Collect()
	require.NoError(t, err)
	require.Equal(t, []string{"a.up.sql"}, names(t, fsys))
}

// A glob of its own selects the files, and one naming a directory is refused.
func TestCollectPattern(t *testing.T) {
	root := tree(t, map[string]string{
		"a/schema/one.sql":    "CREATE TABLE one",
		"a/schema/two.up.sql": "CREATE TABLE two",
	})

	fsys, err := migrate.Collector{Root: root, Dir: "schema", Pattern: "*.sql"}.Collect()
	require.NoError(t, err)

	// Read back under the same pattern: the default one would hide the file
	// the pattern was given for.
	got, err := migrate.Files(fsys, "*.sql")
	require.NoError(t, err)
	require.Equal(t, []string{"one.sql", "two.up.sql"}, got)

	_, err = migrate.Collector{Root: root, Dir: "schema", Pattern: "schema/*.up.sql"}.Collect()
	require.ErrorIs(t, err, migrate.ErrRootedPattern)
}

// Two directories carrying one base name is an error. One would shadow the
// other and the table it creates would never exist.
func TestCollectDuplicate(t *testing.T) {
	root := tree(t, map[string]string{
		"one/schema/user.up.sql": "CREATE TABLE user",
		"two/schema/user.up.sql": "CREATE TABLE user",
	})

	_, err := migrate.Collector{Root: root, Dir: "schema"}.Collect()
	require.ErrorIs(t, err, migrate.ErrDuplicate)
	require.Contains(t, err.Error(), "user.up.sql")
}

// A directory nothing is called, and a name that is no name, are the same
// misconfiguration: a run that applies nothing looks exactly like success.
func TestCollectErrors(t *testing.T) {
	root := tree(t, map[string]string{"a/sql/a.up.sql": "CREATE TABLE a"})

	_, err := migrate.Collector{Root: root, Dir: "schema"}.Collect()
	require.ErrorIs(t, err, migrate.ErrNoDir)

	_, err = migrate.Collector{Root: root}.Collect()
	require.ErrorIs(t, err, migrate.ErrNoDir)

	_, err = migrate.Collector{Root: root, Dir: "/"}.Collect()
	require.ErrorIs(t, err, migrate.ErrNoDir)

	_, err = migrate.Collector{Root: filepath.Join(root, "nope"), Dir: "schema"}.Collect()
	require.Error(t, err, "a root that is not there")
}

// A directory matching the name but holding no migration is still a directory
// that was found: the project is laid out right and has nothing to apply yet,
// which Apply reports as ErrNoMigrations.
func TestCollectEmptyDirectory(t *testing.T) {
	root := tree(t, map[string]string{"a/schema/README.md": "nothing to apply"})

	fsys, err := migrate.Collector{Root: root, Dir: "schema"}.Collect()
	require.NoError(t, err)
	require.Empty(t, names(t, fsys))
}
