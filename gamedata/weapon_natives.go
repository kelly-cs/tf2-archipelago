package gamedata

/*
nativeCuts are the flags a weapon already carries, so the buff that sets the
same flag does nothing on it (gh-47): the Righteous Bison already penetrates,
the Rocket Jumper already takes no blast damage, the Eviction Notice already
speeds up on a hit. A number the weapon already has stacks and stays, which is
why the Direct Hit keeps its damage buff and the Übersaw its ÜberCharge on hit.

The flag comes from items_game.txt, read with prefab inheritance: the effect's
own schema attribute already on every definition of the weapon. Two are spelled
differently there, energy weapon penetration on the Bison and speed_boost_on_hit
on the Eviction Notice. Two more are set in code rather than in the schema:
every fire weapon already ignites, and the Flare Gun already crits a burning
target. Cowser's sheet marks each of these Native, and this is the part of
those cells that is a flag.

Not here on purpose: the Axtinguisher, the Detonator and the Scorch Shot
mini-crit a burning target, so full crits on one is a step up and not a repeat.

Reskins need no row. Eligibility is decided on the family's canonical weapon
and a member never draws on its own, so the Festive Flare Gun is cut with the
Flare Gun.
*/
var nativeCuts = map[string]map[string]bool{
	"Backburner":                 names("back-crits", "ignite"),
	"Big Earner":                 names("speed-on-kill"),
	"Bushwacka":                  names("minicrits-to-crits"),
	"Candy Cane":                 names("drop-health-pack"),
	"Degreaser":                  names("ignite"),
	"Detonator":                  names("ignite"),
	"Direct Hit":                 names("airborne-minicrits"),
	"Dragon's Fury":              names("ignite"),
	"Eviction Notice":            names("speed-on-hit"),
	"Fan O'War":                  names("mark-for-death", "minicrits-to-crits"),
	"Flame Thrower":              names("ignite"),
	"Flare Gun":                  names("crits-vs-burning", "ignite"),
	"Hot Hand":                   names("speed-on-hit"),
	"Jarate":                     names("mark-for-death"),
	"Loose Cannon":               names("knockback"),
	"Mad Milk":                   names("mad-milk"),
	"Manmelter":                  names("ignite"),
	"Market Gardener":            names("airborne-crits"),
	"Natascha":                   names("slow-on-hit"),
	"Phlogistinator":             names("ignite"),
	"Reserve Shooter":            names("airborne-minicrits"),
	"Righteous Bison":            names("projectile-penetration"),
	"Rocket Jumper":              names("no-self-blast"),
	"Scorch Shot":                names("ignite", "knockback"),
	"Sharpened Volcano Fragment": names("ignite"),
	"Sticky Jumper":              names("no-self-blast"),
	"Sun-on-a-Stick":             names("crits-vs-burning"),
	"Sydney Sleeper":             names("mark-for-death"),
}

func cutByNative(name, key string) bool {
	return nativeCuts[name][key]
}
