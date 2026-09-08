package saveplan

import (
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func base() settings.Settings {
	return settings.Settings{
		SrcdsBotTeamSize:     6,
		SrcdsBotTeamComp:     []string{"engineer", "medic"},
		SrcdsBotSeatLoadouts: []string{"ranger", "kritz"},
		SrcdsHostname:        "a server",
		SrcdsPort:            27015,
		MvmMissionCount:      8,
		MvmDifficulty:        "intermediate",
		MvmGoal:              "final_boss",
	}
}

// clone keeps one case from writing into the next one's base.
func clone(s settings.Settings) settings.Settings {
	out := s
	out.SrcdsBotTeamComp = append([]string(nil), s.SrcdsBotTeamComp...)
	out.SrcdsBotSeatLoadouts = append([]string(nil), s.SrcdsBotSeatLoadouts...)
	out.SrcdsBotClassBlacklist = append([]string(nil), s.SrcdsBotClassBlacklist...)
	out.MvmExcludedMissions = append([]string(nil), s.MvmExcludedMissions...)
	return out
}

/*
	The bot team is the one change a running mission takes, and it is the only

one.

The mod re-reads its lineup from a convar and its weapons from a file, so a save
that moved only those hands the team over and the wave carries on. Anything else
on the same save is still a restart.

Written against real fields rather than a list of names, so a setting added to
the struct later fails here rather than silently going live.
*/
func TestOnlyTheBotTeamGoesLive(t *testing.T) {
	for _, test := range []struct {
		name  string
		after func(settings.Settings) settings.Settings
		want  Plan
	}{
		{"nothing moved", func(s settings.Settings) settings.Settings { return s }, Plan{}},
		{
			"a seat changed class",
			func(s settings.Settings) settings.Settings {
				s.SrcdsBotTeamComp = []string{"engineer", "heavyweapons"}
				return s
			},
			Plan{Team: true},
		},
		{
			"the team got bigger",
			func(s settings.Settings) settings.Settings { s.SrcdsBotTeamSize = 4; return s },
			Plan{Team: true},
		},
		{
			"a class was unticked",
			func(s settings.Settings) settings.Settings {
				s.SrcdsBotClassBlacklist = []string{"spy"}
				return s
			},
			Plan{Team: true},
		},
		{
			"a seat changed weapons",
			func(s settings.Settings) settings.Settings {
				s.SrcdsBotSeatLoadouts = []string{"widowmaker", "kritz"}
				return s
			},
			Plan{Team: true},
		},
		{
			// The saved teams are the launcher's own list and reach the server
			// through nothing at all, so naming one is neither a restart nor
			// anything to send.
			"a team was saved under a name",
			func(s settings.Settings) settings.Settings {
				s.SrcdsBotTeamPresets = map[string]settings.BotTeam{"tanks": {}}
				return s
			},
			Plan{},
		},
		{
			"the port moved, which the server reads once",
			func(s settings.Settings) settings.Settings { s.SrcdsPort = 27016; return s },
			Plan{Restart: true},
		},
		{
			"a seat changed and so did the port",
			func(s settings.Settings) settings.Settings {
				s.SrcdsBotTeamComp = []string{"scout"}
				s.SrcdsPort = 27016
				return s
			},
			Plan{Team: true, Restart: true},
		},
	} {
		if got := For(base(), test.after(clone(base()))); got != test.want {
			t.Errorf("%s: plan = %+v, want %+v", test.name, got, test.want)
		}
	}
}

/*
apw-vip: the run shape never reaches a running server, so saving one used to end
the mission the team was playing and bring the server back exactly as it was.

Every field here goes into tf2.yaml and from there into the seed. The run keeps
the shape it was generated with whatever this says.
*/
func TestTheRunShapeAsksTheRunningServerForNothing(t *testing.T) {
	for _, test := range []struct {
		name  string
		after func(settings.Settings) settings.Settings
	}{
		{"the mission count", func(s settings.Settings) settings.Settings { s.MvmMissionCount = 12; return s }},
		{"the difficulty floor", func(s settings.Settings) settings.Settings { s.MvmDifficulty = "advanced"; return s }},
		{"the goal", func(s settings.Settings) settings.Settings { s.MvmGoal = "missionsanity"; return s }},
		{"the pool", func(s settings.Settings) settings.Settings {
			s.MvmExcludedMissions = []string{"mvm_decoy"}
			return s
		}},
		{"a reward weighting", func(s settings.Settings) settings.Settings {
			s.MvmWeaponBuffImportance = "progression"
			return s
		}},
		{"the trap share", func(s settings.Settings) settings.Settings { s.MvmTrapPct = 20; return s }},
		{"death link", func(s settings.Settings) settings.Settings { s.MvmDeathLink = true; return s }},
	} {
		if got := For(base(), test.after(clone(base()))); got != (Plan{}) {
			t.Errorf("%s: plan = %+v, want a save the running server is owed nothing for", test.name, got)
		}
	}
}

// Test mode is the exception, and it has to be: runtime.StartTestRoom builds
// the one-player room out of the run shape when the server starts, so there the
// run shape is a setting the server reads.
func TestTestModeReadsTheRunShapeAtStart(t *testing.T) {
	before := base()
	before.TestMode = true
	after := clone(before)
	after.MvmMissionCount = 12

	if got := For(before, after); !got.Restart {
		t.Fatalf("plan = %+v, want a restart: the test room is built from this", got)
	}
}

// Turning test mode on is itself a change the server reads: it is a different
// room, and the bridge is pointed at it when the server starts.
func TestTurningTestModeOnNeedsARestart(t *testing.T) {
	after := clone(base())
	after.TestMode = true

	if got := For(base(), after); !got.Restart {
		t.Fatalf("plan = %+v, want a restart", got)
	}
}

// A save with the team and the run shape in it hands the team over and leaves
// the mission alone, which is the case the two halves being independent exists
// for.
func TestATeamChangeBesideARunShapeChangeStaysLive(t *testing.T) {
	after := clone(base())
	after.SrcdsBotTeamComp = []string{"scout"}
	after.MvmMissionCount = 12

	if got := For(base(), after); got != (Plan{Team: true}) {
		t.Fatalf("plan = %+v, want the team handed over and no restart", got)
	}
}
