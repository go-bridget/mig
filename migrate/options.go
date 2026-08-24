package migrate

import (
	"fmt"
	"io"
	"log/slog"
	"os"

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

	// Logger receives the progress of a migration. A nil Logger reports
	// nothing.
	Logger *slog.Logger

	// Output receives the SQL the migrations are made of, which Print
	// writes and Verbose adds to a run. A nil Output writes to stdout.
	Output io.Writer
}

// log returns the logger to report through, which drops what it's given
// when the caller bound none.
func (options *Options) log() *slog.Logger {
	if options.Logger == nil {
		return slog.New(slog.DiscardHandler)
	}
	return options.Logger
}

// output returns where the SQL goes. It's stdout by default, so a caller
// can pipe the migrations of a project into a database.
func (options *Options) output() io.Writer {
	if options.Output == nil {
		return os.Stdout
	}
	return options.Output
}

// printf writes a line of SQL to the output.
func (options *Options) printf(format string, args ...any) {
	fmt.Fprintf(options.output(), format+"\n", args...)
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
