package create

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"

	"github.com/titpetric/cli"

	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/migrate"
)

// Name is the command title.
const Name = "Create database schema SQL"

func createDatabaseQuery(driver, name string) string {
	switch driver {
	case "pgx":
		return fmt.Sprintf(`CREATE DATABASE "%s"`, name)
	default:
		return fmt.Sprintf("CREATE DATABASE `%s`", name)
	}
}

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

			driver, _ := db.ParseDSN(config.db.Credentials.DSN)
			query := createDatabaseQuery(driver, config.migrate.Project)

			if config.migrate.Apply {
				handle, err := db.ConnectWithRetry(ctx, config.db)
				if err != nil {
					return errors.Wrap(err, "error connecting to database")
				}

				fmt.Println(query)

				// error is ignored but printed
				if _, err := handle.Exec(query); err != nil {
					fmt.Println("notice:", err)
					return nil
				}
				return nil
			}
			fmt.Println(query)
			return nil
		},
	}
}
