package internal

import (
	"context"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/db/introspect"
	"github.com/go-bridget/mig/model"
)

func ListTables(ctx context.Context, config *db.Options, schema string) ([]*model.Table, error) {
	handle, err := db.ConnectWithRetry(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}

	desc, err := introspect.NewDescriber(config.Credentials.Driver)
	if err != nil {
		return nil, err
	}

	return desc.ListTables(ctx, handle, schema)
}
