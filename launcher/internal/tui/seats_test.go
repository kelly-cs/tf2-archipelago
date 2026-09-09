package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/* From the 1.9.0 play-test. The seats were stored as the classes somebody
 * named, with the draws left out, so naming seat 4 stored one class and the mod
 * read it as seat 1.
 *
 * The rows are declared in internal/form now, so this drives the row by the ID
 * form gave it rather than calling a method the terminal happened to expose.
 * That is the path a player takes and the path the window takes too.
 */
func TestSeatKeepsItsNumber(t *testing.T) {
	f := newSettingsForm(settings.Settings{}, settingsDeps{})

	// Seat 4 plays Engineer, the rest are the mod's to draw.
	set(t, f, "bots.seat.3.class", "engineer")

	if got := f.state.Settings.SrcdsBotTeamComp; len(got) != 4 || got[3] != "engineer" {
		t.Fatalf("team = %q, want three draws and then the engineer", got)
	}
	if got := botloadout.Composition(f.state.Settings.SrcdsBotTeamComp, nil); got != ",,,engineer" {
		t.Errorf("composition = %q", got)
	}

	// And the page still shows it where it was put.
	if got := value(t, f, "bots.seat.3.class"); got != "engineer" {
		t.Errorf("seat 4 shows %q", got)
	}
}

// A seat put back on the draw does not drag the seats after it up one.
func TestClearingASeatLeavesAHole(t *testing.T) {
	f := newSettingsForm(settings.Settings{
		SrcdsBotTeamComp:     []string{"engineer", "medic", "heavyweapons"},
		SrcdsBotSeatLoadouts: []string{"ranger", "kritz", "brass"},
	}, settingsDeps{})

	set(t, f, "bots.seat.1.class", "")

	if got := strings.Join(f.state.Settings.SrcdsBotTeamComp, ","); got != "engineer,,heavyweapons" {
		t.Errorf("team = %q", got)
	}
	if got := strings.Join(f.state.Settings.SrcdsBotSeatLoadouts, ","); got != "ranger,kritz,brass" {
		t.Errorf("seat loadouts = %q", got)
	}
	if got := value(t, f, "bots.seat.2.loadout"); got != "brass" {
		t.Errorf("seat 3 carries %q", got)
	}
}

/*
A team of nothing but draws wrote an empty composition, and the mod then played
the map's default lineup, which beats the blacklist.

The untick goes through the row's own key handling, the way a player does,
rather than writing the blacklist directly.
*/
func TestUntickedClassSurvivesATeamOfDraws(t *testing.T) {
	f := newSettingsForm(settings.Settings{}, settingsDeps{})

	row := find(t, f, "bots.class.sniper.allowed")
	if !row.Handle(tea.KeyMsg{Type: tea.KeyLeft}) {
		t.Fatal("the sniper toggle ignored the key")
	}

	if got := strings.Join(f.state.Settings.SrcdsBotClassBlacklist, ","); got != "sniper" {
		t.Fatalf("blacklist = %q", got)
	}
	comp := botloadout.Composition(f.state.Settings.SrcdsBotTeamComp, f.state.Settings.SrcdsBotClassBlacklist)
	if comp != ",,,,," {
		t.Errorf("composition = %q, want one hole per seat", comp)
	}
}

// find is the row with that ID, on whichever page carries it.
func find(t *testing.T, f *settingsForm, id string) *modelRow {
	t.Helper()
	for _, tab := range f.tabs {
		for _, row := range tab.fields {
			if row.f.ID == id {
				return row
			}
		}
	}
	t.Fatalf("no row %q on any page", id)
	return nil
}

// set writes a row the way any interface does, through form.
func set(t *testing.T, f *settingsForm, id, value string) {
	t.Helper()
	if err := f.applyChange(change(id, value)); err != nil {
		t.Fatalf("setting %s to %q: %v", id, value, err)
	}
}

func value(t *testing.T, f *settingsForm, id string) string {
	t.Helper()
	return find(t, f, id).f.Value
}
