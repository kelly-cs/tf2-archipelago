package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// screen is a model with a size and a supervisor, which is what View needs
// before it can draw anything.
func screen(t *testing.T) *model {
	t.Helper()
	s := settings.Defaults()
	s.APHost, s.APPort = "archipelago.gg", 12345
	s.SrcdsRconPw = "secret"

	m := newModel(s)
	m.supervisor = apruntime.NewSupervisor(s, nil, m.take)
	m.width, m.height, m.ready = 100, 30, true
	return m
}

// The main screen says what the window says: what the server is doing, where
// the room is, what a friend types to join, and which keys do what.
func TestTheMainScreenSaysWhatIsGoingOn(t *testing.T) {
	m := screen(t)
	m.take(apruntime.Line{At: time.Now(), Source: "srcds", Text: "fake server up"})
	m.drain()

	view := m.View()
	for _, want := range []string{
		"stopped",
		"room archipelago.gg:12345",
		"join",
		"Log",
		"Session",
		"fake server up",
		"rcon",
		"settings",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("the screen does not say %q:\n%s", want, view)
		}
	}
}

// A log longer than the screen shows its end, because the line that matters is
// the last one until somebody scrolls.
func TestTheLogShowsItsEnd(t *testing.T) {
	m := screen(t)
	for i := range 200 {
		m.take(apruntime.Line{At: time.Now(), Source: "srcds", Text: "line " + itoa(i)})
	}
	m.drain()

	view := m.View()
	if !strings.Contains(view, "line 199") {
		t.Errorf("the last line is not on screen:\n%s", view)
	}
	if strings.Contains(view, "line 0 ") {
		t.Errorf("the first line is still on screen:\n%s", view)
	}
}

// Every key the footer offers has to do something, or the footer is a lie.
func TestTheKeysDoWhatTheFooterSays(t *testing.T) {
	m := screen(t)

	if _, _ = m.Update(key("tab")); m.view != viewSession {
		t.Error("tab did not change the view")
	}
	if _, _ = m.Update(key("tab")); m.view != viewLog {
		t.Error("tab did not change it back")
	}
	if _, _ = m.Update(key("i")); !m.typing {
		t.Error("i did not reach the rcon line")
	}
	if _, _ = m.Update(key("esc")); m.typing {
		t.Error("esc did not leave the rcon line")
	}
	if _, _ = m.Update(key(",")); m.form == nil {
		t.Error("the settings never opened")
	}
	if _, _ = m.Update(key("esc")); m.form != nil {
		t.Error("esc did not close the settings")
	}
}

// What is typed at the rcon line is the command, and nothing typed there is
// read as a key that starts or stops the server.
func TestTypingGoesToTheCommandLine(t *testing.T) {
	m := screen(t)
	m.Update(key("i"))
	for _, k := range []string{"s", "t", "a", "t", "u", "s"} {
		m.Update(key(k))
	}

	if m.command != "status" {
		t.Errorf("the command line holds %q", m.command)
	}
	if m.supervisor.Running() {
		t.Error("typing s started the server")
	}
}

// The settings are the window's six tabs, and what they change is saved to the
// settings the launcher runs on.
func TestTheSettingsScreenEditsTheRun(t *testing.T) {
	m := screen(t)
	m.Update(key(","))

	if got := len(m.form.tabs); got != 6 {
		t.Errorf("the settings have %d tabs, want 6", got)
	}
	view := m.form.view(100, 30)
	for _, want := range []string{"Player options", "Bots", "Who can join (beta)", "Easiest tier"} {
		if !strings.Contains(view, want) {
			t.Errorf("the settings do not show %q:\n%s", want, view)
		}
	}

	// Death Link is the fifth row of the first tab: move to it and turn it on.
	before := m.form.edited.MvmDeathLink
	for range 4 {
		m.Update(key("down"))
	}
	m.Update(key(" "))
	if m.form.edited.MvmDeathLink == before {
		t.Error("space did not change the row it was on")
	}
}

// The tabs wrap, and every one of them draws: a tab whose fields panic is a
// tab nobody sees until a player opens it.
func TestEveryTabDraws(t *testing.T) {
	m := screen(t)
	m.Update(key(","))

	for i := range m.form.tabs {
		m.form.tab = i
		m.form.focused = 0
		if view := m.form.view(100, 30); !strings.Contains(view, m.form.tabs[i].title) {
			t.Errorf("tab %d does not draw its own title", i)
		}
		for _, row := range m.form.fields() {
			if row.Label() == "" {
				t.Errorf("tab %q has a row with no name", m.form.tabs[i].title)
			}
		}
	}
}

func key(name string) tea.KeyMsg {
	switch name {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
