# Mann vs Archipelago

This project turns a Team Fortress 2 Mann vs Machine server into a
randomizer. The classes, the weapon slots and the missions start locked. The
team clears waves to unlock them. Everybody on the server shares the
unlocks.

The server also fills the RED team with bots, so two people can win a run
that Valve balanced for six.

## Get started

On Windows, download `tf2ap.exe` from the
[latest release](https://github.com/m-this/tf2-archipelago/releases/latest)
and run it. On Linux, `tf2ap-linux-amd64` from the same release does the same
in a terminal. Either asks for your Archipelago room address and installs the
rest.

With Docker:

```sh
cp deploy/.env.example .env   # then set SRCDS_RCONPW
make seed                     # upload the file to archipelago.gg, open a room
                               # then set AP_HOST and AP_PORT in .env
make up
make logs
```

The first start downloads about 14 GB of game files.

[Install on Windows](setup/install-windows.md),
[Install on Linux](setup/install-linux.md) and
[Install with Docker](setup/install.md) cover all three in full.

## Read the book in this order

This book is for the host. It assumes you know Mann vs Machine and have
never used a randomizer before, and it defines every word before it uses it.

1. [Archipelago for MvM players](archipelago-for-mvm-players.md) — the
   vocabulary. Read this first.
2. [What the randomizer changes](what-the-randomizer-changes.md) — what
   differs from a normal MvM server.
3. [Requirements](setup/requirements.md) — what the machine needs.
4. [The shape of the run](setup/shape-of-the-run.md) — the length and
   difficulty of an evening.
5. [Create the session](setup/create-the-session.md) — makes the run and
   puts it on `archipelago.gg`.
6. [Install on Windows](setup/install-windows.md),
   [Install on Linux](setup/install-linux.md) or
   [Install with Docker](setup/install.md) — gets the server running.
7. [Invite your friends](setup/invite-your-friends.md) opens the server.
   [The bots on your team](play/defender-bots.md) says who fills the empty
   slots.
