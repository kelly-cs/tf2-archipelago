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

/*
	page is one page of the model as a walk tab.

A page whose rows name sections becomes a tab widget of its own, one page per
section: the Bots page is six seats, nine classes and two cosmetic ticks, which
is a list on a terminal and a wall in a window. Anything the model nests under
this page joins them as a section of its own.

A page with no sections is the grid it always was.
*/
func (w *windowScreen) page(tab form.Tab, nested []form.Tab, poolView **walk.TableView, pool *poolModel) declarative.TabPage {
	sections := sectionsOf(tab)
	if len(sections) == 0 && len(nested) == 0 {
		return w.sheet(tab, poolView, pool)
	}

	pages := make([]declarative.TabPage, 0, len(sections)+len(nested))
	for _, section := range sections {
		pages = append(pages, w.sheet(section, poolView, pool))
	}
	for _, under := range nested {
		pages = append(pages, w.page(under, nil, poolView, pool))
	}
	return declarative.TabPage{
		Title:    tab.Title,
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{declarative.TabWidget{StretchFactor: 1, Pages: pages}},
	}
}

/*
	sectionsOf splits a page into its sections, in the order the rows declare

them. Nothing comes back for a page whose rows name none, which is how the
caller tells the two layouts apart.

A page that mixes named and unnamed rows would lose the unnamed ones, so it is
refused at the seam rather than drawn wrong: form's own test is what keeps a new
row from being the one that does it.
*/
func sectionsOf(tab form.Tab) []form.Tab {
	var sections []form.Tab
	for _, field := range tab.Fields {
		if field.Bar {
			continue
		}
		if field.Group == "" {
			return nil
		}
		if len(sections) == 0 || sections[len(sections)-1].Title != field.Group {
			sections = append(sections, form.Tab{Title: field.Group})
		}
		at := len(sections) - 1
		sections[at].Fields = append(sections[at].Fields, field)
	}
	if len(sections) > 0 {
		sections[0].Intro = tab.Intro
	}
	return sections
}

/*
	sheet is one page of rows.

The page holding the mission table gives it the slack: the table is the page.
Every other page ends in a spacer, because without somewhere to put the slack a
grid hands it to the rows themselves, which is a page of three settings spread
down the window with the sentence explaining them a hand's width above them.
That is what these looked like before this.

No scroll view. The tall page was Bots and it is four short sections now, so
every page fits the window at its smallest; a view that scrolls costs a page
either a horizontal scrollbar it does not need or, held across, a hand's width
of nothing down the left of the short pages.
*/
func (w *windowScreen) sheet(tab form.Tab, poolView **walk.TableView, pool *poolModel) declarative.TabPage {
	columns, children, tabled := w.grid(tab, poolView, pool)
	if !tabled {
		children = append(children, declarative.VSpacer{ColumnSpan: columns})
	}
	return declarative.TabPage{
		Title:    tab.Title,
		Layout:   declarative.Grid{Columns: columns},
		Children: children,
	}
}

/*
	pairRunLeast is the shortest run of short rows worth putting two to a line.

Nine classes and six seats are a tick and a menu apiece with nothing long in
them. One to a line makes a column of eighteen rows down a window with room for
two of them side by side, which is what the Bots page looked like before this.

Ten, so it is those two runs and nothing else. Player options has six short rows
in a row and reads worse two-up: a difficulty menu the width of a sentence and a
mission count three characters wide are not a pair of anything, and the page had
one column when players last liked it.
*/
const pairRunLeast = 10

// pairable is a row short enough to sit beside another.
func pairable(field form.Field) bool {
	switch field.Kind {
	case form.Toggle, form.Choice, form.Number:
		return !field.Browse
	}
	return false
}

// pairRun is the run of pairable rows at the front of these, empty unless it is
// long enough to be worth laying out two to a line.
func pairRun(fields []form.Field) []form.Field {
	end := 0
	for end < len(fields) && pairable(fields[end]) {
		end++
	}
	if end < pairRunLeast {
		return nil
	}
	return fields[:end]
}

/*
	grid is the rows of one page: how many columns they want, the widgets, and

whether one of them is the mission table.

Four columns is two label-and-control pairs to a line, and a row that is not in
such a run spans the rest of the line so the values still start in one place.
*/
func (w *windowScreen) grid(tab form.Tab, poolView **walk.TableView, pool *poolModel) (int, []declarative.Widget, bool) {
	/* The bar rows are drawn along the bottom of the window by the dialog, so
	   they are out of the page before anything counts a run. */
	rows := make([]form.Field, 0, len(tab.Fields))
	for _, field := range tab.Fields {
		if !field.Bar {
			rows = append(rows, field)
		}
	}

	columns := 2
	for at := range rows {
		if len(pairRun(rows[at:])) > 0 {
			columns = 4
			break
		}
	}

	children := make([]declarative.Widget, 0, 2*len(rows)+4)
	if tab.Intro != "" {
		children = append(children, declarative.TextLabel{
			Text:       tab.Intro,
			ColumnSpan: columns,
			MaxSize:    declarative.Size{Width: sentenceWidth},
		})
	}

	var pooled bool
	for at := 0; at < len(rows); at++ {
		field := rows[at]
		if strings.HasPrefix(field.ID, poolPrefix) {
			pooled = true
			continue
		}
		/* Buttons go along a line of their own, not one to a row with an empty
		   label beside each. Generate seed, Open tf2.yaml and Open the folder
		   are one strip and always were; a row apiece is three lines of window
		   spent on a column that says nothing. */
		if run := buttonRun(rows[at:]); len(run) > 0 {
			children = append(children, w.buttonStrip(run, columns))
			at += len(run) - 1
			continue
		}
		if run := pairRun(rows[at:]); len(run) > 0 {
			children = append(children, w.pairs(run)...)
			at += len(run) - 1
			continue
		}
		children = append(children, w.row(field, columns-1)...)
	}

	/* The pool is twenty-six toggles, which is a sensible list of rows and a
	   poor screen. The window has room for columns, so it draws them as a
	   checkable table with the tier, the wave count and what each needs. The
	   rows are the same rows, written back under the same IDs. */
	if pooled {
		children = append(children,
			declarative.TextLabel{
				Text:        "Missions the run may draw. Untick one to keep it out of every seed generated from here: Caliginous Caper is one wave of 666 robots and an hour on its own. The tier above still applies.",
				ColumnSpan:  columns,
				MaxSize:     declarative.Size{Width: sentenceWidth},
				ToolTipText: "This is the excluded_missions option in tf2.yaml.",
			},
			declarative.TableView{
				AssignTo:         poolView,
				Model:            pool,
				CheckBoxes:       true,
				AlternatingRowBG: true,
				ColumnSpan:       columns,
				StretchFactor:    1,
				Columns: []declarative.TableViewColumn{
					{Title: "Mission", Width: 320},
					{Title: "In the pool", Width: 110},
					{Title: "Compatibility", Width: 220},
				},
			},
		)
	}

	return columns, children, pooled
}

// pairs lays a run of short rows out two to a line, padding an odd last row so
// the next thing on the page starts on a line of its own.
func (w *windowScreen) pairs(fields []form.Field) []declarative.Widget {
	out := make([]declarative.Widget, 0, 2*len(fields)+2)
	for _, field := range fields {
		out = append(out, w.row(field, 1)...)
	}
	if len(fields)%2 == 1 {
		out = append(out, declarative.Label{Text: ""}, declarative.Label{Text: ""})
	}
	return out
}

/*
	buttonRun is the run of buttons starting at the front of these rows.

One is a run. A button's label is the words on the button, so drawing a row for
it puts "Set up / check Funnel" in the label column and again across the width
of the page beside it. Bar rows are not here at all, so a run is never broken by
one the window drew elsewhere.
*/
func buttonRun(fields []form.Field) []form.Field {
	end := 0
	for end < len(fields) && (fields[end].Kind == form.Action || fields[end].Kind == form.Confirm) {
		end++
	}
	if end == 0 {
		return nil
	}
	return fields[:end]
}

// buttonStrip is a run of buttons on one line across the page.
func (w *windowScreen) buttonStrip(fields []form.Field, columns int) declarative.Widget {
	children := make([]declarative.Widget, 0, len(fields)+1)
	for _, field := range fields {
		button, _ := w.control(field, 0).(declarative.PushButton)
		button.MinSize = declarative.Size{Width: buttonWidth(field.Label)}
		children = append(children, button)
	}
	children = append(children, declarative.HSpacer{})
	return declarative.Composite{
		Layout:     declarative.HBox{MarginsZero: true},
		ColumnSpan: columns,
		MaxSize:    declarative.Size{Height: 32},
		Children:   children,
	}
}

// row is the label and the control for one field. span is how many columns the
// control takes, so a lone row still reaches the edge of a four-column page.
func (w *windowScreen) row(field form.Field, span int) []declarative.Widget {
	/* Floored, not capped. A cap here would clip the longest label, and it
	   would not hold the column anyway: the spacer that ends the page spans
	   every column and lifts every maximum with it. What keeps the labels in
	   one narrow column is that every control beside one asks for the slack. */
	label := declarative.Label{
		Text:        field.Label,
		MinSize:     declarative.Size{Width: labelWidth},
		ToolTipText: field.Help,
	}
	return []declarative.Widget{label, w.control(field, span)}
}

// control is the widget for one row, and the closure that reads it back. A span
// of one or less is the widget's own cell, which is what a button bar wants.
func (w *windowScreen) control(field form.Field, span int) declarative.Widget {
	switch field.Kind {
	case form.Toggle:
		var box *walk.CheckBox
		w.read(field, func() string { return strconv.FormatBool(box.Checked()) })
		/* Beside a spacer, like a number. A grid gives its slack to the columns
		   that ask for it and splits it evenly when none do, and a tick asks
		   for nothing: a page whose controls were all ticks put half the window
		   in the label column and the ticks down the middle of it. */
		return declarative.Composite{
			Layout:     declarative.HBox{MarginsZero: true},
			ColumnSpan: span,
			Children: []declarative.Widget{
				declarative.CheckBox{
					AssignTo: &box, Text: field.Hint,
					Checked: field.Value == "true", Enabled: !field.Disabled,
					ToolTipText: field.Help,
				},
				declarative.HSpacer{},
			},
		}

	case form.Number:
		var edit *walk.NumberEdit
		w.read(field, func() string { return strconv.Itoa(int(edit.Value())) })
		/* Held at the width of a number, at the left of its column, with a
		   spacer taking the rest.

		   walk right-aligns the digits inside a NumberEdit and offers no way to
		   ask it not to, so a box stretched across the column puts a two-digit
		   answer as far from the label naming it as the window is wide. Nor
		   does MaxSize hold it back on its own: a grid cell only honours a
		   maximum for a widget that does not grow, and this one grows. */
		return declarative.Composite{
			Layout:     declarative.HBox{MarginsZero: true},
			ColumnSpan: span,
			Children: []declarative.Widget{
				declarative.NumberEdit{
					AssignTo: &edit, Value: numberValue(field),
					MinValue: float64(field.Low), MaxValue: float64(field.High),
					Decimals: 0, Enabled: !field.Disabled,
					MinSize:     declarative.Size{Width: numberWidth},
					MaxSize:     declarative.Size{Width: numberWidth},
					ToolTipText: field.Help,
				},
				declarative.HSpacer{},
			},
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
			StretchFactor: 1, ColumnSpan: span, ToolTipText: field.Help,
		}

	case form.Password:
		var edit *walk.LineEdit
		w.read(field, func() string { return edit.Text() })
		return declarative.LineEdit{
			AssignTo: &edit, Text: field.Value, PasswordMode: true,
			Enabled: !field.Disabled, ColumnSpan: span, ToolTipText: field.Help,
		}

	case form.Action, form.Confirm:
		// A Confirm is a button here too: the message box is what asks again,
		// which the action itself puts up. The terminal wants a second Enter
		// instead, because it has no message box to put up.
		id := field.ID
		return declarative.PushButton{
			Text: field.Label, ToolTipText: field.Help, Enabled: !field.Disabled,
			ColumnSpan: span,
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
			line.ColumnSpan = span
			return line
		}
		// A folder gets the picker beside it. The terminal has none and simply
		// does not offer one, which is the right kind of difference.
		return declarative.Composite{
			Layout:     declarative.HBox{MarginsZero: true},
			ColumnSpan: span,
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

// numberValue is the row's value as a NumberEdit wants it. Saved settings may
// outlive the bounds they were written under: the mission pool can shrink, and
// a newer launcher can tighten a fixed bound. walk refuses to create the whole
// dialog when a NumberEdit starts outside its range, so put a stale value on
// the nearest bound where the player can see and save it.
//
// A row that does not hold a number falls back to the floor rather than to
// zero, which may itself be outside the bounds.
func numberValue(field form.Field) float64 {
	n, err := strconv.Atoi(field.Value)
	if err != nil {
		return float64(field.Low)
	}
	if n < field.Low {
		return float64(field.Low)
	}
	if n > field.High {
		return float64(field.High)
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

/*
	buttonWidth is room for a button's label, estimated from its length.

walk asks Windows for the text extent and adds a fixed padding, and the answer
is a few pixels short of the truth for the longest labels here: "Show the
settings file" came back clipped to "Show the settings fil". A floor of its own
costs nothing, since a button in a strip that is wider than it needs is followed
by a spacer that takes the slack.
*/
func buttonWidth(label string) int {
	const perRune, padding, least = 7, 24, 80
	return max(perRune*len([]rune(label))+padding, least)
}
