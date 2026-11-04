// Package migrate applies SQL migrations to a database and records what it
// applied.
//
// A migration is a file of semicolon-separated SQL statements. Files are read
// from an fs.FS, selected by the glob "*.up.sql", and applied in lexical order.
// Each file is applied in one transaction holding an advisory lock, and the
// outcome is written to the migrations table: the project, the filename, the
// index of the last statement that ran, and either "ok" or the error of the
// statement that failed.
//
// Statements already recorded as applied are skipped, so a file that grew
// statements since the last run has only the new ones applied.
//
//	m, err := migrate.NewManager(handle, os.DirFS("schema"), "events")
//	if err != nil {
//		return err
//	}
//	applied, err := m.Apply(ctx)
//	for _, item := range applied {
//		fmt.Println(item.Filename, item.StatementIndex, item.Status)
//	}
//
// Listing does not apply anything:
//
//	list, err := m.List(ctx)
//
// Printing the migrations of a project needs no database at all:
//
//	err := migrate.Print(os.Stdout, os.DirFS("schema"))
//
// # Migrations beside the code that owns them
//
// A project can keep each migration with the code owning its table rather than
// in one folder. Collect gathers them into the filesystem a Manager takes, over
// a tree carrying access/schema/, keystore/schema/ and modules/notes/schema/:
//
//	fsys, err := migrate.Collect(".", "schema", "")
//
// A nested name is matched as a path suffix, so "db/schema" finds one/db/schema
// and not one/schema. A dot directory is not walked, and one the filesystem
// refuses to read is skipped rather than failing the walk.
//
// The answer is flat, each file at its base name, and that is the point rather
// than a convenience. A run records the name of a file relative to the root of
// the filesystem it came from, so a file that gains a directory is a file
// nobody has run: it applies again, and the second ALTER TABLE of a column that
// is already there fails. Flattening is what lets a migration move between
// directories against a database that already holds it. It also fixes the
// order, which the walk does not: paths sort by their whole string, so
// a/z.up.sql would come before b/a.up.sql, while a migration reading a table
// another one created wants the order the names give.
//
// A base name therefore belongs to one directory. Two carrying the same one is
// ErrDuplicate, naming the file, rather than one shadowing the other with
// nothing to say the loser's table was never created. A pattern holding a "/"
// is ErrRootedPattern for the same reason: the files are matched where they sit
// and answered flat, so one that reads a path could match on the way in and
// never on the way out.
package migrate
