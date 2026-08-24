package migrate

import (
	"fmt"

	"github.com/pkg/errors"
)

// Print outputs database migrations for a project to the bound log sink.
func Print(options *Options) error {
	fs, ok := migrations[options.Project]
	if !ok {
		return errors.Errorf("Migrations for '%s' don't exist", options.Project)
	}

	printQuery := func(idx int, query string) error {
		if options.Verbose {
			options.printf("")
			options.printf("-- Statement index: %d", idx)
			options.printf("%s", query)
			options.printf("")
		}
		return nil
	}

	migrate := func(filename string) error {
		options.printf("-- Migrations file: %s", filename)
		stmts, err := statements(fs.ReadFile(filename))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error reading migration: %s", filename))
		}
		for idx, stmt := range stmts {
			if err := printQuery(idx, stmt); err != nil {
				return err
			}
		}
		return nil
	}

	// print service migrations
	for _, filename := range fs.Migrations() {
		if err := migrate(filename); err != nil {
			return err
		}
	}
	return nil
}
