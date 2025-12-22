package internal

import (
	"strings"

	"github.com/go-bridget/mig/model"
)

// Table is an alias for model.Table for backward compatibility
type Table = model.Table

// Column is an alias for model.Column for backward compatibility
type Column = model.Column

// TableFields is an alias for model.TableFields for backward compatibility
var TableFields = model.TableFields

// ColumnFields is an alias for model.ColumnFields for backward compatibility
var ColumnFields = model.ColumnFields

// Ignore returns true if the table comment indicates it should be ignored
func Ignore(t *Table) bool {
	return strings.TrimSpace(strings.ToLower(t.Comment)) == "ignore"
}
