package migrate

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

func assertDir(location string) error {
	stat, err := os.Stat(location)
	if err != nil {
		return errors.Wrapf(err, "path: '%s'", location)
	}
	if stat.IsDir() {
		return nil
	}
	return errors.Errorf("path is not a directory: '%s'", location)
}

// Load reads migrations from disk
func Load(options Options) error {
	if err := assertDir(options.Path); err != nil {
		return err
	}

	source := path.Join(options.Path, "*", migrationsFile)
	matches, err := filepath.Glob(source)
	if err != nil {
		return err
	}

	for _, match := range matches {
		location := filepath.Dir(match)
		project := filepath.Base(location)
		files, err := filepath.Glob(path.Join(location, "*.sql"))
		if err != nil {
			return err
		}

		migrations[project] = NewFS()
		for _, filename := range files {
			base := filepath.Base(filename)
			contents, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}
			migrations[project][base] = contents
		}
	}
	return nil
}
