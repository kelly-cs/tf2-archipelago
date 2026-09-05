package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func data(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	meta := `{"game": "Team Fortress 2 Mann vs Machine", "classes": [{"name": "Scout"}, {"name": "Heavy"}], "server_mods": [{"key": "bots"}]}`
	missions := `{"missions": [{"name": "Doe's Drill", "pop_file": "mvm_decoy"}, {"name": "Day of Wreckoning", "pop_file": "mvm_decoy_advanced"}]}`
	for name, body := range map[string]string{"meta.json": meta, "missions.json": missions} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func env(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func TestDefaultsAndTranslation(t *testing.T) {
	out, err := build(data(t), "0.6.7", env(map[string]string{
		"MVM_EXCLUDED_MISSIONS": "mvm_decoy, mvm_decoy_advanced",
		"MVM_START_MISSION":     "mvm_decoy",
		"MVM_START_CLASS":       "Heavy",
		"SRCDS_MODS":            "bots",
	}))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"name: \"tf2\"\n", "  version: 0.6.7\n", "  mission_count: 8\n",
		"  start_mission: \"Doe's Drill\"\n", "  start_class: \"Heavy\"\n",
		"  excluded_missions:\n    - \"Doe's Drill\"\n    - \"Day of Wreckoning\"\n",
		"  server_mods:\n    - \"bots\"\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestAnUnknownPopfileIsAnError(t *testing.T) {
	_, err := build(data(t), "0.6.7", env(map[string]string{"MVM_EXCLUDED_MISSIONS": "mvm_nowhere"}))
	if err == nil || !strings.Contains(err.Error(), "mvm_nowhere") {
		t.Fatalf("err = %v", err)
	}
	_, err = build(data(t), "0.6.7", env(map[string]string{"MVM_START_CLASS": "Pyro"}))
	if err == nil || !strings.Contains(err.Error(), "Heavy, Scout") {
		t.Fatalf("err = %v", err)
	}
}
