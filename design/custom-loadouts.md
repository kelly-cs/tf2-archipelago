# Custom bot loadouts

Draft. Launcher side. The mod side is `specs/custom-loadouts.md` in
tf2-mvm-bots.

## The ask

From the Discord thread, 2026-08-25 (Cowser):

> speaking of bots, it'd be nice to have more customization for what they can
> use. Like instead of Scorch Shot for the Phlog build, it's the Gas Passer. Or
> instead of Kunai, it's Big Earner.

So: build your own loadout, name it, keep it. The same shape as the saved
teams already on the Bots tab.

## What is there today

The launcher hands the mod a fixed menu. `botloadout.Classes`
(`launcher/internal/botloadout/botloadout.go:47`) holds nine classes with two
presets each plus stock, and every preset is a literal: four item definition
indexes and a name. A pick is a preset key, per class
(`SrcdsBotLoadouts`) or per seat (`SrcdsBotSeatLoadouts`), and
`Render` turns the pair into `configs/defenderbots/loadout.cfg`.

Saved teams already work: `settings.BotTeam` holds the seats, their loadouts,
the per-class picks and the blacklist. `presetRow`
(`launcher/internal/gui/settings_windows.go:1024`) is the Save / Load / Remove
row. Whatever this feature grows should read like that row, because that is
the control the user is asking for by name.

## The finding that decides the design

The mod does not validate the indexes. `GetServerLoadoutWeapon`
(`source/redbots3/player_pref.sp:133`) reads whatever number the KeyValues
file carries and `PrepareCustomLoadout` (`source/redbots3/loadouts.sp:257`)
gives it. The `WEAPONS_*` arrays at the top of `loadouts.sp` are the pool the
mod draws *random* loadouts from. They are not a filter on the server loadout.

So the weapon half of this feature needs no mod change at all. The whole limit
is the hardcoded `Loadouts` slice in Go. That is what makes this cheap.

Two things still bound what is worth offering:

- A weapon whose behaviour the bot AI does not implement is a legal pick that
  plays badly. The Beggar's Bazooka is already a preset and the bot does not
  charge it. That is the mod-side doc's problem, not this one's, but the menu
  should be able to mark it.
- The Spy's watch is `pda2` and the Spy's sapper is `secondary` in the mod's
  spelling. `Loadout.Second` doubles as both. Keep that.

## Data

ADR 0001 says Go owns the game data, so the weapon catalogue belongs in
`gamedata/`, next to `slots.go` and `classes.go`, not in `botloadout`.

Add `gamedata/weapons.go`: every weapon the mod can hand out, as
`{DefIndex, Name, Class, Slot}`. Seed it from the `WEAPONS_*` arrays so the two
repos cannot drift, and drop the reskins (the 15xxx and 30xxx blocks) from what
the menu shows: a Scout primary list of thirty entries where twenty are
Festive Scatterguns is not a menu.

`botloadout` then reads the catalogue instead of carrying names in prose.
`Loadout.Weapons`, the display string, becomes derived rather than typed by
hand, which also kills the one place where a preset's name and its indexes can
disagree.

Open: whether `gamedata` exports the catalogue to the apworld as well. It has
no use for it. Keep it Go-side and out of `export.go` until something asks.

## Settings

```go
// A loadout somebody built, keyed by the name they gave it.
SrcdsBotCustomLoadouts map[string]CustomLoadout `json:"srcds_bot_custom_loadouts,omitempty"`

type CustomLoadout struct {
    Class   string `json:"class"`
    Primary int    `json:"primary"`
    Second  int    `json:"second"`
    Melee   int    `json:"melee"`
    PDA2    int    `json:"pda2,omitempty"`
}
```

Keyed by name, and the name is the loadout key. That collides with the built-in
keys (`phlog`, `kunai`), so custom keys get a prefix: `custom:Gas runner`. A
prefix rather than a separate field, because everything downstream already
passes loadout keys around as strings and a second field means touching
`BotTeam`, `Seats`, `Render`, both UIs and the config format.

`Class.LoadoutByKey` gains the lookup: prefix means user table, no prefix means
the built-in slice, unknown means stock. Unknown-is-stock is already the rule
(`botloadout.go:117`), so a saved team naming a loadout the user has since
deleted keeps working. That is the behaviour to preserve, and the test to write
first.

`BotTeam` does not change. A saved team stores loadout keys, and a custom key
is a key. Worth stating in the doc that a saved team does not carry the
loadouts it names: delete the loadout and the team that used it falls to stock.
The alternative, copying the weapons into the team, means a team and a loadout
can disagree about what "Gas runner" is. Do not do that.

## UI

A fourth sub-tab, next to Team, Classes and Looks
(`settings_windows.go:1471`):

```go
botsScrollPage("Loadouts", 2, loadoutRows(s, label, team)),
```

The page is one editor, not a list:

- a name box and Load / Save / Remove, copied from `presetRow`
- a class menu
- four slot menus, filtered by the chosen class, with the Spy's watch shown
  only for the Spy
- the slot menus reset when the class changes, the same rule the seat menus
  already follow (`settings_windows.go:1076`)

Nothing new appears on the Team or Classes tabs. Their existing loadout menus
simply grow the custom entries at the bottom of the same list, under a
separator, which is why the key prefix has to be enough on its own.

Watch the build cost. The Bots tab is already two thirds of the settings
dialog's open time and is built lazily for that reason
(`settings_windows.go:463`). Four combo boxes is nothing; a per-slot list of
every Scout primary is not. Filter the reskins out before they reach walk.

The TUI mirrors it (`launcher/internal/tui/settings.go:386`). Same fields, same
order, no new concepts.

## Render

`Render` is unchanged. It already writes four slot keys per block and skips
stock, and it does not care where the indexes came from. `Custom` and
`CustomSeats` gain nothing either: a custom key is not `stock`, so they already
report true.

The only render-side question is validation. A custom loadout can name a
weapon the class cannot hold, because a user can hand-edit `config.json`.
Validate at load: drop a slot whose index is not in the catalogue for that class
and slot, and log it. Boundary validation, not a runtime check in `Render`.

## Order of work

1. `gamedata/weapons.go` and its test. Nothing else can start.
2. `botloadout` reads the catalogue; `Loadout.Weapons` becomes derived. No
   behaviour change, so the existing tests carry it.
3. `settings.CustomLoadout`, the prefix lookup, load-time validation.
4. GUI sub-tab.
5. TUI mirror.
6. Docs in `docs/en` and `docs/fr`, and a CHANGELOG entry.

Steps 1 to 3 are the feature. Steps 4 and 5 are the part somebody sees.

## Open questions

- Does a custom loadout belong to a saved team or to the whole config? Written
  above as config-wide. Per-team means every team carries a copy and two teams
  can hold different "Gas runner"s.
- Should the menu mark weapons the bot AI does not drive well? It needs a
  support list from the mod repo. See the mod-side doc.
- Attributes. `loadouts.sp` carries per-slot runtime attributes
  (`MAX_RUNTIME_ATTRIBUTES`) and `loadout.cfg` says nothing about them. Out of
  scope here, but the KeyValues shape should leave room: a `attributes` block
  under a slot, not a flat key.
