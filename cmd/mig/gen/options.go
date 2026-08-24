package gen

import (
	"github.com/go-bridget/mig/logger"
)

// Options contains code generation options.
type Options struct {
	Language string
	Schema   string
	Output   string

	// LogFn receives code generation progress. It must be set;
	// Render returns logger.ErrNoLogFn when it is nil.
	LogFn logger.LogFn

	Go struct {
		FillJSON bool
		SkipJSON bool
	}

	PHP struct {
		Namespace string
	}
}

// Logf writes a line to the bound sink.
func (options *Options) Logf(format string, args ...any) {
	options.LogFn(format, args...)
}

// checkLogFn reports whether a logging sink has been bound.
func (options *Options) checkLogFn() error {
	if options.LogFn == nil {
		return logger.ErrNoLogFn
	}
	return nil
}
