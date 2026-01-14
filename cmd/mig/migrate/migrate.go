package migrate

import (
	"context"
	"errors"
	"fmt"

	flag "github.com/spf13/pflag"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/migrate"
)

const Name = "Apply SQL migrations to database"

func New() *cli.Command {
	var config struct {
		db      *db.Options
		migrate *migrate.Options
	}

	return &cli.Command{
		Name:  "migrate",
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
				return errors.New("Specify project name as first argument to migrate")
			}

			if err := migrate.Load(config.migrate); err != nil {
				return fmt.Errorf("error loading migrations: %w", err)
			}

			switch {
			case config.migrate.Apply:
				return migrate.Run(ctx, config.migrate, config.db)
			default:
				return migrate.Print(config.migrate)
			}
		},
	}
}
