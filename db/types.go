package db

import (
	"context"
	"os"
	"time"

	"database/sql"

	"github.com/go-bridget/mig/cli"
)

type (
	// Credentials contains database connection DSN
	Credentials struct {
		DSN string
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

// NewOptions provides an initialized *Options object
func NewOptions() *Options {
	return (&Options{}).Init()
}

// Init sets default *Options values
func (options *Options) Init() *Options {
	options.Retries = 100
	options.RetryDelay = 2 * time.Second
	options.ConnectTimeout = 2 * time.Minute
	options.Credentials.DSN = os.Getenv("MIG_DB_DSN")
	return options
}

// Bind binds the options variable flags with `db` prefix for the default database connection
func (options *Options) Bind() *Options {
	return options.BindWithPrefix("db")
}

// BindWithPrefix binds the options variable flags with a custom prefix for multiple database connections
func (options *Options) BindWithPrefix(prefix string) *Options {
	p := func(s string) string {
		if prefix != "" {
			return prefix + "-" + s
		}
		return s
	}
	cli.StringVar(&options.Credentials.DSN, p("dsn"), "", "DSN for database connection (mysql://, postgres://, sqlite://, or driver-specific format)")
	return options
}
