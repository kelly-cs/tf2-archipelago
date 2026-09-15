package gamedata

// ObjectiveKind is the plugin's vocabulary for what Archipelago calls a
// location. The plugin reports in MvM terms and never learns of an id.
type ObjectiveKind uint8

const (
	ObjectiveWaveCleared ObjectiveKind = iota + 1
	ObjectiveMissionCleared
	ObjectiveTankDestroyed
	ObjectiveGiantKilled
)

var objectiveKeys = [...]string{
	ObjectiveWaveCleared:    "wave_cleared",
	ObjectiveMissionCleared: "mission_cleared",
	ObjectiveTankDestroyed:  "tank_destroyed",
	ObjectiveGiantKilled:    "giant_killed",
}

// ObjectiveKinds is every kind that exists, in id order. Whatever walks it
// covers new kinds without a second list to keep in step.
var ObjectiveKinds = []ObjectiveKind{
	ObjectiveWaveCleared, ObjectiveMissionCleared,
	ObjectiveTankDestroyed, ObjectiveGiantKilled,
}

// Key is the string on the wire between the plugin and the bridge.
func (k ObjectiveKind) Key() string { return objectiveKeys[k] }

// Location is one check. Wave is zero for a mission clear. Cache is zero for
// every check but a victory cache, where it counts from 1.
type Location struct {
	ID      int64
	Name    string
	Kind    ObjectiveKind
	Mission MissionID
	Wave    uint8
	Cache   uint8
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

// Locations is every check in the game, mission by mission: the waves in
// order, then the tank and the giant if the mission holds them, then the
// mission clear, then the victory caches its tier pays. The apworld includes
// the caches only when the option asks for them.
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
	}
	return all
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
