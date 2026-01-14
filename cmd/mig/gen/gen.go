package gen

import (
	"context"
	"os"
	"slices"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/db/introspect"
	"github.com/go-bridget/mig/model"
)

// Name is the command title.
const Name = "Generate source code from DB schema"

// New creates a new gen command.
func New() *cli.Command {
	var config struct {
		db      *db.Options
		options Options
	}
	config.options.Language = "go"
	config.options.Output = "types"

	return &cli.Command{
		Name:  "gen",
		Title: Name,
		Bind: func(fs *flag.FlagSet) {
			config.db = db.NewOptions()
			config.db.Bind(fs)

			fs.StringVar(&config.options.Language, "lang", "go", "Programming language")
			fs.StringVar(&config.options.Output, "output", "model", "Output folder where to generate types")

			fs.BoolVar(&config.options.Go.FillJSON, "go.fill-json", false, "Fill JSON tags (go)")
			fs.BoolVar(&config.options.Go.SkipJSON, "go.skip-json", false, "Skip JSON tags (go)")
		},
		Run: func(ctx context.Context, args []string) error {
			handle, err := db.ConnectWithRetry(ctx, config.db)
			if err != nil {
				return err
			}

			desc, err := introspect.NewDescriber(handle)
			if err != nil {
				return err
			}

			tables, err := introspect.ListTablesWithColumns(ctx, handle, desc)
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
