package cli

import (
	"context"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

// This package builds on spf13/pflag functionality.
//
// We omit exposing functions which return a pointer from the spf13/pflag
// public API, so we can encourage defining the flag values into structs.
//
// We also don't expose a lot of the spf13/pflag functionality here, since we
// expect that it will be wrapped in cli.App, and it makes little sense to use
// spf13/pflag "primitives" when we're creating a higher level abstraction.
//
// That being said, it's still possible to use the spf13/pflag API, but there
// should be little reason to do that.

var (
	BoolVar     = flag.BoolVar
	DurationVar = flag.DurationVar
	Int64Var    = flag.Int64Var
	IntVar      = flag.IntVar
	StringVar   = flag.StringVar
	Uint64Var   = flag.Uint64Var
	UintVar     = flag.UintVar

	BoolVarP     = flag.BoolVarP
	DurationVarP = flag.DurationVarP
	Int64VarP    = flag.Int64VarP
	IntVarP      = flag.IntVarP
	StringVarP   = flag.StringVarP
	Uint64VarP   = flag.Uint64VarP
	UintVarP     = flag.UintVarP

	PrintDefaults = flag.PrintDefaults
)

type (
	// App is the cli entrypoint
	App struct {
		Name string

		commands map[string]CommandInfo
	}

	// Command is an individual command
	Command struct {
		Name, Title string

		Bind func(context.Context)
		Init func(context.Context) error
		Run  func(context.Context, []string) error
	}

	// CommandInfo is the constructor info for a command
	CommandInfo struct {
		Name  string
		Title string
		New   func() *Command
	}
)

func Parse() error {
	// parse os.Environ() and include it as cli flags to the command line
	for _, v := range os.Environ() {
		vals := strings.SplitN(v, "=", 2)

		// convert DB_DSN to db-dsn (pflag like)
		flagName := vals[0]
		flagName = strings.ToLower(flagName)
		flagName = strings.Replace(flagName, "_", "-", -1)

		// check if destination flag exists or modified
		fn := flag.CommandLine.Lookup(flagName)
		if fn == nil || fn.Changed {
			continue
		}
		if err := fn.Value.Set(vals[1]); err != nil {
			return err
		}
	}
	flag.Parse()
	return nil
}
