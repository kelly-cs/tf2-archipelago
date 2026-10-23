package gamedata

import "testing"

// A wave counted the way the game spawns it: a squad cycles its members up
// to TotalCount, a Support spawn that runs to the end of the wave has no
// count to give, and a gatebot's MiniBoss sits inside EventChangeAttributes.
func TestCountWaveKillsReadsTheStockSyntax(t *testing.T) {
	body := []byte(`#base robot_giant.pop
WaveSchedule
{
	Templates
	{
		T_Local_Boss { Template T_Base_Giant }
		T_Gate_Giant { EventChangeAttributes { Default { Attributes MiniBoss } } }
	}
	Wave
	{
		WaveSpawn { TotalCount 5 Support 1 TFBot { Template T_Base_Giant } }
		WaveSpawn { TotalCount 10 Squad { TFBot { Template T_Base_Giant } TFBot { Class Scout } TFBot { Template T_Local_Boss } TFBot { Class Scout } TFBot { Class Scout } } }
		WaveSpawn { TotalCount 3 Support Limited TFBot { Template T_Gate_Giant } }
		WaveSpawn { TotalCount 2 Tank { Health 20000 } }
		WaveSpawn { TFBot { Attributes MiniBoss } }
	}
	Wave
	{
		WaveSpawn { TotalCount 4 RandomChoice { TFBot { Template T_Base_Giant } TFBot { Class Heavy } } }
	}
}`)
	resolve := func(name string) []byte {
		if name == "robot_giant.pop" {
			return []byte(`WaveSchedule { Templates { T_Base_Giant { Attributes MiniBoss } } }`)
		}
		return nil
	}
	waves, err := CountWaveKills(body, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if len(waves) != 2 {
		t.Fatalf("%d waves", len(waves))
	}
	// Squad of five over ten: giants at positions 0 and 2 of each pass, four.
	// Plus three limited gatebots and one inline: eight. Two tanks.
	if waves[0].Giants != 8 || waves[0].Tanks != 2 || len(waves[0].Uncertain) != 0 {
		t.Fatalf("wave 1 = %+v", waves[0])
	}
	if waves[1].Giants != 0 || len(waves[1].Uncertain) != 1 {
		t.Fatalf("a random draw was counted as a check: %+v", waves[1])
	}
}

// The anchors of the committed table, against Cowser's hand count of the
// Valve waves (gh-68): the largest wave, the most tanks, and a wave the sheet
// got wrong, where a squad of regular gatebot heavies was taken for giants.
func TestTheCommittedWaveCountsHoldTheirAnchors(t *testing.T) {
	for _, tc := range []struct {
		popFile      string
		wave, giants uint8
		tanks        uint8
	}{
		{"mvm_ghost_town_666", 1, 80, 9},
		{"mvm_mannhattan_advanced1", 5, 32, 0},
		{"mvm_coaltown_expert1", 7, 4, 6},
		{"mvm_mannhattan_advanced2", 3, 15, 0},
		{"mvm_decoy", 1, 0, 0},
	} {
		m, ok := MissionByPopFile(tc.popFile)
		if !ok {
			t.Fatalf("no mission %s", tc.popFile)
		}
		if got := m.WaveKillsAt(tc.wave); got.Giants != tc.giants || got.Tanks != tc.tanks {
			t.Errorf("%s wave %d = %+v, want %d giants and %d tanks", tc.popFile, tc.wave, got, tc.giants, tc.tanks)
		}
	}
	if got := Missions[len(Missions)-1].WaveKillsAt(1); got.Giants != 0 || got.Tanks != 0 {
		t.Errorf("a community mission has per-kill counts nobody scrubbed: %+v", got)
	}
}
