package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"path"
	"strings"

	"github.com/apex/log"
	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cmd/mig/internal"
)

var numericTypes map[string]string = map[string]string{
	"tinyint":  "int8",
	"smallint": "int16",
	// `mediumint` - this one would, technically, be int24 (3 bytes), but
	"mediumint": "int32",
	"int":       "int32",
	"bigint":    "int64",
}

func isNumeric(column *internal.Column) (string, bool) {
	val, ok := numericTypes[column.DataType]
	return val, ok
}

var simpleTypes map[string]string = map[string]string{
	"char":       "string",
	"varchar":    "string",
	"text":       "string",
	"longtext":   "string",
	"mediumtext": "string",
	"tinytext":   "string",
	"longblob":   "[]byte",
	"blob":       "[]byte",
	"varbinary":  "[]byte",
	// `float` and `double` are here since they don't have unsigned modifiers
	"float":  "float32",
	"double": "float64",
	// `decimal` - double stored as string, \o/
	"decimal": "string",
}

func isSimple(column *internal.Column) (string, bool) {
	val, ok := simpleTypes[column.DataType]
	return val, ok
}

type specialType struct {
	Import string
	Type   string
}

var specialTypes map[string]specialType = map[string]specialType{
	"date":      specialType{"time", "*time.Time"},
	"datetime":  specialType{"time", "*time.Time"},
	"time":      specialType{"time", "*time.Time"},
	"timestamp": specialType{"time", "*time.Time"},
	// `enum` and `set` aren't implemented
	// `year` isn't implemented
	"json": specialType{"sqlxTypes github.com/jmoiron/sqlx/types", "sqlxTypes.JSONText"},
}

func isSpecial(column *internal.Column) (specialType, bool) {
	val, ok := specialTypes[column.DataType]
	return val, ok
}

func resolveTypeGo(column *internal.Column) (string, error) {
	if val, ok := isSimple(column); ok {
		return val, nil
	}
	if val, ok := isNumeric(column); ok {
		isUnsigned := strings.Contains(strings.ToLower(column.Type), "unsigned")
		if isUnsigned {
			return "u" + val, nil
		}
		return val, nil
	}
	if val, ok := isSpecial(column); ok {
		return val.Type, nil
	}
	return "", errors.Errorf("Unsupported SQL type: %s", column.DataType)
}

func renderGo(basePath string, service string, tables []*internal.Table) error {
	imports := []string{}

	// Loop through tables/columns, return type error if any
	// This also builds the `imports` slice for codegen lower
	for _, table := range tables {
		for _, column := range table.Columns {
			if val, ok := isSpecial(column); ok {
				importString := fmt.Sprintf("\"%s\"", val.Import)
				// "x y" => import x "y"
				if strings.Contains(val.Import, " ") {
					parts := strings.Split(val.Import, " ")
					importString = fmt.Sprintf("%s \"%s\"", parts[0], parts[1])
				}
				if !internal.Contains(imports, importString) {
					imports = append(imports, importString)
				}
			}
		}
	}

	buf := bytes.NewBuffer([]byte{})

	fmt.Fprintf(buf, "package %s\n", service)
	fmt.Fprintln(buf)

	// Print collected imports
	if len(imports) > 0 {
		fmt.Fprintln(buf, "import (")
		nl := false
		for _, val := range imports {
			if !strings.Contains(val, " ") {
				fmt.Fprintf(buf, "\t%s\n", val)
			} else {
				nl = true
			}
		}
		if nl {
			fmt.Fprintln(buf)
		}
		for _, val := range imports {
			if strings.Contains(val, " ") {
				fmt.Fprintf(buf, "\t%s\n", val)
			}
		}
		fmt.Fprintln(buf, ")")
		fmt.Fprintln(buf)
	}

	for _, table := range tables {
		fields := []string{}
		primary := []string{}
		setters := []string{}

		if table.Ignore() {
			continue
		}

		tableName := internal.Camel(strings.TrimPrefix(table.Name, service+"_"))

		fmt.Fprintf(buf, "// %s generated for db table `%s`\n", tableName, table.Name)
		if table.Comment != "" {
			fmt.Fprintln(buf, "//\n//", table.Comment)
		}
		fmt.Fprintf(buf, "type %s struct {\n", tableName)
		for idx, column := range table.Columns {
			columnName := internal.Camel(column.Name)
			fields = append(fields, column.Name)
			if column.Key == "PRI" {
				primary = append(primary, column.Name)
			}

			if column.Comment != "" {
				if idx > 0 {
					fmt.Fprintln(buf)
				}
				fmt.Fprintf(buf, "	// %s\n", column.Comment)
			}
			columnType, _ := resolveTypeGo(column)
			fmt.Fprintf(buf, "	%s %s `db:\"%s\" json:\"-\"`\n", columnName, columnType, column.Name)
			if columnType == "*time.Time" {
				receiver := strings.ToLower(string(tableName[0]))
				setters = append(setters, []string{
					fmt.Sprintf("// Set%s sets %s which requires a *time.Time", columnName, columnName),
					fmt.Sprintf("func (%s *%s) Set%s(stamp time.Time) { %s.%s = &stamp }", receiver, tableName, columnName, receiver, columnName),
				}...)
			}
		}
		fmt.Fprintln(buf, "}")
		fmt.Fprintln(buf)
		for _, v := range setters {
			fmt.Fprintln(buf, v)
		}
		if len(setters) > 0 {
			fmt.Fprintln(buf)
		}
		// Table name
		fmt.Fprintf(buf, "// %sTable is the name of the table in the DB\n", tableName)
		// Table is SQL backtick quoted so we can allow reserved words like `group`
		fmt.Fprintf(buf, "const %sTable = \"`%s`\"\n", tableName, table.Name)
		// Table fields
		fmt.Fprintf(buf, "// %sFields are all the field names in the DB table\n", tableName)
		fmt.Fprintf(buf, "var %sFields = ", tableName)
		if len(fields) > 0 {
			fmt.Fprintf(buf, "[]string{\"%s\"}", strings.Join(fields, "\", \""))
		} else {
			fmt.Fprintf(buf, "[]string{}")
		}
		fmt.Fprintln(buf)
		// Table primary keys
		fmt.Fprintf(buf, "// %sPrimaryFields are the primary key fields in the DB table\n", tableName)
		fmt.Fprintf(buf, "var %sPrimaryFields = ", tableName)
		if len(primary) > 0 {
			fmt.Fprintf(buf, "[]string{\"%s\"}", strings.Join(primary, "\", \""))
		} else {
			fmt.Fprintf(buf, "[]string{}")
		}
		fmt.Fprintln(buf)
	}

	filename := path.Join(basePath, "types_gen.go")
	contents := buf.Bytes()

	formatted, err := format.Source(contents)
	if err != nil {
		// fall back to unformatted source to inspect
		// the saved file for the error which occurred
		formatted = contents
		log.WithError(err).Errorf("error formatting %s", filename)
	}

	log.Info(filename)

	return ioutil.WriteFile(filename, formatted, 0644)
}
