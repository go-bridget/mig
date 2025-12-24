package gen

import (
	"context"
	"os"
	"slices"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/cmd/mig/internal"
	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/model"
)

const Name = "Generate source code from DB schema"

func New() *cli.Command {
	var config struct {
		db      *db.Options
		options Options
	}
	config.options.Language = "go"
	config.options.Output = "types"

	return &cli.Command{
		Bind: func(_ context.Context) {
			config.db = db.NewOptions()
			config.db.Bind()

			cli.StringVar(&config.options.Language, "lang", "go", "Programming language")
			cli.StringVar(&config.options.Output, "output", "model", "Output folder where to generate types")

			cli.BoolVar(&config.options.Go.FillJSON, "go.fill-json", false, "Fill JSON tags (go)")
			cli.BoolVar(&config.options.Go.SkipJSON, "go.skip-json", false, "Skip JSON tags (go)")
		},
		Run: func(ctx context.Context, commands []string) error {
			tables, err := internal.ListTables(ctx, config.db)
			if err != nil {
				return err
			}
			return cmdGen(config.options, tables)
		},
	}
}

func cmdGen(options Options, tables []*model.Table) error {
	language := options.Language
	languages := []string{
		"go",
	}
	if !slices.Contains(languages, language) {
		return errors.Errorf("invalid language: %s", language)
	}

	// create output folder
	if err := os.MkdirAll(options.Output, 0755); err != nil {
		return err
	}

	switch language {
	case "go":
		return Render(options, tables)
	}
	return nil
}
