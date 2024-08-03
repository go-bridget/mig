package migrate

import (
	"github.com/go-bridget/mig/cli"
)

func NewOptions() *Options {
	return (&Options{}).Init()
}

func (options *Options) Init() *Options {
	options.Path = "schema"
	return options
}

func (options *Options) Bind() {
	cli.StringVar(&options.Path, "migrate-path", options.Path, "Project path for database migrations")
	cli.StringVar(&options.Project, "project", options.Project, "Project name for migrations")
	cli.StringVarP(&options.Filename, "filename", "f", options.Filename, "Single file sql for migrations")
	cli.BoolVar(&options.Apply, "apply", options.Apply, "false = print migrations, true = run migrations")
	cli.BoolVar(&options.Verbose, "verbose", options.Verbose, "false = print summary, true = print details")
}
