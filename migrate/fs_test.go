package migrate_test

import (
	"testing"

	"github.com/go-bridget/mig/migrate"
	"github.com/stretchr/testify/assert"
)

func TestFS(t *testing.T) {
	var dummy = []byte(`-- This is a comment`)

	fs := migrate.FS{}
	fs["foo.txt"] = dummy
	fs["file-1.up.sql"] = dummy
	fs["file-2.up.sql"] = dummy
	fs["folder/file2.sql"] = dummy

	got := fs.Migrations()
	want := []string{
		"file-1.up.sql",
		"file-2.up.sql",
	}

	assert.Equal(t, want, got)
}
