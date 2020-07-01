package lint

import (
	"fmt"
	"strings"

	"github.com/go-bridget/mig/cmd/mig/internal"
)

const (
	errMissingTableComment  = "Table is missing comment: %s"
	errMissingColumnComment = "Column is missing comment: %s.%s"

	errInvalidColumnName = "Invalid column name %s.%s: %w"
	errInvalidTableName  = "Invalid table name %s: %w"
)

// validateTable checks individual columns
func validateTable(table *internal.Table) []error {
	errs := []error{}
	for _, column := range table.Columns {
		if strings.TrimSpace(column.Comment) == "" {
			errs = append(errs, fmt.Errorf(errMissingColumnComment, table.Name, column.Name))
		}
		if err := isColumnNameValid(column.Name); err != nil {
			errs = append(errs, fmt.Errorf(errInvalidColumnName, table.Name, column.Name, err))
		}
	}
	return errs
}

// validate checks each table has a set comment
func validate(tables []*internal.Table) []error {
	errs := []error{}
	for _, table := range tables {
		if table.Ignore() {
			continue
		}
		if strings.TrimSpace(table.Comment) == "" {
			errs = append(errs, fmt.Errorf(errMissingTableComment, table.Name))
		}
		if err := isTableNameValid(table.Name); err != nil {
			errs = append(errs, fmt.Errorf(errInvalidTableName, table.Name, err))
		}
		if tableErrs := validateTable(table); len(tableErrs) > 0 {
			errs = append(errs, tableErrs...)
		}
	}
	return errs
}
