package model

import (
	"strings"
)

// Ignore returns true if the table comment indicates it should be ignored
func Ignore(t *Table) bool {
	return strings.TrimSpace(strings.ToLower(t.Comment)) == "ignore"
}
