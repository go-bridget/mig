package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/titpetric/cli"

	"github.com/go-bridget/mig/cmd/mig/create"
	"github.com/go-bridget/mig/cmd/mig/docs"
	"github.com/go-bridget/mig/cmd/mig/filter"
	"github.com/go-bridget/mig/cmd/mig/gen"
	"github.com/go-bridget/mig/cmd/mig/lint"
	"github.com/go-bridget/mig/cmd/mig/migrate"
)

// mig build info
var (
	BuildVersion string
	BuildTime    string
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	app := cli.NewApp("mig")

	// The SQL a command prints goes to stdout, for a script to read; what
	// a run did goes here.
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	app.AddCommand("create", create.Name, create.New(log))
	app.AddCommand("migrate", migrate.Name, migrate.New(log))
	app.AddCommand("docs", docs.Name, docs.New(log))
	app.AddCommand("filter", filter.Name, filter.New)
	app.AddCommand("lint", lint.Name, lint.New(log))
	app.AddCommand("gen", gen.Name, gen.New(log))

	app.AddCommand("version", "Print version", func() *cli.Command {
		return &cli.Command{
			Run: func(_ context.Context, _ []string) error {
				fmt.Println(app.Name)
				fmt.Println()
				fmt.Println("build version ", BuildVersion)
				fmt.Println("build time    ", BuildTime)
				return nil
			},
		}
	})

	return app.Run()
}
