package migrate

import (
	"context"
	"database/sql"
	"embed"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
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

func TestStatementsAppended(t *testing.T) {
	ctx := context.Background()

	handle, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer handle.Close()

	db := sqlx.NewDb(handle, "sqlite")

	// First run: 3 statements
	initial, err := testdataFS.ReadFile("testdata/pulse.up.sql")
	require.NoError(t, err)

	fs := FS{"pulse.up.sql": initial}
	err = RunWithFS(ctx, db, fs, &Options{
		Project: "test",
		Apply:   true,
	})
	require.NoError(t, err)

	// Verify migration recorded with statement_index=2 (0-based, 3 statements)
	var status Migration
	err = db.GetContext(ctx, &status, "SELECT * FROM migrations WHERE project='test' AND filename='pulse.up.sql'")
	require.NoError(t, err)
	require.Equal(t, "ok", status.Status)
	require.Equal(t, 2, status.StatementIndex)

	// Second run: append a 4th statement
	appended, err := testdataFS.ReadFile("testdata/pulse_appended.up.sql")
	require.NoError(t, err)

	fs["pulse.up.sql"] = appended
	err = RunWithFS(ctx, db, fs, &Options{
		Project: "test",
		Apply:   true,
	})
	require.NoError(t, err)

	// Verify migration updated with statement_index=3
	err = db.GetContext(ctx, &status, "SELECT * FROM migrations WHERE project='test' AND filename='pulse.up.sql'")
	require.NoError(t, err)
	require.Equal(t, "ok", status.Status)
	require.Equal(t, 3, status.StatementIndex)
}
