package php81

import (
	"bytes"
	"strings"

	"text/template"

	"github.com/go-bridget/mig/cmd/mig/internal"
)

func columnDefaultValue(column *internal.Column) string {
	return typeAlias(column.DataType).Default
}

func columnToNativeType(column *internal.Column) string {
	return typeAlias(column.DataType).Type
}

func RenderTable(data *internal.Table, ns string) (string, error) {
	funcMap := template.FuncMap{
		"NL":           func() string { return "\n" },
		"Camel":        internal.Camel,
		"DefaultValue": columnDefaultValue,
		"ToNativeType": columnToNativeType,
	}
	tpl, err := template.New("render-php81-table").Funcs(funcMap).Parse(tableTemplate)
	if err != nil {
		return "", err
	}

	out := new(bytes.Buffer)
	err = tpl.Execute(out, data)
	if err != nil {
		return "", err
	}

	sOut := out.String()
	sOut = strings.ReplaceAll(sOut, "\n\n\n", "\n\n")
	sOut = strings.ReplaceAll(sOut, "-\n", "")
	sOut = strings.ReplaceAll(sOut, "\n-", "")
	sOut = strings.TrimSpace(sOut) + "\n"

	if ns != "" {
		sOut = strings.ReplaceAll(sOut, "%NS%", "namespace "+ns+";")
	} else {
		sOut = strings.ReplaceAll(sOut, "\n%NS%\n", "")
	}

	return sOut, nil
}

const tableTemplate = `
<?php

%NS%

/** {{ .Comment }} */
class {{ .Name | Camel }}
{
	public function __construct(
-{{range .Columns}}
		/** {{ .Comment }} */
		public ?{{. | ToNativeType}} ${{.Name}} = {{. | DefaultValue}},
{{end}}
-	) {}
}
`
