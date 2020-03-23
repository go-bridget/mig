package migrate

import (
	"github.com/go-bridget/mig/cli"
)

func (options *Options) Bind() {
	cli.StringVar(&options.Path, "migrate-path", "schema", "Project path for database migrations")
	cli.StringVar(&options.Project, "project", "", "Project name for migrations")
	cli.BoolVar(&options.Apply, "apply", false, "false = print migrations, true = run migrations")
	cli.BoolVar(&options.Verbose, "verbose", false, "false = print summary, true = print details")
}
