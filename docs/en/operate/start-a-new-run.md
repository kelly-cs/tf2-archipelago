# Start a new run

A seed never changes once it exists. A different run needs a new seed and a
new room. The game files are not touched, so it takes a few minutes.

## With the launcher

1. Change the [run options](../setup/shape-of-the-run.md) in **Settings**, if
   you want a different run.
2. Press **Generate seed** on the **Player options** page.
3. Upload the new file at [archipelago.gg/uploads](https://archipelago.gg/uploads)
   and create a room.
4. Paste the new room address into **Settings**, then **Archipelago room**.
5. Save, then press **Restart**.

## With Docker

1. Edit the `MVM_` lines in `.env`, if you want a different run.
2. Run `make seed`. It writes another file into `seed/`.
3. Upload that file and create a room.
4. Set `AP_PORT` to the port of the new room.
5. Run `make restart`.

Keep the old files in `seed/`. Each one is a whole run, and the room of a run
comes back from its file.

### If you host the session yourself

With `COMPOSE_PROFILES=selfhost`, the seed lives in the
`tf2-archipelago_apoutput` volume. A new run is:

```sh
make down
docker volume rm tf2-archipelago_apoutput
make up
```

Edit `.env` between the first and the third command.

## What happens to the old run

The bridge notices that the room is not the one it holds state for. Then it:

1. moves its state file aside, to `bridge.<seed>.json` in the same folder;
2. starts over with no checks and no unlocks;
3. tells the plugin that the run restarted.

Nothing is overwritten. If you point the server at the wrong room by mistake,
the old run is still on disk.

Nothing else drops a run. Restarting the server, restarting the machine and
stopping for a week all keep it.

## Starting completely over

- **Launcher:** **Reset settings** in **Settings** puts every setting back to
  the defaults and keeps the game files.
- **Docker:** `make clean` stops the stack and deletes every volume, including
  the 14 GB of game files. Use it when you are done with the project, not
  between runs.
