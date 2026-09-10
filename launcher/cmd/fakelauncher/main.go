/*
fakelauncher is the launcher with nothing behind it.

It answers the same contract the real one does, over the same Connect handlers
and the same WebSocket, and serves the same embedded app. What it does not have
is a game server, a bridge, a Steam install or a multiworld: the state is in
memory and the log is a script.

That is what the browser tests want. Playwright drives the real app against the
real proto and the real form model, and the run is the same every time, on a
machine with no Team Fortress 2 on it.

	go run ./launcher/cmd/fakelauncher -addr 127.0.0.1:8099

Never shipped: it is not in any release target, and the real launcher does not
import it.
*/
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m-this/tf2-archipelago/launcher/internal/spa"
)

func main() {
	address := flag.String("addr", "127.0.0.1:8099", "where to listen")
	flag.Parse()

	fake := newFake()
	var config net.ListenConfig
	listener, err := config.Listen(context.Background(), "tcp4", *address)
	if err != nil {
		log.Fatalf("fakelauncher: %v", err)
	}
	authority := listener.Addr().String()

	mux := http.NewServeMux()
	fake.register(mux, authority)
	mux.Handle("/", spa.Handler())

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("fakelauncher: %v", err)
		}
	}()
	fmt.Printf("fakelauncher on http://%s\n", authority)

	signals, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-signals.Done()

	ctx, done := context.WithTimeout(context.Background(), 2*time.Second)
	defer done()
	_ = server.Shutdown(ctx)
}
