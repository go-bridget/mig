package main

import (
	"context"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/migrate"
)

func main() {
	var config struct {
		db      db.Options
		migrate migrate.Options

		real    bool
		service string
	}

	cli.StringVar(&config.migrate.Path, "migrate-path", "schema", "Source path for database migrations")
	cli.StringVar(&config.db.Credentials.Driver, "db-driver", "mysql", "Database driver")
	cli.StringVar(&config.db.Credentials.DSN, "db-dsn", "", "DSN for database connection")
	cli.StringVar(&config.service, "service", "", "Service name for migrations")
	cli.BoolVar(&config.real, "real", false, "false = print migrations, true = run migrations")

	app := new(cli.App)
	app.Init = func(_ context.Context) error {
		if err := migrate.Load(config.migrate); err != nil {
			return errors.Wrap(err, "error loading migrations")
		}
		if config.service == "" {
			return errors.Errorf("Available migration services: [%s]", strings.Join(migrate.List(), ", "))
		}
		return nil
	}
	app.Action = func(ctx context.Context, commands []string) error {
		switch config.real {
		case true:
			handle, err := db.ConnectWithRetry(ctx, config.db)
			if err != nil {
				return errors.Wrap(err, "error connecting to database")
			}
			return migrate.Run(config.service, handle)
		default:
			return migrate.Print(config.service)
		}
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("An error occured: %s", err)
	}
}
