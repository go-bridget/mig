package gen

import (
	"os"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cmd/mig/gen/golang"
	"github.com/go-bridget/mig/cmd/mig/gen/model"
	"github.com/go-bridget/mig/cmd/mig/gen/php81"
	"github.com/go-bridget/mig/cmd/mig/internal"
)

func render(options model.Options, tables []*internal.Table) error {
	language := options.Language
	languages := []string{
		"go",
		"php81",
	}
	if !internal.Contains(languages, language) {
		return errors.Errorf("invalid language: %s", language)
	}

	// create output folder
	if err := os.MkdirAll(options.Output, 0755); err != nil {
		return err
	}

	switch language {
	case "go":
		return golang.Render(options, tables)
	case "php81":
		return php81.Render(options, tables)
	}
	return nil
}
