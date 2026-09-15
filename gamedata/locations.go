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

// Location is one check. Wave is zero for a mission clear. A milestone has
// no mission and a Threshold instead: the total its tally has to reach.
type Location struct {
	ID        int64
	Name      string
	Kind      ObjectiveKind
	Mission   MissionID
	Wave      uint8
	Threshold int
}

// Locations is every check in the game, mission by mission: the waves in
// order, then the tank and the giant if the mission holds them, then the
// mission clear; and after every mission, the milestones.
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
	}
	return append(all, Milestones...)
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
