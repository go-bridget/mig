package migrate

import (
	"embed"
	"fmt"
)

//go:embed *.sql
var schemaFS embed.FS

// drivers maps the driver name of a *sqlx.DB onto the one this package has a
// bookkeeping schema for. The same normalized name goes to db.AcquireLock,
// which knows the same three engines.
var drivers = map[string]string{
	"mysql":      "mysql",
	"postgres":   "postgres",
	"postgresql": "postgres",
	"pgx":        "postgres",
	"sqlite":     "sqlite",
	"sqlite3":    "sqlite",
}

// driver normalizes the driver name of a handle, reporting whether this
// package knows the engine at all.
func driver(name string) (string, bool) {
	normalized, ok := drivers[name]
	return normalized, ok
}

// tableDDL returns the statements creating the bookkeeping table on a driver.
// The postgres schema is five statements, not one, so this is a slice.
func tableDDL(driverName string) ([]string, error) {
	name := fmt.Sprintf("migrations-%s.sql", driverName)

	contents, err := schemaFS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", name, err)
	}

	return split(contents), nil
}
