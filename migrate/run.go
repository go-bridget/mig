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
func Run(options Options, db *sqlx.DB) error {
	fs, ok := migrations[options.Project]
	if !ok {
		return errors.Errorf("Migrations for '%s' don't exist", options.Project)
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
		if _, err := db.Exec(query); err != nil && err != sql.ErrNoRows {
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

		// we can't log the main migrations table
		if filename != migrationsFile {
			if err := db.Get(&status, "select * from migrations where project=? and filename=?", status.Project, status.Filename); err != nil && err != sql.ErrNoRows {
				return err
			}
		}

		up := func() error {
			stmts, err := statements(fs.ReadFile(filename))
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Error reading migration: %s", filename))
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
					if err := execQuery(idx, stmt); err != nil {
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

		err := up()
		if filename != migrationsFile {
			// log the migration status into the database
			set := func(fields []string) string {
				sql := make([]string, len(fields))
				for k, v := range fields {
					sql[k] = v + "=:" + v
				}
				return strings.Join(sql, ", ")
			}
			if _, err := db.NamedExec("replace into migrations set "+set(MigrationFields), status); err != nil {
				return errors.Wrap(err, "updating migration state failed")
			}
		}
		log.Println(filename, strings.ToUpper(status.Status))
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
