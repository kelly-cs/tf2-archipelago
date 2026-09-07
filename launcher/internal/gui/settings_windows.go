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
	runSettingsDialog shows the settings and returns what was saved.

The dialog is modal and the state is edited in a copy, so cancelling is not an
undo but a discard: nothing was written.

Actions are dispatched by the ID form declared them under. A row form declares
as an Action with nothing wired to it here would draw a button that does
nothing, so the dialog says so in the log rather than shipping a dead button
that nobody reports as a bug: they report the feature as missing. The window
cannot be tested on the machine that builds it, so this is a check at run time
where the terminal has one at test time.
*/
func runSettingsDialog(
	owner walk.Form, s settings.Settings,
	repair func() ([]string, error), reset func() error,
	say func(format string, args ...any), openOn string,
) (settings.Settings, bool, error) {
	var (
		dialog    *walk.Dialog
		accept    *walk.PushButton
		cancel    *walk.PushButton
		tabs      *walk.TabWidget
		poolView  *walk.TableView
		warn      *walk.Label
		state     = form.NewState(s)
		available = communityPackNames(s.CommunityContentDir)
		edited    = s
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
			declarative.Label{AssignTo: &warn, Text: ""},
			declarative.Composite{
				Layout:  declarative.HBox{},
				MaxSize: declarative.Size{Height: 34},
				Children: []declarative.Widget{
					declarative.HSpacer{},
					declarative.PushButton{AssignTo: &accept, Text: "Save", OnClicked: func() {
						next, err := collect()
						if err != nil {
							warn.SetText(err.Error())
							return
						}
						/* The address is checked here rather than as it is
						   typed. A disabled button with no explanation is a
						   dead end, and a paste with the mouse sends no
						   keystroke to check on. Test mode never dials a real
						   room, so it does not need one. */
						room, err := settings.ParseRoom(next.Draft.Room)
						if err != nil && !next.Settings.TestMode {
							warn.SetText(err.Error())
							return
						}
						next.Settings.APHost, next.Settings.APPort, next.Settings.APTls = room.Host, room.Port, room.TLS

						/* A reach that leaves the network with no token is a
						   server every client is refused from, and nothing on
						   screen would say why. Refuse the save instead. */
						if complaint := tokenComplaint(next.Settings.SrcdsReach, next.Settings.SrcdsToken); complaint != "" {
							warn.SetText(complaint)
							return
						}
						edited = next.Settings
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

	// What pressing a button does. form said what each is called and what it is
	// for; this is the work, and it is the window's because it ends in a
	// message box or a file dialog.
	screen.actions = map[string]func(){
		"run.generate":            func() { generateSeed(dialog, settingsFrom(collect)) },
		"run.open_player_file":    func() { openPlayerFile(dialog, settingsFrom(collect)) },
		"run.open_folder":         func() { openFolder(dialog, s.InstallRoot) },
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

	// A button that names what it is about to change opens the page it changes,
	// by title rather than by index: the pages a beta flag adds move the numbers.
	showPage(tabs, openOn)

	if dialog.Run() != walk.DlgCmdOK {
		return s, false, nil
	}
	return edited, true, nil
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
