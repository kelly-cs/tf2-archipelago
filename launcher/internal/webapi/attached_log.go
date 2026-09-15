package webapi

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
)

const (
	attachedLogPoll         = 250 * time.Millisecond
	attachedLogHistoryBytes = 8 << 20
)

// WatchAttachedLog follows the -condebug log written by a Compose-managed
// SRCDS. The game volume is mounted read-only into the admin sidecar: the page
// gets the useful console without giving the web process Docker-host control.
func (a *App) WatchAttachedLog(log AttachedLog) {
	path := log.Path
	if path == "" {
		return
	}
	source := log.Source
	if source == "" {
		source = "server"
	}
	var offset int64
	initialized := false
	read := func() {
		offset, initialized = readAttachedLog(path, source, offset, initialized, a.append)
	}
	read()
	ticker := time.NewTicker(attachedLogPoll)
	defer ticker.Stop()
	for {
		select {
		case <-a.quit:
			return
		case <-ticker.C:
			read()
		}
	}
}

// readAttachedLog reads only complete lines. A line being written at the instant
// this runs is left at the current offset and emitted whole on the next pass.
func readAttachedLog(path, source string, offset int64, initialized bool, sink func(apruntime.Line)) (int64, bool) {
	file, err := os.Open(path)
	if err != nil {
		return offset, initialized
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return offset, initialized
	}
	if info.Size() < offset {
		offset = 0
	}
	if !initialized {
		initialized = true
		if info.Size() > attachedLogHistoryBytes {
			offset = info.Size() - attachedLogHistoryBytes
			if _, err := file.Seek(offset, io.SeekStart); err != nil {
				return 0, initialized
			}
			fragment, _ := bufio.NewReader(file).ReadString('\n')
			offset += int64(len(fragment))
		}
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return offset, initialized
	}

	reader := bufio.NewReader(file)
	for {
		start := offset
		line, readErr := reader.ReadString('\n')
		if errors.Is(readErr, io.EOF) && !strings.HasSuffix(line, "\n") {
			return start, initialized
		}
		offset += int64(len(line))
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line != "" {
			sink(apruntime.Line{At: time.Now(), Source: source, Text: line})
		}
		if readErr != nil {
			return offset, initialized
		}
	}
}
