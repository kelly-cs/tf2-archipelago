# The weapon catalogue

`generate.py` writes `gamedata/weapons.go`. Run it when the bot mod's weapon
pools change:

```sh
python3 gamedata/cmd/weapons/generate.py > gamedata/weapons.go && gofmt -w gamedata/weapons.go
```

It reads two files and writes nothing by hand:

| Source | What it gives |
| --- | --- |
| `source/redbots3/loadouts.sp` in tf2-mvm-bots | the `WEAPONS_*` pools: which indexes a class may hold in which slot |
| `tf/scripts/items/items_game.txt` and `tf/resource/tf_english.txt` | what each index is called |

Both paths are at the top of the script. The game files come from any Team
Fortress 2 install; `~/tf2-native` is the one the test-bed uses.

Two things are left out. Stock, because the mod spells it as the default index
and every menu offers it separately. And the repaints: the Botkillers, the
Festives and the Australiums are the same guns in another colour, and a Scout
primary menu of twelve Scatterguns is not a menu.

`gamedata/weapons_test.go` is what keeps the result honest. It fails if a class
loses a slot, if a repaint gets in, or if one index picks up two names.
