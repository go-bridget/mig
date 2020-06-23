package internal

import (
	"strings"
)

// Table is an information_schema record
type Table struct {
	Name    string `db:"TABLE_NAME"`
	Comment string `db:"TABLE_COMMENT"`

	Columns []*Column
}

func (t *Table) Ignore() bool {
	return strings.TrimSpace(strings.ToLower(t.Comment)) == "ignore"
}

// TableFields lists database columns from Table{}
var TableFields = []string{"TABLE_NAME", "TABLE_COMMENT"}

// Column is an information_schema record
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
