package migrate

import (
	"context"
	"fmt"
	"log"
	"strings"

	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/go-bridget/mig/db"
)

// Run takes migrations for a project and executes them against a database.
func Run(ctx context.Context, dbOptions *db.Options, options *Options) error {
	database, err := db.ConnectWithRetry(ctx, dbOptions)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	return RunWithDB(ctx, database, options)
}

// RunWithDB runs the registered migrations from options against a *sqlx.DB with context.
func RunWithDB(ctx context.Context, sqldb *sqlx.DB, options *Options) error {
	fs, ok := migrations[options.Project]
	if !ok {
		return fmt.Errorf("Migrations for '%s' don't exist", options.Project)
	}

	return RunWithFS(ctx, sqldb, fs, options)
}

// RunWithFS runs the passed migrations against a *sqlx.DB with context.
func RunWithFS(ctx context.Context, sqldb *sqlx.DB, fs FS, options *Options) error {
	// Normalize driver name for migration file lookup
	driverName := sqldb.DriverName()
	if driverName == "pgx" {
		driverName = "postgres"
	}
	migrationFile := fmt.Sprintf("migrations-%s.sql", driverName)
	migrationTable, err := statements(migrationsFS.ReadFile(migrationFile))
	if err != nil {
		return fmt.Errorf("error reading %s: %w", migrationFile, err)
	}

	printQuery := func(idx int, query string) {
		if options.Verbose {
			fmt.Println()
			fmt.Println("-- Statement index:", idx)
			fmt.Println(query)
			fmt.Println()
		}
	}

	execQuery := func(idx int, query string) error {
		printQuery(idx, query)
		if _, err := sqldb.ExecContext(ctx, query); err != nil && err != sql.ErrNoRows {
			return err
		}
		return nil
	}

	// execQueryInTransaction executes a query within the context of a transaction
	execQueryInTransaction := func(tx *sqlx.Tx, idx int, query string) error {
		printQuery(idx, query)
		if _, err := tx.ExecContext(ctx, query); err != nil && err != sql.ErrNoRows {
			return err
		}
		return nil
	}

	migrate := func(filename string) error {
		status := Migration{
			Project:        options.Project,
			Filename:       filename,
			StatementIndex: -1,
		}

		// Use a transaction with advisory lock to handle concurrent migrations safely
		tx, err := sqldb.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		// Acquire lock to prevent concurrent migrations from interfering
		lockKey := fmt.Sprintf("%s:%s", status.Project, status.Filename)
		if err := db.AcquireLock(ctx, tx, sqldb.DriverName(), lockKey); err != nil {
			return fmt.Errorf("failed to acquire migration lock: %w", err)
		}

		// Re-check if migration record exists under lock
		query := sqldb.Rebind("select * from migrations where project=? and filename=?")
		exists := true
		if err := tx.GetContext(ctx, &status, query, status.Project, status.Filename); err != nil {
			if err == sql.ErrNoRows {
				exists = false
			} else {
				return err
			}
		}

		// If migration already exists and is marked ok, skip it
		if exists && status.Status == "ok" {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}
			log.Println(filename, "SKIPPED (already applied)")
			return nil
		}

		up := func() error {
			stmts, err := statements(fs.ReadFile(filename))
			if err != nil {
				return fmt.Errorf("Error reading %s: %w", filename, err)
			}

			var isApplied bool
			for idx, stmt := range stmts {
				isApplied = idx <= status.StatementIndex
				if options.Verbose {
					fmt.Printf("-- statement %d/%d is applied? %t\n", idx, status.StatementIndex, isApplied)
				}
				// skip stmt if it has already been applied
				if !isApplied {
					status.StatementIndex = idx
					if err := execQueryInTransaction(tx, idx, stmt); err != nil {
						status.StatementIndex--
						status.Status = err.Error()
						return err
					}
				} else {
					printQuery(idx, stmt)
				}
			}
			status.Status = "ok"
			return nil
		}

		err = up()

		// Save migration status to database within transaction
		if exists {
			// UPDATE existing record
			updateQuery := "UPDATE migrations SET statement_index=:statement_index, status=:status WHERE project=:project AND filename=:filename"
			if _, err := tx.NamedExecContext(ctx, updateQuery, status); err != nil {
				return fmt.Errorf("updating migration state failed: %w", err)
			}
		} else {
			// INSERT new record
			insertQuery := "INSERT INTO migrations (project, filename, statement_index, status) VALUES (:project, :filename, :statement_index, :status)"
			if _, err := tx.NamedExecContext(ctx, insertQuery, status); err != nil {
				return fmt.Errorf("updating migration state failed: %w", err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		log.Println(filename, strings.ToUpper(status.Status))
		return err
	}

	// Run main migration (schema creation for migrations table itself)
	for idx, stmt := range migrationTable {
		if err := execQuery(idx, stmt); err != nil {
			return err
		}
	}

	// Run service migrations
	for _, filename := range fs.Migrations() {
		if err := migrate(filename); err != nil {
			return err
		}
	}
	return nil
}
