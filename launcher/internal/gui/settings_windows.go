//go:build windows

/*
The settings window: the pages internal/form declares, drawn with walk.

Nothing here says what a setting is called or what it is for. form does, once,
and the terminal reads the same list. What is here is what a window does about a
row: a ComboBox for a choice, a NumberEdit for a number, a CheckBox for a
toggle, a Browse button beside a folder.

The one place the window answers differently on purpose is the mission pool.
form declares a Toggle per mission, which is a sensible row and a poor screen
when there are twenty-six of them; the window renders those rows as a checkable
table with columns instead. Same rows, same IDs, same values written back. That
is the kind of difference an interface is for.

Nothing is saved until Save: the widgets are read back in one pass through
form.Apply when the button is pressed, so cancelling leaves the file alone.
*/
package gui

import (
	"context"
	"maps"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/walk"
	declarative "github.com/lxn/walk/declarative"

	"github.com/m-this/tf2-archipelago/launcher/internal/assets"
	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/debugbundle"
	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	"github.com/m-this/tf2-archipelago/launcher/internal/generate"
	"github.com/m-this/tf2-archipelago/launcher/internal/installer"
	"github.com/m-this/tf2-archipelago/launcher/internal/roomcheck"
	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/tailscalefastdl"
	"github.com/m-this/tf2-archipelago/launcher/internal/winproc"
)

// labelWidth keeps the values in one column whichever page is on screen.
const labelWidth = 190

// sentenceWidth caps the paragraphs above a page's rows.
const sentenceWidth = 980

/*
	settingsDialog is the window built but not yet shown.

Building and running are separate so that a test can build the real dialog,
walk the widgets walk actually made and read them back, without a modal loop
nobody can answer. The test runs under Wine, which is enough to create the
window and its controls: settingsdialog_windows_test.go is the whole point of
this split.
*/
type settingsDialog struct {
	dialog *walk.Dialog
	tabs   *walk.TabWidget
	screen *windowScreen
	pool   *poolModel
	model  form.Model

	// collect reads every widget back into a state, which is what Save does
	// and what a test asserts on. persist is the write it then attempts, and
	// refuse is what it does when either says no.
	collect func() (form.State, error)
	persist func(settings.Settings) (settings.Settings, error)

	// note is a refusal short of the message box, which nothing can click.
	note func(page, reason string) string

	// saved is what Save accepted, and ok says whether it was pressed.
	saved settings.Settings
	ok    bool
}

/*
	runSettingsDialog shows the settings and returns what was saved.

The dialog is modal and the state is edited in a copy, so cancelling is not an
undo but a discard: nothing was written.
*/
func runSettingsDialog(
	owner walk.Form, s settings.Settings,
	persist func(settings.Settings) (settings.Settings, error),
	repair func() ([]string, error), reset func() error,
	say func(format string, args ...any), openOn string,
) (settings.Settings, bool, error) {
	built, err := buildSettingsDialog(owner, s, persist, repair, reset, say)
	if err != nil {
		return s, false, err
	}
	// A button that names what it is about to change opens the page it changes,
	// by title rather than by index: the pages a beta flag adds move the numbers.
	showPage(built.tabs, openOn)

	if built.dialog.Run() != walk.DlgCmdOK {
		return s, false, nil
	}
	return built.saved, built.ok, nil
}

/*
	buildSettingsDialog makes the window and everything in it.

Actions are dispatched by the ID form declared them under. A row form declares
as an Action with nothing wired to it here would draw a button that does
nothing, which nobody reports as a bug: they report the feature as missing. The
test refuses one, and the log says so as well for a build nobody tested.
*/
func buildSettingsDialog(
	owner walk.Form, s settings.Settings,
	persist func(settings.Settings) (settings.Settings, error),
	repair func() ([]string, error), reset func() error,
	say func(format string, args ...any),
) (*settingsDialog, error) {
	var (
		dialog    *walk.Dialog
		accept    *walk.PushButton
		cancel    *walk.PushButton
		tabs      *walk.TabWidget
		poolView  *walk.TableView
		state     = form.NewState(s)
		available = communityPackNames(s.CommunityContentDir)
		screen    = &windowScreen{}
		builtAt   = time.Now()
	)

	env := func() form.Env {
		return form.Env{CommunityAvailable: available, AppDirDefault: defaultAppDir()}
	}
	model := form.Build(state, env())
	pool := newPoolModel(model)

	/*
		collect reads every widget back into the state, in one pass.

		Reading before the dialog closes is not optional: closing destroys the
		children, and a destroyed LineEdit reads back empty. It stops at the
		first value form refuses, because a later row's bounds can depend on an
		earlier row's value.
	*/
	collect := func() (form.State, error) {
		next := state
		for _, read := range screen.reads {
			applied, err := form.Apply(next, env(), read())
			if err != nil {
				return next, err
			}
			next = applied
		}
		return pool.apply(next, env())
	}

	pages := make([]declarative.TabPage, 0, len(model.Tabs))
	for _, page := range model.Tabs {
		pages = append(pages, screen.page(page, &poolView, pool))
	}

	built := &settingsDialog{screen: screen, pool: pool, model: model, collect: collect, persist: persist, saved: s}

	err := declarative.Dialog{
		AssignTo:     &dialog,
		Title:        "Settings",
		CancelButton: &cancel,
		// Wide: a seat is a class and the weapons it carries on one line, and
		// the mission table wants six columns.
		Size:    declarative.Size{Width: 1180, Height: 720},
		MinSize: declarative.Size{Width: 900, Height: 560},
		Layout:  declarative.VBox{},
		Children: []declarative.Widget{
			declarative.TabWidget{AssignTo: &tabs, Pages: pages},
			declarative.Composite{
				Layout:  declarative.HBox{},
				MaxSize: declarative.Size{Height: 34},
				Children: []declarative.Widget{
					declarative.HSpacer{},
					declarative.PushButton{AssignTo: &accept, Text: "Save", OnClicked: func() {
						next, err := collect()
						if err != nil {
							refuse(dialog, tabs, say, "", err.Error())
							return
						}
						/* A room that will not parse no longer refuses the
						   save. It used to, and that is how a player lost a
						   login token they were setting two pages away: one
						   field they could not see blocked every other answer
						   in the window. An address that is not one is left out
						   and said afterwards, with the room that was set and
						   did not answer, because both leave the player in the
						   same place. */
						room, roomErr := settings.ParseRoom(next.Draft.Room)
						if roomErr == nil {
							next.Settings.APHost, next.Settings.APPort, next.Settings.APTls = room.Host, room.Port, room.TLS
						} else if strings.TrimSpace(next.Draft.Room) == "" {
							next.Settings.APHost, next.Settings.APPort = "", 0
						}

						/*
							The file is written here, before the dialog closes.

							It used to close first and hand the settings back to
							the caller, which wrote them and put a line in the
							main window's log when it could not. By then this
							dialog was gone and so was everything typed into it,
							and the player was looking at a log pane they had no
							reason to be reading. One on Discord read that as the
							Save button doing nothing at all, went hunting for a
							config.json in the install root, and found none:
							there is none there to find, it goes under the OS's
							own config directory.

							So the write happens while the window is still up,
							and a failure is a message box naming the path.
						*/
						written, err := persist(next.Settings)
						if err != nil {
							refuse(dialog, tabs, say, "", err.Error())
							return
						}

						/* A reach with no token is a server every client is
						   refused from, and nothing on screen would say why. It
						   is not a reason to refuse the save, though: the
						   settings are good and the server simply stays on the
						   local network, so it is said and the window closes. */
						if complaint := tokenComplaint(written.SrcdsReach, written.SrcdsToken); complaint != "" {
							say("settings: %s", complaint)
						}
						say("settings saved")
						built.saved, built.ok = written, true
						dialog.Accept()

						// After the write and after the window is gone: the
						// settings are safe either way and this only ever has
						// advice, never a refusal.
						reportRoom(owner, written, roomErr, strings.TrimSpace(next.Draft.Room), say)
					}},
					declarative.PushButton{AssignTo: &cancel, Text: "Cancel", OnClicked: func() { dialog.Cancel() }},
				},
			},
		},
	}.Create(owner)
	if err != nil {
		return nil, err
	}
	built.dialog, built.tabs = dialog, tabs
	built.note = func(page, reason string) string { return noteRefusal(tabs, say, page, reason) }
	screen.applyCues(say)

	// What pressing a button does. form said what each is called and what it is
	// for; this is the work, and it is the window's because it ends in a
	// message box or a file dialog.
	screen.actions = map[string]func(){
		"run.generate":         func() { generateSeed(dialog, settingsFrom(collect)) },
		"run.open_player_file": func() { openPlayerFile(dialog, settingsFrom(collect)) },
		"run.open_folder": func() {
			next, err := collect()
			if err != nil {
				refuse(dialog, tabs, say, "", err.Error())
				return
			}
			openFolder(dialog, next.Settings.InstallRoot)
		},
		"run.open_settings_file":  func() { openSettingsFile(dialog) },
		"missions.download_packs": func() { downloadSelectedCommunityAssets(dialog, settingsFrom(collect), say) },
		"missions.use_local_packs": func() {
			useLocalCommunityAssets(dialog, settingsFrom(collect), say)
			available = communityPackNames(state.Settings.CommunityContentDir)
		},
		"missions.check_selection": func() { checkRunSelection(dialog, settingsFrom(collect)) },
		"missions.pool_all":        func() { pool.setAll(true) },
		"missions.pool_none":       func() { pool.setAll(false) },
		"server.debug_bundle":      func() { saveDebugBundle(dialog, s) },
		"server.repair":            func() { runRepair(dialog, repair) },
		"server.reset":             func() { runReset(dialog, reset) },
		"net.check_funnel":         func() { checkTailscaleFunnel(dialog, say) },
		"bots.save_team":           func() { saveTeamPreset(dialog, collect, &state) },
		"bots.remove_team":         func() { removeTeamPreset(dialog, collect, &state) },
		"loadout.save":             func() { saveLoadoutPreset(dialog, collect, &state) },
	}

	/* How long the window took to build, in the log a debug bundle carries. A
	   settings dialog that takes a second to appear is a complaint, and a
	   complaint without a number is a guess about which of sixty widgets is the
	   expensive one. */
	for _, page := range model.Tabs {
		for _, field := range page.Fields {
			if field.Kind != form.Action && field.Kind != form.Confirm {
				continue
			}
			if screen.actions[field.ID] == nil {
				say("the settings window draws the button %q and does nothing with it", field.ID)
			}
		}
	}

	say("the settings window took %s to build", time.Since(builtAt).Round(time.Millisecond))

	return built, nil
}

// settingsFrom is a collect that hands back only the settings, for an action
// that has no use for the draft.
func settingsFrom(collect func() (form.State, error)) func() (settings.Settings, error) {
	return func() (settings.Settings, error) {
		state, err := collect()
		return state.Settings, err
	}
}

func checkRunSelection(owner walk.Form, collect func() (settings.Settings, error)) {
	next, err := collect()
	if err != nil {
		walk.MsgBox(owner, "Archipelago run selection", err.Error(), walk.MsgBoxIconWarning)
		return
	}
	result, _ := settings.CheckRunSelection(next)
	walk.MsgBox(owner, "Archipelago run selection", result.Summary(), walk.MsgBoxIconInformation)
}

func saveTeamPreset(owner walk.Form, collect func() (form.State, error), state *form.State) {
	next, err := collect()
	if err != nil {
		walk.MsgBox(owner, "Save this team", err.Error(), walk.MsgBoxIconWarning)
		return
	}
	name := strings.TrimSpace(next.Draft.TeamName)
	if name == "" {
		walk.MsgBox(owner, "Save this team", "Name the team first.", walk.MsgBoxIconWarning)
		return
	}
	presets := map[string]settings.BotTeam{}
	for existing, team := range next.Settings.SrcdsBotTeamPresets {
		presets[existing] = team
	}
	presets[name] = settings.BotTeamOf(next.Settings)
	next.Settings.SrcdsBotTeamPresets = presets
	*state = next
	walk.MsgBox(owner, "Save this team", "Saved the team as "+name+".", walk.MsgBoxIconInformation)
}

func removeTeamPreset(owner walk.Form, collect func() (form.State, error), state *form.State) {
	next, err := collect()
	if err != nil {
		walk.MsgBox(owner, "Remove this team", err.Error(), walk.MsgBoxIconWarning)
		return
	}
	name := strings.TrimSpace(next.Draft.TeamName)
	if _, found := next.Settings.SrcdsBotTeamPresets[name]; !found {
		walk.MsgBox(owner, "Remove this team", "No team is saved as "+name+".", walk.MsgBoxIconWarning)
		return
	}
	presets := map[string]settings.BotTeam{}
	for existing, team := range next.Settings.SrcdsBotTeamPresets {
		if existing != name {
			presets[existing] = team
		}
	}
	if len(presets) == 0 {
		presets = nil
	}
	next.Settings.SrcdsBotTeamPresets = presets
	*state = next
	walk.MsgBox(owner, "Remove this team", "Removed the team "+name+".", walk.MsgBoxIconInformation)
}

func saveLoadoutPreset(owner walk.Form, collect func() (form.State, error), state *form.State) {
	next, err := collect()
	if err != nil {
		walk.MsgBox(owner, "Save this loadout", err.Error(), walk.MsgBoxIconWarning)
		return
	}
	name := strings.TrimSpace(next.Draft.LoadoutName)
	if name == "" {
		walk.MsgBox(owner, "Save this loadout", "Name the loadout first.", walk.MsgBoxIconWarning)
		return
	}
	built := maps.Clone(next.Settings.SrcdsBotCustomLoadouts)
	if built == nil {
		built = map[string]botloadout.Built{}
	}
	built[name] = next.Draft.Loadout
	next.Settings.SrcdsBotCustomLoadouts = built
	*state = next
	walk.MsgBox(owner, "Save this loadout", "Saved the loadout as "+name+".", walk.MsgBoxIconInformation)
}

/* botTeamEditor is the Bots tab's team, as the widgets hold it.
 *
 * Load and Save are two buttons that read and write every seat, every class
 * tick and every weapons menu at once, so they need all of them in one place.
 * Passing nine slices to two callbacks is what this is instead of.
 */

// botSeats is how many bots RED can hold, which is the team size the Bots tab
// caps at. A seat past the team size is simply never filled.
const botSeats = 6

// drawSeatLoadoutLabel is what a seat with no class of its own can hold, which
// is nothing to choose: the mod draws the class and the class holds its own.
const drawSeatLoadoutLabel = "follows the class"

// drawSeatLabel is the first entry of every seat menu: the mod picks.
const drawSeatLabel = "Let the mod pick"

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
		"Wrote "+path+"\n\nIt holds this run's launcher log and the one before "+
			"it, the SourceMod logs, the server console, what the bridge says "+
			"about the run, the player file and the settings. The passwords are "+
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

// runReset puts every setting back to its default, for a player whose answers
// have drifted somewhere they cannot see and cannot undo.
//
// It closes the dialog rather than redrawing it. Every field on screen still
// holds the old answer, and Save would write all of them straight back over
// the reset.
func runReset(owner walk.Form, reset func() error) {
	answer := walk.MsgBox(owner, "Reset settings",
		"This puts every setting back to what a fresh install has: the room, "+
			"the server name and its passwords, the missions, the bots, who can join.\n\n"+
			"It keeps the game files and where they are, so nothing is downloaded again.",
		walk.MsgBoxOKCancel|walk.MsgBoxIconQuestion)
	if answer != walk.DlgCmdOK {
		return
	}
	if err := reset(); err != nil {
		walk.MsgBox(owner, "Reset settings", err.Error(), walk.MsgBoxIconError)
		return
	}
	walk.MsgBox(owner, "Reset settings",
		"Every setting is back to its default. Open Settings again to go through them.",
		walk.MsgBoxIconInformation)
	owner.(*walk.Dialog).Cancel()
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

/*
	The community packs, and Tailscale, as the model-driven window asks for them

Each used to take half a dozen widget pointers so it could grey a button out and
colour a status label. The rows are declared in internal/form now and the window
does not hold a pointer to any of them, so each of these takes what it needs and
says what happened in a message box. The long ones still run off the UI thread,
because a download and a tailnet round trip both outlast a click.
*/
func downloadSelectedCommunityAssets(owner *walk.Dialog, collect func() (settings.Settings, error), say func(string, ...any)) {
	next, err := collect()
	if err != nil {
		walk.MsgBox(owner, "Download community assets", err.Error(), walk.MsgBoxIconWarning)
		return
	}
	folder := strings.TrimSpace(next.CommunityContentDir)
	if folder == "" {
		walk.MsgBox(owner, "Download community assets", "Choose an asset pack folder first.", walk.MsgBoxIconWarning)
		return
	}
	archives := settings.CommunityArchives(next)
	if len(archives) == 0 {
		walk.MsgBox(owner, "Download community assets", "Tick at least one community pack first.", walk.MsgBoxIconWarning)
		return
	}

	go apruntime.Guard("a settings task", func(t string) { say("%s", t) }, func() {
		logf := func(format string, args ...any) { say("community assets: "+format, args...) }
		err := installer.DownloadCommunityArchives(context.Background(), archives, logf)
		owner.Synchronize(func() {
			if owner.IsDisposed() {
				return
			}
			if err != nil {
				walk.MsgBox(owner, "Download community assets", err.Error(), walk.MsgBoxIconError)
				return
			}
			walk.MsgBox(owner, "Download community assets",
				"Selected community packs are ready in "+folder+".\n\nClose and reopen the settings to see their missions.",
				walk.MsgBoxIconInformation)
		})
	})
}

func useLocalCommunityAssets(owner *walk.Dialog, collect func() (settings.Settings, error), say func(string, ...any)) {
	next, err := collect()
	if err != nil {
		walk.MsgBox(owner, "Use local community assets", err.Error(), walk.MsgBoxIconWarning)
		return
	}
	folder := strings.TrimSpace(next.CommunityContentDir)
	if folder == "" {
		walk.MsgBox(owner, "Use local community assets", "Choose an asset pack folder first.", walk.MsgBoxIconWarning)
		return
	}
	packs := communityPackNames(folder)
	if len(packs) == 0 {
		walk.MsgBox(owner, "Use local community assets",
			"No valid archive-assets.zip or mlarchive-assets.zip was found in "+folder+".", walk.MsgBoxIconWarning)
		return
	}
	say("community assets: using local packs from %s", folder)
	walk.MsgBox(owner, "Use local community assets",
		"Using local community packs from "+folder+".\n\nClose and reopen the settings to see their missions.",
		walk.MsgBoxIconInformation)
}

// communityPackNames are the packs with a valid ZIP in that folder, by the name
// the settings key them under.
func communityPackNames(folder string) []string {
	var packs []string
	for _, path := range installer.AvailableCommunityArchives(settings.KnownCommunityArchives(strings.TrimSpace(folder))) {
		name := filepath.Base(path)
		if name == settings.CommunityPackPotato || name == settings.CommunityPackMoonlight {
			packs = append(packs, name)
		}
	}
	return packs
}

func checkTailscaleFunnel(owner *walk.Dialog, say func(string, ...any)) {
	go apruntime.Guard("a settings task", func(text string) { say("%s", text) }, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		result, err := tailscalefastdl.Authorize(ctx)
		owner.Synchronize(func() {
			if owner.IsDisposed() {
				return
			}
			switch {
			case err != nil:
				walk.MsgBox(owner, "Tailscale Funnel", err.Error(), walk.MsgBoxIconError)
			case result.ApprovalURL != "":
				if err := winproc.OpenURL(result.ApprovalURL); err != nil {
					walk.MsgBox(owner, "Enable Tailscale Funnel",
						result.ApprovalURL+"\n\n"+err.Error(), walk.MsgBoxIconWarning)
					return
				}
				walk.MsgBox(owner, "Enable Tailscale Funnel",
					"Approve Funnel in the browser, then run this check again.", walk.MsgBoxIconInformation)
			case result.Ready:
				walk.MsgBox(owner, "Tailscale Funnel", "Funnel is enabled for this tailnet.", walk.MsgBoxIconInformation)
			}
		})
	})
}

// browseForFolder puts the picker on a folder row. It writes into the edit
// rather than into the state, because the state is read from the widgets when
// Save is pressed and this has to look the same as typing the path.
func browseForFolder(edit *walk.LineEdit) {
	dialog := walk.FileDialog{
		Title:          "Which folder?",
		InitialDirPath: strings.TrimSpace(edit.Text()),
	}
	accepted, err := dialog.ShowBrowseFolder(edit.Form())
	if err != nil || !accepted || dialog.FilePath == "" {
		return
	}
	_ = edit.SetText(dialog.FilePath)
}

// showPage opens the tab with that title, and leaves the window on the first
// page for a title no page carries.
func showPage(tabs *walk.TabWidget, title string) {
	if title == "" {
		return
	}
	for i := range tabs.Pages().Len() {
		if tabs.Pages().At(i).Title() == title {
			_ = tabs.SetCurrentIndex(i)
			return
		}
	}
}

/*
	refuse stops the Save, opens the page at fault, and says why.

A debug bundle from a player is why this exists. They were setting a login token
on one page, pressed Save, and nothing happened: the Save was refusing on the
room address two pages away, next to a field they were not looking at. Their log
had no line about saving at all, because a refused Save wrote none, so nothing
in the bundle said what had gone wrong either.

A message box cannot be missed, the page behind it is the one holding the
problem, and the log carries the reason for whoever reads the next bundle. page
is empty for a failure that belongs to no page, such as a file that would not
write.
*/
func refuse(dialog *walk.Dialog, tabs *walk.TabWidget, say func(string, ...any), page, reason string) {
	walk.MsgBox(dialog, "The settings were not saved", noteRefusal(tabs, say, page, reason),
		walk.MsgBoxIconError)
}

/*
	noteRefusal is everything about a refusal except putting the box up.

Split off because a modal message box is the one thing a test cannot drive:
walk.MsgBox does not return until somebody clicks it, and under Wine nobody
does. So the page switch, the log line and the wording are here where they can
be checked, and refuse above is the two lines that cannot be.
*/
func noteRefusal(tabs *walk.TabWidget, say func(string, ...any), page, reason string) string {
	showPage(tabs, page)
	say("settings not saved: %s", reason)
	return reason + "\n\nNothing in this window is lost. Fix what this names and press Save again."
}

/*
	openSettingsFile shows where the launcher keeps its own settings.

Not the install folder, which is the mistake it exists to correct. A player went
looking for config.json in the install folder, found none, and read that as
nothing having saved at all. The file is under the OS's config directory and
nothing in the launcher would show it.

The folder is opened rather than the file, because config.json has no
application to open it with on a fresh Windows install, and the path is in the
message either way so it can be copied.
*/
func openSettingsFile(owner walk.Form) {
	path, err := settings.Path()
	if err != nil {
		walk.MsgBox(owner, "The settings file",
			"Cannot work out where the settings live: "+err.Error(), walk.MsgBoxIconError)
		return
	}
	if err := winproc.Open(filepath.Dir(path)); err != nil {
		walk.MsgBox(owner, "The settings file",
			"The settings are at\n\n"+path+"\n\nand that folder cannot be opened: "+err.Error(),
			walk.MsgBoxIconWarning)
		return
	}
}

/*
	reportRoom says whether the room is there, once the settings are written.

Saving is when a player finds out their address works, and it used to be much
later: Start refused with "AP_PORT is not set", or the bridge retried a dead
room in a log nobody was reading. One player spent seventeen minutes on that.

It runs after the window is gone and cannot undo the save, which is the point.
A room that does not answer is advice: the settings are worth keeping either
way, and a player who has not made the room yet has done nothing wrong. Only the
cases with something to say put a box up, so a working room is a log line and
nothing else.
*/
func reportRoom(owner walk.Form, s settings.Settings, roomErr error, typed string, say func(string, ...any)) {
	if roomErr != nil && typed != "" {
		say("settings: the room address was not saved: %v", roomErr)
		walk.MsgBox(owner, "The Archipelago room",
			"The room address was not saved: "+roomErr.Error()+".\n\n"+roomcheck.NotConfigured.Advice(),
			walk.MsgBoxIconWarning)
		return
	}

	result, err := roomcheck.Check(context.Background(), s)
	switch result {
	case roomcheck.Live:
		say("settings: the Archipelago room answered")
	case roomcheck.Skipped:
		say("settings: test mode, so no room is needed")
	case roomcheck.NotConfigured:
		say("settings: no Archipelago room is set")
		walk.MsgBox(owner, "No Archipelago room", result.Advice(), walk.MsgBoxIconInformation)
	default:
		say("settings: the room did not answer: %v", err)
		walk.MsgBox(owner, "The Archipelago room",
			"The room did not answer: "+err.Error()+".\n\n"+result.Advice(),
			walk.MsgBoxIconWarning)
	}
}
