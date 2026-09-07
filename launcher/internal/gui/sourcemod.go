package gui

/*
	sourcemodUpdate is the launcher's memory of the one restart SourceMod's

updater gets.

The updater fetches gamedata whenever it decides to, then asks in the server's
log for a restart to load it. Taking that request the moment it lands drops a
server restart into the middle of whatever the player is doing, and the settings
window is the one place where that reads as something they did: apw-ylq was
reported as a bot team change restarting the server, and the repro was waiting
in the window rather than anything about the team.

So a request that arrives with the window open waits for it to close, and a save
that is going to restart anyway drops it, because that restart loads the new
gamedata too.
*/
type sourcemodUpdate struct {
	// asked is whether the updater has spoken this run. It never clears: a
	// second request means the first restart did not settle it, and restarting
	// again would be a loop with a 14 GB game server in it.
	asked bool

	// held is a request that landed while the settings window was up and is
	// still owed a restart.
	held bool
}

// ask records the updater's request. restart is for a request to act on now,
// hold for one the settings window is keeping until it closes; a request after
// the first is neither.
func (u *sourcemodUpdate) ask(settingsOpen bool) (restart, hold bool) {
	if u.asked {
		return false, false
	}
	u.asked = true
	u.held = settingsOpen
	return !settingsOpen, settingsOpen
}

// take hands back a held request, once.
func (u *sourcemodUpdate) take() bool {
	held := u.held
	u.held = false
	return held
}

// drop forgets a held request, for a caller whose own restart loads the
// gamedata the updater wrote.
func (u *sourcemodUpdate) drop() { u.held = false }
