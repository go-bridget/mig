package docs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/go-bridget/mig/cmd/mig/internal"
)

func renderMarkdownTable(table *internal.Table) []byte {
	// calculate initial padding from table header
	titles := []string{"Name", "Type", "Key", "Comment"}
	padding := map[string]int{}
	for _, v := range titles {
		padding[v] = len(v)
	}

	max := func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}

	// calculate max length for columns for padding
	for _, column := range table.Columns {
		if column.Comment == "" {
			column.Comment = internal.Title(column.Name)
		}
		padding["Name"] = max(padding["Name"], len(column.Name))
		padding["Type"] = max(padding["Type"], len(column.Type))
		padding["Key"] = max(padding["Key"], len(column.Key))
		padding["Comment"] = max(padding["Comment"], len(column.Comment))
	}

	// use fmt.Sprintf to add padding to columns, left align columns
	format := strings.Repeat("| %%-%ds ", len(padding)) + "|\n"

	// %%-%ds becomes %-10s, which right pads string to len=10
	paddings := []interface{}{
		padding["Name"],
		padding["Type"],
		padding["Key"],
		padding["Comment"],
	}
	format = fmt.Sprintf(format, paddings...)

	// create initial buffer with table name
	buf := bytes.NewBufferString(fmt.Sprintf("# %s\n\n", table.Title()))

	// and comment
	if table.Comment != "" {
		// add trailing dot (godoc)
		if !strings.HasSuffix(table.Comment, ".") {
			table.Comment += "."
		}

		buf.WriteString(fmt.Sprintf("%s\n\n", table.Comment))
	}

	// write header row strings to the buffer
	row := []interface{}{"Name", "Type", "Key", "Comment"}
	buf.WriteString(fmt.Sprintf(format, row...))

	// table header/body delimiter
	row = []interface{}{"", "", "", ""}
	buf.WriteString(strings.Replace(fmt.Sprintf(format, row...), " ", "-", -1))

	// table body
	for _, column := range table.Columns {
		row := []interface{}{column.Name, column.Type, column.Key, column.Comment}
		buf.WriteString(fmt.Sprintf(format, row...))
	}

	// return byte slice for writing to file
	return buf.Bytes()
}

func renderMarkdown(basePath string, filename string, tables []*internal.Table) error {
	// create output folder
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return err
	}

	// Write out single file
	if filename != "" {
		wr, err := os.Create(path.Join(basePath, filename))
		if err != nil {
			return err
		}
		defer wr.Close()

		for k, table := range tables {
			if table.Ignore() {
				continue
			}
			contents := renderMarkdownTable(table)
			if k > 0 {
				contents = append([]byte("\n"), contents...)
			}

			_, err := wr.Write(contents)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// generate individual markdown files with service
	for _, table := range tables {
		if table.Ignore() {
			continue
		}

		filename := path.Join(basePath, internal.Filename(table.Title())+".md")
		contents := renderMarkdownTable(table)

		if err := ioutil.WriteFile(filename, contents, 0644); err != nil {
			return err
		}

		fmt.Println(filename)
	}
	return nil
}
