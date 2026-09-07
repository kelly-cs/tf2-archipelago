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
TestAFreshInstallSavesAndIsToldWhereToGetARoom below is that case on purpose; every
other test here wants to get past it.
*/
func withRoom() settings.Settings {
	s := settings.Defaults()
	s.APHost, s.APPort, s.APTls = "archipelago.gg", 38281, true
	return s
}

/*
	A fresh install saves, and is told where to get a room.

This reverses what it used to do, on purpose. An unparseable room refused the
whole Save, which is how a player lost a login token they were setting two pages
away: one field they could not see blocked every other answer on the screen.
Their config.json held the defaults for the room, ap_port 0, so every Save they
pressed was refused on an address they had not typed yet.

A room they have not made is not a mistake. The settings are written, and the
room is what they are told about afterwards.
*/
func TestAFreshInstallSavesAndIsToldWhereToGetARoom(t *testing.T) {
	var written bool
	f := newSettingsForm(settings.Defaults(), settingsDeps{
		persist: func(s settings.Settings) (settings.Settings, error) {
			written = true
			if s.APPort != 0 {
				t.Errorf("a run with no room saved port %d", s.APPort)
			}
			return s, nil
		},
		saved: func(settings.Settings) tea.Cmd { return nil },
	})

	// On another page, the way the reporter was.
	f.showTab("Game server")
	cmd := f.save()

	if !written {
		t.Fatal("a save with no room address wrote nothing")
	}
	if !f.closed {
		t.Error("the screen stayed open on a save that worked")
	}
	if f.problem != "" {
		t.Errorf("a save that worked put up a problem box: %q", f.problem)
	}
	if cmd == nil {
		t.Fatal("the save said nothing about the missing room")
	}

	notice, ok := cmd().(noticeMsg)
	if !ok {
		t.Fatalf("the save reported a %T", cmd())
	}
	for _, want := range []string{"saved", "archipelago.gg", "Test mode"} {
		if !strings.Contains(string(notice), want) {
			t.Errorf("the notice does not mention %q: %q", want, notice)
		}
	}
}

/*
	The install folder is on a page, and a Save refuses one it cannot use.

The other half of the same report. The player wanted their folder somewhere
other than C:\Users\<name>\tf2-archipelago and could not find a setting for it,
because there was not one: it was chosen by the installer screen and never
offered again. It is a row now, so it can also be typed wrong, and the two ways
that matter are refused where they can be seen rather than at the MkdirAll that
would fail hours later.
*/
func TestTheInstallFolderIsARowAndAnUnusableOneIsRefused(t *testing.T) {
	f := newSettingsForm(withRoom(), settingsDeps{
		persist: settings.Persist,
		saved:   func(settings.Settings) tea.Cmd { return nil },
	})

	row := find(t, f, "run.install_root")
	if row.f.Value == "" {
		t.Error("the install folder row shows nothing, so nobody can see where it is")
	}
	if !row.f.Browse {
		t.Error("the install folder is a folder and offers no way to pick one")
	}

	set(t, f, "run.install_root", "")
	if cmd := f.save(); cmd != nil {
		t.Error("a save with no install folder handed work to the event loop")
	}
	if f.closed {
		t.Error("the screen closed on a save with no install folder")
	}
	if !strings.Contains(f.problem, "install folder") {
		t.Errorf("the problem box says %q", f.problem)
	}
}

// The settings file is somewhere nobody would guess, so there is a button that
// shows it. "Where is the config file?" is a quote from the report.
func TestThereIsAWayToFindTheSettingsFile(t *testing.T) {
	f := newSettingsForm(withRoom(), settingsDeps{
		persist: settings.Persist,
		saved:   func(settings.Settings) tea.Cmd { return nil },
	})

	row := find(t, f, "run.open_settings_file")
	if f.action(row.f.ID) == nil {
		t.Error("the button that shows the settings file does nothing")
	}
	if !strings.Contains(row.f.Help, "not in the install folder") {
		t.Errorf("the help does not correct the mistake it exists for: %q", row.f.Help)
	}
}
