package migrate

import (
	flag "github.com/spf13/pflag"

	"github.com/go-bridget/mig/logger"
)

// Options include migration options.
type Options struct {
	// Path contains sql files with your projects migrations.
	Path string

	// Project contains the project name for tracking migrations.
	Project string

	// Filename imports a single file as a migration source.
	// If filled, it's preferred over path.
	Filename string

	// Apply will apply the migration to the configured database.
	Apply bool

	// Verbose will output more details about migration execution.
	Verbose bool

	// LogFn receives migration progress and SQL output. It must be set;
	// entry points return logger.ErrNoLogFn when it is nil.
	LogFn logger.LogFn
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

// NewOptions creates a new Options instance with default values.
func NewOptions() *Options {
	return &Options{
		Path: "schema",
	}
}

// Bind registers migration flags on the given FlagSet.
func (options *Options) Bind(fs *flag.FlagSet) {
	fs.StringVar(&options.Path, "path", options.Path, "Project path for database migrations")
	fs.StringVar(&options.Project, "project", options.Project, "Project name for migrations (db key)")
	fs.StringVarP(&options.Filename, "filename", "f", options.Filename, "Single file sql for migrations")
	fs.BoolVar(&options.Apply, "apply", options.Apply, "false = print migrations, true = run migrations")
	fs.BoolVar(&options.Verbose, "verbose", options.Verbose, "false = print summary, true = print details")
}
