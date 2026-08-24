package settings

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSaveLoadRoundTrip writes a config, reads it back, and checks the values
// survive. Uses a temp HOME so it does not touch the operator's real config.
func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", filepath.Join(t.TempDir(), "AppData", "Roaming"))

	s := Defaults()
	s.SrcdsRconPw = "hunter2"
	s.APPort = 12345
	s.APHost = "archipelago.gg"
	s.APTls = true
	s.SrcdsAdminSteamIDs = "76561198014216803"

	if err := Save(s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.SrcdsRconPw != s.SrcdsRconPw {
		t.Errorf("RCON password: got %q, want %q", loaded.SrcdsRconPw, s.SrcdsRconPw)
	}
	if loaded.APPort != s.APPort {
		t.Errorf("AP port: got %d, want %d", loaded.APPort, s.APPort)
	}
	if loaded.APHost != s.APHost {
		t.Errorf("AP host: got %q, want %q", loaded.APHost, s.APHost)
	}
	if loaded.APTls != s.APTls {
		t.Errorf("AP TLS: got %v, want %v", loaded.APTls, s.APTls)
	}
}

// TestLoadMissingReturnsDefaults checks that a missing config file yields the
// defaults rather than an error.
func TestLoadMissingReturnsDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", filepath.Join(t.TempDir(), "AppData", "Roaming"))

	s, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	d := Defaults()
	if s.SrcdsPort != d.SrcdsPort {
		t.Errorf("SrcdsPort: got %d, want %d", s.SrcdsPort, d.SrcdsPort)
	}
	if s.MvmDifficulty != d.MvmDifficulty {
		t.Errorf("MvmDifficulty: got %q, want %q", s.MvmDifficulty, d.MvmDifficulty)
	}
}

// TestApplyDefaultsFillsZero checks that an old config missing a field gets the
// default for it.
func TestApplyDefaultsFillsZero(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", filepath.Join(t.TempDir(), "AppData", "Roaming"))

	s := Settings{SrcdsRconPw: "x"}
	if err := Save(s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.SrcdsMaxPlayers != 32 {
		t.Errorf("SrcdsMaxPlayers: got %d, want 32 (default)", loaded.SrcdsMaxPlayers)
	}
}

// TestConfigFilePermissions checks the config file is 0600: it holds the RCON
// password.
func TestConfigFilePermissions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", filepath.Join(t.TempDir(), "AppData", "Roaming"))

	if err := Save(Defaults()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	path, _ := Path()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config file mode: got %o, want 0600", info.Mode().Perm())
	}
}

// A file from before the start mission existed names a map. The mission the
// server starts on is then that map's first mission.
func TestLoadMigratesTheStartMap(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APPDATA", filepath.Join(t.TempDir(), "AppData", "Roaming"))

	path, _ := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"srcds_start_map": "mvm_coaltown"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.SrcdsStartMission != "mvm_coaltown" {
		t.Errorf("start mission = %q, want the first Coal Town mission", loaded.SrcdsStartMission)
	}
	if loaded.SrcdsStartMap != "" {
		t.Errorf("the start map survived the migration: %q", loaded.SrcdsStartMap)
	}
}

func TestLoadMovesAnUnsupportedRafModStartToSafeDefaults(t *testing.T) {
	loaded, err := parse([]byte(`{
		"srcds_start_mission":"mvm_kelly_rc1b_adv_mobocracy",
		"mvm_start_mission":"mvm_kelly_rc1b_adv_mobocracy"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SrcdsStartMission != Defaults().SrcdsStartMission {
		t.Errorf("server start = %q, want %q", loaded.SrcdsStartMission, Defaults().SrcdsStartMission)
	}
	if loaded.MvmStartMission != "" {
		t.Errorf("run still starts on unsupported mission %q", loaded.MvmStartMission)
	}
}

// A config file written before the appearance switches existed does not
// mention them, and a bool nobody wrote reads back as false. Every install that
// had ever saved a setting opened the next version with the bots undressed.
func TestAConfigThatPredatesTheAppearanceSwitchesGetsTheDefaults(t *testing.T) {
	s, err := parse([]byte(`{"srcds_bot_team_size": 6}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if !s.SrcdsBotHats || !s.SrcdsBotHatEffects {
		t.Errorf("hats %v, effects %v; want both on", s.SrcdsBotHats, s.SrcdsBotHatEffects)
	}
}

// A file that says no is not a file that says nothing: unticking one has to
// survive the next start.
func TestTheAppearanceSwitchesStayOffWhenTheFileSaysSo(t *testing.T) {
	s, err := parse([]byte(`{"srcds_bot_hats": false, "srcds_bot_hat_effects": true}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if s.SrcdsBotHats {
		t.Error("hats were off in the file and came back on")
	}
	if !s.SrcdsBotHatEffects {
		t.Error("effects were on in the file and came back off")
	}
}
