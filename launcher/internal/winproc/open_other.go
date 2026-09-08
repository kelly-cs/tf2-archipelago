//go:build !windows

package winproc

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// OpenURL hands a link to the desktop's own handler, which is the same
// program that opens a file here.
func OpenURL(link string) error { return Open(link) }

// Open shows a file or a folder to the player, through the desktop's own
// handler.
func Open(path string) error {
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	if err := exec.CommandContext(context.Background(), opener, path).Start(); err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}
	return nil
}
