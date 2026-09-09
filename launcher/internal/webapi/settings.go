package webapi

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/assets"
	"github.com/m-this/tf2-archipelago/launcher/internal/botlive"
	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/form"
	"github.com/m-this/tf2-archipelago/launcher/internal/generate"
	"github.com/m-this/tf2-archipelago/launcher/internal/installer"
	"github.com/m-this/tf2-archipelago/launcher/internal/roomcheck"
	"github.com/m-this/tf2-archipelago/launcher/internal/runshape"
	"github.com/m-this/tf2-archipelago/launcher/internal/saveplan"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/srcdsconfig"
	"github.com/m-this/tf2-archipelago/launcher/internal/winproc"
)

func (a *App) DraftSettings() (settings.Settings, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft == nil {
		return settings.Settings{}, errors.New("settings are not open")
	}
	return a.draft.Settings, nil
}

func (a *App) OpenSettings(page string) {
	a.mu.Lock()
	state := form.NewState(a.settings)
	a.draft, a.formPage = &state, page
	a.community = availableCommunityPackNames(a.settings.CommunityContentDir)
	a.imported = importedCommunityPackNames(a.settings.CommunityContentDir, a.community)
	a.publishLocked(Event{Name: "state", Data: struct{}{}})
	a.mu.Unlock()
}

func (a *App) Change(c form.Change) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft == nil {
		return errors.New("settings are not open")
	}
	next, err := form.Apply(*a.draft, a.formEnvLocked(), c)
	if err != nil {
		return err
	}
	*a.draft = next
	a.publishLocked(Event{Name: "state", Data: struct{}{}})
	return nil
}

func (a *App) SaveSettings(restart bool) error {
	a.mu.Lock()
	if a.draft == nil {
		a.mu.Unlock()
		return errors.New("settings are not open")
	}
	draft := *a.draft
	a.mu.Unlock()
	room, roomErr := settings.ParseRoom(draft.Draft.Room)
	if roomErr == nil {
		draft.Settings.APHost, draft.Settings.APPort, draft.Settings.APTls = room.Host, room.Port, room.TLS
	} else if strings.TrimSpace(draft.Draft.Room) == "" {
		draft.Settings.APHost, draft.Settings.APPort = "", 0
	}
	written, err := settings.Persist(draft.Settings)
	if err != nil {
		return err
	}
	before := a.supervisor.Settings()
	a.mu.Lock()
	a.settings, a.draft = written, nil
	a.notice = "settings saved"
	a.noticeSeq++
	heldRestart := a.smHeld
	a.smHeld = false
	a.mu.Unlock()
	a.supervisor.SetSettings(written)
	if _, err := settings.WritePlayerFile(written, assets.ArchipelagoVersion); err != nil {
		a.Say("%v", err)
	}
	if a.supervisor.Running() {
		plan := saveplan.For(before, written)
		if plan.Team {
			if err := srcdsconfig.Install(written); err != nil {
				a.Say("cannot write the bot files: %v", err)
			} else {
				for _, command := range botlive.Commands(before, written) {
					a.SendRCON(command)
				}
			}
		}
		switch {
		case plan.Restart && restart:
			a.Say("settings saved. Restarting the server to apply them.")
			a.Restart()
		case heldRestart:
			a.Say("SourceMod updated its gamedata. Restarting the server to load it.")
			a.Restart()
		case plan.Restart:
			a.Say("settings saved. The server is still playing on what it started with: press Restart to apply them.")
		case plan.Quiet():
			a.Say("settings saved. The server keeps playing: nothing here changes a run it is already in.")
		}
	}
	go a.reportRoom(written, draft.Draft.Room, roomErr)
	a.publishState()
	return nil
}

func (a *App) reportRoom(s settings.Settings, typed string, parseErr error) {
	if parseErr != nil && strings.TrimSpace(typed) != "" {
		a.Notify("the room address was not saved: " + parseErr.Error() + ". " + roomcheck.NotConfigured.Advice())
		return
	}
	result, err := roomcheck.Check(context.Background(), s)
	if err != nil {
		a.Notify("settings saved, but the room did not answer: " + err.Error() + ". " + result.Advice())
		return
	}
	a.Notify("settings saved. " + result.Advice())
}

func (a *App) CancelSettings() {
	a.mu.Lock()
	a.draft = nil
	heldRestart := a.smHeld
	a.smHeld = false
	a.publishLocked(Event{Name: "state", Data: struct{}{}})
	a.mu.Unlock()
	if heldRestart {
		a.Say("SourceMod updated its gamedata. Restarting the server to load it.")
		a.Restart()
	}
}

func (a *App) formEnvLocked() form.Env {
	dirs := generate.SearchPath("")
	appDir := ""
	if len(dirs) > 0 {
		appDir = dirs[0]
	}
	return form.Env{CommunityAvailable: slices.Clone(a.community), AppDirDefault: appDir}
}

func (a *App) Dispatch(id string) error {
	a.mu.Lock()
	if a.draft == nil {
		a.mu.Unlock()
		return errors.New("settings are not open")
	}
	if !form.Dispatchable(*a.draft, a.formEnvLocked(), id) {
		a.mu.Unlock()
		return fmt.Errorf("no settings action %q", id)
	}
	s := *a.draft
	a.mu.Unlock()

	switch id {
	case "missions.pool_all", "missions.pool_none":
		a.setPool(id == "missions.pool_all")
	case "bots.save_team":
		a.saveTeam(s)
	case "bots.remove_team":
		a.removeTeam(s)
	case "loadout.save":
		a.saveLoadout(s)
	case "missions.check_selection":
		a.checkMissionSelection(s.Settings)
	case "missions.download_packs":
		go a.downloadPacks(s.Settings)
	case "missions.import_assets":
		return a.useLocalPacks(s.Settings.CommunityContentDir)
	case "server.repair":
		go a.repair(s.Settings.InstallRoot)
	case "server.reset":
		return a.resetSettings()
	default:
		return fmt.Errorf("settings action %q is not wired", id)
	}
	return nil
}

func (a *App) saveTeam(s form.State) {
	name := strings.TrimSpace(s.Draft.TeamName)
	if name == "" {
		a.Notify("name the team first")
		return
	}
	presets := maps.Clone(s.Settings.SrcdsBotTeamPresets)
	if presets == nil {
		presets = map[string]settings.BotTeam{}
	}
	presets[name] = settings.BotTeamOf(s.Settings)
	a.mutateDraft(func(state *form.State) {
		state.Settings.SrcdsBotTeamPresets, state.Draft.TeamName = presets, ""
	})
	a.Notify("saved the team as " + name)
}

func (a *App) removeTeam(s form.State) {
	name := strings.TrimSpace(s.Draft.TeamName)
	if name == "" {
		a.Notify("name the team to remove first")
		return
	}
	if _, ok := s.Settings.SrcdsBotTeamPresets[name]; !ok {
		a.Notify("no team saved as " + name)
		return
	}
	presets := maps.Clone(s.Settings.SrcdsBotTeamPresets)
	delete(presets, name)
	if len(presets) == 0 {
		presets = nil
	}
	a.mutateDraft(func(state *form.State) {
		state.Settings.SrcdsBotTeamPresets, state.Draft.TeamName = presets, ""
	})
	a.Notify("removed the team " + name)
}

func (a *App) saveLoadout(s form.State) {
	name := strings.TrimSpace(s.Draft.LoadoutName)
	if name == "" {
		a.Notify("name the loadout first")
		return
	}
	built := maps.Clone(s.Settings.SrcdsBotCustomLoadouts)
	if built == nil {
		built = map[string]botloadout.Built{}
	}
	built[name] = s.Draft.Loadout
	a.mutateDraft(func(state *form.State) { state.Settings.SrcdsBotCustomLoadouts = built })
	a.Notify("saved the loadout as " + name)
}

func (a *App) checkMissionSelection(s settings.Settings) {
	result, err := settings.CheckRunSelection(s)
	if err != nil {
		a.Notify(err.Error())
		return
	}
	a.Notify(result.Summary())
}

var wiredActions = []string{
	"run.generate", "run.open_player_file", "run.open_folder", "run.open_settings_file",
	"missions.download_packs", "missions.import_assets", "missions.check_selection",
	"missions.pool_all", "missions.pool_none",
	"server.debug_bundle", "server.repair", "server.reset",
	"net.check_funnel", "bots.save_team", "bots.remove_team", "loadout.save",
}

func (a *App) mutateDraft(change func(*form.State)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft != nil {
		change(a.draft)
		a.publishLocked(Event{Name: "state", Data: struct{}{}})
	}
}

func (a *App) setPool(all bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft == nil {
		return
	}
	var excluded []string
	if !all {
		for _, mission := range gamedata.PlayableMissions() {
			excluded = append(excluded, mission.PopFile)
		}
	} else {
		visible := runshape.VisibleMissions(a.community)
		for _, mission := range gamedata.PlayableMissions() {
			if gamedata.MissionPack(mission.ID) != "" && !slices.ContainsFunc(visible, func(candidate gamedata.Mission) bool { return candidate.ID == mission.ID }) {
				excluded = append(excluded, mission.PopFile)
			}
		}
	}
	a.draft.Settings.MvmExcludedMissions = excluded
	if slices.Contains(excluded, a.draft.Settings.MvmStartMission) {
		a.draft.Settings.MvmStartMission = ""
	}
	a.notice = map[bool]string{true: "every mission is in the pool", false: "every mission is left out"}[all]
	a.noticeSeq++
	a.publishLocked(Event{Name: "state", Data: struct{}{}})
}

func (a *App) downloadPacks(s settings.Settings) {
	folder := strings.TrimSpace(s.CommunityContentDir)
	if folder == "" {
		a.Notify("choose an asset pack folder first")
		return
	}
	archives := settings.CommunityArchives(s)
	if len(archives) == 0 {
		a.Notify("select at least one community pack first")
		return
	}
	if err := installer.DownloadCommunityArchives(context.Background(), archives, func(f string, args ...any) { a.Say(f, args...) }); err != nil {
		a.Notify("community assets: " + err.Error())
		return
	}
	a.mu.Lock()
	a.community = availableCommunityPackNames(folder)
	a.mu.Unlock()
	a.Notify("selected community packs are ready in " + folder)
}

func (a *App) repair(root string) {
	a.Stop()
	_, _ = winproc.KillUnder(root)
	removed, err := installer.Clean(root)
	if err != nil {
		a.Notify("repair: " + err.Error())
		return
	}
	if len(removed) == 0 {
		a.Notify("repair: nothing to remove")
	} else {
		a.Notify("repair removed " + strings.Join(removed, ", "))
	}
}

func (a *App) resetSettings() error {
	fresh := settings.Defaults()
	fresh.InstallRoot = a.supervisor.Settings().InstallRoot
	written, err := settings.Persist(fresh)
	if err != nil {
		return err
	}
	a.supervisor.SetSettings(written)
	a.mu.Lock()
	a.settings = written
	state := form.NewState(written)
	a.draft = &state
	a.mu.Unlock()
	a.Notify("every setting is back to its default")
	return nil
}

/*
useLocalPacks takes the asset packs already sitting in the content folder.

The packs run to gigabytes and they are on the same machine as the launcher, so
there is nothing to upload: the player points the content folder at where they
put the zips and this reads what is there. It is what the terminal interface has
always done. The browser's old multipart upload existed only because a page with
no folder picker had no other way to name a file.
*/
func (a *App) useLocalPacks(folder string) error {
	if strings.TrimSpace(folder) == "" {
		return errors.New("choose an asset pack folder first")
	}
	available := availableCommunityPackNames(folder)
	if len(available) == 0 {
		return fmt.Errorf("no archive-assets.zip or mlarchive-assets.zip was found in %s", folder)
	}
	if err := a.AcceptImportedPacks(available); err != nil {
		return err
	}
	a.Notify("using the community packs in " + folder)
	return nil
}
