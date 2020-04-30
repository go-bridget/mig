package grants

import (
	"context"
	"log"

	"github.com/go-bridget/mig/cli"
	"github.com/go-bridget/mig/db"
)

const Name = "Create schemas, users and grant permissions"

func New() *cli.Command {
	var config struct {
		db db.Options
	}

	return &cli.Command{
		Bind: func(_ context.Context) {
			(&config.db).Init().Bind()
		},
		Run: func(ctx context.Context, commands []string) error {
			log.Println("commands", commands)
			return nil
		},
	}
}
