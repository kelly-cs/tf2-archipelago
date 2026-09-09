//go:build windows

package gui

import (
	"sort"
	"strconv"
	"strings"

	"github.com/lxn/walk"

	"github.com/m-this/tf2-archipelago/launcher/internal/form"
)

/*
poolModel is the mission pool as a checkable table.

form declares one Toggle per mission, which is the right row and the wrong
screen when there are twenty-six of them. The window has room for columns, so it
draws the same rows as a table: the label form wrote, whether it is in the pool,
and what a mission the launcher cannot play is missing.

Nothing here knows what a mission is. It holds the resolved rows and the tick
each one currently carries, and hands them back as changes under the IDs form
gave them, which is why adding a mission to the game changes nothing in this
file.
*/
type poolModel struct {
	walk.TableModelBase
	walk.SorterBase

	// One slice, not two. The rows and their ticks were a pair of parallel
	// slices until the headers started sorting, and a sort that reorders one
	// and not the other silently moves every tick onto a different mission.
	rows []poolRow
}

/*
	The table opens sorted by mission name, and that is walk's doing and ours.

A TableView asks a Sorter model to sort by its own default column the moment it
has one, and there is no way to say "no column" from here: the sorted column
lives on the view, unexported, and nothing sets it. So the table cannot both
sort on click and open in the order form declares the rows, which is map by map
and easiest first.

Given the choice, opening alphabetically is the one to take. Eighty-five rows is
what made EZKSupernova ask for this, and finding one mission by name is what
they were trying to do. The header carries the arrow, so the order is stated
rather than mysterious.
*/

// poolRow is one mission as the table holds it: the row form resolved, and
// whether it is ticked in.
type poolRow struct {
	form.Field
	inPool bool
}

func newPoolModel(model form.Model) *poolModel {
	pool := &poolModel{}
	for _, tab := range model.Tabs {
		for _, field := range tab.Fields {
			if !strings.HasPrefix(field.ID, poolPrefix) {
				continue
			}
			pool.rows = append(pool.rows, poolRow{Field: field, inPool: field.Value == "true"})
		}
	}
	return pool
}

func (p *poolModel) RowCount() int { return len(p.rows) }

func (p *poolModel) Value(row, col int) any {
	if row < 0 || row >= len(p.rows) {
		return ""
	}
	entry := p.rows[row]
	switch col {
	case 0:
		return entry.Label
	case 1:
		if entry.inPool {
			return entry.Hint
		}
		return entry.HintOff
	default:
		// The reason a row is unavailable, or nothing when it is fine. It is
		// the column a player scans to find out why a mission they installed a
		// pack for still cannot be ticked.
		return entry.Reason
	}
}

func (p *poolModel) Checked(row int) bool {
	return row >= 0 && row < len(p.rows) && p.rows[row].inPool
}

/*
	SetChecked refuses a mission the launcher cannot play, rather than taking the

tick and dropping it at Save. The Compatibility column says why.

Both paths publish. The "In the pool" column is read off this tick, and nothing
told the table view the row had changed, so the text kept its old answer until
Windows repainted the row for its own reasons, which is what passing the mouse
over it does. A refused tick was the worse half: the control drew the tick the
player clicked and the model had not taken it.
*/
func (p *poolModel) SetChecked(row int, checked bool) error {
	if row < 0 || row >= len(p.rows) {
		return nil
	}
	if p.rows[row].Disabled {
		p.PublishRowChanged(row)
		return nil
	}
	p.rows[row].inPool = checked
	p.PublishRowChanged(row)
	return nil
}

// setAll ticks or unticks every mission the launcher can play. A mission it
// cannot is left out either way: All would otherwise put a mission in the pool
// that the seed can draw and the server cannot load.
func (p *poolModel) setAll(inPool bool) {
	for i := range p.rows {
		p.rows[i].inPool = inPool && !p.rows[i].Disabled
	}
	p.PublishRowsReset()
}

/*
	Sort puts the table in the order the header that was clicked asks for.

The headers are a native list-view's and look pressable whether or not anything
is behind them, so a table that does not sort is a table that looks broken. With
eighty-five missions in it that is what EZKSupernova reported.

Column 0 sorts on the mission's name and not on the whole label, which begins
with the archive it came from: sorting on the label groups by archive, and
somebody looking for one mission by name is the reason the header was clicked.
*/
func (p *poolModel) Sort(col int, order walk.SortOrder) error {
	/* Minus one is the contract's "no column is to be sorted", and it is not a
	   corner case here: a TableView whose model sorts re-sorts on every row it
	   is told changed, so this runs on the first tick with no header ever
	   pressed. Reordering there would rearrange the table under the hand that
	   was ticking a box. */
	if col < 0 {
		return p.SorterBase.Sort(col, order)
	}
	sort.SliceStable(p.rows, func(i, j int) bool {
		if order == walk.SortDescending {
			return p.less(col, j, i)
		}
		return p.less(col, i, j)
	})
	return p.SorterBase.Sort(col, order)
}

func (p *poolModel) less(col, i, j int) bool {
	a, b := p.rows[i], p.rows[j]
	switch col {
	case 1:
		if a.inPool != b.inPool {
			return a.inPool
		}
	case 2:
		if a.Reason != b.Reason {
			return a.Reason < b.Reason
		}
	}
	// The name is the tie-break for every column, so rows that compare equal
	// keep one order instead of shuffling on each click.
	return missionSortKey(a.Label) < missionSortKey(b.Label)
}

// missionSortKey is a row's label without the archive it came from, folded for
// comparison, so "[Potato Archive] Void Voyage (mvm_null_b9c)" sorts under V.
func missionSortKey(label string) string {
	if _, name, found := strings.Cut(label, "] "); found {
		label = name
	}
	return strings.ToLower(label)
}

// apply writes the ticks back through form, under the IDs form gave the rows.
// By ID and never by position, which is what lets the table be sorted at all.
func (p *poolModel) apply(s form.State, env form.Env) (form.State, error) {
	for _, entry := range p.rows {
		next, err := form.Apply(s, env, form.Change{
			Field: entry.ID,
			Value: strconv.FormatBool(entry.inPool),
		})
		if err != nil {
			return s, err
		}
		s = next
	}
	return s, nil
}
