package gen

import (
	"context"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/cmd/mig/gen/model"
	"github.com/go-bridget/mig/cmd/mig/internal"
	"github.com/go-bridget/mig/db"
)

const Name = "Generate source code from DB schema"

func New() *cli.Command {
	var config struct {
		db      db.Options
		options model.Options
	}
	config.options.Language = "go"
	config.options.Output = "types"

	return &cli.Command{
		Bind: func(_ context.Context) {
			(&config.db).Init().Bind()
			cli.StringVar(&config.options.Language, "lang", "go", "Programming language")
			cli.StringVar(&config.options.Schema, "schema", "", "Database schema to list")
			cli.StringVar(&config.options.Output, "output", "types", "Output folder where to generate types")
			cli.BoolVar(&config.options.FillJSON, "go.fill-json", false, "Fill JSON tags (go)")
		},
		Run: func(ctx context.Context, commands []string) error {
			tables, err := internal.ListTables(ctx, config.db, config.options.Schema)
			if err != nil {
				return err
			}
			return render(config.options, tables)
		},
	}
}
