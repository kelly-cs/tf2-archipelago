package roomcheck

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/coder/websocket"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// roomAt points the settings at a server this test is running.
func roomAt(t *testing.T, url string) settings.Settings {
	t.Helper()
	host, port, err := net.SplitHostPort(strings.TrimPrefix(url, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	number, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}
	s := settings.Defaults()
	s.APHost, s.APPort, s.APTls = host, number, false
	return s
}

// A room that speaks websocket is live. Nothing more is asked of it: the seed,
// the slot and the rest are the bridge's business when it connects for real.
func TestARoomThatCompletesTheHandshakeIsLive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.CloseNow() }()
		// A real room sends RoomInfo here. Check reads nothing, so this only
		// has to hold the connection open long enough to be closed.
		<-r.Context().Done()
	}))
	defer server.Close()

	got, err := Check(context.Background(), roomAt(t, server.URL))
	if err != nil {
		t.Fatalf("a live room reported %v", err)
	}
	if got != Live {
		t.Errorf("a live room came back as %v", got)
	}
	if got.Advice() != "" {
		t.Errorf("a live room has advice for the player: %q", got.Advice())
	}
}

// An address with nothing behind it is unreachable, and the advice says the
// settings are saved anyway, because they are.
func TestAnAddressWithNothingBehindItIsUnreachable(t *testing.T) {
	// A port nothing is listening on: take one and give it straight back.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	url := "http://" + listener.Addr().String()
	_ = listener.Close()

	got, err := Check(context.Background(), roomAt(t, url))
	if got != Unreachable {
		t.Fatalf("a dead address came back as %v (%v)", got, err)
	}
	if err == nil {
		t.Error("a dead address gave no reason")
	}
	if !strings.Contains(got.Advice(), "saved") {
		t.Errorf("the advice does not say the settings are saved: %q", got.Advice())
	}
	if !strings.Contains(got.Advice(), "Test mode") {
		t.Error("the advice does not offer the way to play without a room")
	}
}

/*
	A fresh install has no room, and is told where to get one.

This is the case the whole check exists for. The reporter's config.json held
ap_port 0, the defaults, and nothing anywhere said that meant no room: Start
refused with "AP_PORT is not set" and the settings screen said nothing at all.
*/
func TestNoRoomSaysWhereToGetOne(t *testing.T) {
	got, err := Check(context.Background(), settings.Defaults())
	if err != nil {
		t.Fatalf("no room reported an error: %v", err)
	}
	if got != NotConfigured {
		t.Fatalf("the defaults came back as %v", got)
	}
	for _, want := range []string{"archipelago.gg", "Test mode"} {
		if !strings.Contains(got.Advice(), want) {
			t.Errorf("the advice does not mention %q: %q", want, got.Advice())
		}
	}
}

// Test mode is a multiworld of one served on loopback, so there is nothing to
// dial and nothing to say.
func TestTestModeIsNotDialled(t *testing.T) {
	s := settings.Defaults()
	s.TestMode = true
	// An address that would hang if anything tried it.
	s.APHost, s.APPort = "10.255.255.1", 1

	got, err := Check(context.Background(), s)
	if err != nil || got != Skipped {
		t.Errorf("test mode came back as %v (%v)", got, err)
	}
	if got.Advice() != "" {
		t.Errorf("test mode has advice: %q", got.Advice())
	}
}

/*
	A host that does not resolve is named as a typo, not as a resolver error.

reason is fed the error directly rather than a real lookup, because what
resolves and how fast is the machine's business, not this package's. On a box
with no DNS at all the deadline fires before the resolver gives up and the error
carries no DNSError, which is exactly why the switch in reason asks the specific
question first: matching the deadline first reported a typo as "it did not
answer".
*/
func TestAHostThatDoesNotResolveSaysSo(t *testing.T) {
	lookup := &net.DNSError{Err: "no such host", Name: "archipelagoo.gg", IsNotFound: true}

	for _, err := range []error{
		lookup,
		fmt.Errorf("dial: %w", lookup),
		// Both at once, which is the case that was reported wrongly.
		fmt.Errorf("%w: %w", context.DeadlineExceeded, lookup),
	} {
		got := reason(err)
		if got == nil || !strings.Contains(got.Error(), "typo") {
			t.Errorf("reason(%v) is %v, which does not tell the player what to look at", err, got)
		}
	}
}

// A room that is simply slow is named as slow, and the timeout is in the
// message so the number is not a mystery.
func TestARoomThatDoesNotAnswerInTimeSaysHowLongItWaited(t *testing.T) {
	got := reason(fmt.Errorf("dial: %w", context.DeadlineExceeded))
	if got == nil || !strings.Contains(got.Error(), Timeout.String()) {
		t.Errorf("reason is %v, and does not say how long it waited", got)
	}
}
