package form

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/*
The tier menu says how many missions each floor draws from, and the mission
count's ceiling is that number. With SigMod ticked and not yet installed the two
counted different pools: the tiers left out the SigMod missions the ceiling let
a player ask for.
*/
func TestTheTierMenuAndTheCeilingCountOnePool(t *testing.T) {
	state := NewState(settings.Defaults())
	state.Settings.SrcdsMods = []string{"sigsegv-mvm"}
	state.Settings.CommunityPacks = []string{settings.CommunityPackPotato}
	state.Settings.MvmExcludedMissions = slices.DeleteFunc(state.Settings.MvmExcludedMissions,
		func(popFile string) bool { return popFile == "mvm_bronx_rc2_adv_point_of_impact" })
	env := Env{CommunityAvailable: []string{settings.CommunityPackPotato}, Platform: "linux"}
	model := Build(state, env)

	count, ok := model.Field("run.mission_count")
	if !ok {
		t.Fatal("no mission count row")
	}
	tier, ok := model.Field("run.tier")
	if !ok {
		t.Fatal("no tier row")
	}
	for _, option := range tier.Options {
		if option.Value != state.Settings.MvmDifficulty {
			continue
		}
		if want := "draws from " + strconv.Itoa(count.High) + " missions"; !strings.Contains(strings.Join(strings.Fields(option.Label), " "), want) {
			t.Fatalf("tier %q says %q, the ceiling is %d", option.Value, option.Label, count.High)
		}
		return
	}
	t.Fatalf("no tier %q on the menu", state.Settings.MvmDifficulty)
}

/*
Start installs the ticked packs and no other. Unticking one used to leave its
missions in the pool and the ceiling, so the seed drew a mission the server
never got and the bridge refused it on connect.
*/
func TestUntickingAPackTakesItsMissionsOutOfThePool(t *testing.T) {
	const homestead = "mvm_kelly_rc1b_adv_homestead_happenings"
	state := NewState(settings.Defaults())
	state.Settings.CommunityPacks = []string{settings.CommunityPackPotato}
	state.Settings.MvmExcludedMissions = nil
	env := Env{CommunityAvailable: []string{settings.CommunityPackPotato}, Platform: "linux"}

	next, err := Apply(state, env, Change{Field: "missions.potato", Value: "false"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(next.Settings.MvmExcludedMissions, homestead) {
		t.Error("a mission of the unticked pack stayed in the pool")
	}
	if slices.Contains(next.Settings.MvmExcludedMissions, "mvm_decoy") {
		t.Error("unticking a pack took a Valve mission out of the pool")
	}
	field, _ := Build(next, env).Field("missions.pool." + homestead)
	if !field.Disabled || field.Reason != "tick Potato Archive" {
		t.Errorf("the unticked pack's mission row = %+v", field)
	}
}
