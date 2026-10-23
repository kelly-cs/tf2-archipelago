package form

import (
	"slices"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/botnames"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func TestSavingALoadoutKeepsWhatWasBuilt(t *testing.T) {
	s := NewState(settings.Defaults())
	s.Draft.LoadoutName = " pop "
	s.Draft.Loadout = botloadout.Built{Class: "scout", Primary: 448, Second: botloadout.Stock, Melee: botloadout.Stock, PDA2: botloadout.Stock}

	next, said, ok := Act(s, "loadout.save")
	if !ok || said != "saved the loadout as pop" {
		t.Fatalf("Act answered %q, %v", said, ok)
	}
	if got := next.Settings.SrcdsBotCustomLoadouts["pop"]; got.Primary != 448 {
		t.Fatalf("the saved loadout is %+v, want the one built", got)
	}
	if s.Settings.SrcdsBotCustomLoadouts != nil {
		t.Fatal("the state handed in was written to")
	}
}

func TestAnUnnamedSaveIsRefusedWithoutChangingAnything(t *testing.T) {
	s := NewState(settings.Defaults())
	for _, id := range []string{"loadout.save", "bots.save_team", "bots.remove_team"} {
		next, said, ok := Act(s, id)
		if !ok || said == "" {
			t.Errorf("%s: answered %q, %v", id, said, ok)
		}
		if next.Settings.SrcdsBotCustomLoadouts != nil || next.Settings.SrcdsBotTeamPresets != nil {
			t.Errorf("%s: wrote something for no name", id)
		}
	}
}

func TestATeamIsKeptAndForgottenByName(t *testing.T) {
	s := NewState(settings.Defaults())
	s.Settings.SrcdsBotTeamComp = []string{"engineer", "engineer"}
	s.Draft.TeamName = "two engineers"

	kept, _, _ := Act(s, "bots.save_team")
	if kept.Settings.SrcdsBotTeamPresets["two engineers"].Comp[0] != "engineer" {
		t.Fatal("the team was not kept under its name")
	}
	if kept.Draft.TeamName != "" {
		t.Fatal("the name box was not cleared after the save")
	}

	kept.Draft.TeamName = "two engineers"
	gone, said, _ := Act(kept, "bots.remove_team")
	if gone.Settings.SrcdsBotTeamPresets != nil || said != "removed the team two engineers" {
		t.Fatalf("after removing, presets are %v and it said %q", gone.Settings.SrcdsBotTeamPresets, said)
	}

	if _, said, _ := Act(gone, "bots.remove_team"); said != "name the team to remove first" {
		t.Fatalf("removing with the box cleared said %q", said)
	}
	if _, _, ok := Act(s, "server.repair"); ok {
		t.Fatal("a button that is not the draft's was answered here")
	}
}

// A name goes in the pool once, at a length the game keeps, without a comma
// for the Compose list to split on.
func TestAddingABotNameRefusesWhatWouldNotSurvive(t *testing.T) {
	base := State{Settings: settings.Defaults()}

	for _, refusal := range []struct {
		name string
		said string
	}{
		{"", "type a name first"},
		{strings.Repeat("x", botnames.NameMax+1), "longer than"},
		{"Bob, the builder", "comma"},
		{botnames.Shipped()[0], "already in the pool"},
	} {
		s := base
		s.Draft.BotName = refusal.name
		after, said, ok := Act(s, "bots.name_add")
		if !ok || !strings.Contains(said, refusal.said) {
			t.Errorf("%q was answered %q, want %q", refusal.name, said, refusal.said)
		}
		if len(after.Settings.SrcdsBotNamesAdded) != 0 {
			t.Errorf("%q went into the pool anyway", refusal.name)
		}
	}

	s := base
	s.Draft.BotName = "  Gravel Pit Gary  "
	after, said, _ := Act(s, "bots.name_add")
	if !slices.Contains(after.Settings.SrcdsBotNamesAdded, "Gravel Pit Gary") {
		t.Fatalf("the name was not added: %v, said %q", after.Settings.SrcdsBotNamesAdded, said)
	}
	if after.Draft.BotName != "" {
		t.Errorf("the box still holds %q", after.Draft.BotName)
	}
}

// Adding a shipped name that was ticked off puts it back rather than adding a
// second copy of it, so the tick above says what the pool holds.
func TestAddingAShippedNameThatWasTakenOutPutsItBack(t *testing.T) {
	shipped := botnames.Shipped()[0]
	s := State{Settings: settings.Defaults()}
	s.Settings.SrcdsBotNamesExcluded = []string{shipped}
	s.Draft.BotName = shipped

	after, said, _ := Act(s, "bots.name_add")
	if slices.Contains(after.Settings.SrcdsBotNamesExcluded, shipped) {
		t.Errorf("%s is still left out: %q", shipped, said)
	}
	if len(after.Settings.SrcdsBotNamesAdded) != 0 {
		t.Errorf("it was added beside itself: %v", after.Settings.SrcdsBotNamesAdded)
	}
}
