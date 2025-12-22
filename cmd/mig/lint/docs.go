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

type Options struct {
	db *db.Options

	skipComments bool
	skipPlural   bool
}

func New() *cli.Command {
	var config Options

	return &cli.Command{
		Bind: func(_ context.Context) {
			config.db = db.NewOptions()
			config.db.Bind()

			cli.BoolVar(&config.skipComments, "skip-comments", false, "Skip validating table/column comments")
			cli.BoolVar(&config.skipPlural, "skip-plural", false, "Skip validating table name for singular form")
		},
		Run: func(ctx context.Context, commands []string) error {
			tables, err := internal.ListTables(ctx, config.db)
			if err != nil {
				return err
			}
			errs := validate(tables, config)
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
