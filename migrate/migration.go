package migrate

type (
	// Migration holds the DB structure for the migration table.
	Migration struct {
		// Project holds a migration scope. You may have several
		// projects migrated within the same migration table.
		Project string `db:"project"`

		// Filename logs the file used for storing migrations.
		Filename string `db:"filename"`

		// StatementIndex is the current index of applied migrations.
		StatementIndex int `db:"statement_index"`

		// Status contains the status of the migrations.
		// It's expected to be 'ok' for a healthy value.
		Status string `db:"status"`
	}
)

// MigrationFields hold the database column names for Migration{}.
var MigrationFields = []string{"project", "filename", "statement_index", "status"}

// migrations holds loaded migrations
var migrations map[string]FS = map[string]FS{}
