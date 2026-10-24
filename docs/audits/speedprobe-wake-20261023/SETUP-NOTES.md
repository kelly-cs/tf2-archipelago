# Setup observation

The first cold start installed fresh SourceMod while the image staged the mod
files asynchronously. The engine answered RCON with only the stock plugins;
the Archipelago plugin, SigMod extension and wave probe were not yet loaded.
No speed measurements started in that state. `cold-start.log` preserves it.
The isolated container was restarted once after the pinned files appeared on
disk. This benchmark measures wave runs after readiness, not successful cold
installation. This setup issue remains visible and is not a speed-test pass.
