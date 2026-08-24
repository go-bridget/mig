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

// recorder is a Logger owing nothing to slog, which is the point of Logger
// being an interface: a caller brings the logger they already have.
type recorder struct {
	lines []string
}

func (r *recorder) Info(msg string, _ ...any)  { r.lines = append(r.lines, "info:"+msg) }
func (r *recorder) Error(msg string, _ ...any) { r.lines = append(r.lines, "error:"+msg) }

// TestNewOptionsBindsLogger covers the logger arriving through the
// constructor, which is the whole point of it being a parameter: a caller
// can't reach an Options without having said where the progress goes.
func TestNewOptionsBindsLogger(t *testing.T) {
	log := &recorder{}
	options := NewOptions(log)

	require.Same(t, log, options.Logger)
	require.Equal(t, "schema", options.Path, "the constructor still fills the defaults")
}

// TestRunWithLoggerOfOwnMaking covers a run reporting through a Logger that
// isn't a *slog.Logger, which nothing in the package requires it to be.
func TestRunWithLoggerOfOwnMaking(t *testing.T) {
	ctx := context.Background()

	handle, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer handle.Close()

	db := sqlx.NewDb(handle, "sqlite")

	migration, err := testdataFS.ReadFile("testdata/pulse.up.sql")
	require.NoError(t, err)

	log := &recorder{}
	options := NewOptions(log)
	options.Project = "own"
	options.Apply = true
	options.Output = &bytes.Buffer{}

	require.NoError(t, RunWithFS(ctx, db, FS{"pulse.up.sql": migration}, options))
	require.Equal(t, []string{"info:migration"}, log.lines)
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
