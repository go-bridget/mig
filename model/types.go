package model

import (
	"slices"
	"strings"

	stylecheck "honnef.co/go/tools/config"
)

// Table represents a database table with its columns and indexes.
type Table struct {
	Name    string `db:"TABLE_NAME" json:"name" yaml:"name"`
	Comment string `db:"TABLE_COMMENT" json:"comment,omitempty" yaml:"comment,omitempty"`

	Columns []*Column `json:"columns" yaml:"columns"`
	Indexes []*Index  `json:"indexes,omitempty" yaml:"indexes,omitempty"`
}

// Map returns a typed map of the Table fields. Comment may be empty.
func (t *Table) Map() map[string]string {
	return map[string]string{
		"table_name":    t.Name,
		"table_comment": t.Comment,
	}
}

// Title returns a human-readable title for the table.
func (t *Table) Title() string {
	return Title(t.Name)
}

// Ignore returns true if the table comment indicates it should be ignored.
func (t *Table) Ignore() bool {
	return strings.TrimSpace(strings.ToLower(t.Comment)) == "ignore"
}

// TableFields lists the database columns queried from Table.
var TableFields = []string{"TABLE_NAME", "TABLE_COMMENT"}

// Column represents a database column with its metadata.
// Type is the base SQL type (e.g., "varchar", "bigint", "enum").
// DataType is the normalized cross-database type (e.g., "text", "integer", "enum").
// Size is set for types where it has semantic meaning: varchar/char size represents max character length, PostgreSQL int types size represents storage in bytes (2, 4, 8).
// Values contains the allowed values for enum types.
// EnumValues is deprecated; use Values instead.
type Column struct {
	Name       string   `db:"COLUMN_NAME" json:"name" yaml:"name"`
	Type       string   `db:"COLUMN_TYPE" json:"type,omitempty" yaml:"type,omitempty"`
	Key        string   `db:"COLUMN_KEY" json:"key,omitempty" yaml:"key,omitempty"`
	Comment    string   `db:"COLUMN_COMMENT" json:"comment,omitempty" yaml:"comment,omitempty"`
	DataType   string   `db:"DATA_TYPE" json:"datatype,omitempty" yaml:"datatype,omitempty"`
	Size       int      `json:"size,omitempty" yaml:"size,omitempty"`
	Values     []string `json:"values,omitempty" yaml:"values,omitempty"`
	EnumValues []string `json:"enum_values,omitempty" yaml:"enum_values,omitempty"`
}

// ColumnFields lists the database columns queried from Column.
var ColumnFields = []string{"COLUMN_NAME", "COLUMN_TYPE", "COLUMN_KEY", "COLUMN_COMMENT", "DATA_TYPE"}

// Index represents a database index on a table.
type Index struct {
	Name    string   `json:"name,omitempty" yaml:"name,omitempty"`
	Columns []string `json:"columns" yaml:"columns"`
	Primary bool     `json:"primary,omitempty" yaml:"primary,omitempty"`
	Unique  bool     `json:"unique,omitempty" yaml:"unique,omitempty"`
}

// Title returns a human-readable title for the column.
func (c *Column) Title() string {
	return Title(c.Name)
}

// Title converts snake_case to Title Case with spaces.
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
