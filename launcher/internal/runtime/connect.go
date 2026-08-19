package runtime

import (
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// ConnectLines are what a player types in the game's developer console to join
// this server: one line for the machine running it, one for every address the
// friends on the network can reach, and the password when there is one.
//
// The public address is left out on purpose. Nothing here can see what a
// router does with the port, and printing a guess would send people to an
// address that does not answer. Over Steam the address is Valve's to hand
// out, and the server prints it once it is up; see FakeIP.
func ConnectLines(s settings.Settings) []string {
	port := strconv.Itoa(s.SrcdsPort)
	lines := []string{"connect " + net.JoinHostPort("127.0.0.1", port) + "   (on this machine)"}
	for _, address := range LocalAddresses() {
		lines = append(lines, "connect "+net.JoinHostPort(address, port)+"   (from your network)")
	}
	switch s.Reach() {
	case settings.ReachSteam:
		lines = append(lines, "over Steam: the address is in the status bar once the server has one")
	case settings.ReachPort:
		lines = append(lines, "from the internet: your public address, port "+port+" forwarded to this machine")
	case settings.ReachLan:
	}
	if s.SrcdsPw != "" {
		lines = append(lines, fmt.Sprintf("password %s   (before connect, the server asks for it)", s.SrcdsPw))
	}
	return lines
}

// fakeIPPattern is the address Steam's relay hands a server: always in
// 169.254.0.0/16, always with a port, whatever the line around it says. The
// wording of that line is the game's and has changed before; the address has
// not.
var fakeIPPattern = regexp.MustCompile(`\b169\.254\.\d{1,3}\.\d{1,3}:\d{2,5}\b`)

// FakeIP reads the relayed address out of a line of the game server's output.
func FakeIP(line string) (string, bool) {
	address := fakeIPPattern.FindString(line)
	return address, address != ""
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
