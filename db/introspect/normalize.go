package introspect

import (
	"regexp"
	"sort"
	"strings"

	"github.com/go-bridget/mig/model"
)

// EnrichKeyMetadata marks columns with _id suffix or in indexes as "MUL".
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

// NormalizeColumnType sets DataType to the normalized cross-database type based on the source driver.
func NormalizeColumnType(column *model.Column, dbDriver string) {
	typeStr := strings.ToLower(column.Type)
	dataType := strings.ToLower(column.DataType)

	// Check for ENUM first (all drivers)
	// ENUM can be explicit (typeStr contains "enum") or implicit (EnumValues already extracted)
	if strings.Contains(typeStr, "enum") || strings.Contains(dataType, "enum") || len(column.EnumValues) > 0 {
		column.DataType = "enum"
		// Sort enum values for consistency across databases
		sort.Strings(column.EnumValues)
		return
	}

	// Handle database-specific type detection
	switch {
	// Boolean types
	case strings.Contains(typeStr, "bool"), strings.Contains(typeStr, "bit"):
		column.DataType = "boolean"

	// Timestamp/DateTime types
	case strings.Contains(typeStr, "timestamp"), strings.Contains(typeStr, "datetime"):
		column.DataType = "timestamp"

	// Date types
	case strings.Contains(typeStr, "date"):
		column.DataType = "date"

	// Decimal/Float types
	case strings.Contains(typeStr, "decimal"), strings.Contains(typeStr, "numeric"),
		strings.Contains(typeStr, "float"), strings.Contains(typeStr, "double"):
		column.DataType = "decimal"

	// Integer types (all precisions normalized to "integer")
	// Assumes all integers are up to 64-bit capacity
	case strings.Contains(typeStr, "bigint"), strings.Contains(typeStr, "int8"), strings.Contains(typeStr, "long"),
		strings.Contains(typeStr, "int"), strings.Contains(typeStr, "int4"), strings.Contains(typeStr, "serial"),
		strings.Contains(typeStr, "smallint"), strings.Contains(typeStr, "int2"), strings.Contains(typeStr, "tinyint"), strings.Contains(typeStr, "short"):
		column.DataType = "integer"

	// Text/String types (all normalized to "text")
	// Includes JSON/JSONB which are text-based data types
	case strings.Contains(typeStr, "varchar"), strings.Contains(typeStr, "char"), strings.Contains(typeStr, "string"),
		strings.Contains(typeStr, "longtext"), strings.Contains(typeStr, "long varchar"), strings.Contains(typeStr, "text"),
		strings.Contains(typeStr, "json"):
		column.DataType = "text"

	// Binary/BLOB types (raw binary data)
	case strings.Contains(typeStr, "blob"), strings.Contains(dataType, "blob"):
		column.DataType = "blob"

	// Default fallback
	default:
		column.DataType = "unknown"
	}
}

// ExtractEnumValues parses ENUM values from MySQL ENUM('value1','value2',...) format.
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
