package db

import (
	"context"
	"database/sql"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Connect connects to a database and produces the handle for injection
func Connect(ctx context.Context) (*sqlx.DB, error) {
	options := &Options{
		Connector: func(ctx context.Context, credentials Credentials) (*sql.DB, error) {
			db, err := sql.Open(credentials.Driver, credentials.DSN)
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
	options.Credentials.DSN = os.Getenv("DB_DSN")
	options.Credentials.Driver = os.Getenv("DB_DRIVER")
	return ConnectWithRetry(ctx, options)
}

// ConnectWithOptions connect to host based on Options{}
func ConnectWithOptions(ctx context.Context, options *Options) (*sqlx.DB, error) {
	credentials := options.Credentials
	if credentials.DSN == "" {
		return nil, errors.New("DSN not provided")
	}
	if credentials.Driver == "" {
		credentials.Driver = "mysql"
	}
	credentials.DSN = cleanDSN(credentials.DSN)

	connect := func() (*sqlx.DB, error) {
		if options.Connector != nil {
			handle, err := options.Connector(ctx, credentials)
			if err == nil {
				return sqlx.NewDb(handle, credentials.Driver), nil
			}
			return nil, errors.WithStack(err)
		}
		return sqlx.ConnectContext(ctx, credentials.Driver, credentials.DSN)
	}

	db, err := connect()
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(800)
	db.SetMaxIdleConns(800)
	return db, nil
}
