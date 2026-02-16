package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// AcquireLock acquires a database-specific advisory lock for a given key.
// This is used to prevent concurrent migrations from conflicting.
// The lock is released when the transaction commits or rolls back.
func AcquireLock(ctx context.Context, tx *sqlx.Tx, driverName string, lockKey string) error {
	switch driverName {
	case "postgres", "postgresql", "pgx":
		// PostgreSQL: use advisory lock function
		// We hash the key to a 64-bit integer
		hash := hashString(lockKey)
		if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", hash); err != nil {
			return fmt.Errorf("failed to acquire advisory lock: %w", err)
		}
		return nil

	case "mysql":
		// MySQL: use GET_LOCK function
		// GET_LOCK returns 1 on success, 0 on timeout, NULL on error
		// MySQL enforces a maximum lock name length of 64 characters
		lockKey = fmt.Sprintf("%x", hashString(lockKey))
		var result *int
		if err := tx.QueryRowContext(ctx, "SELECT GET_LOCK(?, 30)", lockKey).Scan(&result); err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		if result == nil || *result != 1 {
			return fmt.Errorf("failed to acquire lock (timeout or error)")
		}
		return nil

	case "sqlite":
		// SQLite: transactions are exclusive by default
		// Issue a PRAGMA to ensure we're not in read-only mode
		if _, err := tx.ExecContext(ctx, "PRAGMA query_only = FALSE"); err != nil {
			return fmt.Errorf("failed to set query_only pragma: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("locking not supported for driver: %s", driverName)
	}
}

// hashString converts a string to a 64-bit hash for use with PostgreSQL advisory locks.
func hashString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h
}
