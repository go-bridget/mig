package stats

// Migrations generated for db table `migrations`
type Migrations struct {
	// Project
	Project string `db:"project" json:"-"`

	// Filename
	Filename string `db:"filename" json:"-"`

	// Statement index
	StatementIndex int64 `db:"statement_index" json:"-"`

	// Status
	Status string `db:"status" json:"-"`
}

// MigrationsTable is the name of the table in the DB
const MigrationsTable = "`migrations`"

// MigrationsFields are all the field names in the DB table
var MigrationsFields = []string{"project", "filename", "statement_index", "status"}

// MigrationsPrimaryFields are the primary key fields in the DB table
var MigrationsPrimaryFields = []string{"project"}

// Stats generated for db table `stats`
type Stats struct {
	// Id
	ID int64 `db:"id" json:"-"`

	// Payload
	Payload string `db:"payload" json:"-"`
}

// StatsTable is the name of the table in the DB
const StatsTable = "`stats`"

// StatsFields are all the field names in the DB table
var StatsFields = []string{"id", "payload"}

// StatsPrimaryFields are the primary key fields in the DB table
var StatsPrimaryFields = []string{"id"}
