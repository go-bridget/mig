package create

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/migrate"
)

const Name = "Create database schema SQL"

func New() *cli.Command {
	var config struct {
		db      *db.Options
		migrate *migrate.Options
	}

	return &cli.Command{
		Bind: func(_ context.Context) {
			config.db = db.NewOptions()
			config.db.Bind()
			config.migrate = migrate.NewOptions()
			config.migrate.Bind()
		},
		Init: func(_ context.Context) error {
			if err := migrate.Load(config.migrate); err != nil {
				return errors.Wrap(err, "error loading migrations")
			}
			return nil
		},
		Run: func(ctx context.Context, commands []string) error {
			queries := []string{}
			schemas := migrate.List()
			for _, schema := range schemas {
				queries = append(queries, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;", schema))
			}

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
