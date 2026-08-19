//go:build windows

package gui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/lxn/walk"
	declarative "github.com/lxn/walk/declarative"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/assets"
	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/debugbundle"
	"github.com/m-this/tf2-archipelago/launcher/internal/generate"
	"github.com/m-this/tf2-archipelago/launcher/internal/runshape"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/winproc"
)

// labelWidth keeps every tab's label column the same width, so the fields do
// not jump when the player switches tab.
const labelWidth = 150

// steamNetworkingTab hides the tab that chooses where players come from.
//
// The relay half of it has never been proved end to end: a server needs a
// login token before Valve hands out a relayed address, and no run has yet
// gone from that address to a Team Fortress 2 client that joined. Until one
// has, the tab is off and the launcher offers the local network, which is
// what it offered before and what is known to work.
//
// TF2AP_STEAM_NETWORKING=1 turns it on for whoever is doing that testing.
// SRCDS_REACH still works with the tab off: it is the tab that is hidden, not
// the setting.
var steamNetworkingTab = settings.Truthy(os.Getenv("TF2AP_STEAM_NETWORKING"))

// runSettingsDialog asks for the values worth changing between evenings, in
// five tabs: what the run is, which missions it may draw, where the room is,
// how the game server behaves, and what the bots play. A sixth, who can reach
// it, appears only where steamNetworkingTab is on. Every row carries a
// tooltip, because a name alone does not say what a difficulty floor or a
// login token is.
//
// It returns the edited settings and whether the player accepted them.
func runSettingsDialog(owner walk.Form, s settings.Settings, repair func() ([]string, error)) (settings.Settings, bool, error) {
	var (
		dialog *walk.Dialog
		accept *walk.PushButton
		cancel *walk.PushButton

		testBox  *walk.CheckBox
		roomEdit *walk.LineEdit
		roomWarn *walk.Label
		roomPass *walk.LineEdit
		slotEdit *walk.LineEdit

		tierBox   *walk.ComboBox
		missions  *walk.NumberEdit
		goalBox   *walk.ComboBox
		sanityPct *walk.NumberEdit
		deathLink *walk.CheckBox

		startBox *walk.ComboBox
		poolView *walk.TableView

		appEdit *walk.LineEdit

		nameEdit  *walk.LineEdit
		passEdit  *walk.LineEdit
		portEdit  *walk.NumberEdit
		adminEdit *walk.LineEdit
		botsBox   *walk.CheckBox
		botsSize  *walk.NumberEdit
		buysBox   *walk.CheckBox
		classBox  = make([]*walk.CheckBox, len(botloadout.Classes))
		loadoutBx = make([]*walk.ComboBox, len(botloadout.Classes))

		reachLan   *walk.RadioButton
		reachSteam *walk.RadioButton
		reachPort  *walk.RadioButton
		reachHelp  *walk.TextLabel
		tokenEdit  *walk.LineEdit
		tokenWarn  *walk.Label
	)

	tiers := runshape.Tiers()
	tierLabels := make([]string, 0, len(tiers))
	for _, tier := range tiers {
		tierLabels = append(tierLabels, tier.Label())
	}
	goals := runshape.Goals()
	goalLabels := make([]string, 0, len(goals))
	for _, goal := range goals {
		goalLabels = append(goalLabels, goal.Label())
	}
	choices := runshape.MissionChoices()
	choiceLabels := make([]string, 0, len(choices))
	for _, choice := range choices {
		choiceLabels = append(choiceLabels, choice.Label)
	}
	pool := newPoolModel(s.MvmExcludedMissions)

	current := settings.Room{Host: s.APHost, Port: s.APPort}
	edited := s

	label := func(text, help string) declarative.Label {
		return declarative.Label{
			Text:        text,
			MinSize:     declarative.Size{Width: labelWidth},
			ToolTipText: help,
		}
	}

	// collect reads the widgets into a copy of the settings. Save uses it, and
	// so does the button that writes the player file: both have to work from
	// what is on screen rather than from what was saved last time.
	collect := func() (settings.Settings, error) {
		next := s
		next.TestMode = testBox.Checked()
		room, err := settings.ParseRoom(roomEdit.Text())
		if err != nil {
			// Test mode never dials a real room, so it does not need one.
			if !next.TestMode {
				return next, err
			}
			room = settings.Room{}
		}
		next.APHost, next.APPort, next.APTls = room.Host, room.Port, room.TLS
		next.APPassword = roomPass.Text()
		next.APSlotName = strings.TrimSpace(slotEdit.Text())

		next.MvmDifficulty = tiers[max(tierBox.CurrentIndex(), 0)].Key
		next.MvmMissionCount = int(missions.Value())
		next.MvmGoal = goals[max(goalBox.CurrentIndex(), 0)].Key
		next.MvmMissionsanityPct = int(sanityPct.Value())
		next.MvmDeathLink = deathLink.Checked()
		next.MvmExcludedMissions = pool.excluded()
		next.ArchipelagoDir = strings.TrimSpace(appEdit.Text())

		next.SrcdsStartMission = choices[max(startBox.CurrentIndex(), 0)].PopFile

		next.SrcdsHostname = strings.TrimSpace(nameEdit.Text())
		next.SrcdsPw = strings.TrimSpace(passEdit.Text())
		next.SrcdsPort = int(portEdit.Value())
		next.SrcdsAdminSteamIDs = strings.TrimSpace(adminEdit.Text())
		// With the tab hidden there are no widgets to read, and the reach keeps
		// whatever the config file or the environment set it to.
		if steamNetworkingTab {
			next.SrcdsReach = checkedReach(reachSteam, reachPort)
			next.SrcdsToken = strings.TrimSpace(tokenEdit.Text())
		}
		next.SrcdsBots = botsBox.Checked()
		next.SrcdsBotTeamSize = int(botsSize.Value())
		next.BotUpgradesChat = buysBox.Checked()
		next.SrcdsBotClassBlacklist = nil
		next.SrcdsBotLoadouts = make(map[string]string)
		for i, class := range botloadout.Classes {
			if !classBox[i].Checked() {
				next.SrcdsBotClassBlacklist = append(next.SrcdsBotClassBlacklist, class.Key)
			}
			if pick := class.Loadouts[max(loadoutBx[i].CurrentIndex(), 0)]; pick.Key != botloadout.StockKey {
				next.SrcdsBotLoadouts[class.Key] = pick.Key
			}
		}
		return next, nil
	}

	// The Steam Networking tab is built only when it is turned on, so the
	// widgets behind it stay nil and everything that reads them has to check.
	var extraPages []declarative.TabPage
	if steamNetworkingTab {
		extraPages = append(extraPages, declarative.TabPage{
			Title:  "Steam Networking",
			Layout: declarative.Grid{Columns: 2},
			Children: []declarative.Widget{
				label("Who can reach it", "Where the server takes connections from. A server is on your own network until you change this: it is not open to anybody else by default."),
				// The three buttons are consecutive children of one Composite
				// on purpose. walk groups a radio button with the sibling
				// before it, so a label in between would leave three groups of
				// one, all tickable.
				declarative.Composite{
					// Near, or a VBox centres each button on its own text and
					// the three ends up as a ragged stack.
					Layout: declarative.VBox{MarginsZero: true, SpacingZero: true, Alignment: declarative.AlignHNearVCenter},
					Children: []declarative.Widget{
						declarative.RadioButton{AssignTo: &reachLan, Text: settings.ReachLan.Label()},
						declarative.RadioButton{AssignTo: &reachSteam, Text: settings.ReachSteam.Label()},
						declarative.RadioButton{AssignTo: &reachPort, Text: settings.ReachPort.Label()},
					},
				},
				declarative.Label{Text: ""},
				declarative.TextLabel{
					AssignTo: &reachHelp,
					Text:     s.SrcdsReach.Help(),
					MinSize:  declarative.Size{Width: 330},
				},
				label("Login token", "A Game Server Login Token for app id 440, from steamcommunity.com/dev/managegameservers. The server logs in to Steam with it. Both of the reaches that leave your network need a real one."),
				declarative.LineEdit{AssignTo: &tokenEdit, Text: s.SrcdsToken, CueBanner: "0"},
				declarative.Label{Text: ""},
				declarative.Label{AssignTo: &tokenWarn, Text: "", MaxSize: declarative.Size{Height: 18}},
				declarative.TextLabel{
					Text: "Over Steam, the server prints its address in the log every time it starts, " +
						"in the form connect 169.254.13.42:20232. It is a new address on every start, " +
						"so send your friends the line from the log rather than one you wrote down.",
					ColumnSpan: 2,
					MinSize:    declarative.Size{Width: 470},
				},
			},
		})
	}

	err := declarative.Dialog{
		AssignTo:     &dialog,
		Title:        "Settings",
		CancelButton: &cancel,
		Size:         declarative.Size{Width: 700, Height: 560},
		MinSize:      declarative.Size{Width: 600, Height: 480},
		Layout:       declarative.VBox{},
		Children: []declarative.Widget{
			declarative.TabWidget{
				Pages: append([]declarative.TabPage{
					{
						Title:  "Player options",
						Layout: declarative.Grid{Columns: 2},
						Children: []declarative.Widget{
							label("Easiest tier", "The easiest tier a mission may come from. Harder tiers are always in as well, so the pool shrinks as this rises. Expert leaves four, because Valve made only three expert missions and one haunted one."),
							declarative.ComboBox{
								AssignTo:    &tierBox,
								Model:       tierLabels,
								Value:       tierLabel(tiers, s.MvmDifficulty),
								ToolTipText: "The number is the pool to draw from, not the length of a run.",
							},
							label("Missions used", "How many missions this run uses, out of the pool above. Eight is about fifty waves, which is one evening for a team that knows the mode."),
							declarative.NumberEdit{
								AssignTo:    &missions,
								Value:       float64(s.MvmMissionCount),
								MinValue:    1,
								MaxValue:    float64(runshape.MissionsInPool(s.MvmDifficulty)),
								Decimals:    0,
								ToolTipText: "Asking for more than the pool holds gives you the whole pool.",
							},
							label("Goal", "What ends the run. Final Boss marks the hardest mission the run drew, and clearing it wins. Missionsanity asks for a share of the missions instead, in any order."),
							declarative.ComboBox{AssignTo: &goalBox, Model: goalLabels, Value: goalLabel(goals, s.MvmGoal)},
							label("Missionsanity share", "How much of the run Missionsanity asks for, in percent. It rounds up, and the Final Boss goal ignores it."),
							declarative.NumberEdit{
								AssignTo: &sanityPct, Value: float64(s.MvmMissionsanityPct),
								MinValue: 10, MaxValue: 100, Decimals: 0,
							},
							label("Death Link", "A lost wave kills every other player in the multiworld who has Death Link on, and their deaths wipe your team."),
							declarative.CheckBox{AssignTo: &deathLink, Text: "share deaths", Checked: s.MvmDeathLink},
							label("Archipelago app", "Where the Archipelago app is installed. Leave it blank and the launcher looks where the installer puts it. Set it when the app is on another drive, or in a folder of your own."),
							declarative.Composite{
								Layout: declarative.HBox{MarginsZero: true},
								Children: []declarative.Widget{
									declarative.LineEdit{
										AssignTo:  &appEdit,
										Text:      s.ArchipelagoDir,
										CueBanner: defaultAppDir(),
									},
									declarative.PushButton{
										Text:        "Browse",
										MinSize:     declarative.Size{Width: 70},
										ToolTipText: "Pick the folder holding ArchipelagoGenerate.exe.",
										OnClicked:   func() { browseForApp(dialog, appEdit) },
									},
								},
							},
							declarative.Label{
								Text:        "These are the options the Archipelago website calls player options. They go in tf2.yaml, which the seed is generated from. The Missions tab picks which missions the run may draw.",
								ColumnSpan:  2,
								ToolTipText: "Change them here, then generate again for a new seed. The current run keeps the shape it was generated with.",
							},
							declarative.Composite{
								Layout:     declarative.HBox{MarginsZero: true},
								ColumnSpan: 2,
								MaxSize:    declarative.Size{Height: 32},
								Children: []declarative.Widget{
									declarative.PushButton{
										Text:        "Generate seed",
										ToolTipText: "Make the seed with the Archipelago app installed on this machine: the launcher installs the world file into it, writes the player file, runs the generator and opens the folder with the archive. Upload that archive at archipelago.gg/uploads to open a room.",
										OnClicked:   func() { generateSeed(dialog, collect) },
									},
									declarative.PushButton{
										Text:        "Open tf2.yaml",
										ToolTipText: "Write the player file from what is on screen, then open it. Copy it into the Archipelago app's Players folder to generate the seed.",
										OnClicked:   func() { openPlayerFile(dialog, collect) },
									},
									declarative.PushButton{
										Text:        "Open the folder",
										ToolTipText: "The install root: the game files, the player file, the log and the run's state.",
										OnClicked:   func() { openFolder(dialog, s.InstallRoot) },
									},
									declarative.HSpacer{},
								},
							},
						},
					},
					{
						Title:  "Missions",
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Composite{
								Layout:  declarative.HBox{MarginsZero: true},
								MaxSize: declarative.Size{Height: 28},
								Children: []declarative.Widget{
									label("Start mission", "The mission the server loads first, as map - mission. If the run has not unlocked it, the plugin moves to the first mission it has."),
									declarative.ComboBox{
										AssignTo: &startBox,
										Model:    choiceLabels,
										Value:    runshape.MissionLabel(s.SrcdsStartMission),
									},
								},
							},
							declarative.Label{
								Text:        "Missions the run may draw. Untick one to keep it out of every seed generated from here: Caliginous Caper is one wave of 666 robots and an hour on its own. The tier above still applies.",
								ToolTipText: "This is the excluded_missions option in tf2.yaml.",
							},
							declarative.TableView{
								AssignTo:         &poolView,
								Model:            pool,
								CheckBoxes:       true,
								AlternatingRowBG: true,
								StretchFactor:    1,
								Columns: []declarative.TableViewColumn{
									{Title: "Mission", Width: 200},
									{Title: "Map", Width: 130},
									{Title: "Tier", Width: 90},
									{Title: "Waves", Width: 50},
								},
							},
							declarative.Composite{
								Layout:  declarative.HBox{MarginsZero: true},
								MaxSize: declarative.Size{Height: 30},
								Children: []declarative.Widget{
									declarative.PushButton{Text: "All", OnClicked: func() { pool.setAll(true) }},
									declarative.PushButton{Text: "None", OnClicked: func() { pool.setAll(false) }},
									declarative.HSpacer{},
								},
							},
						},
					},
					{
						Title:  "Archipelago room",
						Layout: declarative.Grid{Columns: 2},
						Children: []declarative.Widget{
							label("Test mode", "Play without Archipelago at all. The launcher serves a multiworld of one, makes up a seed from your run options, and hands out an unlock for every wave you clear. Other players are simulated: they find things, send you things and die."),
							declarative.CheckBox{AssignTo: &testBox, Text: "no room, no seed, just play", Checked: s.TestMode},
							label("Room address", "The line from your room page on archipelago.gg: host and port. A room on your own machine works too, as localhost:38281."),
							declarative.LineEdit{AssignTo: &roomEdit, Text: current.String(), CueBanner: "archipelago.gg:12345"},
							declarative.Label{Text: ""},
							declarative.Label{AssignTo: &roomWarn, Text: "", MaxSize: declarative.Size{Height: 18}},
							label("Room password", "Only if the room asks for one. Leave it blank otherwise."),
							declarative.LineEdit{AssignTo: &roomPass, Text: s.APPassword, PasswordMode: true, CueBanner: "optional"},
							label("Slot name", "The name this server plays under in the multiworld. It has to match the name in tf2.yaml, and the launcher keeps the two in step."),
							declarative.LineEdit{AssignTo: &slotEdit, Text: s.APSlotName},
							declarative.Label{
								Text:        "One slot covers the whole game server: everybody playing here shares its unlocks.",
								ColumnSpan:  2,
								ToolTipText: "Another player who wants a slot of their own plays another game in the same room, from the Archipelago app.",
							},
						},
					},
					{
						Title:  "Game server",
						Layout: declarative.Grid{Columns: 2},
						Children: []declarative.Widget{
							label("Server name", "What the server calls itself in the player list."),
							declarative.LineEdit{AssignTo: &nameEdit, Text: s.SrcdsHostname},
							label("Server password", "What your friends type before connect. Blank means anybody with the address can join."),
							declarative.LineEdit{AssignTo: &passEdit, Text: s.SrcdsPw, CueBanner: "optional, blank for none"},
							label("Game port", "UDP and TCP, 27015 by default. Who can reach it is on the Steam Networking tab."),
							declarative.NumberEdit{
								AssignTo: &portEdit, Value: float64(s.SrcdsPort),
								MinValue: 1024, MaxValue: 65535, Decimals: 0,
							},
							label("Admins by Steam id", "Who may run the admin commands, separated by commas. Either form works: the 17 digit id from a profile URL, or SourceMod's STEAM_0:1:26975537."),
							declarative.LineEdit{AssignTo: &adminEdit, Text: s.SrcdsAdminSteamIDs, CueBanner: "76561198014216803, ..."},
						},
					},
					{
						Title:    "Bots",
						Layout:   declarative.Grid{Columns: 2},
						Children: botsRows(s, label, &botsBox, &botsSize, &buysBox, classBox, loadoutBx),
					},
				}, extraPages...),
			},
			declarative.Composite{
				Layout:  declarative.HBox{},
				MaxSize: declarative.Size{Height: 34},
				Children: []declarative.Widget{
					declarative.PushButton{
						Text:        "Debug logs",
						ToolTipText: "Put the logs, the settings without their passwords, and the player file in one zip, for sending to whoever is helping you.",
						OnClicked:   func() { saveDebugBundle(dialog, s) },
					},
					declarative.PushButton{
						Text:        "Repair",
						ToolTipText: "Throw SteamCMD and the mods away and fetch them again. Keeps the game files and the run.",
						OnClicked:   func() { runRepair(dialog, repair) },
					},
					declarative.HSpacer{},
					declarative.PushButton{AssignTo: &accept, Text: "Save", OnClicked: func() {
						// Read every field before the dialog closes: closing it
						// destroys its children, and a destroyed LineEdit reads
						// back empty.
						//
						// The address is checked here rather than as the player
						// types. A disabled button with no explanation is a dead
						// end, and a paste with the mouse sends no keystroke to
						// check on.
						next, err := collect()
						if err != nil {
							roomWarn.SetText(err.Error())
							return
						}
						// A reach that leaves the network with no token is a
						// server every client is refused from, and nothing on
						// screen would say why. Refuse the save instead.
						//
						// Only where the tab is on. With it off the reach came from
						// the environment or the config file, and refusing would trap
						// the player in a dialog holding nothing that could fix it.
						if steamNetworkingTab {
							if complaint := tokenComplaint(next.SrcdsReach, next.SrcdsToken); complaint != "" {
								tokenWarn.SetText(complaint)
								return
							}
						}
						edited = next
						dialog.Accept()
					}},
					declarative.PushButton{AssignTo: &cancel, Text: "Cancel", OnClicked: func() { dialog.Cancel() }},
				},
			},
		},
	}.Create(owner)
	if err != nil {
		return s, false, err
	}

	// The help under the buttons, and the complaint about a missing token. Both
	// follow the selection, because a reach the player cannot use yet has to
	// say so here rather than in the server log twenty minutes later.
	if steamNetworkingTab {
		explainReach := func() {
			reach := checkedReach(reachSteam, reachPort)
			reachHelp.SetText(reach.Help())
			tokenWarn.SetText(tokenComplaint(reach, tokenEdit.Text()))
		}
		for _, button := range []*walk.RadioButton{reachLan, reachSteam, reachPort} {
			button.CheckedChanged().Attach(explainReach)
		}
		tokenEdit.TextChanged().Attach(explainReach)
		switch s.SrcdsReach {
		case settings.ReachSteam:
			reachSteam.SetChecked(true)
		case settings.ReachPort:
			reachPort.SetChecked(true)
		default:
			reachLan.SetChecked(true)
		}
		explainReach()
	}

	// A harder floor leaves fewer missions to draw from, so the count a run can
	// ask for follows the tier.
	tierBox.CurrentIndexChanged().Attach(func() {
		pool := tiers[max(tierBox.CurrentIndex(), 0)].Missions
		_ = missions.SetRange(1, float64(pool))
		if missions.Value() > float64(pool) {
			_ = missions.SetValue(float64(pool))
		}
	})

	// The complaint under the address: what is missing, or that test mode
	// makes it optional. Cleared as soon as the address looks right.
	explain := func() {
		switch {
		case testBox.Checked():
			roomWarn.SetText("not needed in test mode: the launcher serves its own room")
		case roomEdit.Text() == "":
			roomWarn.SetText("paste the address from your Archipelago room page")
		default:
			if _, err := settings.ParseRoom(roomEdit.Text()); err == nil {
				roomWarn.SetText("")
			}
		}
	}
	roomEdit.TextChanged().Attach(explain)
	testBox.CheckedChanged().Attach(explain)
	explain()

	if dialog.Run() != walk.DlgCmdOK {
		return s, false, nil
	}
	return edited, true, nil
}

// botsRows is the Bots tab: whether the bots play, how many, which classes
// they may pick, and what each class holds.
func botsRows(
	s settings.Settings, label func(text, help string) declarative.Label,
	botsBox **walk.CheckBox, botsSize **walk.NumberEdit, buysBox **walk.CheckBox,
	classBox []*walk.CheckBox, loadoutBx []*walk.ComboBox,
) []declarative.Widget {
	rows := []declarative.Widget{
		label("Defender bots", "Fill the RED team with bots that play, so a wave balanced for six is winnable by fewer. They pick classes, fight and buy their own upgrades. A bot steps aside when a friend joins."),
		declarative.CheckBox{AssignTo: botsBox, Text: "fill the RED team", Checked: s.SrcdsBots},
		label("Fill RED to", "How many players RED holds, humans included. Lower it for a harder run."),
		declarative.NumberEdit{
			AssignTo: botsSize, Value: float64(s.SrcdsBotTeamSize),
			MinValue: 1, MaxValue: 6, Decimals: 0,
		},
		label("Purchases in chat", "Write what the bots buy at the upgrade station to the chat, since the game no longer lets you inspect a teammate's upgrades. One line per purchase, so it is off by default."),
		declarative.CheckBox{AssignTo: buysBox, Text: "say what the bots buy", Checked: s.BotUpgradesChat},
		declarative.Label{
			Text:       "Classes the bots may play, and the weapons each class holds. Bots are poor snipers and spies; untick a class and they never pick it. Stock weapons are the mod's own default.",
			ColumnSpan: 2,
		},
	}
	for i, class := range botloadout.Classes {
		labels := make([]string, 0, len(class.Loadouts))
		for _, loadout := range class.Loadouts {
			labels = append(labels, loadout.Label())
		}
		rows = append(rows,
			declarative.CheckBox{
				AssignTo:    &classBox[i],
				Text:        class.Name,
				Checked:     !slices.Contains(s.SrcdsBotClassBlacklist, class.Key),
				MinSize:     declarative.Size{Width: labelWidth},
				ToolTipText: "Unticked, the bots never play " + class.Name + ".",
			},
			declarative.ComboBox{
				AssignTo:    &loadoutBx[i],
				Model:       labels,
				Value:       class.LoadoutByKey(s.SrcdsBotLoadouts[class.Key]).Label(),
				ToolTipText: "The weapons a " + class.Name + " bot spawns with.",
			},
		)
	}
	return rows
}

// poolModel is the Missions tab's table: every mission the tables know, ticked
// when the run may draw it. The unticked ones are the excluded_missions
// option.
type poolModel struct {
	walk.TableModelBase
	missions []gamedata.Mission
	inPool   []bool
}

func newPoolModel(excluded []string) *poolModel {
	model := &poolModel{missions: gamedata.Missions, inPool: make([]bool, len(gamedata.Missions))}
	for i, mission := range model.missions {
		model.inPool[i] = !slices.Contains(excluded, mission.PopFile)
	}
	return model
}

func (m *poolModel) RowCount() int { return len(m.missions) }

func (m *poolModel) Value(row, col int) any {
	mission := m.missions[row]
	switch col {
	case 0:
		return mission.Name
	case 1:
		played, _ := gamedata.MapByID(mission.Map)
		return played.Name
	case 2:
		return mission.Difficulty.String()
	default:
		return int(mission.Waves)
	}
}

func (m *poolModel) Checked(row int) bool { return m.inPool[row] }

func (m *poolModel) SetChecked(row int, checked bool) error {
	m.inPool[row] = checked
	return nil
}

func (m *poolModel) setAll(checked bool) {
	for i := range m.inPool {
		m.inPool[i] = checked
	}
	m.PublishRowsReset()
}

// excluded is the popfiles the player unticked, in table order.
func (m *poolModel) excluded() []string {
	var out []string
	for i, mission := range m.missions {
		if !m.inPool[i] {
			out = append(out, mission.PopFile)
		}
	}
	return out
}

// generateSeed makes the seed with the Archipelago app and opens the folder the
// archive landed in. It runs off the UI thread and reports through message
// boxes, because the dialog is modal and the log view is behind it.
func generateSeed(owner walk.Form, collect func() (settings.Settings, error)) {
	next, err := collect()
	if err != nil {
		walk.MsgBox(owner, "Generate seed", err.Error(), walk.MsgBoxIconError)
		return
	}
	if _, err := generate.FindApp(next.ArchipelagoDir); err != nil {
		walk.MsgBox(owner, "Generate seed",
			"The Archipelago app was not found.\n\nThe launcher looked in:\n"+
				"    "+strings.Join(generate.SearchPath(next.ArchipelagoDir), "\n    ")+
				"\n\nIf the app is somewhere else, put its folder in Archipelago app "+
				"above and press this again. If it is not installed, get it from "+
				"github.com/ArchipelagoMW/Archipelago/releases.",
			walk.MsgBoxIconWarning)
		return
	}

	var lines []string
	result, err := generate.Run(context.Background(), generate.Options{
		Settings:           next,
		AppDir:             next.ArchipelagoDir,
		Apworld:            assets.Apworld(),
		ArchipelagoVersion: assets.ArchipelagoVersion,
		Log:                func(line string) { lines = append(lines, line) },
	})
	if err != nil {
		tail := lines
		if len(tail) > 12 {
			tail = tail[len(tail)-12:]
		}
		walk.MsgBox(owner, "Generate seed",
			err.Error()+"\n\n"+strings.Join(tail, "\n"), walk.MsgBoxIconError)
		return
	}
	walk.MsgBox(owner, "Generate seed",
		"Seed written to\n"+result.Archive+"\n\nUpload it at archipelago.gg/uploads to open a room, "+
			"then paste the room address into the Archipelago room tab.",
		walk.MsgBoxIconInformation)
	_ = winproc.Open(filepath.Dir(result.Archive))
}

// openPlayerFile writes the player file from what is on screen and opens it.
// Writing first is the point: a player who edits the options then presses this
// wants to see those options, not the ones from the last save.
func openPlayerFile(owner walk.Form, collect func() (settings.Settings, error)) {
	next, err := collect()
	if err != nil {
		walk.MsgBox(owner, "Open tf2.yaml", err.Error(), walk.MsgBoxIconError)
		return
	}
	path, err := settings.WritePlayerFile(next, assets.ArchipelagoVersion)
	if err != nil {
		walk.MsgBox(owner, "Open tf2.yaml", err.Error(), walk.MsgBoxIconError)
		return
	}
	if err := winproc.Open(path); err != nil {
		walk.MsgBox(owner, "Open tf2.yaml", err.Error(), walk.MsgBoxIconError)
	}
}

// browseForApp asks for the app's folder and puts it in the box. The dialog
// starts wherever the box points, so a second try opens where the first one
// left off rather than at the desktop.
func browseForApp(owner walk.Form, edit *walk.LineEdit) {
	dialog := walk.FileDialog{
		Title:          "Where is the Archipelago app?",
		InitialDirPath: strings.TrimSpace(edit.Text()),
	}
	accepted, err := dialog.ShowBrowseFolder(owner)
	if err != nil {
		walk.MsgBox(owner, "Archipelago app", err.Error(), walk.MsgBoxIconError)
		return
	}
	if !accepted || dialog.FilePath == "" {
		return
	}
	_ = edit.SetText(dialog.FilePath)
}

// defaultAppDir is the first place the launcher looks, shown as the box's
// placeholder so a blank field says what it means.
func defaultAppDir() string {
	if dirs := generate.SearchPath(""); len(dirs) > 0 {
		return dirs[0]
	}
	return ""
}

func openFolder(owner walk.Form, path string) {
	if err := winproc.Open(path); err != nil {
		walk.MsgBox(owner, "Open the folder", err.Error(), walk.MsgBoxIconError)
	}
}

// saveDebugBundle writes the zip a play-tester sends on, and opens the folder
// so they can find it.
func saveDebugBundle(owner walk.Form, s settings.Settings) {
	path, err := debugbundle.Write(s, assets.Versions(), time.Now())
	if err != nil {
		walk.MsgBox(owner, "Debug logs", err.Error(), walk.MsgBoxIconError)
		return
	}
	walk.MsgBox(owner, "Debug logs",
		"Wrote "+path+"\n\nIt holds the launcher log, the SourceMod logs, the "+
			"server console, the player file and the settings. The passwords are "+
			"not in it.",
		walk.MsgBoxIconInformation)
	_ = winproc.Open(filepath.Dir(path))
}

// runRepair throws away SteamCMD, the mods and Steam's download record, for a
// player whose install will not go through. The next Start puts them back.
//
// The caller stops the server and any install first, so the button works on the
// first press rather than the third.
func runRepair(owner walk.Form, repair func() ([]string, error)) {
	answer := walk.MsgBox(owner, "Repair",
		"This stops the server, then removes SteamCMD, the mods and Steam's "+
			"record of the download. The next Start fetches them again.\n\n"+
			"It keeps the game files and the run: no 14 GB download, no lost checks.",
		walk.MsgBoxOKCancel|walk.MsgBoxIconQuestion)
	if answer != walk.DlgCmdOK {
		return
	}
	removed, err := repair()
	switch {
	case err != nil:
		walk.MsgBox(owner, "Repair", err.Error(), walk.MsgBoxIconError)
	case len(removed) == 0:
		walk.MsgBox(owner, "Repair", "Nothing to remove.", walk.MsgBoxIconInformation)
	default:
		walk.MsgBox(owner, "Repair",
			"Removed:\n"+strings.Join(removed, "\n")+"\n\nPress Start when you are ready.",
			walk.MsgBoxIconInformation)
	}
}

// tokenComplaint says what is wrong with a reach and a token together, or ""
// when they go together. One sentence, shown under the token field and checked
// again on Save, so the answer is the same in both places.
func tokenComplaint(reach settings.Reach, token string) string {
	if reach.NeedsToken() && !settings.HasToken(token) {
		return "this one needs a login token, or every player is refused"
	}
	return ""
}

// checkedReach reads the selection back. Nothing checked means the private
// default, which is the answer that cannot open a server by mistake.
func checkedReach(steam, port *walk.RadioButton) settings.Reach {
	switch {
	case steam.Checked():
		return settings.ReachSteam
	case port.Checked():
		return settings.ReachPort
	default:
		return settings.ReachLan
	}
}

func tierLabel(tiers []runshape.Tier, key string) string {
	for _, tier := range tiers {
		if tier.Key == key {
			return tier.Label()
		}
	}
	if len(tiers) == 0 {
		return ""
	}
	return tiers[0].Label()
}

func goalLabel(goals []runshape.Goal, key string) string {
	for _, goal := range goals {
		if goal.Key == key {
			return goal.Label()
		}
	}
	return fmt.Sprintf("%v", goals[0].Label())
}
