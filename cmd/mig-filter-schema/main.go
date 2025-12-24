package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/go-bridget/mig/model"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: mig-filter-schema <schema.yaml>\n")
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	var tables []*model.Table
	if err := yaml.Unmarshal(data, &tables); err != nil {
		log.Fatal(err)
	}

	// Filter to keep only name, normalized_type, and enum_values
	filtered := make([]*model.Table, 0, len(tables))
	for _, table := range tables {
		newTable := &model.Table{
			Name: table.Name,
		}
		for _, col := range table.Columns {
			newTable.Columns = append(newTable.Columns, &model.Column{
				Name:           col.Name,
				NormalizedType: col.NormalizedType,
				EnumValues:     col.EnumValues,
			})
		}
		if table.Indexes != nil {
			for _, idx := range table.Indexes {
				newTable.Indexes = append(newTable.Indexes, &model.Index{
					Columns: idx.Columns,
					Primary: idx.Primary,
					Unique:  idx.Unique,
				})
			}
		}
		filtered = append(filtered, newTable)
	}

	// Output as YAML
	out, err := yaml.Marshal(filtered)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(out))
}
