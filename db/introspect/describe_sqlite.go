package introspect

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/go-bridget/mig/model"
)

// sqliteDescriber implements Describer for SQLite
type sqliteDescriber struct{}

// Describe returns column metadata for a SQLite query by creating a temporary view
func (d *sqliteDescriber) Describe(ctx context.Context, db *sqlx.DB, query string) ([]*model.Column, error) {
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

	// Use PRAGMA table_info to get column details
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", viewName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query column metadata with PRAGMA table_info")
	}
	defer rows.Close()

	columns := []*model.Column{}
	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notnull int
		var dfltValue *string
		var pk int

		if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			return nil, errors.Wrap(err, "failed to scan column metadata")
		}

		column := &model.Column{
			Name:     name,
			Type:     colType,
			DataType: strings.ToLower(colType),
		}

		// Map SQLite types to standard types (consistent with existing logic in list.go)
		var sqliteTypeMapping = map[string]string{
			"integer": "bigint",
			"real":    "double",
			"text":    "varchar",
		}
		if mapped, ok := sqliteTypeMapping[column.DataType]; ok {
			column.DataType = mapped
		}

		// Set primary key indicator
		if pk == 1 {
			column.Key = "PRI"
		} else {
			column.Key = ""
		}

		// Generate default comment from column name if not provided
		comment := model.Title(column.Name)
		commentRune := []rune(comment)
		if len(commentRune) > 0 {
			commentRune[0] = unicode.ToUpper(commentRune[0])
		}
		column.Comment = string(commentRune)

		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error reading column metadata")
	}

	return columns, nil
}

// DescribeTable returns the structure of a specific table from the database schema
func (d *sqliteDescriber) DescribeTable(ctx context.Context, db *sqlx.DB, tableName string) (*model.Table, error) {
	table := &model.Table{
		Name:    tableName,
		Comment: "", // SQLite doesn't have table comments
	}

	// Get columns using PRAGMA table_info
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query table info for %s", tableName)
	}
	defer rows.Close()

	columns := []*model.Column{}
	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notnull int
		var dfltValue *string
		var pk int

		if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			return nil, errors.Wrap(err, "failed to scan column metadata")
		}

		column := &model.Column{
			Name:     name,
			Type:     colType,
			DataType: strings.ToLower(colType),
		}

		// Map SQLite types to standard types
		var sqliteTypeMapping = map[string]string{
			"integer": "bigint",
			"real":    "double",
			"text":    "varchar",
		}
		if mapped, ok := sqliteTypeMapping[column.DataType]; ok {
			column.DataType = mapped
		}

		// Set primary key indicator
		if pk == 1 {
			column.Key = "PRI"
		} else {
			column.Key = ""
		}

		// Generate comment from column name
		comment := model.Title(column.Name)
		commentRune := []rune(comment)
		if len(commentRune) > 0 {
			commentRune[0] = unicode.ToUpper(commentRune[0])
		}
		column.Comment = string(commentRune)

		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error reading column metadata")
	}

	table.Columns = columns
	return table, nil
}

// ListTables returns all tables in the database (excluding system tables)
func (d *sqliteDescriber) ListTables(ctx context.Context, db *sqlx.DB) ([]*model.Table, error) {
	tables := []*model.Table{}

	// Get all user-defined tables (excluding sqlite internal tables)
	if err := db.SelectContext(ctx, &tables, "SELECT name as TABLE_NAME, '' as TABLE_COMMENT FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name ASC"); err != nil {
		return nil, errors.Wrap(err, "failed to list tables")
	}

	return tables, nil
}
