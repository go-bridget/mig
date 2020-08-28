package internal

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/db"
)

// only validate base tables, not views
const tableType = "BASE TABLE"

func ListTables(ctx context.Context, config db.Options, schema string) ([]*Table, error) {
	handle, err := db.ConnectWithRetry(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}

	// List tables in schema
	tables := []*Table{}
	fields := strings.Join(TableFields, ", ")
	err = handle.Select(&tables, "select "+fields+" from information_schema.tables where table_schema=? and table_type=? order by table_name asc", schema, tableType)
	if err != nil {
		return nil, errors.Wrap(err, "error listing database tables")
	}

	// List columns in tables
	for _, table := range tables {
		fields := strings.Join(ColumnFields, ", ")
		err := handle.Select(&table.Columns, "select "+fields+" from information_schema.columns where table_schema=? and table_name=? order by ordinal_position asc", schema, table.Name)
		if err != nil {
			return nil, errors.Wrap(err, "error listing database columns for table: "+table.Name)
		}
	}

	return tables, nil
}
