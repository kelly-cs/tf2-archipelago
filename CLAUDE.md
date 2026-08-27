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

## Beads in this repository

The block below has the commands. These facts are local to this repository.

- The prefix is `apw`. P0 crash, P1 costs a player a run, P2 bug, P3 polish.
- Git tracks `.beads/issues.jsonl`. Git ignores the Dolt database.
- There is no Dolt remote. A new clone needs `bd init --from-jsonl`.
- Two sessions share one database and can overwrite each other. Check with
  `bd show` before you set a status.


<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:970c3bf2 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Agent Context Profiles

The managed Beads block is task-tracking guidance, not permission to override repository, user, or orchestrator instructions.

- **Conservative (default)**: Use `bd` for task tracking. Do not run git commits, git pushes, or Dolt remote sync unless explicitly asked. At handoff, report changed files, validation, and suggested next commands.
- **Minimal**: Keep tool instruction files as pointers to `bd prime`; use the same conservative git policy unless active instructions say otherwise.
- **Team-maintainer**: Only when the repository explicitly opts in, agents may close beads, run quality gates, commit, and push as part of session close. A current "do not commit" or "do not push" instruction still wins.

## Session Completion

This protocol applies when ending a Beads implementation workflow. It is subordinate to explicit user, repository, and orchestrator instructions.

1. **File issues for remaining work** - Create beads for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Handle git/sync by active profile**:
   ```bash
   # Conservative/minimal/default: report status and proposed commands; wait for approval.
   git status

   # Team-maintainer opt-in only, unless current instructions forbid it:
   git pull --rebase
   bd dolt push
   git push
   git status
   ```
5. **Hand off** - Summarize changes, validation, issue status, and any blocked sync/commit/push step

**Critical rules:**
- Explicit user or orchestrator instructions override this Beads block.
- Do not commit or push without clear authority from the active profile or the current user request.
- If a required sync or push is blocked, stop and report the exact command and error.
<!-- END BEADS INTEGRATION -->
