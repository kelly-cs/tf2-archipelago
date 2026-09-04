/*
Package guard turns a panic on a goroutine nobody waits on into a line.

The bridge runs its pump and its session on their own goroutines, and Go takes
the whole process down for a panic on any of them. The launcher hosts the
bridge, so that is the launcher closing and the server somebody is playing on
stopping with it.

Not a swallow. The panic and its stack go to the logger the caller already has,
which is louder than the silent exit it replaces.
*/
package guard

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

// Run does the work, and reports a panic rather than letting it end the process.
func Run(name string, logger *slog.Logger, work func()) {
	_ = Result(name, logger, func() error {
		work()
		return nil
	})
}

// Result runs work and turns a panic into an error after logging its stack.
// Callers that coordinate goroutines need the error: swallowing it would leave
// them waiting forever for a completion value the panicking goroutine never
// sent.
func Result(name string, logger *slog.Logger, work func() error) (err error) {
	defer func() {
		reason := recover()
		if reason == nil {
			return
		}
		if logger != nil {
			logger.Error("contained a panic",
				"where", name,
				"reason", fmt.Sprint(reason),
				"stack", string(debug.Stack()))
		}
		err = fmt.Errorf("%s panicked: %v", name, reason)
	}()
	return work()
}
