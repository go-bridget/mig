package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/migrate"
)

func createCmd() *cli.Command {
	var config struct {
		db      db.Options
		migrate migrate.Options

		real    bool
		service string
	}

	return &cli.Command{
		Bind: func(_ context.Context) {
			cli.StringVar(&config.migrate.Path, "migrate-path", "schema", "Source path for database migrations")
			cli.StringVar(&config.db.Credentials.Driver, "db-driver", "mysql", "Database driver")
			cli.StringVar(&config.db.Credentials.DSN, "db-dsn", "", "DSN for database connection")
			cli.StringVar(&config.service, "service", "", "Service name for migrations")
			cli.BoolVar(&config.real, "real", false, "false = print migrations, true = run migrations")
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

			if config.real {
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
