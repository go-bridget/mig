package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-bridget/mig/cli"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

// mig version number
var Version string

func main() {
	app := cli.NewApp("mig")
	app.AddCommand("create", "Create database schema SQL", createCmd)
	app.AddCommand("migrate", "Apply SQL migrations to database", migrateCmd)
	app.AddCommand("version", "Print version", func() *cli.Command {
		return &cli.Command{
			Run: func(_ context.Context, _ []string) error {
				fmt.Println(app.Name, "version", Version)
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
