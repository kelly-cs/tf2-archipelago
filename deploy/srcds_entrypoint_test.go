package deploy_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestChangedAdminFileReloadsSourceModCache(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	game := filepath.Join(root, "tf-dedicated", "tf")
	configs := filepath.Join(game, "addons", "sourcemod", "configs")
	if err := os.MkdirAll(configs, 0o755); err != nil {
		t.Fatal(err)
	}
	calls := filepath.Join(root, "rcon-calls")
	fakeRCON := filepath.Join(root, "rcon")
	if err := os.WriteFile(fakeRCON, []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$TF2AP_RCON_CALLS\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	command := exec.Command("bash", "-c", `. deploy/srcds-entrypoint.sh
install_admin
wait
# An unchanged supervisor pass must not keep flushing SourceMod's cache.
install_admin
wait`)
	command.Dir = ".."
	command.Env = append(os.Environ(),
		"TF2AP_ENTRYPOINT_LIBRARY=1",
		"STEAMAPPDIR="+filepath.Join(root, "tf-dedicated"),
		"STEAMAPP=tf",
		"SRCDS_ADMIN_STEAMIDS=76561198019118556",
		"TF2AP_RCON="+fakeRCON,
		"TF2AP_RCON_CALLS="+calls,
		"TF2AP_ADMIN_RELOAD_ATTEMPTS=1",
		"TF2AP_ADMIN_RELOAD_INTERVAL=0",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("entrypoint test: %v\n%s", err, output)
	}

	admins, err := os.ReadFile(filepath.Join(configs, "admins_simple.ini"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(admins), `"STEAM_0:0:29426414" "99:z"`) {
		t.Fatalf("admin file does not contain converted Steam id:\n%s", admins)
	}
	reloads, err := os.ReadFile(calls)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(reloads); got != "sm_reloadadmins\n" {
		t.Fatalf("RCON calls = %q, want one admin-cache reload", got)
	}
}
