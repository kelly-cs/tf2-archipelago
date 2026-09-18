package config

import (
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ArchipelagoURL != "ws://archipelago:38281" {
		t.Errorf("url = %q", cfg.ArchipelagoURL)
	}
	if cfg.SlotName != "tf2" {
		t.Errorf("slot = %q", cfg.SlotName)
	}
	if cfg.Listen != "127.0.0.1:24680" {
		t.Errorf("listen = %q", cfg.Listen)
	}
	if cfg.PollTimeout != 25*time.Second {
		t.Errorf("poll timeout = %s", cfg.PollTimeout)
	}
}

func TestTLSChangesTheScheme(t *testing.T) {
	t.Setenv("AP_TLS", "true")
	t.Setenv("AP_HOST", "ap.example.org")
	t.Setenv("AP_PORT", "443")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ArchipelagoURL != "wss://ap.example.org:443" {
		t.Errorf("url = %q", cfg.ArchipelagoURL)
	}
}

func TestBadValuesAreRefusedAtStartup(t *testing.T) {
	tests := map[string][2]string{
		"port that is not a number": {"AP_PORT", "http"},
		"tls that is not a boolean": {"AP_TLS", "yes please"},
		"timeout that is not one":   {"BRIDGE_POLL_TIMEOUT", "soon"},
		"timeout that is negative":  {"BRIDGE_POLL_TIMEOUT", "-5s"},
	}
	for name, pair := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv(pair[0], pair[1])
			if _, err := Load(); err == nil {
				t.Fatalf("%s=%q was accepted", pair[0], pair[1])
			}
		})
	}
}

func TestTestModeFromTheEnvironment(t *testing.T) {
	t.Setenv("TF2AP_TEST_MODE", "1")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.TestMode {
		t.Error("TF2AP_TEST_MODE=1 did not turn test mode on")
	}

	t.Setenv("TF2AP_TEST_MODE", "nonsense")
	if _, err := Load(); err == nil {
		t.Error("a value that is not a boolean was accepted")
	}
}

func TestDockerTestRunSettings(t *testing.T) {
	t.Setenv("TF2AP_TEST_MODE", "1")
	t.Setenv("MVM_MISSION_MODIFIERS", "true")
	t.Setenv("MVM_MINIMUM_MISSION_MODIFIERS", "2")
	t.Setenv("MVM_MAXIMUM_MISSION_MODIFIERS", "3")
	t.Setenv("MVM_VICTORY_CACHES", "true")
	t.Setenv("MVM_MILESTONE_CHECKS", "true")
	t.Setenv("MVM_GIANTSANITY", "true")
	t.Setenv("MVM_TANKSANITY", "true")
	t.Setenv("MVM_EXCLUDED_MISSIONS", "mvm_decoy, mvm_mannhattan")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	run := cfg.TestRun
	if !run.MissionModifiers || run.ModifierMin != 2 || run.ModifierMax != 3 || !run.VictoryCaches || !run.MilestoneChecks || !run.Giantsanity || !run.Tanksanity {
		t.Errorf("test run settings were lost: %+v", run)
	}
	if len(run.Excluded) != 2 || run.Excluded[1] != "mvm_mannhattan" {
		t.Errorf("excluded missions = %v", run.Excluded)
	}
}
