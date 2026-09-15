// Command bridge holds the Archipelago session for one Team Fortress 2 server
// and serves the SourceMod plugin over loopback HTTP.
//
// It is the only component that knows the Archipelago protocol and the only
// one that knows the id mapping. See docs/en/adr/0002.
//
// The container image runs this binary directly. The launcher imports the same
// logic from the bridge package and runs it in-process alongside the srcds
// subprocess it supervises.
package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/m-this/tf2-archipelago/bridge"
	"github.com/m-this/tf2-archipelago/bridge/config"
)

func main() {
	output, closeLog := logOutput()
	logger := slog.New(slog.NewTextHandler(output, nil))
	if err := run(logger); err != nil {
		logger.Error("bridge stopped", "error", err)
		closeLog()
		os.Exit(1)
	}
	closeLog()
}

// logOutput keeps the container's normal stderr and gives the admin sidecar a
// read-only file to follow. Health-check processes must not truncate the live
// bridge's log every fifteen seconds.
func logOutput() (io.Writer, func()) {
	path := os.Getenv("BRIDGE_LOG_FILE")
	if path == "" || healthCheckProcess() {
		return os.Stderr, func() {}
	}
	file, err := os.Create(path)
	if err != nil {
		return os.Stderr, func() {}
	}
	return io.MultiWriter(os.Stderr, file), func() { _ = file.Close() }
}

func healthCheckProcess() bool {
	for _, arg := range os.Args[1:] {
		if arg == "-health" || arg == "--health" {
			return true
		}
	}
	return false
}

func run(logger *slog.Logger) error {
	health := flag.Bool("health", false,
		"ask the running bridge for its health and exit; this is the container health check")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if *health {
		return bridge.CheckHealth(cfg.Listen)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return bridge.Run(ctx, cfg, logger)
}
