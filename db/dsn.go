package db

import (
	"strings"
)

// Credentials contains database connection DSN with a schema:// driver hint.
type Credentials struct {
	DSN string
}

// NewCredentials returns a filled Credentials
func NewCredentials(dsn string) Credentials {
	return Credentials{
		DSN: dsn,
	}
}

// ParseDSN will return driver and dsn valid for db.Open.
func ParseDSN(conn string) (string, string) {
	return NewCredentials(conn).Open()
}

// Open returns the driver and the connection string.
func (c Credentials) Open() (string, string) {
	driver, dsn := c.parse(c.DSN)
	switch driver {
	case "mysql":
		dsn = addOptionToDSN(dsn, "?", "?")
		dsn = addOptionToDSN(dsn, "collation=", "&collation=utf8mb4_general_ci")
		dsn = addOptionToDSN(dsn, "parseTime=", "&parseTime=true")
		dsn = addOptionToDSN(dsn, "loc=", "&loc=Local")
		dsn = strings.Replace(dsn, "?&", "?", 1)
	}
	return driver, dsn
}

func (Credentials) parse(dsn string) (string, string) {
	// Extract the schema/protocol prefix as the driver indicator
	if strings.Contains(dsn, "://") {
		scheme := strings.Split(dsn, "://")
		switch scheme[0] {
		case "postgres", "postgresql":
			return "pgx", "postgres://" + scheme[1]
		case "mysql":
			return "mysql", scheme[1]
		case "sqlite", "file":
			return "sqlite", scheme[1]
		}
	}

	// Fallback: check for go-sql-driver/mysql format (user:password@tcp(host:port)/database)
	if strings.Contains(dsn, "@tcp(") || strings.Contains(dsn, "@unix(") {
		return "mysql", dsn
	}

	// Default to sqlite for memory or unknown formats
	return "sqlite", dsn
}

func addOptionToDSN(dsn, match, option string) string {
	if !strings.Contains(dsn, match) {
		dsn += option
	}
	return dsn
}
