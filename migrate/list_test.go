package migrate_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/go-bridget/mig/migrate"
)

// TestListPending covers a migration nobody has run: it is listed, with the
// sentinel index and no status, because the table has no row for it.
func TestListPending(t *testing.T) {
	ctx := context.Background()

	m, err := migrate.NewManager(testDB(t), testFS(t, "pulse.up.sql"), "test")
	require.NoError(t, err)

	list, err := m.List(ctx)
	require.NoError(t, err)

	require.Len(t, list, 1)
	require.Equal(t, "pulse.up.sql", list[0].Filename)
	require.Equal(t, "test", list[0].Project)
	require.Equal(t, -1, list[0].StatementIndex)
	require.Empty(t, list[0].Status)
}

// TestListApplied covers List carrying the row of the migrations table for a
// migration that has run.
func TestListApplied(t *testing.T) {
	ctx := context.Background()

	m, err := migrate.NewManager(testDB(t), testFS(t, "pulse.up.sql"), "test")
	require.NoError(t, err)

	_, err = m.Apply(ctx)
	require.NoError(t, err)

	list, err := m.List(ctx)
	require.NoError(t, err)

	require.Len(t, list, 1)
	require.Equal(t, migrate.StatusOK, list[0].Status)
	require.Equal(t, 2, list[0].StatementIndex)
}

// TestListFailed covers the error of a failed migration being readable from the
// list, which is where the status column puts it.
func TestListFailed(t *testing.T) {
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

	list, err := m.List(ctx)
	require.NoError(t, err)

	require.Len(t, list, 1)
	require.Contains(t, list[0].Status, "dupe")
	require.Equal(t, 0, list[0].StatementIndex)
}

// TestListMissingFile covers a migration applied and since deleted: the row is
// still listed, because forgetting it would hide what the database holds.
func TestListMissingFile(t *testing.T) {
	ctx := context.Background()

	db := testDB(t)

	applied, err := migrate.NewManager(db, testFS(t, "pulse.up.sql"), "test")
	require.NoError(t, err)
	_, err = applied.Apply(ctx)
	require.NoError(t, err)

	// Same database, no files left.
	m, err := migrate.NewManager(db, fstest.MapFS{}, "test")
	require.NoError(t, err)

	list, err := m.List(ctx)
	require.NoError(t, err)

	require.Len(t, list, 1)
	require.Equal(t, "pulse.up.sql", list[0].Filename)
	require.Equal(t, migrate.StatusOK, list[0].Status)
}

// TestListScopedToProject covers several projects sharing one migrations table,
// which is what the project column is for.
func TestListScopedToProject(t *testing.T) {
	ctx := context.Background()

	db := testDB(t)

	one, err := migrate.NewManager(db, testFS(t, "pulse.up.sql"), "one")
	require.NoError(t, err)
	_, err = one.Apply(ctx)
	require.NoError(t, err)

	two, err := migrate.NewManager(db, fstest.MapFS{}, "two")
	require.NoError(t, err)

	list, err := two.List(ctx)
	require.NoError(t, err)
	require.Empty(t, list, "the migrations of 'one' are not the migrations of 'two'")
}

// TestListOrder covers the lexical ordering the run itself follows.
func TestListOrder(t *testing.T) {
	ctx := context.Background()

	fsys := fstest.MapFS{
		"0002-second.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE b (id INTEGER);")},
		"0001-first.up.sql":  &fstest.MapFile{Data: []byte("CREATE TABLE a (id INTEGER);")},
	}

	m, err := migrate.NewManager(testDB(t), fsys, "test")
	require.NoError(t, err)

	list, err := m.List(ctx)
	require.NoError(t, err)

	require.Equal(t, "0001-first.up.sql", list[0].Filename)
	require.Equal(t, "0002-second.up.sql", list[1].Filename)
}
