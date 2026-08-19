/*
The settings, in the six tabs the window uses: what the run is, which missions
it may draw, where the room is, how the game server behaves, what the bots
play, and who can join.

The window puts these in a modal dialog with a control per answer. Here they
are a list per tab, one row each, and the keys do what the mouse does there.
Nothing is saved until Save: cancelling leaves the file alone, the way closing
a dialog does.
*/
package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/assets"
	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/debugbundle"
	"github.com/m-this/tf2-archipelago/launcher/internal/generate"
	"github.com/m-this/tf2-archipelago/launcher/internal/runshape"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/winproc"
)

// botSeats is how many seats the team composition names, which is the largest
// team the mod will field.
const botSeats = 6

// settingsForm is the settings screen: the tabs, their rows, and the copy of
// the settings the rows write into.
type settingsForm struct {
	edited settings.Settings

	tabs    []settingsTab
	tab     int
	focused int
	offset  int

	room   string // the room address as typed, parsed on save
	warn   string
	saved  func(settings.Settings) tea.Cmd
	closed bool
}

type settingsTab struct {
	title  string
	fields []field
}

func newSettingsForm(s settings.Settings, saved func(settings.Settings) tea.Cmd) *settingsForm {
	form := &settingsForm{edited: s, saved: saved}
	form.room = settings.Room{Host: s.APHost, Port: s.APPort, TLS: s.APTls}.String()
	form.tabs = []settingsTab{
		{title: "Player options", fields: form.playerFields()},
		{title: "Missions", fields: form.missionFields()},
		{title: "Archipelago room", fields: form.roomFields()},
		{title: "Game server", fields: form.serverFields()},
		{title: "Bots", fields: form.botFields()},
		{title: "Who can join (beta)", fields: form.reachFields()},
	}
	return form
}

func (f *settingsForm) playerFields() []field {
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

	return []field{
		&choiceField{
			label:   "Easiest tier",
			help:    "The easiest tier a mission may come from. Harder tiers are always in as well, so the pool shrinks as this rises.",
			options: tierLabels,
			index:   max(slices.IndexFunc(tiers, func(t runshape.Tier) bool { return t.Key == f.edited.MvmDifficulty }), 0),
			apply:   func(i int) { f.edited.MvmDifficulty = tiers[i].Key },
		},
		&numberField{
			label: "Missions used",
			help:  "How many missions this run uses, out of the pool above. Eight is about fifty waves.",
			value: &f.edited.MvmMissionCount, low: 1, high: 29,
		},
		&choiceField{
			label:   "Goal",
			help:    "What ends the run.",
			options: goalLabels,
			index:   max(slices.IndexFunc(goals, func(g runshape.Goal) bool { return g.Key == f.edited.MvmGoal }), 0),
			apply:   func(i int) { f.edited.MvmGoal = goals[i].Key },
		},
		&numberField{
			label: "Missionsanity share",
			help:  "How much of the run's checks come from waves rather than whole missions, as a percentage.",
			value: &f.edited.MvmMissionsanityPct, low: 0, high: 100,
		},
		&toggleField{
			label: "Death Link",
			help:  "A lost wave kills every other player in the multiworld who has Death Link on, and their deaths wipe your team.",
			value: &f.edited.MvmDeathLink, on: "share deaths", off: "share deaths",
		},
		&textField{
			label:       "Archipelago app",
			help:        "Where the Archipelago app is installed. Blank means the launcher looks where the installer puts it.",
			value:       &f.edited.ArchipelagoDir,
			placeholder: defaultAppDir(),
		},
		&actionField{
			label: "Generate seed",
			help:  "Make the seed with the Archipelago app on this machine, then upload the archive at archipelago.gg/uploads to open a room.",
			hint:  "enter",
			run:   f.generateSeed,
		},
		&actionField{
			label: "Open tf2.yaml",
			help:  "Write the player file and show it. It is what the seed is generated from.",
			hint:  "enter",
			run:   f.openPlayerFile,
		},
	}
}

func (f *settingsForm) missionFields() []field {
	choices := runshape.StartMissionChoices()
	choiceLabels := make([]string, 0, len(choices))
	for _, choice := range choices {
		choiceLabels = append(choiceLabels, choice.Label)
	}
	classes := runshape.StartClassChoices()

	fields := []field{
		&choiceField{
			label:   "Start mission",
			help:    "Where the run begins. The seed starts there and the server boots there.",
			options: choiceLabels,
			index:   max(slices.IndexFunc(choices, func(c runshape.MissionChoice) bool { return c.PopFile == f.edited.MvmStartMission }), 0),
			apply: func(i int) {
				f.edited.MvmStartMission = choices[i].PopFile
				if choices[i].PopFile != "" {
					f.edited.SrcdsStartMission = choices[i].PopFile
				}
			},
		},
		&choiceField{
			label:   "Start class",
			help:    "The mercenary the run starts with. The tier of the start mission decides how many it starts with.",
			options: classes,
			index:   max(slices.Index(classes, f.edited.MvmStartClass), 0),
			apply:   func(i int) { f.edited.MvmStartClass = startClass(classes, i) },
		},
	}

	// One row per mission, because the pool is what the seed draws from and
	// the window gives it a table with a tick in every row.
	for _, mission := range gamedata.Missions {
		fields = append(fields, f.poolField(mission))
	}
	return fields
}

// poolField is one mission's place in the pool. The setting is the missions
// left out, so the row reads the other way round from what it writes.
func (f *settingsForm) poolField(mission gamedata.Mission) field {
	played, _ := gamedata.MapByID(mission.Map)
	inPool := !slices.Contains(f.edited.MvmExcludedMissions, mission.PopFile)
	held := inPool

	return &poolToggle{
		toggleField: toggleField{
			label: fmt.Sprintf("%s (%s)", mission.Name, played.Name),
			help:  fmt.Sprintf("%s, %d waves. Off means the seed never draws it.", mission.Difficulty.String(), mission.Waves),
			value: &held, on: "in the pool", off: "left out",
		},
		popFile: mission.PopFile,
		form:    f,
		held:    &held,
	}
}

// poolToggle keeps the excluded list in step with the tick.
type poolToggle struct {
	toggleField
	popFile string
	form    *settingsForm
	held    *bool
}

func (p *poolToggle) Handle(msg tea.KeyMsg) bool {
	if !p.toggleField.Handle(msg) {
		return false
	}
	excluded := p.form.edited.MvmExcludedMissions
	excluded = slices.DeleteFunc(excluded, func(popFile string) bool { return popFile == p.popFile })
	if !*p.held {
		excluded = append(excluded, p.popFile)
	}
	p.form.edited.MvmExcludedMissions = excluded
	return true
}

func (f *settingsForm) roomFields() []field {
	return []field{
		&toggleField{
			label: "Test mode",
			help:  "Play without Archipelago at all: the launcher serves a multiworld of one and simulates the other players.",
			value: &f.edited.TestMode, on: "no room needed", off: "use a real room",
		},
		&textField{
			label:       "Room address",
			help:        "The line from your room page on archipelago.gg: host and port.",
			value:       &f.room,
			placeholder: "archipelago.gg:12345",
		},
		&textField{
			label:       "Room password",
			help:        "Only if the room asks for one.",
			value:       &f.edited.APPassword,
			placeholder: "none",
			hidden:      true,
		},
		&textField{
			label:       "Slot name",
			help:        "The name this server plays under in the multiworld. It has to match the name in tf2.yaml.",
			value:       &f.edited.APSlotName,
			placeholder: "tf2",
		},
	}
}

func (f *settingsForm) serverFields() []field {
	return []field{
		&textField{
			label:       "Server name",
			help:        "What the server calls itself in the player list.",
			value:       &f.edited.SrcdsHostname,
			placeholder: "Mann vs Archipelago",
		},
		&textField{
			label:       "Server password",
			help:        "What your friends type before connect. Blank means anybody with the address can join.",
			value:       &f.edited.SrcdsPw,
			placeholder: "none",
			hidden:      true,
		},
		&numberField{
			label: "Game port",
			help:  "UDP and TCP, 27015 by default. Who can reach it is on the last tab.",
			value: &f.edited.SrcdsPort, low: 1024, high: 65535,
		},
		&textField{
			label:       "Admins by Steam id",
			help:        "Who may run the admin commands, separated by commas. The 17 digit id or SourceMod's STEAM_0:1:26975537.",
			value:       &f.edited.SrcdsAdminSteamIDs,
			placeholder: "none",
		},
		&actionField{
			label: "Debug logs",
			help:  "Put the logs, the settings without their passwords and the player file in one zip, for sending to whoever is helping you.",
			hint:  "enter",
			run:   f.debugBundle,
		},
	}
}

func (f *settingsForm) botFields() []field {
	fields := []field{
		&toggleField{
			label: "Fill RED with bots",
			help:  "Valve balances every wave for six players. Off leaves the seats empty until an admin runs sm_addbots.",
			value: &f.edited.SrcdsBots, on: "bots on", off: "bots off",
		},
		&numberField{
			label: "Fill RED to",
			help:  "How many players the server fills RED to, humans included. Lower is harder.",
			value: &f.edited.SrcdsBotTeamSize, low: 1, high: botSeats,
		},
		&toggleField{
			label: "Say what they buy",
			help:  "Write every bot purchase at the upgrade station to the chat. It is a lot of chat.",
			value: &f.edited.BotUpgradesChat, on: "in the chat", off: "quiet",
		},
	}

	// The seats, in the order they fill. This is the one thing a blacklist
	// cannot say, and the reason the team composition exists.
	for seat := range botSeats {
		fields = append(fields, f.seatField(seat))
	}

	for _, class := range botloadout.Classes {
		fields = append(fields, f.classField(class), f.loadoutField(class))
	}
	return fields
}

func (f *settingsForm) seatField(seat int) field {
	options := []string{"the mod decides"}
	for _, class := range botloadout.Classes {
		options = append(options, class.Name)
	}
	index := 0
	if seat < len(f.edited.SrcdsBotTeamComp) {
		if at := slices.IndexFunc(botloadout.Classes, func(c botloadout.Class) bool {
			return c.Key == f.edited.SrcdsBotTeamComp[seat]
		}); at >= 0 {
			index = at + 1
		}
	}

	return &choiceField{
		label:   fmt.Sprintf("Seat %d", seat+1),
		help:    "The classes the bots fill RED with, in this order. The first seats are the ones that always get filled.",
		options: options,
		index:   index,
		apply:   func(i int) { f.setSeat(seat, i) },
	}
}

// setSeat rewrites the composition from the seats that name a class. A seat
// left to the mod contributes nothing, so the list is the picked seats in seat
// order, which is what the convar reads.
func (f *settingsForm) setSeat(seat, index int) {
	comp := make([]string, botSeats)
	copy(comp, f.edited.SrcdsBotTeamComp)
	if index == 0 {
		comp[seat] = ""
	} else {
		comp[seat] = botloadout.Classes[index-1].Key
	}
	f.edited.SrcdsBotTeamComp = slices.DeleteFunc(comp, func(key string) bool { return key == "" })
}

func (f *settingsForm) classField(class botloadout.Class) field {
	allowed := !slices.Contains(f.edited.SrcdsBotClassBlacklist, class.Key)
	held := allowed

	return &classToggle{
		toggleField: toggleField{
			label: class.Name,
			help:  "Off means the bots never play it. A class named in a seat above beats this.",
			value: &held, on: "they play it", off: "never",
		},
		key:  class.Key,
		form: f,
		held: &held,
	}
}

type classToggle struct {
	toggleField
	key  string
	form *settingsForm
	held *bool
}

func (c *classToggle) Handle(msg tea.KeyMsg) bool {
	if !c.toggleField.Handle(msg) {
		return false
	}
	list := c.form.edited.SrcdsBotClassBlacklist
	list = slices.DeleteFunc(list, func(key string) bool { return key == c.key })
	if !*c.held {
		list = append(list, c.key)
	}
	c.form.edited.SrcdsBotClassBlacklist = list
	return true
}

func (f *settingsForm) loadoutField(class botloadout.Class) field {
	options := make([]string, 0, len(class.Loadouts))
	for _, loadout := range class.Loadouts {
		options = append(options, loadout.Label())
	}
	current := f.edited.SrcdsBotLoadouts[class.Key]
	index := max(slices.IndexFunc(class.Loadouts, func(l botloadout.Loadout) bool { return l.Key == current }), 0)

	return &choiceField{
		label:   "  " + class.Name + " loadout",
		help:    "What a bot of this class spawns with. Stock is the game's own.",
		options: options,
		index:   index,
		apply: func(i int) {
			if f.edited.SrcdsBotLoadouts == nil {
				f.edited.SrcdsBotLoadouts = map[string]string{}
			}
			pick := class.Loadouts[i]
			if pick.Key == botloadout.StockKey {
				delete(f.edited.SrcdsBotLoadouts, class.Key)
				return
			}
			f.edited.SrcdsBotLoadouts[class.Key] = pick.Key
		},
	}
}

func (f *settingsForm) reachFields() []field {
	reaches := settings.Reaches()
	labels := make([]string, 0, len(reaches))
	for _, reach := range reaches {
		labels = append(labels, reach.Label())
	}

	return []field{
		&choiceField{
			label:   "Who can reach it",
			help:    "Where the server takes connections from. Without a login token it stays on the local network whatever this says.",
			options: labels,
			index:   max(slices.Index(reaches, f.edited.SrcdsReach), 0),
			apply:   func(i int) { f.edited.SrcdsReach = reaches[i] },
		},
		&textField{
			label:       "Login token",
			help:        "A Game Server Login Token for app id 440, from steamcommunity.com/dev/managegameservers.",
			value:       &f.edited.SrcdsToken,
			placeholder: "0",
		},
	}
}

// save parses what cannot be typed wrong twice, and hands the settings back.
func (f *settingsForm) save() tea.Cmd {
	room, err := settings.ParseRoom(f.room)
	if err != nil && !f.edited.TestMode {
		f.warn = err.Error()
		return nil
	}
	f.edited.APHost, f.edited.APPort, f.edited.APTls = room.Host, room.Port, room.TLS

	if f.edited.SrcdsReach.NeedsToken() && !settings.HasToken(f.edited.SrcdsToken) {
		f.warn = "that reach needs a login token, or the server stays on the local network"
	}
	f.closed = true
	return f.saved(f.edited)
}

func (f *settingsForm) generateSeed() tea.Cmd {
	return func() tea.Msg {
		if _, err := generate.FindApp(f.edited.ArchipelagoDir); err != nil {
			return noticeMsg("the Archipelago app was not found in " +
				strings.Join(generate.SearchPath(f.edited.ArchipelagoDir), ", "))
		}
		result, err := generate.Run(context.Background(), generate.Options{
			Settings:           f.edited,
			AppDir:             f.edited.ArchipelagoDir,
			Apworld:            assets.Apworld(),
			ArchipelagoVersion: assets.ArchipelagoVersion,
		})
		if err != nil {
			return noticeMsg("generate: " + err.Error())
		}
		_ = winproc.Open(filepath.Dir(result.Archive))
		return noticeMsg("wrote " + result.Archive + ": upload it at archipelago.gg/uploads")
	}
}

func (f *settingsForm) openPlayerFile() tea.Cmd {
	return func() tea.Msg {
		path, err := settings.WritePlayerFile(f.edited, assets.ArchipelagoVersion)
		if err != nil {
			return noticeMsg(err.Error())
		}
		_ = winproc.Open(path)
		return noticeMsg("wrote " + path)
	}
}

func (f *settingsForm) debugBundle() tea.Cmd {
	return func() tea.Msg {
		path, err := debugbundle.Write(f.edited, assets.Versions(), time.Now())
		if err != nil {
			return noticeMsg(err.Error())
		}
		return noticeMsg("wrote " + path + ", with no passwords in it")
	}
}

// noticeMsg is a line for the log: what an action did, or why it did not.
type noticeMsg string

func defaultAppDir() string {
	if dirs := generate.SearchPath(""); len(dirs) > 0 {
		return dirs[0]
	}
	return ""
}

// startClass is the class name the run starts with, where the first choice is
// the seed's own pick rather than a class.
func startClass(classes []string, index int) string {
	if index <= 0 || index >= len(classes) {
		return ""
	}
	return classes[index]
}
