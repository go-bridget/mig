package introspect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/go-bridget/mig/model"
)

// postgresDescriber implements Describer for PostgreSQL
type postgresDescriber struct{}

// Describe returns column metadata for a PostgreSQL query by creating a temporary view
func (d *postgresDescriber) Describe(ctx context.Context, db *sqlx.DB, query string) ([]*model.Column, error) {
	var err error

	// Normalize query
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	// Generate unique temporary view name
	viewName := fmt.Sprintf("mig_temp_view_%d", time.Now().UnixNano())

	// Create temporary view
	createViewSQL := fmt.Sprintf("CREATE TEMPORARY VIEW %s AS %s", viewName, query)
	if _, err = db.ExecContext(ctx, createViewSQL); err != nil {
		return nil, errors.Wrapf(err, "failed to create temporary view for query")
	}

	// Defer cleanup - drop the temporary view
	defer func() {
		dropSQL := fmt.Sprintf("DROP VIEW IF EXISTS %s", viewName)
		_, _ = db.ExecContext(context.Background(), dropSQL)
	}()

	// Query information_schema to get column details

	// PostgreSQL information_schema query
	// Note: COLUMN_KEY is not directly equivalent in PostgreSQL, so we check constraints
	pgQuery := `
		SELECT
			c.column_name as COLUMN_NAME,
			c.udt_name as COLUMN_TYPE,
			COALESCE(tc.constraint_type, '') as COLUMN_KEY,
			COALESCE(c.column_default, '') as DATA_TYPE
		FROM information_schema.columns c
		LEFT JOIN information_schema.constraint_column_usage ccu
			ON c.table_catalog = ccu.table_catalog
			AND c.table_schema = ccu.table_schema
			AND c.table_name = ccu.table_name
			AND c.column_name = ccu.column_name
		LEFT JOIN information_schema.table_constraints tc
			ON ccu.constraint_catalog = tc.constraint_catalog
			AND ccu.constraint_schema = tc.constraint_schema
			AND ccu.constraint_name = tc.constraint_name
		WHERE c.table_name = $1
		AND c.table_schema = 'pg_temp'
		ORDER BY c.ordinal_position ASC
	`

	// Map PostgreSQL constraint types to match common KEY values
	type pgColumn struct {
		ColumnName     string `db:"COLUMN_NAME"`
		ColumnType     string `db:"COLUMN_TYPE"`
		ConstraintType string `db:"COLUMN_KEY"`
		ColumnDefault  string `db:"DATA_TYPE"`
	}

	pgCols := []pgColumn{}
	columns := []*model.Column{}
	if err = db.SelectContext(ctx, &pgCols, pgQuery, viewName); err != nil {
		return nil, errors.Wrap(err, "failed to query column metadata from information_schema")
	}

	for _, pgCol := range pgCols {
		column := &model.Column{
			Name:     pgCol.ColumnName,
			Type:     pgCol.ColumnType,
			DataType: pgCol.ColumnType,
		}

		// Map PostgreSQL constraint types
		if pgCol.ConstraintType == "PRIMARY KEY" {
			column.Key = "PRI"
		}

		// Generate default comment from column name
		column.Comment = model.Title(column.Name)

		columns = append(columns, column)
	}

	return columns, nil
}

// DescribeTable returns the structure of a specific table from the database schema
func (d *postgresDescriber) DescribeTable(ctx context.Context, db *sqlx.DB, tableName string) (*model.Table, error) {
	table := &model.Table{
		Name: tableName,
	}

	// Get table comment from pg_description
	var comment *string
	if err := db.GetContext(ctx, &comment, `
		SELECT description FROM pg_description 
		WHERE objoid = (SELECT oid FROM pg_class WHERE relname = $1)
		AND objsubid = 0
	`, tableName); err != nil {
		// Silently ignore if no comment found
		comment = nil
	}
	if comment != nil {
		table.Comment = *comment
	}

	// Get columns from information_schema
	columns := []*model.Column{}
	query := `
		SELECT 
			column_name as COLUMN_NAME,
			udt_name as COLUMN_TYPE,
			COALESCE(col_description(attrelid, attnum), '') as COLUMN_COMMENT,
			udt_name as DATA_TYPE,
			CASE WHEN ix.indexname LIKE 'idx_%' THEN '' ELSE '' END as COLUMN_KEY
		FROM information_schema.columns c
		LEFT JOIN pg_attribute ON c.table_name = pg_class.relname AND c.column_name = pg_attribute.attname
		LEFT JOIN pg_class ON c.table_catalog = current_database() AND c.table_name = pg_class.relname
		LEFT JOIN (SELECT tablename, indexname FROM pg_indexes) ix ON c.table_name = ix.tablename
		WHERE c.table_name = $1 AND c.table_schema = current_schema()
		ORDER BY c.ordinal_position
	`

	if err := db.SelectContext(ctx, &columns, query, tableName); err != nil {
		return nil, errors.Wrapf(err, "failed to get columns for table %s", tableName)
	}

	table.Columns = columns
	return table, nil
}

// ListTables returns all tables in the current schema with their columns
func (d *postgresDescriber) ListTables(ctx context.Context, db *sqlx.DB, schema string) ([]*model.Table, error) {
	tables := []*model.Table{}

	// Get all tables in current schema (excluding system tables)
	if err := db.SelectContext(ctx, &tables, `
		SELECT 
			c.relname as NAME,
			COALESCE(d.description, '') as COMMENT
		FROM pg_class c
		LEFT JOIN pg_description d ON c.oid = d.objoid AND d.objsubid = 0
		WHERE c.relkind = 'r'
		AND c.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = current_schema())
		ORDER BY c.relname
	`); err != nil {
		return nil, errors.Wrap(err, "failed to list tables")
	}

	return tables, nil
}
