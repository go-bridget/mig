package introspect

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/go-bridget/mig/model"
)

// enrichKeysFromInfoSchema retrieves type, key, and comment from information_schema.
func enrichKeysFromInfoSchema(ctx context.Context, db *sqlx.DB, tableName string, columns []*model.Column) {
	var schemaColumns []*model.Column

	// Query information_schema for complete column information
	fields := strings.Join(model.ColumnFields, ", ")
	query := fmt.Sprintf(
		"SELECT %s FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? ORDER BY ordinal_position",
		fields,
	)

	if err := db.SelectContext(ctx, &schemaColumns, query, tableName); err != nil {
		// Silently ignore errors - we'll use DESCRIBE information as fallback
		return
	}

	// Create a map for quick lookup of schema information
	schemaMap := make(map[string]*model.Column)
	for _, sc := range schemaColumns {
		schemaMap[sc.Name] = sc
	}

	// Enrich our columns with schema information
	for _, col := range columns {
		if sc, exists := schemaMap[col.Name]; exists {
			// Update from schema if we got better information
			if col.Key == "" && sc.Key != "" {
				col.Key = sc.Key
			}
			if col.Comment == "" && sc.Comment != "" {
				col.Comment = sc.Comment
			}
			// DataType might be more precise from schema
			if sc.DataType != "" {
				col.DataType = sc.DataType
			}
		}
	}
}

// MysqlDescriber implements Describer for MySQL
type MysqlDescriber struct{}

// parseMySQLType extracts base type and size for varchar/char/numeric/decimal; strips integer display width.
func parseMySQLType(typeStr string) (baseType string, size int) {
	typeStr = strings.TrimSpace(typeStr)
	typeStrLower := strings.ToLower(typeStr)

	// Handle varchar/char - extract max length
	if matches := varcharPattern.FindStringSubmatch(typeStrLower); len(matches) >= 3 {
		if n, err := strconv.Atoi(matches[2]); err == nil {
			return matches[1], n
		}
		return matches[1], 0
	}

	// Handle numeric/decimal - extract precision
	if matches := numericPattern.FindStringSubmatch(typeStrLower); len(matches) >= 3 {
		if n, err := strconv.Atoi(matches[2]); err == nil {
			return matches[1], n
		}
		return matches[1], 0
	}

	// Strip display width from integer types (int(11), bigint(20), etc)
	// Display width has no semantic meaning for storage
	if intDisplayWidth.MatchString(typeStrLower) {
		typeStr = intDisplayWidth.ReplaceAllStringFunc(typeStrLower, func(match string) string {
			parts := strings.Split(match, "(")
			return strings.TrimSpace(parts[0])
		})
		return typeStr, 0
	}

	// Return as-is for other types
	return typeStr, 0
}

// Describe returns column metadata from a query.
func (d *MysqlDescriber) Describe(ctx context.Context, db *sqlx.DB, query string) ([]*model.Column, error) {
	var err error

	// Normalize query
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	// Generate unique temporary table name
	tableName := fmt.Sprintf("mig_temp_tbl_%d", time.Now().UnixNano())

	// Create temporary table from the query with LIMIT 0 to get only structure
	// This is more reliable than using TEMPORARY VIEW in MySQL
	createTableSQL := fmt.Sprintf("CREATE TEMPORARY TABLE `%s` AS %s LIMIT 0", tableName, query)
	if _, err = db.ExecContext(ctx, createTableSQL); err != nil {
		return nil, errors.Wrapf(err, "failed to create temporary table for query")
	}

	// Defer cleanup - drop the temporary table
	defer func() {
		dropSQL := fmt.Sprintf("DROP TEMPORARY TABLE IF EXISTS `%s`", tableName)
		_, _ = db.ExecContext(context.Background(), dropSQL)
	}()

	// Query the temporary table structure using DESCRIBE
	// This is more reliable than information_schema for temporary tables
	type describeRow struct {
		Field   string  `db:"Field"`
		Type    string  `db:"Type"`
		Null    string  `db:"Null"`
		Key     string  `db:"Key"`
		Default *string `db:"Default"`
		Extra   string  `db:"Extra"`
	}

	var describeRows []describeRow
	describeQuery := fmt.Sprintf("DESCRIBE `%s`", tableName)
	if err = db.SelectContext(ctx, &describeRows, describeQuery); err != nil {
		return nil, errors.Wrap(err, "failed to describe temporary table")
	}

	// Convert DESCRIBE output to Column format
	columns := []*model.Column{}
	for _, row := range describeRows {
		column := &model.Column{
			Name:     row.Field,
			Type:     row.Type,
			DataType: row.Type,
			Comment:  "",
		}

		// Map Key values - DESCRIBE shows keys
		column.Key = row.Key // Could be: "PRI", "UNI", "MUL", or empty

		columns = append(columns, column)
	}

	// Try to enrich key information from information_schema for temporary tables
	// This helps preserve primary key and constraint information
	enrichKeysFromInfoSchema(ctx, db, tableName, columns)

	return columns, nil
}

// DescribeTable returns the structure of a table.
func (d *MysqlDescriber) DescribeTable(ctx context.Context, db *sqlx.DB, tableName string) (*model.Table, error) {
	table := &model.Table{
		Name: tableName,
	}

	// Get table comment
	type tableRow struct {
		Comment string `db:"TABLE_COMMENT"`
	}
	var tr tableRow
	if err := db.GetContext(ctx, &tr, "SELECT TABLE_COMMENT FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?", tableName); err != nil {
		return nil, errors.Wrapf(err, "failed to get table comment for %s", tableName)
	}
	table.Comment = tr.Comment

	// Get columns using information_schema
	columns := []*model.Column{}
	fields := strings.Join(model.ColumnFields, ", ")
	query := fmt.Sprintf(
		"SELECT %s FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=? ORDER BY ordinal_position ASC",
		fields,
	)

	if err := db.SelectContext(ctx, &columns, query, tableName); err != nil {
		return nil, errors.Wrapf(err, "failed to get columns for table %s", tableName)
	}

	// Enrich columns with normalized type and extract ENUM values
	for _, col := range columns {
		// Extract ENUM values and set Values field
		if strings.Contains(strings.ToLower(col.Type), "enum") {
			col.Values = ExtractEnumValues(col.Type)
			col.EnumValues = col.Values // Keep for backward compatibility
			// Set type to just "enum" without the values
			col.Type = "enum"
		} else {
			// Parse MySQL type and extract meaningful size
			col.Type, col.Size = parseMySQLType(col.Type)
		}
		// Normalize the type
		NormalizeColumnType(col, "mysql")
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
func (d *MysqlDescriber) ListTables(ctx context.Context, db *sqlx.DB) ([]*model.Table, error) {
	const tableType = "BASE TABLE"

	tables := []*model.Table{}

	// Get all base tables (excluding views)
	if err := db.SelectContext(ctx, &tables, "SELECT TABLE_NAME, TABLE_COMMENT FROM information_schema.tables WHERE table_schema=DATABASE() AND table_type=? ORDER BY table_name ASC", tableType); err != nil {
		return nil, errors.Wrap(err, "failed to list tables")
	}

	return tables, nil
}

// TableIndexes returns all indexes for a MySQL table
func (d *MysqlDescriber) TableIndexes(ctx context.Context, db *sqlx.DB, tableName string) ([]*model.Index, error) {
	type indexStatistic struct {
		IndexName  string `db:"INDEX_NAME"`
		ColumnList string `db:"COLUMN_LIST"`
		NonUnique  int    `db:"NON_UNIQUE"`
	}

	var indexStats []indexStatistic
	query := `
		SELECT 
			INDEX_NAME,
			GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) as COLUMN_LIST,
			MAX(NON_UNIQUE) as NON_UNIQUE
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		GROUP BY INDEX_NAME
		ORDER BY INDEX_NAME
	`

	if err := db.SelectContext(ctx, &indexStats, query, tableName); err != nil {
		return nil, errors.Wrapf(err, "failed to get indexes for table %s", tableName)
	}

	var indexes []*model.Index
	for _, stat := range indexStats {
		indexes = append(indexes, &model.Index{
			Name:    stat.IndexName,
			Columns: strings.Split(stat.ColumnList, ","),
			Primary: stat.IndexName == "PRIMARY",
			Unique:  stat.NonUnique == 0,
		})
	}

	return indexes, nil
}
