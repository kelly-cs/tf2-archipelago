package tui

import (
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/*
	Every button form declares does something here

form says what a button is called and what it is for, and the work is the
terminal's because it ends in a tea.Cmd. That split leaves one way to be wrong:
declaring a row and forgetting to wire it, which draws a button that swallows
the Enter and does nothing. Nobody reports that as a bug, they report the
feature as missing.
*/
func TestEveryButtonDoesSomething(t *testing.T) {
	state := form.NewState(settings.Defaults())
	f := newSettingsForm(settings.Defaults(), settingsDeps{})

	var buttons int
	for _, spec := range form.Specs(state, f.env()) {
		if spec.Kind != form.Action && spec.Kind != form.Confirm {
			continue
		}
		buttons++
		if f.action(spec.ID) == nil {
			t.Errorf("form declares the button %q and the terminal does nothing with it", spec.ID)
		}
	}
	if buttons == 0 {
		t.Fatal("form declares no buttons, so this checked nothing")
	}
}

// And nothing is wired that form does not declare. A stale entry is a method
// kept alive by a dispatcher nobody reaches, which is how dead code survives a
// rename.
func TestNothingIsWiredThatFormDoesNotDeclare(t *testing.T) {
	state := form.NewState(settings.Defaults())
	f := newSettingsForm(settings.Defaults(), settingsDeps{})

	declared := map[string]bool{}
	for _, spec := range form.Specs(state, f.env()) {
		declared[spec.ID] = true
	}
	for _, id := range wiredActions {
		if !declared[id] {
			t.Errorf("the terminal wires %q and form declares no such row", id)
		}
	}
}
