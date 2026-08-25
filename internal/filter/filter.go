package filter

import (
	"context"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/titpetric/cli"

	"github.com/go-bridget/mig/model"
)

// Name is the command title.
const Name = "Filter schema YAML to driver-independent fields"

// New creates a new filter command.
func New() *cli.Command {
	return &cli.Command{
		Name:  "filter",
		Title: Name,
		Run: func(_ context.Context, args []string) error {
			data, err := read(args)
			if err != nil {
				return err
			}

			var tables []*model.Table
			if err := yaml.Unmarshal(data, &tables); err != nil {
				return err
			}

			out, err := yaml.Marshal(Filter(tables))
			if err != nil {
				return err
			}

			fmt.Print(string(out))
			return nil
		},
	}
}

// read returns the schema contents from the first argument, or stdin
// when no argument is given.
func read(args []string) ([]byte, error) {
	if len(args) > 0 {
		return os.ReadFile(args[0])
	}
	return io.ReadAll(os.Stdin)
}

// Filter keeps only the fields which mean the same thing on every
// driver. Dropped: the driver-specific column type and key marker, and
// the generated index names.
func Filter(tables []*model.Table) []*model.Table {
	filtered := make([]*model.Table, 0, len(tables))
	for _, table := range tables {
		newTable := &model.Table{
			Name:    table.Name,
			Comment: table.Comment,
		}
		for _, col := range table.Columns {
			newTable.Columns = append(newTable.Columns, &model.Column{
				Name:     col.Name,
				Comment:  col.Comment,
				DataType: col.DataType,
				Size:     col.Size,
				Values:   col.Values,
			})
		}
		for _, idx := range table.Indexes {
			newTable.Indexes = append(newTable.Indexes, &model.Index{
				Columns: idx.Columns,
				Primary: idx.Primary,
				Unique:  idx.Unique,
			})
		}
		filtered = append(filtered, newTable)
	}
	return filtered
}
