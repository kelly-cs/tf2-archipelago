package form

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/m-this/tf2-archipelago/launcher/internal/botcards"
)

type cardSeat struct {
	class   string
	name    string
	loadout string
}

func cardSeatAt(s State, index int) cardSeat {
	return cardSeat{
		class:   at(s.Settings.SrcdsBotTeamComp, index),
		name:    at(s.Settings.SrcdsBotSeatNames, index),
		loadout: at(s.Settings.SrcdsBotSeatLoadouts, index),
	}
}

func putCardSeat(s State, index int, seat cardSeat) State {
	s.Settings.SrcdsBotTeamComp = withAt(s.Settings.SrcdsBotTeamComp, index, seat.class, Seats)
	s.Settings.SrcdsBotSeatNames = withAt(s.Settings.SrcdsBotSeatNames, index, seat.name, Seats)
	s.Settings.SrcdsBotSeatLoadouts = withAt(s.Settings.SrcdsBotSeatLoadouts, index, seat.loadout, Seats)
	return s
}

// One atomic setting change chooses the card's class, pinned name and starting
// weapons. The older per-seat rows remain available for hand tuning afterward.
func seatCardSpec(index int) Spec {
	return openChoice(fmt.Sprintf("bots.seat.%d.card", index), "Bots",
		fmt.Sprintf("Seat %d card", index+1),
		"Pick a named defender prototype. Its class, name and preferred weapons fill this seat together; the rows below can still be customized.",
		func(State, Env) []Option {
			out := []Option{{Value: "", Label: "no card"}}
			for _, card := range botcards.Cards {
				out = append(out, Option{Value: card.ID, Label: card.Name})
			}
			return out
		},
		func(s State) string {
			seat := cardSeatAt(s, index)
			if card, ok := botcards.BySeat(seat.class, seat.name); ok {
				return card.ID
			}
			return ""
		},
		func(s State, id string) State {
			if id == "" {
				return putCardSeat(s, index, cardSeat{})
			}
			card, ok := botcards.ByID(id)
			if !ok {
				return s
			}
			for other := range Seats {
				if other == index {
					continue
				}
				seat := cardSeatAt(s, other)
				if seat.class == card.Class && seat.name == card.Name {
					s = putCardSeat(s, other, cardSeatAt(s, index))
					return putCardSeat(s, index, seat)
				}
			}
			current := cardSeatAt(s, index)
			s = putCardSeat(s, index, cardSeat{card.Class, card.Name, card.Loadout})
			if current.class != card.Class || current.name != card.Name {
				s = setCardForm(s, card.ID, randomCardForm())
			}
			return s
		})
}

func selectedCardIDs(s State) []string {
	var ids []string
	for index := range Seats {
		seat := cardSeatAt(s, index)
		if card, ok := botcards.BySeat(seat.class, seat.name); ok {
			ids = append(ids, card.ID)
		}
	}
	return ids
}

const (
	cardHuman = "human"
	cardRobot = "robot"
	cardGiant = "giant"
)

func randomCardForm() string {
	return []string{cardHuman, cardRobot, cardGiant}[rand.IntN(3)]
}

func cardForm(s State, id string) string {
	if slices.Contains(s.Settings.SrcdsBotGiantCards, id) {
		return cardGiant
	}
	if slices.Contains(s.Settings.SrcdsBotHumanCards, id) {
		return cardHuman
	}
	return cardRobot
}

func setCardForm(s State, id, form string) State {
	s.Settings.SrcdsBotGiantCards = slices.DeleteFunc(slices.Clone(s.Settings.SrcdsBotGiantCards), func(other string) bool { return other == id })
	s.Settings.SrcdsBotHumanCards = slices.DeleteFunc(slices.Clone(s.Settings.SrcdsBotHumanCards), func(other string) bool { return other == id })
	switch form {
	case cardGiant:
		s.Settings.SrcdsBotGiantCards = append(s.Settings.SrcdsBotGiantCards, id)
	case cardHuman:
		s.Settings.SrcdsBotHumanCards = append(s.Settings.SrcdsBotHumanCards, id)
	}
	return s
}

func cardFormSpec(card botcards.Card) Spec {
	return openChoice("bots.card."+card.ID+".form", "Bots", card.Name+" form",
		"A card draws Human, RED robot or RED robot Giant when recruited. This choice stays with it across maps; change or reroll it here.",
		func(State, Env) []Option {
			return []Option{{Value: cardHuman, Label: "Human"}, {Value: cardRobot, Label: "RED robot"},
				{Value: cardGiant, Label: "RED robot Giant"}, {Value: "reroll", Label: "Reroll at random"}}
		},
		func(s State) string { return cardForm(s, card.ID) },
		func(s State, form string) State {
			if form == "reroll" {
				form = randomCardForm()
			}
			if form != cardHuman && form != cardRobot && form != cardGiant {
				return s
			}
			return setCardForm(s, card.ID, form)
		})
}

func demoLineupSpec() Spec {
	return openChoice("bots.demo_lineup", "Bots", "Demo card lineup",
		"Draft the six prototype cards in priority order. Nothing reaches the server until Apply.",
		func(State, Env) []Option {
			return []Option{{Value: "", Label: "keep this team"}, {Value: "demo", Label: "use all six cards"}}
		},
		func(State) string { return "" },
		func(s State, choice string) State {
			if choice != "demo" {
				return s
			}
			forms := []string{cardHuman, cardHuman, cardRobot, cardRobot, cardGiant, cardGiant}
			order := rand.Perm(len(forms))
			for index, card := range botcards.Cards {
				s = putCardSeat(s, index, cardSeat{card.Class, card.Name, card.Loadout})
				s = setCardForm(s, card.ID, forms[order[index]])
			}
			return s
		})
}

// Priority is a projection of the saved seat order, not a second setting.
// Reordering carries each bot's custom loadout and pinned name with it.
func cardPrioritySpec() Spec {
	return Spec{
		ID: "bots.priority", Tab: "Bots", Kind: Text,
		Label: "Card priority", Help: "Highest priority first. Drag cards in the Bot Switcher; lower priority bots leave first when humans join.",
		Get: func(s State) string { return strings.Join(selectedCardIDs(s), ",") },
		Set: func(s State, raw string) (State, error) {
			wanted := []string{}
			if raw != "" {
				wanted = strings.Split(raw, ",")
			}
			current := selectedCardIDs(s)
			slices.Sort(wanted)
			sorted := slices.Clone(current)
			slices.Sort(sorted)
			if !slices.Equal(wanted, sorted) {
				return s, fmt.Errorf("priority must list each selected card once")
			}
			ordered := strings.Split(raw, ",")
			if raw == "" {
				ordered = nil
			}
			byID := map[string]cardSeat{}
			var other []cardSeat
			for index := range Seats {
				seat := cardSeatAt(s, index)
				if card, ok := botcards.BySeat(seat.class, seat.name); ok {
					byID[card.ID] = seat
				} else if seat.class != "" {
					other = append(other, seat)
				}
			}
			var seats []cardSeat
			for _, id := range ordered {
				seats = append(seats, byID[id])
			}
			seats = append(seats, other...)
			for index := range Seats {
				var seat cardSeat
				if index < len(seats) {
					seat = seats[index]
				}
				s = putCardSeat(s, index, seat)
			}
			return s, nil
		},
	}
}
