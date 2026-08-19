package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

/*
Where a Linux srcds looks for steamclient.so, and how it gets there.

srcds_run dlopens $HOME/.steam/sdk32/steamclient.so on the way up. Without it
the server prints "Could not load: replay_srv.so", names the file it could not
open, and segfaults. Nothing in the game's own install puts it there: SteamCMD
downloads it for itself, and every guide to running a Source server on Linux
tells the operator to link the two by hand.

The link goes inside the install root rather than in the operator's real home,
and srcds runs with HOME pointed at the install root. That directory is the
one this project promises holds everything, so deleting it takes this with it
and leaves nothing of ours behind. See SteamHome.

Windows needs none of this and gets none of it.
*/
func linkSteamClient(installRoot, steamcmdDir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	// linux32 for the 32-bit server, linux64 because srcds_run checks both and
	// a missing sdk64 is one more line of noise in the log.
	for _, sdk := range []struct{ from, to string }{
		{"linux32", "sdk32"},
		{"linux64", "sdk64"},
	} {
		source := filepath.Join(steamcmdDir, sdk.from, "steamclient.so")
		if !exists(source) {
			continue
		}
		dir := filepath.Join(SteamHome(installRoot), ".steam", sdk.to)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("cannot create %s: %w", dir, err)
		}
		target := filepath.Join(dir, "steamclient.so")
		// An install run twice finds its own link already there.
		_ = os.Remove(target)
		if err := os.Symlink(source, target); err != nil {
			return fmt.Errorf("cannot link steamclient.so into %s: %w", dir, err)
		}
	}
	return nil
}

/*
dropMismatchedLoader removes the Metamod loader for the other architecture.

The Source engine loads every .vdf in addons/, and Metamod's archive ships two:
metamod.vdf points at the 32-bit loader and metamod_x64.vdf at the 64-bit one.
The launcher runs the 32-bit server, because that is what the defender bots'
two extensions are built for, so the x64 loader is a file the engine is certain
to try and certain to refuse:

	failed to dlopen .../metamod/bin/linux64/server.so: wrong ELF class

Leaving it there puts a FATAL ERROR in the log of every healthy boot, which is
the kind of line that sends somebody debugging the wrong thing.
*/
func dropMismatchedLoader(modDir string) error {
	unused := filepath.Join(modDir, "addons", "metamod_x64.vdf")
	if err := os.Remove(unused); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove %s: %w", unused, err)
	}
	return nil
}

// SteamHome is the HOME the game server runs with. It is the install root, so
// the .steam directory srcds insists on lands beside the game files instead of
// in the operator's own home.
func SteamHome(installRoot string) string { return installRoot }
