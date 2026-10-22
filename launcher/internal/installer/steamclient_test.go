//go:build !windows

package installer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSteamClient(t *testing.T, steamcmdDir string) string {
	t.Helper()
	source := filepath.Join(steamcmdDir, "linux32", "steamclient.so")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("ELF"), 0o644); err != nil {
		t.Fatal(err)
	}
	return source
}

/*
Lee Harrold's server on Bazzite segfaulted at replay_srv.so on every start, the
signature of srcds finding no $HOME/.steam/sdk32/steamclient.so. linkSteamClient
skipped the link without a word when SteamCMD had no steamclient.so to link, so
the install reported success and the server died 112 times.
*/
func TestLinkSteamClientRefusesAMissingSource(t *testing.T) {
	root := t.TempDir()
	steamcmdDir := steamcmdPath(root)
	if err := os.MkdirAll(steamcmdDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := linkSteamClient(root, steamcmdDir)
	if !errors.Is(err, ErrNoSteamClient) {
		t.Fatalf("linkSteamClient = %v, want ErrNoSteamClient", err)
	}
	want := filepath.Join(steamcmdDir, "linux32", "steamclient.so")
	if !strings.Contains(err.Error(), want) {
		t.Errorf("the error does not name the missing file %s: %v", want, err)
	}
}

func TestInspectSteamClient(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, root string) (target string)
		resolve bool
	}{
		{
			name: "linked by the installer",
			prepare: func(t *testing.T, root string) string {
				t.Helper()
				source := writeSteamClient(t, steamcmdPath(root))
				if err := linkSteamClient(root, steamcmdPath(root)); err != nil {
					t.Fatalf("linkSteamClient: %v", err)
				}
				return source
			},
			resolve: true,
		},
		{
			name:    "never linked",
			prepare: func(*testing.T, string) string { return "" },
		},
		{
			name: "linked, then SteamCMD lost the file",
			prepare: func(t *testing.T, root string) string {
				t.Helper()
				source := writeSteamClient(t, steamcmdPath(root))
				if err := linkSteamClient(root, steamcmdPath(root)); err != nil {
					t.Fatalf("linkSteamClient: %v", err)
				}
				if err := os.Remove(source); err != nil {
					t.Fatal(err)
				}
				return source
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			target := tt.prepare(t, root)

			got := InspectSteamClient(root)
			if want := filepath.Join(root, ".steam", "sdk32", "steamclient.so"); got.Link != want {
				t.Errorf("Link = %q, want %q", got.Link, want)
			}
			if got.Target != target {
				t.Errorf("Target = %q, want %q", got.Target, target)
			}
			if tt.resolve && got.Err != nil {
				t.Errorf("a working link was refused: %v", got.Err)
			}
			if !tt.resolve && !errors.Is(got.Err, ErrNoSteamClient) {
				t.Errorf("Err = %v, want ErrNoSteamClient", got.Err)
			}
		})
	}
}

// Repair throws SteamCMD away and unpacks a fresh one beside an installed game.
// That SteamCMD has never run, so it has no steamclient.so yet, and refusing
// the start there would make Repair the way to break an install.
func TestEnsureSteamClientRunsAFreshSteamcmd(t *testing.T) {
	root := t.TempDir()
	steamcmdDir := steamcmdPath(root)
	if err := os.MkdirAll(steamcmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nmkdir -p \"$(dirname \"$0\")/linux32\"\n: > \"$(dirname \"$0\")/linux32/steamclient.so\"\n"
	if err := os.WriteFile(filepath.Join(steamcmdDir, "steamcmd.sh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := ensureSteamClient(context.Background(), root, steamcmdDir, discard); err != nil {
		t.Fatalf("ensureSteamClient: %v", err)
	}
	if err := InspectSteamClient(root).Err; err != nil {
		t.Errorf("the link does not resolve after SteamCMD ran: %v", err)
	}
}
