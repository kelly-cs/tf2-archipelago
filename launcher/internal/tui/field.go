package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

/*
A field is one row of the settings: a name, a value, and what the keys do to
it. The window has a control per kind of answer, and this is the same set with
the same behaviour written for a keyboard.

Left and right change what has a set of answers, typing changes what does not,
and every field says underneath what it is for, because a name alone does not
tell anybody what a difficulty floor or a login token is.
*/
type field interface {
	Label() string
	Help() string
	Value() string
	// Handle takes the key and says whether it changed anything, so the caller
	// knows a key nobody wanted is still free for the screen itself.
	Handle(tea.KeyMsg) bool
}

// textField is a line of text: a room address, a server name, a token.
type textField struct {
	label, help string
	value       *string
	placeholder string
	hidden      bool
}

func (f *textField) Label() string { return f.label }
func (f *textField) Help() string  { return f.help }

func (f *textField) Value() string {
	switch {
	case *f.value == "":
		return styleMuted.Render(f.placeholder)
	case f.hidden:
		return strings.Repeat("•", len([]rune(*f.value)))
	default:
		return *f.value
	}
}

func (f *textField) Handle(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyRunes:
		*f.value += string(msg.Runes)
	case tea.KeySpace:
		*f.value += " "
	case tea.KeyBackspace:
		runes := []rune(*f.value)
		if len(runes) == 0 {
			return false
		}
		*f.value = string(runes[:len(runes)-1])
	default:
		return false
	}
	return true
}

// numberField is a count with a floor and a ceiling, changed a step at a time
// or typed over.
type numberField struct {
	label, help string
	value       *int
	low, high   int
	typed       string
}

func (f *numberField) Label() string { return f.label }
func (f *numberField) Help() string  { return f.help }

func (f *numberField) Value() string {
	if f.typed != "" {
		return f.typed
	}
	return strconv.Itoa(*f.value)
}

func (f *numberField) Handle(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "left", "h":
		f.set(*f.value - 1)
	case "right", "l":
		f.set(*f.value + 1)
	case "backspace":
		if f.typed == "" {
			return false
		}
		f.typed = f.typed[:len(f.typed)-1]
		f.commit()
	default:
		if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 || msg.Runes[0] < '0' || msg.Runes[0] > '9' {
			return false
		}
		f.typed += string(msg.Runes)
		f.commit()
	}
	return true
}

// commit keeps the typed digits and the number in step, so a value that would
// be out of range is refused rather than silently clamped under the cursor.
func (f *numberField) commit() {
	if f.typed == "" {
		return
	}
	typed, err := strconv.Atoi(f.typed)
	if err != nil || typed > f.high {
		f.typed = f.typed[:len(f.typed)-1]
		return
	}
	*f.value = typed
}

func (f *numberField) set(value int) {
	f.typed = ""
	*f.value = min(max(value, f.low), f.high)
}

// toggleField is a yes or no.
type toggleField struct {
	label, help string
	value       *bool
	on, off     string
}

func (f *toggleField) Label() string { return f.label }
func (f *toggleField) Help() string  { return f.help }

func (f *toggleField) Value() string {
	if *f.value {
		return "[x] " + f.on
	}
	return "[ ] " + f.off
}

func (f *toggleField) Handle(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "left", "right", " ", "h", "l", "x":
		*f.value = !*f.value
		return true
	}
	return false
}

// choiceField is one answer out of a list: a tier, a goal, a class, a reach.
type choiceField struct {
	label, help string
	options     []string
	index       int
	apply       func(index int)
}

func (f *choiceField) Label() string { return f.label }
func (f *choiceField) Help() string  { return f.help }

func (f *choiceField) Value() string {
	if len(f.options) == 0 {
		return ""
	}
	return fmt.Sprintf("%s %s %s",
		styleMuted.Render("<"), f.options[f.index], styleMuted.Render(">"))
}

func (f *choiceField) Handle(msg tea.KeyMsg) bool {
	if len(f.options) == 0 {
		return false
	}
	switch msg.String() {
	case "left", "h":
		f.index = (f.index - 1 + len(f.options)) % len(f.options)
	case "right", "l", " ":
		f.index = (f.index + 1) % len(f.options)
	default:
		return false
	}
	f.apply(f.index)
	return true
}

// actionField is a button: the seed the Archipelago app generates, the folder
// it lands in, the repair. It carries no value, only what pressing it does.
type actionField struct {
	label, help string
	hint        string
	run         func() tea.Cmd
	command     tea.Cmd
}

func (f *actionField) Label() string { return f.label }
func (f *actionField) Help() string  { return f.help }
func (f *actionField) Value() string { return styleMuted.Render(f.hint) }

func (f *actionField) Handle(msg tea.KeyMsg) bool {
	if msg.Type != tea.KeyEnter {
		return false
	}
	f.command = f.run()
	return true
}

// take is what the press left behind, once. The screen asks every field it
// hands a key to, because an action is the only kind that answers with work
// for the event loop rather than a new value.
func (f *actionField) take() tea.Cmd {
	command := f.command
	f.command = nil
	return command
}

func (f *actionField) disarm() {}

// confirmField is an action that cannot be taken back: the repair, the reset.
// The window asks with a message box, and this asks by wanting the second
// enter, so neither one goes off under a finger that was scrolling.
type confirmField struct {
	actionField
	warning string
	armed   bool
}

func (f *confirmField) Value() string {
	if f.armed {
		return styleWarn.Render("enter again to confirm")
	}
	return styleMuted.Render(f.hint)
}

func (f *confirmField) Help() string {
	if f.armed {
		return f.warning
	}
	return f.help
}

func (f *confirmField) Handle(msg tea.KeyMsg) bool {
	if msg.Type != tea.KeyEnter {
		f.armed = false
		return false
	}
	if !f.armed {
		f.armed = true
		return true
	}
	f.armed = false
	f.command = f.run()
	return true
}

// disarm is the change of mind the row never sees: moving off it, or to
// another tab, is handled by the screen, and an armed row left behind would
// fire on an enter meant for wherever the player went.
func (f *confirmField) disarm() { f.armed = false }
