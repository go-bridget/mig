package internal

import (
	"context"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"github.com/stoewer/go-strcase"

	"github.com/go-bridget/mig/db"
)

// only validate base tables, not views
const tableType = "BASE TABLE"

// PRIMARY is the mysql-default schema value hint. SQLite maps `pk=1` into this.
const PRIMARY = "PRI"

var sqliteTypeMapping = map[string]string{
	"integer": "bigint",
	"real":    "double",
	"text":    "varchar",
}

func ListTables(ctx context.Context, config db.Options, schema string) ([]*Table, error) {
	handle, err := db.ConnectWithRetry(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}

	tables := []*Table{}

	switch config.Credentials.Driver {
	case "sqlite":
		err = handle.Select(&tables, "select name TABLE_NAME, '' TABLE_COMMENT from sqlite_schema where type='table' and name not like 'sqlite_%'")
	default:
		err = handle.Select(&tables, "select TABLE_NAME, TABLE_COMMENT from information_schema.tables where table_schema=? and table_type=? order by table_name asc", schema, tableType)
	}

	if err != nil {
		return nil, errors.Wrap(err, "error listing database tables")
	}

	// List columns in tables
	for _, table := range tables {
		switch config.Credentials.Driver {
		case "sqlite":
			err = handle.Select(&table.Columns, "select name COLUMN_NAME, type COLUMN_TYPE, type DATA_TYPE, pk COLUMN_KEY from pragma_table_info((?))", table.Name)
			for _, column := range table.Columns {
				column := column

				comment := strcase.KebabCase(column.Name)
				comment = strings.ReplaceAll(comment, "-", " ")
				commentRune := []rune(comment)
				commentRune[0] = unicode.ToUpper(commentRune[0])
				comment = string(commentRune)
				column.Comment = comment

				column.DataType = strings.ToLower(column.DataType)
				if mapped, ok := sqliteTypeMapping[column.DataType]; ok {
					column.DataType = mapped
				}

				if column.Key == "1" {
					column.Key = PRIMARY
				} else {
					column.Key = ""
				}
			}
		default:
			fields := strings.Join(ColumnFields, ", ")
			err = handle.Select(&table.Columns, "select "+fields+" from information_schema.columns where table_schema=? and table_name=? order by ordinal_position asc", schema, table.Name)
		}

		if err != nil {
			return nil, errors.Wrap(err, "error listing database columns for table: "+table.Name)
		}
	}

	return tables, nil
}
