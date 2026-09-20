package main

import "testing"

func TestMissionRowResolvesIncludesOnlyFromSelectedArchive(t *testing.T) {
	popFile := "mvm_bronx_rc2_adv_point_of_impact"
	mapIDs := map[string]uint8{"mvm_bronx_rc2": 150}
	nav := map[string]bool{"mvm_bronx_rc2": true}
	body := []byte("#base shared.pop\nWaveSchedule { Wave { } }")
	all := map[string]map[string][][]byte{
		"archive-assets.zip":   {"shared.pop": {[]byte("LuaScriptFile example.lua")}},
		"mlarchive-assets.zip": {"shared.pop": {[]byte("WaveSchedule { }")}},
	}
	_, requirement, ok, err := missionRow(popFile, mapIDs, nav, population{body: body, pack: "mlarchive-assets.zip"}, all, 200, map[string]bool{})
	if err != nil || !ok || requirement != "" {
		t.Fatalf("Moonlight row inherited Potato requirement: %q, ok=%t, err=%v", requirement, ok, err)
	}
	_, requirement, ok, err = missionRow(popFile, mapIDs, nav, population{body: body, pack: "archive-assets.zip"}, all, 200, map[string]bool{})
	if err != nil || !ok || requirement != "sigsegv-mvm" {
		t.Fatalf("Potato row lost its requirement: %q, ok=%t, err=%v", requirement, ok, err)
	}
}

func TestMissionIdentityUsesWholeFilenameSegments(t *testing.T) {
	tests := []struct {
		popFile, mapName, difficulty, title string
	}{
		{"mvm_bronx_rc2_adv_point_of_impact", "mvm_bronx_rc2", "advanced", "Point Of Impact"},
		{"mvm_downpour_rc3a_adv_666_last_stand", "mvm_downpour_rc3a", "haunted", "Last Stand"},
		{"mvm_autumnull_rc2_rev_exp_codename_omega", "mvm_autumnull_rc2", "expert", "Codename Omega"},
		{"mvm_cyberia_rc6a_rev_arctic_arrangement", "mvm_cyberia_rc6a", "advanced", "Arctic Arrangement"},
	}
	for _, test := range tests {
		difficulty, title, ok := missionIdentity(test.popFile, test.mapName)
		if !ok || difficulty != test.difficulty || title != test.title {
			t.Errorf("missionIdentity(%q) = %q, %q, %t; want %q, %q, true",
				test.popFile, difficulty, title, ok, test.difficulty, test.title)
		}
	}
}
