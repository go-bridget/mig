package lint

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/cmd/mig/internal"
	"github.com/go-bridget/mig/db"
)

const Name = "Check schema for best practices and comments"

func New() *cli.Command {
	var config struct {
		db db.Options

		schema string

		skipComments bool
	}

	return &cli.Command{
		Bind: func(_ context.Context) {
			(&config.db).Init().Bind()
			cli.StringVar(&config.schema, "schema", "", "Database schema to list")
			cli.BoolVar(&config.skipComments, "skip-comments", false, "Skip validating table/column comments")
		},
		Run: func(ctx context.Context, commands []string) error {
			tables, err := internal.ListTables(ctx, config.db, config.schema)
			if err != nil {
				return err
			}
			errs := validate(tables, config.skipComments)
			if len(errs) > 0 {
				for _, err := range errs {
					fmt.Println(err)
				}
				return errors.New("validation failed")
			}
			return nil
		},
	}
}
