package gen

import (
	"os"

	"github.com/pkg/errors"

	"github.com/go-bridget/mig/cmd/mig/gen/golang"
	"github.com/go-bridget/mig/cmd/mig/gen/php81"
	"github.com/go-bridget/mig/cmd/mig/internal"
)

func render(language string, schema string, basePath string, ns string, tables []*internal.Table) error {
	languages := []string{
		"go",
		"php81",
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
		return golang.Render(basePath, schema, tables)
	case "php81":
		return php81.Render(basePath, schema, ns, tables)
	}
	return nil
}
