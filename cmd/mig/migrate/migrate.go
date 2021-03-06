package migrate

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/migrate"
)

const Name = "Apply SQL migrations to database"

func New() *cli.Command {
	var config struct {
		db      db.Options
		migrate migrate.Options
	}

	return &cli.Command{
		Bind: func(_ context.Context) {
			(&config.db).Init().Bind()
			(&config.migrate).Init().Bind()
		},
		Init: func(_ context.Context) error {
			if err := migrate.Load(config.migrate); err != nil {
				return errors.Wrap(err, "error loading migrations")
			}
			if config.migrate.Project == "" {
				return errors.Errorf("Available migration projects: [%s]", strings.Join(migrate.List(), ", "))
			}
			return nil
		},
		Run: func(ctx context.Context, commands []string) error {
			switch config.migrate.Apply {
			case true:
				handle, err := db.ConnectWithRetry(ctx, config.db)
				if err != nil {
					return errors.Wrap(err, "error connecting to database")
				}
				return migrate.Run(config.migrate, handle)
			default:
				return migrate.Print(config.migrate)
			}
			return nil
		},
	}
}
