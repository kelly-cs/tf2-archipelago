//go:build windows

package winproc

import (
	"context"
	"fmt"
	"os/exec"

	"golang.org/x/sys/windows"
)

// OpenURL hands a link to whatever Windows registered for its scheme. No
// fallback: a steam:// link nothing answers means Steam is not installed, and
// opening it in a text editor helps nobody.
func OpenURL(link string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	target, err := windows.UTF16PtrFromString(link)
	if err != nil {
		return err
	}
	const swShowNormal = 1
	if err := windows.ShellExecute(0, verb, target, nil, nil, swShowNormal); err != nil {
		return fmt.Errorf("cannot open %s: %w", link, err)
	}
	return nil
}

// Open shows a file or a folder to the player, in whatever program Windows
// uses for it.
//
// A .yaml has no program associated with it on a fresh Windows, and
// ShellExecute fails rather than asking. Notepad opens it in that case, which
// is what a player wants from a button called "open".
func Open(path string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	target, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	const swShowNormal = 1
	if err := windows.ShellExecute(0, verb, target, nil, nil, swShowNormal); err == nil {
		return nil
	}
	if err := exec.CommandContext(context.Background(), "notepad.exe", path).Start(); err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}
	return nil
}
