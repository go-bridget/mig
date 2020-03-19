package migrate

import (
	"regexp"
	"strings"
)

func statements(contents []byte, err error) ([]string, error) {
	result := []string{}
	if err != nil {
		return result, err
	}

	stmts := regexp.MustCompilePOSIX(";$").Split(string(contents), -1)
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}
	return result, nil
}
