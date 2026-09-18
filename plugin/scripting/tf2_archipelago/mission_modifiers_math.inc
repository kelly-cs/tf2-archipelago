/** Native-free Mental decisions, also executed by the standalone VM tests. */
bool Mental_CanRecloak(float now, float nextCloak, bool stealthed)
{
    return !stealthed && now >= nextCloak;
}

float Mental_RecloakProgress(float now, float fadeStart)
{
    float progress = now - fadeStart;
    if (progress <= 0.0) return 0.0;
    if (progress >= 1.0) return 1.0;
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
