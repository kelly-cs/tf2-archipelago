//go:build windows

package gui

import (
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

	rows   []form.Field
	inPool []bool
}

func newPoolModel(model form.Model) *poolModel {
	pool := &poolModel{}
	for _, tab := range model.Tabs {
		for _, field := range tab.Fields {
			if !strings.HasPrefix(field.ID, poolPrefix) {
				continue
			}
			pool.rows = append(pool.rows, field)
			pool.inPool = append(pool.inPool, field.Value == "true")
		}
	}
	return pool
}

func (p *poolModel) RowCount() int { return len(p.rows) }

func (p *poolModel) Value(row, col int) any {
	if row < 0 || row >= len(p.rows) {
		return ""
	}
	field := p.rows[row]
	switch col {
	case 0:
		return field.Label
	case 1:
		if p.inPool[row] {
			return field.Hint
		}
		return field.HintOff
	default:
		// The reason a row is unavailable, or nothing when it is fine. It is
		// the column a player scans to find out why a mission they installed a
		// pack for still cannot be ticked.
		return field.Reason
	}
}

func (p *poolModel) Checked(row int) bool {
	return row >= 0 && row < len(p.inPool) && p.inPool[row]
}

// SetChecked refuses a mission the launcher cannot play, rather than taking the
// tick and dropping it at Save. The Compatibility column says why.
func (p *poolModel) SetChecked(row int, checked bool) error {
	if row < 0 || row >= len(p.inPool) {
		return nil
	}
	if p.rows[row].Disabled {
		return nil
	}
	p.inPool[row] = checked
	return nil
}

// setAll ticks or unticks every mission the launcher can play. A mission it
// cannot is left out either way: All would otherwise put a mission in the pool
// that the seed can draw and the server cannot load.
func (p *poolModel) setAll(inPool bool) {
	for i, field := range p.rows {
		if field.Disabled {
			p.inPool[i] = false
			continue
		}
		p.inPool[i] = inPool
	}
	p.PublishRowsReset()
}

// apply writes the ticks back through form, under the IDs form gave the rows.
func (p *poolModel) apply(s form.State, env form.Env) (form.State, error) {
	for i, field := range p.rows {
		next, err := form.Apply(s, env, form.Change{
			Field: field.ID,
			Value: strconv.FormatBool(p.inPool[i]),
		})
		if err != nil {
			return s, err
		}
		s = next
	}
	return s, nil
}
