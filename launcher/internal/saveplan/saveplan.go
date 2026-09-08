/*
Package saveplan answers one question: what does saving these settings ask of a
game server that is already running.

Three things can be true of a save, and the launcher owes the player a different
sentence for each.

  - The bot team moved. The mod re-reads its lineup from a convar and its
    weapons from a file, so botlive hands it over and the wave carries on.
  - Something the server reads once at startup moved: server.cfg, or the command
    line. Nothing short of a restart changes it.
  - Neither. The run shape is the whole of this: the mission count, the pool,
    the goal, the rewards. It goes into tf2.yaml and from there into the seed,
    and the running server never reads any of it.

The third used to be the second, so saving a mission count ended the mission the
team was four waves into and brought the server back exactly as it was. That is
apw-vip, and Likai reported it as the settings being too aggressive.
*/
package saveplan

import (
	"reflect"

	"github.com/m-this/tf2-archipelago/launcher/internal/botlive"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// Plan is what a save asks of a running server. The two are independent: a save
// that moves the team and the port does both, and one that moves the team and
// the mission count hands the team over without a restart.
type Plan struct {
	// Team is whether the bot team moved, so botlive.Commands has something to
	// send.
	Team bool

	// Restart is whether anything that is read once at startup moved.
	Restart bool
}

// Quiet is a save that asks the running server for nothing: either nothing
// moved, or the only thing that moved is a setting the server never reads.
func (p Plan) Quiet() bool { return !p.Team && !p.Restart }

// For reads the plan off the two settings.
func For(before, after settings.Settings) Plan {
	return Plan{
		Team:    botlive.TeamMoved(before, after),
		Restart: !reflect.DeepEqual(readAtStart(before), readAtStart(after)),
	}
}

/*
readAtStart is these settings with everything a running server never reads taken
out, so what is left is what a restart is the only way to change.

Written as a subtraction and not as a list of what needs a restart. A setting
added later needs one until somebody says otherwise, which is the safe way round
and is the same shape botlive uses for the team.

Test mode is the exception and it has to be: runtime.StartTestRoom builds the
one-player room out of the run shape when the server starts, so there the run
shape is a setting the server reads.
*/
func readAtStart(s settings.Settings) settings.Settings {
	s = botlive.WithoutTeam(s)
	if s.TestMode {
		return s
	}
	// The run shape. Every one of these reaches the game through tf2.yaml and
	// the seed generated from it, and through nothing else: no field here is
	// read by internal/srcdsconfig or by the command line.
	s.MvmMissionCount = 0
	s.MvmDifficulty = ""
	s.MvmGoal = ""
	s.MvmMissionsanityPct = 0
	s.MvmMedalOnClear = false
	s.MvmDeathLink = false
	s.MvmExcludedMissions = nil
	s.MvmStartMission = ""
	s.MvmStartClass = ""
	s.MvmCommunityMissions = false
	s.MvmCashRewards = false
	s.MvmTrapPct = 0
	s.MvmWeaponBuffPct = 0
	s.MvmWeaponBuffStackChance = 0
	s.MvmMissionTicketImportance = ""
	s.MvmClassUnlockImportance = ""
	s.MvmWeaponSlotImportance = ""
	s.MvmWeaponBuffImportance = ""
	return s
}
