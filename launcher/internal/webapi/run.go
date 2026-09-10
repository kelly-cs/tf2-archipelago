package webapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m-this/tf2-archipelago/launcher/internal/browser"
	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// shutdownGrace bounds the wait for the last request when the launcher is
// closing. Nothing here is worth making the player wait for.
const shutdownGrace = 3 * time.Second

// Options is how the launcher was asked to serve.
type Options struct {
	// Address to bind. Empty means 127.0.0.1 on a port the operating system
	// picks, which is what a double-click gets: nothing to configure, and no
	// second launcher fighting for the same port.
	Address string

	// OpenBrowser is false for a player who would rather open the link
	// themselves, and for a machine with no desktop to open it on.
	OpenBrowser bool

	// Serving is told the address once it answers, with what Quit does. It is
	// how a tray icon knows what to open and what to close. Nil is fine.
	Serving func(url string, quit func())
}

/*
Run is the launcher: it serves the interface on loopback, opens it, and blocks
until Quit or a signal.

Loopback only, and one player. This is not a daemon: closing the tab leaves the
server running, Quit in the interface stops it, and so does Ctrl-C here.
*/
func Run(s settings.Settings, logger *slog.Logger, options Options) error {
	app := New(s, logger)
	if file, err := apruntime.CreateLogFile(s.InstallRoot); err == nil {
		app.LogTo(file)
		defer func() { _ = file.Close() }()
	}

	address := options.Address
	if address == "" {
		address = "127.0.0.1:0"
	}
	var config net.ListenConfig
	listener, err := config.Listen(context.Background(), "tcp4", address)
	if err != nil {
		return fmt.Errorf("cannot start the launcher interface: %w", err)
	}
	authority := listener.Addr().String()
	url := "http://" + authority

	server := &http.Server{Handler: app.Handler(authority), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Say("interface: %v", err)
			app.Quit()
		}
	}()
	go app.WatchSession()

	announce(app, url, options.OpenBrowser)
	if options.Serving != nil {
		options.Serving(url, app.Quit)
	}
	if s.APPort != 0 || s.TestMode {
		app.Start()
	} else {
		app.OpenSettings("Archipelago room")
	}

	signals, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	select {
	case <-signals.Done():
	case <-app.Quitting():
	}
	app.Stop()
	ctx, stop := context.WithTimeout(context.Background(), shutdownGrace)
	defer stop()
	return server.Shutdown(ctx)
}

// announce puts the address where the player will find it, whether or not a
// browser opened. A desktop with no opener is normal over SSH, in a container
// and on a minimal install, so a failure is a line rather than a reason to stop.
func announce(app *App, url string, open bool) {
	fmt.Fprintf(os.Stderr, "\n    %s\n\n", url)
	app.Say("interface: %s", url)
	if !open {
		return
	}
	if err := browser.Open(url); err != nil {
		app.Say("no browser opened (%v). Open the address above yourself.", err)
	}
}
