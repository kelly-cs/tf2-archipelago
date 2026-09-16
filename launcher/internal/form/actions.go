package form

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/botnames"
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
	case "bots.name_add":
		next, said = addBotName(s)
	case "loadout.save":
		next, said = saveLoadout(s)
	default:
		return s, "", false
	}
	return next, said, true
}

/*
addBotName puts the name in the box into the pool the bots draw from.

Every refusal here is a name that would have gone in and come out looking like
something else: the game cuts a long one, the mod draws by index so a duplicate
comes up twice as often, and a comma is what the Compose stack separates the
list with.
*/
func addBotName(s State) (State, string) {
	name := strings.TrimSpace(s.Draft.BotName)
	switch {
	case name == "":
		return s, "type a name first"
	case len(name) > botnames.NameMax:
		return s, fmt.Sprintf("%q is longer than the %d characters the game keeps", name, botnames.NameMax)
	case strings.Contains(name, ","):
		return s, "a name cannot hold a comma, which is what separates them in a Compose .env"
	case len(s.Settings.SrcdsBotNamesAdded) >= botnames.AddedMax:
		return s, fmt.Sprintf("that is %d names of your own, which is all this page can show", botnames.AddedMax)
	case slices.Contains(botnames.Pool(s.Settings.SrcdsBotNamesExcluded, s.Settings.SrcdsBotNamesAdded), name):
		return s, name + " is already in the pool"
	}

	// A name that was shipped and taken out comes back rather than being added
	// beside itself, so the tick above it says what the pool holds.
	if slices.Contains(botnames.Shipped(), name) {
		s.Settings.SrcdsBotNamesExcluded = slices.DeleteFunc(
			slices.Clone(s.Settings.SrcdsBotNamesExcluded),
			func(one string) bool { return one == name })
		s.Draft.BotName = ""
		return s, "put " + name + " back in the pool"
	}

	s.Settings.SrcdsBotNamesAdded = append(slices.Clone(s.Settings.SrcdsBotNamesAdded), name)
	s.Draft.BotName = ""
	return s, "added " + name
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
