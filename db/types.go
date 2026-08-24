package db

import (
	"context"
	"os"
	"time"

	"database/sql"

	flag "github.com/spf13/pflag"

	"github.com/go-bridget/mig/logger"
)

// Options include database connection options
type Options struct {
	Credentials Credentials

	// Connector is an optional parameter to produce our
	// own *sql.DB, which is then wrapped in *sqlx.DB
	Connector func(context.Context, Credentials) (*sql.DB, error)

	Retries        int
	RetryDelay     time.Duration
	ConnectTimeout time.Duration

	// LogFn receives connection progress. It must be set; entry points
	// return logger.ErrNoLogFn when it is nil.
	LogFn logger.LogFn
}

// Logf writes a line to the bound sink.
func (options *Options) Logf(format string, args ...any) {
	options.LogFn(format, args...)
}

// checkLogFn reports whether a logging sink has been bound.
func (options *Options) checkLogFn() error {
	if options.LogFn == nil {
		return logger.ErrNoLogFn
	}
	return nil
}

// NewOptions provides an initialized *Options object.
func NewOptions() *Options {
	return &Options{
		Retries:        100,
		RetryDelay:     2 * time.Second,
		ConnectTimeout: 2 * time.Minute,
		Credentials: Credentials{
			DSN: os.Getenv("MIG_DB_DSN"),
		},
	}
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
	fs.StringVar(&options.Credentials.DSN, p("dsn"), options.Credentials.DSN, "DSN for database connection (mysql://, postgres://, sqlite://, or driver-specific format)")
	return options
}
