package introspect

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/go-bridget/mig/model"
)

var (
	// MySQL type patterns
	varcharPattern  = regexp.MustCompile(`(varchar|char)\s*\(\s*(\d+)\s*\)`)
	intDisplayWidth = regexp.MustCompile(`(\w+int|serial)\s*\(\s*\d+\s*\)`)
	numericPattern  = regexp.MustCompile(`(numeric|decimal)\s*\(\s*(\d+)\s*,\s*(\d+)\s*\)`)
)

func parseSqliteType(column *model.Column) {
	typeStr := strings.ToLower(strings.TrimSpace(column.Type))
	column.DataType = typeStr

	var sqliteTypeMapping = map[string]string{
		"integer": "bigint",
		"real":    "double",
		"text":    "varchar",
		"blob":    "blob",
	}

	if mapped, ok := sqliteTypeMapping[typeStr]; ok {
		column.DataType = mapped
		if mapped == "bigint" {
			column.Size = 8
		}
	}
}

func parsePostgresType(ctx context.Context, db *sqlx.DB, column *model.Column) {
	typeStr := strings.ToLower(strings.TrimSpace(column.Type))
	column.DataType = typeStr

	switch typeStr {
	case "int2", "smallint", "smallserial":
		column.Type = "integer"
		column.DataType = "smallint"
		column.Size = 2
		return
	case "int4", "integer", "int", "serial":
		column.Type = "integer"
		column.DataType = "int"
		column.Size = 4
		return
	case "int8", "bigint", "bigserial":
		column.Type = "integer"
		column.DataType = "bigint"
		column.Size = 8
		return
	}

	// Try to extract ENUM values for all columns (custom types and enum types)
	// extractPostgresEnumValues will return nil if the type is not an enum
	enumVals := extractPostgresEnumValues(ctx, db, typeStr)
	if enumVals != nil && len(enumVals) > 0 {
		column.Values = enumVals
		column.Type = "enum"
		column.DataType = "enum"
	}
}

func parseMySQLType(column *model.Column) {
	typeStr := strings.ToLower(strings.TrimSpace(column.Type))
	column.DataType = typeStr

	if strings.Contains(typeStr, "enum") {
		column.Values = extractEnumValues(column.Type)
		column.Type = "enum"
		column.DataType = "enum"
		return
	}

	// Handle varchar/char - extract max length
	if matches := varcharPattern.FindStringSubmatch(typeStr); len(matches) >= 3 {
		column.Type = "text"
		column.DataType = matches[1]
		if n, err := strconv.Atoi(matches[2]); err == nil {
			column.Size = n
		}
		return
	}

	// Handle numeric/decimal - extract precision
	if matches := numericPattern.FindStringSubmatch(typeStr); len(matches) >= 3 {
		column.Type = "decimal"
		column.DataType = matches[1]
		if n, err := strconv.Atoi(matches[2]); err == nil {
			column.Size = n
		}
		return
	}

	// Strip display width from integer types (int(11), bigint(20), etc)
	// Display width has no semantic meaning for storage
	if matches := intDisplayWidth.FindStringSubmatch(typeStr); len(matches) >= 3 {
		column.Type = "integer"
		column.DataType = matches[1]
		column.Size = 0
	}

	switch column.DataType {
	case "bigint":
		column.Size = 8
	case "int":
		column.Size = 4
	case "smallint":
		column.Size = 2
	case "tinyint":
		column.Size = 1
	}
}
