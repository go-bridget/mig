package main

import (
	"fmt"
	"os"

	"github.com/go-bridget/mig/cli"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

// mig version number
var version string

func main() {
	app := cli.NewApp("mig")
	app.AddCommand("migrate", "Apply SQL migrations to database", migrateCmd)
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("An error occured: %s", err)
		fmt.Println()
		os.Exit(1)
	}
}
