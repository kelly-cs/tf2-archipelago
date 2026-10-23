package gamedata

import "fmt"

/*
A milestone is a check for a running total across the whole run: robots
destroyed, giants killed, tanks destroyed, whatever mission they fell in. It
is progressed by playing anything the run already has open, so it fills a run
short on checks without adding a mission, and it gives a team stuck on a wave
something to grind (gh-82).

The plugin counts during a wave and reports one tally per counter when the
wave ends, won or lost; the bridge keeps the totals on disk and records every
milestone a total crosses. The thresholds are fixed so the ids are: a seed
that turns the option off simply never holds them.
*/

// MilestoneKind is one running total.
type MilestoneKind uint8

const (
	MilestoneRobots MilestoneKind = iota + 1
	MilestoneGiants
	MilestoneTanks
)

var milestoneKeys = [...]string{
	MilestoneRobots: "robots",
	MilestoneGiants: "giants",
	MilestoneTanks:  "tanks",
}

var milestoneNouns = [...]string{
	MilestoneRobots: "Robots Destroyed",
	MilestoneGiants: "Giants Killed",
	MilestoneTanks:  "Tanks Destroyed",
}

// MilestoneKinds is every total that is counted, in id order.
var MilestoneKinds = []MilestoneKind{MilestoneRobots, MilestoneGiants, MilestoneTanks}

// Key is the word for this total in the export and in the state file.
func (k MilestoneKind) Key() string { return milestoneKeys[k] }

// Tally is the objective kind the plugin reports this total's increments as.
func (k MilestoneKind) Tally() ObjectiveKind {
	switch k {
	case MilestoneRobots:
		return ObjectiveTallyRobots
	case MilestoneGiants:
		return ObjectiveTallyGiants
	default:
		return ObjectiveTallyTanks
	}
}

// milestoneThresholds is the ladder for each total. Append only: the index in
// the ladder is the id.
var milestoneThresholds = map[MilestoneKind][]int{
	MilestoneRobots: {100, 250, 500, 1000, 2000, 4000},
	MilestoneGiants: {10, 25, 50, 100, 200},
	MilestoneTanks:  {5, 10, 20, 40},
}

// Milestones is every milestone location, kind by kind and threshold by
// threshold. It is a slice of Location so the bridge resolves them the way it
// resolves any other check.
var Milestones = buildMilestones()

func buildMilestones() []Location {
	all := make([]Location, 0, 16)
	for _, kind := range MilestoneKinds {
		for index, threshold := range milestoneThresholds[kind] {
			all = append(all, Location{
				ID:        milestoneID(kind, uint8(index+1)),
				Name:      fmt.Sprintf("%d %s", threshold, milestoneNouns[kind]),
				Kind:      kind.Tally(),
				Threshold: threshold,
			})
		}
	}
	return all
}

// Counter is the total a tally feeds. Only meaningful for a tally kind.
func (k ObjectiveKind) Counter() MilestoneKind {
	switch k {
	case ObjectiveTallyRobots:
		return MilestoneRobots
	case ObjectiveTallyGiants:
		return MilestoneGiants
	default:
		return MilestoneTanks
	}
}

// MilestonesReached is every milestone of a total that a count at or past its
// threshold pays, in ladder order.
func MilestonesReached(kind ObjectiveKind, total int) []Location {
	var reached []Location
	for _, milestone := range Milestones {
		if milestone.Kind == kind && total >= milestone.Threshold {
			reached = append(reached, milestone)
		}
	}
	return reached
}
