package runshape

import (
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/gamedata"
)

func valveMissionsAtOrAbove(floor gamedata.Difficulty) int {
	count := 0
	for _, mission := range gamedata.Missions {
		if !gamedata.IsCommunityMission(mission.ID) && mission.Difficulty >= floor {
			count++
		}
	}
	return count
}

/*
apw-2kw: the ceiling counted every mission in the catalogue, the generator drew
from the ones the settings actually left in, and nothing reconciled the two.
A player asked for 82 at the normal floor and their run held 29.
*/
func TestTheCeilingCountsOnlyWhatTheGeneratorWillDraw(t *testing.T) {
	valveOnly := Pool{Community: false}
	want := valveMissionsAtOrAbove(gamedata.DifficultyNormal)
	if got := MissionsInPool(valveOnly, "normal"); got != want {
		t.Fatalf("normal ceiling with community missions off = %d, want %d", got, want)
	}

	withCommunity := Pool{Community: true}
	if MissionsInPool(withCommunity, "normal") <= want {
		t.Fatal("turning community missions on did not widen the pool")
	}
}

// A mission left out on the Missions page is one the generator never sees, so
// it cannot be one the launcher counts towards the number a run may ask for.
func TestExcludedMissionsLeaveTheCeiling(t *testing.T) {
	before := MissionsInPool(Pool{Community: true}, "advanced")
	excluded := Pool{Community: true, Excluded: []string{"mvm_mannworks_advanced"}}
	if got := MissionsInPool(excluded, "advanced"); got != before-1 {
		t.Fatalf("advanced ceiling after one exclusion = %d, want %d", got, before-1)
	}
}

// The label is the sentence a player reads before typing a mission count, so
// it has to describe the same pool the count is capped against.
func TestTierLabelsDescribeThePoolTheCeilingCaps(t *testing.T) {
	pool := Pool{Community: true, Excluded: []string{"mvm_decoy", "mvm_mannworks_advanced"}}
	for _, tier := range Tiers(pool) {
		if tier.Missions != MissionsInPool(pool, tier.Key) {
			t.Fatalf("%s: label says %d missions, ceiling says %d",
				tier.Key, tier.Missions, MissionsInPool(pool, tier.Key))
		}
	}
}

// The generator takes the smaller of the ask and the pool and says so only in
// its own log. A settings file written before the ceiling was fixed still
// holds the unreachable number, so the check has to name both.
func TestSummarySaysTheRunIsSmallerThanTheAsk(t *testing.T) {
	report, err := CheckSelection(Selection{
		Pool:         Pool{},
		Difficulty:   "normal",
		MissionCount: 82,
	})
	if err != nil {
		t.Fatal(err)
	}
	eligible := valveMissionsAtOrAbove(gamedata.DifficultyNormal)
	if report.Requested != 82 || report.Drawn != eligible {
		t.Fatalf("report = %+v, want 82 requested and %d drawn", report, eligible)
	}
	if !strings.Contains(report.Summary(), "asks for 82") {
		t.Fatalf("summary said nothing about the ask: %s", report.Summary())
	}
}
