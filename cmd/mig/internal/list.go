package internal

import (
	"context"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/db"
	"github.com/go-bridget/mig/db/introspect"
	"github.com/go-bridget/mig/model"
)

// ListTables retrieves all tables from the database.
func ListTables(ctx context.Context, config *db.Options) ([]*model.Table, error) {
	handle, err := db.ConnectWithRetry(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}

	desc, err := introspect.NewDescriber(handle)
	if err != nil {
		return nil, err
	}

	return desc.ListTables(ctx, handle)
}
