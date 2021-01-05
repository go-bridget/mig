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
func validateTable(table *internal.Table, options Options) []error {
	errs := []error{}
	validComments := map[string]bool{
		"id": true,
	}
	for _, column := range table.Columns {
		validComment := validComments[strings.ToLower(column.Name)] || options.skipComments
		if !validComment && strings.TrimSpace(column.Comment) == "" {
			errs = append(errs, fmt.Errorf(errMissingColumnComment, table.Name, column.Name))
		}
		if err := isColumnNameValid(column.Name); err != nil {
			errs = append(errs, fmt.Errorf(errInvalidColumnName, table.Name, column.Name, err))
		}
	}
	return errs
}

// validate checks each table has a set comment
func validate(tables []*internal.Table, options Options) []error {
	errs := []error{}
	for _, table := range tables {
		if table.Ignore() {
			continue
		}
		if !options.skipComments && strings.TrimSpace(table.Comment) == "" {
			errs = append(errs, fmt.Errorf(errMissingTableComment, table.Name))
		}
		if err := isTableNameValid(table.Name, options); err != nil {
			errs = append(errs, fmt.Errorf(errInvalidTableName, table.Name, err))
		}
		if tableErrs := validateTable(table, options); len(tableErrs) > 0 {
			errs = append(errs, tableErrs...)
		}
	}
	return errs
}
