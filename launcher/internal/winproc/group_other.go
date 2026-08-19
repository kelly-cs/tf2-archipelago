//go:build !windows

package winproc

import (
	"os/exec"
	"syscall"
	"time"
)

// groupGrace is how long a signalled process group has to exit before it is
// killed outright, and how long Wait blocks on pipes the group still holds.
const groupGrace = 5 * time.Second

/*
KillGroup makes a cancelled command take its children down with it.

srcds_run is a shell script that runs srcds_linux as a child, so killing the
command kills the script and leaves the game server running: it keeps the
game port and the rcon port, and the next Start fails to bind either. Nothing
in the launcher notices, because the process it was watching did exit.

Setpgid puts the script and everything it starts in one group. Cancel then
signals the whole group rather than the script alone, and WaitDelay ends the
wait for a group that ignores the signal or holds the output pipes open.

What WaitDelay does at the end of that wait is kill the process, and the
process is the script. A game server that sat on the SIGTERM would be left
running by exactly the path this exists to close, so the group is killed here
too, a grace later, whatever the script did with the first signal.
*/
func KillGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		group := -cmd.Process.Pid
		// A group that has already gone gives ESRCH, which is the answer this
		// wanted: nothing is left to kill.
		time.AfterFunc(groupGrace, func() { _ = syscall.Kill(group, syscall.SIGKILL) })
		return syscall.Kill(group, syscall.SIGTERM)
	}
	cmd.WaitDelay = groupGrace
}
