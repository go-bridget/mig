package main

import (
	"flag"
	"strings"
	"log"

	"github.com/SentimensRG/sigctx"
	_ "github.com/go-sql-driver/mysql"

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

	flag.StringVar(&config.migrate.Path, "migrate-path", "schema", "Source path for database migrations")
	flag.StringVar(&config.db.Credentials.Driver, "db-driver", "mysql", "Database driver")
	flag.StringVar(&config.db.Credentials.DSN, "db-dsn", "", "DSN for database connection")
	flag.StringVar(&config.service, "service", "", "Service name for migrations")
	flag.BoolVar(&config.real, "real", false, "false = print migrations, true = run migrations")

	flag.Parse()

	if err := migrate.Load(config.migrate); err != nil {
		log.Fatalf("An error occured: %+v", err)
	}
	if config.service == "" {
		log.Fatalf("Available migration services: [%s]", strings.Join(migrate.List(), ", "))
	}

	ctx := sigctx.New()

	switch config.real {
	case true:
		handle, err := db.ConnectWithRetry(ctx, config.db)
		if err != nil {
			log.Fatalf("Error connecting to database: %+v", err)
		}
		if err := migrate.Run(config.service, handle); err != nil {
			log.Fatalf("An error occurred: %+v", err)
		}
	default:
		if err := migrate.Print(config.service); err != nil {
			log.Fatalf("An error occurred: %+v", err)
		}
	}
}
