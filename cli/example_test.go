package cli

import (
	"context"
	"fmt"

	flag "github.com/spf13/pflag"
)

// ExampleCommand demonstrates how to create and register a command with scoped flags.
func ExampleCommand() {
	// Create application
	app := NewApp("myapp")

	// Define a command constructor that registers its own flags
	app.AddCommand("greet", "Greet a person", func() *Command {
		var name string
		var greeting string

		cmd := &Command{
			Name:  "greet",
			Title: "Greet a person",
			Bind: func(fs *flag.FlagSet) {
				// Register flags directly on FlagSet
				fs.StringVar(&name, "name", "World", "Name to greet")
				fs.StringVar(&greeting, "greeting", "Hello", "Greeting prefix")
			},
			Run: func(ctx context.Context, args []string) error {
				fmt.Printf("%s, %s!\n", greeting, name)
				return nil
			},
		}

		return cmd
	})

	// Print available commands
	app.Help()

	// Output:
	// Usage: myapp (command) [--flags]
	// Available commands:
	//
	//    greet    Greet a person
}
