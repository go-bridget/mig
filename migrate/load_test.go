package migrate

import (
	"testing"
)

func TestLoad(t *testing.T) {
	assert := func(ok bool, msg string, args ...interface{}) {
		if !ok {
			t.Fatalf(msg, args...)
		}
	}

	options := &Options{
		Path:    "../testdata/schema/stats",
		Project: "stats",
	}

	err := Load(options)
	assert(err == nil, "Expected error nil, got %+v", err)

	assert(len(migrations) >= 1, "Expected len(migrations)>=1, got %d", len(migrations))

	stats, ok := migrations["stats"]
	assert(ok, "Expected 'stats' key exists in migrations")
	assert(len(stats) >= 1, "Expected len(stats)>=1, got %d", len(stats))
}
