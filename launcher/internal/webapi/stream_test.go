package webapi

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	launcherv1 "github.com/m-this/tf2-archipelago/launcher/internal/gen/tf2ap/launcher/v1"
	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// streamGrace bounds every read in this file. Everything under test is on
// loopback in the same process, so a read that takes longer is a stream that
// stopped rather than a slow one.
const streamGrace = 5 * time.Second

func listening(t *testing.T) (*App, *websocket.Conn) {
	t.Helper()
	app := New(settings.Defaults(), nil)
	server := httptest.NewServer(nil)
	server.Config.Handler = app.Handler(strings.TrimPrefix(server.URL, "http://"))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), streamGrace)
	defer cancel()
	// The handshake response body is the socket; closing it here would close
	// the stream this test is about, and CloseNow below owns it.
	//nolint:bodyclose // The socket owns the handshake response.
	socket, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = socket.CloseNow() })
	return app, socket
}

func nextFrame(t *testing.T, socket *websocket.Conn) *launcherv1.StreamMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), streamGrace)
	defer cancel()
	kind, frame, err := socket.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if kind != websocket.MessageBinary {
		t.Fatalf("the stream sent %v, want a binary frame", kind)
	}
	var message launcherv1.StreamMessage
	if err := proto.Unmarshal(frame, &message); err != nil {
		t.Fatal(err)
	}
	return &message
}

// The first frame is the whole state. A browser that had to ask for it
// separately would lose every line published in between.
func TestTheStreamOpensWithTheWholeState(t *testing.T) {
	app, socket := listening(t)
	app.Say("hello from the launcher")

	first := nextFrame(t, socket)
	state := first.GetSnapshot()
	if state == nil {
		t.Fatal("the first frame was not a snapshot")
	}
	if state.GetTitle() == "" {
		t.Error("the first snapshot carries no title")
	}
}

func TestALineArrivesOnItsOwn(t *testing.T) {
	app, socket := listening(t)
	if nextFrame(t, socket).GetSnapshot() == nil {
		t.Fatal("the first frame was not a snapshot")
	}
	app.Say("the server said something")

	for range 4 {
		if line := nextFrame(t, socket).GetLine(); line != nil {
			if line.GetText() != "the server said something" {
				t.Errorf("the line arrived as %q", line.GetText())
			}
			return
		}
	}
	t.Fatal("no log line reached the stream")
}

// Every frame after the first leaves the logs out: the browser already has
// them, and re-sending twenty thousand lines on every redraw is the difference
// between a tab that costs nothing and one that does not.
func TestARedrawDoesNotResendTheLog(t *testing.T) {
	app, socket := listening(t)
	app.Say("a line to have")
	if nextFrame(t, socket).GetSnapshot() == nil {
		t.Fatal("the first frame was not a snapshot")
	}
	app.OpenSettings("")

	for range 6 {
		if state := nextFrame(t, socket).GetSnapshot(); state != nil {
			if len(state.GetLogs()) != 0 {
				t.Errorf("a redraw carried %d log lines", len(state.GetLogs()))
			}
			if state.GetForm() == nil {
				t.Error("the redraw did not carry the settings screen that opened")
			}
			return
		}
	}
	t.Fatal("opening the settings did not redraw the stream")
}

// A browser that stops reading must not grow the launcher's memory, and must
// not be left drawing a state that has moved on. It gets the whole state back.
func TestAListenerThatFallsBehindIsToldToReadEverything(t *testing.T) {
	app := New(settings.Defaults(), nil)
	listener, done := app.Subscribe()
	defer done()

	for i := range listenerQueue * 2 {
		app.append(apruntime.Line{At: time.Now(), Source: "test", Text: "line"})
		_ = i
	}
	if len(listener.events) > listenerQueue {
		t.Fatalf("the queue grew to %d, past its bound of %d", len(listener.events), listenerQueue)
	}
	if !listener.Behind() {
		t.Fatal("a listener that lost events was not told it had")
	}
	if listener.Behind() {
		t.Error("reading behind twice reported the same drop again")
	}
}

// The log itself is bounded: a long evening of srcds must not be the reason a
// launcher runs out of memory.
func TestTheLogStopsAtItsBound(t *testing.T) {
	app := New(settings.Defaults(), nil)
	for range linesMax + 500 {
		app.append(apruntime.Line{At: time.Now(), Source: "srcds", Text: "line"})
	}
	if got := len(app.Snapshot().Logs); got != linesMax {
		t.Errorf("the log holds %d lines, want %d", got, linesMax)
	}
}
