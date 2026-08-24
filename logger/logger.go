// Package logger provides the printf-style logging sink used across mig.
package logger

import (
	"fmt"

	"github.com/pkg/errors"
)

// LogFn is a printf-style logging sink. The sink terminates each line
// itself; format strings passed to it carry no trailing newline. The
// signature is assignable from log.Printf and testing.T.Logf.
type LogFn func(format string, args ...any)

// ErrNoLogFn is returned when an entry point is called with no sink bound.
var ErrNoLogFn = errors.New("no LogFn bound on Options")

// Printf writes a formatted line to stdout. This is the sink the mig CLI binds.
func Printf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// Discard drops all output. Bind it to silence a package explicitly.
func Discard(string, ...any) {}
