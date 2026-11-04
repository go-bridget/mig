package migrate

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/cli"

	"github.com/go-bridget/mig/migrate"
)

// write puts the files of a map under a fresh directory, chdirs into it for the
// test, and returns it. The recursive mode reads the working directory, so a
// test of it has to have one of its own.
func write(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, contents := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
	}
	t.Chdir(root)
	return root
}

// TestNew covers the wiring: the flags the command binds reach open, and a
// print needs no database, so the whole path from an argument list to the
// migrations found is one call.
func TestNew(t *testing.T) {
	write(t, map[string]string{
		"access/schema/user.up.sql":       "CREATE TABLE user;",
		"keystore/schema/key_info.up.sql": "CREATE TABLE key_info;",
	})

	cmd := New()
	require.Equal(t, "migrate", cmd.Name)

	flags := cli.NewFlagSet("migrate", cli.ContinueOnError)
	cmd.Bind(flags)
	require.NoError(t, flags.Parse([]string{"--recursive", "--path", "schema"}))

	require.NoError(t, cmd.Run(t.Context(), []string{"events"}))
}

// TestOpenPath is the default: the path is one directory, and the names a run
// records are relative to it.
func TestOpenPath(t *testing.T) {
	write(t, map[string]string{
		"schema/user.up.sql": "CREATE TABLE user",
		"schema/run.up.sql":  "CREATE TABLE run",
	})

	fsys, glob, err := options{path: "schema", glob: migrate.Pattern}.open()
	require.NoError(t, err)
	require.Equal(t, migrate.Pattern, glob)

	names, err := migrate.Files(fsys, glob)
	require.NoError(t, err)
	require.Equal(t, []string{"run.up.sql", "user.up.sql"}, names)
}

// A path that is a file, or is not there at all, is an error rather than a run
// of nothing.
func TestOpenPathErrors(t *testing.T) {
	write(t, map[string]string{"schema/user.up.sql": "CREATE TABLE user"})

	_, _, err := options{path: "schema/user.up.sql", glob: migrate.Pattern}.open()
	require.ErrorContains(t, err, "not a directory")

	_, _, err = options{path: "nope", glob: migrate.Pattern}.open()
	require.Error(t, err)
}

// TestOpenRecursive covers the flag: the path is a directory name, every one of
// them under the working directory is found at whatever depth, and the names a
// run records are the base names - the same ones one folder of all of them
// would have recorded.
func TestOpenRecursive(t *testing.T) {
	write(t, map[string]string{
		"access/schema/user.up.sql":         "CREATE TABLE user",
		"keystore/schema/key_info.up.sql":   "CREATE TABLE key_info",
		"app/modules/notes/schema/n.up.sql": "CREATE TABLE n",
		"app/modules/notes/templates/i.tpl": "not a migration",
	})

	fsys, glob, err := options{path: "schema", glob: migrate.Pattern, recursive: true}.open()
	require.NoError(t, err)
	require.Equal(t, migrate.Pattern, glob)

	names, err := migrate.Files(fsys, glob)
	require.NoError(t, err)
	require.Equal(t, []string{"key_info.up.sql", "n.up.sql", "user.up.sql"}, names)

	contents, err := fs.ReadFile(fsys, "user.up.sql")
	require.NoError(t, err)
	require.Equal(t, "CREATE TABLE user", string(contents))
}

// A directory name nothing under the working directory is called is an error,
// because applying nothing is the one failure that looks like success.
func TestOpenRecursiveMissing(t *testing.T) {
	write(t, map[string]string{"access/sql/user.up.sql": "CREATE TABLE user"})

	_, _, err := options{path: "schema", glob: migrate.Pattern, recursive: true}.open()
	require.ErrorIs(t, err, migrate.ErrNoDir)
}

// A filename wins over both modes, and brings its own glob so the file is
// applied whatever it is called.
func TestOpenFilename(t *testing.T) {
	root := write(t, map[string]string{
		"schema/user.up.sql": "CREATE TABLE user",
		"one-off.sql":        "ALTER TABLE user",
	})

	fsys, glob, err := options{
		path:      "schema",
		filename:  filepath.Join(root, "one-off.sql"),
		glob:      migrate.Pattern,
		recursive: true,
	}.open()
	require.NoError(t, err)
	require.Equal(t, "one-off.sql", glob)

	names, err := migrate.Files(fsys, glob)
	require.NoError(t, err)
	require.Equal(t, []string{"one-off.sql"}, names)
}
