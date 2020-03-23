package db

import (
	"context"
	"time"

	"database/sql"

	"github.com/go-bridget/mig/cli"
)

type (
	// Credentials contains DSN and Driver
	Credentials struct {
		DSN    string
		Driver string
	}

	// Options include database connection options
	Options struct {
		Credentials Credentials

		// Connector is an optional parameter to produce our
		// own *sql.DB, which is then wrapped in *sqlx.DB
		Connector func(context.Context, Credentials) (*sql.DB, error)

		Retries        int
		RetryDelay     time.Duration
		ConnectTimeout time.Duration
	}
)

func (options *Options) Bind() {
	options.BindWithPrefix("db")
	return
}

func (options *Options) BindWithPrefix(prefix string) {
	p := func(s string) string {
		return prefix + "-" + s
	}
	cli.StringVar(&options.Credentials.Driver, p("driver"), "mysql", "Database driver")
	cli.StringVar(&options.Credentials.DSN, p("dsn"), "", "DSN for database connection")
}
