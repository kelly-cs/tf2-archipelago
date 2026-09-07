/*
Package roomcheck asks whether an Archipelago room is answering.

Saving the settings is the moment a player finds out whether the address they
pasted works, and until now they found out much later: Start refused with
"AP_PORT is not set", or the bridge retried a dead room in a log nobody was
reading. One player spent seventeen minutes on that and gave up.

So Save asks the room. It is one websocket handshake, and it is deliberately
only that: a room that completes a handshake is listening and speaking the right
protocol, which is the whole question at this point. Whether the slot name is
right, whether the seed matches, whether the room will still be up in an hour
are all things the bridge finds out when it connects for real, and none of them
can be answered here without a session nobody asked for.

Nothing here refuses a save. A player configuring the launcher on a train, or
before making the room, has settings worth keeping and a room that cannot
answer; refusing them would be the same bug this was written to fix, wearing a
different hat.
*/
package roomcheck

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/coder/websocket"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// Timeout is how long a room is given to answer. Long enough for a handshake
// across an ocean, short enough that Save does not feel stuck: a live room on
// archipelago.gg answers in well under a second.
const Timeout = 3 * time.Second

// Result is what asking found.
type Result int

const (
	// Live is a room that completed the handshake.
	Live Result = iota
	// NotConfigured is no room at all: the settings have no address, which is
	// what a fresh install has.
	NotConfigured
	// Unreachable is an address that is there and did not answer. The room may
	// be closed, the seed may not have been uploaded yet, or the machine may
	// have no network.
	Unreachable
	// Skipped is test mode, which never dials a real room.
	Skipped
)

/*
	Check asks the room and says what happened, never why not to save.

The error is kept beside the result rather than returned on its own, because a
room that did not answer is not a failure of this function: it did its job and
the answer was no. The caller shows the reason and carries on.
*/
func Check(ctx context.Context, s settings.Settings) (Result, error) {
	if s.TestMode {
		return Skipped, nil
	}
	if s.APHost == "" || s.APPort == 0 {
		return NotConfigured, nil
	}

	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	conn, handshake, err := websocket.Dial(ctx, roomURL(s), nil)
	if handshake != nil && handshake.Body != nil {
		_ = handshake.Body.Close()
	}
	if err != nil {
		return Unreachable, reason(err)
	}
	// Nothing is read. The handshake is the answer, and a room that has one
	// player's worth of packets queued should not have them drained by a check.
	_ = conn.Close(websocket.StatusNormalClosure, "")
	return Live, nil
}

// roomURL is the address the bridge would dial, built the same way.
func roomURL(s settings.Settings) string {
	scheme := "ws"
	if s.APTls {
		scheme = "wss"
	}
	return scheme + "://" + net.JoinHostPort(s.APHost, strconv.Itoa(s.APPort))
}

/*
	reason turns a dial error into something a player can act on.

The wrapped errors say things like "dial tcp 34.117.x.x:443: i/o timeout", which
names a machine the player has never heard of. What they can do about it is one
of three things, and which one depends on how it failed.
*/
func reason(err error) error {
	/* DNS first, and the order is the finding. A name that does not resolve
	   often takes longer than the timeout to say so, so the deadline is also
	   set by the time the error comes back and matching on it first reported
	   "it did not answer" for what is really a typo. The more specific
	   condition has to be asked first. */
	switch {
	case isDNS(err):
		return errors.New("that host name does not resolve, so check it for a typo")
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("it did not answer within %s", Timeout)
	}
	return err
}

func isDNS(err error) bool {
	var dns *net.DNSError
	return errors.As(err, &dns)
}

/*
	Advice is what to tell the player, and it is the same advice either way.

A room that is not configured and a room that is not answering leave the player
in one place: they need a room. The wording differs because being told to check
an address you have not typed reads as a bug, and being told to go and make one
you already made reads as not listening.
*/
func (r Result) Advice() string {
	switch r {
	case NotConfigured:
		return "No Archipelago room is set. Make one at archipelago.gg by uploading the seed, " +
			"then put its host and port on the Archipelago room page. " +
			"To play on your own without a room, turn Test mode on there instead."
	case Unreachable:
		return "The room is saved and the launcher will keep trying, so this is worth ignoring " +
			"if the room is not open yet. If it should be up, check the address against your room page " +
			"on archipelago.gg. To play on your own without a room, turn Test mode on instead."
	case Live, Skipped:
		// Nothing to advise: there is a room and it answered, or there is
		// deliberately no room to have.
		return ""
	}
	return ""
}
