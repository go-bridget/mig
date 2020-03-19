package migrate

import (
	"fmt"
	"log"
	"strings"

	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Run takes migrations for a project and executes them against a database
func Run(project string, db *sqlx.DB) error {
	fs, ok := migrations[project]
	if !ok {
		return errors.Errorf("Migrations for '%s' don't exist", project)
	}

	execQuery := func(idx int, query string, useLog bool) error {
		if useLog {
			log.Println()
			log.Println("-- Statement index:", idx)
			log.Println(query)
			log.Println()
		}
		if _, err := db.Exec(query); err != nil && err != sql.ErrNoRows {
			return err
		}
		return nil
	}

	migrate := func(filename string) error {
		log.Println("Running migrations from", filename)

		status := Migration{
			Project:  project,
			Filename: filename,
		}

		// we can't log the main migrations table
		useLog := (filename != migrationsFile)
		if useLog {
			if err := db.Get(&status, "select * from migrations where project=? and filename=?", status.Project, status.Filename); err != nil && err != sql.ErrNoRows {
				return err
			}
			if status.Status == "ok" {
				log.Println("Migrations already applied, skipping")
				return nil
			}
		}

		up := func() error {
			stmts, err := statements(fs.ReadFile(filename))
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Error reading migration: %s", filename))
			}

			for idx, stmt := range stmts {
				// skip stmt if it has already been applied
				if idx >= status.StatementIndex {
					status.StatementIndex = idx
					if err := execQuery(idx, stmt, useLog); err != nil {
						status.Status = err.Error()
						return err
					}
				}
			}
			status.Status = "ok"
			return nil
		}

		err := up()
		if useLog {
			// log the migration status into the database
			set := func(fields []string) string {
				sql := make([]string, len(fields))
				for k, v := range fields {
					sql[k] = v + "=:" + v
				}
				return strings.Join(sql, ", ")
			}
			if _, err := db.NamedExec("replace into migrations set "+set(MigrationFields), status); err != nil {
				log.Println("Updating migration status failed:", err)
			}
		}
		return err
	}

	// run main migration
	if err := migrate(migrationsFile); err != nil {
		return err
	}

	// run service migrations
	for _, filename := range fs.Migrations() {
		if err := migrate(filename); err != nil {
			return err
		}
	}
	return nil
}
