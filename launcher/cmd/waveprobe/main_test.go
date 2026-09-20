package main

import (
	"testing"

	"github.com/m-this/tf2-archipelago/gamedata"
)

func TestParseStatus(t *testing.T) {
	got, err := parseStatus("[SM] WAVEPROBE state=passed map=mvm_decoy pop=mvm_decoy_advanced3 max=6 gamewave=2 expected=1 observed=1 bots=42 tanks=1 defender=3 defteam=2 playerteam=2 enemyteam=3 elapsed=70.5")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != "passed" || got.Pop != "mvm_decoy_advanced3" || got.Max != 6 ||
		got.GameWave != 2 || got.Expected != 1 || got.Observed != 1 || got.Bots != 42 || got.Tanks != 1 ||
		got.DefTeam != 2 || got.PlayerTeam != 2 || got.EnemyTeam != 3 {
		t.Fatalf("status = %+v", got)
	}
	if _, err := parseStatus("Unknown command sm_waveprobe_status"); err == nil {
		t.Fatal("missing test plugin looked healthy")
	}
}

func TestTimescaleMatches(t *testing.T) {
	for _, reply := range []string{`host_timescale = "10" ( def. "1" )`, `host_timescale = "10.000000"`} {
		if !timescaleMatches(reply, 10) {
			t.Errorf("did not accept %q", reply)
		}
	}
	if timescaleMatches(`host_timescale = "1" ( def. "1" )`, 10) {
		t.Fatal("accepted unchanged clock")
	}
}

func TestShardsPartitionCatalog(t *testing.T) {
	for _, shards := range []int{1, 2, 6, 17} {
		seen := make(map[string]int)
		for shard := range shards {
			missions, err := selectMissions(options{mission: "all", shards: shards, shard: shard, includeSig: true})
			if err != nil {
				t.Fatal(err)
			}
			for _, mission := range missions {
				seen[mission.PopFile]++
			}
		}
		for _, mission := range gamedata.Missions {
			want := 1
			if gamedata.MissionRequirement(mission.ID) == "no_nav" {
				want = 0
			}
			if seen[mission.PopFile] != want {
				t.Errorf("%d shards: %s occurs %d times, want %d", shards, mission.PopFile, seen[mission.PopFile], want)
			}
		}
	}
}
