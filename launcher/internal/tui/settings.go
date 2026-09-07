/*
The settings screen: the pages internal/form declares, drawn for a keyboard.

Nothing here says what a setting is called or what it is for. form does, once,
and the window reads the same list. What is here is what a terminal does about
a row: left and right change a choice, digits go into a number, a space toggles,
and Enter presses. Those are the parts that should differ between a window and a
terminal, and now they are the only parts that do.

Nothing is saved until Save: cancelling leaves the file alone, the way closing
a dialog does.
*/
package tui

import (
	"context"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/assets"
	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/debugbundle"
	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	"github.com/m-this/tf2-archipelago/launcher/internal/generate"
	"github.com/m-this/tf2-archipelago/launcher/internal/installer"
	"github.com/m-this/tf2-archipelago/launcher/internal/runshape"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/tailscalefastdl"
	"github.com/m-this/tf2-archipelago/launcher/internal/winproc"
)

// settingsForm is the settings screen: the state being edited, the pages form
// resolved from it, and where the cursor is.
type settingsForm struct {
	state form.State

	// communityAvailable is derived from valid local ZIPs, never merely from
	// checkbox state. Community mission rows stay unavailable until their
	// assets can actually be used.
	communityAvailable []string

	tabs    []settingsTab
	tab     int
	focused int
	offset  int

	warn   string
	saved  func(settings.Settings) tea.Cmd
	repair func() ([]string, error)
	reset  func() (settings.Settings, error)
	closed bool
}

type settingsTab struct {
	title  string
	fields []*modelRow
}

type settingsDeps struct {
	saved  func(settings.Settings) tea.Cmd
	repair func() ([]string, error)
	reset  func() (settings.Settings, error)
}

func newSettingsForm(s settings.Settings, deps settingsDeps) *settingsForm {
	f := &settingsForm{
		state:  form.NewState(s),
		saved:  deps.saved,
		repair: deps.repair,
		reset:  deps.reset,
	}
	f.communityAvailable = availableCommunityPackNames(s.CommunityContentDir)
	f.build()
	return f
}

/*
	build lays the pages out from the state as it is now

Every page, every time. It used to patch what was on screen for most changes and
rebuild for a few, and the few were the ones somebody noticed: resetting the
settings, and the pool's All and None. The list is longer than that. Ticking a
community pack adds a row per mission of that pack, choosing a seat's class
changes which loadouts that seat may carry, and changing the drafted loadout's
class changes which slots it even has.

So a change rebuilds. It is a few hundred rows of plain structs and it happens
on a keystroke a human made, which is not a budget worth managing.
*/
func (f *settingsForm) build() {
	model := form.Build(f.state, f.env())
	tabs := make([]settingsTab, 0, len(model.Tabs))
	for _, page := range model.Tabs {
		rows := make([]*modelRow, 0, len(page.Fields))
		for _, resolved := range page.Fields {
			rows = append(rows, &modelRow{f: resolved, apply: f.applyChange, run: f.action(resolved.ID)})
		}
		tabs = append(tabs, settingsTab{title: page.Title, fields: rows})
	}
	f.tabs = tabs
	f.tab = min(f.tab, max(len(f.tabs)-1, 0))
	f.focused = min(f.focused, max(len(f.fields())-1, 0))
}

// env is what the specs need that the state does not hold.
func (f *settingsForm) env() form.Env {
	return form.Env{
		CommunityAvailable: f.communityAvailable,
		AppDirDefault:      defaultAppDir(),
	}
}

/*
	applyChange is the one place the terminal writes a setting

Every row goes through it, so what a row may hold is form's answer and never one
the terminal keeps for itself. A refusal is shown rather than swallowed: the
bounds are on the row, so a number outside them is a typo the player can see.

The rebuild after is not an optimisation anybody skipped: see build.
*/
func (f *settingsForm) applyChange(c form.Change) error {
	next, err := form.Apply(f.state, f.env(), c)
	if err != nil {
		f.warn = err.Error()
		return err
	}
	f.state = next
	f.warn = ""
	f.build()
	return nil
}

// fields is the rows of the open page.
func (f *settingsForm) fields() []*modelRow {
	if f.tab >= len(f.tabs) {
		return nil
	}
	return f.tabs[f.tab].fields
}

/*
	action is what pressing a row does, by the ID form declared it under

form says what a button is called and what it is for; the work is the terminal's
because it ends in a tea.Cmd. A row form declares as an Action with no entry
here would draw a button that swallows the Enter and does nothing, which nobody
reports as a bug: they report the feature as missing. TestEveryButtonDoesSomething
refuses one, and wiredActions below is the same list for the other direction.
*/
var wiredActions = []string{
	"run.generate", "run.open_player_file", "run.open_folder",
	"missions.download_packs", "missions.use_local_packs", "missions.check_selection",
	"missions.pool_all", "missions.pool_none",
	"server.debug_bundle", "server.repair", "server.reset",
	"net.check_funnel",
	"bots.save_team", "bots.remove_team",
	"loadout.save",
}

func (f *settingsForm) action(id string) func() tea.Cmd {
	switch id {
	case "run.generate":
		return f.generateSeed
	case "run.open_player_file":
		return f.openPlayerFile
	case "run.open_folder":
		return f.openInstallRoot
	case "missions.download_packs":
		return f.downloadSelectedCommunityAssets
	case "missions.use_local_packs":
		return f.useLocalCommunityAssets
	case "missions.check_selection":
		return f.checkRunSelection
	case "missions.pool_all":
		return func() tea.Cmd { return f.setPool(true) }
	case "missions.pool_none":
		return func() tea.Cmd { return f.setPool(false) }
	case "server.debug_bundle":
		return f.debugBundle
	case "server.repair":
		return f.runRepair
	case "server.reset":
		return f.runReset
	case "net.check_funnel":
		return f.checkTailscaleFunnel
	case "bots.save_team":
		return f.saveTeam
	case "bots.remove_team":
		return f.removeTeam
	case "loadout.save":
		return f.saveLoadout
	}
	return nil
}

// --- The pool ---

func (f *settingsForm) setPool(inPool bool) tea.Cmd {
	excluded := make([]string, 0)
	if !inPool {
		for _, mission := range gamedata.PlayableMissions() {
			excluded = append(excluded, mission.PopFile)
		}
	} else {
		// A community mission whose pack is not on disk stays out even of All:
		// the seed would draw a mission the server cannot load.
		visible := runshape.VisibleMissions(f.communityAvailable)
		for _, mission := range gamedata.PlayableMissions() {
			if gamedata.MissionPack(mission.ID) != "" && !slices.ContainsFunc(visible, func(candidate gamedata.Mission) bool {
				return candidate.ID == mission.ID
			}) {
				excluded = append(excluded, mission.PopFile)
			}
		}
	}
	f.state.Settings.MvmExcludedMissions = excluded
	f.build()

	if inPool {
		return func() tea.Msg { return noticeMsg("every mission is in the pool") }
	}
	return func() tea.Msg { return noticeMsg("every mission is left out: tick the ones this run may draw") }
}

// --- The bot team and the loadouts ---

func (f *settingsForm) saveTeam() tea.Cmd {
	name := strings.TrimSpace(f.state.Draft.TeamName)
	if name == "" {
		return func() tea.Msg { return noticeMsg("name the team first") }
	}
	presets := maps.Clone(f.state.Settings.SrcdsBotTeamPresets)
	if presets == nil {
		presets = map[string]settings.BotTeam{}
	}
	presets[name] = settings.BotTeamOf(f.state.Settings)
	f.state.Settings.SrcdsBotTeamPresets = presets
	f.state.Draft.TeamName = ""
	f.build()
	return func() tea.Msg { return noticeMsg("saved the team as " + name) }
}

func (f *settingsForm) removeTeam() tea.Cmd {
	name := strings.TrimSpace(f.state.Draft.TeamName)
	if name == "" {
		return func() tea.Msg { return noticeMsg("name the team to remove first") }
	}
	if _, found := f.state.Settings.SrcdsBotTeamPresets[name]; !found {
		return func() tea.Msg { return noticeMsg("no team saved as " + name) }
	}
	presets := make(map[string]settings.BotTeam, len(f.state.Settings.SrcdsBotTeamPresets)-1)
	for existing, team := range f.state.Settings.SrcdsBotTeamPresets {
		if existing != name {
			presets[existing] = team
		}
	}
	if len(presets) == 0 {
		presets = nil
	}
	f.state.Settings.SrcdsBotTeamPresets = presets
	f.state.Draft.TeamName = ""
	f.build()
	return func() tea.Msg { return noticeMsg("removed the team " + name) }
}

func (f *settingsForm) saveLoadout() tea.Cmd {
	name := strings.TrimSpace(f.state.Draft.LoadoutName)
	if name == "" {
		return func() tea.Msg { return noticeMsg("name the loadout first") }
	}
	built := maps.Clone(f.state.Settings.SrcdsBotCustomLoadouts)
	if built == nil {
		built = map[string]botloadout.Built{}
	}
	built[name] = f.state.Draft.Loadout
	f.state.Settings.SrcdsBotCustomLoadouts = built
	f.build()
	return func() tea.Msg { return noticeMsg("saved the loadout as " + name) }
}

// --- Community assets ---

func (f *settingsForm) checkRunSelection() tea.Cmd {
	return func() tea.Msg {
		result, err := settings.CheckRunSelection(f.state.Settings)
		if err != nil {
			return noticeMsg(err.Error())
		}
		return noticeMsg(result.Summary())
	}
}

func (f *settingsForm) downloadSelectedCommunityAssets() tea.Cmd {
	folder := strings.TrimSpace(f.state.Settings.CommunityContentDir)
	selected := f.state.Settings
	selected.CommunityContentDir = folder
	archives := settings.CommunityArchives(selected)
	return func() tea.Msg {
		if folder == "" {
			return noticeMsg("choose an asset pack folder first")
		}
		if len(archives) == 0 {
			return noticeMsg("select at least one community pack first")
		}
		if err := installer.DownloadCommunityArchives(context.Background(), archives, func(string, ...any) {}); err != nil {
			return noticeMsg("community assets: " + err.Error())
		}
		return communityAssetsMsg{
			notice:    "Selected community packs are ready in " + folder,
			available: availableCommunityPackNames(folder),
		}
	}
}

func (f *settingsForm) useLocalCommunityAssets() tea.Cmd {
	folder := strings.TrimSpace(f.state.Settings.CommunityContentDir)
	return func() tea.Msg {
		if folder == "" {
			return noticeMsg("choose an asset pack folder first")
		}
		available := availableCommunityPackNames(folder)
		if len(available) == 0 {
			return noticeMsg("no valid archive-assets.zip or mlarchive-assets.zip was found in " + folder)
		}
		return communityAssetsMsg{
			notice:      "Using local community packs from " + folder,
			available:   available,
			selectPacks: true,
		}
	}
}

func availableCommunityPackNames(folder string) []string {
	paths := installer.AvailableCommunityArchives(settings.KnownCommunityArchives(strings.TrimSpace(folder)))
	packs := make([]string, 0, len(paths))
	for _, path := range paths {
		packs = append(packs, filepath.Base(path))
	}
	return packs
}

// communityAssetsMsg is what a download or a local scan found: which packs are
// usable now, and whether to tick them.
type communityAssetsMsg struct {
	notice      string
	available   []string
	selectPacks bool
}

func (f *settingsForm) applyCommunityAssets(msg communityAssetsMsg) {
	f.communityAvailable = slices.Clone(msg.available)
	if msg.selectPacks {
		for _, pack := range []string{settings.CommunityPackPotato, settings.CommunityPackMoonlight} {
			f.state.Settings.CommunityPacks = slices.DeleteFunc(f.state.Settings.CommunityPacks,
				func(name string) bool { return name == pack })
			selected := slices.Contains(msg.available, pack)
			if selected {
				f.state.Settings.CommunityPacks = append(f.state.Settings.CommunityPacks, pack)
			}
			for _, mission := range gamedata.PlayableMissions() {
				if gamedata.MissionPack(mission.ID) != pack {
					continue
				}
				f.state.Settings.MvmExcludedMissions = slices.DeleteFunc(f.state.Settings.MvmExcludedMissions,
					func(popFile string) bool { return popFile == mission.PopFile })
				if !selected {
					f.state.Settings.MvmExcludedMissions = append(f.state.Settings.MvmExcludedMissions, mission.PopFile)
				}
			}
		}
	}
	f.build()
}

func (f *settingsForm) checkTailscaleFunnel() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		result, err := tailscalefastdl.Authorize(ctx)
		if err != nil {
			return noticeMsg("Tailscale Funnel: " + err.Error())
		}
		if result.ApprovalURL != "" {
			if err := winproc.OpenURL(result.ApprovalURL); err != nil {
				return noticeMsg("approve Funnel at " + result.ApprovalURL)
			}
			return noticeMsg("approve Funnel in the browser, then run Set up / check Funnel again")
		}
		return noticeMsg("Tailscale Funnel is ready for this tailnet")
	}
}

// --- Save, and the rest of the buttons ---

/*
	save parses what a row could not refuse while it was being typed

The room address is the one. It is deferred, so the row keeps what was typed and
this is where an address that never became one is refused. Test mode never dials
a real room, so it does not need one.
*/
func (f *settingsForm) save() tea.Cmd {
	room, err := settings.ParseRoom(f.state.Draft.Room)
	if err != nil && !f.state.Settings.TestMode {
		f.warn = err.Error()
		return nil
	}
	f.state.Settings.APHost, f.state.Settings.APPort, f.state.Settings.APTls = room.Host, room.Port, room.TLS

	if f.state.Settings.SrcdsReach.NeedsToken() && !settings.HasToken(f.state.Settings.SrcdsToken) {
		f.warn = "that reach needs a login token, or the server stays on the local network"
	}
	if _, err := settings.CheckRunSelection(f.state.Settings); err != nil {
		f.warn = err.Error()
		return nil
	}
	f.closed = true
	return f.saved(f.state.Settings)
}

func (f *settingsForm) generateSeed() tea.Cmd {
	return func() tea.Msg {
		if _, err := settings.CheckRunSelection(f.state.Settings); err != nil {
			return noticeMsg(err.Error())
		}
		if _, err := generate.FindApp(f.state.Settings.ArchipelagoDir); err != nil {
			return noticeMsg("the Archipelago app was not found in " +
				strings.Join(generate.SearchPath(f.state.Settings.ArchipelagoDir), ", "))
		}
		result, err := generate.Run(context.Background(), generate.Options{
			Settings:           f.state.Settings,
			AppDir:             f.state.Settings.ArchipelagoDir,
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
		path, err := settings.WritePlayerFile(f.state.Settings, assets.ArchipelagoVersion)
		if err != nil {
			return noticeMsg(err.Error())
		}
		_ = winproc.Open(path)
		return noticeMsg("wrote " + path)
	}
}

func (f *settingsForm) openInstallRoot() tea.Cmd {
	return func() tea.Msg {
		root := f.state.Settings.InstallRoot
		if err := winproc.Open(root); err != nil {
			return noticeMsg("cannot open " + root + ": " + err.Error())
		}
		return noticeMsg("opened " + root)
	}
}

// runRepair stops everything the launcher started and removes what the next
// start can fetch again. It blocks the screen for as long as that takes, which
// is the same wait the window's message box covers.
func (f *settingsForm) runRepair() tea.Cmd {
	return func() tea.Msg {
		removed, err := f.repair()
		switch {
		case err != nil:
			return noticeMsg("repair: " + err.Error())
		case len(removed) == 0:
			return noticeMsg("repair: nothing to remove")
		default:
			return noticeMsg("repair removed " + strings.Join(removed, ", ") + ". Press s when you are ready.")
		}
	}
}

// runReset takes the defaults back into the form as well as onto disk. The
// window closes its dialog instead, because every control on screen still held
// the old answer and Save would have written them straight back.
func (f *settingsForm) runReset() tea.Cmd {
	fresh, err := f.reset()
	if err != nil {
		return func() tea.Msg { return noticeMsg("reset: " + err.Error()) }
	}
	f.state = form.NewState(fresh)
	f.warn = ""
	f.build()
	return func() tea.Msg { return noticeMsg("every setting is back to its default") }
}

func (f *settingsForm) debugBundle() tea.Cmd {
	return func() tea.Msg {
		path, err := debugbundle.Write(f.state.Settings, assets.Versions(), time.Now())
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

// showTab opens on the page with that title, and stays where it is for a title
// no page carries.
func (f *settingsForm) showTab(title string) {
	if title == "" {
		return
	}
	for i, tab := range f.tabs {
		if tab.title == title {
			f.tab = i
			return
		}
	}
}
