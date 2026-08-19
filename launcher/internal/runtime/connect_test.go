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
func TestSteamConnectURL(t *testing.T) {
	s := settings.Settings{SrcdsPort: 27015}
	if got := SteamConnectURL(s); got != "steam://connect/127.0.0.1:27015" {
		t.Errorf("with no password = %q", got)
	}

	s.SrcdsPw = "friends only/2"
	if got := SteamConnectURL(s); got != "steam://connect/127.0.0.1:27015/friends%20only%2F2" {
		t.Errorf("with a password = %q", got)
	}
}
