package php81

import (
	"strings"
)

type Field struct {
	Type    string
	Format  string
	Default string
}

func NewField(kind string, defaultVal string) Field {
	return Field{
		Type:    kind,
		Default: defaultVal,
	}
}

// The definition of PHP native types
var (
	Int    Field = NewField("int", "0")
	String       = NewField("string", "\"\"")
	Float        = NewField("float", "0.0")
	Bool         = NewField("bool", "false")
	Mixed        = NewField("mixed", "null")
)

// The mapping from MySQL to PHP native types
var typeAliases = map[string]Field{
	"tinyint":    Int,
	"smallint":   Int,
	"mediumint":  Int,
	"int":        Int,
	"bigint":     Int,
	"char":       String,
	"varchar":    String,
	"text":       String,
	"longtext":   String,
	"mediumtext": String,
	"tinytext":   String,
	"longblob":   String,
	"blob":       String,
	"binary":     String,
	"varbinary":  String,
	"float":      Float,
	"double":     Float,
	"decimal":    String,
	"enum":       String,
	"year":       Int,
	"date":       String,
	"datetime":   String,
	"time":       String,
	"timestamp":  Int,
}

func typeAlias(kinder string) Field {
	// kinder may be "unsigned int" or something
	for _, kind := range strings.Fields(strings.ToLower(kinder)) {
		if kind == "unsigned" {
			continue
		}
		if val, ok := typeAliases[kind]; ok {
			val.Format = kind
			return val
		}
	}
	return Mixed
}
