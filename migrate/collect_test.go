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

	fsys, err := migrate.Collect(root, "schema", "")
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

	fsys, err := migrate.Collect(root, "db/schema", "")
	require.NoError(t, err)
	require.Equal(t, []string{"a.up.sql", "b.up.sql"}, names(t, fsys))
}

// The root itself counts, so a flat project reads the same either way.
func TestCollectRootDirectory(t *testing.T) {
	root := tree(t, map[string]string{"schema/a.up.sql": "CREATE TABLE a"})

	_, err := migrate.Collect(filepath.Join(root, "schema"), "schema", "")
	require.ErrorIs(t, err, migrate.ErrNoDir, "the root is not named by its own path")

	fsys, err := migrate.Collect(root, "schema", "")
	require.NoError(t, err)
	require.Equal(t, []string{"a.up.sql"}, names(t, fsys))
}

// A glob of its own selects the files, and one naming a directory is refused.
func TestCollectPattern(t *testing.T) {
	root := tree(t, map[string]string{
		"a/schema/one.sql":    "CREATE TABLE one",
		"a/schema/two.up.sql": "CREATE TABLE two",
	})

	fsys, err := migrate.Collect(root, "schema", "*.sql")
	require.NoError(t, err)

	// Read back under the same pattern: the default one would hide the file
	// the pattern was given for.
	got, err := migrate.Files(fsys, "*.sql")
	require.NoError(t, err)
	require.Equal(t, []string{"one.sql", "two.up.sql"}, got)

	_, err = migrate.Collect(root, "schema", "schema/*.up.sql")
	require.ErrorIs(t, err, migrate.ErrRootedPattern)
}

// Two directories carrying one base name is an error. One would shadow the
// other and the table it creates would never exist.
func TestCollectDuplicate(t *testing.T) {
	root := tree(t, map[string]string{
		"one/schema/user.up.sql": "CREATE TABLE user",
		"two/schema/user.up.sql": "CREATE TABLE user",
	})

	_, err := migrate.Collect(root, "schema", "")
	require.ErrorIs(t, err, migrate.ErrDuplicate)
	require.Contains(t, err.Error(), "user.up.sql")
}

// A directory nothing is called, and a name that is no name, are the same
// misconfiguration: a run that applies nothing looks exactly like success.
func TestCollectErrors(t *testing.T) {
	root := tree(t, map[string]string{"a/sql/a.up.sql": "CREATE TABLE a"})

	_, err := migrate.Collect(root, "schema", "")
	require.ErrorIs(t, err, migrate.ErrNoDir)

	_, err = migrate.Collect(root, "", "")
	require.ErrorIs(t, err, migrate.ErrNoDir)

	_, err = migrate.Collect(root, "/", "")
	require.ErrorIs(t, err, migrate.ErrNoDir)

	_, err = migrate.Collect(filepath.Join(root, "nope"), "schema", "")
	require.Error(t, err, "a root that is not there")
}

// A directory matching the name but holding no migration is still a directory
// that was found: the project is laid out right and has nothing to apply yet,
// which Apply reports as ErrNoMigrations.
func TestCollectEmptyDirectory(t *testing.T) {
	root := tree(t, map[string]string{"a/schema/README.md": "nothing to apply"})

	fsys, err := migrate.Collect(root, "schema", "")
	require.NoError(t, err)
	require.Empty(t, names(t, fsys))
}
