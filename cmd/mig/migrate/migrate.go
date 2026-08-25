package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing/fstest"
	"text/tabwriter"

	"github.com/pkg/errors"

	"github.com/titpetric/cli"

	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/migrate"
)

// Name is the command title.
const Name = "Apply SQL migrations to database"

// New creates a new migrate command.
func New() *cli.Command {
	var config struct {
		db *db.Options

		path     string
		project  string
		filename string
		glob     string
		apply    bool
		list     bool
		verbose  bool
	}

	return &cli.Command{
		Name:  "migrate",
		Title: Name,
		Bind: func(fs *cli.FlagSet) {
			config.db = db.NewOptions()
			config.db.Bind(fs)

			fs.StringVar(&config.path, "path", "schema", "Project path for database migrations")
			fs.StringVar(&config.project, "project", "", "Project name for migrations (db key)")
			fs.StringVarP(&config.filename, "filename", "f", "", "Single "+migrate.Pattern+" file as the migration source")
			fs.StringVar(&config.glob, "glob", migrate.Pattern, "Glob selecting migrations under the path")
			fs.BoolVar(&config.apply, "apply", false, "false = print migrations, true = run migrations")
			fs.BoolVar(&config.list, "list", false, "print what the migrations table records and exit")
			fs.BoolVar(&config.verbose, "verbose", false, "false = print summary, true = print details")
		},
		Run: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				config.project = args[0]
			}

			if config.project == "" {
				return errors.New("Specify project name as first argument to migrate")
			}

			fsys, glob, err := open(config.path, config.filename, config.glob)
			if err != nil {
				return fmt.Errorf("error loading migrations: %w", err)
			}

			// Printing needs no database, so it happens before we
			// reach for one.
			if !config.apply && !config.list {
				return migrate.Print(os.Stdout, fsys, glob)
			}

			handle, err := db.ConnectWithRetry(ctx, config.db)
			if err != nil {
				return fmt.Errorf("error connecting to database: %w", err)
			}
			defer handle.Close()

			manager, err := migrate.NewManager(handle, fsys, config.project)
			if err != nil {
				return err
			}
			manager.Load(glob)

			if config.list {
				recorded, err := manager.List(ctx)
				if err != nil {
					return err
				}
				return report(recorded)
			}

			if config.verbose {
				if err := migrate.Print(os.Stdout, fsys, glob); err != nil {
					return err
				}
			}

			applied, applyErr := manager.Apply(ctx)
			if err := report(applied); err != nil {
				return err
			}
			return applyErr
		},
	}
}

// open resolves the migrations to run, and the glob selecting them. A filename
// wins over a path, and brings its own glob: the file is applied whatever it is
// called, and the rest of its directory stays out of the run.
func open(path, filename, glob string) (fs.FS, string, error) {
	if filename == "" {
		stat, err := os.Stat(path)
		if err != nil {
			return nil, "", fmt.Errorf("path: '%s': %w", path, err)
		}
		if !stat.IsDir() {
			return nil, "", fmt.Errorf("path is not a directory: '%s'", path)
		}
		return os.DirFS(path), glob, nil
	}

	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, "", err
	}

	base := filepath.Base(filename)

	return fstest.MapFS{base: &fstest.MapFile{Data: contents}}, base, nil
}

// report prints what the migrations table records for each migration. An empty
// status is one it has no row for.
func report(recorded []migrate.Migration) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, item := range recorded {
		status := item.Status
		if status == "" {
			status = "-"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\n", item.Filename, item.StatementIndex, status)
	}
	return w.Flush()
}
