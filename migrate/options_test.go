package migrate

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

// TestOptionsWithoutLogger covers an Options carrying no Logger, which is
// what a program embedding mig passes when it doesn't want a running
// commentary. It used to be refused before the database was touched.
func TestOptionsWithoutLogger(t *testing.T) {
	ctx := context.Background()

	handle, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer handle.Close()

	db := sqlx.NewDb(handle, "sqlite")

	migration, err := testdataFS.ReadFile("testdata/pulse.up.sql")
	require.NoError(t, err)

	migrations["quiet"] = FS{"pulse.up.sql": migration}
	t.Cleanup(func() { delete(migrations, "quiet") })

	require.NoError(t, RunWithFS(ctx, db, FS{"pulse.up.sql": migration}, &Options{Project: "quiet", Apply: true}))
	require.NoError(t, Print(&Options{Project: "quiet", Output: &bytes.Buffer{}}))
}

// TestOptionsOutput covers where the two streams go: the SQL to the output,
// and what the run did to the logger.
func TestOptionsOutput(t *testing.T) {
	ctx := context.Background()

	handle, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer handle.Close()

	db := sqlx.NewDb(handle, "sqlite")

	migration, err := testdataFS.ReadFile("testdata/pulse.up.sql")
	require.NoError(t, err)

	var out, logged bytes.Buffer

	err = RunWithFS(ctx, db, FS{"pulse.up.sql": migration}, &Options{
		Project: "test",
		Apply:   true,
		Verbose: true,
		Output:  &out,
		Logger:  slog.New(slog.NewTextHandler(&logged, nil)),
	})
	require.NoError(t, err)

	require.Contains(t, out.String(), "CREATE TABLE", "the sql belongs on the output")
	require.NotContains(t, out.String(), "level=", "the output carries no log lines")

	require.Contains(t, logged.String(), "msg=migration", "the run belongs on the logger")
	require.NotContains(t, logged.String(), "CREATE TABLE", "the logger carries no sql")
}

// TestPrintWritesToOutput covers the print path, which is the whole of what
// a project's migrations are.
func TestPrintWritesToOutput(t *testing.T) {
	migration, err := testdataFS.ReadFile("testdata/pulse.up.sql")
	require.NoError(t, err)

	migrations["print"] = FS{"pulse.up.sql": migration}
	t.Cleanup(func() { delete(migrations, "print") })

	var out bytes.Buffer
	require.NoError(t, Print(&Options{Project: "print", Verbose: true, Output: &out}))

	require.Contains(t, out.String(), "-- Migrations file: pulse.up.sql")
	require.Contains(t, out.String(), "CREATE TABLE")
}
