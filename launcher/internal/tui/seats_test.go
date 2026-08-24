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
 */
func TestSeatKeepsItsNumber(t *testing.T) {
	form := newSettingsForm(settings.Settings{}, settingsDeps{})

	// Seat 4 plays Engineer, the rest are the mod's to draw.
	form.setSeat(3, indexOfClass(t, "engineer")+1)

	if got := form.edited.SrcdsBotTeamComp; len(got) != 4 || got[3] != "engineer" {
		t.Fatalf("team = %q, want three draws and then the engineer", got)
	}
	if got := botloadout.Composition(form.edited.SrcdsBotTeamComp, nil); got != ",,,engineer" {
		t.Errorf("composition = %q", got)
	}

	// And the page still shows it where it was put.
	if got := form.seats()[3].Class; got != "engineer" {
		t.Errorf("seat 4 shows %q", got)
	}
}

// A seat put back on the draw does not drag the seats after it up one.
func TestClearingASeatLeavesAHole(t *testing.T) {
	form := newSettingsForm(settings.Settings{
		SrcdsBotTeamComp:     []string{"engineer", "medic", "heavyweapons"},
		SrcdsBotSeatLoadouts: []string{"ranger", "kritz", "brass"},
	}, settingsDeps{})

	form.setSeat(1, 0)

	if got := strings.Join(form.edited.SrcdsBotTeamComp, ","); got != "engineer,,heavyweapons" {
		t.Errorf("team = %q", got)
	}
	if got := strings.Join(form.edited.SrcdsBotSeatLoadouts, ","); got != "ranger,,brass" {
		t.Errorf("seat loadouts = %q", got)
	}
	if got := form.seats()[2].Loadout; got != "brass" {
		t.Errorf("seat 3 carries %q", got)
	}
}

func indexOfClass(t *testing.T, key string) int {
	t.Helper()
	for index, class := range botloadout.Classes {
		if class.Key == key {
			return index
		}
	}
	t.Fatalf("no class %q", key)
	return -1
}

// A team of nothing but draws wrote an empty composition, and the mod then
// played the map's default lineup, which beats the blacklist.
func TestUntickedClassSurvivesATeamOfDraws(t *testing.T) {
	form := newSettingsForm(settings.Settings{}, settingsDeps{})

	untick(t, form, "sniper")

	if got := strings.Join(form.edited.SrcdsBotClassBlacklist, ","); got != "sniper" {
		t.Fatalf("blacklist = %q", got)
	}
	if got := botloadout.Composition(form.edited.SrcdsBotTeamComp, form.edited.SrcdsBotClassBlacklist); got != ",,,,," {
		t.Errorf("composition = %q, want one hole per seat", got)
	}
}

// untick goes through the field's own key handling, the way a player does.
func untick(t *testing.T, form *settingsForm, key string) {
	t.Helper()
	class, found := botloadout.ClassByKey(key)
	if !found {
		t.Fatalf("no class %q", key)
	}
	field := form.classField(class)
	if !field.Handle(tea.KeyMsg{Type: tea.KeyLeft}) {
		t.Fatalf("the %s toggle ignored the key", key)
	}
}
