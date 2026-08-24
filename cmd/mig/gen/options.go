package gen

import (
	"log/slog"
)

// Options contains code generation options.
type Options struct {
	Language string
	Schema   string
	Output   string

	// Logger receives the progress of a render, which is the files it
	// wrote. A nil Logger reports nothing.
	Logger *slog.Logger

	Go struct {
		FillJSON bool
		SkipJSON bool
	}

	PHP struct {
		Namespace string
	}
}

// log returns the logger to report through, which drops what it's given
// when the caller bound none.
func (options *Options) log() *slog.Logger {
	if options.Logger == nil {
		return slog.New(slog.DiscardHandler)
	}
	return options.Logger
}
