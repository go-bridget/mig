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
type SqliteDescriber struct{}

// Describe returns column metadata from a query.
func (d *SqliteDescriber) Describe(ctx context.Context, db *sqlx.DB, query string) ([]*model.Column, error) {
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
	type pragmaColumn struct {
		Cid     int     `db:"cid"`
		Name    string  `db:"name"`
		Type    string  `db:"type"`
		Notnull int     `db:"notnull"`
		Dflt    *string `db:"dflt_value"`
		Pk      int     `db:"pk"`
	}

	var pragmaColumns []pragmaColumn
	if err := db.SelectContext(ctx, &pragmaColumns, fmt.Sprintf("PRAGMA table_info(%s)", viewName)); err != nil {
		return nil, errors.Wrap(err, "failed to query column metadata with PRAGMA table_info")
	}

	columns := []*model.Column{}
	for _, pc := range pragmaColumns {
		column := &model.Column{
			Name:     pc.Name,
			Type:     pc.Type,
			DataType: strings.ToLower(pc.Type),
		}

		// Map SQLite types to standard types (consistent with existing logic in list.go)
		var sqliteTypeMapping = map[string]string{
			"integer": "bigint",
			"real":    "double",
			"text":    "varchar",
			"blob":    "blob",
		}
		if mapped, ok := sqliteTypeMapping[column.DataType]; ok {
			column.DataType = mapped
		}

		// Set primary key indicator - pk > 0 indicates part of primary key
		if pc.Pk > 0 {
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

	return columns, nil
}

// DescribeTable returns the structure of a table.
func (d *SqliteDescriber) DescribeTable(ctx context.Context, db *sqlx.DB, tableName string) (*model.Table, error) {
	table := &model.Table{
		Name:    tableName,
		Comment: "", // SQLite doesn't have table comments
	}

	// Get columns using PRAGMA table_info
	type pragmaColumn struct {
		Cid     int     `db:"cid"`
		Name    string  `db:"name"`
		Type    string  `db:"type"`
		Notnull int     `db:"notnull"`
		Dflt    *string `db:"dflt_value"`
		Pk      int     `db:"pk"`
	}

	var pragmaColumns []pragmaColumn
	if err := db.SelectContext(ctx, &pragmaColumns, fmt.Sprintf("PRAGMA table_info(%s)", tableName)); err != nil {
		return nil, errors.Wrapf(err, "failed to query table info for %s", tableName)
	}

	columns := []*model.Column{}
	for _, pc := range pragmaColumns {
		column := &model.Column{
			Name:     pc.Name,
			Type:     pc.Type,
			DataType: strings.ToLower(pc.Type),
		}

		// Map SQLite types to standard types
		var sqliteTypeMapping = map[string]string{
			"integer": "bigint",
			"real":    "double",
			"text":    "text",
			"blob":    "blob",
		}
		if mapped, ok := sqliteTypeMapping[column.DataType]; ok {
			column.DataType = mapped
		}

		// Set primary key indicator - pk > 0 indicates part of primary key
		if pc.Pk > 0 {
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

	// Extract ENUM values from CHECK constraints and normalize types
	checkConstraints := extractSqliteCheckConstraints(ctx, db, tableName)
	for _, col := range columns {
		// Check if this column has enum-like CHECK constraint
		if constraints, ok := checkConstraints[col.Name]; ok && len(constraints) > 0 {
			col.Values = extractEnumValuesFromCheckConstraint(constraints[0])
			col.EnumValues = col.Values // Keep for backward compatibility
		}
		// Normalize the type
		NormalizeColumnType(col, "sqlite3")
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
func (d *SqliteDescriber) ListTables(ctx context.Context, db *sqlx.DB) ([]*model.Table, error) {
	tables := []*model.Table{}

	// Get all user-defined tables (excluding sqlite internal tables)
	if err := db.SelectContext(ctx, &tables, "SELECT name as TABLE_NAME, '' as TABLE_COMMENT FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name ASC"); err != nil {
		return nil, errors.Wrap(err, "failed to list tables")
	}

	return tables, nil
}

// TableIndexes returns all indexes for a table, synthesizing the primary key from table_info.
func (d *SqliteDescriber) TableIndexes(ctx context.Context, db *sqlx.DB, tableName string) ([]*model.Index, error) {
	// Get all indexes for the table (PRAGMA doesn't support parameterized queries)
	type indexInfo struct {
		Seq     int    `db:"seq"`
		Name    string `db:"name"`
		Unique  int    `db:"unique"`
		Origin  string `db:"origin"`
		Partial int    `db:"partial"`
	}

	var indexInfos []indexInfo
	if err := db.SelectContext(ctx, &indexInfos, fmt.Sprintf("PRAGMA index_list(%s)", tableName)); err != nil {
		return nil, errors.Wrapf(err, "failed to get indexes for table %s", tableName)
	}

	var indexes []*model.Index
	for _, ii := range indexInfos {
		// Get columns for this index
		type indexColumn struct {
			Seqno int    `db:"seqno"`
			Cid   int    `db:"cid"`
			Name  string `db:"name"`
		}

		var indexColumns []indexColumn
		if err := db.SelectContext(ctx, &indexColumns, fmt.Sprintf("PRAGMA index_info(%s)", ii.Name)); err != nil {
			continue
		}

		var columns []string
		for _, ic := range indexColumns {
			columns = append(columns, ic.Name)
		}

		indexes = append(indexes, &model.Index{
			Name:    ii.Name,
			Columns: columns,
			Primary: ii.Origin == "pk",
			Unique:  ii.Unique == 1,
		})
	}

	// SQLite doesn't expose the implicit PRIMARY KEY index via PRAGMA index_list
	// Synthesize it from table_info to match MySQL and PostgreSQL behavior
	type pragmaColumn struct {
		Cid     int     `db:"cid"`
		Name    string  `db:"name"`
		Type    string  `db:"type"`
		Notnull int     `db:"notnull"`
		Dflt    *string `db:"dflt_value"`
		Pk      int     `db:"pk"`
	}

	var pragmaColumns []pragmaColumn
	if err := db.SelectContext(ctx, &pragmaColumns, fmt.Sprintf("PRAGMA table_info(%s)", tableName)); err != nil {
		return indexes, nil // Return what we have if we can't get primary key info
	}

	// Collect primary key columns in order
	var pkColumns []string
	for _, pc := range pragmaColumns {
		if pc.Pk > 0 {
			pkColumns = append(pkColumns, pc.Name)
		}
	}

	// Add synthetic PRIMARY KEY index if table has primary key columns and no explicit pk index
	if len(pkColumns) > 0 {
		// Check if we already have a PRIMARY KEY index
		hasPKIndex := false
		for _, idx := range indexes {
			if idx.Primary {
				hasPKIndex = true
				break
			}
		}

		// Add synthetic PRIMARY KEY index if it doesn't exist
		if !hasPKIndex {
			indexes = append(indexes, &model.Index{
				Columns: pkColumns,
				Primary: true,
				Unique:  true,
			})
		}
	}

	return indexes, nil
}

// extractSqliteCheckConstraints extracts CHECK constraints from table definition
// Returns a map of column name to list of CHECK constraint expressions
func extractSqliteCheckConstraints(ctx context.Context, db *sqlx.DB, tableName string) map[string][]string {
	constraintMap := make(map[string][]string)

	// Get table creation SQL from sqlite_schema
	var sql string
	err := db.GetContext(ctx, &sql, "SELECT sql FROM sqlite_schema WHERE type='table' AND name=?", tableName)
	if err != nil || sql == "" {
		return constraintMap
	}

	// Parse CHECK constraints from the CREATE TABLE statement
	// Look for patterns like: column_name TYPE CHECK (column_name IN ('val1','val2'))
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToUpper(line), "CHECK") {
			// Extract column name and constraint
			parts := strings.Split(line, "CHECK")
			if len(parts) == 2 {
				colName := extractColumnNameFromCheckLine(parts[0])
				constraint := strings.TrimSuffix(strings.TrimSpace(parts[1]), ",")
				if colName != "" {
					constraintMap[colName] = append(constraintMap[colName], constraint)
				}
			}
		}
	}

	return constraintMap
}

// extractColumnNameFromCheckLine extracts column name from a CREATE TABLE line before CHECK
func extractColumnNameFromCheckLine(line string) string {
	// Get the first token in the line (should be the column name)
	// Format is typically: column_name TYPE ... CHECK (...)
	fields := strings.Fields(line)
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}

// extractEnumValuesFromCheckConstraint extracts values from IN ('val1','val2',...) pattern
func extractEnumValuesFromCheckConstraint(constraint string) []string {
	// Look for IN ('value1','value2',...)
	startIdx := strings.Index(strings.ToUpper(constraint), "IN")
	if startIdx == -1 {
		return nil
	}

	// Find the opening paren after IN
	parenIdx := strings.Index(constraint[startIdx:], "(")
	if parenIdx == -1 {
		return nil
	}

	// Extract the part between the IN (...) parens
	startPos := startIdx + parenIdx + 1

	// Find the matching closing paren for IN (...)
	parenCount := 1
	endPos := startPos
	for endPos < len(constraint) && parenCount > 0 {
		if constraint[endPos] == '(' {
			parenCount++
		} else if constraint[endPos] == ')' {
			parenCount--
		}
		endPos++
	}

	if parenCount != 0 || endPos <= startPos+1 {
		return nil
	}

	inClause := constraint[startPos : endPos-1] // -1 to exclude the closing paren we found

	// Split by comma and extract quoted values
	var values []string
	parts := strings.Split(inClause, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Remove quotes
		if strings.HasPrefix(part, "'") && strings.HasSuffix(part, "'") {
			value := strings.TrimPrefix(strings.TrimSuffix(part, "'"), "'")
			values = append(values, value)
		}
	}

	return values
}
