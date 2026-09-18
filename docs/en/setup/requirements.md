# Requirements

## Pick a way to run the server

| Way | Who it is for | Page |
| --- | --- | --- |
| **Windows launcher** | Most people. One exe, nothing else. | [Install on Windows](install-windows.md) |
| **Linux launcher** | The same program, on a Linux machine or over SSH. | [Install on Linux](install-linux.md) |
| **Docker** | A machine that already runs Docker stacks. | [Install with Docker](install.md) |

All three run the same software and have the same settings.

## The machine

| Thing | What you need |
| --- | --- |
| Disk | About 20 GB free. The game server is about 14 GB and downloads once. |
| Memory | 4 GB. |
| Processor | Two cores. |
| Network | Outgoing access to `archipelago.gg`. Nothing to open on the router unless you choose the forwarded-port route. |

## What the host also needs

- The official [Archipelago app](https://github.com/ArchipelagoMW/Archipelago/releases),
  to generate the seed. See [Create the session](create-the-session.md).
- A Steam game server login token, if friends join over the internet. See
  [Invite your friends](invite-your-friends.md). Playing on the local network
  needs none.

## What you do not need

- No Steam account for the server.
- No Team Fortress 2 install on the host. The server downloads its own files.
- No account on `archipelago.gg`.
- Nothing for the players. A normal Team Fortress 2 client is enough.

## Security note

The game server is a large C++ program that reads network traffic from anybody
who knows the address. Run it on a machine where that is acceptable.

Next: [Install on Windows](install-windows.md), [Install on Linux](install-linux.md)
or [Install with Docker](install.md).
