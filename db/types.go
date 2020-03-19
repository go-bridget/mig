package db

import (
	"context"
	"time"

	"database/sql"
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
