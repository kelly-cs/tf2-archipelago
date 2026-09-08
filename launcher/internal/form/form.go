/*
Package form declares what the launcher asks the player.

A setting is one Spec: its words, values and state binding. Build resolves the
specs into the plain-data Model served to the browser, and Apply turns a browser
change back into State:

	Build(state, env) -> Model -> the interface draws it
	                                  |
	                                  v
	                Apply(state, env, Change) -> State -> Build again

# Why the Model holds no functions

A Spec holds closures because it lives in the process that owns the state. A
Model does not, because the web interface has to send one over a socket. Every
answer a closure would have given is resolved when Build runs: bounds that
depend on another setting are numbers by then, a row that a missing asset pack
makes unavailable is already marked Disabled with the reason, and a button is an
ID the caller dispatches rather than a func nobody can serialise.

The Model is a snapshot. The browser rebuilds it after a change because that
change can move another row's bounds or add and remove rows outright; ticking a
community asset pack adds a row for each mission in that pack.
*/
package form

// Kind is what an interface has to draw, and there are seven of them.
type Kind uint8

const (
	// Text is a line: a folder, a server name, a room address.
	Text Kind = iota + 1
	// Password is a line that is never shown back, only replaced.
	Password
	// Number is a whole number with a floor and a ceiling.
	Number
	// Toggle is yes or no.
	Toggle
	// Choice is one of Options.
	Choice
	// Action is a button. It carries no value, only what pressing it does.
	Action
	// Confirm is an Action that cannot be taken back, so the browser asks first.
	Confirm
)

// Option is one answer of a Choice: the value that is saved, and the line the
// player reads. The two are separate because "progression" is what the settings
// file holds and "Required for progression" is what the player is asked.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

/*
Field is one row, resolved for the settings it was built from.

Every kind uses Value, as text, including Number and Toggle. One field type
rather than seven keeps the interfaces from switching on Kind to find out where
the value is, and the text is what a <input> and a LineEdit both hand back
anyway. Apply is what turns it into an int or a bool, in one place, with the
error the player sees.
*/
type Field struct {
	ID    string `json:"id"`
	Kind  Kind   `json:"kind"`
	Label string `json:"label"`
	Help  string `json:"help"`

	// Group is the section of the page this row is in, and Bar marks an action
	// about the whole window rather than about this page. Both are placement an
	// interface may use or ignore; see Spec.
	Group string `json:"group,omitempty"`
	Bar   bool   `json:"bar,omitempty"`

	// Value is empty for Action and Confirm, which have nothing to hold.
	Value string `json:"value,omitempty"`

	// Placeholder is what a blank Text means, spelled out. It is not a default:
	// the setting stays empty and the launcher works the value out at run time.
	Placeholder string `json:"placeholder,omitempty"`

	// Low and High bound a Number. They are resolved here rather than declared
	// as constants because some of them depend on another setting: the mission
	// count's ceiling is however many missions the chosen tier leaves.
	Low  int `json:"low,omitempty"`
	High int `json:"high,omitempty"`

	// Options are the answers to a Choice, in the order they are offered.
	Options []Option `json:"options,omitempty"`

	// Hint is what a Toggle says beside the tick.
	// Warning is what a Confirm says while it is asking.
	Hint    string `json:"hint,omitempty"`
	Warning string `json:"warning,omitempty"`

	// HintOff is what a Toggle says when it is not ticked, where that is worth
	// a different word: "left out" reads better than an unticked "in the pool"
	// down a list of twenty-six missions. Empty means Hint either way, which is
	// most of them.
	HintOff string `json:"hint_off,omitempty"`

	// Browse marks a Text row naming a folder, so an interface with a file
	// picker offers one beside it and one without simply does not.
	Browse bool `json:"browse,omitempty"`

	// Deferred marks a row whose value is parsed rather than stored, so an
	// interface keeps what was typed while it is still half an answer instead
	// of refusing the character that made it one.
	Deferred bool `json:"deferred,omitempty"`

	// Disabled is a row the player can see and cannot use, with Reason saying
	// why. Hiding it instead would leave them looking for a setting that the
	// documentation says exists.
	Disabled bool   `json:"disabled,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// Tab is one page of rows. Intro is the paragraph above them, for a page whose
// rows do not explain themselves: Balancing needs to say that Valve tunes every
// wave for six defenders before "Robot health" means anything.
type Tab struct {
	Title string `json:"title"`
	Intro string `json:"intro,omitempty"`

	// Under is the page this one is a section of, empty for a top-level page.
	Under  string  `json:"under,omitempty"`
	Fields []Field `json:"fields"`
}

// Model is the whole screen for one set of settings.
type Model struct {
	Tabs []Tab `json:"tabs"`
}

// Field finds a row by ID. The interfaces use it to answer a click without
// walking the tabs themselves.
func (m Model) Field(id string) (Field, bool) {
	for _, tab := range m.Tabs {
		for _, f := range tab.Fields {
			if f.ID == id {
				return f, true
			}
		}
	}
	return Field{}, false
}

// Change is one answer, from any interface: a key, a click or a POST. Value is
// text for the same reason Field.Value is.
type Change struct {
	Field string `json:"field"`
	Value string `json:"value"`
}
