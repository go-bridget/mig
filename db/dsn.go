package db

import "strings"

func cleanDSN(dsn string) string {
	dsn = addOptionToDSN(dsn, "?", "?")
	dsn = addOptionToDSN(dsn, "collation=", "&collation=utf8mb4_general_ci")
	dsn = addOptionToDSN(dsn, "parseTime=", "&parseTime=true")
	dsn = addOptionToDSN(dsn, "loc=", "&loc=Local")
	dsn = strings.Replace(dsn, "?&", "?", 1)
	return dsn
}

func addOptionToDSN(dsn, match, option string) string {
	if !strings.Contains(dsn, match) {
		dsn += option
	}
	return dsn
}
