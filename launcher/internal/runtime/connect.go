package runtime

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// ConnectLines are what a player types in the game's developer console to join
// this server: one line for the machine running it, one for every address the
// friends on the network can reach, and the password when there is one.
//
// The public address is left out on purpose. Nothing here can see what a
// router does with the port, and printing a guess would send people to an
// address that does not answer. The relayed address is left out for the same
// reason: it does not exist yet at this point. FakeIPAddress picks it out of
// the server's own output once Valve has handed one over.
func ConnectLines(s settings.Settings) []string {
	port := strconv.Itoa(s.SrcdsPort)
	lines := []string{"connect " + net.JoinHostPort("127.0.0.1", port) + "   (on this machine)"}
	for _, address := range LocalAddresses() {
		lines = append(lines, "connect "+net.JoinHostPort(address, port)+"   (from your network)")
	}
	switch s.SrcdsReach {
	case settings.ReachLan:
		// Nothing to add: the lines above are already the whole answer.
	case settings.ReachSteam:
		lines = append(lines, "the address for friends elsewhere follows, once Steam hands one over")
	case settings.ReachPort:
		lines = append(lines, fmt.Sprintf("friends elsewhere need your public address, with port %s forwarded to this machine", port))
	}
	if s.SrcdsPw != "" {
		lines = append(lines, fmt.Sprintf("password %s   (before connect, the server asks for it)", s.SrcdsPw))
	}
	return lines
}

// fakeIPPrefix is what srcds prints when Valve has handed it a relayed
// address. The rest of the line is "169.254.13.42:20232, 20233": the first
// port carries the game and the second is the query port, which nobody types.
const fakeIPPrefix = "FakeIP allocation succeeded:"

// FakeIPAddress picks the relayed address out of one line of server output, or
// returns "" for every other line. Over Steam this address is the only way in,
// it is new on every start, and it exists nowhere else: reading it back out of
// the log is how the launcher learns it.
func FakeIPAddress(line string) string {
	_, rest, found := strings.Cut(line, fakeIPPrefix)
	if !found {
		return ""
	}
	address, _, _ := strings.Cut(rest, ",")
	address = strings.TrimSpace(address)
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return ""
	}
	// Valve allocates out of 169.254.0.0/16. Anything else on this line is a
	// message that changed shape, and a wrong address is worse than none.
	if ip := net.ParseIP(host); ip == nil || !ip.IsLinkLocalUnicast() {
		return ""
	}
	if _, err := strconv.Atoi(port); err != nil {
		return ""
	}
	return address
}

// LocalAddresses lists this machine's IPv4 addresses, skipping loopback and
// the ones Windows hands out when a network is not really there.
func LocalAddresses() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var found []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			ipNet, ok := address.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			found = append(found, ip.String())
		}
	}
	return found
}
