package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/*
	A save that cannot write the file keeps the screen and says so.

From a Discord report. The settings screen used to close first and hand the
settings to whoever opened it, which wrote them and logged a line when it could
not. By then the screen was gone and so was everything typed into it, and the
line was in a log the player had no reason to be reading. They read it as the
Save button doing nothing.

Three things have to hold, and all three are what that player did not get: the
screen stays, the answers stay, and the reason is put in front of them.
*/
func TestAFailedSaveKeepsTheScreenAndTheAnswers(t *testing.T) {
	refused := errors.New(`cannot write C:\Users\x\AppData\Roaming\tf2ap\config.json: access is denied`)

	var applied bool
	f := newSettingsForm(withRoom(), settingsDeps{
		persist: func(s settings.Settings) (settings.Settings, error) { return s, refused },
		saved:   func(settings.Settings) tea.Cmd { applied = true; return nil },
	})
	set(t, f, "rewards.traps", "42")

	if cmd := f.save(); cmd != nil {
		t.Error("a refused save handed work to the event loop")
	}
	if f.closed {
		t.Error("the screen closed on a save that wrote nothing")
	}
	if applied {
		t.Error("a refused save was applied anyway")
	}
	if f.state.Settings.MvmTrapPct != 42 {
		t.Errorf("the trap share is %d after a refused save, was 42", f.state.Settings.MvmTrapPct)
	}
	if !strings.Contains(f.problem, "access is denied") {
		t.Errorf("the problem box says %q", f.problem)
	}
	// The path is the fact the player needed: they went looking for a
	// config.json in the install root, and it is not there.
	if !strings.Contains(f.problem, "config.json") {
		t.Errorf("the problem box does not name the file it could not write: %q", f.problem)
	}

	view := f.view(100, 30)
	// Word by word: the box wraps, so a path longer than the screen is broken
	// across lines rather than truncated, and truncating it would hide the one
	// fact the player needs.
	for _, want := range []string{"The settings were not saved.", "config.json", "denied", "any key to go back"} {
		if !strings.Contains(view, want) {
			t.Errorf("the screen does not show %q:\n%s", want, view)
		}
	}
	// The rows are behind the box, not beside it.
	if strings.Contains(view, "Easiest tier") {
		t.Error("the problem box is drawn over nothing; the rows are still on screen")
	}
}

// Any key puts the rows back, and that key does nothing else: one meant for the
// box must not also toggle whatever row was under it.
func TestAKeyDismissesTheProblemAndNothingElse(t *testing.T) {
	f := newSettingsForm(withRoom(), settingsDeps{
		persist: func(s settings.Settings) (settings.Settings, error) {
			return s, errors.New("nope")
		},
		saved: func(settings.Settings) tea.Cmd { return nil },
	})
	f.save()
	if f.problem == "" {
		t.Fatal("the save was refused and nothing was shown")
	}

	if !f.dismiss() {
		t.Error("the first key did not take the box off")
	}
	if f.dismiss() {
		t.Error("the box came off twice")
	}
	if !strings.Contains(f.view(100, 30), "Easiest tier") {
		t.Error("the rows did not come back")
	}
}

// A save that works closes the screen and hands on what was written, which is
// not quite what it was given: Persist fills in an RCON password.
func TestASavedScreenHandsOnWhatWasWritten(t *testing.T) {
	var handed settings.Settings
	f := newSettingsForm(withRoom(), settingsDeps{
		persist: func(s settings.Settings) (settings.Settings, error) {
			s.SrcdsRconPw = "made-on-the-way-out"
			return s, nil
		},
		saved: func(s settings.Settings) tea.Cmd { handed = s; return nil },
	})

	f.save()
	if !f.closed {
		t.Error("a successful save left the screen open")
	}
	if handed.SrcdsRconPw != "made-on-the-way-out" {
		t.Errorf("what was handed on carries the password %q", handed.SrcdsRconPw)
	}
	if f.state.Settings.SrcdsRconPw != "made-on-the-way-out" {
		t.Error("the screen kept settings that differ from what was written")
	}
}

/*
	withRoom is the defaults with a room address, because the defaults have none.

settings.Defaults leaves APPort at 0 and Save refuses that, which is not a
detail: it is the bug a player reported. They set a login token, pressed Save,
and nothing happened, because the room two pages away had never been filled in.
TestAFreshInstallIsRefusedOnTheRoomPage below is that case on purpose; every
other test here wants to get past it.
*/
func withRoom() settings.Settings {
	s := settings.Defaults()
	s.APHost, s.APPort, s.APTls = "archipelago.gg", 38281, true
	return s
}

/*
	A fresh install refuses the Save, on the page holding the reason.

From a player's debug bundle. Their config.json held the defaults for the room,
APPort 0, so every Save they pressed was refused on an address they had not
typed yet. They were on another page setting a token at the time, the message
sat beside a field they could not see, and their log had no line about saving at
all. Three separate reasons to read it as the button doing nothing.
*/
func TestAFreshInstallIsRefusedOnTheRoomPage(t *testing.T) {
	f := newSettingsForm(settings.Defaults(), settingsDeps{
		persist: func(s settings.Settings) (settings.Settings, error) {
			t.Error("a save with no room address reached the file")
			return s, nil
		},
		saved: func(settings.Settings) tea.Cmd { return nil },
	})

	// On another page, the way the reporter was.
	f.showTab("Game server")

	if cmd := f.save(); cmd != nil {
		t.Error("a refused save handed work to the event loop")
	}
	if f.closed {
		t.Error("the screen closed on a save that wrote nothing")
	}
	if got := f.tabs[f.tab].title; got != "Archipelago room" {
		t.Errorf("the refusal left the player on %q, not the page holding the reason", got)
	}
	if !strings.Contains(f.problem, "Room address") {
		t.Errorf("the problem box says %q", f.problem)
	}
	if !strings.Contains(f.problem, "Test mode") {
		t.Error("the problem box does not offer the way out for somebody with no room yet")
	}
}
