package create

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/titpetric/cli"

	"github.com/go-bridget/mig/db"
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
		db *db.Options

		project string
		apply   bool
	}

	return &cli.Command{
		Name:  "create",
		Title: Name,
		Bind: func(fs *cli.FlagSet) {
			config.db = db.NewOptions()
			config.db.Bind(fs)

			fs.StringVar(&config.project, "project", "", "Project name for migrations (db key)")
			fs.BoolVar(&config.apply, "apply", false, "false = print query, true = run query")
		},
		Run: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				config.project = args[0]
			}

			if config.project == "" {
				return errors.Errorf("Specify project name as first argument to create")
			}

			driver, _ := db.ParseDSN(config.db.Credentials.DSN)
			query := createDatabaseQuery(driver, config.project)

			if config.apply {
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
