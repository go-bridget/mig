package create

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/migrate"
)

// Name is the command title.
const Name = "Create database schema SQL"

// New creates a new create command.
func New() *cli.Command {
	var config struct {
		db      *db.Options
		migrate *migrate.Options
	}

	return &cli.Command{
		Name:  "create",
		Title: Name,
		Bind: func(fs *flag.FlagSet) {
			config.db = db.NewOptions()
			config.db.Bind(fs)
			config.migrate = migrate.NewOptions()
			config.migrate.Bind(fs)
		},
		Run: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				config.migrate.Project = args[0]
			}

			if config.migrate.Project == "" {
				return errors.Errorf("Specify project name as first argument to create")
			}

			queries := []string{}
			queries = append(queries, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;", config.migrate.Project))

			if config.migrate.Apply {
				handle, err := db.ConnectWithRetry(ctx, config.db)
				if err != nil {
					return errors.Wrap(err, "error connecting to database")
				}

				for _, query := range queries {
					fmt.Println(query)
					if _, err := handle.Exec(query); err != nil {
						return err
					}
				}
				return nil
			}
			for _, query := range queries {
				fmt.Println(query)
			}
			return nil
		},
	}
}
