package migrate_test

import (
	"bytes"
	"os"
	"path"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/go-bridget/mig/migrate"
)

// TestFiles covers what is a migration and what isn't: the glob picks the
// *.up.sql of the root only, and a file too short to hold a statement is not
// one.
func TestFiles(t *testing.T) {
	dummy := []byte(`-- This is a comment`)

	fsys := fstest.MapFS{
		"foo.txt":          {Data: dummy},
		"file-2.up.sql":    {Data: dummy},
		"file-1.up.sql":    {Data: dummy},
		"empty.up.sql":     {Data: []byte("-")},
		"folder/file2.sql": {Data: dummy},
	}

	got, err := migrate.Files(fsys, "")
	require.NoError(t, err)

	require.Equal(t, []string{"file-1.up.sql", "file-2.up.sql"}, got)
}

// TestFilesFromDir covers the path the CLI takes, which is an os.DirFS and
// nothing else.
func TestFilesFromDir(t *testing.T) {
	got, err := migrate.Files(os.DirFS("../testdata/schema/stats"), "")
	require.NoError(t, err)

	require.Equal(t, []string{"2019-12-13-184604-import-initial-schema.up.sql"}, got)
}

func TestFilesNilFS(t *testing.T) {
	_, err := migrate.Files(nil, "")
	require.ErrorIs(t, err, migrate.ErrNoFS)
}

// TestFilesPatternDepth covers what decides the depth: naming a directory in
// the pattern, or not.
func TestFilesPatternDepth(t *testing.T) {
	dummy := []byte(`-- This is a comment`)

	fsys := fstest.MapFS{
		"root.up.sql":            {Data: dummy},
		"schema/one.up.sql":      {Data: dummy},
		"schema/deep/two.up.sql": {Data: dummy},
		"schema/notes.sql":       {Data: dummy},
	}

	for _, test := range []struct {
		pattern string
		want    []string
	}{
		{
			// No directory named, so base names, so any depth.
			pattern: "",
			want: []string{
				"root.up.sql",
				"schema/deep/two.up.sql",
				"schema/one.up.sql",
			},
		},
		{
			// A directory named, so whole paths, so exactly one
			// level: "*" does not cross a "/".
			pattern: "schema/*.sql",
			want: []string{
				"schema/notes.sql",
				"schema/one.up.sql",
			},
		},
		{
			pattern: "schema/deep/*.up.sql",
			want:    []string{"schema/deep/two.up.sql"},
		},
		{
			pattern: "*.sql",
			want: []string{
				"root.up.sql",
				"schema/deep/two.up.sql",
				"schema/notes.sql",
				"schema/one.up.sql",
			},
		},
	} {
		t.Run(test.pattern, func(t *testing.T) {
			got, err := migrate.Files(fsys, test.pattern)
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

// TestFilesBadPattern covers a malformed pattern being reported even when there
// is no name to try it against.
func TestFilesBadPattern(t *testing.T) {
	_, err := migrate.Files(fstest.MapFS{}, "[unterminated")
	require.ErrorIs(t, err, path.ErrBadPattern)
}

func TestStatements(t *testing.T) {
	stmts, err := migrate.Statements(os.DirFS("testdata"), "pulse.up.sql")
	require.NoError(t, err)

	require.Len(t, stmts, 3)
	require.Contains(t, stmts[0], "CREATE TABLE pulse_hourly")
}

// TestPrint covers the dry run, which is the whole of what a project's
// migrations are and needs no database to say so.
func TestPrint(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, migrate.Print(&out, os.DirFS("testdata"), ""))

	require.Contains(t, out.String(), "-- Migrations file: pulse.up.sql")
	require.Contains(t, out.String(), "-- Statement index: 0")
	require.Contains(t, out.String(), "CREATE TABLE pulse_hourly")

	// Every statement is printed, not just the first, and each is
	// terminated - the output is meant to be piped into a database.
	require.Contains(t, out.String(), "CREATE TABLE pulse_hosts")
	require.Equal(t, 7, bytes.Count(out.Bytes(), []byte(";\n")))
}
