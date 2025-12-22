package gen

import (
	"os"
	"slices"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cmd/mig/gen/golang"
	"github.com/go-bridget/mig/cmd/mig/gen/model"
	migmodel "github.com/go-bridget/mig/model"
)

func render(options model.Options, tables []*migmodel.Table) error {
	language := options.Language
	languages := []string{
		"go",
	}
	if !slices.Contains(languages, language) {
		return errors.Errorf("invalid language: %s", language)
	}

	// create output folder
	if err := os.MkdirAll(options.Output, 0755); err != nil {
		return err
	}

	switch language {
	case "go":
		return golang.Render(options, tables)
	}
	return nil
}
