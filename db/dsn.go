package db

import "strings"

// DeriveDriverFromDSN infers the database driver from a DSN connection string.
// Supports:
// - postgres:// or postgresql:// → pgx
// - mysql:// → mysql
// - sqlite:// or file: → sqlite
// - driver-specific formats (e.g., user:pass@tcp(host:port)/db → mysql).
func DeriveDriverFromDSN(dsn string) string {
	// Extract the schema/protocol prefix as the driver indicator
	if strings.Contains(dsn, "://") {
		scheme := strings.Split(dsn, "://")[0]
		switch scheme {
		case "postgres", "postgresql":
			return "pgx"
		case "mysql":
			return "mysql"
		case "sqlite", "file":
			return "sqlite"
		}
	}

	// Fallback: check for go-sql-driver/mysql format (user:password@tcp(host:port)/database)
	if strings.Contains(dsn, "@tcp(") || strings.Contains(dsn, "@unix(") {
		return "mysql"
	}

	// Default to sqlite for memory or unknown formats
	return "sqlite"
}

// CleanDSN applies driver-specific formatting to a DSN connection string.
// It automatically derives the driver from the DSN and applies appropriate transformations.
func CleanDSN(dsn string) string {
	driver := DeriveDriverFromDSN(dsn)
	return cleanDSNForDriver(dsn, driver)
}

func cleanDSNForDriver(dsn, driver string) string {
	// Trim protocol prefix for mysql and sqlite, but keep postgres:// as-is
	if driver == "mysql" && strings.HasPrefix(dsn, "mysql://") {
		dsn = strings.TrimPrefix(dsn, "mysql://")
	} else if driver == "sqlite" && strings.HasPrefix(dsn, "sqlite://") {
		dsn = strings.TrimPrefix(dsn, "sqlite://")
	}
	// postgres:// and postgresql:// are kept as-is (pgx expects the full connection string)

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
