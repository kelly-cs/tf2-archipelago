package gamedata

import (
	"slices"
	"testing"
)

/*
Buying an upgrade and taking it back leaves the balance where it was.

The game charges a purchase against the balance the player holds and pays a
minus press back onto it, bundles included. Only a respec, and a lost wave,
rewrite the balance from the game's own record of the wave, which never held a
bundle credit. Putting the bundle ledger back after a minus press paid every
bundle of the mission again on each press (GitHub #98).

The game's side is simulated here: 400 recorded at wave start, 800 of bundles
written on top, an upgrade that costs 100.
*/
func TestTheUpgradeStationPaysBundlesBackOnlyOnARestore(t *testing.T) {
	got := driver{body: `
    int ledger = 800;
    int held = 1200;
    for (int press = 0; press < 5; press++)
    {
        held -= 100;
        if (MvM_StationRestoresRecord(StationCommand_Upgrade)) held = MvM_CreditsOwed(held, ledger);
        held += 100;
        if (MvM_StationRestoresRecord(StationCommand_Upgrade)) held = MvM_CreditsOwed(held, ledger);
    }
    printnum(held);

    held = 400;
    if (MvM_StationRestoresRecord(StationCommand_Respec)) held = MvM_CreditsOwed(held, ledger);
    printnum(held);

    printnum(MvM_StationRestoresRecord(StationCommand_Other));
    printnum(MvM_CreditsOwed(-300, 200));
`}.run(t)
	want := []int32{1200, 1200, 0, 0}
	if !slices.Equal(got, want) {
		t.Fatalf("station balances = %v, want %v", got, want)
	}
}
