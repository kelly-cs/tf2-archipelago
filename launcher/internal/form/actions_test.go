package form

import (
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
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
