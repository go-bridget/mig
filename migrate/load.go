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
func Load(options *Options) error {
	if options.Filename != "" {
		filename := options.Filename
		project := options.Project

		base := filepath.Base(filename)
		contents, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		migrations[project] = NewFS()
		migrations[project][base] = contents
		return nil
	}

	if err := assertDir(options.Path); err != nil {
		return err
	}

	project := filepath.Base(options.Path)
	files, err := filepath.Glob(path.Join(options.Path, "*.sql"))
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

	return nil
}
