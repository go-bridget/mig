package cli

import (
	"context"

	"github.com/namsral/flag"
)

// This package builds on namsral/flag functionality.
//
// We omit exposing functions which return a pointer from the namsral/flag
// public API, so we can encourage defining the flag values into structs.
//
// We also don't expose a lot of the namsral/flag functionality here, since we
// expect that it will be wrapped in cli.App, and it makes little sense to use
// namsral/flag "primitives" when we're creating a higher level abstraction.
//
// That being said, it's still possible to use the namsral/flag API, but there
// should be little reason to do that.

var (
	BoolVar     = flag.BoolVar
	DurationVar = flag.DurationVar
	Int64Var    = flag.Int64Var
	IntVar      = flag.IntVar
	StringVar   = flag.StringVar
	Uint64Var   = flag.Uint64Var
	UintVar     = flag.UintVar
)

type (
	// App is the cli entrypoint
	App struct {
		Name string

		commands map[string]commandInfo
	}

	// Command is an individual command
	Command struct {
		Name, Title string

		Bind func(context.Context)
		Init func(context.Context) error
		Run  func(context.Context, []string) error
	}

	commandInfo struct {
		Name  string
		Title string
		New   func() *Command
	}
)