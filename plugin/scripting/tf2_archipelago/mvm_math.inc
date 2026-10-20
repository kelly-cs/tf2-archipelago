/** Native-free upgrade station decisions, also executed by the standalone VM tests. */

enum StationCommand
{
    StationCommand_Other,
    StationCommand_Respec,
    StationCommand_Upgrade,
}

// Whether the game rewrites the balance from its own record of the wave, which
// never held a bundle credit, so the bundle ledger has to go back on top of it.
// A respec does. An upgrade does not, in either direction: the minus button pays
// one upgrade's cost back onto the balance the player holds, bundles included,
// and putting the ledger back there paid every bundle again on each press.
bool MvM_StationRestoresRecord(StationCommand command)
{
    return command == StationCommand_Respec;
}

int MvM_CreditsOwed(int held, int ledger)
{
    int owed = held + ledger;
    return owed > 0 ? owed : 0;
}
