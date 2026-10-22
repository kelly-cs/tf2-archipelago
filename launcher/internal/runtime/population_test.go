package runtime

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/installer"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// The bridge refuses a seed whose community missions have no population file,
// so it has to look where the installer put them. A directory one level off
// would refuse every community seed on every server.
func TestTheBridgeLooksForMissionsWhereTheInstallerPutsThem(t *testing.T) {
	const popFile = "mvm_transmission_rc7a_int_binary_broadcast"
	s := settings.Defaults()
	s.InstallRoot = t.TempDir()
	s.SrcdsRconPw = "secret"
	s.TestMode = true

	archive := filepath.Join(t.TempDir(), "pack.zip")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	if _, err := writer.Create("tf/download/scripts/population/" + popFile + ".pop"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	modDir := filepath.Join(s.InstallRoot, "tf-dedicated", "tf")
	if err := installer.InstallCommunityArchives([]string{archive}, modDir, nil, t.Logf); err != nil {
		t.Fatal(err)
	}

	cfg, err := bridgeConfig(s)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PopulationDir == "" {
		t.Fatal("the bridge was given no population directory, so it checks nothing")
	}
	if _, err := os.Stat(filepath.Join(cfg.PopulationDir, popFile+".pop")); err != nil {
		t.Errorf("the bridge looks in %s and the installer put the mission elsewhere: %v", cfg.PopulationDir, err)
	}
}
