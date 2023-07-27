package migrate

import (
	"regexp"
	"strings"

	"github.com/gofrs/uuid"
)

func builtins(s string) string {
	r := regexp.MustCompile(`(?i)uuid\(\)`)
	s = r.ReplaceAllStringFunc(s, func(_ string) string {
		val := uuid.Must(uuid.NewV4())
		return `'` + val.String() + `'`
	})
	return s
}

func statements(contents []byte, err error) ([]string, error) {
	result := []string{}
	if err != nil {
		return result, err
	}

	// remove sql comments from anywhere ([whitespace]--*\n)
	comments := regexp.MustCompile(`\s*--.*`)
	contents = comments.ReplaceAll(contents, nil)

	// split statements by trailing ; at the end of the line
	stmts := regexp.MustCompilePOSIX(`;$`).Split(string(contents), -1)
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, builtins(stmt))
		}
	}

	return result, nil
}
