package stats

import (
	"time"
)

// Asset generated for db table `asset`
//
// Stores asset information for each commit
type Asset struct {
	// Asset ID
	ID int32 `db:"id" json:"-"`

	// Commit ID
	CommitID int32 `db:"commit_id" json:"-"`

	// Filename
	Filename string `db:"filename" json:"-"`

	// File contents
	Contents string `db:"contents" json:"-"`

	// Record creation timestamp
	CreatedAt *time.Time `db:"created_at" json:"-"`

	// Record update timestamp
	UpdatedAt *time.Time `db:"updated_at" json:"-"`
}

// SetCreatedAt sets CreatedAt which requires a *time.Time
func (a *Asset) SetCreatedAt(stamp time.Time) { a.CreatedAt = &stamp }

// SetUpdatedAt sets UpdatedAt which requires a *time.Time
func (a *Asset) SetUpdatedAt(stamp time.Time) { a.UpdatedAt = &stamp }

// AssetTable is the name of the table in the DB
const AssetTable = "`asset`"

// AssetFields are all the field names in the DB table
var AssetFields = []string{"id", "commit_id", "filename", "contents", "created_at", "updated_at"}

// AssetPrimaryFields are the primary key fields in the DB table
var AssetPrimaryFields = []string{"id"}

// Branch generated for db table `branch`
//
// Stores information about branches in repositories
type Branch struct {
	// Branch ID
	ID int32 `db:"id" json:"-"`

	// Repository ID
	RepositoryID int32 `db:"repository_id" json:"-"`

	// Branch name
	Name string `db:"name" json:"-"`

	// Record creation timestamp
	CreatedAt *time.Time `db:"created_at" json:"-"`

	// Record update timestamp
	UpdatedAt *time.Time `db:"updated_at" json:"-"`
}

// SetCreatedAt sets CreatedAt which requires a *time.Time
func (b *Branch) SetCreatedAt(stamp time.Time) { b.CreatedAt = &stamp }

// SetUpdatedAt sets UpdatedAt which requires a *time.Time
func (b *Branch) SetUpdatedAt(stamp time.Time) { b.UpdatedAt = &stamp }

// BranchTable is the name of the table in the DB
const BranchTable = "`branch`"

// BranchFields are all the field names in the DB table
var BranchFields = []string{"id", "repository_id", "name", "created_at", "updated_at"}

// BranchPrimaryFields are the primary key fields in the DB table
var BranchPrimaryFields = []string{"id"}

// Commit generated for db table `commit`
//
// Stores information about commits in branches
type Commit struct {
	// Commit ID
	ID int32 `db:"id" json:"-"`

	// Branch ID
	BranchID int32 `db:"branch_id" json:"-"`

	// Commit hash
	CommitHash string `db:"commit_hash" json:"-"`

	// Commit author
	Author string `db:"author" json:"-"`

	// Commit message
	Message string `db:"message" json:"-"`

	// Commit timestamp
	CommittedAt *time.Time `db:"committed_at" json:"-"`

	// Record creation timestamp
	CreatedAt *time.Time `db:"created_at" json:"-"`

	// Record update timestamp
	UpdatedAt *time.Time `db:"updated_at" json:"-"`
}

// SetCommittedAt sets CommittedAt which requires a *time.Time
func (c *Commit) SetCommittedAt(stamp time.Time) { c.CommittedAt = &stamp }

// SetCreatedAt sets CreatedAt which requires a *time.Time
func (c *Commit) SetCreatedAt(stamp time.Time) { c.CreatedAt = &stamp }

// SetUpdatedAt sets UpdatedAt which requires a *time.Time
func (c *Commit) SetUpdatedAt(stamp time.Time) { c.UpdatedAt = &stamp }

// CommitTable is the name of the table in the DB
const CommitTable = "`commit`"

// CommitFields are all the field names in the DB table
var CommitFields = []string{"id", "branch_id", "commit_hash", "author", "message", "committed_at", "created_at", "updated_at"}

// CommitPrimaryFields are the primary key fields in the DB table
var CommitPrimaryFields = []string{"id"}

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

// Repository generated for db table `repository`
//
// Stores basic information about repositories
type Repository struct {
	// Repository ID
	ID int32 `db:"id" json:"-"`

	// Repository name
	Name string `db:"name" json:"-"`

	// Repository URL
	URL string `db:"url" json:"-"`

	// Record creation timestamp
	CreatedAt *time.Time `db:"created_at" json:"-"`

	// Record update timestamp
	UpdatedAt *time.Time `db:"updated_at" json:"-"`
}

// SetCreatedAt sets CreatedAt which requires a *time.Time
func (r *Repository) SetCreatedAt(stamp time.Time) { r.CreatedAt = &stamp }

// SetUpdatedAt sets UpdatedAt which requires a *time.Time
func (r *Repository) SetUpdatedAt(stamp time.Time) { r.UpdatedAt = &stamp }

// RepositoryTable is the name of the table in the DB
const RepositoryTable = "`repository`"

// RepositoryFields are all the field names in the DB table
var RepositoryFields = []string{"id", "name", "url", "created_at", "updated_at"}

// RepositoryPrimaryFields are the primary key fields in the DB table
var RepositoryPrimaryFields = []string{"id"}
