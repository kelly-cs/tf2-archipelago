package settings

import "testing"

// An unrecognized reach must behave like the private one. A config file from a
// newer build, or one somebody hand-edited, is not allowed to be what opens a
// server to the internet.
func TestUnknownReachStaysPrivate(t *testing.T) {
	unknown := Reach("whatever-comes-next")
	if unknown.Valid() {
		t.Fatal("Valid() = true")
	}
	if !unknown.Lan() {
		t.Error("Lan() = false, want true")
	}
	if unknown.NeedsToken() {
		t.Error("NeedsToken() = true, want false")
	}
	if unknown.SteamNetworking() {
		t.Error("SteamNetworking() = true, want false")
	}
}

func TestReachAnswers(t *testing.T) {
	cases := []struct {
		reach           Reach
		lan             bool
		needsToken      bool
		steamNetworking bool
	}{
		{ReachLan, true, false, false},
		{ReachSteam, false, true, true},
		{ReachPort, false, true, false},
	}
	for _, c := range cases {
		t.Run(string(c.reach), func(t *testing.T) {
			if got := c.reach.Lan(); got != c.lan {
				t.Errorf("Lan() = %v, want %v", got, c.lan)
			}
			if got := c.reach.NeedsToken(); got != c.needsToken {
				t.Errorf("NeedsToken() = %v, want %v", got, c.needsToken)
			}
			if got := c.reach.SteamNetworking(); got != c.steamNetworking {
				t.Errorf("SteamNetworking() = %v, want %v", got, c.steamNetworking)
			}
			if !c.reach.Valid() {
				t.Error("Valid() = false")
			}
			if c.reach.Label() == "" || c.reach.Help() == "" {
				t.Error("a reach with no words on it")
			}
		})
	}
}

func TestParseReach(t *testing.T) {
	cases := []struct {
		value string
		want  Reach
		ok    bool
	}{
		{"lan", ReachLan, true},
		{"steam", ReachSteam, true},
		{"port", ReachPort, true},
		{"", ReachLan, false},
		{"public", ReachLan, false},
		{"LAN", ReachLan, false},
	}
	for _, c := range cases {
		got, ok := ParseReach(c.value)
		if got != c.want || ok != c.ok {
			t.Errorf("ParseReach(%q) = %q, %v; want %q, %v", c.value, got, ok, c.want, c.ok)
		}
	}
}

func TestHasToken(t *testing.T) {
	cases := []struct {
		token string
		want  bool
	}{
		{"", false},
		{"0", false},
		{"  0  ", false},
		{"C7A1B2E3D4F5A6B7C8D9E0F1A2B3C4D5", true},
	}
	for _, c := range cases {
		if got := HasToken(c.token); got != c.want {
			t.Errorf("HasToken(%q) = %v, want %v", c.token, got, c.want)
		}
	}
}

// A server that cannot log in refuses every player, so a reach that needs a
// token and has none is worth less than the local network it came from.
func TestEffectiveReachNeedsItsToken(t *testing.T) {
	const token = "C7A1B2E3D4F5A6B7C8D9E0F1A2B3C4D5"
	cases := []struct {
		reach Reach
		token string
		want  Reach
	}{
		{ReachPort, "", ReachLan},
		{ReachPort, "0", ReachLan},
		{ReachPort, token, ReachPort},
		{ReachSteam, "0", ReachLan},
		{ReachSteam, token, ReachSteam},
		{ReachLan, "0", ReachLan},
		{ReachLan, token, ReachLan},
	}
	for _, c := range cases {
		if got := Effective(c.reach, c.token); got != c.want {
			t.Errorf("Effective(%q, %q) = %q, want %q", c.reach, c.token, got, c.want)
		}
	}
}
