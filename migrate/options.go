package migrate

import (
	flag "github.com/spf13/pflag"
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
}

// NewOptions creates a new Options instance with default values.
func NewOptions() *Options {
	return (&Options{}).Init()
}

func (options *Options) Init() *Options {
	options.Path = "schema"
	return options
}

// Bind registers migration flags on the given FlagSet.
func (options *Options) Bind(fs *flag.FlagSet) {
	fs.StringVar(&options.Path, "path", options.Path, "Project path for database migrations")
	fs.StringVar(&options.Project, "project", options.Project, "Project name for migrations (db key)")
	fs.StringVarP(&options.Filename, "filename", "f", options.Filename, "Single file sql for migrations")
	fs.BoolVar(&options.Apply, "apply", options.Apply, "false = print migrations, true = run migrations")
	fs.BoolVar(&options.Verbose, "verbose", options.Verbose, "false = print summary, true = print details")
}
