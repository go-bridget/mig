package gen

import (
	"context"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/cmd/mig/internal"
	"github.com/go-bridget/mig/db"
)

const Name = "Generate source code from DB schema"

func New() *cli.Command {
	var config struct {
		db db.Options

		lang   string
		schema string
		output string
	}

	return &cli.Command{
		Bind: func(_ context.Context) {
			(&config.db).Init().Bind()
			cli.StringVar(&config.lang, "lang", "go", "Programming language")
			cli.StringVar(&config.schema, "schema", "", "Database schema to list")
			cli.StringVar(&config.output, "output", "types", "Output folder where to generate types")
		},
		Run: func(ctx context.Context, commands []string) error {
			tables, err := internal.ListTables(ctx, config.db, config.schema)
			if err != nil {
				return err
			}
			return render(config.lang, config.schema, config.output, tables)
		},
	}
}
