/**
 * What the plugin has applied of the bridge's grants, and what it paid on top of
 * the game's own record of the wave. Both are money paid once (CONTEXT.md,
 * "State grant / effect grant", "Acknowledgement").
 *
 * No natives, no includes and no SourceMod constant, so the standalone VM runs
 * it. The callbacks in bridge.inc and mvm.inc are adapters: JSON and entity
 * properties in, these calls, then SetEntProp and the acknowledgement out.
 * gamedata/grant_ledger_test.go crosses the same calls with whole scenarios.
 */

// --- Grant cursor ---

enum GrantStep
{
    // At or below the cursor: the bridge resent it before the acknowledgement landed.
    GrantStep_Applied,
    // Counted already, or an effect queued behind the held one: do nothing.
    GrantStep_Skip,
    // A state grant counted now: apply it. It cannot fail to apply.
    GrantStep_Apply,
    // The next effect in line: try it, then report with GrantCursor_Settle.
    GrantStep_Try,
}

// How far every effect has been applied. Polls ask from here and the
// acknowledgement reports it, so it never steps over an unapplied effect.
static int g_GrantSeq;

// State grants already represented by the unlock snapshot or applied from a
// poll. Advances past a held effect without moving the effect cursor, so the
// buffs a held effect keeps resending add no level (#168).
static int g_StateGrantSeq;

// The effect being held, zero for none. Not acknowledged, so the bridge keeps
// sending it, which is also what carries it across a plugin reload.
static int g_HeldSeq;

// How far the bridge has been told the plugin got. Effects at or below it are
// never sent again, which is what stops a reload paying the run's cash twice.
static int g_AckedSeq;

// The unlock set landed. resumeFrom is where the bridge was acknowledged up to,
// snapshotSeq how far the snapshot already counted every state grant.
void GrantCursor_Resume(int resumeFrom, int snapshotSeq)
{
    g_GrantSeq = resumeFrom;
    g_StateGrantSeq = snapshotSeq;
}

// The run changed under the plugin. Everything it counted belonged to another.
void GrantCursor_Restart()
{
    g_GrantSeq = 0;
    g_AckedSeq = 0;
    g_HeldSeq = 0;
    g_StateGrantSeq = 0;
}

// A bridge answering behind the plugin restarted its run. Restarts the cursor
// and says so, so the caller asks for the unlock set again.
bool GrantCursor_RunRestarted(int bridgeSeq)
{
    if (bridgeSeq >= g_GrantSeq && bridgeSeq >= g_StateGrantSeq)
    {
        return false;
    }
    GrantCursor_Restart();
    return true;
}

int GrantCursor_Sequence()
{
    return g_GrantSeq;
}

bool GrantCursor_Holding()
{
    return g_HeldSeq != 0;
}

/* What to do with the grant at seq. effect is true for credits and traps.
 *
 * A state grant is counted here and the effect cursor follows it unless an
 * effect is held. An effect is strictly ordered: one behind the held effect
 * waits, because the cursor is a single number and cannot describe a gap.
 */
GrantStep GrantCursor_Receive(int seq, bool effect)
{
    if (seq <= g_GrantSeq)
    {
        return GrantStep_Applied;
    }
    if (effect)
    {
        return (g_HeldSeq == 0 || g_HeldSeq == seq) ? GrantStep_Try : GrantStep_Skip;
    }
    bool counted = seq <= g_StateGrantSeq;
    if (!counted)
    {
        g_StateGrantSeq = seq;
    }
    if (g_HeldSeq == 0)
    {
        g_GrantSeq = seq;
    }
    return counted ? GrantStep_Skip : GrantStep_Apply;
}

// The effect at seq was tried. True when it has just started waiting, which is
// the one time to say so: the bridge sends it again on every poll.
bool GrantCursor_Settle(int seq, bool applied)
{
    if (applied)
    {
        g_HeldSeq = 0;
        g_GrantSeq = seq;
        return false;
    }
    if (g_HeldSeq == seq)
    {
        return false;
    }
    g_HeldSeq = seq;
    return true;
}

// The sequence to acknowledge, zero when the bridge already has it.
int GrantCursor_AckDue()
{
    return g_GrantSeq > g_AckedSeq ? g_GrantSeq : 0;
}

// The bridge took the acknowledgement of seq. False when it was sent by a run
// that has since restarted: taking it would put the cursor above where the
// plugin now is and stop it acknowledging anything for the rest of the run.
bool GrantCursor_Acknowledged(int seq)
{
    if (seq > g_GrantSeq)
    {
        return false;
    }
    if (seq > g_AckedSeq)
    {
        g_AckedSeq = seq;
    }
    return true;
}

// --- Bundle ledger ---

// MAXPLAYERS + 1 in SourceMod; tf2_archipelago.sp refuses to compile otherwise.
#define BundleLedgerClients 102

enum StationCommand
{
    StationCommand_Other,
    StationCommand_Respec,
    StationCommand_Upgrade,
}

/* Everything a bundle ever added to each client.
 *
 * A bundle written straight onto m_nCurrency is money the game's record of the
 * wave never saw. A respec or a lost wave rewrites the balance from that record,
 * so the bundles vanish unless they go back on top. The ledger only grows, and
 * it is cleared on a map change, where the game clears its own side too.
 */
static int g_BundleCredits[BundleLedgerClients];

// What the balance lacks since the game last restored its record: the whole
// ledger then, nothing once BundleLedger_Restore put it back. A second restore
// before the frame runs lacks the same amount, not twice it.
static int g_BundleMissing[BundleLedgerClients];

void BundleLedger_Pay(int client, int amount)
{
    g_BundleCredits[client] += amount;
}

// Read across the plugin boundary by the defender bot mod, hence the range check.
int BundleLedger_Credits(int client)
{
    return (client >= 0 && client < BundleLedgerClients) ? g_BundleCredits[client] : 0;
}

// Client indexes are reused, and so is the ledger behind them.
void BundleLedger_Forget(int client)
{
    g_BundleCredits[client] = 0;
    g_BundleMissing[client] = 0;
}

void BundleLedger_Clear()
{
    for (int client = 0; client < BundleLedgerClients; client++)
    {
        BundleLedger_Forget(client);
    }
}

/* The upgrade station sent command. True when the game rewrote the balance from
 * its record, so BundleLedger_Restore has to run once the game has written it.
 * Every call that says so asks for a restore; only one of them pays.
 *
 * A respec does. An upgrade does not, in either direction: the minus button pays
 * one upgrade's cost back onto the balance the player holds, bundles included,
 * and putting the ledger back there paid every bundle again on each press (#98).
 */
bool BundleLedger_Station(int client, StationCommand command)
{
    if (command != StationCommand_Respec)
    {
        return false;
    }
    g_BundleMissing[client] = g_BundleCredits[client];
    return true;
}

// A lost wave restores the same record a respec does, to everyone.
void BundleLedger_WaveLost(int client)
{
    g_BundleMissing[client] = g_BundleCredits[client];
}

// The balance to write after the game restored its record to held, and held
// itself when nothing is missing. Never below zero: bundles spent on top of a
// record that never held them run past it.
int BundleLedger_Restore(int client, int held)
{
    int owed = held + g_BundleMissing[client];
    g_BundleMissing[client] = 0;
    return owed > 0 ? owed : 0;
}
