/** Native-free mission decisions, also executed by the standalone VM tests. */

// SigMod refuses a tf_mvm_popfile its map's loose files do not list until ten
// seconds of game time have passed since its last mission vote, and a map starts
// with that time at zero. Half a second over, for the timer's granularity.
#define SigModVoteGuardEnd 10.5

// How long a tf_mvm_popfile sent at gameTime has to wait to be read at all.
float Missions_PopFileWait(float gameTime, bool voteGuard)
{
    if (!voteGuard || gameTime >= SigModVoteGuardEnd)
    {
        return 0.0;
    }
    return SigModVoteGuardEnd - gameTime;
}
