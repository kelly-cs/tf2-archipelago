// Command rcon drives the game server from a shell or from CI, without a game
// client, through the same client the launcher's command box uses. It reads
// SRCDS_RCONPW, and SRCDS_RCON_HOST and SRCDS_PORT for where.
//
// Usage: SRCDS_RCONPW=... go run ./launcher/cmd/rcon <command> [command ...]
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/m-this/tf2-archipelago/launcher/internal/rcon"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(commands []string, stdout, stderr io.Writer) int {
	if len(commands) == 0 {
		say(stderr, "usage: rcon <command> [command ...]")
		return 2
	}
	password := os.Getenv("SRCDS_RCONPW")
	if password == "" {
		say(stderr, "set SRCDS_RCONPW, the same value the server booted with")
		return 2
	}
	// .env leaves a setting empty rather than unset, so empty means the default.
	host := os.Getenv("SRCDS_RCON_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("SRCDS_PORT")
	if port == "" {
		port = "27015"
	}

	client, err := rcon.DialContext(context.Background(), net.JoinHostPort(host, port), password)
	if err != nil {
		if errors.Is(err, rcon.ErrBadPassword) {
			say(stderr, "rcon: the server refused the rcon password. It reads SRCDS_RCONPW at boot, so a value changed since then needs 'make restart'.")
			return 1
		}
		say(stderr, "rcon: "+err.Error())
		return 1
	}
	defer func() { _ = client.Close() }()

	for _, command := range commands {
		reply, err := client.Exec(command)
		if err != nil {
			say(stderr, "rcon: "+err.Error())
			return 1
		}
		say(stdout, "$ "+command)
		if reply != "" {
			say(stdout, reply)
		}
	}
	return 0
}

func say(w io.Writer, line string) { _, _ = fmt.Fprintln(w, line) }
