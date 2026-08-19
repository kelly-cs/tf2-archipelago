package runtime

import (
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func TestFakeIPAddress(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
	}{
		{
			name: "the line srcds prints",
			line: "FakeIP allocation succeeded: 169.254.134.215:57528, 57529",
			want: "169.254.134.215:57528",
		},
		{
			name: "with the timestamp the console adds",
			line: "08/19 04:09:13 FakeIP allocation succeeded: 169.254.13.42:20232, 20233",
			want: "169.254.13.42:20232",
		},
		{
			name: "one port only",
			line: "FakeIP allocation succeeded: 169.254.13.42:20232",
			want: "169.254.13.42:20232",
		},
		{name: "another line entirely", line: "Connection to Steam servers successful.", want: ""},
		{name: "the request, not the answer", line: "Requesting FakeIP as per -enablefakeip", want: ""},
		// Anything outside 169.254.0.0/16 is not a relayed address. Printing a
		// real one here would hand out the address the relay exists to hide.
		{name: "a real address", line: "FakeIP allocation succeeded: 203.0.113.7:27015, 27016", want: ""},
		{name: "no port", line: "FakeIP allocation succeeded: 169.254.13.42, 20233", want: ""},
		{name: "not a port", line: "FakeIP allocation succeeded: 169.254.13.42:none, 20233", want: ""},
		{name: "nothing after it", line: "FakeIP allocation succeeded:", want: ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := FakeIPAddress(c.line); got != c.want {
				t.Errorf("FakeIPAddress(%q) = %q, want %q", c.line, got, c.want)
			}
		})
	}
}

func TestConnectLinesSayWhatEachReachNeeds(t *testing.T) {
	s := settings.Settings{SrcdsPort: 27015, SrcdsReach: settings.ReachLan}

	joined := strings.Join(ConnectLines(s), "\n")
	if !strings.Contains(joined, "connect 127.0.0.1:27015") {
		t.Error("no loopback line")
	}
	if strings.Contains(joined, "forwarded") || strings.Contains(joined, "Steam") {
		t.Errorf("a LAN server talked about getting out:\n%s", joined)
	}

	s.SrcdsReach = settings.ReachSteam
	if joined := strings.Join(ConnectLines(s), "\n"); !strings.Contains(joined, "Steam") {
		t.Errorf("nothing said about the address to come:\n%s", joined)
	}

	s.SrcdsReach = settings.ReachPort
	if joined := strings.Join(ConnectLines(s), "\n"); !strings.Contains(joined, "27015 forwarded") {
		t.Errorf("nothing said about forwarding the port:\n%s", joined)
	}
}

// The Join button hands this to Steam, which starts the game and connects.
// The password rides in the URL, so one that holds a space or a slash has to
// survive the trip.
//
// The link names 440. steam://connect asks the server which game it is, and
// ours answers with the dedicated server's own application, which the client
// refuses with "app id specified by server is invalid".
func TestSteamConnectURL(t *testing.T) {
	s := settings.Settings{SrcdsPort: 27015}

	// This machine's own address, not loopback: a connect to 127.0.0.1 times
	// out against a server on the same machine, and the LAN tab of the server
	// browser lists that server at its network address.
	host := "127.0.0.1"
	if local := LocalAddresses(); len(local) > 0 {
		host = local[0]
	}

	if got := SteamConnectURL(s); got != "steam://run/440//+connect%20"+host+":27015" {
		t.Errorf("with no password = %q", got)
	}

	s.SrcdsPw = "friends only/2"
	want := "steam://run/440//+connect%20" + host + ":27015%20+password%20friends%20only%2F2"
	if got := SteamConnectURL(s); got != want {
		t.Errorf("with a password = %q, want %q", got, want)
	}

	// A port that is not the default has to reach the URL, or the button joins
	// a server that is not the one running.
	s.SrcdsPw, s.SrcdsPort = "", 27045
	if got := SteamConnectURL(s); !strings.Contains(got, host+":27045") {
		t.Errorf("the port did not reach the link: %q", got)
	}
}

// The run moves itself from mission to mission, and the plugin's log line is
// the only place that says which one is on. The window's header showed the
// mission the settings named instead, which stayed on the first one all
// evening.
func TestLoadedMission(t *testing.T) {
	line := `L 08/19/2026 - 17:28:26: [tf2_archipelago.smx] mission switched to Doe's Drill (mvm_decoy) on mvm_decoy`
	if got := LoadedMission(line); got != "Doe's Drill" {
		t.Errorf("LoadedMission = %q, want %q", got, "Doe's Drill")
	}
	// A name with a space and a plus in it, which several missions have.
	line = `[tf2_archipelago.smx] mission switched to Ctrl+Alt+Destruction (mvm_coaltown_advanced) on mvm_coaltown`
	if got := LoadedMission(line); got != "Ctrl+Alt+Destruction" {
		t.Errorf("LoadedMission = %q", got)
	}
	for _, other := range []string{
		"-------- Mapchange to mvm_decoy --------",
		`[UPDATER] Successfully updated gamedata file "sm-tf2.games.txt"`,
		"mission switched to",
		"",
	} {
		if got := LoadedMission(other); got != "" {
			t.Errorf("read a mission out of %q: %q", other, got)
		}
	}
}

// The updater writes its gamedata and then asks whoever is watching the log to
// restart. Reading that line is what lets the launcher do it instead.
func TestSourceModWasUpdated(t *testing.T) {
	for _, line := range []string{
		"L 08/19/2026 - 17:28:35: [UPDATER] SourceMod has been updated, please reload it or restart your server.",
		"[UPDATER] SourceMod has been updated",
	} {
		if !SourceModWasUpdated(line) {
			t.Errorf("missed the updater asking for a restart: %q", line)
		}
	}
	// The same updater says this a few hundred times per start, and none of
	// them is a reason to bring the server round.
	for _, line := range []string{
		`L 08/19/2026 - 17:28:34: [UPDATER] Successfully updated gamedata file "sm-tf2.games.txt"`,
		"L 08/19/2026 - 17:28:26: [tf2_archipelago.smx] mission switched to Doe's Drill",
		"",
	} {
		if SourceModWasUpdated(line) {
			t.Errorf("restarted for an ordinary line: %q", line)
		}
	}
}
