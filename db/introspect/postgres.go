package introspect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/go-bridget/mig/model"
)

// PostgresDescriber implements Describer for PostgreSQL
type PostgresDescriber struct{}

// Describe returns column metadata from a query.
func (d *PostgresDescriber) Describe(ctx context.Context, db *sqlx.DB, query string) ([]*model.Column, error) {
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

// DescribeTable returns the structure of a table.
func (d *PostgresDescriber) DescribeTable(ctx context.Context, db *sqlx.DB, tableName string) (*model.Table, error) {
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
			c.column_name as "COLUMN_NAME",
			c.udt_name as "COLUMN_TYPE",
			COALESCE(col_description(cl.oid, c.ordinal_position), '') as "COLUMN_COMMENT",
			c.udt_name as "DATA_TYPE",
			CASE WHEN pk.conname IS NOT NULL AND a.attnum = ANY(pk.conkey) THEN 'PRI' ELSE '' END as "COLUMN_KEY"
		FROM information_schema.columns c
		LEFT JOIN pg_class cl ON c.table_name = cl.relname AND c.table_schema = cl.relnamespace::regnamespace::name
		LEFT JOIN pg_attribute a ON cl.oid = a.attrelid AND c.column_name = a.attname
		LEFT JOIN pg_constraint pk ON cl.oid = pk.conrelid AND pk.contype = 'p'
		WHERE c.table_name = $1 AND c.table_schema = current_schema()
		ORDER BY c.ordinal_position
	`

	if err := db.SelectContext(ctx, &columns, query, tableName); err != nil {
		return nil, errors.Wrapf(err, "failed to get columns for table %s", tableName)
	}

	// Enrich columns with normalized type and extract ENUM values
	for _, col := range columns {
		// Parse PostgreSQL type aliases (int2, int4, int8) to extract size
		// Keep the original type (e.g., int8, int4) but extract the size in bytes
		_, sizeBytes := ParsePostgresIntType(col.Type)
		if sizeBytes > 0 {
			col.Size = sizeBytes
		}

		// Try to extract ENUM values for all columns (custom types and enum types)
		// extractPostgresEnumValues will return nil if the type is not an enum
		enumVals := extractPostgresEnumValues(ctx, db, col.Type)
		if enumVals != nil && len(enumVals) > 0 {
			col.Values = enumVals
			col.EnumValues = enumVals // Keep for backward compatibility
		}
		// Normalize the type
		NormalizeColumnType(col, "postgres")
	}

	// Get indexes for this table
	indexes, err := d.TableIndexes(ctx, db, tableName)
	if err != nil {
		return nil, err
	}

	// Enrich key metadata based on naming conventions and indexes
	EnrichKeyMetadata(columns, indexes)

	table.Columns = columns
	table.Indexes = indexes
	return table, nil
}

// ListTables returns all tables without columns populated.
func (d *PostgresDescriber) ListTables(ctx context.Context, db *sqlx.DB) ([]*model.Table, error) {
	tables := []*model.Table{}

	// Get all tables in current schema (excluding system tables)
	if err := db.SelectContext(ctx, &tables, `
		SELECT 
			c.relname as "TABLE_NAME",
			COALESCE(d.description, '') as "TABLE_COMMENT"
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

// TableIndexes returns all indexes for a table.
func (d *PostgresDescriber) TableIndexes(ctx context.Context, db *sqlx.DB, tableName string) ([]*model.Index, error) {
	type indexInfo struct {
		Name    string         `db:"name"`
		Columns pq.StringArray `db:"columns"`
		Primary bool           `db:"primary"`
		Unique  bool           `db:"unique"`
	}

	var indexInfos []indexInfo
	query := `
		SELECT 
			i.relname as name,
			ARRAY_AGG(a.attname ORDER BY a.attnum) as columns,
			ix.indisprimary as primary,
			ix.indisunique as unique
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE t.relname = $1 AND t.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = current_schema())
		GROUP BY i.relname, ix.indisprimary, ix.indisunique
		ORDER BY i.relname
	`

	if err := db.SelectContext(ctx, &indexInfos, query, tableName); err != nil {
		return nil, errors.Wrapf(err, "failed to get indexes for table %s", tableName)
	}

	var indexes []*model.Index
	for _, info := range indexInfos {
		indexes = append(indexes, &model.Index{
			Name:    info.Name,
			Columns: []string(info.Columns),
			Primary: info.Primary,
			Unique:  info.Unique,
		})
	}

	return indexes, nil
}

// extractPostgresEnumValues fetches the allowed values for an ENUM type.
func extractPostgresEnumValues(ctx context.Context, db *sqlx.DB, enumTypeName string) []string {
	var values []string

	// Query pg_enum to get all values for this enum type
	query := `
		SELECT enumlabel FROM pg_enum
		WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = $1)
		ORDER BY enumsortorder
	`

	if err := db.SelectContext(ctx, &values, query, enumTypeName); err != nil {
		// Silently ignore if we can't fetch enum values
		return nil
	}

	return values
}
