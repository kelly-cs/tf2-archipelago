package settings

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/gamedata"
)

// packMissions is every mission of pack a server with every mod could play.
func packMissions(t *testing.T, pack string) []gamedata.Mission {
	t.Helper()
	var missions []gamedata.Mission
	for _, mission := range gamedata.Missions {
		if gamedata.MissionPack(mission.ID) == pack && gamedata.IsPlayableMission(mission.ID) {
			missions = append(missions, mission)
		}
	}
	if len(missions) == 0 {
		t.Fatalf("the catalog has no playable mission from %s", pack)
	}
	return missions
}

// A Docker stack that ticks only Moonlight installs only Moonlight, and the
// seed it generates must not draw a Potato mission: the bridge would refuse it.
func TestAnUntickedPackIsOutOfThePoolAndThePlayerFile(t *testing.T) {
	t.Setenv("TF2AP_COMMUNITY_PACKS", CommunityPackMoonlight)
	t.Setenv("MVM_COMMUNITY_MISSIONS", "true")
	t.Setenv("MVM_EXCLUDED_MISSIONS", "")
	s := ApplyEnv(ComposeDefaults())

	for _, mission := range MissionPool(s).Missions() {
		if gamedata.MissionPack(mission.ID) == CommunityPackPotato {
			t.Fatalf("the pool holds %s from the unticked Potato pack", mission.PopFile)
		}
	}
	excluded := ExcludedMissionNames(s)
	for _, mission := range packMissions(t, CommunityPackPotato) {
		if !slices.Contains(excluded, mission.Name) {
			t.Errorf("the player file leaves %q from the unticked Potato pack drawable", mission.Name)
		}
	}
	for _, mission := range packMissions(t, CommunityPackMoonlight) {
		if slices.Contains(excluded, mission.Name) {
			t.Errorf("the player file excludes %q from the ticked Moonlight pack", mission.Name)
		}
	}
}

/*
A .env written before the seed containers were handed TF2AP_COMMUNITY_PACKS
has no such key, and the server it runs installs every pack in
community-content whatever the admin page ticks. Reading that as "no pack"
would quietly shrink every existing stack to Valve's missions, so a missing key
keeps every pack and only a key that is there and empty means none.
*/
func TestComposePacksTellAMissingKeyFromAnEmptyOne(t *testing.T) {
	cases := []struct {
		name  string
		value *string
		want  []string
	}{
		{"missing", nil, []string{CommunityPackPotato, CommunityPackMoonlight}},
		{"empty", new(""), nil},
		{"one", new(CommunityPackPotato), []string{CommunityPackPotato}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			unsetEnv(t, "TF2AP_COMMUNITY_PACKS")
			if c.value != nil {
				t.Setenv("TF2AP_COMMUNITY_PACKS", *c.value)
			}
			t.Setenv("MVM_COMMUNITY_MISSIONS", "true")
			t.Setenv("MVM_EXCLUDED_MISSIONS", "")
			s := ApplyEnv(ComposeDefaults())
			if !slices.Equal(s.CommunityPacks, c.want) {
				t.Fatalf("packs = %v, want %v", s.CommunityPacks, c.want)
			}

			drawn := map[string]bool{}
			for _, mission := range MissionPool(s).Missions() {
				drawn[gamedata.MissionPack(mission.ID)] = true
			}
			yaml := PlayerYAML(s, "")
			for _, pack := range []string{CommunityPackPotato, CommunityPackMoonlight} {
				ticked := slices.Contains(c.want, pack)
				if drawn[pack] != ticked {
					t.Errorf("pool draws from %s: %t, want %t", pack, drawn[pack], ticked)
				}
				named := strings.Contains(yaml, `"`+packMissions(t, pack)[0].Name+`"`)
				if named == ticked {
					t.Errorf("player file excludes %s: %t, want %t", pack, named, !ticked)
				}
			}
		})
	}
}

// unsetEnv removes name for the rest of the test and puts it back afterwards.
func unsetEnv(t *testing.T, name string) {
	t.Helper()
	t.Setenv(name, "")
	if err := os.Unsetenv(name); err != nil {
		t.Fatal(err)
	}
}
