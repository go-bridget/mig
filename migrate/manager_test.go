package migrate_test

import (
	"database/sql"
	"os"
	"testing"
	"testing/fstest"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"

	"github.com/go-bridget/mig/migrate"
)

// testDB returns an in-memory sqlite handle closed with the test.
func testDB(t *testing.T) *sqlx.DB {
	t.Helper()

	handle, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { handle.Close() })

	return sqlx.NewDb(handle, "sqlite")
}

// testFS returns a filesystem holding one migration, named as the migrations
// glob wants it, so a test can swap its contents mid-run.
func testFS(t *testing.T, name string) fstest.MapFS {
	t.Helper()

	contents, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)

	return fstest.MapFS{"pulse.up.sql": &fstest.MapFile{Data: contents}}
}

func TestNewManagerErrors(t *testing.T) {
	fsys := fstest.MapFS{}

	t.Run("nil db", func(t *testing.T) {
		_, err := migrate.NewManager(nil, fsys, "test")
		require.ErrorIs(t, err, migrate.ErrNoDB)
	})

	t.Run("nil fs", func(t *testing.T) {
		_, err := migrate.NewManager(testDB(t), nil, "test")
		require.ErrorIs(t, err, migrate.ErrNoFS)
	})

	t.Run("empty project", func(t *testing.T) {
		_, err := migrate.NewManager(testDB(t), fsys, "")
		require.ErrorIs(t, err, migrate.ErrNoProject)
	})

	t.Run("project too long", func(t *testing.T) {
		_, err := migrate.NewManager(testDB(t), fsys, "a-project-name-that-does-not-fit")
		require.ErrorIs(t, err, migrate.ErrNoProject)
	})

	t.Run("unsupported driver", func(t *testing.T) {
		handle, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		defer handle.Close()

		_, err = migrate.NewManager(sqlx.NewDb(handle, "oracle"), fsys, "test")
		require.ErrorIs(t, err, migrate.ErrDriver)
	})
}

func TestManagerProject(t *testing.T) {
	m, err := migrate.NewManager(testDB(t), fstest.MapFS{}, "events")
	require.NoError(t, err)

	require.Equal(t, "events", m.Project())
}
