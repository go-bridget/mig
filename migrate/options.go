package migrate

import (
	"github.com/go-bridget/mig/cli"
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

func NewOptions() *Options {
	return (&Options{}).Init()
}

func (options *Options) Init() *Options {
	options.Path = "schema"
	return options
}

func (options *Options) Bind() {
	cli.StringVar(&options.Path, "path", options.Path, "Project path for database migrations")
	cli.StringVar(&options.Project, "project", options.Project, "Project name for migrations (db key)")
	cli.StringVarP(&options.Filename, "filename", "f", options.Filename, "Single file sql for migrations")
	cli.BoolVar(&options.Apply, "apply", options.Apply, "false = print migrations, true = run migrations")
	cli.BoolVar(&options.Verbose, "verbose", options.Verbose, "false = print summary, true = print details")
}
