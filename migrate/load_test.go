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

	err := Load(Options{"../test/schema"})
	assert(err == nil, "Expected error nil, got %+v", err)

	assert(len(migrations) >= 1, "Expected len(migrations)>=1, got %d", len(migrations))

	val, ok := migrations["stats"]
	assert(ok, "Expected 'stats' key exists in migrations")
	assert(len(val) == 2, "Expected migration['stats'] length 2, got %d", len(val))
}
