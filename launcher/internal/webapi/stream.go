package webapi

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	launcherv1 "github.com/m-this/tf2-archipelago/launcher/internal/gen/tf2ap/launcher/v1"
	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
)

/*
The live stream, at /ws.

It is a WebSocket rather than a Connect server stream because only one side ever
speaks: the launcher pushes, and everything the browser wants to say is an RPC.

The first frame is the whole state, logs and all, taken after the subscription
started so nothing falls between the two. After it, a log line arrives on its
own and the whole state arrives whenever anything else moved, with the logs left
out because the browser already has them.

Bounded on both sides: the launcher keeps linesMax lines, a listener may be
listenerQueue events behind, and a browser that falls further gets the whole
state again rather than a queue that grows or an event silently lost.
*/

// writeGrace bounds one frame. Loopback either takes it or the browser is gone.
const writeGrace = 10 * time.Second

func (a *App) serveStream(authority string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		socket, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{authority},
		})
		if err != nil {
			return
		}
		defer func() { _ = socket.CloseNow() }()

		// Nothing is ever read from the browser, but a socket nobody reads
		// never sees a close or a ping. CloseRead does that and cancels this
		// context when the browser goes away.
		ctx := socket.CloseRead(r.Context())

		listener, done := a.Subscribe()
		defer done()

		if err := a.sendState(ctx, socket, true); err != nil {
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-a.quit:
				_ = socket.Close(websocket.StatusGoingAway, "the launcher is closing")
				return
			case message := <-listener.Events():
				if err := a.forward(ctx, socket, listener, message); err != nil {
					return
				}
			}
		}
	}
}

// forward turns one event into one frame. A listener that fell behind gets the
// whole state instead of the event it is holding, because the events it lost
// were the ones that said what changed.
func (a *App) forward(ctx context.Context, socket *websocket.Conn, listener *Listener, message Event) error {
	if listener.Behind() {
		return a.sendState(ctx, socket, false)
	}
	if line, ok := message.Data.(apruntime.Line); ok && message.Name == "log" {
		return send(ctx, socket, &launcherv1.StreamMessage{
			Body: &launcherv1.StreamMessage_Line{Line: lineProto(line)},
		})
	}
	return a.sendState(ctx, socket, false)
}

// sendState writes the whole state. withLogs is true only for the first frame:
// after it the browser has every line and is sent each new one as it happens.
//
//nolint:contextcheck // Address discovery owns its timeout instead of the frame.
func (a *App) sendState(ctx context.Context, socket *websocket.Conn, withLogs bool) error {
	state := a.Snapshot().Proto()
	if !withLogs {
		state.Logs = nil
	}
	return send(ctx, socket, &launcherv1.StreamMessage{
		Body: &launcherv1.StreamMessage_Snapshot{Snapshot: state},
	})
}

func send(ctx context.Context, socket *websocket.Conn, message *launcherv1.StreamMessage) error {
	frame, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, writeGrace)
	defer cancel()
	return socket.Write(ctx, websocket.MessageBinary, frame)
}
