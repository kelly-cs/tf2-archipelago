package gamedata

// ObjectiveKind is the plugin's vocabulary for what Archipelago calls a
// location. The plugin reports in MvM terms and never learns of an id.
type ObjectiveKind uint8

const (
	ObjectiveWaveCleared ObjectiveKind = iota + 1
	ObjectiveMissionCleared
	ObjectiveTankDestroyed
	ObjectiveGiantKilled

	// The tallies are what the plugin reports at the end of a wave: how many
	// of each fell in it. They resolve to no location of their own; the
	// bridge adds them to a total and the milestones are what that pays.
	ObjectiveTallyRobots
	ObjectiveTallyGiants
	ObjectiveTallyTanks
)

var objectiveKeys = [...]string{
	ObjectiveWaveCleared:    "wave_cleared",
	ObjectiveMissionCleared: "mission_cleared",
	ObjectiveTankDestroyed:  "tank_destroyed",
	ObjectiveGiantKilled:    "giant_killed",
	ObjectiveTallyRobots:    "tally_robots",
	ObjectiveTallyGiants:    "tally_giants",
	ObjectiveTallyTanks:     "tally_tanks",
}

// ObjectiveKinds is every kind that exists, in id order. Whatever walks it
// covers new kinds without a second list to keep in step.
var ObjectiveKinds = []ObjectiveKind{
	ObjectiveWaveCleared, ObjectiveMissionCleared,
	ObjectiveTankDestroyed, ObjectiveGiantKilled,
	ObjectiveTallyRobots, ObjectiveTallyGiants, ObjectiveTallyTanks,
}

// IsTally reports whether this kind counts towards a milestone rather than
// naming a check of its own.
func (k ObjectiveKind) IsTally() bool {
	return k == ObjectiveTallyRobots || k == ObjectiveTallyGiants || k == ObjectiveTallyTanks
}

// Key is the string on the wire between the plugin and the bridge.
func (k ObjectiveKind) Key() string { return objectiveKeys[k] }

// Location is one check. Wave is zero for a mission clear. Cache is zero for
// every check but a victory cache, where it counts from 1. Index is zero for
// every check but a per-kill one, the nth giant or tank of its wave, where it
// counts from 1: those exist only when a sanity option asks for them. A
// milestone has no mission and a Threshold instead: the total its tally has
// to reach.
type Location struct {
	ID        int64
	Name      string
	Kind      ObjectiveKind
	Mission   MissionID
	Wave      uint8
	Cache     uint8
	Index     uint8
	Threshold int
}

/*
VictoryCaches is how many extra checks a clear at this tier pays when the
victory_caches option is on: a normal clear stays one check, and an expert or
haunted one is worth five.

The ladder is Roseburst's (gh-83): 1, 2, 3, 5. It is the cheap answer to a run
being short on checks, because it adds no mission, scrubs no pop file and
needs no new event from the plugin: a clear is reported once and the bridge
records every cache the seed holds for it.
*/
func (d Difficulty) VictoryCaches() uint8 {
	switch d {
	case DifficultyIntermediate:
		return 1
	case DifficultyAdvanced:
		return 2
	case DifficultyExpert, DifficultyHaunted:
		return 4
	default:
		return 0
	}
}

// VictoryCacheLocations is every cache this mission's clear can pay, in order.
// Empty for a normal mission.
func (m Mission) VictoryCacheLocations() []Location {
	count := m.Difficulty.VictoryCaches()
	caches := make([]Location, 0, count)
	for n := uint8(1); n <= count; n++ {
		caches = append(caches, Location{
			ID:      m.VictoryCacheLocationID(n),
			Name:    m.VictoryCacheLocationName(n),
			Kind:    ObjectiveMissionCleared,
			Mission: m.ID,
			Cache:   n,
		})
	}
	return caches
}

// WaveKillsAt is what wave w of this mission spawns that can be paid for,
// out of the committed count. Nothing for a wave the count does not cover,
// which is every community mission: their files are not scrubbed.
func (m Mission) WaveKillsAt(wave uint8) WaveKills {
	waves := waveKillsByPopFile[m.PopFile]
	if wave < 1 || int(wave) > len(waves) {
		return WaveKills{}
	}
	return waves[wave-1]
}

// WaveKillLocations is every per-kill check this mission can pay, wave by
// wave: the giants of the wave, then its tanks.
func (m Mission) WaveKillLocations() []Location {
	var kills []Location
	for wave := uint8(1); wave <= m.Waves; wave++ {
		counts := m.WaveKillsAt(wave)
		for n := uint8(1); n <= counts.Giants; n++ {
			kills = append(kills, Location{
				ID: m.WaveGiantLocationID(wave, n), Name: m.WaveGiantLocationName(wave, n),
				Kind: ObjectiveGiantKilled, Mission: m.ID, Wave: wave, Index: n,
			})
		}
		for n := uint8(1); n <= counts.Tanks; n++ {
			kills = append(kills, Location{
				ID: m.WaveTankLocationID(wave, n), Name: m.WaveTankLocationName(wave, n),
				Kind: ObjectiveTankDestroyed, Mission: m.ID, Wave: wave, Index: n,
			})
		}
	}
	return kills
}

// Locations is every check in the game, mission by mission: the waves in
// order, then the tank and the giant if the mission holds them, then the
// mission clear, then the victory caches its tier pays, which the apworld
// includes only when the option asks for them, then every giant and tank of
// every wave, held only when giantsanity or tanksanity asks for them; and
// after every mission, the milestones.
var Locations = buildLocations()

func buildLocations() []Location {
	all := make([]Location, 0, len(Missions)*10)
	for _, m := range Missions {
		for wave := uint8(1); wave <= m.Waves; wave++ {
			all = append(all, Location{
				ID:      m.WaveLocationID(wave),
				Name:    m.WaveLocationName(wave),
				Kind:    ObjectiveWaveCleared,
				Mission: m.ID,
				Wave:    wave,
			})
		}
		if m.HasTank {
			all = append(all, Location{
				ID:      m.TankLocationID(),
				Name:    m.TankLocationName(),
				Kind:    ObjectiveTankDestroyed,
				Mission: m.ID,
			})
		}
		if m.HasGiant {
			all = append(all, Location{
				ID:      m.GiantLocationID(),
				Name:    m.GiantLocationName(),
				Kind:    ObjectiveGiantKilled,
				Mission: m.ID,
			})
		}
		all = append(all, Location{
			ID:      m.ClearLocationID(),
			Name:    m.ClearLocationName(),
			Kind:    ObjectiveMissionCleared,
			Mission: m.ID,
		})
		all = append(all, m.VictoryCacheLocations()...)
		all = append(all, m.WaveKillLocations()...)
	}
	return append(all, Milestones...)
}

// LocationByWaveKill resolves the nth giant or tank of a wave, which is what
// the plugin reports beside the mission's own first-of-each check. Nothing for
// a mission whose file was not scrubbed, or a kill past what the wave holds:
// the plugin sends what it sees, and the tables decide what is a check.
func LocationByWaveKill(kind ObjectiveKind, popFile string, wave, index uint8) (Location, bool) {
	m, ok := MissionByPopFile(popFile)
	if !ok || index == 0 {
		return Location{}, false
	}
	counts := m.WaveKillsAt(wave)
	switch kind {
	case ObjectiveGiantKilled:
		if index > counts.Giants {
			return Location{}, false
		}
		return Location{
			ID: m.WaveGiantLocationID(wave, index), Name: m.WaveGiantLocationName(wave, index),
			Kind: kind, Mission: m.ID, Wave: wave, Index: index,
		}, true
	case ObjectiveTankDestroyed:
		if index > counts.Tanks {
			return Location{}, false
		}
		return Location{
			ID: m.WaveTankLocationID(wave, index), Name: m.WaveTankLocationName(wave, index),
			Kind: kind, Mission: m.ID, Wave: wave, Index: index,
		}, true
	default:
		return Location{}, false
	}
}

var locationsByID = indexLocations()

func indexLocations() map[int64]Location {
	byID := make(map[int64]Location, len(Locations))
	for _, l := range Locations {
		byID[l.ID] = l
	}
	return byID
}

// LocationByID returns the mission and wave an id came from.
func LocationByID(id int64) (Location, bool) {
	l, ok := locationsByID[id]
	return l, ok
}

// LocationByObjective resolves what the plugin reported. Wave is ignored for a
// mission clear. It is the whole southbound translation: the bridge holds no
// id table of its own.
func LocationByObjective(kind ObjectiveKind, popFile string, wave uint8) (Location, bool) {
	m, ok := MissionByPopFile(popFile)
	if !ok {
		return Location{}, false
	}
	switch kind {
	case ObjectiveMissionCleared:
		return Location{
			ID:      m.ClearLocationID(),
			Name:    m.ClearLocationName(),
			Kind:    kind,
			Mission: m.ID,
		}, true
	case ObjectiveTankDestroyed:
		// A tank the tables do not know about is a report with no check behind
		// it. The plugin sends what it sees, and this drops what a seed cannot
		// hold, the same as it does for a wave out of range.
		if !m.HasTank {
			return Location{}, false
		}
		return Location{
			ID:      m.TankLocationID(),
			Name:    m.TankLocationName(),
			Kind:    kind,
			Mission: m.ID,
		}, true
	case ObjectiveGiantKilled:
		if !m.HasGiant {
			return Location{}, false
		}
		return Location{
			ID:      m.GiantLocationID(),
			Name:    m.GiantLocationName(),
			Kind:    kind,
			Mission: m.ID,
		}, true
	case ObjectiveWaveCleared:
		if wave < 1 || wave > m.Waves {
			return Location{}, false
		}
		return Location{
			ID:      m.WaveLocationID(wave),
			Name:    m.WaveLocationName(wave),
			Kind:    kind,
			Mission: m.ID,
			Wave:    wave,
		}, true
	default:
		return Location{}, false
	}
}
