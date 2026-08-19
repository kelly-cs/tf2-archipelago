package runtime

import (
	"slices"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func baseSettings() settings.Settings {
	return settings.Settings{
		SrcdsMaxPlayers: 32,
		SrcdsStartMap:   "mvm_decoy",
		SrcdsPort:       27015,
		SrcdsRconPw:     "rcon-secret",
		SrcdsToken:      "0",
		SrcdsReach:      settings.ReachLan,
	}
}

// value returns what follows flag on the command line, or "" when the flag is
// not there at all.
func value(args []string, flag string) string {
	i := slices.Index(args, flag)
	if i < 0 || i+1 >= len(args) {
		return ""
	}
	return args[i+1]
}

func TestLanArgsNeverLeaveTheNetwork(t *testing.T) {
	s := baseSettings()
	// A token left over from an evening spent online must not put a server
	// that is back on LAN mode into the public list.
	s.SrcdsToken = "C7A1B2E3D4F5A6B7C8D9E0F1A2B3C4D5"
	args := srcdsArgs(s, "srcds_run")

	if got := value(args, "+sv_lan"); got != "1" {
		t.Errorf("+sv_lan = %q, want 1", got)
	}
	if slices.Contains(args, "-enablefakeip") {
		t.Error("-enablefakeip on a LAN server")
	}
	if slices.Contains(args, "+sv_setsteamaccount") {
		t.Error("the token was passed to a server that never logs in")
	}
	if slices.Contains(args, "+sv_use_steam_networking") {
		t.Error("+sv_use_steam_networking on a LAN server")
	}
}

func TestSteamArgsAskForARelayedAddress(t *testing.T) {
	s := baseSettings()
	s.SrcdsReach = settings.ReachSteam
	s.SrcdsToken = "C7A1B2E3D4F5A6B7C8D9E0F1A2B3C4D5"
	args := srcdsArgs(s, "srcds_run")

	if !slices.Contains(args, "-enablefakeip") {
		t.Error("no -enablefakeip: Valve hands out no address without it")
	}
	if got := value(args, "+sv_use_steam_networking"); got != "1" {
		t.Errorf("+sv_use_steam_networking = %q, want 1", got)
	}
	if got := value(args, "+sv_lan"); got != "0" {
		t.Errorf("+sv_lan = %q, want 0", got)
	}
	if got := value(args, "+sv_setsteamaccount"); got != s.SrcdsToken {
		t.Errorf("+sv_setsteamaccount = %q, want the token", got)
	}
}

func TestPortArgsGoOnlineWithoutTheRelay(t *testing.T) {
	s := baseSettings()
	s.SrcdsReach = settings.ReachPort
	s.SrcdsToken = "C7A1B2E3D4F5A6B7C8D9E0F1A2B3C4D5"
	args := srcdsArgs(s, "srcds_run")

	if slices.Contains(args, "-enablefakeip") {
		t.Error("-enablefakeip on a forwarded-port server")
	}
	if got := value(args, "+sv_lan"); got != "0" {
		t.Errorf("+sv_lan = %q, want 0", got)
	}
	if got := value(args, "+sv_setsteamaccount"); got != s.SrcdsToken {
		t.Errorf("+sv_setsteamaccount = %q, want the token", got)
	}
}

// "0" is how the compose file and the docs spell "no token". Passing it to
// sv_setsteamaccount is worse than passing nothing.
func TestNoTokenIsNotPassed(t *testing.T) {
	s := baseSettings()
	s.SrcdsReach = settings.ReachSteam
	s.SrcdsToken = "0"
	if args := srcdsArgs(s, "srcds_run"); slices.Contains(args, "+sv_setsteamaccount") {
		t.Error("+sv_setsteamaccount 0 was passed")
	}
}

// srcds reads the dash flags as a set and the + commands in order, so the
// commands must not start before the flags are done.
func TestFlagsComeBeforeCommands(t *testing.T) {
	s := baseSettings()
	s.SrcdsReach = settings.ReachSteam
	args := srcdsArgs(s, "srcds.exe")

	firstCommand := len(args)
	for i, arg := range args {
		if strings.HasPrefix(arg, "+") {
			firstCommand = i
			break
		}
	}
	for i, arg := range args[firstCommand:] {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "-1") {
			t.Errorf("flag %q at %d, after the commands started", arg, firstCommand+i)
		}
	}
	if !slices.Contains(args[:firstCommand], "-console") {
		t.Error("srcds.exe without -console waits for a click nobody makes")
	}
}
