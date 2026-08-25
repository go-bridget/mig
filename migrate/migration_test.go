package migrate_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-bridget/mig/migrate"
)

// TestSetError covers the two things SetError does: it puts the message where
// the column wants it, and keeps the error where Err can return it.
func TestSetError(t *testing.T) {
	sentinel := errors.New("statement failed")

	var m migrate.Migration
	m.SetError(fmt.Errorf("bad.up.sql: statement 1: %w", sentinel))

	require.Equal(t, "bad.up.sql: statement 1: statement failed", m.Status)
	require.ErrorIs(t, m.Err(), sentinel, "the error survives, not just its message")
}

// TestSetErrorNil covers a migration that ran to its last statement, which is
// the only thing that writes StatusOK.
func TestSetErrorNil(t *testing.T) {
	var m migrate.Migration
	m.SetError(nil)

	require.Equal(t, migrate.StatusOK, m.Status)
	require.NoError(t, m.Err())
}

// TestErrFromStatus covers a migration read back from the migrations table,
// where there is no error left to return, only the text of one.
func TestErrFromStatus(t *testing.T) {
	t.Run("never run", func(t *testing.T) {
		require.NoError(t, migrate.Migration{}.Err())
	})

	t.Run("applied", func(t *testing.T) {
		require.NoError(t, migrate.Migration{Status: migrate.StatusOK}.Err())
	})

	t.Run("failed", func(t *testing.T) {
		err := migrate.Migration{Status: "table dupe already exists"}.Err()

		require.EqualError(t, err, "table dupe already exists")
	})
}
