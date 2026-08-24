package gui

import (
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
)

/* What the window's Bots tab means, without the window. From the 1.9.0
 * play-test: the seats named on the page were not the seats the mod filled.
 */
func TestBotTeamFromPicksKeepsTheSeatsAndTheTicks(t *testing.T) {
	engineer, heavy := classIndex(t, "engineer"), classIndex(t, "heavyweapons")
	sniper := classIndex(t, "sniper")

	picks := botTeamPicks{
		// Seat 1 draws, seat 2 is an Engineer, seat 3 draws, seat 4 a Heavy.
		SeatClass:    []int{0, engineer + 1, 0, heavy + 1},
		SeatLoadout:  make([]int, 4),
		Ticked:       ticks(true),
		ClassLoadout: make([]int, len(botloadout.Classes)),
	}
	picks.SeatLoadout[1] = 1 // the Engineer's first preset
	picks.Ticked[sniper] = false

	team := botTeamFrom(picks)

	if got := strings.Join(team.Comp, ","); got != ",engineer,,heavyweapons" {
		t.Errorf("comp = %q", got)
	}
	// A draw seat carries no weapons; a named seat says stock, because that is a
	// choice somebody made.
	if got := strings.Join(team.SeatLoadouts, ","); got != ",ranger,,stock" {
		t.Errorf("seat loadouts = %q", got)
	}
	if got := strings.Join(team.Blacklist, ","); got != "sniper" {
		t.Errorf("blacklist = %q", got)
	}
	// The unticked class reaches the mod, and the seats keep their numbers.
	if got := botloadout.Composition(team.Comp, team.Blacklist); got != ",engineer,,heavyweapons" {
		t.Errorf("composition = %q", got)
	}
}

// Every seat on the draw still writes its seats when a class is unticked, or
// the mod plays the map's own lineup instead.
func TestBotTeamFromPicksAllDraw(t *testing.T) {
	picks := botTeamPicks{
		SeatClass:    make([]int, 6),
		SeatLoadout:  make([]int, 6),
		Ticked:       ticks(true),
		ClassLoadout: make([]int, len(botloadout.Classes)),
	}
	picks.Ticked[classIndex(t, "spy")] = false

	team := botTeamFrom(picks)

	if len(team.Comp) != 0 {
		t.Errorf("comp = %q, want nothing named", team.Comp)
	}
	if got := botloadout.Composition(team.Comp, team.Blacklist); got != ",,,,," {
		t.Errorf("composition = %q", got)
	}
}

func ticks(value bool) []bool {
	out := make([]bool, len(botloadout.Classes))
	for i := range out {
		out[i] = value
	}
	return out
}

func classIndex(t *testing.T, key string) int {
	t.Helper()
	for index, class := range botloadout.Classes {
		if class.Key == key {
			return index
		}
	}
	t.Fatalf("no class %q", key)
	return -1
}
