package settings

import (
	"slices"

	"github.com/m-this/tf2-archipelago/gamedata"
)

/*
ModLoading is when a selected server mod is loaded by the game server.

It is a third answer rather than a second checkbox because the three states are
one decision: a mod can be absent, present for the missions that ask for it, or
present always. A player who wants SigMod's missions and a player who wants
SigMod's behaviour on every map are asking for different things, and the
difference used to be unsayable.
*/
type ModLoading string

const (
	// ModLoadingRequired loads a mod only while the mission pool holds a
	// mission that names it. The default: an extension that patches the engine
	// has no business in a run that never asks for one.
	ModLoadingRequired ModLoading = "required"

	// ModLoadingAlways loads a selected mod on every start, whatever the pool
	// holds. For a host who wants the mod's own behaviour rather than its
	// missions.
	ModLoadingAlways ModLoading = "always"
)

// ModLoadings lists the choices, least surprising first.
func ModLoadings() []ModLoading { return []ModLoading{ModLoadingRequired, ModLoadingAlways} }

// Valid reports whether m is one of the two.
func (m ModLoading) Valid() bool {
	return m == ModLoadingRequired || m == ModLoadingAlways
}

// OrDefault is m, or the default for a settings file written before this
// existed. Those files meant "load it", and required is the safe half of that.
func (m ModLoading) OrDefault() ModLoading {
	if m.Valid() {
		return m
	}
	return ModLoadingRequired
}

// Label is what an interface shows for this choice.
func (m ModLoading) Label() string {
	if m.OrDefault() == ModLoadingAlways {
		return "always, on every mission"
	}
	return "only when a mission needs it"
}

/*
ServerModsToLoad are the selected mods that this run actually loads, which is
what decides whether the launcher writes each one's autoload marker.

A mod with no build on this platform is never loaded, whatever is selected: it
is not there to load.
*/
func ServerModsToLoad(s Settings, goos string) []string {
	loading := s.SrcdsModLoading.OrDefault()
	required := RequiredServerMods(s)
	load := make([]string, 0, len(s.SrcdsMods))
	for _, key := range ServerModKeys(s) {
		if !gamedata.ServerModBuildsOn(key, goos) {
			continue
		}
		if loading == ModLoadingAlways || slices.Contains(required, key) {
			load = append(load, key)
		}
	}
	return load
}
