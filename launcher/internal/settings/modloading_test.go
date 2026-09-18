package settings

import "testing"

func TestServerModsToLoadAcrossTheThreeAnswers(t *testing.T) {
	base := Defaults()
	base.SrcdsMods = []string{"sigsegv-mvm"}
	base.MvmCommunityMissions = false

	for _, tc := range []struct {
		name    string
		loading ModLoading
		start   string
		want    bool
	}{
		{"required, no mission asks", ModLoadingRequired, "mvm_decoy", false},
		{"required, the start mission asks", ModLoadingRequired, "mvm_bronx_rc2_adv_point_of_impact", true},
		{"always, no mission asks", ModLoadingAlways, "mvm_decoy", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := base
			s.SrcdsModLoading = tc.loading
			s.MvmStartMission = tc.start
			got := len(ServerModsToLoad(s, "linux")) == 1
			if got != tc.want {
				t.Errorf("loads = %v, want %v (%v)", got, tc.want, ServerModsToLoad(s, "linux"))
			}
		})
	}

	// Off is off, whatever the pool holds.
	off := base
	off.SrcdsMods = nil
	off.MvmStartMission = "mvm_bronx_rc2_adv_point_of_impact"
	if got := ServerModsToLoad(off, "linux"); len(got) != 0 {
		t.Errorf("an unselected mod loads: %v", got)
	}

	// A platform with no build never loads, however it is selected.
	always := base
	always.SrcdsModLoading = ModLoadingAlways
	if got := ServerModsToLoad(always, "darwin"); len(got) != 0 {
		t.Errorf("a mod with no build here loads: %v", got)
	}
}
