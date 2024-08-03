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
			if config.migrate.Project == "" {
				return errors.Errorf("Available migration projects: [%s]", strings.Join(migrate.List(), ", "))
			}
			return nil
		},
		Run: func(ctx context.Context, commands []string) error {
			switch config.migrate.Apply {
			case true:
				return migrate.Run(config.migrate, config.db)
			default:
				return migrate.Print(config.migrate)
			}
			return nil
		},
	}
}
