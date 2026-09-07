//go:build windows

package gui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/lxn/walk"

	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/*
The settings window, built for real and read back.

Everything else about the port to internal/form is checked on the machine that
writes it. This is not: walk is a Win32 binding, so the window only exists on
Windows, and the one part nobody could check was whether the widgets it makes
are the rows form declared.

So these run under Wine. `make gui-test` cross-compiles this package's tests and
runs the binary under Xvfb, which is enough to create a window, its tabs and its
controls. It is not enough to click anything, and it is not a substitute for a
player opening the real thing on Windows; what it does prove is that every row
form declares becomes a control, that the control matches the kind, and that
reading the widgets back gives the state they were built from.

That last one is the assertion worth having. A window that draws every row and
reads one of them back into the wrong setting looks perfect and loses the
player's answer at Save, which is exactly the class of bug uiparity was written
to catch by reading source and could not.
*/

func testSettings() settings.Settings {
	s := settings.Defaults()
	s.InstallRoot = `C:\tf2ap`
	// Values away from the defaults, so a control that quietly reports its own
	// default rather than what it was built with is caught.
	s.MvmTrapPct = 37
	s.MvmWeaponBuffPct = 12
	s.SrcdsBluHealthPct = 250
	s.MvmCashRewards = false
	s.MvmMissionTicketImportance = "useful"
	s.SrcdsHostname = "a server with a name"
	s.SrcdsPort = 27043
	s.SrcdsBotTeamSize = 4
	s.SrcdsBotTeamComp = []string{"engineer", "medic"}
	return s
}

func build(t *testing.T) *settingsDialog {
	t.Helper()
	built, err := buildSettingsDialog(nil, testSettings(),
		func(s settings.Settings) (settings.Settings, error) { return s, nil },
		func() ([]string, error) { return nil, nil },
		func() error { return nil },
		func(string, ...any) {},
	)
	if err != nil {
		t.Fatalf("the settings window did not build: %v", err)
	}
	t.Cleanup(built.dialog.Dispose)
	return built
}

// The window has a top-level tab per page form declares that is not nested, in
// the same order. A page missing here is a page a player cannot reach.
func TestTheWindowHasAPageForEveryPageFormDeclares(t *testing.T) {
	built := build(t)

	var got []string
	for i := range built.tabs.Pages().Len() {
		got = append(got, built.tabs.Pages().At(i).Title())
	}

	var want []string
	for _, tab := range built.model.Tabs {
		if tab.Under == "" {
			want = append(want, tab.Title)
		}
	}

	if !slices.Equal(got, want) {
		t.Errorf("the window shows\n  %v\nand form declares\n  %v", got, want)
	}
}

/*
	A sectioned page draws a tab for every section its rows name, and the pages

form nests under it as well.

The Bots page is six seats, nine classes, two cosmetic ticks and the loadout
builder. It was four sub-tabs until the rows moved into form and came back as
one list of thirty-nine, which is the regression this asks about.
*/
func TestASectionedPageHasATabForEverySection(t *testing.T) {
	built := build(t)

	for _, tab := range built.model.Tabs {
		if tab.Under != "" {
			continue
		}
		var want []string
		for _, field := range tab.Fields {
			if field.Bar || field.Group == "" {
				continue
			}
			if len(want) == 0 || want[len(want)-1] != field.Group {
				want = append(want, field.Group)
			}
		}
		for _, nested := range built.model.Tabs {
			if nested.Under == tab.Title {
				want = append(want, nested.Title)
			}
		}
		if len(want) == 0 {
			continue
		}

		inner := innerTabs(pageNamed(t, built, tab.Title))
		if inner == nil {
			t.Errorf("%s: form names the sections %v and the window drew one flat page", tab.Title, want)
			continue
		}
		var got []string
		for i := range inner.Pages().Len() {
			got = append(got, inner.Pages().At(i).Title())
		}
		if !slices.Equal(got, want) {
			t.Errorf("%s shows\n  %v\nand form names\n  %v", tab.Title, got, want)
		}
	}
}

/*
	The three rows that are about the launcher are along the bottom of the

window, not one to a line on the Game server page.

Debug logs, Repair and Reset settings were there until the rows moved into form.
A player who has been asked for a debug bundle looks at the bottom of the window
for it, which is the only reason this is worth asserting.
*/
func TestBarRowsAreOnTheBottomBarAndNotOnAPage(t *testing.T) {
	built := build(t)

	var bar int
	for _, tab := range built.model.Tabs {
		for _, field := range tab.Fields {
			if !field.Bar {
				continue
			}
			bar++
			if barButton(built, field.Label) == nil {
				t.Errorf("%q is a bar row and the window drew no button for it", field.Label)
			}
			if controlFor(t, pageNamed(t, built, tab.Title), field) != nil {
				t.Errorf("%q is on the %s page as well as on the bar", field.Label, tab.Title)
			}
		}
	}
	if bar == 0 {
		t.Fatal("form declares no bar rows, so this checked nothing")
	}
}

/*
	Every row form declares becomes a control of the right kind.

The widget tree is walked rather than the builder's own bookkeeping, so this
asks what walk actually made. A row that was declared and never added shows up
as a label with nothing beside it; a row given the wrong control shows up as the
wrong type here.
*/
func TestEveryRowBecomesAControl(t *testing.T) {
	built := build(t)

	for _, tab := range built.model.Tabs {
		for _, field := range tab.Fields {
			if strings.HasPrefix(field.ID, poolPrefix) {
				continue // drawn in the table, checked below
			}
			if field.Bar {
				continue // drawn on the bottom bar, checked there
			}
			widget := controlFor(t, pageOf(t, built, tab, field), field)
			if widget == nil {
				t.Errorf("%s: form declares %q and the window has no control for it", tab.Title, field.Label)
				continue
			}
			if !rightKind(field, widget) {
				t.Errorf("%s: %q is a %s and the window drew a %T", tab.Title, field.Label, field.Kind, widget)
			}
		}
	}
}

/*
	Reading the widgets back gives the state they were built from.

This is the one that matters. Save reads every control in one pass and applies
each answer under the ID form gave the row, so a control wired to the wrong ID
loses one setting and silently changes another. Building from settings that
differ from the defaults in a dozen places and reading them back unchanged is
what says the wiring is right, field by field.
*/
func TestReadingTheWindowBackGivesWhatItWasBuiltFrom(t *testing.T) {
	built := build(t)
	want := testSettings()

	got, err := built.collect()
	if err != nil {
		t.Fatalf("reading the window back: %v", err)
	}

	for _, check := range []struct {
		name string
		got  any
		want any
	}{
		{"MvmTrapPct", got.Settings.MvmTrapPct, want.MvmTrapPct},
		{"MvmWeaponBuffPct", got.Settings.MvmWeaponBuffPct, want.MvmWeaponBuffPct},
		{"SrcdsBluHealthPct", got.Settings.SrcdsBluHealthPct, want.SrcdsBluHealthPct},
		{"MvmCashRewards", got.Settings.MvmCashRewards, want.MvmCashRewards},
		{"MvmMissionTicketImportance", got.Settings.MvmMissionTicketImportance, want.MvmMissionTicketImportance},
		{"MvmMissionCount", got.Settings.MvmMissionCount, want.MvmMissionCount},
		{"MvmDifficulty", got.Settings.MvmDifficulty, want.MvmDifficulty},
		{"MvmGoal", got.Settings.MvmGoal, want.MvmGoal},
		{"MvmMedalOnClear", got.Settings.MvmMedalOnClear, want.MvmMedalOnClear},
		{"MvmDeathLink", got.Settings.MvmDeathLink, want.MvmDeathLink},
		{"SrcdsHostname", got.Settings.SrcdsHostname, want.SrcdsHostname},
		{"SrcdsPort", got.Settings.SrcdsPort, want.SrcdsPort},
		{"SrcdsBots", got.Settings.SrcdsBots, want.SrcdsBots},
		{"SrcdsBotTeamSize", got.Settings.SrcdsBotTeamSize, want.SrcdsBotTeamSize},
		{"SrcdsBotHats", got.Settings.SrcdsBotHats, want.SrcdsBotHats},
		{"SrcdsReach", got.Settings.SrcdsReach, want.SrcdsReach},
		{"TailscaleFastDL", got.Settings.TailscaleFastDL, want.TailscaleFastDL},
		{"FastDLPort", got.Settings.FastDLPort, want.FastDLPort},
	} {
		if check.got != check.want {
			t.Errorf("%s read back as %v, was built from %v", check.name, check.got, check.want)
		}
	}

	if !slices.Equal(got.Settings.SrcdsBotTeamComp, want.SrcdsBotTeamComp) {
		t.Errorf("the team read back as %q, was built from %q",
			got.Settings.SrcdsBotTeamComp, want.SrcdsBotTeamComp)
	}
	if !slices.Equal(got.Settings.MvmExcludedMissions, want.MvmExcludedMissions) {
		t.Errorf("the pool read back %d missions out, was built from %d",
			len(got.Settings.MvmExcludedMissions), len(want.MvmExcludedMissions))
	}
}

/*
	Every setting is reachable through the window, not only through form.

form's own test says every field of settings.Settings is on a page. This says
the window actually draws that page's rows: a page declared and never built
would pass there and lose every setting on it here.

The count is compared rather than the names, because what is being asked is
whether anything was dropped between the declaration and the widget tree.
*/
func TestTheWindowDrawsEveryDeclaredRow(t *testing.T) {
	built := build(t)

	var declared, drawn int
	for _, tab := range built.model.Tabs {
		for _, field := range tab.Fields {
			declared++
			switch {
			case strings.HasPrefix(field.ID, poolPrefix):
				drawn++ // in the table
			case field.Bar:
				if barButton(built, field.Label) != nil {
					drawn++
				}
			case controlFor(t, pageOf(t, built, tab, field), field) != nil:
				drawn++
			}
		}
	}
	if declared == 0 {
		t.Fatal("form declared no rows, so this checked nothing")
	}
	if drawn != declared {
		t.Errorf("form declares %d rows and the window drew %d", declared, drawn)
	}
}

// The mission table holds a row per pool toggle, and ticking one writes it back
// under the ID form gave it.
func TestTheMissionTableIsThePoolRows(t *testing.T) {
	built := build(t)

	var pooled int
	for _, tab := range built.model.Tabs {
		for _, field := range tab.Fields {
			if strings.HasPrefix(field.ID, poolPrefix) {
				pooled++
			}
		}
	}
	if pooled == 0 {
		t.Fatal("form declares no pool rows")
	}
	if built.pool.RowCount() != pooled {
		t.Fatalf("the table has %d rows and form declares %d missions", built.pool.RowCount(), pooled)
	}

	// Untick the first playable mission and read it back out of the state.
	at := slices.IndexFunc(built.pool.rows, func(f form.Field) bool { return !f.Disabled })
	if at < 0 {
		t.Fatal("no mission in the table can be ticked")
	}
	if err := built.pool.SetChecked(at, false); err != nil {
		t.Fatalf("unticking %s: %v", built.pool.rows[at].Label, err)
	}
	got, err := built.collect()
	if err != nil {
		t.Fatalf("reading the window back: %v", err)
	}
	popFile := strings.TrimPrefix(built.pool.rows[at].ID, poolPrefix)
	if !slices.Contains(got.Settings.MvmExcludedMissions, popFile) {
		t.Errorf("unticking %s left it in the pool", popFile)
	}
}

// A mission the launcher cannot play refuses the tick rather than taking it and
// dropping it at Save, and says what is missing in the Compatibility column.
func TestAnUnplayableMissionRefusesTheTick(t *testing.T) {
	built := build(t)

	at := slices.IndexFunc(built.pool.rows, func(f form.Field) bool { return f.Disabled })
	if at < 0 {
		t.Skip("no unplayable mission in this build's tables")
	}
	if err := built.pool.SetChecked(at, true); err != nil {
		t.Fatalf("ticking an unplayable mission: %v", err)
	}
	if built.pool.Checked(at) {
		t.Error("an unplayable mission took the tick")
	}
	if why, _ := built.pool.Value(at, 2).(string); why == "" {
		t.Error("an unplayable mission does not say what it is missing")
	}
}

// Every button form declares has something wired to it. The terminal has the
// same test; this is the window's, and it needs the dialog because the actions
// are closures over it.
func TestEveryButtonIsWired(t *testing.T) {
	built := build(t)

	var buttons int
	for _, tab := range built.model.Tabs {
		for _, field := range tab.Fields {
			if field.Kind != form.Action && field.Kind != form.Confirm {
				continue
			}
			buttons++
			if built.screen.actions[field.ID] == nil {
				t.Errorf("form declares the button %q and the window does nothing with it", field.ID)
			}
		}
	}
	if buttons == 0 {
		t.Fatal("form declares no buttons, so this checked nothing")
	}
	for id := range built.screen.actions {
		if _, ok := built.model.Field(id); !ok {
			t.Errorf("the window wires %q and form declares no such row", id)
		}
	}
}

/*
	A number row carries the bounds form resolved for it.

The first version of this test set a trap share of 1000 and expected Save to
refuse it. It does not, and cannot: walk's NumberEdit clamps to its own
MaxValue, which is form's High, so the value never reaches form to be refused.

That is the better answer and it is worth asserting directly. The bounds crossing
the seam is what makes the window unable to produce a value form would reject,
and it is not free: the mission count's ceiling is the pool the chosen tier
leaves, so a spec whose Bounds were ignored here would let a player ask for
twelve missions out of a pool of four.

form's own refusal is still the second line, and form's tests cover it. This is
about whether the window was told.
*/
func TestNumberRowsCarryTheBoundsFormResolved(t *testing.T) {
	built := build(t)

	var checked int
	for _, tab := range built.model.Tabs {
		for _, field := range tab.Fields {
			if field.Kind != form.Number {
				continue
			}
			edit, ok := controlFor(t, pageOf(t, built, tab, field), field).(*walk.NumberEdit)
			if !ok {
				t.Errorf("%q is a number and the window drew something else", field.Label)
				continue
			}
			checked++
			if got := int(edit.MinValue()); got != field.Low {
				t.Errorf("%q has a floor of %d in the window and %d in form", field.Label, got, field.Low)
			}
			if got := int(edit.MaxValue()); got != field.High {
				t.Errorf("%q has a ceiling of %d in the window and %d in form", field.Label, got, field.High)
			}

			// And the clamp is real, so nothing form would refuse can be typed.
			_ = edit.SetValue(float64(field.High + 1000))
			if int(edit.Value()) > field.High {
				t.Errorf("%q took %v, past its ceiling of %d", field.Label, edit.Value(), field.High)
			}
		}
	}
	if checked == 0 {
		t.Fatal("form declares no number rows, so this checked nothing")
	}
}

// --- walking the widget tree ---

/*
	pageNamed is the tab with that title, at either level.

The Bots page is a tab widget of its own, so Team, Classes, Looks and Loadouts
are pages inside a page. A row lives on the innermost one that names it, which
is what the rest of these tests walk.
*/
func pageNamed(t *testing.T, built *settingsDialog, title string) *walk.TabPage {
	t.Helper()
	if page := findPage(built.tabs, title); page != nil {
		return page
	}
	t.Fatalf("the window has no page %q", title)
	return nil
}

func findPage(tabs *walk.TabWidget, title string) *walk.TabPage {
	for i := range tabs.Pages().Len() {
		page := tabs.Pages().At(i)
		if page.Title() == title {
			return page
		}
		if inner := innerTabs(page); inner != nil {
			if found := findPage(inner, title); found != nil {
				return found
			}
		}
	}
	return nil
}

/*
	pageOf is the page a row is drawn on.

A row that names a section is on the section's page, and one that does not is on
its tab's own. A Bar row is on no page at all: the window draws it along the
bottom, so the caller is told to look there instead.
*/
func pageOf(t *testing.T, built *settingsDialog, tab form.Tab, field form.Field) *walk.TabPage {
	t.Helper()
	if field.Group != "" {
		return pageNamed(t, built, field.Group)
	}
	return pageNamed(t, built, tab.Title)
}

/*
	controlFor is the widget beside a row's label.

The pages are two-column grids of label, control, label, control, so the control
is whatever follows the label with that text. Matching on the label rather than
on an index is what makes this survive a row being added above.
*/
func controlFor(t *testing.T, page *walk.TabPage, field form.Field) walk.Widget {
	t.Helper()
	children := descendants(page)
	for i, widget := range children {
		if field.Kind == form.Action || field.Kind == form.Confirm {
			if button, ok := widget.(*walk.PushButton); ok && button.Text() == field.Label {
				return button
			}
			continue
		}
		label, ok := widget.(*walk.Label)
		if !ok || label.Text() != field.Label {
			continue
		}
		if i+1 >= len(children) {
			return nil
		}
		/* A row that needs the slack held off its control puts the two in a
		   Composite: a folder and its Browse button, a number and the spacer
		   that keeps it the width of a number. The control is what follows,
		   because descendants lists a container before what is inside it. */
		if _, wrapped := children[i+1].(*walk.Composite); wrapped && i+2 < len(children) {
			return children[i+2]
		}
		return children[i+1]
	}
	return nil
}

// descendants is every widget under one, flattened. A folder row puts its edit
// and its Browse button in a Composite, so the tree has to be walked rather
// than the page's own children listed.
func descendants(parent walk.Container) []walk.Widget {
	var out []walk.Widget
	children := parent.Children()
	if children == nil {
		return out
	}
	for i := range children.Len() {
		widget := children.At(i)
		out = append(out, widget)
		// A TabWidget is not a Container in walk: its pages hang off Pages()
		// rather than off Children(), so the rows inside one are invisible to
		// a walk that only follows containers.
		if inner, ok := widget.(*walk.TabWidget); ok {
			for j := range inner.Pages().Len() {
				out = append(out, descendants(inner.Pages().At(j))...)
			}
			continue
		}
		if container, ok := widget.(walk.Container); ok {
			out = append(out, descendants(container)...)
		}
	}
	return out
}

func rightKind(field form.Field, widget walk.Widget) bool {
	switch field.Kind {
	case form.Toggle:
		_, ok := widget.(*walk.CheckBox)
		return ok
	case form.Number:
		_, ok := widget.(*walk.NumberEdit)
		return ok
	case form.Choice:
		_, ok := widget.(*walk.ComboBox)
		return ok
	case form.Action, form.Confirm:
		_, ok := widget.(*walk.PushButton)
		return ok
	default:
		_, ok := widget.(*walk.LineEdit)
		return ok
	}
}

/*
	A Save that cannot write does not close the window and does not report saved.

From a player's debug bundle. The window used to close first and hand the
settings to whoever opened it, which wrote them and put a line in a log pane.
By then the window was gone and so was everything typed into it. Their log has
no line about saving at all, which is how a refused Save looked from outside.

The message box itself cannot be driven from here: walk.MsgBox is modal and
nothing clicks it. What is checked is what surrounds it, which is what was
actually wrong: the write is attempted while the window is up, a failure leaves
ok false so the caller writes nothing either, and the reason reaches the log so
the next bundle carries it.
*/
func TestAFailedWriteIsNotReportedAsSaved(t *testing.T) {
	var logged []string
	built, err := buildSettingsDialog(nil, testSettings(),
		func(s settings.Settings) (settings.Settings, error) {
			return s, errors.New(`cannot write C:\Users\x\AppData\Roaming\tf2ap\config.json: access is denied`)
		},
		func() ([]string, error) { return nil, nil },
		func() error { return nil },
		func(format string, args ...any) { logged = append(logged, fmt.Sprintf(format, args...)) },
	)
	if err != nil {
		t.Fatalf("the settings window did not build: %v", err)
	}
	t.Cleanup(built.dialog.Dispose)

	if built.ok {
		t.Error("the window reports a save nobody asked for yet")
	}

	// What Save does, short of the message box: the state is read back and the
	// write is attempted.
	next, err := built.collect()
	if err != nil {
		t.Fatalf("reading the window back: %v", err)
	}
	if _, err := built.persist(next.Settings); err == nil {
		t.Fatal("the refusing persist did not refuse")
	}
	if built.ok {
		t.Error("a refused write left the window reporting a save")
	}

	/* And the reason reaches the log and the message, because a refused save
	   used to write neither and the reporter's bundle said nothing about it.
	   noteRefusal rather than refuse: the box itself is modal and nothing here
	   can click it. */
	message := built.note("", "cannot write config.json: access is denied")
	if !slices.ContainsFunc(logged, func(line string) bool {
		return strings.Contains(line, "not saved") && strings.Contains(line, "access is denied")
	}) {
		t.Errorf("the refusal is not in the log: %q", logged)
	}
	if !strings.Contains(message, "access is denied") {
		t.Errorf("the message box would say %q", message)
	}
	if !strings.Contains(message, "Nothing in this window is lost") {
		t.Error("the message box does not say the answers are still there")
	}
}

// A refusal that names a page opens it, so the player is looking at the field
// the message is about. The reporter was two pages away from theirs.
func TestARefusalOpensThePageItIsAbout(t *testing.T) {
	built := build(t)

	showPage(built.tabs, "Game server")
	built.note("Archipelago room", "Room address: the address is empty")

	at := built.tabs.CurrentIndex()
	if got := built.tabs.Pages().At(at).Title(); got != "Archipelago room" {
		t.Errorf("the refusal left the window on %q", got)
	}
}

// barButton is the button with that label along the bottom of the window.
func barButton(built *settingsDialog, label string) *walk.PushButton {
	for _, widget := range descendants(built.dialog) {
		if button, ok := widget.(*walk.PushButton); ok && button.Text() == label {
			return button
		}
	}
	return nil
}
