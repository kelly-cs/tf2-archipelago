package tui

import (
	"os"

	"golang.org/x/term"
)

// Available reports whether there is a terminal to draw on. A service, a CI
// job and a launcher whose output is being piped into a file all have none,
// and an interface that draws over the whole screen writes nothing useful into
// any of them: the caller falls back to printing the log.
func Available() bool {
	return term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd()))
}
