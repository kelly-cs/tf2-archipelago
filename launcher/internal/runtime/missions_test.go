package runtime

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// The bridge refuses a seed this server cannot play, so the launcher has to
// hand it what Start sets up: the packs ticked and on disk. A pack on disk and
// not ticked is never installed into the server.
func TestTheBridgeIsHandedThePacksStartInstalls(t *testing.T) {
	binaryBroadcast, ok := gamedata.MissionByPopFile("mvm_transmission_rc7a_int_binary_broadcast")
	if !ok {
		t.Fatal("Binary Broadcast is not in the catalog")
	}
	tests := []struct {
		name     string
		packs    []string
		playable bool
	}{
		{name: "pack ticked", packs: []string{settings.CommunityPackPotato}, playable: true},
		{name: "pack on disk and not ticked", packs: nil, playable: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := settings.Defaults()
			s.InstallRoot = t.TempDir()
			s.CommunityContentDir = t.TempDir()
			s.CommunityPacks = test.packs
			s.SrcdsRconPw = "secret"
			s.TestMode = true
			writeEmptyZip(t, filepath.Join(s.CommunityContentDir, settings.CommunityPackPotato))

			cfg, err := bridgeConfig(s)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Missions == nil {
				t.Fatal("the bridge was handed no server, so it checks nothing")
			}
			verdict := cfg.Missions.Check(binaryBroadcast)
			if verdict.Playable() != test.playable {
				t.Errorf("Binary Broadcast playable = %v, want %v (%s)", verdict.Playable(), test.playable, verdict.Why())
			}
		})
	}
}

func writeEmptyZip(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := zip.NewWriter(file).Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
