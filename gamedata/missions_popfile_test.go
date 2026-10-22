package gamedata

import (
	"slices"
	"testing"
)

/*
A mission named in the first ten seconds of a map is refused under SigMod.

Pop:PopMgr_Extensions detours CPopulationManager::FindPopulationFileByShortName
and, until ten seconds of game time have passed since its last mission vote,
accepts only a name the map's loose population files list. A map starts with
that vote time at zero, so the first ten seconds of every map count as a vote.
Valve's own missions live in tf2_misc_dir.vpk, which that listing never reads,
so "normal" for mvm_decoy came back as "Could not find a valid population file
matching: normal." two seconds after the map loaded, and the map stayed on the
first community file it found, at wave 0 (apw-o96).

The guard is simulated here: a command sent while game time is under ten is
refused. The plugin has to send it after that, and not wait without SigMod.
*/
func TestThePopfileCommandWaitsOutSigModsVoteGuard(t *testing.T) {
	got := driver{body: `
    float starts[4] = {0.0, 2.0, 9.9, 10.0};
    for (int index = 0; index < 4; index++)
    {
        float now = starts[index];
        float sent = now + Missions_PopFileWait(now, true);
        printnum(sent >= 10.0 && sent < 11.0);
        printnum(Missions_PopFileWait(now, false) == 0.0);
    }
    printnum(Missions_PopFileWait(30.0, true) == 0.0);
`}.run(t)
	want := []int32{1, 1, 1, 1, 1, 1, 1, 1, 1}
	if !slices.Equal(got, want) {
		t.Fatalf("popfile sent outside the guard = %v, want %v", got, want)
	}
}
