package lint

import (
	"fmt"
	"strings"

	"github.com/go-bridget/mig/cmd/mig/internal"
)

const (
	errMissingTableComment  = "Table is missing comment: %s"
	errMissingColumnComment = "Column is missing comment: %s.%s"
)

// validateTable checks individual columns
func validateTable(table *internal.Table) []error {
	errs := []error{}
	for _, column := range table.Columns {
		if strings.TrimSpace(column.Comment) == "" {
			errs = append(errs, fmt.Errorf(errMissingColumnComment, table.Name, column.Name))
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
		if tableErrs := validateTable(table); len(tableErrs) > 0 {
			errs = append(errs, tableErrs...)
		}
	}
	return errs
}
