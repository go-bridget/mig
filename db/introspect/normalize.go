package introspect

import (
	"regexp"
	"sort"
	"strings"

	"github.com/go-bridget/mig/model"
)

// EnrichKeyMetadata adds key indicators for columns based on naming conventions and indexes
// Sets Key to "MUL" for columns with _id suffix (foreign keys) or that are part of indexes
func EnrichKeyMetadata(columns []*model.Column, indexes []*model.Index) {
	// Create a map of indexed column names (excluding primary key indexes)
	indexedCols := make(map[string]bool)
	if indexes != nil {
		for _, idx := range indexes {
			// Skip primary key indexes
			if idx.Primary {
				continue
			}
			// Mark all columns in this index
			for _, col := range idx.Columns {
				indexedCols[strings.ToLower(col)] = true
			}
		}
	}

	for _, col := range columns {
		// Skip if already marked as primary key
		if col.Key == "PRI" {
			continue
		}

		// Mark columns that are indexed or end with _id as "MUL"
		colNameLower := strings.ToLower(col.Name)
		if indexedCols[colNameLower] || strings.HasSuffix(colNameLower, "_id") {
			col.Key = "MUL"
		}
	}
}

// NormalizeColumnType maps database-specific types to logical normalized types
func NormalizeColumnType(column *model.Column, dbDriver string) {
	typeStr := strings.ToLower(column.Type)
	dataType := strings.ToLower(column.DataType)

	// Check for ENUM first (all drivers)
	// ENUM can be explicit (typeStr contains "enum") or implicit (EnumValues already extracted)
	if strings.Contains(typeStr, "enum") || strings.Contains(dataType, "enum") || len(column.EnumValues) > 0 {
		column.NormalizedType = "enum"
		// Sort enum values for consistency across databases
		sort.Strings(column.EnumValues)
		return
	}

	// Handle database-specific type detection
	switch {
	// Boolean types
	case strings.Contains(typeStr, "bool"), strings.Contains(typeStr, "bit"):
		column.NormalizedType = "boolean"

	// Timestamp/DateTime types
	case strings.Contains(typeStr, "timestamp"), strings.Contains(typeStr, "datetime"):
		column.NormalizedType = "timestamp"

	// Date types
	case strings.Contains(typeStr, "date"):
		column.NormalizedType = "date"

	// Decimal/Float types
	case strings.Contains(typeStr, "decimal"), strings.Contains(typeStr, "numeric"),
		strings.Contains(typeStr, "float"), strings.Contains(typeStr, "double"):
		column.NormalizedType = "decimal"

	// Integer types (all precisions normalized to "integer")
	// Assumes all integers are up to 64-bit capacity
	case strings.Contains(typeStr, "bigint"), strings.Contains(typeStr, "int8"), strings.Contains(typeStr, "long"),
		strings.Contains(typeStr, "int"), strings.Contains(typeStr, "int4"), strings.Contains(typeStr, "serial"),
		strings.Contains(typeStr, "smallint"), strings.Contains(typeStr, "int2"), strings.Contains(typeStr, "tinyint"), strings.Contains(typeStr, "short"):
		column.NormalizedType = "integer"

	// Text/String types (all normalized to "text")
	// Includes JSON/JSONB which are text-based data types
	case strings.Contains(typeStr, "varchar"), strings.Contains(typeStr, "char"), strings.Contains(typeStr, "string"),
		strings.Contains(typeStr, "longtext"), strings.Contains(typeStr, "long varchar"), strings.Contains(typeStr, "text"),
		strings.Contains(typeStr, "json"):
		column.NormalizedType = "text"

	// Binary/BLOB types (raw binary data)
	case strings.Contains(typeStr, "blob"), strings.Contains(dataType, "blob"):
		column.NormalizedType = "blob"

	// Default fallback
	default:
		column.NormalizedType = "unknown"
	}
}

// ExtractEnumValues parses ENUM values from column type string
// MySQL format: ENUM('value1','value2',...)
// PostgreSQL/SQLite: handled separately by their describers
func ExtractEnumValues(typeStr string) []string {
	// Match ENUM('value1','value2',...)
	re := regexp.MustCompile(`(?i)enum\s*\(\s*'([^']*)'\s*(?:,\s*'([^']*)'\s*)*\)`)
	matches := re.FindStringSubmatch(typeStr)
	if len(matches) < 2 {
		return nil
	}

	// Extract all quoted values
	valueRe := regexp.MustCompile(`'([^']*)'`)
	valueMatches := valueRe.FindAllStringSubmatch(typeStr, -1)
	if len(valueMatches) == 0 {
		return nil
	}

	values := make([]string, 0, len(valueMatches))
	for _, match := range valueMatches {
		if len(match) > 1 {
			values = append(values, match[1])
		}
	}

	return values
}
