//go:build !windows

package debugbundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/*
Lee Harrold's bundle held 112 crashes at replay_srv.so and could not say why:
the link srcds loads steamclient.so through might have been absent, dangling,
or fine. The summary names where it points and whether it leads anywhere.
*/
func TestTheSummaryNamesTheSteamClientLink(t *testing.T) {
	tests := []struct {
		name   string
		source bool
		want   string
	}{
		{name: "resolves", source: true, want: "resolves"},
		{name: "dangling", source: false, want: "DOES NOT RESOLVE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "steamcmd", "linux32", "steamclient.so")
			if tt.source {
				write(t, source, "ELF")
			}
			link := filepath.Join(root, ".steam", "sdk32", "steamclient.so")
			if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(source, link); err != nil {
				t.Fatal(err)
			}

			s := settings.Defaults()
			s.InstallRoot = root
			path, err := Write(s, nil, "", time.Now())
			if err != nil {
				t.Fatalf("Write: %v", err)
			}
			got := read(t, path)["summary.txt"]
			if !strings.Contains(got, "steamclient   "+link+" -> "+source+", "+tt.want) {
				t.Errorf("the summary does not describe the link:\n%s", got)
			}
		})
	}
}
