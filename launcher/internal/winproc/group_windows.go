//go:build windows

package winproc

import "os/exec"

// KillGroup does nothing here. srcds.exe is the process the launcher starts,
// with no script between the two, and what it leaves behind is KillUnder's
// job. Setting SysProcAttr would also undo HideConsole.
func KillGroup(*exec.Cmd) {}
