package runtime

import (
	"slices"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func TestSrcdsArgsFollowTheReach(t *testing.T) {
	base := settings.Defaults()
	base.SrcdsRconPw = "x"
	base.SrcdsToken = "ABCDEF"
	base.SrcdsStartMission = "mvm_ghost_town_666"

	lan := base.WithReach(settings.ReachLan)
	steam := base.WithReach(settings.ReachSteam)
	port := base.WithReach(settings.ReachPort)

	for _, tc := range []struct {
		name string
		s    settings.Settings
		want []string
		none []string
	}{
		{"lan", lan, []string{"+sv_lan 1"}, []string{"sv_setsteamaccount", "-enablefakeip"}},
		{"steam", steam, []string{"+sv_lan 0", "-enablefakeip", "+sv_setsteamaccount ABCDEF"}, nil},
		{"port", port, []string{"+sv_lan 0", "+sv_setsteamaccount ABCDEF"}, []string{"-enablefakeip"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			line := strings.Join(srcdsArgs(tc.s, false), " ")
			for _, want := range tc.want {
				if !strings.Contains(line, want) {
					t.Errorf("missing %q in %q", want, line)
				}
			}
			for _, none := range tc.none {
				if strings.Contains(line, none) {
					t.Errorf("unexpected %q in %q", none, line)
				}
			}
			// The map is the mission's, and Ghost Town is the case a popfile
			// name gets wrong.
			if !strings.Contains(line, "+map mvm_ghost_town ") {
				t.Errorf("the map is not the start mission's: %q", line)
			}
		})
	}
	if args := srcdsArgs(lan, true); !slices.Contains(args, "-console") {
		t.Errorf("srcds.exe runs without -console: %v", args)
	}
}

func TestStartMapTakesAMapNameFromAnOlderSetting(t *testing.T) {
	s := settings.Settings{SrcdsStartMission: "mvm_coaltown"}
	if got := StartMap(s); got != "mvm_coaltown" {
		t.Errorf("StartMap = %q", got)
	}
	s.SrcdsStartMission = "mvm_decoy_advanced2"
	if got := StartMap(s); got != "mvm_decoy" {
		t.Errorf("StartMap = %q", got)
	}
}

func TestFakeIPReadsTheRelayedAddress(t *testing.T) {
	for _, tc := range []struct {
		line string
		want string
	}{
		{"SDR: fake IP 169.254.13.42:20232 assigned", "169.254.13.42:20232"},
		{"udp/ip  : 192.168.1.10:27015  (public IP: 169.254.1.2:3)", ""},
		{"Connection to Steam servers successful.", ""},
	} {
		got, ok := FakeIP(tc.line)
		if ok != (tc.want != "") || got != tc.want {
			t.Errorf("FakeIP(%q) = %q, %v", tc.line, got, ok)
		}
	}
}
