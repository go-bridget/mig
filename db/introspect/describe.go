package introspect

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/go-bridget/mig/model"
)

// Describer defines the interface for database-specific schema introspection operations.
// It can describe tables, queries, and list available tables in the database.
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
}

// ListTablesWithColumns returns all tables with their columns populated.
// For each table returned by ListTables, it calls DescribeTable to fetch column information.
// If a table comment is empty, it's filled with a title-cased version of the table name.
func ListTablesWithColumns(ctx context.Context, db *sqlx.DB, describer Describer) ([]*model.Table, error) {
	// Get list of tables without columns
	tables, err := describer.ListTables(ctx, db)
	if err != nil {
		return nil, err
	}

	// Populate columns for each table
	for _, table := range tables {
		fullTable, err := describer.DescribeTable(ctx, db, table.Name)
		if err != nil {
			return nil, err
		}
		table.Columns = fullTable.Columns

		// Fill in comment if empty
		if table.Comment == "" {
			table.Comment = model.Title(table.Name)
		}
	}

	return tables, nil
}

// NewDescriber returns a Describer implementation for the given database driver name
func NewDescriber(driverName string) (Describer, error) {
	switch driverName {
	case "sqlite":
		return &sqliteDescriber{}, nil
	case "postgres", "postgresql":
		return &postgresDescriber{}, nil
	case "mysql":
		return &mysqlDescriber{}, nil
	}
	return nil, errors.New("Unknown describer: " + driverName)
}
