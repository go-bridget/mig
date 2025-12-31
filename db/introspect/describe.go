package introspect

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/go-bridget/mig/model"
)

// Describer defines the interface for database-specific schema introspection operations.
// It can describe tables, queries, and list available tables in the database.
// The database connection provides both the query execution and driver information.
type Describer interface {
	// Describe returns column metadata for a given SQL query or table.
	// For tables, pass "SELECT * FROM table_name" or just "table_name".
	// For queries, pass any SELECT statement.
	// It works by creating a temporary view from the query, inspecting its columns,
	// and dropping the view afterwards.
	Describe(ctx context.Context, db *sqlx.DB, query string) ([]*model.Column, error)

	// DescribeTable returns the table structure including all columns and metadata.
	// This is more efficient than Describe() for tables as it queries the schema directly.
	DescribeTable(ctx context.Context, db *sqlx.DB, tableName string) (*model.Table, error)

	// ListTables returns all tables in the database (excluding system/temporary tables).
	// Note: Columns are not populated. Use DescribeTable to fetch columns for a specific table.
	ListTables(ctx context.Context, db *sqlx.DB) ([]*model.Table, error)

	// TableIndexes returns all indexes for a given table, including primary keys and unique constraints.
	TableIndexes(ctx context.Context, db *sqlx.DB, tableName string) ([]*model.Index, error)
}

// ListTablesWithColumns returns all tables with their columns populated and indexes sorted.
// For each table returned by ListTables, it calls DescribeTable to fetch column information.
// If a table comment is empty, it's filled with a title-cased version of the table name.
// Indexes are sorted consistently: primary key first, then by column names.
func ListTablesWithColumns(ctx context.Context, db *sqlx.DB, describer Describer) ([]*model.Table, error) {
	// Get list of tables without columns
	tables, err := describer.ListTables(ctx, db)
	if err != nil {
		return nil, err
	}

	// Populate columns and indexes for each table
	for _, table := range tables {
		fullTable, err := describer.DescribeTable(ctx, db, table.Name)
		if err != nil {
			return nil, err
		}
		table.Columns = fullTable.Columns
		table.Indexes = fullTable.Indexes

		// Fill in comment if empty
		if table.Comment == "" {
			table.Comment = model.Title(table.Name)
		}

		// Sort indexes for consistent output
		if len(table.Indexes) > 0 {
			sortIndexes(table.Indexes)
		}
	}

	return tables, nil
}

// NewDescriber returns a Describer implementation for the given database connection.
// The driver type is determined from the database connection's DriverName().
func NewDescriber(db *sqlx.DB) (Describer, error) {
	driverName := db.DriverName()
	switch driverName {
	case "sqlite":
		return &SqliteDescriber{}, nil
	case "postgres", "postgresql":
		return &PostgresDescriber{}, nil
	case "mysql":
		return &MysqlDescriber{}, nil
	}
	return nil, errors.New("Unknown describer: " + driverName)
}
