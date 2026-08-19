package settings

import "strings"

// Reach is how players get to the game server. Three things move together
// here, which is why this is one value and not three checkboxes: whether the
// server logs in to Steam, whether anybody outside the local network can join
// at all, and whether joining needs a port open on the router.
//
// A server is not reachable from the internet by default. It takes either a
// forwarded port or Steam's relay network, and both of those need a login
// token, so the choice is one answer rather than a box to leave ticked.
type Reach string

const (
	// ReachLan keeps the server on the local network. It never logs in to
	// Steam, so it needs no login token and it never appears on the public
	// list. Nothing outside the network can join, whatever the router does.
	ReachLan Reach = "lan"

	// ReachSteam carries the traffic over Steam Datagram Relay. The server
	// asks Valve for an address in 169.254.0.0/16 and hands that to players,
	// so no port has to be forwarded and the real address stays hidden. The
	// address is new on every start.
	ReachSteam Reach = "steam"

	// ReachPort takes connections straight on the game port, which is the
	// classic setup: forward the port on the router and give out the public
	// address.
	ReachPort Reach = "port"
)

// Reaches lists the choices, private first.
func Reaches() []Reach { return []Reach{ReachLan, ReachSteam, ReachPort} }

// Valid reports whether r is one of the three.
func (r Reach) Valid() bool {
	switch r {
	case ReachLan, ReachSteam, ReachPort:
		return true
	}
	return false
}

// Lan reports whether sv_lan is on, which is what keeps the server off Steam
// and off the internet.
//
// An unrecognized value counts as LAN. A config file from the future, or one
// somebody hand-edited, must not be what opens a server to the internet.
func (r Reach) Lan() bool { return r != ReachSteam && r != ReachPort }

// NeedsToken reports whether the server logs in to Steam, which needs a Game
// Server Login Token.
func (r Reach) NeedsToken() bool { return !r.Lan() }

// SteamNetworking reports whether srcds asks Valve for a relayed address.
func (r Reach) SteamNetworking() bool { return r == ReachSteam }

// Label names the choice in one line, for a menu or a radio button.
func (r Reach) Label() string {
	switch r {
	case ReachSteam:
		return "Over Steam, with no port to open"
	case ReachPort:
		return "Over a port forwarded on the router"
	default:
		return "This machine and the local network only"
	}
}

// Help is the sentence under the label: what the choice costs and what it
// gives.
func (r Reach) Help() string {
	switch r {
	case ReachSteam:
		return "Steam relays the traffic. The server asks Valve for an address like " +
			"169.254.13.42:20232 and your friends connect to that from anywhere. " +
			"Nothing to forward, and your own address stays hidden. Needs a login " +
			"token, and the address is a new one every time the server starts."
	case ReachPort:
		return "Your friends connect to your public address, so the game port has to be " +
			"forwarded to this machine on your router and open in its firewall. Needs a " +
			"login token."
	default:
		return "Nobody outside your network can join, whatever the router does. No Steam " +
			"login, no token, never on the public list. This is the default."
	}
}

// HasToken reports whether a real Game Server Login Token was given. The
// setting carries "0" for none, which is the spelling the compose file and the
// documentation have always used, and an empty string means the same thing.
func HasToken(token string) bool {
	token = strings.TrimSpace(token)
	return token != "" && token != "0"
}

// Effective is the reach a server with this token can actually take. Every
// reach that leaves the local network logs in to Steam, and a server with no
// token never gets a session: it then refuses every client that tries to join,
// the ones on the local network included. Serving the local network is worth
// more than serving nobody, so a reach with no token behind it stays there.
func Effective(reach Reach, token string) Reach {
	if reach.NeedsToken() && !HasToken(token) {
		return ReachLan
	}
	return reach
}

// ParseReach reads the value an env var or a config file carries. Anything it
// does not recognize gives ReachLan and false, so a typo keeps the server
// private rather than guessing.
func ParseReach(value string) (Reach, bool) {
	reach := Reach(value)
	if !reach.Valid() {
		return ReachLan, false
	}
	return reach, true
}
