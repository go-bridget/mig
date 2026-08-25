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
package migrate
