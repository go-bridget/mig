package php81

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/go-bridget/mig/cmd/mig/internal"
)

func Render(basePath string, service string, ns string, tables []*internal.Table) error {
	// Loop through tables/columns, return type error if any
	for _, table := range tables {
		tableName := internal.Camel(strings.TrimPrefix(table.Name, service+"_"))
		filename := path.Join(basePath, tableName+".php")

		tmpTable := *table
		tmpTable.Name = tableName

		output, err := RenderTable(&tmpTable, ns)
		if err != nil {
			return fmt.Errorf("Error rendering table template: %w", err)
		}

		if err := ioutil.WriteFile(filename, []byte(output), 0644); err != nil {
			return fmt.Errorf("error writing file %s: %w", filename, err)
		}
	}
	return nil
}
