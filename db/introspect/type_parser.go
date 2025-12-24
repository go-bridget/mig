package introspect

import (
	"regexp"
	"strings"
)

var (
	// MySQL type patterns
	varcharPattern  = regexp.MustCompile(`(varchar|char)\s*\(\s*(\d+)\s*\)`)
	intDisplayWidth = regexp.MustCompile(`(\w+int|serial)\s*\(\s*\d+\s*\)`)
	numericPattern  = regexp.MustCompile(`(numeric|decimal)\s*\(\s*(\d+)\s*,\s*(\d+)\s*\)`)
)

// ParsePostgresIntType maps int2/int4/int8 to base type and size in bytes
func ParsePostgresIntType(typeStr string) (baseType string, sizeBytes int) {
	typeStr = strings.ToLower(strings.TrimSpace(typeStr))

	switch typeStr {
	case "int2", "smallint":
		return "integer", 2
	case "int4", "integer", "int":
		return "integer", 4
	case "int8", "bigint":
		return "integer", 8
	default:
		return typeStr, 0
	}
}
