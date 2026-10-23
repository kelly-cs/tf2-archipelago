package settings

import "github.com/m-this/tf2-archipelago/gamedata"

/*
ComposeDefaults is what a Compose stack's settings start from before ApplyEnv
reads its .env: Defaults, with every community pack ticked.

The srcds container installs every pack in community-content whatever the admin
page ticks, and a .env written before TF2AP_COMMUNITY_PACKS reached the seed
containers has no such key. Starting from no pack would turn every one of those
stacks into a Valve-only pool without a word. So a missing key keeps every
pack, and only a key that is there and empty means none: that is the admin
page unticking the last one.
*/
func ComposeDefaults() Settings {
	s := Defaults()
	s.CommunityPacks = gamedata.CommunityPacks()
	return s
}
