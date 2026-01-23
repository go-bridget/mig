package db

import (
	"context"
	"database/sql"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var ErrEmptyDSN = errors.New("empty dsn")

// Connect connects to a database and produces the handle for injection.
func Connect(ctx context.Context) (*sqlx.DB, error) {
	options := &Options{
		Connector: func(ctx context.Context, credentials Credentials) (*sql.DB, error) {
			driver, dsn := credentials.Open()
			db, err := sql.Open(driver, dsn)
			if err != nil {
				return nil, err
			}
			if err = db.PingContext(ctx); err != nil {
				db.Close()
				return nil, err
			}
			return db, nil
		},
	}
	options.Credentials.DSN = os.Getenv("MIG_DB_DSN")
	return ConnectWithRetry(ctx, options)
}

// ConnectWithOptions connects to host based on Options{}.
func ConnectWithOptions(ctx context.Context, options *Options) (*sqlx.DB, error) {
	driver, dsn := options.Credentials.Open()
	if dsn == "" {
		return nil, ErrEmptyDSN
	}

	connect := func() (*sqlx.DB, error) {
		if options.Connector != nil {
			handle, err := options.Connector(ctx, options.Credentials)
			if err == nil {
				return sqlx.NewDb(handle, driver), nil
			}
			return nil, errors.WithStack(err)
		}
		return sqlx.ConnectContext(ctx, driver, dsn)
	}

	db, err := connect()
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(800)
	db.SetMaxIdleConns(800)
	return db, nil
}
