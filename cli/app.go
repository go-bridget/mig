package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/SentimensRG/sigctx"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

var (
	errNoCommand = errors.New("no command found")
)

// NewApp creates a new App instance
func NewApp(name string) *App {
	return &App{
		Name:     name,
		commands: make(map[string]CommandInfo),
	}
}

// Run passes os.Args without the command name to RunWithArgs()
func (app *App) Run() error {
	return app.RunWithArgs(os.Args[1:])
}

// RunWithArgs is a cli entrypoint which sets up a cancellable context for the command
func (app *App) RunWithArgs(args []string) error {
	ctx := sigctx.New()
	commands := parseCommands(args)
	command, err := app.findCommand(commands, "start")
	if err != nil {
		if !errors.Is(err, errNoCommand) {
			app.Help()
		}
		return err
	}

	flag.Usage = func() {
		app.HelpCommand(command)
	}

	// bind command specific flags
	if command.Bind != nil {
		command.Bind(ctx)
	}

	// parse flags and set from environment
	if err := Parse(); err != nil {
		app.HelpCommand(command)
		return err
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
		err = command.Run(ctx, commands)
		// don't print help with standard "context canceled" exit
		if err != nil && !errors.Is(err, context.Canceled) {
			app.HelpCommand(command)
			return err
		}
		return nil
	}

	return errors.New("Missing Run() for command")
}

// Help prints out registered commands for app
func (app *App) Help() {
	fmt.Println("Usage:", app.Name, "(command) [--flags]")
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
	fmt.Println("Usage:", app.Name, "(command) [--flags]")
	fmt.Println()

	maxLen := len(command.Name)
	pad := "   "
	format := pad + "%-" + fmt.Sprintf("%d", maxLen+3) + "s %s\n"
	fmt.Printf(format, command.Name, command.Title)
	fmt.Println()
	PrintDefaults()
	fmt.Println()
}

// AddCommand adds a command to the app
func (app *App) AddCommand(name, title string, constructor func() *Command) {
	info := CommandInfo{
		Name:  name,
		Title: title,
		New:   constructor,
	}
	app.commands[name] = info
}

// findCommand finds a command for the app
func (app *App) findCommand(commands []string, fallback string) (*Command, error) {
	spawn := func(info CommandInfo) (*Command, error) {
		command := info.New()
		if command.Name == "" {
			command.Name = info.Name
		}
		if command.Title == "" {
			command.Title = info.Title
		}
		return command, nil
	}

	// This is just fully naive, we use the first command we find
	// but we could be smarter and have sub commands? Maybe one day.
	if len(commands) > 0 {
		commandName := commands[0]
		if info, ok := app.commands[commandName]; ok {
			return spawn(info)
		}
	}
	if info, ok := app.commands[fallback]; ok {
		return spawn(info)
	}
	return nil, fmt.Errorf("Can't find commands: [%s, default=%s], err=%w", errNoCommand)
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
