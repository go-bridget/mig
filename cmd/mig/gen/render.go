package gen

import (
	"os"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cmd/mig/internal"
)

func render(language string, schema string, basePath string, tables []*internal.Table) error {
	languages := []string{
		"go",
	}
	if !internal.Contains(languages, language) {
		return errors.Errorf("invalid language: %s", language)
	}

	// create output folder
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return err
	}

	switch language {
	case "go":
		if err := renderGo(basePath, schema, tables); err != nil {
			return err
		}
	}
	return nil
}
