package migrate_test

import (
	"context"
	"os"
	"path"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/go-bridget/mig/migrate"
)

// TestApplySucceeds covers the plain case: every statement of every file runs,
// so there is no error to return and every entry says so.
func TestApplySucceeds(t *testing.T) {
	ctx := context.Background()

	fsys := fstest.MapFS{
		"0001-first.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE a (id INTEGER);\nCREATE TABLE b (id INTEGER);\n",
		)},
		"0002-second.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE c (id INTEGER);\n",
		)},
	}

	m, err := migrate.NewManager(testDB(t), fsys, "test")
	require.NoError(t, err)

	applied, err := m.Apply(ctx)
	require.NoError(t, err, "nothing failed, so nothing is returned")

	require.Len(t, applied, 2)
	for _, item := range applied {
		require.Equal(t, migrate.StatusOK, item.Status)
		require.NoError(t, item.Err())
	}
	require.Equal(t, 1, applied[0].StatementIndex, "two statements, zero based")
	require.Equal(t, 0, applied[1].StatementIndex)
}

// TestApplyAppended covers a migration that grew statements since the last run:
// only the new ones are applied, and the recorded index moves with them.
func TestApplyAppended(t *testing.T) {
	ctx := context.Background()

	db := testDB(t)
	fsys := testFS(t, "pulse.up.sql")

	m, err := migrate.NewManager(db, fsys, "test")
	require.NoError(t, err)

	applied, err := m.Apply(ctx)
	require.NoError(t, err)
	require.Len(t, applied, 1)
	require.Equal(t, migrate.StatusOK, applied[0].Status)
	require.Equal(t, 2, applied[0].StatementIndex, "three statements, zero based")

	// The row is the proof the db tags still bind to a select *.
	var status migrate.Migration
	require.NoError(t, db.GetContext(ctx, &status,
		"SELECT * FROM migrations WHERE project='test' AND filename='pulse.up.sql'"))
	require.Equal(t, migrate.StatusOK, status.Status)
	require.Equal(t, 2, status.StatementIndex)

	// Append a fourth statement to the same file.
	appended, err := os.ReadFile("testdata/pulse_appended.up.sql")
	require.NoError(t, err)
	fsys["pulse.up.sql"] = &fstest.MapFile{Data: appended}

	applied, err = m.Apply(ctx)
	require.NoError(t, err)
	require.Equal(t, 3, applied[0].StatementIndex)

	require.NoError(t, db.GetContext(ctx, &status,
		"SELECT * FROM migrations WHERE project='test' AND filename='pulse.up.sql'"))
	require.Equal(t, migrate.StatusOK, status.Status)
	require.Equal(t, 3, status.StatementIndex)
}

// TestApplyTwice covers the second run being a no-op: nothing is re-applied and
// the recorded index stands still.
func TestApplyTwice(t *testing.T) {
	ctx := context.Background()

	m, err := migrate.NewManager(testDB(t), testFS(t, "pulse.up.sql"), "test")
	require.NoError(t, err)

	_, err = m.Apply(ctx)
	require.NoError(t, err)

	// A re-applied CREATE TABLE would error, so a clean second run is the
	// assertion.
	applied, err := m.Apply(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, applied[0].StatementIndex)
}

// TestApplyFailure covers a migration failing halfway: the error comes back,
// the statements that ran are recorded, and the error text lands in the status.
func TestApplyFailure(t *testing.T) {
	ctx := context.Background()

	fsys := fstest.MapFS{
		"bad.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE dupe (id INTEGER);\nCREATE TABLE dupe (id INTEGER);\n",
		)},
	}

	m, err := migrate.NewManager(testDB(t), fsys, "test")
	require.NoError(t, err)

	applied, err := m.Apply(ctx)
	require.Error(t, err)

	// The error names the statement it came from, so the caller can find it
	// without the log line that used to say the same thing.
	require.ErrorContains(t, err, "bad.up.sql: statement 1:")
	require.ErrorContains(t, err, "dupe")

	// The state of the run comes back with the error, not instead of it.
	require.Len(t, applied, 1)
	require.Equal(t, 0, applied[0].StatementIndex, "the first statement survived")
	require.Contains(t, applied[0].Status, "dupe")
	require.NotEqual(t, migrate.StatusOK, applied[0].Status)

	// The entry of the file that failed carries the error itself, not a
	// copy of its message.
	require.ErrorIs(t, applied[0].Err(), err)
}

// TestApplyResumes covers a failed migration being fixed and re-run: the
// statement that already succeeded is not run again.
func TestApplyResumes(t *testing.T) {
	ctx := context.Background()

	fsys := fstest.MapFS{
		"bad.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE dupe (id INTEGER);\nCREATE TABLE dupe (id INTEGER);\n",
		)},
	}

	m, err := migrate.NewManager(testDB(t), fsys, "test")
	require.NoError(t, err)

	_, err = m.Apply(ctx)
	require.Error(t, err)

	// Fix the second statement. The first would error if it ran again.
	fsys["bad.up.sql"] = &fstest.MapFile{Data: []byte(
		"CREATE TABLE dupe (id INTEGER);\nCREATE TABLE fixed (id INTEGER);\n",
	)}

	applied, err := m.Apply(ctx)
	require.NoError(t, err)
	require.Equal(t, migrate.StatusOK, applied[0].Status)
	require.Equal(t, 1, applied[0].StatementIndex)
}

// TestApplyUnreachableDatabase covers the list coming back even when nothing
// could be read to fill it in: the error says what went wrong, and the slice is
// still there to print.
func TestApplyUnreachableDatabase(t *testing.T) {
	ctx := context.Background()

	db := testDB(t)

	m, err := migrate.NewManager(db, testFS(t, "pulse.up.sql"), "test")
	require.NoError(t, err)

	require.NoError(t, db.Close())

	applied, err := m.Apply(ctx)
	require.Error(t, err, "the error return is what says the run failed")

	require.NotNil(t, applied, "the list is data to print, never nil")
	require.Len(t, applied, 1)
	require.Equal(t, "pulse.up.sql", applied[0].Filename)
	require.Equal(t, -1, applied[0].StatementIndex)

	// Nothing here came from a statement, so nothing here belongs on a
	// migration.
	require.NoError(t, applied[0].Err())
}

// TestApplyNoMigrations covers a filesystem with nothing to apply, which looks
// exactly like a clean run unless it is called out.
func TestApplyNoMigrations(t *testing.T) {
	ctx := context.Background()

	t.Run("empty filesystem", func(t *testing.T) {
		m, err := migrate.NewManager(testDB(t), fstest.MapFS{}, "test")
		require.NoError(t, err)

		applied, err := m.Apply(ctx)
		require.ErrorIs(t, err, migrate.ErrNoMigrations)
		require.NotNil(t, applied)
		require.Empty(t, applied)
	})

	t.Run("nothing the pattern matches", func(t *testing.T) {
		fsys := fstest.MapFS{
			"schema/0001-first.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE a (id INTEGER);")},
		}

		m, err := migrate.NewManager(testDB(t), fsys, "test")
		require.NoError(t, err)
		m.Load("migrations/*.sql") // not where these live

		_, err = m.Apply(ctx)
		require.ErrorIs(t, err, migrate.ErrNoMigrations)
		require.ErrorContains(t, err, "migrations/*.sql", "the error says what it looked for")
	})
}

// TestApplyBelowTheRoot covers the mistake the recursive default is for: an
// embed.FS handed over without an fs.Sub, so every migration sits a directory
// down. It applies, and the recorded filename is the path.
func TestApplyBelowTheRoot(t *testing.T) {
	ctx := context.Background()

	fsys := fstest.MapFS{
		"schema/0001-first.up.sql":         &fstest.MapFile{Data: []byte("CREATE TABLE a (id INTEGER);")},
		"schema/nested/0002-second.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE b (id INTEGER);")},
		"README.md":                        &fstest.MapFile{Data: []byte("not a migration")},
	}

	m, err := migrate.NewManager(testDB(t), fsys, "test")
	require.NoError(t, err)

	applied, err := m.Apply(ctx)
	require.NoError(t, err)

	require.Equal(t, []string{
		"schema/0001-first.up.sql",
		"schema/nested/0002-second.up.sql",
	}, []string{applied[0].Filename, applied[1].Filename})
}

// TestLoadPattern covers Load narrowing what a run applies, which is the point
// of it being settable at all.
func TestLoadPattern(t *testing.T) {
	ctx := context.Background()

	fsys := fstest.MapFS{
		"schema/0001-first.sql":  &fstest.MapFile{Data: []byte("CREATE TABLE a (id INTEGER);")},
		"schema/0002-second.sql": &fstest.MapFile{Data: []byte("CREATE TABLE b (id INTEGER);")},
		"schema/deep/0003.sql":   &fstest.MapFile{Data: []byte("CREATE TABLE c (id INTEGER);")},
		"other/0004.sql":         &fstest.MapFile{Data: []byte("CREATE TABLE d (id INTEGER);")},
	}

	m, err := migrate.NewManager(testDB(t), fsys, "test")
	require.NoError(t, err)

	// The default wants *.up.sql, and none of these are.
	_, err = m.Apply(ctx)
	require.ErrorIs(t, err, migrate.ErrNoMigrations)

	// One directory deep, on purpose: a single "*" does not cross a "/".
	m.Load("schema/*.sql")

	applied, err := m.Apply(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{
		"schema/0001-first.sql",
		"schema/0002-second.sql",
	}, []string{applied[0].Filename, applied[1].Filename})
	require.Len(t, applied, 2, "deep/ and other/ are not one directory deep")
}

// TestLoadEmptyRestoresDefault covers Load("") being the way back, so a caller
// need not know the default string to ask for it.
func TestLoadEmptyRestoresDefault(t *testing.T) {
	ctx := context.Background()

	fsys := fstest.MapFS{
		"schema/0001-first.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE a (id INTEGER);")},
	}

	m, err := migrate.NewManager(testDB(t), fsys, "test")
	require.NoError(t, err)

	m.Load("nothing/matches/this/*.sql")
	_, err = m.Apply(ctx)
	require.ErrorIs(t, err, migrate.ErrNoMigrations)

	m.Load("")
	applied, err := m.Apply(ctx)
	require.NoError(t, err)
	require.Len(t, applied, 1)
}

// TestLoadBadPattern covers Load taking no return: a pattern that doesn't
// compile is the run's problem, not the setter's.
func TestLoadBadPattern(t *testing.T) {
	ctx := context.Background()

	// An empty filesystem too: a malformed pattern is not something to
	// discover only when there happens to be a name to try it against.
	m, err := migrate.NewManager(testDB(t), fstest.MapFS{}, "test")
	require.NoError(t, err)

	m.Load("schema/[unterminated*.sql")

	applied, err := m.Apply(ctx)
	require.ErrorIs(t, err, path.ErrBadPattern)
	require.NotNil(t, applied)

	_, err = m.List(ctx)
	require.ErrorIs(t, err, path.ErrBadPattern)
}

// TestApplyNoStatements covers a migration file that parses to nothing, which
// is a file someone meant to finish writing.
func TestApplyNoStatements(t *testing.T) {
	ctx := context.Background()

	fsys := fstest.MapFS{
		"0001-comments-only.up.sql": &fstest.MapFile{Data: []byte("-- nothing here yet\n-- really\n")},
	}

	m, err := migrate.NewManager(testDB(t), fsys, "test")
	require.NoError(t, err)

	applied, err := m.Apply(ctx)
	require.ErrorIs(t, err, migrate.ErrNoStatements)
	require.ErrorContains(t, err, "0001-comments-only.up.sql")

	// Nothing was recorded for it: an unfinished migration is not applied.
	require.Len(t, applied, 1)
	require.Equal(t, -1, applied[0].StatementIndex)
	require.Empty(t, applied[0].Status)
}
