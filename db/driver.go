package db

import (
	"database/sql"
	"reflect"
	"strings"

	"github.com/titpetric/pdo/client"
)

// DriverName names the engine behind a *sql.DB. database/sql does not keep
// the name the pool was opened with, so the driver's type is read instead:
// every driver this package supports has a package path of its own. The
// names are the ones the bookkeeping schema and the locking strategy are
// keyed by, so aliases (sqlite3, postgresql, pgx) come back normalized.
//
// It reports false for a pool this package has no schema for.
func DriverName(conn *sql.DB) (string, bool) {
	if conn == nil {
		return "", false
	}

	t := reflect.TypeOf(conn.Driver())
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil {
		return "", false
	}

	path := strings.ToLower(t.PkgPath())
	switch {
	case strings.Contains(path, "modernc.org/sqlite"),
		strings.Contains(path, "mattn/go-sqlite3"),
		strings.Contains(path, "sqlite"):
		return "sqlite", true
	case strings.Contains(path, "go-sql-driver/mysql"),
		strings.Contains(path, "mysql"):
		return "mysql", true
	case strings.Contains(path, "jackc/pgx"),
		strings.Contains(path, "lib/pq"),
		strings.Contains(path, "postgres"):
		return "postgres", true
	}

	return "", false
}

// Client returns a pdo client over conn, with the placeholder style of its
// detected driver. It is the handle mig's query layer takes.
func Client(conn *sql.DB) *client.Client {
	name, _ := DriverName(conn)
	return client.NewClient(conn, client.WithDriver(name))
}
