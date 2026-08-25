package migrate

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/*.sql
var testdataFS embed.FS

func TestSplitMultiple(t *testing.T) {
	contents, err := testdataFS.ReadFile("testdata/pulse.up.sql")
	require.NoError(t, err)

	require.Len(t, split(contents), 3)
}

func TestSplitComments(t *testing.T) {
	stmts := split([]byte("-- a comment\nSELECT 1;\n-- another\nSELECT 2;\n"))

	require.Equal(t, []string{"SELECT 1", "SELECT 2"}, stmts)
}

func TestBuiltins(t *testing.T) {
	got := builtins("INSERT INTO t (id) VALUES (uuid())")

	require.NotContains(t, got, "uuid()")
	require.Regexp(t, `VALUES \('[0-9a-f-]{36}'\)`, got)
}
