# The weapon catalogue

`make weapons` writes `gamedata/weapons_generated.go`. Run it when the bot
mod's weapon pools change, which is when `go.mod` moves to a new
`tf2-mvm-bots-go`.

It reads two things and writes nothing by hand:

| Source | What it gives |
| --- | --- |
| `plugin/source/redbots3/generated/loadouts.sp` in the pinned bot mod | the `WEAPONS_*` pools: which indexes a class may hold in which slot |
| `tf/scripts/items/items_game.txt` and `tf/resource/tf_english.txt` | what each index is called |

The game files come from any Team Fortress 2 install; `TF2_DIR` names it and
defaults to `~/tf2-native/tf-dedicated/tf`, the one the bot test-bed uses.

Two things are left out. Stock, because the mod spells it as the default index
and every menu offers it separately. And the repaints: the Botkillers, the
Festives and the Australiums are the same guns in another colour, and a Scout
primary menu of twelve Scatterguns is not a menu.

`gamedata/weapons_test.go` is what keeps the result honest. It fails if a class
loses a slot, if a repaint gets in, or if one index picks up two names.
