package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	_ "modernc.org/sqlite"

	"github.com/go-bridget/mig/cli"

	"github.com/go-bridget/mig/cmd/mig/create"
	"github.com/go-bridget/mig/cmd/mig/docs"
	"github.com/go-bridget/mig/cmd/mig/gen"
	"github.com/go-bridget/mig/cmd/mig/grants"
	"github.com/go-bridget/mig/cmd/mig/lint"
	"github.com/go-bridget/mig/cmd/mig/migrate"
)

// mig build info
var (
	BuildVersion string
	BuildTime    string
)

func main() {
	app := cli.NewApp("mig")

	app.AddCommand("grants", grants.Name, grants.New)
	app.AddCommand("create", create.Name, create.New)
	app.AddCommand("migrate", migrate.Name, migrate.New)
	app.AddCommand("docs", docs.Name, docs.New)
	app.AddCommand("lint", lint.Name, lint.New)
	app.AddCommand("gen", gen.Name, gen.New)

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
	if err := app.Run(); err != nil {
		fmt.Printf("An error occurred: %s", err)
		fmt.Println()
		os.Exit(1)
	}
}
