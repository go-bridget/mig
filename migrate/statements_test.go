package migrate

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/*.sql
var testdataFS embed.FS

func TestStatementsMultiple(t *testing.T) {
	contents, err := testdataFS.ReadFile("testdata/pulse.up.sql")
	require.NoError(t, err)

	stmts, err := statements(contents, nil)
	require.NoError(t, err)
	require.Len(t, stmts, 3)
}
