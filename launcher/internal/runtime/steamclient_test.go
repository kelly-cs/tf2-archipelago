//go:build !windows

package runtime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/installer"
)

// A srcds started without $HOME/.steam/sdk32/steamclient.so prints "Could not
// load: replay_srv.so" and segfaults, and srcds_run restarts it into the same
// crash. The launcher knows before the first start, so it says so instead.
func TestRunSrcdsRefusesAServerWithNoSteamClient(t *testing.T) {
	root := t.TempDir()
	s := fakeServer(t, root)
	if err := os.Remove(installer.InspectSteamClient(root).Link); err != nil {
		t.Fatal(err)
	}
	started := filepath.Join(root, "started")
	script := "#!/bin/sh\ntouch " + started + "\n"
	if err := os.WriteFile(filepath.Join(root, "tf-dedicated", "srcds_run"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	err := runSrcdsWithSink(context.Background(), s, nil, func(Line) {})
	if !errors.Is(err, installer.ErrNoSteamClient) {
		t.Fatalf("runSrcdsWithSink = %v, want ErrNoSteamClient", err)
	}
	if _, statErr := os.Stat(started); statErr == nil {
		t.Error("srcds_run was started anyway")
	}
}
