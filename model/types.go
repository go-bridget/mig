package model

import (
	"slices"
	"strings"

	stylecheck "honnef.co/go/tools/config"
)

// Table is a database table with its columns
type Table struct {
	Name    string `db:"TABLE_NAME"`
	Comment string `db:"TABLE_COMMENT"`

	Columns []*Column
}

// Map returns a typed map of the Table. Comment may be empty.
func (t *Table) Map() map[string]string {
	return map[string]string{
		"table_name":    t.Name,
		"table_comment": t.Comment,
	}
}

// Title returns a human-readable title for the table
func (t *Table) Title() string {
	return Title(t.Name)
}

// Ignore returns true if the table comment indicates it should be ignored
func (t *Table) Ignore() bool {
	return strings.TrimSpace(strings.ToLower(t.Comment)) == "ignore"
}

// TableFields lists database columns from Table{}
var TableFields = []string{"TABLE_NAME", "TABLE_COMMENT"}

// Column is a database column with its metadata
type Column struct {
	Name    string `db:"COLUMN_NAME"`
	Type    string `db:"COLUMN_TYPE"`
	Key     string `db:"COLUMN_KEY"`
	Comment string `db:"COLUMN_COMMENT"`

	// Holds the clean data type
	DataType string `db:"DATA_TYPE"`
}

// ColumnFields lists database columns from Column{}
var ColumnFields = []string{"COLUMN_NAME", "COLUMN_TYPE", "COLUMN_KEY", "COLUMN_COMMENT", "DATA_TYPE"}

// Title returns a human-readable title for the column
func (c *Column) Title() string {
	return Title(c.Name)
}

// Title converts snake_case to Title Case (with spaces)
func Title(input string) string {
	// special case from having camel case `showId` fields in DB
	if len(input) > 2 && input[len(input)-2:] == "Id" {
		input = input[0:len(input)-2] + "_id"
	}

	// split string and check against initialisms
	keys := strings.Split(input, "_")
	for k, v := range keys {
		upper := strings.ToUpper(v)
		if slices.Contains(stylecheck.DefaultConfig.Initialisms, upper) {
			keys[k] = upper
			continue
		}
		if len(v) > 0 {
			keys[k] = upper[0:1] + v[1:]
		}
	}

	return strings.Join(keys, " ")
}
