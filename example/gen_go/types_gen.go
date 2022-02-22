package stats

import (
	"time"
)

// Incoming generated for db table `incoming`
//
// Incoming stats log, writes only
type Incoming struct {
	// Tracking ID
	ID uint64 `db:"id" json:"-"`

	// Property name (human readable, a-z)
	Property string `db:"property" json:"-"`

	// Property Section ID
	PropertySection uint32 `db:"property_section" json:"-"`

	// Property Item ID
	PropertyID uint32 `db:"property_id" json:"-"`

	// Remote IP from user making request
	RemoteIP string `db:"remote_ip" json:"-"`

	// Timestamp of request
	Stamp *time.Time `db:"stamp" json:"-"`
}

// SetStamp sets Stamp which requires a *time.Time
func (i *Incoming) SetStamp(stamp time.Time) { i.Stamp = &stamp }

// IncomingTable is the name of the table in the DB
const IncomingTable = "`incoming`"

// IncomingFields are all the field names in the DB table
var IncomingFields = []string{"id", "property", "property_section", "property_id", "remote_ip", "stamp"}

// IncomingPrimaryFields are the primary key fields in the DB table
var IncomingPrimaryFields = []string{"id"}

// IncomingProc generated for db table `incoming_proc`
//
// Incoming stats log, writes only
type IncomingProc struct {
	// Tracking ID
	ID uint64 `db:"id" json:"-"`

	// Property name (human readable, a-z)
	Property string `db:"property" json:"-"`

	// Property Section ID
	PropertySection uint32 `db:"property_section" json:"-"`

	// Property Item ID
	PropertyID uint32 `db:"property_id" json:"-"`

	// Remote IP from user making request
	RemoteIP string `db:"remote_ip" json:"-"`

	// Timestamp of request
	Stamp *time.Time `db:"stamp" json:"-"`
}

// SetStamp sets Stamp which requires a *time.Time
func (i *IncomingProc) SetStamp(stamp time.Time) { i.Stamp = &stamp }

// IncomingProcTable is the name of the table in the DB
const IncomingProcTable = "`incoming_proc`"

// IncomingProcFields are all the field names in the DB table
var IncomingProcFields = []string{"id", "property", "property_section", "property_id", "remote_ip", "stamp"}

// IncomingProcPrimaryFields are the primary key fields in the DB table
var IncomingProcPrimaryFields = []string{"id"}

// Migrations generated for db table `migrations`
//
// Migration log of applied migrations
type Migrations struct {
	// Microservice or project name
	Project string `db:"project" json:"-"`

	// yyyy-mm-dd-HHMMSS.sql
	Filename string `db:"filename" json:"-"`

	// Statement number from SQL file
	StatementIndex int32 `db:"statement_index" json:"-"`

	// ok or full error message
	Status string `db:"status" json:"-"`
}

// MigrationsTable is the name of the table in the DB
const MigrationsTable = "`migrations`"

// MigrationsFields are all the field names in the DB table
var MigrationsFields = []string{"project", "filename", "statement_index", "status"}

// MigrationsPrimaryFields are the primary key fields in the DB table
var MigrationsPrimaryFields = []string{"project", "filename"}
