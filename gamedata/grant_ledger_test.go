package gamedata

import (
	"slices"
	"testing"
)

/*
The grant cursor and the bundle ledger, run whole.

grant_ledger.inc holds the plugin's only memory of what it applied and what it
paid, and the callbacks in bridge.inc and mvm.inc are adapters around it. So
these drivers play the bridge and the game: they hand it grants in the order a
poll would, write balances the way the upgrade station does, and read back what
the plugin would have done. A step the plugin takes is a GrantCursor_ call, the
same one the callback makes.

receive is the poll loop of OnGrants, with the effect applied or held as the
game allows and the state grants counted as levels.
*/
const grantDriver = `
int levels;
int announced;
bool canPay;

void receive(int seq, bool effect)
{
    GrantStep step = GrantCursor_Receive(seq, effect);
    if (step == GrantStep_Apply)
    {
        levels++;
    }
    else if (step == GrantStep_Try && GrantCursor_Settle(seq, canPay))
    {
        announced++;
    }
}
`

func runGrants(t *testing.T, body string) []int32 {
	t.Helper()
	return driver{prelude: grantDriver, body: body}.run(t)
}

/*
A held Cash Bundle makes the bridge resend every buff behind it, on every poll,
until the station opens. Each copy is a level, so a replay that counted again
took a legitimate x2 to x44 (#168). The unlock snapshot holds items 1 to 5
(two buffs, the bundle at 3, two more buffs) and the bridge was acknowledged up
to 2.
*/
func TestHeldCashDoesNotReplayBuffs(t *testing.T) {
	got := runGrants(t, `
    GrantCursor_Resume(2, 5);
    for (int poll = 0; poll < 21; poll++)
    {
        receive(3, true);
        receive(4, false);
        receive(5, false);
    }
    printnum(levels);
    printnum(announced);
    printnum(GrantCursor_Sequence());
    printnum(view_as<int>(GrantCursor_Holding()));

    // A new buff arrives while the bundle is still held: one level, once.
    for (int poll = 0; poll < 21; poll++)
    {
        receive(3, true);
        receive(4, false);
        receive(5, false);
        receive(6, false);
    }
    printnum(levels);
    printnum(GrantCursor_Sequence());

    // The station opens. The bundle pays, the cursor runs to the end.
    canPay = true;
    receive(3, true);
    receive(4, false);
    receive(5, false);
    receive(6, false);
    printnum(levels);
    printnum(GrantCursor_Sequence());
    printnum(view_as<int>(GrantCursor_Holding()));
    printnum(GrantCursor_AckDue());
`)
	want := []int32{0, 1, 2, 1, 1, 2, 1, 6, 0, 6}
	if !slices.Equal(got, want) {
		t.Fatalf("held cash with replayed buffs = %v, want %v", got, want)
	}
}

// A reload asks from an acknowledgement older than the snapshot even in a run
// with no cash or traps. Those buffs are already in the menu.
func TestTheSnapshotOverlapAddsNoLevel(t *testing.T) {
	got := runGrants(t, `
    GrantCursor_Resume(0, 2);
    receive(1, false);
    receive(2, false);
    receive(1, false);
    printnum(levels);
    printnum(GrantCursor_Sequence());
    printnum(GrantCursor_AckDue());
`)
	want := []int32{0, 2, 2}
	if !slices.Equal(got, want) {
		t.Fatalf("snapshot overlap = %v, want %v", got, want)
	}
}

// Effects are ordered: a trap behind a held bundle waits for it, even once the
// trap could fire, because the cursor cannot describe a gap.
func TestAnEffectWaitsBehindTheHeldOne(t *testing.T) {
	got := runGrants(t, `
    GrantCursor_Resume(0, 0);
    receive(1, true);
    canPay = true;
    receive(2, true);
    printnum(GrantCursor_Sequence());
    printnum(GrantCursor_AckDue());
    receive(1, true);
    receive(2, true);
    printnum(GrantCursor_Sequence());
    printnum(announced);
`)
	want := []int32{0, 0, 2, 1}
	if !slices.Equal(got, want) {
		t.Fatalf("ordered effects = %v, want %v", got, want)
	}
}

// An effect at or below the cursor was paid. The bridge stops sending it once
// acknowledged, and the cursor covers the round trip before that.
func TestAnAppliedEffectIsNeverPaidTwice(t *testing.T) {
	got := runGrants(t, `
    canPay = true;
    GrantCursor_Resume(0, 0);
    int paid;
    for (int poll = 0; poll < 3; poll++)
    {
        if (GrantCursor_Receive(1, true) == GrantStep_Try)
        {
            GrantCursor_Settle(1, true);
            paid++;
        }
    }
    printnum(paid);
    printnum(GrantCursor_AckDue());
    printnum(view_as<int>(GrantCursor_Acknowledged(1)));
    printnum(GrantCursor_AckDue());
`)
	want := []int32{1, 1, 1, 0}
	if !slices.Equal(got, want) {
		t.Fatalf("applied effect = %v, want %v", got, want)
	}
}

/*
A bridge answering behind the plugin restarted its run, whether behind the
effect cursor or behind the snapshot. The cursor starts again from nothing, and
an acknowledgement still in flight from the old run cannot lift it.
*/
func TestARestartedRunStartsTheCursorAgain(t *testing.T) {
	got := runGrants(t, `
    canPay = true;
    GrantCursor_Resume(4, 9);
    printnum(view_as<int>(GrantCursor_RunRestarted(9)));
    printnum(view_as<int>(GrantCursor_RunRestarted(8)));
    printnum(GrantCursor_Sequence());
    printnum(view_as<int>(GrantCursor_Acknowledged(4)));
    printnum(GrantCursor_AckDue());

    GrantCursor_Resume(1, 1);
    receive(2, true);
    printnum(view_as<int>(GrantCursor_Holding()));
    GrantCursor_Restart();
    printnum(view_as<int>(GrantCursor_Holding()));
    printnum(GrantCursor_Sequence());
`)
	want := []int32{0, 1, 0, 0, 0, 0, 0, 0}
	if !slices.Equal(got, want) {
		t.Fatalf("restarted run = %v, want %v", got, want)
	}
}

/*
Buying an upgrade and taking it back leaves the balance where it was.

The game charges a purchase against the balance the player holds and pays a
minus press back onto it, bundles included. Only a respec, and a lost wave,
rewrite the balance from the game's own record of the wave, which never held a
bundle credit. Putting the ledger back after a minus press paid every bundle of
the mission again on each press (#98): five buy-and-refund cycles grew 1200 to
5200.

station plays the game's side: 400 recorded at wave start, 800 of bundles paid
on top, and a restore that runs a frame after the command, as the plugin's does.
*/
const stationDriver = `
int held;

void press(int client, StationCommand command, int cost)
{
    held += cost;
    if (command == StationCommand_Respec)
    {
        held = 400;
    }
    if (BundleLedger_Station(command))
    {
        held = BundleLedger_Restore(client, held);
    }
}
`

func TestTheMinusButtonDoesNotPayBundlesAgain(t *testing.T) {
	got := driver{prelude: stationDriver, body: `
    BundleLedger_Clear();
    held = 400;
    BundleLedger_Pay(1, 500);
    BundleLedger_Pay(1, 300);
    held += 800;
    for (int cycle = 0; cycle < 5; cycle++)
    {
        press(1, StationCommand_Upgrade, -100);
        press(1, StationCommand_Upgrade, 100);
    }
    printnum(held);

    press(1, StationCommand_Respec, 0);
    printnum(held);

    press(1, StationCommand_Other, 0);
    printnum(held);
`}.run(t)
	want := []int32{1200, 1200, 1200}
	if !slices.Equal(got, want) {
		t.Fatalf("station balances = %v, want %v", got, want)
	}
}

// A lost wave can restore a record below what was spent, and the balance comes
// back at zero rather than negative. A reused client index and a new map start
// from an empty ledger.
func TestTheLedgerIsPerClientAndPerMap(t *testing.T) {
	got := driver{body: `
    BundleLedger_Clear();
    BundleLedger_Pay(1, 200);
    BundleLedger_Pay(2, 700);
    printnum(BundleLedger_Restore(1, -300));
    printnum(BundleLedger_Restore(2, 100));
    BundleLedger_Forget(2);
    printnum(BundleLedger_Credits(1));
    printnum(BundleLedger_Credits(2));
    printnum(BundleLedger_Credits(-1));
    printnum(BundleLedger_Credits(BundleLedgerClients));
    BundleLedger_Clear();
    printnum(BundleLedger_Credits(1));
`}.run(t)
	want := []int32{0, 800, 200, 0, 0, 0, 0}
	if !slices.Equal(got, want) {
		t.Fatalf("ledger = %v, want %v", got, want)
	}
}
