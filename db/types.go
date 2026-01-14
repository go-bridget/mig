package db

import (
	"context"
	"os"
	"time"

	"database/sql"

	flag "github.com/spf13/pflag"
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

// Bind registers database flags on the given FlagSet with `db` prefix.
func (options *Options) Bind(fs *flag.FlagSet) *Options {
	return options.BindWithPrefix(fs, "db")
}

// BindWithPrefix registers database flags on the given FlagSet with a custom prefix for multiple database connections.
func (options *Options) BindWithPrefix(fs *flag.FlagSet, prefix string) *Options {
	p := func(s string) string {
		if prefix != "" {
			return prefix + "-" + s
		}
		return s
	}
	fs.StringVar(&options.Credentials.DSN, p("dsn"), "", "DSN for database connection (mysql://, postgres://, sqlite://, or driver-specific format)")
	return options
}
