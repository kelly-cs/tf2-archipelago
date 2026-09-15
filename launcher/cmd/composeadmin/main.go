// Command composeadmin serves the launcher UI beside a Compose-managed TF2
// server. It observes the bridge and reaches SRCDS over their shared loopback
// namespace; Docker remains the lifecycle and configuration authority.
package main

import (
	"log/slog"
	"os"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/webapi"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("admin UI stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	s := settings.ApplyEnv(settings.Defaults())
	authority := env("TF2AP_ADMIN_AUTHORITY", "127.0.0.1:8477")
	logger.Info("TF2 Archipelago admin UI", "url", "http://"+authority)
	logger.Info("Docker Compose owns server lifecycle; the Settings tab writes .env")
	return webapi.Run(s, logger, webapi.Options{
		Address:           env("TF2AP_ADMIN_LISTEN", "0.0.0.0:8477"),
		Authority:         authority,
		Attached:          true,
		AttachedEnvFile:   env("TF2AP_ADMIN_ENV_FILE", "/config/compose.env"),
		TailscaleSocket:   env("TF2AP_TAILSCALE_SOCKET", "/run/tf2ap-fastdl/tailscaled.sock"),
		TailscaleHostname: env("TAILSCALE_HOSTNAME", "tf2-fastdl"),
		AttachedLogs: []webapi.AttachedLog{
			{Path: env("TF2AP_ADMIN_SRCDS_LOG", "/srv/tf-dedicated/tf/console.log"), Source: "srcds"},
			{Path: env("TF2AP_ADMIN_BRIDGE_LOG", "/srv/bridge/bridge.log"), Source: "bridge"},
		},
	})
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
