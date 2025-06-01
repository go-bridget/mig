package php81

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/go-bridget/mig/cmd/mig/gen/model"
	"github.com/go-bridget/mig/cmd/mig/internal"
)

func Render(options model.Options, tables []*internal.Table) error {
	var (
		basePath = options.Output
		service  = options.Schema
	)

	// Loop through tables/columns, return type error if any
	for _, table := range tables {
		tableName := internal.Camel(strings.TrimPrefix(table.Name, service+"_"))
		filename := path.Join(basePath, tableName+".php")

		tmpTable := *table
		tmpTable.Name = tableName

		output, err := RenderTable(&tmpTable, options.PHP.Namespace)
		if err != nil {
			return fmt.Errorf("Error rendering table template: %w", err)
		}

		if err := ioutil.WriteFile(filename, []byte(output), 0644); err != nil {
			return fmt.Errorf("error writing file %s: %w", filename, err)
		}
	}
	return nil
}
