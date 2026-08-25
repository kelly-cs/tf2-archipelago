# tf2-archipelago

An Archipelago randomiser for TF2 Mann vs Machine. It has an apworld, a bridge,
a SourceMod plugin, and a Windows launcher that does the server setup for you.

- `CONTEXT.md` — read it first. Archipelago and TF2 use the same words for
  different things, and this file defines both.
- `apworld/` — the Archipelago world. The logic is in `rules.py`.
- `bridge/` — Go. It connects the game server to the multiworld.
- `plugin/` — the SourceMod side. `launcher/` — the Windows .exe that players run.

## Build

Use the Makefile. `make help` lists every target.

```bash
make test lint                  # Go
make apworld-test apworld-lint
make seed                       # generate a seed
make up down logs rcon          # start the server on this machine
```

`.github/workflows/release.yml` copies the `CHANGELOG.md` section that matches
the tag into the release notes. Players read it. Keep developer notes out.

## Triage from Discord

All discussion happens in Discord and no bot reads it. The maintainer copies the
chat here. Turn it into beads.

- Read `bd list` first. Comment on the open bead. Do not create a second one.
- Quote the reporter and give their name. Put your reading under the quote.
- If the message does not show the cause, say so and leave the bead a report.
- Add the label `discord`. Drop the messages that report nothing, then say so.
- Choose the repository by the cause, not by the symptom. The wave-loss money
  bug appeared in the mod and belonged here.
- Read the close reason before you reopen a bead. It names the trigger measured.

## Beads

Use `bd prime` to list the commands. The prefix is `apw`. P0 crash, P1 costs a
player a run, P2 bug, P3 polish.

- Git tracks `.beads/issues.jsonl`. Git ignores the Dolt database.
- There is no Dolt remote. A new clone needs `bd init --from-jsonl`.
- Two sessions share one database and can overwrite each other. Check with
  `bd show` before you set a status.
- Do not write a markdown TODO list. TODO.md is now in beads.
