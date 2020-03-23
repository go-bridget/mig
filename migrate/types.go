package migrate

type (
	// Options include migration options
	Options struct {
		Path    string
		Project string

		Apply   bool
		Verbose bool
	}

	// Migration holds the DB structure for the migration table
	Migration struct {
		Project        string `db:"project"`
		Filename       string `db:"filename"`
		StatementIndex int    `db:"statement_index"`
		Status         string `db:"status"`
	}
)

// MigrationFields hold the database column names for Migration{}
var MigrationFields = []string{"project", "filename", "statement_index", "status"}

// migrations holds loaded migrations
var migrations map[string]FS = map[string]FS{}
var migrationsFile = "migrations.sql"
