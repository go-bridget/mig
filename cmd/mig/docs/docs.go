package docs

import (
	"context"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/cmd/mig/internal"
	"github.com/go-bridget/mig/db"
)

const Name = "Generate markdown docs from DB schema"

func New() *cli.Command {
	var config struct {
		db db.Options

		schema   string
		output   string
		filename string
	}

	return &cli.Command{
		Bind: func(_ context.Context) {
			(&config.db).Init().Bind()
			cli.StringVar(&config.schema, "schema", "", "Database schema to list")
			cli.StringVar(&config.output, "output", "docs", "Output folder where to generate docs")
			cli.StringVar(&config.filename, "output-file", "", "Output as single filename")
		},
		Run: func(ctx context.Context, commands []string) error {
			tables, err := internal.ListTables(ctx, config.db, config.schema)
			if err != nil {
				return err
			}
			return renderMarkdown(config.output, config.filename, tables)
		},
	}
}
