package installer

import (
	"context"
	"errors"
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
func ensureSteamClient(ctx context.Context, installRoot, steamcmdDir string, logf func(string, ...any)) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	// A SteamCMD unpacked by Repair beside an installed game has never run, and
	// steamclient.so is one of the files its first run downloads.
	if !exists(filepath.Join(steamcmdDir, "linux32", "steamclient.so")) {
		if exe := firstExisting(steamcmdDir, steamcmdNames()); exe != "" {
			warmUpSteamcmd(ctx, exe, steamcmdDir, logf)
		}
	}
	return linkSteamClient(installRoot, steamcmdDir)
}

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
		if !exists(source) && sdk.to == "sdk64" {
			continue
		}
		if !exists(source) {
			return fmt.Errorf("%w: SteamCMD has not downloaded %s, and the game server "+
				"crashes at replay_srv.so without it. %s", ErrNoSteamClient, source, RepairAdvice)
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

// ErrNoSteamClient is a game server that would start without the steamclient.so
// srcds_run loads, and crash at replay_srv.so.
var ErrNoSteamClient = errors.New("the game server has no steamclient.so")

// SteamClient is the link srcds_run loads steamclient.so through, and whether
// it leads to a file.
type SteamClient struct {
	Link   string
	Target string
	Err    error
}

// InspectSteamClient reads the sdk32 link under the install root. Err is nil
// only when the link resolves to a file the server can load.
func InspectSteamClient(installRoot string) SteamClient {
	got := SteamClient{Link: filepath.Join(SteamHome(installRoot), ".steam", "sdk32", "steamclient.so")}
	got.Target, _ = os.Readlink(got.Link)
	info, err := os.Stat(got.Link)
	switch {
	case err != nil:
		got.Err = fmt.Errorf("%w: %s does not lead to a file: %w", ErrNoSteamClient, got.Link, err)
	case !info.Mode().IsRegular():
		got.Err = fmt.Errorf("%w: %s is not a file", ErrNoSteamClient, got.Link)
	}
	return got
}

// SteamHome is the HOME the game server runs with. It is the install root, so
// the .steam directory srcds insists on lands beside the game files instead of
// in the operator's own home.
func SteamHome(installRoot string) string { return installRoot }
