package gamedata

// MissionModifier is one rule a seed may attach to a mission. The text rides
// in slot data so every interface describes the exact rule the plugin applies.
type MissionModifier struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	Description    string `json:"description"`
	ExclusiveGroup string `json:"-"`
}

// MissionModifiers is ordered only for presentation; both real generation and
// test mode shuffle a copy before drawing from it.
var MissionModifiers = []MissionModifier{
	{Key: "low_gravity", Name: "Low Gravity", Kind: "environment", Description: "World gravity is halved, giving everyone much more airtime.", ExclusiveGroup: "gravity"},
	{Key: "high_gravity", Name: "High Gravity", Kind: "environment", Description: "Gravity is increased by 50%, making airtime hard-earned.", ExclusiveGroup: "gravity"},
	{Key: "blast_plating", Name: "Blast Plating", Kind: "robots", Description: "Enemy robots have 50% blast resistance."},
	{Key: "thermal_shielding", Name: "Thermal Shielding", Kind: "robots", Description: "Enemy robots have 50% fire resistance."},
	{Key: "ballistic_plating", Name: "Ballistic Plating", Kind: "robots", Description: "Enemy robots have 50% bullet resistance."},
	{Key: "fragile_mercenaries", Name: "Fragile Mercenaries", Kind: "players", Description: "Human defenders take 25% more damage."},
	{Key: "overclocked_servos", Name: "Overclocked Servos", Kind: "robots", Description: "Enemy robots move 15% faster."},
	{Key: "loaded_dice", Name: "Loaded Dice", Kind: "robots", Description: "Every enemy robot weapon can randomly crit, with a 33% base chance."},
	{Key: "weaponized_tanks", Name: "Weaponized Tanks", Kind: "robots", Description: "Each tank carries an invulnerable level 2 sentry until the tank is destroyed."},
	{Key: "miniature_menace", Name: "Miniature Menace", Kind: "robots", Description: "Robots and tanks are 50% size with 33% less health and 25% more speed."},
	{Key: "bot_surge", Name: "Bot Surge", Kind: "waves", Description: "Robot groups immediately refill open enemy slots, keeping the wave under constant pressure."},
	{Key: "faulty_calibration", Name: "Faulty Calibration", Kind: "players", Description: "Player weapons suffer severe bullet and flame spread, while projectiles and precision shots deviate from the crosshair."},
	{Key: "loose_footing", Name: "Loose Footing", Kind: "environment", Description: "Ground friction and acceleration are heavily reduced, giving everyone icy momentum, while all knockback is tripled."},
}
