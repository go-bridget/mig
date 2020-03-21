package cli

import (
	"github.com/SentimensRG/sigctx"
	"github.com/namsral/flag"
)

// App is a cli entrypoint which sets up a cancellable context for the action
func (app *App) Run(args []string) error {
	flag.Parse()
	// run action flow
	if app.Action != nil {
		ctx := sigctx.New()
		// init actions
		if app.Init != nil {
			if err := app.Init(ctx); err != nil {
				return err
			}
		}
		// run actions
		commands := parseCommands(args)
		return app.Action(ctx, commands)
	}
	// command help
	if app.PrintCommands != nil {
		app.PrintCommands()
	}
	// flags help
	flag.PrintDefaults()
	return nil
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
