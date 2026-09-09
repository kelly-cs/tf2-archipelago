/*
Package browser opens the launcher's own page in the player's browser.

Three desktops, and the awkward one in the middle:

  - Windows: the shell's own handler, through winproc.
  - Linux: xdg-open.
  - WSL: a Linux userland with a Windows desktop. xdg-open there either does
    nothing or opens a browser inside the distribution that the player cannot
    see, so the Windows browser is what gets asked. Windows reaches the
    listener on 127.0.0.1 through WSL's localhost forwarding, so the URL needs
    no rewriting.

Nothing here is fatal. A desktop with no opener is normal over SSH, in a
container and on a minimal install, and the URL still works: Open answers with
the error and the caller prints the address.
*/
package browser

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/m-this/tf2-archipelago/launcher/internal/winproc"
)

// openGrace bounds the wait for the opener to start. It is not the wait for a
// browser to appear: every one of these returns as soon as the request is
// handed over.
const openGrace = 10 * time.Second

// Open shows the URL to the player.
func Open(url string) error {
	if runtime.GOOS == "windows" {
		return winproc.OpenURL(url)
	}
	// WSL_DISTRO_NAME is the operating system saying what it is, not
	// configuration: it is read here because here is where it means something.
	//nolint:forbidigo // The OS owns this one, and it is read where it is used.
	if UnderWSL(os.Getenv("WSL_DISTRO_NAME"), procVersion()) {
		return openThroughWindows(url)
	}
	return start(url, "xdg-open")
}

/*
UnderWSL reports whether this Linux is one Windows is hosting.

Two answers rather than one because neither is always there: WSL_DISTRO_NAME is
unset for a process the init system started rather than the shell, and
/proc/version is unreadable in some containers. Either one naming it is enough.
*/
func UnderWSL(distro, version string) bool {
	if distro != "" {
		return true
	}
	lower := strings.ToLower(version)
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

// openThroughWindows asks the Windows desktop, not this distribution's.
//
// wslview is the tidy way and comes with wslu, which is not installed
// everywhere. cmd.exe is always reachable from WSL, and its start builtin
// treats the first quoted argument as a window title, so the empty title is
// there on purpose: without it "start "http://..."" opens nothing and names a
// window after the URL.
func openThroughWindows(url string) error {
	if _, err := exec.LookPath("wslview"); err == nil {
		return start(url, "wslview")
	}
	return start(url, "cmd.exe", "/c", "start", "")
}

// start hands the URL over and returns. The timeout bounds the handover, not
// the browser: every one of these opens returns as soon as the request is
// taken, and a browser that takes ten seconds to appear is not this to wait on.
func start(url string, command ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), openGrace)
	defer cancel()
	arguments := append(append([]string{}, command[1:]...), url)
	if err := exec.CommandContext(ctx, command[0], arguments...).Start(); err != nil {
		return fmt.Errorf("cannot open a browser: %w", err)
	}
	return nil
}

func procVersion() string {
	version, err := os.ReadFile("/proc/version")
	if err != nil {
		return ""
	}
	return string(version)
}
