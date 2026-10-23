/** Native-free modifier decisions, also executed by the standalone VM tests. */
int Chamber_Quarter(int amount)
{
    if (amount <= 0) return amount;
    int limit = amount / 4;
    return limit > 0 ? limit : 1;
}

int Chamber_ReserveLimit(int amount, int ammoType)
{
    return ammoType <= 0 ? amount : Chamber_Quarter(amount);
}

bool Chamber_ShouldSupply(float now, float nextSupply)
{
    return now >= nextSupply;
}

bool Mental_CanRecloak(float now, float nextCloak, bool stealthed)
{
    return !stealthed && now >= nextCloak;
}

float Mental_RecloakProgress(float now, float fadeStart)
{
    float progress = now - fadeStart;
    if (progress <= 0.0) return 0.0;
    if (progress >= 1.0) return 1.0;
    // Ease the visible-to-hidden transition at both ends.
    return progress * progress * (3.0 - 2.0 * progress);
}

float Mental_EntityVisibility(float now, float shimmerUntil, float revealUntil)
{
    if (now < revealUntil) return 1.0;
    if (now < shimmerUntil) return 0.15;
    if (revealUntil > 0.0 && now < revealUntil + 1.0)
        return 1.0 - Mental_RecloakProgress(now, revealUntil);
    return 0.0;
}
