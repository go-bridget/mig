package docs

import (
	"context"

	flag "github.com/spf13/pflag"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/db/introspect"
)

// Name is the command title.
const Name = "Generate markdown docs from DB schema"

// New creates a new docs command.
func New() *cli.Command {
	var config struct {
		db *db.Options

		output   string
		filename string
		yaml     bool
		jsonOut  bool
	}

	return &cli.Command{
		Name:  "docs",
		Title: Name,
		Bind: func(fs *flag.FlagSet) {
			config.db = db.NewOptions()
			config.db.Bind(fs)

			fs.StringVar(&config.output, "output", "docs", "Output folder where to generate docs")
			fs.StringVar(&config.filename, "output-file", "", "Output as single filename")
			fs.BoolVar(&config.yaml, "yaml", false, "Output as YAML")
			fs.BoolVar(&config.jsonOut, "json", false, "Output as JSON")
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

			if config.yaml {
				return renderYAML(config.output, config.filename, tables)
			}
			if config.jsonOut {
				return renderJSON(config.output, config.filename, tables)
			}
			return renderMarkdown(config.output, config.filename, tables)
		},
	}
}
