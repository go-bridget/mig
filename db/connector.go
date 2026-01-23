package db

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// ConnectWithRetry uses retry options set in Options{}.
func ConnectWithRetry(ctx context.Context, options *Options) (db *sqlx.DB, err error) {
	// by default, retry for 5 minutes, 5 seconds between retries
	if options.Retries == 0 && options.ConnectTimeout.Seconds() == 0 {
		options.ConnectTimeout = 5 * time.Minute
		options.RetryDelay = 5 * time.Second
	}

	connErrCh := make(chan error, 1)
	defer close(connErrCh)

	go func() {
		try := 0
		for {
			try++
			if options.Retries > 0 && options.Retries <= try {
				err = errors.Errorf("could not connect, tries=%d", try)
				break
			}

			db, err = ConnectWithOptions(ctx, options)
			if err != nil {
				log.Printf("can't connect, err=%s, try=%d", err, try)

				if errors.Is(err, ErrEmptyDSN) {
					break
				}

				select {
				case <-ctx.Done():
					break
				case <-time.After(options.RetryDelay):
					continue
				}
			}
			break
		}
		connErrCh <- err
	}()

	select {
	case err = <-connErrCh:
		break
	case <-time.After(options.ConnectTimeout):
		return nil, errors.Errorf("db connect timed out")
	case <-ctx.Done():
		return nil, errors.Errorf("db connection cancelled")
	}

	return db, err
}
