package migrate

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/go-bridget/mig/logger"

	_ "modernc.org/sqlite"
)

func TestRunWithFSRequiresLogFn(t *testing.T) {
	ctx := context.Background()

	handle, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer handle.Close()

	db := sqlx.NewDb(handle, "sqlite")

	err = RunWithFS(ctx, db, FS{}, &Options{Project: "test", Apply: true})
	require.ErrorIs(t, err, logger.ErrNoLogFn)
}

func TestPrintRequiresLogFn(t *testing.T) {
	err := Print(&Options{Project: "test"})
	require.ErrorIs(t, err, logger.ErrNoLogFn)
}
