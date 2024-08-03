package migrate

import (
	"context"
	"errors"
	"fmt"

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
		Bind: func(_ context.Context) {
			config.db = db.NewOptions()
			config.db.Bind()
			config.migrate = migrate.NewOptions()
			config.migrate.Bind()
		},
		Run: func(ctx context.Context, commands []string) error {
			if len(commands) > 1 {
				config.migrate.Project = commands[1]
			}

			if config.migrate.Project == "" {
				return errors.New("Specify project name as first argument to migrate")
			}

			if err := migrate.Load(config.migrate); err != nil {
				return fmt.Errorf("error loading migrations: %w", err)
			}

			switch {
			case config.migrate.Apply:
				return migrate.Run(config.migrate, config.db)
			default:
				return migrate.Print(config.migrate)
			}
			return nil
		},
	}
}
