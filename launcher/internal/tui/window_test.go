package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// rows numbers n lines, so a test can say which of them reached the screen.
func rows(n int) []string {
	out := make([]string, 0, n)
	for i := range n {
		out = append(out, fmt.Sprintf("row%d", i))
	}
	return out
}

// A list that fits is the whole list, and says nothing about scrolling.
func TestAListThatFitsSaysNothing(t *testing.T) {
	m := screen(t)
	m.listOffset = 4

	got := m.window(rows(3), 10)
	if strings.Count(got, "\n") != 9 {
		t.Errorf("the body is not 10 lines:\n%s", got)
	}
	if strings.Contains(got, "below") || strings.Contains(got, "above") {
		t.Errorf("a list that fits counts what is off the screen:\n%s", got)
	}
	// Left over from a longer list, and the short one has nowhere to scroll.
	if m.listOffset != 0 {
		t.Errorf("the offset stayed at %d on a list that fits", m.listOffset)
	}
}

/*
	A list longer than the body scrolls, and says how much is off it

The last line is the count rather than a row, because a list that stops at the
bottom row reads as the whole list. That is the Unlocks tab: one row per item
the multiworld has handed over, past a screenful early in a run.
*/
func TestALongListScrollsAndSaysSo(t *testing.T) {
	m := screen(t)

	got := m.window(rows(30), 10)
	if !strings.Contains(got, "row0") || strings.Contains(got, "row9\n") {
		t.Errorf("the top of the list is wrong:\n%s", got)
	}
	if !strings.Contains(got, "21 below") || strings.Contains(got, "above") {
		t.Errorf("the count of what is off the screen is wrong:\n%s", got)
	}

	m.listOffset = 5
	got = m.window(rows(30), 10)
	if strings.Contains(got, "row4") || !strings.Contains(got, "row5") {
		t.Errorf("scrolling down five did not move the top:\n%s", got)
	}
	if !strings.Contains(got, "5 above, 16 below") {
		t.Errorf("the count of what is off both ends is wrong:\n%s", got)
	}
}

// The end of the list is where scrolling stops, and the keystroke does not
// know where that is: only the view has counted the rows.
func TestScrollingStopsAtTheEnd(t *testing.T) {
	m := screen(t)
	m.listOffset = 500

	got := m.window(rows(30), 10)
	if !strings.Contains(got, "row29") || !strings.Contains(got, "row21") {
		t.Errorf("the bottom of the list is not on screen:\n%s", got)
	}
	if m.listOffset != 21 {
		t.Errorf("the offset clamped to %d, want 21", m.listOffset)
	}
	if !strings.Contains(got, "21 above") || strings.Contains(got, "below") {
		t.Errorf("the count past the end is wrong:\n%s", got)
	}
}

/*
	The Bot Switcher scrolls, checked through the screen rather than the helper

A tab that builds its rows and then cuts them itself is the bug this replaced,
and only the screen shows which of the two it does. The Unlocks tab is wired
the same way and is not checked here: its list is behind a running server, and
this package has no way to start one.
*/
func TestTheSwitcherScrollsWhenTheTeamIsLong(t *testing.T) {
	m := screen(t)
	m.settings = settings.Settings{SrcdsBotTeamSize: 40}
	m.view = viewBots

	got := m.View()
	if !strings.Contains(got, "below") {
		t.Errorf("the Bot Switcher does not say the team runs off the screen:\n%s", got)
	}

	m.listOffset = 10
	if scrolled := m.View(); scrolled == got {
		t.Errorf("scrolling the Bot Switcher changed nothing:\n%s", scrolled)
	}
}
