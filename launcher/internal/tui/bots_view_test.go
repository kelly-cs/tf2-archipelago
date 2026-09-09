package tui

import (
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/botlive"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/*
	The tab names every seat RED holds, not only the ones the lineup spells out.

A lineup of two against a team of six used to read as a team of two, and the
four the mod draws are the four a player wonders about.
*/
func TestTheSwitcherNamesEverySeat(t *testing.T) {
	m := screen(t)
	m.settings = settings.Settings{
		SrcdsBotTeamSize:       6,
		SrcdsBotTeamComp:       []string{"engineer", "medic"},
		SrcdsBotSeatLoadouts:   []string{"gunslinger", ""},
		SrcdsBotClassBlacklist: []string{"spy", "sniper"},
	}
	m.view = viewBots

	got := m.View()
	for _, want := range []string{"Engineer", "Medic", botlive.DrawnClass} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if count := strings.Count(got, botlive.DrawnClass); count != 4 {
		t.Errorf("%d drawn seats on screen, want 4:\n%s", count, got)
	}
	// An unticked class is why a drawn seat never comes up a Spy, so the tab
	// that raises the question answers it.
	if !strings.Contains(got, "Scout, Soldier") || strings.Contains(got, "Spy,") {
		t.Errorf("the classes the mod may draw from are wrong:\n%s", got)
	}
}

// The columns are a fixed width apart, so a seat changing class does not move
// the weapons column of every other row.
func TestTheSwitcherColumnsLineUp(t *testing.T) {
	m := screen(t)
	m.settings = settings.Settings{
		SrcdsBotTeamSize: 3,
		SrcdsBotTeamComp: []string{"heavyweapons", "medic", "engineer"},
	}
	m.view = viewBots

	var at []int
	for line := range strings.SplitSeq(m.View(), "\n") {
		if index := strings.Index(line, "Stock weapons"); index >= 0 {
			at = append(at, index)
		}
	}
	if len(at) != 3 {
		t.Fatalf("%d weapons cells, want 3", len(at))
	}
	for _, index := range at {
		if index != at[0] {
			t.Errorf("the weapons column starts at %v, and a column is one place", at)
		}
	}
}
