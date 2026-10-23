package apclient

import (
	"fmt"
	"strings"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/gamedata/availability"
)

/*
unplayable refuses a seed that draws a mission this server cannot play, and
says what stands in each one's way.

The plugin refuses such a mission when the run switches to it, which can be an
hour in. The seed is known as soon as the room accepts the slot, so that is when
the player hears it. The answer is availability's, the one the launcher's
Missions page gives, so the fix it names is the one that page offers. A nil
server is a bridge that was told nothing about the server it serves, which is a
bridge in its own container: the plugin stays the last word there.
*/
func unplayable(server *availability.Server, missions []string) error {
	if server == nil {
		return nil
	}
	var refused []string
	for _, popFile := range missions {
		mission, known := gamedata.MissionByPopFile(popFile)
		if !known {
			continue
		}
		if verdict := server.Check(mission); !verdict.Playable() {
			refused = append(refused, mission.Name+" ("+verdict.Why()+")")
		}
	}
	if len(refused) == 0 {
		return nil
	}
	return fmt.Errorf(
		"this seed draws %s, and this server cannot play it. "+
			"Fix that in Settings > Missions and press Start again, or play a seed generated without it",
		strings.Join(refused, ", "))
}
