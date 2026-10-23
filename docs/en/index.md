# Mann vs Archipelago

This project turns a Team Fortress 2 Mann vs Machine server into an
[Archipelago](https://archipelago.gg) randomizer.

- The classes, the weapon slots and the missions start locked.
- Every wave your team clears is a check. Checks unlock things, here or in
  another player's game.
- Everybody on the server shares the same unlocks.
- Bots fill the empty seats on RED, so two people can play waves that Valve
  balanced for six.
- Your friends install nothing. They join with a normal Team Fortress 2 client.

## The three steps

Every run is the same three steps. The book follows them in order.

1. **Install the server.** One file on
   [Windows](setup/install-windows.md) or [Linux](setup/install-linux.md), or a
   [Docker](setup/install.md) stack. The first start downloads about 14 GB of
   game files.
2. **Create the Archipelago session.** Choose the [run options](setup/shape-of-the-run.md),
   generate a seed, upload it to `archipelago.gg`, and give the server the room
   address. See [Create the session](setup/create-the-session.md).
3. **Invite your friends.** Give them a connect line. See
   [Invite your friends](setup/invite-your-friends.md).

Then read [The first session](play/first-session.md) so you know what the
first evening looks like. [The launcher, tab by tab](setup/the-launcher.md)
explains every screen, and [the campaign tracker](play/tracker.md) shows the
run to your players.

## Quick start on Windows

1. Download `tf2ap.exe` from the
   [latest release](https://github.com/m-this/tf2-archipelago/releases/latest).
2. Run it. Windows warns about it. Click **More info**, then **Run anyway**.
   The warning is a false positive; see [Install on Windows](setup/install-windows.md).
3. Press **Start**. Wait for the download to finish.
4. Install the [Archipelago app](https://github.com/ArchipelagoMW/Archipelago/releases).
5. In the launcher, open **Settings**, then **Player options**, and press
   **Generate seed**.
6. Upload the generated file at [archipelago.gg/uploads](https://archipelago.gg/uploads)
   and create a room.
7. Paste the room address into **Settings**, then **Archipelago room**, and
   press **Restart**.
8. Send your friends the connect line shown under the buttons.

## New to Archipelago?

Read [Archipelago words, for MvM players](archipelago-for-mvm-players.md)
first. It is one page. Archipelago and Mann vs Machine use the same words for
different things, and the rest of the book assumes you know which is which.

## Where to get help

- [Troubleshooting](operate/troubleshooting.md) finds which part is broken.
- **Debug logs**, in the launcher's Settings, writes one file with everything
  a helper needs. Send that file when you ask for help.
- Issues go to [GitHub](https://github.com/m-this/tf2-archipelago/issues).
