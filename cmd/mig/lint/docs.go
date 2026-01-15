package lint

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"

	"github.com/titpetric/cli"

	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/db/introspect"
)

// Name is the command title.
const Name = "Lint database schema"

// Options contains lint configuration.
type Options struct {
	db *db.Options

	skipComments bool
	skipPlural   bool
}

// New creates a new lint command.
func New() *cli.Command {
	var config Options

	return &cli.Command{
		Name:  "lint",
		Title: Name,
		Bind: func(fs *flag.FlagSet) {
			config.db = db.NewOptions()
			config.db.Bind(fs)

			fs.BoolVar(&config.skipComments, "skip-comments", false, "Skip validating table/column comments")
			fs.BoolVar(&config.skipPlural, "skip-plural", false, "Skip validating table name for singular form")
		},
		Run: func(ctx context.Context, args []string) error {
			handle, err := db.ConnectWithRetry(ctx, config.db)
			if err != nil {
				return err
			}

			desc, err := introspect.NewDescriber(handle)
			if err != nil {
				return err
			}

			tables, err := introspect.ListTablesWithColumns(ctx, handle, desc)
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
