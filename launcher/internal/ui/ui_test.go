package ui

import (
	"bufio"
	"strings"
	"testing"
)

// promptOn builds a Prompt reading from a fixed string, the way a pipe or a
// redirect from /dev/null feeds the real one.
func promptOn(input string) *Prompt {
	return &Prompt{reader: bufio.NewReader(strings.NewReader(input))}
}

// Every prompt that re-asks on a bad answer used to drop the read error, so
// stdin ending left it asking forever. One of them wrote 849 MB of the same
// line in ninety seconds. These finish or the test does not.
func TestPromptsStopWhenStdinEnds(t *testing.T) {
	t.Run("Text", func(t *testing.T) {
		if got := promptOn("").Text("label", "fallback"); got != "fallback" {
			t.Errorf("Text = %q", got)
		}
	})
	t.Run("Int", func(t *testing.T) {
		// "x" is not a number, so this is the path that re-asks.
		if got := promptOn("x\n").Int("label", 7); got != 7 {
			t.Errorf("Int = %d", got)
		}
	})
	t.Run("IntRange", func(t *testing.T) {
		if got := promptOn("99\n").IntRange("label", 4, 1, 6); got != 4 {
			t.Errorf("IntRange = %d", got)
		}
	})
	t.Run("Bool", func(t *testing.T) {
		if got := promptOn("maybe\n").Bool("label", true); !got {
			t.Errorf("Bool = %v", got)
		}
	})
	t.Run("Choice", func(t *testing.T) {
		got := promptOn("nope\n").Choice("label", []string{"a", "b"}, "b")
		if got != "b" {
			t.Errorf("Choice = %q", got)
		}
	})
	t.Run("Select", func(t *testing.T) {
		options := []Option{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}}
		if got := promptOn("42\n").Select("label", options, "two"); got != "two" {
			t.Errorf("Select = %q", got)
		}
	})
}

// Closed is what the caller checks before insisting on a value nobody can
// type. It is false until a read actually runs out.
func TestClosedFollowsStdin(t *testing.T) {
	p := promptOn("archipelago.gg:1\n")
	if p.Closed() {
		t.Error("Closed before reading anything")
	}
	if got := p.Text("label", ""); got != "archipelago.gg:1" {
		t.Errorf("Text = %q", got)
	}
	if p.Closed() {
		t.Error("Closed after an answer that arrived")
	}
	if got := p.Text("label", "fallback"); got != "fallback" {
		t.Errorf("second Text = %q", got)
	}
	if !p.Closed() {
		t.Error("not Closed after stdin ran out")
	}
}

// A last line with no newline on the end is still an answer.
func TestLastLineWithoutANewlineStillCounts(t *testing.T) {
	if got := promptOn("typed").Text("label", "fallback"); got != "typed" {
		t.Errorf("Text = %q", got)
	}
}
