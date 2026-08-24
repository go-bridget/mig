package gen

// Logger receives what a render did. The shape is slog's, so a *slog.Logger
// satisfies it as it stands, but any logger with these two methods does -
// the package doesn't name slog.
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

// Options contains code generation options.
type Options struct {
	Language string
	Schema   string
	Output   string

	// Logger receives the progress of a render, which is the files it
	// wrote. It's required - take an Options from NewOptions and it's the
	// argument you passed.
	Logger Logger

	Go struct {
		FillJSON bool
		SkipJSON bool
	}

	PHP struct {
		Namespace string
	}
}

// NewOptions creates a new Options instance with default values, reporting
// through log.
func NewOptions(log Logger) *Options {
	return &Options{
		Language: "go",
		Output:   "types",
		Logger:   log,
	}
}
