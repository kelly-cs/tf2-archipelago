package gui

import "testing"

/*
apw-ylq: the same bot team change restarted the server or did not, depending
only on how long the settings window had been open.

Nothing on the save path is time-dependent. SourceMod's updater is: it speaks
when it has fetched new gamedata, and the restart it asks for used to land the
moment it did, which for a player mid-edit is their edit restarting the server.
*/
func TestARequestDuringTheSettingsWindowWaitsForIt(t *testing.T) {
	var u sourcemodUpdate

	restart, hold := u.ask(true)
	if restart || !hold {
		t.Fatalf("ask with the window open = (restart %v, hold %v)", restart, hold)
	}
	if !u.take() {
		t.Fatal("the window closed and the held restart was gone")
	}
	if u.take() {
		t.Fatal("the held restart came back a second time")
	}
}

// With no window in the way the request is the restart, as it always was.
func TestARequestWithNoWindowRestartsAtOnce(t *testing.T) {
	var u sourcemodUpdate

	restart, hold := u.ask(false)
	if !restart || hold {
		t.Fatalf("ask with no window = (restart %v, hold %v)", restart, hold)
	}
	if u.take() {
		t.Fatal("a restart already taken was held as well")
	}
}

// The updater only speaks when it found something new, so a second request
// means the first restart did not take. Restarting again is a loop with a
// 14 GB game server in it.
func TestOnlyTheFirstRequestIsEverActedOn(t *testing.T) {
	var u sourcemodUpdate
	u.ask(false)

	if restart, hold := u.ask(false); restart || hold {
		t.Fatalf("second ask = (restart %v, hold %v)", restart, hold)
	}
	if restart, hold := u.ask(true); restart || hold {
		t.Fatalf("second ask with the window open = (restart %v, hold %v)", restart, hold)
	}
}

// A save that restarts the server for its own reasons loads the new gamedata
// on the way, so the held request has nothing left to do.
func TestASaveThatRestartsDropsTheHeldRequest(t *testing.T) {
	var u sourcemodUpdate
	u.ask(true)
	u.drop()

	if u.take() {
		t.Fatal("the settings save restarted and SourceMod restarted again after it")
	}
}
