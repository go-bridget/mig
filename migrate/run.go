package migrate

import (
	"context"
	"fmt"
	"log"
	"strings"

	"database/sql"

	"github.com/go-bridget/mig/db"
)

// Run takes migrations for a project and executes them against a database
func Run(options *Options, dbOptions *db.Options) error {
	ctx := context.Background()

	db, err := db.ConnectWithRetry(ctx, dbOptions)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	fs, ok := migrations[options.Project]
	if !ok {
		return fmt.Errorf("Migrations for '%s' don't exist", options.Project)
	}

	migrationFile := fmt.Sprintf("migrations-%s.sql", dbOptions.Credentials.Driver)
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
					if err := execQuery(idx, stmt); err != nil {
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

		err := up()
		if filename != migrationsFile {
			// log the migration status into the database
			mapFn := func(fields []string, fn func(string) string) string {
				sql := make([]string, len(fields))
				for k, v := range fields {
					sql[k] = fn(v)
				}
				return strings.Join(sql, ", ")
			}
			saveQuery := "replace into migrations (%s) values (%s)"
			fieldNames := strings.Join(MigrationFields, ",")
			fieldValues := mapFn(MigrationFields, func(in string) string {
				return ":" + in
			})

			if _, err := db.NamedExec(fmt.Sprintf(saveQuery, fieldNames, fieldValues), status); err != nil {
				return fmt.Errorf("updating migration state failed: %w", err)
			}
		}
		log.Println(filename, strings.ToUpper(status.Status))
		return err
	}

	// run main migration
	for idx, stmt := range migrationTable {
		if err := execQuery(idx, stmt); err != nil {
			return err
		}
	}

	// run service migrations
	for _, filename := range fs.Migrations() {
		if err := migrate(filename); err != nil {
			return err
		}
	}
	return nil
}
