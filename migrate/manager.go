package migrate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/jmoiron/sqlx"
)

const (
	// Pattern is the glob selecting migrations in an fs.FS, and the one a
	// Manager uses until Load says otherwise. Naming no directory, it is
	// matched against base names, so migrations are found at whatever depth
	// they sit - an embed.FS handed over without an fs.Sub included.
	Pattern = "*.up.sql"

	// Table is the name of the bookkeeping table.
	Table = "migrations"

	// StatusOK is the Status of a migration that ran to its last statement.
	// Any other value is the error of the one that failed.
	StatusOK = "ok"

	// projectLimit is the width of the project column on mysql and
	// postgres, which is what a project name has to fit in.
	projectLimit = 16
)

var (
	// ErrNoDB is returned by NewManager when the handle is nil.
	ErrNoDB = errors.New("migrate: nil database handle")

	// ErrNoFS is returned by NewManager when the migrations fs.FS is nil.
	ErrNoFS = errors.New("migrate: nil migrations filesystem")

	// ErrNoProject is returned by NewManager when the project name is
	// empty, or longer than the 16 characters the column holds.
	ErrNoProject = errors.New("migrate: invalid project name")

	// ErrDriver is returned by NewManager when the driver of the handle is
	// one this package has no bookkeeping schema for.
	ErrDriver = errors.New("migrate: unsupported driver")

	// ErrNoMigrations is returned by Apply when nothing in the filesystem
	// matches the pattern. Applying nothing is not a run, it is a project
	// pointed at the wrong place, and it is the one failure that otherwise
	// looks exactly like success.
	ErrNoMigrations = errors.New("migrate: no migrations found")

	// ErrNoStatements is returned by Apply for a migration that parses to
	// no statements at all, which is a file someone meant to write.
	ErrNoStatements = errors.New("migrate: migration has no statements")
)

// Manager applies the migrations of one project to one database and keeps the
// record of what it applied.
//
// A Manager is safe to use from several processes at once: each file is applied
// under an advisory lock keyed by project and filename.
type Manager struct {
	db      *sqlx.DB
	fsys    fs.FS
	project string

	// pattern selects the migrations of the filesystem. Load sets it.
	pattern string

	// driver is the normalized name of the engine behind db, which picks
	// the bookkeeping schema and the locking strategy.
	driver string
}

// NewManager returns a Manager applying migrations to db, recording them under
// project.
//
// It fails if db or migrations is nil, if project is empty or longer than 16
// characters, or if the driver of db is one this package has no bookkeeping
// schema for. The supported drivers are mysql, postgres (also pgx, postgresql)
// and sqlite (also sqlite3).
//
// No query is made and no file is read until Apply or List is called.
func NewManager(db *sqlx.DB, migrations fs.FS, project string) (*Manager, error) {
	if db == nil {
		return nil, ErrNoDB
	}
	if migrations == nil {
		return nil, ErrNoFS
	}
	if project == "" || len(project) > projectLimit {
		return nil, fmt.Errorf("%w: %q", ErrNoProject, project)
	}

	driverName, ok := driver(db.DriverName())
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDriver, db.DriverName())
	}

	return &Manager{
		db:      db,
		fsys:    migrations,
		project: project,
		pattern: Pattern,
		driver:  driverName,
	}, nil
}

// Project returns the migration scope this Manager records under.
func (m *Manager) Project() string {
	return m.project
}

// Load sets the glob selecting the migrations to apply, in place of Pattern.
// An empty pattern puts Pattern back.
//
//	m.Load("schema/*.sql")
//
// The syntax is path.Match's, and naming a directory is what decides the depth:
// a pattern with a "/" in it is matched against the whole path, so
// "schema/*.sql" is one directory deep and nothing else; a pattern without one
// is matched against base names, so "*.sql" is every depth there is.
//
// Load reads nothing and says nothing about what it was given: a malformed
// pattern is reported by Apply and List, and a pattern matching no migrations is
// ErrNoMigrations from Apply.
func (m *Manager) Load(pattern string) {
	if pattern == "" {
		pattern = Pattern
	}
	m.pattern = pattern
}

// files returns the migrations this Manager applies.
func (m *Manager) files() ([]string, error) {
	return Files(m.fsys, m.pattern)
}

// ensureTable creates the bookkeeping table if it is missing. The DDL runs
// outside a transaction: mysql commits it implicitly anyway, and a migration
// holding a lock on a table that doesn't exist yet is a deadlock waiting to
// happen.
func (m *Manager) ensureTable(ctx context.Context) error {
	stmts, err := tableDDL(m.driver)
	if err != nil {
		return err
	}

	for _, stmt := range stmts {
		if _, err := m.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("error creating %s table: %w", Table, err)
		}
	}

	return nil
}

// rows returns the recorded migrations of the project, keyed by filename.
func (m *Manager) rows(ctx context.Context) (map[string]Migration, error) {
	query := m.db.Rebind(
		"select project, filename, statement_index, status from " + Table + " where project = ?",
	)

	var recorded []Migration
	if err := m.db.SelectContext(ctx, &recorded, query, m.project); err != nil {
		return nil, fmt.Errorf("error reading %s table: %w", Table, err)
	}

	result := make(map[string]Migration, len(recorded))
	for _, row := range recorded {
		result[row.Filename] = row
	}

	return result, nil
}

// pending returns the migrations of the filesystem with nothing recorded
// against them. It stands in for a report when the migrations table can't be
// read, so a caller always has the list to print alongside the error.
func (m *Manager) pending(files []string) []Migration {
	result := make([]Migration, 0, len(files))
	for _, name := range files {
		result = append(result, Migration{
			Project:        m.project,
			Filename:       name,
			StatementIndex: -1,
		})
	}
	return result
}

// report gathers one Migration per file, carrying the row of the migrations
// table where there is one, plus the rows whose file is gone.
func (m *Manager) report(ctx context.Context) ([]Migration, error) {
	files, err := m.files()
	if err != nil {
		return nil, err
	}

	recorded, err := m.rows(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Migration, 0, len(files)+len(recorded))
	for _, name := range files {
		row, ok := recorded[name]
		if !ok {
			// Never run: no row, so nothing but the sentinel index.
			row = Migration{Project: m.project, Filename: name, StatementIndex: -1}
		}
		delete(recorded, name)

		result = append(result, row)
	}

	// Whatever is left has a row but no file behind it any more.
	for _, row := range recorded {
		result = append(result, row)
	}
	sortByFilename(result)

	return result, nil
}
