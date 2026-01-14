package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

var (
	errNoCommand = errors.New("no command found")
)

// App is the cli entrypoint.
type App struct {
	Name string

	commands map[string]CommandInfo
}

// NewApp creates a new App instance.
func NewApp(name string) *App {
	return &App{
		Name:     name,
		commands: make(map[string]CommandInfo),
	}
}

// Run passes os.Args without the command name to RunWithArgs().
func (app *App) Run() error {
	return app.RunWithArgs(os.Args[1:])
}

// RunWithArgs is a cli entrypoint which sets up a cancellable context for the command.
func (app *App) RunWithArgs(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	commands := parseCommands(args)
	command, err := app.findCommand(commands, "")
	if err != nil {
		app.Help()
		return err
	}

	// Create a scoped FlagSet for this command
	fs := flag.NewFlagSet(command.Name, flag.ContinueOnError)
	fs.Usage = func() {
		app.HelpCommand(fs, command)
	}

	// bind command specific flags
	if command.Bind != nil {
		command.Bind(fs)
	}

	// Strip command names from args before parsing
	flagArgs := args
	if len(commands) > 0 {
		flagArgs = args[len(commands):]
	}

	// parse flags and set from environment
	if err := ParseWithFlagSet(fs, flagArgs); err != nil {
		app.HelpCommand(fs, command)
		return err
	}

	// Run command if defined
	if command.Run != nil {
		// Pass remaining command tokens plus any positional args after flags
		// commands[1:] skips the matched command name, leaving sub-commands/args
		remainingArgs := fs.Args()
		if len(commands) > 1 {
			remainingArgs = append(commands[1:], remainingArgs...)
		}
		err = command.Run(ctx, remainingArgs)
		// don't print help with standard "context canceled" exit
		if err != nil && !errors.Is(err, context.Canceled) {
			app.HelpCommand(fs, command)
			return err
		}
		return nil
	}

	return errors.New("Missing Run() for command")
}

// Help prints out registered commands for app.
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

// HelpCommand prints out help for a specific command.
func (app *App) HelpCommand(fs *flag.FlagSet, command *Command) {
	fmt.Println("Usage:", app.Name, command.Name, "[--flags]")
	fmt.Println()
	fs.PrintDefaults()
	fmt.Println()
}

// AddCommand adds a command to the app.
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
	if len(commands) > 0 {
		return nil, fmt.Errorf("unknown command: %q", commands[0])
	}
	return nil, fmt.Errorf("no command specified")

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
		if len(v) > 0 && v[0:1] == "-" {
			break
		}
		result = append(result, v)
	}
	return result
}
