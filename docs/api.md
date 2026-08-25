# Package migrate

```go
import (
	"github.com/go-bridget/mig/migrate"
}
```
Package migrate applies SQL migrations to a database and records what it
applied.

A migration is a file of semicolon-separated SQL statements. Files are read
from an fs.FS, selected by the glob "*.up.sql", and applied in lexical order.
Each file is applied in one transaction holding an advisory lock, and the
outcome is written to the migrations table: the project, the filename, the
index of the last statement that ran, and either "ok" or the error of the
statement that failed.

Statements already recorded as applied are skipped, so a file that grew
statements since the last run has only the new ones applied.

```go
m, err := migrate.NewManager(handle, os.DirFS("schema"), "events")
if err != nil {
	return err
}
applied, err := m.Apply(ctx)
for _, item := range applied {
	fmt.Println(item.Filename, item.StatementIndex, item.Status)
}
```

Listing does not apply anything:

```go
list, err := m.List(ctx)
```

Printing the migrations of a project needs no database at all:

```go
err := migrate.Print(os.Stdout, os.DirFS("schema"))
```

## Types

<details>
<summary><code>type Manager</code></summary>

```go
// Manager applies the migrations of one project to one database and keeps the
// record of what it applied, in a table it creates if it is missing.
//
// A Manager is safe to use from several processes at once: each file is applied
// under an advisory lock keyed by project and filename.
type Manager struct {
	db	*sqlx.DB
	fsys	fs.FS
	project	string

	// pattern selects the migrations of the filesystem. Load sets it.
	pattern	string

	// driver is the normalized name of the engine behind db, which picks
	// the bookkeeping schema and the locking strategy.
	driver	string
}
```

</details>

<details>
<summary><code>type Migration</code></summary>

```go
// Migration holds the DB structure for the migration table.
type Migration struct {
	// Project holds a migration scope. You may have several projects
	// migrated within the same migration table.
	Project	string	`db:"project"`

	// Filename logs the file used for storing migrations.
	Filename	string	`db:"filename"`

	// StatementIndex is the current index of applied migrations. It is -1
	// for a migration the table has no record of.
	StatementIndex	int	`db:"statement_index"`

	// Status contains the status of the migrations. It's expected to be
	// 'ok' for a healthy value, and holds the error of the statement that
	// failed otherwise. It is empty for a migration the table has no record
	// of.
	Status	string	`db:"status"`

	// err is the error behind Status, kept so a caller can unwrap what a
	// run returned. It is not a column: a migration read back from the
	// table has only the text in Status to go on.
	err	error
}
```

</details>

## Consts

<details>
<summary><code>const Pattern, Table, StatusOK, projectLimit</code></summary>

```go
// Defaults of a run: what a Manager looks for, where it records what it
// applied, and what it writes there.
const (
	// Pattern is the glob selecting migrations in an fs.FS, and the one a
	// Manager uses until Load says otherwise. Naming no directory, it is
	// matched against base names, so migrations are found at whatever depth
	// they sit - an embed.FS handed over without an fs.Sub included.
	Pattern	= "*.up.sql"

	// Table is the name of the bookkeeping table.
	Table	= "migrations"

	// StatusOK is the Status of a migration that ran to its last statement.
	// Any other value is the error of the one that failed.
	StatusOK	= "ok"

	// projectLimit is the width of the project column on mysql and
	// postgres, which is what a project name has to fit in.
	projectLimit	= 16
)
```

</details>

## Vars

<details>
<summary><code>var MigrationFields</code></summary>

```go
// MigrationFields hold the database column names for Migration{}.
var MigrationFields = []string{"project", "filename", "statement_index", "status"}
```

</details>

<details>
<summary><code>var ErrNoDB, ErrNoFS, ErrNoProject, ErrDriver, ErrNoMigrations, ErrNoStatements</code></summary>

```go
// The errors of the package, each returned for one thing being wrong, so a
// caller can tell a misconfigured Manager from a run that found no work.
var (
	// ErrNoDB is returned by NewManager when the handle is nil.
	ErrNoDB	= errors.New("migrate: nil database handle")

	// ErrNoFS is returned by NewManager when the migrations fs.FS is nil.
	ErrNoFS	= errors.New("migrate: nil migrations filesystem")

	// ErrNoProject is returned by NewManager when the project name is
	// empty, or longer than the 16 characters the column holds.
	ErrNoProject	= errors.New("migrate: invalid project name")

	// ErrDriver is returned by NewManager when the driver of the handle is
	// one this package has no bookkeeping schema for.
	ErrDriver	= errors.New("migrate: unsupported driver")

	// ErrNoMigrations is returned by Apply when nothing in the filesystem
	// matches the pattern. Applying nothing is not a run, it is a project
	// pointed at the wrong place, and it is the one failure that otherwise
	// looks exactly like success.
	ErrNoMigrations	= errors.New("migrate: no migrations found")

	// ErrNoStatements is returned by Apply for a migration that parses to
	// no statements at all, which is a file someone meant to write.
	ErrNoStatements	= errors.New("migrate: migration has no statements")
)
```

</details>

## Function symbols

- `func Files (fsys fs.FS, pattern string) ([]string, error)`
- `func NewManager (db *sqlx.DB, migrations fs.FS, project string) (*Manager, error)`
- `func Print (w io.Writer, fsys fs.FS, pattern string) error`
- `func Statements (fsys fs.FS, name string) ([]string, error)`
- `func (*Manager) Apply (ctx context.Context) ([]Migration, error)`
- `func (*Manager) List (ctx context.Context) ([]Migration, error)`
- `func (*Manager) Load (pattern string)`
- `func (*Manager) Project () string`
- `func (*Migration) SetError (err error)`
- `func (Migration) Err () error`

### Files

Files returns the names of the migrations in fsys matching pattern, sorted.
An empty pattern is Pattern, which finds them wherever they sit.

Names are paths relative to the root of fsys, and that is what a run records
in the filename column. Files shorter than two bytes are skipped: the
shortest valid statement is "--", a comment.

```go
func Files (fsys fs.FS, pattern string) ([]string, error)
```

### NewManager

NewManager returns a Manager applying migrations to db, recording them under
project.

It fails if db or migrations is nil, if project is empty or longer than 16
characters, or if the driver of db is one this package has no bookkeeping
schema for. The supported drivers are mysql, postgres (also pgx, postgresql)
and sqlite (also sqlite3).

No query is made and no file is read until Apply or List is called.

```go
func NewManager (db *sqlx.DB, migrations fs.FS, project string) (*Manager, error)
```

### Print

Print writes every migration in fsys to w as SQL, each file headed by a
comment naming it and each statement by its index, so the output is a script
applying the project by hand. It touches no database.

An empty pattern is Pattern, as it is for Files.

Because Statements replaces uuid() with a fresh UUID, two calls do not
produce the same bytes for a migration using it.

```go
func Print (w io.Writer, fsys fs.FS, pattern string) error
```

### Statements

Statements returns the statements of one migration, with comments removed and
uuid() calls replaced by a fresh UUID literal.

```go
func Statements (fsys fs.FS, name string) ([]string, error)
```

### Apply

Apply runs the migrations that have not been applied yet, in lexical order of
filename, and returns on the first one that fails.

The slice is the run to print: one Migration per file, in order, carrying the
row recorded for it, and never nil. The entry of a file whose statement failed
carries that error.

Each file is applied in one transaction, recording the index of the last
statement that ran whether it succeeded or not, so a later Apply resumes a
failed file from the statement after that one.

```go
func (*Manager) Apply (ctx context.Context) ([]Migration, error)
```

### List

List returns one Migration per file in the migrations filesystem, in lexical
order of filename, along with the rows of the migrations table whose file is
gone. It applies nothing.

A migration the table has a row for carries the StatementIndex and Status of
that row; one it has no row for carries a StatementIndex of -1 and an empty
Status, which is what says it has never run.

List creates the migrations table if it is missing.

```go
func (*Manager) List (ctx context.Context) ([]Migration, error)
```

### Load

Load sets the glob selecting the migrations to apply, in place of Pattern.
An empty pattern puts Pattern back.

The syntax is path.Match's, and naming a directory is what decides the depth:
a pattern with a "/" in it is matched against the whole path, one without it
against base names.

Load reads nothing and says nothing about what it was given: a malformed
pattern is reported by Apply and List, and a pattern matching no migrations is
ErrNoMigrations from Apply.

```go
func (*Manager) Load (pattern string)
```

### Project

Project returns the migration scope this Manager records under.

```go
func (*Manager) Project () string
```

### SetError

SetError records err as the outcome of the migration. Status becomes the
error text, which is what the column holds, or StatusOK when err is nil.

The error itself is kept as well, so Err returns what was passed rather than
a copy of its message.

```go
func (*Migration) SetError (err error)
```

### Err

Err returns the error the migration failed with, or nil when it succeeded or
has not run.

For a migration read back from the migrations table there is no error left to
return, only the text of one, so Err reports the Status as an error.

```go
func (Migration) Err () error
```


## Examples

<section name="ExampleManager_Apply">

### ExampleManager_Apply

ExampleManager_Apply applies every migration of a project, and prints the run
it returns: the file, the index of the last statement that ran, and the status
recorded for it.

```go
func ExampleManager_Apply() {
	fsys := fstest.MapFS{
		"0001-user.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE user (id INTEGER);\nCREATE TABLE post (id INTEGER);\n",
		)},
		"0002-comment.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE comment (id INTEGER);\n",
		)},
	}

	manager, err := migrate.NewManager(exampleDB(), fsys, "example")
	if err != nil {
		panic(err)
	}

	applied, err := manager.Apply(context.Background())
	if err != nil {
		panic(err)
	}

	for _, item := range applied {
		fmt.Println(item.Filename, item.StatementIndex, item.Status)
	}

	// Output:
	// 0001-user.up.sql 1 ok
	// 0002-comment.up.sql 0 ok
}
```

</section>

<section name="ExampleManager_Apply_failure">

### ExampleManager_Apply_failure

ExampleManager_Apply_failure shows what a failed run reports. The second
statement of the file fails, so Apply returns its error, and the entry of the
file holds the index of the statement before it - where the next Apply
resumes.

```go
func ExampleManager_Apply_failure() {
	fsys := fstest.MapFS{
		"0001-user.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE user (id INTEGER);\nCREATE TABLE user (id INTEGER);\n",
		)},
	}

	manager, err := migrate.NewManager(exampleDB(), fsys, "example")
	if err != nil {
		panic(err)
	}

	applied, err := manager.Apply(context.Background())
	if err == nil {
		panic("the second statement creates the same table twice")
	}

	failed := applied[0]
	fmt.Println("stopped in", failed.Filename, "after statement", failed.StatementIndex)
	fmt.Println("the entry carries the error:", errors.Is(failed.Err(), err))

	// Output:
	// stopped in 0001-user.up.sql after statement 0
	// the entry carries the error: true
}
```

</section>

<section name="ExampleManager_Apply_noMigrations">

### ExampleManager_Apply_noMigrations

ExampleManager_Apply_noMigrations shows the one failure that otherwise looks
like a clean run: a filesystem the pattern matches nothing in.

```go
func ExampleManager_Apply_noMigrations() {
	manager, err := migrate.NewManager(exampleDB(), fstest.MapFS{}, "example")
	if err != nil {
		panic(err)
	}

	_, err = manager.Apply(context.Background())

	fmt.Println(errors.Is(err, migrate.ErrNoMigrations))

	// Output:
	// true
}
```

</section>

<section name="ExampleManager_Load">

### ExampleManager_Load

ExampleManager_Load selects the migrations of one directory. The pattern
names a directory, so it is matched against the whole path, and the migration
sitting elsewhere is left out of the run.

```go
func ExampleManager_Load() {
	fsys := fstest.MapFS{
		"schema/0001-user.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE user (id INTEGER);\n",
		)},
		"vendor/0001-other.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE other (id INTEGER);\n",
		)},
	}

	manager, err := migrate.NewManager(exampleDB(), fsys, "example")
	if err != nil {
		panic(err)
	}
	manager.Load("schema/*.up.sql")

	applied, err := manager.Apply(context.Background())
	if err != nil {
		panic(err)
	}

	for _, item := range applied {
		fmt.Println(item.Filename)
	}

	// Output:
	// schema/0001-user.up.sql
}
```

</section>

<section name="ExampleManager_Load_anyDepth">

### ExampleManager_Load_anyDepth

ExampleManager_Load_anyDepth selects migrations wherever they sit. The pattern
names no directory, so it is matched against base names, which is what the
default Pattern does for an embed.FS handed over whole.

```go
func ExampleManager_Load_anyDepth() {
	fsys := fstest.MapFS{
		"schema/0001-user.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE user (id INTEGER);\n",
		)},
		"vendor/0001-other.up.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE other (id INTEGER);\n",
		)},
	}

	manager, err := migrate.NewManager(exampleDB(), fsys, "example")
	if err != nil {
		panic(err)
	}
	manager.Load("*.up.sql")

	applied, err := manager.Apply(context.Background())
	if err != nil {
		panic(err)
	}

	for _, item := range applied {
		fmt.Println(item.Filename)
	}

	// Output:
	// schema/0001-user.up.sql
	// vendor/0001-other.up.sql
}
```

</section>

