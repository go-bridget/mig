package db

import (
	"strings"
)

// ParseDSN will return driver and dsn valid for db.Open.
func ParseDSN(conn string) (string, string) {
	return NewCredentials(conn).Open()
}

// addOptionToDSN appends option to dsn when match isn't in it already.
func addOptionToDSN(dsn, match, option string) string {
	if !strings.Contains(dsn, match) {
		dsn += option
	}
	return dsn
}
