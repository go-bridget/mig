package migrate

import "errors"

// Migration holds the DB structure for the migration table.
type Migration struct {
	// Project holds a migration scope. You may have several projects
	// migrated within the same migration table.
	Project string `db:"project"`

	// Filename logs the file used for storing migrations.
	Filename string `db:"filename"`

	// StatementIndex is the current index of applied migrations. It is -1
	// for a migration the table has no record of.
	StatementIndex int `db:"statement_index"`

	// Status contains the status of the migrations. It's expected to be
	// 'ok' for a healthy value, and holds the error of the statement that
	// failed otherwise. It is empty for a migration the table has no record
	// of.
	Status string `db:"status"`

	// err is the error behind Status, kept so a caller can unwrap what a
	// run returned. It is not a column: a migration read back from the
	// table has only the text in Status to go on.
	err error
}

// SetError records err as the outcome of the migration. Status becomes the
// error text, which is what the column holds, or StatusOK when err is nil.
//
// The error itself is kept as well, so Err returns what was passed rather than
// a copy of its message.
func (m *Migration) SetError(err error) {
	m.err = err

	if err == nil {
		m.Status = StatusOK
		return
	}
	m.Status = err.Error()
}

// Err returns the error the migration failed with, or nil when it succeeded or
// has not run.
//
// For a migration read back from the migrations table there is no error left to
// return, only the text of one, so Err reports the Status as an error.
func (m Migration) Err() error {
	if m.err != nil {
		return m.err
	}
	if m.Status == "" || m.Status == StatusOK {
		return nil
	}
	return errors.New(m.Status)
}

// MigrationFields hold the database column names for Migration{}.
var MigrationFields = []string{"project", "filename", "statement_index", "status"}
