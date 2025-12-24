package db

import "strings"

func cleanDSN(dsn string) string {
	return cleanDSNForDriver(dsn, "")
}

func cleanDSNForDriver(dsn, driver string) string {
	// Only apply MySQL-specific options for MySQL driver
	if driver != "mysql" {
		return dsn
	}

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
