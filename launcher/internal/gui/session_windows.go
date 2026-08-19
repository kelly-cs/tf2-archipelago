//go:build windows

package gui

import (
	"strings"

	"github.com/lxn/walk"
	declarative "github.com/lxn/walk/declarative"

	"github.com/m-this/tf2-archipelago/launcher/internal/session"
)

// sessionTab is the view of the run: the multiworld line, the run's counters,
// and the missions with what the run has done to them. It is fed by
// watchSession and touched only on the UI thread.
type sessionTab struct {
	multiworld *walk.Label
	run        *walk.Label
	table      *walk.TableView
	switchBt   *walk.PushButton
	model      *missionsModel
	running    bool
}

func newSessionTab() *sessionTab {
	return &sessionTab{model: &missionsModel{}}
}

// page builds the tab. onSwitch gets the popfile of the selected mission.
func (t *sessionTab) page(onSwitch func(popFile string)) declarative.TabPage {
	return declarative.TabPage{
		Title:  "Session",
		Layout: declarative.VBox{},
		Children: []declarative.Widget{
			declarative.Label{AssignTo: &t.multiworld, Text: "The server is not running.", ToolTipText: "The bridge's connection to the Archipelago room."},
			declarative.Label{AssignTo: &t.run, Text: "", TextColor: colorMuted},
			declarative.TableView{
				AssignTo:         &t.table,
				Model:            t.model,
				AlternatingRowBG: true,
				StretchFactor:    1,
				ToolTipText:      "The run's missions in the order the seed drew them. Locked ones wait for their ticket.",
				Columns: []declarative.TableViewColumn{
					{Title: "#", Width: 30},
					{Title: "Mission", Width: 200},
					{Title: "Map", Width: 130},
					{Title: "Waves", Width: 50},
					{Title: "State", Width: 120},
				},
				StyleCell: func(style *walk.CellStyle) {
					if style.Row() < 0 || style.Row() >= len(t.model.missions) {
						return
					}
					mission := t.model.missions[style.Row()]
					switch {
					case mission.Cleared:
						style.TextColor = colorRunning
					case !mission.Unlocked:
						style.TextColor = colorMuted
					}
				},
				OnCurrentIndexChanged: func() { t.refreshButtons() },
			},
			declarative.Composite{
				Layout:  declarative.HBox{MarginsZero: true},
				MaxSize: declarative.Size{Height: 30},
				Children: []declarative.Widget{
					declarative.PushButton{
						AssignTo:    &t.switchBt,
						Text:        "Play this mission",
						ToolTipText: "Load the selected mission now, through the plugin's own switcher. A locked mission is refused.",
						OnClicked: func() {
							if popFile, ok := t.selected(); ok {
								onSwitch(popFile)
							}
						},
					},
					declarative.Label{Text: "In the game, !ap status prints the same picture and !mission the same list.", TextColor: colorMuted},
					declarative.HSpacer{},
				},
			},
		},
	}
}

func (t *sessionTab) selected() (string, bool) {
	index := t.table.CurrentIndex()
	if index < 0 || index >= len(t.model.missions) {
		return "", false
	}
	return t.model.missions[index].PopFile, true
}

func (t *sessionTab) setRunning(running bool) {
	t.running = running
	if !running {
		t.multiworld.SetText("The server is not running.")
		t.run.SetText("")
		t.model.set(nil)
	}
	t.refreshButtons()
}

func (t *sessionTab) refreshButtons() {
	popFile, ok := t.selected()
	t.switchBt.SetEnabled(t.running && ok)
	if ok {
		t.switchBt.SetText("Play " + popFile)
	} else {
		t.switchBt.SetText("Play this mission")
	}
}

// update takes a reading of the run, or the reason there is none.
func (t *sessionTab) update(snapshot session.Snapshot, err error) {
	if !t.running {
		return
	}
	if err != nil {
		t.multiworld.SetText("Bridge: " + err.Error())
		return
	}
	health := snapshot.Health
	t.multiworld.SetText("Multiworld: " + health.Summary())

	var run []string
	if health.Seed != "" {
		run = append(run, "seed "+health.Seed)
	}
	if health.LastCheck != "" {
		run = append(run, "last check: "+health.LastCheck)
	}
	if health.DeathLink {
		run = append(run, "Death Link on")
	}
	if health.GoalSent {
		run = append(run, "goal reached")
	}
	t.run.SetText(strings.Join(run, "   "))
	t.model.set(snapshot.Missions)
	t.refreshButtons()
}

// missionsModel is the table's data: the run's missions, in seed order.
type missionsModel struct {
	walk.TableModelBase
	missions []session.Mission
}

// set replaces the rows. A same-sized list is a change of rows rather than a
// reset, which keeps the player's selection through the five-second refresh.
func (m *missionsModel) set(missions []session.Mission) {
	same := len(missions) == len(m.missions)
	m.missions = missions
	if same && len(missions) > 0 {
		m.PublishRowsChanged(0, len(missions)-1)
		return
	}
	m.PublishRowsReset()
}

func (m *missionsModel) RowCount() int { return len(m.missions) }

func (m *missionsModel) Value(row, col int) any {
	mission := m.missions[row]
	switch col {
	case 0:
		return row + 1
	case 1:
		return mission.Name
	case 2:
		return mission.Map
	case 3:
		return mission.Waves
	default:
		return missionState(mission)
	}
}

func missionState(mission session.Mission) string {
	switch {
	case mission.Cleared:
		return "cleared"
	case mission.Unlocked:
		return "unlocked"
	default:
		return "locked"
	}
}
