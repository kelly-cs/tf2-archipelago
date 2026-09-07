//go:build windows

package gui

import (
	"strconv"
	"strings"

	"github.com/lxn/walk"
	declarative "github.com/lxn/walk/declarative"

	"github.com/m-this/tf2-archipelago/launcher/internal/form"
)

/*
windowScreen builds the settings window out of form.Model.

The window used to carry its own list of rows, and so did the terminal, and the
two said different things: the same row told a window player what the remainder
of the spare checks award and told a terminal player nothing. Neither wording
was chosen. There is one list now and both read it.

What is still the window's own is how a kind is drawn, and it should be. A
Choice is a ComboBox here and a pair of arrow keys in the terminal; a Number is
a NumberEdit here and typed digits there.
*/
type windowScreen struct {
	// reads is one closure per row, in the order the rows were built. Each
	// returns the change that row's widget now holds.
	reads []func() form.Change

	// actions is what a button does, by the ID form declared it under. Filled
	// after the dialog exists, because most of them need it to put a message
	// box on.
	actions map[string]func()

	// cues are the placeholder texts, applied after the window is made rather
	// than while it is being built. See applyCues.
	cues []cue
}

// cue is one placeholder waiting for its edit to exist. The edit is a pointer
// to the pointer walk fills in during Create, because at the time the row is
// declared there is no widget to hold yet.
type cue struct {
	edit **walk.LineEdit
	text string
}

/*
	applyCues puts the placeholder text on, and does not mind if it cannot.

EM_SETCUEBANNER needs comctl32 version 6, which an executable gets from its
application manifest. tf2ap.exe has one; a test binary built with `go test -c`
does not, and neither does anything running where that manifest was lost. walk's
declarative CueBanner field treats the failure as fatal, so the whole settings
window refused to open under Wine over greyed-out placeholder text.

That trade is the wrong way round. The placeholder says what a blank field
means and the window is how the game gets configured, so the error is reported
and the window opens. It is the one error here that is swallowed, and this is
the reason.
*/
func (w *windowScreen) applyCues(say func(string, ...any)) {
	for _, c := range w.cues {
		if *c.edit == nil {
			continue
		}
		if err := (*c.edit).SetCueBanner(c.text); err != nil {
			say("the settings window has no placeholder text: %v", err)
			return
		}
	}
}

// poolPrefix marks the rows the window draws as a table rather than as rows.
const poolPrefix = "missions.pool."

// page is one page of the model as a walk tab.
func (w *windowScreen) page(tab form.Tab, poolView **walk.TableView, pool *poolModel) declarative.TabPage {
	children := make([]declarative.Widget, 0, 2*len(tab.Fields)+4)
	if tab.Intro != "" {
		children = append(children, declarative.TextLabel{
			Text:       tab.Intro,
			ColumnSpan: 2,
			MaxSize:    declarative.Size{Width: sentenceWidth},
		})
	}

	var pooled bool
	for _, field := range tab.Fields {
		if strings.HasPrefix(field.ID, poolPrefix) {
			pooled = true
			continue
		}
		children = append(children, w.row(field)...)
	}

	/* The pool is twenty-six toggles, which is a sensible list of rows and a
	   poor screen. The window has room for columns, so it draws them as a
	   checkable table with the tier, the wave count and what each needs. The
	   rows are the same rows, written back under the same IDs. */
	if pooled {
		children = append(children,
			declarative.Label{
				Text:        "Missions the run may draw. Untick one to keep it out of every seed generated from here: Caliginous Caper is one wave of 666 robots and an hour on its own. The tier above still applies.",
				ColumnSpan:  2,
				ToolTipText: "This is the excluded_missions option in tf2.yaml.",
			},
			declarative.TableView{
				AssignTo:         poolView,
				Model:            pool,
				CheckBoxes:       true,
				AlternatingRowBG: true,
				ColumnSpan:       2,
				StretchFactor:    1,
				Columns: []declarative.TableViewColumn{
					{Title: "Mission", Width: 320},
					{Title: "In the pool", Width: 110},
					{Title: "Compatibility", Width: 220},
				},
			},
		)
	}

	return declarative.TabPage{
		Title:    tab.Title,
		Layout:   declarative.Grid{Columns: 2},
		Children: children,
	}
}

// row is the label and the control for one field.
func (w *windowScreen) row(field form.Field) []declarative.Widget {
	label := declarative.Label{
		Text:        field.Label,
		MinSize:     declarative.Size{Width: labelWidth},
		ToolTipText: field.Help,
	}
	return []declarative.Widget{label, w.control(field)}
}

// control is the widget for one row, and the closure that reads it back.
func (w *windowScreen) control(field form.Field) declarative.Widget {
	switch field.Kind {
	case form.Toggle:
		var box *walk.CheckBox
		w.read(field, func() string { return strconv.FormatBool(box.Checked()) })
		return declarative.CheckBox{
			AssignTo: &box, Text: field.Hint,
			Checked: field.Value == "true", Enabled: !field.Disabled,
			ToolTipText: field.Help,
		}

	case form.Number:
		var edit *walk.NumberEdit
		w.read(field, func() string { return strconv.Itoa(int(edit.Value())) })
		return declarative.NumberEdit{
			AssignTo: &edit, Value: numberValue(field),
			MinValue: float64(field.Low), MaxValue: float64(field.High),
			Decimals: 0, Enabled: !field.Disabled,
			ToolTipText: field.Help,
		}

	case form.Choice:
		var box *walk.ComboBox
		options := field.Options
		w.read(field, func() string {
			at := max(box.CurrentIndex(), 0)
			if at >= len(options) {
				return field.Value
			}
			return options[at].Value
		})
		return declarative.ComboBox{
			AssignTo: &box, Model: optionLabels(options),
			Value: selectedLabel(field), Enabled: !field.Disabled,
			StretchFactor: 1, ToolTipText: field.Help,
		}

	case form.Password:
		var edit *walk.LineEdit
		w.read(field, func() string { return edit.Text() })
		return declarative.LineEdit{
			AssignTo: &edit, Text: field.Value, PasswordMode: true,
			Enabled: !field.Disabled, ToolTipText: field.Help,
		}

	case form.Action, form.Confirm:
		// A Confirm is a button here too: the message box is what asks again,
		// which the action itself puts up. The terminal wants a second Enter
		// instead, because it has no message box to put up.
		id := field.ID
		return declarative.PushButton{
			Text: field.Label, ToolTipText: field.Help, Enabled: !field.Disabled,
			OnClicked: func() {
				if run := w.actions[id]; run != nil {
					run()
				}
			},
		}

	default:
		var edit *walk.LineEdit
		w.read(field, func() string { return edit.Text() })
		if field.Placeholder != "" {
			w.cues = append(w.cues, cue{edit: &edit, text: field.Placeholder})
		}
		line := declarative.LineEdit{
			AssignTo: &edit, Text: field.Value,
			Enabled: !field.Disabled, StretchFactor: 1, ToolTipText: field.Help,
		}
		if !field.Browse {
			return line
		}
		// A folder gets the picker beside it. The terminal has none and simply
		// does not offer one, which is the right kind of difference.
		return declarative.Composite{
			Layout: declarative.HBox{MarginsZero: true},
			Children: []declarative.Widget{
				line,
				declarative.PushButton{Text: "Browse", OnClicked: func() { browseForFolder(edit) }},
			},
		}
	}
}

// read registers how a row is read back. A disabled row reports the value it
// was built with: the widget is greyed out, so whatever it holds is not an
// answer anybody gave.
func (w *windowScreen) read(field form.Field, value func() string) {
	w.reads = append(w.reads, func() form.Change {
		if field.Disabled {
			return form.Change{Field: field.ID, Value: field.Value}
		}
		return form.Change{Field: field.ID, Value: value()}
	})
}

// numberValue is the row's value as a NumberEdit wants it. A row that does not
// hold a number falls back to the floor rather than to zero, which may be
// outside the bounds and which walk refuses to display.
func numberValue(field form.Field) float64 {
	n, err := strconv.Atoi(field.Value)
	if err != nil {
		return float64(field.Low)
	}
	return float64(n)
}

func optionLabels(options []form.Option) []string {
	labels := make([]string, 0, len(options))
	for _, o := range options {
		labels = append(labels, o.Label)
	}
	return labels
}

// selectedLabel is what the combo box shows for the value the state holds. An
// empty string leaves walk on the first entry, which is what a value no option
// matches should look like: nothing claimed.
func selectedLabel(field form.Field) string {
	for _, o := range field.Options {
		if o.Value == field.Value {
			return o.Label
		}
	}
	return ""
}
