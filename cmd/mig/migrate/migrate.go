package migrate

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pkg/errors"

	"github.com/titpetric/cli"

	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/migrate"
)

// Name is the command title.
const Name = "Apply SQL migrations to database"

// New creates a new migrate command. It reports through log.
func New(log *slog.Logger) func() *cli.Command {
	return func() *cli.Command {
		var config struct {
			db      *db.Options
			migrate *migrate.Options
		}

		return &cli.Command{
			Name:  "migrate",
			Title: Name,
			Bind: func(fs *cli.FlagSet) {
				config.db = db.NewOptions(log)
				config.db.Bind(fs)
				config.migrate = migrate.NewOptions(log)
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
					return migrate.Run(ctx, config.db, config.migrate)
				default:
					return migrate.Print(config.migrate)
				}
			},
		}
	}
}
