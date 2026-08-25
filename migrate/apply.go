package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/jmoiron/sqlx"

	"github.com/go-bridget/mig/db"
)

const (
	updateQuery = "UPDATE " + Table + " SET statement_index=:statement_index, status=:status" +
		" WHERE project=:project AND filename=:filename"

	insertQuery = "INSERT INTO " + Table + " (project, filename, statement_index, status)" +
		" VALUES (:project, :filename, :statement_index, :status)"
)

// Apply runs the migrations that have not been applied yet, in lexical order of
// filename, and returns on the first one that fails.
//
// Whatever stopped the run comes back as the error, whether it was a statement
// or the database being out of reach, and that error is the one to check.
//
// The slice is the run alongside it, to print: one Migration per file, in order,
// carrying the row recorded for it. It is never nil, so a failure that never
// reached the migrations table still lists the files it would have applied. The
// entry of a file whose statement failed carries that error, so Err on it
// returns what Apply returned rather than a copy of its message; an error with
// no migration behind it, a connection dropping say, is on the error return
// alone.
//
// Apply creates the migrations table if it is missing. Each file is applied in
// one transaction: statements the table already records as applied are skipped,
// and the index of the last statement that ran is recorded whether the file
// succeeded or not. A file that fails leaves the error in its Status, and a
// later Apply resumes it from the statement after the last that succeeded.
//
// A filesystem holding no migrations is ErrNoMigrations, and a migration that
// parses to no statements is ErrNoStatements. Neither is a run worth reporting
// as a success.
func (m *Manager) Apply(ctx context.Context) ([]Migration, error) {
	files, err := m.files()
	if err != nil {
		return []Migration{}, err
	}

	// Applying nothing is not success. It is the one failure that otherwise
	// looks exactly like a clean run.
	if len(files) == 0 {
		return []Migration{}, fmt.Errorf("%w matching %s", ErrNoMigrations, m.pattern)
	}

	if err := m.ensureTable(ctx); err != nil {
		return m.pending(files), err
	}

	for _, name := range files {
		status, err := m.applyFile(ctx, name)
		if err == nil {
			continue
		}

		// The state of the run is worth more than the error alone, so
		// report as far as it got either way.
		applied, reportErr := m.report(ctx)
		if reportErr != nil {
			applied = m.pending(files)
		}

		// report reads the rows back, which hold the text of the error
		// but not the error. Put the one the run still has in its place.
		if status.Err() != nil {
			for i := range applied {
				if applied[i].Filename == name {
					applied[i].SetError(status.Err())
				}
			}
		}

		return applied, err
	}

	applied, err := m.report(ctx)
	if err != nil {
		return m.pending(files), err
	}

	return applied, nil
}

// applyFile applies one migration, in a transaction holding the lock for it. It
// returns the record it wrote, which holds the error of a statement that failed.
func (m *Manager) applyFile(ctx context.Context, filename string) (Migration, error) {
	status := Migration{
		Project:        m.project,
		Filename:       filename,
		StatementIndex: -1,
	}

	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return status, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Take the lock before reading the row, so a concurrent migration can't
	// decide to apply the same statements.
	lockKey := fmt.Sprintf("%s:%s", status.Project, status.Filename)
	if err := db.AcquireLock(ctx, tx, m.driver, lockKey); err != nil {
		return status, fmt.Errorf("failed to acquire migration lock: %w", err)
	}

	query := m.db.Rebind("select * from " + Table + " where project=? and filename=?")
	exists := true
	if err := tx.GetContext(ctx, &status, query, status.Project, status.Filename); err != nil {
		if err != sql.ErrNoRows {
			return status, fmt.Errorf("error reading %s table: %w", Table, err)
		}
		exists = false
	}

	stmts, err := Statements(m.fsys, filename)
	if err != nil {
		return status, err
	}

	// A file of nothing but comments parses to no statements. Files too
	// short to hold one never get this far - Files leaves them out - so
	// this is a migration someone wrote and left unfinished.
	if len(stmts) == 0 {
		return status, fmt.Errorf("%s: %w", filename, ErrNoStatements)
	}

	// A file recorded as applied is only work again if it grew statements
	// since the last run.
	if exists && status.Status == StatusOK && len(stmts) <= status.StatementIndex+1 {
		if err := tx.Commit(); err != nil {
			return status, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return status, nil
	}

	applyErr := m.execStatements(ctx, tx, filename, stmts, &status)

	// The status row records what ran, so it is written whether the file
	// succeeded or not - that is what makes a failed migration resumable.
	saveQuery := insertQuery
	if exists {
		saveQuery = updateQuery
	}
	if _, err := tx.NamedExecContext(ctx, saveQuery, status); err != nil {
		return status, fmt.Errorf("updating migration state failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return status, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return status, applyErr
}

// execStatements runs the statements the status row has no record of, leaving
// the index of the last one that ran, and its error, on status.
//
// The error names the statement it came from, because that is the whole of what
// a caller needs to find it, and it is what lands in the status column.
func (m *Manager) execStatements(ctx context.Context, tx *sqlx.Tx, filename string, stmts []string, status *Migration) error {
	for idx, stmt := range stmts {
		if idx <= status.StatementIndex {
			continue
		}

		status.StatementIndex = idx
		if _, err := tx.ExecContext(ctx, stmt); err != nil && err != sql.ErrNoRows {
			status.StatementIndex--
			status.SetError(fmt.Errorf("%s: statement %d: %w", filename, idx, err))
			return status.Err()
		}
	}

	status.SetError(nil)

	return nil
}

// sortByFilename orders migrations the way Files does, so the rows with no
// file behind them come out in a stable order too.
func sortByFilename(items []Migration) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Filename < items[j].Filename
	})
}
