package main

import "testing"

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
