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

// options is what the flags of the command set, and what open reads to decide
// where the migrations come from.
type options struct {
	db *db.Options

	path      string
	project   string
	filename  string
	glob      string
	recursive bool
	apply     bool
	list      bool
	verbose   bool
}

// New creates a new migrate command.
func New() *cli.Command {
	var config options

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
			fs.BoolVarP(&config.recursive, "recursive", "r", false, "treat the path as a directory name to find under the working directory, and merge every one of them")
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

			fsys, glob, err := config.open()
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

// open resolves the migrations to run, and the glob selecting them. The three
// modes are exclusive and in this order of precedence.
//
// A filename wins over everything, and brings its own glob: the file is applied
// whatever it is called, and the rest of its directory stays out of the run.
//
// Recursive reads the path as the name of a directory rather than one place,
// and answers every directory under the working directory called that, merged
// flat by migrate.Collect - which is the layout of a project whose migrations
// live with the code owning each table. The glob selects within each of them,
// and the answer is flat, so the same glob still selects what came back: see
// migrate.Collect for why a pattern naming a directory is refused.
//
// Otherwise the path is one directory, and the names a run records are relative
// to it.
func (o options) open() (fs.FS, string, error) {
	if o.filename != "" {
		contents, err := os.ReadFile(o.filename)
		if err != nil {
			return nil, "", err
		}

		base := filepath.Base(o.filename)

		return fstest.MapFS{base: &fstest.MapFile{Data: contents}}, base, nil
	}

	if o.recursive {
		fsys, err := migrate.Collect(".", o.path, o.glob)
		if err != nil {
			return nil, "", err
		}
		return fsys, o.glob, nil
	}

	stat, err := os.Stat(o.path)
	if err != nil {
		return nil, "", fmt.Errorf("path: '%s': %w", o.path, err)
	}
	if !stat.IsDir() {
		return nil, "", fmt.Errorf("path is not a directory: '%s'", o.path)
	}

	return os.DirFS(o.path), o.glob, nil
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
