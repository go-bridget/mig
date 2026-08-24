package db

import (
	"context"
	"os"
	"time"

	"database/sql"

	flag "github.com/spf13/pflag"
)

// Logger receives what a connection did. The shape is slog's, so a
// *slog.Logger satisfies it as it stands, but any logger with these two
// methods does - the package doesn't name slog.
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

// Options include database connection options
type Options struct {
	Credentials Credentials

	// Connector is an optional parameter to produce our
	// own *sql.DB, which is then wrapped in *sqlx.DB
	Connector func(context.Context, Credentials) (*sql.DB, error)

	Retries        int
	RetryDelay     time.Duration
	ConnectTimeout time.Duration

	// Logger receives the progress of a connection, which is the retries
	// it took. It's required - take an Options from NewOptions and it's
	// the argument you passed.
	Logger Logger
}

// NewOptions provides an initialized *Options object reporting through log.
func NewOptions(log Logger) *Options {
	return &Options{
		Retries:        100,
		RetryDelay:     2 * time.Second,
		ConnectTimeout: 2 * time.Minute,
		Credentials: Credentials{
			DSN: os.Getenv("MIG_DB_DSN"),
		},
		Logger: log,
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
