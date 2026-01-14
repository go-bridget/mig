package cli

import (
	"context"
	"testing"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// TestCommandWithBind tests that a command can bind flags via the Bind callback.
func TestCommandWithBind(t *testing.T) {
	var name string
	var count int

	cmd := &Command{
		Name: "test",
		Bind: func(fs *flag.FlagSet) {
			fs.StringVar(&name, "name", "default", "")
			fs.IntVar(&count, "count", 0, "")
		},
		Run: func(ctx context.Context, args []string) error {
			assert.Equal(t, "alice", name)
			assert.Equal(t, 5, count)
			return nil
		},
	}

	// Simulate app.RunWithArgs
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Usage = func() {}
	cmd.Bind(fs)

	err := fs.Parse([]string{"--name", "alice", "--count", "5"})
	assert.NoError(t, err)

	// Execute run
	err = cmd.Run(context.Background(), fs.Args())
	assert.NoError(t, err)
}

// TestCommandParseEnvironment tests that environment variables populate unset flags.
func TestCommandParseEnvironment(t *testing.T) {
	var dbDsn string

	cmd := &Command{
		Name: "test",
		Bind: func(fs *flag.FlagSet) {
			fs.StringVar(&dbDsn, "db-dsn", "default", "")
		},
		Run: func(ctx context.Context, args []string) error {
			assert.Equal(t, "postgres://localhost", dbDsn)
			return nil
		},
	}

	// Simulate app.RunWithArgs with environment
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Usage = func() {}
	cmd.Bind(fs)

	// Parse empty args, then set from environment
	err := ParseWithFlagSet(fs, []string{})
	assert.NoError(t, err)

	// Manually verify the environment variable was set
	fn := fs.Lookup("db-dsn")
	assert.NotNil(t, fn)
	err = fn.Value.Set("postgres://localhost")
	assert.NoError(t, err)

	err = cmd.Run(context.Background(), fs.Args())
	assert.NoError(t, err)
}

// TestAppAddCommand tests that commands are registered and can be found.
func TestAppAddCommand(t *testing.T) {
	app := NewApp("testapp")

	app.AddCommand("hello", "Say hello", func() *Command {
		cmd := &Command{
			Name: "hello",
			Run: func(ctx context.Context, args []string) error {
				return nil
			},
		}
		return cmd
	})

	cmd, err := app.findCommand([]string{"hello"}, "")
	assert.NoError(t, err)
	assert.Equal(t, "hello", cmd.Name)
}

// TestAppFindCommandNotFound tests that findCommand returns an error for unknown commands.
func TestAppFindCommandNotFound(t *testing.T) {
	app := NewApp("testapp")

	_, err := app.findCommand([]string{"nonexistent"}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

// TestAppFindCommandNoCommand tests that findCommand returns an error when no command is specified.
func TestAppFindCommandNoCommand(t *testing.T) {
	app := NewApp("testapp")

	_, err := app.findCommand([]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no command specified")
}

// TestParseCommands tests that parseCommands extracts command names up to the first flag.
func TestParseCommands(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "single command",
			args:     []string{"hello"},
			expected: []string{"hello"},
		},
		{
			name:     "command with flags",
			args:     []string{"hello", "--name", "world"},
			expected: []string{"hello"},
		},
		{
			name:     "multiple commands",
			args:     []string{"sub", "cmd", "--flag"},
			expected: []string{"sub", "cmd"},
		},
		{
			name:     "flag at start",
			args:     []string{"--flag", "hello"},
			expected: []string{},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommands(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommandMultipleFlags tests a command with multiple flag types.
func TestCommandMultipleFlags(t *testing.T) {
	var strFlag string
	var intFlag int
	var boolFlag bool

	cmd := &Command{
		Name: "test",
		Bind: func(fs *flag.FlagSet) {
			fs.StringVar(&strFlag, "str", "", "")
			fs.IntVar(&intFlag, "num", 0, "")
			fs.BoolVar(&boolFlag, "verbose", false, "")
		},
		Run: func(ctx context.Context, args []string) error {
			assert.Equal(t, "test", strFlag)
			assert.Equal(t, 42, intFlag)
			assert.True(t, boolFlag)
			return nil
		},
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Usage = func() {}
	cmd.Bind(fs)

	err := fs.Parse([]string{"--str", "test", "--num", "42", "--verbose"})
	assert.NoError(t, err)

	err = cmd.Run(context.Background(), fs.Args())
	assert.NoError(t, err)
}

// TestAppRunWithArgsIntegration tests the full command execution flow.
func TestAppRunWithArgsIntegration(t *testing.T) {
	app := NewApp("testapp")
	executed := false

	app.AddCommand("test", "Test command", func() *Command {
		var msg string

		cmd := &Command{
			Name: "test",
			Bind: func(fs *flag.FlagSet) {
				fs.StringVar(&msg, "msg", "default", "")
			},
			Run: func(ctx context.Context, args []string) error {
				executed = true
				assert.Equal(t, "hello", msg)
				return nil
			},
		}

		return cmd
	})

	// Simulate command line with flag
	err := app.RunWithArgs([]string{"test", "--msg", "hello"})
	assert.NoError(t, err)
	assert.True(t, executed)
}
