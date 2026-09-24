package form

import (
	"reflect"
	"slices"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func TestBotCardSelectionAndPriorityKeepTheWholeIdentity(t *testing.T) {
	s := NewState(settings.Defaults())
	for _, change := range []Change{
		{Field: "bots.seat.0.card", Value: "chell"},
		{Field: "bots.seat.1.card", Value: "herr-doktor"},
	} {
		var err error
		s, err = Apply(s, Env{}, change)
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := cardSeatAt(s, 0); got != (cardSeat{"engineer", "Chell", "ranger"}) {
		t.Fatalf("first card became %+v", got)
	}
	// A player's custom weapon pick belongs to the card when it is dragged.
	s.Settings.SrcdsBotSeatLoadouts[0] = "custom:my-chelly"
	var err error
	s, err = Apply(s, Env{}, Change{Field: "bots.priority", Value: "herr-doktor,chell"})
	if err != nil {
		t.Fatal(err)
	}
	if got := cardSeatAt(s, 1); got != (cardSeat{"engineer", "Chell", "custom:my-chelly"}) {
		t.Fatalf("dragging split the card from its loadout: %+v", got)
	}
	if got := s.Settings.SrcdsBotSeatNames; !reflect.DeepEqual(got, []string{"Herr Doktor", "Chell"}) {
		t.Fatalf("priority did not carry pinned names: %v", got)
	}
	before := s
	if _, err := Apply(s, Env{}, Change{Field: "bots.priority", Value: "chell,chell"}); err == nil {
		t.Fatal("duplicate priority accepted")
	}
	if !reflect.DeepEqual(s, before) {
		t.Fatal("refused priority changed the team")
	}
}

func TestChoosingAnAlreadySeatedCardSwapsWholeSeats(t *testing.T) {
	s := NewState(settings.Defaults())
	s.Settings.SrcdsBotTeamComp = []string{"scout", "medic"}
	s.Settings.SrcdsBotSeatNames = []string{"CreditToTeam", "Herr Doktor"}
	s.Settings.SrcdsBotSeatLoadouts = []string{"custom:runner", "kritz"}
	next, err := Apply(s, Env{}, Change{Field: "bots.seat.1.card", Value: "credit-to-team"})
	if err != nil {
		t.Fatal(err)
	}
	if got := cardSeatAt(next, 1); got != (cardSeat{"scout", "CreditToTeam", "custom:runner"}) {
		t.Fatalf("selected card lost its custom loadout: %+v", got)
	}
	if got := cardSeatAt(next, 0); got != (cardSeat{"medic", "Herr Doktor", "kritz"}) {
		t.Fatalf("displaced card was not swapped: %+v", got)
	}
}

func TestAnyItemDefinitionCanBeDraftedInAnyOfferedSlot(t *testing.T) {
	s := NewState(settings.Defaults())
	s.Draft.Loadout = StockLoadout("medic")
	next, err := Apply(s, Env{}, Change{Field: "loadout.any_item.secondary", Value: "997"})
	if err != nil {
		t.Fatal(err)
	}
	if got := next.Draft.Slot("secondary"); got != 997 {
		t.Fatalf("Medic secondary became %d, want cross-class item 997", got)
	}
	field, ok := Build(next, Env{}).Field("loadout.slot.secondary")
	if !ok {
		t.Fatal("secondary picker disappeared for a manual item")
	}
	if field.Value != "997" {
		t.Fatalf("manual item disappeared from picker: %q", field.Value)
	}
	if _, err := Apply(next, Env{}, Change{Field: "loadout.any_item.secondary", Value: "not an item"}); err == nil {
		t.Fatal("non-numeric item definition accepted")
	}
}

func TestSpyCanTryAnExperimentalPrimary(t *testing.T) {
	s := NewState(settings.Defaults())
	s.Draft.Loadout = StockLoadout("spy")
	next, err := Apply(s, Env{}, Change{Field: "loadout.any_item.primary", Value: "997"})
	if err != nil || next.Draft.Slot("primary") != 997 {
		t.Fatalf("experimental Spy primary: value %d, error %v", next.Draft.Slot("primary"), err)
	}
}

func TestDemoLineupFillsAllSixNamedSeats(t *testing.T) {
	s := NewState(settings.Defaults())
	next, err := Apply(s, Env{}, Change{Field: "bots.demo_lineup", Value: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(selectedCardIDs(next)) != Seats {
		t.Fatalf("demo selected %d cards, want %d", len(selectedCardIDs(next)), Seats)
	}
	if got := cardSeatAt(next, 5); got != (cardSeat{"spy", "Mentlegen", "diamondback"}) {
		t.Fatalf("last demo seat became %+v", got)
	}
}

func TestGiantChoiceStaysWithCardAcrossPriorityAndSavedTeams(t *testing.T) {
	s := NewState(settings.Defaults())
	for _, change := range []Change{
		{Field: "bots.demo_lineup", Value: "demo"},
		{Field: "bots.card.herr-doktor.form", Value: "giant"},
		{Field: "bots.card.chell.form", Value: "human"},
		{Field: "bots.priority", Value: "herr-doktor,credit-to-team,screamin-eagles,ivan,chell,mentlegen"},
	} {
		var err error
		s, err = Apply(s, Env{}, change)
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := cardSeatAt(s, 0).name; got != "Herr Doktor" {
		t.Fatalf("Giant's priority seat = %q", got)
	}
	team := settings.BotTeamOf(s.Settings)
	restored := settings.WithBotTeam(settings.Defaults(), team)
	if got := restored.SrcdsBotGiantCards; !slices.Contains(got, "herr-doktor") {
		t.Fatalf("saved Giant cards = %v", got)
	}
	if got := cardForm(NewState(restored), "chell"); got != cardHuman {
		t.Fatalf("saved Chell form = %q", got)
	}
}
