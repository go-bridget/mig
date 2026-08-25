package migrate

import (
	"regexp"
	"strings"

	"github.com/gofrs/uuid"
)

var (
	// uuidCall matches the uuid() builtin, in any case.
	uuidCall = regexp.MustCompile(`(?i)uuid\(\)`)

	// comments matches a SQL comment to the end of its line.
	comments = regexp.MustCompile(`\s*--.*`)

	// terminator matches the semicolon ending a statement, which has to be
	// the last thing on its line.
	terminator = regexp.MustCompile(`(?m);$`)
)

// builtins replaces the uuid() builtin with a fresh UUID literal, so a
// migration can seed a row without the database having a UUID function.
func builtins(s string) string {
	return uuidCall.ReplaceAllStringFunc(s, func(_ string) string {
		return `'` + uuid.Must(uuid.NewV4()).String() + `'`
	})
}

// split turns the contents of a migration into its statements, dropping
// comments and expanding builtins.
func split(contents []byte) []string {
	result := []string{}

	contents = comments.ReplaceAll(contents, nil)

	for _, stmt := range terminator.Split(string(contents), -1) {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, builtins(stmt))
		}
	}

	return result
}
