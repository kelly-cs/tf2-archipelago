package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/m-this/tf2-archipelago/launcher/internal/form"
)

/*
modelRow is a settings row drawn from form.Model.

The rows above it in field.go each hold a pointer into the edited settings, so
the terminal names MvmTrapPct and the window names MvmTrapPct and uiparity read
both files to check they still named the same set. This one names nothing: it
has the row form.Build resolved and an ID to send changes back under, and what
that ID writes is form's business.

The keys are still here, and they belong here. Left and right on a choice, a
digit on a number and a space on a toggle are what a terminal does about the
kinds; a combo box and a checkbox are what a window does about the same three.
*/
type modelRow struct {
	f form.Field

	// apply hands the change to form and says whether it took. A refusal is
	// shown rather than swallowed: the bounds are on screen, so a number
	// outside them is a typo the player can see and fix.
	apply func(form.Change) error

	// typed is the digits of a Number as they are being entered, which is not
	// the same as the value. Typing 1 then 0 then 0 into a field bounded at 100
	// passes through 1 and 10, and both are inside the bounds; a fourth digit
	// is not, and the refusal is what drops it.
	typed string

	// armed is a Confirm that has been pressed once. The window asks with a
	// message box and this asks by wanting the second Enter, so neither one
	// goes off under a finger that was scrolling.
	armed bool

	// command is what a press left for the event loop, taken once.
	command tea.Cmd

	// run is what an Action does. Actions are dispatched by ID rather than
	// applied, so the screen supplies the work and form supplies the label.
	run func() tea.Cmd
}

func (r *modelRow) Label() string { return r.f.Label }

func (r *modelRow) Help() string {
	switch {
	case r.armed:
		return r.f.Warning
	case r.f.Disabled:
		return r.f.Reason
	}
	return r.f.Help
}

func (r *modelRow) Value() string {
	// A row that cannot be used says what is missing where its value would be.
	// "unavailable" tells a player nothing they can act on; "missing bot .nav"
	// tells them the pack is short a file and no tick will fix it.
	if r.f.Disabled {
		return styleMuted.Render(r.f.Reason)
	}
	switch r.f.Kind {
	case form.Text:
		if r.f.Value == "" {
			return styleMuted.Render(r.f.Placeholder)
		}
		return r.f.Value
	case form.Password:
		return strings.Repeat("•", len([]rune(r.f.Value)))
	case form.Number:
		if r.typed != "" {
			return r.typed
		}
		return r.f.Value
	case form.Toggle:
		if r.f.Value == "true" {
			return "[x] " + r.f.Hint
		}
		if r.f.HintOff != "" {
			return "[ ] " + r.f.HintOff
		}
		return "[ ] " + r.f.Hint
	case form.Choice:
		return fmt.Sprintf("%s %s %s",
			styleMuted.Render("<"), r.optionLabel(), styleMuted.Render(">"))
	case form.Confirm:
		if r.armed {
			return styleWarn.Render("enter again to confirm")
		}
		return styleMuted.Render(r.f.Hint)
	case form.Action:
		return styleMuted.Render(r.f.Hint)
	}
	return r.f.Value
}

// optionLabel is what the player reads for the value the settings hold. A value
// with no option is shown as itself rather than as a blank: a settings file
// written by an older launcher can hold one, and a blank row says nothing about
// why.
func (r *modelRow) optionLabel() string {
	for _, o := range r.f.Options {
		if o.Value == r.f.Value {
			return o.Label
		}
	}
	return r.f.Value
}

func (r *modelRow) Handle(msg tea.KeyMsg) bool {
	if r.f.Disabled {
		return false
	}
	switch r.f.Kind {
	case form.Text, form.Password:
		return r.handleTyping(msg)
	case form.Number:
		return r.handleNumber(msg)
	case form.Toggle:
		return r.handleToggle(msg)
	case form.Choice:
		return r.handleChoice(msg)
	case form.Action:
		return r.handlePress(msg)
	case form.Confirm:
		return r.handleConfirm(msg)
	}
	return false
}

func (r *modelRow) handleTyping(msg tea.KeyMsg) bool {
	next := r.f.Value
	switch msg.Type {
	case tea.KeyRunes:
		next += string(msg.Runes)
	case tea.KeySpace:
		next += " "
	case tea.KeyBackspace:
		runes := []rune(next)
		if len(runes) == 0 {
			return false
		}
		next = string(runes[:len(runes)-1])
	default:
		return false
	}
	return r.set(next)
}

func (r *modelRow) handleNumber(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "left", "h":
		return r.step(-1)
	case "right", "l":
		return r.step(1)
	case "backspace":
		if r.typed == "" {
			return false
		}
		r.typed = r.typed[:len(r.typed)-1]
		if r.typed != "" {
			r.set(r.typed)
		}
		return true
	}
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 || msg.Runes[0] < '0' || msg.Runes[0] > '9' {
		return false
	}
	// The digit is kept only if the number it makes is one form takes, which is
	// what stops a third digit landing in a field that stops at 100.
	next := r.typed + string(msg.Runes)
	if !r.set(next) {
		return true
	}
	r.typed = next
	return true
}

// step moves a number by one and clamps at the bounds, because a held arrow key
// at the ceiling is not a typo to report.
func (r *modelRow) step(by int) bool {
	n, err := strconv.Atoi(r.f.Value)
	if err != nil {
		return false
	}
	r.typed = ""
	return r.set(strconv.Itoa(min(max(n+by, r.f.Low), r.f.High)))
}

func (r *modelRow) handleToggle(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "left", "right", " ", "h", "l", "x":
		return r.set(strconv.FormatBool(r.f.Value != "true"))
	}
	return false
}

func (r *modelRow) handleChoice(msg tea.KeyMsg) bool {
	if len(r.f.Options) == 0 {
		return false
	}
	at := 0
	for i, o := range r.f.Options {
		if o.Value == r.f.Value {
			at = i
			break
		}
	}
	switch msg.String() {
	case "left", "h":
		at = (at - 1 + len(r.f.Options)) % len(r.f.Options)
	case "right", "l", " ":
		at = (at + 1) % len(r.f.Options)
	default:
		return false
	}
	return r.set(r.f.Options[at].Value)
}

func (r *modelRow) handlePress(msg tea.KeyMsg) bool {
	if msg.Type != tea.KeyEnter || r.run == nil {
		return false
	}
	r.command = r.run()
	return true
}

func (r *modelRow) handleConfirm(msg tea.KeyMsg) bool {
	if msg.Type != tea.KeyEnter || r.run == nil {
		r.armed = false
		return false
	}
	if !r.armed {
		r.armed = true
		return true
	}
	r.armed = false
	r.command = r.run()
	return true
}

// set sends the change and keeps the row's own copy in step with what form
// took, so the row redraws from the value that was actually written.
func (r *modelRow) set(value string) bool {
	if err := r.apply(form.Change{Field: r.f.ID, Value: value}); err != nil {
		return false
	}
	r.f.Value = value
	return true
}

func (r *modelRow) take() tea.Cmd {
	command := r.command
	r.command = nil
	return command
}

func (r *modelRow) disarm() { r.armed = false }

// change is one answer under a row's ID, which is the only thing an interface
// ever sends form.
func change(id, value string) form.Change { return form.Change{Field: id, Value: value} }
