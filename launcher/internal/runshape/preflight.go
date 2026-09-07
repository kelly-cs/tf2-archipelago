package runshape

import (
	"fmt"
	"slices"

	"github.com/m-this/tf2-archipelago/gamedata"
)

// Selection is the part of the launcher settings that determines whether an
// Archipelago run has enough locations to hold its unlock items.
type Selection struct {
	Pool
	Difficulty   string
	MissionCount int
	StartMission string
}

// Preflight is the useful accounting behind a successful selection check.
// Eligible is the whole pool the generator may widen into. Requested is what
// the settings ask for and Drawn is what the run gets, which are two different
// numbers whenever the pool is smaller than the ask.
type Preflight struct {
	Eligible        int
	Requested       int
	Drawn           int
	AvailableChecks int
	RequiredUnlocks int
}

// Summary is suitable for a launcher status line after a successful check.
func (p Preflight) Summary() string {
	message := fmt.Sprintf(
		"Selection is valid: %d eligible mission(s) provide %d checks for %d unlocks.",
		p.Eligible, p.AvailableChecks, p.RequiredUnlocks,
	)
	switch {
	// The generator takes the smaller of the two and says so only in its own
	// log, which is not a place the player who typed the number ever looks.
	case p.Requested > p.Drawn:
		message += fmt.Sprintf(
			" The run asks for %d mission(s) and the pool holds %d, so it uses %d. Turn more missions on, load the mods they need, or lower the difficulty floor.",
			p.Requested, p.Eligible, p.Drawn,
		)
	case p.Requested < p.Eligible:
		message += fmt.Sprintf(
			" The generator requests %d and can add eligible missions automatically if its draw needs more checks.",
			p.Requested,
		)
	}
	return message
}

type missionRequirement struct {
	classes int
	slots   int
}

// These mirror apworld/tf2_mvm/rules.py. The official generator remains the
// final authority if its rules change ahead of the launcher.
var missionRequirements = map[gamedata.Difficulty]missionRequirement{
	gamedata.DifficultyNormal:       {classes: 1, slots: 1},
	gamedata.DifficultyIntermediate: {classes: 2, slots: 1},
	gamedata.DifficultyAdvanced:     {classes: 3, slots: 2},
	gamedata.DifficultyExpert:       {classes: 4, slots: 3},
	gamedata.DifficultyHaunted:      {classes: 5, slots: 3},
}

// CheckSelection performs the same pool-capacity check as the apworld's
// generate_early method. The apworld may add eligible missions beyond the
// requested count; generation fails only when the entire eligible pool still
// has fewer locations than unlock items.
func CheckSelection(selection Selection) (Preflight, error) {
	floor, known := gamedata.DifficultyByKey(selection.Difficulty)
	if !known {
		return Preflight{}, fmt.Errorf("unknown difficulty pool %q", selection.Difficulty)
	}
	if selection.MissionCount < 1 {
		return Preflight{}, fmt.Errorf("mission count must be at least 1")
	}

	var eligible []gamedata.Mission
	for _, mission := range selection.Missions() {
		if mission.Difficulty < floor {
			continue
		}
		eligible = append(eligible, mission)
	}
	if len(eligible) == 0 {
		return Preflight{}, fmt.Errorf(
			"no missions remain in the %s-or-harder pool; select a mission or lower the difficulty floor",
			floor.Key(),
		)
	}

	start := easiestMission(eligible)
	if selection.StartMission != "" {
		at := slices.IndexFunc(eligible, func(mission gamedata.Mission) bool {
			return mission.PopFile == selection.StartMission
		})
		if at < 0 {
			return Preflight{}, fmt.Errorf(
				"start mission %q is excluded or outside the %s-or-harder pool",
				selection.StartMission, floor.Key(),
			)
		}
		start = eligible[at]
	}

	checks := 0
	for _, mission := range eligible {
		checks += missionCheckCount(mission)
	}
	requirement := missionRequirements[start.Difficulty]
	unlocks := len(eligible) - 1 +
		len(gamedata.Classes) - requirement.classes +
		len(gamedata.WeaponSlots) - requirement.slots
	report := Preflight{
		Eligible:        len(eligible),
		Requested:       selection.MissionCount,
		Drawn:           min(selection.MissionCount, len(eligible)),
		AvailableChecks: checks,
		RequiredUnlocks: unlocks,
	}
	if checks < unlocks {
		return report, fmt.Errorf(
			"the %s-or-harder pool has %d mission(s) and %d checks, but needs room for %d unlocks; select another mission or lower the difficulty floor",
			floor.Key(), len(eligible), checks, unlocks,
		)
	}
	return report, nil
}

func easiestMission(missions []gamedata.Mission) gamedata.Mission {
	return slices.MinFunc(missions, func(a, b gamedata.Mission) int {
		if a.Difficulty != b.Difficulty {
			return int(a.Difficulty) - int(b.Difficulty)
		}
		return int(a.ID) - int(b.ID)
	})
}

func missionCheckCount(mission gamedata.Mission) int {
	checks := int(mission.Waves) + 1 // waves plus mission clear
	if mission.HasTank {
		checks++
	}
	if mission.HasGiant {
		checks++
	}
	return checks
}
