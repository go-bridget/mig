package cli

import (
	"fmt"

	"github.com/SentimensRG/sigctx"
	"github.com/pkg/errors"
)

// NewApp creates a new App instance
func NewApp(name string) *App {
	return &App{
		Name:     name,
		commands: make(map[string]commandInfo),
	}
}

// Run is a cli entrypoint which sets up a cancellable context for the command
func (app *App) Run(args []string) error {
	ctx := sigctx.New()
	commands := parseCommands(args)
	if len(commands) == 0 {
		app.Help()
		return nil
	}

	command, err := app.FindCommand(commands)
	if err != nil {
		app.Help()
		return err
	}

	// bind command specific flags
	if command.Bind != nil {
		command.Bind(ctx)
	}
	Parse()

	contains := func(haystack []string, needle string) bool {
		for _, hay := range haystack {
			if hay == needle {
				return true
			}
		}
		return false
	}

	// print help for command(s)
	if contains(args, "-h") || contains(args, "-?") || contains(args, "--help") {
		app.HelpCommand(command)
		return nil
	}

	// initialize command (pre-load data, etc.)
	if command.Init != nil {
		if err := command.Init(ctx); err != nil {
			app.HelpCommand(command)
			return err
		}
	}

	// Run command if defined
	if command.Run != nil {
		if err := command.Run(ctx, commands); err != nil {
			app.HelpCommand(command)
			return err
		}
		return nil
	}

	return errors.New("Missing Run() for command")
}

// Help prints out registered commands for app
func (app *App) Help() {
	fmt.Println("Usage:", app.Name, "(command) [-flags]")
	fmt.Println("Available commands:")
	fmt.Println()

	maxLen := 0
	for _, command := range app.commands {
		if len(command.Name) > maxLen {
			maxLen = len(command.Name)
		}
	}
	pad := "   "
	format := pad + "%-" + fmt.Sprintf("%d", maxLen+3) + "s %s\n"
	for _, command := range app.commands {
		fmt.Printf(format, command.Name, command.Title)
	}
	fmt.Println()
}

// Help prints out registered commands for app
func (app *App) HelpCommand(command *Command) {
	fmt.Println("Usage:", app.Name, "(command) [-flags]")
	fmt.Println()

	maxLen := 0
	for _, command := range app.commands {
		if len(command.Name) > maxLen {
			maxLen = len(command.Name)
		}
	}
	pad := "   "
	format := pad + "%-" + fmt.Sprintf("%d", maxLen+3) + "s %s\n"
	fmt.Printf(format, command.Name, command.Title)
	fmt.Println()
	PrintDefaults()
	fmt.Println()
}

// AddCommand adds a command to the app
func (app *App) AddCommand(name, title string, constructor func() *Command) {
	info := commandInfo{
		Name:  name,
		Title: title,
		New:   constructor,
	}
	app.commands[name] = info
}

// FindCommand finds a command for the app
func (app *App) FindCommand(commands []string) (*Command, error) {
	// This is just fully naive, we use the first command we find
	// but we could be smarter and have sub commands? Maybe one day.
	for _, commandName := range commands {
		info, ok := app.commands[commandName]
		if ok {
			command := info.New()
			if command.Name == "" {
				command.Name = info.Name
			}
			if command.Title == "" {
				command.Title = info.Title
			}
			return command, nil
		}
	}
	return nil, errors.New("no command found")
}

// parseCommand cleans up args[], returning only commands
//
// It looks inside args[] up until the first parameter that starts with "-", a
// flag parameter. We asume all the parameters before are command names.
//
// Example: [a, b, -c, d, e] becomes [a, b]

func parseCommands(args []string) []string {
	result := []string{}
	for _, v := range args {
		if v[0:1] == "-" {
			break
		}
		result = append(result, v)
	}
	return result
}
