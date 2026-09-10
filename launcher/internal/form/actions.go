package form

import (
	"maps"
	"strings"

	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/*
Act is what a button does when its whole effect is on the draft: keeping the
seats under a name, forgetting a kept team, keeping the loadout being built.
It answers with the state after the press and the line the player is told.

Here beside Apply rather than on the launcher, because these are the buttons
whose answer is a State and nothing else: no download, no folder, no server.
The launcher dispatches the rest, and the fake one answers these the same way
the real one does because both ask here. ok is false for a button this does
not know, which the caller then owns.
*/
func Act(s State, id string) (next State, said string, ok bool) {
	switch id {
	case "bots.save_team":
		next, said = saveTeam(s)
	case "bots.remove_team":
		next, said = removeTeam(s)
	case "loadout.save":
		next, said = saveLoadout(s)
	default:
		return s, "", false
	}
	return next, said, true
}

func saveTeam(s State) (State, string) {
	name := strings.TrimSpace(s.Draft.TeamName)
	if name == "" {
		return s, "name the team first"
	}
	presets := maps.Clone(s.Settings.SrcdsBotTeamPresets)
	if presets == nil {
		presets = map[string]settings.BotTeam{}
	}
	presets[name] = settings.BotTeamOf(s.Settings)
	s.Settings.SrcdsBotTeamPresets, s.Draft.TeamName = presets, ""
	return s, "saved the team as " + name
}

func removeTeam(s State) (State, string) {
	name := strings.TrimSpace(s.Draft.TeamName)
	if name == "" {
		return s, "name the team to remove first"
	}
	if _, kept := s.Settings.SrcdsBotTeamPresets[name]; !kept {
		return s, "no team saved as " + name
	}
	presets := maps.Clone(s.Settings.SrcdsBotTeamPresets)
	delete(presets, name)
	if len(presets) == 0 {
		presets = nil
	}
	s.Settings.SrcdsBotTeamPresets, s.Draft.TeamName = presets, ""
	return s, "removed the team " + name
}

func saveLoadout(s State) (State, string) {
	name := strings.TrimSpace(s.Draft.LoadoutName)
	if name == "" {
		return s, "name the loadout first"
	}
	built := maps.Clone(s.Settings.SrcdsBotCustomLoadouts)
	if built == nil {
		built = map[string]botloadout.Built{}
	}
	built[name] = s.Draft.Loadout
	s.Settings.SrcdsBotCustomLoadouts = built
	return s, "saved the loadout as " + name
}
