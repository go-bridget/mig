package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-bridget/mig/cli"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

// mig build info
var (
	BuildVersion string
	BuildTime    string
)

func main() {
	app := cli.NewApp("mig")
	app.AddCommand("create", "Create database schema SQL", createCmd)
	app.AddCommand("migrate", "Apply SQL migrations to database", migrateCmd)
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
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("An error occured: %s", err)
		fmt.Println()
		os.Exit(1)
	}
}
