package runtime

import (
	"os"
	"path/filepath"
)

// The launcher's own log, and the run before it. Named here rather than at each
// use, because the debug bundle has to ask for the same two files.
const (
	LogFileName     = "launcher.log"
	LogPreviousName = "launcher-previous.log"
)

// The game server's console log, written because srcds is started with
// -condebug, and the run before it.
const (
	ConsoleLogName      = "console.log"
	ConsolePreviousName = "console-previous.log"
)

// CreateLogFile opens this run's log next to the game files, keeping the run
// before it.
//
// One run per file, because the interesting one is nearly always the last. The
// exception is the one that matters: a player who hits a bug restarts the
// server before asking anybody about it, and the restart used to truncate the
// evidence. So the previous run is kept, and the bundle carries both.
func CreateLogFile(installRoot string) (*os.File, error) {
	if err := os.MkdirAll(installRoot, 0o755); err != nil {
		return nil, err
	}
	current := filepath.Join(installRoot, LogFileName)
	rotate(current, filepath.Join(installRoot, LogPreviousName))
	return os.Create(current)
}

// RotateConsoleLog keeps the game server's own console log from the run before
// this one, and takes the file out of the way of the run that is starting.
//
// -condebug appends, so without this the file is every evening ever played
// concatenated, and the only thing that ever truncates it is a reinstall. The
// bundle takes the last 8 MB of whatever it finds, which on a file like that
// is the end of one evening and the start of nothing.
func RotateConsoleLog(gameDir string) {
	current := filepath.Join(gameDir, "tf", ConsoleLogName)
	rotate(current, filepath.Join(gameDir, "tf", ConsolePreviousName))
}

// rotate moves current out of the way, if it is there at all.
//
// A rename that fails is a log that is not kept, which is worth no more than a
// truncated one: the run that is starting still gets its file either way, so
// nothing here reports.
func rotate(current, previous string) {
	if _, err := os.Stat(current); err != nil {
		return
	}
	_ = os.Rename(current, previous)
}
