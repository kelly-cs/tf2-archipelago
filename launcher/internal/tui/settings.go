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
	"github.com/m-this/tf2-archipelago/launcher/internal/roomcheck"
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

	warn string

	/*
		problem is what Save could not do, shown over the rows until it is
		dismissed.

		warn under the footer is for something the player can see and fix on the
		row it is about: a number out of range, a room address half typed. A
		save that could not write the file is neither. It is the one thing the
		screen was opened to do, it has just failed, and the reason is a path
		and an operating system error that will scroll past unread at the bottom
		of a screen. So it stops the screen instead.
	*/
	problem string

	// persist writes the settings and returns what it actually wrote, which is
	// not quite what it was given: an empty RCON password is filled in. Separate
	// from saved, because the screen has to still be open when it fails.
	persist func(settings.Settings) (settings.Settings, error)
	saved   func(settings.Settings) tea.Cmd
	repair  func() ([]string, error)
	reset   func() (settings.Settings, error)
	closed  bool
}

type settingsTab struct {
	title  string
	fields []*modelRow
}

type settingsDeps struct {
	// persist writes the settings to disk and returns what it wrote, or what
	// stopped it. The screen stays open on an error, so nothing typed is lost.
	persist func(settings.Settings) (settings.Settings, error)

	// saved is everything after a successful write: the supervisor, the player
	// file, and the restart if one is needed.
	saved  func(settings.Settings) tea.Cmd
	repair func() ([]string, error)
	reset  func() (settings.Settings, error)
}

func newSettingsForm(s settings.Settings, deps settingsDeps) *settingsForm {
	f := &settingsForm{
		state:   form.NewState(s),
		persist: deps.persist,
		saved:   deps.saved,
		repair:  deps.repair,
		reset:   deps.reset,
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
	"run.generate", "run.open_player_file", "run.open_folder", "run.open_settings_file",
	"missions.download_packs", "missions.import_assets", "missions.check_selection",
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
	case "run.open_settings_file":
		return f.openSettingsFile
	case "missions.download_packs":
		return f.downloadSelectedCommunityAssets
	case "missions.import_assets":
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
	save parses what a row could not refuse while it was being typed, then writes.

The room address is the one that could not be refused. It is deferred, so the row
keeps what was typed and this is where an address that never became one is
refused. Test mode never dials a real room, so it does not need one.

The write happens here rather than after the screen closes, and that is the
point of it. It used to close first and hand the settings to the caller, so a
file that could not be written became a line at the bottom of the screen behind
a settings page that was no longer there, and every answer the player had typed
was gone with it. A player on Discord read that as the Save button doing nothing
at all, went looking for a config.json in the install root, and found none:
there is none there to find, it goes under the OS's config directory.

So a failed write keeps the screen open, keeps the answers, and says what it
tried to write and why it could not.
*/
func (f *settingsForm) save() tea.Cmd {
	/* A room that will not parse no longer refuses the save. It used to, and
	   that is how a player lost a login token they were setting two pages
	   away: one field they could not see blocked every other answer on the
	   screen. An address that is not one is left out and said afterwards, along
	   with the room that was configured and did not answer, because the two
	   leave the player in the same place. */
	room, roomErr := settings.ParseRoom(f.state.Draft.Room)
	if roomErr == nil {
		f.state.Settings.APHost, f.state.Settings.APPort, f.state.Settings.APTls = room.Host, room.Port, room.TLS
	} else if strings.TrimSpace(f.state.Draft.Room) == "" {
		f.state.Settings.APHost, f.state.Settings.APPort = "", 0
	}

	if _, err := settings.CheckRunSelection(f.state.Settings); err != nil {
		return f.refuse("Missions", err.Error())
	}

	written, err := f.persist(f.state.Settings)
	if err != nil {
		return f.refuse("", err.Error())
	}
	f.state.Settings = written

	/* A reach with no token is a server every client is refused from, and
	   nothing on screen would say why. It is not a reason to refuse the save,
	   though: the settings are good and the server simply stays on the local
	   network, so it is said and the screen closes. */
	if f.state.Settings.SrcdsReach.NeedsToken() && !settings.HasToken(f.state.Settings.SrcdsToken) {
		f.warn = "that reach needs a login token, or the server stays on the local network"
	}
	f.closed = true
	return tea.Batch(f.saved(written), f.checkRoom(written, roomErr))
}

/*
	checkRoom asks the room whether it is there, once the settings are safely on
	disk.

Saving is when a player finds out their address works, and it used to be much
later: Start refused with "AP_PORT is not set", or the bridge retried a dead
room in a log nobody was reading.

It runs after the write and cannot undo it. A room that does not answer is a
notice, not a refusal, because the settings are worth keeping either way and a
player who has not made the room yet has done nothing wrong.
*/
func (f *settingsForm) checkRoom(s settings.Settings, roomErr error) tea.Cmd {
	return func() tea.Msg {
		if roomErr != nil && strings.TrimSpace(f.state.Draft.Room) != "" {
			return noticeMsg("the room address was not saved: " + roomErr.Error() + ". " +
				roomcheck.NotConfigured.Advice())
		}
		result, err := roomcheck.Check(context.Background(), s)
		switch result {
		case roomcheck.Live:
			return noticeMsg("settings saved, and the Archipelago room answered")
		case roomcheck.Skipped:
			return noticeMsg("settings saved. Test mode: no room is needed")
		case roomcheck.NotConfigured:
			return noticeMsg("settings saved. " + result.Advice())
		case roomcheck.Unreachable:
			return noticeMsg("settings saved, but the room did not answer: " +
				err.Error() + ". " + result.Advice())
		}
		return noticeMsg("settings saved")
	}
}

/*
	refuse stops the save, opens the page at fault, and says why.

Every way a Save can fail comes through here, and that is the change. They used
to be three different things: a room that would not parse set a line under the
footer, a pool too small for the seed set the same line, and a file that would
not write closed the screen and logged somewhere else entirely.

The page matters as much as the reason, which a debug bundle from a player made
plain. They were setting a login token, pressed Save, and nothing happened. The
Save was refusing on the room address, two pages away, with a message beside a
field they were not looking at. Their log has no line about saving at all,
because a refused save wrote none. So this opens the page holding the problem
before it says what the problem is, and page may be empty for a failure that
belongs to no page, like a file that would not write.
*/
func (f *settingsForm) refuse(page, reason string) tea.Cmd {
	f.showTab(page)
	f.problem = reason
	f.warn = ""
	return nil
}

// dismiss takes the problem box off, and says whether there was one. Any key
// does it: the box has one thing to say and nothing to choose.
func (f *settingsForm) dismiss() bool {
	if f.problem == "" {
		return false
	}
	f.problem = ""
	return true
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

/*
	openSettingsFile shows where the launcher keeps its own settings.

Not the install folder, which is the mistake it exists to correct. A player went
looking for config.json in the install folder, found none, and read that as
nothing having saved at all. The file is under the OS's config directory and
nothing in the launcher would show it.

The folder is opened rather than the file: config.json has no application to
open it with on a fresh Windows install, and a file browser sitting on it is
what somebody asking "where is the config file" actually wants.
*/
func (f *settingsForm) openSettingsFile() tea.Cmd {
	return func() tea.Msg {
		path, err := settings.Path()
		if err != nil {
			return noticeMsg("cannot work out where the settings live: " + err.Error())
		}
		if err := winproc.Open(filepath.Dir(path)); err != nil {
			return noticeMsg("the settings are at " + path + ", and it cannot be opened: " + err.Error())
		}
		return noticeMsg("the settings are at " + path)
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
